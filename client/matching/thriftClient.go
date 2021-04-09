// Copyright (c) 2020 Uber Technologies, Inc.
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

package matching

import (
	"context"

	"go.uber.org/yarpc"

	"github.com/uber/cadence/.gen/go/matching/matchingserviceclient"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"
)

type thriftClient struct {
	c matchingserviceclient.Interface
}

// NewThriftClient creates a new instance of Client with thrift protocol
func NewThriftClient(c matchingserviceclient.Interface) Client {
	return thriftClient{c}
}

func (t thriftClient) AddActivityTask(
	ctx context.Context,
	request *types.AddActivityTaskRequest,
	opts ...yarpc.CallOption,
) error {
	err := t.c.AddActivityTask(ctx, thrift.FromAddActivityTaskRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) AddDecisionTask(
	ctx context.Context,
	request *types.AddDecisionTaskRequest,
	opts ...yarpc.CallOption,
) error {
	err := t.c.AddDecisionTask(ctx, thrift.FromAddDecisionTaskRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) CancelOutstandingPoll(
	ctx context.Context,
	request *types.CancelOutstandingPollRequest,
	opts ...yarpc.CallOption,
) error {
	err := t.c.CancelOutstandingPoll(ctx, thrift.FromCancelOutstandingPollRequest(request), opts...)
	return thrift.ToError(err)
}

func (t thriftClient) DescribeTaskList(
	ctx context.Context,
	request *types.MatchingDescribeTaskListRequest,
	opts ...yarpc.CallOption,
) (*types.DescribeTaskListResponse, error) {
	response, err := t.c.DescribeTaskList(ctx, thrift.FromMatchingDescribeTaskListRequest(request), opts...)
	return thrift.ToDescribeTaskListResponse(response), thrift.ToError(err)
}

func (t thriftClient) ListTaskListPartitions(
	ctx context.Context,
	request *types.MatchingListTaskListPartitionsRequest,
	opts ...yarpc.CallOption,
) (*types.ListTaskListPartitionsResponse, error) {
	response, err := t.c.ListTaskListPartitions(ctx, thrift.FromMatchingListTaskListPartitionsRequest(request), opts...)
	return thrift.ToListTaskListPartitionsResponse(response), thrift.ToError(err)
}

func (t thriftClient) PollForActivityTask(
	ctx context.Context,
	request *types.MatchingPollForActivityTaskRequest,
	opts ...yarpc.CallOption,
) (*types.PollForActivityTaskResponse, error) {
	response, err := t.c.PollForActivityTask(ctx, thrift.FromMatchingPollForActivityTaskRequest(request), opts...)
	return thrift.ToPollForActivityTaskResponse(response), thrift.ToError(err)
}

func (t thriftClient) PollForDecisionTask(
	ctx context.Context,
	request *types.MatchingPollForDecisionTaskRequest,
	opts ...yarpc.CallOption,
) (*types.MatchingPollForDecisionTaskResponse, error) {
	response, err := t.c.PollForDecisionTask(ctx, thrift.FromMatchingPollForDecisionTaskRequest(request), opts...)
	return thrift.ToMatchingPollForDecisionTaskResponse(response), thrift.ToError(err)
}

func (t thriftClient) QueryWorkflow(
	ctx context.Context,
	request *types.MatchingQueryWorkflowRequest,
	opts ...yarpc.CallOption,
) (*types.QueryWorkflowResponse, error) {
	response, err := t.c.QueryWorkflow(ctx, thrift.FromMatchingQueryWorkflowRequest(request), opts...)
	return thrift.ToQueryWorkflowResponse(response), thrift.ToError(err)
}

func (t thriftClient) RespondQueryTaskCompleted(
	ctx context.Context,
	request *types.MatchingRespondQueryTaskCompletedRequest,
	opts ...yarpc.CallOption,
) error {
	err := t.c.RespondQueryTaskCompleted(ctx, thrift.FromMatchingRespondQueryTaskCompletedRequest(request), opts...)
	return thrift.ToError(err)
}
