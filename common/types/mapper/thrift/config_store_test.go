// Copyright (c) 2021 Uber Technologies Inc.
//
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

package thrift

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/testdata"
)

func TestDynamicConfigBlob(t *testing.T) {
	testCases := []struct {
		desc  string
		input *types.DynamicConfigBlob
	}{
		{
			desc:  "non-nil input test",
			input: &testdata.DynamicConfigBlob,
		},
		{
			desc:  "empty input test",
			input: &types.DynamicConfigBlob{},
		},
		{
			desc:  "nil input test",
			input: nil,
		},
	}
	for _, tc := range testCases {
		thriftObj := FromDynamicConfigBlob(tc.input)
		roundTripObj := ToDynamicConfigBlob(thriftObj)
		assert.Equal(t, tc.input, roundTripObj)
	}
}

func TestDynamicConfigEntryArray(t *testing.T) {
	testCase := []struct {
		desc  string
		input []*types.DynamicConfigEntry
	}{
		{
			desc:  "non-nil input test",
			input: []*types.DynamicConfigEntry{&testdata.DynamicConfigEntry},
		},
		{
			desc:  "empty input test",
			input: []*types.DynamicConfigEntry{},
		},
		{
			desc:  "nil input test",
			input: nil,
		},
	}
	for _, tc := range testCase {
		thriftObj := FromDynamicConfigEntryArray(tc.input)
		roundTripObj := ToDynamicConfigEntryArray(thriftObj)
		assert.Equal(t, tc.input, roundTripObj)
	}
}

func TestDynamicConfigEntry(t *testing.T) {
	testCases := []struct {
		desc  string
		input *types.DynamicConfigEntry
	}{
		{
			desc:  "non-nil input test",
			input: &testdata.DynamicConfigEntry,
		},
		{
			desc:  "empty input test",
			input: &types.DynamicConfigEntry{},
		},
		{
			desc:  "nil input test",
			input: nil,
		},
	}
	for _, tc := range testCases {
		thriftObj := FromDynamicConfigEntry(tc.input)
		roundTripObj := ToDynamicConfigEntry(thriftObj)
		assert.Equal(t, tc.input, roundTripObj)
	}
}

func TestDynamicConfigValueArray(t *testing.T) {
	testCases := []struct {
		desc  string
		input []*types.DynamicConfigValue
	}{
		{
			desc:  "non-nil input test",
			input: []*types.DynamicConfigValue{&testdata.DynamicConfigValue},
		},
		{
			desc:  "empty input test",
			input: []*types.DynamicConfigValue{},
		},
		{
			desc:  "nil input test",
			input: nil,
		},
	}
	for _, tc := range testCases {
		thriftObj := FromDynamicConfigValueArray(tc.input)
		roundTripObj := ToDynamicConfigValueArray(thriftObj)
		assert.Equal(t, tc.input, roundTripObj)
	}
}

func TestDynamicConfigValue(t *testing.T) {
	testCases := []struct {
		desc  string
		input *types.DynamicConfigValue
	}{
		{
			desc:  "non-nil input test",
			input: &testdata.DynamicConfigValue,
		},
		{
			desc:  "empty input test",
			input: &types.DynamicConfigValue{},
		},
		{
			desc:  "nil input test",
			input: nil,
		},
	}
	for _, tc := range testCases {
		thriftObj := FromDynamicConfigValue(tc.input)
		roundTripObj := ToDynamicConfigValue(thriftObj)
		assert.Equal(t, tc.input, roundTripObj)
	}
}

func TestDynamicConfigFilterArray(t *testing.T) {
	testCases := []struct {
		desc  string
		input []*types.DynamicConfigFilter
	}{
		{
			desc:  "non-nil input test",
			input: []*types.DynamicConfigFilter{&testdata.DynamicConfigFilter},
		},
		{
			desc:  "empty input test",
			input: []*types.DynamicConfigFilter{},
		},
		{
			desc:  "nil input test",
			input: nil,
		},
	}
	for _, tc := range testCases {
		thriftObj := FromDynamicConfigFilterArray(tc.input)
		roundTripObj := ToDynamicConfigFilterArray(thriftObj)
		assert.Equal(t, tc.input, roundTripObj)
	}
}

func TestDynamicConfigFilter(t *testing.T) {
	testCases := []struct {
		desc  string
		input *types.DynamicConfigFilter
	}{
		{
			desc:  "non-nil input test",
			input: &testdata.DynamicConfigFilter,
		},
		{
			desc:  "empty input test",
			input: &types.DynamicConfigFilter{},
		},
		{
			desc:  "nil input test",
			input: nil,
		},
	}
	for _, tc := range testCases {
		thriftObj := FromDynamicConfigFilter(tc.input)
		roundTripObj := ToDynamicConfigFilter(thriftObj)
		assert.Equal(t, tc.input, roundTripObj)
	}
}
