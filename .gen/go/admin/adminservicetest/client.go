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

// Code generated by thriftrw-plugin-yarpc
// @generated

package adminservicetest

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	admin "github.com/uber/cadence/.gen/go/admin"
	adminserviceclient "github.com/uber/cadence/.gen/go/admin/adminserviceclient"
	replicator "github.com/uber/cadence/.gen/go/replicator"
	shared "github.com/uber/cadence/.gen/go/shared"
	yarpc "go.uber.org/yarpc"
)

// MockClient implements a gomock-compatible mock client for service
// AdminService.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *_MockClientRecorder
}

var _ adminserviceclient.Interface = (*MockClient)(nil)

type _MockClientRecorder struct {
	mock *MockClient
}

// Build a new mock client for service AdminService.
//
// 	mockCtrl := gomock.NewController(t)
// 	client := adminservicetest.NewMockClient(mockCtrl)
//
// Use EXPECT() to set expectations on the mock.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &_MockClientRecorder{mock}
	return mock
}

// EXPECT returns an object that allows you to define an expectation on the
// AdminService mock client.
func (m *MockClient) EXPECT() *_MockClientRecorder {
	return m.recorder
}

// AddSearchAttribute responds to a AddSearchAttribute call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().AddSearchAttribute(gomock.Any(), ...).Return(...)
// 	... := client.AddSearchAttribute(...)
func (m *MockClient) AddSearchAttribute(
	ctx context.Context,
	_Request *admin.AddSearchAttributeRequest,
	opts ...yarpc.CallOption,
) (err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "AddSearchAttribute", args...)
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) AddSearchAttribute(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "AddSearchAttribute", args...)
}

// CloseShard responds to a CloseShard call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().CloseShard(gomock.Any(), ...).Return(...)
// 	... := client.CloseShard(...)
func (m *MockClient) CloseShard(
	ctx context.Context,
	_Request *shared.CloseShardRequest,
	opts ...yarpc.CallOption,
) (err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "CloseShard", args...)
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) CloseShard(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "CloseShard", args...)
}

// DescribeCluster responds to a DescribeCluster call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().DescribeCluster(gomock.Any(), ...).Return(...)
// 	... := client.DescribeCluster(...)
func (m *MockClient) DescribeCluster(
	ctx context.Context,
	opts ...yarpc.CallOption,
) (success *admin.DescribeClusterResponse, err error) {

	args := []interface{}{ctx}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "DescribeCluster", args...)
	success, _ = ret[i].(*admin.DescribeClusterResponse)
	i++
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) DescribeCluster(
	ctx interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "DescribeCluster", args...)
}

// DescribeHistoryHost responds to a DescribeHistoryHost call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().DescribeHistoryHost(gomock.Any(), ...).Return(...)
// 	... := client.DescribeHistoryHost(...)
func (m *MockClient) DescribeHistoryHost(
	ctx context.Context,
	_Request *shared.DescribeHistoryHostRequest,
	opts ...yarpc.CallOption,
) (success *shared.DescribeHistoryHostResponse, err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "DescribeHistoryHost", args...)
	success, _ = ret[i].(*shared.DescribeHistoryHostResponse)
	i++
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) DescribeHistoryHost(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "DescribeHistoryHost", args...)
}

// DescribeQueue responds to a DescribeQueue call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().DescribeQueue(gomock.Any(), ...).Return(...)
// 	... := client.DescribeQueue(...)
func (m *MockClient) DescribeQueue(
	ctx context.Context,
	_Request *shared.DescribeQueueRequest,
	opts ...yarpc.CallOption,
) (success *shared.DescribeQueueResponse, err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "DescribeQueue", args...)
	success, _ = ret[i].(*shared.DescribeQueueResponse)
	i++
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) DescribeQueue(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "DescribeQueue", args...)
}

// DescribeWorkflowExecution responds to a DescribeWorkflowExecution call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().DescribeWorkflowExecution(gomock.Any(), ...).Return(...)
// 	... := client.DescribeWorkflowExecution(...)
func (m *MockClient) DescribeWorkflowExecution(
	ctx context.Context,
	_Request *admin.DescribeWorkflowExecutionRequest,
	opts ...yarpc.CallOption,
) (success *admin.DescribeWorkflowExecutionResponse, err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "DescribeWorkflowExecution", args...)
	success, _ = ret[i].(*admin.DescribeWorkflowExecutionResponse)
	i++
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) DescribeWorkflowExecution(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "DescribeWorkflowExecution", args...)
}

// GetDLQReplicationMessages responds to a GetDLQReplicationMessages call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().GetDLQReplicationMessages(gomock.Any(), ...).Return(...)
// 	... := client.GetDLQReplicationMessages(...)
func (m *MockClient) GetDLQReplicationMessages(
	ctx context.Context,
	_Request *replicator.GetDLQReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (success *replicator.GetDLQReplicationMessagesResponse, err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "GetDLQReplicationMessages", args...)
	success, _ = ret[i].(*replicator.GetDLQReplicationMessagesResponse)
	i++
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) GetDLQReplicationMessages(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "GetDLQReplicationMessages", args...)
}

// GetDomainReplicationMessages responds to a GetDomainReplicationMessages call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().GetDomainReplicationMessages(gomock.Any(), ...).Return(...)
// 	... := client.GetDomainReplicationMessages(...)
func (m *MockClient) GetDomainReplicationMessages(
	ctx context.Context,
	_Request *replicator.GetDomainReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (success *replicator.GetDomainReplicationMessagesResponse, err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "GetDomainReplicationMessages", args...)
	success, _ = ret[i].(*replicator.GetDomainReplicationMessagesResponse)
	i++
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) GetDomainReplicationMessages(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "GetDomainReplicationMessages", args...)
}

// GetReplicationMessages responds to a GetReplicationMessages call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().GetReplicationMessages(gomock.Any(), ...).Return(...)
// 	... := client.GetReplicationMessages(...)
func (m *MockClient) GetReplicationMessages(
	ctx context.Context,
	_Request *replicator.GetReplicationMessagesRequest,
	opts ...yarpc.CallOption,
) (success *replicator.GetReplicationMessagesResponse, err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "GetReplicationMessages", args...)
	success, _ = ret[i].(*replicator.GetReplicationMessagesResponse)
	i++
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) GetReplicationMessages(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "GetReplicationMessages", args...)
}

// GetWorkflowExecutionRawHistoryV2 responds to a GetWorkflowExecutionRawHistoryV2 call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().GetWorkflowExecutionRawHistoryV2(gomock.Any(), ...).Return(...)
// 	... := client.GetWorkflowExecutionRawHistoryV2(...)
func (m *MockClient) GetWorkflowExecutionRawHistoryV2(
	ctx context.Context,
	_GetRequest *admin.GetWorkflowExecutionRawHistoryV2Request,
	opts ...yarpc.CallOption,
) (success *admin.GetWorkflowExecutionRawHistoryV2Response, err error) {

	args := []interface{}{ctx, _GetRequest}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "GetWorkflowExecutionRawHistoryV2", args...)
	success, _ = ret[i].(*admin.GetWorkflowExecutionRawHistoryV2Response)
	i++
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) GetWorkflowExecutionRawHistoryV2(
	ctx interface{},
	_GetRequest interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _GetRequest}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "GetWorkflowExecutionRawHistoryV2", args...)
}

// MergeDLQMessages responds to a MergeDLQMessages call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().MergeDLQMessages(gomock.Any(), ...).Return(...)
// 	... := client.MergeDLQMessages(...)
func (m *MockClient) MergeDLQMessages(
	ctx context.Context,
	_Request *replicator.MergeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (success *replicator.MergeDLQMessagesResponse, err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "MergeDLQMessages", args...)
	success, _ = ret[i].(*replicator.MergeDLQMessagesResponse)
	i++
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) MergeDLQMessages(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "MergeDLQMessages", args...)
}

// PurgeDLQMessages responds to a PurgeDLQMessages call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().PurgeDLQMessages(gomock.Any(), ...).Return(...)
// 	... := client.PurgeDLQMessages(...)
func (m *MockClient) PurgeDLQMessages(
	ctx context.Context,
	_Request *replicator.PurgeDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "PurgeDLQMessages", args...)
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) PurgeDLQMessages(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "PurgeDLQMessages", args...)
}

// ReadDLQMessages responds to a ReadDLQMessages call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().ReadDLQMessages(gomock.Any(), ...).Return(...)
// 	... := client.ReadDLQMessages(...)
func (m *MockClient) ReadDLQMessages(
	ctx context.Context,
	_Request *replicator.ReadDLQMessagesRequest,
	opts ...yarpc.CallOption,
) (success *replicator.ReadDLQMessagesResponse, err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "ReadDLQMessages", args...)
	success, _ = ret[i].(*replicator.ReadDLQMessagesResponse)
	i++
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) ReadDLQMessages(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "ReadDLQMessages", args...)
}

// ReapplyEvents responds to a ReapplyEvents call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().ReapplyEvents(gomock.Any(), ...).Return(...)
// 	... := client.ReapplyEvents(...)
func (m *MockClient) ReapplyEvents(
	ctx context.Context,
	_ReapplyEventsRequest *shared.ReapplyEventsRequest,
	opts ...yarpc.CallOption,
) (err error) {

	args := []interface{}{ctx, _ReapplyEventsRequest}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "ReapplyEvents", args...)
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) ReapplyEvents(
	ctx interface{},
	_ReapplyEventsRequest interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _ReapplyEventsRequest}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "ReapplyEvents", args...)
}

// RefreshWorkflowTasks responds to a RefreshWorkflowTasks call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().RefreshWorkflowTasks(gomock.Any(), ...).Return(...)
// 	... := client.RefreshWorkflowTasks(...)
func (m *MockClient) RefreshWorkflowTasks(
	ctx context.Context,
	_Request *shared.RefreshWorkflowTasksRequest,
	opts ...yarpc.CallOption,
) (err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "RefreshWorkflowTasks", args...)
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) RefreshWorkflowTasks(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "RefreshWorkflowTasks", args...)
}

// RemoveTask responds to a RemoveTask call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().RemoveTask(gomock.Any(), ...).Return(...)
// 	... := client.RemoveTask(...)
func (m *MockClient) RemoveTask(
	ctx context.Context,
	_Request *shared.RemoveTaskRequest,
	opts ...yarpc.CallOption,
) (err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "RemoveTask", args...)
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) RemoveTask(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "RemoveTask", args...)
}

// ResendReplicationTasks responds to a ResendReplicationTasks call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().ResendReplicationTasks(gomock.Any(), ...).Return(...)
// 	... := client.ResendReplicationTasks(...)
func (m *MockClient) ResendReplicationTasks(
	ctx context.Context,
	_Request *admin.ResendReplicationTasksRequest,
	opts ...yarpc.CallOption,
) (err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "ResendReplicationTasks", args...)
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) ResendReplicationTasks(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "ResendReplicationTasks", args...)
}

// ResetQueue responds to a ResetQueue call based on the mock expectations. This
// call will fail if the mock does not expect this call. Use EXPECT to expect
// a call to this function.
//
// 	client.EXPECT().ResetQueue(gomock.Any(), ...).Return(...)
// 	... := client.ResetQueue(...)
func (m *MockClient) ResetQueue(
	ctx context.Context,
	_Request *shared.ResetQueueRequest,
	opts ...yarpc.CallOption,
) (err error) {

	args := []interface{}{ctx, _Request}
	for _, o := range opts {
		args = append(args, o)
	}
	i := 0
	ret := m.ctrl.Call(m, "ResetQueue", args...)
	err, _ = ret[i].(error)
	return
}

func (mr *_MockClientRecorder) ResetQueue(
	ctx interface{},
	_Request interface{},
	opts ...interface{},
) *gomock.Call {
	args := append([]interface{}{ctx, _Request}, opts...)
	return mr.mock.ctrl.RecordCall(mr.mock, "ResetQueue", args...)
}
