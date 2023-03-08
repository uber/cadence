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
	"github.com/uber/cadence/common/persistence"

	"github.com/dgryski/go-farm"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/types"
)

const (
	DefaultPartitionConfigWorkerZone = "worker-zone"
	DefaultPartitionConfigRunID      = "wf-run-id"
)

// DefaultPartitionConfig is the open-source default Partition configuration
// for each workflow which ensures that workflows started in the same zone remain there.
type DefaultPartitionConfig struct {
	WorkflowStartZone types.ZoneName
	RunID             string
}

type DefaultPartitioner struct {
	config    Config
	log       log.Logger
	zoneState ZoneState
}

type DefaultZoneStateHandler struct {
	log              log.Logger
	domainCache      cache.DomainCache
	globalZoneDrains persistence.GlobalZoneDrains
	config           Config
}

func NewDefaultZoneStateWatcher(
	logger log.Logger,
	domainCache cache.DomainCache,
	config Config,
	globalZoneDrains persistence.GlobalZoneDrains,
) ZoneState {
	return &DefaultZoneStateHandler{
		log:              logger,
		config:           config,
		domainCache:      domainCache,
		globalZoneDrains: globalZoneDrains,
	}
}

func NewDefaultTaskResolver(
	logger log.Logger,
	zoneState ZoneState,
	cfg Config,
) Partitioner {
	return &DefaultPartitioner{
		log:       logger,
		config:    cfg,
		zoneState: zoneState,
	}
}

func (r *DefaultPartitioner) IsDrained(ctx context.Context, domain string, zone types.ZoneName) (bool, error) {
	state, err := r.zoneState.Get(ctx, domain, zone)
	if err != nil {
		return false, fmt.Errorf("could not determine if drained: %w", err)
	}
	return state.Status == types.ZoneStatusDrained, nil
}

func (r *DefaultPartitioner) IsDrainedByDomainID(ctx context.Context, domainID string, zone types.ZoneName) (bool, error) {
	state, err := r.zoneState.GetByDomainID(ctx, domainID, zone)
	if err != nil {
		return false, fmt.Errorf("could not determine if drained: %w", err)
	}
	return state.Status == types.ZoneStatusDrained, nil
}

func (r *DefaultPartitioner) GetTaskZoneByDomainID(ctx context.Context, DomainID string, key types.PartitionConfig) (*types.ZoneName, error) {
	partitionData := mapPartitionConfigToDefaultPartitionConfig(key)

	isDrained, err := r.IsDrainedByDomainID(ctx, DomainID, partitionData.WorkflowStartZone)
	if err != nil {
		return nil, fmt.Errorf("failed to determine if a zone is drained: %w", err)
	}

	if isDrained {
		zones, err := r.zoneState.ListAll(ctx, DomainID)
		if err != nil {
			return nil, fmt.Errorf("failed to list all zones: %w", err)
		}
		zone := pickZoneAfterDrain(zones, partitionData)
		return &zone, nil
	}

	return &partitionData.WorkflowStartZone, nil
}

func (z *DefaultZoneStateHandler) ListAll(ctx context.Context, domainID string) ([]types.ZonePartition, error) {
	var out []types.ZonePartition

	for _, zone := range z.config.allZones {
		zoneData, err := z.Get(ctx, domainID, zone)
		if err != nil {
			return nil, fmt.Errorf("failed to get zone during listing: %w", err)
		}
		out = append(out, *zoneData)
	}

	return out, nil
}

func (z *DefaultZoneStateHandler) GetByDomainID(ctx context.Context, domainID string, zone types.ZoneName) (*types.ZonePartition, error) {
	domain, err := z.domainCache.GetDomainByID(domainID)
	if err != nil {
		return nil, fmt.Errorf("could not resolve domain in zone handler: %w", err)
	}
	return z.Get(ctx, domain.GetInfo().Name, zone)
}

// Get the statue of a zone, with respect to both domain and global drains. Domain-specific drains override global config
func (z *DefaultZoneStateHandler) Get(ctx context.Context, domain string, zone types.ZoneName) (*types.ZonePartition, error) {
	if !z.config.zonalPartitioningEnabled(domain) {
		return &types.ZonePartition{
			Name:   zone,
			Status: types.ZoneStatusHealthy,
		}, nil
	}

	domainData, err := z.domainCache.GetDomain(domain)
	if err != nil {
		return nil, fmt.Errorf("could not resolve domain in zone handler: %w", err)
	}
	cfg, ok := domainData.GetInfo().ZoneConfig[zone]
	if ok && cfg.Status == types.ZoneStatusDrained {
		return &cfg, nil
	}

	drains, err := z.globalZoneDrains.GetClusterDrains(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not resolve global drains in zone handler: %w", err)
	}
	globalCfg, ok := drains[zone]
	if ok {
		return &globalCfg, nil
	}

	return &types.ZonePartition{
		Name:   zone,
		Status: types.ZoneStatusHealthy,
	}, nil
}

// Simple deterministic zone picker
// which will pick a random healthy zone and place the workflow there
func pickZoneAfterDrain(zones []types.ZonePartition, wfConfig DefaultPartitionConfig) types.ZoneName {
	var availableZones []types.ZoneName
	for _, zone := range zones {
		if zone.Status == types.ZoneStatusHealthy {
			availableZones = append(availableZones, zone.Name)
		}
	}
	hashv := farm.Hash32([]byte(wfConfig.RunID))
	return availableZones[int(hashv)%len(availableZones)]
}

func mapPartitionConfigToDefaultPartitionConfig(config types.PartitionConfig) DefaultPartitionConfig {
	zone, _ := config[DefaultPartitionConfigWorkerZone]
	runID, _ := config[DefaultPartitionConfigRunID]
	return DefaultPartitionConfig{
		WorkflowStartZone: types.ZoneName(zone),
		RunID:             runID,
	}
}
