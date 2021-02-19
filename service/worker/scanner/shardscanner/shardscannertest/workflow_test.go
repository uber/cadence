// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
//
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

package shardscannertest

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/reconciliation/store"
	"github.com/uber/cadence/service/worker/scanner/shardscanner"
)

type workflowsSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestScannerWorkflowSuite(t *testing.T) {
	suite.Run(t, new(workflowsSuite))
}

func (s *workflowsSuite) SetupSuite() {
	workflow.Register(NewTestWorkflow)
	workflow.Register(NewTestFixerWorkflow)
	workflow.Register(shardscanner.GetCorruptedKeys)
}

func (s *workflowsSuite) TestScannerWorkflow_Failure_ScanShard() {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(shardscanner.ActivityScannerConfig, mock.Anything, mock.Anything).Return(shardscanner.ResolvedScannerWorkflowConfig{
		GenericScannerConfig: shardscanner.GenericScannerConfig{
			Enabled:           true,
			Concurrency:       3,
			ActivityBatchSize: 5,
		},
	}, nil)
	shards := shardscanner.Shards{
		Range: &shardscanner.ShardRange{
			Min: 0,
			Max: 30,
		},
	}

	batches := [][]int{
		{0, 3, 6, 9, 12},
		{15, 18, 21, 24, 27},
		{1, 4, 7, 10, 13},
		{16, 19, 22, 25, 28},
		{2, 5, 8, 11, 14},
		{17, 20, 23, 26, 29},
	}

	for i, batch := range batches {

		var reports []shardscanner.ScanReport
		var err error
		if i == len(batches)-1 {
			reports = nil
			err = errors.New("scan shard activity got error")
		} else {
			err = nil
			for _, s := range batch {
				reports = append(reports, shardscanner.ScanReport{
					ShardID: s,
					Stats: shardscanner.ScanStats{
						EntitiesCount: 10,
					},
					Result: shardscanner.ScanResult{
						ControlFlowFailure: &shardscanner.ControlFlowFailure{
							Info: "got control flow failure",
						},
					},
				})
			}
		}
		env.OnActivity(shardscanner.ActivityScanShard, mock.Anything, shardscanner.ScanShardActivityParams{
			Shards: batch,
		}).Return(reports, err)
	}
	env.ExecuteWorkflow(NewTestWorkflow, "test-workflow", shardscanner.ScannerWorkflowParams{
		Shards: shards,
	})
	s.True(env.IsWorkflowCompleted())
	s.Equal("scan shard activity got error", env.GetWorkflowError().Error())
}

func (s *workflowsSuite) TestScannerWorkflow_Failure_ScannerConfigActivity() {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(shardscanner.ActivityScannerConfig, mock.Anything, mock.Anything).Return(shardscanner.ResolvedScannerWorkflowConfig{}, errors.New("got error getting config"))
	env.ExecuteWorkflow(NewTestWorkflow, "test-workflow", shardscanner.ScannerWorkflowParams{
		Shards: shardscanner.Shards{
			List: []int{1, 2, 3},
		},
	})
	s.True(env.IsWorkflowCompleted())
	s.Equal("got error getting config", env.GetWorkflowError().Error())
}

func (s *workflowsSuite) TestScannerWorkflow_Requires_Name() {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(shardscanner.ActivityScannerConfig, mock.Anything, mock.Anything).Return(shardscanner.ResolvedScannerWorkflowConfig{}, errors.New("got error getting config"))
	env.ExecuteWorkflow(NewTestWorkflow, "", shardscanner.ScannerWorkflowParams{
		Shards: shardscanner.Shards{
			List: []int{1, 2, 3},
		},
	})
	s.True(env.IsWorkflowCompleted())
	s.Equal("workflow name is not provided", env.GetWorkflowError().Error())
}

func (s *workflowsSuite) TestScannerWorkflow_Requires_Valid_ShardConfig() {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(shardscanner.ActivityScannerConfig, mock.Anything, mock.Anything).Return(shardscanner.ResolvedScannerWorkflowConfig{}, errors.New("got error getting config"))
	env.ExecuteWorkflow(NewTestWorkflow, "test-workflow", shardscanner.ScannerWorkflowParams{})
	s.True(env.IsWorkflowCompleted())
	s.Equal("must provide either List or Range", env.GetWorkflowError().Error())
}

func (s *workflowsSuite) TestScannerWorkflow_Success_Disabled() {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(shardscanner.ActivityScannerConfig, mock.Anything, mock.Anything).Return(shardscanner.ResolvedScannerWorkflowConfig{
		GenericScannerConfig: shardscanner.GenericScannerConfig{
			Enabled: false,
		},
	}, nil)

	env.ExecuteWorkflow(NewTestWorkflow, "test-workflow", shardscanner.ScannerWorkflowParams{
		Shards: shardscanner.Shards{
			List: []int{1, 2, 3},
		},
	})

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *workflowsSuite) TestFixerWorkflow_Success() {
	env := s.NewTestWorkflowEnvironment()
	corruptedKeys := make([]shardscanner.CorruptedKeysEntry, 30, 30)
	for i := 0; i < 30; i++ {
		corruptedKeys[i] = shardscanner.CorruptedKeysEntry{
			ShardID: i,
		}
	}
	env.OnActivity(shardscanner.ActivityFixerCorruptedKeys, mock.Anything, mock.Anything).Return(&shardscanner.FixerCorruptedKeysActivityResult{
		CorruptedKeys: corruptedKeys,
		MinShard:      common.IntPtr(0),
		MaxShard:      common.IntPtr(29),
		ShardQueryPaginationToken: shardscanner.ShardQueryPaginationToken{
			IsDone:      true,
			NextShardID: nil,
		},
	}, nil)

	fixerWorkflowConfigOverwrites := shardscanner.FixerWorkflowConfigOverwrites{
		Concurrency:             common.IntPtr(3),
		BlobstoreFlushThreshold: common.IntPtr(1000),
		ActivityBatchSize:       common.IntPtr(5),
	}
	resolvedFixerWorkflowConfig := shardscanner.ResolvedFixerWorkflowConfig{
		Concurrency:             3,
		ActivityBatchSize:       5,
		BlobstoreFlushThreshold: 1000,
	}
	batches := [][]int{
		{0, 3, 6, 9, 12},
		{15, 18, 21, 24, 27},
		{1, 4, 7, 10, 13},
		{16, 19, 22, 25, 28},
		{2, 5, 8, 11, 14},
		{17, 20, 23, 26, 29},
	}

	for _, batch := range batches {
		var corruptedKeys []shardscanner.CorruptedKeysEntry
		for _, shard := range batch {
			corruptedKeys = append(corruptedKeys, shardscanner.CorruptedKeysEntry{
				ShardID: shard,
			})
		}
		var reports []shardscanner.FixReport
		for i, s := range batch {
			if i == 0 {
				reports = append(reports, shardscanner.FixReport{
					ShardID: s,
					Stats: shardscanner.FixStats{
						EntitiesCount: 10,
					},
					Result: shardscanner.FixResult{
						ControlFlowFailure: &shardscanner.ControlFlowFailure{
							Info: "got control flow failure",
						},
					},
				})
			} else {
				reports = append(reports, shardscanner.FixReport{
					ShardID: s,
					Stats: shardscanner.FixStats{
						EntitiesCount: 10,
						FixedCount:    2,
						SkippedCount:  1,
						FailedCount:   1,
					},
					Result: shardscanner.FixResult{
						ShardFixKeys: &shardscanner.FixKeys{
							Skipped: &store.Keys{
								UUID: "skipped_keys",
							},
							Failed: &store.Keys{
								UUID: "failed_keys",
							},
							Fixed: &store.Keys{
								UUID: "fixed_keys",
							},
						},
					},
				})
			}
		}
		env.OnActivity(shardscanner.ActivityFixShard, mock.Anything, shardscanner.FixShardActivityParams{
			CorruptedKeysEntries:        corruptedKeys,
			ResolvedFixerWorkflowConfig: resolvedFixerWorkflowConfig,
		}).Return(reports, nil)
	}

	env.ExecuteWorkflow(NewTestFixerWorkflow, shardscanner.FixerWorkflowParams{
		ScannerWorkflowWorkflowID:     "test_wid",
		ScannerWorkflowRunID:          "test_rid",
		FixerWorkflowConfigOverwrites: fixerWorkflowConfigOverwrites,
	})
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	aggValue, err := env.QueryWorkflow(shardscanner.AggregateReportQuery)
	s.NoError(err)
	var agg shardscanner.AggregateFixReportResult
	s.NoError(aggValue.Get(&agg))
	s.Equal(shardscanner.AggregateFixReportResult{
		EntitiesCount: 240,
		FixedCount:    48,
		FailedCount:   24,
		SkippedCount:  24,
	}, agg)

	for i := 0; i < 30; i++ {
		shardReportValue, err := env.QueryWorkflow(shardscanner.ShardReportQuery, i)
		s.NoError(err)
		var shardReport *shardscanner.FixReport
		s.NoError(shardReportValue.Get(&shardReport))
		if i == 0 || i == 1 || i == 2 || i == 15 || i == 16 || i == 17 {
			s.Equal(&shardscanner.FixReport{
				ShardID: i,
				Stats: shardscanner.FixStats{
					EntitiesCount: 10,
				},
				Result: shardscanner.FixResult{
					ControlFlowFailure: &shardscanner.ControlFlowFailure{
						Info: "got control flow failure",
					},
				},
			}, shardReport)
		} else {
			s.Equal(&shardscanner.FixReport{
				ShardID: i,
				Stats: shardscanner.FixStats{
					EntitiesCount: 10,
					FixedCount:    2,
					FailedCount:   1,
					SkippedCount:  1,
				},
				Result: shardscanner.FixResult{
					ShardFixKeys: &shardscanner.FixKeys{
						Skipped: &store.Keys{
							UUID: "skipped_keys",
						},
						Failed: &store.Keys{
							UUID: "failed_keys",
						},
						Fixed: &store.Keys{
							UUID: "fixed_keys",
						},
					},
				},
			}, shardReport)
		}
	}

	statusValue, err := env.QueryWorkflow(shardscanner.ShardStatusQuery, shardscanner.PaginatedShardQueryRequest{})
	s.NoError(err)
	var status *shardscanner.ShardStatusQueryResult
	s.NoError(statusValue.Get(&status))
	expected := make(map[int]shardscanner.ShardStatus)
	for i := 0; i < 30; i++ {
		if i == 0 || i == 1 || i == 2 || i == 15 || i == 16 || i == 17 {
			expected[i] = shardscanner.ShardStatusControlFlowFailure
		} else {
			expected[i] = shardscanner.ShardStatusSuccess
		}
	}
	s.Equal(shardscanner.ShardStatusResult(expected), status.Result)

	// check for paginated query result
	statusValue, err = env.QueryWorkflow(shardscanner.ShardStatusQuery, shardscanner.PaginatedShardQueryRequest{
		StartingShardID: common.IntPtr(5),
		LimitShards:     common.IntPtr(10),
	})
	s.NoError(err)
	status = &shardscanner.ShardStatusQueryResult{}
	s.NoError(statusValue.Get(&status))
	expected = make(map[int]shardscanner.ShardStatus)
	for i := 5; i < 15; i++ {
		if i == 0 || i == 1 || i == 2 || i == 15 || i == 16 || i == 17 {
			expected[i] = shardscanner.ShardStatusControlFlowFailure
		} else {
			expected[i] = shardscanner.ShardStatusSuccess
		}
	}
	s.Equal(shardscanner.ShardStatusResult(expected), status.Result)
	s.False(status.ShardQueryPaginationToken.IsDone)
	s.Equal(15, *status.ShardQueryPaginationToken.NextShardID)
}

func (s *workflowsSuite) TestGetCorruptedKeys_Success() {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(shardscanner.ActivityFixerCorruptedKeys, mock.Anything, shardscanner.FixerCorruptedKeysActivityParams{
		ScannerWorkflowWorkflowID: "test_wid",
		ScannerWorkflowRunID:      "test_rid",
		StartingShardID:           nil,
	}).Return(&shardscanner.FixerCorruptedKeysActivityResult{
		CorruptedKeys: []shardscanner.CorruptedKeysEntry{{ShardID: 1}, {ShardID: 5}, {ShardID: 10}},
		MinShard:      common.IntPtr(1),
		MaxShard:      common.IntPtr(10),
		ShardQueryPaginationToken: shardscanner.ShardQueryPaginationToken{
			NextShardID: common.IntPtr(11),
			IsDone:      false,
		},
	}, nil)
	env.OnActivity(shardscanner.ActivityFixerCorruptedKeys, mock.Anything, shardscanner.FixerCorruptedKeysActivityParams{
		ScannerWorkflowWorkflowID: "test_wid",
		ScannerWorkflowRunID:      "test_rid",
		StartingShardID:           common.IntPtr(11),
	}).Return(&shardscanner.FixerCorruptedKeysActivityResult{
		CorruptedKeys: []shardscanner.CorruptedKeysEntry{{ShardID: 11}, {ShardID: 12}},
		MinShard:      common.IntPtr(11),
		MaxShard:      common.IntPtr(12),
		ShardQueryPaginationToken: shardscanner.ShardQueryPaginationToken{
			NextShardID: common.IntPtr(13),
			IsDone:      false,
		},
	}, nil)
	env.OnActivity(shardscanner.ActivityFixerCorruptedKeys, mock.Anything, shardscanner.FixerCorruptedKeysActivityParams{
		ScannerWorkflowWorkflowID: "test_wid",
		ScannerWorkflowRunID:      "test_rid",
		StartingShardID:           common.IntPtr(13),
	}).Return(&shardscanner.FixerCorruptedKeysActivityResult{
		CorruptedKeys: []shardscanner.CorruptedKeysEntry{{ShardID: 20}, {ShardID: 41}},
		MinShard:      common.IntPtr(20),
		MaxShard:      common.IntPtr(41),
		ShardQueryPaginationToken: shardscanner.ShardQueryPaginationToken{
			NextShardID: common.IntPtr(42),
			IsDone:      false,
		},
	}, nil)
	env.OnActivity(shardscanner.ActivityFixerCorruptedKeys, mock.Anything, shardscanner.FixerCorruptedKeysActivityParams{
		ScannerWorkflowWorkflowID: "test_wid",
		ScannerWorkflowRunID:      "test_rid",
		StartingShardID:           common.IntPtr(42),
	}).Return(&shardscanner.FixerCorruptedKeysActivityResult{
		CorruptedKeys: []shardscanner.CorruptedKeysEntry{},
		MinShard:      nil,
		MaxShard:      nil,
		ShardQueryPaginationToken: shardscanner.ShardQueryPaginationToken{
			NextShardID: nil,
			IsDone:      true,
		},
	}, nil)

	env.ExecuteWorkflow(shardscanner.GetCorruptedKeys, shardscanner.FixerWorkflowParams{
		ScannerWorkflowWorkflowID: "test_wid",
		ScannerWorkflowRunID:      "test_rid",
	})
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var result *shardscanner.FixerCorruptedKeysActivityResult
	s.NoError(env.GetWorkflowResult(&result))
	s.Equal(&shardscanner.FixerCorruptedKeysActivityResult{
		CorruptedKeys: []shardscanner.CorruptedKeysEntry{
			{ShardID: 1},
			{ShardID: 5},
			{ShardID: 10},
			{ShardID: 11},
			{ShardID: 12},
			{ShardID: 20},
			{ShardID: 41},
		},
		MinShard: common.IntPtr(1),
		MaxShard: common.IntPtr(41),
		ShardQueryPaginationToken: shardscanner.ShardQueryPaginationToken{
			NextShardID: nil,
			IsDone:      true,
		},
	}, result)
}

func (s *workflowsSuite) TestGetCorruptedKeys_Error() {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(shardscanner.ActivityFixerCorruptedKeys, mock.Anything, shardscanner.FixerCorruptedKeysActivityParams{
		ScannerWorkflowWorkflowID: "test_wid",
		ScannerWorkflowRunID:      "test_rid",
		StartingShardID:           nil,
	}).Return(&shardscanner.FixerCorruptedKeysActivityResult{
		CorruptedKeys: []shardscanner.CorruptedKeysEntry{{ShardID: 1}, {ShardID: 5}, {ShardID: 10}},
		MinShard:      common.IntPtr(1),
		MaxShard:      common.IntPtr(10),
		ShardQueryPaginationToken: shardscanner.ShardQueryPaginationToken{
			NextShardID: common.IntPtr(11),
			IsDone:      false,
		},
	}, nil)
	env.OnActivity(shardscanner.ActivityFixerCorruptedKeys, mock.Anything, shardscanner.FixerCorruptedKeysActivityParams{
		ScannerWorkflowWorkflowID: "test_wid",
		ScannerWorkflowRunID:      "test_rid",
		StartingShardID:           common.IntPtr(11),
	}).Return(nil, errors.New("got error"))
	env.ExecuteWorkflow(shardscanner.GetCorruptedKeys, shardscanner.FixerWorkflowParams{
		ScannerWorkflowWorkflowID: "test_wid",
		ScannerWorkflowRunID:      "test_rid",
	})
	s.True(env.IsWorkflowCompleted())
	s.Error(env.GetWorkflowError())
}

func (s *workflowsSuite) TestScannerWorkflow_Failure_CorruptedKeysActivity() {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(shardscanner.ActivityFixerCorruptedKeys, mock.Anything, mock.Anything).Return(nil, errors.New("got error getting corrupted keys"))
	env.ExecuteWorkflow(NewTestFixerWorkflow, shardscanner.FixerWorkflowParams{})
	s.True(env.IsWorkflowCompleted())
	s.Equal("got error getting corrupted keys", env.GetWorkflowError().Error())
}
