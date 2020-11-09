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

	"github.com/uber/cadence/.gen/go/admin"
	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/errors"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/types"
	"go.uber.org/yarpc"
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
	request *admin.AddSearchAttributeRequest,
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
			tag.StoreError(clientErr),
		)
		return fakeErr
	}
	return clientErr
}

func (c *errorInjectionClient) DescribeHistoryHost(
	ctx context.Context,
	request *shared.DescribeHistoryHostRequest,
	opts ...yarpc.CallOption,
) (*shared.DescribeHistoryHostResponse, error) {
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
			tag.StoreError(clientErr),
		)
		return nil, fakeErr
	}
	return resp, clientErr
}

func (c *errorInjectionClient) RemoveTask(
	ctx context.Context,
	request *shared.RemoveTaskRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RemoveTask(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *errorInjectionClient) CloseShard(
	ctx context.Context,
	request *shared.CloseShardRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.CloseShard(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *errorInjectionClient) ResetQueue(
	ctx context.Context,
	request *shared.ResetQueueRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.ResetQueue(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *errorInjectionClient) DescribeQueue(
	ctx context.Context,
	request *shared.DescribeQueueRequest,
	opts ...yarpc.CallOption,
) (*shared.DescribeQueueResponse, error) {

	var resp *shared.DescribeQueueResponse
	op := func() error {
		var err error
		resp, err = c.client.DescribeQueue(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *errorInjectionClient) DescribeWorkflowExecution(
	ctx context.Context,
	request *admin.DescribeWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*admin.DescribeWorkflowExecutionResponse, error) {

	var resp *admin.DescribeWorkflowExecutionResponse
	op := func() error {
		var err error
		resp, err = c.client.DescribeWorkflowExecution(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *errorInjectionClient) GetWorkflowExecutionRawHistoryV2(
	ctx context.Context,
	request *admin.GetWorkflowExecutionRawHistoryV2Request,
	opts ...yarpc.CallOption,
) (*admin.GetWorkflowExecutionRawHistoryV2Response, error) {

	var resp *admin.GetWorkflowExecutionRawHistoryV2Response
	op := func() error {
		var err error
		resp, err = c.client.GetWorkflowExecutionRawHistoryV2(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *errorInjectionClient) DescribeCluster(
	ctx context.Context,
	opts ...yarpc.CallOption,
) (*admin.DescribeClusterResponse, error) {

	var resp *admin.DescribeClusterResponse
	op := func() error {
		var err error
		resp, err = c.client.DescribeCluster(ctx, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *errorInjectionClient) GetReplicationMessages(
	ctx context.Context,
	request *replicator.GetReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*replicator.GetReplicationMessagesResponse, error) {
	var resp *replicator.GetReplicationMessagesResponse
	op := func() error {
		var err error
		resp, err = c.client.GetReplicationMessages(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *errorInjectionClient) GetDomainReplicationMessages(
	ctx context.Context,
	request *replicator.GetDomainReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*replicator.GetDomainReplicationMessagesResponse, error) {
	var resp *replicator.GetDomainReplicationMessagesResponse
	op := func() error {
		var err error
		resp, err = c.client.GetDomainReplicationMessages(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *errorInjectionClient) GetDLQReplicationMessages(
	ctx context.Context,
	request *replicator.GetDLQReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*replicator.GetDLQReplicationMessagesResponse, error) {
	var resp *replicator.GetDLQReplicationMessagesResponse
	op := func() error {
		var err error
		resp, err = c.client.GetDLQReplicationMessages(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *errorInjectionClient) ReapplyEvents(
	ctx context.Context,
	request *shared.ReapplyEventsRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.ReapplyEvents(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *errorInjectionClient) ReadDLQMessages(
	ctx context.Context,
	request *replicator.ReadDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*replicator.ReadDLQMessagesResponse, error) {

	var resp *replicator.ReadDLQMessagesResponse
	op := func() error {
		var err error
		resp, err = c.client.ReadDLQMessages(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *errorInjectionClient) PurgeDLQMessages(
	ctx context.Context,
	request *replicator.PurgeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.PurgeDLQMessages(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *errorInjectionClient) MergeDLQMessages(
	ctx context.Context,
	request *replicator.MergeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*replicator.MergeDLQMessagesResponse, error) {

	var resp *replicator.MergeDLQMessagesResponse
	op := func() error {
		var err error
		resp, err = c.client.MergeDLQMessages(ctx, request, opts...)
		return err
	}
	err := backoff.Retry(op, c.policy, c.isRetryable)
	return resp, err
}

func (c *errorInjectionClient) RefreshWorkflowTasks(
	ctx context.Context,
	request *shared.RefreshWorkflowTasksRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.RefreshWorkflowTasks(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}

func (c *errorInjectionClient) ResendReplicationTasks(
	ctx context.Context,
	request *admin.ResendReplicationTasksRequest,
	opts ...yarpc.CallOption,
) error {

	op := func() error {
		return c.client.ResendReplicationTasks(ctx, request, opts...)
	}
	return backoff.Retry(op, c.policy, c.isRetryable)
}
