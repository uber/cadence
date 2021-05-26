// Copyright (c) 2020 Uber Technologies, Inc.
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

	"github.com/uber/cadence/.gen/go/history/historyserviceclient"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"
)

type thriftClient struct {
	c historyserviceclient.Interface
}

// NewThriftClient creates a new instance of Client with thrift protocol
func NewThriftClient(c historyserviceclient.Interface) Client {
	return thriftClient{c}
}

func (t thriftClient) CloseShard(ctx context.Context, request *types.CloseShardRequest, opts ...yarpc.CallOption) error {
	err := t.c.CloseShard(ctx, thrift.FromCloseShardRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) DescribeHistoryHost(ctx context.Context, request *types.DescribeHistoryHostRequest, opts ...yarpc.CallOption) (*types.DescribeHistoryHostResponse, error) {
	response, err := t.c.DescribeHistoryHost(ctx, thrift.FromDescribeHistoryHostRequest(request), opts...)
	return thrift.ToDescribeHistoryHostResponse(response), thrift.ToError(err)
}

func (t thriftClient) DescribeMutableState(ctx context.Context, request *types.DescribeMutableStateRequest, opts ...yarpc.CallOption) (*types.DescribeMutableStateResponse, error) {
	response, err := t.c.DescribeMutableState(ctx, thrift.FromDescribeMutableStateRequest(request), opts...)
	return thrift.ToDescribeMutableStateResponse(response), thrift.ToError(err)
}

func (t thriftClient) DescribeQueue(ctx context.Context, request *types.DescribeQueueRequest, opts ...yarpc.CallOption) (*types.DescribeQueueResponse, error) {
	response, err := t.c.DescribeQueue(ctx, thrift.FromDescribeQueueRequest(request), opts...)
	return thrift.ToDescribeQueueResponse(response), thrift.ToError(err)
}

func (t thriftClient) DescribeWorkflowExecution(ctx context.Context, request *types.HistoryDescribeWorkflowExecutionRequest, opts ...yarpc.CallOption) (*types.DescribeWorkflowExecutionResponse, error) {
	response, err := t.c.DescribeWorkflowExecution(ctx, thrift.FromHistoryDescribeWorkflowExecutionRequest(request), opts...)
	return thrift.ToDescribeWorkflowExecutionResponse(response), thrift.ToError(err)
}

func (t thriftClient) GetDLQReplicationMessages(ctx context.Context, request *types.GetDLQReplicationMessagesRequest, opts ...yarpc.CallOption) (*types.GetDLQReplicationMessagesResponse, error) {
	response, err := t.c.GetDLQReplicationMessages(ctx, thrift.FromGetDLQReplicationMessagesRequest(request), opts...)
	return thrift.ToGetDLQReplicationMessagesResponse(response), thrift.ToError(err)
}

func (t thriftClient) GetMutableState(ctx context.Context, request *types.GetMutableStateRequest, opts ...yarpc.CallOption) (*types.GetMutableStateResponse, error) {
	response, err := t.c.GetMutableState(ctx, thrift.FromGetMutableStateRequest(request), opts...)
	return thrift.ToGetMutableStateResponse(response), thrift.ToError(err)
}

func (t thriftClient) GetReplicationMessages(ctx context.Context, request *types.GetReplicationMessagesRequest, opts ...yarpc.CallOption) (*types.GetReplicationMessagesResponse, error) {
	response, err := t.c.GetReplicationMessages(ctx, thrift.FromGetReplicationMessagesRequest(request), opts...)
	return thrift.ToGetReplicationMessagesResponse(response), thrift.ToError(err)
}

func (t thriftClient) MergeDLQMessages(ctx context.Context, request *types.MergeDLQMessagesRequest, opts ...yarpc.CallOption) (*types.MergeDLQMessagesResponse, error) {
	response, err := t.c.MergeDLQMessages(ctx, thrift.FromMergeDLQMessagesRequest(request), opts...)
	return thrift.ToMergeDLQMessagesResponse(response), thrift.ToError(err)
}

func (t thriftClient) NotifyFailoverMarkers(ctx context.Context, request *types.NotifyFailoverMarkersRequest, opts ...yarpc.CallOption) error {
	err := t.c.NotifyFailoverMarkers(ctx, thrift.FromNotifyFailoverMarkersRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) PollMutableState(ctx context.Context, request *types.PollMutableStateRequest, opts ...yarpc.CallOption) (*types.PollMutableStateResponse, error) {
	response, err := t.c.PollMutableState(ctx, thrift.FromPollMutableStateRequest(request), opts...)
	return thrift.ToPollMutableStateResponse(response), thrift.ToError(err)
}

func (t thriftClient) PurgeDLQMessages(ctx context.Context, request *types.PurgeDLQMessagesRequest, opts ...yarpc.CallOption) error {
	err := t.c.PurgeDLQMessages(ctx, thrift.FromPurgeDLQMessagesRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) QueryWorkflow(ctx context.Context, request *types.HistoryQueryWorkflowRequest, opts ...yarpc.CallOption) (*types.HistoryQueryWorkflowResponse, error) {
	response, err := t.c.QueryWorkflow(ctx, thrift.FromHistoryQueryWorkflowRequest(request), opts...)
	return thrift.ToHistoryQueryWorkflowResponse(response), thrift.ToError(err)
}

func (t thriftClient) ReadDLQMessages(ctx context.Context, request *types.ReadDLQMessagesRequest, opts ...yarpc.CallOption) (*types.ReadDLQMessagesResponse, error) {
	response, err := t.c.ReadDLQMessages(ctx, thrift.FromReadDLQMessagesRequest(request), opts...)
	return thrift.ToReadDLQMessagesResponse(response), thrift.ToError(err)
}

func (t thriftClient) ReapplyEvents(ctx context.Context, request *types.HistoryReapplyEventsRequest, opts ...yarpc.CallOption) error {
	err := t.c.ReapplyEvents(ctx, thrift.FromHistoryReapplyEventsRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) RecordActivityTaskHeartbeat(ctx context.Context, request *types.HistoryRecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) (*types.RecordActivityTaskHeartbeatResponse, error) {
	response, err := t.c.RecordActivityTaskHeartbeat(ctx, thrift.FromHistoryRecordActivityTaskHeartbeatRequest(request), opts...)
	return thrift.ToRecordActivityTaskHeartbeatResponse(response), thrift.ToError(err)
}

func (t thriftClient) RecordActivityTaskStarted(ctx context.Context, request *types.RecordActivityTaskStartedRequest, opts ...yarpc.CallOption) (*types.RecordActivityTaskStartedResponse, error) {
	response, err := t.c.RecordActivityTaskStarted(ctx, thrift.FromRecordActivityTaskStartedRequest(request), opts...)
	return thrift.ToRecordActivityTaskStartedResponse(response), thrift.ToError(err)
}

func (t thriftClient) RecordChildExecutionCompleted(ctx context.Context, request *types.RecordChildExecutionCompletedRequest, opts ...yarpc.CallOption) error {
	err := t.c.RecordChildExecutionCompleted(ctx, thrift.FromRecordChildExecutionCompletedRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) RecordDecisionTaskStarted(ctx context.Context, request *types.RecordDecisionTaskStartedRequest, opts ...yarpc.CallOption) (*types.RecordDecisionTaskStartedResponse, error) {
	response, err := t.c.RecordDecisionTaskStarted(ctx, thrift.FromRecordDecisionTaskStartedRequest(request), opts...)
	return thrift.ToRecordDecisionTaskStartedResponse(response), thrift.ToError(err)
}

func (t thriftClient) RefreshWorkflowTasks(ctx context.Context, request *types.HistoryRefreshWorkflowTasksRequest, opts ...yarpc.CallOption) error {
	err := t.c.RefreshWorkflowTasks(ctx, thrift.FromHistoryRefreshWorkflowTasksRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) RemoveSignalMutableState(ctx context.Context, request *types.RemoveSignalMutableStateRequest, opts ...yarpc.CallOption) error {
	err := t.c.RemoveSignalMutableState(ctx, thrift.FromRemoveSignalMutableStateRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) RemoveTask(ctx context.Context, request *types.RemoveTaskRequest, opts ...yarpc.CallOption) error {
	err := t.c.RemoveTask(ctx, thrift.FromRemoveTaskRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) ReplicateEventsV2(ctx context.Context, request *types.ReplicateEventsV2Request, opts ...yarpc.CallOption) error {
	err := t.c.ReplicateEventsV2(ctx, thrift.FromReplicateEventsV2Request(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) RequestCancelWorkflowExecution(ctx context.Context, request *types.HistoryRequestCancelWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	err := t.c.RequestCancelWorkflowExecution(ctx, thrift.FromHistoryRequestCancelWorkflowExecutionRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) ResetQueue(ctx context.Context, request *types.ResetQueueRequest, opts ...yarpc.CallOption) error {
	err := t.c.ResetQueue(ctx, thrift.FromResetQueueRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) ResetStickyTaskList(ctx context.Context, request *types.HistoryResetStickyTaskListRequest, opts ...yarpc.CallOption) (*types.HistoryResetStickyTaskListResponse, error) {
	response, err := t.c.ResetStickyTaskList(ctx, thrift.FromHistoryResetStickyTaskListRequest(request), opts...)
	return thrift.ToHistoryResetStickyTaskListResponse(response), thrift.ToError(err)
}

func (t thriftClient) ResetWorkflowExecution(ctx context.Context, request *types.HistoryResetWorkflowExecutionRequest, opts ...yarpc.CallOption) (*types.ResetWorkflowExecutionResponse, error) {
	response, err := t.c.ResetWorkflowExecution(ctx, thrift.FromHistoryResetWorkflowExecutionRequest(request), opts...)
	return thrift.ToResetWorkflowExecutionResponse(response), thrift.ToError(err)
}

func (t thriftClient) RespondActivityTaskCanceled(ctx context.Context, request *types.HistoryRespondActivityTaskCanceledRequest, opts ...yarpc.CallOption) error {
	err := t.c.RespondActivityTaskCanceled(ctx, thrift.FromHistoryRespondActivityTaskCanceledRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) RespondActivityTaskCompleted(ctx context.Context, request *types.HistoryRespondActivityTaskCompletedRequest, opts ...yarpc.CallOption) error {
	err := t.c.RespondActivityTaskCompleted(ctx, thrift.FromHistoryRespondActivityTaskCompletedRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) RespondActivityTaskFailed(ctx context.Context, request *types.HistoryRespondActivityTaskFailedRequest, opts ...yarpc.CallOption) error {
	err := t.c.RespondActivityTaskFailed(ctx, thrift.FromHistoryRespondActivityTaskFailedRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) RespondDecisionTaskCompleted(ctx context.Context, request *types.HistoryRespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption) (*types.HistoryRespondDecisionTaskCompletedResponse, error) {
	response, err := t.c.RespondDecisionTaskCompleted(ctx, thrift.FromHistoryRespondDecisionTaskCompletedRequest(request), opts...)
	return thrift.ToHistoryRespondDecisionTaskCompletedResponse(response), thrift.ToError(err)
}

func (t thriftClient) RespondDecisionTaskFailed(ctx context.Context, request *types.HistoryRespondDecisionTaskFailedRequest, opts ...yarpc.CallOption) error {
	err := t.c.RespondDecisionTaskFailed(ctx, thrift.FromHistoryRespondDecisionTaskFailedRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) ScheduleDecisionTask(ctx context.Context, request *types.ScheduleDecisionTaskRequest, opts ...yarpc.CallOption) error {
	err := t.c.ScheduleDecisionTask(ctx, thrift.FromScheduleDecisionTaskRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) SignalWithStartWorkflowExecution(ctx context.Context, request *types.HistorySignalWithStartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*types.StartWorkflowExecutionResponse, error) {
	response, err := t.c.SignalWithStartWorkflowExecution(ctx, thrift.FromHistorySignalWithStartWorkflowExecutionRequest(request), opts...)
	return thrift.ToStartWorkflowExecutionResponse(response), thrift.ToError(err)
}

func (t thriftClient) SignalWorkflowExecution(ctx context.Context, request *types.HistorySignalWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	err := t.c.SignalWorkflowExecution(ctx, thrift.FromHistorySignalWorkflowExecutionRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) StartWorkflowExecution(ctx context.Context, request *types.HistoryStartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*types.StartWorkflowExecutionResponse, error) {
	response, err := t.c.StartWorkflowExecution(ctx, thrift.FromHistoryStartWorkflowExecutionRequest(request), opts...)
	return thrift.ToStartWorkflowExecutionResponse(response), thrift.ToError(err)
}

func (t thriftClient) SyncActivity(ctx context.Context, request *types.SyncActivityRequest, opts ...yarpc.CallOption) error {
	err := t.c.SyncActivity(ctx, thrift.FromSyncActivityRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) SyncShardStatus(ctx context.Context, request *types.SyncShardStatusRequest, opts ...yarpc.CallOption) error {
	err := t.c.SyncShardStatus(ctx, thrift.FromSyncShardStatusRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) TerminateWorkflowExecution(ctx context.Context, request *types.HistoryTerminateWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	err := t.c.TerminateWorkflowExecution(ctx, thrift.FromHistoryTerminateWorkflowExecutionRequest(request), opts...)
	return thrift.ToError(err)
}
