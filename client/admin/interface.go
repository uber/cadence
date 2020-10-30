// Copyright (c) 2017 Uber Technologies, Inc.
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
	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/.gen/go/shared"
)

// Client is the interface exposed by admin service client
type Client interface {
	AddSearchAttribute(context.Context, *admin.AddSearchAttributeRequest, ...yarpc.CallOption) error
	CloseShard(context.Context, *shared.CloseShardRequest, ...yarpc.CallOption) error
	DescribeCluster(context.Context, ...yarpc.CallOption) (*admin.DescribeClusterResponse, error)
	DescribeHistoryHost(context.Context, *shared.DescribeHistoryHostRequest, ...yarpc.CallOption) (*shared.DescribeHistoryHostResponse, error)
	DescribeQueue(context.Context, *shared.DescribeQueueRequest, ...yarpc.CallOption) (*shared.DescribeQueueResponse, error)
	DescribeWorkflowExecution(context.Context, *admin.DescribeWorkflowExecutionRequest, ...yarpc.CallOption) (*admin.DescribeWorkflowExecutionResponse, error)
	GetDLQReplicationMessages(context.Context, *replicator.GetDLQReplicationMessagesRequest, ...yarpc.CallOption) (*replicator.GetDLQReplicationMessagesResponse, error)
	GetDomainReplicationMessages(context.Context, *replicator.GetDomainReplicationMessagesRequest, ...yarpc.CallOption) (*replicator.GetDomainReplicationMessagesResponse, error)
	GetReplicationMessages(context.Context, *replicator.GetReplicationMessagesRequest, ...yarpc.CallOption) (*replicator.GetReplicationMessagesResponse, error)
	GetWorkflowExecutionRawHistoryV2(context.Context, *admin.GetWorkflowExecutionRawHistoryV2Request, ...yarpc.CallOption) (*admin.GetWorkflowExecutionRawHistoryV2Response, error)
	MergeDLQMessages(context.Context, *replicator.MergeDLQMessagesRequest, ...yarpc.CallOption) (*replicator.MergeDLQMessagesResponse, error)
	PurgeDLQMessages(context.Context, *replicator.PurgeDLQMessagesRequest, ...yarpc.CallOption) error
	ReadDLQMessages(context.Context, *replicator.ReadDLQMessagesRequest, ...yarpc.CallOption) (*replicator.ReadDLQMessagesResponse, error)
	ReapplyEvents(context.Context, *shared.ReapplyEventsRequest, ...yarpc.CallOption) error
	RefreshWorkflowTasks(context.Context, *shared.RefreshWorkflowTasksRequest, ...yarpc.CallOption) error
	RemoveTask(context.Context, *shared.RemoveTaskRequest, ...yarpc.CallOption) error
	ResendReplicationTasks(context.Context, *admin.ResendReplicationTasksRequest, ...yarpc.CallOption) error
	ResetQueue(context.Context, *shared.ResetQueueRequest, ...yarpc.CallOption) error
}
