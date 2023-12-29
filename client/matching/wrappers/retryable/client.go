// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package retryable

// Code generated by gowrap. DO NOT EDIT.
// template: ../../../templates/retry.tmpl
// gowrap: http://github.com/hexdigest/gowrap

import (
	"context"

	"go.uber.org/yarpc"

	"github.com/uber/cadence/client/matching"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/types"
)

// retryableClient implements matching.Client interface instrumented with retries
type retryableClient struct {
	client        matching.Client
	throttleRetry *backoff.ThrottleRetry
}

// NewRetryableClient creates a new instance of retryableClient with retry policy
func NewRetryableClient(
	client matching.Client,
	policy backoff.RetryPolicy,
	isRetryable backoff.IsRetryable,
) matching.Client {
	return &retryableClient{
		client: client,
		throttleRetry: backoff.NewThrottleRetry(
			backoff.WithRetryPolicy(policy),
			backoff.WithRetryableError(isRetryable),
		),
	}
}

func (c *retryableClient) AddActivityTask(ctx context.Context, ap1 *types.AddActivityTaskRequest, p1 ...yarpc.CallOption) (err error) {

	op := func() error {
		return c.client.AddActivityTask(ctx, ap1, p1...)
	}
	return c.throttleRetry.Do(ctx, op)
}

func (c *retryableClient) AddDecisionTask(ctx context.Context, ap1 *types.AddDecisionTaskRequest, p1 ...yarpc.CallOption) (err error) {

	op := func() error {
		return c.client.AddDecisionTask(ctx, ap1, p1...)
	}
	return c.throttleRetry.Do(ctx, op)
}

func (c *retryableClient) CancelOutstandingPoll(ctx context.Context, cp1 *types.CancelOutstandingPollRequest, p1 ...yarpc.CallOption) (err error) {

	op := func() error {
		return c.client.CancelOutstandingPoll(ctx, cp1, p1...)
	}
	return c.throttleRetry.Do(ctx, op)
}

func (c *retryableClient) DescribeTaskList(ctx context.Context, mp1 *types.MatchingDescribeTaskListRequest, p1 ...yarpc.CallOption) (dp1 *types.DescribeTaskListResponse, err error) {

	var resp *types.DescribeTaskListResponse
	op := func() error {
		var err error
		resp, err = c.client.DescribeTaskList(ctx, mp1, p1...)
		return err
	}
	err = c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) GetTaskListsByDomain(ctx context.Context, gp1 *types.GetTaskListsByDomainRequest, p1 ...yarpc.CallOption) (gp2 *types.GetTaskListsByDomainResponse, err error) {

	var resp *types.GetTaskListsByDomainResponse
	op := func() error {
		var err error
		resp, err = c.client.GetTaskListsByDomain(ctx, gp1, p1...)
		return err
	}
	err = c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) ListTaskListPartitions(ctx context.Context, mp1 *types.MatchingListTaskListPartitionsRequest, p1 ...yarpc.CallOption) (lp1 *types.ListTaskListPartitionsResponse, err error) {

	var resp *types.ListTaskListPartitionsResponse
	op := func() error {
		var err error
		resp, err = c.client.ListTaskListPartitions(ctx, mp1, p1...)
		return err
	}
	err = c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) PollForActivityTask(ctx context.Context, mp1 *types.MatchingPollForActivityTaskRequest, p1 ...yarpc.CallOption) (pp1 *types.PollForActivityTaskResponse, err error) {

	var resp *types.PollForActivityTaskResponse
	op := func() error {
		var err error
		resp, err = c.client.PollForActivityTask(ctx, mp1, p1...)
		return err
	}
	err = c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) PollForDecisionTask(ctx context.Context, mp1 *types.MatchingPollForDecisionTaskRequest, p1 ...yarpc.CallOption) (mp2 *types.MatchingPollForDecisionTaskResponse, err error) {

	var resp *types.MatchingPollForDecisionTaskResponse
	op := func() error {
		var err error
		resp, err = c.client.PollForDecisionTask(ctx, mp1, p1...)
		return err
	}
	err = c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) QueryWorkflow(ctx context.Context, mp1 *types.MatchingQueryWorkflowRequest, p1 ...yarpc.CallOption) (qp1 *types.QueryWorkflowResponse, err error) {

	var resp *types.QueryWorkflowResponse
	op := func() error {
		var err error
		resp, err = c.client.QueryWorkflow(ctx, mp1, p1...)
		return err
	}
	err = c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) RespondQueryTaskCompleted(ctx context.Context, mp1 *types.MatchingRespondQueryTaskCompletedRequest, p1 ...yarpc.CallOption) (err error) {

	op := func() error {
		return c.client.RespondQueryTaskCompleted(ctx, mp1, p1...)
	}
	return c.throttleRetry.Do(ctx, op)
}
