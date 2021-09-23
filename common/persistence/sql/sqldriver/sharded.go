// Copyright (c) 2021 Uber Technologies, Inc.
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

package sqldriver

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"go.uber.org/multierr"
)

type (
	// sharded is the driver querying a group of SQL databases as sharded solution
	sharded struct {
		dbs    []*sqlx.DB // this is for starting a transaction, or executing any non transaction query
		tx    *sqlx.Tx // this is a reference of a started transaction
		useTx bool     // if tx is not nil, the methods from commonOfDbAndTx should use tx
		currTxShardID int // which shard is current tx started from
	}
)

// NewShardedSQLDriver returns a driver querying a single SQL database, which is the default driver
// xdbs is the list of connections to the sql instances. The length of the list of the list is the totalNumShards
// dbShardID is needed when tx is not nil. It means a started transaction in the shard.
func NewShardedSQLDriver(xdbs []*sqlx.DB, xtx *sqlx.Tx, dbShardID int) Driver {
	driver := &sharded{
		dbs:   xdbs,
		tx:   xtx,
	}
	if xtx != nil {
		driver.useTx = true
		driver.currTxShardID = dbShardID
	}
	return driver
}

// below are shared by transactional and non-transactional, if s.tx is not nil then use s.tx, otherwise use s.db

func (s sharded) ExecContext(ctx context.Context, dbShardID int, query string, args ...interface{}) (sql.Result, error) {
	panic("implement me")
}

func (s sharded) NamedExecContext(ctx context.Context, dbShardID int, query string, arg interface{}) (sql.Result, error) {
	panic("implement me")
}

func (s sharded) GetContext(ctx context.Context, dbShardID int, dest interface{}, query string, args ...interface{}) error {
	panic("implement me")
}

func (s sharded) SelectContext(ctx context.Context, dbShardID int, dest interface{}, query string, args ...interface{}) error {
	panic("implement me")
}

// below are non-transactional methods only

func (s sharded) Exec(dbShardID int, query string, args ...interface{}) (sql.Result, error) {
	panic("implement me")
}

func (s sharded) Select(dbShardID int, dest interface{}, query string, args ...interface{}) error {
	panic("implement me")
}

func (s sharded) Get(dbShardID int, dest interface{}, query string, args ...interface{}) error {
	panic("implement me")
}

func (s sharded) BeginTxx(ctx context.Context, dbShardID int, opts *sql.TxOptions) (*sqlx.Tx, error) {
	panic("implement me")
}

func (s sharded) Close() error {
	var errs []error
	for _, db := range s.dbs{
		err := db.Close()
		if err != nil{
			errs = append(errs, err)
		}
	}
	if len(errs)>0{
		return multierr.Combine(errs...)
	}
	return nil
}

// below are transactional methods only

func (s sharded) Commit() error {
	return s.tx.Commit()
}

func (s sharded) Rollback() error {
	return s.tx.Rollback()
}

