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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/client/frontend"
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
