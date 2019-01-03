// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package sqldb

import (
	"database/sql"
	"time"
)

type (
	// QueryFilter contains the column names that can be used to
	// filter results through a SQL WHERE clause
	QueryFilter struct {
		Closed                bool
		ShardID               int
		TaskType              int
		NextEventID           int
		PageSize              int
		TaskID                int64
		DomainID              string
		DomainName            string
		WorkflowID            string
		RunID                 string
		TaskListName          string
		VisiblityTimestamp    time.Time
		ScheduleID            *int
		InitiatedID           *int
		FirstEventID          *int
		MinTaskID             *int64
		MaxTaskID             *int64
		TimerID               *string
		CloseStatus           *int
		WorkflowTypeName      *string
		SignalID              *string
		MinVisiblityTimestamp *time.Time
		MaxVisiblityTimestamp *time.Time
		MinStartTime          *time.Time
		MaxStartTime          *time.Time
	}

	// DomainRow represents a row in domain table
	DomainRow struct {
		ID                          string
		Name                        string
		Status                      int
		Description                 string
		OwnerEmail                  string
		Data                        []byte
		Retention                   int
		EmitMetric                  bool
		ArchivalBucket              string
		ArchivalStatus              int
		ConfigVersion               int64
		NotificationVersion         int64
		FailoverNotificationVersion int64
		FailoverVersion             int64
		IsGlobalDomain              bool
		ActiveClusterName           string
		Clusters                    []byte
	}

	// DomainMetadataRow represents a row in domain_metadata table
	DomainMetadataRow struct {
		NotificationVersion int64
	}

	// ShardsRow represents a row in shards table
	ShardsRow struct {
		ShardID                   int64
		Owner                     string
		RangeID                   int64
		StolenSinceRenew          int64
		UpdatedAt                 time.Time
		ReplicationAckLevel       int64
		TransferAckLevel          int64
		TimerAckLevel             time.Time
		ClusterTransferAckLevel   []byte
		ClusterTimerAckLevel      []byte
		DomainNotificationVersion int64
	}

	// TransferTasksRow represents a row in transfer_tasks table
	TransferTasksRow struct {
		ShardID                 int
		DomainID                string
		WorkflowID              string
		RunID                   string
		TaskID                  int64
		TaskType                int
		TargetDomainID          string
		TargetWorkflowID        string
		TargetRunID             string
		TargetChildWorkflowOnly bool
		TaskList                string
		ScheduleID              int64
		Version                 int64
		VisibilityTimestamp     time.Time
	}

	// ExecutionsRow represents a row in executions table
	ExecutionsRow struct {
		ShardID                      int
		DomainID                     string
		WorkflowID                   string
		RunID                        string
		ParentDomainID               *string
		ParentWorkflowID             *string
		ParentRunID                  *string
		InitiatedID                  *int64
		CompletionEventBatchID       *int64
		CompletionEvent              *[]byte
		CompletionEventEncoding      *string
		TaskList                     string
		WorkflowTypeName             string
		WorkflowTimeoutSeconds       int64
		DecisionTaskTimeoutMinutes   int64
		ExecutionContext             *[]byte
		State                        int64
		CloseStatus                  int64
		StartVersion                 int64
		CurrentVersion               int64
		LastWriteVersion             int64
		LastWriteEventID             *int64
		LastReplicationInfo          *[]byte
		LastFirstEventID             int64
		NextEventID                  int64
		LastProcessedEvent           int64
		StartTime                    time.Time
		LastUpdatedTime              time.Time
		CreateRequestID              string
		DecisionVersion              int64
		DecisionScheduleID           int64
		DecisionStartedID            int64
		DecisionRequestID            string
		DecisionTimeout              int64
		DecisionAttempt              int64
		DecisionTimestamp            int64
		CancelRequested              *int64
		CancelRequestID              *string
		StickyTaskList               string
		StickyScheduleToStartTimeout int64
		ClientLibraryVersion         string
		ClientFeatureVersion         string
		ClientImpl                   string
		SignalCount                  int
		CronSchedule                 string
	}

	// CurrentExecutionsRow represents a row in current_executions table
	CurrentExecutionsRow struct {
		ShardID          int64
		DomainID         string
		WorkflowID       string
		RunID            string
		CreateRequestID  string
		State            int
		CloseStatus      int
		LastWriteVersion int64
		StartVersion     int64
	}

	// BufferedEventsRow represents a row in buffered_events table
	BufferedEventsRow struct {
		ShardID      int
		DomainID     string
		WorkflowID   string
		RunID        string
		Data         []byte
		DataEncoding string
	}

	// TasksRow represents a row in tasks table
	TasksRow struct {
		DomainID     string
		WorkflowID   string
		RunID        string
		ScheduleID   int64
		TaskID       int64
		TaskType     int64
		TaskListName string
		ExpiryTs     time.Time
	}

	// TaskListsRow represents a row in task_lists table
	TaskListsRow struct {
		DomainID string
		RangeID  int64
		Name     string
		TaskType int64
		AckLevel int64
		Kind     int64
		ExpiryTs time.Time
	}

	// ReplicationTasksRow represents a row in replication_tasks table
	ReplicationTasksRow struct {
		DomainID            string
		WorkflowID          string
		RunID               string
		TaskID              int64
		TaskType            int
		FirstEventID        int64
		NextEventID         int64
		Version             int64
		LastReplicationInfo []byte
		ScheduledID         int64
		ShardID             int
	}

	// TimerTasksRow represents a row in timer_tasks table
	TimerTasksRow struct {
		ShardID             int
		DomainID            string
		WorkflowID          string
		RunID               string
		VisibilityTimestamp time.Time
		TaskID              int64
		TaskType            int
		TimeoutType         int
		EventID             int64
		ScheduleAttempt     int64
		Version             int64
	}

	// EventsRow represents a row in events table
	EventsRow struct {
		DomainID     string
		WorkflowID   string
		RunID        string
		FirstEventID int64
		BatchVersion int64
		RangeID      int64
		TxID         int64
		Data         []byte
		DataEncoding string
	}

	// ActivityInfoMapsRow represents a row in activity_info_maps table
	ActivityInfoMapsRow struct {
		ShardID                  int64
		DomainID                 string
		WorkflowID               string
		RunID                    string
		ScheduleID               int64
		Version                  int64
		ScheduledEventBatchID    int64
		ScheduledEvent           []byte
		ScheduledEventEncoding   string
		ScheduledTime            time.Time
		StartedID                int64
		StartedEvent             *[]byte
		StartedEventEncoding     string
		StartedTime              time.Time
		ActivityID               string
		RequestID                string
		Details                  *[]byte
		ScheduleToStartTimeout   int64
		ScheduleToCloseTimeout   int64
		StartToCloseTimeout      int64
		HeartbeatTimeout         int64
		CancelRequested          int64
		CancelRequestID          int64
		LastHeartbeatUpdatedTime time.Time
		TimerTaskStatus          int64
		Attempt                  int64
		TaskList                 string
		StartedIdentity          string
		HasRetryPolicy           int64
		InitInterval             int64
		BackoffCoefficient       float64
		MaxInterval              int64
		ExpirationTime           time.Time
		MaxAttempts              int64
		NonRetriableErrors       *[]byte
	}

	// TimerInfoMapsRow represents a row in timer_info_maps table
	TimerInfoMapsRow struct {
		ShardID    int64
		DomainID   string
		WorkflowID string
		RunID      string
		TimerID    string
		Version    int64
		StartedID  int64
		ExpiryTime time.Time
		TaskID     int64
	}

	// ChildExecutionInfoMapsRow represents a row in child_execution_info_maps table
	ChildExecutionInfoMapsRow struct {
		ShardID                int64
		DomainID               string
		WorkflowID             string
		RunID                  string
		InitiatedID            int64
		Version                int64
		InitiatedEventBatchID  int64
		InitiatedEvent         *[]byte
		InitiatedEventEncoding string
		StartedID              int64
		StartedWorkflowID      string
		StartedRunID           string
		StartedEvent           *[]byte
		StartedEventEncoding   string
		CreateRequestID        string
		DomainName             string
		WorkflowTypeName       string
	}

	// RequestCancelInfoMapsRow represents a row in request_cancel_info_maps table
	RequestCancelInfoMapsRow struct {
		ShardID         int64
		DomainID        string
		WorkflowID      string
		RunID           string
		InitiatedID     int64
		Version         int64
		CancelRequestID string
	}

	// SignalInfoMapsRow represents a row in signal_info_maps table
	SignalInfoMapsRow struct {
		ShardID         int64
		DomainID        string
		WorkflowID      string
		RunID           string
		InitiatedID     int64
		Version         int64
		SignalRequestID string
		SignalName      string
		Input           *[]byte
		Control         *[]byte
	}

	// BufferedReplicationTaskMapsRow represents a row in buffered_replication_task_maps table
	BufferedReplicationTaskMapsRow struct {
		ShardID               int64
		DomainID              string
		WorkflowID            string
		RunID                 string
		FirstEventID          int64
		NextEventID           int64
		Version               int64
		History               *[]byte
		HistoryEncoding       string
		NewRunHistory         *[]byte
		NewRunHistoryEncoding string
	}

	// SignalsRequestedSetsRow represents a row in signals_requested_sets table
	SignalsRequestedSetsRow struct {
		ShardID    int64
		DomainID   string
		WorkflowID string
		RunID      string
		SignalID   string
	}

	// VisibilityRow represents a row in executions_visibility table
	VisibilityRow struct {
		DomainID         string
		WorkflowID       string
		RunID            string
		WorkflowTypeName string
		StartTime        time.Time
		CloseStatus      *int32
		CloseTime        *time.Time
		HistoryLength    *int64
	}

	// tableCRUD defines the API for interacting with the database tables
	tableCRUD interface {
		InsertIntoDomain(rows *DomainRow) (sql.Result, error)
		UpdateDomain(row *DomainRow) (sql.Result, error)
		SelectFromDomain(filter *QueryFilter) ([]DomainRow, error)
		DeleteFromDomain(filter *QueryFilter) (sql.Result, error)

		LockDomainMetadata() error
		UpdateDomainMetadata(row *DomainMetadataRow) (sql.Result, error)
		SelectFromDomainMetadata() (*DomainMetadataRow, error)

		InsertIntoShards(rows *ShardsRow) (sql.Result, error)
		UpdateShards(row *ShardsRow) (sql.Result, error)
		SelectFromShards(filter *QueryFilter) (*ShardsRow, error)
		ReadLockShards(filter *QueryFilter) (int, error)
		WriteLockShards(filter *QueryFilter) (int, error)

		InsertIntoTasks(rows []TasksRow) (sql.Result, error)
		SelectFromTasks(filter *QueryFilter) ([]TasksRow, error)
		DeleteFromTasks(filter *QueryFilter) (sql.Result, error)

		InsertIntoTaskLists(row *TaskListsRow) (sql.Result, error)
		ReplaceIntoTaskLists(row *TaskListsRow) (sql.Result, error)
		UpdateTaskLists(row *TaskListsRow) (sql.Result, error)
		SelectFromTaskLists(filter *QueryFilter) (*TaskListsRow, error)
		DeleteFromTaskLists(filter *QueryFilter) (sql.Result, error)
		LockTaskLists(filter *QueryFilter) (int64, error)

		InsertIntoEvents(row *EventsRow) (sql.Result, error)
		UpdateEvents(rows *EventsRow) (sql.Result, error)
		SelectFromEvents(filter *QueryFilter) ([]EventsRow, error)
		DeleteFromEvents(filter *QueryFilter) (sql.Result, error)
		LockEvents(filter *QueryFilter) (*EventsRow, error)

		InsertIntoExecutions(row *ExecutionsRow) (sql.Result, error)
		UpdateExecutions(row *ExecutionsRow) (sql.Result, error)
		SelectFromExecutions(filter *QueryFilter) (*ExecutionsRow, error)
		DeleteFromExecutions(filter *QueryFilter) (sql.Result, error)
		LockExecutions(filter *QueryFilter) (int, error)

		LockCurrentExecutionsJoinExecutions(filter *QueryFilter) ([]CurrentExecutionsRow, error)

		InsertIntoCurrentExecutions(row *CurrentExecutionsRow) (sql.Result, error)
		UpdateCurrentExecutions(row *CurrentExecutionsRow) (sql.Result, error)
		SelectFromCurrentExecutions(filter *QueryFilter) ([]CurrentExecutionsRow, error)
		DeleteFromCurrentExecutions(filter *QueryFilter) (sql.Result, error)
		LockCurrentExecutions(filter *QueryFilter) (string, error)

		InsertIntoTransferTasks(rows []TransferTasksRow) (sql.Result, error)
		SelectFromTransferTasks(filter *QueryFilter) ([]TransferTasksRow, error)
		DeleteFromTransferTasks(filter *QueryFilter) (sql.Result, error)

		InsertIntoTimerTasks(rows []TimerTasksRow) (sql.Result, error)
		SelectFromTimerTasks(filter *QueryFilter) ([]TimerTasksRow, error)
		DeleteFromTimerTasks(filter *QueryFilter) (sql.Result, error)

		InsertIntoBufferedEvents(rows []BufferedEventsRow) (sql.Result, error)
		SelectFromBufferedEvents(filter *QueryFilter) ([]BufferedEventsRow, error)
		DeleteFromBufferedEvents(filter *QueryFilter) (sql.Result, error)

		InsertIntoReplicationTasks(rows []ReplicationTasksRow) (sql.Result, error)
		SelectFromReplicationTasks(filter *QueryFilter) ([]ReplicationTasksRow, error)
		DeleteFromReplicationTasks(filter *QueryFilter) (sql.Result, error)

		ReplaceIntoActivityInfoMaps(rows []ActivityInfoMapsRow) (sql.Result, error)
		SelectFromActivityInfoMaps(filter *QueryFilter) ([]ActivityInfoMapsRow, error)
		DeleteFromActivityInfoMaps(filter *QueryFilter) (sql.Result, error)

		ReplaceIntoTimerInfoMaps(rows []TimerInfoMapsRow) (sql.Result, error)
		SelectFromTimerInfoMaps(filter *QueryFilter) ([]TimerInfoMapsRow, error)
		DeleteFromTimerInfoMaps(filter *QueryFilter) (sql.Result, error)

		ReplaceIntoChildExecutionInfoMaps(rows []ChildExecutionInfoMapsRow) (sql.Result, error)
		SelectFromChildExecutionInfoMaps(filter *QueryFilter) ([]ChildExecutionInfoMapsRow, error)
		DeleteFromChildExecutionInfoMaps(filter *QueryFilter) (sql.Result, error)

		ReplaceIntoRequestCancelInfoMaps(rows []RequestCancelInfoMapsRow) (sql.Result, error)
		SelectFromRequestCancelInfoMaps(filter *QueryFilter) ([]RequestCancelInfoMapsRow, error)
		DeleteFromRequestCancelInfoMaps(filter *QueryFilter) (sql.Result, error)

		ReplaceIntoSignalInfoMaps(rows []SignalInfoMapsRow) (sql.Result, error)
		SelectFromSignalInfoMaps(filter *QueryFilter) ([]SignalInfoMapsRow, error)
		DeleteFromSignalInfoMaps(filter *QueryFilter) (sql.Result, error)

		ReplaceIntoBufferedReplicationTasks(rows *BufferedReplicationTaskMapsRow) (sql.Result, error)
		SelectFromBufferedReplicationTasks(filter *QueryFilter) ([]BufferedReplicationTaskMapsRow, error)
		DeleteFromBufferedReplicationTasks(filter *QueryFilter) (sql.Result, error)

		InsertIntoSignalsRequestedSets(rows []SignalsRequestedSetsRow) (sql.Result, error)
		SelectFromSignalsRequestedSets(filter *QueryFilter) ([]SignalsRequestedSetsRow, error)
		DeleteFromSignalsRequestedSets(filter *QueryFilter) (sql.Result, error)

		InsertIntoVisibility(row *VisibilityRow) (sql.Result, error)
		UpdateVisibility(row *VisibilityRow) (sql.Result, error)
		SelectFromVisibility(filter *QueryFilter) ([]VisibilityRow, error)
	}

	// Tx defines the API for a SQL transaction
	Tx interface {
		tableCRUD
		Commit() error
		Rollback() error
	}

	// Interface defines the API for a SQL database
	Interface interface {
		tableCRUD
		BeginTx() (Tx, error)
		DriverName() string
		Close() error
	}

	// Conn defines the API for a single database connection
	Conn interface {
		Exec(query string, args ...interface{}) (sql.Result, error)
		NamedExec(query string, arg interface{}) (sql.Result, error)
		Get(dest interface{}, query string, args ...interface{}) error
		Select(dest interface{}, query string, args ...interface{}) error
	}
)
