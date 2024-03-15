package testdata

import (
	"github.com/uber/cadence/common/types"
)

const (
	DynamicConfigValueName         = "test_config"
	DynamicConfigFilterName        = "test_filter"
	DynamicConfigEntryName         = "test_entry"
	DynamicConfigBlobSchemaVersion = int64(1)
)

var (
	DynamicConfigFilter = types.DynamicConfigFilter{
		Name:  DynamicConfigFilterName,
		Value: &DataBlob,
	}
	DynamicConfigValue = types.DynamicConfigValue{
		Value:   &DataBlob,
		Filters: []*types.DynamicConfigFilter{&DynamicConfigFilter},
	}
	DynamicConfigEntry = types.DynamicConfigEntry{
		Name:   DynamicConfigEntryName,
		Values: []*types.DynamicConfigValue{&DynamicConfigValue},
	}
	DynamicConfigBlob = types.DynamicConfigBlob{
		SchemaVersion: DynamicConfigBlobSchemaVersion,
		Entries:       []*types.DynamicConfigEntry{&DynamicConfigEntry},
	}
)
