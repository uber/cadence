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

	"github.com/uber/cadence/.gen/go/history"
	"github.com/uber/cadence/.gen/go/history/historyserviceclient"
	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/.gen/go/shared"
)

type thriftClient struct {
	c historyserviceclient.Interface
}

// NewThriftClient creates a new instance of Client with thrift protocol
func NewThriftClient(c historyserviceclient.Interface) Client {
	return thriftClient{c}
}

func (t thriftClient) CloseShard(ctx context.Context, request *shared.CloseShardRequest, opts ...yarpc.CallOption) error {
	return t.c.CloseShard(ctx, request, opts...)
}

func (t thriftClient) DescribeHistoryHost(ctx context.Context, request *shared.DescribeHistoryHostRequest, opts ...yarpc.CallOption) (*shared.DescribeHistoryHostResponse, error) {
	return t.c.DescribeHistoryHost(ctx, request, opts...)
}

func (t thriftClient) DescribeMutableState(ctx context.Context, request *history.DescribeMutableStateRequest, opts ...yarpc.CallOption) (*history.DescribeMutableStateResponse, error) {
	return t.c.DescribeMutableState(ctx, request, opts...)
}

func (t thriftClient) DescribeQueue(ctx context.Context, request *shared.DescribeQueueRequest, opts ...yarpc.CallOption) (*shared.DescribeQueueResponse, error) {
	return t.c.DescribeQueue(ctx, request, opts...)
}

func (t thriftClient) DescribeWorkflowExecution(ctx context.Context, request *history.DescribeWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.DescribeWorkflowExecutionResponse, error) {
	return t.c.DescribeWorkflowExecution(ctx, request, opts...)
}

func (t thriftClient) GetDLQReplicationMessages(ctx context.Context, request *replicator.GetDLQReplicationMessagesRequest, opts ...yarpc.CallOption) (*replicator.GetDLQReplicationMessagesResponse, error) {
	return t.c.GetDLQReplicationMessages(ctx, request, opts...)
}

func (t thriftClient) GetMutableState(ctx context.Context, request *history.GetMutableStateRequest, opts ...yarpc.CallOption) (*history.GetMutableStateResponse, error) {
	return t.c.GetMutableState(ctx, request, opts...)
}

func (t thriftClient) GetReplicationMessages(ctx context.Context, request *replicator.GetReplicationMessagesRequest, opts ...yarpc.CallOption) (*replicator.GetReplicationMessagesResponse, error) {
	return t.c.GetReplicationMessages(ctx, request, opts...)
}

func (t thriftClient) MergeDLQMessages(ctx context.Context, request *replicator.MergeDLQMessagesRequest, opts ...yarpc.CallOption) (*replicator.MergeDLQMessagesResponse, error) {
	return t.c.MergeDLQMessages(ctx, request, opts...)
}

func (t thriftClient) NotifyFailoverMarkers(ctx context.Context, request *history.NotifyFailoverMarkersRequest, opts ...yarpc.CallOption) error {
	return t.c.NotifyFailoverMarkers(ctx, request, opts...)
}

func (t thriftClient) PollMutableState(ctx context.Context, request *history.PollMutableStateRequest, opts ...yarpc.CallOption) (*history.PollMutableStateResponse, error) {
	return t.c.PollMutableState(ctx, request, opts...)
}

func (t thriftClient) PurgeDLQMessages(ctx context.Context, request *replicator.PurgeDLQMessagesRequest, opts ...yarpc.CallOption) error {
	return t.c.PurgeDLQMessages(ctx, request, opts...)
}

func (t thriftClient) QueryWorkflow(ctx context.Context, request *history.QueryWorkflowRequest, opts ...yarpc.CallOption) (*history.QueryWorkflowResponse, error) {
	return t.c.QueryWorkflow(ctx, request, opts...)
}

func (t thriftClient) ReadDLQMessages(ctx context.Context, request *replicator.ReadDLQMessagesRequest, opts ...yarpc.CallOption) (*replicator.ReadDLQMessagesResponse, error) {
	return t.c.ReadDLQMessages(ctx, request, opts...)
}

func (t thriftClient) ReapplyEvents(ctx context.Context, request *history.ReapplyEventsRequest, opts ...yarpc.CallOption) error {
	return t.c.ReapplyEvents(ctx, request, opts...)
}

func (t thriftClient) RecordActivityTaskHeartbeat(ctx context.Context, request *history.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	return t.c.RecordActivityTaskHeartbeat(ctx, request, opts...)
}

func (t thriftClient) RecordActivityTaskStarted(ctx context.Context, request *history.RecordActivityTaskStartedRequest, opts ...yarpc.CallOption) (*history.RecordActivityTaskStartedResponse, error) {
	return t.c.RecordActivityTaskStarted(ctx, request, opts...)
}

func (t thriftClient) RecordChildExecutionCompleted(ctx context.Context, request *history.RecordChildExecutionCompletedRequest, opts ...yarpc.CallOption) error {
	return t.c.RecordChildExecutionCompleted(ctx, request, opts...)
}

func (t thriftClient) RecordDecisionTaskStarted(ctx context.Context, request *history.RecordDecisionTaskStartedRequest, opts ...yarpc.CallOption) (*history.RecordDecisionTaskStartedResponse, error) {
	return t.c.RecordDecisionTaskStarted(ctx, request, opts...)
}

func (t thriftClient) RefreshWorkflowTasks(ctx context.Context, request *history.RefreshWorkflowTasksRequest, opts ...yarpc.CallOption) error {
	return t.c.RefreshWorkflowTasks(ctx, request, opts...)
}

func (t thriftClient) RemoveSignalMutableState(ctx context.Context, request *history.RemoveSignalMutableStateRequest, opts ...yarpc.CallOption) error {
	return t.c.RemoveSignalMutableState(ctx, request, opts...)
}

func (t thriftClient) RemoveTask(ctx context.Context, request *shared.RemoveTaskRequest, opts ...yarpc.CallOption) error {
	return t.c.RemoveTask(ctx, request, opts...)
}

func (t thriftClient) ReplicateEventsV2(ctx context.Context, request *history.ReplicateEventsV2Request, opts ...yarpc.CallOption) error {
	return t.c.ReplicateEventsV2(ctx, request, opts...)
}

func (t thriftClient) RequestCancelWorkflowExecution(ctx context.Context, request *history.RequestCancelWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	return t.c.RequestCancelWorkflowExecution(ctx, request, opts...)
}

func (t thriftClient) ResetQueue(ctx context.Context, request *shared.ResetQueueRequest, opts ...yarpc.CallOption) error {
	return t.c.ResetQueue(ctx, request, opts...)
}

func (t thriftClient) ResetStickyTaskList(ctx context.Context, request *history.ResetStickyTaskListRequest, opts ...yarpc.CallOption) (*history.ResetStickyTaskListResponse, error) {
	return t.c.ResetStickyTaskList(ctx, request, opts...)
}

func (t thriftClient) ResetWorkflowExecution(ctx context.Context, request *history.ResetWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.ResetWorkflowExecutionResponse, error) {
	return t.c.ResetWorkflowExecution(ctx, request, opts...)
}

func (t thriftClient) RespondActivityTaskCanceled(ctx context.Context, request *history.RespondActivityTaskCanceledRequest, opts ...yarpc.CallOption) error {
	return t.c.RespondActivityTaskCanceled(ctx, request, opts...)
}

func (t thriftClient) RespondActivityTaskCompleted(ctx context.Context, request *history.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption) error {
	return t.c.RespondActivityTaskCompleted(ctx, request, opts...)
}

func (t thriftClient) RespondActivityTaskFailed(ctx context.Context, request *history.RespondActivityTaskFailedRequest, opts ...yarpc.CallOption) error {
	return t.c.RespondActivityTaskFailed(ctx, request, opts...)
}

func (t thriftClient) RespondDecisionTaskCompleted(ctx context.Context, request *history.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption) (*history.RespondDecisionTaskCompletedResponse, error) {
	return t.c.RespondDecisionTaskCompleted(ctx, request, opts...)
}

func (t thriftClient) RespondDecisionTaskFailed(ctx context.Context, request *history.RespondDecisionTaskFailedRequest, opts ...yarpc.CallOption) error {
	return t.c.RespondDecisionTaskFailed(ctx, request, opts...)
}

func (t thriftClient) ScheduleDecisionTask(ctx context.Context, request *history.ScheduleDecisionTaskRequest, opts ...yarpc.CallOption) error {
	return t.c.ScheduleDecisionTask(ctx, request, opts...)
}

func (t thriftClient) SignalWithStartWorkflowExecution(ctx context.Context, request *history.SignalWithStartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
	return t.c.SignalWithStartWorkflowExecution(ctx, request, opts...)
}

func (t thriftClient) SignalWorkflowExecution(ctx context.Context, request *history.SignalWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	return t.c.SignalWorkflowExecution(ctx, request, opts...)
}

func (t thriftClient) StartWorkflowExecution(ctx context.Context, request *history.StartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
	return t.c.StartWorkflowExecution(ctx, request, opts...)
}

func (t thriftClient) SyncActivity(ctx context.Context, request *history.SyncActivityRequest, opts ...yarpc.CallOption) error {
	return t.c.SyncActivity(ctx, request, opts...)
}

func (t thriftClient) SyncShardStatus(ctx context.Context, request *history.SyncShardStatusRequest, opts ...yarpc.CallOption) error {
	return t.c.SyncShardStatus(ctx, request, opts...)
}

func (t thriftClient) TerminateWorkflowExecution(ctx context.Context, request *history.TerminateWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	return t.c.TerminateWorkflowExecution(ctx, request, opts...)
}
