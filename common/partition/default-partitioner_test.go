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
	"fmt"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/persistence"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/types"
)

func TestDefaultPartitionerZonePicking(t *testing.T) {

	zonesAllHealthy := []types.ZonePartition{
		{
			Name:   "zone1",
			Status: types.ZoneStatusHealthy,
		},
		{
			Name:   "zone2",
			Status: types.ZoneStatusHealthy,
		},
		{
			Name:   "zone3",
			Status: types.ZoneStatusHealthy,
		},
	}

	zonesWithOneDrain := []types.ZonePartition{
		{
			Name:   "zone1",
			Status: types.ZoneStatusHealthy,
		},
		{
			Name:   "zone2",
			Status: types.ZoneStatusDrained,
		},
		{
			Name:   "zone3",
			Status: types.ZoneStatusHealthy,
		},
	}

	tests := map[string]struct {
		zones    []types.ZonePartition
		wfConfig DefaultPartitionConfig
		expected types.ZoneName
	}{
		"default behaviour - no drains - should start in the zone it's created in 1/3": {
			zones: zonesAllHealthy,
			wfConfig: DefaultPartitionConfig{
				WorkflowStartZone: "zone1",
				RunID:             "d1254200-e457-4998-a2ba-d69c8f59a75f",
			},
			expected: "zone1",
		},
		"default behaviour - no drains - should start in the zone it's created in 2/3": {
			zones: zonesAllHealthy,
			wfConfig: DefaultPartitionConfig{
				WorkflowStartZone: "zone2",
				RunID:             "d1254200-e457-4998-a2ba-d69c8f59a75f",
			},
			expected: "zone2",
		},
		"default behaviour - no drains - should start in the zone it's created in 3/3": {
			zones: zonesAllHealthy,
			wfConfig: DefaultPartitionConfig{
				WorkflowStartZone: "zone3",
				RunID:             "d1254200-e457-4998-a2ba-d69c8f59a75f",
			},
			expected: "zone3",
		},
		"failover behaviour - should be able to fallback to a random deterministic partition if the zone isn't found": {
			zones: zonesWithOneDrain,
			wfConfig: DefaultPartitionConfig{
				WorkflowStartZone: "zone2",
				RunID:             "d1254200-e457-4998-a2ba-d69c8f59a75f",
			},
			expected: "zone3",
		},
		"regional failover case - should be able to fallback to a random deterministic partition if the zone isn't found": {
			zones: zonesAllHealthy,
			wfConfig: DefaultPartitionConfig{
				WorkflowStartZone: "not-found-zone",
				RunID:             "d1254200-e457-4998-a2ba-d69c8f59a75f",
			},
			expected: "zone1",
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expected, pickZoneAfterDrain(td.zones, td.wfConfig))
		})
	}
}

func TestDefaultPartitionerPickerDistribution(t *testing.T) {

	countHealthy := make(map[types.ZoneName]int)

	var zones []types.ZonePartition

	for i := 0; i < 100; i++ {
		healthy := types.ZonePartition{
			Name:   types.ZoneName(fmt.Sprintf("zone-%d", i)),
			Status: types.ZoneStatusHealthy,
		}

		unhealthy := types.ZonePartition{
			Name:   types.ZoneName(fmt.Sprintf("zone-%d", i)),
			Status: types.ZoneStatusHealthy,
		}

		zones = append(zones, healthy)
		zones = append(zones, unhealthy)
		countHealthy[healthy.Name] = 0
	}

	for i := 0; i < 100000; i++ {
		result := pickZoneAfterDrain(zones, DefaultPartitionConfig{
			WorkflowStartZone: "not-a-present-zone", // always force a fallback to the simple hash
			RunID:             uuid.New().String(),
		})

		count, ok := countHealthy[result]
		if !ok {
			t.Fatal("the result wasn't found in the healthy list, something is wrong with the logic for selecting healthy zones")
		}
		countHealthy[result] = count + 1
	}

	for k, v := range countHealthy {
		assert.True(t, v > 0, "failed to pick a zone %s", k)
	}
}

func TestDefaultPartitioner_GetTaskZone(t *testing.T) {

	domainID := "4271399E-10B5-4F9E-B5D1-B538EFD73125"
	domainName := "domain-name"
	zone2 := types.ZoneName("zone2")

	partitionKey := types.PartitionConfig{
		DefaultPartitionConfigWorkerZone: string(zone2),
		DefaultPartitionConfigRunID:      "A7758FF3-CBC2-4414-B517-A37823F335D3",
	}

	tests := map[string]struct {
		featureEnabled        bool
		zoneDrainAffordance   func(drains *persistence.MockGlobalZoneDrains)
		domainCacheAffordance func(drains *cache.MockDomainCache)
		expected              *types.ZoneName
		expectedErr           error
	}{
		"Zones not drained, normal case - feature is not enabled, all zones should always report healthy and there should be no additional DB calls": {
			featureEnabled: false,
			zoneDrainAffordance: func(drains *persistence.MockGlobalZoneDrains) {
			},
			domainCacheAffordance: func(dc *cache.MockDomainCache) {
				o := int64(0)
				ce := cache.NewDomainCacheEntryForTest(&persistence.DomainInfo{
					Name:       domainName,
					ZoneConfig: types.ZoneConfiguration{},
				}, nil, true, nil, 0, &o)
				dc.EXPECT().GetDomainByID(domainID).Return(ce, nil)
			},
			expected: &zone2,
		},
		"Zones not drained, normal case - workflow is started in zone 2, so it should remain there": {
			featureEnabled: true,
			zoneDrainAffordance: func(drains *persistence.MockGlobalZoneDrains) {
				drains.EXPECT().GetClusterDrains(gomock.Any()).Return(types.ZoneConfiguration{}, nil) // empty drains
			},
			domainCacheAffordance: func(dc *cache.MockDomainCache) {
				o := int64(0)
				ce := cache.NewDomainCacheEntryForTest(&persistence.DomainInfo{
					Name:       domainName,
					ZoneConfig: types.ZoneConfiguration{},
				}, nil, true, nil, 0, &o)
				dc.EXPECT().GetDomainByID(domainID).Return(ce, nil) // resolve the domainName first
				dc.EXPECT().GetDomain(domainName).Return(ce, nil)   // no domain drains
			},
			expected: &zone2,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {

			cfg := Config{
				zonalPartitioningEnabled: func(string) bool { return td.featureEnabled },
				allZones:                 []types.ZoneName{"zone1", "zone2", "zone3"},
			}

			mockCtl := gomock.NewController(t)

			zoneDrainsMock := persistence.NewMockGlobalZoneDrains(mockCtl)
			cacheMock := cache.NewMockDomainCache(mockCtl)
			td.zoneDrainAffordance(zoneDrainsMock)
			td.domainCacheAffordance(cacheMock)

			zoneState := NewDefaultZoneStateWatcher(loggerimpl.NewNopLogger(), cacheMock, cfg, zoneDrainsMock)
			resolver := NewDefaultTaskResolver(loggerimpl.NewNopLogger(), zoneState, cfg)

			res, err := resolver.GetTaskZoneByDomainID(context.Background(), domainID, partitionKey)

			assert.Equal(t, td.expectedErr, err)
			assert.Equal(t, td.expected, res)
		})
	}
}
