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

	templateCompleteReplicationTaskQuery = templateCompleteTransferTaskQuery

	templateRangeCompleteReplicationTaskQuery = templateRangeCompleteTransferTaskQuery
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
	executions, token, err := d.db.SelectAllCurrentWorkflows(ctx, d.shardID, request.PageToken, request.PageSize)
	if err != nil {
		return nil, convertCommonErrors(d.client, "ListCurrentExecutions", err)
	}
	return &p.ListCurrentExecutionsResponse{
		Executions: executions,
		PageToken:  token,
	}, nil
}

func (d *cassandraPersistence) IsWorkflowExecutionExists(
	ctx context.Context,
	request *p.IsWorkflowExecutionExistsRequest,
) (*p.IsWorkflowExecutionExistsResponse, error) {
	exists, err := d.db.IsWorkflowExecutionExists(ctx, d.shardID, request.DomainID, request.WorkflowID, request.RunID)
	if err != nil {
		return nil, convertCommonErrors(d.client, "IsWorkflowExecutionExists", err)
	}
	return &p.IsWorkflowExecutionExistsResponse{
		Exists: exists,
	}, nil
}

func (d *cassandraPersistence) ListConcreteExecutions(
	ctx context.Context,
	request *p.ListConcreteExecutionsRequest,
) (*p.InternalListConcreteExecutionsResponse, error) {
	executions, nextPageToken, err := d.db.SelectAllWorkflowExecutions(ctx, d.shardID, request.PageToken, request.PageSize)
	if err != nil {
		return nil, convertCommonErrors(d.client, "ListConcreteExecutions", err)
	}
	return &p.InternalListConcreteExecutionsResponse{
		Executions:    executions,
		NextPageToken: nextPageToken,
	}, nil
}

func (d *cassandraPersistence) GetTransferTasks(
	ctx context.Context,
	request *p.GetTransferTasksRequest,
) (*p.GetTransferTasksResponse, error) {

	tasks, nextPageToken, err := d.db.SelectTransferTasksOrderByTaskID(ctx, d.shardID, request.BatchSize, request.NextPageToken, request.ReadLevel, request.MaxReadLevel)
	if err != nil {
		return nil, convertCommonErrors(d.client, "GetTransferTasks", err)
	}

	return &p.GetTransferTasksResponse{
		Tasks:         tasks,
		NextPageToken: nextPageToken,
	}, nil
}

func (d *cassandraPersistence) GetCrossClusterTasks(
	ctx context.Context,
	request *p.GetCrossClusterTasksRequest,
) (*p.GetCrossClusterTasksResponse, error) {

	cTasks, nextPageToken, err := d.db.SelectCrossClusterTasksOrderByTaskID(ctx, d.shardID, request.BatchSize, request.NextPageToken, request.TargetCluster, request.ReadLevel, request.MaxReadLevel)

	if err != nil {
		return nil, convertCommonErrors(d.client, "GetCrossClusterTasks", err)
	}

	var tTasks []*p.CrossClusterTaskInfo
	for _, t := range cTasks {
		tTasks = append(tTasks, &t.TransferTask)
	}
	return &p.GetCrossClusterTasksResponse{
		Tasks:         tTasks,
		NextPageToken: nextPageToken,
	}, nil
}

func (d *cassandraPersistence) GetReplicationTasks(
	ctx context.Context,
	request *p.GetReplicationTasksRequest,
) (*p.InternalGetReplicationTasksResponse, error) {

	tasks, nextPageToken, err := d.db.SelectReplicationTasksOrderByTaskID(ctx, d.shardID, request.BatchSize, request.NextPageToken, request.ReadLevel, request.MaxReadLevel)
	if err != nil {
		return nil, convertCommonErrors(d.client, "GetReplicationTasks", err)
	}
	return &p.InternalGetReplicationTasksResponse{
		Tasks:         tasks,
		NextPageToken: nextPageToken,
	}, nil
}

func (d *cassandraPersistence) RangeCompleteTransferTask(
	ctx context.Context,
	request *p.RangeCompleteTransferTaskRequest,
) error {
	err := d.db.RangeDeleteTransferTasks(ctx, d.shardID, request.ExclusiveBeginTaskID, request.InclusiveEndTaskID)
	if err != nil {
		return convertCommonErrors(d.client, "RangeCompleteTransferTask", err)
	}

	return nil
}

func (d *cassandraPersistence) CompleteCrossClusterTask(
	ctx context.Context,
	request *p.CompleteCrossClusterTaskRequest,
) error {

	err := d.db.DeleteCrossClusterTask(ctx, d.shardID, request.TargetCluster, request.TaskID)
	if err != nil {
		return convertCommonErrors(d.client, "CompleteCrossClusterTask", err)
	}

	return nil
}

func (d *cassandraPersistence) RangeCompleteCrossClusterTask(
	ctx context.Context,
	request *p.RangeCompleteCrossClusterTaskRequest,
) error {

	err := d.db.RangeDeleteCrossClusterTasks(ctx, d.shardID, request.TargetCluster, request.ExclusiveBeginTaskID, request.InclusiveEndTaskID)
	if err != nil {
		return convertCommonErrors(d.client, "RangeCompleteCrossClusterTask", err)
	}

	return nil
}

func (d *cassandraPersistence) CompleteReplicationTask(
	ctx context.Context,
	request *p.CompleteReplicationTaskRequest,
) error {
	err := d.db.DeleteReplicationTask(ctx, d.shardID, request.TaskID)
	if err != nil {
		return convertCommonErrors(d.client, "CompleteReplicationTask", err)
	}

	return nil
}

func (d *cassandraPersistence) RangeCompleteReplicationTask(
	ctx context.Context,
	request *p.RangeCompleteReplicationTaskRequest,
) error {

	err := d.db.RangeDeleteReplicationTasks(ctx, d.shardID, request.InclusiveEndTaskID)
	if err != nil {
		return convertCommonErrors(d.client, "RangeCompleteReplicationTask", err)
	}

	return nil
}

func (d *cassandraPersistence) CompleteTimerTask(
	ctx context.Context,
	request *p.CompleteTimerTaskRequest,
) error {
	err := d.db.DeleteTimerTask(ctx, d.shardID, request.TaskID, request.VisibilityTimestamp)
	if err != nil {
		return convertCommonErrors(d.client, "CompleteTimerTask", err)
	}

	return nil
}

func (d *cassandraPersistence) RangeCompleteTimerTask(
	ctx context.Context,
	request *p.RangeCompleteTimerTaskRequest,
) error {
	err := d.db.RangeDeleteTimerTasks(ctx, d.shardID, request.InclusiveBeginTimestamp, request.ExclusiveEndTimestamp)
	if err != nil {
		return convertCommonErrors(d.client, "RangeCompleteTimerTask", err)
	}

	return nil
}

func (d *cassandraPersistence) GetTimerIndexTasks(
	ctx context.Context,
	request *p.GetTimerIndexTasksRequest,
) (*p.GetTimerIndexTasksResponse, error) {

	timers, nextPageToken, err := d.db.SelectTimerTasksOrderByVisibilityTime(ctx, d.shardID, request.BatchSize, request.NextPageToken, request.MinTimestamp, request.MaxTimestamp)
	if err != nil {
		return nil, convertCommonErrors(d.client, "GetTimerTasks", err)
	}

	return &p.GetTimerIndexTasksResponse{
		Timers:        timers,
		NextPageToken: nextPageToken,
	}, nil
}

func (d *cassandraPersistence) PutReplicationTaskToDLQ(
	ctx context.Context,
	request *p.InternalPutReplicationTaskToDLQRequest,
) error {

	err := d.db.InsertReplicationDLQTask(ctx, d.shardID, request.SourceClusterName, *request.TaskInfo)
	if err != nil {
		return convertCommonErrors(d.client, "PutReplicationTaskToDLQ", err)
	}

	return nil
}

func (d *cassandraPersistence) GetReplicationTasksFromDLQ(
	ctx context.Context,
	request *p.GetReplicationTasksFromDLQRequest,
) (*p.InternalGetReplicationTasksFromDLQResponse, error) {
	tasks, nextPageToken, err := d.db.SelectReplicationDLQTasksOrderByTaskID(ctx, d.shardID, request.SourceClusterName, request.BatchSize, request.NextPageToken, request.ReadLevel, request.MaxReadLevel)
	if err != nil {
		return nil, convertCommonErrors(d.client, "GetReplicationTasks", err)
	}
	return &p.InternalGetReplicationTasksResponse{
		Tasks:         tasks,
		NextPageToken: nextPageToken,
	}, nil
}

func (d *cassandraPersistence) GetReplicationDLQSize(
	ctx context.Context,
	request *p.GetReplicationDLQSizeRequest,
) (*p.GetReplicationDLQSizeResponse, error) {

	size, err := d.db.SelectReplicationDLQTasksCount(ctx, d.shardID, request.SourceClusterName)
	if err != nil {
		return nil, convertCommonErrors(d.client, "GetReplicationDLQSize", err)
	}
	return &p.GetReplicationDLQSizeResponse{
		Size: size,
	}, nil
}

func (d *cassandraPersistence) DeleteReplicationTaskFromDLQ(
	ctx context.Context,
	request *p.DeleteReplicationTaskFromDLQRequest,
) error {

	err := d.db.DeleteReplicationDLQTask(ctx, d.shardID, request.SourceClusterName, request.TaskID)
	if err != nil {
		return convertCommonErrors(d.client, "DeleteReplicationTaskFromDLQ", err)
	}

	return nil
}

func (d *cassandraPersistence) RangeDeleteReplicationTaskFromDLQ(
	ctx context.Context,
	request *p.RangeDeleteReplicationTaskFromDLQRequest,
) error {

	err := d.db.RangeDeleteReplicationDLQTasks(ctx, d.shardID, request.SourceClusterName, request.ExclusiveBeginTaskID, request.InclusiveEndTaskID)
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