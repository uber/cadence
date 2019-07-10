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
	"errors"
	"sync/atomic"

	gen "github.com/uber/cadence/.gen/go/matching"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/client/matching"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/quotas"
)

type (
	// Forwarder is the type that contains state pertaining to
	// the api call forwarder component
	Forwarder struct {
		cfg          *forwarderConfig
		taskListID   *taskListID
		taskListKind shared.TaskListKind
		client       matching.Client
		taskToken    forwarderToken
		pollToken    forwarderToken
		scope        struct {
			forwardTask  metrics.Scope
			forwardQuery metrics.Scope
			forwardPoll  metrics.Scope
		}
		scopeFunc func() metrics.Scope
	}
	// forwarderToken contains the state related to rate limiting
	// of forwarded calls
	forwarderToken struct {
		nOutstanding   int32
		maxOutstanding func() int
		rpsFunc        func() float64
		limiter        *quotas.RateLimiter
	}
)

var (
	errForwarderRateLimit  = errors.New("limit exceeded")
	errNoParent            = errors.New("cannot find parent task list for forwarding")
	errTaskListKind        = errors.New("forwarding is not supported on sticky task list")
	errInvalidTaskListType = errors.New("unrecognized task list type")
)

// newForwarder returns an instance of Forwarder object which
// can be used to forward api request calls from a task list
// child partition to a task list parent partition. The returned
// forwarder is tied to a single task list. All of the exposed
// methods can return the following errors:
// Returns following errors:
//  - errNoParent: If this task list doesn't have a parent to forward to
//  - errTaskListKind: If the task list is a sticky task list. Sticky task lists are never partitioned
//  - errForwarderRateLimit: When the rate limit is exceeded
//  - errInvalidTaskType: If the task list type is invalid
func newForwarder(
	cfg *forwarderConfig,
	taskListID *taskListID,
	kind shared.TaskListKind,
	client matching.Client,
	scopeFunc func() metrics.Scope,
) *Forwarder {
	rps := func() float64 { return float64(cfg.ForwarderMaxRatePerSecond()) }
	return &Forwarder{
		cfg:          cfg,
		client:       client,
		taskListID:   taskListID,
		taskListKind: kind,
		taskToken:    newForwarderToken(cfg.ForwarderMaxOutstandingTasks, rps),
		pollToken:    newForwarderToken(cfg.ForwarderMaxOutstandingPolls, nil),
		scopeFunc:    scopeFunc,
	}
}

// ForwardTask forwards an activity or decision task to the parent task list partition if it exist
func (fwdr *Forwarder) ForwardTask(ctx context.Context, task *internalTask) error {
	if fwdr.taskListKind == shared.TaskListKindSticky {
		return errTaskListKind
	}

	name := fwdr.taskListID.Parent(fwdr.cfg.ForwarderMaxChildrenPerNode())
	if name == "" {
		return errNoParent
	}

	release, err := fwdr.acquireToken(
		&fwdr.taskToken, metrics.ForwardTaskCalls, metrics.ForwardTaskErrors, metrics.ForwardTaskLatency)
	if err != nil {
		return err
	}
	defer release()

	switch fwdr.taskListID.taskType {
	case persistence.TaskListTypeDecision:
		return fwdr.client.AddDecisionTask(ctx, &gen.AddDecisionTaskRequest{
			DomainUUID: &task.event.DomainID,
			Execution:  task.workflowExecution(),
			TaskList: &shared.TaskList{
				Name: &name,
				Kind: &fwdr.taskListKind,
			},
			ScheduleId:                    &task.event.ScheduleID,
			ScheduleToStartTimeoutSeconds: &task.event.ScheduleToStartTimeout,
			ForwardedFrom:                 &fwdr.taskListID.name,
		})
	case persistence.TaskListTypeActivity:
		return fwdr.client.AddActivityTask(ctx, &gen.AddActivityTaskRequest{
			DomainUUID: &task.event.DomainID,
			Execution:  task.workflowExecution(),
			TaskList: &shared.TaskList{
				Name: &name,
				Kind: &fwdr.taskListKind,
			},
			ScheduleId:                    &task.event.ScheduleID,
			ScheduleToStartTimeoutSeconds: &task.event.ScheduleToStartTimeout,
			ForwardedFrom:                 &fwdr.taskListID.name,
		})
	}
	return errInvalidTaskListType
}

// ForwardQueryTask forwards a query task to parent task list partition, if it exist
func (fwdr *Forwarder) ForwardQueryTask(
	ctx context.Context,
	task *internalTask,
) (*shared.QueryWorkflowResponse, error) {

	if fwdr.taskListKind == shared.TaskListKindSticky {
		return nil, errTaskListKind
	}

	name := fwdr.taskListID.Parent(fwdr.cfg.ForwarderMaxChildrenPerNode())
	if name == "" {
		return nil, errNoParent
	}

	release, err := fwdr.acquireToken(
		&fwdr.taskToken, metrics.ForwardQueryCalls, metrics.ForwardQueryErrors, metrics.ForwardQueryLatency)
	if err != nil {
		return nil, err
	}
	defer release()

	return fwdr.client.QueryWorkflow(ctx, &gen.QueryWorkflowRequest{
		DomainUUID: task.query.request.DomainUUID,
		TaskList: &shared.TaskList{
			Name: &name,
			Kind: &fwdr.taskListKind,
		},
		QueryRequest:  task.query.request.QueryRequest,
		ForwardedFrom: &fwdr.taskListID.name,
	})
}

// ForwardPoll forwards a poll request to parent task list partition if it exist
func (fwdr *Forwarder) ForwardPoll(ctx context.Context) (*internalTask, error) {
	if fwdr.taskListKind == shared.TaskListKindSticky {
		return nil, errTaskListKind
	}

	name := fwdr.taskListID.Parent(fwdr.cfg.ForwarderMaxChildrenPerNode())
	if name == "" {
		return nil, errNoParent
	}

	release, err := fwdr.acquireToken(
		&fwdr.pollToken, metrics.ForwardPollCalls, metrics.ForwardPollErrors, metrics.ForwardPollLatency)
	if err != nil {
		return nil, err
	}
	defer release()

	pollerID, _ := ctx.Value(pollerIDKey).(string)
	identity, _ := ctx.Value(identityKey).(string)

	switch fwdr.taskListID.taskType {
	case persistence.TaskListTypeDecision:
		resp, err := fwdr.client.PollForDecisionTask(ctx, &gen.PollForDecisionTaskRequest{
			DomainUUID: &fwdr.taskListID.domainID,
			PollerID:   &pollerID,
			PollRequest: &shared.PollForDecisionTaskRequest{
				TaskList: &shared.TaskList{
					Name: &name,
					Kind: &fwdr.taskListKind,
				},
				Identity: &identity,
			},
			ForwardedFrom: &fwdr.taskListID.name,
		})
		if err != nil {
			return nil, err
		}
		return newInternalStartedTask(&startedTaskInfo{decisionTaskInfo: resp}), nil
	case persistence.TaskListTypeActivity:
		resp, err := fwdr.client.PollForActivityTask(ctx, &gen.PollForActivityTaskRequest{
			DomainUUID: &fwdr.taskListID.domainID,
			PollerID:   &pollerID,
			PollRequest: &shared.PollForActivityTaskRequest{
				TaskList: &shared.TaskList{
					Name: &name,
					Kind: &fwdr.taskListKind,
				},
				Identity: &identity,
			},
			ForwardedFrom: &fwdr.taskListID.name,
		})
		if err != nil {
			return nil, err
		}
		return newInternalStartedTask(&startedTaskInfo{activityTaskInfo: resp}), nil
	}

	return nil, errInvalidTaskListType
}

func (fwdr *Forwarder) acquireToken(token *forwarderToken, calls int, errs int, latency int) (func(), error) {
	fwdr.scopeFunc().IncCounter(calls)
	if !token.acquire() {
		fwdr.scopeFunc().IncCounter(errs)
		return nil, errForwarderRateLimit
	}
	sw := fwdr.scopeFunc().StartTimer(latency)
	return func() {
		token.release()
		sw.Stop()
	}, nil
}

func newForwarderToken(maxOutstanding func() int, rpsFunc func() float64) forwarderToken {
	var limiter *quotas.RateLimiter
	if rpsFunc != nil {
		rate := rpsFunc()
		limiter = quotas.NewRateLimiter(&rate, _defaultTaskDispatchRPSTTL, 1)
	}
	return forwarderToken{
		maxOutstanding: maxOutstanding,
		limiter:        limiter,
		rpsFunc:        rpsFunc,
	}
}

func (token *forwarderToken) acquire() bool {
	curr := atomic.LoadInt32(&token.nOutstanding)
	for {
		if curr >= int32(token.maxOutstanding()) {
			return false
		}
		if atomic.CompareAndSwapInt32(&token.nOutstanding, curr, curr+1) {
			if !token.rateLimit() {
				token.release()
				return false
			}
			return true
		}
		curr = atomic.LoadInt32(&token.nOutstanding)
	}
}

func (token *forwarderToken) rateLimit() bool {
	if token.limiter == nil {
		return true
	}
	if token.rpsFunc != nil {
		rate := token.rpsFunc()
		token.limiter.UpdateMaxDispatch(&rate)
	}
	return token.limiter.Allow()
}

func (token *forwarderToken) release() {
	atomic.AddInt32(&token.nOutstanding, -1)
}
