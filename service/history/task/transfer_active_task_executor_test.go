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

package task

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	hclient "github.com/uber/cadence/client/history"
	"github.com/uber/cadence/client/matching"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/archiver/provider"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	dc "github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/engine"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
	test "github.com/uber/cadence/service/history/testing"
	warchiver "github.com/uber/cadence/service/worker/archiver"
	"github.com/uber/cadence/service/worker/parentclosepolicy"
)

type (
	transferActiveTaskExecutorSuite struct {
		suite.Suite
		*require.Assertions

		controller          *gomock.Controller
		mockShard           *shard.TestContext
		mockEngine          *engine.MockEngine
		mockDomainCache     *cache.MockDomainCache
		mockHistoryClient   *hclient.MockClient
		mockMatchingClient  *matching.MockClient
		mockClusterMetadata *cluster.MockMetadata

		mockVisibilityMgr           *mocks.VisibilityManager
		mockExecutionMgr            *mocks.ExecutionManager
		mockHistoryV2Mgr            *mocks.HistoryV2Manager
		mockArchivalClient          *warchiver.ClientMock
		mockArchivalMetadata        *archiver.MockArchivalMetadata
		mockArchiverProvider        *provider.MockArchiverProvider
		mockParentClosePolicyClient *parentclosepolicy.ClientMock

		logger                     log.Logger
		domainID                   string
		domainName                 string
		domainEntry                *cache.DomainCacheEntry
		targetDomainID             string
		targetDomainName           string
		targetDomainEntry          *cache.DomainCacheEntry
		remoteTargetDomainID       string
		remoteTargetDomainName     string
		remoteTargetDomainEntry    *cache.DomainCacheEntry
		childDomainID              string
		childDomainName            string
		childDomainEntry           *cache.DomainCacheEntry
		version                    int64
		now                        time.Time
		timeSource                 *clock.EventTimeSource
		transferActiveTaskExecutor *transferActiveTaskExecutor
	}
)

func TestTransferActiveTaskExecutorSuite(t *testing.T) {
	s := new(transferActiveTaskExecutorSuite)
	suite.Run(t, s)
}

func (s *transferActiveTaskExecutorSuite) SetupSuite() {

}

func (s *transferActiveTaskExecutorSuite) TearDownSuite() {

}

func (s *transferActiveTaskExecutorSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.domainID = constants.TestDomainID
	s.domainName = constants.TestDomainName
	s.domainEntry = constants.TestGlobalDomainEntry
	s.targetDomainID = constants.TestTargetDomainID
	s.targetDomainName = constants.TestTargetDomainName
	s.targetDomainEntry = constants.TestGlobalTargetDomainEntry
	s.remoteTargetDomainID = constants.TestRemoteTargetDomainID
	s.remoteTargetDomainName = constants.TestRemoteTargetDomainName
	s.remoteTargetDomainEntry = constants.TestGlobalRemoteTargetDomainEntry
	s.childDomainID = constants.TestChildDomainID
	s.childDomainName = constants.TestChildDomainName
	s.childDomainEntry = constants.TestGlobalChildDomainEntry
	s.version = s.domainEntry.GetFailoverVersion()
	s.now = time.Now()
	s.timeSource = clock.NewEventTimeSource().Update(s.now)

	s.controller = gomock.NewController(s.T())

	config := config.NewForTest()
	s.mockShard = shard.NewTestContext(
		s.controller,
		&persistence.ShardInfo{
			ShardID:          0,
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

	s.mockEngine = engine.NewMockEngine(s.controller)
	s.mockEngine.EXPECT().NotifyNewHistoryEvent(gomock.Any()).AnyTimes()
	s.mockEngine.EXPECT().NotifyNewTransferTasks(gomock.Any(), gomock.Any()).AnyTimes()
	s.mockEngine.EXPECT().NotifyNewTimerTasks(gomock.Any(), gomock.Any()).AnyTimes()
	s.mockEngine.EXPECT().NotifyNewCrossClusterTasks(gomock.Any(), gomock.Any()).AnyTimes()
	s.mockShard.SetEngine(s.mockEngine)

	s.mockParentClosePolicyClient = &parentclosepolicy.ClientMock{}
	s.mockArchivalClient = &warchiver.ClientMock{}
	s.mockMatchingClient = s.mockShard.Resource.MatchingClient
	s.mockHistoryClient = s.mockShard.Resource.HistoryClient
	s.mockExecutionMgr = s.mockShard.Resource.ExecutionMgr
	s.mockHistoryV2Mgr = s.mockShard.Resource.HistoryMgr
	s.mockVisibilityMgr = s.mockShard.Resource.VisibilityMgr
	s.mockClusterMetadata = s.mockShard.Resource.ClusterMetadata
	s.mockArchivalMetadata = s.mockShard.Resource.ArchivalMetadata
	s.mockArchiverProvider = s.mockShard.Resource.ArchiverProvider
	s.mockDomainCache = s.mockShard.Resource.DomainCache
	s.mockDomainCache.EXPECT().GetDomainByID(s.domainID).Return(s.domainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(s.domainID).Return(s.domainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(s.domainName).Return(s.domainID, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(s.targetDomainID).Return(s.targetDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(s.targetDomainID).Return(s.targetDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(s.targetDomainName).Return(s.targetDomainID, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(s.remoteTargetDomainID).Return(s.remoteTargetDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(s.remoteTargetDomainID).Return(s.remoteTargetDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(s.remoteTargetDomainName).Return(s.remoteTargetDomainID, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestParentDomainID).Return(constants.TestGlobalParentDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestParentDomainID).Return(constants.TestParentDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(constants.TestParentDomainName).Return(constants.TestParentDomainID, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(s.childDomainID).Return(s.childDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(s.childDomainID).Return(s.childDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(s.childDomainName).Return(s.childDomainID, nil).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetAllClusterInfo().Return(cluster.TestAllClusterInfo).AnyTimes()
	s.mockClusterMetadata.EXPECT().IsGlobalDomainEnabled().Return(true).AnyTimes()
	s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(s.version).Return(s.mockClusterMetadata.GetCurrentClusterName()).AnyTimes()

	s.logger = s.mockShard.GetLogger()
	s.transferActiveTaskExecutor = NewTransferActiveTaskExecutor(
		s.mockShard,
		s.mockArchivalClient,
		execution.NewCache(s.mockShard),
		nil,
		s.logger,
		config,
	).(*transferActiveTaskExecutor)
	s.transferActiveTaskExecutor.parentClosePolicyClient = s.mockParentClosePolicyClient
}

func (s *transferActiveTaskExecutorSuite) TearDownTest() {
	s.controller.Finish()
	s.mockShard.Finish(s.T())
	s.mockArchivalClient.AssertExpectations(s.T())
}

func (s *transferActiveTaskExecutorSuite) TestProcessActivityTask_Success() {

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
		s.domainEntry,
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
	event, ai := test.AddActivityTaskScheduledEvent(mutableState, event.GetEventID(), activityID, activityType, taskListName, []byte{}, 1, 1, 1, 1)
	mutableState.FlushBufferedEvents()
	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:        s.version,
		DomainID:       s.domainID,
		TargetDomainID: s.targetDomainID,
		WorkflowID:     workflowExecution.GetWorkflowID(),
		RunID:          workflowExecution.GetRunID(),
		TaskID:         taskID,
		TaskList:       taskListName,
		TaskType:       persistence.TransferTaskTypeActivityTask,
		ScheduleID:     event.GetEventID(),
	})

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockMatchingClient.EXPECT().AddActivityTask(gomock.Any(), s.createAddActivityTaskRequest(transferTask, ai)).Return(nil).Times(1)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessActivityTask_Duplication() {

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
		s.domainEntry,
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
	event, ai := test.AddActivityTaskScheduledEvent(mutableState, event.GetEventID(), activityID, activityType, taskListName, []byte{}, 1, 1, 1, 1)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:        s.version,
		DomainID:       s.domainID,
		TargetDomainID: s.targetDomainID,
		WorkflowID:     workflowExecution.GetWorkflowID(),
		RunID:          workflowExecution.GetRunID(),
		TaskID:         taskID,
		TaskList:       taskListName,
		TaskType:       persistence.TransferTaskTypeActivityTask,
		ScheduleID:     event.GetEventID(),
	})

	event = test.AddActivityTaskStartedEvent(mutableState, event.GetEventID(), "")
	ai.StartedID = event.GetEventID()
	event = test.AddActivityTaskCompletedEvent(mutableState, ai.ScheduleID, ai.StartedID, nil, "")
	mutableState.FlushBufferedEvents()

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessDecisionTask_FirstDecision() {

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
		s.domainEntry,
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

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeDecisionTask,
		ScheduleID: di.ScheduleID,
	})

	persistenceMutableState := s.createPersistenceMutableState(mutableState, di.ScheduleID, di.Version)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockMatchingClient.EXPECT().AddDecisionTask(gomock.Any(), s.createAddDecisionTaskRequest(transferTask, mutableState)).Return(nil).Times(1)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessDecisionTask_NonFirstDecision() {

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
		s.domainEntry,
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
	s.NotNil(event)

	// make another round of decision
	taskID := int64(59)
	di = test.AddDecisionTaskScheduledEvent(mutableState)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeDecisionTask,
		ScheduleID: di.ScheduleID,
	})

	persistenceMutableState := s.createPersistenceMutableState(mutableState, di.ScheduleID, di.Version)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockMatchingClient.EXPECT().AddDecisionTask(gomock.Any(), s.createAddDecisionTaskRequest(transferTask, mutableState)).Return(nil).Times(1)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessDecisionTask_Sticky_NonFirstDecision() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"
	stickyTaskListName := "some random sticky task list"
	stickyTaskListTimeout := int32(233)

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		s.domainEntry,
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
	s.NotNil(event)
	// set the sticky tasklist attr
	executionInfo := mutableState.GetExecutionInfo()
	executionInfo.StickyTaskList = stickyTaskListName
	executionInfo.StickyScheduleToStartTimeout = stickyTaskListTimeout

	// make another round of decision
	taskID := int64(59)
	di = test.AddDecisionTaskScheduledEvent(mutableState)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   stickyTaskListName,
		TaskType:   persistence.TransferTaskTypeDecisionTask,
		ScheduleID: di.ScheduleID,
	})

	persistenceMutableState := s.createPersistenceMutableState(mutableState, di.ScheduleID, di.Version)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockMatchingClient.EXPECT().AddDecisionTask(gomock.Any(), s.createAddDecisionTaskRequest(transferTask, mutableState)).Return(nil).Times(1)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessDecisionTask_DecisionNotSticky_MutableStateSticky() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"
	stickyTaskListName := "some random sticky task list"
	stickyTaskListTimeout := int32(233)

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		s.domainEntry,
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
	s.NotNil(event)
	// set the sticky tasklist attr
	executionInfo := mutableState.GetExecutionInfo()
	executionInfo.StickyTaskList = stickyTaskListName
	executionInfo.StickyScheduleToStartTimeout = stickyTaskListTimeout

	// make another round of decision
	taskID := int64(59)
	di = test.AddDecisionTaskScheduledEvent(mutableState)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeDecisionTask,
		ScheduleID: di.ScheduleID,
	})

	persistenceMutableState := s.createPersistenceMutableState(mutableState, di.ScheduleID, di.Version)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockMatchingClient.EXPECT().AddDecisionTask(gomock.Any(), s.createAddDecisionTaskRequest(transferTask, mutableState)).Return(nil).Times(1)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessDecisionTask_Duplication() {

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
		s.domainEntry,
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

	taskID := int64(4096)
	di := test.AddDecisionTaskScheduledEvent(mutableState)
	event := test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, taskListName, uuid.New())
	di.StartedID = event.GetEventID()
	event = test.AddDecisionTaskCompletedEvent(mutableState, di.ScheduleID, di.StartedID, nil, "some random identity")

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeDecisionTask,
		ScheduleID: di.ScheduleID,
	})

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCloseExecution_HasParent() {
	s.testProcessCloseExecution_HasParent(
		s.domainEntry,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
		) {
			s.mockVisibilityMgr.On("RecordWorkflowExecutionClosed", mock.Anything, mock.Anything).Return(nil).Once()
			s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewDisabledArchvialConfig())
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCloseExecution_HasParentCrossCluster() {
	s.testProcessCloseExecution_HasParent(
		s.remoteTargetDomainEntry,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
		) {
			s.mockVisibilityMgr.On("RecordWorkflowExecutionClosed", mock.Anything, mock.Anything).Return(nil).Once()
			s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewDisabledArchvialConfig())
			s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(int64(common.EmptyVersion)).Return(s.mockClusterMetadata.GetCurrentClusterName()).AnyTimes()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.MatchedBy(func(request *persistence.UpdateWorkflowExecutionRequest) bool {
				crossClusterTasks := request.UpdateWorkflowMutation.CrossClusterTasks
				s.Len(crossClusterTasks, 1)
				s.Equal(persistence.CrossClusterTaskTypeRecordChildWorkflowExeuctionComplete, crossClusterTasks[0].GetType())
				return true
			})).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
		},
	)
}

func (s *transferActiveTaskExecutorSuite) testProcessCloseExecution_HasParent(
	domainEntry *cache.DomainCacheEntry,
	setupMockFn func(
		mutableState execution.MutableState,
		workflowExecution, targetExecution types.WorkflowExecution,
	),
) {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	parentDomainID := "some random parent domain ID"
	parentInitiatedID := int64(3222)
	parentDomainName := "some random parent domain Name"
	parentExecution := types.WorkflowExecution{
		WorkflowID: "some random parent workflow ID",
		RunID:      uuid.New(),
	}
	s.mockDomainCache.EXPECT().GetDomainByID(parentDomainID).Return(domainEntry, nil).AnyTimes()

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		s.domainEntry,
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
			ParentExecutionInfo: &types.ParentExecutionInfo{
				DomainUUID:  parentDomainID,
				Domain:      parentDomainName,
				Execution:   &parentExecution,
				InitiatedID: parentInitiatedID,
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

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeCloseExecution,
		ScheduleID: event.GetEventID(),
	})

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockHistoryClient.EXPECT().RecordChildExecutionCompleted(gomock.Any(), &types.RecordChildExecutionCompletedRequest{
		DomainUUID:         parentDomainID,
		WorkflowExecution:  &parentExecution,
		InitiatedID:        parentInitiatedID,
		CompletedExecution: &workflowExecution,
		CompletionEvent:    event,
	}).Return(nil).AnyTimes()
	s.mockParentClosePolicyClient.On("SendParentClosePolicyRequest", mock.Anything, mock.Anything).Return(nil).Times(1)
	setupMockFn(mutableState, workflowExecution, parentExecution)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCloseExecution_NoParent() {

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
		s.domainEntry,
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

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeCloseExecution,
		ScheduleID: event.GetEventID(),
	})

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockVisibilityMgr.On("RecordWorkflowExecutionClosed", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), dc.GetBoolPropertyFn(true), "disabled", "random URI"))
	s.mockArchivalClient.On("Archive", mock.Anything, mock.Anything).Return(nil, nil).Once()

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCloseExecution_NoParent_HasFewChildren() {

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
		s.domainEntry,
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

	dt := types.DecisionTypeStartChildWorkflowExecution
	parentClosePolicy1 := types.ParentClosePolicyAbandon
	parentClosePolicy2 := types.ParentClosePolicyTerminate
	parentClosePolicy3 := types.ParentClosePolicyRequestCancel

	event, _ = mutableState.AddDecisionTaskCompletedEvent(di.ScheduleID, di.StartedID, &types.RespondDecisionTaskCompletedRequest{
		ExecutionContext: nil,
		Identity:         "some random identity",
		Decisions: []*types.Decision{
			{
				DecisionType: &dt,
				StartChildWorkflowExecutionDecisionAttributes: &types.StartChildWorkflowExecutionDecisionAttributes{
					WorkflowID: "child workflow1",
					WorkflowType: &types.WorkflowType{
						Name: "child workflow type",
					},
					TaskList:          &types.TaskList{Name: taskListName},
					Input:             []byte("random input"),
					ParentClosePolicy: &parentClosePolicy1,
				},
			},
			{
				DecisionType: &dt,
				StartChildWorkflowExecutionDecisionAttributes: &types.StartChildWorkflowExecutionDecisionAttributes{
					WorkflowID: "child workflow2",
					WorkflowType: &types.WorkflowType{
						Name: "child workflow type",
					},
					TaskList:          &types.TaskList{Name: taskListName},
					Input:             []byte("random input"),
					ParentClosePolicy: &parentClosePolicy2,
				},
			},
			{
				DecisionType: &dt,
				StartChildWorkflowExecutionDecisionAttributes: &types.StartChildWorkflowExecutionDecisionAttributes{
					WorkflowID: "child workflow3",
					WorkflowType: &types.WorkflowType{
						Name: "child workflow type",
					},
					TaskList:          &types.TaskList{Name: taskListName},
					Input:             []byte("random input"),
					ParentClosePolicy: &parentClosePolicy3,
				},
			},
		},
	}, config.DefaultHistoryMaxAutoResetPoints)

	_, _, err = mutableState.AddStartChildWorkflowExecutionInitiatedEvent(event.GetEventID(), uuid.New(), &types.StartChildWorkflowExecutionDecisionAttributes{
		WorkflowID: "child workflow1",
		WorkflowType: &types.WorkflowType{
			Name: "child workflow type",
		},
		TaskList:          &types.TaskList{Name: taskListName},
		Input:             []byte("random input"),
		ParentClosePolicy: &parentClosePolicy1,
	})
	s.Nil(err)
	_, _, err = mutableState.AddStartChildWorkflowExecutionInitiatedEvent(event.GetEventID(), uuid.New(), &types.StartChildWorkflowExecutionDecisionAttributes{
		WorkflowID: "child workflow2",
		WorkflowType: &types.WorkflowType{
			Name: "child workflow type",
		},
		TaskList:          &types.TaskList{Name: taskListName},
		Input:             []byte("random input"),
		ParentClosePolicy: &parentClosePolicy2,
	})
	s.Nil(err)
	_, _, err = mutableState.AddStartChildWorkflowExecutionInitiatedEvent(event.GetEventID(), uuid.New(), &types.StartChildWorkflowExecutionDecisionAttributes{
		WorkflowID: "child workflow3",
		WorkflowType: &types.WorkflowType{
			Name: "child workflow type",
		},
		TaskList:          &types.TaskList{Name: taskListName},
		Input:             []byte("random input"),
		ParentClosePolicy: &parentClosePolicy3,
	})
	s.Nil(err)

	s.NoError(mutableState.FlushBufferedEvents())

	taskID := int64(59)
	event = test.AddCompleteWorkflowEvent(mutableState, event.GetEventID(), nil)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeCloseExecution,
		ScheduleID: event.GetEventID(),
	})

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockVisibilityMgr.On("RecordWorkflowExecutionClosed", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewDisabledArchvialConfig())
	s.mockHistoryClient.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}) error {
		errors := []error{nil, &types.CancellationAlreadyRequestedError{}, &types.EntityNotExistsError{}}
		return errors[rand.Intn(len(errors))]
	}).Times(1)
	s.mockHistoryClient.EXPECT().TerminateWorkflowExecution(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}) error {
		errors := []error{nil, &types.EntityNotExistsError{}}
		return errors[rand.Intn(len(errors))]
	}).Times(1)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCloseExecution_NoParent_HasManyChildren() {

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
		s.domainEntry,
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

	dt := types.DecisionTypeStartChildWorkflowExecution
	parentClosePolicy := types.ParentClosePolicyTerminate
	var decisions []*types.Decision
	for i := 0; i < 10; i++ {
		decisions = append(decisions, &types.Decision{
			DecisionType: &dt,
			StartChildWorkflowExecutionDecisionAttributes: &types.StartChildWorkflowExecutionDecisionAttributes{
				WorkflowID: "child workflow" + strconv.Itoa(i),
				WorkflowType: &types.WorkflowType{
					Name: "child workflow type",
				},
				TaskList:          &types.TaskList{Name: taskListName},
				Input:             []byte("random input"),
				ParentClosePolicy: &parentClosePolicy,
			},
		})
	}

	event, _ = mutableState.AddDecisionTaskCompletedEvent(di.ScheduleID, di.StartedID, &types.RespondDecisionTaskCompletedRequest{
		ExecutionContext: nil,
		Identity:         "some random identity",
		Decisions:        decisions,
	}, config.DefaultHistoryMaxAutoResetPoints)

	for i := 0; i < 10; i++ {
		_, _, err = mutableState.AddStartChildWorkflowExecutionInitiatedEvent(event.GetEventID(), uuid.New(), &types.StartChildWorkflowExecutionDecisionAttributes{
			WorkflowID: "child workflow" + strconv.Itoa(i),
			WorkflowType: &types.WorkflowType{
				Name: "child workflow type",
			},
			TaskList:          &types.TaskList{Name: taskListName},
			Input:             []byte("random input"),
			ParentClosePolicy: &parentClosePolicy,
		})
		s.Nil(err)
	}

	s.NoError(mutableState.FlushBufferedEvents())

	taskID := int64(59)
	event = test.AddCompleteWorkflowEvent(mutableState, event.GetEventID(), nil)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeCloseExecution,
		ScheduleID: event.GetEventID(),
	})

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockVisibilityMgr.On("RecordWorkflowExecutionClosed", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewDisabledArchvialConfig())
	s.mockParentClosePolicyClient.On("SendParentClosePolicyRequest", mock.Anything, mock.Anything).Return(nil).Times(1)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCloseExecution_NoParent_HasManyAbandonedChildren() {

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
		s.domainEntry,
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

	dt := types.DecisionTypeStartChildWorkflowExecution
	parentClosePolicy := types.ParentClosePolicyAbandon
	var decisions []*types.Decision
	for i := 0; i < 10; i++ {
		decisions = append(decisions, &types.Decision{
			DecisionType: &dt,
			StartChildWorkflowExecutionDecisionAttributes: &types.StartChildWorkflowExecutionDecisionAttributes{
				WorkflowID: "child workflow" + strconv.Itoa(i),
				WorkflowType: &types.WorkflowType{
					Name: "child workflow type",
				},
				TaskList:          &types.TaskList{Name: taskListName},
				Input:             []byte("random input"),
				ParentClosePolicy: &parentClosePolicy,
			},
		})
	}

	event, _ = mutableState.AddDecisionTaskCompletedEvent(di.ScheduleID, di.StartedID, &types.RespondDecisionTaskCompletedRequest{
		ExecutionContext: nil,
		Identity:         "some random identity",
		Decisions:        decisions,
	}, config.DefaultHistoryMaxAutoResetPoints)

	for i := 0; i < 10; i++ {
		_, _, err = mutableState.AddStartChildWorkflowExecutionInitiatedEvent(event.GetEventID(), uuid.New(), &types.StartChildWorkflowExecutionDecisionAttributes{
			WorkflowID: "child workflow" + strconv.Itoa(i),
			WorkflowType: &types.WorkflowType{
				Name: "child workflow type",
			},
			TaskList:          &types.TaskList{Name: taskListName},
			Input:             []byte("random input"),
			ParentClosePolicy: &parentClosePolicy,
		})
		s.Nil(err)
	}

	s.NoError(mutableState.FlushBufferedEvents())

	taskID := int64(59)
	event = test.AddCompleteWorkflowEvent(mutableState, event.GetEventID(), nil)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeCloseExecution,
		ScheduleID: event.GetEventID(),
	})

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockVisibilityMgr.On("RecordWorkflowExecutionClosed", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewDisabledArchvialConfig())

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCancelExecution_Success() {
	s.testProcessCancelExecution(
		s.targetDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			event *types.HistoryEvent,
			transferTask Task,
			requestCancelInfo *persistence.RequestCancelInfo,
		) {
			persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			cancelRequest := createTestRequestCancelWorkflowExecutionRequest(s.targetDomainName, transferTask.GetInfo().(*persistence.TransferTaskInfo), requestCancelInfo.CancelRequestID)
			s.mockHistoryClient.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), cancelRequest).Return(nil).Times(1)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
			s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(s.version).Return(cluster.TestCurrentClusterName).AnyTimes()
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCancelExecution_Failure() {
	s.testProcessCancelExecution(
		s.targetDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			event *types.HistoryEvent,
			transferTask Task,
			requestCancelInfo *persistence.RequestCancelInfo,
		) {
			persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			cancelRequest := createTestRequestCancelWorkflowExecutionRequest(s.targetDomainName, transferTask.GetInfo().(*persistence.TransferTaskInfo), requestCancelInfo.CancelRequestID)
			s.mockHistoryClient.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), cancelRequest).Return(&types.EntityNotExistsError{}).Times(1)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
			s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(s.version).Return(cluster.TestCurrentClusterName).AnyTimes()
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCancelExecution_Duplication() {
	s.testProcessCancelExecution(
		s.targetDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			event *types.HistoryEvent,
			transferTask Task,
			requestCancelInfo *persistence.RequestCancelInfo,
		) {
			event = test.AddCancelRequestedEvent(mutableState, event.GetEventID(), s.targetDomainID, targetExecution.GetWorkflowID(), targetExecution.GetRunID())
			mutableState.FlushBufferedEvents()

			persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCancelExecution_CrossCluster() {
	s.testProcessCancelExecution(
		s.remoteTargetDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			event *types.HistoryEvent,
			transferTask Task,
			requestCancelInfo *persistence.RequestCancelInfo,
		) {
			persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.MatchedBy(func(request *persistence.UpdateWorkflowExecutionRequest) bool {
				crossClusterTasks := request.UpdateWorkflowMutation.CrossClusterTasks
				s.Len(crossClusterTasks, 1)
				s.Equal(persistence.CrossClusterTaskTypeCancelExecution, crossClusterTasks[0].GetType())
				return true
			})).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
		},
	)
}

func (s *transferActiveTaskExecutorSuite) testProcessCancelExecution(
	targetDomainID string,
	setupMockFn func(
		mutableState execution.MutableState,
		workflowExecution, targetExecution types.WorkflowExecution,
		cancelInitEvent *types.HistoryEvent,
		transferTask Task,
		requestCancelInfo *persistence.RequestCancelInfo,
	),
) {
	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflow(s.mockShard, s.domainID)
	s.NoError(err)
	targetExecution := types.WorkflowExecution{
		WorkflowID: "some random target workflow ID",
		RunID:      uuid.New(),
	}

	event, rci := test.AddRequestCancelInitiatedEvent(
		mutableState,
		decisionCompletionID,
		uuid.New(),
		s.targetDomainName,
		targetExecution.GetWorkflowID(),
		targetExecution.GetRunID(),
	)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:          s.version,
		DomainID:         s.domainID,
		WorkflowID:       workflowExecution.GetWorkflowID(),
		RunID:            workflowExecution.GetRunID(),
		TargetDomainID:   targetDomainID,
		TargetWorkflowID: targetExecution.GetWorkflowID(),
		TargetRunID:      targetExecution.GetRunID(),
		TaskID:           int64(59),
		TaskList:         mutableState.GetExecutionInfo().TaskList,
		TaskType:         persistence.TransferTaskTypeCancelExecution,
		ScheduleID:       event.GetEventID(),
	})

	setupMockFn(mutableState, workflowExecution, targetExecution, event, transferTask, rci)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessSignalExecution_Success() {
	s.testProcessSignalExecution(
		s.targetDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			event *types.HistoryEvent,
			transferTask Task,
			signalInfo *persistence.SignalInfo,
		) {
			persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			signalRequest := createTestSignalWorkflowExecutionRequest(s.targetDomainName, transferTask.GetInfo().(*persistence.TransferTaskInfo), signalInfo)
			s.mockHistoryClient.EXPECT().SignalWorkflowExecution(gomock.Any(), signalRequest).Return(nil).Times(1)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()

			taskInfo := transferTask.GetInfo().(*persistence.TransferTaskInfo)
			s.mockHistoryClient.EXPECT().RemoveSignalMutableState(gomock.Any(), &types.RemoveSignalMutableStateRequest{
				DomainUUID: taskInfo.TargetDomainID,
				WorkflowExecution: &types.WorkflowExecution{
					WorkflowID: taskInfo.TargetWorkflowID,
					RunID:      taskInfo.TargetRunID,
				},
				RequestID: signalInfo.SignalRequestID,
			}).Return(nil).Times(1)
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestProcessSignalExecution_Failure() {
	s.testProcessSignalExecution(
		s.targetDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			event *types.HistoryEvent,
			transferTask Task,
			signalInfo *persistence.SignalInfo,
		) {
			persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			signalRequest := createTestSignalWorkflowExecutionRequest(s.targetDomainName, transferTask.GetInfo().(*persistence.TransferTaskInfo), signalInfo)
			s.mockHistoryClient.EXPECT().SignalWorkflowExecution(gomock.Any(), signalRequest).Return(&types.EntityNotExistsError{}).Times(1)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestProcessSignalExecution_Duplication() {
	s.testProcessSignalExecution(
		s.targetDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			event *types.HistoryEvent,
			transferTask Task,
			signalInfo *persistence.SignalInfo,
		) {
			event = test.AddSignaledEvent(mutableState, event.GetEventID(), s.targetDomainName, targetExecution.GetWorkflowID(), targetExecution.GetRunID(), nil)
			mutableState.FlushBufferedEvents()

			persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestProcessSignalExecution_CrossCluster() {
	s.testProcessSignalExecution(
		s.remoteTargetDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
			event *types.HistoryEvent,
			transferTask Task,
			signalInfo *persistence.SignalInfo,
		) {
			persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.MatchedBy(func(request *persistence.UpdateWorkflowExecutionRequest) bool {
				crossClusterTasks := request.UpdateWorkflowMutation.CrossClusterTasks
				s.Len(crossClusterTasks, 1)
				s.Equal(persistence.CrossClusterTaskTypeSignalExecution, crossClusterTasks[0].GetType())
				return true
			})).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
		},
	)
}

func (s *transferActiveTaskExecutorSuite) testProcessSignalExecution(
	targetDomainID string,
	setupMockFn func(
		mutableState execution.MutableState,
		workflowExecution, targetExecution types.WorkflowExecution,
		signalInitEvent *types.HistoryEvent,
		transferTask Task,
		signalInfo *persistence.SignalInfo,
	),
) {
	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflow(s.mockShard, s.domainID)
	s.NoError(err)
	targetExecution := types.WorkflowExecution{
		WorkflowID: "some random target workflow ID",
		RunID:      uuid.New(),
	}

	event, signalInfo := test.AddRequestSignalInitiatedEvent(
		mutableState,
		decisionCompletionID,
		uuid.New(),
		s.targetDomainName,
		targetExecution.GetWorkflowID(),
		targetExecution.GetRunID(),
		"some random signal name",
		[]byte("some random signal input"),
		[]byte("some random signal control"),
	)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:          s.version,
		DomainID:         s.domainID,
		WorkflowID:       workflowExecution.GetWorkflowID(),
		RunID:            workflowExecution.GetRunID(),
		TargetDomainID:   targetDomainID,
		TargetWorkflowID: targetExecution.GetWorkflowID(),
		TargetRunID:      targetExecution.GetRunID(),
		TaskID:           int64(59),
		TaskList:         mutableState.GetExecutionInfo().TaskList,
		TaskType:         persistence.TransferTaskTypeSignalExecution,
		ScheduleID:       event.GetEventID(),
	})

	setupMockFn(mutableState, workflowExecution, targetExecution, event, transferTask, signalInfo)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessStartChildExecution_Success() {
	s.testProcessStartChildExecution(
		s.childDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, childExecution types.WorkflowExecution,
			event *types.HistoryEvent,
			transferTask Task,
			childInfo *persistence.ChildExecutionInfo,
		) {
			persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			taskInfo := transferTask.GetInfo().(*persistence.TransferTaskInfo)
			event, err := mutableState.GetChildExecutionInitiatedEvent(context.Background(), taskInfo.ScheduleID)
			s.NoError(err)
			historyReq := createTestChildWorkflowExecutionRequest(
				s.domainName,
				s.childDomainName,
				taskInfo,
				event.StartChildWorkflowExecutionInitiatedEventAttributes,
				childInfo.CreateRequestID,
				s.mockShard.GetTimeSource().Now(),
			)
			s.mockHistoryClient.EXPECT().StartWorkflowExecution(gomock.Any(), historyReq).Return(&types.StartWorkflowExecutionResponse{RunID: childExecution.RunID}, nil).Times(1)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
			s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(s.version).Return(cluster.TestCurrentClusterName).AnyTimes()
			s.mockHistoryClient.EXPECT().ScheduleDecisionTask(gomock.Any(), &types.ScheduleDecisionTaskRequest{
				DomainUUID: s.childDomainID,
				WorkflowExecution: &types.WorkflowExecution{
					WorkflowID: childExecution.WorkflowID,
					RunID:      childExecution.RunID,
				},
				IsFirstDecision: true,
			}).Return(nil).Times(1)
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestProcessStartChildExecution_Failure() {
	s.testProcessStartChildExecution(
		s.childDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, childExecution types.WorkflowExecution,
			event *types.HistoryEvent,
			transferTask Task,
			childInfo *persistence.ChildExecutionInfo,
		) {
			persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			taskInfo := transferTask.GetInfo().(*persistence.TransferTaskInfo)
			event, err := mutableState.GetChildExecutionInitiatedEvent(context.Background(), taskInfo.ScheduleID)
			s.NoError(err)
			historyReq := createTestChildWorkflowExecutionRequest(
				s.domainName,
				s.childDomainName,
				taskInfo,
				event.StartChildWorkflowExecutionInitiatedEventAttributes,
				childInfo.CreateRequestID,
				s.mockShard.GetTimeSource().Now(),
			)
			s.mockHistoryClient.EXPECT().StartWorkflowExecution(gomock.Any(), historyReq).Return(nil, &types.WorkflowExecutionAlreadyStartedError{}).Times(1)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
			s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(s.version).Return(cluster.TestCurrentClusterName).AnyTimes()
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestProcessStartChildExecution_Success_Dup() {
	s.testProcessStartChildExecution(
		s.childDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, childExecution types.WorkflowExecution,
			event *types.HistoryEvent,
			transferTask Task,
			childInfo *persistence.ChildExecutionInfo,
		) {
			startEvent := test.AddChildWorkflowExecutionStartedEvent(mutableState, event.GetEventID(), s.childDomainID, childExecution.WorkflowID, childExecution.RunID, childInfo.WorkflowTypeName)
			childInfo.StartedID = startEvent.GetEventID()
			mutableState.FlushBufferedEvents()

			persistenceMutableState := s.createPersistenceMutableState(mutableState, startEvent.GetEventID(), startEvent.GetVersion())
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			s.mockHistoryClient.EXPECT().ScheduleDecisionTask(gomock.Any(), &types.ScheduleDecisionTaskRequest{
				DomainUUID: s.childDomainID,
				WorkflowExecution: &types.WorkflowExecution{
					WorkflowID: childExecution.WorkflowID,
					RunID:      childExecution.RunID,
				},
				IsFirstDecision: true,
			}).Return(nil).Times(1)
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestProcessStartChildExecution_Duplication() {
	s.testProcessStartChildExecution(
		s.childDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, childExecution types.WorkflowExecution,
			event *types.HistoryEvent,
			transferTask Task,
			childInfo *persistence.ChildExecutionInfo,
		) {
			startEvent := test.AddChildWorkflowExecutionStartedEvent(mutableState, event.GetEventID(), s.childDomainID, childExecution.GetWorkflowID(), childExecution.GetRunID(), childInfo.WorkflowTypeName)
			childInfo.StartedID = startEvent.GetEventID()
			startEvent = test.AddChildWorkflowExecutionCompletedEvent(mutableState, childInfo.InitiatedID, &childExecution, &types.WorkflowExecutionCompletedEventAttributes{
				Result:                       []byte("some random child workflow execution result"),
				DecisionTaskCompletedEventID: transferTask.GetInfo().(*persistence.TransferTaskInfo).ScheduleID,
			})
			mutableState.FlushBufferedEvents()

			persistenceMutableState := s.createPersistenceMutableState(mutableState, startEvent.GetEventID(), startEvent.GetVersion())
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestProcessStartChildExecution_CrossCluster() {
	s.testProcessStartChildExecution(
		s.remoteTargetDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, childExecution types.WorkflowExecution,
			event *types.HistoryEvent,
			transferTask Task,
			childInfo *persistence.ChildExecutionInfo,
		) {
			persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.MatchedBy(func(request *persistence.UpdateWorkflowExecutionRequest) bool {
				crossClusterTasks := request.UpdateWorkflowMutation.CrossClusterTasks
				s.Len(crossClusterTasks, 1)
				s.Equal(persistence.CrossClusterTaskTypeStartChildExecution, crossClusterTasks[0].GetType())
				return true
			})).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
		},
	)
}

func (s *transferActiveTaskExecutorSuite) testProcessStartChildExecution(
	targetDomainID string,
	setupMockFn func(
		mutableState execution.MutableState,
		workflowExecution, targetExecution types.WorkflowExecution,
		childInitEvent *types.HistoryEvent,
		transferTask Task,
		childInfo *persistence.ChildExecutionInfo,
	),
) {
	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflow(s.mockShard, s.domainID)
	s.NoError(err)
	childExecution := types.WorkflowExecution{
		WorkflowID: "some random child workflow ID",
		RunID:      uuid.New(),
	}

	event, ci := test.AddStartChildWorkflowExecutionInitiatedEvent(
		mutableState,
		decisionCompletionID,
		uuid.New(),
		s.childDomainName,
		childExecution.WorkflowID,
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

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:          s.version,
		DomainID:         s.domainID,
		WorkflowID:       workflowExecution.GetWorkflowID(),
		RunID:            workflowExecution.GetRunID(),
		TargetDomainID:   targetDomainID,
		TargetWorkflowID: childExecution.WorkflowID,
		TargetRunID:      "",
		TaskID:           int64(59),
		TaskList:         mutableState.GetExecutionInfo().TaskList,
		TaskType:         persistence.TransferTaskTypeStartChildExecution,
		ScheduleID:       event.GetEventID(),
	})

	setupMockFn(mutableState, workflowExecution, childExecution, event, transferTask, ci)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessRecordWorkflowStartedTask() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"
	cronSchedule := "@every 5s"
	backoffSeconds := int32(5)

	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		s.mockShard,
		s.logger,
		s.version,
		workflowExecution.GetRunID(),
		s.domainEntry,
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
				CronSchedule:                        cronSchedule,
			},
			FirstDecisionTaskBackoffSeconds: common.Int32Ptr(backoffSeconds),
		},
	)
	s.Nil(err)

	taskID := int64(59)
	di := test.AddDecisionTaskScheduledEvent(mutableState)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeRecordWorkflowStarted,
		ScheduleID: event.GetEventID(),
	})

	persistenceMutableState := s.createPersistenceMutableState(mutableState, di.ScheduleID, di.Version)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockVisibilityMgr.On("RecordWorkflowExecutionStarted", mock.Anything, s.createRecordWorkflowExecutionStartedRequest(s.domainName, event, transferTask, mutableState, backoffSeconds)).Once().Return(nil)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessUpsertWorkflowSearchAttributes() {

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
		s.domainEntry,
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

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeUpsertWorkflowSearchAttributes,
		ScheduleID: event.GetEventID(),
	})

	persistenceMutableState := s.createPersistenceMutableState(mutableState, di.ScheduleID, di.Version)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockVisibilityMgr.On("UpsertWorkflowExecution", mock.Anything, s.createUpsertWorkflowSearchAttributesRequest(s.domainName, event, transferTask, mutableState)).Once().Return(nil)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestCopySearchAttributes() {
	var input map[string][]byte
	s.Nil(copySearchAttributes(input))

	key := "key"
	val := []byte{'1', '2', '3'}
	input = map[string][]byte{
		key: val,
	}
	result := copySearchAttributes(input)
	s.Equal(input, result)
	result[key][0] = '0'
	s.Equal(byte('1'), val[0])
}

func (s *transferActiveTaskExecutorSuite) createAddActivityTaskRequest(
	transferTask Task,
	ai *persistence.ActivityInfo,
) *types.AddActivityTaskRequest {

	taskInfo := transferTask.GetInfo().(*persistence.TransferTaskInfo)
	workflowExecution := types.WorkflowExecution{
		WorkflowID: taskInfo.WorkflowID,
		RunID:      taskInfo.RunID,
	}
	taskList := &types.TaskList{Name: taskInfo.TaskList}

	return &types.AddActivityTaskRequest{
		DomainUUID:                    taskInfo.TargetDomainID,
		SourceDomainUUID:              taskInfo.DomainID,
		Execution:                     &workflowExecution,
		TaskList:                      taskList,
		ScheduleID:                    taskInfo.ScheduleID,
		ScheduleToStartTimeoutSeconds: common.Int32Ptr(ai.ScheduleToStartTimeout),
	}
}

func (s *transferActiveTaskExecutorSuite) createAddDecisionTaskRequest(
	transferTask Task,
	mutableState execution.MutableState,
) *types.AddDecisionTaskRequest {

	taskInfo := transferTask.GetInfo().(*persistence.TransferTaskInfo)
	workflowExecution := types.WorkflowExecution{
		WorkflowID: taskInfo.WorkflowID,
		RunID:      taskInfo.RunID,
	}
	taskList := &types.TaskList{Name: taskInfo.TaskList}
	executionInfo := mutableState.GetExecutionInfo()
	timeout := executionInfo.WorkflowTimeout
	if mutableState.GetExecutionInfo().TaskList != taskInfo.TaskList {
		taskListStickyKind := types.TaskListKindSticky
		taskList.Kind = &taskListStickyKind
		timeout = executionInfo.StickyScheduleToStartTimeout
	}

	return &types.AddDecisionTaskRequest{
		DomainUUID:                    taskInfo.DomainID,
		Execution:                     &workflowExecution,
		TaskList:                      taskList,
		ScheduleID:                    taskInfo.ScheduleID,
		ScheduleToStartTimeoutSeconds: common.Int32Ptr(timeout),
	}
}

func (s *transferActiveTaskExecutorSuite) createRecordWorkflowExecutionStartedRequest(
	domainName string,
	startEvent *types.HistoryEvent,
	transferTask Task,
	mutableState execution.MutableState,
	backoffSeconds int32,
) *persistence.RecordWorkflowExecutionStartedRequest {
	taskInfo := transferTask.GetInfo().(*persistence.TransferTaskInfo)
	workflowExecution := types.WorkflowExecution{
		WorkflowID: taskInfo.WorkflowID,
		RunID:      taskInfo.RunID,
	}
	executionInfo := mutableState.GetExecutionInfo()
	executionTimestamp := time.Unix(0, startEvent.GetTimestamp()).Add(time.Duration(backoffSeconds) * time.Second)
	isCron := len(executionInfo.CronSchedule) > 0

	return &persistence.RecordWorkflowExecutionStartedRequest{
		Domain:             domainName,
		DomainUUID:         taskInfo.DomainID,
		Execution:          workflowExecution,
		WorkflowTypeName:   executionInfo.WorkflowTypeName,
		StartTimestamp:     startEvent.GetTimestamp(),
		ExecutionTimestamp: executionTimestamp.UnixNano(),
		WorkflowTimeout:    int64(executionInfo.WorkflowTimeout),
		TaskID:             taskInfo.TaskID,
		TaskList:           taskInfo.TaskList,
		IsCron:             isCron,
	}
}

func createTestRequestCancelWorkflowExecutionRequest(
	targetDomainName string,
	taskInfo *persistence.TransferTaskInfo,
	requestID string,
) *types.HistoryRequestCancelWorkflowExecutionRequest {

	sourceExecution := types.WorkflowExecution{
		WorkflowID: taskInfo.WorkflowID,
		RunID:      taskInfo.RunID,
	}
	targetExecution := types.WorkflowExecution{
		WorkflowID: taskInfo.TargetWorkflowID,
		RunID:      taskInfo.TargetRunID,
	}

	return &types.HistoryRequestCancelWorkflowExecutionRequest{
		DomainUUID: taskInfo.TargetDomainID,
		CancelRequest: &types.RequestCancelWorkflowExecutionRequest{
			Domain:            targetDomainName,
			WorkflowExecution: &targetExecution,
			Identity:          execution.IdentityHistoryService,
			// Use the same request ID to dedupe RequestCancelWorkflowExecution calls
			RequestID: requestID,
		},
		ExternalInitiatedEventID:  common.Int64Ptr(taskInfo.ScheduleID),
		ExternalWorkflowExecution: &sourceExecution,
		ChildWorkflowOnly:         taskInfo.TargetChildWorkflowOnly,
	}
}

func createTestSignalWorkflowExecutionRequest(
	targetDomainName string,
	taskInfo *persistence.TransferTaskInfo,
	si *persistence.SignalInfo,
) *types.HistorySignalWorkflowExecutionRequest {

	sourceExecution := types.WorkflowExecution{
		WorkflowID: taskInfo.WorkflowID,
		RunID:      taskInfo.RunID,
	}
	targetExecution := types.WorkflowExecution{
		WorkflowID: taskInfo.TargetWorkflowID,
		RunID:      taskInfo.TargetRunID,
	}

	return &types.HistorySignalWorkflowExecutionRequest{
		DomainUUID: taskInfo.TargetDomainID,
		SignalRequest: &types.SignalWorkflowExecutionRequest{
			Domain:            targetDomainName,
			WorkflowExecution: &targetExecution,
			Identity:          execution.IdentityHistoryService,
			SignalName:        si.SignalName,
			Input:             si.Input,
			RequestID:         si.SignalRequestID,
			Control:           si.Control,
		},
		ExternalWorkflowExecution: &sourceExecution,
		ChildWorkflowOnly:         taskInfo.TargetChildWorkflowOnly,
	}
}

func createTestChildWorkflowExecutionRequest(
	domainName string,
	childDomainName string,
	taskInfo *persistence.TransferTaskInfo,
	attributes *types.StartChildWorkflowExecutionInitiatedEventAttributes,
	requestID string,
	now time.Time,
) *types.HistoryStartWorkflowExecutionRequest {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: taskInfo.WorkflowID,
		RunID:      taskInfo.RunID,
	}
	frontendStartReq := &types.StartWorkflowExecutionRequest{
		Domain:                              childDomainName,
		WorkflowID:                          attributes.WorkflowID,
		WorkflowType:                        attributes.WorkflowType,
		TaskList:                            attributes.TaskList,
		Input:                               attributes.Input,
		ExecutionStartToCloseTimeoutSeconds: attributes.ExecutionStartToCloseTimeoutSeconds,
		TaskStartToCloseTimeoutSeconds:      attributes.TaskStartToCloseTimeoutSeconds,
		// Use the same request ID to dedupe StartWorkflowExecution calls
		RequestID:             requestID,
		WorkflowIDReusePolicy: attributes.WorkflowIDReusePolicy,
		RetryPolicy:           attributes.RetryPolicy,
	}

	parentInfo := &types.ParentExecutionInfo{
		DomainUUID:  taskInfo.DomainID,
		Domain:      domainName,
		Execution:   &workflowExecution,
		InitiatedID: taskInfo.ScheduleID,
	}

	historyStartReq := common.CreateHistoryStartWorkflowRequest(
		taskInfo.TargetDomainID, frontendStartReq, now)

	historyStartReq.ParentExecutionInfo = parentInfo
	return historyStartReq
}

func (s *transferActiveTaskExecutorSuite) createUpsertWorkflowSearchAttributesRequest(
	domainName string,
	startEvent *types.HistoryEvent,
	transferTask Task,
	mutableState execution.MutableState,
) *persistence.UpsertWorkflowExecutionRequest {

	taskInfo := transferTask.GetInfo().(*persistence.TransferTaskInfo)
	workflowExecution := types.WorkflowExecution{
		WorkflowID: taskInfo.WorkflowID,
		RunID:      taskInfo.RunID,
	}
	executionInfo := mutableState.GetExecutionInfo()

	return &persistence.UpsertWorkflowExecutionRequest{
		Domain:           domainName,
		DomainUUID:       taskInfo.DomainID,
		Execution:        workflowExecution,
		WorkflowTypeName: executionInfo.WorkflowTypeName,
		StartTimestamp:   startEvent.GetTimestamp(),
		WorkflowTimeout:  int64(executionInfo.WorkflowTimeout),
		TaskID:           taskInfo.TaskID,
		TaskList:         taskInfo.TaskList,
	}
}

func (s *transferActiveTaskExecutorSuite) createPersistenceMutableState(
	ms execution.MutableState,
	lastEventID int64,
	lastEventVersion int64,
) *persistence.WorkflowMutableState {

	if ms.GetVersionHistories() != nil {
		currentVersionHistory, err := ms.GetVersionHistories().GetCurrentVersionHistory()
		s.NoError(err)
		err = currentVersionHistory.AddOrUpdateItem(persistence.NewVersionHistoryItem(
			lastEventID,
			lastEventVersion,
		))
		s.NoError(err)
	}

	return execution.CreatePersistenceMutableState(ms)
}

func (s *transferActiveTaskExecutorSuite) newTransferTaskFromInfo(
	info *persistence.TransferTaskInfo,
) Task {
	return NewTransferTask(s.mockShard, info, QueueTypeActiveTransfer, s.logger, nil, nil, nil, nil, nil)
}
