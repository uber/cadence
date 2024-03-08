package os2

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log/testlogger"
)

func TestNewClient(t *testing.T) {
	logger := testlogger.New(t)
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			// You might need to return a valid OpenSearch ping response JSON here
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
			sharedClient := testServer.Client()
			client, err := NewClient(tt.config, logger, sharedClient)

			if !tt.expectedErr {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
