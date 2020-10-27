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

package admin

import (
	"context"

	"go.uber.org/yarpc"

	"github.com/uber/cadence/.gen/go/admin"
	"github.com/uber/cadence/.gen/go/admin/adminserviceclient"
	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/.gen/go/shared"
)

type thriftClient struct {
	c adminserviceclient.Interface
}

// NewThriftClient creates a new instance of Client with thrift protocol
func NewThriftClient(c adminserviceclient.Interface) Client {
	return thriftClient{c}
}

func (t thriftClient) AddSearchAttribute(ctx context.Context, request *admin.AddSearchAttributeRequest, opts ...yarpc.CallOption) error {
	return t.c.AddSearchAttribute(ctx, request, opts...)
}

func (t thriftClient) CloseShard(ctx context.Context, request *shared.CloseShardRequest, opts ...yarpc.CallOption) error {
	return t.c.CloseShard(ctx, request, opts...)
}

func (t thriftClient) DescribeCluster(ctx context.Context, opts ...yarpc.CallOption) (*admin.DescribeClusterResponse, error) {
	return t.c.DescribeCluster(ctx, opts...)
}

func (t thriftClient) DescribeHistoryHost(ctx context.Context, request *shared.DescribeHistoryHostRequest, opts ...yarpc.CallOption) (*shared.DescribeHistoryHostResponse, error) {
	return t.c.DescribeHistoryHost(ctx, request, opts...)
}

func (t thriftClient) DescribeQueue(ctx context.Context, request *shared.DescribeQueueRequest, opts ...yarpc.CallOption) (*shared.DescribeQueueResponse, error) {
	return t.c.DescribeQueue(ctx, request, opts...)
}

func (t thriftClient) DescribeWorkflowExecution(ctx context.Context, request *admin.DescribeWorkflowExecutionRequest, opts ...yarpc.CallOption) (*admin.DescribeWorkflowExecutionResponse, error) {
	return t.c.DescribeWorkflowExecution(ctx, request, opts...)
}

func (t thriftClient) GetDLQReplicationMessages(ctx context.Context, request *replicator.GetDLQReplicationMessagesRequest, opts ...yarpc.CallOption) (*replicator.GetDLQReplicationMessagesResponse, error) {
	return t.c.GetDLQReplicationMessages(ctx, request, opts...)
}

func (t thriftClient) GetDomainReplicationMessages(ctx context.Context, request *replicator.GetDomainReplicationMessagesRequest, opts ...yarpc.CallOption) (*replicator.GetDomainReplicationMessagesResponse, error) {
	return t.c.GetDomainReplicationMessages(ctx, request, opts...)
}

func (t thriftClient) GetReplicationMessages(ctx context.Context, request *replicator.GetReplicationMessagesRequest, opts ...yarpc.CallOption) (*replicator.GetReplicationMessagesResponse, error) {
	return t.c.GetReplicationMessages(ctx, request, opts...)
}

func (t thriftClient) GetWorkflowExecutionRawHistoryV2(ctx context.Context, request *admin.GetWorkflowExecutionRawHistoryV2Request, opts ...yarpc.CallOption) (*admin.GetWorkflowExecutionRawHistoryV2Response, error) {
	return t.c.GetWorkflowExecutionRawHistoryV2(ctx, request, opts...)
}

func (t thriftClient) MergeDLQMessages(ctx context.Context, request *replicator.MergeDLQMessagesRequest, opts ...yarpc.CallOption) (*replicator.MergeDLQMessagesResponse, error) {
	return t.c.MergeDLQMessages(ctx, request, opts...)
}

func (t thriftClient) PurgeDLQMessages(ctx context.Context, request *replicator.PurgeDLQMessagesRequest, opts ...yarpc.CallOption) error {
	return t.c.PurgeDLQMessages(ctx, request, opts...)
}

func (t thriftClient) ReadDLQMessages(ctx context.Context, request *replicator.ReadDLQMessagesRequest, opts ...yarpc.CallOption) (*replicator.ReadDLQMessagesResponse, error) {
	return t.c.ReadDLQMessages(ctx, request, opts...)
}

func (t thriftClient) ReapplyEvents(ctx context.Context, request *shared.ReapplyEventsRequest, opts ...yarpc.CallOption) error {
	return t.c.ReapplyEvents(ctx, request, opts...)
}

func (t thriftClient) RefreshWorkflowTasks(ctx context.Context, request *shared.RefreshWorkflowTasksRequest, opts ...yarpc.CallOption) error {
	return t.c.RefreshWorkflowTasks(ctx, request, opts...)
}

func (t thriftClient) RemoveTask(ctx context.Context, request *shared.RemoveTaskRequest, opts ...yarpc.CallOption) error {
	return t.c.RemoveTask(ctx, request, opts...)
}

func (t thriftClient) ResendReplicationTasks(ctx context.Context, request *admin.ResendReplicationTasksRequest, opts ...yarpc.CallOption) error {
	return t.c.ResendReplicationTasks(ctx, request, opts...)
}

func (t thriftClient) ResetQueue(ctx context.Context, request *shared.ResetQueueRequest, opts ...yarpc.CallOption) error {
	return t.c.ResetQueue(ctx, request, opts...)
}
