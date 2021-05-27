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

package tests

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/environment"
	"github.com/uber/cadence/tools/cassandra"
)

type (
	VersionTestSuite struct {
		*require.Assertions // override suite.Suite.Assertions with require.Assertions; this means that s.NotNil(nil) will stop the test, not merely log an error
		suite.Suite
	}
)

func TestVersionTestSuite(t *testing.T) {
	suite.Run(t, new(VersionTestSuite))
}

func (s *VersionTestSuite) SetupTest() {
	s.Assertions = require.New(s.T()) // Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
}

func (s *VersionTestSuite) TestVerifyCompatibleVersion() {
	keyspace := "cadence_test"
	visKeyspace := "cadence_visibility_test"
	cqlFile := rootRelativePath + "schema/cassandra/cadence/schema.cql"
	visCqlFile := rootRelativePath + "schema/cassandra/visibility/schema.cql"

	defer s.createKeyspace(keyspace)()
	defer s.createKeyspace(visKeyspace)()
	s.Nil(cassandra.RunTool([]string{
		"./tool", "-k", keyspace, "-q", "setup-schema", "-f", cqlFile, "-version", "10.0", "-o",
	}))
	s.Nil(cassandra.RunTool([]string{
		"./tool", "-k", visKeyspace, "-q", "setup-schema", "-f", visCqlFile, "-version", "10.0", "-o",
	}))

	defaultCfg := config.Cassandra{
		Hosts:    environment.GetCassandraAddress(),
		Port:     cassandra.DefaultCassandraPort,
		User:     "",
		Password: "",
		Keyspace: keyspace,
	}
	visibilityCfg := defaultCfg
	visibilityCfg.Keyspace = visKeyspace
	cfg := config.Persistence{
		DefaultStore:    "default",
		VisibilityStore: "visibility",
		DataStores: map[string]config.DataStore{
			"default":    {NoSQL: &defaultCfg},
			"visibility": {NoSQL: &visibilityCfg},
		},
		TransactionSizeLimit: dynamicconfig.GetIntPropertyFn(common.DefaultTransactionSizeLimit),
		ErrorInjectionRate:   dynamicconfig.GetFloatPropertyFn(0),
	}
	s.NoError(cassandra.VerifyCompatibleVersion(cfg))
}

func (s *VersionTestSuite) TestCheckCompatibleVersion() {
	flags := []struct {
		expectedVersion string
		actualVersion   string
		errStr          string
		expectedFail    bool
	}{
		{"2.0", "1.0", "version mismatch", false},
		{"1.0", "1.0", "", false},
		{"1.0", "2.0", "", false},
		{"1.0", "abc", "unable to read cassandra schema version", false},
	}
	for _, flag := range flags {
		s.runCheckCompatibleVersion(flag.expectedVersion, flag.actualVersion, flag.errStr, flag.expectedFail)
	}
}

func (s *VersionTestSuite) createKeyspace(keyspace string) func() {
	cfg := &cassandra.CQLClientConfig{
		Hosts:       environment.GetCassandraAddress(),
		Port:        cassandra.DefaultCassandraPort,
		Keyspace:    "system",
		Timeout:     cassandra.DefaultTimeout,
		NumReplicas: 1,
	}
	client, err := cassandra.NewCQLClient(cfg)
	s.NoError(err)

	err = client.CreateKeyspace(keyspace)
	if err != nil {
		log.Fatalf("error creating Keyspace, err=%v", err)
	}
	return func() {
		s.NoError(client.DropKeyspace(keyspace))
		client.Close()
	}
}

func (s *VersionTestSuite) runCheckCompatibleVersion(
	expected string, actual string, errStr string, expectedFail bool,
) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	keyspace := fmt.Sprintf("version_test_%v", r.Int63())
	defer s.createKeyspace(keyspace)()

	dir := "check_version"
	tmpDir, err := ioutil.TempDir("", dir)
	s.NoError(err)
	defer os.RemoveAll(tmpDir)

	subdir := tmpDir + "/" + keyspace
	s.NoError(os.Mkdir(subdir, os.FileMode(0744)))

	s.createSchemaForVersion(subdir, actual)
	if expected != actual {
		s.createSchemaForVersion(subdir, expected)
	}

	cqlFile := subdir + "/v" + actual + "/tmp.cql"
	s.NoError(cassandra.RunTool([]string{
		"./tool", "-k", keyspace, "-q", "setup-schema", "-f", cqlFile, "-version", actual, "-o",
	}))
	if expectedFail {
		os.RemoveAll(subdir + "/v" + actual)
	}

	cfg := config.Cassandra{
		Hosts:    environment.GetCassandraAddress(),
		Port:     cassandra.DefaultCassandraPort,
		User:     "",
		Password: "",
		Keyspace: keyspace,
	}
	err = cassandra.CheckCompatibleVersion(cfg, expected)
	if len(errStr) > 0 {
		s.Error(err)
		s.Contains(err.Error(), errStr)
	} else {
		s.NoError(err)
	}
}

func (s *VersionTestSuite) createSchemaForVersion(subdir string, v string) {
	vDir := subdir + "/v" + v
	s.NoError(os.Mkdir(vDir, os.FileMode(0744)))
	cqlFile := vDir + "/tmp.cql"
	s.NoError(ioutil.WriteFile(cqlFile, []byte{}, os.FileMode(0644)))
}
