package executions

import (
	"context"
	"time"

	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/service/worker/scanner/executions/common"
	"github.com/uber/cadence/service/worker/scanner/executions/shard"
)

const (
	ScannerConfigActivityName    = "cadence-sys-executions-scanner-config-activity"
	ScannerScanShardActivityName = "cadence-sys-executions-scanner-scan-shard-activity"
	ScannerEmitMetricsActivityName = "cadence-sys-executions-scanner-emit-metrics-activity"
)

type (
	scannerConfigActivityParams struct {
		overwrites *ScannerWorkflowConfigOverwrites
	}

	scanShardActivityParams struct {
		shardID int
		executionsPageSize int
		blobstoreFlushThreshold int
		invariantCollections []common.InvariantCollection
	}

	scannerEmitMetricsActivityParams struct {
		latency time.Duration
		shardStatusResult ShardStatusResult
		aggregateReportResult AggregateReportResult
	}
)

// ScannerEmitMetricsActivity will emit metrics for a complete run of scanner
func ScannerEmitMetricsActivity(
	activityCtx context.Context,
	params scannerEmitMetricsActivityParams,
) {
	scope := activityCtx.Value(ScannerContextKey).(ScannerContext).Scope.Tagged(metrics.ActivityTypeTag(ScannerEmitMetricsActivityName))
	scope.RecordTimer(metrics.CadenceLatency, params.latency)
	shardSuccess := 0
	shardControlFlowFailure := 0
	for _, v := range params.shardStatusResult {
		switch v {
		case ShardStatusSuccess:
			shardSuccess++
		case ShardStatusControlFlowFailure:
			shardControlFlowFailure++
		}
	}
	scope.UpdateGauge(metrics.CadenceShardSuccessGauge, float64(shardSuccess))
	scope.UpdateGauge(metrics.CadenceShardFailureGauge, float64(shardControlFlowFailure))

	scope.UpdateGauge(metrics.ScannerExecutionsGauge, float64(params.aggregateReportResult.ExecutionsCount))
	scope.UpdateGauge(metrics.ScannerCorruptedGauge, float64(params.aggregateReportResult.CorruptedCount))
	scope.UpdateGauge(metrics.ScannerCheckFailedGauge, float64(params.aggregateReportResult.CheckFailedCount))
	scope.UpdateGauge(metrics.ScannerCorruptedOpenExecutionGauge, float64(params.aggregateReportResult.CorruptedOpenExecutionCount))
	for k, v := range params.aggregateReportResult.CorruptionByType {
		scope.Tagged(metrics.InvariantTypeTag(string(k))).UpdateGauge(metrics.ScannerCorruptionByTypeGauge, float64(v))
	}
}

// ScanShardActivity will scan all executions in a shard and check for invariant violations.
func ScanShardActivity(
	activityCtx context.Context,
	params scanShardActivityParams,
) (*common.ShardScanReport, error) {
	ctx := activityCtx.Value(ScannerContextKey).(ScannerContext)
	resources := ctx.Resource
	scope := ctx.Scope.Tagged(metrics.ActivityTypeTag(ScannerScanShardActivityName))
	sw := scope.StartTimer(metrics.CadenceLatency)
	defer sw.Stop()
	execManager, err := resources.GetExecutionManager(params.shardID)
	if err != nil {
		scope.IncCounter(metrics.CadenceFailures)
		return nil, err
	}
	pr := common.NewPersistenceRetryer(execManager, resources.GetHistoryManager())
	scanner := shard.NewScanner(
		params.shardID,
		pr,
		params.executionsPageSize,
		resources.GetBlobstoreClient(),
		params.blobstoreFlushThreshold,
		params.invariantCollections)
	report := scanner.Scan()
	if report.Result.ControlFlowFailure != nil {
		scope.IncCounter(metrics.CadenceFailures)
	}
	return &report, nil
}

// ScannerConfigActivity will read dynamic config, apply overwrites and return a resolved config.
func ScannerConfigActivity(
	activityCtx context.Context,
	params scannerConfigActivityParams,
) (ResolvedScannerWorkflowConfig, error) {
	dc := activityCtx.Value(ScannerContextKey).(ScannerContext).ScannerWorkflowDynamicConfig
	overwrites := params.overwrites
	result := ResolvedScannerWorkflowConfig{
		Enabled: dc.Enabled(),
		Concurrency: dc.Concurrency(),
		ExecutionsPageSize: dc.ExecutionsPageSize(),
		BlobstoreFlushThreshold: dc.BlobstoreFlushThreshold(),
		InvariantCollections: InvariantCollections{
			InvariantCollectionMutableState: dc.DynamicConfigInvariantCollections.InvariantCollectionMutableState(),
			InvariantCollectionHistory: dc.DynamicConfigInvariantCollections.InvariantCollectionHistory(),
		},
	}
	if overwrites.Enabled != nil {
		result.Enabled = *overwrites.Enabled
	}
	if overwrites.Concurrency != nil {
		result.Concurrency = *overwrites.Concurrency
	}
	if overwrites.ExecutionsPageSize != nil {
		result.ExecutionsPageSize = *overwrites.ExecutionsPageSize
	}
	if overwrites.BlobstoreFlushThreshold != nil {
		result.BlobstoreFlushThreshold = *overwrites.BlobstoreFlushThreshold
	}
	if overwrites.InvariantCollections != nil {
		result.InvariantCollections = *overwrites.InvariantCollections
	}
	return result, nil
}