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
	"fmt"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/service/worker/scanner/executions/common"
	"go.uber.org/cadence/workflow"
	"time"
)

const (
	// ScannerContextKey is the key used to access ScannerContext in activities
	ScannerContextKey = ContextKey(0)

	// ShardReportQuery is the query name for the query used to get a single shard's report
	ShardReportQuery = "shard_report"
	// ShardStatusQuery is the query name for the query used to get the status of all shards
	ShardStatusQuery = "shard_status_query"
	// AggregateReportQuery is the query name for the query used to get the aggregate result of all finished shards
	AggregateReportQuery = "aggregate_report"

	// ShardStatusRunning indicates the shard has not completed yet
	ShardStatusRunning ShardStatus = "running"
	// ShardStatusSuccess indicates the scan on the shard ran successfully
	ShardStatusSuccess ShardStatus = "success"
	// ShardStatusControlFlowFailure indicates the scan on the shard failed
	ShardStatusControlFlowFailure ShardStatus = "control_flow_failure"

	shardReportChan = "share"
)

type (
	// ContextKey is the type which identifies context keys
	ContextKey int

	// ScannerContext is the resource that is available in activities under ScannerContextKey context key
	ScannerContext struct {
		Resource resource.Resource
		Scope metrics.Scope
		ScannerWorkflowDynamicConfig *ScannerWorkflowDynamicConfig
	}

	// ScannerWorkflowParams are the parameters to the scan workflow
	ScannerWorkflowParams struct {
		Shards Shards
		ScannerWorkflowConfigOverwrites ScannerWorkflowConfigOverwrites
	}

	// Shards identify the shards that should be scanned.
	// Exactly one of List of Range should be non-nil.
	Shards struct {
		List []int
		Range *ShardRange
	}

	// ShardRange identifies a set of shards based on min (inclusive) and max (exclusive)
	ShardRange struct {
		Min int
		Max int
	}

	// ShardStatusResult indicates the status for all shards
	ShardStatusResult map[int]ShardStatus

	// AggregateReportResult indicates the result of summing together all
	// shard reports which have finished.
	AggregateReportResult common.ShardScanStats

	// StatusStatus is the type which indicates the status of a shard scan.
	ShardStatus string

	// ReportError is a type that is used to send either error or report on a channel.
	// Exactly one of Report and ErrorStr should be non-nil.
	ReportError struct {
		Report *common.ShardScanReport
		ErrorStr *string
	}
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

	activityOptions := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:   	time.Minute,
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOptions)
	var resolvedConfig ResolvedScannerWorkflowConfig
	if err := workflow.ExecuteActivity(activityCtx, ScannerConfigActivityName, ScannerConfigActivityParams{
		Overwrites: params.ScannerWorkflowConfigOverwrites,
	}).Get(ctx, &resolvedConfig); err != nil {
		return err
	}

	if !resolvedConfig.Enabled {
		return nil
	}

	shardReportChan := workflow.GetSignalChannel(ctx, shardReportChan)
	for i := 0; i < resolvedConfig.Concurrency; i++ {
		idx := i
		workflow.Go(ctx, func(ctx workflow.Context) {
			for _, shard := range shards {
				if shard % resolvedConfig.Concurrency == idx {
					activityOptions.StartToCloseTimeout = time.Hour * 3
					activityCtx = workflow.WithActivityOptions(ctx, activityOptions)
					var report *common.ShardScanReport
					if err := workflow.ExecuteActivity(activityCtx, ScannerScanShardActivityName, ScanShardActivityParams{
						ShardID: shard,
						ExecutionsPageSize: resolvedConfig.ExecutionsPageSize,
						BlobstoreFlushThreshold: resolvedConfig.BlobstoreFlushThreshold,
						InvariantCollections: resolvedConfig.InvariantCollections,
					}).Get(ctx, &report); err != nil {
						errStr := err.Error()
						shardReportChan.Send(ctx, ReportError{
							Report: nil,
							ErrorStr: &errStr,
						})
						return
					}
					shardReportChan.Send(ctx, ReportError{
						Report: report,
						ErrorStr: nil,
					})
				}
			}
		})
	}

	for i := 0; i < len(shards); i++ {
		var reportErr ReportError
		shardReportChan.Receive(ctx, &reportErr)
		if reportErr.ErrorStr != nil {
			return errors.New(*reportErr.ErrorStr)
		}
		aggregator.addReport(*reportErr.Report)
	}

	activityOptions.StartToCloseTimeout = time.Minute
	activityCtx = workflow.WithActivityOptions(ctx, activityOptions)
	if err := workflow.ExecuteActivity(activityCtx, ScannerEmitMetricsActivityName, ScannerEmitMetricsActivityParams{
		ShardStatusResult: aggregator.status,
		AggregateReportResult: aggregator.aggregation,
	}).Get(ctx, nil); err != nil {
		return err
	}
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