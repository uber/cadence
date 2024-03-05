package v6

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/olivere/elastic"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log/testlogger"
)

func TestNewV6Client(t *testing.T) {
	logger := testlogger.New(t)
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()
	url, _ := url.Parse(testServer.URL)
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
	badURL, _ := url.Parse("http://nonexistent.elasticsearch.server:9200")
	connectionConfig.DisableHealthCheck = false
	connectionConfig.URL = *badURL
	_, err = NewV6Client(connectionConfig, logger, nil, nil)
	assert.Error(t, err)
}

func TestCreateIndex(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"acknowledged": true}`))
	}))
	defer testServer.Close()
	// Create a new MockESClient
	mockClient, err := elastic.NewClient(
		elastic.SetURL(testServer.URL),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetHttpClient(testServer.Client()),
	)
	assert.NoError(t, err)

	// Create an instance of ElasticV6 with the mock ES client
	elasticV6 := ElasticV6{
		client: mockClient,
	}

	// Call the method under test
	err = elasticV6.CreateIndex(context.Background(), "testindex")
	assert.NoError(t, err)
}

func TestPutMapping(t *testing.T) {
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"acknowledged": true}`))
	}))
	defer testServer.Close()
	// Create a new MockESClient
	mockClient, err := elastic.NewClient(
		elastic.SetURL(testServer.URL),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetHttpClient(testServer.Client()))
	assert.NoError(t, err)

	// Create an instance of ElasticV6 with the mock ES client
	elasticV6 := ElasticV6{
		client: mockClient,
	}

	err = elasticV6.PutMapping(context.Background(), "testindex", `{
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
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/testIndex/_count" {
			// Simulate a response with a count of 42 documents
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"count": 42}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer testServer.Close()
	// Create a new MockESClient
	mockClient, err := elastic.NewClient(
		elastic.SetURL(testServer.URL),
		elastic.SetSniff(false),
		elastic.SetHealthcheck(false),
		elastic.SetHttpClient(testServer.Client()))
	assert.NoError(t, err)

	// Create an instance of ElasticV6 with the mock ES client
	elasticV6 := ElasticV6{
		client: mockClient,
	}

	count, err := elasticV6.Count(context.Background(), "testIndex", "")
	assert.NoError(t, err)
	assert.Equal(t, int64(42), count)
}
