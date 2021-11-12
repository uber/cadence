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

package dynamicconfig

import (
	"reflect"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/uber/cadence/common/types"
)

// MockClient is a mock of Client interface
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// GetValue mocks base method
func (m *MockClient) GetValue(name Key, sysDefaultValue interface{}) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValue", name, sysDefaultValue)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetValue indicates an expected call of GetValue
func (mr *MockClientMockRecorder) GetValue(name, sysDefaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValue", reflect.TypeOf((*MockClient)(nil).GetValue), name, sysDefaultValue)
}

// GetValueWithFilters mocks base method
func (m *MockClient) GetValueWithFilters(name Key, filters map[Filter]interface{}, sysDefaultValue interface{}) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValueWithFilters", name, filters, sysDefaultValue)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetValueWithFilters indicates an expected call of GetValueWithFilters
func (mr *MockClientMockRecorder) GetValueWithFilters(name, filters, sysDefaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValueWithFilters", reflect.TypeOf((*MockClient)(nil).GetValueWithFilters), name, filters, sysDefaultValue)
}

// GetIntValue mocks base method
func (m *MockClient) GetIntValue(name Key, filters map[Filter]interface{}, sysDefaultValue int) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetIntValue", name, filters, sysDefaultValue)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetIntValue indicates an expected call of GetIntValue
func (mr *MockClientMockRecorder) GetIntValue(name, filters, sysDefaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIntValue", reflect.TypeOf((*MockClient)(nil).GetIntValue), name, filters, sysDefaultValue)
}

// GetFloatValue mocks base method
func (m *MockClient) GetFloatValue(name Key, filters map[Filter]interface{}, sysDefaultValue float64) (float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFloatValue", name, filters, sysDefaultValue)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFloatValue indicates an expected call of GetFloatValue
func (mr *MockClientMockRecorder) GetFloatValue(name, filters, sysDefaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFloatValue", reflect.TypeOf((*MockClient)(nil).GetFloatValue), name, filters, sysDefaultValue)
}

// GetBoolValue mocks base method
func (m *MockClient) GetBoolValue(name Key, filters map[Filter]interface{}, sysDefaultValue bool) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBoolValue", name, filters, sysDefaultValue)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBoolValue indicates an expected call of GetBoolValue
func (mr *MockClientMockRecorder) GetBoolValue(name, filters, sysDefaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBoolValue", reflect.TypeOf((*MockClient)(nil).GetBoolValue), name, filters, sysDefaultValue)
}

// GetStringValue mocks base method
func (m *MockClient) GetStringValue(name Key, filters map[Filter]interface{}, sysDefaultValue string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStringValue", name, filters, sysDefaultValue)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStringValue indicates an expected call of GetStringValue
func (mr *MockClientMockRecorder) GetStringValue(name, filters, sysDefaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStringValue", reflect.TypeOf((*MockClient)(nil).GetStringValue), name, filters, sysDefaultValue)
}

// GetMapValue mocks base method
func (m *MockClient) GetMapValue(name Key, filters map[Filter]interface{}, sysDefaultValue map[string]interface{}) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMapValue", name, filters, sysDefaultValue)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMapValue indicates an expected call of GetMapValue
func (mr *MockClientMockRecorder) GetMapValue(name, filters, sysDefaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMapValue", reflect.TypeOf((*MockClient)(nil).GetMapValue), name, filters, sysDefaultValue)
}

// GetDurationValue mocks base method
func (m *MockClient) GetDurationValue(name Key, filters map[Filter]interface{}, sysDefaultValue time.Duration) (time.Duration, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDurationValue", name, filters, sysDefaultValue)
	ret0, _ := ret[0].(time.Duration)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDurationValue indicates an expected call of GetDurationValue
func (mr *MockClientMockRecorder) GetDurationValue(name, filters, sysDefaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDurationValue", reflect.TypeOf((*MockClient)(nil).GetDurationValue), name, filters, sysDefaultValue)
}

// UpdateValues mocks base method
func (m *MockClient) UpdateValues(name Key, values []*types.DynamicConfigValue) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateValues", name, values)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateValues indicates an expected call of UpdateValues
func (mr *MockClientMockRecorder) UpdateValues(name, values interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateValues", reflect.TypeOf((*MockClient)(nil).UpdateValues), name, values)
}

// UpdateFallbackRawValue mocks base method
func (m *MockClient) UpdateFallbackRawValue(name Key, value interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateFallbackRawValue", name, value)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateFallbackRawValue indicates an expected call of UpdateFallbackRawValue
func (mr *MockClientMockRecorder) UpdateFallbackRawValue(name, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateFallbackRawValue", reflect.TypeOf((*MockClient)(nil).UpdateFallbackRawValue), name, value)
}

// RestoreValues mocks base method
func (m *MockClient) RestoreValues(name Key, filters map[Filter]interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RestoreValues", name, filters)
	ret0, _ := ret[0].(error)
	return ret0
}

// RestoreValues indicates an expected call of RestoreValues
func (mr *MockClientMockRecorder) RestoreValues(name, filters interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RestoreValues", reflect.TypeOf((*MockClient)(nil).RestoreValues), name, filters)
}

// ListConfigEntries mocks base method
func (m *MockClient) ListConfigEntries() ([]*types.DynamicConfigEntry, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListConfigEntries")
	ret0, _ := ret[0].([]*types.DynamicConfigEntry)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListConfigEntries indicates an expected call of ListConfigEntries
func (mr *MockClientMockRecorder) ListConfigEntries() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListConfigEntries", reflect.TypeOf((*MockClient)(nil).ListConfigEntries))
}
