// Copyright (c) 2019 Uber Technologies, Inc.
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

// Code generated by thriftrw-plugin-yarpc
// @generated

package matchingservicetest

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	matching "github.com/uber/cadence/.gen/go/matching"
	matchingserviceclient "github.com/uber/cadence/.gen/go/matching/matchingserviceclient"
	shared "github.com/uber/cadence/.gen/go/shared"
	yarpc "go.uber.org/yarpc"
)

// MockClient implements a gomock-compatible mock client for service
// MatchingService.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *_MockClientRecorder
}

var _ matchingserviceclient.Interface = (*MockClient)(nil)

type _MockClientRecorder struct {
	mock *MockClient
}

// Build a new mock client for service MatchingService.
//
// 	mockCtrl := gomock.NewController(t)
// 	client := matchingservicetest.NewMockClient(mockCtrl)
//
// Use EXPECT() to set expectations on the mock.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &_MockClientRecorder{mock}
	return mock
}

// EXPECT returns an object that allows you to define an expectation on the
// MatchingService mock client.
func (m *MockClient) EXPECT() *_MockClientRecorder {
	return m.recorder
}

// AddActivityTask responds to a AddActivityTask call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().AddActivityTask(gomock.Any(), ...).Return(...)
// 	... := client.AddActivityTask(...)
func (m *MockClient) AddActivityTask(
	ctx context.Context,
	_AddRequest *matching.AddActivityTaskRequest,
	opts ...yarpc.CallOption,
) (err error) {

	args := []interface{}{ctx, _AddRequest}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "AddActivityTask", args...)
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) AddActivityTask(
	ctx interface{},
	_AddRequest interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _AddRequest}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "AddActivityTask", args...)
}

// AddDecisionTask responds to a AddDecisionTask call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().AddDecisionTask(gomock.Any(), ...).Return(...)
// 	... := client.AddDecisionTask(...)
func (m *MockClient) AddDecisionTask(
	ctx context.Context,
	_AddRequest *matching.AddDecisionTaskRequest,
	opts ...yarpc.CallOption,
) (err error) {

	args := []interface{}{ctx, _AddRequest}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "AddDecisionTask", args...)
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) AddDecisionTask(
	ctx interface{},
	_AddRequest interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _AddRequest}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "AddDecisionTask", args...)
}

// CancelOutstandingPoll responds to a CancelOutstandingPoll call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().CancelOutstandingPoll(gomock.Any(), ...).Return(...)
// 	... := client.CancelOutstandingPoll(...)
func (m *MockClient) CancelOutstandingPoll(
	ctx context.Context,
	_Request *matching.CancelOutstandingPollRequest,
	opts ...yarpc.CallOption,
) (err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "CancelOutstandingPoll", args...)
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) CancelOutstandingPoll(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "CancelOutstandingPoll", args...)
}

// DescribeTaskList responds to a DescribeTaskList call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().DescribeTaskList(gomock.Any(), ...).Return(...)
// 	... := client.DescribeTaskList(...)
func (m *MockClient) DescribeTaskList(
	ctx context.Context,
	_Request *matching.DescribeTaskListRequest,
	opts ...yarpc.CallOption,
) (success *shared.DescribeTaskListResponse, err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "DescribeTaskList", args...)
	success, _ = ret[i].(*shared.DescribeTaskListResponse)
	i++
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) DescribeTaskList(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "DescribeTaskList", args...)
}

// PollForActivityTask responds to a PollForActivityTask call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().PollForActivityTask(gomock.Any(), ...).Return(...)
// 	... := client.PollForActivityTask(...)
func (m *MockClient) PollForActivityTask(
	ctx context.Context,
	_PollRequest *matching.PollForActivityTaskRequest,
	opts ...yarpc.CallOption,
) (success *shared.PollForActivityTaskResponse, err error) {

	args := []interface{}{ctx, _PollRequest}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "PollForActivityTask", args...)
	success, _ = ret[i].(*shared.PollForActivityTaskResponse)
	i++
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) PollForActivityTask(
	ctx interface{},
	_PollRequest interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _PollRequest}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "PollForActivityTask", args...)
}

// PollForDecisionTask responds to a PollForDecisionTask call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().PollForDecisionTask(gomock.Any(), ...).Return(...)
// 	... := client.PollForDecisionTask(...)
func (m *MockClient) PollForDecisionTask(
	ctx context.Context,
	_PollRequest *matching.PollForDecisionTaskRequest,
	opts ...yarpc.CallOption,
) (success *matching.PollForDecisionTaskResponse, err error) {

	args := []interface{}{ctx, _PollRequest}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "PollForDecisionTask", args...)
	success, _ = ret[i].(*matching.PollForDecisionTaskResponse)
	i++
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) PollForDecisionTask(
	ctx interface{},
	_PollRequest interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _PollRequest}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "PollForDecisionTask", args...)
}

// QueryWorkflow responds to a QueryWorkflow call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().QueryWorkflow(gomock.Any(), ...).Return(...)
// 	... := client.QueryWorkflow(...)
func (m *MockClient) QueryWorkflow(
	ctx context.Context,
	_QueryRequest *matching.QueryWorkflowRequest,
	opts ...yarpc.CallOption,
) (success *shared.QueryWorkflowResponse, err error) {

	args := []interface{}{ctx, _QueryRequest}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "QueryWorkflow", args...)
	success, _ = ret[i].(*shared.QueryWorkflowResponse)
	i++
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) QueryWorkflow(
	ctx interface{},
	_QueryRequest interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _QueryRequest}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "QueryWorkflow", args...)
}

// RespondQueryTaskCompleted responds to a RespondQueryTaskCompleted call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().RespondQueryTaskCompleted(gomock.Any(), ...).Return(...)
// 	... := client.RespondQueryTaskCompleted(...)
func (m *MockClient) RespondQueryTaskCompleted(
	ctx context.Context,
	_Request *matching.RespondQueryTaskCompletedRequest,
	opts ...yarpc.CallOption,
) (err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "RespondQueryTaskCompleted", args...)
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) RespondQueryTaskCompleted(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "RespondQueryTaskCompleted", args...)
}
