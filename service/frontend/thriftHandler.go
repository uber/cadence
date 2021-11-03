// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

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

	"github.com/uber/cadence/.gen/go/cadence/workflowserviceserver"
	"github.com/uber/cadence/.gen/go/health"
	"github.com/uber/cadence/.gen/go/health/metaserver"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/types/mapper/thrift"
)

// ThriftHandler wrap underlying handler and handles Thrift related type conversions
type ThriftHandler struct {
	h Handler
}

// NewThriftHandler creates Thrift handler on top of underlying handler
func NewThriftHandler(h Handler) ThriftHandler {
	return ThriftHandler{h}
}

func (t ThriftHandler) register(dispatcher *yarpc.Dispatcher) {
	dispatcher.Register(workflowserviceserver.New(t))
	dispatcher.Register(metaserver.New(t))
}

// Health forwards request to the underlying handler
func (t ThriftHandler) Health(ctx context.Context) (*health.HealthStatus, error) {
	response, err := t.h.Health(ctx)
	return thrift.FromHealthStatus(response), thrift.FromError(err)
}

// CountWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) CountWorkflowExecutions(ctx context.Context, request *shared.CountWorkflowExecutionsRequest) (*shared.CountWorkflowExecutionsResponse, error) {
	response, err := t.h.CountWorkflowExecutions(ctx, thrift.ToCountWorkflowExecutionsRequest(request))
	return thrift.FromCountWorkflowExecutionsResponse(response), thrift.FromError(err)
}

// DeprecateDomain forwards request to the underlying handler
func (t ThriftHandler) DeprecateDomain(ctx context.Context, request *shared.DeprecateDomainRequest) error {
	err := t.h.DeprecateDomain(ctx, thrift.ToDeprecateDomainRequest(request))
	return thrift.FromError(err)
}

// DescribeDomain forwards request to the underlying handler
func (t ThriftHandler) DescribeDomain(ctx context.Context, request *shared.DescribeDomainRequest) (*shared.DescribeDomainResponse, error) {
	response, err := t.h.DescribeDomain(ctx, thrift.ToDescribeDomainRequest(request))
	return thrift.FromDescribeDomainResponse(response), thrift.FromError(err)
}

// DescribeTaskList forwards request to the underlying handler
func (t ThriftHandler) DescribeTaskList(ctx context.Context, request *shared.DescribeTaskListRequest) (*shared.DescribeTaskListResponse, error) {
	response, err := t.h.DescribeTaskList(ctx, thrift.ToDescribeTaskListRequest(request))
	return thrift.FromDescribeTaskListResponse(response), thrift.FromError(err)
}

// DescribeWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) DescribeWorkflowExecution(ctx context.Context, request *shared.DescribeWorkflowExecutionRequest) (*shared.DescribeWorkflowExecutionResponse, error) {
	response, err := t.h.DescribeWorkflowExecution(ctx, thrift.ToDescribeWorkflowExecutionRequest(request))
	return thrift.FromDescribeWorkflowExecutionResponse(response), thrift.FromError(err)
}

// GetClusterInfo forwards request to the underlying handler
func (t ThriftHandler) GetClusterInfo(ctx context.Context) (*shared.ClusterInfo, error) {
	response, err := t.h.GetClusterInfo(ctx)
	return thrift.FromClusterInfo(response), thrift.FromError(err)
}

// GetSearchAttributes forwards request to the underlying handler
func (t ThriftHandler) GetSearchAttributes(ctx context.Context) (*shared.GetSearchAttributesResponse, error) {
	response, err := t.h.GetSearchAttributes(ctx)
	return thrift.FromGetSearchAttributesResponse(response), thrift.FromError(err)
}

// GetWorkflowExecutionHistory forwards request to the underlying handler
func (t ThriftHandler) GetWorkflowExecutionHistory(ctx context.Context, request *shared.GetWorkflowExecutionHistoryRequest) (*shared.GetWorkflowExecutionHistoryResponse, error) {
	response, err := t.h.GetWorkflowExecutionHistory(ctx, thrift.ToGetWorkflowExecutionHistoryRequest(request))
	return thrift.FromGetWorkflowExecutionHistoryResponse(response), thrift.FromError(err)
}

// ListArchivedWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) ListArchivedWorkflowExecutions(ctx context.Context, request *shared.ListArchivedWorkflowExecutionsRequest) (*shared.ListArchivedWorkflowExecutionsResponse, error) {
	response, err := t.h.ListArchivedWorkflowExecutions(ctx, thrift.ToListArchivedWorkflowExecutionsRequest(request))
	return thrift.FromListArchivedWorkflowExecutionsResponse(response), thrift.FromError(err)
}

// ListClosedWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) ListClosedWorkflowExecutions(ctx context.Context, request *shared.ListClosedWorkflowExecutionsRequest) (*shared.ListClosedWorkflowExecutionsResponse, error) {
	response, err := t.h.ListClosedWorkflowExecutions(ctx, thrift.ToListClosedWorkflowExecutionsRequest(request))
	return thrift.FromListClosedWorkflowExecutionsResponse(response), thrift.FromError(err)
}

// ListDomains forwards request to the underlying handler
func (t ThriftHandler) ListDomains(ctx context.Context, request *shared.ListDomainsRequest) (*shared.ListDomainsResponse, error) {
	response, err := t.h.ListDomains(ctx, thrift.ToListDomainsRequest(request))
	return thrift.FromListDomainsResponse(response), thrift.FromError(err)
}

// ListOpenWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) ListOpenWorkflowExecutions(ctx context.Context, request *shared.ListOpenWorkflowExecutionsRequest) (*shared.ListOpenWorkflowExecutionsResponse, error) {
	response, err := t.h.ListOpenWorkflowExecutions(ctx, thrift.ToListOpenWorkflowExecutionsRequest(request))
	return thrift.FromListOpenWorkflowExecutionsResponse(response), thrift.FromError(err)
}

// ListTaskListPartitions forwards request to the underlying handler
func (t ThriftHandler) ListTaskListPartitions(ctx context.Context, request *shared.ListTaskListPartitionsRequest) (*shared.ListTaskListPartitionsResponse, error) {
	response, err := t.h.ListTaskListPartitions(ctx, thrift.ToListTaskListPartitionsRequest(request))
	return thrift.FromListTaskListPartitionsResponse(response), thrift.FromError(err)
}

// GetTaskListsByDomain forwards request to the underlying handler
func (t ThriftHandler) GetTaskListsByDomain(ctx context.Context, request *shared.GetTaskListsByDomainRequest) (*shared.GetTaskListsByDomainResponse, error) {
	response, err := t.h.GetTaskListsByDomain(ctx, thrift.ToGetTaskListsByDomainRequest(request))
	return thrift.FromGetTaskListsByDomainResponse(response), thrift.FromError(err)
}

// ListWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) ListWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest) (*shared.ListWorkflowExecutionsResponse, error) {
	response, err := t.h.ListWorkflowExecutions(ctx, thrift.ToListWorkflowExecutionsRequest(request))
	return thrift.FromListWorkflowExecutionsResponse(response), thrift.FromError(err)
}

// PollForActivityTask forwards request to the underlying handler
func (t ThriftHandler) PollForActivityTask(ctx context.Context, request *shared.PollForActivityTaskRequest) (*shared.PollForActivityTaskResponse, error) {
	response, err := t.h.PollForActivityTask(ctx, thrift.ToPollForActivityTaskRequest(request))
	return thrift.FromPollForActivityTaskResponse(response), thrift.FromError(err)
}

// PollForDecisionTask forwards request to the underlying handler
func (t ThriftHandler) PollForDecisionTask(ctx context.Context, request *shared.PollForDecisionTaskRequest) (*shared.PollForDecisionTaskResponse, error) {
	response, err := t.h.PollForDecisionTask(ctx, thrift.ToPollForDecisionTaskRequest(request))
	return thrift.FromPollForDecisionTaskResponse(response), thrift.FromError(err)
}

// QueryWorkflow forwards request to the underlying handler
func (t ThriftHandler) QueryWorkflow(ctx context.Context, request *shared.QueryWorkflowRequest) (*shared.QueryWorkflowResponse, error) {
	response, err := t.h.QueryWorkflow(ctx, thrift.ToQueryWorkflowRequest(request))
	return thrift.FromQueryWorkflowResponse(response), thrift.FromError(err)
}

// RecordActivityTaskHeartbeat forwards request to the underlying handler
func (t ThriftHandler) RecordActivityTaskHeartbeat(ctx context.Context, request *shared.RecordActivityTaskHeartbeatRequest) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	response, err := t.h.RecordActivityTaskHeartbeat(ctx, thrift.ToRecordActivityTaskHeartbeatRequest(request))
	return thrift.FromRecordActivityTaskHeartbeatResponse(response), thrift.FromError(err)
}

// RecordActivityTaskHeartbeatByID forwards request to the underlying handler
func (t ThriftHandler) RecordActivityTaskHeartbeatByID(ctx context.Context, request *shared.RecordActivityTaskHeartbeatByIDRequest) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	response, err := t.h.RecordActivityTaskHeartbeatByID(ctx, thrift.ToRecordActivityTaskHeartbeatByIDRequest(request))
	return thrift.FromRecordActivityTaskHeartbeatResponse(response), thrift.FromError(err)
}

// RegisterDomain forwards request to the underlying handler
func (t ThriftHandler) RegisterDomain(ctx context.Context, request *shared.RegisterDomainRequest) error {
	err := t.h.RegisterDomain(ctx, thrift.ToRegisterDomainRequest(request))
	return thrift.FromError(err)
}

// RequestCancelWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) RequestCancelWorkflowExecution(ctx context.Context, request *shared.RequestCancelWorkflowExecutionRequest) error {
	err := t.h.RequestCancelWorkflowExecution(ctx, thrift.ToRequestCancelWorkflowExecutionRequest(request))
	return thrift.FromError(err)
}

// ResetStickyTaskList forwards request to the underlying handler
func (t ThriftHandler) ResetStickyTaskList(ctx context.Context, request *shared.ResetStickyTaskListRequest) (*shared.ResetStickyTaskListResponse, error) {
	response, err := t.h.ResetStickyTaskList(ctx, thrift.ToResetStickyTaskListRequest(request))
	return thrift.FromResetStickyTaskListResponse(response), thrift.FromError(err)
}

// ResetWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) ResetWorkflowExecution(ctx context.Context, request *shared.ResetWorkflowExecutionRequest) (*shared.ResetWorkflowExecutionResponse, error) {
	response, err := t.h.ResetWorkflowExecution(ctx, thrift.ToResetWorkflowExecutionRequest(request))
	return thrift.FromResetWorkflowExecutionResponse(response), thrift.FromError(err)
}

// RespondActivityTaskCanceled forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCanceled(ctx context.Context, request *shared.RespondActivityTaskCanceledRequest) error {
	err := t.h.RespondActivityTaskCanceled(ctx, thrift.ToRespondActivityTaskCanceledRequest(request))
	return thrift.FromError(err)
}

// RespondActivityTaskCanceledByID forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCanceledByID(ctx context.Context, request *shared.RespondActivityTaskCanceledByIDRequest) error {
	err := t.h.RespondActivityTaskCanceledByID(ctx, thrift.ToRespondActivityTaskCanceledByIDRequest(request))
	return thrift.FromError(err)
}

// RespondActivityTaskCompleted forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCompleted(ctx context.Context, request *shared.RespondActivityTaskCompletedRequest) error {
	err := t.h.RespondActivityTaskCompleted(ctx, thrift.ToRespondActivityTaskCompletedRequest(request))
	return thrift.FromError(err)
}

// RespondActivityTaskCompletedByID forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCompletedByID(ctx context.Context, request *shared.RespondActivityTaskCompletedByIDRequest) error {
	err := t.h.RespondActivityTaskCompletedByID(ctx, thrift.ToRespondActivityTaskCompletedByIDRequest(request))
	return thrift.FromError(err)
}

// RespondActivityTaskFailed forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskFailed(ctx context.Context, request *shared.RespondActivityTaskFailedRequest) error {
	err := t.h.RespondActivityTaskFailed(ctx, thrift.ToRespondActivityTaskFailedRequest(request))
	return thrift.FromError(err)
}

// RespondActivityTaskFailedByID forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskFailedByID(ctx context.Context, request *shared.RespondActivityTaskFailedByIDRequest) error {
	err := t.h.RespondActivityTaskFailedByID(ctx, thrift.ToRespondActivityTaskFailedByIDRequest(request))
	return thrift.FromError(err)
}

// RespondDecisionTaskCompleted forwards request to the underlying handler
func (t ThriftHandler) RespondDecisionTaskCompleted(ctx context.Context, request *shared.RespondDecisionTaskCompletedRequest) (*shared.RespondDecisionTaskCompletedResponse, error) {
	response, err := t.h.RespondDecisionTaskCompleted(ctx, thrift.ToRespondDecisionTaskCompletedRequest(request))
	return thrift.FromRespondDecisionTaskCompletedResponse(response), thrift.FromError(err)
}

// RespondDecisionTaskFailed forwards request to the underlying handler
func (t ThriftHandler) RespondDecisionTaskFailed(ctx context.Context, request *shared.RespondDecisionTaskFailedRequest) error {
	err := t.h.RespondDecisionTaskFailed(ctx, thrift.ToRespondDecisionTaskFailedRequest(request))
	return thrift.FromError(err)
}

// RespondQueryTaskCompleted forwards request to the underlying handler
func (t ThriftHandler) RespondQueryTaskCompleted(ctx context.Context, request *shared.RespondQueryTaskCompletedRequest) error {
	err := t.h.RespondQueryTaskCompleted(ctx, thrift.ToRespondQueryTaskCompletedRequest(request))
	return thrift.FromError(err)
}

// ScanWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) ScanWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest) (*shared.ListWorkflowExecutionsResponse, error) {
	response, err := t.h.ScanWorkflowExecutions(ctx, thrift.ToListWorkflowExecutionsRequest(request))
	return thrift.FromListWorkflowExecutionsResponse(response), thrift.FromError(err)
}

// SignalWithStartWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) SignalWithStartWorkflowExecution(ctx context.Context, request *shared.SignalWithStartWorkflowExecutionRequest) (*shared.StartWorkflowExecutionResponse, error) {
	response, err := t.h.SignalWithStartWorkflowExecution(ctx, thrift.ToSignalWithStartWorkflowExecutionRequest(request))
	return thrift.FromStartWorkflowExecutionResponse(response), thrift.FromError(err)
}

// SignalWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) SignalWorkflowExecution(ctx context.Context, request *shared.SignalWorkflowExecutionRequest) error {
	err := t.h.SignalWorkflowExecution(ctx, thrift.ToSignalWorkflowExecutionRequest(request))
	return thrift.FromError(err)
}

// StartWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) StartWorkflowExecution(ctx context.Context, request *shared.StartWorkflowExecutionRequest) (*shared.StartWorkflowExecutionResponse, error) {
	response, err := t.h.StartWorkflowExecution(ctx, thrift.ToStartWorkflowExecutionRequest(request))
	return thrift.FromStartWorkflowExecutionResponse(response), thrift.FromError(err)
}

// TerminateWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) TerminateWorkflowExecution(ctx context.Context, request *shared.TerminateWorkflowExecutionRequest) error {
	err := t.h.TerminateWorkflowExecution(ctx, thrift.ToTerminateWorkflowExecutionRequest(request))
	return thrift.FromError(err)
}

// UpdateDomain forwards request to the underlying handler
func (t ThriftHandler) UpdateDomain(ctx context.Context, request *shared.UpdateDomainRequest) (*shared.UpdateDomainResponse, error) {
	response, err := t.h.UpdateDomain(ctx, thrift.ToUpdateDomainRequest(request))
	return thrift.FromUpdateDomainResponse(response), thrift.FromError(err)
}
