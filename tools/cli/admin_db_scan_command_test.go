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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/tools/cli/clitest"
	"github.com/urfave/cli/v2"
)

func TestAdminDBScan(t *testing.T) {
	td := newCLITestData(t)

	expectWorkFlow(td, "test-workflow-id1")
	expectWorkFlow(td, "test-workflow-id2")
	expectWorkFlow(td, "test-workflow-id3")

	cliCtx := clitest.NewCLIContext(t, td.app,
		clitest.StringArgument("scan_type", "CurrentExecutionType"),
		clitest.IntArgument("number_of_shards", 16384),
		clitest.StringSliceArgument("invariant_collection", "CollectionMutableState"),
		clitest.StringArgument("input_file", "testdata/scan_input.json"),
	)

	err := AdminDBScan(cliCtx)
	assert.NoError(t, err)
}

func expectWorkFlow(td *cliTestData, workflowID string) {
	shardId1 := common.WorkflowIDToHistoryShard(workflowID, 16384)
	mockExecutionManager := persistence.NewMockExecutionManager(td.ctrl)
	mockExecutionManager.EXPECT().Close().Times(1)
	td.mockManagerFactory.EXPECT().
		initializeExecutionManager(gomock.Any(), shardId1).
		Return(mockExecutionManager, nil).
		Times(1)

	mockHistoryManager := persistence.NewMockHistoryManager(td.ctrl)
	mockHistoryManager.EXPECT().Close().Times(1)
	td.mockManagerFactory.EXPECT().
		initializeHistoryManager(gomock.Any()).
		Return(mockHistoryManager, nil).
		Times(1)

	mockExecutionManager.EXPECT().GetCurrentExecution(gomock.Any(), gomock.Any()).
		Return(&persistence.GetCurrentExecutionResponse{
			RunID: "test-run-id1",
			State: persistence.WorkflowStateCompleted,
		}, nil).
		Times(1)
	mockExecutionManager.EXPECT().GetShardID().Return(shardId1).Times(1)
	mockExecutionManager.EXPECT().IsWorkflowExecutionExists(gomock.Any(), gomock.Any()).
		Return(&persistence.IsWorkflowExecutionExistsResponse{
			Exists: true,
		}, nil).
		Times(1)
}

func TestAdminDBScanErrorCases(t *testing.T) {
	cases := []struct {
		name        string
		testSetup   func(td *cliTestData) *cli.Context
		errContains string // empty if no error is expected
	}{
		{
			name: "scan type not provided",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app)
			},
			errContains: "unknown scan type",
		},
		{
			name: "unknown scan type provided",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app,
					clitest.StringArgument("scan_type", "some_unknown_scan_type"),
				)
			},
			errContains: "unknown scan type",
		},
		{
			name: "number of shards not provided",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app,
					clitest.StringArgument("scan_type", "ConcreteExecutionType"),
				)
			},
			errContains: "Required flag not found",
		},
		{
			name: "invariant collection not provided",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app,
					clitest.StringArgument("scan_type", "ConcreteExecutionType"),
					clitest.IntArgument("number_of_shards", 16384),
				)
			},
			errContains: "no invariants for scan type \"ConcreteExecutionType\" and collections []",
		},
		{
			name: "invalid invariant collection provided",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app,
					clitest.StringArgument("scan_type", "ConcreteExecutionType"),
					clitest.IntArgument("number_of_shards", 16384),
					clitest.StringSliceArgument("invariant_collection", "some_unknown_invariant_collection"),
				)
			},
			errContains: "unknown invariant collection: some_unknown_invariant_collection",
		},
		{
			name: "input file not found",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app,
					clitest.StringArgument("scan_type", "ConcreteExecutionType"),
					clitest.IntArgument("number_of_shards", 16384),
					clitest.StringSliceArgument("invariant_collection", "CollectionHistory"),
					clitest.StringArgument("input_file", "testdata/non-existant-file.json"),
				)
			},
			errContains: "Input file not found",
		},
		{
			name: "input file is empty",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app,
					clitest.StringArgument("scan_type", "ConcreteExecutionType"),
					clitest.IntArgument("number_of_shards", 16384),
					clitest.StringSliceArgument("invariant_collection", "CollectionHistory"),
					clitest.StringArgument("input_file", "testdata/scan_input_empty.json"),
				)
			},
			errContains: "Input file contained no data to scan",
		},
		{
			name: "bad data in input file",
			testSetup: func(td *cliTestData) *cli.Context {
				return clitest.NewCLIContext(t, td.app,
					clitest.StringArgument("scan_type", "ConcreteExecutionType"),
					clitest.IntArgument("number_of_shards", 16384),
					clitest.StringSliceArgument("invariant_collection", "CollectionHistory"),
					clitest.StringArgument("input_file", "testdata/scan_input_bad_data.json"),
				)
			},
			errContains: "Error decoding input file",
		},
		{
			name: "execution manager initialization error",
			testSetup: func(td *cliTestData) *cli.Context {
				td.mockManagerFactory.EXPECT().
					initializeExecutionManager(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)

				return clitest.NewCLIContext(t, td.app,
					clitest.StringArgument("scan_type", "ConcreteExecutionType"),
					clitest.IntArgument("number_of_shards", 16384),
					clitest.StringSliceArgument("invariant_collection", "CollectionHistory"),
					clitest.StringArgument("input_file", "testdata/scan_input.json"),
				)
			},
			errContains: "Execution check failed: initialize execution manager: assert.AnError general error for testing",
		},
		{
			name: "historyV2 manager initialization error",
			testSetup: func(td *cliTestData) *cli.Context {
				shardId1 := common.WorkflowIDToHistoryShard("test-workflow-id1", 16384)
				mockExecutionManager := persistence.NewMockExecutionManager(td.ctrl)
				mockExecutionManager.EXPECT().Close()
				td.mockManagerFactory.EXPECT().
					initializeExecutionManager(gomock.Any(), shardId1).
					Return(mockExecutionManager, nil)

				td.mockManagerFactory.EXPECT().
					initializeHistoryManager(gomock.Any()).
					Return(nil, assert.AnError)

				return clitest.NewCLIContext(t, td.app,
					clitest.StringArgument("scan_type", "ConcreteExecutionType"),
					clitest.IntArgument("number_of_shards", 16384),
					clitest.StringSliceArgument("invariant_collection", "CollectionHistory"),
					clitest.StringArgument("input_file", "testdata/scan_input.json"),
				)
			},
			errContains: "Execution check failed: initialize history manager: assert.AnError general error for testing",
		},
		{
			name: "failed to fetch execution",
			testSetup: func(td *cliTestData) *cli.Context {
				shardId1 := common.WorkflowIDToHistoryShard("test-workflow-id1", 16384)
				mockExecutionManager := persistence.NewMockExecutionManager(td.ctrl)
				mockExecutionManager.EXPECT().Close()
				td.mockManagerFactory.EXPECT().
					initializeExecutionManager(gomock.Any(), shardId1).
					Return(mockExecutionManager, nil)

				mockHistoryManager := persistence.NewMockHistoryManager(td.ctrl)
				mockHistoryManager.EXPECT().Close()
				td.mockManagerFactory.EXPECT().
					initializeHistoryManager(gomock.Any()).
					Return(mockHistoryManager, nil)

				mockExecutionManager.EXPECT().GetWorkflowExecution(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError)

				return clitest.NewCLIContext(t, td.app,
					clitest.StringArgument("scan_type", "ConcreteExecutionType"),
					clitest.IntArgument("number_of_shards", 16384),
					clitest.StringSliceArgument("invariant_collection", "CollectionHistory"),
					clitest.StringArgument("input_file", "testdata/scan_input.json"),
				)
			},
			errContains: "Execution check failed: fetching execution: assert.AnError general error for testing",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			td := newCLITestData(t)
			testCtx := tc.testSetup(td)
			err := AdminDBScan(testCtx)
			if tc.errContains != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
