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

func (c *metricClient) StartWorkflowExecution(
	context context.Context,
	request *types.HistoryStartWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.StartWorkflowExecutionResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientStartWorkflowExecutionScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientStartWorkflowExecutionScope, metrics.CadenceClientLatency)
	resp, err := c.client.StartWorkflowExecution(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientStartWorkflowExecutionScope, metrics.CadenceClientFailures)
	}

	return resp, err
}

func (c *metricClient) DescribeHistoryHost(
	context context.Context,
	request *types.DescribeHistoryHostRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeHistoryHostResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientDescribeHistoryHostScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientDescribeHistoryHostScope, metrics.CadenceClientLatency)
	resp, err := c.client.DescribeHistoryHost(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientDescribeHistoryHostScope, metrics.CadenceClientFailures)
	}

	return resp, err
}

func (c *metricClient) RemoveTask(
	context context.Context,
	request *types.RemoveTaskRequest,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.HistoryClientRemoveTaskScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRemoveTaskScope, metrics.CadenceClientLatency)
	err := c.client.RemoveTask(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRemoveTaskScope, metrics.CadenceClientFailures)
	}

	return err
}

func (c *metricClient) CloseShard(
	context context.Context,
	request *types.CloseShardRequest,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.HistoryClientCloseShardScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientCloseShardScope, metrics.CadenceClientLatency)
	err := c.client.CloseShard(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientCloseShardScope, metrics.CadenceClientFailures)
	}

	return err
}

func (c *metricClient) ResetQueue(
	context context.Context,
	request *types.ResetQueueRequest,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.HistoryClientResetQueueScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientResetQueueScope, metrics.CadenceClientLatency)
	err := c.client.ResetQueue(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientResetQueueScope, metrics.CadenceClientFailures)
	}

	return err
}

func (c *metricClient) DescribeQueue(
	context context.Context,
	request *types.DescribeQueueRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeQueueResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientDescribeQueueScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientDescribeQueueScope, metrics.CadenceClientLatency)
	resp, err := c.client.DescribeQueue(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientDescribeQueueScope, metrics.CadenceClientFailures)
	}

	return resp, err
}

func (c *metricClient) DescribeMutableState(
	context context.Context,
	request *types.DescribeMutableStateRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeMutableStateResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientDescribeMutableStateScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientDescribeMutableStateScope, metrics.CadenceClientLatency)
	resp, err := c.client.DescribeMutableState(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientDescribeMutableStateScope, metrics.CadenceClientFailures)
	}

	return resp, err
}

func (c *metricClient) GetMutableState(
	context context.Context,
	request *types.GetMutableStateRequest,
	opts ...yarpc.CallOption,
) (*types.GetMutableStateResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientGetMutableStateScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientGetMutableStateScope, metrics.CadenceClientLatency)
	resp, err := c.client.GetMutableState(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientGetMutableStateScope, metrics.CadenceClientFailures)
	}

	return resp, err
}

func (c *metricClient) PollMutableState(
	context context.Context,
	request *types.PollMutableStateRequest,
	opts ...yarpc.CallOption,
) (*types.PollMutableStateResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientPollMutableStateScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientPollMutableStateScope, metrics.CadenceClientLatency)
	resp, err := c.client.PollMutableState(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientPollMutableStateScope, metrics.CadenceClientFailures)
	}

	return resp, err
}

func (c *metricClient) ResetStickyTaskList(
	context context.Context,
	request *types.HistoryResetStickyTaskListRequest,
	opts ...yarpc.CallOption,
) (*types.HistoryResetStickyTaskListResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientResetStickyTaskListScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientResetStickyTaskListScope, metrics.CadenceClientLatency)
	resp, err := c.client.ResetStickyTaskList(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientResetStickyTaskListScope, metrics.CadenceClientFailures)
	}

	return resp, err
}

func (c *metricClient) DescribeWorkflowExecution(
	context context.Context,
	request *types.HistoryDescribeWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeWorkflowExecutionResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientDescribeWorkflowExecutionScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientDescribeWorkflowExecutionScope, metrics.CadenceClientLatency)
	resp, err := c.client.DescribeWorkflowExecution(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientDescribeWorkflowExecutionScope, metrics.CadenceClientFailures)
	}

	return resp, err
}

func (c *metricClient) RecordDecisionTaskStarted(
	context context.Context,
	request *types.RecordDecisionTaskStartedRequest,
	opts ...yarpc.CallOption,
) (*types.RecordDecisionTaskStartedResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientRecordDecisionTaskStartedScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRecordDecisionTaskStartedScope, metrics.CadenceClientLatency)
	resp, err := c.client.RecordDecisionTaskStarted(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRecordDecisionTaskStartedScope, metrics.CadenceClientFailures)
	}

	return resp, err
}

func (c *metricClient) RecordActivityTaskStarted(
	context context.Context,
	request *types.RecordActivityTaskStartedRequest,
	opts ...yarpc.CallOption,
) (*types.RecordActivityTaskStartedResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientRecordActivityTaskStartedScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRecordActivityTaskStartedScope, metrics.CadenceClientLatency)
	resp, err := c.client.RecordActivityTaskStarted(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRecordActivityTaskStartedScope, metrics.CadenceClientFailures)
	}

	return resp, err
}

func (c *metricClient) RespondDecisionTaskCompleted(
	context context.Context,
	request *types.HistoryRespondDecisionTaskCompletedRequest,
	opts ...yarpc.CallOption,
) (*types.HistoryRespondDecisionTaskCompletedResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientRespondDecisionTaskCompletedScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRespondDecisionTaskCompletedScope, metrics.CadenceClientLatency)
	response, err := c.client.RespondDecisionTaskCompleted(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRespondDecisionTaskCompletedScope, metrics.CadenceClientFailures)
	}

	return response, err
}

func (c *metricClient) RespondDecisionTaskFailed(
	context context.Context,
	request *types.HistoryRespondDecisionTaskFailedRequest,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.HistoryClientRespondDecisionTaskFailedScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRespondDecisionTaskFailedScope, metrics.CadenceClientLatency)
	err := c.client.RespondDecisionTaskFailed(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRespondDecisionTaskFailedScope, metrics.CadenceClientFailures)
	}

	return err
}

func (c *metricClient) RespondActivityTaskCompleted(
	context context.Context,
	request *types.HistoryRespondActivityTaskCompletedRequest,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.HistoryClientRespondActivityTaskCompletedScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRespondActivityTaskCompletedScope, metrics.CadenceClientLatency)
	err := c.client.RespondActivityTaskCompleted(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRespondActivityTaskCompletedScope, metrics.CadenceClientFailures)
	}

	return err
}

func (c *metricClient) RespondActivityTaskFailed(
	context context.Context,
	request *types.HistoryRespondActivityTaskFailedRequest,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.HistoryClientRespondActivityTaskFailedScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRespondActivityTaskFailedScope, metrics.CadenceClientLatency)
	err := c.client.RespondActivityTaskFailed(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRespondActivityTaskFailedScope, metrics.CadenceClientFailures)
	}

	return err
}

func (c *metricClient) RespondActivityTaskCanceled(
	context context.Context,
	request *types.HistoryRespondActivityTaskCanceledRequest,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.HistoryClientRespondActivityTaskCanceledScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRespondActivityTaskCanceledScope, metrics.CadenceClientLatency)
	err := c.client.RespondActivityTaskCanceled(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRespondActivityTaskCanceledScope, metrics.CadenceClientFailures)
	}

	return err
}

func (c *metricClient) RecordActivityTaskHeartbeat(
	context context.Context,
	request *types.HistoryRecordActivityTaskHeartbeatRequest,
	opts ...yarpc.CallOption,
) (*types.RecordActivityTaskHeartbeatResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientRecordActivityTaskHeartbeatScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRecordActivityTaskHeartbeatScope, metrics.CadenceClientLatency)
	resp, err := c.client.RecordActivityTaskHeartbeat(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRecordActivityTaskHeartbeatScope, metrics.CadenceClientFailures)
	}

	return resp, err
}

func (c *metricClient) RequestCancelWorkflowExecution(
	context context.Context,
	request *types.HistoryRequestCancelWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.HistoryClientRequestCancelWorkflowExecutionScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRequestCancelWorkflowExecutionScope, metrics.CadenceClientLatency)
	err := c.client.RequestCancelWorkflowExecution(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRequestCancelWorkflowExecutionScope, metrics.CadenceClientFailures)
	}

	return err
}

func (c *metricClient) SignalWorkflowExecution(
	context context.Context,
	request *types.HistorySignalWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.HistoryClientSignalWorkflowExecutionScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientSignalWorkflowExecutionScope, metrics.CadenceClientLatency)
	err := c.client.SignalWorkflowExecution(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientSignalWorkflowExecutionScope, metrics.CadenceClientFailures)
	}

	return err
}

func (c *metricClient) SignalWithStartWorkflowExecution(
	context context.Context,
	request *types.HistorySignalWithStartWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.StartWorkflowExecutionResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientSignalWithStartWorkflowExecutionScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientSignalWithStartWorkflowExecutionScope, metrics.CadenceClientLatency)
	resp, err := c.client.SignalWithStartWorkflowExecution(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientSignalWithStartWorkflowExecutionScope, metrics.CadenceClientFailures)
	}

	return resp, err
}

func (c *metricClient) RemoveSignalMutableState(
	context context.Context,
	request *types.RemoveSignalMutableStateRequest,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.HistoryClientRemoveSignalMutableStateScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRemoveSignalMutableStateScope, metrics.CadenceClientLatency)
	err := c.client.RemoveSignalMutableState(context, request)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRemoveSignalMutableStateScope, metrics.CadenceClientFailures)
	}

	return err
}

func (c *metricClient) TerminateWorkflowExecution(
	context context.Context,
	request *types.HistoryTerminateWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.HistoryClientTerminateWorkflowExecutionScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientTerminateWorkflowExecutionScope, metrics.CadenceClientLatency)
	err := c.client.TerminateWorkflowExecution(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientTerminateWorkflowExecutionScope, metrics.CadenceClientFailures)
	}

	return err
}

func (c *metricClient) ResetWorkflowExecution(
	context context.Context,
	request *types.HistoryResetWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.ResetWorkflowExecutionResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientResetWorkflowExecutionScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientResetWorkflowExecutionScope, metrics.CadenceClientLatency)
	resp, err := c.client.ResetWorkflowExecution(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientResetWorkflowExecutionScope, metrics.CadenceClientFailures)
	}

	return resp, err
}

func (c *metricClient) ScheduleDecisionTask(
	context context.Context,
	request *types.ScheduleDecisionTaskRequest,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.HistoryClientScheduleDecisionTaskScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientScheduleDecisionTaskScope, metrics.CadenceClientLatency)
	err := c.client.ScheduleDecisionTask(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientScheduleDecisionTaskScope, metrics.CadenceClientFailures)
	}

	return err
}

func (c *metricClient) RecordChildExecutionCompleted(
	context context.Context,
	request *types.RecordChildExecutionCompletedRequest,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.HistoryClientRecordChildExecutionCompletedScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRecordChildExecutionCompletedScope, metrics.CadenceClientLatency)
	err := c.client.RecordChildExecutionCompleted(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRecordChildExecutionCompletedScope, metrics.CadenceClientFailures)
	}

	return err
}

func (c *metricClient) ReplicateEventsV2(
	context context.Context,
	request *types.ReplicateEventsV2Request,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.HistoryClientReplicateEventsV2Scope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientReplicateEventsV2Scope, metrics.CadenceClientLatency)
	err := c.client.ReplicateEventsV2(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientReplicateEventsV2Scope, metrics.CadenceClientFailures)
	}

	return err
}

func (c *metricClient) SyncShardStatus(
	context context.Context,
	request *types.SyncShardStatusRequest,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.HistoryClientSyncShardStatusScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientSyncShardStatusScope, metrics.CadenceClientLatency)
	err := c.client.SyncShardStatus(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientSyncShardStatusScope, metrics.CadenceClientFailures)
	}

	return err
}

func (c *metricClient) SyncActivity(
	context context.Context,
	request *types.SyncActivityRequest,
	opts ...yarpc.CallOption,
) error {
	c.metricsClient.IncCounter(metrics.HistoryClientSyncActivityScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientSyncActivityScope, metrics.CadenceClientLatency)
	err := c.client.SyncActivity(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientSyncActivityScope, metrics.CadenceClientFailures)
	}

	return err
}

func (c *metricClient) GetReplicationMessages(
	ctx context.Context,
	request *types.GetReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.GetReplicationMessagesResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientGetReplicationTasksScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientGetReplicationTasksScope, metrics.CadenceClientLatency)
	resp, err := c.client.GetReplicationMessages(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientGetReplicationTasksScope, metrics.CadenceClientFailures)
	}

	return resp, err
}

func (c *metricClient) GetDLQReplicationMessages(
	ctx context.Context,
	request *types.GetDLQReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.GetDLQReplicationMessagesResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientGetDLQReplicationTasksScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientGetDLQReplicationTasksScope, metrics.CadenceClientLatency)
	resp, err := c.client.GetDLQReplicationMessages(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientGetDLQReplicationTasksScope, metrics.CadenceClientFailures)
	}

	return resp, err
}

func (c *metricClient) QueryWorkflow(
	ctx context.Context,
	request *types.HistoryQueryWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.HistoryQueryWorkflowResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientQueryWorkflowScope, metrics.CadenceClientRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientQueryWorkflowScope, metrics.CadenceClientLatency)
	resp, err := c.client.QueryWorkflow(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientQueryWorkflowScope, metrics.CadenceClientFailures)
	}

	return resp, err
}

func (c *metricClient) ReapplyEvents(
	ctx context.Context,
	request *types.HistoryReapplyEventsRequest,
	opts ...yarpc.CallOption,
) error {

	c.metricsClient.IncCounter(metrics.HistoryClientReapplyEventsScope, metrics.CadenceClientRequests)
	sw := c.metricsClient.StartTimer(metrics.HistoryClientReapplyEventsScope, metrics.CadenceClientLatency)
	err := c.client.ReapplyEvents(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientReapplyEventsScope, metrics.CadenceClientFailures)
	}
	return err
}

func (c *metricClient) ReadDLQMessages(
	ctx context.Context,
	request *types.ReadDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.ReadDLQMessagesResponse, error) {

	c.metricsClient.IncCounter(metrics.HistoryClientReadDLQMessagesScope, metrics.CadenceClientRequests)
	sw := c.metricsClient.StartTimer(metrics.HistoryClientReadDLQMessagesScope, metrics.CadenceClientLatency)
	resp, err := c.client.ReadDLQMessages(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientReadDLQMessagesScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) PurgeDLQMessages(
	ctx context.Context,
	request *types.PurgeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) error {

	c.metricsClient.IncCounter(metrics.HistoryClientPurgeDLQMessagesScope, metrics.CadenceClientRequests)
	sw := c.metricsClient.StartTimer(metrics.HistoryClientPurgeDLQMessagesScope, metrics.CadenceClientLatency)
	err := c.client.PurgeDLQMessages(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientPurgeDLQMessagesScope, metrics.CadenceClientFailures)
	}
	return err
}

func (c *metricClient) MergeDLQMessages(
	ctx context.Context,
	request *types.MergeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (*types.MergeDLQMessagesResponse, error) {

	c.metricsClient.IncCounter(metrics.HistoryClientMergeDLQMessagesScope, metrics.CadenceClientRequests)
	sw := c.metricsClient.StartTimer(metrics.HistoryClientMergeDLQMessagesScope, metrics.CadenceClientLatency)
	resp, err := c.client.MergeDLQMessages(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientMergeDLQMessagesScope, metrics.CadenceClientFailures)
	}
	return resp, err
}

func (c *metricClient) RefreshWorkflowTasks(
	ctx context.Context,
	request *types.HistoryRefreshWorkflowTasksRequest,
	opts ...yarpc.CallOption,
) error {

	c.metricsClient.IncCounter(metrics.HistoryClientRefreshWorkflowTasksScope, metrics.CadenceClientRequests)
	sw := c.metricsClient.StartTimer(metrics.HistoryClientRefreshWorkflowTasksScope, metrics.CadenceClientLatency)
	err := c.client.RefreshWorkflowTasks(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRefreshWorkflowTasksScope, metrics.CadenceClientFailures)
	}
	return err
}

func (c *metricClient) NotifyFailoverMarkers(
	ctx context.Context,
	request *types.NotifyFailoverMarkersRequest,
	opts ...yarpc.CallOption,
) error {

	c.metricsClient.IncCounter(metrics.HistoryClientNotifyFailoverMarkersScope, metrics.CadenceClientRequests)
	sw := c.metricsClient.StartTimer(metrics.HistoryClientNotifyFailoverMarkersScope, metrics.CadenceClientLatency)
	err := c.client.NotifyFailoverMarkers(ctx, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientNotifyFailoverMarkersScope, metrics.CadenceClientFailures)
	}
	return err
}
