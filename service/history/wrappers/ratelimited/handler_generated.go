// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package ratelimited

// Code generated by gowrap. DO NOT EDIT.
// template: ../../templates/ratelimited.tmpl
// gowrap: http://github.com/hexdigest/gowrap

//go:generate gowrap gen -p github.com/uber/cadence/service/history/handler -i Handler -t ../../templates/ratelimited.tmpl -o handler_generated.go -v handler=History -l ""

import (
	"context"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/frontend/validate"
	"github.com/uber/cadence/service/history/handler"
	"github.com/uber/cadence/service/history/workflowcache"
)

// historyHandler implements handler.Handler interface instrumented with rate limiter.
type historyHandler struct {
	wrapped                        handler.Handler
	workflowIDCache                workflowcache.WFCache
	ratelimitExternalPerWorkflowID dynamicconfig.BoolPropertyFnWithDomainFilter
	domainCache                    cache.DomainCache
	logger                         log.Logger
	allowFunc                      func(domainID string, workflowID string) bool
}

// NewHistoryHandler creates a new instance of Handler with ratelimiter.
func NewHistoryHandler(
	wrapped handler.Handler,
	workflowIDCache workflowcache.WFCache,
	ratelimitExternalPerWorkflowID dynamicconfig.BoolPropertyFnWithDomainFilter,
	domainCache cache.DomainCache,
	logger log.Logger,
) handler.Handler {
	wrapper := &historyHandler{
		wrapped:                        wrapped,
		workflowIDCache:                workflowIDCache,
		ratelimitExternalPerWorkflowID: ratelimitExternalPerWorkflowID,
		domainCache:                    domainCache,
		logger:                         logger,
	}
	wrapper.allowFunc = wrapper.allowWfID

	return wrapper
}

func (h *historyHandler) CloseShard(ctx context.Context, cp1 *types.CloseShardRequest) (err error) {
	return h.wrapped.CloseShard(ctx, cp1)
}

func (h *historyHandler) CountDLQMessages(ctx context.Context, cp1 *types.CountDLQMessagesRequest) (hp1 *types.HistoryCountDLQMessagesResponse, err error) {
	return h.wrapped.CountDLQMessages(ctx, cp1)
}

func (h *historyHandler) DescribeHistoryHost(ctx context.Context, dp1 *types.DescribeHistoryHostRequest) (dp2 *types.DescribeHistoryHostResponse, err error) {
	return h.wrapped.DescribeHistoryHost(ctx, dp1)
}

func (h *historyHandler) DescribeMutableState(ctx context.Context, dp1 *types.DescribeMutableStateRequest) (dp2 *types.DescribeMutableStateResponse, err error) {
	return h.wrapped.DescribeMutableState(ctx, dp1)
}

func (h *historyHandler) DescribeQueue(ctx context.Context, dp1 *types.DescribeQueueRequest) (dp2 *types.DescribeQueueResponse, err error) {
	return h.wrapped.DescribeQueue(ctx, dp1)
}

func (h *historyHandler) DescribeWorkflowExecution(ctx context.Context, hp1 *types.HistoryDescribeWorkflowExecutionRequest) (dp1 *types.DescribeWorkflowExecutionResponse, err error) {

	if hp1 == nil {
		err = validate.ErrRequestNotSet
		return
	}

	if hp1.GetDomainUUID() == "" {
		err = validate.ErrDomainNotSet
		return
	}

	if hp1.Request.GetExecution().GetWorkflowID() == "" {
		err = validate.ErrWorkflowIDNotSet
		return
	}

	if !h.allowFunc(hp1.GetDomainUUID(), hp1.Request.GetExecution().GetWorkflowID()) {
		err = &types.ServiceBusyError{
			Message: "Too many requests for the workflow ID",
			Reason:  common.WorkflowIDRateLimitReason,
		}
		return
	}
	return h.wrapped.DescribeWorkflowExecution(ctx, hp1)
}

func (h *historyHandler) GetCrossClusterTasks(ctx context.Context, gp1 *types.GetCrossClusterTasksRequest) (gp2 *types.GetCrossClusterTasksResponse, err error) {
	return h.wrapped.GetCrossClusterTasks(ctx, gp1)
}

func (h *historyHandler) GetDLQReplicationMessages(ctx context.Context, gp1 *types.GetDLQReplicationMessagesRequest) (gp2 *types.GetDLQReplicationMessagesResponse, err error) {
	return h.wrapped.GetDLQReplicationMessages(ctx, gp1)
}

func (h *historyHandler) GetFailoverInfo(ctx context.Context, gp1 *types.GetFailoverInfoRequest) (gp2 *types.GetFailoverInfoResponse, err error) {
	return h.wrapped.GetFailoverInfo(ctx, gp1)
}

func (h *historyHandler) GetMutableState(ctx context.Context, gp1 *types.GetMutableStateRequest) (gp2 *types.GetMutableStateResponse, err error) {
	return h.wrapped.GetMutableState(ctx, gp1)
}

func (h *historyHandler) GetReplicationMessages(ctx context.Context, gp1 *types.GetReplicationMessagesRequest) (gp2 *types.GetReplicationMessagesResponse, err error) {
	return h.wrapped.GetReplicationMessages(ctx, gp1)
}

func (h *historyHandler) Health(ctx context.Context) (hp1 *types.HealthStatus, err error) {
	return h.wrapped.Health(ctx)
}

func (h *historyHandler) MergeDLQMessages(ctx context.Context, mp1 *types.MergeDLQMessagesRequest) (mp2 *types.MergeDLQMessagesResponse, err error) {
	return h.wrapped.MergeDLQMessages(ctx, mp1)
}

func (h *historyHandler) NotifyFailoverMarkers(ctx context.Context, np1 *types.NotifyFailoverMarkersRequest) (err error) {
	return h.wrapped.NotifyFailoverMarkers(ctx, np1)
}

func (h *historyHandler) PollMutableState(ctx context.Context, pp1 *types.PollMutableStateRequest) (pp2 *types.PollMutableStateResponse, err error) {
	return h.wrapped.PollMutableState(ctx, pp1)
}

func (h *historyHandler) PrepareToStop(d1 time.Duration) (d2 time.Duration) {
	return h.wrapped.PrepareToStop(d1)
}

func (h *historyHandler) PurgeDLQMessages(ctx context.Context, pp1 *types.PurgeDLQMessagesRequest) (err error) {
	return h.wrapped.PurgeDLQMessages(ctx, pp1)
}

func (h *historyHandler) QueryWorkflow(ctx context.Context, hp1 *types.HistoryQueryWorkflowRequest) (hp2 *types.HistoryQueryWorkflowResponse, err error) {
	return h.wrapped.QueryWorkflow(ctx, hp1)
}

func (h *historyHandler) ReadDLQMessages(ctx context.Context, rp1 *types.ReadDLQMessagesRequest) (rp2 *types.ReadDLQMessagesResponse, err error) {
	return h.wrapped.ReadDLQMessages(ctx, rp1)
}

func (h *historyHandler) ReapplyEvents(ctx context.Context, hp1 *types.HistoryReapplyEventsRequest) (err error) {
	return h.wrapped.ReapplyEvents(ctx, hp1)
}

func (h *historyHandler) RecordActivityTaskHeartbeat(ctx context.Context, hp1 *types.HistoryRecordActivityTaskHeartbeatRequest) (rp1 *types.RecordActivityTaskHeartbeatResponse, err error) {
	return h.wrapped.RecordActivityTaskHeartbeat(ctx, hp1)
}

func (h *historyHandler) RecordActivityTaskStarted(ctx context.Context, rp1 *types.RecordActivityTaskStartedRequest) (rp2 *types.RecordActivityTaskStartedResponse, err error) {
	return h.wrapped.RecordActivityTaskStarted(ctx, rp1)
}

func (h *historyHandler) RecordChildExecutionCompleted(ctx context.Context, rp1 *types.RecordChildExecutionCompletedRequest) (err error) {
	return h.wrapped.RecordChildExecutionCompleted(ctx, rp1)
}

func (h *historyHandler) RecordDecisionTaskStarted(ctx context.Context, rp1 *types.RecordDecisionTaskStartedRequest) (rp2 *types.RecordDecisionTaskStartedResponse, err error) {
	return h.wrapped.RecordDecisionTaskStarted(ctx, rp1)
}

func (h *historyHandler) RefreshWorkflowTasks(ctx context.Context, hp1 *types.HistoryRefreshWorkflowTasksRequest) (err error) {
	return h.wrapped.RefreshWorkflowTasks(ctx, hp1)
}

func (h *historyHandler) RemoveSignalMutableState(ctx context.Context, rp1 *types.RemoveSignalMutableStateRequest) (err error) {
	return h.wrapped.RemoveSignalMutableState(ctx, rp1)
}

func (h *historyHandler) RemoveTask(ctx context.Context, rp1 *types.RemoveTaskRequest) (err error) {
	return h.wrapped.RemoveTask(ctx, rp1)
}

func (h *historyHandler) ReplicateEventsV2(ctx context.Context, rp1 *types.ReplicateEventsV2Request) (err error) {
	return h.wrapped.ReplicateEventsV2(ctx, rp1)
}

func (h *historyHandler) RequestCancelWorkflowExecution(ctx context.Context, hp1 *types.HistoryRequestCancelWorkflowExecutionRequest) (err error) {
	return h.wrapped.RequestCancelWorkflowExecution(ctx, hp1)
}

func (h *historyHandler) ResetQueue(ctx context.Context, rp1 *types.ResetQueueRequest) (err error) {
	return h.wrapped.ResetQueue(ctx, rp1)
}

func (h *historyHandler) ResetStickyTaskList(ctx context.Context, hp1 *types.HistoryResetStickyTaskListRequest) (hp2 *types.HistoryResetStickyTaskListResponse, err error) {
	return h.wrapped.ResetStickyTaskList(ctx, hp1)
}

func (h *historyHandler) ResetWorkflowExecution(ctx context.Context, hp1 *types.HistoryResetWorkflowExecutionRequest) (rp1 *types.ResetWorkflowExecutionResponse, err error) {
	return h.wrapped.ResetWorkflowExecution(ctx, hp1)
}

func (h *historyHandler) RespondActivityTaskCanceled(ctx context.Context, hp1 *types.HistoryRespondActivityTaskCanceledRequest) (err error) {
	return h.wrapped.RespondActivityTaskCanceled(ctx, hp1)
}

func (h *historyHandler) RespondActivityTaskCompleted(ctx context.Context, hp1 *types.HistoryRespondActivityTaskCompletedRequest) (err error) {
	return h.wrapped.RespondActivityTaskCompleted(ctx, hp1)
}

func (h *historyHandler) RespondActivityTaskFailed(ctx context.Context, hp1 *types.HistoryRespondActivityTaskFailedRequest) (err error) {
	return h.wrapped.RespondActivityTaskFailed(ctx, hp1)
}

func (h *historyHandler) RespondCrossClusterTasksCompleted(ctx context.Context, rp1 *types.RespondCrossClusterTasksCompletedRequest) (rp2 *types.RespondCrossClusterTasksCompletedResponse, err error) {
	return h.wrapped.RespondCrossClusterTasksCompleted(ctx, rp1)
}

func (h *historyHandler) RespondDecisionTaskCompleted(ctx context.Context, hp1 *types.HistoryRespondDecisionTaskCompletedRequest) (hp2 *types.HistoryRespondDecisionTaskCompletedResponse, err error) {
	return h.wrapped.RespondDecisionTaskCompleted(ctx, hp1)
}

func (h *historyHandler) RespondDecisionTaskFailed(ctx context.Context, hp1 *types.HistoryRespondDecisionTaskFailedRequest) (err error) {
	return h.wrapped.RespondDecisionTaskFailed(ctx, hp1)
}

func (h *historyHandler) ScheduleDecisionTask(ctx context.Context, sp1 *types.ScheduleDecisionTaskRequest) (err error) {
	return h.wrapped.ScheduleDecisionTask(ctx, sp1)
}

func (h *historyHandler) SignalWithStartWorkflowExecution(ctx context.Context, hp1 *types.HistorySignalWithStartWorkflowExecutionRequest) (sp1 *types.StartWorkflowExecutionResponse, err error) {

	if hp1 == nil {
		err = validate.ErrRequestNotSet
		return
	}

	if hp1.GetDomainUUID() == "" {
		err = validate.ErrDomainNotSet
		return
	}

	if hp1.SignalWithStartRequest.GetWorkflowID() == "" {
		err = validate.ErrWorkflowIDNotSet
		return
	}

	if !h.allowFunc(hp1.GetDomainUUID(), hp1.SignalWithStartRequest.GetWorkflowID()) {
		err = &types.ServiceBusyError{
			Message: "Too many requests for the workflow ID",
			Reason:  common.WorkflowIDRateLimitReason,
		}
		return
	}
	return h.wrapped.SignalWithStartWorkflowExecution(ctx, hp1)
}

func (h *historyHandler) SignalWorkflowExecution(ctx context.Context, hp1 *types.HistorySignalWorkflowExecutionRequest) (err error) {

	if hp1 == nil {
		err = validate.ErrRequestNotSet
		return
	}

	if hp1.GetDomainUUID() == "" {
		err = validate.ErrDomainNotSet
		return
	}

	if hp1.SignalRequest.GetWorkflowExecution().GetWorkflowID() == "" {
		err = validate.ErrWorkflowIDNotSet
		return
	}

	if !h.allowFunc(hp1.GetDomainUUID(), hp1.SignalRequest.GetWorkflowExecution().GetWorkflowID()) {
		err = &types.ServiceBusyError{
			Message: "Too many requests for the workflow ID",
			Reason:  common.WorkflowIDRateLimitReason,
		}
		return
	}
	return h.wrapped.SignalWorkflowExecution(ctx, hp1)
}

func (h *historyHandler) Start() {
	h.wrapped.Start()
	return
}

func (h *historyHandler) StartWorkflowExecution(ctx context.Context, hp1 *types.HistoryStartWorkflowExecutionRequest) (sp1 *types.StartWorkflowExecutionResponse, err error) {

	if hp1 == nil {
		err = validate.ErrRequestNotSet
		return
	}

	if hp1.GetDomainUUID() == "" {
		err = validate.ErrDomainNotSet
		return
	}

	if hp1.StartRequest.GetWorkflowID() == "" {
		err = validate.ErrWorkflowIDNotSet
		return
	}

	if !h.allowFunc(hp1.GetDomainUUID(), hp1.StartRequest.GetWorkflowID()) {
		err = &types.ServiceBusyError{
			Message: "Too many requests for the workflow ID",
			Reason:  common.WorkflowIDRateLimitReason,
		}
		return
	}
	return h.wrapped.StartWorkflowExecution(ctx, hp1)
}

func (h *historyHandler) Stop() {
	h.wrapped.Stop()
	return
}

func (h *historyHandler) SyncActivity(ctx context.Context, sp1 *types.SyncActivityRequest) (err error) {
	return h.wrapped.SyncActivity(ctx, sp1)
}

func (h *historyHandler) SyncShardStatus(ctx context.Context, sp1 *types.SyncShardStatusRequest) (err error) {
	return h.wrapped.SyncShardStatus(ctx, sp1)
}

func (h *historyHandler) TerminateWorkflowExecution(ctx context.Context, hp1 *types.HistoryTerminateWorkflowExecutionRequest) (err error) {
	return h.wrapped.TerminateWorkflowExecution(ctx, hp1)
}
