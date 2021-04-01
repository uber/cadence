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
	"time"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/.gen/go/sqlblobs"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
)

func shardInfoToThrift(info *ShardInfo) *sqlblobs.ShardInfo {
	if info == nil {
		return nil
	}
	result := &sqlblobs.ShardInfo{
		StolenSinceRenew:                      info.StolenSinceRenew,
		ReplicationAckLevel:                   info.ReplicationAckLevel,
		TransferAckLevel:                      info.TransferAckLevel,
		DomainNotificationVersion:             info.DomainNotificationVersion,
		ClusterTransferAckLevel:               info.ClusterTransferAckLevel,
		Owner:                                 info.Owner,
		ClusterReplicationLevel:               info.ClusterReplicationLevel,
		PendingFailoverMarkers:                info.PendingFailoverMarkers,
		PendingFailoverMarkersEncoding:        info.PendingFailoverMarkersEncoding,
		ReplicationDlqAckLevel:                info.ReplicationDlqAckLevel,
		TransferProcessingQueueStates:         info.TransferProcessingQueueStates,
		TransferProcessingQueueStatesEncoding: info.TransferProcessingQueueStatesEncoding,
		TimerProcessingQueueStates:            info.TimerProcessingQueueStates,
		TimerProcessingQueueStatesEncoding:    info.TimerProcessingQueueStatesEncoding,
		UpdatedAtNanos:                        unixNanoPtr(info.UpdatedAt),
		TimerAckLevelNanos:                    unixNanoPtr(info.TimerAckLevel),
	}
	if info.ClusterTimerAckLevel != nil {
		result.ClusterTimerAckLevel = make(map[string]int64, len(info.ClusterTimerAckLevel))
		for k, v := range info.ClusterTimerAckLevel {
			result.ClusterTimerAckLevel[k] = v.UnixNano()
		}
	}
	return result
}

func shardInfoFromThrift(info *sqlblobs.ShardInfo) *ShardInfo {
	if info == nil {
		return nil
	}

	result := &ShardInfo{
		StolenSinceRenew:                      info.StolenSinceRenew,
		ReplicationAckLevel:                   info.ReplicationAckLevel,
		TransferAckLevel:                      info.TransferAckLevel,
		DomainNotificationVersion:             info.DomainNotificationVersion,
		ClusterTransferAckLevel:               info.ClusterTransferAckLevel,
		Owner:                                 info.Owner,
		ClusterReplicationLevel:               info.ClusterReplicationLevel,
		PendingFailoverMarkers:                info.PendingFailoverMarkers,
		PendingFailoverMarkersEncoding:        info.PendingFailoverMarkersEncoding,
		ReplicationDlqAckLevel:                info.ReplicationDlqAckLevel,
		TransferProcessingQueueStates:         info.TransferProcessingQueueStates,
		TransferProcessingQueueStatesEncoding: info.TransferProcessingQueueStatesEncoding,
		TimerProcessingQueueStates:            info.TimerProcessingQueueStates,
		TimerProcessingQueueStatesEncoding:    info.TimerProcessingQueueStatesEncoding,
		UpdatedAt:                             timePtr(info.UpdatedAtNanos),
		TimerAckLevel:                         timePtr(info.TimerAckLevelNanos),
	}
	if info.ClusterTimerAckLevel != nil {
		result.ClusterTimerAckLevel = make(map[string]time.Time, len(info.ClusterTimerAckLevel))
		for k, v := range info.ClusterTimerAckLevel {
			result.ClusterTimerAckLevel[k] = time.Unix(0, v)
		}
	}
	return result
}

func domainInfoToThrift(info *DomainInfo) *sqlblobs.DomainInfo {
	if info == nil {
		return nil
	}
	return &sqlblobs.DomainInfo{
		Name:                        info.Name,
		Description:                 info.Description,
		Owner:                       info.Owner,
		Status:                      info.Status,
		EmitMetric:                  info.EmitMetric,
		ArchivalBucket:              info.ArchivalBucket,
		ArchivalStatus:              info.ArchivalStatus,
		ConfigVersion:               info.ConfigVersion,
		NotificationVersion:         info.NotificationVersion,
		FailoverNotificationVersion: info.FailoverNotificationVersion,
		FailoverVersion:             info.FailoverVersion,
		ActiveClusterName:           info.ActiveClusterName,
		Clusters:                    info.Clusters,
		Data:                        info.Data,
		BadBinaries:                 info.BadBinaries,
		BadBinariesEncoding:         info.BadBinariesEncoding,
		HistoryArchivalStatus:       info.HistoryArchivalStatus,
		HistoryArchivalURI:          info.HistoryArchivalURI,
		VisibilityArchivalStatus:    info.VisibilityArchivalStatus,
		VisibilityArchivalURI:       info.VisibilityArchivalURI,
		PreviousFailoverVersion:     info.PreviousFailoverVersion,
		RetentionDays:               durationToDays(info.Retention),
		FailoverEndTime:             unixNanoPtr(info.FailoverEndTimestamp),
		LastUpdatedTime:             unixNanoPtr(info.LastUpdatedTimestamp),
	}
}

func domainInfoFromThrift(info *sqlblobs.DomainInfo) *DomainInfo {
	if info == nil {
		return nil
	}
	return &DomainInfo{
		Name:                        info.Name,
		Description:                 info.Description,
		Owner:                       info.Owner,
		Status:                      info.Status,
		EmitMetric:                  info.EmitMetric,
		ArchivalBucket:              info.ArchivalBucket,
		ArchivalStatus:              info.ArchivalStatus,
		ConfigVersion:               info.ConfigVersion,
		NotificationVersion:         info.NotificationVersion,
		FailoverNotificationVersion: info.FailoverNotificationVersion,
		FailoverVersion:             info.FailoverVersion,
		ActiveClusterName:           info.ActiveClusterName,
		Clusters:                    info.Clusters,
		Data:                        info.Data,
		BadBinaries:                 info.BadBinaries,
		BadBinariesEncoding:         info.BadBinariesEncoding,
		HistoryArchivalStatus:       info.HistoryArchivalStatus,
		HistoryArchivalURI:          info.HistoryArchivalURI,
		VisibilityArchivalStatus:    info.VisibilityArchivalStatus,
		VisibilityArchivalURI:       info.VisibilityArchivalURI,
		PreviousFailoverVersion:     info.PreviousFailoverVersion,
		Retention:                   daysToDuration(info.RetentionDays),
		FailoverEndTimestamp:        timePtr(info.FailoverEndTime),
		LastUpdatedTimestamp:        timePtr(info.LastUpdatedTime),
	}
}

func historyTreeInfoToThrift(info *HistoryTreeInfo) *sqlblobs.HistoryTreeInfo {
	if info == nil {
		return nil
	}
	result := &sqlblobs.HistoryTreeInfo{
		CreatedTimeNanos: unixNanoPtr(info.CreatedTimestamp),
		Info:             info.Info,
	}
	if info.Ancestors != nil {
		result.Ancestors = make([]*shared.HistoryBranchRange, len(info.Ancestors), len(info.Ancestors))
		for i, a := range info.Ancestors {
			result.Ancestors[i] = &shared.HistoryBranchRange{
				BranchID:    a.BranchID,
				BeginNodeID: a.BeginNodeID,
				EndNodeID:   a.EndNodeID,
			}
		}
	}
	return result
}

func historyTreeInfoFromThrift(info *sqlblobs.HistoryTreeInfo) *HistoryTreeInfo {
	if info == nil {
		return nil
	}
	result := &HistoryTreeInfo{
		CreatedTimestamp: timePtr(info.CreatedTimeNanos),
		Info:             info.Info,
	}
	if info.Ancestors != nil {
		result.Ancestors = make([]*types.HistoryBranchRange, len(info.Ancestors), len(info.Ancestors))
		for i, a := range info.Ancestors {
			result.Ancestors[i] = &types.HistoryBranchRange{
				BranchID:    a.BranchID,
				BeginNodeID: a.BeginNodeID,
				EndNodeID:   a.EndNodeID,
			}
		}
	}
	return result
}

func workflowExecutionInfoToThrift(info *WorkflowExecutionInfo) *sqlblobs.WorkflowExecutionInfo {
	if info == nil {
		return nil
	}
	return &sqlblobs.WorkflowExecutionInfo{
		ParentDomainID:                          info.ParentDomainID,
		ParentWorkflowID:                        info.ParentWorkflowID,
		ParentRunID:                             info.ParentRunID,
		InitiatedID:                             info.InitiatedID,
		CompletionEventBatchID:                  info.CompletionEventBatchID,
		CompletionEvent:                         info.CompletionEvent,
		CompletionEventEncoding:                 info.CompletionEventEncoding,
		TaskList:                                info.TaskList,
		WorkflowTypeName:                        info.WorkflowTypeName,
		WorkflowTimeoutSeconds:                  durationToSeconds(info.WorkflowTimeout),
		DecisionTaskTimeoutSeconds:              durationToSeconds(info.DecisionTaskTimeout),
		ExecutionContext:                        info.ExecutionContext,
		State:                                   info.State,
		CloseStatus:                             info.CloseStatus,
		StartVersion:                            info.StartVersion,
		LastWriteEventID:                        info.LastWriteEventID,
		LastEventTaskID:                         info.LastEventTaskID,
		LastFirstEventID:                        info.LastFirstEventID,
		LastProcessedEvent:                      info.LastProcessedEvent,
		StartTimeNanos:                          unixNanoPtr(info.StartTimestamp),
		LastUpdatedTimeNanos:                    unixNanoPtr(info.LastUpdatedTimestamp),
		DecisionVersion:                         info.DecisionVersion,
		DecisionScheduleID:                      info.DecisionScheduleID,
		DecisionStartedID:                       info.DecisionStartedID,
		DecisionTimeout:                         durationToSeconds(info.DecisionTimeout),
		DecisionAttempt:                         info.DecisionAttempt,
		DecisionStartedTimestampNanos:           unixNanoPtr(info.DecisionStartedTimestamp),
		DecisionScheduledTimestampNanos:         unixNanoPtr(info.DecisionScheduledTimestamp),
		CancelRequested:                         info.CancelRequested,
		DecisionOriginalScheduledTimestampNanos: unixNanoPtr(info.DecisionOriginalScheduledTimestamp),
		CreateRequestID:                         info.CreateRequestID,
		DecisionRequestID:                       info.DecisionRequestID,
		CancelRequestID:                         info.CancelRequestID,
		StickyTaskList:                          info.StickyTaskList,
		StickyScheduleToStartTimeout:            durationToSecondsInt64(info.StickyScheduleToStartTimeout),
		RetryAttempt:                            info.RetryAttempt,
		RetryInitialIntervalSeconds:             durationToSeconds(info.RetryInitialInterval),
		RetryMaximumIntervalSeconds:             durationToSeconds(info.RetryMaximumInterval),
		RetryMaximumAttempts:                    info.RetryMaximumAttempts,
		RetryExpirationSeconds:                  durationToSeconds(info.RetryExpiration),
		RetryBackoffCoefficient:                 info.RetryBackoffCoefficient,
		RetryExpirationTimeNanos:                unixNanoPtr(info.RetryExpirationTimestamp),
		RetryNonRetryableErrors:                 info.RetryNonRetryableErrors,
		HasRetryPolicy:                          info.HasRetryPolicy,
		CronSchedule:                            info.CronSchedule,
		EventStoreVersion:                       info.EventStoreVersion,
		EventBranchToken:                        info.EventBranchToken,
		SignalCount:                             info.SignalCount,
		HistorySize:                             info.HistorySize,
		ClientLibraryVersion:                    info.ClientLibraryVersion,
		ClientFeatureVersion:                    info.ClientFeatureVersion,
		ClientImpl:                              info.ClientImpl,
		AutoResetPoints:                         info.AutoResetPoints,
		AutoResetPointsEncoding:                 info.AutoResetPointsEncoding,
		SearchAttributes:                        info.SearchAttributes,
		Memo:                                    info.Memo,
		VersionHistories:                        info.VersionHistories,
		VersionHistoriesEncoding:                info.VersionHistoriesEncoding,
	}
}

func workflowExecutionInfoFromThrift(info *sqlblobs.WorkflowExecutionInfo) *WorkflowExecutionInfo {
	if info == nil {
		return nil
	}
	return &WorkflowExecutionInfo{
		ParentDomainID:                     info.ParentDomainID,
		ParentWorkflowID:                   info.ParentWorkflowID,
		ParentRunID:                        info.ParentRunID,
		InitiatedID:                        info.InitiatedID,
		CompletionEventBatchID:             info.CompletionEventBatchID,
		CompletionEvent:                    info.CompletionEvent,
		CompletionEventEncoding:            info.CompletionEventEncoding,
		TaskList:                           info.TaskList,
		WorkflowTypeName:                   info.WorkflowTypeName,
		WorkflowTimeout:                    secondsToDuration(info.WorkflowTimeoutSeconds),
		DecisionTaskTimeout:                secondsToDuration(info.DecisionTaskTimeoutSeconds),
		ExecutionContext:                   info.ExecutionContext,
		State:                              info.State,
		CloseStatus:                        info.CloseStatus,
		StartVersion:                       info.StartVersion,
		LastWriteEventID:                   info.LastWriteEventID,
		LastEventTaskID:                    info.LastEventTaskID,
		LastFirstEventID:                   info.LastFirstEventID,
		LastProcessedEvent:                 info.LastProcessedEvent,
		StartTimestamp:                     timePtr(info.StartTimeNanos),
		LastUpdatedTimestamp:               timePtr(info.LastUpdatedTimeNanos),
		DecisionVersion:                    info.DecisionVersion,
		DecisionScheduleID:                 info.DecisionScheduleID,
		DecisionStartedID:                  info.DecisionStartedID,
		DecisionTimeout:                    secondsToDuration(info.DecisionTimeout),
		DecisionAttempt:                    info.DecisionAttempt,
		DecisionStartedTimestamp:           timePtr(info.DecisionStartedTimestampNanos),
		DecisionScheduledTimestamp:         timePtr(info.DecisionScheduledTimestampNanos),
		CancelRequested:                    info.CancelRequested,
		DecisionOriginalScheduledTimestamp: timePtr(info.DecisionOriginalScheduledTimestampNanos),
		CreateRequestID:                    info.CreateRequestID,
		DecisionRequestID:                  info.DecisionRequestID,
		CancelRequestID:                    info.CancelRequestID,
		StickyTaskList:                     info.StickyTaskList,
		StickyScheduleToStartTimeout:       secondsInt64ToDuration(info.StickyScheduleToStartTimeout),
		RetryAttempt:                       info.RetryAttempt,
		RetryInitialInterval:               secondsToDuration(info.RetryInitialIntervalSeconds),
		RetryMaximumInterval:               secondsToDuration(info.RetryMaximumIntervalSeconds),
		RetryMaximumAttempts:               info.RetryMaximumAttempts,
		RetryExpiration:                    secondsToDuration(info.RetryExpirationSeconds),
		RetryBackoffCoefficient:            info.RetryBackoffCoefficient,
		RetryExpirationTimestamp:           timePtr(info.RetryExpirationTimeNanos),
		RetryNonRetryableErrors:            info.RetryNonRetryableErrors,
		HasRetryPolicy:                     info.HasRetryPolicy,
		CronSchedule:                       info.CronSchedule,
		EventStoreVersion:                  info.EventStoreVersion,
		EventBranchToken:                   info.EventBranchToken,
		SignalCount:                        info.SignalCount,
		HistorySize:                        info.HistorySize,
		ClientLibraryVersion:               info.ClientLibraryVersion,
		ClientFeatureVersion:               info.ClientFeatureVersion,
		ClientImpl:                         info.ClientImpl,
		AutoResetPoints:                    info.AutoResetPoints,
		AutoResetPointsEncoding:            info.AutoResetPointsEncoding,
		SearchAttributes:                   info.SearchAttributes,
		Memo:                               info.Memo,
		VersionHistories:                   info.VersionHistories,
		VersionHistoriesEncoding:           info.VersionHistoriesEncoding,
	}
}

func activityInfoToThrift(info *ActivityInfo) *sqlblobs.ActivityInfo {
	if info == nil {
		return nil
	}
	return &sqlblobs.ActivityInfo{
		Version:                       info.Version,
		ScheduledEventBatchID:         info.ScheduledEventBatchID,
		ScheduledEvent:                info.ScheduledEvent,
		ScheduledEventEncoding:        info.ScheduledEventEncoding,
		ScheduledTimeNanos:            unixNanoPtr(info.ScheduledTimestamp),
		StartedID:                     info.StartedID,
		StartedEvent:                  info.StartedEvent,
		StartedEventEncoding:          info.StartedEventEncoding,
		StartedTimeNanos:              unixNanoPtr(info.StartedTimestamp),
		ActivityID:                    info.ActivityID,
		RequestID:                     info.RequestID,
		ScheduleToStartTimeoutSeconds: durationToSeconds(info.ScheduleToStartTimeout),
		ScheduleToCloseTimeoutSeconds: durationToSeconds(info.ScheduleToCloseTimeout),
		StartToCloseTimeoutSeconds:    durationToSeconds(info.StartToCloseTimeout),
		HeartbeatTimeoutSeconds:       durationToSeconds(info.HeartbeatTimeout),
		CancelRequested:               info.CancelRequested,
		CancelRequestID:               info.CancelRequestID,
		TimerTaskStatus:               info.TimerTaskStatus,
		Attempt:                       info.Attempt,
		TaskList:                      info.TaskList,
		StartedIdentity:               info.StartedIdentity,
		HasRetryPolicy:                info.HasRetryPolicy,
		RetryInitialIntervalSeconds:   durationToSeconds(info.RetryInitialInterval),
		RetryMaximumIntervalSeconds:   durationToSeconds(info.RetryMaximumInterval),
		RetryMaximumAttempts:          info.RetryMaximumAttempts,
		RetryExpirationTimeNanos:      unixNanoPtr(info.RetryExpirationTimestamp),
		RetryBackoffCoefficient:       info.RetryBackoffCoefficient,
		RetryNonRetryableErrors:       info.RetryNonRetryableErrors,
		RetryLastFailureReason:        info.RetryLastFailureReason,
		RetryLastWorkerIdentity:       info.RetryLastWorkerIdentity,
		RetryLastFailureDetails:       info.RetryLastFailureDetails,
	}
}

func activityInfoFromThrift(info *sqlblobs.ActivityInfo) *ActivityInfo {
	if info == nil {
		return nil
	}
	return &ActivityInfo{
		Version:                  info.Version,
		ScheduledEventBatchID:    info.ScheduledEventBatchID,
		ScheduledEvent:           info.ScheduledEvent,
		ScheduledEventEncoding:   info.ScheduledEventEncoding,
		ScheduledTimestamp:       timePtr(info.ScheduledTimeNanos),
		StartedID:                info.StartedID,
		StartedEvent:             info.StartedEvent,
		StartedEventEncoding:     info.StartedEventEncoding,
		StartedTimestamp:         timePtr(info.StartedTimeNanos),
		ActivityID:               info.ActivityID,
		RequestID:                info.RequestID,
		ScheduleToStartTimeout:   secondsToDuration(info.ScheduleToStartTimeoutSeconds),
		ScheduleToCloseTimeout:   secondsToDuration(info.ScheduleToCloseTimeoutSeconds),
		StartToCloseTimeout:      secondsToDuration(info.StartToCloseTimeoutSeconds),
		HeartbeatTimeout:         secondsToDuration(info.HeartbeatTimeoutSeconds),
		CancelRequested:          info.CancelRequested,
		CancelRequestID:          info.CancelRequestID,
		TimerTaskStatus:          info.TimerTaskStatus,
		Attempt:                  info.Attempt,
		TaskList:                 info.TaskList,
		StartedIdentity:          info.StartedIdentity,
		HasRetryPolicy:           info.HasRetryPolicy,
		RetryInitialInterval:     secondsToDuration(info.RetryInitialIntervalSeconds),
		RetryMaximumInterval:     secondsToDuration(info.RetryMaximumIntervalSeconds),
		RetryMaximumAttempts:     info.RetryMaximumAttempts,
		RetryExpirationTimestamp: timePtr(info.RetryExpirationTimeNanos),
		RetryBackoffCoefficient:  info.RetryBackoffCoefficient,
		RetryNonRetryableErrors:  info.RetryNonRetryableErrors,
		RetryLastFailureReason:   info.RetryLastFailureReason,
		RetryLastWorkerIdentity:  info.RetryLastWorkerIdentity,
		RetryLastFailureDetails:  info.RetryLastFailureDetails,
	}
}

func childExecutionInfoToThrift(info *ChildExecutionInfo) *sqlblobs.ChildExecutionInfo {
	if info == nil {
		return nil
	}
	return &sqlblobs.ChildExecutionInfo{
		Version:                info.Version,
		InitiatedEventBatchID:  info.InitiatedEventBatchID,
		StartedID:              info.StartedID,
		InitiatedEvent:         info.InitiatedEvent,
		InitiatedEventEncoding: info.InitiatedEventEncoding,
		StartedWorkflowID:      info.StartedWorkflowID,
		StartedRunID:           info.StartedRunID,
		StartedEvent:           info.StartedEvent,
		StartedEventEncoding:   info.StartedEventEncoding,
		CreateRequestID:        info.CreateRequestID,
		DomainName:             info.DomainName,
		WorkflowTypeName:       info.WorkflowTypeName,
		ParentClosePolicy:      info.ParentClosePolicy,
	}
}

func childExecutionInfoFromThrift(info *sqlblobs.ChildExecutionInfo) *ChildExecutionInfo {
	if info == nil {
		return nil
	}
	return &ChildExecutionInfo{
		Version:                info.Version,
		InitiatedEventBatchID:  info.InitiatedEventBatchID,
		StartedID:              info.StartedID,
		InitiatedEvent:         info.InitiatedEvent,
		InitiatedEventEncoding: info.InitiatedEventEncoding,
		StartedWorkflowID:      info.StartedWorkflowID,
		StartedRunID:           info.StartedRunID,
		StartedEvent:           info.StartedEvent,
		StartedEventEncoding:   info.StartedEventEncoding,
		CreateRequestID:        info.CreateRequestID,
		DomainName:             info.DomainName,
		WorkflowTypeName:       info.WorkflowTypeName,
		ParentClosePolicy:      info.ParentClosePolicy,
	}
}

func signalInfoToThrift(info *SignalInfo) *sqlblobs.SignalInfo {
	if info == nil {
		return nil
	}
	return &sqlblobs.SignalInfo{
		Version:               info.Version,
		InitiatedEventBatchID: info.InitiatedEventBatchID,
		RequestID:             info.RequestID,
		Name:                  info.Name,
		Input:                 info.Input,
		Control:               info.Control,
	}
}

func signalInfoFromThrift(info *sqlblobs.SignalInfo) *SignalInfo {
	if info == nil {
		return nil
	}
	return &SignalInfo{
		Version:               info.Version,
		InitiatedEventBatchID: info.InitiatedEventBatchID,
		RequestID:             info.RequestID,
		Name:                  info.Name,
		Input:                 info.Input,
		Control:               info.Control,
	}
}

func requestCancelInfoToThrift(info *RequestCancelInfo) *sqlblobs.RequestCancelInfo {
	if info == nil {
		return nil
	}
	return &sqlblobs.RequestCancelInfo{
		Version:               info.Version,
		InitiatedEventBatchID: info.InitiatedEventBatchID,
		CancelRequestID:       info.CancelRequestID,
	}
}

func requestCancelInfoFromThrift(info *sqlblobs.RequestCancelInfo) *RequestCancelInfo {
	if info == nil {
		return nil
	}
	return &RequestCancelInfo{
		Version:               info.Version,
		InitiatedEventBatchID: info.InitiatedEventBatchID,
		CancelRequestID:       info.CancelRequestID,
	}
}

func timerInfoToThrift(info *TimerInfo) *sqlblobs.TimerInfo {
	if info == nil {
		return nil
	}
	return &sqlblobs.TimerInfo{
		Version:         info.Version,
		StartedID:       info.StartedID,
		ExpiryTimeNanos: unixNanoPtr(info.ExpiryTimestamp),
		TaskID:          info.TaskID,
	}
}

func timerInfoFromThrift(info *sqlblobs.TimerInfo) *TimerInfo {
	if info == nil {
		return nil
	}
	return &TimerInfo{
		Version:         info.Version,
		StartedID:       info.StartedID,
		ExpiryTimestamp: timePtr(info.ExpiryTimeNanos),
		TaskID:          info.TaskID,
	}
}

func taskInfoToThrift(info *TaskInfo) *sqlblobs.TaskInfo {
	if info == nil {
		return nil
	}
	return &sqlblobs.TaskInfo{
		WorkflowID:       info.WorkflowID,
		RunID:            info.RunID,
		ScheduleID:       info.ScheduleID,
		ExpiryTimeNanos:  unixNanoPtr(info.ExpiryTimestamp),
		CreatedTimeNanos: unixNanoPtr(info.CreatedTimestamp),
	}
}

func taskInfoFromThrift(info *sqlblobs.TaskInfo) *TaskInfo {
	if info == nil {
		return nil
	}
	return &TaskInfo{
		WorkflowID:       info.WorkflowID,
		RunID:            info.RunID,
		ScheduleID:       info.ScheduleID,
		ExpiryTimestamp:  timePtr(info.ExpiryTimeNanos),
		CreatedTimestamp: timePtr(info.CreatedTimeNanos),
	}
}

func taskListInfoToThrift(info *TaskListInfo) *sqlblobs.TaskListInfo {
	if info == nil {
		return nil
	}
	return &sqlblobs.TaskListInfo{
		Kind:             info.Kind,
		AckLevel:         info.AckLevel,
		ExpiryTimeNanos:  unixNanoPtr(info.ExpiryTimestamp),
		LastUpdatedNanos: unixNanoPtr(info.LastUpdated),
	}
}

func taskListInfoFromThrift(info *sqlblobs.TaskListInfo) *TaskListInfo {
	if info == nil {
		return nil
	}
	return &TaskListInfo{
		Kind:            info.Kind,
		AckLevel:        info.AckLevel,
		ExpiryTimestamp: timePtr(info.ExpiryTimeNanos),
		LastUpdated:     timePtr(info.LastUpdatedNanos),
	}
}

func transferTaskInfoToThrift(info *TransferTaskInfo) *sqlblobs.TransferTaskInfo {
	if info == nil {
		return nil
	}
	return &sqlblobs.TransferTaskInfo{
		DomainID:                 info.DomainID,
		WorkflowID:               info.WorkflowID,
		RunID:                    info.RunID,
		TaskType:                 info.TaskType,
		TargetDomainID:           info.TargetDomainID,
		TargetWorkflowID:         info.TargetWorkflowID,
		TargetRunID:              info.TargetRunID,
		TaskList:                 info.TaskList,
		TargetChildWorkflowOnly:  info.TargetChildWorkflowOnly,
		ScheduleID:               info.ScheduleID,
		Version:                  info.Version,
		VisibilityTimestampNanos: unixNanoPtr(info.VisibilityTimestamp),
	}
}

func transferTaskInfoFromThrift(info *sqlblobs.TransferTaskInfo) *TransferTaskInfo {
	if info == nil {
		return nil
	}
	return &TransferTaskInfo{
		DomainID:                info.DomainID,
		WorkflowID:              info.WorkflowID,
		RunID:                   info.RunID,
		TaskType:                info.TaskType,
		TargetDomainID:          info.TargetDomainID,
		TargetWorkflowID:        info.TargetWorkflowID,
		TargetRunID:             info.TargetRunID,
		TaskList:                info.TaskList,
		TargetChildWorkflowOnly: info.TargetChildWorkflowOnly,
		ScheduleID:              info.ScheduleID,
		Version:                 info.Version,
		VisibilityTimestamp:     timePtr(info.VisibilityTimestampNanos),
	}
}

func timerTaskInfoToThrift(info *TimerTaskInfo) *sqlblobs.TimerTaskInfo {
	if info == nil {
		return nil
	}
	return &sqlblobs.TimerTaskInfo{
		DomainID:        info.DomainID,
		WorkflowID:      info.WorkflowID,
		RunID:           info.RunID,
		TaskType:        info.TaskType,
		TimeoutType:     info.TimeoutType,
		Version:         info.Version,
		ScheduleAttempt: info.ScheduleAttempt,
		EventID:         info.EventID,
	}
}

func timerTaskInfoFromThrift(info *sqlblobs.TimerTaskInfo) *TimerTaskInfo {
	if info == nil {
		return nil
	}
	return &TimerTaskInfo{
		DomainID:        info.DomainID,
		WorkflowID:      info.WorkflowID,
		RunID:           info.RunID,
		TaskType:        info.TaskType,
		TimeoutType:     info.TimeoutType,
		Version:         info.Version,
		ScheduleAttempt: info.ScheduleAttempt,
		EventID:         info.EventID,
	}
}

func replicationTaskInfoToThrift(info *ReplicationTaskInfo) *sqlblobs.ReplicationTaskInfo {
	if info == nil {
		return nil
	}
	return &sqlblobs.ReplicationTaskInfo{
		DomainID:                info.DomainID,
		WorkflowID:              info.WorkflowID,
		RunID:                   info.RunID,
		TaskType:                info.TaskType,
		Version:                 info.Version,
		FirstEventID:            info.FirstEventID,
		NextEventID:             info.NextEventID,
		ScheduledID:             info.ScheduledID,
		EventStoreVersion:       info.EventStoreVersion,
		NewRunEventStoreVersion: info.NewRunEventStoreVersion,
		BranchToken:             info.BranchToken,
		NewRunBranchToken:       info.NewRunBranchToken,
		CreationTime:            unixNanoPtr(info.CreationTimestamp),
	}
}

func replicationTaskInfoFromThrift(info *sqlblobs.ReplicationTaskInfo) *ReplicationTaskInfo {
	if info == nil {
		return nil
	}
	return &ReplicationTaskInfo{
		DomainID:                info.DomainID,
		WorkflowID:              info.WorkflowID,
		RunID:                   info.RunID,
		TaskType:                info.TaskType,
		Version:                 info.Version,
		FirstEventID:            info.FirstEventID,
		NextEventID:             info.NextEventID,
		ScheduledID:             info.ScheduledID,
		EventStoreVersion:       info.EventStoreVersion,
		NewRunEventStoreVersion: info.NewRunEventStoreVersion,
		BranchToken:             info.BranchToken,
		NewRunBranchToken:       info.NewRunBranchToken,
		CreationTimestamp:       timePtr(info.CreationTime),
	}
}

var zeroTimeNanos = time.Time{}.UnixNano()

func unixNanoPtr(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	return common.Int64Ptr(t.UnixNano())
}

func timePtr(t *int64) *time.Time {
	if t == nil {
		return nil
	}
	// Calling UnixNano() on zero time is undefined an results in a number that is converted back to 1754-08-30T22:43:41Z
	// Handle such case explicitly
	if *t == zeroTimeNanos {
		return common.TimePtr(time.Time{})
	}
	return common.TimePtr(time.Unix(0, *t))
}

func durationToSeconds(t *time.Duration) *int32 {
	if t == nil {
		return nil
	}
	return common.Int32Ptr(int32(common.DurationToSeconds(*t)))
}

func durationToSecondsInt64(t *time.Duration) *int64 {
	if t == nil {
		return nil
	}
	return common.Int64Ptr(common.DurationToSeconds(*t))
}

func secondsInt64ToDuration(t *int64) *time.Duration {
	if t == nil {
		return nil
	}
	return common.DurationPtr(common.SecondsToDuration(*t))
}

func secondsToDuration(t *int32) *time.Duration {
	if t == nil {
		return nil
	}
	return common.DurationPtr(common.SecondsToDuration(int64(*t)))
}

func durationToDays(t *time.Duration) *int16 {
	if t == nil {
		return nil
	}
	return common.Int16Ptr(int16(common.DurationToDays(*t)))
}

func daysToDuration(t *int16) *time.Duration {
	if t == nil {
		return nil
	}
	return common.DurationPtr(common.DaysToDuration(int32(*t)))
}
