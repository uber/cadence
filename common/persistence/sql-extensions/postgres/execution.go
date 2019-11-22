// Copyright (c) 2019 Uber Technologies, Inc.
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

package postgres

const (
	executionsColumns = `shard_id, domain_id, workflow_id, run_id, next_event_id, last_write_version, data, data_encoding`

	createExecutionQuery = `INSERT INTO executions(` + executionsColumns + `)
 VALUES(:shard_id, :domain_id, :workflow_id, :run_id, :next_event_id, :last_write_version, :data, :data_encoding)`

	updateExecutionQuery = `UPDATE executions SET
 next_event_id = :next_event_id, last_write_version = :last_write_version, data = :data, data_encoding = :data_encoding
 WHERE shard_id = :shard_id AND domain_id = :domain_id AND workflow_id = :workflow_id AND run_id = :run_id`

	getExecutionQuery = `SELECT ` + executionsColumns + ` FROM executions
 WHERE shard_id = ? AND domain_id = ? AND workflow_id = ? AND run_id = ?`

	deleteExecutionQuery = `DELETE FROM executions 
 WHERE shard_id = ? AND domain_id = ? AND workflow_id = ? AND run_id = ?`

	lockExecutionQueryBase = `SELECT next_event_id FROM executions 
 WHERE shard_id = ? AND domain_id = ? AND workflow_id = ? AND run_id = ?`

	writeLockExecutionQuery = lockExecutionQueryBase + ` FOR UPDATE`
	readLockExecutionQuery  = lockExecutionQueryBase + ` LOCK IN SHARE MODE`

	createCurrentExecutionQuery = `INSERT INTO current_executions
(shard_id, domain_id, workflow_id, run_id, create_request_id, state, close_status, start_version, last_write_version) VALUES
(:shard_id, :domain_id, :workflow_id, :run_id, :create_request_id, :state, :close_status, :start_version, :last_write_version)`

	deleteCurrentExecutionQuery = "DELETE FROM current_executions WHERE shard_id=? AND domain_id=? AND workflow_id=? AND run_id=?"

	getCurrentExecutionQuery = `SELECT
shard_id, domain_id, workflow_id, run_id, create_request_id, state, close_status, start_version, last_write_version
FROM current_executions WHERE shard_id = ? AND domain_id = ? AND workflow_id = ?`

	lockCurrentExecutionJoinExecutionsQuery = `SELECT
ce.shard_id, ce.domain_id, ce.workflow_id, ce.run_id, ce.create_request_id, ce.state, ce.close_status, ce.start_version, e.last_write_version
FROM current_executions ce
INNER JOIN executions e ON e.shard_id = ce.shard_id AND e.domain_id = ce.domain_id AND e.workflow_id = ce.workflow_id AND e.run_id = ce.run_id
WHERE ce.shard_id = ? AND ce.domain_id = ? AND ce.workflow_id = ? FOR UPDATE`

	lockCurrentExecutionQuery = getCurrentExecutionQuery + ` FOR UPDATE`

	updateCurrentExecutionsQuery = `UPDATE current_executions SET
run_id = :run_id,
create_request_id = :create_request_id,
state = :state,
close_status = :close_status,
start_version = :start_version,
last_write_version = :last_write_version
WHERE
shard_id = :shard_id AND
domain_id = :domain_id AND
workflow_id = :workflow_id
`

	getTransferTasksQuery = `SELECT task_id, data, data_encoding 
 FROM transfer_tasks WHERE shard_id = ? AND task_id > ? AND task_id <= ?`

	createTransferTasksQuery = `INSERT INTO transfer_tasks(shard_id, task_id, data, data_encoding) 
 VALUES(:shard_id, :task_id, :data, :data_encoding)`

	deleteTransferTaskQuery      = `DELETE FROM transfer_tasks WHERE shard_id = ? AND task_id = ?`
	rangeDeleteTransferTaskQuery = `DELETE FROM transfer_tasks WHERE shard_id = ? AND task_id > ? AND task_id <= ?`

	createTimerTasksQuery = `INSERT INTO timer_tasks (shard_id, visibility_timestamp, task_id, data, data_encoding)
  VALUES (:shard_id, :visibility_timestamp, :task_id, :data, :data_encoding)`

	getTimerTasksQuery = `SELECT visibility_timestamp, task_id, data, data_encoding FROM timer_tasks 
  WHERE shard_id = ? 
  AND ((visibility_timestamp >= ? AND task_id >= ?) OR visibility_timestamp > ?) 
  AND visibility_timestamp < ?
  ORDER BY visibility_timestamp,task_id LIMIT ?`

	deleteTimerTaskQuery      = `DELETE FROM timer_tasks WHERE shard_id = ? AND visibility_timestamp = ? AND task_id = ?`
	rangeDeleteTimerTaskQuery = `DELETE FROM timer_tasks WHERE shard_id = ? AND visibility_timestamp >= ? AND visibility_timestamp < ?`

	createReplicationTasksQuery = `INSERT INTO replication_tasks (shard_id, task_id, data, data_encoding) 
  VALUES(:shard_id, :task_id, :data, :data_encoding)`

	getReplicationTasksQuery = `SELECT task_id, data, data_encoding FROM replication_tasks WHERE 
shard_id = ? AND
task_id > ? AND
task_id <= ? 
ORDER BY task_id LIMIT ?`

	deleteReplicationTaskQuery = `DELETE FROM replication_tasks WHERE shard_id = ? AND task_id = ?`

	getReplicationTasksDLQQuery = `SELECT task_id, data, data_encoding FROM replication_tasks_dlq WHERE 
source_cluster_name = ? AND
shard_id = ? AND
task_id > ? AND
task_id <= ?
ORDER BY task_id LIMIT ?`

	bufferedEventsColumns     = `shard_id, domain_id, workflow_id, run_id, data, data_encoding`
	createBufferedEventsQuery = `INSERT INTO buffered_events(` + bufferedEventsColumns + `)
VALUES (:shard_id, :domain_id, :workflow_id, :run_id, :data, :data_encoding)`

	deleteBufferedEventsQuery = `DELETE FROM buffered_events WHERE shard_id=? AND domain_id=? AND workflow_id=? AND run_id=?`
	getBufferedEventsQuery    = `SELECT data, data_encoding FROM buffered_events WHERE
shard_id=? AND domain_id=? AND workflow_id=? AND run_id=?`

	insertReplicationTaskDLQQuery = `
INSERT INTO replication_tasks_dlq 
            (source_cluster_name, 
             shard_id, 
             task_id, 
             data, 
             data_encoding) 
VALUES     (:source_cluster_name, 
            :shard_id, 
            :task_id, 
            :data, 
            :data_encoding)
`
)

func (d *driver) CreateExecutionQuery() string {
	return createExecutionQuery
}

func (d *driver) UpdateExecutionQuery() string {
	return updateExecutionQuery
}

func (d *driver) GetExecutionQuery() string {
	return getExecutionQuery
}

func (d *driver) DeleteExecutionQuery() string {
	return deleteExecutionQuery
}

func (d *driver) WriteLockExecutionQuery() string {
	return writeLockExecutionQuery
}

func (d *driver) ReadLockExecutionQuery() string {
	return readLockExecutionQuery
}

func (d *driver) CreateCurrentExecutionQuery() string {
	return createCurrentExecutionQuery
}

func (d *driver) DeleteCurrentExecutionQuery() string {
	return deleteCurrentExecutionQuery
}

func (d *driver) GetCurrentExecutionQuery() string {
	return getCurrentExecutionQuery
}

func (d *driver) LockCurrentExecutionJoinExecutionsQuery() string {
	return lockCurrentExecutionJoinExecutionsQuery
}

func (d *driver) LockCurrentExecutionQuery() string {
	return lockCurrentExecutionQuery
}

func (d *driver) UpdateCurrentExecutionsQuery() string {
	return updateCurrentExecutionsQuery
}

func (d *driver) GetTransferTasksQuery() string {
	return getTransferTasksQuery
}

func (d *driver) CreateTransferTasksQuery() string {
	return createTransferTasksQuery
}

func (d *driver) DeleteTransferTaskQuery() string {
	return deleteTransferTaskQuery
}

func (d *driver) RangeDeleteTransferTaskQuery() string {
	return rangeDeleteTransferTaskQuery
}

func (d *driver) CreateTimerTasksQuery() string {
	return createTimerTasksQuery
}

func (d *driver) GetTimerTasksQuery() string {
	return getTimerTasksQuery
}

func (d *driver) DeleteTimerTaskQuery() string {
	return deleteTimerTaskQuery
}

func (d *driver) RangeDeleteTimerTaskQuery() string {
	return rangeDeleteTimerTaskQuery
}

func (d *driver) CreateReplicationTasksQuery() string {
	return createReplicationTasksQuery
}

func (d *driver) GetReplicationTasksQuery() string {
	return getReplicationTasksQuery
}

func (d *driver) DeleteReplicationTaskQuery() string {
	return deleteReplicationTaskQuery
}

func (d *driver) GetReplicationTasksDLQQuery() string {
	return getReplicationTasksDLQQuery
}

func (d *driver) CreateBufferedEventsQuery() string {
	return createBufferedEventsQuery
}

func (d *driver) DeleteBufferedEventsQuery() string {
	return deleteBufferedEventsQuery
}

func (d *driver) GetBufferedEventsQuery() string {
	return getBufferedEventsQuery
}

func (d *driver) InsertReplicationTaskDLQQuery() string {
	return insertReplicationTaskDLQQuery
}
