// Copyright (c) 2021 Uber Technologies, Inc.
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

	adminv1 "github.com/uber/cadence/.gen/proto/admin/v1"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/proto"
)

type grpcClient struct {
	c adminv1.AdminAPIYARPCClient
}

func NewGRPCClient(c adminv1.AdminAPIYARPCClient) Client {
	return grpcClient{c}
}

func (g grpcClient) AddSearchAttribute(ctx context.Context, request *types.AddSearchAttributeRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.AddSearchAttribute(ctx, proto.FromAdminAddSearchAttributeRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) CloseShard(ctx context.Context, request *types.CloseShardRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.CloseShard(ctx, proto.FromAdminCloseShardRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) DescribeCluster(ctx context.Context, opts ...yarpc.CallOption) (*types.DescribeClusterResponse, error) {
	response, err := g.c.DescribeCluster(ctx, &adminv1.DescribeClusterRequest{}, opts...)
	return proto.ToAdminDescribeClusterResponse(response), proto.ToError(err)
}

func (g grpcClient) DescribeShardDistribution(ctx context.Context, request *types.DescribeShardDistributionRequest, opts ...yarpc.CallOption) (*types.DescribeShardDistributionResponse, error) {
	response, err := g.c.DescribeShardDistribution(ctx, proto.FromAdminDescribeShardDistributionRequest(request), opts...)
	return proto.ToAdminDescribeShardDistributionResponse(response), proto.ToError(err)
}

func (g grpcClient) DescribeHistoryHost(ctx context.Context, request *types.DescribeHistoryHostRequest, opts ...yarpc.CallOption) (*types.DescribeHistoryHostResponse, error) {
	response, err := g.c.DescribeHistoryHost(ctx, proto.FromAdminDescribeHistoryHostRequest(request), opts...)
	return proto.ToAdminDescribeHistoryHostResponse(response), proto.ToError(err)
}

func (g grpcClient) DescribeQueue(ctx context.Context, request *types.DescribeQueueRequest, opts ...yarpc.CallOption) (*types.DescribeQueueResponse, error) {
	response, err := g.c.DescribeQueue(ctx, proto.FromAdminDescribeQueueRequest(request), opts...)
	return proto.ToAdminDescribeQueueResponse(response), proto.ToError(err)
}

func (g grpcClient) DescribeWorkflowExecution(ctx context.Context, request *types.AdminDescribeWorkflowExecutionRequest, opts ...yarpc.CallOption) (*types.AdminDescribeWorkflowExecutionResponse, error) {
	response, err := g.c.DescribeWorkflowExecution(ctx, proto.FromAdminDescribeWorkflowExecutionRequest(request), opts...)
	return proto.ToAdminDescribeWorkflowExecutionResponse(response), proto.ToError(err)
}

func (g grpcClient) GetDLQReplicationMessages(ctx context.Context, request *types.GetDLQReplicationMessagesRequest, opts ...yarpc.CallOption) (*types.GetDLQReplicationMessagesResponse, error) {
	response, err := g.c.GetDLQReplicationMessages(ctx, proto.FromAdminGetDLQReplicationMessagesRequest(request), opts...)
	return proto.ToAdminGetDLQReplicationMessagesResponse(response), proto.ToError(err)
}

func (g grpcClient) GetDomainReplicationMessages(ctx context.Context, request *types.GetDomainReplicationMessagesRequest, opts ...yarpc.CallOption) (*types.GetDomainReplicationMessagesResponse, error) {
	response, err := g.c.GetDomainReplicationMessages(ctx, proto.FromAdminGetDomainReplicationMessagesRequest(request), opts...)
	return proto.ToAdminGetDomainReplicationMessagesResponse(response), proto.ToError(err)
}

func (g grpcClient) GetReplicationMessages(ctx context.Context, request *types.GetReplicationMessagesRequest, opts ...yarpc.CallOption) (*types.GetReplicationMessagesResponse, error) {
	response, err := g.c.GetReplicationMessages(ctx, proto.FromAdminGetReplicationMessagesRequest(request), opts...)
	return proto.ToAdminGetReplicationMessagesResponse(response), proto.ToError(err)
}

func (g grpcClient) GetWorkflowExecutionRawHistoryV2(ctx context.Context, request *types.GetWorkflowExecutionRawHistoryV2Request, opts ...yarpc.CallOption) (*types.GetWorkflowExecutionRawHistoryV2Response, error) {
	response, err := g.c.GetWorkflowExecutionRawHistoryV2(ctx, proto.FromAdminGetWorkflowExecutionRawHistoryV2Request(request), opts...)
	return proto.ToAdminGetWorkflowExecutionRawHistoryV2Response(response), proto.ToError(err)
}

func (g grpcClient) CountDLQMessages(ctx context.Context, request *types.CountDLQMessagesRequest, opts ...yarpc.CallOption) (*types.CountDLQMessagesResponse, error) {
	response, err := g.c.CountDLQMessages(ctx, proto.FromAdminCountDLQMessagesRequest(request), opts...)
	return proto.ToAdminCountDLQMessagesResponse(response), proto.ToError(err)
}

func (g grpcClient) MergeDLQMessages(ctx context.Context, request *types.MergeDLQMessagesRequest, opts ...yarpc.CallOption) (*types.MergeDLQMessagesResponse, error) {
	response, err := g.c.MergeDLQMessages(ctx, proto.FromAdminMergeDLQMessagesRequest(request), opts...)
	return proto.ToAdminMergeDLQMessagesResponse(response), proto.ToError(err)
}

func (g grpcClient) PurgeDLQMessages(ctx context.Context, request *types.PurgeDLQMessagesRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.PurgeDLQMessages(ctx, proto.FromAdminPurgeDLQMessagesRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) ReadDLQMessages(ctx context.Context, request *types.ReadDLQMessagesRequest, opts ...yarpc.CallOption) (*types.ReadDLQMessagesResponse, error) {
	response, err := g.c.ReadDLQMessages(ctx, proto.FromAdminReadDLQMessagesRequest(request), opts...)
	return proto.ToAdminReadDLQMessagesResponse(response), proto.ToError(err)
}

func (g grpcClient) ReapplyEvents(ctx context.Context, request *types.ReapplyEventsRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.ReapplyEvents(ctx, proto.FromAdminReapplyEventsRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RefreshWorkflowTasks(ctx context.Context, request *types.RefreshWorkflowTasksRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.RefreshWorkflowTasks(ctx, proto.FromAdminRefreshWorkflowTasksRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RemoveTask(ctx context.Context, request *types.RemoveTaskRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.RemoveTask(ctx, proto.FromAdminRemoveTaskRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) ResendReplicationTasks(ctx context.Context, request *types.ResendReplicationTasksRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.ResendReplicationTasks(ctx, proto.FromAdminResendReplicationTasksRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) ResetQueue(ctx context.Context, request *types.ResetQueueRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.ResetQueue(ctx, proto.FromAdminResetQueueRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) GetCrossClusterTasks(ctx context.Context, request *types.GetCrossClusterTasksRequest, opts ...yarpc.CallOption) (*types.GetCrossClusterTasksResponse, error) {
	response, err := g.c.GetCrossClusterTasks(ctx, proto.FromAdminGetCrossClusterTasksRequest(request), opts...)
	return proto.ToAdminGetCrossClusterTasksResponse(response), proto.ToError(err)
}

func (g grpcClient) RespondCrossClusterTasksCompleted(ctx context.Context, request *types.RespondCrossClusterTasksCompletedRequest, opts ...yarpc.CallOption) (*types.RespondCrossClusterTasksCompletedResponse, error) {
	response, err := g.c.RespondCrossClusterTasksCompleted(ctx, proto.FromAdminRespondCrossClusterTasksCompletedRequest(request), opts...)
	return proto.ToAdminRespondCrossClusterTasksCompletedResponse(response), proto.ToError(err)
}

func (g grpcClient) GetDynamicConfig(ctx context.Context, request *types.GetDynamicConfigRequest, opts ...yarpc.CallOption) (*types.GetDynamicConfigResponse, error) {
	response, err := g.c.GetDynamicConfig(ctx, proto.FromGetDynamicConfigRequest(request), opts...)
	return proto.ToGetDynamicConfigResponse(response), proto.ToError(err)
}

func (g grpcClient) UpdateDynamicConfig(ctx context.Context, request *types.UpdateDynamicConfigRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.UpdateDynamicConfig(ctx, proto.FromUpdateDynamicConfigRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RestoreDynamicConfig(ctx context.Context, request *types.RestoreDynamicConfigRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.RestoreDynamicConfig(ctx, proto.FromRestoreDynamicConfigRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) DeleteWorkflow(ctx context.Context, request *types.AdminDeleteWorkflowRequest, opts ...yarpc.CallOption) (*types.AdminDeleteWorkflowResponse, error) {
	response, err := g.c.DeleteWorkflow(ctx, proto.FromAdminDeleteWorkflowRequest(request), opts...)
	return proto.ToAdminDeleteWorkflowResponse(response), proto.ToError(err)
}

func (g grpcClient) MaintainCorruptWorkflow(ctx context.Context, request *types.AdminMaintainWorkflowRequest, opts ...yarpc.CallOption) (*types.AdminMaintainWorkflowResponse, error) {
	response, err := g.c.MaintainCorruptWorkflow(ctx, proto.FromAdminMaintainWorkflowRequest(request), opts...)
	return proto.ToAdminMaintainWorkflowResponse(response), proto.ToError(err)
}

func (g grpcClient) ListDynamicConfig(ctx context.Context, request *types.ListDynamicConfigRequest, opts ...yarpc.CallOption) (*types.ListDynamicConfigResponse, error) {
	response, err := g.c.ListDynamicConfig(ctx, proto.FromListDynamicConfigRequest(request), opts...)
	return proto.ToListDynamicConfigResponse(response), proto.ToError(err)
}
