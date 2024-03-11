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
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/opensearch-project/opensearch-go/v2"
	osapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log/testlogger"
)

func TestNewClient(t *testing.T) {
	logger := testlogger.New(t)
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{ "status": "green" }`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer testServer.Close()
	url, err := url.Parse(testServer.URL)
	if err != nil {
		t.Fatalf("Failed to parse bad URL: %v", err)
	}
	badURL, err := url.Parse("http://nonexistent.elasticsearch.server:9200")
	if err != nil {
		t.Fatalf("Failed to parse bad URL: %v", err)
	}
	tests := []struct {
		name        string
		config      *config.ElasticSearchConfig
		handlerFunc http.HandlerFunc
		expectedErr bool
	}{
		{
			name: "without aws signing config",
			config: &config.ElasticSearchConfig{
				URL:          *url,
				DisableSniff: false,
			},
			expectedErr: false,
		},
		{
			name: "with wrong aws signing config",
			config: &config.ElasticSearchConfig{
				URL:          *badURL,
				DisableSniff: false,
				AWSSigning: config.AWSSigning{
					Enable: true,
				},
			},
			expectedErr: true, //will fail to ping os sever
		},
		{
			name: "with aws signing config",
			config: &config.ElasticSearchConfig{
				URL:          *url,
				DisableSniff: false,
				AWSSigning: config.AWSSigning{
					Enable: true,
					EnvironmentCredential: &config.AWSEnvironmentCredential{
						Region: "us-west-2",
					},
				},
			},
			expectedErr: true, //will fail to ping os sever
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config, logger, testServer.Client())

			if !tt.expectedErr {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestCreateIndex(t *testing.T) {
	tests := []struct {
		name      string
		handler   http.HandlerFunc
		expectErr bool
		secure    bool
	}{
		{
			name: "normal case",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "PUT" && r.URL.Path == "/test-index" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"acknowledged": true, "shards_acknowledged": true, "index": "test-index"}`))
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}),
			expectErr: false,
			secure:    true,
		},
		{
			name: "error case",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.NotFound(w, r)
			}),
			expectErr: true,
			secure:    true,
		},
		{
			name: "not valid config",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.NotFound(w, r)
			}),
			expectErr: true,
			secure:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os2Client, testServer := getSecureMockOS2Client(t, tt.handler, tt.secure)
			defer testServer.Close()

			err := os2Client.CreateIndex(context.Background(), "test-index")
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOSError(t *testing.T) {
	tests := []struct {
		name          string
		givenError    osError
		expectedError string
	}{
		{
			name: "document missing error",
			givenError: osError{
				Status: 404,
				Details: &errorDetails{
					Type:   "document_missing_exception",
					Reason: "document missing [doc-id]",
				},
			},
			expectedError: "Status code: 404, Type: document_missing_exception, Reason: document missing [doc-id]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualError := tt.givenError.Error()
			assert.Equal(t, tt.expectedError, actualError, "The formatted error message did not match the expected value")
		})
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:           "Invalid decoder",
			responseBody:   `{"error": "index_not_found_exception"}`,
			expectError:    true,
			expectedErrMsg: "index_not_found_exception",
		},
		{
			name: "valid error response",
			responseBody: `{
				"status": 404,
				"error": {
					"type": "index_not_found_exception",
					"reason": "index_not_found_exception: no such index",
					"index": "test-index"
				}
			}`,
			expectError:    false,
			expectedErrMsg: "index_not_found_exception: no such index",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := osapi.Response{
				StatusCode: 400,
				Body:       io.NopCloser(bytes.NewBufferString(tt.responseBody)),
			}

			os2Client, testServer := getSecureMockOS2Client(t, nil, false)
			defer testServer.Close()

			err := os2Client.parseError(&response)

			if !tt.expectError {
				if parsedErr, ok := err.(*osError); ok && parsedErr.Details != nil {
					assert.Equal(t, tt.expectedErrMsg, parsedErr.Details.Reason, "Error message mismatch for case: %s", tt.name)
				} else {
					t.Errorf("Failed to assert error reason for case: %s", tt.name)
				}
			} else {
				assert.Error(t, err, "Expected an error for case: %s", tt.name)
			}
		})
	}
}

func getSecureMockOS2Client(t *testing.T, handler http.HandlerFunc, secure bool) (*OS2, *httptest.Server) {
	testServer := httptest.NewTLSServer(handler)
	osConfig := opensearch.Config{
		Addresses: []string{testServer.URL},
	}

	if secure {
		osConfig.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	client, err := opensearch.NewClient(osConfig)
	if err != nil {
		t.Fatalf("Failed to create open search client: %v", err)
	}
	mockClient := &OS2{
		client:  client,
		logger:  testlogger.New(t),
		decoder: &NumberDecoder{},
	}
	assert.NoError(t, err)
	return mockClient, testServer
}

func TestCloseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test response body"))
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request to test server: %v", err)
	}

	osResponse := &osapi.Response{
		StatusCode: resp.StatusCode,
		Body:       resp.Body,
		Header:     resp.Header,
	}

	// Assert that the response body is open before calling closeBody
	_, err = osResponse.Body.Read(make([]byte, 1))
	assert.NoError(t, err, "Expected response body to be open before calling closeBody")

	closeBody(osResponse)

	// Attempt to read from the body again should result in an error because it's closed
	_, err = osResponse.Body.Read(make([]byte, 1))
	assert.Error(t, err, "Expected response body to be closed after calling closeBody")
}
