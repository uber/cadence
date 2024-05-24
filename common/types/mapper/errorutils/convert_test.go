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

package errorutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertError(t *testing.T) {
	t.Run("sample error", func(t *testing.T) {
		err := &sampleError{
			message: "test",
		}
		isError, converted := ConvertError(err, sampleErrorConvertor)
		assert.True(t, isError, "is error")
		assert.Error(t, converted, "converted error")
		assert.Equal(t, err.message, converted.message)
	})
	t.Run("nil error is propagated as nil", func(t *testing.T) {
		isError, converted := ConvertError(nil, sampleErrorConvertor)
		assert.False(t, isError, "is error")
		assert.Nil(t, converted, "converted error")
	})
}

type sampleError struct {
	message string
}

func (s *sampleError) Error() string {
	return "sample error"
}

type convertedError struct {
	message string
}

func (c *convertedError) Error() string {
	return "converted error"
}

func sampleErrorConvertor(e *sampleError) *convertedError {
	return &convertedError{
		message: e.message,
	}
}
