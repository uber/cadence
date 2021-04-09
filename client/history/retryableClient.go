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

package history

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

func (c *retryableClient) StartWorkflowExecution(
	ctx context.Context,
	request *types.HistoryStartWorkflowExecutionRequest,
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

	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) CloseShard(
	ctx context.Context,
	request *types.CloseShardRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		err := c.client.CloseShard(ctx, request, opts...)
		return err
	}

	err := backoff.Retry(op, c.policy, c.isRetryable)
	return err
}

func (c *retryableClient) ResetQueue(
	ctx context.Context,
	request *types.ResetQueueRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		err := c.client.ResetQueue(ctx, request, opts...)
		return err
	}

	err := backoff.Retry(op, c.policy, c.isRetryable)
	return err
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

	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) RemoveTask(
	ctx context.Context,
	request *types.RemoveTaskRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		err := c.client.RemoveTask(ctx, request, opts...)
		return err
	}

	err := backoff.Retry(op, c.policy, c.isRetryable)
	return err
}

func (c *retryableClient) DescribeMutableState(
	ctx context.Context,
	request *types.DescribeMutableStateRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeMutableStateResponse, error) {

	var resp *types.DescribeMutableStateResponse
	op := func() error {
		var err error
		resp, err = c.client.DescribeMutableState(ctx, request, opts...)
		return err
	}

	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) GetMutableState(
	ctx context.Context,
	request *types.GetMutableStateRequest,
	opts ...yarpc.CallOption,
) (*types.GetMutableStateResponse, error) {

	var resp *types.GetMutableStateResponse
	op := func() error {
		var err error
		resp, err = c.client.GetMutableState(ctx, request, opts...)
		return err
	}

	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) PollMutableState(
	ctx context.Context,
	request *types.PollMutableStateRequest,
	opts ...yarpc.CallOption,
) (*types.PollMutableStateResponse, error) {

	var resp *types.PollMutableStateResponse
	op := func() error {
		var err error
		resp, err = c.client.PollMutableState(ctx, request, opts...)
		return err
	}

	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) ResetStickyTaskList(
	ctx context.Context,
	request *types.HistoryResetStickyTaskListRequest,
	opts ...yarpc.CallOption,
) (*types.HistoryResetStickyTaskListResponse, error) {

	var resp *types.HistoryResetStickyTaskListResponse
	op := func() error {
		var err error
		resp, err = c.client.ResetStickyTaskList(ctx, request, opts...)
		return err
	}

	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) DescribeWorkflowExecution(
	ctx context.Context,
	request *types.HistoryDescribeWorkflowExecutionRequest,
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

func (c *retryableClient) RecordDecisionTaskStarted(
	ctx context.Context,
	request *types.RecordDecisionTaskStartedRequest,
	opts ...yarpc.CallOption,
) (*types.RecordDecisionTaskStartedResponse, error) {

	var resp *types.RecordDecisionTaskStartedResponse
	op := func() error {
		var err error
		resp, err = c.client.RecordDecisionTaskStarted(ctx, request, opts...)
		return err
	}

	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) RecordActivityTaskStarted(
	ctx context.Context,
	request *types.RecordActivityTaskStartedRequest,
	opts ...yarpc.CallOption,
) (*types.RecordActivityTaskStartedResponse, error) {

	var resp *types.RecordActivityTaskStartedResponse
	op := func() error {
		var err error
		resp, err = c.client.RecordActivityTaskStarted(ctx, request, opts...)
		return err
	}

	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) RespondDecisionTaskCompleted(
	ctx context.Context,
	request *types.HistoryRespondDecisionTaskCompletedRequest,
	opts ...yarpc.CallOption,
) (*types.HistoryRespondDecisionTaskCompletedResponse, error) {

	var resp *types.HistoryRespondDecisionTaskCompletedResponse
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
	request *types.HistoryRespondDecisionTaskFailedRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RespondDecisionTaskFailed(ctx, request, opts...)
	}

	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) RespondActivityTaskCompleted(
	ctx context.Context,
	request *types.HistoryRespondActivityTaskCompletedRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RespondActivityTaskCompleted(ctx, request, opts...)
	}

	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) RespondActivityTaskFailed(
	ctx context.Context,
	request *types.HistoryRespondActivityTaskFailedRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RespondActivityTaskFailed(ctx, request, opts...)
	}

	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) RespondActivityTaskCanceled(
	ctx context.Context,
	request *types.HistoryRespondActivityTaskCanceledRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RespondActivityTaskCanceled(ctx, request, opts...)
	}

	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) RecordActivityTaskHeartbeat(
	ctx context.Context,
	request *types.HistoryRecordActivityTaskHeartbeatRequest,
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

func (c *retryableClient) RequestCancelWorkflowExecution(
	ctx context.Context,
	request *types.HistoryRequestCancelWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RequestCancelWorkflowExecution(ctx, request, opts...)
	}

	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) SignalWorkflowExecution(
	ctx context.Context,
	request *types.HistorySignalWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.SignalWorkflowExecution(ctx, request, opts...)
	}

	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) SignalWithStartWorkflowExecution(
	ctx context.Context,
	request *types.HistorySignalWithStartWorkflowExecutionRequest,
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

func (c *retryableClient) RemoveSignalMutableState(
	ctx context.Context,
	request *types.RemoveSignalMutableStateRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RemoveSignalMutableState(ctx, request, opts...)
	}

	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) TerminateWorkflowExecution(
	ctx context.Context,
	request *types.HistoryTerminateWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.TerminateWorkflowExecution(ctx, request, opts...)
	}

	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) ResetWorkflowExecution(
	ctx context.Context,
	request *types.HistoryResetWorkflowExecutionRequest,
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

func (c *retryableClient) ScheduleDecisionTask(
	ctx context.Context,
	request *types.ScheduleDecisionTaskRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.ScheduleDecisionTask(ctx, request, opts...)
	}

	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) RecordChildExecutionCompleted(
	ctx context.Context,
	request *types.RecordChildExecutionCompletedRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RecordChildExecutionCompleted(ctx, request, opts...)
	}

	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) ReplicateEventsV2(
	ctx context.Context,
	request *types.ReplicateEventsV2Request,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.ReplicateEventsV2(ctx, request, opts...)
	}

	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) SyncShardStatus(
	ctx context.Context,
	request *types.SyncShardStatusRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.SyncShardStatus(ctx, request, opts...)
	}

	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) SyncActivity(
	ctx context.Context,
	request *types.SyncActivityRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.SyncActivity(ctx, request, opts...)
	}

	return backoff.Retry(op, c.policy, c.isRetryable)
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

	err := backoff.Retry(op, c.policy, c.isRetryable)
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

	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) QueryWorkflow(
	ctx context.Context,
	request *types.HistoryQueryWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.HistoryQueryWorkflowResponse, error) {
	var resp *types.HistoryQueryWorkflowResponse
	op := func() error {
		var err error
		resp, err = c.client.QueryWorkflow(ctx, request, opts...)
		return err
	}

	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) ReapplyEvents(
	ctx context.Context,
	request *types.HistoryReapplyEventsRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.ReapplyEvents(ctx, request, opts...)
	}

	return backoff.Retry(op, c.policy, c.isRetryable)
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

	err := backoff.Retry(op, c.policy, c.isRetryable)
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

	return backoff.Retry(op, c.policy, c.isRetryable)
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

	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *retryableClient) RefreshWorkflowTasks(
	ctx context.Context,
	request *types.HistoryRefreshWorkflowTasksRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RefreshWorkflowTasks(ctx, request, opts...)
	}

	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *retryableClient) NotifyFailoverMarkers(
	ctx context.Context,
	request *types.NotifyFailoverMarkersRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.NotifyFailoverMarkers(ctx, request, opts...)
	}

	return backoff.Retry(op, c.policy, c.isRetryable)
}
