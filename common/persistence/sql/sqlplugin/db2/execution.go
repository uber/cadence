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

package db2

import (
	"database/sql"
	"fmt"

	"github.com/uber/cadence/common/persistence/sql/sqlplugin"
)

const (
	executionsColumns = `shard_id, domain_id, workflow_id, run_id, next_event_id, last_write_version, data, data_encoding`

	createExecutionQuery = `INSERT INTO cadence.executions(` + executionsColumns + `)
 VALUES(:SHARD_ID, :DOMAIN_ID, :WORKFLOW_ID, :RUN_ID, :NEXT_EVENT_ID, :LAST_WRITE_VERSION, :DATA, :DATA_ENCODING)`

	updateExecutionQuery = `UPDATE cadence.executions SET
 next_event_id = :NEXT_EVENT_ID, last_write_version = :LAST_WRITE_VERSION, data = :DATA, data_encoding = :DATA_ENCODING
 WHERE shard_id = :SHARD_ID AND domain_id = :DOMAIN_ID AND workflow_id = :WORKFLOW_ID AND run_id = :RUN_ID`

	getExecutionQuery = `SELECT ` + executionsColumns + ` FROM cadence.executions
 WHERE shard_id = ? AND domain_id = ? AND workflow_id = ? AND run_id = ?`

	deleteExecutionQuery = `DELETE FROM cadence.executions 
 WHERE shard_id = ? AND domain_id = ? AND workflow_id = ? AND run_id = ?`

	lockExecutionQueryBase = `SELECT next_event_id FROM cadence.executions 
 WHERE shard_id = ? AND domain_id = ? AND workflow_id = ? AND run_id = ?`

	writeLockExecutionQuery = lockExecutionQueryBase + ` FOR UPDATE`
	readLockExecutionQuery  = lockExecutionQueryBase + ` LOCK IN SHARE MODE`

	createCurrentExecutionQuery = `INSERT INTO cadence.current_executions
(shard_id, domain_id, workflow_id, run_id, create_request_id, state, close_status, start_version, last_write_version) VALUES
(:SHARD_ID, :DOMAIN_ID, :WORKFLOW_ID, :RUN_ID, :CREATE_REQUEST_ID, :STATE, :CLOSE_STATUS, :START_VERSION, :LAST_WRITE_VERSION)`

	deleteCurrentExecutionQuery = "DELETE FROM cadence.current_executions WHERE shard_id=? AND domain_id=? AND workflow_id=? AND run_id=?"

	getCurrentExecutionQuery = `SELECT
shard_id, domain_id, workflow_id, run_id, create_request_id, state, close_status, start_version, last_write_version
FROM cadence.current_executions WHERE shard_id = ? AND domain_id = ? AND workflow_id = ?`

	lockCurrentExecutionJoinExecutionsQuery = `SELECT
	ce.shard_id, ce.domain_id, ce.workflow_id, ce.run_id, ce.create_request_id, ce.state, 
	ce.close_status, ce.start_version, 
	(SELECT LAST_WRITE_VERSION FROM cadence.executions e WHERE e.shard_id = ce.shard_id AND e.domain_id = ce.domain_id AND e.workflow_id = ce.workflow_id AND e.run_id = ce.run_id ) last_write_version
	FROM cadence.current_executions ce 
	WHERE ce.shard_id = ? AND ce.domain_id = ? AND ce.workflow_id = ? 
	AND EXISTS (SELECT * FROM cadence.executions e WHERE e.shard_id = ce.shard_id AND e.domain_id = ce.domain_id AND e.workflow_id = ce.workflow_id AND e.run_id = ce.run_id) 
	FOR UPDATE`

	lockCurrentExecutionQuery = getCurrentExecutionQuery + ` FOR UPDATE`

	updateCurrentExecutionsQuery = `UPDATE cadence.current_executions SET
run_id = :RUN_ID,
create_request_id = :CREATE_REQUEST_ID,
state = :STATE,
close_status = :CLOSE_STATUS,
start_version = :START_VERSION,
last_write_version = :LAST_WRITE_VERSION
WHERE
shard_id = :SHARD_ID AND
domain_id = :DOMAIN_ID AND
workflow_id = :WORKFLOW_ID
`

	getTransferTasksQuery = `SELECT task_id, data, data_encoding 
 FROM cadence.transfer_tasks WHERE shard_id = ? AND task_id > ? AND task_id <= ? ORDER BY shard_id, task_id`

	createTransferTasksQuery = `INSERT INTO cadence.transfer_tasks(shard_id, task_id, data, data_encoding) 
 VALUES(:SHARD_ID, :TASK_ID, :DATA, :DATA_ENCODING)`

	deleteTransferTaskQuery      = `DELETE FROM cadence.transfer_tasks WHERE shard_id = ? AND task_id = ?`
	rangeDeleteTransferTaskQuery = `DELETE FROM cadence.transfer_tasks WHERE shard_id = ? AND task_id > ? AND task_id <= ?`

	createTimerTasksQuery = `INSERT INTO cadence.timer_tasks (shard_id, visibility_timestamp, task_id, data, data_encoding)
  VALUES (:SHARD_ID, :VISIBILITY_TIMESTAMP, :TASK_ID, :DATA, :DATA_ENCODING)`

	getTimerTasksQuery = `SELECT visibility_timestamp, task_id, data, data_encoding FROM cadence.timer_tasks 
  WHERE shard_id = ? 
  AND ((visibility_timestamp >= ? AND task_id >= ?) OR visibility_timestamp > ?) 
  AND visibility_timestamp < ?
  ORDER BY visibility_timestamp,task_id LIMIT ?`

	deleteTimerTaskQuery      = `DELETE FROM cadence.timer_tasks WHERE shard_id = ? AND visibility_timestamp = ? AND task_id = ?`
	rangeDeleteTimerTaskQuery = `DELETE FROM cadence.timer_tasks WHERE shard_id = ? AND visibility_timestamp >= ? AND visibility_timestamp < ?`

	createReplicationTasksQuery = `INSERT INTO cadence.replication_tasks (shard_id, task_id, data, data_encoding) 
  VALUES(:SHARD_ID, :TASK_ID, :DATA, :DATA_ENCODING)`

	getReplicationTasksQuery = `SELECT task_id, data, data_encoding FROM cadence.replication_tasks WHERE 
shard_id = ? AND
task_id > ? AND
task_id <= ? 
ORDER BY task_id LIMIT ?`

	deleteReplicationTaskQuery      = `DELETE FROM cadence.replication_tasks WHERE shard_id = ? AND task_id = ?`
	rangeDeleteReplicationTaskQuery = `DELETE FROM cadence.replication_tasks WHERE shard_id = ? AND task_id <= ?`

	getReplicationTasksDLQQuery = `SELECT task_id, data, data_encoding FROM cadence.replication_tasks_dlq WHERE 
source_cluster_name = ? AND
shard_id = ? AND
task_id > ? AND
task_id <= ?
ORDER BY task_id LIMIT ?`

	getReplicationTaskDLQQuery = `SELECT count(1) as count FROM cadence.replication_tasks_dlq WHERE 
source_cluster_name = ? AND
shard_id = ?`

	bufferedEventsColumns     = `shard_id, domain_id, workflow_id, run_id, data, data_encoding`
	createBufferedEventsQuery = `INSERT INTO cadence.buffered_events(` + bufferedEventsColumns + `)
VALUES (:SHARD_ID, :DOMAIN_ID, :WORKFLOW_ID, :RUN_ID, :DATA, :DATA_ENCODING)`

	deleteBufferedEventsQuery = `DELETE FROM cadence.buffered_events WHERE shard_id=? AND domain_id=? AND workflow_id=? AND run_id=?`
	getBufferedEventsQuery    = `SELECT data, data_encoding FROM cadence.buffered_events WHERE
shard_id=? AND domain_id=? AND workflow_id=? AND run_id=?`

	insertReplicationTaskDLQQuery = `
INSERT INTO cadence.replication_tasks_dlq 
            (source_cluster_name, 
             shard_id, 
             task_id, 
             data, 
             data_encoding) 
VALUES     (:SOURCE_CLUSTER_NAME, 
            :SHARD_ID, 
            :TASK_ID, 
            :DATA, 
            :DATA_ENCODING)
`
	deleteReplicationTaskFromDLQQuery = `
	DELETE FROM cadence.replication_tasks_dlq 
		WHERE source_cluster_name = ? 
		AND shard_id = ? 
		AND task_id = ?`

	rangeDeleteReplicationTaskFromDLQQuery = `
	DELETE FROM cadence.replication_tasks_dlq 
		WHERE source_cluster_name = ? 
		AND shard_id = ? 
		AND task_id > ?
		AND task_id <= ?`
)

// InsertIntoExecutions inserts a row into executions table
func (mdb *db) InsertIntoExecutions(row *sqlplugin.ExecutionsRow) (sql.Result, error) {
	return mdb.conn.NamedExec(createExecutionQuery, row)
}

// UpdateExecutions updates a single row in executions table
func (mdb *db) UpdateExecutions(row *sqlplugin.ExecutionsRow) (sql.Result, error) {
	return mdb.conn.NamedExec(updateExecutionQuery, row)
}

// SelectFromExecutions reads a single row from executions table
func (mdb *db) SelectFromExecutions(filter *sqlplugin.ExecutionsFilter) (*sqlplugin.ExecutionsRow, error) {
	var row sqlplugin.ExecutionsRow
	err := mdb.Get(&row, getExecutionQuery, filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
	if err != nil {
		return nil, err
	}
	return &row, err
}

// DeleteFromExecutions deletes a single row from executions table
func (mdb *db) DeleteFromExecutions(filter *sqlplugin.ExecutionsFilter) (sql.Result, error) {
	return mdb.conn.Exec(deleteExecutionQuery, filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
}

// ReadLockExecutions acquires a write lock on a single row in executions table
func (mdb *db) ReadLockExecutions(filter *sqlplugin.ExecutionsFilter) (int, error) {
	var nextEventID int
	fmt.Printf("ReadLockExecutions.TX: ", mdb.tx)
	err := mdb.Get(&nextEventID, readLockExecutionQuery, filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
	return nextEventID, err
}

// WriteLockExecutions acquires a write lock on a single row in executions table
func (mdb *db) WriteLockExecutions(filter *sqlplugin.ExecutionsFilter) (int, error) {
	var nextEventID int
	err := mdb.Get(&nextEventID, writeLockExecutionQuery, filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
	return nextEventID, err
}

// InsertIntoCurrentExecutions inserts a single row into current_executions table
func (mdb *db) InsertIntoCurrentExecutions(row *sqlplugin.CurrentExecutionsRow) (sql.Result, error) {
	return mdb.NamedExec(createCurrentExecutionQuery, row)
}

// UpdateCurrentExecutions updates a single row in current_executions table
func (mdb *db) UpdateCurrentExecutions(row *sqlplugin.CurrentExecutionsRow) (sql.Result, error) {
	return mdb.NamedExec(updateCurrentExecutionsQuery, row)
}

// SelectFromCurrentExecutions reads one or more rows from current_executions table
func (mdb *db) SelectFromCurrentExecutions(filter *sqlplugin.CurrentExecutionsFilter) (*sqlplugin.CurrentExecutionsRow, error) {
	var row sqlplugin.CurrentExecutionsRow
	err := mdb.Get(&row, getCurrentExecutionQuery, filter.ShardID, filter.DomainID, filter.WorkflowID)
	return &row, err
}

// DeleteFromCurrentExecutions deletes a single row in current_executions table
func (mdb *db) DeleteFromCurrentExecutions(filter *sqlplugin.CurrentExecutionsFilter) (sql.Result, error) {
	return mdb.conn.Exec(deleteCurrentExecutionQuery, filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
}

// LockCurrentExecutions acquires a write lock on a single row in current_executions table
func (mdb *db) LockCurrentExecutions(filter *sqlplugin.CurrentExecutionsFilter) (*sqlplugin.CurrentExecutionsRow, error) {
	var row sqlplugin.CurrentExecutionsRow
	err := mdb.Get(&row, lockCurrentExecutionQuery, filter.ShardID, filter.DomainID, filter.WorkflowID)
	return &row, err
}

// LockCurrentExecutionsJoinExecutions joins a row in current_executions with executions table and acquires a
// write lock on the result
func (mdb *db) LockCurrentExecutionsJoinExecutions(filter *sqlplugin.CurrentExecutionsFilter) ([]sqlplugin.CurrentExecutionsRow, error) {
	var rows []sqlplugin.CurrentExecutionsRow
	err := mdb.Select(&rows, lockCurrentExecutionJoinExecutionsQuery, filter.ShardID, filter.DomainID, filter.WorkflowID)
	return rows, err
}

// InsertIntoTransferTasks inserts one or more rows into transfer_tasks table
func (mdb *db) InsertIntoTransferTasks(rows []sqlplugin.TransferTasksRow) (sql.Result, error) {
	return mdb.conn.NamedExec(createTransferTasksQuery, rows)
}

// SelectFromTransferTasks reads one or more rows from transfer_tasks table
func (mdb *db) SelectFromTransferTasks(filter *sqlplugin.TransferTasksFilter) ([]sqlplugin.TransferTasksRow, error) {
	var rows []sqlplugin.TransferTasksRow
	err := mdb.Select(&rows, getTransferTasksQuery, filter.ShardID, *filter.MinTaskID, *filter.MaxTaskID)
	if err != nil {
		return nil, err
	}
	return rows, err
}

// DeleteFromTransferTasks deletes one or more rows from transfer_tasks table
func (mdb *db) DeleteFromTransferTasks(filter *sqlplugin.TransferTasksFilter) (sql.Result, error) {
	if filter.MinTaskID != nil {
		return mdb.conn.Exec(rangeDeleteTransferTaskQuery, filter.ShardID, *filter.MinTaskID, *filter.MaxTaskID)
	}
	return mdb.conn.Exec(deleteTransferTaskQuery, filter.ShardID, *filter.TaskID)
}

// InsertIntoTimerTasks inserts one or more rows into timer_tasks table
func (mdb *db) InsertIntoTimerTasks(rows []sqlplugin.TimerTasksRow) (sql.Result, error) {
	for i := range rows {
		rows[i].VisibilityTimestamp = mdb.converter.ToMySQLDateTime(rows[i].VisibilityTimestamp)
	}
	return mdb.conn.NamedExec(createTimerTasksQuery, rows)
}

// SelectFromTimerTasks reads one or more rows from timer_tasks table
func (mdb *db) SelectFromTimerTasks(filter *sqlplugin.TimerTasksFilter) ([]sqlplugin.TimerTasksRow, error) {
	var rows []sqlplugin.TimerTasksRow
	*filter.MinVisibilityTimestamp = mdb.converter.ToMySQLDateTime(*filter.MinVisibilityTimestamp)
	*filter.MaxVisibilityTimestamp = mdb.converter.ToMySQLDateTime(*filter.MaxVisibilityTimestamp)
	err := mdb.Select(&rows, getTimerTasksQuery, filter.ShardID, *filter.MinVisibilityTimestamp,
		filter.TaskID, *filter.MinVisibilityTimestamp, *filter.MaxVisibilityTimestamp, *filter.PageSize)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].VisibilityTimestamp = mdb.converter.FromMySQLDateTime(rows[i].VisibilityTimestamp)
	}
	return rows, err
}

// DeleteFromTimerTasks deletes one or more rows from timer_tasks table
func (mdb *db) DeleteFromTimerTasks(filter *sqlplugin.TimerTasksFilter) (sql.Result, error) {
	if filter.MinVisibilityTimestamp != nil {
		*filter.MinVisibilityTimestamp = mdb.converter.ToMySQLDateTime(*filter.MinVisibilityTimestamp)
		*filter.MaxVisibilityTimestamp = mdb.converter.ToMySQLDateTime(*filter.MaxVisibilityTimestamp)
		return mdb.conn.Exec(rangeDeleteTimerTaskQuery, filter.ShardID, *filter.MinVisibilityTimestamp, *filter.MaxVisibilityTimestamp)
	}
	*filter.VisibilityTimestamp = mdb.converter.ToMySQLDateTime(*filter.VisibilityTimestamp)
	return mdb.conn.Exec(deleteTimerTaskQuery, filter.ShardID, *filter.VisibilityTimestamp, filter.TaskID)
}

// InsertIntoBufferedEvents inserts one or more rows into buffered_events table
func (mdb *db) InsertIntoBufferedEvents(rows []sqlplugin.BufferedEventsRow) (sql.Result, error) {
	return mdb.conn.NamedExec(createBufferedEventsQuery, rows)
}

// SelectFromBufferedEvents reads one or more rows from buffered_events table
func (mdb *db) SelectFromBufferedEvents(filter *sqlplugin.BufferedEventsFilter) ([]sqlplugin.BufferedEventsRow, error) {
	var rows []sqlplugin.BufferedEventsRow
	err := mdb.Select(&rows, getBufferedEventsQuery, filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
	for i := 0; i < len(rows); i++ {
		rows[i].DomainID = filter.DomainID
		rows[i].WorkflowID = filter.WorkflowID
		rows[i].RunID = filter.RunID
		rows[i].ShardID = filter.ShardID
	}
	return rows, err
}

// DeleteFromBufferedEvents deletes one or more rows from buffered_events table
func (mdb *db) DeleteFromBufferedEvents(filter *sqlplugin.BufferedEventsFilter) (sql.Result, error) {
	return mdb.conn.Exec(deleteBufferedEventsQuery, filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
}

// InsertIntoReplicationTasks inserts one or more rows into replication_tasks table
func (mdb *db) InsertIntoReplicationTasks(rows []sqlplugin.ReplicationTasksRow) (sql.Result, error) {
	return mdb.conn.NamedExec(createReplicationTasksQuery, rows)
}

// SelectFromReplicationTasks reads one or more rows from replication_tasks table
func (mdb *db) SelectFromReplicationTasks(filter *sqlplugin.ReplicationTasksFilter) ([]sqlplugin.ReplicationTasksRow, error) {
	var rows []sqlplugin.ReplicationTasksRow
	err := mdb.Select(&rows, getReplicationTasksQuery, filter.ShardID, filter.MinTaskID, filter.MaxTaskID, filter.PageSize)
	return rows, err
}

// DeleteFromReplicationTasks deletes one row from replication_tasks table
func (mdb *db) DeleteFromReplicationTasks(filter *sqlplugin.ReplicationTasksFilter) (sql.Result, error) {
	return mdb.conn.Exec(deleteReplicationTaskQuery, filter.ShardID, filter.TaskID)
}

// RangeDeleteFromReplicationTasks deletes multi rows from replication_tasks table
func (mdb *db) RangeDeleteFromReplicationTasks(filter *sqlplugin.ReplicationTasksFilter) (sql.Result, error) {
	return mdb.conn.Exec(rangeDeleteReplicationTaskQuery, filter.ShardID, filter.InclusiveEndTaskID)
}

// InsertIntoReplicationTasksDLQ inserts one or more rows into replication_tasks_dlq table
func (mdb *db) InsertIntoReplicationTasksDLQ(row *sqlplugin.ReplicationTaskDLQRow) (sql.Result, error) {
	return mdb.conn.NamedExec(insertReplicationTaskDLQQuery, row)
}

// SelectFromReplicationTasksDLQ reads one or more rows from replication_tasks_dlq table
func (mdb *db) SelectFromReplicationTasksDLQ(filter *sqlplugin.ReplicationTasksDLQFilter) ([]sqlplugin.ReplicationTasksRow, error) {
	var rows []sqlplugin.ReplicationTasksRow
	err := mdb.Select(
		&rows, getReplicationTasksDLQQuery,
		filter.SourceClusterName,
		filter.ShardID,
		filter.MinTaskID,
		filter.MaxTaskID,
		filter.PageSize)
	return rows, err
}

// SelectFromReplicationDLQ reads one row from replication_tasks_dlq table
func (mdb *db) SelectFromReplicationDLQ(filter *sqlplugin.ReplicationTaskDLQFilter) (int64, error) {
	var size []int64
	if err := mdb.Select(
		&size, getReplicationTaskDLQQuery,
		filter.SourceClusterName,
		filter.ShardID,
	); err != nil {
		return 0, err
	}
	return size[0], nil
}

// DeleteMessageFromReplicationTasksDLQ deletes one row from replication_tasks_dlq table
func (mdb *db) DeleteMessageFromReplicationTasksDLQ(
	filter *sqlplugin.ReplicationTasksDLQFilter,
) (sql.Result, error) {

	return mdb.conn.Exec(
		deleteReplicationTaskFromDLQQuery,
		filter.SourceClusterName,
		filter.ShardID,
		filter.TaskID,
	)
}

// DeleteMessageFromReplicationTasksDLQ deletes one or more rows from replication_tasks_dlq table
func (mdb *db) RangeDeleteMessageFromReplicationTasksDLQ(
	filter *sqlplugin.ReplicationTasksDLQFilter,
) (sql.Result, error) {

	return mdb.conn.Exec(
		rangeDeleteReplicationTaskFromDLQQuery,
		filter.SourceClusterName,
		filter.ShardID,
		filter.TaskID,
		filter.InclusiveEndTaskID,
	)
}
