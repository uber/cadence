package serialization

import (
	"time"

	"github.com/uber/cadence/.gen/go/sqlblobs"
	"github.com/uber/cadence/common"

)

func shardInfoToThrift(info *ShardInfo) *sqlblobs.ShardInfo {
	if info == nil {
		return nil
	}
	result := &sqlblobs.ShardInfo{
		StolenSinceRenew: info.StolenSinceRenew,
		ReplicationAckLevel: info.ReplicationAckLevel,
		TransferAckLevel: info.TransferAckLevel,
		DomainNotificationVersion: info.DomainNotificationVersion,
		ClusterTransferAckLevel: info.ClusterTransferAckLevel,
		Owner: info.Owner,
		ClusterReplicationLevel: info.ClusterReplicationLevel,
		PendingFailoverMarkers: info.PendingFailoverMarkers,
		PendingFailoverMarkersEncoding: info.PendingFailoverMarkersEncoding,
		ReplicationDlqAckLevel: info.ReplicationDlqAckLevel,
		TransferProcessingQueueStates: info.TransferProcessingQueueStates,
		TransferProcessingQueueStatesEncoding: info.TransferProcessingQueueStatesEncoding,
		TimerProcessingQueueStates: info.TimerProcessingQueueStates,
		TimerProcessingQueueStatesEncoding: info.TimerProcessingQueueStatesEncoding,
		UpdatedAtNanos: unixNanoPtr(info.UpdatedAt),
		TimerAckLevelNanos: unixNanoPtr(info.TimerAckLevel),
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
		StolenSinceRenew: info.StolenSinceRenew,
		ReplicationAckLevel: info.ReplicationAckLevel,
		TransferAckLevel: info.TransferAckLevel,
		DomainNotificationVersion: info.DomainNotificationVersion,
		ClusterTransferAckLevel: info.ClusterTransferAckLevel,
		Owner: info.Owner,
		ClusterReplicationLevel: info.ClusterReplicationLevel,
		PendingFailoverMarkers: info.PendingFailoverMarkers,
		PendingFailoverMarkersEncoding: info.PendingFailoverMarkersEncoding,
		ReplicationDlqAckLevel: info.ReplicationDlqAckLevel,
		TransferProcessingQueueStates: info.TransferProcessingQueueStates,
		TransferProcessingQueueStatesEncoding: info.TransferProcessingQueueStatesEncoding,
		TimerProcessingQueueStates: info.TimerProcessingQueueStates,
		TimerProcessingQueueStatesEncoding: info.TimerProcessingQueueStatesEncoding,
		UpdatedAt: timePtr(info.UpdatedAtNanos),
		TimerAckLevel: timePtr(info.TimerAckLevelNanos),
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
		Name: info.Name,
		Description: info.Description,
		Owner: info.Owner,
		Status: info.Status,
		EmitMetric: info.EmitMetric,
		ArchivalBucket: info.ArchivalBucket,
		ArchivalStatus: info.ArchivalStatus,
		ConfigVersion: info.ConfigVersion,
		NotificationVersion: info.NotificationVersion,
		FailoverNotificationVersion: info.FailoverNotificationVersion,
		FailoverVersion: info.FailoverVersion,
		ActiveClusterName: info.ActiveClusterName,
		Clusters: info.Clusters,
		Data: info.Data,
		BadBinaries: info.BadBinaries,
		BadBinariesEncoding: info.BadBinariesEncoding,
		HistoryArchivalStatus: info.HistoryArchivalStatus,
		HistoryArchivalURI: info.HistoryArchivalURI,
		VisibilityArchivalStatus: info.VisibilityArchivalStatus,
		VisibilityArchivalURI: info.VisibilityArchivalURI,
		PreviousFailoverVersion: info.PreviousFailoverVersion,
		RetentionDays: durationToDays(info.RetentionDays),
		FailoverEndTime: unixNanoPtr(info.FailoverEndTime),
		LastUpdatedTime: unixNanoPtr(info.LastUpdatedTime),
	}
}

func domainInfoFromThrift(info *sqlblobs.DomainInfo) *DomainInfo {
	if info == nil {
		return nil
	}
	return &DomainInfo{
		Name: info.Name,
		Description: info.Description,
		Owner: info.Owner,
		Status: info.Status,
		EmitMetric: info.EmitMetric,
		ArchivalBucket: info.ArchivalBucket,
		ArchivalStatus: info.ArchivalStatus,
		ConfigVersion: info.ConfigVersion,
		NotificationVersion: info.NotificationVersion,
		FailoverNotificationVersion: info.FailoverNotificationVersion,
		FailoverVersion: info.FailoverVersion,
		ActiveClusterName: info.ActiveClusterName,
		Clusters: info.Clusters,
		Data: info.Data,
		BadBinaries: info.BadBinaries,
		BadBinariesEncoding: info.BadBinariesEncoding,
		HistoryArchivalStatus: info.HistoryArchivalStatus,
		HistoryArchivalURI: info.HistoryArchivalURI,
		VisibilityArchivalStatus: info.VisibilityArchivalStatus,
		VisibilityArchivalURI: info.VisibilityArchivalURI,
		PreviousFailoverVersion: info.PreviousFailoverVersion,
		RetentionDays: daysToDuration(info.RetentionDays),
		FailoverEndTime: timePtr(info.FailoverEndTime),
		LastUpdatedTime: timePtr(info.LastUpdatedTime),
	}
}

func historyTreeInfoToThrift(info *HistoryTreeInfo) *sqlblobs.HistoryTreeInfo {
	if info == nil {
		return nil
	}
}

func historyTreeInfoFromThrift(info *sqlblobs.HistoryTreeInfo) *HistoryTreeInfo {
	if info == nil {
		return nil
	}
}

func workflowExecutionInfoToThrift(info *WorkflowExecutionInfo) *sqlblobs.WorkflowExecutionInfo {
	if info == nil {
		return nil
	}
}

func workflowExecutionInfoFromThrift(info *sqlblobs.WorkflowExecutionInfo) *WorkflowExecutionInfo {
	if info == nil {
		return nil
	}
}

func activityInfoToThrift(info *ActivityInfo) *sqlblobs.ActivityInfo {
	if info == nil {
		return nil
	}
}

func activityInfoFromThrift(info *sqlblobs.ActivityInfo) *ActivityInfo {
	if info == nil {
		return nil
	}
}

func childExecutionInfoToThrift(info *ChildExecutionInfo) *sqlblobs.ChildExecutionInfo {
	if info == nil {
		return nil
	}
}

func childExecutionInfoFromThrift(info *sqlblobs.ChildExecutionInfo) *ChildExecutionInfo {
	if info == nil {
		return nil
	}
}

func signalInfoToThrift(info *SignalInfo) *sqlblobs.SignalInfo {
	if info == nil {
		return nil
	}
}

func signalInfoFromThrift(info *sqlblobs.SignalInfo) *SignalInfo {
	if info == nil {
		return nil
	}
}

func requestCancelInfoToThrift(info *RequestCancelInfo) *sqlblobs.RequestCancelInfo {
	if info == nil {
		return nil
	}
}

func requestCancelInfoFromThrift(info *sqlblobs.RequestCancelInfo) *RequestCancelInfo {
	if info == nil {
		return nil
	}
}

func timerInfoToThrift(info *TimerInfo) *sqlblobs.TimerInfo {
	if info == nil {
		return nil
	}
}

func timerInfoFromThrift(info *sqlblobs.TimerInfo) *TimerInfo {
	if info == nil {
		return nil
	}
}

func taskInfoToThrift(info *TaskInfo) *sqlblobs.TaskListInfo {
	if info == nil {
		return nil
	}
}

func taskInfoFromThrift(info *sqlblobs.TaskInfo) *TaskInfo {
	if info == nil {
		return nil
	}
}

func taskListInfoToThrift(info *TaskListInfo) sqlblobs.TaskListInfo {
	if info == nil {
		return nil
	}
}

func taskListInfoFromThrift(info *sqlblobs.TaskListInfo) *TaskListInfo {
	if info == nil {
		return nil
	}
}

func transferTaskInfoToThrift(info *TransferTaskInfo) *sqlblobs.TransferTaskInfo {
	if info == nil {
		return nil
	}
}

func transferTaskInfoFromThrift(info *sqlblobs.TransferTaskInfo) *TransferTaskInfo {
	if info == nil {
		return nil
	}
}

func timerTaskInfoToThrift(info *TimerTaskInfo) *sqlblobs.TimerTaskInfo {
	if info == nil {
		return nil
	}
}

func timerTaskInfoFromThrift(info *sqlblobs.TimerTaskInfo) *TimerTaskInfo {
	if info == nil {
		return nil
	}
}

func replicationTaskInfoToThrift(info *ReplicationTaskInfo) *sqlblobs.ReplicationTaskInfo {
	if info == nil {
		return nil
	}
}

func replicationTaskInfoFromThrift(info *sqlblobs.ReplicationTaskInfo) *ReplicationTaskInfo {
	if info == nil {
		return nil
	}
}

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
	return common.TimePtr(time.Unix(0, *t))
}

func durationToSeconds(t *time.Duration) *int32 {
	if t == nil {
		return nil
	}
	return common.Int32Ptr(int32(common.DurationToSeconds(*t)))
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