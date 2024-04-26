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

func TestConsistency_MarshalText(t *testing.T) {
	tests := []struct {
		c        Consistency
		wantText []byte
		testFn   func(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool
	}{
		{c: Any, wantText: []byte("ANY"), testFn: assert.Equal},
		{c: One, wantText: []byte("ONE"), testFn: assert.Equal},
		{c: Two, wantText: []byte("TWO"), testFn: assert.Equal},
		{c: Three, wantText: []byte("THREE"), testFn: assert.Equal},
		{c: Quorum, wantText: []byte("QUORUM"), testFn: assert.Equal},
		{c: All, wantText: []byte("ALL"), testFn: assert.Equal},
		{c: LocalQuorum, wantText: []byte("LOCAL_QUORUM"), testFn: assert.Equal},
		{c: EachQuorum, wantText: []byte("EACH_QUORUM"), testFn: assert.Equal},
		{c: LocalOne, wantText: []byte("LOCAL_ONE"), testFn: assert.Equal},
		{c: LocalOne, wantText: []byte("WRONG_VALUE"), testFn: assert.NotEqualValues},
	}
	for _, tt := range tests {
		t.Run(tt.c.String(), func(t *testing.T) {
			gotText, err := tt.c.MarshalText()
			assert.NoError(t, err)
			tt.testFn(t, tt.wantText, gotText)
		})
	}
}

func TestConsistency_String(t *testing.T) {
	c := Consistency(9)
	assert.Equal(t, c.String(), "invalid consistency: 9")
}

func TestConsistency_UnmarshalText(t *testing.T) {
	tests := []struct {
		destConsistency Consistency
		inputText       []byte
		testFn          func(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool
		wantErr         bool
	}{
		{destConsistency: Any, inputText: []byte("ANY"), testFn: assert.Equal},
		{destConsistency: One, inputText: []byte("ONE"), testFn: assert.Equal},
		{destConsistency: Two, inputText: []byte("TWO"), testFn: assert.Equal},
		{destConsistency: Three, inputText: []byte("THREE"), testFn: assert.Equal},
		{destConsistency: Quorum, inputText: []byte("QUORUM"), testFn: assert.Equal},
		{destConsistency: All, inputText: []byte("ALL"), testFn: assert.Equal},
		{destConsistency: LocalQuorum, inputText: []byte("LOCAL_QUORUM"), testFn: assert.Equal},
		{destConsistency: EachQuorum, inputText: []byte("EACH_QUORUM"), testFn: assert.Equal},
		{destConsistency: LocalOne, inputText: []byte("LOCAL_ONE"), testFn: assert.Equal},
		{destConsistency: LocalOne, inputText: []byte("WRONG_VALUE"), testFn: assert.NotEqualValues, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.destConsistency.String(), func(t *testing.T) {
			var c Consistency
			err := c.UnmarshalText(tt.inputText)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			tt.testFn(t, tt.destConsistency, c)
		})
	}
}

func TestParseConsistency(t *testing.T) {
	tests := []struct {
		destConsistency Consistency
		inputText       string
		testFn          func(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool
		wantErr         bool
	}{
		{destConsistency: Any, inputText: "any", testFn: assert.Equal},
		{destConsistency: One, inputText: "ONE", testFn: assert.Equal},
		{destConsistency: Two, inputText: "TWO", testFn: assert.Equal},
		{destConsistency: Three, inputText: "THREE", testFn: assert.Equal},
		{destConsistency: Quorum, inputText: "QUORUM", testFn: assert.Equal},
		{destConsistency: All, inputText: "all", testFn: assert.Equal},
		{destConsistency: LocalQuorum, inputText: "LOCAL_QUORUM", testFn: assert.Equal},
		{destConsistency: EachQuorum, inputText: "EACH_QUORUM", testFn: assert.Equal},
		{destConsistency: LocalOne, inputText: "LOCAL_ONE", testFn: assert.Equal},
		{destConsistency: Any, inputText: "WRONG_VALUE_FAILBACK_TO_ANY", testFn: assert.Equal, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(string(tt.inputText), func(t *testing.T) {
			got, err := ParseConsistency(tt.inputText)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			tt.testFn(t, tt.destConsistency, got)
		})
	}
}

func TestParseSerialConsistency(t *testing.T) {
	tests := []struct {
		destConsistency SerialConsistency
		inputText       string
		testFn          func(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool
		wantErr         bool
	}{
		{destConsistency: Serial, inputText: "serial", testFn: assert.Equal},
		{destConsistency: LocalSerial, inputText: "LOCAL_SERIAL", testFn: assert.Equal},
		{destConsistency: Serial, inputText: "WRONG_VALUE_FAILBACK_TO_ANY", testFn: assert.Equal, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(string(tt.inputText), func(t *testing.T) {
			got, err := ParseSerialConsistency(tt.inputText)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			tt.testFn(t, tt.destConsistency, got)
		})
	}
}

func TestSerialConsistency_MarshalText(t *testing.T) {
	tests := []struct {
		c        SerialConsistency
		wantText []byte
		testFn   func(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool
	}{
		{c: Serial, wantText: []byte("SERIAL"), testFn: assert.Equal},
		{c: LocalSerial, wantText: []byte("LOCAL_SERIAL"), testFn: assert.Equal},
		{c: LocalSerial, wantText: []byte("WRONG_VALUE"), testFn: assert.NotEqualValues},
	}
	for _, tt := range tests {
		t.Run(tt.c.String(), func(t *testing.T) {
			gotText, err := tt.c.MarshalText()
			assert.NoError(t, err)
			tt.testFn(t, tt.wantText, gotText)
		})
	}
}

func TestSerialConsistency_String(t *testing.T) {
	c := SerialConsistency(2)
	assert.Equal(t, c.String(), "invalid serial consistency 2")
}

func TestSerialConsistency_UnmarshalText(t *testing.T) {
	tests := []struct {
		destSerialConsistency SerialConsistency
		inputText             []byte
		wantErr               bool
	}{
		{destSerialConsistency: Serial, inputText: []byte("SERIAL")},
		{destSerialConsistency: LocalSerial, inputText: []byte("LOCAL_SERIAL")},
		{destSerialConsistency: Serial, inputText: []byte("WRONG_VALUE_DEFAULTS_TO_SERIAL"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.destSerialConsistency.String(), func(t *testing.T) {
			var c SerialConsistency
			err := c.UnmarshalText(tt.inputText)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.destSerialConsistency, c)
		})
	}
}

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
