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

package frontend

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/.gen/go/admin"
	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/.gen/go/shared"
)

func TestAdminThriftHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := NewMockAdminHandler(ctrl)
	th := NewAdminThriftHandler(h)
	ctx := context.Background()

	t.Run("AddSearchAttribute", func(t *testing.T) {
		h.EXPECT().AddSearchAttribute(ctx, &admin.AddSearchAttributeRequest{}).Return(assert.AnError).Times(1)
		err := th.AddSearchAttribute(ctx, &admin.AddSearchAttributeRequest{})
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("CloseShard", func(t *testing.T) {
		h.EXPECT().CloseShard(ctx, &shared.CloseShardRequest{}).Return(assert.AnError).Times(1)
		err := th.CloseShard(ctx, &shared.CloseShardRequest{})
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("DescribeCluster", func(t *testing.T) {
		h.EXPECT().DescribeCluster(ctx).Return(&admin.DescribeClusterResponse{}, assert.AnError).Times(1)
		resp, err := th.DescribeCluster(ctx)
		assert.Equal(t, admin.DescribeClusterResponse{}, *resp)
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("DescribeHistoryHost", func(t *testing.T) {
		h.EXPECT().DescribeHistoryHost(ctx, &shared.DescribeHistoryHostRequest{}).Return(&shared.DescribeHistoryHostResponse{}, assert.AnError).Times(1)
		resp, err := th.DescribeHistoryHost(ctx, &shared.DescribeHistoryHostRequest{})
		assert.Equal(t, shared.DescribeHistoryHostResponse{}, *resp)
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("DescribeQueue", func(t *testing.T) {
		h.EXPECT().DescribeQueue(ctx, &shared.DescribeQueueRequest{}).Return(&shared.DescribeQueueResponse{}, assert.AnError).Times(1)
		resp, err := th.DescribeQueue(ctx, &shared.DescribeQueueRequest{})
		assert.Equal(t, shared.DescribeQueueResponse{}, *resp)
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("DescribeWorkflowExecution", func(t *testing.T) {
		h.EXPECT().DescribeWorkflowExecution(ctx, &admin.DescribeWorkflowExecutionRequest{}).Return(&admin.DescribeWorkflowExecutionResponse{}, assert.AnError).Times(1)
		resp, err := th.DescribeWorkflowExecution(ctx, &admin.DescribeWorkflowExecutionRequest{})
		assert.Equal(t, admin.DescribeWorkflowExecutionResponse{}, *resp)
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("GetDLQReplicationMessages", func(t *testing.T) {
		h.EXPECT().GetDLQReplicationMessages(ctx, &replicator.GetDLQReplicationMessagesRequest{}).Return(&replicator.GetDLQReplicationMessagesResponse{}, assert.AnError).Times(1)
		resp, err := th.GetDLQReplicationMessages(ctx, &replicator.GetDLQReplicationMessagesRequest{})
		assert.Equal(t, replicator.GetDLQReplicationMessagesResponse{}, *resp)
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("GetDomainReplicationMessages", func(t *testing.T) {
		h.EXPECT().GetDomainReplicationMessages(ctx, &replicator.GetDomainReplicationMessagesRequest{}).Return(&replicator.GetDomainReplicationMessagesResponse{}, assert.AnError).Times(1)
		resp, err := th.GetDomainReplicationMessages(ctx, &replicator.GetDomainReplicationMessagesRequest{})
		assert.Equal(t, replicator.GetDomainReplicationMessagesResponse{}, *resp)
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("GetReplicationMessages", func(t *testing.T) {
		h.EXPECT().GetReplicationMessages(ctx, &replicator.GetReplicationMessagesRequest{}).Return(&replicator.GetReplicationMessagesResponse{}, assert.AnError).Times(1)
		resp, err := th.GetReplicationMessages(ctx, &replicator.GetReplicationMessagesRequest{})
		assert.Equal(t, replicator.GetReplicationMessagesResponse{}, *resp)
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("GetWorkflowExecutionRawHistoryV2", func(t *testing.T) {
		h.EXPECT().GetWorkflowExecutionRawHistoryV2(ctx, &admin.GetWorkflowExecutionRawHistoryV2Request{}).Return(&admin.GetWorkflowExecutionRawHistoryV2Response{}, assert.AnError).Times(1)
		resp, err := th.GetWorkflowExecutionRawHistoryV2(ctx, &admin.GetWorkflowExecutionRawHistoryV2Request{})
		assert.Equal(t, admin.GetWorkflowExecutionRawHistoryV2Response{}, *resp)
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("MergeDLQMessages", func(t *testing.T) {
		h.EXPECT().MergeDLQMessages(ctx, &replicator.MergeDLQMessagesRequest{}).Return(&replicator.MergeDLQMessagesResponse{}, assert.AnError).Times(1)
		resp, err := th.MergeDLQMessages(ctx, &replicator.MergeDLQMessagesRequest{})
		assert.Equal(t, replicator.MergeDLQMessagesResponse{}, *resp)
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("PurgeDLQMessages", func(t *testing.T) {
		h.EXPECT().PurgeDLQMessages(ctx, &replicator.PurgeDLQMessagesRequest{}).Return(assert.AnError).Times(1)
		err := th.PurgeDLQMessages(ctx, &replicator.PurgeDLQMessagesRequest{})
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("ReadDLQMessages", func(t *testing.T) {
		h.EXPECT().ReadDLQMessages(ctx, &replicator.ReadDLQMessagesRequest{}).Return(&replicator.ReadDLQMessagesResponse{}, assert.AnError).Times(1)
		resp, err := th.ReadDLQMessages(ctx, &replicator.ReadDLQMessagesRequest{})
		assert.Equal(t, replicator.ReadDLQMessagesResponse{}, *resp)
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("ReapplyEvents", func(t *testing.T) {
		h.EXPECT().ReapplyEvents(ctx, &shared.ReapplyEventsRequest{}).Return(assert.AnError).Times(1)
		err := th.ReapplyEvents(ctx, &shared.ReapplyEventsRequest{})
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("RefreshWorkflowTasks", func(t *testing.T) {
		h.EXPECT().RefreshWorkflowTasks(ctx, &shared.RefreshWorkflowTasksRequest{}).Return(assert.AnError).Times(1)
		err := th.RefreshWorkflowTasks(ctx, &shared.RefreshWorkflowTasksRequest{})
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("RemoveTask", func(t *testing.T) {
		h.EXPECT().RemoveTask(ctx, &shared.RemoveTaskRequest{}).Return(assert.AnError).Times(1)
		err := th.RemoveTask(ctx, &shared.RemoveTaskRequest{})
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("ResendReplicationTasks", func(t *testing.T) {
		h.EXPECT().ResendReplicationTasks(ctx, &admin.ResendReplicationTasksRequest{}).Return(assert.AnError).Times(1)
		err := th.ResendReplicationTasks(ctx, &admin.ResendReplicationTasksRequest{})
		assert.Equal(t, assert.AnError, err)
	})
	t.Run("ResetQueue", func(t *testing.T) {
		h.EXPECT().ResetQueue(ctx, &shared.ResetQueueRequest{}).Return(assert.AnError).Times(1)
		err := th.ResetQueue(ctx, &shared.ResetQueueRequest{})
		assert.Equal(t, assert.AnError, err)
	})
}
