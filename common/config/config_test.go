// Copyright (c) 2019 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package config

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/service"
)

func TestToString(t *testing.T) {
	var cfg Config
	err := Load("", "../../config", "", &cfg)
	assert.NoError(t, err)
	assert.NotEmpty(t, cfg.String())
}

func TestFillingDefaultSQLEncodingDecodingTypes(t *testing.T) {
	cfg := &Config{
		Persistence: Persistence{
			DataStores: map[string]DataStore{
				"sql": {
					SQL: &SQL{},
				},
			},
		},
		ClusterGroupMetadata: &ClusterGroupMetadata{},
	}
	cfg.fillDefaults()
	assert.Equal(t, string(common.EncodingTypeThriftRW), cfg.Persistence.DataStores["sql"].SQL.EncodingType)
	assert.Equal(t, []string{string(common.EncodingTypeThriftRW)}, cfg.Persistence.DataStores["sql"].SQL.DecodingTypes)
}

func getValidMultipleDatabasseConfig() *Config {
	metadata := validClusterGroupMetadata()
	cfg := &Config{
		ClusterGroupMetadata: metadata,
		Persistence: Persistence{
			DefaultStore:            "default",
			AdvancedVisibilityStore: "esv7",
			DataStores: map[string]DataStore{
				"default": {
					SQL: &SQL{
						PluginName:           "fake",
						ConnectProtocol:      "tcp",
						NumShards:            2,
						UseMultipleDatabases: true,
						MultipleDatabasesConfig: []MultipleDatabasesConfigEntry{
							{
								DatabaseName: "db1",
								ConnectAddr:  "192.168.0.1:3306",
							},
							{
								DatabaseName: "db2",
								ConnectAddr:  "192.168.0.2:3306",
							},
						},
					},
				},
				"esv7": {
					ElasticSearch: &ElasticSearchConfig{
						Version: "v7",
						URL: url.URL{Scheme: "http",
							Host: "127.0.0.1:9200",
						},
						Indices: map[string]string{
							"visibility": "cadence-visibility-dev",
						},
					},
					// no sql or nosql, should be populated from cassandra
				},
			},
		},
	}
	return cfg
}

func TestValidMultipleDatabaseConfig(t *testing.T) {
	cfg := getValidMultipleDatabasseConfig()
	err := cfg.ValidateAndFillDefaults()
	require.NoError(t, err)
}

func TestInvalidMultipleDatabaseConfig_useBasicVisibility(t *testing.T) {
	cfg := getValidMultipleDatabasseConfig()
	cfg.Persistence.VisibilityStore = "basic"
	cfg.Persistence.DataStores["basic"] = DataStore{
		SQL: &SQL{},
	}
	err := cfg.ValidateAndFillDefaults()
	require.EqualError(t, err, "sql persistence config: multipleSQLDatabases can only be used with advanced visibility only")
}

func TestInvalidMultipleDatabaseConfig_wrongNumDBShards(t *testing.T) {
	cfg := getValidMultipleDatabasseConfig()
	sqlds := cfg.Persistence.DataStores["default"]
	sqlds.SQL.NumShards = 3
	cfg.Persistence.DataStores["default"] = sqlds
	err := cfg.ValidateAndFillDefaults()
	require.EqualError(t, err, "sql persistence config: nShards must be greater than one and equal to the length of multipleDatabasesConfig")
}

func TestInvalidMultipleDatabaseConfig_nonEmptySQLUser(t *testing.T) {
	cfg := getValidMultipleDatabasseConfig()
	sqlds := cfg.Persistence.DataStores["default"]
	sqlds.SQL.User = "user"
	cfg.Persistence.DataStores["default"] = sqlds
	err := cfg.ValidateAndFillDefaults()
	require.EqualError(t, err, "sql persistence config: user can only be configured in multipleDatabasesConfig when UseMultipleDatabases is true")
}

func TestInvalidMultipleDatabaseConfig_nonEmptySQLPassword(t *testing.T) {
	cfg := getValidMultipleDatabasseConfig()
	sqlds := cfg.Persistence.DataStores["default"]
	sqlds.SQL.Password = "pw"
	cfg.Persistence.DataStores["default"] = sqlds
	err := cfg.ValidateAndFillDefaults()
	require.EqualError(t, err, "sql persistence config: password can only be configured in multipleDatabasesConfig when UseMultipleDatabases is true")
}

func TestInvalidMultipleDatabaseConfig_nonEmptySQLDatabaseName(t *testing.T) {
	cfg := getValidMultipleDatabasseConfig()
	sqlds := cfg.Persistence.DataStores["default"]
	sqlds.SQL.DatabaseName = "db"
	cfg.Persistence.DataStores["default"] = sqlds
	err := cfg.ValidateAndFillDefaults()
	require.EqualError(t, err, "sql persistence config: databaseName can only be configured in multipleDatabasesConfig when UseMultipleDatabases is true")
}

func TestInvalidMultipleDatabaseConfig_nonEmptySQLConnAddr(t *testing.T) {
	cfg := getValidMultipleDatabasseConfig()
	sqlds := cfg.Persistence.DataStores["default"]
	sqlds.SQL.ConnectAddr = "127.0.0.1:3306"
	cfg.Persistence.DataStores["default"] = sqlds
	err := cfg.ValidateAndFillDefaults()
	require.EqualError(t, err, "sql persistence config: connectAddr can only be configured in multipleDatabasesConfig when UseMultipleDatabases is true")
}

func TestConfigFallbacks(t *testing.T) {
	metadata := validClusterGroupMetadata()
	cfg := &Config{
		Services: map[string]Service{
			"frontend": {
				RPC: RPC{
					Port: 7900,
				},
			},
		},
		ClusterGroupMetadata: metadata,
		Persistence: Persistence{
			DefaultStore:    "default",
			VisibilityStore: "cass",
			DataStores: map[string]DataStore{
				"default": {
					SQL: &SQL{
						PluginName:      "fake",
						ConnectProtocol: "tcp",
						ConnectAddr:     "192.168.0.1:3306",
						DatabaseName:    "db1",
						NumShards:       0, // default value, should be changed
					},
				},
				"cass": {
					Cassandra: &NoSQL{
						Hosts: "127.0.0.1",
					},
					// no sql or nosql, should be populated from cassandra
				},
			},
		},
	}
	err := cfg.ValidateAndFillDefaults()
	require.NoError(t, err) // sanity check, must be valid or the later tests are potentially useless

	assert.NotEmpty(t, cfg.Persistence.DataStores["cass"].Cassandra, "cassandra config should remain after update")
	assert.NotEmpty(t, cfg.Persistence.DataStores["cass"].NoSQL, "nosql config should contain cassandra config / not be empty")
	assert.NotZero(t, cfg.Persistence.DataStores["default"].SQL.NumShards, "num shards should be nonzero")
	assert.Equal(t, "localhost:7833", cfg.PublicClient.HostPort)
}

func TestConfigErrorInAuthorizationConfig(t *testing.T) {
	cfg := &Config{
		Authorization: Authorization{
			OAuthAuthorizer: OAuthAuthorizer{
				Enable: true,
			},
			NoopAuthorizer: NoopAuthorizer{
				Enable: true,
			},
		},
		ClusterGroupMetadata: &ClusterGroupMetadata{},
	}

	err := cfg.ValidateAndFillDefaults()
	require.Error(t, err)
}

func TestGetServiceConfig(t *testing.T) {
	cfg := Config{}
	_, err := cfg.GetServiceConfig(service.Frontend)
	assert.EqualError(t, err, "no config section for service: frontend")

	cfg = Config{Services: map[string]Service{"frontend": {RPC: RPC{GRPCPort: 123}}}}
	svc, err := cfg.GetServiceConfig(service.Frontend)
	assert.NoError(t, err)
	assert.NotEmpty(t, svc)
}
