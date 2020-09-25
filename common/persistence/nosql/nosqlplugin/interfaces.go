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
	"github.com/uber/cadence/.gen/go/shared"
)

type (
	// DB defines the API for regular NoSQL operations of a Cadence server
	DB interface {
		PluginName() string
		IsNotFoundError(err error) bool
		IsTimeoutError(err error) bool
		IsThrottlingError(err error) bool
		Close()

		tableCRUD
	}
	// tableCRUD defines the API for interacting with the database tables
	tableCRUD interface {
		historyEventsCRUD
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
		ShardID                     int
		TreeID                      string
		BranchID                    string
		Ancestors                   []*shared.HistoryBranchRange
		CreateTimestampMilliseconds int64
		Info                        string
	}

	// HistoryTreeFilter contains the column names within history_tree table that
	// can be used to filter results through a WHERE clause
	HistoryTreeFilter struct {
		ShardID  int
		TreeID   string
		BranchID *string
	}
)
