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

	h "github.com/uber/cadence/.gen/go/history"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/metrics"
	"go.uber.org/yarpc"
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
	request *h.StartWorkflowExecutionRequest,
	opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientStartWorkflowExecutionScope, metrics.CadenceRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientStartWorkflowExecutionScope, metrics.CadenceLatency)
	resp, err := c.client.StartWorkflowExecution(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientStartWorkflowExecutionScope, metrics.CadenceFailures)
	}

	return resp, err
}

func (c *metricClient) GetMutableState(
	context context.Context,
	request *h.GetMutableStateRequest,
	opts ...yarpc.CallOption) (*h.GetMutableStateResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientGetMutableStateScope, metrics.CadenceRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientGetMutableStateScope, metrics.CadenceLatency)
	resp, err := c.client.GetMutableState(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientGetMutableStateScope, metrics.CadenceFailures)
	}

	return resp, err
}

func (c *metricClient) DescribeWorkflowExecution(
	context context.Context,
	request *h.DescribeWorkflowExecutionRequest,
	opts ...yarpc.CallOption) (*shared.DescribeWorkflowExecutionResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientDescribeWorkflowExecutionScope, metrics.CadenceRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientDescribeWorkflowExecutionScope, metrics.CadenceLatency)
	resp, err := c.client.DescribeWorkflowExecution(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientDescribeWorkflowExecutionScope, metrics.CadenceFailures)
	}

	return resp, err
}

func (c *metricClient) RecordDecisionTaskStarted(
	context context.Context,
	request *h.RecordDecisionTaskStartedRequest,
	opts ...yarpc.CallOption) (*h.RecordDecisionTaskStartedResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientRecordDecisionTaskStartedScope, metrics.CadenceRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRecordDecisionTaskStartedScope, metrics.CadenceLatency)
	resp, err := c.client.RecordDecisionTaskStarted(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRecordDecisionTaskStartedScope, metrics.CadenceFailures)
	}

	return resp, err
}

func (c *metricClient) RecordActivityTaskStarted(
	context context.Context,
	request *h.RecordActivityTaskStartedRequest,
	opts ...yarpc.CallOption) (*h.RecordActivityTaskStartedResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientRecordActivityTaskStartedScope, metrics.CadenceRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRecordActivityTaskStartedScope, metrics.CadenceLatency)
	resp, err := c.client.RecordActivityTaskStarted(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRecordActivityTaskStartedScope, metrics.CadenceFailures)
	}

	return resp, err
}

func (c *metricClient) RespondDecisionTaskCompleted(
	context context.Context,
	request *h.RespondDecisionTaskCompletedRequest,
	opts ...yarpc.CallOption) error {
	c.metricsClient.IncCounter(metrics.HistoryClientRespondDecisionTaskCompletedScope, metrics.CadenceRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRespondDecisionTaskCompletedScope, metrics.CadenceLatency)
	err := c.client.RespondDecisionTaskCompleted(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRespondDecisionTaskCompletedScope, metrics.CadenceFailures)
	}

	return err
}

func (c *metricClient) RespondDecisionTaskFailed(
	context context.Context,
	request *h.RespondDecisionTaskFailedRequest,
	opts ...yarpc.CallOption) error {
	c.metricsClient.IncCounter(metrics.HistoryClientRespondDecisionTaskFailedScope, metrics.CadenceRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRespondDecisionTaskFailedScope, metrics.CadenceLatency)
	err := c.client.RespondDecisionTaskFailed(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRespondDecisionTaskFailedScope, metrics.CadenceFailures)
	}

	return err
}

func (c *metricClient) RespondActivityTaskCompleted(
	context context.Context,
	request *h.RespondActivityTaskCompletedRequest,
	opts ...yarpc.CallOption) error {
	c.metricsClient.IncCounter(metrics.HistoryClientRespondActivityTaskCompletedScope, metrics.CadenceRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRespondActivityTaskCompletedScope, metrics.CadenceLatency)
	err := c.client.RespondActivityTaskCompleted(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRespondActivityTaskCompletedScope, metrics.CadenceFailures)
	}

	return err
}

func (c *metricClient) RespondActivityTaskFailed(
	context context.Context,
	request *h.RespondActivityTaskFailedRequest,
	opts ...yarpc.CallOption) error {
	c.metricsClient.IncCounter(metrics.HistoryClientRespondActivityTaskFailedScope, metrics.CadenceRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRespondActivityTaskFailedScope, metrics.CadenceLatency)
	err := c.client.RespondActivityTaskFailed(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRespondActivityTaskFailedScope, metrics.CadenceFailures)
	}

	return err
}

func (c *metricClient) RespondActivityTaskCanceled(
	context context.Context,
	request *h.RespondActivityTaskCanceledRequest,
	opts ...yarpc.CallOption) error {
	c.metricsClient.IncCounter(metrics.HistoryClientRespondActivityTaskCanceledScope, metrics.CadenceRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRespondActivityTaskCanceledScope, metrics.CadenceLatency)
	err := c.client.RespondActivityTaskCanceled(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRespondActivityTaskCanceledScope, metrics.CadenceFailures)
	}

	return err
}

func (c *metricClient) RecordActivityTaskHeartbeat(
	context context.Context,
	request *h.RecordActivityTaskHeartbeatRequest,
	opts ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	c.metricsClient.IncCounter(metrics.HistoryClientRecordActivityTaskHeartbeatScope, metrics.CadenceRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRecordActivityTaskHeartbeatScope, metrics.CadenceLatency)
	resp, err := c.client.RecordActivityTaskHeartbeat(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRecordActivityTaskHeartbeatScope, metrics.CadenceFailures)
	}

	return resp, err
}

func (c *metricClient) RequestCancelWorkflowExecution(
	context context.Context,
	request *h.RequestCancelWorkflowExecutionRequest,
	opts ...yarpc.CallOption) error {
	c.metricsClient.IncCounter(metrics.HistoryClientRequestCancelWorkflowExecutionScope, metrics.CadenceRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRequestCancelWorkflowExecutionScope, metrics.CadenceLatency)
	err := c.client.RequestCancelWorkflowExecution(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRequestCancelWorkflowExecutionScope, metrics.CadenceFailures)
	}

	return err
}

func (c *metricClient) SignalWorkflowExecution(
	context context.Context,
	request *h.SignalWorkflowExecutionRequest,
	opts ...yarpc.CallOption) error {
	c.metricsClient.IncCounter(metrics.HistoryClientSignalWorkflowExecutionScope, metrics.CadenceRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientSignalWorkflowExecutionScope, metrics.CadenceLatency)
	err := c.client.SignalWorkflowExecution(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientSignalWorkflowExecutionScope, metrics.CadenceFailures)
	}

	return err
}

func (c *metricClient) TerminateWorkflowExecution(
	context context.Context,
	request *h.TerminateWorkflowExecutionRequest,
	opts ...yarpc.CallOption) error {
	c.metricsClient.IncCounter(metrics.HistoryClientTerminateWorkflowExecutionScope, metrics.CadenceRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientTerminateWorkflowExecutionScope, metrics.CadenceLatency)
	err := c.client.TerminateWorkflowExecution(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientTerminateWorkflowExecutionScope, metrics.CadenceFailures)
	}

	return err
}

func (c *metricClient) ScheduleDecisionTask(
	context context.Context,
	request *h.ScheduleDecisionTaskRequest,
	opts ...yarpc.CallOption) error {
	c.metricsClient.IncCounter(metrics.HistoryClientScheduleDecisionTaskScope, metrics.CadenceRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientScheduleDecisionTaskScope, metrics.CadenceLatency)
	err := c.client.ScheduleDecisionTask(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientScheduleDecisionTaskScope, metrics.CadenceFailures)
	}

	return err
}

func (c *metricClient) RecordChildExecutionCompleted(
	context context.Context,
	request *h.RecordChildExecutionCompletedRequest,
	opts ...yarpc.CallOption) error {
	c.metricsClient.IncCounter(metrics.HistoryClientRecordChildExecutionCompletedScope, metrics.CadenceRequests)

	sw := c.metricsClient.StartTimer(metrics.HistoryClientRecordChildExecutionCompletedScope, metrics.CadenceLatency)
	err := c.client.RecordChildExecutionCompleted(context, request, opts...)
	sw.Stop()

	if err != nil {
		c.metricsClient.IncCounter(metrics.HistoryClientRecordChildExecutionCompletedScope, metrics.CadenceFailures)
	}

	return err
}
