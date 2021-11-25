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
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/mongodb"
	persistencetests "github.com/uber/cadence/common/persistence/persistence-tests"
	"github.com/uber/cadence/environment"
)

func TestConfigStorePersistence(t *testing.T) {
	s := new(persistencetests.ConfigStorePersistenceSuite)
	s.TestBase = NewTestBaseWithMongo()
	s.TestBase.Setup()
	suite.Run(t, s)
}

// TODO uncomment the test once HistoryEventsCRUD is implemented
// func TestMongoDBHistoryPersistence(t *testing.T) {
// 	s := new(persistencetests.HistoryV2PersistenceSuite)
// 	s.TestBase = NewTestBaseWithMongo()
// 	s.TestBase.Setup()
// 	suite.Run(t, s)
// }

// TODO uncomment the test once TaskCRUD is implemented
// func TestMongoDBMatchingPersistence(t *testing.T) {
// 	s := new(persistencetests.MatchingPersistenceSuite)
// 	s.TestBase = NewTestBaseWithMongo()
// 	s.TestBase.Setup()
// 	suite.Run(t, s)
// }

// TODO uncomment the test once DomainCRUD is implemented
// func TestMongoDBDomainPersistence(t *testing.T) {
// 	s := new(persistencetests.MetadataPersistenceSuiteV2)
// 	s.TestBase = NewTestBaseWithMongo()
// 	s.TestBase.Setup()
// 	suite.Run(t, s)
// }

// TODO uncomment the test once MessageQueueCRUD is implemented
// func TestQueuePersistence(t *testing.T) {
// 	s := new(persistencetests.QueuePersistenceSuite)
// 	s.TestBase = NewTestBaseWithMongo()
// 	s.TestBase.Setup()
// 	suite.Run(t, s)
// }

// TODO uncomment the test once ShardCRUD is implemented
// func TestCassandraShardPersistence(t *testing.T) {
// 	s := new(persistencetests.ShardPersistenceSuite)
// 	s.TestBase = NewTestBaseWithMongo()
// 	s.TestBase.Setup()
// 	suite.Run(t, s)
// }

// TODO uncomment the test once VisibilityCRUD is implemented
// func TestCassandraVisibilityPersistence(t *testing.T) {
// 	s := new(persistencetests.DBVisibilityPersistenceSuite)
// 	s.TestBase = NewTestBaseWithMongo()
// 	s.TestBase.Setup()
// 	suite.Run(t, s)
// }

// TODO uncomment the test once WorkflowCRUD is implemented
// func TestCassandraExecutionManager(t *testing.T) {
// 	s := new(persistencetests.ExecutionManagerSuite)
// 	s.TestBase = NewTestBaseWithMongo()
// 	s.TestBase.Setup()
// 	suite.Run(t, s)
// }

// TODO uncomment the test once WorkflowCRUD is implemented
// func TestCassandraExecutionManagerWithEventsV2(t *testing.T) {
// 	s := new(persistencetests.ExecutionManagerSuiteForEventsV2)
// 	s.TestBase = NewTestBaseWithMongo()
// 	s.TestBase.Setup()
// 	suite.Run(t, s)
// }

func NewTestBaseWithMongo() persistencetests.TestBase {
	options := &persistencetests.TestBaseOptions{
		DBPluginName: mongodb.PluginName,
		DBHost:       getTestConfig().Hosts,
		DBUsername:   getTestConfig().User,
		DBPassword:   getTestConfig().Password,
		DBPort:       getTestConfig().Port,
	}
	return persistencetests.NewTestBaseWithNoSQL(options)
}

func getTestConfig() *config.NoSQL {
	return &config.NoSQL{
		PluginName: mongodb.PluginName,
		User:       "root",
		Password:   "cadence",
		Hosts:      environment.GetMongoAddress(),
		Port:       environment.GetMongoPort(),
	}
}
