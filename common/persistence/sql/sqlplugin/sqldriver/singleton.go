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
		db   *sqlx.DB
		tx   *sqlx.Tx
		conn conn
	}
)

// NewSingletonSQLDriver returns a driver querying a single SQL database, which is the default driver
func NewSingletonSQLDriver(xdb *sqlx.DB, tx *sqlx.Tx) Driver {
	driver := &singleton{
		db:   xdb,
		tx:   tx,
		conn: xdb,
	}
	if tx != nil {
		// when tx is not nil, conn will be same as tx
		driver.conn = tx
	}
	return driver
}

func (s *singleton) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args)
}

func (s *singleton) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return s.db.NamedExecContext(ctx, query, arg)
}

func (s *singleton) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return s.db.GetContext(ctx, dest, query, args)
}

func (s *singleton) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return s.db.SelectContext(ctx, dest, query, args)
}

func (s *singleton) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.db.Exec(query, args)
}

func (s *singleton) Select(dest interface{}, query string, args ...interface{}) error {
	return s.db.Select(dest, query, args)
}

func (s *singleton) Get(dest interface{}, query string, args ...interface{}) error {
	return s.db.Get(dest, query, args)
}

func (s *singleton) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	return s.db.BeginTxx(ctx, opts)
}

func (s *singleton) Close() error {
	panic("implement me")
}

func (s *singleton) Commit() error {
	return s.tx.Commit()
}

func (s *singleton) Rollback() error {
	return s.tx.Rollback()
}
