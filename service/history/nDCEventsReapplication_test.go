// Copyright (c) 2019 Uber Technologies, Inc.
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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/client"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/definition"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service"
)

type (
	nDCEventReapplicationSuite struct {
		suite.Suite
		logger              log.Logger
		mockCtrl            *gomock.Controller
		mockExecutionMgr    *mocks.ExecutionManager
		mockHistoryMgr      *mocks.HistoryManager
		mockHistoryV2Mgr    *mocks.HistoryV2Manager
		mockShardManager    *mocks.ShardManager
		mockClusterMetadata *mocks.ClusterMetadata
		mockProducer        *mocks.KafkaProducer
		mockMetadataMgr     *mocks.MetadataManager
		mockMessagingClient messaging.Client
		mockService         service.Service
		mockShard           *shardContextImpl
		mockClientBean      *client.MockClientBean
		mockWorkflowResetor *mockWorkflowResetor
		mockDomainCache     *cache.DomainCacheMock
		nDCReapplication    *nDCEventsReapplication
	}
)

func TestNDCEventReapplicationSuite(t *testing.T) {
	s := new(nDCEventReapplicationSuite)
	suite.Run(t, s)
}

func (s *nDCEventReapplicationSuite) SetupTest() {
	s.logger = loggerimpl.NewDevelopmentForTest(s.Suite)
	s.mockCtrl = gomock.NewController(s.T())
	s.mockHistoryMgr = &mocks.HistoryManager{}
	s.mockHistoryV2Mgr = &mocks.HistoryV2Manager{}
	s.mockExecutionMgr = &mocks.ExecutionManager{}
	s.mockClusterMetadata = &mocks.ClusterMetadata{}
	s.mockShardManager = &mocks.ShardManager{}
	s.mockProducer = &mocks.KafkaProducer{}
	s.mockMessagingClient = mocks.NewMockMessagingClient(s.mockProducer, nil)
	s.mockMetadataMgr = &mocks.MetadataManager{}
	metricsClient := metrics.NewClient(tally.NoopScope, metrics.History)
	s.mockClientBean = &client.MockClientBean{}
	s.mockService = service.NewTestService(s.mockClusterMetadata, s.mockMessagingClient, metricsClient, s.mockClientBean, nil, nil)
	domainCache := cache.NewDomainCache(s.mockMetadataMgr, s.mockClusterMetadata, metricsClient, s.logger)
	s.mockShard = &shardContextImpl{
		service:                   s.mockService,
		shardInfo:                 &persistence.ShardInfo{ShardID: testShardID, RangeID: 1, TransferAckLevel: 0},
		shardID:                   testShardID,
		transferSequenceNumber:    1,
		executionManager:          s.mockExecutionMgr,
		shardManager:              s.mockShardManager,
		clusterMetadata:           s.mockClusterMetadata,
		historyMgr:                s.mockHistoryMgr,
		historyV2Mgr:              s.mockHistoryV2Mgr,
		maxTransferSequenceNumber: 100000,
		closeCh:                   make(chan int, 100),
		config:                    NewDynamicConfigForTest(),
		logger:                    s.logger,
		domainCache:               domainCache,
		metricsClient:             metrics.NewClient(tally.NoopScope, metrics.History),
		standbyClusterCurrentTime: make(map[string]time.Time),
		timeSource:                clock.NewRealTimeSource(),
	}
	s.mockClusterMetadata.On("GetCurrentClusterName").Return(cluster.TestCurrentClusterName)
	s.mockClusterMetadata.On("GetAllClusterInfo").Return(cluster.TestAllClusterInfo)
	s.mockClusterMetadata.On("IsGlobalDomainEnabled").Return(true)

	historyCache := newHistoryCache(s.mockShard)
	engine := &historyEngineImpl{
		currentClusterName:   s.mockShard.GetService().GetClusterMetadata().GetCurrentClusterName(),
		shard:                s.mockShard,
		clusterMetadata:      s.mockClusterMetadata,
		historyMgr:           s.mockHistoryMgr,
		executionManager:     s.mockExecutionMgr,
		historyCache:         historyCache,
		logger:               s.logger,
		tokenSerializer:      common.NewJSONTaskTokenSerializer(),
		metricsClient:        s.mockShard.GetMetricsClient(),
		timeSource:           s.mockShard.timeSource,
		historyEventNotifier: newHistoryEventNotifier(clock.NewRealTimeSource(), metrics.NewClient(tally.NoopScope, metrics.History), func(string) int { return 0 }),
	}
	s.mockShard.SetEngine(engine)
	s.mockWorkflowResetor = &mockWorkflowResetor{}
	s.mockDomainCache = &cache.DomainCacheMock{}
	s.nDCReapplication = newNDCEventsReapplication(
		clock.NewRealTimeSource(),
		s.mockWorkflowResetor,
		historyCache,
		s.mockDomainCache,
		metricsClient,
		s.logger,
		s.mockClusterMetadata,
	)
}

func (s *nDCEventReapplicationSuite) TestGetCurrentWorkflowMutableState() {
	domainUUID := uuid.New()
	workflowID := uuid.New()
	currentRunID := uuid.New()
	contextCurrent := &mockWorkflowExecutionContext{}
	defer contextCurrent.AssertExpectations(s.T())
	contextCurrent.On("lock", mock.Anything).Return(nil)
	contextCurrent.On("unlock")

	msBuilderCurrent := &mockMutableState{}
	defer msBuilderCurrent.AssertExpectations(s.T())

	contextCurrent.On("loadWorkflowExecution").Return(msBuilderCurrent, nil).Once()
	contextCurrentCacheKey := definition.NewWorkflowIdentifier(domainUUID, workflowID, currentRunID)
	s.nDCReapplication.historyCache.PutIfNotExist(contextCurrentCacheKey, contextCurrent)
	s.mockExecutionMgr.On("GetCurrentExecution", &persistence.GetCurrentExecutionRequest{
		DomainID:   domainUUID,
		WorkflowID: workflowID,
	}).Return(&persistence.GetCurrentExecutionResponse{
		RunID: currentRunID,
		// other attributes are not used
	}, nil)

	workflowContext, msBuilder, release, err := s.nDCReapplication.getCurrentWorkflowMutableState(context.Background(), domainUUID, workflowID)
	release(err)
	s.NotNil(workflowContext)
	s.NotNil(msBuilder)
	s.NotNil(release)
	s.NoError(err)
}

func (s *nDCEventReapplicationSuite) TestPrepareWorkflowMutation_LastWriteVersionActive() {
	msBuilderCurrent := &mockMutableState{}
	msBuilderCurrent.On("GetLastWriteVersion").Return(int64(1), nil)
	msBuilderCurrent.On("UpdateReplicationStateVersion", int64(1), true).Return()

	s.mockClusterMetadata.On("ClusterNameForFailoverVersion", int64(1)).Return(cluster.TestCurrentClusterName)
	s.mockClusterMetadata.On("GetCurrentClusterName").Return(cluster.TestCurrentClusterName)
	isCurrent, err := s.nDCReapplication.prepareWorkflowMutation(msBuilderCurrent)
	s.True(isCurrent)
	s.NoError(err)
}

func (s *nDCEventReapplicationSuite) TestPrepareWorkflowMutation_GetDomainCache() {
	execution := &persistence.WorkflowExecutionInfo{
		DomainID: uuid.New(),
	}
	msBuilderCurrent := &mockMutableState{}
	msBuilderCurrent.On("GetLastWriteVersion").Return(int64(1), nil)
	msBuilderCurrent.On("UpdateReplicationStateVersion", int64(1), true).Return()
	msBuilderCurrent.On("GetExecutionInfo").Return(execution)
	s.mockClusterMetadata.On("ClusterNameForFailoverVersion", int64(1)).Return(cluster.TestCurrentClusterName)
	s.mockClusterMetadata.On("GetCurrentClusterName").Return(cluster.TestAlternativeClusterName)

	replicationConfig := &persistence.DomainReplicationConfig{
		ActiveClusterName: cluster.TestCurrentClusterName,
	}
	domainEntry := cache.NewDomainCacheEntryForTest(nil, nil, false, replicationConfig, int64(2), s.mockClusterMetadata)
	s.mockDomainCache.On("GetDomainByID", execution.DomainID).Return(domainEntry)
	isCurrent, err := s.nDCReapplication.prepareWorkflowMutation(msBuilderCurrent)
	s.True(isCurrent)
	s.NoError(err)
}

func (s *nDCEventReapplicationSuite) TestPrepareWorkflowMutation_GetDomainCache_Error() {
	execution := &persistence.WorkflowExecutionInfo{
		DomainID: uuid.New(),
	}
	msBuilderCurrent := &mockMutableState{}
	msBuilderCurrent.On("GetLastWriteVersion").Return(int64(1), nil)
	msBuilderCurrent.On("UpdateReplicationStateVersion", int64(1), true).Return()
	msBuilderCurrent.On("GetExecutionInfo").Return(execution)
	s.mockClusterMetadata.On("ClusterNameForFailoverVersion", mock.Anything).Return(cluster.TestAlternativeClusterName)
	replicationConfig := &persistence.DomainReplicationConfig{
		ActiveClusterName: cluster.TestAlternativeClusterName,
	}
	domainEntry := cache.NewDomainCacheEntryForTest(nil, nil, false, replicationConfig, int64(0), s.mockClusterMetadata)
	s.mockDomainCache.On("GetDomainByID", execution.DomainID).Return(domainEntry, nil)
	isCurrent, err := s.nDCReapplication.prepareWorkflowMutation(msBuilderCurrent)
	s.False(isCurrent)
	s.NoError(err)
}

func (s *nDCEventReapplicationSuite) TestReapplyEventsToCurrentRunningWorkflow() {
	execution := &persistence.WorkflowExecutionInfo{
		DomainID: uuid.New(),
	}
	event := &shared.HistoryEvent{
		EventId:   common.Int64Ptr(1),
		EventType: common.EventTypePtr(shared.EventTypeWorkflowExecutionSignaled),
		WorkflowExecutionSignaledEventAttributes: &shared.WorkflowExecutionSignaledEventAttributes{
			Identity:   common.StringPtr("test"),
			SignalName: common.StringPtr("signal"),
			Input:      []byte{},
		},
	}
	msBuilderCurrent := &mockMutableState{}
	msBuilderCurrent.On("GetLastWriteVersion").Return(int64(1), nil)
	msBuilderCurrent.On("UpdateReplicationStateVersion", int64(1), true).Return()
	msBuilderCurrent.On("GetExecutionInfo").Return(execution)
	msBuilderCurrent.On("AddWorkflowExecutionSignaled", mock.Anything, mock.Anything, mock.Anything).Return(event, nil)
	s.mockClusterMetadata.On("ClusterNameForFailoverVersion", int64(1)).Return(cluster.TestCurrentClusterName)
	s.mockClusterMetadata.On("GetCurrentClusterName").Return(cluster.TestAlternativeClusterName)

	contextCurrent := &mockWorkflowExecutionContext{}
	contextCurrent.On("updateWorkflowExecutionAsActive", mock.Anything).Return(nil)
	contextCurrent.On("lock", mock.Anything).Return(nil)
	contextCurrent.On("unlock")

	replicationConfig := &persistence.DomainReplicationConfig{
		ActiveClusterName: cluster.TestCurrentClusterName,
	}
	domainEntry := cache.NewDomainCacheEntryForTest(nil, nil, false, replicationConfig, int64(2), s.mockClusterMetadata)
	s.mockDomainCache.On("GetDomainByID", execution.DomainID).Return(domainEntry)

	events := []*shared.HistoryEvent{event}
	err := s.nDCReapplication.reapplyEventsToCurrentRunningWorkflow(context.Background(), contextCurrent, msBuilderCurrent, events)
	s.NoError(err)
}

func (s *nDCEventReapplicationSuite) TestReapplyEventsToCurrentClosedWorkflow() {
	domainUUID := uuid.New()
	workflowID := uuid.New()
	currentRunID := uuid.New()
	execution := &persistence.WorkflowExecutionInfo{
		DomainID:   domainUUID,
		WorkflowID: workflowID,
	}
	event := &shared.HistoryEvent{
		EventId:   common.Int64Ptr(1),
		EventType: common.EventTypePtr(shared.EventTypeWorkflowExecutionSignaled),
		WorkflowExecutionSignaledEventAttributes: &shared.WorkflowExecutionSignaledEventAttributes{
			Identity:   common.StringPtr("test"),
			SignalName: common.StringPtr("signal"),
			Input:      []byte{},
		},
	}

	msBuilderCurrent := &mockMutableState{}
	msBuilderCurrent.On("GetLastWriteVersion").Return(int64(1), nil)
	msBuilderCurrent.On("UpdateReplicationStateVersion", int64(1), true).Return()
	msBuilderCurrent.On("GetExecutionInfo").Return(execution)
	msBuilderCurrent.On("AddWorkflowExecutionSignaled", mock.Anything, mock.Anything, mock.Anything).Return(event, nil)
	msBuilderCurrent.On("GetPreviousStartedEventID").Return(int64(3))
	msBuilderCurrent.On("IsWorkflowExecutionRunning").Return(false)
	s.mockClusterMetadata.On("ClusterNameForFailoverVersion", int64(1)).Return(cluster.TestCurrentClusterName)
	s.mockClusterMetadata.On("GetCurrentClusterName").Return(cluster.TestAlternativeClusterName)

	sharedExecution := &shared.WorkflowExecution{
		RunId:      common.StringPtr(currentRunID),
		WorkflowId: common.StringPtr(workflowID),
	}

	contextCurrent := &mockWorkflowExecutionContext{}
	contextCurrent.On("updateWorkflowExecutionAsActive", mock.Anything).Return(nil)
	contextCurrent.On("loadWorkflowExecution").Return(msBuilderCurrent, nil).Once()
	contextCurrent.On("getExecution").Return(sharedExecution)
	contextCurrent.On("lock", mock.Anything).Return(nil)
	contextCurrent.On("unlock")

	replicationConfig := &persistence.DomainReplicationConfig{
		ActiveClusterName: cluster.TestCurrentClusterName,
	}
	domainInfo := &persistence.DomainInfo{
		ID:   domainUUID,
		Name: domainUUID,
	}
	domainEntry := cache.NewDomainCacheEntryForTest(domainInfo, nil, false, replicationConfig, int64(1), s.mockClusterMetadata)
	s.mockDomainCache.On("GetDomainByID", execution.DomainID).Return(domainEntry, nil)

	contextCurrentCacheKey := definition.NewWorkflowIdentifier(domainUUID, workflowID, currentRunID)
	s.nDCReapplication.historyCache.PutIfNotExist(contextCurrentCacheKey, contextCurrent)
	s.mockExecutionMgr.On("GetCurrentExecution", &persistence.GetCurrentExecutionRequest{
		DomainID:   domainUUID,
		WorkflowID: workflowID,
	}).Return(&persistence.GetCurrentExecutionResponse{
		RunID: currentRunID,
		// other attributes are not used
	}, nil)

	s.mockWorkflowResetor.On(
		"ResetWorkflowExecution",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(&shared.ResetWorkflowExecutionResponse{}, nil)
	events := []*shared.HistoryEvent{event}
	err := s.nDCReapplication.reapplyEventsToCurrentClosedWorkflow(context.Background(), contextCurrent, msBuilderCurrent, events)
	s.NoError(err)
}

func (s *nDCEventReapplicationSuite) TestReapplyEvents() {
	domainUUID := uuid.New()
	workflowID := uuid.New()
	currentRunID := uuid.New()
	execution := &persistence.WorkflowExecutionInfo{
		DomainID: uuid.New(),
	}
	event := &shared.HistoryEvent{
		EventId:   common.Int64Ptr(1),
		EventType: common.EventTypePtr(shared.EventTypeWorkflowExecutionSignaled),
		WorkflowExecutionSignaledEventAttributes: &shared.WorkflowExecutionSignaledEventAttributes{
			Identity:   common.StringPtr("test"),
			SignalName: common.StringPtr("signal"),
			Input:      []byte{},
		},
	}
	msBuilderCurrent := &mockMutableState{}
	msBuilderCurrent.On("GetLastWriteVersion").Return(int64(1), nil)
	msBuilderCurrent.On("UpdateReplicationStateVersion", int64(1), true).Return()
	msBuilderCurrent.On("GetExecutionInfo").Return(execution)
	msBuilderCurrent.On("AddWorkflowExecutionSignaled", mock.Anything, mock.Anything, mock.Anything).Return(event, nil).Once()
	msBuilderCurrent.On("IsWorkflowExecutionRunning").Return(true)
	s.mockClusterMetadata.On("ClusterNameForFailoverVersion", int64(1)).Return(cluster.TestCurrentClusterName)
	s.mockClusterMetadata.On("GetCurrentClusterName").Return(cluster.TestAlternativeClusterName)

	contextCurrent := &mockWorkflowExecutionContext{}
	contextCurrent.On("updateWorkflowExecutionAsActive", mock.Anything).Return(nil)
	contextCurrent.On("lock", mock.Anything).Return(nil)
	contextCurrent.On("unlock")
	contextCurrent.On("loadWorkflowExecution").Return(msBuilderCurrent, nil).Once()
	contextCurrentCacheKey := definition.NewWorkflowIdentifier(domainUUID, workflowID, currentRunID)
	s.nDCReapplication.historyCache.PutIfNotExist(contextCurrentCacheKey, contextCurrent)
	s.mockExecutionMgr.On("GetCurrentExecution", &persistence.GetCurrentExecutionRequest{
		DomainID:   domainUUID,
		WorkflowID: workflowID,
	}).Return(&persistence.GetCurrentExecutionResponse{
		RunID: currentRunID,
		// other attributes are not used
	}, nil)

	replicationConfig := &persistence.DomainReplicationConfig{
		ActiveClusterName: cluster.TestCurrentClusterName,
	}
	domainEntry := cache.NewDomainCacheEntryForTest(nil, nil, false, replicationConfig, int64(2), s.mockClusterMetadata)
	s.mockDomainCache.On("GetDomainByID", execution.DomainID).Return(domainEntry)

	events := []*shared.HistoryEvent{
		{EventType: common.EventTypePtr(shared.EventTypeWorkflowExecutionStarted)},
		event,
	}
	err := s.nDCReapplication.reapplyEvents(context.Background(), domainUUID, workflowID, events)
	s.NoError(err)
}
