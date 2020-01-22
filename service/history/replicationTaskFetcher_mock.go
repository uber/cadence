// The MIT License (MIT)
//
// Copyright (c) 2020 Uber Technologies, Inc.
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
//

// Code generated by MockGen. DO NOT EDIT.
// Source: replicationTaskFetcher.go

// Package history is a generated GoMock package.
package history

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockReplicationTaskFetcher is a mock of ReplicationTaskFetcher interface
type MockReplicationTaskFetcher struct {
	ctrl     *gomock.Controller
	recorder *MockReplicationTaskFetcherMockRecorder
}

// MockReplicationTaskFetcherMockRecorder is the mock recorder for MockReplicationTaskFetcher
type MockReplicationTaskFetcherMockRecorder struct {
	mock *MockReplicationTaskFetcher
}

// NewMockReplicationTaskFetcher creates a new mock instance
func NewMockReplicationTaskFetcher(ctrl *gomock.Controller) *MockReplicationTaskFetcher {
	mock := &MockReplicationTaskFetcher{ctrl: ctrl}
	mock.recorder = &MockReplicationTaskFetcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockReplicationTaskFetcher) EXPECT() *MockReplicationTaskFetcherMockRecorder {
	return m.recorder
}

// Start mocks base method
func (m *MockReplicationTaskFetcher) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start
func (mr *MockReplicationTaskFetcherMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockReplicationTaskFetcher)(nil).Start))
}

// Stop mocks base method
func (m *MockReplicationTaskFetcher) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop
func (mr *MockReplicationTaskFetcherMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockReplicationTaskFetcher)(nil).Stop))
}

// GetSourceCluster mocks base method
func (m *MockReplicationTaskFetcher) GetSourceCluster() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSourceCluster")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetSourceCluster indicates an expected call of GetSourceCluster
func (mr *MockReplicationTaskFetcherMockRecorder) GetSourceCluster() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSourceCluster", reflect.TypeOf((*MockReplicationTaskFetcher)(nil).GetSourceCluster))
}

// GetRequestChan mocks base method
func (m *MockReplicationTaskFetcher) GetRequestChan() chan<- *request {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRequestChan")
	ret0, _ := ret[0].(chan<- *request)
	return ret0
}

// GetRequestChan indicates an expected call of GetRequestChan
func (mr *MockReplicationTaskFetcherMockRecorder) GetRequestChan() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRequestChan", reflect.TypeOf((*MockReplicationTaskFetcher)(nil).GetRequestChan))
}

// MockReplicationTaskFetchers is a mock of ReplicationTaskFetchers interface
type MockReplicationTaskFetchers struct {
	ctrl     *gomock.Controller
	recorder *MockReplicationTaskFetchersMockRecorder
}

// MockReplicationTaskFetchersMockRecorder is the mock recorder for MockReplicationTaskFetchers
type MockReplicationTaskFetchersMockRecorder struct {
	mock *MockReplicationTaskFetchers
}

// NewMockReplicationTaskFetchers creates a new mock instance
func NewMockReplicationTaskFetchers(ctrl *gomock.Controller) *MockReplicationTaskFetchers {
	mock := &MockReplicationTaskFetchers{ctrl: ctrl}
	mock.recorder = &MockReplicationTaskFetchersMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockReplicationTaskFetchers) EXPECT() *MockReplicationTaskFetchersMockRecorder {
	return m.recorder
}

// Start mocks base method
func (m *MockReplicationTaskFetchers) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start
func (mr *MockReplicationTaskFetchersMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockReplicationTaskFetchers)(nil).Start))
}

// Stop mocks base method
func (m *MockReplicationTaskFetchers) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop
func (mr *MockReplicationTaskFetchersMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockReplicationTaskFetchers)(nil).Stop))
}

// GetFetchers mocks base method
func (m *MockReplicationTaskFetchers) GetFetchers() []ReplicationTaskFetcher {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFetchers")
	ret0, _ := ret[0].([]ReplicationTaskFetcher)
	return ret0
}

// GetFetchers indicates an expected call of GetFetchers
func (mr *MockReplicationTaskFetchersMockRecorder) GetFetchers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFetchers", reflect.TypeOf((*MockReplicationTaskFetchers)(nil).GetFetchers))
}
