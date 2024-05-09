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
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/resource"
	"github.com/uber/cadence/service/history/shard"
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

	// test newBranchManager function in history replicator
	mockExecutionContext := execution.NewMockContext(ctrl)
	mockExecutionMutableState := execution.NewMockMutableState(ctrl)
	assert.NotNil(t, testReplicatorImpl.newBranchManager(mockExecutionContext, mockExecutionMutableState, log.NewNoop()))
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

	// test newConflictResolver function in history replicator
	mockEventsCache := events.NewMockCache(ctrl)
	mockShard.EXPECT().GetEventsCache().Return(mockEventsCache).Times(1)
	mockShard.EXPECT().GetShardID().Return(0).Times(1)

	mockExecutionContext := execution.NewMockContext(ctrl)
	mockExecutionMutableState := execution.NewMockMutableState(ctrl)
	assert.NotNil(t, testReplicatorImpl.newConflictResolver(mockExecutionContext, mockExecutionMutableState, log.NewNoop()))
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

	// test newWorkflowResetter function in history replicator
	mockEventsCache := events.NewMockCache(ctrl)
	mockShard.EXPECT().GetEventsCache().Return(mockEventsCache).Times(1)
	mockShard.EXPECT().GetShardID().Return(0).Times(1)

	mockExecutionContext := execution.NewMockContext(ctrl)
	assert.NotNil(t, testReplicatorImpl.newWorkflowResetter(
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

	// test newStateBuilder function in history replicator
	mockExecutionMutableState := execution.NewMockMutableState(ctrl)
	assert.NotNil(t, testReplicatorImpl.newStateBuilder(mockExecutionMutableState, log.NewNoop()))
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

	// test newMutableState function in history replicator
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
	assert.NotNil(t, testReplicatorImpl.newMutableState(mockDomainCacheEntry, log.NewNoop()))
}

func TestApplyEvents_newReplicationTask_validateReplicateEventsRequest(t *testing.T) {
	encodingType1 := types.EncodingTypeJSON
	encodingType2 := types.EncodingTypeThriftRW

	historyEvent1 := &types.HistoryEvent{
		Version: 0,
	}
	historyEvent2 := &types.HistoryEvent{
		Version: 1,
	}

	historyEvents := []*types.HistoryEvent{historyEvent1, historyEvent2}
	serializer := persistence.NewPayloadSerializer()
	serializedEvents, err := serializer.SerializeBatchEvents(historyEvents, common.EncodingTypeThriftRW)
	assert.NoError(t, err)

	historyEvent3 := &types.HistoryEvent{
		ID:      0,
		Version: 2,
	}
	historyEvent4 := &types.HistoryEvent{
		ID:      1,
		Version: 3,
	}

	historyEvents2 := []*types.HistoryEvent{historyEvent3}
	serializedEvents2, err := serializer.SerializeBatchEvents(historyEvents2, common.EncodingTypeThriftRW)
	assert.NoError(t, err)

	historyEvents3 := []*types.HistoryEvent{historyEvent3, historyEvent4}
	serializedEvents3, err := serializer.SerializeBatchEvents(historyEvents3, common.EncodingTypeThriftRW)
	assert.NoError(t, err)

	historyEvents4 := []*types.HistoryEvent{historyEvent4}
	serializedEvents4, err := serializer.SerializeBatchEvents(historyEvents4, common.EncodingTypeThriftRW)
	assert.NoError(t, err)

	tests := map[string]struct {
		request          *types.ReplicateEventsV2Request
		eventsAffordance func() []*types.HistoryEvent
		expectedErrorMsg string
	}{
		"Case1: empty case": {
			request:          &types.ReplicateEventsV2Request{},
			expectedErrorMsg: "invalid domain ID",
		},
		"Case2: nil case": {
			request:          nil,
			expectedErrorMsg: "invalid domain ID",
		},
		"Case3-1: fail case with invalid domain id": {
			request: &types.ReplicateEventsV2Request{
				DomainUUID: "12345678-1234-5678-9012-",
				WorkflowExecution: &types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				},
				VersionHistoryItems: nil,
				Events:              nil,
				NewRunEvents:        nil,
			},
			expectedErrorMsg: "invalid domain ID",
		},
		"Case3-2: fail case in validateReplicateEventsRequest with invalid run id": {
			request: &types.ReplicateEventsV2Request{
				DomainUUID: "12345678-1234-5678-9012-123456789011",
				WorkflowExecution: &types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				},
				VersionHistoryItems: nil,
				Events:              nil,
				NewRunEvents:        nil,
			},
			expectedErrorMsg: "invalid run ID",
		},
		"Case3-3: fail case in validateReplicateEventsRequest with invalid run id": {
			request: &types.ReplicateEventsV2Request{
				DomainUUID: "12345678-1234-5678-9012-123456789011",
				WorkflowExecution: &types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				},
				VersionHistoryItems: nil,
				Events:              nil,
				NewRunEvents:        nil,
			},
			expectedErrorMsg: "invalid run ID",
		},
		"Case3-4: fail case in validateReplicateEventsRequest with invalid workflow execution": {
			request: &types.ReplicateEventsV2Request{
				DomainUUID:          "12345678-1234-5678-9012-123456789011",
				WorkflowExecution:   nil,
				VersionHistoryItems: nil,
				Events:              nil,
				NewRunEvents:        nil,
			},
			expectedErrorMsg: "invalid execution",
		},
		"Case3-5: fail case in validateReplicateEventsRequest with event is empty": {
			request: &types.ReplicateEventsV2Request{
				DomainUUID: "12345678-1234-5678-9012-123456789011",
				WorkflowExecution: &types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "12345678-1234-5678-9012-123456789012",
				},
				VersionHistoryItems: nil,
				Events:              nil,
				NewRunEvents:        nil,
			},
			expectedErrorMsg: "encounter empty history batch",
		},
		"Case3-6: fail case in validateReplicateEventsRequest with DeserializeBatchEvents error": {
			request: &types.ReplicateEventsV2Request{
				DomainUUID: "12345678-1234-5678-9012-123456789011",
				WorkflowExecution: &types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "12345678-1234-5678-9012-123456789012",
				},
				VersionHistoryItems: nil,
				Events: &types.DataBlob{
					EncodingType: &encodingType1,
					Data:         []byte("test-data"),
				},
				NewRunEvents: nil,
			},
			expectedErrorMsg: "cadence deserialization error: DeserializeBatchEvents encoding: \"thriftrw\", error: Invalid binary encoding version.",
		},
		"Case3-7: fail case in validateReplicateEventsRequest with event ID mismatch": {
			request: &types.ReplicateEventsV2Request{
				DomainUUID: "12345678-1234-5678-9012-123456789011",
				WorkflowExecution: &types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "12345678-1234-5678-9012-123456789012",
				},
				VersionHistoryItems: nil,
				Events: &types.DataBlob{
					EncodingType: &encodingType2,
					Data:         serializedEvents.Data,
				},
				NewRunEvents: nil,
			},
			expectedErrorMsg: "event ID mismatch",
		},
		"Case3-8: fail case in validateReplicateEventsRequest with event version mismatch": {
			request: &types.ReplicateEventsV2Request{
				DomainUUID: "12345678-1234-5678-9012-123456789011",
				WorkflowExecution: &types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "12345678-1234-5678-9012-123456789012",
				},
				VersionHistoryItems: nil,
				Events: &types.DataBlob{
					EncodingType: &encodingType2,
					Data:         serializedEvents2.Data,
				},
				NewRunEvents: &types.DataBlob{
					EncodingType: &encodingType2,
					Data:         serializedEvents3.Data,
				},
			},
			expectedErrorMsg: "event version mismatch",
		},
		"Case3-9: fail case in validateReplicateEventsRequest with ErrEventVersionMismatch": {
			request: &types.ReplicateEventsV2Request{
				DomainUUID: "12345678-1234-5678-9012-123456789011",
				WorkflowExecution: &types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "12345678-1234-5678-9012-123456789012",
				},
				VersionHistoryItems: nil,
				Events: &types.DataBlob{
					EncodingType: &encodingType2,
					Data:         serializedEvents2.Data,
				},
				NewRunEvents: &types.DataBlob{
					EncodingType: &encodingType2,
					Data:         serializedEvents4.Data,
				},
			},
			expectedErrorMsg: "event version mismatch",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			replicator := createTestHistoryReplicator(t)
			err := replicator.ApplyEvents(context.Background(), test.request)
			assert.Equal(t, test.expectedErrorMsg, err.Error())
		})
	}
}
