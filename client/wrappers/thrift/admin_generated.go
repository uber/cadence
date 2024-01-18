// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package thrift

// Code generated by gowrap. DO NOT EDIT.
// template: ../../templates/thrift.tmpl
// gowrap: http://github.com/hexdigest/gowrap

import (
	"context"

	"go.uber.org/yarpc"

	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"
)

func (g adminClient) AddSearchAttribute(ctx context.Context, ap1 *types.AddSearchAttributeRequest, p1 ...yarpc.CallOption) (err error) {
	err = g.c.AddSearchAttribute(ctx, thrift.FromAdminAddSearchAttributeRequest(ap1), p1...)
	return thrift.ToError(err)
}

func (g adminClient) CloseShard(ctx context.Context, cp1 *types.CloseShardRequest, p1 ...yarpc.CallOption) (err error) {
	err = g.c.CloseShard(ctx, thrift.FromAdminCloseShardRequest(cp1), p1...)
	return thrift.ToError(err)
}

func (g adminClient) CountDLQMessages(ctx context.Context, cp1 *types.CountDLQMessagesRequest, p1 ...yarpc.CallOption) (cp2 *types.CountDLQMessagesResponse, err error) {
	return nil, thrift.ToError(&types.BadRequestError{Message: "Feature not supported on TChannel"})
}

func (g adminClient) DeleteWorkflow(ctx context.Context, ap1 *types.AdminDeleteWorkflowRequest, p1 ...yarpc.CallOption) (ap2 *types.AdminDeleteWorkflowResponse, err error) {
	response, err := g.c.DeleteWorkflow(ctx, thrift.FromAdminDeleteWorkflowRequest(ap1), p1...)
	return thrift.ToAdminDeleteWorkflowResponse(response), thrift.ToError(err)
}

func (g adminClient) DescribeCluster(ctx context.Context, p1 ...yarpc.CallOption) (dp1 *types.DescribeClusterResponse, err error) {
	response, err := g.c.DescribeCluster(ctx, p1...)
	return thrift.ToAdminDescribeClusterResponse(response), thrift.ToError(err)
}

func (g adminClient) DescribeHistoryHost(ctx context.Context, dp1 *types.DescribeHistoryHostRequest, p1 ...yarpc.CallOption) (dp2 *types.DescribeHistoryHostResponse, err error) {
	response, err := g.c.DescribeHistoryHost(ctx, thrift.FromAdminDescribeHistoryHostRequest(dp1), p1...)
	return thrift.ToAdminDescribeHistoryHostResponse(response), thrift.ToError(err)
}

func (g adminClient) DescribeQueue(ctx context.Context, dp1 *types.DescribeQueueRequest, p1 ...yarpc.CallOption) (dp2 *types.DescribeQueueResponse, err error) {
	response, err := g.c.DescribeQueue(ctx, thrift.FromAdminDescribeQueueRequest(dp1), p1...)
	return thrift.ToAdminDescribeQueueResponse(response), thrift.ToError(err)
}

func (g adminClient) DescribeShardDistribution(ctx context.Context, dp1 *types.DescribeShardDistributionRequest, p1 ...yarpc.CallOption) (dp2 *types.DescribeShardDistributionResponse, err error) {
	response, err := g.c.DescribeShardDistribution(ctx, thrift.FromAdminDescribeShardDistributionRequest(dp1), p1...)
	return thrift.ToAdminDescribeShardDistributionResponse(response), thrift.ToError(err)
}

func (g adminClient) DescribeWorkflowExecution(ctx context.Context, ap1 *types.AdminDescribeWorkflowExecutionRequest, p1 ...yarpc.CallOption) (ap2 *types.AdminDescribeWorkflowExecutionResponse, err error) {
	response, err := g.c.DescribeWorkflowExecution(ctx, thrift.FromAdminDescribeWorkflowExecutionRequest(ap1), p1...)
	return thrift.ToAdminDescribeWorkflowExecutionResponse(response), thrift.ToError(err)
}

func (g adminClient) GetCrossClusterTasks(ctx context.Context, gp1 *types.GetCrossClusterTasksRequest, p1 ...yarpc.CallOption) (gp2 *types.GetCrossClusterTasksResponse, err error) {
	response, err := g.c.GetCrossClusterTasks(ctx, thrift.FromAdminGetCrossClusterTasksRequest(gp1), p1...)
	return thrift.ToAdminGetCrossClusterTasksResponse(response), thrift.ToError(err)
}

func (g adminClient) GetDLQReplicationMessages(ctx context.Context, gp1 *types.GetDLQReplicationMessagesRequest, p1 ...yarpc.CallOption) (gp2 *types.GetDLQReplicationMessagesResponse, err error) {
	response, err := g.c.GetDLQReplicationMessages(ctx, thrift.FromAdminGetDLQReplicationMessagesRequest(gp1), p1...)
	return thrift.ToAdminGetDLQReplicationMessagesResponse(response), thrift.ToError(err)
}

func (g adminClient) GetDomainIsolationGroups(ctx context.Context, request *types.GetDomainIsolationGroupsRequest, opts ...yarpc.CallOption) (gp1 *types.GetDomainIsolationGroupsResponse, err error) {
	response, err := g.c.GetDomainIsolationGroups(ctx, thrift.FromAdminGetDomainIsolationGroupsRequest(request), opts...)
	return thrift.ToAdminGetDomainIsolationGroupsResponse(response), thrift.ToError(err)
}

func (g adminClient) GetDomainReplicationMessages(ctx context.Context, gp1 *types.GetDomainReplicationMessagesRequest, p1 ...yarpc.CallOption) (gp2 *types.GetDomainReplicationMessagesResponse, err error) {
	response, err := g.c.GetDomainReplicationMessages(ctx, thrift.FromAdminGetDomainReplicationMessagesRequest(gp1), p1...)
	return thrift.ToAdminGetDomainReplicationMessagesResponse(response), thrift.ToError(err)
}

func (g adminClient) GetDynamicConfig(ctx context.Context, gp1 *types.GetDynamicConfigRequest, p1 ...yarpc.CallOption) (gp2 *types.GetDynamicConfigResponse, err error) {
	response, err := g.c.GetDynamicConfig(ctx, thrift.FromAdminGetDynamicConfigRequest(gp1), p1...)
	return thrift.ToAdminGetDynamicConfigResponse(response), thrift.ToError(err)
}

func (g adminClient) GetGlobalIsolationGroups(ctx context.Context, request *types.GetGlobalIsolationGroupsRequest, opts ...yarpc.CallOption) (gp1 *types.GetGlobalIsolationGroupsResponse, err error) {
	response, err := g.c.GetGlobalIsolationGroups(ctx, thrift.FromAdminGetGlobalIsolationGroupsRequest(request), opts...)
	return thrift.ToAdminGetGlobalIsolationGroupsResponse(response), thrift.ToError(err)
}

func (g adminClient) GetReplicationMessages(ctx context.Context, gp1 *types.GetReplicationMessagesRequest, p1 ...yarpc.CallOption) (gp2 *types.GetReplicationMessagesResponse, err error) {
	response, err := g.c.GetReplicationMessages(ctx, thrift.FromAdminGetReplicationMessagesRequest(gp1), p1...)
	return thrift.ToAdminGetReplicationMessagesResponse(response), thrift.ToError(err)
}

func (g adminClient) GetWorkflowExecutionRawHistoryV2(ctx context.Context, gp1 *types.GetWorkflowExecutionRawHistoryV2Request, p1 ...yarpc.CallOption) (gp2 *types.GetWorkflowExecutionRawHistoryV2Response, err error) {
	response, err := g.c.GetWorkflowExecutionRawHistoryV2(ctx, thrift.FromAdminGetWorkflowExecutionRawHistoryV2Request(gp1), p1...)
	return thrift.ToAdminGetWorkflowExecutionRawHistoryV2Response(response), thrift.ToError(err)
}

func (g adminClient) ListDynamicConfig(ctx context.Context, lp1 *types.ListDynamicConfigRequest, p1 ...yarpc.CallOption) (lp2 *types.ListDynamicConfigResponse, err error) {
	response, err := g.c.ListDynamicConfig(ctx, thrift.FromAdminListDynamicConfigRequest(lp1), p1...)
	return thrift.ToAdminListDynamicConfigResponse(response), thrift.ToError(err)
}

func (g adminClient) MaintainCorruptWorkflow(ctx context.Context, ap1 *types.AdminMaintainWorkflowRequest, p1 ...yarpc.CallOption) (ap2 *types.AdminMaintainWorkflowResponse, err error) {
	response, err := g.c.MaintainCorruptWorkflow(ctx, thrift.FromAdminMaintainCorruptWorkflowRequest(ap1), p1...)
	return thrift.ToAdminMaintainCorruptWorkflowResponse(response), thrift.ToError(err)
}

func (g adminClient) MergeDLQMessages(ctx context.Context, mp1 *types.MergeDLQMessagesRequest, p1 ...yarpc.CallOption) (mp2 *types.MergeDLQMessagesResponse, err error) {
	response, err := g.c.MergeDLQMessages(ctx, thrift.FromAdminMergeDLQMessagesRequest(mp1), p1...)
	return thrift.ToAdminMergeDLQMessagesResponse(response), thrift.ToError(err)
}

func (g adminClient) PurgeDLQMessages(ctx context.Context, pp1 *types.PurgeDLQMessagesRequest, p1 ...yarpc.CallOption) (err error) {
	err = g.c.PurgeDLQMessages(ctx, thrift.FromAdminPurgeDLQMessagesRequest(pp1), p1...)
	return thrift.ToError(err)
}

func (g adminClient) ReadDLQMessages(ctx context.Context, rp1 *types.ReadDLQMessagesRequest, p1 ...yarpc.CallOption) (rp2 *types.ReadDLQMessagesResponse, err error) {
	response, err := g.c.ReadDLQMessages(ctx, thrift.FromAdminReadDLQMessagesRequest(rp1), p1...)
	return thrift.ToAdminReadDLQMessagesResponse(response), thrift.ToError(err)
}

func (g adminClient) ReapplyEvents(ctx context.Context, rp1 *types.ReapplyEventsRequest, p1 ...yarpc.CallOption) (err error) {
	err = g.c.ReapplyEvents(ctx, thrift.FromAdminReapplyEventsRequest(rp1), p1...)
	return thrift.ToError(err)
}

func (g adminClient) RefreshWorkflowTasks(ctx context.Context, rp1 *types.RefreshWorkflowTasksRequest, p1 ...yarpc.CallOption) (err error) {
	err = g.c.RefreshWorkflowTasks(ctx, thrift.FromAdminRefreshWorkflowTasksRequest(rp1), p1...)
	return thrift.ToError(err)
}

func (g adminClient) RemoveTask(ctx context.Context, rp1 *types.RemoveTaskRequest, p1 ...yarpc.CallOption) (err error) {
	err = g.c.RemoveTask(ctx, thrift.FromAdminRemoveTaskRequest(rp1), p1...)
	return thrift.ToError(err)
}

func (g adminClient) ResendReplicationTasks(ctx context.Context, rp1 *types.ResendReplicationTasksRequest, p1 ...yarpc.CallOption) (err error) {
	err = g.c.ResendReplicationTasks(ctx, thrift.FromAdminResendReplicationTasksRequest(rp1), p1...)
	return thrift.ToError(err)
}

func (g adminClient) ResetQueue(ctx context.Context, rp1 *types.ResetQueueRequest, p1 ...yarpc.CallOption) (err error) {
	err = g.c.ResetQueue(ctx, thrift.FromAdminResetQueueRequest(rp1), p1...)
	return thrift.ToError(err)
}

func (g adminClient) RespondCrossClusterTasksCompleted(ctx context.Context, rp1 *types.RespondCrossClusterTasksCompletedRequest, p1 ...yarpc.CallOption) (rp2 *types.RespondCrossClusterTasksCompletedResponse, err error) {
	response, err := g.c.RespondCrossClusterTasksCompleted(ctx, thrift.FromAdminRespondCrossClusterTasksCompletedRequest(rp1), p1...)
	return thrift.ToAdminRespondCrossClusterTasksCompletedResponse(response), thrift.ToError(err)
}

func (g adminClient) RestoreDynamicConfig(ctx context.Context, rp1 *types.RestoreDynamicConfigRequest, p1 ...yarpc.CallOption) (err error) {
	err = g.c.RestoreDynamicConfig(ctx, thrift.FromAdminRestoreDynamicConfigRequest(rp1), p1...)
	return thrift.ToError(err)
}

func (g adminClient) UpdateDomainIsolationGroups(ctx context.Context, request *types.UpdateDomainIsolationGroupsRequest, opts ...yarpc.CallOption) (up1 *types.UpdateDomainIsolationGroupsResponse, err error) {
	response, err := g.c.UpdateDomainIsolationGroups(ctx, thrift.FromAdminUpdateDomainIsolationGroupsRequest(request), opts...)
	return thrift.ToAdminUpdateDomainIsolationGroupsResponse(response), thrift.ToError(err)
}

func (g adminClient) UpdateDynamicConfig(ctx context.Context, up1 *types.UpdateDynamicConfigRequest, p1 ...yarpc.CallOption) (err error) {
	err = g.c.UpdateDynamicConfig(ctx, thrift.FromAdminUpdateDynamicConfigRequest(up1), p1...)
	return thrift.ToError(err)
}

func (g adminClient) UpdateGlobalIsolationGroups(ctx context.Context, request *types.UpdateGlobalIsolationGroupsRequest, opts ...yarpc.CallOption) (up1 *types.UpdateGlobalIsolationGroupsResponse, err error) {
	response, err := g.c.UpdateGlobalIsolationGroups(ctx, thrift.FromAdminUpdateGlobalIsolationGroupsRequest(request), opts...)
	return thrift.ToAdminUpdateGlobalIsolationGroupsResponse(response), thrift.ToError(err)
}
