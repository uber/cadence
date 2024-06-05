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

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
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
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestParentDomainID).Return(constants.TestGlobalParentDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestChildDomainID).Return(constants.TestGlobalChildDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(constants.TestChildDomainName).Return(constants.TestChildDomainID, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestTargetDomainID).Return(constants.TestGlobalTargetDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(constants.TestTargetDomainName).Return(constants.TestTargetDomainID, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestRemoteTargetDomainID).Return(constants.TestGlobalRemoteTargetDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(constants.TestRemoteTargetDomainName).Return(constants.TestRemoteTargetDomainID, nil).AnyTimes()

	s.taskGenerator = NewMutableStateTaskGenerator(
		constants.TestClusterMetadata,
		s.mockDomainCache,
		s.mockMutableState,
	).(*mutableStateTaskGeneratorImpl)
}

func (s *mutableStateTaskGeneratorSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *mutableStateTaskGeneratorSuite) TestGenerateWorkflowCloseTasks_JitteredDeletion() {
	now := time.Now()
	version := int64(123)
	closeEvent := &types.HistoryEvent{
		EventType: types.EventTypeWorkflowExecutionCompleted.Ptr(),
		Timestamp: common.Int64Ptr(now.UnixNano()),
		Version:   version,
	}
	domainEntry, err := s.mockDomainCache.GetDomainByID(constants.TestDomainID)
	s.NoError(err)
	retention := time.Duration(domainEntry.GetRetentionDays(constants.TestWorkflowID)) * time.Hour * 24
	testCases := GenerateWorkflowCloseTasksTestCases(retention, closeEvent, now)

	for _, tc := range testCases {
		// create new mockMutableState so can we can setup separete mock for each test case
		mockMutableState := NewMockMutableState(s.controller)
		taskGenerator := NewMutableStateTaskGenerator(
			constants.TestClusterMetadata,
			s.mockDomainCache,
			mockMutableState,
		)

		var transferTasks []persistence.Task
		var crossClusterTasks []persistence.Task
		var timerTasks []persistence.Task
		mockMutableState.EXPECT().GetDomainEntry().Return(domainEntry).AnyTimes()
		mockMutableState.EXPECT().AddTransferTasks(gomock.Any()).Do(func(tasks ...persistence.Task) {
			transferTasks = tasks
		}).MaxTimes(1)
		mockMutableState.EXPECT().AddTimerTasks(gomock.Any()).Do(func(tasks ...persistence.Task) {
			timerTasks = tasks
		}).MaxTimes(1)

		tc.setupFn(mockMutableState)
		err := taskGenerator.GenerateWorkflowCloseTasks(closeEvent, 60)
		s.NoError(err)

		actualGeneratedTasks := append(transferTasks, crossClusterTasks...)
		for _, task := range actualGeneratedTasks {
			// force set visibility timestamp since that field is not assigned
			// for transfer and cross cluster in GenerateWorkflowCloseTasks
			// it will be set by shard context
			// set it to now so that we can easily test if other fields are equal
			task.SetVisibilityTimestamp(now)
		}
		for _, task := range timerTasks {
			// force set timer tasks because with jittering the timertask visibility time stamp
			// is not consistent for each run.
			// as long as code doesn't break during generation we should be ok
			task.SetVisibilityTimestamp(time.Unix(0, closeEvent.GetTimestamp()).Add(retention))
		}
		actualGeneratedTasks = append(actualGeneratedTasks, timerTasks...)
		s.Equal(tc.generatedTasks, actualGeneratedTasks)
	}
}

func (s *mutableStateTaskGeneratorSuite) TestGenerateWorkflowCloseTasks() {
	now := time.Now()
	version := int64(123)
	closeEvent := &types.HistoryEvent{
		EventType: types.EventTypeWorkflowExecutionCompleted.Ptr(),
		Timestamp: common.Int64Ptr(now.UnixNano()),
		Version:   version,
	}
	domainEntry, err := s.mockDomainCache.GetDomainByID(constants.TestDomainID)
	s.NoError(err)
	retention := time.Duration(domainEntry.GetRetentionDays(constants.TestWorkflowID)) * time.Hour * 24
	testCases := GenerateWorkflowCloseTasksTestCases(retention, closeEvent, now)

	for _, tc := range testCases {
		// create new mockMutableState so can we can setup separete mock for each test case
		mockMutableState := NewMockMutableState(s.controller)
		taskGenerator := NewMutableStateTaskGenerator(
			constants.TestClusterMetadata,
			s.mockDomainCache,
			mockMutableState,
		)

		var transferTasks []persistence.Task
		var timerTasks []persistence.Task
		mockMutableState.EXPECT().GetDomainEntry().Return(domainEntry).AnyTimes()
		mockMutableState.EXPECT().AddTransferTasks(gomock.Any()).Do(func(tasks ...persistence.Task) {
			transferTasks = tasks
		}).MaxTimes(1)
		mockMutableState.EXPECT().AddTimerTasks(gomock.Any()).Do(func(tasks ...persistence.Task) {
			timerTasks = tasks
		}).MaxTimes(1)

		tc.setupFn(mockMutableState)
		err := taskGenerator.GenerateWorkflowCloseTasks(closeEvent, 1)
		s.NoError(err)

		actualGeneratedTasks := transferTasks
		for _, task := range actualGeneratedTasks {
			// force set visibility timestamp since that field is not assigned
			// for transfer and cross cluster in GenerateWorkflowCloseTasks
			// it will be set by shard context
			// set it to now so that we can easily test if other fields are equal
			task.SetVisibilityTimestamp(now)
		}
		actualGeneratedTasks = append(actualGeneratedTasks, timerTasks...)
		s.Equal(tc.generatedTasks, actualGeneratedTasks)
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

func GenerateWorkflowCloseTasksTestCases(retention time.Duration, closeEvent *types.HistoryEvent, now time.Time) []struct {
	setupFn        func(mockMutableState *MockMutableState)
	generatedTasks []persistence.Task
} {
	version := int64(123)
	return []struct {
		setupFn        func(mockMutableState *MockMutableState)
		generatedTasks []persistence.Task
	}{
		{
			// no parent, no children
			setupFn: func(mockMutableState *MockMutableState) {
				mockMutableState.EXPECT().GetExecutionInfo().Return(&persistence.WorkflowExecutionInfo{
					DomainID:   constants.TestDomainID,
					WorkflowID: constants.TestWorkflowID,
					RunID:      constants.TestRunID,
				}).AnyTimes()
				mockMutableState.EXPECT().HasParentExecution().Return(false).AnyTimes()
				mockMutableState.EXPECT().GetPendingChildExecutionInfos().Return(nil).AnyTimes()
			},
			generatedTasks: []persistence.Task{
				&persistence.CloseExecutionTask{
					TaskData: persistence.TaskData{
						VisibilityTimestamp: now,
						Version:             version,
					},
				},
				&persistence.DeleteHistoryEventTask{
					TaskData: persistence.TaskData{
						VisibilityTimestamp: time.Unix(0, closeEvent.GetTimestamp()).Add(retention),
						Version:             version,
					},
				},
			},
		},
		{
			// parent and children all active in current cluster
			setupFn: func(mockMutableState *MockMutableState) {
				mockMutableState.EXPECT().GetExecutionInfo().Return(&persistence.WorkflowExecutionInfo{
					DomainID:         constants.TestDomainID,
					WorkflowID:       constants.TestWorkflowID,
					RunID:            constants.TestRunID,
					ParentDomainID:   constants.TestParentDomainID,
					ParentWorkflowID: "parent workflowID",
					ParentRunID:      "parent runID",
					InitiatedID:      101,
					CloseStatus:      persistence.WorkflowCloseStatusCompleted,
				}).AnyTimes()
				mockMutableState.EXPECT().HasParentExecution().Return(true).AnyTimes()
				mockMutableState.EXPECT().GetPendingChildExecutionInfos().Return(map[int64]*persistence.ChildExecutionInfo{
					102: {DomainID: constants.TestTargetDomainID, ParentClosePolicy: types.ParentClosePolicyTerminate},
					103: {DomainID: constants.TestRemoteTargetDomainID, ParentClosePolicy: types.ParentClosePolicyAbandon},
				}).AnyTimes()
			},
			generatedTasks: []persistence.Task{
				&persistence.CloseExecutionTask{
					TaskData: persistence.TaskData{
						VisibilityTimestamp: now,
						Version:             version,
					},
				},
				&persistence.DeleteHistoryEventTask{
					TaskData: persistence.TaskData{
						VisibilityTimestamp: time.Unix(0, closeEvent.GetTimestamp()).Add(retention),
						Version:             version,
					},
				},
			},
		},
	}
}
