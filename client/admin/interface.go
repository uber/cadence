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

	"github.com/uber/cadence/common/types"
)

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination interface_mock.go -package admin github.com/uber/cadence/client/admin Client

// Client is the interface exposed by admin service client
type Client interface {
	AddSearchAttribute(context.Context, *types.AddSearchAttributeRequest, ...yarpc.CallOption) error
	CloseShard(context.Context, *types.CloseShardRequest, ...yarpc.CallOption) error
	DescribeCluster(context.Context, ...yarpc.CallOption) (*types.DescribeClusterResponse, error)
	DescribeShardDistribution(context.Context, *types.DescribeShardDistributionRequest, ...yarpc.CallOption) (*types.DescribeShardDistributionResponse, error)
	DescribeHistoryHost(context.Context, *types.DescribeHistoryHostRequest, ...yarpc.CallOption) (*types.DescribeHistoryHostResponse, error)
	DescribeQueue(context.Context, *types.DescribeQueueRequest, ...yarpc.CallOption) (*types.DescribeQueueResponse, error)
	DescribeWorkflowExecution(context.Context, *types.AdminDescribeWorkflowExecutionRequest, ...yarpc.CallOption) (*types.AdminDescribeWorkflowExecutionResponse, error)
	GetDLQReplicationMessages(context.Context, *types.GetDLQReplicationMessagesRequest, ...yarpc.CallOption) (*types.GetDLQReplicationMessagesResponse, error)
	GetDomainReplicationMessages(context.Context, *types.GetDomainReplicationMessagesRequest, ...yarpc.CallOption) (*types.GetDomainReplicationMessagesResponse, error)
	GetReplicationMessages(context.Context, *types.GetReplicationMessagesRequest, ...yarpc.CallOption) (*types.GetReplicationMessagesResponse, error)
	GetWorkflowExecutionRawHistoryV2(context.Context, *types.GetWorkflowExecutionRawHistoryV2Request, ...yarpc.CallOption) (*types.GetWorkflowExecutionRawHistoryV2Response, error)
	CountDLQMessages(context.Context, *types.CountDLQMessagesRequest, ...yarpc.CallOption) (*types.CountDLQMessagesResponse, error)
	MergeDLQMessages(context.Context, *types.MergeDLQMessagesRequest, ...yarpc.CallOption) (*types.MergeDLQMessagesResponse, error)
	PurgeDLQMessages(context.Context, *types.PurgeDLQMessagesRequest, ...yarpc.CallOption) error
	ReadDLQMessages(context.Context, *types.ReadDLQMessagesRequest, ...yarpc.CallOption) (*types.ReadDLQMessagesResponse, error)
	ReapplyEvents(context.Context, *types.ReapplyEventsRequest, ...yarpc.CallOption) error
	RefreshWorkflowTasks(context.Context, *types.RefreshWorkflowTasksRequest, ...yarpc.CallOption) error
	RemoveTask(context.Context, *types.RemoveTaskRequest, ...yarpc.CallOption) error
	ResendReplicationTasks(context.Context, *types.ResendReplicationTasksRequest, ...yarpc.CallOption) error
	ResetQueue(context.Context, *types.ResetQueueRequest, ...yarpc.CallOption) error
	GetCrossClusterTasks(context.Context, *types.GetCrossClusterTasksRequest, ...yarpc.CallOption) (*types.GetCrossClusterTasksResponse, error)
	RespondCrossClusterTasksCompleted(context.Context, *types.RespondCrossClusterTasksCompletedRequest, ...yarpc.CallOption) (*types.RespondCrossClusterTasksCompletedResponse, error)
	GetDynamicConfig(context.Context, *types.GetDynamicConfigRequest, ...yarpc.CallOption) (*types.GetDynamicConfigResponse, error)
	UpdateDynamicConfig(context.Context, *types.UpdateDynamicConfigRequest, ...yarpc.CallOption) error
	RestoreDynamicConfig(context.Context, *types.RestoreDynamicConfigRequest, ...yarpc.CallOption) error
	ListDynamicConfig(context.Context, *types.ListDynamicConfigRequest, ...yarpc.CallOption) (*types.ListDynamicConfigResponse, error)
	DeleteWorkflow(context.Context, *types.AdminDeleteWorkflowRequest, ...yarpc.CallOption) (*types.AdminDeleteWorkflowResponse, error)
	MaintainCorruptWorkflow(context.Context, *types.AdminMaintainWorkflowRequest, ...yarpc.CallOption) (*types.AdminMaintainWorkflowResponse, error)
}
