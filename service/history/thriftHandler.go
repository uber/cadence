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
func (t ThriftHandler) Health(ctx context.Context) (*health.HealthStatus, error) {
	response, err := t.h.Health(ctx)
	return thrift.FromHealthStatus(response), thrift.FromError(err)
}

// CloseShard forwards request to the underlying handler
func (t ThriftHandler) CloseShard(ctx context.Context, request *shared.CloseShardRequest) error {
	err := t.h.CloseShard(ctx, thrift.ToCloseShardRequest(request))
	return thrift.FromError(err)
}

// DescribeHistoryHost forwards request to the underlying handler
func (t ThriftHandler) DescribeHistoryHost(ctx context.Context, request *shared.DescribeHistoryHostRequest) (*shared.DescribeHistoryHostResponse, error) {
	response, err := t.h.DescribeHistoryHost(ctx, thrift.ToDescribeHistoryHostRequest(request))
	return thrift.FromDescribeHistoryHostResponse(response), thrift.FromError(err)
}

// DescribeMutableState forwards request to the underlying handler
func (t ThriftHandler) DescribeMutableState(ctx context.Context, request *h.DescribeMutableStateRequest) (*h.DescribeMutableStateResponse, error) {
	response, err := t.h.DescribeMutableState(ctx, thrift.ToDescribeMutableStateRequest(request))
	return thrift.FromDescribeMutableStateResponse(response), thrift.FromError(err)
}

// DescribeQueue forwards request to the underlying handler
func (t ThriftHandler) DescribeQueue(ctx context.Context, request *shared.DescribeQueueRequest) (*shared.DescribeQueueResponse, error) {
	response, err := t.h.DescribeQueue(ctx, thrift.ToDescribeQueueRequest(request))
	return thrift.FromDescribeQueueResponse(response), thrift.FromError(err)
}

// DescribeWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) DescribeWorkflowExecution(ctx context.Context, request *h.DescribeWorkflowExecutionRequest) (*shared.DescribeWorkflowExecutionResponse, error) {
	response, err := t.h.DescribeWorkflowExecution(ctx, thrift.ToHistoryDescribeWorkflowExecutionRequest(request))
	return thrift.FromDescribeWorkflowExecutionResponse(response), thrift.FromError(err)
}

// GetDLQReplicationMessages forwards request to the underlying handler
func (t ThriftHandler) GetDLQReplicationMessages(ctx context.Context, request *replicator.GetDLQReplicationMessagesRequest) (*replicator.GetDLQReplicationMessagesResponse, error) {
	response, err := t.h.GetDLQReplicationMessages(ctx, thrift.ToGetDLQReplicationMessagesRequest(request))
	return thrift.FromGetDLQReplicationMessagesResponse(response), thrift.FromError(err)
}

// GetMutableState forwards request to the underlying handler
func (t ThriftHandler) GetMutableState(ctx context.Context, request *h.GetMutableStateRequest) (*h.GetMutableStateResponse, error) {
	response, err := t.h.GetMutableState(ctx, thrift.ToGetMutableStateRequest(request))
	return thrift.FromGetMutableStateResponse(response), thrift.FromError(err)
}

// GetReplicationMessages forwards request to the underlying handler
func (t ThriftHandler) GetReplicationMessages(ctx context.Context, request *replicator.GetReplicationMessagesRequest) (*replicator.GetReplicationMessagesResponse, error) {
	response, err := t.h.GetReplicationMessages(ctx, thrift.ToGetReplicationMessagesRequest(request))
	return thrift.FromGetReplicationMessagesResponse(response), thrift.FromError(err)
}

// MergeDLQMessages forwards request to the underlying handler
func (t ThriftHandler) MergeDLQMessages(ctx context.Context, request *replicator.MergeDLQMessagesRequest) (*replicator.MergeDLQMessagesResponse, error) {
	response, err := t.h.MergeDLQMessages(ctx, thrift.ToMergeDLQMessagesRequest(request))
	return thrift.FromMergeDLQMessagesResponse(response), thrift.FromError(err)
}

// NotifyFailoverMarkers forwards request to the underlying handler
func (t ThriftHandler) NotifyFailoverMarkers(ctx context.Context, request *h.NotifyFailoverMarkersRequest) error {
	err := t.h.NotifyFailoverMarkers(ctx, thrift.ToNotifyFailoverMarkersRequest(request))
	return thrift.FromError(err)
}

// PollMutableState forwards request to the underlying handler
func (t ThriftHandler) PollMutableState(ctx context.Context, request *h.PollMutableStateRequest) (*h.PollMutableStateResponse, error) {
	response, err := t.h.PollMutableState(ctx, thrift.ToPollMutableStateRequest(request))
	return thrift.FromPollMutableStateResponse(response), thrift.FromError(err)
}

// PurgeDLQMessages forwards request to the underlying handler
func (t ThriftHandler) PurgeDLQMessages(ctx context.Context, request *replicator.PurgeDLQMessagesRequest) error {
	err := t.h.PurgeDLQMessages(ctx, thrift.ToPurgeDLQMessagesRequest(request))
	return thrift.FromError(err)
}

// QueryWorkflow forwards request to the underlying handler
func (t ThriftHandler) QueryWorkflow(ctx context.Context, request *h.QueryWorkflowRequest) (*h.QueryWorkflowResponse, error) {
	response, err := t.h.QueryWorkflow(ctx, thrift.ToHistoryQueryWorkflowRequest(request))
	return thrift.FromHistoryQueryWorkflowResponse(response), thrift.FromError(err)
}

// ReadDLQMessages forwards request to the underlying handler
func (t ThriftHandler) ReadDLQMessages(ctx context.Context, request *replicator.ReadDLQMessagesRequest) (*replicator.ReadDLQMessagesResponse, error) {
	response, err := t.h.ReadDLQMessages(ctx, thrift.ToReadDLQMessagesRequest(request))
	return thrift.FromReadDLQMessagesResponse(response), thrift.FromError(err)
}

// ReapplyEvents forwards request to the underlying handler
func (t ThriftHandler) ReapplyEvents(ctx context.Context, request *h.ReapplyEventsRequest) error {
	err := t.h.ReapplyEvents(ctx, thrift.ToHistoryReapplyEventsRequest(request))
	return thrift.FromError(err)
}

// RecordActivityTaskHeartbeat forwards request to the underlying handler
func (t ThriftHandler) RecordActivityTaskHeartbeat(ctx context.Context, request *h.RecordActivityTaskHeartbeatRequest) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	response, err := t.h.RecordActivityTaskHeartbeat(ctx, thrift.ToHistoryRecordActivityTaskHeartbeatRequest(request))
	return thrift.FromRecordActivityTaskHeartbeatResponse(response), thrift.FromError(err)
}

// RecordActivityTaskStarted forwards request to the underlying handler
func (t ThriftHandler) RecordActivityTaskStarted(ctx context.Context, request *h.RecordActivityTaskStartedRequest) (*h.RecordActivityTaskStartedResponse, error) {
	response, err := t.h.RecordActivityTaskStarted(ctx, thrift.ToRecordActivityTaskStartedRequest(request))
	return thrift.FromRecordActivityTaskStartedResponse(response), thrift.FromError(err)
}

// RecordChildExecutionCompleted forwards request to the underlying handler
func (t ThriftHandler) RecordChildExecutionCompleted(ctx context.Context, request *h.RecordChildExecutionCompletedRequest) error {
	err := t.h.RecordChildExecutionCompleted(ctx, thrift.ToRecordChildExecutionCompletedRequest(request))
	return thrift.FromError(err)
}

// RecordDecisionTaskStarted forwards request to the underlying handler
func (t ThriftHandler) RecordDecisionTaskStarted(ctx context.Context, request *h.RecordDecisionTaskStartedRequest) (*h.RecordDecisionTaskStartedResponse, error) {
	response, err := t.h.RecordDecisionTaskStarted(ctx, thrift.ToRecordDecisionTaskStartedRequest(request))
	return thrift.FromRecordDecisionTaskStartedResponse(response), thrift.FromError(err)
}

// RefreshWorkflowTasks forwards request to the underlying handler
func (t ThriftHandler) RefreshWorkflowTasks(ctx context.Context, request *h.RefreshWorkflowTasksRequest) error {
	err := t.h.RefreshWorkflowTasks(ctx, thrift.ToHistoryRefreshWorkflowTasksRequest(request))
	return thrift.FromError(err)
}

// RemoveSignalMutableState forwards request to the underlying handler
func (t ThriftHandler) RemoveSignalMutableState(ctx context.Context, request *h.RemoveSignalMutableStateRequest) error {
	err := t.h.RemoveSignalMutableState(ctx, thrift.ToRemoveSignalMutableStateRequest(request))
	return thrift.FromError(err)
}

// RemoveTask forwards request to the underlying handler
func (t ThriftHandler) RemoveTask(ctx context.Context, request *shared.RemoveTaskRequest) error {
	err := t.h.RemoveTask(ctx, thrift.ToRemoveTaskRequest(request))
	return thrift.FromError(err)
}

// ReplicateEventsV2 forwards request to the underlying handler
func (t ThriftHandler) ReplicateEventsV2(ctx context.Context, request *h.ReplicateEventsV2Request) error {
	err := t.h.ReplicateEventsV2(ctx, thrift.ToReplicateEventsV2Request(request))
	return thrift.FromError(err)
}

// RequestCancelWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) RequestCancelWorkflowExecution(ctx context.Context, request *h.RequestCancelWorkflowExecutionRequest) error {
	err := t.h.RequestCancelWorkflowExecution(ctx, thrift.ToHistoryRequestCancelWorkflowExecutionRequest(request))
	return thrift.FromError(err)
}

// ResetQueue forwards request to the underlying handler
func (t ThriftHandler) ResetQueue(ctx context.Context, request *shared.ResetQueueRequest) error {
	err := t.h.ResetQueue(ctx, thrift.ToResetQueueRequest(request))
	return thrift.FromError(err)
}

// ResetStickyTaskList forwards request to the underlying handler
func (t ThriftHandler) ResetStickyTaskList(ctx context.Context, request *h.ResetStickyTaskListRequest) (*h.ResetStickyTaskListResponse, error) {
	response, err := t.h.ResetStickyTaskList(ctx, thrift.ToHistoryResetStickyTaskListRequest(request))
	return thrift.FromHistoryResetStickyTaskListResponse(response), thrift.FromError(err)
}

// ResetWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) ResetWorkflowExecution(ctx context.Context, request *h.ResetWorkflowExecutionRequest) (*shared.ResetWorkflowExecutionResponse, error) {
	response, err := t.h.ResetWorkflowExecution(ctx, thrift.ToHistoryResetWorkflowExecutionRequest(request))
	return thrift.FromResetWorkflowExecutionResponse(response), thrift.FromError(err)
}

// RespondActivityTaskCanceled forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCanceled(ctx context.Context, request *h.RespondActivityTaskCanceledRequest) error {
	err := t.h.RespondActivityTaskCanceled(ctx, thrift.ToHistoryRespondActivityTaskCanceledRequest(request))
	return thrift.FromError(err)
}

// RespondActivityTaskCompleted forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCompleted(ctx context.Context, request *h.RespondActivityTaskCompletedRequest) error {
	err := t.h.RespondActivityTaskCompleted(ctx, thrift.ToHistoryRespondActivityTaskCompletedRequest(request))
	return thrift.FromError(err)
}

// RespondActivityTaskFailed forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskFailed(ctx context.Context, request *h.RespondActivityTaskFailedRequest) error {
	err := t.h.RespondActivityTaskFailed(ctx, thrift.ToHistoryRespondActivityTaskFailedRequest(request))
	return thrift.FromError(err)
}

// RespondDecisionTaskCompleted forwards request to the underlying handler
func (t ThriftHandler) RespondDecisionTaskCompleted(ctx context.Context, request *h.RespondDecisionTaskCompletedRequest) (*h.RespondDecisionTaskCompletedResponse, error) {
	response, err := t.h.RespondDecisionTaskCompleted(ctx, thrift.ToHistoryRespondDecisionTaskCompletedRequest(request))
	return thrift.FromHistoryRespondDecisionTaskCompletedResponse(response), thrift.FromError(err)
}

// RespondDecisionTaskFailed forwards request to the underlying handler
func (t ThriftHandler) RespondDecisionTaskFailed(ctx context.Context, request *h.RespondDecisionTaskFailedRequest) error {
	err := t.h.RespondDecisionTaskFailed(ctx, thrift.ToHistoryRespondDecisionTaskFailedRequest(request))
	return thrift.FromError(err)
}

// ScheduleDecisionTask forwards request to the underlying handler
func (t ThriftHandler) ScheduleDecisionTask(ctx context.Context, request *h.ScheduleDecisionTaskRequest) error {
	err := t.h.ScheduleDecisionTask(ctx, thrift.ToScheduleDecisionTaskRequest(request))
	return thrift.FromError(err)
}

// SignalWithStartWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) SignalWithStartWorkflowExecution(ctx context.Context, request *h.SignalWithStartWorkflowExecutionRequest) (*shared.StartWorkflowExecutionResponse, error) {
	response, err := t.h.SignalWithStartWorkflowExecution(ctx, thrift.ToHistorySignalWithStartWorkflowExecutionRequest(request))
	return thrift.FromStartWorkflowExecutionResponse(response), thrift.FromError(err)
}

// SignalWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) SignalWorkflowExecution(ctx context.Context, request *h.SignalWorkflowExecutionRequest) error {
	err := t.h.SignalWorkflowExecution(ctx, thrift.ToHistorySignalWorkflowExecutionRequest(request))
	return thrift.FromError(err)
}

// StartWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) StartWorkflowExecution(ctx context.Context, request *h.StartWorkflowExecutionRequest) (*shared.StartWorkflowExecutionResponse, error) {
	response, err := t.h.StartWorkflowExecution(ctx, thrift.ToHistoryStartWorkflowExecutionRequest(request))
	return thrift.FromStartWorkflowExecutionResponse(response), thrift.FromError(err)
}

// SyncActivity forwards request to the underlying handler
func (t ThriftHandler) SyncActivity(ctx context.Context, request *h.SyncActivityRequest) error {
	err := t.h.SyncActivity(ctx, thrift.ToSyncActivityRequest(request))
	return thrift.FromError(err)
}

// SyncShardStatus forwards request to the underlying handler
func (t ThriftHandler) SyncShardStatus(ctx context.Context, request *h.SyncShardStatusRequest) error {
	err := t.h.SyncShardStatus(ctx, thrift.ToSyncShardStatusRequest(request))
	return thrift.FromError(err)
}

// TerminateWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) TerminateWorkflowExecution(ctx context.Context, request *h.TerminateWorkflowExecutionRequest) error {
	err := t.h.TerminateWorkflowExecution(ctx, thrift.ToHistoryTerminateWorkflowExecutionRequest(request))
	return thrift.FromError(err)
}
