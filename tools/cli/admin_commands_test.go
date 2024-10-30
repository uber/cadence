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
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/tools/cli/clitest"
)

const (
	testShardID      = 1234
	testCluster      = "test-cluster"
	testDomain       = "test-domain"
	testTaskList     = "test-tasklist"
	testTaskListType = "decision"
	testQueueType    = 2 // transfer queue
	testWorkflowID   = "test-workflow-id"
	testRunID        = "test-run-id"
)

type cliTestData struct {
	mockFrontendClient *frontend.MockClient
	mockAdminClient    *admin.MockClient
	ioHandler          *testIOHandler
	app                *cli.App
	mockManagerFactory *MockManagerFactory
}

func newCLITestData(t *testing.T) *cliTestData {
	var td cliTestData

	ctrl := gomock.NewController(t)

	td.mockFrontendClient = frontend.NewMockClient(ctrl)
	td.mockAdminClient = admin.NewMockClient(ctrl)
	td.mockManagerFactory = NewMockManagerFactory(ctrl)
	td.ioHandler = &testIOHandler{}

	// Create a new CLI app with client factory and persistence manager factory
	td.app = NewCliApp(
		&clientFactoryMock{
			serverFrontendClient: td.mockFrontendClient,
			serverAdminClient:    td.mockAdminClient,
		},
		WithIOHandler(td.ioHandler),
		WithManagerFactory(td.mockManagerFactory), // Inject the mocked persistence manager factory
	)
	return &td
}

func (td *cliTestData) consoleOutput() string {
	return td.ioHandler.outputBytes.String()
}

func TestAdminResetQueue(t *testing.T) {
	tests := []struct {
		name           string
		testSetup      func(td *cliTestData) *cli.Context
		errContains    string // empty if no error is expected
		expectedOutput string
	}{
		{
			name: "no shardID argument",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app /* arguments are missing */)
			},
			errContains: "Required flag not found",
		},
		{
			name: "missing cluster argument",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(
					t,
					td.app,
					clitest.IntArgument(FlagShardID, testShardID),
					/* no cluster argument */
				)
			},
			errContains: "Required flag not found",
		},
		{
			name: "missing queue type argument",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(
					t,
					td.app,
					clitest.IntArgument(FlagShardID, testShardID),
					clitest.StringArgument(FlagCluster, testCluster),
					/* no queue type argument */
				)
			},
			errContains: "Required flag not found",
		},
		{
			name: "all arguments provided",
			testSetup: func(td *cliTestData) *cli.Context {
				cliCtx := clitest.NewCLIContext(
					t,
					td.app,
					clitest.IntArgument(FlagShardID, testShardID),
					clitest.StringArgument(FlagCluster, testCluster),
					clitest.IntArgument(FlagQueueType, testQueueType),
				)

				td.mockAdminClient.EXPECT().ResetQueue(gomock.Any(), &types.ResetQueueRequest{
					ShardID:     testShardID,
					ClusterName: testCluster,
					Type:        common.Int32Ptr(testQueueType),
				})

				return cliCtx
			},
			errContains:    "",
			expectedOutput: "Reset queue state succeeded\n",
		},
		{
			name: "ResetQueue returns an error",
			testSetup: func(td *cliTestData) *cli.Context {
				cliCtx := clitest.NewCLIContext(
					t,
					td.app,
					clitest.IntArgument(FlagShardID, testShardID),
					clitest.StringArgument(FlagCluster, testCluster),
					clitest.IntArgument(FlagQueueType, testQueueType),
				)

				td.mockAdminClient.EXPECT().ResetQueue(gomock.Any(), gomock.Any()).
					Return(errors.New("critical error"))

				return cliCtx
			},
			errContains: "Failed to reset queue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := newCLITestData(t)
			cliCtx := tt.testSetup(td)

			err := AdminResetQueue(cliCtx)
			if tt.errContains == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errContains)
			}
			assert.Equal(t, tt.expectedOutput, td.consoleOutput())
		})
	}
}

func TestAdminDescribeQueue(t *testing.T) {
	tests := []struct {
		name           string
		testSetup      func(td *cliTestData) *cli.Context
		errContains    string // empty if no error is expected
		expectedOutput string
	}{
		{
			name: "no shardID argument",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app /* arguments are missing */)
			},
			errContains: "Required flag not found",
		},
		{
			name: "missing cluster argument",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(
					t,
					td.app,
					clitest.IntArgument(FlagShardID, testShardID),
					/* no cluster argument */
				)
			},
			errContains: "Required flag not found",
		},
		{
			name: "missing queue type argument",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(
					t,
					td.app,
					clitest.IntArgument(FlagShardID, testShardID),
					clitest.StringArgument(FlagCluster, testCluster),
					/* no queue type argument */
				)
			},
			errContains: "Required flag not found",
		},
		{
			name: "all arguments provided",
			testSetup: func(td *cliTestData) *cli.Context {
				cliCtx := clitest.NewCLIContext(
					t,
					td.app,
					clitest.IntArgument(FlagShardID, testShardID),
					clitest.StringArgument(FlagCluster, testCluster),
					clitest.IntArgument(FlagQueueType, testQueueType),
				)

				td.mockAdminClient.EXPECT().DescribeQueue(gomock.Any(), &types.DescribeQueueRequest{
					ShardID:     testShardID,
					ClusterName: testCluster,
					Type:        common.Int32Ptr(testQueueType),
				}).Return(&types.DescribeQueueResponse{
					ProcessingQueueStates: []string{"state1", "state2"},
				}, nil)

				return cliCtx
			},
			errContains:    "",
			expectedOutput: "state1\nstate2\n",
		},
		{
			name: "DescribeQueue returns an error",
			testSetup: func(td *cliTestData) *cli.Context {
				cliCtx := clitest.NewCLIContext(
					t,
					td.app,
					clitest.IntArgument(FlagShardID, testShardID),
					clitest.StringArgument(FlagCluster, testCluster),
					clitest.IntArgument(FlagQueueType, testQueueType),
				)

				td.mockAdminClient.EXPECT().DescribeQueue(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("critical error"))

				return cliCtx
			},
			errContains: "Failed to describe queue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := newCLITestData(t)
			cliCtx := tt.testSetup(td)

			err := AdminDescribeQueue(cliCtx)
			if tt.errContains == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errContains)
			}
			assert.Equal(t, tt.expectedOutput, td.consoleOutput())
		})
	}
}

func TestAdminRefreshWorkflowTasks(t *testing.T) {
	tests := []struct {
		name           string
		testSetup      func(td *cliTestData) *cli.Context
		errContains    string // empty if no error is expected
		expectedOutput string
	}{
		{
			name: "no domain argument",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app /* arguments are missing */)
			},
			errContains: "Required flag not found",
		},
		{
			name: "missing workflowID argument",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(
					t,
					td.app,
					clitest.StringArgument(FlagDomain, testDomain),
					/* no workflowID argument */
				)
			},
			errContains: "Required flag not found",
		},
		{
			name: "all arguments provided",
			testSetup: func(td *cliTestData) *cli.Context {
				cliCtx := clitest.NewCLIContext(
					t,
					td.app,
					clitest.StringArgument(FlagDomain, testDomain),
					clitest.StringArgument(FlagWorkflowID, testWorkflowID),
					clitest.StringArgument(FlagRunID, testRunID),
				)

				td.mockAdminClient.EXPECT().RefreshWorkflowTasks(gomock.Any(), &types.RefreshWorkflowTasksRequest{
					Domain: testDomain,
					Execution: &types.WorkflowExecution{
						WorkflowID: testWorkflowID,
						RunID:      testRunID,
					},
				}).Return(nil)

				return cliCtx
			},
			errContains:    "",
			expectedOutput: "Refresh workflow task succeeded.\n",
		},
		{
			name: "RefreshWorkflowTasks returns an error",
			testSetup: func(td *cliTestData) *cli.Context {
				cliCtx := clitest.NewCLIContext(
					t,
					td.app,
					clitest.StringArgument(FlagDomain, testDomain),
					clitest.StringArgument(FlagWorkflowID, testWorkflowID),
					clitest.StringArgument(FlagRunID, testRunID),
				)

				td.mockAdminClient.EXPECT().RefreshWorkflowTasks(gomock.Any(), gomock.Any()).
					Return(errors.New("critical error"))

				return cliCtx
			},
			errContains: "Refresh workflow task failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := newCLITestData(t)
			cliCtx := tt.testSetup(td)

			err := AdminRefreshWorkflowTasks(cliCtx)
			if tt.errContains == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errContains)
			}
			assert.Equal(t, tt.expectedOutput, td.consoleOutput())
		})
	}
}

func TestAdminDescribeHistoryHost(t *testing.T) {
	tests := []struct {
		name           string
		testSetup      func(td *cliTestData) *cli.Context
		errContains    string // empty if no error is expected
		expectedOutput string
	}{
		{
			name: "no arguments provided",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app /* arguments are missing */)
			},
			errContains: "at least one of them is required to provide to lookup host: workflowID, shardID and host address",
		},
		{
			name: "workflow-id provided, but empty",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app,
					clitest.StringArgument(FlagWorkflowID, ""))
			},
			errContains: "at least one of them is required to provide to lookup host: workflowID, shardID and host address",
		},
		{
			name: "addr provided, but empty",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app,
					clitest.StringArgument(FlagHistoryAddress, ""))
			},
			errContains: "at least one of them is required to provide to lookup host: workflowID, shardID and host address",
		},
		{
			name: "calling with all arguments",
			testSetup: func(td *cliTestData) *cli.Context {
				cliCtx := clitest.NewCLIContext(
					t,
					td.app,
					clitest.IntArgument(FlagShardID, testShardID),
					clitest.StringArgument(FlagWorkflowID, testWorkflowID),
					clitest.StringArgument(FlagHistoryAddress, "host:port"),
					clitest.BoolArgument(FlagPrintFullyDetail, false),
				)

				td.mockAdminClient.EXPECT().DescribeHistoryHost(gomock.Any(),
					&types.DescribeHistoryHostRequest{
						ExecutionForHost: &types.WorkflowExecution{WorkflowID: testWorkflowID},
						ShardIDForHost:   common.Int32Ptr(int32(testShardID)),
						HostAddress:      common.StringPtr("host:port"),
					}).Return(
					&types.DescribeHistoryHostResponse{NumberOfShards: 12, Address: "host:port"},
					nil,
				)

				return cliCtx
			},
			errContains: "",
			expectedOutput: `{
  "numberOfShards": 12,
  "address": "host:port"
}
`,
		},
		{
			name: "DescribeHistoryHost returns an error",
			testSetup: func(td *cliTestData) *cli.Context {
				cliCtx := clitest.NewCLIContext(
					t,
					td.app,
					clitest.StringArgument(FlagWorkflowID, testWorkflowID),
				)

				td.mockAdminClient.EXPECT().DescribeHistoryHost(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("critical error"))

				return cliCtx
			},
			errContains: "Describe history host failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := newCLITestData(t)
			cliCtx := tt.testSetup(td)

			err := AdminDescribeHistoryHost(cliCtx)
			if tt.errContains == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errContains)
			}
			assert.Equal(t, tt.expectedOutput, td.consoleOutput())
		})
	}
}

func TestAdminCloseShard(t *testing.T) {
	tests := []struct {
		name        string
		testSetup   func(td *cliTestData) *cli.Context
		errContains string // empty if no error is expected
	}{
		{
			name: "no arguments provided",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app /* arguments are missing */)
			},
			errContains: "Required option not found",
		},
		{
			name: "calling with all arguments",
			testSetup: func(td *cliTestData) *cli.Context {
				cliCtx := clitest.NewCLIContext(
					t,
					td.app,
					clitest.IntArgument(FlagShardID, testShardID),
				)

				td.mockAdminClient.EXPECT().CloseShard(gomock.Any(),
					&types.CloseShardRequest{
						ShardID: testShardID,
					}).Return(nil)

				return cliCtx
			},
			errContains: "",
		},
		{
			name: "CloseShard returns an error",
			testSetup: func(td *cliTestData) *cli.Context {
				cliCtx := clitest.NewCLIContext(
					t,
					td.app,
					clitest.IntArgument(FlagShardID, testShardID),
				)

				td.mockAdminClient.EXPECT().CloseShard(gomock.Any(), gomock.Any()).
					Return(errors.New("critical error"))

				return cliCtx
			},
			errContains: "Close shard task has failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := newCLITestData(t)
			cliCtx := tt.testSetup(td)

			err := AdminCloseShard(cliCtx)
			if tt.errContains == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errContains)
			}
		})
	}
}
