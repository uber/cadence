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
	"errors"
	"fmt"
	"slices"

	"github.com/uber/cadence/common/isolationgroup"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
)

const (
	IsolationGroupKey         = "isolation-group"
	OriginalIsolationGroupKey = "original-isolation-group"
	WorkflowIDKey             = "wf-id"
)

var (
	IsolationLeakCauseError           = metrics.IsolationLeakCause("error")
	IsolationLeakCauseGroupUnknown    = metrics.IsolationLeakCause("group_unknown")
	IsolationLeakCauseGroupDrained    = metrics.IsolationLeakCause("group_drained")
	IsolationLeakCauseNoRecentPollers = metrics.IsolationLeakCause("no_recent_pollers")
	IsolationLeakCauseExpired         = metrics.IsolationLeakCause("expired")
)

// ErrInvalidPartitionConfig is returned when the required partitioning configuration
// is missing due to misconfiguration
var ErrInvalidPartitionConfig = errors.New("invalid partition config")

// DefaultWorkflowPartitionConfig Is the default dataset expected to be passed around in the
// execution records for workflows which is used for partitioning. It contains the IsolationGroup
// where the workflow was started, and is expected to be pinned, and a workflow ID  for a fallback means
// to partition data deterministically.
type defaultWorkflowPartitionConfig struct {
	WorkflowStartIsolationGroup string
	WFID                        string
}

// defaultPartitioner is a business-agnostic implementation of partitioning
// which is used by the Cadence system to allocate workflows in matching by isolation-group
type defaultPartitioner struct {
	log                 log.Logger
	isolationGroupState isolationgroup.State
}

func NewDefaultPartitioner(
	logger log.Logger,
	isolationGroupState isolationgroup.State,
) Partitioner {
	return &defaultPartitioner{
		log:                 logger,
		isolationGroupState: isolationGroupState,
	}
}

func (r *defaultPartitioner) GetIsolationGroupByDomainID(ctx context.Context, scope metrics.Scope, pollerInfo PollerInfo, wfPartitionData PartitionConfig) (string, error) {
	if wfPartitionData == nil {
		return "", ErrInvalidPartitionConfig
	}
	wfPartition := mapPartitionConfigToDefaultPartitionConfig(wfPartitionData)
	if wfPartition.WorkflowStartIsolationGroup == "" || wfPartition.WFID == "" {
		return "", ErrInvalidPartitionConfig
	}

	isolationGroups, err := r.isolationGroupState.IsolationGroupsByDomainID(ctx, pollerInfo.DomainID)
	if err != nil {
		return "", fmt.Errorf("failed to get available isolation groups: %w", err)
	}
	scope = scope.Tagged(metrics.IsolationGroupTag(wfPartition.WorkflowStartIsolationGroup))
	group, ok := isolationGroups[wfPartition.WorkflowStartIsolationGroup]
	if !ok {
		scope.Tagged(IsolationLeakCauseGroupUnknown).IncCounter(metrics.TaskIsolationLeakPerTaskList)
		return "", nil
	}
	if group.State != types.IsolationGroupStateHealthy {
		scope.Tagged(IsolationLeakCauseGroupDrained).IncCounter(metrics.TaskIsolationLeakPerTaskList)
		return "", nil
	}
	if !slices.Contains(pollerInfo.AvailableIsolationGroups, wfPartition.WorkflowStartIsolationGroup) {
		scope.Tagged(IsolationLeakCauseNoRecentPollers).IncCounter(metrics.TaskIsolationLeakPerTaskList)
		return "", nil
	}
	return wfPartition.WorkflowStartIsolationGroup, nil
}

func mapPartitionConfigToDefaultPartitionConfig(config PartitionConfig) defaultWorkflowPartitionConfig {
	isolationGroup, _ := config[IsolationGroupKey]
	wfID, _ := config[WorkflowIDKey]
	return defaultWorkflowPartitionConfig{
		WorkflowStartIsolationGroup: isolationGroup,
		WFID:                        wfID,
	}
}
