// The MIT License (MIT)
//
// Copyright (c) 2017-2022 Uber Technologies Inc.
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

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

	hydrator := NewTaskHydrator(testShardID, nil, log.NewNoop(), nil)
	actual := hydrator.HydrateFailoverMarkerTask(task)
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
		name         string
		task         persistence.ReplicationTaskInfo
		mutableState MutableState
		expectTask   *types.ReplicationTask
		expectErr    string
	}{
		{
			name: "hydrates sync activity task",
			task: task,
			mutableState: &fakeMutableState{
				isWorkflowExecutionRunning: true,
				versionHistories:           versionHistories,
				activityInfos:              map[int64]persistence.ActivityInfo{testScheduleID: activityInfo},
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
			name:         "workflow is not running - return nil, no error",
			task:         task,
			mutableState: &fakeMutableState{isWorkflowExecutionRunning: false},
			expectTask:   nil,
		},
		{
			name: "no activity info - return nil, no error",
			task: task,
			mutableState: &fakeMutableState{
				isWorkflowExecutionRunning: true,
				activityInfos:              map[int64]persistence.ActivityInfo{},
			},
			expectTask: nil,
		},
		{
			name:         "no mutable state - treat as workflow does not exist, return nil, no error",
			task:         task,
			mutableState: nil,
			expectTask:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hydrator := NewTaskHydrator(testShardID, nil, log.NewNoop(), nil)

			actualTask, err := hydrator.HydrateSyncActivityTask(context.Background(), tt.task, tt.mutableState)
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
		name           string
		task           persistence.ReplicationTaskInfo
		mutableState   MutableState
		prepareHistory func(hm *mocks.HistoryV2Manager)
		expectTask     *types.ReplicationTask
		expectErr      string
	}{
		{
			name:         "hydrates history with given branch token",
			task:         task,
			mutableState: &fakeMutableState{versionHistories: versionHistories},
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
			name:         "hydrates history with branch token from version histories",
			task:         taskWithoutBranchToken,
			mutableState: &fakeMutableState{versionHistories: versionHistories},
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
			name:         "no version histories - return nil, no error",
			task:         task,
			mutableState: &fakeMutableState{versionHistories: nil},
			expectTask:   nil,
		},
		{
			name:         "bad version histories - return error",
			task:         task,
			mutableState: &fakeMutableState{versionHistories: &persistence.VersionHistories{}},
			expectErr:    "version histories does not contains given item.",
		},
		{
			name:         "failed reading history - return error",
			task:         task,
			mutableState: &fakeMutableState{versionHistories: versionHistories},
			prepareHistory: func(hm *mocks.HistoryV2Manager) {
				hm.On("ReadRawHistoryBranch", mock.Anything, mock.Anything).Return(nil, errors.New("failed reading history"))
			},
			expectErr: "failed reading history",
		},
		{
			name:         "no mutable state - treat as workflow does not exist, return nil, no error",
			task:         task,
			mutableState: nil,
			expectTask:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := &mocks.HistoryV2Manager{}
			if tt.prepareHistory != nil {
				tt.prepareHistory(history)
			}

			hydrator := NewTaskHydrator(testShardID, history, log.NewNoop(), dynamicconfig.GetIntPropertyFn(5))

			actualTask, err := hydrator.HydrateHistoryReplicationTask(context.Background(), tt.task, tt.mutableState)
			if tt.expectErr != "" {
				assert.EqualError(t, err, tt.expectErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectTask, actualTask)
			}
		})
	}
}

type fakeMutableState struct {
	isWorkflowExecutionRunning bool
	versionHistories           *persistence.VersionHistories
	activityInfos              map[int64]persistence.ActivityInfo
}

func (ms fakeMutableState) IsWorkflowExecutionRunning() bool {
	return ms.isWorkflowExecutionRunning
}
func (ms fakeMutableState) GetActivityInfo(scheduleID int64) (*persistence.ActivityInfo, bool) {
	ai, ok := ms.activityInfos[scheduleID]
	return &ai, ok
}
func (ms fakeMutableState) GetVersionHistories() *persistence.VersionHistories {
	return ms.versionHistories
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
