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

package admin

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
	msgInjectedFakeErr = "Injected fake admin client error"
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

func (c *errorInjectionClient) AddSearchAttribute(
	ctx context.Context,
	request *types.AddSearchAttributeRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.AddSearchAttribute(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.AdminClientOperationAddSearchAttribute,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) DescribeShardDistribution(
	ctx context.Context,
	request *types.DescribeShardDistributionRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeShardDistributionResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.DescribeShardDistributionResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.DescribeShardDistribution(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.AdminClientOperationDescribeShardDistribution,
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
			tag.AdminClientOperationDescribeHistoryHost,
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
			tag.AdminClientOperationRemoveTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
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
			tag.AdminClientOperationCloseShard,
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
			tag.AdminClientOperationResetQueue,
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
			tag.AdminClientOperationDescribeQueue,
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
	request *types.AdminDescribeWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.AdminDescribeWorkflowExecutionResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.AdminDescribeWorkflowExecutionResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.DescribeWorkflowExecution(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.AdminClientOperationDescribeWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) GetWorkflowExecutionRawHistoryV2(
	ctx context.Context,
	request *types.GetWorkflowExecutionRawHistoryV2Request,
	opts ...yarpc.CallOption,
) (*types.GetWorkflowExecutionRawHistoryV2Response, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.GetWorkflowExecutionRawHistoryV2Response
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.GetWorkflowExecutionRawHistoryV2(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.AdminClientOperationGetWorkflowExecutionRawHistoryV2,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) DescribeCluster(
	ctx context.Context,
	opts ...yarpc.CallOption,
) (*types.DescribeClusterResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.DescribeClusterResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.DescribeCluster(ctx, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.AdminClientOperationDescribeCluster,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
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
			tag.AdminClientOperationGetReplicationMessages,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) GetDomainReplicationMessages(
	ctx context.Context,
	request *types.GetDomainReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.GetDomainReplicationMessagesResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.GetDomainReplicationMessagesResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.GetDomainReplicationMessages(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.AdminClientOperationGetDomainReplicationMessages,
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
			tag.AdminClientOperationGetDLQReplicationMessages,
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
	request *types.ReapplyEventsRequest,
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
			tag.AdminClientOperationReapplyEvents,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) CountDLQMessages(
	ctx context.Context,
	request *types.CountDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.CountDLQMessagesResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.CountDLQMessagesResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.CountDLQMessages(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.AdminClientOperationCountDLQMessages,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
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
			tag.AdminClientOperationReadDLQMessages,
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
			tag.AdminClientOperationPurgeDLQMessages,
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
			tag.AdminClientOperationMergeDLQMessages,
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
	request *types.RefreshWorkflowTasksRequest,
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
			tag.AdminClientOperationRefreshWorkflowTasks,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) ResendReplicationTasks(
	ctx context.Context,
	request *types.ResendReplicationTasksRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.ResendReplicationTasks(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.AdminClientOperationResendReplicationTasks,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) GetCrossClusterTasks(
	ctx context.Context,
	request *types.GetCrossClusterTasksRequest,
	opts ...yarpc.CallOption,
) (*types.GetCrossClusterTasksResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.GetCrossClusterTasksResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.GetCrossClusterTasks(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.AdminClientOperationGetCrossClusterTasks,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) RespondCrossClusterTasksCompleted(
	ctx context.Context,
	request *types.RespondCrossClusterTasksCompletedRequest,
	opts ...yarpc.CallOption,
) (*types.RespondCrossClusterTasksCompletedResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.RespondCrossClusterTasksCompletedResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.RespondCrossClusterTasksCompleted(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.AdminClientOperationRespondCrossClusterTasksCompleted,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) GetDynamicConfig(
	ctx context.Context,
	request *types.GetDynamicConfigRequest,
	opts ...yarpc.CallOption,
) (*types.GetDynamicConfigResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.GetDynamicConfigResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.GetDynamicConfig(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.AdminClientOperationGetDynamicConfig,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) UpdateDynamicConfig(
	ctx context.Context,
	request *types.UpdateDynamicConfigRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.UpdateDynamicConfig(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.AdminClientOperationUpdateDynamicConfig,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) RestoreDynamicConfig(
	ctx context.Context,
	request *types.RestoreDynamicConfigRequest,
	opts ...yarpc.CallOption,
) error {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		clientErr = c.client.RestoreDynamicConfig(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.AdminClientOperationRestoreDynamicConfig,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) DeleteWorkflow(
	ctx context.Context,
	request *types.AdminDeleteWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.AdminDeleteWorkflowResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.AdminDeleteWorkflowResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.DeleteWorkflow(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.AdminDeleteWorkflow,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) MaintainCorruptWorkflow(
	ctx context.Context,
	request *types.AdminMaintainWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.AdminMaintainWorkflowResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.AdminMaintainWorkflowResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.MaintainCorruptWorkflow(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.MaintainCorruptWorkflow,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) ListDynamicConfig(
	ctx context.Context,
	request *types.ListDynamicConfigRequest,
	opts ...yarpc.CallOption,
) (*types.ListDynamicConfigResponse, error) {
	fakeErr := errors.GenerateFakeError(c.errorRate)

	var resp *types.ListDynamicConfigResponse
	var clientErr error
	var forwardCall bool
	if forwardCall = errors.ShouldForwardCall(fakeErr); forwardCall {
		resp, clientErr = c.client.ListDynamicConfig(ctx, request, opts...)
	}

	if fakeErr != nil {
		c.logger.Error(msgInjectedFakeErr,
			tag.AdminClientOperationListDynamicConfig,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.ClientError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}
