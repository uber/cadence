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

package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
)

func TestConstructStartWorkflowRequest(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	set.String(FlagDomain, "test-domain", "domain")
	set.String(FlagTaskList, "test-task-list", "tasklist")
	set.String(FlagWorkflowType, "test-workflow-type", "workflow-type")
	set.Int("execution_timeout", 100, "execution_timeout")
	set.Int("decision_timeout", 50, "decision_timeout")
	set.String("workflow_id", "test-workflow-id", "workflow_id")
	set.Int("workflow_id_reuse_policy", 1, "workflow_id_reuse_policy")
	set.String("input", "{}", "input")
	set.String("cron_schedule", "* * * * *", "cron_schedule")
	set.Int("retry_attempts", 5, "retry_attempts")
	set.Int("retry_expiration", 600, "retry_expiration")
	set.Int("retry_interval", 10, "retry_interval")
	set.Float64("retry_backoff", 2.0, "retry_backoff")
	set.Int("retry_max_interval", 100, "retry_max_interval")
	set.Int(DelayStartSeconds, 5, DelayStartSeconds)
	set.Int(JitterStartSeconds, 2, JitterStartSeconds)
	set.String("first_run_at_time", "2024-07-24T12:00:00Z", "first-run-at-time")

	c := cli.NewContext(nil, set, nil)
	// inject context with span
	tracer := mocktracer.New()
	span, ctx := opentracing.StartSpanFromContextWithTracer(context.Background(), tracer, "test-span")
	span.SetBaggageItem("tracer-test-key", "tracer-test-value")
	defer span.Finish()
	c.Context = ctx

	assert.NoError(t, c.Set(FlagDomain, "test-domain"))
	assert.NoError(t, c.Set(FlagTaskList, "test-task-list"))
	assert.NoError(t, c.Set(FlagWorkflowType, "test-workflow-type"))
	assert.NoError(t, c.Set("execution_timeout", "100"))
	assert.NoError(t, c.Set("decision_timeout", "50"))
	assert.NoError(t, c.Set("workflow_id", "test-workflow-id"))
	assert.NoError(t, c.Set("workflow_id_reuse_policy", "1"))
	assert.NoError(t, c.Set("input", "{}"))
	assert.NoError(t, c.Set("cron_schedule", "* * * * *"))
	assert.NoError(t, c.Set("retry_attempts", "5"))
	assert.NoError(t, c.Set("retry_expiration", "600"))
	assert.NoError(t, c.Set("retry_interval", "10"))
	assert.NoError(t, c.Set("retry_backoff", "2.0"))
	assert.NoError(t, c.Set("retry_max_interval", "100"))
	assert.NoError(t, c.Set(DelayStartSeconds, "5"))
	assert.NoError(t, c.Set(JitterStartSeconds, "2"))
	assert.NoError(t, c.Set("first_run_at_time", "2024-07-24T12:00:00Z"))

	request, err := constructStartWorkflowRequest(c)
	assert.NoError(t, err)
	assert.NotNil(t, request)
	assert.Equal(t, "test-domain", request.Domain)
	assert.Equal(t, "test-task-list", request.TaskList.Name)
	assert.Equal(t, "test-workflow-type", request.WorkflowType.Name)
	assert.Equal(t, int32(100), *request.ExecutionStartToCloseTimeoutSeconds)
	assert.Equal(t, int32(50), *request.TaskStartToCloseTimeoutSeconds)
	assert.Equal(t, "test-workflow-id", request.WorkflowID)
	assert.NotNil(t, request.WorkflowIDReusePolicy)
	assert.Equal(t, int32(5), *request.DelayStartSeconds)
	assert.Equal(t, int32(2), *request.JitterStartSeconds)
	assert.Contains(t, request.Header.Fields, "mockpfx-baggage-tracer-test-key")
	assert.Equal(t, []byte("tracer-test-value"), request.Header.Fields["mockpfx-baggage-tracer-test-key"])

	firstRunAt, err := time.Parse(time.RFC3339, "2024-07-24T12:00:00Z")
	assert.NoError(t, err)
	assert.Equal(t, firstRunAt.UnixNano(), *request.FirstRunAtTimeStamp)
}

func Test_PrintAutoResetPoints(t *testing.T) {
	tests := []struct {
		name string
		resp *types.DescribeWorkflowExecutionResponse
	}{
		{
			name: "empty reset points",
			resp: &types.DescribeWorkflowExecutionResponse{
				WorkflowExecutionInfo: &types.WorkflowExecutionInfo{},
			},
		},
		{
			name: "normal case",
			resp: &types.DescribeWorkflowExecutionResponse{
				WorkflowExecutionInfo: &types.WorkflowExecutionInfo{
					AutoResetPoints: &types.ResetPoints{
						Points: []*types.ResetPointInfo{
							{
								BinaryChecksum: "test-binary-checksum",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := printAutoResetPoints(tt.resp)
			assert.NoError(t, err)
		})
	}
}

func Test_DescribeWorkflow(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)
	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})
	serverFrontendClient.EXPECT().DescribeWorkflowExecution(gomock.Any(), gomock.Any()).Return(&types.DescribeWorkflowExecutionResponse{
		WorkflowExecutionInfo: &types.WorkflowExecutionInfo{
			Execution: &types.WorkflowExecution{
				WorkflowID: "test-workflow-id",
				RunID:      "test-run-id",
			},
		},
	}, nil).Times(1)
	c := getMockContext(t, nil, app)
	err := DescribeWorkflow(c)
	assert.NoError(t, err)
}

func Test_DescribeWorkflowWithID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)
	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})
	serverFrontendClient.EXPECT().DescribeWorkflowExecution(gomock.Any(), gomock.Any()).Return(&types.DescribeWorkflowExecutionResponse{
		WorkflowExecutionInfo: &types.WorkflowExecutionInfo{
			Execution: &types.WorkflowExecution{
				WorkflowID: "test-workflow-id",
				RunID:      "test-run-id",
			},
		},
	}, nil).Times(1)
	c := getMockContext(t, nil, app)
	err := DescribeWorkflowWithID(c)
	assert.NoError(t, err)
}

func Test_DescribeWorkflowWithID_Error(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	err := DescribeWorkflowWithID(cli.NewContext(nil, set, nil))
	assert.Error(t, err)

	// WF helper describe failed
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)
	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})
	serverFrontendClient.EXPECT().DescribeWorkflowExecution(gomock.Any(), gomock.Any()).Return(&types.DescribeWorkflowExecutionResponse{
		WorkflowExecutionInfo: &types.WorkflowExecutionInfo{
			Execution: &types.WorkflowExecution{
				WorkflowID: "test-workflow-id",
				RunID:      "test-run-id",
			},
			SearchAttributes: &types.SearchAttributes{
				IndexedFields: map[string][]byte{
					"CustomKeywordField": []byte("test-value"),
				},
			},
		},
	}, nil).Times(1)
	serverFrontendClient.EXPECT().GetSearchAttributes(gomock.Any()).Return(nil, errors.New("test-error")).Times(1)

	c := getMockContext(t, nil, app)
	err = DescribeWorkflowWithID(c)
	assert.Error(t, err)
}

func getMockContext(t *testing.T, set *flag.FlagSet, app *cli.App) *cli.Context {
	if set == nil {
		set = flag.NewFlagSet("test", 0)
		set.String(FlagDomain, "test-domain", "domain")
		set.String("workflow_id", "test-workflow-id", "workflow_id")
		set.String("run_id", "test-run-id", "run_id")
		set.Bool("print_reset_points", true, "print_reset_points")
		set.Parse([]string{"test-workflow-id", "test-run-id"})
	}

	c := cli.NewContext(app, set, nil)
	assert.NoError(t, c.Set(FlagDomain, "test-domain"))
	assert.NoError(t, c.Set("workflow_id", "test-workflow-id"))
	assert.NoError(t, c.Set("run_id", "test-run-id"))

	return c
}

func Test_ListAllWorkflow(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
	})
	serverFrontendClient.EXPECT().CountWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&types.CountWorkflowExecutionsResponse{
		Count: int64(1),
	}, nil).AnyTimes()
	serverFrontendClient.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&types.ListWorkflowExecutionsResponse{}, nil).AnyTimes()
	serverFrontendClient.EXPECT().ListClosedWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&types.ListClosedWorkflowExecutionsResponse{}, nil).AnyTimes()
	set := flag.NewFlagSet("test", 0)
	set.String(FlagDomain, "test-domain", "domain")
	set.String("workflow_id", "test-workflow-id", "workflow_id")
	set.String("run_id", "test-run-id", "run_id")
	set.String("status", "open", "status")
	c := getMockContext(t, set, app)
	err := ListAllWorkflow(c)
	assert.NoError(t, err)
}

func Test_ConvertSearchAttributesToMapOfInterface(t *testing.T) {
	tests := []struct {
		name          string
		in            *types.SearchAttributes
		out           map[string]interface{}
		expectedError bool
		mockResponse  *types.GetSearchAttributesResponse
		mockError     error
	}{
		{
			name:          "empty search attributes",
			out:           nil,
			expectedError: false,
		},
		{
			name: "error when get search attributes",
			in: &types.SearchAttributes{
				IndexedFields: map[string][]byte{
					"CustomKeywordField": []byte("test-value"),
				},
			},
			out:           nil,
			mockError:     errors.New("test-error"),
			expectedError: true,
		},
		{
			name: "error deserialize search attributes",
			in: &types.SearchAttributes{
				IndexedFields: map[string][]byte{
					"CustomKeywordField": []byte("test-value"),
				},
			},
			out:           nil,
			mockError:     nil,
			expectedError: true,
		},
		{
			name: "normal case",
			in: &types.SearchAttributes{
				IndexedFields: map[string][]byte{
					"CustomKeywordField": []byte(`"test-value"`), // Assuming the value is a serialized string
				},
			},
			out: map[string]interface{}{
				"CustomKeywordField": "test-value", // Expected deserialized value
			},
			mockResponse: &types.GetSearchAttributesResponse{
				Keys: map[string]types.IndexedValueType{
					"CustomKeywordField": types.IndexedValueTypeKeyword,
				},
			},
			mockError:     nil,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			serverFrontendClient := frontend.NewMockClient(mockCtrl)
			serverAdminClient := admin.NewMockClient(mockCtrl)
			app := NewCliApp(&clientFactoryMock{
				serverFrontendClient: serverFrontendClient,
				serverAdminClient:    serverAdminClient,
			})
			c := getMockContext(t, nil, app)
			serverFrontendClient.EXPECT().GetSearchAttributes(gomock.Any()).Return(tt.mockResponse, tt.mockError).AnyTimes()
			out, err := convertSearchAttributesToMapOfInterface(tt.in, serverFrontendClient, c)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.out, out)
			}
		})
	}
}

func Test_GetAllWorkflowIDsByQuery(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)
	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})
	// missing required flag
	set := flag.NewFlagSet("test", 0)
	set.String("workflow_id", "test-workflow-id", "workflow_id")
	c := cli.NewContext(app, set, nil)
	_, err := getAllWorkflowIDsByQuery(c, "WorkflowType='test-workflow-type'")
	assert.Error(t, err)

	c = getMockContext(t, nil, app)
	serverFrontendClient.EXPECT().ScanWorkflowExecutions(gomock.Any(), &types.ListWorkflowExecutionsRequest{
		Query:    "WorkflowType='test-workflow-type'",
		Domain:   "test-domain",
		PageSize: 1000,
	}).Return(&types.ListWorkflowExecutionsResponse{}, nil).Times(1)

	resp, err := getAllWorkflowIDsByQuery(c, "WorkflowType='test-workflow-type'")
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	serverFrontendClient.EXPECT().ScanWorkflowExecutions(gomock.Any(), &types.ListWorkflowExecutionsRequest{
		Query:    "WorkflowType='test-workflow-type'",
		Domain:   "test-domain",
		PageSize: 1000,
	}).Return(&types.ListWorkflowExecutionsResponse{}, errors.New("test-error")).Times(1)

	_, err = getAllWorkflowIDsByQuery(c, "WorkflowType='test-workflow-type'")
	assert.Error(t, err)
}

func Test_GetWorkflowStatus(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      types.WorkflowExecutionCloseStatus
		expectedError bool
	}{
		{
			name:          "Valid status - completed",
			input:         "completed",
			expected:      types.WorkflowExecutionCloseStatusCompleted,
			expectedError: false,
		},
		{
			name:          "Valid alias - fail",
			input:         "fail",
			expected:      types.WorkflowExecutionCloseStatusFailed,
			expectedError: false,
		},
		{
			name:          "Valid status - timed_out",
			input:         "timed_out",
			expected:      types.WorkflowExecutionCloseStatusTimedOut,
			expectedError: false,
		},
		{
			name:          "Invalid status",
			input:         "invalid",
			expected:      -1,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := getWorkflowStatus(tt.input)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, -1, int(status))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, status)
			}
		})
	}
}

func Test_ConvertDescribeWorkflowExecutionResponse(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	mockResp := &types.DescribeWorkflowExecutionResponse{
		WorkflowExecutionInfo: &types.WorkflowExecutionInfo{
			Execution: &types.WorkflowExecution{
				WorkflowID: "test-workflow-id",
				RunID:      "test-run-id",
			},
		},
		PendingActivities: []*types.PendingActivityInfo{
			{
				ActivityID: "test-activity-id",
				ActivityType: &types.ActivityType{
					Name: "test-activity-type",
				},
				HeartbeatDetails:   []byte("test-heartbeat-details"),
				LastFailureDetails: []byte("test-failure-details"),
			},
		},
		PendingDecision: &types.PendingDecisionInfo{
			State: nil,
		},
	}

	resp, err := convertDescribeWorkflowExecutionResponse(mockResp, serverFrontendClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, "test-workflow-id", resp.WorkflowExecutionInfo.Execution.WorkflowID)
}

func Test_PrintRunStatus(t *testing.T) {
	// this method only prints results, no need to test the output
	tests := []struct {
		name  string
		event *types.HistoryEvent
	}{
		{
			name: "COMPLETED",
			event: &types.HistoryEvent{
				EventType: types.EventTypeWorkflowExecutionCompleted.Ptr(),
				WorkflowExecutionCompletedEventAttributes: &types.WorkflowExecutionCompletedEventAttributes{
					Result: []byte("workflow completed successfully"),
				},
			},
		},
		{
			name: "FAILED",
			event: &types.HistoryEvent{
				EventType: types.EventTypeWorkflowExecutionFailed.Ptr(),
				WorkflowExecutionFailedEventAttributes: &types.WorkflowExecutionFailedEventAttributes{
					Reason:  nil,
					Details: []byte("failure details"),
				},
			},
		},
		{
			name: "TIMEOUT",
			event: &types.HistoryEvent{
				EventType: types.EventTypeWorkflowExecutionTimedOut.Ptr(),
				WorkflowExecutionTimedOutEventAttributes: &types.WorkflowExecutionTimedOutEventAttributes{
					TimeoutType: types.TimeoutTypeStartToClose.Ptr(),
				},
			},
		},
		{
			name: "CANCELED",
			event: &types.HistoryEvent{
				EventType: types.EventTypeWorkflowExecutionCanceled.Ptr(),
				WorkflowExecutionCanceledEventAttributes: &types.WorkflowExecutionCanceledEventAttributes{
					Details: []byte("canceled details"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				printRunStatus(tt.event)
			})
		})
	}
}

func Test_ListWorkflowExecutions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)
	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})
	c := getMockContext(t, nil, app)
	listFn := listWorkflowExecutions(serverFrontendClient, 100, "test-domain", "WorkflowType='test-workflow-type'", c)
	assert.NotNil(t, listFn)
	expectedResp := &types.ListWorkflowExecutionsResponse{
		Executions: []*types.WorkflowExecutionInfo{
			{
				Execution: &types.WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				},
			},
		},
		NextPageToken: []byte("test-next-page-token"),
	}
	serverFrontendClient.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any()).Return(expectedResp, nil).Times(1)
	executions, nextPageToken, err := listFn(nil)
	assert.NoError(t, err)
	assert.NotNil(t, executions)
	assert.Equal(t, expectedResp.Executions, executions)
	assert.NotNil(t, nextPageToken)

	serverFrontendClient.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any()).Return(nil, errors.New("test-error")).Times(1)
	_, _, err = listFn(nil)
	assert.Error(t, err)
}

func Test_PrintListResults(t *testing.T) {
	executions := []*types.WorkflowExecutionInfo{
		{
			Execution: &types.WorkflowExecution{
				WorkflowID: "test-workflow-id-1",
				RunID:      "test-run-id-1",
			},
			Type: &types.WorkflowType{
				Name: "test-workflow-type-1",
			},
		},
		{
			Execution: &types.WorkflowExecution{
				WorkflowID: "test-workflow-id-2",
				RunID:      "test-run-id-2",
			},
			Type: &types.WorkflowType{
				Name: "test-workflow-type-2",
			},
		},
	}

	assert.NotPanics(t, func() {
		printListResults(executions, true, false)
		printListResults(executions, false, false)
		printListResults(executions, true, true)
		printListResults(nil, true, false)
	})
}

func Test_ResetWorkflow(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)
	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})

	set := flag.NewFlagSet("test", 0)
	c := cli.NewContext(app, set, nil)
	// missing domain flag
	err := ResetWorkflow(c)
	assert.Error(t, err)

	set.String(FlagDomain, "test-domain", "domain")
	// missing workflowID flag
	err = ResetWorkflow(c)
	assert.Error(t, err)

	set.String("workflow_id", "test-workflow-id", "workflow_id")
	set.Parse([]string{"test-workflow-id", "test-run-id"})
	// missing reason flag
	err = ResetWorkflow(c)
	assert.Error(t, err)

	set.String("reason", "test", "reason")
	set.String("decision_offset", "-1", "decision_offset")
	// invalid event ID
	err = ResetWorkflow(c)
	assert.Error(t, err)

	set.String("reset_type", "LastDecisionCompleted", "reset_type")
	set.String("run_id", "test-run-id", "run_id")
	serverFrontendClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any()).Return(&types.GetWorkflowExecutionHistoryResponse{
		History: &types.History{
			Events: []*types.HistoryEvent{
				{
					ID:        1,
					EventType: types.EventTypeDecisionTaskCompleted.Ptr(),
				},
				{
					ID:        2,
					EventType: types.EventTypeDecisionTaskScheduled.Ptr(),
				},
			},
		},
	}, nil).Times(2)
	serverFrontendClient.EXPECT().ResetWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
	serverFrontendClient.EXPECT().DescribeWorkflowExecution(gomock.Any(), gomock.Any()).Return(&types.DescribeWorkflowExecutionResponse{
		WorkflowExecutionInfo: &types.WorkflowExecutionInfo{},
	}, nil).AnyTimes()
	err = ResetWorkflow(c)
	assert.NoError(t, err)

	// reset failed
	serverFrontendClient.EXPECT().ResetWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, errors.New("test-error")).Times(1)
	err = ResetWorkflow(c)
	assert.Error(t, err)

	// getResetEventIDByType failed
	serverFrontendClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any()).Return(nil, errors.New("test-error")).Times(1)
	err = ResetWorkflow(c)
	assert.Error(t, err)
}

func Test_ResetWorkflow_Invalid_Decision_Offset(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)
	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})

	set := flag.NewFlagSet("test", 0)
	set.String(FlagDomain, "test-domain", "domain")
	set.String("workflow_id", "test-workflow-id", "workflow_id")
	set.Parse([]string{"test-workflow-id", "test-run-id"})
	set.String("reason", "test", "reason")
	set.String("decision_offset", "100", "decision_offset")
	c := cli.NewContext(app, set, nil)
	err := ResetWorkflow(c)
	assert.Error(t, err)
}

func Test_ResetWorkflow_Missing_RunID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)
	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})

	set := flag.NewFlagSet("test", 0)
	set.String(FlagDomain, "test-domain", "domain")
	set.String("workflow_id", "test-workflow-id", "workflow_id")
	set.String("reason", "test", "reason")
	set.String("decision_offset", "-1", "decision_offset")
	set.String("reset_type", "BadBinary", "reset_type")
	set.String("reset_bad_binary_checksum", "test-bad-binary-checksum", "reset_bad_binary_checksum")
	serverFrontendClient.EXPECT().DescribeWorkflowExecution(gomock.Any(), gomock.Any()).Return(&types.DescribeWorkflowExecutionResponse{
		WorkflowExecutionInfo: &types.WorkflowExecutionInfo{},
	}, errors.New("test-error")).AnyTimes()
	c := cli.NewContext(app, set, nil)
	err := ResetWorkflow(c)
	assert.Error(t, err)
}

func (s *cliAppSuite) TestCompleteActivity() {
	s.testcaseHelper([]testcase{
		{
			name:    "happy",
			command: `cadence --do test-domain wf activity complete -w wid -r rid -aid 3 -result result --identity tester`,
			err:     "",
			mock: func() {
				s.serverFrontendClient.EXPECT().RespondActivityTaskCompletedByID(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
	})
}

func (s *cliAppSuite) TestDescribeWorkflow() {
	s.testcaseHelper([]testcase{
		{
			"happy",
			"cadence --do test-domain wf describe -w wid",
			"",
			func() {
				resp := &types.DescribeWorkflowExecutionResponse{
					WorkflowExecutionInfo: &types.WorkflowExecutionInfo{
						Execution: &types.WorkflowExecution{
							WorkflowID: "wid",
						},
						Type: &types.WorkflowType{
							Name: "workflow-type",
						},
						StartTime: common.Int64Ptr(time.Now().UnixNano()),
						CloseTime: common.Int64Ptr(time.Now().UnixNano()),
					},
					PendingActivities: []*types.PendingActivityInfo{
						{},
					},
					PendingDecision: &types.PendingDecisionInfo{},
				}
				s.serverFrontendClient.EXPECT().DescribeWorkflowExecution(gomock.Any(), gomock.Any()).Return(resp, nil)
			},
		},
	})
}

func (s *cliAppSuite) TestFailActivity() {
	s.testcaseHelper([]testcase{
		{
			name:    "happy",
			command: `cadence --do test-domain wf activity fail -w wid -r rid -aid 3 --reason somereason --detail somedetail --identity tester`,
			err:     "",
			mock: func() {
				s.serverFrontendClient.EXPECT().RespondActivityTaskFailedByID(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
	})
}

func (s *cliAppSuite) TestListAllWorkflow() {
	s.testcaseHelper([]testcase{
		{
			name:    "happy",
			command: `cadence --do test-domain wf listall`,
			mock: func() {
				countWorkflowResp := &types.CountWorkflowExecutionsResponse{}
				s.serverFrontendClient.EXPECT().CountWorkflowExecutions(gomock.Any(), gomock.Any()).Return(countWorkflowResp, nil)
				s.serverFrontendClient.EXPECT().ListClosedWorkflowExecutions(gomock.Any(), gomock.Any()).Return(listClosedWorkflowExecutionsResponse, nil)
			},
		},
	})
}

func (s *cliAppSuite) TestQueryWorkflow() {
	s.testcaseHelper([]testcase{
		{
			name:    "happy",
			command: `cadence --do test-domain wf query -w wid -qt query-type-test`,
			err:     "",
			mock: func() {
				resp := &types.QueryWorkflowResponse{
					QueryResult: []byte("query-result"),
				}
				s.serverFrontendClient.EXPECT().QueryWorkflow(gomock.Any(), gomock.Any()).Return(resp, nil)
			},
		},
		{
			name:    "query with reject not_open",
			command: `cadence --do test-domain wf query -w wid -qt query-type-test --query_reject_condition not_open`,
			err:     "",
			mock: func() {
				resp := &types.QueryWorkflowResponse{
					QueryResult: []byte("query-result"),
				}
				s.serverFrontendClient.EXPECT().
					QueryWorkflow(gomock.Any(), &types.QueryWorkflowRequest{
						Domain: "test-domain",
						Execution: &types.WorkflowExecution{
							WorkflowID: "wid",
						},
						Query: &types.WorkflowQuery{
							QueryType: "query-type-test",
						},
						QueryRejectCondition: types.QueryRejectConditionNotOpen.Ptr(),
					}).
					Return(resp, nil)
			},
		},
		{
			name:    "query with reject not_completed_cleanly",
			command: `cadence --do test-domain wf query -w wid -qt query-type-test --query_reject_condition not_completed_cleanly`,
			err:     "",
			mock: func() {
				resp := &types.QueryWorkflowResponse{
					QueryResult: []byte("query-result"),
				}
				s.serverFrontendClient.EXPECT().
					QueryWorkflow(gomock.Any(), &types.QueryWorkflowRequest{
						Domain: "test-domain",
						Execution: &types.WorkflowExecution{
							WorkflowID: "wid",
						},
						Query: &types.WorkflowQuery{
							QueryType: "query-type-test",
						},
						QueryRejectCondition: types.QueryRejectConditionNotCompletedCleanly.Ptr(),
					}).
					Return(resp, nil)
			},
		},
		{
			name:    "query with unknown reject",
			command: `cadence --do test-domain wf query -w wid -qt query-type-test --query_reject_condition unknown`,
			err:     "invalid reject condition",
		},
		{
			name:    "query with eventual consistency",
			command: `cadence --do test-domain wf query -w wid -qt query-type-test --query_consistency_level eventual`,
			err:     "",
			mock: func() {
				resp := &types.QueryWorkflowResponse{
					QueryResult: []byte("query-result"),
				}
				s.serverFrontendClient.EXPECT().
					QueryWorkflow(gomock.Any(), &types.QueryWorkflowRequest{
						Domain: "test-domain",
						Execution: &types.WorkflowExecution{
							WorkflowID: "wid",
						},
						Query: &types.WorkflowQuery{
							QueryType: "query-type-test",
						},
						QueryConsistencyLevel: types.QueryConsistencyLevelEventual.Ptr(),
					}).
					Return(resp, nil)
			},
		},
		{
			name:    "query with strong consistency",
			command: `cadence --do test-domain wf query -w wid -qt query-type-test --query_consistency_level strong`,
			err:     "",
			mock: func() {
				resp := &types.QueryWorkflowResponse{
					QueryResult: []byte("query-result"),
				}
				s.serverFrontendClient.EXPECT().
					QueryWorkflow(gomock.Any(), &types.QueryWorkflowRequest{
						Domain: "test-domain",
						Execution: &types.WorkflowExecution{
							WorkflowID: "wid",
						},
						Query: &types.WorkflowQuery{
							QueryType: "query-type-test",
						},
						QueryConsistencyLevel: types.QueryConsistencyLevelStrong.Ptr(),
					}).
					Return(resp, nil)
			},
		},
		{
			name:    "query with invalid consistency",
			command: `cadence --do test-domain wf query -w wid -qt query-type-test --query_consistency_level invalid`,
			err:     "invalid query consistency level",
		},
		{
			name:    "failed",
			command: `cadence --do test-domain wf query -w wid -qt query-type-test`,
			err:     "some error",
			mock: func() {
				s.serverFrontendClient.EXPECT().QueryWorkflow(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some error"))
			},
		},
		{
			name:    "missing flags",
			command: "cadence wf query",
			err:     "Required flag not found",
		},
	})
}

func (s *cliAppSuite) TestResetWorkflow() {
	s.testcaseHelper([]testcase{
		{
			name:    "happy",
			command: `cadence --do test-domain wf reset -w wid -r rid -reason test-reason --event_id 1`,
			err:     "",
			mock: func() {
				resp := &types.ResetWorkflowExecutionResponse{RunID: uuid.New()}
				s.serverFrontendClient.EXPECT().ResetWorkflowExecution(gomock.Any(), gomock.Any()).Return(resp, nil)
			},
		},
	})
}

func (s *cliAppSuite) TestScanAllWorkflow() {
	s.testcaseHelper([]testcase{
		{
			name:    "happy",
			command: `cadence --do test-domain wf scanall`,
			mock: func() {
				s.serverFrontendClient.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&types.ListWorkflowExecutionsResponse{}, nil)
			},
		},
	})
}

func (s *cliAppSuite) TestSignalWithStartWorkflowExecution() {
	s.testcaseHelper([]testcase{
		{
			name:    "happy",
			command: `cadence --do test-domain wf signalwithstart --et 100 --workflow_type sometype --tasklist tasklist -w wid -n signal-name --signal_input []`,
			err:     "",
			mock: func() {
				s.serverFrontendClient.EXPECT().SignalWithStartWorkflowExecution(gomock.Any(), gomock.Any()).Return(&types.StartWorkflowExecutionResponse{}, nil)
			},
		},
		{
			name:    "failed",
			command: `cadence --do test-domain wf signalwithstart --et 100 --workflow_type sometype --tasklist tasklist -w wid -n signal-name --signal_input []`,
			err:     "some error",
			mock: func() {
				s.serverFrontendClient.EXPECT().SignalWithStartWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some error"))
			},
		},
		{
			name:    "missing flags",
			command: "cadence wf signalwithstart",
			err:     "Required flag not found",
		},
	})
}

func Test_DoReset(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)
	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})

	set := flag.NewFlagSet("test", 0)

	set.String(FlagDomain, "test-domain", "domain")
	set.String("workflow_id", "test-workflow-id", "workflow_id")
	set.Parse([]string{"test-workflow-id", "test-run-id"})
	set.String("reason", "test", "reason")
	set.String("decision_offset", "-1", "decision_offset")
	set.String("reset_type", "LastDecisionCompleted", "reset_type")
	serverFrontendClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any()).Return(nil, errors.New("test-error")).Times(1)
	serverFrontendClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any()).Return(&types.GetWorkflowExecutionHistoryResponse{
		History: &types.History{
			Events: []*types.HistoryEvent{
				{
					ID:        1,
					EventType: types.EventTypeDecisionTaskCompleted.Ptr(),
				},
				{
					ID:        2,
					EventType: types.EventTypeDecisionTaskScheduled.Ptr(),
				},
			},
		},
	}, nil).AnyTimes()
	serverFrontendClient.EXPECT().ResetWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, errors.New("test-error")).Times(1)
	serverFrontendClient.EXPECT().ResetWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	serverFrontendClient.EXPECT().DescribeWorkflowExecution(gomock.Any(), gomock.Any()).Return(&types.DescribeWorkflowExecutionResponse{
		WorkflowExecutionInfo: &types.WorkflowExecutionInfo{},
	}, errors.New("test-error")).Times(1)
	serverFrontendClient.EXPECT().DescribeWorkflowExecution(gomock.Any(), gomock.Any()).Return(&types.DescribeWorkflowExecutionResponse{
		WorkflowExecutionInfo: &types.WorkflowExecutionInfo{
			Execution: &types.WorkflowExecution{
				WorkflowID: "test-workflow-id",
				RunID:      "test-run-id",
			},
		},
	}, nil).AnyTimes()

	c := cli.NewContext(app, set, nil)
	params := &batchResetParamsType{
		reason:    "test",
		resetType: "LastDecisionCompleted",
	}
	// describe workflow execution failed
	err := doReset(c, "test-domain", "test-workflow-id", "test-run-id", *params)
	assert.Error(t, err)

	// get reset event id failure
	err = doReset(c, "test-domain", "test-workflow-id", "test-run-id", *params)
	assert.Error(t, err)

	// reset failure
	err = doReset(c, "test-domain", "test-workflow-id", "", *params)
	assert.Error(t, err)

	// normal case
	err = doReset(c, "test-domain", "test-workflow-id", "", *params)
	assert.NoError(t, err)

	// dry run
	params.dryRun = true
	err = doReset(c, "test-domain", "test-workflow-id", "", *params)
	assert.NoError(t, err)

	// skip current open
	params.skipCurrentOpen = true
	err = doReset(c, "test-domain", "test-workflow-id", "", *params)
	assert.NoError(t, err)

	// current run id not match with input rid
	params.skipBaseNotCurrent = true
	err = doReset(c, "test-domain", "test-workflow-id", "test-not-matched-rid", *params)
	assert.NoError(t, err)

	// dry run
	params.dryRun = true
	err = doReset(c, "test-domain", "test-workflow-id", "", *params)
	assert.NoError(t, err)
}

func Test_DoReset_SkipCurrentCompleted(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	serverFrontendClient := frontend.NewMockClient(mockCtrl)
	serverAdminClient := admin.NewMockClient(mockCtrl)
	app := NewCliApp(&clientFactoryMock{
		serverFrontendClient: serverFrontendClient,
		serverAdminClient:    serverAdminClient,
	})

	set := flag.NewFlagSet("test", 0)

	set.String(FlagDomain, "test-domain", "domain")
	set.String("workflow_id", "test-workflow-id", "workflow_id")
	set.Parse([]string{"test-workflow-id", "test-run-id"})
	set.String("reason", "test", "reason")
	set.String("decision_offset", "-1", "decision_offset")
	set.String("reset_type", "LastDecisionCompleted", "reset_type")
	serverFrontendClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any()).Return(nil, errors.New("test-error")).Times(1)
	serverFrontendClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any()).Return(&types.GetWorkflowExecutionHistoryResponse{
		History: &types.History{
			Events: []*types.HistoryEvent{
				{
					ID:        1,
					EventType: types.EventTypeDecisionTaskCompleted.Ptr(),
				},
			},
		},
	}, nil).AnyTimes()
	serverFrontendClient.EXPECT().ResetWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	serverFrontendClient.EXPECT().DescribeWorkflowExecution(gomock.Any(), gomock.Any()).Return(&types.DescribeWorkflowExecutionResponse{
		WorkflowExecutionInfo: &types.WorkflowExecutionInfo{
			Execution: &types.WorkflowExecution{
				WorkflowID: "test-workflow-id",
				RunID:      "test-run-id",
			},
			CloseStatus: types.WorkflowExecutionCloseStatusCompleted.Ptr(),
			CloseTime:   common.Int64Ptr(time.Now().UnixNano()),
		},
	}, nil).AnyTimes()

	c := cli.NewContext(app, set, nil)
	params := &batchResetParamsType{
		reason:               "test",
		resetType:            "LastDecisionCompleted",
		nonDeterministicOnly: true,
	}
	// check non determinism failed when get history
	err := doReset(c, "test-domain", "test-workflow-id", "test-run-id", *params)
	assert.Error(t, err)

	err = doReset(c, "test-domain", "test-workflow-id", "test-run-id", *params)
	assert.NoError(t, err)

	// describe workflow execution failed
	params.skipCurrentCompleted = true
	err = doReset(c, "test-domain", "test-workflow-id", "test-run-id", *params)
	assert.NoError(t, err)
}

func createTempFileWithContent(t *testing.T, content string) (string, func()) {
	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}

	_, err = tmpFile.Write([]byte(content))
	if err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}

	tmpFileName := tmpFile.Name()
	tmpFile.Close()

	// Return a cleanup function to delete the file after the test
	cleanup := func() {
		os.Remove(tmpFileName)
	}

	return tmpFileName, cleanup
}

func TestLoadWorkflowIDsFromFile_Success(t *testing.T) {
	content := "wid1,wid2,wid3\n\nwid4,wid5\nwid6\n"
	fileName, cleanup := createTempFileWithContent(t, content)
	defer cleanup()

	workflowIDs, err := loadWorkflowIDsFromFile(fileName, ",")
	assert.NoError(t, err)

	expected := map[string]bool{
		"wid1": true,
		"wid4": true,
		"wid6": true,
	}
	assert.Equal(t, expected, workflowIDs)
}

func TestLoadWorkflowIDsFromFile_Failure(t *testing.T) {
	// open failed
	_, err := loadWorkflowIDsFromFile("non exist file", ",")
	assert.Error(t, err)
}
