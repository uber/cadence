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
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"
)

type thriftClient struct {
	c workflowserviceclient.Interface
}

// NewThriftClient creates a new instance of Client with thrift protocol
func NewThriftClient(c workflowserviceclient.Interface) Client {
	return thriftClient{c}
}

func (t thriftClient) CountWorkflowExecutions(ctx context.Context, request *types.CountWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*types.CountWorkflowExecutionsResponse, error) {
	response, err := t.c.CountWorkflowExecutions(ctx, thrift.FromCountWorkflowExecutionsRequest(request), opts...)
	return thrift.ToCountWorkflowExecutionsResponse(response), thrift.ToError(err)
}

func (t thriftClient) DeprecateDomain(ctx context.Context, request *types.DeprecateDomainRequest, opts ...yarpc.CallOption) error {
	err := t.c.DeprecateDomain(ctx, thrift.FromDeprecateDomainRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) DescribeDomain(ctx context.Context, request *types.DescribeDomainRequest, opts ...yarpc.CallOption) (*types.DescribeDomainResponse, error) {
	response, err := t.c.DescribeDomain(ctx, thrift.FromDescribeDomainRequest(request), opts...)
	return thrift.ToDescribeDomainResponse(response), thrift.ToError(err)
}

func (t thriftClient) DescribeTaskList(ctx context.Context, request *types.DescribeTaskListRequest, opts ...yarpc.CallOption) (*types.DescribeTaskListResponse, error) {
	response, err := t.c.DescribeTaskList(ctx, thrift.FromDescribeTaskListRequest(request), opts...)
	return thrift.ToDescribeTaskListResponse(response), thrift.ToError(err)
}

func (t thriftClient) DescribeWorkflowExecution(ctx context.Context, request *types.DescribeWorkflowExecutionRequest, opts ...yarpc.CallOption) (*types.DescribeWorkflowExecutionResponse, error) {
	response, err := t.c.DescribeWorkflowExecution(ctx, thrift.FromDescribeWorkflowExecutionRequest(request), opts...)
	return thrift.ToDescribeWorkflowExecutionResponse(response), thrift.ToError(err)
}

func (t thriftClient) GetClusterInfo(ctx context.Context, opts ...yarpc.CallOption) (*types.ClusterInfo, error) {
	response, err := t.c.GetClusterInfo(ctx, opts...)
	return thrift.ToClusterInfo(response), thrift.ToError(err)
}

func (t thriftClient) GetSearchAttributes(ctx context.Context, opts ...yarpc.CallOption) (*types.GetSearchAttributesResponse, error) {
	response, err := t.c.GetSearchAttributes(ctx, opts...)
	return thrift.ToGetSearchAttributesResponse(response), thrift.ToError(err)
}

func (t thriftClient) GetWorkflowExecutionHistory(ctx context.Context, request *types.GetWorkflowExecutionHistoryRequest, opts ...yarpc.CallOption) (*types.GetWorkflowExecutionHistoryResponse, error) {
	response, err := t.c.GetWorkflowExecutionHistory(ctx, thrift.FromGetWorkflowExecutionHistoryRequest(request), opts...)
	return thrift.ToGetWorkflowExecutionHistoryResponse(response), thrift.ToError(err)
}

func (t thriftClient) ListArchivedWorkflowExecutions(ctx context.Context, request *types.ListArchivedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*types.ListArchivedWorkflowExecutionsResponse, error) {
	response, err := t.c.ListArchivedWorkflowExecutions(ctx, thrift.FromListArchivedWorkflowExecutionsRequest(request), opts...)
	return thrift.ToListArchivedWorkflowExecutionsResponse(response), thrift.ToError(err)
}

func (t thriftClient) ListClosedWorkflowExecutions(ctx context.Context, request *types.ListClosedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*types.ListClosedWorkflowExecutionsResponse, error) {
	response, err := t.c.ListClosedWorkflowExecutions(ctx, thrift.FromListClosedWorkflowExecutionsRequest(request), opts...)
	return thrift.ToListClosedWorkflowExecutionsResponse(response), thrift.ToError(err)
}

func (t thriftClient) ListDomains(ctx context.Context, request *types.ListDomainsRequest, opts ...yarpc.CallOption) (*types.ListDomainsResponse, error) {
	response, err := t.c.ListDomains(ctx, thrift.FromListDomainsRequest(request), opts...)
	return thrift.ToListDomainsResponse(response), thrift.ToError(err)
}

func (t thriftClient) ListOpenWorkflowExecutions(ctx context.Context, request *types.ListOpenWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*types.ListOpenWorkflowExecutionsResponse, error) {
	response, err := t.c.ListOpenWorkflowExecutions(ctx, thrift.FromListOpenWorkflowExecutionsRequest(request), opts...)
	return thrift.ToListOpenWorkflowExecutionsResponse(response), thrift.ToError(err)
}

func (t thriftClient) ListTaskListPartitions(ctx context.Context, request *types.ListTaskListPartitionsRequest, opts ...yarpc.CallOption) (*types.ListTaskListPartitionsResponse, error) {
	response, err := t.c.ListTaskListPartitions(ctx, thrift.FromListTaskListPartitionsRequest(request), opts...)
	return thrift.ToListTaskListPartitionsResponse(response), thrift.ToError(err)
}

func (t thriftClient) GetTaskListsByDomain(ctx context.Context, request *types.GetTaskListsByDomainRequest, opts ...yarpc.CallOption) (*types.GetTaskListsByDomainResponse, error) {
	response, err := t.c.GetTaskListsByDomain(ctx, thrift.FromGetTaskListsByDomainRequest(request), opts...)
	return thrift.ToGetTaskListsByDomainResponse(response), thrift.ToError(err)
}

func (t thriftClient) ListWorkflowExecutions(ctx context.Context, request *types.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*types.ListWorkflowExecutionsResponse, error) {
	response, err := t.c.ListWorkflowExecutions(ctx, thrift.FromListWorkflowExecutionsRequest(request), opts...)
	return thrift.ToListWorkflowExecutionsResponse(response), thrift.ToError(err)
}

func (t thriftClient) PollForActivityTask(ctx context.Context, request *types.PollForActivityTaskRequest, opts ...yarpc.CallOption) (*types.PollForActivityTaskResponse, error) {
	response, err := t.c.PollForActivityTask(ctx, thrift.FromPollForActivityTaskRequest(request), opts...)
	return thrift.ToPollForActivityTaskResponse(response), thrift.ToError(err)
}

func (t thriftClient) PollForDecisionTask(ctx context.Context, request *types.PollForDecisionTaskRequest, opts ...yarpc.CallOption) (*types.PollForDecisionTaskResponse, error) {
	response, err := t.c.PollForDecisionTask(ctx, thrift.FromPollForDecisionTaskRequest(request), opts...)
	return thrift.ToPollForDecisionTaskResponse(response), thrift.ToError(err)
}

func (t thriftClient) QueryWorkflow(ctx context.Context, request *types.QueryWorkflowRequest, opts ...yarpc.CallOption) (*types.QueryWorkflowResponse, error) {
	response, err := t.c.QueryWorkflow(ctx, thrift.FromQueryWorkflowRequest(request), opts...)
	return thrift.ToQueryWorkflowResponse(response), thrift.ToError(err)
}

func (t thriftClient) RecordActivityTaskHeartbeat(ctx context.Context, request *types.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) (*types.RecordActivityTaskHeartbeatResponse, error) {
	response, err := t.c.RecordActivityTaskHeartbeat(ctx, thrift.FromRecordActivityTaskHeartbeatRequest(request), opts...)
	return thrift.ToRecordActivityTaskHeartbeatResponse(response), thrift.ToError(err)
}

func (t thriftClient) RecordActivityTaskHeartbeatByID(ctx context.Context, request *types.RecordActivityTaskHeartbeatByIDRequest, opts ...yarpc.CallOption) (*types.RecordActivityTaskHeartbeatResponse, error) {
	response, err := t.c.RecordActivityTaskHeartbeatByID(ctx, thrift.FromRecordActivityTaskHeartbeatByIDRequest(request), opts...)
	return thrift.ToRecordActivityTaskHeartbeatResponse(response), thrift.ToError(err)
}

func (t thriftClient) RegisterDomain(ctx context.Context, request *types.RegisterDomainRequest, opts ...yarpc.CallOption) error {
	err := t.c.RegisterDomain(ctx, thrift.FromRegisterDomainRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) RequestCancelWorkflowExecution(ctx context.Context, request *types.RequestCancelWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	err := t.c.RequestCancelWorkflowExecution(ctx, thrift.FromRequestCancelWorkflowExecutionRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) ResetStickyTaskList(ctx context.Context, request *types.ResetStickyTaskListRequest, opts ...yarpc.CallOption) (*types.ResetStickyTaskListResponse, error) {
	response, err := t.c.ResetStickyTaskList(ctx, thrift.FromResetStickyTaskListRequest(request), opts...)
	return thrift.ToResetStickyTaskListResponse(response), thrift.ToError(err)
}

func (t thriftClient) ResetWorkflowExecution(ctx context.Context, request *types.ResetWorkflowExecutionRequest, opts ...yarpc.CallOption) (*types.ResetWorkflowExecutionResponse, error) {
	response, err := t.c.ResetWorkflowExecution(ctx, thrift.FromResetWorkflowExecutionRequest(request), opts...)
	return thrift.ToResetWorkflowExecutionResponse(response), thrift.ToError(err)
}

func (t thriftClient) RespondActivityTaskCanceled(ctx context.Context, request *types.RespondActivityTaskCanceledRequest, opts ...yarpc.CallOption) error {
	err := t.c.RespondActivityTaskCanceled(ctx, thrift.FromRespondActivityTaskCanceledRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) RespondActivityTaskCanceledByID(ctx context.Context, request *types.RespondActivityTaskCanceledByIDRequest, opts ...yarpc.CallOption) error {
	err := t.c.RespondActivityTaskCanceledByID(ctx, thrift.FromRespondActivityTaskCanceledByIDRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) RespondActivityTaskCompleted(ctx context.Context, request *types.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption) error {
	err := t.c.RespondActivityTaskCompleted(ctx, thrift.FromRespondActivityTaskCompletedRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) RespondActivityTaskCompletedByID(ctx context.Context, request *types.RespondActivityTaskCompletedByIDRequest, opts ...yarpc.CallOption) error {
	err := t.c.RespondActivityTaskCompletedByID(ctx, thrift.FromRespondActivityTaskCompletedByIDRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) RespondActivityTaskFailed(ctx context.Context, request *types.RespondActivityTaskFailedRequest, opts ...yarpc.CallOption) error {
	err := t.c.RespondActivityTaskFailed(ctx, thrift.FromRespondActivityTaskFailedRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) RespondActivityTaskFailedByID(ctx context.Context, request *types.RespondActivityTaskFailedByIDRequest, opts ...yarpc.CallOption) error {
	err := t.c.RespondActivityTaskFailedByID(ctx, thrift.FromRespondActivityTaskFailedByIDRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) RespondDecisionTaskCompleted(ctx context.Context, request *types.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption) (*types.RespondDecisionTaskCompletedResponse, error) {
	response, err := t.c.RespondDecisionTaskCompleted(ctx, thrift.FromRespondDecisionTaskCompletedRequest(request), opts...)
	return thrift.ToRespondDecisionTaskCompletedResponse(response), thrift.ToError(err)
}

func (t thriftClient) RespondDecisionTaskFailed(ctx context.Context, request *types.RespondDecisionTaskFailedRequest, opts ...yarpc.CallOption) error {
	err := t.c.RespondDecisionTaskFailed(ctx, thrift.FromRespondDecisionTaskFailedRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) RespondQueryTaskCompleted(ctx context.Context, request *types.RespondQueryTaskCompletedRequest, opts ...yarpc.CallOption) error {
	err := t.c.RespondQueryTaskCompleted(ctx, thrift.FromRespondQueryTaskCompletedRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) ScanWorkflowExecutions(ctx context.Context, request *types.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*types.ListWorkflowExecutionsResponse, error) {
	response, err := t.c.ScanWorkflowExecutions(ctx, thrift.FromListWorkflowExecutionsRequest(request), opts...)
	return thrift.ToListWorkflowExecutionsResponse(response), thrift.ToError(err)
}

func (t thriftClient) SignalWithStartWorkflowExecution(ctx context.Context, request *types.SignalWithStartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*types.StartWorkflowExecutionResponse, error) {
	response, err := t.c.SignalWithStartWorkflowExecution(ctx, thrift.FromSignalWithStartWorkflowExecutionRequest(request), opts...)
	return thrift.ToStartWorkflowExecutionResponse(response), thrift.ToError(err)
}

func (t thriftClient) SignalWorkflowExecution(ctx context.Context, request *types.SignalWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	err := t.c.SignalWorkflowExecution(ctx, thrift.FromSignalWorkflowExecutionRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) StartWorkflowExecution(ctx context.Context, request *types.StartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*types.StartWorkflowExecutionResponse, error) {
	response, err := t.c.StartWorkflowExecution(ctx, thrift.FromStartWorkflowExecutionRequest(request), opts...)
	return thrift.ToStartWorkflowExecutionResponse(response), thrift.ToError(err)
}

func (t thriftClient) TerminateWorkflowExecution(ctx context.Context, request *types.TerminateWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	err := t.c.TerminateWorkflowExecution(ctx, thrift.FromTerminateWorkflowExecutionRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) UpdateDomain(ctx context.Context, request *types.UpdateDomainRequest, opts ...yarpc.CallOption) (*types.UpdateDomainResponse, error) {
	response, err := t.c.UpdateDomain(ctx, thrift.FromUpdateDomainRequest(request), opts...)
	return thrift.ToUpdateDomainResponse(response), thrift.ToError(err)
}
