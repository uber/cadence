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
	"github.com/uber/cadence/common/types"
)

const (
	DefaultPartitionConfigWorkerIsolationGroup = "worker-isolation-group"
	DefaultPartitionConfigRunID                = "wf-run-id"
)

// DefaultWorkflowPartitionConfig Is the default dataset expected to be passed around in the
// execution records for workflows which is used for partitioning. It contains the IsolationGroup
// where the workflow was started, and is expected to be pinned, and a RunID for a fallback means
// to partition data deterministically.
type DefaultWorkflowPartitionConfig struct {
	WorkflowStartIsolationGroup types.IsolationGroupName
	RunID                       string
}

// DefaultPartitioner is a business-agnositic implementation of partitioning
// which is used by the Cadence system to allocate workflows in matching by isolation-group
type DefaultPartitioner struct {
	config              Config
	log                 log.Logger
	domainCache         cache.DomainCache
	isolationGroupState IsolationGroupState
}

func NewDefaultPartitioner(
	logger log.Logger,
	isolationGroupState IsolationGroupState,
	cfg Config,
) Partitioner {
	return &DefaultPartitioner{
		log:                 logger,
		config:              cfg,
		isolationGroupState: isolationGroupState,
	}
}

func (r *DefaultPartitioner) IsDrained(ctx context.Context, domain string, isolationGroup types.IsolationGroupName) (bool, error) {
	state, err := r.isolationGroupState.Get(ctx, domain, isolationGroup)
	if err != nil {
		return false, fmt.Errorf("could not determine if drained: %w", err)
	}
	return isDrained(isolationGroup, state.Global, state.Domain), nil
}

func (r *DefaultPartitioner) IsDrainedByDomainID(ctx context.Context, domainID string, isolationGroup types.IsolationGroupName) (bool, error) {
	domain, err := r.domainCache.GetDomainByID(domainID)
	if err != nil {
		return false, fmt.Errorf("could not determine if drained: %w", err)
	}
	return r.IsDrained(ctx, domain.GetInfo().Name, isolationGroup)
}

func (r *DefaultPartitioner) GetTaskIsolationGroupByDomainID(ctx context.Context, domainID string, key types.PartitionConfig) (*types.IsolationGroupName, error) {
	if !r.config.zonalPartitioningEnabledGlobally(domainID) {
		return nil, nil
	}

	partitionData := mapPartitionConfigToDefaultPartitionConfig(key)

	r.isolationGroupState.GetByDomainID(ctx, domainID)

	availableIG()
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

func mapPartitionConfigToDefaultPartitionConfig(config types.PartitionConfig) DefaultWorkflowPartitionConfig {
	isolationGroup, _ := config[DefaultPartitionConfigWorkerIsolationGroup]
	runID, _ := config[DefaultPartitionConfigRunID]
	return DefaultWorkflowPartitionConfig{
		WorkflowStartIsolationGroup: types.IsolationGroupName(isolationGroup),
		RunID:                       runID,
	}
}
