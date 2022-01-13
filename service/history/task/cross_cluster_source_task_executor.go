// Copyright (c) 2021 Uber Technologies, Inc.
// Portions of the Software are attributed to Copyright (c) 2021 Temporal Technologies Inc.
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

package task

import (
	"context"
	"errors"
	"fmt"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	ctask "github.com/uber/cadence/common/task"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
)

var (
	errContinueExecution = errors.New("cross cluster task should continue execution")
)

type (
	crossClusterSourceTaskExecutor struct {
		shard          shard.Context
		executionCache *execution.Cache
		logger         log.Logger
		metricsClient  metrics.Client
	}
)

func NewCrossClusterSourceTaskExecutor(
	shard shard.Context,
	executionCache *execution.Cache,
	logger log.Logger,
) Executor {
	return &crossClusterSourceTaskExecutor{
		shard:          shard,
		executionCache: executionCache,
		logger:         logger,
		metricsClient:  shard.GetMetricsClient(),
	}
}

func (t *crossClusterSourceTaskExecutor) Execute(
	task Task,
	shouldProcessTask bool,
) error {
	sourceTask, ok := task.(*crossClusterSourceTask)
	if !ok {
		return errUnexpectedTask
	}

	if !shouldProcessTask {
		// this should not happen for cross cluster task
		// shouldProcessTask parameter will be deprecated after
		// 3+DC task lifecycle is done.
		return errors.New("encounter shouldProcessTask equals to false when processing cross cluster task at source cluster")
	}

	if err := t.verifyDomainActive(sourceTask); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), taskDefaultTimeout)
	defer cancel()

	var err error
	switch task.GetTaskType() {
	case persistence.CrossClusterTaskTypeStartChildExecution:
		err = t.executeStartChildExecutionTask(ctx, sourceTask)
	case persistence.CrossClusterTaskTypeCancelExecution:
		err = t.executeCancelExecutionTask(ctx, sourceTask)
	case persistence.CrossClusterTaskTypeSignalExecution:
		err = t.executeSignalExecutionTask(ctx, sourceTask)
	case persistence.CrossClusterTaskTypeRecordChildExeuctionCompleted:
		err = t.executeRecordChildWorkflowExecutionCompleteTask(ctx, sourceTask)
	case persistence.CrossClusterTaskTypeApplyParentClosePolicy:
		err = t.executeApplyParentClosePolicyTask(ctx, sourceTask)
	default:
		err = errUnknownCrossClusterTask
	}

	if err == nil {
		t.setTaskState(sourceTask, ctask.TaskStateAcked, processingStateResponseReported)
	} else if err == errContinueExecution {
		err = nil
	}

	return err
}

func (t *crossClusterSourceTaskExecutor) executeStartChildExecutionTask(
	ctx context.Context,
	task *crossClusterSourceTask,
) (retError error) {

	taskInfo := task.GetInfo().(*persistence.CrossClusterTaskInfo)
	wfContext, mutableState, release, err := loadWorkflowForCrossClusterTask(ctx, t.executionCache, taskInfo, t.metricsClient, t.logger)
	if err != nil || mutableState == nil {
		return err
	}
	defer func() { release(retError) }()

	initiatedEventID := taskInfo.ScheduleID
	childInfo, ok := mutableState.GetChildExecutionInfo(initiatedEventID)
	if !ok {
		return nil
	}

	ok, err = verifyTaskVersion(t.shard, t.logger, taskInfo.DomainID, childInfo.Version, taskInfo.Version, task)
	if err != nil || !ok {
		return err
	}

	if !mutableState.IsWorkflowExecutionRunning() &&
		(childInfo.StartedID == common.EmptyEventID ||
			childInfo.ParentClosePolicy != types.ParentClosePolicyAbandon) {
		return nil
	}

	if task.ProcessingState() == processingStateInvalidated {
		return t.generateNewTask(ctx, wfContext, mutableState, task)
	}

	// handle common errors
	failedCause := task.response.FailedCause
	if failedCause != nil {
		switch *failedCause {
		case types.CrossClusterTaskFailedCauseDomainNotActive:
			return t.generateNewTask(ctx, wfContext, mutableState, task)
		case types.CrossClusterTaskFailedCauseWorkflowNotExists:
			return &types.EntityNotExistsError{}
		case types.CrossClusterTaskFailedCauseWorkflowAlreadyCompleted:
			return &types.WorkflowExecutionAlreadyCompletedError{}
		case types.CrossClusterTaskFailedCauseUncategorized:
			// ideally this case should not happen
			// crossClusterSourceTask.Update() will reject response with this failed cause
			t.setTaskState(task, ctask.TaskStatePending, processingState(task.response.TaskState))
			return errContinueExecution
		}
	}

	switch processingState(task.response.TaskState) {
	case processingStateInitialized:
		if childInfo.StartedID != common.EmptyEventID {
			t.setTaskState(task, ctask.TaskStatePending, processingStateResponseRecorded)
			return errContinueExecution
		}

		initiatedEvent, err := mutableState.GetChildExecutionInitiatedEvent(ctx, initiatedEventID)
		if err != nil {
			return err
		}

		attributes := initiatedEvent.StartChildWorkflowExecutionInitiatedEventAttributes
		now := t.shard.GetTimeSource().Now()
		if failedCause != nil &&
			(*failedCause == types.CrossClusterTaskFailedCauseDomainNotExists ||
				*failedCause == types.CrossClusterTaskFailedCauseWorkflowAlreadyRunning) {
			return recordStartChildExecutionFailed(ctx, taskInfo, wfContext, attributes, now)
		}

		childRunID := task.response.StartChildExecutionAttributes.GetRunID()
		err = recordChildExecutionStarted(ctx, taskInfo, wfContext, attributes, childRunID, now)
		if err != nil {
			return err
		}

		// advance the task state so that it will be available for polling again
		t.setTaskState(task, ctask.TaskStatePending, processingStateResponseRecorded)
		return errContinueExecution
	case processingStateResponseRecorded:
		// there's nothing we need to do for the remaining two failed causes
		// ack the task
		return nil
	default:
		return errUnknownTaskProcessingState
	}
}

func (t *crossClusterSourceTaskExecutor) executeCancelExecutionTask(
	ctx context.Context,
	task *crossClusterSourceTask,
) (retError error) {

	taskInfo := task.GetInfo().(*persistence.CrossClusterTaskInfo)
	wfContext, mutableState, release, err := loadWorkflowForCrossClusterTask(ctx, t.executionCache, taskInfo, t.metricsClient, t.logger)
	if err != nil || mutableState == nil {
		return err
	}
	defer func() { release(retError) }()

	if !mutableState.IsWorkflowExecutionRunning() {
		return nil
	}

	initiatedEventID := taskInfo.ScheduleID
	requestCancelInfo, ok := mutableState.GetRequestCancelInfo(initiatedEventID)
	if !ok {
		return nil
	}
	ok, err = verifyTaskVersion(t.shard, t.logger, taskInfo.DomainID, requestCancelInfo.Version, taskInfo.Version, task)
	if err != nil || !ok {
		return err
	}

	if task.ProcessingState() == processingStateInvalidated {
		return t.generateNewTask(ctx, wfContext, mutableState, task)
	}

	// handle common errors
	failedCause := task.response.FailedCause
	if failedCause != nil {
		switch *failedCause {
		case types.CrossClusterTaskFailedCauseDomainNotActive:
			return t.generateNewTask(ctx, wfContext, mutableState, task)
		case types.CrossClusterTaskFailedCauseWorkflowAlreadyRunning,
			types.CrossClusterTaskFailedCauseUncategorized:
			// ideally this case should not happen
			// crossClusterSourceTask.Update() will reject response with those failed causes
			t.setTaskState(task, ctask.TaskStatePending, processingState(task.response.TaskState))
			return errContinueExecution
		}
	}

	switch processingState(task.response.TaskState) {
	case processingStateInitialized:
		now := t.shard.GetTimeSource().Now()
		targetDomainName, err := t.shard.GetDomainCache().GetDomainName(taskInfo.TargetDomainID)
		if err != nil {
			return err
		}

		if failedCause != nil {
			// remaining errors are non-retryable
			return requestCancelExternalExecutionFailed(
				ctx,
				taskInfo,
				wfContext,
				targetDomainName,
				taskInfo.TargetWorkflowID,
				taskInfo.TargetRunID,
				now,
			)
		}
		return requestCancelExternalExecutionCompleted(
			ctx,
			taskInfo,
			wfContext,
			targetDomainName,
			taskInfo.TargetWorkflowID,
			taskInfo.TargetRunID,
			now,
		)

	default:
		return errUnknownTaskProcessingState
	}
}

func (t *crossClusterSourceTaskExecutor) executeApplyParentClosePolicyTask(
	ctx context.Context,
	task *crossClusterSourceTask,
) (retError error) {
	taskInfo := task.GetInfo().(*persistence.CrossClusterTaskInfo)
	wfContext, mutableState, release, err := loadWorkflowForCrossClusterTask(ctx, t.executionCache, taskInfo, t.metricsClient, t.logger)
	if err != nil || mutableState == nil {
		return err
	}
	defer func() { release(retError) }()

	if mutableState.IsWorkflowExecutionRunning() {
		return nil
	}

	verified, err := task.VerifyLastWriteVersion(mutableState, taskInfo)
	if err != nil || !verified {
		return err
	}

	if task.ProcessingState() == processingStateInvalidated {
		return t.generateNewTask(ctx, wfContext, mutableState, task)
	}

	if task.response == nil || task.response.ApplyParentClosePolicyAttributes == nil {
		// this should never happen but we can't fix it by retrying. So log the event and return nil
		t.logger.Error(fmt.Sprintf(
			"Cross Cluster ApplyParentClosePolicy task response is invalid. Task: %#v, response: %#v.",
			task,
			task.response))
		t.setTaskState(task, ctask.TaskStatePending, processingState(task.response.TaskState))
		return errContinueExecution
	}

	// When processing state = processingStateInitialized, there's nothing we need to do
	// task is already complete, ack the task. All the other states are invalid
	if processingState(task.response.TaskState) != processingStateInitialized {
		return errUnknownTaskProcessingState
	}

	childrenStatus := task.response.ApplyParentClosePolicyAttributes.ChildrenStatus

	failedDomains := map[string]struct{}{}
	domainsToRegenerateTask := map[string]struct{}{}
	scope := t.metricsClient.Scope(metrics.CrossClusterSourceTaskApplyParentClosePolicyScope)

	for _, result := range childrenStatus {
		// handle common errors
		failedCause := result.FailedCause
		if failedCause != nil {
			switch *failedCause {
			case types.CrossClusterTaskFailedCauseDomainNotActive:
				domainsToRegenerateTask[result.Child.ChildDomainID] = struct{}{}
				scope.IncCounter(metrics.ParentClosePolicyProcessorFailures)
			case types.CrossClusterTaskFailedCauseWorkflowNotExists,
				types.CrossClusterTaskFailedCauseWorkflowAlreadyCompleted:
				// Do nothing, these errors are expected if the target workflow is already closed
				scope.IncCounter(metrics.ParentClosePolicyProcessorSuccess)
			default:
				t.logger.Error(fmt.Sprintf(
					"Unexpected CrossCluster ApplyParentClosePolicy Error: %#v",
					failedCause.String()))
				failedDomains[result.Child.ChildDomainID] = struct{}{}
				scope.IncCounter(metrics.ParentClosePolicyProcessorFailures)
			}
		} else {
			scope.IncCounter(metrics.ParentClosePolicyProcessorSuccess)
		}
	}

	if len(domainsToRegenerateTask) > 0 {
		// since there are domains we can retry, we need to combine them with failed ones too for the retry
		for domainID := range failedDomains {
			domainsToRegenerateTask[domainID] = struct{}{}
		}
		taskInfo := task.GetInfo().(*persistence.CrossClusterTaskInfo)
		taskInfo.TargetDomainIDs = domainsToRegenerateTask
		return t.generateNewTask(ctx, wfContext, mutableState, task)
	}
	if len(failedDomains) > 0 {
		taskInfo.TargetDomainIDs = failedDomains
		t.setTaskState(task, ctask.TaskStatePending, processingStateInitialized)

		return errContinueExecution
	}

	return nil
}

func (t *crossClusterSourceTaskExecutor) executeRecordChildWorkflowExecutionCompleteTask(
	ctx context.Context,
	task *crossClusterSourceTask,
) (retError error) {
	taskInfo := task.GetInfo().(*persistence.CrossClusterTaskInfo)
	wfContext, mutableState, release, err := loadWorkflowForCrossClusterTask(ctx, t.executionCache, taskInfo, t.metricsClient, t.logger)
	if err != nil || mutableState == nil {
		return err
	}
	defer func() { release(retError) }()

	if mutableState.IsWorkflowExecutionRunning() {
		return nil
	}

	verified, err := task.VerifyLastWriteVersion(mutableState, taskInfo)
	if err != nil || !verified {
		return err
	}

	if task.ProcessingState() == processingStateInvalidated {
		return t.generateNewTask(ctx, wfContext, mutableState, task)
	}

	// handle common errors
	failedCause := task.response.FailedCause
	if failedCause != nil {
		switch *failedCause {
		case types.CrossClusterTaskFailedCauseDomainNotActive:
			return t.generateNewTask(ctx, wfContext, mutableState, task)
		case types.CrossClusterTaskFailedCauseWorkflowNotExists,
			types.CrossClusterTaskFailedCauseWorkflowAlreadyCompleted:
			// Do nothing, these errors are expected if the target workflow is already closed
		default:
			t.setTaskState(task, ctask.TaskStatePending, processingState(task.response.TaskState))
			return errContinueExecution
		}
	}

	switch processingState(task.response.TaskState) {
	case processingStateInitialized:
		// there's nothing we need to do when record child completion is complete
		// ack the task
		return nil
	// processingStateResponseRecorded state is invalid for this operation
	default:
		return errUnknownTaskProcessingState
	}
}

func (t *crossClusterSourceTaskExecutor) executeSignalExecutionTask(
	ctx context.Context,
	task *crossClusterSourceTask,
) (retError error) {

	taskInfo := task.GetInfo().(*persistence.CrossClusterTaskInfo)
	wfContext, mutableState, release, err := loadWorkflowForCrossClusterTask(ctx, t.executionCache, taskInfo, t.metricsClient, t.logger)
	if err != nil || mutableState == nil {
		return err
	}
	defer func() { release(retError) }()

	if !mutableState.IsWorkflowExecutionRunning() {
		return nil
	}

	initiatedEventID := taskInfo.ScheduleID
	signalInfo, ok := mutableState.GetSignalInfo(initiatedEventID)
	if !ok {
		// TODO: here we should also RemoveSignalMutableState from target workflow
		// Otherwise, target SignalRequestID still can leak if shard restart after signalExternalExecutionCompleted
		// To do that, probably need to add the SignalRequestID in task info
		return nil
	}
	ok, err = verifyTaskVersion(t.shard, t.logger, taskInfo.DomainID, signalInfo.Version, taskInfo.Version, task)
	if err != nil || !ok {
		return err
	}

	if task.ProcessingState() == processingStateInvalidated {
		return t.generateNewTask(ctx, wfContext, mutableState, task)
	}

	// handle common errors
	failedCause := task.response.FailedCause
	if failedCause != nil {
		switch *failedCause {
		case types.CrossClusterTaskFailedCauseDomainNotActive:
			return t.generateNewTask(ctx, wfContext, mutableState, task)
		case types.CrossClusterTaskFailedCauseWorkflowAlreadyRunning,
			types.CrossClusterTaskFailedCauseUncategorized:
			// ideally this case should not happen
			// crossClusterSourceTask.Update() will reject response with those failed causes
			t.setTaskState(task, ctask.TaskStatePending, processingState(task.response.TaskState))
			return errContinueExecution
		}
	}

	switch processingState(task.response.TaskState) {
	case processingStateInitialized:
		now := t.shard.GetTimeSource().Now()
		targetDomainName, err := t.shard.GetDomainCache().GetDomainName(taskInfo.TargetDomainID)
		if err != nil {
			return err
		}

		if failedCause != nil {
			// remaining errors are non-retryable
			return signalExternalExecutionFailed(
				ctx,
				taskInfo,
				wfContext,
				targetDomainName,
				taskInfo.TargetWorkflowID,
				taskInfo.TargetRunID,
				signalInfo.Control,
				now,
			)
		}

		if err := signalExternalExecutionCompleted(
			ctx,
			taskInfo,
			wfContext,
			targetDomainName,
			taskInfo.TargetWorkflowID,
			taskInfo.TargetRunID,
			signalInfo.Control,
			now,
		); err != nil {
			return err
		}

		// advance the task state so that it will be available for polling again
		t.setTaskState(task, ctask.TaskStatePending, processingStateResponseRecorded)
		return errContinueExecution
	case processingStateResponseRecorded:
		// there's nothing we need to do for the response or remaining failed cause
		// ack the task
		return nil
	default:
		return errUnknownTaskProcessingState
	}
}

func (t *crossClusterSourceTaskExecutor) generateNewTask(
	ctx context.Context,
	wfContext execution.Context,
	mutableState execution.MutableState,
	task *crossClusterSourceTask,
) error {
	clusterMetadata := t.shard.GetClusterMetadata()
	taskGenerator := execution.NewMutableStateTaskGenerator(
		clusterMetadata,
		t.shard.GetDomainCache(),
		t.logger,
		mutableState,
	)
	taskInfo := task.GetInfo().(*persistence.CrossClusterTaskInfo)
	if err := taskGenerator.GenerateFromCrossClusterTask(taskInfo); err != nil {
		return err
	}

	err := wfContext.UpdateWorkflowExecutionTasks(ctx, t.shard.GetTimeSource().Now())
	if err != nil {
		return err
	}

	t.setTaskState(task, ctask.TaskStateAcked, processingStateInvalidated)
	return nil
}

func (t *crossClusterSourceTaskExecutor) verifyDomainActive(
	task *crossClusterSourceTask,
) error {
	entry, err := t.shard.GetDomainCache().GetDomainByID(task.GetDomainID())
	if err != nil {
		return err
	}

	if entry.IsDomainPendingActive() {
		// return error so that the task can be retried
		return ErrTaskPendingActive
	}

	if !entry.IsDomainActive() {
		// set processing state to invalidated so that a new task can be created
		t.setTaskState(task, ctask.TaskStatePending, processingStateInvalidated)
		return nil
	}

	return nil
}

func (t *crossClusterSourceTaskExecutor) setTaskState(
	task *crossClusterSourceTask,
	state ctask.State,
	processingState processingState,
) {
	task.Lock()
	task.state = state
	task.processingState = processingState
	task.response = nil
	task.Unlock()
}
