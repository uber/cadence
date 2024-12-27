// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

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

// Code generated by MockGen. DO NOT EDIT.
// Source: partition_config_provider.go
//
// Generated by this command:
//
//	mockgen -package matching -source partition_config_provider.go -destination partition_config_provider_mock.go -package matching github.com/uber/cadence/client/matching PartitionConfigProvider
//

// Package matching is a generated GoMock package.
package matching

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"

	types "github.com/uber/cadence/common/types"
)

// MockPartitionConfigProvider is a mock of PartitionConfigProvider interface.
type MockPartitionConfigProvider struct {
	ctrl     *gomock.Controller
	recorder *MockPartitionConfigProviderMockRecorder
	isgomock struct{}
}

// MockPartitionConfigProviderMockRecorder is the mock recorder for MockPartitionConfigProvider.
type MockPartitionConfigProviderMockRecorder struct {
	mock *MockPartitionConfigProvider
}

// NewMockPartitionConfigProvider creates a new mock instance.
func NewMockPartitionConfigProvider(ctrl *gomock.Controller) *MockPartitionConfigProvider {
	mock := &MockPartitionConfigProvider{ctrl: ctrl}
	mock.recorder = &MockPartitionConfigProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPartitionConfigProvider) EXPECT() *MockPartitionConfigProviderMockRecorder {
	return m.recorder
}

// GetNumberOfReadPartitions mocks base method.
func (m *MockPartitionConfigProvider) GetNumberOfReadPartitions(domainID string, taskList types.TaskList, taskListType int) int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNumberOfReadPartitions", domainID, taskList, taskListType)
	ret0, _ := ret[0].(int)
	return ret0
}

// GetNumberOfReadPartitions indicates an expected call of GetNumberOfReadPartitions.
func (mr *MockPartitionConfigProviderMockRecorder) GetNumberOfReadPartitions(domainID, taskList, taskListType any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNumberOfReadPartitions", reflect.TypeOf((*MockPartitionConfigProvider)(nil).GetNumberOfReadPartitions), domainID, taskList, taskListType)
}

// GetNumberOfWritePartitions mocks base method.
func (m *MockPartitionConfigProvider) GetNumberOfWritePartitions(domainID string, taskList types.TaskList, taskListType int) int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNumberOfWritePartitions", domainID, taskList, taskListType)
	ret0, _ := ret[0].(int)
	return ret0
}

// GetNumberOfWritePartitions indicates an expected call of GetNumberOfWritePartitions.
func (mr *MockPartitionConfigProviderMockRecorder) GetNumberOfWritePartitions(domainID, taskList, taskListType any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNumberOfWritePartitions", reflect.TypeOf((*MockPartitionConfigProvider)(nil).GetNumberOfWritePartitions), domainID, taskList, taskListType)
}

// UpdatePartitionConfig mocks base method.
func (m *MockPartitionConfigProvider) UpdatePartitionConfig(domainID string, taskList types.TaskList, taskListType int, config *types.TaskListPartitionConfig) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "UpdatePartitionConfig", domainID, taskList, taskListType, config)
}

// UpdatePartitionConfig indicates an expected call of UpdatePartitionConfig.
func (mr *MockPartitionConfigProviderMockRecorder) UpdatePartitionConfig(domainID, taskList, taskListType, config any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePartitionConfig", reflect.TypeOf((*MockPartitionConfigProvider)(nil).UpdatePartitionConfig), domainID, taskList, taskListType, config)
}
