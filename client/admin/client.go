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
	"time"

	"go.uber.org/yarpc"

	"github.com/uber/cadence/common/types"
)

var _ Client = (*clientImpl)(nil)

const (
	// DefaultTimeout is the default timeout used to make calls
	DefaultTimeout = 10 * time.Second
	// DefaultLargeTimeout is the default timeout used to make calls
	DefaultLargeTimeout = time.Minute
)

type clientImpl struct {
	timeout      time.Duration
	largeTimeout time.Duration
	client       Client
}

// NewClient creates a new admin service TChannel client
func NewClient(
	timeout time.Duration,
	largeTimeout time.Duration,
	client Client,
) Client {
	return &clientImpl{
		timeout:      timeout,
		largeTimeout: largeTimeout,
		client:       client,
	}
}

func (c *clientImpl) AddSearchAttribute(
	ctx context.Context,
	request *types.AddSearchAttributeRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.AddSearchAttribute(ctx, request, opts...)
}

func (c *clientImpl) DescribeShardDistribution(
	ctx context.Context,
	request *types.DescribeShardDistributionRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeShardDistributionResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.DescribeShardDistribution(ctx, request, opts...)
}

func (c *clientImpl) DescribeHistoryHost(
	ctx context.Context,
	request *types.DescribeHistoryHostRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeHistoryHostResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.DescribeHistoryHost(ctx, request, opts...)
}

func (c *clientImpl) RemoveTask(
	ctx context.Context,
	request *types.RemoveTaskRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RemoveTask(ctx, request, opts...)
}

func (c *clientImpl) CloseShard(
	ctx context.Context,
	request *types.CloseShardRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.CloseShard(ctx, request, opts...)
}

func (c *clientImpl) ResetQueue(
	ctx context.Context,
	request *types.ResetQueueRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.ResetQueue(ctx, request, opts...)
}

func (c *clientImpl) DescribeQueue(
	ctx context.Context,
	request *types.DescribeQueueRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeQueueResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.DescribeQueue(ctx, request, opts...)
}

func (c *clientImpl) DescribeWorkflowExecution(
	ctx context.Context,
	request *types.AdminDescribeWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.AdminDescribeWorkflowExecutionResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.DescribeWorkflowExecution(ctx, request, opts...)
}

func (c *clientImpl) GetWorkflowExecutionRawHistoryV2(
	ctx context.Context,
	request *types.GetWorkflowExecutionRawHistoryV2Request,
	opts ...yarpc.CallOption,
) (*types.GetWorkflowExecutionRawHistoryV2Response, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.GetWorkflowExecutionRawHistoryV2(ctx, request, opts...)
}

func (c *clientImpl) DescribeCluster(
	ctx context.Context,
	opts ...yarpc.CallOption,
) (*types.DescribeClusterResponse, error) {
	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.DescribeCluster(ctx, opts...)
}

func (c *clientImpl) GetReplicationMessages(
	ctx context.Context,
	request *types.GetReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.GetReplicationMessagesResponse, error) {
	ctx, cancel := c.createContextWithLargeTimeout(ctx)
	defer cancel()
	return c.client.GetReplicationMessages(ctx, request, opts...)
}

func (c *clientImpl) GetDomainReplicationMessages(
	ctx context.Context,
	request *types.GetDomainReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.GetDomainReplicationMessagesResponse, error) {
	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.GetDomainReplicationMessages(ctx, request, opts...)
}

func (c *clientImpl) GetDLQReplicationMessages(
	ctx context.Context,
	request *types.GetDLQReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.GetDLQReplicationMessagesResponse, error) {
	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.GetDLQReplicationMessages(ctx, request, opts...)
}

func (c *clientImpl) ReapplyEvents(
	ctx context.Context,
	request *types.ReapplyEventsRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.ReapplyEvents(ctx, request, opts...)
}

func (c *clientImpl) CountDLQMessages(
	ctx context.Context,
	request *types.CountDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.CountDLQMessagesResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.CountDLQMessages(ctx, request, opts...)
}

func (c *clientImpl) ReadDLQMessages(
	ctx context.Context,
	request *types.ReadDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.ReadDLQMessagesResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.ReadDLQMessages(ctx, request, opts...)
}

func (c *clientImpl) PurgeDLQMessages(
	ctx context.Context,
	request *types.PurgeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.PurgeDLQMessages(ctx, request, opts...)
}

func (c *clientImpl) MergeDLQMessages(
	ctx context.Context,
	request *types.MergeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.MergeDLQMessagesResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.MergeDLQMessages(ctx, request, opts...)

}

func (c *clientImpl) RefreshWorkflowTasks(
	ctx context.Context,
	request *types.RefreshWorkflowTasksRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RefreshWorkflowTasks(ctx, request, opts...)
}

func (c *clientImpl) ResendReplicationTasks(
	ctx context.Context,
	request *types.ResendReplicationTasksRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.ResendReplicationTasks(ctx, request, opts...)
}

func (c *clientImpl) GetCrossClusterTasks(
	ctx context.Context,
	request *types.GetCrossClusterTasksRequest,
	opts ...yarpc.CallOption,
) (*types.GetCrossClusterTasksResponse, error) {
	ctx, cancel := c.createContextWithLargeTimeout(ctx)
	defer cancel()
	return c.client.GetCrossClusterTasks(ctx, request, opts...)
}

func (c *clientImpl) RespondCrossClusterTasksCompleted(
	ctx context.Context,
	request *types.RespondCrossClusterTasksCompletedRequest,
	opts ...yarpc.CallOption,
) (*types.RespondCrossClusterTasksCompletedResponse, error) {
	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RespondCrossClusterTasksCompleted(ctx, request, opts...)
}

func (c *clientImpl) GetDynamicConfig(
	ctx context.Context,
	request *types.GetDynamicConfigRequest,
	opts ...yarpc.CallOption,
) (*types.GetDynamicConfigResponse, error) {
	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.GetDynamicConfig(ctx, request, opts...)
}

func (c *clientImpl) UpdateDynamicConfig(
	ctx context.Context,
	request *types.UpdateDynamicConfigRequest,
	opts ...yarpc.CallOption,
) error {
	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.UpdateDynamicConfig(ctx, request, opts...)
}

func (c *clientImpl) RestoreDynamicConfig(
	ctx context.Context,
	request *types.RestoreDynamicConfigRequest,
	opts ...yarpc.CallOption,
) error {
	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RestoreDynamicConfig(ctx, request, opts...)
}

func (c *clientImpl) DeleteWorkflow(
	ctx context.Context,
	request *types.AdminDeleteWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.AdminDeleteWorkflowResponse, error) {
	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.DeleteWorkflow(ctx, request, opts...)
}

func (c *clientImpl) MaintainCorruptWorkflow(
	ctx context.Context,
	request *types.AdminMaintainWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.AdminMaintainWorkflowResponse, error) {
	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.MaintainCorruptWorkflow(ctx, request, opts...)
}

func (c *clientImpl) ListDynamicConfig(
	ctx context.Context,
	request *types.ListDynamicConfigRequest,
	opts ...yarpc.CallOption,
) (*types.ListDynamicConfigResponse, error) {
	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.ListDynamicConfig(ctx, request, opts...)
}

func (c *clientImpl) createContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		return context.WithTimeout(context.Background(), c.timeout)
	}
	return context.WithTimeout(parent, c.timeout)
}

func (c *clientImpl) createContextWithLargeTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		return context.WithTimeout(context.Background(), c.largeTimeout)
	}
	return context.WithTimeout(parent, c.largeTimeout)
}
