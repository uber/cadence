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

package pinot

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/definition"
	"github.com/uber/cadence/common/dynamicconfig"
)

func TestValidateQuery(t *testing.T) {
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
			validated: "(JSON_MATCH(Attr, '\"$.CustomStringField\" is not null') AND REGEXP_LIKE(JSON_EXTRACT_SCALAR(Attr, '$.CustomStringField', 'string'), 'custom*'))",
		},
		"Case4: custom field query with or in string": {
			query:     "CustomStringField='Or'",
			validated: "(JSON_MATCH(Attr, '\"$.CustomStringField\" is not null') AND REGEXP_LIKE(JSON_EXTRACT_SCALAR(Attr, '$.CustomStringField', 'string'), 'Or*'))",
		},
		"Case5: custom keyword field query": {
			query:     "CustomKeywordField = 'custom'",
			validated: "(JSON_MATCH(Attr, '\"$.CustomKeywordField\"=''custom''') or JSON_MATCH(Attr, '\"$.CustomKeywordField[*]\"=''custom'''))",
		},
		"Case6-1: complex query I: with parenthesis": {
			query:     "(CustomStringField = 'custom and custom2 or custom3 order by') or CustomIntField between 1 and 10",
			validated: "((JSON_MATCH(Attr, '\"$.CustomStringField\" is not null') AND REGEXP_LIKE(JSON_EXTRACT_SCALAR(Attr, '$.CustomStringField', 'string'), 'custom and custom2 or custom3 order by*')) or (JSON_MATCH(Attr, '\"$.CustomIntField\" is not null') AND CAST(JSON_EXTRACT_SCALAR(Attr, '$.CustomIntField') AS INT) >= 1 AND CAST(JSON_EXTRACT_SCALAR(Attr, '$.CustomIntField') AS INT) <= 10))",
		},
		"Case6-2: complex query II: with only system keys": {
			query:     "DomainID = 'd-id' and (RunID = 'run-id' or WorkflowID = 'wid')",
			validated: "DomainID = 'd-id' and (RunID = 'run-id' or WorkflowID = 'wid')",
		},
		"Case6-3: complex query III: operation priorities": {
			query:     "DomainID = 'd-id' or RunID = 'run-id' and WorkflowID = 'wid'",
			validated: "(DomainID = 'd-id' or RunID = 'run-id' and WorkflowID = 'wid')",
		},
		"Case6-4: complex query IV": {
			query:     "WorkflowID = 'wid' and (CustomStringField = 'custom and custom2 or custom3 order by' or CustomIntField between 1 and 10)",
			validated: "WorkflowID = 'wid' and ((JSON_MATCH(Attr, '\"$.CustomStringField\" is not null') AND REGEXP_LIKE(JSON_EXTRACT_SCALAR(Attr, '$.CustomStringField', 'string'), 'custom and custom2 or custom3 order by*')) or (JSON_MATCH(Attr, '\"$.CustomIntField\" is not null') AND CAST(JSON_EXTRACT_SCALAR(Attr, '$.CustomIntField') AS INT) >= 1 AND CAST(JSON_EXTRACT_SCALAR(Attr, '$.CustomIntField') AS INT) <= 10))",
		},
		"Case7: invalid sql query": {
			query: "Invalid SQL",
			err:   "Invalid query.",
		},
		"Case8: query with missing val": {
			query:     "CloseTime = missing",
			validated: "CloseTime = `-1`",
		},
		"Case9: invalid where expression": {
			query: "InvalidWhereExpr",
			err:   "invalid where clause",
		},
		"Case10: invalid search attribute": {
			query: "Invalid = 'a' and 1 < 2",
			err:   "invalid search attribute \"Invalid\"",
		},
		"Case11-1: order by clause": {
			query:     "order by CloseTime desc",
			validated: " order by CloseTime desc",
		},
		"Case11-2: only order by clause with custom field": {
			query:     "order by CustomIntField desc",
			validated: " order by CustomIntField desc",
		},
		"Case11-3: order by clause with custom field": {
			query:     "WorkflowID = 'wid' order by CloseTime desc",
			validated: "WorkflowID = 'wid' order by CloseTime desc",
		},
		"Case12-1: security SQL injection - with another statement": {
			query: "WorkflowID = 'wid'; SELECT * FROM important_table;",
			err:   "Invalid query.",
		},
		"Case12-2: security SQL injection - with union": {
			query: "WorkflowID = 'wid' union select * from dummy",
			err:   "Invalid select query.",
		},
		"Case13: or clause": {
			query:     "CustomIntField = 1 or CustomIntField = 2",
			validated: "(JSON_MATCH(Attr, '\"$.CustomIntField\"=''1''') or JSON_MATCH(Attr, '\"$.CustomIntField\"=''2'''))",
		},
		"Case14-1: range query: custom filed": {
			query:     "CustomIntField BETWEEN 1 AND 2",
			validated: "(JSON_MATCH(Attr, '\"$.CustomIntField\" is not null') AND CAST(JSON_EXTRACT_SCALAR(Attr, '$.CustomIntField') AS INT) >= 1 AND CAST(JSON_EXTRACT_SCALAR(Attr, '$.CustomIntField') AS INT) <= 2)",
		},
		"Case14-2: range query: system filed": {
			query:     "NumClusters BETWEEN 1 AND 2",
			validated: "NumClusters between 1 and 2",
		},
		"Case15-1: custom date attribute less than": {
			query:     "CustomDatetimeField < 1697754674",
			validated: "(JSON_MATCH(Attr, '\"$.CustomDatetimeField\" is not null') AND CAST(JSON_EXTRACT_SCALAR(Attr, '$.CustomDatetimeField') AS BIGINT) < 1697754674)",
		},
		"Case15-2: custom date attribute greater than or equal to": {
			query:     "CustomDatetimeField >= 1697754674",
			validated: "(JSON_MATCH(Attr, '\"$.CustomDatetimeField\" is not null') AND CAST(JSON_EXTRACT_SCALAR(Attr, '$.CustomDatetimeField') AS BIGINT) >= 1697754674)",
		},
		"Case15-3: system date attribute greater than or equal to": {
			query:     "StartTime >= 1697754674",
			validated: "StartTime >= 1697754674",
		},
		"Case15-4: unix nano converts to milli seconds for equal statements": {
			query:     "StartTime = 1707319950934000128",
			validated: "StartTime = 1707319950934",
		},
		"Case15-5: unix nano converts to milli seconds for range query": {
			query:     "StartTime = 1707319950934000128",
			validated: "StartTime = 1707319950934",
		},
		"Case15-6: open workflows": {
			query:     "CloseTime = -1",
			validated: "CloseTime = -1",
		},
		"Case15-6: startTime for range query": {
			query:     "StartTime BETWEEN 1707319950934000128 AND 1707319950935000128",
			validated: "StartTime between 1707319950934 and 1707319950935",
		},
		"Case16-1: custom int attribute greater than or equal to": {
			query:     "CustomIntField >= 0",
			validated: "(JSON_MATCH(Attr, '\"$.CustomIntField\" is not null') AND CAST(JSON_EXTRACT_SCALAR(Attr, '$.CustomIntField') AS INT) >= 0)",
		},
		"Case16-2: custom double attribute greater than or equal to": {
			query:     "CustomDoubleField >= 0",
			validated: "(JSON_MATCH(Attr, '\"$.CustomDoubleField\" is not null') AND CAST(JSON_EXTRACT_SCALAR(Attr, '$.CustomDoubleField') AS DOUBLE) >= 0)",
		},
		"Case17: custom keyword attribute greater than or equal to. Will return error run time": {
			query:     "CustomKeywordField < 0",
			validated: "(JSON_MATCH(Attr, '\"$.CustomKeywordField\"<''0''') or JSON_MATCH(Attr, '\"$.CustomKeywordField[*]\"<''0'''))",
		},
		// TODO
		"Case18: custom int order by. Will have errors at run time. Doesn't support for now": {
			query:     "CustomIntField = 0 order by CustomIntField desc",
			validated: "JSON_MATCH(Attr, '\"$.CustomIntField\"=''0''') order by CustomIntField desc",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			validSearchAttr := dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys())
			qv := NewPinotQueryValidator(validSearchAttr())
			validated, err := qv.ValidateQuery(test.query)
			if err != nil {
				assert.Equal(t, test.err, err.Error())
			} else {
				assert.Equal(t, test.validated, validated)
			}
		})
	}
}
