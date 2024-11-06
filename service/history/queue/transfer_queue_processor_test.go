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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"go.uber.org/goleak"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
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

type (
	transferQueueProcessorSuite struct {
		suite.Suite
		*require.Assertions

		controller           *gomock.Controller
		mockShard            *shard.TestContext
		mockTaskProcessor    *task.MockProcessor
		mockQueueSplitPolicy *MockProcessingQueueSplitPolicy
		mockResetter         *reset.MockWorkflowResetter
		mockArchiver         *archiver.ClientMock
		mockInvariant        *invariant.MockInvariant
		mockWorkflowCache    *workflowcache.MockWFCache

		logger        log.Logger
		metricsClient metrics.Client
		metricsScope  metrics.Scope
	}
)

func TestTransferQueueProcessorSuite(t *testing.T) {
	s := new(transferQueueProcessorSuite)
	suite.Run(t, s)
}

func (s *transferQueueProcessorSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockShard = shard.NewTestContext(
		s.T(),
		s.controller,
		&persistence.ShardInfo{
			ShardID:          10,
			RangeID:          1,
			TransferAckLevel: 0,
		},
		config.NewForTest(),
	)
	s.mockShard.Resource.DomainCache.EXPECT().GetDomainName(gomock.Any()).Return(constants.TestDomainName, nil).AnyTimes()
	s.mockQueueSplitPolicy = NewMockProcessingQueueSplitPolicy(s.controller)
	s.mockTaskProcessor = task.NewMockProcessor(s.controller)

	s.logger = testlogger.New(s.Suite.T())
	s.metricsClient = metrics.NewClient(tally.NoopScope, metrics.History)
	s.metricsScope = s.metricsClient.Scope(metrics.TimerQueueProcessorScope)
	s.mockResetter = reset.NewMockWorkflowResetter(s.controller)
	s.mockArchiver = &archiver.ClientMock{}
	s.mockInvariant = invariant.NewMockInvariant(s.controller)
	s.mockWorkflowCache = workflowcache.NewMockWFCache(s.controller)
}

func (s *transferQueueProcessorSuite) TearDownTest() {
	s.controller.Finish()
	s.mockShard.Finish(s.T())
	// some goroutine leak not from this test
	defer goleak.VerifyNone(s.T(),
		// TODO(CDNC-8881):  TimerGate should not start background goroutine in constructor. Make it start/stoppable
		goleak.IgnoreTopFunction("github.com/uber/cadence/service/history/queue.NewLocalTimerGate.func1"),
	)
}

func (s *transferQueueProcessorSuite) NewProcessor() *transferQueueProcessor {
	return NewTransferQueueProcessor(
		s.mockShard,
		s.mockShard.GetEngine(),
		s.mockTaskProcessor,
		execution.NewCache(s.mockShard),
		s.mockResetter,
		s.mockArchiver,
		s.mockInvariant,
		s.mockWorkflowCache,
		func(domain string) bool { return false },
	).(*transferQueueProcessor)
}

func (s *transferQueueProcessorSuite) NewProcessorWithConfig(cfg *config.Config) *transferQueueProcessor {
	mockShard := shard.NewTestContext(
		s.T(),
		s.controller,
		&persistence.ShardInfo{
			ShardID:          10,
			RangeID:          1,
			TransferAckLevel: 0,
		},
		cfg,
	)

	s.mockShard = mockShard
	return s.NewProcessor()
}

func (s *transferQueueProcessorSuite) TestTransferQueueProcessor_RequireStartStop() {
	processor := s.NewProcessor()

	s.Equal(processor.status, common.DaemonStatusInitialized)
	processor.Start()
	s.Equal(processor.status, common.DaemonStatusStarted)
	// noop start
	processor.Start()

	processor.Stop()
	s.Equal(processor.status, common.DaemonStatusStopped)
	// noop stop
	processor.Stop()
}

func (s *transferQueueProcessorSuite) TestTransferQueueProcessor_RequireStartNotGracefulStop() {
	cfg := config.NewForTest()
	cfg.QueueProcessorEnableGracefulSyncShutdown = dynamicconfig.GetBoolPropertyFn(false)

	processor := s.NewProcessorWithConfig(cfg)

	processor.Start()
	processor.Stop()
}

func (s *transferQueueProcessorSuite) TestNotifyNewTask() {
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
		s.T().Run(name, func(t *testing.T) {
			processor := s.NewProcessor()

			info := &hcommon.NotifyTaskInfo{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					DomainID:   constants.TestDomainID,
					WorkflowID: constants.TestWorkflowID,
					RunID:      constants.TestRunID,
				},
				Tasks: tc.tasks,
			}

			if tc.shouldPanic {
				s.Panics(func() {
					processor.NotifyNewTask(tc.clusterName, info)
				})
			} else {
				processor.NotifyNewTask(tc.clusterName, info)
				tc.checkNotification(processor)
			}
		})
	}
}

func (s *transferQueueProcessorSuite) TestFailoverDomain() {
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
		s.T().Run(name, func(t *testing.T) {
			processor := s.NewProcessor()

			if tc.processorStarted {
				defer processor.Stop()
				processor.Start()
			}

			tc.setupMocks(s.mockShard)

			processor.FailoverDomain(tc.domainIDs)

			processor.ackLevel = 10

			if tc.processorStarted {
				s.Equal(1, len(processor.failoverQueueProcessors))
				s.Equal(common.DaemonStatusStarted, processor.failoverQueueProcessors[0].status)
			}

			if tc.processorStarted {
				processor.drain()
			}
		})
	}
}

func (s *transferQueueProcessorSuite) TestHandleAction() {
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
		s.T().Run(name, func(t *testing.T) {
			processor := s.NewProcessor()
			defer processor.Stop()
			processor.Start()

			ctx := context.Background()

			action := &Action{
				ActionType:               ActionTypeGetState,
				GetStateActionAttributes: &GetStateActionAttributes{},
			}

			actionResult, err := processor.HandleAction(ctx, tc.clusterName, action)

			if tc.err != nil {
				s.Nil(actionResult)
				s.ErrorContains(err, tc.err.Error())
			} else {
				s.NoError(err)
				s.NotNil(actionResult)
				s.Equal(action.ActionType, actionResult.ActionType)
			}
		})
	}
}

func (s *transferQueueProcessorSuite) TestLockTaskProcessing() {
	processor := s.NewProcessor()

	locked := make(chan struct{}, 1)

	processor.LockTaskProcessing()

	go func() {
		defer processor.taskAllocator.Unlock()
		processor.taskAllocator.Lock()
		locked <- struct{}{}
	}()

	select {
	case <-locked:
		s.Fail("Expected mutex to be locked, but it was unlocked")
	case <-time.After(50 * time.Millisecond):
		processor.UnlockTaskProcessing()
		s.True(true, "Mutex is locked as expected")
	}
}

func (s *transferQueueProcessorSuite) Test_completeTransfer() {
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
		s.T().Run(name, func(t *testing.T) {
			processor := s.NewProcessor()
			processor.ackLevel = tt.ackLevel

			tt.mockSetup(s.mockShard)

			defer processor.Stop()
			processor.Start()

			err := processor.completeTransfer()

			if tt.err != nil {
				s.Error(err)
				s.ErrorIs(err, tt.err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *transferQueueProcessorSuite) Test_completeTransferLoop() {
	processor := s.NewProcessor()

	processor.config.TransferProcessorCompleteTransferInterval = dynamicconfig.GetDurationPropertyFn(10 * time.Millisecond)

	processor.activeQueueProcessor.Start()
	for _, standbyQueueProcessor := range processor.standbyQueueProcessors {
		standbyQueueProcessor.Start()
	}

	s.mockShard.Resource.ExecutionMgr.On("RangeCompleteTransferTask", mock.Anything, mock.Anything).
		Return(&persistence.RangeCompleteTransferTaskResponse{}, nil)

	s.mockShard.GetShardManager().(*mocks.ShardManager).On("UpdateShard", mock.Anything, mock.Anything).Return(nil)

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

func (s *transferQueueProcessorSuite) Test_completeTransferLoop_ErrShardClosed() {
	processor := s.NewProcessor()

	processor.config.TransferProcessorCompleteTransferInterval = dynamicconfig.GetDurationPropertyFn(30 * time.Millisecond)

	processor.activeQueueProcessor.Start()
	for _, standbyQueueProcessor := range processor.standbyQueueProcessors {
		standbyQueueProcessor.Start()
	}

	s.mockShard.Resource.ExecutionMgr.On("RangeCompleteTransferTask", mock.Anything, mock.Anything).
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

func (s *transferQueueProcessorSuite) Test_completeTransferLoop_ErrShardClosedNotGraceful() {
	cfg := config.NewForTest()
	cfg.QueueProcessorEnableGracefulSyncShutdown = dynamicconfig.GetBoolPropertyFn(false)

	processor := s.NewProcessorWithConfig(cfg)

	processor.config.TransferProcessorCompleteTransferInterval = dynamicconfig.GetDurationPropertyFn(30 * time.Millisecond)

	processor.activeQueueProcessor.Start()
	for _, standbyQueueProcessor := range processor.standbyQueueProcessors {
		standbyQueueProcessor.Start()
	}

	s.mockShard.Resource.ExecutionMgr.On("RangeCompleteTransferTask", mock.Anything, mock.Anything).
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

func (s *transferQueueProcessorSuite) Test_completeTransferLoop_OtherError() {
	processor := s.NewProcessor()

	processor.config.TransferProcessorCompleteTransferInterval = dynamicconfig.GetDurationPropertyFn(30 * time.Millisecond)

	processor.activeQueueProcessor.Start()
	for _, standbyQueueProcessor := range processor.standbyQueueProcessors {
		standbyQueueProcessor.Start()
	}

	s.mockShard.Resource.ExecutionMgr.On("RangeCompleteTransferTask", mock.Anything, mock.Anything).
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

func (s *transferQueueProcessorSuite) Test_transferQueueActiveProcessor_taskFilter() {
	tests := map[string]struct {
		mockSetup func(*shard.TestContext)
		task      task.Info
		err       error
	}{
		"error - errUnexpectedQueueTask": {
			mockSetup: func(testContext *shard.TestContext) {},
			task:      &persistence.TimerTaskInfo{},
			err:       errUnexpectedQueueTask,
		},
		"noop - domain not registered": {
			mockSetup: func(testContext *shard.TestContext) {
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
		s.T().Run(name, func(t *testing.T) {
			processor := s.NewProcessor()

			tt.mockSetup(s.mockShard)

			err := processor.activeQueueProcessor.taskInitializer(tt.task).Execute()

			if tt.err != nil {
				s.Error(err)
				s.ErrorContains(err, tt.err.Error())
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *transferQueueProcessorSuite) Test_transferQueueActiveProcessor_updateClusterAckLevel() {
	processor := s.NewProcessor()

	taskID := int64(11)

	key := transferTaskKey{
		taskID: taskID,
	}

	s.mockShard.GetShardManager().(*mocks.ShardManager).On("UpdateShard", mock.Anything, mock.Anything).Return(nil).Once()

	err := processor.activeQueueProcessor.processorBase.updateClusterAckLevel(key)

	s.NoError(err)
	s.Equal(taskID, s.mockShard.ShardInfo().ClusterTransferAckLevel[constants.TestClusterMetadata.GetCurrentClusterName()])
}

func (s *transferQueueProcessorSuite) Test_transferQueueActiveProcessor_updateProcessingQueueStates() {
	processor := s.NewProcessor()

	taskID := int64(11)

	key := transferTaskKey{
		taskID: taskID,
	}

	state := NewProcessingQueueState(12, key, key, DomainFilter{})

	states := []ProcessingQueueState{state}

	s.mockShard.GetShardManager().(*mocks.ShardManager).On("UpdateShard", mock.Anything, mock.Anything).Return(nil).Once()

	err := processor.activeQueueProcessor.processorBase.updateProcessingQueueStates(states)

	s.NoError(err)
	s.Equal(taskID, s.mockShard.ShardInfo().ClusterTransferAckLevel[constants.TestClusterMetadata.GetCurrentClusterName()])
	s.Equal(1, len(s.mockShard.ShardInfo().TransferProcessingQueueStates.StatesByCluster[constants.TestClusterMetadata.GetCurrentClusterName()]))
	s.Equal(int32(state.Level()), *s.mockShard.ShardInfo().TransferProcessingQueueStates.StatesByCluster[constants.TestClusterMetadata.GetCurrentClusterName()][0].Level)
}

func (s *transferQueueProcessorSuite) Test_transferQueueActiveProcessor_queueShutdown() {
	processor := s.NewProcessor()

	err := processor.activeQueueProcessor.queueShutdown()

	s.NoError(err)
}

func (s *transferQueueProcessorSuite) Test_transferQueueStandbyProcessor_taskFilter() {
	tests := map[string]struct {
		mockSetup func(*shard.TestContext)
		task      task.Info
		err       error
	}{
		"error - errUnexpectedQueueTask": {
			mockSetup: func(testContext *shard.TestContext) {},
			task:      &persistence.TimerTaskInfo{},
			err:       errUnexpectedQueueTask,
		},
		"noop - domain not registered": {
			mockSetup: func(testContext *shard.TestContext) {
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
		s.T().Run(name, func(t *testing.T) {
			processor := s.NewProcessor()

			tt.mockSetup(s.mockShard)

			err := processor.standbyQueueProcessors["standby"].taskInitializer(tt.task).Execute()

			if tt.err != nil {
				s.Error(err)
				s.ErrorContains(err, tt.err.Error())
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *transferQueueProcessorSuite) Test_transferQueueStandbyProcessor_updateClusterAckLevel() {
	processor := s.NewProcessor()

	taskID := int64(11)

	key := transferTaskKey{
		taskID: taskID,
	}

	s.mockShard.GetShardManager().(*mocks.ShardManager).On("UpdateShard", mock.Anything, mock.Anything).Return(nil).Once()

	err := processor.standbyQueueProcessors["standby"].processorBase.updateClusterAckLevel(key)

	s.NoError(err)
	s.Equal(taskID, s.mockShard.ShardInfo().ClusterTransferAckLevel["standby"])
}

func (s *transferQueueProcessorSuite) Test_transferQueueStandbyProcessor_updateProcessingQueueStates() {
	processor := s.NewProcessor()

	taskID := int64(11)

	key := transferTaskKey{
		taskID: taskID,
	}

	state := NewProcessingQueueState(12, key, key, DomainFilter{})

	states := []ProcessingQueueState{state}

	s.mockShard.GetShardManager().(*mocks.ShardManager).On("UpdateShard", mock.Anything, mock.Anything).Return(nil).Once()

	err := processor.standbyQueueProcessors["standby"].processorBase.updateProcessingQueueStates(states)

	s.NoError(err)
	s.Equal(taskID, s.mockShard.ShardInfo().ClusterTransferAckLevel["standby"])
	s.Equal(1, len(s.mockShard.ShardInfo().TransferProcessingQueueStates.StatesByCluster["standby"]))
	s.Equal(int32(state.Level()), *s.mockShard.ShardInfo().TransferProcessingQueueStates.StatesByCluster["standby"][0].Level)
}

func (s *transferQueueProcessorSuite) Test_transferQueueStandbyProcessor_queueShutdown() {
	processor := s.NewProcessor()

	err := processor.standbyQueueProcessors["standby"].queueShutdown()

	s.NoError(err)
}

func (s *transferQueueProcessorSuite) Test_transferQueueFailoverProcessor_taskFilter() {
	tests := map[string]struct {
		mockSetup func(*shard.TestContext)
		task      task.Info
		err       error
	}{
		"error - errUnexpectedQueueTask": {
			mockSetup: func(testContext *shard.TestContext) {},
			task:      &persistence.TimerTaskInfo{},
			err:       errUnexpectedQueueTask,
		},
		"noop - domain not registered": {
			mockSetup: func(testContext *shard.TestContext) {
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
		s.T().Run(name, func(t *testing.T) {
			processor := s.NewProcessor()

			tt.mockSetup(s.mockShard)

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
				s.Error(err)
				s.ErrorContains(err, tt.err.Error())
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *transferQueueProcessorSuite) Test_transferQueueFailoverProcessor_updateClusterAckLevel() {
	processor := s.NewProcessor()

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

	s.NoError(err)
}

func (s *transferQueueProcessorSuite) Test_transferQueueFailoverProcessor_queueShutdown() {
	processor := s.NewProcessor()

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

	s.NoError(err)
}

func (s *transferQueueProcessorSuite) Test_loadTransferProcessingQueueStates() {
	tests := map[string]struct {
		enableLoadQueueStates bool
		clusterName           string
		taskID                int64
	}{
		"load queue states true": {
			enableLoadQueueStates: true,
			clusterName:           constants.TestClusterMetadata.GetCurrentClusterName(),
			taskID:                s.mockShard.GetTransferClusterAckLevel(constants.TestClusterMetadata.GetCurrentClusterName()),
		},
		"load queue states false": {
			enableLoadQueueStates: false,
			clusterName:           "standby",
			taskID:                s.mockShard.GetTransferClusterAckLevel("standby"),
		},
	}

	for name, tt := range tests {
		s.T().Run(name, func(t *testing.T) {
			opts := &queueProcessorOptions{
				EnableLoadQueueStates: func(opts ...dynamicconfig.FilterOption) bool {
					return tt.enableLoadQueueStates
				},
			}

			pqs := loadTransferProcessingQueueStates(tt.clusterName, s.mockShard, opts, s.logger)

			s.NotNil(pqs)
			s.Equal(1, len(pqs))
			s.Equal(tt.taskID, pqs[0].AckLevel().(transferTaskKey).taskID)
		})
	}
}
