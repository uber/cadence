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

package types

type DynamicConfigBlob struct {
	SchemaVersion int64                 `json:"schemaVersion,omitempty"`
	Entries       []*DynamicConfigEntry `json:"entries,omitempty"`
}

func (v *DynamicConfigBlob) GetSchemaVersion() (o int64) {
	if v != nil {
		return v.SchemaVersion
	}
	return
}

func (v *DynamicConfigBlob) GetEntries() (o []*DynamicConfigEntry) {
	if v != nil && v.Entries != nil {
		return v.Entries
	}
	return
}

type DynamicConfigEntry struct {
	Name   string                `json:"name,omitempty"`
	Values []*DynamicConfigValue `json:"values,omitempty"`
}

func (v *DynamicConfigEntry) GetName() (o string) {
	if v != nil {
		return v.Name
	}
	return
}

func (v *DynamicConfigEntry) GetValues() (o []*DynamicConfigValue) {
	if v != nil && v.Values != nil {
		return v.Values
	}
	return
}

type DynamicConfigValue struct {
	Value   *DataBlob              `json:"value,omitempty"`
	Filters []*DynamicConfigFilter `json:"filters,omitempty"`
}

func (v *DynamicConfigValue) GetValue() (o *DataBlob) {
	if v != nil {
		return v.Value
	}
	return
}

func (v *DynamicConfigValue) GetFilters() (o []*DynamicConfigFilter) {
	if v != nil && v.Filters != nil {
		return v.Filters
	}
	return
}

type DynamicConfigFilter struct {
	Name  string    `json:"name,omitempty"`
	Value *DataBlob `json:"value,omitempty"`
}

func (v *DynamicConfigFilter) GetName() (o string) {
	if v != nil {
		return v.Name
	}
	return
}

func (v *DynamicConfigFilter) GetValue() (o *DataBlob) {
	if v != nil {
		return v.Value
	}
	return
}
