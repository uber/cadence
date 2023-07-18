package persistence

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/types"
	"testing"
)

var (
	testIndex             = "test-index"
	testDomain            = "test-domain"
	testDomainID          = "bfd5c907-f899-4baf-a7b2-2ab85e623ebd"
	testPageSize          = 10
	testEarliestTime      = int64(1547596872371000000)
	testLatestTime        = int64(2547596872371000000)
	testWorkflowType      = "test-wf-type"
	testWorkflowID        = "test-wid"
	testCloseStatus       = int32(1)
	testTableName         = "test-table-name"
	testRunID             = "test-run-id"
	testSearchAttributes1 = map[string]interface{}{"TestAttr1": "val1", "TestAttr2": 2, "TestAttr3": false}
	testSearchAttributes2 = map[string]interface{}{"TestAttr1": "val2", "TestAttr2": 2, "TestAttr3": false}
	testSearchAttributes3 = map[string]interface{}{"TestAttr2": 2, "TestAttr3": false}
)

func TestInterfaceToMap(t *testing.T) {
	tests := map[string]struct {
		input          interface{}
		expectedResult map[string][]byte
	}{
		"Case1: nil input case": {
			input:          nil,
			expectedResult: map[string][]byte{},
		},
		"Case2: empty input case": {
			input:          "",
			expectedResult: map[string][]byte{},
		},
		"Case3: normal input case": {
			input:          transferMap(testSearchAttributes1),
			expectedResult: transferMap(testSearchAttributes1),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				actualResult := interfaceToMap(test.input)
				assert.Equal(t, test.expectedResult, actualResult)
			})
		})
	}

	// Case4: panic case
	assert.Panics(t, func() {
		interfaceToMap(0)
	})
	assert.Panics(t, func() {
		interfaceToMap("0")
	})
	assert.Panics(t, func() {
		interfaceToMap(true)
	})
}

func TestCompareSearchAttributes(t *testing.T) {
	tests := map[string]struct {
		pinotInput     interface{}
		esInput        interface{}
		expectedResult error
	}{
		"Case1: pass case": {
			pinotInput:     &types.SearchAttributes{IndexedFields: transferMap(testSearchAttributes1)},
			esInput:        &types.SearchAttributes{IndexedFields: transferMap(testSearchAttributes1)},
			expectedResult: nil,
		},
		"Case2: error case": {
			pinotInput:     &types.SearchAttributes{IndexedFields: transferMap(testSearchAttributes1)},
			esInput:        &types.SearchAttributes{IndexedFields: transferMap(testSearchAttributes2)},
			expectedResult: fmt.Errorf(fmt.Sprintf("Comparison Failed: response.%s are not equal. ES value = \"%s\", Pinot value = \"%s\"", "TestAttr1", "val2", "val1")),
		},
		"Case3: pass case with different response": {
			pinotInput:     &types.SearchAttributes{IndexedFields: transferMap(testSearchAttributes1)},
			esInput:        &types.SearchAttributes{IndexedFields: transferMap(testSearchAttributes3)},
			expectedResult: nil,
		},
		"Case4: error case with different response": {
			pinotInput:     &types.SearchAttributes{IndexedFields: transferMap(testSearchAttributes3)},
			esInput:        &types.SearchAttributes{IndexedFields: transferMap(testSearchAttributes2)},
			expectedResult: fmt.Errorf(fmt.Sprintf("Comparison Failed: response.%s are not equal. ES value = \"%s\", Pinot value = %s", "TestAttr1", "val2", "")),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				actualResult := compareSearchAttributes(test.esInput, test.pinotInput)
				assert.Equal(t, test.expectedResult, actualResult)
			})
		})
	}
}

func TestCompareExecutions(t *testing.T) {
	tests := map[string]struct {
		pinotInput     interface{}
		esInput        interface{}
		expectedResult error
	}{
		"Case1: pass case": {
			pinotInput: &types.WorkflowExecution{
				WorkflowID: testWorkflowID,
				RunID:      testRunID,
			},
			esInput: &types.WorkflowExecution{
				WorkflowID: testWorkflowID,
				RunID:      testRunID,
			}, expectedResult: nil,
		},
		"Case2: error case": {
			pinotInput: &types.WorkflowExecution{
				WorkflowID: testWorkflowID,
				RunID:      "testRunID",
			},
			esInput: &types.WorkflowExecution{
				WorkflowID: testWorkflowID,
				RunID:      testRunID,
			}, expectedResult: fmt.Errorf(fmt.Sprintf("Comparison Failed: Execution.RunID are not equal. ES value = test-run-id, Pinot value = testRunID")),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				actualResult := compareExecutions(test.esInput, test.pinotInput)
				assert.Equal(t, test.expectedResult, actualResult)
			})
		})
	}
}

func TestCompareType(t *testing.T) {
	tests := map[string]struct {
		pinotInput     interface{}
		esInput        interface{}
		expectedResult error
	}{
		"Case1: pass case": {
			pinotInput:     &types.WorkflowType{Name: testWorkflowType},
			esInput:        &types.WorkflowType{Name: testWorkflowType},
			expectedResult: nil,
		},
		"Case2: error case": {
			pinotInput:     &types.WorkflowType{Name: "testWorkflowType"},
			esInput:        &types.WorkflowType{Name: testWorkflowType},
			expectedResult: fmt.Errorf("Comparison Failed: WorkflowTypes are not equal. ES value = test-wf-type, Pinot value = testWorkflowType"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				actualResult := compareType(test.esInput, test.pinotInput)
				assert.Equal(t, test.expectedResult, actualResult)
			})
		})
	}
}

func TestCompareCloseStatus(t *testing.T) {
	testVal1 := types.WorkflowExecutionCloseStatus(0)
	testVal2 := types.WorkflowExecutionCloseStatus(1)

	tests := map[string]struct {
		pinotInput     interface{}
		esInput        interface{}
		expectedResult error
	}{
		"Case1: pass case": {
			pinotInput:     &testVal1,
			esInput:        &testVal1,
			expectedResult: nil,
		},
		"Case2: error case": {
			pinotInput:     &testVal1,
			esInput:        &testVal2,
			expectedResult: fmt.Errorf("Comparison Failed: WorkflowExecutionCloseStatus are not equal. ES value = FAILED, Pinot value = COMPLETED"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				actualResult := compareCloseStatus(test.esInput, test.pinotInput)
				assert.Equal(t, test.expectedResult, actualResult)
			})
		})
	}
}

func TestCompareListWorkflowExecutionInfo(t *testing.T) {
	testSearchAttributeMap1 := transferMap(testSearchAttributes1)
	testSearchAttributeMap2 := transferMap(testSearchAttributes2)

	tests := map[string]struct {
		esInfo         *types.WorkflowExecutionInfo
		pinotInfo      *types.WorkflowExecutionInfo
		expectedResult error
	}{
		"Case1: pass case": {
			esInfo: &types.WorkflowExecutionInfo{
				Execution: &types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      testRunID,
				},
				Type:             &types.WorkflowType{Name: testWorkflowType},
				StartTime:        &testEarliestTime,
				CloseTime:        &testLatestTime,
				SearchAttributes: nil,
			},
			pinotInfo: &types.WorkflowExecutionInfo{
				Execution: &types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      testRunID,
				},
				Type:             &types.WorkflowType{Name: testWorkflowType},
				StartTime:        &testEarliestTime,
				CloseTime:        &testLatestTime,
				SearchAttributes: nil,
			},
			expectedResult: nil,
		},
		"Case2: pass case with search attributes": {
			esInfo: &types.WorkflowExecutionInfo{
				Execution: &types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      testRunID,
				},
				Type:             &types.WorkflowType{Name: testWorkflowType},
				StartTime:        &testEarliestTime,
				CloseTime:        &testLatestTime,
				SearchAttributes: &types.SearchAttributes{IndexedFields: testSearchAttributeMap1},
			},
			pinotInfo: &types.WorkflowExecutionInfo{
				Execution: &types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      testRunID,
				},
				Type:             &types.WorkflowType{Name: testWorkflowType},
				StartTime:        &testEarliestTime,
				CloseTime:        &testLatestTime,
				SearchAttributes: &types.SearchAttributes{IndexedFields: testSearchAttributeMap1},
			},
			expectedResult: nil,
		},
		"Case3: error case with wrong type": {
			esInfo: &types.WorkflowExecutionInfo{
				Execution: &types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      testRunID,
				},
				Type:             &types.WorkflowType{Name: "testWorkflowType"},
				StartTime:        &testEarliestTime,
				CloseTime:        &testLatestTime,
				SearchAttributes: &types.SearchAttributes{IndexedFields: testSearchAttributeMap1},
			},
			pinotInfo: &types.WorkflowExecutionInfo{
				Execution: &types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      testRunID,
				},
				Type:             &types.WorkflowType{Name: testWorkflowType},
				StartTime:        &testEarliestTime,
				CloseTime:        &testLatestTime,
				SearchAttributes: &types.SearchAttributes{IndexedFields: testSearchAttributeMap1},
			},
			expectedResult: fmt.Errorf("Comparison Failed: WorkflowTypes are not equal. ES value = testWorkflowType, Pinot value = test-wf-type"),
		},
		"Case4: error case with wrong workflowID": {
			esInfo: &types.WorkflowExecutionInfo{
				Execution: &types.WorkflowExecution{
					WorkflowID: "testWorkflowID",
					RunID:      testRunID,
				},
				Type:             &types.WorkflowType{Name: testWorkflowType},
				StartTime:        &testEarliestTime,
				CloseTime:        &testLatestTime,
				SearchAttributes: &types.SearchAttributes{IndexedFields: testSearchAttributeMap1},
			},
			pinotInfo: &types.WorkflowExecutionInfo{
				Execution: &types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      testRunID,
				},
				Type:             &types.WorkflowType{Name: testWorkflowType},
				StartTime:        &testEarliestTime,
				CloseTime:        &testLatestTime,
				SearchAttributes: &types.SearchAttributes{IndexedFields: testSearchAttributeMap1},
			},
			expectedResult: fmt.Errorf("Comparison Failed: Execution.WorkflowID are not equal. ES value = testWorkflowID, Pinot value = test-wid"),
		},
		"Case5: error case with wrong runID": {
			esInfo: &types.WorkflowExecutionInfo{
				Execution: &types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      "testRunID",
				},
				Type:             &types.WorkflowType{Name: testWorkflowType},
				StartTime:        &testEarliestTime,
				CloseTime:        &testLatestTime,
				SearchAttributes: &types.SearchAttributes{IndexedFields: testSearchAttributeMap1},
			},
			pinotInfo: &types.WorkflowExecutionInfo{
				Execution: &types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      testRunID,
				},
				Type:             &types.WorkflowType{Name: testWorkflowType},
				StartTime:        &testEarliestTime,
				CloseTime:        &testLatestTime,
				SearchAttributes: &types.SearchAttributes{IndexedFields: testSearchAttributeMap1},
			},
			expectedResult: fmt.Errorf("Comparison Failed: Execution.RunID are not equal. ES value = testRunID, Pinot value = test-run-id"),
		},
		"Case6: error case with wrong SearchAttributes": {
			esInfo: &types.WorkflowExecutionInfo{
				Execution: &types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      testRunID,
				},
				Type:             &types.WorkflowType{Name: testWorkflowType},
				StartTime:        &testEarliestTime,
				CloseTime:        &testLatestTime,
				SearchAttributes: &types.SearchAttributes{IndexedFields: testSearchAttributeMap1},
			},
			pinotInfo: &types.WorkflowExecutionInfo{
				Execution: &types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      testRunID,
				},
				Type:             &types.WorkflowType{Name: testWorkflowType},
				StartTime:        &testEarliestTime,
				CloseTime:        &testLatestTime,
				SearchAttributes: &types.SearchAttributes{IndexedFields: testSearchAttributeMap2},
			},
			expectedResult: fmt.Errorf("Comparison Failed: response.TestAttr1 are not equal. ES value = \"val1\", Pinot value = \"val2\""),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				actualResult := compareListWorkflowExecutionInfo(test.esInfo, test.pinotInfo)
				assert.Equal(t, test.expectedResult, actualResult)
			})
		})
	}
}

func TestCompareListWorkflowExecutions(t *testing.T) {
	testSearchAttributeMap1 := transferMap(testSearchAttributes1)

	tests := map[string]struct {
		esInfo         []*types.WorkflowExecutionInfo
		pinotInfo      []*types.WorkflowExecutionInfo
		expectedResult error
	}{
		"Case1: pass case": {
			esInfo: []*types.WorkflowExecutionInfo{&types.WorkflowExecutionInfo{
				Execution: &types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      testRunID,
				},
				Type:      &types.WorkflowType{Name: testWorkflowType},
				StartTime: &testEarliestTime,
				CloseTime: &testLatestTime,
			},
				&types.WorkflowExecutionInfo{
					Execution: &types.WorkflowExecution{
						WorkflowID: "testWorkflowID",
						RunID:      testRunID,
					},
					Type:             &types.WorkflowType{Name: testWorkflowType},
					StartTime:        &testEarliestTime,
					CloseTime:        &testLatestTime,
					SearchAttributes: &types.SearchAttributes{IndexedFields: testSearchAttributeMap1},
				}},
			pinotInfo: []*types.WorkflowExecutionInfo{&types.WorkflowExecutionInfo{
				Execution: &types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      testRunID,
				},
				Type:      &types.WorkflowType{Name: testWorkflowType},
				StartTime: &testEarliestTime,
				CloseTime: &testLatestTime,
			},
				&types.WorkflowExecutionInfo{
					Execution: &types.WorkflowExecution{
						WorkflowID: "testWorkflowID",
						RunID:      testRunID,
					},
					Type:             &types.WorkflowType{Name: testWorkflowType},
					StartTime:        &testEarliestTime,
					CloseTime:        &testLatestTime,
					SearchAttributes: &types.SearchAttributes{IndexedFields: testSearchAttributeMap1},
				}},
			expectedResult: nil,
		},
		"Case2: nil case": {
			esInfo:         nil,
			pinotInfo:      nil,
			expectedResult: nil,
		},
		"Case3: one nil case": {
			esInfo: []*types.WorkflowExecutionInfo{&types.WorkflowExecutionInfo{
				Execution: &types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      testRunID,
				},
				Type:             &types.WorkflowType{Name: testWorkflowType},
				StartTime:        &testEarliestTime,
				CloseTime:        &testLatestTime,
				SearchAttributes: &types.SearchAttributes{IndexedFields: testSearchAttributeMap1},
			}},
			pinotInfo:      nil,
			expectedResult: fmt.Errorf("Comparison failed. One of the response is nil. "),
		},
		"Case4: length not equal case": {
			esInfo: []*types.WorkflowExecutionInfo{&types.WorkflowExecutionInfo{
				Execution: &types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      testRunID,
				},
				Type:             &types.WorkflowType{Name: testWorkflowType},
				StartTime:        &testEarliestTime,
				CloseTime:        &testLatestTime,
				SearchAttributes: &types.SearchAttributes{IndexedFields: testSearchAttributeMap1},
			},
				&types.WorkflowExecutionInfo{
					Execution: &types.WorkflowExecution{
						WorkflowID: testWorkflowID,
						RunID:      testRunID,
					},
					Type:             &types.WorkflowType{Name: testWorkflowType},
					StartTime:        &testEarliestTime,
					CloseTime:        &testLatestTime,
					SearchAttributes: &types.SearchAttributes{IndexedFields: testSearchAttributeMap1},
				}},
			pinotInfo: []*types.WorkflowExecutionInfo{&types.WorkflowExecutionInfo{
				Execution: &types.WorkflowExecution{
					WorkflowID: testWorkflowID,
					RunID:      testRunID,
				},
				Type:             &types.WorkflowType{Name: testWorkflowType},
				StartTime:        &testEarliestTime,
				CloseTime:        &testLatestTime,
				SearchAttributes: &types.SearchAttributes{IndexedFields: testSearchAttributeMap1},
			}},
			expectedResult: fmt.Errorf("Comparison failed. result length doesn't equal. "),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				actualResult := compareListWorkflowExecutions(test.esInfo, test.pinotInfo)
				assert.Equal(t, test.expectedResult, actualResult)
			})
		})
	}
}

func transferMap(input map[string]interface{}) map[string][]byte {
	res := make(map[string][]byte)
	for key, _ := range input {
		marshalVal, _ := json.Marshal(input[key])
		res[key] = marshalVal
	}
	return res
}
