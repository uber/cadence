package pinot

import (
	"github.com/stretchr/testify/assert"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"testing"
	"time"
)

var (
	neutrinoClient = NeutrinoClient{
		client: nil,
		logger: nil,
	}
)

func TestNeutinoGetInternalListWorkflowExecutionsResponse(t *testing.T) {
	cols := []struct {
		Name string `json:"name"`
	}{
		{
			Name: "WorkflowID",
		},
		{
			Name: "RunID",
		},
		{
			Name: "WorkflowType",
		},
		{
			Name: "DomainID",
		},
		{
			Name: "StartTime",
		},
		{
			Name: "ExecutionTime",
		},
		{
			Name: "CloseTime",
		},
		{
			Name: "CloseStatus",
		},
		{
			Name: "HistoryLength",
		},
		{
			Name: "Encoding",
		},
		{
			Name: "TaskList",
		},
		{
			Name: "IsCron",
		},
		{
			Name: "NumClusters",
		},
		{
			Name: "UpdateTime",
		},
	}

	hit1 := []interface{}{"wfid1", "rid1", "wftype1", "domainid1", testEarliestTime, testEarliestTime, testLatestTime, 1, 1, "encode1", "tsklst1", true, 1, testEarliestTime}
	hit2 := []interface{}{"wfid2", "rid2", "wftype2", "domainid2", testEarliestTime, testEarliestTime, testLatestTime, 1, 1, "encode2", "tsklst2", false, 1, testEarliestTime}

	response := &queryResponse{
		Columns: cols,
		Data: [][]interface{}{
			hit1,
			hit2,
		},
		Error: nil,
	}

	// Cannot use a table test, because they are not checking the same fields
	result, err := neutrinoClient.getInternalListWorkflowExecutionsResponse(response, nil)

	assert.Equal(t, "wfid1", result.Executions[0].WorkflowID)
	assert.Equal(t, "rid1", result.Executions[0].RunID)
	assert.Equal(t, "wftype1", result.Executions[0].WorkflowType)
	assert.Equal(t, "domainid1", result.Executions[0].DomainID)
	assert.Equal(t, time.UnixMilli(testEarliestTime), result.Executions[0].StartTime)
	assert.Equal(t, time.UnixMilli(testEarliestTime), result.Executions[0].ExecutionTime)
	assert.Equal(t, time.UnixMilli(testLatestTime), result.Executions[0].CloseTime)
	assert.Equal(t, types.WorkflowExecutionCloseStatus(1), *result.Executions[0].Status)
	assert.Equal(t, int64(1), result.Executions[0].HistoryLength)
	assert.Equal(t, "tsklst1", result.Executions[0].TaskList)
	assert.Equal(t, true, result.Executions[0].IsCron)
	assert.Equal(t, int16(1), result.Executions[0].NumClusters)
	assert.Equal(t, time.UnixMilli(testEarliestTime), result.Executions[0].UpdateTime)

	assert.Equal(t, "wfid2", result.Executions[1].WorkflowID)
	assert.Equal(t, "rid2", result.Executions[1].RunID)
	assert.Equal(t, "wftype2", result.Executions[1].WorkflowType)
	assert.Equal(t, "domainid2", result.Executions[1].DomainID)
	assert.Equal(t, time.UnixMilli(testEarliestTime), result.Executions[1].StartTime)
	assert.Equal(t, time.UnixMilli(testEarliestTime), result.Executions[1].ExecutionTime)
	assert.Equal(t, time.UnixMilli(testLatestTime), result.Executions[1].CloseTime)
	assert.Equal(t, types.WorkflowExecutionCloseStatus(1), *result.Executions[1].Status)
	assert.Equal(t, int64(1), result.Executions[1].HistoryLength)
	assert.Equal(t, "tsklst2", result.Executions[1].TaskList)
	assert.Equal(t, false, result.Executions[1].IsCron)
	assert.Equal(t, int16(1), result.Executions[1].NumClusters)
	assert.Equal(t, time.UnixMilli(testEarliestTime), result.Executions[1].UpdateTime)

	assert.Nil(t, err)

	// check if record is not valid
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		return false
	}
	emptyResult, err := neutrinoClient.getInternalListWorkflowExecutionsResponse(response, isRecordValid)
	assert.Equal(t, 0, len(emptyResult.Executions))
	assert.Nil(t, err)

	// check nil input
	nilResult, err := neutrinoClient.getInternalListWorkflowExecutionsResponse(nil, isRecordValid)
	assert.Nil(t, nilResult)
	assert.Nil(t, err)
}
