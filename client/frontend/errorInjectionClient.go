// Copyright (c) 2017-2020 Uber Technologies, Inc.
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

	"github.com/uber/cadence/common/errors"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/types"
)

var _ Client = (*errorInjectionClient)(nil)

const (
	msgInjectedFakeErr = "Injected fake frontend client error"
)

type errorInjectionClient struct {
	client    Client
	errorRate float64
	logger    log.Logger
}

// NewErrorInjectionClient creates a new instance of Client that injects fake error
func NewErrorInjectionClient(
	client Client,
	errorRate float64,
	logger log.Logger,
) Client {
	return &errorInjectionClient{
		client:    client,
		errorRate: errorRate,
		logger:    logger,
	}
}

func (c *errorInjectionClient) DeprecateDomain(
	ctx context.Context,
	request *types.DeprecateDomainRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.DeprecateDomain(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationDeprecateDomain,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) DescribeDomain(
	ctx context.Context,
	request *types.DescribeDomainRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeDomainResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.DescribeDomainResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.DescribeDomain(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationDescribeDomain,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) DescribeTaskList(
	ctx context.Context,
	request *types.DescribeTaskListRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeTaskListResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.DescribeTaskListResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.DescribeTaskList(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationDescribeTaskList,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) DescribeWorkflowExecution(
	ctx context.Context,
	request *types.DescribeWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeWorkflowExecutionResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.DescribeWorkflowExecutionResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.DescribeWorkflowExecution(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationDescribeWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) GetWorkflowExecutionHistory(
	ctx context.Context,
	request *types.GetWorkflowExecutionHistoryRequest,
	opts ...yarpc.CallOption,
) (*types.GetWorkflowExecutionHistoryResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.GetWorkflowExecutionHistoryResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.GetWorkflowExecutionHistory(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationDescribeWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) ListArchivedWorkflowExecutions(
	ctx context.Context,
	request *types.ListArchivedWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListArchivedWorkflowExecutionsResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.ListArchivedWorkflowExecutionsResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.ListArchivedWorkflowExecutions(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationListArchivedWorkflowExecutions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *types.ListClosedWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListClosedWorkflowExecutionsResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.ListClosedWorkflowExecutionsResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.ListClosedWorkflowExecutions(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationListClosedWorkflowExecutions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) ListDomains(
	ctx context.Context,
	request *types.ListDomainsRequest,
	opts ...yarpc.CallOption,
) (*types.ListDomainsResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.ListDomainsResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.ListDomains(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationListClosedWorkflowExecutions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *types.ListOpenWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListOpenWorkflowExecutionsResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.ListOpenWorkflowExecutionsResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.ListOpenWorkflowExecutions(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationListOpenWorkflowExecutions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) ListWorkflowExecutions(
	ctx context.Context,
	request *types.ListWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListWorkflowExecutionsResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.ListWorkflowExecutionsResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.ListWorkflowExecutions(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationListWorkflowExecutions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) ScanWorkflowExecutions(
	ctx context.Context,
	request *types.ListWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListWorkflowExecutionsResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.ListWorkflowExecutionsResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.ScanWorkflowExecutions(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationScanWorkflowExecutions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) CountWorkflowExecutions(
	ctx context.Context,
	request *types.CountWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.CountWorkflowExecutionsResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.CountWorkflowExecutionsResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.CountWorkflowExecutions(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationCountWorkflowExecutions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) GetSearchAttributes(
	ctx context.Context,
	opts ...yarpc.CallOption,
) (*types.GetSearchAttributesResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.GetSearchAttributesResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.GetSearchAttributes(ctx, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationGetSearchAttributes,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) PollForActivityTask(
	ctx context.Context,
	request *types.PollForActivityTaskRequest,
	opts ...yarpc.CallOption,
) (*types.PollForActivityTaskResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.PollForActivityTaskResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.PollForActivityTask(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationPollForActivityTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) PollForDecisionTask(
	ctx context.Context,
	request *types.PollForDecisionTaskRequest,
	opts ...yarpc.CallOption,
) (*types.PollForDecisionTaskResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.PollForDecisionTaskResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.PollForDecisionTask(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationPollForDecisionTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) QueryWorkflow(
	ctx context.Context,
	request *types.QueryWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.QueryWorkflowResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.QueryWorkflowResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.QueryWorkflow(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationQueryWorkflow,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) RecordActivityTaskHeartbeat(
	ctx context.Context,
	request *types.RecordActivityTaskHeartbeatRequest,
	opts ...yarpc.CallOption,
) (*types.RecordActivityTaskHeartbeatResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.RecordActivityTaskHeartbeatResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.RecordActivityTaskHeartbeat(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationRecordActivityTaskHeartbeat,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) RecordActivityTaskHeartbeatByID(
	ctx context.Context,
	request *types.RecordActivityTaskHeartbeatByIDRequest,
	opts ...yarpc.CallOption,
) (*types.RecordActivityTaskHeartbeatResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.RecordActivityTaskHeartbeatResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.RecordActivityTaskHeartbeatByID(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationRecordActivityTaskHeartbeatByID,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) RegisterDomain(
	ctx context.Context,
	request *types.RegisterDomainRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.RegisterDomain(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationRegisterDomain,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) RequestCancelWorkflowExecution(
	ctx context.Context,
	request *types.RequestCancelWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.RequestCancelWorkflowExecution(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationRequestCancelWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) ResetStickyTaskList(
	ctx context.Context,
	request *types.ResetStickyTaskListRequest,
	opts ...yarpc.CallOption,
) (*types.ResetStickyTaskListResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.ResetStickyTaskListResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.ResetStickyTaskList(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationResetStickyTaskList,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) ResetWorkflowExecution(
	ctx context.Context,
	request *types.ResetWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.ResetWorkflowExecutionResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.ResetWorkflowExecutionResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.ResetWorkflowExecution(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationResetWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) RespondActivityTaskCanceled(
	ctx context.Context,
	request *types.RespondActivityTaskCanceledRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.RespondActivityTaskCanceled(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationRespondActivityTaskCanceled,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) RespondActivityTaskCanceledByID(
	ctx context.Context,
	request *types.RespondActivityTaskCanceledByIDRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.RespondActivityTaskCanceledByID(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationRespondActivityTaskCanceledByID,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) RespondActivityTaskCompleted(
	ctx context.Context,
	request *types.RespondActivityTaskCompletedRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.RespondActivityTaskCompleted(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationRespondActivityTaskCompleted,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) RespondActivityTaskCompletedByID(
	ctx context.Context,
	request *types.RespondActivityTaskCompletedByIDRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.RespondActivityTaskCompletedByID(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationRespondActivityTaskCompletedByID,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) RespondActivityTaskFailed(
	ctx context.Context,
	request *types.RespondActivityTaskFailedRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.RespondActivityTaskFailed(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationRespondActivityTaskFailed,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) RespondActivityTaskFailedByID(
	ctx context.Context,
	request *types.RespondActivityTaskFailedByIDRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.RespondActivityTaskFailedByID(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationRespondActivityTaskFailedByID,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) RespondDecisionTaskCompleted(
	ctx context.Context,
	request *types.RespondDecisionTaskCompletedRequest,
	opts ...yarpc.CallOption,
) (*types.RespondDecisionTaskCompletedResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.RespondDecisionTaskCompletedResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.RespondDecisionTaskCompleted(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationRespondDecisionTaskCompleted,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) RespondDecisionTaskFailed(
	ctx context.Context,
	request *types.RespondDecisionTaskFailedRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.RespondDecisionTaskFailed(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationRespondDecisionTaskFailed,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) RespondQueryTaskCompleted(
	ctx context.Context,
	request *types.RespondQueryTaskCompletedRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.RespondQueryTaskCompleted(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationRespondQueryTaskCompleted,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) SignalWithStartWorkflowExecution(
	ctx context.Context,
	request *types.SignalWithStartWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.StartWorkflowExecutionResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.StartWorkflowExecutionResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.SignalWithStartWorkflowExecution(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationSignalWithStartWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) SignalWorkflowExecution(
	ctx context.Context,
	request *types.SignalWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.SignalWorkflowExecution(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationSignalWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) StartWorkflowExecution(
	ctx context.Context,
	request *types.StartWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.StartWorkflowExecutionResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.StartWorkflowExecutionResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.StartWorkflowExecution(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationStartWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) TerminateWorkflowExecution(
	ctx context.Context,
	request *types.TerminateWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.TerminateWorkflowExecution(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationTerminateWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) UpdateDomain(
	ctx context.Context,
	request *types.UpdateDomainRequest,
	opts ...yarpc.CallOption,
) (*types.UpdateDomainResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.UpdateDomainResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.UpdateDomain(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationUpdateDomain,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) GetClusterInfo(
	ctx context.Context,
	opts ...yarpc.CallOption,
) (*types.ClusterInfo, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.ClusterInfo
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.GetClusterInfo(ctx, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationGetClusterInfo,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) ListTaskListPartitions(
	ctx context.Context,
	request *types.ListTaskListPartitionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListTaskListPartitionsResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.ListTaskListPartitionsResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.ListTaskListPartitions(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationListTaskListPartitions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) GetTaskListsByDomain(
	ctx context.Context,
	request *types.GetTaskListsByDomainRequest,
	opts ...yarpc.CallOption,
) (*types.GetTaskListsByDomainResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.GetTaskListsByDomainResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.GetTaskListsByDomain(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.FrontendClientOperationGetTaskListsByDomain,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}
