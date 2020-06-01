package executions

import (
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/uber/cadence/service/worker/scanner/executions/common"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/workflow"
	"testing"
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
		Enabled: true,
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
		if i % 5 == 0 {
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
					ExecutionsCount: 10,
					CorruptedCount: 2,
					CheckFailedCount: 1,
					CorruptionByType: map[common.InvariantType]int64{
						common.HistoryExistsInvariantType: 1,
						common.ValidFirstEventInvariantType: 1,
					},
					CorruptedOpenExecutionCount: 0,
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
	var agg AggregateReportResult
	s.NoError(aggValue.Get(&agg))
	s.Equal(AggregateReportResult{
		ExecutionsCount: 240,
		CorruptedCount: 48,
		CheckFailedCount: 24,
		CorruptionByType: map[common.InvariantType]int64{
			common.HistoryExistsInvariantType: 24,
			common.ValidFirstEventInvariantType: 24,
		},
	}, agg)
	for i := 0; i < 30; i++ {
		shardReportValue, err := env.QueryWorkflow(ShardReportQuery, i)
		s.NoError(err)
		var shardReport *common.ShardScanReport
		s.NoError(shardReportValue.Get(&shardReport))
		if i % 5 == 0 {
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
					ExecutionsCount: 10,
					CorruptedCount: 2,
					CheckFailedCount: 1,
					CorruptionByType: map[common.InvariantType]int64{
						common.HistoryExistsInvariantType: 1,
						common.ValidFirstEventInvariantType: 1,
					},
					CorruptedOpenExecutionCount: 0,
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
		if i % 5 == 0 {
			expected[i] = ShardStatusControlFlowFailure
		} else {
			expected[i] = ShardStatusSuccess
		}
	}
	s.Equal(ShardStatusResult(expected), status)
}

func (s *workflowsSuite) TestScannerWorkflow_Failure_ScanShard() {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(ScannerConfigActivityName, mock.Anything, mock.Anything).Return(ResolvedScannerWorkflowConfig{
		Enabled: true,
		Concurrency: 3,
	}, nil)
	shards := Shards{
		Range: &ShardRange{
			Min: 0,
			Max: 30,
		},
	}
	for i := 0; i < 30; i++ {
		if i % 5 == 0 {
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
					ExecutionsCount: 10,
					CorruptedCount: 2,
					CheckFailedCount: 1,
					CorruptionByType: map[common.InvariantType]int64{
						common.HistoryExistsInvariantType: 1,
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

func (s *workflowsSuite) TestFlattenShards() {
	testCases := []struct{
		input Shards
		expected []int
	}{
		{
			input: Shards{
				List: []int{1, 2, 3},
			},
			expected: []int{1, 2, 3},
		},
		{
			input: Shards{
				Range: &ShardRange{
					Min: 5,
					Max: 10,
				},
			},
			expected: []int{5, 6, 7, 8, 9},
		},
	}
	for _, tc := range testCases {
		s.Equal(tc.expected, flattenShards(tc.input))
	}
}

func (s *workflowsSuite) TestShardResultAggregator() {
	agg := newShardResultAggregator([]int{1, 2, 3})
	expected := &shardResultAggregator{
		reports: map[int]common.ShardScanReport{},
		status: map[int]ShardStatus{
			1: ShardStatusRunning,
			2: ShardStatusRunning,
			3: ShardStatusRunning,
		},
		aggregation: AggregateReportResult{
			CorruptionByType: make(map[common.InvariantType]int64),
		},
	}
	s.Equal(expected, agg)
	report, err := agg.getReport(1)
	s.Nil(report)
	s.Error(err)
	firstReport := common.ShardScanReport{
		ShardID: 1,
		Stats: common.ShardScanStats{
			ExecutionsCount: 10,
			CorruptedCount: 3,
			CheckFailedCount: 1,
			CorruptionByType: map[common.InvariantType]int64{
				common.HistoryExistsInvariantType: 2,
				common.OpenCurrentExecutionInvariantType: 1,
			},
			CorruptedOpenExecutionCount: 1,
		},
		Result: common.ShardScanResult{
			ShardScanKeys: &common.ShardScanKeys{},
		},
	}
	agg.addReport(firstReport)
	expected.status[1] = ShardStatusSuccess
	expected.reports[1] = firstReport
	expected.aggregation.ExecutionsCount = 10
	expected.aggregation.CorruptedCount = 3
	expected.aggregation.CheckFailedCount = 1
	expected.aggregation.CorruptionByType = map[common.InvariantType]int64{
		common.HistoryExistsInvariantType: 2,
		common.OpenCurrentExecutionInvariantType: 1,
	}
	expected.aggregation.CorruptedOpenExecutionCount = 1
	s.Equal(expected, agg)
	agg.addReport(firstReport)
	s.Equal(expected, agg)
	report, err = agg.getReport(1)
	s.NoError(err)
	s.Equal(firstReport, *report)
	secondReport := common.ShardScanReport{
		ShardID: 2,
		Stats: common.ShardScanStats{
			ExecutionsCount: 10,
			CorruptedCount: 3,
			CheckFailedCount: 1,
			CorruptionByType: map[common.InvariantType]int64{
				common.HistoryExistsInvariantType: 2,
				common.OpenCurrentExecutionInvariantType: 1,
			},
			CorruptedOpenExecutionCount: 1,
		},
		Result: common.ShardScanResult{
			ControlFlowFailure: &common.ControlFlowFailure{},
		},
	}
	agg.addReport(secondReport)
	expected.status[2] = ShardStatusControlFlowFailure
	expected.reports[2] = secondReport
	s.Equal(expected, agg)
}