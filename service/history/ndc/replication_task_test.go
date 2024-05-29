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
	"testing"
	"time"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

func TestReplicationTaskResetEvent(t *testing.T) {
	clusterMetadata := cluster.GetTestClusterMetadata(true)
	historySerializer := persistence.NewPayloadSerializer()
	taskStartTime := time.Now()
	domainID := uuid.New()
	workflowID := uuid.New()
	runID := uuid.New()
	versionHistoryItems := []*types.VersionHistoryItem{}
	versionHistoryItems = append(versionHistoryItems, persistence.NewVersionHistoryItem(3, 0).ToInternalType())

	eventType := types.EventTypeDecisionTaskFailed
	resetCause := types.DecisionTaskFailedCauseResetWorkflow
	event := &types.HistoryEvent{
		ID:        3,
		EventType: &eventType,
		Version:   0,
		DecisionTaskFailedEventAttributes: &types.DecisionTaskFailedEventAttributes{
			BaseRunID:        uuid.New(),
			ForkEventVersion: 0,
			NewRunID:         runID,
			Cause:            &resetCause,
		},
	}
	events := []*types.HistoryEvent{}
	events = append(events, event)
	eventsBlob, err := historySerializer.SerializeBatchEvents(events, common.EncodingTypeThriftRW)
	require.NoError(t, err)
	request := &types.ReplicateEventsV2Request{
		DomainUUID: domainID,
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: workflowID,
			RunID:      runID,
		},
		VersionHistoryItems: versionHistoryItems,
		Events:              eventsBlob.ToInternal(),
		NewRunEvents:        nil,
	}

	task, err := newReplicationTask(clusterMetadata, historySerializer, taskStartTime, testlogger.New(t), request)
	require.NoError(t, err)
	assert.True(t, task.isWorkflowReset())
}

func TestNewReplicationTask(t *testing.T) {
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
		"Case3-3: fail case in validateReplicateEventsRequest with invalid workflow execution": {
			request: &types.ReplicateEventsV2Request{
				DomainUUID:          "12345678-1234-5678-9012-123456789011",
				WorkflowExecution:   nil,
				VersionHistoryItems: nil,
				Events:              nil,
				NewRunEvents:        nil,
			},
			expectedErrorMsg: "invalid execution",
		},
		"Case3-4: fail case in validateReplicateEventsRequest with event is empty": {
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
		"Case3-5: fail case in validateReplicateEventsRequest with DeserializeBatchEvents error": {
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
		"Case3-6: fail case in validateReplicateEventsRequest with event ID mismatch": {
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
		"Case3-7: fail case in validateReplicateEventsRequest with event version mismatch": {
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
		"Case3-8: fail case in validateReplicateEventsRequest with ErrEventVersionMismatch": {
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
			_, err := newReplicationTask(replicator.clusterMetadata,
				replicator.historySerializer,
				time.Now(),
				replicator.logger,
				test.request,
			)
			assert.Equal(t, test.expectedErrorMsg, err.Error())
		})
	}
}
