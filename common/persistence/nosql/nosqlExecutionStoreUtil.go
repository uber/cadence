// Copyright (c) 2017-2021 Uber Technologies, Inc.
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

package nosql

import (
	"context"
	"fmt"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/checksum"
	"github.com/uber/cadence/common/log/tag"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/types"
)

func (d *nosqlExecutionStore) prepareCreateWorkflowExecutionRequestWithMaps(
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

func (d *nosqlExecutionStore) prepareResetWorkflowExecutionRequestWithMapsAndEventBuffer(
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

func (d *nosqlExecutionStore) prepareUpdateWorkflowExecutionRequestWithMapsAndEventBuffer(
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

func (d *nosqlExecutionStore) prepareTimerTasksForWorkflowTxn(
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
			TaskType:   task.GetType(),
			DomainID:   domainID,
			WorkflowID: workflowID,
			RunID:      runID,

			VisibilityTimestamp: task.GetVisibilityTimestamp(),
			TaskID:              task.GetTaskID(),

			TimeoutType:     timeoutType,
			EventID:         eventID,
			ScheduleAttempt: attempt,
			Version:         task.GetVersion(),
		}
		tasks = append(tasks, nt)
	}

	return tasks, nil
}

func (d *nosqlExecutionStore) prepareReplicationTasksForWorkflowTxn(
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
				Message: fmt.Sprintf("Unknown replication type: %v", task.GetType()),
			}
		}

		nt := &nosqlplugin.ReplicationTask{
			TaskType:          task.GetType(),
			DomainID:          domainID,
			WorkflowID:        workflowID,
			RunID:             runID,
			CreationTime:      task.GetVisibilityTimestamp(),
			TaskID:            task.GetTaskID(),
			FirstEventID:      firstEventID,
			NextEventID:       nextEventID,
			Version:           version,
			ScheduledID:       activityScheduleID,
			BranchToken:       branchToken,
			NewRunBranchToken: newRunBranchToken,
		}
		tasks = append(tasks, nt)
	}

	return tasks, nil
}

func (d *nosqlExecutionStore) prepareCrossClusterTasksForWorkflowTxn(
	domainID, workflowID, runID string,
	crossClusterTasks []p.Task,
) ([]*nosqlplugin.CrossClusterTask, error) {
	var tasks []*nosqlplugin.CrossClusterTask

	for _, task := range crossClusterTasks {
		var taskList string
		var scheduleID int64
		var targetCluster string
		targetDomainID := domainID // default to source domain, can't be empty, since empty string is not valid UUID
		targetDomainIDs := map[string]struct{}{}
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
			if targetRunID == "" {
				targetRunID = p.CrossClusterTaskDefaultTargetRunID
			}
			targetChildWorkflowOnly = task.(*p.CrossClusterCancelExecutionTask).TargetChildWorkflowOnly
			scheduleID = task.(*p.CrossClusterCancelExecutionTask).InitiatedID

		case p.CrossClusterTaskTypeSignalExecution:
			targetCluster = task.(*p.CrossClusterSignalExecutionTask).TargetCluster
			targetDomainID = task.(*p.CrossClusterSignalExecutionTask).TargetDomainID
			targetWorkflowID = task.(*p.CrossClusterSignalExecutionTask).TargetWorkflowID
			targetRunID = task.(*p.CrossClusterSignalExecutionTask).TargetRunID
			if targetRunID == "" {
				targetRunID = p.CrossClusterTaskDefaultTargetRunID
			}
			targetChildWorkflowOnly = task.(*p.CrossClusterSignalExecutionTask).TargetChildWorkflowOnly
			scheduleID = task.(*p.CrossClusterSignalExecutionTask).InitiatedID

		case p.CrossClusterTaskTypeRecordChildExeuctionCompleted:
			targetCluster = task.(*p.CrossClusterRecordChildExecutionCompletedTask).TargetCluster
			targetDomainID = task.(*p.CrossClusterRecordChildExecutionCompletedTask).TargetDomainID
			targetWorkflowID = task.(*p.CrossClusterRecordChildExecutionCompletedTask).TargetWorkflowID
			targetRunID = task.(*p.CrossClusterRecordChildExecutionCompletedTask).TargetRunID
			if targetRunID == "" {
				targetRunID = p.CrossClusterTaskDefaultTargetRunID
			}

		case p.CrossClusterTaskTypeApplyParentClosePolicy:
			targetCluster = task.(*p.CrossClusterApplyParentClosePolicyTask).TargetCluster
			targetDomainIDs = task.(*p.CrossClusterApplyParentClosePolicyTask).TargetDomainIDs

		default:
			return nil, &types.InternalServiceError{
				Message: fmt.Sprintf("Unknown cross-cluster task type: %v", task.GetType()),
			}
		}

		nt := &nosqlplugin.CrossClusterTask{
			TransferTask: nosqlplugin.TransferTask{
				TaskType:                task.GetType(),
				DomainID:                domainID,
				WorkflowID:              workflowID,
				RunID:                   runID,
				VisibilityTimestamp:     task.GetVisibilityTimestamp(),
				TaskID:                  task.GetTaskID(),
				TargetDomainID:          targetDomainID,
				TargetDomainIDs:         targetDomainIDs,
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

func (d *nosqlExecutionStore) prepareNoSQLTasksForWorkflowTxn(
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

func (d *nosqlExecutionStore) prepareTransferTasksForWorkflowTxn(
	domainID, workflowID, runID string,
	transferTasks []p.Task,
) ([]*nosqlplugin.TransferTask, error) {
	var tasks []*nosqlplugin.TransferTask

	for _, task := range transferTasks {
		var taskList string
		var scheduleID int64
		targetDomainID := domainID
		targetDomainIDs := map[string]struct{}{}
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

		case p.TransferTaskTypeRecordChildExecutionCompleted:
			targetDomainID = task.(*p.RecordChildExecutionCompletedTask).TargetDomainID
			targetWorkflowID = task.(*p.RecordChildExecutionCompletedTask).TargetWorkflowID
			targetRunID = task.(*p.RecordChildExecutionCompletedTask).TargetRunID
			if targetRunID == "" {
				targetRunID = p.TransferTaskTransferTargetRunID
			}

		case p.TransferTaskTypeApplyParentClosePolicy:
			targetDomainIDs = task.(*p.ApplyParentClosePolicyTask).TargetDomainIDs

		case p.TransferTaskTypeCloseExecution,
			p.TransferTaskTypeRecordWorkflowStarted,
			p.TransferTaskTypeResetWorkflow,
			p.TransferTaskTypeUpsertWorkflowSearchAttributes,
			p.TransferTaskTypeRecordWorkflowClosed:
			// No explicit property needs to be set

		default:
			return nil, &types.InternalServiceError{
				Message: fmt.Sprintf("Unknown transfer type: %v", task.GetType()),
			}
		}
		t := &nosqlplugin.TransferTask{
			TaskType:                task.GetType(),
			DomainID:                domainID,
			WorkflowID:              workflowID,
			RunID:                   runID,
			VisibilityTimestamp:     task.GetVisibilityTimestamp(),
			TaskID:                  task.GetTaskID(),
			TargetDomainID:          targetDomainID,
			TargetDomainIDs:         targetDomainIDs,
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

func (d *nosqlExecutionStore) prepareActivityInfosForWorkflowTxn(
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

func (d *nosqlExecutionStore) prepareTimerInfosForWorkflowTxn(
	timerInfo []*p.TimerInfo,
) (map[string]*p.TimerInfo, error) {
	m := map[string]*p.TimerInfo{}
	for _, a := range timerInfo {
		m[a.TimerID] = a
	}
	return m, nil
}

func (d *nosqlExecutionStore) prepareChildWFInfosForWorkflowTxn(
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

func (d *nosqlExecutionStore) prepareRequestCancelsForWorkflowTxn(
	requestCancels []*p.RequestCancelInfo,
) (map[int64]*p.RequestCancelInfo, error) {
	m := map[int64]*p.RequestCancelInfo{}
	for _, c := range requestCancels {
		m[c.InitiatedID] = c
	}
	return m, nil
}

func (d *nosqlExecutionStore) prepareSignalInfosForWorkflowTxn(
	signalInfos []*p.SignalInfo,
) (map[int64]*p.SignalInfo, error) {
	m := map[int64]*p.SignalInfo{}
	for _, c := range signalInfos {
		m[c.InitiatedID] = c
	}
	return m, nil
}

func (d *nosqlExecutionStore) prepareUpdateWorkflowExecutionTxn(
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

func (d *nosqlExecutionStore) prepareCreateWorkflowExecutionTxn(
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

	if executionInfo.StartTimestamp.IsZero() {
		executionInfo.StartTimestamp = nowTimestamp
		d.logger.Error("Workflow startTimestamp not set, fallback to now",
			tag.WorkflowDomainID(executionInfo.DomainID),
			tag.WorkflowID(executionInfo.WorkflowID),
			tag.WorkflowRunID(executionInfo.RunID),
		)
	}
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

func (d *nosqlExecutionStore) prepareCurrentWorkflowRequestForCreateWorkflowTxn(
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

func (d *nosqlExecutionStore) processUpdateWorkflowResult(err error, rangeID int64) error {
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
		return convertCommonErrors(d.db, "UpdateWorkflowExecution", err)
	}

	return nil
}

func (d *nosqlExecutionStore) assertNotCurrentExecution(
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
