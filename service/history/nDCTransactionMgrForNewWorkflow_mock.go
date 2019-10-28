// The MIT License (MIT)
//
// Copyright (c) 2019 Uber Technologies, Inc.
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
// Source: nDCTransactionMgrForNewWorkflow.go

// Package history is a generated GoMock package.
package history

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MocknDCTransactionMgrForNewWorkflow is a mock of nDCTransactionMgrForNewWorkflow interface
type MocknDCTransactionMgrForNewWorkflow struct {
	ctrl     *gomock.Controller
	recorder *MocknDCTransactionMgrForNewWorkflowMockRecorder
}

// MocknDCTransactionMgrForNewWorkflowMockRecorder is the mock recorder for MocknDCTransactionMgrForNewWorkflow
type MocknDCTransactionMgrForNewWorkflowMockRecorder struct {
	mock *MocknDCTransactionMgrForNewWorkflow
}

// NewMocknDCTransactionMgrForNewWorkflow creates a new mock instance
func NewMocknDCTransactionMgrForNewWorkflow(ctrl *gomock.Controller) *MocknDCTransactionMgrForNewWorkflow {
	mock := &MocknDCTransactionMgrForNewWorkflow{ctrl: ctrl}
	mock.recorder = &MocknDCTransactionMgrForNewWorkflowMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MocknDCTransactionMgrForNewWorkflow) EXPECT() *MocknDCTransactionMgrForNewWorkflowMockRecorder {
	return m.recorder
}

// dispatchForNewWorkflow mocks base method
func (m *MocknDCTransactionMgrForNewWorkflow) dispatchForNewWorkflow(ctx context.Context, now time.Time, targetWorkflow nDCWorkflow) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "dispatchForNewWorkflow", ctx, now, targetWorkflow)
	ret0, _ := ret[0].(error)
	return ret0
}

// dispatchForNewWorkflow indicates an expected call of dispatchForNewWorkflow
func (mr *MocknDCTransactionMgrForNewWorkflowMockRecorder) dispatchForNewWorkflow(ctx, now, targetWorkflow interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "dispatchForNewWorkflow", reflect.TypeOf((*MocknDCTransactionMgrForNewWorkflow)(nil).dispatchForNewWorkflow), ctx, now, targetWorkflow)
}
