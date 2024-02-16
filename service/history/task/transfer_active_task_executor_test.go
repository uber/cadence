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
	"go.uber.org/yarpc"

	hclient "github.com/uber/cadence/client/history"
	"github.com/uber/cadence/client/matching"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/archiver/provider"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	dc "github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/engine"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
	test "github.com/uber/cadence/service/history/testing"
	"github.com/uber/cadence/service/history/workflowcache"
	warchiver "github.com/uber/cadence/service/worker/archiver"
	"github.com/uber/cadence/service/worker/parentclosepolicy"
)

type (
	transferActiveTaskExecutorSuite struct {
		suite.Suite
		*require.Assertions

		controller         *gomock.Controller
		mockShard          *shard.TestContext
		mockEngine         *engine.MockEngine
		mockDomainCache    *cache.MockDomainCache
		mockWFCache        *workflowcache.MockWFCache
		mockHistoryClient  *hclient.MockClient
		mockMatchingClient *matching.MockClient

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
		timeSource                 clock.MockedTimeSource
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
	s.timeSource = clock.NewMockedTimeSource()

	s.controller = gomock.NewController(s.T())

	config := config.NewForTest()
	s.mockShard = shard.NewTestContext(
		s.T(),
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
		s.mockShard.GetDomainCache(),
	))
	s.mockShard.Resource.TimeSource = s.timeSource

	s.mockEngine = engine.NewMockEngine(s.controller)
	s.mockEngine.EXPECT().NotifyNewHistoryEvent(gomock.Any()).AnyTimes()
	s.mockEngine.EXPECT().NotifyNewTransferTasks(gomock.Any()).AnyTimes()
	s.mockEngine.EXPECT().NotifyNewTimerTasks(gomock.Any()).AnyTimes()
	s.mockEngine.EXPECT().NotifyNewCrossClusterTasks(gomock.Any()).AnyTimes()
	s.mockEngine.EXPECT().NotifyNewReplicationTasks(gomock.Any()).AnyTimes()
	s.mockShard.SetEngine(s.mockEngine)

	s.mockParentClosePolicyClient = &parentclosepolicy.ClientMock{}
	s.mockArchivalClient = &warchiver.ClientMock{}
	s.mockMatchingClient = s.mockShard.Resource.MatchingClient
	s.mockHistoryClient = s.mockShard.Resource.HistoryClient
	s.mockExecutionMgr = s.mockShard.Resource.ExecutionMgr
	s.mockHistoryV2Mgr = s.mockShard.Resource.HistoryMgr
	s.mockVisibilityMgr = s.mockShard.Resource.VisibilityMgr
	s.mockArchivalMetadata = s.mockShard.Resource.ArchivalMetadata
	s.mockArchiverProvider = s.mockShard.Resource.ArchiverProvider
	s.mockDomainCache = s.mockShard.Resource.DomainCache
	s.mockWFCache = workflowcache.NewMockWFCache(s.controller)
	s.mockDomainCache.EXPECT().GetDomainByID(s.domainID).Return(s.domainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(s.domainID).Return(s.domainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(s.domainName).Return(s.domainID, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomain(s.domainName).Return(s.domainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(s.targetDomainID).Return(s.targetDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(s.targetDomainID).Return(s.targetDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(s.targetDomainName).Return(s.targetDomainID, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomain(s.targetDomainName).Return(s.targetDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(s.remoteTargetDomainID).Return(s.remoteTargetDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(s.remoteTargetDomainID).Return(s.remoteTargetDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(s.remoteTargetDomainName).Return(s.remoteTargetDomainID, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomain(s.remoteTargetDomainName).Return(s.remoteTargetDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestParentDomainID).Return(constants.TestGlobalParentDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestParentDomainID).Return(constants.TestParentDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(constants.TestParentDomainName).Return(constants.TestParentDomainID, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomain(constants.TestParentDomainName).Return(constants.TestGlobalParentDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(s.childDomainID).Return(s.childDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(s.childDomainID).Return(s.childDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainID(s.childDomainName).Return(s.childDomainID, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomain(s.childDomainName).Return(s.childDomainEntry, nil).AnyTimes()

	s.logger = s.mockShard.GetLogger()
	s.transferActiveTaskExecutor = NewTransferActiveTaskExecutor(
		s.mockShard,
		s.mockArchivalClient,
		execution.NewCache(s.mockShard),
		nil,
		s.logger,
		config,
		s.mockWFCache,
	).(*transferActiveTaskExecutor)
	s.transferActiveTaskExecutor.parentClosePolicyClient = s.mockParentClosePolicyClient
}

func (s *transferActiveTaskExecutorSuite) TearDownTest() {
	s.controller.Finish()
	s.mockShard.Finish(s.T())
	s.mockArchivalClient.AssertExpectations(s.T())
	s.mockParentClosePolicyClient.AssertExpectations(s.T())
}

func (s *transferActiveTaskExecutorSuite) TestProcessActivityTask_Success() {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	event, ai := test.AddActivityTaskScheduledEvent(
		mutableState,
		decisionCompletionID,
		"activity-1",
		"some random activity type",
		mutableState.GetExecutionInfo().TaskList,
		[]byte{}, 1, 1, 1, 1,
	)
	mutableState.FlushBufferedEvents()

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:        s.version,
		DomainID:       s.domainID,
		TargetDomainID: s.targetDomainID,
		WorkflowID:     workflowExecution.GetWorkflowID(),
		RunID:          workflowExecution.GetRunID(),
		TaskID:         int64(59),
		TaskList:       mutableState.GetExecutionInfo().TaskList,
		TaskType:       persistence.TransferTaskTypeActivityTask,
		ScheduleID:     event.ID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockMatchingClient.EXPECT().AddActivityTask(gomock.Any(), createAddActivityTaskRequest(transferTask, ai, mutableState.GetExecutionInfo().PartitionConfig)).Return(nil).Times(1)
	s.mockWFCache.EXPECT().AllowInternal(constants.TestDomainID, constants.TestWorkflowID).Return(true).Times(1)
	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessActivityTask_Duplication() {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	event, ai := test.AddActivityTaskScheduledEvent(
		mutableState,
		decisionCompletionID,
		"activity-1",
		"some random activity type",
		mutableState.GetExecutionInfo().TaskList,
		[]byte{}, 1, 1, 1, 1,
	)

	event = test.AddActivityTaskStartedEvent(mutableState, event.ID, "")
	ai.StartedID = event.ID
	event = test.AddActivityTaskCompletedEvent(mutableState, ai.ScheduleID, ai.StartedID, nil, "")
	mutableState.FlushBufferedEvents()

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:        s.version,
		DomainID:       s.domainID,
		TargetDomainID: s.targetDomainID,
		WorkflowID:     workflowExecution.GetWorkflowID(),
		RunID:          workflowExecution.GetRunID(),
		TaskID:         int64(59),
		TaskList:       mutableState.GetExecutionInfo().TaskList,
		TaskType:       persistence.TransferTaskTypeActivityTask,
		ScheduleID:     event.ID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessDecisionTask_FirstDecision() {

	workflowExecution, mutableState, err := test.StartWorkflow(s.mockShard, s.domainID)
	s.NoError(err)

	di := test.AddDecisionTaskScheduledEvent(mutableState)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     int64(59),
		TaskList:   mutableState.GetExecutionInfo().TaskList,
		TaskType:   persistence.TransferTaskTypeDecisionTask,
		ScheduleID: di.ScheduleID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, di.ScheduleID, di.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockWFCache.EXPECT().AllowInternal(constants.TestDomainID, constants.TestWorkflowID).Return(true).Times(1)
	s.mockMatchingClient.EXPECT().AddDecisionTask(gomock.Any(), createAddDecisionTaskRequest(transferTask, mutableState)).Return(nil).Times(1)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessDecisionTask_NonFirstDecision() {

	workflowExecution, mutableState, _, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	// make another round of decision
	di := test.AddDecisionTaskScheduledEvent(mutableState)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     int64(59),
		TaskList:   mutableState.GetExecutionInfo().TaskList,
		TaskType:   persistence.TransferTaskTypeDecisionTask,
		ScheduleID: di.ScheduleID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, di.ScheduleID, di.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockWFCache.EXPECT().AllowInternal(constants.TestDomainID, constants.TestWorkflowID).Return(true).Times(1)
	s.mockMatchingClient.EXPECT().AddDecisionTask(gomock.Any(), createAddDecisionTaskRequest(transferTask, mutableState)).Return(nil).Times(1)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessDecisionTask_Sticky_NonFirstDecision() {

	workflowExecution, mutableState, _, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	// set the sticky tasklist attr
	executionInfo := mutableState.GetExecutionInfo()
	executionInfo.StickyTaskList = "some random sticky task list"
	executionInfo.StickyScheduleToStartTimeout = int32(233)

	// make another round of decision
	di := test.AddDecisionTaskScheduledEvent(mutableState)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     int64(59),
		TaskList:   executionInfo.StickyTaskList,
		TaskType:   persistence.TransferTaskTypeDecisionTask,
		ScheduleID: di.ScheduleID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, di.ScheduleID, di.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockWFCache.EXPECT().AllowInternal(constants.TestDomainID, constants.TestWorkflowID).Return(true).Times(1)
	s.mockMatchingClient.EXPECT().AddDecisionTask(gomock.Any(), createAddDecisionTaskRequest(transferTask, mutableState)).Return(nil).Times(1)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessDecisionTask_DecisionNotSticky_MutableStateSticky() {

	workflowExecution, mutableState, _, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	// set the sticky tasklist attr
	executionInfo := mutableState.GetExecutionInfo()
	executionInfo.StickyTaskList = "some random sticky task list"
	executionInfo.StickyScheduleToStartTimeout = int32(233)

	// make another round of decision
	di := test.AddDecisionTaskScheduledEvent(mutableState)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     int64(59),
		TaskList:   mutableState.GetExecutionInfo().TaskList,
		TaskType:   persistence.TransferTaskTypeDecisionTask,
		ScheduleID: di.ScheduleID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, di.ScheduleID, di.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockWFCache.EXPECT().AllowInternal(constants.TestDomainID, constants.TestWorkflowID).Return(true).Times(1)
	s.mockMatchingClient.EXPECT().AddDecisionTask(gomock.Any(), createAddDecisionTaskRequest(transferTask, mutableState)).Return(nil).Times(1)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessDecisionTask_Duplication() {

	workflowExecution, mutableState, _, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     int64(4096),
		TaskList:   mutableState.GetExecutionInfo().TaskList,
		TaskType:   persistence.TransferTaskTypeDecisionTask,
		ScheduleID: mutableState.GetPreviousStartedEventID() - 1,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, mutableState.GetNextEventID()-1, mutableState.GetCurrentVersion())
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCloseExecution_HasParent_Success() {
	s.testProcessCloseExecutionWithParent(
		s.targetDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
		) {
			s.mockVisibilityMgr.On("RecordWorkflowExecutionClosed", mock.Anything, mock.Anything).Return(nil).Once()
			s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewDisabledArchvialConfig())
		},
		false,
	)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCloseExecution_HasParent_Failure() {
	s.testProcessCloseExecutionWithParent(
		s.targetDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
		) {
			s.mockVisibilityMgr.On("RecordWorkflowExecutionClosed", mock.Anything, mock.Anything).Return(nil).Once()
			s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewDisabledArchvialConfig())
		},
		true,
	)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCloseExecution_HasParentCrossCluster() {
	s.testProcessCloseExecutionWithParent(
		s.remoteTargetDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, targetExecution types.WorkflowExecution,
		) {
			s.mockVisibilityMgr.On("RecordWorkflowExecutionClosed", mock.Anything, mock.Anything).Return(nil).Once()
			s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewDisabledArchvialConfig())
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.MatchedBy(func(request *persistence.UpdateWorkflowExecutionRequest) bool {
				s.Equal(persistence.UpdateWorkflowModeIgnoreCurrent, request.Mode)
				crossClusterTasks := request.UpdateWorkflowMutation.CrossClusterTasks
				s.Len(crossClusterTasks, 1)
				s.Equal(persistence.CrossClusterTaskTypeRecordChildExeuctionCompleted, crossClusterTasks[0].GetType())
				return true
			})).Return(&persistence.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &persistence.MutableStateUpdateSessionStats{}}, nil).Once()
		},
		false,
	)
}

func (s *transferActiveTaskExecutorSuite) testProcessCloseExecutionWithParent(
	targetDomainID string,
	setupMockFn func(
		mutableState execution.MutableState,
		workflowExecution, targetExecution types.WorkflowExecution,
	),
	failRecordChild bool,
) {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)
	executionInfo := mutableState.GetExecutionInfo()
	parentInitiatedID := int64(3222)
	parentExecution := types.WorkflowExecution{
		WorkflowID: "some random parent workflow ID",
		RunID:      uuid.New(),
	}
	executionInfo.ParentDomainID = targetDomainID
	executionInfo.ParentWorkflowID = parentExecution.WorkflowID
	executionInfo.ParentRunID = parentExecution.RunID
	executionInfo.InitiatedID = parentInitiatedID

	event := test.AddCompleteWorkflowEvent(mutableState, decisionCompletionID, nil)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     int64(59),
		TaskList:   executionInfo.TaskList,
		TaskType:   persistence.TransferTaskTypeCloseExecution,
		ScheduleID: event.ID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	if targetDomainID == s.targetDomainID {
		var recordChildErr error
		if failRecordChild {
			recordChildErr = &types.DomainNotActiveError{}
		}
		s.mockHistoryClient.EXPECT().RecordChildExecutionCompleted(gomock.Any(), &types.RecordChildExecutionCompletedRequest{
			DomainUUID:         targetDomainID,
			WorkflowExecution:  &parentExecution,
			InitiatedID:        parentInitiatedID,
			CompletedExecution: &workflowExecution,
			CompletionEvent:    event,
		}).Return(recordChildErr).Times(1)
	}

	setupMockFn(mutableState, workflowExecution, parentExecution)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	if failRecordChild {
		s.Equal(errTargetDomainNotActive, err)
	} else {
		s.NoError(err)
	}
}

func (s *transferActiveTaskExecutorSuite) TestProcessCloseExecution_NoParent() {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	event := test.AddCompleteWorkflowEvent(mutableState, decisionCompletionID, nil)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     int64(59),
		TaskList:   mutableState.GetExecutionInfo().TaskList,
		TaskType:   persistence.TransferTaskTypeCloseExecution,
		ScheduleID: event.ID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockVisibilityMgr.On("RecordWorkflowExecutionClosed", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewArchivalConfig("enabled", dc.GetStringPropertyFn("enabled"), true, dc.GetBoolPropertyFn(true), "disabled", "random URI"))
	s.mockArchivalClient.On("Archive", mock.Anything, mock.Anything).Return(nil, nil).Once()

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCloseExecution_NoParent_HasFewChildren() {
	s.testProcessCloseExecutionNoParentHasFewChildren(
		map[string]string{
			"child_abandon":   s.domainName,
			"child_terminate": s.childDomainName,
			"child_cancel":    s.domainName,
		},
		func() {
			s.expectCancelRequest(s.domainName)
			s.expectTerminateRequest(s.childDomainName)
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestApplyParentPolicy_SameClusterChild_TargetNotActive() {
	s.testProcessCloseExecutionNoParentHasFewChildrenWithError(
		map[string]string{
			"child_terminate": s.domainName,
			"child_cancel":    s.domainName,
		},
		func() {
			s.mockHistoryClient.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), gomock.Any()).
				Return(&types.DomainNotActiveError{}).MaxTimes(1)
			s.mockHistoryClient.EXPECT().TerminateWorkflowExecution(gomock.Any(), gomock.Any()).
				Return(&types.DomainNotActiveError{}).MaxTimes(1)
		},
		errTargetDomainNotActive,
	)
}

func (s *transferActiveTaskExecutorSuite) TestApplyParentPolicy_CrossClusterAbandonedChild_Abandon() {
	s.testProcessCloseExecutionNoParentHasFewChildren(
		map[string]string{
			"child_abandon":   s.remoteTargetDomainName,
			"child_terminate": s.domainName,
			"child_cancel":    s.domainName,
		},
		func() {
			s.expectCancelRequest(s.domainName)
			s.expectTerminateRequest(s.domainName)
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestApplyParentPolicy_CrossClusterChild_Terminate() {
	s.testProcessCloseExecutionNoParentHasFewChildren(
		map[string]string{
			"child_abandon":   s.domainName,
			"child_terminate": s.remoteTargetDomainName,
			"child_cancel":    s.childDomainName,
		},
		func() {
			s.expectCrossClusterApplyParentPolicyCalls()
			s.expectCancelRequest(s.childDomainName)
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestApplyParentPolicy_CrossClusterChild_Cancel() {
	s.testProcessCloseExecutionNoParentHasFewChildren(
		map[string]string{
			"child_abandon":   s.domainName,
			"child_terminate": s.domainName,
			"child_cancel":    s.remoteTargetDomainName,
		},
		func() {
			s.expectCrossClusterApplyParentPolicyCalls()
			s.expectTerminateRequest(s.domainName)
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestApplyParentPolicy_CrossClusterChildren_Mixed() {
	s.testProcessCloseExecutionNoParentHasFewChildren(
		map[string]string{
			"child_abandon":   s.remoteTargetDomainName,
			"child_terminate": s.remoteTargetDomainName,
			"child_cancel":    s.remoteTargetDomainName,
		},
		func() {
			s.expectCrossClusterApplyParentPolicyCalls()
		},
	)
}

func (s *transferActiveTaskExecutorSuite) expectCancelRequest(childDomainName string) {
	childDomainID, err := s.mockDomainCache.GetDomainID(childDomainName)
	s.NoError(err)
	s.mockHistoryClient.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), gomock.Any()).DoAndReturn(
		func(
			_ context.Context,
			request *types.HistoryRequestCancelWorkflowExecutionRequest,
			option ...yarpc.CallOption,
		) error {
			s.Equal(childDomainID, request.DomainUUID)
			s.Equal(childDomainName, request.CancelRequest.Domain)
			s.True(request.GetChildWorkflowOnly())
			errors := []error{nil, &types.CancellationAlreadyRequestedError{}, &types.EntityNotExistsError{}}
			return errors[rand.Intn(len(errors))]
		},
	).Times(1)
}

func (s *transferActiveTaskExecutorSuite) expectTerminateRequest(childDomainName string) {
	childDomainID, err := s.mockDomainCache.GetDomainID(childDomainName)
	s.NoError(err)
	s.mockHistoryClient.EXPECT().TerminateWorkflowExecution(gomock.Any(), gomock.Any()).DoAndReturn(
		func(
			_ context.Context,
			request *types.HistoryTerminateWorkflowExecutionRequest,
			option ...yarpc.CallOption,
		) error {
			s.Equal(childDomainID, request.DomainUUID)
			s.Equal(childDomainName, request.TerminateRequest.Domain)
			errors := []error{nil, &types.EntityNotExistsError{}}
			return errors[rand.Intn(len(errors))]
		},
	).Times(1)
}

func (s *transferActiveTaskExecutorSuite) expectCrossClusterApplyParentPolicyCalls() {
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.MatchedBy(func(request *persistence.UpdateWorkflowExecutionRequest) bool {
		s.Equal(persistence.UpdateWorkflowModeIgnoreCurrent, request.Mode)
		crossClusterTasks := request.UpdateWorkflowMutation.CrossClusterTasks
		s.Len(crossClusterTasks, 1)
		s.Equal(persistence.CrossClusterTaskTypeApplyParentClosePolicy, crossClusterTasks[0].GetType())
		return true
	})).Return(&persistence.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &persistence.MutableStateUpdateSessionStats{}}, nil).Once()
	s.mockDomainCache.EXPECT().GetDomain(s.remoteTargetDomainName).Return(s.remoteTargetDomainEntry, nil).AnyTimes()
}

func (s *transferActiveTaskExecutorSuite) testProcessCloseExecutionNoParentHasFewChildren(
	childrenDomainNames map[string]string,
	setupMockFn func(),
) {
	s.testProcessCloseExecutionNoParentHasFewChildrenWithError(childrenDomainNames, setupMockFn, nil)
}

func (s *transferActiveTaskExecutorSuite) testProcessCloseExecutionNoParentHasFewChildrenWithError(
	childrenDomainNames map[string]string,
	setupMockFn func(),
	expectedErr error,
) {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	parentClosePolicy1 := types.ParentClosePolicyAbandon
	parentClosePolicy2 := types.ParentClosePolicyTerminate
	parentClosePolicy3 := types.ParentClosePolicyRequestCancel

	_, _, err = mutableState.AddStartChildWorkflowExecutionInitiatedEvent(decisionCompletionID, uuid.New(), &types.StartChildWorkflowExecutionDecisionAttributes{
		Domain:     childrenDomainNames["child_abandon"],
		WorkflowID: "child workflow1",
		WorkflowType: &types.WorkflowType{
			Name: "child workflow type",
		},
		TaskList:          &types.TaskList{Name: mutableState.GetExecutionInfo().TaskList},
		Input:             []byte("random input"),
		ParentClosePolicy: &parentClosePolicy1,
	})
	s.Nil(err)
	_, _, err = mutableState.AddStartChildWorkflowExecutionInitiatedEvent(decisionCompletionID, uuid.New(), &types.StartChildWorkflowExecutionDecisionAttributes{
		Domain:     childrenDomainNames["child_terminate"],
		WorkflowID: "child workflow2",
		WorkflowType: &types.WorkflowType{
			Name: "child workflow type",
		},
		TaskList:          &types.TaskList{Name: mutableState.GetExecutionInfo().TaskList},
		Input:             []byte("random input"),
		ParentClosePolicy: &parentClosePolicy2,
	})
	s.Nil(err)
	_, _, err = mutableState.AddStartChildWorkflowExecutionInitiatedEvent(decisionCompletionID, uuid.New(), &types.StartChildWorkflowExecutionDecisionAttributes{
		Domain:     childrenDomainNames["child_cancel"],
		WorkflowID: "child workflow3",
		WorkflowType: &types.WorkflowType{
			Name: "child workflow type",
		},
		TaskList:          &types.TaskList{Name: mutableState.GetExecutionInfo().TaskList},
		Input:             []byte("random input"),
		ParentClosePolicy: &parentClosePolicy3,
	})
	s.Nil(err)
	s.NoError(mutableState.FlushBufferedEvents())

	event := test.AddCompleteWorkflowEvent(mutableState, decisionCompletionID, nil)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     int64(59),
		TaskList:   mutableState.GetExecutionInfo().TaskList,
		TaskType:   persistence.TransferTaskTypeCloseExecution,
		ScheduleID: event.ID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockVisibilityMgr.On("RecordWorkflowExecutionClosed", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewDisabledArchvialConfig())
	setupMockFn()

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Equal(expectedErr, err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCloseExecution_NoParent_HasManyChildren() {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	numChildWorkflows := 10
	for i := 0; i < numChildWorkflows; i++ {
		_, _, err = mutableState.AddStartChildWorkflowExecutionInitiatedEvent(decisionCompletionID, uuid.New(), &types.StartChildWorkflowExecutionDecisionAttributes{
			Domain:     s.childDomainName,
			WorkflowID: "child workflow" + strconv.Itoa(i),
			WorkflowType: &types.WorkflowType{
				Name: "child workflow type",
			},
			TaskList:          &types.TaskList{Name: mutableState.GetExecutionInfo().TaskList},
			Input:             []byte("random input"),
			ParentClosePolicy: types.ParentClosePolicyTerminate.Ptr(),
		})
		s.Nil(err)
	}

	s.NoError(mutableState.FlushBufferedEvents())

	event := test.AddCompleteWorkflowEvent(mutableState, decisionCompletionID, nil)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     int64(59),
		TaskList:   mutableState.GetExecutionInfo().TaskList,
		TaskType:   persistence.TransferTaskTypeCloseExecution,
		ScheduleID: event.ID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	s.mockVisibilityMgr.On("RecordWorkflowExecutionClosed", mock.Anything, mock.Anything).Return(nil).Once()
	s.mockArchivalMetadata.On("GetVisibilityConfig").Return(archiver.NewDisabledArchvialConfig())
	s.mockParentClosePolicyClient.On("SendParentClosePolicyRequest", mock.Anything, mock.MatchedBy(
		func(request parentclosepolicy.Request) bool {
			if len(request.Executions) != numChildWorkflows {
				return false
			}
			for _, executions := range request.Executions {
				if executions.DomainName != s.childDomainName {
					return false
				}
			}
			return true
		},
	)).Return(nil).Times(1)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessCloseExecution_NoParent_HasManyAbandonedChildren() {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	for i := 0; i < 10; i++ {
		_, _, err = mutableState.AddStartChildWorkflowExecutionInitiatedEvent(decisionCompletionID, uuid.New(), &types.StartChildWorkflowExecutionDecisionAttributes{
			WorkflowID: "child workflow" + strconv.Itoa(i),
			WorkflowType: &types.WorkflowType{
				Name: "child workflow type",
			},
			TaskList:          &types.TaskList{Name: mutableState.GetExecutionInfo().TaskList},
			Input:             []byte("random input"),
			ParentClosePolicy: types.ParentClosePolicyAbandon.Ptr(),
		})
		s.Nil(err)
	}

	s.NoError(mutableState.FlushBufferedEvents())

	event := test.AddCompleteWorkflowEvent(mutableState, decisionCompletionID, nil)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     int64(59),
		TaskList:   mutableState.GetExecutionInfo().TaskList,
		TaskType:   persistence.TransferTaskTypeCloseExecution,
		ScheduleID: event.ID,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
	s.NoError(err)
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
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			cancelRequest := createTestRequestCancelWorkflowExecutionRequest(s.targetDomainName, transferTask.GetInfo().(*persistence.TransferTaskInfo), requestCancelInfo.CancelRequestID)
			s.mockHistoryClient.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), cancelRequest).Return(nil).Times(1)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&persistence.AppendHistoryNodesResponse{}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &persistence.MutableStateUpdateSessionStats{}}, nil).Once()
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
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			cancelRequest := createTestRequestCancelWorkflowExecutionRequest(s.targetDomainName, transferTask.GetInfo().(*persistence.TransferTaskInfo), requestCancelInfo.CancelRequestID)
			s.mockHistoryClient.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), cancelRequest).Return(&types.EntityNotExistsError{}).Times(1)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&persistence.AppendHistoryNodesResponse{}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &persistence.MutableStateUpdateSessionStats{}}, nil).Once()
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
			event = test.AddCancelRequestedEvent(mutableState, event.ID, s.targetDomainID, targetExecution.GetWorkflowID(), targetExecution.GetRunID())
			mutableState.FlushBufferedEvents()

			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
			s.NoError(err)
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
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.MatchedBy(func(request *persistence.UpdateWorkflowExecutionRequest) bool {
				s.Equal(persistence.UpdateWorkflowModeIgnoreCurrent, request.Mode)
				crossClusterTasks := request.UpdateWorkflowMutation.CrossClusterTasks
				s.Len(crossClusterTasks, 1)
				s.Equal(persistence.CrossClusterTaskTypeCancelExecution, crossClusterTasks[0].GetType())
				return true
			})).Return(&persistence.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &persistence.MutableStateUpdateSessionStats{}}, nil).Once()
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
	s.testProcessCancelExecutionWithError(targetDomainID, setupMockFn, nil)
}

func (s *transferActiveTaskExecutorSuite) testProcessCancelExecutionWithError(
	targetDomainID string,
	setupMockFn func(
		mutableState execution.MutableState,
		workflowExecution, targetExecution types.WorkflowExecution,
		cancelInitEvent *types.HistoryEvent,
		transferTask Task,
		requestCancelInfo *persistence.RequestCancelInfo,
	),
	expectedErr error,
) {
	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
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
		ScheduleID:       event.ID,
	})

	setupMockFn(mutableState, workflowExecution, targetExecution, event, transferTask, rci)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Equal(expectedErr, err)
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
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			signalRequest := createTestSignalWorkflowExecutionRequest(s.targetDomainName, transferTask.GetInfo().(*persistence.TransferTaskInfo), signalInfo)
			s.mockHistoryClient.EXPECT().SignalWorkflowExecution(gomock.Any(), signalRequest).Return(nil).Times(1)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&persistence.AppendHistoryNodesResponse{}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &persistence.MutableStateUpdateSessionStats{}}, nil).Once()

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
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			signalRequest := createTestSignalWorkflowExecutionRequest(s.targetDomainName, transferTask.GetInfo().(*persistence.TransferTaskInfo), signalInfo)
			s.mockHistoryClient.EXPECT().SignalWorkflowExecution(gomock.Any(), signalRequest).Return(&types.EntityNotExistsError{}).Times(1)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&persistence.AppendHistoryNodesResponse{}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &persistence.MutableStateUpdateSessionStats{}}, nil).Once()
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
			event = test.AddSignaledEvent(mutableState, event.ID, s.targetDomainName, targetExecution.GetWorkflowID(), targetExecution.GetRunID(), nil)
			mutableState.FlushBufferedEvents()

			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
			s.NoError(err)
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
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.MatchedBy(func(request *persistence.UpdateWorkflowExecutionRequest) bool {
				s.Equal(persistence.UpdateWorkflowModeIgnoreCurrent, request.Mode)
				crossClusterTasks := request.UpdateWorkflowMutation.CrossClusterTasks
				s.Len(crossClusterTasks, 1)
				s.Equal(persistence.CrossClusterTaskTypeSignalExecution, crossClusterTasks[0].GetType())
				return true
			})).Return(&persistence.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &persistence.MutableStateUpdateSessionStats{}}, nil).Once()
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
	s.testProcessSignalExecutionWithError(targetDomainID, setupMockFn, nil)
}

func (s *transferActiveTaskExecutorSuite) testProcessSignalExecutionWithError(
	targetDomainID string,
	setupMockFn func(
		mutableState execution.MutableState,
		workflowExecution, targetExecution types.WorkflowExecution,
		signalInitEvent *types.HistoryEvent,
		transferTask Task,
		signalInfo *persistence.SignalInfo,
	),
	expectedErr error,
) {
	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
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
		ScheduleID:       event.ID,
	})

	setupMockFn(mutableState, workflowExecution, targetExecution, event, transferTask, signalInfo)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Equal(expectedErr, err)
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
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			taskInfo := transferTask.GetInfo().(*persistence.TransferTaskInfo)
			event, err = mutableState.GetChildExecutionInitiatedEvent(context.Background(), taskInfo.ScheduleID)
			s.NoError(err)
			historyReq, err := createTestChildWorkflowExecutionRequest(
				s.domainName,
				s.childDomainName,
				taskInfo,
				event.StartChildWorkflowExecutionInitiatedEventAttributes,
				childInfo.CreateRequestID,
				s.mockShard.GetTimeSource().Now(),
				mutableState.GetExecutionInfo().PartitionConfig,
			)
			require.NoError(s.T(), err)
			s.mockHistoryClient.EXPECT().StartWorkflowExecution(gomock.Any(), historyReq).Return(&types.StartWorkflowExecutionResponse{RunID: childExecution.RunID}, nil).Times(1)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&persistence.AppendHistoryNodesResponse{}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &persistence.MutableStateUpdateSessionStats{}}, nil).Once()
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
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			taskInfo := transferTask.GetInfo().(*persistence.TransferTaskInfo)
			event, err = mutableState.GetChildExecutionInitiatedEvent(context.Background(), taskInfo.ScheduleID)
			s.NoError(err)
			historyReq, err := createTestChildWorkflowExecutionRequest(
				s.domainName,
				s.childDomainName,
				taskInfo,
				event.StartChildWorkflowExecutionInitiatedEventAttributes,
				childInfo.CreateRequestID,
				s.mockShard.GetTimeSource().Now(),
				mutableState.GetExecutionInfo().PartitionConfig,
			)
			require.NoError(s.T(), err)
			s.mockHistoryClient.EXPECT().StartWorkflowExecution(gomock.Any(), historyReq).Return(nil, &types.WorkflowExecutionAlreadyStartedError{}).Times(1)
			s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&persistence.AppendHistoryNodesResponse{}, nil).Once()
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &persistence.MutableStateUpdateSessionStats{}}, nil).Once()
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestProcessStartChildExecution_TargetNotActive() {
	s.testProcessStartChildExecutionWithError(
		s.childDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, childExecution types.WorkflowExecution,
			event *types.HistoryEvent,
			transferTask Task,
			childInfo *persistence.ChildExecutionInfo,
		) {
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			taskInfo := transferTask.GetInfo().(*persistence.TransferTaskInfo)
			_, err = mutableState.GetChildExecutionInitiatedEvent(context.Background(), taskInfo.ScheduleID)
			s.NoError(err)
			s.mockHistoryClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, &types.DomainNotActiveError{}).Times(1)
		},
		errTargetDomainNotActive,
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
			startEvent := test.AddChildWorkflowExecutionStartedEvent(mutableState, event.ID, s.childDomainID, childExecution.WorkflowID, childExecution.RunID, childInfo.WorkflowTypeName)
			childInfo.StartedID = startEvent.ID
			mutableState.FlushBufferedEvents()

			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, startEvent.ID, startEvent.Version)
			s.NoError(err)
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

func (s *transferActiveTaskExecutorSuite) TestProcessStartChildExecution_Dup_TargetNotActive() {
	s.testProcessStartChildExecutionWithError(
		s.childDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, childExecution types.WorkflowExecution,
			event *types.HistoryEvent,
			transferTask Task,
			childInfo *persistence.ChildExecutionInfo,
		) {
			startEvent := test.AddChildWorkflowExecutionStartedEvent(mutableState, event.ID, s.childDomainID, childExecution.WorkflowID, childExecution.RunID, childInfo.WorkflowTypeName)
			childInfo.StartedID = startEvent.ID
			mutableState.FlushBufferedEvents()

			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, startEvent.ID, startEvent.Version)
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			s.mockHistoryClient.EXPECT().ScheduleDecisionTask(gomock.Any(), gomock.Any()).Return(&types.DomainNotActiveError{}).Times(1)
		},
		errTargetDomainNotActive,
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
			startEvent := test.AddChildWorkflowExecutionStartedEvent(mutableState, event.ID, s.childDomainID, childExecution.GetWorkflowID(), childExecution.GetRunID(), childInfo.WorkflowTypeName)
			childInfo.StartedID = startEvent.ID
			startEvent = test.AddChildWorkflowExecutionCompletedEvent(mutableState, childInfo.InitiatedID, &childExecution, &types.WorkflowExecutionCompletedEventAttributes{
				Result:                       []byte("some random child workflow execution result"),
				DecisionTaskCompletedEventID: transferTask.GetInfo().(*persistence.TransferTaskInfo).ScheduleID,
			})
			mutableState.FlushBufferedEvents()

			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, startEvent.ID, startEvent.Version)
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestProcessStartChildExecution_StartedAbandonChild_ParentClosed() {
	s.testProcessStartChildExecution(
		s.childDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, childExecution types.WorkflowExecution,
			event *types.HistoryEvent,
			transferTask Task,
			childInfo *persistence.ChildExecutionInfo,
		) {
			event = test.AddChildWorkflowExecutionStartedEvent(mutableState, event.ID, s.childDomainID, childExecution.WorkflowID, childExecution.RunID, childInfo.WorkflowTypeName)
			childInfo.StartedID = event.ID
			di := test.AddDecisionTaskScheduledEvent(mutableState)
			event = test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, mutableState.GetExecutionInfo().TaskList, "some random identity")
			event = test.AddDecisionTaskCompletedEvent(mutableState, di.ScheduleID, event.ID, nil, "some random identity")
			event = test.AddCompleteWorkflowEvent(mutableState, event.ID, nil)
			mutableState.FlushBufferedEvents()

			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
			s.NoError(err)
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
			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.MatchedBy(func(request *persistence.UpdateWorkflowExecutionRequest) bool {
				s.Equal(persistence.UpdateWorkflowModeIgnoreCurrent, request.Mode)
				crossClusterTasks := request.UpdateWorkflowMutation.CrossClusterTasks
				s.Len(crossClusterTasks, 1)
				s.Equal(persistence.CrossClusterTaskTypeStartChildExecution, crossClusterTasks[0].GetType())
				return true
			})).Return(&persistence.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &persistence.MutableStateUpdateSessionStats{}}, nil).Once()
		},
	)
}

func (s *transferActiveTaskExecutorSuite) TestProcessStartChildExecution_CrossCluster_Dup() {
	s.testProcessStartChildExecution(
		s.remoteTargetDomainID,
		func(
			mutableState execution.MutableState,
			workflowExecution, childExecution types.WorkflowExecution,
			event *types.HistoryEvent,
			transferTask Task,
			childInfo *persistence.ChildExecutionInfo,
		) {
			_ = test.AddChildWorkflowExecutionStartedEvent(mutableState, event.ID, s.childDomainID, childExecution.WorkflowID, childExecution.RunID, childInfo.WorkflowTypeName)
			di := test.AddDecisionTaskScheduledEvent(mutableState)
			event = test.AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, mutableState.GetExecutionInfo().TaskList, "some random identity")
			event = test.AddDecisionTaskCompletedEvent(mutableState, di.ScheduleID, event.ID, nil, "some random identity")
			event = test.AddCompleteWorkflowEvent(mutableState, event.ID, nil)
			mutableState.FlushBufferedEvents()

			persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, event.ID, event.Version)
			s.NoError(err)
			s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
			s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.MatchedBy(func(request *persistence.UpdateWorkflowExecutionRequest) bool {
				s.Equal(persistence.UpdateWorkflowModeIgnoreCurrent, request.Mode)
				crossClusterTasks := request.UpdateWorkflowMutation.CrossClusterTasks
				s.Len(crossClusterTasks, 1)
				s.Equal(persistence.CrossClusterTaskTypeStartChildExecution, crossClusterTasks[0].GetType())
				return true
			})).Return(&persistence.UpdateWorkflowExecutionResponse{MutableStateUpdateSessionStats: &persistence.MutableStateUpdateSessionStats{}}, nil).Once()
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
	s.testProcessStartChildExecutionWithError(targetDomainID, setupMockFn, nil)
}

func (s *transferActiveTaskExecutorSuite) testProcessStartChildExecutionWithError(
	targetDomainID string,
	setupMockFn func(
		mutableState execution.MutableState,
		workflowExecution, targetExecution types.WorkflowExecution,
		childInitEvent *types.HistoryEvent,
		transferTask Task,
		childInfo *persistence.ChildExecutionInfo,
	),
	expectedErr error,
) {
	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
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
		ScheduleID:       event.ID,
	})

	setupMockFn(mutableState, workflowExecution, childExecution, event, transferTask, ci)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Equal(expectedErr, err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessRecordWorkflowStartedTask() {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)
	executionInfo := mutableState.GetExecutionInfo()
	executionInfo.CronSchedule = "@every 5s"
	startEvent, err := mutableState.GetStartEvent(context.Background())
	s.NoError(err)
	startEvent.WorkflowExecutionStartedEventAttributes.FirstDecisionTaskBackoffSeconds = common.Int32Ptr(5)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     int64(59),
		TaskList:   mutableState.GetExecutionInfo().TaskList,
		TaskType:   persistence.TransferTaskTypeRecordWorkflowStarted,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, decisionCompletionID, mutableState.GetCurrentVersion())
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	if s.mockShard.GetConfig().EnableRecordWorkflowExecutionUninitialized(s.domainName) {
		s.mockVisibilityMgr.On(
			"RecordWorkflowExecutionUninitialized",
			mock.Anything,
			createRecordWorkflowExecutionUninitializedRequest(transferTask, mutableState, s.mockShard.GetTimeSource().Now(), 1234),
		).Once().Return(nil)
	}
	s.mockVisibilityMgr.On(
		"RecordWorkflowExecutionStarted",
		mock.Anything,
		createRecordWorkflowExecutionStartedRequest(
			s.domainName, startEvent, transferTask, mutableState, 2, s.mockShard.GetTimeSource().Now()),
	).Once().Return(nil)

	err = s.transferActiveTaskExecutor.Execute(transferTask, true)
	s.Nil(err)
}

func (s *transferActiveTaskExecutorSuite) TestProcessUpsertWorkflowSearchAttributes() {

	workflowExecution, mutableState, decisionCompletionID, err := test.SetupWorkflowWithCompletedDecision(s.mockShard, s.domainID)
	s.NoError(err)

	transferTask := s.newTransferTaskFromInfo(&persistence.TransferTaskInfo{
		Version:    s.version,
		DomainID:   s.domainID,
		WorkflowID: workflowExecution.GetWorkflowID(),
		RunID:      workflowExecution.GetRunID(),
		TaskID:     int64(59),
		TaskList:   mutableState.GetExecutionInfo().TaskList,
		TaskType:   persistence.TransferTaskTypeUpsertWorkflowSearchAttributes,
	})

	persistenceMutableState, err := test.CreatePersistenceMutableState(mutableState, decisionCompletionID, mutableState.GetCurrentVersion())
	s.NoError(err)
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{State: persistenceMutableState}, nil)
	startEvent, err := mutableState.GetStartEvent(context.Background())
	s.NoError(err)
	s.mockVisibilityMgr.On(
		"UpsertWorkflowExecution",
		mock.Anything,
		createUpsertWorkflowSearchAttributesRequest(
			s.domainName, startEvent, transferTask, mutableState, 2, s.mockShard.GetTimeSource().Now()),
	).Once().Return(nil)

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

func (s *transferActiveTaskExecutorSuite) newTransferTaskFromInfo(
	info *persistence.TransferTaskInfo,
) Task {
	return NewTransferTask(s.mockShard, info, QueueTypeActiveTransfer, s.logger, nil, nil, nil, nil, nil)
}

func createAddActivityTaskRequest(
	transferTask Task,
	ai *persistence.ActivityInfo,
	partitionConfig map[string]string,
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
		PartitionConfig:               partitionConfig,
	}
}

func createAddDecisionTaskRequest(
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
		PartitionConfig:               executionInfo.PartitionConfig,
	}
}

func createRecordWorkflowExecutionStartedRequest(
	domainName string,
	startEvent *types.HistoryEvent,
	transferTask Task,
	mutableState execution.MutableState,
	numClusters int16,
	updateTime time.Time,
) *persistence.RecordWorkflowExecutionStartedRequest {
	taskInfo := transferTask.GetInfo().(*persistence.TransferTaskInfo)
	workflowExecution := types.WorkflowExecution{
		WorkflowID: taskInfo.WorkflowID,
		RunID:      taskInfo.RunID,
	}
	executionInfo := mutableState.GetExecutionInfo()
	backoffSeconds := startEvent.WorkflowExecutionStartedEventAttributes.GetFirstDecisionTaskBackoffSeconds()
	executionTimestamp := int64(0)
	if backoffSeconds != 0 {
		executionTimestamp = startEvent.GetTimestamp() + int64(backoffSeconds)*int64(time.Second)
	}
	return &persistence.RecordWorkflowExecutionStartedRequest{
		Domain:             domainName,
		DomainUUID:         taskInfo.DomainID,
		Execution:          workflowExecution,
		WorkflowTypeName:   executionInfo.WorkflowTypeName,
		StartTimestamp:     startEvent.GetTimestamp(),
		ExecutionTimestamp: executionTimestamp,
		WorkflowTimeout:    int64(executionInfo.WorkflowTimeout),
		TaskID:             taskInfo.TaskID,
		TaskList:           taskInfo.TaskList,
		IsCron:             len(executionInfo.CronSchedule) > 0,
		NumClusters:        numClusters,
		UpdateTimestamp:    updateTime.UnixNano(),
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
	partitionConfig map[string]string,
) (*types.HistoryStartWorkflowExecutionRequest, error) {

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

	historyStartReq, err := common.CreateHistoryStartWorkflowRequest(
		taskInfo.TargetDomainID, frontendStartReq, now, partitionConfig)
	if err != nil {
		return nil, err
	}

	historyStartReq.ParentExecutionInfo = parentInfo
	return historyStartReq, nil
}

func createUpsertWorkflowSearchAttributesRequest(
	domainName string,
	startEvent *types.HistoryEvent,
	transferTask Task,
	mutableState execution.MutableState,
	numClusters int16,
	updateTime time.Time,
) *persistence.UpsertWorkflowExecutionRequest {

	taskInfo := transferTask.GetInfo().(*persistence.TransferTaskInfo)
	workflowExecution := types.WorkflowExecution{
		WorkflowID: taskInfo.WorkflowID,
		RunID:      taskInfo.RunID,
	}
	executionInfo := mutableState.GetExecutionInfo()
	backoffSeconds := startEvent.WorkflowExecutionStartedEventAttributes.GetFirstDecisionTaskBackoffSeconds()
	executionTimestamp := int64(0)
	if backoffSeconds != 0 {
		executionTimestamp = startEvent.GetTimestamp() + int64(backoffSeconds)*int64(time.Second)
	}

	return &persistence.UpsertWorkflowExecutionRequest{
		Domain:             domainName,
		DomainUUID:         taskInfo.DomainID,
		Execution:          workflowExecution,
		WorkflowTypeName:   executionInfo.WorkflowTypeName,
		StartTimestamp:     startEvent.GetTimestamp(),
		ExecutionTimestamp: executionTimestamp,
		WorkflowTimeout:    int64(executionInfo.WorkflowTimeout),
		TaskID:             taskInfo.TaskID,
		TaskList:           taskInfo.TaskList,
		IsCron:             len(executionInfo.CronSchedule) > 0,
		NumClusters:        numClusters,
		UpdateTimestamp:    updateTime.UnixNano(),
	}
}

func createRecordWorkflowExecutionUninitializedRequest(
	transferTask Task,
	mutableState execution.MutableState,
	updateTime time.Time,
	shardID int64,
) *persistence.RecordWorkflowExecutionUninitializedRequest {
	taskInfo := transferTask.GetInfo().(*persistence.TransferTaskInfo)
	workflowExecution := types.WorkflowExecution{
		WorkflowID: taskInfo.WorkflowID,
		RunID:      taskInfo.RunID,
	}
	executionInfo := mutableState.GetExecutionInfo()
	return &persistence.RecordWorkflowExecutionUninitializedRequest{
		DomainUUID:       taskInfo.DomainID,
		Execution:        workflowExecution,
		WorkflowTypeName: executionInfo.WorkflowTypeName,
		UpdateTimestamp:  updateTime.UnixNano(),
		ShardID:          shardID,
	}
}
