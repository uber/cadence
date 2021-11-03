// Copyright (c) 2017 Uber Technologies, Inc.
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
	"fmt"

	"github.com/uber/cadence/common"
)

const (
	// StoreTypeSQL refers to sql based storage as persistence store
	StoreTypeSQL = "sql"
	// StoreTypeCassandra refers to cassandra as persistence store
	StoreTypeCassandra = "cassandra"
)

// DefaultStoreType returns the storeType for the default persistence store
func (c *Persistence) DefaultStoreType() string {
	if c.DataStores[c.DefaultStore].SQL != nil {
		return StoreTypeSQL
	}
	return StoreTypeCassandra
}

// FillDefaults populates default values for unspecified fields in persistence config
func (c *Persistence) FillDefaults() {
	for k, store := range c.DataStores {
		if store.Cassandra != nil && store.NoSQL == nil {
			// for backward-compatibility
			store.NoSQL = store.Cassandra
			store.NoSQL.PluginName = "cassandra"
		}

		if store.SQL != nil {
			// filling default encodingType/decodingTypes for SQL persistence
			if store.SQL.EncodingType == "" {
				store.SQL.EncodingType = string(common.EncodingTypeThriftRW)
			}
			if len(store.SQL.DecodingTypes) == 0 {
				store.SQL.DecodingTypes = []string{
					string(common.EncodingTypeThriftRW),
				}
			}

			if store.SQL.NumShards == 0 {
				store.SQL.NumShards = 1
			}
		}

		// write changes back to DataStores, as ds is a value object
		c.DataStores[k] = store
	}
}

// Validate validates the persistence config
func (c *Persistence) Validate() error {
	dbStoreKeys := []string{c.DefaultStore}

	useAdvancedVisibilityOnly := false
	if _, ok := c.DataStores[c.VisibilityStore]; ok {
		dbStoreKeys = append(dbStoreKeys, c.VisibilityStore)
	} else {
		if _, ok := c.DataStores[c.AdvancedVisibilityStore]; !ok {
			return fmt.Errorf("must provide one of VisibilityStore and AdvancedVisibilityStore")
		}
		useAdvancedVisibilityOnly = true
	}

	for _, st := range dbStoreKeys {
		ds, ok := c.DataStores[st]
		if !ok {
			return fmt.Errorf("persistence config: missing config for datastore %v", st)
		}
		if ds.Cassandra != nil && ds.NoSQL != nil && ds.Cassandra != ds.NoSQL {
			return fmt.Errorf("persistence config: datastore %v: only one of Cassandra or NoSQL can be specified", st)
		}
		if ds.SQL == nil && ds.NoSQL == nil {
			return fmt.Errorf("persistence config: datastore %v: must provide config for one of SQL or NoSQL stores", st)
		}
		if ds.SQL != nil && ds.NoSQL != nil {
			return fmt.Errorf("persistence config: datastore %v: only one of SQL or NoSQL can be specified", st)
		}
		if ds.SQL != nil {
			if ds.SQL.UseMultipleDatabases {
				if !useAdvancedVisibilityOnly {
					return fmt.Errorf("sql persistence config: multipleSQLDatabases can only be used with advanced visibility only")
				}
				if ds.SQL.DatabaseName != "" {
					return fmt.Errorf("sql persistence config: databaseName can only be configured in multipleDatabasesConfig when UseMultipleDatabases is true")
				}
				if ds.SQL.ConnectAddr != "" {
					return fmt.Errorf("sql persistence config: connectAddr can only be configured in multipleDatabasesConfig when UseMultipleDatabases is true")
				}
				if ds.SQL.User != "" {
					return fmt.Errorf("sql persistence config: user can only be configured in multipleDatabasesConfig when UseMultipleDatabases is true")
				}
				if ds.SQL.Password != "" {
					return fmt.Errorf("sql persistence config: password can only be configured in multipleDatabasesConfig when UseMultipleDatabases is true")
				}
				if ds.SQL.NumShards <= 1 || len(ds.SQL.MultipleDatabasesConfig) != ds.SQL.NumShards {
					return fmt.Errorf("sql persistence config: nShards must be greater than one and equal to the length of multipleDatabasesConfig")
				}
				for _, entry := range ds.SQL.MultipleDatabasesConfig {
					if entry.DatabaseName == "" {
						return fmt.Errorf("sql multipleDatabasesConfig persistence config: databaseName can not be empty")
					}
					if entry.ConnectAddr == "" {
						return fmt.Errorf("sql multipleDatabasesConfig persistence config: connectAddr can not be empty")
					}
				}
			} else {
				if ds.SQL.DatabaseName == "" {
					return fmt.Errorf("sql persistence config: databaseName can not be empty")
				}
				if ds.SQL.ConnectAddr == "" {
					return fmt.Errorf("sql persistence config: connectAddr can not be empty")
				}
			}
		}
	}

	return nil
}

// IsAdvancedVisibilityConfigExist returns whether user specified advancedVisibilityStore in config
func (c *Persistence) IsAdvancedVisibilityConfigExist() bool {
	return len(c.AdvancedVisibilityStore) != 0
}
