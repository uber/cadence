package executions

import "github.com/uber/cadence/common/service/dynamicconfig"

type (
	// ScannerWorkflowDynamicConfig is the dynamic config for scanner workflow
	ScannerWorkflowDynamicConfig struct {
		Enabled dynamicconfig.BoolPropertyFn
		Concurrency dynamicconfig.IntPropertyFn
		ExecutionsPageSize dynamicconfig.IntPropertyFn
		BlobstoreFlushThreshold dynamicconfig.IntPropertyFn
		DynamicConfigInvariantCollections DynamicConfigInvariantCollections
	}

	// DynamicConfigInvariantCollections is the portion of ScannerWorkflowDynamicConfig
	// which indicates which collections of invariants should be run
	DynamicConfigInvariantCollections struct {
		InvariantCollectionMutableState dynamicconfig.BoolPropertyFn
		InvariantCollectionHistory dynamicconfig.BoolPropertyFn
	}

	// ScannerWorkflowOverwrites enables overwriting the values in dynamic config.
	// If provided workflow will favor overwrites over dynamic config.
	ScannerWorkflowConfigOverwrites struct {
		Enabled *bool
		Concurrency *int
		ExecutionsPageSize *int
		BlobstoreFlushThreshold *int
		InvariantCollections *InvariantCollections
	}

	// ResolvedScannerWorkflowConfig is the resolved config after reading dynamic config
	// and applying overwrites.
	ResolvedScannerWorkflowConfig struct {
		Enabled bool
		Concurrency int
		ExecutionsPageSize int
		BlobstoreFlushThreshold int
		InvariantCollections InvariantCollections
	}

	// InvariantCollection represents the resolved set of invariant collections
	// that scanner workflow should run
	InvariantCollections struct {
		InvariantCollectionMutableState bool
		InvariantCollectionHistory bool
	}
)

// GetScannerWorkflowDynamicConfig returns the dynamic config for ScannerWorkflow
func GetScannerWorkflowDynamicConfig() *ScannerWorkflowDynamicConfig {
	return &ScannerWorkflowDynamicConfig{
		// TODO: populate this with some sane defaults
	}
}