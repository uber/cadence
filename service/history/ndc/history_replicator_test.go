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

package ndc

import (
	ctx "context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/resource"
	"github.com/uber/cadence/service/history/shard"
)

var (
	testStopwatch = metrics.NoopScope(metrics.ReplicateHistoryEventsScope).StartTimer(metrics.CacheLatency)
)

func createTestHistoryReplicator(t *testing.T) historyReplicatorImpl {
	ctrl := gomock.NewController(t)

	mockShard := shard.NewMockContext(ctrl)
	mockShard.EXPECT().GetConfig().Return(&config.Config{
		NumberOfShards:           0,
		IsAdvancedVisConfigExist: false,
		MaxResponseSize:          0,
		HistoryCacheInitialSize:  dynamicconfig.GetIntPropertyFn(10),
		HistoryCacheMaxSize:      dynamicconfig.GetIntPropertyFn(10),
		HistoryCacheTTL:          dynamicconfig.GetDurationPropertyFn(10),
		HostName:                 "test-host",
	}).Times(1)

	// before going into NewHistoryReplicator
	mockExecutionManager := persistence.NewMockExecutionManager(ctrl)
	mockShard.EXPECT().GetExecutionManager().Return(mockExecutionManager).Times(1)
	mockShard.EXPECT().GetLogger().Return(log.NewNoop()).Times(1)
	mockShard.EXPECT().GetMetricsClient().Return(nil).Times(3)

	testExecutionCache := execution.NewCache(mockShard)
	mockEventsReapplier := NewMockEventsReapplier(ctrl)

	// going into NewHistoryReplicator -> newTransactionManager()
	mockShard.EXPECT().GetClusterMetadata().Return(cluster.Metadata{}).Times(2)
	mockHistoryManager := persistence.NewMockHistoryManager(ctrl)
	mockShard.EXPECT().GetHistoryManager().Return(mockHistoryManager).Times(3)

	mockHistoryResource := resource.NewMockResource(ctrl)
	mockShard.EXPECT().GetService().Return(mockHistoryResource).Times(2)
	mockPayloadSerializer := persistence.NewMockPayloadSerializer(ctrl)
	mockHistoryResource.EXPECT().GetPayloadSerializer().Return(mockPayloadSerializer).Times(1)

	// going into NewHistoryReplicator -> newTransactionManager -> reset.NewWorkflowResetter
	mockDomainCache := cache.NewMockDomainCache(ctrl)
	mockShard.EXPECT().GetDomainCache().Return(mockDomainCache).Times(2)

	// going back to NewHistoryReplicator
	mockHistoryResource.EXPECT().GetClusterMetadata().Return(cluster.Metadata{}).Times(1)

	replicator := NewHistoryReplicator(mockShard, testExecutionCache, mockEventsReapplier, log.NewNoop())
	replicatorImpl := replicator.(*historyReplicatorImpl)
	return *replicatorImpl
}

func TestNewHistoryReplicator(t *testing.T) {
	assert.NotNil(t, createTestHistoryReplicator(t))
}

func TestNewHistoryReplicator_newBranchManager(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockShard := shard.NewMockContext(ctrl)
	mockShard.EXPECT().GetConfig().Return(&config.Config{
		NumberOfShards:           0,
		IsAdvancedVisConfigExist: false,
		MaxResponseSize:          0,
		HistoryCacheInitialSize:  dynamicconfig.GetIntPropertyFn(10),
		HistoryCacheMaxSize:      dynamicconfig.GetIntPropertyFn(10),
		HistoryCacheTTL:          dynamicconfig.GetDurationPropertyFn(10),
		HostName:                 "test-host",
	}).Times(1)

	// before going into NewHistoryReplicator
	mockExecutionManager := persistence.NewMockExecutionManager(ctrl)
	mockShard.EXPECT().GetExecutionManager().Return(mockExecutionManager).Times(1)
	mockShard.EXPECT().GetLogger().Return(log.NewNoop()).Times(1)
	mockShard.EXPECT().GetMetricsClient().Return(nil).Times(3)

	testExecutionCache := execution.NewCache(mockShard)
	mockEventsReapplier := NewMockEventsReapplier(ctrl)

	// going into NewHistoryReplicator -> newTransactionManager()
	mockShard.EXPECT().GetClusterMetadata().Return(cluster.Metadata{}).Times(2)
	mockHistoryManager := persistence.NewMockHistoryManager(ctrl)
	mockShard.EXPECT().GetHistoryManager().Return(mockHistoryManager).Times(4)

	mockHistoryResource := resource.NewMockResource(ctrl)
	mockShard.EXPECT().GetService().Return(mockHistoryResource).Times(3)
	mockPayloadSerializer := persistence.NewMockPayloadSerializer(ctrl)
	mockHistoryResource.EXPECT().GetPayloadSerializer().Return(mockPayloadSerializer).Times(1)

	// going into NewHistoryReplicator -> newTransactionManager -> reset.NewWorkflowResetter
	mockDomainCache := cache.NewMockDomainCache(ctrl)
	mockShard.EXPECT().GetDomainCache().Return(mockDomainCache).Times(2)

	// going back to NewHistoryReplicator
	mockHistoryResource.EXPECT().GetClusterMetadata().Return(cluster.Metadata{}).Times(2)

	testReplicator := NewHistoryReplicator(mockShard, testExecutionCache, mockEventsReapplier, log.NewNoop())
	testReplicatorImpl := testReplicator.(*historyReplicatorImpl)

	// test newBranchManagerFn function in history replicator
	mockExecutionContext := execution.NewMockContext(ctrl)
	mockExecutionMutableState := execution.NewMockMutableState(ctrl)
	assert.NotNil(t, testReplicatorImpl.newBranchManagerFn(mockExecutionContext, mockExecutionMutableState, log.NewNoop()))
}

func TestNewHistoryReplicator_newConflictResolver(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockShard := shard.NewMockContext(ctrl)
	mockShard.EXPECT().GetConfig().Return(&config.Config{
		NumberOfShards:           0,
		IsAdvancedVisConfigExist: false,
		MaxResponseSize:          0,
		HistoryCacheInitialSize:  dynamicconfig.GetIntPropertyFn(10),
		HistoryCacheMaxSize:      dynamicconfig.GetIntPropertyFn(10),
		HistoryCacheTTL:          dynamicconfig.GetDurationPropertyFn(10),
		HostName:                 "test-host",
	}).Times(2)

	// before going into NewHistoryReplicator
	mockExecutionManager := persistence.NewMockExecutionManager(ctrl)
	mockShard.EXPECT().GetExecutionManager().Return(mockExecutionManager).Times(1)
	mockShard.EXPECT().GetLogger().Return(log.NewNoop()).Times(1)
	mockShard.EXPECT().GetMetricsClient().Return(nil).Times(3)

	testExecutionCache := execution.NewCache(mockShard)
	mockEventsReapplier := NewMockEventsReapplier(ctrl)

	// going into NewHistoryReplicator -> newTransactionManager()
	mockShard.EXPECT().GetClusterMetadata().Return(cluster.Metadata{}).Times(3)
	mockHistoryManager := persistence.NewMockHistoryManager(ctrl)
	mockShard.EXPECT().GetHistoryManager().Return(mockHistoryManager).Times(4)

	mockHistoryResource := resource.NewMockResource(ctrl)
	mockShard.EXPECT().GetService().Return(mockHistoryResource).Times(3)
	mockPayloadSerializer := persistence.NewMockPayloadSerializer(ctrl)
	mockHistoryResource.EXPECT().GetPayloadSerializer().Return(mockPayloadSerializer).Times(1)

	// going into NewHistoryReplicator -> newTransactionManager -> reset.NewWorkflowResetter
	mockDomainCache := cache.NewMockDomainCache(ctrl)
	mockShard.EXPECT().GetDomainCache().Return(mockDomainCache).Times(4)

	// going back to NewHistoryReplicator
	mockHistoryResource.EXPECT().GetClusterMetadata().Return(cluster.Metadata{}).Times(2)

	testReplicator := NewHistoryReplicator(mockShard, testExecutionCache, mockEventsReapplier, log.NewNoop())
	testReplicatorImpl := testReplicator.(*historyReplicatorImpl)

	// test newConflictResolverFn function in history replicator
	mockEventsCache := events.NewMockCache(ctrl)
	mockShard.EXPECT().GetEventsCache().Return(mockEventsCache).Times(1)
	mockShard.EXPECT().GetShardID().Return(0).Times(1)

	mockExecutionContext := execution.NewMockContext(ctrl)
	mockExecutionMutableState := execution.NewMockMutableState(ctrl)
	assert.NotNil(t, testReplicatorImpl.newConflictResolverFn(mockExecutionContext, mockExecutionMutableState, log.NewNoop()))
}

func TestNewHistoryReplicator_newWorkflowResetter(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockShard := shard.NewMockContext(ctrl)
	mockShard.EXPECT().GetConfig().Return(&config.Config{
		NumberOfShards:           0,
		IsAdvancedVisConfigExist: false,
		MaxResponseSize:          0,
		HistoryCacheInitialSize:  dynamicconfig.GetIntPropertyFn(10),
		HistoryCacheMaxSize:      dynamicconfig.GetIntPropertyFn(10),
		HistoryCacheTTL:          dynamicconfig.GetDurationPropertyFn(10),
		HostName:                 "test-host",
	}).Times(2)

	// before going into NewHistoryReplicator
	mockExecutionManager := persistence.NewMockExecutionManager(ctrl)
	mockShard.EXPECT().GetExecutionManager().Return(mockExecutionManager).Times(1)
	mockShard.EXPECT().GetLogger().Return(log.NewNoop()).Times(1)
	mockShard.EXPECT().GetMetricsClient().Return(nil).Times(3)

	testExecutionCache := execution.NewCache(mockShard)
	mockEventsReapplier := NewMockEventsReapplier(ctrl)

	// going into NewHistoryReplicator -> newTransactionManager()
	mockShard.EXPECT().GetClusterMetadata().Return(cluster.Metadata{}).Times(3)
	mockHistoryManager := persistence.NewMockHistoryManager(ctrl)
	mockShard.EXPECT().GetHistoryManager().Return(mockHistoryManager).Times(5)

	mockHistoryResource := resource.NewMockResource(ctrl)
	mockShard.EXPECT().GetService().Return(mockHistoryResource).Times(3)
	mockPayloadSerializer := persistence.NewMockPayloadSerializer(ctrl)
	mockHistoryResource.EXPECT().GetPayloadSerializer().Return(mockPayloadSerializer).Times(1)

	// going into NewHistoryReplicator -> newTransactionManager -> reset.NewWorkflowResetter
	mockDomainCache := cache.NewMockDomainCache(ctrl)
	mockShard.EXPECT().GetDomainCache().Return(mockDomainCache).Times(4)

	// going back to NewHistoryReplicator
	mockHistoryResource.EXPECT().GetClusterMetadata().Return(cluster.Metadata{}).Times(2)

	testReplicator := NewHistoryReplicator(mockShard, testExecutionCache, mockEventsReapplier, log.NewNoop())
	testReplicatorImpl := testReplicator.(*historyReplicatorImpl)

	// test newWorkflowResetterFn function in history replicator
	mockEventsCache := events.NewMockCache(ctrl)
	mockShard.EXPECT().GetEventsCache().Return(mockEventsCache).Times(1)
	mockShard.EXPECT().GetShardID().Return(0).Times(1)

	mockExecutionContext := execution.NewMockContext(ctrl)
	assert.NotNil(t, testReplicatorImpl.newWorkflowResetterFn(
		"test-domain-id",
		"test-workflow-id",
		"test-base-run-id",
		mockExecutionContext,
		"test-run-id",
		log.NewNoop(),
	))
}

func TestNewHistoryReplicator_newStateBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockShard := shard.NewMockContext(ctrl)
	mockShard.EXPECT().GetConfig().Return(&config.Config{
		NumberOfShards:           0,
		IsAdvancedVisConfigExist: false,
		MaxResponseSize:          0,
		HistoryCacheInitialSize:  dynamicconfig.GetIntPropertyFn(10),
		HistoryCacheMaxSize:      dynamicconfig.GetIntPropertyFn(10),
		HistoryCacheTTL:          dynamicconfig.GetDurationPropertyFn(10),
		HostName:                 "test-host",
	}).Times(1)

	// before going into NewHistoryReplicator
	mockExecutionManager := persistence.NewMockExecutionManager(ctrl)
	mockShard.EXPECT().GetExecutionManager().Return(mockExecutionManager).Times(1)
	mockShard.EXPECT().GetLogger().Return(log.NewNoop()).Times(1)
	mockShard.EXPECT().GetMetricsClient().Return(nil).Times(3)

	testExecutionCache := execution.NewCache(mockShard)
	mockEventsReapplier := NewMockEventsReapplier(ctrl)

	// going into NewHistoryReplicator -> newTransactionManager()
	mockShard.EXPECT().GetClusterMetadata().Return(cluster.Metadata{}).Times(2)
	mockHistoryManager := persistence.NewMockHistoryManager(ctrl)
	mockShard.EXPECT().GetHistoryManager().Return(mockHistoryManager).Times(3)

	mockHistoryResource := resource.NewMockResource(ctrl)
	mockShard.EXPECT().GetService().Return(mockHistoryResource).Times(3)
	mockPayloadSerializer := persistence.NewMockPayloadSerializer(ctrl)
	mockHistoryResource.EXPECT().GetPayloadSerializer().Return(mockPayloadSerializer).Times(1)

	// going into NewHistoryReplicator -> newTransactionManager -> reset.NewWorkflowResetter
	mockDomainCache := cache.NewMockDomainCache(ctrl)
	mockShard.EXPECT().GetDomainCache().Return(mockDomainCache).Times(3)

	// going back to NewHistoryReplicator
	mockHistoryResource.EXPECT().GetClusterMetadata().Return(cluster.Metadata{}).Times(2)

	testReplicator := NewHistoryReplicator(mockShard, testExecutionCache, mockEventsReapplier, log.NewNoop())
	testReplicatorImpl := testReplicator.(*historyReplicatorImpl)

	// test newStateBuilderFn function in history replicator
	mockExecutionMutableState := execution.NewMockMutableState(ctrl)
	assert.NotNil(t, testReplicatorImpl.newStateBuilderFn(mockExecutionMutableState, log.NewNoop()))
}

func TestNewHistoryReplicator_newMutableState(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockShard := shard.NewMockContext(ctrl)
	mockShard.EXPECT().GetConfig().Return(&config.Config{
		NumberOfShards:           0,
		IsAdvancedVisConfigExist: false,
		MaxResponseSize:          0,
		HistoryCacheInitialSize:  dynamicconfig.GetIntPropertyFn(10),
		HistoryCacheMaxSize:      dynamicconfig.GetIntPropertyFn(10),
		HistoryCacheTTL:          dynamicconfig.GetDurationPropertyFn(10),
		HostName:                 "test-host",
	}).Times(2)

	// before going into NewHistoryReplicator
	mockExecutionManager := persistence.NewMockExecutionManager(ctrl)
	mockShard.EXPECT().GetExecutionManager().Return(mockExecutionManager).Times(1)
	mockShard.EXPECT().GetLogger().Return(log.NewNoop()).Times(1)
	mockShard.EXPECT().GetMetricsClient().Return(nil).Times(4)

	testExecutionCache := execution.NewCache(mockShard)
	mockEventsReapplier := NewMockEventsReapplier(ctrl)

	// going into NewHistoryReplicator -> newTransactionManager()
	mockShard.EXPECT().GetClusterMetadata().Return(cluster.Metadata{}).Times(4)
	mockHistoryManager := persistence.NewMockHistoryManager(ctrl)
	mockShard.EXPECT().GetHistoryManager().Return(mockHistoryManager).Times(3)

	mockHistoryResource := resource.NewMockResource(ctrl)
	mockShard.EXPECT().GetService().Return(mockHistoryResource).Times(2)
	mockPayloadSerializer := persistence.NewMockPayloadSerializer(ctrl)
	mockHistoryResource.EXPECT().GetPayloadSerializer().Return(mockPayloadSerializer).Times(1)

	// going into NewHistoryReplicator -> newTransactionManager -> reset.NewWorkflowResetter
	mockDomainCache := cache.NewMockDomainCache(ctrl)
	mockShard.EXPECT().GetDomainCache().Return(mockDomainCache).Times(3)

	// going back to NewHistoryReplicator
	mockHistoryResource.EXPECT().GetClusterMetadata().Return(cluster.Metadata{}).Times(1)

	testReplicator := NewHistoryReplicator(mockShard, testExecutionCache, mockEventsReapplier, log.NewNoop())
	testReplicatorImpl := testReplicator.(*historyReplicatorImpl)

	// test newMutableStateFn function in history replicator
	deadline := int64(0)
	mockShard.EXPECT().GetTimeSource().Return(clock.NewMockedTimeSource()).Times(1)
	mockEventsCache := events.NewMockCache(ctrl)
	mockShard.EXPECT().GetEventsCache().Return(mockEventsCache).Times(1)
	mockDomainCacheEntry := cache.NewDomainCacheEntryForTest(
		&persistence.DomainInfo{Name: "test-domain"},
		nil,
		true,
		&persistence.DomainReplicationConfig{ActiveClusterName: "test-active-cluster"},
		0,
		&deadline,
	)
	assert.NotNil(t, testReplicatorImpl.newMutableStateFn(mockDomainCacheEntry, log.NewNoop()))
}

func TestApplyEvents(t *testing.T) {
	replicator := createTestHistoryReplicator(t)
	replicator.newReplicationTaskFn = func(
		clusterMetadata cluster.Metadata,
		historySerializer persistence.PayloadSerializer,
		taskStartTime time.Time,
		logger log.Logger,
		request *types.ReplicateEventsV2Request,
	) (replicationTask, error) {
		return nil, nil
	}

	// Intentionally panic result. Will test applyEvents function seperately
	assert.Panics(t, func() { replicator.ApplyEvents(nil, nil) })
}

func Test_applyEvents_EventTypeWorkflowExecutionStarted(t *testing.T) {
	workflowExecutionStartedType := types.EventTypeWorkflowExecutionStarted

	tests := map[string]struct {
		mockExecutionCacheAffordance  func(mockExecutionCache *execution.MockCache)
		mockReplicationTaskAffordance func(mockReplicationTask *MockreplicationTask)
		applyStartEventsFnAffordance  func() func(
			ctx ctx.Context,
			context execution.Context,
			releaseFn execution.ReleaseFunc,
			task replicationTask,
			domainCache cache.DomainCache,
			newMutableState newMutableStateFn,
			newStateBuilder newStateBuilderFn,
			transactionManager transactionManager,
			logger log.Logger,
			shard shard.Context,
			clusterMetadata cluster.Metadata,
		) error
		expectError error
	}{
		"Case1: success case": {
			mockExecutionCacheAffordance: func(mockExecutionCache *execution.MockCache) {
				mockExecutionCache.EXPECT().GetOrCreateWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, func(err error) {}, nil).Times(1)
			},
			mockReplicationTaskAffordance: func(mockReplicationTask *MockreplicationTask) {
				mockReplicationTask.EXPECT().getDomainID().Return("test-domain-id").Times(1)
				mockReplicationTask.EXPECT().getExecution().Return(&types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				}).Times(1)
				mockReplicationTask.EXPECT().getFirstEvent().Return(&types.HistoryEvent{
					EventType: &workflowExecutionStartedType,
				}).Times(1)
			},
			applyStartEventsFnAffordance: func() func(
				ctx ctx.Context,
				context execution.Context,
				releaseFn execution.ReleaseFunc,
				task replicationTask,
				domainCache cache.DomainCache,
				newMutableState newMutableStateFn,
				newStateBuilder newStateBuilderFn,
				transactionManager transactionManager,
				logger log.Logger,
				shard shard.Context,
				clusterMetadata cluster.Metadata,
			) error {
				fn := func(
					ctx ctx.Context,
					context execution.Context,
					releaseFn execution.ReleaseFunc,
					task replicationTask,
					domainCache cache.DomainCache,
					newMutableState newMutableStateFn,
					newStateBuilder newStateBuilderFn,
					transactionManager transactionManager,
					logger log.Logger,
					shard shard.Context,
					clusterMetadata cluster.Metadata,
				) error {
					return nil
				}
				return fn
			},
			expectError: nil,
		},
		"Case2: error case": {
			mockExecutionCacheAffordance: func(mockExecutionCache *execution.MockCache) {
				mockExecutionCache.EXPECT().GetOrCreateWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, func(err error) {}, fmt.Errorf("test error")).Times(1)
			},
			mockReplicationTaskAffordance: func(mockReplicationTask *MockreplicationTask) {
				mockReplicationTask.EXPECT().getDomainID().Return("test-domain-id").Times(1)
				mockReplicationTask.EXPECT().getExecution().Return(&types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				}).Times(1)
			},
			applyStartEventsFnAffordance: func() func(
				ctx ctx.Context,
				context execution.Context,
				releaseFn execution.ReleaseFunc,
				task replicationTask,
				domainCache cache.DomainCache,
				newMutableState newMutableStateFn,
				newStateBuilder newStateBuilderFn,
				transactionManager transactionManager,
				logger log.Logger,
				shard shard.Context,
				clusterMetadata cluster.Metadata,
			) error {
				fn := func(
					ctx ctx.Context,
					context execution.Context,
					releaseFn execution.ReleaseFunc,
					task replicationTask,
					domainCache cache.DomainCache,
					newMutableState newMutableStateFn,
					newStateBuilder newStateBuilderFn,
					transactionManager transactionManager,
					logger log.Logger,
					shard shard.Context,
					clusterMetadata cluster.Metadata,
				) error {
					return nil
				}
				return fn
			},
			expectError: fmt.Errorf("test error"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			// mock objects
			replicator := createTestHistoryReplicator(t)
			mockReplicationTask := NewMockreplicationTask(ctrl)
			mockExecutionCache := execution.NewMockCache(ctrl)
			replicator.executionCache = mockExecutionCache
			// mock functions
			test.mockReplicationTaskAffordance(mockReplicationTask)
			test.mockExecutionCacheAffordance(mockExecutionCache)

			// parameter functions affordance
			replicator.applyStartEventsFn = test.applyStartEventsFnAffordance()

			assert.Equal(t, replicator.applyEvents(ctx.Background(), mockReplicationTask), test.expectError)
		})
	}
}

func Test_applyEvents_defaultCase_noErrorBranch(t *testing.T) {
	workflowExecutionType := types.EventTypeWorkflowExecutionCompleted

	tests := map[string]struct {
		mockExecutionCacheAffordance                 func(mockExecutionCache *execution.MockCache, mockExecutionContext *execution.MockContext)
		mockReplicationTaskAffordance                func(mockReplicationTask *MockreplicationTask)
		mockExecutionContextAffordance               func(mockExecutionContext *execution.MockContext, mockExecutionMutableState *execution.MockMutableState)
		mockMutableStateAffordance                   func(mockExecutionMutableState *execution.MockMutableState)
		applyNonStartEventsPrepareBranchFnAffordance func() func(ctx ctx.Context,
			context execution.Context,
			mutableState execution.MutableState,
			task replicationTask,
			newBranchManager newBranchManagerFn,
		) (bool, int, error)
		applyNonStartEventsPrepareMutableStateFnAffordance func(mockExecutionMutableState *execution.MockMutableState) func(ctx ctx.Context,
			context execution.Context,
			mutableState execution.MutableState,
			branchIndex int,
			task replicationTask,
			newConflictResolver newConflictResolverFn,
		) (execution.MutableState, bool, error)
		applyNonStartEventsToCurrentBranchFnAffordance func() func(ctx ctx.Context,
			context execution.Context,
			mutableState execution.MutableState,
			isRebuilt bool,
			releaseFn execution.ReleaseFunc,
			task replicationTask,
			newStateBuilder newStateBuilderFn,
			clusterMetadata cluster.Metadata,
			shard shard.Context,
			logger log.Logger,
			transactionManager transactionManager,
		) error
		applyNonStartEventsToNoneCurrentBranchFnAffordance func() func(ctx ctx.Context,
			context execution.Context,
			mutableState execution.MutableState,
			branchIndex int,
			releaseFn execution.ReleaseFunc,
			task replicationTask,
			r *historyReplicatorImpl,
		) error
		expectError error
	}{
		"Case1: case nil with GetVersionHistories is nil": {
			mockExecutionCacheAffordance: func(mockExecutionCache *execution.MockCache, mockExecutionContext *execution.MockContext) {
				mockExecutionCache.EXPECT().GetOrCreateWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockExecutionContext, func(err error) {}, nil).Times(1)
			},
			mockReplicationTaskAffordance: func(mockReplicationTask *MockreplicationTask) {
				mockReplicationTask.EXPECT().getDomainID().Return("test-domain-id").Times(1)
				mockReplicationTask.EXPECT().getExecution().Return(&types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				}).Times(1)
				mockReplicationTask.EXPECT().getFirstEvent().Return(&types.HistoryEvent{
					EventType: &workflowExecutionType,
				}).Times(1)
				mockReplicationTask.EXPECT().getVersion().Return(int64(1)).Times(1)
			},
			mockExecutionContextAffordance: func(mockExecutionContext *execution.MockContext, mockExecutionMutableState *execution.MockMutableState) {
				mockExecutionContext.EXPECT().LoadWorkflowExecutionWithTaskVersion(gomock.Any(), gomock.Any()).
					Return(mockExecutionMutableState, nil).Times(1)
			},
			mockMutableStateAffordance: func(mockExecutionMutableState *execution.MockMutableState) {
				mockExecutionMutableState.EXPECT().GetVersionHistories().Return(nil).Times(1)
			},
			applyNonStartEventsPrepareBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				task replicationTask,
				newBranchManager newBranchManagerFn,
			) (bool, int, error) {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					task replicationTask,
					newBranchManager newBranchManagerFn,
				) (bool, int, error) {
					return false, 0, nil
				}
				return fn
			},
			applyNonStartEventsPrepareMutableStateFnAffordance: func(mockExecutionMutableState *execution.MockMutableState) func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				branchIndex int,
				task replicationTask,
				newConflictResolver newConflictResolverFn,
			) (execution.MutableState, bool, error) {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					branchIndex int,
					task replicationTask,
					newConflictResolver newConflictResolverFn,
				) (execution.MutableState, bool, error) {
					return nil, false, fmt.Errorf("test error")
				}
				return fn
			},
			applyNonStartEventsToCurrentBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				isRebuilt bool,
				releaseFn execution.ReleaseFunc,
				task replicationTask,
				newStateBuilder newStateBuilderFn,
				clusterMetadata cluster.Metadata,
				shard shard.Context,
				logger log.Logger,
				transactionManager transactionManager,
			) error {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					isRebuilt bool,
					releaseFn execution.ReleaseFunc,
					task replicationTask,
					newStateBuilder newStateBuilderFn,
					clusterMetadata cluster.Metadata,
					shard shard.Context,
					logger log.Logger,
					transactionManager transactionManager,
				) error {
					return nil
				}
				return fn
			},
			applyNonStartEventsToNoneCurrentBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				branchIndex int,
				releaseFn execution.ReleaseFunc,
				task replicationTask,
				r *historyReplicatorImpl,
			) error {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					branchIndex int,
					releaseFn execution.ReleaseFunc,
					task replicationTask,
					r *historyReplicatorImpl,
				) error {
					return nil
				}
				return fn
			},
			expectError: execution.ErrMissingVersionHistories,
		},
		"Case2: case nil with applyNonStartEventsPrepareBranchFn error": {
			mockExecutionCacheAffordance: func(mockExecutionCache *execution.MockCache, mockExecutionContext *execution.MockContext) {
				mockExecutionCache.EXPECT().GetOrCreateWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockExecutionContext, func(err error) {}, nil).Times(1)
			},
			mockReplicationTaskAffordance: func(mockReplicationTask *MockreplicationTask) {
				mockReplicationTask.EXPECT().getDomainID().Return("test-domain-id").Times(1)
				mockReplicationTask.EXPECT().getExecution().Return(&types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				}).Times(1)
				mockReplicationTask.EXPECT().getFirstEvent().Return(&types.HistoryEvent{
					EventType: &workflowExecutionType,
				}).Times(1)
				mockReplicationTask.EXPECT().getVersion().Return(int64(1)).Times(1)
			},
			mockExecutionContextAffordance: func(mockExecutionContext *execution.MockContext, mockExecutionMutableState *execution.MockMutableState) {
				mockExecutionContext.EXPECT().LoadWorkflowExecutionWithTaskVersion(gomock.Any(), gomock.Any()).
					Return(mockExecutionMutableState, nil).Times(1)
			},
			mockMutableStateAffordance: func(mockExecutionMutableState *execution.MockMutableState) {
				mockExecutionMutableState.EXPECT().GetVersionHistories().Return(&persistence.VersionHistories{
					CurrentVersionHistoryIndex: 1,
					Histories:                  nil,
				}).Times(1)
			},
			applyNonStartEventsPrepareBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				task replicationTask,
				newBranchManager newBranchManagerFn,
			) (bool, int, error) {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					task replicationTask,
					newBranchManager newBranchManagerFn,
				) (bool, int, error) {
					return false, 0, fmt.Errorf("test error")
				}
				return fn
			},
			applyNonStartEventsPrepareMutableStateFnAffordance: func(mockExecutionMutableState *execution.MockMutableState) func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				branchIndex int,
				task replicationTask,
				newConflictResolver newConflictResolverFn,
			) (execution.MutableState, bool, error) {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					branchIndex int,
					task replicationTask,
					newConflictResolver newConflictResolverFn,
				) (execution.MutableState, bool, error) {
					return nil, false, fmt.Errorf("test error")
				}
				return fn
			},
			applyNonStartEventsToCurrentBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				isRebuilt bool,
				releaseFn execution.ReleaseFunc,
				task replicationTask,
				newStateBuilder newStateBuilderFn,
				clusterMetadata cluster.Metadata,
				shard shard.Context,
				logger log.Logger,
				transactionManager transactionManager,
			) error {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					isRebuilt bool,
					releaseFn execution.ReleaseFunc,
					task replicationTask,
					newStateBuilder newStateBuilderFn,
					clusterMetadata cluster.Metadata,
					shard shard.Context,
					logger log.Logger,
					transactionManager transactionManager,
				) error {
					return nil
				}
				return fn
			},
			applyNonStartEventsToNoneCurrentBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				branchIndex int,
				releaseFn execution.ReleaseFunc,
				task replicationTask,
				r *historyReplicatorImpl,
			) error {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					branchIndex int,
					releaseFn execution.ReleaseFunc,
					task replicationTask,
					r *historyReplicatorImpl,
				) error {
					return nil
				}
				return fn
			},
			expectError: fmt.Errorf("test error"),
		},
		"Case3: case nil with !doContinue": {
			mockExecutionCacheAffordance: func(mockExecutionCache *execution.MockCache, mockExecutionContext *execution.MockContext) {
				mockExecutionCache.EXPECT().GetOrCreateWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockExecutionContext, func(err error) {}, nil).Times(1)
			},
			mockReplicationTaskAffordance: func(mockReplicationTask *MockreplicationTask) {
				mockReplicationTask.EXPECT().getDomainID().Return("test-domain-id").Times(1)
				mockReplicationTask.EXPECT().getExecution().Return(&types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				}).Times(1)
				mockReplicationTask.EXPECT().getFirstEvent().Return(&types.HistoryEvent{
					EventType: &workflowExecutionType,
				}).Times(1)
				mockReplicationTask.EXPECT().getVersion().Return(int64(1)).Times(1)
			},
			mockExecutionContextAffordance: func(mockExecutionContext *execution.MockContext, mockExecutionMutableState *execution.MockMutableState) {
				mockExecutionContext.EXPECT().LoadWorkflowExecutionWithTaskVersion(gomock.Any(), gomock.Any()).
					Return(mockExecutionMutableState, nil).Times(1)
			},
			mockMutableStateAffordance: func(mockExecutionMutableState *execution.MockMutableState) {
				mockExecutionMutableState.EXPECT().GetVersionHistories().Return(&persistence.VersionHistories{
					CurrentVersionHistoryIndex: 1,
					Histories:                  nil,
				}).Times(1)
			},
			applyNonStartEventsPrepareBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				task replicationTask,
				newBranchManager newBranchManagerFn,
			) (bool, int, error) {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					task replicationTask,
					newBranchManager newBranchManagerFn,
				) (bool, int, error) {
					return false, 0, nil
				}
				return fn
			},
			applyNonStartEventsPrepareMutableStateFnAffordance: func(mockExecutionMutableState *execution.MockMutableState) func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				branchIndex int,
				task replicationTask,
				newConflictResolver newConflictResolverFn,
			) (execution.MutableState, bool, error) {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					branchIndex int,
					task replicationTask,
					newConflictResolver newConflictResolverFn,
				) (execution.MutableState, bool, error) {
					return nil, false, fmt.Errorf("test error")
				}
				return fn
			},
			applyNonStartEventsToCurrentBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				isRebuilt bool,
				releaseFn execution.ReleaseFunc,
				task replicationTask,
				newStateBuilder newStateBuilderFn,
				clusterMetadata cluster.Metadata,
				shard shard.Context,
				logger log.Logger,
				transactionManager transactionManager,
			) error {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					isRebuilt bool,
					releaseFn execution.ReleaseFunc,
					task replicationTask,
					newStateBuilder newStateBuilderFn,
					clusterMetadata cluster.Metadata,
					shard shard.Context,
					logger log.Logger,
					transactionManager transactionManager,
				) error {
					return nil
				}
				return fn
			},
			applyNonStartEventsToNoneCurrentBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				branchIndex int,
				releaseFn execution.ReleaseFunc,
				task replicationTask,
				r *historyReplicatorImpl,
			) error {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					branchIndex int,
					releaseFn execution.ReleaseFunc,
					task replicationTask,
					r *historyReplicatorImpl,
				) error {
					return nil
				}
				return fn
			},
			expectError: nil,
		},
		"Case4: case nil with applyNonStartEventsPrepareMutableStateFn error": {
			mockExecutionCacheAffordance: func(mockExecutionCache *execution.MockCache, mockExecutionContext *execution.MockContext) {
				mockExecutionCache.EXPECT().GetOrCreateWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockExecutionContext, func(err error) {}, nil).Times(1)
			},
			mockReplicationTaskAffordance: func(mockReplicationTask *MockreplicationTask) {
				mockReplicationTask.EXPECT().getDomainID().Return("test-domain-id").Times(1)
				mockReplicationTask.EXPECT().getExecution().Return(&types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				}).Times(1)
				mockReplicationTask.EXPECT().getFirstEvent().Return(&types.HistoryEvent{
					EventType: &workflowExecutionType,
				}).Times(1)
				mockReplicationTask.EXPECT().getVersion().Return(int64(1)).Times(1)
			},
			mockExecutionContextAffordance: func(mockExecutionContext *execution.MockContext, mockExecutionMutableState *execution.MockMutableState) {
				mockExecutionContext.EXPECT().LoadWorkflowExecutionWithTaskVersion(gomock.Any(), gomock.Any()).
					Return(mockExecutionMutableState, nil).Times(1)
			},
			mockMutableStateAffordance: func(mockExecutionMutableState *execution.MockMutableState) {
				mockExecutionMutableState.EXPECT().GetVersionHistories().Return(&persistence.VersionHistories{
					CurrentVersionHistoryIndex: 1,
					Histories:                  nil,
				}).Times(1)
			},
			applyNonStartEventsPrepareBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				task replicationTask,
				newBranchManager newBranchManagerFn,
			) (bool, int, error) {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					task replicationTask,
					newBranchManager newBranchManagerFn,
				) (bool, int, error) {
					return true, 0, nil
				}
				return fn
			},
			applyNonStartEventsPrepareMutableStateFnAffordance: func(mockExecutionMutableState *execution.MockMutableState) func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				branchIndex int,
				task replicationTask,
				newConflictResolver newConflictResolverFn,
			) (execution.MutableState, bool, error) {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					branchIndex int,
					task replicationTask,
					newConflictResolver newConflictResolverFn,
				) (execution.MutableState, bool, error) {
					return nil, false, fmt.Errorf("test error")
				}
				return fn
			},
			applyNonStartEventsToCurrentBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				isRebuilt bool,
				releaseFn execution.ReleaseFunc,
				task replicationTask,
				newStateBuilder newStateBuilderFn,
				clusterMetadata cluster.Metadata,
				shard shard.Context,
				logger log.Logger,
				transactionManager transactionManager,
			) error {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					isRebuilt bool,
					releaseFn execution.ReleaseFunc,
					task replicationTask,
					newStateBuilder newStateBuilderFn,
					clusterMetadata cluster.Metadata,
					shard shard.Context,
					logger log.Logger,
					transactionManager transactionManager,
				) error {
					return nil
				}
				return fn
			},
			applyNonStartEventsToNoneCurrentBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				branchIndex int,
				releaseFn execution.ReleaseFunc,
				task replicationTask,
				r *historyReplicatorImpl,
			) error {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					branchIndex int,
					releaseFn execution.ReleaseFunc,
					task replicationTask,
					r *historyReplicatorImpl,
				) error {
					return nil
				}
				return fn
			},
			expectError: fmt.Errorf("test error"),
		},
		"Case5: case nil with CurrentVersionHistoryIndex() == branchIndex": {
			mockExecutionCacheAffordance: func(mockExecutionCache *execution.MockCache, mockExecutionContext *execution.MockContext) {
				mockExecutionCache.EXPECT().GetOrCreateWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockExecutionContext, func(err error) {}, nil).Times(1)
			},
			mockReplicationTaskAffordance: func(mockReplicationTask *MockreplicationTask) {
				mockReplicationTask.EXPECT().getDomainID().Return("test-domain-id").Times(1)
				mockReplicationTask.EXPECT().getExecution().Return(&types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				}).Times(1)
				mockReplicationTask.EXPECT().getFirstEvent().Return(&types.HistoryEvent{
					EventType: &workflowExecutionType,
				}).Times(1)
				mockReplicationTask.EXPECT().getVersion().Return(int64(1)).Times(1)
			},
			mockExecutionContextAffordance: func(mockExecutionContext *execution.MockContext, mockExecutionMutableState *execution.MockMutableState) {
				mockExecutionContext.EXPECT().LoadWorkflowExecutionWithTaskVersion(gomock.Any(), gomock.Any()).
					Return(mockExecutionMutableState, nil).Times(1)
			},
			mockMutableStateAffordance: func(mockExecutionMutableState *execution.MockMutableState) {
				mockExecutionMutableState.EXPECT().GetVersionHistories().Return(&persistence.VersionHistories{
					CurrentVersionHistoryIndex: 1,
					Histories:                  nil,
				}).Times(2)
			},
			applyNonStartEventsPrepareBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				task replicationTask,
				newBranchManager newBranchManagerFn,
			) (bool, int, error) {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					task replicationTask,
					newBranchManager newBranchManagerFn,
				) (bool, int, error) {
					return true, 1, nil
				}
				return fn
			},
			applyNonStartEventsPrepareMutableStateFnAffordance: func(mockExecutionMutableState *execution.MockMutableState) func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				branchIndex int,
				task replicationTask,
				newConflictResolver newConflictResolverFn,
			) (execution.MutableState, bool, error) {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					branchIndex int,
					task replicationTask,
					newConflictResolver newConflictResolverFn,
				) (execution.MutableState, bool, error) {
					return mockExecutionMutableState, false, nil
				}
				return fn
			},
			applyNonStartEventsToCurrentBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				isRebuilt bool,
				releaseFn execution.ReleaseFunc,
				task replicationTask,
				newStateBuilder newStateBuilderFn,
				clusterMetadata cluster.Metadata,
				shard shard.Context,
				logger log.Logger,
				transactionManager transactionManager,
			) error {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					isRebuilt bool,
					releaseFn execution.ReleaseFunc,
					task replicationTask,
					newStateBuilder newStateBuilderFn,
					clusterMetadata cluster.Metadata,
					shard shard.Context,
					logger log.Logger,
					transactionManager transactionManager,
				) error {
					return nil
				}
				return fn
			},
			applyNonStartEventsToNoneCurrentBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				branchIndex int,
				releaseFn execution.ReleaseFunc,
				task replicationTask,
				r *historyReplicatorImpl,
			) error {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					branchIndex int,
					releaseFn execution.ReleaseFunc,
					task replicationTask,
					r *historyReplicatorImpl,
				) error {
					return nil
				}
				return fn
			},
			expectError: nil,
		},
		"Case6: case nil with CurrentVersionHistoryIndex() != branchIndex": {
			mockExecutionCacheAffordance: func(mockExecutionCache *execution.MockCache, mockExecutionContext *execution.MockContext) {
				mockExecutionCache.EXPECT().GetOrCreateWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockExecutionContext, func(err error) {}, nil).Times(1)
			},
			mockReplicationTaskAffordance: func(mockReplicationTask *MockreplicationTask) {
				mockReplicationTask.EXPECT().getDomainID().Return("test-domain-id").Times(1)
				mockReplicationTask.EXPECT().getExecution().Return(&types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				}).Times(1)
				mockReplicationTask.EXPECT().getFirstEvent().Return(&types.HistoryEvent{
					EventType: &workflowExecutionType,
				}).Times(1)
				mockReplicationTask.EXPECT().getVersion().Return(int64(1)).Times(1)
			},
			mockExecutionContextAffordance: func(mockExecutionContext *execution.MockContext, mockExecutionMutableState *execution.MockMutableState) {
				mockExecutionContext.EXPECT().LoadWorkflowExecutionWithTaskVersion(gomock.Any(), gomock.Any()).
					Return(mockExecutionMutableState, nil).Times(1)
			},
			mockMutableStateAffordance: func(mockExecutionMutableState *execution.MockMutableState) {
				mockExecutionMutableState.EXPECT().GetVersionHistories().Return(&persistence.VersionHistories{
					CurrentVersionHistoryIndex: 1,
					Histories:                  nil,
				}).Times(2)
			},
			applyNonStartEventsPrepareBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				task replicationTask,
				newBranchManager newBranchManagerFn,
			) (bool, int, error) {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					task replicationTask,
					newBranchManager newBranchManagerFn,
				) (bool, int, error) {
					return true, 2, nil
				}
				return fn
			},
			applyNonStartEventsPrepareMutableStateFnAffordance: func(mockExecutionMutableState *execution.MockMutableState) func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				branchIndex int,
				task replicationTask,
				newConflictResolver newConflictResolverFn,
			) (execution.MutableState, bool, error) {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					branchIndex int,
					task replicationTask,
					newConflictResolver newConflictResolverFn,
				) (execution.MutableState, bool, error) {
					return mockExecutionMutableState, false, nil
				}
				return fn
			},
			applyNonStartEventsToCurrentBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				isRebuilt bool,
				releaseFn execution.ReleaseFunc,
				task replicationTask,
				newStateBuilder newStateBuilderFn,
				clusterMetadata cluster.Metadata,
				shard shard.Context,
				logger log.Logger,
				transactionManager transactionManager,
			) error {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					isRebuilt bool,
					releaseFn execution.ReleaseFunc,
					task replicationTask,
					newStateBuilder newStateBuilderFn,
					clusterMetadata cluster.Metadata,
					shard shard.Context,
					logger log.Logger,
					transactionManager transactionManager,
				) error {
					return nil
				}
				return fn
			},
			applyNonStartEventsToNoneCurrentBranchFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				branchIndex int,
				releaseFn execution.ReleaseFunc,
				task replicationTask,
				r *historyReplicatorImpl,
			) error {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					branchIndex int,
					releaseFn execution.ReleaseFunc,
					task replicationTask,
					r *historyReplicatorImpl,
				) error {
					return nil
				}
				return fn
			},
			expectError: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			// mock objects
			replicator := createTestHistoryReplicator(t)
			mockReplicationTask := NewMockreplicationTask(ctrl)
			mockExecutionCache := execution.NewMockCache(ctrl)
			mockExecutionContext := execution.NewMockContext(ctrl)
			mockExecutionMutableState := execution.NewMockMutableState(ctrl)
			mockMetricsClient := metrics.NewNoopMetricsClient()
			replicator.executionCache = mockExecutionCache
			replicator.metricsClient = mockMetricsClient

			// mock functions
			test.mockReplicationTaskAffordance(mockReplicationTask)
			test.mockExecutionCacheAffordance(mockExecutionCache, mockExecutionContext)
			test.mockExecutionContextAffordance(mockExecutionContext, mockExecutionMutableState)
			test.mockMutableStateAffordance(mockExecutionMutableState)

			// parameter functions affordance
			replicator.applyNonStartEventsPrepareBranchFn = test.applyNonStartEventsPrepareBranchFnAffordance()
			replicator.applyNonStartEventsPrepareMutableStateFn = test.applyNonStartEventsPrepareMutableStateFnAffordance(mockExecutionMutableState)
			replicator.applyNonStartEventsToCurrentBranchFn = test.applyNonStartEventsToCurrentBranchFnAffordance()
			replicator.applyNonStartEventsToNoneCurrentBranchFn = test.applyNonStartEventsToNoneCurrentBranchFnAffordance()

			assert.Equal(t, replicator.applyEvents(ctx.Background(), mockReplicationTask), test.expectError)
		})
	}
}

func Test_applyEvents_defaultCase_errorAndDefault(t *testing.T) {
	workflowExecutionType := types.EventTypeWorkflowExecutionCompleted

	tests := map[string]struct {
		mockExecutionCacheAffordance                       func(mockExecutionCache *execution.MockCache, mockExecutionContext *execution.MockContext)
		mockReplicationTaskAffordance                      func(mockReplicationTask *MockreplicationTask)
		mockExecutionContextAffordance                     func(mockExecutionContext *execution.MockContext, mockExecutionMutableState *execution.MockMutableState)
		mockMutableStateAffordance                         func(mockExecutionMutableState *execution.MockMutableState)
		applyNonStartEventsMissingMutableStateFnAffordance func() func(ctx ctx.Context,
			newContext execution.Context,
			task replicationTask,
			newWorkflowResetter newWorkflowResetterFn,
		) (execution.MutableState, error)
		applyNonStartEventsResetWorkflowFnAffordance func() func(ctx ctx.Context,
			context execution.Context,
			mutableState execution.MutableState,
			task replicationTask,
			newStateBuilder newStateBuilderFn,
			transactionManager transactionManager,
			clusterMetadata cluster.Metadata,
			logger log.Logger,
			shard shard.Context,
		) error
		expectError error
	}{
		"Case1-1: case EntityNotExistsError + applyNonStartEventsMissingMutableStateFn error": {
			mockExecutionCacheAffordance: func(mockExecutionCache *execution.MockCache, mockExecutionContext *execution.MockContext) {
				mockExecutionCache.EXPECT().GetOrCreateWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockExecutionContext, func(err error) {}, nil).Times(1)
			},
			mockReplicationTaskAffordance: func(mockReplicationTask *MockreplicationTask) {
				mockReplicationTask.EXPECT().getDomainID().Return("test-domain-id").Times(1)
				mockReplicationTask.EXPECT().getExecution().Return(&types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				}).Times(1)
				mockReplicationTask.EXPECT().getFirstEvent().Return(&types.HistoryEvent{
					EventType: &workflowExecutionType,
				}).Times(1)
				mockReplicationTask.EXPECT().getVersion().Return(int64(1)).Times(1)
			},
			mockExecutionContextAffordance: func(mockExecutionContext *execution.MockContext, mockExecutionMutableState *execution.MockMutableState) {
				mockExecutionContext.EXPECT().LoadWorkflowExecutionWithTaskVersion(gomock.Any(), gomock.Any()).
					Return(mockExecutionMutableState, &types.EntityNotExistsError{}).Times(1)
			},
			mockMutableStateAffordance: func(mockExecutionMutableState *execution.MockMutableState) {
				return
			},
			applyNonStartEventsMissingMutableStateFnAffordance: func() func(ctx ctx.Context,
				newContext execution.Context,
				task replicationTask,
				newWorkflowResetter newWorkflowResetterFn,
			) (execution.MutableState, error) {
				fn := func(ctx ctx.Context,
					newContext execution.Context,
					task replicationTask,
					newWorkflowResetter newWorkflowResetterFn,
				) (execution.MutableState, error) {
					return nil, fmt.Errorf("test error")
				}
				return fn
			},
			applyNonStartEventsResetWorkflowFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				task replicationTask,
				newStateBuilder newStateBuilderFn,
				transactionManager transactionManager,
				clusterMetadata cluster.Metadata,
				logger log.Logger,
				shard shard.Context,
			) error {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					task replicationTask,
					newStateBuilder newStateBuilderFn,
					transactionManager transactionManager,
					clusterMetadata cluster.Metadata,
					logger log.Logger,
					shard shard.Context,
				) error {
					return nil
				}
				return fn
			},
			expectError: fmt.Errorf("test error"),
		},
		"Case1-2: case EntityNotExistsError": {
			mockExecutionCacheAffordance: func(mockExecutionCache *execution.MockCache, mockExecutionContext *execution.MockContext) {
				mockExecutionCache.EXPECT().GetOrCreateWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockExecutionContext, func(err error) {}, nil).Times(1)
			},
			mockReplicationTaskAffordance: func(mockReplicationTask *MockreplicationTask) {
				mockReplicationTask.EXPECT().getDomainID().Return("test-domain-id").Times(1)
				mockReplicationTask.EXPECT().getExecution().Return(&types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				}).Times(1)
				mockReplicationTask.EXPECT().getFirstEvent().Return(&types.HistoryEvent{
					EventType: &workflowExecutionType,
				}).Times(1)
				mockReplicationTask.EXPECT().getVersion().Return(int64(1)).Times(1)
			},
			mockExecutionContextAffordance: func(mockExecutionContext *execution.MockContext, mockExecutionMutableState *execution.MockMutableState) {
				mockExecutionContext.EXPECT().LoadWorkflowExecutionWithTaskVersion(gomock.Any(), gomock.Any()).
					Return(mockExecutionMutableState, &types.EntityNotExistsError{}).Times(1)
			},
			mockMutableStateAffordance: func(mockExecutionMutableState *execution.MockMutableState) {
				return
			},
			applyNonStartEventsMissingMutableStateFnAffordance: func() func(ctx ctx.Context,
				newContext execution.Context,
				task replicationTask,
				newWorkflowResetter newWorkflowResetterFn,
			) (execution.MutableState, error) {
				fn := func(ctx ctx.Context,
					newContext execution.Context,
					task replicationTask,
					newWorkflowResetter newWorkflowResetterFn,
				) (execution.MutableState, error) {
					return nil, nil
				}
				return fn
			},
			expectError: nil,
			applyNonStartEventsResetWorkflowFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				task replicationTask,
				newStateBuilder newStateBuilderFn,
				transactionManager transactionManager,
				clusterMetadata cluster.Metadata,
				logger log.Logger,
				shard shard.Context,
			) error {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					task replicationTask,
					newStateBuilder newStateBuilderFn,
					transactionManager transactionManager,
					clusterMetadata cluster.Metadata,
					logger log.Logger,
					shard shard.Context,
				) error {
					return nil
				}
				return fn
			},
		},
		"Case1-3: case other errors": {
			mockExecutionCacheAffordance: func(mockExecutionCache *execution.MockCache, mockExecutionContext *execution.MockContext) {
				mockExecutionCache.EXPECT().GetOrCreateWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(mockExecutionContext, func(err error) {}, nil).Times(1)
			},
			mockReplicationTaskAffordance: func(mockReplicationTask *MockreplicationTask) {
				mockReplicationTask.EXPECT().getDomainID().Return("test-domain-id").Times(1)
				mockReplicationTask.EXPECT().getExecution().Return(&types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				}).Times(1)
				mockReplicationTask.EXPECT().getFirstEvent().Return(&types.HistoryEvent{
					EventType: &workflowExecutionType,
				}).Times(1)
				mockReplicationTask.EXPECT().getVersion().Return(int64(1)).Times(1)
			},
			mockExecutionContextAffordance: func(mockExecutionContext *execution.MockContext, mockExecutionMutableState *execution.MockMutableState) {
				mockExecutionContext.EXPECT().LoadWorkflowExecutionWithTaskVersion(gomock.Any(), gomock.Any()).
					Return(mockExecutionMutableState, fmt.Errorf("test-error")).Times(1)
			},
			mockMutableStateAffordance: func(mockExecutionMutableState *execution.MockMutableState) {
				return
			},
			applyNonStartEventsMissingMutableStateFnAffordance: func() func(ctx ctx.Context,
				newContext execution.Context,
				task replicationTask,
				newWorkflowResetter newWorkflowResetterFn,
			) (execution.MutableState, error) {
				fn := func(ctx ctx.Context,
					newContext execution.Context,
					task replicationTask,
					newWorkflowResetter newWorkflowResetterFn,
				) (execution.MutableState, error) {
					return nil, nil
				}
				return fn
			},
			expectError: fmt.Errorf("test-error"),
			applyNonStartEventsResetWorkflowFnAffordance: func() func(ctx ctx.Context,
				context execution.Context,
				mutableState execution.MutableState,
				task replicationTask,
				newStateBuilder newStateBuilderFn,
				transactionManager transactionManager,
				clusterMetadata cluster.Metadata,
				logger log.Logger,
				shard shard.Context,
			) error {
				fn := func(ctx ctx.Context,
					context execution.Context,
					mutableState execution.MutableState,
					task replicationTask,
					newStateBuilder newStateBuilderFn,
					transactionManager transactionManager,
					clusterMetadata cluster.Metadata,
					logger log.Logger,
					shard shard.Context,
				) error {
					return nil
				}
				return fn
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			// mock objects
			replicator := createTestHistoryReplicator(t)
			mockReplicationTask := NewMockreplicationTask(ctrl)
			mockExecutionCache := execution.NewMockCache(ctrl)
			mockExecutionContext := execution.NewMockContext(ctrl)
			mockExecutionMutableState := execution.NewMockMutableState(ctrl)
			replicator.executionCache = mockExecutionCache

			// mock functions
			test.mockReplicationTaskAffordance(mockReplicationTask)
			test.mockExecutionCacheAffordance(mockExecutionCache, mockExecutionContext)
			test.mockExecutionContextAffordance(mockExecutionContext, mockExecutionMutableState)
			test.mockMutableStateAffordance(mockExecutionMutableState)

			// parameter functions affordance
			replicator.applyNonStartEventsMissingMutableStateFn = test.applyNonStartEventsMissingMutableStateFnAffordance()
			replicator.applyNonStartEventsResetWorkflowFn = test.applyNonStartEventsResetWorkflowFnAffordance()

			assert.Equal(t, replicator.applyEvents(ctx.Background(), mockReplicationTask), test.expectError)
		})
	}
}
