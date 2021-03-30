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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
)

type (
	taskAckManagerSuite struct {
		suite.Suite
		*require.Assertions

		controller          *gomock.Controller
		mockShard           *shard.TestContext
		mockDomainCache     *cache.MockDomainCache
		mockMutableState    *execution.MockMutableState
		mockClusterMetadata *cluster.MockMetadata

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
	s.mockClusterMetadata = s.mockShard.Resource.ClusterMetadata
	s.mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockClusterMetadata.EXPECT().IsGlobalDomainEnabled().Return(true).AnyTimes()

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

func (s *taskAckManagerSuite) TestGetPaginationFunc() {
	firstEventID := int64(0)
	nextEventID := int64(1)
	var branchToken []byte
	shardID := 0
	historyCount := 0
	pagingFunc := s.ackManager.getPaginationFunc(context.Background(), firstEventID, nextEventID, branchToken, shardID, &historyCount)

	pageToken := []byte{1}
	event := &types.HistoryEvent{
		EventID: 1,
	}
	s.mockHistoryMgr.On("ReadHistoryBranch", mock.Anything, mock.Anything).Return(&persistence.ReadHistoryBranchResponse{
		HistoryEvents:    []*types.HistoryEvent{event},
		NextPageToken:    pageToken,
		Size:             1,
		LastFirstEventID: 1,
	}, nil)
	events, token, err := pagingFunc(nil)
	s.NoError(err)
	s.Equal(pageToken, token)
	s.Len(events, 1)
	s.Equal(events[0].(*types.HistoryEvent), event)
	s.Equal(historyCount, 1)
}

func (s *taskAckManagerSuite) TestGetAllHistory_OK() {
	firstEventID := int64(0)
	nextEventID := int64(1)
	var branchToken []byte
	event := &types.HistoryEvent{
		EventID: 1,
	}

	s.mockHistoryMgr.On("ReadHistoryBranch", mock.Anything, mock.Anything).Return(&persistence.ReadHistoryBranchResponse{
		HistoryEvents:    []*types.HistoryEvent{event},
		NextPageToken:    nil,
		Size:             1,
		LastFirstEventID: 1,
	}, nil)

	history, err := s.ackManager.getAllHistory(context.Background(), firstEventID, nextEventID, branchToken)
	s.NoError(err)
	s.Len(history.GetEvents(), 1)
	s.Equal(event, history.GetEvents()[0])
}

func (s *taskAckManagerSuite) TestGetAllHistory_Error() {
	firstEventID := int64(0)
	nextEventID := int64(1)
	var branchToken []byte
	s.mockHistoryMgr.On("ReadHistoryBranch", mock.Anything, mock.Anything).Return(nil, errors.New("test"))

	history, err := s.ackManager.getAllHistory(context.Background(), firstEventID, nextEventID, branchToken)
	s.Error(err)
	s.Nil(history)
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

func (s *taskAckManagerSuite) TestIsNewRunNDCEnabled_True() {
	domainID := uuid.New()
	workflowID := uuid.New()
	runID := uuid.New()
	workflowContext, release, _ := s.ackManager.executionCache.GetOrCreateWorkflowExecutionForBackground(
		domainID,
		types.WorkflowExecution{
			WorkflowID: workflowID,
			RunID:      runID,
		},
	)
	workflowContext.SetWorkflowExecution(s.mockMutableState)
	release(nil)

	s.mockMutableState.EXPECT().StartTransaction(gomock.Any()).Return(false, nil).Times(1)
	s.mockMutableState.EXPECT().IsWorkflowExecutionRunning().Return(false).AnyTimes()
	s.mockMutableState.EXPECT().GetVersionHistories().Return(&persistence.VersionHistories{})
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
		nil,
	), nil).AnyTimes()

	isNDC, err := s.ackManager.isNewRunNDCEnabled(
		context.Background(),
		domainID,
		workflowID,
		runID,
	)
	s.NoError(err)
	s.True(isNDC)
}

func (s *taskAckManagerSuite) TestIsNewRunNDCEnabled_False() {
	domainID := uuid.New()
	workflowID := uuid.New()
	runID := uuid.New()
	workflowContext, release, _ := s.ackManager.executionCache.GetOrCreateWorkflowExecutionForBackground(
		domainID,
		types.WorkflowExecution{
			WorkflowID: workflowID,
			RunID:      runID,
		},
	)
	workflowContext.SetWorkflowExecution(s.mockMutableState)
	release(nil)

	s.mockMutableState.EXPECT().StartTransaction(gomock.Any()).Return(false, nil).Times(1)
	s.mockMutableState.EXPECT().IsWorkflowExecutionRunning().Return(false).AnyTimes()
	s.mockMutableState.EXPECT().GetVersionHistories().Return(nil)
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
		nil,
	), nil).AnyTimes()

	isNDC, err := s.ackManager.isNewRunNDCEnabled(
		context.Background(),
		domainID,
		workflowID,
		runID,
	)
	s.NoError(err)
	s.False(isNDC)
}

func (s *taskAckManagerSuite) TestGetVersionHistoryItems_Error() {
	_, _, err := getVersionHistoryItems(nil, 0, 0)
	s.Error(err)
}

func (s *taskAckManagerSuite) TestGetVersionHistoryItems_OK() {
	versionHistories := &persistence.VersionHistories{
		CurrentVersionHistoryIndex: 0,
		Histories: []*persistence.VersionHistory{
			{
				BranchToken: []byte{1},
				Items: []*persistence.VersionHistoryItem{
					{
						EventID: 1,
						Version: 1,
					},
				},
			},
		},
	}
	versionHistory, branchToken, err := getVersionHistoryItems(versionHistories, 1, 1)
	s.NoError(err)
	s.Equal(versionHistories.Histories[0].GetBranchToken(), branchToken)
	s.Equal(versionHistories.Histories[0].Items[0].GetVersion(), versionHistory[0].GetVersion())
}

func (s *taskAckManagerSuite) TestGetEventsBlob_OK() {
	branchToken := []byte{}
	firstEventID := int64(1)
	nextEventID := int64(2)

	s.mockHistoryMgr.On("ReadRawHistoryBranch", mock.Anything, mock.Anything).Return(
		&persistence.ReadRawHistoryBranchResponse{
			HistoryEventBlobs: []*persistence.DataBlob{
				{
					Encoding: common.EncodingTypeJSON,
					Data:     []byte{},
				},
			},
			Size: 1,
		}, nil)
	_, err := s.ackManager.getEventsBlob(context.Background(), branchToken, firstEventID, nextEventID)
	s.NoError(err)
}

func (s *taskAckManagerSuite) TestGetEventsBlob_Errors() {
	branchToken := []byte{}
	firstEventID := int64(1)
	nextEventID := int64(2)

	s.mockHistoryMgr.On("ReadRawHistoryBranch", mock.Anything, mock.Anything).Return(
		&persistence.ReadRawHistoryBranchResponse{
			HistoryEventBlobs: []*persistence.DataBlob{},
			Size:              0,
		}, nil)
	_, err := s.ackManager.getEventsBlob(context.Background(), branchToken, firstEventID, nextEventID)
	s.Error(err)

	s.mockHistoryMgr.On("ReadRawHistoryBranch", mock.Anything, mock.Anything).Return(
		&persistence.ReadRawHistoryBranchResponse{
			HistoryEventBlobs: []*persistence.DataBlob{
				{
					Encoding: common.EncodingTypeJSON,
					Data:     []byte{},
				},
				{
					Encoding: common.EncodingTypeJSON,
					Data:     []byte{},
				},
			},
			Size: 2,
		}, nil)
	_, err = s.ackManager.getEventsBlob(context.Background(), branchToken, firstEventID, nextEventID)
	s.Error(err)

	s.mockHistoryMgr.On("ReadRawHistoryBranch", mock.Anything, mock.Anything).Return(nil, errors.New("test"))
	_, err = s.ackManager.getEventsBlob(context.Background(), branchToken, firstEventID, nextEventID)
	s.Error(err)
}

func (s *taskAckManagerSuite) TestProcessReplication_OK() {
	domainID := uuid.New()
	workflowID := uuid.New()
	runID := uuid.New()
	taskInfo := &persistence.ReplicationTaskInfo{
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
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

	s.mockMutableState.EXPECT().StartTransaction(gomock.Any()).Return(false, nil).AnyTimes()
	s.mockMutableState.EXPECT().IsWorkflowExecutionRunning().Return(false).AnyTimes()
	s.mockMutableState.EXPECT().GetVersionHistories().Return(nil).AnyTimes()
	s.mockMutableState.EXPECT().GetActivityInfo(gomock.Any()).Return(nil, false).AnyTimes()
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
		nil,
	), nil).AnyTimes()

	_, err := s.ackManager.processReplication(
		context.Background(),
		false,
		taskInfo,
		func(
			activityInfo *persistence.ActivityInfo,
			versionHistories *persistence.VersionHistories,
		) (*types.ReplicationTask, error) {
			_, release, err := s.ackManager.executionCache.GetOrCreateWorkflowExecutionForBackground(
				domainID,
				types.WorkflowExecution{
					WorkflowID: workflowID,
					RunID:      runID,
				},
			)
			s.NoError(err)
			defer release(nil)
			return nil, nil
		},
	)
	s.NoError(err)
	_, release, err = s.ackManager.executionCache.GetOrCreateWorkflowExecutionForBackground(
		domainID,
		types.WorkflowExecution{
			WorkflowID: workflowID,
			RunID:      runID,
		},
	)
	s.NoError(err)
	release(nil)

	_, err = s.ackManager.processReplication(
		context.Background(),
		true,
		taskInfo,
		func(
			activityInfo *persistence.ActivityInfo,
			versionHistories *persistence.VersionHistories,
		) (*types.ReplicationTask, error) {
			_, release, err := s.ackManager.executionCache.GetOrCreateWorkflowExecutionForBackground(
				domainID,
				types.WorkflowExecution{
					WorkflowID: workflowID,
					RunID:      runID,
				},
			)
			s.NoError(err)
			release(nil)
			return nil, nil
		},
	)
	s.NoError(err)
	_, release, err = s.ackManager.executionCache.GetOrCreateWorkflowExecutionForBackground(
		domainID,
		types.WorkflowExecution{
			WorkflowID: workflowID,
			RunID:      runID,
		},
	)
	s.NoError(err)
	release(nil)
}

func (s *taskAckManagerSuite) TestProcessReplication_Error() {
	domainID := uuid.New()
	workflowID := uuid.New()
	runID := uuid.New()
	taskInfo := &persistence.ReplicationTaskInfo{
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
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

	s.mockMutableState.EXPECT().StartTransaction(gomock.Any()).Return(false, errors.New("test")).Times(1)
	s.mockMutableState.EXPECT().IsWorkflowExecutionRunning().Return(false).AnyTimes()
	s.mockMutableState.EXPECT().GetVersionHistories().Return(nil).AnyTimes()
	s.mockMutableState.EXPECT().GetActivityInfo(gomock.Any()).Return(nil, false).AnyTimes()
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
		nil,
	), nil).AnyTimes()

	_, err := s.ackManager.processReplication(
		context.Background(),
		true,
		taskInfo,
		func(
			activityInfo *persistence.ActivityInfo,
			versionHistories *persistence.VersionHistories,
		) (*types.ReplicationTask, error) {
			return nil, nil
		},
	)
	s.Error(err)
	_, release, err = s.ackManager.executionCache.GetOrCreateWorkflowExecutionForBackground(
		domainID,
		types.WorkflowExecution{
			WorkflowID: workflowID,
			RunID:      runID,
		},
	)
	s.NoError(err)
	release(nil)
}

func (s *taskAckManagerSuite) TestGenerateFailoverMarkerTask() {
	domainID := uuid.New()
	taskInfo := &persistence.ReplicationTaskInfo{
		DomainID:     domainID,
		TaskID:       1,
		Version:      2,
		CreationTime: 3,
	}
	task := s.ackManager.generateFailoverMarkerTask(taskInfo)
	s.Equal(task.GetSourceTaskID(), int64(1))
	s.NotNil(task.GetFailoverMarkerAttributes())
	s.Equal(types.ReplicationTaskTypeFailoverMarker, task.GetTaskType())
	s.Equal(domainID, task.GetFailoverMarkerAttributes().GetDomainID())
	s.Equal(int64(2), task.GetFailoverMarkerAttributes().GetFailoverVersion())
	s.Equal(int64(3), task.GetCreationTime())
}

func (s *taskAckManagerSuite) TestGenerateSyncActivityTask_OK() {
	domainID := uuid.New()
	workflowID := uuid.New()
	runID := uuid.New()
	taskInfo := &persistence.ReplicationTaskInfo{
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
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
	activityInfo := &persistence.ActivityInfo{
		Version:                  1,
		ScheduleID:               5,
		ScheduledTime:            time.Now(),
		StartedID:                6,
		StartedTime:              time.Now(),
		DomainID:                 domainID,
		RequestID:                uuid.New(),
		LastHeartBeatUpdatedTime: time.Now(),
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

	s.mockMutableState.EXPECT().StartTransaction(gomock.Any()).Return(false, nil).AnyTimes()
	s.mockMutableState.EXPECT().IsWorkflowExecutionRunning().Return(true).AnyTimes()
	s.mockMutableState.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
	s.mockMutableState.EXPECT().GetActivityInfo(gomock.Any()).Return(activityInfo, true).AnyTimes()
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
		nil,
	), nil).AnyTimes()

	task, err := s.ackManager.generateSyncActivityTask(context.Background(), taskInfo)
	s.NoError(err)
	s.NotNil(task)
	s.NotNil(task.SyncActivityTaskAttributes)
	s.Equal(types.ReplicationTaskTypeSyncActivity, task.GetTaskType())
	s.Equal(activityInfo.DomainID, task.GetSyncActivityTaskAttributes().GetDomainID())
	s.Equal(activityInfo.ScheduleID, task.GetSyncActivityTaskAttributes().GetScheduledID())
	s.Equal(activityInfo.ScheduledTime.UnixNano(), task.GetSyncActivityTaskAttributes().GetScheduledTime())
	s.Equal(activityInfo.StartedID, task.GetSyncActivityTaskAttributes().GetStartedID())
	s.Equal(activityInfo.StartedTime.UnixNano(), task.GetSyncActivityTaskAttributes().GetStartedTime())
	s.Equal(activityInfo.LastHeartBeatUpdatedTime.UnixNano(), task.GetSyncActivityTaskAttributes().GetLastHeartbeatTime())
	s.Equal(activityInfo.Version, task.GetSyncActivityTaskAttributes().GetVersion())
	s.Equal(versionHistories.Histories[0].ToInternalType(), task.GetSyncActivityTaskAttributes().GetVersionHistory())
}

func (s *taskAckManagerSuite) TestGenerateSyncActivityTask_Empty() {
	domainID := uuid.New()
	workflowID := uuid.New()
	runID := uuid.New()
	taskInfo := &persistence.ReplicationTaskInfo{
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
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

	s.mockMutableState.EXPECT().StartTransaction(gomock.Any()).Return(false, nil).AnyTimes()
	s.mockMutableState.EXPECT().IsWorkflowExecutionRunning().Return(true).AnyTimes()
	s.mockMutableState.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
	s.mockMutableState.EXPECT().GetActivityInfo(gomock.Any()).Return(nil, false).AnyTimes()
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
		nil,
	), nil).AnyTimes()

	task, err := s.ackManager.generateSyncActivityTask(context.Background(), taskInfo)
	s.NoError(err)
	s.Nil(task)
}

func (s *taskAckManagerSuite) TestGenerateHistoryReplicationTask() {
	domainID := uuid.New()
	workflowID := uuid.New()
	runID := uuid.New()
	taskInfo := &persistence.ReplicationTaskInfo{
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

	s.mockMutableState.EXPECT().StartTransaction(gomock.Any()).Return(false, nil).AnyTimes()
	s.mockMutableState.EXPECT().IsWorkflowExecutionRunning().Return(true).AnyTimes()
	s.mockMutableState.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
	s.mockMutableState.EXPECT().GetActivityInfo(gomock.Any()).Return(nil, false).AnyTimes()
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
		nil,
	), nil).AnyTimes()
	s.mockHistoryMgr.On("ReadRawHistoryBranch", mock.Anything, mock.Anything).Return(
		&persistence.ReadRawHistoryBranchResponse{
			HistoryEventBlobs: []*persistence.DataBlob{
				{
					Encoding: common.EncodingTypeJSON,
					Data:     []byte{},
				},
			},
			Size: 1,
		},
		nil,
	)

	task, err := s.ackManager.generateHistoryReplicationTask(context.Background(), taskInfo)
	s.NoError(err)
	s.NotNil(task)
	s.NotNil(task.HistoryTaskV2Attributes)
	s.Equal(types.ReplicationTaskTypeHistoryV2, task.GetTaskType())
}

func (s *taskAckManagerSuite) TestToReplicationTask_FailoverMarker() {
	domainID := uuid.New()
	workflowID := uuid.New()
	runID := uuid.New()
	taskInfo := &persistence.ReplicationTaskInfo{
		TaskType:     persistence.ReplicationTaskTypeFailoverMarker,
		DomainID:     domainID,
		WorkflowID:   workflowID,
		RunID:        runID,
		FirstEventID: 6,
		Version:      1,
	}

	task, err := s.ackManager.toReplicationTask(context.Background(), taskInfo)
	s.NoError(err)
	s.NotNil(task)
	s.Equal(types.ReplicationTaskTypeFailoverMarker, task.GetTaskType())
}

func (s *taskAckManagerSuite) TestToReplicationTask_SyncActivity() {
	domainID := uuid.New()
	workflowID := uuid.New()
	runID := uuid.New()
	taskInfo := &persistence.ReplicationTaskInfo{
		TaskType:   persistence.ReplicationTaskTypeSyncActivity,
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
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
	activityInfo := &persistence.ActivityInfo{
		Version:                  1,
		ScheduleID:               5,
		ScheduledTime:            time.Now(),
		StartedID:                6,
		StartedTime:              time.Now(),
		DomainID:                 domainID,
		RequestID:                uuid.New(),
		LastHeartBeatUpdatedTime: time.Now(),
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

	s.mockMutableState.EXPECT().StartTransaction(gomock.Any()).Return(false, nil).AnyTimes()
	s.mockMutableState.EXPECT().IsWorkflowExecutionRunning().Return(true).AnyTimes()
	s.mockMutableState.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
	s.mockMutableState.EXPECT().GetActivityInfo(gomock.Any()).Return(activityInfo, true).AnyTimes()
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
		nil,
	), nil).AnyTimes()

	task, err := s.ackManager.toReplicationTask(context.Background(), taskInfo)
	s.NoError(err)
	s.NotNil(task)
	s.NotNil(task.SyncActivityTaskAttributes)
	s.Equal(types.ReplicationTaskTypeSyncActivity, task.GetTaskType())
	s.Equal(activityInfo.DomainID, task.GetSyncActivityTaskAttributes().GetDomainID())
	s.Equal(activityInfo.ScheduleID, task.GetSyncActivityTaskAttributes().GetScheduledID())
	s.Equal(activityInfo.ScheduledTime.UnixNano(), task.GetSyncActivityTaskAttributes().GetScheduledTime())
	s.Equal(activityInfo.StartedID, task.GetSyncActivityTaskAttributes().GetStartedID())
	s.Equal(activityInfo.StartedTime.UnixNano(), task.GetSyncActivityTaskAttributes().GetStartedTime())
	s.Equal(activityInfo.LastHeartBeatUpdatedTime.UnixNano(), task.GetSyncActivityTaskAttributes().GetLastHeartbeatTime())
	s.Equal(activityInfo.Version, task.GetSyncActivityTaskAttributes().GetVersion())
	s.Equal(versionHistories.Histories[0].ToInternalType(), task.GetSyncActivityTaskAttributes().GetVersionHistory())
}

func (s *taskAckManagerSuite) TestToReplicationTask_History() {
	domainID := uuid.New()
	workflowID := uuid.New()
	runID := uuid.New()
	taskInfo := &persistence.ReplicationTaskInfo{
		TaskType:     persistence.ReplicationTaskTypeHistory,
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

	s.mockMutableState.EXPECT().StartTransaction(gomock.Any()).Return(false, nil).AnyTimes()
	s.mockMutableState.EXPECT().IsWorkflowExecutionRunning().Return(true).AnyTimes()
	s.mockMutableState.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
	s.mockMutableState.EXPECT().GetActivityInfo(gomock.Any()).Return(nil, false).AnyTimes()
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
		nil,
	), nil).AnyTimes()
	s.mockHistoryMgr.On("ReadRawHistoryBranch", mock.Anything, mock.Anything).Return(
		&persistence.ReadRawHistoryBranchResponse{
			HistoryEventBlobs: []*persistence.DataBlob{
				{
					Encoding: common.EncodingTypeJSON,
					Data:     []byte{},
				},
			},
			Size: 1,
		},
		nil,
	)

	task, err := s.ackManager.generateHistoryReplicationTask(context.Background(), taskInfo)
	s.NoError(err)
	s.NotNil(task)
	s.NotNil(task.HistoryTaskV2Attributes)
	s.Equal(types.ReplicationTaskTypeHistoryV2, task.GetTaskType())
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
		nil,
	), nil).AnyTimes()

	_, err := s.ackManager.GetTasks(context.Background(), clusterName, 10)
	s.NoError(err)
	ackLevel := s.mockShard.GetClusterReplicationLevel(clusterName)
	s.Equal(int64(10), ackLevel)
}

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
	s.mockMutableState.EXPECT().StartTransaction(gomock.Any()).Return(false, nil).AnyTimes()
	s.mockMutableState.EXPECT().IsWorkflowExecutionRunning().Return(true).AnyTimes()
	s.mockMutableState.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
	s.mockMutableState.EXPECT().GetActivityInfo(gomock.Any()).Return(nil, false).AnyTimes()
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
		nil,
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
}

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
		nil,
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
		nil,
	)
	s.False(skipTask(cluster.TestAlternativeClusterName, domainEntity))
}
