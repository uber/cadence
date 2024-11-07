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

package queue

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/goleak"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/invariant"
	"github.com/uber/cadence/common/types"
	hcommon "github.com/uber/cadence/service/history/common"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/reset"
	"github.com/uber/cadence/service/history/shard"
	"github.com/uber/cadence/service/history/task"
	"github.com/uber/cadence/service/history/workflowcache"
	"github.com/uber/cadence/service/worker/archiver"
)

func TestMain(m *testing.M) {
	defer goleak.VerifyTestMain(m)

	os.Exit(m.Run())
}

func setupMockEnvironment(t *testing.T, cfg *config.Config) (*gomock.Controller, *shard.TestContext) {
	ctrl := gomock.NewController(t)

	if cfg == nil {
		cfg = config.NewForTest()
	}

	mockShard := shard.NewTestContext(
		t,
		ctrl,
		&persistence.ShardInfo{
			ShardID:          10,
			RangeID:          1,
			TransferAckLevel: 0,
		},
		cfg,
	)
	return ctrl, mockShard
}

func setupProcessor(ctrl *gomock.Controller, mockShard *shard.TestContext) *transferQueueProcessor {
	return NewTransferQueueProcessor(
		mockShard,
		mockShard.GetEngine(),
		task.NewMockProcessor(ctrl),
		execution.NewCache(mockShard),
		reset.NewMockWorkflowResetter(ctrl),
		&archiver.ClientMock{},
		invariant.NewMockInvariant(ctrl),
		workflowcache.NewMockWFCache(ctrl),
		func(domain string) bool { return false },
	).(*transferQueueProcessor)
}

func TestTransferQueueProcessorRequireStartStop(t *testing.T) {
	ctrl, mockShard := setupMockEnvironment(t, nil)
	defer ctrl.Finish()

	processor := setupProcessor(ctrl, mockShard)

	assert.Equal(t, common.DaemonStatusInitialized, processor.status)

	processor.Start()
	assert.Equal(t, common.DaemonStatusStarted, processor.status)

	// noop start
	processor.Start()

	processor.Stop()
	assert.Equal(t, common.DaemonStatusStopped, processor.status)

	// noop stop
	processor.Stop()
}

func TestTransferQueueProcessorRequireStartNotGracefulStop(t *testing.T) {
	cfg := config.NewForTest()
	cfg.QueueProcessorEnableGracefulSyncShutdown = dynamicconfig.GetBoolPropertyFn(false)

	ctrl, mockShard := setupMockEnvironment(t, cfg)
	defer ctrl.Finish()

	processor := setupProcessor(ctrl, mockShard)

	assert.Equal(t, common.DaemonStatusInitialized, processor.status)
	processor.Start()

	assert.Equal(t, common.DaemonStatusStarted, processor.status)

	processor.Stop()
	assert.Equal(t, common.DaemonStatusStopped, processor.status)
}

func TestNotifyNewTask(t *testing.T) {
	tests := map[string]struct {
		tasks             []persistence.Task
		clusterName       string
		checkNotification func(processor *transferQueueProcessor)
		shouldPanic       bool
	}{
		"no task": {
			tasks:             []persistence.Task{},
			clusterName:       constants.TestClusterMetadata.GetCurrentClusterName(),
			checkNotification: func(processor *transferQueueProcessor) {},
		},
		"notify active queue processor": {
			tasks: []persistence.Task{
				&persistence.ActivityTask{},
			},
			clusterName: constants.TestClusterMetadata.GetCurrentClusterName(),
			checkNotification: func(processor *transferQueueProcessor) {
				<-processor.activeQueueProcessor.notifyCh
			},
		},
		"notify standby queue processor": {
			tasks: []persistence.Task{
				&persistence.ActivityTask{},
			},
			clusterName: "standby",
			checkNotification: func(processor *transferQueueProcessor) {
				<-processor.standbyQueueProcessors["standby"].notifyCh
			},
		},
		"panic on unknown cluster": {
			tasks: []persistence.Task{
				&persistence.ActivityTask{},
			},
			clusterName:       "unknown",
			checkNotification: func(processor *transferQueueProcessor) {},
			shouldPanic:       true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl, mockShard := setupMockEnvironment(t, nil)
			defer ctrl.Finish()

			info := &hcommon.NotifyTaskInfo{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					DomainID:   constants.TestDomainID,
					WorkflowID: constants.TestWorkflowID,
					RunID:      constants.TestRunID,
				},
				Tasks: tc.tasks,
			}

			processor := setupProcessor(ctrl, mockShard)

			if tc.shouldPanic {
				assert.Panics(t, func() {
					processor.NotifyNewTask(tc.clusterName, info)
				})
			} else {
				processor.NotifyNewTask(tc.clusterName, info)
				tc.checkNotification(processor)
			}
		})
	}
}

func TestFailoverDomain(t *testing.T) {
	tests := map[string]struct {
		domainIDs        map[string]struct{}
		setupMocks       func(mockShard *shard.TestContext)
		processorStarted bool
	}{
		"processor not started": {
			domainIDs:        map[string]struct{}{},
			setupMocks:       func(mockShard *shard.TestContext) {},
			processorStarted: false,
		},
		"processor started": {
			domainIDs: map[string]struct{}{"domainID": {}},
			setupMocks: func(mockShard *shard.TestContext) {
				response := &persistence.GetTransferTasksResponse{
					Tasks: []*persistence.TransferTaskInfo{
						{
							DomainID:   constants.TestDomainID,
							WorkflowID: constants.TestWorkflowID,
							RunID:      constants.TestRunID,
							TaskID:     1,
						},
					},
				}
				mockShard.GetExecutionManager().(*mocks.ExecutionManager).On("GetTransferTasks", context.Background(), mock.Anything).Return(response, nil).Once()
			},
			processorStarted: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl, mockShard := setupMockEnvironment(t, nil)
			defer ctrl.Finish()

			processor := setupProcessor(ctrl, mockShard)

			if tc.processorStarted {
				defer processor.Stop()
				processor.Start()
			}

			tc.setupMocks(mockShard)

			processor.FailoverDomain(tc.domainIDs)

			processor.ackLevel = 10

			if tc.processorStarted {
				assert.Equal(t, 1, len(processor.failoverQueueProcessors))
				assert.Equal(t, common.DaemonStatusStarted, processor.failoverQueueProcessors[0].status)
			}

			if tc.processorStarted {
				processor.drain()
			}
		})
	}
}

func TestHandleAction(t *testing.T) {
	tests := map[string]struct {
		clusterName string
		err         error
	}{
		"active cluster": {
			clusterName: constants.TestClusterMetadata.GetCurrentClusterName(),
		},
		"standby cluster": {
			clusterName: "standby",
		},
		"unknown cluster": {
			clusterName: "unknown",
			err:         errors.New("unknown cluster name: unknown"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl, mockShard := setupMockEnvironment(t, nil)
			defer ctrl.Finish()

			processor := setupProcessor(ctrl, mockShard)
			defer processor.Stop()
			processor.Start()

			ctx := context.Background()

			action := &Action{
				ActionType:               ActionTypeGetState,
				GetStateActionAttributes: &GetStateActionAttributes{},
			}

			actionResult, err := processor.HandleAction(ctx, tc.clusterName, action)

			if tc.err != nil {
				assert.Nil(t, actionResult)
				assert.ErrorContains(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, actionResult)
				assert.Equal(t, action.ActionType, actionResult.ActionType)
			}
		})
	}
}

func TestLockTaskProcessing(t *testing.T) {
	ctrl, mockShard := setupMockEnvironment(t, nil)
	defer ctrl.Finish()

	processor := setupProcessor(ctrl, mockShard)
	locked := make(chan struct{}, 1)

	processor.LockTaskProcessing()

	go func() {
		defer processor.taskAllocator.Unlock()
		processor.taskAllocator.Lock()
		locked <- struct{}{}
	}()

	select {
	case <-locked:
		assert.Fail(t, "Expected mutex to be locked, but it was unlocked")
	case <-time.After(50 * time.Millisecond):
		processor.UnlockTaskProcessing()
		assert.True(t, true, "Mutex is locked as expected")
	}
}

func Test_completeTransfer(t *testing.T) {
	tests := map[string]struct {
		ackLevel  int64
		mockSetup func(*shard.TestContext)
		err       error
	}{
		"noop - ackLevel >= newAckLevelTaskID": {
			ackLevel:  10,
			mockSetup: func(mockShard *shard.TestContext) {},
			err:       nil,
		},
		"error - ackLevel < newAckLevelTaskID - RangeCompleteTransferTask error": {
			ackLevel: 1,
			mockSetup: func(mockShard *shard.TestContext) {
				mockShard.Resource.ExecutionMgr.On("RangeCompleteTransferTask", mock.Anything, mock.Anything).
					Return(&persistence.RangeCompleteTransferTaskResponse{}, assert.AnError).Once()
			},
			err: assert.AnError,
		},
		"success - ackLevel < newAckLevelTaskID": {
			ackLevel: 1,
			mockSetup: func(mockShard *shard.TestContext) {
				mockShard.Resource.ExecutionMgr.On("RangeCompleteTransferTask", mock.Anything, mock.Anything).
					Return(&persistence.RangeCompleteTransferTaskResponse{}, nil).Once()
				mockShard.GetShardManager().(*mocks.ShardManager).On("UpdateShard", mock.Anything, mock.Anything).Return(nil).Once()
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl, mockShard := setupMockEnvironment(t, nil)
			defer ctrl.Finish()

			processor := setupProcessor(ctrl, mockShard)

			processor.ackLevel = tt.ackLevel

			tt.mockSetup(mockShard)

			defer processor.Stop()
			processor.Start()

			err := processor.completeTransfer()

			if tt.err != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_completeTransferLoop(t *testing.T) {
	ctrl, mockShard := setupMockEnvironment(t, nil)
	defer ctrl.Finish()

	processor := setupProcessor(ctrl, mockShard)

	processor.config.TransferProcessorCompleteTransferInterval = dynamicconfig.GetDurationPropertyFn(10 * time.Millisecond)

	processor.activeQueueProcessor.Start()
	for _, standbyQueueProcessor := range processor.standbyQueueProcessors {
		standbyQueueProcessor.Start()
	}

	mockShard.Resource.ExecutionMgr.On("RangeCompleteTransferTask", mock.Anything, mock.Anything).
		Return(&persistence.RangeCompleteTransferTaskResponse{}, nil)

	mockShard.GetShardManager().(*mocks.ShardManager).On("UpdateShard", mock.Anything, mock.Anything).Return(nil)

	processor.shutdownWG.Add(1)

	go func() {
		time.Sleep(200 * time.Millisecond)
		processor.activeQueueProcessor.Stop()
		for _, standbyQueueProcessor := range processor.standbyQueueProcessors {
			standbyQueueProcessor.Stop()
		}

		close(processor.shutdownChan)
		common.AwaitWaitGroup(&processor.shutdownWG, time.Minute)
	}()

	processor.completeTransferLoop()
}

func Test_completeTransferLoop_ErrShardClosed(t *testing.T) {
	ctrl, mockShard := setupMockEnvironment(t, nil)
	defer ctrl.Finish()

	processor := setupProcessor(ctrl, mockShard)

	processor.config.TransferProcessorCompleteTransferInterval = dynamicconfig.GetDurationPropertyFn(30 * time.Millisecond)

	processor.activeQueueProcessor.Start()
	for _, standbyQueueProcessor := range processor.standbyQueueProcessors {
		standbyQueueProcessor.Start()
	}

	mockShard.Resource.ExecutionMgr.On("RangeCompleteTransferTask", mock.Anything, mock.Anything).
		Return(&persistence.RangeCompleteTransferTaskResponse{}, &shard.ErrShardClosed{}).Once()

	processor.shutdownWG.Add(1)

	go func() {
		time.Sleep(50 * time.Millisecond)
		processor.activeQueueProcessor.Stop()
		for _, standbyQueueProcessor := range processor.standbyQueueProcessors {
			standbyQueueProcessor.Stop()
		}

		close(processor.shutdownChan)
		common.AwaitWaitGroup(&processor.shutdownWG, time.Minute)
	}()

	processor.completeTransferLoop()
}

func Test_completeTransferLoop_ErrShardClosedNotGraceful(t *testing.T) {
	cfg := config.NewForTest()
	cfg.QueueProcessorEnableGracefulSyncShutdown = dynamicconfig.GetBoolPropertyFn(false)

	ctrl, mockShard := setupMockEnvironment(t, cfg)
	defer ctrl.Finish()

	processor := setupProcessor(ctrl, mockShard)

	processor.config.TransferProcessorCompleteTransferInterval = dynamicconfig.GetDurationPropertyFn(30 * time.Millisecond)

	processor.activeQueueProcessor.Start()
	for _, standbyQueueProcessor := range processor.standbyQueueProcessors {
		standbyQueueProcessor.Start()
	}

	mockShard.Resource.ExecutionMgr.On("RangeCompleteTransferTask", mock.Anything, mock.Anything).
		Return(&persistence.RangeCompleteTransferTaskResponse{}, &shard.ErrShardClosed{}).Once()

	processor.shutdownWG.Add(1)

	go func() {
		time.Sleep(50 * time.Millisecond)
		processor.activeQueueProcessor.Stop()
		for _, standbyQueueProcessor := range processor.standbyQueueProcessors {
			standbyQueueProcessor.Stop()
		}

		close(processor.shutdownChan)
		common.AwaitWaitGroup(&processor.shutdownWG, time.Minute)
	}()

	processor.completeTransferLoop()
}

func Test_completeTransferLoop_OtherError(t *testing.T) {
	ctrl, mockShard := setupMockEnvironment(t, nil)
	defer ctrl.Finish()

	processor := setupProcessor(ctrl, mockShard)

	processor.config.TransferProcessorCompleteTransferInterval = dynamicconfig.GetDurationPropertyFn(30 * time.Millisecond)

	processor.activeQueueProcessor.Start()
	for _, standbyQueueProcessor := range processor.standbyQueueProcessors {
		standbyQueueProcessor.Start()
	}

	mockShard.Resource.ExecutionMgr.On("RangeCompleteTransferTask", mock.Anything, mock.Anything).
		Return(&persistence.RangeCompleteTransferTaskResponse{}, assert.AnError)

	processor.shutdownWG.Add(1)

	go func() {
		time.Sleep(50 * time.Millisecond)
		processor.activeQueueProcessor.Stop()
		for _, standbyQueueProcessor := range processor.standbyQueueProcessors {
			standbyQueueProcessor.Stop()
		}

		close(processor.shutdownChan)
		common.AwaitWaitGroup(&processor.shutdownWG, time.Minute)
	}()

	processor.completeTransferLoop()
}

func Test_transferQueueActiveProcessor_taskFilter(t *testing.T) {
	tests := map[string]struct {
		mockSetup func(*shard.TestContext)
		task      task.Info
		err       error
	}{
		"error - errUnexpectedQueueTask": {
			mockSetup: func(testContext *shard.TestContext) {
				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainName(constants.TestDomainID).
					Return(constants.TestDomainName, nil).Times(1)
			},
			task: &persistence.TimerTaskInfo{
				DomainID: constants.TestDomainID,
			},
			err: errUnexpectedQueueTask,
		},
		"noop - domain not registered": {
			mockSetup: func(testContext *shard.TestContext) {
				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainName(constants.TestDomainID).
					Return(constants.TestDomainName, nil).Times(1)

				cacheEntry := cache.NewDomainCacheEntryForTest(&persistence.DomainInfo{Status: persistence.DomainStatusDeprecated}, nil, true, nil, 1, nil, 0, 0, 0)

				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainByID(constants.TestDomainID).
					Return(cacheEntry, nil).Times(1)
			},
			task: &persistence.TransferTaskInfo{
				DomainID: constants.TestDomainID,
			},
			err: nil,
		},
		"taskFilter success": {
			mockSetup: func(testContext *shard.TestContext) {
				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainName(constants.TestDomainID).
					Return(constants.TestDomainName, nil).Times(1)

				cacheEntry := cache.NewDomainCacheEntryForTest(
					&persistence.DomainInfo{Status: persistence.DomainStatusRegistered},
					nil,
					true,
					&persistence.DomainReplicationConfig{
						ActiveClusterName: constants.TestClusterMetadata.GetCurrentClusterName(),
					},
					1,
					nil,
					0,
					0,
					0)

				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainByID(constants.TestDomainID).
					Return(cacheEntry, nil).Times(2)
			},
			task: &persistence.TransferTaskInfo{
				DomainID: constants.TestDomainID,
			},
			// Error to Execute only. Since taskImpl is not exported, and the filter is a private field, had to use the Execute method to execute the filter function.
			// The filter returned no error
			err: &types.BadRequestError{Message: "Can't load workflow execution.  WorkflowId not set."},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl, mockShard := setupMockEnvironment(t, nil)
			defer ctrl.Finish()

			processor := setupProcessor(ctrl, mockShard)

			tt.mockSetup(mockShard)

			err := processor.activeQueueProcessor.taskInitializer(tt.task).Execute()

			if tt.err != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_transferQueueActiveProcessor_updateClusterAckLevel(t *testing.T) {
	ctrl, mockShard := setupMockEnvironment(t, nil)
	defer ctrl.Finish()

	processor := setupProcessor(ctrl, mockShard)

	taskID := int64(11)

	key := transferTaskKey{
		taskID: taskID,
	}

	mockShard.GetShardManager().(*mocks.ShardManager).On("UpdateShard", mock.Anything, mock.Anything).Return(nil).Once()

	err := processor.activeQueueProcessor.processorBase.updateClusterAckLevel(key)

	assert.NoError(t, err)
	assert.Equal(t, taskID, mockShard.ShardInfo().ClusterTransferAckLevel[constants.TestClusterMetadata.GetCurrentClusterName()])
}

func Test_transferQueueActiveProcessor_updateProcessingQueueStates(t *testing.T) {
	ctrl, mockShard := setupMockEnvironment(t, nil)
	defer ctrl.Finish()

	processor := setupProcessor(ctrl, mockShard)

	taskID := int64(11)

	key := transferTaskKey{
		taskID: taskID,
	}

	state := NewProcessingQueueState(12, key, key, DomainFilter{})

	states := []ProcessingQueueState{state}

	mockShard.GetShardManager().(*mocks.ShardManager).On("UpdateShard", mock.Anything, mock.Anything).Return(nil).Once()

	err := processor.activeQueueProcessor.processorBase.updateProcessingQueueStates(states)

	assert.NoError(t, err)
	assert.Equal(t, taskID, mockShard.ShardInfo().ClusterTransferAckLevel[constants.TestClusterMetadata.GetCurrentClusterName()])
	assert.Equal(t, 1, len(mockShard.ShardInfo().TransferProcessingQueueStates.StatesByCluster[constants.TestClusterMetadata.GetCurrentClusterName()]))
	assert.Equal(t, int32(state.Level()), *mockShard.ShardInfo().TransferProcessingQueueStates.StatesByCluster[constants.TestClusterMetadata.GetCurrentClusterName()][0].Level)
}

func Test_transferQueueActiveProcessor_queueShutdown(t *testing.T) {
	ctrl, mockShard := setupMockEnvironment(t, nil)
	defer ctrl.Finish()

	processor := setupProcessor(ctrl, mockShard)

	err := processor.activeQueueProcessor.queueShutdown()

	assert.NoError(t, err)
}

func Test_transferQueueStandbyProcessor_taskFilter(t *testing.T) {
	tests := map[string]struct {
		mockSetup func(*shard.TestContext)
		task      task.Info
		err       error
	}{
		"error - errUnexpectedQueueTask": {
			mockSetup: func(testContext *shard.TestContext) {
				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainName(constants.TestDomainID).
					Return(constants.TestDomainName, nil).Times(1)
			},
			task: &persistence.TimerTaskInfo{
				DomainID: constants.TestDomainID,
			},
			err: errUnexpectedQueueTask,
		},
		"noop - domain not registered": {
			mockSetup: func(testContext *shard.TestContext) {
				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainName(constants.TestDomainID).
					Return(constants.TestDomainName, nil).Times(1)

				cacheEntry := cache.NewDomainCacheEntryForTest(&persistence.DomainInfo{Status: persistence.DomainStatusDeprecated}, nil, true, nil, 1, nil, 0, 0, 0)

				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainByID(constants.TestDomainID).
					Return(cacheEntry, nil).Times(1)
			},
			task: &persistence.TransferTaskInfo{
				DomainID: constants.TestDomainID,
			},
			err: nil,
		},
		"no error - TransferTaskTypeCloseExecution or TransferTaskTypeRecordWorkflowClosed": {
			mockSetup: func(testContext *shard.TestContext) {
				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainName(constants.TestDomainID).
					Return(constants.TestDomainName, nil).Times(1)

				cacheEntry := cache.NewDomainCacheEntryForTest(
					&persistence.DomainInfo{Status: persistence.DomainStatusRegistered},
					nil,
					true,
					&persistence.DomainReplicationConfig{
						ActiveClusterName: constants.TestClusterMetadata.GetCurrentClusterName(),
						Clusters: []*persistence.ClusterReplicationConfig{
							{ClusterName: constants.TestClusterMetadata.GetCurrentClusterName()},
							{ClusterName: "standby"},
						},
					},
					1,
					nil,
					0,
					0,
					0)

				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainByID(constants.TestDomainID).
					Return(cacheEntry, nil).Times(2)
			},
			task: &persistence.TransferTaskInfo{
				DomainID: constants.TestDomainID,
				TaskType: persistence.TransferTaskTypeCloseExecution,
			},
			// Error to Execute only. Since taskImpl is not exported, and the filter is a private field, had to use the Execute method to execute the filter function.
			// The filter returned no error
			err: &types.BadRequestError{Message: "Can't load workflow execution.  WorkflowId not set."},
		},
		"error - TransferTaskTypeCloseExecution or TransferTaskTypeRecordWorkflowClosed - cannot find domain - retry": {
			mockSetup: func(testContext *shard.TestContext) {
				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainName(constants.TestDomainID).
					Return(constants.TestDomainName, nil).Times(1)

				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainByID(constants.TestDomainID).
					Return(nil, assert.AnError).Times(2)
			},
			task: &persistence.TransferTaskInfo{
				DomainID: constants.TestDomainID,
				TaskType: persistence.TransferTaskTypeCloseExecution,
			},
			err: assert.AnError,
		},
		"noop - TransferTaskTypeCloseExecution or TransferTaskTypeRecordWorkflowClosed - EntityNotExistsError": {
			mockSetup: func(testContext *shard.TestContext) {
				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainName(constants.TestDomainID).
					Return(constants.TestDomainName, nil).Times(1)

				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainByID(constants.TestDomainID).
					Return(nil, &types.EntityNotExistsError{Message: "domain doesn't exist"}).Times(2)
			},
			task: &persistence.TransferTaskInfo{
				DomainID: constants.TestDomainID,
				TaskType: persistence.TransferTaskTypeCloseExecution,
			},
			err: nil,
		},
		"taskFilter success": {
			mockSetup: func(testContext *shard.TestContext) {
				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainName(constants.TestDomainID).
					Return(constants.TestDomainName, nil).Times(1)

				cacheEntry := cache.NewDomainCacheEntryForTest(
					&persistence.DomainInfo{Status: persistence.DomainStatusRegistered},
					nil,
					true,
					&persistence.DomainReplicationConfig{
						ActiveClusterName: constants.TestClusterMetadata.GetCurrentClusterName(),
					},
					1,
					nil,
					0,
					0,
					0)

				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainByID(constants.TestDomainID).
					Return(cacheEntry, nil).Times(2)
			},
			task: &persistence.TransferTaskInfo{
				DomainID: constants.TestDomainID,
			},
			err: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl, mockShard := setupMockEnvironment(t, nil)
			defer ctrl.Finish()

			processor := setupProcessor(ctrl, mockShard)

			tt.mockSetup(mockShard)

			err := processor.standbyQueueProcessors["standby"].taskInitializer(tt.task).Execute()

			if tt.err != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_transferQueueStandbyProcessor_updateClusterAckLevel(t *testing.T) {
	ctrl, mockShard := setupMockEnvironment(t, nil)
	defer ctrl.Finish()

	processor := setupProcessor(ctrl, mockShard)

	taskID := int64(11)

	key := transferTaskKey{
		taskID: taskID,
	}

	mockShard.GetShardManager().(*mocks.ShardManager).On("UpdateShard", mock.Anything, mock.Anything).Return(nil).Once()

	err := processor.standbyQueueProcessors["standby"].processorBase.updateClusterAckLevel(key)

	assert.NoError(t, err)
	assert.Equal(t, taskID, mockShard.ShardInfo().ClusterTransferAckLevel["standby"])
}

func Test_transferQueueStandbyProcessor_updateProcessingQueueStates(t *testing.T) {
	ctrl, mockShard := setupMockEnvironment(t, nil)
	defer ctrl.Finish()

	processor := setupProcessor(ctrl, mockShard)

	taskID := int64(11)

	key := transferTaskKey{
		taskID: taskID,
	}

	state := NewProcessingQueueState(12, key, key, DomainFilter{})

	states := []ProcessingQueueState{state}

	mockShard.GetShardManager().(*mocks.ShardManager).On("UpdateShard", mock.Anything, mock.Anything).Return(nil).Once()

	err := processor.standbyQueueProcessors["standby"].processorBase.updateProcessingQueueStates(states)

	assert.NoError(t, err)
	assert.Equal(t, taskID, mockShard.ShardInfo().ClusterTransferAckLevel["standby"])
	assert.Equal(t, 1, len(mockShard.ShardInfo().TransferProcessingQueueStates.StatesByCluster["standby"]))
	assert.Equal(t, int32(state.Level()), *mockShard.ShardInfo().TransferProcessingQueueStates.StatesByCluster["standby"][0].Level)
}

func Test_transferQueueStandbyProcessor_queueShutdown(t *testing.T) {
	ctrl, mockShard := setupMockEnvironment(t, nil)
	defer ctrl.Finish()

	processor := setupProcessor(ctrl, mockShard)

	err := processor.standbyQueueProcessors["standby"].queueShutdown()

	assert.NoError(t, err)
}

func Test_transferQueueFailoverProcessor_taskFilter(t *testing.T) {
	tests := map[string]struct {
		mockSetup func(*shard.TestContext)
		task      task.Info
		err       error
	}{
		"error - errUnexpectedQueueTask": {
			mockSetup: func(testContext *shard.TestContext) {
				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainName(constants.TestDomainID).
					Return(constants.TestDomainName, nil).Times(1)
			},
			task: &persistence.TimerTaskInfo{
				DomainID: constants.TestDomainID,
			},
			err: errUnexpectedQueueTask,
		},
		"noop - domain not registered": {
			mockSetup: func(testContext *shard.TestContext) {
				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainName(constants.TestDomainID).
					Return(constants.TestDomainName, nil).Times(1)

				cacheEntry := cache.NewDomainCacheEntryForTest(&persistence.DomainInfo{Status: persistence.DomainStatusDeprecated}, nil, true, nil, 1, nil, 0, 0, 0)

				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainByID(constants.TestDomainID).
					Return(cacheEntry, nil).Times(1)
			},
			task: &persistence.TransferTaskInfo{
				DomainID: constants.TestDomainID,
			},
			err: nil,
		},
		"taskFilter success": {
			mockSetup: func(testContext *shard.TestContext) {
				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainName(constants.TestDomainID).
					Return(constants.TestDomainName, nil).Times(1)

				cacheEntry := cache.NewDomainCacheEntryForTest(
					&persistence.DomainInfo{Status: persistence.DomainStatusRegistered},
					nil,
					true,
					&persistence.DomainReplicationConfig{
						ActiveClusterName: constants.TestClusterMetadata.GetCurrentClusterName(),
					},
					1,
					nil,
					0,
					0,
					0)

				testContext.GetDomainCache().(*cache.MockDomainCache).EXPECT().GetDomainByID(constants.TestDomainID).
					Return(cacheEntry, nil).Times(1)
			},
			task: &persistence.TransferTaskInfo{
				DomainID: constants.TestDomainID,
			},
			err: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl, mockShard := setupMockEnvironment(t, nil)
			defer ctrl.Finish()

			processor := setupProcessor(ctrl, mockShard)

			tt.mockSetup(mockShard)

			domainIDs := map[string]struct{}{"standby": {}}

			_, failoverQueueProcessor := newTransferQueueFailoverProcessor(
				processor.shard,
				processor.taskProcessor,
				processor.taskAllocator,
				processor.activeTaskExecutor,
				processor.logger,
				0,
				10,
				domainIDs,
				constants.TestClusterMetadata.GetCurrentClusterName(),
			)

			err := failoverQueueProcessor.taskInitializer(tt.task).Execute()

			if tt.err != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_transferQueueFailoverProcessor_updateClusterAckLevel(t *testing.T) {
	ctrl, mockShard := setupMockEnvironment(t, nil)
	defer ctrl.Finish()

	processor := setupProcessor(ctrl, mockShard)

	taskID := int64(11)

	key := transferTaskKey{
		taskID: taskID,
	}

	domainIDs := map[string]struct{}{"standby": {}}

	updateClusterAckLevel, _ := newTransferQueueFailoverProcessor(
		processor.shard,
		processor.taskProcessor,
		processor.taskAllocator,
		processor.activeTaskExecutor,
		processor.logger,
		0,
		10,
		domainIDs,
		constants.TestClusterMetadata.GetCurrentClusterName(),
	)

	err := updateClusterAckLevel(key)

	assert.NoError(t, err)
}

func Test_transferQueueFailoverProcessor_queueShutdown(t *testing.T) {
	ctrl, mockShard := setupMockEnvironment(t, nil)
	defer ctrl.Finish()

	processor := setupProcessor(ctrl, mockShard)

	domainIDs := map[string]struct{}{"standby": {}}

	_, failoverQueueProcessor := newTransferQueueFailoverProcessor(
		processor.shard,
		processor.taskProcessor,
		processor.taskAllocator,
		processor.activeTaskExecutor,
		processor.logger,
		0,
		10,
		domainIDs,
		constants.TestClusterMetadata.GetCurrentClusterName(),
	)

	err := failoverQueueProcessor.queueShutdown()

	assert.NoError(t, err)
}

func Test_loadTransferProcessingQueueStates(t *testing.T) {
	tests := map[string]struct {
		enableLoadQueueStates bool
		clusterName           string
		taskID                func(testContext *shard.TestContext) int64
	}{
		"load queue states true": {
			enableLoadQueueStates: true,
			clusterName:           constants.TestClusterMetadata.GetCurrentClusterName(),
			taskID: func(testContext *shard.TestContext) int64 {
				return testContext.GetTransferClusterAckLevel(constants.TestClusterMetadata.GetCurrentClusterName())
			},
		},
		"load queue states false": {
			enableLoadQueueStates: false,
			clusterName:           "standby",
			taskID: func(testContext *shard.TestContext) int64 {
				return testContext.GetTransferClusterAckLevel("standby")
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl, mockShard := setupMockEnvironment(t, nil)
			defer ctrl.Finish()

			opts := &queueProcessorOptions{
				EnableLoadQueueStates: func(opts ...dynamicconfig.FilterOption) bool {
					return tt.enableLoadQueueStates
				},
			}

			pqs := loadTransferProcessingQueueStates(tt.clusterName, mockShard, opts, mockShard.GetLogger())

			assert.NotNil(t, pqs)
			assert.Equal(t, 1, len(pqs))
			assert.Equal(t, tt.taskID(mockShard), pqs[0].AckLevel().(transferTaskKey).taskID)
		})
	}
}
