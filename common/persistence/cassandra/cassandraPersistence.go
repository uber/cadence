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
	"strings"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/checksum"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
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

	templateCreateReplicationTaskQuery = `INSERT INTO executions (` +
		`shard_id, type, domain_id, workflow_id, run_id, replication, visibility_ts, task_id) ` +
		`VALUES(?, ?, ?, ?, ?, ` + templateReplicationTaskType + `, ?, ?)`

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

	templateDeleteWorkflowExecutionMutableStateQuery = `DELETE FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? `

	templateDeleteWorkflowExecutionCurrentRowQuery = templateDeleteWorkflowExecutionMutableStateQuery + " if current_run_id = ? "

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

	currentWorkflowWriteReq, err := d.prepareCurrentWorkflowRequestForCreateWorkflowTxn(domainID, workflowID, runID, executionInfo, lastWriteVersion, request)
	if err != nil {
		return nil, err
	}

	workflowExecutionWriteReq, err := d.prepareCreateWorkflowExecutionRequestWithMaps(&newWorkflow)
	if err != nil {
		return nil, err
	}

	transferTasks, crossClusterTasks, replicationTasks, timerTasks, err := d.prepareNoSQLTasksForWorkflowTxn(
		domainID, workflowID, runID,
		newWorkflow.TransferTasks, newWorkflow.CrossClusterTasks, newWorkflow.ReplicationTasks, newWorkflow.TimerTasks,
		nil, nil, nil, nil,
	)
	if err != nil {
		return nil, err
	}

	shardCondition := &nosqlplugin.ShardCondition{
		ShardID: d.shardID,
		RangeID: request.RangeID,
	}

	err = d.db.InsertWorkflowExecutionWithTasks(
		ctx,
		currentWorkflowWriteReq, workflowExecutionWriteReq,
		transferTasks, crossClusterTasks, replicationTasks, timerTasks,
		shardCondition,
	)
	if err != nil {
		conditionFailureErr, isConditionFailedError := err.(*nosqlplugin.WorkflowOperationConditionFailure)
		if isConditionFailedError {
			switch {
			case conditionFailureErr.UnknownConditionFailureDetails != nil:
				return nil, &p.ShardOwnershipLostError{
					ShardID: d.shardID,
					Msg:     *conditionFailureErr.UnknownConditionFailureDetails,
				}
			case conditionFailureErr.ShardRangeIDNotMatch != nil:
				return nil, &p.ShardOwnershipLostError{
					ShardID: d.shardID,
					Msg: fmt.Sprintf("Failed to create workflow execution.  Request RangeID: %v, Actual RangeID: %v",
						request.RangeID, *conditionFailureErr.ShardRangeIDNotMatch),
				}
			case conditionFailureErr.CurrentWorkflowConditionFailInfo != nil:
				return nil, &p.CurrentWorkflowConditionFailedError{
					Msg: *conditionFailureErr.CurrentWorkflowConditionFailInfo,
				}
			case conditionFailureErr.WorkflowExecutionAlreadyExists != nil:
				return nil, &p.WorkflowExecutionAlreadyStartedError{
					Msg:              conditionFailureErr.WorkflowExecutionAlreadyExists.OtherInfo,
					StartRequestID:   conditionFailureErr.WorkflowExecutionAlreadyExists.CreateRequestID,
					RunID:            conditionFailureErr.WorkflowExecutionAlreadyExists.RunID,
					State:            conditionFailureErr.WorkflowExecutionAlreadyExists.State,
					CloseStatus:      conditionFailureErr.WorkflowExecutionAlreadyExists.CloseStatus,
					LastWriteVersion: conditionFailureErr.WorkflowExecutionAlreadyExists.LastWriteVersion,
				}
			default:
				// If ever runs into this branch, there is bug in the code either in here, or in the implementation of nosql plugin
				err := fmt.Errorf("unsupported conditionFailureReason error")
				d.logger.Error("A code bug exists in persistence layer, please investigate ASAP", tag.Error(err))
				return nil, err
			}
		}
		return nil, convertCommonErrors(d.client, "CreateWorkflowExecution", err)
	}

	return &p.CreateWorkflowExecutionResponse{}, nil
}

func (d *cassandraPersistence) prepareCreateWorkflowExecutionRequestWithMaps(
	newWorkflow *p.InternalWorkflowSnapshot,
) (*nosqlplugin.WorkflowExecutionRequest, error) {
	executionInfo := newWorkflow.ExecutionInfo
	lastWriteVersion := newWorkflow.LastWriteVersion
	checkSum := newWorkflow.Checksum
	versionHistories := newWorkflow.VersionHistories
	nowTimestamp := time.Now()

	executionRequest, err := d.prepareCreateWorkflowExecutionTxn(
		executionInfo, versionHistories, checkSum,
		nowTimestamp, lastWriteVersion,
	)
	if err != nil {
		return nil, err
	}

	executionRequest.ActivityInfos, err = d.prepareActivityInfosForWorkflowTxn(newWorkflow.ActivityInfos)
	if err != nil {
		return nil, err
	}
	executionRequest.TimerInfos, err = d.prepareTimerInfosForWorkflowTxn(newWorkflow.TimerInfos)
	if err != nil {
		return nil, err
	}
	executionRequest.ChildWorkflowInfos, err = d.prepareChildWFInfosForWorkflowTxn(newWorkflow.ChildExecutionInfos)
	if err != nil {
		return nil, err
	}
	executionRequest.RequestCancelInfos, err = d.prepareRequestCancelsForWorkflowTxn(newWorkflow.RequestCancelInfos)
	if err != nil {
		return nil, err
	}
	executionRequest.SignalInfos, err = d.prepareSignalInfosForWorkflowTxn(newWorkflow.SignalInfos)
	if err != nil {
		return nil, err
	}
	executionRequest.SignalRequestedIDs = newWorkflow.SignalRequestedIDs
	executionRequest.MapsWriteMode = nosqlplugin.WorkflowExecutionMapsWriteModeCreate
	return executionRequest, nil
}

func (d *cassandraPersistence) prepareResetWorkflowExecutionRequestWithMapsAndEventBuffer(
	resetWorkflow *p.InternalWorkflowSnapshot,
) (*nosqlplugin.WorkflowExecutionRequest, error) {
	executionInfo := resetWorkflow.ExecutionInfo
	lastWriteVersion := resetWorkflow.LastWriteVersion
	checkSum := resetWorkflow.Checksum
	versionHistories := resetWorkflow.VersionHistories
	nowTimestamp := time.Now()

	executionRequest, err := d.prepareUpdateWorkflowExecutionTxn(
		executionInfo, versionHistories, checkSum,
		nowTimestamp, lastWriteVersion,
	)
	if err != nil {
		return nil, err
	}
	//reset 6 maps
	executionRequest.ActivityInfos, err = d.prepareActivityInfosForWorkflowTxn(resetWorkflow.ActivityInfos)
	if err != nil {
		return nil, err
	}
	executionRequest.TimerInfos, err = d.prepareTimerInfosForWorkflowTxn(resetWorkflow.TimerInfos)
	if err != nil {
		return nil, err
	}
	executionRequest.ChildWorkflowInfos, err = d.prepareChildWFInfosForWorkflowTxn(resetWorkflow.ChildExecutionInfos)
	if err != nil {
		return nil, err
	}
	executionRequest.RequestCancelInfos, err = d.prepareRequestCancelsForWorkflowTxn(resetWorkflow.RequestCancelInfos)
	if err != nil {
		return nil, err
	}
	executionRequest.SignalInfos, err = d.prepareSignalInfosForWorkflowTxn(resetWorkflow.SignalInfos)
	if err != nil {
		return nil, err
	}
	executionRequest.SignalRequestedIDs = resetWorkflow.SignalRequestedIDs
	executionRequest.MapsWriteMode = nosqlplugin.WorkflowExecutionMapsWriteModeReset
	// delete buffered events
	executionRequest.EventBufferWriteMode = nosqlplugin.EventBufferWriteModeClear
	// condition
	executionRequest.PreviousNextEventIDCondition = &resetWorkflow.Condition
	return executionRequest, nil
}

func (d *cassandraPersistence) prepareUpdateWorkflowExecutionRequestWithMapsAndEventBuffer(
	workflowMutation *p.InternalWorkflowMutation,
) (*nosqlplugin.WorkflowExecutionRequest, error) {
	executionInfo := workflowMutation.ExecutionInfo
	lastWriteVersion := workflowMutation.LastWriteVersion
	checkSum := workflowMutation.Checksum
	versionHistories := workflowMutation.VersionHistories
	nowTimestamp := time.Now()

	executionRequest, err := d.prepareUpdateWorkflowExecutionTxn(
		executionInfo, versionHistories, checkSum,
		nowTimestamp, lastWriteVersion,
	)
	if err != nil {
		return nil, err
	}

	// merge 6 maps
	executionRequest.ActivityInfos, err = d.prepareActivityInfosForWorkflowTxn(workflowMutation.UpsertActivityInfos)
	if err != nil {
		return nil, err
	}
	executionRequest.TimerInfos, err = d.prepareTimerInfosForWorkflowTxn(workflowMutation.UpsertTimerInfos)
	if err != nil {
		return nil, err
	}
	executionRequest.ChildWorkflowInfos, err = d.prepareChildWFInfosForWorkflowTxn(workflowMutation.UpsertChildExecutionInfos)
	if err != nil {
		return nil, err
	}
	executionRequest.RequestCancelInfos, err = d.prepareRequestCancelsForWorkflowTxn(workflowMutation.UpsertRequestCancelInfos)
	if err != nil {
		return nil, err
	}
	executionRequest.SignalInfos, err = d.prepareSignalInfosForWorkflowTxn(workflowMutation.UpsertSignalInfos)
	if err != nil {
		return nil, err
	}
	executionRequest.SignalRequestedIDs = workflowMutation.UpsertSignalRequestedIDs

	// delete from 6 maps
	executionRequest.ActivityInfoKeysToDelete = workflowMutation.DeleteActivityInfos
	executionRequest.TimerInfoKeysToDelete = workflowMutation.DeleteTimerInfos
	executionRequest.ChildWorkflowInfoKeysToDelete = workflowMutation.DeleteChildExecutionInfos
	executionRequest.RequestCancelInfoKeysToDelete = workflowMutation.DeleteRequestCancelInfos
	executionRequest.SignalInfoKeysToDelete = workflowMutation.DeleteSignalInfos
	executionRequest.SignalRequestedIDsKeysToDelete = workflowMutation.DeleteSignalRequestedIDs

	// map write mode
	executionRequest.MapsWriteMode = nosqlplugin.WorkflowExecutionMapsWriteModeUpdate

	// prepare to write buffer event
	executionRequest.EventBufferWriteMode = nosqlplugin.EventBufferWriteModeNone
	if workflowMutation.ClearBufferedEvents {
		executionRequest.EventBufferWriteMode = nosqlplugin.EventBufferWriteModeClear
	} else if workflowMutation.NewBufferedEvents != nil {
		executionRequest.EventBufferWriteMode = nosqlplugin.EventBufferWriteModeAppend
		executionRequest.NewBufferedEventBatch = workflowMutation.NewBufferedEvents
	}

	// condition
	executionRequest.PreviousNextEventIDCondition = &workflowMutation.Condition
	return executionRequest, nil
}

func (d *cassandraPersistence) prepareTimerTasksForWorkflowTxn(
	domainID, workflowID, runID string,
	timerTasks []p.Task,
) ([]*nosqlplugin.TimerTask, error) {
	var tasks []*nosqlplugin.TimerTask

	for _, task := range timerTasks {
		var eventID int64
		var attempt int64

		timeoutType := 0

		switch t := task.(type) {
		case *p.DecisionTimeoutTask:
			eventID = t.EventID
			timeoutType = t.TimeoutType
			attempt = t.ScheduleAttempt

		case *p.ActivityTimeoutTask:
			eventID = t.EventID
			timeoutType = t.TimeoutType
			attempt = t.Attempt

		case *p.UserTimerTask:
			eventID = t.EventID

		case *p.ActivityRetryTimerTask:
			eventID = t.EventID
			attempt = int64(t.Attempt)

		case *p.WorkflowBackoffTimerTask:
			eventID = t.EventID
			timeoutType = t.TimeoutType

		case *p.WorkflowTimeoutTask:
			// noop

		case *p.DeleteHistoryEventTask:
			// noop

		default:
			return nil, &types.InternalServiceError{
				Message: fmt.Sprintf("Unknow timer type: %v", task.GetType()),
			}
		}

		nt := &nosqlplugin.TimerTask{
			Type:       task.GetType(),
			DomainID:   domainID,
			WorkflowID: workflowID,
			RunID:      runID,

			VisibilityTimestamp: task.GetVisibilityTimestamp(),
			TaskID:              task.GetTaskID(),

			TimeoutType: timeoutType,
			EventID:     eventID,
			Attempt:     attempt,
			Version:     task.GetVersion(),
		}
		tasks = append(tasks, nt)
	}

	return tasks, nil
}

func (d *cassandraPersistence) prepareReplicationTasksForWorkflowTxn(
	domainID, workflowID, runID string,
	replicationTasks []p.Task,
) ([]*nosqlplugin.ReplicationTask, error) {
	var tasks []*nosqlplugin.ReplicationTask

	for _, task := range replicationTasks {
		// Replication task specific information
		firstEventID := common.EmptyEventID
		nextEventID := common.EmptyEventID
		version := common.EmptyVersion //nolint:ineffassign
		activityScheduleID := common.EmptyEventID
		var branchToken, newRunBranchToken []byte

		switch task.GetType() {
		case p.ReplicationTaskTypeHistory:
			histTask := task.(*p.HistoryReplicationTask)
			branchToken = histTask.BranchToken
			newRunBranchToken = histTask.NewRunBranchToken
			firstEventID = histTask.FirstEventID
			nextEventID = histTask.NextEventID
			version = task.GetVersion()

		case p.ReplicationTaskTypeSyncActivity:
			version = task.GetVersion()
			activityScheduleID = task.(*p.SyncActivityTask).ScheduledID

		case p.ReplicationTaskTypeFailoverMarker:
			version = task.GetVersion()

		default:
			return nil, &types.InternalServiceError{
				Message: fmt.Sprintf("Unknow replication type: %v", task.GetType()),
			}
		}

		nt := &nosqlplugin.ReplicationTask{
			Type:                task.GetType(),
			DomainID:            domainID,
			WorkflowID:          workflowID,
			RunID:               runID,
			VisibilityTimestamp: task.GetVisibilityTimestamp(),
			TaskID:              task.GetTaskID(),
			FirstEventID:        firstEventID,
			NextEventID:         nextEventID,
			Version:             version,
			ActivityScheduleID:  activityScheduleID,
			EventStoreVersion:   p.EventStoreVersion,
			BranchToken:         branchToken,
			NewRunBranchToken:   newRunBranchToken,
		}
		tasks = append(tasks, nt)
	}

	return tasks, nil
}

func (d *cassandraPersistence) prepareCrossClusterTasksForWorkflowTxn(
	domainID, workflowID, runID string,
	crossClusterTasks []p.Task,
) ([]*nosqlplugin.CrossClusterTask, error) {
	var tasks []*nosqlplugin.CrossClusterTask

	for _, task := range crossClusterTasks {
		var taskList string
		var scheduleID int64
		var targetCluster string
		var targetDomainID string
		var targetWorkflowID string
		targetRunID := p.CrossClusterTaskDefaultTargetRunID
		targetChildWorkflowOnly := false
		recordVisibility := false

		switch task.GetType() {
		case p.CrossClusterTaskTypeStartChildExecution:
			targetCluster = task.(*p.CrossClusterStartChildExecutionTask).TargetCluster
			targetDomainID = task.(*p.CrossClusterStartChildExecutionTask).TargetDomainID
			targetWorkflowID = task.(*p.CrossClusterStartChildExecutionTask).TargetWorkflowID
			scheduleID = task.(*p.CrossClusterStartChildExecutionTask).InitiatedID

		case p.CrossClusterTaskTypeCancelExecution:
			targetCluster = task.(*p.CrossClusterCancelExecutionTask).TargetCluster
			targetDomainID = task.(*p.CrossClusterCancelExecutionTask).TargetDomainID
			targetWorkflowID = task.(*p.CrossClusterCancelExecutionTask).TargetWorkflowID
			targetRunID = task.(*p.CrossClusterCancelExecutionTask).TargetRunID
			targetChildWorkflowOnly = task.(*p.CrossClusterCancelExecutionTask).TargetChildWorkflowOnly
			scheduleID = task.(*p.CrossClusterCancelExecutionTask).InitiatedID

		case p.CrossClusterTaskTypeSignalExecution:
			targetCluster = task.(*p.CrossClusterSignalExecutionTask).TargetCluster
			targetDomainID = task.(*p.CrossClusterSignalExecutionTask).TargetDomainID
			targetWorkflowID = task.(*p.CrossClusterSignalExecutionTask).TargetWorkflowID
			targetRunID = task.(*p.CrossClusterSignalExecutionTask).TargetRunID
			targetChildWorkflowOnly = task.(*p.CrossClusterSignalExecutionTask).TargetChildWorkflowOnly
			scheduleID = task.(*p.CrossClusterSignalExecutionTask).InitiatedID

		default:
			return nil, &types.InternalServiceError{
				Message: fmt.Sprintf("Unknown cross-cluster task type: %v", task.GetType()),
			}
		}

		nt := &nosqlplugin.CrossClusterTask{
			TransferTask: nosqlplugin.TransferTask{
				Type:                    task.GetType(),
				DomainID:                domainID,
				WorkflowID:              workflowID,
				RunID:                   runID,
				VisibilityTimestamp:     task.GetVisibilityTimestamp(),
				TaskID:                  task.GetTaskID(),
				TargetDomainID:          targetDomainID,
				TargetWorkflowID:        targetWorkflowID,
				TargetRunID:             targetRunID,
				TargetChildWorkflowOnly: targetChildWorkflowOnly,
				TaskList:                taskList,
				ScheduleID:              scheduleID,
				RecordVisibility:        recordVisibility,
				Version:                 task.GetVersion(),
			},
			TargetCluster: targetCluster,
		}
		tasks = append(tasks, nt)
	}

	return tasks, nil
}

func (d *cassandraPersistence) prepareNoSQLTasksForWorkflowTxn(
	domainID, workflowID, runID string,
	persistenceTransferTasks, persistenceCrossClusterTasks, persistenceReplicationTasks, persistenceTimerTasks []p.Task,
	transferTasksToAppend []*nosqlplugin.TransferTask,
	crossClusterTasksToAppend []*nosqlplugin.CrossClusterTask,
	replicationTasksToAppend []*nosqlplugin.ReplicationTask,
	timerTasksToAppend []*nosqlplugin.TimerTask,
) ([]*nosqlplugin.TransferTask, []*nosqlplugin.CrossClusterTask, []*nosqlplugin.ReplicationTask, []*nosqlplugin.TimerTask, error) {
	transferTasks, err := d.prepareTransferTasksForWorkflowTxn(domainID, workflowID, runID, persistenceTransferTasks)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	transferTasksToAppend = append(transferTasksToAppend, transferTasks...)

	crossClusterTasks, err := d.prepareCrossClusterTasksForWorkflowTxn(domainID, workflowID, runID, persistenceCrossClusterTasks)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	crossClusterTasksToAppend = append(crossClusterTasksToAppend, crossClusterTasks...)

	replicationTasks, err := d.prepareReplicationTasksForWorkflowTxn(domainID, workflowID, runID, persistenceReplicationTasks)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	replicationTasksToAppend = append(replicationTasksToAppend, replicationTasks...)

	timerTasks, err := d.prepareTimerTasksForWorkflowTxn(domainID, workflowID, runID, persistenceTimerTasks)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	timerTasksToAppend = append(timerTasksToAppend, timerTasks...)
	return transferTasksToAppend, crossClusterTasksToAppend, replicationTasksToAppend, timerTasksToAppend, nil
}

func (d *cassandraPersistence) prepareTransferTasksForWorkflowTxn(
	domainID, workflowID, runID string,
	transferTasks []p.Task,
) ([]*nosqlplugin.TransferTask, error) {
	var tasks []*nosqlplugin.TransferTask

	targetDomainID := domainID
	for _, task := range transferTasks {
		var taskList string
		var scheduleID int64
		targetWorkflowID := p.TransferTaskTransferTargetWorkflowID
		targetRunID := p.TransferTaskTransferTargetRunID
		targetChildWorkflowOnly := false
		recordVisibility := false

		switch task.GetType() {
		case p.TransferTaskTypeActivityTask:
			targetDomainID = task.(*p.ActivityTask).DomainID
			taskList = task.(*p.ActivityTask).TaskList
			scheduleID = task.(*p.ActivityTask).ScheduleID

		case p.TransferTaskTypeDecisionTask:
			targetDomainID = task.(*p.DecisionTask).DomainID
			taskList = task.(*p.DecisionTask).TaskList
			scheduleID = task.(*p.DecisionTask).ScheduleID
			recordVisibility = task.(*p.DecisionTask).RecordVisibility

		case p.TransferTaskTypeCancelExecution:
			targetDomainID = task.(*p.CancelExecutionTask).TargetDomainID
			targetWorkflowID = task.(*p.CancelExecutionTask).TargetWorkflowID
			targetRunID = task.(*p.CancelExecutionTask).TargetRunID
			if targetRunID == "" {
				targetRunID = p.TransferTaskTransferTargetRunID
			}
			targetChildWorkflowOnly = task.(*p.CancelExecutionTask).TargetChildWorkflowOnly
			scheduleID = task.(*p.CancelExecutionTask).InitiatedID

		case p.TransferTaskTypeSignalExecution:
			targetDomainID = task.(*p.SignalExecutionTask).TargetDomainID
			targetWorkflowID = task.(*p.SignalExecutionTask).TargetWorkflowID
			targetRunID = task.(*p.SignalExecutionTask).TargetRunID
			if targetRunID == "" {
				targetRunID = p.TransferTaskTransferTargetRunID
			}
			targetChildWorkflowOnly = task.(*p.SignalExecutionTask).TargetChildWorkflowOnly
			scheduleID = task.(*p.SignalExecutionTask).InitiatedID

		case p.TransferTaskTypeStartChildExecution:
			targetDomainID = task.(*p.StartChildExecutionTask).TargetDomainID
			targetWorkflowID = task.(*p.StartChildExecutionTask).TargetWorkflowID
			scheduleID = task.(*p.StartChildExecutionTask).InitiatedID

		case p.TransferTaskTypeCloseExecution,
			p.TransferTaskTypeRecordWorkflowStarted,
			p.TransferTaskTypeResetWorkflow,
			p.TransferTaskTypeUpsertWorkflowSearchAttributes:
			// No explicit property needs to be set
		default:
			return nil, &types.InternalServiceError{
				Message: fmt.Sprintf("Unknown transfer type: %v", task.GetType()),
			}
		}
		t := &nosqlplugin.TransferTask{
			Type:                    task.GetType(),
			DomainID:                domainID,
			WorkflowID:              workflowID,
			RunID:                   runID,
			VisibilityTimestamp:     task.GetVisibilityTimestamp(),
			TaskID:                  task.GetTaskID(),
			TargetDomainID:          targetDomainID,
			TargetWorkflowID:        targetWorkflowID,
			TargetRunID:             targetRunID,
			TargetChildWorkflowOnly: targetChildWorkflowOnly,
			TaskList:                taskList,
			ScheduleID:              scheduleID,
			RecordVisibility:        recordVisibility,
			Version:                 task.GetVersion(),
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (d *cassandraPersistence) prepareActivityInfosForWorkflowTxn(
	activityInfos []*p.InternalActivityInfo,
) (map[int64]*p.InternalActivityInfo, error) {
	m := map[int64]*p.InternalActivityInfo{}
	for _, a := range activityInfos {
		_, scheduleEncoding := p.FromDataBlob(a.ScheduledEvent)
		_, startEncoding := p.FromDataBlob(a.StartedEvent)
		if a.StartedEvent != nil && scheduleEncoding != startEncoding {
			return nil, p.NewCadenceSerializationError(fmt.Sprintf("expect to have the same encoding, but %v != %v", scheduleEncoding, startEncoding))
		}
		a.ScheduledEvent = a.ScheduledEvent.ToNilSafeDataBlob()
		a.StartedEvent = a.StartedEvent.ToNilSafeDataBlob()
		m[a.ScheduleID] = a
	}
	return m, nil
}

func (d *cassandraPersistence) prepareTimerInfosForWorkflowTxn(
	timerInfo []*p.TimerInfo,
) (map[string]*p.TimerInfo, error) {
	m := map[string]*p.TimerInfo{}
	for _, a := range timerInfo {
		m[a.TimerID] = a
	}
	return m, nil
}

func (d *cassandraPersistence) prepareChildWFInfosForWorkflowTxn(
	childWFInfos []*p.InternalChildExecutionInfo,
) (map[int64]*p.InternalChildExecutionInfo, error) {
	m := map[int64]*p.InternalChildExecutionInfo{}
	for _, c := range childWFInfos {
		_, initiatedEncoding := p.FromDataBlob(c.InitiatedEvent)
		_, startEncoding := p.FromDataBlob(c.StartedEvent)
		if c.StartedEvent != nil && initiatedEncoding != startEncoding {
			return nil, p.NewCadenceSerializationError(fmt.Sprintf("expect to have the same encoding, but %v != %v", initiatedEncoding, startEncoding))
		}

		if c.StartedRunID == "" {
			c.StartedRunID = emptyRunID
		}

		c.InitiatedEvent = c.InitiatedEvent.ToNilSafeDataBlob()
		c.StartedEvent = c.StartedEvent.ToNilSafeDataBlob()
		m[c.InitiatedID] = c
	}
	return m, nil
}

func (d *cassandraPersistence) prepareRequestCancelsForWorkflowTxn(
	requestCancels []*p.RequestCancelInfo,
) (map[int64]*p.RequestCancelInfo, error) {
	m := map[int64]*p.RequestCancelInfo{}
	for _, c := range requestCancels {
		m[c.InitiatedID] = c
	}
	return m, nil
}

func (d *cassandraPersistence) prepareSignalInfosForWorkflowTxn(
	signalInfos []*p.SignalInfo,
) (map[int64]*p.SignalInfo, error) {
	m := map[int64]*p.SignalInfo{}
	for _, c := range signalInfos {
		m[c.InitiatedID] = c
	}
	return m, nil
}

func (d *cassandraPersistence) prepareUpdateWorkflowExecutionTxn(
	executionInfo *p.InternalWorkflowExecutionInfo,
	versionHistories *p.DataBlob,
	checksum checksum.Checksum,
	nowTimestamp time.Time,
	lastWriteVersion int64,
) (*nosqlplugin.WorkflowExecutionRequest, error) {
	// validate workflow state & close status
	if err := p.ValidateUpdateWorkflowStateCloseStatus(
		executionInfo.State,
		executionInfo.CloseStatus); err != nil {
		return nil, err
	}

	if executionInfo.ParentDomainID == "" {
		executionInfo.ParentDomainID = emptyDomainID
		executionInfo.ParentWorkflowID = ""
		executionInfo.ParentRunID = emptyRunID
		executionInfo.InitiatedID = emptyInitiatedID
	}

	// TODO we should set the last update time on business logic layer
	executionInfo.LastUpdatedTimestamp = nowTimestamp

	executionInfo.CompletionEvent = executionInfo.CompletionEvent.ToNilSafeDataBlob()
	executionInfo.AutoResetPoints = executionInfo.AutoResetPoints.ToNilSafeDataBlob()
	// TODO also need to set the start / current / last write version
	versionHistories = versionHistories.ToNilSafeDataBlob()
	return &nosqlplugin.WorkflowExecutionRequest{
		InternalWorkflowExecutionInfo: *executionInfo,
		VersionHistories:              versionHistories,
		Checksums:                     &checksum,
		LastWriteVersion:              lastWriteVersion,
	}, nil
}

func (d *cassandraPersistence) prepareCreateWorkflowExecutionTxn(
	executionInfo *p.InternalWorkflowExecutionInfo,
	versionHistories *p.DataBlob,
	checksum checksum.Checksum,
	nowTimestamp time.Time,
	lastWriteVersion int64,
) (*nosqlplugin.WorkflowExecutionRequest, error) {
	// validate workflow state & close status
	if err := p.ValidateCreateWorkflowStateCloseStatus(
		executionInfo.State,
		executionInfo.CloseStatus); err != nil {
		return nil, err
	}

	if executionInfo.ParentDomainID == "" {
		executionInfo.ParentDomainID = emptyDomainID
		executionInfo.ParentWorkflowID = ""
		executionInfo.ParentRunID = emptyRunID
		executionInfo.InitiatedID = emptyInitiatedID
	}

	// TODO we should set the start time and last update time on business logic layer
	executionInfo.StartTimestamp = nowTimestamp
	executionInfo.LastUpdatedTimestamp = nowTimestamp
	executionInfo.CompletionEvent = executionInfo.CompletionEvent.ToNilSafeDataBlob()
	executionInfo.AutoResetPoints = executionInfo.AutoResetPoints.ToNilSafeDataBlob()
	if versionHistories == nil {
		return nil, &types.InternalServiceError{Message: "encounter empty version histories in createExecution"}
	}
	versionHistories = versionHistories.ToNilSafeDataBlob()
	return &nosqlplugin.WorkflowExecutionRequest{
		InternalWorkflowExecutionInfo: *executionInfo,
		VersionHistories:              versionHistories,
		Checksums:                     &checksum,
		LastWriteVersion:              lastWriteVersion,
	}, nil
}

func (d *cassandraPersistence) prepareCurrentWorkflowRequestForCreateWorkflowTxn(
	domainID, workflowID, runID string,
	executionInfo *p.InternalWorkflowExecutionInfo,
	lastWriteVersion int64,
	request *p.InternalCreateWorkflowExecutionRequest,
) (*nosqlplugin.CurrentWorkflowWriteRequest, error) {
	currentWorkflowWriteReq := &nosqlplugin.CurrentWorkflowWriteRequest{
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
	return currentWorkflowWriteReq, nil
}

func (d *cassandraPersistence) GetWorkflowExecution(
	ctx context.Context,
	request *p.InternalGetWorkflowExecutionRequest,
) (*p.InternalGetWorkflowExecutionResponse, error) {

	execution := request.Execution
	state, err := d.db.SelectWorkflowExecution(ctx, d.shardID, request.DomainID, execution.WorkflowID, execution.RunID)
	if err != nil {
		if d.client.IsNotFoundError(err) {
			return nil, &types.EntityNotExistsError{
				Message: fmt.Sprintf("Workflow execution not found.  WorkflowId: %v, RunId: %v",
					execution.WorkflowID, execution.RunID),
			}
		}

		return nil, convertCommonErrors(d.client, "GetWorkflowExecution", err)
	}

	return &p.InternalGetWorkflowExecutionResponse{State: state}, nil
}

func (d *cassandraPersistence) UpdateWorkflowExecution(
	ctx context.Context,
	request *p.InternalUpdateWorkflowExecutionRequest,
) error {
	updateWorkflow := request.UpdateWorkflowMutation
	newWorkflow := request.NewWorkflowSnapshot

	executionInfo := updateWorkflow.ExecutionInfo
	domainID := executionInfo.DomainID
	workflowID := executionInfo.WorkflowID
	runID := executionInfo.RunID

	if err := p.ValidateUpdateWorkflowModeState(
		request.Mode,
		updateWorkflow,
		newWorkflow,
	); err != nil {
		return err
	}

	var currentWorkflowWriteReq *nosqlplugin.CurrentWorkflowWriteRequest

	switch request.Mode {
	case p.UpdateWorkflowModeBypassCurrent:
		if err := d.assertNotCurrentExecution(
			ctx,
			domainID,
			workflowID,
			runID); err != nil {
			return err
		}
		currentWorkflowWriteReq = &nosqlplugin.CurrentWorkflowWriteRequest{
			WriteMode: nosqlplugin.CurrentWorkflowWriteModeNoop,
		}

	case p.UpdateWorkflowModeUpdateCurrent:
		if newWorkflow != nil {
			newExecutionInfo := newWorkflow.ExecutionInfo
			newLastWriteVersion := newWorkflow.LastWriteVersion
			newDomainID := newExecutionInfo.DomainID
			// TODO: ?? would it change at all ??
			newWorkflowID := newExecutionInfo.WorkflowID
			newRunID := newExecutionInfo.RunID

			if domainID != newDomainID {
				return &types.InternalServiceError{
					Message: fmt.Sprintf("UpdateWorkflowExecution: cannot continue as new to another domain"),
				}
			}

			currentWorkflowWriteReq = &nosqlplugin.CurrentWorkflowWriteRequest{
				WriteMode: nosqlplugin.CurrentWorkflowWriteModeUpdate,
				Row: nosqlplugin.CurrentWorkflowRow{
					ShardID:          d.shardID,
					DomainID:         newDomainID,
					WorkflowID:       newWorkflowID,
					RunID:            newRunID,
					State:            newExecutionInfo.State,
					CloseStatus:      newExecutionInfo.CloseStatus,
					CreateRequestID:  newExecutionInfo.CreateRequestID,
					LastWriteVersion: newLastWriteVersion,
				},
				Condition: &nosqlplugin.CurrentWorkflowWriteCondition{
					CurrentRunID: &runID,
				},
			}
		} else {
			lastWriteVersion := updateWorkflow.LastWriteVersion

			currentWorkflowWriteReq = &nosqlplugin.CurrentWorkflowWriteRequest{
				WriteMode: nosqlplugin.CurrentWorkflowWriteModeUpdate,
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
				Condition: &nosqlplugin.CurrentWorkflowWriteCondition{
					CurrentRunID: &runID,
				},
			}
		}

	default:
		return &types.InternalServiceError{
			Message: fmt.Sprintf("UpdateWorkflowExecution: unknown mode: %v", request.Mode),
		}
	}

	var mutateExecution, insertExecution *nosqlplugin.WorkflowExecutionRequest
	var nosqlTransferTasks []*nosqlplugin.TransferTask
	var nosqlCrossClusterTasks []*nosqlplugin.CrossClusterTask
	var nosqlReplicationTasks []*nosqlplugin.ReplicationTask
	var nosqlTimerTasks []*nosqlplugin.TimerTask
	var err error

	// 1. current
	mutateExecution, err = d.prepareUpdateWorkflowExecutionRequestWithMapsAndEventBuffer(&updateWorkflow)
	if err != nil {
		return err
	}
	nosqlTransferTasks, nosqlCrossClusterTasks, nosqlReplicationTasks, nosqlTimerTasks, err = d.prepareNoSQLTasksForWorkflowTxn(
		domainID, workflowID, updateWorkflow.ExecutionInfo.RunID,
		updateWorkflow.TransferTasks, updateWorkflow.CrossClusterTasks, updateWorkflow.ReplicationTasks, updateWorkflow.TimerTasks,
		nosqlTransferTasks, nosqlCrossClusterTasks, nosqlReplicationTasks, nosqlTimerTasks,
	)
	if err != nil {
		return err
	}

	// 2. new
	if newWorkflow != nil {
		insertExecution, err = d.prepareCreateWorkflowExecutionRequestWithMaps(newWorkflow)

		nosqlTransferTasks, nosqlCrossClusterTasks, nosqlReplicationTasks, nosqlTimerTasks, err = d.prepareNoSQLTasksForWorkflowTxn(
			domainID, workflowID, newWorkflow.ExecutionInfo.RunID,
			newWorkflow.TransferTasks, newWorkflow.CrossClusterTasks, newWorkflow.ReplicationTasks, newWorkflow.TimerTasks,
			nosqlTransferTasks, nosqlCrossClusterTasks, nosqlReplicationTasks, nosqlTimerTasks,
		)
		if err != nil {
			return err
		}
	}

	shardCondition := &nosqlplugin.ShardCondition{
		ShardID: d.shardID,
		RangeID: request.RangeID,
	}

	err = d.db.UpdateWorkflowExecutionWithTasks(
		ctx, currentWorkflowWriteReq,
		mutateExecution, insertExecution, nil, // no workflow to reset here
		nosqlTransferTasks, nosqlCrossClusterTasks, nosqlReplicationTasks, nosqlTimerTasks,
		shardCondition)

	return d.processUpdateWorkflowResult(err, request.RangeID)
}

func (d *cassandraPersistence) ConflictResolveWorkflowExecution(
	ctx context.Context,
	request *p.InternalConflictResolveWorkflowExecutionRequest,
) error {
	currentWorkflow := request.CurrentWorkflowMutation
	resetWorkflow := request.ResetWorkflowSnapshot
	newWorkflow := request.NewWorkflowSnapshot

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

	var currentWorkflowWriteReq *nosqlplugin.CurrentWorkflowWriteRequest
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
		currentWorkflowWriteReq = &nosqlplugin.CurrentWorkflowWriteRequest{
			WriteMode: nosqlplugin.CurrentWorkflowWriteModeNoop,
		}
	case p.ConflictResolveWorkflowModeUpdateCurrent:
		executionInfo := resetWorkflow.ExecutionInfo
		lastWriteVersion := resetWorkflow.LastWriteVersion
		if newWorkflow != nil {
			executionInfo = newWorkflow.ExecutionInfo
			lastWriteVersion = newWorkflow.LastWriteVersion
		}

		if currentWorkflow != nil {
			prevRunID = currentWorkflow.ExecutionInfo.RunID
		} else {
			// reset workflow is current
			prevRunID = resetWorkflow.ExecutionInfo.RunID
		}
		currentWorkflowWriteReq = &nosqlplugin.CurrentWorkflowWriteRequest{
			WriteMode: nosqlplugin.CurrentWorkflowWriteModeUpdate,
			Row: nosqlplugin.CurrentWorkflowRow{
				ShardID:          d.shardID,
				DomainID:         domainID,
				WorkflowID:       workflowID,
				RunID:            executionInfo.RunID,
				State:            executionInfo.State,
				CloseStatus:      executionInfo.CloseStatus,
				CreateRequestID:  executionInfo.CreateRequestID,
				LastWriteVersion: lastWriteVersion,
			},
			Condition: &nosqlplugin.CurrentWorkflowWriteCondition{
				CurrentRunID: &prevRunID,
			},
		}

	default:
		return &types.InternalServiceError{
			Message: fmt.Sprintf("ConflictResolveWorkflowExecution: unknown mode: %v", request.Mode),
		}
	}

	var mutateExecution, insertExecution, resetExecution *nosqlplugin.WorkflowExecutionRequest
	var nosqlTransferTasks []*nosqlplugin.TransferTask
	var nosqlCrossClusterTasks []*nosqlplugin.CrossClusterTask
	var nosqlReplicationTasks []*nosqlplugin.ReplicationTask
	var nosqlTimerTasks []*nosqlplugin.TimerTask
	var err error

	// 1. current
	if currentWorkflow != nil {
		mutateExecution, err = d.prepareUpdateWorkflowExecutionRequestWithMapsAndEventBuffer(currentWorkflow)
		if err != nil {
			return err
		}
		nosqlTransferTasks, nosqlCrossClusterTasks, nosqlReplicationTasks, nosqlTimerTasks, err = d.prepareNoSQLTasksForWorkflowTxn(
			domainID, workflowID, currentWorkflow.ExecutionInfo.RunID,
			currentWorkflow.TransferTasks, currentWorkflow.CrossClusterTasks, currentWorkflow.ReplicationTasks, currentWorkflow.TimerTasks,
			nosqlTransferTasks, nosqlCrossClusterTasks, nosqlReplicationTasks, nosqlTimerTasks,
		)
		if err != nil {
			return err
		}
	}

	// 2. reset
	resetExecution, err = d.prepareResetWorkflowExecutionRequestWithMapsAndEventBuffer(&resetWorkflow)
	if err != nil {
		return err
	}
	nosqlTransferTasks, nosqlCrossClusterTasks, nosqlReplicationTasks, nosqlTimerTasks, err = d.prepareNoSQLTasksForWorkflowTxn(
		domainID, workflowID, resetWorkflow.ExecutionInfo.RunID,
		resetWorkflow.TransferTasks, resetWorkflow.CrossClusterTasks, resetWorkflow.ReplicationTasks, resetWorkflow.TimerTasks,
		nosqlTransferTasks, nosqlCrossClusterTasks, nosqlReplicationTasks, nosqlTimerTasks,
	)
	if err != nil {
		return err
	}

	// 3. new
	if newWorkflow != nil {
		insertExecution, err = d.prepareCreateWorkflowExecutionRequestWithMaps(newWorkflow)

		nosqlTransferTasks, nosqlCrossClusterTasks, nosqlReplicationTasks, nosqlTimerTasks, err = d.prepareNoSQLTasksForWorkflowTxn(
			domainID, workflowID, newWorkflow.ExecutionInfo.RunID,
			newWorkflow.TransferTasks, newWorkflow.CrossClusterTasks, newWorkflow.ReplicationTasks, newWorkflow.TimerTasks,
			nosqlTransferTasks, nosqlCrossClusterTasks, nosqlReplicationTasks, nosqlTimerTasks,
		)
		if err != nil {
			return err
		}
	}

	shardCondition := &nosqlplugin.ShardCondition{
		ShardID: d.shardID,
		RangeID: request.RangeID,
	}

	err = d.db.UpdateWorkflowExecutionWithTasks(
		ctx, currentWorkflowWriteReq,
		mutateExecution, insertExecution, resetExecution,
		nosqlTransferTasks, nosqlCrossClusterTasks, nosqlReplicationTasks, nosqlTimerTasks,
		shardCondition)
	return d.processUpdateWorkflowResult(err, request.RangeID)
}

func (d *cassandraPersistence) processUpdateWorkflowResult(err error, rangeID int64) error {
	if err != nil {
		conditionFailureErr, isConditionFailedError := err.(*nosqlplugin.WorkflowOperationConditionFailure)
		if isConditionFailedError {
			switch {
			case conditionFailureErr.UnknownConditionFailureDetails != nil:
				return &p.ConditionFailedError{
					Msg: *conditionFailureErr.UnknownConditionFailureDetails,
				}
			case conditionFailureErr.ShardRangeIDNotMatch != nil:
				return &p.ShardOwnershipLostError{
					ShardID: d.shardID,
					Msg: fmt.Sprintf("Failed to update workflow execution.  Request RangeID: %v, Actual RangeID: %v",
						rangeID, *conditionFailureErr.ShardRangeIDNotMatch),
				}
			case conditionFailureErr.CurrentWorkflowConditionFailInfo != nil:
				return &p.CurrentWorkflowConditionFailedError{
					Msg: *conditionFailureErr.CurrentWorkflowConditionFailInfo,
				}
			default:
				// If ever runs into this branch, there is bug in the code either in here, or in the implementation of nosql plugin
				err := fmt.Errorf("unexpected conditionFailureReason error")
				d.logger.Error("A code bug exists in persistence layer, please investigate ASAP", tag.Error(err))
				return err
			}
		}
		return convertCommonErrors(d.client, "UpdateWorkflowExecution", err)
	}

	return nil
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
	err := d.db.DeleteWorkflowExecution(ctx, d.shardID, request.DomainID, request.WorkflowID, request.RunID)
	if err != nil {
		return convertCommonErrors(d.client, "DeleteWorkflowExecution", err)
	}

	return nil
}

func (d *cassandraPersistence) DeleteCurrentWorkflowExecution(
	ctx context.Context,
	request *p.DeleteCurrentWorkflowExecutionRequest,
) error {
	err := d.db.DeleteCurrentWorkflow(ctx, d.shardID, request.DomainID, request.WorkflowID, request.RunID)
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
	result, err := d.db.SelectCurrentWorkflow(ctx, d.shardID, request.DomainID, request.WorkflowID)

	if err != nil {
		if d.client.IsNotFoundError(err) {
			return nil, &types.EntityNotExistsError{
				Message: fmt.Sprintf("Workflow execution not found.  WorkflowId: %v",
					request.WorkflowID),
			}
		}
		return nil, convertCommonErrors(d.client, "GetCurrentExecution", err)
	}

	return &p.GetCurrentExecutionResponse{
		RunID:            result.RunID,
		StartRequestID:   result.CreateRequestID,
		State:            result.State,
		CloseStatus:      result.CloseStatus,
		LastWriteVersion: result.LastWriteVersion,
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
			_ = iter.Close()
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
