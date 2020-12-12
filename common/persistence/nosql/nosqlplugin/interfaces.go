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
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
	"github.com/uber/cadence/common/types"
)

type (
	// DB defines the API for regular NoSQL operations of a Cadence server
	DB interface {
		PluginName() string
		Close()

		IsConditionFailedError(err error) bool
		gocql.ErrorChecker

		tableCRUD
	}
	// tableCRUD defines the API for interacting with the database tables
	// NOTE: All SELECT interfaces require strong consistency. Using eventual consistency will not work.
	tableCRUD interface {
		historyEventsCRUD
		messageQueueCRUD
		domainCRUD
	}

	// historyEventsCRUD is for History events storage system
	historyEventsCRUD interface {
		/**
		* It can be implemented with two tables(history_tree for branch records and history_node for node records)
		* ShardID is passed from application layer as the same shardID of the workflow. But it is not required for History events
		* to be in the same shard as workflows. The pro of being the same shard is that when one DB partition goes down, the impact is lower.
		* However, being in the same shard can cause some hot partition issue. Because sometimes history can grow very large, this could be worse.
		* Therefore, Cadence built-in Cassandra plugin doesn't take use of ShardID at all.
		**/

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

	// messageQueueCRUD is for the message queue storage system
	// Typically two tables(queue_message,and queue_metadata) are needed to implement this interface
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

	// domainCRUD is for domain + domain metadata storage system
	// Ideally a domain operation should be implemented in transaction to be atomic if using multiple tables.
	// For example, Cassandra uses two table, domains and domains_by_name_v2.
	// But it is okay if not, as the nosqlMetadataManager will handle the edge cases.
	//
	// However, there is a special record as "domain metadata". Right now it is an integer number as notification version.
	// The main purpose of it is to notify clusters that there is some changes in domains, so domain cache needs to refresh.
	// It always increase by one, whenever a domain is updated or inserted.
	// Updating this failover metadata with domain insert/update needs to be atomic.
	// Because Batch LWTs is only allowed within one table and same partition.
	// The Cassandra implementation stores it in the same table as domain in domains_by_name_v2.
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
