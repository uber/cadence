// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package metered

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/frontend/api"
	"github.com/uber/cadence/service/frontend/config"
)

func TestContextMetricsTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockHandler := api.NewMockHandler(ctrl)
	mockHandler.EXPECT().CountWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&types.CountWorkflowExecutionsResponse{}, nil).Times(1)
	mockDomainCache := cache.NewMockDomainCache(ctrl)
	testScope := tally.NewTestScope("test", nil)
	metricsClient := metrics.NewClient(testScope, metrics.Frontend)
	handler := NewAPIHandler(mockHandler, testlogger.New(t), metricsClient, mockDomainCache, nil)

	tag := metrics.TransportTag("grpc")
	ctx := metrics.TagContext(context.Background(), tag)
	handler.CountWorkflowExecutions(ctx, nil)

	snapshot := testScope.Snapshot()
	for _, counter := range snapshot.Counters() {
		if counter.Name() == "test.cadence_requests" {
			assert.Equal(t, tag.Value(), counter.Tags()[tag.Key()])
			return
		}
	}
	assert.Fail(t, "counter not found")
}

func TestSignalMetricHasSignalName(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockHandler := api.NewMockHandler(ctrl)
	mockHandler.EXPECT().SignalWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	mockDomainCache := cache.NewMockDomainCache(ctrl)
	testScope := tally.NewTestScope("test", nil)
	metricsClient := metrics.NewClient(testScope, metrics.Frontend)
	handler := NewAPIHandler(mockHandler, testlogger.New(t), metricsClient, mockDomainCache, &config.Config{EmitSignalNameMetricsTag: dynamicconfig.GetBoolPropertyFnFilteredByDomain(true)})

	signalRequest := &types.SignalWorkflowExecutionRequest{
		SignalName: "test_signal",
	}
	handler.SignalWorkflowExecution(context.Background(), signalRequest)

	expectedMetrics := make(map[string]bool)
	expectedMetrics["test.cadence_requests"] = false

	snapshot := testScope.Snapshot()
	for _, counter := range snapshot.Counters() {
		if _, ok := expectedMetrics[counter.Name()]; ok {
			expectedMetrics[counter.Name()] = true
		}
		if val, ok := counter.Tags()["signalName"]; ok {
			assert.Equal(t, val, "test_signal")
		} else {
			assert.Fail(t, "Couldn't find signalName tag")
		}
	}
	assert.True(t, expectedMetrics["test.cadence_requests"])
}

func TestHandleErr_InternalServiceError(t *testing.T) {
	logger := testlogger.New(t)
	testScope := tally.NewTestScope("test", nil)
	metricsClient := metrics.NewClient(testScope, metrics.Frontend)
	handler := &apiHandler{}

	err := handler.handleErr(&types.InternalServiceError{Message: "internal error"}, metricsClient.Scope(0), logger)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "cadence internal error")
	assert.Contains(t, err.Error(), "internal error")
}

func TestHandleErr_UncategorizedError(t *testing.T) {
	logger := testlogger.New(t)
	testScope := tally.NewTestScope("test", nil)
	metricsClient := metrics.NewClient(testScope, metrics.Frontend)
	handler := &apiHandler{}

	err := handler.handleErr(errors.New("unknown error"), metricsClient.Scope(0), logger)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "cadence internal uncategorized error")
}

func TestToSignalWithStartWorkflowExecutionRequestTags(t *testing.T) {
	req := &types.SignalWithStartWorkflowExecutionRequest{
		Domain:     "test-domain",
		WorkflowID: "test-workflow-id",
		WorkflowType: &types.WorkflowType{
			Name: "test-workflow-type",
		},
		SignalName: "test-signal",
	}

	tags := toSignalWithStartWorkflowExecutionRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowID("test-workflow-id"),
		tag.WorkflowType("test-workflow-type"),
		tag.WorkflowSignalName("test-signal"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}

func TestToDescribeWorkflowExecutionRequestTags(t *testing.T) {
	req := &types.DescribeWorkflowExecutionRequest{
		Domain: "test-domain",
		Execution: &types.WorkflowExecution{
			WorkflowID: "test-workflow-id",
			RunID:      "test-run-id",
		},
	}

	tags := toDescribeWorkflowExecutionRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowID("test-workflow-id"),
		tag.WorkflowRunID("test-run-id"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}

func TestToStartWorkflowExecutionRequestTags(t *testing.T) {
	req := &types.StartWorkflowExecutionRequest{
		Domain:     "test-domain",
		WorkflowID: "test-workflow-id",
		WorkflowType: &types.WorkflowType{
			Name: "test-workflow-type",
		},
		CronSchedule: "test-cron-schedule",
	}

	tags := toStartWorkflowExecutionRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowID("test-workflow-id"),
		tag.WorkflowType("test-workflow-type"),
		tag.WorkflowCronSchedule("test-cron-schedule"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}
func TestToStartWorkflowExecutionAsyncRequestTags(t *testing.T) {
	req := &types.StartWorkflowExecutionAsyncRequest{
		StartWorkflowExecutionRequest: &types.StartWorkflowExecutionRequest{
			Domain:     "test-domain",
			WorkflowID: "test-workflow-id",
			WorkflowType: &types.WorkflowType{
				Name: "test-workflow-type",
			},
			CronSchedule: "test-cron-schedule",
		},
	}

	tags := toStartWorkflowExecutionAsyncRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowID("test-workflow-id"),
		tag.WorkflowType("test-workflow-type"),
		tag.WorkflowCronSchedule("test-cron-schedule"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}

func TestToTerminateWorkflowExecutionRequestTags(t *testing.T) {
	req := &types.TerminateWorkflowExecutionRequest{
		Domain: "test-domain",
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: "test-workflow-id",
			RunID:      "test-run-id",
		},
	}

	tags := toTerminateWorkflowExecutionRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowID("test-workflow-id"),
		tag.WorkflowRunID("test-run-id"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}

func TestToScanWorkflowExecutionsRequestTags(t *testing.T) {
	req := &types.ListWorkflowExecutionsRequest{
		Domain: "test-domain",
	}

	tags := toScanWorkflowExecutionsRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}

func TestToListTaskListPartitionsRequestTags(t *testing.T) {
	kind := types.TaskListKindNormal
	req := &types.ListTaskListPartitionsRequest{
		Domain: "test-domain",
		TaskList: &types.TaskList{
			Name: "test-task-list",
			Kind: &kind,
		},
	}

	tags := toListTaskListPartitionsRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowTaskListName("test-task-list"),
		tag.WorkflowTaskListKind(int32(types.TaskListKindNormal)),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}

func TestToGetTaskListsByDomainRequestTags(t *testing.T) {
	req := &types.GetTaskListsByDomainRequest{
		Domain: "test-domain",
	}

	tags := toGetTaskListsByDomainRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}

func TestToGetWorkflowExecutionHistoryRequestTags(t *testing.T) {
	req := &types.GetWorkflowExecutionHistoryRequest{
		Domain: "test-domain",
		Execution: &types.WorkflowExecution{
			WorkflowID: "test-workflow-id",
			RunID:      "test-run-id",
		},
	}

	tags := toGetWorkflowExecutionHistoryRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowID("test-workflow-id"),
		tag.WorkflowRunID("test-run-id"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}

func TestToListArchivedWorkflowExecutionsRequestTags(t *testing.T) {
	req := &types.ListArchivedWorkflowExecutionsRequest{
		Domain: "test-domain",
	}

	tags := toListArchivedWorkflowExecutionsRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}

func TestToListClosedWorkflowExecutionsRequestTags(t *testing.T) {
	req := &types.ListClosedWorkflowExecutionsRequest{
		Domain: "test-domain",
	}

	tags := toListClosedWorkflowExecutionsRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}

func TestToListOpenWorkflowExecutionsRequestTags(t *testing.T) {
	req := &types.ListOpenWorkflowExecutionsRequest{
		Domain: "test-domain",
	}

	tags := toListOpenWorkflowExecutionsRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}

func TestToRestartWorkflowExecutionRequestTags(t *testing.T) {
	req := &types.RestartWorkflowExecutionRequest{
		Domain: "test-domain",
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: "test-workflow-id",
			RunID:      "test-run-id",
		},
	}

	tags := toRestartWorkflowExecutionRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowID("test-workflow-id"),
		tag.WorkflowRunID("test-run-id"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}

func TestToRespondActivityTaskFailedByIDRequestTags(t *testing.T) {
	req := &types.RespondActivityTaskFailedByIDRequest{
		Domain:     "test-domain",
		WorkflowID: "test-workflow-id",
		RunID:      "test-run-id",
	}

	tags := toRespondActivityTaskFailedByIDRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowID("test-workflow-id"),
		tag.WorkflowRunID("test-run-id"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}

func TestToRespondActivityTaskCompletedByIDRequestTags(t *testing.T) {
	req := &types.RespondActivityTaskCompletedByIDRequest{
		Domain:     "test-domain",
		WorkflowID: "test-workflow-id",
		RunID:      "test-run-id",
	}

	tags := toRespondActivityTaskCompletedByIDRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowID("test-workflow-id"),
		tag.WorkflowRunID("test-run-id"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}
func TestToRespondActivityTaskCanceledByIDRequestTags(t *testing.T) {
	req := &types.RespondActivityTaskCanceledByIDRequest{
		Domain:     "test-domain",
		WorkflowID: "test-workflow-id",
		RunID:      "test-run-id",
	}

	tags := toRespondActivityTaskCanceledByIDRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowID("test-workflow-id"),
		tag.WorkflowRunID("test-run-id"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}

func TestToResetWorkflowExecutionRequestTags(t *testing.T) {
	req := &types.ResetWorkflowExecutionRequest{
		Domain: "test-domain",
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: "test-workflow-id",
			RunID:      "test-run-id",
		},
	}

	tags := toResetWorkflowExecutionRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowID("test-workflow-id"),
		tag.WorkflowRunID("test-run-id"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}
func TestToResetStickyTaskListRequestTags(t *testing.T) {
	req := &types.ResetStickyTaskListRequest{
		Domain: "test-domain",
		Execution: &types.WorkflowExecution{
			WorkflowID: "test-workflow-id",
			RunID:      "test-run-id",
		},
	}

	tags := toResetStickyTaskListRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowID("test-workflow-id"),
		tag.WorkflowRunID("test-run-id"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}
func TestToRequestCancelWorkflowExecutionRequestTags(t *testing.T) {
	req := &types.RequestCancelWorkflowExecutionRequest{
		Domain: "test-domain",
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: "test-workflow-id",
			RunID:      "test-run-id",
		},
	}

	tags := toRequestCancelWorkflowExecutionRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowID("test-workflow-id"),
		tag.WorkflowRunID("test-run-id"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}
func TestToRefreshWorkflowTasksRequestTags(t *testing.T) {
	req := &types.RefreshWorkflowTasksRequest{
		Domain: "test-domain",
		Execution: &types.WorkflowExecution{
			WorkflowID: "test-workflow-id",
			RunID:      "test-run-id",
		},
	}

	tags := toRefreshWorkflowTasksRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowID("test-workflow-id"),
		tag.WorkflowRunID("test-run-id"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}
func TestToRecordActivityTaskHeartbeatByIDRequestTags(t *testing.T) {
	req := &types.RecordActivityTaskHeartbeatByIDRequest{
		Domain:     "test-domain",
		WorkflowID: "test-workflow-id",
		RunID:      "test-run-id",
	}

	tags := toRecordActivityTaskHeartbeatByIDRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowID("test-workflow-id"),
		tag.WorkflowRunID("test-run-id"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}
func TestToQueryWorkflowRequestTags(t *testing.T) {
	req := &types.QueryWorkflowRequest{
		Domain: "test-domain",
		Execution: &types.WorkflowExecution{
			WorkflowID: "test-workflow-id",
			RunID:      "test-run-id",
		},
	}

	tags := toQueryWorkflowRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowID("test-workflow-id"),
		tag.WorkflowRunID("test-run-id"),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}
func TestToPollForDecisionTaskRequestTags(t *testing.T) {
	kind := types.TaskListKindNormal
	req := &types.PollForDecisionTaskRequest{
		Domain: "test-domain",
		TaskList: &types.TaskList{
			Name: "test-task-list",
			Kind: &kind,
		},
	}

	tags := toPollForDecisionTaskRequestTags(req)

	expectedTags := []tag.Tag{
		tag.WorkflowDomainName("test-domain"),
		tag.WorkflowTaskListName("test-task-list"),
		tag.WorkflowTaskListKind(int32(types.TaskListKindNormal)),
	}

	assert.ElementsMatch(t, expectedTags, tags)
}
