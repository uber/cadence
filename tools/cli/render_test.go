// Copyright (c) 2022 Uber Technologies, Inc.
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

package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_RenderTable(t *testing.T) {
	tests := []struct {
		name         string
		data         interface{}
		opts         RenderOptions
		expectOutput string
		expectErr    string
	}{
		{
			name:         "nil slice",
			data:         ([]testRow)(nil),
			expectOutput: "",
		},
		{
			name:         "empty slice",
			data:         []testRow{},
			expectOutput: "",
		},
		{
			name:         "nil struct",
			data:         (*testRow)(nil),
			expectOutput: "",
		},
		{
			name: "single struct",
			data: testTable[0],
			expectOutput: "" +
				"  STRING | INTEGER | BOOL |   TIME   |    MAP     |  SLICE   \n" +
				"  text   |     123 | true | 03:04:05 | A:AA, B:BB | 1, 2, 3  \n",
		},
		{
			name: "pointer to single struct",
			data: &testTable[0],
			expectOutput: "" +
				"  STRING | INTEGER | BOOL |   TIME   |    MAP     |  SLICE   \n" +
				"  text   |     123 | true | 03:04:05 | A:AA, B:BB | 1, 2, 3  \n",
		},
		{
			name: "a slice",
			data: testTable,
			expectOutput: "" +
				"        STRING        | INTEGER | BOOL  |   TIME   |    MAP     |  SLICE   \n" +
				"  text                |     123 | true  | 03:04:05 | A:AA, B:BB | 1, 2, 3  \n" +
				"  .../ string this is |     456 | false | 13:14:15 |            |          \n",
		},
		{
			name: "a slice with rendering options",
			data: testTable,
			opts: RenderOptions{OptionalColumns: map[string]bool{"map": true, "slice": false}, PrintDateTime: true},
			expectOutput: "" +
				"        STRING        | INTEGER | BOOL  |         TIME         |    MAP      \n" +
				"  text                |     123 | true  | 2000-01-02T03:04:05Z | A:AA, B:BB  \n" +
				"  .../ string this is |     456 | false | 2000-11-12T13:14:15Z |             \n",
		},
		{
			name:      "non-struct element",
			data:      123,
			expectErr: "table slice element must be a struct, provided: int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &strings.Builder{}
			err := RenderTable(builder, tt.data, tt.opts)
			if tt.expectErr != "" {
				assert.EqualError(t, err, tt.expectErr)
			} else {
				assert.Equal(t, tt.expectOutput, builder.String())
			}
		})
	}
}

func Test_RenderTemplate(t *testing.T) {
	tests := []struct {
		name         string
		data         interface{}
		template     string
		opts         RenderOptions
		expectOutput string
		expectErr    string
	}{
		{
			name:         "a field from nil struct",
			data:         (*testRow)(nil),
			template:     "{{.StringField}}",
			expectOutput: "",
		},
		{
			name:         "a field from struct",
			data:         testTable[0],
			template:     "{{.StringField}}",
			expectOutput: "text",
		},
		{
			name:         "json function",
			data:         testTable[0],
			template:     "{{json .MapField}}",
			expectOutput: "{\n  \"A\": \"AA\",\n  \"B\": \"BB\"\n}",
		},
		{
			name:     "table function",
			data:     testTable,
			template: "{{table .}}",
			expectOutput: "" +
				"        STRING        | INTEGER | BOOL  |   TIME   |    MAP     |  SLICE   \n" +
				"  text                |     123 | true  | 03:04:05 | A:AA, B:BB | 1, 2, 3  \n" +
				"  .../ string this is |     456 | false | 13:14:15 |            |          \n",
		},
		{
			name:      "invalid template",
			data:      testTable,
			template:  "{{invalid}}",
			expectErr: "invalid template \"{{invalid}}\": template: :1: function \"invalid\" not defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &strings.Builder{}
			err := RenderTemplate(builder, tt.data, tt.template, tt.opts)
			if tt.expectErr != "" {
				assert.EqualError(t, err, tt.expectErr)
			} else {
				assert.Equal(t, tt.expectOutput, builder.String())
			}
		})
	}
}

type testRow struct {
	StringField  string            `header:"string" maxLength:"16"`
	IntField     int               `header:"integer"`
	BoolField    bool              `header:"bool"`
	TimeField    time.Time         `header:"time"`
	MapField     map[string]string `header:"map"`
	SliceField   []int             `header:"slice"`
	IgnoredField int
}

var testTable = []testRow{
	{
		StringField: "text",
		IntField:    123,
		BoolField:   true,
		TimeField:   time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC),
		MapField:    map[string]string{"A": "AA", "B": "BB"},
		SliceField:  []int{1, 2, 3},
	},
	{
		StringField: "very long/ string this is",
		IntField:    456,
		BoolField:   false,
		TimeField:   time.Date(2000, 11, 12, 13, 14, 15, 16, time.FixedZone("UTC+3", 0)),
		MapField:    nil,
		SliceField:  nil,
	},
}
