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
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/shard"
	"github.com/uber/cadence/service/worker/parentclosepolicy"
)

var (
	errUnknownTaskProcessingState   = errors.New("unknown cross cluster task processing state")
	errMissingTaskRequestAttributes = errors.New("request attributes not specified")
	errDomainNotExists              = errors.New("domain not exists")
	errDomainStandby                = errors.New("domain is standby in current cluster")
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
		TaskID:   targetTask.GetTaskID(),
		TaskType: targetTask.request.TaskInfo.TaskType,
	}

	if !shouldProcessTask {
		// this should not happen,
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
	default:
		err = errUnknownTransferTask
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

	targetDomainName, err := t.verifyAndTargetDomainName(attributes.TargetDomainID)
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

	targetDomainName, err := t.verifyAndTargetDomainName(attributes.TargetDomainID)
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

func (t *crossClusterTargetTaskExecutor) executeSignalExecutionTask(
	ctx context.Context,
	task *crossClusterTargetTask,
) (*types.CrossClusterSignalExecutionResponseAttributes, error) {
	attributes := task.request.SignalExecutionAttributes
	if attributes == nil {
		return nil, errMissingTaskRequestAttributes
	}

	targetDomainName, err := t.verifyAndTargetDomainName(attributes.TargetDomainID)
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

	if err == errDomainNotExists {
		return types.CrossClusterTaskFailedCauseDomainNotExists.Ptr(), false
	}
	if err == errDomainStandby {
		return types.CrossClusterTaskFailedCauseDomainNotActive.Ptr(), false
	}
	if err == ErrTaskPendingActive {
		return types.CrossClusterTaskFailedCauseDomainNotActive.Ptr(), true
	}
	if err == errUnknownTaskProcessingState || err == errMissingTaskRequestAttributes {
		return types.CrossClusterTaskFailedCauseUncategorized.Ptr(), false
	}
	return types.CrossClusterTaskFailedCauseUncategorized.Ptr(), true
}

func (t *crossClusterTargetTaskExecutor) verifyAndTargetDomainName(
	targetDomainID string,
) (string, error) {
	entry, err := t.shard.GetDomainCache().GetDomainByID(targetDomainID)
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
		return "", errDomainStandby
	}

	return entry.GetInfo().Name, nil
}
