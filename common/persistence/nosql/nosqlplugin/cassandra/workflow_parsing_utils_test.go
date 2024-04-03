package cassandra

import (
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/persistence"
	"testing"
)

type mockUUID struct {
	uuid string
}

func (m mockUUID) String() string {
	return m.uuid
}

func Test_parseWorkflowExecutionInfo(t *testing.T) {

	tests := []struct {
		args map[string]interface{}
		want *persistence.InternalWorkflowExecutionInfo
	}{
		{
			args: map[string]interface{}{
				"domain_id":                 mockUUID{"domain_id"},
				"workflow_id":               "workflow_id",
				"run_id":                    mockUUID{"run_id"},
				"parent_workflow_id":        "parent_workflow_id",
				"initiated_id":              int64(1),
				"completion_event_batch_id": int64(2),
				"task_list":                 "task_list",
				"workflow_type_name":        "workflow_type_name",
			},
			want: &persistence.InternalWorkflowExecutionInfo{
				DomainID:               "domain_id",
				WorkflowID:             "workflow_id",
				RunID:                  "run_id",
				ParentWorkflowID:       "parent_workflow_id",
				InitiatedID:            int64(1),
				CompletionEventBatchID: int64(2),
				TaskList:               "task_list",
				WorkflowTypeName:       "workflow_type_name",
			},
		},
	}
	for _, tt := range tests {
		result := parseWorkflowExecutionInfo(tt.args)
		assert.Equal(t, result.DomainID, tt.want.DomainID)
		assert.Equal(t, result.WorkflowID, tt.want.WorkflowID)
		assert.Equal(t, result.RunID, tt.want.RunID)
		assert.Equal(t, result.ParentWorkflowID, tt.want.ParentWorkflowID)
		assert.Equal(t, result.InitiatedID, tt.want.InitiatedID)
		assert.Equal(t, result.CompletionEventBatchID, tt.want.CompletionEventBatchID)
		assert.Equal(t, result.TaskList, tt.want.TaskList)
		assert.Equal(t, result.WorkflowTypeName, tt.want.WorkflowTypeName)
	}
}
