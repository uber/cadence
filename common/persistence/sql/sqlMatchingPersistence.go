// Copyright (c) 2018 Uber Technologies, Inc.
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

package sql

import (
	"database/sql"
	"fmt"
	"time"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/persistence"

	"github.com/hmgle/sqlx"
	"github.com/uber/cadence/common"
	"github.com/uber-common/bark"
)

type (
	sqlMatchingManager struct {
		db      *sqlx.DB
		shardID int
		logger             bark.Logger
	}

	FlatCreateWorkflowExecutionRequest struct {
		DomainID               string  `db:"domain_id"`
		WorkflowID             string  `db:"workflow_id"`
		RunID                  string  `db:"run_id"`
		ParentDomainID         *string `db:"parent_domain_id"`
		ParentWorkflowID       *string `db:"parent_workflow_id"`
		ParentRunID            *string `db:"parent_run_id"`
		InitiatedID            *int64  `db:"initiated_id"`
		TaskList               string  `db:"task_list"`
		WorkflowTypeName       string  `db:"workflow_type_name"`
		WorkflowTimeoutSeconds int64   `db:"workflow_timeout_seconds"`
		DecisionTimeoutValue   int64   `db:"decision_timeout_value"`
		ExecutionContext       []byte  `db:"execution_context"`
		NextEventID            int64   `db:"next_event_id"`
		LastProcessedEvent     int64   `db:"last_processed_event"`
		// maybe i don't need this.
	}

	executionRow struct {
		DomainID                     string    `db:"domain_id"`
		WorkflowID                   string    `db:"workflow_id"`
		RunID                        string    `db:"run_id"`
		ParentDomainID               *string   `db:"parent_domain_id"`
		ParentWorkflowID             *string   `db:"parent_workflow_id"`
		ParentRunID                  *string   `db:"parent_run_id"`
		InitiatedID                  *int64    `db:"initiated_id"`
		CompletionEvent              *[]byte   `db:"completion_event"`
		TaskList                     string    `db:"task_list"`
		WorkflowTypeName             string    `db:"workflow_type_name"`
		WorkflowTimeoutSeconds       int64     `db:"workflow_timeout_seconds"`
		DecisionTaskTimeoutMinutes   int64     `db:"decision_task_timeout_minutes"`
		ExecutionContext             *[]byte   `db:"execution_context"`
		State                        int64     `db:"state"`
		CloseStatus                  int64     `db:"close_status"`
		StartVersion                 *int64    `db:"start_version"`
		LastFirstEventID             int64     `db:"last_first_event_id"`
		NextEventID                  int64     `db:"next_event_id"`
		LastProcessedEvent           int64     `db:"last_processed_event"`
		StartTime                    time.Time `db:"start_time"`
		LastUpdatedTime              time.Time `db:"last_updated_time"`
		CreateRequestID              string    `db:"create_request_id"`
		DecisionVersion              int64     `db:"decision_version"`
		DecisionScheduleID           int64     `db:"decision_schedule_id"`
		DecisionStartedID            int64     `db:"decision_started_id"`
		DecisionRequestID            string    `db:"decision_request_id"`
		DecisionTimeout              int64     `db:"decision_timeout"`
		DecisionAttempt              int64     `db:"decision_attempt"`
		DecisionTimestamp            int64     `db:"decision_timestamp"`
		CancelRequested              *int64    `db:"cancel_requested"`
		CancelRequestID              *string   `db:"cancel_request_id"`
		StickyTaskList               string    `db:"sticky_task_list"`
		StickyScheduleToStartTimeout int64     `db:"sticky_schedule_to_start_timeout"`
		ClientLibraryVersion         string    `db:"client_library_version"`
		ClientFeatureVersion         string    `db:"client_feature_version"`
		ClientImpl                   string    `db:"client_impl"`
		ShardID                      int       `db:"shard_id"`
	}

	currentExecutionRow struct {
		ShardID    int64  `db:"shard_id"`
		DomainID   string `db:"domain_id"`
		WorkflowID string `db:"workflow_id"`

		RunID           string `db:"run_id"`
		CreateRequestID string `db:"create_request_id"`
		State           int64  `db:"state"`
		CloseStatus     int64  `db:"close_status"`
		StartVersion    *int64 `db:"start_version"`
	}
)

const (
	executionsNonNullableColumns = `shard_id,
domain_id, 
workflow_id, 
run_id, 
task_list, 
workflow_type_name, 
workflow_timeout_seconds,
decision_task_timeout_minutes,
state,
close_status,
start_version,
last_first_event_id,
next_event_id,
last_processed_event,
start_time,
last_updated_time,
create_request_id,
decision_version,
decision_schedule_id,
decision_started_id,
decision_timeout,
decision_attempt,
decision_timestamp,
sticky_task_list,
sticky_schedule_to_start_timeout,
client_library_version,
client_feature_version,
client_impl`

	executionsNonNullableColumnsTags = `:shard_id,
:domain_id,
:workflow_id,
:run_id,
:task_list,
:workflow_type_name,
:workflow_timeout_seconds,
:decision_task_timeout_minutes,
:state,
:close_status,
:start_version,
:last_first_event_id,
:next_event_id,
:last_processed_event,
:start_time,
:last_updated_time,
:create_request_id,
:decision_version,
:decision_schedule_id,
:decision_started_id,
:decision_timeout,
:decision_attempt,
:decision_timestamp,
:sticky_task_list,
:sticky_schedule_to_start_timeout,
:client_library_version,
:client_feature_version,
:client_impl`

	executionsBlobColumns = `completion_event,
execution_context`

	executionsBlobColumnsTags = `:completion_event,
:execution_context`

	// Excluding completion_event
	executionsNonblobParentColumns = `parent_domain_id,
parent_workflow_id,
parent_run_id,
initiated_id`

	executionsNonblobParentColumnsTags = `:parent_domain_id,
:parent_workflow_id,
:parent_run_id,
:initiated_id`

	executionsCancelColumns = `cancel_requested,
cancel_request_id`

	executionsCancelColumnsTags = `:cancel_requested,
:cancel_request_id`

	executionsReplicationStateColumns     = `start_version`
	executionsReplicationStateColumnsTags = `:start_version`

	createExecutionWithNoParentSQLQuery = `INSERT INTO executions 
(` + executionsNonNullableColumns +
		`,
execution_context,
cancel_requested,
cancel_request_id)
VALUES
(` + executionsNonNullableColumnsTags + `,
:execution_context,
:cancel_requested,
:cancel_request_id)
`

	updateExecutionSQLQuery = `UPDATE executions SET
domain_id = :domain_id,
workflow_id = :workflow_id,
run_id = :run_id,
parent_domain_id = :parent_domain_id,
parent_workflow_id = :parent_workflow_id,
parent_run_id = :parent_run_id,
initiated_id = :initiated_id,
completion_event = :completion_event,
task_list = :task_list,
workflow_type_name = :workflow_type_name,
workflow_timeout_seconds = :workflow_timeout_seconds,
decision_task_timeout_minutes = :decision_task_timeout_minutes,
execution_context = :execution_context,
state = :state,
close_status = :close_status,
last_first_event_id = :last_first_event_id,
next_event_id = :next_event_id,
last_processed_event = :last_processed_event,
start_time = :start_time,
last_updated_time = :last_updated_time,
create_request_id = :create_request_id,
decision_version = :decision_version,
decision_schedule_id = :decision_schedule_id,
decision_started_id = :decision_started_id,
decision_request_id = :decision_request_id,
decision_timeout = :decision_timeout,
decision_attempt = :decision_attempt,
decision_timestamp = :decision_timestamp,
cancel_requested = :cancel_requested,
cancel_request_id = :cancel_request_id,
sticky_task_list = :sticky_task_list,
sticky_schedule_to_start_timeout = :sticky_schedule_to_start_timeout,
client_library_version = :client_library_version,
client_feature_version = :client_feature_version,
client_impl = :client_impl
WHERE
shard_id = :shard_id AND
domain_id = :domain_id AND
workflow_id = :workflow_id AND
run_id = :run_id
`

	getExecutionSQLQuery = `SELECT ` +
		executionsNonNullableColumns + "," +
		executionsBlobColumns + "," +
		executionsNonblobParentColumns + "," +
		executionsCancelColumns + "," +
		executionsReplicationStateColumns +
		` FROM executions WHERE
shard_id = ? AND
domain_id = ? AND
workflow_id = ? AND
run_id = ?`

	deleteExecutionSQLQuery = `DELETE FROM executions WHERE
shard_id = ? AND
domain_id = ? AND
workflow_id = ? AND
run_id = ?`

	transferTaskInfoColumns = `domain_id,
workflow_id,
run_id,
task_id,
type,
target_domain_id,
target_workflow_id,
target_run_id,
target_child_workflow_only,
task_list,
schedule_id,
version`

	transferTaskInfoColumnsTags = `:domain_id,
:workflow_id,
:run_id,
:task_id,
:type,
:target_domain_id,
:target_workflow_id,
:target_run_id,
:target_child_workflow_only,
:task_list,
:schedule_id,
:version`

	transferTasksColumns = `shard_id,` + transferTaskInfoColumns

	transferTasksColumnsTags = `:shard_id,` + transferTaskInfoColumnsTags

	getTransferTasksSQLQuery = `SELECT
` + transferTaskInfoColumns +
		`
FROM transfer_tasks WHERE
task_id > ? AND
task_id <= ?
`

	createCurrentExecutionSQLQuery = `INSERT INTO current_executions
(shard_id, domain_id, workflow_id, run_id, create_request_id, state, close_status, start_version) VALUES
(:shard_id, :domain_id, :workflow_id, :run_id, :create_request_id, :state, :close_status, :start_version)`

	getCurrentExecutionSQLQuery = `SELECT 
shard_id, domain_id, workflow_id, run_id, state, close_status, start_version 
FROM current_executions
WHERE
shard_id = ? AND domain_id = ? AND workflow_id = ?
`

	createTransferTasksSQLQuery = `INSERT INTO transfer_tasks
(` + transferTasksColumns + `)
VALUES
(` + transferTasksColumnsTags + `
)
`

	completeTransferTaskSQLQuery = `DELETE FROM transfer_tasks WHERE shard_id = :shard_id AND task_id = :task_id`

	lockAndCheckNextEventIdSQLQuery = `SELECT next_event_id FROM executions WHERE
shard_id = ? AND
domain_id = ? AND
workflow_id = ? AND
run_id = ?
FOR UPDATE`
)

func (m *sqlMatchingManager) Close() {
	if m.db != nil {
		m.db.Close()
	}
}

func (m *sqlMatchingManager) CreateWorkflowExecution(request *persistence.CreateWorkflowExecutionRequest) (*persistence.CreateWorkflowExecutionResponse, error) {
	/*
		make a transaction
		check for a parent
		check for continueasnew, update the curret exec or create one for this new workflow
		create a workflow with/without cross dc
	*/

	tx, err := m.db.Beginx()
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateWorkflowExecution failed. Failed to start transaction. Error: %v", err),
		}
	}
	defer tx.Rollback()

	if row, err := getCurrentExecutionIfExists(tx, m.shardID, request.DomainID, *request.Execution.WorkflowId); err == nil {
		// Workflow already exists.
		startVersion := common.EmptyVersion
		if row.StartVersion != nil {
			startVersion = *row.StartVersion
		}

		return nil, &persistence.WorkflowExecutionAlreadyStartedError{
			Msg:            fmt.Sprintf("Workflow execution already running. WorkflowId: %v", row.WorkflowID),
			StartRequestID: row.CreateRequestID,
			RunID:          row.RunID,
			State:          int(row.State),
			CloseStatus:    int(row.CloseStatus),
			StartVersion:   startVersion,
		}
	}

	if err := createCurrentExecution(tx, request, m.shardID); err != nil {
		return nil, err
	}

	if err := createExecution(tx, request, m.shardID); err != nil {
		return nil, err
	}

	if err := m.createTransferTasks(tx, request.TransferTasks, m.shardID, request.DomainID, *request.Execution.WorkflowId, *request.Execution.RunId); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateWorkflowExecution operation failed. Failed to create transfer tasks. Error: %v", err),
		}
	}

	if err := lockShard(tx, m.shardID, request.RangeID); err != nil {
		switch err.(type) {
		case *persistence.ShardOwnershipLostError:
			return nil, err
		default:
			return nil, &workflow.InternalServiceError{
				Message: fmt.Sprintf("CreateWorkflowExecution operation failed. Error: %v", err),
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateWorkflowExecution operation failed. Failed to commit transaction. Error: %v", err),
		}
	}

	return &persistence.CreateWorkflowExecutionResponse{}, nil
}

func (m *sqlMatchingManager) GetWorkflowExecution(request *persistence.GetWorkflowExecutionRequest) (*persistence.GetWorkflowExecutionResponse, error) {
	var execution executionRow
	if err := m.db.Get(&execution, getExecutionSQLQuery,
		m.shardID,
		request.DomainID,
		*request.Execution.WorkflowId,
		*request.Execution.RunId); err != nil {
		if err == sql.ErrNoRows {
			return nil, &workflow.EntityNotExistsError{
				Message: fmt.Sprintf("Workflow execution not found.  WorkflowId: %v, RunId: %v",
					*request.Execution.WorkflowId,
					*request.Execution.RunId),
			}
		}
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("GetWorkflowExecution failed. Error: %v", err),
		}
	}

	var state persistence.WorkflowMutableState
	state.ExecutionInfo = &persistence.WorkflowExecutionInfo{
		DomainID:                     execution.DomainID,
		WorkflowID:                   execution.WorkflowID,
		RunID:                        execution.RunID,
		TaskList:                     execution.TaskList,
		WorkflowTypeName:             execution.WorkflowTypeName,
		WorkflowTimeout:              int32(execution.WorkflowTimeoutSeconds),
		DecisionTimeoutValue:         int32(execution.DecisionTaskTimeoutMinutes),
		State:                        int(execution.State),
		CloseStatus:                  int(execution.CloseStatus),
		LastFirstEventID:             execution.LastFirstEventID,
		NextEventID:                  execution.NextEventID,
		LastProcessedEvent:           execution.LastProcessedEvent,
		StartTimestamp:               execution.StartTime,
		LastUpdatedTimestamp:         execution.LastUpdatedTime,
		CreateRequestID:              execution.CreateRequestID,
		DecisionVersion:              execution.DecisionVersion,
		DecisionScheduleID:           execution.DecisionScheduleID,
		DecisionStartedID:            execution.DecisionStartedID,
		DecisionRequestID:            execution.DecisionRequestID,
		DecisionTimeout:              int32(execution.DecisionTimeout),
		DecisionAttempt:              execution.DecisionAttempt,
		DecisionTimestamp:            execution.DecisionTimestamp,
		StickyTaskList:               execution.StickyTaskList,
		StickyScheduleToStartTimeout: int32(execution.StickyScheduleToStartTimeout),
		ClientLibraryVersion:         execution.ClientLibraryVersion,
		ClientFeatureVersion:         execution.ClientFeatureVersion,
		ClientImpl:                   execution.ClientImpl,
	}

	if execution.ExecutionContext != nil {
		state.ExecutionInfo.ExecutionContext = *execution.ExecutionContext
	}

	if execution.StartVersion != nil {
		state.ReplicationState = &persistence.ReplicationState{
			StartVersion: *execution.StartVersion,
		}
	}

	if execution.ParentDomainID != nil {
		state.ExecutionInfo.ParentDomainID = *execution.ParentDomainID
		state.ExecutionInfo.ParentWorkflowID = *execution.ParentWorkflowID
		state.ExecutionInfo.ParentRunID = *execution.ParentRunID
		state.ExecutionInfo.InitiatedID = *execution.InitiatedID
		if state.ExecutionInfo.CompletionEvent != nil {
			state.ExecutionInfo.CompletionEvent = *execution.CompletionEvent
		}
	}

	if execution.CancelRequested != nil && (*execution.CancelRequested != 0) {
		state.ExecutionInfo.CancelRequested = true
		state.ExecutionInfo.CancelRequestID = *execution.CancelRequestID
	}

	return &persistence.GetWorkflowExecutionResponse{State: &state}, nil
}

func (m *sqlMatchingManager) UpdateWorkflowExecution(request *persistence.UpdateWorkflowExecutionRequest) error {
	tx, err := m.db.Beginx()
	if err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("UpdateWorkflowExecution operation failed. Failed to begin transaction. Erorr: %v", err),
		}
	}
	defer tx.Rollback()

	if err := m.createTransferTasks(tx,
		request.TransferTasks,
		m.shardID,
		request.ExecutionInfo.DomainID,
		request.ExecutionInfo.WorkflowID,
		request.ExecutionInfo.RunID); err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("UpdateWorkflowExecution operation failed. Failed to create transfer tasks. Error: %v", err),
		}
	}

	if err := lockShard(tx, m.shardID, request.RangeID); err != nil {
		switch err.(type) {
		case *persistence.ShardOwnershipLostError:
			return err
		default:
			return &workflow.InternalServiceError{
				Message: fmt.Sprintf("CreateWorkflowExecution operation failed. Error: %v", err),
			}
		}
	}

	// TODO Remove me if UPDATE holds the lock to the end of a transaction
	if err := lockAndCheckNextEventID(tx,
		m.shardID,
		request.ExecutionInfo.DomainID,
		request.ExecutionInfo.WorkflowID,
		request.ExecutionInfo.RunID,
		request.Condition); err != nil {
		switch err.(type) {
		case *persistence.ConditionFailedError:
			return err
		default:
			return &workflow.InternalServiceError{
				Message: fmt.Sprintf("UpdateWorkflowExecution operation failed. Failed to lock executions row. Error: %v", err),
			}
		}
	}

	args := struct {
		executionRow
		Condition int64 `db:"old_next_event_id"`
	}{
		executionRow{
			DomainID:                   request.ExecutionInfo.DomainID,
			WorkflowID:                 request.ExecutionInfo.WorkflowID,
			RunID:                      request.ExecutionInfo.RunID,
			ParentDomainID:             &request.ExecutionInfo.ParentDomainID,
			ParentWorkflowID:           &request.ExecutionInfo.ParentWorkflowID,
			ParentRunID:                &request.ExecutionInfo.ParentRunID,
			InitiatedID:                &request.ExecutionInfo.InitiatedID,
			CompletionEvent:            &request.ExecutionInfo.CompletionEvent,
			TaskList:                   request.ExecutionInfo.TaskList,
			WorkflowTypeName:           request.ExecutionInfo.WorkflowTypeName,
			WorkflowTimeoutSeconds:     int64(request.ExecutionInfo.WorkflowTimeout),
			DecisionTaskTimeoutMinutes: int64(request.ExecutionInfo.DecisionTimeoutValue),
			State:                        int64(request.ExecutionInfo.State),
			CloseStatus:                  int64(request.ExecutionInfo.CloseStatus),
			LastFirstEventID:             int64(request.ExecutionInfo.LastFirstEventID),
			NextEventID:                  int64(request.ExecutionInfo.NextEventID),
			LastProcessedEvent:           int64(request.ExecutionInfo.LastProcessedEvent),
			StartTime:                    request.ExecutionInfo.StartTimestamp,
			LastUpdatedTime:              request.ExecutionInfo.LastUpdatedTimestamp,
			CreateRequestID:              request.ExecutionInfo.CreateRequestID,
			DecisionVersion:              request.ExecutionInfo.DecisionVersion,
			DecisionScheduleID:           request.ExecutionInfo.DecisionScheduleID,
			DecisionStartedID:            request.ExecutionInfo.DecisionStartedID,
			DecisionRequestID:            request.ExecutionInfo.DecisionRequestID,
			DecisionTimeout:              int64(request.ExecutionInfo.DecisionTimeout),
			DecisionAttempt:              request.ExecutionInfo.DecisionAttempt,
			DecisionTimestamp:            request.ExecutionInfo.DecisionTimestamp,
			StickyTaskList:               request.ExecutionInfo.StickyTaskList,
			StickyScheduleToStartTimeout: int64(request.ExecutionInfo.StickyScheduleToStartTimeout),
			ClientLibraryVersion:         request.ExecutionInfo.ClientLibraryVersion,
			ClientFeatureVersion:         request.ExecutionInfo.ClientFeatureVersion,
			ClientImpl:                   request.ExecutionInfo.ClientImpl,
		},
		request.Condition,
	}

	if request.ExecutionInfo.ExecutionContext != nil {
		args.executionRow.ExecutionContext = &request.ExecutionInfo.ExecutionContext
	}

	if request.ReplicationState != nil {
		args.StartVersion = &request.ReplicationState.StartVersion
	}

	if request.ExecutionInfo.ParentDomainID != "" {
		args.ParentDomainID = &request.ExecutionInfo.ParentDomainID
		args.ParentWorkflowID = &request.ExecutionInfo.ParentWorkflowID
		args.ParentRunID = &request.ExecutionInfo.ParentRunID
		args.InitiatedID = &request.ExecutionInfo.InitiatedID
		args.CompletionEvent = &request.ExecutionInfo.CompletionEvent
	}

	if request.ExecutionInfo.CancelRequested {
		var i int64 = 1
		args.CancelRequested = &i
		args.CancelRequestID = &request.ExecutionInfo.CancelRequestID
	}

	result, err := tx.NamedExec(updateExecutionSQLQuery, args)
	if err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("UpdateWorkflowExecution operation failed. Failed to update executions row. Erorr: %v", err),
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("UpdateWorkflowExecution operation failed. Failed to verify number of rows affected. Erorr: %v", err),
		}
	}
	if rowsAffected != 1 {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("UpdateWorkflowExecution operation failed. %v rows updated instead of 1.", rowsAffected),
		}
	}

	if err := tx.Commit(); err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("UpdateWorkflowExecution operation failed. Failed to commit transaction. Error: %v", err),
		}
	}

	return nil
}

func (*sqlMatchingManager) ResetMutableState(request *persistence.ResetMutableStateRequest) error {
	return nil
}

func (m *sqlMatchingManager) DeleteWorkflowExecution(request *persistence.DeleteWorkflowExecutionRequest) error {
	if _, err := m.db.Exec(deleteExecutionSQLQuery, m.shardID, request.DomainID, request.WorkflowID, request.RunID); err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("DeleteWorkflowExecution operation failed. Error: %v", err),
		}
	}
	return nil
}

func (m *sqlMatchingManager) GetCurrentExecution(request *persistence.GetCurrentExecutionRequest) (*persistence.GetCurrentExecutionResponse, error) {
	var row currentExecutionRow
	if err := m.db.Get(&row, getCurrentExecutionSQLQuery, m.shardID, request.DomainID, request.WorkflowID); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("GetCurrentExecution operation failed. Error: %v", err),
		}
	}
	return &persistence.GetCurrentExecutionResponse{
		StartRequestID: row.CreateRequestID,
		RunID:          row.RunID,
		State:          int(row.State),
		CloseStatus:    int(row.CloseStatus),
	}, nil
}

func (m *sqlMatchingManager) GetTransferTasks(request *persistence.GetTransferTasksRequest) (*persistence.GetTransferTasksResponse, error) {
	var resp persistence.GetTransferTasksResponse
	if err := m.db.Select(&resp.Tasks,
		getTransferTasksSQLQuery,
		request.ReadLevel,
		request.MaxReadLevel); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("GetTransferTasks operation failed. Select failed. Error: %v", err),
		}
	}

	return &resp, nil
}

func (m *sqlMatchingManager) CompleteTransferTask(request *persistence.CompleteTransferTaskRequest) error {
	if _, err := m.db.NamedExec(completeTransferTaskSQLQuery, struct {
		ShardID int   `db:"shard_id"`
		TaskID  int64 `db:"task_id"`
	}{m.shardID, request.TaskID}); err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("CompleteTransferTask operation failed. Error: %v", err),
		}
	}
	return nil
}

func (*sqlMatchingManager) GetReplicationTasks(request *persistence.GetReplicationTasksRequest) (*persistence.GetReplicationTasksResponse, error) {
	return &persistence.GetReplicationTasksResponse{}, nil
}

func (*sqlMatchingManager) CompleteReplicationTask(request *persistence.CompleteReplicationTaskRequest) error {
	return nil
}

func (*sqlMatchingManager) GetTimerIndexTasks(request *persistence.GetTimerIndexTasksRequest) (*persistence.GetTimerIndexTasksResponse, error) {
	return &persistence.GetTimerIndexTasksResponse{}, nil
}

func (*sqlMatchingManager) CompleteTimerTask(request *persistence.CompleteTimerTaskRequest) error {
	return nil
}

func NewSqlMatchingPersistence(username, password, host, port, dbName string, logger bark.Logger) (persistence.ExecutionManager, error) {
	var db, err = sqlx.Connect("mysql",
		fmt.Sprintf(Dsn, username, password, host, port, dbName))
	if err != nil {
		return nil, err
	}

	return &sqlMatchingManager{
		db: db,
		logger: logger,
	}, nil
}

func getCurrentExecutionIfExists(tx *sqlx.Tx, shardID int, domainID string, workflowID string) (*currentExecutionRow, error) {
	var row currentExecutionRow
	if err := tx.Get(&row, getCurrentExecutionSQLQuery, shardID, domainID, workflowID); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("Failed to get current_executions row for (shard,domain,workflow) = (%v, %v, %v). Error: %v", shardID, domainID, workflowID, err),
		}
	}
	return &row, nil
}

func createExecution(tx *sqlx.Tx, request *persistence.CreateWorkflowExecutionRequest, shardID int) error {
	nowTimestamp := time.Now()
	_, err := tx.NamedExec(createExecutionWithNoParentSQLQuery, executionRow{
		ShardID:                    shardID,
		DomainID:                   request.DomainID,
		WorkflowID:                 *request.Execution.WorkflowId,
		RunID:                      *request.Execution.RunId,
		TaskList:                   request.TaskList,
		WorkflowTypeName:           request.WorkflowTypeName,
		WorkflowTimeoutSeconds:     int64(request.WorkflowTimeout),
		DecisionTaskTimeoutMinutes: int64(request.DecisionTimeoutValue),
		State:                        persistence.WorkflowStateCreated,
		CloseStatus:                  persistence.WorkflowCloseStatusNone,
		LastFirstEventID:             common.FirstEventID,
		NextEventID:                  request.NextEventID,
		LastProcessedEvent:           request.LastProcessedEvent,
		StartTime:                    nowTimestamp,
		LastUpdatedTime:              nowTimestamp,
		CreateRequestID:              request.RequestID,
		DecisionVersion:              int64(request.DecisionVersion),
		DecisionScheduleID:           int64(request.DecisionScheduleID),
		DecisionStartedID:            int64(request.DecisionStartedID),
		DecisionTimeout:              int64(request.DecisionStartToCloseTimeout),
		DecisionAttempt:              0,
		DecisionTimestamp:            0,
		StickyTaskList:               "",
		StickyScheduleToStartTimeout: 0,
		ClientLibraryVersion:         "",
		ClientFeatureVersion:         "",
		ClientImpl:                   "",
	})
	if err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateWorkflowExecution operation failed. Failed to insert into executions table. Error: %v", err),
		}
	}
	return nil
}

func createCurrentExecution(tx *sqlx.Tx, request *persistence.CreateWorkflowExecutionRequest, shardID int) error {
	arg := currentExecutionRow{
		ShardID:         int64(shardID),
		DomainID:        request.DomainID,
		WorkflowID:      *request.Execution.WorkflowId,
		RunID:           *request.Execution.RunId,
		CreateRequestID: request.RequestID,
		State:           persistence.WorkflowStateRunning,
		CloseStatus:     persistence.WorkflowCloseStatusNone,
	}
	if request.ReplicationState != nil {
		arg.StartVersion = &request.ReplicationState.StartVersion
	}

	if _, err := tx.NamedExec(createCurrentExecutionSQLQuery, arg); err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateWorkflowExecution operation failed. Failed to insert into current_executions table. Error: %v", err),
		}
	}
	return nil
}

func (m *sqlMatchingManager) createTransferTasks(tx *sqlx.Tx, transferTasks []persistence.Task, shardID int, domainID, workflowID, runID string) error {
	transferTaskRows := make([]struct {
		persistence.TransferTaskInfo
		ShardID int `db:"shard_id"`
	}, len(transferTasks))

	for i, task := range transferTasks {
		transferTaskRows[i].ShardID = shardID
		transferTaskRows[i].DomainID = domainID
		transferTaskRows[i].WorkflowID = workflowID
		transferTaskRows[i].RunID = runID
		transferTaskRows[i].TargetDomainID = domainID
		transferTaskRows[i].TargetWorkflowID = persistence.TransferTaskTransferTargetWorkflowID
		transferTaskRows[i].TargetChildWorkflowOnly = false
		transferTaskRows[i].TaskList = ""
		transferTaskRows[i].ScheduleID = 0

		switch task.GetType() {
		case persistence.TransferTaskTypeActivityTask:
			transferTaskRows[i].TargetDomainID = task.(*persistence.ActivityTask).DomainID
			transferTaskRows[i].TaskList = task.(*persistence.ActivityTask).TaskList
			transferTaskRows[i].ScheduleID = task.(*persistence.ActivityTask).ScheduleID

		case persistence.TransferTaskTypeDecisionTask:
			transferTaskRows[i].TargetDomainID = task.(*persistence.DecisionTask).DomainID
			transferTaskRows[i].TaskList = task.(*persistence.DecisionTask).TaskList
			transferTaskRows[i].ScheduleID = task.(*persistence.DecisionTask).ScheduleID

		case persistence.TransferTaskTypeCancelExecution:
			transferTaskRows[i].TargetDomainID = task.(*persistence.CancelExecutionTask).TargetDomainID
			transferTaskRows[i].TargetWorkflowID = task.(*persistence.CancelExecutionTask).TargetWorkflowID
			if task.(*persistence.CancelExecutionTask).TargetRunID != "" {
				transferTaskRows[i].TargetRunID = task.(*persistence.CancelExecutionTask).TargetRunID
			}
			transferTaskRows[i].TargetChildWorkflowOnly = task.(*persistence.CancelExecutionTask).TargetChildWorkflowOnly
			transferTaskRows[i].ScheduleID = task.(*persistence.CancelExecutionTask).InitiatedID

		case persistence.TransferTaskTypeSignalExecution:
			transferTaskRows[i].TargetDomainID = task.(*persistence.SignalExecutionTask).TargetDomainID
			transferTaskRows[i].TargetWorkflowID = task.(*persistence.SignalExecutionTask).TargetWorkflowID
			if task.(*persistence.SignalExecutionTask).TargetRunID != "" {
				transferTaskRows[i].TargetRunID = task.(*persistence.SignalExecutionTask).TargetRunID
			}
			transferTaskRows[i].TargetChildWorkflowOnly = task.(*persistence.SignalExecutionTask).TargetChildWorkflowOnly
			transferTaskRows[i].ScheduleID = task.(*persistence.SignalExecutionTask).InitiatedID

		case persistence.TransferTaskTypeStartChildExecution:
			transferTaskRows[i].TargetDomainID = task.(*persistence.StartChildExecutionTask).TargetDomainID
			transferTaskRows[i].TargetWorkflowID = task.(*persistence.StartChildExecutionTask).TargetWorkflowID
			transferTaskRows[i].ScheduleID = task.(*persistence.StartChildExecutionTask).InitiatedID

		case persistence.TransferTaskTypeCloseExecution:
			// No explicit property needs to be set

		default:
			m.logger.Fatal("Unknown Transfer Task.")
		}

		transferTaskRows[i].TaskID = task.GetTaskID()
		transferTaskRows[i].TaskType = task.GetType()
		transferTaskRows[i].Version = task.GetVersion()
	}

	query, args, err := m.db.BindNamed(createTransferTasksSQLQuery, transferTaskRows)
	if err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("Failed to create transfer tasks. Failed to bind query. Error: %v", err),
		}
	}

	result, err := tx.Exec(query, args...)
	if err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("Failed to create transfer tasks. Error: %v", err),
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("Failed to create transfer tasks. Could not verify number of rows inserted.Error: %v", err),
		}
	}

	if int(rowsAffected) != len(transferTasks) {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("Failed to create transfer tasks. Inserted %v instead of %v rows into transfer_tasks. Error: %v", rowsAffected, len(transferTasks), err),
		}
	}

	return nil
}

func lockAndCheckNextEventID(tx *sqlx.Tx, shardID int, domainID, workflowID, runID string, condition int64) error {
	var nextEventID int64
	if err := tx.Get(&nextEventID, lockAndCheckNextEventIdSQLQuery, shardID, domainID, workflowID, runID); err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("Failed to lock executions row to check next_event_id. Error: %v", err),
		}
	}
	if nextEventID != condition {
		return &persistence.ConditionFailedError{
			Msg: fmt.Sprintf("next_event_id was %v when it should have been %v.", nextEventID, condition),
		}
	}
	return nil
}
