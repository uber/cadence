package executions

import (
	"fmt"

	"github.com/uber/cadence/service/worker/scanner/executions/common"
)

type shardFixResultAggregator struct {
	reports     map[int]common.ShardFixReport
	status      ShardStatusResult
	aggregation AggregateFixReportResult
}

func newShardFixResultAggregator(shards []int) *shardFixResultAggregator {
	status := make(map[int]ShardStatus)
	for _, s := range shards {
		status[s] = ShardStatusRunning
	}
	return &shardFixResultAggregator{
		reports:     make(map[int]common.ShardFixReport),
		status:      status,
		aggregation: AggregateFixReportResult{},
	}
}

func (a *shardFixResultAggregator) addReport(report common.ShardFixReport) {
	a.removeReport(report.ShardID)
	a.reports[report.ShardID] = report
	if report.Result.ControlFlowFailure != nil {
		a.status[report.ShardID] = ShardStatusControlFlowFailure
	} else {
		a.status[report.ShardID] = ShardStatusSuccess
	}
	if report.Result.ShardFixKeys != nil {
		a.adjustAggregation(report.Stats, func(a, b int64) int64 { return a + b })
	}
}

func (a *shardFixResultAggregator) removeReport(shardID int) {
	report, ok := a.reports[shardID]
	if !ok {
		return
	}
	delete(a.reports, shardID)
	delete(a.status, shardID)
	if report.Result.ShardFixKeys != nil {
		a.adjustAggregation(report.Stats, func(a, b int64) int64 { return a - b })
	}
}

func (a *shardFixResultAggregator) getReport(shardID int) (*common.ShardFixReport, error) {
	if report, ok := a.reports[shardID]; ok {
		return &report, nil
	}
	if _, ok := a.status[shardID]; !ok {
		return nil, fmt.Errorf("shard %v is not included in the shards that will be processed", shardID)
	}
	return nil, fmt.Errorf("shard %v has not finished yet, check back later for report", shardID)
}

func (a *shardFixResultAggregator) adjustAggregation(stats common.ShardFixStats, fn func(a, b int64) int64) {
	a.aggregation.ExecutionCount = fn(a.aggregation.ExecutionCount, stats.ExecutionCount)
	a.aggregation.SkippedCount = fn(a.aggregation.SkippedCount, stats.SkippedCount)
	a.aggregation.FailedCount = fn(a.aggregation.FailedCount, stats.FailedCount)
	a.aggregation.FixedCount = fn(a.aggregation.FixedCount, stats.FixedCount)
}

type shardScanResultAggregator struct {
	reports        map[int]common.ShardScanReport
	status         ShardStatusResult
	aggregation    AggregateScanReportResult
	corruptionKeys map[int]common.Keys
}

func newShardScanResultAggregator(shards []int) *shardScanResultAggregator {
	status := make(map[int]ShardStatus)
	for _, s := range shards {
		status[s] = ShardStatusRunning
	}
	return &shardScanResultAggregator{
		reports: make(map[int]common.ShardScanReport),
		status:  status,
		aggregation: AggregateScanReportResult{
			CorruptionByType: make(map[common.InvariantType]int64),
		},
		corruptionKeys: make(map[int]common.Keys),
	}
}

func (a *shardScanResultAggregator) addReport(report common.ShardScanReport) {
	a.removeReport(report.ShardID)
	a.reports[report.ShardID] = report
	if report.Result.ControlFlowFailure != nil {
		a.status[report.ShardID] = ShardStatusControlFlowFailure
	} else {
		a.status[report.ShardID] = ShardStatusSuccess
	}
	if report.Result.ShardScanKeys != nil {
		a.adjustAggregation(report.Stats, func(a, b int64) int64 { return a + b })
		if report.Result.ShardScanKeys.Corrupt != nil {
			a.corruptionKeys[report.ShardID] = *report.Result.ShardScanKeys.Corrupt
		}
	}
}

func (a *shardScanResultAggregator) removeReport(shardID int) {
	report, ok := a.reports[shardID]
	if !ok {
		return
	}
	delete(a.reports, shardID)
	delete(a.status, shardID)
	delete(a.corruptionKeys, shardID)
	if report.Result.ShardScanKeys != nil {
		a.adjustAggregation(report.Stats, func(a, b int64) int64 { return a - b })
	}
}

func (a *shardScanResultAggregator) getReport(shardID int) (*common.ShardScanReport, error) {
	if report, ok := a.reports[shardID]; ok {
		return &report, nil
	}
	if _, ok := a.status[shardID]; !ok {
		return nil, fmt.Errorf("shard %v is not included in the shards that will be processed", shardID)
	}
	return nil, fmt.Errorf("shard %v has not finished yet, check back later for report", shardID)
}

func (a *shardScanResultAggregator) adjustAggregation(stats common.ShardScanStats, fn func(a, b int64) int64) {
	a.aggregation.ExecutionsCount = fn(a.aggregation.ExecutionsCount, stats.ExecutionsCount)
	a.aggregation.CorruptedCount = fn(a.aggregation.CorruptedCount, stats.CorruptedCount)
	a.aggregation.CheckFailedCount = fn(a.aggregation.CheckFailedCount, stats.CheckFailedCount)
	a.aggregation.CorruptedOpenExecutionCount = fn(a.aggregation.CorruptedOpenExecutionCount, stats.CorruptedOpenExecutionCount)
	for k, v := range stats.CorruptionByType {
		a.aggregation.CorruptionByType[k] = fn(a.aggregation.CorruptionByType[k], v)
	}
}
