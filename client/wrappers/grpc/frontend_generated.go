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

// Code generated by gowrap. DO NOT EDIT.
// template: ../../templates/grpc.tmpl
// gowrap: http://github.com/hexdigest/gowrap

package grpc

import (
	"context"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	"go.uber.org/yarpc"

	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/proto"
)

func (g frontendClient) CountWorkflowExecutions(ctx context.Context, cp1 *types.CountWorkflowExecutionsRequest, p1 ...yarpc.CallOption) (cp2 *types.CountWorkflowExecutionsResponse, err error) {
	response, err := g.c.CountWorkflowExecutions(ctx, proto.FromCountWorkflowExecutionsRequest(cp1), p1...)
	return proto.ToCountWorkflowExecutionsResponse(response), proto.ToError(err)
}

func (g frontendClient) DeprecateDomain(ctx context.Context, dp1 *types.DeprecateDomainRequest, p1 ...yarpc.CallOption) (err error) {
	_, err = g.c.DeprecateDomain(ctx, proto.FromDeprecateDomainRequest(dp1), p1...)
	return proto.ToError(err)
}

func (g frontendClient) DescribeDomain(ctx context.Context, dp1 *types.DescribeDomainRequest, p1 ...yarpc.CallOption) (dp2 *types.DescribeDomainResponse, err error) {
	response, err := g.c.DescribeDomain(ctx, proto.FromDescribeDomainRequest(dp1), p1...)
	return proto.ToDescribeDomainResponse(response), proto.ToError(err)
}

func (g frontendClient) DescribeTaskList(ctx context.Context, dp1 *types.DescribeTaskListRequest, p1 ...yarpc.CallOption) (dp2 *types.DescribeTaskListResponse, err error) {
	response, err := g.c.DescribeTaskList(ctx, proto.FromDescribeTaskListRequest(dp1), p1...)
	return proto.ToDescribeTaskListResponse(response), proto.ToError(err)
}

func (g frontendClient) DescribeWorkflowExecution(ctx context.Context, dp1 *types.DescribeWorkflowExecutionRequest, p1 ...yarpc.CallOption) (dp2 *types.DescribeWorkflowExecutionResponse, err error) {
	response, err := g.c.DescribeWorkflowExecution(ctx, proto.FromDescribeWorkflowExecutionRequest(dp1), p1...)
	return proto.ToDescribeWorkflowExecutionResponse(response), proto.ToError(err)
}

func (g frontendClient) GetClusterInfo(ctx context.Context, p1 ...yarpc.CallOption) (cp1 *types.ClusterInfo, err error) {
	response, err := g.c.GetClusterInfo(ctx, &apiv1.GetClusterInfoRequest{}, p1...)
	return proto.ToGetClusterInfoResponse(response), proto.ToError(err)
}

func (g frontendClient) GetSearchAttributes(ctx context.Context, p1 ...yarpc.CallOption) (gp1 *types.GetSearchAttributesResponse, err error) {
	response, err := g.c.GetSearchAttributes(ctx, &apiv1.GetSearchAttributesRequest{}, p1...)
	return proto.ToGetSearchAttributesResponse(response), proto.ToError(err)
}

func (g frontendClient) GetTaskListsByDomain(ctx context.Context, gp1 *types.GetTaskListsByDomainRequest, p1 ...yarpc.CallOption) (gp2 *types.GetTaskListsByDomainResponse, err error) {
	response, err := g.c.GetTaskListsByDomain(ctx, proto.FromGetTaskListsByDomainRequest(gp1), p1...)
	return proto.ToGetTaskListsByDomainResponse(response), proto.ToError(err)
}

func (g frontendClient) GetWorkflowExecutionHistory(ctx context.Context, gp1 *types.GetWorkflowExecutionHistoryRequest, p1 ...yarpc.CallOption) (gp2 *types.GetWorkflowExecutionHistoryResponse, err error) {
	response, err := g.c.GetWorkflowExecutionHistory(ctx, proto.FromGetWorkflowExecutionHistoryRequest(gp1), p1...)
	return proto.ToGetWorkflowExecutionHistoryResponse(response), proto.ToError(err)
}

func (g frontendClient) ListAllWorkflowExecutions(ctx context.Context, lp1 *types.ListAllWorkflowExecutionsRequest) (lp2 *types.ListAllWorkflowExecutionsResponse, err error) {
	response, err := g.c.ListAllWorkflowExecutions(ctx, &apiv1.ListAllWorkflowExecutionsRequest{}, lp1)
	return proto.ToListAllWorkflowExecutionsResponse(response), proto.ToError(err)
}

func (g frontendClient) ListArchivedWorkflowExecutions(ctx context.Context, lp1 *types.ListArchivedWorkflowExecutionsRequest, p1 ...yarpc.CallOption) (lp2 *types.ListArchivedWorkflowExecutionsResponse, err error) {
	response, err := g.c.ListArchivedWorkflowExecutions(ctx, proto.FromListArchivedWorkflowExecutionsRequest(lp1), p1...)
	return proto.ToListArchivedWorkflowExecutionsResponse(response), proto.ToError(err)
}

func (g frontendClient) ListClosedWorkflowExecutions(ctx context.Context, lp1 *types.ListClosedWorkflowExecutionsRequest, p1 ...yarpc.CallOption) (lp2 *types.ListClosedWorkflowExecutionsResponse, err error) {
	response, err := g.c.ListClosedWorkflowExecutions(ctx, proto.FromListClosedWorkflowExecutionsRequest(lp1), p1...)
	return proto.ToListClosedWorkflowExecutionsResponse(response), proto.ToError(err)
}

func (g frontendClient) ListDomains(ctx context.Context, lp1 *types.ListDomainsRequest, p1 ...yarpc.CallOption) (lp2 *types.ListDomainsResponse, err error) {
	response, err := g.c.ListDomains(ctx, proto.FromListDomainsRequest(lp1), p1...)
	return proto.ToListDomainsResponse(response), proto.ToError(err)
}

func (g frontendClient) ListOpenWorkflowExecutions(ctx context.Context, lp1 *types.ListOpenWorkflowExecutionsRequest, p1 ...yarpc.CallOption) (lp2 *types.ListOpenWorkflowExecutionsResponse, err error) {
	response, err := g.c.ListOpenWorkflowExecutions(ctx, proto.FromListOpenWorkflowExecutionsRequest(lp1), p1...)
	return proto.ToListOpenWorkflowExecutionsResponse(response), proto.ToError(err)
}

func (g frontendClient) ListTaskListPartitions(ctx context.Context, lp1 *types.ListTaskListPartitionsRequest, p1 ...yarpc.CallOption) (lp2 *types.ListTaskListPartitionsResponse, err error) {
	response, err := g.c.ListTaskListPartitions(ctx, proto.FromListTaskListPartitionsRequest(lp1), p1...)
	return proto.ToListTaskListPartitionsResponse(response), proto.ToError(err)
}

func (g frontendClient) ListWorkflowExecutions(ctx context.Context, lp1 *types.ListWorkflowExecutionsRequest, p1 ...yarpc.CallOption) (lp2 *types.ListWorkflowExecutionsResponse, err error) {
	response, err := g.c.ListWorkflowExecutions(ctx, proto.FromListWorkflowExecutionsRequest(lp1), p1...)
	return proto.ToListWorkflowExecutionsResponse(response), proto.ToError(err)
}

func (g frontendClient) PollForActivityTask(ctx context.Context, pp1 *types.PollForActivityTaskRequest, p1 ...yarpc.CallOption) (pp2 *types.PollForActivityTaskResponse, err error) {
	response, err := g.c.PollForActivityTask(ctx, proto.FromPollForActivityTaskRequest(pp1), p1...)
	return proto.ToPollForActivityTaskResponse(response), proto.ToError(err)
}

func (g frontendClient) PollForDecisionTask(ctx context.Context, pp1 *types.PollForDecisionTaskRequest, p1 ...yarpc.CallOption) (pp2 *types.PollForDecisionTaskResponse, err error) {
	response, err := g.c.PollForDecisionTask(ctx, proto.FromPollForDecisionTaskRequest(pp1), p1...)
	return proto.ToPollForDecisionTaskResponse(response), proto.ToError(err)
}

func (g frontendClient) QueryWorkflow(ctx context.Context, qp1 *types.QueryWorkflowRequest, p1 ...yarpc.CallOption) (qp2 *types.QueryWorkflowResponse, err error) {
	response, err := g.c.QueryWorkflow(ctx, proto.FromQueryWorkflowRequest(qp1), p1...)
	return proto.ToQueryWorkflowResponse(response), proto.ToError(err)
}

func (g frontendClient) RecordActivityTaskHeartbeat(ctx context.Context, rp1 *types.RecordActivityTaskHeartbeatRequest, p1 ...yarpc.CallOption) (rp2 *types.RecordActivityTaskHeartbeatResponse, err error) {
	response, err := g.c.RecordActivityTaskHeartbeat(ctx, proto.FromRecordActivityTaskHeartbeatRequest(rp1), p1...)
	return proto.ToRecordActivityTaskHeartbeatResponse(response), proto.ToError(err)
}

func (g frontendClient) RecordActivityTaskHeartbeatByID(ctx context.Context, rp1 *types.RecordActivityTaskHeartbeatByIDRequest, p1 ...yarpc.CallOption) (rp2 *types.RecordActivityTaskHeartbeatResponse, err error) {
	response, err := g.c.RecordActivityTaskHeartbeatByID(ctx, proto.FromRecordActivityTaskHeartbeatByIDRequest(rp1), p1...)
	return proto.ToRecordActivityTaskHeartbeatByIDResponse(response), proto.ToError(err)
}

func (g frontendClient) RefreshWorkflowTasks(ctx context.Context, rp1 *types.RefreshWorkflowTasksRequest, p1 ...yarpc.CallOption) (err error) {
	_, err = g.c.RefreshWorkflowTasks(ctx, proto.FromRefreshWorkflowTasksRequest(rp1), p1...)
	return proto.ToError(err)
}

func (g frontendClient) RegisterDomain(ctx context.Context, rp1 *types.RegisterDomainRequest, p1 ...yarpc.CallOption) (err error) {
	_, err = g.c.RegisterDomain(ctx, proto.FromRegisterDomainRequest(rp1), p1...)
	return proto.ToError(err)
}

func (g frontendClient) RequestCancelWorkflowExecution(ctx context.Context, rp1 *types.RequestCancelWorkflowExecutionRequest, p1 ...yarpc.CallOption) (err error) {
	_, err = g.c.RequestCancelWorkflowExecution(ctx, proto.FromRequestCancelWorkflowExecutionRequest(rp1), p1...)
	return proto.ToError(err)
}

func (g frontendClient) ResetStickyTaskList(ctx context.Context, rp1 *types.ResetStickyTaskListRequest, p1 ...yarpc.CallOption) (rp2 *types.ResetStickyTaskListResponse, err error) {
	response, err := g.c.ResetStickyTaskList(ctx, proto.FromResetStickyTaskListRequest(rp1), p1...)
	return proto.ToResetStickyTaskListResponse(response), proto.ToError(err)
}

func (g frontendClient) ResetWorkflowExecution(ctx context.Context, rp1 *types.ResetWorkflowExecutionRequest, p1 ...yarpc.CallOption) (rp2 *types.ResetWorkflowExecutionResponse, err error) {
	response, err := g.c.ResetWorkflowExecution(ctx, proto.FromResetWorkflowExecutionRequest(rp1), p1...)
	return proto.ToResetWorkflowExecutionResponse(response), proto.ToError(err)
}

func (g frontendClient) RespondActivityTaskCanceled(ctx context.Context, rp1 *types.RespondActivityTaskCanceledRequest, p1 ...yarpc.CallOption) (err error) {
	_, err = g.c.RespondActivityTaskCanceled(ctx, proto.FromRespondActivityTaskCanceledRequest(rp1), p1...)
	return proto.ToError(err)
}

func (g frontendClient) RespondActivityTaskCanceledByID(ctx context.Context, rp1 *types.RespondActivityTaskCanceledByIDRequest, p1 ...yarpc.CallOption) (err error) {
	_, err = g.c.RespondActivityTaskCanceledByID(ctx, proto.FromRespondActivityTaskCanceledByIDRequest(rp1), p1...)
	return proto.ToError(err)
}

func (g frontendClient) RespondActivityTaskCompleted(ctx context.Context, rp1 *types.RespondActivityTaskCompletedRequest, p1 ...yarpc.CallOption) (err error) {
	_, err = g.c.RespondActivityTaskCompleted(ctx, proto.FromRespondActivityTaskCompletedRequest(rp1), p1...)
	return proto.ToError(err)
}

func (g frontendClient) RespondActivityTaskCompletedByID(ctx context.Context, rp1 *types.RespondActivityTaskCompletedByIDRequest, p1 ...yarpc.CallOption) (err error) {
	_, err = g.c.RespondActivityTaskCompletedByID(ctx, proto.FromRespondActivityTaskCompletedByIDRequest(rp1), p1...)
	return proto.ToError(err)
}

func (g frontendClient) RespondActivityTaskFailed(ctx context.Context, rp1 *types.RespondActivityTaskFailedRequest, p1 ...yarpc.CallOption) (err error) {
	_, err = g.c.RespondActivityTaskFailed(ctx, proto.FromRespondActivityTaskFailedRequest(rp1), p1...)
	return proto.ToError(err)
}

func (g frontendClient) RespondActivityTaskFailedByID(ctx context.Context, rp1 *types.RespondActivityTaskFailedByIDRequest, p1 ...yarpc.CallOption) (err error) {
	_, err = g.c.RespondActivityTaskFailedByID(ctx, proto.FromRespondActivityTaskFailedByIDRequest(rp1), p1...)
	return proto.ToError(err)
}

func (g frontendClient) RespondDecisionTaskCompleted(ctx context.Context, rp1 *types.RespondDecisionTaskCompletedRequest, p1 ...yarpc.CallOption) (rp2 *types.RespondDecisionTaskCompletedResponse, err error) {
	response, err := g.c.RespondDecisionTaskCompleted(ctx, proto.FromRespondDecisionTaskCompletedRequest(rp1), p1...)
	return proto.ToRespondDecisionTaskCompletedResponse(response), proto.ToError(err)
}

func (g frontendClient) RespondDecisionTaskFailed(ctx context.Context, rp1 *types.RespondDecisionTaskFailedRequest, p1 ...yarpc.CallOption) (err error) {
	_, err = g.c.RespondDecisionTaskFailed(ctx, proto.FromRespondDecisionTaskFailedRequest(rp1), p1...)
	return proto.ToError(err)
}

func (g frontendClient) RespondQueryTaskCompleted(ctx context.Context, rp1 *types.RespondQueryTaskCompletedRequest, p1 ...yarpc.CallOption) (err error) {
	_, err = g.c.RespondQueryTaskCompleted(ctx, proto.FromRespondQueryTaskCompletedRequest(rp1), p1...)
	return proto.ToError(err)
}

func (g frontendClient) RestartWorkflowExecution(ctx context.Context, rp1 *types.RestartWorkflowExecutionRequest, p1 ...yarpc.CallOption) (rp2 *types.RestartWorkflowExecutionResponse, err error) {
	response, err := g.c.RestartWorkflowExecution(ctx, proto.FromRestartWorkflowExecutionRequest(rp1), p1...)
	return proto.ToRestartWorkflowExecutionResponse(response), proto.ToError(err)
}

func (g frontendClient) ScanWorkflowExecutions(ctx context.Context, lp1 *types.ListWorkflowExecutionsRequest, p1 ...yarpc.CallOption) (lp2 *types.ListWorkflowExecutionsResponse, err error) {
	response, err := g.c.ScanWorkflowExecutions(ctx, proto.FromScanWorkflowExecutionsRequest(lp1), p1...)
	return proto.ToScanWorkflowExecutionsResponse(response), proto.ToError(err)
}

func (g frontendClient) SignalWithStartWorkflowExecution(ctx context.Context, sp1 *types.SignalWithStartWorkflowExecutionRequest, p1 ...yarpc.CallOption) (sp2 *types.StartWorkflowExecutionResponse, err error) {
	response, err := g.c.SignalWithStartWorkflowExecution(ctx, proto.FromSignalWithStartWorkflowExecutionRequest(sp1), p1...)
	return proto.ToSignalWithStartWorkflowExecutionResponse(response), proto.ToError(err)
}

func (g frontendClient) SignalWithStartWorkflowExecutionAsync(ctx context.Context, sp1 *types.SignalWithStartWorkflowExecutionAsyncRequest, p1 ...yarpc.CallOption) (sp2 *types.SignalWithStartWorkflowExecutionAsyncResponse, err error) {
	response, err := g.c.SignalWithStartWorkflowExecutionAsync(ctx, proto.FromSignalWithStartWorkflowExecutionAsyncRequest(sp1), p1...)
	return proto.ToSignalWithStartWorkflowExecutionAsyncResponse(response), proto.ToError(err)
}

func (g frontendClient) SignalWorkflowExecution(ctx context.Context, sp1 *types.SignalWorkflowExecutionRequest, p1 ...yarpc.CallOption) (err error) {
	_, err = g.c.SignalWorkflowExecution(ctx, proto.FromSignalWorkflowExecutionRequest(sp1), p1...)
	return proto.ToError(err)
}

func (g frontendClient) StartWorkflowExecution(ctx context.Context, sp1 *types.StartWorkflowExecutionRequest, p1 ...yarpc.CallOption) (sp2 *types.StartWorkflowExecutionResponse, err error) {
	response, err := g.c.StartWorkflowExecution(ctx, proto.FromStartWorkflowExecutionRequest(sp1), p1...)
	return proto.ToStartWorkflowExecutionResponse(response), proto.ToError(err)
}

func (g frontendClient) StartWorkflowExecutionAsync(ctx context.Context, sp1 *types.StartWorkflowExecutionAsyncRequest, p1 ...yarpc.CallOption) (sp2 *types.StartWorkflowExecutionAsyncResponse, err error) {
	response, err := g.c.StartWorkflowExecutionAsync(ctx, proto.FromStartWorkflowExecutionAsyncRequest(sp1), p1...)
	return proto.ToStartWorkflowExecutionAsyncResponse(response), proto.ToError(err)
}

func (g frontendClient) TerminateWorkflowExecution(ctx context.Context, tp1 *types.TerminateWorkflowExecutionRequest, p1 ...yarpc.CallOption) (err error) {
	_, err = g.c.TerminateWorkflowExecution(ctx, proto.FromTerminateWorkflowExecutionRequest(tp1), p1...)
	return proto.ToError(err)
}

func (g frontendClient) UpdateDomain(ctx context.Context, up1 *types.UpdateDomainRequest, p1 ...yarpc.CallOption) (up2 *types.UpdateDomainResponse, err error) {
	response, err := g.c.UpdateDomain(ctx, proto.FromUpdateDomainRequest(up1), p1...)
	return proto.ToUpdateDomainResponse(response), proto.ToError(err)
}
