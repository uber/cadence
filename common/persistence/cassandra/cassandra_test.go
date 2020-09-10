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

package cassandra_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common/persistence/cassandra"
	pt "github.com/uber/cadence/common/persistence/persistence-tests"
)

// NewTestBaseWithCassandra returns a persistence test base backed by cassandra datastore
func NewTestBaseWithCassandra(options *pt.TestBaseOptions) pt.TestBase {
	if options.DBName == "" {
		options.DBName = "test_" + pt.GenerateRandomDBName(10)
	}
	testCluster := cassandra.NewTestCluster(
		options.DBName,
		options.DBUsername,
		options.DBPassword,
		options.DBHost,
		options.DBPort,
		options.SchemaDir,
	)

	return pt.NewTestBase(options, testCluster)
}

func TestCassandraHistoryV2Persistence(t *testing.T) {
	s := new(pt.HistoryV2PersistenceSuite)
	s.TestBase = NewTestBaseWithCassandra(&pt.TestBaseOptions{})
	s.TestBase.Setup()
	suite.Run(t, s)
}

func TestCassandraMatchingPersistence(t *testing.T) {
	s := new(pt.MatchingPersistenceSuite)
	s.TestBase = NewTestBaseWithCassandra(&pt.TestBaseOptions{})
	s.TestBase.Setup()
	suite.Run(t, s)
}

func TestCassandraMetadataPersistenceV2(t *testing.T) {
	s := new(pt.MetadataPersistenceSuiteV2)
	s.TestBase = NewTestBaseWithCassandra(&pt.TestBaseOptions{})
	s.TestBase.Setup()
	suite.Run(t, s)
}

func TestCassandraShardPersistence(t *testing.T) {
	s := new(pt.ShardPersistenceSuite)
	s.TestBase = NewTestBaseWithCassandra(&pt.TestBaseOptions{})
	s.TestBase.Setup()
	suite.Run(t, s)
}

func TestCassandraVisibilityPersistence(t *testing.T) {
	s := new(pt.VisibilityPersistenceSuite)
	s.TestBase = NewTestBaseWithCassandra(&pt.TestBaseOptions{})
	s.TestBase.Setup()
	suite.Run(t, s)
}

func TestCassandraExecutionManager(t *testing.T) {
	s := new(pt.ExecutionManagerSuite)
	s.TestBase = NewTestBaseWithCassandra(&pt.TestBaseOptions{})
	s.TestBase.Setup()
	suite.Run(t, s)
}

func TestCassandraExecutionManagerWithEventsV2(t *testing.T) {
	s := new(pt.ExecutionManagerSuiteForEventsV2)
	s.TestBase = NewTestBaseWithCassandra(&pt.TestBaseOptions{})
	s.TestBase.Setup()
	suite.Run(t, s)
}

func TestQueuePersistence(t *testing.T) {
	s := new(pt.QueuePersistenceSuite)
	s.TestBase = NewTestBaseWithCassandra(&pt.TestBaseOptions{})
	s.TestBase.Setup()
	suite.Run(t, s)
}
