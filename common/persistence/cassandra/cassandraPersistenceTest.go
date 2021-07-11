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

package cassandra

import (
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
	"github.com/uber/cadence/environment"
)

const (
	testSchemaDir = "schema/cassandra/"
)

// TestCluster allows executing cassandra operations in testing.
type TestCluster struct {
	keyspace  string
	schemaDir string
	cluster   *gocql.ClusterConfig
	session   gocql.Session
	cfg       config.Cassandra
}

// NewTestCluster returns a new cassandra test cluster
func NewTestCluster(keyspace, username, password, host string, port int, schemaDir string, protoVersion int) *TestCluster {
	var result TestCluster
	result.keyspace = keyspace
	if port == 0 {
		port = environment.GetCassandraPort()
	}
	if schemaDir == "" {
		schemaDir = testSchemaDir
	}
	if host == "" {
		host = environment.GetCassandraAddress()
	}
	if protoVersion == 0 {
		protoVersion = environment.GetCassandraProtoVersion()
	}
	result.schemaDir = schemaDir
	result.cfg = config.Cassandra{
		User:         username,
		Password:     password,
		Hosts:        host,
		Port:         port,
		MaxConns:     2,
		Keyspace:     keyspace,
		ProtoVersion: protoVersion,
	}
	return &result
}

// Config returns the persistence config for connecting to this test cluster
func (s *TestCluster) Config() config.Persistence {
	cfg := s.cfg
	return config.Persistence{
		DefaultStore:    "test",
		VisibilityStore: "test",
		DataStores: map[string]config.DataStore{
			"test": {NoSQL: &cfg},
		},
		TransactionSizeLimit: dynamicconfig.GetIntPropertyFn(common.DefaultTransactionSizeLimit),
		ErrorInjectionRate:   dynamicconfig.GetFloatPropertyFn(0),
	}
}

// databaseName from PersistenceTestCluster interface
func (s *TestCluster) databaseName() string {
	return s.keyspace
}

// SetupTestDatabase from PersistenceTestCluster interface
func (s *TestCluster) SetupTestDatabase() {
	s.createSession()
	s.createDatabase()
	schemaDir := s.schemaDir + "/"

	if !strings.HasPrefix(schemaDir, "/") && !strings.HasPrefix(schemaDir, "../") {
		cadencePackageDir, err := getCadencePackageDir()
		if err != nil {
			log.Fatal(err)
		}
		schemaDir = cadencePackageDir + schemaDir
	}

	s.loadSchema([]string{"schema.cql"}, schemaDir)
	s.loadVisibilitySchema([]string{"schema.cql"}, schemaDir)
}

// TearDownTestDatabase from PersistenceTestCluster interface
func (s *TestCluster) TearDownTestDatabase() {
	s.dropDatabase()
	s.session.Close()
}

// createSession from PersistenceTestCluster interface
func (s *TestCluster) createSession() {
	s.cluster = &gocql.ClusterConfig{
		Hosts:        s.cfg.Hosts,
		Port:         s.cfg.Port,
		User:         s.cfg.User,
		Password:     s.cfg.Password,
		ProtoVersion: s.cfg.ProtoVersion,
		Keyspace:     "system",
		Consistency:  gocql.One,
		Timeout:      40 * time.Second,
	}

	var err error
	s.session, err = gocql.GetOrCreateClient().CreateSession(*s.cluster)
	if err != nil {
		log.Fatal(`CreateSession`, err)
	}
}

// createDatabase from PersistenceTestCluster interface
func (s *TestCluster) createDatabase() {
	err := createCassandraKeyspace(s.session, s.databaseName(), 1, true)
	if err != nil {
		log.Fatal(err)
	}

	s.cluster.Keyspace = s.databaseName()
}

// dropDatabase from PersistenceTestCluster interface
func (s *TestCluster) dropDatabase() {
	err := dropCassandraKeyspace(s.session, s.databaseName())
	if err != nil && !strings.Contains(err.Error(), "AlreadyExists") {
		log.Fatal(err)
	}
}

// loadSchema from PersistenceTestCluster interface
func (s *TestCluster) loadSchema(fileNames []string, schemaDir string) {
	workflowSchemaDir := schemaDir + "/cadence"
	err := loadCassandraSchema(workflowSchemaDir, fileNames, s.cluster.Hosts, s.cluster.Port, s.databaseName(), true, nil, s.cluster.ProtoVersion)
	if err != nil && !strings.Contains(err.Error(), "AlreadyExists") {
		log.Fatal(err)
	}
}

// loadVisibilitySchema from PersistenceTestCluster interface
func (s *TestCluster) loadVisibilitySchema(fileNames []string, schemaDir string) {
	workflowSchemaDir := schemaDir + "visibility"
	err := loadCassandraSchema(workflowSchemaDir, fileNames, s.cluster.Hosts, s.cluster.Port, s.databaseName(), false, nil, s.cluster.ProtoVersion)
	if err != nil && !strings.Contains(err.Error(), "AlreadyExists") {
		log.Fatal(err)
	}
}

func getCadencePackageDir() (string, error) {
	cadencePackageDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cadenceIndex := strings.LastIndex(cadencePackageDir, "/cadence/")
	cadencePackageDir = cadencePackageDir[:cadenceIndex+len("/cadence/")]
	if err != nil {
		panic(err)
	}
	return cadencePackageDir, err
}
