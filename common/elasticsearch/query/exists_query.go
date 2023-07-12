// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package query

// ExistsQuery is a query that only matches on documents that the field
// has a value in them.
//
// For more details, see:
// https://www.elastic.co/guide/en/elasticsearch/reference/6.8/query-dsl-exists-query.html
type ExistsQuery struct {
	name      string
	queryName string
}

// NewExistsQuery creates and initializes a new exists query.
func NewExistsQuery(name string) *ExistsQuery {
	return &ExistsQuery{
		name: name,
	}
}

// QueryName sets the query name for the filter that can be used
// when searching for matched queries per hit.
func (q *ExistsQuery) QueryName(queryName string) *ExistsQuery {
	q.queryName = queryName
	return q
}

// Source returns the JSON serializable content for this query.
func (q *ExistsQuery) Source() (interface{}, error) {
	// {
	//   "exists" : {
	//     "field" : "user"
	//   }
	// }

	query := make(map[string]interface{})
	params := make(map[string]interface{})
	query["exists"] = params

	params["field"] = q.name
	if q.queryName != "" {
		params["_name"] = q.queryName
	}

	return query, nil
}
