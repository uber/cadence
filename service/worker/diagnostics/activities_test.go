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

package diagnostics

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/client"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/worker/diagnostics/invariants"
)

func Test__retrieveExecutionHistory(t *testing.T) {
	dwtest := testDiagnosticWorkflow(t)
	result, err := dwtest.retrieveExecutionHistory(context.Background(), retrieveExecutionHistoryInputParams{
		domain: "test",
		execution: &types.WorkflowExecution{
			WorkflowID: "123",
			RunID:      "abc",
		},
	})
	require.NoError(t, err)
	require.Equal(t, testWorkflowExecutionHistoryResponse(), result)
}

func Test__identifyTimeouts(t *testing.T) {
	dwtest := testDiagnosticWorkflow(t)
	workflowTimeoutSecondInBytes, err := json.Marshal(int32(10))
	require.NoError(t, err)
	expectedResult := []invariants.InvariantCheckResult{
		{
			InvariantType: invariants.TimeoutTypeExecution.String(),
			Reason:        "START_TO_CLOSE",
			Metadata:      workflowTimeoutSecondInBytes,
		},
	}
	result, err := dwtest.identifyTimeouts(context.Background(), identifyTimeoutsInputParams{history: testWorkflowExecutionHistoryResponse()})
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func testDiagnosticWorkflow(t *testing.T) *dw {
	ctrl := gomock.NewController(t)
	mockClientBean := client.NewMockBean(ctrl)
	mockFrontendClient := frontend.NewMockClient(ctrl)
	mockClientBean.EXPECT().GetFrontendClient().Return(mockFrontendClient).AnyTimes()
	mockFrontendClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any()).Return(testWorkflowExecutionHistoryResponse(), nil).AnyTimes()
	return &dw{
		clientBean: mockClientBean,
	}
}

func testWorkflowExecutionHistoryResponse() *types.GetWorkflowExecutionHistoryResponse {
	return &types.GetWorkflowExecutionHistoryResponse{
		History: &types.History{
			Events: []*types.HistoryEvent{
				{
					ID: 1,
					WorkflowExecutionStartedEventAttributes: &types.WorkflowExecutionStartedEventAttributes{
						ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(10),
					},
				},
				{
					WorkflowExecutionTimedOutEventAttributes: &types.WorkflowExecutionTimedOutEventAttributes{TimeoutType: types.TimeoutTypeStartToClose.Ptr()},
				},
			},
		},
	}
}
