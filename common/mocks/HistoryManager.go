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

import "github.com/uber/cadence/common/persistence"
import "github.com/stretchr/testify/mock"

type HistoryManager struct {
	mock.Mock
}

// GetName provides a mock function with given fields:
func (_m *HistoryManager) GetName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// AppendHistoryEvents provides a mock function with given fields: request
func (_m *HistoryManager) AppendHistoryEvents(request *persistence.AppendHistoryEventsRequest) (*persistence.AppendHistoryEventsResponse, error) {
	ret := _m.Called(request)

	var r0 *persistence.AppendHistoryEventsResponse
	if rf, ok := ret.Get(0).(func(*persistence.AppendHistoryEventsRequest) *persistence.AppendHistoryEventsResponse); ok {
		r0 = rf(request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*persistence.AppendHistoryEventsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*persistence.AppendHistoryEventsRequest) error); ok {
		r1 = rf(request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetWorkflowExecutionHistory provides a mock function with given fields: request
func (_m *HistoryManager) GetWorkflowExecutionHistory(request *persistence.GetWorkflowExecutionHistoryRequest) (*persistence.GetWorkflowExecutionHistoryResponse, error) {
	ret := _m.Called(request)

	var r0 *persistence.GetWorkflowExecutionHistoryResponse
	if rf, ok := ret.Get(0).(func(*persistence.GetWorkflowExecutionHistoryRequest) *persistence.GetWorkflowExecutionHistoryResponse); ok {
		r0 = rf(request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*persistence.GetWorkflowExecutionHistoryResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*persistence.GetWorkflowExecutionHistoryRequest) error); ok {
		r1 = rf(request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteWorkflowExecutionHistory provides a mock function with given fields: request
func (_m *HistoryManager) DeleteWorkflowExecutionHistory(request *persistence.DeleteWorkflowExecutionHistoryRequest) error {
	ret := _m.Called(request)

	var r0 error
	if rf, ok := ret.Get(0).(func(*persistence.DeleteWorkflowExecutionHistoryRequest) error); ok {
		r0 = rf(request)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewHistoryBranch provides a mock function with given fields: request
func (_m *HistoryManager) NewHistoryBranch(request *persistence.NewHistoryBranchRequest) error {
	ret := _m.Called(request)
	var r0 error
	if rf, ok := ret.Get(0).(func(*persistence.NewHistoryBranchRequest) error); ok {
		r0 = rf(request)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// AppendHistoryNode provides a mock function with given fields: request
func (_m *HistoryManager) AppendHistoryNode(request *persistence.AppendHistoryNodeRequest) error {
	ret := _m.Called(request)
	var r0 error
	if rf, ok := ret.Get(0).(func(*persistence.AppendHistoryNodeRequest) error); ok {
		r0 = rf(request)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// ReadHistoryBranch provides a mock function with given fields: request
func (_m *HistoryManager) ReadHistoryBranch(request *persistence.ReadHistoryBranchRequest) (*persistence.ReadHistoryBranchResponse, error) {
	ret := _m.Called(request)
	var r0 *persistence.ReadHistoryBranchResponse
	if rf, ok := ret.Get(0).(func(*persistence.ReadHistoryBranchRequest) *persistence.ReadHistoryBranchResponse); ok {
		r0 = rf(request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*persistence.ReadHistoryBranchResponse)
		}
	}
	var r1 error
	if rf, ok := ret.Get(1).(func(*persistence.ReadHistoryBranchRequest) error); ok {
		r1 = rf(request)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// ForkHistoryBranch provides a mock function with given fields: request
func (_m *HistoryManager) ForkHistoryBranch(request *persistence.ForkHistoryBranchRequest) (*persistence.ForkHistoryBranchResponse, error) {
	ret := _m.Called(request)
	var r0 *persistence.ForkHistoryBranchResponse
	if rf, ok := ret.Get(0).(func(*persistence.ForkHistoryBranchRequest) *persistence.ForkHistoryBranchResponse); ok {
		r0 = rf(request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*persistence.ForkHistoryBranchResponse)
		}
	}
	var r1 error
	if rf, ok := ret.Get(1).(func(*persistence.ForkHistoryBranchRequest) error); ok {
		r1 = rf(request)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// DeleteHistoryBranch provides a mock function with given fields: request
func (_m *HistoryManager) DeleteHistoryBranch(request *persistence.DeleteHistoryBranchRequest) error {
	ret := _m.Called(request)
	var r0 error
	if rf, ok := ret.Get(0).(func(*persistence.DeleteHistoryBranchRequest) error); ok {
		r0 = rf(request)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// GetHistoryTree provides a mock function with given fields: request
func (_m *HistoryManager) GetHistoryTree(request *persistence.GetHistoryTreeRequest) (*persistence.GetHistoryTreeResponse, error) {
	ret := _m.Called(request)
	var r0 *persistence.GetHistoryTreeResponse
	if rf, ok := ret.Get(0).(func(*persistence.GetHistoryTreeRequest) *persistence.GetHistoryTreeResponse); ok {
		r0 = rf(request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*persistence.GetHistoryTreeResponse)
		}
	}
	var r1 error
	if rf, ok := ret.Get(1).(func(*persistence.GetHistoryTreeRequest) error); ok {
		r1 = rf(request)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Close provides a mock function with given fields:
func (_m *HistoryManager) Close() {
	_m.Called()
}

var _ persistence.HistoryManager = (*HistoryManager)(nil)
