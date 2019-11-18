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

package sqlshared

import (
	"database/sql"

	"github.com/uber/cadence/common/persistence/sql/storage/sqldb"
)

// For history_node table:

// InsertIntoHistoryNode inserts a row into history_node table
func (mdb *DB) InsertIntoHistoryNode(row *sqldb.HistoryNodeRow) (sql.Result, error) {
	// NOTE: MySQL 5.6 doesn't support clustering order, to workaround, we let txn_id multiple by -1
	*row.TxnID *= -1
	return mdb.conn.NamedExec(mdb.driver.AddHistoryNodesQuery(), row)
}

// SelectFromHistoryNode reads one or more rows from history_node table
func (mdb *DB) SelectFromHistoryNode(filter *sqldb.HistoryNodeFilter) ([]sqldb.HistoryNodeRow, error) {
	var rows []sqldb.HistoryNodeRow
	err := mdb.conn.Select(&rows, mdb.driver.GetHistoryNodesQuery(),
		filter.ShardID, filter.TreeID, filter.BranchID, *filter.MinNodeID, *filter.MaxNodeID, *filter.PageSize)
	// NOTE: since we let txn_id multiple by -1 when inserting, we have to revert it back here
	for _, row := range rows {
		*row.TxnID *= -1
	}
	return rows, err
}

// DeleteFromHistoryNode deletes one or more rows from history_node table
func (mdb *DB) DeleteFromHistoryNode(filter *sqldb.HistoryNodeFilter) (sql.Result, error) {
	return mdb.conn.Exec(mdb.driver.DeleteHistoryNodesQuery(), filter.ShardID, filter.TreeID, filter.BranchID, *filter.MinNodeID)
}

// For history_tree table:

// InsertIntoHistoryTree inserts a row into history_tree table
func (mdb *DB) InsertIntoHistoryTree(row *sqldb.HistoryTreeRow) (sql.Result, error) {
	return mdb.conn.NamedExec(mdb.driver.AddHistoryTreeQuery(), row)
}

// SelectFromHistoryTree reads one or more rows from history_tree table
func (mdb *DB) SelectFromHistoryTree(filter *sqldb.HistoryTreeFilter) ([]sqldb.HistoryTreeRow, error) {
	var rows []sqldb.HistoryTreeRow
	err := mdb.conn.Select(&rows, mdb.driver.GetHistoryTreeQuery(), filter.ShardID, filter.TreeID)
	return rows, err
}

// DeleteFromHistoryTree deletes one or more rows from history_tree table
func (mdb *DB) DeleteFromHistoryTree(filter *sqldb.HistoryTreeFilter) (sql.Result, error) {
	return mdb.conn.Exec(mdb.driver.DeleteHistoryTreeQuery(), filter.ShardID, filter.TreeID, *filter.BranchID)
}
