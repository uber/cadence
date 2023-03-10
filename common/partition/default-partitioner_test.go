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

package partition

import (
	"context"
	"testing"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/persistence"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/types"
)

func TestDefaultPartitioner_GetTaskIsolationGroup(t *testing.T) {

	domainID := "4271399E-10B5-4F9E-B5D1-B538EFD73125"
	domainName := "domain-name"
	isolationGroup2 := types.IsolationGroupName("isolationGroup2")

	partitionKey := types.PartitionConfig{
		DefaultPartitionConfigWorkerIsolationGroup: string(isolationGroup2),
		DefaultPartitionConfigRunID:                "A7758FF3-CBC2-4414-B517-A37823F335D3",
	}

	tests := map[string]struct {
		featureEnabledGlobally        bool
		featureEnabledForDomain       bool
		isolationGroupDrainAffordance func(drains *persistence.MockGlobalIsolationGroupDrains)
		domainCacheAffordance         func(drains *cache.MockDomainCache)
		expected                      *types.IsolationGroupName
		expectedErr                   error
	}{
		"normal case - feature not enabled": {
			featureEnabledGlobally:  false,
			featureEnabledForDomain: false,
			isolationGroupDrainAffordance: func(drains *persistence.MockGlobalIsolationGroupDrains) {
			},
			domainCacheAffordance: func(dc *cache.MockDomainCache) {
			},
			expected: nil,
		},
		"normal case - feature is not enabled for domain, all isolationGroups should always report healthy and there should be no additional DB calls": {
			featureEnabledGlobally:  true,
			featureEnabledForDomain: false,
			isolationGroupDrainAffordance: func(drains *persistence.MockGlobalIsolationGroupDrains) {
			},
			domainCacheAffordance: func(dc *cache.MockDomainCache) {
				o := int64(0)
				ce := cache.NewDomainCacheEntryForTest(&persistence.DomainInfo{
					Name:                 domainName,
					IsolationGroupConfig: types.IsolationGroupConfiguration{},
				}, nil, true, nil, 0, &o)
				dc.EXPECT().GetDomainByID(domainID).Return(ce, nil)
			},
			expected: &isolationGroup2,
		},
		"IsolationGroups not drained, feature enabled, normal case - workflow is started in isolationGroup 2, so it should remain there": {
			featureEnabledGlobally:  true,
			featureEnabledForDomain: true,
			isolationGroupDrainAffordance: func(drains *persistence.MockGlobalIsolationGroupDrains) {
				drains.EXPECT().GetClusterDrains(gomock.Any()).Return(types.IsolationGroupConfiguration{}, nil) // empty drains
			},
			domainCacheAffordance: func(dc *cache.MockDomainCache) {
				o := int64(0)
				ce := cache.NewDomainCacheEntryForTest(&persistence.DomainInfo{
					Name:                 domainName,
					IsolationGroupConfig: types.IsolationGroupConfiguration{},
				}, nil, true, nil, 0, &o)
				dc.EXPECT().GetDomainByID(domainID).Return(ce, nil) // resolve the domainName first
				dc.EXPECT().GetDomain(domainName).Return(ce, nil)   // no domain drains
			},
			expected: &isolationGroup2,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {

			cfg := Config{
				zonalPartitioningEnabledGlobally:  func(string) bool { return td.featureEnabledGlobally },
				zonalPartitioningEnabledForDomain: func(string) bool { return td.featureEnabledForDomain },
				allIsolationGroups:                []types.IsolationGroupName{"isolationGroup1", "isolationGroup2", "isolationGroup3"},
			}

			mockCtl := gomock.NewController(t)

			isolationGroupDrainsMock := persistence.NewMockGlobalIsolationGroupDrains(mockCtl)
			cacheMock := cache.NewMockDomainCache(mockCtl)
			td.isolationGroupDrainAffordance(isolationGroupDrainsMock)
			td.domainCacheAffordance(cacheMock)

			isolationGroupState := NewDefaultIsolationGroupStateWatcher(loggerimpl.NewNopLogger(), cacheMock, cfg, isolationGroupDrainsMock)
			resolver := NewDefaultTaskPartitioner(loggerimpl.NewNopLogger(), isolationGroupState, cfg)

			res, err := resolver.GetTaskIsolationGroupByDomainID(context.Background(), domainID, partitionKey)

			assert.Equal(t, td.expectedErr, err)
			assert.Equal(t, td.expected, res)
		})
	}
}
