// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to qvom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, qvETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package validator

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
			validated: "`Attr.CustomStringField` = 'custom'",
		},
		{
			msg:       "complex query",
			query:     "WorkflowID = 'wid' and ((CustomStringField = 'custom') or CustomIntField between 1 and 10)",
			validated: "WorkflowID = 'wid' and ((`Attr.CustomStringField` = 'custom') or `Attr.CustomIntField` between 1 and 10)",
		},
		{
			msg:   "invalid SQL",
			query: "Invalid SQL",
			err:   "BadRequestError{Message: Invalid query.}",
		},
		{
			msg:   "invalid where expression",
			query: "InvalidWhereExpr",
			err:   "BadRequestError{Message: invalid where clause}",
		},
		{
			msg:   "invalid comparison",
			query: "WorkflowID = 'wid' and 1 < 2",
			err:   "BadRequestError{Message: invalid comparison expression}",
		},
		{
			msg:   "invalid range",
			query: "1 between 1 and 2 or WorkflowID = 'wid'",
			err:   "BadRequestError{Message: invalid range expression}",
		},
		{
			msg:   "invalid search attribute in comparison",
			query: "Invalid = 'a' and 1 < 2",
			err:   "BadRequestError{Message: invalid search attribute}",
		},
		{
			msg:   "invalid search attribute in range",
			query: "Invalid between 1 and 2 or WorkflowID = 'wid'",
			err:   "BadRequestError{Message: invalid search attribute}",
		},
		{
			msg:       "only order by",
			query:     "order by CloseTime desc",
			validated: " order by CloseTime desc",
		},
		{
			msg:       "only order by search attribute",
			query:     "order by CustomIntField desc",
			validated: " order by `Attr.CustomIntField` desc",
		},
		{
			msg:       "condition + order by",
			query:     "WorkflowID = 'wid' order by CloseTime desc",
			validated: "WorkflowID = 'wid' order by CloseTime desc",
		},
		{
			msg:   "invalid order by attribute",
			query: "order by InvalidField desc",
			err:   "BadRequestError{Message: invalid order by attribute}",
		},
		{
			msg:   "invalid order by attribute expr",
			query: "order by 123",
			err:   "BadRequestError{Message: invalid order by expression}",
		},
		{
			msg:   "security SQL injection - with another statement",
			query: "WorkflowID = 'wid'; SELECT * FROM important_table;",
			err:   "BadRequestError{Message: Invalid query.}",
		},
		{
			msg:   "security SQL injection - with always true expression",
			query: "WorkflowID = 'wid' and (RunID = 'rid' or 1 = 1)",
			err:   "BadRequestError{Message: invalid comparison expression}",
		},
		{
			msg:   "security SQL injection - with union",
			query: "WorkflowID = 'wid' union select * from dummy",
			err:   "BadRequestError{Message: Invalid select query.}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			validSearchAttr := dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys())
			qv := NewQueryValidator(validSearchAttr)
			validated, err := qv.ValidateQuery(tt.query)
			if err != nil {
				assert.Equal(t, tt.err, err.Error())
			} else {
				assert.Equal(t, tt.validated, validated)
			}
		})
	}
}
