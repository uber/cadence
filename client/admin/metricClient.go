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

	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
)

var _ Client = (*metricClient)(nil)

type metricClient struct {
	client        Client
	metricsClient metrics.Client
}

// NewMetricClient creates a new instance of Client that emits metrics
func NewMetricClient(client Client, metricsClient metrics.Client) Client {
	return &metricClient{
		client:        client,
		metricsClient: metricsClient,
	}
}

func (c *metricClient) AddSearchAttribute(
	ctx context.Context,
	request *types.AddSearchAttributeRequest,
	opts ...yarpc.CallOption,
) error {

	c.metricsClient.IncCounter(metrics.AdminClientAddSearchAttributeScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.AdminClientAddSearchAttributeScope, metrics.CadenceClientLatency)
	err := c.client.AddSearchAttribute(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientAddSearchAttributeScope, metrics.CadenceClientFailures)
	}
	return err
}

func (c *metricClient) DescribeShardDistribution(
	ctx context.Context,
	request *types.DescribeShardDistributionRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeShardDistributionResponse, error) {

	c.metricsClient.IncCounter(metrics.AdminClientDescribeShardDistributionScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.AdminClientDescribeHistoryHostScope, metrics.CadenceClientLatency)
	resp, err := c.client.DescribeShardDistribution(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientDescribeHistoryHostScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) DescribeHistoryHost(
	ctx context.Context,
	request *types.DescribeHistoryHostRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeHistoryHostResponse, error) {

	c.metricsClient.IncCounter(metrics.AdminClientDescribeHistoryHostScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.AdminClientDescribeHistoryHostScope, metrics.CadenceClientLatency)
	resp, err := c.client.DescribeHistoryHost(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientDescribeHistoryHostScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) RemoveTask(
	ctx context.Context,
	request *types.RemoveTaskRequest,
	opts ...yarpc.CallOption,
) error {

	c.metricsClient.IncCounter(metrics.AdminClientRemoveTaskScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.AdminClientRemoveTaskScope, metrics.CadenceClientLatency)
	err := c.client.RemoveTask(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientRemoveTaskScope, metrics.CadenceClientFailures)
	}
	return err
}

func (c *metricClient) CloseShard(
	ctx context.Context,
	request *types.CloseShardRequest,
	opts ...yarpc.CallOption,
) error {

	c.metricsClient.IncCounter(metrics.AdminClientCloseShardScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.AdminClientCloseShardScope, metrics.CadenceClientLatency)
	err := c.client.CloseShard(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientCloseShardScope, metrics.CadenceClientFailures)
	}
	return err
}

func (c *metricClient) ResetQueue(
	ctx context.Context,
	request *types.ResetQueueRequest,
	opts ...yarpc.CallOption,
) error {

	c.metricsClient.IncCounter(metrics.AdminClientResetQueueScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.AdminClientResetQueueScope, metrics.CadenceClientLatency)
	err := c.client.ResetQueue(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientResetQueueScope, metrics.CadenceClientFailures)
	}
	return err
}

func (c *metricClient) DescribeQueue(
	ctx context.Context,
	request *types.DescribeQueueRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeQueueResponse, error) {

	c.metricsClient.IncCounter(metrics.AdminClientDescribeQueueScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.AdminClientDescribeQueueScope, metrics.CadenceClientLatency)
	resp, err := c.client.DescribeQueue(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientDescribeQueueScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) DescribeWorkflowExecution(
	ctx context.Context,
	request *types.AdminDescribeWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.AdminDescribeWorkflowExecutionResponse, error) {

	c.metricsClient.IncCounter(metrics.AdminClientDescribeWorkflowExecutionScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.AdminClientDescribeWorkflowExecutionScope, metrics.CadenceClientLatency)
	resp, err := c.client.DescribeWorkflowExecution(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientDescribeWorkflowExecutionScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) GetWorkflowExecutionRawHistoryV2(
	ctx context.Context,
	request *types.GetWorkflowExecutionRawHistoryV2Request,
	opts ...yarpc.CallOption,
) (*types.GetWorkflowExecutionRawHistoryV2Response, error) {

	c.metricsClient.IncCounter(metrics.AdminClientGetWorkflowExecutionRawHistoryV2Scope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.AdminClientGetWorkflowExecutionRawHistoryV2Scope, metrics.CadenceClientLatency)
	resp, err := c.client.GetWorkflowExecutionRawHistoryV2(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientGetWorkflowExecutionRawHistoryV2Scope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) DescribeCluster(
	ctx context.Context,
	opts ...yarpc.CallOption,
) (*types.DescribeClusterResponse, error) {

	c.metricsClient.IncCounter(metrics.AdminClientDescribeClusterScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.AdminClientDescribeClusterScope, metrics.CadenceClientLatency)
	resp, err := c.client.DescribeCluster(ctx, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientDescribeClusterScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) GetReplicationMessages(
	ctx context.Context,
	request *types.GetReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.GetReplicationMessagesResponse, error) {
	c.metricsClient.IncCounter(metrics.FrontendClientGetReplicationTasksScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.FrontendClientGetReplicationTasksScope, metrics.CadenceClientLatency)
	resp, err := c.client.GetReplicationMessages(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.FrontendClientGetReplicationTasksScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) GetDomainReplicationMessages(
	ctx context.Context,
	request *types.GetDomainReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.GetDomainReplicationMessagesResponse, error) {
	c.metricsClient.IncCounter(metrics.FrontendClientGetDomainReplicationTasksScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.FrontendClientGetDomainReplicationTasksScope, metrics.CadenceClientLatency)
	resp, err := c.client.GetDomainReplicationMessages(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.FrontendClientGetDomainReplicationTasksScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) GetDLQReplicationMessages(
	ctx context.Context,
	request *types.GetDLQReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.GetDLQReplicationMessagesResponse, error) {
	c.metricsClient.IncCounter(metrics.FrontendClientGetDLQReplicationTasksScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.FrontendClientGetDLQReplicationTasksScope, metrics.CadenceClientLatency)
	resp, err := c.client.GetDLQReplicationMessages(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.FrontendClientGetDLQReplicationTasksScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) ReapplyEvents(
	ctx context.Context,
	request *types.ReapplyEventsRequest,
	opts ...yarpc.CallOption,
) error {

	c.metricsClient.IncCounter(metrics.FrontendClientReapplyEventsScope, metrics.CadenceClientRequests)
	sw := c.metricsClient.StartTimer(metrics.FrontendClientReapplyEventsScope, metrics.CadenceClientLatency)
	err := c.client.ReapplyEvents(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.FrontendClientReapplyEventsScope, metrics.CadenceClientFailures)
	}
	return err
}

func (c *metricClient) CountDLQMessages(
	ctx context.Context,
	request *types.CountDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.CountDLQMessagesResponse, error) {

	c.metricsClient.IncCounter(metrics.AdminClientCountDLQMessagesScope, metrics.CadenceClientRequests)
	sw := c.metricsClient.StartTimer(metrics.AdminClientCountDLQMessagesScope, metrics.CadenceClientLatency)
	resp, err := c.client.CountDLQMessages(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientCountDLQMessagesScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) ReadDLQMessages(
	ctx context.Context,
	request *types.ReadDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.ReadDLQMessagesResponse, error) {

	c.metricsClient.IncCounter(metrics.AdminClientReadDLQMessagesScope, metrics.CadenceClientRequests)
	sw := c.metricsClient.StartTimer(metrics.AdminClientReadDLQMessagesScope, metrics.CadenceClientLatency)
	resp, err := c.client.ReadDLQMessages(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientReadDLQMessagesScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) PurgeDLQMessages(
	ctx context.Context,
	request *types.PurgeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) error {

	c.metricsClient.IncCounter(metrics.AdminClientPurgeDLQMessagesScope, metrics.CadenceClientRequests)
	sw := c.metricsClient.StartTimer(metrics.AdminClientPurgeDLQMessagesScope, metrics.CadenceClientLatency)
	err := c.client.PurgeDLQMessages(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientPurgeDLQMessagesScope, metrics.CadenceClientFailures)
	}
	return err
}

func (c *metricClient) MergeDLQMessages(
	ctx context.Context,
	request *types.MergeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.MergeDLQMessagesResponse, error) {

	c.metricsClient.IncCounter(metrics.AdminClientMergeDLQMessagesScope, metrics.CadenceClientRequests)
	sw := c.metricsClient.StartTimer(metrics.AdminClientMergeDLQMessagesScope, metrics.CadenceClientLatency)
	resp, err := c.client.MergeDLQMessages(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientMergeDLQMessagesScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) RefreshWorkflowTasks(
	ctx context.Context,
	request *types.RefreshWorkflowTasksRequest,
	opts ...yarpc.CallOption,
) error {

	c.metricsClient.IncCounter(metrics.AdminClientRefreshWorkflowTasksScope, metrics.CadenceClientRequests)
	sw := c.metricsClient.StartTimer(metrics.AdminClientRefreshWorkflowTasksScope, metrics.CadenceClientLatency)
	err := c.client.RefreshWorkflowTasks(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientRefreshWorkflowTasksScope, metrics.CadenceClientFailures)
	}
	return err
}

func (c *metricClient) ResendReplicationTasks(
	ctx context.Context,
	request *types.ResendReplicationTasksRequest,
	opts ...yarpc.CallOption,
) error {

	c.metricsClient.IncCounter(metrics.AdminClientResendReplicationTasksScope, metrics.CadenceClientRequests)
	sw := c.metricsClient.StartTimer(metrics.AdminClientResendReplicationTasksScope, metrics.CadenceClientLatency)
	err := c.client.ResendReplicationTasks(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientResendReplicationTasksScope, metrics.CadenceClientFailures)
	}
	return err
}

func (c *metricClient) GetCrossClusterTasks(
	ctx context.Context,
	request *types.GetCrossClusterTasksRequest,
	opts ...yarpc.CallOption,
) (*types.GetCrossClusterTasksResponse, error) {
	c.metricsClient.IncCounter(metrics.AdminClientGetCrossClusterTasksScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.AdminClientGetCrossClusterTasksScope, metrics.CadenceClientLatency)
	resp, err := c.client.GetCrossClusterTasks(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientGetCrossClusterTasksScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) RespondCrossClusterTasksCompleted(
	ctx context.Context,
	request *types.RespondCrossClusterTasksCompletedRequest,
	opts ...yarpc.CallOption,
) (*types.RespondCrossClusterTasksCompletedResponse, error) {
	c.metricsClient.IncCounter(metrics.AdminClientRespondCrossClusterTasksCompletedScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.AdminClientRespondCrossClusterTasksCompletedScope, metrics.CadenceClientLatency)
	resp, err := c.client.RespondCrossClusterTasksCompleted(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientRespondCrossClusterTasksCompletedScope, metrics.CadenceClientFailures)
	}

	return resp, err
}

func (c *metricClient) GetDynamicConfig(
	ctx context.Context,
	request *types.GetDynamicConfigRequest,
	opts ...yarpc.CallOption,
) (*types.GetDynamicConfigResponse, error) {
	c.metricsClient.IncCounter(metrics.AdminClientGetDynamicConfigScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.AdminClientGetDynamicConfigScope, metrics.CadenceClientLatency)
	resp, err := c.client.GetDynamicConfig(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientGetDynamicConfigScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) UpdateDynamicConfig(
	ctx context.Context,
	request *types.UpdateDynamicConfigRequest,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.AdminClientUpdateDynamicConfigScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.AdminClientUpdateDynamicConfigScope, metrics.CadenceClientLatency)
	err := c.client.UpdateDynamicConfig(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientUpdateDynamicConfigScope, metrics.CadenceClientFailures)
	}
	return err
}

func (c *metricClient) RestoreDynamicConfig(
	ctx context.Context,
	request *types.RestoreDynamicConfigRequest,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.AdminClientRestoreDynamicConfigScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.AdminClientRestoreDynamicConfigScope, metrics.CadenceClientLatency)
	err := c.client.RestoreDynamicConfig(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientRestoreDynamicConfigScope, metrics.CadenceClientFailures)
	}
	return err
}

func (c *metricClient) DeleteWorkflow(
	ctx context.Context,
	request *types.AdminDeleteWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.AdminDeleteWorkflowResponse, error) {
	c.metricsClient.IncCounter(metrics.AdminDeleteWorkflowScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.AdminDeleteWorkflowScope, metrics.CadenceClientLatency)
	resp, err := c.client.DeleteWorkflow(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminDeleteWorkflowScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) MaintainCorruptWorkflow(
	ctx context.Context,
	request *types.AdminMaintainWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.AdminMaintainWorkflowResponse, error) {
	c.metricsClient.IncCounter(metrics.MaintainCorruptWorkflowScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.MaintainCorruptWorkflowScope, metrics.CadenceClientLatency)
	resp, err := c.client.MaintainCorruptWorkflow(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.MaintainCorruptWorkflowScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) ListDynamicConfig(
	ctx context.Context,
	request *types.ListDynamicConfigRequest,
	opts ...yarpc.CallOption,
) (*types.ListDynamicConfigResponse, error) {
	c.metricsClient.IncCounter(metrics.AdminClientListDynamicConfigScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.AdminClientListDynamicConfigScope, metrics.CadenceClientLatency)
	resp, err := c.client.ListDynamicConfig(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.AdminClientListDynamicConfigScope, metrics.CadenceClientFailures)
	}
	return resp, err
}
