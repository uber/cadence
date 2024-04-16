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

package configstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	dc "github.com/uber/cadence/common/dynamicconfig"
	c "github.com/uber/cadence/common/dynamicconfig/configstore/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/types"
)

const (
	retryAttempts                = 2
	unknownKey    unknownKeyType = 0
)

type configStoreClientSuite struct {
	suite.Suite
	*require.Assertions
	client         *configStoreClient
	mockManager    *p.MockConfigStoreManager
	mockController *gomock.Controller
	doneCh         chan struct{}
}

var snapshot1 *p.DynamicConfigSnapshot

func TestConfigStoreClientSuite(t *testing.T) {
	s := new(configStoreClientSuite)
	suite.Run(t, s)
}

func (s *configStoreClientSuite) SetupSuite() {
	s.doneCh = make(chan struct{})
	s.mockController = gomock.NewController(s.T())

	mockPlugin := nosqlplugin.NewMockPlugin(s.mockController)
	mockPlugin.EXPECT().
		CreateDB(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil).AnyTimes()
	nosql.RegisterPlugin("cassandra", mockPlugin)
}

func (s *configStoreClientSuite) TearDownSuite() {
	close(s.doneCh)
}

func (s *configStoreClientSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	snapshot1 = &p.DynamicConfigSnapshot{
		Version: 1,
		Values: &types.DynamicConfigBlob{
			SchemaVersion: 1,
			Entries: []*types.DynamicConfigEntry{
				{
					Name: dc.TestGetBoolPropertyKey.String(),
					Values: []*types.DynamicConfigValue{
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeJSON.Ptr(),
								Data:         jsonMarshalHelper(false),
							},
							Filters: nil,
						},
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeJSON.Ptr(),
								Data:         jsonMarshalHelper(true),
							},
							Filters: []*types.DynamicConfigFilter{
								{
									Name: "domainName",
									Value: &types.DataBlob{
										EncodingType: types.EncodingTypeJSON.Ptr(),
										Data:         jsonMarshalHelper("global-samples-domain"),
									},
								},
							},
						},
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeJSON.Ptr(),
								Data:         jsonMarshalHelper(true),
							},
							Filters: []*types.DynamicConfigFilter{
								{
									Name: "domainName",
									Value: &types.DataBlob{
										EncodingType: types.EncodingTypeJSON.Ptr(),
										Data:         jsonMarshalHelper("samples-domain"),
									},
								},
							},
						},
					},
				},
				{
					Name: dc.TestGetIntPropertyKey.String(),
					Values: []*types.DynamicConfigValue{
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeJSON.Ptr(),
								Data:         jsonMarshalHelper(1000),
							},
							Filters: nil,
						},
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeJSON.Ptr(),
								Data:         jsonMarshalHelper(1000.1),
							},
							Filters: []*types.DynamicConfigFilter{
								{
									Name: "domainName",
									Value: &types.DataBlob{
										EncodingType: types.EncodingTypeJSON.Ptr(),
										Data:         jsonMarshalHelper("global-samples-domain"),
									},
								},
							},
						},
					},
				},
				{
					Name: dc.TestGetFloat64PropertyKey.String(),
					Values: []*types.DynamicConfigValue{
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeJSON.Ptr(),
								Data:         jsonMarshalHelper(12),
							},
							Filters: nil,
						},
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeJSON.Ptr(),
								Data:         jsonMarshalHelper("wrong type"),
							},
							Filters: []*types.DynamicConfigFilter{
								{
									Name: "domainName",
									Value: &types.DataBlob{
										EncodingType: types.EncodingTypeJSON.Ptr(),
										Data:         jsonMarshalHelper("samples-domain"),
									},
								},
							},
						},
					},
				},
				{
					Name: dc.TestGetStringPropertyKey.String(),
					Values: []*types.DynamicConfigValue{
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeJSON.Ptr(),
								Data:         jsonMarshalHelper("some random string"),
							},
							Filters: nil,
						},
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeJSON.Ptr(),
								Data:         jsonMarshalHelper("constrained-string"),
							},
							Filters: []*types.DynamicConfigFilter{
								{
									Name: "taskListName",
									Value: &types.DataBlob{
										EncodingType: types.EncodingTypeJSON.Ptr(),
										Data:         jsonMarshalHelper("random tasklist"),
									},
								},
							},
						},
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeJSON.Ptr(),
								Data:         jsonMarshalHelper(10),
							},
							Filters: []*types.DynamicConfigFilter{
								{
									Name: "DomainName",
									Value: &types.DataBlob{
										EncodingType: types.EncodingTypeJSON.Ptr(),
										Data:         jsonMarshalHelper("wrong-type-domain"),
									},
								},
							},
						},
					},
				},
				{
					Name: dc.TestGetMapPropertyKey.String(),
					Values: []*types.DynamicConfigValue{
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeJSON.Ptr(),
								Data: jsonMarshalHelper(map[string]interface{}{
									"key1": "1",
									"key2": 1,
									"key3": []interface{}{
										false,
										map[string]interface{}{
											"key4": true,
											"key5": 2.1,
										},
									},
								}),
							},
							Filters: nil,
						},
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeJSON.Ptr(),
								Data:         jsonMarshalHelper("1"),
							},
							Filters: []*types.DynamicConfigFilter{
								{
									Name: "taskListName",
									Value: &types.DataBlob{
										EncodingType: types.EncodingTypeJSON.Ptr(),
										Data:         jsonMarshalHelper("random tasklist"),
									},
								},
							},
						},
					},
				},
				{
					Name: dc.TestGetDurationPropertyKey.String(),
					Values: []*types.DynamicConfigValue{
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeJSON.Ptr(),
								Data:         jsonMarshalHelper("1m"),
							},
							Filters: nil,
						},
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeJSON.Ptr(),
								Data:         jsonMarshalHelper("wrong duration string"),
							},
							Filters: []*types.DynamicConfigFilter{
								{
									Name: "domainName",
									Value: &types.DataBlob{
										EncodingType: types.EncodingTypeJSON.Ptr(),
										Data:         jsonMarshalHelper("samples-domain"),
									},
								},
								{
									Name: "taskListName",
									Value: &types.DataBlob{
										EncodingType: types.EncodingTypeJSON.Ptr(),
										Data:         jsonMarshalHelper("longIdleTimeTaskList"),
									},
								},
							},
						},
						{
							Value: &types.DataBlob{
								EncodingType: types.EncodingTypeJSON.Ptr(),
								Data:         jsonMarshalHelper(2),
							},
							Filters: []*types.DynamicConfigFilter{
								{
									Name: "domainName",
									Value: &types.DataBlob{
										EncodingType: types.EncodingTypeJSON.Ptr(),
										Data:         jsonMarshalHelper("samples-domain"),
									},
								},
							},
						},
					},
				},
				{
					Name: dc.TestGetListPropertyKey.String(),
					Values: []*types.DynamicConfigValue{
						{
							Value:   &types.DataBlob{EncodingType: types.EncodingTypeJSON.Ptr(), Data: jsonMarshalHelper([]interface{}{1, 2, 3})},
							Filters: nil,
						},
						{
							Value: &types.DataBlob{EncodingType: types.EncodingTypeJSON.Ptr(), Data: jsonMarshalHelper([]interface{}{"a", "b", "c"})},
							Filters: []*types.DynamicConfigFilter{
								{
									Name: "domainName",
									Value: &types.DataBlob{
										EncodingType: types.EncodingTypeJSON.Ptr(),
										Data:         jsonMarshalHelper("sample-domain"),
									},
								},
							},
						},
						{
							Value: &types.DataBlob{EncodingType: types.EncodingTypeJSON.Ptr(), Data: jsonMarshalHelper("text")},
							Filters: []*types.DynamicConfigFilter{
								{
									Name: "domainName",
									Value: &types.DataBlob{
										EncodingType: types.EncodingTypeJSON.Ptr(),
										Data:         jsonMarshalHelper("wrong-type-domain"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	connections := make(map[string]config.DBShardConnection)
	connections[config.NonShardedStoreName] = config.DBShardConnection{
		NoSQLPlugin: &config.NoSQL{
			PluginName: "cassandra",
		},
	}

	var err error
	s.client, err = newConfigStoreClient(
		&c.ClientConfig{
			PollInterval:        time.Second * 2,
			UpdateRetryAttempts: retryAttempts,
			FetchTimeout:        time.Second * 1,
			UpdateTimeout:       time.Second * 1,
		},
		&config.DataStore{
			ShardedNoSQL: &config.ShardedNoSQL{
				DefaultShard: config.NonShardedStoreName,
				Connections:  connections,
			},
		}, log.NewNoop(), p.DynamicConfig)
	s.Require().NoError(err)

	s.mockManager = p.NewMockConfigStoreManager(s.mockController)
	s.client.configStoreManager = s.mockManager
}

func defaultTestSetup(s *configStoreClientSuite) {
	s.mockManager.EXPECT().
		FetchDynamicConfig(gomock.Any(), p.DynamicConfig).
		Return(&p.FetchDynamicConfigResponse{
			Snapshot: snapshot1,
		}, nil).
		AnyTimes()
	err := s.client.startUpdate()
	s.NoError(err)
}

func (s *configStoreClientSuite) TestGetValue() {
	defaultTestSetup(s)
	v, err := s.client.GetValue(dc.TestGetBoolPropertyKey)
	s.NoError(err)
	s.Equal(false, v)
}

func (s *configStoreClientSuite) TestGetValue_NonExistKey() {
	defaultTestSetup(s)
	v, err := s.client.GetValue(dc.MaxRetentionDays)
	s.Error(err)
	s.Equal(dc.MaxRetentionDays.DefaultInt(), v)
	v, err = s.client.GetValue(dc.EnableVisibilitySampling)
	s.Error(err)
	s.Equal(dc.EnableVisibilitySampling.DefaultBool(), v)
	v, err = s.client.GetValue(dc.FrontendErrorInjectionRate)
	s.Error(err)
	s.Equal(dc.FrontendErrorInjectionRate.DefaultFloat(), v)
	v, err = s.client.GetValue(dc.AdvancedVisibilityWritingMode)
	s.Error(err)
	s.Equal(dc.AdvancedVisibilityWritingMode.DefaultString(), v)
	v, err = s.client.GetValue(dc.FrontendShutdownDrainDuration)
	s.Error(err)
	s.Equal(dc.FrontendShutdownDrainDuration.DefaultDuration(), v)
	v, err = s.client.GetValue(dc.RequiredDomainDataKeys)
	s.Error(err)
	s.Equal(dc.RequiredDomainDataKeys.DefaultMap(), v)
}

func (s *configStoreClientSuite) TestGetValueWithFilters() {
	defaultTestSetup(s)

	filters := map[dc.Filter]interface{}{
		dc.DomainName: "global-samples-domain",
	}

	v, err := s.client.GetValueWithFilters(dc.TestGetBoolPropertyKey, filters)
	s.NoError(err)
	s.Equal(true, v)

	filters = map[dc.Filter]interface{}{
		dc.DomainName: "non-exist-domain",
	}
	v, err = s.client.GetValueWithFilters(dc.TestGetBoolPropertyKey, filters)
	s.NoError(err)
	s.Equal(false, v)

	filters = map[dc.Filter]interface{}{
		dc.DomainName:   "samples-domain",
		dc.TaskListName: "non-exist-tasklist",
	}
	v, err = s.client.GetValueWithFilters(dc.TestGetBoolPropertyKey, filters)
	s.NoError(err)
	s.Equal(true, v)
}

func (s *configStoreClientSuite) TestGetValueWithFilters_UnknownFilter() {
	defaultTestSetup(s)
	filters := map[dc.Filter]interface{}{
		dc.DomainName:    "global-samples-domain1",
		dc.UnknownFilter: "unknown-filter1",
	}
	v, err := s.client.GetValueWithFilters(dc.TestGetBoolPropertyKey, filters)
	s.NoError(err)
	s.Equal(false, v)
}

func (s *configStoreClientSuite) TestGetIntValue() {
	defaultTestSetup(s)
	v, err := s.client.GetIntValue(dc.TestGetIntPropertyKey, nil)
	s.NoError(err)
	s.Equal(1000, v)
}

func (s *configStoreClientSuite) TestGetIntValue_FilterNotMatch() {
	defaultTestSetup(s)
	filters := map[dc.Filter]interface{}{
		dc.DomainName: "samples-domain",
	}
	v, err := s.client.GetIntValue(dc.TestGetIntPropertyKey, filters)
	s.NoError(err)
	s.Equal(1000, v)
}

func (s *configStoreClientSuite) TestGetIntValue_WrongType() {
	defaultTestSetup(s)
	filters := map[dc.Filter]interface{}{
		dc.DomainName: "global-samples-domain",
	}
	v, err := s.client.GetIntValue(dc.TestGetIntPropertyKey, filters)
	s.Error(err)
	s.Equal(dc.TestGetIntPropertyKey.DefaultInt(), v)
}

func (s *configStoreClientSuite) TestGetIntValueNonExistentKey() {
	defaultTestSetup(s)
	v, err := s.client.GetIntValue(dc.UnknownIntKey, nil)
	s.Equal(err, dc.NotFoundError)
	s.Equal(dc.UnknownIntKey.DefaultInt(), v)
}

func (s *configStoreClientSuite) TestGetFloatValue() {
	defaultTestSetup(s)
	v, err := s.client.GetFloatValue(dc.TestGetFloat64PropertyKey, nil)
	s.NoError(err)
	s.Equal(12.0, v)
}

func (s *configStoreClientSuite) TestGetFloatValue_WrongType() {
	defaultTestSetup(s)
	filters := map[dc.Filter]interface{}{
		dc.DomainName: "samples-domain",
	}
	v, err := s.client.GetFloatValue(dc.TestGetFloat64PropertyKey, filters)
	s.Error(err)
	s.Equal(dc.TestGetFloat64PropertyKey.DefaultFloat(), v)
}

func (s *configStoreClientSuite) TestGetFloatValueNonExistentKey() {
	defaultTestSetup(s)
	v, err := s.client.GetFloatValue(dc.UnknownFloatKey, nil)
	s.Equal(err, dc.NotFoundError)
	s.Equal(dc.UnknownFloatKey.DefaultFloat(), v)
}

func (s *configStoreClientSuite) TestGetBoolValue() {
	defaultTestSetup(s)
	v, err := s.client.GetBoolValue(dc.TestGetBoolPropertyKey, nil)
	s.NoError(err)
	s.Equal(false, v)
}

func (s *configStoreClientSuite) TestGetBoolValueNonExistentKey() {
	defaultTestSetup(s)
	v, err := s.client.GetBoolValue(dc.UnknownBoolKey, nil)
	s.Equal(dc.NotFoundError, err)
	s.Equal(dc.UnknownBoolKey.DefaultValue(), v)
}

func (s *configStoreClientSuite) TestGetStringValue() {
	defaultTestSetup(s)
	filters := map[dc.Filter]interface{}{
		dc.TaskListName: "random tasklist",
	}
	v, err := s.client.GetStringValue(dc.TestGetStringPropertyKey, filters)
	s.NoError(err)
	s.Equal("constrained-string", v)
}

func (s *configStoreClientSuite) TestGetStringValueNonExistentKey() {
	defaultTestSetup(s)
	v, err := s.client.GetStringValue(dc.UnknownStringKey, nil)
	s.Equal(dc.NotFoundError, err)
	s.Equal(dc.UnknownStringKey.DefaultValue(), v)
}

func (s *configStoreClientSuite) TestGetMapValue() {
	defaultTestSetup(s)
	v, err := s.client.GetMapValue(dc.TestGetMapPropertyKey, nil)
	s.NoError(err)
	expectedVal := map[string]interface{}{
		"key1": "1",
		"key2": float64(1),
		"key3": []interface{}{
			false,
			map[string]interface{}{
				"key4": true,
				"key5": 2.1,
			},
		},
	}
	s.Equal(expectedVal, v)
}

func (s *configStoreClientSuite) TestGetMapValueNonExistentKey() {
	defaultTestSetup(s)
	v, err := s.client.GetMapValue(dc.UnknownMapKey, nil)
	s.Equal(dc.NotFoundError, err)
	s.Equal(dc.UnknownMapKey.DefaultValue(), v)
}

func (s *configStoreClientSuite) TestGetMapValue_WrongType() {
	defaultTestSetup(s)
	filters := map[dc.Filter]interface{}{
		dc.TaskListName: "random tasklist",
	}
	v, err := s.client.GetMapValue(dc.TestGetMapPropertyKey, filters)
	s.Error(err)
	s.Equal(dc.TestGetMapPropertyKey.DefaultMap(), v)
}

func (s *configStoreClientSuite) TestGetDurationValue() {
	defaultTestSetup(s)
	v, err := s.client.GetDurationValue(dc.TestGetDurationPropertyKey, nil)
	s.NoError(err)
	s.Equal(time.Minute, v)
}

func (s *configStoreClientSuite) TestGetDurationValueNonExistentKey() {
	defaultTestSetup(s)
	v, err := s.client.GetDurationValue(dc.UnknownDurationKey, nil)
	s.Equal(dc.NotFoundError, err)
	s.Equal(dc.UnknownDurationKey.DefaultValue(), v)
}

func (s *configStoreClientSuite) TestGetDurationValue_NotStringRepresentation() {
	defaultTestSetup(s)
	filters := map[dc.Filter]interface{}{
		dc.DomainName: "samples-domain",
	}
	v, err := s.client.GetDurationValue(dc.TestGetDurationPropertyKey, filters)
	s.Error(err)
	s.Equal(dc.TestGetDurationPropertyKey.DefaultDuration(), v)
}

func (s *configStoreClientSuite) TestGetDurationValue_ParseFailed() {
	defaultTestSetup(s)
	filters := map[dc.Filter]interface{}{
		dc.DomainName:   "samples-domain",
		dc.TaskListName: "longIdleTimeTaskList",
	}
	v, err := s.client.GetDurationValue(dc.TestGetDurationPropertyKey, filters)
	s.Error(err)
	s.Equal(dc.TestGetDurationPropertyKey.DefaultDuration(), v)
}

func (s *configStoreClientSuite) TestValidateConfig_InvalidConfig() {
	err := validateClientConfig(
		&c.ClientConfig{
			PollInterval:        time.Second * 1,
			UpdateRetryAttempts: 0,
			FetchTimeout:        time.Second * 3,
			UpdateTimeout:       time.Second * 4,
		},
	)
	s.Error(err)

	err = validateClientConfig(
		&c.ClientConfig{
			PollInterval:        time.Second * 2,
			UpdateRetryAttempts: -1,
			FetchTimeout:        time.Second * 2,
			UpdateTimeout:       time.Second * 2,
		},
	)
	s.Error(err)

	err = validateClientConfig(
		&c.ClientConfig{
			PollInterval:        time.Second * 2,
			UpdateRetryAttempts: 0,
			FetchTimeout:        time.Second * 0,
			UpdateTimeout:       time.Second * 0,
		},
	)
	s.Error(err)

	err = validateClientConfig(
		&c.ClientConfig{
			PollInterval:        time.Second * 2,
			UpdateRetryAttempts: 1,
			FetchTimeout:        time.Second * 1,
			UpdateTimeout:       time.Second * 0,
		},
	)
	s.Error(err)
}

func (s *configStoreClientSuite) TestMatchFilters() {
	testCases := []struct {
		v       *types.DynamicConfigValue
		filters map[dc.Filter]interface{}
		matched bool
	}{
		{
			v: &types.DynamicConfigValue{
				Value:   nil,
				Filters: nil,
			},
			filters: map[dc.Filter]interface{}{
				dc.DomainName: "some random domain",
			},
			matched: true,
		},
		{
			v: &types.DynamicConfigValue{
				Value: nil,
				Filters: []*types.DynamicConfigFilter{
					{
						Name: "some key",
						Value: &types.DataBlob{
							EncodingType: types.EncodingTypeJSON.Ptr(),
							Data:         jsonMarshalHelper("some value"),
						},
					},
				},
			},
			filters: map[dc.Filter]interface{}{},
			matched: false,
		},
		{
			v: &types.DynamicConfigValue{
				Value: nil,
				Filters: []*types.DynamicConfigFilter{
					{
						Name: "domainName",
						Value: &types.DataBlob{
							EncodingType: types.EncodingTypeJSON.Ptr(),
							Data:         jsonMarshalHelper("samples-domain"),
						},
					},
				},
			},
			filters: map[dc.Filter]interface{}{
				dc.DomainName: "some random domain",
			},
			matched: false,
		},
		{
			v: &types.DynamicConfigValue{
				Value: nil,
				Filters: []*types.DynamicConfigFilter{
					{
						Name: "domainName",
						Value: &types.DataBlob{
							EncodingType: types.EncodingTypeJSON.Ptr(),
							Data:         jsonMarshalHelper("samples-domain"),
						},
					},
					{
						Name: "taskListName",
						Value: &types.DataBlob{
							EncodingType: types.EncodingTypeJSON.Ptr(),
							Data:         jsonMarshalHelper("sample-task-list"),
						},
					},
				},
			},
			filters: map[dc.Filter]interface{}{
				dc.DomainName:   "samples-domain",
				dc.TaskListName: "sample-task-list",
			},
			matched: true,
		},
		{
			v: &types.DynamicConfigValue{
				Value: nil,
				Filters: []*types.DynamicConfigFilter{
					{
						Name: "domainName",
						Value: &types.DataBlob{
							EncodingType: types.EncodingTypeJSON.Ptr(),
							Data:         jsonMarshalHelper("samples-domain"),
						},
					},
					{
						Name: "some-other-filter",
						Value: &types.DataBlob{
							EncodingType: types.EncodingTypeJSON.Ptr(),
							Data:         jsonMarshalHelper("sample-task-list"),
						},
					},
				},
			},
			filters: map[dc.Filter]interface{}{
				dc.DomainName:   "samples-domain",
				dc.TaskListName: "sample-task-list",
			},
			matched: false,
		},
		{
			v: &types.DynamicConfigValue{
				Value: nil,
				Filters: []*types.DynamicConfigFilter{
					{
						Name: "domainName",
						Value: &types.DataBlob{
							EncodingType: types.EncodingTypeJSON.Ptr(),
							Data:         jsonMarshalHelper("samples-domain"),
						},
					},
				},
			},
			filters: map[dc.Filter]interface{}{
				dc.TaskListName: "sample-task-list",
			},
			matched: false,
		},
	}

	for index, tc := range testCases {
		matched := matchFilters(tc.v, tc.filters)
		s.Equal(tc.matched, matched, fmt.Sprintf("Test case %v failed", index))
	}
}

func (s *configStoreClientSuite) TestUpdateValue_NilOverwrite() {
	defaultTestSetup(s)

	s.mockManager.EXPECT().
		UpdateDynamicConfig(gomock.Any(), gomock.Any(), p.DynamicConfig).
		DoAndReturn(func(_ context.Context, request *p.UpdateDynamicConfigRequest, cfgType p.ConfigType) error {
			if request.Snapshot.Values.Entries[0].Name != dc.TestGetBoolPropertyKey.String() {
				return nil
			}
			return errors.New("entry not removed")
		}).AnyTimes()

	err := s.client.UpdateValue(dc.TestGetBoolPropertyKey, nil)
	s.NoError(err)
}

func (s *configStoreClientSuite) TestUpdateValue_NoRetrySuccess() {
	defaultTestSetup(s)

	s.mockManager.EXPECT().
		UpdateDynamicConfig(gomock.Any(), EqSnapshotVersion(2), p.DynamicConfig).
		Return(nil).MaxTimes(1)

	values := []*types.DynamicConfigValue{
		{
			Value: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper(true),
			},
			Filters: nil,
		},
	}

	err := s.client.UpdateValue(dc.TestGetBoolPropertyKey, values)
	s.NoError(err)

	snapshot2 := snapshot1
	snapshot2.Values.Entries[0].Values = values
	s.mockManager.EXPECT().
		FetchDynamicConfig(gomock.Any(), p.DynamicConfig).
		Return(&p.FetchDynamicConfigResponse{
			Snapshot: snapshot2,
		}, nil).MaxTimes(1)

	err = s.client.update()
	s.NoError(err)

	v, err := s.client.GetValue(dc.TestGetBoolPropertyKey)
	s.NoError(err)
	s.Equal(true, v)
}

func (s *configStoreClientSuite) TestUpdateValue_SuccessNewKey() {
	values := []*types.DynamicConfigValue{
		{
			Value: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper(true),
			},
			Filters: nil,
		},
	}

	s.mockManager.EXPECT().
		FetchDynamicConfig(gomock.Any(), p.DynamicConfig).
		Return(&p.FetchDynamicConfigResponse{
			Snapshot: &p.DynamicConfigSnapshot{
				Version: 1,
				Values: &types.DynamicConfigBlob{
					SchemaVersion: 1,
					Entries:       nil,
				},
			},
		}, nil).
		AnyTimes()

	s.mockManager.EXPECT().
		UpdateDynamicConfig(gomock.Any(), gomock.Any(), p.DynamicConfig).
		DoAndReturn(func(_ context.Context, request *p.UpdateDynamicConfigRequest, cfgType p.ConfigType) error {
			s.Equal(1, len(request.Snapshot.Values.Entries))
			s.Equal(request.Snapshot.Values.Entries[0].Values, values)
			return nil
		}).AnyTimes()

	s.client.update()
	err := s.client.UpdateValue(dc.TestGetBoolPropertyKey, values)
	s.NoError(err)
}

func (s *configStoreClientSuite) TestUpdateValue_RetrySuccess() {
	s.mockManager.EXPECT().
		UpdateDynamicConfig(gomock.Any(), EqSnapshotVersion(2), p.DynamicConfig).
		Return(&p.ConditionFailedError{}).AnyTimes()

	s.mockManager.EXPECT().
		UpdateDynamicConfig(gomock.Any(), EqSnapshotVersion(3), p.DynamicConfig).
		Return(nil).AnyTimes()

	snapshot1.Version = 2
	s.mockManager.EXPECT().
		FetchDynamicConfig(gomock.Any(), p.DynamicConfig).
		Return(&p.FetchDynamicConfigResponse{
			Snapshot: snapshot1,
		}, nil).AnyTimes()

	s.client.update()

	err := s.client.UpdateValue(dc.TestGetBoolPropertyKey, []*types.DynamicConfigValue{})
	s.NoError(err)
}

func (s *configStoreClientSuite) TestUpdateValue_RetryFailure() {
	defaultTestSetup(s)

	s.mockManager.EXPECT().
		UpdateDynamicConfig(gomock.Any(), gomock.Any(), p.DynamicConfig).
		Return(&p.ConditionFailedError{}).MaxTimes(retryAttempts + 1)

	err := s.client.UpdateValue(dc.TestGetFloat64PropertyKey, []*types.DynamicConfigValue{})
	s.Error(err)
}

func (s *configStoreClientSuite) TestUpdateValue_Timeout() {
	defaultTestSetup(s)
	s.mockManager.EXPECT().
		UpdateDynamicConfig(gomock.Any(), gomock.Any(), p.DynamicConfig).
		DoAndReturn(func(_ context.Context, _ *p.UpdateDynamicConfigRequest, cfgType p.ConfigType) error {
			time.Sleep(2 * time.Second)
			return nil
		}).AnyTimes()

	err := s.client.UpdateValue(dc.TestGetDurationPropertyKey, []*types.DynamicConfigValue{})
	s.Error(err)
}

func (s *configStoreClientSuite) TestRestoreValue_NoFilter() {
	defaultTestSetup(s)
	s.mockManager.EXPECT().
		UpdateDynamicConfig(gomock.Any(), gomock.Any(), p.DynamicConfig).
		DoAndReturn(func(_ context.Context, request *p.UpdateDynamicConfigRequest, cfgType p.ConfigType) error {
			for _, entry := range request.Snapshot.Values.Entries {
				if entry.Name == dc.TestGetBoolPropertyKey.String() {
					for _, value := range entry.Values {
						s.Equal(value.Value.Data, jsonMarshalHelper(true))
						if value.Filters == nil {
							return errors.New("fallback value not restored")
						}
					}
				}
			}
			return nil
		}).AnyTimes()

	err := s.client.RestoreValue(dc.TestGetBoolPropertyKey, nil)
	s.NoError(err)
}

func (s *configStoreClientSuite) TestRestoreValue_FilterNoMatch() {
	defaultTestSetup(s)

	s.mockManager.EXPECT().
		UpdateDynamicConfig(gomock.Any(), gomock.Any(), p.DynamicConfig).
		DoAndReturn(func(_ context.Context, request *p.UpdateDynamicConfigRequest, cfgType p.ConfigType) error {
			for _, resEntry := range request.Snapshot.Values.Entries {
				for _, oriEntry := range snapshot1.Values.Entries {
					if oriEntry.Name == resEntry.Name {
						s.Equal(resEntry.Values, oriEntry.Values)
					}
				}
			}
			return nil
		}).AnyTimes()

	noMatchFilter := map[dc.Filter]interface{}{
		dc.DomainName: "unknown-domain",
	}

	err := s.client.RestoreValue(dc.TestGetBoolPropertyKey, noMatchFilter)
	s.NoError(err)
}

func (s *configStoreClientSuite) TestRestoreValue_FilterMatch() {
	defaultTestSetup(s)
	s.mockManager.EXPECT().
		UpdateDynamicConfig(gomock.Any(), gomock.Any(), p.DynamicConfig).
		DoAndReturn(func(_ context.Context, request *p.UpdateDynamicConfigRequest, cfgType p.ConfigType) error {
			for _, resEntry := range request.Snapshot.Values.Entries {
				if resEntry.Name == dc.TestGetBoolPropertyKey.String() {
					s.Equal(2, len(resEntry.Values))
				}
			}
			return nil
		}).AnyTimes()

	filters := map[dc.Filter]interface{}{
		dc.DomainName: "samples-domain",
	}

	err := s.client.RestoreValue(dc.TestGetBoolPropertyKey, filters)
	s.NoError(err)
}

func (s *configStoreClientSuite) TestListValues() {
	defaultTestSetup(s)
	val, err := s.client.ListValue(nil)
	s.NoError(err)
	for _, resEntry := range val {
		for _, oriEntry := range snapshot1.Values.Entries {
			if oriEntry.Name == resEntry.Name {
				s.Equal(resEntry.Values, oriEntry.Values)
			}
		}
	}
}

func (s *configStoreClientSuite) TestGetListValue() {
	tests := []struct {
		desc    string
		key     dc.ListKey
		filters map[dc.Filter]interface{}
		res     interface{}
		err     error
	}{
		{
			desc:    "return error for non-existing key",
			key:     dc.UnknownListKey,
			filters: nil,
			res:     *(new([]interface{})),
			err:     dc.NotFoundError,
		},
		{
			desc:    "return correct value for existing key",
			key:     dc.TestGetListPropertyKey,
			filters: nil,
			res:     []interface{}{float64(1), float64(2), float64(3)},
			err:     nil,
		},
		{
			desc: "return empty filter value for existing key with no matching filter",
			key:  dc.TestGetListPropertyKey,
			filters: map[dc.Filter]interface{}{
				dc.DomainName: "no-match-domain",
			},
			res: []interface{}{float64(1), float64(2), float64(3)},
			err: nil,
		},
		{
			desc: "return correct value for existing key with matching filter",
			key:  dc.TestGetListPropertyKey,
			filters: map[dc.Filter]interface{}{
				dc.DomainName: "sample-domain",
			},
			res: []interface{}{"a", "b", "c"},
			err: nil,
		},
		{
			desc: "return error on wrong type of value",
			key:  dc.TestGetListPropertyKey,
			filters: map[dc.Filter]interface{}{
				dc.DomainName: "wrong-type-domain",
			},
			res: *(new([]interface{})),
			err: errors.New("value type is not list"),
		},
	}
	defaultTestSetup(s)
	for _, test := range tests {
		val, err := s.client.GetListValue(test.key, test.filters)
		s.Equal(test.err, err)
		s.Equal(test.res, val)
	}
}

func (s *configStoreClientSuite) TestListValues_EmptyCache() {
	s.mockManager.EXPECT().
		FetchDynamicConfig(gomock.Any(), p.DynamicConfig).
		Return(&p.FetchDynamicConfigResponse{
			Snapshot: &p.DynamicConfigSnapshot{
				Version: 1,
				Values: &types.DynamicConfigBlob{
					SchemaVersion: 1,
					Entries:       nil,
				},
			},
		}, nil).
		MaxTimes(1)

	s.client.update()

	val, err := s.client.ListValue(nil)
	s.NoError(err)
	s.Nil(val)
}

func (s *configStoreClientSuite) TestConfigStoreClientStart() {
	defaultTestSetup(s)
	s.client.Start()
	s.Equal(common.DaemonStatusStarted, s.client.status)
}

func (s *configStoreClientSuite) TestConfigStoreClientStop() {
	s.mockManager.EXPECT().Close()
	s.client.Stop()
	s.Equal(common.DaemonStatusStopped, s.client.status)
}

func jsonMarshalHelper(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

type eqSnapshotVersionMatcher struct {
	version int64
}

func (e eqSnapshotVersionMatcher) Matches(x interface{}) bool {
	arg, ok := x.(*p.UpdateDynamicConfigRequest)
	if !ok {
		return false
	}
	return e.version == arg.Snapshot.Version
}

func (e eqSnapshotVersionMatcher) String() string {
	return fmt.Sprintf("Version match %d.\n", e.version)
}

func EqSnapshotVersion(version int64) gomock.Matcher {
	return eqSnapshotVersionMatcher{version}
}

func TestValidateClientConfig(t *testing.T) {
	tests := []struct {
		desc          string
		input         *c.ClientConfig
		checkForError bool
	}{
		{
			desc:          "return error for nil config",
			input:         nil,
			checkForError: true,
		},
		{
			desc: "return error for minimum PollInterval",
			input: &c.ClientConfig{
				PollInterval: 1 * time.Second,
			},
			checkForError: true,
		},
		{
			desc: "return error for invalid UpdateRetryAttempts",
			input: &c.ClientConfig{
				PollInterval:        3 * time.Second,
				UpdateRetryAttempts: -1,
			},
			checkForError: true,
		},
		{
			desc: "return error for invalid FetchTimeout",
			input: &c.ClientConfig{
				PollInterval:        3 * time.Second,
				UpdateRetryAttempts: 1,
				FetchTimeout:        -1 * time.Second,
			},
			checkForError: true,
		},
		{
			desc: "return error for invalid UpdateTimeout",
			input: &c.ClientConfig{
				PollInterval:        3 * time.Second,
				UpdateRetryAttempts: 1,
				FetchTimeout:        1 * time.Second,
				UpdateTimeout:       -1 * time.Second,
			},
			checkForError: true,
		},
		{
			desc: "no error for valid config",
			input: &c.ClientConfig{
				PollInterval:        3 * time.Second,
				UpdateRetryAttempts: 1,
				FetchTimeout:        1 * time.Second,
				UpdateTimeout:       1 * time.Second,
			},
			checkForError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			err := validateClientConfig(test.input)
			if !test.checkForError {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestConvertFromDataBlob(t *testing.T) {
	tests := []struct {
		desc       string
		blob       *types.DataBlob
		checkError bool
	}{
		{
			desc: "return error for unsupported encoding type",
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeThriftRW.Ptr(),
				Data:         nil,
			},
			checkError: true,
		},
		{
			desc: "no error for valid json",
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper("test"),
			},
			checkError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			_, err := convertFromDataBlob(test.blob)
			if test.checkError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type unknownKeyType int

func (uk unknownKeyType) String() string {
	return ""
}

func (uk unknownKeyType) Description() string {
	return ""
}

func (uk unknownKeyType) DefaultValue() interface{} {
	return struct{}{}
}

func (uk unknownKeyType) Filters() []dc.Filter {
	return nil
}

func TestValidateKeyDataBlobPair(t *testing.T) {
	tests := []struct {
		desc       string
		key        dc.Key
		blob       *types.DataBlob
		checkError bool
	}{
		{
			desc: "return error for unsupported encoding type",
			key:  dc.TestGetIntPropertyKey,
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeThriftRW.Ptr(),
				Data:         jsonMarshalHelper(1),
			},
			checkError: true,
		},
		{
			desc: "return error for invalid value type for int key",
			key:  dc.TestGetIntPropertyKey,
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper("text"),
			},
			checkError: true,
		},
		{
			desc: "return no error for valid value type for int key",
			key:  dc.TestGetIntPropertyKey,
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper(1),
			},
			checkError: false,
		},
		{
			desc: "return error for invalid value type for bool key",
			key:  dc.TestGetBoolPropertyKey,
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper(1),
			},
			checkError: true,
		},
		{
			desc: "return no error for valid value type for bool key",
			key:  dc.TestGetBoolPropertyKey,
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper(true),
			},
			checkError: false,
		},
		{
			desc: "return error for invalid value type for float key",
			key:  dc.TestGetFloat64PropertyKey,
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper("text"),
			},
			checkError: true,
		},
		{
			desc: "return no error for valid value type for float key",
			key:  dc.TestGetFloat64PropertyKey,
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper(2.5),
			},
			checkError: false,
		},
		{
			desc: "return error for invalid value type for string key",
			key:  dc.TestGetStringPropertyKey,
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper(false),
			},
			checkError: true,
		},
		{
			desc: "return no error for valid value type for string key",
			key:  dc.TestGetStringPropertyKey,
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper("text"),
			},
			checkError: false,
		},
		{
			desc: "return error for invalid value type for duration key",
			key:  dc.TestGetDurationPropertyKey,
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper("text"),
			},
			checkError: true,
		},
		{
			desc: "return no error for valid value type for duration key",
			key:  dc.TestGetDurationPropertyKey,
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper("1m"),
			},
			checkError: false,
		},
		{
			desc: "return error for invalid value type for map key",
			key:  dc.TestGetMapPropertyKey,
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper("text"),
			},
			checkError: true,
		},
		{
			desc: "return no error for valid value type for map key",
			key:  dc.TestGetMapPropertyKey,
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper(map[string]interface{}{"key": 1}),
			},
			checkError: false,
		},
		{
			desc: "return error for invalid value type for list key",
			key:  dc.TestGetListPropertyKey,
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper("text"),
			},
			checkError: true,
		},
		{
			desc: "return no error for valid value type for list key",
			key:  dc.TestGetListPropertyKey,
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper([]interface{}{1, 2, 3}),
			},
			checkError: false,
		},
		{
			desc: "return error for unknown key",
			key:  unknownKey,
			blob: &types.DataBlob{
				EncodingType: types.EncodingTypeJSON.Ptr(),
				Data:         jsonMarshalHelper(1),
			},
			checkError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			err := validateKeyDataBlobPair(test.key, test.blob)
			if test.checkError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewConfigStoreClient(t *testing.T) {
	tests := []struct {
		desc              string
		clientConfig      *c.ClientConfig
		persistenceConfig *config.Persistence
		configType        persistence.ConfigType
		logger            log.Logger
		checkError        bool
	}{
		{
			desc:              "return error when persistenceConfig is nil",
			clientConfig:      nil,
			persistenceConfig: nil,
			checkError:        true,
		},
		{
			desc:              "return error when defaultstore not exists in persistenceConfig datastores",
			clientConfig:      &c.ClientConfig{},
			persistenceConfig: &config.Persistence{DefaultStore: "test"},
			configType:        0,
			logger:            log.NewNoop(),
			checkError:        true,
		},
		{
			desc:              "return error when invalid datastore configuration provided",
			clientConfig:      &c.ClientConfig{},
			persistenceConfig: &config.Persistence{DefaultStore: "test", DataStores: map[string]config.DataStore{"test": {}}},
			configType:        0,
			logger:            log.NewNoop(),
			checkError:        true,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			client, err := NewConfigStoreClient(test.clientConfig, test.persistenceConfig, test.logger, test.configType)
			if test.checkError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}
