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
	"sort"
	"sync"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/types"
)

type (
	pollerWeight struct {
		sync.RWMutex
		weights     []int64
		initialized bool
	}

	weightedLoadBalancer struct {
		fallbackLoadBalancer LoadBalancer
		nReadPartitions      dynamicconfig.IntPropertyFnWithTaskListInfoFilters
		domainIDToName       func(string) (string, error)
		weightCache          cache.Cache
		logger               log.Logger
	}
)

func newPollerWeight(n int) *pollerWeight {
	pw := &pollerWeight{weights: make([]int64, n)}
	for i := range pw.weights {
		pw.weights[i] = -1
	}
	return pw
}

func (pw *pollerWeight) pick() int {
	cumulativeWeights := make([]int64, len(pw.weights))
	totalWeight := int64(0)
	pw.RLock()
	defer pw.RUnlock()
	if !pw.initialized {
		return -1
	}
	for i, w := range pw.weights {
		totalWeight += w
		cumulativeWeights[i] = totalWeight
	}
	if totalWeight <= 0 {
		return -1
	}
	r := rand.Int63n(totalWeight)
	index := sort.Search(len(cumulativeWeights), func(i int) bool {
		return cumulativeWeights[i] > r
	})
	return index
}

func (pw *pollerWeight) updateWeightAndPartition(n, p int, weight int64) {
	pw.Lock()
	defer pw.Unlock()
	if n > len(pw.weights) {
		newWeights := make([]int64, n)
		copy(newWeights, pw.weights)
		for i := len(pw.weights); i < n; i++ {
			newWeights[i] = -1
		}
		pw.weights = newWeights
		pw.initialized = false
	}
	pw.weights[p] = weight
	for _, w := range pw.weights {
		if w == -1 {
			return
		}
	}
	pw.initialized = true
}

func NewWeightedLoadBalancer(
	lb LoadBalancer,
	domainIDToName func(string) (string, error),
	dc *dynamicconfig.Collection,
	logger log.Logger,
) LoadBalancer {
	return &weightedLoadBalancer{
		fallbackLoadBalancer: lb,
		domainIDToName:       domainIDToName,
		nReadPartitions:      dc.GetIntPropertyFilteredByTaskListInfo(dynamicconfig.MatchingNumTasklistReadPartitions),
		weightCache: cache.New(&cache.Options{
			TTL:             0,
			InitialCapacity: 100,
			Pin:             false,
			MaxCount:        3000,
			ActivelyEvict:   false,
		}),
		logger: logger,
	}
}

func (lb *weightedLoadBalancer) PickWritePartition(
	domainID string,
	taskList types.TaskList,
	taskListType int,
	forwardedFrom string,
) int {
	return lb.fallbackLoadBalancer.PickWritePartition(domainID, taskList, taskListType, forwardedFrom)
}

func (lb *weightedLoadBalancer) PickReadPartition(
	domainID string,
	taskList types.TaskList,
	taskListType int,
	forwardedFrom string,
) int {
	taskListKey := key{
		domainID:     domainID,
		taskListName: taskList.GetName(),
		taskListType: taskListType,
	}
	wI := lb.weightCache.Get(taskListKey)
	if wI == nil {
		return lb.fallbackLoadBalancer.PickReadPartition(domainID, taskList, taskListType, forwardedFrom)
	}
	w, ok := wI.(*pollerWeight)
	if !ok || !w.initialized {
		return lb.fallbackLoadBalancer.PickReadPartition(domainID, taskList, taskListType, forwardedFrom)
	}
	p := w.pick()
	lb.logger.Info("pick partition", tag.Dynamic("result", p), tag.Dynamic("weights", w.weights))
	if p < 0 {
		return lb.fallbackLoadBalancer.PickReadPartition(domainID, taskList, taskListType, forwardedFrom)
	}
	return p
}

func (lb *weightedLoadBalancer) UpdateWeight(
	domainID string,
	taskList types.TaskList,
	taskListType int,
	partition int,
	weight int64,
) {
	if taskList.GetKind() == types.TaskListKindSticky {
		return
	}
	domainName, err := lb.domainIDToName(domainID)
	if err != nil {
		return
	}
	taskListKey := key{
		domainID:     domainID,
		taskListName: taskList.GetName(),
		taskListType: taskListType,
	}
	n := lb.nReadPartitions(domainName, taskList.GetName(), taskListType)
	if n <= 1 {
		lb.weightCache.Delete(taskListKey)
		return
	}
	wI := lb.weightCache.Get(taskListKey)
	if wI == nil {
		w := newPollerWeight(n)
		wI, err = lb.weightCache.PutIfNotExist(taskListKey, w)
		if err != nil {
			return
		}
	}
	w, ok := wI.(*pollerWeight)
	if !ok {
		return
	}
	w.updateWeightAndPartition(n, partition, weight)
}
