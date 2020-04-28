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
// Source: interface.go

// Package task is a generated GoMock package.
package task

import (
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"

	task "github.com/uber/cadence/common/task"
	shard "github.com/uber/cadence/service/history/shard"
)

// MockInfo is a mock of Info interface
type MockInfo struct {
	ctrl     *gomock.Controller
	recorder *MockInfoMockRecorder
}

// MockInfoMockRecorder is the mock recorder for MockInfo
type MockInfoMockRecorder struct {
	mock *MockInfo
}

// NewMockInfo creates a new mock instance
func NewMockInfo(ctrl *gomock.Controller) *MockInfo {
	mock := &MockInfo{ctrl: ctrl}
	mock.recorder = &MockInfoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockInfo) EXPECT() *MockInfoMockRecorder {
	return m.recorder
}

// GetVersion mocks base method
func (m *MockInfo) GetVersion() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVersion")
	ret0, _ := ret[0].(int64)
	return ret0
}

// GetVersion indicates an expected call of GetVersion
func (mr *MockInfoMockRecorder) GetVersion() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVersion", reflect.TypeOf((*MockInfo)(nil).GetVersion))
}

// GetTaskID mocks base method
func (m *MockInfo) GetTaskID() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTaskID")
	ret0, _ := ret[0].(int64)
	return ret0
}

// GetTaskID indicates an expected call of GetTaskID
func (mr *MockInfoMockRecorder) GetTaskID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTaskID", reflect.TypeOf((*MockInfo)(nil).GetTaskID))
}

// GetTaskType mocks base method
func (m *MockInfo) GetTaskType() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTaskType")
	ret0, _ := ret[0].(int)
	return ret0
}

// GetTaskType indicates an expected call of GetTaskType
func (mr *MockInfoMockRecorder) GetTaskType() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTaskType", reflect.TypeOf((*MockInfo)(nil).GetTaskType))
}

// GetVisibilityTimestamp mocks base method
func (m *MockInfo) GetVisibilityTimestamp() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVisibilityTimestamp")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// GetVisibilityTimestamp indicates an expected call of GetVisibilityTimestamp
func (mr *MockInfoMockRecorder) GetVisibilityTimestamp() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVisibilityTimestamp", reflect.TypeOf((*MockInfo)(nil).GetVisibilityTimestamp))
}

// GetWorkflowID mocks base method
func (m *MockInfo) GetWorkflowID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWorkflowID")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetWorkflowID indicates an expected call of GetWorkflowID
func (mr *MockInfoMockRecorder) GetWorkflowID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWorkflowID", reflect.TypeOf((*MockInfo)(nil).GetWorkflowID))
}

// GetRunID mocks base method
func (m *MockInfo) GetRunID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRunID")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetRunID indicates an expected call of GetRunID
func (mr *MockInfoMockRecorder) GetRunID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRunID", reflect.TypeOf((*MockInfo)(nil).GetRunID))
}

// GetDomainID mocks base method
func (m *MockInfo) GetDomainID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDomainID")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetDomainID indicates an expected call of GetDomainID
func (mr *MockInfoMockRecorder) GetDomainID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDomainID", reflect.TypeOf((*MockInfo)(nil).GetDomainID))
}

// MockTask is a mock of Task interface
type MockTask struct {
	ctrl     *gomock.Controller
	recorder *MockTaskMockRecorder
}

// MockTaskMockRecorder is the mock recorder for MockTask
type MockTaskMockRecorder struct {
	mock *MockTask
}

// NewMockTask creates a new mock instance
func NewMockTask(ctrl *gomock.Controller) *MockTask {
	mock := &MockTask{ctrl: ctrl}
	mock.recorder = &MockTaskMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTask) EXPECT() *MockTaskMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockTask) Execute() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute")
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockTaskMockRecorder) Execute() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockTask)(nil).Execute))
}

// HandleErr mocks base method
func (m *MockTask) HandleErr(err error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HandleErr", err)
	ret0, _ := ret[0].(error)
	return ret0
}

// HandleErr indicates an expected call of HandleErr
func (mr *MockTaskMockRecorder) HandleErr(err interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleErr", reflect.TypeOf((*MockTask)(nil).HandleErr), err)
}

// RetryErr mocks base method
func (m *MockTask) RetryErr(err error) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RetryErr", err)
	ret0, _ := ret[0].(bool)
	return ret0
}

// RetryErr indicates an expected call of RetryErr
func (mr *MockTaskMockRecorder) RetryErr(err interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RetryErr", reflect.TypeOf((*MockTask)(nil).RetryErr), err)
}

// Ack mocks base method
func (m *MockTask) Ack() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Ack")
}

// Ack indicates an expected call of Ack
func (mr *MockTaskMockRecorder) Ack() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ack", reflect.TypeOf((*MockTask)(nil).Ack))
}

// Nack mocks base method
func (m *MockTask) Nack() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Nack")
}

// Nack indicates an expected call of Nack
func (mr *MockTaskMockRecorder) Nack() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Nack", reflect.TypeOf((*MockTask)(nil).Nack))
}

// State mocks base method
func (m *MockTask) State() task.State {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "State")
	ret0, _ := ret[0].(task.State)
	return ret0
}

// State indicates an expected call of State
func (mr *MockTaskMockRecorder) State() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "State", reflect.TypeOf((*MockTask)(nil).State))
}

// Priority mocks base method
func (m *MockTask) Priority() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Priority")
	ret0, _ := ret[0].(int)
	return ret0
}

// Priority indicates an expected call of Priority
func (mr *MockTaskMockRecorder) Priority() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Priority", reflect.TypeOf((*MockTask)(nil).Priority))
}

// SetPriority mocks base method
func (m *MockTask) SetPriority(arg0 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetPriority", arg0)
}

// SetPriority indicates an expected call of SetPriority
func (mr *MockTaskMockRecorder) SetPriority(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetPriority", reflect.TypeOf((*MockTask)(nil).SetPriority), arg0)
}

// GetVersion mocks base method
func (m *MockTask) GetVersion() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVersion")
	ret0, _ := ret[0].(int64)
	return ret0
}

// GetVersion indicates an expected call of GetVersion
func (mr *MockTaskMockRecorder) GetVersion() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVersion", reflect.TypeOf((*MockTask)(nil).GetVersion))
}

// GetTaskID mocks base method
func (m *MockTask) GetTaskID() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTaskID")
	ret0, _ := ret[0].(int64)
	return ret0
}

// GetTaskID indicates an expected call of GetTaskID
func (mr *MockTaskMockRecorder) GetTaskID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTaskID", reflect.TypeOf((*MockTask)(nil).GetTaskID))
}

// GetTaskType mocks base method
func (m *MockTask) GetTaskType() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTaskType")
	ret0, _ := ret[0].(int)
	return ret0
}

// GetTaskType indicates an expected call of GetTaskType
func (mr *MockTaskMockRecorder) GetTaskType() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTaskType", reflect.TypeOf((*MockTask)(nil).GetTaskType))
}

// GetVisibilityTimestamp mocks base method
func (m *MockTask) GetVisibilityTimestamp() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVisibilityTimestamp")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// GetVisibilityTimestamp indicates an expected call of GetVisibilityTimestamp
func (mr *MockTaskMockRecorder) GetVisibilityTimestamp() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVisibilityTimestamp", reflect.TypeOf((*MockTask)(nil).GetVisibilityTimestamp))
}

// GetWorkflowID mocks base method
func (m *MockTask) GetWorkflowID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWorkflowID")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetWorkflowID indicates an expected call of GetWorkflowID
func (mr *MockTaskMockRecorder) GetWorkflowID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWorkflowID", reflect.TypeOf((*MockTask)(nil).GetWorkflowID))
}

// GetRunID mocks base method
func (m *MockTask) GetRunID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRunID")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetRunID indicates an expected call of GetRunID
func (mr *MockTaskMockRecorder) GetRunID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRunID", reflect.TypeOf((*MockTask)(nil).GetRunID))
}

// GetDomainID mocks base method
func (m *MockTask) GetDomainID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDomainID")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetDomainID indicates an expected call of GetDomainID
func (mr *MockTaskMockRecorder) GetDomainID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDomainID", reflect.TypeOf((*MockTask)(nil).GetDomainID))
}

// GetQueueType mocks base method
func (m *MockTask) GetQueueType() QueueType {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetQueueType")
	ret0, _ := ret[0].(QueueType)
	return ret0
}

// GetQueueType indicates an expected call of GetQueueType
func (mr *MockTaskMockRecorder) GetQueueType() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetQueueType", reflect.TypeOf((*MockTask)(nil).GetQueueType))
}

// GetShard mocks base method
func (m *MockTask) GetShard() shard.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetShard")
	ret0, _ := ret[0].(shard.Context)
	return ret0
}

// GetShard indicates an expected call of GetShard
func (mr *MockTaskMockRecorder) GetShard() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetShard", reflect.TypeOf((*MockTask)(nil).GetShard))
}

// GetAttempt mocks base method
func (m *MockTask) GetAttempt() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAttempt")
	ret0, _ := ret[0].(int)
	return ret0
}

// GetAttempt indicates an expected call of GetAttempt
func (mr *MockTaskMockRecorder) GetAttempt() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAttempt", reflect.TypeOf((*MockTask)(nil).GetAttempt))
}

// MockKey is a mock of Key interface
type MockKey struct {
	ctrl     *gomock.Controller
	recorder *MockKeyMockRecorder
}

// MockKeyMockRecorder is the mock recorder for MockKey
type MockKeyMockRecorder struct {
	mock *MockKey
}

// NewMockKey creates a new mock instance
func NewMockKey(ctrl *gomock.Controller) *MockKey {
	mock := &MockKey{ctrl: ctrl}
	mock.recorder = &MockKeyMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockKey) EXPECT() *MockKeyMockRecorder {
	return m.recorder
}

// Less mocks base method
func (m *MockKey) Less(arg0 Key) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Less", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Less indicates an expected call of Less
func (mr *MockKeyMockRecorder) Less(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Less", reflect.TypeOf((*MockKey)(nil).Less), arg0)
}

// MockExecutor is a mock of Executor interface
type MockExecutor struct {
	ctrl     *gomock.Controller
	recorder *MockExecutorMockRecorder
}

// MockExecutorMockRecorder is the mock recorder for MockExecutor
type MockExecutorMockRecorder struct {
	mock *MockExecutor
}

// NewMockExecutor creates a new mock instance
func NewMockExecutor(ctrl *gomock.Controller) *MockExecutor {
	mock := &MockExecutor{ctrl: ctrl}
	mock.recorder = &MockExecutorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockExecutor) EXPECT() *MockExecutorMockRecorder {
	return m.recorder
}

// Execute mocks base method
func (m *MockExecutor) Execute(taskInfo Info, shouldProcessTask bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Execute", taskInfo, shouldProcessTask)
	ret0, _ := ret[0].(error)
	return ret0
}

// Execute indicates an expected call of Execute
func (mr *MockExecutorMockRecorder) Execute(taskInfo, shouldProcessTask interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockExecutor)(nil).Execute), taskInfo, shouldProcessTask)
}

// MockPriorityAssigner is a mock of PriorityAssigner interface
type MockPriorityAssigner struct {
	ctrl     *gomock.Controller
	recorder *MockPriorityAssignerMockRecorder
}

// MockPriorityAssignerMockRecorder is the mock recorder for MockPriorityAssigner
type MockPriorityAssignerMockRecorder struct {
	mock *MockPriorityAssigner
}

// NewMockPriorityAssigner creates a new mock instance
func NewMockPriorityAssigner(ctrl *gomock.Controller) *MockPriorityAssigner {
	mock := &MockPriorityAssigner{ctrl: ctrl}
	mock.recorder = &MockPriorityAssignerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPriorityAssigner) EXPECT() *MockPriorityAssignerMockRecorder {
	return m.recorder
}

// Assign mocks base method
func (m *MockPriorityAssigner) Assign(arg0 Task) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Assign", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Assign indicates an expected call of Assign
func (mr *MockPriorityAssignerMockRecorder) Assign(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Assign", reflect.TypeOf((*MockPriorityAssigner)(nil).Assign), arg0)
}

// MockProcessor is a mock of Processor interface
type MockProcessor struct {
	ctrl     *gomock.Controller
	recorder *MockProcessorMockRecorder
}

// MockProcessorMockRecorder is the mock recorder for MockProcessor
type MockProcessorMockRecorder struct {
	mock *MockProcessor
}

// NewMockProcessor creates a new mock instance
func NewMockProcessor(ctrl *gomock.Controller) *MockProcessor {
	mock := &MockProcessor{ctrl: ctrl}
	mock.recorder = &MockProcessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockProcessor) EXPECT() *MockProcessorMockRecorder {
	return m.recorder
}

// Start mocks base method
func (m *MockProcessor) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start
func (mr *MockProcessorMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockProcessor)(nil).Start))
}

// Stop mocks base method
func (m *MockProcessor) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop
func (mr *MockProcessorMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockProcessor)(nil).Stop))
}

// StopShardProcessor mocks base method
func (m *MockProcessor) StopShardProcessor(arg0 shard.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "StopShardProcessor", arg0)
}

// StopShardProcessor indicates an expected call of StopShardProcessor
func (mr *MockProcessorMockRecorder) StopShardProcessor(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopShardProcessor", reflect.TypeOf((*MockProcessor)(nil).StopShardProcessor), arg0)
}

// Submit mocks base method
func (m *MockProcessor) Submit(arg0 Task) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Submit", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Submit indicates an expected call of Submit
func (mr *MockProcessorMockRecorder) Submit(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Submit", reflect.TypeOf((*MockProcessor)(nil).Submit), arg0)
}

// TrySubmit mocks base method
func (m *MockProcessor) TrySubmit(arg0 Task) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TrySubmit", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TrySubmit indicates an expected call of TrySubmit
func (mr *MockProcessorMockRecorder) TrySubmit(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TrySubmit", reflect.TypeOf((*MockProcessor)(nil).TrySubmit), arg0)
}
