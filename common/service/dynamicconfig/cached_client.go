// Copyright (c) 2017-2020 Uber Technologies, Inc.
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

package dynamicconfig

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/uber/cadence/common/types"
)

type (
	getValueFunc func(Key, map[Filter]interface{}, interface{}) (interface{}, error)

	cachedKey struct {
		name      Key
		filterStr string
	}

	cachedValue struct {
		name         Key
		filters      map[Filter]interface{}
		getValueFunc getValueFunc
		value        interface{}
		err          error
	}

	cachedClient struct {
		sync.RWMutex

		client Client
		values map[cachedKey]cachedValue
	}
)

const (
	// DefaultRefreshInterval is the default refresh interval for cached client
	DefaultRefreshInterval = 30 * time.Second
)

// NewCachedClient creates a new cached dynamic config client
func NewCachedClient(
	client Client,
	updateInterval time.Duration,
	doneCh chan struct{},
) Client {
	cachedClient := &cachedClient{
		client: client,
		values: make(map[cachedKey]cachedValue),
	}

	go cachedClient.refreshLoop(updateInterval, doneCh)

	return cachedClient
}

func (c *cachedClient) GetValue(
	name Key,
	defaultValue interface{},
) (interface{}, error) {
	return c.getValueWithFilters(
		name,
		nil,
		defaultValue,
		c.getValueFunc,
	)
}

func (c *cachedClient) GetValueWithFilters(
	name Key,
	filters map[Filter]interface{},
	defaultValue interface{},
) (interface{}, error) {
	return c.getValueWithFilters(
		name,
		filters,
		defaultValue,
		c.client.GetValueWithFilters,
	)
}

func (c *cachedClient) GetIntValue(
	name Key,
	filters map[Filter]interface{},
	defaultValue int,
) (int, error) {
	value, err := c.getValueWithFilters(
		name,
		filters,
		defaultValue,
		c.getIntValueFunc,
	)

	if err != nil {
		return defaultValue, err
	}

	if intValue, ok := value.(int); ok {
		return intValue, nil
	}
	return defaultValue, fmt.Errorf("unable to convert value of type %T to int", value)
}

func (c *cachedClient) GetFloatValue(
	name Key,
	filters map[Filter]interface{},
	defaultValue float64,
) (float64, error) {
	value, err := c.getValueWithFilters(
		name,
		filters,
		defaultValue,
		c.getFloatValueFunc,
	)

	if err != nil {
		return defaultValue, err
	}

	if floatValue, ok := value.(float64); ok {
		return floatValue, nil
	} else if intVal, ok := value.(int); ok {
		return float64(intVal), nil
	}
	return defaultValue, fmt.Errorf("unable to convert value of type %T to float64", value)
}

func (c *cachedClient) GetBoolValue(
	name Key,
	filters map[Filter]interface{},
	defaultValue bool,
) (bool, error) {
	value, err := c.getValueWithFilters(
		name,
		filters,
		defaultValue,
		c.getBoolValueFunc,
	)

	if err != nil {
		return defaultValue, err
	}

	if boolVal, ok := value.(bool); ok {
		return boolVal, nil
	}
	return defaultValue, fmt.Errorf("unable to convert value of type %T to bool", value)
}

func (c *cachedClient) GetStringValue(
	name Key,
	filters map[Filter]interface{},
	defaultValue string,
) (string, error) {
	value, err := c.getValueWithFilters(
		name,
		filters,
		defaultValue,
		c.getStringValueFunc,
	)

	if err != nil {
		return defaultValue, err
	}

	if stringVal, ok := value.(string); ok {
		return stringVal, nil
	}
	return defaultValue, fmt.Errorf("unable to convert value of type %T to string", value)
}

func (c *cachedClient) GetMapValue(
	name Key,
	filters map[Filter]interface{},
	defaultValue map[string]interface{},
) (map[string]interface{}, error) {
	value, err := c.getValueWithFilters(
		name,
		filters,
		defaultValue,
		c.getMapValueFunc,
	)

	if err != nil {
		return defaultValue, err
	}

	if mapVal, ok := value.(map[string]interface{}); ok {
		return mapVal, nil
	}
	return defaultValue, fmt.Errorf("unable to convert value of type %T to map", value)
}

func (c *cachedClient) GetDurationValue(
	name Key,
	filters map[Filter]interface{},
	defaultValue time.Duration,
) (time.Duration, error) {
	value, err := c.getValueWithFilters(
		name,
		filters,
		defaultValue,
		c.getDurationValueFunc,
	)

	if err != nil {
		return defaultValue, err
	}

	if durationVal, ok := value.(time.Duration); ok {
		return durationVal, nil
	}
	if durationString, ok := value.(string); ok {
		if durationVal, err := time.ParseDuration(durationString); err == nil {
			return durationVal, nil
		}
	}
	return defaultValue, fmt.Errorf("unable to convert value of type %T to duration", value)
}

func (c *cachedClient) UpdateValue(
	name Key,
	value interface{},
) error {
	if err := c.client.UpdateValue(name, value); err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	c.values[newCachedKey(name, nil)] = cachedValue{
		name:         name,
		getValueFunc: c.getValueFunc,
		value:        value,
	}
	return nil
}

func (c *cachedClient) refreshLoop(
	updateInterval time.Duration,
	doneCh chan struct{},
) {
	ticker := time.NewTicker(updateInterval)
	for {
		select {
		case <-ticker.C:
			c.refresh()
		case <-doneCh:
			ticker.Stop()
			return
		}
	}
}

func (c *cachedClient) refresh() {
	c.RLock()
	newValues := make(map[cachedKey]cachedValue)
	for key, value := range c.values {
		newValues[key] = value
	}
	c.RUnlock()

	for key, value := range newValues {
		newValue, err := value.getValueFunc(value.name, value.filters, value.value)
		switch err.(type) {
		case nil, *types.EntityNotExistsError:
			c.values[key] = cachedValue{
				name:         value.name,
				filters:      value.filters,
				getValueFunc: value.getValueFunc,
				value:        newValue,
				err:          err,
			}
		default:
			delete(newValues, key)
		}
	}

	c.Lock()
	defer c.Unlock()

	c.values = newValues
}

func (c *cachedClient) getValueWithFilters(
	name Key,
	filters map[Filter]interface{},
	defaultValue interface{},
	getValueFunc getValueFunc,
) (interface{}, error) {
	constrainedKey := newCachedKey(name, filters)

	c.RLock()
	if cacheHit, value, err := c.loadValueLocked(constrainedKey); cacheHit {
		c.RUnlock()
		return value, err
	}
	c.RUnlock()

	// value not in cache, read from the underlying client
	c.Lock()
	defer c.Unlock()

	if cacheHit, value, err := c.loadValueLocked(constrainedKey); cacheHit {
		return value, err
	}

	value, err := getValueFunc(name, filters, defaultValue)
	c.storeValueLocked(constrainedKey, filters, getValueFunc, value, err)

	return value, err
}

func (c *cachedClient) loadValueLocked(
	key cachedKey,
) (bool, interface{}, error) {
	cachedValue, ok := c.values[key]
	if !ok {
		return false, nil, nil
	}

	return true, cachedValue.value, cachedValue.err
}

func (c *cachedClient) storeValueLocked(
	key cachedKey,
	filters map[Filter]interface{},
	getValueFunc getValueFunc,
	value interface{},
	err error,
) {
	switch err.(type) {
	case nil, *types.EntityNotExistsError:
		c.values[key] = cachedValue{
			name:         key.name,
			filters:      filters,
			getValueFunc: getValueFunc,
			value:        value,
			err:          err,
		}
	}
}

func (c *cachedClient) getValueFunc(
	name Key,
	_ map[Filter]interface{},
	defaultValue interface{},
) (interface{}, error) {
	return c.client.GetValue(name, defaultValue)
}

func (c *cachedClient) getIntValueFunc(
	name Key,
	filters map[Filter]interface{},
	defaultValue interface{},
) (interface{}, error) {
	return c.client.GetIntValue(name, filters, defaultValue.(int))
}

func (c *cachedClient) getFloatValueFunc(
	name Key,
	filters map[Filter]interface{},
	defaultValue interface{},
) (interface{}, error) {
	return c.client.GetFloatValue(name, filters, defaultValue.(float64))
}

func (c *cachedClient) getBoolValueFunc(
	name Key,
	filters map[Filter]interface{},
	defaultValue interface{},
) (interface{}, error) {
	return c.client.GetBoolValue(name, filters, defaultValue.(bool))
}

func (c *cachedClient) getStringValueFunc(
	name Key,
	filters map[Filter]interface{},
	defaultValue interface{},
) (interface{}, error) {
	return c.client.GetStringValue(name, filters, defaultValue.(string))
}

func (c *cachedClient) getMapValueFunc(
	name Key,
	filters map[Filter]interface{},
	defaultValue interface{},
) (interface{}, error) {
	return c.client.GetMapValue(name, filters, defaultValue.(map[string]interface{}))
}

func (c *cachedClient) getDurationValueFunc(
	name Key,
	filters map[Filter]interface{},
	defaultValue interface{},
) (interface{}, error) {
	return c.client.GetDurationValue(name, filters, defaultValue.(time.Duration))
}

func newCachedKey(
	name Key,
	filters map[Filter]interface{},
) cachedKey {
	var filterStrBuilder strings.Builder
	for filter, value := range filters {
		filterStrBuilder.WriteString(fmt.Sprintf("%v:%v;", filter.String(), value))
	}
	return cachedKey{
		name:      name,
		filterStr: filterStrBuilder.String(),
	}
}
