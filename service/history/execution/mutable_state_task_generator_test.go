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
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log/loggerimpl"
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
	testCases := []struct {
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
					VisibilityTimestamp: now,
					Version:             version,
				},
				&persistence.DeleteHistoryEventTask{
					VisibilityTimestamp: time.Unix(0, closeEvent.GetTimestamp()).Add(retention),
					Version:             version,
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
					VisibilityTimestamp: now,
					Version:             version,
				},
				&persistence.DeleteHistoryEventTask{
					VisibilityTimestamp: time.Unix(0, closeEvent.GetTimestamp()).Add(retention),
					Version:             version,
				},
			},
		},
		{
			// parent active in cluster cluster,
			// one child active in cluster, one child active in remote
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
					103: {DomainID: constants.TestRemoteTargetDomainID, ParentClosePolicy: types.ParentClosePolicyRequestCancel},
				}).AnyTimes()
			},
			generatedTasks: []persistence.Task{
				&persistence.RecordChildExecutionCompletedTask{
					VisibilityTimestamp: now,
					TargetDomainID:      constants.TestParentDomainID,
					TargetWorkflowID:    "parent workflowID",
					TargetRunID:         "parent runID",
					Version:             version,
				},
				&persistence.ApplyParentClosePolicyTask{
					VisibilityTimestamp: now,
					TargetDomainIDs:     map[string]struct{}{constants.TestTargetDomainID: {}},
					Version:             version,
				},
				&persistence.RecordWorkflowClosedTask{
					VisibilityTimestamp: now,
					Version:             version,
				},
				&persistence.CrossClusterApplyParentClosePolicyTask{
					TargetCluster: cluster.TestAlternativeClusterName,
					ApplyParentClosePolicyTask: persistence.ApplyParentClosePolicyTask{
						VisibilityTimestamp: now,
						TargetDomainIDs:     map[string]struct{}{constants.TestRemoteTargetDomainID: {}},
						Version:             version,
					},
				},
				&persistence.DeleteHistoryEventTask{
					VisibilityTimestamp: time.Unix(0, closeEvent.GetTimestamp()).Add(retention),
					Version:             version,
				},
			},
		},
		{
			// parent active in remote cluster, all children active in current cluster
			setupFn: func(mockMutableState *MockMutableState) {
				mockMutableState.EXPECT().GetExecutionInfo().Return(&persistence.WorkflowExecutionInfo{
					DomainID:         constants.TestDomainID,
					WorkflowID:       constants.TestWorkflowID,
					RunID:            constants.TestRunID,
					ParentDomainID:   constants.TestRemoteTargetDomainID,
					ParentWorkflowID: "parent workflowID",
					ParentRunID:      "parent runID",
					InitiatedID:      101,
					CloseStatus:      persistence.WorkflowCloseStatusCompleted,
				}).AnyTimes()
				mockMutableState.EXPECT().HasParentExecution().Return(true).AnyTimes()
				mockMutableState.EXPECT().GetPendingChildExecutionInfos().Return(map[int64]*persistence.ChildExecutionInfo{
					102: {DomainID: constants.TestTargetDomainID, ParentClosePolicy: types.ParentClosePolicyTerminate},
					103: {DomainID: constants.TestChildDomainID, ParentClosePolicy: types.ParentClosePolicyRequestCancel},
				}).AnyTimes()
			},
			generatedTasks: []persistence.Task{
				&persistence.ApplyParentClosePolicyTask{
					VisibilityTimestamp: now,
					TargetDomainIDs:     map[string]struct{}{constants.TestTargetDomainID: {}, constants.TestChildDomainID: {}},
					Version:             version,
				},
				&persistence.RecordWorkflowClosedTask{
					VisibilityTimestamp: now,
					Version:             version,
				},
				&persistence.CrossClusterRecordChildExecutionCompletedTask{
					TargetCluster: cluster.TestAlternativeClusterName,
					RecordChildExecutionCompletedTask: persistence.RecordChildExecutionCompletedTask{
						VisibilityTimestamp: now,
						TargetDomainID:      constants.TestRemoteTargetDomainID,
						TargetWorkflowID:    "parent workflowID",
						TargetRunID:         "parent runID",
						Version:             version,
					},
				},
				&persistence.DeleteHistoryEventTask{
					VisibilityTimestamp: time.Unix(0, closeEvent.GetTimestamp()).Add(retention),
					Version:             version,
				},
			},
		},
		{
			// parent active in remote cluster
			// two children active in current cluster, one child active in remote cluster
			setupFn: func(mockMutableState *MockMutableState) {
				mockMutableState.EXPECT().GetExecutionInfo().Return(&persistence.WorkflowExecutionInfo{
					DomainID:         constants.TestDomainID,
					WorkflowID:       constants.TestWorkflowID,
					RunID:            constants.TestRunID,
					ParentDomainID:   constants.TestRemoteTargetDomainID,
					ParentWorkflowID: "parent workflowID",
					ParentRunID:      "parent runID",
					InitiatedID:      101,
					CloseStatus:      persistence.WorkflowCloseStatusCompleted,
				}).AnyTimes()
				mockMutableState.EXPECT().HasParentExecution().Return(true).AnyTimes()
				mockMutableState.EXPECT().GetPendingChildExecutionInfos().Return(map[int64]*persistence.ChildExecutionInfo{
					102: {DomainID: constants.TestTargetDomainID, ParentClosePolicy: types.ParentClosePolicyTerminate},
					103: {DomainID: constants.TestChildDomainID, ParentClosePolicy: types.ParentClosePolicyRequestCancel},
					104: {DomainID: constants.TestRemoteTargetDomainID, ParentClosePolicy: types.ParentClosePolicyRequestCancel},
				}).AnyTimes()
			},
			generatedTasks: []persistence.Task{
				&persistence.ApplyParentClosePolicyTask{
					VisibilityTimestamp: now,
					TargetDomainIDs:     map[string]struct{}{constants.TestTargetDomainID: {}, constants.TestChildDomainID: {}},
					Version:             version,
				},
				&persistence.RecordWorkflowClosedTask{
					VisibilityTimestamp: now,
					Version:             version,
				},
				&persistence.CrossClusterRecordChildExecutionCompletedTask{
					TargetCluster: cluster.TestAlternativeClusterName,
					RecordChildExecutionCompletedTask: persistence.RecordChildExecutionCompletedTask{
						VisibilityTimestamp: now,
						TargetDomainID:      constants.TestRemoteTargetDomainID,
						TargetWorkflowID:    "parent workflowID",
						TargetRunID:         "parent runID",
						Version:             version,
					},
				},
				&persistence.CrossClusterApplyParentClosePolicyTask{
					TargetCluster: cluster.TestAlternativeClusterName,
					ApplyParentClosePolicyTask: persistence.ApplyParentClosePolicyTask{
						VisibilityTimestamp: now,
						TargetDomainIDs:     map[string]struct{}{constants.TestRemoteTargetDomainID: {}},
						Version:             version,
					},
				},
				&persistence.DeleteHistoryEventTask{
					VisibilityTimestamp: time.Unix(0, closeEvent.GetTimestamp()).Add(retention),
					Version:             version,
				},
			},
		},
	}

	for _, tc := range testCases {
		// create new mockMutableState so can we can setup separete mock for each test case
		mockMutableState := NewMockMutableState(s.controller)
		taskGenerator := NewMutableStateTaskGenerator(
			constants.TestClusterMetadata,
			s.mockDomainCache,
			loggerimpl.NewLoggerForTest(s.Suite),
			mockMutableState,
		)

		var transferTasks []persistence.Task
		var crossClusterTasks []persistence.Task
		var timerTasks []persistence.Task
		mockMutableState.EXPECT().GetDomainEntry().Return(domainEntry).AnyTimes()
		mockMutableState.EXPECT().AddTransferTasks(gomock.Any()).Do(func(tasks ...persistence.Task) {
			transferTasks = tasks
		}).MaxTimes(1)
		mockMutableState.EXPECT().AddCrossClusterTasks(gomock.Any()).Do(func(tasks ...persistence.Task) {
			crossClusterTasks = tasks
		}).MaxTimes(1)
		mockMutableState.EXPECT().AddTimerTasks(gomock.Any()).Do(func(tasks ...persistence.Task) {
			timerTasks = tasks
		}).MaxTimes(1)

		tc.setupFn(mockMutableState)
		err := taskGenerator.GenerateWorkflowCloseTasks(closeEvent)
		s.NoError(err)

		actualGeneratedTasks := append(transferTasks, crossClusterTasks...)
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

func (s *mutableStateTaskGeneratorSuite) TestGenerateFromTransferTask() {
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

		err := s.taskGenerator.GenerateFromTransferTask(tc.transferTask, targetCluster)
		if tc.expectError {
			s.Error(err)
		} else {
			s.Equal(tc.expectedCrossClusterTask, actualCrossClusterTask)
		}
	}
}

func (s *mutableStateTaskGeneratorSuite) TestGenerateCrossClusterRecordChildCompletedTask() {
	targetCluster := cluster.TestAlternativeClusterName
	transferTask := &persistence.TransferTaskInfo{
		TaskType:            persistence.TransferTaskTypeCloseExecution,
		VisibilityTimestamp: time.Now(),
		Version:             int64(101),
	}
	parentInfo := &types.ParentExecutionInfo{
		DomainUUID: constants.TestParentDomainID,
		Domain:     constants.TestParentDomainName,
		Execution: &types.WorkflowExecution{
			WorkflowID: constants.TestWorkflowID,
			RunID:      constants.TestRunID,
		},
		InitiatedID: 123,
	}
	expectedTask := &persistence.CrossClusterRecordChildExecutionCompletedTask{
		TargetCluster: targetCluster,
		RecordChildExecutionCompletedTask: persistence.RecordChildExecutionCompletedTask{
			VisibilityTimestamp: transferTask.GetVisibilityTimestamp(),
			TargetDomainID:      constants.TestParentDomainID,
			TargetWorkflowID:    constants.TestWorkflowID,
			TargetRunID:         constants.TestRunID,
			Version:             101,
		},
	}

	var actualTask persistence.Task
	s.mockMutableState.EXPECT().AddCrossClusterTasks(gomock.Any()).Do(
		func(crossClusterTasks ...persistence.Task) {
			s.Len(crossClusterTasks, 1)
			actualTask = crossClusterTasks[0]
		},
	).Times(1)

	err := s.taskGenerator.GenerateCrossClusterRecordChildCompletedTask(
		transferTask,
		targetCluster,
		parentInfo,
	)
	s.NoError(err)
	s.Equal(expectedTask, actualTask)
}

func (s *mutableStateTaskGeneratorSuite) TestGenerateCrossClusterApplyParentClosePolicyTask() {
	targetCluster := cluster.TestAlternativeClusterName
	transferTask := &persistence.TransferTaskInfo{
		TaskType:            persistence.TransferTaskTypeCloseExecution,
		VisibilityTimestamp: time.Now(),
		Version:             int64(101),
	}
	childDomainIDs := map[string]struct{}{constants.TestRemoteTargetDomainID: {}}
	expectedTask := &persistence.CrossClusterApplyParentClosePolicyTask{
		TargetCluster: targetCluster,
		ApplyParentClosePolicyTask: persistence.ApplyParentClosePolicyTask{
			VisibilityTimestamp: transferTask.GetVisibilityTimestamp(),
			TargetDomainIDs:     map[string]struct{}{constants.TestRemoteTargetDomainID: {}},
			Version:             101,
		},
	}

	var actualTask persistence.Task
	s.mockMutableState.EXPECT().AddCrossClusterTasks(gomock.Any()).Do(
		func(crossClusterTasks ...persistence.Task) {
			s.Len(crossClusterTasks, 1)
			actualTask = crossClusterTasks[0]
		},
	).Times(1)

	err := s.taskGenerator.GenerateCrossClusterApplyParentClosePolicyTask(
		transferTask,
		targetCluster,
		childDomainIDs,
	)
	s.NoError(err)
	s.Equal(expectedTask, actualTask)
}

func (s *mutableStateTaskGeneratorSuite) TestGenerateFromCrossClusterTask() {
	testCases := []struct {
		sourceActive     bool
		crossClusterTask *persistence.CrossClusterTaskInfo
		generatedTasks   []persistence.Task
	}{
		{
			sourceActive: true,
			crossClusterTask: &persistence.CrossClusterTaskInfo{
				TaskType:         persistence.CrossClusterTaskTypeStartChildExecution,
				TargetDomainID:   constants.TestTargetDomainID,
				TargetWorkflowID: constants.TestWorkflowID,
				ScheduleID:       int64(123),
			},
			generatedTasks: []persistence.Task{
				&persistence.StartChildExecutionTask{
					TargetDomainID:   constants.TestTargetDomainID,
					TargetWorkflowID: constants.TestWorkflowID,
					InitiatedID:      int64(123),
				},
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
			generatedTasks: []persistence.Task{
				&persistence.CrossClusterSignalExecutionTask{
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
			generatedTasks: []persistence.Task{
				&persistence.CancelExecutionTask{
					TargetDomainID:          constants.TestTargetDomainID,
					TargetWorkflowID:        constants.TestWorkflowID,
					TargetRunID:             constants.TestRunID,
					TargetChildWorkflowOnly: false,
					InitiatedID:             int64(123),
				},
			},
		},
		{
			sourceActive: true,
			crossClusterTask: &persistence.CrossClusterTaskInfo{
				TaskType:         persistence.CrossClusterTaskTypeRecordChildExeuctionCompleted,
				TargetDomainID:   constants.TestRemoteTargetDomainID,
				TargetWorkflowID: constants.TestWorkflowID,
				TargetRunID:      constants.TestRunID,
				ScheduleID:       int64(123),
			},
			generatedTasks: []persistence.Task{
				&persistence.CrossClusterRecordChildExecutionCompletedTask{
					TargetCluster: cluster.TestAlternativeClusterName,
					RecordChildExecutionCompletedTask: persistence.RecordChildExecutionCompletedTask{
						TargetDomainID:   constants.TestRemoteTargetDomainID,
						TargetWorkflowID: constants.TestWorkflowID,
						TargetRunID:      constants.TestRunID,
					},
				},
			},
		},
		{
			sourceActive: false,
			crossClusterTask: &persistence.CrossClusterTaskInfo{
				TaskType:         persistence.CrossClusterTaskTypeRecordChildExeuctionCompleted,
				TargetDomainID:   constants.TestRemoteTargetDomainID,
				TargetWorkflowID: constants.TestWorkflowID,
				TargetRunID:      constants.TestRunID,
				ScheduleID:       int64(123),
			},
			generatedTasks: []persistence.Task{
				&persistence.RecordChildExecutionCompletedTask{
					TargetDomainID:   constants.TestRemoteTargetDomainID,
					TargetWorkflowID: constants.TestWorkflowID,
					TargetRunID:      constants.TestRunID,
				},
			},
		},
		{
			sourceActive: true,
			crossClusterTask: &persistence.CrossClusterTaskInfo{
				TaskType:         persistence.CrossClusterTaskTypeRecordChildExeuctionCompleted,
				TargetDomainID:   constants.TestTargetDomainID,
				TargetWorkflowID: constants.TestWorkflowID,
				TargetRunID:      constants.TestRunID,
				ScheduleID:       int64(123),
			},
			generatedTasks: []persistence.Task{
				&persistence.RecordChildExecutionCompletedTask{
					TargetDomainID:   constants.TestTargetDomainID,
					TargetWorkflowID: constants.TestWorkflowID,
					TargetRunID:      constants.TestRunID,
				},
			},
		},
		{
			sourceActive: true,
			crossClusterTask: &persistence.CrossClusterTaskInfo{
				TaskType:        persistence.CrossClusterTaskTypeApplyParentClosePolicy,
				TargetDomainIDs: map[string]struct{}{constants.TestRemoteTargetDomainID: {}},
				ScheduleID:      int64(123),
			},
			generatedTasks: []persistence.Task{
				&persistence.CrossClusterApplyParentClosePolicyTask{
					TargetCluster: cluster.TestAlternativeClusterName,
					ApplyParentClosePolicyTask: persistence.ApplyParentClosePolicyTask{
						TargetDomainIDs: map[string]struct{}{constants.TestRemoteTargetDomainID: {}},
					},
				},
			},
		},
		{
			sourceActive: false,
			crossClusterTask: &persistence.CrossClusterTaskInfo{
				TaskType:        persistence.CrossClusterTaskTypeApplyParentClosePolicy,
				TargetDomainIDs: map[string]struct{}{constants.TestRemoteTargetDomainID: {}},
				ScheduleID:      int64(123),
			},
			generatedTasks: []persistence.Task{
				&persistence.ApplyParentClosePolicyTask{
					TargetDomainIDs: map[string]struct{}{
						constants.TestRemoteTargetDomainID: {},
					},
				},
			},
		},
		{
			sourceActive: true,
			crossClusterTask: &persistence.CrossClusterTaskInfo{
				TaskType:        persistence.CrossClusterTaskTypeApplyParentClosePolicy,
				TargetDomainIDs: map[string]struct{}{constants.TestTargetDomainID: {}},
				ScheduleID:      int64(123),
			},
			generatedTasks: []persistence.Task{
				&persistence.ApplyParentClosePolicyTask{
					TargetDomainIDs: map[string]struct{}{
						constants.TestTargetDomainID: {},
					},
				},
			},
		},
		{
			sourceActive: true,
			crossClusterTask: &persistence.CrossClusterTaskInfo{
				TaskType: persistence.CrossClusterTaskTypeApplyParentClosePolicy,
				TargetDomainIDs: map[string]struct{}{
					constants.TestTargetDomainID:       {},
					constants.TestRemoteTargetDomainID: {},
				},
				ScheduleID: int64(123),
			},
			generatedTasks: []persistence.Task{
				&persistence.ApplyParentClosePolicyTask{
					TargetDomainIDs: map[string]struct{}{
						constants.TestTargetDomainID: {},
					},
				},
				&persistence.CrossClusterApplyParentClosePolicyTask{
					TargetCluster: cluster.TestAlternativeClusterName,
					ApplyParentClosePolicyTask: persistence.ApplyParentClosePolicyTask{
						TargetDomainIDs: map[string]struct{}{
							constants.TestRemoteTargetDomainID: {},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		// create new mockMutableState so can we can setup separete mock for each test case
		mockMutableState := NewMockMutableState(s.controller)
		taskGenerator := NewMutableStateTaskGenerator(
			constants.TestClusterMetadata,
			s.mockDomainCache,
			loggerimpl.NewLoggerForTest(s.Suite),
			mockMutableState,
		)

		if tc.sourceActive {
			tc.crossClusterTask.DomainID = constants.TestDomainID
			mockMutableState.EXPECT().GetDomainEntry().Return(constants.TestGlobalDomainEntry).Times(1)
		} else {
			tc.crossClusterTask.DomainID = constants.TestRemoteTargetDomainID
			mockMutableState.EXPECT().GetDomainEntry().Return(constants.TestGlobalRemoteTargetDomainEntry).Times(1)
		}

		var transferTasks []persistence.Task
		var crossClusterTasks []persistence.Task
		mockMutableState.EXPECT().AddTransferTasks(gomock.Any()).Do(func(tasks ...persistence.Task) {
			transferTasks = tasks
		}).MaxTimes(1)
		mockMutableState.EXPECT().AddCrossClusterTasks(gomock.Any()).Do(func(tasks ...persistence.Task) {
			crossClusterTasks = tasks
		}).MaxTimes(1)

		err := taskGenerator.GenerateFromCrossClusterTask(tc.crossClusterTask)
		s.NoError(err)

		actualGeneratedTasks := append(transferTasks, crossClusterTasks...)
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
