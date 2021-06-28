// Copyright (c) 2017-2020 Uber Technologies, Inc.
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
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/client/matching"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/archiver/provider"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/ndc"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
	test "github.com/uber/cadence/service/history/testing"
	warchiver "github.com/uber/cadence/service/worker/archiver"
)

type (
	transferStandbyTaskExecutorSuite struct {
		suite.Suite
		*require.Assertions

		controller             *gomock.Controller
		mockShard              *shard.TestContext
		mockDomainCache        *cache.MockDomainCache
		mockClusterMetadata    *cluster.MockMetadata
		mockNDCHistoryResender *ndc.MockHistoryResender
		mockMatchingClient     *matching.MockClient

		mockVisibilityMgr    *mocks.VisibilityManager
		mockExecutionMgr     *mocks.ExecutionManager
		mockArchivalClient   *warchiver.ClientMock
		mockArchivalMetadata *archiver.MockArchivalMetadata
		mockArchiverProvider *provider.MockArchiverProvider

		logger               log.Logger
		domainID             string
		domainEntry          *cache.DomainCacheEntry
		version              int64
		clusterName          string
		now                  time.Time
		timeSource           *clock.EventTimeSource
		fetchHistoryDuration time.Duration
		discardDuration      time.Duration

		transferStandbyTaskExecutor *transferStandbyTaskExecutor
	}
)

func TestTransferStandbyTaskExecutorSuite(t *testing.T) {
	s := new(transferStandbyTaskExecutorSuite)
	suite.Run(t, s)
}

func (s *transferStandbyTaskExecutorSuite) SetupSuite() {

}

func (s *transferStandbyTaskExecutorSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	config := config.NewForTest()
	s.domainID = constants.TestDomainID
	s.domainEntry = constants.TestGlobalDomainEntry
	s.version = s.domainEntry.GetFailoverVersion()
	s.now = time.Now()
	s.timeSource = clock.NewEventTimeSource().Update(s.now)
	s.fetchHistoryDuration = config.StandbyTaskMissingEventsResendDelay() +
		(config.StandbyTaskMissingEventsDiscardDelay()-config.StandbyTaskMissingEventsResendDelay())/2
	s.discardDuration = config.StandbyTaskMissingEventsDiscardDelay() * 2

	s.controller = gomock.NewController(s.T())
	s.mockNDCHistoryResender = ndc.NewMockHistoryResender(s.controller)

	s.mockShard = shard.NewTestContext(
		s.controller,
		&persistence.ShardInfo{
			RangeID:          1,
			TransferAckLevel: 0,
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
	s.mockShard.Resource.TimeSource = s.timeSource

	s.mockMatchingClient = s.mockShard.Resource.MatchingClient
	s.mockExecutionMgr = s.mockShard.Resource.ExecutionMgr
	s.mockVisibilityMgr = s.mockShard.Resource.VisibilityMgr
	s.mockClusterMetadata = s.mockShard.Resource.ClusterMetadata
	s.mockArchivalClient = &warchiver.ClientMock{}
	s.mockArchivalMetadata = s.mockShard.Resource.ArchivalMetadata
	s.mockArchiverProvider = s.mockShard.Resource.ArchiverProvider
	s.mockDomainCache = s.mockShard.Resource.DomainCache
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestDomainID).Return(constants.TestGlobalDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestDomainID).Return(constants.TestDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomain(constants.TestDomainName).Return(constants.TestGlobalDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestTargetDomainID).Return(constants.TestGlobalTargetDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestTargetDomainID).Return(constants.TestTargetDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(constants.TestTargetDomainName).Return(constants.TestTargetDomainID, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestParentDomainID).Return(constants.TestGlobalParentDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestParentDomainID).Return(constants.TestParentDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomain(constants.TestParentDomainName).Return(constants.TestGlobalParentDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestChildDomainID).Return(constants.TestGlobalChildDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestChildDomainID).Return(constants.TestChildDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(constants.TestChildDomainName).Return(constants.TestChildDomainID, nil).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetAllClusterInfo().Return(cluster.TestAllClusterInfo).AnyTimes()
	s.mockClusterMetadata.EXPECT().IsGlobalDomainEnabled().Return(true).AnyTimes()
	s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(s.version).Return(s.clusterName).AnyTimes()

	s.logger = s.mockShard.GetLogger()
	s.clusterName = cluster.TestAlternativeClusterName
	s.transferStandbyTaskExecutor = NewTransferStandbyTaskExecutor(
		s.mockShard,
		s.mockArchivalClient,
		execution.NewCache(s.mockShard),
		s.mockNDCHistoryResender,
		s.logger,
		s.mockShard.GetMetricsClient(),
		s.clusterName,
		config,
	).(*transferStandbyTaskExecutor)
}

func (s *transferStandbyTaskExecutorSuite) TearDownTest() {
	s.controller.Finish()
	s.mockShard.Finish(s.T())
	s.mockArchivalClient.AssertExpectations(s.T())
}

func (s *transferStandbyTaskExecutorSuite) TestProcessActivityTask_Pending() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		constants.TestGlobalDomainEntry,
	)
	_, err := mutableState.AddWorkflowExecutionStartedEvent(
		workflowExecution,
		&types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: s.domainID,
			StartRequest: &types.StartWorkflowExecutionRequest{
				WorkflowType:                        &types.WorkflowType{Name: workflowType},
				TaskList:                            &types.TaskList{Name: taskListName},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			},
		},
	)
	s.Nil(err)

	di := test.AddDecisionTaskScheduledEvent(mutableState)
	event := test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, taskListName, uuid.New())
	di.StartedID = event.GetEventID()
	event = test.AddDecisionTaskCompletedEvent(mutableState, di.ScheduleID, di.StartedID, nil, "some random identity")

	taskID := int64(59)
	activityID := "activity-1"
	activityType := "some random activity type"
	event, _ = test.AddActivityTaskScheduledEvent(mutableState, event.GetEventID(), activityID, activityType, taskListName, []byte{}, 1, 1, 1, 1)

	now := time.Now()
	transferTask := &persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              taskID,
		TaskList:            taskListName,
		TaskType:            persistence.TransferTaskTypeActivityTask,
		ScheduleID:          event.GetEventID(),
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Equal(ErrTaskRedispatch, err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessActivityTask_Pending_PushToMatching() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		constants.TestGlobalDomainEntry,
	)
	_, err := mutableState.AddWorkflowExecutionStartedEvent(
		workflowExecution,
		&types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: s.domainID,
			StartRequest: &types.StartWorkflowExecutionRequest{
				WorkflowType:                        &types.WorkflowType{Name: workflowType},
				TaskList:                            &types.TaskList{Name: taskListName},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			},
		},
	)
	s.Nil(err)

	di := test.AddDecisionTaskScheduledEvent(mutableState)
	event := test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, taskListName, uuid.New())
	di.StartedID = event.GetEventID()
	event = test.AddDecisionTaskCompletedEvent(mutableState, di.ScheduleID, di.StartedID, nil, "some random identity")

	taskID := int64(59)
	activityID := "activity-1"
	activityType := "some random activity type"
	event, _ = test.AddActivityTaskScheduledEvent(mutableState, event.GetEventID(), activityID, activityType, taskListName, []byte{}, 1, 1, 1, 1)

	now := time.Now()
	s.mockShard.SetCurrentTime(s.clusterName, now.Add(s.fetchHistoryDuration))
	transferTask := &persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              taskID,
		TaskList:            taskListName,
		TaskType:            persistence.TransferTaskTypeActivityTask,
		ScheduleID:          event.GetEventID(),
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockMatchingClient.EXPECT().AddActivityTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessActivityTask_Success() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		constants.TestGlobalDomainEntry,
	)
	_, err := mutableState.AddWorkflowExecutionStartedEvent(
		workflowExecution,
		&types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: s.domainID,
			StartRequest: &types.StartWorkflowExecutionRequest{
				WorkflowType:                        &types.WorkflowType{Name: workflowType},
				TaskList:                            &types.TaskList{Name: taskListName},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			},
		},
	)
	s.Nil(err)

	di := test.AddDecisionTaskScheduledEvent(mutableState)
	event := test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, taskListName, uuid.New())
	di.StartedID = event.GetEventID()
	event = test.AddDecisionTaskCompletedEvent(mutableState, di.ScheduleID, di.StartedID, nil, "some random identity")

	taskID := int64(59)
	activityID := "activity-1"
	activityType := "some random activity type"
	event, _ = test.AddActivityTaskScheduledEvent(mutableState, event.GetEventID(), activityID, activityType, taskListName, []byte{}, 1, 1, 1, 1)

	now := time.Now()
	transferTask := &persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              taskID,
		TaskList:            taskListName,
		TaskType:            persistence.TransferTaskTypeActivityTask,
		ScheduleID:          event.GetEventID(),
	}

	event = test.AddActivityTaskStartedEvent(mutableState, event.GetEventID(), "")
	mutableState.FlushBufferedEvents()

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessDecisionTask_Pending() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		constants.TestGlobalDomainEntry,
	)
	_, err := mutableState.AddWorkflowExecutionStartedEvent(
		workflowExecution,
		&types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: s.domainID,
			StartRequest: &types.StartWorkflowExecutionRequest{
				WorkflowType:                        &types.WorkflowType{Name: workflowType},
				TaskList:                            &types.TaskList{Name: taskListName},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			},
		},
	)
	s.Nil(err)

	taskID := int64(59)
	di := test.AddDecisionTaskScheduledEvent(mutableState)

	now := time.Now()
	transferTask := &persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              taskID,
		TaskList:            taskListName,
		TaskType:            persistence.TransferTaskTypeDecisionTask,
		ScheduleID:          di.ScheduleID,
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, di.ScheduleID, di.Version)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Equal(ErrTaskRedispatch, err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessDecisionTask_Pending_PushToMatching() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		constants.TestGlobalDomainEntry,
	)
	_, err := mutableState.AddWorkflowExecutionStartedEvent(
		workflowExecution,
		&types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: s.domainID,
			StartRequest: &types.StartWorkflowExecutionRequest{
				WorkflowType:                        &types.WorkflowType{Name: workflowType},
				TaskList:                            &types.TaskList{Name: taskListName},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			},
		},
	)
	s.Nil(err)

	taskID := int64(59)
	di := test.AddDecisionTaskScheduledEvent(mutableState)

	now := time.Now()
	s.mockShard.SetCurrentTime(s.clusterName, now.Add(s.fetchHistoryDuration))
	transferTask := &persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              taskID,
		TaskList:            taskListName,
		TaskType:            persistence.TransferTaskTypeDecisionTask,
		ScheduleID:          di.ScheduleID,
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, di.ScheduleID, di.Version)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockMatchingClient.EXPECT().AddDecisionTask(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessDecisionTask_Success_FirstDecision() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		constants.TestGlobalDomainEntry,
	)
	_, err := mutableState.AddWorkflowExecutionStartedEvent(
		workflowExecution,
		&types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: s.domainID,
			StartRequest: &types.StartWorkflowExecutionRequest{
				WorkflowType:                        &types.WorkflowType{Name: workflowType},
				TaskList:                            &types.TaskList{Name: taskListName},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			},
		},
	)
	s.Nil(err)

	taskID := int64(59)
	di := test.AddDecisionTaskScheduledEvent(mutableState)

	now := time.Now()
	transferTask := &persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              taskID,
		TaskList:            taskListName,
		TaskType:            persistence.TransferTaskTypeDecisionTask,
		ScheduleID:          di.ScheduleID,
	}

	event := test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, taskListName, uuid.New())
	di.StartedID = event.GetEventID()

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessDecisionTask_Success_NonFirstDecision() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		constants.TestGlobalDomainEntry,
	)
	_, err := mutableState.AddWorkflowExecutionStartedEvent(
		workflowExecution,
		&types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: s.domainID,
			StartRequest: &types.StartWorkflowExecutionRequest{
				WorkflowType:                        &types.WorkflowType{Name: workflowType},
				TaskList:                            &types.TaskList{Name: taskListName},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			},
		},
	)
	s.Nil(err)

	di := test.AddDecisionTaskScheduledEvent(mutableState)
	event := test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, taskListName, uuid.New())
	di.StartedID = event.GetEventID()
	test.AddDecisionTaskCompletedEvent(mutableState, di.ScheduleID, di.StartedID, nil, "some random identity")

	taskID := int64(59)
	di = test.AddDecisionTaskScheduledEvent(mutableState)

	now := time.Now()
	transferTask := &persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              taskID,
		TaskList:            taskListName,
		TaskType:            persistence.TransferTaskTypeDecisionTask,
		ScheduleID:          di.ScheduleID,
	}

	event = test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, taskListName, uuid.New())
	di.StartedID = event.GetEventID()

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessCloseExecution() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		constants.TestGlobalDomainEntry,
	)
	_, err := mutableState.AddWorkflowExecutionStartedEvent(
		workflowExecution,
		&types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: s.domainID,
			StartRequest: &types.StartWorkflowExecutionRequest{
				WorkflowType:                        &types.WorkflowType{Name: workflowType},
				TaskList:                            &types.TaskList{Name: taskListName},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			},
		},
	)
	s.Nil(err)

	di := test.AddDecisionTaskScheduledEvent(mutableState)
	event := test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, taskListName, uuid.New())
	di.StartedID = event.GetEventID()
	event = test.AddDecisionTaskCompletedEvent(mutableState, di.ScheduleID, di.StartedID, nil, "some random identity")

	taskID := int64(59)
	event = test.AddCompleteWorkflowEvent(mutableState, event.GetEventID(), nil)

	now := time.Now()
	transferTask := &persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              taskID,
		TaskList:            taskListName,
		TaskType:            persistence.TransferTaskTypeCloseExecution,
		ScheduleID:          event.GetEventID(),
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockVisibilityMgr.On("RecordWorkflowExecutionClosed", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewDisabledArchvialConfig())

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessCancelExecution_Pending() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	targetExecution := types.WorkflowExecution{
		WorkflowID: "some random target workflow ID",
		RunID:      uuid.New(),
	}

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		constants.TestGlobalDomainEntry,
	)
	_, err := mutableState.AddWorkflowExecutionStartedEvent(
		workflowExecution,
		&types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: s.domainID,
			StartRequest: &types.StartWorkflowExecutionRequest{
				WorkflowType:                        &types.WorkflowType{Name: workflowType},
				TaskList:                            &types.TaskList{Name: taskListName},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			},
		},
	)
	s.Nil(err)

	di := test.AddDecisionTaskScheduledEvent(mutableState)
	event := test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, taskListName, uuid.New())
	di.StartedID = event.GetEventID()
	event = test.AddDecisionTaskCompletedEvent(mutableState, di.ScheduleID, di.StartedID, nil, "some random identity")
	taskID := int64(59)
	event, _ = test.AddRequestCancelInitiatedEvent(mutableState, event.GetEventID(), uuid.New(), constants.TestTargetDomainName, targetExecution.GetWorkflowID(), targetExecution.GetRunID())
	nextEventID := event.GetEventID()

	now := time.Now()
	transferTask := &persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TargetDomainID:      constants.TestTargetDomainID,
		TargetWorkflowID:    targetExecution.GetWorkflowID(),
		TargetRunID:         targetExecution.GetRunID(),
		TaskID:              taskID,
		TaskList:            taskListName,
		TaskType:            persistence.TransferTaskTypeCancelExecution,
		ScheduleID:          event.GetEventID(),
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Equal(ErrTaskRedispatch, err)

	s.mockShard.SetCurrentTime(s.clusterName, now.Add(s.fetchHistoryDuration))
	s.mockNDCHistoryResender.EXPECT().SendSingleWorkflowHistory(
		transferTask.GetDomainID(),
		transferTask.GetWorkflowID(),
		transferTask.GetRunID(),
		common.Int64Ptr(nextEventID),
		common.Int64Ptr(s.version),
		nil,
		nil,
	).Return(nil).Times(1)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Equal(ErrTaskRedispatch, err)

	s.mockShard.SetCurrentTime(s.clusterName, now.Add(s.discardDuration))
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Equal(ErrTaskDiscarded, err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessCancelExecution_Success() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	targetExecution := types.WorkflowExecution{
		WorkflowID: "some random target workflow ID",
		RunID:      uuid.New(),
	}

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		constants.TestGlobalDomainEntry,
	)
	_, err := mutableState.AddWorkflowExecutionStartedEvent(
		workflowExecution,
		&types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: s.domainID,
			StartRequest: &types.StartWorkflowExecutionRequest{
				WorkflowType:                        &types.WorkflowType{Name: workflowType},
				TaskList:                            &types.TaskList{Name: taskListName},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			},
		},
	)
	s.Nil(err)

	di := test.AddDecisionTaskScheduledEvent(mutableState)
	event := test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, taskListName, uuid.New())
	di.StartedID = event.GetEventID()
	event = test.AddDecisionTaskCompletedEvent(mutableState, di.ScheduleID, di.StartedID, nil, "some random identity")

	taskID := int64(59)
	event, _ = test.AddRequestCancelInitiatedEvent(mutableState, event.GetEventID(), uuid.New(), constants.TestTargetDomainName, targetExecution.GetWorkflowID(), targetExecution.GetRunID())

	now := time.Now()
	transferTask := &persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TargetDomainID:      constants.TestTargetDomainID,
		TargetWorkflowID:    targetExecution.GetWorkflowID(),
		TargetRunID:         targetExecution.GetRunID(),
		TaskID:              taskID,
		TaskList:            taskListName,
		TaskType:            persistence.TransferTaskTypeCancelExecution,
		ScheduleID:          event.GetEventID(),
	}

	event = test.AddCancelRequestedEvent(mutableState, event.GetEventID(), constants.TestTargetDomainID, targetExecution.GetWorkflowID(), targetExecution.GetRunID())
	mutableState.FlushBufferedEvents()

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessSignalExecution_Pending() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	targetExecution := types.WorkflowExecution{
		WorkflowID: "some random target workflow ID",
		RunID:      uuid.New(),
	}
	signalName := "some random signal name"

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		constants.TestGlobalDomainEntry,
	)
	_, err := mutableState.AddWorkflowExecutionStartedEvent(
		workflowExecution,
		&types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: s.domainID,
			StartRequest: &types.StartWorkflowExecutionRequest{
				WorkflowType:                        &types.WorkflowType{Name: workflowType},
				TaskList:                            &types.TaskList{Name: taskListName},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			},
		},
	)
	s.Nil(err)

	di := test.AddDecisionTaskScheduledEvent(mutableState)
	event := test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, taskListName, uuid.New())
	di.StartedID = event.GetEventID()
	event = test.AddDecisionTaskCompletedEvent(mutableState, di.ScheduleID, di.StartedID, nil, "some random identity")

	taskID := int64(59)
	event, _ = test.AddRequestSignalInitiatedEvent(mutableState, event.GetEventID(), uuid.New(),
		constants.TestTargetDomainName, targetExecution.GetWorkflowID(), targetExecution.GetRunID(), signalName, nil, nil)
	nextEventID := event.GetEventID()

	now := time.Now()
	transferTask := &persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TargetDomainID:      constants.TestTargetDomainID,
		TargetWorkflowID:    targetExecution.GetWorkflowID(),
		TargetRunID:         targetExecution.GetRunID(),
		TaskID:              taskID,
		TaskList:            taskListName,
		TaskType:            persistence.TransferTaskTypeSignalExecution,
		ScheduleID:          event.GetEventID(),
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Equal(ErrTaskRedispatch, err)

	s.mockShard.SetCurrentTime(s.clusterName, now.Add(s.fetchHistoryDuration))
	s.mockNDCHistoryResender.EXPECT().SendSingleWorkflowHistory(
		transferTask.GetDomainID(),
		transferTask.GetWorkflowID(),
		transferTask.GetRunID(),
		common.Int64Ptr(nextEventID),
		common.Int64Ptr(s.version),
		nil,
		nil,
	).Return(nil).Times(1)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Equal(ErrTaskRedispatch, err)

	s.mockShard.SetCurrentTime(s.clusterName, now.Add(s.discardDuration))
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Equal(ErrTaskDiscarded, err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessSignalExecution_Success() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	targetExecution := types.WorkflowExecution{
		WorkflowID: "some random target workflow ID",
		RunID:      uuid.New(),
	}
	signalName := "some random signal name"

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		constants.TestGlobalDomainEntry,
	)
	_, err := mutableState.AddWorkflowExecutionStartedEvent(
		workflowExecution,
		&types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: s.domainID,
			StartRequest: &types.StartWorkflowExecutionRequest{
				WorkflowType:                        &types.WorkflowType{Name: workflowType},
				TaskList:                            &types.TaskList{Name: taskListName},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			},
		},
	)
	s.Nil(err)

	di := test.AddDecisionTaskScheduledEvent(mutableState)
	event := test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, taskListName, uuid.New())
	di.StartedID = event.GetEventID()
	event = test.AddDecisionTaskCompletedEvent(mutableState, di.ScheduleID, di.StartedID, nil, "some random identity")

	taskID := int64(59)
	event, _ = test.AddRequestSignalInitiatedEvent(mutableState, event.GetEventID(), uuid.New(),
		constants.TestTargetDomainName, targetExecution.GetWorkflowID(), targetExecution.GetRunID(), signalName, nil, nil)

	now := time.Now()
	transferTask := &persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TargetDomainID:      constants.TestTargetDomainID,
		TargetWorkflowID:    targetExecution.GetWorkflowID(),
		TargetRunID:         targetExecution.GetRunID(),
		TaskID:              taskID,
		TaskList:            taskListName,
		TaskType:            persistence.TransferTaskTypeSignalExecution,
		ScheduleID:          event.GetEventID(),
	}

	event = test.AddSignaledEvent(mutableState, event.GetEventID(), constants.TestTargetDomainName, targetExecution.GetWorkflowID(), targetExecution.GetRunID(), nil)
	mutableState.FlushBufferedEvents()

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessStartChildExecution_Pending() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	childWorkflowID := "some random child workflow ID"
	childWorkflowType := "some random child workflow type"
	childTaskListName := "some random child task list"

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		constants.TestGlobalDomainEntry,
	)
	_, err := mutableState.AddWorkflowExecutionStartedEvent(
		workflowExecution,
		&types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: s.domainID,
			StartRequest: &types.StartWorkflowExecutionRequest{
				WorkflowType:                        &types.WorkflowType{Name: workflowType},
				TaskList:                            &types.TaskList{Name: taskListName},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			},
		},
	)
	s.Nil(err)

	di := test.AddDecisionTaskScheduledEvent(mutableState)
	event := test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, taskListName, uuid.New())
	di.StartedID = event.GetEventID()
	event = test.AddDecisionTaskCompletedEvent(mutableState, di.ScheduleID, di.StartedID, nil, "some random identity")

	taskID := int64(59)
	event, _ = test.AddStartChildWorkflowExecutionInitiatedEvent(mutableState, event.GetEventID(), uuid.New(),
		constants.TestChildDomainName, childWorkflowID, childWorkflowType, childTaskListName, nil, 1, 1, nil)
	nextEventID := event.GetEventID()

	now := time.Now()
	transferTask := &persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TargetDomainID:      constants.TestChildDomainID,
		TargetWorkflowID:    childWorkflowID,
		TargetRunID:         "",
		TaskID:              taskID,
		TaskList:            taskListName,
		TaskType:            persistence.TransferTaskTypeStartChildExecution,
		ScheduleID:          event.GetEventID(),
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Equal(ErrTaskRedispatch, err)

	s.mockShard.SetCurrentTime(s.clusterName, now.Add(s.fetchHistoryDuration))
	s.mockNDCHistoryResender.EXPECT().SendSingleWorkflowHistory(
		transferTask.GetDomainID(),
		transferTask.GetWorkflowID(),
		transferTask.GetRunID(),
		common.Int64Ptr(nextEventID),
		common.Int64Ptr(s.version),
		nil,
		nil,
	).Return(nil).Times(1)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Equal(ErrTaskRedispatch, err)

	s.mockShard.SetCurrentTime(s.clusterName, now.Add(s.discardDuration))
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Equal(ErrTaskDiscarded, err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessStartChildExecution_Success() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	childWorkflowID := "some random child workflow ID"
	childWorkflowType := "some random child workflow type"
	childTaskListName := "some random child task list"

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		constants.TestGlobalDomainEntry,
	)
	_, err := mutableState.AddWorkflowExecutionStartedEvent(
		workflowExecution,
		&types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: s.domainID,
			StartRequest: &types.StartWorkflowExecutionRequest{
				WorkflowType:                        &types.WorkflowType{Name: workflowType},
				TaskList:                            &types.TaskList{Name: taskListName},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			},
		},
	)
	s.Nil(err)

	di := test.AddDecisionTaskScheduledEvent(mutableState)
	event := test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, taskListName, uuid.New())
	di.StartedID = event.GetEventID()
	event = test.AddDecisionTaskCompletedEvent(mutableState, di.ScheduleID, di.StartedID, nil, "some random identity")

	taskID := int64(59)
	event, childInfo := test.AddStartChildWorkflowExecutionInitiatedEvent(mutableState, event.GetEventID(), uuid.New(),
		constants.TestChildDomainName, childWorkflowID, childWorkflowType, childTaskListName, nil, 1, 1, nil)

	now := time.Now()
	transferTask := &persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TargetDomainID:      constants.TestChildDomainID,
		TargetWorkflowID:    childWorkflowID,
		TargetRunID:         "",
		TaskID:              taskID,
		TaskList:            taskListName,
		TaskType:            persistence.TransferTaskTypeStartChildExecution,
		ScheduleID:          event.GetEventID(),
	}

	event = test.AddChildWorkflowExecutionStartedEvent(mutableState, event.GetEventID(), constants.TestChildDomainName, childWorkflowID, uuid.New(), childWorkflowType)
	mutableState.FlushBufferedEvents()
	childInfo.StartedID = event.GetEventID()

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessRecordWorkflowStartedTask() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		constants.TestGlobalDomainEntry,
	)
	event, err := mutableState.AddWorkflowExecutionStartedEvent(
		workflowExecution,
		&types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: s.domainID,
			StartRequest: &types.StartWorkflowExecutionRequest{
				WorkflowType:                        &types.WorkflowType{Name: workflowType},
				TaskList:                            &types.TaskList{Name: taskListName},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			},
		},
	)
	s.Nil(err)

	taskID := int64(59)

	di := test.AddDecisionTaskScheduledEvent(mutableState)

	now := time.Now()
	transferTask := &persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              taskID,
		TaskList:            taskListName,
		TaskType:            persistence.TransferTaskTypeRecordWorkflowStarted,
		ScheduleID:          event.GetEventID(),
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, di.ScheduleID, di.Version)
	executionInfo := mutableState.GetExecutionInfo()
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockVisibilityMgr.On("RecordWorkflowExecutionStarted", mock.Anything, &persistence.RecordWorkflowExecutionStartedRequest{
		DomainUUID: constants.TestDomainID,
		Domain:     constants.TestDomainName,
		Execution: types.WorkflowExecution{
			WorkflowID: executionInfo.WorkflowID,
			RunID:      executionInfo.RunID,
		},
		WorkflowTypeName: executionInfo.WorkflowTypeName,
		StartTimestamp:   event.GetTimestamp(),
		WorkflowTimeout:  int64(executionInfo.WorkflowTimeout),
		TaskID:           taskID,
		TaskList:         taskListName,
		IsCron:           false,
	}).Return(nil).Once()

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) TestProcessUpsertWorkflowSearchAttributesTask() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		constants.TestGlobalDomainEntry,
	)
	event, err := mutableState.AddWorkflowExecutionStartedEvent(
		workflowExecution,
		&types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: s.domainID,
			StartRequest: &types.StartWorkflowExecutionRequest{
				WorkflowType:                        &types.WorkflowType{Name: workflowType},
				TaskList:                            &types.TaskList{Name: taskListName},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			},
		},
	)
	s.Nil(err)

	taskID := int64(59)

	di := test.AddDecisionTaskScheduledEvent(mutableState)

	now := time.Now()
	transferTask := &persistence.TransferTaskInfo{
		Version:             s.version,
		DomainID:            s.domainID,
		WorkflowID:          workflowExecution.GetWorkflowID(),
		RunID:               workflowExecution.GetRunID(),
		VisibilityTimestamp: now,
		TaskID:              taskID,
		TaskList:            taskListName,
		TaskType:            persistence.TransferTaskTypeUpsertWorkflowSearchAttributes,
		ScheduleID:          event.GetEventID(),
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, di.ScheduleID, di.Version)
	executionInfo := mutableState.GetExecutionInfo()
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockVisibilityMgr.On("UpsertWorkflowExecution", mock.Anything, &persistence.UpsertWorkflowExecutionRequest{
		DomainUUID: constants.TestDomainID,
		Domain:     constants.TestDomainName,
		Execution: types.WorkflowExecution{
			WorkflowID: executionInfo.WorkflowID,
			RunID:      executionInfo.RunID,
		},
		WorkflowTypeName: executionInfo.WorkflowTypeName,
		StartTimestamp:   event.GetTimestamp(),
		WorkflowTimeout:  int64(executionInfo.WorkflowTimeout),
		TaskID:           taskID,
		TaskList:         taskListName,
	}).Return(nil).Once()

	s.mockShard.SetCurrentTime(s.clusterName, now)
	err = s.transferStandbyTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferStandbyTaskExecutorSuite) createPersistenceMutableState(
	ms execution.MutableState,
	lastEventID int64,
	lastEventVersion int64,
) *persistence.WorkflowMutableState {

	if ms.GetVersionHistories() != nil {
		currentVersionHistory, err := ms.GetVersionHistories().GetCurrentVersionHistory()
		s.NoError(err)
		err = currentVersionHistory.AddOrUpdateItem(persistence.NewVersionHistoryItem(
			lastEventID, lastEventVersion,
		))
		s.NoError(err)
	}

	return execution.CreatePersistenceMutableState(ms)
}
