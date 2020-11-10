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

package history

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
	msgInjectedFakeErr = "Injected fake history client error"
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

func (c *errorInjectionClient) StartWorkflowExecution(
	ctx context.Context,
	request *types.HistoryStartWorkflowExecutionRequest,
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
			tag.HistoryClientOperationStartWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) DescribeHistoryHost(
	ctx context.Context,
	request *types.DescribeHistoryHostRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeHistoryHostResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.DescribeHistoryHostResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.DescribeHistoryHost(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationDescribeHistoryHost,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) CloseShard(
	ctx context.Context,
	request *types.CloseShardRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.CloseShard(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationCloseShard,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) ResetQueue(
	ctx context.Context,
	request *types.ResetQueueRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.ResetQueue(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationResetQueue,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) DescribeQueue(
	ctx context.Context,
	request *types.DescribeQueueRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeQueueResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.DescribeQueueResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.DescribeQueue(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationDescribeQueue,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) RemoveTask(
	ctx context.Context,
	request *types.RemoveTaskRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.RemoveTask(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationRemoveTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) DescribeMutableState(
	ctx context.Context,
	request *types.DescribeMutableStateRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeMutableStateResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.DescribeMutableStateResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.DescribeMutableState(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationDescribeMutableState,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) GetMutableState(
	ctx context.Context,
	request *types.GetMutableStateRequest,
	opts ...yarpc.CallOption,
) (*types.GetMutableStateResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.GetMutableStateResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.GetMutableState(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationGetMutableState,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) PollMutableState(
	ctx context.Context,
	request *types.PollMutableStateRequest,
	opts ...yarpc.CallOption,
) (*types.PollMutableStateResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.PollMutableStateResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.PollMutableState(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationPollMutableState,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) ResetStickyTaskList(
	ctx context.Context,
	request *types.HistoryResetStickyTaskListRequest,
	opts ...yarpc.CallOption,
) (*types.HistoryResetStickyTaskListResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.HistoryResetStickyTaskListResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.ResetStickyTaskList(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationResetStickyTaskList,
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
	request *types.HistoryDescribeWorkflowExecutionRequest,
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
			tag.HistoryClientOperationDescribeWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) RecordDecisionTaskStarted(
	ctx context.Context,
	request *types.RecordDecisionTaskStartedRequest,
	opts ...yarpc.CallOption,
) (*types.RecordDecisionTaskStartedResponse, error) {

	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.RecordDecisionTaskStartedResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.RecordDecisionTaskStarted(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationRecordDecisionTaskStarted,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) RecordActivityTaskStarted(
	ctx context.Context,
	request *types.RecordActivityTaskStartedRequest,
	opts ...yarpc.CallOption,
) (*types.RecordActivityTaskStartedResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.RecordActivityTaskStartedResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.RecordActivityTaskStarted(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationRecordActivityTaskStarted,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) RespondDecisionTaskCompleted(
	ctx context.Context,
	request *types.HistoryRespondDecisionTaskCompletedRequest,
	opts ...yarpc.CallOption,
) (*types.HistoryRespondDecisionTaskCompletedResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.HistoryRespondDecisionTaskCompletedResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.RespondDecisionTaskCompleted(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationRecordDecisionTaskCompleted,
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
	request *types.HistoryRespondDecisionTaskFailedRequest,
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
			tag.HistoryClientOperationRecordDecisionTaskFailed,
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
	request *types.HistoryRespondActivityTaskCompletedRequest,
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
			tag.HistoryClientOperationRecordActivityTaskCompleted,
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
	request *types.HistoryRespondActivityTaskFailedRequest,
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
			tag.HistoryClientOperationRecordActivityTaskFailed,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) RespondActivityTaskCanceled(
	ctx context.Context,
	request *types.HistoryRespondActivityTaskCanceledRequest,
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
			tag.HistoryClientOperationRecordActivityTaskCanceled,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) RecordActivityTaskHeartbeat(
	ctx context.Context,
	request *types.HistoryRecordActivityTaskHeartbeatRequest,
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
			tag.HistoryClientOperationRecordActivityTaskHeartbeat,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) RequestCancelWorkflowExecution(
	ctx context.Context,
	request *types.HistoryRequestCancelWorkflowExecutionRequest,
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
			tag.HistoryClientOperationRequestCancelWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) SignalWorkflowExecution(
	ctx context.Context,
	request *types.HistorySignalWorkflowExecutionRequest,
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
			tag.HistoryClientOperationSignalWorkflowExecution,
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
	request *types.HistorySignalWithStartWorkflowExecutionRequest,
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
			tag.HistoryClientOperationSignalWithStartWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) RemoveSignalMutableState(
	ctx context.Context,
	request *types.RemoveSignalMutableStateRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.RemoveSignalMutableState(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationRemoveSignalMutableState,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) TerminateWorkflowExecution(
	ctx context.Context,
	request *types.HistoryTerminateWorkflowExecutionRequest,
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
			tag.HistoryClientOperationTerminateWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) ResetWorkflowExecution(
	ctx context.Context,
	request *types.HistoryResetWorkflowExecutionRequest,
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
			tag.HistoryClientOperationResetWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) ScheduleDecisionTask(
	ctx context.Context,
	request *types.ScheduleDecisionTaskRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.ScheduleDecisionTask(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationScheduleDecisionTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) RecordChildExecutionCompleted(
	ctx context.Context,
	request *types.RecordChildExecutionCompletedRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.RecordChildExecutionCompleted(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationRecordChildExecutionCompleted,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) ReplicateEventsV2(
	ctx context.Context,
	request *types.ReplicateEventsV2Request,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.ReplicateEventsV2(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationReplicateEventsV2,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) SyncShardStatus(
	ctx context.Context,
	request *types.SyncShardStatusRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.SyncShardStatus(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationSyncShardStatus,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) SyncActivity(
	ctx context.Context,
	request *types.SyncActivityRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.SyncActivity(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationSyncActivity,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) GetReplicationMessages(
	ctx context.Context,
	request *types.GetReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.GetReplicationMessagesResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.GetReplicationMessagesResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.GetReplicationMessages(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationGetReplicationMessages,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) GetDLQReplicationMessages(
	ctx context.Context,
	request *types.GetDLQReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.GetDLQReplicationMessagesResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.GetDLQReplicationMessagesResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.GetDLQReplicationMessages(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationGetDLQReplicationMessages,
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
	request *types.HistoryQueryWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.HistoryQueryWorkflowResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.HistoryQueryWorkflowResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.QueryWorkflow(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationQueryWorkflow,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) ReapplyEvents(
	ctx context.Context,
	request *types.HistoryReapplyEventsRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.ReapplyEvents(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationReapplyEvents,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) ReadDLQMessages(
	ctx context.Context,
	request *types.ReadDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.ReadDLQMessagesResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.ReadDLQMessagesResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.ReadDLQMessages(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationReadDLQMessages,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) PurgeDLQMessages(
	ctx context.Context,
	request *types.PurgeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.PurgeDLQMessages(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationPurgeDLQMessages,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) MergeDLQMessages(
	ctx context.Context,
	request *types.MergeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.MergeDLQMessagesResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.MergeDLQMessagesResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.MergeDLQMessages(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationMergeDLQMessages,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) RefreshWorkflowTasks(
	ctx context.Context,
	request *types.HistoryRefreshWorkflowTasksRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.RefreshWorkflowTasks(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationRefreshWorkflowTasks,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) NotifyFailoverMarkers(
	ctx context.Context,
	request *types.NotifyFailoverMarkersRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.NotifyFailoverMarkers(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.HistoryClientOperationNotifyFailoverMarkers,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}
