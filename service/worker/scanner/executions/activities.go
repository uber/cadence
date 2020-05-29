package executions

import "context"

// ScannerConfigActivity will read dynamic config, apply overwrites and return a resolved config.
func ScannerConfigActivity(
	activityCtx context.Context,
	overwrites *ScannerWorkflowConfigOverwrites,
) (ResolvedScannerWorkflowConfig, error) {
	dc := activityCtx.Value(ScannerContextKey).(ScannerContext).ScannerWorkflowDynamicConfig
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