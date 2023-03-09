// Copyright (c) 2021 Uber Technologies, Inc.
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
	"math"
	"sync/atomic"
	"time"

	"github.com/uber/cadence/common/config"
	dc "github.com/uber/cadence/common/dynamicconfig"
	csc "github.com/uber/cadence/common/dynamicconfig/configstore/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql"
	"github.com/uber/cadence/common/types"
)

var _ dc.Client = (*configStoreClient)(nil)

const (
	configStoreMinPollInterval = time.Second * 2
)

var defaultConfigValues = &csc.ClientConfig{
	PollInterval:        time.Second * 10,
	UpdateRetryAttempts: 1,
	FetchTimeout:        2,
	UpdateTimeout:       2,
}

type configStoreClient struct {
	values             atomic.Value
	lastUpdatedTime    time.Time
	config             *csc.ClientConfig
	configStoreManager persistence.ConfigStoreManager
	doneCh             chan struct{}
	logger             log.Logger
}

type cacheEntry struct {
	cacheVersion  int64
	schemaVersion int64
	dcEntries     map[string]*types.DynamicConfigEntry
}

// NewConfigStoreClient creates a config store client
func NewConfigStoreClient(clientCfg *csc.ClientConfig, persistenceCfg *config.Persistence, logger log.Logger, doneCh chan struct{}) (dc.Client, error) {
	if err := validateClientConfig(clientCfg); err != nil {
		logger.Error("Invalid Client Config Values, Using Default Values")
		clientCfg = defaultConfigValues
	}

	if persistenceCfg == nil {
		return nil, errors.New("persistence cfg is nil")
	} else if persistenceCfg.DefaultStore != "cass-default" {
		return nil, fmt.Errorf("persistence cfg default store is not Cassandra")
	} else if store, ok := persistenceCfg.DataStores[persistenceCfg.DefaultStore]; !ok {
		return nil, errors.New("persistence cfg datastores missing Cassandra")
	} else if store.NoSQL == nil {
		return nil, errors.New("NoSQL struct is nil")
	}

	client, err := newConfigStoreClient(clientCfg, persistenceCfg.DataStores[persistenceCfg.DefaultStore].NoSQL, logger, doneCh)
	if err != nil {
		return nil, err
	}
	err = client.startUpdate()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func newConfigStoreClient(clientCfg *csc.ClientConfig, persistenceCfg *config.NoSQL, logger log.Logger, doneCh chan struct{}) (*configStoreClient, error) {
	store, err := nosql.NewNoSQLConfigStore(*persistenceCfg, logger, nil)
	if err != nil {
		return nil, err
	}

	client := &configStoreClient{
		config:             clientCfg,
		doneCh:             doneCh,
		configStoreManager: persistence.NewConfigStoreManagerImpl(store, logger),
		logger:             logger,
	}

	return client, nil
}

func (csc *configStoreClient) startUpdate() error {
	if err := csc.update(); err != nil {
		return err
	}
	go func() {
		ticker := time.NewTicker(csc.config.PollInterval)
		for {
			select {
			case <-ticker.C:
				err := csc.update()
				if err != nil {
					csc.logger.Error("Failed to update dynamic config", tag.Error(err))
				}
			case <-csc.doneCh:
				ticker.Stop()
				return
			}
		}
	}()
	return nil
}

func (csc *configStoreClient) GetValue(name dc.Key) (interface{}, error) {
	return csc.getValueWithFilters(name, nil, name.DefaultValue())
}

func (csc *configStoreClient) GetValueWithFilters(name dc.Key, filters map[dc.Filter]interface{}) (interface{}, error) {
	return csc.getValueWithFilters(name, filters, name.DefaultValue())
}

func (csc *configStoreClient) GetIntValue(name dc.IntKey, filters map[dc.Filter]interface{}) (int, error) {
	defaultValue := name.DefaultInt()
	val, err := csc.getValueWithFilters(name, filters, defaultValue)
	if err != nil {
		return defaultValue, err
	}

	floatVal, ok := val.(float64)
	if !ok {
		return defaultValue, errors.New("value type is not int")
	}

	if floatVal != math.Trunc(floatVal) {
		return defaultValue, errors.New("value type is not int")
	}

	return int(floatVal), nil
}

func (csc *configStoreClient) GetFloatValue(name dc.FloatKey, filters map[dc.Filter]interface{}) (float64, error) {
	defaultValue := name.DefaultFloat()
	val, err := csc.getValueWithFilters(name, filters, defaultValue)
	if err != nil {
		return defaultValue, err
	}

	if floatVal, ok := val.(float64); ok {
		return floatVal, nil
	}
	return defaultValue, errors.New("value type is not float64")
}

func (csc *configStoreClient) GetBoolValue(name dc.BoolKey, filters map[dc.Filter]interface{}) (bool, error) {
	defaultValue := name.DefaultBool()
	val, err := csc.getValueWithFilters(name, filters, defaultValue)
	if err != nil {
		return defaultValue, err
	}

	if boolVal, ok := val.(bool); ok {
		return boolVal, nil
	}
	return defaultValue, errors.New("value type is not bool")
}

func (csc *configStoreClient) GetStringValue(name dc.StringKey, filters map[dc.Filter]interface{}) (string, error) {
	defaultValue := name.DefaultString()
	val, err := csc.getValueWithFilters(name, filters, defaultValue)
	if err != nil {
		return defaultValue, err
	}

	if stringVal, ok := val.(string); ok {
		return stringVal, nil
	}
	return defaultValue, errors.New("value type is not string")
}

// Note that all number types (ex: ints) will be returned as float64.
// It is the caller's responsibility to convert based on their context for value type.
func (csc *configStoreClient) GetMapValue(name dc.MapKey, filters map[dc.Filter]interface{}) (map[string]interface{}, error) {
	defaultValue := name.DefaultMap()
	val, err := csc.getValueWithFilters(name, filters, defaultValue)
	if err != nil {
		return defaultValue, err
	}
	if mapVal, ok := val.(map[string]interface{}); ok {
		return mapVal, nil
	}
	return defaultValue, errors.New("value type is not map")
}

func (csc *configStoreClient) GetDurationValue(name dc.DurationKey, filters map[dc.Filter]interface{}) (time.Duration, error) {
	defaultValue := name.DefaultDuration()
	val, err := csc.getValueWithFilters(name, filters, defaultValue)
	if err != nil {
		return defaultValue, err
	}

	var durVal time.Duration
	switch v := val.(type) {
	case string:
		durVal, err = time.ParseDuration(v)
		if err != nil {
			return defaultValue, errors.New("value string encoding cannot be parsed into duration")
		}
	case time.Duration:
		durVal = v
	default:
		return defaultValue, errors.New("value type is not duration")
	}

	return durVal, nil
}

func (csc *configStoreClient) UpdateValue(name dc.Key, value interface{}) error {
	dcValues, ok := value.([]*types.DynamicConfigValue)
	if !ok && value != nil {
		return errors.New("invalid value")
	}
	return csc.updateValue(name, dcValues, csc.config.UpdateRetryAttempts)
}

func (csc *configStoreClient) RestoreValue(name dc.Key, filters map[dc.Filter]interface{}) error {
	//if empty filter provided, update fallback value.
	//if u want to remove entire entry, just do update value with empty
	loaded := csc.values.Load()
	if loaded == nil {
		return dc.NotFoundError
	}
	currentCached := loaded.(cacheEntry)

	if currentCached.dcEntries == nil {
		return dc.NotFoundError
	}

	val, ok := currentCached.dcEntries[name.String()]
	if !ok {
		return dc.NotFoundError
	}

	newValues := make([]*types.DynamicConfigValue, 0, len(val.Values))
	if filters == nil {
		for _, dcValue := range val.Values {
			if dcValue.Filters != nil || len(dcValue.Filters) != 0 {
				newValues = append(newValues, copyDynamicConfigValue(dcValue))
			}
		}
	} else {
		for _, dcValue := range val.Values {
			if !matchFilters(dcValue, filters) || dcValue.Filters == nil || len(dcValue.Filters) == 0 {
				newValues = append(newValues, copyDynamicConfigValue(dcValue))
			}
		}
	}

	return csc.updateValue(name, newValues, csc.config.UpdateRetryAttempts)
}

func (csc *configStoreClient) ListValue(name dc.Key) ([]*types.DynamicConfigEntry, error) {
	var resList []*types.DynamicConfigEntry

	loaded := csc.values.Load()
	if loaded == nil {
		return nil, nil
	}
	currentCached := loaded.(cacheEntry)

	if currentCached.dcEntries == nil {
		return nil, nil
	}
	listAll := false
	if name == nil {
		//if key is not specified, return all entries
		listAll = true
	} else if _, ok := currentCached.dcEntries[name.String()]; !ok {
		//if key is not known, return all entries
		listAll = true
	}
	if listAll {
		//if key is not known/specified, return all entries
		resList = make([]*types.DynamicConfigEntry, 0, len(currentCached.dcEntries))
		for _, entry := range currentCached.dcEntries {
			resList = append(resList, copyDynamicConfigEntry(entry))
		}
	} else {
		//if key is known, return just that specific entry
		resList = make([]*types.DynamicConfigEntry, 0, 1)
		resList = append(resList, currentCached.dcEntries[name.String()])
	}

	return resList, nil
}

func (csc *configStoreClient) updateValue(name dc.Key, dcValues []*types.DynamicConfigValue, retryAttempts int) error {
	//since values are not unique, no way to know if you are trying to update a specific value
	//or if you want to add another of the same value with different filters.
	//UpdateValue will replace everything associated with dc key.
	for _, dcValue := range dcValues {
		if err := validateKeyDataBlobPair(name, dcValue.Value); err != nil {
			return err
		}
	}
	loaded := csc.values.Load()
	var currentCached cacheEntry
	if loaded == nil {
		currentCached = cacheEntry{
			cacheVersion:  0,
			schemaVersion: 0,
			dcEntries:     map[string]*types.DynamicConfigEntry{},
		}
	} else {
		currentCached = loaded.(cacheEntry)
	}

	keyName := name.String()
	var newEntries []*types.DynamicConfigEntry

	existingEntry, entryExists := currentCached.dcEntries[keyName]

	if dcValues == nil || len(dcValues) == 0 {
		newEntries = make([]*types.DynamicConfigEntry, 0, len(currentCached.dcEntries))

		for _, entry := range currentCached.dcEntries {
			if entryExists && entry == existingEntry {
				continue
			} else {
				newEntries = append(newEntries, copyDynamicConfigEntry(entry))
			}
		}
	} else {
		if entryExists {
			newEntries = make([]*types.DynamicConfigEntry, 0, len(currentCached.dcEntries))
		} else {
			newEntries = make([]*types.DynamicConfigEntry, 0, len(currentCached.dcEntries)+1)
			newEntries = append(newEntries,
				&types.DynamicConfigEntry{
					Name:   keyName,
					Values: dcValues,
				})
		}

		for _, entry := range currentCached.dcEntries {
			if entryExists && entry.Name == keyName {
				newEntries = append(newEntries,
					&types.DynamicConfigEntry{
						Name:   keyName,
						Values: dcValues,
					})
			} else {
				newEntries = append(newEntries, copyDynamicConfigEntry(entry))
			}
		}
	}

	newSnapshot := &persistence.DynamicConfigSnapshot{
		Version: currentCached.cacheVersion + 1,
		Values: &types.DynamicConfigBlob{
			SchemaVersion: currentCached.schemaVersion,
			Entries:       newEntries,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), csc.config.UpdateTimeout)
	defer cancel()

	err := csc.configStoreManager.UpdateDynamicConfig(
		ctx,
		&persistence.UpdateDynamicConfigRequest{
			Snapshot: newSnapshot,
		},
	)

	select {
	case <-ctx.Done():
		//potentially we can retry on timeout
		return errors.New("timeout error on update")
	default:
		if err != nil {
			if _, ok := err.(*persistence.ConditionFailedError); ok && retryAttempts > 0 {
				//fetch new config and retry
				err := csc.update()
				if err != nil {
					return err
				}
				return csc.updateValue(name, dcValues, retryAttempts-1)
			}

			if retryAttempts == 0 {
				return errors.New("ran out of retry attempts on update")
			}
			return err
		}
		return nil
	}
}

func copyDynamicConfigEntry(entry *types.DynamicConfigEntry) *types.DynamicConfigEntry {
	if entry == nil {
		return nil
	}

	newValues := make([]*types.DynamicConfigValue, 0, len(entry.Values))
	for _, value := range entry.Values {
		newValues = append(newValues, copyDynamicConfigValue(value))
	}

	return &types.DynamicConfigEntry{
		Name:   entry.Name,
		Values: newValues,
	}
}

func copyDynamicConfigValue(value *types.DynamicConfigValue) *types.DynamicConfigValue {
	if value == nil {
		return nil
	}

	var newFilters []*types.DynamicConfigFilter
	if value.Filters == nil {
		newFilters = nil
	} else {
		newFilters = make([]*types.DynamicConfigFilter, 0, len(value.Filters))
		for _, filter := range value.Filters {
			newFilters = append(newFilters, copyDynamicConfigFilter(filter))
		}
	}

	return &types.DynamicConfigValue{
		Value:   copyDataBlob(value.Value),
		Filters: newFilters,
	}
}

func copyDynamicConfigFilter(filter *types.DynamicConfigFilter) *types.DynamicConfigFilter {
	if filter == nil {
		return nil
	}

	return &types.DynamicConfigFilter{
		Name:  filter.Name,
		Value: copyDataBlob(filter.Value),
	}
}

func copyDataBlob(blob *types.DataBlob) *types.DataBlob {
	if blob == nil {
		return nil
	}

	newData := make([]byte, len(blob.Data))
	copy(newData, blob.Data)

	return &types.DataBlob{
		EncodingType: blob.EncodingType,
		Data:         newData,
	}
}

func (csc *configStoreClient) update() error {
	ctx, cancel := context.WithTimeout(context.Background(), csc.config.FetchTimeout)
	defer cancel()

	res, err := csc.configStoreManager.FetchDynamicConfig(ctx)

	select {
	case <-ctx.Done():
		return errors.New("timeout error on fetch")
	default:
		if err != nil {
			return fmt.Errorf("failed to fetch dynamic config snapshot %v", err)
		}

		if res != nil && res.Snapshot != nil {
			defer func() {
				csc.lastUpdatedTime = time.Now()
			}()

			return csc.storeValues(res.Snapshot)
		}
	}
	return nil
}

func (csc *configStoreClient) storeValues(snapshot *persistence.DynamicConfigSnapshot) error {
	//Converting the list of dynamic config entries into a map for better lookup performance
	var dcEntryMap map[string]*types.DynamicConfigEntry
	if snapshot.Values.Entries == nil {
		dcEntryMap = nil
	} else {
		dcEntryMap = make(map[string]*types.DynamicConfigEntry)
		for _, entry := range snapshot.Values.Entries {
			dcEntryMap[entry.Name] = entry
		}
	}

	csc.values.Store(cacheEntry{
		cacheVersion:  snapshot.Version,
		schemaVersion: snapshot.Values.SchemaVersion,
		dcEntries:     dcEntryMap,
	})
	csc.logger.Info("Updated dynamic config")
	return nil
}

func (csc *configStoreClient) getValueWithFilters(key dc.Key, filters map[dc.Filter]interface{}, defaultValue interface{}) (interface{}, error) {
	keyName := key.String()
	loaded := csc.values.Load()
	if loaded == nil {
		return defaultValue, nil
	}
	cached := loaded.(cacheEntry)
	found := false

	if entry, ok := cached.dcEntries[keyName]; ok && entry != nil {
		for _, dcValue := range entry.Values {
			if len(dcValue.Filters) == 0 {
				parsedVal, err := convertFromDataBlob(dcValue.Value)

				if err == nil {
					defaultValue = parsedVal
					found = true
				}
				continue
			}

			if matchFilters(dcValue, filters) {
				return convertFromDataBlob(dcValue.Value)
			}
		}
	}
	if found {
		return defaultValue, nil
	}
	return defaultValue, dc.NotFoundError
}

func matchFilters(dcValue *types.DynamicConfigValue, filters map[dc.Filter]interface{}) bool {
	if len(dcValue.Filters) > len(filters) {
		return false
	}

	for _, valueFilter := range dcValue.Filters {
		filterKey := dc.ParseFilter(valueFilter.Name)
		if filters[filterKey] == nil {
			return false
		}

		requestValue, err := convertFromDataBlob(valueFilter.Value)
		if err != nil || filters[filterKey] != requestValue {
			return false
		}
	}
	return true
}

func validateClientConfig(config *csc.ClientConfig) error {
	if config == nil {
		return errors.New("no config found for config store based dynamic config client")
	}
	if config.PollInterval < configStoreMinPollInterval {
		return fmt.Errorf("poll interval should be at least %v", configStoreMinPollInterval)
	}
	if config.UpdateRetryAttempts < 0 {
		return errors.New("UpdateRetryAttempts must be non-negative")
	}
	if config.FetchTimeout <= 0 {
		return errors.New("FetchTimeout must be positive")
	}
	if config.UpdateTimeout <= 0 {
		return errors.New("UpdateTimeout must be positive")
	}
	return nil
}

func convertFromDataBlob(blob *types.DataBlob) (interface{}, error) {
	switch *blob.EncodingType {
	case types.EncodingTypeJSON:
		var v interface{}
		err := json.Unmarshal(blob.Data, &v)
		return v, err
	default:
		return nil, errors.New("unsupported blob encoding")
	}
}

func validateKeyDataBlobPair(key dc.Key, blob *types.DataBlob) error {
	value, err := convertFromDataBlob(blob)
	if err != nil {
		return err
	}
	err = fmt.Errorf("key value pair mismatch, key type: %T, value type: %T", key, value)
	switch key.(type) {
	case dc.IntKey:
		if _, ok := value.(int); !ok {
			floatVal, ok := value.(float64)
			if !ok { // int can be decoded as float64
				return err
			}
			if floatVal != math.Trunc(floatVal) {
				return errors.New("value type is not int")
			}
		}
	case dc.BoolKey:
		if _, ok := value.(bool); !ok {
			return err
		}
	case dc.FloatKey:
		if _, ok := value.(float64); !ok {
			return err
		}
	case dc.StringKey:
		if _, ok := value.(string); !ok {
			return err
		}
	case dc.DurationKey:
		if _, ok := value.(time.Duration); !ok {
			durationStr, ok := value.(string)
			if !ok {
				return err
			}
			if _, err = time.ParseDuration(durationStr); err != nil {
				return errors.New("value string encoding cannot be parsed into duration")
			}
		}
	case dc.MapKey:
		if _, ok := value.(map[string]interface{}); !ok {
			return err
		}
	default:
		return fmt.Errorf("unknown key type: %T", key)
	}
	return nil
}
