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

package invariants

import (
	"context"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/uber/cadence/client"
	"github.com/uber/cadence/client/frontend"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
)

const (
	workflowTimeoutSecond = int32(110)
	taskTimeoutSecond     = int32(50)
	testTimeStamp         = int64(2547596872371000000)
	timeUnit              = time.Second
	testTasklist          = "test-tasklist"
	testDomain            = "test-domain"
	testTaskListBacklog   = int64(10)
)

func Test__Check(t *testing.T) {
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
					Metadata:      wfTimeoutDataInBytes(t),
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
					Metadata:      childWfTimeoutDataInBytes(t),
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
					Metadata:      activityTimeoutDataInBytes(t),
				},
				{
					InvariantType: TimeoutTypeActivity.String(),
					Reason:        "HEARTBEAT",
					Metadata:      activityHeartBeatTimeoutDataInBytes(t),
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
	ctrl := gomock.NewController(t)
	mockClientBean := client.NewMockBean(ctrl)
	for _, tc := range testCases {
		inv := NewTimeout(NewTimeoutParams{
			WorkflowExecutionHistory: tc.testData,
			Domain:                   testDomain,
			ClientBean:               mockClientBean,
		})
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
					ID:        1,
					Timestamp: common.Int64Ptr(testTimeStamp),
					WorkflowExecutionStartedEventAttributes: &types.WorkflowExecutionStartedEventAttributes{
						ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(workflowTimeoutSecond),
						TaskList: &types.TaskList{
							Name: testTasklist,
							Kind: nil,
						},
					},
				},
				{
					ID:                                       2,
					Timestamp:                                common.Int64Ptr(testTimeStamp + int64(workflowTimeoutSecond)*timeUnit.Nanoseconds()),
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
					ID: 1,
					StartChildWorkflowExecutionInitiatedEventAttributes: &types.StartChildWorkflowExecutionInitiatedEventAttributes{
						ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(workflowTimeoutSecond),
					},
				},
				{
					ID:        2,
					Timestamp: common.Int64Ptr(testTimeStamp),
					ChildWorkflowExecutionStartedEventAttributes: &types.ChildWorkflowExecutionStartedEventAttributes{
						InitiatedEventID: 1,
					},
				},
				{
					ID:        3,
					Timestamp: common.Int64Ptr(testTimeStamp + int64(workflowTimeoutSecond)*timeUnit.Nanoseconds()),
					ChildWorkflowExecutionTimedOutEventAttributes: &types.ChildWorkflowExecutionTimedOutEventAttributes{
						InitiatedEventID: 1,
						StartedEventID:   2,
						TimeoutType:      types.TimeoutTypeStartToClose.Ptr(),
						WorkflowExecution: &types.WorkflowExecution{
							WorkflowID: "123",
							RunID:      "abc",
						},
					},
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
					ID:        1,
					Timestamp: common.Int64Ptr(testTimeStamp),
					ActivityTaskScheduledEventAttributes: &types.ActivityTaskScheduledEventAttributes{
						ScheduleToStartTimeoutSeconds: common.Int32Ptr(taskTimeoutSecond),
						TaskList: &types.TaskList{
							Name: testTasklist,
							Kind: nil,
						},
					},
				},
				{
					ID:        2,
					Timestamp: common.Int64Ptr(testTimeStamp + int64(taskTimeoutSecond)*timeUnit.Nanoseconds()),
					ActivityTaskTimedOutEventAttributes: &types.ActivityTaskTimedOutEventAttributes{
						ScheduledEventID: 1,
						TimeoutType:      types.TimeoutTypeScheduleToStart.Ptr(),
					},
				},
				{
					ID: 3,
					ActivityTaskScheduledEventAttributes: &types.ActivityTaskScheduledEventAttributes{
						HeartbeatTimeoutSeconds: common.Int32Ptr(taskTimeoutSecond),
						TaskList: &types.TaskList{
							Name: testTasklist,
							Kind: nil,
						},
					},
				},
				{
					ID:        4,
					Timestamp: common.Int64Ptr(testTimeStamp),
					ActivityTaskStartedEventAttributes: &types.ActivityTaskStartedEventAttributes{
						ScheduledEventID: 21,
					},
				},
				{
					ID:        5,
					Timestamp: common.Int64Ptr(testTimeStamp + int64(taskTimeoutSecond)*timeUnit.Nanoseconds()),
					ActivityTaskTimedOutEventAttributes: &types.ActivityTaskTimedOutEventAttributes{
						ScheduledEventID: 3,
						StartedEventID:   4,
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
						Cause:    types.DecisionTaskTimedOutCauseReset.Ptr(),
						Reason:   "workflow reset",
						NewRunID: "new run ID",
					},
				},
			},
		},
	}
}

func wfTimeoutDataInBytes(t *testing.T) []byte {
	data := ExecutionTimeoutMetadata{
		ExecutionTime:     110 * time.Second,
		ConfiguredTimeout: 110 * time.Second,
		LastOngoingEvent: &types.HistoryEvent{
			ID:        1,
			Timestamp: common.Int64Ptr(testTimeStamp),
			WorkflowExecutionStartedEventAttributes: &types.WorkflowExecutionStartedEventAttributes{
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(workflowTimeoutSecond),
				TaskList: &types.TaskList{
					Name: testTasklist,
					Kind: nil,
				},
			},
		},
		Tasklist: &types.TaskList{
			Name: testTasklist,
			Kind: nil,
		},
	}
	dataInBytes, err := json.Marshal(data)
	require.NoError(t, err)
	return dataInBytes
}

func activityTimeoutData() ActivityTimeoutMetadata {
	return ActivityTimeoutMetadata{
		TimeoutType:       types.TimeoutTypeScheduleToStart.Ptr(),
		ConfiguredTimeout: 50 * time.Second,
		TimeElapsed:       50 * time.Second,
		RetryPolicy:       nil,
		HeartBeatTimeout:  0,
		Tasklist: &types.TaskList{
			Name: testTasklist,
			Kind: nil,
		},
	}
}
func activityTimeoutDataInBytes(t *testing.T) []byte {
	data := activityTimeoutData()
	dataInBytes, err := json.Marshal(data)
	require.NoError(t, err)
	return dataInBytes
}

func activityHeartBeatTimeoutDataInBytes(t *testing.T) []byte {
	actTimeoutData := activityTimeoutData()
	actTimeoutData.TimeoutType = types.TimeoutTypeHeartbeat.Ptr()
	actTimeoutData.HeartBeatTimeout = 50 * time.Second
	actHeartBeatTimeoutDataInBytes, err := json.Marshal(actTimeoutData)
	require.NoError(t, err)
	return actHeartBeatTimeoutDataInBytes
}

func childWfTimeoutDataInBytes(t *testing.T) []byte {
	data := ChildWfTimeoutMetadata{
		ExecutionTime:     110 * time.Second,
		ConfiguredTimeout: 110 * time.Second,
		Execution: &types.WorkflowExecution{
			WorkflowID: "123",
			RunID:      "abc",
		},
	}
	dataInBytes, err := json.Marshal(data)
	require.NoError(t, err)
	return dataInBytes
}

func Test__RootCause(t *testing.T) {
	actTimeoutData := activityTimeoutData()
	testCases := []struct {
		name           string
		input          []InvariantCheckResult
		clientExpects  func(*frontend.MockClient)
		expectedResult []InvariantRootCauseResult
		err            error
	}{
		{
			name: "workflow execution timeout without pollers",
			input: []InvariantCheckResult{
				{
					InvariantType: TimeoutTypeExecution.String(),
					Reason:        "START_TO_CLOSE",
					Metadata:      wfTimeoutDataInBytes(t),
				},
			},
			clientExpects: func(client *frontend.MockClient) {
				client.EXPECT().DescribeTaskList(gomock.Any(), gomock.Any()).Return(&types.DescribeTaskListResponse{
					Pollers: nil,
					TaskListStatus: &types.TaskListStatus{
						BacklogCountHint: testTaskListBacklog,
					},
				}, nil)
			},
			expectedResult: []InvariantRootCauseResult{
				{
					RootCause: RootCauseTypeMissingPollers.String(),
					Metadata:  taskListBacklogInBytes(testTaskListBacklog),
				},
			},
			err: nil,
		},
		{
			name: "workflow execution timeout with pollers",
			input: []InvariantCheckResult{
				{
					InvariantType: TimeoutTypeExecution.String(),
					Reason:        "START_TO_CLOSE",
					Metadata:      wfTimeoutDataInBytes(t),
				},
			},
			clientExpects: func(client *frontend.MockClient) {
				client.EXPECT().DescribeTaskList(gomock.Any(), gomock.Any()).Return(&types.DescribeTaskListResponse{
					Pollers: []*types.PollerInfo{
						{
							Identity: "dca24-xy",
						},
					},
					TaskListStatus: &types.TaskListStatus{
						BacklogCountHint: testTaskListBacklog,
					},
				}, nil)
			},
			expectedResult: []InvariantRootCauseResult{
				{
					RootCause: RootCauseTypePollersStatus.String(),
					Metadata:  taskListBacklogInBytes(testTaskListBacklog),
				},
			},
			err: nil,
		},
		{
			name: "activity timeout and heart beating not enabled",
			input: []InvariantCheckResult{
				{
					InvariantType: TimeoutTypeActivity.String(),
					Reason:        "START_TO_CLOSE",
					Metadata:      activityTimeoutDataInBytes(t),
				},
			},
			clientExpects: func(client *frontend.MockClient) {
				client.EXPECT().DescribeTaskList(gomock.Any(), gomock.Any()).Return(&types.DescribeTaskListResponse{
					Pollers: []*types.PollerInfo{
						{
							Identity: "dca24-xy",
						},
					},
					TaskListStatus: &types.TaskListStatus{
						BacklogCountHint: testTaskListBacklog,
					},
				}, nil)
			},
			expectedResult: []InvariantRootCauseResult{
				{
					RootCause: RootCauseTypePollersStatus.String(),
					Metadata:  taskListBacklogInBytes(testTaskListBacklog),
				},
				{
					RootCause: RootCauseTypeHeartBeatingNotEnabled.String(),
					Metadata:  []byte(actTimeoutData.TimeElapsed.String()),
				},
			},
			err: nil,
		},
		{
			name: "activity timeout and heart beating not enabled",
			input: []InvariantCheckResult{
				{
					InvariantType: TimeoutTypeActivity.String(),
					Reason:        "START_TO_CLOSE",
					Metadata:      activityTimeoutDataInBytes(t),
				},
			},
			clientExpects: func(client *frontend.MockClient) {
				client.EXPECT().DescribeTaskList(gomock.Any(), gomock.Any()).Return(&types.DescribeTaskListResponse{
					Pollers: []*types.PollerInfo{
						{
							Identity: "dca24-xy",
						},
					},
					TaskListStatus: &types.TaskListStatus{
						BacklogCountHint: testTaskListBacklog,
					},
				}, nil)
			},
			expectedResult: []InvariantRootCauseResult{
				{
					RootCause: RootCauseTypePollersStatus.String(),
					Metadata:  taskListBacklogInBytes(testTaskListBacklog),
				},
				{
					RootCause: RootCauseTypeHeartBeatingNotEnabled.String(),
					Metadata:  []byte(actTimeoutData.TimeElapsed.String()),
				},
			},
			err: nil,
		},
		{
			name: "activity timeout and heart beating enabled",
			input: []InvariantCheckResult{
				{
					InvariantType: TimeoutTypeActivity.String(),
					Reason:        "START_TO_CLOSE",
					Metadata:      activityHeartBeatTimeoutDataInBytes(t),
				},
			},
			clientExpects: func(client *frontend.MockClient) {
				client.EXPECT().DescribeTaskList(gomock.Any(), gomock.Any()).Return(&types.DescribeTaskListResponse{
					Pollers: []*types.PollerInfo{
						{
							Identity: "dca24-xy",
						},
					},
					TaskListStatus: &types.TaskListStatus{
						BacklogCountHint: testTaskListBacklog,
					},
				}, nil)
			},
			expectedResult: []InvariantRootCauseResult{
				{
					RootCause: RootCauseTypePollersStatus.String(),
					Metadata:  taskListBacklogInBytes(testTaskListBacklog),
				},
				{
					RootCause: RootCauseTypeHeartBeatingEnabledMissingHeartbeat.String(),
					Metadata:  []byte(actTimeoutData.TimeElapsed.String()),
				},
			},
			err: nil,
		},
	}
	ctrl := gomock.NewController(t)
	mockClientBean := client.NewMockBean(ctrl)
	mockFrontendClient := frontend.NewMockClient(ctrl)
	mockClientBean.EXPECT().GetFrontendClient().Return(mockFrontendClient).AnyTimes()
	inv := NewTimeout(NewTimeoutParams{
		Domain:     testDomain,
		ClientBean: mockClientBean,
	})
	for _, tc := range testCases {
		tc.clientExpects(mockFrontendClient)
		result, err := inv.RootCause(context.Background(), tc.input)
		require.Equal(t, tc.err, err)
		require.Equal(t, len(tc.expectedResult), len(result))
		for i := range result {
			require.Equal(t, tc.expectedResult[i], result[i])
		}

	}
}
