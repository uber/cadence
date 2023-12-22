// Copyright (c) 2017-2020 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package thrift

import (
	"github.com/uber/cadence/.gen/go/config"
	"github.com/uber/cadence/common/types"
)

// FromDynamicConfigBlob converts internal DynamicConfigBlob type to thrift
func FromDynamicConfigBlob(t *types.DynamicConfigBlob) *config.DynamicConfigBlob {
	if t == nil {
		return nil
	}
	return &config.DynamicConfigBlob{
		SchemaVersion: &t.SchemaVersion,
		Entries:       FromDynamicConfigEntryArray(t.Entries),
	}
}

// ToDynamicConfigBlob converts thrift DynamicConfigBlob type to internal
func ToDynamicConfigBlob(t *config.DynamicConfigBlob) *types.DynamicConfigBlob {
	if t == nil {
		return nil
	}
	return &types.DynamicConfigBlob{
		SchemaVersion: t.GetSchemaVersion(),
		Entries:       ToDynamicConfigEntryArray(t.Entries),
	}
}

// FromDynamicConfigEntryArray converts internal DynamicConfigEntry array type to thrift
func FromDynamicConfigEntryArray(t []*types.DynamicConfigEntry) []*config.DynamicConfigEntry {
	if t == nil {
		return nil
	}
	v := make([]*config.DynamicConfigEntry, len(t))
	for i := range t {
		v[i] = FromDynamicConfigEntry(t[i])
	}
	return v
}

// ToDynamicConfigEntryArray converts thrift DynamicConfigEntry array type to internal
func ToDynamicConfigEntryArray(t []*config.DynamicConfigEntry) []*types.DynamicConfigEntry {
	if t == nil {
		return nil
	}
	v := make([]*types.DynamicConfigEntry, len(t))
	for i := range t {
		v[i] = ToDynamicConfigEntry(t[i])
	}
	return v
}

// FromDynamicConfigEntry converts internal DynamicConfigEntry type to thrift
func FromDynamicConfigEntry(t *types.DynamicConfigEntry) *config.DynamicConfigEntry {
	if t == nil {
		return nil
	}
	return &config.DynamicConfigEntry{
		Name:   &t.Name,
		Values: FromDynamicConfigValueArray(t.Values),
	}
}

// ToDynamicConfigEntry converts thrift DynamicConfigEntry type to internal
func ToDynamicConfigEntry(t *config.DynamicConfigEntry) *types.DynamicConfigEntry {
	if t == nil {
		return nil
	}
	return &types.DynamicConfigEntry{
		Name:   t.GetName(),
		Values: ToDynamicConfigValueArray(t.Values),
	}
}

// FromDynamicConfigValueArray converts internal DynamicConfigValue array type to thrift
func FromDynamicConfigValueArray(t []*types.DynamicConfigValue) []*config.DynamicConfigValue {
	if t == nil {
		return nil
	}
	v := make([]*config.DynamicConfigValue, len(t))
	for i := range t {
		v[i] = FromDynamicConfigValue(t[i])
	}
	return v
}

// ToDynamicConfigValueArray converts thrift DynamicConfigValue array type to internal
func ToDynamicConfigValueArray(t []*config.DynamicConfigValue) []*types.DynamicConfigValue {
	if t == nil {
		return nil
	}
	v := make([]*types.DynamicConfigValue, len(t))
	for i := range t {
		v[i] = ToDynamicConfigValue(t[i])
	}
	return v
}

// FromDynamicConfigValue converts internal DynamicConfigValue type to thrift
func FromDynamicConfigValue(t *types.DynamicConfigValue) *config.DynamicConfigValue {
	if t == nil {
		return nil
	}
	return &config.DynamicConfigValue{
		Value:   FromDataBlob(t.Value),
		Filters: FromDynamicConfigFilterArray(t.Filters),
	}
}

// ToDynamicConfigValue converts thrift DynamicConfigValue type to internal
func ToDynamicConfigValue(t *config.DynamicConfigValue) *types.DynamicConfigValue {
	if t == nil {
		return nil
	}
	return &types.DynamicConfigValue{
		Value:   ToDataBlob(t.Value),
		Filters: ToDynamicConfigFilterArray(t.Filters),
	}
}

// FromDynamicConfigFilterArray converts internal DynamicConfigFilter array type to thrift
func FromDynamicConfigFilterArray(t []*types.DynamicConfigFilter) []*config.DynamicConfigFilter {
	if t == nil {
		return nil
	}
	v := make([]*config.DynamicConfigFilter, len(t))
	for i := range t {
		v[i] = FromDynamicConfigFilter(t[i])
	}
	return v
}

// ToDynamicConfigFilterArray converts thrift DynamicConfigFilter array type to internal
func ToDynamicConfigFilterArray(t []*config.DynamicConfigFilter) []*types.DynamicConfigFilter {
	if t == nil {
		return nil
	}
	v := make([]*types.DynamicConfigFilter, len(t))
	for i := range t {
		v[i] = ToDynamicConfigFilter(t[i])
	}
	return v
}

// FromDynamicConfigFilter converts internal DynamicConfigFilter type to thrift
func FromDynamicConfigFilter(t *types.DynamicConfigFilter) *config.DynamicConfigFilter {
	if t == nil {
		return nil
	}
	return &config.DynamicConfigFilter{
		Name:  &t.Name,
		Value: FromDataBlob(t.Value),
	}
}

// ToDynamicConfigFilter converts thrift DynamicConfigFilter type to internal
func ToDynamicConfigFilter(t *config.DynamicConfigFilter) *types.DynamicConfigFilter {
	if t == nil {
		return nil
	}
	return &types.DynamicConfigFilter{
		Name:  t.GetName(),
		Value: ToDataBlob(t.Value),
	}
}
