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

type Query interface {
	// Source returns the JSON-serializable query request.
	Source() (interface{}, error)
}

type BoolQuery struct {
	mustClauses    []Query
	mustNotClauses []Query
	filterClauses  []Query
}

// Creates a new bool query.
func NewBoolQuery() *BoolQuery {
	return &BoolQuery{
		mustClauses:    make([]Query, 0),
		mustNotClauses: make([]Query, 0),
		filterClauses:  make([]Query, 0),
	}
}

func (q *BoolQuery) Must(queries ...Query) *BoolQuery {
	q.mustClauses = append(q.mustClauses, queries...)
	return q
}

func (q *BoolQuery) MustNot(queries ...Query) *BoolQuery {
	q.mustNotClauses = append(q.mustNotClauses, queries...)
	return q
}

func (q *BoolQuery) Filter(filters ...Query) *BoolQuery {
	q.filterClauses = append(q.filterClauses, filters...)
	return q
}

func (q *BoolQuery) Source() (interface{}, error) {
	query := make(map[string]interface{})

	boolClause := make(map[string]interface{})
	query["bool"] = boolClause

	// must
	if len(q.mustClauses) == 1 {
		src, err := q.mustClauses[0].Source()
		if err != nil {
			return nil, err
		}
		boolClause["must"] = src
	} else if len(q.mustClauses) > 1 {
		var clauses []interface{}
		for _, subQuery := range q.mustClauses {
			src, err := subQuery.Source()
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, src)
		}
		boolClause["must"] = clauses
	}

	// must_not
	if len(q.mustNotClauses) == 1 {
		src, err := q.mustNotClauses[0].Source()
		if err != nil {
			return nil, err
		}
		boolClause["must_not"] = src
	} else if len(q.mustNotClauses) > 1 {
		var clauses []interface{}
		for _, subQuery := range q.mustNotClauses {
			src, err := subQuery.Source()
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, src)
		}
		boolClause["must_not"] = clauses
	}

	// filter
	if len(q.filterClauses) == 1 {
		src, err := q.filterClauses[0].Source()
		if err != nil {
			return nil, err
		}
		boolClause["filter"] = src
	} else if len(q.filterClauses) > 1 {
		var clauses []interface{}
		for _, subQuery := range q.filterClauses {
			src, err := subQuery.Source()
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, src)
		}
		boolClause["filter"] = clauses
	}

	return query, nil
}
