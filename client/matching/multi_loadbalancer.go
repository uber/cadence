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
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/types"
)

type (
	multiLoadBalancer struct {
		random               LoadBalancer
		roundRobin           LoadBalancer
		domainIDToName       func(string) (string, error)
		loadbalancerStrategy dynamicconfig.StringPropertyFnWithTaskListInfoFilters
	}
)

func NewMultiLoadBalancer(
	random LoadBalancer,
	roundRobin LoadBalancer,
	domainIDToName func(string) (string, error),
	dc *dynamicconfig.Collection,
) LoadBalancer {
	return &multiLoadBalancer{
		random:               random,
		roundRobin:           roundRobin,
		domainIDToName:       domainIDToName,
		loadbalancerStrategy: dc.GetStringPropertyFilteredByTaskListInfo(dynamicconfig.TasklistLoadBalancerStrategy),
	}
}

func (lb *multiLoadBalancer) PickWritePartition(
	domainID string,
	taskList types.TaskList,
	taskListType int,
	forwardedFrom string,
) string {
	domainName, err := lb.domainIDToName(domainID)
	if err != nil {
		return lb.random.PickWritePartition(domainID, taskList, taskListType, forwardedFrom)
	}
	if lb.loadbalancerStrategy(domainName, taskList.GetName(), taskListType) == "round-robin" {
		return lb.roundRobin.PickWritePartition(domainID, taskList, taskListType, forwardedFrom)
	}
	return lb.random.PickWritePartition(domainID, taskList, taskListType, forwardedFrom)
}

func (lb *multiLoadBalancer) PickReadPartition(
	domainID string,
	taskList types.TaskList,
	taskListType int,
	forwardedFrom string,
) string {
	domainName, err := lb.domainIDToName(domainID)
	if err != nil {
		return lb.random.PickReadPartition(domainID, taskList, taskListType, forwardedFrom)
	}
	if lb.loadbalancerStrategy(domainName, taskList.GetName(), taskListType) == "round-robin" {
		return lb.roundRobin.PickReadPartition(domainID, taskList, taskListType, forwardedFrom)
	}
	return lb.random.PickReadPartition(domainID, taskList, taskListType, forwardedFrom)
}
