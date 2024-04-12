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

package pinot

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

func TestConvertSearchResultToVisibilityRecord(t *testing.T) {
	columnName := []string{"WorkflowID", "RunID", "WorkflowType", "DomainID", "StartTime", "ExecutionTime", "CloseTime", "CloseStatus", "HistoryLength", "TaskList", "IsCron", "NumClusters", "UpdateTime", "Attr"}
	closeStatus := types.WorkflowExecutionCloseStatusFailed

	tests := map[string]struct {
		inputColumnNames         []string
		inputHit                 []interface{}
		expectedVisibilityRecord *p.InternalVisibilityWorkflowExecutionInfo
		memoCheck                bool
	}{
		"Case1: nil result": {
			inputColumnNames:         nil,
			inputHit:                 []interface{}{"wfid", "rid", "wftype", "domainid", testEarliestTime, testEarliestTime, testLatestTime, -1, 1, "tsklst", true, 1, testEarliestTime, "{}"},
			expectedVisibilityRecord: nil,
		},
		"Case2-1: marshal system key error case": {
			inputColumnNames:         columnName,
			inputHit:                 []interface{}{"wfid", "rid", "wftype", "domainid", testEarliestTime, testEarliestTime, testLatestTime, 1, 1, "tsklst", true, 1, testEarliestTime, make(chan int)},
			expectedVisibilityRecord: nil,
		},
		"Case2-2: unmarshal system key error case": {
			inputColumnNames:         columnName,
			inputHit:                 []interface{}{"wfid", "rid", "wftype", "domainid", testEarliestTime, testEarliestTime, testLatestTime, 1, "1", "tsklst", true, 1, testEarliestTime, `{"CustomStringField": "customA and customB or customC", "CustomDoubleField": 3.14}`},
			expectedVisibilityRecord: nil,
		},
		"Case3-1: closed wf with everything except for an empty Attr": {
			inputColumnNames: columnName,
			inputHit:         []interface{}{"wfid", "rid", "wftype", "domainid", testEarliestTime, testEarliestTime, testLatestTime, -1, 1, "tsklst", true, 1, testEarliestTime, "{}"},
			expectedVisibilityRecord: &p.InternalVisibilityWorkflowExecutionInfo{
				DomainID:         "domainid",
				WorkflowType:     "wftype",
				WorkflowID:       "wfid",
				RunID:            "rid",
				TypeName:         "wftype",
				StartTime:        time.UnixMilli(testEarliestTime),
				ExecutionTime:    time.UnixMilli(testEarliestTime),
				CloseTime:        time.UnixMilli(testLatestTime),
				Status:           nil,
				HistoryLength:    1,
				Memo:             nil,
				TaskList:         "tsklst",
				IsCron:           true,
				NumClusters:      1,
				UpdateTime:       time.UnixMilli(testEarliestTime),
				SearchAttributes: map[string]interface{}{},
				ShardID:          0,
			},
		},
		"Case3-2: closed wf with everything": {
			inputColumnNames: columnName,
			inputHit:         []interface{}{"wfid", "rid", "wftype", "domainid", testEarliestTime, testEarliestTime, testLatestTime, 1, 1, "tsklst", true, 1, testEarliestTime, `{"CustomStringField": "customA and customB or customC", "CustomDoubleField": 3.14}`},
			expectedVisibilityRecord: &p.InternalVisibilityWorkflowExecutionInfo{
				DomainID:         "domainid",
				WorkflowType:     "wftype",
				WorkflowID:       "wfid",
				RunID:            "rid",
				TypeName:         "wftype",
				StartTime:        time.UnixMilli(testEarliestTime),
				ExecutionTime:    time.UnixMilli(testEarliestTime),
				CloseTime:        time.UnixMilli(testLatestTime),
				Status:           &closeStatus,
				HistoryLength:    1,
				Memo:             nil,
				TaskList:         "tsklst",
				IsCron:           true,
				NumClusters:      1,
				UpdateTime:       time.UnixMilli(testEarliestTime),
				SearchAttributes: map[string]interface{}{"CustomStringField": "customA and customB or customC", "CustomDoubleField": 3.14},
				ShardID:          0,
			},
		},
		"Case4: open wf with everything": {
			inputColumnNames: columnName,
			inputHit: []interface{}{"wfid", "rid", "wftype", "domainid", testEarliestTime, testEarliestTime, -1, -1, -1,
				"tsklst", true, 1, testEarliestTime, fmt.Sprint(`{"CustomStringField": "customA and customB or customC", "CustomDoubleField": 3.14}`)},
			expectedVisibilityRecord: &p.InternalVisibilityWorkflowExecutionInfo{
				DomainID:         "domainid",
				WorkflowType:     "wftype",
				WorkflowID:       "wfid",
				RunID:            "rid",
				TypeName:         "wftype",
				StartTime:        time.UnixMilli(testEarliestTime),
				ExecutionTime:    time.UnixMilli(testEarliestTime),
				Memo:             nil,
				TaskList:         "tsklst",
				IsCron:           true,
				NumClusters:      1,
				UpdateTime:       time.UnixMilli(testEarliestTime),
				SearchAttributes: map[string]interface{}{"CustomStringField": "customA and customB or customC", "CustomDoubleField": 3.14},
				ShardID:          0,
			},
			memoCheck: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				visibilityRecord := ConvertSearchResultToVisibilityRecord(test.inputHit, test.inputColumnNames, log.NewNoop())
				assert.Equal(t, test.expectedVisibilityRecord, visibilityRecord)
			})
		})
	}
}

func TestConvertMemo_easeCase(t *testing.T) {
	tests := map[string]struct {
		memo        *p.DataBlob
		badData     interface{}
		badEncoding interface{}
	}{
		"Case1: easy case": {
			memo: p.NewDataBlob([]byte("test memo"), common.EncodingTypeJSON),
		},
		"Case2: weird case": {
			memo: p.NewDataBlob([]byte{0, 0, 0, 0}, common.EncodingTypeJSON),
		},
		"Case3: nil case": {
			memo: p.NewDataBlob(nil, common.EncodingTypeJSON),
		},
		"Case4-1: badData case": {
			badData:     make(chan int),
			badEncoding: common.EncodingTypeJSON,
		},
		"Case4-2: badEncoding case": {
			badEncoding: make(chan int),
			badData:     []byte("test"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				testMemoData, testMemoEncoding, err := test.memo.GetVisibilityStoreInfo()
				assert.NoError(t, err)
				res, err := convertMemo(testMemoData, testMemoEncoding)
				if test.badData != nil || test.badEncoding != nil {
					res, err = convertMemo(test.badData, test.badEncoding)
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, test.memo, res)
				}
			})
		})
	}
}
