// Copyright (c) 2017 Uber Technologies, Inc.
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

package history

import (
	"context"
	"time"

	"github.com/uber/cadence/.gen/go/history"
	"github.com/uber/cadence/.gen/go/history/historyserviceclient"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/service/worker/scanner"
	"go.uber.org/cadence/activity"
	"golang.org/x/time/rate"
)

type (
	// Scavenger is the type that holds the state for history scavenger daemon
	Scavenger struct {
		db             p.HistoryV2Manager
		workflowClient historyserviceclient.Interface
		hbd            scanner.HistoryScavengerActivityHeartbeatDetails
		rps            int
		limiter        *rate.Limiter
		metrics        metrics.Client
		logger         log.Logger
	}

	taskDetail struct {
		domainID   string
		workflowID string
		runID      string
		treeID     string
		branchID   string

		// passing along the current heartbeat details to make heartbeat within a task so that it won't timeout
		hbd scanner.HistoryScavengerActivityHeartbeatDetails
	}
)

const (
	// used this to decide how many goroutines to process
	rpsPerConcurrency = 50
	pageSize          = 1000
	// only clean up history branches that older than this threshold
	cleanUpThreshold = time.Hour * 24
)

// NewScavenger returns an instance of history scavenger daemon
// The Scavenger can be started by calling the Run() method on the
// returned object. Calling the Run() method will result in one
// complete iteration over all of the history branches in the system. For
// each branch, the scavenger will attempt
//  - describe the corresponding workflow execution
//  - deletion of history itself, if there are no workflow execution
func NewScavenger(
	db p.HistoryV2Manager,
	rps int,
	client historyserviceclient.Interface,
	hbd scanner.HistoryScavengerActivityHeartbeatDetails,
	metricsClient metrics.Client,
	logger log.Logger,
) *Scavenger {

	rateLimiter := rate.NewLimiter(rate.Limit(rps), rps)

	return &Scavenger{
		db:             db,
		workflowClient: client,
		hbd:            hbd,
		rps:            rps,
		limiter:        rateLimiter,
		metrics:        metricsClient,
		logger:         logger,
	}
}

// Start starts the scavenger
func (s *Scavenger) Run(ctx context.Context) (scanner.HistoryScavengerActivityHeartbeatDetails, error) {
	taskCh := make(chan taskDetail, pageSize)
	respCh := make(chan error, pageSize)
	concurrency := s.rps/rpsPerConcurrency + 1

	for i := 0; i < concurrency; i++ {
		go s.startTaskProcessor(ctx, taskCh, respCh)
	}

	for {
		resp, err := s.db.GetAllHistoryTreeBranches(&p.GetAllHistoryTreeBranchesRequest{
			NextPageToken: s.hbd.NextPageToken,
		})
		if err != nil {
			return s.hbd, err
		}
		batchCount := len(resp.Branches)
		if batchCount <= 0 {
			break
		}

		skips := 0
		errsSplit := 0
		// send all tasks
		for _, br := range resp.Branches {
			if time.Now().Add(-cleanUpThreshold).Before(br.ForkTime) {
				//metric
				batchCount--
				skips++
				continue
			}

			domainID, wid, rid, err := p.SplitHistoryGarbageCleanupInfo(br.Info)
			if err != nil {
				//metric, log
				batchCount--
				errsSplit++
				continue
			}

			taskCh <- taskDetail{
				domainID:   domainID,
				workflowID: wid,
				runID:      rid,
				treeID:     br.TreeID,
				branchID:   br.BranchID,

				hbd: s.hbd,
			}
		}

		succCount := 0
		errCount := 0
		// wait for counters indicate this batch is done
	Loop:
		for {
			select {
			case err := <-respCh:
				if err == nil {
					succCount++
				} else {
					errCount++
				}
				if succCount+errCount == batchCount {
					break Loop
				}
			case <-ctx.Done():
				return s.hbd, ctx.Err()
			}
		}

		s.hbd.CurrentPage++
		s.hbd.NextPageToken = resp.NextPageToken
		s.hbd.SkipCount += succCount
		s.hbd.ErrorCount += errCount + errsSplit
		s.hbd.SkipCount += skips
		activity.RecordHeartbeat(ctx, s.hbd)

		if len(s.hbd.NextPageToken) == 0 {
			break
		}
	}
	return s.hbd, nil
}

func (s *Scavenger) startTaskProcessor(ctx context.Context, taskCh chan taskDetail, respCh chan error) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-taskCh:
			if isDone(ctx) {
				return
			}

			activity.RecordHeartbeat(ctx, s.hbd)

			err := s.limiter.Wait(ctx)
			if err != nil {
				respCh <- err
				continue
			}

			// this checks if the mutableState still exists
			// if not then delete the history branch
			_, err = s.workflowClient.DescribeMutableState(ctx, &history.DescribeMutableStateRequest{
				DomainUUID: common.StringPtr(task.domainID),
				Execution: &shared.WorkflowExecution{
					WorkflowId: common.StringPtr(task.workflowID),
					RunId:      common.StringPtr(task.runID),
				},
			})

			if err != nil {
				if _, ok := err.(*shared.EntityNotExistsError); ok {
					//delete history
					branchToken, err := p.NewHistoryBranchTokenByBranchID(task.treeID, task.branchID)
					if err != nil {
						respCh <- err
						continue
					}

					err = s.db.DeleteHistoryBranch(&p.DeleteHistoryBranchRequest{
						BranchToken: branchToken,
						// This is a required argument but it is not needed for Cassandra.
						// Since this scanner is only for Cassandra,
						// we can fill any number here to let to code go through
						ShardID: common.IntPtr(1),
					})
					if err != nil {
						respCh <- err
						continue
					} else {
						// deleted garbage
					}
				} else {
					respCh <- err
					continue
				}
			} else {
				// no garbage
				respCh <- nil
			}
		}
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
