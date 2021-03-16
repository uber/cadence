// Copyright (c) 2017-2020 Uber Technologies, Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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

package gocql

import (
	"context"

	"github.com/gocql/gocql"
)

var _ Query = (*query)(nil)

type (
	query struct {
		*gocql.Query

		session *session
	}
)

func newQuery(
	session *session,
	gocqlQuery *gocql.Query,
) *query {
	return &query{
		Query:   gocqlQuery,
		session: session,
	}
}

func (q *query) Exec() error {
	err := q.Query.Exec()
	return q.handleError(err)
}

func (q *query) Scan(
	dest ...interface{},
) error {
	err := q.Query.Scan(dest...)
	return q.handleError(err)
}

func (q *query) ScanCAS(
	dest ...interface{},
) (bool, error) {
	applied, err := q.Query.ScanCAS(dest...)
	return applied, q.handleError(err)
}

func (q *query) MapScan(
	m map[string]interface{},
) error {
	err := q.Query.MapScan(m)
	return q.handleError(err)
}

func (q *query) MapScanCAS(
	dest map[string]interface{},
) (bool, error) {
	applied, err := q.Query.MapScanCAS(dest)
	return applied, q.handleError(err)
}

func (q *query) Iter() Iter {
	iter := q.Query.Iter()
	if iter == nil {
		return nil
	}
	return iter
}

func (q *query) PageSize(n int) Query {
	q.Query.PageSize(n)
	return q
}

func (q *query) PageState(state []byte) Query {
	q.Query.PageState(state)
	return q
}

func (q *query) Consistency(c Consistency) Query {
	q.Query.Consistency(mustConvertConsistency(c))
	return q
}

func (q *query) WithTimestamp(timestamp int64) Query {
	q.Query.WithTimestamp(timestamp)
	return q
}

func (q *query) WithContext(ctx context.Context) Query {
	q2 := q.Query.WithContext(ctx)
	if q2 == nil {
		return nil
	}
	return newQuery(q.session, q2)
}

func (q *query) Bind(v ...interface{}) Query {
	q.Query.Bind(v...)
	return q
}

func (q *query) handleError(err error) error {
	return q.session.handleError(err)
}
