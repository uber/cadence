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
