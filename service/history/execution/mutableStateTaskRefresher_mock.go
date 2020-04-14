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
// Source: mutableStateTaskRefresher.go

// Package execution is a generated GoMock package.
package execution

import (
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockMutableStateTaskRefresher is a mock of MutableStateTaskRefresher interface
type MockMutableStateTaskRefresher struct {
	ctrl     *gomock.Controller
	recorder *MockMutableStateTaskRefresherMockRecorder
}

// MockMutableStateTaskRefresherMockRecorder is the mock recorder for MockMutableStateTaskRefresher
type MockMutableStateTaskRefresherMockRecorder struct {
	mock *MockMutableStateTaskRefresher
}

// NewMockMutableStateTaskRefresher creates a new mock instance
func NewMockMutableStateTaskRefresher(ctrl *gomock.Controller) *MockMutableStateTaskRefresher {
	mock := &MockMutableStateTaskRefresher{ctrl: ctrl}
	mock.recorder = &MockMutableStateTaskRefresherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMutableStateTaskRefresher) EXPECT() *MockMutableStateTaskRefresherMockRecorder {
	return m.recorder
}

// RefreshTasks mocks base method
func (m *MockMutableStateTaskRefresher) RefreshTasks(now time.Time, mutableState MutableState) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RefreshTasks", now, mutableState)
	ret0, _ := ret[0].(error)
	return ret0
}

// RefreshTasks indicates an expected call of RefreshTasks
func (mr *MockMutableStateTaskRefresherMockRecorder) RefreshTasks(now, mutableState interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshTasks", reflect.TypeOf((*MockMutableStateTaskRefresher)(nil).RefreshTasks), now, mutableState)
}
