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
func (t ThriftHandler) Health(ctx context.Context) (response *health.HealthStatus, err error) {
	response, err = t.h.Health(ctx)
	return response, thrift.FromError(err)
}

// CountWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) CountWorkflowExecutions(ctx context.Context, request *shared.CountWorkflowExecutionsRequest) (*shared.CountWorkflowExecutionsResponse, error) {
	response, err := t.h.CountWorkflowExecutions(ctx, thrift.ToCountWorkflowExecutionsRequest(request))
	return thrift.FromCountWorkflowExecutionsResponse(response), thrift.FromError(err)
}

// DeprecateDomain forwards request to the underlying handler
func (t ThriftHandler) DeprecateDomain(ctx context.Context, request *shared.DeprecateDomainRequest) (err error) {
	err = t.h.DeprecateDomain(ctx, request)
	return thrift.FromError(err)
}

// DescribeDomain forwards request to the underlying handler
func (t ThriftHandler) DescribeDomain(ctx context.Context, request *shared.DescribeDomainRequest) (response *shared.DescribeDomainResponse, err error) {
	response, err = t.h.DescribeDomain(ctx, request)
	return response, thrift.FromError(err)
}

// DescribeTaskList forwards request to the underlying handler
func (t ThriftHandler) DescribeTaskList(ctx context.Context, request *shared.DescribeTaskListRequest) (response *shared.DescribeTaskListResponse, err error) {
	response, err = t.h.DescribeTaskList(ctx, request)
	return response, thrift.FromError(err)
}

// DescribeWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) DescribeWorkflowExecution(ctx context.Context, request *shared.DescribeWorkflowExecutionRequest) (response *shared.DescribeWorkflowExecutionResponse, err error) {
	response, err = t.h.DescribeWorkflowExecution(ctx, request)
	return response, thrift.FromError(err)
}

// GetClusterInfo forwards request to the underlying handler
func (t ThriftHandler) GetClusterInfo(ctx context.Context) (response *shared.ClusterInfo, err error) {
	response, err = t.h.GetClusterInfo(ctx)
	return response, thrift.FromError(err)
}

// GetSearchAttributes forwards request to the underlying handler
func (t ThriftHandler) GetSearchAttributes(ctx context.Context) (response *shared.GetSearchAttributesResponse, err error) {
	response, err = t.h.GetSearchAttributes(ctx)
	return response, thrift.FromError(err)
}

// GetWorkflowExecutionHistory forwards request to the underlying handler
func (t ThriftHandler) GetWorkflowExecutionHistory(ctx context.Context, request *shared.GetWorkflowExecutionHistoryRequest) (response *shared.GetWorkflowExecutionHistoryResponse, err error) {
	response, err = t.h.GetWorkflowExecutionHistory(ctx, request)
	return response, thrift.FromError(err)
}

// ListArchivedWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) ListArchivedWorkflowExecutions(ctx context.Context, request *shared.ListArchivedWorkflowExecutionsRequest) (response *shared.ListArchivedWorkflowExecutionsResponse, err error) {
	response, err = t.h.ListArchivedWorkflowExecutions(ctx, request)
	return response, thrift.FromError(err)
}

// ListClosedWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) ListClosedWorkflowExecutions(ctx context.Context, request *shared.ListClosedWorkflowExecutionsRequest) (*shared.ListClosedWorkflowExecutionsResponse, error) {
	response, err := t.h.ListClosedWorkflowExecutions(ctx, thrift.ToListClosedWorkflowExecutionsRequest(request))
	return thrift.FromListClosedWorkflowExecutionsResponse(response), thrift.FromError(err)
}

// ListDomains forwards request to the underlying handler
func (t ThriftHandler) ListDomains(ctx context.Context, request *shared.ListDomainsRequest) (response *shared.ListDomainsResponse, err error) {
	response, err = t.h.ListDomains(ctx, request)
	return response, thrift.FromError(err)
}

// ListOpenWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) ListOpenWorkflowExecutions(ctx context.Context, request *shared.ListOpenWorkflowExecutionsRequest) (*shared.ListOpenWorkflowExecutionsResponse, error) {
	response, err := t.h.ListOpenWorkflowExecutions(ctx, thrift.ToListOpenWorkflowExecutionsRequest(request))
	return thrift.FromListOpenWorkflowExecutionsResponse(response), thrift.FromError(err)
}

// ListTaskListPartitions forwards request to the underlying handler
func (t ThriftHandler) ListTaskListPartitions(ctx context.Context, request *shared.ListTaskListPartitionsRequest) (response *shared.ListTaskListPartitionsResponse, err error) {
	response, err = t.h.ListTaskListPartitions(ctx, request)
	return response, thrift.FromError(err)
}

// ListWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) ListWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest) (*shared.ListWorkflowExecutionsResponse, error) {
	response, err := t.h.ListWorkflowExecutions(ctx, thrift.ToListWorkflowExecutionsRequest(request))
	return thrift.FromListWorkflowExecutionsResponse(response), thrift.FromError(err)
}

// PollForActivityTask forwards request to the underlying handler
func (t ThriftHandler) PollForActivityTask(ctx context.Context, request *shared.PollForActivityTaskRequest) (response *shared.PollForActivityTaskResponse, err error) {
	response, err = t.h.PollForActivityTask(ctx, request)
	return response, thrift.FromError(err)
}

// PollForDecisionTask forwards request to the underlying handler
func (t ThriftHandler) PollForDecisionTask(ctx context.Context, request *shared.PollForDecisionTaskRequest) (response *shared.PollForDecisionTaskResponse, err error) {
	response, err = t.h.PollForDecisionTask(ctx, request)
	return response, thrift.FromError(err)
}

// QueryWorkflow forwards request to the underlying handler
func (t ThriftHandler) QueryWorkflow(ctx context.Context, request *shared.QueryWorkflowRequest) (response *shared.QueryWorkflowResponse, err error) {
	response, err = t.h.QueryWorkflow(ctx, request)
	return response, thrift.FromError(err)
}

// RecordActivityTaskHeartbeat forwards request to the underlying handler
func (t ThriftHandler) RecordActivityTaskHeartbeat(ctx context.Context, request *shared.RecordActivityTaskHeartbeatRequest) (response *shared.RecordActivityTaskHeartbeatResponse, err error) {
	response, err = t.h.RecordActivityTaskHeartbeat(ctx, request)
	return response, thrift.FromError(err)
}

// RecordActivityTaskHeartbeatByID forwards request to the underlying handler
func (t ThriftHandler) RecordActivityTaskHeartbeatByID(ctx context.Context, request *shared.RecordActivityTaskHeartbeatByIDRequest) (response *shared.RecordActivityTaskHeartbeatResponse, err error) {
	response, err = t.h.RecordActivityTaskHeartbeatByID(ctx, request)
	return response, thrift.FromError(err)
}

// RegisterDomain forwards request to the underlying handler
func (t ThriftHandler) RegisterDomain(ctx context.Context, request *shared.RegisterDomainRequest) (err error) {
	err = t.h.RegisterDomain(ctx, request)
	return thrift.FromError(err)
}

// RequestCancelWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) RequestCancelWorkflowExecution(ctx context.Context, request *shared.RequestCancelWorkflowExecutionRequest) (err error) {
	err = t.h.RequestCancelWorkflowExecution(ctx, request)
	return thrift.FromError(err)
}

// ResetStickyTaskList forwards request to the underlying handler
func (t ThriftHandler) ResetStickyTaskList(ctx context.Context, request *shared.ResetStickyTaskListRequest) (response *shared.ResetStickyTaskListResponse, err error) {
	response, err = t.h.ResetStickyTaskList(ctx, request)
	return response, thrift.FromError(err)
}

// ResetWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) ResetWorkflowExecution(ctx context.Context, request *shared.ResetWorkflowExecutionRequest) (response *shared.ResetWorkflowExecutionResponse, err error) {
	response, err = t.h.ResetWorkflowExecution(ctx, request)
	return response, thrift.FromError(err)
}

// RespondActivityTaskCanceled forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCanceled(ctx context.Context, request *shared.RespondActivityTaskCanceledRequest) (err error) {
	err = t.h.RespondActivityTaskCanceled(ctx, request)
	return thrift.FromError(err)
}

// RespondActivityTaskCanceledByID forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCanceledByID(ctx context.Context, request *shared.RespondActivityTaskCanceledByIDRequest) (err error) {
	err = t.h.RespondActivityTaskCanceledByID(ctx, request)
	return thrift.FromError(err)
}

// RespondActivityTaskCompleted forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCompleted(ctx context.Context, request *shared.RespondActivityTaskCompletedRequest) (err error) {
	err = t.h.RespondActivityTaskCompleted(ctx, request)
	return thrift.FromError(err)
}

// RespondActivityTaskCompletedByID forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCompletedByID(ctx context.Context, request *shared.RespondActivityTaskCompletedByIDRequest) (err error) {
	err = t.h.RespondActivityTaskCompletedByID(ctx, request)
	return thrift.FromError(err)
}

// RespondActivityTaskFailed forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskFailed(ctx context.Context, request *shared.RespondActivityTaskFailedRequest) (err error) {
	err = t.h.RespondActivityTaskFailed(ctx, request)
	return thrift.FromError(err)
}

// RespondActivityTaskFailedByID forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskFailedByID(ctx context.Context, request *shared.RespondActivityTaskFailedByIDRequest) (err error) {
	err = t.h.RespondActivityTaskFailedByID(ctx, request)
	return thrift.FromError(err)
}

// RespondDecisionTaskCompleted forwards request to the underlying handler
func (t ThriftHandler) RespondDecisionTaskCompleted(ctx context.Context, request *shared.RespondDecisionTaskCompletedRequest) (response *shared.RespondDecisionTaskCompletedResponse, err error) {
	response, err = t.h.RespondDecisionTaskCompleted(ctx, request)
	return response, thrift.FromError(err)
}

// RespondDecisionTaskFailed forwards request to the underlying handler
func (t ThriftHandler) RespondDecisionTaskFailed(ctx context.Context, request *shared.RespondDecisionTaskFailedRequest) (err error) {
	err = t.h.RespondDecisionTaskFailed(ctx, request)
	return thrift.FromError(err)
}

// RespondQueryTaskCompleted forwards request to the underlying handler
func (t ThriftHandler) RespondQueryTaskCompleted(ctx context.Context, request *shared.RespondQueryTaskCompletedRequest) (err error) {
	err = t.h.RespondQueryTaskCompleted(ctx, request)
	return thrift.FromError(err)
}

// ScanWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) ScanWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest) (*shared.ListWorkflowExecutionsResponse, error) {
	response, err := t.h.ScanWorkflowExecutions(ctx, thrift.ToListWorkflowExecutionsRequest(request))
	return thrift.FromListWorkflowExecutionsResponse(response), thrift.FromError(err)
}

// SignalWithStartWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) SignalWithStartWorkflowExecution(ctx context.Context, request *shared.SignalWithStartWorkflowExecutionRequest) (response *shared.StartWorkflowExecutionResponse, err error) {
	response, err = t.h.SignalWithStartWorkflowExecution(ctx, request)
	return response, thrift.FromError(err)
}

// SignalWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) SignalWorkflowExecution(ctx context.Context, request *shared.SignalWorkflowExecutionRequest) (err error) {
	err = t.h.SignalWorkflowExecution(ctx, request)
	return thrift.FromError(err)
}

// StartWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) StartWorkflowExecution(ctx context.Context, request *shared.StartWorkflowExecutionRequest) (response *shared.StartWorkflowExecutionResponse, err error) {
	response, err = t.h.StartWorkflowExecution(ctx, request)
	return response, thrift.FromError(err)
}

// TerminateWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) TerminateWorkflowExecution(ctx context.Context, request *shared.TerminateWorkflowExecutionRequest) (err error) {
	err = t.h.TerminateWorkflowExecution(ctx, request)
	return thrift.FromError(err)
}

// UpdateDomain forwards request to the underlying handler
func (t ThriftHandler) UpdateDomain(ctx context.Context, request *shared.UpdateDomainRequest) (response *shared.UpdateDomainResponse, err error) {
	response, err = t.h.UpdateDomain(ctx, request)
	return response, thrift.FromError(err)
}
