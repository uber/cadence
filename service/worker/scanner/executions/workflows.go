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
	"time"

	"go.uber.org/cadence"

	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/service/worker/scanner/executions/common"
)

const (
	// ScannerContextKey is the key used to access ScannerContext in activities
	ScannerContextKey = ContextKey(0)
	// FixerContextKey is the key used to access FixerContext in activities
	FixerContextKey = ContextKey(1)

	// ShardReportQuery is the query name for the query used to get a single shard's report
	ShardReportQuery = "shard_report"
	// ShardStatusQuery is the query name for the query used to get the status of all shards
	ShardStatusQuery = "shard_status_query"
	// AggregateReportQuery is the query name for the query used to get the aggregate result of all finished shards
	AggregateReportQuery = "aggregate_report"
	// ShardCorruptKeysQuery is the query name for the query used to get all completed shards with at least one corruption
	ShardCorruptKeysQuery = "shard_corrupt_keys"

	// ShardStatusRunning indicates the shard has not completed yet
	ShardStatusRunning ShardStatus = "running"
	// ShardStatusSuccess indicates the scan on the shard ran successfully
	ShardStatusSuccess ShardStatus = "success"
	// ShardStatusControlFlowFailure indicates the scan on the shard failed
	ShardStatusControlFlowFailure ShardStatus = "control_flow_failure"

	scanShardReportChan = "scanShardReportChan"
	fixShardReportChan = "fixShardReportChan"
)

type (
	// ContextKey is the type which identifies context keys
	ContextKey int

	// ScannerContext is the resource that is available in activities under ScannerContextKey context key
	ScannerContext struct {
		Resource                     resource.Resource
		Scope                        metrics.Scope
		ScannerWorkflowDynamicConfig *ScannerWorkflowDynamicConfig
	}

	// FixerContext is the resource that is available to activities under FixerContextKey
	FixerContext struct {
		Resource resource.Resource
		Scope metrics.Scope
	}

	// ScannerWorkflowParams are the parameters to the scan workflow
	ScannerWorkflowParams struct {
		Shards                          Shards
		ScannerWorkflowConfigOverwrites ScannerWorkflowConfigOverwrites
	}

	// FixerWorkflowParams are the parameters to the fix workflow
	FixerWorkflowParams struct {
		FixerCorruptedKeysActivityParams FixerCorruptedKeysActivityParams
		FixerWorkflowConfigOverwrites FixerWorkflowConfigOverwrites
	}

	// Shards identify the shards that should be scanned.
	// Exactly one of List or Range should be non-nil.
	Shards struct {
		List  []int
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

	// ShardCorruptKeysResult is a map of all shards which have finished scan successfully and have at least one corruption
	ShardCorruptKeysResult map[int]common.Keys

	// ShardStatus is the type which indicates the status of a shard scan.
	ShardStatus string

	// ScanReportError is a type that is used to send either error or report on a channel.
	// Exactly one of Report and ErrorStr should be non-nil.
	ScanReportError struct {
		Report   *common.ShardScanReport
		ErrorStr *string
	}

	// FixReportError is a type that is used to send either error or report on a channel.
	// Exactly one of Report and ErrorStr should be non-nil.
	FixReportError struct {
		Report   *common.ShardFixReport
		ErrorStr *string
	}
)

// ScannerWorkflow is the workflow that scans over all concrete executions
func ScannerWorkflow(
	ctx workflow.Context,
	params ScannerWorkflowParams,
) error {
	shards := flattenShards(params.Shards)
	aggregator := newShardScanResultAggregator(shards)
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
	if err := workflow.SetQueryHandler(ctx, ShardCorruptKeysQuery, func() (ShardCorruptKeysResult, error) {
		return aggregator.corruptionKeys, nil
	}); err != nil {
		return err
	}

	activityOptions := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 1.7,
			ExpirationInterval: 5 * time.Minute,
		},
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

	shardReportChan := workflow.GetSignalChannel(ctx, scanShardReportChan)
	for i := 0; i < resolvedConfig.Concurrency; i++ {
		idx := i
		workflow.Go(ctx, func(ctx workflow.Context) {
			for _, shard := range shards {
				if shard%resolvedConfig.Concurrency == idx {
					activityOptions.RetryPolicy.ExpirationInterval = time.Hour * 5
					activityOptions.HeartbeatTimeout = time.Minute
					activityCtx = workflow.WithActivityOptions(ctx, activityOptions)
					var report *common.ShardScanReport
					if err := workflow.ExecuteActivity(activityCtx, ScannerScanShardActivityName, ScanShardActivityParams{
						ShardID:                 shard,
						ExecutionsPageSize:      resolvedConfig.ExecutionsPageSize,
						BlobstoreFlushThreshold: resolvedConfig.BlobstoreFlushThreshold,
						InvariantCollections:    resolvedConfig.InvariantCollections,
					}).Get(ctx, &report); err != nil {
						errStr := err.Error()
						shardReportChan.Send(ctx, ScanReportError{
							Report:   nil,
							ErrorStr: &errStr,
						})
						return
					}
					shardReportChan.Send(ctx, ScanReportError{
						Report:   report,
						ErrorStr: nil,
					})
				}
			}
		})
	}

	for i := 0; i < len(shards); i++ {
		var reportErr ScanReportError
		shardReportChan.Receive(ctx, &reportErr)
		if reportErr.ErrorStr != nil {
			return errors.New(*reportErr.ErrorStr)
		}
		aggregator.addReport(*reportErr.Report)
	}

	activityOptions.RetryPolicy.ExpirationInterval = time.Minute * 5
	activityCtx = workflow.WithActivityOptions(ctx, activityOptions)
	if err := workflow.ExecuteActivity(activityCtx, ScannerEmitMetricsActivityName, ScannerEmitMetricsActivityParams{
		ShardStatusResult:     aggregator.status,
		AggregateReportResult: aggregator.aggregation,
	}).Get(ctx, nil); err != nil {
		return err
	}
	return nil
}

type shardScanResultAggregator struct {
	reports        map[int]common.ShardScanReport
	status         ShardStatusResult
	aggregation    AggregateReportResult
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
		aggregation: AggregateReportResult{
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
	if report.Result.ControlFlowFailure == nil {
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
	if report.Result.ControlFlowFailure == nil {
		a.adjustAggregation(report.Stats, func(a, b int64) int64 { return a - b })
	}
}

func (a *shardScanResultAggregator) getReport(shardID int) (*common.ShardScanReport, error) {
	if report, ok := a.reports[shardID]; ok {
		return &report, nil
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

func FixerWorkflow(
	ctx workflow.Context,
	params FixerWorkflowParams,
) error {
	// 1. get the corrupted keys from scanner workflow
	// 2. resolve the config using overwrites and sane defaults
	// 3. start concurrency number of go routines and have them consume


	activityOptions := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 1.7,
			ExpirationInterval: 5 * time.Minute,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOptions)
	var corruptedKeys []CorruptedKeysEntry
	if err := workflow.ExecuteActivity(
		activityCtx, FixerCorruptedKeysActivityName, params.FixerCorruptedKeysActivityParams).Get(ctx, &corruptedKeys); err != nil {
		return err
	}

	resolvedConfig := resolveFixerConfig(params.FixerWorkflowConfigOverwrites)



	shardReportChan := workflow.GetSignalChannel(ctx, fixShardReportChan)
	for i := 0; i < resolvedConfig.Concurrency; i++ {
		idx := i
		workflow.Go(ctx, func(ctx workflow.Context) {
			for j, corrupted := range corruptedKeys {
				if j%resolvedConfig.Concurrency == idx {
					activityOptions.RetryPolicy.ExpirationInterval = time.Hour * 5
					activityOptions.HeartbeatTimeout = time.Minute
					activityCtx = workflow.WithActivityOptions(ctx, activityOptions)
					var report *common.ShardFixReport
					if err := workflow.ExecuteActivity(activityCtx, FixerFixShardActivityName, FixShardActivityParams{
						CorruptedKeysEntry: corrupted,
						ResolvedFixerWorkflowConfig: resolvedConfig,
					}).Get(ctx, &report); err != nil {
						errStr := err.Error()
						shardReportChan.Send(ctx, FixReportError{
							Report:   nil,
							ErrorStr: &errStr,
						})
						return
					}
					shardReportChan.Send(ctx, FixReportError{
						Report:   report,
						ErrorStr: nil,
					})
				}
			}
		})
	}

	for i := 0; i < len(corruptedKeys); i++ {
		var reportErr FixReportError
		shardReportChan.Receive(ctx, &reportErr)
		if reportErr.ErrorStr != nil {
			return errors.New(*reportErr.ErrorStr)
		}
		aggregator.addReport(*reportErr.Report)
	}


	return nil
}

func resolveFixerConfig(overwrites FixerWorkflowConfigOverwrites) ResolvedFixerWorkflowConfig {
	resolvedConfig := ResolvedFixerWorkflowConfig{
		Concurrency: 25,
		BlobstoreFlushThreshold: 1000,
		InvariantCollections: InvariantCollections{
			InvariantCollectionMutableState: true,
			InvariantCollectionHistory: true,
		},
	}
	if overwrites.Concurrency != nil {
		resolvedConfig.Concurrency = *overwrites.Concurrency
	}
	if overwrites.BlobstoreFlushThreshold != nil {
		resolvedConfig.BlobstoreFlushThreshold = *overwrites.BlobstoreFlushThreshold
	}
	if overwrites.InvariantCollections != nil {
		resolvedConfig.InvariantCollections = *overwrites.InvariantCollections
	}
	return resolvedConfig
}
