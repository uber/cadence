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

// InsertIntoShards inserts one or more rows into shards table
func (mdb *DB) InsertIntoShards(row *sqldb.ShardsRow) (sql.Result, error) {
	return mdb.conn.Exec(mdb.driver.CreateShardQuery(), row.ShardID, row.RangeID, row.Data, row.DataEncoding)
}

// UpdateShards updates one or more rows into shards table
func (mdb *DB) UpdateShards(row *sqldb.ShardsRow) (sql.Result, error) {
	return mdb.conn.Exec(mdb.driver.UpdateShardQuery(), row.RangeID, row.Data, row.DataEncoding, row.ShardID)
}

// SelectFromShards reads one or more rows from shards table
func (mdb *DB) SelectFromShards(filter *sqldb.ShardsFilter) (*sqldb.ShardsRow, error) {
	var row sqldb.ShardsRow
	err := mdb.conn.Get(&row, mdb.driver.GetShardQuery(), filter.ShardID)
	if err != nil {
		return nil, err
	}
	return &row, err
}

// ReadLockShards acquires a read lock on a single row in shards table
func (mdb *DB) ReadLockShards(filter *sqldb.ShardsFilter) (int, error) {
	var rangeID int
	err := mdb.conn.Get(&rangeID, mdb.driver.ReadLockShardQuery(), filter.ShardID)
	return rangeID, err
}

// WriteLockShards acquires a write lock on a single row in shards table
func (mdb *DB) WriteLockShards(filter *sqldb.ShardsFilter) (int, error) {
	var rangeID int
	err := mdb.conn.Get(&rangeID, mdb.driver.LockShardQuery(), filter.ShardID)
	return rangeID, err
}
