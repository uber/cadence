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
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/uber/cadence/common/persistence/sql/storage/sqldb"
)

// db represents a logical connection to mysql database
type db struct {
	db        *sqlx.DB
	tx        *sqlx.Tx
	conn      sqldb.Conn
	converter DataConverter
}

var _ sqldb.Tx = (*db)(nil)
var _ sqldb.DB = (*db)(nil)

func (mdb *db) GetDriverName() string {
	return DriverName
}

// ErrDupEntry indicates a duplicate primary key i.e. the row already exists,
// check http://www.postgresql.org/docs/9.3/static/errcodes-appendix.html
const ErrDupEntry = "23505"

func (mdb *db) IsDupEntryError(err error) bool {
	sqlErr, ok := err.(*pq.Error)
	return ok && sqlErr.Code == ErrDupEntry
}

// NewDB returns an instance of DB, which is a logical
// connection to the underlying mysql database
func NewDB(xdb *sqlx.DB, tx *sqlx.Tx) sqldb.DB {
	mdb := &db{db: xdb, tx: tx}
	mdb.conn = xdb
	if tx != nil {
		mdb.conn = tx
	}
	mdb.converter = &converter{}
	return mdb
}

// BeginTx starts a new transaction and returns a reference to the Tx object
func (mdb *db) BeginTx() (sqldb.Tx, error) {
	xtx, err := mdb.db.Beginx()
	if err != nil {
		return nil, err
	}
	return NewDB(mdb.db, xtx), nil
}

// Commit commits a previously started transaction
func (mdb *db) Commit() error {
	return mdb.tx.Commit()
}

// Rollback triggers rollback of a previously started transaction
func (mdb *db) Rollback() error {
	return mdb.tx.Rollback()
}

// Close closes the connection to the mysql db
func (mdb *db) Close() error {
	return mdb.db.Close()
}

// DriverName returns the name of the mysql driver
func (mdb *db) DriverName() string {
	return DriverName
}
