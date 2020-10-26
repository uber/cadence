// Copyright (c) 2020 Uber Technologies Inc.
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
	"testing"

	"github.com/uber/cadence/.gen/go/health"
	m "github.com/uber/cadence/.gen/go/matching"
	s "github.com/uber/cadence/.gen/go/shared"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestThriftHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := NewMockHandler(ctrl)
	th := NewThriftHandler(h)
	ctx := context.Background()

	t.Run("Health", func(t *testing.T) {
		h.EXPECT().Health(ctx).Return(&health.HealthStatus{}, assert.AnError).Times(1)
		resp, err := th.Health(ctx)
		assert.Equal(t, health.HealthStatus{}, *resp)
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("AddActivityTask", func(t *testing.T) {
		h.EXPECT().AddActivityTask(ctx, &m.AddActivityTaskRequest{}).Return(assert.AnError).Times(1)
		err := th.AddActivityTask(ctx, &m.AddActivityTaskRequest{})
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("AddDecisionTask", func(t *testing.T) {
		h.EXPECT().AddDecisionTask(ctx, &m.AddDecisionTaskRequest{}).Return(assert.AnError).Times(1)
		err := th.AddDecisionTask(ctx, &m.AddDecisionTaskRequest{})
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("CancelOutstandingPoll", func(t *testing.T) {
		h.EXPECT().CancelOutstandingPoll(ctx, &m.CancelOutstandingPollRequest{}).Return(assert.AnError).Times(1)
		err := th.CancelOutstandingPoll(ctx, &m.CancelOutstandingPollRequest{})
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("DescribeTaskList", func(t *testing.T) {
		h.EXPECT().DescribeTaskList(ctx, &m.DescribeTaskListRequest{}).Return(&s.DescribeTaskListResponse{}, assert.AnError).Times(1)
		resp, err := th.DescribeTaskList(ctx, &m.DescribeTaskListRequest{})
		assert.Equal(t, s.DescribeTaskListResponse{}, *resp)
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("ListTaskListPartitions", func(t *testing.T) {
		h.EXPECT().ListTaskListPartitions(ctx, &m.ListTaskListPartitionsRequest{}).Return(&s.ListTaskListPartitionsResponse{}, assert.AnError).Times(1)
		resp, err := th.ListTaskListPartitions(ctx, &m.ListTaskListPartitionsRequest{})
		assert.Equal(t, s.ListTaskListPartitionsResponse{}, *resp)
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("PollForActivityTask", func(t *testing.T) {
		h.EXPECT().PollForActivityTask(ctx, &m.PollForActivityTaskRequest{}).Return(&s.PollForActivityTaskResponse{}, assert.AnError).Times(1)
		resp, err := th.PollForActivityTask(ctx, &m.PollForActivityTaskRequest{})
		assert.Equal(t, s.PollForActivityTaskResponse{}, *resp)
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("PollForDecisionTask", func(t *testing.T) {
		h.EXPECT().PollForDecisionTask(ctx, &m.PollForDecisionTaskRequest{}).Return(&m.PollForDecisionTaskResponse{}, assert.AnError).Times(1)
		resp, err := th.PollForDecisionTask(ctx, &m.PollForDecisionTaskRequest{})
		assert.Equal(t, m.PollForDecisionTaskResponse{}, *resp)
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("QueryWorkflow", func(t *testing.T) {
		h.EXPECT().QueryWorkflow(ctx, &m.QueryWorkflowRequest{}).Return(&s.QueryWorkflowResponse{}, assert.AnError).Times(1)
		resp, err := th.QueryWorkflow(ctx, &m.QueryWorkflowRequest{})
		assert.Equal(t, s.QueryWorkflowResponse{}, *resp)
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("RespondQueryTaskCompleted", func(t *testing.T) {
		h.EXPECT().RespondQueryTaskCompleted(ctx, &m.RespondQueryTaskCompletedRequest{}).Return(assert.AnError).Times(1)
		err := th.RespondQueryTaskCompleted(ctx, &m.RespondQueryTaskCompletedRequest{})
		assert.Equal(t, assert.AnError, err)
	})
}
