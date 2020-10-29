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
	"github.com/uber/cadence/common/types/mapper/thrift"
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
	response, err = t.h.Health(ctx)
	return response, thrift.FromError(err)
}

// CloseShard forwards request to the underlying handler
func (t ThriftHandler) CloseShard(ctx context.Context, request *shared.CloseShardRequest) (err error) {
	err = t.h.CloseShard(ctx, request)
	return thrift.FromError(err)
}

// DescribeHistoryHost forwards request to the underlying handler
func (t ThriftHandler) DescribeHistoryHost(ctx context.Context, request *shared.DescribeHistoryHostRequest) (response *shared.DescribeHistoryHostResponse, err error) {
	response, err = t.h.DescribeHistoryHost(ctx, request)
	return response, thrift.FromError(err)
}

// DescribeMutableState forwards request to the underlying handler
func (t ThriftHandler) DescribeMutableState(ctx context.Context, request *h.DescribeMutableStateRequest) (response *h.DescribeMutableStateResponse, err error) {
	response, err = t.h.DescribeMutableState(ctx, request)
	return response, thrift.FromError(err)
}

// DescribeQueue forwards request to the underlying handler
func (t ThriftHandler) DescribeQueue(ctx context.Context, request *shared.DescribeQueueRequest) (response *shared.DescribeQueueResponse, err error) {
	response, err = t.h.DescribeQueue(ctx, request)
	return response, thrift.FromError(err)
}

// DescribeWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) DescribeWorkflowExecution(ctx context.Context, request *h.DescribeWorkflowExecutionRequest) (response *shared.DescribeWorkflowExecutionResponse, err error) {
	response, err = t.h.DescribeWorkflowExecution(ctx, request)
	return response, thrift.FromError(err)
}

// GetDLQReplicationMessages forwards request to the underlying handler
func (t ThriftHandler) GetDLQReplicationMessages(ctx context.Context, request *replicator.GetDLQReplicationMessagesRequest) (response *replicator.GetDLQReplicationMessagesResponse, err error) {
	response, err = t.h.GetDLQReplicationMessages(ctx, request)
	return response, thrift.FromError(err)
}

// GetMutableState forwards request to the underlying handler
func (t ThriftHandler) GetMutableState(ctx context.Context, request *h.GetMutableStateRequest) (response *h.GetMutableStateResponse, err error) {
	response, err = t.h.GetMutableState(ctx, request)
	return response, thrift.FromError(err)
}

// GetReplicationMessages forwards request to the underlying handler
func (t ThriftHandler) GetReplicationMessages(ctx context.Context, request *replicator.GetReplicationMessagesRequest) (response *replicator.GetReplicationMessagesResponse, err error) {
	response, err = t.h.GetReplicationMessages(ctx, request)
	return response, thrift.FromError(err)
}

// MergeDLQMessages forwards request to the underlying handler
func (t ThriftHandler) MergeDLQMessages(ctx context.Context, request *replicator.MergeDLQMessagesRequest) (response *replicator.MergeDLQMessagesResponse, err error) {
	response, err = t.h.MergeDLQMessages(ctx, request)
	return response, thrift.FromError(err)
}

// NotifyFailoverMarkers forwards request to the underlying handler
func (t ThriftHandler) NotifyFailoverMarkers(ctx context.Context, request *h.NotifyFailoverMarkersRequest) (err error) {
	err = t.h.NotifyFailoverMarkers(ctx, request)
	return thrift.FromError(err)
}

// PollMutableState forwards request to the underlying handler
func (t ThriftHandler) PollMutableState(ctx context.Context, request *h.PollMutableStateRequest) (response *h.PollMutableStateResponse, err error) {
	response, err = t.h.PollMutableState(ctx, request)
	return response, thrift.FromError(err)
}

// PurgeDLQMessages forwards request to the underlying handler
func (t ThriftHandler) PurgeDLQMessages(ctx context.Context, request *replicator.PurgeDLQMessagesRequest) (err error) {
	err = t.h.PurgeDLQMessages(ctx, request)
	return thrift.FromError(err)
}

// QueryWorkflow forwards request to the underlying handler
func (t ThriftHandler) QueryWorkflow(ctx context.Context, request *h.QueryWorkflowRequest) (response *h.QueryWorkflowResponse, err error) {
	response, err = t.h.QueryWorkflow(ctx, request)
	return response, thrift.FromError(err)
}

// ReadDLQMessages forwards request to the underlying handler
func (t ThriftHandler) ReadDLQMessages(ctx context.Context, request *replicator.ReadDLQMessagesRequest) (response *replicator.ReadDLQMessagesResponse, err error) {
	response, err = t.h.ReadDLQMessages(ctx, request)
	return response, thrift.FromError(err)
}

// ReapplyEvents forwards request to the underlying handler
func (t ThriftHandler) ReapplyEvents(ctx context.Context, request *h.ReapplyEventsRequest) (err error) {
	err = t.h.ReapplyEvents(ctx, request)
	return thrift.FromError(err)
}

// RecordActivityTaskHeartbeat forwards request to the underlying handler
func (t ThriftHandler) RecordActivityTaskHeartbeat(ctx context.Context, request *h.RecordActivityTaskHeartbeatRequest) (response *shared.RecordActivityTaskHeartbeatResponse, err error) {
	response, err = t.h.RecordActivityTaskHeartbeat(ctx, request)
	return response, thrift.FromError(err)
}

// RecordActivityTaskStarted forwards request to the underlying handler
func (t ThriftHandler) RecordActivityTaskStarted(ctx context.Context, request *h.RecordActivityTaskStartedRequest) (response *h.RecordActivityTaskStartedResponse, err error) {
	response, err = t.h.RecordActivityTaskStarted(ctx, request)
	return response, thrift.FromError(err)
}

// RecordChildExecutionCompleted forwards request to the underlying handler
func (t ThriftHandler) RecordChildExecutionCompleted(ctx context.Context, request *h.RecordChildExecutionCompletedRequest) (err error) {
	err = t.h.RecordChildExecutionCompleted(ctx, request)
	return thrift.FromError(err)
}

// RecordDecisionTaskStarted forwards request to the underlying handler
func (t ThriftHandler) RecordDecisionTaskStarted(ctx context.Context, request *h.RecordDecisionTaskStartedRequest) (response *h.RecordDecisionTaskStartedResponse, err error) {
	response, err = t.h.RecordDecisionTaskStarted(ctx, request)
	return response, thrift.FromError(err)
}

// RefreshWorkflowTasks forwards request to the underlying handler
func (t ThriftHandler) RefreshWorkflowTasks(ctx context.Context, request *h.RefreshWorkflowTasksRequest) (err error) {
	err = t.h.RefreshWorkflowTasks(ctx, request)
	return thrift.FromError(err)
}

// RemoveSignalMutableState forwards request to the underlying handler
func (t ThriftHandler) RemoveSignalMutableState(ctx context.Context, request *h.RemoveSignalMutableStateRequest) (err error) {
	err = t.h.RemoveSignalMutableState(ctx, request)
	return thrift.FromError(err)
}

// RemoveTask forwards request to the underlying handler
func (t ThriftHandler) RemoveTask(ctx context.Context, request *shared.RemoveTaskRequest) (err error) {
	err = t.h.RemoveTask(ctx, request)
	return thrift.FromError(err)
}

// ReplicateEventsV2 forwards request to the underlying handler
func (t ThriftHandler) ReplicateEventsV2(ctx context.Context, request *h.ReplicateEventsV2Request) (err error) {
	err = t.h.ReplicateEventsV2(ctx, request)
	return thrift.FromError(err)
}

// RequestCancelWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) RequestCancelWorkflowExecution(ctx context.Context, request *h.RequestCancelWorkflowExecutionRequest) (err error) {
	err = t.h.RequestCancelWorkflowExecution(ctx, request)
	return thrift.FromError(err)
}

// ResetQueue forwards request to the underlying handler
func (t ThriftHandler) ResetQueue(ctx context.Context, request *shared.ResetQueueRequest) (err error) {
	err = t.h.ResetQueue(ctx, request)
	return thrift.FromError(err)
}

// ResetStickyTaskList forwards request to the underlying handler
func (t ThriftHandler) ResetStickyTaskList(ctx context.Context, request *h.ResetStickyTaskListRequest) (response *h.ResetStickyTaskListResponse, err error) {
	response, err = t.h.ResetStickyTaskList(ctx, request)
	return response, thrift.FromError(err)
}

// ResetWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) ResetWorkflowExecution(ctx context.Context, request *h.ResetWorkflowExecutionRequest) (response *shared.ResetWorkflowExecutionResponse, err error) {
	response, err = t.h.ResetWorkflowExecution(ctx, request)
	return response, thrift.FromError(err)
}

// RespondActivityTaskCanceled forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCanceled(ctx context.Context, request *h.RespondActivityTaskCanceledRequest) (err error) {
	err = t.h.RespondActivityTaskCanceled(ctx, request)
	return thrift.FromError(err)
}

// RespondActivityTaskCompleted forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCompleted(ctx context.Context, request *h.RespondActivityTaskCompletedRequest) (err error) {
	err = t.h.RespondActivityTaskCompleted(ctx, request)
	return thrift.FromError(err)
}

// RespondActivityTaskFailed forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskFailed(ctx context.Context, request *h.RespondActivityTaskFailedRequest) (err error) {
	err = t.h.RespondActivityTaskFailed(ctx, request)
	return thrift.FromError(err)
}

// RespondDecisionTaskCompleted forwards request to the underlying handler
func (t ThriftHandler) RespondDecisionTaskCompleted(ctx context.Context, request *h.RespondDecisionTaskCompletedRequest) (response *h.RespondDecisionTaskCompletedResponse, err error) {
	response, err = t.h.RespondDecisionTaskCompleted(ctx, request)
	return response, thrift.FromError(err)
}

// RespondDecisionTaskFailed forwards request to the underlying handler
func (t ThriftHandler) RespondDecisionTaskFailed(ctx context.Context, request *h.RespondDecisionTaskFailedRequest) (err error) {
	err = t.h.RespondDecisionTaskFailed(ctx, request)
	return thrift.FromError(err)
}

// ScheduleDecisionTask forwards request to the underlying handler
func (t ThriftHandler) ScheduleDecisionTask(ctx context.Context, request *h.ScheduleDecisionTaskRequest) (err error) {
	err = t.h.ScheduleDecisionTask(ctx, request)
	return thrift.FromError(err)
}

// SignalWithStartWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) SignalWithStartWorkflowExecution(ctx context.Context, request *h.SignalWithStartWorkflowExecutionRequest) (response *shared.StartWorkflowExecutionResponse, err error) {
	response, err = t.h.SignalWithStartWorkflowExecution(ctx, request)
	return response, thrift.FromError(err)
}

// SignalWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) SignalWorkflowExecution(ctx context.Context, request *h.SignalWorkflowExecutionRequest) (err error) {
	err = t.h.SignalWorkflowExecution(ctx, request)
	return thrift.FromError(err)
}

// StartWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) StartWorkflowExecution(ctx context.Context, request *h.StartWorkflowExecutionRequest) (response *shared.StartWorkflowExecutionResponse, err error) {
	response, err = t.h.StartWorkflowExecution(ctx, request)
	return response, thrift.FromError(err)
}

// SyncActivity forwards request to the underlying handler
func (t ThriftHandler) SyncActivity(ctx context.Context, request *h.SyncActivityRequest) (err error) {
	err = t.h.SyncActivity(ctx, request)
	return thrift.FromError(err)
}

// SyncShardStatus forwards request to the underlying handler
func (t ThriftHandler) SyncShardStatus(ctx context.Context, request *h.SyncShardStatusRequest) (err error) {
	err = t.h.SyncShardStatus(ctx, request)
	return thrift.FromError(err)
}

// TerminateWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) TerminateWorkflowExecution(ctx context.Context, request *h.TerminateWorkflowExecutionRequest) (err error) {
	err = t.h.TerminateWorkflowExecution(ctx, request)
	return thrift.FromError(err)
}
