// Copyright (c) 2020 Uber Technologies, Inc.
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

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"
)

const (
	domainPartition = 0
)

const (
	///////////////// Open Executions /////////////////
	openExecutionsColumnsForSelect = " workflow_id, run_id, start_time, execution_time, workflow_type_name, memo, encoding, task_list, is_cron "

	openExecutionsColumnsForInsert = "(domain_id, domain_partition, " + openExecutionsColumnsForSelect + ")"

	templateCreateWorkflowExecutionStartedWithTTL = `INSERT INTO open_executions ` +
		openExecutionsColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) using TTL ?`

	templateCreateWorkflowExecutionStarted = `INSERT INTO open_executions` +
		openExecutionsColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	templateDeleteWorkflowExecutionStarted = `DELETE FROM open_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND start_time = ? ` +
		`AND run_id = ?`

	templateGetOpenWorkflowExecutions = `SELECT ` + openExecutionsColumnsForSelect +
		`FROM open_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition IN (?) ` +
		`AND start_time >= ? ` +
		`AND start_time <= ? `

	templateGetOpenWorkflowExecutionsByType = `SELECT ` + openExecutionsColumnsForSelect +
		`FROM open_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND start_time >= ? ` +
		`AND start_time <= ? ` +
		`AND workflow_type_name = ? `

	templateGetOpenWorkflowExecutionsByID = `SELECT ` + openExecutionsColumnsForSelect +
		`FROM open_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND start_time >= ? ` +
		`AND start_time <= ? ` +
		`AND workflow_id = ? `

	///////////////// Closed Executions /////////////////
	closedExecutionColumnsForSelect = " workflow_id, run_id, start_time, execution_time, close_time, workflow_type_name, status, history_length, memo, encoding, task_list, is_cron "

	closedExecutionColumnsForInsert = "(domain_id, domain_partition, " + closedExecutionColumnsForSelect + ")"

	templateCreateWorkflowExecutionClosedWithTTL = `INSERT INTO closed_executions ` +
		closedExecutionColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) using TTL ?`

	templateCreateWorkflowExecutionClosed = `INSERT INTO closed_executions ` +
		closedExecutionColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	templateCreateWorkflowExecutionClosedWithTTLV2 = `INSERT INTO closed_executions_v2 ` +
		closedExecutionColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) using TTL ?`

	templateCreateWorkflowExecutionClosedV2 = `INSERT INTO closed_executions_v2 ` +
		closedExecutionColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	templateGetClosedWorkflowExecutions = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition IN (?) ` +
		`AND start_time >= ? ` +
		`AND start_time <= ? `

	templateGetClosedWorkflowExecutionsByType = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND start_time >= ? ` +
		`AND start_time <= ? ` +
		`AND workflow_type_name = ? `

	templateGetClosedWorkflowExecutionsByID = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND start_time >= ? ` +
		`AND start_time <= ? ` +
		`AND workflow_id = ? `

	templateGetClosedWorkflowExecutionsByStatus = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND start_time >= ? ` +
		`AND start_time <= ? ` +
		`AND status = ? `

	templateGetClosedWorkflowExecution = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND workflow_id = ? ` +
		`AND run_id = ? ALLOW FILTERING `

	templateGetClosedWorkflowExecutionsSortByCloseTime = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions_v2 ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition IN (?) ` +
		`AND close_time >= ? ` +
		`AND close_time <= ? `

	templateGetClosedWorkflowExecutionsByTypeSortByCloseTime = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions_v2 ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND close_time >= ? ` +
		`AND close_time <= ? ` +
		`AND workflow_type_name = ? `

	templateGetClosedWorkflowExecutionsByIDSortByCloseTime = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions_v2 ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND close_time >= ? ` +
		`AND close_time <= ? ` +
		`AND workflow_id = ? `

	templateGetClosedWorkflowExecutionsByStatusSortByClosedTime = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions_v2 ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND close_time >= ? ` +
		`AND close_time <= ? ` +
		`AND status = ? `
)

// InsertVisibility creates a new visibility record, return error is there is any.
func (db *cdb) InsertVisibility(ctx context.Context, ttlSeconds int64, row *nosqlplugin.VisibilityRowForWrite) error {
	var query gocql.Query
	if ttlSeconds > maxCassandraTTL {
		query = db.session.Query(templateCreateWorkflowExecutionStarted,
			row.DomainID,
			domainPartition,
			row.WorkflowID,
			row.RunID,
			persistence.UnixNanoToDBTimestamp(row.StartTime.UnixNano()),
			persistence.UnixNanoToDBTimestamp(row.ExecutionTime.UnixNano()),
			row.WorkflowTypeName,
			row.Memo,
			row.Encoding,
			row.TaskList,
			row.IsCron,
		).WithContext(ctx)
	} else {
		query = db.session.Query(templateCreateWorkflowExecutionStartedWithTTL,
			row.DomainID,
			domainPartition,
			row.WorkflowID,
			row.RunID,
			persistence.UnixNanoToDBTimestamp(row.StartTime.UnixNano()),
			persistence.UnixNanoToDBTimestamp(row.ExecutionTime.UnixNano()),
			row.WorkflowTypeName,
			row.Memo,
			row.Encoding,
			row.TaskList,
			row.IsCron,
			ttlSeconds,
		).WithContext(ctx)
	}
	query = query.WithTimestamp(persistence.UnixNanoToDBTimestamp(row.StartTime.UnixNano()))
	return query.Exec()
}

func (db *cdb) UpdateVisibility(ctx context.Context, ttlSeconds int64, row *nosqlplugin.VisibilityRowForWrite) error {
	batch := db.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)

	// First, remove execution from the open table
	batch.Query(templateDeleteWorkflowExecutionStarted,
		row.DomainID,
		domainPartition,
		persistence.UnixNanoToDBTimestamp(row.StartTime.UnixNano()),
		row.RunID,
	)

	// Next, add a row in the closed table.
	if ttlSeconds > maxCassandraTTL {
		batch.Query(templateCreateWorkflowExecutionClosed,
			row.DomainID,
			domainPartition,
			row.WorkflowID,
			row.RunID,
			persistence.UnixNanoToDBTimestamp(row.StartTime.UnixNano()),
			persistence.UnixNanoToDBTimestamp(row.ExecutionTime.UnixNano()),
			persistence.UnixNanoToDBTimestamp(row.CloseTime.UnixNano()),
			row.WorkflowTypeName,
			row.CloseStatus,
			row.HistoryLength,
			row.Memo,
			row.Encoding,
			row.TaskList,
			row.IsCron,
		)
		// duplicate write to v2 to order by close time
		batch.Query(templateCreateWorkflowExecutionClosedV2,
			row.DomainID,
			domainPartition,
			row.WorkflowID,
			row.RunID,
			persistence.UnixNanoToDBTimestamp(row.StartTime.UnixNano()),
			persistence.UnixNanoToDBTimestamp(row.ExecutionTime.UnixNano()),
			persistence.UnixNanoToDBTimestamp(row.CloseTime.UnixNano()),
			row.WorkflowTypeName,
			row.CloseStatus,
			row.HistoryLength,
			row.Memo,
			row.Encoding,
			row.TaskList,
			row.IsCron,
		)
	} else {
		batch.Query(templateCreateWorkflowExecutionClosedWithTTL,
			row.DomainID,
			domainPartition,
			row.WorkflowID,
			row.RunID,
			persistence.UnixNanoToDBTimestamp(row.StartTime.UnixNano()),
			persistence.UnixNanoToDBTimestamp(row.ExecutionTime.UnixNano()),
			persistence.UnixNanoToDBTimestamp(row.CloseTime.UnixNano()),
			row.WorkflowTypeName,
			row.CloseStatus,
			row.HistoryLength,
			row.Memo,
			row.Encoding,
			row.TaskList,
			row.IsCron,
			ttlSeconds,
		)
		// duplicate write to v2 to order by close time
		batch.Query(templateCreateWorkflowExecutionClosedWithTTLV2,
			row.DomainID,
			domainPartition,
			row.WorkflowID,
			row.RunID,
			persistence.UnixNanoToDBTimestamp(row.StartTime.UnixNano()),
			persistence.UnixNanoToDBTimestamp(row.ExecutionTime.UnixNano()),
			persistence.UnixNanoToDBTimestamp(row.CloseTime.UnixNano()),
			row.WorkflowTypeName,
			row.CloseStatus,
			row.HistoryLength,
			row.Memo,
			row.Encoding,
			row.TaskList,
			row.IsCron,
			ttlSeconds,
		)
	}

	// RecordWorkflowExecutionStarted is using StartTimestamp as
	// the timestamp to issue query to Cassandra
	// due to the fact that cross DC using mutable state creation time as workflow start time
	// and visibility using event time instead of last update time (#1501)
	// CloseTimestamp can be before StartTimestamp, meaning using CloseTimestamp
	// can cause the deletion of open visibility record to be ignored.
	queryTimeStamp := row.CloseTime
	if queryTimeStamp.Before(row.StartTime) {
		queryTimeStamp = row.StartTime.Add(time.Second)
	}
	batch = batch.WithTimestamp(persistence.UnixNanoToDBTimestamp(queryTimeStamp.UnixNano()))
	return db.session.ExecuteBatch(batch)
}

func (db *cdb) SelectOneClosedWorkflow(
	ctx context.Context,
	domainID, workflowID, runID string,
) (*nosqlplugin.VisibilityRowForRead, error) {
	query := db.session.Query(templateGetClosedWorkflowExecution,
		domainID,
		domainPartition,
		workflowID,
		runID,
	).WithContext(ctx)

	iter := query.Iter()
	if iter == nil {
		return nil, fmt.Errorf("not able to create query iterator")
	}

	wfexecution, has := readClosedWorkflowExecutionRecord(iter)
	if !has {
		return nil, &types.EntityNotExistsError{
			Message: fmt.Sprintf("Workflow execution not found.  WorkflowId: %v, RunId: %v",
				workflowID, runID),
		}
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return wfexecution, nil
}

// Noop for Cassandra as it already handle by TTL
func (db *cdb) DeleteVisibility(ctx context.Context, domainID, workflowID, runID string) error {
	return nil
}

func (db *cdb) SelectVisibility(ctx context.Context, filter *nosqlplugin.VisibilityFilter) (*nosqlplugin.SelectVisibilityResponse, error) {
	switch filter.FilterType {
	case nosqlplugin.OpenSortedByStartTime:
		return db.openSortedyByStartTime(ctx, &filter.ListRequest)
	case nosqlplugin.ClosedSortedByStartTime:
		return db.closedSortedByStartTime(ctx, &filter.ListRequest)
	case nosqlplugin.ClosedSortedByClosedTime:
		return db.closedSortedByClosedTime(ctx, &filter.ListRequest)

	// workflowType
	case nosqlplugin.OpenFilteredByWorkflowTypeSortedByStartTime:
		return db.openFilteredByWorkflowTypeSortedByStartTime(ctx, &filter.ListRequest, filter.WorkflowType)
	case nosqlplugin.ClosedFilteredByWorkflowTypeSortedByStartTime:
		return db.closedFilteredByWorkflowTypeSortedByStartTime(ctx, &filter.ListRequest, filter.WorkflowType)
	case nosqlplugin.ClosedFilteredByWorkflowTypeSortedByClosedTime:
		return db.closedFilteredByWorkflowTypeSortedByClosedTime(ctx, &filter.ListRequest, filter.WorkflowType)

	// workflowID
	case nosqlplugin.OpenFilteredByWorkflowIDSortedByStartTime:
		return db.openFilteredByWorkflowIDSortedByStartTime(ctx, &filter.ListRequest, filter.WorkflowID)
	case nosqlplugin.ClosedFilteredByWorkflowIDSortedByStartTime:
		return db.closedFilteredByWorkflowIDSortedByStartTime(ctx, &filter.ListRequest, filter.WorkflowID)
	case nosqlplugin.ClosedFilteredByWorkflowIDSortedByClosedTime:
		return db.closedFilteredByWorkflowIDSortedByClosedTime(ctx, &filter.ListRequest, filter.WorkflowID)

	// closeStatus
	case nosqlplugin.ClosedFilteredByClosedStatusSortedByStartTime:
		return db.closedFilteredByClosedStatusSortedByStartTime(ctx, &filter.ListRequest, filter.CloseStatus)
	case nosqlplugin.ClosedFilteredByClosedStatusSortedByClosedTIme:
		return db.closedFilteredByClosedStatusSortedByClosedTime(ctx, &filter.ListRequest, filter.CloseStatus)
	default:
		panic("no supported filter type")
	}
}

func (db *cdb) openFilteredByWorkflowTypeSortedByStartTime(
	ctx context.Context,
	request *persistence.InternalListWorkflowExecutionsRequest,
	workflowType string,
) (*nosqlplugin.SelectVisibilityResponse, error) {
	query := db.session.Query(templateGetOpenWorkflowExecutionsByType,
		request.DomainUUID,
		domainPartition,
		persistence.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		persistence.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
		workflowType,
	).Consistency(cassandraLowConslevel).WithContext(ctx)
	return processQuery(query, request, readOpenWorkflowExecutionRecord)
}

func (db *cdb) closedFilteredByWorkflowTypeSortedByStartTime(
	ctx context.Context,
	request *persistence.InternalListWorkflowExecutionsRequest,
	workflowType string,
) (*nosqlplugin.SelectVisibilityResponse, error) {
	query := db.session.Query(templateGetClosedWorkflowExecutionsByType,
		request.DomainUUID,
		domainPartition,
		persistence.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		persistence.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
		workflowType,
	).Consistency(cassandraLowConslevel).WithContext(ctx)
	return processQuery(query, request, readOpenWorkflowExecutionRecord)
}

func (db *cdb) closedFilteredByWorkflowTypeSortedByClosedTime(
	ctx context.Context,
	request *persistence.InternalListWorkflowExecutionsRequest,
	workflowType string,
) (*nosqlplugin.SelectVisibilityResponse, error) {
	query := db.session.Query(templateGetClosedWorkflowExecutionsByTypeSortByCloseTime,
		request.DomainUUID,
		domainPartition,
		persistence.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		persistence.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
		workflowType,
	).Consistency(cassandraLowConslevel).WithContext(ctx)
	return processQuery(query, request, readOpenWorkflowExecutionRecord)
}

func (db *cdb) openFilteredByWorkflowIDSortedByStartTime(
	ctx context.Context,
	request *persistence.InternalListWorkflowExecutionsRequest,
	workflowID string,
) (*nosqlplugin.SelectVisibilityResponse, error) {
	query := db.session.Query(templateGetOpenWorkflowExecutionsByID,
		request.DomainUUID,
		domainPartition,
		persistence.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		persistence.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
		workflowID,
	).Consistency(cassandraLowConslevel).WithContext(ctx)
	return processQuery(query, request, readOpenWorkflowExecutionRecord)
}

func (db *cdb) closedFilteredByWorkflowIDSortedByStartTime(
	ctx context.Context,
	request *persistence.InternalListWorkflowExecutionsRequest,
	workflowID string,
) (*nosqlplugin.SelectVisibilityResponse, error) {
	query := db.session.Query(templateGetClosedWorkflowExecutionsByID,
		request.DomainUUID,
		domainPartition,
		persistence.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		persistence.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
		workflowID,
	).Consistency(cassandraLowConslevel).WithContext(ctx)
	return processQuery(query, request, readOpenWorkflowExecutionRecord)
}

func (db *cdb) closedFilteredByWorkflowIDSortedByClosedTime(
	ctx context.Context,
	request *persistence.InternalListWorkflowExecutionsRequest,
	workflowID string,
) (*nosqlplugin.SelectVisibilityResponse, error) {
	query := db.session.Query(templateGetClosedWorkflowExecutionsByIDSortByCloseTime,
		request.DomainUUID,
		domainPartition,
		persistence.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		persistence.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
		workflowID,
	).Consistency(cassandraLowConslevel).WithContext(ctx)
	return processQuery(query, request, readOpenWorkflowExecutionRecord)
}

func (db *cdb) closedFilteredByClosedStatusSortedByStartTime(
	ctx context.Context,
	request *persistence.InternalListWorkflowExecutionsRequest,
	closeStatus int32,
) (*nosqlplugin.SelectVisibilityResponse, error) {
	query := db.session.Query(templateGetClosedWorkflowExecutionsByStatus,
		request.DomainUUID,
		domainPartition,
		persistence.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		persistence.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
		closeStatus,
	).Consistency(cassandraLowConslevel).WithContext(ctx)
	return processQuery(query, request, readOpenWorkflowExecutionRecord)
}

func (db *cdb) closedFilteredByClosedStatusSortedByClosedTime(
	ctx context.Context,
	request *persistence.InternalListWorkflowExecutionsRequest,
	closeStatus int32,
) (*nosqlplugin.SelectVisibilityResponse, error) {
	query := db.session.Query(templateGetClosedWorkflowExecutionsByStatusSortByClosedTime,
		request.DomainUUID,
		domainPartition,
		persistence.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		persistence.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
		closeStatus,
	).Consistency(cassandraLowConslevel).WithContext(ctx)
	return processQuery(query, request, readOpenWorkflowExecutionRecord)
}

func (db *cdb) openSortedyByStartTime(
	ctx context.Context,
	request *persistence.InternalListWorkflowExecutionsRequest,
) (*nosqlplugin.SelectVisibilityResponse, error) {
	query := db.session.Query(templateGetOpenWorkflowExecutions,
		request.DomainUUID,
		domainPartition,
		persistence.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		persistence.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
	).Consistency(cassandraLowConslevel).WithContext(ctx)

	return processQuery(query, request, readOpenWorkflowExecutionRecord)
}

func (db *cdb) closedSortedByStartTime(
	ctx context.Context,
	request *persistence.InternalListWorkflowExecutionsRequest,
) (*nosqlplugin.SelectVisibilityResponse, error) {
	query := db.session.Query(templateGetClosedWorkflowExecutions,
		request.DomainUUID,
		domainPartition,
		persistence.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		persistence.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
	).Consistency(cassandraLowConslevel).WithContext(ctx)
	return processQuery(query, request, readClosedWorkflowExecutionRecord)
}

func (db *cdb) closedSortedByClosedTime(
	ctx context.Context,
	request *persistence.InternalListWorkflowExecutionsRequest,
) (*nosqlplugin.SelectVisibilityResponse, error) {
	query := db.session.Query(templateGetClosedWorkflowExecutionsSortByCloseTime,
		request.DomainUUID,
		domainPartition,
		persistence.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		persistence.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
	).Consistency(cassandraLowConslevel).WithContext(ctx)
	return processQuery(query, request, readClosedWorkflowExecutionRecord)
}

type recorderReaderFunc func(iter gocql.Iter) (*persistence.InternalVisibilityWorkflowExecutionInfo, bool)

func processQuery(
	query gocql.Query,
	request *persistence.InternalListWorkflowExecutionsRequest,
	recorderReader recorderReaderFunc,
) (*nosqlplugin.SelectVisibilityResponse, error) {
	iter := query.PageSize(request.PageSize).PageState(request.NextPageToken).Iter()
	if iter == nil {
		// TODO: may return badRequestError
		return nil, fmt.Errorf("not able to create query iterator")
	}

	response := &nosqlplugin.SelectVisibilityResponse{}
	response.Executions = make([]*persistence.InternalVisibilityWorkflowExecutionInfo, 0)
	wfexecution, has := recorderReader(iter)
	for has {
		response.Executions = append(response.Executions, wfexecution)
		wfexecution, has = recorderReader(iter)
	}

	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)
	if err := iter.Close(); err != nil {
		return nil, err
	}

	return response, nil
}

func readOpenWorkflowExecutionRecord(
	iter gocql.Iter,
) (*persistence.InternalVisibilityWorkflowExecutionInfo, bool) {
	var workflowID string
	var runID string
	var typeName string
	var startTime time.Time
	var executionTime time.Time
	var memo []byte
	var encoding string
	var taskList string
	var isCron bool
	if iter.Scan(&workflowID, &runID, &startTime, &executionTime, &typeName, &memo, &encoding, &taskList, &isCron) {
		record := &persistence.InternalVisibilityWorkflowExecutionInfo{
			WorkflowID:    workflowID,
			RunID:         runID,
			TypeName:      typeName,
			StartTime:     startTime,
			ExecutionTime: executionTime,
			Memo:          persistence.NewDataBlob(memo, common.EncodingType(encoding)),
			TaskList:      taskList,
			IsCron:        isCron,
		}
		return record, true
	}
	return nil, false
}

func readClosedWorkflowExecutionRecord(
	iter gocql.Iter,
) (*persistence.InternalVisibilityWorkflowExecutionInfo, bool) {
	var workflowID string
	var runID string
	var typeName string
	var startTime time.Time
	var executionTime time.Time
	var closeTime time.Time
	var status workflow.WorkflowExecutionCloseStatus
	var historyLength int64
	var memo []byte
	var encoding string
	var taskList string
	var isCron bool
	if iter.Scan(&workflowID, &runID, &startTime, &executionTime, &closeTime, &typeName, &status, &historyLength, &memo, &encoding, &taskList, &isCron) {
		record := &persistence.InternalVisibilityWorkflowExecutionInfo{
			WorkflowID:    workflowID,
			RunID:         runID,
			TypeName:      typeName,
			StartTime:     startTime,
			ExecutionTime: executionTime,
			CloseTime:     closeTime,
			Status:        thrift.ToWorkflowExecutionCloseStatus(&status),
			HistoryLength: historyLength,
			Memo:          persistence.NewDataBlob(memo, common.EncodingType(encoding)),
			TaskList:      taskList,
			IsCron:        isCron,
		}
		return record, true
	}
	return nil, false
}
