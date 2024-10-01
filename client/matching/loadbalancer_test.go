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
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package matching

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/types"
)

func TestConstructor(t *testing.T) {
	assert.NotPanics(t, func() {
		lb := NewLoadBalancer(func(string) (string, error) { return "", nil }, dynamicconfig.NewNopCollection())
		assert.NotNil(t, lb)
	})
}

func Test_defaultLoadBalancer_PickReadPartition(t *testing.T) {
	type fields struct {
		nReadPartitions  dynamicconfig.IntPropertyFnWithTaskListInfoFilters
		nWritePartitions dynamicconfig.IntPropertyFnWithTaskListInfoFilters
		domainIDToName   func(string) (string, error)
	}
	type args struct {
		domainID      string
		taskList      types.TaskList
		taskListType  int
		forwardedFrom string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		expectedVal []string
	}{
		{name: "Testing Read",
			fields: fields{
				nReadPartitions: func(domainName string, taskListName string, taskListType int) int {
					return 3
				},
				nWritePartitions: nil,
				domainIDToName: func(domainID string) (string, error) {
					return "domainName", nil
				},
			},
			args: args{
				domainID:      "domainID",
				taskList:      types.TaskList{Name: "taskListName1"},
				taskListType:  1,
				forwardedFrom: "",
			},
			expectedVal: []string{"taskListName1", "/__cadence_sys/taskListName1/1", "/__cadence_sys/taskListName1/2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := &defaultLoadBalancer{
				nReadPartitions:    tt.fields.nReadPartitions,
				nWritePartitions:   tt.fields.nWritePartitions,
				domainIDToName:     tt.fields.domainIDToName,
				maxChildrenPerNode: func(domain string, taskList string, taskType int) int { return 20 },
			}
			for i := 0; i < 100; i++ {
				got := lb.PickReadPartition(tt.args.domainID, tt.args.taskList, tt.args.taskListType, tt.args.forwardedFrom)
				assert.Contains(t, tt.expectedVal, got)
			}

		})
	}
}

func Test_defaultLoadBalancer_PickWritePartition(t *testing.T) {
	type fields struct {
		nReadPartitions  dynamicconfig.IntPropertyFnWithTaskListInfoFilters
		nWritePartitions dynamicconfig.IntPropertyFnWithTaskListInfoFilters
		domainIDToName   func(string) (string, error)
	}
	type args struct {
		domainID      string
		taskList      types.TaskList
		taskListType  int
		forwardedFrom string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		expectedVal []string
	}{
		{
			name: "Case: Writes and Reads are same amount",
			fields: fields{
				nReadPartitions:  func(domain string, taskList string, taskType int) int { return 2 },
				nWritePartitions: func(domain string, taskList string, taskType int) int { return 2 },
				domainIDToName:   func(string) (string, error) { return "domainName", nil },
			},
			args: args{
				domainID:      "exampleDomainID",
				taskList:      types.TaskList{Name: "taskListName1"},
				taskListType:  1,
				forwardedFrom: "",
			},
			expectedVal: []string{"taskListName1", "/__cadence_sys/taskListName1/1"},
		},

		{
			name: "Case: More Writes than Reads ",
			fields: fields{
				nReadPartitions:  func(domain string, taskList string, taskType int) int { return 2 },
				nWritePartitions: func(domain string, taskList string, taskType int) int { return 3 },
				domainIDToName:   func(string) (string, error) { return "domainName", nil },
			},
			args: args{
				domainID:      "exampleDomainID",
				taskList:      types.TaskList{Name: "taskListName1"},
				taskListType:  1,
				forwardedFrom: "",
			},
			expectedVal: []string{"taskListName1", "/__cadence_sys/taskListName1/1"},
		},

		{
			name: "Case: More Reads than Writes ",
			fields: fields{
				nReadPartitions:  func(domain string, taskList string, taskType int) int { return 4 },
				nWritePartitions: func(domain string, taskList string, taskType int) int { return 2 },
				domainIDToName:   func(string) (string, error) { return "domainName", nil },
			},
			args: args{
				domainID:      "exampleDomainID",
				taskList:      types.TaskList{Name: "taskListName3"},
				taskListType:  1,
				forwardedFrom: "",
			},
			expectedVal: []string{"taskListName3", "/__cadence_sys/taskListName3/1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := &defaultLoadBalancer{
				nReadPartitions:    tt.fields.nReadPartitions,
				nWritePartitions:   tt.fields.nWritePartitions,
				domainIDToName:     tt.fields.domainIDToName,
				maxChildrenPerNode: func(domain string, taskList string, taskType int) int { return 20 },
			}
			for i := 0; i < 100; i++ {
				got := lb.PickWritePartition(tt.args.domainID, tt.args.taskList, tt.args.taskListType, tt.args.forwardedFrom)
				assert.Contains(t, tt.expectedVal, got)

			}

		})
	}
}

func Test_defaultLoadBalancer_pickPartition(t *testing.T) {
	type fields struct {
		nReadPartitions  dynamicconfig.IntPropertyFnWithTaskListInfoFilters
		nWritePartitions dynamicconfig.IntPropertyFnWithTaskListInfoFilters
		domainIDToName   func(string) (string, error)
	}
	type args struct {
		taskList      types.TaskList
		forwardedFrom string
		nPartitions   int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "Test: ForwardedFrom not empty",
			fields: fields{},
			args: args{
				taskList: types.TaskList{
					Name: "taskList1",
					Kind: types.TaskListKindSticky.Ptr(),
				},
				forwardedFrom: "forwardedFromVal",
				nPartitions:   10,
			},
			want: "taskList1",
		},
		{
			name:   "Test: TaskList kind is Sticky",
			fields: fields{},
			args: args{
				taskList: types.TaskList{
					Name: "taskList2",
					Kind: types.TaskListKindSticky.Ptr(),
				},
				forwardedFrom: "",
				nPartitions:   10,
			},
			want: "taskList2",
		},
		{
			name:   "Test: TaskList name starts with ReservedTaskListPrefix",
			fields: fields{},
			args: args{
				taskList: types.TaskList{
					Name: common.ReservedTaskListPrefix + "taskList3",
					Kind: types.TaskListKindNormal.Ptr(),
				},
				forwardedFrom: "",
				nPartitions:   10,
			},
			want: common.ReservedTaskListPrefix + "taskList3",
		},
		// This behaviour will be preserved, but this is being removed as an expected value
		// for the partitioning. The function to generate a partition count *should never
		// send traffic to the root partition directly* so therefore this is an invalid value.
		// All traffic should be directed to root partitions and then only if there's a problem
		// finding an immediate match, to therefore forward to the root partition.
		//
		// The reason for this is that it adds needless complexity to the partitioned tasklists -
		// it's hard to follow what's going on when 1/n of traffic is being forwarded directly and the
		// rest is being forwarded, so spotting misconfigurations due to overpartitioning is harder
		// and it introduces additional load on the root partition which risks being a heavy contention
		// point if the number of partitions is too great for the number of pollers (a very
		// frequent problem with zonal isolation).
		{
			name:   "Test: nPartitions <= 0",
			fields: fields{},
			args: args{
				taskList: types.TaskList{
					Name: "taskList4",
					Kind: types.TaskListKindNormal.Ptr(),
				},
				forwardedFrom: "",
				nPartitions:   0,
			},
			want: "taskList4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := &defaultLoadBalancer{
				nReadPartitions:  tt.fields.nReadPartitions,
				nWritePartitions: tt.fields.nWritePartitions,
				domainIDToName:   tt.fields.domainIDToName,
			}
			got := lb.pickPartition(tt.args.taskList, tt.args.forwardedFrom, tt.args.nPartitions, 20)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGeneratingAPartitionCount(t *testing.T) {

	tests := map[string]struct {
		numberOfPartitions       int
		expectedResultLowerBound int
		expectedResultUpperBound int
	}{
		"valid partition count = 5, expected values are between 1 to 4": {
			numberOfPartitions:       5,
			expectedResultLowerBound: 1,
			expectedResultUpperBound: 4,
		},
		"min sane partition count = 2 expected values are between 1 to 2": {
			numberOfPartitions:       2,
			expectedResultLowerBound: 1,
			expectedResultUpperBound: 2,
		},
		"large partition count": {
			numberOfPartitions:       30,
			expectedResultLowerBound: 1,
			expectedResultUpperBound: 29,
		},
		"weird, probably a mistake, partition count 1, just ensure that everything is sent to root": {
			numberOfPartitions:       1,
			expectedResultLowerBound: 0,
			expectedResultUpperBound: 1,
		},
		"weird, probably a mistake, partition count 0, just ensure that everything is sent to root": {
			numberOfPartitions:       0,
			expectedResultLowerBound: 0,
			expectedResultUpperBound: 0,
		},
		"invalid partition count, just ensure that everything is sent to root": {
			numberOfPartitions:       -1,
			expectedResultLowerBound: 0,
			expectedResultUpperBound: 0,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				lb := defaultLoadBalancer{}

				p := lb.generateRandomPartitionID(td.numberOfPartitions, 20)

				assert.LessOrEqual(t, p, td.expectedResultUpperBound)

				assert.GreaterOrEqual(t, p, td.expectedResultLowerBound)
			}
		})
	}
}
