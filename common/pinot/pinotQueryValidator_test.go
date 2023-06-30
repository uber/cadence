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
	tests := []struct {
		msg       string
		query     string
		validated string
		err       string
	}{
		{
			msg:       "empty query",
			query:     "",
			validated: "",
		},
		{
			msg:       "simple query",
			query:     "WorkflowID = 'wid'",
			validated: "WorkflowID = 'wid'",
		},
		{
			msg:       "custom field",
			query:     "CustomStringField = 'custom'",
			validated: "CustomStringField like '%custom%'",
		},
		{
			msg:       "custom field with Or",
			query:     "CustomStringField='Or'",
			validated: "CustomStringField like '%Or%'",
		},
		{
			msg:       "custom keyword field",
			query:     "CustomKeywordField = 'custom'",
			validated: "CustomKeywordField = 'custom'",
		},
		{
			msg:       "complex query",
			query:     "WorkflowID = 'wid' and ((CustomStringField = 'custom and custom2 or custom3 order by') or CustomIntField between 1 and 10)",
			validated: "WorkflowID = 'wid' and ((CustomStringField like '%custom and custom2 or custom3 order by%') or CustomIntField between 1 and 10)",
		},
		{
			msg:   "invalid SQL",
			query: "Invalid SQL",
			err:   "Invalid query.",
		},
		{
			msg:       "SQL with missing val",
			query:     "CloseTime = missing",
			validated: "CloseTime = `-1`",
		},
		{
			msg:   "invalid where expression",
			query: "InvalidWhereExpr",
			err:   "invalid where clause",
		},
		{
			msg:   "invalid search attribute in comparison",
			query: "Invalid = 'a' and 1 < 2",
			err:   "invalid search attribute \"Invalid\"",
		},
		{
			msg:       "only order by",
			query:     "order by CloseTime desc",
			validated: " order by CloseTime desc",
		},
		{
			msg:       "only order by search attribute",
			query:     "order by CustomIntField desc",
			validated: " order by CustomIntField desc",
		},
		{
			msg:       "condition + order by",
			query:     "WorkflowID = 'wid' order by CloseTime desc",
			validated: "WorkflowID = 'wid' order by CloseTime desc",
		},
		{
			msg:   "security SQL injection - with another statement",
			query: "WorkflowID = 'wid'; SELECT * FROM important_table;",
			err:   "Invalid query.",
		},
		{
			msg:   "security SQL injection - with union",
			query: "WorkflowID = 'wid' union select * from dummy",
			err:   "Invalid select query.",
		},
		{
			msg:       "or clause",
			query:     "CustomIntField = 1 or CustomIntField = 2",
			validated: "CustomIntField = 1 or CustomIntField = 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			validSearchAttr := dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys())
			qv := NewPinotQueryValidator(validSearchAttr())
			validated, err := qv.ValidateQuery(tt.query)
			if err != nil {
				assert.Equal(t, tt.err, err.Error())
			} else {
				assert.Equal(t, tt.validated, validated)
			}
		})
	}
}
