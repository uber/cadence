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

package nosql

import (
	log "github.com/sirupsen/logrus"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/persistence/persistence-tests/testcluster"
)

// testCluster allows executing cassandra operations in testing.
type testCluster struct {
	keyspace      string
	schemaBaseDir string
	cfg           config.NoSQL
}

var _ testcluster.PersistenceTestCluster = (*testCluster)(nil)

// NewTestCluster returns a new cassandra test cluster
// if schemaBaseDir is empty, it will be auto-resolved based on os.Getwd()
// otherwise the specified value will be used (used by internal tests)
func NewTestCluster(
	pluginName, keyspace, username, password, host string,
	port, protoVersion int,
	schemaBaseDir string,
) testcluster.PersistenceTestCluster {
	return &testCluster{
		keyspace:      keyspace,
		schemaBaseDir: schemaBaseDir,
		cfg: config.NoSQL{
			PluginName:   pluginName,
			User:         username,
			Password:     password,
			Hosts:        host,
			Port:         port,
			MaxConns:     2,
			Keyspace:     keyspace,
			ProtoVersion: protoVersion,
		},
	}
}

// Config returns the persistence config for connecting to this test cluster
func (s *testCluster) Config() config.Persistence {
	return config.Persistence{
		DefaultStore:    "test",
		VisibilityStore: "test",
		DataStores: map[string]config.DataStore{
			"test": {NoSQL: &s.cfg},
		},
		TransactionSizeLimit: dynamicconfig.GetIntPropertyFn(common.DefaultTransactionSizeLimit),
		ErrorInjectionRate:   dynamicconfig.GetFloatPropertyFn(0),
	}
}

// SetupTestDatabase from PersistenceTestCluster interface
func (s *testCluster) SetupTestDatabase() {
	adminDB, err := NewNoSQLAdminDB(&s.cfg, loggerimpl.NewNopLogger())

	if err != nil {
		log.Fatal(err)
	}
	err = adminDB.SetupTestDatabase(s.schemaBaseDir)
	if err != nil {
		log.Fatal(err)
	}
}

// TearDownTestDatabase from PersistenceTestCluster interface
func (s *testCluster) TearDownTestDatabase() {
	adminDB, err := NewNoSQLAdminDB(&s.cfg, loggerimpl.NewNopLogger())
	if err != nil {
		log.Fatal(err)
	}
	err = adminDB.TeardownTestDatabase()
	if err != nil {
		log.Fatal(err)
	}
}
