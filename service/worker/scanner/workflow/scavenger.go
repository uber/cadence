// Copyright (c) 2019 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package workflow

import (
	"context"
	"github.com/uber/cadence/.gen/go/shared"
	"time"

	"go.uber.org/cadence/activity"
	"golang.org/x/time/rate"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/resource"
)

type (
	// ScavengerHeartbeatDetails is the heartbeat detail for MutableStateScavengerActivity
	ScavengerHeartbeatDetails struct {
		NextPageToken  []byte
		CurrentShardID int
		CurrentPage    int

		// number of executions processed
		ProcessedCount      int
		// number of workflows that are corrupted
		CorruptionCount int
		// number of task failure due to some errors
		ErrorCount int
	}

	// Scavenger is the type that holds the state for Workflow scavenger daemon:
	// 1. It scan each shard for current workflow record and concrete record(a.k.a mutableState)
	// 2. Detect corruption of workflow based
	// 3. Turns open workflow into corrupted state if startTime not over threshold
	// 4. Deleted corrupted record if over threshold
	//
	// However, step 2 can be quite complicated:
	//Here are all the cases that a workflow run(with unique runID) should be considered as "corrupted":
	//a) a run has both current record and mutableState, workflowState is open, but history is corrupted
	//b) a run has both current record and mutableState, workflowState is closed, but history is corrupted
	//c) a run has only current record but no mutableState
	//d) a run has only mutableState but no current record, workflow state is open
	//e) a run with completed history without corruption, but timerMap/activityMap/etc are corrupted.
	//
	// Today this Scavenger only detecting for a) without performing actions on it.
	// TODO https://github.com/uber/cadence/issues/2926 for more details.
	Scavenger struct {
		shardDB     p.ShardManager
		historyDB   p.HistoryManager
		resource    resource.Resource
		hbd      ScavengerHeartbeatDetails
		rps      int
		numShards int
		limiter  *rate.Limiter
		metrics  metrics.Client
		logger   log.Logger
		isInTest bool
	}

	taskDetail struct {
		shardID    int
		domainID   string
		workflowID string
		runID      string

		// passing along the current heartbeat details to make heartbeat within a task so that it won't timeout
		hbd ScavengerHeartbeatDetails
	}

	taskResult struct{
		err  error
		corrupted bool
	}

)



const (
	// used this to decide how many goroutines to process
	rpsPerConcurrency = 50
	pageSize          = 1000
	// clean up mutable state records if the started time is older than this threshold
	cleanUpThreshold = time.Hour * 24 * 365
)

// NewScavenger returns an instance of history scavenger daemon
// The Scavenger can be started by calling the Run() method on the
// returned object. Calling the Run() method will result in one
// complete iteration over all of the history branches in the system. For
// each branch, the scavenger will attempt
//  - describe the corresponding workflow execution
//  - deletion of history itself, if there are no workflow execution
func NewScavenger(
	shardDB   p.ShardManager,
	historyDB p.HistoryManager,
	resource resource.Resource,
	rps int,
	numShards int,
	hbd ScavengerHeartbeatDetails,
	metricsClient metrics.Client,
	logger log.Logger,
) *Scavenger {

	rateLimiter := rate.NewLimiter(rate.Limit(rps), rps)

	return &Scavenger{
		shardDB:shardDB,
		historyDB:historyDB,
		resource:resource,
		hbd:     hbd,
		rps:     rps,
		numShards:numShards,
		limiter: rateLimiter,
		metrics: metricsClient,
		logger:  logger,
	}
}

var persistenceOperationRetryPolicy = common.CreatePersistanceRetryPolicy()

// Run runs the scavenger
func (s *Scavenger) Run(ctx context.Context) (ScavengerHeartbeatDetails, error) {
	taskCh := make(chan taskDetail, pageSize)
	respCh := make(chan taskResult, pageSize)
	concurrency := s.rps/rpsPerConcurrency + 1

	for i := 0; i < concurrency; i++ {
		go s.startTaskProcessor(ctx, taskCh, respCh)
	}

	for {
		var resp *p.ScanWorkflowsResponse
		var err error
		op := func() error {
			resp, err = s.shardDB.ScanWorkflows(&p.ScanWorkflowsRequest{
				PageSize:      pageSize,
				NextPageToken: s.hbd.NextPageToken,
				ShardID: s.hbd.CurrentShardID,
			})
			return err
		}

		err = backoff.Retry(op, persistenceOperationRetryPolicy, common.IsPersistenceTransientError)

		if err != nil {
			return s.hbd, err
		}

		taskCount := 0
		corruptionCount := 0
		errorCount := 0
		for _, wf := range resp.WorkflowExecutions {
			if wf.IsCurrentWorkflow && p.IsWorkflowRunning(wf.State){
				taskCount ++
				taskCh <- taskDetail{
					shardID:    s.hbd.CurrentShardID,
					domainID:   wf.DomainID,
					workflowID: wf.WorkflowID,
					runID:      wf.RunID,
					hbd:s.hbd,
				}
			}
		}

		if taskCount > 0 {
			// wait for task counters indicate this batch is done
		Loop:
			for {
				select {
				case resp := <-respCh:
					taskCount --
					if resp.err != nil {
						s.metrics.IncCounter(metrics.MutableStateScavengerScope, metrics.MutableStateScavengerErrorCount)
						errorCount++
					}
					if resp.corrupted {
						s.metrics.IncCounter(metrics.MutableStateScavengerScope, metrics.MutableStateScavengerCorruptionCount)
						corruptionCount++
					}
					if taskCount == 0 {
						break Loop
					}
				case <-ctx.Done():
					return s.hbd, ctx.Err()
				}
			}
		}

		s.hbd.CurrentPage++
		s.hbd.NextPageToken = resp.NextPageToken
		s.hbd.ErrorCount += errorCount
		s.hbd.ProcessedCount += len(resp.WorkflowExecutions)
		s.hbd.CorruptionCount += corruptionCount

		if len(s.hbd.NextPageToken) == 0 {
			if s.hbd.CurrentShardID == s.numShards - 1{
				break
			}
			s.hbd.CurrentShardID ++
		}
		if !s.isInTest {
			activity.RecordHeartbeat(ctx, s.hbd)
		}
	}
	return s.hbd, nil
}

func (s *Scavenger) startTaskProcessor(
	ctx context.Context,
	taskCh chan taskDetail,
	respCh chan taskResult,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-taskCh:
			if isDone(ctx) {
				return
			}

			if !s.isInTest {
				activity.RecordHeartbeat(ctx, s.hbd)
			}

			res := taskResult{}
			res.err = s.limiter.Wait(ctx)
			if res.err != nil {
				respCh <- res
				s.logger.Error("encounter error when wait for rate limiter",
					getTaskLoggingTags(res.err, task)...)
				continue
			}

			res.corrupted, res.err = s.processHistoryCorruptionTask(task)
			if res.err != nil{
				s.logger.Error("encounter error when process task",
					getTaskLoggingTags(res.err, task)...)
			}
			respCh <- res
		}
	}
}

func (s *Scavenger) processHistoryCorruptionTask(task taskDetail) (bool, error) {
	mgr, err := s.resource.GetExecutionManager(task.shardID)
	if err != nil{
		return false, err
	}

	var resp *p.GetWorkflowExecutionResponse
	op := func() error {
		resp, err = mgr.GetWorkflowExecution(&p.GetWorkflowExecutionRequest{
			DomainID:task.domainID,
			Execution:shared.WorkflowExecution{
				WorkflowId:common.StringPtr(task.workflowID),
				RunId:common.StringPtr(task.runID),
			},
		})
		return err
	}
	err = backoff.Retry(op, persistenceOperationRetryPolicy, common.IsPersistenceTransientError)
	if err != nil{
		// TODO if mutable state doesn't exists, we should delete the current record
		// this is case c) in https://github.com/uber/cadence/issues/2926
		return false, err
	}

	op = func() error {
		_, err = s.historyDB.ReadHistoryBranch(&p.ReadHistoryBranchRequest{
			BranchToken:resp.State.ExecutionInfo.BranchToken,
			MinEventID:common.FirstEventID,
			MaxEventID:resp.State.ExecutionInfo.NextEventID,
			PageSize:pageSize,
			ShardID: &task.shardID,
		})
		return err
	}

	err = backoff.Retry(op, persistenceOperationRetryPolicy, common.IsPersistenceTransientError)
	if err == nil{
		// not corrupted
		return false, nil
	}
	if _, ok := err.(*shared.InternalDataCorruptionError); !ok{
		// some other error
		return false, err
	}

	// for corrupted, we need to
	// 1. delete current record
	// 2. covert into corrupted state
	// TODO  https://github.com/uber/cadence/issues/2926
	return true, nil
}

func getTaskLoggingTags(err error, task taskDetail) []tag.Tag {
	if err != nil {
		return []tag.Tag{
			tag.Error(err),
			tag.WorkflowDomainID(task.domainID),
			tag.WorkflowID(task.workflowID),
			tag.WorkflowRunID(task.runID),
		}
	}
	return []tag.Tag{
		tag.WorkflowDomainID(task.domainID),
		tag.WorkflowID(task.workflowID),
		tag.WorkflowRunID(task.runID),
	}
}

func isDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
