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
// Source: interface.go

// Package isolationgroup is a generated GoMock package.
package isolationgroup

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"

	types "github.com/uber/cadence/common/types"
)

// MockController is a mock of Controller interface.
type MockController struct {
	ctrl     *gomock.Controller
	recorder *MockControllerMockRecorder
}

// MockControllerMockRecorder is the mock recorder for MockController.
type MockControllerMockRecorder struct {
	mock *MockController
}

// NewMockController creates a new mock instance.
func NewMockController(ctrl *gomock.Controller) *MockController {
	mock := &MockController{ctrl: ctrl}
	mock.recorder = &MockControllerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockController) EXPECT() *MockControllerMockRecorder {
	return m.recorder
}

// GetClusterDrains mocks base method.
func (m *MockController) GetClusterDrains(ctx context.Context) (types.IsolationGroupConfiguration, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClusterDrains", ctx)
	ret0, _ := ret[0].(types.IsolationGroupConfiguration)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetClusterDrains indicates an expected call of GetClusterDrains.
func (mr *MockControllerMockRecorder) GetClusterDrains(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClusterDrains", reflect.TypeOf((*MockController)(nil).GetClusterDrains), ctx)
}

// GetDomainDrains mocks base method.
func (m *MockController) GetDomainDrains(ctx context.Context) (types.IsolationGroupConfiguration, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDomainDrains", ctx)
	ret0, _ := ret[0].(types.IsolationGroupConfiguration)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDomainDrains indicates an expected call of GetDomainDrains.
func (mr *MockControllerMockRecorder) GetDomainDrains(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDomainDrains", reflect.TypeOf((*MockController)(nil).GetDomainDrains), ctx)
}

// SetClusterDrains mocks base method.
func (m *MockController) SetClusterDrains(ctx context.Context, partition types.IsolationGroupPartition) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetClusterDrains", ctx, partition)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetClusterDrains indicates an expected call of SetClusterDrains.
func (mr *MockControllerMockRecorder) SetClusterDrains(ctx, partition interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetClusterDrains", reflect.TypeOf((*MockController)(nil).SetClusterDrains), ctx, partition)
}

// SetDomainDrains mocks base method.
func (m *MockController) SetDomainDrains(ctx context.Context, partition types.IsolationGroupPartition) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetDomainDrains", ctx, partition)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetDomainDrains indicates an expected call of SetDomainDrains.
func (mr *MockControllerMockRecorder) SetDomainDrains(ctx, partition interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetDomainDrains", reflect.TypeOf((*MockController)(nil).SetDomainDrains), ctx, partition)
}

// MockState is a mock of State interface.
type MockState struct {
	ctrl     *gomock.Controller
	recorder *MockStateMockRecorder
}

// MockStateMockRecorder is the mock recorder for MockState.
type MockStateMockRecorder struct {
	mock *MockState
}

// NewMockState creates a new mock instance.
func NewMockState(ctrl *gomock.Controller) *MockState {
	mock := &MockState{ctrl: ctrl}
	mock.recorder = &MockStateMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockState) EXPECT() *MockStateMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockState) Get(ctx context.Context, domain string) (*IsolationGroups, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, domain)
	ret0, _ := ret[0].(*IsolationGroups)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockStateMockRecorder) Get(ctx, domain interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockState)(nil).Get), ctx, domain)
}

// GetByDomainID mocks base method.
func (m *MockState) GetByDomainID(ctx context.Context, domainID string) (*IsolationGroups, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByDomainID", ctx, domainID)
	ret0, _ := ret[0].(*IsolationGroups)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByDomainID indicates an expected call of GetByDomainID.
func (mr *MockStateMockRecorder) GetByDomainID(ctx, domainID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByDomainID", reflect.TypeOf((*MockState)(nil).GetByDomainID), ctx, domainID)
}

// Start mocks base method.
func (m *MockState) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start.
func (mr *MockStateMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockState)(nil).Start))
}

// Stop mocks base method.
func (m *MockState) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop.
func (mr *MockStateMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockState)(nil).Stop))
}

// Subscribe mocks base method.
func (m *MockState) Subscribe(domainID, key string, notifyChannel chan<- ChangeEvent) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Subscribe", domainID, key, notifyChannel)
	ret0, _ := ret[0].(error)
	return ret0
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockStateMockRecorder) Subscribe(domainID, key, notifyChannel interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockState)(nil).Subscribe), domainID, key, notifyChannel)
}

// Unsubscribe mocks base method.
func (m *MockState) Unsubscribe(domainID, key string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unsubscribe", domainID, key)
	ret0, _ := ret[0].(error)
	return ret0
}

// Unsubscribe indicates an expected call of Unsubscribe.
func (mr *MockStateMockRecorder) Unsubscribe(domainID, key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unsubscribe", reflect.TypeOf((*MockState)(nil).Unsubscribe), domainID, key)
}
