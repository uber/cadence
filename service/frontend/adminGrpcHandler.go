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

package frontend

import (
	"context"

	"go.uber.org/yarpc"

	adminv1 "github.com/uber/cadence/.gen/proto/admin/v1"
	"github.com/uber/cadence/common/types/mapper/proto"
)

type adminGRPCHandler struct {
	h AdminHandler
}

func newAdminGRPCHandler(h AdminHandler) adminGRPCHandler {
	return adminGRPCHandler{h}
}

func (g adminGRPCHandler) register(dispatcher *yarpc.Dispatcher) {
	dispatcher.Register(adminv1.BuildAdminAPIYARPCProcedures(g))
}

func (g adminGRPCHandler) AddSearchAttribute(ctx context.Context, request *adminv1.AddSearchAttributeRequest) (*adminv1.AddSearchAttributeResponse, error) {
	err := g.h.AddSearchAttribute(ctx, proto.ToAdminAddSearchAttributeRequest(request))
	return &adminv1.AddSearchAttributeResponse{}, proto.FromError(err)
}

func (g adminGRPCHandler) CloseShard(ctx context.Context, request *adminv1.CloseShardRequest) (*adminv1.CloseShardResponse, error) {
	err := g.h.CloseShard(ctx, proto.ToAdminCloseShardRequest(request))
	return &adminv1.CloseShardResponse{}, proto.FromError(err)
}

func (g adminGRPCHandler) DescribeCluster(ctx context.Context, _ *adminv1.DescribeClusterRequest) (*adminv1.DescribeClusterResponse, error) {
	response, err := g.h.DescribeCluster(ctx)
	return proto.FromAdminDescribeClusterResponse(response), proto.FromError(err)
}

func (g adminGRPCHandler) DescribeShardDistribution(ctx context.Context, request *adminv1.DescribeShardDistributionRequest) (*adminv1.DescribeShardDistributionResponse, error) {
	response, err := g.h.DescribeShardDistribution(ctx, proto.ToAdminDescribeShardDistributionRequest(request))
	return proto.FromAdminDescribeShardDistributionResponse(response), proto.FromError(err)
}

func (g adminGRPCHandler) DescribeHistoryHost(ctx context.Context, request *adminv1.DescribeHistoryHostRequest) (*adminv1.DescribeHistoryHostResponse, error) {
	response, err := g.h.DescribeHistoryHost(ctx, proto.ToAdminDescribeHistoryHostRequest(request))
	return proto.FromAdminDescribeHistoryHostResponse(response), proto.FromError(err)
}

func (g adminGRPCHandler) DescribeQueue(ctx context.Context, request *adminv1.DescribeQueueRequest) (*adminv1.DescribeQueueResponse, error) {
	response, err := g.h.DescribeQueue(ctx, proto.ToAdminDescribeQueueRequest(request))
	return proto.FromAdminDescribeQueueResponse(response), proto.FromError(err)
}

func (g adminGRPCHandler) DescribeWorkflowExecution(ctx context.Context, request *adminv1.DescribeWorkflowExecutionRequest) (*adminv1.DescribeWorkflowExecutionResponse, error) {
	response, err := g.h.DescribeWorkflowExecution(ctx, proto.ToAdminDescribeWorkflowExecutionRequest(request))
	return proto.FromAdminDescribeWorkflowExecutionResponse(response), proto.FromError(err)
}

func (g adminGRPCHandler) GetDLQReplicationMessages(ctx context.Context, request *adminv1.GetDLQReplicationMessagesRequest) (*adminv1.GetDLQReplicationMessagesResponse, error) {
	response, err := g.h.GetDLQReplicationMessages(ctx, proto.ToAdminGetDLQReplicationMessagesRequest(request))
	return proto.FromAdminGetDLQReplicationMessagesResponse(response), proto.FromError(err)
}

func (g adminGRPCHandler) GetDomainReplicationMessages(ctx context.Context, request *adminv1.GetDomainReplicationMessagesRequest) (*adminv1.GetDomainReplicationMessagesResponse, error) {
	response, err := g.h.GetDomainReplicationMessages(ctx, proto.ToAdminGetDomainReplicationMessagesRequest(request))
	return proto.FromAdminGetDomainReplicationMessagesResponse(response), proto.FromError(err)
}

func (g adminGRPCHandler) GetReplicationMessages(ctx context.Context, request *adminv1.GetReplicationMessagesRequest) (*adminv1.GetReplicationMessagesResponse, error) {
	response, err := g.h.GetReplicationMessages(ctx, proto.ToAdminGetReplicationMessagesRequest(request))
	return proto.FromAdminGetReplicationMessagesResponse(response), proto.FromError(err)
}

func (g adminGRPCHandler) GetWorkflowExecutionRawHistoryV2(ctx context.Context, request *adminv1.GetWorkflowExecutionRawHistoryV2Request) (*adminv1.GetWorkflowExecutionRawHistoryV2Response, error) {
	response, err := g.h.GetWorkflowExecutionRawHistoryV2(ctx, proto.ToAdminGetWorkflowExecutionRawHistoryV2Request(request))
	return proto.FromAdminGetWorkflowExecutionRawHistoryV2Response(response), proto.FromError(err)
}

func (g adminGRPCHandler) CountDLQMessages(ctx context.Context, request *adminv1.CountDLQMessagesRequest) (*adminv1.CountDLQMessagesResponse, error) {
	response, err := g.h.CountDLQMessages(ctx, proto.ToAdminCountDLQMessagesRequest(request))
	return proto.FromAdminCountDLQMessagesResponse(response), proto.FromError(err)
}

func (g adminGRPCHandler) MergeDLQMessages(ctx context.Context, request *adminv1.MergeDLQMessagesRequest) (*adminv1.MergeDLQMessagesResponse, error) {
	response, err := g.h.MergeDLQMessages(ctx, proto.ToAdminMergeDLQMessagesRequest(request))
	return proto.FromAdminMergeDLQMessagesResponse(response), proto.FromError(err)
}

func (g adminGRPCHandler) PurgeDLQMessages(ctx context.Context, request *adminv1.PurgeDLQMessagesRequest) (*adminv1.PurgeDLQMessagesResponse, error) {
	err := g.h.PurgeDLQMessages(ctx, proto.ToAdminPurgeDLQMessagesRequest(request))
	return &adminv1.PurgeDLQMessagesResponse{}, proto.FromError(err)
}

func (g adminGRPCHandler) ReadDLQMessages(ctx context.Context, request *adminv1.ReadDLQMessagesRequest) (*adminv1.ReadDLQMessagesResponse, error) {
	response, err := g.h.ReadDLQMessages(ctx, proto.ToAdminReadDLQMessagesRequest(request))
	return proto.FromAdminReadDLQMessagesResponse(response), proto.FromError(err)
}

func (g adminGRPCHandler) ReapplyEvents(ctx context.Context, request *adminv1.ReapplyEventsRequest) (*adminv1.ReapplyEventsResponse, error) {
	err := g.h.ReapplyEvents(ctx, proto.ToAdminReapplyEventsRequest(request))
	return &adminv1.ReapplyEventsResponse{}, proto.FromError(err)
}

func (g adminGRPCHandler) RefreshWorkflowTasks(ctx context.Context, request *adminv1.RefreshWorkflowTasksRequest) (*adminv1.RefreshWorkflowTasksResponse, error) {
	err := g.h.RefreshWorkflowTasks(ctx, proto.ToAdminRefreshWorkflowTasksRequest(request))
	return &adminv1.RefreshWorkflowTasksResponse{}, proto.FromError(err)
}

func (g adminGRPCHandler) RemoveTask(ctx context.Context, request *adminv1.RemoveTaskRequest) (*adminv1.RemoveTaskResponse, error) {
	err := g.h.RemoveTask(ctx, proto.ToAdminRemoveTaskRequest(request))
	return &adminv1.RemoveTaskResponse{}, proto.FromError(err)
}

func (g adminGRPCHandler) ResendReplicationTasks(ctx context.Context, request *adminv1.ResendReplicationTasksRequest) (*adminv1.ResendReplicationTasksResponse, error) {
	err := g.h.ResendReplicationTasks(ctx, proto.ToAdminResendReplicationTasksRequest(request))
	return &adminv1.ResendReplicationTasksResponse{}, proto.FromError(err)
}

func (g adminGRPCHandler) ResetQueue(ctx context.Context, request *adminv1.ResetQueueRequest) (*adminv1.ResetQueueResponse, error) {
	err := g.h.ResetQueue(ctx, proto.ToAdminResetQueueRequest(request))
	return &adminv1.ResetQueueResponse{}, proto.FromError(err)
}

func (g adminGRPCHandler) GetCrossClusterTasks(ctx context.Context, request *adminv1.GetCrossClusterTasksRequest) (*adminv1.GetCrossClusterTasksResponse, error) {
	response, err := g.h.GetCrossClusterTasks(ctx, proto.ToAdminGetCrossClusterTasksRequest(request))
	return proto.FromAdminGetCrossClusterTasksResponse(response), proto.FromError(err)
}

func (g adminGRPCHandler) RespondCrossClusterTasksCompleted(ctx context.Context, request *adminv1.RespondCrossClusterTasksCompletedRequest) (*adminv1.RespondCrossClusterTasksCompletedResponse, error) {
	response, err := g.h.RespondCrossClusterTasksCompleted(ctx, proto.ToAdminRespondCrossClusterTasksCompletedRequest(request))
	return proto.FromAdminRespondCrossClusterTasksCompletedResponse(response), proto.FromError(err)
}

func (g adminGRPCHandler) GetDynamicConfig(ctx context.Context, request *adminv1.GetDynamicConfigRequest) (*adminv1.GetDynamicConfigResponse, error) {
	response, err := g.h.GetDynamicConfig(ctx, proto.ToGetDynamicConfigRequest(request))
	return proto.FromGetDynamicConfigResponse(response), proto.FromError(err)
}

func (g adminGRPCHandler) UpdateDynamicConfig(ctx context.Context, request *adminv1.UpdateDynamicConfigRequest) (*adminv1.UpdateDynamicConfigResponse, error) {
	err := g.h.UpdateDynamicConfig(ctx, proto.ToUpdateDynamicConfigRequest(request))
	return &adminv1.UpdateDynamicConfigResponse{}, proto.FromError(err)
}

func (g adminGRPCHandler) RestoreDynamicConfig(ctx context.Context, request *adminv1.RestoreDynamicConfigRequest) (*adminv1.RestoreDynamicConfigResponse, error) {
	err := g.h.RestoreDynamicConfig(ctx, proto.ToRestoreDynamicConfigRequest(request))
	return &adminv1.RestoreDynamicConfigResponse{}, proto.FromError(err)
}

func (g adminGRPCHandler) DeleteWorkflow(ctx context.Context, request *adminv1.AdminDeleteWorkflowRequest) (*adminv1.AdminDeleteWorkflowResponse, error) {
	response, err := g.h.DeleteWorkflow(ctx, proto.ToAdminDeleteWorkflowRequest(request))
	return proto.FromAdminDeleteWorkflowResponse(response), proto.FromError(err)
}

func (g adminGRPCHandler) MaintainCorruptWorkflow(ctx context.Context, request *adminv1.AdminMaintainWorkflowRequest) (*adminv1.AdminMaintainWorkflowResponse, error) {
	response, err := g.h.MaintainCorruptWorkflow(ctx, proto.ToAdminMaintainWorkflowRequest(request))
	return proto.FromAdminMaintainWorkflowResponse(response), proto.FromError(err)
}

func (g adminGRPCHandler) ListDynamicConfig(ctx context.Context, request *adminv1.ListDynamicConfigRequest) (*adminv1.ListDynamicConfigResponse, error) {
	response, err := g.h.ListDynamicConfig(ctx, proto.ToListDynamicConfigRequest(request))
	return proto.FromListDynamicConfigResponse(response), proto.FromError(err)
}
