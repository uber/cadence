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
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestDomainID).Return(constants.TestGlobalDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestDomainID).Return(constants.TestDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomain(constants.TestDomainName).Return(constants.TestGlobalDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestTargetDomainID).Return(constants.TestGlobalTargetDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestTargetDomainID).Return(constants.TestTargetDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomain(constants.TestTargetDomainName).Return(constants.TestGlobalTargetDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestParentDomainID).Return(constants.TestGlobalParentDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestParentDomainID).Return(constants.TestParentDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomain(constants.TestParentDomainName).Return(constants.TestGlobalParentDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestChildDomainID).Return(constants.TestGlobalChildDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestChildDomainID).Return(constants.TestChildDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomain(constants.TestChildDomainName).Return(constants.TestGlobalChildDomainEntry, nil).AnyTimes()
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
		s.mockShard.GetMetricsClient(),
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
	event, ai := test.AddActivityTaskScheduledEvent(mutableState, event.GetEventID(), activityID, activityType, taskListName, []byte{}, 1, 1, 1, 1)
	mutableState.FlushBufferedEvents()
	transferTask := &persistence.TransferTaskInfo{
		Version:        s.version,
		DomainID:       s.domainID,
		TargetDomainID: constants.TestTargetDomainID,
		WorkflowID:     workflowExecution.GetWorkflowID(),
		RunID:          workflowExecution.GetRunID(),
		TaskID:         taskID,
		TaskList:       taskListName,
		TaskType:       persistence.TransferTaskTypeActivityTask,
		ScheduleID:     event.GetEventID(),
	}

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
	event, ai := test.AddActivityTaskScheduledEvent(mutableState, event.GetEventID(), activityID, activityType, taskListName, []byte{}, 1, 1, 1, 1)

	transferTask := &persistence.TransferTaskInfo{
		Version:        s.version,
		DomainID:       s.domainID,
		TargetDomainID: s.targetDomainID,
		WorkflowID:     workflowExecution.GetWorkflowID(),
		RunID:          workflowExecution.GetRunID(),
		TaskID:         taskID,
		TaskList:       taskListName,
		TaskType:       persistence.TransferTaskTypeActivityTask,
		ScheduleID:     event.GetEventID(),
	}

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

	transferTask := &persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeDecisionTask,
		ScheduleID: di.ScheduleID,
	}

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
	s.NotNil(event)

	// make another round of decision
	taskID := int64(59)
	di = test.AddDecisionTaskScheduledEvent(mutableState)

	transferTask := &persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeDecisionTask,
		ScheduleID: di.ScheduleID,
	}

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
	s.NotNil(event)
	// set the sticky tasklist attr
	executionInfo := mutableState.GetExecutionInfo()
	executionInfo.StickyTaskList = stickyTaskListName
	executionInfo.StickyScheduleToStartTimeout = stickyTaskListTimeout

	// make another round of decision
	taskID := int64(59)
	di = test.AddDecisionTaskScheduledEvent(mutableState)

	transferTask := &persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   stickyTaskListName,
		TaskType:   persistence.TransferTaskTypeDecisionTask,
		ScheduleID: di.ScheduleID,
	}

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
	s.NotNil(event)
	// set the sticky tasklist attr
	executionInfo := mutableState.GetExecutionInfo()
	executionInfo.StickyTaskList = stickyTaskListName
	executionInfo.StickyScheduleToStartTimeout = stickyTaskListTimeout

	// make another round of decision
	taskID := int64(59)
	di = test.AddDecisionTaskScheduledEvent(mutableState)

	transferTask := &persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeDecisionTask,
		ScheduleID: di.ScheduleID,
	}

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

	taskID := int64(4096)
	di := test.AddDecisionTaskScheduledEvent(mutableState)
	event := test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, taskListName, uuid.New())
	di.StartedID = event.GetEventID()
	event = test.AddDecisionTaskCompletedEvent(mutableState, di.ScheduleID, di.StartedID, nil, "some random identity")

	transferTask := &persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeDecisionTask,
		ScheduleID: di.ScheduleID,
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCloseExecution_HasParent() {

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

	transferTask := &persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeCloseExecution,
		ScheduleID: event.GetEventID(),
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockHistoryClient.EXPECT().RecordChildExecutionCompleted(gomock.Any(), &types.RecordChildExecutionCompletedRequest{
		DomainUUID:         parentDomainID,
		WorkflowExecution:  &parentExecution,
		InitiatedID:        parentInitiatedID,
		CompletedExecution: &workflowExecution,
		CompletionEvent:    event,
	}).Return(nil).AnyTimes()
	s.mockVisibilityMgr.On("RecordWorkflowExecutionClosed", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewDisabledArchvialConfig())

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

	transferTask := &persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeCloseExecution,
		ScheduleID: event.GetEventID(),
	}

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

	transferTask := &persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeCloseExecution,
		ScheduleID: event.GetEventID(),
	}

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

	transferTask := &persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeCloseExecution,
		ScheduleID: event.GetEventID(),
	}

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

	transferTask := &persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeCloseExecution,
		ScheduleID: event.GetEventID(),
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockVisibilityMgr.On("RecordWorkflowExecutionClosed", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewDisabledArchvialConfig())

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCancelExecution_Success() {

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
	event, rci := test.AddRequestCancelInitiatedEvent(mutableState, event.GetEventID(), uuid.New(), constants.TestTargetDomainName, targetExecution.GetWorkflowID(), targetExecution.GetRunID())

	transferTask := &persistence.TransferTaskInfo{
		Version:          s.version,
		DomainID:         s.domainID,
		WorkflowID:       workflowExecution.GetWorkflowID(),
		RunID:            workflowExecution.GetRunID(),
		TargetDomainID:   s.targetDomainID,
		TargetWorkflowID: targetExecution.GetWorkflowID(),
		TargetRunID:      targetExecution.GetRunID(),
		TaskID:           taskID,
		TaskList:         taskListName,
		TaskType:         persistence.TransferTaskTypeCancelExecution,
		ScheduleID:       event.GetEventID(),
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockHistoryClient.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), s.createRequestCancelWorkflowExecutionRequest(s.targetDomainName, transferTask, rci)).Return(nil).Times(1)
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
	s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(s.version).Return(cluster.TestCurrentClusterName).AnyTimes()

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCancelExecution_Failure() {

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
	event, rci := test.AddRequestCancelInitiatedEvent(mutableState, event.GetEventID(), uuid.New(), constants.TestTargetDomainName, targetExecution.GetWorkflowID(), targetExecution.GetRunID())
	mutableState.FlushBufferedEvents()

	transferTask := &persistence.TransferTaskInfo{
		Version:          s.version,
		DomainID:         s.domainID,
		WorkflowID:       workflowExecution.GetWorkflowID(),
		RunID:            workflowExecution.GetRunID(),
		TargetDomainID:   s.targetDomainID,
		TargetWorkflowID: targetExecution.GetWorkflowID(),
		TargetRunID:      targetExecution.GetRunID(),
		TaskID:           taskID,
		TaskList:         taskListName,
		TaskType:         persistence.TransferTaskTypeCancelExecution,
		ScheduleID:       event.GetEventID(),
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockHistoryClient.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), s.createRequestCancelWorkflowExecutionRequest(s.targetDomainName, transferTask, rci)).Return(&types.EntityNotExistsError{}).Times(1)
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
	s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(s.version).Return(cluster.TestCurrentClusterName).AnyTimes()

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCancelExecution_Duplication() {

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

	transferTask := &persistence.TransferTaskInfo{
		Version:          s.version,
		DomainID:         s.domainID,
		WorkflowID:       workflowExecution.GetWorkflowID(),
		RunID:            workflowExecution.GetRunID(),
		TargetDomainID:   s.targetDomainID,
		TargetWorkflowID: targetExecution.GetWorkflowID(),
		TargetRunID:      targetExecution.GetRunID(),
		TaskID:           taskID,
		TaskList:         taskListName,
		TaskType:         persistence.TransferTaskTypeCancelExecution,
		ScheduleID:       event.GetEventID(),
	}

	event = test.AddCancelRequestedEvent(mutableState, event.GetEventID(), constants.TestTargetDomainID, targetExecution.GetWorkflowID(), targetExecution.GetRunID())
	mutableState.FlushBufferedEvents()

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessSignalExecution_Success() {

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
	signalInput := []byte("some random signal input")
	signalControl := []byte("some random signal control")

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
	event, si := test.AddRequestSignalInitiatedEvent(mutableState, event.GetEventID(), uuid.New(),
		constants.TestTargetDomainName, targetExecution.GetWorkflowID(), targetExecution.GetRunID(), signalName, signalInput, signalControl)

	transferTask := &persistence.TransferTaskInfo{
		Version:          s.version,
		DomainID:         s.domainID,
		WorkflowID:       workflowExecution.GetWorkflowID(),
		RunID:            workflowExecution.GetRunID(),
		TargetDomainID:   s.targetDomainID,
		TargetWorkflowID: targetExecution.GetWorkflowID(),
		TargetRunID:      targetExecution.GetRunID(),
		TaskID:           taskID,
		TaskList:         taskListName,
		TaskType:         persistence.TransferTaskTypeSignalExecution,
		ScheduleID:       event.GetEventID(),
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockHistoryClient.EXPECT().SignalWorkflowExecution(gomock.Any(), s.createSignalWorkflowExecutionRequest(s.targetDomainName, transferTask, si)).Return(nil).Times(1)
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
	s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(s.version).Return(cluster.TestCurrentClusterName).AnyTimes()

	s.mockHistoryClient.EXPECT().RemoveSignalMutableState(gomock.Any(), &types.RemoveSignalMutableStateRequest{
		DomainUUID: transferTask.TargetDomainID,
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: transferTask.TargetWorkflowID,
			RunID:      transferTask.TargetRunID,
		},
		RequestID: si.SignalRequestID,
	}).Return(nil).Times(1)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessSignalExecution_Failure() {

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
	signalInput := []byte("some random signal input")
	signalControl := []byte("some random signal control")

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
	event, si := test.AddRequestSignalInitiatedEvent(mutableState, event.GetEventID(), uuid.New(),
		constants.TestTargetDomainName, targetExecution.GetWorkflowID(), targetExecution.GetRunID(), signalName, signalInput, signalControl)

	transferTask := &persistence.TransferTaskInfo{
		Version:          s.version,
		DomainID:         s.domainID,
		WorkflowID:       workflowExecution.GetWorkflowID(),
		RunID:            workflowExecution.GetRunID(),
		TargetDomainID:   s.targetDomainID,
		TargetWorkflowID: targetExecution.GetWorkflowID(),
		TargetRunID:      targetExecution.GetRunID(),
		TaskID:           taskID,
		TaskList:         taskListName,
		TaskType:         persistence.TransferTaskTypeSignalExecution,
		ScheduleID:       event.GetEventID(),
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockHistoryClient.EXPECT().SignalWorkflowExecution(gomock.Any(), s.createSignalWorkflowExecutionRequest(s.targetDomainName, transferTask, si)).Return(&types.EntityNotExistsError{}).Times(1)
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
	s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(s.version).Return(cluster.TestCurrentClusterName).AnyTimes()

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessSignalExecution_Duplication() {

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
	signalInput := []byte("some random signal input")
	signalControl := []byte("some random signal control")

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
		constants.TestTargetDomainName, targetExecution.GetWorkflowID(), targetExecution.GetRunID(), signalName, signalInput, signalControl)

	transferTask := &persistence.TransferTaskInfo{
		Version:          s.version,
		DomainID:         s.domainID,
		WorkflowID:       workflowExecution.GetWorkflowID(),
		RunID:            workflowExecution.GetRunID(),
		TargetDomainID:   s.targetDomainID,
		TargetWorkflowID: targetExecution.GetWorkflowID(),
		TargetRunID:      targetExecution.GetRunID(),
		TaskID:           taskID,
		TaskList:         taskListName,
		TaskType:         persistence.TransferTaskTypeSignalExecution,
		ScheduleID:       event.GetEventID(),
	}

	event = test.AddSignaledEvent(mutableState, event.GetEventID(), constants.TestTargetDomainName, targetExecution.GetWorkflowID(), targetExecution.GetRunID(), nil)
	mutableState.FlushBufferedEvents()

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessStartChildExecution_Success() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	childWorkflowID := "some random child workflow ID"
	childRunID := uuid.New()
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

	event, ci := test.AddStartChildWorkflowExecutionInitiatedEvent(mutableState, event.GetEventID(), uuid.New(),
		s.childDomainName, childWorkflowID, childWorkflowType, childTaskListName, nil, 1, 1, nil)

	transferTask := &persistence.TransferTaskInfo{
		Version:          s.version,
		DomainID:         s.domainID,
		WorkflowID:       workflowExecution.GetWorkflowID(),
		RunID:            workflowExecution.GetRunID(),
		TargetDomainID:   constants.TestChildDomainID,
		TargetWorkflowID: childWorkflowID,
		TargetRunID:      "",
		TaskID:           taskID,
		TaskList:         taskListName,
		TaskType:         persistence.TransferTaskTypeStartChildExecution,
		ScheduleID:       event.GetEventID(),
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	historyReq, _ := s.createChildWorkflowExecutionRequest(
		s.domainName,
		s.childDomainName,
		transferTask,
		mutableState,
		ci,
		time.Now(),
	)
	s.mockHistoryClient.EXPECT().StartWorkflowExecution(gomock.Any(), historyReq).Return(&types.StartWorkflowExecutionResponse{RunID: childRunID}, nil).Times(1)
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
	s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(s.version).Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockHistoryClient.EXPECT().ScheduleDecisionTask(gomock.Any(), &types.ScheduleDecisionTaskRequest{
		DomainUUID: constants.TestChildDomainID,
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: childWorkflowID,
			RunID:      childRunID,
		},
		IsFirstDecision: true,
	}).Return(nil).Times(1)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessStartChildExecution_WithRetry_Success() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	childWorkflowID := "some random child workflow ID"
	childRunID := uuid.New()
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

	retryPolicy := &types.RetryPolicy{
		ExpirationIntervalInSeconds: 100,
		MaximumAttempts:             3,
		InitialIntervalInSeconds:    1,
		MaximumIntervalInSeconds:    2,
		BackoffCoefficient:          1,
	}

	event, ci := test.AddStartChildWorkflowExecutionInitiatedEvent(mutableState, event.GetEventID(), uuid.New(),
		s.childDomainName, childWorkflowID, childWorkflowType, childTaskListName, nil, 1, 1, retryPolicy)

	transferTask := &persistence.TransferTaskInfo{
		Version:          s.version,
		DomainID:         s.domainID,
		WorkflowID:       workflowExecution.GetWorkflowID(),
		RunID:            workflowExecution.GetRunID(),
		TargetDomainID:   constants.TestChildDomainID,
		TargetWorkflowID: childWorkflowID,
		TargetRunID:      "",
		TaskID:           taskID,
		TaskList:         taskListName,
		TaskType:         persistence.TransferTaskTypeStartChildExecution,
		ScheduleID:       event.GetEventID(),
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	historyReq, _ := s.createChildWorkflowExecutionRequest(
		s.domainName,
		s.childDomainName,
		transferTask,
		mutableState,
		ci,
		s.mockShard.GetTimeSource().Now(),
	)
	s.mockHistoryClient.EXPECT().StartWorkflowExecution(
		gomock.Any(),
		gomock.Eq(historyReq)).Return(&types.StartWorkflowExecutionResponse{RunID: childRunID}, nil).Times(1)
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
	s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(s.version).Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockHistoryClient.EXPECT().ScheduleDecisionTask(gomock.Any(), &types.ScheduleDecisionTaskRequest{
		DomainUUID: constants.TestChildDomainID,
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: childWorkflowID,
			RunID:      childRunID,
		},
		IsFirstDecision: true,
	}).Return(nil).Times(1)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessStartChildExecution_Failure() {

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

	event, ci := test.AddStartChildWorkflowExecutionInitiatedEvent(
		mutableState,
		event.GetEventID(),
		uuid.New(),
		s.childDomainName,
		childWorkflowID,
		childWorkflowType,
		childTaskListName,
		nil,
		1,
		1,
		nil,
	)

	transferTask := &persistence.TransferTaskInfo{
		Version:          s.version,
		DomainID:         s.domainID,
		WorkflowID:       workflowExecution.GetWorkflowID(),
		RunID:            workflowExecution.GetRunID(),
		TargetDomainID:   constants.TestChildDomainID,
		TargetWorkflowID: childWorkflowID,
		TargetRunID:      "",
		TaskID:           taskID,
		TaskList:         taskListName,
		TaskType:         persistence.TransferTaskTypeStartChildExecution,
		ScheduleID:       event.GetEventID(),
	}

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	historyReq, _ := s.createChildWorkflowExecutionRequest(
		s.domainName,
		s.childDomainName,
		transferTask,
		mutableState,
		ci,
		time.Now(),
	)
	s.mockHistoryClient.EXPECT().StartWorkflowExecution(gomock.Any(), historyReq).Return(nil, &types.WorkflowExecutionAlreadyStartedError{}).Times(1)
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{}}, nil).Once()
	s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(s.version).Return(cluster.TestCurrentClusterName).AnyTimes()

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessStartChildExecution_Success_Dup() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	childWorkflowID := "some random child workflow ID"
	childRunID := uuid.New()
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

	event, ci := test.AddStartChildWorkflowExecutionInitiatedEvent(
		mutableState,
		event.GetEventID(),
		uuid.New(),
		s.childDomainName,
		childWorkflowID,
		childWorkflowType,
		childTaskListName,
		nil,
		1,
		1,
		nil,
	)

	transferTask := &persistence.TransferTaskInfo{
		Version:          s.version,
		DomainID:         s.domainID,
		WorkflowID:       workflowExecution.GetWorkflowID(),
		RunID:            workflowExecution.GetRunID(),
		TargetDomainID:   constants.TestChildDomainID,
		TargetWorkflowID: childWorkflowID,
		TargetRunID:      "",
		TaskID:           taskID,
		TaskList:         taskListName,
		TaskType:         persistence.TransferTaskTypeStartChildExecution,
		ScheduleID:       event.GetEventID(),
	}

	event = test.AddChildWorkflowExecutionStartedEvent(mutableState, event.GetEventID(), constants.TestChildDomainID, childWorkflowID, childRunID, childWorkflowType)
	ci.StartedID = event.GetEventID()
	mutableState.FlushBufferedEvents()

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockHistoryClient.EXPECT().ScheduleDecisionTask(gomock.Any(), &types.ScheduleDecisionTaskRequest{
		DomainUUID: constants.TestChildDomainID,
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: childWorkflowID,
			RunID:      childRunID,
		},
		IsFirstDecision: true,
	}).Return(nil).Times(1)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessStartChildExecution_Duplication() {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: "some random workflow ID",
		RunID:      uuid.New(),
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	childExecution := types.WorkflowExecution{
		WorkflowID: "some random child workflow ID",
		RunID:      uuid.New(),
	}
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

	event, ci := test.AddStartChildWorkflowExecutionInitiatedEvent(
		mutableState,
		event.GetEventID(),
		uuid.New(),
		s.childDomainName,
		childExecution.GetWorkflowID(),
		childWorkflowType,
		childTaskListName,
		nil,
		1,
		1,
		nil,
	)

	transferTask := &persistence.TransferTaskInfo{
		Version:          s.version,
		DomainID:         s.domainID,
		WorkflowID:       workflowExecution.GetWorkflowID(),
		RunID:            workflowExecution.GetRunID(),
		TargetDomainID:   constants.TestChildDomainID,
		TargetWorkflowID: childExecution.GetWorkflowID(),
		TargetRunID:      "",
		TaskID:           taskID,
		TaskList:         taskListName,
		TaskType:         persistence.TransferTaskTypeStartChildExecution,
		ScheduleID:       event.GetEventID(),
	}

	event = test.AddChildWorkflowExecutionStartedEvent(mutableState, event.GetEventID(), constants.TestChildDomainID, childExecution.GetWorkflowID(), childExecution.GetRunID(), childWorkflowType)
	ci.StartedID = event.GetEventID()
	event = test.AddChildWorkflowExecutionCompletedEvent(mutableState, ci.InitiatedID, &childExecution, &types.WorkflowExecutionCompletedEventAttributes{
		Result:                       []byte("some random child workflow execution result"),
		DecisionTaskCompletedEventID: transferTask.ScheduleID,
	})
	mutableState.FlushBufferedEvents()

	persistenceMutableState := s.createPersistenceMutableState(mutableState, event.GetEventID(), event.GetVersion())
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

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
				CronSchedule:                        cronSchedule,
			},
			FirstDecisionTaskBackoffSeconds: common.Int32Ptr(backoffSeconds),
		},
	)
	s.Nil(err)

	taskID := int64(59)
	di := test.AddDecisionTaskScheduledEvent(mutableState)

	transferTask := &persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeRecordWorkflowStarted,
		ScheduleID: event.GetEventID(),
	}

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

	transferTask := &persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     taskID,
		TaskList:   taskListName,
		TaskType:   persistence.TransferTaskTypeUpsertWorkflowSearchAttributes,
		ScheduleID: event.GetEventID(),
	}

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
	task *persistence.TransferTaskInfo,
	ai *persistence.ActivityInfo,
) *types.AddActivityTaskRequest {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: task.WorkflowID,
		RunID:      task.RunID,
	}
	taskList := &types.TaskList{Name: task.TaskList}

	return &types.AddActivityTaskRequest{
		DomainUUID:                    task.TargetDomainID,
		SourceDomainUUID:              task.DomainID,
		Execution:                     &workflowExecution,
		TaskList:                      taskList,
		ScheduleID:                    task.ScheduleID,
		ScheduleToStartTimeoutSeconds: common.Int32Ptr(ai.ScheduleToStartTimeout),
	}
}

func (s *transferActiveTaskExecutorSuite) createAddDecisionTaskRequest(
	task *persistence.TransferTaskInfo,
	mutableState execution.MutableState,
) *types.AddDecisionTaskRequest {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: task.WorkflowID,
		RunID:      task.RunID,
	}
	taskList := &types.TaskList{Name: task.TaskList}
	executionInfo := mutableState.GetExecutionInfo()
	timeout := executionInfo.WorkflowTimeout
	if mutableState.GetExecutionInfo().TaskList != task.TaskList {
		taskListStickyKind := types.TaskListKindSticky
		taskList.Kind = &taskListStickyKind
		timeout = executionInfo.StickyScheduleToStartTimeout
	}

	return &types.AddDecisionTaskRequest{
		DomainUUID:                    task.DomainID,
		Execution:                     &workflowExecution,
		TaskList:                      taskList,
		ScheduleID:                    task.ScheduleID,
		ScheduleToStartTimeoutSeconds: common.Int32Ptr(timeout),
	}
}

func (s *transferActiveTaskExecutorSuite) createRecordWorkflowExecutionStartedRequest(
	domainName string,
	startEvent *types.HistoryEvent,
	task *persistence.TransferTaskInfo,
	mutableState execution.MutableState,
	backoffSeconds int32,
) *persistence.RecordWorkflowExecutionStartedRequest {
	workflowExecution := types.WorkflowExecution{
		WorkflowID: task.WorkflowID,
		RunID:      task.RunID,
	}
	executionInfo := mutableState.GetExecutionInfo()
	executionTimestamp := time.Unix(0, startEvent.GetTimestamp()).Add(time.Duration(backoffSeconds) * time.Second)

	return &persistence.RecordWorkflowExecutionStartedRequest{
		Domain:             domainName,
		DomainUUID:         task.DomainID,
		Execution:          workflowExecution,
		WorkflowTypeName:   executionInfo.WorkflowTypeName,
		StartTimestamp:     startEvent.GetTimestamp(),
		ExecutionTimestamp: executionTimestamp.UnixNano(),
		WorkflowTimeout:    int64(executionInfo.WorkflowTimeout),
		TaskID:             task.TaskID,
		TaskList:           task.TaskList,
	}
}

func (s *transferActiveTaskExecutorSuite) createRequestCancelWorkflowExecutionRequest(
	targetDomainName string,
	task *persistence.TransferTaskInfo,
	rci *persistence.RequestCancelInfo,
) *types.HistoryRequestCancelWorkflowExecutionRequest {

	sourceExecution := types.WorkflowExecution{
		WorkflowID: task.WorkflowID,
		RunID:      task.RunID,
	}
	targetExecution := types.WorkflowExecution{
		WorkflowID: task.TargetWorkflowID,
		RunID:      task.TargetRunID,
	}

	return &types.HistoryRequestCancelWorkflowExecutionRequest{
		DomainUUID: task.TargetDomainID,
		CancelRequest: &types.RequestCancelWorkflowExecutionRequest{
			Domain:            targetDomainName,
			WorkflowExecution: &targetExecution,
			Identity:          execution.IdentityHistoryService,
			// Use the same request ID to dedupe RequestCancelWorkflowExecution calls
			RequestID: rci.CancelRequestID,
		},
		ExternalInitiatedEventID:  common.Int64Ptr(task.ScheduleID),
		ExternalWorkflowExecution: &sourceExecution,
		ChildWorkflowOnly:         task.TargetChildWorkflowOnly,
	}
}

func (s *transferActiveTaskExecutorSuite) createSignalWorkflowExecutionRequest(
	targetDomainName string,
	task *persistence.TransferTaskInfo,
	si *persistence.SignalInfo,
) *types.HistorySignalWorkflowExecutionRequest {

	sourceExecution := types.WorkflowExecution{
		WorkflowID: task.WorkflowID,
		RunID:      task.RunID,
	}
	targetExecution := types.WorkflowExecution{
		WorkflowID: task.TargetWorkflowID,
		RunID:      task.TargetRunID,
	}

	return &types.HistorySignalWorkflowExecutionRequest{
		DomainUUID: task.TargetDomainID,
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
		ChildWorkflowOnly:         task.TargetChildWorkflowOnly,
	}
}

func (s *transferActiveTaskExecutorSuite) createChildWorkflowExecutionRequest(
	domainName string,
	childDomainName string,
	task *persistence.TransferTaskInfo,
	mutableState execution.MutableState,
	ci *persistence.ChildExecutionInfo,
	now time.Time,
) (historyReq *types.HistoryStartWorkflowExecutionRequest, retError error) {

	event, err := mutableState.GetChildExecutionInitiatedEvent(context.Background(), task.ScheduleID)
	s.NoError(err)
	attributes := event.StartChildWorkflowExecutionInitiatedEventAttributes
	workflowExecution := types.WorkflowExecution{
		WorkflowID: task.WorkflowID,
		RunID:      task.RunID,
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
		RequestID:             ci.CreateRequestID,
		WorkflowIDReusePolicy: attributes.WorkflowIDReusePolicy,
		RetryPolicy:           attributes.RetryPolicy,
	}

	parentInfo := &types.ParentExecutionInfo{
		DomainUUID:  task.DomainID,
		Domain:      domainName,
		Execution:   &workflowExecution,
		InitiatedID: task.ScheduleID,
	}

	historyStartReq, err := common.CreateHistoryStartWorkflowRequest(
		task.TargetDomainID, frontendStartReq, now)
	if err != nil {
		return nil, err
	}

	historyStartReq.ParentExecutionInfo = parentInfo
	return historyStartReq, nil
}

func (s *transferActiveTaskExecutorSuite) createUpsertWorkflowSearchAttributesRequest(
	domainName string,
	startEvent *types.HistoryEvent,
	task *persistence.TransferTaskInfo,
	mutableState execution.MutableState,
) *persistence.UpsertWorkflowExecutionRequest {

	workflowExecution := types.WorkflowExecution{
		WorkflowID: task.WorkflowID,
		RunID:      task.RunID,
	}
	executionInfo := mutableState.GetExecutionInfo()

	return &persistence.UpsertWorkflowExecutionRequest{
		Domain:           domainName,
		DomainUUID:       task.DomainID,
		Execution:        workflowExecution,
		WorkflowTypeName: executionInfo.WorkflowTypeName,
		StartTimestamp:   startEvent.GetTimestamp(),
		WorkflowTimeout:  int64(executionInfo.WorkflowTimeout),
		TaskID:           task.TaskID,
		TaskList:         task.TaskList,
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
