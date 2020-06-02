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

package executions

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/service/worker/scanner/executions/common"
)

type workflowsSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestWorkflowsSuite(t *testing.T) {
	suite.Run(t, new(workflowsSuite))
}

func (s *workflowsSuite) SetupSuite() {
	workflow.Register(ScannerWorkflow)
	activity.RegisterWithOptions(ScannerConfigActivity, activity.RegisterOptions{Name: ScannerConfigActivityName})
	activity.RegisterWithOptions(ScanShardActivity, activity.RegisterOptions{Name: ScannerScanShardActivityName})
	activity.RegisterWithOptions(ScannerEmitMetricsActivity, activity.RegisterOptions{Name: ScannerEmitMetricsActivityName})
}

func (s *workflowsSuite) TestScannerWorkflow_Failure_ScannerConfigActivity() {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(ScannerConfigActivityName, mock.Anything, mock.Anything).Return(ResolvedScannerWorkflowConfig{}, errors.New("got error getting config"))
	env.ExecuteWorkflow(ScannerWorkflow, ScannerWorkflowParams{
		Shards: Shards{
			List: []int{1, 2, 3},
		},
	})
	s.True(env.IsWorkflowCompleted())
	s.Equal("got error getting config", env.GetWorkflowError().Error())
}

func (s *workflowsSuite) TestScannerWorkflow_Success_Disabled() {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(ScannerConfigActivityName, mock.Anything, mock.Anything).Return(ResolvedScannerWorkflowConfig{
		Enabled: false,
	}, nil)
	env.ExecuteWorkflow(ScannerWorkflow, ScannerWorkflowParams{
		Shards: Shards{
			List: []int{1, 2, 3},
		},
	})
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *workflowsSuite) TestScannerWorkflow_Success() {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(ScannerConfigActivityName, mock.Anything, mock.Anything).Return(ResolvedScannerWorkflowConfig{
		Enabled:     true,
		Concurrency: 3,
	}, nil)
	env.OnActivity(ScannerEmitMetricsActivityName, mock.Anything, mock.Anything).Return(nil)
	shards := Shards{
		Range: &ShardRange{
			Min: 0,
			Max: 30,
		},
	}
	for i := 0; i < 30; i++ {
		if i%5 == 0 {
			env.OnActivity(ScannerScanShardActivityName, mock.Anything, ScanShardActivityParams{
				ShardID: i,
			}).Return(&common.ShardScanReport{
				ShardID: i,
				Stats: common.ShardScanStats{
					ExecutionsCount: 10,
				},
				Result: common.ShardScanResult{
					ControlFlowFailure: &common.ControlFlowFailure{
						Info: "got control flow failure",
					},
				},
			}, nil)
		} else {
			env.OnActivity(ScannerScanShardActivityName, mock.Anything, ScanShardActivityParams{
				ShardID: i,
			}).Return(&common.ShardScanReport{
				ShardID: i,
				Stats: common.ShardScanStats{
					ExecutionsCount:  10,
					CorruptedCount:   2,
					CheckFailedCount: 1,
					CorruptionByType: map[common.InvariantType]int64{
						common.HistoryExistsInvariantType:   1,
						common.ValidFirstEventInvariantType: 1,
					},
					CorruptedOpenExecutionCount: 0,
				},
				Result: common.ShardScanResult{
					ShardScanKeys: &common.ShardScanKeys{
						Corrupt: &common.Keys{
							UUID:    "test_uuid",
							MinPage: 0,
							MaxPage: i,
						},
					},
				},
			}, nil)
		}
	}
	env.ExecuteWorkflow(ScannerWorkflow, ScannerWorkflowParams{
		Shards: shards,
	})
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	aggValue, err := env.QueryWorkflow(AggregateReportQuery)
	s.NoError(err)
	var agg AggregateScanReportResult
	s.NoError(aggValue.Get(&agg))
	s.Equal(AggregateScanReportResult{
		ExecutionsCount:  240,
		CorruptedCount:   48,
		CheckFailedCount: 24,
		CorruptionByType: map[common.InvariantType]int64{
			common.HistoryExistsInvariantType:   24,
			common.ValidFirstEventInvariantType: 24,
		},
	}, agg)
	for i := 0; i < 30; i++ {
		shardReportValue, err := env.QueryWorkflow(ShardReportQuery, i)
		s.NoError(err)
		var shardReport *common.ShardScanReport
		s.NoError(shardReportValue.Get(&shardReport))
		if i%5 == 0 {
			s.Equal(&common.ShardScanReport{
				ShardID: i,
				Stats: common.ShardScanStats{
					ExecutionsCount: 10,
				},
				Result: common.ShardScanResult{
					ControlFlowFailure: &common.ControlFlowFailure{
						Info: "got control flow failure",
					},
				},
			}, shardReport)
		} else {
			s.Equal(&common.ShardScanReport{
				ShardID: i,
				Stats: common.ShardScanStats{
					ExecutionsCount:  10,
					CorruptedCount:   2,
					CheckFailedCount: 1,
					CorruptionByType: map[common.InvariantType]int64{
						common.HistoryExistsInvariantType:   1,
						common.ValidFirstEventInvariantType: 1,
					},
					CorruptedOpenExecutionCount: 0,
				},
				Result: common.ShardScanResult{
					ShardScanKeys: &common.ShardScanKeys{
						Corrupt: &common.Keys{
							UUID:    "test_uuid",
							MinPage: 0,
							MaxPage: i,
						},
					},
				},
			}, shardReport)
		}
	}
	statusValue, err := env.QueryWorkflow(ShardStatusQuery)
	s.NoError(err)
	var status ShardStatusResult
	s.NoError(statusValue.Get(&status))
	expected := make(map[int]ShardStatus)
	for i := 0; i < 30; i++ {
		if i%5 == 0 {
			expected[i] = ShardStatusControlFlowFailure
		} else {
			expected[i] = ShardStatusSuccess
		}
	}
	s.Equal(ShardStatusResult(expected), status)
	corruptionKeysValue, err := env.QueryWorkflow(ShardCorruptKeysQuery)
	s.NoError(err)
	var shardCorruptKeysResult ShardCorruptKeysResult
	s.NoError(corruptionKeysValue.Get(&shardCorruptKeysResult))
	expectedCorrupted := make(map[int]common.Keys)
	for i := 0; i < 30; i++ {
		if i%5 != 0 {
			expectedCorrupted[i] = common.Keys{
				UUID:    "test_uuid",
				MinPage: 0,
				MaxPage: i,
			}
		}
	}
	s.Equal(ShardCorruptKeysResult(expectedCorrupted), shardCorruptKeysResult)
}

func (s *workflowsSuite) TestScannerWorkflow_Failure_ScanShard() {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(ScannerConfigActivityName, mock.Anything, mock.Anything).Return(ResolvedScannerWorkflowConfig{
		Enabled:     true,
		Concurrency: 3,
	}, nil)
	shards := Shards{
		Range: &ShardRange{
			Min: 0,
			Max: 30,
		},
	}
	for i := 0; i < 30; i++ {
		if i%5 == 0 {
			env.OnActivity(ScannerScanShardActivityName, mock.Anything, ScanShardActivityParams{
				ShardID: i,
			}).Return(&common.ShardScanReport{
				ShardID: i,
				Stats: common.ShardScanStats{
					ExecutionsCount: 10,
				},
				Result: common.ShardScanResult{
					ControlFlowFailure: &common.ControlFlowFailure{
						Info: "got control flow failure",
					},
				},
			}, nil)
		} else if i < 29 {
			env.OnActivity(ScannerScanShardActivityName, mock.Anything, ScanShardActivityParams{
				ShardID: i,
			}).Return(&common.ShardScanReport{
				ShardID: i,
				Stats: common.ShardScanStats{
					ExecutionsCount:  10,
					CorruptedCount:   2,
					CheckFailedCount: 1,
					CorruptionByType: map[common.InvariantType]int64{
						common.HistoryExistsInvariantType:   1,
						common.ValidFirstEventInvariantType: 1,
					},
					CorruptedOpenExecutionCount: 0,
				},
			}, nil)
		} else {
			env.OnActivity(ScannerScanShardActivityName, mock.Anything, ScanShardActivityParams{
				ShardID: i,
			}).Return(nil, errors.New("scan shard activity got error"))
		}
	}
	env.ExecuteWorkflow(ScannerWorkflow, ScannerWorkflowParams{
		Shards: shards,
	})
	s.True(env.IsWorkflowCompleted())
	s.Error(env.GetWorkflowError())
}
