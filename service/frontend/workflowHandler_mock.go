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

package frontend

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/uber/cadence/.gen/go/shared"
)

// MockWorkflowHandler is an autogenerated mock type for the Interface type
type MockWorkflowHandler struct {
	mock.Mock
}

// CountWorkflowExecutions provides a mock function with given fields: ctx, CountRequest
func (_m *MockWorkflowHandler) CountWorkflowExecutions(ctx context.Context, CountRequest *shared.CountWorkflowExecutionsRequest) (*shared.CountWorkflowExecutionsResponse, error) {
	ret := _m.Called(ctx, CountRequest)

	var r0 *shared.CountWorkflowExecutionsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.CountWorkflowExecutionsRequest) *shared.CountWorkflowExecutionsResponse); ok {
		r0 = rf(ctx, CountRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.CountWorkflowExecutionsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.CountWorkflowExecutionsRequest) error); ok {
		r1 = rf(ctx, CountRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeprecateDomain provides a mock function with given fields: ctx, DeprecateRequest
func (_m *MockWorkflowHandler) DeprecateDomain(ctx context.Context, DeprecateRequest *shared.DeprecateDomainRequest) error {
	ret := _m.Called(ctx, DeprecateRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *shared.DeprecateDomainRequest) error); ok {
		r0 = rf(ctx, DeprecateRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DescribeDomain provides a mock function with given fields: ctx, DescribeRequest
func (_m *MockWorkflowHandler) DescribeDomain(ctx context.Context, DescribeRequest *shared.DescribeDomainRequest) (*shared.DescribeDomainResponse, error) {
	ret := _m.Called(ctx, DescribeRequest)

	var r0 *shared.DescribeDomainResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.DescribeDomainRequest) *shared.DescribeDomainResponse); ok {
		r0 = rf(ctx, DescribeRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.DescribeDomainResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.DescribeDomainRequest) error); ok {
		r1 = rf(ctx, DescribeRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DescribeTaskList provides a mock function with given fields: ctx, Request
func (_m *MockWorkflowHandler) DescribeTaskList(ctx context.Context, Request *shared.DescribeTaskListRequest) (*shared.DescribeTaskListResponse, error) {
	ret := _m.Called(ctx, Request)

	var r0 *shared.DescribeTaskListResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.DescribeTaskListRequest) *shared.DescribeTaskListResponse); ok {
		r0 = rf(ctx, Request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.DescribeTaskListResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.DescribeTaskListRequest) error); ok {
		r1 = rf(ctx, Request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DescribeWorkflowExecution provides a mock function with given fields: ctx, DescribeRequest
func (_m *MockWorkflowHandler) DescribeWorkflowExecution(ctx context.Context, DescribeRequest *shared.DescribeWorkflowExecutionRequest) (*shared.DescribeWorkflowExecutionResponse, error) {
	ret := _m.Called(ctx, DescribeRequest)

	var r0 *shared.DescribeWorkflowExecutionResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.DescribeWorkflowExecutionRequest) *shared.DescribeWorkflowExecutionResponse); ok {
		r0 = rf(ctx, DescribeRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.DescribeWorkflowExecutionResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.DescribeWorkflowExecutionRequest) error); ok {
		r1 = rf(ctx, DescribeRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSearchAttributes provides a mock function with given fields: ctx
func (_m *MockWorkflowHandler) GetSearchAttributes(ctx context.Context) (*shared.GetSearchAttributesResponse, error) {
	ret := _m.Called(ctx)

	var r0 *shared.GetSearchAttributesResponse
	if rf, ok := ret.Get(0).(func(context.Context) *shared.GetSearchAttributesResponse); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.GetSearchAttributesResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetWorkflowExecutionHistory provides a mock function with given fields: ctx, GetRequest
func (_m *MockWorkflowHandler) GetWorkflowExecutionHistory(ctx context.Context, GetRequest *shared.GetWorkflowExecutionHistoryRequest) (*shared.GetWorkflowExecutionHistoryResponse, error) {
	ret := _m.Called(ctx, GetRequest)

	var r0 *shared.GetWorkflowExecutionHistoryResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.GetWorkflowExecutionHistoryRequest) *shared.GetWorkflowExecutionHistoryResponse); ok {
		r0 = rf(ctx, GetRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.GetWorkflowExecutionHistoryResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.GetWorkflowExecutionHistoryRequest) error); ok {
		r1 = rf(ctx, GetRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListClosedWorkflowExecutions provides a mock function with given fields: ctx, ListRequest
func (_m *MockWorkflowHandler) ListClosedWorkflowExecutions(ctx context.Context, ListRequest *shared.ListClosedWorkflowExecutionsRequest) (*shared.ListClosedWorkflowExecutionsResponse, error) {
	ret := _m.Called(ctx, ListRequest)

	var r0 *shared.ListClosedWorkflowExecutionsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.ListClosedWorkflowExecutionsRequest) *shared.ListClosedWorkflowExecutionsResponse); ok {
		r0 = rf(ctx, ListRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.ListClosedWorkflowExecutionsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.ListClosedWorkflowExecutionsRequest) error); ok {
		r1 = rf(ctx, ListRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListDomains provides a mock function with given fields: ctx, ListRequest
func (_m *MockWorkflowHandler) ListDomains(ctx context.Context, ListRequest *shared.ListDomainsRequest) (*shared.ListDomainsResponse, error) {
	ret := _m.Called(ctx, ListRequest)

	var r0 *shared.ListDomainsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.ListDomainsRequest) *shared.ListDomainsResponse); ok {
		r0 = rf(ctx, ListRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.ListDomainsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.ListDomainsRequest) error); ok {
		r1 = rf(ctx, ListRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListOpenWorkflowExecutions provides a mock function with given fields: ctx, ListRequest
func (_m *MockWorkflowHandler) ListOpenWorkflowExecutions(ctx context.Context, ListRequest *shared.ListOpenWorkflowExecutionsRequest) (*shared.ListOpenWorkflowExecutionsResponse, error) {
	ret := _m.Called(ctx, ListRequest)

	var r0 *shared.ListOpenWorkflowExecutionsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.ListOpenWorkflowExecutionsRequest) *shared.ListOpenWorkflowExecutionsResponse); ok {
		r0 = rf(ctx, ListRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.ListOpenWorkflowExecutionsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.ListOpenWorkflowExecutionsRequest) error); ok {
		r1 = rf(ctx, ListRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListWorkflowExecutions provides a mock function with given fields: ctx, ListRequest
func (_m *MockWorkflowHandler) ListWorkflowExecutions(ctx context.Context, ListRequest *shared.ListWorkflowExecutionsRequest) (*shared.ListWorkflowExecutionsResponse, error) {
	ret := _m.Called(ctx, ListRequest)

	var r0 *shared.ListWorkflowExecutionsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.ListWorkflowExecutionsRequest) *shared.ListWorkflowExecutionsResponse); ok {
		r0 = rf(ctx, ListRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.ListWorkflowExecutionsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.ListWorkflowExecutionsRequest) error); ok {
		r1 = rf(ctx, ListRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PollForActivityTask provides a mock function with given fields: ctx, PollRequest
func (_m *MockWorkflowHandler) PollForActivityTask(ctx context.Context, PollRequest *shared.PollForActivityTaskRequest) (*shared.PollForActivityTaskResponse, error) {
	ret := _m.Called(ctx, PollRequest)

	var r0 *shared.PollForActivityTaskResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.PollForActivityTaskRequest) *shared.PollForActivityTaskResponse); ok {
		r0 = rf(ctx, PollRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.PollForActivityTaskResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.PollForActivityTaskRequest) error); ok {
		r1 = rf(ctx, PollRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PollForDecisionTask provides a mock function with given fields: ctx, PollRequest
func (_m *MockWorkflowHandler) PollForDecisionTask(ctx context.Context, PollRequest *shared.PollForDecisionTaskRequest) (*shared.PollForDecisionTaskResponse, error) {
	ret := _m.Called(ctx, PollRequest)

	var r0 *shared.PollForDecisionTaskResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.PollForDecisionTaskRequest) *shared.PollForDecisionTaskResponse); ok {
		r0 = rf(ctx, PollRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.PollForDecisionTaskResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.PollForDecisionTaskRequest) error); ok {
		r1 = rf(ctx, PollRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// QueryWorkflow provides a mock function with given fields: ctx, QueryRequest
func (_m *MockWorkflowHandler) QueryWorkflow(ctx context.Context, QueryRequest *shared.QueryWorkflowRequest) (*shared.QueryWorkflowResponse, error) {
	ret := _m.Called(ctx, QueryRequest)

	var r0 *shared.QueryWorkflowResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.QueryWorkflowRequest) *shared.QueryWorkflowResponse); ok {
		r0 = rf(ctx, QueryRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.QueryWorkflowResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.QueryWorkflowRequest) error); ok {
		r1 = rf(ctx, QueryRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RecordActivityTaskHeartbeat provides a mock function with given fields: ctx, HeartbeatRequest
func (_m *MockWorkflowHandler) RecordActivityTaskHeartbeat(ctx context.Context, HeartbeatRequest *shared.RecordActivityTaskHeartbeatRequest) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	ret := _m.Called(ctx, HeartbeatRequest)

	var r0 *shared.RecordActivityTaskHeartbeatResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.RecordActivityTaskHeartbeatRequest) *shared.RecordActivityTaskHeartbeatResponse); ok {
		r0 = rf(ctx, HeartbeatRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.RecordActivityTaskHeartbeatResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.RecordActivityTaskHeartbeatRequest) error); ok {
		r1 = rf(ctx, HeartbeatRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RecordActivityTaskHeartbeatByID provides a mock function with given fields: ctx, HeartbeatRequest
func (_m *MockWorkflowHandler) RecordActivityTaskHeartbeatByID(ctx context.Context, HeartbeatRequest *shared.RecordActivityTaskHeartbeatByIDRequest) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	ret := _m.Called(ctx, HeartbeatRequest)

	var r0 *shared.RecordActivityTaskHeartbeatResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.RecordActivityTaskHeartbeatByIDRequest) *shared.RecordActivityTaskHeartbeatResponse); ok {
		r0 = rf(ctx, HeartbeatRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.RecordActivityTaskHeartbeatResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.RecordActivityTaskHeartbeatByIDRequest) error); ok {
		r1 = rf(ctx, HeartbeatRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RegisterDomain provides a mock function with given fields: ctx, RegisterRequest
func (_m *MockWorkflowHandler) RegisterDomain(ctx context.Context, RegisterRequest *shared.RegisterDomainRequest) error {
	ret := _m.Called(ctx, RegisterRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *shared.RegisterDomainRequest) error); ok {
		r0 = rf(ctx, RegisterRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RequestCancelWorkflowExecution provides a mock function with given fields: ctx, CancelRequest
func (_m *MockWorkflowHandler) RequestCancelWorkflowExecution(ctx context.Context, CancelRequest *shared.RequestCancelWorkflowExecutionRequest) error {
	ret := _m.Called(ctx, CancelRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *shared.RequestCancelWorkflowExecutionRequest) error); ok {
		r0 = rf(ctx, CancelRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ResetStickyTaskList provides a mock function with given fields: ctx, ResetRequest
func (_m *MockWorkflowHandler) ResetStickyTaskList(ctx context.Context, ResetRequest *shared.ResetStickyTaskListRequest) (*shared.ResetStickyTaskListResponse, error) {
	ret := _m.Called(ctx, ResetRequest)

	var r0 *shared.ResetStickyTaskListResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.ResetStickyTaskListRequest) *shared.ResetStickyTaskListResponse); ok {
		r0 = rf(ctx, ResetRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.ResetStickyTaskListResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.ResetStickyTaskListRequest) error); ok {
		r1 = rf(ctx, ResetRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ResetWorkflowExecution provides a mock function with given fields: ctx, ResetRequest
func (_m *MockWorkflowHandler) ResetWorkflowExecution(ctx context.Context, ResetRequest *shared.ResetWorkflowExecutionRequest) (*shared.ResetWorkflowExecutionResponse, error) {
	ret := _m.Called(ctx, ResetRequest)

	var r0 *shared.ResetWorkflowExecutionResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.ResetWorkflowExecutionRequest) *shared.ResetWorkflowExecutionResponse); ok {
		r0 = rf(ctx, ResetRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.ResetWorkflowExecutionResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.ResetWorkflowExecutionRequest) error); ok {
		r1 = rf(ctx, ResetRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RespondActivityTaskCanceled provides a mock function with given fields: ctx, CanceledRequest
func (_m *MockWorkflowHandler) RespondActivityTaskCanceled(ctx context.Context, CanceledRequest *shared.RespondActivityTaskCanceledRequest) error {
	ret := _m.Called(ctx, CanceledRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *shared.RespondActivityTaskCanceledRequest) error); ok {
		r0 = rf(ctx, CanceledRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RespondActivityTaskCanceledByID provides a mock function with given fields: ctx, CanceledRequest
func (_m *MockWorkflowHandler) RespondActivityTaskCanceledByID(ctx context.Context, CanceledRequest *shared.RespondActivityTaskCanceledByIDRequest) error {
	ret := _m.Called(ctx, CanceledRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *shared.RespondActivityTaskCanceledByIDRequest) error); ok {
		r0 = rf(ctx, CanceledRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RespondActivityTaskCompleted provides a mock function with given fields: ctx, CompleteRequest
func (_m *MockWorkflowHandler) RespondActivityTaskCompleted(ctx context.Context, CompleteRequest *shared.RespondActivityTaskCompletedRequest) error {
	ret := _m.Called(ctx, CompleteRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *shared.RespondActivityTaskCompletedRequest) error); ok {
		r0 = rf(ctx, CompleteRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RespondActivityTaskCompletedByID provides a mock function with given fields: ctx, CompleteRequest
func (_m *MockWorkflowHandler) RespondActivityTaskCompletedByID(ctx context.Context, CompleteRequest *shared.RespondActivityTaskCompletedByIDRequest) error {
	ret := _m.Called(ctx, CompleteRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *shared.RespondActivityTaskCompletedByIDRequest) error); ok {
		r0 = rf(ctx, CompleteRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RespondActivityTaskFailed provides a mock function with given fields: ctx, FailRequest
func (_m *MockWorkflowHandler) RespondActivityTaskFailed(ctx context.Context, FailRequest *shared.RespondActivityTaskFailedRequest) error {
	ret := _m.Called(ctx, FailRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *shared.RespondActivityTaskFailedRequest) error); ok {
		r0 = rf(ctx, FailRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RespondActivityTaskFailedByID provides a mock function with given fields: ctx, FailRequest
func (_m *MockWorkflowHandler) RespondActivityTaskFailedByID(ctx context.Context, FailRequest *shared.RespondActivityTaskFailedByIDRequest) error {
	ret := _m.Called(ctx, FailRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *shared.RespondActivityTaskFailedByIDRequest) error); ok {
		r0 = rf(ctx, FailRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RespondDecisionTaskCompleted provides a mock function with given fields: ctx, CompleteRequest
func (_m *MockWorkflowHandler) RespondDecisionTaskCompleted(ctx context.Context, CompleteRequest *shared.RespondDecisionTaskCompletedRequest) (*shared.RespondDecisionTaskCompletedResponse, error) {
	ret := _m.Called(ctx, CompleteRequest)

	var r0 *shared.RespondDecisionTaskCompletedResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.RespondDecisionTaskCompletedRequest) *shared.RespondDecisionTaskCompletedResponse); ok {
		r0 = rf(ctx, CompleteRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.RespondDecisionTaskCompletedResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.RespondDecisionTaskCompletedRequest) error); ok {
		r1 = rf(ctx, CompleteRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RespondDecisionTaskFailed provides a mock function with given fields: ctx, FailedRequest
func (_m *MockWorkflowHandler) RespondDecisionTaskFailed(ctx context.Context, FailedRequest *shared.RespondDecisionTaskFailedRequest) error {
	ret := _m.Called(ctx, FailedRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *shared.RespondDecisionTaskFailedRequest) error); ok {
		r0 = rf(ctx, FailedRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RespondQueryTaskCompleted provides a mock function with given fields: ctx, CompleteRequest
func (_m *MockWorkflowHandler) RespondQueryTaskCompleted(ctx context.Context, CompleteRequest *shared.RespondQueryTaskCompletedRequest) error {
	ret := _m.Called(ctx, CompleteRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *shared.RespondQueryTaskCompletedRequest) error); ok {
		r0 = rf(ctx, CompleteRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ScanWorkflowExecutions provides a mock function with given fields: ctx, ListRequest
func (_m *MockWorkflowHandler) ScanWorkflowExecutions(ctx context.Context, ListRequest *shared.ListWorkflowExecutionsRequest) (*shared.ListWorkflowExecutionsResponse, error) {
	ret := _m.Called(ctx, ListRequest)

	var r0 *shared.ListWorkflowExecutionsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.ListWorkflowExecutionsRequest) *shared.ListWorkflowExecutionsResponse); ok {
		r0 = rf(ctx, ListRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.ListWorkflowExecutionsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.ListWorkflowExecutionsRequest) error); ok {
		r1 = rf(ctx, ListRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SignalWithStartWorkflowExecution provides a mock function with given fields: ctx, SignalWithStartRequest
func (_m *MockWorkflowHandler) SignalWithStartWorkflowExecution(ctx context.Context, SignalWithStartRequest *shared.SignalWithStartWorkflowExecutionRequest) (*shared.StartWorkflowExecutionResponse, error) {
	ret := _m.Called(ctx, SignalWithStartRequest)

	var r0 *shared.StartWorkflowExecutionResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.SignalWithStartWorkflowExecutionRequest) *shared.StartWorkflowExecutionResponse); ok {
		r0 = rf(ctx, SignalWithStartRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.StartWorkflowExecutionResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.SignalWithStartWorkflowExecutionRequest) error); ok {
		r1 = rf(ctx, SignalWithStartRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SignalWorkflowExecution provides a mock function with given fields: ctx, SignalRequest
func (_m *MockWorkflowHandler) SignalWorkflowExecution(ctx context.Context, SignalRequest *shared.SignalWorkflowExecutionRequest) error {
	ret := _m.Called(ctx, SignalRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *shared.SignalWorkflowExecutionRequest) error); ok {
		r0 = rf(ctx, SignalRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StartWorkflowExecution provides a mock function with given fields: ctx, StartRequest
func (_m *MockWorkflowHandler) StartWorkflowExecution(ctx context.Context, StartRequest *shared.StartWorkflowExecutionRequest) (*shared.StartWorkflowExecutionResponse, error) {
	ret := _m.Called(ctx, StartRequest)

	var r0 *shared.StartWorkflowExecutionResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.StartWorkflowExecutionRequest) *shared.StartWorkflowExecutionResponse); ok {
		r0 = rf(ctx, StartRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.StartWorkflowExecutionResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.StartWorkflowExecutionRequest) error); ok {
		r1 = rf(ctx, StartRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TerminateWorkflowExecution provides a mock function with given fields: ctx, TerminateRequest
func (_m *MockWorkflowHandler) TerminateWorkflowExecution(ctx context.Context, TerminateRequest *shared.TerminateWorkflowExecutionRequest) error {
	ret := _m.Called(ctx, TerminateRequest)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *shared.TerminateWorkflowExecutionRequest) error); ok {
		r0 = rf(ctx, TerminateRequest)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateDomain provides a mock function with given fields: ctx, UpdateRequest
func (_m *MockWorkflowHandler) UpdateDomain(ctx context.Context, UpdateRequest *shared.UpdateDomainRequest) (*shared.UpdateDomainResponse, error) {
	ret := _m.Called(ctx, UpdateRequest)

	var r0 *shared.UpdateDomainResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.UpdateDomainRequest) *shared.UpdateDomainResponse); ok {
		r0 = rf(ctx, UpdateRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.UpdateDomainResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.UpdateDomainRequest) error); ok {
		r1 = rf(ctx, UpdateRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
