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

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
	"github.com/uber/cadence/common/service/config"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"
)

// Fixed domain values for now
const (
	domainPartition        = 0
	defaultCloseTTLSeconds = 86400
	openExecutionTTLBuffer = int64(86400) // setting it to a day to account for shard going down

	maxCassandraTTL = int64(157680000) // Cassandra max support time is 2038-01-19T03:14:06+00:00. Updated this to 5 years to support until year 2033
)

const (
	///////////////// Open Executions /////////////////
	openExecutionsColumnsForSelect = " workflow_id, run_id, start_time, execution_time, workflow_type_name, memo, encoding, task_list "

	openExecutionsColumnsForInsert = "(domain_id, domain_partition, " + openExecutionsColumnsForSelect + ")"

	templateCreateWorkflowExecutionStartedWithTTL = `INSERT INTO open_executions ` +
		openExecutionsColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) using TTL ?`

	templateCreateWorkflowExecutionStarted = `INSERT INTO open_executions` +
		openExecutionsColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

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
	closedExecutionColumnsForSelect = " workflow_id, run_id, start_time, execution_time, close_time, workflow_type_name, status, history_length, memo, encoding, task_list "

	closedExecutionColumnsForInsert = "(domain_id, domain_partition, " + closedExecutionColumnsForSelect + ")"

	templateCreateWorkflowExecutionClosedWithTTL = `INSERT INTO closed_executions ` +
		closedExecutionColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) using TTL ?`

	templateCreateWorkflowExecutionClosed = `INSERT INTO closed_executions ` +
		closedExecutionColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	templateCreateWorkflowExecutionClosedWithTTLV2 = `INSERT INTO closed_executions_v2 ` +
		closedExecutionColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) using TTL ?`

	templateCreateWorkflowExecutionClosedV2 = `INSERT INTO closed_executions_v2 ` +
		closedExecutionColumnsForInsert +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

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

	templateGetClosedWorkflowExecutionsV2 = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions_v2 ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition IN (?) ` +
		`AND close_time >= ? ` +
		`AND close_time <= ? `

	templateGetClosedWorkflowExecutionsByTypeV2 = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions_v2 ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND close_time >= ? ` +
		`AND close_time <= ? ` +
		`AND workflow_type_name = ? `

	templateGetClosedWorkflowExecutionsByIDV2 = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions_v2 ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND close_time >= ? ` +
		`AND close_time <= ? ` +
		`AND workflow_id = ? `

	templateGetClosedWorkflowExecutionsByStatusV2 = `SELECT ` + closedExecutionColumnsForSelect +
		`FROM closed_executions_v2 ` +
		`WHERE domain_id = ? ` +
		`AND domain_partition = ? ` +
		`AND close_time >= ? ` +
		`AND close_time <= ? ` +
		`AND status = ? `
)

type (
	cassandraVisibilityPersistence struct {
		sortByCloseTime bool
		cassandraStore
		lowConslevel gocql.Consistency
	}
)

// newVisibilityPersistence is used to create an instance of VisibilityManager implementation
func newVisibilityPersistence(
	listClosedOrderingByCloseTime bool,
	cfg config.Cassandra,
	logger log.Logger,
) (p.VisibilityStore, error) {
	session, err := cassandra.CreateSession(cfg)
	if err != nil {
		return nil, err
	}

	return &cassandraVisibilityPersistence{
		sortByCloseTime: listClosedOrderingByCloseTime,
		cassandraStore: cassandraStore{
			client:  cfg.CQLClient,
			session: session,
			logger:  logger,
		},
		lowConslevel: gocql.One,
	}, nil
}

// Close releases the resources held by this object
func (v *cassandraVisibilityPersistence) Close() {
	if v.session != nil {
		v.session.Close()
	}
}

func (v *cassandraVisibilityPersistence) RecordWorkflowExecutionStarted(
	ctx context.Context,
	request *p.InternalRecordWorkflowExecutionStartedRequest,
) error {
	ttl := int64(request.WorkflowTimeout.Seconds()) + openExecutionTTLBuffer
	var query gocql.Query
	if ttl > maxCassandraTTL {
		query = v.session.Query(templateCreateWorkflowExecutionStarted,
			request.DomainUUID,
			domainPartition,
			request.WorkflowID,
			request.RunID,
			p.UnixNanoToDBTimestamp(request.StartTimestamp.UnixNano()),
			p.UnixNanoToDBTimestamp(request.ExecutionTimestamp.UnixNano()),
			request.WorkflowTypeName,
			request.Memo.Data,
			string(request.Memo.GetEncoding()),
			request.TaskList,
		).WithContext(ctx)
	} else {
		query = v.session.Query(templateCreateWorkflowExecutionStartedWithTTL,
			request.DomainUUID,
			domainPartition,
			request.WorkflowID,
			request.RunID,
			p.UnixNanoToDBTimestamp(request.StartTimestamp.UnixNano()),
			p.UnixNanoToDBTimestamp(request.ExecutionTimestamp.UnixNano()),
			request.WorkflowTypeName,
			request.Memo.Data,
			string(request.Memo.GetEncoding()),
			request.TaskList,
			ttl,
		).WithContext(ctx)
	}
	query = query.WithTimestamp(p.UnixNanoToDBTimestamp(request.StartTimestamp.UnixNano()))
	err := query.Exec()
	if err != nil {
		return convertCommonErrors(v.client, "RecordWorkflowExecutionStarted", err)
	}

	return nil
}

func (v *cassandraVisibilityPersistence) RecordWorkflowExecutionClosed(
	ctx context.Context,
	request *p.InternalRecordWorkflowExecutionClosedRequest,
) error {
	batch := v.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)

	// First, remove execution from the open table
	batch.Query(templateDeleteWorkflowExecutionStarted,
		request.DomainUUID,
		domainPartition,
		p.UnixNanoToDBTimestamp(request.StartTimestamp.UnixNano()),
		request.RunID,
	)

	// Next, add a row in the closed table.

	// Find how long to keep the row
	retention := request.RetentionSeconds
	if retention == 0 {
		retention = defaultCloseTTLSeconds * time.Second
	}
	if int64(retention.Seconds()) > maxCassandraTTL {
		batch.Query(templateCreateWorkflowExecutionClosed,
			request.DomainUUID,
			domainPartition,
			request.WorkflowID,
			request.RunID,
			p.UnixNanoToDBTimestamp(request.StartTimestamp.UnixNano()),
			p.UnixNanoToDBTimestamp(request.ExecutionTimestamp.UnixNano()),
			p.UnixNanoToDBTimestamp(request.CloseTimestamp.UnixNano()),
			request.WorkflowTypeName,
			*thrift.FromWorkflowExecutionCloseStatus(&request.Status),
			request.HistoryLength,
			request.Memo.Data,
			string(request.Memo.GetEncoding()),
			request.TaskList,
		)
		// duplicate write to v2 to order by close time
		batch.Query(templateCreateWorkflowExecutionClosedV2,
			request.DomainUUID,
			domainPartition,
			request.WorkflowID,
			request.RunID,
			p.UnixNanoToDBTimestamp(request.StartTimestamp.UnixNano()),
			p.UnixNanoToDBTimestamp(request.ExecutionTimestamp.UnixNano()),
			p.UnixNanoToDBTimestamp(request.CloseTimestamp.UnixNano()),
			request.WorkflowTypeName,
			*thrift.FromWorkflowExecutionCloseStatus(&request.Status),
			request.HistoryLength,
			request.Memo.Data,
			string(request.Memo.GetEncoding()),
			request.TaskList,
		)
	} else {
		batch.Query(templateCreateWorkflowExecutionClosedWithTTL,
			request.DomainUUID,
			domainPartition,
			request.WorkflowID,
			request.RunID,
			p.UnixNanoToDBTimestamp(request.StartTimestamp.UnixNano()),
			p.UnixNanoToDBTimestamp(request.ExecutionTimestamp.UnixNano()),
			p.UnixNanoToDBTimestamp(request.CloseTimestamp.UnixNano()),
			request.WorkflowTypeName,
			*thrift.FromWorkflowExecutionCloseStatus(&request.Status),
			request.HistoryLength,
			request.Memo.Data,
			string(request.Memo.GetEncoding()),
			request.TaskList,
			int64(retention.Seconds()),
		)
		// duplicate write to v2 to order by close time
		batch.Query(templateCreateWorkflowExecutionClosedWithTTLV2,
			request.DomainUUID,
			domainPartition,
			request.WorkflowID,
			request.RunID,
			p.UnixNanoToDBTimestamp(request.StartTimestamp.UnixNano()),
			p.UnixNanoToDBTimestamp(request.ExecutionTimestamp.UnixNano()),
			p.UnixNanoToDBTimestamp(request.CloseTimestamp.UnixNano()),
			request.WorkflowTypeName,
			*thrift.FromWorkflowExecutionCloseStatus(&request.Status),
			request.HistoryLength,
			request.Memo.Data,
			string(request.Memo.GetEncoding()),
			request.TaskList,
			int64(retention.Seconds()),
		)
	}

	// RecordWorkflowExecutionStarted is using StartTimestamp as
	// the timestamp to issue query to Cassandra
	// due to the fact that cross DC using mutable state creation time as workflow start time
	// and visibility using event time instead of last update time (#1501)
	// CloseTimestamp can be before StartTimestamp, meaning using CloseTimestamp
	// can cause the deletion of open visibility record to be ignored.
	queryTimeStamp := request.CloseTimestamp
	if queryTimeStamp.Before(request.StartTimestamp) {
		queryTimeStamp = request.StartTimestamp.Add(time.Second)
	}
	batch = batch.WithTimestamp(p.UnixNanoToDBTimestamp(queryTimeStamp.UnixNano()))
	err := v.session.ExecuteBatch(batch)
	if err != nil {
		return convertCommonErrors(v.client, "RecordWorkflowExecutionClosed", err)
	}
	return nil
}

func (v *cassandraVisibilityPersistence) UpsertWorkflowExecution(
	ctx context.Context,
	request *p.InternalUpsertWorkflowExecutionRequest,
) error {
	if p.IsNopUpsertWorkflowRequest(request) {
		return nil
	}
	return p.NewOperationNotSupportErrorForVis()
}

func (v *cassandraVisibilityPersistence) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	query := v.session.Query(templateGetOpenWorkflowExecutions,
		request.DomainUUID,
		domainPartition,
		p.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		p.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
	).Consistency(v.lowConslevel).WithContext(ctx)
	iter := query.PageSize(request.PageSize).PageState(request.NextPageToken).Iter()
	if iter == nil {
		// TODO: should return a bad request error if the token is invalid
		return nil, &types.InternalServiceError{
			Message: "ListOpenWorkflowExecutions operation failed.  Not able to create query iterator.",
		}
	}

	response := &p.InternalListWorkflowExecutionsResponse{}
	response.Executions = make([]*p.InternalVisibilityWorkflowExecutionInfo, 0)
	wfexecution, has := readOpenWorkflowExecutionRecord(iter)
	for has {
		response.Executions = append(response.Executions, wfexecution)
		wfexecution, has = readOpenWorkflowExecutionRecord(iter)
	}

	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)
	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(v.client, "ListOpenWorkflowExecutions", err)
	}

	return response, nil
}

func (v *cassandraVisibilityPersistence) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	if v.sortByCloseTime {
		return v.listClosedWorkflowExecutionsOrderByClosedTime(ctx, request)
	}
	query := v.session.Query(templateGetClosedWorkflowExecutions,
		request.DomainUUID,
		domainPartition,
		p.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		p.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
	).Consistency(v.lowConslevel).WithContext(ctx)
	iter := query.PageSize(request.PageSize).PageState(request.NextPageToken).Iter()
	if iter == nil {
		// TODO: should return a bad request error if the token is invalid
		return nil, &types.InternalServiceError{
			Message: "ListClosedWorkflowExecutions operation failed.  Not able to create query iterator.",
		}
	}

	response := &p.InternalListWorkflowExecutionsResponse{}
	response.Executions = make([]*p.InternalVisibilityWorkflowExecutionInfo, 0)
	wfexecution, has := readClosedWorkflowExecutionRecord(iter)
	for has {
		response.Executions = append(response.Executions, wfexecution)
		wfexecution, has = readClosedWorkflowExecutionRecord(iter)
	}

	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)
	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(v.client, "ListClosedWorkflowExecutions", err)
	}

	return response, nil
}

func (v *cassandraVisibilityPersistence) ListOpenWorkflowExecutionsByType(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByTypeRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	query := v.session.Query(templateGetOpenWorkflowExecutionsByType,
		request.DomainUUID,
		domainPartition,
		p.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		p.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
		request.WorkflowTypeName,
	).Consistency(v.lowConslevel).WithContext(ctx)
	iter := query.PageSize(request.PageSize).PageState(request.NextPageToken).Iter()
	if iter == nil {
		// TODO: should return a bad request error if the token is invalid
		return nil, &types.InternalServiceError{
			Message: "ListOpenWorkflowExecutionsByType operation failed.  Not able to create query iterator.",
		}
	}

	response := &p.InternalListWorkflowExecutionsResponse{}
	response.Executions = make([]*p.InternalVisibilityWorkflowExecutionInfo, 0)
	wfexecution, has := readOpenWorkflowExecutionRecord(iter)
	for has {
		response.Executions = append(response.Executions, wfexecution)
		wfexecution, has = readOpenWorkflowExecutionRecord(iter)
	}

	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)
	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(v.client, "ListOpenWorkflowExecutionsByType", err)
	}

	return response, nil
}

func (v *cassandraVisibilityPersistence) ListClosedWorkflowExecutionsByType(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByTypeRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	if v.sortByCloseTime {
		return v.listClosedWorkflowExecutionsByTypeOrderByClosedTime(ctx, request)
	}
	query := v.session.Query(templateGetClosedWorkflowExecutionsByType,
		request.DomainUUID,
		domainPartition,
		p.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		p.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
		request.WorkflowTypeName,
	).Consistency(v.lowConslevel).WithContext(ctx)
	iter := query.PageSize(request.PageSize).PageState(request.NextPageToken).Iter()
	if iter == nil {
		// TODO: should return a bad request error if the token is invalid
		return nil, &types.InternalServiceError{
			Message: "ListClosedWorkflowExecutionsByType operation failed.  Not able to create query iterator.",
		}
	}

	response := &p.InternalListWorkflowExecutionsResponse{}
	response.Executions = make([]*p.InternalVisibilityWorkflowExecutionInfo, 0)
	wfexecution, has := readClosedWorkflowExecutionRecord(iter)
	for has {
		response.Executions = append(response.Executions, wfexecution)
		wfexecution, has = readClosedWorkflowExecutionRecord(iter)
	}

	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)
	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(v.client, "ListClosedWorkflowExecutionsByType", err)
	}

	return response, nil
}

func (v *cassandraVisibilityPersistence) ListOpenWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByWorkflowIDRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	query := v.session.Query(templateGetOpenWorkflowExecutionsByID,
		request.DomainUUID,
		domainPartition,
		p.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		p.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
		request.WorkflowID,
	).Consistency(v.lowConslevel).WithContext(ctx)
	iter := query.PageSize(request.PageSize).PageState(request.NextPageToken).Iter()
	if iter == nil {
		// TODO: should return a bad request error if the token is invalid
		return nil, &types.InternalServiceError{
			Message: "ListOpenWorkflowExecutionsByWorkflowID operation failed.  Not able to create query iterator.",
		}
	}

	response := &p.InternalListWorkflowExecutionsResponse{}
	response.Executions = make([]*p.InternalVisibilityWorkflowExecutionInfo, 0)
	wfexecution, has := readOpenWorkflowExecutionRecord(iter)
	for has {
		response.Executions = append(response.Executions, wfexecution)
		wfexecution, has = readOpenWorkflowExecutionRecord(iter)
	}

	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)
	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(v.client, "ListOpenWorkflowExecutionsByWorkflowID", err)
	}

	return response, nil
}

func (v *cassandraVisibilityPersistence) ListClosedWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByWorkflowIDRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	if v.sortByCloseTime {
		return v.listClosedWorkflowExecutionsByWorkflowIDOrderByClosedTime(ctx, request)
	}
	query := v.session.Query(templateGetClosedWorkflowExecutionsByID,
		request.DomainUUID,
		domainPartition,
		p.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		p.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
		request.WorkflowID,
	).Consistency(v.lowConslevel).WithContext(ctx)
	iter := query.PageSize(request.PageSize).PageState(request.NextPageToken).Iter()
	if iter == nil {
		// TODO: should return a bad request error if the token is invalid
		return nil, &types.InternalServiceError{
			Message: "ListClosedWorkflowExecutionsByWorkflowID operation failed.  Not able to create query iterator.",
		}
	}

	response := &p.InternalListWorkflowExecutionsResponse{}
	response.Executions = make([]*p.InternalVisibilityWorkflowExecutionInfo, 0)
	wfexecution, has := readClosedWorkflowExecutionRecord(iter)
	for has {
		response.Executions = append(response.Executions, wfexecution)
		wfexecution, has = readClosedWorkflowExecutionRecord(iter)
	}

	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)
	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(v.client, "ListClosedWorkflowExecutionsByWorkflowID", err)
	}

	return response, nil
}

func (v *cassandraVisibilityPersistence) ListClosedWorkflowExecutionsByStatus(
	ctx context.Context,
	request *p.InternalListClosedWorkflowExecutionsByStatusRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	if v.sortByCloseTime {
		return v.listClosedWorkflowExecutionsByStatusOrderByClosedTime(ctx, request)
	}
	query := v.session.Query(templateGetClosedWorkflowExecutionsByStatus,
		request.DomainUUID,
		domainPartition,
		p.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		p.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
		*thrift.FromWorkflowExecutionCloseStatus(&request.Status),
	).Consistency(v.lowConslevel).WithContext(ctx)
	iter := query.PageSize(request.PageSize).PageState(request.NextPageToken).Iter()
	if iter == nil {
		// TODO: should return a bad request error if the token is invalid
		return nil, &types.InternalServiceError{
			Message: "ListClosedWorkflowExecutionsByStatus operation failed.  Not able to create query iterator.",
		}
	}

	response := &p.InternalListWorkflowExecutionsResponse{}
	response.Executions = make([]*p.InternalVisibilityWorkflowExecutionInfo, 0)
	wfexecution, has := readClosedWorkflowExecutionRecord(iter)
	for has {
		response.Executions = append(response.Executions, wfexecution)
		wfexecution, has = readClosedWorkflowExecutionRecord(iter)
	}

	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)
	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(v.client, "ListClosedWorkflowExecutionsByStatus", err)
	}

	return response, nil
}

func (v *cassandraVisibilityPersistence) GetClosedWorkflowExecution(
	ctx context.Context,
	request *p.InternalGetClosedWorkflowExecutionRequest,
) (*p.InternalGetClosedWorkflowExecutionResponse, error) {
	execution := request.Execution
	query := v.session.Query(templateGetClosedWorkflowExecution,
		request.DomainUUID,
		domainPartition,
		execution.GetWorkflowID(),
		execution.GetRunID(),
	).WithContext(ctx)

	iter := query.Iter()
	if iter == nil {
		return nil, &types.InternalServiceError{
			Message: "GetClosedWorkflowExecution operation failed.  Not able to create query iterator.",
		}
	}

	wfexecution, has := readClosedWorkflowExecutionRecord(iter)
	if !has {
		return nil, &types.EntityNotExistsError{
			Message: fmt.Sprintf("Workflow execution not found.  WorkflowId: %v, RunId: %v",
				execution.GetWorkflowID(), execution.GetRunID()),
		}
	}

	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(v.client, "GetClosedWorkflowExecution", err)
	}

	return &p.InternalGetClosedWorkflowExecutionResponse{
		Execution: wfexecution,
	}, nil
}

// DeleteWorkflowExecution is a no-op since deletes are auto-handled by cassandra TTLs
func (v *cassandraVisibilityPersistence) DeleteWorkflowExecution(
	ctx context.Context,
	request *p.VisibilityDeleteWorkflowExecutionRequest,
) error {
	return nil
}

func (v *cassandraVisibilityPersistence) ListWorkflowExecutions(
	ctx context.Context,
	request *p.ListWorkflowExecutionsByQueryRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	return nil, p.NewOperationNotSupportErrorForVis()
}

func (v *cassandraVisibilityPersistence) ScanWorkflowExecutions(
	ctx context.Context,
	request *p.ListWorkflowExecutionsByQueryRequest) (*p.InternalListWorkflowExecutionsResponse, error) {
	return nil, p.NewOperationNotSupportErrorForVis()
}

func (v *cassandraVisibilityPersistence) CountWorkflowExecutions(
	ctx context.Context,
	request *p.CountWorkflowExecutionsRequest,
) (*p.CountWorkflowExecutionsResponse, error) {
	return nil, p.NewOperationNotSupportErrorForVis()
}

func (v *cassandraVisibilityPersistence) listClosedWorkflowExecutionsOrderByClosedTime(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	query := v.session.Query(templateGetClosedWorkflowExecutionsV2,
		request.DomainUUID,
		domainPartition,
		p.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		p.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
	).Consistency(v.lowConslevel).WithContext(ctx)
	iter := query.PageSize(request.PageSize).PageState(request.NextPageToken).Iter()
	if iter == nil {
		// TODO: should return a bad request error if the token is invalid
		return nil, &types.InternalServiceError{
			Message: "ListClosedWorkflowExecutions operation failed.  Not able to create query iterator.",
		}
	}

	response := &p.InternalListWorkflowExecutionsResponse{}
	response.Executions = make([]*p.InternalVisibilityWorkflowExecutionInfo, 0)
	wfexecution, has := readClosedWorkflowExecutionRecord(iter)
	for has {
		response.Executions = append(response.Executions, wfexecution)
		wfexecution, has = readClosedWorkflowExecutionRecord(iter)
	}

	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)
	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(v.client, "ListClosedWorkflowExecutions", err)
	}

	return response, nil
}

func (v *cassandraVisibilityPersistence) listClosedWorkflowExecutionsByTypeOrderByClosedTime(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByTypeRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	query := v.session.Query(templateGetClosedWorkflowExecutionsByTypeV2,
		request.DomainUUID,
		domainPartition,
		p.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		p.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
		request.WorkflowTypeName,
	).Consistency(v.lowConslevel).WithContext(ctx)
	iter := query.PageSize(request.PageSize).PageState(request.NextPageToken).Iter()
	if iter == nil {
		// TODO: should return a bad request error if the token is invalid
		return nil, &types.InternalServiceError{
			Message: "ListClosedWorkflowExecutionsByType operation failed.  Not able to create query iterator.",
		}
	}

	response := &p.InternalListWorkflowExecutionsResponse{}
	response.Executions = make([]*p.InternalVisibilityWorkflowExecutionInfo, 0)
	wfexecution, has := readClosedWorkflowExecutionRecord(iter)
	for has {
		response.Executions = append(response.Executions, wfexecution)
		wfexecution, has = readClosedWorkflowExecutionRecord(iter)
	}

	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)
	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(v.client, "ListClosedWorkflowExecutionsByType", err)
	}

	return response, nil
}

func (v *cassandraVisibilityPersistence) listClosedWorkflowExecutionsByWorkflowIDOrderByClosedTime(
	ctx context.Context,
	request *p.InternalListWorkflowExecutionsByWorkflowIDRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	query := v.session.Query(templateGetClosedWorkflowExecutionsByIDV2,
		request.DomainUUID,
		domainPartition,
		p.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		p.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
		request.WorkflowID,
	).Consistency(v.lowConslevel).WithContext(ctx)
	iter := query.PageSize(request.PageSize).PageState(request.NextPageToken).Iter()
	if iter == nil {
		// TODO: should return a bad request error if the token is invalid
		return nil, &types.InternalServiceError{
			Message: "ListClosedWorkflowExecutionsByWorkflowID operation failed.  Not able to create query iterator.",
		}
	}

	response := &p.InternalListWorkflowExecutionsResponse{}
	response.Executions = make([]*p.InternalVisibilityWorkflowExecutionInfo, 0)
	wfexecution, has := readClosedWorkflowExecutionRecord(iter)
	for has {
		response.Executions = append(response.Executions, wfexecution)
		wfexecution, has = readClosedWorkflowExecutionRecord(iter)
	}

	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)
	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(v.client, "ListClosedWorkflowExecutionsByWorkflowID", err)
	}

	return response, nil
}

func (v *cassandraVisibilityPersistence) listClosedWorkflowExecutionsByStatusOrderByClosedTime(
	ctx context.Context,
	request *p.InternalListClosedWorkflowExecutionsByStatusRequest,
) (*p.InternalListWorkflowExecutionsResponse, error) {
	query := v.session.Query(templateGetClosedWorkflowExecutionsByStatusV2,
		request.DomainUUID,
		domainPartition,
		p.UnixNanoToDBTimestamp(request.EarliestTime.UnixNano()),
		p.UnixNanoToDBTimestamp(request.LatestTime.UnixNano()),
		request.Status,
	).Consistency(v.lowConslevel).WithContext(ctx)
	iter := query.PageSize(request.PageSize).PageState(request.NextPageToken).Iter()
	if iter == nil {
		// TODO: should return a bad request error if the token is invalid
		return nil, &types.InternalServiceError{
			Message: "ListClosedWorkflowExecutionsByStatus operation failed.  Not able to create query iterator.",
		}
	}

	response := &p.InternalListWorkflowExecutionsResponse{}
	response.Executions = make([]*p.InternalVisibilityWorkflowExecutionInfo, 0)
	wfexecution, has := readClosedWorkflowExecutionRecord(iter)
	for has {
		response.Executions = append(response.Executions, wfexecution)
		wfexecution, has = readClosedWorkflowExecutionRecord(iter)
	}

	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)
	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(v.client, "ListClosedWorkflowExecutionsByStatus", err)
	}

	return response, nil
}

func readOpenWorkflowExecutionRecord(
	iter gocql.Iter,
) (*p.InternalVisibilityWorkflowExecutionInfo, bool) {
	var workflowID string
	var runID string
	var typeName string
	var startTime time.Time
	var executionTime time.Time
	var memo []byte
	var encoding string
	var taskList string
	if iter.Scan(&workflowID, &runID, &startTime, &executionTime, &typeName, &memo, &encoding, &taskList) {
		record := &p.InternalVisibilityWorkflowExecutionInfo{
			WorkflowID:    workflowID,
			RunID:         runID,
			TypeName:      typeName,
			StartTime:     startTime,
			ExecutionTime: executionTime,
			Memo:          p.NewDataBlob(memo, common.EncodingType(encoding)),
			TaskList:      taskList,
		}
		return record, true
	}
	return nil, false
}

func readClosedWorkflowExecutionRecord(
	iter gocql.Iter,
) (*p.InternalVisibilityWorkflowExecutionInfo, bool) {
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
	if iter.Scan(&workflowID, &runID, &startTime, &executionTime, &closeTime, &typeName, &status, &historyLength, &memo, &encoding, &taskList) {
		record := &p.InternalVisibilityWorkflowExecutionInfo{
			WorkflowID:    workflowID,
			RunID:         runID,
			TypeName:      typeName,
			StartTime:     startTime,
			ExecutionTime: executionTime,
			CloseTime:     closeTime,
			Status:        thrift.ToWorkflowExecutionCloseStatus(&status),
			HistoryLength: historyLength,
			Memo:          p.NewDataBlob(memo, common.EncodingType(encoding)),
			TaskList:      taskList,
		}
		return record, true
	}
	return nil, false
}
