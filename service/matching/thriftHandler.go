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
	return t.h.Health(ctx)
}

// AddActivityTask forwards request to the underlying handler
func (t ThriftHandler) AddActivityTask(ctx context.Context, request *m.AddActivityTaskRequest) (err error) {
	return t.h.AddActivityTask(ctx, request)
}

// AddDecisionTask forwards request to the underlying handler
func (t ThriftHandler) AddDecisionTask(ctx context.Context, request *m.AddDecisionTaskRequest) (err error) {
	return t.h.AddDecisionTask(ctx, request)
}

// CancelOutstandingPoll forwards request to the underlying handler
func (t ThriftHandler) CancelOutstandingPoll(ctx context.Context, request *m.CancelOutstandingPollRequest) (err error) {
	return t.h.CancelOutstandingPoll(ctx, request)
}

// DescribeTaskList forwards request to the underlying handler
func (t ThriftHandler) DescribeTaskList(ctx context.Context, request *m.DescribeTaskListRequest) (response *s.DescribeTaskListResponse, err error) {
	return t.h.DescribeTaskList(ctx, request)
}

// ListTaskListPartitions forwards request to the underlying handler
func (t ThriftHandler) ListTaskListPartitions(ctx context.Context, request *m.ListTaskListPartitionsRequest) (response *s.ListTaskListPartitionsResponse, err error) {
	return t.h.ListTaskListPartitions(ctx, request)
}

// PollForActivityTask forwards request to the underlying handler
func (t ThriftHandler) PollForActivityTask(ctx context.Context, request *m.PollForActivityTaskRequest) (response *s.PollForActivityTaskResponse, err error) {
	return t.h.PollForActivityTask(ctx, request)
}

// PollForDecisionTask forwards request to the underlying handler
func (t ThriftHandler) PollForDecisionTask(ctx context.Context, request *m.PollForDecisionTaskRequest) (response *m.PollForDecisionTaskResponse, err error) {
	return t.h.PollForDecisionTask(ctx, request)
}

// QueryWorkflow forwards request to the underlying handler
func (t ThriftHandler) QueryWorkflow(ctx context.Context, request *m.QueryWorkflowRequest) (response *s.QueryWorkflowResponse, err error) {
	return t.h.QueryWorkflow(ctx, request)
}

// RespondQueryTaskCompleted forwards request to the underlying handler
func (t ThriftHandler) RespondQueryTaskCompleted(ctx context.Context, request *m.RespondQueryTaskCompletedRequest) (err error) {
	return t.h.RespondQueryTaskCompleted(ctx, request)
}
