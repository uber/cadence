package invariants

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
	"testing"
)

const (
	workflowTimeoutSecond = int32(110)
	taskTimeoutSecond     = int32(50)
)

func Test__Check(t *testing.T) {
	workflowTimeoutSecondInBytes, err := json.Marshal(workflowTimeoutSecond)
	taskTimeoutSecondInBytes, err := json.Marshal(taskTimeoutSecond)
	require.NoError(t, err)
	testCases := []struct {
		name           string
		testData       *types.GetWorkflowExecutionHistoryResponse
		expectedResult []InvariantCheckResult
		err            error
	}{
		{
			name:     "workflow execution timeout",
			testData: wfTimeoutHistory(),
			expectedResult: []InvariantCheckResult{
				{
					InvariantType: TimeoutTypeExecution.String(),
					Reason:        "START_TO_CLOSE",
					Metadata:      workflowTimeoutSecondInBytes,
				},
			},
			err: nil,
		},
		{
			name:     "child workflow execution timeout",
			testData: childWfTimeoutHistory(),
			expectedResult: []InvariantCheckResult{
				{
					InvariantType: TimeoutTypeChildWorkflow.String(),
					Reason:        "START_TO_CLOSE",
					Metadata:      workflowTimeoutSecondInBytes,
				},
			},
			err: nil,
		},
		{
			name:     "activity timeout",
			testData: activityTimeoutHistory(),
			expectedResult: []InvariantCheckResult{
				{
					InvariantType: TimeoutTypeActivity.String(),
					Reason:        "SCHEDULE_TO_START",
					Metadata:      taskTimeoutSecondInBytes,
				},
				{
					InvariantType: TimeoutTypeActivity.String(),
					Reason:        "HEARTBEAT",
					Metadata:      taskTimeoutSecondInBytes,
				},
			},
			err: nil,
		},
		{
			name:     "decision timeout",
			testData: decisionTimeoutHistory(),
			expectedResult: []InvariantCheckResult{
				{
					InvariantType: TimeoutTypeDecision.String(),
					Reason:        "START_TO_CLOSE",
					Metadata:      taskTimeoutSecondInBytes,
				},
				{
					InvariantType: TimeoutTypeDecision.String(),
					Reason:        "workflow reset",
					Metadata:      []byte("new run ID"),
				},
			},
			err: nil,
		},
	}
	for _, tc := range testCases {
		inv := NewTimeout(tc.testData)
		result, err := inv.Check(context.Background())
		require.Equal(t, tc.err, err)
		require.Equal(t, len(tc.expectedResult), len(result))
		for i := range result {
			require.Equal(t, tc.expectedResult[i], result[i])
		}

	}
}

func wfTimeoutHistory() *types.GetWorkflowExecutionHistoryResponse {
	return &types.GetWorkflowExecutionHistoryResponse{
		History: &types.History{
			Events: []*types.HistoryEvent{
				{
					ID: 1,
					WorkflowExecutionStartedEventAttributes: &types.WorkflowExecutionStartedEventAttributes{
						ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(workflowTimeoutSecond),
					},
				},
				{
					WorkflowExecutionTimedOutEventAttributes: &types.WorkflowExecutionTimedOutEventAttributes{TimeoutType: types.TimeoutTypeStartToClose.Ptr()},
				},
			},
		},
	}
}

func childWfTimeoutHistory() *types.GetWorkflowExecutionHistoryResponse {
	return &types.GetWorkflowExecutionHistoryResponse{
		History: &types.History{
			Events: []*types.HistoryEvent{
				{
					ID: 22,
					StartChildWorkflowExecutionInitiatedEventAttributes: &types.StartChildWorkflowExecutionInitiatedEventAttributes{
						ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(workflowTimeoutSecond),
					},
				},
				{
					ChildWorkflowExecutionTimedOutEventAttributes: &types.ChildWorkflowExecutionTimedOutEventAttributes{
						InitiatedEventID: 22,
						TimeoutType:      types.TimeoutTypeStartToClose.Ptr()},
				},
			},
		},
	}
}

func activityTimeoutHistory() *types.GetWorkflowExecutionHistoryResponse {
	return &types.GetWorkflowExecutionHistoryResponse{
		History: &types.History{
			Events: []*types.HistoryEvent{
				{
					ID: 5,
					ActivityTaskScheduledEventAttributes: &types.ActivityTaskScheduledEventAttributes{
						ScheduleToStartTimeoutSeconds: common.Int32Ptr(taskTimeoutSecond),
					},
				},
				{
					ActivityTaskTimedOutEventAttributes: &types.ActivityTaskTimedOutEventAttributes{
						ScheduledEventID: 5,
						StartedEventID:   6,
						TimeoutType:      types.TimeoutTypeScheduleToStart.Ptr(),
					},
				},
				{
					ID: 21,
					ActivityTaskScheduledEventAttributes: &types.ActivityTaskScheduledEventAttributes{
						HeartbeatTimeoutSeconds: common.Int32Ptr(taskTimeoutSecond),
					},
				},
				{
					ActivityTaskTimedOutEventAttributes: &types.ActivityTaskTimedOutEventAttributes{
						ScheduledEventID: 21,
						StartedEventID:   22,
						TimeoutType:      types.TimeoutTypeHeartbeat.Ptr(),
					},
				},
			},
		},
	}
}

func decisionTimeoutHistory() *types.GetWorkflowExecutionHistoryResponse {
	return &types.GetWorkflowExecutionHistoryResponse{
		History: &types.History{
			Events: []*types.HistoryEvent{
				{
					ID: 13,
					DecisionTaskScheduledEventAttributes: &types.DecisionTaskScheduledEventAttributes{
						StartToCloseTimeoutSeconds: common.Int32Ptr(taskTimeoutSecond),
					},
				},
				{
					DecisionTaskTimedOutEventAttributes: &types.DecisionTaskTimedOutEventAttributes{
						ScheduledEventID: 13,
						StartedEventID:   14,
						Cause:            types.DecisionTaskTimedOutCauseTimeout.Ptr(),
						TimeoutType:      types.TimeoutTypeStartToClose.Ptr(),
					},
				},
				{
					DecisionTaskTimedOutEventAttributes: &types.DecisionTaskTimedOutEventAttributes{
						ScheduledEventID: 3,
						StartedEventID:   14,
						Cause:            types.DecisionTaskTimedOutCauseReset.Ptr(),
						Reason:           "workflow reset",
						NewRunID:         "new run ID",
					},
				},
			},
		},
	}
}
