// Copyright (c) 2017-2020 Uber Technologies, Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uber/cadence/common/persistence/serialization"

	"github.com/uber/cadence/common"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/sql/sqlplugin"
	"github.com/uber/cadence/common/types"
)

func applyWorkflowMutationTx(
	ctx context.Context,
	tx sqlplugin.Tx,
	shardID int,
	workflowMutation *p.InternalWorkflowMutation,
	parser serialization.Parser,
) error {

	executionInfo := workflowMutation.ExecutionInfo
	versionHistories := workflowMutation.VersionHistories
	startVersion := workflowMutation.StartVersion
	lastWriteVersion := workflowMutation.LastWriteVersion
	domainID := serialization.MustParseUUID(executionInfo.DomainID)
	workflowID := executionInfo.WorkflowID
	runID := serialization.MustParseUUID(executionInfo.RunID)

	// TODO Remove me if UPDATE holds the lock to the end of a transaction
	if err := lockAndCheckNextEventID(
		ctx,
		tx,
		shardID,
		domainID,
		workflowID,
		runID,
		workflowMutation.Condition); err != nil {
		return err
	}

	if err := updateExecution(
		ctx,
		tx,
		executionInfo,
		versionHistories,
		startVersion,
		lastWriteVersion,
		shardID,
		parser); err != nil {
		return err
	}

	if err := applyTasks(
		ctx,
		tx,
		shardID,
		domainID,
		workflowID,
		runID,
		workflowMutation.TransferTasks,
		workflowMutation.CrossClusterTasks,
		workflowMutation.ReplicationTasks,
		workflowMutation.TimerTasks,
		parser,
	); err != nil {
		return err
	}

	if err := updateActivityInfos(
		ctx,
		tx,
		workflowMutation.UpsertActivityInfos,
		workflowMutation.DeleteActivityInfos,
		shardID,
		domainID,
		workflowID,
		runID,
		parser,
	); err != nil {
		return err
	}

	if err := updateTimerInfos(
		ctx,
		tx,
		workflowMutation.UpsertTimerInfos,
		workflowMutation.DeleteTimerInfos,
		shardID,
		domainID,
		workflowID,
		runID,
		parser,
	); err != nil {
		return err
	}

	if err := updateChildExecutionInfos(
		ctx,
		tx,
		workflowMutation.UpsertChildExecutionInfos,
		workflowMutation.DeleteChildExecutionInfos,
		shardID,
		domainID,
		workflowID,
		runID,
		parser,
	); err != nil {
		return err
	}

	if err := updateRequestCancelInfos(
		ctx,
		tx,
		workflowMutation.UpsertRequestCancelInfos,
		workflowMutation.DeleteRequestCancelInfos,
		shardID,
		domainID,
		workflowID,
		runID,
		parser,
	); err != nil {
		return err
	}

	if err := updateSignalInfos(
		ctx,
		tx,
		workflowMutation.UpsertSignalInfos,
		workflowMutation.DeleteSignalInfos,
		shardID,
		domainID,
		workflowID,
		runID,
		parser,
	); err != nil {
		return err
	}

	if err := updateSignalsRequested(
		ctx,
		tx,
		workflowMutation.UpsertSignalRequestedIDs,
		workflowMutation.DeleteSignalRequestedIDs,
		shardID,
		domainID,
		workflowID,
		runID,
	); err != nil {
		return err
	}

	if workflowMutation.ClearBufferedEvents {
		if err := deleteBufferedEvents(
			ctx,
			tx,
			shardID,
			domainID,
			workflowID,
			runID,
		); err != nil {
			return err
		}
	}

	return updateBufferedEvents(
		ctx,
		tx,
		workflowMutation.NewBufferedEvents,
		shardID,
		domainID,
		workflowID,
		runID,
	)
}

func applyWorkflowSnapshotTxAsReset(
	ctx context.Context,
	tx sqlplugin.Tx,
	shardID int,
	workflowSnapshot *p.InternalWorkflowSnapshot,
	parser serialization.Parser,
) error {

	executionInfo := workflowSnapshot.ExecutionInfo
	versionHistories := workflowSnapshot.VersionHistories
	startVersion := workflowSnapshot.StartVersion
	lastWriteVersion := workflowSnapshot.LastWriteVersion
	domainID := serialization.MustParseUUID(executionInfo.DomainID)
	workflowID := executionInfo.WorkflowID
	runID := serialization.MustParseUUID(executionInfo.RunID)

	// TODO Is there a way to modify the various map tables without fear of other people adding rows after we delete, without locking the executions row?
	if err := lockAndCheckNextEventID(
		ctx,
		tx,
		shardID,
		domainID,
		workflowID,
		runID,
		workflowSnapshot.Condition); err != nil {
		return err
	}

	if err := updateExecution(
		ctx,
		tx,
		executionInfo,
		versionHistories,
		startVersion,
		lastWriteVersion,
		shardID,
		parser); err != nil {
		return err
	}

	if err := applyTasks(
		ctx,
		tx,
		shardID,
		domainID,
		workflowID,
		runID,
		workflowSnapshot.TransferTasks,
		workflowSnapshot.CrossClusterTasks,
		workflowSnapshot.ReplicationTasks,
		workflowSnapshot.TimerTasks,
		parser,
	); err != nil {
		return err
	}

	if err := deleteActivityInfoMap(
		ctx,
		tx,
		shardID,
		domainID,
		workflowID,
		runID); err != nil {
		return err
	}

	if err := updateActivityInfos(
		ctx,
		tx,
		workflowSnapshot.ActivityInfos,
		nil,
		shardID,
		domainID,
		workflowID,
		runID,
		parser); err != nil {
		return err
	}

	if err := deleteTimerInfoMap(
		ctx,
		tx,
		shardID,
		domainID,
		workflowID,
		runID); err != nil {
		return err
	}

	if err := updateTimerInfos(
		ctx,
		tx,
		workflowSnapshot.TimerInfos,
		nil,
		shardID,
		domainID,
		workflowID,
		runID,
		parser); err != nil {
		return err
	}

	if err := deleteChildExecutionInfoMap(
		ctx,
		tx,
		shardID,
		domainID,
		workflowID,
		runID); err != nil {
		return err
	}

	if err := updateChildExecutionInfos(
		ctx,
		tx,
		workflowSnapshot.ChildExecutionInfos,
		nil,
		shardID,
		domainID,
		workflowID,
		runID,
		parser); err != nil {
		return err
	}

	if err := deleteRequestCancelInfoMap(
		ctx,
		tx,
		shardID,
		domainID,
		workflowID,
		runID); err != nil {
		return err
	}

	if err := updateRequestCancelInfos(
		ctx,
		tx,
		workflowSnapshot.RequestCancelInfos,
		nil,
		shardID,
		domainID,
		workflowID,
		runID,
		parser); err != nil {
		return err
	}

	if err := deleteSignalInfoMap(
		ctx,
		tx,
		shardID,
		domainID,
		workflowID,
		runID); err != nil {
		return err
	}

	if err := updateSignalInfos(
		ctx,
		tx,
		workflowSnapshot.SignalInfos,
		nil,
		shardID,
		domainID,
		workflowID,
		runID,
		parser); err != nil {
		return err
	}

	if err := deleteSignalsRequestedSet(
		ctx,
		tx,
		shardID,
		domainID,
		workflowID,
		runID); err != nil {
		return err
	}

	if err := updateSignalsRequested(
		ctx,
		tx,
		workflowSnapshot.SignalRequestedIDs,
		nil,
		shardID,
		domainID,
		workflowID,
		runID); err != nil {
		return err
	}

	return deleteBufferedEvents(
		ctx,
		tx,
		shardID,
		domainID,
		workflowID,
		runID)
}

func (m *sqlExecutionStore) applyWorkflowSnapshotTxAsNew(
	ctx context.Context,
	tx sqlplugin.Tx,
	shardID int,
	workflowSnapshot *p.InternalWorkflowSnapshot,
	parser serialization.Parser,
) error {

	executionInfo := workflowSnapshot.ExecutionInfo
	versionHistories := workflowSnapshot.VersionHistories
	startVersion := workflowSnapshot.StartVersion
	lastWriteVersion := workflowSnapshot.LastWriteVersion
	domainID := serialization.MustParseUUID(executionInfo.DomainID)
	workflowID := executionInfo.WorkflowID
	runID := serialization.MustParseUUID(executionInfo.RunID)

	if err := m.createExecution(
		ctx,
		tx,
		executionInfo,
		versionHistories,
		startVersion,
		lastWriteVersion,
		shardID,
		parser); err != nil {
		return err
	}

	if err := applyTasks(
		ctx,
		tx,
		shardID,
		domainID,
		workflowID,
		runID,
		workflowSnapshot.TransferTasks,
		workflowSnapshot.CrossClusterTasks,
		workflowSnapshot.ReplicationTasks,
		workflowSnapshot.TimerTasks,
		parser,
	); err != nil {
		return err
	}

	if err := updateActivityInfos(
		ctx,
		tx,
		workflowSnapshot.ActivityInfos,
		nil,
		shardID,
		domainID,
		workflowID,
		runID,
		parser); err != nil {
		return err
	}

	if err := updateTimerInfos(
		ctx,
		tx,
		workflowSnapshot.TimerInfos,
		nil,
		shardID,
		domainID,
		workflowID,
		runID,
		parser); err != nil {
		return err
	}

	if err := updateChildExecutionInfos(
		ctx,
		tx,
		workflowSnapshot.ChildExecutionInfos,
		nil,
		shardID,
		domainID,
		workflowID,
		runID,
		parser); err != nil {
		return err
	}

	if err := updateRequestCancelInfos(
		ctx,
		tx,
		workflowSnapshot.RequestCancelInfos,
		nil,
		shardID,
		domainID,
		workflowID,
		runID,
		parser); err != nil {
		return err
	}

	if err := updateSignalInfos(
		ctx,
		tx,
		workflowSnapshot.SignalInfos,
		nil,
		shardID,
		domainID,
		workflowID,
		runID,
		parser); err != nil {
		return err
	}

	return updateSignalsRequested(
		ctx,
		tx,
		workflowSnapshot.SignalRequestedIDs,
		nil,
		shardID,
		domainID,
		workflowID,
		runID)
}

func applyTasks(
	ctx context.Context,
	tx sqlplugin.Tx,
	shardID int,
	domainID serialization.UUID,
	workflowID string,
	runID serialization.UUID,
	transferTasks []p.Task,
	crossClusterTasks []p.Task,
	replicationTasks []p.Task,
	timerTasks []p.Task,
	parser serialization.Parser,
) error {

	if err := createTransferTasks(
		ctx,
		tx,
		transferTasks,
		shardID,
		domainID,
		workflowID,
		runID,
		parser); err != nil {
		return err
	}

	if err := createCrossClusterTasks(
		ctx,
		tx,
		crossClusterTasks,
		shardID,
		domainID,
		workflowID,
		runID,
		parser); err != nil {
		return err
	}

	if err := createReplicationTasks(
		ctx,
		tx,
		replicationTasks,
		shardID,
		domainID,
		workflowID,
		runID,
		parser,
	); err != nil {
		return err
	}

	return createTimerTasks(
		ctx,
		tx,
		timerTasks,
		shardID,
		domainID,
		workflowID,
		runID,
		parser,
	)
}

// lockCurrentExecutionIfExists returns current execution or nil if none is found for the workflowID
// locking it in the DB
func lockCurrentExecutionIfExists(
	ctx context.Context,
	tx sqlplugin.Tx,
	shardID int,
	domainID serialization.UUID,
	workflowID string,
) (*sqlplugin.CurrentExecutionsRow, error) {

	rows, err := tx.LockCurrentExecutionsJoinExecutions(ctx, &sqlplugin.CurrentExecutionsFilter{
		ShardID:    int64(shardID),
		DomainID:   domainID,
		WorkflowID: workflowID,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, convertCommonErrors(tx, "lockCurrentExecutionIfExists", fmt.Sprintf("Failed to get current_executions row for (shard,domain,workflow) = (%v, %v, %v).", shardID, domainID, workflowID), err)
		}
	}
	size := len(rows)
	if size > 1 {
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("lockCurrentExecutionIfExists failed. Multiple current_executions rows for (shard,domain,workflow) = (%v, %v, %v).", shardID, domainID, workflowID),
		}
	}
	if size == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func createOrUpdateCurrentExecution(
	ctx context.Context,
	tx sqlplugin.Tx,
	createMode p.CreateWorkflowMode,
	shardID int,
	domainID serialization.UUID,
	workflowID string,
	runID serialization.UUID,
	state int,
	closeStatus int,
	createRequestID string,
	startVersion int64,
	lastWriteVersion int64,
) error {

	row := sqlplugin.CurrentExecutionsRow{
		ShardID:          int64(shardID),
		DomainID:         domainID,
		WorkflowID:       workflowID,
		RunID:            runID,
		CreateRequestID:  createRequestID,
		State:            state,
		CloseStatus:      closeStatus,
		StartVersion:     startVersion,
		LastWriteVersion: lastWriteVersion,
	}

	switch createMode {
	case p.CreateWorkflowModeContinueAsNew:
		if err := updateCurrentExecution(
			ctx,
			tx,
			shardID,
			domainID,
			workflowID,
			runID,
			createRequestID,
			state,
			closeStatus,
			row.StartVersion,
			row.LastWriteVersion); err != nil {
			return err
		}
	case p.CreateWorkflowModeWorkflowIDReuse:
		if err := updateCurrentExecution(
			ctx,
			tx,
			shardID,
			domainID,
			workflowID,
			runID,
			createRequestID,
			state,
			closeStatus,
			row.StartVersion,
			row.LastWriteVersion); err != nil {
			return err
		}
	case p.CreateWorkflowModeBrandNew:
		if _, err := tx.InsertIntoCurrentExecutions(ctx, &row); err != nil {
			return convertCommonErrors(tx, "createOrUpdateCurrentExecution", "Failed to insert into current_executions table.", err)
		}
	case p.CreateWorkflowModeZombie:
		// noop
	default:
		return fmt.Errorf("createOrUpdateCurrentExecution failed. Unknown workflow creation mode: %v", createMode)
	}

	return nil
}

func lockAndCheckNextEventID(
	ctx context.Context,
	tx sqlplugin.Tx,
	shardID int,
	domainID serialization.UUID,
	workflowID string,
	runID serialization.UUID,
	condition int64,
) error {

	nextEventID, err := lockNextEventID(
		ctx,
		tx,
		shardID,
		domainID,
		workflowID,
		runID,
	)

	if err != nil {
		return err
	}
	if *nextEventID != condition {
		return &p.ConditionFailedError{
			Msg: fmt.Sprintf("lockAndCheckNextEventID failed. Next_event_id was %v when it should have been %v.", nextEventID, condition),
		}
	}
	return nil
}

func lockNextEventID(
	ctx context.Context,
	tx sqlplugin.Tx,
	shardID int,
	domainID serialization.UUID,
	workflowID string,
	runID serialization.UUID,
) (*int64, error) {

	nextEventID, err := tx.WriteLockExecutions(ctx, &sqlplugin.ExecutionsFilter{
		ShardID:    shardID,
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &types.EntityNotExistsError{
				Message: fmt.Sprintf(
					"lockNextEventID failed. Unable to lock executions row with (shard, domain, workflow, run) = (%v,%v,%v,%v) which does not exist.",
					shardID,
					domainID,
					workflowID,
					runID,
				),
			}
		}
		return nil, convertCommonErrors(tx, "lockNextEventID", "", err)
	}
	result := int64(nextEventID)
	return &result, nil
}

func createCrossClusterTasks(
	ctx context.Context,
	tx sqlplugin.Tx,
	crossClusterTasks []p.Task,
	shardID int,
	domainID serialization.UUID,
	workflowID string,
	runID serialization.UUID,
	parser serialization.Parser,
) error {
	if len(crossClusterTasks) == 0 {
		return nil
	}

	crossClusterTasksRows := make([]sqlplugin.CrossClusterTasksRow, len(crossClusterTasks))
	for i, task := range crossClusterTasks {
		info := &serialization.CrossClusterTaskInfo{
			DomainID:            domainID,
			WorkflowID:          workflowID,
			RunID:               runID,
			TaskType:            int16(task.GetType()),
			TargetDomainID:      domainID,
			TargetWorkflowID:    p.TransferTaskTransferTargetWorkflowID,
			TargetRunID:         serialization.UUID(p.CrossClusterTaskDefaultTargetRunID),
			ScheduleID:          0,
			Version:             task.GetVersion(),
			VisibilityTimestamp: task.GetVisibilityTimestamp(),
		}

		crossClusterTasksRows[i].ShardID = shardID
		crossClusterTasksRows[i].TaskID = task.GetTaskID()

		switch task.GetType() {
		case p.CrossClusterTaskTypeStartChildExecution:
			crossClusterTasksRows[i].TargetCluster = task.(*p.CrossClusterStartChildExecutionTask).TargetCluster
			info.TargetDomainID = serialization.MustParseUUID(task.(*p.CrossClusterStartChildExecutionTask).TargetDomainID)
			info.TargetWorkflowID = task.(*p.CrossClusterStartChildExecutionTask).TargetWorkflowID
			info.ScheduleID = task.(*p.CrossClusterStartChildExecutionTask).InitiatedID

		case p.CrossClusterTaskTypeCancelExecution:
			crossClusterTasksRows[i].TargetCluster = task.(*p.CrossClusterCancelExecutionTask).TargetCluster
			info.TargetDomainID = serialization.MustParseUUID(task.(*p.CrossClusterCancelExecutionTask).TargetDomainID)
			info.TargetWorkflowID = task.(*p.CrossClusterCancelExecutionTask).TargetWorkflowID
			info.TargetRunID = serialization.MustParseUUID(task.(*p.CrossClusterCancelExecutionTask).TargetRunID)
			info.TargetChildWorkflowOnly = task.(*p.CrossClusterCancelExecutionTask).TargetChildWorkflowOnly
			info.ScheduleID = task.(*p.CrossClusterCancelExecutionTask).InitiatedID

		case p.CrossClusterTaskTypeSignalExecution:
			crossClusterTasksRows[i].TargetCluster = task.(*p.CrossClusterSignalExecutionTask).TargetCluster
			info.TargetDomainID = serialization.MustParseUUID(task.(*p.CrossClusterSignalExecutionTask).TargetDomainID)
			info.TargetWorkflowID = task.(*p.CrossClusterSignalExecutionTask).TargetWorkflowID
			info.TargetRunID = serialization.MustParseUUID(task.(*p.CrossClusterSignalExecutionTask).TargetRunID)
			info.TargetChildWorkflowOnly = task.(*p.CrossClusterSignalExecutionTask).TargetChildWorkflowOnly
			info.ScheduleID = task.(*p.CrossClusterSignalExecutionTask).InitiatedID

		case p.CrossClusterTaskTypeRecordChildWorkflowExeuctionComplete:
			crossClusterTasksRows[i].TargetCluster = task.(*p.CrossClusterRecordChildWorkflowExecutionCompleteTask).TargetCluster

		case p.CrossClusterTaskTypeApplyParentPolicy:
			crossClusterTasksRows[i].TargetCluster = task.(*p.CrossClusterApplyParentClosePolicyTask).TargetCluster

		default:
			return &types.InternalServiceError{
				Message: fmt.Sprintf("Unknown cross-cluster task type: %v", task.GetType()),
			}
		}

		blob, err := parser.CrossClusterTaskInfoToBlob(info)
		if err != nil {
			return err
		}
		crossClusterTasksRows[i].Data = blob.Data
		crossClusterTasksRows[i].DataEncoding = string(blob.Encoding)
	}

	result, err := tx.InsertIntoCrossClusterTasks(ctx, crossClusterTasksRows)
	if err != nil {
		return convertCommonErrors(tx, "createCrossClusterTasks", "", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("createTransferTasks failed. Could not verify number of rows inserted. Error: %v", err),
		}
	}

	if int(rowsAffected) != len(crossClusterTasks) {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("createCrossClusterTasks failed. Inserted %v instead of %v rows into transfer_tasks. Error: %v", rowsAffected, len(crossClusterTasks), err),
		}
	}

	return nil
}

func createTransferTasks(
	ctx context.Context,
	tx sqlplugin.Tx,
	transferTasks []p.Task,
	shardID int,
	domainID serialization.UUID,
	workflowID string,
	runID serialization.UUID,
	parser serialization.Parser,
) error {

	if len(transferTasks) == 0 {
		return nil
	}

	transferTasksRows := make([]sqlplugin.TransferTasksRow, len(transferTasks))
	for i, task := range transferTasks {
		info := &serialization.TransferTaskInfo{
			DomainID:            domainID,
			WorkflowID:          workflowID,
			RunID:               runID,
			TaskType:            int16(task.GetType()),
			TargetDomainID:      domainID,
			TargetWorkflowID:    p.TransferTaskTransferTargetWorkflowID,
			ScheduleID:          0,
			Version:             task.GetVersion(),
			VisibilityTimestamp: task.GetVisibilityTimestamp(),
		}

		transferTasksRows[i].ShardID = shardID
		transferTasksRows[i].TaskID = task.GetTaskID()

		switch task.GetType() {
		case p.TransferTaskTypeActivityTask:
			info.TargetDomainID = serialization.MustParseUUID(task.(*p.ActivityTask).DomainID)
			info.TaskList = task.(*p.ActivityTask).TaskList
			info.ScheduleID = task.(*p.ActivityTask).ScheduleID

		case p.TransferTaskTypeDecisionTask:
			info.TargetDomainID = serialization.MustParseUUID(task.(*p.DecisionTask).DomainID)
			info.TaskList = task.(*p.DecisionTask).TaskList
			info.ScheduleID = task.(*p.DecisionTask).ScheduleID

		case p.TransferTaskTypeCancelExecution:
			info.TargetDomainID = serialization.MustParseUUID(task.(*p.CancelExecutionTask).TargetDomainID)
			info.TargetWorkflowID = task.(*p.CancelExecutionTask).TargetWorkflowID
			if task.(*p.CancelExecutionTask).TargetRunID != "" {
				info.TargetRunID = serialization.MustParseUUID(task.(*p.CancelExecutionTask).TargetRunID)
			}
			info.TargetChildWorkflowOnly = task.(*p.CancelExecutionTask).TargetChildWorkflowOnly
			info.ScheduleID = task.(*p.CancelExecutionTask).InitiatedID

		case p.TransferTaskTypeSignalExecution:
			info.TargetDomainID = serialization.MustParseUUID(task.(*p.SignalExecutionTask).TargetDomainID)
			info.TargetWorkflowID = task.(*p.SignalExecutionTask).TargetWorkflowID
			if task.(*p.SignalExecutionTask).TargetRunID != "" {
				info.TargetRunID = serialization.MustParseUUID(task.(*p.SignalExecutionTask).TargetRunID)
			}
			info.TargetChildWorkflowOnly = task.(*p.SignalExecutionTask).TargetChildWorkflowOnly
			info.ScheduleID = task.(*p.SignalExecutionTask).InitiatedID

		case p.TransferTaskTypeStartChildExecution:
			info.TargetDomainID = serialization.MustParseUUID(task.(*p.StartChildExecutionTask).TargetDomainID)
			info.TargetWorkflowID = task.(*p.StartChildExecutionTask).TargetWorkflowID
			info.ScheduleID = task.(*p.StartChildExecutionTask).InitiatedID

		case p.TransferTaskTypeCloseExecution,
			p.TransferTaskTypeRecordWorkflowStarted,
			p.TransferTaskTypeResetWorkflow,
			p.TransferTaskTypeUpsertWorkflowSearchAttributes:
			// No explicit property needs to be set

		default:
			return &types.InternalServiceError{
				Message: fmt.Sprintf("createTransferTasks failed. Unknow transfer type: %v", task.GetType()),
			}
		}

		blob, err := parser.TransferTaskInfoToBlob(info)
		if err != nil {
			return err
		}
		transferTasksRows[i].Data = blob.Data
		transferTasksRows[i].DataEncoding = string(blob.Encoding)
	}

	result, err := tx.InsertIntoTransferTasks(ctx, transferTasksRows)
	if err != nil {
		return convertCommonErrors(tx, "createTransferTasks", "", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("createTransferTasks failed. Could not verify number of rows inserted. Error: %v", err),
		}
	}

	if int(rowsAffected) != len(transferTasks) {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("createTransferTasks failed. Inserted %v instead of %v rows into transfer_tasks. Error: %v", rowsAffected, len(transferTasks), err),
		}
	}

	return nil
}

func createReplicationTasks(
	ctx context.Context,
	tx sqlplugin.Tx,
	replicationTasks []p.Task,
	shardID int,
	domainID serialization.UUID,
	workflowID string,
	runID serialization.UUID,
	parser serialization.Parser,
) error {

	if len(replicationTasks) == 0 {
		return nil
	}
	replicationTasksRows := make([]sqlplugin.ReplicationTasksRow, len(replicationTasks))

	for i, task := range replicationTasks {

		firstEventID := common.EmptyEventID
		nextEventID := common.EmptyEventID
		version := common.EmptyVersion
		activityScheduleID := common.EmptyEventID
		var branchToken, newRunBranchToken []byte

		switch task.GetType() {
		case p.ReplicationTaskTypeHistory:
			historyReplicationTask, ok := task.(*p.HistoryReplicationTask)
			if !ok {
				return &types.InternalServiceError{
					Message: fmt.Sprintf("createReplicationTasks failed. Failed to cast %v to HistoryReplicationTask", task),
				}
			}
			firstEventID = historyReplicationTask.FirstEventID
			nextEventID = historyReplicationTask.NextEventID
			version = task.GetVersion()
			branchToken = historyReplicationTask.BranchToken
			newRunBranchToken = historyReplicationTask.NewRunBranchToken

		case p.ReplicationTaskTypeSyncActivity:
			version = task.GetVersion()
			activityScheduleID = task.(*p.SyncActivityTask).ScheduledID

		case p.ReplicationTaskTypeFailoverMarker:
			version = task.GetVersion()

		default:
			return &types.InternalServiceError{
				Message: fmt.Sprintf("Unknown replication task: %v", task.GetType()),
			}
		}

		blob, err := parser.ReplicationTaskInfoToBlob(&serialization.ReplicationTaskInfo{
			DomainID:                domainID,
			WorkflowID:              workflowID,
			RunID:                   runID,
			TaskType:                int16(task.GetType()),
			FirstEventID:            firstEventID,
			NextEventID:             nextEventID,
			Version:                 version,
			ScheduledID:             activityScheduleID,
			EventStoreVersion:       p.EventStoreVersion,
			NewRunEventStoreVersion: p.EventStoreVersion,
			BranchToken:             branchToken,
			NewRunBranchToken:       newRunBranchToken,
			CreationTimestamp:       task.GetVisibilityTimestamp(),
		})
		if err != nil {
			return err
		}
		replicationTasksRows[i].ShardID = shardID
		replicationTasksRows[i].TaskID = task.GetTaskID()
		replicationTasksRows[i].Data = blob.Data
		replicationTasksRows[i].DataEncoding = string(blob.Encoding)
	}

	result, err := tx.InsertIntoReplicationTasks(ctx, replicationTasksRows)
	if err != nil {
		return convertCommonErrors(tx, "createReplicationTasks", "", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("createReplicationTasks failed. Could not verify number of rows inserted. Error: %v", err),
		}
	}

	if int(rowsAffected) != len(replicationTasks) {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("createReplicationTasks failed. Inserted %v instead of %v rows into transfer_tasks. Error: %v", rowsAffected, len(replicationTasks), err),
		}
	}

	return nil
}

func createTimerTasks(
	ctx context.Context,
	tx sqlplugin.Tx,
	timerTasks []p.Task,
	shardID int,
	domainID serialization.UUID,
	workflowID string,
	runID serialization.UUID,
	parser serialization.Parser,
) error {

	if len(timerTasks) == 0 {
		return nil
	}

	timerTasksRows := make([]sqlplugin.TimerTasksRow, len(timerTasks))

	for i, task := range timerTasks {
		info := &serialization.TimerTaskInfo{
			DomainID:        domainID,
			WorkflowID:      workflowID,
			RunID:           runID,
			TaskType:        int16(task.GetType()),
			Version:         task.GetVersion(),
			EventID:         common.EmptyEventID,
			ScheduleAttempt: 0,
		}

		switch t := task.(type) {
		case *p.DecisionTimeoutTask:
			info.EventID = t.EventID
			info.TimeoutType = common.Int16Ptr(int16(t.TimeoutType))
			info.ScheduleAttempt = t.ScheduleAttempt

		case *p.ActivityTimeoutTask:
			info.EventID = t.EventID
			info.TimeoutType = common.Int16Ptr(int16(t.TimeoutType))
			info.ScheduleAttempt = t.Attempt

		case *p.UserTimerTask:
			info.EventID = t.EventID

		case *p.ActivityRetryTimerTask:
			info.EventID = t.EventID
			info.ScheduleAttempt = int64(t.Attempt)

		case *p.WorkflowBackoffTimerTask:
			info.EventID = t.EventID
			info.TimeoutType = common.Int16Ptr(int16(t.TimeoutType))

		case *p.WorkflowTimeoutTask:
			// noop

		case *p.DeleteHistoryEventTask:
			// noop

		default:
			return &types.InternalServiceError{
				Message: fmt.Sprintf("createTimerTasks failed. Unknown timer task: %v", task.GetType()),
			}
		}

		blob, err := parser.TimerTaskInfoToBlob(info)
		if err != nil {
			return err
		}

		timerTasksRows[i].ShardID = shardID
		timerTasksRows[i].VisibilityTimestamp = task.GetVisibilityTimestamp()
		timerTasksRows[i].TaskID = task.GetTaskID()
		timerTasksRows[i].Data = blob.Data
		timerTasksRows[i].DataEncoding = string(blob.Encoding)
	}

	result, err := tx.InsertIntoTimerTasks(ctx, timerTasksRows)
	if err != nil {
		return convertCommonErrors(tx, "createTimerTasks", "", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("createTimerTasks failed. Could not verify number of rows inserted. Error: %v", err),
		}
	}

	if int(rowsAffected) != len(timerTasks) {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("createTimerTasks failed. Inserted %v instead of %v rows into timer_tasks.", rowsAffected, len(timerTasks)),
		}
	}

	return nil
}

func assertNotCurrentExecution(
	ctx context.Context,
	tx sqlplugin.Tx,
	shardID int,
	domainID serialization.UUID,
	workflowID string,
	runID serialization.UUID,
) error {
	currentRow, err := tx.LockCurrentExecutions(ctx, &sqlplugin.CurrentExecutionsFilter{
		ShardID:    int64(shardID),
		DomainID:   domainID,
		WorkflowID: workflowID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			// allow bypassing no current record
			return nil
		}
		return convertCommonErrors(tx, "assertCurrentExecution", "Unable to load current record.", err)
	}
	return assertRunIDMismatch(runID, currentRow.RunID)
}

func assertRunIDAndUpdateCurrentExecution(
	ctx context.Context,
	tx sqlplugin.Tx,
	shardID int,
	domainID serialization.UUID,
	workflowID string,
	newRunID serialization.UUID,
	previousRunID serialization.UUID,
	createRequestID string,
	state int,
	closeStatus int,
	startVersion int64,
	lastWriteVersion int64,
) error {

	assertFn := func(currentRow *sqlplugin.CurrentExecutionsRow) error {
		if !bytes.Equal(currentRow.RunID, previousRunID) {
			return &p.ConditionFailedError{Msg: fmt.Sprintf(
				"assertRunIDAndUpdateCurrentExecution failed. Current run ID was %v, expected %v",
				currentRow.RunID,
				previousRunID,
			)}
		}
		return nil
	}
	if err := assertCurrentExecution(ctx, tx, shardID, domainID, workflowID, assertFn); err != nil {
		return err
	}

	return updateCurrentExecution(ctx, tx, shardID, domainID, workflowID, newRunID, createRequestID, state, closeStatus, startVersion, lastWriteVersion)
}

func assertAndUpdateCurrentExecution(
	ctx context.Context,
	tx sqlplugin.Tx,
	shardID int,
	domainID serialization.UUID,
	workflowID string,
	newRunID serialization.UUID,
	previousRunID serialization.UUID,
	previousLastWriteVersion int64,
	previousState int,
	createRequestID string,
	state int,
	closeStatus int,
	startVersion int64,
	lastWriteVersion int64,
) error {

	assertFn := func(currentRow *sqlplugin.CurrentExecutionsRow) error {
		if !bytes.Equal(currentRow.RunID, previousRunID) {
			return &p.ConditionFailedError{Msg: fmt.Sprintf(
				"assertAndUpdateCurrentExecution failed. Current run ID was %v, expected %v",
				currentRow.RunID,
				previousRunID,
			)}
		}
		if currentRow.LastWriteVersion != previousLastWriteVersion {
			return &p.ConditionFailedError{Msg: fmt.Sprintf(
				"assertAndUpdateCurrentExecution failed. Current last write version was %v, expected %v",
				currentRow.LastWriteVersion,
				previousLastWriteVersion,
			)}
		}
		if currentRow.State != previousState {
			return &p.ConditionFailedError{Msg: fmt.Sprintf(
				"assertAndUpdateCurrentExecution failed. Current state %v, expected %v",
				currentRow.State,
				previousState,
			)}
		}
		return nil
	}
	if err := assertCurrentExecution(ctx, tx, shardID, domainID, workflowID, assertFn); err != nil {
		return err
	}

	return updateCurrentExecution(ctx, tx, shardID, domainID, workflowID, newRunID, createRequestID, state, closeStatus, startVersion, lastWriteVersion)
}

func assertCurrentExecution(
	ctx context.Context,
	tx sqlplugin.Tx,
	shardID int,
	domainID serialization.UUID,
	workflowID string,
	assertFn func(currentRow *sqlplugin.CurrentExecutionsRow) error,
) error {

	currentRow, err := tx.LockCurrentExecutions(ctx, &sqlplugin.CurrentExecutionsFilter{
		ShardID:    int64(shardID),
		DomainID:   domainID,
		WorkflowID: workflowID,
	})
	if err != nil {
		return convertCommonErrors(tx, "assertCurrentExecution", "Unable to load current record.", err)
	}
	return assertFn(currentRow)
}

func assertRunIDMismatch(runID serialization.UUID, currentRunID serialization.UUID) error {
	// zombie workflow creation with existence of current record, this is a noop
	if bytes.Equal(currentRunID, runID) {
		return &p.ConditionFailedError{Msg: fmt.Sprintf(
			"assertRunIDMismatch failed. Current run ID was %v, input %v",
			currentRunID,
			runID,
		)}
	}
	return nil
}

func updateCurrentExecution(
	ctx context.Context,
	tx sqlplugin.Tx,
	shardID int,
	domainID serialization.UUID,
	workflowID string,
	runID serialization.UUID,
	createRequestID string,
	state int,
	closeStatus int,
	startVersion int64,
	lastWriteVersion int64,
) error {

	result, err := tx.UpdateCurrentExecutions(ctx, &sqlplugin.CurrentExecutionsRow{
		ShardID:          int64(shardID),
		DomainID:         domainID,
		WorkflowID:       workflowID,
		RunID:            runID,
		CreateRequestID:  createRequestID,
		State:            state,
		CloseStatus:      closeStatus,
		StartVersion:     startVersion,
		LastWriteVersion: lastWriteVersion,
	})
	if err != nil {
		return convertCommonErrors(tx, "updateCurrentExecution", "", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("updateCurrentExecution failed. Failed to check number of rows updated in current_executions table. Error: %v", err),
		}
	}
	if rowsAffected != 1 {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("updateCurrentExecution failed. %v rows of current_executions updated instead of 1.", rowsAffected),
		}
	}
	return nil
}

func buildExecutionRow(
	executionInfo *p.InternalWorkflowExecutionInfo,
	versionHistories *p.DataBlob,
	startVersion int64,
	lastWriteVersion int64,
	shardID int,
	parser serialization.Parser,
) (row *sqlplugin.ExecutionsRow, err error) {

	info := &serialization.WorkflowExecutionInfo{
		TaskList:                           executionInfo.TaskList,
		WorkflowTypeName:                   executionInfo.WorkflowTypeName,
		WorkflowTimeout:                    executionInfo.WorkflowTimeout,
		DecisionTaskTimeout:                executionInfo.DecisionStartToCloseTimeout,
		ExecutionContext:                   executionInfo.ExecutionContext,
		State:                              int32(executionInfo.State),
		CloseStatus:                        int32(executionInfo.CloseStatus),
		LastFirstEventID:                   executionInfo.LastFirstEventID,
		LastEventTaskID:                    executionInfo.LastEventTaskID,
		LastProcessedEvent:                 executionInfo.LastProcessedEvent,
		StartTimestamp:                     executionInfo.StartTimestamp,
		LastUpdatedTimestamp:               executionInfo.LastUpdatedTimestamp,
		CreateRequestID:                    executionInfo.CreateRequestID,
		DecisionVersion:                    executionInfo.DecisionVersion,
		DecisionScheduleID:                 executionInfo.DecisionScheduleID,
		DecisionStartedID:                  executionInfo.DecisionStartedID,
		DecisionRequestID:                  executionInfo.DecisionRequestID,
		DecisionTimeout:                    executionInfo.DecisionTimeout,
		DecisionAttempt:                    executionInfo.DecisionAttempt,
		DecisionStartedTimestamp:           executionInfo.DecisionStartedTimestamp,
		DecisionScheduledTimestamp:         executionInfo.DecisionScheduledTimestamp,
		DecisionOriginalScheduledTimestamp: executionInfo.DecisionOriginalScheduledTimestamp,
		StickyTaskList:                     executionInfo.StickyTaskList,
		StickyScheduleToStartTimeout:       executionInfo.StickyScheduleToStartTimeout,
		ClientLibraryVersion:               executionInfo.ClientLibraryVersion,
		ClientFeatureVersion:               executionInfo.ClientFeatureVersion,
		ClientImpl:                         executionInfo.ClientImpl,
		SignalCount:                        int64(executionInfo.SignalCount),
		HistorySize:                        executionInfo.HistorySize,
		CronSchedule:                       executionInfo.CronSchedule,
		CompletionEventBatchID:             &executionInfo.CompletionEventBatchID,
		HasRetryPolicy:                     executionInfo.HasRetryPolicy,
		RetryAttempt:                       int64(executionInfo.Attempt),
		RetryInitialInterval:               executionInfo.InitialInterval,
		RetryBackoffCoefficient:            executionInfo.BackoffCoefficient,
		RetryMaximumInterval:               executionInfo.MaximumInterval,
		RetryMaximumAttempts:               executionInfo.MaximumAttempts,
		RetryExpiration:                    executionInfo.ExpirationSeconds,
		RetryExpirationTimestamp:           executionInfo.ExpirationTime,
		RetryNonRetryableErrors:            executionInfo.NonRetriableErrors,
		EventStoreVersion:                  p.EventStoreVersion,
		EventBranchToken:                   executionInfo.BranchToken,
		AutoResetPoints:                    executionInfo.AutoResetPoints.Data,
		AutoResetPointsEncoding:            string(executionInfo.AutoResetPoints.GetEncoding()),
		SearchAttributes:                   executionInfo.SearchAttributes,
		Memo:                               executionInfo.Memo,
		CompletionEventEncoding:            string(common.EncodingTypeEmpty),
		VersionHistoriesEncoding:           string(common.EncodingTypeEmpty),
		InitiatedID:                        common.EmptyEventID,
	}

	completionEvent := executionInfo.CompletionEvent
	if completionEvent != nil {
		info.CompletionEvent = completionEvent.Data
		info.CompletionEventEncoding = string(completionEvent.Encoding)
	}

	info.StartVersion = startVersion
	if versionHistories == nil {
		// this is allowed
	} else {
		info.VersionHistories = versionHistories.Data
		info.VersionHistoriesEncoding = string(versionHistories.GetEncoding())
	}

	if executionInfo.ParentDomainID != "" {
		info.ParentDomainID = serialization.MustParseUUID(executionInfo.ParentDomainID)
		info.ParentWorkflowID = executionInfo.ParentWorkflowID
		info.ParentRunID = serialization.MustParseUUID(executionInfo.ParentRunID)
		info.InitiatedID = executionInfo.InitiatedID
		info.CompletionEvent = nil
	}

	if executionInfo.CancelRequested {
		info.CancelRequested = true
		info.CancelRequestID = executionInfo.CancelRequestID
	}

	blob, err := parser.WorkflowExecutionInfoToBlob(info)
	if err != nil {
		return nil, err
	}

	return &sqlplugin.ExecutionsRow{
		ShardID:          shardID,
		DomainID:         serialization.MustParseUUID(executionInfo.DomainID),
		WorkflowID:       executionInfo.WorkflowID,
		RunID:            serialization.MustParseUUID(executionInfo.RunID),
		NextEventID:      int64(executionInfo.NextEventID),
		LastWriteVersion: lastWriteVersion,
		Data:             blob.Data,
		DataEncoding:     string(blob.Encoding),
	}, nil
}

func (m *sqlExecutionStore) createExecution(
	ctx context.Context,
	tx sqlplugin.Tx,
	executionInfo *p.InternalWorkflowExecutionInfo,
	versionHistories *p.DataBlob,
	startVersion int64,
	lastWriteVersion int64,
	shardID int,
	parser serialization.Parser,
) error {

	// validate workflow state & close status
	if err := p.ValidateCreateWorkflowStateCloseStatus(
		executionInfo.State,
		executionInfo.CloseStatus); err != nil {
		return err
	}

	// TODO we should set the start time and last update time on business logic layer
	executionInfo.StartTimestamp = time.Now()
	executionInfo.LastUpdatedTimestamp = executionInfo.StartTimestamp

	row, err := buildExecutionRow(
		executionInfo,
		versionHistories,
		startVersion,
		lastWriteVersion,
		shardID,
		parser,
	)
	if err != nil {
		return err
	}
	result, err := tx.InsertIntoExecutions(ctx, row)
	if err != nil {
		if m.db.IsDupEntryError(err) {
			return &p.WorkflowExecutionAlreadyStartedError{
				Msg:              fmt.Sprintf("Workflow execution already running. WorkflowId: %v", executionInfo.WorkflowID),
				StartRequestID:   executionInfo.CreateRequestID,
				RunID:            executionInfo.RunID,
				State:            executionInfo.State,
				CloseStatus:      executionInfo.CloseStatus,
				LastWriteVersion: row.LastWriteVersion,
			}
		}
		return convertCommonErrors(tx, "createExecution", "", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("createExecution failed. Failed to verify number of rows affected. Erorr: %v", err),
		}
	}
	if rowsAffected != 1 {
		return &types.EntityNotExistsError{
			Message: fmt.Sprintf("createExecution failed. Affected %v rows updated instead of 1.", rowsAffected),
		}
	}

	return nil
}

func updateExecution(
	ctx context.Context,
	tx sqlplugin.Tx,
	executionInfo *p.InternalWorkflowExecutionInfo,
	versionHistories *p.DataBlob,
	startVersion int64,
	lastWriteVersion int64,
	shardID int,
	parser serialization.Parser,
) error {

	// validate workflow state & close status
	if err := p.ValidateUpdateWorkflowStateCloseStatus(
		executionInfo.State,
		executionInfo.CloseStatus); err != nil {
		return err
	}

	// TODO we should set the last update time on business logic layer
	executionInfo.LastUpdatedTimestamp = time.Now()

	row, err := buildExecutionRow(
		executionInfo,
		versionHistories,
		startVersion,
		lastWriteVersion,
		shardID,
		parser,
	)
	if err != nil {
		return err
	}
	result, err := tx.UpdateExecutions(ctx, row)
	if err != nil {
		return convertCommonErrors(tx, "updateExecution", "", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("updateExecution failed. Failed to verify number of rows affected. Erorr: %v", err),
		}
	}
	if rowsAffected != 1 {
		return &types.EntityNotExistsError{
			Message: fmt.Sprintf("updateExecution failed. Affected %v rows updated instead of 1.", rowsAffected),
		}
	}

	return nil
}
