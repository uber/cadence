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
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

const (
	DefaultPartitionConfigWorkerIsolationGroup = "worker-isolationGroup"
	DefaultPartitionConfigRunID                = "wf-run-id"
)

type DefaultIsolationGroupState struct {
	log                        log.Logger
	domainCache                cache.DomainCache
	globalIsolationGroupDrains persistence.GlobalIsolationGroupDrains
	config                     Config
}

func NewDefaultIsolationGroupStateWatcher(
	logger log.Logger,
	domainCache cache.DomainCache,
	config Config,
	globalIsolationGroupDrains persistence.GlobalIsolationGroupDrains,
) IsolationGroupState {
	return &DefaultIsolationGroupState{
		log:                        logger,
		config:                     config,
		domainCache:                domainCache,
		globalIsolationGroupDrains: globalIsolationGroupDrains,
	}
}

func (z *DefaultIsolationGroupState) GetByDomainID(ctx context.Context, domainID string, isolationGroup types.IsolationGroupName) (*State, error) {
	domain, err := z.domainCache.GetDomainByID(domainID)
	if err != nil {
		return nil, fmt.Errorf("could not resolve domain in isolationGroup handler: %w", err)
	}
	return z.Get(ctx, domain.GetInfo().Name, isolationGroup)
}

// Get the state of a isolationGroup, with respect to both domain and global drains. Domain-specific drains override global config
func (z *DefaultIsolationGroupState) Get(ctx context.Context, domain string, isolationGroup types.IsolationGroupName) (*State, error) {
	if !z.config.zonalPartitioningEnabledForDomain(domain) {
		return &State{
			Global: types.IsolationGroupConfiguration{
				isolationGroup: types.IsolationGroupPartition{
					Name:   isolationGroup,
					Status: types.IsolationGroupStatusHealthy,
				},
			},
			Domain: types.IsolationGroupConfiguration{
				isolationGroup: types.IsolationGroupPartition{
					Name:   isolationGroup,
					Status: types.IsolationGroupStatusHealthy,
				},
			},
		}, nil
	}

	domainData, err := z.domainCache.GetDomain(domain)
	if err != nil {
		return nil, fmt.Errorf("could not resolve domain in isolationGroup handler: %w", err)
	}
	domainCfg := domainData.GetInfo().IsolationGroupConfig

	// todo (david.porter) wrap this in an in-memory cache
	global, err := z.globalIsolationGroupDrains.GetClusterDrains(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not resolve global drains in isolationGroup handler: %w", err)
	}
	return &State{
		Global: global,
		Domain: domainCfg,
	}, nil
}

func isDrained(isolationGroup types.IsolationGroupName, global types.IsolationGroupConfiguration, domain types.IsolationGroupConfiguration) bool {
	globalCfg, hasGlobalConfig := global[isolationGroup]
	domainCfg, hasDomainConfig := domain[isolationGroup]
	if hasGlobalConfig {
		if globalCfg.Status == types.IsolationGroupStatusDrained {
			return true
		}
	}
	if hasDomainConfig {
		if domainCfg.Status == types.IsolationGroupStatusDrained {
			return true
		}
	}
	return false
}

// A simple explicit deny-based isolation group implementation
func availableIG(all []types.IsolationGroupName, global types.IsolationGroupConfiguration, domain types.IsolationGroupConfiguration) types.IsolationGroupConfiguration {
	out := types.IsolationGroupConfiguration{}
	for _, isolationGroup := range all {
		globalCfg, hasGlobalConfig := global[isolationGroup]
		domainCfg, hasDomainConfig := domain[isolationGroup]
		if hasGlobalConfig {
			if globalCfg.Status == types.IsolationGroupStatusDrained {
				continue
			}
		}
		if hasDomainConfig {
			if domainCfg.Status == types.IsolationGroupStatusDrained {
				continue
			}
		}
		out[isolationGroup] = types.IsolationGroupPartition{
			Name:   isolationGroup,
			Status: types.IsolationGroupStatusHealthy,
		}
	}
	return out
}