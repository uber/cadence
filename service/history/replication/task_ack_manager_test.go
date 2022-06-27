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
	"github.com/stretchr/testify/assert"
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

const (
	testShardID           = 0
	testDomainID          = "11111111-1111-1111-1111-111111111111"
	testWorkflowID        = "workflow-id"
	testRunID             = "22222222-2222-2222-2222-222222222222"
	testTaskID            = 111
	testCreationTime      = int64(333)
	testFirstEventID      = 6
	testNextEventID       = 8
	testVersion           = 456
	testScheduleID        = int64(10)
	testStartedID         = int64(11)
	testLastFailureReason = "failure-reason"
	testWorkerIdentity    = "worker-identity"
	testAttempt           = 42
)

var (
	testBlobTask                  = []byte{1, 2, 3}
	testBlobTaskNewRun            = []byte{4, 5, 6}
	testBlobTokenVersionHistory   = []byte{4, 5, 6}
	testBranchTokenTask           = []byte{91, 92, 93}
	testBranchTokenTaskNewRun     = []byte{94, 95, 96}
	testBranchTokenVersionHistory = []byte{97, 98, 99}
	testDetails                   = []byte{100, 101, 102}
	testLastFailureDetails        = []byte{103, 104, 105}
	testScheduleTime              = time.Now()
	testStartedTime               = time.Now()
	testHeartbeatTime             = time.Now()
)

func TestHydration_FailoverMarker(t *testing.T) {
	task := persistence.ReplicationTaskInfo{
		TaskType:     persistence.ReplicationTaskTypeFailoverMarker,
		DomainID:     testDomainID,
		TaskID:       testTaskID,
		Version:      testVersion,
		CreationTime: testCreationTime,
	}

	expected := types.ReplicationTask{
		TaskType:     types.ReplicationTaskTypeFailoverMarker.Ptr(),
		SourceTaskID: testTaskID,
		FailoverMarkerAttributes: &types.FailoverMarkerAttributes{
			DomainID:        testDomainID,
			FailoverVersion: testVersion,
		},
		CreationTime: common.Int64Ptr(testCreationTime),
	}

	ackManager, _, _ := setupTaskAckManager(t)
	actual, err := ackManager.toReplicationTask(context.Background(), &task)
	assert.NoError(t, err)
	assert.Equal(t, &expected, actual)
}

func TestHydration_SyncActivity(t *testing.T) {
	task := persistence.ReplicationTaskInfo{
		TaskType:     persistence.ReplicationTaskTypeSyncActivity,
		TaskID:       testTaskID,
		DomainID:     testDomainID,
		WorkflowID:   testWorkflowID,
		RunID:        testRunID,
		ScheduledID:  testScheduleID,
		CreationTime: testCreationTime,
	}

	versionHistories := &persistence.VersionHistories{
		CurrentVersionHistoryIndex: 0,
		Histories: []*persistence.VersionHistory{
			{
				BranchToken: testBranchTokenVersionHistory,
				Items: []*persistence.VersionHistoryItem{
					{EventID: testFirstEventID, Version: testVersion},
				},
			},
		},
	}
	activityInfo := persistence.ActivityInfo{
		Version:                  testVersion,
		ScheduleID:               testScheduleID,
		ScheduledTime:            testScheduleTime,
		StartedID:                testStartedID,
		StartedTime:              testStartedTime,
		DomainID:                 testDomainID,
		LastHeartBeatUpdatedTime: testHeartbeatTime,
		Details:                  testDetails,
		Attempt:                  testAttempt,
		LastFailureReason:        testLastFailureReason,
		LastFailureDetails:       testLastFailureDetails,
		LastWorkerIdentity:       testWorkerIdentity,
	}

	tests := []struct {
		name                string
		task                persistence.ReplicationTaskInfo
		prepareMutableState func(*execution.MockMutableState)
		expectTask          *types.ReplicationTask
		expectErr           string
	}{
		{
			name: "hydrates sync activity task",
			task: task,
			prepareMutableState: func(ms *execution.MockMutableState) {
				ms.EXPECT().IsWorkflowExecutionRunning().Return(true)
				ms.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
				ms.EXPECT().GetActivityInfo(testScheduleID).Return(&activityInfo, true)
			},
			expectTask: &types.ReplicationTask{
				TaskType:     types.ReplicationTaskTypeSyncActivity.Ptr(),
				SourceTaskID: testTaskID,
				CreationTime: common.Int64Ptr(testCreationTime),
				SyncActivityTaskAttributes: &types.SyncActivityTaskAttributes{
					DomainID:           testDomainID,
					WorkflowID:         testWorkflowID,
					RunID:              testRunID,
					Version:            testVersion,
					ScheduledID:        testScheduleID,
					ScheduledTime:      common.Int64Ptr(testScheduleTime.UnixNano()),
					StartedID:          testStartedID,
					StartedTime:        common.Int64Ptr(testStartedTime.UnixNano()),
					LastHeartbeatTime:  common.Int64Ptr(testHeartbeatTime.UnixNano()),
					Details:            testDetails,
					Attempt:            testAttempt,
					LastFailureReason:  common.StringPtr(testLastFailureReason),
					LastWorkerIdentity: testWorkerIdentity,
					LastFailureDetails: testLastFailureDetails,
					VersionHistory: &types.VersionHistory{
						Items:       []*types.VersionHistoryItem{{EventID: testFirstEventID, Version: testVersion}},
						BranchToken: testBranchTokenVersionHistory,
					},
				},
			},
		},
		{
			name: "workflow is not running - return nil, no error",
			task: task,
			prepareMutableState: func(ms *execution.MockMutableState) {
				ms.EXPECT().IsWorkflowExecutionRunning().Return(false)
				ms.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
			},
			expectTask: nil,
		},
		{
			name: "no activity info - return nil, no error",
			task: task,
			prepareMutableState: func(ms *execution.MockMutableState) {
				ms.EXPECT().IsWorkflowExecutionRunning().Return(true)
				ms.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
				ms.EXPECT().GetActivityInfo(testScheduleID).Return(nil, false)
			},
			expectTask: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ackManager, ms, _ := setupTaskAckManager(t)

			if tt.prepareMutableState != nil {
				tt.prepareMutableState(ms)
			}

			actualTask, err := ackManager.toReplicationTask(context.Background(), &tt.task)
			if tt.expectErr != "" {
				assert.EqualError(t, err, tt.expectErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectTask, actualTask)
			}
		})
	}
}

func TestHydration_History(t *testing.T) {
	task := persistence.ReplicationTaskInfo{
		TaskType:          persistence.ReplicationTaskTypeHistory,
		TaskID:            testTaskID,
		DomainID:          testDomainID,
		WorkflowID:        testWorkflowID,
		RunID:             testRunID,
		FirstEventID:      testFirstEventID,
		NextEventID:       testNextEventID,
		BranchToken:       testBranchTokenTask,
		NewRunBranchToken: testBranchTokenTaskNewRun,
		Version:           testVersion,
		CreationTime:      testCreationTime,
	}
	taskWithoutBranchToken := task
	taskWithoutBranchToken.BranchToken = nil

	versionHistories := &persistence.VersionHistories{
		CurrentVersionHistoryIndex: 0,
		Histories: []*persistence.VersionHistory{
			{
				BranchToken: testBranchTokenVersionHistory,
				Items: []*persistence.VersionHistoryItem{
					{EventID: testFirstEventID, Version: testVersion},
				},
			},
		},
	}

	tests := []struct {
		name                string
		task                persistence.ReplicationTaskInfo
		prepareMutableState func(*execution.MockMutableState)
		prepareHistory      func(*mocks.HistoryV2Manager)
		expectTask          *types.ReplicationTask
		expectErr           string
	}{
		{
			name: "hydrates history with given branch token",
			task: task,
			prepareMutableState: func(ms *execution.MockMutableState) {
				ms.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
				ms.EXPECT().GetActivityInfo(gomock.Any()).Return(nil, false)
			},
			prepareHistory: func(hm *mocks.HistoryV2Manager) {
				mockHistory(hm, testFirstEventID, testNextEventID, testBranchTokenTask, testBlobTask)
				mockHistory(hm, 1, 2, testBranchTokenTaskNewRun, testBlobTaskNewRun)
			},
			expectTask: &types.ReplicationTask{
				TaskType:     types.ReplicationTaskTypeHistoryV2.Ptr(),
				SourceTaskID: testTaskID,
				CreationTime: common.Int64Ptr(testCreationTime),
				HistoryTaskV2Attributes: &types.HistoryTaskV2Attributes{
					DomainID:            testDomainID,
					WorkflowID:          testWorkflowID,
					RunID:               testRunID,
					VersionHistoryItems: []*types.VersionHistoryItem{{EventID: testFirstEventID, Version: testVersion}},
					Events:              &types.DataBlob{EncodingType: types.EncodingTypeJSON.Ptr(), Data: testBlobTask},
					NewRunEvents:        &types.DataBlob{EncodingType: types.EncodingTypeJSON.Ptr(), Data: testBlobTaskNewRun},
				},
			},
		},
		{
			name: "hydrates history with branch token from version histories",
			task: taskWithoutBranchToken,
			prepareMutableState: func(ms *execution.MockMutableState) {
				ms.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
				ms.EXPECT().GetActivityInfo(gomock.Any()).Return(nil, false)
			},
			prepareHistory: func(hm *mocks.HistoryV2Manager) {
				mockHistory(hm, testFirstEventID, testNextEventID, testBranchTokenVersionHistory, testBlobTokenVersionHistory)
				mockHistory(hm, 1, 2, testBranchTokenTaskNewRun, testBlobTaskNewRun)
			},
			expectTask: &types.ReplicationTask{
				TaskType:     types.ReplicationTaskTypeHistoryV2.Ptr(),
				SourceTaskID: testTaskID,
				CreationTime: common.Int64Ptr(testCreationTime),
				HistoryTaskV2Attributes: &types.HistoryTaskV2Attributes{
					DomainID:            testDomainID,
					WorkflowID:          testWorkflowID,
					RunID:               testRunID,
					VersionHistoryItems: []*types.VersionHistoryItem{{EventID: testFirstEventID, Version: testVersion}},
					Events:              &types.DataBlob{EncodingType: types.EncodingTypeJSON.Ptr(), Data: testBlobTokenVersionHistory},
					NewRunEvents:        &types.DataBlob{EncodingType: types.EncodingTypeJSON.Ptr(), Data: testBlobTaskNewRun},
				},
			},
		},
		{
			name: "no version histories - return nil, no error",
			task: task,
			prepareMutableState: func(ms *execution.MockMutableState) {
				ms.EXPECT().GetVersionHistories().Return(nil).AnyTimes()
				ms.EXPECT().GetActivityInfo(gomock.Any()).Return(nil, false)
			},
			expectTask: nil,
		},
		{
			name: "bad version histories - return error",
			task: task,
			prepareMutableState: func(ms *execution.MockMutableState) {
				ms.EXPECT().GetVersionHistories().Return(&persistence.VersionHistories{}).AnyTimes()
				ms.EXPECT().GetActivityInfo(gomock.Any()).Return(nil, false)
			},
			expectErr: "version histories does not contains given item.",
		},
		{
			name: "failed reading history - return error",
			task: task,
			prepareMutableState: func(ms *execution.MockMutableState) {
				ms.EXPECT().GetVersionHistories().Return(versionHistories).AnyTimes()
				ms.EXPECT().GetActivityInfo(gomock.Any()).Return(nil, false)
			},
			prepareHistory: func(hm *mocks.HistoryV2Manager) {
				hm.On("ReadRawHistoryBranch", mock.Anything, mock.Anything).Return(nil, errors.New("failed reading history"))
			},
			expectErr: "failed reading history",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ackManager, ms, hm := setupTaskAckManager(t)

			if tt.prepareMutableState != nil {
				tt.prepareMutableState(ms)
			}
			if tt.prepareHistory != nil {
				tt.prepareHistory(hm)
			}

			actualTask, err := ackManager.toReplicationTask(context.Background(), &tt.task)
			if tt.expectErr != "" {
				assert.EqualError(t, err, tt.expectErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectTask, actualTask)
			}
		})
	}
}

func mockHistory(hm *mocks.HistoryV2Manager, minID, maxID int64, branchToken []byte, returnedBlob []byte) {
	historyResponse := persistence.ReadRawHistoryBranchResponse{
		HistoryEventBlobs: []*persistence.DataBlob{
			{Encoding: common.EncodingTypeJSON, Data: returnedBlob},
		},
		Size: 1,
	}
	hm.On("ReadRawHistoryBranch", mock.Anything, &persistence.ReadHistoryBranchRequest{
		BranchToken: branchToken,
		MinEventID:  minID,
		MaxEventID:  maxID,
		PageSize:    5,
		ShardID:     common.IntPtr(testShardID),
	}).Return(&historyResponse, nil)
}

func setupTaskAckManager(t *testing.T) (*taskAckManagerImpl, *execution.MockMutableState, *mocks.HistoryV2Manager) {
	controller := gomock.NewController(t)
	mockMutableState := execution.NewMockMutableState(controller)

	mockShard := shard.NewTestContext(
		controller,
		&persistence.ShardInfo{
			ShardID:                 testShardID,
			RangeID:                 1,
			TransferAckLevel:        0,
			ClusterReplicationLevel: make(map[string]int64),
		},
		config.NewForTest(),
	)

	mockDomainCache := mockShard.Resource.DomainCache
	mockHistoryMgr := mockShard.Resource.HistoryMgr
	executionCache := execution.NewCache(mockShard)

	workflowContext, release, _ := executionCache.GetOrCreateWorkflowExecutionForBackground(testDomainID, types.WorkflowExecution{testWorkflowID, testRunID})
	workflowContext.SetWorkflowExecution(mockMutableState)
	release(nil)

	mockMutableState.EXPECT().StartTransaction(gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	mockDomainCache.EXPECT().GetDomainByID(testDomainID).Return(cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{ID: testDomainID, Name: "domainName"},
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

	ackManager := NewTaskAckManager(mockShard, executionCache).(*taskAckManagerImpl)
	return ackManager, mockMutableState, mockHistoryMgr
}
