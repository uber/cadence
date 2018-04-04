// Copyright (c) 2017 Uber Technologies, Inc.
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

package mocks

import "context"
import "go.uber.org/yarpc"
import history "github.com/uber/cadence/.gen/go/history"
import mock "github.com/stretchr/testify/mock"
import shared "github.com/uber/cadence/.gen/go/shared"

// Client is an autogenerated mock type for the Client type
type HistoryClient struct {
	mock.Mock
}

// GetMutableState provides a mock function with given fields: ctx, getRequest
func (_m *HistoryClient) GetMutableState(ctx context.Context, getRequest *history.GetMutableStateRequest, opts ...yarpc.CallOption) (*history.GetMutableStateResponse, error) {
	ret := _m.Called(ctx, getRequest)

	var r0 *history.GetMutableStateResponse
	if rf, ok := ret.Get(0).(func(context.Context, *history.GetMutableStateRequest) *history.GetMutableStateResponse); ok {
		r0 = rf(ctx, getRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*history.GetMutableStateResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *history.GetMutableStateRequest) error); ok {
		r1 = rf(ctx, getRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ResetStickyTaskList provides a mock function with given fields: ctx, getRequest
func (_m *HistoryClient) ResetStickyTaskList(ctx context.Context, request *history.ResetStickyTaskListRequest, opts ...yarpc.CallOption) (*history.ResetStickyTaskListResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *history.ResetStickyTaskListResponse
	if rf, ok := ret.Get(0).(func(context.Context, *history.ResetStickyTaskListRequest) *history.ResetStickyTaskListResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*history.ResetStickyTaskListResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *history.ResetStickyTaskListRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DescribeWorkflowExecution provides a mock function with given fields: ctx, request
func (_m *HistoryClient) DescribeWorkflowExecution(ctx context.Context, request *history.DescribeWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.DescribeWorkflowExecutionResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *shared.DescribeWorkflowExecutionResponse
	if rf, ok := ret.Get(0).(func(context.Context, *history.DescribeWorkflowExecutionRequest) *shared.DescribeWorkflowExecutionResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.DescribeWorkflowExecutionResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *history.DescribeWorkflowExecutionRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RecordActivityTaskHeartbeat provides a mock function with given fields: ctx, heartbeatRequest
func (_m *HistoryClient) RecordActivityTaskHeartbeat(ctx context.Context, heartbeatRequest *history.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	ret := _m.Called(ctx, heartbeatRequest)

	var r0 *shared.RecordActivityTaskHeartbeatResponse
	if rf, ok := ret.Get(0).(func(context.Context, *history.RecordActivityTaskHeartbeatRequest) *shared.RecordActivityTaskHeartbeatResponse); ok {
		r0 = rf(ctx, heartbeatRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.RecordActivityTaskHeartbeatResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *history.RecordActivityTaskHeartbeatRequest) error); ok {
		r1 = rf(ctx, heartbeatRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RecordActivityTaskStarted provides a mock function with given fields: ctx, addRequest
func (_m *HistoryClient) RecordActivityTaskStarted(ctx context.Context, addRequest *history.RecordActivityTaskStartedRequest, opts ...yarpc.CallOption) (*history.RecordActivityTaskStartedResponse, error) {
	ret := _m.Called(ctx, addRequest)

	var r0 *history.RecordActivityTaskStartedResponse
	if rf, ok := ret.Get(0).(func(context.Context, *history.RecordActivityTaskStartedRequest) *history.RecordActivityTaskStartedResponse); ok {
		r0 = rf(ctx, addRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*history.RecordActivityTaskStartedResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *history.RecordActivityTaskStartedRequest) error); ok {
		r1 = rf(ctx, addRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RecordDecisionTaskStarted provides a mock function with given fields: ctx, addRequest
func (_m *HistoryClient) RecordDecisionTaskStarted(ctx context.Context, addRequest *history.RecordDecisionTaskStartedRequest, opts ...yarpc.CallOption) (*history.RecordDecisionTaskStartedResponse, error) {
	ret := _m.Called(ctx, addRequest)

	var r0 *history.RecordDecisionTaskStartedResponse
	if rf, ok := ret.Get(0).(func(context.Context, *history.RecordDecisionTaskStartedRequest) *history.RecordDecisionTaskStartedResponse); ok {
		r0 = rf(ctx, addRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*history.RecordDecisionTaskStartedResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *history.RecordDecisionTaskStartedRequest) error); ok {
		r1 = rf(ctx, addRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RespondActivityTaskCanceled provides a mock function with given fields: ctx, canceledRequest
func (_m *HistoryClient) RespondActivityTaskCanceled(ctx context.Context, canceledRequest *history.RespondActivityTaskCanceledRequest, opts ...yarpc.CallOption) error {
	ret := _m.Called(ctx, canceledRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *history.RespondActivityTaskCanceledRequest) error); ok {
		r0 = rf(ctx, canceledRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RespondActivityTaskCompleted provides a mock function with given fields: ctx, completeRequest
func (_m *HistoryClient) RespondActivityTaskCompleted(ctx context.Context, completeRequest *history.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption) error {
	ret := _m.Called(ctx, completeRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *history.RespondActivityTaskCompletedRequest) error); ok {
		r0 = rf(ctx, completeRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RespondActivityTaskFailed provides a mock function with given fields: ctx, failRequest
func (_m *HistoryClient) RespondActivityTaskFailed(ctx context.Context, failRequest *history.RespondActivityTaskFailedRequest, opts ...yarpc.CallOption) error {
	ret := _m.Called(ctx, failRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *history.RespondActivityTaskFailedRequest) error); ok {
		r0 = rf(ctx, failRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RespondDecisionTaskCompleted provides a mock function with given fields: ctx, completeRequest
func (_m *HistoryClient) RespondDecisionTaskCompleted(ctx context.Context, completeRequest *history.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption) error {
	ret := _m.Called(ctx, completeRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *history.RespondDecisionTaskCompletedRequest) error); ok {
		r0 = rf(ctx, completeRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RespondDecisionTaskFailed provides a mock function with given fields: ctx, failedRequest
func (_m *HistoryClient) RespondDecisionTaskFailed(ctx context.Context, failedRequest *history.RespondDecisionTaskFailedRequest, opts ...yarpc.CallOption) error {
	ret := _m.Called(ctx, failedRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *history.RespondDecisionTaskFailedRequest) error); ok {
		r0 = rf(ctx, failedRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SignalWorkflowExecution provides a mock function with given fields: ctx, signalRequest
func (_m *HistoryClient) SignalWorkflowExecution(ctx context.Context, signalRequest *history.SignalWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	ret := _m.Called(ctx, signalRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *history.SignalWorkflowExecutionRequest) error); ok {
		r0 = rf(ctx, signalRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SignalWithStartWorkflowExecution provides a mock function with given fields: ctx, signalWithStartRequest
func (_m *HistoryClient) SignalWithStartWorkflowExecution(ctx context.Context,
	signalWithStartRequest *history.SignalWithStartWorkflowExecutionRequest,
	opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
	ret := _m.Called(ctx, signalWithStartRequest)

	var r0 *shared.StartWorkflowExecutionResponse
	if rf, ok := ret.Get(0).(func(context.Context, *history.SignalWithStartWorkflowExecutionRequest) *shared.StartWorkflowExecutionResponse); ok {
		r0 = rf(ctx, signalWithStartRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.StartWorkflowExecutionResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *history.SignalWithStartWorkflowExecutionRequest) error); ok {
		r1 = rf(ctx, signalWithStartRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveSignalMutableState provides a mock function with given fields: ctx, removeRequest
func (_m *HistoryClient) RemoveSignalMutableState(ctx context.Context, removeRequest *history.RemoveSignalMutableStateRequest, opts ...yarpc.CallOption) error {
	ret := _m.Called(ctx, removeRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *history.RemoveSignalMutableStateRequest) error); ok {
		r0 = rf(ctx, removeRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StartWorkflowExecution provides a mock function with given fields: ctx, startRequest
func (_m *HistoryClient) StartWorkflowExecution(ctx context.Context, startRequest *history.StartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
	ret := _m.Called(ctx, startRequest)

	var r0 *shared.StartWorkflowExecutionResponse
	if rf, ok := ret.Get(0).(func(context.Context, *history.StartWorkflowExecutionRequest) *shared.StartWorkflowExecutionResponse); ok {
		r0 = rf(ctx, startRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.StartWorkflowExecutionResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *history.StartWorkflowExecutionRequest) error); ok {
		r1 = rf(ctx, startRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RequestCancelWorkflowExecution is mock implementation for RequestCancelWorkflowExecution of HistoryEngine
func (_m *HistoryClient) RequestCancelWorkflowExecution(ctx context.Context, request *history.RequestCancelWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	ret := _m.Called(request)

	var r0 error
	if rf, ok := ret.Get(0).(func(*history.RequestCancelWorkflowExecutionRequest) error); ok {
		r0 = rf(request)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TerminateWorkflowExecution provides a mock function with given fields: ctx, terminateRequest
func (_m *HistoryClient) TerminateWorkflowExecution(ctx context.Context, terminateRequest *history.TerminateWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	ret := _m.Called(ctx, terminateRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *history.TerminateWorkflowExecutionRequest) error); ok {
		r0 = rf(ctx, terminateRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ScheduleDecisionTask provides a mock function with given fields: ctx, request
func (_m *HistoryClient) ScheduleDecisionTask(ctx context.Context, request *history.ScheduleDecisionTaskRequest, opts ...yarpc.CallOption) error {
	ret := _m.Called(ctx, request)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *history.ScheduleDecisionTaskRequest) error); ok {
		r0 = rf(ctx, request)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RecordChildExecutionCompleted provides a mock function with given fields: ctx, request
func (_m *HistoryClient) RecordChildExecutionCompleted(ctx context.Context, request *history.RecordChildExecutionCompletedRequest, opts ...yarpc.CallOption) error {
	ret := _m.Called(ctx, request)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *history.RecordChildExecutionCompletedRequest) error); ok {
		r0 = rf(ctx, request)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RecordChildExecutionCompleted provides a mock function with given fields: ctx, request
func (_m *HistoryClient) ReplicateEvents(ctx context.Context, request *history.ReplicateEventsRequest, opts ...yarpc.CallOption) error {
	ret := _m.Called(ctx, request)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *history.ReplicateEventsRequest) error); ok {
		r0 = rf(ctx, request)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
