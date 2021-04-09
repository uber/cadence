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

	"github.com/uber/cadence/.gen/go/health"
	"github.com/uber/cadence/.gen/go/health/metaserver"
	m "github.com/uber/cadence/.gen/go/matching"
	"github.com/uber/cadence/.gen/go/matching/matchingserviceserver"
	s "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types/mapper/thrift"
)

// ThriftHandler wrap underlying handler and handles Thrift related type conversions
type ThriftHandler struct {
	h Handler
}

// NewThriftHandler creates Thrift handler on top of underlying handler
func NewThriftHandler(h Handler) ThriftHandler {
	return ThriftHandler{h}
}

func (t ThriftHandler) register(dispatcher *yarpc.Dispatcher) {
	dispatcher.Register(matchingserviceserver.New(t))
	dispatcher.Register(metaserver.New(t))
}

// Health forwards request to the underlying handler
func (t ThriftHandler) Health(ctx context.Context) (*health.HealthStatus, error) {
	response, err := t.h.Health(withThriftTag(ctx))
	return thrift.FromHealthStatus(response), thrift.FromError(err)
}

// AddActivityTask forwards request to the underlying handler
func (t ThriftHandler) AddActivityTask(ctx context.Context, request *m.AddActivityTaskRequest) error {
	err := t.h.AddActivityTask(withThriftTag(ctx), thrift.ToAddActivityTaskRequest(request))
	return thrift.FromError(err)
}

// AddDecisionTask forwards request to the underlying handler
func (t ThriftHandler) AddDecisionTask(ctx context.Context, request *m.AddDecisionTaskRequest) error {
	err := t.h.AddDecisionTask(withThriftTag(ctx), thrift.ToAddDecisionTaskRequest(request))
	return thrift.FromError(err)
}

// CancelOutstandingPoll forwards request to the underlying handler
func (t ThriftHandler) CancelOutstandingPoll(ctx context.Context, request *m.CancelOutstandingPollRequest) error {
	err := t.h.CancelOutstandingPoll(withThriftTag(ctx), thrift.ToCancelOutstandingPollRequest(request))
	return thrift.FromError(err)
}

// DescribeTaskList forwards request to the underlying handler
func (t ThriftHandler) DescribeTaskList(ctx context.Context, request *m.DescribeTaskListRequest) (*s.DescribeTaskListResponse, error) {
	response, err := t.h.DescribeTaskList(withThriftTag(ctx), thrift.ToMatchingDescribeTaskListRequest(request))
	return thrift.FromDescribeTaskListResponse(response), thrift.FromError(err)
}

// ListTaskListPartitions forwards request to the underlying handler
func (t ThriftHandler) ListTaskListPartitions(ctx context.Context, request *m.ListTaskListPartitionsRequest) (*s.ListTaskListPartitionsResponse, error) {
	response, err := t.h.ListTaskListPartitions(withThriftTag(ctx), thrift.ToMatchingListTaskListPartitionsRequest(request))
	return thrift.FromListTaskListPartitionsResponse(response), thrift.FromError(err)
}

// PollForActivityTask forwards request to the underlying handler
func (t ThriftHandler) PollForActivityTask(ctx context.Context, request *m.PollForActivityTaskRequest) (*s.PollForActivityTaskResponse, error) {
	response, err := t.h.PollForActivityTask(withThriftTag(ctx), thrift.ToMatchingPollForActivityTaskRequest(request))
	return thrift.FromPollForActivityTaskResponse(response), thrift.FromError(err)
}

// PollForDecisionTask forwards request to the underlying handler
func (t ThriftHandler) PollForDecisionTask(ctx context.Context, request *m.PollForDecisionTaskRequest) (*m.PollForDecisionTaskResponse, error) {
	response, err := t.h.PollForDecisionTask(withThriftTag(ctx), thrift.ToMatchingPollForDecisionTaskRequest(request))
	return thrift.FromMatchingPollForDecisionTaskResponse(response), thrift.FromError(err)
}

// QueryWorkflow forwards request to the underlying handler
func (t ThriftHandler) QueryWorkflow(ctx context.Context, request *m.QueryWorkflowRequest) (*s.QueryWorkflowResponse, error) {
	response, err := t.h.QueryWorkflow(withThriftTag(ctx), thrift.ToMatchingQueryWorkflowRequest(request))
	return thrift.FromQueryWorkflowResponse(response), thrift.FromError(err)
}

// RespondQueryTaskCompleted forwards request to the underlying handler
func (t ThriftHandler) RespondQueryTaskCompleted(ctx context.Context, request *m.RespondQueryTaskCompletedRequest) error {
	err := t.h.RespondQueryTaskCompleted(withThriftTag(ctx), thrift.ToMatchingRespondQueryTaskCompletedRequest(request))
	return thrift.FromError(err)
}

func withThriftTag(ctx context.Context) context.Context {
	return metrics.TagContext(ctx, metrics.ThriftTransportTag())
}
