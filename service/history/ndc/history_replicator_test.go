package ndc

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/resource"
	"github.com/uber/cadence/service/history/shard"

	"testing"
)

func TestNewHistoryReplicator(t *testing.T) {
	assert.NotPanics(t, func() {
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

		NewHistoryReplicator(mockShard, testExecutionCache, mockEventsReapplier, log.NewNoop())
	})
}

func TestNewHistoryReplicator_newBranchManager(t *testing.T) {
	assert.NotPanics(t, func() {
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

		// test newBranchManager function in history replicator
		assert.NotPanics(t, func() {
			mockExecutionContext := execution.NewMockContext(ctrl)
			mockExecutionMutableState := execution.NewMockMutableState(ctrl)
			testReplicatorImpl.newBranchManager(mockExecutionContext, mockExecutionMutableState, log.NewNoop())
		})
	})
}

func TestNewHistoryReplicator_newConflictResolver(t *testing.T) {
	assert.NotPanics(t, func() {
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

		// test newConflictResolver function in history replicator
		assert.NotPanics(t, func() {
			mockEventsCache := events.NewMockCache(ctrl)
			mockShard.EXPECT().GetEventsCache().Return(mockEventsCache).Times(1)
			mockShard.EXPECT().GetShardID().Return(0).Times(1)

			mockExecutionContext := execution.NewMockContext(ctrl)
			mockExecutionMutableState := execution.NewMockMutableState(ctrl)
			testReplicatorImpl.newConflictResolver(mockExecutionContext, mockExecutionMutableState, log.NewNoop())
		})
	})
}

func TestNewHistoryReplicator_newWorkflowResetter(t *testing.T) {
	assert.NotPanics(t, func() {
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

		// test newWorkflowResetter function in history replicator
		assert.NotPanics(t, func() {
			mockEventsCache := events.NewMockCache(ctrl)
			mockShard.EXPECT().GetEventsCache().Return(mockEventsCache).Times(1)
			mockShard.EXPECT().GetShardID().Return(0).Times(1)

			mockExecutionContext := execution.NewMockContext(ctrl)
			testReplicatorImpl.newWorkflowResetter(
				"test-domain-id",
				"test-workflow-id",
				"test-base-run-id",
				mockExecutionContext,
				"test-run-id",
				log.NewNoop(),
			)
		})
	})
}

func TestNewHistoryReplicator_newStateBuilder(t *testing.T) {
	assert.NotPanics(t, func() {
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

		// test newStateBuilder function in history replicator
		assert.NotPanics(t, func() {
			mockExecutionMutableState := execution.NewMockMutableState(ctrl)
			testReplicatorImpl.newStateBuilder(mockExecutionMutableState, log.NewNoop())
		})
	})
}

func TestNewHistoryReplicator_newMutableState(t *testing.T) {
	assert.NotPanics(t, func() {
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

		// test newMutableState function in history replicator
		assert.NotPanics(t, func() {
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
			testReplicatorImpl.newMutableState(mockDomainCacheEntry, log.NewNoop())
		})
	})
}
