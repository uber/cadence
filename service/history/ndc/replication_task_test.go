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
	"github.com/uber/cadence/common/log/loggerimpl"
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

	task, err := newReplicationTask(clusterMetadata, historySerializer, taskStartTime, loggerimpl.NewNopLogger(), request)
	require.NoError(t, err)
	assert.True(t, task.isWorkflowReset())
}
