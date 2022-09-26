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

package nosql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNextID(t *testing.T) {
	tests := map[string]struct {
		acks     map[string]int64
		lastID   int64
		expected int64
	}{
		"expected case - last ID is equal to ack-levels": {
			acks:     map[string]int64{"a": 3},
			lastID:   3,
			expected: 4,
		},
		"expected case - last ID is equal to ack-levels haven't caught up": {
			acks:     map[string]int64{"a": 2},
			lastID:   3,
			expected: 4,
		},
		"error case - ack-levels are ahead for some reason": {
			acks:     map[string]int64{"a": 3},
			lastID:   2,
			expected: 4,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expected, getNextID(td.acks, td.lastID))
		})
	}
}
