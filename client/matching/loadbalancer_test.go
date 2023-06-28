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
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/types"
	"testing"
)

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
				nReadPartitions:  tt.fields.nReadPartitions,
				nWritePartitions: tt.fields.nWritePartitions,
				domainIDToName:   tt.fields.domainIDToName,
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
				nReadPartitions:  tt.fields.nReadPartitions,
				nWritePartitions: tt.fields.nWritePartitions,
				domainIDToName:   tt.fields.domainIDToName,
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
			got := lb.pickPartition(tt.args.taskList, tt.args.forwardedFrom, tt.args.nPartitions)
			assert.Equal(t, tt.want, got)
		})
	}
}
