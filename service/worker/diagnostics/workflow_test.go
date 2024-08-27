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
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/worker/diagnostics/invariants"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/resource"
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
		svcClient:  mockResource.GetSDKClient(),
		clientBean: mockResource.ClientBean,
	}

	s.T().Cleanup(func() {
		mockResource.Finish(s.T())
	})

	s.workflowEnv.RegisterWorkflowWithOptions(s.dw.DiagnosticsWorkflow, workflow.RegisterOptions{Name: diagnosticsWorkflow})
	s.workflowEnv.RegisterActivityWithOptions(s.dw.retrieveExecutionHistory, activity.RegisterOptions{Name: retrieveWfExecutionHistoryActivity})
	s.workflowEnv.RegisterActivityWithOptions(s.dw.identifyTimeouts, activity.RegisterOptions{Name: identifyTimeoutsActivity})
	s.workflowEnv.RegisterActivityWithOptions(s.dw.rootCauseTimeouts, activity.RegisterOptions{Name: rootCauseTimeoutsActivity})
}

func (s *diagnosticsWorkflowTestSuite) TearDownTest() {
	s.workflowEnv.AssertExpectations(s.T())
}

func (s *diagnosticsWorkflowTestSuite) TestWorkflow() {
	params := &DiagnosticsWorkflowInput{
		Domain:     "test",
		WorkflowID: "123",
		RunID:      "abc",
	}
	workflowTimeoutData := invariants.ExecutionTimeoutMetadata{
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
	issues := []invariants.InvariantCheckResult{
		{
			InvariantType: invariants.TimeoutTypeExecution.String(),
			Reason:        "START_TO_CLOSE",
			Metadata:      workflowTimeoutDataInBytes,
		},
	}
	taskListBacklog := int64(10)
	taskListBacklogInBytes, err := json.Marshal(taskListBacklog)
	s.NoError(err)
	rootCause := []invariants.InvariantRootCauseResult{
		{
			RootCause: invariants.RootCauseTypePollersStatus,
			Metadata:  taskListBacklogInBytes,
		},
	}
	s.workflowEnv.OnActivity(retrieveWfExecutionHistoryActivity, mock.Anything, mock.Anything).Return(nil, nil)
	s.workflowEnv.OnActivity(identifyTimeoutsActivity, mock.Anything, mock.Anything).Return(issues, nil)
	s.workflowEnv.OnActivity(rootCauseTimeoutsActivity, mock.Anything, mock.Anything).Return(rootCause, nil)
	s.workflowEnv.ExecuteWorkflow(diagnosticsWorkflow, params)
	s.True(s.workflowEnv.IsWorkflowCompleted())
	var result DiagnosticsWorkflowResult
	s.NoError(s.workflowEnv.GetWorkflowResult(&result))
	s.ElementsMatch(issues, result.Issues)
	s.ElementsMatch(rootCause, result.RootCause)
}

func (s *diagnosticsWorkflowTestSuite) TestWorkflow_Error() {
	params := &DiagnosticsWorkflowInput{
		Domain:     "test",
		WorkflowID: "123",
		RunID:      "abc",
	}
	mockErr := errors.New("mockErr")
	errExpected := fmt.Errorf("IdentifyTimeouts: %w", mockErr)
	s.workflowEnv.OnActivity(retrieveWfExecutionHistoryActivity, mock.Anything, mock.Anything).Return(nil, nil)
	s.workflowEnv.OnActivity(identifyTimeoutsActivity, mock.Anything, mock.Anything).Return(nil, mockErr)
	s.workflowEnv.ExecuteWorkflow(diagnosticsWorkflow, params)
	s.True(s.workflowEnv.IsWorkflowCompleted())
	s.Error(s.workflowEnv.GetWorkflowError())
	s.EqualError(s.workflowEnv.GetWorkflowError(), errExpected.Error())
}
