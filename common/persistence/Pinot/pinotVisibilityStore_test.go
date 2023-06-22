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

	"github.com/uber/cadence/common/log"
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
		logger:      log.NewNoop(),
		config:      nil,
	}
)

func TestGetCountWorkflowExecutionsQuery(t *testing.T) {
	request := &p.CountWorkflowExecutionsRequest{
		DomainUUID: testDomainID,
		Domain:     testDomain,
		Query:      "WorkflowID = 'wfid'",
	}

	result := visibilityStore.getCountWorkflowExecutionsQuery(testTableName, request, nil)
	expectResult := fmt.Sprintf(`SELECT COUNT(*)
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND WorkflowID = 'wfid'
`, testTableName)

	assert.Equal(t, result, expectResult)

	nilResult := visibilityStore.getCountWorkflowExecutionsQuery(testTableName, nil, nil)
	assert.Equal(t, nilResult, "")
}

func TestGetListWorkflowExecutionQuery(t *testing.T) {
	testValidMap := make(map[string]interface{})
	testValidMap["CustomizedKeyword"] = types.IndexedValueTypeKeyword
	testValidMap["CustomStringField"] = types.IndexedValueTypeString
	testValidMap["CustomIntField"] = types.IndexedValueTypeInt
	testValidMap["CustomKeywordField"] = types.IndexedValueTypeDouble
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
				Query:         "`Attr.CustomizedKeyword` = 'keywordCustomized'",
			},
			expectedOutput: fmt.Sprintf(
				`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND CustomizedKeyword = 'keywordCustomized'
Order BY CloseTime DESC
LIMIT 0, 10
`, testTableName),
		},

		"complete request from search attribute worker": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "CustomIntField=2 and CustomKeywordField='Update2' order by `Attr.CustomDatetimeField` DESC",
			},
			expectedOutput: fmt.Sprintf(
				`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND CustomIntField = 2
AND CustomKeywordField = 'Update2'
order by CustomDatetimeField DESC
LIMIT 0, 10
`, testTableName),
		},

		"complete request with keyword query and other customized query": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "CustomizedKeyword = 'keywordCustomized' and CustomStringField = 'String field is for text' Order by CloseTime DESC, RunID DESC",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND CustomizedKeyword = 'keywordCustomized'
AND CustomStringField LIKE '%%String field is for text%%'
Order by CloseTime DESC, RunID DESC
LIMIT 0, 10
`, testTableName),
		},

		"complete request with or query & customized attributes": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "CustomStringField = 'String' or CustomStringField = 'field' Order by CloseTime DESC, RunID DESC",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND (CustomStringField LIKE '%%String%%' or CustomStringField LIKE '%%field%%')
Order by CloseTime DESC, RunID DESC
LIMIT 0, 10
`, testTableName),
		},

		"complete request with or query & all attributes": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "(DocID = 'String' or CustomStringField = 'field' or CustomIntField = 10) and RunID='test-runid' Order by CloseTime DESC, RunID DESC",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND (DocID = 'String' or CustomStringField LIKE '%%field%%' or CustomIntField = 10)
AND RunID='test-runid'
Order by CloseTime DESC, RunID DESC
LIMIT 0, 10
`, testTableName),
		},

		"complete request with customized query with not registered attribute": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "CustomizedKeyword = 'keywordCustomized' and CustomStringField = 'String field is for text' and unregistered <= 100",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND CustomizedKeyword = 'keywordCustomized'
AND CustomStringField LIKE '%%String field is for text%%'
Order BY CloseTime DESC
LIMIT 0, 10
`, testTableName),
		},

		"complete request with customized query with all customized attributes with all cases AND & a invalid string input": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "CloseTime = missing anD WorkflowType = some-test-workflow and CustomizedKeyword = 'keywordCustomized' AND CustomStringField = 'String field is for text' And unregistered <= 100 aNd Order by DomainId Desc",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND CloseTime = -1
AND WorkflowType = some-test-workflow
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
				Query:         "CloseStatus < 0 and CustomizedKeyword = 'keywordCustomized' AND CustomIntField<=10 and CustomStringField = 'String field is for text' And unregistered <= 100 aNd Order by DomainId Desc",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND CloseStatus < 0
AND CustomizedKeyword = 'keywordCustomized'
AND CustomIntField <= 10
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
Order BY CloseTime DESC
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
				output := visibilityStore.getListWorkflowExecutionsByQueryQuery(testTableName, test.input, testValidMap)
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
AND CloseTime BETWEEN 1547596871371 AND 2547596873371
AND CloseStatus >= 0
Order BY CloseTime DESC
LIMIT 0, 10
`, testTableName)
	expectOpenResult := fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND StartTime BETWEEN 1547596871371 AND 2547596873371
AND CloseStatus < 0
AND CloseTime = -1
Order BY RunID DESC
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
AND CloseTime BETWEEN 1547596871371 AND 2547596873371
AND CloseStatus >= 0
Order BY CloseTime DESC
`, testTableName)
	expectOpenResult := fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND WorkflowType = 'test-wf-type'
AND StartTime BETWEEN 1547596871371 AND 2547596873371
AND CloseStatus < 0
AND CloseTime = -1
Order BY RunID DESC
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
AND CloseTime BETWEEN 1547596871371 AND 2547596873371
AND CloseStatus >= 0
Order BY CloseTime DESC
`, testTableName)
	expectOpenResult := fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND WorkflowID = 'test-wid'
AND StartTime BETWEEN 1547596871371 AND 2547596873371
AND CloseStatus < 0
AND CloseTime = -1
Order BY CloseTime DESC
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

func TestStringFormatting(t *testing.T) {
	key := "CustomizedStringField"
	val := "When query; select * from users_secret_table;"

	assert.Equal(t, `CustomizedStringField LIKE '%When query; select * from users_secret_table;%'`, getPartialFormatString(key, val))
}

func TestParseLastElement(t *testing.T) {
	tests := map[string]struct {
		input           string
		expectedElement string
		expectedOrderBy string
	}{
		"Case1: only contains order by": {
			input:           "Order by TestInt DESC",
			expectedElement: "",
			expectedOrderBy: "Order by TestInt DESC",
		},
		"Case2: only contains order by": {
			input:           "TestString = 'cannot be used in order by'",
			expectedElement: "TestString = 'cannot be used in order by'",
			expectedOrderBy: "",
		},
		"Case3: not contains any order by": {
			input:           "TestInt = 1",
			expectedElement: "TestInt = 1",
			expectedOrderBy: "",
		},
		"Case4-1: with order by in string & real order by": {
			input:           "TestString = 'cannot be used in order by' Order by TestInt DESC",
			expectedElement: "TestString = 'cannot be used in order by'",
			expectedOrderBy: "Order by TestInt DESC",
		},
		"Case4-2: with non-string attribute & real order by": {
			input:           "TestDouble = 1.0 Order by TestInt DESC",
			expectedElement: "TestDouble = 1.0",
			expectedOrderBy: "Order by TestInt DESC",
		},
		"Case5: with random case order by": {
			input:           "TestString = 'cannot be used in OrDer by' ORdeR by TestInt DESC",
			expectedElement: "TestString = 'cannot be used in OrDer by'",
			expectedOrderBy: "ORdeR by TestInt DESC",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				element, orderBy := parseLastElement(test.input)
				assert.Equal(t, test.expectedElement, element)
				assert.Equal(t, test.expectedOrderBy, orderBy)
			})
		})
	}
}

func TestSplitElement(t *testing.T) {
	tests := map[string]struct {
		input       string
		expectedKey string
		expectedVal string
		expectedOp  string
	}{
		"Case1-1: A=B": {
			input:       "CustomizedTestField=Test",
			expectedKey: "CustomizedTestField",
			expectedVal: "Test",
			expectedOp:  "=",
		},
		"Case1-2: A=\"B\"": {
			input:       "CustomizedTestField=\"Test\"",
			expectedKey: "CustomizedTestField",
			expectedVal: "\"Test\"",
			expectedOp:  "=",
		},
		"Case1-3: A='B'": {
			input:       "CustomizedTestField='Test'",
			expectedKey: "CustomizedTestField",
			expectedVal: "'Test'",
			expectedOp:  "=",
		},
		"Case2: A<=B": {
			input:       "CustomizedTestField<=Test",
			expectedKey: "CustomizedTestField",
			expectedVal: "Test",
			expectedOp:  "<=",
		},
		"Case3: A>=B": {
			input:       "CustomizedTestField>=Test",
			expectedKey: "CustomizedTestField",
			expectedVal: "Test",
			expectedOp:  ">=",
		},
		"Case4: A = B": {
			input:       "CustomizedTestField = Test",
			expectedKey: "CustomizedTestField",
			expectedVal: "Test",
			expectedOp:  "=",
		},
		"Case5: A <= B": {
			input:       "CustomizedTestField <= Test",
			expectedKey: "CustomizedTestField",
			expectedVal: "Test",
			expectedOp:  "<=",
		},
		"Case6: A >= B": {
			input:       "CustomizedTestField >= Test",
			expectedKey: "CustomizedTestField",
			expectedVal: "Test",
			expectedOp:  ">=",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				key, val, op := splitElement(test.input)
				assert.Equal(t, test.expectedKey, key)
				assert.Equal(t, test.expectedVal, val)
				assert.Equal(t, test.expectedOp, op)
			})
		})
	}
}
