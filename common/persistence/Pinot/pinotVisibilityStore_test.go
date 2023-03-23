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

package pinotVisibility

import (
	"github.com/startreedata/pinot-client-go/pinot"
	"github.com/stretchr/testify/assert"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"testing"
	"time"
)

var (
	testIndex        = "test-index"
	testDomain       = "test-domain"
	testDomainID     = "bfd5c907-f899-4baf-a7b2-2ab85e623ebd"
	testPageSize     = 10
	testEarliestTime = int64(1547596872371000000)
	testLatestTime   = int64(2547596872371000000)
	testWorkflowType = "test-wf-type"
	testWorkflowID   = "test-wid"
	testCloseStatus  = int32(1)

	visibilityStore = pinotVisibilityStore{
		pinotClient: nil,
		producer:    nil,
		logger:      nil,
		config:      nil,
	}
)

func TestGetCountWorkflowExecutionsQuery(t *testing.T) {
	request := &p.CountWorkflowExecutionsRequest{
		DomainUUID: testDomainID,
		Domain:     testDomain,
		Query:      "WorkflowId = wfid",
	}

	result := getCountWorkflowExecutionsQuery(request)
	expectResult := "SELECT COUNT(*)\nFROM cadence-visibility-pinot\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND WorkflowId = wfid\n"

	assert.Equal(t, result, expectResult)

	nilResult := getCountWorkflowExecutionsQuery(nil)
	assert.Equal(t, nilResult, "")
}

func TestGetListWorkflowExecutionQuery(t *testing.T) {
	tests := map[string]struct {
		input          *p.ListWorkflowExecutionsByQueryRequest
		expectedOutput string
	}{
		"complete request with empty query": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "",
			},
			expectedOutput: "SELECT *\nFROM cadence-visibility-pinot\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nLIMIT 10\n",
		},

		"complete request with order by query": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "Order by DomainId Desc",
			},
			expectedOutput: "SELECT *\nFROM cadence-visibility-pinot\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nOrder by DomainId Desc\nLIMIT 10\n",
		},

		"complete request with filter query": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "Status < 0",
			},
			expectedOutput: "SELECT *\nFROM cadence-visibility-pinot\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND Status < 0\nLIMIT 10\n",
		},

		"empty request": {
			input:          &p.ListWorkflowExecutionsByQueryRequest{},
			expectedOutput: "SELECT *\nFROM cadence-visibility-pinot\nWHERE DomainID = \nLIMIT 0\n",
		},

		"nil request": {
			input:          nil,
			expectedOutput: "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				output := getListWorkflowExecutionsByQueryQuery(test.input)
				assert.Equal(t, test.expectedOutput, output)
			})
		})
	}
}

func TestGetListWorkflowExecutionsQuery(t *testing.T) {
	request := &p.InternalListWorkflowExecutionsRequest{
		DomainUUID:    testDomainID,
		Domain:        testDomain,
		EarliestTime:  time.Unix(0, testEarliestTime),
		LatestTime:    time.Unix(0, testLatestTime),
		PageSize:      testPageSize,
		NextPageToken: nil,
	}

	closeResult := getListWorkflowExecutionsQuery(request, true)
	openResult := getListWorkflowExecutionsQuery(request, false)
	nilResult := getListWorkflowExecutionsQuery(nil, true)
	expectCloseResult := "SELECT *\nFROM cadence-visibility-pinot\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND CloseTime BETWEEN 1547596872371 AND 2547596872371\nAND CloseStatus >= 0\nOrder BY CloseTime DESC\nOrder BY RunID DESC\n"
	expectOpenResult := "SELECT *\nFROM cadence-visibility-pinot\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND CloseTime BETWEEN 1547596872371 AND 2547596872371\nAND CloseStatus < 0\nOrder BY CloseTime DESC\nOrder BY RunID DESC\n"
	expectNilResult := ""

	assert.Equal(t, closeResult, expectCloseResult)
	assert.Equal(t, openResult, expectOpenResult)
	assert.Equal(t, nilResult, expectNilResult)
}

func TestGetListWorkflowExecutionsByTypeQuery(t *testing.T) {
	request := &p.InternalListWorkflowExecutionsByTypeRequest{
		InternalListWorkflowExecutionsRequest: p.InternalListWorkflowExecutionsRequest{
			DomainUUID:    testDomainID,
			Domain:        testDomain,
			EarliestTime:  time.Unix(0, testEarliestTime),
			LatestTime:    time.Unix(0, testLatestTime),
			PageSize:      testPageSize,
			NextPageToken: nil,
		},
		WorkflowTypeName: testWorkflowType,
	}

	closeResult := getListWorkflowExecutionsByTypeQuery(request, true)
	openResult := getListWorkflowExecutionsByTypeQuery(request, false)
	nilResult := getListWorkflowExecutionsByTypeQuery(nil, true)
	expectCloseResult := "SELECT *\nFROM cadence-visibility-pinot\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND WorkflowType = test-wf-type\nAND CloseTime BETWEEN 1547596872371 AND 2547596872371\nAND CloseStatus >= 0\nOrder BY CloseTime DESC\nOrder BY RunID DESC\n"
	expectOpenResult := "SELECT *\nFROM cadence-visibility-pinot\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND WorkflowType = test-wf-type\nAND CloseTime BETWEEN 1547596872371 AND 2547596872371\nAND CloseStatus < 0\nOrder BY CloseTime DESC\nOrder BY RunID DESC\n"
	expectNilResult := ""

	assert.Equal(t, closeResult, expectCloseResult)
	assert.Equal(t, openResult, expectOpenResult)
	assert.Equal(t, nilResult, expectNilResult)
}

func TestGetListWorkflowExecutionsByWorkflowIDQuery(t *testing.T) {
	request := &p.InternalListWorkflowExecutionsByWorkflowIDRequest{
		InternalListWorkflowExecutionsRequest: p.InternalListWorkflowExecutionsRequest{
			DomainUUID:    testDomainID,
			Domain:        testDomain,
			EarliestTime:  time.Unix(0, testEarliestTime),
			LatestTime:    time.Unix(0, testLatestTime),
			PageSize:      testPageSize,
			NextPageToken: nil,
		},
		WorkflowID: testWorkflowID,
	}

	closeResult := getListWorkflowExecutionsByWorkflowIDQuery(request, true)
	openResult := getListWorkflowExecutionsByWorkflowIDQuery(request, false)
	nilResult := getListWorkflowExecutionsByWorkflowIDQuery(nil, true)
	expectCloseResult := "SELECT *\nFROM cadence-visibility-pinot\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND WorkflowID = test-wid\nAND CloseTime BETWEEN 1547596872371 AND 2547596872371\nAND CloseStatus >= 0\nOrder BY CloseTime DESC\nOrder BY RunID DESC\n"
	expectOpenResult := "SELECT *\nFROM cadence-visibility-pinot\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND WorkflowID = test-wid\nAND CloseTime BETWEEN 1547596872371 AND 2547596872371\nAND CloseStatus < 0\nOrder BY CloseTime DESC\nOrder BY RunID DESC\n"
	expectNilResult := ""

	assert.Equal(t, closeResult, expectCloseResult)
	assert.Equal(t, openResult, expectOpenResult)
	assert.Equal(t, nilResult, expectNilResult)
}

func TestGetListWorkflowExecutionsByStatusQuery(t *testing.T) {
	request := &p.InternalListClosedWorkflowExecutionsByStatusRequest{
		InternalListWorkflowExecutionsRequest: p.InternalListWorkflowExecutionsRequest{
			DomainUUID:    testDomainID,
			Domain:        testDomain,
			EarliestTime:  time.Unix(0, testEarliestTime),
			LatestTime:    time.Unix(0, testLatestTime),
			PageSize:      testPageSize,
			NextPageToken: nil,
		},
		Status: types.WorkflowExecutionCloseStatus(0),
	}

	closeResult := getListWorkflowExecutionsByStatusQuery(request)
	nilResult := getListWorkflowExecutionsByStatusQuery(nil)
	expectCloseResult := "SELECT *\nFROM cadence-visibility-pinot\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND CloseStatus = 0\nAND CloseTime BETWEEN 1547596872371 AND 2547596872371\nOrder BY CloseTime DESC\nOrder BY RunID DESC\n"
	expectNilResult := ""

	assert.Equal(t, expectCloseResult, closeResult)
	assert.Equal(t, expectNilResult, nilResult)
}

func TestGetGetClosedWorkflowExecutionQuery(t *testing.T) {
	tests := map[string]struct {
		input          *p.InternalGetClosedWorkflowExecutionRequest
		expectedOutput string
	}{
		"complete request with empty RunId": {
			input: &p.InternalGetClosedWorkflowExecutionRequest{
				DomainUUID: testDomainID,
				Domain:     testDomain,
				Execution: types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      "",
				},
			},
			expectedOutput: "SELECT *\nFROM cadence-visibility-pinot\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND CloseStatus >= 0\nAND WorkflowID = test-wid\n",
		},

		"complete request with runId": {
			input: &p.InternalGetClosedWorkflowExecutionRequest{
				DomainUUID: testDomainID,
				Domain:     testDomain,
				Execution: types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      "runid",
				},
			},
			expectedOutput: "SELECT *\nFROM cadence-visibility-pinot\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND CloseStatus >= 0\nAND WorkflowID = test-wid\nAND RunID = runid\n",
		},

		"empty request": {
			input:          &p.InternalGetClosedWorkflowExecutionRequest{},
			expectedOutput: "SELECT *\nFROM cadence-visibility-pinot\nWHERE DomainID = \nAND CloseStatus >= 0\nAND WorkflowID = \n",
		},

		"nil request": {
			input:          nil,
			expectedOutput: "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				output := getGetClosedWorkflowExecutionQuery(test.input)
				assert.Equal(t, test.expectedOutput, output)
			})
		})
	}
}

func TestBuildMap(t *testing.T) {
	nilResult := buildMap(nil, nil)
	expectNilResult := map[string]interface{}{}
	assert.Equal(t, nilResult, expectNilResult)
}

func TestConvertSearchResultToVisibilityRecord(t *testing.T) {
	badHit := []interface{}{1, 2, 3}
	badColumnName := []string{"testName1", "testName2"}

	badResult := visibilityStore.convertSearchResultToVisibilityRecord(badHit, badColumnName)
	assert.Nil(t, badResult)

	columnName := []string{"WorkflowID", "RunID", "WorkflowType", "DomainID", "StartTime", "ExecutionTime", "CloseTime", "CloseStatus", "HistoryLength", "Encoding", "TaskList", "IsCron", "NumClusters", "UpdateTime"}
	hit := []interface{}{"wfid", "rid", "wftype", "domainid", 10, 10, 10, 1, 1, "encode", "tsklst", true, 1, 10}

	result := *visibilityStore.convertSearchResultToVisibilityRecord(hit, columnName)

	assert.Equal(t, "wfid", result.WorkflowID)
	assert.Equal(t, "rid", result.RunID)
	assert.Equal(t, "wftype", result.WorkflowType)
	assert.Equal(t, "domainid", result.DomainID)
	assert.Equal(t, time.Date(1969, time.December, 31, 16, 0, 0, 10000000, time.Local), result.StartTime)
	assert.Equal(t, time.Date(1969, time.December, 31, 16, 0, 0, 10000000, time.Local), result.ExecutionTime)
	assert.Equal(t, time.Date(1969, time.December, 31, 16, 0, 0, 10000000, time.Local), result.CloseTime)
	assert.Equal(t, types.WorkflowExecutionCloseStatus(1), *result.Status)
	assert.Equal(t, int64(1), result.HistoryLength)
	assert.Equal(t, "tsklst", result.TaskList)
	assert.Equal(t, true, result.IsCron)
	assert.Equal(t, int16(1), result.NumClusters)
	assert.Equal(t, time.Date(1969, time.December, 31, 16, 0, 0, 10000000, time.Local), result.UpdateTime)
}

func TestGetInternalListWorkflowExecutionsResponse(t *testing.T) {
	columnName := []string{"WorkflowID", "RunID", "WorkflowType", "DomainID", "StartTime", "ExecutionTime", "CloseTime", "CloseStatus", "HistoryLength", "Encoding", "TaskList", "IsCron", "NumClusters", "UpdateTime"}
	hit1 := []interface{}{"wfid1", "rid1", "wftype1", "domainid1", 10, 10, 10, 1, 1, "encode1", "tsklst1", true, 1, 10}
	hit2 := []interface{}{"wfid2", "rid2", "wftype2", "domainid2", 10, 10, 10, 1, 1, "encode2", "tsklst2", false, 1, 10}

	brokerResponse := &pinot.BrokerResponse{
		AggregationResults: nil,
		SelectionResults:   nil,
		ResultTable: &pinot.ResultTable{
			DataSchema: pinot.RespSchema{
				nil,
				columnName,
			},
			Rows: [][]interface{}{
				hit1,
				hit2,
			},
		},
		Exceptions:                  nil,
		TraceInfo:                   nil,
		NumServersQueried:           1,
		NumServersResponded:         1,
		NumSegmentsQueried:          1,
		NumSegmentsProcessed:        1,
		NumSegmentsMatched:          1,
		NumConsumingSegmentsQueried: 1,
		NumDocsScanned:              1,
		NumEntriesScannedInFilter:   1,
		NumEntriesScannedPostFilter: 1,
		NumGroupsLimitReached:       false,
		TotalDocs:                   1,
		TimeUsedMs:                  1,
		MinConsumingFreshnessTimeMs: 1,
	}

	// Cannot use a table test, because they are not checking the same fields
	result, err := visibilityStore.getInternalListWorkflowExecutionsResponse(brokerResponse, nil)

	assert.Equal(t, "wfid1", result.Executions[0].WorkflowID)
	assert.Equal(t, "rid1", result.Executions[0].RunID)
	assert.Equal(t, "wftype1", result.Executions[0].WorkflowType)
	assert.Equal(t, "domainid1", result.Executions[0].DomainID)
	assert.Equal(t, time.Date(1969, time.December, 31, 16, 0, 0, 10000000, time.Local), result.Executions[0].StartTime)
	assert.Equal(t, time.Date(1969, time.December, 31, 16, 0, 0, 10000000, time.Local), result.Executions[0].ExecutionTime)
	assert.Equal(t, time.Date(1969, time.December, 31, 16, 0, 0, 10000000, time.Local), result.Executions[0].CloseTime)
	assert.Equal(t, types.WorkflowExecutionCloseStatus(1), *result.Executions[0].Status)
	assert.Equal(t, int64(1), result.Executions[0].HistoryLength)
	assert.Equal(t, "tsklst1", result.Executions[0].TaskList)
	assert.Equal(t, true, result.Executions[0].IsCron)
	assert.Equal(t, int16(1), result.Executions[0].NumClusters)
	assert.Equal(t, time.Date(1969, time.December, 31, 16, 0, 0, 10000000, time.Local), result.Executions[0].UpdateTime)

	assert.Equal(t, "wfid2", result.Executions[1].WorkflowID)
	assert.Equal(t, "rid2", result.Executions[1].RunID)
	assert.Equal(t, "wftype2", result.Executions[1].WorkflowType)
	assert.Equal(t, "domainid2", result.Executions[1].DomainID)
	assert.Equal(t, time.Date(1969, time.December, 31, 16, 0, 0, 10000000, time.Local), result.Executions[1].StartTime)
	assert.Equal(t, time.Date(1969, time.December, 31, 16, 0, 0, 10000000, time.Local), result.Executions[1].ExecutionTime)
	assert.Equal(t, time.Date(1969, time.December, 31, 16, 0, 0, 10000000, time.Local), result.Executions[1].CloseTime)
	assert.Equal(t, types.WorkflowExecutionCloseStatus(1), *result.Executions[1].Status)
	assert.Equal(t, int64(1), result.Executions[1].HistoryLength)
	assert.Equal(t, "tsklst2", result.Executions[1].TaskList)
	assert.Equal(t, false, result.Executions[1].IsCron)
	assert.Equal(t, int16(1), result.Executions[1].NumClusters)
	assert.Equal(t, time.Date(1969, time.December, 31, 16, 0, 0, 10000000, time.Local), result.Executions[1].UpdateTime)

	assert.Nil(t, err)

	// check if record is not valid
	isRecordValid := func(rec *p.InternalVisibilityWorkflowExecutionInfo) bool {
		return false
	}
	emptyResult, err := visibilityStore.getInternalListWorkflowExecutionsResponse(brokerResponse, isRecordValid)
	assert.Equal(t, 0, len(emptyResult.Executions))
	assert.Nil(t, err)

	// check nil input
	nilResult, err := visibilityStore.getInternalListWorkflowExecutionsResponse(nil, isRecordValid)
	assert.Nil(t, nilResult)
	assert.Nil(t, err)
}

func TestGetInternalGetClosedWorkflowExecutionResponse(t *testing.T) {
	columnName := []string{"WorkflowID", "RunID", "WorkflowType", "DomainID", "StartTime", "ExecutionTime", "CloseTime", "CloseStatus", "HistoryLength", "Encoding", "TaskList", "IsCron", "NumClusters", "UpdateTime"}
	hit1 := []interface{}{"wfid1", "rid1", "wftype1", "domainid1", 10, 10, 10, 1, 1, "encode1", "tsklst1", true, 1, 10}

	brokerResponse := &pinot.BrokerResponse{
		AggregationResults: nil,
		SelectionResults:   nil,
		ResultTable: &pinot.ResultTable{
			DataSchema: pinot.RespSchema{
				nil,
				columnName,
			},
			Rows: [][]interface{}{
				hit1,
			},
		},
		Exceptions:                  nil,
		TraceInfo:                   nil,
		NumServersQueried:           1,
		NumServersResponded:         1,
		NumSegmentsQueried:          1,
		NumSegmentsProcessed:        1,
		NumSegmentsMatched:          1,
		NumConsumingSegmentsQueried: 1,
		NumDocsScanned:              1,
		NumEntriesScannedInFilter:   1,
		NumEntriesScannedPostFilter: 1,
		NumGroupsLimitReached:       false,
		TotalDocs:                   1,
		TimeUsedMs:                  1,
		MinConsumingFreshnessTimeMs: 1,
	}

	result, err := visibilityStore.getInternalGetClosedWorkflowExecutionResponse(brokerResponse)

	assert.Equal(t, "wfid1", result.Execution.WorkflowID)
	assert.Equal(t, "rid1", result.Execution.RunID)
	assert.Equal(t, "wftype1", result.Execution.WorkflowType)
	assert.Equal(t, "domainid1", result.Execution.DomainID)
	assert.Equal(t, time.Date(1969, time.December, 31, 16, 0, 0, 10000000, time.Local), result.Execution.StartTime)
	assert.Equal(t, time.Date(1969, time.December, 31, 16, 0, 0, 10000000, time.Local), result.Execution.ExecutionTime)
	assert.Equal(t, time.Date(1969, time.December, 31, 16, 0, 0, 10000000, time.Local), result.Execution.CloseTime)
	assert.Equal(t, types.WorkflowExecutionCloseStatus(1), *result.Execution.Status)
	assert.Equal(t, int64(1), result.Execution.HistoryLength)
	assert.Equal(t, "tsklst1", result.Execution.TaskList)
	assert.Equal(t, true, result.Execution.IsCron)
	assert.Equal(t, int16(1), result.Execution.NumClusters)
	assert.Equal(t, time.Date(1969, time.December, 31, 16, 0, 0, 10000000, time.Local), result.Execution.UpdateTime)

	assert.Nil(t, err)
}
