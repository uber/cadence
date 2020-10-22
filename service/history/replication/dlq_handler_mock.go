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
// Source: dlq_handler.go

// Package replication is a generated GoMock package.
package replication

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"

	replicator "github.com/uber/cadence/.gen/go/replicator"
)

// MockDLQHandler is a mock of DLQHandler interface
type MockDLQHandler struct {
	ctrl     *gomock.Controller
	recorder *MockDLQHandlerMockRecorder
}

// MockDLQHandlerMockRecorder is the mock recorder for MockDLQHandler
type MockDLQHandlerMockRecorder struct {
	mock *MockDLQHandler
}

// NewMockDLQHandler creates a new mock instance
func NewMockDLQHandler(ctrl *gomock.Controller) *MockDLQHandler {
	mock := &MockDLQHandler{ctrl: ctrl}
	mock.recorder = &MockDLQHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDLQHandler) EXPECT() *MockDLQHandlerMockRecorder {
	return m.recorder
}

// ReadMessages mocks base method
func (m *MockDLQHandler) ReadMessages(ctx context.Context, sourceCluster string, lastMessageID int64, pageSize int, pageToken []byte) ([]*replicator.ReplicationTask, []byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadMessages", ctx, sourceCluster, lastMessageID, pageSize, pageToken)
	ret0, _ := ret[0].([]*replicator.ReplicationTask)
	ret1, _ := ret[1].([]byte)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ReadMessages indicates an expected call of ReadMessages
func (mr *MockDLQHandlerMockRecorder) ReadMessages(ctx, sourceCluster, lastMessageID, pageSize, pageToken interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadMessages", reflect.TypeOf((*MockDLQHandler)(nil).ReadMessages), ctx, sourceCluster, lastMessageID, pageSize, pageToken)
}

// PurgeMessages mocks base method
func (m *MockDLQHandler) PurgeMessages(ctx context.Context, sourceCluster string, lastMessageID int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PurgeMessages", ctx, sourceCluster, lastMessageID)
	ret0, _ := ret[0].(error)
	return ret0
}

// PurgeMessages indicates an expected call of PurgeMessages
func (mr *MockDLQHandlerMockRecorder) PurgeMessages(ctx, sourceCluster, lastMessageID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PurgeMessages", reflect.TypeOf((*MockDLQHandler)(nil).PurgeMessages), ctx, sourceCluster, lastMessageID)
}

// MergeMessages mocks base method
func (m *MockDLQHandler) MergeMessages(ctx context.Context, sourceCluster string, lastMessageID int64, pageSize int, pageToken []byte) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MergeMessages", ctx, sourceCluster, lastMessageID, pageSize, pageToken)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MergeMessages indicates an expected call of MergeMessages
func (mr *MockDLQHandlerMockRecorder) MergeMessages(ctx, sourceCluster, lastMessageID, pageSize, pageToken interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MergeMessages", reflect.TypeOf((*MockDLQHandler)(nil).MergeMessages), ctx, sourceCluster, lastMessageID, pageSize, pageToken)
}
