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

package cli

import (
	"fmt"
	"errors"
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"testing"

	// "time"
	"os"

	"github.com/uber/cadence/common"
	// "github.com/opentracing/opentracing-go"
	// "github.com/opentracing/opentracing-go/mocktracer"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/types"
	"github.com/urfave/cli/v2"

	"github.com/uber/cadence/service/worker/batcher"
	"github.com/uber/cadence/client/frontend"
)



func TestStartBatchJob(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*frontend.MockClient)
		flags          map[string]interface{}
		expectedError  string
		expectedOutput map[string]interface{}
	}{
		{
			name: "Valid Start Batch Job",
			setup: func(mockClient *frontend.MockClient) {
				mockClient.EXPECT().CountWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&types.CountWorkflowExecutionsResponse{
					Count: 100,
				}, nil)
				mockClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any()).Return(&types.StartWorkflowExecutionResponse{
					RunID: "run-id-example", 
				},nil)
			},
			flags: map[string]interface{}{
				FlagDomain:     "test-domain",
				FlagListQuery:  "workflowType='batch'",
				FlagReason:     "Testing batch job",
				FlagBatchType:  batcher.BatchTypeSignal,
				FlagSignalName: "test-signal",
				FlagInput:      "test-input",
				FlagYes:        true, // Automatically confirm
			},
			expectedError:  "",
		},
		{
			name: "Missing Domain",
			setup: func(mockClient *frontend.MockClient) {},
			flags: map[string]interface{}{
				FlagListQuery: "workflowType='batch'",
				FlagReason:    "Testing batch job",
				FlagBatchType: batcher.BatchTypeSignal,
			},
			expectedError: "Option domain is required",
		},
		{
			name: "Missing ListQuery",
			setup: func(mockClient *frontend.MockClient) {},
			flags: map[string]interface{}{
				FlagDomain: "test-domain",
				FlagReason: "Testing batch job",
				FlagBatchType: batcher.BatchTypeSignal,
			},
			expectedError: "Option query is required",
		},
		{
			name: "Missing Reason",
			setup: func(mockClient *frontend.MockClient) {},
			flags: map[string]interface{}{
				FlagDomain:    "test-domain",
				FlagListQuery: "workflowType='batch'",
				FlagBatchType: batcher.BatchTypeSignal,
			},
			expectedError: "Option reason is required",
		},
		{
			name: "Invalid Batch Type",
			setup: func(mockClient *frontend.MockClient) {},
			flags: map[string]interface{}{
				FlagDomain:     "test-domain",
				FlagListQuery:  "workflowType='batch'",
				FlagReason:     "Testing batch job",
				FlagBatchType:  "invalidBatchType",
			},
			expectedError: "batchType is not valid, supported:terminate,cancel,signal,replicate",
		},
		{
			name: "Count Workflow Executions Failure",
			setup: func(mockClient *frontend.MockClient) {
				mockClient.EXPECT().CountWorkflowExecutions(gomock.Any(), gomock.Any()).Return(nil, errors.New("count error"))
			},
			flags: map[string]interface{}{
				FlagDomain:    "test-domain",
				FlagListQuery: "workflowType='batch'",
				FlagReason:    "Testing batch job",
				FlagBatchType: batcher.BatchTypeSignal,
				FlagSignalName: "test-signal",
				FlagInput:     "test-input",
				FlagYes:       true, // Automatically confirm
			},
			expectedError: "Failed to count impacting workflows for starting a batch job: count error",
		},
		{
			name: "Start Workflow Execution Failure",
			setup: func(mockClient *frontend.MockClient) {
				mockClient.EXPECT().CountWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&types.CountWorkflowExecutionsResponse{
					Count: 100,
				}, nil)
				mockClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, errors.New("start error"))
			},
			flags: map[string]interface{}{
				FlagDomain:    "test-domain",
				FlagListQuery: "workflowType='batch'",
				FlagReason:    "Testing batch job",
				FlagBatchType: batcher.BatchTypeSignal,
				FlagSignalName: "test-signal",
				FlagInput:     "test-input",
				FlagYes:       true, // Automatically confirm
			},
			expectedError: "Failed to start batch job: start error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockClient := frontend.NewMockClient(mockCtrl)
			app := NewCliApp(&clientFactoryMock{
				serverFrontendClient: mockClient,
			})

			set := flag.NewFlagSet("test", 0)
			for k, v := range tt.flags {
				switch val := v.(type) {
				case string:
					_ = set.String(k, val, "")
				case int:
					_ = set.Int(k, val, "")
				case bool:
					_ = set.Bool(k, val, "")
				}
			}
			c := cli.NewContext(app, set, nil)
			tt.setup(mockClient)

			// Redirect stdout to capture output
			reader, writer, _ := os.Pipe()
			oldStdout := os.Stdout
			os.Stdout = writer

			err := StartBatchJob(c)
			writer.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, reader)
			reader.Close()

			fmt.Println("Captured Output:", buf.String())

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				var output map[string]interface{}
				err = json.Unmarshal(buf.Bytes(), &output)
				// fmt.Println(err)
				// assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestTerminateBatchJob(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*frontend.MockClient)
		flags          map[string]interface{}
		expectedError  string
		expectedOutput map[string]interface{}
	}{
		{
			name: "Valid Termination",
			setup: func(mockClient *frontend.MockClient) {
				mockClient.EXPECT().TerminateWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil)
			},
			flags: map[string]interface{}{
				FlagJobID:  "example-workflow-1",
				FlagReason: "Testing termination",
			},
			expectedError:  "",
			expectedOutput: map[string]interface{}{"msg": "batch job is terminated"},
		},
		{
			name: "Missing JobID",
			setup: func(mockClient *frontend.MockClient) {},
			flags: map[string]interface{}{
				FlagReason: "Testing termination",
			},
			expectedError: "Option job_id is required",
		},
		{
			name: "Missing Reason",
			setup: func(mockClient *frontend.MockClient) {},
			flags: map[string]interface{}{
				FlagJobID: "example-workflow-1",
			},
			expectedError: "Option reason is required",
		},
		{
			name: "Terminate Failure",
			setup: func(mockClient *frontend.MockClient) {
				mockClient.EXPECT().TerminateWorkflowExecution(gomock.Any(), gomock.Any()).Return(errors.New("termination error"))
			},
			flags: map[string]interface{}{
				FlagJobID:  "example-workflow-1",
				FlagReason: "Testing termination",
			},
			expectedError: "Failed to terminate batch job: termination error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockClient := frontend.NewMockClient(mockCtrl)
			app := NewCliApp(&clientFactoryMock{
				serverFrontendClient: mockClient,
			})

			set := flag.NewFlagSet("test", 0)
			for k, v := range tt.flags {
				switch val := v.(type) {
				case string:
					_ = set.String(k, val, "")
				case int:
					_ = set.Int(k, val, "")
				}
			}
			c := cli.NewContext(app, set, nil)
			tt.setup(mockClient)

			// Redirect stdout to capture output
			reader, writer, _ := os.Pipe()
			oldStdout := os.Stdout
			os.Stdout = writer

			err := TerminateBatchJob(c)
			writer.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, reader)
			reader.Close()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				var output map[string]interface{}
				err = json.Unmarshal(buf.Bytes(), &output)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestDescribeBatchJob(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*frontend.MockClient)
		flags          map[string]interface{}
		expectedError  string
		expectedOutput map[string]interface{}
	}{
				{
			name: "Valid Batch Job",
			setup: func(mockClient *frontend.MockClient) {
				mockClient.EXPECT().DescribeWorkflowExecution(gomock.Any(), gomock.Any()).Return(&types.DescribeWorkflowExecutionResponse{
					WorkflowExecutionInfo: &types.WorkflowExecutionInfo{
						CloseStatus: types.WorkflowExecutionCloseStatusCompleted.Ptr(),
						Execution: &types.WorkflowExecution{
							WorkflowID: "example-workflow-1",
						},
						StartTime:   common.Int64Ptr(1697018400),
					},
					PendingActivities: []*types.PendingActivityInfo{
						{
							HeartbeatDetails: json.RawMessage(`{"PageToken": null, "CurrentPage": 50, "TotalEstimate": 100, "SuccessCount": 0, "ErrorCount": 0}`),
						},
					},
				}, nil)
			},
			flags: map[string]interface{}{
				FlagJobID: "example-workflow-1",
			},
			expectedError: "",
			expectedOutput: map[string]interface{}{
				"msg": "batch job is finished successfully",
			},
		},
		{
			name: "Batch Job Running",
			setup: func(mockClient *frontend.MockClient) {
				mockClient.EXPECT().DescribeWorkflowExecution(gomock.Any(), gomock.Any()).Return(&types.DescribeWorkflowExecutionResponse{
					WorkflowExecutionInfo: &types.WorkflowExecutionInfo{
						CloseStatus: nil,
						Execution: &types.WorkflowExecution{
							WorkflowID: "example-workflow-2",
						},
					},
					PendingActivities: []*types.PendingActivityInfo{
						{
							HeartbeatDetails: json.RawMessage(`{"PageToken": null, "CurrentPage": 30, "TotalEstimate": 100, "SuccessCount": 10, "ErrorCount": 0}`),
						},
					},
				}, nil)
			},
			flags: map[string]interface{}{
				FlagJobID: "example-workflow-2",
			},
			expectedError: "",
			expectedOutput: map[string]interface{}{
				"msg": "batch job is running",
				"progress": batcher.HeartBeatDetails{
					PageToken:     nil,
					CurrentPage:   30,
					TotalEstimate: 100,
					SuccessCount:  10,
					ErrorCount:    0,
				},
			},
		},
		{
			name: "Error when describing job",
			setup: func(mockClient *frontend.MockClient) {
				mockClient.EXPECT().DescribeWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, errors.New("service error"))
			},
			flags: map[string]interface{}{
				FlagJobID: "error-job",
			},
			expectedError: "Failed to describe batch job: service error",
		},
		{
			name:          "Missing Job ID",
			setup:        func(mockClient *frontend.MockClient) {},
			flags:        map[string]interface{}{},
			expectedError: "Option job_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockClient := frontend.NewMockClient(mockCtrl)
			app := NewCliApp(&clientFactoryMock{
				serverFrontendClient: mockClient,
			})

			set := flag.NewFlagSet("test", 0)
			for k, v := range tt.flags {
				switch val := v.(type) {
				case string:
					_ = set.String(k, val, "")
				case int:
					_ = set.Int(k, val, "")
				}
			}
			c := cli.NewContext(app, set, nil)
			tt.setup(mockClient)

			// Redirect stdout to capture output
			reader, writer, _ := os.Pipe()
			oldStdout := os.Stdout
			os.Stdout = writer

			err := DescribeBatchJob(c)
			writer.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, reader)
			reader.Close()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				var output map[string]interface{}
				err = json.Unmarshal(buf.Bytes(), &output)
				assert.NoError(t, err)
	
				// Type assertion for progress
				if progressRaw, exists := output["progress"]; exists {
					var progress batcher.HeartBeatDetails
					progressBytes, err := json.Marshal(progressRaw)
					if err == nil {
						json.Unmarshal(progressBytes, &progress)
						output["progress"] = progress
					}
				}

				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestListBatchJobs(t *testing.T) {
	tests := []struct {
        name            string
        setup           func(*frontend.MockClient)
        flags           map[string]interface{}
        expectedError   string
        expectedOutput  []map[string]string
    }{
		{
			name: "Valid Batch Job",
			setup : func(mockClient *frontend.MockClient) {
                mockClient.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any()).Return(&types.ListWorkflowExecutionsResponse{
					Executions: []*types.WorkflowExecutionInfo {
						{
							Execution: &types.WorkflowExecution{
								WorkflowID: "example-workflow-1",
							},
							StartTime: common.Int64Ptr(1697018400),
							Memo: &types.Memo{
								Fields: map[string][]byte{
									"Reason": []byte("Testing reason"),
								},
							},
							SearchAttributes: &types.SearchAttributes{
								IndexedFields: map[string][]byte{
									"Operator": []byte("test-operator"),
								},
							},
							CloseStatus: types.WorkflowExecutionCloseStatusCompleted.Ptr(),
						},
					},
					NextPageToken:  []byte("test-next-token"),
                }, nil)
            },
			flags: map[string]interface{}{
				FlagDomain:       "test-domain",
				FlagPageSize:     100,
			},
			expectedError: "",
			expectedOutput: []map[string]string{
				{
					"jobID":     "example-workflow-1",
					"startTime": "1970-01-01T05:30:01+05:30",
					"reason":    "Testing reason",
					"operator":  "test-operator",
					"status":    "COMPLETED",
					"closeTime": "1970-01-01T05:30:00+05:30",
				},
			},
		},
		{
			name: "Missing Domain",
            setup: func(mockClient *frontend.MockClient) {},
            flags: map[string]interface{}{},
			expectedError: "Option domain is required",
		},

	}

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			serverFrontendClient := frontend.NewMockClient(mockCtrl)
			app := NewCliApp(&clientFactoryMock{
				serverFrontendClient: serverFrontendClient,
			})
			set := flag.NewFlagSet("test", 0)
            for k, v := range tt.flags {
				switch val := v.(type) {
				case string:
					_ = set.String(k, val, "")
				case int:
					_ = set.Int(k, val, "")
				}
            }
			c := cli.NewContext(app, set, nil)
            tt.setup(serverFrontendClient)
			reader, writer, _ := os.Pipe()
			oldStdout := os.Stdout
			os.Stdout = writer
			err := ListBatchJobs(c)
			writer.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, reader)
			reader.Close()
			if tt.expectedError != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
            } else {
                assert.NoError(t, err)
				var output []map[string]string
				err = json.Unmarshal(buf.Bytes(), &output)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
            }
        })
    }
}

func TestValidateBatchType(t *testing.T) {
    // Mock batch types for testing
    batcher.AllBatchTypes = []string{"signal", "replicate", "terminate"}

    tests := []struct {
        name       string
        batchType  string
        expected   bool
    }{
        {
            name:      "Valid batch type - signal",
            batchType: "signal",
            expected:  true,
        },
        {
            name:      "Valid batch type - replicate",
            batchType: "replicate",
            expected:  true,
        },
        {
            name:      "Invalid batch type",
            batchType: "invalid",
            expected:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := validateBatchType(tt.batchType)
            if result != tt.expected {
                t.Errorf("expected %v, got %v for batch type %v", tt.expected, result, tt.batchType)
            }
        })
    }
}