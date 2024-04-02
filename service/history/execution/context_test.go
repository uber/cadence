// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

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

package execution

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	hcommon "github.com/uber/cadence/service/history/common"
	"github.com/uber/cadence/service/history/engine"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/shard"
)

func TestIsOperationPossiblySuccessfulError(t *testing.T) {
	assert.False(t, isOperationPossiblySuccessfulError(nil))
	assert.False(t, isOperationPossiblySuccessfulError(&types.WorkflowExecutionAlreadyStartedError{}))
	assert.False(t, isOperationPossiblySuccessfulError(&persistence.WorkflowExecutionAlreadyStartedError{}))
	assert.False(t, isOperationPossiblySuccessfulError(&persistence.CurrentWorkflowConditionFailedError{}))
	assert.False(t, isOperationPossiblySuccessfulError(&persistence.ConditionFailedError{}))
	assert.False(t, isOperationPossiblySuccessfulError(&types.ServiceBusyError{}))
	assert.False(t, isOperationPossiblySuccessfulError(&types.LimitExceededError{}))
	assert.False(t, isOperationPossiblySuccessfulError(&persistence.ShardOwnershipLostError{}))
	assert.True(t, isOperationPossiblySuccessfulError(&persistence.TimeoutError{}))
	assert.False(t, isOperationPossiblySuccessfulError(NewConflictError(t, &persistence.ConditionFailedError{})))
	assert.True(t, isOperationPossiblySuccessfulError(context.DeadlineExceeded))
}

func TestMergeContinueAsNewReplicationTasks(t *testing.T) {
	testCases := []struct {
		name                    string
		updateMode              persistence.UpdateWorkflowMode
		currentWorkflowMutation *persistence.WorkflowMutation
		newWorkflowSnapshot     *persistence.WorkflowSnapshot
		wantErr                 bool
		assertErr               func(*testing.T, error)
	}{
		{
			name: "current workflow does not continue as new",
			currentWorkflowMutation: &persistence.WorkflowMutation{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					CloseStatus: persistence.WorkflowCloseStatusCompleted,
				},
			},
			wantErr: false,
		},
		{
			name: "update workflow as zombie and continue as new without new zombie workflow",
			currentWorkflowMutation: &persistence.WorkflowMutation{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					CloseStatus: persistence.WorkflowCloseStatusContinuedAsNew,
				},
			},
			updateMode: persistence.UpdateWorkflowModeBypassCurrent,
			wantErr:    false,
		},
		{
			name: "continue as new on the passive side",
			currentWorkflowMutation: &persistence.WorkflowMutation{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					CloseStatus: persistence.WorkflowCloseStatusContinuedAsNew,
				},
			},
			updateMode: persistence.UpdateWorkflowModeUpdateCurrent,
			wantErr:    false,
		},
		{
			name: "continue as new on the active side, but new workflow is not provided",
			currentWorkflowMutation: &persistence.WorkflowMutation{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					CloseStatus: persistence.WorkflowCloseStatusContinuedAsNew,
				},
				ReplicationTasks: []persistence.Task{
					&persistence.HistoryReplicationTask{},
				},
			},
			updateMode: persistence.UpdateWorkflowModeUpdateCurrent,
			wantErr:    true,
			assertErr: func(t *testing.T, err error) {
				assert.IsType(t, &types.InternalServiceError{}, err)
				assert.Contains(t, err.Error(), "unable to find replication task from new workflow for continue as new replication")
			},
		},
		{
			name: "continue as new on the active side, but new workflow has no replication task",
			currentWorkflowMutation: &persistence.WorkflowMutation{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					CloseStatus: persistence.WorkflowCloseStatusContinuedAsNew,
				},
				ReplicationTasks: []persistence.Task{
					&persistence.HistoryReplicationTask{},
				},
			},
			newWorkflowSnapshot: &persistence.WorkflowSnapshot{},
			updateMode:          persistence.UpdateWorkflowModeUpdateCurrent,
			wantErr:             true,
			assertErr: func(t *testing.T, err error) {
				assert.IsType(t, &types.InternalServiceError{}, err)
				assert.Contains(t, err.Error(), "unable to find replication task from new workflow for continue as new replication")
			},
		},
		{
			name: "continue as new on the active side, but current workflow has no history replication task",
			currentWorkflowMutation: &persistence.WorkflowMutation{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					CloseStatus: persistence.WorkflowCloseStatusContinuedAsNew,
				},
				ReplicationTasks: []persistence.Task{
					&persistence.SyncActivityTask{},
				},
			},
			newWorkflowSnapshot: &persistence.WorkflowSnapshot{
				ReplicationTasks: []persistence.Task{
					&persistence.HistoryReplicationTask{},
				},
			},
			updateMode: persistence.UpdateWorkflowModeUpdateCurrent,
			wantErr:    true,
			assertErr: func(t *testing.T, err error) {
				assert.IsType(t, &types.InternalServiceError{}, err)
				assert.Contains(t, err.Error(), "unable to find replication task from current workflow for continue as new replication")
			},
		},
		{
			name: "continue as new on the active side",
			currentWorkflowMutation: &persistence.WorkflowMutation{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					CloseStatus: persistence.WorkflowCloseStatusContinuedAsNew,
				},
				ReplicationTasks: []persistence.Task{
					&persistence.HistoryReplicationTask{},
				},
			},
			newWorkflowSnapshot: &persistence.WorkflowSnapshot{
				ReplicationTasks: []persistence.Task{
					&persistence.HistoryReplicationTask{},
				},
			},
			updateMode: persistence.UpdateWorkflowModeUpdateCurrent,
			wantErr:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := mergeContinueAsNewReplicationTasks(tc.updateMode, tc.currentWorkflowMutation, tc.newWorkflowSnapshot)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.assertErr != nil {
					tc.assertErr(t, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNotifyTasksFromWorkflowSnapshot(t *testing.T) {
	testCases := []struct {
		name             string
		workflowSnapShot *persistence.WorkflowSnapshot
		history          events.PersistedBlobs
		persistenceError bool
		mockSetup        func(*engine.MockEngine)
	}{
		{
			name: "Success case",
			workflowSnapShot: &persistence.WorkflowSnapshot{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					DomainID:   "test-domain-id",
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				},
				VersionHistories: &persistence.VersionHistories{
					CurrentVersionHistoryIndex: 0,
					Histories: []*persistence.VersionHistory{
						{
							BranchToken: []byte{1, 2, 3},
						},
					},
				},
				ActivityInfos: []*persistence.ActivityInfo{
					{
						Version:    1,
						ScheduleID: 11,
					},
				},
				TransferTasks: []persistence.Task{
					&persistence.ActivityTask{
						TaskList: "test-tl",
					},
				},
				TimerTasks: []persistence.Task{
					&persistence.ActivityTimeoutTask{
						Attempt: 10,
					},
				},
				CrossClusterTasks: []persistence.Task{
					&persistence.CrossClusterStartChildExecutionTask{
						StartChildExecutionTask: persistence.StartChildExecutionTask{
							TargetDomainID: "target-domain",
						},
						TargetCluster: "target",
					},
				},
				ReplicationTasks: []persistence.Task{
					&persistence.HistoryReplicationTask{
						FirstEventID: 1,
						NextEventID:  10,
					},
				},
			},
			history: events.PersistedBlobs{
				events.PersistedBlob{},
			},
			persistenceError: true,
			mockSetup: func(mockEngine *engine.MockEngine) {
				mockEngine.EXPECT().NotifyNewTransferTasks(&hcommon.NotifyTaskInfo{
					ExecutionInfo: &persistence.WorkflowExecutionInfo{
						DomainID:   "test-domain-id",
						WorkflowID: "test-workflow-id",
						RunID:      "test-run-id",
					},
					Tasks: []persistence.Task{
						&persistence.ActivityTask{
							TaskList: "test-tl",
						},
					},
					PersistenceError: true,
				})
				mockEngine.EXPECT().NotifyNewTimerTasks(&hcommon.NotifyTaskInfo{
					ExecutionInfo: &persistence.WorkflowExecutionInfo{
						DomainID:   "test-domain-id",
						WorkflowID: "test-workflow-id",
						RunID:      "test-run-id",
					},
					Tasks: []persistence.Task{
						&persistence.ActivityTimeoutTask{
							Attempt: 10,
						},
					},
					PersistenceError: true,
				})
				mockEngine.EXPECT().NotifyNewCrossClusterTasks(&hcommon.NotifyTaskInfo{
					ExecutionInfo: &persistence.WorkflowExecutionInfo{
						DomainID:   "test-domain-id",
						WorkflowID: "test-workflow-id",
						RunID:      "test-run-id",
					},
					Tasks: []persistence.Task{
						&persistence.CrossClusterStartChildExecutionTask{
							StartChildExecutionTask: persistence.StartChildExecutionTask{
								TargetDomainID: "target-domain",
							},
							TargetCluster: "target",
						},
					},
					PersistenceError: true,
				})
				mockEngine.EXPECT().NotifyNewReplicationTasks(&hcommon.NotifyTaskInfo{
					ExecutionInfo: &persistence.WorkflowExecutionInfo{
						DomainID:   "test-domain-id",
						WorkflowID: "test-workflow-id",
						RunID:      "test-run-id",
					},
					Tasks: []persistence.Task{
						&persistence.HistoryReplicationTask{
							FirstEventID: 1,
							NextEventID:  10,
						},
					},
					VersionHistories: &persistence.VersionHistories{
						CurrentVersionHistoryIndex: 0,
						Histories: []*persistence.VersionHistory{
							{
								BranchToken: []byte{1, 2, 3},
							},
						},
					},
					Activities: map[int64]*persistence.ActivityInfo{
						11: {
							Version:    1,
							ScheduleID: 11,
						},
					},
					History: events.PersistedBlobs{
						events.PersistedBlob{},
					},
					PersistenceError: true,
				})
			},
		},
		{
			name: "nil snapshot",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			mockEngine := engine.NewMockEngine(mockCtrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockEngine)
			}
			notifyTasksFromWorkflowSnapshot(mockEngine, tc.workflowSnapShot, tc.history, tc.persistenceError)
		})
	}
}

func TestNotifyTasksFromWorkflowMutation(t *testing.T) {
	testCases := []struct {
		name             string
		workflowMutation *persistence.WorkflowMutation
		history          events.PersistedBlobs
		persistenceError bool
		mockSetup        func(*engine.MockEngine)
	}{
		{
			name: "Success case",
			workflowMutation: &persistence.WorkflowMutation{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					DomainID:   "test-domain-id",
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				},
				VersionHistories: &persistence.VersionHistories{
					CurrentVersionHistoryIndex: 0,
					Histories: []*persistence.VersionHistory{
						{
							BranchToken: []byte{1, 2, 3},
						},
					},
				},
				UpsertActivityInfos: []*persistence.ActivityInfo{
					{
						Version:    1,
						ScheduleID: 11,
					},
				},
				TransferTasks: []persistence.Task{
					&persistence.ActivityTask{
						TaskList: "test-tl",
					},
				},
				TimerTasks: []persistence.Task{
					&persistence.ActivityTimeoutTask{
						Attempt: 10,
					},
				},
				CrossClusterTasks: []persistence.Task{
					&persistence.CrossClusterStartChildExecutionTask{
						StartChildExecutionTask: persistence.StartChildExecutionTask{
							TargetDomainID: "target-domain",
						},
						TargetCluster: "target",
					},
				},
				ReplicationTasks: []persistence.Task{
					&persistence.HistoryReplicationTask{
						FirstEventID: 1,
						NextEventID:  10,
					},
				},
			},
			history: events.PersistedBlobs{
				events.PersistedBlob{},
			},
			persistenceError: true,
			mockSetup: func(mockEngine *engine.MockEngine) {
				mockEngine.EXPECT().NotifyNewTransferTasks(&hcommon.NotifyTaskInfo{
					ExecutionInfo: &persistence.WorkflowExecutionInfo{
						DomainID:   "test-domain-id",
						WorkflowID: "test-workflow-id",
						RunID:      "test-run-id",
					},
					Tasks: []persistence.Task{
						&persistence.ActivityTask{
							TaskList: "test-tl",
						},
					},
					PersistenceError: true,
				})
				mockEngine.EXPECT().NotifyNewTimerTasks(&hcommon.NotifyTaskInfo{
					ExecutionInfo: &persistence.WorkflowExecutionInfo{
						DomainID:   "test-domain-id",
						WorkflowID: "test-workflow-id",
						RunID:      "test-run-id",
					},
					Tasks: []persistence.Task{
						&persistence.ActivityTimeoutTask{
							Attempt: 10,
						},
					},
					PersistenceError: true,
				})
				mockEngine.EXPECT().NotifyNewCrossClusterTasks(&hcommon.NotifyTaskInfo{
					ExecutionInfo: &persistence.WorkflowExecutionInfo{
						DomainID:   "test-domain-id",
						WorkflowID: "test-workflow-id",
						RunID:      "test-run-id",
					},
					Tasks: []persistence.Task{
						&persistence.CrossClusterStartChildExecutionTask{
							StartChildExecutionTask: persistence.StartChildExecutionTask{
								TargetDomainID: "target-domain",
							},
							TargetCluster: "target",
						},
					},
					PersistenceError: true,
				})
				mockEngine.EXPECT().NotifyNewReplicationTasks(&hcommon.NotifyTaskInfo{
					ExecutionInfo: &persistence.WorkflowExecutionInfo{
						DomainID:   "test-domain-id",
						WorkflowID: "test-workflow-id",
						RunID:      "test-run-id",
					},
					Tasks: []persistence.Task{
						&persistence.HistoryReplicationTask{
							FirstEventID: 1,
							NextEventID:  10,
						},
					},
					VersionHistories: &persistence.VersionHistories{
						CurrentVersionHistoryIndex: 0,
						Histories: []*persistence.VersionHistory{
							{
								BranchToken: []byte{1, 2, 3},
							},
						},
					},
					Activities: map[int64]*persistence.ActivityInfo{
						11: {
							Version:    1,
							ScheduleID: 11,
						},
					},
					History: events.PersistedBlobs{
						events.PersistedBlob{},
					},
					PersistenceError: true,
				})
			},
		},
		{
			name: "nil mutation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			mockEngine := engine.NewMockEngine(mockCtrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockEngine)
			}
			notifyTasksFromWorkflowMutation(mockEngine, tc.workflowMutation, tc.history, tc.persistenceError)
		})
	}
}

func TestActivityInfosToMap(t *testing.T) {
	testCases := []struct {
		name       string
		activities []*persistence.ActivityInfo
		want       map[int64]*persistence.ActivityInfo
	}{
		{
			name: "non-empty",
			activities: []*persistence.ActivityInfo{
				{
					Version:    1,
					ScheduleID: 11,
				},
				{
					Version:    2,
					ScheduleID: 12,
				},
			},
			want: map[int64]*persistence.ActivityInfo{
				11: {
					Version:    1,
					ScheduleID: 11,
				},
				12: {
					Version:    2,
					ScheduleID: 12,
				},
			},
		},
		{
			name:       "empty slice",
			activities: []*persistence.ActivityInfo{},
			want:       map[int64]*persistence.ActivityInfo{},
		},
		{
			name: "nil slice",
			want: map[int64]*persistence.ActivityInfo{},
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.want, activityInfosToMap(tc.activities))
	}
}

func TestCreateWorkflowExecutionWithRetry(t *testing.T) {
	testCases := []struct {
		name      string
		request   *persistence.CreateWorkflowExecutionRequest
		mockSetup func(*shard.MockContext)
		want      *persistence.CreateWorkflowExecutionResponse
		wantErr   bool
		assertErr func(*testing.T, error)
	}{
		{
			name: "Success case",
			request: &persistence.CreateWorkflowExecutionRequest{
				RangeID: 100,
			},
			mockSetup: func(mockShard *shard.MockContext) {
				mockShard.EXPECT().CreateWorkflowExecution(gomock.Any(), &persistence.CreateWorkflowExecutionRequest{
					RangeID: 100,
				}).Return(&persistence.CreateWorkflowExecutionResponse{
					MutableStateUpdateSessionStats: &persistence.MutableStateUpdateSessionStats{
						MutableStateSize: 123,
					},
				}, nil)
			},
			want: &persistence.CreateWorkflowExecutionResponse{
				MutableStateUpdateSessionStats: &persistence.MutableStateUpdateSessionStats{
					MutableStateSize: 123,
				},
			},
			wantErr: false,
		},
		{
			name: "workflow already started error",
			request: &persistence.CreateWorkflowExecutionRequest{
				RangeID: 100,
			},
			mockSetup: func(mockShard *shard.MockContext) {
				mockShard.EXPECT().CreateWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, &persistence.WorkflowExecutionAlreadyStartedError{})
			},
			wantErr: true,
			assertErr: func(t *testing.T, err error) {
				assert.IsType(t, err, &persistence.WorkflowExecutionAlreadyStartedError{})
			},
		},
		{
			name: "timeout error",
			request: &persistence.CreateWorkflowExecutionRequest{
				RangeID: 100,
			},
			mockSetup: func(mockShard *shard.MockContext) {
				mockShard.EXPECT().CreateWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, &persistence.TimeoutError{})
			},
			wantErr: true,
			assertErr: func(t *testing.T, err error) {
				assert.IsType(t, err, &persistence.TimeoutError{})
			},
		},
		{
			name: "retry succeeds",
			request: &persistence.CreateWorkflowExecutionRequest{
				RangeID: 100,
			},
			mockSetup: func(mockShard *shard.MockContext) {
				mockShard.EXPECT().CreateWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, &types.ServiceBusyError{})
				mockShard.EXPECT().CreateWorkflowExecution(gomock.Any(), gomock.Any()).Return(&persistence.CreateWorkflowExecutionResponse{
					MutableStateUpdateSessionStats: &persistence.MutableStateUpdateSessionStats{
						MutableStateSize: 123,
					},
				}, nil)
			},
			want: &persistence.CreateWorkflowExecutionResponse{
				MutableStateUpdateSessionStats: &persistence.MutableStateUpdateSessionStats{
					MutableStateSize: 123,
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			mockShard := shard.NewMockContext(mockCtrl)
			policy := backoff.NewExponentialRetryPolicy(time.Millisecond)
			policy.SetMaximumAttempts(1)
			if tc.mockSetup != nil {
				tc.mockSetup(mockShard)
			}
			resp, err := createWorkflowExecutionWithRetry(context.Background(), mockShard, testlogger.New(t), policy, tc.request)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.assertErr != nil {
					tc.assertErr(t, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, resp)
			}
		})
	}
}

func TestAppendHistoryV2EventsWithRetry(t *testing.T) {
	testCases := []struct {
		name      string
		domainID  string
		execution types.WorkflowExecution
		request   *persistence.AppendHistoryNodesRequest
		mockSetup func(*shard.MockContext)
		want      *persistence.AppendHistoryNodesResponse
		wantErr   bool
	}{
		{
			name:     "Success case",
			domainID: "test-domain-id",
			execution: types.WorkflowExecution{
				WorkflowID: "test-workflow-id",
				RunID:      "test-run-id",
			},
			request: &persistence.AppendHistoryNodesRequest{
				IsNewBranch: true,
			},
			mockSetup: func(mockShard *shard.MockContext) {
				mockShard.EXPECT().AppendHistoryV2Events(gomock.Any(), &persistence.AppendHistoryNodesRequest{
					IsNewBranch: true,
				}, "test-domain-id", types.WorkflowExecution{WorkflowID: "test-workflow-id", RunID: "test-run-id"}).Return(&persistence.AppendHistoryNodesResponse{
					DataBlob: persistence.DataBlob{},
				}, nil)
			},
			want: &persistence.AppendHistoryNodesResponse{
				DataBlob: persistence.DataBlob{},
			},
			wantErr: false,
		},
		{
			name:     "retry success",
			domainID: "test-domain-id",
			execution: types.WorkflowExecution{
				WorkflowID: "test-workflow-id",
				RunID:      "test-run-id",
			},
			request: &persistence.AppendHistoryNodesRequest{
				IsNewBranch: true,
			},
			mockSetup: func(mockShard *shard.MockContext) {
				mockShard.EXPECT().AppendHistoryV2Events(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, &types.ServiceBusyError{})
				mockShard.EXPECT().AppendHistoryV2Events(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&persistence.AppendHistoryNodesResponse{
					DataBlob: persistence.DataBlob{},
				}, nil)
			},
			want: &persistence.AppendHistoryNodesResponse{
				DataBlob: persistence.DataBlob{},
			},
			wantErr: false,
		},
		{
			name:     "non retryable error",
			domainID: "test-domain-id",
			execution: types.WorkflowExecution{
				WorkflowID: "test-workflow-id",
				RunID:      "test-run-id",
			},
			request: &persistence.AppendHistoryNodesRequest{
				IsNewBranch: true,
			},
			mockSetup: func(mockShard *shard.MockContext) {
				mockShard.EXPECT().AppendHistoryV2Events(gomock.Any(), &persistence.AppendHistoryNodesRequest{
					IsNewBranch: true,
				}, "test-domain-id", types.WorkflowExecution{WorkflowID: "test-workflow-id", RunID: "test-run-id"}).Return(nil, errors.New("some error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			mockShard := shard.NewMockContext(mockCtrl)
			policy := backoff.NewExponentialRetryPolicy(time.Millisecond)
			policy.SetMaximumAttempts(1)
			if tc.mockSetup != nil {
				tc.mockSetup(mockShard)
			}
			resp, err := appendHistoryV2EventsWithRetry(context.Background(), mockShard, policy, tc.domainID, tc.execution, tc.request)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, resp)
			}
		})
	}
}

func TestPersistStartWorkflowBatchEvents(t *testing.T) {
	testCases := []struct {
		name                     string
		workflowEvents           *persistence.WorkflowEvents
		mockSetup                func(*shard.MockContext, *cache.MockDomainCache)
		mockAppendHistoryNodesFn func(context.Context, string, types.WorkflowExecution, *persistence.AppendHistoryNodesRequest) (*persistence.AppendHistoryNodesResponse, error)
		wantErr                  bool
		want                     events.PersistedBlob
		assertErr                func(*testing.T, error)
	}{
		{
			name:           "empty events",
			workflowEvents: &persistence.WorkflowEvents{},
			wantErr:        true,
			assertErr: func(t *testing.T, err error) {
				assert.IsType(t, err, &types.InternalServiceError{})
				assert.Contains(t, err.Error(), "cannot persist first workflow events with empty events")
			},
		},
		{
			name: "failed to get domain name",
			workflowEvents: &persistence.WorkflowEvents{
				Events: []*types.HistoryEvent{
					{
						ID: 1,
					},
				},
			},
			mockSetup: func(mockShard *shard.MockContext, mockDomainCache *cache.MockDomainCache) {
				mockShard.EXPECT().GetDomainCache().Return(mockDomainCache)
				mockDomainCache.EXPECT().GetDomainName(gomock.Any()).Return("", errors.New("some error"))
			},
			wantErr: true,
			assertErr: func(t *testing.T, err error) {
				assert.Equal(t, err, errors.New("some error"))
			},
		},
		{
			name: "failed to append history nodes",
			workflowEvents: &persistence.WorkflowEvents{
				Events: []*types.HistoryEvent{
					{
						ID: 1,
					},
				},
			},
			mockSetup: func(mockShard *shard.MockContext, mockDomainCache *cache.MockDomainCache) {
				mockShard.EXPECT().GetDomainCache().Return(mockDomainCache)
				mockDomainCache.EXPECT().GetDomainName(gomock.Any()).Return("test-domain", nil)
			},
			mockAppendHistoryNodesFn: func(context.Context, string, types.WorkflowExecution, *persistence.AppendHistoryNodesRequest) (*persistence.AppendHistoryNodesResponse, error) {
				return nil, errors.New("some error")
			},
			wantErr: true,
			assertErr: func(t *testing.T, err error) {
				assert.Equal(t, err, errors.New("some error"))
			},
		},
		{
			name: "success",
			workflowEvents: &persistence.WorkflowEvents{
				Events: []*types.HistoryEvent{
					{
						ID: 1,
					},
				},
				BranchToken: []byte{1, 2, 3},
			},
			mockSetup: func(mockShard *shard.MockContext, mockDomainCache *cache.MockDomainCache) {
				mockShard.EXPECT().GetDomainCache().Return(mockDomainCache)
				mockDomainCache.EXPECT().GetDomainName(gomock.Any()).Return("test-domain", nil)
			},
			mockAppendHistoryNodesFn: func(ctx context.Context, domainID string, execution types.WorkflowExecution, req *persistence.AppendHistoryNodesRequest) (*persistence.AppendHistoryNodesResponse, error) {
				assert.Equal(t, &persistence.AppendHistoryNodesRequest{
					IsNewBranch: true,
					Info:        "::",
					BranchToken: []byte{1, 2, 3},
					Events: []*types.HistoryEvent{
						{
							ID: 1,
						},
					},
					DomainName: "test-domain",
				}, req)
				return &persistence.AppendHistoryNodesResponse{
					DataBlob: persistence.DataBlob{
						Data: []byte("123"),
					},
				}, nil
			},
			want: events.PersistedBlob{
				DataBlob: persistence.DataBlob{
					Data: []byte("123"),
				},
				BranchToken:  []byte{1, 2, 3},
				FirstEventID: 1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			mockShard := shard.NewMockContext(mockCtrl)
			mockDomainCache := cache.NewMockDomainCache(mockCtrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockShard, mockDomainCache)
			}
			ctx := &contextImpl{
				shard: mockShard,
			}
			if tc.mockAppendHistoryNodesFn != nil {
				ctx.appendHistoryNodesFn = tc.mockAppendHistoryNodesFn
			}
			got, err := ctx.PersistStartWorkflowBatchEvents(context.Background(), tc.workflowEvents)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestPersistNonStartWorkflowBatchEvents(t *testing.T) {
	testCases := []struct {
		name                     string
		workflowEvents           *persistence.WorkflowEvents
		mockSetup                func(*shard.MockContext, *cache.MockDomainCache)
		mockAppendHistoryNodesFn func(context.Context, string, types.WorkflowExecution, *persistence.AppendHistoryNodesRequest) (*persistence.AppendHistoryNodesResponse, error)
		wantErr                  bool
		want                     events.PersistedBlob
		assertErr                func(*testing.T, error)
	}{
		{
			name:           "empty events",
			workflowEvents: &persistence.WorkflowEvents{},
			wantErr:        false,
			want:           events.PersistedBlob{},
		},
		{
			name: "failed to get domain name",
			workflowEvents: &persistence.WorkflowEvents{
				Events: []*types.HistoryEvent{
					{
						ID: 1,
					},
				},
			},
			mockSetup: func(mockShard *shard.MockContext, mockDomainCache *cache.MockDomainCache) {
				mockShard.EXPECT().GetDomainCache().Return(mockDomainCache)
				mockDomainCache.EXPECT().GetDomainName(gomock.Any()).Return("", errors.New("some error"))
			},
			wantErr: true,
			assertErr: func(t *testing.T, err error) {
				assert.Equal(t, err, errors.New("some error"))
			},
		},
		{
			name: "failed to append history nodes",
			workflowEvents: &persistence.WorkflowEvents{
				Events: []*types.HistoryEvent{
					{
						ID: 1,
					},
				},
			},
			mockSetup: func(mockShard *shard.MockContext, mockDomainCache *cache.MockDomainCache) {
				mockShard.EXPECT().GetDomainCache().Return(mockDomainCache)
				mockDomainCache.EXPECT().GetDomainName(gomock.Any()).Return("test-domain", nil)
			},
			mockAppendHistoryNodesFn: func(context.Context, string, types.WorkflowExecution, *persistence.AppendHistoryNodesRequest) (*persistence.AppendHistoryNodesResponse, error) {
				return nil, errors.New("some error")
			},
			wantErr: true,
			assertErr: func(t *testing.T, err error) {
				assert.Equal(t, err, errors.New("some error"))
			},
		},
		{
			name: "success",
			workflowEvents: &persistence.WorkflowEvents{
				Events: []*types.HistoryEvent{
					{
						ID: 1,
					},
				},
				BranchToken: []byte{1, 2, 3},
			},
			mockSetup: func(mockShard *shard.MockContext, mockDomainCache *cache.MockDomainCache) {
				mockShard.EXPECT().GetDomainCache().Return(mockDomainCache)
				mockDomainCache.EXPECT().GetDomainName(gomock.Any()).Return("test-domain", nil)
			},
			mockAppendHistoryNodesFn: func(ctx context.Context, domainID string, execution types.WorkflowExecution, req *persistence.AppendHistoryNodesRequest) (*persistence.AppendHistoryNodesResponse, error) {
				assert.Equal(t, &persistence.AppendHistoryNodesRequest{
					IsNewBranch: false,
					BranchToken: []byte{1, 2, 3},
					Events: []*types.HistoryEvent{
						{
							ID: 1,
						},
					},
					DomainName: "test-domain",
				}, req)
				return &persistence.AppendHistoryNodesResponse{
					DataBlob: persistence.DataBlob{
						Data: []byte("123"),
					},
				}, nil
			},
			want: events.PersistedBlob{
				DataBlob: persistence.DataBlob{
					Data: []byte("123"),
				},
				BranchToken:  []byte{1, 2, 3},
				FirstEventID: 1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			mockShard := shard.NewMockContext(mockCtrl)
			mockDomainCache := cache.NewMockDomainCache(mockCtrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mockShard, mockDomainCache)
			}
			ctx := &contextImpl{
				shard: mockShard,
			}
			if tc.mockAppendHistoryNodesFn != nil {
				ctx.appendHistoryNodesFn = tc.mockAppendHistoryNodesFn
			}
			got, err := ctx.PersistNonStartWorkflowBatchEvents(context.Background(), tc.workflowEvents)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestCreateWorkflowExecution(t *testing.T) {
	testCases := []struct {
		name                                  string
		newWorkflow                           *persistence.WorkflowSnapshot
		history                               events.PersistedBlob
		createMode                            persistence.CreateWorkflowMode
		prevRunID                             string
		prevLastWriteVersion                  int64
		mockCreateWorkflowExecutionFn         func(context.Context, *persistence.CreateWorkflowExecutionRequest) (*persistence.CreateWorkflowExecutionResponse, error)
		mockNotifyTasksFromWorkflowSnapshotFn func(*persistence.WorkflowSnapshot, events.PersistedBlobs, bool)
		mockEmitSessionUpdateStatsFn          func(string, *persistence.MutableStateUpdateSessionStats)
		wantErr                               bool
	}{
		{
			name: "failed to create workflow execution with possibly success error",
			newWorkflow: &persistence.WorkflowSnapshot{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					DomainID:   "test-domain-id",
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				},
			},
			history: events.PersistedBlob{
				DataBlob: persistence.DataBlob{
					Data: []byte("123"),
				},
				BranchToken:  []byte{1, 2, 3},
				FirstEventID: 1,
			},
			createMode:           persistence.CreateWorkflowModeContinueAsNew,
			prevRunID:            "test-prev-run-id",
			prevLastWriteVersion: 123,
			mockCreateWorkflowExecutionFn: func(context.Context, *persistence.CreateWorkflowExecutionRequest) (*persistence.CreateWorkflowExecutionResponse, error) {
				return nil, &types.InternalServiceError{}
			},
			mockNotifyTasksFromWorkflowSnapshotFn: func(_ *persistence.WorkflowSnapshot, _ events.PersistedBlobs, persistenceError bool) {
				assert.Equal(t, true, persistenceError)
			},
			wantErr: true,
		},
		{
			name: "success",
			newWorkflow: &persistence.WorkflowSnapshot{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					DomainID:   "test-domain-id",
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				},
			},
			history: events.PersistedBlob{
				DataBlob: persistence.DataBlob{
					Data: []byte("123"),
				},
				BranchToken:  []byte{1, 2, 3},
				FirstEventID: 1,
			},
			createMode:           persistence.CreateWorkflowModeContinueAsNew,
			prevRunID:            "test-prev-run-id",
			prevLastWriteVersion: 123,
			mockCreateWorkflowExecutionFn: func(ctx context.Context, req *persistence.CreateWorkflowExecutionRequest) (*persistence.CreateWorkflowExecutionResponse, error) {
				assert.Equal(t, &persistence.CreateWorkflowExecutionRequest{
					Mode:                     persistence.CreateWorkflowModeContinueAsNew,
					PreviousRunID:            "test-prev-run-id",
					PreviousLastWriteVersion: 123,
					NewWorkflowSnapshot: persistence.WorkflowSnapshot{
						ExecutionInfo: &persistence.WorkflowExecutionInfo{
							DomainID:   "test-domain-id",
							WorkflowID: "test-workflow-id",
							RunID:      "test-run-id",
						},
						ExecutionStats: &persistence.ExecutionStats{
							HistorySize: 3,
						},
					},
					DomainName: "test-domain",
				}, req)
				return &persistence.CreateWorkflowExecutionResponse{
					MutableStateUpdateSessionStats: &persistence.MutableStateUpdateSessionStats{
						MutableStateSize: 123,
					},
				}, nil
			},
			mockNotifyTasksFromWorkflowSnapshotFn: func(newWorkflow *persistence.WorkflowSnapshot, history events.PersistedBlobs, persistenceError bool) {
				assert.Equal(t, &persistence.WorkflowSnapshot{
					ExecutionInfo: &persistence.WorkflowExecutionInfo{
						DomainID:   "test-domain-id",
						WorkflowID: "test-workflow-id",
						RunID:      "test-run-id",
					},
				}, newWorkflow)
				assert.Equal(t, events.PersistedBlobs{
					{
						DataBlob: persistence.DataBlob{
							Data: []byte("123"),
						},
						BranchToken:  []byte{1, 2, 3},
						FirstEventID: 1,
					},
				}, history)
				assert.Equal(t, false, persistenceError)
			},
			mockEmitSessionUpdateStatsFn: func(domainName string, stats *persistence.MutableStateUpdateSessionStats) {
				assert.Equal(t, "test-domain", domainName)
				assert.Equal(t, &persistence.MutableStateUpdateSessionStats{
					MutableStateSize: 123,
				}, stats)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			mockShard := shard.NewMockContext(mockCtrl)
			mockDomainCache := cache.NewMockDomainCache(mockCtrl)
			mockShard.EXPECT().GetDomainCache().Return(mockDomainCache)
			mockDomainCache.EXPECT().GetDomainName(gomock.Any()).Return("test-domain", nil)
			ctx := &contextImpl{
				shard:         mockShard,
				stats:         &persistence.ExecutionStats{},
				metricsClient: metrics.NewNoopMetricsClient(),
			}
			if tc.mockCreateWorkflowExecutionFn != nil {
				ctx.createWorkflowExecutionFn = tc.mockCreateWorkflowExecutionFn
			}
			if tc.mockNotifyTasksFromWorkflowSnapshotFn != nil {
				ctx.notifyTasksFromWorkflowSnapshotFn = tc.mockNotifyTasksFromWorkflowSnapshotFn
			}
			if tc.mockEmitSessionUpdateStatsFn != nil {
				ctx.emitSessionUpdateStatsFn = tc.mockEmitSessionUpdateStatsFn
			}
			err := ctx.CreateWorkflowExecution(context.Background(), tc.newWorkflow, tc.history, tc.createMode, tc.prevRunID, tc.prevLastWriteVersion)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
