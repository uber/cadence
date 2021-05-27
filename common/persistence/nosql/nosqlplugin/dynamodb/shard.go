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

package dynamodb

import (
	"context"
	"fmt"
	"strings"

	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
)

// InsertShard creates a new shard, return error is there is any.
// When error is nil, return applied=true if there is a conflict, and return the conflicted row as previous
func (db *ddb) InsertShard(ctx context.Context, row *nosqlplugin.ShardRow) (*nosqlplugin.ConflictedShardRow, error) {
	panic("TODO")
}

func convertToConflictedShardRow(shardID int, previousRangeID int64, previous map[string]interface{}) *nosqlplugin.ConflictedShardRow {
	var columns []string
	for k, v := range previous {
		columns = append(columns, fmt.Sprintf("%s=%v", k, v))
	}
	return &nosqlplugin.ConflictedShardRow{
		ShardID:         shardID,
		PreviousRangeID: previousRangeID,
		Details:         strings.Join(columns, ","),
	}
}

// SelectShard gets a shard
func (db *ddb) SelectShard(ctx context.Context, shardID int, currentClusterName string) (int64, *nosqlplugin.ShardRow, error) {
	panic("TODO")
}

// UpdateRangeID updates the rangeID, return error is there is any
// When error is nil, return applied=true if there is a conflict, and return the conflicted row as previous
func (db *ddb) UpdateRangeID(ctx context.Context, shardID int, rangeID int64, previousRangeID int64) (*nosqlplugin.ConflictedShardRow, error) {
	panic("TODO")
}

// UpdateShard updates a shard, return error is there is any.
// When error is nil, return applied=true if there is a conflict, and return the conflicted row as previous
func (db *ddb) UpdateShard(ctx context.Context, row *nosqlplugin.ShardRow, previousRangeID int64) (*nosqlplugin.ConflictedShardRow, error) {
	panic("TODO")
}
