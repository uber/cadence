package executions

import (
	"fmt"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/service/worker/scanner/executions/common"
	"go.uber.org/cadence/workflow"
)

const (
	ScannerContextKey = ContextKey(0)

	ShardReportQuery = "shard_report"
	ShardStatusQuery = "shard_status_query"
	AggregateReportQuery = "aggregate_report"

	ShardStatusRunning ShardStatus = "running"
	ShardStatusSuccess ShardStatus = "success"
	ShardStatusControlFlowFailure ShardStatus = "control_flow_failure"
)

type (
	ContextKey int

	ScannerContext struct {
		Resource resource.Resource
		ScannerWorkflowDynamicConfig *ScannerWorkflowDynamicConfig
	}

	ScannerWorkflowParams struct {
		Shards Shards
	}

	Shards struct {
		List []int
		Range *ShardRange
	}

	ShardRange struct {
		Min int
		Max int
	}

	ShardStatusResult map[int]ShardStatus

	AggregateReportResult common.ShardScanStats

	ShardStatus string
)

// ScannerWorkflow is the workflow that scans over all concrete executions
func ScannerWorkflow(
	ctx workflow.Context,
	params ScannerWorkflowParams,
) error {
	shards := flattenShards(params.Shards)
	aggregator := newShardResultAggregator(shards)
	if err := workflow.SetQueryHandler(ctx, ShardReportQuery, func(shardID int) (*common.ShardScanReport, error) {
		return aggregator.getReport(shardID)
	}); err != nil {
		return err
	}
	if err := workflow.SetQueryHandler(ctx, ShardStatusQuery, func() (ShardStatusResult, error) {
		return aggregator.status, nil
	}); err != nil {
		return err
	}
	if err := workflow.SetQueryHandler(ctx, AggregateReportQuery, func() (AggregateReportResult, error) {
		return aggregator.aggregation, nil
	}); err != nil {
		return err
	}







	// 2. start some number of cadence go routines
	// 3. iterate over all shards and for each one a goroutine will run the activity
	// 4. keep track of a map from shardID -> report, this will get used in query
	// 5. there should also be queries to get aggregate information this will also be kept in workflow and updated each time a shard finishes



	return nil
}

type shardResultAggregator struct {
	reports map[int]common.ShardScanReport
	status ShardStatusResult
	aggregation AggregateReportResult
}

func newShardResultAggregator(shards []int) *shardResultAggregator {
	status := make(map[int]ShardStatus)
	for _, s := range shards {
		status[s] = ShardStatusRunning
	}
	return &shardResultAggregator{
		reports: make(map[int]common.ShardScanReport),
		status: status,
		aggregation: AggregateReportResult{
			CorruptionByType: make(map[common.InvariantType]int64),
		},
	}
}

func (a *shardResultAggregator) addReport(report common.ShardScanReport) {
	a.removeReport(report.ShardID)
	a.reports[report.ShardID] = report
	if report.Result.ControlFlowFailure != nil {
		a.status[report.ShardID] = ShardStatusControlFlowFailure
	} else {
		a.status[report.ShardID] = ShardStatusSuccess
	}
	a.adjustAggregation(report.Stats, func(a, b int64) int64 { return a + b })
}

func (a *shardResultAggregator) removeReport(shardID int) {
	report, ok := a.reports[shardID]
	if !ok {
		return
	}
	delete(a.reports, shardID)
	delete(a.status, shardID)
	a.adjustAggregation(report.Stats, func(a, b int64) int64 { return a - b })
}

func (a *shardResultAggregator) getReport(shardID int) (*common.ShardScanReport, error) {
	if report, ok := a.reports[shardID]; ok {
		return &report, nil
	}
	return nil, fmt.Errorf("shard %v has not finished yet, check back later for report", shardID)
}

func (a *shardResultAggregator) adjustAggregation(stats common.ShardScanStats, fn func(a, b int64) int64) {
	a.aggregation.ExecutionsCount = fn(a.aggregation.ExecutionsCount, stats.ExecutionsCount)
	a.aggregation.CorruptedCount = fn(a.aggregation.CorruptedCount, stats.CorruptedCount)
	a.aggregation.CheckFailedCount = fn(a.aggregation.CheckFailedCount, stats.CheckFailedCount)
	a.aggregation.CorruptedOpenExecutionCount = fn(a.aggregation.CorruptedOpenExecutionCount, stats.CorruptedOpenExecutionCount)
	for k, v := range stats.CorruptionByType {
		a.aggregation.CorruptionByType[k] = fn(a.aggregation.CorruptionByType[k], v)
	}
}

func flattenShards(shards Shards) []int {
	if len(shards.List) != 0 {
		return shards.List
	}
	var result []int
	for i := shards.Range.Min; i < shards.Range.Max; i++ {
		result = append(result, i)
	}
	return result
}