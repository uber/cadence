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
	"strings"
	"sync/atomic"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/types"
)

type (
	key struct {
		domainID     string
		taskListName string
		taskListType int
	}

	roundRobinLoadBalancer struct {
		nReadPartitions  dynamicconfig.IntPropertyFnWithTaskListInfoFilters
		nWritePartitions dynamicconfig.IntPropertyFnWithTaskListInfoFilters
		domainIDToName   func(string) (string, error)
		readCache        cache.Cache
		writeCache       cache.Cache

		pickPartitionFn func(domainName string, taskList types.TaskList, taskListType int, forwardedFrom string, nPartitions int, partitionCache cache.Cache) int
	}
)

func NewRoundRobinLoadBalancer(
	domainIDToName func(string) (string, error),
	dc *dynamicconfig.Collection,
) LoadBalancer {
	return &roundRobinLoadBalancer{
		domainIDToName:   domainIDToName,
		nReadPartitions:  dc.GetIntPropertyFilteredByTaskListInfo(dynamicconfig.MatchingNumTasklistReadPartitions),
		nWritePartitions: dc.GetIntPropertyFilteredByTaskListInfo(dynamicconfig.MatchingNumTasklistWritePartitions),
		readCache: cache.New(&cache.Options{
			TTL:             0,
			InitialCapacity: 100,
			Pin:             false,
			MaxCount:        3000,
			ActivelyEvict:   false,
		}),
		writeCache: cache.New(&cache.Options{
			TTL:             0,
			InitialCapacity: 100,
			Pin:             false,
			MaxCount:        3000,
			ActivelyEvict:   false,
		}),
		pickPartitionFn: pickPartition,
	}
}

func (lb *roundRobinLoadBalancer) PickWritePartition(
	domainID string,
	taskList types.TaskList,
	taskListType int,
	forwardedFrom string,
) int {
	domainName, err := lb.domainIDToName(domainID)
	if err != nil {
		return 0
	}
	nPartitions := lb.nWritePartitions(domainName, taskList.GetName(), taskListType)

	// checks to make sure number of writes never exceeds number of reads
	if nRead := lb.nReadPartitions(domainName, taskList.GetName(), taskListType); nPartitions > nRead {
		nPartitions = nRead
	}
	return lb.pickPartitionFn(domainID, taskList, taskListType, forwardedFrom, nPartitions, lb.writeCache)
}

func (lb *roundRobinLoadBalancer) PickReadPartition(
	domainID string,
	taskList types.TaskList,
	taskListType int,
	forwardedFrom string,
) int {
	domainName, err := lb.domainIDToName(domainID)
	if err != nil {
		return 0
	}
	n := lb.nReadPartitions(domainName, taskList.GetName(), taskListType)
	return lb.pickPartitionFn(domainID, taskList, taskListType, forwardedFrom, n, lb.readCache)
}

func pickPartition(
	domainID string,
	taskList types.TaskList,
	taskListType int,
	forwardedFrom string,
	nPartitions int,
	partitionCache cache.Cache,
) int {
	taskListName := taskList.GetName()
	if forwardedFrom != "" || taskList.GetKind() == types.TaskListKindSticky {
		return 0
	}
	if strings.HasPrefix(taskListName, common.ReservedTaskListPrefix) {
		// this should never happen when forwardedFrom is empty
		return 0
	}
	if nPartitions <= 1 {
		return 0
	}

	taskListKey := key{
		domainID:     domainID,
		taskListName: taskListName,
		taskListType: taskListType,
	}

	valI := partitionCache.Get(taskListKey)
	if valI == nil {
		val := int64(-1)
		var err error
		valI, err = partitionCache.PutIfNotExist(taskListKey, &val)
		if err != nil {
			return 0
		}
	}
	valAddr, ok := valI.(*int64)
	if !ok {
		return 0
	}

	return int(atomic.AddInt64(valAddr, 1) % int64(nPartitions))
}

func (lb *roundRobinLoadBalancer) UpdateWeight(
	domainID string,
	taskList types.TaskList,
	taskListType int,
	partition int,
	weight int64,
) {
}
