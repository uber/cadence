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

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination clientInterface_mock.go -self_package github.com/uber/cadence/common/dynamicconfig

package dynamicconfig

import (
	"time"

	"github.com/uber/cadence/common/types"
)

const (
	DynamicConfigConfigStoreClient = "configstore"
	DynamicConfigFileBasedClient   = "filebased"
	DynamicConfigInMemoryClient    = "memory"
	DynamicConfigNopClient         = "nop"
)

// Client allows fetching values from a dynamic configuration system NOTE: This does not have async
// options right now. In the interest of keeping it minimal, we can add when requirement arises.
type Client interface {
	// GetValue get the default value of a dynamic config entry(without filter), return sysDefaultValue if no config is set
	// name is the key to identify a dynamic config entry
	GetValue(name Key, sysDefaultValue interface{}) (interface{}, error)
	// GetValueWithFilters get the actual value of a dynamic config entry under certain condition using filter, return sysDefaultValue if no config is set
	// name is the key to identify a dynamic config entry
	GetValueWithFilters(name Key, filters map[Filter]interface{}, sysDefaultValue interface{}) (interface{}, error)

	// GetIntValue is the strongly typed form of GetValueWithFilters
	// name is the key to identify a dynamic config entry
	GetIntValue(name Key, filters map[Filter]interface{}, sysDefaultValue int) (int, error)
	// GetFloatValue is the strongly typed form of GetValueWithFilters
	// name is the key to identify a dynamic config entry
	GetFloatValue(name Key, filters map[Filter]interface{}, sysDefaultValue float64) (float64, error)
	// GetBoolValue is the strongly typed form of GetValueWithFilters
	// name is the key to identify a dynamic config entry
	GetBoolValue(name Key, filters map[Filter]interface{}, sysDefaultValue bool) (bool, error)
	// GetStringValue is the strongly typed form of GetValueWithFilters
	// name is the key to identify a dynamic config entry
	GetStringValue(name Key, filters map[Filter]interface{}, sysDefaultValue string) (string, error)
	// GetMapValue is the strongly typed form of GetValueWithFilters
	// name is the key to identify a dynamic config entry
	GetMapValue(
		name Key, filters map[Filter]interface{}, sysDefaultValue map[string]interface{},
	) (map[string]interface{}, error)
	// GetDurationValue is the strongly typed form of GetValueWithFilters
	// name is the key to identify a dynamic config entry
	GetDurationValue(
		name Key, filters map[Filter]interface{}, sysDefaultValue time.Duration,
	) (time.Duration, error)

	// UpdateValue update all the values(all the filters) of the dynamic config entry
	// name is the key to identify a dynamic config entry
	UpdateFallbackRawValue(name Key, value interface{}) error
	// RestoreValue is a shortcut or special case of UpdateValue -- delete some filters from the dynamic config value in a safer way
	// When filters is nil, it will delete the value with empty filters -- which is the fallback value. So that the fallback value becomes the system default value(defined in code)
	// When filters is not nil, it will delete the values with the matched filters
	// name is the key to identify a dynamic config entry
	RestoreValues(name Key, filters map[Filter]interface{}) error

	// ListConfigEntries returns all the existing dynamic config entries
	ListConfigEntries() ([]*types.DynamicConfigEntry, error)
}

var NotFoundError = &types.EntityNotExistsError{
	Message: "unable to find key",
}
