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

package admin

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
func NewRetryableClient(client Client, policy backoff.RetryPolicy, isRetryable backoff.IsRetryable) Client {
	return &retryableClient{
		client: client,
		throttleRetry: backoff.NewThrottleRetry(
			backoff.WithRetryPolicy(policy),
			backoff.WithRetryableError(isRetryable),
		),
	}
}

func (c *retryableClient) AddSearchAttribute(
	ctx context.Context,
	request *types.AddSearchAttributeRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.AddSearchAttribute(ctx, request, opts...)
	}
	return c.throttleRetry.Do(ctx, op)
}

func (c *retryableClient) DescribeShardDistribution(
	ctx context.Context,
	request *types.DescribeShardDistributionRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeShardDistributionResponse, error) {

	var resp *types.DescribeShardDistributionResponse
	op := func() error {
		var err error
		resp, err = c.client.DescribeShardDistribution(ctx, request, opts...)
		return err
	}
	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) DescribeHistoryHost(
	ctx context.Context,
	request *types.DescribeHistoryHostRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeHistoryHostResponse, error) {

	var resp *types.DescribeHistoryHostResponse
	op := func() error {
		var err error
		resp, err = c.client.DescribeHistoryHost(ctx, request, opts...)
		return err
	}
	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) RemoveTask(
	ctx context.Context,
	request *types.RemoveTaskRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RemoveTask(ctx, request, opts...)
	}
	return c.throttleRetry.Do(ctx, op)
}

func (c *retryableClient) CloseShard(
	ctx context.Context,
	request *types.CloseShardRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.CloseShard(ctx, request, opts...)
	}
	return c.throttleRetry.Do(ctx, op)
}

func (c *retryableClient) ResetQueue(
	ctx context.Context,
	request *types.ResetQueueRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.ResetQueue(ctx, request, opts...)
	}
	return c.throttleRetry.Do(ctx, op)
}

func (c *retryableClient) DescribeQueue(
	ctx context.Context,
	request *types.DescribeQueueRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeQueueResponse, error) {

	var resp *types.DescribeQueueResponse
	op := func() error {
		var err error
		resp, err = c.client.DescribeQueue(ctx, request, opts...)
		return err
	}
	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) DescribeWorkflowExecution(
	ctx context.Context,
	request *types.AdminDescribeWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.AdminDescribeWorkflowExecutionResponse, error) {

	var resp *types.AdminDescribeWorkflowExecutionResponse
	op := func() error {
		var err error
		resp, err = c.client.DescribeWorkflowExecution(ctx, request, opts...)
		return err
	}
	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) GetWorkflowExecutionRawHistoryV2(
	ctx context.Context,
	request *types.GetWorkflowExecutionRawHistoryV2Request,
	opts ...yarpc.CallOption,
) (*types.GetWorkflowExecutionRawHistoryV2Response, error) {

	var resp *types.GetWorkflowExecutionRawHistoryV2Response
	op := func() error {
		var err error
		resp, err = c.client.GetWorkflowExecutionRawHistoryV2(ctx, request, opts...)
		return err
	}
	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) DescribeCluster(
	ctx context.Context,
	opts ...yarpc.CallOption,
) (*types.DescribeClusterResponse, error) {

	var resp *types.DescribeClusterResponse
	op := func() error {
		var err error
		resp, err = c.client.DescribeCluster(ctx, opts...)
		return err
	}
	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) GetReplicationMessages(
	ctx context.Context,
	request *types.GetReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.GetReplicationMessagesResponse, error) {
	var resp *types.GetReplicationMessagesResponse
	op := func() error {
		var err error
		resp, err = c.client.GetReplicationMessages(ctx, request, opts...)
		return err
	}
	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) GetDomainReplicationMessages(
	ctx context.Context,
	request *types.GetDomainReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.GetDomainReplicationMessagesResponse, error) {
	var resp *types.GetDomainReplicationMessagesResponse
	op := func() error {
		var err error
		resp, err = c.client.GetDomainReplicationMessages(ctx, request, opts...)
		return err
	}
	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) GetDLQReplicationMessages(
	ctx context.Context,
	request *types.GetDLQReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.GetDLQReplicationMessagesResponse, error) {
	var resp *types.GetDLQReplicationMessagesResponse
	op := func() error {
		var err error
		resp, err = c.client.GetDLQReplicationMessages(ctx, request, opts...)
		return err
	}
	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) ReapplyEvents(
	ctx context.Context,
	request *types.ReapplyEventsRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.ReapplyEvents(ctx, request, opts...)
	}
	return c.throttleRetry.Do(ctx, op)
}

func (c *retryableClient) CountDLQMessages(
	ctx context.Context,
	request *types.CountDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.CountDLQMessagesResponse, error) {

	var resp *types.CountDLQMessagesResponse
	op := func() error {
		var err error
		resp, err = c.client.CountDLQMessages(ctx, request, opts...)
		return err
	}
	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) ReadDLQMessages(
	ctx context.Context,
	request *types.ReadDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.ReadDLQMessagesResponse, error) {

	var resp *types.ReadDLQMessagesResponse
	op := func() error {
		var err error
		resp, err = c.client.ReadDLQMessages(ctx, request, opts...)
		return err
	}
	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) PurgeDLQMessages(
	ctx context.Context,
	request *types.PurgeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.PurgeDLQMessages(ctx, request, opts...)
	}
	return c.throttleRetry.Do(ctx, op)
}

func (c *retryableClient) MergeDLQMessages(
	ctx context.Context,
	request *types.MergeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.MergeDLQMessagesResponse, error) {

	var resp *types.MergeDLQMessagesResponse
	op := func() error {
		var err error
		resp, err = c.client.MergeDLQMessages(ctx, request, opts...)
		return err
	}
	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) RefreshWorkflowTasks(
	ctx context.Context,
	request *types.RefreshWorkflowTasksRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RefreshWorkflowTasks(ctx, request, opts...)
	}
	return c.throttleRetry.Do(ctx, op)
}

func (c *retryableClient) ResendReplicationTasks(
	ctx context.Context,
	request *types.ResendReplicationTasksRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.ResendReplicationTasks(ctx, request, opts...)
	}
	return c.throttleRetry.Do(ctx, op)
}

func (c *retryableClient) GetCrossClusterTasks(
	ctx context.Context,
	request *types.GetCrossClusterTasksRequest,
	opts ...yarpc.CallOption,
) (*types.GetCrossClusterTasksResponse, error) {
	var resp *types.GetCrossClusterTasksResponse
	op := func() error {
		var err error
		resp, err = c.client.GetCrossClusterTasks(ctx, request, opts...)
		return err
	}
	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) RespondCrossClusterTasksCompleted(
	ctx context.Context,
	request *types.RespondCrossClusterTasksCompletedRequest,
	opts ...yarpc.CallOption,
) (*types.RespondCrossClusterTasksCompletedResponse, error) {
	var resp *types.RespondCrossClusterTasksCompletedResponse
	op := func() error {
		var err error
		resp, err = c.client.RespondCrossClusterTasksCompleted(ctx, request, opts...)
		return err
	}

	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) GetDynamicConfig(
	ctx context.Context,
	request *types.GetDynamicConfigRequest,
	opts ...yarpc.CallOption,
) (*types.GetDynamicConfigResponse, error) {
	var resp *types.GetDynamicConfigResponse
	op := func() error {
		var err error
		resp, err = c.client.GetDynamicConfig(ctx, request, opts...)
		return err
	}
	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) UpdateDynamicConfig(
	ctx context.Context,
	request *types.UpdateDynamicConfigRequest,
	opts ...yarpc.CallOption,
) error {
	op := func() error {
		return c.client.UpdateDynamicConfig(ctx, request, opts...)
	}
	return c.throttleRetry.Do(ctx, op)
}

func (c *retryableClient) RestoreDynamicConfig(
	ctx context.Context,
	request *types.RestoreDynamicConfigRequest,
	opts ...yarpc.CallOption,
) error {
	op := func() error {
		return c.client.RestoreDynamicConfig(ctx, request, opts...)
	}
	return c.throttleRetry.Do(ctx, op)
}

func (c *retryableClient) DeleteWorkflow(
	ctx context.Context,
	request *types.AdminDeleteWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.AdminDeleteWorkflowResponse, error) {
	var resp *types.AdminDeleteWorkflowResponse
	op := func() error {
		var err error
		resp, err = c.client.DeleteWorkflow(ctx, request, opts...)
		return err
	}
	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) MaintainCorruptWorkflow(
	ctx context.Context,
	request *types.AdminMaintainWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.AdminMaintainWorkflowResponse, error) {
	var resp *types.AdminMaintainWorkflowResponse
	op := func() error {
		var err error
		resp, err = c.client.MaintainCorruptWorkflow(ctx, request, opts...)
		return err
	}
	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}

func (c *retryableClient) ListDynamicConfig(
	ctx context.Context,
	request *types.ListDynamicConfigRequest,
	opts ...yarpc.CallOption,
) (*types.ListDynamicConfigResponse, error) {
	var resp *types.ListDynamicConfigResponse
	op := func() error {
		var err error
		resp, err = c.client.ListDynamicConfig(ctx, request, opts...)
		return err
	}
	err := c.throttleRetry.Do(ctx, op)
	return resp, err
}
