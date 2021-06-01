// Copyright (c) 2020 Uber Technologies, Inc.
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
	"context"
	"time"

	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

type (
	// DB defines the API for regular NoSQL operations of a Cadence server
	DB interface {
		PluginName() string
		Close()

		NoSQLErrorChecker
		tableCRUD
	}
	// tableCRUD defines the API for interacting with the database tables
	// NOTE 1: All SELECT interfaces require strong consistency (eventual consistency will not work) unless specify in the method.
	//
	// NOTE 2: About schema: only the columns that need to be used directly in queries are considered 'significant',
	// including partition key, range key or index key and conditional columns, are required to be in the schema.
	// All other non 'significant' columns are opaque for implementation.
	// Therefore, it's recommended to use a data blob to store all the other columns, so that adding new
	// column will not require schema changes. This approach has been proved very successful in MySQL/Postgres implementation of SQL interfaces.
	// Cassandra implementation cannot do it due to backward-compatibility. Any other NoSQL implementation should use datablob for non-significant columns.
	// Follow the comment for each tableCRUD for what are 'significant' columns.
	tableCRUD interface {
		historyEventsCRUD
		messageQueueCRUD
		domainCRUD
		shardCRUD
		visibilityCRUD
	}

	// NoSQLErrorChecker checks for common nosql errors
	NoSQLErrorChecker interface {
		IsConditionFailedError(err error) bool
		ClientErrorChecker
	}

	// ClientErrorChecker checks for common nosql errors on client
	ClientErrorChecker interface {
		IsTimeoutError(error) bool
		IsNotFoundError(error) bool
		IsThrottlingError(error) bool
	}

	/**
	 * historyEventsCRUD is for History events storage system
	 * Recommendation: use two tables: history_tree for branch records and history_node for node records
	 * if a single update query can operate on two tables.
	 *
	 * Significant columns:
	 * history_tree partition key: (shardID, treeID), range key: (branchID)
	 * history_node partition key: (shardID, treeID), range key: (branchID, nodeID ASC, txnID DESC)
	 */
	historyEventsCRUD interface {
		// InsertIntoHistoryTreeAndNode inserts one or two rows: tree row and node row(at least one of them)
		InsertIntoHistoryTreeAndNode(ctx context.Context, treeRow *HistoryTreeRow, nodeRow *HistoryNodeRow) error

		// SelectFromHistoryNode read nodes based on a filter
		SelectFromHistoryNode(ctx context.Context, filter *HistoryNodeFilter) ([]*HistoryNodeRow, []byte, error)

		// DeleteFromHistoryTreeAndNode delete a branch record, and a list of ranges of nodes.
		// for each range, it will delete all nodes starting from MinNodeID(inclusive)
		DeleteFromHistoryTreeAndNode(ctx context.Context, treeFilter *HistoryTreeFilter, nodeFilters []*HistoryNodeFilter) error

		// SelectAllHistoryTrees will return all tree branches with pagination
		SelectAllHistoryTrees(ctx context.Context, nextPageToken []byte, pageSize int) ([]*HistoryTreeRow, []byte, error)

		// SelectFromHistoryTree read branch records for a tree.
		// It returns without pagination, because we assume one tree won't have too many branches.
		SelectFromHistoryTree(ctx context.Context, filter *HistoryTreeFilter) ([]*HistoryTreeRow, error)
	}

	/***
	 * messageQueueCRUD is for the message queue storage system
	 *
	 * Recommendation: use two tables(queue_message,and queue_metadata) to implement this interface
	 *
	 * Significant columns:
	 * queue_message partition key: (queueType), range key: (messageID)
	 * queue_metadata partition key: (queueType), range key: N/A, query condition column(version)
	 */
	messageQueueCRUD interface {
		//Insert message into queue, return error if failed or already exists
		// Must return conditionFailed error if row already exists
		InsertIntoQueue(ctx context.Context, row *QueueMessageRow) error
		// Get the ID of last message inserted into the queue
		SelectLastEnqueuedMessageID(ctx context.Context, queueType persistence.QueueType) (int64, error)
		// Read queue messages starting from the exclusiveBeginMessageID
		SelectMessagesFrom(ctx context.Context, queueType persistence.QueueType, exclusiveBeginMessageID int64, maxRows int) ([]*QueueMessageRow, error)
		// Read queue message starting from exclusiveBeginMessageID int64, inclusiveEndMessageID int64
		SelectMessagesBetween(ctx context.Context, request SelectMessagesBetweenRequest) (*SelectMessagesBetweenResponse, error)
		// Delete all messages before exclusiveBeginMessageID
		DeleteMessagesBefore(ctx context.Context, queueType persistence.QueueType, exclusiveBeginMessageID int64) error
		// Delete all messages in a range between exclusiveBeginMessageID and inclusiveEndMessageID
		DeleteMessagesInRange(ctx context.Context, queueType persistence.QueueType, exclusiveBeginMessageID int64, inclusiveEndMessageID int64) error
		// Delete one message
		DeleteMessage(ctx context.Context, queueType persistence.QueueType, messageID int64) error

		// Insert an empty metadata row, starting from a version
		InsertQueueMetadata(ctx context.Context, queueType persistence.QueueType, version int64) error
		// **Conditionally** update a queue metadata row, if current version is matched(meaning current == row.Version - 1),
		// then the current version will increase by one when updating the metadata row
		// Must return conditionFailed error if the condition is not met
		UpdateQueueMetadataCas(ctx context.Context, row QueueMetadataRow) error
		// Read a QueueMetadata
		SelectQueueMetadata(ctx context.Context, queueType persistence.QueueType) (*QueueMetadataRow, error)
		// GetQueueSize return the queue size
		GetQueueSize(ctx context.Context, queueType persistence.QueueType) (int64, error)
	}

	/***
	* domainCRUD is for domain + domain metadata storage system
	*
	* Recommendation: two tables(domain, domain_metadata) to implement if conditional updates on two tables is supported
	*
	* Significant columns:
	* domain: partition key( a constant value), range key(domainName), local secondary index(domainID)
	* domain_metadata: partition key( a constant value), range key(N/A), query condition column(notificationVersion)
	*
	* Note 1: About Cassandra's implementation: Because of historical reasons, Cassandra uses two table,
	* domains and domains_by_name_v2. Therefore, Cassandra implementation lost the atomicity causing some edge cases,
	* and the implementation is more complicated than it should be.
	*
	* Note 2: Cassandra doesn't support conditional updates on multiple tables. Hence the domain_metadata table is implemented
	* as a special record as "domain metadata". It is an integer number as notification version.
	* The main purpose of it is to notify clusters that there is some changes in domains, so domain cache needs to refresh.
	* It always increase by one, whenever a domain is updated or inserted.
	* Updating this failover metadata with domain insert/update needs to be atomic.
	* Because Batch LWTs is only allowed within one table and same partition.
	* The Cassandra implementation stores it in the same table as domain in domains_by_name_v2.
	*
	* Note 3: It's okay to use a constant value for partition key because domain table is serving very small volume of traffic.
	 */
	domainCRUD interface {
		// Insert a new record to domain, return error if failed or already exists
		// Must return conditionFailed error if domainName already exists
		InsertDomain(ctx context.Context, row *DomainRow) error
		// Update domain data
		UpdateDomain(ctx context.Context, row *DomainRow) error
		// Get one domain data, either by domainID or domainName
		SelectDomain(ctx context.Context, domainID *string, domainName *string) (*DomainRow, error)
		// Get all domain data
		SelectAllDomains(ctx context.Context, pageSize int, pageToken []byte) ([]*DomainRow, []byte, error)
		//  Delete a domain, either by domainID or domainName
		DeleteDomain(ctx context.Context, domainID *string, domainName *string) error
		// right now domain metadata is just an integer as notification version
		SelectDomainMetadata(ctx context.Context) (int64, error)
	}

	/**
	* shardCRUD is for shard storage of workflow execution.

	* Recommendation: use one table if database support batch conditional update on multiple tables, otherwise combine with workflowCRUD (likeCassandra)
	*
	* Significant columns:
	* domain: partition key(shardID), range key(N/A), local secondary index(domainID), query condition column(rangeID)
	*
	* Note 1: shard will be required to run conditional update with workflowCRUD. So in some nosql database like Cassandra,
	* shardCRUD and workflowCRUD must be implemented within the same table. Because Cassandra only allows LightWeight transaction
	* executed within a single table.
	* Note 2: unlike Cassandra, most NoSQL databases don't return the previous rows when conditional write fails. In this case,
	* an extra read query is needed to get the previous row.
	 */
	shardCRUD interface {
		// InsertShard creates a new shard.
		// Return error is there is any thing wrong
		// When error IsConditionFailedError, also return the row that doesn't meet the condition
		InsertShard(ctx context.Context, row *ShardRow) (previous *ConflictedShardRow, err error)
		// SelectShard gets a shard, rangeID is the current rangeID in shard row
		SelectShard(ctx context.Context, shardID int, currentClusterName string) (rangeID int64, shard *ShardRow, err error)
		// UpdateRangeID updates the rangeID
		// Return error is there is any thing wrong
		// When error IsConditionFailedError, also return the row that doesn't meet the condition
		UpdateRangeID(ctx context.Context, shardID int, rangeID int64, previousRangeID int64) (previous *ConflictedShardRow, err error)
		// UpdateShard updates a shard
		// Return error is there is any thing wrong
		// When error IsConditionFailedError, also return the row that doesn't meet the condition
		UpdateShard(ctx context.Context, row *ShardRow, previousRangeID int64) (previous *ConflictedShardRow, err error)
	}

	/**
	* visibilityCRUD is for visibility storage
	*
	* Recommendation: use one table with multiple indexes
	*
	* Significant columns:
	* domain: partition key(domainID), range key(workflowID, runID),
	*         local secondary index #1(startTime),
	*         local secondary index #2(closedTime),
	*         local secondary index #3(workflowType, startTime),
	*         local secondary index #4(workflowType, closedTime),
	*         local secondary index #5(workflowID, startTime),
	*         local secondary index #6(workflowID, closedTime),
	*         local secondary index #7(closeStatus, closedTime),
	*
	* NOTE 1: Cassandra implementation of visibility uses three tables: open_executions, closed_executions and closed_executions_v2,
	* because Cassandra doesn't support cross-partition indexing.
	* Records in open_executions and closed_executions are clustered by start_time. Records in  closed_executions_v2 are by close_time.
	* This optimizes the performance, but introduce a lot of complexity.
	* In some other databases, this may be be necessary. Please refer to MySQL/Postgres implementation which uses only
	* one table with multiple indexes.
	*
	* NOTE 2: TTL(time to live records) is for auto-deleting expired records in visibility. For databases that don't support TTL,
	* please implement DeleteVisibility method. If TTL is supported, then DeleteVisibility can be a noop.
	 */
	visibilityCRUD interface {
		InsertVisibility(ctx context.Context, ttlSeconds int64, row *VisibilityRowForInsert) error
		UpdateVisibility(ctx context.Context, ttlSeconds int64, row *VisibilityRowForUpdate) error
		SelectVisibility(ctx context.Context, filter *VisibilityFilter) (*SelectVisibilityResponse, error)
		DeleteVisibility(ctx context.Context, domainID, workflowID, runID string) error
		// TODO deprecated this in the future in favor of SelectVisibility
		// Special case: return nil,nil if not found(since we will deprecate it, it's not worth refactor to be consistent)
		SelectOneClosedWorkflow(ctx context.Context, domainID, workflowID, runID string) (*VisibilityRow, error)
	}

	VisibilityRowForInsert struct {
		VisibilityRow
		DomainID string
	}

	VisibilityRowForUpdate struct {
		VisibilityRow
		DomainID string
		// NOTE: this is only for some implementation (e.g. Cassandra) that uses multiple tables,
		// they needs to delete record from the open execution table. Ignore this field if not need it
		UpdateOpenToClose bool
		//  Similar as UpdateOpenToClose
		UpdateCloseToOpen bool
	}

	// TODO separate in the future when need it
	VisibilityRow = persistence.InternalVisibilityWorkflowExecutionInfo

	SelectVisibilityResponse struct {
		Executions    []*VisibilityRow
		NextPageToken []byte
	}

	// VisibilityFilter contains the column names within executions_visibility table that
	// can be used to filter results through a WHERE clause
	VisibilityFilter struct {
		ListRequest  persistence.InternalListWorkflowExecutionsRequest
		FilterType   VisibilityFilterType
		SortType     VisibilitySortType
		WorkflowType string
		WorkflowID   string
		CloseStatus  int32
	}

	VisibilityFilterType int
	VisibilitySortType   int

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
