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
	"fmt"
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
	expectResult := fmt.Sprintf("SELECT COUNT(*)\nFROM %s\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND WorkflowId = wfid\n", tableName)

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
			expectedOutput: fmt.Sprintf("SELECT *\nFROM %s\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nLIMIT 10\n", tableName),
		},

		"complete request with order by query": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "Order by DomainId Desc",
			},
			expectedOutput: fmt.Sprintf("SELECT *\nFROM %s\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nOrder by DomainId Desc\nLIMIT 10\n", tableName),
		},

		"complete request with filter query": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "Status < 0",
			},
			expectedOutput: fmt.Sprintf("SELECT *\nFROM %s\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND Status < 0\nLIMIT 10\n", tableName),
		},

		"empty request": {
			input:          &p.ListWorkflowExecutionsByQueryRequest{},
			expectedOutput: fmt.Sprintf("SELECT *\nFROM %s\nWHERE DomainID = \nLIMIT 0\n", tableName),
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
	expectCloseResult := fmt.Sprintf("SELECT *\nFROM %s\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND CloseTime BETWEEN 1547596872371 AND 2547596872371\nAND CloseStatus >= 0\nOrder BY CloseTime DESC\nOrder BY RunID DESC\n", tableName)
	expectOpenResult := fmt.Sprintf("SELECT *\nFROM %s\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND CloseTime BETWEEN 1547596872371 AND 2547596872371\nAND CloseStatus < 0\nOrder BY CloseTime DESC\nOrder BY RunID DESC\n", tableName)
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
	expectCloseResult := fmt.Sprintf("SELECT *\nFROM %s\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND WorkflowType = test-wf-type\nAND CloseTime BETWEEN 1547596872371 AND 2547596872371\nAND CloseStatus >= 0\nOrder BY CloseTime DESC\nOrder BY RunID DESC\n", tableName)
	expectOpenResult := fmt.Sprintf("SELECT *\nFROM %s\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND WorkflowType = test-wf-type\nAND CloseTime BETWEEN 1547596872371 AND 2547596872371\nAND CloseStatus < 0\nOrder BY CloseTime DESC\nOrder BY RunID DESC\n", tableName)
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
	expectCloseResult := fmt.Sprintf("SELECT *\nFROM %s\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND WorkflowID = test-wid\nAND CloseTime BETWEEN 1547596872371 AND 2547596872371\nAND CloseStatus >= 0\nOrder BY CloseTime DESC\nOrder BY RunID DESC\n", tableName)
	expectOpenResult := fmt.Sprintf("SELECT *\nFROM %s\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND WorkflowID = test-wid\nAND CloseTime BETWEEN 1547596872371 AND 2547596872371\nAND CloseStatus < 0\nOrder BY CloseTime DESC\nOrder BY RunID DESC\n", tableName)
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
	expectCloseResult := fmt.Sprintf("SELECT *\nFROM %s\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND CloseStatus = 0\nAND CloseTime BETWEEN 1547596872371 AND 2547596872371\nOrder BY CloseTime DESC\nOrder BY RunID DESC\n", tableName)
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
			expectedOutput: fmt.Sprintf("SELECT *\nFROM %s\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND CloseStatus >= 0\nAND WorkflowID = test-wid\n", tableName),
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
			expectedOutput: fmt.Sprintf("SELECT *\nFROM %s\nWHERE DomainID = bfd5c907-f899-4baf-a7b2-2ab85e623ebd\nAND CloseStatus >= 0\nAND WorkflowID = test-wid\nAND RunID = runid\n", tableName),
		},

		"empty request": {
			input:          &p.InternalGetClosedWorkflowExecutionRequest{},
			expectedOutput: fmt.Sprintf("SELECT *\nFROM %s\nWHERE DomainID = \nAND CloseStatus >= 0\nAND WorkflowID = \n", tableName),
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
