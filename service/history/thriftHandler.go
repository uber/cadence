// Copyright (c) 2020 Uber Technologies Inc.
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

	"github.com/uber/cadence/.gen/go/health"
	"github.com/uber/cadence/.gen/go/health/metaserver"
	h "github.com/uber/cadence/.gen/go/history"
	"github.com/uber/cadence/.gen/go/history/historyserviceserver"
	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/.gen/go/shared"
)

// ThriftHandler wrap underlying handler and handles Thrift related type conversions
type ThriftHandler struct {
	h Handler
}

// NewThriftHandler creates Thrift handler on top of underlying handler
func NewThriftHandler(h Handler) ThriftHandler {
	return ThriftHandler{h}
}

func (t ThriftHandler) register(dispatcher *yarpc.Dispatcher) {
	dispatcher.Register(historyserviceserver.New(&t))
	dispatcher.Register(metaserver.New(&t))
}

// Health forwards request to the underlying handler
func (t ThriftHandler) Health(ctx context.Context) (response *health.HealthStatus, err error) {
	return t.h.Health(ctx)
}

// CloseShard forwards request to the underlying handler
func (t ThriftHandler) CloseShard(ctx context.Context, request *shared.CloseShardRequest) (err error) {
	return t.h.CloseShard(ctx, request)
}

// DescribeHistoryHost forwards request to the underlying handler
func (t ThriftHandler) DescribeHistoryHost(ctx context.Context, request *shared.DescribeHistoryHostRequest) (response *shared.DescribeHistoryHostResponse, err error) {
	return t.h.DescribeHistoryHost(ctx, request)
}

// DescribeMutableState forwards request to the underlying handler
func (t ThriftHandler) DescribeMutableState(ctx context.Context, request *h.DescribeMutableStateRequest) (response *h.DescribeMutableStateResponse, err error) {
	return t.h.DescribeMutableState(ctx, request)
}

// DescribeQueue forwards request to the underlying handler
func (t ThriftHandler) DescribeQueue(ctx context.Context, request *shared.DescribeQueueRequest) (response *shared.DescribeQueueResponse, err error) {
	return t.h.DescribeQueue(ctx, request)
}

// DescribeWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) DescribeWorkflowExecution(ctx context.Context, request *h.DescribeWorkflowExecutionRequest) (response *shared.DescribeWorkflowExecutionResponse, err error) {
	return t.h.DescribeWorkflowExecution(ctx, request)
}

// GetDLQReplicationMessages forwards request to the underlying handler
func (t ThriftHandler) GetDLQReplicationMessages(ctx context.Context, request *replicator.GetDLQReplicationMessagesRequest) (response *replicator.GetDLQReplicationMessagesResponse, err error) {
	return t.h.GetDLQReplicationMessages(ctx, request)
}

// GetMutableState forwards request to the underlying handler
func (t ThriftHandler) GetMutableState(ctx context.Context, request *h.GetMutableStateRequest) (response *h.GetMutableStateResponse, err error) {
	return t.h.GetMutableState(ctx, request)
}

// GetReplicationMessages forwards request to the underlying handler
func (t ThriftHandler) GetReplicationMessages(ctx context.Context, request *replicator.GetReplicationMessagesRequest) (response *replicator.GetReplicationMessagesResponse, err error) {
	return t.h.GetReplicationMessages(ctx, request)
}

// MergeDLQMessages forwards request to the underlying handler
func (t ThriftHandler) MergeDLQMessages(ctx context.Context, request *replicator.MergeDLQMessagesRequest) (response *replicator.MergeDLQMessagesResponse, err error) {
	return t.h.MergeDLQMessages(ctx, request)
}

// NotifyFailoverMarkers forwards request to the underlying handler
func (t ThriftHandler) NotifyFailoverMarkers(ctx context.Context, request *h.NotifyFailoverMarkersRequest) (err error) {
	return t.h.NotifyFailoverMarkers(ctx, request)
}

// PollMutableState forwards request to the underlying handler
func (t ThriftHandler) PollMutableState(ctx context.Context, request *h.PollMutableStateRequest) (response *h.PollMutableStateResponse, err error) {
	return t.h.PollMutableState(ctx, request)
}

// PurgeDLQMessages forwards request to the underlying handler
func (t ThriftHandler) PurgeDLQMessages(ctx context.Context, request *replicator.PurgeDLQMessagesRequest) (err error) {
	return t.h.PurgeDLQMessages(ctx, request)
}

// QueryWorkflow forwards request to the underlying handler
func (t ThriftHandler) QueryWorkflow(ctx context.Context, request *h.QueryWorkflowRequest) (response *h.QueryWorkflowResponse, err error) {
	return t.h.QueryWorkflow(ctx, request)
}

// ReadDLQMessages forwards request to the underlying handler
func (t ThriftHandler) ReadDLQMessages(ctx context.Context, request *replicator.ReadDLQMessagesRequest) (response *replicator.ReadDLQMessagesResponse, err error) {
	return t.h.ReadDLQMessages(ctx, request)
}

// ReapplyEvents forwards request to the underlying handler
func (t ThriftHandler) ReapplyEvents(ctx context.Context, request *h.ReapplyEventsRequest) (err error) {
	return t.h.ReapplyEvents(ctx, request)
}

// RecordActivityTaskHeartbeat forwards request to the underlying handler
func (t ThriftHandler) RecordActivityTaskHeartbeat(ctx context.Context, request *h.RecordActivityTaskHeartbeatRequest) (response *shared.RecordActivityTaskHeartbeatResponse, err error) {
	return t.h.RecordActivityTaskHeartbeat(ctx, request)
}

// RecordActivityTaskStarted forwards request to the underlying handler
func (t ThriftHandler) RecordActivityTaskStarted(ctx context.Context, request *h.RecordActivityTaskStartedRequest) (response *h.RecordActivityTaskStartedResponse, err error) {
	return t.h.RecordActivityTaskStarted(ctx, request)
}

// RecordChildExecutionCompleted forwards request to the underlying handler
func (t ThriftHandler) RecordChildExecutionCompleted(ctx context.Context, request *h.RecordChildExecutionCompletedRequest) (err error) {
	return t.h.RecordChildExecutionCompleted(ctx, request)
}

// RecordDecisionTaskStarted forwards request to the underlying handler
func (t ThriftHandler) RecordDecisionTaskStarted(ctx context.Context, request *h.RecordDecisionTaskStartedRequest) (response *h.RecordDecisionTaskStartedResponse, err error) {
	return t.h.RecordDecisionTaskStarted(ctx, request)
}

// RefreshWorkflowTasks forwards request to the underlying handler
func (t ThriftHandler) RefreshWorkflowTasks(ctx context.Context, request *h.RefreshWorkflowTasksRequest) (err error) {
	return t.h.RefreshWorkflowTasks(ctx, request)
}

// RemoveSignalMutableState forwards request to the underlying handler
func (t ThriftHandler) RemoveSignalMutableState(ctx context.Context, request *h.RemoveSignalMutableStateRequest) (err error) {
	return t.h.RemoveSignalMutableState(ctx, request)
}

// RemoveTask forwards request to the underlying handler
func (t ThriftHandler) RemoveTask(ctx context.Context, request *shared.RemoveTaskRequest) (err error) {
	return t.h.RemoveTask(ctx, request)
}

// ReplicateEventsV2 forwards request to the underlying handler
func (t ThriftHandler) ReplicateEventsV2(ctx context.Context, request *h.ReplicateEventsV2Request) (err error) {
	return t.h.ReplicateEventsV2(ctx, request)
}

// RequestCancelWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) RequestCancelWorkflowExecution(ctx context.Context, request *h.RequestCancelWorkflowExecutionRequest) (err error) {
	return t.h.RequestCancelWorkflowExecution(ctx, request)
}

// ResetQueue forwards request to the underlying handler
func (t ThriftHandler) ResetQueue(ctx context.Context, request *shared.ResetQueueRequest) (err error) {
	return t.h.ResetQueue(ctx, request)
}

// ResetStickyTaskList forwards request to the underlying handler
func (t ThriftHandler) ResetStickyTaskList(ctx context.Context, request *h.ResetStickyTaskListRequest) (response *h.ResetStickyTaskListResponse, err error) {
	return t.h.ResetStickyTaskList(ctx, request)
}

// ResetWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) ResetWorkflowExecution(ctx context.Context, request *h.ResetWorkflowExecutionRequest) (response *shared.ResetWorkflowExecutionResponse, err error) {
	return t.h.ResetWorkflowExecution(ctx, request)
}

// RespondActivityTaskCanceled forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCanceled(ctx context.Context, request *h.RespondActivityTaskCanceledRequest) (err error) {
	return t.h.RespondActivityTaskCanceled(ctx, request)
}

// RespondActivityTaskCompleted forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCompleted(ctx context.Context, request *h.RespondActivityTaskCompletedRequest) (err error) {
	return t.h.RespondActivityTaskCompleted(ctx, request)
}

// RespondActivityTaskFailed forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskFailed(ctx context.Context, request *h.RespondActivityTaskFailedRequest) (err error) {
	return t.h.RespondActivityTaskFailed(ctx, request)
}

// RespondDecisionTaskCompleted forwards request to the underlying handler
func (t ThriftHandler) RespondDecisionTaskCompleted(ctx context.Context, request *h.RespondDecisionTaskCompletedRequest) (response *h.RespondDecisionTaskCompletedResponse, err error) {
	return t.h.RespondDecisionTaskCompleted(ctx, request)
}

// RespondDecisionTaskFailed forwards request to the underlying handler
func (t ThriftHandler) RespondDecisionTaskFailed(ctx context.Context, request *h.RespondDecisionTaskFailedRequest) (err error) {
	return t.h.RespondDecisionTaskFailed(ctx, request)
}

// ScheduleDecisionTask forwards request to the underlying handler
func (t ThriftHandler) ScheduleDecisionTask(ctx context.Context, request *h.ScheduleDecisionTaskRequest) (err error) {
	return t.h.ScheduleDecisionTask(ctx, request)
}

// SignalWithStartWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) SignalWithStartWorkflowExecution(ctx context.Context, request *h.SignalWithStartWorkflowExecutionRequest) (response *shared.StartWorkflowExecutionResponse, err error) {
	return t.h.SignalWithStartWorkflowExecution(ctx, request)
}

// SignalWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) SignalWorkflowExecution(ctx context.Context, request *h.SignalWorkflowExecutionRequest) (err error) {
	return t.h.SignalWorkflowExecution(ctx, request)
}

// StartWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) StartWorkflowExecution(ctx context.Context, request *h.StartWorkflowExecutionRequest) (response *shared.StartWorkflowExecutionResponse, err error) {
	return t.h.StartWorkflowExecution(ctx, request)
}

// SyncActivity forwards request to the underlying handler
func (t ThriftHandler) SyncActivity(ctx context.Context, request *h.SyncActivityRequest) (err error) {
	return t.h.SyncActivity(ctx, request)
}

// SyncShardStatus forwards request to the underlying handler
func (t ThriftHandler) SyncShardStatus(ctx context.Context, request *h.SyncShardStatusRequest) (err error) {
	return t.h.SyncShardStatus(ctx, request)
}

// TerminateWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) TerminateWorkflowExecution(ctx context.Context, request *h.TerminateWorkflowExecutionRequest) (err error) {
	return t.h.TerminateWorkflowExecution(ctx, request)
}
