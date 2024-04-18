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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsDuplicateRequestError(t *testing.T) {
	testCases := []struct {
		name        string
		err         error
		expectedErr *DuplicateRequestError
		ok          bool
	}{
		{
			name:        "unwrapped",
			err:         &DuplicateRequestError{RunID: "a"},
			expectedErr: &DuplicateRequestError{RunID: "a"},
			ok:          true,
		},
		{
			name:        "wrapped",
			err:         fmt.Errorf("%w", &DuplicateRequestError{RunID: "b"}),
			expectedErr: &DuplicateRequestError{RunID: "b"},
			ok:          true,
		},
		{
			name: "not same type",
			err:  fmt.Errorf("adasdf"),
			ok:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e, ok := AsDuplicateRequestError(tc.err)
			assert.Equal(t, tc.ok, ok)
			if ok {
				assert.Equal(t, tc.expectedErr, e)
			}
		})
	}
}
