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

package replication

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
)

type (
	taskAckManagerSuite struct {
		suite.Suite
		*require.Assertions

		controller       *gomock.Controller
		mockShard        *shard.TestContext
		mockDomainCache  *cache.MockDomainCache
		mockMutableState *execution.MockMutableState

		mockExecutionMgr *mocks.ExecutionManager
		mockHistoryMgr   *mocks.HistoryV2Manager

		logger log.Logger

		ackManager *taskAckManagerImpl
	}
)

func TestTaskAckManagerSuite(t *testing.T) {
	s := new(taskAckManagerSuite)
	suite.Run(t, s)
}

func (s *taskAckManagerSuite) SetupSuite() {

}

func (s *taskAckManagerSuite) TearDownSuite() {

}

func (s *taskAckManagerSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockMutableState = execution.NewMockMutableState(s.controller)

	s.mockShard = shard.NewTestContext(
		s.controller,
		&persistence.ShardInfo{
			ShardID:                 0,
			RangeID:                 1,
			TransferAckLevel:        0,
			ClusterReplicationLevel: make(map[string]int64),
		},
		config.NewForTest(),
	)

	s.mockDomainCache = s.mockShard.Resource.DomainCache
	s.mockExecutionMgr = s.mockShard.Resource.ExecutionMgr
	s.mockHistoryMgr = s.mockShard.Resource.HistoryMgr

	s.logger = s.mockShard.GetLogger()
	executionCache := execution.NewCache(s.mockShard)

	s.ackManager = NewTaskAckManager(
		s.mockShard,
		executionCache,
	).(*taskAckManagerImpl)
}

func (s *taskAckManagerSuite) TearDownTest() {
	s.controller.Finish()
	s.mockShard.Finish(s.T())
}

func (s *taskAckManagerSuite) TestReadTasksWithBatchSize_OK() {
	task := &persistence.ReplicationTaskInfo{
		DomainID: uuid.New(),
	}
	s.mockExecutionMgr.On("GetReplicationTasks", mock.Anything, mock.Anything).Return(&persistence.GetReplicationTasksResponse{
		Tasks:         []*persistence.ReplicationTaskInfo{task},
		NextPageToken: []byte{1},
	}, nil)

	taskInfo, hasMore, err := s.ackManager.readTasksWithBatchSize(context.Background(), 0, 1)
	s.NoError(err)
	s.True(hasMore)
	s.Len(taskInfo, 1)
	s.Equal(task.GetDomainID(), taskInfo[0].GetDomainID())
}

func (s *taskAckManagerSuite) TestReadTasksWithBatchSize_Error() {
	s.mockExecutionMgr.On("GetReplicationTasks", mock.Anything, mock.Anything).Return(nil, errors.New("test"))

	taskInfo, hasMore, err := s.ackManager.readTasksWithBatchSize(context.Background(), 0, 1)
	s.Error(err)
	s.False(hasMore)
	s.Len(taskInfo, 0)
}

func (s *taskAckManagerSuite) TestGetTasks() {
	domainID := uuid.New()
	workflowID := uuid.New()
	runID := uuid.New()
	clusterName := cluster.TestCurrentClusterName
	taskInfo := &persistence.ReplicationTaskInfo{
		TaskType:     persistence.ReplicationTaskTypeFailoverMarker,
		DomainID:     domainID,
		WorkflowID:   workflowID,
		RunID:        runID,
		FirstEventID: 6,
		Version:      1,
	}
	s.mockExecutionMgr.On("GetReplicationTasks", mock.Anything, mock.Anything).Return(&persistence.GetReplicationTasksResponse{
		Tasks:         []*persistence.ReplicationTaskInfo{taskInfo},
		NextPageToken: []byte{1},
	}, nil)
	s.mockShard.Resource.ShardMgr.On("UpdateShard", mock.Anything, mock.Anything).Return(nil)
	s.mockDomainCache.EXPECT().GetDomainByID(domainID).Return(cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{ID: domainID, Name: "domainName"},
		&persistence.DomainConfig{Retention: 1},
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
				{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		1,
	), nil).AnyTimes()

	_, err := s.ackManager.GetTasks(context.Background(), clusterName, 10)
	s.NoError(err)
	ackLevel := s.mockShard.GetClusterReplicationLevel(clusterName)
	s.Equal(int64(10), ackLevel)
}

/*
func (s *taskAckManagerSuite) TestGetTasks_ReturnDataErrors() {
	domainID := uuid.New()
	workflowID := uuid.New()
	runID := uuid.New()
	clusterName := cluster.TestCurrentClusterName
	taskID := int64(10)
	taskInfo := &persistence.ReplicationTaskInfo{
		TaskType:     persistence.ReplicationTaskTypeHistory,
		TaskID:       taskID + 1,
		DomainID:     domainID,
		WorkflowID:   workflowID,
		RunID:        runID,
		FirstEventID: 6,
		Version:      1,
	}
	versionHistories := &persistence.VersionHistories{
		CurrentVersionHistoryIndex: 0,
		Histories: []*persistence.VersionHistory{
			{
				BranchToken: []byte{1},
				Items: []*persistence.VersionHistoryItem{
					{
						EventID: 6,
						Version: 1,
					},
				},
			},
		},
	}
	workflowContext, release, _ := s.ackManager.executionCache.GetOrCreateWorkflowExecutionForBackground(
		domainID,
		types.WorkflowExecution{
			WorkflowID: workflowID,
			RunID:      runID,
		},
	)
	workflowContext.SetWorkflowExecution(s.mockMutableState)
	release(nil)
	s.mockMutableState.EXPECT().StartTransaction(gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	s.mockMutableState.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(domainID).Return(cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{ID: domainID, Name: "domainName"},
		&persistence.DomainConfig{Retention: 1},
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
				{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		1,
	), nil).AnyTimes()
	s.mockExecutionMgr.On("GetReplicationTasks", mock.Anything, mock.Anything).Return(&persistence.GetReplicationTasksResponse{
		Tasks:         []*persistence.ReplicationTaskInfo{taskInfo},
		NextPageToken: nil,
	}, nil)
	s.mockShard.Resource.ShardMgr.On("UpdateShard", mock.Anything, mock.Anything).Return(nil)
	// Test BadRequestError
	s.mockHistoryMgr.On("ReadRawHistoryBranch", mock.Anything, mock.Anything).Return(
		nil,
		&types.BadRequestError{},
	).Times(1)
	msg, err := s.ackManager.GetTasks(context.Background(), clusterName, taskID)
	s.NoError(err)
	s.Equal(taskID+1, msg.GetLastRetrievedMessageID())

	// Test InternalDataInconsistencyError
	s.mockHistoryMgr.On("ReadRawHistoryBranch", mock.Anything, mock.Anything).Return(
		nil,
		&types.InternalDataInconsistencyError{},
	).Times(1)
	msg, err = s.ackManager.GetTasks(context.Background(), clusterName, taskID)
	s.NoError(err)
	s.Equal(taskID+1, msg.GetLastRetrievedMessageID())

	// Test EntityNotExistsError
	s.mockHistoryMgr.On("ReadRawHistoryBranch", mock.Anything, mock.Anything).Return(
		nil,
		&types.EntityNotExistsError{},
	).Times(1)
	msg, err = s.ackManager.GetTasks(context.Background(), clusterName, taskID)
	s.NoError(err)
	s.Equal(taskID+1, msg.GetLastRetrievedMessageID())
}*/

func (s *taskAckManagerSuite) TestSkipTask_ReturnTrue() {
	domainID := uuid.New()
	domainEntity := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{ID: domainID, Name: "domainName"},
		&persistence.DomainConfig{Retention: 1},
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
				{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		1,
	)
	s.True(skipTask("test", domainEntity))
}

func (s *taskAckManagerSuite) TestSkipTask_ReturnFalse() {
	domainID := uuid.New()
	domainEntity := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{ID: domainID, Name: "domainName"},
		&persistence.DomainConfig{Retention: 1},
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
				{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		1,
	)
	s.False(skipTask(cluster.TestAlternativeClusterName, domainEntity))
}

func (s *taskAckManagerSuite) TestGetBatchSize_UpperLimit() {
	s.ackManager.lastTaskCreationTime = atomic.Value{}
	s.ackManager.lastTaskCreationTime.Store(time.Now().Add(time.Duration(-s.ackManager.maxAllowedLatencyFn()) * time.Second))
	size := s.ackManager.getBatchSize()
	s.Equal(s.ackManager.fetchTasksBatchSize(0), size)
}

func (s *taskAckManagerSuite) TestGetBatchSize_ValidRange() {
	s.ackManager.lastTaskCreationTime = atomic.Value{}
	s.ackManager.lastTaskCreationTime.Store(time.Now().Add(-8 * time.Second))
	size := s.ackManager.getBatchSize()
	s.True(minReadTaskSize+5 <= size)
}

func (s *taskAckManagerSuite) TestGetBatchSize_InvalidRange() {
	s.ackManager.lastTaskCreationTime = atomic.Value{}
	s.ackManager.lastTaskCreationTime.Store(time.Now().Add(time.Minute))
	size := s.ackManager.getBatchSize()
	s.Equal(minReadTaskSize, size)
}
