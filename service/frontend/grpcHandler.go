// Copyright (c) 2021 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package frontend

import (
	"context"

	"go.uber.org/yarpc"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	"github.com/uber/cadence/common/types/mapper/proto"
)

type grpcHandler struct {
	h Handler
}

func newGrpcHandler(h Handler) grpcHandler {
	return grpcHandler{h}
}

func (g grpcHandler) register(dispatcher *yarpc.Dispatcher) {
	dispatcher.Register(apiv1.BuildDomainAPIYARPCProcedures(g))
	dispatcher.Register(apiv1.BuildWorkflowAPIYARPCProcedures(g))
	dispatcher.Register(apiv1.BuildWorkerAPIYARPCProcedures(g))
	dispatcher.Register(apiv1.BuildVisibilityAPIYARPCProcedures(g))
	dispatcher.Register(apiv1.BuildMetaAPIYARPCProcedures(g))
}

func (g grpcHandler) Health(ctx context.Context, _ *apiv1.HealthRequest) (*apiv1.HealthResponse, error) {
	response, err := g.h.Health(ctx)
	return proto.FromHealthResponse(response), proto.FromError(err)
}

func (g grpcHandler) CountWorkflowExecutions(ctx context.Context, request *apiv1.CountWorkflowExecutionsRequest) (*apiv1.CountWorkflowExecutionsResponse, error) {
	response, err := g.h.CountWorkflowExecutions(ctx, proto.ToCountWorkflowExecutionsRequest(request))
	return proto.FromCountWorkflowExecutionsResponse(response), proto.FromError(err)
}

func (g grpcHandler) DeprecateDomain(ctx context.Context, request *apiv1.DeprecateDomainRequest) (*apiv1.DeprecateDomainResponse, error) {
	err := g.h.DeprecateDomain(ctx, proto.ToDeprecateDomainRequest(request))
	return &apiv1.DeprecateDomainResponse{}, proto.FromError(err)
}

func (g grpcHandler) DescribeDomain(ctx context.Context, request *apiv1.DescribeDomainRequest) (*apiv1.DescribeDomainResponse, error) {
	response, err := g.h.DescribeDomain(ctx, proto.ToDescribeDomainRequest(request))
	return proto.FromDescribeDomainResponse(response), proto.FromError(err)
}

func (g grpcHandler) DescribeTaskList(ctx context.Context, request *apiv1.DescribeTaskListRequest) (*apiv1.DescribeTaskListResponse, error) {
	response, err := g.h.DescribeTaskList(ctx, proto.ToDescribeTaskListRequest(request))
	return proto.FromDescribeTaskListResponse(response), proto.FromError(err)
}

func (g grpcHandler) DescribeWorkflowExecution(ctx context.Context, request *apiv1.DescribeWorkflowExecutionRequest) (*apiv1.DescribeWorkflowExecutionResponse, error) {
	response, err := g.h.DescribeWorkflowExecution(ctx, proto.ToDescribeWorkflowExecutionRequest(request))
	return proto.FromDescribeWorkflowExecutionResponse(response), proto.FromError(err)
}

func (g grpcHandler) GetClusterInfo(ctx context.Context, _ *apiv1.GetClusterInfoRequest) (*apiv1.GetClusterInfoResponse, error) {
	response, err := g.h.GetClusterInfo(ctx)
	return proto.FromGetClusterInfoResponse(response), proto.FromError(err)
}

func (g grpcHandler) GetSearchAttributes(ctx context.Context, _ *apiv1.GetSearchAttributesRequest) (*apiv1.GetSearchAttributesResponse, error) {
	response, err := g.h.GetSearchAttributes(ctx)
	return proto.FromGetSearchAttributesResponse(response), proto.FromError(err)
}

func (g grpcHandler) GetWorkflowExecutionHistory(ctx context.Context, request *apiv1.GetWorkflowExecutionHistoryRequest) (*apiv1.GetWorkflowExecutionHistoryResponse, error) {
	response, err := g.h.GetWorkflowExecutionHistory(ctx, proto.ToGetWorkflowExecutionHistoryRequest(request))
	return proto.FromGetWorkflowExecutionHistoryResponse(response), proto.FromError(err)
}

func (g grpcHandler) ListArchivedWorkflowExecutions(ctx context.Context, request *apiv1.ListArchivedWorkflowExecutionsRequest) (*apiv1.ListArchivedWorkflowExecutionsResponse, error) {
	response, err := g.h.ListArchivedWorkflowExecutions(ctx, proto.ToListArchivedWorkflowExecutionsRequest(request))
	return proto.FromListArchivedWorkflowExecutionsResponse(response), proto.FromError(err)
}

func (g grpcHandler) ListClosedWorkflowExecutions(ctx context.Context, request *apiv1.ListClosedWorkflowExecutionsRequest) (*apiv1.ListClosedWorkflowExecutionsResponse, error) {
	response, err := g.h.ListClosedWorkflowExecutions(ctx, proto.ToListClosedWorkflowExecutionsRequest(request))
	return proto.FromListClosedWorkflowExecutionsResponse(response), proto.FromError(err)
}

func (g grpcHandler) ListDomains(ctx context.Context, request *apiv1.ListDomainsRequest) (*apiv1.ListDomainsResponse, error) {
	response, err := g.h.ListDomains(ctx, proto.ToListDomainsRequest(request))
	return proto.FromListDomainsResponse(response), proto.FromError(err)
}

func (g grpcHandler) ListOpenWorkflowExecutions(ctx context.Context, request *apiv1.ListOpenWorkflowExecutionsRequest) (*apiv1.ListOpenWorkflowExecutionsResponse, error) {
	response, err := g.h.ListOpenWorkflowExecutions(ctx, proto.ToListOpenWorkflowExecutionsRequest(request))
	return proto.FromListOpenWorkflowExecutionsResponse(response), proto.FromError(err)
}

func (g grpcHandler) ListTaskListPartitions(ctx context.Context, request *apiv1.ListTaskListPartitionsRequest) (*apiv1.ListTaskListPartitionsResponse, error) {
	response, err := g.h.ListTaskListPartitions(ctx, proto.ToListTaskListPartitionsRequest(request))
	return proto.FromListTaskListPartitionsResponse(response), proto.FromError(err)
}

func (g grpcHandler) GetTaskListsByDomain(ctx context.Context, request *apiv1.GetTaskListsByDomainRequest) (*apiv1.GetTaskListsByDomainResponse, error) {
	response, err := g.h.GetTaskListsByDomain(ctx, proto.ToGetTaskListsByDomainRequest(request))
	return proto.FromGetTaskListsByDomainResponse(response), proto.FromError(err)
}

func (g grpcHandler) ListWorkflowExecutions(ctx context.Context, request *apiv1.ListWorkflowExecutionsRequest) (*apiv1.ListWorkflowExecutionsResponse, error) {
	response, err := g.h.ListWorkflowExecutions(ctx, proto.ToListWorkflowExecutionsRequest(request))
	return proto.FromListWorkflowExecutionsResponse(response), proto.FromError(err)
}

func (g grpcHandler) PollForActivityTask(ctx context.Context, request *apiv1.PollForActivityTaskRequest) (*apiv1.PollForActivityTaskResponse, error) {
	response, err := g.h.PollForActivityTask(ctx, proto.ToPollForActivityTaskRequest(request))
	return proto.FromPollForActivityTaskResponse(response), proto.FromError(err)
}

func (g grpcHandler) PollForDecisionTask(ctx context.Context, request *apiv1.PollForDecisionTaskRequest) (*apiv1.PollForDecisionTaskResponse, error) {
	response, err := g.h.PollForDecisionTask(ctx, proto.ToPollForDecisionTaskRequest(request))
	return proto.FromPollForDecisionTaskResponse(response), proto.FromError(err)
}

func (g grpcHandler) QueryWorkflow(ctx context.Context, request *apiv1.QueryWorkflowRequest) (*apiv1.QueryWorkflowResponse, error) {
	response, err := g.h.QueryWorkflow(ctx, proto.ToQueryWorkflowRequest(request))
	return proto.FromQueryWorkflowResponse(response), proto.FromError(err)
}

func (g grpcHandler) RecordActivityTaskHeartbeat(ctx context.Context, request *apiv1.RecordActivityTaskHeartbeatRequest) (*apiv1.RecordActivityTaskHeartbeatResponse, error) {
	response, err := g.h.RecordActivityTaskHeartbeat(ctx, proto.ToRecordActivityTaskHeartbeatRequest(request))
	return proto.FromRecordActivityTaskHeartbeatResponse(response), proto.FromError(err)
}

func (g grpcHandler) RecordActivityTaskHeartbeatByID(ctx context.Context, request *apiv1.RecordActivityTaskHeartbeatByIDRequest) (*apiv1.RecordActivityTaskHeartbeatByIDResponse, error) {
	response, err := g.h.RecordActivityTaskHeartbeatByID(ctx, proto.ToRecordActivityTaskHeartbeatByIDRequest(request))
	return proto.FromRecordActivityTaskHeartbeatByIDResponse(response), proto.FromError(err)
}

func (g grpcHandler) RegisterDomain(ctx context.Context, request *apiv1.RegisterDomainRequest) (*apiv1.RegisterDomainResponse, error) {
	err := g.h.RegisterDomain(ctx, proto.ToRegisterDomainRequest(request))
	return &apiv1.RegisterDomainResponse{}, proto.FromError(err)
}

func (g grpcHandler) RequestCancelWorkflowExecution(ctx context.Context, request *apiv1.RequestCancelWorkflowExecutionRequest) (*apiv1.RequestCancelWorkflowExecutionResponse, error) {
	err := g.h.RequestCancelWorkflowExecution(ctx, proto.ToRequestCancelWorkflowExecutionRequest(request))
	return &apiv1.RequestCancelWorkflowExecutionResponse{}, proto.FromError(err)
}

func (g grpcHandler) ResetStickyTaskList(ctx context.Context, request *apiv1.ResetStickyTaskListRequest) (*apiv1.ResetStickyTaskListResponse, error) {
	_, err := g.h.ResetStickyTaskList(ctx, proto.ToResetStickyTaskListRequest(request))
	return &apiv1.ResetStickyTaskListResponse{}, proto.FromError(err)
}

func (g grpcHandler) ResetWorkflowExecution(ctx context.Context, request *apiv1.ResetWorkflowExecutionRequest) (*apiv1.ResetWorkflowExecutionResponse, error) {
	response, err := g.h.ResetWorkflowExecution(ctx, proto.ToResetWorkflowExecutionRequest(request))
	return proto.FromResetWorkflowExecutionResponse(response), proto.FromError(err)
}

func (g grpcHandler) RespondActivityTaskCanceled(ctx context.Context, request *apiv1.RespondActivityTaskCanceledRequest) (*apiv1.RespondActivityTaskCanceledResponse, error) {
	err := g.h.RespondActivityTaskCanceled(ctx, proto.ToRespondActivityTaskCanceledRequest(request))
	return &apiv1.RespondActivityTaskCanceledResponse{}, proto.FromError(err)
}

func (g grpcHandler) RespondActivityTaskCanceledByID(ctx context.Context, request *apiv1.RespondActivityTaskCanceledByIDRequest) (*apiv1.RespondActivityTaskCanceledByIDResponse, error) {
	err := g.h.RespondActivityTaskCanceledByID(ctx, proto.ToRespondActivityTaskCanceledByIDRequest(request))
	return &apiv1.RespondActivityTaskCanceledByIDResponse{}, proto.FromError(err)
}

func (g grpcHandler) RespondActivityTaskCompleted(ctx context.Context, request *apiv1.RespondActivityTaskCompletedRequest) (*apiv1.RespondActivityTaskCompletedResponse, error) {
	err := g.h.RespondActivityTaskCompleted(ctx, proto.ToRespondActivityTaskCompletedRequest(request))
	return &apiv1.RespondActivityTaskCompletedResponse{}, proto.FromError(err)
}

func (g grpcHandler) RespondActivityTaskCompletedByID(ctx context.Context, request *apiv1.RespondActivityTaskCompletedByIDRequest) (*apiv1.RespondActivityTaskCompletedByIDResponse, error) {
	err := g.h.RespondActivityTaskCompletedByID(ctx, proto.ToRespondActivityTaskCompletedByIDRequest(request))
	return &apiv1.RespondActivityTaskCompletedByIDResponse{}, proto.FromError(err)
}

func (g grpcHandler) RespondActivityTaskFailed(ctx context.Context, request *apiv1.RespondActivityTaskFailedRequest) (*apiv1.RespondActivityTaskFailedResponse, error) {
	err := g.h.RespondActivityTaskFailed(ctx, proto.ToRespondActivityTaskFailedRequest(request))
	return &apiv1.RespondActivityTaskFailedResponse{}, proto.FromError(err)
}

func (g grpcHandler) RespondActivityTaskFailedByID(ctx context.Context, request *apiv1.RespondActivityTaskFailedByIDRequest) (*apiv1.RespondActivityTaskFailedByIDResponse, error) {
	err := g.h.RespondActivityTaskFailedByID(ctx, proto.ToRespondActivityTaskFailedByIDRequest(request))
	return &apiv1.RespondActivityTaskFailedByIDResponse{}, proto.FromError(err)
}

func (g grpcHandler) RespondDecisionTaskCompleted(ctx context.Context, request *apiv1.RespondDecisionTaskCompletedRequest) (*apiv1.RespondDecisionTaskCompletedResponse, error) {
	response, err := g.h.RespondDecisionTaskCompleted(ctx, proto.ToRespondDecisionTaskCompletedRequest(request))
	return proto.FromRespondDecisionTaskCompletedResponse(response), proto.FromError(err)
}

func (g grpcHandler) RespondDecisionTaskFailed(ctx context.Context, request *apiv1.RespondDecisionTaskFailedRequest) (*apiv1.RespondDecisionTaskFailedResponse, error) {
	err := g.h.RespondDecisionTaskFailed(ctx, proto.ToRespondDecisionTaskFailedRequest(request))
	return &apiv1.RespondDecisionTaskFailedResponse{}, proto.FromError(err)
}

func (g grpcHandler) RespondQueryTaskCompleted(ctx context.Context, request *apiv1.RespondQueryTaskCompletedRequest) (*apiv1.RespondQueryTaskCompletedResponse, error) {
	err := g.h.RespondQueryTaskCompleted(ctx, proto.ToRespondQueryTaskCompletedRequest(request))
	return &apiv1.RespondQueryTaskCompletedResponse{}, proto.FromError(err)
}

func (g grpcHandler) ScanWorkflowExecutions(ctx context.Context, request *apiv1.ScanWorkflowExecutionsRequest) (*apiv1.ScanWorkflowExecutionsResponse, error) {
	response, err := g.h.ScanWorkflowExecutions(ctx, proto.ToScanWorkflowExecutionsRequest(request))
	return proto.FromScanWorkflowExecutionsResponse(response), proto.FromError(err)
}

func (g grpcHandler) SignalWithStartWorkflowExecution(ctx context.Context, request *apiv1.SignalWithStartWorkflowExecutionRequest) (*apiv1.SignalWithStartWorkflowExecutionResponse, error) {
	response, err := g.h.SignalWithStartWorkflowExecution(ctx, proto.ToSignalWithStartWorkflowExecutionRequest(request))
	return proto.FromSignalWithStartWorkflowExecutionResponse(response), proto.FromError(err)
}

func (g grpcHandler) SignalWorkflowExecution(ctx context.Context, request *apiv1.SignalWorkflowExecutionRequest) (*apiv1.SignalWorkflowExecutionResponse, error) {
	err := g.h.SignalWorkflowExecution(ctx, proto.ToSignalWorkflowExecutionRequest(request))
	return &apiv1.SignalWorkflowExecutionResponse{}, proto.FromError(err)
}

func (g grpcHandler) StartWorkflowExecution(ctx context.Context, request *apiv1.StartWorkflowExecutionRequest) (*apiv1.StartWorkflowExecutionResponse, error) {
	response, err := g.h.StartWorkflowExecution(ctx, proto.ToStartWorkflowExecutionRequest(request))
	return proto.FromStartWorkflowExecutionResponse(response), proto.FromError(err)
}

func (g grpcHandler) TerminateWorkflowExecution(ctx context.Context, request *apiv1.TerminateWorkflowExecutionRequest) (*apiv1.TerminateWorkflowExecutionResponse, error) {
	err := g.h.TerminateWorkflowExecution(ctx, proto.ToTerminateWorkflowExecutionRequest(request))
	return &apiv1.TerminateWorkflowExecutionResponse{}, proto.FromError(err)
}

func (g grpcHandler) UpdateDomain(ctx context.Context, request *apiv1.UpdateDomainRequest) (*apiv1.UpdateDomainResponse, error) {
	response, err := g.h.UpdateDomain(ctx, proto.ToUpdateDomainRequest(request))
	return proto.FromUpdateDomainResponse(response), proto.FromError(err)
}
