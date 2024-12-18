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

package matching

import (
	"math/rand"
	"slices"

	"golang.org/x/exp/maps"

	"github.com/uber/cadence/common/partition"
	"github.com/uber/cadence/common/types"
)

type isolationLoadBalancer struct {
	provider           PartitionConfigProvider
	fallback           LoadBalancer
	allIsolationGroups func() []string
}

func NewIsolationLoadBalancer(fallback LoadBalancer, provider PartitionConfigProvider, allIsolationGroups func() []string) LoadBalancer {
	return &isolationLoadBalancer{
		provider:           provider,
		fallback:           fallback,
		allIsolationGroups: allIsolationGroups,
	}
}

func (i *isolationLoadBalancer) PickWritePartition(taskListType int, req WriteRequest) string {
	taskList := *req.GetTaskList()
	nPartitions := i.provider.GetNumberOfWritePartitions(req.GetDomainUUID(), taskList, taskListType)
	taskListName := req.GetTaskList().Name

	if nPartitions <= 1 {
		return taskListName
	}

	taskGroup, ok := req.GetPartitionConfig()[partition.IsolationGroupKey]
	if !ok {
		return i.fallback.PickWritePartition(taskListType, req)
	}

	partitions, ok := i.getPartitionsForGroup(taskGroup, nPartitions)
	if !ok {
		return i.fallback.PickWritePartition(taskListType, req)
	}

	p := i.pickBetween(partitions)

	return getPartitionTaskListName(taskList.GetName(), p)
}

func (i *isolationLoadBalancer) PickReadPartition(taskListType int, req ReadRequest, isolationGroup string) string {
	taskList := *req.GetTaskList()
	nRead := i.provider.GetNumberOfReadPartitions(req.GetDomainUUID(), taskList, taskListType)
	taskListName := taskList.Name

	if nRead <= 1 {
		return taskListName
	}

	partitions, ok := i.getPartitionsForGroup(isolationGroup, nRead)
	if !ok {
		return i.fallback.PickReadPartition(taskListType, req, isolationGroup)
	}

	// Scaling down, we need to consider both sets of partitions
	if numWrite := i.provider.GetNumberOfWritePartitions(req.GetDomainUUID(), taskList, taskListType); numWrite != nRead {
		writePartitions, ok := i.getPartitionsForGroup(isolationGroup, numWrite)
		if ok {
			for p := range writePartitions {
				partitions[p] = struct{}{}
			}
		}
	}

	p := i.pickBetween(partitions)

	return getPartitionTaskListName(taskList.GetName(), p)
}

func (i *isolationLoadBalancer) UpdateWeight(taskListType int, req ReadRequest, partition string, info *types.LoadBalancerHints) {
}

func (i *isolationLoadBalancer) getPartitionsForGroup(taskGroup string, partitionCount int) (map[int]any, bool) {
	if taskGroup == "" {
		return nil, false
	}
	isolationGroups := slices.Clone(i.allIsolationGroups())
	slices.Sort(isolationGroups)
	index := slices.Index(isolationGroups, taskGroup)
	if index == -1 {
		return nil, false
	}
	partitions := make(map[int]any, 1)
	// 3 groups [a, b, c] and 4 partitions gives us a mapping like this:
	// 0, 3: a
	// 1: b
	// 2: c
	// 4 groups [a, b, c, d] and 10 partitions gives us a mapping like this:
	// 0, 4, 8: a
	// 1, 5, 9: b
	// 2, 6: c
	// 3, 7: d
	if len(isolationGroups) <= partitionCount {
		for j := index; j < partitionCount; j += len(isolationGroups) {
			partitions[j] = struct{}{}
		}
		// 4 groups [a,b,c,d] and 3 partitions gives us a mapping like this:
		// 0: a, d
		// 1: b
		// 2: c
	} else {
		partitions[index%partitionCount] = struct{}{}
	}
	if len(partitions) == 0 {
		return nil, false
	}
	return partitions, true
}

func (i *isolationLoadBalancer) pickBetween(partitions map[int]any) int {
	// Could alternatively use backlog weights to make a smarter choice
	total := len(partitions)
	picked := rand.Intn(total)
	return maps.Keys(partitions)[picked]
}
