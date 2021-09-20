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
)

type (
	// singleton is the driver querying a single SQL database, which is the default driver
	singleton struct {
		db   *sqlx.DB        // this is for starting a transaction, or executing any non transaction query
		tx   *sqlx.Tx        // this is a reference of a started transaction
		conn commonOfDbAndTx // this is a merged of the above two. In some case it' either db or tx depends on whether or not a transaction has started
	}

	// a wrapper to help xdb to fit commonOfDbAndTx interface
	xdbWrapper struct {
		xdb *sqlx.DB
	}

	// a wrapper to help xtx to fit commonOfDbAndTx interface
	xtxWrapper struct {
		xtx *sqlx.Tx
	}
)

// NewSingletonSQLDriver returns a driver querying a single SQL database, which is the default driver
// typically dbShardID is needed when tx is not nil, because it means a started transaction in a shard.
// But this singleton doesn't have sharding so omitting it.
func NewSingletonSQLDriver(xdb *sqlx.DB, xtx *sqlx.Tx, _ int) Driver {
	driver := &singleton{
		db:   xdb,
		tx:   xtx,
		conn: newXdbWrapper(xdb),
	}
	if xtx != nil {
		// when tx is not nil, commonOfDbAndTx will be same as tx
		driver.conn = newXtxWrapper(xtx)
	}
	return driver
}

// below are shared by transactional and non-transactional

func (s *singleton) ExecContext(ctx context.Context, dbShardID int, query string, args ...interface{}) (sql.Result, error) {
	return s.conn.ExecContext(ctx, dbShardID, query, args...)
}

func (s *singleton) NamedExecContext(ctx context.Context, dbShardID int, query string, arg interface{}) (sql.Result, error) {
	return s.conn.NamedExecContext(ctx, dbShardID, query, arg)
}

func (s *singleton) GetContext(ctx context.Context, dbShardID int, dest interface{}, query string, args ...interface{}) error {
	return s.conn.GetContext(ctx, dbShardID, dest, query, args...)
}

func (s *singleton) SelectContext(ctx context.Context, dbShardID int, dest interface{}, query string, args ...interface{}) error {
	return s.conn.SelectContext(ctx, dbShardID, dest, query, args...)
}

// below are non-transactional methods only

func (s *singleton) Exec(_ int, query string, args ...interface{}) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

func (s *singleton) Select(_ int, dest interface{}, query string, args ...interface{}) error {
	return s.db.Select(dest, query, args...)
}

func (s *singleton) Get(_ int, dest interface{}, query string, args ...interface{}) error {
	return s.db.Get(dest, query, args...)
}

func (s *singleton) BeginTxx(ctx context.Context, _ int, opts *sql.TxOptions) (*sqlx.Tx, error) {
	return s.db.BeginTxx(ctx, opts)
}

func (s *singleton) Close() error {
	return s.db.Close()
}

// below are transactional methods only

func (s *singleton) Commit() error {
	return s.tx.Commit()
}

func (s *singleton) Rollback() error {
	return s.tx.Rollback()
}

func newXdbWrapper(xdb *sqlx.DB, ) commonOfDbAndTx {
	return &xdbWrapper{
		xdb: xdb,
	}
}

func (x *xdbWrapper) ExecContext(ctx context.Context, _ int, query string, args ...interface{}) (sql.Result, error) {
	return x.xdb.ExecContext(ctx, query, args...)
}

func (x *xdbWrapper) NamedExecContext(ctx context.Context, _ int, query string, arg interface{}) (sql.Result, error) {
	return x.xdb.NamedExecContext(ctx, query, arg)
}

func (x *xdbWrapper) GetContext(ctx context.Context, _ int, dest interface{}, query string, args ...interface{}) error {
	return x.xdb.GetContext(ctx, dest, query, args...)
}

func (x *xdbWrapper) SelectContext(ctx context.Context, _ int, dest interface{}, query string, args ...interface{}) error {
	return x.xdb.SelectContext(ctx, dest, query, args...)
}

func newXtxWrapper(xtx *sqlx.Tx, ) commonOfDbAndTx {
	return &xtxWrapper{
		xtx: xtx,
	}
}

func (x *xtxWrapper) ExecContext(ctx context.Context, _ int, query string, args ...interface{}) (sql.Result, error) {
	return x.xtx.ExecContext(ctx, query, args...)
}

func (x xtxWrapper) NamedExecContext(ctx context.Context, _ int, query string, arg interface{}) (sql.Result, error) {
	return x.xtx.NamedExecContext(ctx, query, arg)
}

func (x xtxWrapper) GetContext(ctx context.Context, _ int, dest interface{}, query string, args ...interface{}) error {
	return x.xtx.GetContext(ctx, dest, query, args...)
}

func (x xtxWrapper) SelectContext(ctx context.Context, _ int, dest interface{}, query string, args ...interface{}) error {
	return x.xtx.SelectContext(ctx, dest, query, args...)
}