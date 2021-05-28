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

	apiv1 "github.com/uber/cadence/.gen/proto/api/v1"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/proto"
)

type grpcClient struct {
	domain     apiv1.DomainAPIYARPCClient
	workflow   apiv1.WorkflowAPIYARPCClient
	worker     apiv1.WorkerAPIYARPCClient
	visibility apiv1.VisibilityAPIYARPCClient
}

func NewGRPCClient(
	domain apiv1.DomainAPIYARPCClient,
	workflow apiv1.WorkflowAPIYARPCClient,
	worker apiv1.WorkerAPIYARPCClient,
	visibility apiv1.VisibilityAPIYARPCClient,
) Client {
	return grpcClient{domain, workflow, worker, visibility}
}

func (g grpcClient) CountWorkflowExecutions(ctx context.Context, request *types.CountWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*types.CountWorkflowExecutionsResponse, error) {
	response, err := g.visibility.CountWorkflowExecutions(ctx, proto.FromCountWorkflowExecutionsRequest(request), opts...)
	return proto.ToCountWorkflowExecutionsResponse(response), proto.ToError(err)
}

func (g grpcClient) DeprecateDomain(ctx context.Context, request *types.DeprecateDomainRequest, opts ...yarpc.CallOption) error {
	_, err := g.domain.DeprecateDomain(ctx, proto.FromDeprecateDomainRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) DescribeDomain(ctx context.Context, request *types.DescribeDomainRequest, opts ...yarpc.CallOption) (*types.DescribeDomainResponse, error) {
	response, err := g.domain.DescribeDomain(ctx, proto.FromDescribeDomainRequest(request), opts...)
	return proto.ToDescribeDomainResponse(response), proto.ToError(err)
}

func (g grpcClient) DescribeTaskList(ctx context.Context, request *types.DescribeTaskListRequest, opts ...yarpc.CallOption) (*types.DescribeTaskListResponse, error) {
	response, err := g.workflow.DescribeTaskList(ctx, proto.FromDescribeTaskListRequest(request), opts...)
	return proto.ToDescribeTaskListResponse(response), proto.ToError(err)
}

func (g grpcClient) DescribeWorkflowExecution(ctx context.Context, request *types.DescribeWorkflowExecutionRequest, opts ...yarpc.CallOption) (*types.DescribeWorkflowExecutionResponse, error) {
	response, err := g.workflow.DescribeWorkflowExecution(ctx, proto.FromDescribeWorkflowExecutionRequest(request), opts...)
	return proto.ToDescribeWorkflowExecutionResponse(response), proto.ToError(err)
}

func (g grpcClient) GetClusterInfo(ctx context.Context, opts ...yarpc.CallOption) (*types.ClusterInfo, error) {
	response, err := g.workflow.GetClusterInfo(ctx, &apiv1.GetClusterInfoRequest{}, opts...)
	return proto.ToGetClusterInfoResponse(response), proto.ToError(err)
}

func (g grpcClient) GetSearchAttributes(ctx context.Context, opts ...yarpc.CallOption) (*types.GetSearchAttributesResponse, error) {
	response, err := g.visibility.GetSearchAttributes(ctx, &apiv1.GetSearchAttributesRequest{}, opts...)
	return proto.ToGetSearchAttributesResponse(response), proto.ToError(err)
}

func (g grpcClient) GetWorkflowExecutionHistory(ctx context.Context, request *types.GetWorkflowExecutionHistoryRequest, opts ...yarpc.CallOption) (*types.GetWorkflowExecutionHistoryResponse, error) {
	response, err := g.workflow.GetWorkflowExecutionHistory(ctx, proto.FromGetWorkflowExecutionHistoryRequest(request), opts...)
	return proto.ToGetWorkflowExecutionHistoryResponse(response), proto.ToError(err)
}

func (g grpcClient) ListArchivedWorkflowExecutions(ctx context.Context, request *types.ListArchivedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*types.ListArchivedWorkflowExecutionsResponse, error) {
	response, err := g.visibility.ListArchivedWorkflowExecutions(ctx, proto.FromListArchivedWorkflowExecutionsRequest(request), opts...)
	return proto.ToListArchivedWorkflowExecutionsResponse(response), proto.ToError(err)
}

func (g grpcClient) ListClosedWorkflowExecutions(ctx context.Context, request *types.ListClosedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*types.ListClosedWorkflowExecutionsResponse, error) {
	response, err := g.visibility.ListClosedWorkflowExecutions(ctx, proto.FromListClosedWorkflowExecutionsRequest(request), opts...)
	return proto.ToListClosedWorkflowExecutionsResponse(response), proto.ToError(err)
}

func (g grpcClient) ListDomains(ctx context.Context, request *types.ListDomainsRequest, opts ...yarpc.CallOption) (*types.ListDomainsResponse, error) {
	response, err := g.domain.ListDomains(ctx, proto.FromListDomainsRequest(request), opts...)
	return proto.ToListDomainsResponse(response), proto.ToError(err)
}

func (g grpcClient) ListOpenWorkflowExecutions(ctx context.Context, request *types.ListOpenWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*types.ListOpenWorkflowExecutionsResponse, error) {
	response, err := g.visibility.ListOpenWorkflowExecutions(ctx, proto.FromListOpenWorkflowExecutionsRequest(request), opts...)
	return proto.ToListOpenWorkflowExecutionsResponse(response), proto.ToError(err)
}

func (g grpcClient) ListTaskListPartitions(ctx context.Context, request *types.ListTaskListPartitionsRequest, opts ...yarpc.CallOption) (*types.ListTaskListPartitionsResponse, error) {
	response, err := g.workflow.ListTaskListPartitions(ctx, proto.FromListTaskListPartitionsRequest(request), opts...)
	return proto.ToListTaskListPartitionsResponse(response), proto.ToError(err)
}

func (g grpcClient) ListWorkflowExecutions(ctx context.Context, request *types.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*types.ListWorkflowExecutionsResponse, error) {
	response, err := g.visibility.ListWorkflowExecutions(ctx, proto.FromListWorkflowExecutionsRequest(request), opts...)
	return proto.ToListWorkflowExecutionsResponse(response), proto.ToError(err)
}

func (g grpcClient) PollForActivityTask(ctx context.Context, request *types.PollForActivityTaskRequest, opts ...yarpc.CallOption) (*types.PollForActivityTaskResponse, error) {
	response, err := g.worker.PollForActivityTask(ctx, proto.FromPollForActivityTaskRequest(request), opts...)
	return proto.ToPollForActivityTaskResponse(response), proto.ToError(err)
}

func (g grpcClient) PollForDecisionTask(ctx context.Context, request *types.PollForDecisionTaskRequest, opts ...yarpc.CallOption) (*types.PollForDecisionTaskResponse, error) {
	response, err := g.worker.PollForDecisionTask(ctx, proto.FromPollForDecisionTaskRequest(request), opts...)
	return proto.ToPollForDecisionTaskResponse(response), proto.ToError(err)
}

func (g grpcClient) QueryWorkflow(ctx context.Context, request *types.QueryWorkflowRequest, opts ...yarpc.CallOption) (*types.QueryWorkflowResponse, error) {
	response, err := g.workflow.QueryWorkflow(ctx, proto.FromQueryWorkflowRequest(request), opts...)
	return proto.ToQueryWorkflowResponse(response), proto.ToError(err)
}

func (g grpcClient) RecordActivityTaskHeartbeat(ctx context.Context, request *types.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) (*types.RecordActivityTaskHeartbeatResponse, error) {
	response, err := g.worker.RecordActivityTaskHeartbeat(ctx, proto.FromRecordActivityTaskHeartbeatRequest(request), opts...)
	return proto.ToRecordActivityTaskHeartbeatResponse(response), proto.ToError(err)
}

func (g grpcClient) RecordActivityTaskHeartbeatByID(ctx context.Context, request *types.RecordActivityTaskHeartbeatByIDRequest, opts ...yarpc.CallOption) (*types.RecordActivityTaskHeartbeatResponse, error) {
	response, err := g.worker.RecordActivityTaskHeartbeatByID(ctx, proto.FromRecordActivityTaskHeartbeatByIDRequest(request), opts...)
	return proto.ToRecordActivityTaskHeartbeatByIDResponse(response), proto.ToError(err)
}

func (g grpcClient) RegisterDomain(ctx context.Context, request *types.RegisterDomainRequest, opts ...yarpc.CallOption) error {
	_, err := g.domain.RegisterDomain(ctx, proto.FromRegisterDomainRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RequestCancelWorkflowExecution(ctx context.Context, request *types.RequestCancelWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	_, err := g.workflow.RequestCancelWorkflowExecution(ctx, proto.FromRequestCancelWorkflowExecutionRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) ResetStickyTaskList(ctx context.Context, request *types.ResetStickyTaskListRequest, opts ...yarpc.CallOption) (*types.ResetStickyTaskListResponse, error) {
	_, err := g.worker.ResetStickyTaskList(ctx, proto.FromResetStickyTaskListRequest(request), opts...)
	return &types.ResetStickyTaskListResponse{}, proto.ToError(err)
}

func (g grpcClient) ResetWorkflowExecution(ctx context.Context, request *types.ResetWorkflowExecutionRequest, opts ...yarpc.CallOption) (*types.ResetWorkflowExecutionResponse, error) {
	response, err := g.workflow.ResetWorkflowExecution(ctx, proto.FromResetWorkflowExecutionRequest(request), opts...)
	return proto.ToResetWorkflowExecutionResponse(response), proto.ToError(err)
}

func (g grpcClient) RespondActivityTaskCanceled(ctx context.Context, request *types.RespondActivityTaskCanceledRequest, opts ...yarpc.CallOption) error {
	_, err := g.worker.RespondActivityTaskCanceled(ctx, proto.FromRespondActivityTaskCanceledRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RespondActivityTaskCanceledByID(ctx context.Context, request *types.RespondActivityTaskCanceledByIDRequest, opts ...yarpc.CallOption) error {
	_, err := g.worker.RespondActivityTaskCanceledByID(ctx, proto.FromRespondActivityTaskCanceledByIDRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RespondActivityTaskCompleted(ctx context.Context, request *types.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption) error {
	_, err := g.worker.RespondActivityTaskCompleted(ctx, proto.FromRespondActivityTaskCompletedRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RespondActivityTaskCompletedByID(ctx context.Context, request *types.RespondActivityTaskCompletedByIDRequest, opts ...yarpc.CallOption) error {
	_, err := g.worker.RespondActivityTaskCompletedByID(ctx, proto.FromRespondActivityTaskCompletedByIDRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RespondActivityTaskFailed(ctx context.Context, request *types.RespondActivityTaskFailedRequest, opts ...yarpc.CallOption) error {
	_, err := g.worker.RespondActivityTaskFailed(ctx, proto.FromRespondActivityTaskFailedRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RespondActivityTaskFailedByID(ctx context.Context, request *types.RespondActivityTaskFailedByIDRequest, opts ...yarpc.CallOption) error {
	_, err := g.worker.RespondActivityTaskFailedByID(ctx, proto.FromRespondActivityTaskFailedByIDRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RespondDecisionTaskCompleted(ctx context.Context, request *types.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption) (*types.RespondDecisionTaskCompletedResponse, error) {
	response, err := g.worker.RespondDecisionTaskCompleted(ctx, proto.FromRespondDecisionTaskCompletedRequest(request), opts...)
	return proto.ToRespondDecisionTaskCompletedResponse(response), proto.ToError(err)
}

func (g grpcClient) RespondDecisionTaskFailed(ctx context.Context, request *types.RespondDecisionTaskFailedRequest, opts ...yarpc.CallOption) error {
	_, err := g.worker.RespondDecisionTaskFailed(ctx, proto.FromRespondDecisionTaskFailedRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) RespondQueryTaskCompleted(ctx context.Context, request *types.RespondQueryTaskCompletedRequest, opts ...yarpc.CallOption) error {
	_, err := g.worker.RespondQueryTaskCompleted(ctx, proto.FromRespondQueryTaskCompletedRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) ScanWorkflowExecutions(ctx context.Context, request *types.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*types.ListWorkflowExecutionsResponse, error) {
	response, err := g.visibility.ScanWorkflowExecutions(ctx, proto.FromScanWorkflowExecutionsRequest(request), opts...)
	return proto.ToScanWorkflowExecutionsResponse(response), proto.ToError(err)
}

func (g grpcClient) SignalWithStartWorkflowExecution(ctx context.Context, request *types.SignalWithStartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*types.StartWorkflowExecutionResponse, error) {
	response, err := g.workflow.SignalWithStartWorkflowExecution(ctx, proto.FromSignalWithStartWorkflowExecutionRequest(request), opts...)
	return proto.ToSignalWithStartWorkflowExecutionResponse(response), proto.ToError(err)
}

func (g grpcClient) SignalWorkflowExecution(ctx context.Context, request *types.SignalWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	_, err := g.workflow.SignalWorkflowExecution(ctx, proto.FromSignalWorkflowExecutionRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) StartWorkflowExecution(ctx context.Context, request *types.StartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*types.StartWorkflowExecutionResponse, error) {
	response, err := g.workflow.StartWorkflowExecution(ctx, proto.FromStartWorkflowExecutionRequest(request), opts...)
	return proto.ToStartWorkflowExecutionResponse(response), proto.ToError(err)
}

func (g grpcClient) TerminateWorkflowExecution(ctx context.Context, request *types.TerminateWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	_, err := g.workflow.TerminateWorkflowExecution(ctx, proto.FromTerminateWorkflowExecutionRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) UpdateDomain(ctx context.Context, request *types.UpdateDomainRequest, opts ...yarpc.CallOption) (*types.UpdateDomainResponse, error) {
	response, err := g.domain.UpdateDomain(ctx, proto.FromUpdateDomainRequest(request), opts...)
	return proto.ToUpdateDomainResponse(response), proto.ToError(err)
}
