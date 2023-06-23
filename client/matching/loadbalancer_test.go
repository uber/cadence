package matching

import (
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/types"
	"testing"
)

func TestNewLoadBalancer(t *testing.T) {
	type args struct {
		domainIDToName func(string) (string, error)
		dc             *dynamicconfig.Collection
	}
	tests := []struct {
		name string
		args args
		want LoadBalancer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewLoadBalancer(tt.args.domainIDToName, tt.args.dc), "NewLoadBalancer(%v, %v)", tt.args.domainIDToName, tt.args.dc)
		})
	}
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
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := &defaultLoadBalancer{
				nReadPartitions:  tt.fields.nReadPartitions,
				nWritePartitions: tt.fields.nWritePartitions,
				domainIDToName:   tt.fields.domainIDToName,
			}
			assert.Equalf(t, tt.want, lb.PickReadPartition(tt.args.domainID, tt.args.taskList, tt.args.taskListType, tt.args.forwardedFrom), "PickReadPartition(%v, %v, %v, %v)", tt.args.domainID, tt.args.taskList, tt.args.taskListType, tt.args.forwardedFrom)
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
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "Case: Writes and Reads are same amount",
			fields: fields{
				nReadPartitions:  func(domain string, taskList string, taskType int) int { return 2 },
				nWritePartitions: func(domain string, taskList string, taskType int) int { return 2 },
				domainIDToName:   func(string) (string, error) { return "domainName", nil },
			},
			args: args{
				domainID:      "exampleDomainID",
				taskList:      types.TaskList{Name: "exampleTaskList1"},
				taskListType:  1,
				forwardedFrom: "exampleForwardedFrom",
			},
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
				taskList:      types.TaskList{Name: "exampleTaskList1"},
				taskListType:  1,
				forwardedFrom: "exampleForwardedFrom",
			},
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
				taskList:      types.TaskList{Name: "exampleTaskList3"},
				taskListType:  1,
				forwardedFrom: "exampleForwardedFrom",
			},
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
				lastDigit := int(got[len(got)-1] - '0')
				assert.True(t, lastDigit <= tt.fields.nReadPartitions(tt.args.domainID, tt.args.taskList.Name, tt.args.taskListType))
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := &defaultLoadBalancer{
				nReadPartitions:  tt.fields.nReadPartitions,
				nWritePartitions: tt.fields.nWritePartitions,
				domainIDToName:   tt.fields.domainIDToName,
			}
			assert.Equalf(t, tt.want, lb.pickPartition(tt.args.taskList, tt.args.forwardedFrom, tt.args.nPartitions), "pickPartition(%v, %v, %v)", tt.args.taskList, tt.args.forwardedFrom, tt.args.nPartitions)
		})
	}
}
