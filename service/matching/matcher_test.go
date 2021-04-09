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

package matching

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/client/matching"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

type MatcherTestSuite struct {
	suite.Suite
	controller  *gomock.Controller
	client      *matching.MockClient
	fwdr        *Forwarder
	cfg         *taskListConfig
	taskList    *taskListID
	matcher     *TaskMatcher // matcher for child partition
	rootMatcher *TaskMatcher // matcher for parent partition
}

func TestMatcherSuite(t *testing.T) {
	suite.Run(t, new(MatcherTestSuite))
}

func (t *MatcherTestSuite) SetupTest() {
	t.controller = gomock.NewController(t.T())
	t.client = matching.NewMockClient(t.controller)
	cfg := NewConfig(dynamicconfig.NewNopCollection())
	t.taskList = newTestTaskListID(uuid.New(), common.ReservedTaskListPrefix+"tl0/1", persistence.TaskListTypeDecision)
	tlCfg, err := newTaskListConfig(t.taskList, cfg, t.newDomainCache())
	t.NoError(err)
	tlCfg.forwarderConfig = forwarderConfig{
		ForwarderMaxOutstandingPolls: func() int { return 1 },
		ForwarderMaxOutstandingTasks: func() int { return 1 },
		ForwarderMaxRatePerSecond:    func() int { return 2 },
		ForwarderMaxChildrenPerNode:  func() int { return 20 },
	}
	t.cfg = tlCfg
	t.fwdr = newForwarder(&t.cfg.forwarderConfig, t.taskList, types.TaskListKindNormal, t.client)
	t.matcher = newTaskMatcher(tlCfg, t.fwdr, func() metrics.Scope { return metrics.NoopScope(metrics.Matching) })

	rootTaskList := newTestTaskListID(t.taskList.domainID, t.taskList.Parent(20), persistence.TaskListTypeDecision)
	rootTasklistCfg, err := newTaskListConfig(rootTaskList, cfg, t.newDomainCache())
	t.NoError(err)
	t.rootMatcher = newTaskMatcher(rootTasklistCfg, nil, func() metrics.Scope { return metrics.NoopScope(metrics.Matching) })
}

func (t *MatcherTestSuite) TearDownTest() {
	t.controller.Finish()
}

func (t *MatcherTestSuite) TestLocalSyncMatch() {
	// force disable remote forwarding
	<-t.fwdr.AddReqTokenC()
	<-t.fwdr.PollReqTokenC()

	wait := ensureAsyncReady(time.Second, func(ctx context.Context) {
		task, err := t.matcher.Poll(ctx)
		if err == nil {
			task.finish(nil)
		}
	})

	task := newInternalTask(t.newTaskInfo(), nil, types.TaskSourceHistory, "", true)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	syncMatch, err := t.matcher.Offer(ctx, task)
	cancel()
	wait()
	t.NoError(err)
	t.True(syncMatch)
}

func (t *MatcherTestSuite) TestRemoteSyncMatch() {
	t.testRemoteSyncMatch(types.TaskSourceHistory)
}

func (t *MatcherTestSuite) TestRemoteSyncMatchBlocking() {
	t.testRemoteSyncMatch(types.TaskSourceDbBacklog)
}

func (t *MatcherTestSuite) testRemoteSyncMatch(taskSource types.TaskSource) {
	pollSigC := make(chan struct{})

	bgctx, bgcancel := context.WithTimeout(context.Background(), time.Second)
	go func() {
		<-pollSigC
		if taskSource == types.TaskSourceDbBacklog {
			// when task is from dbBacklog, sync match SHOULD block
			// so lets delay polling by a bit to verify that
			time.Sleep(time.Millisecond * 10)
		}
		task, err := t.matcher.Poll(bgctx)
		bgcancel()
		if err == nil && !task.isStarted() {
			task.finish(nil)
		}
	}()

	t.client.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any()).DoAndReturn(
		func(arg0 context.Context, arg1 *types.MatchingPollForDecisionTaskRequest) (*types.MatchingPollForDecisionTaskResponse, error) {
			task, err := t.rootMatcher.Poll(arg0)
			if err != nil {
				return nil, err
			}

			task.finish(nil)
			return &types.MatchingPollForDecisionTaskResponse{
				WorkflowExecution: task.workflowExecution(),
			}, nil
		},
	).AnyTimes()

	task := newInternalTask(t.newTaskInfo(), nil, taskSource, "", true)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	var err error
	var remoteSyncMatch bool
	var req *types.AddDecisionTaskRequest
	t.client.EXPECT().AddDecisionTask(gomock.Any(), gomock.Any()).Do(
		func(arg0 context.Context, arg1 *types.AddDecisionTaskRequest) {
			req = arg1
			task.forwardedFrom = req.GetForwardedFrom()
			close(pollSigC)
			if taskSource != types.TaskSourceDbBacklog {
				// when task is not from backlog, wait a bit for poller
				// to arrive first - when task is from backlog, offer
				// blocks - so we don't need to do this
				time.Sleep(10 * time.Millisecond)
			}
			remoteSyncMatch, err = t.rootMatcher.Offer(ctx, task)
		},
	).Return(nil)

	_, err0 := t.matcher.Offer(ctx, task)
	t.NoError(err0)
	cancel()
	<-bgctx.Done() // wait for async work to finish
	t.NotNil(req)
	t.NoError(err)
	t.True(remoteSyncMatch)
	t.Equal(t.taskList.name, req.GetForwardedFrom())
	t.Equal(t.taskList.Parent(20), req.GetTaskList().GetName())
}

func (t *MatcherTestSuite) TestSyncMatchFailure() {
	task := newInternalTask(t.newTaskInfo(), nil, types.TaskSourceHistory, "", true)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	var req *types.AddDecisionTaskRequest
	t.client.EXPECT().AddDecisionTask(gomock.Any(), gomock.Any()).Do(
		func(arg0 context.Context, arg1 *types.AddDecisionTaskRequest) {
			req = arg1
		},
	).Return(errMatchingHostThrottle)

	syncMatch, err := t.matcher.Offer(ctx, task)
	cancel()
	t.NotNil(req)
	t.NoError(err)
	t.False(syncMatch)
}

func (t *MatcherTestSuite) TestQueryLocalSyncMatch() {
	// force disable remote forwarding
	<-t.fwdr.AddReqTokenC()
	<-t.fwdr.PollReqTokenC()

	wait := ensureAsyncReady(time.Second, func(ctx context.Context) {
		task, err := t.matcher.PollForQuery(ctx)
		if err == nil && task.isQuery() {
			task.finish(nil)
		}
	})

	task := newInternalQueryTask(uuid.New(), &types.MatchingQueryWorkflowRequest{})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := t.matcher.OfferQuery(ctx, task)
	cancel()
	wait()
	t.NoError(err)
	t.Nil(resp)
}

func (t *MatcherTestSuite) TestQueryRemoteSyncMatch() {
	ready, wait := ensureAsyncAfterReady(time.Second, func(ctx context.Context) {
		task, err := t.matcher.PollForQuery(ctx)
		if err == nil && task.isQuery() {
			task.finish(nil)
		}
	})

	var remotePollResp *types.MatchingPollForDecisionTaskResponse
	t.client.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any()).DoAndReturn(
		func(arg0 context.Context, arg1 *types.MatchingPollForDecisionTaskRequest) (*types.MatchingPollForDecisionTaskResponse, error) {
			task, err := t.rootMatcher.PollForQuery(arg0)
			if err != nil {
				return nil, err
			} else if task.isQuery() {
				task.finish(nil)
				res := &types.MatchingPollForDecisionTaskResponse{
					Query: &types.WorkflowQuery{},
				}
				remotePollResp = res
				return res, nil
			}
			return nil, nil
		},
	).AnyTimes()

	task := newInternalQueryTask(uuid.New(), &types.MatchingQueryWorkflowRequest{})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	var req *types.MatchingQueryWorkflowRequest
	t.client.EXPECT().QueryWorkflow(gomock.Any(), gomock.Any()).Do(
		func(arg0 context.Context, arg1 *types.MatchingQueryWorkflowRequest) {
			req = arg1
			task.forwardedFrom = req.GetForwardedFrom()
			ready()
			t.rootMatcher.OfferQuery(ctx, task)
		},
	).Return(&types.QueryWorkflowResponse{QueryResult: []byte("answer")}, nil)

	result, err := t.matcher.OfferQuery(ctx, task)
	cancel()
	wait()
	t.NotNil(req)
	t.NoError(err)
	t.NotNil(result)
	t.NotNil(remotePollResp.Query)
	t.Equal("answer", string(result.QueryResult))
	t.Equal(t.taskList.name, req.GetForwardedFrom())
	t.Equal(t.taskList.Parent(20), req.GetTaskList().GetName())
}

func (t *MatcherTestSuite) TestQueryRemoteSyncMatchError() {
	<-t.fwdr.PollReqTokenC()

	matched := false
	ready, wait := ensureAsyncAfterReady(time.Second, func(ctx context.Context) {
		task, err := t.matcher.PollForQuery(ctx)
		if err == nil && task.isQuery() {
			matched = true
			task.finish(nil)
		}
	})

	task := newInternalQueryTask(uuid.New(), &types.MatchingQueryWorkflowRequest{})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	var req *types.MatchingQueryWorkflowRequest
	t.client.EXPECT().QueryWorkflow(gomock.Any(), gomock.Any()).Do(
		func(arg0 context.Context, arg1 *types.MatchingQueryWorkflowRequest) {
			req = arg1
			ready()
		},
	).Return(nil, errMatchingHostThrottle)

	result, err := t.matcher.OfferQuery(ctx, task)
	cancel()
	wait()
	t.NotNil(req)
	t.NoError(err)
	t.Nil(result)
	t.True(matched)
}

func (t *MatcherTestSuite) TestMustOfferLocalMatch() {
	// force disable remote forwarding
	<-t.fwdr.AddReqTokenC()
	<-t.fwdr.PollReqTokenC()

	wait := ensureAsyncReady(time.Second, func(ctx context.Context) {
		task, err := t.matcher.Poll(ctx)
		if err == nil {
			task.finish(nil)
		}
	})

	task := newInternalTask(t.newTaskInfo(), nil, types.TaskSourceHistory, "", false)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	err := t.matcher.MustOffer(ctx, task)
	cancel()
	wait()
	t.NoError(err)
}

func (t *MatcherTestSuite) TestMustOfferRemoteMatch() {
	pollSigC := make(chan struct{})

	t.client.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any()).DoAndReturn(
		func(arg0 context.Context, arg1 *types.MatchingPollForDecisionTaskRequest) (*types.MatchingPollForDecisionTaskResponse, error) {
			<-pollSigC
			time.Sleep(time.Millisecond * 500) // delay poll to verify that offer blocks on parent
			task, err := t.rootMatcher.Poll(arg0)
			if err != nil {
				return nil, err
			}

			task.finish(nil)
			return &types.MatchingPollForDecisionTaskResponse{
				WorkflowExecution: task.workflowExecution(),
			}, nil
		},
	).AnyTimes()

	taskCompleted := false
	completionFunc := func(*persistence.TaskInfo, error) {
		taskCompleted = true
	}

	task := newInternalTask(t.newTaskInfo(), completionFunc, types.TaskSourceDbBacklog, "", false)
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)

	var err error
	var remoteSyncMatch bool
	var req *types.AddDecisionTaskRequest
	t.client.EXPECT().AddDecisionTask(gomock.Any(), gomock.Any()).Return(errMatchingHostThrottle).Times(1)
	t.client.EXPECT().AddDecisionTask(gomock.Any(), gomock.Any()).Do(
		func(arg0 context.Context, arg1 *types.AddDecisionTaskRequest) {
			req = arg1
			task := newInternalTask(task.event.TaskInfo, nil, types.TaskSourceDbBacklog, req.GetForwardedFrom(), true)
			close(pollSigC)
			remoteSyncMatch, err = t.rootMatcher.Offer(ctx, task)
		},
	).Return(nil)

	// Poll needs to happen before MustOffer, or else it goes into the non-blocking path.
	ensureAsyncReady(time.Second, func(ctx context.Context) {
		task, err := t.matcher.Poll(ctx)
		t.Nil(err)
		t.NotNil(task)
	})

	t.NoError(t.matcher.MustOffer(ctx, task))
	cancel()
	t.NotNil(req)
	t.NoError(err)
	t.True(remoteSyncMatch)
	t.True(taskCompleted)
	t.Equal(t.taskList.name, req.GetForwardedFrom())
	t.Equal(t.taskList.Parent(20), req.GetTaskList().GetName())
}

func (t *MatcherTestSuite) TestRemotePoll() {
	pollToken := <-t.fwdr.PollReqTokenC()

	var req *types.MatchingPollForDecisionTaskRequest
	t.client.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any()).DoAndReturn(
		func(arg0 context.Context, arg1 *types.MatchingPollForDecisionTaskRequest) (*types.MatchingPollForDecisionTaskResponse, error) {
			req = arg1
			return nil, nil
		},
	)

	ready, wait := ensureAsyncAfterReady(0, func(_ context.Context) {
		pollToken.release()
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	ready()
	task, err := t.matcher.Poll(ctx)
	cancel()
	wait()
	t.NoError(err)
	t.NotNil(req)
	t.NotNil(task)
	t.True(task.isStarted())
}

func (t *MatcherTestSuite) TestRemotePollForQuery() {
	pollToken := <-t.fwdr.PollReqTokenC()

	var req *types.MatchingPollForDecisionTaskRequest
	t.client.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any()).DoAndReturn(
		func(arg0 context.Context, arg1 *types.MatchingPollForDecisionTaskRequest) (*types.MatchingPollForDecisionTaskResponse, error) {
			req = arg1
			return nil, nil
		},
	)

	ready, wait := ensureAsyncAfterReady(0, func(_ context.Context) {
		pollToken.release()
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ready()
	task, err := t.matcher.PollForQuery(ctx)
	wait()
	t.NoError(err)
	t.NotNil(req)
	t.NotNil(task)
	t.True(task.isStarted())
}

func (t *MatcherTestSuite) newDomainCache() cache.DomainCache {
	domainName := "test-domain"
	dc := cache.NewMockDomainCache(t.controller)
	dc.EXPECT().GetDomainName(gomock.Any()).Return(domainName, nil).AnyTimes()
	return dc
}

func (t *MatcherTestSuite) newTaskInfo() *persistence.TaskInfo {
	return &persistence.TaskInfo{
		DomainID:               uuid.New(),
		WorkflowID:             uuid.New(),
		RunID:                  uuid.New(),
		TaskID:                 rand.Int63(),
		ScheduleID:             rand.Int63(),
		ScheduleToStartTimeout: rand.Int31(),
	}
}

// Try to ensure a blocking callback in a goroutine is not running until the thing immediately
// after `ready()` has blocked, so tests can ensure that the callback contents happen last.
//
// Try to delay calling `ready()` until *immediately* before the blocking call for best results.
//
// This is a best-effort technique, as there is no way to reliably synchronize this kind of thing
// without exposing internal latches or having a more sophisticated locking library than Go offers.
// In case of flakiness, increase the time.Sleep and hope for the best.
//
// Note that adding fmt.Println() calls touches synchronization code (for I/O), so it may change behavior.
func ensureAsyncAfterReady(ctxTimeout time.Duration, cb func(ctx context.Context)) (ready func(), wait func()) {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	ready = func() {
		go func() {
			defer cancel()

			// since `go func()` is non-blocking, the ready()-ing goroutine should generally continue,
			// and read whatever blocking point is relevant before this goroutine runs.
			// in many cases this sleep is unnecessary (especially with -cpu=1), but it does help.
			time.Sleep(1 * time.Millisecond)

			cb(ctx)

		}()
	}
	wait = func() {
		<-ctx.Done()
	}
	return ready, wait
}

// Try to ensure a blocking callback is actively blocked in a goroutine before returning, so tests can
// ensure that the callback contents happen first.
//
// This is a best-effort technique, as there is no way to reliably synchronize this kind of thing
// without exposing internal latches or having a more sophisticated locking library than Go offers.
// In case of flakiness, increase the time.Sleep and hope for the best.
//
// Note that adding fmt.Println() calls touches synchronization code (for I/O), so it may change behavior.
func ensureAsyncReady(ctxTimeout time.Duration, cb func(ctx context.Context)) (wait func()) {
	running := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	go func() {
		defer cancel()

		close(running)
		cb(ctx)
	}()
	<-running // ensure the goroutine is alive

	// `close(running)` is non-blocking, so it should generally begin polling before yielding control to other goroutines,
	// but there is still a race to reach whatever blocking sync point exists between the code being tested.
	// In many cases this sleep is completely unnecessary (especially with -cpu=1), but it does help.
	time.Sleep(1 * time.Millisecond)

	return func() {
		<-ctx.Done()
	}
}
