package diagnostics

import (
	"context"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/uber/cadence/client"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/worker/diagnostics/invariants"
	"testing"
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
