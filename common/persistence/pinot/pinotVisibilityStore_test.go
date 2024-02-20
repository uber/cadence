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

package pinotvisibility

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/definition"
	"github.com/uber/cadence/common/log"
	p "github.com/uber/cadence/common/persistence"
	pnt "github.com/uber/cadence/common/pinot"
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

	validSearchAttr = definition.GetDefaultIndexedKeys()

	visibilityStore = pinotVisibilityStore{
		pinotClient:         nil,
		producer:            nil,
		logger:              log.NewNoop(),
		config:              nil,
		pinotQueryValidator: pnt.NewPinotQueryValidator(validSearchAttr),
	}
)

func TestGetCountWorkflowExecutionsQuery(t *testing.T) {
	request := &p.CountWorkflowExecutionsRequest{
		DomainUUID: testDomainID,
		Domain:     testDomain,
		Query:      "WorkflowID = 'wfid'",
	}

	result := visibilityStore.getCountWorkflowExecutionsQuery(testTableName, request)
	expectResult := fmt.Sprintf(`SELECT COUNT(*)
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND IsDeleted = false
AND WorkflowID = 'wfid'
`, testTableName)

	assert.Equal(t, result, expectResult)

	nilResult := visibilityStore.getCountWorkflowExecutionsQuery(testTableName, nil)
	assert.Equal(t, nilResult, "")
}

func TestGetListWorkflowExecutionQuery(t *testing.T) {

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
				Query:         "`Attr.CustomKeywordField` = 'keywordCustomized'",
			},
			expectedOutput: fmt.Sprintf(
				`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND IsDeleted = false
AND (JSON_MATCH(Attr, '"$.CustomKeywordField"=''keywordCustomized''') or JSON_MATCH(Attr, '"$.CustomKeywordField[*]"=''keywordCustomized'''))
Order BY StartTime DESC
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
AND IsDeleted = false
AND JSON_MATCH(Attr, '"$.CustomIntField"=''2''') and (JSON_MATCH(Attr, '"$.CustomKeywordField"=''Update2''') or JSON_MATCH(Attr, '"$.CustomKeywordField[*]"=''Update2'''))
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
				Query:         "CustomKeywordField = 'keywordCustomized' and CustomStringField = 'String and or order by'",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND IsDeleted = false
AND (JSON_MATCH(Attr, '"$.CustomKeywordField"=''keywordCustomized''') or JSON_MATCH(Attr, '"$.CustomKeywordField[*]"=''keywordCustomized''')) and (JSON_MATCH(Attr, '"$.CustomStringField" is not null') AND REGEXP_LIKE(JSON_EXTRACT_SCALAR(Attr, '$.CustomStringField', 'string'), 'String and or order by*'))
Order BY StartTime DESC
LIMIT 0, 10
`, testTableName),
		},

		"complete request with or query & customized attributes": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "CustomStringField = 'Or' or CustomStringField = 'and' Order by StartTime DESC",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND IsDeleted = false
AND ((JSON_MATCH(Attr, '"$.CustomStringField" is not null') AND REGEXP_LIKE(JSON_EXTRACT_SCALAR(Attr, '$.CustomStringField', 'string'), 'Or*')) or (JSON_MATCH(Attr, '"$.CustomStringField" is not null') AND REGEXP_LIKE(JSON_EXTRACT_SCALAR(Attr, '$.CustomStringField', 'string'), 'and*')))
Order by StartTime DESC
LIMIT 0, 10
`, testTableName),
		},

		"complex query": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "WorkflowID = 'wid' and ((CustomStringField = 'custom and custom2 or custom3 order by') or CustomIntField between 1 and 10)",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND IsDeleted = false
AND WorkflowID = 'wid' and ((JSON_MATCH(Attr, '"$.CustomStringField" is not null') AND REGEXP_LIKE(JSON_EXTRACT_SCALAR(Attr, '$.CustomStringField', 'string'), 'custom and custom2 or custom3 order by*')) or (JSON_MATCH(Attr, '"$.CustomIntField" is not null') AND CAST(JSON_EXTRACT_SCALAR(Attr, '$.CustomIntField') AS INT) >= 1 AND CAST(JSON_EXTRACT_SCALAR(Attr, '$.CustomIntField') AS INT) <= 10))
Order BY StartTime DESC
LIMIT 0, 10
`, testTableName),
		},

		"or clause with custom attributes": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "CustomIntField = 1 or CustomIntField = 2",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND IsDeleted = false
AND (JSON_MATCH(Attr, '"$.CustomIntField"=''1''') or JSON_MATCH(Attr, '"$.CustomIntField"=''2'''))
Order BY StartTime DESC
LIMIT 0, 10
`, testTableName),
		},

		"complete request with customized query with missing": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: nil,
				Query:         "CloseTime = missing anD WorkflowType = 'some-test-workflow'",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND IsDeleted = false
AND CloseTime = -1 and WorkflowType = 'some-test-workflow'
Order BY StartTime DESC
LIMIT 0, 10
`, testTableName),
		},

		"complete request with customized query with NextPageToken": {
			input: &p.ListWorkflowExecutionsByQueryRequest{
				DomainUUID:    testDomainID,
				Domain:        testDomain,
				PageSize:      testPageSize,
				NextPageToken: serializedToken,
				Query:         "CloseStatus < 0 and CustomKeywordField = 'keywordCustomized' AND CustomIntField<=10 and CustomStringField = 'String field is for text' Order by DomainID Desc",
			},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND IsDeleted = false
AND CloseStatus < 0 and (JSON_MATCH(Attr, '"$.CustomKeywordField"=''keywordCustomized''') or JSON_MATCH(Attr, '"$.CustomKeywordField[*]"=''keywordCustomized''')) and (JSON_MATCH(Attr, '"$.CustomIntField" is not null') AND CAST(JSON_EXTRACT_SCALAR(Attr, '$.CustomIntField') AS INT) <= 10) and (JSON_MATCH(Attr, '"$.CustomStringField" is not null') AND REGEXP_LIKE(JSON_EXTRACT_SCALAR(Attr, '$.CustomStringField', 'string'), 'String field is for text*'))
Order by DomainID Desc
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
AND IsDeleted = false
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
AND IsDeleted = false
AND CloseStatus < 0
Order BY StartTime DESC
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
AND IsDeleted = false
LIMIT 0, 10
`, testTableName),
		},

		"empty request": {
			input: &p.ListWorkflowExecutionsByQueryRequest{},
			expectedOutput: fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = ''
AND IsDeleted = false
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
			output, err := visibilityStore.getListWorkflowExecutionsByQueryQuery(testTableName, test.input)
			assert.Equal(t, test.expectedOutput, output)
			assert.NoError(t, err)
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

	closeResult, err1 := getListWorkflowExecutionsQuery(testTableName, request, true)
	openResult, err2 := getListWorkflowExecutionsQuery(testTableName, request, false)
	nilResult, err3 := getListWorkflowExecutionsQuery(testTableName, nil, true)
	expectCloseResult := fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND IsDeleted = false
AND CloseTime BETWEEN 1547596871371 AND 2547596873371
AND CloseStatus >= 0
Order BY StartTime DESC
LIMIT 0, 10
`, testTableName)
	expectOpenResult := fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND IsDeleted = false
AND StartTime BETWEEN 1547596871371 AND 2547596873371
AND CloseStatus < 0
AND CloseTime = -1
Order BY StartTime DESC
LIMIT 0, 10
`, testTableName)
	expectNilResult := ""

	assert.Equal(t, closeResult, expectCloseResult)
	assert.Equal(t, openResult, expectOpenResult)
	assert.Equal(t, nilResult, expectNilResult)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
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

	closeResult, err1 := getListWorkflowExecutionsByTypeQuery(testTableName, request, true)
	openResult, err2 := getListWorkflowExecutionsByTypeQuery(testTableName, request, false)
	nilResult, err3 := getListWorkflowExecutionsByTypeQuery(testTableName, nil, true)
	expectCloseResult := fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND IsDeleted = false
AND WorkflowType = 'test-wf-type'
AND CloseTime BETWEEN 1547596871371 AND 2547596873371
AND CloseStatus >= 0
Order BY StartTime DESC
LIMIT 0, 10
`, testTableName)
	expectOpenResult := fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND IsDeleted = false
AND WorkflowType = 'test-wf-type'
AND StartTime BETWEEN 1547596871371 AND 2547596873371
AND CloseStatus < 0
AND CloseTime = -1
Order BY StartTime DESC
LIMIT 0, 10
`, testTableName)
	expectNilResult := ""

	assert.Equal(t, closeResult, expectCloseResult)
	assert.Equal(t, openResult, expectOpenResult)
	assert.Equal(t, nilResult, expectNilResult)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
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

	closeResult, err1 := getListWorkflowExecutionsByWorkflowIDQuery(testTableName, request, true)
	openResult, err2 := getListWorkflowExecutionsByWorkflowIDQuery(testTableName, request, false)
	nilResult, err3 := getListWorkflowExecutionsByWorkflowIDQuery(testTableName, nil, true)
	expectCloseResult := fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND IsDeleted = false
AND WorkflowID = 'test-wid'
AND CloseTime BETWEEN 1547596871371 AND 2547596873371
AND CloseStatus >= 0
Order BY StartTime DESC
LIMIT 0, 10
`, testTableName)
	expectOpenResult := fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND IsDeleted = false
AND WorkflowID = 'test-wid'
AND StartTime BETWEEN 1547596871371 AND 2547596873371
AND CloseStatus < 0
AND CloseTime = -1
Order BY StartTime DESC
LIMIT 0, 10
`, testTableName)
	expectNilResult := ""

	assert.Equal(t, closeResult, expectCloseResult)
	assert.Equal(t, openResult, expectOpenResult)
	assert.Equal(t, nilResult, expectNilResult)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
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

	closeResult, err1 := getListWorkflowExecutionsByStatusQuery(testTableName, request)
	nilResult, err2 := getListWorkflowExecutionsByStatusQuery(testTableName, nil)
	expectCloseResult := fmt.Sprintf(`SELECT *
FROM %s
WHERE DomainID = 'bfd5c907-f899-4baf-a7b2-2ab85e623ebd'
AND IsDeleted = false
AND CloseStatus = '0'
AND CloseTime BETWEEN 1547596872371 AND 2547596872371
Order BY StartTime DESC
LIMIT 0, 10
`, testTableName)
	expectNilResult := ""

	assert.Equal(t, expectCloseResult, closeResult)
	assert.Equal(t, expectNilResult, nilResult)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
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
AND IsDeleted = false
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
AND IsDeleted = false
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
AND IsDeleted = false
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
				element, orderBy := parseOrderBy(test.input)
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
