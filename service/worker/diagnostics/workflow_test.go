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

package diagnostics

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/workflow"
	"go.uber.org/mock/gomock"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/worker/diagnostics/invariant"
	"github.com/uber/cadence/service/worker/diagnostics/invariant/failure"
	"github.com/uber/cadence/service/worker/diagnostics/invariant/retry"
	"github.com/uber/cadence/service/worker/diagnostics/invariant/timeout"
)

type diagnosticsWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	workflowEnv *testsuite.TestWorkflowEnvironment
	dw          *dw
}

func TestDiagnosticsWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(diagnosticsWorkflowTestSuite))
}

func (s *diagnosticsWorkflowTestSuite) SetupTest() {
	s.workflowEnv = s.NewTestWorkflowEnvironment()
	controller := gomock.NewController(s.T())
	mockResource := resource.NewTest(s.T(), controller, metrics.Worker)

	s.dw = &dw{
		svcClient:     mockResource.GetSDKClient(),
		clientBean:    mockResource.ClientBean,
		metricsClient: mockResource.GetMetricsClient(),
	}

	s.T().Cleanup(func() {
		mockResource.Finish(s.T())
	})

	s.workflowEnv.RegisterWorkflowWithOptions(s.dw.DiagnosticsStarterWorkflow, workflow.RegisterOptions{Name: diagnosticsStarterWorkflow})
	s.workflowEnv.RegisterWorkflowWithOptions(s.dw.DiagnosticsWorkflow, workflow.RegisterOptions{Name: diagnosticsWorkflow})
	s.workflowEnv.RegisterActivityWithOptions(s.dw.retrieveExecutionHistory, activity.RegisterOptions{Name: retrieveWfExecutionHistoryActivity})
	s.workflowEnv.RegisterActivityWithOptions(s.dw.identifyIssues, activity.RegisterOptions{Name: identifyIssuesActivity})
	s.workflowEnv.RegisterActivityWithOptions(s.dw.rootCauseIssues, activity.RegisterOptions{Name: rootCauseIssuesActivity})
	s.workflowEnv.RegisterActivityWithOptions(s.dw.emitUsageLogs, activity.RegisterOptions{Name: emitUsageLogsActivity})
}

func (s *diagnosticsWorkflowTestSuite) TearDownTest() {
	s.workflowEnv.AssertExpectations(s.T())
}

func (s *diagnosticsWorkflowTestSuite) TestWorkflow() {
	params := &DiagnosticsStarterWorkflowInput{
		Domain:     "test",
		WorkflowID: "123",
		RunID:      "abc",
	}
	workflowTimeoutData := timeout.ExecutionTimeoutMetadata{
		ExecutionTime:     110 * time.Second,
		ConfiguredTimeout: 110 * time.Second,
		LastOngoingEvent: &types.HistoryEvent{
			ID:        1,
			Timestamp: common.Int64Ptr(testTimeStamp),
			WorkflowExecutionStartedEventAttributes: &types.WorkflowExecutionStartedEventAttributes{
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(workflowTimeoutSecond),
			},
		},
	}
	workflowTimeoutDataInBytes, err := json.Marshal(workflowTimeoutData)
	s.NoError(err)
	issues := []invariant.InvariantCheckResult{
		{
			InvariantType: timeout.TimeoutTypeExecution.String(),
			Reason:        "START_TO_CLOSE",
			Metadata:      workflowTimeoutDataInBytes,
		},
	}
	timeoutIssues := []*timeoutIssuesResult{
		{
			InvariantType:    timeout.TimeoutTypeExecution.String(),
			Reason:           "START_TO_CLOSE",
			ExecutionTimeout: &workflowTimeoutData,
		},
	}
	taskListBacklog := int64(10)
	pollersMetadataInBytes, err := json.Marshal(timeout.PollersMetadata{TaskListBacklog: taskListBacklog})
	s.NoError(err)
	rootCause := []invariant.InvariantRootCauseResult{
		{
			RootCause: invariant.RootCauseTypePollersStatus,
			Metadata:  pollersMetadataInBytes,
		},
	}
	timeoutRootCause := []*timeoutRootCauseResult{
		{
			RootCauseType:   invariant.RootCauseTypePollersStatus.String(),
			PollersMetadata: &timeout.PollersMetadata{TaskListBacklog: taskListBacklog},
		},
	}
	s.workflowEnv.OnActivity(retrieveWfExecutionHistoryActivity, mock.Anything, mock.Anything).Return(nil, nil)
	s.workflowEnv.OnActivity(identifyIssuesActivity, mock.Anything, mock.Anything).Return(issues, nil)
	s.workflowEnv.OnActivity(rootCauseIssuesActivity, mock.Anything, mock.Anything).Return(rootCause, nil)
	s.workflowEnv.OnActivity(emitUsageLogsActivity, mock.Anything, mock.Anything).Return(nil)
	s.workflowEnv.ExecuteWorkflow(diagnosticsStarterWorkflow, params)
	s.True(s.workflowEnv.IsWorkflowCompleted())
	var result DiagnosticsStarterWorkflowResult
	s.NoError(s.workflowEnv.GetWorkflowResult(&result))
	s.ElementsMatch(timeoutIssues, result.DiagnosticsResult.Timeouts.Issues)
	s.ElementsMatch(timeoutRootCause, result.DiagnosticsResult.Timeouts.RootCause)

	queriedResult := s.queryDiagnostics()
	s.ElementsMatch(queriedResult.DiagnosticsResult.Timeouts.Issues, result.DiagnosticsResult.Timeouts.Issues)
	s.ElementsMatch(queriedResult.DiagnosticsResult.Timeouts.RootCause, result.DiagnosticsResult.Timeouts.RootCause)
}

func (s *diagnosticsWorkflowTestSuite) TestWorkflow_Error() {
	params := &DiagnosticsWorkflowInput{
		Domain:     "test",
		WorkflowID: "123",
		RunID:      "abc",
	}
	mockErr := errors.New("mockErr")
	errExpected := fmt.Errorf("IdentifyIssues: %w", mockErr)
	s.workflowEnv.OnActivity(retrieveWfExecutionHistoryActivity, mock.Anything, mock.Anything).Return(nil, nil)
	s.workflowEnv.OnActivity(identifyIssuesActivity, mock.Anything, mock.Anything).Return(nil, mockErr)
	s.workflowEnv.ExecuteWorkflow(diagnosticsWorkflow, params)
	s.True(s.workflowEnv.IsWorkflowCompleted())
	s.Error(s.workflowEnv.GetWorkflowError())
	s.EqualError(s.workflowEnv.GetWorkflowError(), errExpected.Error())
}

func (s *diagnosticsWorkflowTestSuite) queryDiagnostics() DiagnosticsStarterWorkflowResult {
	queryFuture, err := s.workflowEnv.QueryWorkflow(queryDiagnosticsReport)
	s.NoError(err)

	var result DiagnosticsStarterWorkflowResult
	err = queryFuture.Get(&result)
	s.NoError(err)
	return result
}

func (s *diagnosticsWorkflowTestSuite) Test__retrieveTimeoutIssues() {
	workflowTimeoutData := timeout.ExecutionTimeoutMetadata{
		ExecutionTime:     110 * time.Second,
		ConfiguredTimeout: 110 * time.Second,
		LastOngoingEvent: &types.HistoryEvent{
			ID:        1,
			Timestamp: common.Int64Ptr(testTimeStamp),
			WorkflowExecutionStartedEventAttributes: &types.WorkflowExecutionStartedEventAttributes{
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(workflowTimeoutSecond),
			},
		},
	}
	workflowTimeoutDataInBytes, err := json.Marshal(workflowTimeoutData)
	s.NoError(err)
	childWorkflowTimeoutData := timeout.ChildWfTimeoutMetadata{
		ExecutionTime:     110 * time.Second,
		ConfiguredTimeout: 110 * time.Second,
	}
	childWorkflowTimeoutDataInBytes, err := json.Marshal(childWorkflowTimeoutData)
	s.NoError(err)
	activityTimeoutData := timeout.ActivityTimeoutMetadata{
		TimeoutType:       types.TimeoutTypeStartToClose.Ptr(),
		ConfiguredTimeout: 5 * time.Second,
		TimeElapsed:       5 * time.Second,
		HeartBeatTimeout:  0,
	}
	activityTimeoutDataInBytes, err := json.Marshal(activityTimeoutData)
	s.NoError(err)
	descTimeoutData := timeout.DecisionTimeoutMetadata{
		ConfiguredTimeout: 5 * time.Second,
	}
	descTimeoutDataInBytes, err := json.Marshal(activityTimeoutData)
	s.NoError(err)
	issues := []invariant.InvariantCheckResult{
		{
			InvariantType: timeout.TimeoutTypeExecution.String(),
			Reason:        "START_TO_CLOSE",
			Metadata:      workflowTimeoutDataInBytes,
		},
		{
			InvariantType: timeout.TimeoutTypeActivity.String(),
			Reason:        "START_TO_CLOSE",
			Metadata:      activityTimeoutDataInBytes,
		},
		{
			InvariantType: timeout.TimeoutTypeDecision.String(),
			Reason:        "START_TO_CLOSE",
			Metadata:      descTimeoutDataInBytes,
		},
		{
			InvariantType: timeout.TimeoutTypeChildWorkflow.String(),
			Reason:        "START_TO_CLOSE",
			Metadata:      childWorkflowTimeoutDataInBytes,
		},
	}
	timeoutIssues := []*timeoutIssuesResult{
		{
			InvariantType:    timeout.TimeoutTypeExecution.String(),
			Reason:           "START_TO_CLOSE",
			ExecutionTimeout: &workflowTimeoutData,
		},
		{
			InvariantType:   timeout.TimeoutTypeActivity.String(),
			Reason:          "START_TO_CLOSE",
			ActivityTimeout: &activityTimeoutData,
		},
		{
			InvariantType:   timeout.TimeoutTypeDecision.String(),
			Reason:          "START_TO_CLOSE",
			DecisionTimeout: &descTimeoutData,
		},
		{
			InvariantType:  timeout.TimeoutTypeChildWorkflow.String(),
			Reason:         "START_TO_CLOSE",
			ChildWfTimeout: &childWorkflowTimeoutData,
		},
	}
	result, err := retrieveTimeoutIssues(issues)
	s.NoError(err)
	s.Equal(timeoutIssues, result)
}

func (s *diagnosticsWorkflowTestSuite) Test__retrieveTimeoutRootCause() {
	taskListBacklog := int64(10)
	pollersMetadataInBytes, err := json.Marshal(timeout.PollersMetadata{TaskListBacklog: taskListBacklog})
	s.NoError(err)
	heartBeatingMetadataInBytes, err := json.Marshal(timeout.HeartbeatingMetadata{TimeElapsed: 5 * time.Second})
	s.NoError(err)
	rootCause := []invariant.InvariantRootCauseResult{
		{
			RootCause: invariant.RootCauseTypePollersStatus,
			Metadata:  pollersMetadataInBytes,
		},
		{
			RootCause: invariant.RootCauseTypeNoHeartBeatTimeoutNoRetryPolicy,
			Metadata:  heartBeatingMetadataInBytes,
		},
	}
	timeoutRootCause := []*timeoutRootCauseResult{
		{
			RootCauseType:   invariant.RootCauseTypePollersStatus.String(),
			PollersMetadata: &timeout.PollersMetadata{TaskListBacklog: taskListBacklog},
		},
		{
			RootCauseType:        invariant.RootCauseTypeNoHeartBeatTimeoutNoRetryPolicy.String(),
			HeartBeatingMetadata: &timeout.HeartbeatingMetadata{TimeElapsed: 5 * time.Second},
		},
	}
	result, err := retrieveTimeoutRootCause(rootCause)
	s.NoError(err)
	s.Equal(timeoutRootCause, result)
}

func (s *diagnosticsWorkflowTestSuite) Test__retrieveFailureIssues() {
	actMetadata := failure.FailureMetadata{
		Identity: "localhost",
		ActivityScheduled: &types.ActivityTaskScheduledEventAttributes{
			ActivityID:   "101",
			ActivityType: &types.ActivityType{Name: "test-activity"},
		},
		ActivityStarted: &types.ActivityTaskStartedEventAttributes{
			Identity: "localhost",
			Attempt:  0,
		},
	}
	actMetadataInBytes, err := json.Marshal(actMetadata)
	s.NoError(err)
	issues := []invariant.InvariantCheckResult{
		{
			InvariantType: failure.ActivityFailed.String(),
			Reason:        failure.CustomError.String(),
			Metadata:      actMetadataInBytes,
		},
	}
	failureIssues := []*failureIssuesResult{
		{
			InvariantType: failure.ActivityFailed.String(),
			Reason:        failure.CustomError.String(),
			Metadata:      &actMetadata,
		},
	}
	result, err := retrieveFailureIssues(issues)
	s.NoError(err)
	s.Equal(failureIssues, result)
}

func (s *diagnosticsWorkflowTestSuite) Test__retrieveFailureRootCause() {
	blobSizeMetadataInBytes, err := json.Marshal(failure.BlobSizeMetadata{
		BlobSizeWarnLimit:  5,
		BlobSizeErrorLimit: 10,
	})
	s.NoError(err)
	rootCause := []invariant.InvariantRootCauseResult{
		{
			RootCause: invariant.RootCauseTypeServiceSideIssue,
		},
		{
			RootCause: invariant.RootCauseTypeBlobSizeLimit,
			Metadata:  blobSizeMetadataInBytes,
		},
	}
	failureRootCause := []*failureRootCauseResult{
		{
			RootCauseType: invariant.RootCauseTypeServiceSideIssue.String(),
		},
		{
			RootCauseType: invariant.RootCauseTypeBlobSizeLimit.String(),
			BlobSizeMetadata: &failure.BlobSizeMetadata{
				BlobSizeWarnLimit:  5,
				BlobSizeErrorLimit: 10,
			},
		},
	}
	result, err := retrieveFailureRootCause(rootCause)
	s.NoError(err)
	s.Equal(failureRootCause, result)
}

func (s *diagnosticsWorkflowTestSuite) Test__retrieveRetryIssues() {
	retryMetadata := retry.RetryMetadata{
		RetryPolicy: &types.RetryPolicy{
			InitialIntervalInSeconds: 1,
			MaximumAttempts:          1,
		},
	}
	retryMetadataInBytes, err := json.Marshal(retryMetadata)
	s.NoError(err)
	issues := []invariant.InvariantCheckResult{
		{
			InvariantType: retry.ActivityRetryIssue.String(),
			Reason:        retry.RetryPolicyValidationMaxAttempts.String(),
			Metadata:      retryMetadataInBytes,
		},
		{
			InvariantType: retry.WorkflowRetryIssue.String(),
			Reason:        retry.RetryPolicyValidationMaxAttempts.String(),
			Metadata:      retryMetadataInBytes,
		},
	}
	retryIssues := []*retryIssuesResult{
		{
			InvariantType: retry.ActivityRetryIssue.String(),
			Reason:        retry.RetryPolicyValidationMaxAttempts.String(),
			Metadata:      retryMetadata,
		},
		{
			InvariantType: retry.WorkflowRetryIssue.String(),
			Reason:        retry.RetryPolicyValidationMaxAttempts.String(),
			Metadata:      retryMetadata,
		},
	}
	result, err := retrieveRetryIssues(issues)
	s.NoError(err)
	s.Equal(retryIssues, result)
}
