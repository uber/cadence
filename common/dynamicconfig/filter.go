// Copyright (c) 2021 Uber Technologies, Inc.
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

package dynamicconfig

// Filter represents a filter on the dynamic config key
type Filter int

func (f Filter) String() string {
	if f <= unknownFilter || f > ClusterName {
		return filters[unknownFilter]
	}
	return filters[f]
}

func parseFilter(filterName string) Filter {
	switch filterName {
	case "domainName":
		return DomainName
	case "domainID":
		return DomainID
	case "taskListName":
		return TaskListName
	case "taskType":
		return TaskType
	case "shardID":
		return ShardID
	case "clusterName":
		return ClusterName
	default:
		return unknownFilter
	}
}

var filters = []string{
	"unknownFilter",
	"domainName",
	"domainID",
	"taskListName",
	"taskType",
	"shardID",
	"clusterName",
}

const (
	unknownFilter Filter = iota
	// DomainName is the domain name
	DomainName
	// DomainID is the domain id
	DomainID
	// TaskListName is the tasklist name
	TaskListName
	// TaskType is the task type (0:Decision, 1:Activity)
	TaskType
	// ShardID is the shard id
	ShardID
	// ClusterName is the cluster name in a multi-region setup
	ClusterName

	// lastFilterTypeForTest must be the last one in this const group for testing purpose
	lastFilterTypeForTest
)

// FilterOption is used to provide filters for dynamic config keys
type FilterOption func(filterMap map[Filter]interface{})

// TaskListFilter filters by task list name
func TaskListFilter(name string) FilterOption {
	return func(filterMap map[Filter]interface{}) {
		filterMap[TaskListName] = name
	}
}

// DomainFilter filters by domain name
func DomainFilter(name string) FilterOption {
	return func(filterMap map[Filter]interface{}) {
		filterMap[DomainName] = name
	}
}

// DomainIDFilter filters by domain id
func DomainIDFilter(domainID string) FilterOption {
	return func(filterMap map[Filter]interface{}) {
		filterMap[DomainID] = domainID
	}
}

// TaskTypeFilter filters by task type
func TaskTypeFilter(taskType int) FilterOption {
	return func(filterMap map[Filter]interface{}) {
		filterMap[TaskType] = taskType
	}
}

// ShardIDFilter filters by shard id
func ShardIDFilter(shardID int) FilterOption {
	return func(filterMap map[Filter]interface{}) {
		filterMap[ShardID] = shardID
	}
}

// ClusterNameFilter filters by cluster name
func ClusterNameFilter(clusterName string) FilterOption {
	return func(filterMap map[Filter]interface{}) {
		filterMap[ClusterName] = clusterName
	}
}
