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

package mongodb

import (
	"context"

	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
)

// InsertIntoHistoryTreeAndNode inserts one or two rows: tree row and node row(at least one of them)
func (db *mdb) InsertIntoHistoryTreeAndNode(ctx context.Context, treeRow *nosqlplugin.HistoryTreeRow, nodeRow *nosqlplugin.HistoryNodeRow) error {
	panic("TODO")
}

// SelectFromHistoryNode read nodes based on a filter
func (db *mdb) SelectFromHistoryNode(ctx context.Context, filter *nosqlplugin.HistoryNodeFilter) ([]*nosqlplugin.HistoryNodeRow, []byte, error) {
	panic("TODO")
}

// DeleteFromHistoryTreeAndNode delete a branch record, and a list of ranges of nodes.
func (db *mdb) DeleteFromHistoryTreeAndNode(ctx context.Context, treeFilter *nosqlplugin.HistoryTreeFilter, nodeFilters []*nosqlplugin.HistoryNodeFilter) error {
	panic("TODO")
}

// SelectAllHistoryTrees will return all tree branches with pagination
func (db *mdb) SelectAllHistoryTrees(ctx context.Context, nextPageToken []byte, pageSize int) ([]*nosqlplugin.HistoryTreeRow, []byte, error) {
	panic("TODO")
}

// SelectFromHistoryTree read branch records for a tree
func (db *mdb) SelectFromHistoryTree(ctx context.Context, filter *nosqlplugin.HistoryTreeFilter) ([]*nosqlplugin.HistoryTreeRow, error) {
	panic("TODO")
}
