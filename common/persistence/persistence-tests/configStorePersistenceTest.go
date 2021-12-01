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

package persistencetests

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/common/config"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/mongodb"
	"github.com/uber/cadence/common/types"
)

var supportedPlugins = map[string]bool{
	cassandra.PluginName: true,
	mongodb.PluginName:   true,
}

// Currently you cannot clear or remove any entries in cluster_config table
// Therefore, Teardown and Setup of Test DB is required before every test.

type (
	// ConfigStorePersistenceSuite contains config store persistence tests
	ConfigStorePersistenceSuite struct {
		TestBase
		// override suite.Suite.Assertions with require.Assertions; this means that s.NotNil(nil) will stop the test,
		// not merely log an error
		*require.Assertions
	}
)

// SetupSuite implementation
func (s *ConfigStorePersistenceSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}
}

// SetupTest implementation
func (s *ConfigStorePersistenceSuite) SetupTest() {
	// Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
	s.Assertions = require.New(s.T())
}

// TearDownSuite implementation
func (s *ConfigStorePersistenceSuite) TearDownSuite() {
	s.TearDownWorkflowStore()
}

// Tests if error is returned when trying to fetch dc values from empty table
func (s *ConfigStorePersistenceSuite) TestFetchFromEmptyTable() {
	if !validDatabaseCheck(s.Config()) {
		s.T().Skip()
	}

	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	s.DefaultTestCluster.TearDownTestDatabase()
	s.DefaultTestCluster.SetupTestDatabase()

	snapshot, err := s.FetchDynamicConfig(ctx)
	s.Nil(snapshot)
	s.NotNil(err)
}

func (s *ConfigStorePersistenceSuite) TestUpdateSimpleSuccess() {
	if !validDatabaseCheck(s.Config()) {
		s.T().Skip()
	}

	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	s.DefaultTestCluster.TearDownTestDatabase()
	s.DefaultTestCluster.SetupTestDatabase()

	snapshot := generateRandomSnapshot(1)
	err := s.UpdateDynamicConfig(ctx, snapshot)
	s.Nil(err)

	retSnapshot, err := s.FetchDynamicConfig(ctx)
	s.NotNil(snapshot)
	s.Nil(err)
	s.Equal(snapshot.Version, retSnapshot.Version)
	s.Equal(snapshot.Values.Entries[0].Name, retSnapshot.Values.Entries[0].Name)
}

func (s *ConfigStorePersistenceSuite) TestUpdateVersionCollisionFailure() {
	if !validDatabaseCheck(s.Config()) {
		s.T().Skip()
	}

	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	s.DefaultTestCluster.TearDownTestDatabase()
	s.DefaultTestCluster.SetupTestDatabase()

	snapshot := generateRandomSnapshot(1)
	err := s.UpdateDynamicConfig(ctx, snapshot)
	s.Nil(err)

	err = s.UpdateDynamicConfig(ctx, snapshot)
	var condErr *p.ConditionFailedError
	s.True(errors.As(err, &condErr))
}

func (s *ConfigStorePersistenceSuite) TestUpdateIncrementalVersionSuccess() {
	if !validDatabaseCheck(s.Config()) {
		s.T().Skip()
	}

	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	s.DefaultTestCluster.TearDownTestDatabase()
	s.DefaultTestCluster.SetupTestDatabase()

	snapshot2 := generateRandomSnapshot(2)
	err := s.UpdateDynamicConfig(ctx, snapshot2)
	s.Nil(err)
	snapshot3 := generateRandomSnapshot(3)
	err = s.UpdateDynamicConfig(ctx, snapshot3)
	s.Nil(err)
}

func (s *ConfigStorePersistenceSuite) TestFetchLatestVersionSuccess() {
	if !validDatabaseCheck(s.Config()) {
		s.T().Skip()
	}

	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	s.DefaultTestCluster.TearDownTestDatabase()
	s.DefaultTestCluster.SetupTestDatabase()

	snapshot2 := generateRandomSnapshot(2)
	err := s.UpdateDynamicConfig(ctx, snapshot2)
	s.Nil(err)
	snapshot3 := generateRandomSnapshot(3)
	err = s.UpdateDynamicConfig(ctx, snapshot3)
	s.Nil(err)

	snapshot, err := s.FetchDynamicConfig(ctx)
	s.NotNil(snapshot)
	s.Nil(err)
	s.Equal(int64(3), snapshot.Version)
}

func generateRandomSnapshot(version int64) *p.DynamicConfigSnapshot {
	data, _ := json.Marshal("test_value")

	values := make([]*types.DynamicConfigValue, 1)
	values[0] = &types.DynamicConfigValue{
		Value: &types.DataBlob{
			EncodingType: types.EncodingTypeJSON.Ptr(),
			Data:         data,
		},
		Filters: nil,
	}

	entries := make([]*types.DynamicConfigEntry, 1)
	entries[0] = &types.DynamicConfigEntry{
		Name:   "test_parameter",
		Values: values,
	}

	return &p.DynamicConfigSnapshot{
		Version: version,
		Values: &types.DynamicConfigBlob{
			SchemaVersion: 1,
			Entries:       entries,
		},
	}
}

func validDatabaseCheck(cfg config.Persistence) bool {
	if datastore, ok := cfg.DataStores[cfg.DefaultStore]; ok {
		if datastore.NoSQL != nil {
			return supportedPlugins[datastore.NoSQL.PluginName]
		}
	}
	return false
}

func (s *ConfigStorePersistenceSuite) FetchDynamicConfig(ctx context.Context) (*p.DynamicConfigSnapshot, error) {
	response, err := s.ConfigStoreManager.FetchDynamicConfig(ctx)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.New("nil FetchDynamicConfig response")
	}
	return response.Snapshot, nil
}

func (s *ConfigStorePersistenceSuite) UpdateDynamicConfig(ctx context.Context, snapshot *p.DynamicConfigSnapshot) error {
	return s.ConfigStoreManager.UpdateDynamicConfig(ctx, &p.UpdateDynamicConfigRequest{Snapshot: snapshot})
}
