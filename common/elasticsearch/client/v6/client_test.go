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

package v6

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/olivere/elastic"
	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log/testlogger"
)

func TestNewV6Client(t *testing.T) {
	logger := testlogger.New(t)
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()
	url, err := url.Parse(testServer.URL)
	if err != nil {
		t.Fatalf("Failed to parse bad URL: %v", err)
	}

	connectionConfig := &config.ElasticSearchConfig{
		URL:                *url,
		DisableSniff:       true,
		DisableHealthCheck: true,
	}
	sharedClient := testServer.Client()
	client, err := NewV6Client(connectionConfig, logger, sharedClient, sharedClient)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// failed case due to an unreachable Elasticsearch server
	badURL, err := url.Parse("http://nonexistent.elasticsearch.server:9200")
	if err != nil {
		t.Fatalf("Failed to parse bad URL: %v", err)
	}
	connectionConfig.DisableHealthCheck = false
	connectionConfig.URL = *badURL
	_, err = NewV6Client(connectionConfig, logger, nil, nil)
	assert.Error(t, err)
}

func TestCreateIndex(t *testing.T) {
	testServer := getTestServer(t)
	defer testServer.Close()
	// Create a new MockESClient
	mockClient, err := elastic.NewClient(
		elastic.SetURL(testServer.URL),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetHttpClient(testServer.Client()),
	)
	assert.NoError(t, err)

	elasticV6 := ElasticV6{
		client: mockClient,
	}
	err = elasticV6.CreateIndex(context.Background(), "testIndex")
	assert.NoError(t, err)
}

func TestPutMapping(t *testing.T) {
	testServer := getTestServer(t)
	defer testServer.Close()
	// Create a new MockESClient
	mockClient, err := elastic.NewClient(
		elastic.SetURL(testServer.URL),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetHttpClient(testServer.Client()))
	assert.NoError(t, err)

	elasticV6 := ElasticV6{
		client: mockClient,
	}
	err = elasticV6.PutMapping(context.Background(), "testIndex", `{
        "properties": {
            "title": {
                "type": "text"
            },
            "publish_date": {
                "type": "date"
            }
        }
    }`)
	assert.NoError(t, err)
}

func TestCount(t *testing.T) {
	testServer := getTestServer(t)
	defer testServer.Close()
	// Create a new MockESClient
	mockClient, err := elastic.NewClient(
		elastic.SetURL(testServer.URL),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetHttpClient(testServer.Client()))
	assert.NoError(t, err)

	elasticV6 := ElasticV6{
		client: mockClient,
	}
	count, err := elasticV6.Count(context.Background(), "testIndex", `{"query":{"match":{"WorkflowID":"test-workflow-id"}}}`)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), count)
}

func getTestServer(t *testing.T) *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		defer r.Body.Close()

		switch r.URL.Path {
		case "/testIndex":
			if r.Method == "PUT" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"acknowledged": true}`))
			}
		case "/testIndex/_mapping/_doc":
			if r.Method == "PUT" {
				var receivedMapping map[string]interface{}
				if err := json.Unmarshal(body, &receivedMapping); err != nil {
					t.Fatalf("Failed to unmarshal request body: %v", err)
				}

				// Define expected mapping structurally
				expectedMapping := map[string]interface{}{
					"properties": map[string]interface{}{
						"title": map[string]interface{}{
							"type": "text",
						},
						"publish_date": map[string]interface{}{
							"type": "date",
						},
					},
				}

				// Compare structurally
				if !assert.Equal(t, expectedMapping, receivedMapping) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"acknowledged": true}`))
			}
		case "/testIndex/_count":
			expectedQuery := `{"query":{"match":{"WorkflowID":"test-workflow-id"}}}`
			if string(body) != expectedQuery {
				t.Fatalf("Expected query %s, got %s", expectedQuery, body)
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"count": 42}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}
