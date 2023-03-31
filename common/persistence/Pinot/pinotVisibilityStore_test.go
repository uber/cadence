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
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/definition"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/mocks"
	pnt "github.com/uber/cadence/common/pinot"
	"github.com/uber/cadence/common/service"
	"testing"
	"time"

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

	testRequest = &p.InternalListWorkflowExecutionsRequest{
		DomainUUID:   testDomainID,
		Domain:       testDomain,
		PageSize:     testPageSize,
		EarliestTime: time.Unix(0, testEarliestTime),
		LatestTime:   time.Unix(0, testLatestTime),
	}
	testSearchResult = &p.InternalListWorkflowExecutionsResponse{}
	errTestPinot     = errors.New("Pinot error")

	testContextTimeout = 5 * time.Second

	pinotIndexMaxResultWindow = 3

	logger = log.NewNoop()

	config = &service.Config{
		ESIndexMaxResultWindow: dynamicconfig.GetIntPropertyFn(pinotIndexMaxResultWindow),
		ValidSearchAttributes:  dynamicconfig.GetMapPropertyFn(definition.GetDefaultIndexedKeys()),
	}

	visibilityStore = pinotVisibilityStore{
		pinotClient: nil,
		producer:    nil,
		logger:      logger,
		config:      config,
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

func TestRecordWorkflowExecutionStarted(t *testing.T) {
	mockProducer := &mocks.KafkaProducer{}
	mockProducer.On("Publish", mock.Anything, mock.Anything).Return(nil).Once()
	visibilityStore.producer = mockProducer

	// test non-empty request fields match
	request := &p.InternalRecordWorkflowExecutionStartedRequest{}
	request.DomainUUID = "domainID"
	request.WorkflowID = "wid"
	request.RunID = "rid"
	request.WorkflowTypeName = "wfType"
	request.StartTimestamp = time.Unix(0, int64(123))
	request.ExecutionTimestamp = time.Unix(0, int64(321))
	request.TaskID = int64(111)
	request.IsCron = true
	request.NumClusters = 2
	memoBytes := []byte(`test bytes`)
	request.Memo = p.NewDataBlob(memoBytes, common.EncodingTypeThriftRW)
	request.ShardID = 1234

	assert.NotPanics(t, func() {
		visibilityStore.RecordWorkflowExecutionStarted(nil, request)
	})

	panicRequest := &p.InternalRecordWorkflowExecutionStartedRequest{}
	assert.Panics(t, func() {
		visibilityStore.RecordWorkflowExecutionStarted(nil, panicRequest)
	})
}

func TestRecordWorkflowExecutionClosed(t *testing.T) {
	mockProducer := &mocks.KafkaProducer{}
	mockProducer.On("Publish", mock.Anything, mock.Anything).Return(nil).Once()
	visibilityStore.producer = mockProducer

	// test non-empty request fields match
	request := &p.InternalRecordWorkflowExecutionClosedRequest{}
	request.DomainUUID = "domainID"
	request.WorkflowID = "wid"
	request.RunID = "rid"
	request.WorkflowTypeName = "wfType"
	request.StartTimestamp = time.Unix(0, int64(123))
	request.ExecutionTimestamp = time.Unix(0, int64(321))
	request.TaskID = int64(111)
	memoBytes := []byte(`test bytes`)
	request.Memo = p.NewDataBlob(memoBytes, common.EncodingTypeThriftRW)
	request.CloseTimestamp = time.Unix(0, int64(999))
	request.Status = types.WorkflowExecutionCloseStatusTerminated
	request.HistoryLength = int64(20)
	request.IsCron = false
	request.NumClusters = 2
	request.UpdateTimestamp = time.Unix(0, int64(213))
	request.ShardID = 1234

	assert.NotPanics(t, func() {
		visibilityStore.RecordWorkflowExecutionClosed(nil, request)
	})

	panicRequest := &p.InternalRecordWorkflowExecutionClosedRequest{}
	assert.Panics(t, func() {
		visibilityStore.RecordWorkflowExecutionClosed(nil, panicRequest)
	})
}

func TestRecordWorkflowExecutionUninitialized(t *testing.T) {
	mockProducer := &mocks.KafkaProducer{}
	mockProducer.On("Publish", mock.Anything, mock.Anything).Return(nil).Once()
	visibilityStore.producer = mockProducer

	// test non-empty request fields match
	request := &p.InternalRecordWorkflowExecutionUninitializedRequest{}
	request.DomainUUID = "domainID"
	request.WorkflowID = "wid"
	request.RunID = "rid"
	request.WorkflowTypeName = "wfType"
	request.UpdateTimestamp = time.Unix(0, int64(213))
	request.ShardID = 1234

	assert.NotPanics(t, func() {
		visibilityStore.RecordWorkflowExecutionUninitialized(nil, request)
	})

	panicRequest := &p.InternalRecordWorkflowExecutionUninitializedRequest{}
	assert.Panics(t, func() {
		visibilityStore.RecordWorkflowExecutionUninitialized(nil, panicRequest)
	})
}

func TestUpsertWorkflowExecution(t *testing.T) {
	mockProducer := &mocks.KafkaProducer{}
	mockProducer.On("Publish", mock.Anything, mock.Anything).Return(nil).Once()
	visibilityStore.producer = mockProducer

	// test non-empty request fields match
	request := &p.InternalUpsertWorkflowExecutionRequest{
		DomainUUID:         "1",
		WorkflowID:         "1",
		RunID:              "1",
		WorkflowTypeName:   "1",
		StartTimestamp:     time.Time{},
		ExecutionTimestamp: time.Time{},
		WorkflowTimeout:    1,
		TaskID:             1,
		Memo:               &p.DataBlob{},
		TaskList:           "",
		IsCron:             false,
		NumClusters:        0,
		UpdateTimestamp:    time.Time{},
		SearchAttributes:   map[string][]byte{},
		ShardID:            0,
	}

	assert.NotPanics(t, func() {
		visibilityStore.UpsertWorkflowExecution(nil, request)
	})

	panicRequest := &p.InternalUpsertWorkflowExecutionRequest{}
	assert.Panics(t, func() {
		visibilityStore.UpsertWorkflowExecution(nil, panicRequest)
	})
}

func TestListOpenWorkflowExecutions(t *testing.T) {
	client := pnt.NewMockGenericClient(gomock.NewController(t))
	visibilityStore.pinotClient = client

	request := &p.InternalListWorkflowExecutionsRequest{
		DomainUUID:    testDomainID,
		Domain:        testDomain,
		EarliestTime:  time.Unix(0, testEarliestTime),
		LatestTime:    time.Unix(0, testLatestTime),
		PageSize:      testPageSize,
		NextPageToken: nil,
	}

	response := &p.InternalListWorkflowExecutionsResponse{
		Executions:    nil,
		NextPageToken: nil,
	}

	client.EXPECT().Search(gomock.Any()).Return(response, nil)

	assert.NotPanics(t, func() {
		visibilityStore.ListOpenWorkflowExecutions(nil, request)
	})
}

func TestListClosedWorkflowExecutions(t *testing.T) {
	client := pnt.NewMockGenericClient(gomock.NewController(t))
	visibilityStore.pinotClient = client

	request := &p.InternalListWorkflowExecutionsRequest{
		DomainUUID:    testDomainID,
		Domain:        testDomain,
		EarliestTime:  time.Unix(0, testEarliestTime),
		LatestTime:    time.Unix(0, testLatestTime),
		PageSize:      testPageSize,
		NextPageToken: nil,
	}

	response := &p.InternalListWorkflowExecutionsResponse{
		Executions:    nil,
		NextPageToken: nil,
	}

	client.EXPECT().Search(gomock.Any()).Return(response, nil)

	assert.NotPanics(t, func() {
		visibilityStore.ListClosedWorkflowExecutions(nil, request)
	})
}

func TestListOpenWorkflowExecutionsByType(t *testing.T) {
	client := pnt.NewMockGenericClient(gomock.NewController(t))
	visibilityStore.pinotClient = client

	request := &p.InternalListWorkflowExecutionsByTypeRequest{
		InternalListWorkflowExecutionsRequest: p.InternalListWorkflowExecutionsRequest{},
		WorkflowTypeName:                      "",
	}

	response := &p.InternalListWorkflowExecutionsResponse{
		Executions:    nil,
		NextPageToken: nil,
	}

	client.EXPECT().Search(gomock.Any()).Return(response, nil)

	assert.NotPanics(t, func() {
		visibilityStore.ListOpenWorkflowExecutionsByType(nil, request)
	})
}

func TestListClosedWorkflowExecutionsByType(t *testing.T) {
	client := pnt.NewMockGenericClient(gomock.NewController(t))
	visibilityStore.pinotClient = client

	request := &p.InternalListWorkflowExecutionsByTypeRequest{
		InternalListWorkflowExecutionsRequest: p.InternalListWorkflowExecutionsRequest{},
		WorkflowTypeName:                      "",
	}

	response := &p.InternalListWorkflowExecutionsResponse{
		Executions:    nil,
		NextPageToken: nil,
	}

	client.EXPECT().Search(gomock.Any()).Return(response, nil)

	assert.NotPanics(t, func() {
		visibilityStore.ListClosedWorkflowExecutionsByType(nil, request)
	})
}

func TestListOpenWorkflowExecutionsByWorkflowID(t *testing.T) {
	client := pnt.NewMockGenericClient(gomock.NewController(t))
	visibilityStore.pinotClient = client

	request := &p.InternalListWorkflowExecutionsByWorkflowIDRequest{
		InternalListWorkflowExecutionsRequest: p.InternalListWorkflowExecutionsRequest{},
		WorkflowID:                            testWorkflowID,
	}

	response := &p.InternalListWorkflowExecutionsResponse{
		Executions:    nil,
		NextPageToken: nil,
	}

	client.EXPECT().Search(gomock.Any()).Return(response, nil)

	assert.NotPanics(t, func() {
		visibilityStore.ListOpenWorkflowExecutionsByWorkflowID(nil, request)
	})
}

func TestListClosedWorkflowExecutionsByWorkflowID(t *testing.T) {
	client := pnt.NewMockGenericClient(gomock.NewController(t))
	visibilityStore.pinotClient = client

	request := &p.InternalListWorkflowExecutionsByWorkflowIDRequest{
		InternalListWorkflowExecutionsRequest: p.InternalListWorkflowExecutionsRequest{},
		WorkflowID:                            testWorkflowID,
	}

	response := &p.InternalListWorkflowExecutionsResponse{
		Executions:    nil,
		NextPageToken: nil,
	}

	client.EXPECT().Search(gomock.Any()).Return(response, nil)

	assert.NotPanics(t, func() {
		visibilityStore.ListClosedWorkflowExecutionsByWorkflowID(nil, request)
	})
}

func TestListClosedWorkflowExecutionsByStatus(t *testing.T) {
	client := pnt.NewMockGenericClient(gomock.NewController(t))
	visibilityStore.pinotClient = client

	request := &p.InternalListClosedWorkflowExecutionsByStatusRequest{
		InternalListWorkflowExecutionsRequest: p.InternalListWorkflowExecutionsRequest{},
		Status:                                0,
	}

	response := &p.InternalListWorkflowExecutionsResponse{
		Executions:    nil,
		NextPageToken: nil,
	}

	client.EXPECT().Search(gomock.Any()).Return(response, nil)

	assert.NotPanics(t, func() {
		visibilityStore.ListClosedWorkflowExecutionsByStatus(nil, request)
	})
}

func TestScanWorkflowExecutions(t *testing.T) {
	client := pnt.NewMockGenericClient(gomock.NewController(t))
	visibilityStore.pinotClient = client

	request := &p.ListWorkflowExecutionsByQueryRequest{
		DomainUUID:    testDomainID,
		Domain:        testDomain,
		PageSize:      testPageSize,
		NextPageToken: nil,
		Query:         "",
	}

	response := &p.InternalListWorkflowExecutionsResponse{
		Executions:    nil,
		NextPageToken: nil,
	}

	client.EXPECT().Search(gomock.Any()).Return(response, nil)

	assert.NotPanics(t, func() {
		visibilityStore.ScanWorkflowExecutions(nil, request)
	})
}

func TestCountWorkflowExecutions(t *testing.T) {
	client := pnt.NewMockGenericClient(gomock.NewController(t))
	visibilityStore.pinotClient = client

	request := &p.CountWorkflowExecutionsRequest{
		DomainUUID: testDomainID,
		Domain:     testDomain,
		Query:      "",
	}

	client.EXPECT().CountByQuery(gomock.Any()).Return(int64(0), nil)

	assert.NotPanics(t, func() {
		visibilityStore.CountWorkflowExecutions(nil, request)
	})
}
