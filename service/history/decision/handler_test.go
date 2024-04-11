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

package decision

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/client"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	commonConfig "github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/query"
	"github.com/uber/cadence/service/history/shard"
)

const (
	_testDomainUUID        = "00000000000000000000000000000001"
	_testInvalidDomainUUID = "some-invalid-UUID"
	_testDomainName        = "test-domain"
	_testWorkflowID        = "test-wfID"
	_testRunID             = "00000000000000000000000000000002"
	_testCluster           = "test-cluster"
	_testShardID           = 0
)

type (
	DecisionHandlerSuite struct {
		*require.Assertions
		suite.Suite

		controller       *gomock.Controller
		mockMutableState *execution.MockMutableState

		decisionHandler       *handlerImpl
		queryRegistry         query.Registry
		localDomainCacheEntry *cache.DomainCacheEntry
		clusterMetadata       cluster.Metadata
	}
)

func TestDecisionHandlerSuite(t *testing.T) {
	suite.Run(t, new(DecisionHandlerSuite))
}

func (s *DecisionHandlerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.controller = gomock.NewController(s.T())
	domainInfo := &persistence.DomainInfo{
		ID:   _testDomainUUID,
		Name: _testDomainName,
	}
	s.localDomainCacheEntry = cache.NewLocalDomainCacheEntryForTest(domainInfo, &persistence.DomainConfig{}, _testCluster)
	s.clusterMetadata = cluster.NewMetadata(0, _testCluster, _testCluster, map[string]commonConfig.ClusterInformation{}, func(domain string) bool {
		return false
	}, metrics.NewClient(tally.NoopScope, metrics.History), testlogger.New(s.T()))
	s.decisionHandler = &handlerImpl{
		versionChecker: client.NewVersionChecker(),
		metricsClient:  metrics.NewClient(tally.NoopScope, metrics.History),
		config:         config.NewForTest(),
		logger:         testlogger.New(s.T()),
		timeSource:     clock.NewRealTimeSource(),
	}
	s.queryRegistry = s.constructQueryRegistry(10)
	s.mockMutableState = execution.NewMockMutableState(s.controller)
	workflowInfo := &persistence.WorkflowExecutionInfo{
		WorkflowID: constants.TestWorkflowID,
		RunID:      constants.TestRunID,
	}
	s.mockMutableState.EXPECT().GetExecutionInfo().Return(workflowInfo).AnyTimes()
}

func (s *DecisionHandlerSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *DecisionHandlerSuite) TestNewHandler() {
	shardContext := shard.NewMockContext(s.controller)
	tokenSerializer := common.NewMockTaskTokenSerializer(s.controller)
	shardContext.EXPECT().GetConfig().Times(1).Return(&config.Config{})
	shardContext.EXPECT().GetLogger().Times(2).Return(testlogger.New(s.T()))
	shardContext.EXPECT().GetTimeSource().Times(1)
	shardContext.EXPECT().GetDomainCache().Times(2)
	shardContext.EXPECT().GetMetricsClient().Times(2)
	shardContext.EXPECT().GetThrottledLogger().Times(1).Return(testlogger.New(s.T()))
	h := NewHandler(shardContext, &execution.Cache{}, tokenSerializer)
	s.NotNil(h)
	s.Equal("handlerImpl", reflect.ValueOf(h).Elem().Type().Name())
}

func (s *DecisionHandlerSuite) TestHandleDecisionTaskScheduled() {
	tests := []struct {
		name            string
		domainID        string
		mutablestate    *persistence.WorkflowMutableState
		isfirstDecision bool
		expectCalls     func(shardContext *shard.MockContext)
		expectErr       bool
	}{
		{
			name:     "test HandleDecisionTaskScheduled - fail to retrieve domain From ID",
			domainID: _testInvalidDomainUUID,
			mutablestate: &persistence.WorkflowMutableState{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{},
			},
			expectCalls: func(shardContext *shard.MockContext) {},
			expectErr:   true,
		},
		{
			name:     "test HandleDecisionTaskScheduled - success",
			domainID: _testDomainUUID,
			mutablestate: &persistence.WorkflowMutableState{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{},
			},
			expectCalls: func(shardContext *shard.MockContext) {
				shardContext.EXPECT().GetEventsCache().Times(1).Return(events.NewMockCache(s.controller))
			},
			expectErr: false,
		},
		{
			name:     "test HandleDecisionTaskScheduled - with completed workflow",
			domainID: _testDomainUUID,
			mutablestate: &persistence.WorkflowMutableState{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					// WorkflowStateCompleted = 2 from persistence WorkflowExecutionInfo.IsRunning()
					State: 2,
				},
			},
			expectCalls: func(shardContext *shard.MockContext) {
				shardContext.EXPECT().GetEventsCache().Times(1).Return(events.NewMockCache(s.controller))
			},
			expectErr: true,
		},
		{
			name:     "test HandleDecisionTaskScheduled - with failure to get start event",
			domainID: _testDomainUUID,
			mutablestate: &persistence.WorkflowMutableState{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					// execution has no event yet
					DecisionScheduleID: -23,
					LastProcessedEvent: -23,
				},
			},
			expectCalls: func(shardContext *shard.MockContext) {
				eventsCache := events.NewMockCache(s.controller)
				shardContext.EXPECT().GetEventsCache().Times(1).Return(eventsCache)
				eventsCache.EXPECT().
					GetEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, &persistence.TimeoutError{Msg: "failed to get start event: request timeout"})
				shardContext.EXPECT().GetShardID().Return(_testShardID).Times(1)
			},
			expectErr: true,
		},
		{
			name:     "test HandleDecisionTaskScheduled - first decision task scheduled failure",
			domainID: _testDomainUUID,
			mutablestate: &persistence.WorkflowMutableState{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					DecisionScheduleID: -23,
					LastProcessedEvent: -23,
				},
				BufferedEvents: append([]*types.HistoryEvent{}, &types.HistoryEvent{}),
			},
			expectCalls: func(shardContext *shard.MockContext) {
				eventsCache := events.NewMockCache(s.controller)
				shardContext.EXPECT().GetEventsCache().Times(1).Return(eventsCache)
				eventsCache.EXPECT().
					GetEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(&types.HistoryEvent{}, nil)
				shardContext.EXPECT().GetShardID().Return(_testShardID).Times(1)
				shardContext.EXPECT().GenerateTransferTaskIDs(gomock.Any()).Times(1).Return([]int64{}, errors.New("some random error to avoid going too deep in call stack unrelated to this unit"))
			},
			expectErr:       true,
			isfirstDecision: true,
		},
		{
			name:     "test HandleDecisionTaskScheduled - first decision task scheduled success",
			domainID: _testDomainUUID,
			mutablestate: &persistence.WorkflowMutableState{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					DecisionScheduleID: -23,
					LastProcessedEvent: -23,
				},
			},
			expectCalls: func(shardContext *shard.MockContext) {
				eventsCache := events.NewMockCache(s.controller)
				shardContext.EXPECT().GetEventsCache().Times(1).Return(eventsCache)
				eventsCache.EXPECT().
					GetEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(&types.HistoryEvent{}, nil)
				shardContext.EXPECT().GetShardID().Return(_testShardID).Times(1)
				shardContext.EXPECT().GenerateTransferTaskIDs(gomock.Any()).Times(1).Return([]int64{}, errors.New("some random error to avoid going too deep in call stack unrelated to this unit"))
			},
			expectErr:       true,
			isfirstDecision: true,
		},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			request := &types.ScheduleDecisionTaskRequest{
				DomainUUID: test.domainID,
				WorkflowExecution: &types.WorkflowExecution{
					WorkflowID: _testWorkflowID,
					RunID:      _testRunID,
				},
				IsFirstDecision: test.isfirstDecision,
			}
			shardContext := shard.NewMockContext(s.controller)
			test.expectCalls(shardContext)
			s.expectHandleDecisionTaskScheduledCalls(test.domainID, test.mutablestate, shardContext)

			handler := *s.decisionHandler
			handler.executionCache = execution.NewCache(shardContext)
			handler.shard = shardContext
			err := handler.HandleDecisionTaskScheduled(context.Background(), request)
			s.Equal(test.expectErr, err != nil)
		})
	}
}

func (s *DecisionHandlerSuite) TestHandleBufferedQueries_ClientNotSupports() {
	s.mockMutableState.EXPECT().GetQueryRegistry().Return(s.queryRegistry)
	s.assertQueryCounts(s.queryRegistry, 10, 0, 0, 0)
	s.decisionHandler.handleBufferedQueries(s.mockMutableState, client.GoSDK, "0.0.0", nil, false, constants.TestGlobalDomainEntry, false)
	s.assertQueryCounts(s.queryRegistry, 0, 0, 0, 10)
}

func (s *DecisionHandlerSuite) TestHandleBufferedQueries_HeartbeatDecision() {
	s.mockMutableState.EXPECT().GetQueryRegistry().Return(s.queryRegistry)
	s.assertQueryCounts(s.queryRegistry, 10, 0, 0, 0)
	queryResults := s.constructQueryResults(s.queryRegistry.GetBufferedIDs()[0:5], 10)
	s.decisionHandler.handleBufferedQueries(s.mockMutableState, client.GoSDK, client.GoWorkerConsistentQueryVersion, queryResults, false, constants.TestGlobalDomainEntry, true)
	s.assertQueryCounts(s.queryRegistry, 10, 0, 0, 0)
}

func (s *DecisionHandlerSuite) TestHandleBufferedQueries_NewDecisionTask() {
	s.mockMutableState.EXPECT().GetQueryRegistry().Return(s.queryRegistry)
	s.assertQueryCounts(s.queryRegistry, 10, 0, 0, 0)
	queryResults := s.constructQueryResults(s.queryRegistry.GetBufferedIDs()[0:5], 10)
	s.decisionHandler.handleBufferedQueries(s.mockMutableState, client.GoSDK, client.GoWorkerConsistentQueryVersion, queryResults, true, constants.TestGlobalDomainEntry, false)
	s.assertQueryCounts(s.queryRegistry, 5, 5, 0, 0)
}

func (s *DecisionHandlerSuite) TestHandleBufferedQueries_NoNewDecisionTask() {
	s.mockMutableState.EXPECT().GetQueryRegistry().Return(s.queryRegistry)
	s.assertQueryCounts(s.queryRegistry, 10, 0, 0, 0)
	queryResults := s.constructQueryResults(s.queryRegistry.GetBufferedIDs()[0:5], 10)
	s.decisionHandler.handleBufferedQueries(s.mockMutableState, client.GoSDK, client.GoWorkerConsistentQueryVersion, queryResults, false, constants.TestGlobalDomainEntry, false)
	s.assertQueryCounts(s.queryRegistry, 0, 5, 5, 0)
}

func (s *DecisionHandlerSuite) TestHandleBufferedQueries_QueryTooLarge() {
	s.mockMutableState.EXPECT().GetQueryRegistry().Return(s.queryRegistry)
	s.assertQueryCounts(s.queryRegistry, 10, 0, 0, 0)
	bufferedIDs := s.queryRegistry.GetBufferedIDs()
	queryResults := s.constructQueryResults(bufferedIDs[0:5], 10)
	largeQueryResults := s.constructQueryResults(bufferedIDs[5:10], 10*1024*1024)
	for k, v := range largeQueryResults {
		queryResults[k] = v
	}
	s.decisionHandler.handleBufferedQueries(s.mockMutableState, client.GoSDK, client.GoWorkerConsistentQueryVersion, queryResults, false, constants.TestGlobalDomainEntry, false)
	s.assertQueryCounts(s.queryRegistry, 0, 5, 0, 5)
}

func (s *DecisionHandlerSuite) constructQueryResults(ids []string, resultSize int) map[string]*types.WorkflowQueryResult {
	results := make(map[string]*types.WorkflowQueryResult)
	for _, id := range ids {
		results[id] = &types.WorkflowQueryResult{
			ResultType: types.QueryResultTypeAnswered.Ptr(),
			Answer:     make([]byte, resultSize),
		}
	}
	return results
}

func (s *DecisionHandlerSuite) constructQueryRegistry(numQueries int) query.Registry {
	queryRegistry := query.NewRegistry()
	for i := 0; i < numQueries; i++ {
		queryRegistry.BufferQuery(&types.WorkflowQuery{})
	}
	return queryRegistry
}

func (s *DecisionHandlerSuite) assertQueryCounts(queryRegistry query.Registry, buffered, completed, unblocked, failed int) {
	s.Len(queryRegistry.GetBufferedIDs(), buffered)
	s.Len(queryRegistry.GetCompletedIDs(), completed)
	s.Len(queryRegistry.GetUnblockedIDs(), unblocked)
	s.Len(queryRegistry.GetFailedIDs(), failed)
}

func (s *DecisionHandlerSuite) expectHandleDecisionTaskScheduledCalls(domainID string, state *persistence.WorkflowMutableState, shardContex *shard.MockContext) {
	workflowExecutionResponse := &persistence.GetWorkflowExecutionResponse{
		State:             state,
		MutableStateStats: &persistence.MutableStateStats{},
	}
	workflowExecutionResponse.State.ExecutionStats = &persistence.ExecutionStats{}
	workflowExecutionResponse.State.ExecutionInfo.DomainID = domainID
	workflowExecutionResponse.State.ExecutionInfo.WorkflowID = _testWorkflowID
	workflowExecutionResponse.State.ExecutionInfo.RunID = _testRunID
	shardContextConfig := config.NewForTest()
	shardContextLogger := testlogger.New(s.T())
	shardContextTimeSource := clock.NewMockedTimeSource()
	shardContextMetricClient := metrics.NewClient(tally.NoopScope, metrics.History)
	domainCacheMock := cache.NewMockDomainCache(s.controller)

	shardContex.EXPECT().GetWorkflowExecution(gomock.Any(), gomock.Any()).AnyTimes().Return(workflowExecutionResponse, nil)
	shardContex.EXPECT().GetConfig().AnyTimes().Return(shardContextConfig)
	shardContex.EXPECT().GetLogger().AnyTimes().Return(shardContextLogger)
	shardContex.EXPECT().GetTimeSource().AnyTimes().Return(shardContextTimeSource)
	shardContex.EXPECT().GetDomainCache().AnyTimes().Return(domainCacheMock)
	shardContex.EXPECT().GetClusterMetadata().AnyTimes().Return(s.clusterMetadata)
	shardContex.EXPECT().GetMetricsClient().AnyTimes().Return(shardContextMetricClient)
	domainCacheMock.EXPECT().GetDomainByID(domainID).AnyTimes().Return(s.localDomainCacheEntry, nil)
	domainCacheMock.EXPECT().GetDomainName(domainID).AnyTimes().Return(_testDomainName, nil)
	shardContex.EXPECT().GetExecutionManager().Times(1)
}
