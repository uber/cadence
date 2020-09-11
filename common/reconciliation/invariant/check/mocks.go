// The MIT License (MIT)
//
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
//

// Code generated by MockGen. DO NOT EDIT.
// Source: types.go

// Package check is a generated GoMock package.
package check

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockInvariant is a mock of Invariant interface
type MockInvariant struct {
	ctrl     *gomock.Controller
	recorder *MockInvariantMockRecorder
}

// MockInvariantMockRecorder is the mock recorder for MockInvariant
type MockInvariantMockRecorder struct {
	mock *MockInvariant
}

// NewMockInvariant creates a new mock instance
func NewMockInvariant(ctrl *gomock.Controller) *MockInvariant {
	mock := &MockInvariant{ctrl: ctrl}
	mock.recorder = &MockInvariantMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockInvariant) EXPECT() *MockInvariantMockRecorder {
	return m.recorder
}

// Check mocks base method
func (m *MockInvariant) Check(arg0 interface{}) CheckResult {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Check", arg0)
	ret0, _ := ret[0].(CheckResult)
	return ret0
}

// Check indicates an expected call of Check
func (mr *MockInvariantMockRecorder) Check(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Check", reflect.TypeOf((*MockInvariant)(nil).Check), arg0)
}

// Fix mocks base method
func (m *MockInvariant) Fix(arg0 interface{}) FixResult {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Fix", arg0)
	ret0, _ := ret[0].(FixResult)
	return ret0
}

// Fix indicates an expected call of Fix
func (mr *MockInvariantMockRecorder) Fix(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fix", reflect.TypeOf((*MockInvariant)(nil).Fix), arg0)
}

// InvariantType mocks base method
func (m *MockInvariant) InvariantType() InvariantType {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InvariantType")
	ret0, _ := ret[0].(InvariantType)
	return ret0
}

// InvariantType indicates an expected call of InvariantType
func (mr *MockInvariantMockRecorder) InvariantType() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InvariantType", reflect.TypeOf((*MockInvariant)(nil).InvariantType))
}
