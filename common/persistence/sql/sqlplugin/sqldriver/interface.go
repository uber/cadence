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
	// Driver interface is an abstraction to query SQL.
	//The layer is added so that we can have a adapter to support multiple SQL databases behind a single Cadence cluster
	Driver interface {

		// refactored from conn(xdb)
		conn

		// refactored from db(xdb)
		Exec(query string, args ...interface{}) (sql.Result, error)
		Select(dest interface{}, query string, args ...interface{}) error
		Get(dest interface{}, query string, args ...interface{}) error
		BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
		Close() error

		// refactored from tx(tx)
		Commit() error
		Rollback() error
	}

	conn interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
		NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
		GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
		SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	}
)