// Copyright (c) 2021 Uber Technologies, Inc.
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
package nosqlplugin

import (
	"time"

	"github.com/uber/cadence/common/checksum"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

type (
	WorkflowExecution = persistence.InternalWorkflowMutableState

	WorkflowExecutionRequest struct {
		// basic information/data
		persistence.InternalWorkflowExecutionInfo
		VersionHistories *persistence.DataBlob
		Checksums        *checksum.Checksum
		LastWriteVersion int64
		// condition checking for updating execution info
		PreviousNextEventIDCondition *int64

		// MapsWriteMode controls how to write into the six maps(activityInfoMap, timerInfoMap, childWorkflowInfoMap, signalInfoMap and signalRequestedIDs)
		MapsWriteMode WorkflowExecutionMapsWriteMode

		// For WorkflowExecutionMapsWriteMode of create, update and reset
		ActivityInfos      map[int64]*persistence.InternalActivityInfo
		TimerInfos         map[string]*persistence.TimerInfo
		ChildWorkflowInfos map[int64]*persistence.InternalChildExecutionInfo
		RequestCancelInfos map[int64]*persistence.RequestCancelInfo
		SignalInfos        map[int64]*persistence.SignalInfo
		SignalRequestedIDs []string // This map has no value, hence use array to store keys

		// For WorkflowExecutionMapsWriteMode of update only
		ActivityInfoKeysToDelete       []int64
		TimerInfoKeysToDelete          []string
		ChildWorkflowInfoKeysToDelete  []int64
		RequestCancelInfoKeysToDelete  []int64
		SignalInfoKeysToDelete         []int64
		SignalRequestedIDsKeysToDelete []string

		// EventBufferWriteMode controls how to write into the buffered event list
		// only needed for UpdateWorkflowExecutionWithTasks API
		EventBufferWriteMode EventBufferWriteMode
		// the batch of event to be appended, only for EventBufferWriteModeAppend
		NewBufferedEventBatch *persistence.DataBlob
	}

	WorkflowExecutionMapsWriteMode int
	EventBufferWriteMode           int

	TimerTask = persistence.TimerTaskInfo

	ReplicationTask = persistence.InternalReplicationTaskInfo

	CrossClusterTask struct {
		TransferTask
		TargetCluster string
	}

	TransferTask = persistence.TransferTaskInfo

	ShardCondition struct {
		ShardID int
		RangeID int64
	}

	CurrentWorkflowWriteRequest struct {
		WriteMode CurrentWorkflowWriteMode
		Row       CurrentWorkflowRow
		Condition *CurrentWorkflowWriteCondition
	}

	CurrentWorkflowWriteCondition struct {
		CurrentRunID     *string
		LastWriteVersion *int64
		State            *int
	}

	CurrentWorkflowWriteMode int

	CurrentWorkflowRow struct {
		ShardID          int
		DomainID         string
		WorkflowID       string
		RunID            string
		State            int
		CloseStatus      int
		CreateRequestID  string
		LastWriteVersion int64
	}

	TasksFilter struct {
		TaskListFilter
		// Exclusive
		MinTaskID int64
		// Inclusive
		MaxTaskID int64
		BatchSize int
	}

	TaskRowForInsert struct {
		TaskRow
		// <= 0 means no TTL
		TTLSeconds int
	}

	TaskRow struct {
		DomainID     string
		TaskListName string
		TaskListType int
		TaskID       int64

		WorkflowID  string
		RunID       string
		ScheduledID int64
		CreatedTime time.Time
	}

	TaskListFilter struct {
		DomainID     string
		TaskListName string
		TaskListType int
	}

	TaskListRow struct {
		DomainID     string
		TaskListName string
		TaskListType int

		RangeID         int64
		TaskListKind    int
		AckLevel        int64
		LastUpdatedTime time.Time
	}

	ListTaskListResult struct {
		TaskLists     []*TaskListRow
		NextPageToken []byte
	}

	// For now ShardRow is the same as persistence.InternalShardInfo
	// Separate them later when there is a need.
	ShardRow = persistence.InternalShardInfo

	// ConflictedShardRow contains the partial information about a shard returned when a conditional write fails
	ConflictedShardRow struct {
		ShardID int
		// PreviousRangeID is the condition of previous change that used for conditional update
		PreviousRangeID int64
		// optional detailed information for logging purpose
		Details string
	}

	// DomainRow defines the row struct for queue message
	DomainRow struct {
		Info                        *persistence.DomainInfo
		Config                      *NoSQLInternalDomainConfig
		ReplicationConfig           *persistence.DomainReplicationConfig
		ConfigVersion               int64
		FailoverVersion             int64
		FailoverNotificationVersion int64
		PreviousFailoverVersion     int64
		FailoverEndTime             *time.Time
		NotificationVersion         int64
		LastUpdatedTime             time.Time
		IsGlobalDomain              bool
	}

	// NoSQLInternalDomainConfig defines the struct for the domainConfig
	NoSQLInternalDomainConfig struct {
		Retention                time.Duration
		EmitMetric               bool                 // deprecated
		ArchivalBucket           string               // deprecated
		ArchivalStatus           types.ArchivalStatus // deprecated
		HistoryArchivalStatus    types.ArchivalStatus
		HistoryArchivalURI       string
		VisibilityArchivalStatus types.ArchivalStatus
		VisibilityArchivalURI    string
		BadBinaries              *persistence.DataBlob
	}

	// SelectMessagesBetweenRequest is a request struct for SelectMessagesBetween
	SelectMessagesBetweenRequest struct {
		QueueType               persistence.QueueType
		ExclusiveBeginMessageID int64
		InclusiveEndMessageID   int64
		PageSize                int
		NextPageToken           []byte
	}

	// SelectMessagesBetweenResponse is a response struct for SelectMessagesBetween
	SelectMessagesBetweenResponse struct {
		Rows          []QueueMessageRow
		NextPageToken []byte
	}

	// QueueMessageRow defines the row struct for queue message
	QueueMessageRow struct {
		QueueType persistence.QueueType
		ID        int64
		Payload   []byte
	}

	// QueueMetadataRow defines the row struct for metadata
	QueueMetadataRow struct {
		QueueType        persistence.QueueType
		ClusterAckLevels map[string]int64
		Version          int64
	}

	// HistoryNodeRow represents a row in history_node table
	HistoryNodeRow struct {
		ShardID  int
		TreeID   string
		BranchID string
		NodeID   int64
		// Note: use pointer so that it's easier to multiple by -1 if needed
		TxnID        *int64
		Data         []byte
		DataEncoding string
	}

	// HistoryNodeFilter contains the column names within history_node table that
	// can be used to filter results through a WHERE clause
	HistoryNodeFilter struct {
		ShardID  int
		TreeID   string
		BranchID string
		// Inclusive
		MinNodeID int64
		// Exclusive
		MaxNodeID     int64
		NextPageToken []byte
		PageSize      int
	}

	// HistoryTreeRow represents a row in history_tree table
	HistoryTreeRow struct {
		ShardID         int
		TreeID          string
		BranchID        string
		Ancestors       []*types.HistoryBranchRange
		CreateTimestamp time.Time
		Info            string
	}

	// HistoryTreeFilter contains the column names within history_tree table that
	// can be used to filter results through a WHERE clause
	HistoryTreeFilter struct {
		ShardID  int
		TreeID   string
		BranchID *string
	}
)

const (
	AllOpen VisibilityFilterType = iota
	AllClosed
	OpenByWorkflowType
	ClosedByWorkflowType
	OpenByWorkflowID
	ClosedByWorkflowID
	ClosedByClosedStatus
)

const (
	SortByStartTime VisibilitySortType = iota
	SortByClosedTime
)

const (
	CurrentWorkflowWriteModeNoop CurrentWorkflowWriteMode = iota
	CurrentWorkflowWriteModeUpdate
	CurrentWorkflowWriteModeInsert
)

const (
	// WorkflowExecutionMapsWriteModeCreate will upsert new entry to maps
	WorkflowExecutionMapsWriteModeCreate WorkflowExecutionMapsWriteMode = iota
	// WorkflowExecutionMapsWriteModeUpdate will upsert new entry to maps and also delete entries from maps
	WorkflowExecutionMapsWriteModeUpdate
	// WorkflowExecutionMapsWriteModeReset will reset(override) the whole maps
	WorkflowExecutionMapsWriteModeReset
)

const (
	// EventBufferWriteModeNone is for not doing anything to the event buffer
	EventBufferWriteModeNone EventBufferWriteMode = iota
	// EventBufferWriteModeAppend will append a new event to the event buffer
	EventBufferWriteModeAppend
	// EventBufferWriteModeClear will clear(delete all event from) the event buffer
	EventBufferWriteModeClear
)

func (w *CurrentWorkflowWriteCondition) GetCurrentRunID() string {
	if w == nil || w.CurrentRunID == nil {
		return ""
	}
	return *w.CurrentRunID
}

