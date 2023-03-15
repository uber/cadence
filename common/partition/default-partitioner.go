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
	"sort"

	"github.com/dgryski/go-farm"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/types"
)

const (
	defaultPartitionConfigWorkerIsolationGroup = "worker-isolation-group"
	defaultPartitionConfigRunID                = "wf-run-id"
)

// DefaultWorkflowPartitionConfig Is the default dataset expected to be passed around in the
// execution records for workflows which is used for partitioning. It contains the IsolationGroup
// where the workflow was started, and is expected to be pinned, and a RunID for a fallback means
// to partition data deterministically.
type defaultWorkflowPartitionConfig struct {
	WorkflowStartIsolationGroup IsolationGroupName
	RunID                       string
}

// defaultPartitioner is a business-agnostic implementation of partitioning
// which is used by the Cadence system to allocate workflows in matching by isolation-group
type defaultPartitioner struct {
	config              Config
	log                 log.Logger
	domainCache         cache.DomainCache
	isolationGroupState IsolationGroupState
}

func NewdefaultPartitioner(
	logger log.Logger,
	isolationGroupState IsolationGroupState,
	cfg Config,
) Partitioner {
	return &defaultPartitioner{
		log:                 logger,
		config:              cfg,
		isolationGroupState: isolationGroupState,
	}
}

func (r *defaultPartitioner) IsDrained(ctx context.Context, domain string, isolationGroup IsolationGroupName) (bool, error) {
	state, err := r.isolationGroupState.Get(ctx, domain)
	if err != nil {
		return false, fmt.Errorf("could not determine if drained: %w", err)
	}
	return isDrained(isolationGroup, state.Global, state.Domain), nil
}

func (r *defaultPartitioner) IsDrainedByDomainID(ctx context.Context, domainID string, isolationGroup IsolationGroupName) (bool, error) {
	domain, err := r.domainCache.GetDomainByID(domainID)
	if err != nil {
		return false, fmt.Errorf("could not determine if drained: %w", err)
	}
	return r.IsDrained(ctx, domain.GetInfo().Name, isolationGroup)
}

func (r *defaultPartitioner) GetIsolationGroupByDomainID(ctx context.Context, domainID string, wfPartitionData PartitionConfig) (*IsolationGroupName, error) {
	if !r.config.zonalPartitioningEnabledGlobally(domainID) {
		return nil, nil
	}

	wfPartition := mapPartitionConfigToDefaultPartitionConfig(wfPartitionData)
	state, err := r.isolationGroupState.GetByDomainID(ctx, domainID)
	if err != nil {
		return nil, fmt.Errorf("unable to get isolation group state: %w", err)
	}
	available := availableIG(r.config.allIsolationGroups, state.Global, state.Domain)

	ig := pickIsolationGroup(wfPartition, available)
	return &ig, nil
}

func isDrained(isolationGroup IsolationGroupName, global types.IsolationGroupConfiguration, domain types.IsolationGroupConfiguration) bool {
	globalCfg, hasGlobalConfig := global[string(isolationGroup)]
	domainCfg, hasDomainConfig := domain[string(isolationGroup)]
	if hasGlobalConfig {
		if globalCfg.State == types.IsolationGroupStateDrained {
			return true
		}
	}
	if hasDomainConfig {
		if domainCfg.State == types.IsolationGroupStateDrained {
			return true
		}
	}
	return false
}

// A simple explicit deny-based isolation group implementation
func availableIG(all []IsolationGroupName, global types.IsolationGroupConfiguration, domain types.IsolationGroupConfiguration) types.IsolationGroupConfiguration {
	out := types.IsolationGroupConfiguration{}
	for _, isolationGroup := range all {
		globalCfg, hasGlobalConfig := global[string(isolationGroup)]
		domainCfg, hasDomainConfig := domain[string(isolationGroup)]
		if hasGlobalConfig {
			if globalCfg.State == types.IsolationGroupStateDrained {
				continue
			}
		}
		if hasDomainConfig {
			if domainCfg.State == types.IsolationGroupStateDrained {
				continue
			}
		}
		out[isolationGroup] = types.IsolationGroupPartition{
			Name:   isolationGroup,
			Status: types.IsolationGroupStateHealthy,
		}
	}
	return out
}

func mapPartitionConfigToDefaultPartitionConfig(config PartitionConfig) defaultWorkflowPartitionConfig {
	isolationGroup, _ := config[defaultPartitionConfigWorkerIsolationGroup]
	runID, _ := config[defaultPartitionConfigRunID]
	return defaultWorkflowPartitionConfig{
		WorkflowStartIsolationGroup: IsolationGroupName(isolationGroup),
		RunID:                       runID,
	}
}

// picks an isolation group to run in. if the workflow was started there, it'll attempt to pin it, unless there is an explicit
// drain.
func pickIsolationGroup(wfPartition defaultWorkflowPartitionConfig, available types.IsolationGroupConfiguration) IsolationGroupName {
	wfIG, isAvailable := available[wfPartition.WorkflowStartIsolationGroup]
	if isAvailable && wfIG.Status != types.IsolationGroupStateDrained {
		return wfPartition.WorkflowStartIsolationGroup
	}

	// it's drained, fall back to picking a deterministic but random group
	availableList := []IsolationGroupName{}
	for k, v := range available {
		if v.Status == types.IsolationGroupStateDrained {
			continue
		}
		availableList = append(availableList, k)
	}
	// sort the slice to ensure it's deterministic
	sort.Slice(availableList, func(i int, j int) bool {
		return availableList[i] > availableList[j]
	})
	return pickIsolationGroupFallback(availableList, wfPartition)
}

// Simple deterministic isolationGroup picker
// which will pick a random healthy isolationGroup and place the workflow there
func pickIsolationGroupFallback(available []IsolationGroupName, wfConfig defaultWorkflowPartitionConfig) IsolationGroupName {
	hashv := farm.Hash32([]byte(wfConfig.RunID))
	return available[int(hashv)%len(available)]
}
