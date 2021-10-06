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

package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/VividCortex/mysqlerr"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/uber/cadence/common/persistence/sql/sqldriver"
	"github.com/uber/cadence/common/persistence/sql/sqlplugin"
)

type (
	db struct {
		converter   DataConverter
		driver      sqldriver.Driver
		originalDBs []*sqlx.DB
		numDBShards int
	}
)

func (mdb *db) GetTotalNumDBShards() int {
	return mdb.numDBShards
}

var _ sqlplugin.AdminDB = (*db)(nil)
var _ sqlplugin.DB = (*db)(nil)
var _ sqlplugin.Tx = (*db)(nil)

func (mdb *db) IsDupEntryError(err error) bool {
	sqlErr, ok := err.(*mysql.MySQLError)
	// ErrDupEntry MySQL Error 1062 indicates a duplicate primary key i.e. the row already exists,
	// so we don't do the insert and return a ConditionalUpdate error.
	return ok && sqlErr.Number == mysqlerr.ER_DUP_ENTRY
}

func (mdb *db) IsNotFoundError(err error) bool {
	if err == sql.ErrNoRows {
		return true
	}
	return false
}

func (mdb *db) IsTimeoutError(err error) bool {
	if err == context.DeadlineExceeded {
		return true
	}
	sqlErr, ok := err.(*mysql.MySQLError)
	if ok {
		if sqlErr.Number == mysqlerr.ER_NET_READ_INTERRUPTED ||
			sqlErr.Number == mysqlerr.ER_NET_WRITE_INTERRUPTED ||
			sqlErr.Number == mysqlerr.ER_LOCK_WAIT_TIMEOUT ||
			sqlErr.Number == mysqlerr.ER_XA_RBTIMEOUT ||
			sqlErr.Number == mysqlerr.ER_QUERY_TIMEOUT ||
			sqlErr.Number == mysqlerr.ER_LOCKING_SERVICE_TIMEOUT ||
			sqlErr.Number == mysqlerr.ER_REGEXP_TIME_OUT {
			return true
		}
	}
	return false
}

func (mdb *db) IsThrottlingError(err error) bool {
	sqlErr, ok := err.(*mysql.MySQLError)
	if ok {
		if sqlErr.Number == mysqlerr.ER_CON_COUNT_ERROR ||
			sqlErr.Number == mysqlerr.ER_TOO_MANY_USER_CONNECTIONS ||
			sqlErr.Number == mysqlerr.ER_TOO_MANY_CONCURRENT_TRXS ||
			sqlErr.Number == mysqlerr.ER_CLONE_TOO_MANY_CONCURRENT_CLONES {
			return true
		}
	}
	return false
}

// newDB returns an instance of DB, which is a logical
// connection to the underlying mysql database
// dbShardID is needed when tx is not nil
func newDB(xdbs []*sqlx.DB, tx *sqlx.Tx, dbShardID int, numDBShards int) (*db, error) {
	driver, err := sqldriver.NewDriver(xdbs, tx, dbShardID)
	if err != nil {
		return nil, err
	}

	db := &db{
		converter:   &converter{},
		originalDBs: xdbs, // this is kept because newDB will be called again when starting a transaction
		driver:      driver,
		numDBShards: numDBShards,
	}

	return db, nil
}

// BeginTx starts a new transaction and returns a reference to the Tx object
func (mdb *db) BeginTx(dbShardID int, ctx context.Context) (sqlplugin.Tx, error) {
	xtx, err := mdb.driver.BeginTxx(ctx, dbShardID, nil)
	if err != nil {
		return nil, err
	}
	return newDB(mdb.originalDBs, xtx, dbShardID, mdb.numDBShards)
}

// Commit commits a previously started transaction
func (mdb *db) Commit() error {
	return mdb.driver.Commit()
}

// Rollback triggers rollback of a previously started transaction
func (mdb *db) Rollback() error {
	return mdb.driver.Rollback()
}

// Close closes the connection to the mysql db
func (mdb *db) Close() error {
	return mdb.driver.Close()
}

// PluginName returns the name of the mysql plugin
func (mdb *db) PluginName() string {
	return PluginName
}

// SupportsTTL returns weather MySQL supports TTL
func (mdb *db) SupportsTTL() bool {
	return false
}

// MaxAllowedTTL returns the max allowed ttl MySQL supports
func (mdb *db) MaxAllowedTTL() (*time.Duration, error) {
	return nil, sqlplugin.ErrTTLNotSupported
}
