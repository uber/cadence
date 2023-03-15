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
// Source: partitioning.go

// Package partition is a generated GoMock package.
package partition

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockPartitioner is a mock of Partitioner interface.
type MockPartitioner struct {
	ctrl     *gomock.Controller
	recorder *MockPartitionerMockRecorder
}

// MockPartitionerMockRecorder is the mock recorder for MockPartitioner.
type MockPartitionerMockRecorder struct {
	mock *MockPartitioner
}

// NewMockPartitioner creates a new mock instance.
func NewMockPartitioner(ctrl *gomock.Controller) *MockPartitioner {
	mock := &MockPartitioner{ctrl: ctrl}
	mock.recorder = &MockPartitionerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPartitioner) EXPECT() *MockPartitionerMockRecorder {
	return m.recorder
}

// GetIsolationGroupByDomainID mocks base method.
func (m *MockPartitioner) GetIsolationGroupByDomainID(ctx context.Context, DomainID string, partitionKey PartitionConfig) (*string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetIsolationGroupByDomainID", ctx, DomainID, partitionKey)
	ret0, _ := ret[0].(*string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetIsolationGroupByDomainID indicates an expected call of GetIsolationGroupByDomainID.
func (mr *MockPartitionerMockRecorder) GetIsolationGroupByDomainID(ctx, DomainID, partitionKey interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIsolationGroupByDomainID", reflect.TypeOf((*MockPartitioner)(nil).GetIsolationGroupByDomainID), ctx, DomainID, partitionKey)
}

// IsDrained mocks base method.
func (m *MockPartitioner) IsDrained(ctx context.Context, Domain, IsolationGroup string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsDrained", ctx, Domain, IsolationGroup)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsDrained indicates an expected call of IsDrained.
func (mr *MockPartitionerMockRecorder) IsDrained(ctx, Domain, IsolationGroup interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsDrained", reflect.TypeOf((*MockPartitioner)(nil).IsDrained), ctx, Domain, IsolationGroup)
}

// IsDrainedByDomainID mocks base method.
func (m *MockPartitioner) IsDrainedByDomainID(ctx context.Context, DomainID, IsolationGroup string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsDrainedByDomainID", ctx, DomainID, IsolationGroup)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsDrainedByDomainID indicates an expected call of IsDrainedByDomainID.
func (mr *MockPartitionerMockRecorder) IsDrainedByDomainID(ctx, DomainID, IsolationGroup interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsDrainedByDomainID", reflect.TypeOf((*MockPartitioner)(nil).IsDrainedByDomainID), ctx, DomainID, IsolationGroup)
}
