// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package dynamicconfig

import (
	"errors"
	"sync"
	"time"

	"github.com/uber/cadence/common/types"
)

type inMemoryClient struct {
	sync.RWMutex

	globalValues map[Key]interface{}
}

// NewInMemoryClient creates a new in memory dynamic config client for testing purpose
func NewInMemoryClient() Client {
	return &inMemoryClient{
		globalValues: make(map[Key]interface{}),
	}
}

func (mc *inMemoryClient) SetValue(key Key, value interface{}) {
	mc.Lock()
	defer mc.Unlock()

	mc.globalValues[key] = value
}

func (mc *inMemoryClient) GetValue(key Key, sysDefaultValue interface{}) (interface{}, error) {
	mc.RLock()
	defer mc.RUnlock()

	if val, ok := mc.globalValues[key]; ok {
		return val, nil
	}
	return sysDefaultValue, NotFoundError
}

func (mc *inMemoryClient) GetValueWithFilters(
	name Key, filters map[Filter]interface{}, sysDefaultValue interface{},
) (interface{}, error) {
	mc.RLock()
	defer mc.RUnlock()

	return mc.GetValue(name, sysDefaultValue)
}

func (mc *inMemoryClient) GetIntValue(name Key, filters map[Filter]interface{}, sysDefaultValue int) (int, error) {
	mc.RLock()
	defer mc.RUnlock()

	if val, ok := mc.globalValues[name]; ok {
		return val.(int), nil
	}
	return sysDefaultValue, NotFoundError
}

func (mc *inMemoryClient) GetFloatValue(name Key, filters map[Filter]interface{}, sysDefaultValue float64) (float64, error) {
	mc.RLock()
	defer mc.RUnlock()

	if val, ok := mc.globalValues[name]; ok {
		return val.(float64), nil
	}
	return sysDefaultValue, NotFoundError
}

func (mc *inMemoryClient) GetBoolValue(name Key, filters map[Filter]interface{}, sysDefaultValue bool) (bool, error) {
	mc.RLock()
	defer mc.RUnlock()

	if val, ok := mc.globalValues[name]; ok {
		return val.(bool), nil
	}
	return sysDefaultValue, NotFoundError
}

func (mc *inMemoryClient) GetStringValue(name Key, filters map[Filter]interface{}, sysDefaultValue string) (string, error) {
	mc.RLock()
	defer mc.RUnlock()

	if val, ok := mc.globalValues[name]; ok {
		return val.(string), nil
	}
	return sysDefaultValue, NotFoundError
}

func (mc *inMemoryClient) GetMapValue(
	name Key, filters map[Filter]interface{}, sysDefaultValue map[string]interface{},
) (map[string]interface{}, error) {
	mc.RLock()
	defer mc.RUnlock()

	if val, ok := mc.globalValues[name]; ok {
		return val.(map[string]interface{}), nil
	}
	return sysDefaultValue, NotFoundError
}

func (mc *inMemoryClient) GetDurationValue(
	name Key, filters map[Filter]interface{}, sysDefaultValue time.Duration,
) (time.Duration, error) {
	mc.RLock()
	defer mc.RUnlock()

	if val, ok := mc.globalValues[name]; ok {
		return val.(time.Duration), nil
	}
	return sysDefaultValue, NotFoundError
}

func (mc *inMemoryClient) UpdateValue(key Key, value interface{}) error {
	mc.SetValue(key, value)
	return nil
}

func (mc *inMemoryClient) RestoreValue(name Key, filters map[Filter]interface{}) error {
	return errors.New("not supported for file based client")
}

func (mc *inMemoryClient) ListValue(name Key) ([]*types.DynamicConfigEntry, error) {
	return nil, errors.New("not supported for file based client")
}
