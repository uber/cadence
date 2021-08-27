// Copyright (c) 2021 Uber Technologies, Inc.
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

package task

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/mocks"
	p "github.com/uber/cadence/common/persistence"
	ctask "github.com/uber/cadence/common/task"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/engine"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
	test "github.com/uber/cadence/service/history/testing"
)

type (
	crossClusterSourceTaskExecutorSuite struct {
		suite.Suite
		*require.Assertions

		controller          *gomock.Controller
		mockShard           *shard.TestContext
		mockEngine          *engine.MockEngine
		mockDomainCache     *cache.MockDomainCache
		mockClusterMetadata *cluster.MockMetadata
		mockExecutionMgr    *mocks.ExecutionManager
		mockHistoryV2Mgr    *mocks.HistoryV2Manager
		executionCache      *execution.Cache

		executor Executor
	}
)

func TestCrossClusterSourceTaskExecutorSuite(t *testing.T) {
	s := new(crossClusterSourceTaskExecutorSuite)
	suite.Run(t, s)
}

func (s *crossClusterSourceTaskExecutorSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	config := config.NewForTest()
	s.mockShard = shard.NewTestContext(
		s.controller,
		&p.ShardInfo{
			ShardID: 0,
			RangeID: 1,
		},
		config,
	)
	s.mockShard.SetEventsCache(events.NewCache(
		s.mockShard.GetShardID(),
		s.mockShard.GetHistoryManager(),
		s.mockShard.GetConfig(),
		s.mockShard.GetLogger(),
		s.mockShard.GetMetricsClient(),
	))
	s.mockEngine = engine.NewMockEngine(s.controller)
	s.mockEngine.EXPECT().NotifyNewHistoryEvent(gomock.Any()).AnyTimes()
	s.mockEngine.EXPECT().NotifyNewTransferTasks(gomock.Any(), gomock.Any()).AnyTimes()
	s.mockEngine.EXPECT().NotifyNewTimerTasks(gomock.Any(), gomock.Any()).AnyTimes()
	s.mockEngine.EXPECT().NotifyNewCrossClusterTasks(gomock.Any(), gomock.Any()).AnyTimes()
	s.mockShard.SetEngine(s.mockEngine)

	s.mockClusterMetadata = s.mockShard.Resource.ClusterMetadata
	s.mockDomainCache = s.mockShard.Resource.DomainCache
	s.mockExecutionMgr = s.mockShard.Resource.ExecutionMgr
	s.mockHistoryV2Mgr = s.mockShard.Resource.HistoryMgr

	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestDomainID).Return(constants.TestGlobalDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestDomainID).Return(constants.TestDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(constants.TestDomainName).Return(constants.TestDomainID, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestTargetDomainID).Return(constants.TestGlobalTargetDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestTargetDomainID).Return(constants.TestTargetDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(constants.TestTargetDomainName).Return(constants.TestTargetDomainID, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestRemoteTargetDomainID).Return(constants.TestGlobalRemoteTargetDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestRemoteTargetDomainID).Return(constants.TestRemoteTargetDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(constants.TestRemoteTargetDomainName).Return(constants.TestRemoteTargetDomainID, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestParentDomainID).Return(constants.TestGlobalParentDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestParentDomainID).Return(constants.TestParentDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(constants.TestParentDomainName).Return(constants.TestParentDomainID, nil).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetAllClusterInfo().Return(cluster.TestAllClusterInfo).AnyTimes()
	s.mockClusterMetadata.EXPECT().IsGlobalDomainEnabled().Return(true).AnyTimes()

	s.executionCache = execution.NewCache(s.mockShard)
	s.executor = NewCrossClusterSourceTaskExecutor(
		s.mockShard,
		s.executionCache,
		s.mockShard.GetLogger(),
		config,
	)
}

func (s *crossClusterSourceTaskExecutorSuite) TearDownTest() {
	s.controller.Finish()
	s.mockShard.Finish(s.T())
}

func (s *crossClusterSourceTaskExecutorSuite) TestExecute_UnexpectedTask() {
	transferTask := NewTransferTask(
		s.mockShard,
		&p.TransferTaskInfo{},
		QueueTypeActiveTransfer,
		nil, nil, nil, nil, nil, nil,
	)

	err := s.executor.Execute(transferTask, true)
	s.Equal(errUnexpectedTask, err)
}

func (s *crossClusterSourceTaskExecutorSuite) TestExecute_DomainNotActive() {
	for _, processingState := range []processingState{processingStateResponseReported, processingStateInvalidated} {
		s.testProcessCancelExecution(
			constants.TestRemoteTargetDomainID,
			processingState,
			nil,
			func(
				mutableState execution.MutableState,
				workflowExecution, targetExecution types.WorkflowExecution,
				lastEvent *types.HistoryEvent,
				transferTask Task,
				requestCancelInfo *p.RequestCancelInfo,
			) {
				persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
				s.NoError(err)
				s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&p.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
				s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.MatchedBy(
					func(req *p.UpdateWorkflowExecutionRequest) bool {
						if len(req.UpdateWorkflowMutation.CrossClusterTasks) != 0 || len(req.UpdateWorkflowMutation.TransferTasks) != 1 {
							return false
						}
						transferTask := req.UpdateWorkflowMutation.TransferTasks[0]
						return transferTask.GetType() == p.TransferTaskTypeCancelExecution
					},
				)).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
				s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(mutableState.GetCurrentVersion()).Return(cluster.TestAlternativeClusterName).AnyTimes()
			},
			func(task *crossClusterSourceTask) {
				s.Equal(ctask.TaskStateAcked, task.state)
			},
		)
	}
}

func (s *crossClusterSourceTaskExecutorSuite) TestExecuteCancelExecution_Success() {
	s.testProcessCancelExecution(
		constants.TestDomainID,
		processingStateResponseReported,
		&types.CrossClusterTaskResponse{
			TaskType:                  types.CrossClusterTaskTypeCancelExecution.Ptr(),
			TaskState:                 int16(processingStateInitialized),
			FailedCause:               nil,
			CancelExecutionAttributes: &types.CrossClusterCancelExecutionResponseAttributes{},
		},
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			transferTask Task,
			requestCancelInfo *p.RequestCancelInfo,
		) {
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&p.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.MatchedBy(
				func(req *p.AppendHistoryNodesRequest) bool {
					return req.Events[0].GetEventType() == types.EventTypeExternalWorkflowExecutionCancelRequested
				},
			)).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.MatchedBy(
				func(req *p.UpdateWorkflowExecutionRequest) bool {
					return len(req.UpdateWorkflowMutation.CrossClusterTasks) == 0 &&
						len(req.UpdateWorkflowMutation.TransferTasks) == 1 // one decision task
				},
			)).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
			s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(mutableState.GetCurrentVersion()).Return(cluster.TestCurrentClusterName).AnyTimes()
		},
		func(task *crossClusterSourceTask) {
			s.Equal(ctask.TaskStateAcked, task.state)
		},
	)
}

func (s *crossClusterSourceTaskExecutorSuite) TestExecuteCancelExecution_Failure() {
	s.testProcessCancelExecution(
		constants.TestDomainID,
		processingStateResponseReported,
		&types.CrossClusterTaskResponse{
			TaskType:    types.CrossClusterTaskTypeCancelExecution.Ptr(),
			TaskState:   int16(processingStateInitialized),
			FailedCause: types.CrossClusterTaskFailedCauseWorkflowNotExists.Ptr(),
		},
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			transferTask Task,
			requestCancelInfo *p.RequestCancelInfo,
		) {
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&p.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.MatchedBy(
				func(req *p.AppendHistoryNodesRequest) bool {
					return req.Events[0].GetEventType() == types.EventTypeRequestCancelExternalWorkflowExecutionFailed
				},
			)).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.MatchedBy(
				func(req *p.UpdateWorkflowExecutionRequest) bool {
					return len(req.UpdateWorkflowMutation.CrossClusterTasks) == 0 &&
						len(req.UpdateWorkflowMutation.TransferTasks) == 1 // one decision task
				},
			)).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
			s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(mutableState.GetCurrentVersion()).Return(cluster.TestCurrentClusterName).AnyTimes()
		},
		func(task *crossClusterSourceTask) {
			s.Equal(ctask.TaskStateAcked, task.state)
		},
	)
}

func (s *crossClusterSourceTaskExecutorSuite) TestExecuteCancelExecution_Duplication() {
	s.testProcessCancelExecution(
		constants.TestDomainID,
		processingStateResponseReported,
		&types.CrossClusterTaskResponse{
			TaskType:                  types.CrossClusterTaskTypeCancelExecution.Ptr(),
			TaskState:                 int16(processingStateInitialized),
			FailedCause:               nil,
			CancelExecutionAttributes: &types.CrossClusterCancelExecutionResponseAttributes{},
		},
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			transferTask Task,
			requestCancelInfo *p.RequestCancelInfo,
		) {
			lastEvent = test.AddCancelRequestedEvent(mutableState, lastEvent.GetEventID(), constants.TestTargetDomainID, targetExecution.GetWorkflowID(), targetExecution.GetRunID())
			mutableState.FlushBufferedEvents()

			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&p.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
		},
		func(task *crossClusterSourceTask) {
			s.Equal(ctask.TaskStateAcked, task.state)
		},
	)
}

func (s *crossClusterSourceTaskExecutorSuite) testProcessCancelExecution(
	sourceDomainID string,
	proessingState processingState,
	response *types.CrossClusterTaskResponse,
	setupMockFn func(
		mutableState execution.MutableState,
		workflowExecution, targetExecution types.WorkflowExecution,
		lastEvent *types.HistoryEvent,
		transferTask Task,
		requestCancelInfo *p.RequestCancelInfo,
	),
	taskStateValidationFn func(
		task *crossClusterSourceTask,
	),
) {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, sourceDomainID)
	s.NoError(err)
	targetExecution := types.WorkflowExecution{
		WorkflowID: "some random target workflow ID",
		RunID:      "some random target runID",
	}

	event, rci := test.AddRequestCancelInitiatedEvent(
		mutableState,
		decisionCompletionID,
		uuid.New(),
		constants.TestTargetDomainName,
		targetExecution.GetWorkflowID(),
		targetExecution.GetRunID(),
	)

	crossClusterTask := s.getTestCrossClusterSourceTask(
		&p.CrossClusterTaskInfo{
			Version:          mutableState.GetCurrentVersion(),
			DomainID:         sourceDomainID,
			WorkflowID:       workflowExecution.GetWorkflowID(),
			RunID:            workflowExecution.GetRunID(),
			TargetDomainID:   constants.TestTargetDomainID,
			TargetWorkflowID: targetExecution.GetWorkflowID(),
			TargetRunID:      targetExecution.GetRunID(),
			TaskID:           int64(59),
			TaskList:         mutableState.GetExecutionInfo().TaskList,
			TaskType:         p.CrossClusterTaskTypeCancelExecution,
			ScheduleID:       event.GetEventID(),
		},
		response,
		proessingState,
	)

	setupMockFn(mutableState, workflowExecution, targetExecution, event, crossClusterTask, rci)

	err = s.executor.Execute(crossClusterTask, true)
	s.Nil(err)

	if taskStateValidationFn != nil {
		taskStateValidationFn(crossClusterTask)
	}
}

func (s *crossClusterSourceTaskExecutorSuite) TestExecuteSignalExecution_InitState_Success() {
	s.testProcessSignalExecution(
		constants.TestDomainID,
		processingStateResponseReported,
		&types.CrossClusterTaskResponse{
			TaskType:                  types.CrossClusterTaskTypeSignalExecution.Ptr(),
			TaskState:                 int16(processingStateInitialized),
			FailedCause:               nil,
			SignalExecutionAttributes: &types.CrossClusterSignalExecutionResponseAttributes{},
		},
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			transferTask Task,
			signalInfo *p.SignalInfo,
		) {
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&p.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.MatchedBy(
				func(req *p.AppendHistoryNodesRequest) bool {
					return req.Events[0].GetEventType() == types.EventTypeExternalWorkflowExecutionSignaled
				},
			)).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.MatchedBy(
				func(req *p.UpdateWorkflowExecutionRequest) bool {
					return len(req.UpdateWorkflowMutation.CrossClusterTasks) == 0 &&
						len(req.UpdateWorkflowMutation.TransferTasks) == 1 // one decision task
				},
			)).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
			s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(mutableState.GetCurrentVersion()).Return(cluster.TestCurrentClusterName).AnyTimes()
		},
		func(task *crossClusterSourceTask) {
			s.Equal(ctask.TaskStatePending, task.state)
			s.Equal(processingStateResponseRecorded, task.processingState)
		},
	)
}

func (s *crossClusterSourceTaskExecutorSuite) TestExecuteSignalExecution_InitState_Failure() {
	s.testProcessSignalExecution(
		constants.TestDomainID,
		processingStateResponseReported,
		&types.CrossClusterTaskResponse{
			TaskType:    types.CrossClusterTaskTypeSignalExecution.Ptr(),
			TaskState:   int16(processingStateInitialized),
			FailedCause: types.CrossClusterTaskFailedCauseWorkflowAlreadyCompleted.Ptr(),
		},
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			transferTask Task,
			signalInfo *p.SignalInfo,
		) {
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&p.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.MatchedBy(
				func(req *p.AppendHistoryNodesRequest) bool {
					return req.Events[0].GetEventType() == types.EventTypeSignalExternalWorkflowExecutionFailed
				},
			)).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.MatchedBy(
				func(req *p.UpdateWorkflowExecutionRequest) bool {
					return len(req.UpdateWorkflowMutation.CrossClusterTasks) == 0 &&
						len(req.UpdateWorkflowMutation.TransferTasks) == 1 // one decision task
				},
			)).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
			s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(mutableState.GetCurrentVersion()).Return(cluster.TestCurrentClusterName).AnyTimes()
		},
		func(task *crossClusterSourceTask) {
			s.Equal(ctask.TaskStateAcked, task.state)
		},
	)
}

func (s *crossClusterSourceTaskExecutorSuite) TestExecuteSignalExecution_InitState_Duplication() {
	s.testProcessSignalExecution(
		constants.TestDomainID,
		processingStateResponseReported,
		&types.CrossClusterTaskResponse{
			TaskType:    types.CrossClusterTaskTypeSignalExecution.Ptr(),
			TaskState:   int16(processingStateInitialized),
			FailedCause: types.CrossClusterTaskFailedCauseWorkflowAlreadyCompleted.Ptr(),
		},
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			transferTask Task,
			signalInfo *p.SignalInfo,
		) {
			lastEvent = test.AddSignaledEvent(mutableState, lastEvent.GetEventID(), constants.TestTargetDomainID, targetExecution.GetWorkflowID(), targetExecution.GetRunID(), signalInfo.Control)
			mutableState.FlushBufferedEvents()

			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&p.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
		},
		func(task *crossClusterSourceTask) {
			s.Equal(ctask.TaskStateAcked, task.state)
		},
	)
}

func (s *crossClusterSourceTaskExecutorSuite) TestExecuteSignalExecution_RecordedState() {
	s.testProcessSignalExecution(
		constants.TestDomainID,
		processingStateResponseReported,
		&types.CrossClusterTaskResponse{
			TaskType:    types.CrossClusterTaskTypeSignalExecution.Ptr(),
			TaskState:   int16(processingStateResponseRecorded),
			FailedCause: types.CrossClusterTaskFailedCauseWorkflowAlreadyCompleted.Ptr(),
		},
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			transferTask Task,
			signalInfo *p.SignalInfo,
		) {
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&p.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
		},
		func(task *crossClusterSourceTask) {
			s.Equal(ctask.TaskStateAcked, task.state)
		},
	)
}

func (s *crossClusterSourceTaskExecutorSuite) testProcessSignalExecution(
	sourceDomainID string,
	proessingState processingState,
	response *types.CrossClusterTaskResponse,
	setupMockFn func(
		mutableState execution.MutableState,
		workflowExecution, targetExecution types.WorkflowExecution,
		lastEvent *types.HistoryEvent,
		crossClusterTask Task,
		signalInfo *p.SignalInfo,
	),
	taskStateValidationFn func(
		task *crossClusterSourceTask,
	),
) {
	workflowExecution := types.WorkflowExecution{
		WorkflowID: constants.TestWorkflowID,
		RunID:      constants.TestRunID,
	}
	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, sourceDomainID)
	s.NoError(err)
	targetExecution := types.WorkflowExecution{
		WorkflowID: "some random target workflow ID",
		RunID:      "some random target runID",
	}

	event, signalInfo := test.AddRequestSignalInitiatedEvent(
		mutableState,
		decisionCompletionID,
		uuid.New(),
		constants.TestTargetDomainName,
		targetExecution.GetWorkflowID(),
		targetExecution.GetRunID(),
		"some random signal name",
		[]byte("some random signal input"),
		[]byte("some random signal control"),
	)

	crossClusterTask := s.getTestCrossClusterSourceTask(
		&p.CrossClusterTaskInfo{
			Version:          mutableState.GetCurrentVersion(),
			DomainID:         sourceDomainID,
			WorkflowID:       workflowExecution.GetWorkflowID(),
			RunID:            workflowExecution.GetRunID(),
			TargetDomainID:   constants.TestTargetDomainID,
			TargetWorkflowID: targetExecution.GetWorkflowID(),
			TargetRunID:      targetExecution.GetRunID(),
			TaskID:           int64(59),
			TaskList:         mutableState.GetExecutionInfo().TaskList,
			TaskType:         p.CrossClusterTaskTypeSignalExecution,
			ScheduleID:       event.GetEventID(),
		},
		response,
		proessingState,
	)

	setupMockFn(mutableState, workflowExecution, targetExecution, event, crossClusterTask, signalInfo)

	err = s.executor.Execute(crossClusterTask, true)
	s.Nil(err)

	if taskStateValidationFn != nil {
		taskStateValidationFn(crossClusterTask)
	}
}

func (s *crossClusterSourceTaskExecutorSuite) TestExecuteStartChildExecution_InitState_Success() {
	s.testProcessStartChildExecution(
		constants.TestDomainID,
		processingStateResponseReported,
		&types.CrossClusterTaskResponse{
			TaskType:    types.CrossClusterTaskTypeStartChildExecution.Ptr(),
			TaskState:   int16(processingStateInitialized),
			FailedCause: nil,
			StartChildExecutionAttributes: &types.CrossClusterStartChildExecutionResponseAttributes{
				RunID: "some random child runID",
			},
		},
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			crossClusterTask Task,
			childInfo *p.ChildExecutionInfo,
		) {
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&p.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.MatchedBy(
				func(req *p.AppendHistoryNodesRequest) bool {
					return req.Events[0].GetEventType() == types.EventTypeChildWorkflowExecutionStarted
				},
			)).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.MatchedBy(
				func(req *p.UpdateWorkflowExecutionRequest) bool {
					return len(req.UpdateWorkflowMutation.CrossClusterTasks) == 0 &&
						len(req.UpdateWorkflowMutation.TransferTasks) == 1 // one decision task
				},
			)).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
			s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(mutableState.GetCurrentVersion()).Return(cluster.TestCurrentClusterName).AnyTimes()
		},
		func(task *crossClusterSourceTask) {
			s.Equal(ctask.TaskStatePending, task.state)
			s.Equal(processingStateResponseRecorded, task.processingState)
		},
	)
}

func (s *crossClusterSourceTaskExecutorSuite) TestExecuteStartChildExecution_InitState_Failure() {
	s.testProcessStartChildExecution(
		constants.TestDomainID,
		processingStateResponseReported,
		&types.CrossClusterTaskResponse{
			TaskType:    types.CrossClusterTaskTypeStartChildExecution.Ptr(),
			TaskState:   int16(processingStateInitialized),
			FailedCause: types.CrossClusterTaskFailedCauseWorkflowAlreadyRunning.Ptr(),
		},
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			crossClusterTask Task,
			childInfo *p.ChildExecutionInfo,
		) {
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&p.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.MatchedBy(
				func(req *p.AppendHistoryNodesRequest) bool {
					return req.Events[0].GetEventType() == types.EventTypeStartChildWorkflowExecutionFailed
				},
			)).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.MatchedBy(
				func(req *p.UpdateWorkflowExecutionRequest) bool {
					return len(req.UpdateWorkflowMutation.CrossClusterTasks) == 0 &&
						len(req.UpdateWorkflowMutation.TransferTasks) == 1 // one decision task
				},
			)).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
			s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(mutableState.GetCurrentVersion()).Return(cluster.TestCurrentClusterName).AnyTimes()
		},
		func(task *crossClusterSourceTask) {
			s.Equal(ctask.TaskStateAcked, task.state)
		},
	)
}

func (s *crossClusterSourceTaskExecutorSuite) TestExecuteStartChildExecution_InitState_Duplication() {
	s.testProcessStartChildExecution(
		constants.TestDomainID,
		processingStateResponseReported,
		&types.CrossClusterTaskResponse{
			TaskType:    types.CrossClusterTaskTypeStartChildExecution.Ptr(),
			TaskState:   int16(processingStateInitialized),
			FailedCause: types.CrossClusterTaskFailedCauseWorkflowAlreadyRunning.Ptr(),
		},
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			crossClusterTask Task,
			childInfo *p.ChildExecutionInfo,
		) {
			lastEvent = test.AddChildWorkflowExecutionStartedEvent(mutableState, lastEvent.GetEventID(), constants.TestTargetDomainID, targetExecution.GetWorkflowID(), targetExecution.GetRunID(), childInfo.WorkflowTypeName)
			mutableState.FlushBufferedEvents()

			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&p.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
		},
		func(task *crossClusterSourceTask) {
			s.Equal(ctask.TaskStatePending, task.state)
			s.Equal(processingStateResponseRecorded, task.processingState)
		},
	)
}

func (s *crossClusterSourceTaskExecutorSuite) TestExecuteStartChildExecution_RecordedState() {
	s.testProcessStartChildExecution(
		constants.TestDomainID,
		processingStateResponseReported,
		&types.CrossClusterTaskResponse{
			TaskType:    types.CrossClusterTaskTypeStartChildExecution.Ptr(),
			TaskState:   int16(processingStateResponseRecorded),
			FailedCause: nil,
			StartChildExecutionAttributes: &types.CrossClusterStartChildExecutionResponseAttributes{
				RunID: "some random child runID",
			},
		},
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			crossClusterTask Task,
			childInfo *p.ChildExecutionInfo,
		) {
			lastEvent = test.AddChildWorkflowExecutionStartedEvent(mutableState, lastEvent.GetEventID(), constants.TestTargetDomainID, targetExecution.GetWorkflowID(), targetExecution.GetRunID(), childInfo.WorkflowTypeName)
			mutableState.FlushBufferedEvents()

			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&p.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
		},
		func(task *crossClusterSourceTask) {
			s.Equal(ctask.TaskStateAcked, task.state)
		},
	)
}

func (s *crossClusterSourceTaskExecutorSuite) testProcessStartChildExecution(
	sourceDomainID string,
	proessingState processingState,
	response *types.CrossClusterTaskResponse,
	setupMockFn func(
		mutableState execution.MutableState,
		workflowExecution, targetExecution types.WorkflowExecution,
		lastEvent *types.HistoryEvent,
		crossClusterTask Task,
		childInfo *p.ChildExecutionInfo,
	),
	taskStateValidationFn func(
		task *crossClusterSourceTask,
	),
) {
	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, sourceDomainID)
	s.NoError(err)
	targetExecution := types.WorkflowExecution{
		WorkflowID: "some random target workflow ID",
	}

	event, childInfo := test.AddStartChildWorkflowExecutionInitiatedEvent(
		mutableState,
		decisionCompletionID,
		uuid.New(),
		constants.TestTargetDomainName,
		targetExecution.WorkflowID,
		"some random child workflow type",
		"some random child task list",
		nil,
		1,
		1,
		&types.RetryPolicy{
			ExpirationIntervalInSeconds: 100,
			MaximumAttempts:             3,
			InitialIntervalInSeconds:    1,
			MaximumIntervalInSeconds:    2,
			BackoffCoefficient:          1,
		},
	)

	crossClusterTask := s.getTestCrossClusterSourceTask(
		&p.CrossClusterTaskInfo{
			Version:          mutableState.GetCurrentVersion(),
			DomainID:         sourceDomainID,
			WorkflowID:       workflowExecution.GetWorkflowID(),
			RunID:            workflowExecution.GetRunID(),
			TargetDomainID:   constants.TestTargetDomainID,
			TargetWorkflowID: targetExecution.GetWorkflowID(),
			TaskID:           int64(59),
			TaskList:         mutableState.GetExecutionInfo().TaskList,
			TaskType:         p.CrossClusterTaskTypeStartChildExecution,
			ScheduleID:       event.GetEventID(),
		},
		response,
		proessingState,
	)

	setupMockFn(mutableState, workflowExecution, targetExecution, event, crossClusterTask, childInfo)

	err = s.executor.Execute(crossClusterTask, true)
	s.Nil(err)

	if taskStateValidationFn != nil {
		taskStateValidationFn(crossClusterTask)
	}
}

func (s *crossClusterSourceTaskExecutorSuite) getTestCrossClusterSourceTask(
	taskInfo *p.CrossClusterTaskInfo,
	response *types.CrossClusterTaskResponse,
	processingState processingState,
) *crossClusterSourceTask {
	task := NewCrossClusterSourceTask(
		s.mockShard,
		cluster.TestAlternativeClusterName,
		s.executionCache,
		taskInfo,
		s.executor,
		nil,
		s.mockShard.GetLogger(),
		nil,
		dynamicconfig.GetIntPropertyFn(1),
	).(*crossClusterSourceTask)
	task.response = response
	task.processingState = processingState
	return task
}
