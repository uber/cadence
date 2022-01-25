// Copyright (c) 2021 Uber Technologies, Inc.
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

	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/shard"
	"github.com/uber/cadence/service/worker/parentclosepolicy"
)

var (
	errUnknownCrossClusterTask      = errors.New("Unknown cross cluster task")
	errUnknownTaskProcessingState   = errors.New("unknown cross cluster task processing state")
	errMissingTaskRequestAttributes = errors.New("request attributes not specified")
	errDomainNotExists              = errors.New("domain not exists")
)

type (
	crossClusterTargetTaskExecutor struct {
		shard                   shard.Context
		logger                  log.Logger
		metricsClient           metrics.Client
		historyClient           history.Client
		parentClosePolicyClient parentclosepolicy.Client
		config                  *config.Config
	}
)

// NewCrossClusterTargetTaskExecutor creates a new task Executor for
// processing cross cluster tasks at target cluster
func NewCrossClusterTargetTaskExecutor(
	shard shard.Context,
	logger log.Logger,
	config *config.Config,
) Executor {
	return &crossClusterTargetTaskExecutor{
		shard:         shard,
		logger:        logger,
		metricsClient: shard.GetMetricsClient(),
		historyClient: shard.GetService().GetHistoryClient(),
		parentClosePolicyClient: parentclosepolicy.NewClient(
			shard.GetMetricsClient(),
			shard.GetLogger(),
			shard.GetService().GetSDKClient(),
			config.NumParentClosePolicySystemWorkflows(),
		),
		config: config,
	}
}

// Execute executes a cross cluster task at the target cluster.
//
// task.response field will be populated as long as the passed in
// task has the right type (*crossClusterTargetTask).
//
// Returned error will be non-nil only when the error is retryable.
// Non-retryable errors, such as domain/workflow not exists or
// errMissingTaskRequestAttributes, will not be returned,
// as when those error happens the task execution is considered as
// completed, and the error will be returned to the source cluster
// via the failedCause field in task.response
func (t *crossClusterTargetTaskExecutor) Execute(
	task Task,
	shouldProcessTask bool,
) error {
	targetTask, ok := task.(*crossClusterTargetTask)
	if !ok {
		return errUnexpectedTask
	}

	response := &types.CrossClusterTaskResponse{
		TaskID:    targetTask.GetTaskID(),
		TaskType:  targetTask.request.TaskInfo.TaskType,
		TaskState: targetTask.request.TaskInfo.TaskState,
	}

	if !shouldProcessTask {
		// this should not happen for cross cluster task
		// shouldProcessTask parameter will be deprecated after
		// 3+DC task lifecycle is done.
		response.FailedCause = types.CrossClusterTaskFailedCauseUncategorized.Ptr()
		targetTask.response = response
		return errors.New("encounter shouldProcessTask equals to false when processing cross cluster task at target cluster")
	}

	ctx, cancel := context.WithTimeout(context.Background(), taskDefaultTimeout)
	defer cancel()

	var err error
	switch task.GetTaskType() {
	case persistence.CrossClusterTaskTypeStartChildExecution:
		response.StartChildExecutionAttributes, err = t.executeStartChildExecutionTask(ctx, targetTask)
	case persistence.CrossClusterTaskTypeCancelExecution:
		response.CancelExecutionAttributes, err = t.executeCancelExecutionTask(ctx, targetTask)
	case persistence.CrossClusterTaskTypeSignalExecution:
		response.SignalExecutionAttributes, err = t.executeSignalExecutionTask(ctx, targetTask)
	case persistence.CrossClusterTaskTypeRecordChildExeuctionCompleted:
		response.RecordChildWorkflowExecutionCompleteAttributes, err = t.executeRecordChildWorkflowExecutionCompleteTask(ctx, targetTask)
	case persistence.CrossClusterTaskTypeApplyParentClosePolicy:
		response.ApplyParentClosePolicyAttributes, err = t.executeApplyParentClosePolicyTask(ctx, targetTask)
	default:
		err = errUnknownCrossClusterTask
	}

	var retryable bool
	response.FailedCause, retryable = t.convertErrorToFailureCause(err)
	targetTask.response = response

	// only return retryable/unexpected error so that
	// tasks won't be retried for expected errors like workflow not exists.
	// expected errors will be responded to source cluster through the failedCause field
	if retryable {
		return err
	}
	return nil
}

func (t *crossClusterTargetTaskExecutor) executeStartChildExecutionTask(
	ctx context.Context,
	task *crossClusterTargetTask,
) (*types.CrossClusterStartChildExecutionResponseAttributes, error) {
	attributes := task.request.StartChildExecutionAttributes
	if attributes == nil {
		return nil, errMissingTaskRequestAttributes
	}

	targetDomainName, err := t.verifyDomainActive(attributes.TargetDomainID)
	if err != nil {
		return nil, err
	}

	switch task.processingState {
	case processingStateInitialized:
		runID, err := startWorkflowWithRetry(
			ctx,
			t.historyClient,
			t.shard.GetTimeSource(),
			t.shard.GetDomainCache(),
			task.Info.(*persistence.CrossClusterTaskInfo),
			targetDomainName,
			attributes.RequestID,
			attributes.GetInitiatedEventAttributes(),
		)
		if err != nil {
			return nil, err
		}
		return &types.CrossClusterStartChildExecutionResponseAttributes{
			RunID: runID,
		}, nil
	case processingStateResponseRecorded:
		if err := createFirstDecisionTask(
			ctx,
			t.historyClient,
			attributes.TargetDomainID,
			&types.WorkflowExecution{
				WorkflowID: attributes.InitiatedEventAttributes.WorkflowID,
				RunID:      attributes.GetTargetRunID(),
			},
		); err != nil {
			return nil, err
		}
		return &types.CrossClusterStartChildExecutionResponseAttributes{
			RunID: attributes.GetTargetRunID(),
		}, nil
	default:
		return nil, errUnknownTaskProcessingState
	}
}

func (t *crossClusterTargetTaskExecutor) executeCancelExecutionTask(
	ctx context.Context,
	task *crossClusterTargetTask,
) (*types.CrossClusterCancelExecutionResponseAttributes, error) {
	attributes := task.request.CancelExecutionAttributes
	if attributes == nil {
		return nil, errMissingTaskRequestAttributes
	}

	targetDomainName, err := t.verifyDomainActive(attributes.TargetDomainID)
	if err != nil {
		return nil, err
	}

	if task.processingState != processingStateInitialized {
		return nil, errUnknownTaskProcessingState
	}

	if err := requestCancelExternalExecutionWithRetry(
		ctx,
		t.historyClient,
		task.Info.(*persistence.CrossClusterTaskInfo),
		targetDomainName,
		attributes.RequestID,
	); err != nil {
		return nil, err
	}
	return &types.CrossClusterCancelExecutionResponseAttributes{}, nil
}

func (t *crossClusterTargetTaskExecutor) executeApplyParentClosePolicyTask(
	ctx context.Context,
	task *crossClusterTargetTask,
) (*types.CrossClusterApplyParentClosePolicyResponseAttributes, error) {
	if task.processingState != processingStateInitialized {
		return nil, errUnknownTaskProcessingState
	}

	attributes := task.request.ApplyParentClosePolicyAttributes
	if attributes == nil {
		return nil, errMissingTaskRequestAttributes
	}

	response := types.CrossClusterApplyParentClosePolicyResponseAttributes{}
	var anyErr error

	for _, child := range attributes.Children {
		// DON'T RETURN ERROR INSIDE THIS LOOP, CONVERT THEM AND ASSIGN IT TO EACH CHILD'S FailedCause
		childAttrs := child.Child
		if child.Status != nil && child.Status.Completed {
			// This means we are in retry and this child was already successfully processed before
			response.ChildrenStatus = append(
				response.ChildrenStatus,
				&types.ApplyParentClosePolicyResult{
					Child:       childAttrs,
					FailedCause: child.Status.FailedCause,
				})
			continue
		}
		if child.Status == nil {
			child.Status = &types.ApplyParentClosePolicyStatus{}
		}
		child.Status.Completed = false
		var failedCause *types.CrossClusterTaskFailedCause
		retriable := false

		targetDomainName, err := t.verifyDomainActive(childAttrs.ChildDomainID)
		if err == nil {
			err = applyParentClosePolicy(
				ctx,
				t.historyClient,
				&types.WorkflowExecution{
					WorkflowID: task.GetWorkflowID(),
					RunID:      task.GetRunID(),
				},
				childAttrs.ChildDomainID,
				targetDomainName,
				childAttrs.ChildWorkflowID,
				childAttrs.ChildRunID,
				*childAttrs.ParentClosePolicy,
			)
		}

		switch err.(type) {
		case *types.EntityNotExistsError,
			*types.WorkflowExecutionAlreadyCompletedError,
			*types.CancellationAlreadyRequestedError:
			// expected error, no-op
			child.Status.Completed = true
		case nil:
			child.Status.Completed = true
		default:
			failedCause, retriable = t.convertErrorToFailureCause(err)
			child.Status.FailedCause = failedCause
			child.Status.Completed = !retriable
			// return error only if the error is retriable
			if retriable {
				anyErr = err
			}
		}
		// Optimize SOURCE SIDE RETRIES by returning each child status separately
		response.ChildrenStatus = append(
			response.ChildrenStatus,
			&types.ApplyParentClosePolicyResult{
				Child:       childAttrs,
				FailedCause: failedCause,
			})
	}

	return &response, anyErr
}

func (t *crossClusterTargetTaskExecutor) executeRecordChildWorkflowExecutionCompleteTask(
	ctx context.Context,
	task *crossClusterTargetTask,
) (*types.CrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes, error) {
	if task.processingState != processingStateInitialized {
		return nil, errUnknownTaskProcessingState
	}

	// parent info is in attributes
	attributes := task.request.RecordChildWorkflowExecutionCompleteAttributes
	if attributes == nil {
		return nil, errMissingTaskRequestAttributes
	}

	_, err := t.verifyDomainActive(attributes.TargetDomainID)
	if err != nil {
		return nil, err
	}

	// child info is in taskInfo
	taskInfo, ok := task.Info.(*persistence.CrossClusterTaskInfo)
	if !ok {
		return nil, errUnexpectedTask
	}

	recordChildCompletionCtx, cancel := context.WithTimeout(ctx, taskRPCCallTimeout)
	defer cancel()
	err = t.historyClient.RecordChildExecutionCompleted(recordChildCompletionCtx, &types.RecordChildExecutionCompletedRequest{
		DomainUUID: attributes.TargetDomainID, // parent domain id
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: attributes.TargetWorkflowID,
			RunID:      attributes.TargetRunID,
		},
		InitiatedID: attributes.InitiatedEventID,
		CompletedExecution: &types.WorkflowExecution{
			WorkflowID: taskInfo.WorkflowID,
			RunID:      taskInfo.RunID,
		},
		CompletionEvent: attributes.CompletionEvent,
	})

	if err != nil {
		return nil, err
	}
	return &types.CrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes{}, nil
}

func (t *crossClusterTargetTaskExecutor) executeSignalExecutionTask(
	ctx context.Context,
	task *crossClusterTargetTask,
) (*types.CrossClusterSignalExecutionResponseAttributes, error) {
	attributes := task.request.SignalExecutionAttributes
	if attributes == nil {
		return nil, errMissingTaskRequestAttributes
	}

	targetDomainName, err := t.verifyDomainActive(attributes.TargetDomainID)
	if err != nil {
		return nil, err
	}

	switch task.processingState {
	case processingStateInitialized:
		err = signalExternalExecutionWithRetry(
			ctx,
			t.historyClient,
			task.Info.(*persistence.CrossClusterTaskInfo),
			targetDomainName,
			&persistence.SignalInfo{
				InitiatedID:     attributes.InitiatedEventID,
				SignalName:      attributes.SignalName,
				Input:           attributes.SignalInput,
				SignalRequestID: attributes.RequestID,
				Control:         attributes.Control,
			},
		)
	case processingStateResponseRecorded:
		err = removeSignalMutableStateWithRetry(
			ctx,
			t.historyClient,
			task.Info.(*persistence.CrossClusterTaskInfo),
			attributes.RequestID,
		)
	default:
		err = errUnknownTaskProcessingState
	}

	if err != nil {
		return nil, err
	}
	return &types.CrossClusterSignalExecutionResponseAttributes{}, nil
}

func (t *crossClusterTargetTaskExecutor) convertErrorToFailureCause(
	err error,
) (*types.CrossClusterTaskFailedCause, bool) {
	switch err.(type) {
	case nil:
		return nil, false
	case *types.EntityNotExistsError:
		return types.CrossClusterTaskFailedCauseWorkflowNotExists.Ptr(), false
	case *types.WorkflowExecutionAlreadyStartedError:
		return types.CrossClusterTaskFailedCauseWorkflowAlreadyRunning.Ptr(), false
	case *types.WorkflowExecutionAlreadyCompletedError:
		return types.CrossClusterTaskFailedCauseWorkflowAlreadyCompleted.Ptr(), false
	case *types.DomainNotActiveError:
		return types.CrossClusterTaskFailedCauseDomainNotActive.Ptr(), false
	}

	switch err {
	case errDomainNotExists:
		return types.CrossClusterTaskFailedCauseDomainNotExists.Ptr(), false
	case errTargetDomainNotActive:
		return types.CrossClusterTaskFailedCauseDomainNotActive.Ptr(), false
	case ErrTaskPendingActive:
		return types.CrossClusterTaskFailedCauseDomainNotActive.Ptr(), true
	default:
		retryable := err != errUnknownTaskProcessingState && err != errMissingTaskRequestAttributes
		t.logger.Error("encountered uncategorized error when processing target cross cluster task", tag.Error(err))
		return types.CrossClusterTaskFailedCauseUncategorized.Ptr(), retryable
	}
}

func (t *crossClusterTargetTaskExecutor) verifyDomainActive(
	domainID string,
) (string, error) {
	entry, err := t.shard.GetDomainCache().GetDomainByID(domainID)
	if err != nil {
		if common.IsEntityNotExistsError(err) {
			// return a special error here so that we can tell the difference from
			// workflow not exists when handling the error
			return "", errDomainNotExists
		}
		return "", err
	}

	if entry.IsDomainPendingActive() {
		return "", ErrTaskPendingActive
	}

	if !entry.IsDomainActive() {
		return "", errTargetDomainNotActive
	}

	return entry.GetInfo().Name, nil
}
