// Copyright (c) 2021 Uber Technologies, Inc.
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

	matchingv1 "github.com/uber/cadence/.gen/proto/matching/v1"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/proto"
)

type grpcClient struct {
	c matchingv1.MatchingAPIYARPCClient
}

func NewGRPCClient(c matchingv1.MatchingAPIYARPCClient) Client {
	return grpcClient{c}
}

func (g grpcClient) AddActivityTask(ctx context.Context, request *types.AddActivityTaskRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.AddActivityTask(ctx, proto.FromMatchingAddActivityTaskRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) AddDecisionTask(ctx context.Context, request *types.AddDecisionTaskRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.AddDecisionTask(ctx, proto.FromMatchingAddDecisionTaskRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) CancelOutstandingPoll(ctx context.Context, request *types.CancelOutstandingPollRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.CancelOutstandingPoll(ctx, proto.FromMatchingCancelOutstandingPollRequest(request), opts...)
	return proto.ToError(err)
}

func (g grpcClient) DescribeTaskList(ctx context.Context, request *types.MatchingDescribeTaskListRequest, opts ...yarpc.CallOption) (*types.DescribeTaskListResponse, error) {
	response, err := g.c.DescribeTaskList(ctx, proto.FromMatchingDescribeTaskListRequest(request), opts...)
	return proto.ToMatchingDescribeTaskListResponse(response), proto.ToError(err)
}

func (g grpcClient) ListTaskListPartitions(ctx context.Context, request *types.MatchingListTaskListPartitionsRequest, opts ...yarpc.CallOption) (*types.ListTaskListPartitionsResponse, error) {
	response, err := g.c.ListTaskListPartitions(ctx, proto.FromMatchingListTaskListPartitionsRequest(request), opts...)
	return proto.ToMatchingListTaskListPartitionsResponse(response), proto.ToError(err)
}

func (g grpcClient) PollForActivityTask(ctx context.Context, request *types.MatchingPollForActivityTaskRequest, opts ...yarpc.CallOption) (*types.PollForActivityTaskResponse, error) {
	response, err := g.c.PollForActivityTask(ctx, proto.FromMatchingPollForActivityTaskRequest(request), opts...)
	return proto.ToMatchingPollForActivityTaskResponse(response), proto.ToError(err)
}

func (g grpcClient) PollForDecisionTask(ctx context.Context, request *types.MatchingPollForDecisionTaskRequest, opts ...yarpc.CallOption) (*types.MatchingPollForDecisionTaskResponse, error) {
	response, err := g.c.PollForDecisionTask(ctx, proto.FromMatchingPollForDecisionTaskRequest(request), opts...)
	return proto.ToMatchingPollForDecisionTaskResponse(response), proto.ToError(err)
}

func (g grpcClient) QueryWorkflow(ctx context.Context, request *types.MatchingQueryWorkflowRequest, opts ...yarpc.CallOption) (*types.QueryWorkflowResponse, error) {
	response, err := g.c.QueryWorkflow(ctx, proto.FromMatchingQueryWorkflowRequest(request), opts...)
	return proto.ToMatchingQueryWorkflowResponse(response), proto.ToError(err)
}

func (g grpcClient) RespondQueryTaskCompleted(ctx context.Context, request *types.MatchingRespondQueryTaskCompletedRequest, opts ...yarpc.CallOption) error {
	_, err := g.c.RespondQueryTaskCompleted(ctx, proto.FromMatchingRespondQueryTaskCompletedRequest(request), opts...)
	return proto.ToError(err)
}
