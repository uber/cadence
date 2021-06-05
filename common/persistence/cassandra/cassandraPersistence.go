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
	"reflect"
	"strings"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
	"github.com/uber/cadence/common/types"
)

type (
	cassandraStore struct {
		client  gocql.Client
		session gocql.Session
		logger  log.Logger
	}

	// Implements ExecutionStore
	cassandraPersistence struct {
		cassandraStore
		shardID int
		db      nosqlplugin.DB
	}
)

var _ p.ExecutionStore = (*cassandraPersistence)(nil)

// Guidelines for creating new special UUID constants
// Each UUID should be of the form: E0000000-R000-f000-f000-00000000000x
// Where x is any hexadecimal value, E represents the entity type valid values are:
// E = {DomainID = 1, WorkflowID = 2, RunID = 3}
// R represents row type in executions table, valid values are:
// R = {Shard = 1, Execution = 2, Transfer = 3, Timer = 4, Replication = 5, Replication_DLQ = 6, CrossCluster = 7}
const (
	// Special Domains related constants
	emptyDomainID = "10000000-0000-f000-f000-000000000000"
	// Special Run IDs
	emptyRunID     = "30000000-0000-f000-f000-000000000000"
	permanentRunID = "30000000-0000-f000-f000-000000000001"
	// Row Constants for Shard Row
	rowTypeShardDomainID   = "10000000-1000-f000-f000-000000000000"
	rowTypeShardWorkflowID = "20000000-1000-f000-f000-000000000000"
	rowTypeShardRunID      = "30000000-1000-f000-f000-000000000000"
	// Row Constants for Transfer Task Row
	rowTypeTransferDomainID   = "10000000-3000-f000-f000-000000000000"
	rowTypeTransferWorkflowID = "20000000-3000-f000-f000-000000000000"
	rowTypeTransferRunID      = "30000000-3000-f000-f000-000000000000"
	// Row Constants for Timer Task Row
	rowTypeTimerDomainID   = "10000000-4000-f000-f000-000000000000"
	rowTypeTimerWorkflowID = "20000000-4000-f000-f000-000000000000"
	rowTypeTimerRunID      = "30000000-4000-f000-f000-000000000000"
	// Row Constants for Replication Task Row
	rowTypeReplicationDomainID   = "10000000-5000-f000-f000-000000000000"
	rowTypeReplicationWorkflowID = "20000000-5000-f000-f000-000000000000"
	rowTypeReplicationRunID      = "30000000-5000-f000-f000-000000000000"
	// Row Constants for Replication Task DLQ Row. Source cluster name will be used as WorkflowID.
	rowTypeDLQDomainID = "10000000-6000-f000-f000-000000000000"
	rowTypeDLQRunID    = "30000000-6000-f000-f000-000000000000"
	// Row Constants for Cross Cluster Task Row
	rowTypeCrossClusterDomainID = "10000000-7000-f000-f000-000000000000"
	rowTypeCrossClusterRunID    = "30000000-7000-f000-f000-000000000000"
	// Special TaskId constants
	rowTypeExecutionTaskID = int64(-10)
	rowTypeShardTaskID     = int64(-11)
	emptyInitiatedID       = int64(-7)

	stickyTaskListTTL = int64(24 * time.Hour / time.Second) // if sticky task_list stopped being updated, remove it in one day
)

const (
	// Row types for table executions
	rowTypeShard = iota
	rowTypeExecution
	rowTypeTransferTask
	rowTypeTimerTask
	rowTypeReplicationTask
	rowTypeDLQ
	rowTypeCrossClusterTask
)

const (
	templateWorkflowExecutionType = `{` +
		`domain_id: ?, ` +
		`workflow_id: ?, ` +
		`run_id: ?, ` +
		`parent_domain_id: ?, ` +
		`parent_workflow_id: ?, ` +
		`parent_run_id: ?, ` +
		`initiated_id: ?, ` +
		`completion_event_batch_id: ?, ` +
		`completion_event: ?, ` +
		`completion_event_data_encoding: ?, ` +
		`task_list: ?, ` +
		`workflow_type_name: ?, ` +
		`workflow_timeout: ?, ` +
		`decision_task_timeout: ?, ` +
		`execution_context: ?, ` +
		`state: ?, ` +
		`close_status: ?, ` +
		`last_first_event_id: ?, ` +
		`last_event_task_id: ?, ` +
		`next_event_id: ?, ` +
		`last_processed_event: ?, ` +
		`start_time: ?, ` +
		`last_updated_time: ?, ` +
		`create_request_id: ?, ` +
		`signal_count: ?, ` +
		`history_size: ?, ` +
		`decision_version: ?, ` +
		`decision_schedule_id: ?, ` +
		`decision_started_id: ?, ` +
		`decision_request_id: ?, ` +
		`decision_timeout: ?, ` +
		`decision_attempt: ?, ` +
		`decision_timestamp: ?, ` +
		`decision_scheduled_timestamp: ?, ` +
		`decision_original_scheduled_timestamp: ?, ` +
		`cancel_requested: ?, ` +
		`cancel_request_id: ?, ` +
		`sticky_task_list: ?, ` +
		`sticky_schedule_to_start_timeout: ?,` +
		`client_library_version: ?, ` +
		`client_feature_version: ?, ` +
		`client_impl: ?, ` +
		`auto_reset_points: ?, ` +
		`auto_reset_points_encoding: ?, ` +
		`attempt: ?, ` +
		`has_retry_policy: ?, ` +
		`init_interval: ?, ` +
		`backoff_coefficient: ?, ` +
		`max_interval: ?, ` +
		`expiration_time: ?, ` +
		`max_attempts: ?, ` +
		`non_retriable_errors: ?, ` +
		`event_store_version: ?, ` +
		`branch_token: ?, ` +
		`cron_schedule: ?, ` +
		`expiration_seconds: ?, ` +
		`search_attributes: ?, ` +
		`memo: ? ` +
		`}`

	templateTransferTaskType = `{` +
		`domain_id: ?, ` +
		`workflow_id: ?, ` +
		`run_id: ?, ` +
		`visibility_ts: ?, ` +
		`task_id: ?, ` +
		`target_domain_id: ?, ` +
		`target_workflow_id: ?, ` +
		`target_run_id: ?, ` +
		`target_child_workflow_only: ?, ` +
		`task_list: ?, ` +
		`type: ?, ` +
		`schedule_id: ?, ` +
		`record_visibility: ?, ` +
		`version: ?` +
		`}`

	templateCrossClusterTaskType = templateTransferTaskType

	templateReplicationTaskType = `{` +
		`domain_id: ?, ` +
		`workflow_id: ?, ` +
		`run_id: ?, ` +
		`task_id: ?, ` +
		`type: ?, ` +
		`first_event_id: ?,` +
		`next_event_id: ?,` +
		`version: ?,` +
		`scheduled_id: ?, ` +
		`event_store_version: ?, ` +
		`branch_token: ?, ` +
		`new_run_event_store_version: ?, ` +
		`new_run_branch_token: ?, ` +
		`created_time: ? ` +
		`}`

	templateTimerTaskType = `{` +
		`domain_id: ?, ` +
		`workflow_id: ?, ` +
		`run_id: ?, ` +
		`visibility_ts: ?, ` +
		`task_id: ?, ` +
		`type: ?, ` +
		`timeout_type: ?, ` +
		`event_id: ?, ` +
		`schedule_attempt: ?, ` +
		`version: ?` +
		`}`

	templateActivityInfoType = `{` +
		`version: ?,` +
		`schedule_id: ?, ` +
		`scheduled_event_batch_id: ?, ` +
		`scheduled_event: ?, ` +
		`scheduled_time: ?, ` +
		`started_id: ?, ` +
		`started_event: ?, ` +
		`started_time: ?, ` +
		`activity_id: ?, ` +
		`request_id: ?, ` +
		`details: ?, ` +
		`schedule_to_start_timeout: ?, ` +
		`schedule_to_close_timeout: ?, ` +
		`start_to_close_timeout: ?, ` +
		`heart_beat_timeout: ?, ` +
		`cancel_requested: ?, ` +
		`cancel_request_id: ?, ` +
		`last_hb_updated_time: ?, ` +
		`timer_task_status: ?, ` +
		`attempt: ?, ` +
		`task_list: ?, ` +
		`started_identity: ?, ` +
		`has_retry_policy: ?, ` +
		`init_interval: ?, ` +
		`backoff_coefficient: ?, ` +
		`max_interval: ?, ` +
		`expiration_time: ?, ` +
		`max_attempts: ?, ` +
		`non_retriable_errors: ?, ` +
		`last_failure_reason: ?, ` +
		`last_worker_identity: ?, ` +
		`last_failure_details: ?, ` +
		`event_data_encoding: ?` +
		`}`

	templateTimerInfoType = `{` +
		`version: ?,` +
		`timer_id: ?, ` +
		`started_id: ?, ` +
		`expiry_time: ?, ` +
		`task_id: ?` +
		`}`

	templateChildExecutionInfoType = `{` +
		`version: ?,` +
		`initiated_id: ?, ` +
		`initiated_event_batch_id: ?, ` +
		`initiated_event: ?, ` +
		`started_id: ?, ` +
		`started_workflow_id: ?, ` +
		`started_run_id: ?, ` +
		`started_event: ?, ` +
		`create_request_id: ?, ` +
		`event_data_encoding: ?, ` +
		`domain_name: ?, ` +
		`workflow_type_name: ?, ` +
		`parent_close_policy: ?` +
		`}`

	templateRequestCancelInfoType = `{` +
		`version: ?,` +
		`initiated_id: ?, ` +
		`initiated_event_batch_id: ?, ` +
		`cancel_request_id: ? ` +
		`}`

	templateSignalInfoType = `{` +
		`version: ?,` +
		`initiated_id: ?, ` +
		`initiated_event_batch_id: ?, ` +
		`signal_request_id: ?, ` +
		`signal_name: ?, ` +
		`input: ?, ` +
		`control: ?` +
		`}`

	templateChecksumType = `{` +
		`version: ?, ` +
		`flavor: ?, ` +
		`value: ? ` +
		`}`

	templateUpdateCurrentWorkflowExecutionQuery = `UPDATE executions USING TTL 0 ` +
		`SET current_run_id = ?,
execution = {run_id: ?, create_request_id: ?, state: ?, close_status: ?},
workflow_last_write_version = ?,
workflow_state = ? ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? ` +
		`IF current_run_id = ? `

	templateUpdateCurrentWorkflowExecutionForNewQuery = templateUpdateCurrentWorkflowExecutionQuery +
		`and workflow_last_write_version = ? ` +
		`and workflow_state = ? `

	templateCreateCurrentWorkflowExecutionQuery = `INSERT INTO executions (` +
		`shard_id, type, domain_id, workflow_id, run_id, visibility_ts, task_id, current_run_id, execution, workflow_last_write_version, workflow_state) ` +
		`VALUES(?, ?, ?, ?, ?, ?, ?, ?, {run_id: ?, create_request_id: ?, state: ?, close_status: ?}, ?, ?) IF NOT EXISTS USING TTL 0 `

	templateCreateWorkflowExecutionWithVersionHistoriesQuery = `INSERT INTO executions (` +
		`shard_id, domain_id, workflow_id, run_id, type, execution, next_event_id, visibility_ts, task_id, version_histories, version_histories_encoding, checksum, workflow_last_write_version, workflow_state) ` +
		`VALUES(?, ?, ?, ?, ?, ` + templateWorkflowExecutionType + `, ?, ?, ?, ?, ?, ` + templateChecksumType + `, ?, ?) IF NOT EXISTS `

	templateCreateTransferTaskQuery = `INSERT INTO executions (` +
		`shard_id, type, domain_id, workflow_id, run_id, transfer, visibility_ts, task_id) ` +
		`VALUES(?, ?, ?, ?, ?, ` + templateTransferTaskType + `, ?, ?)`

	templateCreateCrossClusterTaskQuery = `INSERT INTO executions (` +
		`shard_id, type, domain_id, workflow_id, run_id, cross_cluster, visibility_ts, task_id) ` +
		`VALUES(?, ?, ?, ?, ?, ` + templateCrossClusterTaskType + `, ?, ?)`

	templateCreateReplicationTaskQuery = `INSERT INTO executions (` +
		`shard_id, type, domain_id, workflow_id, run_id, replication, visibility_ts, task_id) ` +
		`VALUES(?, ?, ?, ?, ?, ` + templateReplicationTaskType + `, ?, ?)`

	templateCreateTimerTaskQuery = `INSERT INTO executions (` +
		`shard_id, type, domain_id, workflow_id, run_id, timer, visibility_ts, task_id) ` +
		`VALUES(?, ?, ?, ?, ?, ` + templateTimerTaskType + `, ?, ?)`

	templateUpdateLeaseQuery = `UPDATE executions ` +
		`SET range_id = ? ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? ` +
		`IF range_id = ?`

	// TODO: remove replication_state after all 2DC workflows complete
	templateGetWorkflowExecutionQuery = `SELECT execution, replication_state, activity_map, timer_map, ` +
		`child_executions_map, request_cancel_map, signal_map, signal_requested, buffered_events_list, ` +
		`buffered_replication_tasks_map, version_histories, version_histories_encoding, checksum ` +
		`FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ?`

	templateGetCurrentExecutionQuery = `SELECT current_run_id, execution, workflow_last_write_version ` +
		`FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ?`

	templateListCurrentExecutionsQuery = `SELECT domain_id, workflow_id, run_id, current_run_id, workflow_state ` +
		`FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ?`

	templateIsWorkflowExecutionExistsQuery = `SELECT shard_id, type, domain_id, workflow_id, run_id, visibility_ts, task_id ` +
		`FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ?`

	templateListWorkflowExecutionQuery = `SELECT run_id, execution, version_histories, version_histories_encoding ` +
		`FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ?`

	templateCheckWorkflowExecutionQuery = `UPDATE executions ` +
		`SET next_event_id = ? ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? ` +
		`IF next_event_id = ?`

	templateUpdateWorkflowExecutionWithVersionHistoriesQuery = `UPDATE executions ` +
		`SET execution = ` + templateWorkflowExecutionType +
		`, next_event_id = ? ` +
		`, version_histories = ? ` +
		`, version_histories_encoding = ? ` +
		`, checksum = ` + templateChecksumType +
		`, workflow_last_write_version = ? ` +
		`, workflow_state = ? ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? ` +
		`IF next_event_id = ? `

	templateUpdateActivityInfoQuery = `UPDATE executions ` +
		`SET activity_map[ ? ] =` + templateActivityInfoType + ` ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateResetActivityInfoQuery = `UPDATE executions ` +
		`SET activity_map = ?` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateUpdateTimerInfoQuery = `UPDATE executions ` +
		`SET timer_map[ ? ] =` + templateTimerInfoType + ` ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateResetTimerInfoQuery = `UPDATE executions ` +
		`SET timer_map = ?` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateUpdateChildExecutionInfoQuery = `UPDATE executions ` +
		`SET child_executions_map[ ? ] =` + templateChildExecutionInfoType + ` ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateResetChildExecutionInfoQuery = `UPDATE executions ` +
		`SET child_executions_map = ?` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateUpdateRequestCancelInfoQuery = `UPDATE executions ` +
		`SET request_cancel_map[ ? ] =` + templateRequestCancelInfoType + ` ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateResetRequestCancelInfoQuery = `UPDATE executions ` +
		`SET request_cancel_map = ?` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateUpdateSignalInfoQuery = `UPDATE executions ` +
		`SET signal_map[ ? ] =` + templateSignalInfoType + ` ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateResetSignalInfoQuery = `UPDATE executions ` +
		`SET signal_map = ?` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateUpdateSignalRequestedQuery = `UPDATE executions ` +
		`SET signal_requested = signal_requested + ? ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateResetSignalRequestedQuery = `UPDATE executions ` +
		`SET signal_requested = ?` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateAppendBufferedEventsQuery = `UPDATE executions ` +
		`SET buffered_events_list = buffered_events_list + ? ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateDeleteBufferedEventsQuery = `UPDATE executions ` +
		`SET buffered_events_list = [] ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateDeleteActivityInfoQuery = `DELETE activity_map[ ? ] ` +
		`FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateDeleteTimerInfoQuery = `DELETE timer_map[ ? ] ` +
		`FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateDeleteChildExecutionInfoQuery = `DELETE child_executions_map[ ? ] ` +
		`FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateDeleteRequestCancelInfoQuery = `DELETE request_cancel_map[ ? ] ` +
		`FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateDeleteSignalInfoQuery = `DELETE signal_map[ ? ] ` +
		`FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateDeleteWorkflowExecutionMutableStateQuery = `DELETE FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateDeleteWorkflowExecutionCurrentRowQuery = templateDeleteWorkflowExecutionMutableStateQuery + " if current_run_id = ? "

	templateDeleteWorkflowExecutionSignalRequestedQuery = `UPDATE executions ` +
		`SET signal_requested = signal_requested - ? ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateGetTransferTasksQuery = `SELECT transfer ` +
		`FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id > ? ` +
		`and task_id <= ?`

	templateGetCrossClusterTasksQuery = `SELECT cross_cluster ` +
		`FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id > ? ` +
		`and task_id <= ?`

	templateGetReplicationTasksQuery = `SELECT replication ` +
		`FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id > ? ` +
		`and task_id <= ?`

	templateGetDLQSizeQuery = `SELECT count(1) as count ` +
		`FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ?`

	templateCompleteTransferTaskQuery = `DELETE FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ?`

	templateRangeCompleteTransferTaskQuery = `DELETE FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id > ? ` +
		`and task_id <= ?`

	templateCompleteCrossClusterTaskQuery = templateCompleteTransferTaskQuery

	templateRangeCompleteCrossClusterTaskQuery = templateRangeCompleteTransferTaskQuery

	templateCompleteReplicationTaskBeforeQuery = `DELETE FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id <= ?`

	templateCompleteReplicationTaskQuery = templateCompleteTransferTaskQuery

	templateRangeCompleteReplicationTaskQuery = templateRangeCompleteTransferTaskQuery

	templateGetTimerTasksQuery = `SELECT timer ` +
		`FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ?` +
		`and domain_id = ? ` +
		`and workflow_id = ?` +
		`and run_id = ?` +
		`and visibility_ts >= ? ` +
		`and visibility_ts < ?`

	templateCompleteTimerTaskQuery = `DELETE FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ?` +
		`and run_id = ?` +
		`and visibility_ts = ? ` +
		`and task_id = ?`

	templateRangeCompleteTimerTaskQuery = `DELETE FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ?` +
		`and run_id = ?` +
		`and visibility_ts >= ? ` +
		`and visibility_ts < ?`
)

var (
	defaultDateTime            = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	defaultVisibilityTimestamp = p.UnixNanoToDBTimestamp(defaultDateTime.UnixNano())
)

func (d *cassandraStore) GetName() string {
	return cassandraPersistenceName
}

// Close releases the underlying resources held by this object
func (d *cassandraStore) Close() {
	if d.session != nil {
		d.session.Close()
	}
}

// NewWorkflowExecutionPersistence is used to create an instance of workflowExecutionManager implementation
func NewWorkflowExecutionPersistence(
	shardID int,
	client gocql.Client,
	session gocql.Session,
	logger log.Logger,
) (p.ExecutionStore, error) {
	db := cassandra.NewCassandraDBFromSession(client, session, logger)

	return &cassandraPersistence{
		cassandraStore: cassandraStore{
			client:  client,
			session: session,
			logger:  logger,
		},
		shardID: shardID,
		db:      db,
	}, nil
}

func (d *cassandraPersistence) GetShardID() int {
	return d.shardID
}

func (d *cassandraPersistence) CreateWorkflowExecution(
	ctx context.Context,
	request *p.InternalCreateWorkflowExecutionRequest,
) (*p.CreateWorkflowExecutionResponse, error) {

	batch := d.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)

	newWorkflow := request.NewWorkflowSnapshot
	executionInfo := newWorkflow.ExecutionInfo
	lastWriteVersion := newWorkflow.LastWriteVersion
	domainID := executionInfo.DomainID
	workflowID := executionInfo.WorkflowID
	runID := executionInfo.RunID

	if err := p.ValidateCreateWorkflowModeState(
		request.Mode,
		newWorkflow,
	); err != nil {
		return nil, err
	}

	currentWorkflowWriteReq := nosqlplugin.CurrentWorkflowWriteRequest{
		Row: nosqlplugin.CurrentWorkflowRow{
			ShardID:          d.shardID,
			DomainID:         domainID,
			WorkflowID:       workflowID,
			RunID:            runID,
			State:            executionInfo.State,
			CloseStatus:      executionInfo.CloseStatus,
			CreateRequestID:  executionInfo.CreateRequestID,
			LastWriteVersion: lastWriteVersion,
		},
	}
	switch request.Mode {
	case p.CreateWorkflowModeZombie:
		// noop
		currentWorkflowWriteReq.WriteMode = nosqlplugin.CurrentWorkflowWriteModeNoop
	case p.CreateWorkflowModeContinueAsNew:
		currentWorkflowWriteReq.WriteMode = nosqlplugin.CurrentWorkflowWriteModeUpdate
		currentWorkflowWriteReq.Condition = &nosqlplugin.CurrentWorkflowWriteCondition{
			CurrentRunID: common.StringPtr(request.PreviousRunID),
		}
	case p.CreateWorkflowModeWorkflowIDReuse:
		currentWorkflowWriteReq.WriteMode = nosqlplugin.CurrentWorkflowWriteModeUpdate
		currentWorkflowWriteReq.Condition = &nosqlplugin.CurrentWorkflowWriteCondition{
			CurrentRunID:     common.StringPtr(request.PreviousRunID),
			State:            common.IntPtr(p.WorkflowStateCompleted),
			LastWriteVersion: common.Int64Ptr(request.PreviousLastWriteVersion),
		}
	case p.CreateWorkflowModeBrandNew:
		currentWorkflowWriteReq.WriteMode = nosqlplugin.CurrentWorkflowWriteModeInsert
	default:
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("unknown mode: %v", request.Mode),
		}
	}

	//if err := applyWorkflowSnapshotBatchAsNew(
	//	batch,
	//	d.shardID,
	//	&newWorkflow,
	//); err != nil {
	//	return nil, err
	//}

	//execution *WorkflowExecutionRow,
	//	transferTasks []*TransferTask,
	//	crossClusterTasks []*CrossClusterTask,
	//	replicationTasks []*ReplicationTask,
	//	timerTasks []*TimerTask,
	//	activityInfoMap map[int64]*persistence.InternalActivityInfo,
	//	timerInfoMap map[string]*persistence.TimerInfo,
	//	childWorkflowInfoMap map[int64]*persistence.InternalChildExecutionInfo,
	//	requestCancelInfoMap map[int64]*persistence.RequestCancelInfo,
	//	signalInfoMap map[int64]*persistence.SignalInfo,
	//	signalRequestedIDs []string,
	//	shardCondition *ShardCondition,
	//

	err := prepareWorkflowSnapshotBatchAsNew(
		d.shardID,
		&newWorkflow,
	)

	if err != nil {
		return nil, err
	}

	shardCondition := nosqlplugin.ShardCondition{
		ShardID: d.shardID,
		RangeID: request.RangeID,
	}
	//batch.Query(templateUpdateLeaseQuery,
	//	request.RangeID,
	//	d.shardID,
	//	rowTypeShard,
	//	rowTypeShardDomainID,
	//	rowTypeShardWorkflowID,
	//	rowTypeShardRunID,
	//	defaultVisibilityTimestamp,
	//	rowTypeShardTaskID,
	//	request.RangeID,
	//)

	previous := make(map[string]interface{})
	applied, iter, err := d.session.MapExecuteBatchCAS(batch, previous)
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()

	if err != nil {
		return nil, convertCommonErrors(d.client, "CreateWorkflowExecution", err)
	}

	if !applied {
		// There can be two reasons why the query does not get applied. Either the RangeID has changed, or
		// the workflow is already started. Check the row info returned by Cassandra to figure out which one it is.
	GetFailureReasonLoop:
		for {
			rowType, ok := previous["type"].(int)
			if !ok {
				// This should never happen, as all our rows have the type field.
				break GetFailureReasonLoop
			}
			runID := previous["run_id"].(gocql.UUID).String()

			if rowType == rowTypeShard {
				if rangeID, ok := previous["range_id"].(int64); ok && rangeID != request.RangeID {
					// CreateWorkflowExecution failed because rangeID was modified
					return nil, &p.ShardOwnershipLostError{
						ShardID: d.shardID,
						Msg: fmt.Sprintf("Failed to create workflow execution.  Request RangeID: %v, Actual RangeID: %v",
							request.RangeID, rangeID),
					}
				}

			} else if rowType == rowTypeExecution && runID == permanentRunID {
				var columns []string
				for k, v := range previous {
					columns = append(columns, fmt.Sprintf("%s=%v", k, v))
				}

				if execution, ok := previous["execution"].(map[string]interface{}); ok {
					// CreateWorkflowExecution failed because it already exists
					executionInfo := createWorkflowExecutionInfo(execution)
					lastWriteVersion := common.EmptyVersion
					if previous["workflow_last_write_version"] != nil {
						lastWriteVersion = previous["workflow_last_write_version"].(int64)
					}

					msg := fmt.Sprintf("Workflow execution already running. WorkflowId: %v, RunId: %v, rangeID: %v, columns: (%v)",
						executionInfo.WorkflowID, executionInfo.RunID, request.RangeID, strings.Join(columns, ","))
					if request.Mode == p.CreateWorkflowModeBrandNew {
						return nil, &p.WorkflowExecutionAlreadyStartedError{
							Msg:              msg,
							StartRequestID:   executionInfo.CreateRequestID,
							RunID:            executionInfo.RunID,
							State:            executionInfo.State,
							CloseStatus:      executionInfo.CloseStatus,
							LastWriteVersion: lastWriteVersion,
						}
					}
					return nil, &p.CurrentWorkflowConditionFailedError{Msg: msg}

				}

				if prevRunID := previous["current_run_id"].(gocql.UUID).String(); prevRunID != request.PreviousRunID {
					// currentRunID on previous run has been changed, return to caller to handle
					msg := fmt.Sprintf("Workflow execution creation condition failed by mismatch runID. WorkflowId: %v, Expected Current RunID: %v, Actual Current RunID: %v",
						executionInfo.WorkflowID, request.PreviousRunID, prevRunID)
					return nil, &p.CurrentWorkflowConditionFailedError{Msg: msg}
				}

				msg := fmt.Sprintf("Workflow execution creation condition failed. WorkflowId: %v, CurrentRunID: %v, columns: (%v)",
					executionInfo.WorkflowID, executionInfo.RunID, strings.Join(columns, ","))
				return nil, &p.CurrentWorkflowConditionFailedError{Msg: msg}
			} else if rowType == rowTypeExecution && runID == executionInfo.RunID {
				msg := fmt.Sprintf("Workflow execution already running. WorkflowId: %v, RunId: %v, rangeID: %v",
					executionInfo.WorkflowID, executionInfo.RunID, request.RangeID)
				lastWriteVersion := common.EmptyVersion
				if previous["workflow_last_write_version"] != nil {
					lastWriteVersion = previous["workflow_last_write_version"].(int64)
				}
				return nil, &p.WorkflowExecutionAlreadyStartedError{
					Msg:              msg,
					StartRequestID:   executionInfo.CreateRequestID,
					RunID:            executionInfo.RunID,
					State:            executionInfo.State,
					CloseStatus:      executionInfo.CloseStatus,
					LastWriteVersion: lastWriteVersion,
				}
			}

			previous = make(map[string]interface{})
			if !iter.MapScan(previous) {
				// Cassandra returns the actual row that caused a condition failure, so we should always return
				// from the checks above, but just in case.
				break GetFailureReasonLoop
			}
		}

		return nil, newShardOwnershipLostError(d.shardID, request.RangeID, previous)
	}

	return &p.CreateWorkflowExecutionResponse{}, nil
}

func (d *cassandraPersistence) GetWorkflowExecution(
	ctx context.Context,
	request *p.InternalGetWorkflowExecutionRequest,
) (*p.InternalGetWorkflowExecutionResponse, error) {

	execution := request.Execution
	query := d.session.Query(templateGetWorkflowExecutionQuery,
		d.shardID,
		rowTypeExecution,
		request.DomainID,
		execution.WorkflowID,
		execution.RunID,
		defaultVisibilityTimestamp,
		rowTypeExecutionTaskID,
	).WithContext(ctx)

	result := make(map[string]interface{})
	if err := query.MapScan(result); err != nil {
		if d.client.IsNotFoundError(err) {
			return nil, &types.EntityNotExistsError{
				Message: fmt.Sprintf("Workflow execution not found.  WorkflowId: %v, RunId: %v",
					execution.WorkflowID, execution.RunID),
			}
		}

		return nil, convertCommonErrors(d.client, "GetWorkflowExecution", err)
	}

	state := &p.InternalWorkflowMutableState{}
	info := createWorkflowExecutionInfo(result["execution"].(map[string]interface{}))
	state.ExecutionInfo = info
	state.VersionHistories = p.NewDataBlob(result["version_histories"].([]byte), common.EncodingType(result["version_histories_encoding"].(string)))
	// TODO: remove this after all 2DC workflows complete
	replicationState := createReplicationState(result["replication_state"].(map[string]interface{}))
	state.ReplicationState = replicationState

	activityInfos := make(map[int64]*p.InternalActivityInfo)
	aMap := result["activity_map"].(map[int64]map[string]interface{})
	for key, value := range aMap {
		info := createActivityInfo(request.DomainID, value)
		activityInfos[key] = info
	}
	state.ActivityInfos = activityInfos

	timerInfos := make(map[string]*p.TimerInfo)
	tMap := result["timer_map"].(map[string]map[string]interface{})
	for key, value := range tMap {
		info := createTimerInfo(value)
		timerInfos[key] = info
	}
	state.TimerInfos = timerInfos

	childExecutionInfos := make(map[int64]*p.InternalChildExecutionInfo)
	cMap := result["child_executions_map"].(map[int64]map[string]interface{})
	for key, value := range cMap {
		info := createChildExecutionInfo(value)
		childExecutionInfos[key] = info
	}
	state.ChildExecutionInfos = childExecutionInfos

	requestCancelInfos := make(map[int64]*p.RequestCancelInfo)
	rMap := result["request_cancel_map"].(map[int64]map[string]interface{})
	for key, value := range rMap {
		info := createRequestCancelInfo(value)
		requestCancelInfos[key] = info
	}
	state.RequestCancelInfos = requestCancelInfos

	signalInfos := make(map[int64]*p.SignalInfo)
	sMap := result["signal_map"].(map[int64]map[string]interface{})
	for key, value := range sMap {
		info := createSignalInfo(value)
		signalInfos[key] = info
	}
	state.SignalInfos = signalInfos

	signalRequestedIDs := make(map[string]struct{})
	sList := mustConvertToSlice(result["signal_requested"])
	for _, v := range sList {
		signalRequestedIDs[v.(gocql.UUID).String()] = struct{}{}
	}
	state.SignalRequestedIDs = signalRequestedIDs

	eList := result["buffered_events_list"].([]map[string]interface{})
	bufferedEventsBlobs := make([]*p.DataBlob, 0, len(eList))
	for _, v := range eList {
		blob := createHistoryEventBatchBlob(v)
		bufferedEventsBlobs = append(bufferedEventsBlobs, blob)
	}
	state.BufferedEvents = bufferedEventsBlobs

	state.Checksum = createChecksum(result["checksum"].(map[string]interface{}))

	return &p.InternalGetWorkflowExecutionResponse{State: state}, nil
}

func (d *cassandraPersistence) UpdateWorkflowExecution(
	ctx context.Context,
	request *p.InternalUpdateWorkflowExecutionRequest,
) error {

	batch := d.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)

	updateWorkflow := request.UpdateWorkflowMutation
	newWorkflow := request.NewWorkflowSnapshot

	executionInfo := updateWorkflow.ExecutionInfo
	domainID := executionInfo.DomainID
	workflowID := executionInfo.WorkflowID
	runID := executionInfo.RunID
	shardID := d.shardID

	if err := p.ValidateUpdateWorkflowModeState(
		request.Mode,
		updateWorkflow,
		newWorkflow,
	); err != nil {
		return err
	}

	switch request.Mode {
	case p.UpdateWorkflowModeBypassCurrent:
		if err := d.assertNotCurrentExecution(
			ctx,
			domainID,
			workflowID,
			runID); err != nil {
			return err
		}

	case p.UpdateWorkflowModeUpdateCurrent:
		if newWorkflow != nil {
			newExecutionInfo := newWorkflow.ExecutionInfo
			newStartVersion := newWorkflow.StartVersion
			newLastWriteVersion := newWorkflow.LastWriteVersion
			newDomainID := newExecutionInfo.DomainID
			newWorkflowID := newExecutionInfo.WorkflowID
			newRunID := newExecutionInfo.RunID

			if domainID != newDomainID {
				return &types.InternalServiceError{
					Message: fmt.Sprintf("UpdateWorkflowExecution: cannot continue as new to another domain"),
				}
			}

			if err := createOrUpdateCurrentExecution(batch,
				p.CreateWorkflowModeContinueAsNew,
				d.shardID,
				newDomainID,
				newWorkflowID,
				newRunID,
				newExecutionInfo.State,
				newExecutionInfo.CloseStatus,
				newExecutionInfo.CreateRequestID,
				newStartVersion,
				newLastWriteVersion,
				runID,
				0, // for continue as new, this is not used
			); err != nil {
				return err
			}

		} else {
			lastWriteVersion := updateWorkflow.LastWriteVersion
			batch.Query(templateUpdateCurrentWorkflowExecutionQuery,
				runID,
				runID,
				executionInfo.CreateRequestID,
				executionInfo.State,
				executionInfo.CloseStatus,
				lastWriteVersion,
				executionInfo.State,
				d.shardID,
				rowTypeExecution,
				domainID,
				workflowID,
				permanentRunID,
				defaultVisibilityTimestamp,
				rowTypeExecutionTaskID,
				runID,
			)
		}

	default:
		return &types.InternalServiceError{
			Message: fmt.Sprintf("UpdateWorkflowExecution: unknown mode: %v", request.Mode),
		}
	}

	if err := applyWorkflowMutationBatch(batch, shardID, &updateWorkflow); err != nil {
		return err
	}
	if newWorkflow != nil {
		if err := applyWorkflowSnapshotBatchAsNew(batch,
			d.shardID,
			newWorkflow,
		); err != nil {
			return err
		}
	}

	// Verifies that the RangeID has not changed
	batch.Query(templateUpdateLeaseQuery,
		request.RangeID,
		d.shardID,
		rowTypeShard,
		rowTypeShardDomainID,
		rowTypeShardWorkflowID,
		rowTypeShardRunID,
		defaultVisibilityTimestamp,
		rowTypeShardTaskID,
		request.RangeID,
	)

	previous := make(map[string]interface{})
	applied, iter, err := d.session.MapExecuteBatchCAS(batch, previous)
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()

	if err != nil {
		return convertCommonErrors(d.client, "UpdateWorkflowExecution", err)
	}

	if !applied {
		return d.getExecutionConditionalUpdateFailure(previous, iter, executionInfo.RunID, updateWorkflow.Condition, request.RangeID, executionInfo.RunID)
	}
	return nil
}

func (d *cassandraPersistence) ConflictResolveWorkflowExecution(
	ctx context.Context,
	request *p.InternalConflictResolveWorkflowExecutionRequest,
) error {
	batch := d.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)

	currentWorkflow := request.CurrentWorkflowMutation
	resetWorkflow := request.ResetWorkflowSnapshot
	newWorkflow := request.NewWorkflowSnapshot

	shardID := d.shardID

	domainID := resetWorkflow.ExecutionInfo.DomainID
	workflowID := resetWorkflow.ExecutionInfo.WorkflowID

	if err := p.ValidateConflictResolveWorkflowModeState(
		request.Mode,
		resetWorkflow,
		newWorkflow,
		currentWorkflow,
	); err != nil {
		return err
	}

	var prevRunID string

	switch request.Mode {
	case p.ConflictResolveWorkflowModeBypassCurrent:
		if err := d.assertNotCurrentExecution(
			ctx,
			domainID,
			workflowID,
			resetWorkflow.ExecutionInfo.RunID); err != nil {
			return err
		}

	case p.ConflictResolveWorkflowModeUpdateCurrent:
		executionInfo := resetWorkflow.ExecutionInfo
		lastWriteVersion := resetWorkflow.LastWriteVersion
		if newWorkflow != nil {
			executionInfo = newWorkflow.ExecutionInfo
			lastWriteVersion = newWorkflow.LastWriteVersion
		}
		runID := executionInfo.RunID
		createRequestID := executionInfo.CreateRequestID
		state := executionInfo.State
		closeStatus := executionInfo.CloseStatus

		if currentWorkflow != nil {
			prevRunID = currentWorkflow.ExecutionInfo.RunID

			batch.Query(templateUpdateCurrentWorkflowExecutionQuery,
				runID,
				runID,
				createRequestID,
				state,
				closeStatus,
				lastWriteVersion,
				state,
				shardID,
				rowTypeExecution,
				domainID,
				workflowID,
				permanentRunID,
				defaultVisibilityTimestamp,
				rowTypeExecutionTaskID,
				prevRunID,
			)
		} else {
			// reset workflow is current
			prevRunID = resetWorkflow.ExecutionInfo.RunID

			batch.Query(templateUpdateCurrentWorkflowExecutionQuery,
				runID,
				runID,
				createRequestID,
				state,
				closeStatus,
				lastWriteVersion,
				state,
				shardID,
				rowTypeExecution,
				domainID,
				workflowID,
				permanentRunID,
				defaultVisibilityTimestamp,
				rowTypeExecutionTaskID,
				prevRunID,
			)
		}

	default:
		return &types.InternalServiceError{
			Message: fmt.Sprintf("ConflictResolveWorkflowExecution: unknown mode: %v", request.Mode),
		}
	}

	if err := applyWorkflowSnapshotBatchAsReset(batch,
		shardID,
		&resetWorkflow); err != nil {
		return err
	}

	if currentWorkflow != nil {
		if err := applyWorkflowMutationBatch(batch, shardID, currentWorkflow); err != nil {
			return err
		}
	}
	if newWorkflow != nil {
		if err := applyWorkflowSnapshotBatchAsNew(batch, shardID, newWorkflow); err != nil {
			return err
		}
	}

	// Verifies that the RangeID has not changed
	batch.Query(templateUpdateLeaseQuery,
		request.RangeID,
		d.shardID,
		rowTypeShard,
		rowTypeShardDomainID,
		rowTypeShardWorkflowID,
		rowTypeShardRunID,
		defaultVisibilityTimestamp,
		rowTypeShardTaskID,
		request.RangeID,
	)

	previous := make(map[string]interface{})
	applied, iter, err := d.session.MapExecuteBatchCAS(batch, previous)
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()

	if err != nil {
		return convertCommonErrors(d.client, "ConflictResolveWorkflowExecution", err)
	}

	if !applied {
		return d.getExecutionConditionalUpdateFailure(previous, iter, resetWorkflow.ExecutionInfo.RunID, request.ResetWorkflowSnapshot.Condition, request.RangeID, prevRunID)
	}
	return nil
}

func (d *cassandraPersistence) getExecutionConditionalUpdateFailure(
	previous map[string]interface{},
	iter gocql.Iter,
	requestRunID string,
	requestCondition int64,
	requestRangeID int64,
	requestConditionalRunID string,
) error {
	// There can be three reasons why the query does not get applied: the RangeID has changed, or the next_event_id or current_run_id check failed.
	// Check the row info returned by Cassandra to figure out which one it is.
	rangeIDUnmatch := false
	actualRangeID := int64(0)
	nextEventIDUnmatch := false
	actualNextEventID := int64(0)
	runIDUnmatch := false
	actualCurrRunID := ""
	allPrevious := []map[string]interface{}{}

GetFailureReasonLoop:
	for {
		rowType, ok := previous["type"].(int)
		if !ok {
			// This should never happen, as all our rows have the type field.
			break GetFailureReasonLoop
		}

		runID := previous["run_id"].(gocql.UUID).String()

		if rowType == rowTypeShard {
			if actualRangeID, ok = previous["range_id"].(int64); ok && actualRangeID != requestRangeID {
				// UpdateWorkflowExecution failed because rangeID was modified
				rangeIDUnmatch = true
			}
		} else if rowType == rowTypeExecution && runID == requestRunID {
			if actualNextEventID, ok = previous["next_event_id"].(int64); ok && actualNextEventID != requestCondition {
				// UpdateWorkflowExecution failed because next event ID is unexpected
				nextEventIDUnmatch = true
			}
		} else if rowType == rowTypeExecution && runID == permanentRunID {
			// UpdateWorkflowExecution failed because current_run_id is unexpected
			if actualCurrRunID = previous["current_run_id"].(gocql.UUID).String(); actualCurrRunID != requestConditionalRunID {
				// UpdateWorkflowExecution failed because next event ID is unexpected
				runIDUnmatch = true
			}
		}

		allPrevious = append(allPrevious, previous)
		previous = make(map[string]interface{})
		if !iter.MapScan(previous) {
			// Cassandra returns the actual row that caused a condition failure, so we should always return
			// from the checks above, but just in case.
			break GetFailureReasonLoop
		}
	}

	if rangeIDUnmatch {
		return &p.ShardOwnershipLostError{
			ShardID: d.shardID,
			Msg: fmt.Sprintf("Failed to update mutable state.  Request RangeID: %v, Actual RangeID: %v",
				requestRangeID, actualRangeID),
		}
	}

	if runIDUnmatch {
		return &p.CurrentWorkflowConditionFailedError{
			Msg: fmt.Sprintf("Failed to update mutable state.  Request Condition: %v, Actual Value: %v, Request Current RunID: %v, Actual Value: %v",
				requestCondition, actualNextEventID, requestConditionalRunID, actualCurrRunID),
		}
	}

	if nextEventIDUnmatch {
		return &p.ConditionFailedError{
			Msg: fmt.Sprintf("Failed to update mutable state.  Request Condition: %v, Actual Value: %v, Request Current RunID: %v, Actual Value: %v",
				requestCondition, actualNextEventID, requestConditionalRunID, actualCurrRunID),
		}
	}

	// At this point we only know that the write was not applied.
	var columns []string
	columnID := 0
	for _, previous := range allPrevious {
		for k, v := range previous {
			columns = append(columns, fmt.Sprintf("%v: %s=%v", columnID, k, v))
		}
		columnID++
	}
	return &p.ConditionFailedError{
		Msg: fmt.Sprintf("Failed to reset mutable state. ShardID: %v, RangeID: %v, Condition: %v, Request Current RunID: %v, columns: (%v)",
			d.shardID, requestRangeID, requestCondition, requestConditionalRunID, strings.Join(columns, ",")),
	}
}

func (d *cassandraPersistence) assertNotCurrentExecution(
	ctx context.Context,
	domainID string,
	workflowID string,
	runID string,
) error {

	if resp, err := d.GetCurrentExecution(ctx, &p.GetCurrentExecutionRequest{
		DomainID:   domainID,
		WorkflowID: workflowID,
	}); err != nil {
		if _, ok := err.(*types.EntityNotExistsError); ok {
			// allow bypassing no current record
			return nil
		}
		return err
	} else if resp.RunID == runID {
		return &p.ConditionFailedError{
			Msg: fmt.Sprintf("Assertion on current record failed. Current run ID is not expected: %v", resp.RunID),
		}
	}

	return nil
}

func (d *cassandraPersistence) DeleteWorkflowExecution(
	ctx context.Context,
	request *p.DeleteWorkflowExecutionRequest,
) error {
	query := d.session.Query(templateDeleteWorkflowExecutionMutableStateQuery,
		d.shardID,
		rowTypeExecution,
		request.DomainID,
		request.WorkflowID,
		request.RunID,
		defaultVisibilityTimestamp,
		rowTypeExecutionTaskID,
	).WithContext(ctx)

	err := query.Exec()
	if err != nil {
		return convertCommonErrors(d.client, "DeleteWorkflowExecution", err)
	}

	return nil
}

func (d *cassandraPersistence) DeleteCurrentWorkflowExecution(
	ctx context.Context,
	request *p.DeleteCurrentWorkflowExecutionRequest,
) error {
	query := d.session.Query(templateDeleteWorkflowExecutionCurrentRowQuery,
		d.shardID,
		rowTypeExecution,
		request.DomainID,
		request.WorkflowID,
		permanentRunID,
		defaultVisibilityTimestamp,
		rowTypeExecutionTaskID,
		request.RunID,
	).WithContext(ctx)

	err := query.Exec()
	if err != nil {
		return convertCommonErrors(d.client, "DeleteWorkflowCurrentRow", err)
	}

	return nil
}

func (d *cassandraPersistence) GetCurrentExecution(
	ctx context.Context,
	request *p.GetCurrentExecutionRequest,
) (*p.GetCurrentExecutionResponse,
	error) {
	query := d.session.Query(templateGetCurrentExecutionQuery,
		d.shardID,
		rowTypeExecution,
		request.DomainID,
		request.WorkflowID,
		permanentRunID,
		defaultVisibilityTimestamp,
		rowTypeExecutionTaskID,
	).WithContext(ctx)

	result := make(map[string]interface{})
	if err := query.MapScan(result); err != nil {
		if d.client.IsNotFoundError(err) {
			return nil, &types.EntityNotExistsError{
				Message: fmt.Sprintf("Workflow execution not found.  WorkflowId: %v",
					request.WorkflowID),
			}
		}

		return nil, convertCommonErrors(d.client, "GetCurrentExecution", err)
	}

	currentRunID := result["current_run_id"].(gocql.UUID).String()
	executionInfo := createWorkflowExecutionInfo(result["execution"].(map[string]interface{}))
	lastWriteVersion := common.EmptyVersion
	if result["workflow_last_write_version"] != nil {
		lastWriteVersion = result["workflow_last_write_version"].(int64)
	}
	return &p.GetCurrentExecutionResponse{
		RunID:            currentRunID,
		StartRequestID:   executionInfo.CreateRequestID,
		State:            executionInfo.State,
		CloseStatus:      executionInfo.CloseStatus,
		LastWriteVersion: lastWriteVersion,
	}, nil
}

func (d *cassandraPersistence) ListCurrentExecutions(
	ctx context.Context,
	request *p.ListCurrentExecutionsRequest,
) (*p.ListCurrentExecutionsResponse, error) {
	query := d.session.Query(
		templateListCurrentExecutionsQuery,
		d.shardID,
		rowTypeExecution,
	).PageSize(request.PageSize).PageState(request.PageToken).WithContext(ctx)

	iter := query.Iter()
	if iter == nil {
		return nil, &types.InternalServiceError{
			Message: "ListCurrentExecutions operation failed. Not able to create query iterator.",
		}
	}
	response := &p.ListCurrentExecutionsResponse{}
	result := make(map[string]interface{})
	for iter.MapScan(result) {
		runID := result["run_id"].(gocql.UUID).String()
		if runID != permanentRunID {
			result = make(map[string]interface{})
			continue
		}
		response.Executions = append(response.Executions, &p.CurrentWorkflowExecution{
			DomainID:     result["domain_id"].(gocql.UUID).String(),
			WorkflowID:   result["workflow_id"].(string),
			RunID:        permanentRunID,
			State:        result["workflow_state"].(int),
			CurrentRunID: result["current_run_id"].(gocql.UUID).String(),
		})
		result = make(map[string]interface{})
	}
	nextPageToken := iter.PageState()
	response.PageToken = make([]byte, len(nextPageToken))
	copy(response.PageToken, nextPageToken)

	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(d.client, "ListCurrentExecutions", err)
	}
	return response, nil
}

func (d *cassandraPersistence) IsWorkflowExecutionExists(
	ctx context.Context,
	request *p.IsWorkflowExecutionExistsRequest,
) (*p.IsWorkflowExecutionExistsResponse, error) {
	query := d.session.Query(templateIsWorkflowExecutionExistsQuery,
		d.shardID,
		rowTypeExecution,
		request.DomainID,
		request.WorkflowID,
		request.RunID,
		defaultVisibilityTimestamp,
		rowTypeExecutionTaskID,
	).WithContext(ctx)

	result := make(map[string]interface{})
	if err := query.MapScan(result); err != nil {
		if d.client.IsNotFoundError(err) {
			return &p.IsWorkflowExecutionExistsResponse{Exists: false}, nil
		}

		return nil, convertCommonErrors(d.client, "IsWorkflowExecutionExists", err)
	}
	return &p.IsWorkflowExecutionExistsResponse{Exists: true}, nil
}

func (d *cassandraPersistence) ListConcreteExecutions(
	ctx context.Context,
	request *p.ListConcreteExecutionsRequest,
) (*p.InternalListConcreteExecutionsResponse, error) {
	query := d.session.Query(
		templateListWorkflowExecutionQuery,
		d.shardID,
		rowTypeExecution,
	).PageSize(request.PageSize).PageState(request.PageToken).WithContext(ctx)

	iter := query.Iter()
	if iter == nil {
		return nil, &types.InternalServiceError{
			Message: "ListConcreteExecutions operation failed.  Not able to create query iterator.",
		}
	}

	response := &p.InternalListConcreteExecutionsResponse{}
	result := make(map[string]interface{})
	for iter.MapScan(result) {
		runID := result["run_id"].(gocql.UUID).String()
		if runID == permanentRunID {
			result = make(map[string]interface{})
			continue
		}
		response.Executions = append(response.Executions, &p.InternalListConcreteExecutionsEntity{
			ExecutionInfo:    createWorkflowExecutionInfo(result["execution"].(map[string]interface{})),
			VersionHistories: p.NewDataBlob(result["version_histories"].([]byte), common.EncodingType(result["version_histories_encoding"].(string))),
		})
		result = make(map[string]interface{})
	}
	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)

	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(d.client, "ListConcreteExecutions", err)
	}

	return response, nil
}

func (d *cassandraPersistence) GetTransferTasks(
	ctx context.Context,
	request *p.GetTransferTasksRequest,
) (*p.GetTransferTasksResponse, error) {

	// Reading transfer tasks need to be quorum level consistent, otherwise we could loose task
	query := d.session.Query(templateGetTransferTasksQuery,
		d.shardID,
		rowTypeTransferTask,
		rowTypeTransferDomainID,
		rowTypeTransferWorkflowID,
		rowTypeTransferRunID,
		defaultVisibilityTimestamp,
		request.ReadLevel,
		request.MaxReadLevel,
	).PageSize(request.BatchSize).PageState(request.NextPageToken).WithContext(ctx)

	iter := query.Iter()
	if iter == nil {
		return nil, &types.InternalServiceError{
			Message: "GetTransferTasks operation failed.  Not able to create query iterator.",
		}
	}

	response := &p.GetTransferTasksResponse{}
	task := make(map[string]interface{})
	for iter.MapScan(task) {
		t := createTransferTaskInfo(task["transfer"].(map[string]interface{}))
		// Reset task map to get it ready for next scan
		task = make(map[string]interface{})

		response.Tasks = append(response.Tasks, t)
	}
	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)

	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(d.client, "GetTransferTasks", err)
	}

	return response, nil
}

func (d *cassandraPersistence) GetCrossClusterTasks(
	ctx context.Context,
	request *p.GetCrossClusterTasksRequest,
) (*p.GetCrossClusterTasksResponse, error) {

	// Reading cross-cluster tasks need to be quorum level consistent, otherwise we could loose task
	query := d.session.Query(templateGetCrossClusterTasksQuery,
		d.shardID,
		rowTypeCrossClusterTask,
		rowTypeCrossClusterDomainID,
		request.TargetCluster, // workflowID field is used to store target cluster
		rowTypeCrossClusterRunID,
		defaultVisibilityTimestamp,
		request.ReadLevel,
		request.MaxReadLevel,
	).PageSize(request.BatchSize).PageState(request.NextPageToken).WithContext(ctx)

	iter := query.Iter()
	if iter == nil {
		return nil, &types.InternalServiceError{
			Message: "GetCrossClusterTasks operation failed.  Not able to create query iterator.",
		}
	}

	response := &p.GetCrossClusterTasksResponse{}
	task := make(map[string]interface{})
	for iter.MapScan(task) {
		t := createCrossClusterTaskInfo(task["cross_cluster"].(map[string]interface{}))
		// Reset task map to get it ready for next scan
		task = make(map[string]interface{})

		response.Tasks = append(response.Tasks, t)
	}
	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)

	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(d.client, "GetCrossClusterTasks", err)
	}

	return response, nil
}

func (d *cassandraPersistence) GetReplicationTasks(
	ctx context.Context,
	request *p.GetReplicationTasksRequest,
) (*p.InternalGetReplicationTasksResponse, error) {

	// Reading replication tasks need to be quorum level consistent, otherwise we could loose task
	query := d.session.Query(templateGetReplicationTasksQuery,
		d.shardID,
		rowTypeReplicationTask,
		rowTypeReplicationDomainID,
		rowTypeReplicationWorkflowID,
		rowTypeReplicationRunID,
		defaultVisibilityTimestamp,
		request.ReadLevel,
		request.MaxReadLevel,
	).PageSize(request.BatchSize).PageState(request.NextPageToken).WithContext(ctx)

	return d.populateGetReplicationTasksResponse(query)
}

func (d *cassandraPersistence) populateGetReplicationTasksResponse(
	query gocql.Query,
) (*p.InternalGetReplicationTasksResponse, error) {
	iter := query.Iter()
	if iter == nil {
		return nil, &types.InternalServiceError{
			Message: "GetReplicationTasks operation failed.  Not able to create query iterator.",
		}
	}

	response := &p.InternalGetReplicationTasksResponse{}
	task := make(map[string]interface{})
	for iter.MapScan(task) {
		t := createReplicationTaskInfo(task["replication"].(map[string]interface{}))
		// Reset task map to get it ready for next scan
		task = make(map[string]interface{})

		response.Tasks = append(response.Tasks, t)
	}
	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)

	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(d.client, "GetReplicationTasks", err)
	}

	return response, nil
}

func (d *cassandraPersistence) CompleteTransferTask(
	ctx context.Context,
	request *p.CompleteTransferTaskRequest,
) error {
	query := d.session.Query(templateCompleteTransferTaskQuery,
		d.shardID,
		rowTypeTransferTask,
		rowTypeTransferDomainID,
		rowTypeTransferWorkflowID,
		rowTypeTransferRunID,
		defaultVisibilityTimestamp,
		request.TaskID,
	).WithContext(ctx)

	err := query.Exec()
	if err != nil {
		return convertCommonErrors(d.client, "CompleteTransferTask", err)
	}

	return nil
}

func (d *cassandraPersistence) RangeCompleteTransferTask(
	ctx context.Context,
	request *p.RangeCompleteTransferTaskRequest,
) error {
	query := d.session.Query(templateRangeCompleteTransferTaskQuery,
		d.shardID,
		rowTypeTransferTask,
		rowTypeTransferDomainID,
		rowTypeTransferWorkflowID,
		rowTypeTransferRunID,
		defaultVisibilityTimestamp,
		request.ExclusiveBeginTaskID,
		request.InclusiveEndTaskID,
	).WithContext(ctx)

	err := query.Exec()
	if err != nil {
		return convertCommonErrors(d.client, "RangeCompleteTransferTask", err)
	}

	return nil
}

func (d *cassandraPersistence) CompleteCrossClusterTask(
	ctx context.Context,
	request *p.CompleteCrossClusterTaskRequest,
) error {
	query := d.session.Query(templateCompleteCrossClusterTaskQuery,
		d.shardID,
		rowTypeCrossClusterTask,
		rowTypeCrossClusterDomainID,
		request.TargetCluster,
		rowTypeCrossClusterRunID,
		defaultVisibilityTimestamp,
		request.TaskID,
	).WithContext(ctx)

	err := query.Exec()
	if err != nil {
		return convertCommonErrors(d.client, "CompleteCrossClusterTask", err)
	}

	return nil
}

func (d *cassandraPersistence) RangeCompleteCrossClusterTask(
	ctx context.Context,
	request *p.RangeCompleteCrossClusterTaskRequest,
) error {
	query := d.session.Query(templateRangeCompleteCrossClusterTaskQuery,
		d.shardID,
		rowTypeCrossClusterTask,
		rowTypeCrossClusterDomainID,
		request.TargetCluster,
		rowTypeCrossClusterRunID,
		defaultVisibilityTimestamp,
		request.ExclusiveBeginTaskID,
		request.InclusiveEndTaskID,
	).WithContext(ctx)

	err := query.Exec()
	if err != nil {
		return convertCommonErrors(d.client, "RangeCompleteCrossClusterTask", err)
	}

	return nil
}

func (d *cassandraPersistence) CompleteReplicationTask(
	ctx context.Context,
	request *p.CompleteReplicationTaskRequest,
) error {
	query := d.session.Query(templateCompleteReplicationTaskQuery,
		d.shardID,
		rowTypeReplicationTask,
		rowTypeReplicationDomainID,
		rowTypeReplicationWorkflowID,
		rowTypeReplicationRunID,
		defaultVisibilityTimestamp,
		request.TaskID,
	).WithContext(ctx)

	err := query.Exec()
	if err != nil {
		return convertCommonErrors(d.client, "CompleteReplicationTask", err)
	}

	return nil
}

func (d *cassandraPersistence) RangeCompleteReplicationTask(
	ctx context.Context,
	request *p.RangeCompleteReplicationTaskRequest,
) error {

	query := d.session.Query(templateCompleteReplicationTaskBeforeQuery,
		d.shardID,
		rowTypeReplicationTask,
		rowTypeReplicationDomainID,
		rowTypeReplicationWorkflowID,
		rowTypeReplicationRunID,
		defaultVisibilityTimestamp,
		request.InclusiveEndTaskID,
	).WithContext(ctx)

	err := query.Exec()
	if err != nil {
		return convertCommonErrors(d.client, "RangeCompleteReplicationTask", err)
	}

	return nil
}

func (d *cassandraPersistence) CompleteTimerTask(
	ctx context.Context,
	request *p.CompleteTimerTaskRequest,
) error {
	ts := p.UnixNanoToDBTimestamp(request.VisibilityTimestamp.UnixNano())
	query := d.session.Query(templateCompleteTimerTaskQuery,
		d.shardID,
		rowTypeTimerTask,
		rowTypeTimerDomainID,
		rowTypeTimerWorkflowID,
		rowTypeTimerRunID,
		ts,
		request.TaskID,
	).WithContext(ctx)

	err := query.Exec()
	if err != nil {
		return convertCommonErrors(d.client, "CompleteTimerTask", err)
	}

	return nil
}

func (d *cassandraPersistence) RangeCompleteTimerTask(
	ctx context.Context,
	request *p.RangeCompleteTimerTaskRequest,
) error {
	start := p.UnixNanoToDBTimestamp(request.InclusiveBeginTimestamp.UnixNano())
	end := p.UnixNanoToDBTimestamp(request.ExclusiveEndTimestamp.UnixNano())
	query := d.session.Query(templateRangeCompleteTimerTaskQuery,
		d.shardID,
		rowTypeTimerTask,
		rowTypeTimerDomainID,
		rowTypeTimerWorkflowID,
		rowTypeTimerRunID,
		start,
		end,
	).WithContext(ctx)

	err := query.Exec()
	if err != nil {
		return convertCommonErrors(d.client, "RangeCompleteTimerTask", err)
	}

	return nil
}

func (d *cassandraPersistence) GetTimerIndexTasks(
	ctx context.Context,
	request *p.GetTimerIndexTasksRequest,
) (*p.GetTimerIndexTasksResponse, error) {
	// Reading timer tasks need to be quorum level consistent, otherwise we could loose task
	minTimestamp := p.UnixNanoToDBTimestamp(request.MinTimestamp.UnixNano())
	maxTimestamp := p.UnixNanoToDBTimestamp(request.MaxTimestamp.UnixNano())
	query := d.session.Query(templateGetTimerTasksQuery,
		d.shardID,
		rowTypeTimerTask,
		rowTypeTimerDomainID,
		rowTypeTimerWorkflowID,
		rowTypeTimerRunID,
		minTimestamp,
		maxTimestamp,
	).PageSize(request.BatchSize).PageState(request.NextPageToken).WithContext(ctx)

	iter := query.Iter()
	if iter == nil {
		return nil, &types.InternalServiceError{
			Message: "GetTimerTasks operation failed.  Not able to create query iterator.",
		}
	}

	response := &p.GetTimerIndexTasksResponse{}
	task := make(map[string]interface{})
	for iter.MapScan(task) {
		t := createTimerTaskInfo(task["timer"].(map[string]interface{}))
		// Reset task map to get it ready for next scan
		task = make(map[string]interface{})

		response.Timers = append(response.Timers, t)
	}
	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)

	if err := iter.Close(); err != nil {
		return nil, convertCommonErrors(d.client, "GetTimerTasks", err)
	}

	return response, nil
}

func (d *cassandraPersistence) PutReplicationTaskToDLQ(
	ctx context.Context,
	request *p.InternalPutReplicationTaskToDLQRequest,
) error {
	task := request.TaskInfo

	// Use source cluster name as the workflow id for replication dlq
	query := d.session.Query(templateCreateReplicationTaskQuery,
		d.shardID,
		rowTypeDLQ,
		rowTypeDLQDomainID,
		request.SourceClusterName,
		rowTypeDLQRunID,
		task.DomainID,
		task.WorkflowID,
		task.RunID,
		task.TaskID,
		task.TaskType,
		task.FirstEventID,
		task.NextEventID,
		task.Version,
		task.ScheduledID,
		p.EventStoreVersion,
		task.BranchToken,
		p.EventStoreVersion,
		task.NewRunBranchToken,
		defaultVisibilityTimestamp,
		defaultVisibilityTimestamp,
		task.TaskID,
	).WithContext(ctx)

	err := query.Exec()
	if err != nil {
		return convertCommonErrors(d.client, "PutReplicationTaskToDLQ", err)
	}

	return nil
}

func (d *cassandraPersistence) GetReplicationTasksFromDLQ(
	ctx context.Context,
	request *p.GetReplicationTasksFromDLQRequest,
) (*p.InternalGetReplicationTasksFromDLQResponse, error) {
	// Reading replication tasks need to be quorum level consistent, otherwise we could loose task
	query := d.session.Query(templateGetReplicationTasksQuery,
		d.shardID,
		rowTypeDLQ,
		rowTypeDLQDomainID,
		request.SourceClusterName,
		rowTypeDLQRunID,
		defaultVisibilityTimestamp,
		request.ReadLevel,
		request.MaxReadLevel,
	).PageSize(request.BatchSize).PageState(request.NextPageToken).WithContext(ctx)

	return d.populateGetReplicationTasksResponse(query)
}

func (d *cassandraPersistence) GetReplicationDLQSize(
	ctx context.Context,
	request *p.GetReplicationDLQSizeRequest,
) (*p.GetReplicationDLQSizeResponse, error) {

	// Reading replication tasks need to be quorum level consistent, otherwise we could loose task
	query := d.session.Query(templateGetDLQSizeQuery,
		d.shardID,
		rowTypeDLQ,
		rowTypeDLQDomainID,
		request.SourceClusterName,
		rowTypeDLQRunID,
	).WithContext(ctx)

	result := make(map[string]interface{})
	if err := query.MapScan(result); err != nil {
		return nil, convertCommonErrors(d.client, "GetReplicationDLQSize", err)
	}

	queueSize := result["count"].(int64)
	return &p.GetReplicationDLQSizeResponse{
		Size: queueSize,
	}, nil
}

func (d *cassandraPersistence) DeleteReplicationTaskFromDLQ(
	ctx context.Context,
	request *p.DeleteReplicationTaskFromDLQRequest,
) error {

	query := d.session.Query(templateCompleteReplicationTaskQuery,
		d.shardID,
		rowTypeDLQ,
		rowTypeDLQDomainID,
		request.SourceClusterName,
		rowTypeDLQRunID,
		defaultVisibilityTimestamp,
		request.TaskID,
	).WithContext(ctx)

	err := query.Exec()
	if err != nil {
		return convertCommonErrors(d.client, "DeleteReplicationTaskFromDLQ", err)
	}

	return nil
}

func (d *cassandraPersistence) RangeDeleteReplicationTaskFromDLQ(
	ctx context.Context,
	request *p.RangeDeleteReplicationTaskFromDLQRequest,
) error {

	query := d.session.Query(templateRangeCompleteReplicationTaskQuery,
		d.shardID,
		rowTypeDLQ,
		rowTypeDLQDomainID,
		request.SourceClusterName,
		rowTypeDLQRunID,
		defaultVisibilityTimestamp,
		request.ExclusiveBeginTaskID,
		request.InclusiveEndTaskID,
	).WithContext(ctx)

	err := query.Exec()
	if err != nil {
		return convertCommonErrors(d.client, "RangeDeleteReplicationTaskFromDLQ", err)
	}

	return nil
}

func (d *cassandraPersistence) CreateFailoverMarkerTasks(
	ctx context.Context,
	request *p.CreateFailoverMarkersRequest,
) error {

	batch := d.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)
	for _, task := range request.Markers {
		t := []p.Task{task}
		if err := createReplicationTasks(
			batch,
			t,
			d.shardID,
			task.DomainID,
			rowTypeReplicationWorkflowID,
			rowTypeReplicationRunID,
		); err != nil {
			return err
		}
	}

	// Verifies that the RangeID has not changed
	batch.Query(templateUpdateLeaseQuery,
		request.RangeID,
		d.shardID,
		rowTypeShard,
		rowTypeShardDomainID,
		rowTypeShardWorkflowID,
		rowTypeShardRunID,
		defaultVisibilityTimestamp,
		rowTypeShardTaskID,
		request.RangeID,
	)

	previous := make(map[string]interface{})
	applied, iter, err := d.session.MapExecuteBatchCAS(batch, previous)
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()
	if err != nil {
		return convertCommonErrors(d.client, "CreateFailoverMarkerTasks", err)
	}

	if !applied {
		rowType, ok := previous["type"].(int)
		if !ok {
			// This should never happen, as all our rows have the type field.
			panic("Encounter row type not found")
		}
		if rowType == rowTypeShard {
			if rangeID, ok := previous["range_id"].(int64); ok && rangeID != request.RangeID {
				// CreateWorkflowExecution failed because rangeID was modified
				return &p.ShardOwnershipLostError{
					ShardID: d.shardID,
					Msg: fmt.Sprintf("Failed to create workflow execution.  Request RangeID: %v, Actual RangeID: %v",
						request.RangeID, rangeID),
				}
			}
		}
		return newShardOwnershipLostError(d.shardID, request.RangeID, previous)
	}
	return nil
}

func newShardOwnershipLostError(
	shardID int,
	rangeID int64,
	row map[string]interface{},
) error {
	// At this point we only know that the write was not applied.
	// It's much safer to return ShardOwnershipLostError as the default to force the application to reload
	// shard to recover from such errors
	var columns []string
	for k, v := range row {
		columns = append(columns, fmt.Sprintf("%s=%v", k, v))
	}
	return &p.ShardOwnershipLostError{
		ShardID: shardID,
		Msg: fmt.Sprintf("Failed to create workflow execution.  Request RangeID: %v, columns: (%v)",
			rangeID, strings.Join(columns, ",")),
	}
}

// TODO: remove this after all 2DC workflows complete
func createReplicationState(
	result map[string]interface{},
) *p.ReplicationState {

	if len(result) == 0 {
		return nil
	}

	info := &p.ReplicationState{}
	for k, v := range result {
		switch k {
		case "current_version":
			info.CurrentVersion = v.(int64)
		case "start_version":
			info.StartVersion = v.(int64)
		case "last_write_version":
			info.LastWriteVersion = v.(int64)
		case "last_write_event_id":
			info.LastWriteEventID = v.(int64)
		case "last_replication_info":
			info.LastReplicationInfo = make(map[string]*p.ReplicationInfo)
			replicationInfoMap := v.(map[string]map[string]interface{})
			for key, value := range replicationInfoMap {
				info.LastReplicationInfo[key] = createReplicationInfo(value)
			}
		}
	}

	return info
}

func mustConvertToSlice(value interface{}) []interface{} {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		result := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = v.Index(i).Interface()
		}
		return result
	default:
		panic(fmt.Sprintf("Unable to convert %v to slice", value))
	}
}
