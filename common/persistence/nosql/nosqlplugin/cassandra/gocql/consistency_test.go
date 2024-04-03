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

package gocql

import (
	"testing"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
)

func Test_mustConvertConsistency(t *testing.T) {
	tests := []struct {
		input  Consistency
		output gocql.Consistency
	}{
		{Any, gocql.Any},
		{One, gocql.One},
		{Two, gocql.Two},
		{Three, gocql.Three},
		{Quorum, gocql.Quorum},
		{All, gocql.All},
		{LocalQuorum, gocql.LocalQuorum},
		{EachQuorum, gocql.EachQuorum},
		{LocalOne, gocql.LocalOne},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.output, mustConvertConsistency(tt.input))
	}
	assert.Panics(t, func() { mustConvertConsistency(Consistency(9999)) })
}

func Test_mustConvertSerialConsistency(t *testing.T) {
	tests := []struct {
		input  SerialConsistency
		output gocql.SerialConsistency
	}{
		{Serial, gocql.Serial},
		{LocalSerial, gocql.LocalSerial},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.output, mustConvertSerialConsistency(tt.input))
	}
	assert.Panics(t, func() { mustConvertSerialConsistency(SerialConsistency(9999)) })
}
