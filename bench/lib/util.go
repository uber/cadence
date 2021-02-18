// Copyright (c) 2017-2021 Uber Technologies Inc.

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

package lib

import (
	crypto "crypto/rand"
	"encoding/binary"
	"time"
)

// TODO: remove duplicated util functions once all bench tests are open sourced.

// MaxInt64 returns the max of given two int64's
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// MinInt64 returns the min of given two int64's
func MinInt64(a, b int64) int64 {
	if a > b {
		return b
	}
	return a
}

// RandInt64 returns a random int64
func RandInt64() int64 {
	bytes := make([]byte, 8)
	_, err := crypto.Read(bytes)
	if err != nil {
		return time.Now().UnixNano()
	}
	val := binary.BigEndian.Uint64(bytes)
	return int64(val & 0x3FFFFFFFFFFFFFFF)
}

// StringPtr returns string pointer
func StringPtr(str string) *string {
	return &str
}

// Int64Ptr returns int64 pointer
func Int64Ptr(v int64) *int64 {
	return &v
}
