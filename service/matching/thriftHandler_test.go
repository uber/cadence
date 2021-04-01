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
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestThriftHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := NewMockHandler(ctrl)
	th := NewThriftHandler(h)
	ctx := context.Background()
	taggedCtx := metrics.TagContext(ctx, metrics.ThriftTransportTag())
	internalErr := &types.InternalServiceError{Message: "test"}
	expectedErr := &s.InternalServiceError{Message: "test"}

	t.Run("Health", func(t *testing.T) {
		h.EXPECT().Health(taggedCtx).Return(&types.HealthStatus{}, internalErr).Times(1)
		resp, err := th.Health(ctx)
		assert.Equal(t, health.HealthStatus{Msg: common.StringPtr("")}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("AddActivityTask", func(t *testing.T) {
		h.EXPECT().AddActivityTask(taggedCtx, &types.AddActivityTaskRequest{}).Return(internalErr).Times(1)
		err := th.AddActivityTask(ctx, &m.AddActivityTaskRequest{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("AddDecisionTask", func(t *testing.T) {
		h.EXPECT().AddDecisionTask(taggedCtx, &types.AddDecisionTaskRequest{}).Return(internalErr).Times(1)
		err := th.AddDecisionTask(ctx, &m.AddDecisionTaskRequest{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("CancelOutstandingPoll", func(t *testing.T) {
		h.EXPECT().CancelOutstandingPoll(taggedCtx, &types.CancelOutstandingPollRequest{}).Return(internalErr).Times(1)
		err := th.CancelOutstandingPoll(ctx, &m.CancelOutstandingPollRequest{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("DescribeTaskList", func(t *testing.T) {
		h.EXPECT().DescribeTaskList(taggedCtx, &types.MatchingDescribeTaskListRequest{}).Return(&types.DescribeTaskListResponse{}, internalErr).Times(1)
		resp, err := th.DescribeTaskList(ctx, &m.DescribeTaskListRequest{})
		assert.Equal(t, s.DescribeTaskListResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("ListTaskListPartitions", func(t *testing.T) {
		h.EXPECT().ListTaskListPartitions(taggedCtx, &types.MatchingListTaskListPartitionsRequest{}).Return(&types.ListTaskListPartitionsResponse{}, internalErr).Times(1)
		resp, err := th.ListTaskListPartitions(ctx, &m.ListTaskListPartitionsRequest{})
		assert.Equal(t, s.ListTaskListPartitionsResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("PollForActivityTask", func(t *testing.T) {
		h.EXPECT().PollForActivityTask(taggedCtx, &types.MatchingPollForActivityTaskRequest{}).Return(&types.PollForActivityTaskResponse{}, internalErr).Times(1)
		resp, err := th.PollForActivityTask(ctx, &m.PollForActivityTaskRequest{})
		assert.Equal(t, s.PollForActivityTaskResponse{WorkflowDomain: common.StringPtr(""), ActivityId: common.StringPtr(""), Attempt: common.Int32Ptr(0)}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("PollForDecisionTask", func(t *testing.T) {
		h.EXPECT().PollForDecisionTask(taggedCtx, &types.MatchingPollForDecisionTaskRequest{}).Return(&types.MatchingPollForDecisionTaskResponse{}, internalErr).Times(1)
		resp, err := th.PollForDecisionTask(ctx, &m.PollForDecisionTaskRequest{})
		assert.Equal(t, m.PollForDecisionTaskResponse{
			StartedEventId:         common.Int64Ptr(0),
			NextEventId:            common.Int64Ptr(0),
			Attempt:                common.Int64Ptr(0),
			StickyExecutionEnabled: common.BoolPtr(false),
			BacklogCountHint:       common.Int64Ptr(0),
			EventStoreVersion:      common.Int32Ptr(0),
		}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("QueryWorkflow", func(t *testing.T) {
		h.EXPECT().QueryWorkflow(taggedCtx, &types.MatchingQueryWorkflowRequest{}).Return(&types.QueryWorkflowResponse{}, internalErr).Times(1)
		resp, err := th.QueryWorkflow(ctx, &m.QueryWorkflowRequest{})
		assert.Equal(t, s.QueryWorkflowResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("RespondQueryTaskCompleted", func(t *testing.T) {
		h.EXPECT().RespondQueryTaskCompleted(taggedCtx, &types.MatchingRespondQueryTaskCompletedRequest{}).Return(internalErr).Times(1)
		err := th.RespondQueryTaskCompleted(ctx, &m.RespondQueryTaskCompletedRequest{})
		assert.Equal(t, expectedErr, err)
	})
}
