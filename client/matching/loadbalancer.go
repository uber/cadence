// Copyright (c) 2019 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package matching

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/types"
)

type (
	// LoadBalancer is the interface for implementers of
	// component that distributes add/poll api calls across
	// available task list partitions when possible
	LoadBalancer interface {
		// PickWritePartition returns the task list partition for adding
		// an activity or decision task. The input is the name of the
		// original task list (with no partition info). When forwardedFrom
		// is non-empty, this call is forwardedFrom from a child partition
		// to a parent partition in which case, no load balancing should be
		// performed
		PickWritePartition(
			domainID string,
			taskList types.TaskList,
			taskListType int,
			forwardedFrom string,
		) string

		// PickReadPartition returns the task list partition to send a poller to.
		// Input is name of the original task list as specified by caller. When
		// forwardedFrom is non-empty, no load balancing should be done.
		PickReadPartition(
			domainID string,
			taskList types.TaskList,
			taskListType int,
			forwardedFrom string,
		) string
	}

	defaultLoadBalancer struct {
		nReadPartitions  dynamicconfig.IntPropertyFnWithTaskListInfoFilters
		nWritePartitions dynamicconfig.IntPropertyFnWithTaskListInfoFilters
		domainIDToName   func(string) (string, error)
	}
)

// NewLoadBalancer returns an instance of matching load balancer that
// can help distribute api calls across task list partitions
func NewLoadBalancer(
	domainIDToName func(string) (string, error),
	dc *dynamicconfig.Collection,
) LoadBalancer {
	return &defaultLoadBalancer{
		domainIDToName:   domainIDToName,
		nReadPartitions:  dc.GetIntPropertyFilteredByTaskListInfo(dynamicconfig.MatchingNumTasklistReadPartitions),
		nWritePartitions: dc.GetIntPropertyFilteredByTaskListInfo(dynamicconfig.MatchingNumTasklistWritePartitions),
	}
}

func (lb *defaultLoadBalancer) PickWritePartition(
	domainID string,
	taskList types.TaskList,
	taskListType int,
	forwardedFrom string,
) string {
	domainName, err := lb.domainIDToName(domainID)
	if err != nil {
		return taskList.GetName()
	}
	nPartitions := lb.nWritePartitions(domainName, taskList.GetName(), taskListType)

	// checks to make sure number of writes never exceeds number of reads
	if nRead := lb.nReadPartitions(domainName, taskList.GetName(), taskListType); nPartitions > nRead {
		nPartitions = nRead
	}
	partition := lb.pickPartition(taskList, forwardedFrom, nPartitions)
	return partition
}

func (lb *defaultLoadBalancer) PickReadPartition(
	domainID string,
	taskList types.TaskList,
	taskListType int,
	forwardedFrom string,
) string {
	domainName, err := lb.domainIDToName(domainID)
	if err != nil {
		return taskList.GetName()
	}
	n := lb.nReadPartitions(domainName, taskList.GetName(), taskListType)
	return lb.pickPartition(taskList, forwardedFrom, n)

}

func (lb *defaultLoadBalancer) pickPartition(
	taskList types.TaskList,
	forwardedFrom string,
	nPartitions int,
) string {

	if forwardedFrom != "" || taskList.GetKind() == types.TaskListKindSticky {
		return taskList.GetName()
	}

	if strings.HasPrefix(taskList.GetName(), common.ReservedTaskListPrefix) {
		// this should never happen when forwardedFrom is empty
		return taskList.GetName()
	}

	p := generateRandomPartitionID(nPartitions)
	if p == 0 {
		return taskList.GetName()
	}

	return fmt.Sprintf("%v%v/%v", common.ReservedTaskListPrefix, taskList.GetName(), p)
}

// generates a number within the range of partitions for partitions > 1, but excluding
// the root partition. Given n=5 partitions, valid values are 1-4 with 0 reserved for the root partition
//
// edge-cases:
// nPartitions == 0 is actually not a valid number of partitions so just return root (0)
// nPartitions == 1 similarly is not a really sane partition count, so just return root (0)
func generateRandomPartitionID(nPartitions int) int {
	if nPartitions <= 0 {
		return 0
	}
	if nPartitions == 1 {
		return 0
	}
	p := rand.Intn(nPartitions - 1)
	return p + 1
}
