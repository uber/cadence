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
func (t AdminThriftHandler) AddSearchAttribute(ctx context.Context, request *admin.AddSearchAttributeRequest) (err error) {
	err = t.h.AddSearchAttribute(ctx, request)
	return thrift.FromError(err)
}

// CloseShard forwards request to the underlying handler
func (t AdminThriftHandler) CloseShard(ctx context.Context, request *shared.CloseShardRequest) (err error) {
	err = t.h.CloseShard(ctx, request)
	return thrift.FromError(err)
}

// DescribeCluster forwards request to the underlying handler
func (t AdminThriftHandler) DescribeCluster(ctx context.Context) (response *admin.DescribeClusterResponse, err error) {
	response, err = t.h.DescribeCluster(ctx)
	return response, thrift.FromError(err)
}

// DescribeHistoryHost forwards request to the underlying handler
func (t AdminThriftHandler) DescribeHistoryHost(ctx context.Context, request *shared.DescribeHistoryHostRequest) (response *shared.DescribeHistoryHostResponse, err error) {
	response, err = t.h.DescribeHistoryHost(ctx, request)
	return response, thrift.FromError(err)
}

// DescribeQueue forwards request to the underlying handler
func (t AdminThriftHandler) DescribeQueue(ctx context.Context, request *shared.DescribeQueueRequest) (response *shared.DescribeQueueResponse, err error) {
	response, err = t.h.DescribeQueue(ctx, request)
	return response, thrift.FromError(err)
}

// DescribeWorkflowExecution forwards request to the underlying handler
func (t AdminThriftHandler) DescribeWorkflowExecution(ctx context.Context, request *admin.DescribeWorkflowExecutionRequest) (response *admin.DescribeWorkflowExecutionResponse, err error) {
	response, err = t.h.DescribeWorkflowExecution(ctx, request)
	return response, thrift.FromError(err)
}

// GetDLQReplicationMessages forwards request to the underlying handler
func (t AdminThriftHandler) GetDLQReplicationMessages(ctx context.Context, request *replicator.GetDLQReplicationMessagesRequest) (response *replicator.GetDLQReplicationMessagesResponse, err error) {
	response, err = t.h.GetDLQReplicationMessages(ctx, request)
	return response, thrift.FromError(err)
}

// GetDomainReplicationMessages forwards request to the underlying handler
func (t AdminThriftHandler) GetDomainReplicationMessages(ctx context.Context, request *replicator.GetDomainReplicationMessagesRequest) (response *replicator.GetDomainReplicationMessagesResponse, err error) {
	response, err = t.h.GetDomainReplicationMessages(ctx, request)
	return response, thrift.FromError(err)
}

// GetReplicationMessages forwards request to the underlying handler
func (t AdminThriftHandler) GetReplicationMessages(ctx context.Context, request *replicator.GetReplicationMessagesRequest) (response *replicator.GetReplicationMessagesResponse, err error) {
	response, err = t.h.GetReplicationMessages(ctx, request)
	return response, thrift.FromError(err)
}

// GetWorkflowExecutionRawHistoryV2 forwards request to the underlying handler
func (t AdminThriftHandler) GetWorkflowExecutionRawHistoryV2(ctx context.Context, request *admin.GetWorkflowExecutionRawHistoryV2Request) (response *admin.GetWorkflowExecutionRawHistoryV2Response, err error) {
	response, err = t.h.GetWorkflowExecutionRawHistoryV2(ctx, request)
	return response, thrift.FromError(err)
}

// MergeDLQMessages forwards request to the underlying handler
func (t AdminThriftHandler) MergeDLQMessages(ctx context.Context, request *replicator.MergeDLQMessagesRequest) (response *replicator.MergeDLQMessagesResponse, err error) {
	response, err = t.h.MergeDLQMessages(ctx, request)
	return response, thrift.FromError(err)
}

// PurgeDLQMessages forwards request to the underlying handler
func (t AdminThriftHandler) PurgeDLQMessages(ctx context.Context, request *replicator.PurgeDLQMessagesRequest) (err error) {
	err = t.h.PurgeDLQMessages(ctx, request)
	return thrift.FromError(err)
}

// ReadDLQMessages forwards request to the underlying handler
func (t AdminThriftHandler) ReadDLQMessages(ctx context.Context, request *replicator.ReadDLQMessagesRequest) (response *replicator.ReadDLQMessagesResponse, err error) {
	response, err = t.h.ReadDLQMessages(ctx, request)
	return response, thrift.FromError(err)
}

// ReapplyEvents forwards request to the underlying handler
func (t AdminThriftHandler) ReapplyEvents(ctx context.Context, request *shared.ReapplyEventsRequest) (err error) {
	err = t.h.ReapplyEvents(ctx, request)
	return thrift.FromError(err)
}

// RefreshWorkflowTasks forwards request to the underlying handler
func (t AdminThriftHandler) RefreshWorkflowTasks(ctx context.Context, request *shared.RefreshWorkflowTasksRequest) (err error) {
	err = t.h.RefreshWorkflowTasks(ctx, request)
	return thrift.FromError(err)
}

// RemoveTask forwards request to the underlying handler
func (t AdminThriftHandler) RemoveTask(ctx context.Context, request *shared.RemoveTaskRequest) (err error) {
	err = t.h.RemoveTask(ctx, request)
	return thrift.FromError(err)
}

// ResendReplicationTasks forwards request to the underlying handler
func (t AdminThriftHandler) ResendReplicationTasks(ctx context.Context, request *admin.ResendReplicationTasksRequest) (err error) {
	err = t.h.ResendReplicationTasks(ctx, request)
	return thrift.FromError(err)
}

// ResetQueue forwards request to the underlying handler
func (t AdminThriftHandler) ResetQueue(ctx context.Context, request *shared.ResetQueueRequest) (err error) {
	err = t.h.ResetQueue(ctx, request)
	return thrift.FromError(err)
}
