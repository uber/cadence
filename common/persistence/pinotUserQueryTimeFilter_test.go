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

package persistence

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterQuery(t *testing.T) {
	tests := map[string]struct {
		query     string
		validated string
		err       string
	}{
		"Case1: empty query": {
			query:     "",
			validated: "",
		},
		"Case2: simple query": {
			query:     "WorkflowID = 'wid'",
			validated: "WorkflowID = 'wid'",
		},
		"Case3: query with custom field": {
			query:     "CustomStringField = 'custom'",
			validated: "CustomStringField = 'custom'",
		},
		"Case4: startTime query": {
			query:     "StartTime=123456789",
			validated: "",
		},
		"Case5-1: complex query I: with parenthesis": {
			query:     "(CustomStringField = 'custom and custom2 or custom3 order by') or StartTime between 1 and 10",
			validated: "CustomStringField = 'custom and custom2 or custom3 order by'",
		},
		"Case5-2: complex query II: with startTime": {
			query:     "DomainID = 'd-id' and (RunID = 'run-id' or StartTime < 123456778)",
			validated: "DomainID = 'd-id' and RunID = 'run-id'",
		},
		"Case5-3: complex query II: without startTime": {
			query:     "DomainID = 'd-id' or RunID = 'run-id' and WorkflowID = 'wid'",
			validated: "DomainID = 'd-id' or RunID = 'run-id' and WorkflowID = 'wid'",
		},
		"Case5-4: complex query IV": {
			query:     "WorkflowID = 'wid' and (CustomStringField = 'custom and custom2 or custom3 order by' or StartTime between 1 and 10)",
			validated: "WorkflowID = 'wid' and CustomStringField = 'custom and custom2 or custom3 order by'",
		},
		"Case6: invalid sql query": {
			query: "Invalid SQL",
			err:   "Invalid query.",
		},
		"Case7: query with missing val": {
			query:     "CloseTime = missing",
			validated: "CloseTime = missing",
		},
		"Case8: invalid where expression": {
			query: "InvalidWhereExpr",
			err:   "invalid where clause",
		},
		"Case9: or clause": {
			query:     "CustomIntField = 1 or CustomIntField = 2",
			validated: "CustomIntField = 1 or CustomIntField = 2",
		},
		"Case10: range query: custom filed": {
			query:     "CustomIntField BETWEEN 1 AND 2",
			validated: "CustomIntField between 1 and 2",
		},
		"Case11: system date attribute greater than or equal to": {
			query:     "StartTime >= 1697754674",
			validated: "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			qv := NewVisibilityQueryFilter()
			validated, err := qv.FilterQuery(test.query)
			if err != nil {
				assert.Equal(t, test.err, err.Error())
			} else {
				assert.Equal(t, test.validated, validated)
			}
		})
	}
}
