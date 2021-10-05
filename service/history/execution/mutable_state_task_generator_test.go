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

package execution

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/service/history/constants"
)

type (
	mutableStateTaskGeneratorSuite struct {
		suite.Suite
		*require.Assertions

		controller       *gomock.Controller
		mockDomainCache  *cache.MockDomainCache
		mockMutableState *MockMutableState

		taskGenerator *mutableStateTaskGeneratorImpl
	}
)

func TestMutableStateTaskGeneratorSuite(t *testing.T) {
	s := new(mutableStateTaskGeneratorSuite)
	suite.Run(t, s)
}

func (s *mutableStateTaskGeneratorSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockDomainCache = cache.NewMockDomainCache(s.controller)
	s.mockMutableState = NewMockMutableState(s.controller)

	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestDomainID).Return(constants.TestGlobalDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestTargetDomainID).Return(constants.TestGlobalTargetDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestRemoteTargetDomainID).Return(constants.TestGlobalRemoteTargetDomainEntry, nil).AnyTimes()

	s.taskGenerator = NewMutableStateTaskGenerator(
		constants.TestClusterMetadata,
		s.mockDomainCache,
		loggerimpl.NewLoggerForTest(s.Suite),
		s.mockMutableState,
	).(*mutableStateTaskGeneratorImpl)
}

func (s *mutableStateTaskGeneratorSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *mutableStateTaskGeneratorSuite) TestIsCrossClusterTask() {
	testCases := []struct {
		sourceDomainID string
		targetDomainID string
		isCrossCluster bool
		targetCluster  string
	}{
		{
			sourceDomainID: constants.TestDomainID,
			targetDomainID: constants.TestDomainID,
			isCrossCluster: false,
			targetCluster:  "",
		},
		{
			// source domain is passive in the current cluster
			sourceDomainID: constants.TestRemoteTargetDomainID,
			targetDomainID: constants.TestDomainID,
			isCrossCluster: false,
			targetCluster:  "",
		},
		{
			sourceDomainID: constants.TestDomainID,
			targetDomainID: constants.TestTargetDomainID,
			isCrossCluster: false,
			targetCluster:  "",
		},
		{
			sourceDomainID: constants.TestDomainID,
			targetDomainID: constants.TestRemoteTargetDomainID,
			isCrossCluster: true,
			targetCluster:  cluster.TestAlternativeClusterName,
		},
	}

	for _, tc := range testCases {
		s.mockMutableState.EXPECT().GetExecutionInfo().Return(&persistence.WorkflowExecutionInfo{
			DomainID: constants.TestDomainID,
		})

		targetCluster, isCrossCluster, err := s.taskGenerator.isCrossClusterTask(tc.targetDomainID)
		s.NoError(err)
		s.Equal(tc.isCrossCluster, isCrossCluster)
		s.Equal(tc.targetCluster, targetCluster)
	}
}

func (s *mutableStateTaskGeneratorSuite) TestGenerateCrossClusterTaskFromTransferTask() {
	targetCluster := cluster.TestAlternativeClusterName
	now := time.Now()
	testCases := []struct {
		transferTask             *persistence.TransferTaskInfo
		expectError              bool
		expectedCrossClusterTask persistence.Task
	}{
		{
			transferTask: &persistence.TransferTaskInfo{
				TaskType: persistence.TransferTaskTypeActivityTask,
			},
			expectError: true,
		},
		{
			transferTask: &persistence.TransferTaskInfo{
				TaskType:                persistence.TransferTaskTypeCancelExecution,
				TargetDomainID:          constants.TestTargetDomainID,
				TargetWorkflowID:        constants.TestWorkflowID,
				TargetRunID:             constants.TestRunID,
				TargetChildWorkflowOnly: false,
				ScheduleID:              int64(123),
			},
			expectError: false,
			expectedCrossClusterTask: &persistence.CrossClusterCancelExecutionTask{
				TargetCluster: targetCluster,
				CancelExecutionTask: persistence.CancelExecutionTask{
					TargetDomainID:          constants.TestTargetDomainID,
					TargetWorkflowID:        constants.TestWorkflowID,
					TargetRunID:             constants.TestRunID,
					TargetChildWorkflowOnly: false,
					InitiatedID:             int64(123),
				},
			},
		},
		{
			transferTask: &persistence.TransferTaskInfo{
				TaskType:                persistence.TransferTaskTypeSignalExecution,
				TargetDomainID:          constants.TestTargetDomainID,
				TargetWorkflowID:        constants.TestWorkflowID,
				TargetRunID:             constants.TestRunID,
				TargetChildWorkflowOnly: false,
				ScheduleID:              int64(123),
			},
			expectError: false,
			expectedCrossClusterTask: &persistence.CrossClusterSignalExecutionTask{
				TargetCluster: targetCluster,
				SignalExecutionTask: persistence.SignalExecutionTask{
					TargetDomainID:          constants.TestTargetDomainID,
					TargetWorkflowID:        constants.TestWorkflowID,
					TargetRunID:             constants.TestRunID,
					TargetChildWorkflowOnly: false,
					InitiatedID:             int64(123),
				},
			},
		},
		{
			transferTask: &persistence.TransferTaskInfo{
				TaskType:         persistence.TransferTaskTypeStartChildExecution,
				TargetDomainID:   constants.TestTargetDomainID,
				TargetWorkflowID: constants.TestWorkflowID,
				ScheduleID:       int64(123),
			},
			expectError: false,
			expectedCrossClusterTask: &persistence.CrossClusterStartChildExecutionTask{
				TargetCluster: targetCluster,
				StartChildExecutionTask: persistence.StartChildExecutionTask{
					TargetDomainID:   constants.TestTargetDomainID,
					TargetWorkflowID: constants.TestWorkflowID,
					InitiatedID:      int64(123),
				},
			},
		},
		{
			transferTask: &persistence.TransferTaskInfo{
				TaskType:         persistence.TransferTaskTypeCloseExecution,
				TargetDomainID:   constants.TestTargetDomainID,
				TargetWorkflowID: constants.TestWorkflowID,
				TargetRunID:      constants.TestRunID,
				ScheduleID:       int64(123),
			},
			expectError: false,
			expectedCrossClusterTask: &persistence.CrossClusterRecordChildWorkflowExecutionCompleteTask{
				TargetCluster:                       targetCluster,
				RecordWorkflowExecutionCompleteTask: persistence.RecordWorkflowExecutionCompleteTask{},
			},
		},
	}

	for _, tc := range testCases {
		var actualCrossClusterTask persistence.Task
		if !tc.expectError {
			tc.transferTask.Version = int64(101)
			tc.expectedCrossClusterTask.SetVersion(int64(101))
			tc.transferTask.VisibilityTimestamp = now
			tc.expectedCrossClusterTask.SetVisibilityTimestamp(now)

			s.mockMutableState.EXPECT().AddCrossClusterTasks(gomock.Any()).Do(
				func(crossClusterTasks ...persistence.Task) {
					actualCrossClusterTask = crossClusterTasks[0]
				},
			).MaxTimes(1)
		}

		err := s.taskGenerator.GenerateCrossClusterTaskFromTransferTask(tc.transferTask, targetCluster)
		if tc.expectError {
			s.Error(err)
		} else {
			s.Equal(tc.expectedCrossClusterTask, actualCrossClusterTask)
		}
	}
}

func (s *mutableStateTaskGeneratorSuite) TestGenerateFromCrossClusterTask() {
	testCases := []struct {
		sourceActive     bool
		crossClusterTask *persistence.CrossClusterTaskInfo
		generatedTask    persistence.Task
	}{
		{
			sourceActive: true,
			crossClusterTask: &persistence.CrossClusterTaskInfo{
				TaskType:         persistence.CrossClusterTaskTypeStartChildExecution,
				TargetDomainID:   constants.TestTargetDomainID,
				TargetWorkflowID: constants.TestWorkflowID,
				ScheduleID:       int64(123),
			},
			generatedTask: &persistence.StartChildExecutionTask{
				TargetDomainID:   constants.TestTargetDomainID,
				TargetWorkflowID: constants.TestWorkflowID,
				InitiatedID:      int64(123),
			},
		},
		{
			sourceActive: true,
			crossClusterTask: &persistence.CrossClusterTaskInfo{
				TaskType:                persistence.CrossClusterTaskTypeSignalExecution,
				TargetDomainID:          constants.TestRemoteTargetDomainID,
				TargetWorkflowID:        constants.TestWorkflowID,
				TargetRunID:             constants.TestRunID,
				TargetChildWorkflowOnly: false,
				ScheduleID:              int64(123),
			},
			generatedTask: &persistence.CrossClusterSignalExecutionTask{
				TargetCluster: cluster.TestAlternativeClusterName,
				SignalExecutionTask: persistence.SignalExecutionTask{
					TargetDomainID:          constants.TestRemoteTargetDomainID,
					TargetWorkflowID:        constants.TestWorkflowID,
					TargetRunID:             constants.TestRunID,
					TargetChildWorkflowOnly: false,
					InitiatedID:             int64(123),
				},
			},
		},
		{
			sourceActive: false,
			crossClusterTask: &persistence.CrossClusterTaskInfo{
				TaskType:                persistence.CrossClusterTaskTypeCancelExecution,
				TargetDomainID:          constants.TestTargetDomainID,
				TargetWorkflowID:        constants.TestWorkflowID,
				TargetRunID:             constants.TestRunID,
				TargetChildWorkflowOnly: false,
				ScheduleID:              int64(123),
			},
			generatedTask: &persistence.CancelExecutionTask{
				TargetDomainID:          constants.TestTargetDomainID,
				TargetWorkflowID:        constants.TestWorkflowID,
				TargetRunID:             constants.TestRunID,
				TargetChildWorkflowOnly: false,
				InitiatedID:             int64(123),
			},
		},
		{
			sourceActive: true,
			crossClusterTask: &persistence.CrossClusterTaskInfo{
				TaskType:         persistence.CrossClusterTaskTypeRecordChildWorkflowExeuctionComplete,
				TargetDomainID:   constants.TestRemoteTargetDomainID,
				TargetWorkflowID: constants.TestWorkflowID,
				TargetRunID:      constants.TestRunID,
				ScheduleID:       int64(123),
			},
			generatedTask: &persistence.CrossClusterRecordChildWorkflowExecutionCompleteTask{
				TargetCluster:                       cluster.TestAlternativeClusterName,
				RecordWorkflowExecutionCompleteTask: persistence.RecordWorkflowExecutionCompleteTask{},
			},
		},
		{
			sourceActive: false,
			crossClusterTask: &persistence.CrossClusterTaskInfo{
				TaskType:         persistence.CrossClusterTaskTypeRecordChildWorkflowExeuctionComplete,
				TargetDomainID:   constants.TestRemoteTargetDomainID,
				TargetWorkflowID: constants.TestWorkflowID,
				TargetRunID:      constants.TestRunID,
				ScheduleID:       int64(123),
			},
			generatedTask: &persistence.CloseExecutionTask{},
		},
		{
			sourceActive: true,
			crossClusterTask: &persistence.CrossClusterTaskInfo{
				TaskType:         persistence.CrossClusterTaskTypeRecordChildWorkflowExeuctionComplete,
				TargetDomainID:   constants.TestTargetDomainID,
				TargetWorkflowID: constants.TestWorkflowID,
				TargetRunID:      constants.TestRunID,
				ScheduleID:       int64(123),
			},
			generatedTask: &persistence.CloseExecutionTask{},
		},
		{
			sourceActive: true,
			crossClusterTask: &persistence.CrossClusterTaskInfo{
				TaskType:         persistence.CrossClusterTaskTypeApplyParentPolicy,
				TargetDomainID:   constants.TestRemoteTargetDomainID,
				TargetWorkflowID: constants.TestWorkflowID,
				TargetRunID:      constants.TestRunID,
				ScheduleID:       int64(123),
			},
			generatedTask: &persistence.CrossClusterApplyParentClosePolicyTask{
				TargetCluster:              cluster.TestAlternativeClusterName,
				ApplyParentClosePolicyTask: persistence.ApplyParentClosePolicyTask{},
			},
		},
		{
			sourceActive: false,
			crossClusterTask: &persistence.CrossClusterTaskInfo{
				TaskType:         persistence.CrossClusterTaskTypeApplyParentPolicy,
				TargetDomainID:   constants.TestRemoteTargetDomainID,
				TargetWorkflowID: constants.TestWorkflowID,
				TargetRunID:      constants.TestRunID,
				ScheduleID:       int64(123),
			},
			generatedTask: &persistence.CloseExecutionTask{},
		},
		{
			sourceActive: true,
			crossClusterTask: &persistence.CrossClusterTaskInfo{
				TaskType:         persistence.CrossClusterTaskTypeApplyParentPolicy,
				TargetDomainID:   constants.TestTargetDomainID,
				TargetWorkflowID: constants.TestWorkflowID,
				TargetRunID:      constants.TestRunID,
				ScheduleID:       int64(123),
			},
			generatedTask: &persistence.CloseExecutionTask{},
		},
	}

	for _, tc := range testCases {
		if tc.sourceActive {
			tc.crossClusterTask.DomainID = constants.TestDomainID
			s.mockMutableState.EXPECT().GetDomainEntry().Return(constants.TestGlobalDomainEntry).Times(1)
		} else {
			tc.crossClusterTask.DomainID = constants.TestRemoteTargetDomainID
			s.mockMutableState.EXPECT().GetDomainEntry().Return(constants.TestGlobalRemoteTargetDomainEntry).Times(1)
		}
		targetActive := tc.crossClusterTask.TargetDomainID == constants.TestTargetDomainID

		var actualGeneratedTask persistence.Task
		mockDoFn := func(tasks ...persistence.Task) {
			actualGeneratedTask = tasks[0]
		}
		if !tc.sourceActive || targetActive {
			s.mockMutableState.EXPECT().AddTransferTasks(gomock.Any()).Do(mockDoFn).Times(1)
		} else {
			s.mockMutableState.EXPECT().AddCrossClusterTasks(gomock.Any()).Do(mockDoFn).Times(1)
		}

		err := s.taskGenerator.GenerateFromCrossClusterTask(tc.crossClusterTask)
		s.NoError(err)
		s.Equal(tc.generatedTask, actualGeneratedTask)
	}
}

func (s *mutableStateTaskGeneratorSuite) TestGetNextDecisionTimeout() {
	defaultStartToCloseTimeout := 10 * time.Second
	expectedResult := []time.Duration{
		defaultStartToCloseTimeout,
		defaultStartToCloseTimeout,
		defaultInitIntervalForDecisionRetry,
		defaultInitIntervalForDecisionRetry * 2,
		defaultInitIntervalForDecisionRetry * 4,
		defaultMaxIntervalForDecisionRetry,
		defaultMaxIntervalForDecisionRetry,
		defaultMaxIntervalForDecisionRetry,
	}
	for i := 0; i < len(expectedResult); i++ {
		next := getNextDecisionTimeout(int64(i), defaultStartToCloseTimeout)
		expected := expectedResult[i]
		min, max := getNextBackoffRange(expected)
		s.True(next >= min, "NextBackoff too low: actual: %v, expected: %v", next, expected)
		s.True(next <= max, "NextBackoff too high: actual: %v, expected: %v", next, expected)
	}
}

func getNextBackoffRange(duration time.Duration) (time.Duration, time.Duration) {
	rangeMin := time.Duration((1 - defaultJitterCoefficient) * float64(duration))
	return rangeMin, duration
}
