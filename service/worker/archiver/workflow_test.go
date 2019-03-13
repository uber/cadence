// Copyright (c) 2017 Uber Technologies, Inc.
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

package archiver

import (
	"errors"
	"go.uber.org/cadence/workflow"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/uber-common/bark"
	"github.com/uber-go/tally"
	"github.com/uber/cadence/common/metrics"
	mmocks "github.com/uber/cadence/common/metrics/mocks"
	"go.uber.org/cadence/testsuite"
)

type workflowSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestWorkflowSuite(t *testing.T) {
	suite.Run(t, new(workflowSuite))
}

func (s *workflowSuite) TestArchivalWorkflow_Fail_ReadConfigError() {
	mockMetricsClient := &mmocks.Client{}
	mockMetricsClient.On("IncCounter", metrics.ArchiverArchivalWorkflowScope, metrics.ArchiverWorkflowStartedCount).Once()
	mockMetricsClient.On("StartTimer", metrics.ArchiverArchivalWorkflowScope, metrics.CadenceLatency).Return(tally.NewStopwatch(time.Now(), &nopStopwatchRecorder{})).Once()
	mockMetricsClient.On("IncCounter", metrics.ArchiverArchivalWorkflowScope, metrics.ArchiverReadDynamicConfigErrorCount).Once()
	globalLogger = bark.NewNopLogger()
	globalMetricsClient = mockMetricsClient

	env := s.NewTestWorkflowEnvironment()
	actErr := errors.New("error reading dynamic config")
	env.OnActivity(readConfigActivity, mock.Anything).Return(readConfigActivityResult{}, actErr)
	env.ExecuteWorkflow(archivalWorkflow, nil)

	env.AssertExpectations(s.T())
	s.True(env.IsWorkflowCompleted())
	s.Equal(actErr.Error(), env.GetWorkflowError().Error())
}

func (s *workflowSuite) TestArchivalWorkflow_Fail_HashesDoNotEqual() {
	mockMetricsClient := &mmocks.Client{}
	mockMetricsClient.On("IncCounter", metrics.ArchiverArchivalWorkflowScope, metrics.ArchiverWorkflowStartedCount).Once()
	mockMetricsClient.On("StartTimer", metrics.ArchiverArchivalWorkflowScope, metrics.CadenceLatency).Return(tally.NewStopwatch(time.Now(), &nopStopwatchRecorder{})).Once()
	mockMetricsClient.On("StartTimer", metrics.ArchiverArchivalWorkflowScope, metrics.ArchiverHandleAllRequestsLatency).Return(tally.NewStopwatch(time.Now(), &nopStopwatchRecorder{})).Once()
	mockMetricsClient.On("AddCounter", metrics.ArchiverArchivalWorkflowScope, metrics.ArchiverNumPumpedRequestsCount, int64(3)).Once()
	mockMetricsClient.On("AddCounter", metrics.ArchiverArchivalWorkflowScope, metrics.ArchiverNumHandledRequestsCount, int64(3)).Once()
	mockMetricsClient.On("IncCounter", metrics.ArchiverArchivalWorkflowScope, metrics.ArchiverPumpedNotEqualHandledCount).Once()
	mockArchiver := &ArchiverMock{}
	mockArchiver.On("Start").Once()
	mockArchiver.On("Finished").Return([]uint64{9, 7, 0}).Once()
	mockPump := &PumpMock{}
	mockPump.On("Run").Return(PumpResult{
		PumpedHashes: []uint64{8, 7, 0},
	}).Once()
	globalLogger = bark.NewNopLogger()
	globalMetricsClient = mockMetricsClient
	testOverridePump = mockPump
	testOverrideArchiver = mockArchiver

	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(readConfigActivity, mock.Anything).Return(readConfigActivityResult{}, nil)
	env.ExecuteWorkflow(archivalWorkflow, nil)

	s.True(env.IsWorkflowCompleted())
	_, ok := env.GetWorkflowError().(*workflow.ContinueAsNewError)
	s.True(ok, "Called ContinueAsNew")
	env.AssertExpectations(s.T())
}

func (s *workflowSuite) TestArchivalWorkflow_Exit_TimeoutWithoutSignals() {
	mockMetricsClient := &mmocks.Client{}
	mockMetricsClient.On("IncCounter", metrics.ArchiverArchivalWorkflowScope, metrics.ArchiverWorkflowStartedCount).Once()
	mockMetricsClient.On("StartTimer", metrics.ArchiverArchivalWorkflowScope, metrics.CadenceLatency).Return(tally.NewStopwatch(time.Now(), &nopStopwatchRecorder{})).Once()
	mockMetricsClient.On("StartTimer", metrics.ArchiverArchivalWorkflowScope, metrics.ArchiverHandleAllRequestsLatency).Return(tally.NewStopwatch(time.Now(), &nopStopwatchRecorder{})).Once()
	mockMetricsClient.On("AddCounter", metrics.ArchiverArchivalWorkflowScope, metrics.ArchiverNumPumpedRequestsCount, int64(0)).Once()
	mockMetricsClient.On("AddCounter", metrics.ArchiverArchivalWorkflowScope, metrics.ArchiverNumHandledRequestsCount, int64(0)).Once()
	mockMetricsClient.On("IncCounter", metrics.ArchiverArchivalWorkflowScope, metrics.ArchiverWorkflowStoppingCount).Once()
	mockArchiver := &ArchiverMock{}
	mockArchiver.On("Start").Once()
	mockArchiver.On("Finished").Return([]uint64{}).Once()
	mockPump := &PumpMock{}
	mockPump.On("Run").Return(PumpResult{
		PumpedHashes: []uint64{},
		TimeoutWithoutSignals: true,
	}).Once()
	globalLogger = bark.NewNopLogger()
	globalMetricsClient = mockMetricsClient
	testOverridePump = mockPump
	testOverrideArchiver = mockArchiver

	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(readConfigActivity, mock.Anything).Return(readConfigActivityResult{}, nil)
	env.ExecuteWorkflow(archivalWorkflow, nil)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	env.AssertExpectations(s.T())
}

func (s *workflowSuite) TestArchivalWorkflow_Success() {
	mockMetricsClient := &mmocks.Client{}
	mockMetricsClient.On("IncCounter", metrics.ArchiverArchivalWorkflowScope, metrics.ArchiverWorkflowStartedCount).Once()
	mockMetricsClient.On("StartTimer", metrics.ArchiverArchivalWorkflowScope, metrics.CadenceLatency).Return(tally.NewStopwatch(time.Now(), &nopStopwatchRecorder{})).Once()
	mockMetricsClient.On("StartTimer", metrics.ArchiverArchivalWorkflowScope, metrics.ArchiverHandleAllRequestsLatency).Return(tally.NewStopwatch(time.Now(), &nopStopwatchRecorder{})).Once()
	mockMetricsClient.On("AddCounter", metrics.ArchiverArchivalWorkflowScope, metrics.ArchiverNumPumpedRequestsCount, int64(5)).Once()
	mockMetricsClient.On("AddCounter", metrics.ArchiverArchivalWorkflowScope, metrics.ArchiverNumHandledRequestsCount, int64(5)).Once()
	mockArchiver := &ArchiverMock{}
	mockArchiver.On("Start").Once()
	mockArchiver.On("Finished").Return([]uint64{1, 2, 3, 4, 5}).Once()
	mockPump := &PumpMock{}
	mockPump.On("Run").Return(PumpResult{
		PumpedHashes: []uint64{1, 2, 3, 4, 5},
	}).Once()
	globalLogger = bark.NewNopLogger()
	globalMetricsClient = mockMetricsClient
	testOverridePump = mockPump
	testOverrideArchiver = mockArchiver

	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(readConfigActivity, mock.Anything).Return(readConfigActivityResult{}, nil)
	env.ExecuteWorkflow(archivalWorkflow, nil)

	s.True(env.IsWorkflowCompleted())
	_, ok := env.GetWorkflowError().(*workflow.ContinueAsNewError)
	s.True(ok, "Called ContinueAsNew")
	env.AssertExpectations(s.T())
}