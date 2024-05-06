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

package os2

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/opensearch-project/opensearch-go/v2/opensearchutil"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/elasticsearch/bulk"
)

func TestStart(t *testing.T) {
	ctx := context.Background()
	processor := &bulkProcessor{}
	if err := processor.Start(ctx); err != nil {
		t.Errorf("Start() error = %v, wantErr %v", err, nil)
	}
}

func TestStopAndClose(t *testing.T) {
	osClient, testServer := getSecureMockOS2Client(t, nil, false)
	defer testServer.Close()

	params := bulk.BulkProcessorParameters{
		FlushInterval: 1,
		NumOfWorkers:  1,
	}
	processor, err := osClient.RunBulkProcessor(context.Background(), &params)
	assert.NoError(t, err)

	if err := processor.Stop(); err != nil {
		t.Errorf("Stop() error = %v, wantErr %v", err, nil)
	}
}

func TestFlush(t *testing.T) {
	processor := &bulkProcessor{}
	if err := processor.Flush(); err != nil {
		t.Errorf("Flush() error = %v, wantErr %v", err, nil)
	}
}

func TestAdd(t *testing.T) {
	testcases := []struct {
		name      string
		input     bulk.GenericBulkableAddRequest
		handler   http.HandlerFunc
		expectErr bool
	}{
		{
			name: "delete request",
			input: bulk.GenericBulkableAddRequest{
				RequestType: bulk.BulkableDeleteRequest,
				Index:       "test-index",
				ID:          "test-id",
				Version:     1,
				VersionType: "external",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			expectErr: false,
		},
		{
			name: "index request failure",
			input: bulk.GenericBulkableAddRequest{
				RequestType: bulk.BulkableIndexRequest,
				Index:       "test-index",
				ID:          "test-id",
				Doc:         make(chan int),
			},
			handler:   nil,
			expectErr: false,
		},
		{
			name: "index request success",
			input: bulk.GenericBulkableAddRequest{
				RequestType: bulk.BulkableIndexRequest,
				Index:       "test-index",
				ID:          "test-id",
				Doc:         map[string]interface{}{"field": "value"},
			},
			handler:   nil,
			expectErr: false,
		},
		{
			name: "create request failure",
			input: bulk.GenericBulkableAddRequest{
				RequestType: bulk.BulkableCreateRequest,
				Index:       "test-index",
				ID:          "test-id",
				Doc:         make(chan int),
			},
			handler:   nil,
			expectErr: false,
		},
		{
			name: "create request success",
			input: bulk.GenericBulkableAddRequest{
				RequestType: bulk.BulkableCreateRequest,
				Index:       "test-index",
				ID:          "test-id",
				Doc:         map[string]interface{}{"field": "value"},
			},
			handler:   nil,
			expectErr: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			osClient, testServer := getSecureMockOS2Client(t, tc.handler, false)
			defer testServer.Close()

			params := bulk.BulkProcessorParameters{
				FlushInterval: 1,
				NumOfWorkers:  1,
			}
			processor, err := osClient.RunBulkProcessor(context.Background(), &params)
			assert.NoError(t, err)
			processor.Add(&tc.input)
		})
	}
}

func TestProcessCallback(t *testing.T) {
	osClient, testServer := getSecureMockOS2Client(t, nil, false)
	defer testServer.Close()
	processor, _ := opensearchutil.NewBulkIndexer(opensearchutil.BulkIndexerConfig{
		Client:        osClient.client,
		FlushInterval: 1,
		NumWorkers:    1,
		Decoder:       &NumberDecoder{},
	})
	bulkProcessor := &bulkProcessor{
		processor: processor,
		before:    beforeFunc,
		after:     func(int64, []bulk.GenericBulkableRequest, *bulk.GenericBulkResponse, *bulk.GenericError) {},
	}

	callBackRequest := bulk.NewBulkIndexRequest().ID("test-id").Index("test-index").Doc(map[string]interface{}{"field": "value"})
	req := opensearchutil.BulkIndexerItem{Index: "test-index", DocumentID: "test-id"}
	res := opensearchutil.BulkIndexerResponseItem{Index: "test-index", DocumentID: "test-id", Status: 200}

	tests := []struct {
		name        string
		onSuccess   bool
		errorPassed error
		expectError bool
	}{
		{"Success Callback", true, nil, false},
		{"Failure Callback", false, errors.New("test error"), true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bulkProcessor.processCallback(0, callBackRequest, req, res, test.onSuccess, test.errorPassed)
		})
	}
}

func beforeFunc(executionId int64, requests []bulk.GenericBulkableRequest) {}

func TestRunBulkProcessor(t *testing.T) {
	osClient, testServer := getSecureMockOS2Client(t, nil, false)
	defer testServer.Close()

	testcases := []struct {
		name      string
		input     bulk.BulkProcessorParameters
		expectErr bool
	}{
		{
			name: "normal",
			input: bulk.BulkProcessorParameters{
				FlushInterval: 1,
				NumOfWorkers:  1,
			},
			expectErr: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			processor, err := osClient.RunBulkProcessor(context.Background(), &tc.input)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, processor)
			}
		})
	}
}

func TestUnmarshalFromReader(t *testing.T) {
	// test unmarshalFromReader
	testcases := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "normal",
			input:     `{"status":"ok"}`,
			expectErr: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			decoder := &NumberDecoder{}
			reader := createReaderFromString(tc.input)

			var response opensearchutil.BulkIndexerResponse
			err := decoder.UnmarshalFromReader(reader, &response)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	// test decode
	testcases := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "normal",
			input:     `{"status":"ok"}`,
			expectErr: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			decoder := &NumberDecoder{}
			reader := createReaderFromString(tc.input)

			var response opensearchutil.BulkIndexerResponse
			err := decoder.Decode(reader, &response)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Function to create a sample reader from a string
func createReaderFromString(s string) io.Reader {
	return bytes.NewBufferString(s)
}
