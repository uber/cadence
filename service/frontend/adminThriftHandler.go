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
	return t.h.AddSearchAttribute(ctx, request)
}

// CloseShard forwards request to the underlying handler
func (t AdminThriftHandler) CloseShard(ctx context.Context, request *shared.CloseShardRequest) (err error) {
	return t.h.CloseShard(ctx, request)
}

// DescribeCluster forwards request to the underlying handler
func (t AdminThriftHandler) DescribeCluster(ctx context.Context) (response *admin.DescribeClusterResponse, err error) {
	return t.h.DescribeCluster(ctx)
}

// DescribeHistoryHost forwards request to the underlying handler
func (t AdminThriftHandler) DescribeHistoryHost(ctx context.Context, request *shared.DescribeHistoryHostRequest) (response *shared.DescribeHistoryHostResponse, err error) {
	return t.h.DescribeHistoryHost(ctx, request)
}

// DescribeQueue forwards request to the underlying handler
func (t AdminThriftHandler) DescribeQueue(ctx context.Context, request *shared.DescribeQueueRequest) (response *shared.DescribeQueueResponse, err error) {
	return t.h.DescribeQueue(ctx, request)
}

// DescribeWorkflowExecution forwards request to the underlying handler
func (t AdminThriftHandler) DescribeWorkflowExecution(ctx context.Context, request *admin.DescribeWorkflowExecutionRequest) (response *admin.DescribeWorkflowExecutionResponse, err error) {
	return t.h.DescribeWorkflowExecution(ctx, request)
}

// GetDLQReplicationMessages forwards request to the underlying handler
func (t AdminThriftHandler) GetDLQReplicationMessages(ctx context.Context, request *replicator.GetDLQReplicationMessagesRequest) (response *replicator.GetDLQReplicationMessagesResponse, err error) {
	return t.h.GetDLQReplicationMessages(ctx, request)
}

// GetDomainReplicationMessages forwards request to the underlying handler
func (t AdminThriftHandler) GetDomainReplicationMessages(ctx context.Context, request *replicator.GetDomainReplicationMessagesRequest) (response *replicator.GetDomainReplicationMessagesResponse, err error) {
	return t.h.GetDomainReplicationMessages(ctx, request)
}

// GetReplicationMessages forwards request to the underlying handler
func (t AdminThriftHandler) GetReplicationMessages(ctx context.Context, request *replicator.GetReplicationMessagesRequest) (response *replicator.GetReplicationMessagesResponse, err error) {
	return t.h.GetReplicationMessages(ctx, request)
}

// GetWorkflowExecutionRawHistoryV2 forwards request to the underlying handler
func (t AdminThriftHandler) GetWorkflowExecutionRawHistoryV2(ctx context.Context, request *admin.GetWorkflowExecutionRawHistoryV2Request) (response *admin.GetWorkflowExecutionRawHistoryV2Response, err error) {
	return t.h.GetWorkflowExecutionRawHistoryV2(ctx, request)
}

// MergeDLQMessages forwards request to the underlying handler
func (t AdminThriftHandler) MergeDLQMessages(ctx context.Context, request *replicator.MergeDLQMessagesRequest) (response *replicator.MergeDLQMessagesResponse, err error) {
	return t.h.MergeDLQMessages(ctx, request)
}

// PurgeDLQMessages forwards request to the underlying handler
func (t AdminThriftHandler) PurgeDLQMessages(ctx context.Context, request *replicator.PurgeDLQMessagesRequest) (err error) {
	return t.h.PurgeDLQMessages(ctx, request)
}

// ReadDLQMessages forwards request to the underlying handler
func (t AdminThriftHandler) ReadDLQMessages(ctx context.Context, request *replicator.ReadDLQMessagesRequest) (response *replicator.ReadDLQMessagesResponse, err error) {
	return t.h.ReadDLQMessages(ctx, request)
}

// ReapplyEvents forwards request to the underlying handler
func (t AdminThriftHandler) ReapplyEvents(ctx context.Context, request *shared.ReapplyEventsRequest) (err error) {
	return t.h.ReapplyEvents(ctx, request)
}

// RefreshWorkflowTasks forwards request to the underlying handler
func (t AdminThriftHandler) RefreshWorkflowTasks(ctx context.Context, request *shared.RefreshWorkflowTasksRequest) (err error) {
	return t.h.RefreshWorkflowTasks(ctx, request)
}

// RemoveTask forwards request to the underlying handler
func (t AdminThriftHandler) RemoveTask(ctx context.Context, request *shared.RemoveTaskRequest) (err error) {
	return t.h.RemoveTask(ctx, request)
}

// ResendReplicationTasks forwards request to the underlying handler
func (t AdminThriftHandler) ResendReplicationTasks(ctx context.Context, request *admin.ResendReplicationTasksRequest) (err error) {
	return t.h.ResendReplicationTasks(ctx, request)
}

// ResetQueue forwards request to the underlying handler
func (t AdminThriftHandler) ResetQueue(ctx context.Context, request *shared.ResetQueueRequest) (err error) {
	return t.h.ResetQueue(ctx, request)
}
