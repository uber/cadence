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

	"github.com/uber/cadence/.gen/go/matching"
	"github.com/uber/cadence/.gen/go/matching/matchingserviceclient"
	"github.com/uber/cadence/.gen/go/shared"
)

type thriftClient struct {
	c matchingserviceclient.Interface
}

// NewThriftClient creates a new instance of Client with thrift protocol
func NewThriftClient(c matchingserviceclient.Interface) Client {
	return thriftClient{c}
}

func (t thriftClient) AddActivityTask(ctx context.Context, request *matching.AddActivityTaskRequest, opts ...yarpc.CallOption) error {
	return t.c.AddActivityTask(ctx, request, opts...)
}

func (t thriftClient) AddDecisionTask(ctx context.Context, request *matching.AddDecisionTaskRequest, opts ...yarpc.CallOption) error {
	return t.c.AddDecisionTask(ctx, request, opts...)
}

func (t thriftClient) CancelOutstandingPoll(ctx context.Context, request *matching.CancelOutstandingPollRequest, opts ...yarpc.CallOption) error {
	return t.c.CancelOutstandingPoll(ctx, request, opts...)
}

func (t thriftClient) DescribeTaskList(ctx context.Context, request *matching.DescribeTaskListRequest, opts ...yarpc.CallOption) (*shared.DescribeTaskListResponse, error) {
	return t.c.DescribeTaskList(ctx, request, opts...)
}

func (t thriftClient) ListTaskListPartitions(ctx context.Context, request *matching.ListTaskListPartitionsRequest, opts ...yarpc.CallOption) (*shared.ListTaskListPartitionsResponse, error) {
	return t.c.ListTaskListPartitions(ctx, request, opts...)
}

func (t thriftClient) PollForActivityTask(ctx context.Context, request *matching.PollForActivityTaskRequest, opts ...yarpc.CallOption) (*shared.PollForActivityTaskResponse, error) {
	return t.c.PollForActivityTask(ctx, request, opts...)
}

func (t thriftClient) PollForDecisionTask(ctx context.Context, request *matching.PollForDecisionTaskRequest, opts ...yarpc.CallOption) (*matching.PollForDecisionTaskResponse, error) {
	return t.c.PollForDecisionTask(ctx, request, opts...)
}

func (t thriftClient) QueryWorkflow(ctx context.Context, request *matching.QueryWorkflowRequest, opts ...yarpc.CallOption) (*shared.QueryWorkflowResponse, error) {
	return t.c.QueryWorkflow(ctx, request, opts...)
}

func (t thriftClient) RespondQueryTaskCompleted(ctx context.Context, request *matching.RespondQueryTaskCompletedRequest, opts ...yarpc.CallOption) error {
	return t.c.RespondQueryTaskCompleted(ctx, request, opts...)
}
