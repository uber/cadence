package matching

import (
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/types"
	"strings"
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
				forwardedFrom: "forwardedFrom",
			},
			expectedVal: []string{"/__cadence_sys/taskListName1/0", "/__cadence_sys/taskListName1/1", "/__cadence_sys/taskListName1/2"},
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
				found := false
				for _, val := range tt.expectedVal {
					if strings.Contains(val, got) {
						found = true
						break
					}
				}
				assert.True(t, found)

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
				forwardedFrom: "exampleForwardedFrom",
			},
			expectedVal: []string{"/__cadence_sys/taskListName1/0", "/__cadence_sys/taskListName1/1"},
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
				forwardedFrom: "exampleForwardedFrom",
			},
			expectedVal: []string{"/__cadence_sys/taskListName1/0", "/__cadence_sys/taskListName1/1"},
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
				forwardedFrom: "exampleForwardedFrom",
			},
			expectedVal: []string{"/__cadence_sys/taskListName3/0", "/__cadence_sys/taskListName3/1", "/__cadence_sys/taskListName3/2", "/__cadence_sys/taskListName3/3"},
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
				found := false
				for _, val := range tt.expectedVal {
					if strings.Contains(val, got) {
						found = true
						break
					}
				}
				assert.True(t, found)
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
