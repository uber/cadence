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

package persistence

import (
	"time"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
)

type (

	//Persistence interface is a lower layer of dataInterface. The intention is to let different persistence implementation(SQL,Cassandra/etc) share some common logic
	// Right now the only common part is serialization/deserialization, and only ExecutionManager/HistoryManager need it.
	// ShardManager/TaskManager/MetadataManager are the same.
	// PersistenceShardManager is a lower level of ShardManager
	PersistenceShardManager = ShardManager
	// PersistenceTaskManager is a lower level of TaskManager
	PersistenceTaskManager = TaskManager
	// PersistenceMetadataManager is a lower level of MetadataManager
	PersistenceMetadataManager = MetadataManager

	// PersistenceExecutionManager is used to manage workflow executions for Persistence layer
	PersistenceExecutionManager interface {
		Closeable

		//The below three APIs are related to serialization/deserialization
		GetWorkflowExecution(request *GetWorkflowExecutionRequest) (*PersistenceGetWorkflowExecutionResponse, error)
		UpdateWorkflowExecution(request *PersistenceUpdateWorkflowExecutionRequest) error
		ResetMutableState(request *PersistenceResetMutableStateRequest) error

		CreateWorkflowExecution(request *CreateWorkflowExecutionRequest) (*CreateWorkflowExecutionResponse, error)
		DeleteWorkflowExecution(request *DeleteWorkflowExecutionRequest) error
		GetCurrentExecution(request *GetCurrentExecutionRequest) (*GetCurrentExecutionResponse, error)

		// Transfer task related methods
		GetTransferTasks(request *GetTransferTasksRequest) (*GetTransferTasksResponse, error)
		CompleteTransferTask(request *CompleteTransferTaskRequest) error
		RangeCompleteTransferTask(request *RangeCompleteTransferTaskRequest) error

		// Replication task related methods
		GetReplicationTasks(request *GetReplicationTasksRequest) (*GetReplicationTasksResponse, error)
		CompleteReplicationTask(request *CompleteReplicationTaskRequest) error

		// Timer related methods.
		GetTimerIndexTasks(request *GetTimerIndexTasksRequest) (*GetTimerIndexTasksResponse, error)
		CompleteTimerTask(request *CompleteTimerTaskRequest) error
		RangeCompleteTimerTask(request *RangeCompleteTimerTaskRequest) error
	}

	// PersistenceHistoryManager is used to manage Workflow Execution HistoryEventBatch for Persistence layer
	PersistenceHistoryManager interface {
		Closeable

		//The below two APIs are related to serialization/deserialization
		AppendHistoryEvents(request *PersistenceAppendHistoryEventsRequest) error
		GetWorkflowExecutionHistory(request *PersistenceGetWorkflowExecutionHistoryRequest) (*PersistenceGetWorkflowExecutionHistoryResponse, error)

		DeleteWorkflowExecutionHistory(request *DeleteWorkflowExecutionHistoryRequest) error
	}

	// DataBlob represents a blob for any binary data.
	// It contains raw data, and metadata(right now only encoding) in other field
	// Note that it should be only used for Persistence layer, below dataInterface and application(historyEngine/etc)
	DataBlob struct {
		Encoding common.EncodingType
		Data     []byte
	}

	// PersistenceWorkflowExecutionInfo describes a workflow execution for Persistence Interface
	PersistenceWorkflowExecutionInfo struct {
		DomainID                     string
		WorkflowID                   string
		RunID                        string
		ParentDomainID               string
		ParentWorkflowID             string
		ParentRunID                  string
		InitiatedID                  int64
		CompletionEvent              *DataBlob
		TaskList                     string
		WorkflowTypeName             string
		WorkflowTimeout              int32
		DecisionTimeoutValue         int32
		ExecutionContext             []byte
		State                        int
		CloseStatus                  int
		LastFirstEventID             int64
		NextEventID                  int64
		LastProcessedEvent           int64
		StartTimestamp               time.Time
		LastUpdatedTimestamp         time.Time
		CreateRequestID              string
		HistorySize                  int64
		DecisionVersion              int64
		DecisionScheduleID           int64
		DecisionStartedID            int64
		DecisionRequestID            string
		DecisionTimeout              int32
		DecisionAttempt              int64
		DecisionTimestamp            int64
		CancelRequested              bool
		CancelRequestID              string
		StickyTaskList               string
		StickyScheduleToStartTimeout int32
		ClientLibraryVersion         string
		ClientFeatureVersion         string
		ClientImpl                   string
		// for retry
		Attempt            int32
		HasRetryPolicy     bool
		InitialInterval    int32
		BackoffCoefficient float64
		MaximumInterval    int32
		ExpirationTime     time.Time
		MaximumAttempts    int32
		NonRetriableErrors []string
	}

	// PersistenceWorkflowMutableState indicates workflow related state for Persistence Interface
	PersistenceWorkflowMutableState struct {
		ActivitInfos             map[int64]*PersistenceActivityInfo
		TimerInfos               map[string]*TimerInfo
		ChildExecutionInfos      map[int64]*PersistenceChildExecutionInfo
		RequestCancelInfos       map[int64]*RequestCancelInfo
		SignalInfos              map[int64]*SignalInfo
		SignalRequestedIDs       map[string]struct{}
		ExecutionInfo            *PersistenceWorkflowExecutionInfo
		ReplicationState         *ReplicationState
		BufferedEvents           []*DataBlob
		BufferedReplicationTasks map[int64]*PersistenceBufferedReplicationTask
	}

	// PersistenceActivityInfo details  for Persistence Interface
	PersistenceActivityInfo struct {
		Version                  int64
		ScheduleID               int64
		ScheduledEvent           *DataBlob
		ScheduledTime            time.Time
		StartedID                int64
		StartedEvent             *DataBlob
		StartedTime              time.Time
		ActivityID               string
		RequestID                string
		Details                  []byte
		ScheduleToStartTimeout   int32
		ScheduleToCloseTimeout   int32
		StartToCloseTimeout      int32
		HeartbeatTimeout         int32
		CancelRequested          bool
		CancelRequestID          int64
		LastHeartBeatUpdatedTime time.Time
		TimerTaskStatus          int32
		// For retry
		Attempt            int32
		DomainID           string
		StartedIdentity    string
		TaskList           string
		HasRetryPolicy     bool
		InitialInterval    int32
		BackoffCoefficient float64
		MaximumInterval    int32
		ExpirationTime     time.Time
		MaximumAttempts    int32
		NonRetriableErrors []string
		// Not written to database - This is used only for deduping heartbeat timer creation
		LastTimeoutVisibility int64
	}

	// PersistenceChildExecutionInfo has details for pending child executions  for Persistence Interface
	PersistenceChildExecutionInfo struct {
		Version         int64
		InitiatedID     int64
		InitiatedEvent  *DataBlob
		StartedID       int64
		StartedEvent    *DataBlob
		CreateRequestID string
	}

	// PersistenceBufferedReplicationTask has details to handle out of order receive of history events  for Persistence Interface
	PersistenceBufferedReplicationTask struct {
		FirstEventID  int64
		NextEventID   int64
		Version       int64
		History       *DataBlob
		NewRunHistory *DataBlob
	}

	// PersistenceUpdateWorkflowExecutionRequest is used to update a workflow execution  for Persistence Interface
	PersistenceUpdateWorkflowExecutionRequest struct {
		ExecutionInfo        *PersistenceWorkflowExecutionInfo
		ReplicationState     *ReplicationState
		TransferTasks        []Task
		TimerTasks           []Task
		ReplicationTasks     []Task
		DeleteTimerTask      Task
		Condition            int64
		RangeID              int64
		ContinueAsNew        *CreateWorkflowExecutionRequest
		FinishExecution      bool
		FinishedExecutionTTL int32

		// Mutable state
		UpsertActivityInfos           []*PersistenceActivityInfo
		DeleteActivityInfos           []int64
		UpserTimerInfos               []*TimerInfo
		DeleteTimerInfos              []string
		UpsertChildExecutionInfos     []*PersistenceChildExecutionInfo
		DeleteChildExecutionInfo      *int64
		UpsertRequestCancelInfos      []*RequestCancelInfo
		DeleteRequestCancelInfo       *int64
		UpsertSignalInfos             []*SignalInfo
		DeleteSignalInfo              *int64
		UpsertSignalRequestedIDs      []string
		DeleteSignalRequestedID       string
		NewBufferedEvents             *DataBlob
		ClearBufferedEvents           bool
		NewBufferedReplicationTask    *PersistenceBufferedReplicationTask
		DeleteBufferedReplicationTask *int64
	}

	// PersistenceResetMutableStateRequest is used to reset workflow execution state  for Persistence Interface
	PersistenceResetMutableStateRequest struct {
		PrevRunID        string
		ExecutionInfo    *PersistenceWorkflowExecutionInfo
		ReplicationState *ReplicationState
		Condition        int64
		RangeID          int64

		// Mutable state
		InsertActivityInfos       []*PersistenceActivityInfo
		InsertTimerInfos          []*TimerInfo
		InsertChildExecutionInfos []*PersistenceChildExecutionInfo
		InsertRequestCancelInfos  []*RequestCancelInfo
		InsertSignalInfos         []*SignalInfo
		InsertSignalRequestedIDs  []string
	}

	// PersistenceAppendHistoryEventsRequest is used to append new events to workflow execution history  for Persistence Interface
	PersistenceAppendHistoryEventsRequest struct {
		DomainID          string
		Execution         workflow.WorkflowExecution
		FirstEventID      int64
		EventBatchVersion int64
		RangeID           int64
		TransactionID     int64
		Events            *DataBlob
		Overwrite         bool
	}

	// PersistenceGetWorkflowExecutionResponse is the response to GetworkflowExecutionRequest for Persistence Interface
	PersistenceGetWorkflowExecutionResponse struct {
		State *PersistenceWorkflowMutableState
	}

	// PersistenceGetWorkflowExecutionHistoryRequest is used to retrieve history of a workflow execution
	PersistenceGetWorkflowExecutionHistoryRequest struct {
		// an extra field passing from GetWorkflowExecutionHistoryRequest
		LastEventBatchVersion int64

		DomainID  string
		Execution workflow.WorkflowExecution
		// Get the history events from FirstEventID. Inclusive.
		FirstEventID int64
		// Get the history events upto NextEventID.  Not Inclusive.
		NextEventID int64
		// Maximum number of history append transactions per page
		PageSize int
		// Token to continue reading next page of history append transactions.  Pass in empty slice for first page
		NextPageToken []byte
	}

	// PersistenceGetWorkflowExecutionHistoryResponse is the response to GetWorkflowExecutionHistoryRequest for Persistence Interface
	PersistenceGetWorkflowExecutionHistoryResponse struct {
		History []*DataBlob
		// Token to read next page if there are more events beyond page size.
		// Use this to set NextPageToken on GetworkflowExecutionHistoryRequest to read the next page.
		NextPageToken []byte
		// an extra field passing to DataInterface
		LastEventBatchVersion int64
	}
)

// NewDataBlob returns a new DataBlob
func NewDataBlob(data []byte, encodingType common.EncodingType) *DataBlob {
	return &DataBlob{
		Data:     data,
		Encoding: encodingType,
	}
}

// GetEncoding returns encoding type
func (d *DataBlob) GetEncoding() common.EncodingType {
	encodingStr := string(d.Encoding)

	switch common.EncodingType(encodingStr) {
	case common.EncodingTypeGob:
		return common.EncodingTypeGob
	case common.EncodingTypeJSON:
		return common.EncodingTypeJSON
	case common.EncodingTypeThriftRW:
		return common.EncodingTypeThriftRW
	default:
		return common.EncodingTypeUnknown
	}
}
