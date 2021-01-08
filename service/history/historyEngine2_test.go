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

package history

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/mocks"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/query"
	"github.com/uber/cadence/service/history/queue"
	"github.com/uber/cadence/service/history/shard"
	test "github.com/uber/cadence/service/history/testing"
)

type (
	engine2Suite struct {
		suite.Suite
		*require.Assertions

		controller          *gomock.Controller
		mockShard           *shard.TestContext
		mockTxProcessor     *queue.MockProcessor
		mockTimerProcessor  *queue.MockProcessor
		mockEventsCache     *events.MockCache
		mockDomainCache     *cache.MockDomainCache
		mockClusterMetadata *cluster.MockMetadata

		historyEngine    *historyEngineImpl
		mockExecutionMgr *mocks.ExecutionManager
		mockHistoryV2Mgr *mocks.HistoryV2Manager

		config *config.Config
		logger log.Logger
	}
)

func TestEngine2Suite(t *testing.T) {
	s := new(engine2Suite)
	suite.Run(t, s)
}

func (s *engine2Suite) SetupSuite() {
	s.config = config.NewForTest()
}

func (s *engine2Suite) TearDownSuite() {
}

func (s *engine2Suite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())

	s.mockTxProcessor = queue.NewMockProcessor(s.controller)
	s.mockTimerProcessor = queue.NewMockProcessor(s.controller)
	s.mockTxProcessor.EXPECT().NotifyNewTask(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	s.mockTimerProcessor.EXPECT().NotifyNewTask(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	s.mockShard = shard.NewTestContext(
		s.controller,
		&p.ShardInfo{
			ShardID:          0,
			RangeID:          1,
			TransferAckLevel: 0,
		},
		s.config,
	)

	s.mockDomainCache = s.mockShard.Resource.DomainCache
	s.mockExecutionMgr = s.mockShard.Resource.ExecutionMgr
	s.mockHistoryV2Mgr = s.mockShard.Resource.HistoryMgr
	s.mockClusterMetadata = s.mockShard.Resource.ClusterMetadata
	s.mockEventsCache = s.mockShard.MockEventsCache
	s.mockDomainCache.EXPECT().GetDomainByID(gomock.Any()).Return(cache.NewLocalDomainCacheEntryForTest(
		&p.DomainInfo{ID: constants.TestDomainID}, &p.DomainConfig{}, "", nil,
	), nil).AnyTimes()
	s.mockEventsCache.EXPECT().PutEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	s.mockClusterMetadata.EXPECT().IsGlobalDomainEnabled().Return(false).AnyTimes()
	s.mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockClusterMetadata.EXPECT().ClusterNameForFailoverVersion(common.EmptyVersion).Return(cluster.TestCurrentClusterName).AnyTimes()

	s.logger = s.mockShard.GetLogger()

	executionCache := execution.NewCache(s.mockShard)
	h := &historyEngineImpl{
		currentClusterName:   s.mockShard.GetClusterMetadata().GetCurrentClusterName(),
		shard:                s.mockShard,
		clusterMetadata:      s.mockClusterMetadata,
		executionManager:     s.mockExecutionMgr,
		historyV2Mgr:         s.mockHistoryV2Mgr,
		executionCache:       executionCache,
		logger:               s.logger,
		throttledLogger:      s.logger,
		metricsClient:        metrics.NewClient(tally.NoopScope, metrics.History),
		tokenSerializer:      common.NewJSONTaskTokenSerializer(),
		config:               s.config,
		timeSource:           s.mockShard.GetTimeSource(),
		historyEventNotifier: events.NewNotifier(clock.NewRealTimeSource(), metrics.NewClient(tally.NoopScope, metrics.History), func(string) int { return 0 }),
		txProcessor:          s.mockTxProcessor,
		timerProcessor:       s.mockTimerProcessor,
	}
	s.mockShard.SetEngine(h)
	h.decisionHandler = newDecisionHandler(h)

	s.historyEngine = h
}

func (s *engine2Suite) TearDownTest() {
	s.controller.Finish()
	s.mockShard.Finish(s.T())
}

func (s *engine2Suite) TestRecordDecisionTaskStartedSuccessStickyExpired() {
	domainID := constants.TestDomainID
	we := types.WorkflowExecution{
		WorkflowID: common.StringPtr("wId"),
		RunID:      common.StringPtr(constants.TestRunID),
	}
	tl := "testTaskList"
	stickyTl := "stickyTaskList"
	identity := "testIdentity"

	msBuilder := execution.NewMutableStateBuilderWithEventV2(
		s.historyEngine.shard,
		loggerimpl.NewDevelopmentForTest(s.Suite),
		we.GetRunID(),
		constants.TestLocalDomainEntry,
	)
	executionInfo := msBuilder.GetExecutionInfo()
	executionInfo.StickyTaskList = stickyTl

	test.AddWorkflowExecutionStartedEvent(msBuilder, we, "wType", tl, []byte("input"), 100, 200, identity)
	di := test.AddDecisionTaskScheduledEvent(msBuilder)

	ms := execution.CreatePersistenceMutableState(msBuilder)

	gwmsResponse := &p.GetWorkflowExecutionResponse{State: ms}

	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse, nil).Once()
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{
		MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{},
	}, nil).Once()

	request := types.RecordDecisionTaskStartedRequest{
		DomainUUID:        common.StringPtr(domainID),
		WorkflowExecution: &we,
		ScheduleID:        common.Int64Ptr(2),
		TaskID:            common.Int64Ptr(100),
		RequestID:         common.StringPtr("reqId"),
		PollRequest: &types.PollForDecisionTaskRequest{
			TaskList: &types.TaskList{
				Name: common.StringPtr(stickyTl),
			},
			Identity: common.StringPtr(identity),
		},
	}

	expectedResponse := types.RecordDecisionTaskStartedResponse{}
	expectedResponse.WorkflowType = msBuilder.GetWorkflowType()
	executionInfo = msBuilder.GetExecutionInfo()
	if executionInfo.LastProcessedEvent != common.EmptyEventID {
		expectedResponse.PreviousStartedEventID = common.Int64Ptr(executionInfo.LastProcessedEvent)
	}
	expectedResponse.ScheduledEventID = common.Int64Ptr(di.ScheduleID)
	expectedResponse.StartedEventID = common.Int64Ptr(di.ScheduleID + 1)
	expectedResponse.StickyExecutionEnabled = common.BoolPtr(false)
	expectedResponse.NextEventID = common.Int64Ptr(msBuilder.GetNextEventID() + 1)
	expectedResponse.Attempt = common.Int64Ptr(di.Attempt)
	expectedResponse.WorkflowExecutionTaskList = &types.TaskList{
		Name: &executionInfo.TaskList,
		Kind: types.TaskListKindNormal.Ptr(),
	}
	expectedResponse.BranchToken, _ = msBuilder.GetCurrentBranchToken()

	response, err := s.historyEngine.RecordDecisionTaskStarted(context.Background(), &request)
	s.Nil(err)
	s.NotNil(response)
	expectedResponse.StartedTimestamp = response.StartedTimestamp
	expectedResponse.ScheduledTimestamp = common.Int64Ptr(0)
	expectedResponse.Queries = make(map[string]*types.WorkflowQuery)
	s.Equal(&expectedResponse, response)
}

func (s *engine2Suite) TestRecordDecisionTaskStartedSuccessStickyEnabled() {
	domainID := constants.TestDomainID
	we := types.WorkflowExecution{
		WorkflowID: common.StringPtr("wId"),
		RunID:      common.StringPtr(constants.TestRunID),
	}
	tl := "testTaskList"
	stickyTl := "stickyTaskList"
	identity := "testIdentity"

	msBuilder := execution.NewMutableStateBuilderWithEventV2(
		s.historyEngine.shard,
		loggerimpl.NewDevelopmentForTest(s.Suite),
		we.GetRunID(),
		constants.TestLocalDomainEntry,
	)
	executionInfo := msBuilder.GetExecutionInfo()
	executionInfo.LastUpdatedTimestamp = time.Now()
	executionInfo.StickyTaskList = stickyTl

	test.AddWorkflowExecutionStartedEvent(msBuilder, we, "wType", tl, []byte("input"), 100, 200, identity)
	di := test.AddDecisionTaskScheduledEvent(msBuilder)

	ms := execution.CreatePersistenceMutableState(msBuilder)

	gwmsResponse := &p.GetWorkflowExecutionResponse{State: ms}

	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse, nil).Once()
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{
		MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{},
	}, nil).Once()

	request := types.RecordDecisionTaskStartedRequest{
		DomainUUID:        common.StringPtr(domainID),
		WorkflowExecution: &we,
		ScheduleID:        common.Int64Ptr(2),
		TaskID:            common.Int64Ptr(100),
		RequestID:         common.StringPtr("reqId"),
		PollRequest: &types.PollForDecisionTaskRequest{
			TaskList: &types.TaskList{
				Name: common.StringPtr(stickyTl),
			},
			Identity: common.StringPtr(identity),
		},
	}

	expectedResponse := types.RecordDecisionTaskStartedResponse{}
	expectedResponse.WorkflowType = msBuilder.GetWorkflowType()
	executionInfo = msBuilder.GetExecutionInfo()
	if executionInfo.LastProcessedEvent != common.EmptyEventID {
		expectedResponse.PreviousStartedEventID = common.Int64Ptr(executionInfo.LastProcessedEvent)
	}
	expectedResponse.ScheduledEventID = common.Int64Ptr(di.ScheduleID)
	expectedResponse.StartedEventID = common.Int64Ptr(di.ScheduleID + 1)
	expectedResponse.StickyExecutionEnabled = common.BoolPtr(true)
	expectedResponse.NextEventID = common.Int64Ptr(msBuilder.GetNextEventID() + 1)
	expectedResponse.Attempt = common.Int64Ptr(di.Attempt)
	expectedResponse.WorkflowExecutionTaskList = &types.TaskList{
		Name: &executionInfo.TaskList,
		Kind: types.TaskListKindNormal.Ptr(),
	}
	currentBranchTokken, err := msBuilder.GetCurrentBranchToken()
	s.NoError(err)
	expectedResponse.BranchToken = currentBranchTokken

	response, err := s.historyEngine.RecordDecisionTaskStarted(context.Background(), &request)
	s.Nil(err)
	s.NotNil(response)
	expectedResponse.StartedTimestamp = response.StartedTimestamp
	expectedResponse.ScheduledTimestamp = common.Int64Ptr(0)
	expectedResponse.Queries = make(map[string]*types.WorkflowQuery)
	s.Equal(&expectedResponse, response)
}

func (s *engine2Suite) TestRecordDecisionTaskStartedIfNoExecution() {
	domainID := constants.TestDomainID
	workflowExecution := &types.WorkflowExecution{
		WorkflowID: common.StringPtr("wId"),
		RunID:      common.StringPtr(constants.TestRunID),
	}

	identity := "testIdentity"
	tl := "testTaskList"

	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(nil, &types.EntityNotExistsError{}).Once()

	response, err := s.historyEngine.RecordDecisionTaskStarted(context.Background(), &types.RecordDecisionTaskStartedRequest{
		DomainUUID:        common.StringPtr(domainID),
		WorkflowExecution: workflowExecution,
		ScheduleID:        common.Int64Ptr(2),
		TaskID:            common.Int64Ptr(100),
		RequestID:         common.StringPtr("reqId"),
		PollRequest: &types.PollForDecisionTaskRequest{
			TaskList: &types.TaskList{
				Name: common.StringPtr(tl),
			},
			Identity: common.StringPtr(identity),
		},
	})
	s.Nil(response)
	s.NotNil(err)
	s.IsType(&types.EntityNotExistsError{}, err)
}

func (s *engine2Suite) TestRecordDecisionTaskStartedIfGetExecutionFailed() {
	domainID := constants.TestDomainID
	workflowExecution := &types.WorkflowExecution{
		WorkflowID: common.StringPtr("wId"),
		RunID:      common.StringPtr(constants.TestRunID),
	}

	identity := "testIdentity"
	tl := "testTaskList"

	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(nil, errors.New("FAILED")).Once()

	response, err := s.historyEngine.RecordDecisionTaskStarted(context.Background(), &types.RecordDecisionTaskStartedRequest{
		DomainUUID:        common.StringPtr(domainID),
		WorkflowExecution: workflowExecution,
		ScheduleID:        common.Int64Ptr(2),
		TaskID:            common.Int64Ptr(100),
		RequestID:         common.StringPtr("reqId"),
		PollRequest: &types.PollForDecisionTaskRequest{
			TaskList: &types.TaskList{
				Name: common.StringPtr(tl),
			},
			Identity: common.StringPtr(identity),
		},
	})
	s.Nil(response)
	s.NotNil(err)
	s.EqualError(err, "FAILED")
}

func (s *engine2Suite) TestRecordDecisionTaskStartedIfTaskAlreadyStarted() {
	domainID := constants.TestDomainID
	workflowExecution := types.WorkflowExecution{
		WorkflowID: common.StringPtr("wId"),
		RunID:      common.StringPtr(constants.TestRunID),
	}

	identity := "testIdentity"
	tl := "testTaskList"

	msBuilder := s.createExecutionStartedState(workflowExecution, tl, identity, true)
	ms := execution.CreatePersistenceMutableState(msBuilder)
	gwmsResponse := &p.GetWorkflowExecutionResponse{State: ms}
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse, nil).Once()

	response, err := s.historyEngine.RecordDecisionTaskStarted(context.Background(), &types.RecordDecisionTaskStartedRequest{
		DomainUUID:        common.StringPtr(domainID),
		WorkflowExecution: &workflowExecution,
		ScheduleID:        common.Int64Ptr(2),
		TaskID:            common.Int64Ptr(100),
		RequestID:         common.StringPtr("reqId"),
		PollRequest: &types.PollForDecisionTaskRequest{
			TaskList: &types.TaskList{
				Name: common.StringPtr(tl),
			},
			Identity: common.StringPtr(identity),
		},
	})
	s.Nil(response)
	s.NotNil(err)
	s.IsType(&types.EventAlreadyStartedError{}, err)
	s.logger.Error("RecordDecisionTaskStarted failed with", tag.Error(err))
}

func (s *engine2Suite) TestRecordDecisionTaskStartedIfTaskAlreadyCompleted() {
	domainID := constants.TestDomainID
	workflowExecution := types.WorkflowExecution{
		WorkflowID: common.StringPtr("wId"),
		RunID:      common.StringPtr(constants.TestRunID),
	}

	identity := "testIdentity"
	tl := "testTaskList"

	msBuilder := s.createExecutionStartedState(workflowExecution, tl, identity, true)
	test.AddDecisionTaskCompletedEvent(msBuilder, int64(2), int64(3), nil, identity)

	ms := execution.CreatePersistenceMutableState(msBuilder)
	gwmsResponse := &p.GetWorkflowExecutionResponse{State: ms}

	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse, nil).Once()

	response, err := s.historyEngine.RecordDecisionTaskStarted(context.Background(), &types.RecordDecisionTaskStartedRequest{
		DomainUUID:        common.StringPtr(domainID),
		WorkflowExecution: &workflowExecution,
		ScheduleID:        common.Int64Ptr(2),
		TaskID:            common.Int64Ptr(100),
		RequestID:         common.StringPtr("reqId"),
		PollRequest: &types.PollForDecisionTaskRequest{
			TaskList: &types.TaskList{
				Name: common.StringPtr(tl),
			},
			Identity: common.StringPtr(identity),
		},
	})
	s.Nil(response)
	s.NotNil(err)
	s.IsType(&types.EntityNotExistsError{}, err)
	s.logger.Error("RecordDecisionTaskStarted failed with", tag.Error(err))
}

func (s *engine2Suite) TestRecordDecisionTaskStartedConflictOnUpdate() {
	domainID := constants.TestDomainID
	workflowExecution := types.WorkflowExecution{
		WorkflowID: common.StringPtr("wId"),
		RunID:      common.StringPtr(constants.TestRunID),
	}

	identity := "testIdentity"
	tl := "testTaskList"

	msBuilder := s.createExecutionStartedState(workflowExecution, tl, identity, false)

	ms := execution.CreatePersistenceMutableState(msBuilder)
	gwmsResponse := &p.GetWorkflowExecutionResponse{State: ms}

	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse, nil).Once()
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(nil, &p.ConditionFailedError{}).Once()

	ms2 := execution.CreatePersistenceMutableState(msBuilder)
	gwmsResponse2 := &p.GetWorkflowExecutionResponse{State: ms2}

	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse2, nil).Once()
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{
		MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{},
	}, nil).Once()

	response, err := s.historyEngine.RecordDecisionTaskStarted(context.Background(), &types.RecordDecisionTaskStartedRequest{
		DomainUUID:        common.StringPtr(domainID),
		WorkflowExecution: &workflowExecution,
		ScheduleID:        common.Int64Ptr(2),
		TaskID:            common.Int64Ptr(100),
		RequestID:         common.StringPtr("reqId"),
		PollRequest: &types.PollForDecisionTaskRequest{
			TaskList: &types.TaskList{
				Name: common.StringPtr(tl),
			},
			Identity: common.StringPtr(identity),
		},
	})
	s.Nil(err)
	s.NotNil(response)
	s.Equal("wType", *response.WorkflowType.Name)
	s.True(response.PreviousStartedEventID == nil)
	s.Equal(int64(3), *response.StartedEventID)
}

func (s *engine2Suite) TestRecordDecisionTaskRetrySameRequest() {
	domainID := constants.TestDomainID
	workflowExecution := types.WorkflowExecution{
		WorkflowID: common.StringPtr("wId"),
		RunID:      common.StringPtr(constants.TestRunID),
	}

	tl := "testTaskList"
	identity := "testIdentity"
	requestID := "testRecordDecisionTaskRetrySameRequestID"

	msBuilder := s.createExecutionStartedState(workflowExecution, tl, identity, false)
	ms := execution.CreatePersistenceMutableState(msBuilder)
	gwmsResponse := &p.GetWorkflowExecutionResponse{State: ms}

	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse, nil).Once()
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(nil, &p.ConditionFailedError{}).Once()

	startedEventID := test.AddDecisionTaskStartedEventWithRequestID(msBuilder, int64(2), requestID, tl, identity)
	ms2 := execution.CreatePersistenceMutableState(msBuilder)
	gwmsResponse2 := &p.GetWorkflowExecutionResponse{State: ms2}
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse2, nil).Once()

	response, err := s.historyEngine.RecordDecisionTaskStarted(context.Background(), &types.RecordDecisionTaskStartedRequest{
		DomainUUID:        common.StringPtr(domainID),
		WorkflowExecution: &workflowExecution,
		ScheduleID:        common.Int64Ptr(2),
		TaskID:            common.Int64Ptr(100),
		RequestID:         common.StringPtr(requestID),
		PollRequest: &types.PollForDecisionTaskRequest{
			TaskList: &types.TaskList{
				Name: common.StringPtr(tl),
			},
			Identity: common.StringPtr(identity),
		},
	})

	s.Nil(err)
	s.NotNil(response)
	s.Equal("wType", *response.WorkflowType.Name)
	s.True(response.PreviousStartedEventID == nil)
	s.Equal(startedEventID.EventID, response.StartedEventID)
}

func (s *engine2Suite) TestRecordDecisionTaskRetryDifferentRequest() {
	domainID := constants.TestDomainID
	workflowExecution := types.WorkflowExecution{
		WorkflowID: common.StringPtr("wId"),
		RunID:      common.StringPtr(constants.TestRunID),
	}

	tl := "testTaskList"
	identity := "testIdentity"
	requestID := "testRecordDecisionTaskRetrySameRequestID"

	msBuilder := s.createExecutionStartedState(workflowExecution, tl, identity, false)
	ms := execution.CreatePersistenceMutableState(msBuilder)
	gwmsResponse := &p.GetWorkflowExecutionResponse{State: ms}
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse, nil).Once()
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(nil, &p.ConditionFailedError{}).Once()

	// Add event.
	test.AddDecisionTaskStartedEventWithRequestID(msBuilder, int64(2), "some_other_req", tl, identity)
	ms2 := execution.CreatePersistenceMutableState(msBuilder)
	gwmsResponse2 := &p.GetWorkflowExecutionResponse{State: ms2}
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse2, nil).Once()

	response, err := s.historyEngine.RecordDecisionTaskStarted(context.Background(), &types.RecordDecisionTaskStartedRequest{
		DomainUUID:        common.StringPtr(domainID),
		WorkflowExecution: &workflowExecution,
		ScheduleID:        common.Int64Ptr(2),
		TaskID:            common.Int64Ptr(100),
		RequestID:         common.StringPtr(requestID),
		PollRequest: &types.PollForDecisionTaskRequest{
			TaskList: &types.TaskList{
				Name: common.StringPtr(tl),
			},
			Identity: common.StringPtr(identity),
		},
	})

	s.Nil(response)
	s.NotNil(err)
	s.IsType(&types.EventAlreadyStartedError{}, err)
	s.logger.Info("Failed with error", tag.Error(err))
}

func (s *engine2Suite) TestRecordDecisionTaskStartedMaxAttemptsExceeded() {
	domainID := constants.TestDomainID
	workflowExecution := types.WorkflowExecution{
		WorkflowID: common.StringPtr("wId"),
		RunID:      common.StringPtr(constants.TestRunID),
	}

	tl := "testTaskList"
	identity := "testIdentity"

	msBuilder := s.createExecutionStartedState(workflowExecution, tl, identity, false)
	for i := 0; i < conditionalRetryCount; i++ {
		ms := execution.CreatePersistenceMutableState(msBuilder)
		gwmsResponse := &p.GetWorkflowExecutionResponse{State: ms}

		s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse, nil).Once()
	}

	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Times(
		conditionalRetryCount)
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(nil,
		&p.ConditionFailedError{}).Times(conditionalRetryCount)

	response, err := s.historyEngine.RecordDecisionTaskStarted(context.Background(), &types.RecordDecisionTaskStartedRequest{
		DomainUUID:        common.StringPtr(domainID),
		WorkflowExecution: &workflowExecution,
		ScheduleID:        common.Int64Ptr(2),
		TaskID:            common.Int64Ptr(100),
		RequestID:         common.StringPtr("reqId"),
		PollRequest: &types.PollForDecisionTaskRequest{
			TaskList: &types.TaskList{
				Name: common.StringPtr(tl),
			},
			Identity: common.StringPtr(identity),
		},
	})

	s.NotNil(err)
	s.Nil(response)
	s.Equal(ErrMaxAttemptsExceeded, err)
}

func (s *engine2Suite) TestRecordDecisionTaskSuccess() {
	domainID := constants.TestDomainID
	workflowExecution := types.WorkflowExecution{
		WorkflowID: common.StringPtr("wId"),
		RunID:      common.StringPtr(constants.TestRunID),
	}

	tl := "testTaskList"
	identity := "testIdentity"

	msBuilder := s.createExecutionStartedState(workflowExecution, tl, identity, false)
	ms := execution.CreatePersistenceMutableState(msBuilder)
	gwmsResponse := &p.GetWorkflowExecutionResponse{State: ms}
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse, nil).Once()
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{
		MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{},
	}, nil).Once()

	// load mutable state such that it already exists in memory when respond decision task is called
	// this enables us to set query registry on it
	ctx, release, err := s.historyEngine.executionCache.GetOrCreateWorkflowExecutionForBackground(constants.TestDomainID, workflowExecution)
	s.NoError(err)
	loadedMS, err := ctx.LoadWorkflowExecution(context.Background())
	s.NoError(err)
	qr := query.NewRegistry()
	id1, _ := qr.BufferQuery(&types.WorkflowQuery{})
	id2, _ := qr.BufferQuery(&types.WorkflowQuery{})
	id3, _ := qr.BufferQuery(&types.WorkflowQuery{})
	loadedMS.SetQueryRegistry(qr)
	release(nil)

	response, err := s.historyEngine.RecordDecisionTaskStarted(context.Background(), &types.RecordDecisionTaskStartedRequest{
		DomainUUID:        common.StringPtr(domainID),
		WorkflowExecution: &workflowExecution,
		ScheduleID:        common.Int64Ptr(2),
		TaskID:            common.Int64Ptr(100),
		RequestID:         common.StringPtr("reqId"),
		PollRequest: &types.PollForDecisionTaskRequest{
			TaskList: &types.TaskList{
				Name: common.StringPtr(tl),
			},
			Identity: common.StringPtr(identity),
		},
	})

	s.Nil(err)
	s.NotNil(response)
	s.Equal("wType", *response.WorkflowType.Name)
	s.True(response.PreviousStartedEventID == nil)
	s.Equal(int64(3), *response.StartedEventID)
	expectedQueryMap := map[string]*types.WorkflowQuery{
		id1: {},
		id2: {},
		id3: {},
	}
	s.Equal(expectedQueryMap, response.Queries)
}

func (s *engine2Suite) TestRecordActivityTaskStartedIfNoExecution() {
	domainID := constants.TestDomainID
	workflowExecution := &types.WorkflowExecution{
		WorkflowID: common.StringPtr("wId"),
		RunID:      common.StringPtr(constants.TestRunID),
	}

	identity := "testIdentity"
	tl := "testTaskList"

	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(nil, &types.EntityNotExistsError{}).Once()

	response, err := s.historyEngine.RecordActivityTaskStarted(context.Background(), &types.RecordActivityTaskStartedRequest{
		DomainUUID:        common.StringPtr(domainID),
		WorkflowExecution: workflowExecution,
		ScheduleID:        common.Int64Ptr(5),
		TaskID:            common.Int64Ptr(100),
		RequestID:         common.StringPtr("reqId"),
		PollRequest: &types.PollForActivityTaskRequest{
			TaskList: &types.TaskList{
				Name: common.StringPtr(tl),
			},
			Identity: common.StringPtr(identity),
		},
	})
	if err != nil {
		s.logger.Error("Unexpected Error", tag.Error(err))
	}
	s.Nil(response)
	s.NotNil(err)
	s.IsType(&types.EntityNotExistsError{}, err)
}

func (s *engine2Suite) TestRecordActivityTaskStartedSuccess() {
	domainID := constants.TestDomainID
	workflowExecution := types.WorkflowExecution{
		WorkflowID: common.StringPtr("wId"),
		RunID:      common.StringPtr(constants.TestRunID),
	}

	identity := "testIdentity"
	tl := "testTaskList"

	activityID := "activity1_id"
	activityType := "activity_type1"
	activityInput := []byte("input1")

	msBuilder := s.createExecutionStartedState(workflowExecution, tl, identity, true)
	decisionCompletedEvent := test.AddDecisionTaskCompletedEvent(msBuilder, int64(2), int64(3), nil, identity)
	scheduledEvent, _ := test.AddActivityTaskScheduledEvent(msBuilder, *decisionCompletedEvent.EventID, activityID,
		activityType, tl, activityInput, 100, 10, 1, 5)

	ms1 := execution.CreatePersistenceMutableState(msBuilder)
	gwmsResponse1 := &p.GetWorkflowExecutionResponse{State: ms1}

	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse1, nil).Once()
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{
		MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{},
	}, nil).Once()

	s.mockEventsCache.EXPECT().GetEvent(
		gomock.Any(), gomock.Any(), domainID, workflowExecution.GetWorkflowID(), workflowExecution.GetRunID(),
		decisionCompletedEvent.GetEventID(), scheduledEvent.GetEventID(), gomock.Any(),
	).Return(scheduledEvent, nil)
	response, err := s.historyEngine.RecordActivityTaskStarted(context.Background(), &types.RecordActivityTaskStartedRequest{
		DomainUUID:        common.StringPtr(domainID),
		WorkflowExecution: &workflowExecution,
		ScheduleID:        common.Int64Ptr(5),
		TaskID:            common.Int64Ptr(100),
		RequestID:         common.StringPtr("reqId"),
		PollRequest: &types.PollForActivityTaskRequest{
			TaskList: &types.TaskList{
				Name: common.StringPtr(tl),
			},
			Identity: common.StringPtr(identity),
		},
	})
	s.Nil(err)
	s.NotNil(response)
	s.Equal(scheduledEvent, response.ScheduledEvent)
}

func (s *engine2Suite) TestRequestCancelWorkflowExecutionSuccess() {
	domainID := constants.TestDomainID
	workflowExecution := types.WorkflowExecution{
		WorkflowID: common.StringPtr("wId"),
		RunID:      common.StringPtr(constants.TestRunID),
	}

	identity := "testIdentity"
	tl := "testTaskList"

	msBuilder := s.createExecutionStartedState(workflowExecution, tl, identity, false)
	ms1 := execution.CreatePersistenceMutableState(msBuilder)
	gwmsResponse1 := &p.GetWorkflowExecutionResponse{State: ms1}

	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse1, nil).Once()
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{
		MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{},
	}, nil).Once()

	err := s.historyEngine.RequestCancelWorkflowExecution(context.Background(), &types.HistoryRequestCancelWorkflowExecutionRequest{
		DomainUUID: common.StringPtr(domainID),
		CancelRequest: &types.RequestCancelWorkflowExecutionRequest{
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: workflowExecution.WorkflowID,
				RunID:      workflowExecution.RunID,
			},
			Identity: common.StringPtr("identity"),
		},
	})
	s.Nil(err)

	executionBuilder := s.getBuilder(domainID, workflowExecution)
	s.Equal(int64(4), executionBuilder.GetNextEventID())
}

func (s *engine2Suite) TestRequestCancelWorkflowExecutionFail() {
	domainID := constants.TestDomainID
	workflowExecution := types.WorkflowExecution{
		WorkflowID: common.StringPtr("wId"),
		RunID:      common.StringPtr(constants.TestRunID),
	}

	identity := "testIdentity"
	tl := "testTaskList"

	msBuilder := s.createExecutionStartedState(workflowExecution, tl, identity, false)
	msBuilder.GetExecutionInfo().State = p.WorkflowStateCompleted
	ms1 := execution.CreatePersistenceMutableState(msBuilder)
	gwmsResponse1 := &p.GetWorkflowExecutionResponse{State: ms1}

	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse1, nil).Once()

	err := s.historyEngine.RequestCancelWorkflowExecution(context.Background(), &types.HistoryRequestCancelWorkflowExecutionRequest{
		DomainUUID: common.StringPtr(domainID),
		CancelRequest: &types.RequestCancelWorkflowExecutionRequest{
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: workflowExecution.WorkflowID,
				RunID:      workflowExecution.RunID,
			},
			Identity: common.StringPtr("identity"),
		},
	})
	s.NotNil(err)
	s.IsType(&types.EntityNotExistsError{}, err)
}

func (s *engine2Suite) createExecutionStartedState(we types.WorkflowExecution, tl, identity string,
	startDecision bool) execution.MutableState {
	msBuilder := execution.NewMutableStateBuilderWithEventV2(
		s.historyEngine.shard,
		s.logger,
		we.GetRunID(),
		constants.TestLocalDomainEntry,
	)
	test.AddWorkflowExecutionStartedEvent(msBuilder, we, "wType", tl, []byte("input"), 100, 200, identity)
	di := test.AddDecisionTaskScheduledEvent(msBuilder)
	if startDecision {
		test.AddDecisionTaskStartedEvent(msBuilder, di.ScheduleID, tl, identity)
	}
	_ = msBuilder.SetHistoryTree(we.GetRunID())

	return msBuilder
}

//nolint:unused
func (s *engine2Suite) printHistory(builder execution.MutableState) string {
	return thrift.FromHistory(builder.GetHistoryBuilder().GetHistory()).String()
}

func (s *engine2Suite) TestRespondDecisionTaskCompletedRecordMarkerDecision() {
	domainID := constants.TestDomainID
	we := types.WorkflowExecution{
		WorkflowID: common.StringPtr("wId"),
		RunID:      common.StringPtr(constants.TestRunID),
	}
	tl := "testTaskList"
	taskToken, _ := json.Marshal(&common.TaskToken{
		WorkflowID: "wId",
		RunID:      we.GetRunID(),
		ScheduleID: 2,
	})
	identity := "testIdentity"
	markerDetails := []byte("marker details")
	markerName := "marker name"

	msBuilder := execution.NewMutableStateBuilderWithEventV2(
		s.historyEngine.shard,
		loggerimpl.NewDevelopmentForTest(s.Suite),
		we.GetRunID(),
		constants.TestLocalDomainEntry,
	)
	test.AddWorkflowExecutionStartedEvent(msBuilder, we, "wType", tl, []byte("input"), 100, 200, identity)
	di := test.AddDecisionTaskScheduledEvent(msBuilder)
	test.AddDecisionTaskStartedEvent(msBuilder, di.ScheduleID, tl, identity)

	decisions := []*types.Decision{{
		DecisionType: types.DecisionTypeRecordMarker.Ptr(),
		RecordMarkerDecisionAttributes: &types.RecordMarkerDecisionAttributes{
			MarkerName: common.StringPtr(markerName),
			Details:    markerDetails,
		},
	}}

	ms := execution.CreatePersistenceMutableState(msBuilder)
	gwmsResponse := &p.GetWorkflowExecutionResponse{State: ms}

	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse, nil).Once()
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{
		MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{},
	}, nil).Once()

	_, err := s.historyEngine.RespondDecisionTaskCompleted(context.Background(), &types.HistoryRespondDecisionTaskCompletedRequest{
		DomainUUID: common.StringPtr(domainID),
		CompleteRequest: &types.RespondDecisionTaskCompletedRequest{
			TaskToken:        taskToken,
			Decisions:        decisions,
			ExecutionContext: nil,
			Identity:         &identity,
		},
	})
	s.Nil(err)
	executionBuilder := s.getBuilder(domainID, we)
	s.Equal(int64(6), executionBuilder.GetExecutionInfo().NextEventID)
	s.Equal(int64(3), executionBuilder.GetExecutionInfo().LastProcessedEvent)
	s.Equal(p.WorkflowStateRunning, executionBuilder.GetExecutionInfo().State)
	s.False(executionBuilder.HasPendingDecision())
}

func (s *engine2Suite) TestStartWorkflowExecution_BrandNew() {
	domainID := constants.TestDomainID
	workflowID := "workflowID"
	workflowType := "workflowType"
	taskList := "testTaskList"
	identity := "testIdentity"

	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("CreateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.CreateWorkflowExecutionResponse{}, nil).Once()

	requestID := uuid.New()
	resp, err := s.historyEngine.StartWorkflowExecution(context.Background(), &types.HistoryStartWorkflowExecutionRequest{
		DomainUUID: common.StringPtr(domainID),
		StartRequest: &types.StartWorkflowExecutionRequest{
			Domain:                              common.StringPtr(domainID),
			WorkflowID:                          common.StringPtr(workflowID),
			WorkflowType:                        &types.WorkflowType{Name: common.StringPtr(workflowType)},
			TaskList:                            &types.TaskList{Name: common.StringPtr(taskList)},
			ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(1),
			TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(2),
			Identity:                            common.StringPtr(identity),
			RequestID:                           common.StringPtr(requestID),
		},
	})
	s.Nil(err)
	s.NotNil(resp.RunID)
}

func (s *engine2Suite) TestStartWorkflowExecution_StillRunning_Dedup() {
	domainID := constants.TestDomainID
	workflowID := "workflowID"
	runID := "runID"
	workflowType := "workflowType"
	taskList := "testTaskList"
	identity := "testIdentity"
	requestID := "requestID"
	lastWriteVersion := common.EmptyVersion

	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("CreateWorkflowExecution", mock.Anything, mock.Anything).Return(nil, &p.WorkflowExecutionAlreadyStartedError{
		Msg:              "random message",
		StartRequestID:   requestID,
		RunID:            runID,
		State:            p.WorkflowStateRunning,
		CloseStatus:      p.WorkflowCloseStatusNone,
		LastWriteVersion: lastWriteVersion,
	}).Once()

	resp, err := s.historyEngine.StartWorkflowExecution(context.Background(), &types.HistoryStartWorkflowExecutionRequest{
		DomainUUID: common.StringPtr(domainID),
		StartRequest: &types.StartWorkflowExecutionRequest{
			Domain:                              common.StringPtr(domainID),
			WorkflowID:                          common.StringPtr(workflowID),
			WorkflowType:                        &types.WorkflowType{Name: common.StringPtr(workflowType)},
			TaskList:                            &types.TaskList{Name: common.StringPtr(taskList)},
			ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(1),
			TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(2),
			Identity:                            common.StringPtr(identity),
			RequestID:                           common.StringPtr(requestID),
		},
	})
	s.Nil(err)
	s.Equal(runID, resp.GetRunID())
}

func (s *engine2Suite) TestStartWorkflowExecution_StillRunning_NonDeDup() {
	domainID := constants.TestDomainID
	workflowID := "workflowID"
	runID := "runID"
	workflowType := "workflowType"
	taskList := "testTaskList"
	identity := "testIdentity"
	lastWriteVersion := common.EmptyVersion

	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("CreateWorkflowExecution", mock.Anything, mock.Anything).Return(nil, &p.WorkflowExecutionAlreadyStartedError{
		Msg:              "random message",
		StartRequestID:   "oldRequestID",
		RunID:            runID,
		State:            p.WorkflowStateRunning,
		CloseStatus:      p.WorkflowCloseStatusNone,
		LastWriteVersion: lastWriteVersion,
	}).Once()

	resp, err := s.historyEngine.StartWorkflowExecution(context.Background(), &types.HistoryStartWorkflowExecutionRequest{
		DomainUUID: common.StringPtr(domainID),
		StartRequest: &types.StartWorkflowExecutionRequest{
			Domain:                              common.StringPtr(domainID),
			WorkflowID:                          common.StringPtr(workflowID),
			WorkflowType:                        &types.WorkflowType{Name: common.StringPtr(workflowType)},
			TaskList:                            &types.TaskList{Name: common.StringPtr(taskList)},
			ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(1),
			TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(2),
			Identity:                            common.StringPtr(identity),
			RequestID:                           common.StringPtr("newRequestID"),
		},
	})
	if _, ok := err.(*types.WorkflowExecutionAlreadyStartedError); !ok {
		s.Fail("return err is not *types.WorkflowExecutionAlreadyStartedError")
	}
	s.Nil(resp)
}

func (s *engine2Suite) TestStartWorkflowExecution_NotRunning_PrevSuccess() {
	domainID := constants.TestDomainID
	workflowID := "workflowID"
	runID := "runID"
	workflowType := "workflowType"
	taskList := "testTaskList"
	identity := "testIdentity"
	lastWriteVersion := common.EmptyVersion

	options := []types.WorkflowIDReusePolicy{
		types.WorkflowIDReusePolicyAllowDuplicateFailedOnly,
		types.WorkflowIDReusePolicyAllowDuplicate,
		types.WorkflowIDReusePolicyRejectDuplicate,
	}

	expecedErrs := []bool{true, false, true}

	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Times(len(expecedErrs))
	s.mockExecutionMgr.On(
		"CreateWorkflowExecution",
		mock.Anything,
		mock.MatchedBy(func(request *p.CreateWorkflowExecutionRequest) bool {
			return request.Mode == p.CreateWorkflowModeBrandNew
		}),
	).Return(nil, &p.WorkflowExecutionAlreadyStartedError{
		Msg:              "random message",
		StartRequestID:   "oldRequestID",
		RunID:            runID,
		State:            p.WorkflowStateCompleted,
		CloseStatus:      p.WorkflowCloseStatusCompleted,
		LastWriteVersion: lastWriteVersion,
	}).Times(len(expecedErrs))

	for index, option := range options {
		if !expecedErrs[index] {
			s.mockExecutionMgr.On(
				"CreateWorkflowExecution",
				mock.Anything,
				mock.MatchedBy(func(request *p.CreateWorkflowExecutionRequest) bool {
					return request.Mode == p.CreateWorkflowModeWorkflowIDReuse &&
						request.PreviousRunID == runID &&
						request.PreviousLastWriteVersion == lastWriteVersion
				}),
			).Return(&p.CreateWorkflowExecutionResponse{}, nil).Once()
		}

		resp, err := s.historyEngine.StartWorkflowExecution(context.Background(), &types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: common.StringPtr(domainID),
			StartRequest: &types.StartWorkflowExecutionRequest{
				Domain:                              common.StringPtr(domainID),
				WorkflowID:                          common.StringPtr(workflowID),
				WorkflowType:                        &types.WorkflowType{Name: common.StringPtr(workflowType)},
				TaskList:                            &types.TaskList{Name: common.StringPtr(taskList)},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(1),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(2),
				Identity:                            common.StringPtr(identity),
				RequestID:                           common.StringPtr("newRequestID"),
				WorkflowIDReusePolicy:               &option,
			},
		})

		if expecedErrs[index] {
			if _, ok := err.(*types.WorkflowExecutionAlreadyStartedError); !ok {
				s.Fail("return err is not *types.WorkflowExecutionAlreadyStartedError")
			}
			s.Nil(resp)
		} else {
			s.Nil(err)
			s.NotNil(resp)
		}
	}
}

func (s *engine2Suite) TestStartWorkflowExecution_NotRunning_PrevFail() {
	domainID := constants.TestDomainID
	workflowID := "workflowID"
	workflowType := "workflowType"
	taskList := "testTaskList"
	identity := "testIdentity"
	lastWriteVersion := common.EmptyVersion

	options := []types.WorkflowIDReusePolicy{
		types.WorkflowIDReusePolicyAllowDuplicateFailedOnly,
		types.WorkflowIDReusePolicyAllowDuplicate,
		types.WorkflowIDReusePolicyRejectDuplicate,
	}

	expecedErrs := []bool{false, false, true}

	closeStates := []int{
		p.WorkflowCloseStatusFailed,
		p.WorkflowCloseStatusCanceled,
		p.WorkflowCloseStatusTerminated,
		p.WorkflowCloseStatusTimedOut,
	}
	runIDs := []string{"1", "2", "3", "4"}

	for i, closeState := range closeStates {

		s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Times(len(expecedErrs))
		s.mockExecutionMgr.On(
			"CreateWorkflowExecution",
			mock.Anything,
			mock.MatchedBy(func(request *p.CreateWorkflowExecutionRequest) bool {
				return request.Mode == p.CreateWorkflowModeBrandNew
			}),
		).Return(nil, &p.WorkflowExecutionAlreadyStartedError{
			Msg:              "random message",
			StartRequestID:   "oldRequestID",
			RunID:            runIDs[i],
			State:            p.WorkflowStateCompleted,
			CloseStatus:      closeState,
			LastWriteVersion: lastWriteVersion,
		}).Times(len(expecedErrs))

		for j, option := range options {

			if !expecedErrs[j] {
				s.mockExecutionMgr.On(
					"CreateWorkflowExecution",
					mock.Anything,
					mock.MatchedBy(func(request *p.CreateWorkflowExecutionRequest) bool {
						return request.Mode == p.CreateWorkflowModeWorkflowIDReuse &&
							request.PreviousRunID == runIDs[i] &&
							request.PreviousLastWriteVersion == lastWriteVersion
					}),
				).Return(&p.CreateWorkflowExecutionResponse{}, nil).Once()
			}

			resp, err := s.historyEngine.StartWorkflowExecution(context.Background(), &types.HistoryStartWorkflowExecutionRequest{
				DomainUUID: common.StringPtr(domainID),
				StartRequest: &types.StartWorkflowExecutionRequest{
					Domain:                              common.StringPtr(domainID),
					WorkflowID:                          common.StringPtr(workflowID),
					WorkflowType:                        &types.WorkflowType{Name: common.StringPtr(workflowType)},
					TaskList:                            &types.TaskList{Name: common.StringPtr(taskList)},
					ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(1),
					TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(2),
					Identity:                            common.StringPtr(identity),
					RequestID:                           common.StringPtr("newRequestID"),
					WorkflowIDReusePolicy:               &option,
				},
			})

			if expecedErrs[j] {
				if _, ok := err.(*types.WorkflowExecutionAlreadyStartedError); !ok {
					s.Fail("return err is not *types.WorkflowExecutionAlreadyStartedError")
				}
				s.Nil(resp)
			} else {
				s.Nil(err)
				s.NotNil(resp)
			}
		}
	}
}

func (s *engine2Suite) TestSignalWithStartWorkflowExecution_JustSignal() {
	sRequest := &types.HistorySignalWithStartWorkflowExecutionRequest{}
	_, err := s.historyEngine.SignalWithStartWorkflowExecution(context.Background(), sRequest)
	s.EqualError(err, "BadRequestError{Message: Missing domain UUID.}")

	domainID := constants.TestDomainID
	workflowID := "wId"
	runID := constants.TestRunID
	identity := "testIdentity"
	signalName := "my signal name"
	input := []byte("test input")
	sRequest = &types.HistorySignalWithStartWorkflowExecutionRequest{
		DomainUUID: common.StringPtr(domainID),
		SignalWithStartRequest: &types.SignalWithStartWorkflowExecutionRequest{
			Domain:     common.StringPtr(domainID),
			WorkflowID: common.StringPtr(workflowID),
			Identity:   common.StringPtr(identity),
			SignalName: common.StringPtr(signalName),
			Input:      input,
		},
	}

	msBuilder := execution.NewMutableStateBuilderWithEventV2(
		s.historyEngine.shard,
		loggerimpl.NewDevelopmentForTest(s.Suite),
		runID,
		constants.TestLocalDomainEntry,
	)
	ms := execution.CreatePersistenceMutableState(msBuilder)
	gwmsResponse := &p.GetWorkflowExecutionResponse{State: ms}
	gceResponse := &p.GetCurrentExecutionResponse{RunID: runID}

	s.mockExecutionMgr.On("GetCurrentExecution", mock.Anything, mock.Anything).Return(gceResponse, nil).Once()
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse, nil).Once()
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("UpdateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.UpdateWorkflowExecutionResponse{
		MutableStateUpdateSessionStats: &p.MutableStateUpdateSessionStats{},
	}, nil).Once()

	resp, err := s.historyEngine.SignalWithStartWorkflowExecution(context.Background(), sRequest)
	s.Nil(err)
	s.Equal(runID, resp.GetRunID())
}

func (s *engine2Suite) TestSignalWithStartWorkflowExecution_WorkflowNotExist() {
	sRequest := &types.HistorySignalWithStartWorkflowExecutionRequest{}
	_, err := s.historyEngine.SignalWithStartWorkflowExecution(context.Background(), sRequest)
	s.EqualError(err, "BadRequestError{Message: Missing domain UUID.}")

	domainID := constants.TestDomainID
	workflowID := "wId"
	workflowType := "workflowType"
	taskList := "testTaskList"
	identity := "testIdentity"
	signalName := "my signal name"
	input := []byte("test input")
	requestID := uuid.New()

	sRequest = &types.HistorySignalWithStartWorkflowExecutionRequest{
		DomainUUID: common.StringPtr(domainID),
		SignalWithStartRequest: &types.SignalWithStartWorkflowExecutionRequest{
			Domain:                              common.StringPtr(domainID),
			WorkflowID:                          common.StringPtr(workflowID),
			WorkflowType:                        &types.WorkflowType{Name: common.StringPtr(workflowType)},
			TaskList:                            &types.TaskList{Name: common.StringPtr(taskList)},
			ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(1),
			TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(2),
			Identity:                            common.StringPtr(identity),
			SignalName:                          common.StringPtr(signalName),
			Input:                               input,
			RequestID:                           common.StringPtr(requestID),
		},
	}

	notExistErr := &types.EntityNotExistsError{Message: "Workflow not exist"}

	s.mockExecutionMgr.On("GetCurrentExecution", mock.Anything, mock.Anything).Return(nil, notExistErr).Once()
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("CreateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.CreateWorkflowExecutionResponse{}, nil).Once()

	resp, err := s.historyEngine.SignalWithStartWorkflowExecution(context.Background(), sRequest)
	s.Nil(err)
	s.NotNil(resp.GetRunID())
}

func (s *engine2Suite) TestSignalWithStartWorkflowExecution_CreateTimeout() {
	sRequest := &types.HistorySignalWithStartWorkflowExecutionRequest{}
	_, err := s.historyEngine.SignalWithStartWorkflowExecution(context.Background(), sRequest)
	s.EqualError(err, "BadRequestError{Message: Missing domain UUID.}")

	domainID := constants.TestDomainID
	workflowID := "wId"
	workflowType := "workflowType"
	taskList := "testTaskList"
	identity := "testIdentity"
	signalName := "my signal name"
	input := []byte("test input")
	requestID := uuid.New()

	sRequest = &types.HistorySignalWithStartWorkflowExecutionRequest{
		DomainUUID: common.StringPtr(domainID),
		SignalWithStartRequest: &types.SignalWithStartWorkflowExecutionRequest{
			Domain:                              common.StringPtr(domainID),
			WorkflowID:                          common.StringPtr(workflowID),
			WorkflowType:                        &types.WorkflowType{Name: common.StringPtr(workflowType)},
			TaskList:                            &types.TaskList{Name: common.StringPtr(taskList)},
			ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(1),
			TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(2),
			Identity:                            common.StringPtr(identity),
			SignalName:                          common.StringPtr(signalName),
			Input:                               input,
			RequestID:                           common.StringPtr(requestID),
		},
	}

	notExistErr := &types.EntityNotExistsError{Message: "Workflow not exist"}

	s.mockExecutionMgr.On("GetCurrentExecution", mock.Anything, mock.Anything).Return(nil, notExistErr).Once()
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("CreateWorkflowExecution", mock.Anything, mock.Anything).Return(nil, &p.TimeoutError{}).Once()

	resp, err := s.historyEngine.SignalWithStartWorkflowExecution(context.Background(), sRequest)
	s.True(p.IsTimeoutError(err))
	s.NotNil(resp.GetRunID())
}

func (s *engine2Suite) TestSignalWithStartWorkflowExecution_WorkflowNotRunning() {
	sRequest := &types.HistorySignalWithStartWorkflowExecutionRequest{}
	_, err := s.historyEngine.SignalWithStartWorkflowExecution(context.Background(), sRequest)
	s.EqualError(err, "BadRequestError{Message: Missing domain UUID.}")

	domainID := constants.TestDomainID
	workflowID := "wId"
	runID := constants.TestRunID
	workflowType := "workflowType"
	taskList := "testTaskList"
	identity := "testIdentity"
	signalName := "my signal name"
	input := []byte("test input")
	requestID := uuid.New()
	policy := types.WorkflowIDReusePolicyAllowDuplicate
	sRequest = &types.HistorySignalWithStartWorkflowExecutionRequest{
		DomainUUID: common.StringPtr(domainID),
		SignalWithStartRequest: &types.SignalWithStartWorkflowExecutionRequest{
			Domain:                              common.StringPtr(domainID),
			WorkflowID:                          common.StringPtr(workflowID),
			WorkflowType:                        &types.WorkflowType{Name: common.StringPtr(workflowType)},
			TaskList:                            &types.TaskList{Name: common.StringPtr(taskList)},
			ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(1),
			TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(2),
			Identity:                            common.StringPtr(identity),
			SignalName:                          common.StringPtr(signalName),
			Input:                               input,
			RequestID:                           common.StringPtr(requestID),
			WorkflowIDReusePolicy:               &policy,
		},
	}

	msBuilder := execution.NewMutableStateBuilderWithEventV2(
		s.historyEngine.shard,
		loggerimpl.NewDevelopmentForTest(s.Suite),
		runID,
		constants.TestLocalDomainEntry,
	)
	ms := execution.CreatePersistenceMutableState(msBuilder)
	ms.ExecutionInfo.State = p.WorkflowStateCompleted
	gwmsResponse := &p.GetWorkflowExecutionResponse{State: ms}
	gceResponse := &p.GetCurrentExecutionResponse{RunID: runID}

	s.mockExecutionMgr.On("GetCurrentExecution", mock.Anything, mock.Anything).Return(gceResponse, nil).Once()
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse, nil).Once()
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("CreateWorkflowExecution", mock.Anything, mock.Anything).Return(&p.CreateWorkflowExecutionResponse{}, nil).Once()

	resp, err := s.historyEngine.SignalWithStartWorkflowExecution(context.Background(), sRequest)
	s.Nil(err)
	s.NotNil(resp.GetRunID())
	s.NotEqual(runID, resp.GetRunID())
}

func (s *engine2Suite) TestSignalWithStartWorkflowExecution_Start_DuplicateRequests() {
	domainID := constants.TestDomainID
	workflowID := "wId"
	runID := constants.TestRunID
	workflowType := "workflowType"
	taskList := "testTaskList"
	identity := "testIdentity"
	signalName := "my signal name"
	input := []byte("test input")
	requestID := "testRequestID"
	policy := types.WorkflowIDReusePolicyAllowDuplicate
	sRequest := &types.HistorySignalWithStartWorkflowExecutionRequest{
		DomainUUID: common.StringPtr(domainID),
		SignalWithStartRequest: &types.SignalWithStartWorkflowExecutionRequest{
			Domain:                              common.StringPtr(domainID),
			WorkflowID:                          common.StringPtr(workflowID),
			WorkflowType:                        &types.WorkflowType{Name: common.StringPtr(workflowType)},
			TaskList:                            &types.TaskList{Name: common.StringPtr(taskList)},
			ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(1),
			TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(2),
			Identity:                            common.StringPtr(identity),
			SignalName:                          common.StringPtr(signalName),
			Input:                               input,
			RequestID:                           common.StringPtr(requestID),
			WorkflowIDReusePolicy:               &policy,
		},
	}

	msBuilder := execution.NewMutableStateBuilderWithEventV2(
		s.historyEngine.shard,
		loggerimpl.NewDevelopmentForTest(s.Suite),
		runID,
		constants.TestLocalDomainEntry,
	)
	ms := execution.CreatePersistenceMutableState(msBuilder)
	ms.ExecutionInfo.State = p.WorkflowStateCompleted
	gwmsResponse := &p.GetWorkflowExecutionResponse{State: ms}
	gceResponse := &p.GetCurrentExecutionResponse{RunID: runID}
	workflowAlreadyStartedErr := &p.WorkflowExecutionAlreadyStartedError{
		Msg:              "random message",
		StartRequestID:   requestID, // use same requestID
		RunID:            runID,
		State:            p.WorkflowStateRunning,
		CloseStatus:      p.WorkflowCloseStatusNone,
		LastWriteVersion: common.EmptyVersion,
	}

	s.mockExecutionMgr.On("GetCurrentExecution", mock.Anything, mock.Anything).Return(gceResponse, nil).Once()
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse, nil).Once()
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("CreateWorkflowExecution", mock.Anything, mock.Anything).Return(nil, workflowAlreadyStartedErr).Once()

	resp, err := s.historyEngine.SignalWithStartWorkflowExecution(context.Background(), sRequest)
	s.Nil(err)
	s.NotNil(resp.GetRunID())
	s.Equal(runID, resp.GetRunID())
}

func (s *engine2Suite) TestSignalWithStartWorkflowExecution_Start_WorkflowAlreadyStarted() {
	domainID := constants.TestDomainID
	workflowID := "wId"
	runID := constants.TestRunID
	workflowType := "workflowType"
	taskList := "testTaskList"
	identity := "testIdentity"
	signalName := "my signal name"
	input := []byte("test input")
	requestID := "testRequestID"
	policy := types.WorkflowIDReusePolicyAllowDuplicate
	sRequest := &types.HistorySignalWithStartWorkflowExecutionRequest{
		DomainUUID: common.StringPtr(domainID),
		SignalWithStartRequest: &types.SignalWithStartWorkflowExecutionRequest{
			Domain:                              common.StringPtr(domainID),
			WorkflowID:                          common.StringPtr(workflowID),
			WorkflowType:                        &types.WorkflowType{Name: common.StringPtr(workflowType)},
			TaskList:                            &types.TaskList{Name: common.StringPtr(taskList)},
			ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(1),
			TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(2),
			Identity:                            common.StringPtr(identity),
			SignalName:                          common.StringPtr(signalName),
			Input:                               input,
			RequestID:                           common.StringPtr(requestID),
			WorkflowIDReusePolicy:               &policy,
		},
	}

	msBuilder := execution.NewMutableStateBuilderWithEventV2(
		s.historyEngine.shard,
		loggerimpl.NewDevelopmentForTest(s.Suite),
		runID,
		constants.TestLocalDomainEntry,
	)
	ms := execution.CreatePersistenceMutableState(msBuilder)
	ms.ExecutionInfo.State = p.WorkflowStateCompleted
	gwmsResponse := &p.GetWorkflowExecutionResponse{State: ms}
	gceResponse := &p.GetCurrentExecutionResponse{RunID: runID}
	workflowAlreadyStartedErr := &p.WorkflowExecutionAlreadyStartedError{
		Msg:              "random message",
		StartRequestID:   "new request ID",
		RunID:            runID,
		State:            p.WorkflowStateRunning,
		CloseStatus:      p.WorkflowCloseStatusNone,
		LastWriteVersion: common.EmptyVersion,
	}

	s.mockExecutionMgr.On("GetCurrentExecution", mock.Anything, mock.Anything).Return(gceResponse, nil).Once()
	s.mockExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(gwmsResponse, nil).Once()
	s.mockHistoryV2Mgr.On("AppendHistoryNodes", mock.Anything, mock.Anything).Return(&p.AppendHistoryNodesResponse{Size: 0}, nil).Once()
	s.mockExecutionMgr.On("CreateWorkflowExecution", mock.Anything, mock.Anything).Return(nil, workflowAlreadyStartedErr).Once()

	resp, err := s.historyEngine.SignalWithStartWorkflowExecution(context.Background(), sRequest)
	s.Nil(resp)
	s.NotNil(err)
}

func (s *engine2Suite) TestNewChildContext() {
	ctx := context.Background()
	childCtx, childCancel := s.historyEngine.newChildContext(ctx)
	defer childCancel()
	_, ok := childCtx.Deadline()
	s.True(ok)

	ctx, cancel := context.WithTimeout(ctx, time.Hour)
	defer cancel()
	childCtx, childCancel = s.historyEngine.newChildContext(ctx)
	deadline, ok := childCtx.Deadline()
	s.True(ok)
	s.True(deadline.Sub(time.Now()) < 10*time.Minute)
}

func (s *engine2Suite) getBuilder(domainID string, we types.WorkflowExecution) execution.MutableState {
	context, release, err := s.historyEngine.executionCache.GetOrCreateWorkflowExecutionForBackground(domainID, we)
	if err != nil {
		return nil
	}
	defer release(nil)

	return context.GetWorkflowExecution()
}
