// Copyright (c) 2021 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package nosqlplugin

import (
	"context"
	"time"

	"github.com/uber/cadence/common/persistence"
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
		SortType     VisibilitySortType
		WorkflowType string
		WorkflowID   string
		CloseStatus  int32
	}

	VisibilityFilterType int
	VisibilitySortType   int
)

const (
	// Sorted by StartTime
	AllOpen VisibilityFilterType = iota
	AllClosed
	OpenByWorkflowType
	ClosedByWorkflowType
	OpenByWorkflowID
	ClosedByWorkflowID
	ClosedByClosedStatus
)

const (
	SortByStartTime VisibilitySortType = iota
	SortByClosedTime
)
