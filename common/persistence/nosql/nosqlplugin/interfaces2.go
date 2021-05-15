package nosqlplugin

import (
	"context"
	"github.com/uber/cadence/common/persistence"
	"time"
)

type (
	// visibilityCRUD is for visibility storage
	// NOTE: Cassandra implementation of visibility uses three tables: open_executions, closed_executions and closed_executions_v2.
	// Records in open_executions and closed_executions are clustered by start_time. Records in  closed_executions_v2 are by close_time.
	// This is for optimizing the performance, but introduce a lot of complexity.
	// In some other databases, this may be be necessary. Please refer to MySQL/Postgres implementation which uses only
	// one table with multiple indexes.
	// TTL(time to live records) is for auto-deleting expired records in visibility. For databases that don't support TTL,
	// please implement DeleteVisibility method. If TTL is supported, then DeleteVisibility can be a noop.
	visibilityCRUD interface {
		InsertVisibility(ctx context.Context, ttlSeconds int64, row *VisibilityRowForWrite) error
		UpdateVisibility(ctx context.Context, ttlSeconds int64, row *VisibilityRowForWrite) error
		SelectVisibility(ctx context.Context, filter *VisibilityFilter) (*SelectVisibilityResponse, error)
		DeleteVisibility(ctx context.Context, domainID, workflowID, runID string) error
		// TODO deprecated this in the future in favor of SelectVisibility
		SelectOneClosedWorkflow(ctx context.Context, domainID, workflowID, runID string) (*VisibilityRowForRead, error)
	}

	VisibilityRowForWrite struct {
		DomainID         string
		RunID            string
		WorkflowTypeName string
		WorkflowID       string
		StartTime        time.Time
		ExecutionTime    time.Time
		CloseStatus      int32
		CloseTime        time.Time
		HistoryLength    int64
		Memo             []byte
		Encoding         string
		TaskList         string
		IsCron           bool
	}

	// TODO unify with VisibilityRowForWrite in the future for consistency
	VisibilityRowForRead = persistence.InternalVisibilityWorkflowExecutionInfo

	SelectVisibilityResponse struct {
		Executions    []*VisibilityRowForRead
		NextPageToken []byte
	}

	// VisibilityFilter contains the column names within executions_visibility table that
	// can be used to filter results through a WHERE clause
	VisibilityFilter struct {
		ListRequest  persistence.InternalListWorkflowExecutionsRequest
		FilterType   VisibilityFilterType
		WorkflowType string
		WorkflowID   string
		CloseStatus  int32
	}

	VisibilityFilterType int
)

const (
	// Sorted by StartTime
	OpenSortedByStartTime VisibilityFilterType = iota
	ClosedSortedByStartTime
	OpenFilteredByWorkflowTypeSortedByStartTime
	ClosedFilteredByWorkflowTypeSortedByStartTime
	OpenFilteredByWorkflowIDSortedByStartTime
	ClosedFilteredByWorkflowIDSortedByStartTime
	ClosedFilteredByClosedStatusSortedByStartTime

	// Sorted by ClosedTime only valid when workflows are closed
	ClosedSortedByClosedTime
	ClosedFilteredByWorkflowTypeSortedByClosedTime
	ClosedFilteredByWorkflowIDSortedByClosedTime
	ClosedFilteredByClosedStatusSortedByClosedTIme
)
