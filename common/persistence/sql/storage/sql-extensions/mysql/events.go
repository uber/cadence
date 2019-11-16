// Copyright (c) 2019 Uber Technologies, Inc.
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

package mysql

const (
    // below are templates for history_node table
    addHistoryNodesQry = `INSERT INTO history_node (` +
        `shard_id, tree_id, branch_id, node_id, txn_id, data, data_encoding) ` +
        `VALUES (:shard_id, :tree_id, :branch_id, :node_id, :txn_id, :data, :data_encoding) `

    getHistoryNodesQry = `SELECT node_id, txn_id, data, data_encoding FROM history_node ` +
        `WHERE shard_id = ? AND tree_id = ? AND branch_id = ? AND node_id >= ? and node_id < ? ORDER BY shard_id, tree_id, branch_id, node_id, txn_id LIMIT ? `

    deleteHistoryNodesQry = `DELETE FROM history_node WHERE shard_id = ? AND tree_id = ? AND branch_id = ? AND node_id >= ? `

    // below are templates for history_tree table
    addHistoryTreeQry = `INSERT INTO history_tree (` +
        `shard_id, tree_id, branch_id, data, data_encoding) ` +
        `VALUES (:shard_id, :tree_id, :branch_id, :data, :data_encoding) `

    getHistoryTreeQry = `SELECT branch_id, data, data_encoding FROM history_tree WHERE shard_id = ? AND tree_id = ? `

    deleteHistoryTreeQry = `DELETE FROM history_tree WHERE shard_id = ? AND tree_id = ? AND branch_id = ? `
)

func (d *driver) AddHistoryNodesQry() string {
    return addHistoryNodesQry
}

func (d *driver) GetHistoryNodesQry() string {
    return getHistoryNodesQry
}

func (d *driver) DeleteHistoryNodesQry() string {
    return deleteHistoryNodesQry
}

func (d *driver) AddHistoryTreeQry() string {
    return addHistoryTreeQry
}

func (d *driver) GetHistoryTreeQry() string {
    return getHistoryTreeQry
}

func (d *driver) DeleteHistoryTreeQry() string {
    return deleteHistoryTreeQry
}

