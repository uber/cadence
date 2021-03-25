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

	apiv1 "github.com/uber/cadence/.gen/proto/api/v1"
	matchingv1 "github.com/uber/cadence/.gen/proto/matching/v1"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types/mapper/proto"
)

type grpcHandler struct {
	h Handler
}

func newGRPCHandler(h Handler) grpcHandler {
	return grpcHandler{h}
}

func (g grpcHandler) register(dispatcher *yarpc.Dispatcher) {
	dispatcher.Register(matchingv1.BuildMatchingAPIYARPCProcedures(g))
	dispatcher.Register(apiv1.BuildMetaAPIYARPCProcedures(g))
}

func (g grpcHandler) Health(ctx context.Context, _ *apiv1.HealthRequest) (*apiv1.HealthResponse, error) {
	response, err := g.h.Health(withGRPCTag(ctx))
	return proto.FromHealthResponse(response), proto.FromError(err)
}

func (g grpcHandler) AddActivityTask(ctx context.Context, request *matchingv1.AddActivityTaskRequest) (*matchingv1.AddActivityTaskResponse, error) {
	err := g.h.AddActivityTask(withGRPCTag(ctx), proto.ToMatchingAddActivityTaskRequest(request))
	return &matchingv1.AddActivityTaskResponse{}, proto.FromError(err)
}

func (g grpcHandler) AddDecisionTask(ctx context.Context, request *matchingv1.AddDecisionTaskRequest) (*matchingv1.AddDecisionTaskResponse, error) {
	err := g.h.AddDecisionTask(withGRPCTag(ctx), proto.ToMatchingAddDecisionTaskRequest(request))
	return &matchingv1.AddDecisionTaskResponse{}, proto.FromError(err)
}

func (g grpcHandler) CancelOutstandingPoll(ctx context.Context, request *matchingv1.CancelOutstandingPollRequest) (*matchingv1.CancelOutstandingPollResponse, error) {
	err := g.h.CancelOutstandingPoll(withGRPCTag(ctx), proto.ToMatchingCancelOutstandingPollRequest(request))
	return &matchingv1.CancelOutstandingPollResponse{}, proto.FromError(err)
}

func (g grpcHandler) DescribeTaskList(ctx context.Context, request *matchingv1.DescribeTaskListRequest) (*matchingv1.DescribeTaskListResponse, error) {
	response, err := g.h.DescribeTaskList(withGRPCTag(ctx), proto.ToMatchingDescribeTaskListRequest(request))
	return proto.FromMatchingDescribeTaskListResponse(response), proto.FromError(err)
}

func (g grpcHandler) ListTaskListPartitions(ctx context.Context, request *matchingv1.ListTaskListPartitionsRequest) (*matchingv1.ListTaskListPartitionsResponse, error) {
	response, err := g.h.ListTaskListPartitions(withGRPCTag(ctx), proto.ToMatchingListTaskListPartitionsRequest(request))
	return proto.FromMatchingListTaskListPartitionsResponse(response), proto.FromError(err)
}

func (g grpcHandler) PollForActivityTask(ctx context.Context, request *matchingv1.PollForActivityTaskRequest) (*matchingv1.PollForActivityTaskResponse, error) {
	response, err := g.h.PollForActivityTask(withGRPCTag(ctx), proto.ToMatchingPollForActivityTaskRequest(request))
	return proto.FromMatchingPollForActivityTaskResponse(response), proto.FromError(err)
}

func (g grpcHandler) PollForDecisionTask(ctx context.Context, request *matchingv1.PollForDecisionTaskRequest) (*matchingv1.PollForDecisionTaskResponse, error) {
	response, err := g.h.PollForDecisionTask(withGRPCTag(ctx), proto.ToMatchingPollForDecisionTaskRequest(request))
	return proto.FromMatchingPollForDecisionTaskResponse(response), proto.FromError(err)
}

func (g grpcHandler) QueryWorkflow(ctx context.Context, request *matchingv1.QueryWorkflowRequest) (*matchingv1.QueryWorkflowResponse, error) {
	response, err := g.h.QueryWorkflow(withGRPCTag(ctx), proto.ToMatchingQueryWorkflowRequest(request))
	return proto.FromMatchingQueryWorkflowResponse(response), proto.FromError(err)
}

func (g grpcHandler) RespondQueryTaskCompleted(ctx context.Context, request *matchingv1.RespondQueryTaskCompletedRequest) (*matchingv1.RespondQueryTaskCompletedResponse, error) {
	err := g.h.RespondQueryTaskCompleted(withGRPCTag(ctx), proto.ToMatchingRespondQueryTaskCompletedRequest(request))
	return &matchingv1.RespondQueryTaskCompletedResponse{}, proto.FromError(err)
}

func withGRPCTag(ctx context.Context) context.Context {
	return metrics.TagContext(ctx, metrics.GPRCTransportTag())
}
