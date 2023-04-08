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
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/persistence"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/types"
)

func TestIsDrainedHandler(t *testing.T) {

	validInput := types.IsolationGroupConfiguration{
		"zone-1": types.IsolationGroupPartition{
			Name:  "zone-1",
			State: types.IsolationGroupStateHealthy,
		},
		"zone-2": types.IsolationGroupPartition{
			Name:  "zone-2",
			State: types.IsolationGroupStateDrained,
		},
	}

	validCfg, _ := mapUpdateGlobalIsolationGroupsRequest(validInput)

	validCfgData := validCfg[0].Value.GetData()
	dynamicConfigResponse := []interface{}{}
	json.Unmarshal(validCfgData, &dynamicConfigResponse)

	tests := map[string]struct {
		requestIsolationgroup string
		dcAffordance          func(client *dynamicconfig.MockClient)
		domainAffordance      func(mock *cache.MockDomainCache)
		cfg                   defaultConfig
		expected              bool
		expectedErr           error
	}{
		"normal case - no drains present - no configuration specifying a drain - feature is enabled": {
			requestIsolationgroup: "zone-3", // no config specified for this
			cfg: defaultConfig{
				IsolationGroupEnabled: func(string) bool { return true },
				AllIsolationGroups:    []string{"zone-1", "zone-2", "zone-3"},
			},
			dcAffordance: func(client *dynamicconfig.MockClient) {
				client.EXPECT().GetListValue(
					dynamicconfig.DefaultIsolationGroupConfigStoreManagerGlobalMapping,
					gomock.Any(), // covering the mapping in the mapper unit-test instead
				).Return(dynamicConfigResponse, nil)
			},
			domainAffordance: func(mock *cache.MockDomainCache) {
				domainResponse := cache.NewDomainCacheEntryForTest(&persistence.DomainInfo{ID: "domain-id", Name: "domain"}, nil, true, nil, 0, nil)
				mock.EXPECT().GetDomainByID("domain-id").Return(domainResponse, nil)
				mock.EXPECT().GetDomain("domain").Return(domainResponse, nil)
			},
			expected: false,
		},
		"normal case - drains present - feature is enabled": {
			requestIsolationgroup: "zone-2", // no config specified for this
			cfg: defaultConfig{
				IsolationGroupEnabled: func(string) bool { return true },
				AllIsolationGroups:    []string{"zone-1", "zone-2", "zone-3"},
			},
			dcAffordance: func(client *dynamicconfig.MockClient) {
				client.EXPECT().GetListValue(
					dynamicconfig.DefaultIsolationGroupConfigStoreManagerGlobalMapping,
					gomock.Any(), // covering the mapping in the mapper unit-test instead
				).Return(dynamicConfigResponse, nil)
			},
			domainAffordance: func(mock *cache.MockDomainCache) {
				domainResponse := cache.NewDomainCacheEntryForTest(&persistence.DomainInfo{ID: "domain-id", Name: "domain"}, nil, true, nil, 0, nil)
				mock.EXPECT().GetDomainByID("domain-id").Return(domainResponse, nil)
				mock.EXPECT().GetDomain("domain").Return(domainResponse, nil)
			},
			expected: true,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			mockCtl := gomock.NewController(t)
			dcMock := dynamicconfig.NewMockClient(mockCtl)
			domaincacheMock := cache.NewMockDomainCache(mockCtl)
			td.dcAffordance(dcMock)
			td.domainAffordance(domaincacheMock)
			handler := defaultIsolationGroupStateHandler{
				log:                        loggerimpl.NewNopLogger(),
				globalIsolationGroupDrains: dcMock,
				domainCache:                domaincacheMock,
				config:                     td.cfg,
			}
			res, err := handler.IsDrainedByDomainID(context.Background(), "domain-id", td.requestIsolationgroup)
			assert.Equal(t, td.expected, res)
			assert.Equal(t, td.expectedErr, err)
		})
	}
}

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

func TestUpdateGlobalState(t *testing.T) {

	tests := map[string]struct {
		in           types.UpdateGlobalIsolationGroupsRequest
		dcAffordance func(client *dynamicconfig.MockClient)
		expectedErr  error
	}{
		"updating normal value": {
			in: types.UpdateGlobalIsolationGroupsRequest{IsolationGroups: types.IsolationGroupConfiguration{
				"zone-1": {Name: "zone-1", State: types.IsolationGroupStateHealthy},
				"zone-2": {Name: "zone-2", State: types.IsolationGroupStateDrained},
			}},
			dcAffordance: func(client *dynamicconfig.MockClient) {
				client.EXPECT().UpdateValue(
					dynamicconfig.DefaultIsolationGroupConfigStoreManagerGlobalMapping,
					gomock.Any(), // covering the mapping in the mapper unit-test instead
				)
			},
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			dcMock := dynamicconfig.NewMockClient(gomock.NewController(t))
			td.dcAffordance(dcMock)
			handler := defaultIsolationGroupStateHandler{
				log:                        loggerimpl.NewNopLogger(),
				globalIsolationGroupDrains: dcMock,
			}
			err := handler.UpdateGlobalState(context.Background(), td.in)
			assert.Equal(t, td.expectedErr, err)
		})
	}
}

func TestGetGlobalState(t *testing.T) {

	validInput := types.IsolationGroupConfiguration{
		"zone-1": types.IsolationGroupPartition{
			Name:  "zone-1",
			State: types.IsolationGroupStateDrained,
		},
		"zone-2": types.IsolationGroupPartition{
			Name:  "zone-2",
			State: types.IsolationGroupStateHealthy,
		},
	}

	validCfg, _ := mapUpdateGlobalIsolationGroupsRequest(validInput)

	validCfgData := validCfg[0].Value.GetData()
	dynamicConfigResponse := []interface{}{}
	json.Unmarshal(validCfgData, &dynamicConfigResponse)

	tests := map[string]struct {
		in           types.GetGlobalIsolationGroupsRequest
		dcAffordance func(client *dynamicconfig.MockClient)
		expected     *types.GetGlobalIsolationGroupsResponse
		expectedErr  error
	}{
		"updating normal value": {
			in: types.GetGlobalIsolationGroupsRequest{},
			dcAffordance: func(client *dynamicconfig.MockClient) {
				client.EXPECT().GetListValue(
					dynamicconfig.DefaultIsolationGroupConfigStoreManagerGlobalMapping,
					gomock.Any(),
				).Return(dynamicConfigResponse, nil)
			},
			expected: &types.GetGlobalIsolationGroupsResponse{IsolationGroups: validInput},
		},
		"an error was returned": {
			in: types.GetGlobalIsolationGroupsRequest{},
			dcAffordance: func(client *dynamicconfig.MockClient) {
				client.EXPECT().GetListValue(
					dynamicconfig.DefaultIsolationGroupConfigStoreManagerGlobalMapping,
					gomock.Any(),
				).Return(nil, errors.New("an error"))
			},
			expectedErr: errors.New("failed to get global isolation groups from datastore: an error"),
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			dcMock := dynamicconfig.NewMockClient(gomock.NewController(t))
			td.dcAffordance(dcMock)
			handler := defaultIsolationGroupStateHandler{
				log:                        loggerimpl.NewNopLogger(),
				globalIsolationGroupDrains: dcMock,
			}
			res, err := handler.GetGlobalState(context.Background())
			assert.Equal(t, td.expected, res)
			if td.expectedErr != nil {
				// only compare strings because full wrapping makes it fiddly otherwise
				assert.Equal(t, td.expectedErr.Error(), err.Error())
			}
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

	z1 := types.IsolationGroupPartition{
		Name:  "zone-1",
		State: types.IsolationGroupStateHealthy,
	}

	z2 := types.IsolationGroupPartition{
		Name:  "zone-2",
		State: types.IsolationGroupStateDrained,
	}

	// JSON serialization inside the dynamic config library makes a mess of things because
	// it doesn't have any type information. So mimicing this type information loss to ensure
	// any field name types or similar serialization quirks are picked up by the test
	zMarshalled, _ := json.Marshal([]types.IsolationGroupPartition{z1, z2})
	var rawIsolationGroupDataMarshalled []interface{}
	json.Unmarshal(zMarshalled, &rawIsolationGroupDataMarshalled)

	tests := map[string]struct {
		in          []interface{}
		expected    types.IsolationGroupConfiguration
		expectedErr error
	}{
		"valid mapping": {
			in: rawIsolationGroupDataMarshalled,
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
		"empty mapping": {
			in:       nil,
			expected: nil,
		},
		"invalid mapping 1": {
			in:          []interface{}{"invalid"},
			expectedErr: errors.New("failed parse a dynamic config entry, map[], (got invalid)"),
		},
		"invalid mapping 2": {
			in:          []interface{}{`{"Name": "some zone", "State": "not the right type"}`},
			expectedErr: errors.New(`failed parse a dynamic config entry, map[], (got {"Name": "some zone", "State": "not the right type"})`),
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
			expectedErr: errors.New("failed to get all-isolation-groups response from dynamic config: got 1 (int)"),
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
						EncodingType: types.EncodingTypeJSON.Ptr(),
						Data:         []byte(`[{"Name":"zone-1","State":1},{"Name":"zone-2","State":2}]`),
					},
					Filters: nil,
				},
			},
		},
		"empty mapping": {
			in: types.IsolationGroupConfiguration{},
			expected: []*types.DynamicConfigValue{
				{
					Value: &types.DataBlob{
						EncodingType: types.EncodingTypeJSON.Ptr(),
						Data:         []byte(`null`),
					},
					Filters: nil,
				},
			},
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
