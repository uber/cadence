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
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

// Currently you cannot clear or remove any entries in cluster_config table
// Therefore, Teardown and Setup of Test DB is required before every test.

type (
	// ConfigStorePersistenceSuite contains config store persistence tests
	ConfigStorePersistenceSuite struct {
		*TestBase
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
	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	snapshot, err := s.FetchDynamicConfig(ctx, 0)
	s.Nil(snapshot)
	s.NotNil(err)
}

func (s *ConfigStorePersistenceSuite) TestUpdateSimpleSuccess() {
	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	snapshot := generateRandomSnapshot(1)
	err := s.UpdateDynamicConfig(ctx, snapshot, 1)
	s.Nil(err)

	retSnapshot, err := s.FetchDynamicConfig(ctx, 1)
	s.NotNil(retSnapshot)
	s.Nil(err)
	s.Equal(snapshot.Version, retSnapshot.Version)
	s.Equal(snapshot.Values.Entries[0].Name, retSnapshot.Values.Entries[0].Name)
}

func (s *ConfigStorePersistenceSuite) TestUpdateVersionCollisionFailure() {
	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	snapshot := generateRandomSnapshot(1)
	err := s.UpdateDynamicConfig(ctx, snapshot, 2)
	s.Nil(err)

	err = s.UpdateDynamicConfig(ctx, snapshot, 2)
	var condErr *p.ConditionFailedError
	s.True(errors.As(err, &condErr))
}

func (s *ConfigStorePersistenceSuite) TestUpdateIncrementalVersionSuccess() {
	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	snapshot2 := generateRandomSnapshot(2)
	err := s.UpdateDynamicConfig(ctx, snapshot2, 3)
	s.Nil(err)
	snapshot3 := generateRandomSnapshot(3)
	err = s.UpdateDynamicConfig(ctx, snapshot3, 3)
	s.Nil(err)
}

func (s *ConfigStorePersistenceSuite) TestFetchLatestVersionSuccess() {
	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	snapshot2 := generateRandomSnapshot(2)
	err := s.UpdateDynamicConfig(ctx, snapshot2, 4)
	s.Nil(err)
	snapshot3 := generateRandomSnapshot(3)
	err = s.UpdateDynamicConfig(ctx, snapshot3, 4)
	s.Nil(err)

	snapshot, err := s.FetchDynamicConfig(ctx, 4)
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

func (s *ConfigStorePersistenceSuite) FetchDynamicConfig(ctx context.Context, rowType int) (*p.DynamicConfigSnapshot, error) {
	response, err := s.ConfigStoreManager.FetchDynamicConfig(ctx, p.ConfigType(rowType))
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.New("nil FetchDynamicConfig response")
	}
	return response.Snapshot, nil
}

func (s *ConfigStorePersistenceSuite) UpdateDynamicConfig(ctx context.Context, snapshot *p.DynamicConfigSnapshot, rowType int) error {
	return s.ConfigStoreManager.UpdateDynamicConfig(ctx, &p.UpdateDynamicConfigRequest{Snapshot: snapshot}, p.ConfigType(rowType))
}
