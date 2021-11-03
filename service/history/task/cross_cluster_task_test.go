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
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
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
	crossClusterTaskSuite struct {
		suite.Suite
		*require.Assertions

		controller          *gomock.Controller
		mockShard           *shard.TestContext
		mockEngine          *engine.MockEngine
		mockDomainCache     *cache.MockDomainCache
		mockClusterMetadata *cluster.MockMetadata
		mockExecutionMgr    *mocks.ExecutionManager
		mockExecutor        *MockExecutor
		mockProcessor       *MockProcessor
		mockRedispatcher    *MockRedispatcher
		executionCache      *execution.Cache
	}
)

func TestCrossClusterTaskSuite(t *testing.T) {
	s := new(crossClusterTaskSuite)
	suite.Run(t, s)
}

func (s *crossClusterTaskSuite) SetupTest() {
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

	s.mockClusterMetadata = s.mockShard.Resource.ClusterMetadata
	s.mockDomainCache = s.mockShard.Resource.DomainCache
	s.mockExecutionMgr = s.mockShard.Resource.ExecutionMgr

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

	s.mockExecutor = NewMockExecutor(s.controller)
	s.mockProcessor = NewMockProcessor(s.controller)
	s.mockRedispatcher = NewMockRedispatcher(s.controller)
	s.executionCache = execution.NewCache(s.mockShard)
}

func (s *crossClusterTaskSuite) TestSourceTask_Execute() {
	sourceTask := s.newTestSourceTask(
		cluster.TestAlternativeClusterName,
		&p.CrossClusterTaskInfo{
			DomainID: constants.TestDomainID,
		},
	)

	s.mockExecutor.EXPECT().Execute(sourceTask, true).Return(nil).Times(1)
	s.NoError(sourceTask.Execute())

	s.mockExecutor.EXPECT().Execute(sourceTask, true).Return(errors.New("some random error")).Times(1)
	s.Error(sourceTask.Execute())
}

func (s *crossClusterTaskSuite) TestSourceTask_Ack() {
	sourceTask := s.newTestSourceTask(
		cluster.TestAlternativeClusterName,
		&p.CrossClusterTaskInfo{
			DomainID: constants.TestDomainID,
		},
	)

	readyForPollFnCalled := false
	sourceTask.readyForPollFn = func(t CrossClusterTask) {
		readyForPollFnCalled = true
		s.Equal(sourceTask, t)
	}

	// case 1: task execution finished, not available for polling
	sourceTask.state = ctask.TaskStateAcked
	sourceTask.Ack()
	// make sure task state is not changed by the Ack function
	s.Equal(ctask.TaskStateAcked, sourceTask.state)
	s.False(readyForPollFnCalled)

	// case 2: task execution not finished, still available for polling
	sourceTask.state = ctask.TaskStatePending
	sourceTask.processingState = processingStateResponseRecorded
	sourceTask.Ack()
	// make sure task state is not changed by the Ack function
	s.Equal(ctask.TaskStatePending, sourceTask.state)
	s.Equal(processingStateResponseRecorded, sourceTask.processingState)
	s.True(readyForPollFnCalled)
}

func (s *crossClusterTaskSuite) TestSourceTask_Nack() {
	sourceTask := s.newTestSourceTask(
		cluster.TestAlternativeClusterName,
		&p.CrossClusterTaskInfo{
			DomainID: constants.TestDomainID,
		},
	)
	sourceTask.state = ctask.TaskStatePending
	sourceTask.processingState = processingStateResponseRecorded
	sourceTask.attempt = 0

	s.mockProcessor.EXPECT().TrySubmit(sourceTask).Return(true, nil).Times(1)
	sourceTask.Nack()
	// make sure task state is not changed by the Nack function
	s.Equal(ctask.TaskStatePending, sourceTask.state)
	s.Equal(processingStateResponseRecorded, sourceTask.processingState)

	s.mockProcessor.EXPECT().TrySubmit(sourceTask).Return(false, nil).Times(1)
	s.mockRedispatcher.EXPECT().AddTask(sourceTask).Times(1)
	sourceTask.Nack()
	s.Equal(ctask.TaskStatePending, sourceTask.state)
	s.Equal(processingStateResponseRecorded, sourceTask.processingState)

	sourceTask.attempt = activeTaskResubmitMaxAttempts + 1
	s.mockRedispatcher.EXPECT().AddTask(sourceTask).Times(1)
	sourceTask.Nack()
	s.Equal(ctask.TaskStatePending, sourceTask.state)
	s.Equal(processingStateResponseRecorded, sourceTask.processingState)
}

func (s *crossClusterTaskSuite) TestSourceTask_HandleError() {
	unknownErr := errors.New("some unknown error")
	testCases := []struct {
		err          error
		expectRetErr error
	}{
		{
			err:          unknownErr,
			expectRetErr: unknownErr,
		},
		{
			err:          &types.EntityNotExistsError{},
			expectRetErr: nil,
		},
		{
			err:          &types.WorkflowExecutionAlreadyCompletedError{},
			expectRetErr: nil,
		},
		{
			err:          errWorkflowBusy,
			expectRetErr: errWorkflowBusy,
		},
		{
			err:          ErrTaskPendingActive,
			expectRetErr: ErrTaskPendingActive,
		},
		{
			err:          &types.DomainNotActiveError{},
			expectRetErr: &types.DomainNotActiveError{},
		},
		{
			err:          context.DeadlineExceeded,
			expectRetErr: context.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		sourceTask := s.newTestSourceTask(
			cluster.TestAlternativeClusterName,
			&p.CrossClusterTaskInfo{
				DomainID: constants.TestDomainID,
			},
		)

		retErr := sourceTask.HandleErr(tc.err)
		s.Equal(tc.expectRetErr, retErr)
		if tc.expectRetErr != nil {
			s.Equal(1, sourceTask.GetAttempt())
		}
	}
}

func (s *crossClusterTaskSuite) TestSourceTask_IsReadyForPoll() {
	testCases := []struct {
		state           ctask.State
		processingState processingState
		readyForPoll    bool
	}{
		{
			state:           ctask.TaskStatePending,
			processingState: processingStateInitialized,
			readyForPoll:    true,
		},
		{
			state:           ctask.TaskStatePending,
			processingState: processingStateResponseReported,
			readyForPoll:    false,
		},
		{
			state:           ctask.TaskStatePending,
			processingState: processingStateResponseRecorded,
			readyForPoll:    true,
		},
		{
			state:           ctask.TaskStatePending,
			processingState: processingStateInvalidated,
			readyForPoll:    false,
		},
		{
			state:           ctask.TaskStateAcked,
			processingState: processingStateInitialized,
			readyForPoll:    false,
		},
	}

	for _, tc := range testCases {
		sourceTask := s.newTestSourceTask(
			cluster.TestAlternativeClusterName,
			&p.CrossClusterTaskInfo{
				DomainID: constants.TestDomainID,
			},
		)
		sourceTask.state = tc.state
		sourceTask.processingState = tc.processingState
		s.Equal(tc.readyForPoll, sourceTask.IsReadyForPoll())
	}
}

func (s *crossClusterTaskSuite) TestSourceTask_IsValid() {
	testCases := []struct {
		sourceDomainID string
		targetDomainID string
		targetCluster  string
		isValid        bool
	}{
		{
			sourceDomainID: constants.TestDomainID,
			targetDomainID: constants.TestRemoteTargetDomainID,
			targetCluster:  cluster.TestAlternativeClusterName,
			isValid:        true,
		},
		{
			sourceDomainID: constants.TestDomainID,
			targetDomainID: constants.TestTargetDomainID, // active in current cluster
			targetCluster:  cluster.TestAlternativeClusterName,
			isValid:        false,
		},
		{
			sourceDomainID: constants.TestRemoteTargetDomainID, // active is alternative cluster
			targetDomainID: constants.TestRemoteTargetDomainID,
			targetCluster:  cluster.TestAlternativeClusterName,
			isValid:        false,
		},
		{
			sourceDomainID: constants.TestDomainID,
			targetDomainID: constants.TestRemoteTargetDomainID,
			targetCluster:  "doesn't not match active cluster of target domain",
			isValid:        false,
		},
	}

	for _, tc := range testCases {
		sourceTask := s.newTestSourceTask(
			tc.targetCluster,
			&p.CrossClusterTaskInfo{
				DomainID:       tc.sourceDomainID,
				TargetDomainID: tc.targetDomainID,
			},
		)
		s.Equal(tc.isValid, sourceTask.IsValid())

		if !tc.isValid {
			s.Equal(processingStateInvalidated, sourceTask.ProcessingState())
		}
	}
}

func (s *crossClusterTaskSuite) TestSourceTask_RecordResponse() {
	testCases := []struct {
		response        *types.CrossClusterTaskResponse
		taskType        int
		state           ctask.State
		processingState processingState
		expectErr       bool
	}{
		{
			response: &types.CrossClusterTaskResponse{
				TaskType:                  types.CrossClusterTaskTypeSignalExecution.Ptr(),
				TaskState:                 int16(processingStateInitialized),
				FailedCause:               nil,
				SignalExecutionAttributes: &types.CrossClusterSignalExecutionResponseAttributes{},
			},
			taskType:        persistence.CrossClusterTaskTypeSignalExecution,
			state:           ctask.TaskStatePending,
			processingState: processingStateInitialized,
			expectErr:       false,
		},
		{
			response: &types.CrossClusterTaskResponse{
				TaskType:    types.CrossClusterTaskTypeStartChildExecution.Ptr(),
				TaskState:   int16(processingStateInitialized),
				FailedCause: types.CrossClusterTaskFailedCauseDomainNotActive.Ptr(),
			},
			taskType:        persistence.CrossClusterTaskTypeStartChildExecution,
			state:           ctask.TaskStatePending,
			processingState: processingStateInitialized,
			expectErr:       false,
		},
		{
			response: &types.CrossClusterTaskResponse{
				TaskType:                  types.CrossClusterTaskTypeSignalExecution.Ptr(),
				TaskState:                 int16(processingStateInitialized),
				FailedCause:               nil,
				SignalExecutionAttributes: &types.CrossClusterSignalExecutionResponseAttributes{},
			},
			taskType:        persistence.CrossClusterTaskTypeSignalExecution,
			state:           ctask.TaskStatePending,
			processingState: processingStateResponseReported, // response already reported
			expectErr:       true,
		},
		{
			response: &types.CrossClusterTaskResponse{
				TaskType:                  types.CrossClusterTaskTypeSignalExecution.Ptr(),
				TaskState:                 int16(processingStateInitialized),
				FailedCause:               nil,
				SignalExecutionAttributes: &types.CrossClusterSignalExecutionResponseAttributes{},
			},
			taskType:        persistence.CrossClusterTaskTypeSignalExecution,
			state:           ctask.TaskStatePending,
			processingState: processingStateInvalidated, // task already invalidated
			expectErr:       true,
		},
		{
			response: &types.CrossClusterTaskResponse{
				TaskType:                  types.CrossClusterTaskTypeSignalExecution.Ptr(),
				TaskState:                 int16(processingStateInitialized),
				FailedCause:               nil,
				SignalExecutionAttributes: &types.CrossClusterSignalExecutionResponseAttributes{},
			},
			taskType:        persistence.CrossClusterTaskTypeSignalExecution,
			state:           ctask.TaskStateAcked, // task already completed
			processingState: processingStateResponseReported,
			expectErr:       true,
		},
		{
			response: &types.CrossClusterTaskResponse{
				TaskType:  types.CrossClusterTaskTypeSignalExecution.Ptr(),
				TaskState: int16(processingStateInitialized),
				// empty response
			},
			taskType:        persistence.CrossClusterTaskTypeSignalExecution,
			state:           ctask.TaskStatePending,
			processingState: processingStateInitialized,
			expectErr:       true,
		},
		{
			response: &types.CrossClusterTaskResponse{
				TaskType:    types.CrossClusterTaskTypeSignalExecution.Ptr(),
				TaskState:   int16(processingStateInitialized),
				FailedCause: types.CrossClusterTaskFailedCauseUncategorized.Ptr(), // unexpected failed cause
			},
			taskType:        persistence.CrossClusterTaskTypeSignalExecution,
			state:           ctask.TaskStatePending,
			processingState: processingStateInitialized,
			expectErr:       true,
		},
		{
			response: &types.CrossClusterTaskResponse{
				TaskType:    types.CrossClusterTaskTypeCancelExecution.Ptr(),
				TaskState:   int16(processingStateInitialized),
				FailedCause: types.CrossClusterTaskFailedCauseWorkflowAlreadyRunning.Ptr(), // unexpected failed cause
			},
			taskType:        persistence.CrossClusterTaskTypeCancelExecution,
			state:           ctask.TaskStatePending,
			processingState: processingStateInitialized,
			expectErr:       true,
		},
	}

	for _, tc := range testCases {
		sourceTask := s.newTestSourceTask(
			cluster.TestAlternativeClusterName,
			&p.CrossClusterTaskInfo{
				DomainID: constants.TestDomainID,
				TaskType: tc.taskType,
			},
		)
		sourceTask.state = tc.state
		sourceTask.processingState = tc.processingState

		err := sourceTask.RecordResponse(tc.response)
		if tc.expectErr {
			s.Error(err)
			// state should not be changed
			s.Equal(tc.state, sourceTask.State())
			s.Equal(tc.processingState, sourceTask.ProcessingState())
		} else {
			s.NoError(err)
			s.Equal(tc.state, sourceTask.State())
			s.Equal(processingStateResponseReported, sourceTask.ProcessingState())
		}
	}
}

func (s *crossClusterTaskSuite) TestSourceTask_GetStartChildRequest_Invalidated() {
	s.testGetStartChildExecutionRequest(
		constants.TestRemoteTargetDomainID,
		nil,
		func(
			request *types.CrossClusterTaskRequest,
			getRequestErr error,
			sourceTask *crossClusterSourceTask,
			childInfo *p.ChildExecutionInfo,
			initiatedEvent *types.HistoryEvent,
		) {
			s.Error(getRequestErr)
			s.Nil(request)
			s.Equal(ctask.TaskStatePending, sourceTask.State())
			s.Equal(processingStateInvalidated, sourceTask.ProcessingState())
			s.False(sourceTask.IsValid())
		},
	)
}

func (s *crossClusterTaskSuite) TestSourceTask_GetStartChildRequest_AlreadyStarted() {
	childExecutionRunID := "some random target run ID"
	s.testGetStartChildExecutionRequest(
		constants.TestDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			sourceTask *crossClusterSourceTask,
			childInfo *p.ChildExecutionInfo,
		) {
			// child already started, advance processing to recorded
			lastEvent = test.AddChildWorkflowExecutionStartedEvent(mutableState, lastEvent.GetEventID(), constants.TestTargetDomainID, targetExecution.GetWorkflowID(), childExecutionRunID, childInfo.WorkflowTypeName)
			mutableState.FlushBufferedEvents()

			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
		},
		func(
			request *types.CrossClusterTaskRequest,
			getRequestErr error,
			sourceTask *crossClusterSourceTask,
			childInfo *p.ChildExecutionInfo,
			initiatedEvent *types.HistoryEvent,
		) {
			s.NoError(getRequestErr)
			s.NotNil(request)
			s.Equal(&types.CrossClusterTaskInfo{
				DomainID:            sourceTask.GetDomainID(),
				WorkflowID:          sourceTask.GetWorkflowID(),
				RunID:               sourceTask.GetRunID(),
				TaskType:            types.CrossClusterTaskTypeStartChildExecution.Ptr(),
				TaskState:           int16(processingStateResponseRecorded),
				TaskID:              sourceTask.GetTaskID(),
				VisibilityTimestamp: common.Int64Ptr(sourceTask.GetVisibilityTimestamp().UnixNano()),
			}, request.TaskInfo)
			taskInfo := sourceTask.GetInfo().(*p.CrossClusterTaskInfo)
			s.Equal(&types.CrossClusterStartChildExecutionRequestAttributes{
				TargetDomainID:           taskInfo.TargetDomainID,
				RequestID:                childInfo.CreateRequestID,
				InitiatedEventID:         taskInfo.ScheduleID,
				InitiatedEventAttributes: initiatedEvent.StartChildWorkflowExecutionInitiatedEventAttributes,
				TargetRunID:              common.StringPtr(childExecutionRunID),
			}, request.StartChildExecutionAttributes)
			s.Equal(ctask.TaskStatePending, sourceTask.State())
			s.Equal(processingStateResponseRecorded, sourceTask.ProcessingState())
		},
	)
}

func (s *crossClusterTaskSuite) TestSourceTask_GetStartChildRequest_AlreadyStarted_ParentClosed() {
	childExecutionRunID := "some random target run ID"
	s.testGetStartChildExecutionRequest(
		constants.TestDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			sourceTask *crossClusterSourceTask,
			childInfo *p.ChildExecutionInfo,
		) {
			// child already started, advance processing to recorded
			lastEvent = test.AddChildWorkflowExecutionStartedEvent(mutableState, lastEvent.GetEventID(), constants.TestTargetDomainID, targetExecution.GetWorkflowID(), childExecutionRunID, childInfo.WorkflowTypeName)
			di := test.AddDecisionTaskScheduledEvent(mutableState)
			lastEvent = test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, mutableState.GetExecutionInfo().TaskList, "some random identity")
			lastEvent = test.AddDecisionTaskCompletedEvent(mutableState, di.ScheduleID, lastEvent.GetEventID(), nil, "some random identity")
			lastEvent = test.AddCompleteWorkflowEvent(mutableState, lastEvent.EventID, nil)
			mutableState.FlushBufferedEvents()

			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
		},
		func(
			request *types.CrossClusterTaskRequest,
			getRequestErr error,
			sourceTask *crossClusterSourceTask,
			childInfo *p.ChildExecutionInfo,
			initiatedEvent *types.HistoryEvent,
		) {
			s.NoError(getRequestErr)
			s.NotNil(request)
			s.Equal(&types.CrossClusterTaskInfo{
				DomainID:            sourceTask.GetDomainID(),
				WorkflowID:          sourceTask.GetWorkflowID(),
				RunID:               sourceTask.GetRunID(),
				TaskType:            types.CrossClusterTaskTypeStartChildExecution.Ptr(),
				TaskState:           int16(processingStateResponseRecorded),
				TaskID:              sourceTask.GetTaskID(),
				VisibilityTimestamp: common.Int64Ptr(sourceTask.GetVisibilityTimestamp().UnixNano()),
			}, request.TaskInfo)
			taskInfo := sourceTask.GetInfo().(*p.CrossClusterTaskInfo)
			s.Equal(&types.CrossClusterStartChildExecutionRequestAttributes{
				TargetDomainID:           taskInfo.TargetDomainID,
				RequestID:                childInfo.CreateRequestID,
				InitiatedEventID:         taskInfo.ScheduleID,
				InitiatedEventAttributes: initiatedEvent.StartChildWorkflowExecutionInitiatedEventAttributes,
				TargetRunID:              common.StringPtr(childExecutionRunID),
			}, request.StartChildExecutionAttributes)
			s.Equal(ctask.TaskStatePending, sourceTask.State())
			s.Equal(processingStateResponseRecorded, sourceTask.ProcessingState())
		},
	)
}

func (s *crossClusterTaskSuite) TestSourceTask_GetStartChildRequest_Duplication() {
	s.testGetStartChildExecutionRequest(
		constants.TestDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			sourceTask *crossClusterSourceTask,
			childInfo *p.ChildExecutionInfo,
		) {
			// complete the child workflow, task should be no-op
			lastEvent = test.AddChildWorkflowExecutionStartedEvent(mutableState, lastEvent.GetEventID(), constants.TestTargetDomainID, targetExecution.GetWorkflowID(), targetExecution.GetRunID(), childInfo.WorkflowTypeName)
			childInfo.StartedID = lastEvent.GetEventID()
			lastEvent = test.AddChildWorkflowExecutionCompletedEvent(mutableState, childInfo.InitiatedID, &targetExecution, &types.WorkflowExecutionCompletedEventAttributes{
				Result:                       []byte("some random child workflow execution result"),
				DecisionTaskCompletedEventID: sourceTask.GetInfo().(*persistence.CrossClusterTaskInfo).ScheduleID,
			})
			mutableState.FlushBufferedEvents()

			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
		},
		func(
			request *types.CrossClusterTaskRequest,
			getRequestErr error,
			sourceTask *crossClusterSourceTask,
			childInfo *p.ChildExecutionInfo,
			initiatedEvent *types.HistoryEvent,
		) {
			s.NoError(getRequestErr)
			s.Nil(request)
			s.Equal(ctask.TaskStateAcked, sourceTask.State())
		},
	)
}

func (s *crossClusterTaskSuite) testGetStartChildExecutionRequest(
	sourceDomainID string,
	setupMockFn func(
		mutableState execution.MutableState,
		workflowExecution, targetExecution types.WorkflowExecution,
		lastEvent *types.HistoryEvent,
		sourceTask *crossClusterSourceTask,
		childInfo *p.ChildExecutionInfo,
	),
	validationFn func(
		request *types.CrossClusterTaskRequest,
		getRequestErr error,
		sourceTask *crossClusterSourceTask,
		childInfo *p.ChildExecutionInfo,
		initiatedEvent *types.HistoryEvent,
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

	sourceTask := s.newTestSourceTask(
		cluster.TestAlternativeClusterName,
		&p.CrossClusterTaskInfo{
			Version:          mutableState.GetCurrentVersion(),
			DomainID:         sourceDomainID,
			WorkflowID:       workflowExecution.GetWorkflowID(),
			RunID:            workflowExecution.GetRunID(),
			TargetDomainID:   constants.TestRemoteTargetDomainID,
			TargetWorkflowID: targetExecution.GetWorkflowID(),
			TaskID:           int64(59),
			TaskList:         mutableState.GetExecutionInfo().TaskList,
			TaskType:         p.CrossClusterTaskTypeStartChildExecution,
			ScheduleID:       event.GetEventID(),
		},
	)

	if setupMockFn != nil {
		setupMockFn(mutableState, workflowExecution, targetExecution, event, sourceTask, childInfo)
	}

	request, err := sourceTask.GetCrossClusterRequest()
	validationFn(request, err, sourceTask, childInfo, event)
}

func (s *crossClusterTaskSuite) TestSourceTask_GetCancelRequest_Duplication() {
	s.testGetCancelExecutionRequest(
		constants.TestDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			sourceTask *crossClusterSourceTask,
			cancelInfo *p.RequestCancelInfo,
		) {
			lastEvent = test.AddCancelRequestedEvent(mutableState, lastEvent.GetEventID(), constants.TestTargetDomainID, targetExecution.GetWorkflowID(), targetExecution.GetRunID())
			mutableState.FlushBufferedEvents()

			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
		},
		func(
			request *types.CrossClusterTaskRequest,
			getRequestErr error,
			sourceTask *crossClusterSourceTask,
			cancelInfo *p.RequestCancelInfo,
		) {
			s.NoError(getRequestErr)
			s.Nil(request)
			s.Equal(ctask.TaskStateAcked, sourceTask.State())
		},
	)
}

func (s *crossClusterTaskSuite) TestSourceTask_GetCancelRequest_Success() {
	s.testGetCancelExecutionRequest(
		constants.TestDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			sourceTask *crossClusterSourceTask,
			cancelInfo *p.RequestCancelInfo,
		) {
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
		},
		func(
			request *types.CrossClusterTaskRequest,
			getRequestErr error,
			sourceTask *crossClusterSourceTask,
			cancelInfo *p.RequestCancelInfo,
		) {
			s.NoError(getRequestErr)
			s.NotNil(request)
			s.Equal(&types.CrossClusterTaskInfo{
				DomainID:            sourceTask.GetDomainID(),
				WorkflowID:          sourceTask.GetWorkflowID(),
				RunID:               sourceTask.GetRunID(),
				TaskType:            types.CrossClusterTaskTypeCancelExecution.Ptr(),
				TaskState:           int16(processingStateInitialized),
				TaskID:              sourceTask.GetTaskID(),
				VisibilityTimestamp: common.Int64Ptr(sourceTask.GetVisibilityTimestamp().UnixNano()),
			}, request.TaskInfo)
			taskInfo := sourceTask.GetInfo().(*p.CrossClusterTaskInfo)
			s.Equal(&types.CrossClusterCancelExecutionRequestAttributes{
				TargetDomainID:    taskInfo.TargetDomainID,
				TargetWorkflowID:  taskInfo.TargetWorkflowID,
				TargetRunID:       taskInfo.TargetRunID,
				RequestID:         cancelInfo.CancelRequestID,
				InitiatedEventID:  taskInfo.ScheduleID,
				ChildWorkflowOnly: taskInfo.TargetChildWorkflowOnly,
			}, request.CancelExecutionAttributes)
			s.Equal(ctask.TaskStatePending, sourceTask.State())
			s.Equal(processingStateInitialized, sourceTask.ProcessingState())
		},
	)
}

func (s *crossClusterTaskSuite) testGetCancelExecutionRequest(
	sourceDomainID string,
	setupMockFn func(
		mutableState execution.MutableState,
		workflowExecution, targetExecution types.WorkflowExecution,
		lastEvent *types.HistoryEvent,
		sourceTask *crossClusterSourceTask,
		cancelInfo *p.RequestCancelInfo,
	),
	validationFn func(
		request *types.CrossClusterTaskRequest,
		getRequestErr error,
		sourceTask *crossClusterSourceTask,
		cancelInfo *p.RequestCancelInfo,
	),
) {
	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, sourceDomainID)
	s.NoError(err)
	targetExecution := types.WorkflowExecution{
		WorkflowID: "some random target workflow ID",
		RunID:      "some randome target run ID",
	}

	event, cancelInfo := test.AddRequestCancelInitiatedEvent(
		mutableState,
		decisionCompletionID,
		uuid.New(),
		constants.TestRemoteTargetDomainName,
		targetExecution.GetWorkflowID(),
		targetExecution.GetRunID(),
	)

	sourceTask := s.newTestSourceTask(
		cluster.TestAlternativeClusterName,
		&p.CrossClusterTaskInfo{
			Version:          mutableState.GetCurrentVersion(),
			DomainID:         sourceDomainID,
			WorkflowID:       workflowExecution.GetWorkflowID(),
			RunID:            workflowExecution.GetRunID(),
			TargetDomainID:   constants.TestRemoteTargetDomainID,
			TargetWorkflowID: targetExecution.GetWorkflowID(),
			TargetRunID:      targetExecution.GetRunID(),
			TaskID:           int64(59),
			TaskList:         mutableState.GetExecutionInfo().TaskList,
			TaskType:         p.CrossClusterTaskTypeCancelExecution,
			ScheduleID:       event.GetEventID(),
		},
	)

	if setupMockFn != nil {
		setupMockFn(mutableState, workflowExecution, targetExecution, event, sourceTask, cancelInfo)
	}

	request, err := sourceTask.GetCrossClusterRequest()
	validationFn(request, err, sourceTask, cancelInfo)
}

func (s *crossClusterTaskSuite) TestSourceTask_GetSignalRequest_Duplication() {
	s.testGetSignalExecutionRequest(
		constants.TestDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			sourceTask *crossClusterSourceTask,
			signalInfo *p.SignalInfo,
		) {
			lastEvent = test.AddSignaledEvent(mutableState, lastEvent.GetEventID(), constants.TestTargetDomainID, targetExecution.GetWorkflowID(), targetExecution.GetRunID(), signalInfo.Control)
			mutableState.FlushBufferedEvents()

			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
		},
		func(
			request *types.CrossClusterTaskRequest,
			getRequestErr error,
			sourceTask *crossClusterSourceTask,
			signalInfo *p.SignalInfo,
		) {
			s.NoError(getRequestErr)
			s.Nil(request)
			s.Equal(ctask.TaskStateAcked, sourceTask.State())
		},
	)
}

func (s *crossClusterTaskSuite) TestSourceTask_GetSignalRequest_Success() {
	s.testGetSignalExecutionRequest(
		constants.TestDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			sourceTask *crossClusterSourceTask,
			signalInfo *p.SignalInfo,
		) {
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, lastEvent.GetEventID(), lastEvent.GetVersion())
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
		},
		func(
			request *types.CrossClusterTaskRequest,
			getRequestErr error,
			sourceTask *crossClusterSourceTask,
			signalInfo *p.SignalInfo,
		) {
			s.NoError(getRequestErr)
			s.NotNil(request)
			s.Equal(&types.CrossClusterTaskInfo{
				DomainID:            sourceTask.GetDomainID(),
				WorkflowID:          sourceTask.GetWorkflowID(),
				RunID:               sourceTask.GetRunID(),
				TaskType:            types.CrossClusterTaskTypeSignalExecution.Ptr(),
				TaskState:           int16(processingStateInitialized),
				TaskID:              sourceTask.GetTaskID(),
				VisibilityTimestamp: common.Int64Ptr(sourceTask.GetVisibilityTimestamp().UnixNano()),
			}, request.TaskInfo)
			taskInfo := sourceTask.GetInfo().(*p.CrossClusterTaskInfo)
			s.Equal(&types.CrossClusterSignalExecutionRequestAttributes{
				TargetDomainID:    taskInfo.TargetDomainID,
				TargetWorkflowID:  taskInfo.TargetWorkflowID,
				TargetRunID:       taskInfo.TargetRunID,
				RequestID:         signalInfo.SignalRequestID,
				InitiatedEventID:  taskInfo.ScheduleID,
				ChildWorkflowOnly: taskInfo.TargetChildWorkflowOnly,
				SignalName:        signalInfo.SignalName,
				SignalInput:       signalInfo.Input,
				Control:           signalInfo.Control,
			}, request.SignalExecutionAttributes)
			s.Equal(ctask.TaskStatePending, sourceTask.State())
			s.Equal(processingStateInitialized, sourceTask.ProcessingState())
		},
	)
}

func (s *crossClusterTaskSuite) TestSourceTask_GetSignalRequest_Failure() {
	s.testGetSignalExecutionRequest(
		constants.TestDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			lastEvent *types.HistoryEvent,
			sourceTask *crossClusterSourceTask,
			signalInfo *p.SignalInfo,
		) {
			sourceTask.processingState = processingStateResponseRecorded
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(nil, errors.New("some random error"))
		},
		func(
			request *types.CrossClusterTaskRequest,
			getRequestErr error,
			sourceTask *crossClusterSourceTask,
			signalInfo *p.SignalInfo,
		) {
			s.Error(getRequestErr)
			s.Nil(request)
			s.Equal(ctask.TaskStatePending, sourceTask.State())
			s.Equal(processingStateResponseRecorded, sourceTask.ProcessingState())
		},
	)
}

func (s *crossClusterTaskSuite) testGetSignalExecutionRequest(
	sourceDomainID string,
	setupMockFn func(
		mutableState execution.MutableState,
		workflowExecution, targetExecution types.WorkflowExecution,
		lastEvent *types.HistoryEvent,
		sourceTask *crossClusterSourceTask,
		signalInfo *p.SignalInfo,
	),
	validationFn func(
		request *types.CrossClusterTaskRequest,
		getRequestErr error,
		sourceTask *crossClusterSourceTask,
		signalInfo *p.SignalInfo,
	),
) {
	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, sourceDomainID)
	s.NoError(err)
	targetExecution := types.WorkflowExecution{
		WorkflowID: "some random target workflow ID",
		RunID:      "some randome target run ID",
	}

	event, signalInfo := test.AddRequestSignalInitiatedEvent(
		mutableState,
		decisionCompletionID,
		uuid.New(),
		constants.TestRemoteTargetDomainName,
		targetExecution.GetWorkflowID(),
		targetExecution.GetRunID(),
		"some random signal name",
		[]byte("some random signal input"),
		[]byte("some random signal control"),
	)

	sourceTask := s.newTestSourceTask(
		cluster.TestAlternativeClusterName,
		&p.CrossClusterTaskInfo{
			Version:          mutableState.GetCurrentVersion(),
			DomainID:         sourceDomainID,
			WorkflowID:       workflowExecution.GetWorkflowID(),
			RunID:            workflowExecution.GetRunID(),
			TargetDomainID:   constants.TestRemoteTargetDomainID,
			TargetWorkflowID: targetExecution.GetWorkflowID(),
			TargetRunID:      targetExecution.GetRunID(),
			TaskID:           int64(59),
			TaskList:         mutableState.GetExecutionInfo().TaskList,
			TaskType:         p.CrossClusterTaskTypeSignalExecution,
			ScheduleID:       event.GetEventID(),
		},
	)

	if setupMockFn != nil {
		setupMockFn(mutableState, workflowExecution, targetExecution, event, sourceTask, signalInfo)
	}

	request, err := sourceTask.GetCrossClusterRequest()
	validationFn(request, err, sourceTask, signalInfo)
}

func (s *crossClusterTaskSuite) newTestSourceTask(
	targetCluster string,
	taskInfo *p.CrossClusterTaskInfo,
) *crossClusterSourceTask {
	return NewCrossClusterSourceTask(
		s.mockShard,
		targetCluster,
		s.executionCache,
		taskInfo,
		s.mockExecutor,
		s.mockProcessor,
		s.mockShard.GetLogger(),
		func(t Task) {
			s.mockRedispatcher.AddTask(t)
		},
		nil,
		dynamicconfig.GetIntPropertyFn(1),
	).(*crossClusterSourceTask)
}
