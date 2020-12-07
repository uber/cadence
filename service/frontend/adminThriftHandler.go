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

package frontend

import (
	"context"

	"go.uber.org/yarpc"

	"github.com/uber/cadence/.gen/go/admin"
	"github.com/uber/cadence/.gen/go/admin/adminserviceserver"
	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/types/mapper/thrift"
)

// AdminThriftHandler wrap underlying handler and handles Thrift related type conversions
type AdminThriftHandler struct {
	h AdminHandler
}

// NewAdminThriftHandler creates Thrift handler on top of underlying handler
func NewAdminThriftHandler(h AdminHandler) AdminThriftHandler {
	return AdminThriftHandler{h}
}

func (t AdminThriftHandler) register(dispatcher *yarpc.Dispatcher) {
	dispatcher.Register(adminserviceserver.New(t))
}

// AddSearchAttribute forwards request to the underlying handler
func (t AdminThriftHandler) AddSearchAttribute(ctx context.Context, request *admin.AddSearchAttributeRequest) error {
	err := t.h.AddSearchAttribute(ctx, thrift.ToAddSearchAttributeRequest(request))
	return thrift.FromError(err)
}

// CloseShard forwards request to the underlying handler
func (t AdminThriftHandler) CloseShard(ctx context.Context, request *shared.CloseShardRequest) error {
	err := t.h.CloseShard(ctx, thrift.ToCloseShardRequest(request))
	return thrift.FromError(err)
}

// DescribeCluster forwards request to the underlying handler
func (t AdminThriftHandler) DescribeCluster(ctx context.Context) (*admin.DescribeClusterResponse, error) {
	response, err := t.h.DescribeCluster(ctx)
	return thrift.FromDescribeClusterResponse(response), thrift.FromError(err)
}

// DescribeHistoryHost forwards request to the underlying handler
func (t AdminThriftHandler) DescribeHistoryHost(ctx context.Context, request *shared.DescribeHistoryHostRequest) (*shared.DescribeHistoryHostResponse, error) {
	response, err := t.h.DescribeHistoryHost(ctx, thrift.ToDescribeHistoryHostRequest(request))
	return thrift.FromDescribeHistoryHostResponse(response), thrift.FromError(err)
}

// DescribeQueue forwards request to the underlying handler
func (t AdminThriftHandler) DescribeQueue(ctx context.Context, request *shared.DescribeQueueRequest) (*shared.DescribeQueueResponse, error) {
	response, err := t.h.DescribeQueue(ctx, thrift.ToDescribeQueueRequest(request))
	return thrift.FromDescribeQueueResponse(response), thrift.FromError(err)
}

// DescribeWorkflowExecution forwards request to the underlying handler
func (t AdminThriftHandler) DescribeWorkflowExecution(ctx context.Context, request *admin.DescribeWorkflowExecutionRequest) (*admin.DescribeWorkflowExecutionResponse, error) {
	response, err := t.h.DescribeWorkflowExecution(ctx, thrift.ToAdminDescribeWorkflowExecutionRequest(request))
	return thrift.FromAdminDescribeWorkflowExecutionResponse(response), thrift.FromError(err)
}

// GetDLQReplicationMessages forwards request to the underlying handler
func (t AdminThriftHandler) GetDLQReplicationMessages(ctx context.Context, request *replicator.GetDLQReplicationMessagesRequest) (*replicator.GetDLQReplicationMessagesResponse, error) {
	response, err := t.h.GetDLQReplicationMessages(ctx, thrift.ToGetDLQReplicationMessagesRequest(request))
	return thrift.FromGetDLQReplicationMessagesResponse(response), thrift.FromError(err)
}

// GetDomainReplicationMessages forwards request to the underlying handler
func (t AdminThriftHandler) GetDomainReplicationMessages(ctx context.Context, request *replicator.GetDomainReplicationMessagesRequest) (*replicator.GetDomainReplicationMessagesResponse, error) {
	response, err := t.h.GetDomainReplicationMessages(ctx, thrift.ToGetDomainReplicationMessagesRequest(request))
	return thrift.FromGetDomainReplicationMessagesResponse(response), thrift.FromError(err)
}

// GetReplicationMessages forwards request to the underlying handler
func (t AdminThriftHandler) GetReplicationMessages(ctx context.Context, request *replicator.GetReplicationMessagesRequest) (*replicator.GetReplicationMessagesResponse, error) {
	response, err := t.h.GetReplicationMessages(ctx, thrift.ToGetReplicationMessagesRequest(request))
	return thrift.FromGetReplicationMessagesResponse(response), thrift.FromError(err)
}

// GetWorkflowExecutionRawHistoryV2 forwards request to the underlying handler
func (t AdminThriftHandler) GetWorkflowExecutionRawHistoryV2(ctx context.Context, request *admin.GetWorkflowExecutionRawHistoryV2Request) (*admin.GetWorkflowExecutionRawHistoryV2Response, error) {
	response, err := t.h.GetWorkflowExecutionRawHistoryV2(ctx, thrift.ToGetWorkflowExecutionRawHistoryV2Request(request))
	return thrift.FromGetWorkflowExecutionRawHistoryV2Response(response), thrift.FromError(err)
}

// MergeDLQMessages forwards request to the underlying handler
func (t AdminThriftHandler) MergeDLQMessages(ctx context.Context, request *replicator.MergeDLQMessagesRequest) (*replicator.MergeDLQMessagesResponse, error) {
	response, err := t.h.MergeDLQMessages(ctx, thrift.ToMergeDLQMessagesRequest(request))
	return thrift.FromMergeDLQMessagesResponse(response), thrift.FromError(err)
}

// PurgeDLQMessages forwards request to the underlying handler
func (t AdminThriftHandler) PurgeDLQMessages(ctx context.Context, request *replicator.PurgeDLQMessagesRequest) error {
	err := t.h.PurgeDLQMessages(ctx, thrift.ToPurgeDLQMessagesRequest(request))
	return thrift.FromError(err)
}

// ReadDLQMessages forwards request to the underlying handler
func (t AdminThriftHandler) ReadDLQMessages(ctx context.Context, request *replicator.ReadDLQMessagesRequest) (*replicator.ReadDLQMessagesResponse, error) {
	response, err := t.h.ReadDLQMessages(ctx, thrift.ToReadDLQMessagesRequest(request))
	return thrift.FromReadDLQMessagesResponse(response), thrift.FromError(err)
}

// ReapplyEvents forwards request to the underlying handler
func (t AdminThriftHandler) ReapplyEvents(ctx context.Context, request *shared.ReapplyEventsRequest) error {
	err := t.h.ReapplyEvents(ctx, thrift.ToReapplyEventsRequest(request))
	return thrift.FromError(err)
}

// RefreshWorkflowTasks forwards request to the underlying handler
func (t AdminThriftHandler) RefreshWorkflowTasks(ctx context.Context, request *shared.RefreshWorkflowTasksRequest) error {
	err := t.h.RefreshWorkflowTasks(ctx, thrift.ToRefreshWorkflowTasksRequest(request))
	return thrift.FromError(err)
}

// RemoveTask forwards request to the underlying handler
func (t AdminThriftHandler) RemoveTask(ctx context.Context, request *shared.RemoveTaskRequest) error {
	err := t.h.RemoveTask(ctx, thrift.ToRemoveTaskRequest(request))
	return thrift.FromError(err)
}

// ResendReplicationTasks forwards request to the underlying handler
func (t AdminThriftHandler) ResendReplicationTasks(ctx context.Context, request *admin.ResendReplicationTasksRequest) error {
	err := t.h.ResendReplicationTasks(ctx, thrift.ToResendReplicationTasksRequest(request))
	return thrift.FromError(err)
}

// ResetQueue forwards request to the underlying handler
func (t AdminThriftHandler) ResetQueue(ctx context.Context, request *shared.ResetQueueRequest) error {
	err := t.h.ResetQueue(ctx, thrift.ToResetQueueRequest(request))
	return thrift.FromError(err)
}
