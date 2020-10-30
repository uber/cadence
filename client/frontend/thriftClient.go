// Copyright (c) 2020 Uber Technologies, Inc.
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

	"github.com/uber/cadence/.gen/go/cadence/workflowserviceclient"
	"github.com/uber/cadence/.gen/go/shared"
)

type thriftClient struct {
	c workflowserviceclient.Interface
}

// NewThriftClient creates a new instance of Client with thrift protocol
func NewThriftClient(c workflowserviceclient.Interface) Client {
	return thriftClient{c}
}

func (t thriftClient) CountWorkflowExecutions(ctx context.Context, request *shared.CountWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.CountWorkflowExecutionsResponse, error) {
	return t.c.CountWorkflowExecutions(ctx, request, opts...)
}

func (t thriftClient) DeprecateDomain(ctx context.Context, request *shared.DeprecateDomainRequest, opts ...yarpc.CallOption) error {
	return t.c.DeprecateDomain(ctx, request, opts...)
}

func (t thriftClient) DescribeDomain(ctx context.Context, request *shared.DescribeDomainRequest, opts ...yarpc.CallOption) (*shared.DescribeDomainResponse, error) {
	return t.c.DescribeDomain(ctx, request, opts...)
}

func (t thriftClient) DescribeTaskList(ctx context.Context, request *shared.DescribeTaskListRequest, opts ...yarpc.CallOption) (*shared.DescribeTaskListResponse, error) {
	return t.c.DescribeTaskList(ctx, request, opts...)
}

func (t thriftClient) DescribeWorkflowExecution(ctx context.Context, request *shared.DescribeWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.DescribeWorkflowExecutionResponse, error) {
	return t.c.DescribeWorkflowExecution(ctx, request, opts...)
}

func (t thriftClient) GetClusterInfo(ctx context.Context, opts ...yarpc.CallOption) (*shared.ClusterInfo, error) {
	return t.c.GetClusterInfo(ctx, opts...)
}

func (t thriftClient) GetSearchAttributes(ctx context.Context, opts ...yarpc.CallOption) (*shared.GetSearchAttributesResponse, error) {
	return t.c.GetSearchAttributes(ctx, opts...)
}

func (t thriftClient) GetWorkflowExecutionHistory(ctx context.Context, request *shared.GetWorkflowExecutionHistoryRequest, opts ...yarpc.CallOption) (*shared.GetWorkflowExecutionHistoryResponse, error) {
	return t.c.GetWorkflowExecutionHistory(ctx, request, opts...)
}

func (t thriftClient) ListArchivedWorkflowExecutions(ctx context.Context, request *shared.ListArchivedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListArchivedWorkflowExecutionsResponse, error) {
	return t.c.ListArchivedWorkflowExecutions(ctx, request, opts...)
}

func (t thriftClient) ListClosedWorkflowExecutions(ctx context.Context, request *shared.ListClosedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListClosedWorkflowExecutionsResponse, error) {
	return t.c.ListClosedWorkflowExecutions(ctx, request, opts...)
}

func (t thriftClient) ListDomains(ctx context.Context, request *shared.ListDomainsRequest, opts ...yarpc.CallOption) (*shared.ListDomainsResponse, error) {
	return t.c.ListDomains(ctx, request, opts...)
}

func (t thriftClient) ListOpenWorkflowExecutions(ctx context.Context, request *shared.ListOpenWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListOpenWorkflowExecutionsResponse, error) {
	return t.c.ListOpenWorkflowExecutions(ctx, request, opts...)
}

func (t thriftClient) ListTaskListPartitions(ctx context.Context, request *shared.ListTaskListPartitionsRequest, opts ...yarpc.CallOption) (*shared.ListTaskListPartitionsResponse, error) {
	return t.c.ListTaskListPartitions(ctx, request, opts...)
}

func (t thriftClient) ListWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListWorkflowExecutionsResponse, error) {
	return t.c.ListWorkflowExecutions(ctx, request, opts...)
}

func (t thriftClient) PollForActivityTask(ctx context.Context, request *shared.PollForActivityTaskRequest, opts ...yarpc.CallOption) (*shared.PollForActivityTaskResponse, error) {
	return t.c.PollForActivityTask(ctx, request, opts...)
}

func (t thriftClient) PollForDecisionTask(ctx context.Context, request *shared.PollForDecisionTaskRequest, opts ...yarpc.CallOption) (*shared.PollForDecisionTaskResponse, error) {
	return t.c.PollForDecisionTask(ctx, request, opts...)
}

func (t thriftClient) QueryWorkflow(ctx context.Context, request *shared.QueryWorkflowRequest, opts ...yarpc.CallOption) (*shared.QueryWorkflowResponse, error) {
	return t.c.QueryWorkflow(ctx, request, opts...)
}

func (t thriftClient) RecordActivityTaskHeartbeat(ctx context.Context, request *shared.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	return t.c.RecordActivityTaskHeartbeat(ctx, request, opts...)
}

func (t thriftClient) RecordActivityTaskHeartbeatByID(ctx context.Context, request *shared.RecordActivityTaskHeartbeatByIDRequest, opts ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	return t.c.RecordActivityTaskHeartbeatByID(ctx, request, opts...)
}

func (t thriftClient) RegisterDomain(ctx context.Context, request *shared.RegisterDomainRequest, opts ...yarpc.CallOption) error {
	return t.c.RegisterDomain(ctx, request, opts...)
}

func (t thriftClient) RequestCancelWorkflowExecution(ctx context.Context, request *shared.RequestCancelWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	return t.c.RequestCancelWorkflowExecution(ctx, request, opts...)
}

func (t thriftClient) ResetStickyTaskList(ctx context.Context, request *shared.ResetStickyTaskListRequest, opts ...yarpc.CallOption) (*shared.ResetStickyTaskListResponse, error) {
	return t.c.ResetStickyTaskList(ctx, request, opts...)
}

func (t thriftClient) ResetWorkflowExecution(ctx context.Context, request *shared.ResetWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.ResetWorkflowExecutionResponse, error) {
	return t.c.ResetWorkflowExecution(ctx, request, opts...)
}

func (t thriftClient) RespondActivityTaskCanceled(ctx context.Context, request *shared.RespondActivityTaskCanceledRequest, opts ...yarpc.CallOption) error {
	return t.c.RespondActivityTaskCanceled(ctx, request, opts...)
}

func (t thriftClient) RespondActivityTaskCanceledByID(ctx context.Context, request *shared.RespondActivityTaskCanceledByIDRequest, opts ...yarpc.CallOption) error {
	return t.c.RespondActivityTaskCanceledByID(ctx, request, opts...)
}

func (t thriftClient) RespondActivityTaskCompleted(ctx context.Context, request *shared.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption) error {
	return t.c.RespondActivityTaskCompleted(ctx, request, opts...)
}

func (t thriftClient) RespondActivityTaskCompletedByID(ctx context.Context, request *shared.RespondActivityTaskCompletedByIDRequest, opts ...yarpc.CallOption) error {
	return t.c.RespondActivityTaskCompletedByID(ctx, request, opts...)
}

func (t thriftClient) RespondActivityTaskFailed(ctx context.Context, request *shared.RespondActivityTaskFailedRequest, opts ...yarpc.CallOption) error {
	return t.c.RespondActivityTaskFailed(ctx, request, opts...)
}

func (t thriftClient) RespondActivityTaskFailedByID(ctx context.Context, request *shared.RespondActivityTaskFailedByIDRequest, opts ...yarpc.CallOption) error {
	return t.c.RespondActivityTaskFailedByID(ctx, request, opts...)
}

func (t thriftClient) RespondDecisionTaskCompleted(ctx context.Context, request *shared.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption) (*shared.RespondDecisionTaskCompletedResponse, error) {
	return t.c.RespondDecisionTaskCompleted(ctx, request, opts...)
}

func (t thriftClient) RespondDecisionTaskFailed(ctx context.Context, request *shared.RespondDecisionTaskFailedRequest, opts ...yarpc.CallOption) error {
	return t.c.RespondDecisionTaskFailed(ctx, request, opts...)
}

func (t thriftClient) RespondQueryTaskCompleted(ctx context.Context, request *shared.RespondQueryTaskCompletedRequest, opts ...yarpc.CallOption) error {
	return t.c.RespondQueryTaskCompleted(ctx, request, opts...)
}

func (t thriftClient) ScanWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListWorkflowExecutionsResponse, error) {
	return t.c.ScanWorkflowExecutions(ctx, request, opts...)
}

func (t thriftClient) SignalWithStartWorkflowExecution(ctx context.Context, request *shared.SignalWithStartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
	return t.c.SignalWithStartWorkflowExecution(ctx, request, opts...)
}

func (t thriftClient) SignalWorkflowExecution(ctx context.Context, request *shared.SignalWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	return t.c.SignalWorkflowExecution(ctx, request, opts...)
}

func (t thriftClient) StartWorkflowExecution(ctx context.Context, request *shared.StartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
	return t.c.StartWorkflowExecution(ctx, request, opts...)
}

func (t thriftClient) TerminateWorkflowExecution(ctx context.Context, request *shared.TerminateWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	return t.c.TerminateWorkflowExecution(ctx, request, opts...)
}

func (t thriftClient) UpdateDomain(ctx context.Context, request *shared.UpdateDomainRequest, opts ...yarpc.CallOption) (*shared.UpdateDomainResponse, error) {
	return t.c.UpdateDomain(ctx, request, opts...)
}
