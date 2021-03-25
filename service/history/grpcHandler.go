// Copyright (c) 2021 Uber Technologies Inc.
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

	"go.uber.org/yarpc"

	apiv1 "github.com/uber/cadence/.gen/proto/api/v1"
	historyv1 "github.com/uber/cadence/.gen/proto/history/v1"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types/mapper/proto"
)

type grpcHandler struct {
	h Handler
}

func newGRPCHandler(h Handler) grpcHandler {
	return grpcHandler{h}
}

func (g grpcHandler) register(dispatcher *yarpc.Dispatcher) {
	dispatcher.Register(historyv1.BuildHistoryAPIYARPCProcedures(g))
	dispatcher.Register(apiv1.BuildMetaAPIYARPCProcedures(g))
}

func (g grpcHandler) Health(ctx context.Context, _ *apiv1.HealthRequest) (*apiv1.HealthResponse, error) {
	response, err := g.h.Health(withGRPCTag(ctx))
	return proto.FromHealthResponse(response), proto.FromError(err)
}

func (g grpcHandler) CloseShard(ctx context.Context, request *historyv1.CloseShardRequest) (*historyv1.CloseShardResponse, error) {
	err := g.h.CloseShard(withGRPCTag(ctx), proto.ToHistoryCloseShardRequest(request))
	return &historyv1.CloseShardResponse{}, proto.FromError(err)
}

func (g grpcHandler) DescribeHistoryHost(ctx context.Context, request *historyv1.DescribeHistoryHostRequest) (*historyv1.DescribeHistoryHostResponse, error) {
	response, err := g.h.DescribeHistoryHost(withGRPCTag(ctx), proto.ToHistoryDescribeHistoryHostRequest(request))
	return proto.FromHistoryDescribeHistoryHostResponse(response), proto.FromError(err)
}

func (g grpcHandler) DescribeMutableState(ctx context.Context, request *historyv1.DescribeMutableStateRequest) (*historyv1.DescribeMutableStateResponse, error) {
	response, err := g.h.DescribeMutableState(withGRPCTag(ctx), proto.ToHistoryDescribeMutableStateRequest(request))
	return proto.FromHistoryDescribeMutableStateResponse(response), proto.FromError(err)
}

func (g grpcHandler) DescribeQueue(ctx context.Context, request *historyv1.DescribeQueueRequest) (*historyv1.DescribeQueueResponse, error) {
	response, err := g.h.DescribeQueue(withGRPCTag(ctx), proto.ToHistoryDescribeQueueRequest(request))
	return proto.FromHistoryDescribeQueueResponse(response), proto.FromError(err)
}

func (g grpcHandler) DescribeWorkflowExecution(ctx context.Context, request *historyv1.DescribeWorkflowExecutionRequest) (*historyv1.DescribeWorkflowExecutionResponse, error) {
	response, err := g.h.DescribeWorkflowExecution(withGRPCTag(ctx), proto.ToHistoryDescribeWorkflowExecutionRequest(request))
	return proto.FromHistoryDescribeWorkflowExecutionResponse(response), proto.FromError(err)
}

func (g grpcHandler) GetDLQReplicationMessages(ctx context.Context, request *historyv1.GetDLQReplicationMessagesRequest) (*historyv1.GetDLQReplicationMessagesResponse, error) {
	response, err := g.h.GetDLQReplicationMessages(withGRPCTag(ctx), proto.ToHistoryGetDLQReplicationMessagesRequest(request))
	return proto.FromHistoryGetDLQReplicationMessagesResponse(response), proto.FromError(err)
}

func (g grpcHandler) GetMutableState(ctx context.Context, request *historyv1.GetMutableStateRequest) (*historyv1.GetMutableStateResponse, error) {
	response, err := g.h.GetMutableState(withGRPCTag(ctx), proto.ToHistoryGetMutableStateRequest(request))
	return proto.FromHistoryGetMutableStateResponse(response), proto.FromError(err)
}

func (g grpcHandler) GetReplicationMessages(ctx context.Context, request *historyv1.GetReplicationMessagesRequest) (*historyv1.GetReplicationMessagesResponse, error) {
	response, err := g.h.GetReplicationMessages(withGRPCTag(ctx), proto.ToHistoryGetReplicationMessagesRequest(request))
	return proto.FromHistoryGetReplicationMessagesResponse(response), proto.FromError(err)
}

func (g grpcHandler) MergeDLQMessages(ctx context.Context, request *historyv1.MergeDLQMessagesRequest) (*historyv1.MergeDLQMessagesResponse, error) {
	response, err := g.h.MergeDLQMessages(withGRPCTag(ctx), proto.ToHistoryMergeDLQMessagesRequest(request))
	return proto.FromHistoryMergeDLQMessagesResponse(response), proto.FromError(err)
}

func (g grpcHandler) NotifyFailoverMarkers(ctx context.Context, request *historyv1.NotifyFailoverMarkersRequest) (*historyv1.NotifyFailoverMarkersResponse, error) {
	err := g.h.NotifyFailoverMarkers(withGRPCTag(ctx), proto.ToHistoryNotifyFailoverMarkersRequest(request))
	return &historyv1.NotifyFailoverMarkersResponse{}, proto.FromError(err)
}

func (g grpcHandler) PollMutableState(ctx context.Context, request *historyv1.PollMutableStateRequest) (*historyv1.PollMutableStateResponse, error) {
	response, err := g.h.PollMutableState(withGRPCTag(ctx), proto.ToHistoryPollMutableStateRequest(request))
	return proto.FromHistoryPollMutableStateResponse(response), proto.FromError(err)
}

func (g grpcHandler) PurgeDLQMessages(ctx context.Context, request *historyv1.PurgeDLQMessagesRequest) (*historyv1.PurgeDLQMessagesResponse, error) {
	err := g.h.PurgeDLQMessages(withGRPCTag(ctx), proto.ToHistoryPurgeDLQMessagesRequest(request))
	return &historyv1.PurgeDLQMessagesResponse{}, proto.FromError(err)
}

func (g grpcHandler) QueryWorkflow(ctx context.Context, request *historyv1.QueryWorkflowRequest) (*historyv1.QueryWorkflowResponse, error) {
	response, err := g.h.QueryWorkflow(withGRPCTag(ctx), proto.ToHistoryQueryWorkflowRequest(request))
	return proto.FromHistoryQueryWorkflowResponse(response), proto.FromError(err)
}

func (g grpcHandler) ReadDLQMessages(ctx context.Context, request *historyv1.ReadDLQMessagesRequest) (*historyv1.ReadDLQMessagesResponse, error) {
	response, err := g.h.ReadDLQMessages(withGRPCTag(ctx), proto.ToHistoryReadDLQMessagesRequest(request))
	return proto.FromHistoryReadDLQMessagesResponse(response), proto.FromError(err)
}

func (g grpcHandler) ReapplyEvents(ctx context.Context, request *historyv1.ReapplyEventsRequest) (*historyv1.ReapplyEventsResponse, error) {
	err := g.h.ReapplyEvents(withGRPCTag(ctx), proto.ToHistoryReapplyEventsRequest(request))
	return &historyv1.ReapplyEventsResponse{}, proto.FromError(err)
}

func (g grpcHandler) RecordActivityTaskHeartbeat(ctx context.Context, request *historyv1.RecordActivityTaskHeartbeatRequest) (*historyv1.RecordActivityTaskHeartbeatResponse, error) {
	response, err := g.h.RecordActivityTaskHeartbeat(withGRPCTag(ctx), proto.ToHistoryRecordActivityTaskHeartbeatRequest(request))
	return proto.FromHistoryRecordActivityTaskHeartbeatResponse(response), proto.FromError(err)
}

func (g grpcHandler) RecordActivityTaskStarted(ctx context.Context, request *historyv1.RecordActivityTaskStartedRequest) (*historyv1.RecordActivityTaskStartedResponse, error) {
	response, err := g.h.RecordActivityTaskStarted(withGRPCTag(ctx), proto.ToHistoryRecordActivityTaskStartedRequest(request))
	return proto.FromHistoryRecordActivityTaskStartedResponse(response), proto.FromError(err)
}

func (g grpcHandler) RecordChildExecutionCompleted(ctx context.Context, request *historyv1.RecordChildExecutionCompletedRequest) (*historyv1.RecordChildExecutionCompletedResponse, error) {
	err := g.h.RecordChildExecutionCompleted(withGRPCTag(ctx), proto.ToHistoryRecordChildExecutionCompletedRequest(request))
	return &historyv1.RecordChildExecutionCompletedResponse{}, proto.FromError(err)
}

func (g grpcHandler) RecordDecisionTaskStarted(ctx context.Context, request *historyv1.RecordDecisionTaskStartedRequest) (*historyv1.RecordDecisionTaskStartedResponse, error) {
	response, err := g.h.RecordDecisionTaskStarted(withGRPCTag(ctx), proto.ToHistoryRecordDecisionTaskStartedRequest(request))
	return proto.FromHistoryRecordDecisionTaskStartedResponse(response), proto.FromError(err)
}

func (g grpcHandler) RefreshWorkflowTasks(ctx context.Context, request *historyv1.RefreshWorkflowTasksRequest) (*historyv1.RefreshWorkflowTasksResponse, error) {
	err := g.h.RefreshWorkflowTasks(withGRPCTag(ctx), proto.ToHistoryRefreshWorkflowTasksRequest(request))
	return &historyv1.RefreshWorkflowTasksResponse{}, proto.FromError(err)
}

func (g grpcHandler) RemoveSignalMutableState(ctx context.Context, request *historyv1.RemoveSignalMutableStateRequest) (*historyv1.RemoveSignalMutableStateResponse, error) {
	err := g.h.RemoveSignalMutableState(withGRPCTag(ctx), proto.ToHistoryRemoveSignalMutableStateRequest(request))
	return &historyv1.RemoveSignalMutableStateResponse{}, proto.FromError(err)
}

func (g grpcHandler) RemoveTask(ctx context.Context, request *historyv1.RemoveTaskRequest) (*historyv1.RemoveTaskResponse, error) {
	err := g.h.RemoveTask(withGRPCTag(ctx), proto.ToHistoryRemoveTaskRequest(request))
	return &historyv1.RemoveTaskResponse{}, proto.FromError(err)
}

func (g grpcHandler) ReplicateEventsV2(ctx context.Context, request *historyv1.ReplicateEventsV2Request) (*historyv1.ReplicateEventsV2Response, error) {
	err := g.h.ReplicateEventsV2(withGRPCTag(ctx), proto.ToHistoryReplicateEventsV2Request(request))
	return &historyv1.ReplicateEventsV2Response{}, proto.FromError(err)
}

func (g grpcHandler) RequestCancelWorkflowExecution(ctx context.Context, request *historyv1.RequestCancelWorkflowExecutionRequest) (*historyv1.RequestCancelWorkflowExecutionResponse, error) {
	err := g.h.RequestCancelWorkflowExecution(withGRPCTag(ctx), proto.ToHistoryRequestCancelWorkflowExecutionRequest(request))
	return &historyv1.RequestCancelWorkflowExecutionResponse{}, proto.FromError(err)
}

func (g grpcHandler) ResetQueue(ctx context.Context, request *historyv1.ResetQueueRequest) (*historyv1.ResetQueueResponse, error) {
	err := g.h.ResetQueue(withGRPCTag(ctx), proto.ToHistoryResetQueueRequest(request))
	return &historyv1.ResetQueueResponse{}, proto.FromError(err)
}

func (g grpcHandler) ResetStickyTaskList(ctx context.Context, request *historyv1.ResetStickyTaskListRequest) (*historyv1.ResetStickyTaskListResponse, error) {
	_, err := g.h.ResetStickyTaskList(withGRPCTag(ctx), proto.ToHistoryResetStickyTaskListRequest(request))
	return &historyv1.ResetStickyTaskListResponse{}, proto.FromError(err)
}

func (g grpcHandler) ResetWorkflowExecution(ctx context.Context, request *historyv1.ResetWorkflowExecutionRequest) (*historyv1.ResetWorkflowExecutionResponse, error) {
	response, err := g.h.ResetWorkflowExecution(withGRPCTag(ctx), proto.ToHistoryResetWorkflowExecutionRequest(request))
	return proto.FromHistoryResetWorkflowExecutionResponse(response), proto.FromError(err)
}

func (g grpcHandler) RespondActivityTaskCanceled(ctx context.Context, request *historyv1.RespondActivityTaskCanceledRequest) (*historyv1.RespondActivityTaskCanceledResponse, error) {
	err := g.h.RespondActivityTaskCanceled(withGRPCTag(ctx), proto.ToHistoryRespondActivityTaskCanceledRequest(request))
	return &historyv1.RespondActivityTaskCanceledResponse{}, proto.FromError(err)
}

func (g grpcHandler) RespondActivityTaskCompleted(ctx context.Context, request *historyv1.RespondActivityTaskCompletedRequest) (*historyv1.RespondActivityTaskCompletedResponse, error) {
	err := g.h.RespondActivityTaskCompleted(withGRPCTag(ctx), proto.ToHistoryRespondActivityTaskCompletedRequest(request))
	return &historyv1.RespondActivityTaskCompletedResponse{}, proto.FromError(err)
}

func (g grpcHandler) RespondActivityTaskFailed(ctx context.Context, request *historyv1.RespondActivityTaskFailedRequest) (*historyv1.RespondActivityTaskFailedResponse, error) {
	err := g.h.RespondActivityTaskFailed(withGRPCTag(ctx), proto.ToHistoryRespondActivityTaskFailedRequest(request))
	return &historyv1.RespondActivityTaskFailedResponse{}, proto.FromError(err)
}

func (g grpcHandler) RespondDecisionTaskCompleted(ctx context.Context, request *historyv1.RespondDecisionTaskCompletedRequest) (*historyv1.RespondDecisionTaskCompletedResponse, error) {
	response, err := g.h.RespondDecisionTaskCompleted(withGRPCTag(ctx), proto.ToHistoryRespondDecisionTaskCompletedRequest(request))
	return proto.FromHistoryRespondDecisionTaskCompletedResponse(response), proto.FromError(err)
}

func (g grpcHandler) RespondDecisionTaskFailed(ctx context.Context, request *historyv1.RespondDecisionTaskFailedRequest) (*historyv1.RespondDecisionTaskFailedResponse, error) {
	err := g.h.RespondDecisionTaskFailed(withGRPCTag(ctx), proto.ToHistoryRespondDecisionTaskFailedRequest(request))
	return &historyv1.RespondDecisionTaskFailedResponse{}, proto.FromError(err)
}

func (g grpcHandler) ScheduleDecisionTask(ctx context.Context, request *historyv1.ScheduleDecisionTaskRequest) (*historyv1.ScheduleDecisionTaskResponse, error) {
	err := g.h.ScheduleDecisionTask(withGRPCTag(ctx), proto.ToHistoryScheduleDecisionTaskRequest(request))
	return &historyv1.ScheduleDecisionTaskResponse{}, proto.FromError(err)
}

func (g grpcHandler) SignalWithStartWorkflowExecution(ctx context.Context, request *historyv1.SignalWithStartWorkflowExecutionRequest) (*historyv1.SignalWithStartWorkflowExecutionResponse, error) {
	response, err := g.h.SignalWithStartWorkflowExecution(withGRPCTag(ctx), proto.ToHistorySignalWithStartWorkflowExecutionRequest(request))
	return proto.FromHistorySignalWithStartWorkflowExecutionResponse(response), proto.FromError(err)
}

func (g grpcHandler) SignalWorkflowExecution(ctx context.Context, request *historyv1.SignalWorkflowExecutionRequest) (*historyv1.SignalWorkflowExecutionResponse, error) {
	err := g.h.SignalWorkflowExecution(withGRPCTag(ctx), proto.ToHistorySignalWorkflowExecutionRequest(request))
	return &historyv1.SignalWorkflowExecutionResponse{}, proto.FromError(err)
}

func (g grpcHandler) StartWorkflowExecution(ctx context.Context, request *historyv1.StartWorkflowExecutionRequest) (*historyv1.StartWorkflowExecutionResponse, error) {
	response, err := g.h.StartWorkflowExecution(withGRPCTag(ctx), proto.ToHistoryStartWorkflowExecutionRequest(request))
	return proto.FromHistoryStartWorkflowExecutionResponse(response), proto.FromError(err)
}

func (g grpcHandler) SyncActivity(ctx context.Context, request *historyv1.SyncActivityRequest) (*historyv1.SyncActivityResponse, error) {
	err := g.h.SyncActivity(withGRPCTag(ctx), proto.ToHistorySyncActivityRequest(request))
	return &historyv1.SyncActivityResponse{}, proto.FromError(err)
}

func (g grpcHandler) SyncShardStatus(ctx context.Context, request *historyv1.SyncShardStatusRequest) (*historyv1.SyncShardStatusResponse, error) {
	err := g.h.SyncShardStatus(withGRPCTag(ctx), proto.ToHistorySyncShardStatusRequest(request))
	return &historyv1.SyncShardStatusResponse{}, proto.FromError(err)
}

func (g grpcHandler) TerminateWorkflowExecution(ctx context.Context, request *historyv1.TerminateWorkflowExecutionRequest) (*historyv1.TerminateWorkflowExecutionResponse, error) {
	err := g.h.TerminateWorkflowExecution(withGRPCTag(ctx), proto.ToHistoryTerminateWorkflowExecutionRequest(request))
	return &historyv1.TerminateWorkflowExecutionResponse{}, proto.FromError(err)
}

func withGRPCTag(ctx context.Context) context.Context {
	return metrics.TagContext(ctx, metrics.GPRCTransportTag())
}
