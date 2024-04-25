// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package shard

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/resource"
)

type (
	contextTestSuite struct {
		suite.Suite
		*require.Assertions

		controller       *gomock.Controller
		mockResource     *resource.Test
		mockShardManager *mocks.ShardManager

		metricsClient metrics.Client
		logger        log.Logger

		context *contextImpl
	}
)

func TestContextSuite(t *testing.T) {
	s := new(contextTestSuite)
	suite.Run(t, s)
}

func (s *contextTestSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockResource = resource.NewTest(s.T(), s.controller, metrics.History)
	s.mockShardManager = s.mockResource.ShardMgr

	s.metricsClient = metrics.NewClient(tally.NoopScope, metrics.History)
	s.logger = testlogger.New(s.T())

	s.context = s.newContext()
}

func (s *contextTestSuite) newContext() *contextImpl {
	eventsCache := events.NewMockCache(s.controller)
	config := config.NewForTest()
	shardInfo := &persistence.ShardInfo{
		ShardID: 0,
		RangeID: 1,
		// the following fields will be initialized
		// when acquiring the shard if they are nil
		ClusterTransferAckLevel: make(map[string]int64),
		ClusterTimerAckLevel:    make(map[string]time.Time),
		TransferProcessingQueueStates: &types.ProcessingQueueStates{
			StatesByCluster: make(map[string][]*types.ProcessingQueueState),
		},
		CrossClusterProcessingQueueStates: &types.ProcessingQueueStates{
			StatesByCluster: make(map[string][]*types.ProcessingQueueState),
		},
		TimerProcessingQueueStates: &types.ProcessingQueueStates{
			StatesByCluster: make(map[string][]*types.ProcessingQueueState),
		},
	}
	context := &contextImpl{
		Resource:                  s.mockResource,
		shardID:                   shardInfo.ShardID,
		rangeID:                   shardInfo.RangeID,
		shardInfo:                 shardInfo,
		executionManager:          s.mockResource.ExecutionMgr,
		config:                    config,
		logger:                    s.logger,
		throttledLogger:           s.logger,
		transferSequenceNumber:    1,
		transferMaxReadLevel:      0,
		maxTransferSequenceNumber: 100000,
		timerMaxReadLevelMap:      make(map[string]time.Time),
		remoteClusterCurrentTime:  make(map[string]time.Time),
		eventsCache:               eventsCache,
	}
	return context
}

func (s *contextTestSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *contextTestSuite) TestRenewRangeLockedSuccess() {
	s.mockShardManager.On("UpdateShard", mock.Anything, mock.Anything).Once().Return(nil)

	err := s.context.renewRangeLocked(false)
	s.NoError(err)
}

func (s *contextTestSuite) TestRenewRangeLockedSuccessAfterRetries() {
	s.mockShardManager.On("UpdateShard", mock.Anything, mock.Anything).Return(nil)

	err := s.context.renewRangeLocked(false)
	s.NoError(err)
}

func (s *contextTestSuite) TestRenewRangeLockedRetriesExceeded() {
	someError := errors.New("some error")
	s.mockShardManager.On("UpdateShard", mock.Anything, mock.Anything).Return(someError)

	err := s.context.renewRangeLocked(false)
	s.Error(err)
}

func (s *contextTestSuite) TestReplicateFailoverMarkersSuccess() {
	s.mockResource.ExecutionMgr.On("CreateFailoverMarkerTasks", mock.Anything, mock.Anything).Once().Return(nil)

	markers := make([]*persistence.FailoverMarkerTask, 0)
	err := s.context.ReplicateFailoverMarkers(context.Background(), markers)
	s.NoError(err)
}

func (s *contextTestSuite) TestGetAndUpdateProcessingQueueStates() {
	clusterName := cluster.TestCurrentClusterName
	var initialQueueStates [][]*types.ProcessingQueueState
	initialQueueStates = append(initialQueueStates, s.context.GetTransferProcessingQueueStates(clusterName))
	initialQueueStates = append(initialQueueStates, s.context.GetCrossClusterProcessingQueueStates(clusterName))
	initialQueueStates = append(initialQueueStates, s.context.GetTimerProcessingQueueStates(clusterName))
	for _, queueStates := range initialQueueStates {
		s.Len(queueStates, 1)
		s.Zero(queueStates[0].GetLevel())
		ackLevel := queueStates[0].GetAckLevel()
		if ackLevel != 0 {
			// for timer queue
			s.Equal(time.Time{}.UnixNano(), ackLevel)
		}
		s.Equal(int64(math.MaxInt64), queueStates[0].GetMaxLevel())
		s.Equal(&types.DomainFilter{ReverseMatch: true}, queueStates[0].GetDomainFilter())
	}

	s.mockShardManager.On("UpdateShard", mock.Anything, mock.Anything).Return(nil)
	updatedTransferQueueStates := []*types.ProcessingQueueState{
		{
			Level:    common.Int32Ptr(0),
			AckLevel: common.Int64Ptr(123),
			MaxLevel: common.Int64Ptr(math.MaxInt64),
			DomainFilter: &types.DomainFilter{
				ReverseMatch: true,
			},
		},
	}
	updatedCrossClusterQueueStates := []*types.ProcessingQueueState{
		{
			Level:    common.Int32Ptr(0),
			AckLevel: common.Int64Ptr(456),
			MaxLevel: common.Int64Ptr(math.MaxInt64),
			DomainFilter: &types.DomainFilter{
				ReverseMatch: true,
				DomainIDs:    []string{"testDomainID"},
			},
		},
		{
			Level:    common.Int32Ptr(1),
			AckLevel: common.Int64Ptr(123),
			MaxLevel: common.Int64Ptr(math.MaxInt64),
			DomainFilter: &types.DomainFilter{
				ReverseMatch: false,
				DomainIDs:    []string{"testDomainID"},
			},
		},
	}
	updatedTimerQueueStates := []*types.ProcessingQueueState{
		{
			Level:    common.Int32Ptr(0),
			AckLevel: common.Int64Ptr(time.Now().UnixNano()),
			MaxLevel: common.Int64Ptr(math.MaxInt64),
			DomainFilter: &types.DomainFilter{
				ReverseMatch: true,
			},
		},
	}
	err := s.context.UpdateTransferProcessingQueueStates(clusterName, updatedTransferQueueStates)
	s.NoError(err)
	err = s.context.UpdateCrossClusterProcessingQueueStates(clusterName, updatedCrossClusterQueueStates)
	s.NoError(err)
	err = s.context.UpdateTimerProcessingQueueStates(clusterName, updatedTimerQueueStates)
	s.NoError(err)

	s.Equal(updatedTransferQueueStates, s.context.GetTransferProcessingQueueStates(clusterName))
	s.Equal(updatedCrossClusterQueueStates, s.context.GetCrossClusterProcessingQueueStates(clusterName))
	s.Equal(updatedTimerQueueStates, s.context.GetTimerProcessingQueueStates(clusterName))

	// check if cluster ack level for transfer and timer is backfilled for backward compatibility
	s.Equal(updatedTransferQueueStates[0].GetAckLevel(), s.context.GetTransferClusterAckLevel(clusterName))
	s.Equal(time.Unix(0, updatedTimerQueueStates[0].GetAckLevel()), s.context.GetTimerClusterAckLevel(clusterName))
}

func TestGetWorkflowExecution(t *testing.T) {
	testCases := []struct {
		name           string
		closedAt       *time.Time
		request        *persistence.GetWorkflowExecutionRequest
		mockSetup      func(*mocks.ExecutionManager)
		expectedResult *persistence.GetWorkflowExecutionResponse
		expectedError  error
	}{
		{
			name: "Success",
			request: &persistence.GetWorkflowExecutionRequest{
				DomainID:  "testDomain",
				Execution: types.WorkflowExecution{WorkflowID: "testWorkflowID", RunID: "testRunID"},
			},
			mockSetup: func(mgr *mocks.ExecutionManager) {
				mgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(&persistence.GetWorkflowExecutionResponse{
					State: &persistence.WorkflowMutableState{
						ExecutionInfo: &persistence.WorkflowExecutionInfo{
							DomainID:   "testDomain",
							WorkflowID: "testWorkflowID",
							RunID:      "testRunID",
						},
					},
				}, nil)
			},
			expectedResult: &persistence.GetWorkflowExecutionResponse{
				State: &persistence.WorkflowMutableState{
					ExecutionInfo: &persistence.WorkflowExecutionInfo{
						DomainID:   "testDomain",
						WorkflowID: "testWorkflowID",
						RunID:      "testRunID",
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "Error",
			request: &persistence.GetWorkflowExecutionRequest{
				DomainID:  "testDomain",
				Execution: types.WorkflowExecution{WorkflowID: "testWorkflowID", RunID: "testRunID"},
			},
			mockSetup: func(mgr *mocks.ExecutionManager) {
				mgr.On("GetWorkflowExecution", mock.Anything, mock.Anything).Return(nil, errors.New("some random error"))
			},
			expectedResult: nil,
			expectedError:  errors.New("some random error"),
		},
		{
			name:     "Shard closed",
			closedAt: common.TimePtr(time.Now().Add(-time.Minute)),
			request: &persistence.GetWorkflowExecutionRequest{
				DomainID:  "testDomain",
				Execution: types.WorkflowExecution{WorkflowID: "testWorkflowID", RunID: "testRunID"},
			},
			mockSetup:      func(mgr *mocks.ExecutionManager) {},
			expectedResult: nil,
			expectedError:  ErrShardClosed,
		},
		{
			name:     "Shard recently closed",
			closedAt: common.TimePtr(time.Now()),
			request: &persistence.GetWorkflowExecutionRequest{
				DomainID:  "testDomain",
				Execution: types.WorkflowExecution{WorkflowID: "testWorkflowID", RunID: "testRunID"},
			},
			mockSetup:      func(mgr *mocks.ExecutionManager) {},
			expectedResult: nil,
			expectedError:  ErrShardRecentlyClosed,
		},
	}

	for _, tc := range testCases {
		mockExecutionMgr := &mocks.ExecutionManager{}
		shardContext := &contextImpl{
			executionManager: mockExecutionMgr,
			shardInfo: &persistence.ShardInfo{
				RangeID: 12,
			},
		}
		shardContext.closedAt.Store(tc.closedAt)
		tc.mockSetup(mockExecutionMgr)

		result, err := shardContext.GetWorkflowExecution(context.Background(), tc.request)
		assert.Equal(t, tc.expectedResult, result)
		assert.Equal(t, tc.expectedError, err)
	}
}

func TestCloseShard(t *testing.T) {
	closeCallback := make(chan struct{})

	mockExecutionMgr := &mocks.ExecutionManager{}
	shardContext := &contextImpl{
		executionManager: mockExecutionMgr,
		shardInfo: &persistence.ShardInfo{
			RangeID: 12,
		},
		closeCallback: func(i int, item *historyShardsItem) {
			close(closeCallback)
		},
	}
	shardContext.closeShard()

	select {
	case <-closeCallback:
	case <-time.After(time.Second):
		assert.Fail(t, "closeCallback not called")
	}

	assert.WithinDuration(t, time.Now(), *shardContext.closedAt.Load(), time.Second)
}

func TestShardClosedGuard(t *testing.T) {
	shardContext := &contextImpl{
		shardInfo: &persistence.ShardInfo{},
	}

	testCases := []struct {
		name string
		call func() error
	}{
		{
			name: "GetWorkflowExecution",
			call: func() error {
				_, err := shardContext.GetWorkflowExecution(
					context.Background(),
					&persistence.GetWorkflowExecutionRequest{RangeID: 0},
				)
				return err
			},
		},
		{
			name: "CreateWorkflowExecution",
			call: func() error {
				_, err := shardContext.CreateWorkflowExecution(context.Background(), nil)
				return err
			},
		},
		{
			name: "UpdateWorkflowExecution",
			call: func() error {
				_, err := shardContext.UpdateWorkflowExecution(context.Background(), nil)
				return err
			},
		},
		{
			name: "ConflictResolveWorkflowExecution",
			call: func() error {
				_, err := shardContext.ConflictResolveWorkflowExecution(context.Background(), nil)
				return err
			},
		},
		{
			name: "AppendHistoryV2Events",
			call: func() error {
				_, err := shardContext.AppendHistoryV2Events(context.Background(), nil, "", types.WorkflowExecution{})
				return err
			},
		},
		{
			name: "renewRangeLocked",
			call: func() error {
				return shardContext.renewRangeLocked(false)
			},
		},
		{
			name: "persistShardInfoLocked",
			call: func() error {
				return shardContext.persistShardInfoLocked(false)
			},
		},
		{
			name: "ReplicateFailoverMarkers",
			call: func() error {
				return shardContext.ReplicateFailoverMarkers(context.Background(), nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shardContext.closedAt.Store(common.TimePtr(time.Now()))
			err := tc.call()
			assert.Equal(t, ErrShardRecentlyClosed, err)

			shardContext.closedAt.Store(common.TimePtr(time.Now().Add(-time.Minute)))
			err = tc.call()
			assert.Equal(t, ErrShardClosed, err)
		})
	}
}
