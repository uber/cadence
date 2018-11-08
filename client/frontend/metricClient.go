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

package frontend

import (
	"context"

	"github.com/uber/cadence/common/metrics"
	"go.uber.org/cadence/.gen/go/shared"
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

func (c *metricClient) DeprecateDomain(
	ctx context.Context,
	request *shared.DeprecateDomainRequest,
	opts ...yarpc.CallOption,
) error {
	return nil
	// c.metricsClient.IncCounter(metrics.FrontendClient)
}

func (c *metricClient) DescribeDomain(
	ctx context.Context,
	request *shared.DescribeDomainRequest,
	opts ...yarpc.CallOption,
) (*shared.DescribeDomainResponse, error) {
	return nil, nil
}

func (c *metricClient) DescribeTaskList(
	ctx context.Context,
	request *shared.DescribeTaskListRequest,
	opts ...yarpc.CallOption,
) (*shared.DescribeTaskListResponse, error) {
	return nil, nil
}

func (c *metricClient) DescribeWorkflowExecution(
	ctx context.Context,
	request *shared.DescribeWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*shared.DescribeWorkflowExecutionResponse, error) {
	return nil, nil
}

func (c *metricClient) GetWorkflowExecutionHistory(
	ctx context.Context,
	request *shared.GetWorkflowExecutionHistoryRequest,
	opts ...yarpc.CallOption,
) (*shared.GetWorkflowExecutionHistoryResponse, error) {
	return nil, nil
}

func (c *metricClient) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *shared.ListClosedWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*shared.ListClosedWorkflowExecutionsResponse, error) {
	return nil, nil
}

func (c *metricClient) ListDomains(
	ctx context.Context,
	request *shared.ListDomainsRequest,
	opts ...yarpc.CallOption,
) (*shared.ListDomainsResponse, error) {
	return nil, nil
}

func (c *metricClient) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *shared.ListOpenWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*shared.ListOpenWorkflowExecutionsResponse, error) {
	return nil, nil
}

func (c *metricClient) PollForActivityTask(
	ctx context.Context,
	request *shared.PollForActivityTaskRequest,
	opts ...yarpc.CallOption,
) (*shared.PollForActivityTaskResponse, error) {
	return nil, nil
}

func (c *metricClient) PollForDecisionTask(
	ctx context.Context,
	request *shared.PollForDecisionTaskRequest,
	opts ...yarpc.CallOption,
) (*shared.PollForDecisionTaskResponse, error) {
	return nil, nil
}

func (c *metricClient) QueryWorkflow(
	ctx context.Context,
	request *shared.QueryWorkflowRequest,
	opts ...yarpc.CallOption,
) (*shared.QueryWorkflowResponse, error) {
	return nil, nil
}

func (c *metricClient) RecordActivityTaskHeartbeat(
	ctx context.Context,
	request *shared.RecordActivityTaskHeartbeatRequest,
	opts ...yarpc.CallOption,
) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	return nil, nil
}

func (c *metricClient) RecordActivityTaskHeartbeatByID(
	ctx context.Context,
	request *shared.RecordActivityTaskHeartbeatByIDRequest,
	opts ...yarpc.CallOption,
) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	return nil, nil
}

func (c *metricClient) RegisterDomain(
	ctx context.Context,
	request *shared.RegisterDomainRequest,
	opts ...yarpc.CallOption,
) error {
	return nil
}

func (c *metricClient) RequestCancelWorkflowExecution(
	ctx context.Context,
	request *shared.RequestCancelWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {
	return nil
}

func (c *metricClient) ResetStickyTaskList(
	ctx context.Context,
	request *shared.ResetStickyTaskListRequest,
	opts ...yarpc.CallOption,
) (*shared.ResetStickyTaskListResponse, error) {
	return nil, nil
}

func (c *metricClient) RespondActivityTaskCanceled(
	ctx context.Context,
	request *shared.RespondActivityTaskCanceledRequest,
	opts ...yarpc.CallOption,
) error {
	return nil
}

func (c *metricClient) RespondActivityTaskCanceledByID(
	ctx context.Context,
	request *shared.RespondActivityTaskCanceledByIDRequest,
	opts ...yarpc.CallOption,
) error {
	return nil
}

func (c *metricClient) RespondActivityTaskCompleted(
	ctx context.Context,
	request *shared.RespondActivityTaskCompletedRequest,
	opts ...yarpc.CallOption,
) error {
	return nil
}

func (c *metricClient) RespondActivityTaskCompletedByID(
	ctx context.Context,
	request *shared.RespondActivityTaskCompletedByIDRequest,
	opts ...yarpc.CallOption,
) error {
	return nil
}

func (c *metricClient) RespondActivityTaskFailed(
	ctx context.Context,
	request *shared.RespondActivityTaskFailedRequest,
	opts ...yarpc.CallOption,
) error {
	return nil
}

func (c *metricClient) RespondActivityTaskFailedByID(
	ctx context.Context,
	request *shared.RespondActivityTaskFailedByIDRequest,
	opts ...yarpc.CallOption,
) error {
	return nil
}

func (c *metricClient) RespondDecisionTaskCompleted(
	ctx context.Context,
	request *shared.RespondDecisionTaskCompletedRequest,
	opts ...yarpc.CallOption,
) (*shared.RespondDecisionTaskCompletedResponse, error) {
	return nil, nil
}

func (c *metricClient) RespondDecisionTaskFailed(
	ctx context.Context,
	request *shared.RespondDecisionTaskFailedRequest,
	opts ...yarpc.CallOption,
) error {
	return nil
}

func (c *metricClient) RespondQueryTaskCompleted(
	ctx context.Context,
	request *shared.RespondQueryTaskCompletedRequest,
	opts ...yarpc.CallOption,
) error {
	return nil
}

func (c *metricClient) SignalWithStartWorkflowExecution(
	ctx context.Context,
	request *shared.SignalWithStartWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*shared.StartWorkflowExecutionResponse, error) {
	return nil, nil
}

func (c *metricClient) SignalWorkflowExecution(
	ctx context.Context,
	request *shared.SignalWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {
	return nil
}

func (c *metricClient) StartWorkflowExecution(
	ctx context.Context,
	request *shared.StartWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*shared.StartWorkflowExecutionResponse, error) {
	return nil, nil
}

func (c *metricClient) TerminateWorkflowExecution(
	ctx context.Context,
	request *shared.TerminateWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {
	return nil
}

func (c *metricClient) UpdateDomain(
	ctx context.Context,
	request *shared.UpdateDomainRequest,
	opts ...yarpc.CallOption,
) (*shared.UpdateDomainResponse, error) {
	return nil, nil
}

//func (c *metricClient) AddActivityTask(
//	ctx context.Context,
//	addRequest *m.AddActivityTaskRequest,
//	opts ...yarpc.CallOption) error {
//	c.metricsClient.IncCounter(metrics.MatchingClientAddActivityTaskScope, metrics.CadenceClientRequests)
//
//	sw := c.metricsClient.StartTimer(metrics.MatchingClientAddActivityTaskScope, metrics.CadenceClientLatency)
//	err := c.client.AddActivityTask(ctx, addRequest, opts...)
//	sw.Stop()
//
//	if err != nil {
//		c.metricsClient.IncCounter(metrics.MatchingClientAddActivityTaskScope, metrics.CadenceClientFailures)
//	}
//
//	return err
//}
