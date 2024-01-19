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

package taskvalidator

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/service/history/constants"
)

type MockStaleChecker struct {
	CheckAgeFunc func(response *persistence.GetWorkflowExecutionResponse) (bool, error)
}

func (m *MockStaleChecker) CheckAge(response *persistence.GetWorkflowExecutionResponse) (bool, error) {
	return m.CheckAgeFunc(response)
}

func TestWorkflowCheckforValidation(t *testing.T) {
	testCases := []struct {
		name          string
		workflowID    string
		domainID      string
		domainName    string
		runID         string
		isStale       bool
		simulateError bool
	}{
		{"NonStaleWorkflow", "workflow-1", "domain-1", "domain-name-1", "run-1", false, false},
		{"ErrorInGetWorkflowExecution", "workflow-3", "domain-3", "domain-name-3", "run-3", false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockLogger := zap.NewNop()
			mockMetricsClient := metrics.NewNoopMetricsClient()
			mockDomainCache := cache.NewMockDomainCache(mockCtrl)
			mockPersistenceRetryer := persistence.NewMockRetryer(mockCtrl)
			mockStaleChecker := &MockStaleChecker{
				CheckAgeFunc: func(response *persistence.GetWorkflowExecutionResponse) (bool, error) {
					return tc.isStale, nil
				},
			}
			checker := NewWfChecker(mockLogger, mockMetricsClient, mockDomainCache, mockPersistenceRetryer)
			checker.(*checkerImpl).staleCheck = mockStaleChecker
			mockDomainCache.EXPECT().
				GetDomainByID(tc.domainID).
				Return(constants.TestGlobalDomainEntry, nil).AnyTimes()

			if tc.isStale {
				mockPersistenceRetryer.EXPECT().
					DeleteWorkflowExecution(gomock.Any(), gomock.Any()).
					Return(nil).Times(1)
			}

			mockPersistenceRetryer.EXPECT().
				GetWorkflowExecution(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, request *persistence.GetWorkflowExecutionRequest) (*persistence.GetWorkflowExecutionResponse, error) {
					if tc.simulateError {
						return nil, errors.New("database error")
					}
					// Return a valid response object to trigger the deletion calls
					return &persistence.GetWorkflowExecutionResponse{
						State: &persistence.WorkflowMutableState{
							ExecutionInfo: &persistence.WorkflowExecutionInfo{
								DomainID:   constants.TestDomainID,
								WorkflowID: constants.TestWorkflowID,
							},
						},
					}, nil
				}).AnyTimes()

			ctx := context.Background()
			err := checker.WorkflowCheckforValidation(ctx, tc.workflowID, tc.domainID, tc.domainName, tc.runID)

			if tc.simulateError {
				assert.Error(t, err, "Expected error when GetWorkflowExecution fails")
			} else {
				assert.NoError(t, err, "Expected no error for valid workflow execution")
			}
		})
	}
}
