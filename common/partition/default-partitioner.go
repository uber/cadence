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
	"github.com/dgryski/go-farm"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/types"
)

// DefaultPartitionConfig is the open-source default Partition configuration
// for each workflow which ensures that workflows started in the same isolationGroup remain there.
type DefaultPartitionConfig struct {
	WorkflowStartIsolationGroup types.IsolationGroupName
	RunID                       string
}

type DefaultPartitioner struct {
	config              Config
	log                 log.Logger
	domainCache         cache.DomainCache
	isolationGroupState IsolationGroupState
}

func NewDefaultTaskPartitioner(
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
	return state.Status == types.IsolationGroupStatusDrained, nil
}

func (r *DefaultPartitioner) IsDrainedByDomainID(ctx context.Context, domainID string, isolationGroup types.IsolationGroupName) (bool, error) {
	state, err := r.isolationGroupState.GetByDomainID(ctx, domainID, isolationGroup)
	if err != nil {
		return false, fmt.Errorf("could not determine if drained: %w", err)
	}
	return state.Status == types.IsolationGroupStatusDrained, nil
}

func (r *DefaultPartitioner) GetTaskIsolationGroupByDomainID(ctx context.Context, domainID string, key types.PartitionConfig) (*types.IsolationGroupName, error) {
	if !r.config.zonalPartitioningEnabledGlobally(domainID) {
		return nil, nil
	}

	partitionData := mapPartitionConfigToDefaultPartitionConfig(key)

	isDrained, err := r.IsDrainedByDomainID(ctx, domainID, partitionData.WorkflowStartIsolationGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to determine if a isolationGroup is drained: %w", err)
	}

	if isDrained {
		isolationGroups, err := r.isolationGroupState.ListAll(ctx, domainID)
		if err != nil {
			return nil, fmt.Errorf("failed to list all isolationGroups: %w", err)
		}
		isolationGroup := pickIsolationGroupAfterDrain(isolationGroups, partitionData)
		return &isolationGroup, nil
	}

	return &partitionData.WorkflowStartIsolationGroup, nil
}

// Simple deterministic isolationGroup picker
// which will pick a random healthy isolationGroup and place the workflow there
func pickIsolationGroupAfterDrain(isolationGroups []types.IsolationGroupPartition, wfConfig DefaultPartitionConfig) types.IsolationGroupName {
	var availableIsolationGroups []types.IsolationGroupName
	for _, isolationGroup := range isolationGroups {
		if isolationGroup.Status == types.IsolationGroupStatusHealthy {
			availableIsolationGroups = append(availableIsolationGroups, isolationGroup.Name)
		}
	}
	hashv := farm.Hash32([]byte(wfConfig.RunID))
	return availableIsolationGroups[int(hashv)%len(availableIsolationGroups)]
}

func mapPartitionConfigToDefaultPartitionConfig(config types.PartitionConfig) DefaultPartitionConfig {
	isolationGroup, _ := config[DefaultPartitionConfigWorkerIsolationGroup]
	runID, _ := config[DefaultPartitionConfigRunID]
	return DefaultPartitionConfig{
		WorkflowStartIsolationGroup: types.IsolationGroupName(isolationGroup),
		RunID:                       runID,
	}
}
