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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_ParseIntMultiRange(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectOutput []int
		expectError  string
	}{
		{
			name:         "empty",
			input:        "",
			expectOutput: []int{},
		},
		{
			name:         "single number",
			input:        " 1 ",
			expectOutput: []int{1},
		},
		{
			name:         "single range",
			input:        "1 - 3 ",
			expectOutput: []int{1, 2, 3},
		},
		{
			name:         "multi range",
			input:        "1 - 3 ,,  6",
			expectOutput: []int{1, 2, 3, 6},
		},
		{
			name:         "overlapping ranges",
			input:        "1-3,2-4",
			expectOutput: []int{1, 2, 3, 4},
		},
		{
			name:        "invalid single number",
			input:       "1a",
			expectError: "single number \"1a\": strconv.Atoi: parsing \"1a\": invalid syntax",
		},
		{
			name:        "invalid lower bound",
			input:       "1a-2",
			expectError: "lower range of \"1a-2\": strconv.Atoi: parsing \"1a\": invalid syntax",
		},
		{
			name:        "invalid upper bound",
			input:       "1-2a",
			expectError: "upper range of \"1-2a\": strconv.Atoi: parsing \"2a\": invalid syntax",
		},
		{
			name:        "invalid range",
			input:       "1-2-3",
			expectError: "invalid range \"1-2-3\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := parseIntMultiRange(tt.input)
			if tt.expectError != "" {
				assert.EqualError(t, err, tt.expectError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectOutput, output)
			}
		})
	}
}

func Test_anyToStringWorksWithTime(t *testing.T) {
	tm := time.Date(2020, 1, 15, 14, 30, 45, 0, time.UTC)

	assert.Equal(
		t,
		"2020-01-15 14:30:45 +0000 UTC",
		anyToString(tm, false, 100),
	)
	assert.Equal(
		t,
		"2020-01-15 ...  +0000 UTC",
		anyToString(tm, false, 20),
		"trimming should work for time.Time as well",
	)
}

func Test_anyToString(t *testing.T) {
	info := struct {
		Name   string
		Number int
		Time   time.Time
	}{
		"Joel",
		1234,
		time.Date(2019, 1, 15, 14, 30, 45, 0, time.UTC),
	}

	res := anyToString(info, false, 100)
	assert.Equal(t, "{Name:Joel, Number:1234, Time:2019-01-15 14:30:45 +0000 UTC}", res)
}
