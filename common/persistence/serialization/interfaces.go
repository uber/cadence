// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
//
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

	"go.uber.org/thriftrw/wire"

	"github.com/uber/cadence/.gen/go/sqlblobs"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
)

type (
	// ShardInfo blob in a serialization agnostic format
	ShardInfo struct {
		StolenSinceRenew                      *int32
		UpdatedAt                             *time.Time
		ReplicationAckLevel                   *int64
		TransferAckLevel                      *int64
		TimerAckLevel                         *time.Time
		DomainNotificationVersion             *int64
		ClusterTransferAckLevel               map[string]int64
		ClusterTimerAckLevel                  map[string]time.Time
		Owner                                 *string
		ClusterReplicationLevel               map[string]int64
		PendingFailoverMarkers                []byte
		PendingFailoverMarkersEncoding        *string
		ReplicationDlqAckLevel                map[string]int64
		TransferProcessingQueueStates         []byte
		TransferProcessingQueueStatesEncoding *string
		TimerProcessingQueueStates            []byte
		TimerProcessingQueueStatesEncoding    *string
	}

	// DomainInfo blob in a serialization agnostic format
	DomainInfo struct {
		Name                        *string
		Description                 *string
		Owner                       *string
		Status                      *int32
		Retention                   *time.Duration
		EmitMetric                  *bool
		ArchivalBucket              *string
		ArchivalStatus              *int16
		ConfigVersion               *int64
		NotificationVersion         *int64
		FailoverNotificationVersion *int64
		FailoverVersion             *int64
		ActiveClusterName           *string
		Clusters                    []string
		Data                        map[string]string
		BadBinaries                 []byte
		BadBinariesEncoding         *string
		HistoryArchivalStatus       *int16
		HistoryArchivalURI          *string
		VisibilityArchivalStatus    *int16
		VisibilityArchivalURI       *string
		FailoverEndTimestamp        *time.Time
		PreviousFailoverVersion     *int64
		LastUpdatedTimestamp        *time.Time
	}

	// HistoryBranchRange blob in a serialization agnostic format
	HistoryBranchRange struct {
		BranchID    *string
		BeginNodeID *int64
		EndNodeID   *int64
	}

	// HistoryTreeInfo blob in a serialization agnostic format
	HistoryTreeInfo struct {
		CreatedTimestamp *time.Time
		Ancestors        []*HistoryBranchRange
		Info             *string
	}

	// WorkflowExecutionInfo blob in a serialization agnostic format
	WorkflowExecutionInfo struct {
		ParentDomainID                     UUID
		ParentWorkflowID                   *string
		ParentRunID                        UUID
		InitiatedID                        *int64
		CompletionEventBatchID             *int64
		CompletionEvent                    []byte
		CompletionEventEncoding            *string
		TaskList                           *string
		WorkflowTypeName                   *string
		WorkflowTimeout                    *time.Duration
		DecisionTaskTimeout                *time.Duration
		ExecutionContext                   []byte
		State                              *int32
		CloseStatus                        *int32
		StartVersion                       *int64
		LastWriteEventID                   *int64
		LastEventTaskID                    *int64
		LastFirstEventID                   *int64
		LastProcessedEvent                 *int64
		StartTimestamp                     *time.Time
		LastUpdatedTimestamp               *time.Time
		DecisionVersion                    *int64
		DecisionScheduleID                 *int64
		DecisionStartedID                  *int64
		DecisionTimeout                    *time.Duration
		DecisionAttempt                    *int64
		DecisionStartedTimestamp           *time.Time
		DecisionScheduledTimestamp         *time.Time
		CancelRequested                    *bool
		DecisionOriginalScheduledTimestamp *time.Time
		CreateRequestID                    *string
		DecisionRequestID                  *string
		CancelRequestID                    *string
		StickyTaskList                     *string
		StickyScheduleToStartTimeout       *time.Duration
		RetryAttempt                       *int64
		RetryInitialInterval               *time.Duration
		RetryMaximumInterval               *time.Duration
		RetryMaximumAttempts               *int32
		RetryExpiration                    *time.Duration
		RetryBackoffCoefficient            *float64
		RetryExpirationTimestamp           *time.Time
		RetryNonRetryableErrors            []string
		HasRetryPolicy                     *bool
		CronSchedule                       *string
		EventStoreVersion                  *int32
		EventBranchToken                   []byte
		SignalCount                        *int64
		HistorySize                        *int64
		ClientLibraryVersion               *string
		ClientFeatureVersion               *string
		ClientImpl                         *string
		AutoResetPoints                    []byte
		AutoResetPointsEncoding            *string
		SearchAttributes                   map[string][]byte
		Memo                               map[string][]byte
		VersionHistories                   []byte
		VersionHistoriesEncoding           *string
	}

	// ActivityInfo blob in a serialization agnostic format
	ActivityInfo struct {
		Version                  *int64
		ScheduledEventBatchID    *int64
		ScheduledEvent           []byte
		ScheduledEventEncoding   *string
		ScheduledTimestamp       *time.Time
		StartedID                *int64
		StartedEvent             []byte
		StartedEventEncoding     *string
		StartedTimestamp         *time.Time
		ActivityID               *string
		RequestID                *string
		ScheduleToStartTimeout   *time.Duration
		ScheduleToCloseTimeout   *time.Duration
		StartToCloseTimeout      *time.Duration
		HeartbeatTimeout         *time.Duration
		CancelRequested          *bool
		CancelRequestID          *int64
		TimerTaskStatus          *int32
		Attempt                  *int32
		TaskList                 *string
		StartedIdentity          *string
		HasRetryPolicy           *bool
		RetryInitialInterval     *time.Duration
		RetryMaximumInterval     *time.Duration
		RetryMaximumAttempts     *int32
		RetryExpirationTimestamp *time.Time
		RetryBackoffCoefficient  *float64
		RetryNonRetryableErrors  []string
		RetryLastFailureReason   *string
		RetryLastWorkerIdentity  *string
		RetryLastFailureDetails  []byte
	}

	// ChildExecutionInfo blob in a serialization agnostic format
	ChildExecutionInfo struct {
		Version                *int64
		InitiatedEventBatchID  *int64
		StartedID              *int64
		InitiatedEvent         []byte
		InitiatedEventEncoding *string
		StartedWorkflowID      *string
		StartedRunID           UUID
		StartedEvent           []byte
		StartedEventEncoding   *string
		CreateRequestID        *string
		DomainName             *string
		WorkflowTypeName       *string
		ParentClosePolicy      *int32
	}

	// SignalInfo blob in a serialization agnostic format
	SignalInfo struct {
		Version               *int64
		InitiatedEventBatchID *int64
		RequestID             *string
		Name                  *string
		Input                 []byte
		Control               []byte
	}

	// RequestCancelInfo blob in a serialization agnostic format
	RequestCancelInfo struct {
		Version               *int64
		InitiatedEventBatchID *int64
		CancelRequestID       *string
	}

	// TimerInfo blob in a serialization agnostic format
	TimerInfo struct {
		Version         *int64
		StartedID       *int64
		ExpiryTimestamp *time.Time
		TaskID          *int64
	}

	// TaskInfo blob in a serialization agnostic format
	TaskInfo struct {
		WorkflowID       *string
		RunID            UUID
		ScheduleID       *int64
		ExpiryTimestamp  *time.Time
		CreatedTimestamp *time.Time
	}

	// TaskListInfo blob in a serialization agnostic format
	TaskListInfo struct {
		Kind            *int16
		AckLevel        *int64
		ExpiryTimestamp *time.Time
		LastUpdated     *time.Time
	}

	// TransferTaskInfo blob in a serialization agnostic format
	TransferTaskInfo struct {
		DomainID                UUID
		WorkflowID              *string
		RunID                   UUID
		TaskType                *int16
		TargetDomainID          UUID
		TargetWorkflowID        *string
		TargetRunID             UUID
		TaskList                *string
		TargetChildWorkflowOnly *bool
		ScheduleID              *int64
		Version                 *int64
		VisibilityTimestamp     *time.Time
	}

	// TimerTaskInfo blob in a serialization agnostic format
	TimerTaskInfo struct {
		DomainID        UUID
		WorkflowID      *string
		RunID           UUID
		TaskType        *int16
		TimeoutType     *int16
		Version         *int64
		ScheduleAttempt *int64
		EventID         *int64
	}

	// ReplicationTaskInfo blob in a serialization agnostic format
	ReplicationTaskInfo struct {
		DomainID                UUID
		WorkflowID              *string
		RunID                   UUID
		TaskType                *int16
		Version                 *int64
		FirstEventID            *int64
		NextEventID             *int64
		ScheduledID             *int64
		EventStoreVersion       *int32
		NewRunEventStoreVersion *int32
		BranchToken             []byte
		NewRunBranchToken       []byte
		CreationTimestamp       *time.Time
	}
)

type (
	// Parser is used to do serialization and deserialization. A parser is backed by a
	// a single encoder which encodes into one format and a collection of decoders.
	// Parser selects the appropriate decoder for the provided blob.
	Parser interface {
		ShardInfoToBlob(*sqlblobs.ShardInfo) (persistence.DataBlob, error)
		DomainInfoToBlob(*sqlblobs.DomainInfo) (persistence.DataBlob, error)
		HistoryTreeInfoToBlob(*sqlblobs.HistoryTreeInfo) (persistence.DataBlob, error)
		WorkflowExecutionInfoToBlob(*sqlblobs.WorkflowExecutionInfo) (persistence.DataBlob, error)
		ActivityInfoToBlob(*sqlblobs.ActivityInfo) (persistence.DataBlob, error)
		ChildExecutionInfoToBlob(*sqlblobs.ChildExecutionInfo) (persistence.DataBlob, error)
		SignalInfoToBlob(*sqlblobs.SignalInfo) (persistence.DataBlob, error)
		RequestCancelInfoToBlob(*sqlblobs.RequestCancelInfo) (persistence.DataBlob, error)
		TimerInfoToBlob(*sqlblobs.TimerInfo) (persistence.DataBlob, error)
		TaskInfoToBlob(*sqlblobs.TaskInfo) (persistence.DataBlob, error)
		TaskListInfoToBlob(*sqlblobs.TaskListInfo) (persistence.DataBlob, error)
		TransferTaskInfoToBlob(*sqlblobs.TransferTaskInfo) (persistence.DataBlob, error)
		TimerTaskInfoToBlob(*sqlblobs.TimerTaskInfo) (persistence.DataBlob, error)
		ReplicationTaskInfoToBlob(*sqlblobs.ReplicationTaskInfo) (persistence.DataBlob, error)

		ShardInfoFromBlob([]byte, string) (*sqlblobs.ShardInfo, error)
		DomainInfoFromBlob([]byte, string) (*sqlblobs.DomainInfo, error)
		HistoryTreeInfoFromBlob([]byte, string) (*sqlblobs.HistoryTreeInfo, error)
		WorkflowExecutionInfoFromBlob([]byte, string) (*sqlblobs.WorkflowExecutionInfo, error)
		ActivityInfoFromBlob([]byte, string) (*sqlblobs.ActivityInfo, error)
		ChildExecutionInfoFromBlob([]byte, string) (*sqlblobs.ChildExecutionInfo, error)
		SignalInfoFromBlob([]byte, string) (*sqlblobs.SignalInfo, error)
		RequestCancelInfoFromBlob([]byte, string) (*sqlblobs.RequestCancelInfo, error)
		TimerInfoFromBlob([]byte, string) (*sqlblobs.TimerInfo, error)
		TaskInfoFromBlob([]byte, string) (*sqlblobs.TaskInfo, error)
		TaskListInfoFromBlob([]byte, string) (*sqlblobs.TaskListInfo, error)
		TransferTaskInfoFromBlob([]byte, string) (*sqlblobs.TransferTaskInfo, error)
		TimerTaskInfoFromBlob([]byte, string) (*sqlblobs.TimerTaskInfo, error)
		ReplicationTaskInfoFromBlob([]byte, string) (*sqlblobs.ReplicationTaskInfo, error)
	}

	// encoder is used to serialize structs. Each encoder implementation uses one serialization format.
	encoder interface {
		shardInfoToBlob(*sqlblobs.ShardInfo) ([]byte, error)
		domainInfoToBlob(*sqlblobs.DomainInfo) ([]byte, error)
		historyTreeInfoToBlob(*sqlblobs.HistoryTreeInfo) ([]byte, error)
		workflowExecutionInfoToBlob(*sqlblobs.WorkflowExecutionInfo) ([]byte, error)
		activityInfoToBlob(*sqlblobs.ActivityInfo) ([]byte, error)
		childExecutionInfoToBlob(*sqlblobs.ChildExecutionInfo) ([]byte, error)
		signalInfoToBlob(*sqlblobs.SignalInfo) ([]byte, error)
		requestCancelInfoToBlob(*sqlblobs.RequestCancelInfo) ([]byte, error)
		timerInfoToBlob(*sqlblobs.TimerInfo) ([]byte, error)
		taskInfoToBlob(*sqlblobs.TaskInfo) ([]byte, error)
		taskListInfoToBlob(*sqlblobs.TaskListInfo) ([]byte, error)
		transferTaskInfoToBlob(*sqlblobs.TransferTaskInfo) ([]byte, error)
		timerTaskInfoToBlob(*sqlblobs.TimerTaskInfo) ([]byte, error)
		replicationTaskInfoToBlob(*sqlblobs.ReplicationTaskInfo) ([]byte, error)
		encodingType() common.EncodingType
	}

	// decoder is used to deserialize structs. Each decoder implementation uses one serialization format.
	decoder interface {
		shardInfoFromBlob([]byte) (*sqlblobs.ShardInfo, error)
		domainInfoFromBlob([]byte) (*sqlblobs.DomainInfo, error)
		historyTreeInfoFromBlob([]byte) (*sqlblobs.HistoryTreeInfo, error)
		workflowExecutionInfoFromBlob([]byte) (*sqlblobs.WorkflowExecutionInfo, error)
		activityInfoFromBlob([]byte) (*sqlblobs.ActivityInfo, error)
		childExecutionInfoFromBlob([]byte) (*sqlblobs.ChildExecutionInfo, error)
		signalInfoFromBlob([]byte) (*sqlblobs.SignalInfo, error)
		requestCancelInfoFromBlob([]byte) (*sqlblobs.RequestCancelInfo, error)
		timerInfoFromBlob([]byte) (*sqlblobs.TimerInfo, error)
		taskInfoFromBlob([]byte) (*sqlblobs.TaskInfo, error)
		taskListInfoFromBlob([]byte) (*sqlblobs.TaskListInfo, error)
		transferTaskInfoFromBlob([]byte) (*sqlblobs.TransferTaskInfo, error)
		timerTaskInfoFromBlob([]byte) (*sqlblobs.TimerTaskInfo, error)
		replicationTaskInfoFromBlob([]byte) (*sqlblobs.ReplicationTaskInfo, error)
	}

	thriftRWType interface {
		ToWire() (wire.Value, error)
		FromWire(w wire.Value) error
	}
)
