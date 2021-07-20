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
	"sync/atomic"
	"time"

	dc "github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

var _ dc.Client = (*configStoreClient)(nil)

const (
	configStoreMinPollInterval = time.Second * 5
)

// ConfigStoreClientConfig is the config for the config store based dynamic config client.
// It specifies how often the cached config should be updated by checking underlying database.
type ConfigStoreClientConfig struct {
	PollInterval time.Duration `yaml:"pollInterval"`
}

type configStoreClient struct {
	values             atomic.Value
	lastUpdatedTime    time.Time
	config             *ConfigStoreClientConfig
	configStoreManager persistence.ConfigStoreManager
	doneCh             chan struct{}
	logger             log.Logger
}

type cacheEntry struct {
	cache_version  int64
	schema_version int64
	dc_entries     map[string]*types.DynamicConfigEntry
}

// NewConfigStoreClient creates a config store client
func NewConfigStoreClient(client_cfg *ConfigStoreClientConfig, manager persistence.ConfigStoreManager, logger log.Logger, doneCh chan struct{}) (dc.Client, error) {
	//persistence_cfg config.NoSQL
	if err := validateConfigStoreClientConfig(client_cfg); err != nil {
		return nil, err
	}

	// store, err := nosql.NewNoSQLConfigStore(persistence_cfg, logger)
	// if err != nil {
	// 	return nil, err
	// }

	client := &configStoreClient{
		config:             client_cfg,
		doneCh:             doneCh,
		configStoreManager: manager, //persistence.NewConfigStoreManagerImpl(store, logger),
		logger:             logger,
	}
	if err := client.update(); err != nil {
		return nil, err
	}
	go func() {
		ticker := time.NewTicker(client.config.PollInterval)
		for {
			select {
			case <-ticker.C:
				err := client.update()
				if err != nil {
					client.logger.Error("Failed to update cached dynamic config", tag.Error(err))
				}
			case <-client.doneCh:
				ticker.Stop()
				return
			}
		}
	}()
	return client, nil
}

func (csc *configStoreClient) GetValue(name dc.Key, defaultValue interface{}) (interface{}, error) {
	return csc.getValueWithFilters(name, nil, defaultValue)
}

func (csc *configStoreClient) GetValueWithFilters(name dc.Key, filters map[dc.Filter]interface{}, defaultValue interface{}) (interface{}, error) {
	return csc.getValueWithFilters(name, filters, defaultValue)
}

func (csc *configStoreClient) GetIntValue(name dc.Key, filters map[dc.Filter]interface{}, defaultValue int) (int, error) {
	val, err := csc.getValueWithFilters(name, filters, defaultValue)
	if err != nil {
		return defaultValue, err
	}

	if intVal, ok := val.(int); ok {
		return intVal, nil
	}
	return defaultValue, errors.New("value type is not int")
}

func (csc *configStoreClient) GetFloatValue(name dc.Key, filters map[dc.Filter]interface{}, defaultValue float64) (float64, error) {
	val, err := csc.getValueWithFilters(name, filters, defaultValue)
	if err != nil {
		return defaultValue, err
	}

	if floatVal, ok := val.(float64); ok {
		return floatVal, nil
	} else if intVal, ok := val.(int); ok {
		return float64(intVal), nil
	}
	return defaultValue, errors.New("value type is not float64")
}

func (csc *configStoreClient) GetBoolValue(name dc.Key, filters map[dc.Filter]interface{}, defaultValue bool) (bool, error) {
	val, err := csc.getValueWithFilters(name, filters, defaultValue)
	if err != nil {
		return defaultValue, err
	}

	if boolVal, ok := val.(bool); ok {
		return boolVal, nil
	}
	return defaultValue, errors.New("value type is not bool")
}

func (csc *configStoreClient) GetStringValue(name dc.Key, filters map[dc.Filter]interface{}, defaultValue string) (string, error) {
	val, err := csc.getValueWithFilters(name, filters, defaultValue)
	if err != nil {
		return defaultValue, err
	}

	if stringVal, ok := val.(string); ok {
		return stringVal, nil
	}
	return defaultValue, errors.New("value type is not string")
}

func (csc *configStoreClient) GetMapValue(
	name dc.Key, filters map[dc.Filter]interface{}, defaultValue map[string]interface{},
) (map[string]interface{}, error) {
	val, err := csc.getValueWithFilters(name, filters, defaultValue)
	if err != nil {
		return defaultValue, err
	}
	if mapVal, ok := val.(map[string]interface{}); ok {
		return mapVal, nil
	}
	return defaultValue, errors.New("value type is not map")
}

func (csc *configStoreClient) GetDurationValue(
	name dc.Key, filters map[dc.Filter]interface{}, defaultValue time.Duration,
) (time.Duration, error) {
	val, err := csc.getValueWithFilters(name, filters, defaultValue)
	if err != nil {
		return defaultValue, err
	}

	durationString, ok := val.(string)
	if !ok {
		return defaultValue, errors.New("value type is not string")
	}

	durationVal, err := time.ParseDuration(durationString)
	if err != nil {
		return defaultValue, fmt.Errorf("failed to parse duration: %v", err)
	}
	return durationVal, nil
}

func (csc *configStoreClient) UpdateValue(name dc.Key, value interface{}) error {
	//add retry logic
	currentCached := csc.values.Load().(cacheEntry)

	newEntries := make([]*types.DynamicConfigEntry, 0, len(currentCached.dc_entries))
	for _, v := range currentCached.dc_entries {
		newEntries = append(newEntries, v)
	}

	newSnapshot := &persistence.DynamicConfigSnapshot{
		Version: currentCached.cache_version + 1,
		Values: &types.DynamicConfigBlob{
			SchemaVersion: currentCached.schema_version,
			Entries:       newEntries,
		},
	}
	csc.configStoreManager.UpdateDynamicConfig(context.TODO(), newSnapshot)

	return nil
}

func (csc *configStoreClient) update() error {
	defer func() {
		csc.lastUpdatedTime = time.Now()
	}()

	dc_snapshot, err := csc.configStoreManager.FetchDynamicConfig(context.TODO())
	//if same version, then no need to store again (not yet implemented)

	if err != nil {
		return fmt.Errorf("failed to fetch dynamic config snapshot %v", err)
	}

	return csc.storeValues(dc_snapshot)
}

func (csc *configStoreClient) storeValues(snapshot *persistence.DynamicConfigSnapshot) error {
	//Converting the list of dynamic config entries into a map for better lookup performance
	dc_entry_map := make(map[string]*types.DynamicConfigEntry)
	for _, entry := range snapshot.Values.Entries {
		dc_entry_map[entry.Name] = entry
	}

	csc.values.Store(cacheEntry{
		cache_version:  snapshot.Version,
		schema_version: snapshot.Values.SchemaVersion,
		dc_entries:     dc_entry_map,
	})
	csc.logger.Info("Updated dynamic config")
	return nil
}

func validateConfigStoreClientConfig(config *ConfigStoreClientConfig) error {
	if config == nil {
		return errors.New("no config found for config store based dynamic config client")
	}
	if config.PollInterval < configStoreMinPollInterval {
		return fmt.Errorf("poll interval should be at least %v", configStoreMinPollInterval)
	}
	return nil
}

func convertFromDataBlob(blob *types.DataBlob) (interface{}, error) {
	switch *blob.EncodingType {
	case types.EncodingTypeJSON: //
		var v interface{}
		err := json.Unmarshal(blob.Data, v)
		return v, err
	default:
		return nil, errors.New("unsupported blob encoding")
	}
}

func (csc *configStoreClient) getValueWithFilters(key dc.Key, filters map[dc.Filter]interface{}, defaultValue interface{}) (interface{}, error) {
	keyName := dc.Keys[key]
	cached := csc.values.Load().(cacheEntry)
	dc_entries := cached.dc_entries
	found := false

	for _, dc_value := range dc_entries[keyName].Values {
		if len(dc_value.Filters) == 0 {
			parsed_val, err := convertFromDataBlob(dc_value.Value)
			if err == nil {
				defaultValue = parsed_val
				found = true
			}
			continue
		}

		if matchFilters(dc_value, filters) {
			return convertFromDataBlob(dc_value.Value)
		}
	}

	if !found {
		return defaultValue, dc.NotFoundError
	}
	return defaultValue, nil
}

func matchFilters(dc_value *types.DynamicConfigValue, filters map[dc.Filter]interface{}) bool {
	if len(dc_value.Filters) > len(filters) {
		return false
	}

	for _, value_filter := range dc_value.Filters {
		filterKey := dc.ParseFilter(value_filter.Name)
		if filters[filterKey] == nil {
			return false
		}

		request_value, err := convertFromDataBlob(value_filter.Value)
		if err != nil || filters[filterKey] != request_value {
			return false
		}
	}
	return true
}
