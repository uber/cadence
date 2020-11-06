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
	"context"
	"fmt"
	"time"

	"github.com/uber/cadence/common/types"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/checksum"
)

type (
	//////////////////////////////////////////////////////////////////////
	// Persistence interface is a lower layer of dataInterface.
	// The intention is to let different persistence implementation(SQL,Cassandra/etc) share some common logic
	// Right now the only common part is serialization/deserialization, and only ExecutionManager/HistoryManager need it.
	// TaskManager are the same.
	//////////////////////////////////////////////////////////////////////

	// ShardStore is the lower level of ShardManager
	ShardStore interface {
		Closeable
		GetName() string
		CreateShard(ctx context.Context, request *InternalCreateShardRequest) error
		GetShard(ctx context.Context, request *InternalGetShardRequest) (*InternalGetShardResponse, error)
		UpdateShard(ctx context.Context, request *InternalUpdateShardRequest) error
	}

	// TaskStore is a lower level of TaskManager
	TaskStore interface {
		Closeable
		GetName() string
		LeaseTaskList(ctx context.Context, request *LeaseTaskListRequest) (*LeaseTaskListResponse, error)
		UpdateTaskList(ctx context.Context, request *UpdateTaskListRequest) (*UpdateTaskListResponse, error)
		ListTaskList(ctx context.Context, request *ListTaskListRequest) (*ListTaskListResponse, error)
		DeleteTaskList(ctx context.Context, request *DeleteTaskListRequest) error
		CreateTasks(ctx context.Context, request *InternalCreateTasksRequest) (*CreateTasksResponse, error)
		GetTasks(ctx context.Context, request *GetTasksRequest) (*InternalGetTasksResponse, error)
		CompleteTask(ctx context.Context, request *CompleteTaskRequest) error
		// CompleteTasksLessThan completes tasks less than or equal to the given task id
		// This API takes a limit parameter which specifies the count of maxRows that
		// can be deleted. This parameter may be ignored by the underlying storage, but
		// its mandatory to specify it. On success this method returns the number of rows
		// actually deleted. If the underlying storage doesn't support "limit", all rows
		// less than or equal to taskID will be deleted.
		// On success, this method returns:
		//  - number of rows actually deleted, if limit is honored
		//  - UnknownNumRowsDeleted, when all rows below value are deleted
		CompleteTasksLessThan(ctx context.Context, request *CompleteTasksLessThanRequest) (int, error)
	}

	// MetadataStore is a lower level of MetadataManager
	MetadataStore interface {
		Closeable
		GetName() string
		CreateDomain(ctx context.Context, request *InternalCreateDomainRequest) (*CreateDomainResponse, error)
		GetDomain(ctx context.Context, request *GetDomainRequest) (*InternalGetDomainResponse, error)
		UpdateDomain(ctx context.Context, request *InternalUpdateDomainRequest) error
		DeleteDomain(ctx context.Context, request *DeleteDomainRequest) error
		DeleteDomainByName(ctx context.Context, request *DeleteDomainByNameRequest) error
		ListDomains(ctx context.Context, request *ListDomainsRequest) (*InternalListDomainsResponse, error)
		GetMetadata(ctx context.Context) (*GetMetadataResponse, error)
	}

	// ExecutionStore is used to manage workflow executions for Persistence layer
	ExecutionStore interface {
		Closeable
		GetName() string
		GetShardID() int
		//The below three APIs are related to serialization/deserialization
		GetWorkflowExecution(ctx context.Context, request *InternalGetWorkflowExecutionRequest) (*InternalGetWorkflowExecutionResponse, error)
		UpdateWorkflowExecution(ctx context.Context, request *InternalUpdateWorkflowExecutionRequest) error
		ConflictResolveWorkflowExecution(ctx context.Context, request *InternalConflictResolveWorkflowExecutionRequest) error
		ResetWorkflowExecution(ctx context.Context, request *InternalResetWorkflowExecutionRequest) error

		CreateWorkflowExecution(ctx context.Context, request *InternalCreateWorkflowExecutionRequest) (*CreateWorkflowExecutionResponse, error)
		DeleteWorkflowExecution(ctx context.Context, request *DeleteWorkflowExecutionRequest) error
		DeleteCurrentWorkflowExecution(ctx context.Context, request *DeleteCurrentWorkflowExecutionRequest) error
		GetCurrentExecution(ctx context.Context, request *GetCurrentExecutionRequest) (*GetCurrentExecutionResponse, error)
		IsWorkflowExecutionExists(ctx context.Context, request *IsWorkflowExecutionExistsRequest) (*IsWorkflowExecutionExistsResponse, error)

		// Transfer task related methods
		GetTransferTasks(ctx context.Context, request *GetTransferTasksRequest) (*GetTransferTasksResponse, error)
		CompleteTransferTask(ctx context.Context, request *CompleteTransferTaskRequest) error
		RangeCompleteTransferTask(ctx context.Context, request *RangeCompleteTransferTaskRequest) error

		// Replication task related methods
		GetReplicationTasks(ctx context.Context, request *GetReplicationTasksRequest) (*InternalGetReplicationTasksResponse, error)
		CompleteReplicationTask(ctx context.Context, request *CompleteReplicationTaskRequest) error
		RangeCompleteReplicationTask(ctx context.Context, request *RangeCompleteReplicationTaskRequest) error
		PutReplicationTaskToDLQ(ctx context.Context, request *InternalPutReplicationTaskToDLQRequest) error
		GetReplicationTasksFromDLQ(ctx context.Context, request *GetReplicationTasksFromDLQRequest) (*InternalGetReplicationTasksFromDLQResponse, error)
		GetReplicationDLQSize(ctx context.Context, request *GetReplicationDLQSizeRequest) (*GetReplicationDLQSizeResponse, error)
		DeleteReplicationTaskFromDLQ(ctx context.Context, request *DeleteReplicationTaskFromDLQRequest) error
		RangeDeleteReplicationTaskFromDLQ(ctx context.Context, request *RangeDeleteReplicationTaskFromDLQRequest) error
		CreateFailoverMarkerTasks(ctx context.Context, request *CreateFailoverMarkersRequest) error

		// Timer related methods.
		GetTimerIndexTasks(ctx context.Context, request *GetTimerIndexTasksRequest) (*GetTimerIndexTasksResponse, error)
		CompleteTimerTask(ctx context.Context, request *CompleteTimerTaskRequest) error
		RangeCompleteTimerTask(ctx context.Context, request *RangeCompleteTimerTaskRequest) error

		// Scan related methods
		ListConcreteExecutions(ctx context.Context, request *ListConcreteExecutionsRequest) (*InternalListConcreteExecutionsResponse, error)
		ListCurrentExecutions(ctx context.Context, request *ListCurrentExecutionsRequest) (*ListCurrentExecutionsResponse, error)
	}

	// HistoryStore is to manager workflow history events
	HistoryStore interface {
		Closeable
		GetName() string

		// The below are history V2 APIs
		// V2 regards history events growing as a tree, decoupled from workflow concepts

		// AppendHistoryNodes add(or override) a node to a history branch
		AppendHistoryNodes(ctx context.Context, request *InternalAppendHistoryNodesRequest) error
		// ReadHistoryBranch returns history node data for a branch
		ReadHistoryBranch(ctx context.Context, request *InternalReadHistoryBranchRequest) (*InternalReadHistoryBranchResponse, error)
		// ForkHistoryBranch forks a new branch from a old branch
		ForkHistoryBranch(ctx context.Context, request *InternalForkHistoryBranchRequest) (*InternalForkHistoryBranchResponse, error)
		// DeleteHistoryBranch removes a branch
		DeleteHistoryBranch(ctx context.Context, request *InternalDeleteHistoryBranchRequest) error
		// GetHistoryTree returns all branch information of a tree
		GetHistoryTree(ctx context.Context, request *InternalGetHistoryTreeRequest) (*InternalGetHistoryTreeResponse, error)
		// GetAllHistoryTreeBranches returns all branches of all trees
		GetAllHistoryTreeBranches(ctx context.Context, request *GetAllHistoryTreeBranchesRequest) (*GetAllHistoryTreeBranchesResponse, error)
	}

	// VisibilityStore is the store interface for visibility
	VisibilityStore interface {
		Closeable
		GetName() string
		RecordWorkflowExecutionStarted(ctx context.Context, request *InternalRecordWorkflowExecutionStartedRequest) error
		RecordWorkflowExecutionClosed(ctx context.Context, request *InternalRecordWorkflowExecutionClosedRequest) error
		UpsertWorkflowExecution(ctx context.Context, request *InternalUpsertWorkflowExecutionRequest) error
		ListOpenWorkflowExecutions(ctx context.Context, request *InternalListWorkflowExecutionsRequest) (*InternalListWorkflowExecutionsResponse, error)
		ListClosedWorkflowExecutions(ctx context.Context, request *InternalListWorkflowExecutionsRequest) (*InternalListWorkflowExecutionsResponse, error)
		ListOpenWorkflowExecutionsByType(ctx context.Context, request *InternalListWorkflowExecutionsByTypeRequest) (*InternalListWorkflowExecutionsResponse, error)
		ListClosedWorkflowExecutionsByType(ctx context.Context, request *InternalListWorkflowExecutionsByTypeRequest) (*InternalListWorkflowExecutionsResponse, error)
		ListOpenWorkflowExecutionsByWorkflowID(ctx context.Context, request *InternalListWorkflowExecutionsByWorkflowIDRequest) (*InternalListWorkflowExecutionsResponse, error)
		ListClosedWorkflowExecutionsByWorkflowID(ctx context.Context, request *InternalListWorkflowExecutionsByWorkflowIDRequest) (*InternalListWorkflowExecutionsResponse, error)
		ListClosedWorkflowExecutionsByStatus(ctx context.Context, request *InternalListClosedWorkflowExecutionsByStatusRequest) (*InternalListWorkflowExecutionsResponse, error)
		GetClosedWorkflowExecution(ctx context.Context, request *InternalGetClosedWorkflowExecutionRequest) (*InternalGetClosedWorkflowExecutionResponse, error)
		DeleteWorkflowExecution(ctx context.Context, request *VisibilityDeleteWorkflowExecutionRequest) error
		ListWorkflowExecutions(ctx context.Context, request *ListWorkflowExecutionsByQueryRequest) (*InternalListWorkflowExecutionsResponse, error)
		ScanWorkflowExecutions(ctx context.Context, request *ListWorkflowExecutionsByQueryRequest) (*InternalListWorkflowExecutionsResponse, error)
		CountWorkflowExecutions(ctx context.Context, request *CountWorkflowExecutionsRequest) (*CountWorkflowExecutionsResponse, error)
	}

	// Queue is a store to enqueue and get messages
	Queue interface {
		Closeable
		EnqueueMessage(ctx context.Context, messagePayload []byte) error
		ReadMessages(ctx context.Context, lastMessageID int64, maxCount int) ([]*InternalQueueMessage, error)
		DeleteMessagesBefore(ctx context.Context, messageID int64) error
		UpdateAckLevel(ctx context.Context, messageID int64, clusterName string) error
		GetAckLevels(ctx context.Context) (map[string]int64, error)
		EnqueueMessageToDLQ(ctx context.Context, messagePayload []byte) (int64, error)
		ReadMessagesFromDLQ(ctx context.Context, firstMessageID int64, lastMessageID int64, pageSize int, pageToken []byte) ([]*InternalQueueMessage, []byte, error)
		DeleteMessageFromDLQ(ctx context.Context, messageID int64) error
		RangeDeleteMessagesFromDLQ(ctx context.Context, firstMessageID int64, lastMessageID int64) error
		UpdateDLQAckLevel(ctx context.Context, messageID int64, clusterName string) error
		GetDLQAckLevels(ctx context.Context) (map[string]int64, error)
	}

	// InternalQueueMessage is the message that stores in the queue
	InternalQueueMessage struct {
		ID        int64     `json:"message_id"`
		QueueType QueueType `json:"queue_type"`
		Payload   []byte    `json:"message_payload"`
	}

	// DataBlob represents a blob for any binary data.
	// It contains raw data, and metadata(right now only encoding) in other field
	// Note that it should be only used for Persistence layer, below dataInterface and application(historyEngine/etc)
	DataBlob struct {
		Encoding common.EncodingType
		Data     []byte
	}

	// InternalCreateWorkflowExecutionRequest is used to write a new workflow execution
	InternalCreateWorkflowExecutionRequest struct {
		RangeID int64

		Mode CreateWorkflowMode

		PreviousRunID            string
		PreviousLastWriteVersion int64

		NewWorkflowSnapshot InternalWorkflowSnapshot
	}

	// InternalGetReplicationTasksResponse is the response to GetReplicationTask
	InternalGetReplicationTasksResponse struct {
		Tasks         []*InternalReplicationTaskInfo
		NextPageToken []byte
	}

	// InternalPutReplicationTaskToDLQRequest is used to put a replication task to dlq
	InternalPutReplicationTaskToDLQRequest struct {
		SourceClusterName string
		TaskInfo          *InternalReplicationTaskInfo
	}

	// InternalGetReplicationTasksFromDLQResponse is the response for GetReplicationTasksFromDLQ
	InternalGetReplicationTasksFromDLQResponse = InternalGetReplicationTasksResponse

	// InternalReplicationTaskInfo describes the replication task created for replication of history events
	InternalReplicationTaskInfo struct {
		DomainID          string
		WorkflowID        string
		RunID             string
		TaskID            int64
		TaskType          int
		FirstEventID      int64
		NextEventID       int64
		Version           int64
		ScheduledID       int64
		BranchToken       []byte
		NewRunBranchToken []byte
		CreationTime      int64
	}

	// InternalWorkflowExecutionInfo describes a workflow execution for Persistence Interface
	InternalWorkflowExecutionInfo struct {
		DomainID                           string
		WorkflowID                         string
		RunID                              string
		ParentDomainID                     string
		ParentWorkflowID                   string
		ParentRunID                        string
		InitiatedID                        int64
		CompletionEventBatchID             int64
		CompletionEvent                    *DataBlob
		TaskList                           string
		WorkflowTypeName                   string
		WorkflowTimeout                    int32
		DecisionStartToCloseTimeout        int32
		ExecutionContext                   []byte
		State                              int
		CloseStatus                        int
		LastFirstEventID                   int64
		LastEventTaskID                    int64
		NextEventID                        int64
		LastProcessedEvent                 int64
		StartTimestamp                     time.Time
		LastUpdatedTimestamp               time.Time
		CreateRequestID                    string
		SignalCount                        int32
		DecisionVersion                    int64
		DecisionScheduleID                 int64
		DecisionStartedID                  int64
		DecisionRequestID                  string
		DecisionTimeout                    int32
		DecisionAttempt                    int64
		DecisionStartedTimestamp           int64
		DecisionScheduledTimestamp         int64
		DecisionOriginalScheduledTimestamp int64
		CancelRequested                    bool
		CancelRequestID                    string
		StickyTaskList                     string
		StickyScheduleToStartTimeout       int32
		ClientLibraryVersion               string
		ClientFeatureVersion               string
		ClientImpl                         string
		AutoResetPoints                    *DataBlob
		// for retry
		Attempt            int32
		HasRetryPolicy     bool
		InitialInterval    int32
		BackoffCoefficient float64
		MaximumInterval    int32
		ExpirationTime     time.Time
		MaximumAttempts    int32
		NonRetriableErrors []string
		BranchToken        []byte
		CronSchedule       string
		ExpirationSeconds  int32
		Memo               map[string][]byte
		SearchAttributes   map[string][]byte

		// attributes which are not related to mutable state at all
		HistorySize int64
	}

	// InternalWorkflowMutableState indicates workflow related state for Persistence Interface
	InternalWorkflowMutableState struct {
		ExecutionInfo    *InternalWorkflowExecutionInfo
		VersionHistories *DataBlob
		ReplicationState *ReplicationState // TODO: remove this after all 2DC workflows complete
		ActivityInfos    map[int64]*InternalActivityInfo

		TimerInfos          map[string]*TimerInfo
		ChildExecutionInfos map[int64]*InternalChildExecutionInfo
		RequestCancelInfos  map[int64]*RequestCancelInfo
		SignalInfos         map[int64]*SignalInfo
		SignalRequestedIDs  map[string]struct{}
		BufferedEvents      []*DataBlob

		Checksum checksum.Checksum
	}

	// InternalActivityInfo details  for Persistence Interface
	InternalActivityInfo struct {
		Version                  int64
		ScheduleID               int64
		ScheduledEventBatchID    int64
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
		LastFailureReason  string
		LastWorkerIdentity string
		LastFailureDetails []byte
		// Not written to database - This is used only for deduping heartbeat timer creation
		LastHeartbeatTimeoutVisibilityInSeconds int64
	}

	// InternalChildExecutionInfo has details for pending child executions for Persistence Interface
	InternalChildExecutionInfo struct {
		Version               int64
		InitiatedID           int64
		InitiatedEventBatchID int64
		InitiatedEvent        *DataBlob
		StartedID             int64
		StartedWorkflowID     string
		StartedRunID          string
		StartedEvent          *DataBlob
		CreateRequestID       string
		DomainName            string
		WorkflowTypeName      string
		ParentClosePolicy     workflow.ParentClosePolicy
	}

	// InternalUpdateWorkflowExecutionRequest is used to update a workflow execution for Persistence Interface
	InternalUpdateWorkflowExecutionRequest struct {
		RangeID int64

		Mode UpdateWorkflowMode

		UpdateWorkflowMutation InternalWorkflowMutation

		NewWorkflowSnapshot *InternalWorkflowSnapshot
	}

	// InternalConflictResolveWorkflowExecutionRequest is used to reset workflow execution state for Persistence Interface
	InternalConflictResolveWorkflowExecutionRequest struct {
		RangeID int64

		Mode ConflictResolveWorkflowMode

		// workflow to be resetted
		ResetWorkflowSnapshot InternalWorkflowSnapshot

		// maybe new workflow
		NewWorkflowSnapshot *InternalWorkflowSnapshot

		// current workflow
		CurrentWorkflowMutation *InternalWorkflowMutation
	}

	// InternalResetWorkflowExecutionRequest is used to reset workflow execution state for Persistence Interface
	InternalResetWorkflowExecutionRequest struct {
		RangeID int64

		// for base run (we need to make sure the baseRun hasn't been deleted after forking)
		BaseRunID          string
		BaseRunNextEventID int64

		// for current workflow record
		CurrentRunID          string
		CurrentRunNextEventID int64

		// for current mutable state
		CurrentWorkflowMutation *InternalWorkflowMutation

		// For new mutable state
		NewWorkflowSnapshot InternalWorkflowSnapshot
	}

	// InternalWorkflowMutation is used as generic workflow execution state mutation for Persistence Interface
	InternalWorkflowMutation struct {
		ExecutionInfo    *InternalWorkflowExecutionInfo
		VersionHistories *DataBlob
		StartVersion     int64
		LastWriteVersion int64

		UpsertActivityInfos       []*InternalActivityInfo
		DeleteActivityInfos       []int64
		UpsertTimerInfos          []*TimerInfo
		DeleteTimerInfos          []string
		UpsertChildExecutionInfos []*InternalChildExecutionInfo
		DeleteChildExecutionInfo  *int64
		UpsertRequestCancelInfos  []*RequestCancelInfo
		DeleteRequestCancelInfo   *int64
		UpsertSignalInfos         []*SignalInfo
		DeleteSignalInfo          *int64
		UpsertSignalRequestedIDs  []string
		DeleteSignalRequestedID   string
		NewBufferedEvents         *DataBlob
		ClearBufferedEvents       bool

		TransferTasks    []Task
		TimerTasks       []Task
		ReplicationTasks []Task

		Condition int64

		Checksum checksum.Checksum
	}

	// InternalWorkflowSnapshot is used as generic workflow execution state snapshot for Persistence Interface
	InternalWorkflowSnapshot struct {
		ExecutionInfo    *InternalWorkflowExecutionInfo
		VersionHistories *DataBlob
		StartVersion     int64
		LastWriteVersion int64

		ActivityInfos       []*InternalActivityInfo
		TimerInfos          []*TimerInfo
		ChildExecutionInfos []*InternalChildExecutionInfo
		RequestCancelInfos  []*RequestCancelInfo
		SignalInfos         []*SignalInfo
		SignalRequestedIDs  []string

		TransferTasks    []Task
		TimerTasks       []Task
		ReplicationTasks []Task

		Condition int64

		Checksum checksum.Checksum
	}

	// InternalAppendHistoryEventsRequest is used to append new events to workflow execution history  for Persistence Interface
	InternalAppendHistoryEventsRequest struct {
		DomainID          string
		Execution         workflow.WorkflowExecution
		FirstEventID      int64
		EventBatchVersion int64
		RangeID           int64
		TransactionID     int64
		Events            *DataBlob
		Overwrite         bool
	}

	// InternalAppendHistoryNodesRequest is used to append a batch of history nodes
	InternalAppendHistoryNodesRequest struct {
		// True if it is the first append request to the branch
		IsNewBranch bool
		// The info for clean up data in background
		Info string
		// The branch to be appended
		BranchInfo types.HistoryBranch
		// The first eventID becomes the nodeID to be appended
		NodeID int64
		// The events to be appended
		Events *DataBlob
		// Requested TransactionID for conditional update
		TransactionID int64
		// Used in sharded data stores to identify which shard to use
		ShardID int
	}

	// InternalGetWorkflowExecutionRequest is used to retrieve the info of a workflow execution
	InternalGetWorkflowExecutionRequest struct {
		DomainID  string
		Execution workflow.WorkflowExecution
	}

	// InternalGetWorkflowExecutionResponse is the response to GetWorkflowExecution for Persistence Interface
	InternalGetWorkflowExecutionResponse struct {
		State *InternalWorkflowMutableState
	}

	// InternalListConcreteExecutionsResponse is the response to ListConcreteExecutions for Persistence Interface
	InternalListConcreteExecutionsResponse struct {
		Executions    []*InternalListConcreteExecutionsEntity
		NextPageToken []byte
	}

	// InternalListConcreteExecutionsEntity is a single entity in InternalListConcreteExecutionsResponse
	InternalListConcreteExecutionsEntity struct {
		ExecutionInfo    *InternalWorkflowExecutionInfo
		VersionHistories *DataBlob
	}

	// InternalForkHistoryBranchRequest is used to fork a history branch
	InternalForkHistoryBranchRequest struct {
		// The base branch to fork from
		ForkBranchInfo types.HistoryBranch
		// The nodeID to fork from, the new branch will start from ( inclusive ), the base branch will stop at(exclusive)
		ForkNodeID int64
		// branchID of the new branch
		NewBranchID string
		// the info for clean up data in background
		Info string
		// Used in sharded data stores to identify which shard to use
		ShardID int
	}

	// InternalForkHistoryBranchResponse is the response to ForkHistoryBranchRequest
	InternalForkHistoryBranchResponse struct {
		// branchInfo to represent the new branch
		NewBranchInfo types.HistoryBranch
	}

	// InternalDeleteHistoryBranchRequest is used to remove a history branch
	InternalDeleteHistoryBranchRequest struct {
		// branch to be deleted
		BranchInfo types.HistoryBranch
		// Used in sharded data stores to identify which shard to use
		ShardID int
	}

	// InternalReadHistoryBranchRequest is used to read a history branch
	InternalReadHistoryBranchRequest struct {
		// The tree of branch range to be read
		TreeID string
		// The branch range to be read
		BranchID string
		// Get the history nodes from MinNodeID. Inclusive.
		MinNodeID int64
		// Get the history nodes upto MaxNodeID.  Exclusive.
		MaxNodeID int64
		// passing thru for pagination
		PageSize int
		// Pagination token
		NextPageToken []byte
		// LastNodeID is the last known node ID attached to a history node
		LastNodeID int64
		// LastTransactionID is the last known transaction ID attached to a history node
		LastTransactionID int64
		// Used in sharded data stores to identify which shard to use
		ShardID int
	}

	// InternalCompleteForkBranchRequest is used to update some tree/branch meta data for forking
	InternalCompleteForkBranchRequest struct {
		// branch to be updated
		BranchInfo workflow.HistoryBranch
		// whether fork is successful
		Success bool
		// Used in sharded data stores to identify which shard to use
		ShardID int
	}

	// InternalReadHistoryBranchResponse is the response to ReadHistoryBranchRequest
	InternalReadHistoryBranchResponse struct {
		// History events
		History []*DataBlob
		// Pagination token
		NextPageToken []byte
		// LastNodeID is the last known node ID attached to a history node
		LastNodeID int64
		// LastTransactionID is the last known transaction ID attached to a history node
		LastTransactionID int64
	}

	// InternalGetHistoryTreeRequest is used to get history tree
	InternalGetHistoryTreeRequest struct {
		// A UUID of a tree
		TreeID string
		// Get data from this shard
		ShardID *int
		// optional: can provide treeID via branchToken if treeID is empty
		BranchToken []byte
	}

	// InternalGetHistoryTreeResponse is the response to GetHistoryTree
	InternalGetHistoryTreeResponse struct {
		// all branches of a tree
		Branches []*types.HistoryBranch
	}

	// InternalVisibilityWorkflowExecutionInfo is visibility info for internal response
	InternalVisibilityWorkflowExecutionInfo struct {
		WorkflowID       string
		RunID            string
		TypeName         string
		StartTime        time.Time
		ExecutionTime    time.Time
		CloseTime        time.Time
		Status           *types.WorkflowExecutionCloseStatus
		HistoryLength    int64
		Memo             *DataBlob
		TaskList         string
		SearchAttributes map[string]interface{}
	}

	// InternalListWorkflowExecutionsResponse is response from ListWorkflowExecutions
	InternalListWorkflowExecutionsResponse struct {
		Executions []*InternalVisibilityWorkflowExecutionInfo
		// Token to read next page if there are more workflow executions beyond page size.
		// Use this to set NextPageToken on ListWorkflowExecutionsRequest to read the next page.
		NextPageToken []byte
	}

	// InternalGetClosedWorkflowExecutionRequest is used retrieve the record for a specific execution
	InternalGetClosedWorkflowExecutionRequest struct {
		DomainUUID string
		Domain     string // domain name is not persisted, but used as config filter key
		Execution  types.WorkflowExecution
	}

	// InternalListClosedWorkflowExecutionsByStatusRequest is used to list executions that have specific close status
	InternalListClosedWorkflowExecutionsByStatusRequest struct {
		InternalListWorkflowExecutionsRequest
		Status types.WorkflowExecutionCloseStatus
	}

	// InternalListWorkflowExecutionsByWorkflowIDRequest is used to list executions that have specific WorkflowID in a domain
	InternalListWorkflowExecutionsByWorkflowIDRequest struct {
		InternalListWorkflowExecutionsRequest
		WorkflowID string
	}

	// InternalListWorkflowExecutionsByTypeRequest is used to list executions of a specific type in a domain
	InternalListWorkflowExecutionsByTypeRequest struct {
		InternalListWorkflowExecutionsRequest
		WorkflowTypeName string
	}

	// InternalGetClosedWorkflowExecutionResponse is response from GetWorkflowExecution
	InternalGetClosedWorkflowExecutionResponse struct {
		Execution *InternalVisibilityWorkflowExecutionInfo
	}

	// InternalRecordWorkflowExecutionStartedRequest request to RecordWorkflowExecutionStarted
	InternalRecordWorkflowExecutionStartedRequest struct {
		DomainUUID         string
		WorkflowID         string
		RunID              string
		WorkflowTypeName   string
		StartTimestamp     int64
		ExecutionTimestamp int64
		WorkflowTimeout    int64
		TaskID             int64
		Memo               *DataBlob
		TaskList           string
		SearchAttributes   map[string][]byte
	}

	// InternalRecordWorkflowExecutionClosedRequest is request to RecordWorkflowExecutionClosed
	InternalRecordWorkflowExecutionClosedRequest struct {
		DomainUUID         string
		WorkflowID         string
		RunID              string
		WorkflowTypeName   string
		StartTimestamp     int64
		ExecutionTimestamp int64
		TaskID             int64
		Memo               *DataBlob
		TaskList           string
		SearchAttributes   map[string][]byte
		CloseTimestamp     int64
		Status             types.WorkflowExecutionCloseStatus
		HistoryLength      int64
		RetentionSeconds   int64
	}

	// InternalUpsertWorkflowExecutionRequest is request to UpsertWorkflowExecution
	InternalUpsertWorkflowExecutionRequest struct {
		DomainUUID         string
		WorkflowID         string
		RunID              string
		WorkflowTypeName   string
		StartTimestamp     int64
		ExecutionTimestamp int64
		WorkflowTimeout    int64
		TaskID             int64
		Memo               *DataBlob
		TaskList           string
		SearchAttributes   map[string][]byte
	}

	// InternalListWorkflowExecutionsRequest is used to list executions in a domain
	InternalListWorkflowExecutionsRequest struct {
		DomainUUID string
		Domain     string // domain name is not persisted, but used as config filter key
		// The earliest end of the time range
		EarliestTime int64
		// The latest end of the time range
		LatestTime int64
		// Maximum number of workflow executions per page
		PageSize int
		// Token to continue reading next page of workflow executions.
		// Pass in empty slice for first page.
		NextPageToken []byte
	}

	// InternalDomainConfig describes the domain configuration
	InternalDomainConfig struct {
		Retention                time.Duration
		EmitMetric               bool                 // deprecated
		ArchivalBucket           string               // deprecated
		ArchivalStatus           types.ArchivalStatus // deprecated
		HistoryArchivalStatus    types.ArchivalStatus
		HistoryArchivalURI       string
		VisibilityArchivalStatus types.ArchivalStatus
		VisibilityArchivalURI    string
		BadBinaries              *DataBlob
	}

	// InternalCreateDomainRequest is used to create the domain
	InternalCreateDomainRequest struct {
		Info              *DomainInfo
		Config            *InternalDomainConfig
		ReplicationConfig *DomainReplicationConfig
		IsGlobalDomain    bool
		ConfigVersion     int64
		FailoverVersion   int64
		LastUpdatedTime   time.Time
	}

	// InternalGetDomainResponse is the response for GetDomain
	InternalGetDomainResponse struct {
		Info                        *DomainInfo
		Config                      *InternalDomainConfig
		ReplicationConfig           *DomainReplicationConfig
		IsGlobalDomain              bool
		ConfigVersion               int64
		FailoverVersion             int64
		FailoverNotificationVersion int64
		PreviousFailoverVersion     int64
		FailoverEndTime             *time.Time
		LastUpdatedTime             time.Time
		NotificationVersion         int64
	}

	// InternalUpdateDomainRequest is used to update domain
	InternalUpdateDomainRequest struct {
		Info                        *DomainInfo
		Config                      *InternalDomainConfig
		ReplicationConfig           *DomainReplicationConfig
		ConfigVersion               int64
		FailoverVersion             int64
		FailoverNotificationVersion int64
		PreviousFailoverVersion     int64
		FailoverEndTime             *time.Time
		LastUpdatedTime             time.Time
		NotificationVersion         int64
	}

	// InternalListDomainsResponse is the response for GetDomain
	InternalListDomainsResponse struct {
		Domains       []*InternalGetDomainResponse
		NextPageToken []byte
	}

	// InternalShardInfo describes a shard
	InternalShardInfo struct {
		ShardID                       int                              `json:"shard_id"`
		Owner                         string                           `json:"owner"`
		RangeID                       int64                            `json:"range_id"`
		StolenSinceRenew              int                              `json:"stolen_since_renew"`
		UpdatedAt                     time.Time                        `json:"updated_at"`
		ReplicationAckLevel           int64                            `json:"replication_ack_level"`
		ReplicationDLQAckLevel        map[string]int64                 `json:"replication_dlq_ack_level"`
		TransferAckLevel              int64                            `json:"transfer_ack_level"`
		TimerAckLevel                 time.Time                        `json:"timer_ack_level"`
		ClusterTransferAckLevel       map[string]int64                 `json:"cluster_transfer_ack_level"`
		ClusterTimerAckLevel          map[string]time.Time             `json:"cluster_timer_ack_level"`
		TransferProcessingQueueStates *DataBlob                        `json:"transfer_processing_queue_states"`
		TimerProcessingQueueStates    *DataBlob                        `json:"timer_processing_queue_states"`
		TransferFailoverLevels        map[string]TransferFailoverLevel // uuid -> TransferFailoverLevel
		TimerFailoverLevels           map[string]TimerFailoverLevel    // uuid -> TimerFailoverLevel
		ClusterReplicationLevel       map[string]int64                 `json:"cluster_replication_level"`
		DomainNotificationVersion     int64                            `json:"domain_notification_version"`
		PendingFailoverMarkers        *DataBlob                        `json:"pending_failover_markers"`
	}

	// InternalCreateShardRequest is request to CreateShard
	InternalCreateShardRequest struct {
		ShardInfo *InternalShardInfo
	}

	// InternalGetShardRequest is used to get shard information
	InternalGetShardRequest struct {
		ShardID int
	}

	// InternalUpdateShardRequest  is used to update shard information
	InternalUpdateShardRequest struct {
		ShardInfo       *InternalShardInfo
		PreviousRangeID int64
	}

	// InternalGetShardResponse is the response to GetShard
	InternalGetShardResponse struct {
		ShardInfo *InternalShardInfo
	}

	// InternalTaskInfo describes a Task
	InternalTaskInfo struct {
		DomainID               string
		WorkflowID             string
		RunID                  string
		TaskID                 int64
		ScheduleID             int64
		ScheduleToStartTimeout int32
		Expiry                 time.Time
		CreatedTime            time.Time
	}

	// InternalCreateTasksInfo describes a task to be created in InternalCreateTasksRequest
	InternalCreateTasksInfo struct {
		Execution types.WorkflowExecution
		Data      *InternalTaskInfo
		TaskID    int64
	}

	// InternalCreateTasksRequest is request to CreateTasks
	InternalCreateTasksRequest struct {
		TaskListInfo *TaskListInfo
		Tasks        []*InternalCreateTasksInfo
	}

	// InternalGetTasksResponse is response from GetTasks
	InternalGetTasksResponse struct {
		Tasks []*InternalTaskInfo
	}
)

// NewDataBlob returns a new DataBlob
func NewDataBlob(data []byte, encodingType common.EncodingType) *DataBlob {
	if data == nil || len(data) == 0 {
		return nil
	}
	if encodingType != "thriftrw" && data[0] == 'Y' {
		panic(fmt.Sprintf("Invalid incoding: \"%v\"", encodingType))
	}
	return &DataBlob{
		Data:     data,
		Encoding: encodingType,
	}
}

// FromDataBlob decodes a datablob into a (payload, encodingType) tuple
func FromDataBlob(blob *DataBlob) ([]byte, string) {
	if blob == nil || len(blob.Data) == 0 {
		return nil, ""
	}
	return blob.Data, string(blob.Encoding)
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
	case common.EncodingTypeEmpty:
		return common.EncodingTypeEmpty
	default:
		return common.EncodingTypeUnknown
	}
}

// ToThrift convert data blob to thrift representation
func (d *DataBlob) ToThrift() *workflow.DataBlob {
	switch d.Encoding {
	case common.EncodingTypeJSON:
		return &workflow.DataBlob{
			EncodingType: workflow.EncodingTypeJSON.Ptr(),
			Data:         d.Data,
		}
	case common.EncodingTypeThriftRW:
		return &workflow.DataBlob{
			EncodingType: workflow.EncodingTypeThriftRW.Ptr(),
			Data:         d.Data,
		}
	default:
		panic(fmt.Sprintf("DataBlob seeing unsupported enconding type: %v", d.Encoding))
	}
}

// NewDataBlobFromThrift convert data blob from thrift representation
func NewDataBlobFromThrift(blob *workflow.DataBlob) *DataBlob {
	switch blob.GetEncodingType() {
	case workflow.EncodingTypeJSON:
		return &DataBlob{
			Encoding: common.EncodingTypeJSON,
			Data:     blob.Data,
		}
	case workflow.EncodingTypeThriftRW:
		return &DataBlob{
			Encoding: common.EncodingTypeThriftRW,
			Data:     blob.Data,
		}
	default:
		panic(fmt.Sprintf("NewDataBlobFromThrift seeing unsupported enconding type: %v", blob.GetEncodingType()))
	}
}
