// Copyright (c) 2021 Uber Technologies, Inc.
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

	historyv1 "github.com/uber/cadence/.gen/proto/history/v1"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/proto"
)

type grpcClient struct {
	c historyv1.HistoryAPIYARPCClient
}

func NewGRPCClient(c historyv1.HistoryAPIYARPCClient) Client {
	return grpcClient{c}
}

func (g grpcClient) CloseShard(ctx context.Context, request *types.CloseShardRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.CloseShard(ctx, proto.FromHistoryCloseShardRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) DescribeHistoryHost(ctx context.Context, request *types.DescribeHistoryHostRequest, opts ...yarpc.CallOption) (*types.DescribeHistoryHostResponse, error) {
	response, err := g.c.DescribeHistoryHost(ctx, proto.FromHistoryDescribeHistoryHostRequest(request), opts...)
	return proto.ToHistoryDescribeHistoryHostResponse(response), proto.ToError(err)
}

func (g grpcClient) DescribeMutableState(ctx context.Context, request *types.DescribeMutableStateRequest, opts ...yarpc.CallOption) (*types.DescribeMutableStateResponse, error) {
	response, err := g.c.DescribeMutableState(ctx, proto.FromHistoryDescribeMutableStateRequest(request), opts...)
	return proto.ToHistoryDescribeMutableStateResponse(response), proto.ToError(err)
}

func (g grpcClient) DescribeQueue(ctx context.Context, request *types.DescribeQueueRequest, opts ...yarpc.CallOption) (*types.DescribeQueueResponse, error) {
	response, err := g.c.DescribeQueue(ctx, proto.FromHistoryDescribeQueueRequest(request), opts...)
	return proto.ToHistoryDescribeQueueResponse(response), proto.ToError(err)
}

func (g grpcClient) DescribeWorkflowExecution(ctx context.Context, request *types.HistoryDescribeWorkflowExecutionRequest, opts ...yarpc.CallOption) (*types.DescribeWorkflowExecutionResponse, error) {
	response, err := g.c.DescribeWorkflowExecution(ctx, proto.FromHistoryDescribeWorkflowExecutionRequest(request), opts...)
	return proto.ToHistoryDescribeWorkflowExecutionResponse(response), proto.ToError(err)
}

func (g grpcClient) GetCrossClusterTasks(ctx context.Context, request *types.GetCrossClusterTasksRequest, opts ...yarpc.CallOption) (*types.GetCrossClusterTasksResponse, error) {
	response, err := g.c.GetCrossClusterTasks(ctx, proto.FromHistoryGetCrossClusterTasksRequest(request), opts...)
	return proto.ToHistoryGetCrossClusterTasksResponse(response), proto.ToError(err)
}

func (g grpcClient) GetDLQReplicationMessages(ctx context.Context, request *types.GetDLQReplicationMessagesRequest, opts ...yarpc.CallOption) (*types.GetDLQReplicationMessagesResponse, error) {
	response, err := g.c.GetDLQReplicationMessages(ctx, proto.FromHistoryGetDLQReplicationMessagesRequest(request), opts...)
	return proto.ToHistoryGetDLQReplicationMessagesResponse(response), proto.ToError(err)
}

func (g grpcClient) GetMutableState(ctx context.Context, request *types.GetMutableStateRequest, opts ...yarpc.CallOption) (*types.GetMutableStateResponse, error) {
	response, err := g.c.GetMutableState(ctx, proto.FromHistoryGetMutableStateRequest(request), opts...)
	return proto.ToHistoryGetMutableStateResponse(response), proto.ToError(err)
}

func (g grpcClient) GetReplicationMessages(ctx context.Context, request *types.GetReplicationMessagesRequest, opts ...yarpc.CallOption) (*types.GetReplicationMessagesResponse, error) {
	response, err := g.c.GetReplicationMessages(ctx, proto.FromHistoryGetReplicationMessagesRequest(request), opts...)
	return proto.ToHistoryGetReplicationMessagesResponse(response), proto.ToError(err)
}

func (g grpcClient) MergeDLQMessages(ctx context.Context, request *types.MergeDLQMessagesRequest, opts ...yarpc.CallOption) (*types.MergeDLQMessagesResponse, error) {
	response, err := g.c.MergeDLQMessages(ctx, proto.FromHistoryMergeDLQMessagesRequest(request), opts...)
	return proto.ToHistoryMergeDLQMessagesResponse(response), proto.ToError(err)
}

func (g grpcClient) NotifyFailoverMarkers(ctx context.Context, request *types.NotifyFailoverMarkersRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.NotifyFailoverMarkers(ctx, proto.FromHistoryNotifyFailoverMarkersRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) PollMutableState(ctx context.Context, request *types.PollMutableStateRequest, opts ...yarpc.CallOption) (*types.PollMutableStateResponse, error) {
	response, err := g.c.PollMutableState(ctx, proto.FromHistoryPollMutableStateRequest(request), opts...)
	return proto.ToHistoryPollMutableStateResponse(response), proto.ToError(err)
}

func (g grpcClient) PurgeDLQMessages(ctx context.Context, request *types.PurgeDLQMessagesRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.PurgeDLQMessages(ctx, proto.FromHistoryPurgeDLQMessagesRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) QueryWorkflow(ctx context.Context, request *types.HistoryQueryWorkflowRequest, opts ...yarpc.CallOption) (*types.HistoryQueryWorkflowResponse, error) {
	response, err := g.c.QueryWorkflow(ctx, proto.FromHistoryQueryWorkflowRequest(request), opts...)
	return proto.ToHistoryQueryWorkflowResponse(response), proto.ToError(err)
}

func (g grpcClient) ReadDLQMessages(ctx context.Context, request *types.ReadDLQMessagesRequest, opts ...yarpc.CallOption) (*types.ReadDLQMessagesResponse, error) {
	response, err := g.c.ReadDLQMessages(ctx, proto.FromHistoryReadDLQMessagesRequest(request), opts...)
	return proto.ToHistoryReadDLQMessagesResponse(response), proto.ToError(err)
}

func (g grpcClient) ReapplyEvents(ctx context.Context, request *types.HistoryReapplyEventsRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.ReapplyEvents(ctx, proto.FromHistoryReapplyEventsRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RecordActivityTaskHeartbeat(ctx context.Context, request *types.HistoryRecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) (*types.RecordActivityTaskHeartbeatResponse, error) {
	response, err := g.c.RecordActivityTaskHeartbeat(ctx, proto.FromHistoryRecordActivityTaskHeartbeatRequest(request), opts...)
	return proto.ToHistoryRecordActivityTaskHeartbeatResponse(response), proto.ToError(err)
}

func (g grpcClient) RecordActivityTaskStarted(ctx context.Context, request *types.RecordActivityTaskStartedRequest, opts ...yarpc.CallOption) (*types.RecordActivityTaskStartedResponse, error) {
	response, err := g.c.RecordActivityTaskStarted(ctx, proto.FromHistoryRecordActivityTaskStartedRequest(request), opts...)
	return proto.ToHistoryRecordActivityTaskStartedResponse(response), proto.ToError(err)
}

func (g grpcClient) RecordChildExecutionCompleted(ctx context.Context, request *types.RecordChildExecutionCompletedRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.RecordChildExecutionCompleted(ctx, proto.FromHistoryRecordChildExecutionCompletedRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RecordDecisionTaskStarted(ctx context.Context, request *types.RecordDecisionTaskStartedRequest, opts ...yarpc.CallOption) (*types.RecordDecisionTaskStartedResponse, error) {
	response, err := g.c.RecordDecisionTaskStarted(ctx, proto.FromHistoryRecordDecisionTaskStartedRequest(request), opts...)
	return proto.ToHistoryRecordDecisionTaskStartedResponse(response), proto.ToError(err)
}

func (g grpcClient) RefreshWorkflowTasks(ctx context.Context, request *types.HistoryRefreshWorkflowTasksRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.RefreshWorkflowTasks(ctx, proto.FromHistoryRefreshWorkflowTasksRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RemoveSignalMutableState(ctx context.Context, request *types.RemoveSignalMutableStateRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.RemoveSignalMutableState(ctx, proto.FromHistoryRemoveSignalMutableStateRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RemoveTask(ctx context.Context, request *types.RemoveTaskRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.RemoveTask(ctx, proto.FromHistoryRemoveTaskRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) ReplicateEventsV2(ctx context.Context, request *types.ReplicateEventsV2Request, opts ...yarpc.CallOption) error {
	_, err := g.c.ReplicateEventsV2(ctx, proto.FromHistoryReplicateEventsV2Request(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RequestCancelWorkflowExecution(ctx context.Context, request *types.HistoryRequestCancelWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.RequestCancelWorkflowExecution(ctx, proto.FromHistoryRequestCancelWorkflowExecutionRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) ResetQueue(ctx context.Context, request *types.ResetQueueRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.ResetQueue(ctx, proto.FromHistoryResetQueueRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) ResetStickyTaskList(ctx context.Context, request *types.HistoryResetStickyTaskListRequest, opts ...yarpc.CallOption) (*types.HistoryResetStickyTaskListResponse, error) {
	_, err := g.c.ResetStickyTaskList(ctx, proto.FromHistoryResetStickyTaskListRequest(request), opts...)
	return &types.HistoryResetStickyTaskListResponse{}, proto.ToError(err)
}

func (g grpcClient) ResetWorkflowExecution(ctx context.Context, request *types.HistoryResetWorkflowExecutionRequest, opts ...yarpc.CallOption) (*types.ResetWorkflowExecutionResponse, error) {
	response, err := g.c.ResetWorkflowExecution(ctx, proto.FromHistoryResetWorkflowExecutionRequest(request), opts...)
	return proto.ToHistoryResetWorkflowExecutionResponse(response), proto.ToError(err)
}

func (g grpcClient) RespondActivityTaskCanceled(ctx context.Context, request *types.HistoryRespondActivityTaskCanceledRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.RespondActivityTaskCanceled(ctx, proto.FromHistoryRespondActivityTaskCanceledRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RespondActivityTaskCompleted(ctx context.Context, request *types.HistoryRespondActivityTaskCompletedRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.RespondActivityTaskCompleted(ctx, proto.FromHistoryRespondActivityTaskCompletedRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RespondActivityTaskFailed(ctx context.Context, request *types.HistoryRespondActivityTaskFailedRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.RespondActivityTaskFailed(ctx, proto.FromHistoryRespondActivityTaskFailedRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RespondCrossClusterTasksCompleted(ctx context.Context, request *types.RespondCrossClusterTasksCompletedRequest, opts ...yarpc.CallOption) (*types.RespondCrossClusterTasksCompletedResponse, error) {
	response, err := g.c.RespondCrossClusterTasksCompleted(ctx, proto.FromHistoryRespondCrossClusterTasksCompletedRequest(request), opts...)
	return proto.ToHistoryRespondCrossClusterTasksCompletedResponse(response), proto.ToError(err)
}

func (g grpcClient) RespondDecisionTaskCompleted(ctx context.Context, request *types.HistoryRespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption) (*types.HistoryRespondDecisionTaskCompletedResponse, error) {
	response, err := g.c.RespondDecisionTaskCompleted(ctx, proto.FromHistoryRespondDecisionTaskCompletedRequest(request), opts...)
	return proto.ToHistoryRespondDecisionTaskCompletedResponse(response), proto.ToError(err)
}

func (g grpcClient) RespondDecisionTaskFailed(ctx context.Context, request *types.HistoryRespondDecisionTaskFailedRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.RespondDecisionTaskFailed(ctx, proto.FromHistoryRespondDecisionTaskFailedRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) ScheduleDecisionTask(ctx context.Context, request *types.ScheduleDecisionTaskRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.ScheduleDecisionTask(ctx, proto.FromHistoryScheduleDecisionTaskRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) SignalWithStartWorkflowExecution(ctx context.Context, request *types.HistorySignalWithStartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*types.StartWorkflowExecutionResponse, error) {
	response, err := g.c.SignalWithStartWorkflowExecution(ctx, proto.FromHistorySignalWithStartWorkflowExecutionRequest(request), opts...)
	return proto.ToHistorySignalWithStartWorkflowExecutionResponse(response), proto.ToError(err)
}

func (g grpcClient) SignalWorkflowExecution(ctx context.Context, request *types.HistorySignalWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.SignalWorkflowExecution(ctx, proto.FromHistorySignalWorkflowExecutionRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) StartWorkflowExecution(ctx context.Context, request *types.HistoryStartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*types.StartWorkflowExecutionResponse, error) {
	response, err := g.c.StartWorkflowExecution(ctx, proto.FromHistoryStartWorkflowExecutionRequest(request), opts...)
	return proto.ToHistoryStartWorkflowExecutionResponse(response), proto.ToError(err)
}

func (g grpcClient) SyncActivity(ctx context.Context, request *types.SyncActivityRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.SyncActivity(ctx, proto.FromHistorySyncActivityRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) SyncShardStatus(ctx context.Context, request *types.SyncShardStatusRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.SyncShardStatus(ctx, proto.FromHistorySyncShardStatusRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) TerminateWorkflowExecution(ctx context.Context, request *types.HistoryTerminateWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.TerminateWorkflowExecution(ctx, proto.FromHistoryTerminateWorkflowExecutionRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) GetFailoverInfo(ctx context.Context, request *types.GetFailoverInfoRequest, opts ...yarpc.CallOption) (*types.GetFailoverInfoResponse, error) {
	response, err := g.c.GetFailoverInfo(ctx, proto.FromHistoryGetFailoverInfoRequest(request), opts...)
	return proto.ToHistoryGetFailoverInfoResponse(response), proto.ToError(err)
}
