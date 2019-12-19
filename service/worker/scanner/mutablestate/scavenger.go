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

package mutablestate

import (
	"context"
	"encoding/json"
	"github.com/uber/cadence/common/backoff"
	"time"

	"go.uber.org/cadence/activity"
	"golang.org/x/time/rate"

	"github.com/uber/cadence/.gen/go/history"
	"github.com/uber/cadence/.gen/go/history/historyserviceclient"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	p "github.com/uber/cadence/common/persistence"
)

type (
	// ScavengerHeartbeatDetails is the heartbeat detail for MutableStateScavengerActivity
	ScavengerHeartbeatDetails struct {
		NextPageToken  []byte
		CurrentShardID int
		CurrentPage    int

		// skip if current workflow is not open, or corrupted state with start time less than threshold to delete
		SkipCount      int
		// some unknown errors
		ErrorCount     int
		// number of workflows that converted into corrupted state
		CorruptedCount int
		// number of deleted workflows
		DeletedCount   int
	}

	// Scavenger is the type that holds the state for MutableState scavenger daemon
	Scavenger struct {
		shardDB     p.ShardManager
		historyDB   p.HistoryManager
		client historyserviceclient.Interface
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
		skip bool
		err  error
		corrupted bool
		deleted bool
	}
)

const (
	// used this to decide how many goroutines to process
	rpsPerConcurrency = 50
	pageSize          = 1000
	// only clean up current workflow record and mutable state records if the started time is older than the threshold
	// otherwise only put the workflow into corrupted state
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
	client historyserviceclient.Interface,
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
		client:client,
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
		var resp *p.ScanMutableStateResponse
		var err error
		op := func() error {
			resp, err = s.shardDB.ScanMutableState(&p.ScanMutableStateRequest{
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
		skipCount := 0
		// send all tasks
		for _, wf := range resp.WorkflowInfos {
			if wf.State != p.WorkflowStateRunning && wf.State != p.WorkflowStateCorrupted {
				skipCount++
			}
			taskCh <- taskDetail{
				shardID: s.hbd.CurrentShardID,
				domainID:   wf.DomainID,
				workflowID: wf.WorkflowID,
				runID:      wf.RunID,

				hbd: s.hbd,
			}
			taskCount ++
		}

		corruptedCount := 0
		deletedCount := 0
		errorCount := 0
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
						s.metrics.IncCounter(metrics.MutableStateScavengerScope, metrics.MutableStateScavengerCorruptedCount)
						corruptedCount++
					}
					if resp.deleted {
						s.metrics.IncCounter(metrics.MutableStateScavengerScope, metrics.MutableStateScavengerDeleteCount)
						deletedCount++
					}
					if resp.skip {
						s.metrics.IncCounter(metrics.MutableStateScavengerScope, metrics.MutableStateScavengerSkipCount)
						skipCount++
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
		s.hbd.CorruptedCount += corruptedCount
		s.hbd.ErrorCount += errorCount
		s.hbd.DeletedCount += deletedCount
		s.hbd.SkipCount += skipCount

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

			// this checks if the mutableState still exists
			// if not then the history branch is garbage, we need to delete the history branch
			resp, err := s.client.DescribeMutableState(ctx, &history.DescribeMutableStateRequest{
				DomainUUID: common.StringPtr(task.domainID),
				Execution: &shared.WorkflowExecution{
					WorkflowId: common.StringPtr(task.workflowID),
					RunId:      common.StringPtr(task.runID),
				},
			})

			if err != nil {
				res.err = err
				s.logger.Error("encounter error when describeMutableState",
					getTaskLoggingTags(err, task)...)
				respCh <- res
			} else {
				msStr := resp.GetMutableStateInDatabase()
				ms := p.WorkflowMutableState{}
				res.err = json.Unmarshal([]byte(msStr), &ms)
				if err != nil {
					respCh <- res
					continue
				}


			}
		}
	}
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
