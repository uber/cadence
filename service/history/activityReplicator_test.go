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
	ctx "context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"

	h "github.com/uber/cadence/.gen/go/history"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/client"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service"
)

type (
	activityReplicatorSuite struct {
		suite.Suite
		logger                   log.Logger
		mockCtrl                 *gomock.Controller
		mockExecutionMgr         *mocks.ExecutionManager
		mockHistoryV2Mgr         *mocks.HistoryV2Manager
		mockShardManager         *mocks.ShardManager
		mockClusterMetadata      *mocks.ClusterMetadata
		mockProducer             *mocks.KafkaProducer
		mockDomainCache          *cache.DomainCacheMock
		mockMessagingClient      messaging.Client
		mockService              service.Service
		mockShard                *shardContextImpl
		mockClientBean           *client.MockClientBean
		mockTxProcessor          *MockTransferQueueProcessor
		mockReplicationProcessor *MockReplicatorQueueProcessor
		mockTimerProcessor       *MockTimerQueueProcessor
		historyCache             *historyCache

		activityReplicator activityReplicator
	}
)

func TestActivityReplicatorSuite(t *testing.T) {
	s := new(activityReplicatorSuite)
	suite.Run(t, s)
}

func (s *activityReplicatorSuite) SetupSuite() {

}

func (s *activityReplicatorSuite) TearDownSuite() {

}

func (s *activityReplicatorSuite) SetupTest() {
	s.logger = loggerimpl.NewDevelopmentForTest(s.Suite)
	s.mockCtrl = gomock.NewController(s.T())
	s.mockHistoryV2Mgr = &mocks.HistoryV2Manager{}
	s.mockExecutionMgr = &mocks.ExecutionManager{}
	s.mockClusterMetadata = &mocks.ClusterMetadata{}
	s.mockShardManager = &mocks.ShardManager{}
	s.mockProducer = &mocks.KafkaProducer{}
	s.mockMessagingClient = mocks.NewMockMessagingClient(s.mockProducer, nil)
	s.mockDomainCache = &cache.DomainCacheMock{}
	metricsClient := metrics.NewClient(tally.NoopScope, metrics.History)
	s.mockClientBean = &client.MockClientBean{}
	s.mockService = service.NewTestService(
		s.mockClusterMetadata,
		s.mockMessagingClient,
		metricsClient,
		s.mockClientBean,
		nil,
		nil,
		nil)

	s.mockShard = &shardContextImpl{
		service:                   s.mockService,
		shardInfo:                 &persistence.ShardInfo{ShardID: testShardID, RangeID: 1, TransferAckLevel: 0},
		shardID:                   testShardID,
		transferSequenceNumber:    1,
		executionManager:          s.mockExecutionMgr,
		shardManager:              s.mockShardManager,
		clusterMetadata:           s.mockClusterMetadata,
		historyV2Mgr:              s.mockHistoryV2Mgr,
		maxTransferSequenceNumber: 100000,
		closeCh:                   make(chan int, 100),
		config:                    NewDynamicConfigForTest(),
		logger:                    s.logger,
		domainCache:               s.mockDomainCache,
		metricsClient:             metrics.NewClient(tally.NoopScope, metrics.History),
		standbyClusterCurrentTime: make(map[string]time.Time),
		timeSource:                clock.NewRealTimeSource(),
	}
	s.mockClusterMetadata.On("GetCurrentClusterName").Return(cluster.TestCurrentClusterName)
	s.mockClusterMetadata.On("GetAllClusterInfo").Return(cluster.TestAllClusterInfo)
	s.mockClusterMetadata.On("IsGlobalDomainEnabled").Return(true)
	s.mockTxProcessor = &MockTransferQueueProcessor{}
	s.mockTxProcessor.On("NotifyNewTask", mock.Anything, mock.Anything).Maybe()
	s.mockReplicationProcessor = NewMockReplicatorQueueProcessor(s.mockCtrl)
	s.mockReplicationProcessor.EXPECT().notifyNewTask().AnyTimes()
	s.mockTimerProcessor = &MockTimerQueueProcessor{}
	s.mockTimerProcessor.On("NotifyNewTimers", mock.Anything, mock.Anything).Maybe()

	s.historyCache = newHistoryCache(s.mockShard)
	engine := &historyEngineImpl{
		currentClusterName:   s.mockShard.GetService().GetClusterMetadata().GetCurrentClusterName(),
		shard:                s.mockShard,
		clusterMetadata:      s.mockClusterMetadata,
		executionManager:     s.mockExecutionMgr,
		historyCache:         s.historyCache,
		logger:               s.logger,
		tokenSerializer:      common.NewJSONTaskTokenSerializer(),
		metricsClient:        s.mockShard.GetMetricsClient(),
		timeSource:           s.mockShard.timeSource,
		historyEventNotifier: newHistoryEventNotifier(clock.NewRealTimeSource(), metrics.NewClient(tally.NoopScope, metrics.History), func(string) int { return 0 }),
		txProcessor:          s.mockTxProcessor,
		replicatorProcessor:  s.mockReplicationProcessor,
		timerProcessor:       s.mockTimerProcessor,
	}
	s.mockShard.SetEngine(engine)

	s.activityReplicator = newActivityReplicator(
		s.mockShard,
		s.historyCache,
		s.logger,
	)
}

func (s *activityReplicatorSuite) TearDownTest() {
	s.mockCtrl.Finish()
	s.mockExecutionMgr.AssertExpectations(s.T())
	s.mockShardManager.AssertExpectations(s.T())
	s.mockProducer.AssertExpectations(s.T())
	s.mockDomainCache.AssertExpectations(s.T())
	s.mockTxProcessor.AssertExpectations(s.T())
	s.mockTimerProcessor.AssertExpectations(s.T())
	s.mockClientBean.AssertExpectations(s.T())
	s.mockTxProcessor.AssertExpectations(s.T())
	s.mockTimerProcessor.AssertExpectations(s.T())
}

func (s *activityReplicatorSuite) TestSyncActivity_WorkflowNotFound() {
	domainName := "some random domain name"
	domainID := testDomainID
	workflowID := "some random workflow ID"
	runID := uuid.New()
	version := int64(100)

	request := &h.SyncActivityRequest{
		DomainId:   common.StringPtr(domainID),
		WorkflowId: common.StringPtr(workflowID),
		RunId:      common.StringPtr(runID),
	}
	s.mockExecutionMgr.On("GetWorkflowExecution", &persistence.GetWorkflowExecutionRequest{
		DomainID: domainID,
		Execution: shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
	}).Return(nil, &shared.EntityNotExistsError{})
	s.mockDomainCache.On("GetDomainByID", domainID).Return(
		cache.NewGlobalDomainCacheEntryForTest(
			&persistence.DomainInfo{ID: domainID, Name: domainName},
			&persistence.DomainConfig{Retention: 1},
			&persistence.DomainReplicationConfig{
				ActiveClusterName: cluster.TestCurrentClusterName,
				Clusters: []*persistence.ClusterReplicationConfig{
					{ClusterName: cluster.TestCurrentClusterName},
					{ClusterName: cluster.TestAlternativeClusterName},
				},
			},
			version,
			nil,
		), nil,
	)

	err := s.activityReplicator.SyncActivity(ctx.Background(), request)
	s.Nil(err)
}

func (s *activityReplicatorSuite) TestSyncActivity_WorkflowClosed() {
	domainName := "some random domain name"
	domainID := testDomainID
	workflowID := "some random workflow ID"
	runID := uuid.New()
	version := int64(100)

	context, release, err := s.historyCache.getOrCreateWorkflowExecutionForBackground(
		domainID,
		shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
	)
	s.Nil(err)
	msBuilder := &mockMutableState{}
	defer msBuilder.AssertExpectations(s.T())
	context.(*workflowExecutionContextImpl).msBuilder = msBuilder
	release(nil)
	request := &h.SyncActivityRequest{
		DomainId:   common.StringPtr(domainID),
		WorkflowId: common.StringPtr(workflowID),
		RunId:      common.StringPtr(runID),
	}
	msBuilder.On("StartTransaction", mock.Anything).Return(false, nil).Once()
	msBuilder.On("IsWorkflowExecutionRunning").Return(false)
	var versionHistories *persistence.VersionHistories
	msBuilder.On("GetVersionHistories").Return(versionHistories)
	msBuilder.On("GetReplicationState").Return(&persistence.ReplicationState{})
	s.mockDomainCache.On("GetDomainByID", domainID).Return(
		cache.NewGlobalDomainCacheEntryForTest(
			&persistence.DomainInfo{ID: domainID, Name: domainName},
			&persistence.DomainConfig{Retention: 1},
			&persistence.DomainReplicationConfig{
				ActiveClusterName: cluster.TestCurrentClusterName,
				Clusters: []*persistence.ClusterReplicationConfig{
					{ClusterName: cluster.TestCurrentClusterName},
					{ClusterName: cluster.TestAlternativeClusterName},
				},
			},
			version,
			nil,
		), nil,
	)

	err = s.activityReplicator.SyncActivity(ctx.Background(), request)
	s.Nil(err)
}

func (s *activityReplicatorSuite) TestSyncActivity_IncomingScheduleIDLarger_IncomingVersionSmaller() {
	domainName := "some random domain name"
	domainID := testDomainID
	workflowID := "some random workflow ID"
	runID := uuid.New()
	scheduleID := int64(144)
	version := int64(100)

	lastWriteVersion := version + 100
	nextEventID := scheduleID - 10

	context, release, err := s.historyCache.getOrCreateWorkflowExecutionForBackground(
		domainID,
		shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
	)
	s.Nil(err)
	msBuilder := &mockMutableState{}
	defer msBuilder.AssertExpectations(s.T())
	context.(*workflowExecutionContextImpl).msBuilder = msBuilder
	release(nil)
	request := &h.SyncActivityRequest{
		DomainId:    common.StringPtr(domainID),
		WorkflowId:  common.StringPtr(workflowID),
		RunId:       common.StringPtr(runID),
		Version:     common.Int64Ptr(version),
		ScheduledId: common.Int64Ptr(scheduleID),
	}
	msBuilder.On("StartTransaction", mock.Anything).Return(false, nil).Once()
	msBuilder.On("IsWorkflowExecutionRunning").Return(true)
	msBuilder.On("GetNextEventID").Return(nextEventID)
	msBuilder.On("GetLastWriteVersion").Return(lastWriteVersion, nil)
	var versionHistories *persistence.VersionHistories
	msBuilder.On("GetVersionHistories").Return(versionHistories)
	msBuilder.On("GetReplicationState").Return(&persistence.ReplicationState{})
	s.mockDomainCache.On("GetDomainByID", domainID).Return(
		cache.NewGlobalDomainCacheEntryForTest(
			&persistence.DomainInfo{ID: domainID, Name: domainName},
			&persistence.DomainConfig{Retention: 1},
			&persistence.DomainReplicationConfig{
				ActiveClusterName: cluster.TestCurrentClusterName,
				Clusters: []*persistence.ClusterReplicationConfig{
					{ClusterName: cluster.TestCurrentClusterName},
					{ClusterName: cluster.TestAlternativeClusterName},
				},
			},
			lastWriteVersion,
			nil,
		), nil,
	)

	err = s.activityReplicator.SyncActivity(ctx.Background(), request)
	s.Nil(err)
}

func (s *activityReplicatorSuite) TestSyncActivity_IncomingScheduleIDLarger_IncomingVersionLarger() {
	domainName := "some random domain name"
	domainID := testDomainID
	workflowID := "some random workflow ID"
	runID := uuid.New()
	scheduleID := int64(144)
	version := int64(100)

	lastWriteVersion := version - 100
	nextEventID := scheduleID - 10

	context, release, err := s.historyCache.getOrCreateWorkflowExecutionForBackground(
		domainID,
		shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
	)
	s.Nil(err)
	msBuilder := &mockMutableState{}
	defer msBuilder.AssertExpectations(s.T())
	context.(*workflowExecutionContextImpl).msBuilder = msBuilder
	release(nil)
	request := &h.SyncActivityRequest{
		DomainId:    common.StringPtr(domainID),
		WorkflowId:  common.StringPtr(workflowID),
		RunId:       common.StringPtr(runID),
		Version:     common.Int64Ptr(version),
		ScheduledId: common.Int64Ptr(scheduleID),
	}
	msBuilder.On("StartTransaction", mock.Anything).Return(false, nil).Once()
	msBuilder.On("IsWorkflowExecutionRunning").Return(true)
	msBuilder.On("GetNextEventID").Return(nextEventID)
	msBuilder.On("GetLastWriteVersion").Return(lastWriteVersion, nil)
	var versionHistories *persistence.VersionHistories
	msBuilder.On("GetVersionHistories").Return(versionHistories)
	msBuilder.On("GetReplicationState").Return(&persistence.ReplicationState{})
	s.mockDomainCache.On("GetDomainByID", domainID).Return(
		cache.NewGlobalDomainCacheEntryForTest(
			&persistence.DomainInfo{ID: domainID, Name: domainName},
			&persistence.DomainConfig{Retention: 1},
			&persistence.DomainReplicationConfig{
				ActiveClusterName: cluster.TestCurrentClusterName,
				Clusters: []*persistence.ClusterReplicationConfig{
					{ClusterName: cluster.TestCurrentClusterName},
					{ClusterName: cluster.TestAlternativeClusterName},
				},
			},
			lastWriteVersion,
			nil,
		), nil,
	)

	err = s.activityReplicator.SyncActivity(ctx.Background(), request)
	s.Equal(newRetryTaskErrorWithHint(ErrRetrySyncActivityMsg, domainID, workflowID, runID, nextEventID), err)
}

func (s *activityReplicatorSuite) TestSyncActivity_VersionHistories_IncomingVersionSmaller_DiscardTask() {
	domainName := "some random domain name"
	domainID := testDomainID
	workflowID := "some random workflow ID"
	runID := uuid.New()
	scheduleID := int64(144)
	version := int64(99)

	lastWriteVersion := version - 100
	nextEventID := scheduleID - 10
	incomingVersionHistory := persistence.VersionHistory{
		BranchToken: []byte{},
		Items: []*persistence.VersionHistoryItem{
			{
				EventID: 50,
				Version: 2,
			},
			{
				EventID: 144,
				Version: 99,
			},
		},
	}
	context, release, err := s.historyCache.getOrCreateWorkflowExecutionForBackground(
		domainID,
		shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
	)
	s.Nil(err)
	msBuilder := &mockMutableState{}
	defer msBuilder.AssertExpectations(s.T())
	context.(*workflowExecutionContextImpl).msBuilder = msBuilder
	release(nil)
	request := &h.SyncActivityRequest{
		DomainId:       common.StringPtr(domainID),
		WorkflowId:     common.StringPtr(workflowID),
		RunId:          common.StringPtr(runID),
		Version:        common.Int64Ptr(version),
		ScheduledId:    common.Int64Ptr(scheduleID),
		VersionHistory: incomingVersionHistory.ToThrift(),
	}
	msBuilder.On("StartTransaction", mock.Anything).Return(false, nil).Once()
	localVersionHistories := &persistence.VersionHistories{
		CurrentVersionHistoryIndex: 0,
		Histories: []*persistence.VersionHistory{
			{
				BranchToken: []byte{},
				Items: []*persistence.VersionHistoryItem{
					{
						EventID: nextEventID - 1,
						Version: 100,
					},
				},
			},
		},
	}
	msBuilder.On("GetVersionHistories").Return(localVersionHistories)
	s.mockDomainCache.On("GetDomainByID", domainID).Return(
		cache.NewGlobalDomainCacheEntryForTest(
			&persistence.DomainInfo{ID: domainID, Name: domainName},
			&persistence.DomainConfig{Retention: 1},
			&persistence.DomainReplicationConfig{
				ActiveClusterName: cluster.TestCurrentClusterName,
				Clusters: []*persistence.ClusterReplicationConfig{
					{ClusterName: cluster.TestCurrentClusterName},
					{ClusterName: cluster.TestAlternativeClusterName},
				},
			},
			lastWriteVersion,
			nil,
		), nil,
	)

	err = s.activityReplicator.SyncActivity(ctx.Background(), request)
	s.Nil(err)
}

func (s *activityReplicatorSuite) TestSyncActivity_DifferentVersionHistories_IncomingVersionSLarger_ReturnRetryError() {
	domainName := "some random domain name"
	domainID := testDomainID
	workflowID := "some random workflow ID"
	runID := uuid.New()
	scheduleID := int64(144)
	version := int64(100)
	lastWriteVersion := version - 100

	incomingVersionHistory := persistence.VersionHistory{
		BranchToken: []byte{},
		Items: []*persistence.VersionHistoryItem{
			{
				EventID: 50,
				Version: 2,
			},
			{
				EventID: scheduleID,
				Version: version,
			},
		},
	}
	context, release, err := s.historyCache.getOrCreateWorkflowExecutionForBackground(
		domainID,
		shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
	)
	s.Nil(err)
	msBuilder := &mockMutableState{}
	defer msBuilder.AssertExpectations(s.T())
	context.(*workflowExecutionContextImpl).msBuilder = msBuilder
	release(nil)
	request := &h.SyncActivityRequest{
		DomainId:       common.StringPtr(domainID),
		WorkflowId:     common.StringPtr(workflowID),
		RunId:          common.StringPtr(runID),
		Version:        common.Int64Ptr(version),
		ScheduledId:    common.Int64Ptr(scheduleID),
		VersionHistory: incomingVersionHistory.ToThrift(),
	}
	msBuilder.On("StartTransaction", mock.Anything).Return(false, nil).Once()
	localVersionHistories := &persistence.VersionHistories{
		CurrentVersionHistoryIndex: 0,
		Histories: []*persistence.VersionHistory{
			{
				BranchToken: []byte{},
				Items: []*persistence.VersionHistoryItem{
					{
						EventID: 100,
						Version: 2,
					},
				},
			},
		},
	}
	msBuilder.On("GetVersionHistories").Return(localVersionHistories)
	s.mockDomainCache.On("GetDomainByID", domainID).Return(
		cache.NewGlobalDomainCacheEntryForTest(
			&persistence.DomainInfo{ID: domainID, Name: domainName},
			&persistence.DomainConfig{Retention: 1},
			&persistence.DomainReplicationConfig{
				ActiveClusterName: cluster.TestCurrentClusterName,
				Clusters: []*persistence.ClusterReplicationConfig{
					{ClusterName: cluster.TestCurrentClusterName},
					{ClusterName: cluster.TestAlternativeClusterName},
				},
			},
			lastWriteVersion,
			nil,
		), nil,
	)

	err = s.activityReplicator.SyncActivity(ctx.Background(), request)
	s.Equal(newNDCRetryTaskErrorWithHint(
		domainID,
		workflowID,
		runID,
		common.Int64Ptr(50),
		common.Int64Ptr(2),
		common.Int64Ptr(scheduleID),
		common.Int64Ptr(version),
	),
		err,
	)
}

func (s *activityReplicatorSuite) TestSyncActivity_VersionHistories_IncomingScheduleIDLarger_ReturnRetryError() {
	domainName := "some random domain name"
	domainID := testDomainID
	workflowID := "some random workflow ID"
	runID := uuid.New()
	scheduleID := int64(144)
	version := int64(100)

	lastWriteVersion := version - 100
	incomingVersionHistory := persistence.VersionHistory{
		BranchToken: []byte{},
		Items: []*persistence.VersionHistoryItem{
			{
				EventID: 50,
				Version: 2,
			},
			{
				EventID: scheduleID,
				Version: version,
			},
		},
	}
	context, release, err := s.historyCache.getOrCreateWorkflowExecutionForBackground(
		domainID,
		shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
	)
	s.Nil(err)
	msBuilder := &mockMutableState{}
	defer msBuilder.AssertExpectations(s.T())
	context.(*workflowExecutionContextImpl).msBuilder = msBuilder
	release(nil)
	request := &h.SyncActivityRequest{
		DomainId:       common.StringPtr(domainID),
		WorkflowId:     common.StringPtr(workflowID),
		RunId:          common.StringPtr(runID),
		Version:        common.Int64Ptr(version),
		ScheduledId:    common.Int64Ptr(scheduleID),
		VersionHistory: incomingVersionHistory.ToThrift(),
	}
	msBuilder.On("StartTransaction", mock.Anything).Return(false, nil).Once()
	localVersionHistories := &persistence.VersionHistories{
		CurrentVersionHistoryIndex: 0,
		Histories: []*persistence.VersionHistory{
			{
				BranchToken: []byte{},
				Items: []*persistence.VersionHistoryItem{
					{
						EventID: 100,
						Version: version,
					},
				},
			},
		},
	}
	msBuilder.On("GetVersionHistories").Return(localVersionHistories)
	s.mockDomainCache.On("GetDomainByID", domainID).Return(
		cache.NewGlobalDomainCacheEntryForTest(
			&persistence.DomainInfo{ID: domainID, Name: domainName},
			&persistence.DomainConfig{Retention: 1},
			&persistence.DomainReplicationConfig{
				ActiveClusterName: cluster.TestCurrentClusterName,
				Clusters: []*persistence.ClusterReplicationConfig{
					{ClusterName: cluster.TestCurrentClusterName},
					{ClusterName: cluster.TestAlternativeClusterName},
				},
			},
			lastWriteVersion,
			nil,
		), nil,
	)

	err = s.activityReplicator.SyncActivity(ctx.Background(), request)
	s.Equal(newNDCRetryTaskErrorWithHint(
		domainID,
		workflowID,
		runID,
		common.Int64Ptr(100),
		common.Int64Ptr(version),
		common.Int64Ptr(scheduleID),
		common.Int64Ptr(version),
	),
		err,
	)
}

func (s *activityReplicatorSuite) TestSyncActivity_ActivityCompleted() {
	domainName := "some random domain name"
	domainID := testDomainID
	workflowID := "some random workflow ID"
	runID := uuid.New()
	scheduleID := int64(144)
	version := int64(100)

	lastWriteVersion := version
	nextEventID := scheduleID + 10

	context, release, err := s.historyCache.getOrCreateWorkflowExecutionForBackground(
		domainID,
		shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
	)
	s.Nil(err)
	msBuilder := &mockMutableState{}
	defer msBuilder.AssertExpectations(s.T())
	context.(*workflowExecutionContextImpl).msBuilder = msBuilder
	release(nil)
	request := &h.SyncActivityRequest{
		DomainId:    common.StringPtr(domainID),
		WorkflowId:  common.StringPtr(workflowID),
		RunId:       common.StringPtr(runID),
		Version:     common.Int64Ptr(version),
		ScheduledId: common.Int64Ptr(scheduleID),
	}
	msBuilder.On("StartTransaction", mock.Anything).Return(false, nil).Once()
	msBuilder.On("IsWorkflowExecutionRunning").Return(true)
	msBuilder.On("GetNextEventID").Return(nextEventID)
	var versionHistories *persistence.VersionHistories
	msBuilder.On("GetVersionHistories").Return(versionHistories)
	msBuilder.On("GetReplicationState").Return(&persistence.ReplicationState{})
	s.mockDomainCache.On("GetDomainByID", domainID).Return(
		cache.NewGlobalDomainCacheEntryForTest(
			&persistence.DomainInfo{ID: domainID, Name: domainName},
			&persistence.DomainConfig{Retention: 1},
			&persistence.DomainReplicationConfig{
				ActiveClusterName: cluster.TestCurrentClusterName,
				Clusters: []*persistence.ClusterReplicationConfig{
					{ClusterName: cluster.TestCurrentClusterName},
					{ClusterName: cluster.TestAlternativeClusterName},
				},
			},
			lastWriteVersion,
			nil,
		), nil,
	)
	msBuilder.On("GetActivityInfo", scheduleID).Return(nil, false)

	err = s.activityReplicator.SyncActivity(ctx.Background(), request)
	s.Nil(err)
}

func (s *activityReplicatorSuite) TestSyncActivity_ActivityRunning_LocalActivityVersionLarger() {
	domainName := "some random domain name"
	domainID := testDomainID
	workflowID := "some random workflow ID"
	runID := uuid.New()
	scheduleID := int64(144)
	version := int64(100)

	lastWriteVersion := version + 10
	nextEventID := scheduleID + 10

	context, release, err := s.historyCache.getOrCreateWorkflowExecutionForBackground(
		domainID,
		shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
	)
	s.Nil(err)
	msBuilder := &mockMutableState{}
	defer msBuilder.AssertExpectations(s.T())
	context.(*workflowExecutionContextImpl).msBuilder = msBuilder
	release(nil)
	request := &h.SyncActivityRequest{
		DomainId:    common.StringPtr(domainID),
		WorkflowId:  common.StringPtr(workflowID),
		RunId:       common.StringPtr(runID),
		Version:     common.Int64Ptr(version),
		ScheduledId: common.Int64Ptr(scheduleID),
	}
	msBuilder.On("StartTransaction", mock.Anything).Return(false, nil).Once()
	msBuilder.On("IsWorkflowExecutionRunning").Return(true)
	msBuilder.On("GetNextEventID").Return(nextEventID)
	var versionHistories *persistence.VersionHistories
	msBuilder.On("GetVersionHistories").Return(versionHistories)
	msBuilder.On("GetReplicationState").Return(&persistence.ReplicationState{})
	s.mockDomainCache.On("GetDomainByID", domainID).Return(
		cache.NewGlobalDomainCacheEntryForTest(
			&persistence.DomainInfo{ID: domainID, Name: domainName},
			&persistence.DomainConfig{Retention: 1},
			&persistence.DomainReplicationConfig{
				ActiveClusterName: cluster.TestCurrentClusterName,
				Clusters: []*persistence.ClusterReplicationConfig{
					{ClusterName: cluster.TestCurrentClusterName},
					{ClusterName: cluster.TestAlternativeClusterName},
				},
			},
			lastWriteVersion,
			nil,
		), nil,
	)
	msBuilder.On("GetActivityInfo", scheduleID).Return(&persistence.ActivityInfo{
		Version: lastWriteVersion - 1,
	}, true)

	err = s.activityReplicator.SyncActivity(ctx.Background(), request)
	s.Nil(err)
}

func (s *activityReplicatorSuite) TestSyncActivity_ActivityRunning_Update_SameVersionSameAttempt() {
	domainName := "some random domain name"
	domainID := testDomainID
	workflowID := "some random workflow ID"
	runID := uuid.New()
	version := int64(100)
	scheduleID := int64(144)
	scheduledTime := time.Now()
	startedID := scheduleID + 1
	startedTime := scheduledTime.Add(time.Minute)
	heartBeatUpdatedTime := startedTime.Add(time.Minute)
	attempt := int32(0)
	details := []byte("some random activity heartbeat progress")

	nextEventID := scheduleID + 10

	context, release, err := s.historyCache.getOrCreateWorkflowExecutionForBackground(
		domainID,
		shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
	)
	s.Nil(err)
	msBuilder := &mockMutableState{}
	defer msBuilder.AssertExpectations(s.T())
	context.(*workflowExecutionContextImpl).msBuilder = msBuilder
	release(nil)
	request := &h.SyncActivityRequest{
		DomainId:          common.StringPtr(domainID),
		WorkflowId:        common.StringPtr(workflowID),
		RunId:             common.StringPtr(runID),
		Version:           common.Int64Ptr(version),
		ScheduledId:       common.Int64Ptr(scheduleID),
		ScheduledTime:     common.Int64Ptr(scheduledTime.UnixNano()),
		StartedId:         common.Int64Ptr(startedID),
		StartedTime:       common.Int64Ptr(startedTime.UnixNano()),
		Attempt:           common.Int32Ptr(attempt),
		LastHeartbeatTime: common.Int64Ptr(heartBeatUpdatedTime.UnixNano()),
		Details:           details,
	}
	msBuilder.On("StartTransaction", mock.Anything).Return(false, nil).Once()
	msBuilder.On("IsWorkflowExecutionRunning").Return(true)
	msBuilder.On("GetNextEventID").Return(nextEventID)
	var versionHistories *persistence.VersionHistories
	msBuilder.On("GetVersionHistories").Return(versionHistories)
	msBuilder.On("GetReplicationState").Return(&persistence.ReplicationState{})
	s.mockDomainCache.On("GetDomainByID", domainID).Return(
		cache.NewGlobalDomainCacheEntryForTest(
			&persistence.DomainInfo{ID: domainID, Name: domainName},
			&persistence.DomainConfig{Retention: 1},
			&persistence.DomainReplicationConfig{
				ActiveClusterName: cluster.TestCurrentClusterName,
				Clusters: []*persistence.ClusterReplicationConfig{
					{ClusterName: cluster.TestCurrentClusterName},
					{ClusterName: cluster.TestAlternativeClusterName},
				},
			},
			version,
			nil,
		), nil,
	)
	activityInfo := &persistence.ActivityInfo{
		Version:    version,
		ScheduleID: scheduleID,
		Attempt:    attempt,
	}
	msBuilder.On("GetActivityInfo", scheduleID).Return(activityInfo, true)
	s.mockClusterMetadata.On("IsVersionFromSameCluster", version, activityInfo.Version).Return(true)

	expectedErr := errors.New("this is error is used to by pass lots of mocking")
	msBuilder.On("ReplicateActivityInfo", request, false).Return(expectedErr)

	err = s.activityReplicator.SyncActivity(ctx.Background(), request)
	s.Equal(expectedErr, err)
}

func (s *activityReplicatorSuite) TestSyncActivity_ActivityRunning_Update_SameVersionLargerAttempt() {
	domainName := "some random domain name"
	domainID := testDomainID
	workflowID := "some random workflow ID"
	runID := uuid.New()
	version := int64(100)
	scheduleID := int64(144)
	scheduledTime := time.Now()
	startedID := scheduleID + 1
	startedTime := scheduledTime.Add(time.Minute)
	heartBeatUpdatedTime := startedTime.Add(time.Minute)
	attempt := int32(100)
	details := []byte("some random activity heartbeat progress")

	nextEventID := scheduleID + 10

	context, release, err := s.historyCache.getOrCreateWorkflowExecutionForBackground(
		domainID,
		shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
	)
	s.Nil(err)
	msBuilder := &mockMutableState{}
	defer msBuilder.AssertExpectations(s.T())
	context.(*workflowExecutionContextImpl).msBuilder = msBuilder
	release(nil)
	request := &h.SyncActivityRequest{
		DomainId:          common.StringPtr(domainID),
		WorkflowId:        common.StringPtr(workflowID),
		RunId:             common.StringPtr(runID),
		Version:           common.Int64Ptr(version),
		ScheduledId:       common.Int64Ptr(scheduleID),
		ScheduledTime:     common.Int64Ptr(scheduledTime.UnixNano()),
		StartedId:         common.Int64Ptr(startedID),
		StartedTime:       common.Int64Ptr(startedTime.UnixNano()),
		Attempt:           common.Int32Ptr(attempt),
		LastHeartbeatTime: common.Int64Ptr(heartBeatUpdatedTime.UnixNano()),
		Details:           details,
	}
	msBuilder.On("StartTransaction", mock.Anything).Return(false, nil).Once()
	msBuilder.On("IsWorkflowExecutionRunning").Return(true)
	msBuilder.On("GetNextEventID").Return(nextEventID)
	var versionHistories *persistence.VersionHistories
	msBuilder.On("GetVersionHistories").Return(versionHistories)
	msBuilder.On("GetReplicationState").Return(&persistence.ReplicationState{})
	s.mockDomainCache.On("GetDomainByID", domainID).Return(
		cache.NewGlobalDomainCacheEntryForTest(
			&persistence.DomainInfo{ID: domainID, Name: domainName},
			&persistence.DomainConfig{Retention: 1},
			&persistence.DomainReplicationConfig{
				ActiveClusterName: cluster.TestCurrentClusterName,
				Clusters: []*persistence.ClusterReplicationConfig{
					{ClusterName: cluster.TestCurrentClusterName},
					{ClusterName: cluster.TestAlternativeClusterName},
				},
			},
			version,
			nil,
		), nil,
	)
	activityInfo := &persistence.ActivityInfo{
		Version:    version,
		ScheduleID: scheduleID,
		Attempt:    attempt - 1,
	}
	msBuilder.On("GetActivityInfo", scheduleID).Return(activityInfo, true)
	s.mockClusterMetadata.On("IsVersionFromSameCluster", version, activityInfo.Version).Return(true)

	expectedErr := errors.New("this is error is used to by pass lots of mocking")
	msBuilder.On("ReplicateActivityInfo", request, true).Return(expectedErr)

	err = s.activityReplicator.SyncActivity(ctx.Background(), request)
	s.Equal(expectedErr, err)
}

func (s *activityReplicatorSuite) TestSyncActivity_ActivityRunning_Update_LargerVersion() {
	domainName := "some random domain name"
	domainID := testDomainID
	workflowID := "some random workflow ID"
	runID := uuid.New()
	version := int64(100)
	scheduleID := int64(144)
	scheduledTime := time.Now()
	startedID := scheduleID + 1
	startedTime := scheduledTime.Add(time.Minute)
	heartBeatUpdatedTime := startedTime.Add(time.Minute)
	attempt := int32(100)
	details := []byte("some random activity heartbeat progress")

	nextEventID := scheduleID + 10

	context, release, err := s.historyCache.getOrCreateWorkflowExecutionForBackground(
		domainID,
		shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
	)
	s.Nil(err)
	msBuilder := &mockMutableState{}
	defer msBuilder.AssertExpectations(s.T())
	context.(*workflowExecutionContextImpl).msBuilder = msBuilder
	release(nil)
	request := &h.SyncActivityRequest{
		DomainId:          common.StringPtr(domainID),
		WorkflowId:        common.StringPtr(workflowID),
		RunId:             common.StringPtr(runID),
		Version:           common.Int64Ptr(version),
		ScheduledId:       common.Int64Ptr(scheduleID),
		ScheduledTime:     common.Int64Ptr(scheduledTime.UnixNano()),
		StartedId:         common.Int64Ptr(startedID),
		StartedTime:       common.Int64Ptr(startedTime.UnixNano()),
		Attempt:           common.Int32Ptr(attempt),
		LastHeartbeatTime: common.Int64Ptr(heartBeatUpdatedTime.UnixNano()),
		Details:           details,
	}
	msBuilder.On("StartTransaction", mock.Anything).Return(false, nil).Once()
	msBuilder.On("IsWorkflowExecutionRunning").Return(true)
	msBuilder.On("GetNextEventID").Return(nextEventID)
	var versionHistories *persistence.VersionHistories
	msBuilder.On("GetVersionHistories").Return(versionHistories)
	msBuilder.On("GetReplicationState").Return(&persistence.ReplicationState{})
	s.mockDomainCache.On("GetDomainByID", domainID).Return(
		cache.NewGlobalDomainCacheEntryForTest(
			&persistence.DomainInfo{ID: domainID, Name: domainName},
			&persistence.DomainConfig{Retention: 1},
			&persistence.DomainReplicationConfig{
				ActiveClusterName: cluster.TestCurrentClusterName,
				Clusters: []*persistence.ClusterReplicationConfig{
					{ClusterName: cluster.TestCurrentClusterName},
					{ClusterName: cluster.TestAlternativeClusterName},
				},
			},
			version,
			nil,
		), nil,
	)
	activityInfo := &persistence.ActivityInfo{
		Version:    version - 1,
		ScheduleID: scheduleID,
		Attempt:    attempt + 1,
	}
	msBuilder.On("GetActivityInfo", scheduleID).Return(activityInfo, true)
	s.mockClusterMetadata.On("IsVersionFromSameCluster", version, activityInfo.Version).Return(false)

	expectedErr := errors.New("this is error is used to by pass lots of mocking")
	msBuilder.On("ReplicateActivityInfo", request, true).Return(expectedErr)

	err = s.activityReplicator.SyncActivity(ctx.Background(), request)
	s.Equal(expectedErr, err)
}
