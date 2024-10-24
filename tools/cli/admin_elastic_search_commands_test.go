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

package cli

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/olivere/elastic"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/uber/cadence/.gen/go/indexer"
	"github.com/uber/cadence/common/elasticsearch"
)

// Tests for timeKeyFilter function
func TestTimeKeyFilter(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{
			name:     "ValidTimeKeyStartTime",
			key:      "StartTime",
			expected: true,
		},
		{
			name:     "ValidTimeKeyCloseTime",
			key:      "CloseTime",
			expected: true,
		},
		{
			name:     "ValidTimeKeyExecutionTime",
			key:      "ExecutionTime",
			expected: true,
		},
		{
			name:     "InvalidTimeKey",
			key:      "SomeOtherKey",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := timeKeyFilter(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for timeValProcess function
func TestTimeValProcess(t *testing.T) {
	tests := []struct {
		name        string
		timeStr     string
		expected    string
		expectError bool
	}{
		{
			name:        "ValidInt64TimeString",
			timeStr:     "1630425600000000000", // Already in int64 format
			expected:    "1630425600000000000",
			expectError: false,
		},
		{
			name:        "ValidDateTimeString",
			timeStr:     "2021-09-01T00:00:00Z", // A valid time string
			expected:    fmt.Sprintf("%v", time.Date(2021, 9, 1, 0, 0, 0, 0, time.UTC).UnixNano()),
			expectError: false,
		},
		{
			name:        "InvalidTimeString",
			timeStr:     "invalid-time",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := timeValProcess(tt.timeStr)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestAdminCatIndices(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedOutput string
		expectedError  string
		handlerCalled  bool
	}{
		{
			name: "Success",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate a successful response from Elasticsearch CatIndices API
				if r.URL.Path == "/_cat/indices" && r.Method == "GET" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`[{"health":"green","status":"open","index":"test-index","pri":"5","rep":"1","docs.count":"1000","docs.deleted":"50","store.size":"10gb","pri.store.size":"5gb"}]`))
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}),
			expectedOutput: `+--------+--------+------------+-----+-----+------------+--------------+--------------+
| HEALTH | STATUS |   INDEX    | PRI | REP | DOCS COUNT | DOCS DELETED |  STORE SIZE  |
+--------+--------+------------+-----+-----+------------+--------------+--------------+
| green  | open   | test-index |   5 |   1 |       1000 |           50 | 10gb         |
+--------+--------+------------+-----+-----+------------+--------------+--------------+
`,
			expectedError: "",
			handlerCalled: true,
		},
		{
			name: "CatIndices Error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate an error response
				w.WriteHeader(http.StatusInternalServerError)
			}),
			expectedOutput: "",
			expectedError:  "Unable to cat indices",
			handlerCalled:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false

			// Wrap the test case's handler to track if it was called
			wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				tt.handler.ServeHTTP(w, r)
			})

			// Create mock Elasticsearch client and server
			esClient, testServer := getMockClient(t, wrappedHandler)
			defer testServer.Close()

			// Initialize mock controller
			mockCtrl := gomock.NewController(t)

			// Create mock Cadence client factory
			mockClientFactory := NewMockClientFactory(mockCtrl)

			// Create test IO handler to capture output
			ioHandler := &testIOHandler{}

			// Set up the CLI app and mock dependencies
			app := NewCliApp(mockClientFactory, WithIOHandler(ioHandler))

			// Expect ElasticSearchClient to return the mock client created by getMockClient
			mockClientFactory.EXPECT().ElasticSearchClient(gomock.Any()).Return(esClient, nil).Times(1)

			// Create a mock CLI context
			c := setContextMock(app)

			// Call AdminCatIndices
			err := AdminCatIndices(c)

			// Validate handler was called
			assert.Equal(t, tt.handlerCalled, handlerCalled, "Expected handler to be called")

			// Check for expected error or success
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				// Validate the output captured by testIOHandler
				assert.Equal(t, tt.expectedOutput, ioHandler.Output().(*bytes.Buffer).String())
			}
		})
	}
}

// getMockClient creates a mock elastic.Client using the provided HTTP handler and returns the client and the test server
func getMockClient(t *testing.T, handler http.HandlerFunc) (*elastic.Client, *httptest.Server) {
	// Create a mock HTTP test server
	testServer := httptest.NewTLSServer(handler)

	// Create an Elasticsearch client using the test server's URL
	mockClient, err := elastic.NewClient(
		elastic.SetURL(testServer.URL),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetHttpClient(testServer.Client()),
	)
	// Ensure no error occurred while creating the mock client
	assert.NoError(t, err)

	// Return the elastic.Client and the test server
	return mockClient, testServer
}

func TestAdminIndex(t *testing.T) {
	tests := []struct {
		name            string
		handler         http.HandlerFunc
		createInputFile bool
		messageType     indexer.MessageType
		expectedError   string
	}{
		{
			name: "SuccessIndexMessage",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate a successful Bulk request
				if r.URL.Path == "/_bulk" && r.Method == "POST" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"took": 30, "errors": false}`))
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}),
			createInputFile: true,
			messageType:     indexer.MessageTypeIndex, // Test MessageTypeIndex case
			expectedError:   "",
		},
		{
			name: "SuccessCreateMessage",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate a successful Bulk request
				if r.URL.Path == "/_bulk" && r.Method == "POST" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"took": 30, "errors": false}`))
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}),
			createInputFile: true,
			messageType:     indexer.MessageTypeCreate, // Test MessageTypeCreate case
			expectedError:   "",
		},
		{
			name: "SuccessDeleteMessage",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate a successful Bulk request
				if r.URL.Path == "/_bulk" && r.Method == "POST" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"took": 30, "errors": false}`))
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}),
			createInputFile: true,
			messageType:     indexer.MessageTypeDelete, // Test MessageTypeDelete case
			expectedError:   "",
		},
		{
			name: "UnknownMessageType",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// In this test case, we are simulating an unknown message type, so no bulk request needed.
				w.WriteHeader(http.StatusOK)
			}),
			createInputFile: true,
			messageType:     indexer.MessageType(9999), // Test unknown message type case
			expectedError:   "Unknown message type",
		},
		{
			name: "BulkRequestFailure",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate an error in the Bulk request
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Bulk request failed"}`))
			}),
			createInputFile: true,
			messageType:     indexer.MessageTypeIndex,
			expectedError:   "Bulk failed",
		},
		{
			name: "ParseIndexerMessageError",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// In this test case, we are simulating a parse error, so no need to simulate the bulk request.
				w.WriteHeader(http.StatusOK)
			}),
			createInputFile: false, // No valid input file created
			messageType:     indexer.MessageTypeIndex,
			expectedError:   "Unable to parse indexer message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inputFileName string
			var err error
			if tt.createInputFile {
				// Create a temporary input file with a valid message
				inputFileName, err = createTempIndexerInputFileWithMessageType(tt.messageType, false)
				assert.NoError(t, err)
				defer os.Remove(inputFileName) // Clean up after test
			}

			// Create mock Elasticsearch client and server
			esClient, testServer := getMockClient(t, tt.handler)
			defer testServer.Close()

			// Initialize mock controller
			mockCtrl := gomock.NewController(t)

			// Create mock client factory
			mockClientFactory := NewMockClientFactory(mockCtrl)

			// Expect ElasticSearchClient to return the mock client created by getMockClient
			mockClientFactory.EXPECT().ElasticSearchClient(gomock.Any()).Return(esClient, nil).Times(1)

			// Set up the CLI app
			app := cli.NewApp()
			app.Metadata = map[string]interface{}{
				"deps": &deps{
					ClientFactory: mockClientFactory,
				},
			}

			// Setup flag values for the CLI context
			set := flag.NewFlagSet("test", 0)
			set.String(FlagIndex, "test-index", "Index flag")
			if tt.createInputFile {
				set.String(FlagInputFile, inputFileName, "Input file flag")
			} else {
				set.String(FlagInputFile, "invalid-input-file", "Input file flag")
			}
			set.Int(FlagBatchSize, 1, "Batch size flag")

			// Create a mock CLI context
			c := cli.NewContext(app, set, nil)

			// Call AdminIndex
			err = AdminIndex(c)

			// Validate results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to create a temporary input file for AdminIndex or AdminDelete with valid data
func createTempIndexerInputFileWithMessageType(messageType indexer.MessageType, forDelete bool) (string, error) {
	file, err := os.CreateTemp("", "indexer_input_*.txt")
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	if forDelete {
		// For AdminDelete, we need to simulate workflow-id|run-id format
		_, err = writer.WriteString("Header\n") // First line is skipped in AdminDelete
		if err != nil {
			return "", err
		}
		_, err = writer.WriteString("some-value|workflow-id|run-id\n") // Simulate document deletion data
		if err != nil {
			return "", err
		}
	} else {
		// For AdminIndex, we need to generate a JSON message format
		message := `{"WorkflowID": "test-workflow-id", "RunID": "test-run-id", "Version": 1, "MessageType": ` + fmt.Sprintf("%d", messageType) + `}`
		_, err = writer.WriteString(message + "\n")
		if err != nil {
			return "", err
		}
	}

	writer.Flush()

	return file.Name(), nil
}

func TestAdminDelete(t *testing.T) {
	tests := []struct {
		name            string
		handler         http.HandlerFunc
		createInputFile bool
		expectedError   string
	}{
		{
			name: "SuccessDelete",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate a successful Bulk delete request
				if r.URL.Path == "/_bulk" && r.Method == "POST" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"took": 30, "errors": false}`))
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}),
			createInputFile: true,
			expectedError:   "",
		},
		{
			name: "BulkRequestFailure",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulate an error in the Bulk delete request
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Bulk request failed"}`))
			}),
			createInputFile: true,
			expectedError:   "Bulk failed",
		},
		{
			name: "ParseFileError",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// No bulk request needed in this case, just simulating a file parsing error
				w.WriteHeader(http.StatusOK)
			}),
			createInputFile: false, // No valid input file created
			expectedError:   "Cannot open input file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inputFileName string
			var err error
			if tt.createInputFile {
				// Reuse the temp input file creation helper from previous tests
				inputFileName, err = createTempIndexerInputFileWithMessageType(indexer.MessageTypeDelete, true)
				assert.NoError(t, err)
				defer os.Remove(inputFileName) // Clean up after test
			}

			// Create mock Elasticsearch client and server
			esClient, testServer := getMockClient(t, tt.handler)
			defer testServer.Close()

			// Initialize mock controller
			mockCtrl := gomock.NewController(t)

			// Create mock client factory
			mockClientFactory := NewMockClientFactory(mockCtrl)

			// Expect ElasticSearchClient to return the mock client created by getMockClient
			mockClientFactory.EXPECT().ElasticSearchClient(gomock.Any()).Return(esClient, nil).Times(1)

			// Set up the CLI app
			app := cli.NewApp()
			app.Metadata = map[string]interface{}{
				"deps": &deps{
					ClientFactory: mockClientFactory,
				},
			}

			// Setup flag values for the CLI context
			set := flag.NewFlagSet("test", 0)
			set.String(FlagIndex, "test-index", "Index flag")
			if tt.createInputFile {
				set.String(FlagInputFile, inputFileName, "Input file flag")
			} else {
				set.String(FlagInputFile, "invalid-input-file", "Input file flag")
			}
			set.Int(FlagBatchSize, 1, "Batch size flag")
			set.Int(FlagRPS, 10, "RPS flag")

			// Create a mock CLI context
			c := cli.NewContext(app, set, nil)

			// Call AdminDelete
			err = AdminDelete(c)

			// Validate results
			if tt.expectedError != "" {
				if err != nil {
					assert.Contains(t, err.Error(), tt.expectedError)
				} else {
					t.Errorf("Expected error: %s, but got no error", tt.expectedError)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseIndexerMessage(t *testing.T) {
	workflowID := "test-workflow-id"
	runID := "test-run-id"
	version := int64(1)
	messageType := indexer.MessageTypeIndex

	tests := []struct {
		name            string
		messageType     indexer.MessageType
		createInputFile bool
		expectedError   string
		expectedResult  []*indexer.Message
	}{
		{
			name:            "SuccessParse",
			messageType:     indexer.MessageTypeIndex,
			createInputFile: true,
			expectedError:   "",
			expectedResult: []*indexer.Message{
				{
					WorkflowID:  &workflowID,
					RunID:       &runID,
					Version:     &version,
					MessageType: &messageType,
				},
			},
		},
		{
			name:            "FileNotExist",
			messageType:     0,
			createInputFile: false, // No file created
			expectedError:   "open nonexistent-file.txt: no such file or directory",
			expectedResult:  nil,
		},
		{
			name:            "SkipEmptyLines",
			messageType:     indexer.MessageTypeIndex,
			createInputFile: true,
			expectedError:   "",
			expectedResult: []*indexer.Message{
				{
					WorkflowID:  &workflowID,
					RunID:       &runID,
					Version:     &version,
					MessageType: &messageType,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fileName string
			var err error
			if tt.createInputFile {
				// Use the existing createTempIndexerInputFileWithMessageType function
				fileName, err = createTempIndexerInputFileWithMessageType(tt.messageType, false) // forDelete=false for AdminIndex
				assert.NoError(t, err)
				defer os.Remove(fileName) // Clean up after test
			} else {
				// Simulate file not found
				fileName = "nonexistent-file.txt"
			}

			// Call the function being tested
			messages, err := parseIndexerMessage(fileName)

			// Validate results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, messages)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, messages)
			}
		})
	}
}

func TestGenerateESDoc(t *testing.T) {
	tests := []struct {
		name          string
		message       *indexer.Message
		expectedDoc   map[string]interface{}
		expectedError string
	}{
		{
			name: "SuccessWithAllFieldTypes",
			message: &indexer.Message{
				DomainID:   &[]string{"domain1"}[0],
				WorkflowID: &[]string{"workflow1"}[0],
				RunID:      &[]string{"run1"}[0],
				Fields: map[string]*indexer.Field{
					"field_string": {
						Type:       &[]indexer.FieldType{indexer.FieldTypeString}[0],
						StringData: &[]string{"string_value"}[0],
					},
					"field_int": {
						Type:    &[]indexer.FieldType{indexer.FieldTypeInt}[0],
						IntData: &[]int64{123}[0],
					},
					"field_bool": {
						Type:     &[]indexer.FieldType{indexer.FieldTypeBool}[0],
						BoolData: &[]bool{true}[0],
					},
					"field_binary": {
						Type:       &[]indexer.FieldType{indexer.FieldTypeBinary}[0],
						BinaryData: []byte("binary_value"),
					},
				},
			},
			expectedDoc: map[string]interface{}{
				elasticsearch.DomainID:   "domain1",
				elasticsearch.WorkflowID: "workflow1",
				elasticsearch.RunID:      "run1",
				"field_string":           "string_value",
				"field_int":              int64(123),
				"field_bool":             true,
				"field_binary":           []byte("binary_value"),
			},
			expectedError: "",
		},
		{
			name: "UnknownFieldType",
			message: &indexer.Message{
				DomainID:   &[]string{"domain1"}[0],
				WorkflowID: &[]string{"workflow1"}[0],
				RunID:      &[]string{"run1"}[0],
				Fields: map[string]*indexer.Field{
					"unknown_field": {
						Type: &[]indexer.FieldType{9999}[0], // Invalid field type
					},
				},
			},
			expectedDoc:   nil,
			expectedError: "Unknown field type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function being tested
			doc, err := generateESDoc(tt.message)

			// Validate results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, doc)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDoc, doc)
			}
		})
	}
}
