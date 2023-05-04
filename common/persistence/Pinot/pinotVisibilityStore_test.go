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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	pnt "github.com/uber/cadence/common/pinot"

	"github.com/stretchr/testify/assert"

	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
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
	testTableName    = "test-table-name"

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
		Query:      "WorkflowID = 'wfid'",
	}

	result := getCountWorkflowExecutionsQuery(testTableName, request, nil)
	expectResult := fmt.Sprintf(`SELECT COUNT(*)
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND WorkflowID = 'wfid'
`, testTableName)

	assert.Equal(t, result, expectResult)

	nilResult := getCountWorkflowExecutionsQuery(testTableName, nil, nil)
	assert.Equal(t, nilResult, "")
}

func TestGetListWorkflowExecutionQuery(t *testing.T) {
	testValidMap := make(map[string]interface{})
	testValidMap["CustomizedKeyword"] = types.IndexedValueTypeKeyword
	testValidMap["CustomStringField"] = types.IndexedValueTypeString
	testValidMap["IndexedValueTypeInt"] = types.IndexedValueTypeInt
	testValidMap["IndexedValueTypeDouble"] = types.IndexedValueTypeDouble
	testValidMap["IndexedValueTypeBool"] = types.IndexedValueTypeBool
	testValidMap["IndexedValueTypeDatetime"] = types.IndexedValueTypeDatetime

	token := pnt.PinotVisibilityPageToken{
		From: 11,
	}

	serializedToken, err := json.Marshal(token)
	if err != nil {
		panic(fmt.Sprintf("Serialized error in PinotVisibilityStoreTest!!!, %s", err))
	}

	tests := map[string]struct {
		input          *p.ListWorkflowExecutionsByQueryRequest
		expectedOutput string
	}{
		"complete request with keyword query only": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "CustomizedKeyword = keywordCustomized",
			},
			expectedOutput: fmt.Sprintf(
				`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND CustomizedKeyword = 'keywordCustomized'
LIMIT 0, 10
`, testTableName),
		},

		"complete request with keyword query and other customized query": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "CustomizedKeyword = keywordCustomized and CustomStringField = String field is for text",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND CustomizedKeyword = 'keywordCustomized'
AND CustomStringField LIKE '%%String field is for text%%'
LIMIT 0, 10
`, testTableName),
		},

		"complete request with customized query with not registered attribute": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "CustomizedKeyword = keywordCustomized and CustomStringField = String field is for text and unregistered <= 100",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND CustomizedKeyword = 'keywordCustomized'
AND CustomStringField LIKE '%%String field is for text%%'
LIMIT 0, 10
`, testTableName),
		},

		"complete request with customized query with all customized attributes with all cases AND": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "CloseStatus < 0 anD WorkflowType = some-test-workflow and CustomizedKeyword = keywordCustomized AND CustomStringField = String field is for text And unregistered <= 100 aNd Order by DomainId Desc",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND CloseStatus < 0
AND WorkflowType = 'some-test-workflow'
AND CustomizedKeyword = 'keywordCustomized'
AND CustomStringField LIKE '%%String field is for text%%'
Order by DomainId Desc
LIMIT 0, 10
`, testTableName),
		},

		"complete request with customized query with NextPageToken": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: serializedToken,
				Query:         "CloseStatus < 0 and CustomizedKeyword = keywordCustomized AND CustomStringField = String field is for text And unregistered <= 100 aNd Order by DomainId Desc",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND CloseStatus < 0
AND CustomizedKeyword = 'keywordCustomized'
AND CustomStringField LIKE '%%String field is for text%%'
Order by DomainId Desc
LIMIT 11, 10
`, testTableName),
		},

		"complete request with order by query": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "Order by DomainId Desc",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
Order by DomainId Desc
LIMIT 0, 10
`, testTableName),
		},

		"complete request with filter query": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "CloseStatus < 0",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND CloseStatus < 0
LIMIT 0, 10
`, testTableName),
		},

		"complete request with empty query": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
LIMIT 0, 10
`, testTableName),
		},

		"empty request": {
			input: &p.ListWorkflowExecutionsByQueryRequest{},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = ''
LIMIT 0, 0
`, testTableName),
		},

		"nil request": {
			input:          nil,
			expectedOutput: "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				output := getListWorkflowExecutionsByQueryQuery(testTableName, test.input, testValidMap)
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

	closeResult := getListWorkflowExecutionsQuery(testTableName, request, true)
	openResult := getListWorkflowExecutionsQuery(testTableName, request, false)
	nilResult := getListWorkflowExecutionsQuery(testTableName, nil, true)
	expectCloseResult := fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND CloseTime BETWEEN 1547596872371 AND 2547596872371
AND CloseStatus >= 0
Order BY CloseTime DESC
, RunID DESC
LIMIT 0, 10
`, testTableName)
	expectOpenResult := fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND CloseTime BETWEEN 1547596872371 AND 2547596872371
AND CloseStatus < 0
Order BY CloseTime DESC
, RunID DESC
LIMIT 0, 10
`, testTableName)
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

	closeResult := getListWorkflowExecutionsByTypeQuery(testTableName, request, true)
	openResult := getListWorkflowExecutionsByTypeQuery(testTableName, request, false)
	nilResult := getListWorkflowExecutionsByTypeQuery(testTableName, nil, true)
	expectCloseResult := fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND WorkflowType = 'test-wf-type'
AND CloseTime BETWEEN 1547596872371 AND 2547596872371
AND CloseStatus >= 0
Order BY CloseTime DESC
, RunID DESC
`, testTableName)
	expectOpenResult := fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND WorkflowType = 'test-wf-type'
AND CloseTime BETWEEN 1547596872371 AND 2547596872371
AND CloseStatus < 0
Order BY CloseTime DESC
, RunID DESC
`, testTableName)
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

	closeResult := getListWorkflowExecutionsByWorkflowIDQuery(testTableName, request, true)
	openResult := getListWorkflowExecutionsByWorkflowIDQuery(testTableName, request, false)
	nilResult := getListWorkflowExecutionsByWorkflowIDQuery(testTableName, nil, true)
	expectCloseResult := fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND WorkflowID = 'test-wid'
AND CloseTime BETWEEN 1547596872371 AND 2547596872371
AND CloseStatus >= 0
Order BY CloseTime DESC
, RunID DESC
`, testTableName)
	expectOpenResult := fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND WorkflowID = 'test-wid'
AND CloseTime BETWEEN 1547596872371 AND 2547596872371
AND CloseStatus < 0
Order BY CloseTime DESC
, RunID DESC
`, testTableName)
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

	closeResult := getListWorkflowExecutionsByStatusQuery(testTableName, request)
	nilResult := getListWorkflowExecutionsByStatusQuery(testTableName, nil)
	expectCloseResult := fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND CloseStatus = '0'
AND CloseTime BETWEEN 1547596872371 AND 2547596872371
Order BY CloseTime DESC
, RunID DESC
`, testTableName)
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
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND CloseStatus >= 0
AND WorkflowID = 'test-wid'
`, testTableName),
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
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND CloseStatus >= 0
AND WorkflowID = 'test-wid'
AND RunID = 'runid'
`, testTableName),
		},

		"empty request": {
			input: &p.InternalGetClosedWorkflowExecutionRequest{},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = ''
AND CloseStatus >= 0
AND WorkflowID = ''
`, testTableName),
		},

		"nil request": {
			input:          nil,
			expectedOutput: "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				output := getGetClosedWorkflowExecutionQuery(testTableName, test.input)
				assert.Equal(t, test.expectedOutput, output)
			})
		})
	}
}

func TestCheckIfValueSurroundedByQoute(t *testing.T) {
	notSurroundedVal := "testVal"
	singeQuoteVal := "'testVal'"
	doubleQuoteVal := "\"testVal\""

	assert.Equal(t, addSingleQoute(notSurroundedVal), addSingleQoute(singeQuoteVal))
	assert.Equal(t, addSingleQoute(notSurroundedVal), addSingleQoute(doubleQuoteVal))
	assert.Equal(t, addSingleQoute(doubleQuoteVal), addSingleQoute(singeQuoteVal))
}

func TestStringFormatting(t *testing.T) {
	key := "CustomizedStringField"
	val := "When query; select * from users_secret_table;"

	assert.Equal(t, `CustomizedStringField LIKE '%When query; select * from users_secret_table;%'
`, getPartialFormatString(key, val))
}
