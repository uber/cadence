// Copyright (c) 2017 Uber Technologies, Inc.
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

package cassandra

import (
	"context"
	"fmt"
	"time"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra"
	"github.com/uber/cadence/common/types"
)

// Fixed domain values for now
const (
	defaultCloseTTLSeconds = 86400
	openExecutionTTLBuffer = int64(86400) // setting it to a day to account for shard going down

	maxCassandraTTL = int64(157680000) // Cassandra max support time is 2038-01-19T03:14:06+00:00. Updated this to 5 years to support until year 2033
)

type (
	nosqlVisibilityManager struct {
		sortByCloseTime bool
		db              nosqlplugin.DB
	}
)

// newVisibilityPersistence is used to create an instance of VisibilityManager implementation
func newVisibilityPersistence(
	listClosedOrderingByCloseTime bool,
	cfg config.Cassandra,
	logger log.Logger,
) (p.VisibilityStore, error) {
	// TODO hardcoding to Cassandra for now, will switch to dynamically loading later
	db, err := cassandra.NewCassandraDB(cfg, logger)
	if err != nil {
		return nil, err
	}

	return &nosqlVisibilityManager{
		sortByCloseTime: listClosedOrderingByCloseTime,
		db:              db,
	}, nil
}

func (v *nosqlVisibilityManager) GetName() string {
	return v.db.PluginName()
}

// Close releases the underlying resources held by this object
func (v *nosqlVisibilityManager) Close() {
	v.db.Close()
}

func (v *nosqlVisibilityManager) RecordWorkflowExecutionStarted(
	ctx context.Context,
	request *p.InternalRecordWorkflowExecutionStartedRequest,
) error {
	ttl := int64(request.WorkflowTimeout.Seconds()) + openExecutionTTLBuffer

	err := v.db.InsertVisibility(ctx, ttl, &nosqlplugin.VisibilityRowForInsert{
		DomainID: request.DomainUUID,
		VisibilityRow: nosqlplugin.VisibilityRow{
			WorkflowID:    request.WorkflowID,
			RunID:         request.RunID,
			TypeName:      request.WorkflowTypeName,
			StartTime:     request.StartTimestamp,
			ExecutionTime: request.ExecutionTimestamp,
			Memo:          request.Memo,
			TaskList:      request.TaskList,
			IsCron:        request.IsCron,
		},
	})
	if err != nil {
		return convertCommonErrors(v.db, "RecordWorkflowExecutionStarted", err)
	}

	return nil
}

func (v *nosqlVisibilityManager) RecordWorkflowExecutionClosed(
	ctx context.Context,
	request *p.InternalRecordWorkflowExecutionClosedRequest,
) error {
	// Find how long to keep the row
	retention := request.RetentionSeconds
	if retention == 0 {
		retention = defaultCloseTTLSeconds * time.Second
	}

	err := v.db.UpdateVisibility(ctx, int64(retention.Seconds()), &nosqlplugin.VisibilityRowForUpdate{
		DomainID:          request.DomainUUID,
		UpdateOpenToClose: true,
		VisibilityRow: nosqlplugin.VisibilityRow{
			WorkflowID:    request.WorkflowID,
			RunID:         request.RunID,
			TypeName:      request.WorkflowTypeName,
			StartTime:     request.StartTimestamp,
			ExecutionTime: request.ExecutionTimestamp,
			Memo:          request.Memo,
			TaskList:      request.TaskList,
			IsCron:        request.IsCron,
			//closed workflow attributes
			Status:        &request.Status,
			CloseTime:     request.CloseTimestamp,
			HistoryLength: request.HistoryLength,
		},
	})

	if err != nil {
		return convertCommonErrors(v.db, "RecordWorkflowExecutionClosed", err)
	}
	return nil
}

func (v *nosqlVisibilityManager) UpsertWorkflowExecution(
	ctx context.Context,
	request *p.InternalUpsertWorkflowExecutionRequest,
) error {
	if p.IsNopUpsertWorkflowRequest(request) {
		return nil
	}
	return p.NewOperationNotSupportErrorForVis()
}

func (v *nosqlVisibilityManager) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	resp, err := v.db.SelectVisibility(ctx, &nosqlplugin.VisibilityFilter{
		ListRequest: *request,
		FilterType:  nosqlplugin.AllOpen,
		SortType:    nosqlplugin.SortByStartTime,
	})
	if err != nil {
		return nil, convertCommonErrors(v.db, "ListOpenWorkflowExecutions", err)
	}

	return &p.InternalListWorkflowExecutionsResponse{
		Executions:    resp.Executions,
		NextPageToken: resp.NextPageToken,
	}, nil
}

func (v *nosqlVisibilityManager) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	var filter *nosqlplugin.VisibilityFilter
	if v.sortByCloseTime {
		filter = &nosqlplugin.VisibilityFilter{
			ListRequest: *request,
			FilterType:  nosqlplugin.AllClosed,
			SortType:    nosqlplugin.SortByClosedTime,
		}
	} else {
		filter = &nosqlplugin.VisibilityFilter{
			ListRequest: *request,
			FilterType:  nosqlplugin.AllClosed,
			SortType:    nosqlplugin.SortByStartTime,
		}
	}
	resp, err := v.db.SelectVisibility(ctx, filter)
	if err != nil {
		return nil, convertCommonErrors(v.db, "ListClosedWorkflowExecutions", err)
	}

	return &p.InternalListWorkflowExecutionsResponse{
		Executions:    resp.Executions,
		NextPageToken: resp.NextPageToken,
	}, nil
}

func (v *nosqlVisibilityManager) ListOpenWorkflowExecutionsByType(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByTypeRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	resp, err := v.db.SelectVisibility(ctx, &nosqlplugin.VisibilityFilter{
		ListRequest:  request.InternalListWorkflowExecutionsRequest,
		FilterType:   nosqlplugin.OpenByWorkflowType,
		SortType:     nosqlplugin.SortByStartTime,
		WorkflowType: request.WorkflowTypeName,
	})
	if err != nil {
		return nil, convertCommonErrors(v.db, "ListOpenWorkflowExecutionsByType", err)
	}

	return &p.InternalListWorkflowExecutionsResponse{
		Executions:    resp.Executions,
		NextPageToken: resp.NextPageToken,
	}, nil
}

func (v *nosqlVisibilityManager) ListClosedWorkflowExecutionsByType(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByTypeRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	var filter *nosqlplugin.VisibilityFilter
	if v.sortByCloseTime {
		filter = &nosqlplugin.VisibilityFilter{
			ListRequest:  request.InternalListWorkflowExecutionsRequest,
			FilterType:   nosqlplugin.ClosedByWorkflowType,
			SortType:     nosqlplugin.SortByClosedTime,
			WorkflowType: request.WorkflowTypeName,
		}
	} else {
		filter = &nosqlplugin.VisibilityFilter{
			ListRequest:  request.InternalListWorkflowExecutionsRequest,
			FilterType:   nosqlplugin.ClosedByWorkflowType,
			SortType:     nosqlplugin.SortByStartTime,
			WorkflowType: request.WorkflowTypeName,
		}
	}
	resp, err := v.db.SelectVisibility(ctx, filter)
	if err != nil {
		return nil, convertCommonErrors(v.db, "ListClosedWorkflowExecutionsByType", err)
	}

	return &p.InternalListWorkflowExecutionsResponse{
		Executions:    resp.Executions,
		NextPageToken: resp.NextPageToken,
	}, nil
}

func (v *nosqlVisibilityManager) ListOpenWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByWorkflowIDRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	resp, err := v.db.SelectVisibility(ctx, &nosqlplugin.VisibilityFilter{
		ListRequest: request.InternalListWorkflowExecutionsRequest,
		FilterType:  nosqlplugin.OpenByWorkflowID,
		SortType:    nosqlplugin.SortByStartTime,
		WorkflowID:  request.WorkflowID,
	})
	if err != nil {
		return nil, convertCommonErrors(v.db, "ListOpenWorkflowExecutionsByWorkflowID", err)
	}

	return &p.InternalListWorkflowExecutionsResponse{
		Executions:    resp.Executions,
		NextPageToken: resp.NextPageToken,
	}, nil
}

func (v *nosqlVisibilityManager) ListClosedWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByWorkflowIDRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	var filter *nosqlplugin.VisibilityFilter
	if v.sortByCloseTime {
		filter = &nosqlplugin.VisibilityFilter{
			ListRequest: request.InternalListWorkflowExecutionsRequest,
			FilterType:  nosqlplugin.ClosedByWorkflowID,
			SortType:    nosqlplugin.SortByClosedTime,
			WorkflowID:  request.WorkflowID,
		}
	} else {
		filter = &nosqlplugin.VisibilityFilter{
			ListRequest: request.InternalListWorkflowExecutionsRequest,
			FilterType:  nosqlplugin.ClosedByWorkflowID,
			SortType:    nosqlplugin.SortByStartTime,
			WorkflowID:  request.WorkflowID,
		}
	}
	resp, err := v.db.SelectVisibility(ctx, filter)
	if err != nil {
		return nil, convertCommonErrors(v.db, "ListClosedWorkflowExecutionsByWorkflowID", err)
	}

	return &p.InternalListWorkflowExecutionsResponse{
		Executions:    resp.Executions,
		NextPageToken: resp.NextPageToken,
	}, nil
}

func (v *nosqlVisibilityManager) ListClosedWorkflowExecutionsByStatus(
	ctx context.Context,
	request *p.InternalListClosedWorkflowExecutionsByStatusRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	var filter *nosqlplugin.VisibilityFilter
	if v.sortByCloseTime {
		filter = &nosqlplugin.VisibilityFilter{
			ListRequest: request.InternalListWorkflowExecutionsRequest,
			FilterType:  nosqlplugin.ClosedByClosedStatus,
			SortType:    nosqlplugin.SortByClosedTime,
			CloseStatus: int32(request.Status),
		}
	} else {
		filter = &nosqlplugin.VisibilityFilter{
			ListRequest: request.InternalListWorkflowExecutionsRequest,
			FilterType:  nosqlplugin.ClosedByClosedStatus,
			SortType:    nosqlplugin.SortByStartTime,
			CloseStatus: int32(request.Status),
		}
	}
	resp, err := v.db.SelectVisibility(ctx, filter)
	if err != nil {
		return nil, convertCommonErrors(v.db, "ListClosedWorkflowExecutionsByStatus", err)
	}

	return &p.InternalListWorkflowExecutionsResponse{
		Executions:    resp.Executions,
		NextPageToken: resp.NextPageToken,
	}, nil
}

func (v *nosqlVisibilityManager) GetClosedWorkflowExecution(
	ctx context.Context,
	request *p.InternalGetClosedWorkflowExecutionRequest,
) (*p.InternalGetClosedWorkflowExecutionResponse, error) {
	wfexecution, err := v.db.SelectOneClosedWorkflow(ctx, request.DomainUUID, request.Execution.GetWorkflowID(), request.Execution.GetRunID())

	if err != nil {
		return nil, convertCommonErrors(v.db, "GetClosedWorkflowExecution", err)
	}
	if wfexecution == nil {
		// Special case: this API return nil,nil if not found(since we will deprecate it, it's not worth refactor to be consistent)
		return nil, &types.EntityNotExistsError{
			Message: fmt.Sprintf("Workflow execution not found.  WorkflowId: %v, RunId: %v",
				request.Execution.GetWorkflowID(), request.Execution.GetRunID()),
		}
	}
	return &p.InternalGetClosedWorkflowExecutionResponse{
		Execution: wfexecution,
	}, nil
}

func (v *nosqlVisibilityManager) DeleteWorkflowExecution(
	ctx context.Context,
	request *p.VisibilityDeleteWorkflowExecutionRequest,
) error {
	err := v.db.DeleteVisibility(ctx, request.DomainID, request.WorkflowID, request.RunID)
	if err != nil {
		return convertCommonErrors(v.db, "DeleteWorkflowExecution", err)
	}
	return nil
}

func (v *nosqlVisibilityManager) ListWorkflowExecutions(
	ctx context.Context,
	request *p.ListWorkflowExecutionsByQueryRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return nil, p.NewOperationNotSupportErrorForVis()
}

func (v *nosqlVisibilityManager) ScanWorkflowExecutions(
	ctx context.Context,
	request *p.ListWorkflowExecutionsByQueryRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	return nil, p.NewOperationNotSupportErrorForVis()
}

func (v *nosqlVisibilityManager) CountWorkflowExecutions(
	ctx context.Context,
	request *p.CountWorkflowExecutionsRequest,
) (*p.CountWorkflowExecutionsResponse, error) {
	return nil, p.NewOperationNotSupportErrorForVis()
}
