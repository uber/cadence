// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package frontend

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/.gen/go/health"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/types"
)

func TestThriftHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	h := NewMockHandler(ctrl)
	th := NewThriftHandler(h)
	ctx := context.Background()
	internalErr := &types.InternalServiceError{Message: "test"}
	expectedErr := &shared.InternalServiceError{Message: "test"}

	t.Run("Health", func(t *testing.T) {
		h.EXPECT().Health(ctx).Return(&health.HealthStatus{}, internalErr).Times(1)
		resp, err := th.Health(ctx)
		assert.Equal(t, health.HealthStatus{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("CountWorkflowExecutions", func(t *testing.T) {
		h.EXPECT().CountWorkflowExecutions(ctx, &types.CountWorkflowExecutionsRequest{}).Return(&types.CountWorkflowExecutionsResponse{}, internalErr).Times(1)
		resp, err := th.CountWorkflowExecutions(ctx, &shared.CountWorkflowExecutionsRequest{})
		assert.Equal(t, shared.CountWorkflowExecutionsResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("DeprecateDomain", func(t *testing.T) {
		h.EXPECT().DeprecateDomain(ctx, &types.DeprecateDomainRequest{}).Return(internalErr).Times(1)
		err := th.DeprecateDomain(ctx, &shared.DeprecateDomainRequest{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("DescribeDomain", func(t *testing.T) {
		h.EXPECT().DescribeDomain(ctx, &types.DescribeDomainRequest{}).Return(&types.DescribeDomainResponse{}, internalErr).Times(1)
		resp, err := th.DescribeDomain(ctx, &shared.DescribeDomainRequest{})
		assert.Equal(t, shared.DescribeDomainResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("DescribeTaskList", func(t *testing.T) {
		h.EXPECT().DescribeTaskList(ctx, &shared.DescribeTaskListRequest{}).Return(&shared.DescribeTaskListResponse{}, internalErr).Times(1)
		resp, err := th.DescribeTaskList(ctx, &shared.DescribeTaskListRequest{})
		assert.Equal(t, shared.DescribeTaskListResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("DescribeWorkflowExecution", func(t *testing.T) {
		h.EXPECT().DescribeWorkflowExecution(ctx, &shared.DescribeWorkflowExecutionRequest{}).Return(&shared.DescribeWorkflowExecutionResponse{}, internalErr).Times(1)
		resp, err := th.DescribeWorkflowExecution(ctx, &shared.DescribeWorkflowExecutionRequest{})
		assert.Equal(t, shared.DescribeWorkflowExecutionResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("GetClusterInfo", func(t *testing.T) {
		h.EXPECT().GetClusterInfo(ctx).Return(&shared.ClusterInfo{}, internalErr).Times(1)
		resp, err := th.GetClusterInfo(ctx)
		assert.Equal(t, shared.ClusterInfo{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("GetSearchAttributes", func(t *testing.T) {
		h.EXPECT().GetSearchAttributes(ctx).Return(&shared.GetSearchAttributesResponse{}, internalErr).Times(1)
		resp, err := th.GetSearchAttributes(ctx)
		assert.Equal(t, shared.GetSearchAttributesResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("GetWorkflowExecutionHistory", func(t *testing.T) {
		h.EXPECT().GetWorkflowExecutionHistory(ctx, &shared.GetWorkflowExecutionHistoryRequest{}).Return(&shared.GetWorkflowExecutionHistoryResponse{}, internalErr).Times(1)
		resp, err := th.GetWorkflowExecutionHistory(ctx, &shared.GetWorkflowExecutionHistoryRequest{})
		assert.Equal(t, shared.GetWorkflowExecutionHistoryResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("ListArchivedWorkflowExecutions", func(t *testing.T) {
		h.EXPECT().ListArchivedWorkflowExecutions(ctx, &shared.ListArchivedWorkflowExecutionsRequest{}).Return(&shared.ListArchivedWorkflowExecutionsResponse{}, internalErr).Times(1)
		resp, err := th.ListArchivedWorkflowExecutions(ctx, &shared.ListArchivedWorkflowExecutionsRequest{})
		assert.Equal(t, shared.ListArchivedWorkflowExecutionsResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("ListClosedWorkflowExecutions", func(t *testing.T) {
		h.EXPECT().ListClosedWorkflowExecutions(ctx, &types.ListClosedWorkflowExecutionsRequest{}).Return(&types.ListClosedWorkflowExecutionsResponse{}, internalErr).Times(1)
		resp, err := th.ListClosedWorkflowExecutions(ctx, &shared.ListClosedWorkflowExecutionsRequest{})
		assert.Equal(t, shared.ListClosedWorkflowExecutionsResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("ListDomains", func(t *testing.T) {
		h.EXPECT().ListDomains(ctx, &types.ListDomainsRequest{}).Return(&types.ListDomainsResponse{}, internalErr).Times(1)
		resp, err := th.ListDomains(ctx, &shared.ListDomainsRequest{})
		assert.Equal(t, shared.ListDomainsResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("ListOpenWorkflowExecutions", func(t *testing.T) {
		h.EXPECT().ListOpenWorkflowExecutions(ctx, &types.ListOpenWorkflowExecutionsRequest{}).Return(&types.ListOpenWorkflowExecutionsResponse{}, internalErr).Times(1)
		resp, err := th.ListOpenWorkflowExecutions(ctx, &shared.ListOpenWorkflowExecutionsRequest{})
		assert.Equal(t, shared.ListOpenWorkflowExecutionsResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("ListTaskListPartitions", func(t *testing.T) {
		h.EXPECT().ListTaskListPartitions(ctx, &shared.ListTaskListPartitionsRequest{}).Return(&shared.ListTaskListPartitionsResponse{}, internalErr).Times(1)
		resp, err := th.ListTaskListPartitions(ctx, &shared.ListTaskListPartitionsRequest{})
		assert.Equal(t, shared.ListTaskListPartitionsResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("ListWorkflowExecutions", func(t *testing.T) {
		h.EXPECT().ListWorkflowExecutions(ctx, &types.ListWorkflowExecutionsRequest{}).Return(&types.ListWorkflowExecutionsResponse{}, internalErr).Times(1)
		resp, err := th.ListWorkflowExecutions(ctx, &shared.ListWorkflowExecutionsRequest{})
		assert.Equal(t, shared.ListWorkflowExecutionsResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("PollForActivityTask", func(t *testing.T) {
		h.EXPECT().PollForActivityTask(ctx, &shared.PollForActivityTaskRequest{}).Return(&shared.PollForActivityTaskResponse{}, internalErr).Times(1)
		resp, err := th.PollForActivityTask(ctx, &shared.PollForActivityTaskRequest{})
		assert.Equal(t, shared.PollForActivityTaskResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("PollForDecisionTask", func(t *testing.T) {
		h.EXPECT().PollForDecisionTask(ctx, &shared.PollForDecisionTaskRequest{}).Return(&shared.PollForDecisionTaskResponse{}, internalErr).Times(1)
		resp, err := th.PollForDecisionTask(ctx, &shared.PollForDecisionTaskRequest{})
		assert.Equal(t, shared.PollForDecisionTaskResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("QueryWorkflow", func(t *testing.T) {
		h.EXPECT().QueryWorkflow(ctx, &shared.QueryWorkflowRequest{}).Return(&shared.QueryWorkflowResponse{}, internalErr).Times(1)
		resp, err := th.QueryWorkflow(ctx, &shared.QueryWorkflowRequest{})
		assert.Equal(t, shared.QueryWorkflowResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("RecordActivityTaskHeartbeat", func(t *testing.T) {
		h.EXPECT().RecordActivityTaskHeartbeat(ctx, &shared.RecordActivityTaskHeartbeatRequest{}).Return(&shared.RecordActivityTaskHeartbeatResponse{}, internalErr).Times(1)
		resp, err := th.RecordActivityTaskHeartbeat(ctx, &shared.RecordActivityTaskHeartbeatRequest{})
		assert.Equal(t, shared.RecordActivityTaskHeartbeatResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("RecordActivityTaskHeartbeatByID", func(t *testing.T) {
		h.EXPECT().RecordActivityTaskHeartbeatByID(ctx, &shared.RecordActivityTaskHeartbeatByIDRequest{}).Return(&shared.RecordActivityTaskHeartbeatResponse{}, internalErr).Times(1)
		resp, err := th.RecordActivityTaskHeartbeatByID(ctx, &shared.RecordActivityTaskHeartbeatByIDRequest{})
		assert.Equal(t, shared.RecordActivityTaskHeartbeatResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("RegisterDomain", func(t *testing.T) {
		h.EXPECT().RegisterDomain(ctx, &types.RegisterDomainRequest{}).Return(internalErr).Times(1)
		err := th.RegisterDomain(ctx, &shared.RegisterDomainRequest{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("RequestCancelWorkflowExecution", func(t *testing.T) {
		h.EXPECT().RequestCancelWorkflowExecution(ctx, &shared.RequestCancelWorkflowExecutionRequest{}).Return(internalErr).Times(1)
		err := th.RequestCancelWorkflowExecution(ctx, &shared.RequestCancelWorkflowExecutionRequest{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("ResetStickyTaskList", func(t *testing.T) {
		h.EXPECT().ResetStickyTaskList(ctx, &shared.ResetStickyTaskListRequest{}).Return(&shared.ResetStickyTaskListResponse{}, internalErr).Times(1)
		resp, err := th.ResetStickyTaskList(ctx, &shared.ResetStickyTaskListRequest{})
		assert.Equal(t, shared.ResetStickyTaskListResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("ResetWorkflowExecution", func(t *testing.T) {
		h.EXPECT().ResetWorkflowExecution(ctx, &shared.ResetWorkflowExecutionRequest{}).Return(&shared.ResetWorkflowExecutionResponse{}, internalErr).Times(1)
		resp, err := th.ResetWorkflowExecution(ctx, &shared.ResetWorkflowExecutionRequest{})
		assert.Equal(t, shared.ResetWorkflowExecutionResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("RespondActivityTaskCanceled", func(t *testing.T) {
		h.EXPECT().RespondActivityTaskCanceled(ctx, &shared.RespondActivityTaskCanceledRequest{}).Return(internalErr).Times(1)
		err := th.RespondActivityTaskCanceled(ctx, &shared.RespondActivityTaskCanceledRequest{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("RespondActivityTaskCanceledByID", func(t *testing.T) {
		h.EXPECT().RespondActivityTaskCanceledByID(ctx, &shared.RespondActivityTaskCanceledByIDRequest{}).Return(internalErr).Times(1)
		err := th.RespondActivityTaskCanceledByID(ctx, &shared.RespondActivityTaskCanceledByIDRequest{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("RespondActivityTaskCompleted", func(t *testing.T) {
		h.EXPECT().RespondActivityTaskCompleted(ctx, &shared.RespondActivityTaskCompletedRequest{}).Return(internalErr).Times(1)
		err := th.RespondActivityTaskCompleted(ctx, &shared.RespondActivityTaskCompletedRequest{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("RespondActivityTaskCompletedByID", func(t *testing.T) {
		h.EXPECT().RespondActivityTaskCompletedByID(ctx, &shared.RespondActivityTaskCompletedByIDRequest{}).Return(internalErr).Times(1)
		err := th.RespondActivityTaskCompletedByID(ctx, &shared.RespondActivityTaskCompletedByIDRequest{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("RespondActivityTaskFailed", func(t *testing.T) {
		h.EXPECT().RespondActivityTaskFailed(ctx, &shared.RespondActivityTaskFailedRequest{}).Return(internalErr).Times(1)
		err := th.RespondActivityTaskFailed(ctx, &shared.RespondActivityTaskFailedRequest{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("RespondActivityTaskFailedByID", func(t *testing.T) {
		h.EXPECT().RespondActivityTaskFailedByID(ctx, &shared.RespondActivityTaskFailedByIDRequest{}).Return(internalErr).Times(1)
		err := th.RespondActivityTaskFailedByID(ctx, &shared.RespondActivityTaskFailedByIDRequest{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("RespondDecisionTaskCompleted", func(t *testing.T) {
		h.EXPECT().RespondDecisionTaskCompleted(ctx, &shared.RespondDecisionTaskCompletedRequest{}).Return(&shared.RespondDecisionTaskCompletedResponse{}, internalErr).Times(1)
		resp, err := th.RespondDecisionTaskCompleted(ctx, &shared.RespondDecisionTaskCompletedRequest{})
		assert.Equal(t, shared.RespondDecisionTaskCompletedResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("RespondDecisionTaskFailed", func(t *testing.T) {
		h.EXPECT().RespondDecisionTaskFailed(ctx, &shared.RespondDecisionTaskFailedRequest{}).Return(internalErr).Times(1)
		err := th.RespondDecisionTaskFailed(ctx, &shared.RespondDecisionTaskFailedRequest{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("RespondQueryTaskCompleted", func(t *testing.T) {
		h.EXPECT().RespondQueryTaskCompleted(ctx, &shared.RespondQueryTaskCompletedRequest{}).Return(internalErr).Times(1)
		err := th.RespondQueryTaskCompleted(ctx, &shared.RespondQueryTaskCompletedRequest{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("ScanWorkflowExecutions", func(t *testing.T) {
		h.EXPECT().ScanWorkflowExecutions(ctx, &types.ListWorkflowExecutionsRequest{}).Return(&types.ListWorkflowExecutionsResponse{}, internalErr).Times(1)
		resp, err := th.ScanWorkflowExecutions(ctx, &shared.ListWorkflowExecutionsRequest{})
		assert.Equal(t, shared.ListWorkflowExecutionsResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("SignalWithStartWorkflowExecution", func(t *testing.T) {
		h.EXPECT().SignalWithStartWorkflowExecution(ctx, &shared.SignalWithStartWorkflowExecutionRequest{}).Return(&shared.StartWorkflowExecutionResponse{}, internalErr).Times(1)
		resp, err := th.SignalWithStartWorkflowExecution(ctx, &shared.SignalWithStartWorkflowExecutionRequest{})
		assert.Equal(t, shared.StartWorkflowExecutionResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("SignalWorkflowExecution", func(t *testing.T) {
		h.EXPECT().SignalWorkflowExecution(ctx, &shared.SignalWorkflowExecutionRequest{}).Return(internalErr).Times(1)
		err := th.SignalWorkflowExecution(ctx, &shared.SignalWorkflowExecutionRequest{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("StartWorkflowExecution", func(t *testing.T) {
		h.EXPECT().StartWorkflowExecution(ctx, &shared.StartWorkflowExecutionRequest{}).Return(&shared.StartWorkflowExecutionResponse{}, internalErr).Times(1)
		resp, err := th.StartWorkflowExecution(ctx, &shared.StartWorkflowExecutionRequest{})
		assert.Equal(t, shared.StartWorkflowExecutionResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("TerminateWorkflowExecution", func(t *testing.T) {
		h.EXPECT().TerminateWorkflowExecution(ctx, &shared.TerminateWorkflowExecutionRequest{}).Return(internalErr).Times(1)
		err := th.TerminateWorkflowExecution(ctx, &shared.TerminateWorkflowExecutionRequest{})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("UpdateDomain", func(t *testing.T) {
		h.EXPECT().UpdateDomain(ctx, &types.UpdateDomainRequest{}).Return(&types.UpdateDomainResponse{}, internalErr).Times(1)
		resp, err := th.UpdateDomain(ctx, &shared.UpdateDomainRequest{})
		assert.Equal(t, shared.UpdateDomainResponse{}, *resp)
		assert.Equal(t, expectedErr, err)
	})
}
