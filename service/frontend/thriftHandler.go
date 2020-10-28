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
	return t.h.Health(ctx)
}

// CountWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) CountWorkflowExecutions(ctx context.Context, request *shared.CountWorkflowExecutionsRequest) (response *shared.CountWorkflowExecutionsResponse, err error) {
	return t.h.CountWorkflowExecutions(ctx, request)
}

// DeprecateDomain forwards request to the underlying handler
func (t ThriftHandler) DeprecateDomain(ctx context.Context, request *shared.DeprecateDomainRequest) (err error) {
	return t.h.DeprecateDomain(ctx, request)
}

// DescribeDomain forwards request to the underlying handler
func (t ThriftHandler) DescribeDomain(ctx context.Context, request *shared.DescribeDomainRequest) (response *shared.DescribeDomainResponse, err error) {
	return t.h.DescribeDomain(ctx, request)
}

// DescribeTaskList forwards request to the underlying handler
func (t ThriftHandler) DescribeTaskList(ctx context.Context, request *shared.DescribeTaskListRequest) (response *shared.DescribeTaskListResponse, err error) {
	return t.h.DescribeTaskList(ctx, request)
}

// DescribeWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) DescribeWorkflowExecution(ctx context.Context, request *shared.DescribeWorkflowExecutionRequest) (response *shared.DescribeWorkflowExecutionResponse, err error) {
	return t.h.DescribeWorkflowExecution(ctx, request)
}

// GetClusterInfo forwards request to the underlying handler
func (t ThriftHandler) GetClusterInfo(ctx context.Context) (response *shared.ClusterInfo, err error) {
	return t.h.GetClusterInfo(ctx)
}

// GetSearchAttributes forwards request to the underlying handler
func (t ThriftHandler) GetSearchAttributes(ctx context.Context) (response *shared.GetSearchAttributesResponse, err error) {
	return t.h.GetSearchAttributes(ctx)
}

// GetWorkflowExecutionHistory forwards request to the underlying handler
func (t ThriftHandler) GetWorkflowExecutionHistory(ctx context.Context, request *shared.GetWorkflowExecutionHistoryRequest) (response *shared.GetWorkflowExecutionHistoryResponse, err error) {
	return t.h.GetWorkflowExecutionHistory(ctx, request)
}

// ListArchivedWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) ListArchivedWorkflowExecutions(ctx context.Context, request *shared.ListArchivedWorkflowExecutionsRequest) (response *shared.ListArchivedWorkflowExecutionsResponse, err error) {
	return t.h.ListArchivedWorkflowExecutions(ctx, request)
}

// ListClosedWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) ListClosedWorkflowExecutions(ctx context.Context, request *shared.ListClosedWorkflowExecutionsRequest) (response *shared.ListClosedWorkflowExecutionsResponse, err error) {
	return t.h.ListClosedWorkflowExecutions(ctx, request)
}

// ListDomains forwards request to the underlying handler
func (t ThriftHandler) ListDomains(ctx context.Context, request *shared.ListDomainsRequest) (response *shared.ListDomainsResponse, err error) {
	return t.h.ListDomains(ctx, request)
}

// ListOpenWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) ListOpenWorkflowExecutions(ctx context.Context, request *shared.ListOpenWorkflowExecutionsRequest) (response *shared.ListOpenWorkflowExecutionsResponse, err error) {
	return t.h.ListOpenWorkflowExecutions(ctx, request)
}

// ListTaskListPartitions forwards request to the underlying handler
func (t ThriftHandler) ListTaskListPartitions(ctx context.Context, request *shared.ListTaskListPartitionsRequest) (response *shared.ListTaskListPartitionsResponse, err error) {
	return t.h.ListTaskListPartitions(ctx, request)
}

// ListWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) ListWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest) (response *shared.ListWorkflowExecutionsResponse, err error) {
	return t.h.ListWorkflowExecutions(ctx, request)
}

// PollForActivityTask forwards request to the underlying handler
func (t ThriftHandler) PollForActivityTask(ctx context.Context, request *shared.PollForActivityTaskRequest) (response *shared.PollForActivityTaskResponse, err error) {
	return t.h.PollForActivityTask(ctx, request)
}

// PollForDecisionTask forwards request to the underlying handler
func (t ThriftHandler) PollForDecisionTask(ctx context.Context, request *shared.PollForDecisionTaskRequest) (response *shared.PollForDecisionTaskResponse, err error) {
	return t.h.PollForDecisionTask(ctx, request)
}

// QueryWorkflow forwards request to the underlying handler
func (t ThriftHandler) QueryWorkflow(ctx context.Context, request *shared.QueryWorkflowRequest) (response *shared.QueryWorkflowResponse, err error) {
	return t.h.QueryWorkflow(ctx, request)
}

// RecordActivityTaskHeartbeat forwards request to the underlying handler
func (t ThriftHandler) RecordActivityTaskHeartbeat(ctx context.Context, request *shared.RecordActivityTaskHeartbeatRequest) (response *shared.RecordActivityTaskHeartbeatResponse, err error) {
	return t.h.RecordActivityTaskHeartbeat(ctx, request)
}

// RecordActivityTaskHeartbeatByID forwards request to the underlying handler
func (t ThriftHandler) RecordActivityTaskHeartbeatByID(ctx context.Context, request *shared.RecordActivityTaskHeartbeatByIDRequest) (response *shared.RecordActivityTaskHeartbeatResponse, err error) {
	return t.h.RecordActivityTaskHeartbeatByID(ctx, request)
}

// RegisterDomain forwards request to the underlying handler
func (t ThriftHandler) RegisterDomain(ctx context.Context, request *shared.RegisterDomainRequest) (err error) {
	return t.h.RegisterDomain(ctx, request)
}

// RequestCancelWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) RequestCancelWorkflowExecution(ctx context.Context, request *shared.RequestCancelWorkflowExecutionRequest) (err error) {
	return t.h.RequestCancelWorkflowExecution(ctx, request)
}

// ResetStickyTaskList forwards request to the underlying handler
func (t ThriftHandler) ResetStickyTaskList(ctx context.Context, request *shared.ResetStickyTaskListRequest) (response *shared.ResetStickyTaskListResponse, err error) {
	return t.h.ResetStickyTaskList(ctx, request)
}

// ResetWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) ResetWorkflowExecution(ctx context.Context, request *shared.ResetWorkflowExecutionRequest) (response *shared.ResetWorkflowExecutionResponse, err error) {
	return t.h.ResetWorkflowExecution(ctx, request)
}

// RespondActivityTaskCanceled forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCanceled(ctx context.Context, request *shared.RespondActivityTaskCanceledRequest) (err error) {
	return t.h.RespondActivityTaskCanceled(ctx, request)
}

// RespondActivityTaskCanceledByID forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCanceledByID(ctx context.Context, request *shared.RespondActivityTaskCanceledByIDRequest) (err error) {
	return t.h.RespondActivityTaskCanceledByID(ctx, request)
}

// RespondActivityTaskCompleted forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCompleted(ctx context.Context, request *shared.RespondActivityTaskCompletedRequest) (err error) {
	return t.h.RespondActivityTaskCompleted(ctx, request)
}

// RespondActivityTaskCompletedByID forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskCompletedByID(ctx context.Context, request *shared.RespondActivityTaskCompletedByIDRequest) (err error) {
	return t.h.RespondActivityTaskCompletedByID(ctx, request)
}

// RespondActivityTaskFailed forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskFailed(ctx context.Context, request *shared.RespondActivityTaskFailedRequest) (err error) {
	return t.h.RespondActivityTaskFailed(ctx, request)
}

// RespondActivityTaskFailedByID forwards request to the underlying handler
func (t ThriftHandler) RespondActivityTaskFailedByID(ctx context.Context, request *shared.RespondActivityTaskFailedByIDRequest) (err error) {
	return t.h.RespondActivityTaskFailedByID(ctx, request)
}

// RespondDecisionTaskCompleted forwards request to the underlying handler
func (t ThriftHandler) RespondDecisionTaskCompleted(ctx context.Context, request *shared.RespondDecisionTaskCompletedRequest) (response *shared.RespondDecisionTaskCompletedResponse, err error) {
	return t.h.RespondDecisionTaskCompleted(ctx, request)
}

// RespondDecisionTaskFailed forwards request to the underlying handler
func (t ThriftHandler) RespondDecisionTaskFailed(ctx context.Context, request *shared.RespondDecisionTaskFailedRequest) (err error) {
	return t.h.RespondDecisionTaskFailed(ctx, request)
}

// RespondQueryTaskCompleted forwards request to the underlying handler
func (t ThriftHandler) RespondQueryTaskCompleted(ctx context.Context, request *shared.RespondQueryTaskCompletedRequest) (err error) {
	return t.h.RespondQueryTaskCompleted(ctx, request)
}

// ScanWorkflowExecutions forwards request to the underlying handler
func (t ThriftHandler) ScanWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest) (response *shared.ListWorkflowExecutionsResponse, err error) {
	return t.h.ScanWorkflowExecutions(ctx, request)
}

// SignalWithStartWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) SignalWithStartWorkflowExecution(ctx context.Context, request *shared.SignalWithStartWorkflowExecutionRequest) (response *shared.StartWorkflowExecutionResponse, err error) {
	return t.h.SignalWithStartWorkflowExecution(ctx, request)
}

// SignalWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) SignalWorkflowExecution(ctx context.Context, request *shared.SignalWorkflowExecutionRequest) (err error) {
	return t.h.SignalWorkflowExecution(ctx, request)
}

// StartWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) StartWorkflowExecution(ctx context.Context, request *shared.StartWorkflowExecutionRequest) (response *shared.StartWorkflowExecutionResponse, err error) {
	return t.h.StartWorkflowExecution(ctx, request)
}

// TerminateWorkflowExecution forwards request to the underlying handler
func (t ThriftHandler) TerminateWorkflowExecution(ctx context.Context, request *shared.TerminateWorkflowExecutionRequest) (err error) {
	return t.h.TerminateWorkflowExecution(ctx, request)
}

// UpdateDomain forwards request to the underlying handler
func (t ThriftHandler) UpdateDomain(ctx context.Context, request *shared.UpdateDomainRequest) (response *shared.UpdateDomainResponse, err error) {
	return t.h.UpdateDomain(ctx, request)
}
