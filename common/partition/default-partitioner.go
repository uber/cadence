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

	"github.com/uber/cadence/common/isolationgroup"

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
	WorkflowStartIsolationGroup string
	RunID                       string
}

// defaultPartitioner is a business-agnostic implementation of partitioning
// which is used by the Cadence system to allocate workflows in matching by isolation-group
type defaultPartitioner struct {
	config              Config
	log                 log.Logger
	domainCache         cache.DomainCache
	isolationGroupState isolationgroup.State
}

func NewDefaultPartitioner(
	logger log.Logger,
	isolationGroupState isolationgroup.State,
	cfg Config,
) Partitioner {
	return &defaultPartitioner{
		log:                 logger,
		config:              cfg,
		isolationGroupState: isolationGroupState,
	}
}

func (r *defaultPartitioner) GetIsolationGroupByDomainID(ctx context.Context, domainID string, wfPartitionData PartitionConfig) (*string, error) {
	if !r.config.IsolationGroupEnabled(domainID) {
		return nil, nil
	}

	wfPartition := mapPartitionConfigToDefaultPartitionConfig(wfPartitionData)
	available, err := r.isolationGroupState.AvailableIsolationGroupsByDomainID(ctx, domainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get available isolation groups: %w", err)
	}
	ig := pickIsolationGroup(wfPartition, available)
	return &ig, nil
}

func mapPartitionConfigToDefaultPartitionConfig(config PartitionConfig) defaultWorkflowPartitionConfig {
	isolationGroup, _ := config[defaultPartitionConfigWorkerIsolationGroup]
	runID, _ := config[defaultPartitionConfigRunID]
	return defaultWorkflowPartitionConfig{
		WorkflowStartIsolationGroup: isolationGroup,
		RunID:                       runID,
	}
}

// picks an isolation group to run in. if the workflow was started there, it'll attempt to pin it, unless there is an explicit
// drain.
func pickIsolationGroup(wfPartition defaultWorkflowPartitionConfig, available types.IsolationGroupConfiguration) string {
	_, isAvailable := available[wfPartition.WorkflowStartIsolationGroup]
	if isAvailable {
		return wfPartition.WorkflowStartIsolationGroup
	}

	// it's drained, fall back to picking a deterministic but random group
	var availableList []string
	for k := range available {
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
func pickIsolationGroupFallback(available []string, wfConfig defaultWorkflowPartitionConfig) string {
	hashv := farm.Hash32([]byte(wfConfig.RunID))
	return available[int(hashv)%len(available)]
}
