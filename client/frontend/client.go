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
	"time"

	"go.uber.org/yarpc"

	"github.com/uber/cadence/common/types"
)

var _ Client = (*clientImpl)(nil)

const (
	// DefaultTimeout is the default timeout used to make calls
	DefaultTimeout = 10 * time.Second
	// DefaultLongPollTimeout is the long poll default timeout used to make calls
	DefaultLongPollTimeout = time.Minute * 3
)

type clientImpl struct {
	timeout         time.Duration
	longPollTimeout time.Duration
	client          Client
}

// NewClient creates a new frontend service TChannel client
func NewClient(
	timeout time.Duration,
	longPollTimeout time.Duration,
	client Client,
) Client {
	return &clientImpl{
		timeout:         timeout,
		longPollTimeout: longPollTimeout,
		client:          client,
	}
}

func (c *clientImpl) DeprecateDomain(
	ctx context.Context,
	request *types.DeprecateDomainRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.DeprecateDomain(ctx, request, opts...)
}

func (c *clientImpl) DescribeDomain(
	ctx context.Context,
	request *types.DescribeDomainRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeDomainResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.DescribeDomain(ctx, request, opts...)
}

func (c *clientImpl) DescribeTaskList(
	ctx context.Context,
	request *types.DescribeTaskListRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeTaskListResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.DescribeTaskList(ctx, request, opts...)
}

func (c *clientImpl) DescribeWorkflowExecution(
	ctx context.Context,
	request *types.DescribeWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeWorkflowExecutionResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.DescribeWorkflowExecution(ctx, request, opts...)
}

func (c *clientImpl) GetWorkflowExecutionHistory(
	ctx context.Context,
	request *types.GetWorkflowExecutionHistoryRequest,
	opts ...yarpc.CallOption,
) (*types.GetWorkflowExecutionHistoryResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.GetWorkflowExecutionHistory(ctx, request, opts...)
}

func (c *clientImpl) ListArchivedWorkflowExecutions(
	ctx context.Context,
	request *types.ListArchivedWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListArchivedWorkflowExecutionsResponse, error) {

	ctx, cancel := c.createLongPollContext(ctx)
	defer cancel()
	return c.client.ListArchivedWorkflowExecutions(ctx, request, opts...)
}

func (c *clientImpl) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *types.ListClosedWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListClosedWorkflowExecutionsResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.ListClosedWorkflowExecutions(ctx, request, opts...)
}

func (c *clientImpl) ListDomains(
	ctx context.Context,
	request *types.ListDomainsRequest,
	opts ...yarpc.CallOption,
) (*types.ListDomainsResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.ListDomains(ctx, request, opts...)
}

func (c *clientImpl) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *types.ListOpenWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListOpenWorkflowExecutionsResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.ListOpenWorkflowExecutions(ctx, request, opts...)
}

func (c *clientImpl) ListWorkflowExecutions(
	ctx context.Context,
	request *types.ListWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListWorkflowExecutionsResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.ListWorkflowExecutions(ctx, request, opts...)
}

func (c *clientImpl) ScanWorkflowExecutions(
	ctx context.Context,
	request *types.ListWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListWorkflowExecutionsResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.ScanWorkflowExecutions(ctx, request, opts...)
}

func (c *clientImpl) CountWorkflowExecutions(
	ctx context.Context,
	request *types.CountWorkflowExecutionsRequest,
	opts ...yarpc.CallOption,
) (*types.CountWorkflowExecutionsResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.CountWorkflowExecutions(ctx, request, opts...)
}

func (c *clientImpl) GetSearchAttributes(
	ctx context.Context,
	opts ...yarpc.CallOption,
) (*types.GetSearchAttributesResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.GetSearchAttributes(ctx, opts...)
}

func (c *clientImpl) PollForActivityTask(
	ctx context.Context,
	request *types.PollForActivityTaskRequest,
	opts ...yarpc.CallOption,
) (*types.PollForActivityTaskResponse, error) {

	ctx, cancel := c.createLongPollContext(ctx)
	defer cancel()
	return c.client.PollForActivityTask(ctx, request, opts...)
}

func (c *clientImpl) PollForDecisionTask(
	ctx context.Context,
	request *types.PollForDecisionTaskRequest,
	opts ...yarpc.CallOption,
) (*types.PollForDecisionTaskResponse, error) {

	ctx, cancel := c.createLongPollContext(ctx)
	defer cancel()
	return c.client.PollForDecisionTask(ctx, request, opts...)
}

func (c *clientImpl) QueryWorkflow(
	ctx context.Context,
	request *types.QueryWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.QueryWorkflowResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.QueryWorkflow(ctx, request, opts...)
}

func (c *clientImpl) RecordActivityTaskHeartbeat(
	ctx context.Context,
	request *types.RecordActivityTaskHeartbeatRequest,
	opts ...yarpc.CallOption,
) (*types.RecordActivityTaskHeartbeatResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RecordActivityTaskHeartbeat(ctx, request, opts...)
}

func (c *clientImpl) RecordActivityTaskHeartbeatByID(
	ctx context.Context,
	request *types.RecordActivityTaskHeartbeatByIDRequest,
	opts ...yarpc.CallOption,
) (*types.RecordActivityTaskHeartbeatResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RecordActivityTaskHeartbeatByID(ctx, request, opts...)
}

func (c *clientImpl) RegisterDomain(
	ctx context.Context,
	request *types.RegisterDomainRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RegisterDomain(ctx, request, opts...)
}

func (c *clientImpl) RequestCancelWorkflowExecution(
	ctx context.Context,
	request *types.RequestCancelWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RequestCancelWorkflowExecution(ctx, request, opts...)
}

func (c *clientImpl) ResetStickyTaskList(
	ctx context.Context,
	request *types.ResetStickyTaskListRequest,
	opts ...yarpc.CallOption,
) (*types.ResetStickyTaskListResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.ResetStickyTaskList(ctx, request, opts...)
}

func (c *clientImpl) ResetWorkflowExecution(
	ctx context.Context,
	request *types.ResetWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.ResetWorkflowExecutionResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.ResetWorkflowExecution(ctx, request, opts...)
}

func (c *clientImpl) RespondActivityTaskCanceled(
	ctx context.Context,
	request *types.RespondActivityTaskCanceledRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RespondActivityTaskCanceled(ctx, request, opts...)
}

func (c *clientImpl) RespondActivityTaskCanceledByID(
	ctx context.Context,
	request *types.RespondActivityTaskCanceledByIDRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RespondActivityTaskCanceledByID(ctx, request, opts...)
}

func (c *clientImpl) RespondActivityTaskCompleted(
	ctx context.Context,
	request *types.RespondActivityTaskCompletedRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RespondActivityTaskCompleted(ctx, request, opts...)
}

func (c *clientImpl) RespondActivityTaskCompletedByID(
	ctx context.Context,
	request *types.RespondActivityTaskCompletedByIDRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RespondActivityTaskCompletedByID(ctx, request, opts...)
}

func (c *clientImpl) RespondActivityTaskFailed(
	ctx context.Context,
	request *types.RespondActivityTaskFailedRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RespondActivityTaskFailed(ctx, request, opts...)
}

func (c *clientImpl) RespondActivityTaskFailedByID(
	ctx context.Context,
	request *types.RespondActivityTaskFailedByIDRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RespondActivityTaskFailedByID(ctx, request, opts...)
}

func (c *clientImpl) RespondDecisionTaskCompleted(
	ctx context.Context,
	request *types.RespondDecisionTaskCompletedRequest,
	opts ...yarpc.CallOption,
) (*types.RespondDecisionTaskCompletedResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RespondDecisionTaskCompleted(ctx, request, opts...)
}

func (c *clientImpl) RespondDecisionTaskFailed(
	ctx context.Context,
	request *types.RespondDecisionTaskFailedRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RespondDecisionTaskFailed(ctx, request, opts...)
}

func (c *clientImpl) RespondQueryTaskCompleted(
	ctx context.Context,
	request *types.RespondQueryTaskCompletedRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.RespondQueryTaskCompleted(ctx, request, opts...)
}

func (c *clientImpl) SignalWithStartWorkflowExecution(
	ctx context.Context,
	request *types.SignalWithStartWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.StartWorkflowExecutionResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.SignalWithStartWorkflowExecution(ctx, request, opts...)
}

func (c *clientImpl) SignalWorkflowExecution(
	ctx context.Context,
	request *types.SignalWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.SignalWorkflowExecution(ctx, request, opts...)
}

func (c *clientImpl) StartWorkflowExecution(
	ctx context.Context,
	request *types.StartWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (*types.StartWorkflowExecutionResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.StartWorkflowExecution(ctx, request, opts...)
}

func (c *clientImpl) TerminateWorkflowExecution(
	ctx context.Context,
	request *types.TerminateWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) error {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.TerminateWorkflowExecution(ctx, request, opts...)
}

func (c *clientImpl) UpdateDomain(
	ctx context.Context,
	request *types.UpdateDomainRequest,
	opts ...yarpc.CallOption,
) (*types.UpdateDomainResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.UpdateDomain(ctx, request, opts...)
}

func (c *clientImpl) createContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		return context.WithTimeout(context.Background(), c.timeout)
	}
	return context.WithTimeout(parent, c.timeout)
}

func (c *clientImpl) createLongPollContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		return context.WithTimeout(context.Background(), c.longPollTimeout)
	}
	return context.WithTimeout(parent, c.longPollTimeout)
}

func (c *clientImpl) GetClusterInfo(
	ctx context.Context,
	opts ...yarpc.CallOption,
) (*types.ClusterInfo, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()
	return c.client.GetClusterInfo(ctx, opts...)
}

func (c *clientImpl) ListTaskListPartitions(
	ctx context.Context,
	request *types.ListTaskListPartitionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListTaskListPartitionsResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()

	return c.client.ListTaskListPartitions(ctx, request, opts...)
}

func (c *clientImpl) GetTaskListsByDomain(
	ctx context.Context,
	request *types.GetTaskListsByDomainRequest,
	opts ...yarpc.CallOption,
) (*types.GetTaskListsByDomainResponse, error) {

	ctx, cancel := c.createContext(ctx)
	defer cancel()

	return c.client.GetTaskListsByDomain(ctx, request, opts...)
}
