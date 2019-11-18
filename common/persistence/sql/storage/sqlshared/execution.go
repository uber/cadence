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

package sqlshared

import (
	"database/sql"

	"github.com/uber/cadence/common/persistence/sql/storage/sqldb"
)

// InsertIntoExecutions inserts a row into executions table
func (mdb *DB) InsertIntoExecutions(row *sqldb.ExecutionsRow) (sql.Result, error) {
	return mdb.conn.NamedExec(mdb.driver.CreateExecutionQuery(), row)
}

// UpdateExecutions updates a single row in executions table
func (mdb *DB) UpdateExecutions(row *sqldb.ExecutionsRow) (sql.Result, error) {
	return mdb.conn.NamedExec(mdb.driver.UpdateExecutionQuery(), row)
}

// SelectFromExecutions reads a single row from executions table
func (mdb *DB) SelectFromExecutions(filter *sqldb.ExecutionsFilter) (*sqldb.ExecutionsRow, error) {
	var row sqldb.ExecutionsRow
	err := mdb.conn.Get(&row, mdb.driver.GetExecutionQuery(), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
	if err != nil {
		return nil, err
	}
	return &row, err
}

// DeleteFromExecutions deletes a single row from executions table
func (mdb *DB) DeleteFromExecutions(filter *sqldb.ExecutionsFilter) (sql.Result, error) {
	return mdb.conn.Exec(mdb.driver.DeleteExecutionQuery(), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
}

// ReadLockExecutions acquires a write lock on a single row in executions table
func (mdb *DB) ReadLockExecutions(filter *sqldb.ExecutionsFilter) (int, error) {
	var nextEventID int
	err := mdb.conn.Get(&nextEventID, mdb.driver.ReadLockExecutionQuery(), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
	return nextEventID, err
}

// WriteLockExecutions acquires a write lock on a single row in executions table
func (mdb *DB) WriteLockExecutions(filter *sqldb.ExecutionsFilter) (int, error) {
	var nextEventID int
	err := mdb.conn.Get(&nextEventID, mdb.driver.WriteLockExecutionQuery(), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
	return nextEventID, err
}

// InsertIntoCurrentExecutions inserts a single row into current_executions table
func (mdb *DB) InsertIntoCurrentExecutions(row *sqldb.CurrentExecutionsRow) (sql.Result, error) {
	return mdb.conn.NamedExec(mdb.driver.CreateCurrentExecutionQuery(), row)
}

// UpdateCurrentExecutions updates a single row in current_executions table
func (mdb *DB) UpdateCurrentExecutions(row *sqldb.CurrentExecutionsRow) (sql.Result, error) {
	return mdb.conn.NamedExec(mdb.driver.UpdateCurrentExecutionsQuery(), row)
}

// SelectFromCurrentExecutions reads one or more rows from current_executions table
func (mdb *DB) SelectFromCurrentExecutions(filter *sqldb.CurrentExecutionsFilter) (*sqldb.CurrentExecutionsRow, error) {
	var row sqldb.CurrentExecutionsRow
	err := mdb.conn.Get(&row, mdb.driver.GetCurrentExecutionQuery(), filter.ShardID, filter.DomainID, filter.WorkflowID)
	return &row, err
}

// DeleteFromCurrentExecutions deletes a single row in current_executions table
func (mdb *DB) DeleteFromCurrentExecutions(filter *sqldb.CurrentExecutionsFilter) (sql.Result, error) {
	return mdb.conn.Exec(mdb.driver.DeleteCurrentExecutionQuery(), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
}

// LockCurrentExecutions acquires a write lock on a single row in current_executions table
func (mdb *DB) LockCurrentExecutions(filter *sqldb.CurrentExecutionsFilter) (*sqldb.CurrentExecutionsRow, error) {
	var row sqldb.CurrentExecutionsRow
	err := mdb.conn.Get(&row, mdb.driver.LockCurrentExecutionQuery(), filter.ShardID, filter.DomainID, filter.WorkflowID)
	return &row, err
}

// LockCurrentExecutionsJoinExecutions joins a row in current_executions with executions table and acquires a
// write lock on the result
func (mdb *DB) LockCurrentExecutionsJoinExecutions(filter *sqldb.CurrentExecutionsFilter) ([]sqldb.CurrentExecutionsRow, error) {
	var rows []sqldb.CurrentExecutionsRow
	err := mdb.conn.Select(&rows, mdb.driver.LockCurrentExecutionJoinExecutionsQuery(), filter.ShardID, filter.DomainID, filter.WorkflowID)
	return rows, err
}

// InsertIntoTransferTasks inserts one or more rows into transfer_tasks table
func (mdb *DB) InsertIntoTransferTasks(rows []sqldb.TransferTasksRow) (sql.Result, error) {
	return mdb.conn.NamedExec(mdb.driver.CreateTransferTasksQuery(), rows)
}

// SelectFromTransferTasks reads one or more rows from transfer_tasks table
func (mdb *DB) SelectFromTransferTasks(filter *sqldb.TransferTasksFilter) ([]sqldb.TransferTasksRow, error) {
	var rows []sqldb.TransferTasksRow
	err := mdb.conn.Select(&rows, mdb.driver.GetTransferTasksQuery(), filter.ShardID, *filter.MinTaskID, *filter.MaxTaskID)
	if err != nil {
		return nil, err
	}
	return rows, err
}

// DeleteFromTransferTasks deletes one or more rows from transfer_tasks table
func (mdb *DB) DeleteFromTransferTasks(filter *sqldb.TransferTasksFilter) (sql.Result, error) {
	if filter.MinTaskID != nil {
		return mdb.conn.Exec(mdb.driver.RangeDeleteTransferTaskQuery(), filter.ShardID, *filter.MinTaskID, *filter.MaxTaskID)
	}
	return mdb.conn.Exec(mdb.driver.DeleteTransferTaskQuery(), filter.ShardID, *filter.TaskID)
}

// InsertIntoTimerTasks inserts one or more rows into timer_tasks table
func (mdb *DB) InsertIntoTimerTasks(rows []sqldb.TimerTasksRow) (sql.Result, error) {
	for i := range rows {
		rows[i].VisibilityTimestamp = mdb.converter.ToMySQLDateTime(rows[i].VisibilityTimestamp)
	}
	return mdb.conn.NamedExec(mdb.driver.CreateTimerTasksQuery(), rows)
}

// SelectFromTimerTasks reads one or more rows from timer_tasks table
func (mdb *DB) SelectFromTimerTasks(filter *sqldb.TimerTasksFilter) ([]sqldb.TimerTasksRow, error) {
	var rows []sqldb.TimerTasksRow
	*filter.MinVisibilityTimestamp = mdb.converter.ToMySQLDateTime(*filter.MinVisibilityTimestamp)
	*filter.MaxVisibilityTimestamp = mdb.converter.ToMySQLDateTime(*filter.MaxVisibilityTimestamp)
	err := mdb.conn.Select(&rows, mdb.driver.GetTimerTasksQuery(), filter.ShardID, *filter.MinVisibilityTimestamp,
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
func (mdb *DB) DeleteFromTimerTasks(filter *sqldb.TimerTasksFilter) (sql.Result, error) {
	if filter.MinVisibilityTimestamp != nil {
		*filter.MinVisibilityTimestamp = mdb.converter.ToMySQLDateTime(*filter.MinVisibilityTimestamp)
		*filter.MaxVisibilityTimestamp = mdb.converter.ToMySQLDateTime(*filter.MaxVisibilityTimestamp)
		return mdb.conn.Exec(mdb.driver.RangeDeleteTimerTaskQuery(), filter.ShardID, *filter.MinVisibilityTimestamp, *filter.MaxVisibilityTimestamp)
	}
	*filter.VisibilityTimestamp = mdb.converter.ToMySQLDateTime(*filter.VisibilityTimestamp)
	return mdb.conn.Exec(mdb.driver.DeleteTimerTaskQuery(), filter.ShardID, *filter.VisibilityTimestamp, filter.TaskID)
}

// InsertIntoBufferedEvents inserts one or more rows into buffered_events table
func (mdb *DB) InsertIntoBufferedEvents(rows []sqldb.BufferedEventsRow) (sql.Result, error) {
	return mdb.conn.NamedExec(mdb.driver.CreateBufferedEventsQuery(), rows)
}

// SelectFromBufferedEvents reads one or more rows from buffered_events table
func (mdb *DB) SelectFromBufferedEvents(filter *sqldb.BufferedEventsFilter) ([]sqldb.BufferedEventsRow, error) {
	var rows []sqldb.BufferedEventsRow
	err := mdb.conn.Select(&rows, mdb.driver.GetBufferedEventsQuery(), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
	for i := 0; i < len(rows); i++ {
		rows[i].DomainID = filter.DomainID
		rows[i].WorkflowID = filter.WorkflowID
		rows[i].RunID = filter.RunID
		rows[i].ShardID = filter.ShardID
	}
	return rows, err
}

// DeleteFromBufferedEvents deletes one or more rows from buffered_events table
func (mdb *DB) DeleteFromBufferedEvents(filter *sqldb.BufferedEventsFilter) (sql.Result, error) {
	return mdb.conn.Exec(mdb.driver.DeleteBufferedEventsQuery(), filter.ShardID, filter.DomainID, filter.WorkflowID, filter.RunID)
}

// InsertIntoReplicationTasks inserts one or more rows into replication_tasks table
func (mdb *DB) InsertIntoReplicationTasks(rows []sqldb.ReplicationTasksRow) (sql.Result, error) {
	return mdb.conn.NamedExec(mdb.driver.CreateReplicationTasksQuery(), rows)
}

// SelectFromReplicationTasks reads one or more rows from replication_tasks table
func (mdb *DB) SelectFromReplicationTasks(filter *sqldb.ReplicationTasksFilter) ([]sqldb.ReplicationTasksRow, error) {
	var rows []sqldb.ReplicationTasksRow
	err := mdb.conn.Select(&rows, mdb.driver.GetReplicationTasksQuery(), filter.ShardID, filter.MinTaskID, filter.MaxTaskID, filter.PageSize)
	return rows, err
}

// DeleteFromReplicationTasks deletes one or more rows from replication_tasks table
func (mdb *DB) DeleteFromReplicationTasks(shardID, taskID int) (sql.Result, error) {
	return mdb.conn.Exec(mdb.driver.DeleteReplicationTaskQuery(), shardID, taskID)
}

// InsertIntoReplicationTasksDLQ inserts one or more rows into replication_tasks_dlq table
func (mdb *DB) InsertIntoReplicationTasksDLQ(row *sqldb.ReplicationTaskDLQRow) (sql.Result, error) {
	return mdb.conn.NamedExec(mdb.driver.InsertReplicationTaskDLQQuery(), row)
}

// SelectFromReplicationTasksDLQ reads one or more rows from replication_tasks_dlq table
func (mdb *DB) SelectFromReplicationTasksDLQ(filter *sqldb.ReplicationTasksDLQFilter) ([]sqldb.ReplicationTasksRow, error) {
	var rows []sqldb.ReplicationTasksRow
	err := mdb.conn.Select(
		&rows, mdb.driver.GetReplicationTasksDLQQuery(),
		filter.SourceClusterName,
		filter.ShardID,
		filter.MinTaskID,
		filter.MaxTaskID,
		filter.PageSize)
	return rows, err
}
