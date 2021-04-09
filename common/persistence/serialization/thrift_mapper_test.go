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

package serialization

import (
	"math/rand"
	"testing"
	"time"

	"github.com/uber/cadence/common/types"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
)

func TestShardInfo(t *testing.T) {
	expected := &ShardInfo{
		StolenSinceRenew:                      common.Int32Ptr(int32(rand.Intn(1000))),
		UpdatedAt:                             common.TimePtr(time.Now()),
		ReplicationAckLevel:                   common.Int64Ptr(int64(rand.Intn(1000))),
		TransferAckLevel:                      common.Int64Ptr(int64(rand.Intn(1000))),
		TimerAckLevel:                         common.TimePtr(time.Now()),
		DomainNotificationVersion:             common.Int64Ptr(int64(rand.Intn(1000))),
		ClusterTransferAckLevel:               map[string]int64{"key_1": int64(rand.Intn(1000)), "key_2": int64(rand.Intn(1000))},
		ClusterTimerAckLevel:                  map[string]time.Time{"key_1": time.Now(), "key_2": time.Now()},
		Owner:                                 common.StringPtr("test_owner"),
		ClusterReplicationLevel:               map[string]int64{"key_1": int64(rand.Intn(1000)), "key_2": int64(rand.Intn(1000))},
		PendingFailoverMarkers:                []byte("PendingFailoverMarkers"),
		PendingFailoverMarkersEncoding:        common.StringPtr("PendingFailoverMarkersEncoding"),
		ReplicationDlqAckLevel:                map[string]int64{"key_1": int64(rand.Intn(1000)), "key_2": int64(rand.Intn(1000))},
		TransferProcessingQueueStates:         []byte("TransferProcessingQueueStates"),
		TransferProcessingQueueStatesEncoding: common.StringPtr("TransferProcessingQueueStatesEncoding"),
		TimerProcessingQueueStates:            []byte("TimerProcessingQueueStates"),
		TimerProcessingQueueStatesEncoding:    common.StringPtr("TimerProcessingQueueStatesEncoding"),
	}
	actual := shardInfoFromThrift(shardInfoToThrift(expected))
	assert.Equal(t, expected.StolenSinceRenew, actual.StolenSinceRenew)
	assert.Equal(t, expected.UpdatedAt.Sub(*actual.UpdatedAt), time.Duration(0))
	assert.Equal(t, expected.ReplicationAckLevel, actual.ReplicationAckLevel)
	assert.Equal(t, expected.TransferAckLevel, actual.TransferAckLevel)
	assert.Equal(t, expected.TimerAckLevel.Sub(*actual.TimerAckLevel), time.Duration(0))
	assert.Equal(t, expected.DomainNotificationVersion, actual.DomainNotificationVersion)
	assert.Equal(t, expected.ClusterTransferAckLevel, actual.ClusterTransferAckLevel)
	assert.Equal(t, expected.Owner, actual.Owner)
	assert.Equal(t, expected.ClusterReplicationLevel, actual.ClusterReplicationLevel)
	assert.Equal(t, expected.PendingFailoverMarkers, actual.PendingFailoverMarkers)
	assert.Equal(t, expected.PendingFailoverMarkersEncoding, actual.PendingFailoverMarkersEncoding)
	assert.Equal(t, expected.ReplicationDlqAckLevel, actual.ReplicationDlqAckLevel)
	assert.Equal(t, expected.TransferProcessingQueueStates, actual.TransferProcessingQueueStates)
	assert.Equal(t, expected.TransferProcessingQueueStatesEncoding, actual.TransferProcessingQueueStatesEncoding)
	assert.Equal(t, expected.TimerProcessingQueueStates, actual.TimerProcessingQueueStates)
	assert.Equal(t, expected.TimerProcessingQueueStatesEncoding, actual.TimerProcessingQueueStatesEncoding)
	assert.Len(t, actual.ClusterTimerAckLevel, 2)
	assert.Contains(t, actual.ClusterTimerAckLevel, "key_1")
	assert.Contains(t, actual.ClusterTimerAckLevel, "key_2")
	assert.Equal(t, expected.ClusterTimerAckLevel["key_1"].Sub(actual.ClusterTimerAckLevel["key_1"]), time.Duration(0))
	assert.Equal(t, expected.ClusterTimerAckLevel["key_2"].Sub(actual.ClusterTimerAckLevel["key_2"]), time.Duration(0))
}

func TestDomainInfo(t *testing.T) {
	expected := &DomainInfo{
		Name:                        common.StringPtr("domain_name"),
		Description:                 common.StringPtr("description"),
		Owner:                       common.StringPtr("owner"),
		Status:                      common.Int32Ptr(int32(rand.Intn(1000))),
		Retention:                   common.DurationPtr(time.Duration(int64(rand.Intn(1000)))),
		EmitMetric:                  common.BoolPtr(true),
		ArchivalBucket:              common.StringPtr("archival_bucket"),
		ArchivalStatus:              common.Int16Ptr(int16(rand.Intn(1000))),
		ConfigVersion:               common.Int64Ptr(int64(rand.Intn(1000))),
		NotificationVersion:         common.Int64Ptr(int64(rand.Intn(1000))),
		FailoverNotificationVersion: common.Int64Ptr(int64(rand.Intn(1000))),
		FailoverVersion:             common.Int64Ptr(int64(rand.Intn(1000))),
		ActiveClusterName:           common.StringPtr("ActiveClusterName"),
		Clusters:                    []string{"cluster_a", "cluster_b"},
		Data:                        map[string]string{"key_1": "value_1", "key_2": "value_2"},
		BadBinaries:                 []byte("BadBinaries"),
		BadBinariesEncoding:         common.StringPtr("BadBinariesEncoding"),
		HistoryArchivalStatus:       common.Int16Ptr(int16(rand.Intn(1000))),
		HistoryArchivalURI:          common.StringPtr("HistoryArchivalURI"),
		VisibilityArchivalStatus:    common.Int16Ptr(int16(rand.Intn(1000))),
		VisibilityArchivalURI:       common.StringPtr("VisibilityArchivalURI"),
		FailoverEndTimestamp:        common.TimePtr(time.Now()),
		PreviousFailoverVersion:     common.Int64Ptr(int64(rand.Intn(1000))),
		LastUpdatedTimestamp:        common.TimePtr(time.Now()),
	}
	actual := domainInfoFromThrift(domainInfoToThrift(expected))
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Description, actual.Description)
	assert.Equal(t, expected.Owner, actual.Owner)
	assert.Equal(t, expected.Status, actual.Status)
	assert.True(t, (*expected.Retention-*actual.Retention) < time.Second)
	assert.Equal(t, expected.EmitMetric, actual.EmitMetric)
	assert.Equal(t, expected.ArchivalBucket, actual.ArchivalBucket)
	assert.Equal(t, expected.ArchivalStatus, actual.ArchivalStatus)
	assert.Equal(t, expected.ConfigVersion, actual.ConfigVersion)
	assert.Equal(t, expected.NotificationVersion, actual.NotificationVersion)
	assert.Equal(t, expected.FailoverNotificationVersion, actual.FailoverNotificationVersion)
	assert.Equal(t, expected.ActiveClusterName, actual.ActiveClusterName)
	assert.Equal(t, expected.Clusters, actual.Clusters)
	assert.Equal(t, expected.Data, actual.Data)
	assert.Equal(t, expected.BadBinaries, actual.BadBinaries)
	assert.Equal(t, expected.BadBinariesEncoding, actual.BadBinariesEncoding)
	assert.Equal(t, expected.HistoryArchivalStatus, actual.HistoryArchivalStatus)
	assert.Equal(t, expected.HistoryArchivalURI, actual.HistoryArchivalURI)
	assert.Equal(t, expected.VisibilityArchivalStatus, actual.VisibilityArchivalStatus)
	assert.Equal(t, expected.VisibilityArchivalURI, actual.VisibilityArchivalURI)
	assert.Equal(t, expected.FailoverEndTimestamp.Sub(*actual.FailoverEndTimestamp), time.Duration(0))
	assert.Equal(t, expected.PreviousFailoverVersion, actual.PreviousFailoverVersion)
	assert.Equal(t, expected.LastUpdatedTimestamp.Sub(*actual.LastUpdatedTimestamp), time.Duration(0))
}

func TestHistoryTreeInfo(t *testing.T) {
	expected := &HistoryTreeInfo{
		CreatedTimestamp: common.TimePtr(time.Now()),
		Ancestors: []*types.HistoryBranchRange{
			{
				BranchID:    common.StringPtr("branch_id"),
				BeginNodeID: common.Int64Ptr(int64(rand.Intn(1000))),
				EndNodeID:   common.Int64Ptr(int64(rand.Intn(1000))),
			},
			{
				BranchID:    common.StringPtr("branch_id"),
				BeginNodeID: common.Int64Ptr(int64(rand.Intn(1000))),
				EndNodeID:   common.Int64Ptr(int64(rand.Intn(1000))),
			},
		},
		Info: common.StringPtr("info"),
	}
	actual := historyTreeInfoFromThrift(historyTreeInfoToThrift(expected))
	assert.Equal(t, expected.CreatedTimestamp.Sub(*actual.CreatedTimestamp), time.Duration(0))
	assert.Equal(t, expected.Ancestors, actual.Ancestors)
	assert.Equal(t, expected.Info, actual.Info)
}

func TestWorkflowExecutionInfo(t *testing.T) {
	expected := &WorkflowExecutionInfo{
		ParentDomainID:                     UUID(uuid.New()),
		ParentWorkflowID:                   common.StringPtr("ParentWorkflowID"),
		ParentRunID:                        UUID(uuid.New()),
		InitiatedID:                        common.Int64Ptr(int64(rand.Intn(1000))),
		CompletionEventBatchID:             common.Int64Ptr(int64(rand.Intn(1000))),
		CompletionEvent:                    []byte("CompletionEvent"),
		CompletionEventEncoding:            common.StringPtr("CompletionEventEncoding"),
		TaskList:                           common.StringPtr("TaskList"),
		WorkflowTypeName:                   common.StringPtr("WorkflowTypeName"),
		WorkflowTimeout:                    common.DurationPtr(time.Minute * time.Duration(rand.Intn(10))),
		DecisionTaskTimeout:                common.DurationPtr(time.Minute * time.Duration(rand.Intn(10))),
		ExecutionContext:                   []byte("ExecutionContext"),
		State:                              common.Int32Ptr(int32(rand.Intn(1000))),
		CloseStatus:                        common.Int32Ptr(int32(rand.Intn(1000))),
		StartVersion:                       common.Int64Ptr(int64(rand.Intn(1000))),
		LastWriteEventID:                   common.Int64Ptr(int64(rand.Intn(1000))),
		LastEventTaskID:                    common.Int64Ptr(int64(rand.Intn(1000))),
		LastFirstEventID:                   common.Int64Ptr(int64(rand.Intn(1000))),
		LastProcessedEvent:                 common.Int64Ptr(int64(rand.Intn(1000))),
		StartTimestamp:                     common.TimePtr(time.Now()),
		LastUpdatedTimestamp:               common.TimePtr(time.Now()),
		DecisionVersion:                    common.Int64Ptr(int64(rand.Intn(1000))),
		DecisionScheduleID:                 common.Int64Ptr(int64(rand.Intn(1000))),
		DecisionStartedID:                  common.Int64Ptr(int64(rand.Intn(1000))),
		DecisionTimeout:                    common.DurationPtr(time.Minute * time.Duration(rand.Intn(10))),
		DecisionAttempt:                    common.Int64Ptr(int64(rand.Intn(1000))),
		DecisionStartedTimestamp:           common.TimePtr(time.Now()),
		DecisionScheduledTimestamp:         common.TimePtr(time.Now()),
		CancelRequested:                    common.BoolPtr(true),
		DecisionOriginalScheduledTimestamp: common.TimePtr(time.Now()),
		CreateRequestID:                    common.StringPtr("CreateRequestID"),
		DecisionRequestID:                  common.StringPtr("DecisionRequestID"),
		CancelRequestID:                    common.StringPtr("CancelRequestID"),
		StickyTaskList:                     common.StringPtr("StickyTaskList"),
		StickyScheduleToStartTimeout:       common.DurationPtr(time.Minute * time.Duration(rand.Intn(10))),
		RetryAttempt:                       common.Int64Ptr(int64(rand.Intn(1000))),
		RetryInitialInterval:               common.DurationPtr(time.Minute * time.Duration(rand.Intn(10))),
		RetryMaximumInterval:               common.DurationPtr(time.Minute * time.Duration(rand.Intn(10))),
		RetryMaximumAttempts:               common.Int32Ptr(int32(rand.Intn(1000))),
		RetryExpiration:                    common.DurationPtr(time.Minute * time.Duration(rand.Intn(10))),
		RetryBackoffCoefficient:            common.Float64Ptr(rand.Float64() * 1000),
		RetryExpirationTimestamp:           common.TimePtr(time.Now()),
		RetryNonRetryableErrors:            []string{"RetryNonRetryableErrors"},
		HasRetryPolicy:                     common.BoolPtr(true),
		CronSchedule:                       common.StringPtr("CronSchedule"),
		EventStoreVersion:                  common.Int32Ptr(int32(rand.Intn(1000))),
		EventBranchToken:                   []byte("EventBranchToken"),
		SignalCount:                        common.Int64Ptr(int64(rand.Intn(1000))),
		HistorySize:                        common.Int64Ptr(int64(rand.Intn(1000))),
		ClientLibraryVersion:               common.StringPtr("ClientLibraryVersion"),
		ClientFeatureVersion:               common.StringPtr("ClientFeatureVersion"),
		ClientImpl:                         common.StringPtr("ClientImpl"),
		AutoResetPoints:                    []byte("AutoResetPoints"),
		AutoResetPointsEncoding:            common.StringPtr("AutoResetPointsEncoding"),
		SearchAttributes:                   map[string][]byte{"key_1": []byte("SearchAttributes")},
		Memo:                               map[string][]byte{"key_1": []byte("Memo")},
		VersionHistories:                   []byte("VersionHistories"),
		VersionHistoriesEncoding:           common.StringPtr("VersionHistoriesEncoding"),
	}
	actual := workflowExecutionInfoFromThrift(workflowExecutionInfoToThrift(expected))
	assert.Equal(t, expected.ParentDomainID, actual.ParentDomainID)
	assert.Equal(t, expected.ParentWorkflowID, actual.ParentWorkflowID)
	assert.Equal(t, expected.ParentRunID, actual.ParentRunID)
	assert.Equal(t, expected.InitiatedID, actual.InitiatedID)
	assert.Equal(t, expected.CompletionEventBatchID, actual.CompletionEventBatchID)
	assert.Equal(t, expected.CompletionEvent, actual.CompletionEvent)
	assert.Equal(t, expected.CompletionEventEncoding, actual.CompletionEventEncoding)
	assert.Equal(t, expected.TaskList, actual.TaskList)
	assert.Equal(t, expected.WorkflowTypeName, actual.WorkflowTypeName)
	assert.True(t, (*expected.WorkflowTimeout-*actual.WorkflowTimeout) < time.Second)
	assert.True(t, (*expected.DecisionTaskTimeout-*actual.DecisionTaskTimeout) < time.Second)
	assert.Equal(t, expected.ExecutionContext, actual.ExecutionContext)
	assert.Equal(t, expected.State, actual.State)
	assert.Equal(t, expected.CloseStatus, actual.CloseStatus)
	assert.Equal(t, expected.StartVersion, actual.StartVersion)
	assert.Equal(t, expected.LastWriteEventID, actual.LastWriteEventID)
	assert.Equal(t, expected.LastEventTaskID, actual.LastEventTaskID)
	assert.Equal(t, expected.LastFirstEventID, actual.LastFirstEventID)
	assert.Equal(t, expected.LastProcessedEvent, actual.LastProcessedEvent)
	assert.Equal(t, expected.StartTimestamp.Sub(*actual.StartTimestamp), time.Duration(0))
	assert.Equal(t, expected.LastUpdatedTimestamp.Sub(*actual.LastUpdatedTimestamp), time.Duration(0))
	assert.Equal(t, expected.DecisionVersion, actual.DecisionVersion)
	assert.Equal(t, expected.DecisionScheduleID, actual.DecisionScheduleID)
	assert.Equal(t, expected.DecisionStartedID, actual.DecisionStartedID)
	assert.True(t, (*expected.DecisionTimeout-*actual.DecisionTimeout) < time.Second)
	assert.Equal(t, expected.DecisionAttempt, actual.DecisionAttempt)
	assert.Equal(t, expected.DecisionStartedTimestamp.Sub(*actual.DecisionStartedTimestamp), time.Duration(0))
	assert.Equal(t, expected.DecisionScheduledTimestamp.Sub(*actual.DecisionScheduledTimestamp), time.Duration(0))
	assert.Equal(t, expected.DecisionOriginalScheduledTimestamp.Sub(*actual.DecisionOriginalScheduledTimestamp), time.Duration(0))
	assert.Equal(t, expected.CancelRequested, actual.CancelRequested)
	assert.Equal(t, expected.DecisionRequestID, actual.DecisionRequestID)
	assert.Equal(t, expected.CancelRequestID, actual.CancelRequestID)
	assert.Equal(t, expected.StickyTaskList, actual.StickyTaskList)
	assert.Equal(t, expected.RetryAttempt, actual.RetryAttempt)
	assert.Equal(t, expected.RetryMaximumAttempts, actual.RetryMaximumAttempts)
	assert.Equal(t, expected.RetryBackoffCoefficient, actual.RetryBackoffCoefficient)
	assert.Equal(t, expected.RetryNonRetryableErrors, actual.RetryNonRetryableErrors)
	assert.Equal(t, expected.HasRetryPolicy, actual.HasRetryPolicy)
	assert.Equal(t, expected.CronSchedule, actual.CronSchedule)
	assert.Equal(t, expected.EventStoreVersion, actual.EventStoreVersion)
	assert.Equal(t, expected.EventBranchToken, actual.EventBranchToken)
	assert.Equal(t, expected.SignalCount, actual.SignalCount)
	assert.Equal(t, expected.HistorySize, actual.HistorySize)
	assert.Equal(t, expected.ClientLibraryVersion, actual.ClientLibraryVersion)
	assert.Equal(t, expected.ClientFeatureVersion, actual.ClientFeatureVersion)
	assert.Equal(t, expected.ClientImpl, actual.ClientImpl)
	assert.Equal(t, expected.AutoResetPoints, actual.AutoResetPoints)
	assert.Equal(t, expected.AutoResetPointsEncoding, actual.AutoResetPointsEncoding)
	assert.Equal(t, expected.SearchAttributes, actual.SearchAttributes)
	assert.Equal(t, expected.Memo, actual.Memo)
	assert.Equal(t, expected.VersionHistories, actual.VersionHistories)
	assert.Equal(t, expected.VersionHistoriesEncoding, actual.VersionHistoriesEncoding)
	assert.Equal(t, expected.RetryExpirationTimestamp.Sub(*actual.RetryExpirationTimestamp), time.Duration(0))
	assert.True(t, (*expected.StickyScheduleToStartTimeout-*actual.StickyScheduleToStartTimeout) < time.Second)
	assert.True(t, (*expected.RetryInitialInterval-*actual.RetryInitialInterval) < time.Second)
	assert.True(t, (*expected.RetryMaximumInterval-*actual.RetryMaximumInterval) < time.Second)
	assert.True(t, (*expected.RetryExpiration-*actual.RetryExpiration) < time.Second)
}

func TestActivityInfo(t *testing.T) {
	expected := &ActivityInfo{
		Version:                  common.Int64Ptr(int64(rand.Intn(1000))),
		ScheduledEventBatchID:    common.Int64Ptr(int64(rand.Intn(1000))),
		ScheduledEvent:           []byte("ScheduledEvent"),
		ScheduledEventEncoding:   common.StringPtr("ScheduledEventEncoding"),
		ScheduledTimestamp:       common.TimePtr(time.Now()),
		StartedID:                common.Int64Ptr(int64(rand.Intn(1000))),
		StartedEvent:             []byte("StartedEvent"),
		StartedEventEncoding:     common.StringPtr("StartedEventEncoding"),
		StartedTimestamp:         common.TimePtr(time.Now()),
		ActivityID:               common.StringPtr("ActivityID"),
		RequestID:                common.StringPtr("RequestID"),
		ScheduleToStartTimeout:   common.DurationPtr(time.Minute * time.Duration(rand.Intn(10))),
		ScheduleToCloseTimeout:   common.DurationPtr(time.Minute * time.Duration(rand.Intn(10))),
		StartToCloseTimeout:      common.DurationPtr(time.Minute * time.Duration(rand.Intn(10))),
		HeartbeatTimeout:         common.DurationPtr(time.Minute * time.Duration(rand.Intn(10))),
		CancelRequested:          common.BoolPtr(true),
		CancelRequestID:          common.Int64Ptr(int64(rand.Intn(1000))),
		TimerTaskStatus:          common.Int32Ptr(int32(rand.Intn(1000))),
		Attempt:                  common.Int32Ptr(int32(rand.Intn(1000))),
		TaskList:                 common.StringPtr("TaskList"),
		StartedIdentity:          common.StringPtr("StartedIdentity"),
		HasRetryPolicy:           common.BoolPtr(true),
		RetryInitialInterval:     common.DurationPtr(time.Minute * time.Duration(rand.Intn(10))),
		RetryMaximumInterval:     common.DurationPtr(time.Minute * time.Duration(rand.Intn(10))),
		RetryMaximumAttempts:     common.Int32Ptr(int32(rand.Intn(1000))),
		RetryExpirationTimestamp: common.TimePtr(time.Time{}),
		RetryBackoffCoefficient:  common.Float64Ptr(rand.Float64() * 1000),
		RetryNonRetryableErrors:  []string{"RetryNonRetryableErrors"},
		RetryLastFailureReason:   common.StringPtr("RetryLastFailureReason"),
		RetryLastWorkerIdentity:  common.StringPtr("RetryLastWorkerIdentity"),
		RetryLastFailureDetails:  []byte("RetryLastFailureDetails"),
	}
	actual := activityInfoFromThrift(activityInfoToThrift(expected))
	assert.Equal(t, expected.Version, actual.Version)
	assert.Equal(t, expected.ScheduledEventBatchID, actual.ScheduledEventBatchID)
	assert.Equal(t, expected.ScheduledEvent, actual.ScheduledEvent)
	assert.Equal(t, expected.ScheduledEventEncoding, actual.ScheduledEventEncoding)
	assert.Equal(t, expected.StartedID, actual.StartedID)
	assert.Equal(t, expected.StartedEvent, actual.StartedEvent)
	assert.Equal(t, expected.StartedEventEncoding, actual.StartedEventEncoding)
	assert.Equal(t, expected.ActivityID, actual.ActivityID)
	assert.Equal(t, expected.RequestID, actual.RequestID)
	assert.Equal(t, expected.CancelRequested, actual.CancelRequested)
	assert.Equal(t, expected.CancelRequestID, actual.CancelRequestID)
	assert.Equal(t, expected.TimerTaskStatus, actual.TimerTaskStatus)
	assert.Equal(t, expected.Attempt, actual.Attempt)
	assert.Equal(t, expected.TaskList, actual.TaskList)
	assert.Equal(t, expected.StartedIdentity, actual.StartedIdentity)
	assert.Equal(t, expected.HasRetryPolicy, actual.HasRetryPolicy)
	assert.Equal(t, expected.RetryMaximumAttempts, actual.RetryMaximumAttempts)
	assert.Equal(t, expected.RetryBackoffCoefficient, actual.RetryBackoffCoefficient)
	assert.Equal(t, expected.RetryNonRetryableErrors, actual.RetryNonRetryableErrors)
	assert.Equal(t, expected.RetryLastFailureReason, actual.RetryLastFailureReason)
	assert.Equal(t, expected.RetryLastWorkerIdentity, actual.RetryLastWorkerIdentity)
	assert.Equal(t, expected.RetryLastFailureDetails, actual.RetryLastFailureDetails)
	assert.True(t, (*expected.ScheduleToStartTimeout-*actual.ScheduleToStartTimeout) < time.Second)
	assert.True(t, (*expected.ScheduleToCloseTimeout-*actual.ScheduleToCloseTimeout) < time.Second)
	assert.True(t, (*expected.StartToCloseTimeout-*actual.StartToCloseTimeout) < time.Second)
	assert.True(t, (*expected.HeartbeatTimeout-*actual.HeartbeatTimeout) < time.Second)
	assert.True(t, (*expected.RetryInitialInterval-*actual.RetryInitialInterval) < time.Second)
	assert.True(t, (*expected.RetryMaximumInterval-*actual.RetryMaximumInterval) < time.Second)
	assert.Equal(t, expected.ScheduledTimestamp.Sub(*actual.ScheduledTimestamp), time.Duration(0))
	assert.Equal(t, expected.StartedTimestamp.Sub(*actual.StartedTimestamp), time.Duration(0))
	assert.Equal(t, expected.RetryExpirationTimestamp.Sub(*actual.RetryExpirationTimestamp), time.Duration(0))
}

func TestChildExecutionInfo(t *testing.T) {
	expected := &ChildExecutionInfo{
		Version:                common.Int64Ptr(int64(rand.Intn(1000))),
		InitiatedEventBatchID:  common.Int64Ptr(int64(rand.Intn(1000))),
		StartedID:              common.Int64Ptr(int64(rand.Intn(1000))),
		InitiatedEvent:         []byte("InitiatedEvent"),
		InitiatedEventEncoding: common.StringPtr("InitiatedEventEncoding"),
		StartedWorkflowID:      common.StringPtr("InitiatedEventEncoding"),
		StartedRunID:           UUID(uuid.New()),
		StartedEvent:           []byte("StartedEvent"),
		StartedEventEncoding:   common.StringPtr("StartedEventEncoding"),
		CreateRequestID:        common.StringPtr("CreateRequestID"),
		DomainName:             common.StringPtr("DomainName"),
		WorkflowTypeName:       common.StringPtr("WorkflowTypeName"),
		ParentClosePolicy:      common.Int32Ptr(int32(rand.Intn(1000))),
	}
	actual := childExecutionInfoFromThrift(childExecutionInfoToThrift(expected))
	assert.Equal(t, expected, actual)
}

func TestSignalInfo(t *testing.T) {
	expected := &SignalInfo{
		Version:               common.Int64Ptr(int64(rand.Intn(1000))),
		InitiatedEventBatchID: common.Int64Ptr(int64(rand.Intn(1000))),
		RequestID:             common.StringPtr("RequestID"),
		Name:                  common.StringPtr("Name"),
		Input:                 []byte("Input"),
		Control:               []byte("Control"),
	}
	actual := signalInfoFromThrift(signalInfoToThrift(expected))
	assert.Equal(t, expected, actual)
}

func TestRequestCancelInfo(t *testing.T) {
	expected := &RequestCancelInfo{
		Version:               common.Int64Ptr(int64(rand.Intn(1000))),
		InitiatedEventBatchID: common.Int64Ptr(int64(rand.Intn(1000))),
		CancelRequestID:       common.StringPtr("CancelRequestID"),
	}
	actual := requestCancelInfoFromThrift(requestCancelInfoToThrift(expected))
	assert.Equal(t, expected, actual)
}

func TestTimerInfo(t *testing.T) {
	expected := &TimerInfo{
		Version:         common.Int64Ptr(int64(rand.Intn(1000))),
		StartedID:       common.Int64Ptr(int64(rand.Intn(1000))),
		ExpiryTimestamp: common.TimePtr(time.Now()),
		TaskID:          common.Int64Ptr(int64(rand.Intn(1000))),
	}
	actual := timerInfoFromThrift(timerInfoToThrift(expected))
	assert.Equal(t, expected.Version, actual.Version)
	assert.Equal(t, expected.StartedID, actual.StartedID)
	assert.Equal(t, expected.TaskID, actual.TaskID)
	assert.Equal(t, expected.ExpiryTimestamp.Sub(*actual.ExpiryTimestamp), time.Duration(0))
}

func TestTaskInfo(t *testing.T) {
	expected := &TaskInfo{
		WorkflowID:       common.StringPtr("WorkflowID"),
		RunID:            UUID(uuid.New()),
		ScheduleID:       common.Int64Ptr(int64(rand.Intn(1000))),
		ExpiryTimestamp:  common.TimePtr(time.Now()),
		CreatedTimestamp: common.TimePtr(time.Now()),
	}
	actual := taskInfoFromThrift(taskInfoToThrift(expected))
	assert.Equal(t, expected.WorkflowID, actual.WorkflowID)
	assert.Equal(t, expected.RunID, actual.RunID)
	assert.Equal(t, expected.ScheduleID, actual.ScheduleID)
	assert.Equal(t, expected.ExpiryTimestamp.Sub(*actual.ExpiryTimestamp), time.Duration(0))
	assert.Equal(t, expected.CreatedTimestamp.Sub(*actual.CreatedTimestamp), time.Duration(0))
}

func TestTaskListInfo(t *testing.T) {
	expected := &TaskListInfo{
		Kind:            common.Int16Ptr(int16(rand.Intn(1000))),
		AckLevel:        common.Int64Ptr(int64(rand.Intn(1000))),
		ExpiryTimestamp: common.TimePtr(time.Now()),
		LastUpdated:     common.TimePtr(time.Now()),
	}
	actual := taskListInfoFromThrift(taskListInfoToThrift(expected))
	assert.Equal(t, expected.Kind, actual.Kind)
	assert.Equal(t, expected.AckLevel, actual.AckLevel)
	assert.Equal(t, expected.LastUpdated.Sub(*actual.LastUpdated), time.Duration(0))
	assert.Equal(t, expected.ExpiryTimestamp.Sub(*actual.ExpiryTimestamp), time.Duration(0))
}

func TestTransferTaskInfo(t *testing.T) {
	expected := &TransferTaskInfo{
		DomainID:                UUID(uuid.New()),
		WorkflowID:              common.StringPtr("WorkflowID"),
		RunID:                   UUID(uuid.New()),
		TaskType:                common.Int16Ptr(int16(rand.Intn(1000))),
		TargetDomainID:          UUID(uuid.New()),
		TargetWorkflowID:        common.StringPtr("TargetWorkflowID"),
		TargetRunID:             UUID(uuid.New()),
		TaskList:                common.StringPtr("TaskList"),
		TargetChildWorkflowOnly: common.BoolPtr(true),
		ScheduleID:              common.Int64Ptr(int64(rand.Intn(1000))),
		Version:                 common.Int64Ptr(int64(rand.Intn(1000))),
	}
	actual := transferTaskInfoFromThrift(transferTaskInfoToThrift(expected))
	assert.Equal(t, expected, actual)
}

func TestTimerTaskInfo(t *testing.T) {
	expected := &TimerTaskInfo{
		DomainID:        UUID(uuid.New()),
		WorkflowID:      common.StringPtr("WorkflowID"),
		RunID:           UUID(uuid.New()),
		TaskType:        common.Int16Ptr(int16(rand.Intn(1000))),
		TimeoutType:     common.Int16Ptr(int16(rand.Intn(1000))),
		Version:         common.Int64Ptr(int64(rand.Intn(1000))),
		ScheduleAttempt: common.Int64Ptr(int64(rand.Intn(1000))),
		EventID:         common.Int64Ptr(int64(rand.Intn(1000))),
	}
	actual := timerTaskInfoFromThrift(timerTaskInfoToThrift(expected))
	assert.Equal(t, expected, actual)
}

func TestReplicationTaskInfo(t *testing.T) {
	expected := &ReplicationTaskInfo{
		DomainID:                UUID(uuid.New()),
		WorkflowID:              common.StringPtr("WorkflowID"),
		RunID:                   UUID(uuid.New()),
		TaskType:                common.Int16Ptr(int16(rand.Intn(1000))),
		Version:                 common.Int64Ptr(int64(rand.Intn(1000))),
		FirstEventID:            common.Int64Ptr(int64(rand.Intn(1000))),
		NextEventID:             common.Int64Ptr(int64(rand.Intn(1000))),
		ScheduledID:             common.Int64Ptr(int64(rand.Intn(1000))),
		EventStoreVersion:       common.Int32Ptr(int32(rand.Intn(1000))),
		NewRunEventStoreVersion: common.Int32Ptr(int32(rand.Intn(1000))),
		BranchToken:             []byte("BranchToken"),
		NewRunBranchToken:       []byte("NewRunBranchToken"),
	}
	actual := replicationTaskInfoFromThrift(replicationTaskInfoToThrift(expected))
	assert.Equal(t, expected, actual)
}
