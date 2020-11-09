// Copyright (c) 2017 Uber Technologies, Inc.
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

	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/types"
)

var _ Client = (*retryableClient)(nil)

type retryableClient struct {
	client      Client
	policy      backoff.RetryPolicy
	isRetryable backoff.IsRetryable
}

// NewRetryableClient creates a new instance of Client with retry policy
func NewRetryableClient(
	client Client,
	policy backoff.RetryPolicy,
	isRetryable backoff.IsRetryable,
) Client {
	return &retryableClient{
		client:      client,
		policy:      policy,
		isRetryable: isRetryable,
	}
}

func (c *retryableClient) DeprecateDomain(
	ctx context.Context,
	request *types.DeprecateDomainRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.DeprecateDomain(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) DescribeDomain(
	ctx context.Context,
	request *types.DescribeDomainRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeDomainResponse, error) {

	var resp *types.DescribeDomainResponse
	op := func() error {
		var err error
		resp, err = c.client.DescribeDomain(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) DescribeTaskList(
	ctx context.Context,
	request *types.DescribeTaskListRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeTaskListResponse, error) {

	var resp *types.DescribeTaskListResponse
	op := func() error {
		var err error
		resp, err = c.client.DescribeTaskList(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) DescribeWorkflowExecution(
	ctx context.Context,
	request *types.DescribeWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeWorkflowExecutionResponse, error) {

	var resp *types.DescribeWorkflowExecutionResponse
	op := func() error {
		var err error
		resp, err = c.client.DescribeWorkflowExecution(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) GetWorkflowExecutionHistory(
	ctx context.Context,
	request *types.GetWorkflowExecutionHistoryRequest,
	opts ...yarpc.CallOption,
) (*types.GetWorkflowExecutionHistoryResponse, error) {

	var resp *types.GetWorkflowExecutionHistoryResponse
	op := func() error {
		var err error
		resp, err = c.client.GetWorkflowExecutionHistory(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) ListArchivedWorkflowExecutions(
	ctx context.Context,
	request *types.ListArchivedWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListArchivedWorkflowExecutionsResponse, error) {

	var resp *types.ListArchivedWorkflowExecutionsResponse
	op := func() error {
		var err error
		resp, err = c.client.ListArchivedWorkflowExecutions(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *types.ListClosedWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListClosedWorkflowExecutionsResponse, error) {

	var resp *types.ListClosedWorkflowExecutionsResponse
	op := func() error {
		var err error
		resp, err = c.client.ListClosedWorkflowExecutions(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) ListDomains(
	ctx context.Context,
	request *types.ListDomainsRequest,
	opts ...yarpc.CallOption,
) (*types.ListDomainsResponse, error) {

	var resp *types.ListDomainsResponse
	op := func() error {
		var err error
		resp, err = c.client.ListDomains(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *types.ListOpenWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListOpenWorkflowExecutionsResponse, error) {

	var resp *types.ListOpenWorkflowExecutionsResponse
	op := func() error {
		var err error
		resp, err = c.client.ListOpenWorkflowExecutions(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) ListWorkflowExecutions(
	ctx context.Context,
	request *types.ListWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListWorkflowExecutionsResponse, error) {

	var resp *types.ListWorkflowExecutionsResponse
	op := func() error {
		var err error
		resp, err = c.client.ListWorkflowExecutions(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) ScanWorkflowExecutions(
	ctx context.Context,
	request *types.ListWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListWorkflowExecutionsResponse, error) {

	var resp *types.ListWorkflowExecutionsResponse
	op := func() error {
		var err error
		resp, err = c.client.ScanWorkflowExecutions(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) CountWorkflowExecutions(
	ctx context.Context,
	request *types.CountWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.CountWorkflowExecutionsResponse, error) {

	var resp *types.CountWorkflowExecutionsResponse
	op := func() error {
		var err error
		resp, err = c.client.CountWorkflowExecutions(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) GetSearchAttributes(
	ctx context.Context,
	opts ...yarpc.CallOption,
) (*types.GetSearchAttributesResponse, error) {

	var resp *types.GetSearchAttributesResponse
	op := func() error {
		var err error
		resp, err = c.client.GetSearchAttributes(ctx, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) PollForActivityTask(
	ctx context.Context,
	request *types.PollForActivityTaskRequest,
	opts ...yarpc.CallOption,
) (*types.PollForActivityTaskResponse, error) {

	var resp *types.PollForActivityTaskResponse
	op := func() error {
		var err error
		resp, err = c.client.PollForActivityTask(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) PollForDecisionTask(
	ctx context.Context,
	request *types.PollForDecisionTaskRequest,
	opts ...yarpc.CallOption,
) (*types.PollForDecisionTaskResponse, error) {

	var resp *types.PollForDecisionTaskResponse
	op := func() error {
		var err error
		resp, err = c.client.PollForDecisionTask(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) QueryWorkflow(
	ctx context.Context,
	request *types.QueryWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.QueryWorkflowResponse, error) {

	var resp *types.QueryWorkflowResponse
	op := func() error {
		var err error
		resp, err = c.client.QueryWorkflow(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) RecordActivityTaskHeartbeat(
	ctx context.Context,
	request *types.RecordActivityTaskHeartbeatRequest,
	opts ...yarpc.CallOption,
) (*types.RecordActivityTaskHeartbeatResponse, error) {

	var resp *types.RecordActivityTaskHeartbeatResponse
	op := func() error {
		var err error
		resp, err = c.client.RecordActivityTaskHeartbeat(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) RecordActivityTaskHeartbeatByID(
	ctx context.Context,
	request *types.RecordActivityTaskHeartbeatByIDRequest,
	opts ...yarpc.CallOption,
) (*types.RecordActivityTaskHeartbeatResponse, error) {

	var resp *types.RecordActivityTaskHeartbeatResponse
	op := func() error {
		var err error
		resp, err = c.client.RecordActivityTaskHeartbeatByID(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) RegisterDomain(
	ctx context.Context,
	request *types.RegisterDomainRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RegisterDomain(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) RequestCancelWorkflowExecution(
	ctx context.Context,
	request *types.RequestCancelWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RequestCancelWorkflowExecution(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) ResetStickyTaskList(
	ctx context.Context,
	request *types.ResetStickyTaskListRequest,
	opts ...yarpc.CallOption,
) (*types.ResetStickyTaskListResponse, error) {

	var resp *types.ResetStickyTaskListResponse
	op := func() error {
		var err error
		resp, err = c.client.ResetStickyTaskList(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) ResetWorkflowExecution(
	ctx context.Context,
	request *types.ResetWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.ResetWorkflowExecutionResponse, error) {

	var resp *types.ResetWorkflowExecutionResponse
	op := func() error {
		var err error
		resp, err = c.client.ResetWorkflowExecution(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) RespondActivityTaskCanceled(
	ctx context.Context,
	request *types.RespondActivityTaskCanceledRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RespondActivityTaskCanceled(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) RespondActivityTaskCanceledByID(
	ctx context.Context,
	request *types.RespondActivityTaskCanceledByIDRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RespondActivityTaskCanceledByID(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) RespondActivityTaskCompleted(
	ctx context.Context,
	request *types.RespondActivityTaskCompletedRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RespondActivityTaskCompleted(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) RespondActivityTaskCompletedByID(
	ctx context.Context,
	request *types.RespondActivityTaskCompletedByIDRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RespondActivityTaskCompletedByID(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) RespondActivityTaskFailed(
	ctx context.Context,
	request *types.RespondActivityTaskFailedRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RespondActivityTaskFailed(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) RespondActivityTaskFailedByID(
	ctx context.Context,
	request *types.RespondActivityTaskFailedByIDRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RespondActivityTaskFailedByID(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) RespondDecisionTaskCompleted(
	ctx context.Context,
	request *types.RespondDecisionTaskCompletedRequest,
	opts ...yarpc.CallOption,
) (*types.RespondDecisionTaskCompletedResponse, error) {

	var resp *types.RespondDecisionTaskCompletedResponse
	op := func() error {
		var err error
		resp, err = c.client.RespondDecisionTaskCompleted(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) RespondDecisionTaskFailed(
	ctx context.Context,
	request *types.RespondDecisionTaskFailedRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RespondDecisionTaskFailed(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) RespondQueryTaskCompleted(
	ctx context.Context,
	request *types.RespondQueryTaskCompletedRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RespondQueryTaskCompleted(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) SignalWithStartWorkflowExecution(
	ctx context.Context,
	request *types.SignalWithStartWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.StartWorkflowExecutionResponse, error) {

	var resp *types.StartWorkflowExecutionResponse
	op := func() error {
		var err error
		resp, err = c.client.SignalWithStartWorkflowExecution(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) SignalWorkflowExecution(
	ctx context.Context,
	request *types.SignalWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.SignalWorkflowExecution(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) StartWorkflowExecution(
	ctx context.Context,
	request *types.StartWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.StartWorkflowExecutionResponse, error) {

	var resp *types.StartWorkflowExecutionResponse
	op := func() error {
		var err error
		resp, err = c.client.StartWorkflowExecution(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) TerminateWorkflowExecution(
	ctx context.Context,
	request *types.TerminateWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.TerminateWorkflowExecution(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) UpdateDomain(
	ctx context.Context,
	request *types.UpdateDomainRequest,
	opts ...yarpc.CallOption,
) (*types.UpdateDomainResponse, error) {

	var resp *types.UpdateDomainResponse
	op := func() error {
		var err error
		resp, err = c.client.UpdateDomain(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) GetClusterInfo(
	ctx context.Context,
	opts ...yarpc.CallOption,
) (*types.ClusterInfo, error) {
	var resp *types.ClusterInfo
	op := func() error {
		var err error
		resp, err = c.client.GetClusterInfo(ctx, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) ListTaskListPartitions(
	ctx context.Context,
	request *types.ListTaskListPartitionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListTaskListPartitionsResponse, error) {
	var resp *types.ListTaskListPartitionsResponse
	op := func() error {
		var err error
		resp, err = c.client.ListTaskListPartitions(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}
