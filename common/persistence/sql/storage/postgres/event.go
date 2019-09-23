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

package postgres

import (
	"database/sql"
	"log"

	"github.com/uber/cadence/common/persistence/sql/storage/sqldb"
)

// InsertIntoEvents inserts a row into events table
func (mdb *DB) InsertIntoEvents(row *sqldb.EventsRow) (sql.Result, error) {
	log.Panic("do not call events v1 api")
	return nil, nil
}

// UpdateEvents updates a row in events table
func (mdb *DB) UpdateEvents(row *sqldb.EventsRow) (sql.Result, error) {
	log.Panic("do not call events v1 api")
	return nil, nil
}

// SelectFromEvents reads one or more rows from events table
func (mdb *DB) SelectFromEvents(filter *sqldb.EventsFilter) ([]sqldb.EventsRow, error) {
	log.Panic("do not call events v1 api")
	return nil, nil
}

// DeleteFromEvents deletes one or more rows from events table
func (mdb *DB) DeleteFromEvents(filter *sqldb.EventsFilter) (sql.Result, error) {
	log.Panic("do not call events v1 api")
	return nil, nil
}

// LockEvents acquires a write lock on a single row in events table
func (mdb *DB) LockEvents(filter *sqldb.EventsFilter) (*sqldb.EventsRow, error) {
	log.Panic("do not call events v1 api")
	return nil, nil
}
