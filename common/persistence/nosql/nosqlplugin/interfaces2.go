package nosqlplugin

import (
	"context"
	"time"

	"github.com/uber/cadence/common/persistence"
)

type (
	// visibilityCRUD is for visibility storage
	visibilityCRUD interface {
		InsertVisibility(ctx context.Context, row *persistence.InternalShardInfo) error

	}

	VisibilityRow struct {
		TTL int64
		DomainUUID string
		WorkflowID string
		RunID              string
		WorkflowTypeName   string
		StartTimestamp     time.Time
		ExecutionTimestamp time.Time
		WorkflowTimeout    time.Duration
		TaskID             int64
		Memo               *DataBlob
		TaskList           string
		IsCron             bool
	}
)
