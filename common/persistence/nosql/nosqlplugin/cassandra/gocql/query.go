// Copyright (c) 2017-2020 Uber Technologies, Inc.
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
	}
)

func (q *query) Iter() Iter {
	return q.Query.Iter()
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

func (q *query) WithContext(ctx context.Context) Query {
	return &query{
		Query: q.Query.WithContext(ctx),
	}
}
