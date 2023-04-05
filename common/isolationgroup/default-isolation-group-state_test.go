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

package isolationgroup

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/types"
)

func TestAvailableIsolationGroups(t *testing.T) {

	igA := string("isolationGroupA")
	igB := string("isolationGroupB")
	igC := string("isolationGroupC")

	all := []string{igA, igB, igC}

	isolationGroupsAllHealthy := types.IsolationGroupConfiguration{
		igA: {
			Name:  igA,
			State: types.IsolationGroupStateHealthy,
		},
		igB: {
			Name:  igB,
			State: types.IsolationGroupStateHealthy,
		},
		igC: {
			Name:  igC,
			State: types.IsolationGroupStateHealthy,
		},
	}

	isolationGroupsSetB := types.IsolationGroupConfiguration{
		igA: {
			Name:  igA,
			State: types.IsolationGroupStateHealthy,
		},
		igB: {
			Name:  igB,
			State: types.IsolationGroupStateHealthy,
		},
	}

	isolationGroupsSetC := types.IsolationGroupConfiguration{
		igC: {
			Name:  igC,
			State: types.IsolationGroupStateDrained,
		},
	}

	isolationGroupsSetBDrained := types.IsolationGroupConfiguration{
		igA: {
			Name:  igA,
			State: types.IsolationGroupStateDrained,
		},
		igB: {
			Name:  igB,
			State: types.IsolationGroupStateDrained,
		},
	}

	tests := map[string]struct {
		globalIGCfg types.IsolationGroupConfiguration
		domainIGCfg types.IsolationGroupConfiguration
		expected    types.IsolationGroupConfiguration
	}{
		"default behaviour - no drains - everything should be healthy": {
			globalIGCfg: types.IsolationGroupConfiguration{},
			domainIGCfg: types.IsolationGroupConfiguration{},
			expected:    isolationGroupsAllHealthy,
		},
		"default behaviour - one is not healthy - should return remaining 1/2": {
			globalIGCfg: types.IsolationGroupConfiguration{},
			domainIGCfg: isolationGroupsSetC, // C is drained
			expected:    isolationGroupsSetB, // A and B
		},
		"default behaviour - one is not healthy - should return remaining 2/2": {
			globalIGCfg: isolationGroupsSetC, // C is drained
			domainIGCfg: types.IsolationGroupConfiguration{},
			expected:    isolationGroupsSetB, // A and B
		},
		"both": {
			globalIGCfg: isolationGroupsSetC,                 // C is drained
			domainIGCfg: isolationGroupsSetBDrained,          // A, B
			expected:    types.IsolationGroupConfiguration{}, // nothing should be available
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expected, availableIG(all, td.globalIGCfg, td.domainIGCfg))
		})
	}
}

func TestIsDrained(t *testing.T) {

	igA := "isolationGroupA"
	igB := "isolationGroupB"

	isolationGroupsAllHealthy := types.IsolationGroupConfiguration{
		igA: {
			Name:  igA,
			State: types.IsolationGroupStateHealthy,
		},
		igB: {
			Name:  igB,
			State: types.IsolationGroupStateHealthy,
		},
	}

	isolationGroupsOneDrain := types.IsolationGroupConfiguration{
		igA: {
			Name:  igA,
			State: types.IsolationGroupStateDrained,
		},
	}

	tests := map[string]struct {
		globalIGCfg    types.IsolationGroupConfiguration
		domainIGCfg    types.IsolationGroupConfiguration
		isolationGroup string
		expected       bool
	}{
		"default behaviour - no drains - isolationGroup is specified": {
			globalIGCfg:    isolationGroupsAllHealthy,
			domainIGCfg:    isolationGroupsAllHealthy,
			isolationGroup: igA,
			expected:       false,
		},
		"default behaviour - no drains - isolationGroup is not specified": {
			globalIGCfg:    isolationGroupsAllHealthy,
			domainIGCfg:    isolationGroupsAllHealthy,
			isolationGroup: "some-not-specified-drain",
			expected:       false,
		},
		"default behaviour - globalDrain": {
			globalIGCfg:    isolationGroupsOneDrain,
			domainIGCfg:    isolationGroupsAllHealthy,
			isolationGroup: igA,
			expected:       true,
		},
		"default behaviour - domainDrain": {
			globalIGCfg:    isolationGroupsAllHealthy,
			domainIGCfg:    isolationGroupsOneDrain,
			isolationGroup: igA,
			expected:       true,
		},
		"default behaviour - both ": {
			globalIGCfg:    isolationGroupsOneDrain,
			domainIGCfg:    isolationGroupsOneDrain,
			isolationGroup: igA,
			expected:       true,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expected, isDrained(td.isolationGroup, td.globalIGCfg, td.domainIGCfg))
		})
	}
}

func TestIsolationGroupStateMapping(t *testing.T) {

	tests := map[string]struct {
		in          []*types.DynamicConfigEntry
		expected    types.IsolationGroupConfiguration
		expectedErr error
	}{
		"valid mapping": {
			in: []*types.DynamicConfigEntry{
				&types.DynamicConfigEntry{
					Name: "",
					Values: []*types.DynamicConfigValue{
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeJSON.Ptr(),
								Data:         []byte(`{"Name":"zone-1","State":1}`),
							},
						},
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeJSON.Ptr(),
								Data:         []byte(`{"Name":"zone-2","State":2}`),
							},
						},
					},
				},
			},
			expected: map[string]types.IsolationGroupPartition{
				"zone-1": {
					Name:  "zone-1",
					State: types.IsolationGroupStateHealthy,
				},
				"zone-2": {
					Name:  "zone-2",
					State: types.IsolationGroupStateDrained,
				},
			},
		},
		"invalid values - encoding": {
			in: []*types.DynamicConfigEntry{
				&types.DynamicConfigEntry{
					Name: "",
					Values: []*types.DynamicConfigValue{
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeThriftRW.Ptr(), //invalid, not expected to use thrift here
								Data:         []byte(`{"Name":"zone-1","State":2}`),
							},
						},
					},
				},
			},
			expectedErr: errors.New("failed to decode values: &{ThriftRW [123 34 78 97 109 101 34 58 34 122 111 110 101 45 49 34 44 34 83 116 97 116 101 34 58 50 125]}, (*types.EncodingType)"),
		},
		"invalid encoding - incomplete input - 1": {
			in:       []*types.DynamicConfigEntry{},
			expected: map[string]types.IsolationGroupPartition{},
		},
		"invalid encoding - incomplete input - 2": {
			in: []*types.DynamicConfigEntry{
				{
					Values: nil,
				},
			},
			expected: map[string]types.IsolationGroupPartition{},
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := mapDynamicConfigResponse(td.in)
			assert.Equal(t, td.expected, res)
			assert.Equal(t, td.expectedErr, err)
		})
	}
}

func TestMapAllIsolationGroupStates(t *testing.T) {

	tests := map[string]struct {
		in          []interface{}
		expected    []string
		expectedErr error
	}{
		"valid mapping": {
			in:       []interface{}{"zone-1", "zone-2", "zone-3"},
			expected: []string{"zone-1", "zone-2", "zone-3"},
		},
		"invalid mapping": {
			in:          []interface{}{1, 2, 3},
			expectedErr: errors.New("failed to get all-isolation-groups resonse from dynamic config: got 1 (int)"),
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := mapAllIsolationGroupsResponse(td.in)
			assert.Equal(t, td.expected, res)
			assert.Equal(t, td.expectedErr, err)
		})
	}
}

func TestUpdateRequest(t *testing.T) {

	json := types.EncodingTypeJSON

	tests := map[string]struct {
		in          types.IsolationGroupConfiguration
		expected    []*types.DynamicConfigValue
		expectedErr error
	}{
		"valid mapping": {
			in: types.IsolationGroupConfiguration{
				"zone-1": {
					Name:  "zone-1",
					State: types.IsolationGroupStateHealthy,
				},
				"zone-2": {
					Name:  "zone-2",
					State: types.IsolationGroupStateDrained,
				},
			},
			expected: []*types.DynamicConfigValue{
				{
					Value: &types.DataBlob{
						EncodingType: &json,
						Data:         []byte(`{"Name":"zone-1","State":1}`),
					},
					Filters: nil,
				},
				{
					Value: &types.DataBlob{
						EncodingType: &json,
						Data:         []byte(`{"Name":"zone-2","State":2}`),
					},
					Filters: nil,
				},
			},
		},
		"empty mapping": {
			in:       types.IsolationGroupConfiguration{},
			expected: nil,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := mapUpdateGlobalIsolationGroupsRequest(td.in)
			assert.Equal(t, td.expected, res)
			assert.Equal(t, td.expectedErr, err)
		})
	}
}

func TestIsolationGroupShutdown(t *testing.T) {
	var v defaultIsolationGroupStateHandler
	assert.NotPanics(t, func() {
		v.Stop()
	})
}
