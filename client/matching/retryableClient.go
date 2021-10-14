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

package matching

import (
	"context"

	"go.uber.org/yarpc"

	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/types"
)

var _ Client = (*retryableClient)(nil)

type retryableClient struct {
	client        Client
	throttleRetry *backoff.ThrottleRetry
}

// NewRetryableClient creates a new instance of Client with retry policy
func NewRetryableClient(
	client Client,
	policy backoff.RetryPolicy,
	isRetryable backoff.IsRetryable,
) Client {
	return &retryableClient{
		client: client,
		throttleRetry: backoff.NewThrottleRetry(
			backoff.WithRetryPolicy(policy),
			backoff.WithRetryableError(isRetryable),
		),
	}
}

func (c *retryableClient) AddActivityTask(
	ctx context.Context,
	addRequest *types.AddActivityTaskRequest,
	opts ...yarpc.CallOption,
) error {
	op := func() error {
		return c.client.AddActivityTask(ctx, addRequest, opts...)
	}

	return c.throttleRetry.Do(ctx, op)
}

func (c *retryableClient) AddDecisionTask(
	ctx context.Context,
	addRequest *types.AddDecisionTaskRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.AddDecisionTask(ctx, addRequest, opts...)
	}

	return c.throttleRetry.Do(ctx, op)
}

func (c *retryableClient) PollForActivityTask(
	ctx context.Context,
	pollRequest *types.MatchingPollForActivityTaskRequest,
	opts ...yarpc.CallOption,
) (*types.PollForActivityTaskResponse, error) {

	var resp *types.PollForActivityTaskResponse
	op := func() error {
		var err error
		resp, err = c.client.PollForActivityTask(ctx, pollRequest, opts...)
		return err
	}

	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) PollForDecisionTask(
	ctx context.Context,
	pollRequest *types.MatchingPollForDecisionTaskRequest,
	opts ...yarpc.CallOption,
) (*types.MatchingPollForDecisionTaskResponse, error) {

	var resp *types.MatchingPollForDecisionTaskResponse
	op := func() error {
		var err error
		resp, err = c.client.PollForDecisionTask(ctx, pollRequest, opts...)
		return err
	}

	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) QueryWorkflow(
	ctx context.Context,
	queryRequest *types.MatchingQueryWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.QueryWorkflowResponse, error) {

	var resp *types.QueryWorkflowResponse
	op := func() error {
		var err error
		resp, err = c.client.QueryWorkflow(ctx, queryRequest, opts...)
		return err
	}

	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) RespondQueryTaskCompleted(
	ctx context.Context,
	request *types.MatchingRespondQueryTaskCompletedRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RespondQueryTaskCompleted(ctx, request, opts...)
	}

	return c.throttleRetry.Do(ctx, op)
}

func (c *retryableClient) CancelOutstandingPoll(
	ctx context.Context,
	request *types.CancelOutstandingPollRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.CancelOutstandingPoll(ctx, request, opts...)
	}

	return c.throttleRetry.Do(ctx, op)
}

func (c *retryableClient) DescribeTaskList(
	ctx context.Context,
	request *types.MatchingDescribeTaskListRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeTaskListResponse, error) {

	var resp *types.DescribeTaskListResponse
	op := func() error {
		var err error
		resp, err = c.client.DescribeTaskList(ctx, request, opts...)
		return err
	}

	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) ListTaskListPartitions(
	ctx context.Context,
	request *types.MatchingListTaskListPartitionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListTaskListPartitionsResponse, error) {

	var resp *types.ListTaskListPartitionsResponse
	op := func() error {
		var err error
		resp, err = c.client.ListTaskListPartitions(ctx, request, opts...)
		return err
	}

	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) GetTaskListsByDomain(
	ctx context.Context,
	request *types.GetTaskListsByDomainRequest,
	opts ...yarpc.CallOption,
) (*types.GetTaskListsByDomainResponse, error) {

	var resp *types.GetTaskListsByDomainResponse
	op := func() error {
		var err error
		resp, err = c.client.GetTaskListsByDomain(ctx, request, opts...)
		return err
	}

	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}
