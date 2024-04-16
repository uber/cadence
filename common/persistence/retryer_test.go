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

package persistence

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/types"
	"testing"
	"time"
)

func TestPersistenceRetryerListConcreteExecutions(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := map[string]struct {
		request                        *ListConcreteExecutionsRequest
		mockExecutionManager           ExecutionManager
		mockExecutionManagerAccordance func(mockExecutionManager *MockExecutionManager)
		expectedResponse               *ListConcreteExecutionsResponse
		expectedError                  error
	}{
		"Case 1: Success": {
			request:              &ListConcreteExecutionsRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().ListConcreteExecutions(gomock.Any(), gomock.Eq(&ListConcreteExecutionsRequest{})).Return(&ListConcreteExecutionsResponse{}, nil)
			},
			expectedResponse: &ListConcreteExecutionsResponse{},
			expectedError:    nil,
		},
		"Case 2: Transient Error": {
			request:              &ListConcreteExecutionsRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				gomock.InOrder(
					mockExecutionManager.EXPECT().ListConcreteExecutions(gomock.Any(), gomock.Eq(&ListConcreteExecutionsRequest{})).Return(nil, &types.InternalServiceError{}),
					mockExecutionManager.EXPECT().ListConcreteExecutions(gomock.Any(), gomock.Eq(&ListConcreteExecutionsRequest{})).Return(&ListConcreteExecutionsResponse{}, nil),
				)
			},
			expectedResponse: &ListConcreteExecutionsResponse{},
			expectedError:    nil,
		},
		"Case 3: Fatal Error": {
			request:              &ListConcreteExecutionsRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().ListConcreteExecutions(gomock.Any(), gomock.Eq(&ListConcreteExecutionsRequest{})).Return(nil, &types.AccessDeniedError{}).Times(1)
			},
			expectedResponse: nil,
			expectedError:    &types.AccessDeniedError{},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockExecutionManager != nil {
				test.mockExecutionManagerAccordance(test.mockExecutionManager.(*MockExecutionManager))
			}
			retryer := NewPersistenceRetryer(test.mockExecutionManager, NewMockHistoryManager(ctrl), backoff.NewExponentialRetryPolicy(time.Nanosecond))

			resp, err := retryer.ListConcreteExecutions(context.Background(), test.request)
			assert.Equal(t, test.expectedResponse, resp)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestPersistenceRetryerGetWorkflowExecution(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := map[string]struct {
		request                        *GetWorkflowExecutionRequest
		mockExecutionManager           ExecutionManager
		mockExecutionManagerAccordance func(mockExecutionManager *MockExecutionManager)
		expectedResponse               *GetWorkflowExecutionResponse
		expectedError                  error
	}{
		"Case 1: Success": {
			request:              &GetWorkflowExecutionRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().GetWorkflowExecution(gomock.Any(), gomock.Eq(&GetWorkflowExecutionRequest{})).Return(&GetWorkflowExecutionResponse{}, nil)
			},
			expectedResponse: &GetWorkflowExecutionResponse{},
			expectedError:    nil,
		},
		"Case 2: Transient Error": {
			request:              &GetWorkflowExecutionRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				gomock.InOrder(
					mockExecutionManager.EXPECT().GetWorkflowExecution(gomock.Any(), gomock.Eq(&GetWorkflowExecutionRequest{})).Return(nil, &types.InternalServiceError{}),
					mockExecutionManager.EXPECT().GetWorkflowExecution(gomock.Any(), gomock.Eq(&GetWorkflowExecutionRequest{})).Return(&GetWorkflowExecutionResponse{}, nil),
				)
			},
			expectedResponse: &GetWorkflowExecutionResponse{},
			expectedError:    nil,
		},
		"Case 3: Fatal Error": {
			request:              &GetWorkflowExecutionRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().GetWorkflowExecution(gomock.Any(), gomock.Eq(&GetWorkflowExecutionRequest{})).Return(nil, &types.AccessDeniedError{}).Times(1)
			},
			expectedResponse: nil,
			expectedError:    &types.AccessDeniedError{},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockExecutionManager != nil {
				test.mockExecutionManagerAccordance(test.mockExecutionManager.(*MockExecutionManager))
			}
			retryer := NewPersistenceRetryer(test.mockExecutionManager, NewMockHistoryManager(ctrl), backoff.NewExponentialRetryPolicy(time.Nanosecond))

			resp, err := retryer.GetWorkflowExecution(context.Background(), test.request)
			assert.Equal(t, test.expectedResponse, resp)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestPersistenceRetryerGetCurrentExecution(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := map[string]struct {
		request                        *GetCurrentExecutionRequest
		mockExecutionManager           ExecutionManager
		mockExecutionManagerAccordance func(mockExecutionManager *MockExecutionManager)
		expectedResponse               *GetCurrentExecutionResponse
		expectedError                  error
	}{
		"Case 1: Success": {
			request:              &GetCurrentExecutionRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().GetCurrentExecution(gomock.Any(), gomock.Eq(&GetCurrentExecutionRequest{})).Return(&GetCurrentExecutionResponse{}, nil)
			},
			expectedResponse: &GetCurrentExecutionResponse{},
			expectedError:    nil,
		},
		"Case 2: Transient Error": {
			request:              &GetCurrentExecutionRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				gomock.InOrder(
					mockExecutionManager.EXPECT().GetCurrentExecution(gomock.Any(), gomock.Eq(&GetCurrentExecutionRequest{})).Return(nil, &types.InternalServiceError{}),
					mockExecutionManager.EXPECT().GetCurrentExecution(gomock.Any(), gomock.Eq(&GetCurrentExecutionRequest{})).Return(&GetCurrentExecutionResponse{}, nil),
				)
			},
			expectedResponse: &GetCurrentExecutionResponse{},
			expectedError:    nil,
		},
		"Case 3: Fatal Error": {
			request:              &GetCurrentExecutionRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().GetCurrentExecution(gomock.Any(), gomock.Eq(&GetCurrentExecutionRequest{})).Return(nil, &types.AccessDeniedError{}).Times(1)
			},
			expectedResponse: nil,
			expectedError:    &types.AccessDeniedError{},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockExecutionManager != nil {
				test.mockExecutionManagerAccordance(test.mockExecutionManager.(*MockExecutionManager))
			}
			retryer := NewPersistenceRetryer(test.mockExecutionManager, NewMockHistoryManager(ctrl), backoff.NewExponentialRetryPolicy(time.Nanosecond))

			resp, err := retryer.GetCurrentExecution(context.Background(), test.request)
			assert.Equal(t, test.expectedResponse, resp)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestPersistenceRetryerListCurrentExecutions(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := map[string]struct {
		request                        *ListCurrentExecutionsRequest
		mockExecutionManager           ExecutionManager
		mockExecutionManagerAccordance func(mockExecutionManager *MockExecutionManager)
		expectedResponse               *ListCurrentExecutionsResponse
		expectedError                  error
	}{
		"Case 1: Success": {
			request:              &ListCurrentExecutionsRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().ListCurrentExecutions(gomock.Any(), gomock.Eq(&ListCurrentExecutionsRequest{})).Return(&ListCurrentExecutionsResponse{}, nil)
			},
			expectedResponse: &ListCurrentExecutionsResponse{},
			expectedError:    nil,
		},
		"Case 2: Transient Error": {
			request:              &ListCurrentExecutionsRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				gomock.InOrder(
					mockExecutionManager.EXPECT().ListCurrentExecutions(gomock.Any(), gomock.Eq(&ListCurrentExecutionsRequest{})).Return(nil, &types.InternalServiceError{}),
					mockExecutionManager.EXPECT().ListCurrentExecutions(gomock.Any(), gomock.Eq(&ListCurrentExecutionsRequest{})).Return(&ListCurrentExecutionsResponse{}, nil),
				)
			},
			expectedResponse: &ListCurrentExecutionsResponse{},
			expectedError:    nil,
		},
		"Case 3: Fatal Error": {
			request:              &ListCurrentExecutionsRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().ListCurrentExecutions(gomock.Any(), gomock.Eq(&ListCurrentExecutionsRequest{})).Return(nil, &types.AccessDeniedError{}).Times(1)
			},
			expectedResponse: nil,
			expectedError:    &types.AccessDeniedError{},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockExecutionManager != nil {
				test.mockExecutionManagerAccordance(test.mockExecutionManager.(*MockExecutionManager))
			}
			retryer := NewPersistenceRetryer(test.mockExecutionManager, NewMockHistoryManager(ctrl), backoff.NewExponentialRetryPolicy(time.Nanosecond))

			resp, err := retryer.ListCurrentExecutions(context.Background(), test.request)
			assert.Equal(t, test.expectedResponse, resp)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestPersistenceRetryerIsWorkflowExecutionExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := map[string]struct {
		request                        *IsWorkflowExecutionExistsRequest
		mockExecutionManager           ExecutionManager
		mockExecutionManagerAccordance func(mockExecutionManager *MockExecutionManager)
		expectedResponse               *IsWorkflowExecutionExistsResponse
		expectedError                  error
	}{
		"Case 1: Success": {
			request:              &IsWorkflowExecutionExistsRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().IsWorkflowExecutionExists(gomock.Any(), gomock.Eq(&IsWorkflowExecutionExistsRequest{})).Return(&IsWorkflowExecutionExistsResponse{}, nil)
			},
			expectedResponse: &IsWorkflowExecutionExistsResponse{},
			expectedError:    nil,
		},
		"Case 2: Transient Error": {
			request:              &IsWorkflowExecutionExistsRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				gomock.InOrder(
					mockExecutionManager.EXPECT().IsWorkflowExecutionExists(gomock.Any(), gomock.Eq(&IsWorkflowExecutionExistsRequest{})).Return(nil, &types.InternalServiceError{}),
					mockExecutionManager.EXPECT().IsWorkflowExecutionExists(gomock.Any(), gomock.Eq(&IsWorkflowExecutionExistsRequest{})).Return(&IsWorkflowExecutionExistsResponse{}, nil),
				)
			},
			expectedResponse: &IsWorkflowExecutionExistsResponse{},
			expectedError:    nil,
		},
		"Case 3: Fatal Error": {
			request:              &IsWorkflowExecutionExistsRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().IsWorkflowExecutionExists(gomock.Any(), gomock.Eq(&IsWorkflowExecutionExistsRequest{})).Return(nil, &types.AccessDeniedError{}).Times(1)
			},
			expectedResponse: nil,
			expectedError:    &types.AccessDeniedError{},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockExecutionManager != nil {
				test.mockExecutionManagerAccordance(test.mockExecutionManager.(*MockExecutionManager))
			}
			retryer := NewPersistenceRetryer(test.mockExecutionManager, NewMockHistoryManager(ctrl), backoff.NewExponentialRetryPolicy(time.Nanosecond))

			resp, err := retryer.IsWorkflowExecutionExists(context.Background(), test.request)
			assert.Equal(t, test.expectedResponse, resp)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestPersistenceRetryerReadHistoryBranch(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := map[string]struct {
		request                      *ReadHistoryBranchRequest
		mockHistoryManager           HistoryManager
		mockHistoryManagerAccordance func(mockHistoryManager *MockHistoryManager)
		expectedResponse             *ReadHistoryBranchResponse
		expectedError                error
	}{
		"Case 1: Success": {
			request:            &ReadHistoryBranchRequest{},
			mockHistoryManager: NewMockHistoryManager(ctrl),
			mockHistoryManagerAccordance: func(mockHistoryManager *MockHistoryManager) {
				mockHistoryManager.EXPECT().ReadHistoryBranch(gomock.Any(), gomock.Eq(&ReadHistoryBranchRequest{})).Return(&ReadHistoryBranchResponse{}, nil)
			},
			expectedResponse: &ReadHistoryBranchResponse{},
			expectedError:    nil,
		},
		"Case 2: Transient Error": {
			request:            &ReadHistoryBranchRequest{},
			mockHistoryManager: NewMockHistoryManager(ctrl),
			mockHistoryManagerAccordance: func(mockHistoryManager *MockHistoryManager) {
				gomock.InOrder(
					mockHistoryManager.EXPECT().ReadHistoryBranch(gomock.Any(), gomock.Eq(&ReadHistoryBranchRequest{})).Return(nil, &types.InternalServiceError{}),
					mockHistoryManager.EXPECT().ReadHistoryBranch(gomock.Any(), gomock.Eq(&ReadHistoryBranchRequest{})).Return(&ReadHistoryBranchResponse{}, nil),
				)
			},
			expectedResponse: &ReadHistoryBranchResponse{},
			expectedError:    nil,
		},
		"Case 3: Fatal Error": {
			request:            &ReadHistoryBranchRequest{},
			mockHistoryManager: NewMockHistoryManager(ctrl),
			mockHistoryManagerAccordance: func(mockHistoryManager *MockHistoryManager) {
				mockHistoryManager.EXPECT().ReadHistoryBranch(gomock.Any(), gomock.Eq(&ReadHistoryBranchRequest{})).Return(nil, &types.AccessDeniedError{}).Times(1)
			},
			expectedResponse: nil,
			expectedError:    &types.AccessDeniedError{},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockHistoryManager != nil {
				test.mockHistoryManagerAccordance(test.mockHistoryManager.(*MockHistoryManager))
			}
			retryer := NewPersistenceRetryer(NewMockExecutionManager(ctrl), test.mockHistoryManager, backoff.NewExponentialRetryPolicy(time.Nanosecond))

			resp, err := retryer.ReadHistoryBranch(context.Background(), test.request)
			assert.Equal(t, test.expectedResponse, resp)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestPersistenceRetryerDeleteWorkflowExecution(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := map[string]struct {
		request                        *DeleteWorkflowExecutionRequest
		mockExecutionManager           ExecutionManager
		mockExecutionManagerAccordance func(mockExecutionManager *MockExecutionManager)
		expectedError                  error
	}{
		"Case 1: Success": {
			request:              &DeleteWorkflowExecutionRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().DeleteWorkflowExecution(gomock.Any(), gomock.Eq(&DeleteWorkflowExecutionRequest{})).Return(nil)
			},
			expectedError: nil,
		},
		"Case 2: Transient Error": {
			request:              &DeleteWorkflowExecutionRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				gomock.InOrder(
					mockExecutionManager.EXPECT().DeleteWorkflowExecution(gomock.Any(), gomock.Eq(&DeleteWorkflowExecutionRequest{})).Return(&types.InternalServiceError{}),
					mockExecutionManager.EXPECT().DeleteWorkflowExecution(gomock.Any(), gomock.Eq(&DeleteWorkflowExecutionRequest{})).Return(nil),
				)
			},
			expectedError: nil,
		},
		"Case 3: Fatal Error": {
			request:              &DeleteWorkflowExecutionRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().DeleteWorkflowExecution(gomock.Any(), gomock.Eq(&DeleteWorkflowExecutionRequest{})).Return(&types.AccessDeniedError{}).Times(1)
			},
			expectedError: &types.AccessDeniedError{},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockExecutionManager != nil {
				test.mockExecutionManagerAccordance(test.mockExecutionManager.(*MockExecutionManager))
			}
			retryer := NewPersistenceRetryer(test.mockExecutionManager, NewMockHistoryManager(ctrl), backoff.NewExponentialRetryPolicy(time.Nanosecond))

			err := retryer.DeleteWorkflowExecution(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPersistenceRetryerDeleteCurrentWorkflowExecution(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := map[string]struct {
		request                        *DeleteCurrentWorkflowExecutionRequest
		mockExecutionManager           ExecutionManager
		mockExecutionManagerAccordance func(mockExecutionManager *MockExecutionManager)
		expectedError                  error
	}{
		"Case 1: Success": {
			request:              &DeleteCurrentWorkflowExecutionRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().DeleteCurrentWorkflowExecution(gomock.Any(), gomock.Eq(&DeleteCurrentWorkflowExecutionRequest{})).Return(nil)
			},
			expectedError: nil,
		},
		"Case 2: Transient Error": {
			request:              &DeleteCurrentWorkflowExecutionRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				gomock.InOrder(
					mockExecutionManager.EXPECT().DeleteCurrentWorkflowExecution(gomock.Any(), gomock.Eq(&DeleteCurrentWorkflowExecutionRequest{})).Return(&types.InternalServiceError{}),
					mockExecutionManager.EXPECT().DeleteCurrentWorkflowExecution(gomock.Any(), gomock.Eq(&DeleteCurrentWorkflowExecutionRequest{})).Return(nil),
				)
			},
			expectedError: nil,
		},
		"Case 3: Fatal Error": {
			request:              &DeleteCurrentWorkflowExecutionRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().DeleteCurrentWorkflowExecution(gomock.Any(), gomock.Eq(&DeleteCurrentWorkflowExecutionRequest{})).Return(&types.AccessDeniedError{}).Times(1)
			},
			expectedError: &types.AccessDeniedError{},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockExecutionManager != nil {
				test.mockExecutionManagerAccordance(test.mockExecutionManager.(*MockExecutionManager))
			}
			retryer := NewPersistenceRetryer(test.mockExecutionManager, NewMockHistoryManager(ctrl), backoff.NewExponentialRetryPolicy(time.Nanosecond))

			err := retryer.DeleteCurrentWorkflowExecution(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPersistenceRetryerGetShardID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockExecutionManager := NewMockExecutionManager(ctrl)
	mockExecutionManager.EXPECT().GetShardID().Return(42)
	retryer := NewPersistenceRetryer(mockExecutionManager, NewMockHistoryManager(ctrl), backoff.NewExponentialRetryPolicy(time.Nanosecond))

	assert.Equal(t, 42, retryer.GetShardID())
}

func TestPersistenceRetryerGetTimerIndexTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := map[string]struct {
		request                        *GetTimerIndexTasksRequest
		mockExecutionManager           ExecutionManager
		mockExecutionManagerAccordance func(mockExecutionManager *MockExecutionManager)
		expectedResponse               *GetTimerIndexTasksResponse
		expectedError                  error
	}{
		"Case 1: Success": {
			request:              &GetTimerIndexTasksRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().GetTimerIndexTasks(gomock.Any(), gomock.Eq(&GetTimerIndexTasksRequest{})).Return(&GetTimerIndexTasksResponse{}, nil)
			},
			expectedResponse: &GetTimerIndexTasksResponse{},
			expectedError:    nil,
		},
		"Case 2: Transient Error": {
			request:              &GetTimerIndexTasksRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				gomock.InOrder(
					mockExecutionManager.EXPECT().GetTimerIndexTasks(gomock.Any(), gomock.Eq(&GetTimerIndexTasksRequest{})).Return(nil, &types.InternalServiceError{}),
					mockExecutionManager.EXPECT().GetTimerIndexTasks(gomock.Any(), gomock.Eq(&GetTimerIndexTasksRequest{})).Return(&GetTimerIndexTasksResponse{}, nil),
				)
			},
			expectedResponse: &GetTimerIndexTasksResponse{},
			expectedError:    nil,
		},
		"Case 3: Fatal Error": {
			request:              &GetTimerIndexTasksRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().GetTimerIndexTasks(gomock.Any(), gomock.Eq(&GetTimerIndexTasksRequest{})).Return(nil, &types.AccessDeniedError{}).Times(1)
			},
			expectedResponse: nil,
			expectedError:    &types.AccessDeniedError{},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockExecutionManager != nil {
				test.mockExecutionManagerAccordance(test.mockExecutionManager.(*MockExecutionManager))
			}
			retryer := NewPersistenceRetryer(test.mockExecutionManager, NewMockHistoryManager(ctrl), backoff.NewExponentialRetryPolicy(time.Nanosecond))

			resp, err := retryer.GetTimerIndexTasks(context.Background(), test.request)
			assert.Equal(t, test.expectedResponse, resp)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestPersistenceRetryerCompleteTimerTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := map[string]struct {
		request                        *CompleteTimerTaskRequest
		mockExecutionManager           ExecutionManager
		mockExecutionManagerAccordance func(mockExecutionManager *MockExecutionManager)
		expectedError                  error
	}{
		"Case 1: Success": {
			request:              &CompleteTimerTaskRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().CompleteTimerTask(gomock.Any(), gomock.Eq(&CompleteTimerTaskRequest{})).Return(nil)
			},
			expectedError: nil,
		},
		"Case 2: Transient Error": {
			request:              &CompleteTimerTaskRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				gomock.InOrder(
					mockExecutionManager.EXPECT().CompleteTimerTask(gomock.Any(), gomock.Eq(&CompleteTimerTaskRequest{})).Return(&types.InternalServiceError{}),
					mockExecutionManager.EXPECT().CompleteTimerTask(gomock.Any(), gomock.Eq(&CompleteTimerTaskRequest{})).Return(nil),
				)
			},
			expectedError: nil,
		},
		"Case 3: Fatal Error": {
			request:              &CompleteTimerTaskRequest{},
			mockExecutionManager: NewMockExecutionManager(ctrl),
			mockExecutionManagerAccordance: func(mockExecutionManager *MockExecutionManager) {
				mockExecutionManager.EXPECT().CompleteTimerTask(gomock.Any(), gomock.Eq(&CompleteTimerTaskRequest{})).Return(&types.AccessDeniedError{}).Times(1)
			},
			expectedError: &types.AccessDeniedError{},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockExecutionManager != nil {
				test.mockExecutionManagerAccordance(test.mockExecutionManager.(*MockExecutionManager))
			}
			retryer := NewPersistenceRetryer(test.mockExecutionManager, NewMockHistoryManager(ctrl), backoff.NewExponentialRetryPolicy(time.Nanosecond))

			err := retryer.CompleteTimerTask(context.Background(), test.request)
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
