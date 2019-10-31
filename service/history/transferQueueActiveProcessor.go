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

package history

import (
	ctx "context"
	"fmt"

	"github.com/pborman/uuid"

	h "github.com/uber/cadence/.gen/go/history"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/client/matching"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/service/worker/parentclosepolicy"
)

const identityHistoryService = "history-service"

type (
	transferQueueActiveProcessorImpl struct {
		currentClusterName      string
		shard                   ShardContext
		domainCache             cache.DomainCache
		historyService          *historyEngineImpl
		options                 *QueueProcessorOptions
		historyClient           history.Client
		cache                   *historyCache
		transferTaskFilter      taskFilter
		logger                  log.Logger
		metricsClient           metrics.Client
		parentClosePolicyClient parentclosepolicy.Client
		maxReadAckLevel         maxReadAckLevel
		*transferQueueProcessorBase
		*queueProcessorBase
		queueAckMgr
	}
)

func newTransferQueueActiveProcessor(
	shard ShardContext,
	historyService *historyEngineImpl,
	visibilityMgr persistence.VisibilityManager,
	matchingClient matching.Client,
	historyClient history.Client,
	taskAllocator taskAllocator,
	logger log.Logger,
) *transferQueueActiveProcessorImpl {

	config := shard.GetConfig()
	options := &QueueProcessorOptions{
		BatchSize:                          config.TransferTaskBatchSize,
		WorkerCount:                        config.TransferTaskWorkerCount,
		MaxPollRPS:                         config.TransferProcessorMaxPollRPS,
		MaxPollInterval:                    config.TransferProcessorMaxPollInterval,
		MaxPollIntervalJitterCoefficient:   config.TransferProcessorMaxPollIntervalJitterCoefficient,
		UpdateAckInterval:                  config.TransferProcessorUpdateAckInterval,
		UpdateAckIntervalJitterCoefficient: config.TransferProcessorUpdateAckIntervalJitterCoefficient,
		MaxRetryCount:                      config.TransferTaskMaxRetryCount,
		MetricScope:                        metrics.TransferActiveQueueProcessorScope,
	}
	currentClusterName := shard.GetService().GetClusterMetadata().GetCurrentClusterName()
	logger = logger.WithTags(tag.ClusterName(currentClusterName))
	transferTaskFilter := func(taskInfo *taskInfo) (bool, error) {
		task, ok := taskInfo.task.(*persistence.TransferTaskInfo)
		if !ok {
			return false, errUnexpectedQueueTask
		}
		return taskAllocator.verifyActiveTask(task.DomainID, task)
	}
	maxReadAckLevel := func() int64 {
		return shard.GetTransferMaxReadLevel()
	}
	updateTransferAckLevel := func(ackLevel int64) error {
		return shard.UpdateTransferClusterAckLevel(currentClusterName, ackLevel)
	}

	transferQueueShutdown := func() error {
		return nil
	}

	parentClosePolicyClient := parentclosepolicy.NewClient(
		shard.GetMetricsClient(),
		shard.GetLogger(),
		historyService.publicClient,
		shard.GetConfig().NumParentClosePolicySystemWorkflows())

	processor := &transferQueueActiveProcessorImpl{
		currentClusterName:      currentClusterName,
		shard:                   shard,
		domainCache:             shard.GetDomainCache(),
		historyService:          historyService,
		options:                 options,
		historyClient:           historyClient,
		logger:                  logger,
		metricsClient:           historyService.metricsClient,
		parentClosePolicyClient: parentClosePolicyClient,

		cache:              historyService.historyCache,
		transferTaskFilter: transferTaskFilter,
		transferQueueProcessorBase: newTransferQueueProcessorBase(
			shard,
			options,
			visibilityMgr,
			matchingClient,
			maxReadAckLevel,
			updateTransferAckLevel,
			transferQueueShutdown,
			logger,
		),
	}

	queueAckMgr := newQueueAckMgr(shard, options, processor, shard.GetTransferClusterAckLevel(currentClusterName), logger)
	queueProcessorBase := newQueueProcessorBase(currentClusterName, shard, options, processor, queueAckMgr, historyService.historyCache, logger)
	processor.queueAckMgr = queueAckMgr
	processor.queueProcessorBase = queueProcessorBase

	return processor
}

func newTransferQueueFailoverProcessor(
	shard ShardContext,
	historyService *historyEngineImpl,
	visibilityMgr persistence.VisibilityManager,
	matchingClient matching.Client,
	historyClient history.Client,
	domainIDs map[string]struct{},
	standbyClusterName string,
	minLevel int64,
	maxLevel int64,
	taskAllocator taskAllocator,
	logger log.Logger,
) (func(ackLevel int64) error, *transferQueueActiveProcessorImpl) {

	config := shard.GetConfig()
	options := &QueueProcessorOptions{
		BatchSize:                          config.TransferTaskBatchSize,
		WorkerCount:                        config.TransferTaskWorkerCount,
		MaxPollRPS:                         config.TransferProcessorFailoverMaxPollRPS,
		MaxPollInterval:                    config.TransferProcessorMaxPollInterval,
		MaxPollIntervalJitterCoefficient:   config.TransferProcessorMaxPollIntervalJitterCoefficient,
		UpdateAckInterval:                  config.TransferProcessorUpdateAckInterval,
		UpdateAckIntervalJitterCoefficient: config.TransferProcessorUpdateAckIntervalJitterCoefficient,
		MaxRetryCount:                      config.TransferTaskMaxRetryCount,
		MetricScope:                        metrics.TransferActiveQueueProcessorScope,
	}
	currentClusterName := shard.GetService().GetClusterMetadata().GetCurrentClusterName()
	failoverUUID := uuid.New()
	logger = logger.WithTags(
		tag.ClusterName(currentClusterName),
		tag.WorkflowDomainIDs(domainIDs),
		tag.FailoverMsg("from: "+standbyClusterName),
	)

	transferTaskFilter := func(taskInfo *taskInfo) (bool, error) {
		task, ok := taskInfo.task.(*persistence.TransferTaskInfo)
		if !ok {
			return false, errUnexpectedQueueTask
		}
		return taskAllocator.verifyFailoverActiveTask(domainIDs, task.DomainID, task)
	}
	maxReadAckLevel := func() int64 {
		return maxLevel // this is a const
	}
	failoverStartTime := shard.GetTimeSource().Now()
	updateTransferAckLevel := func(ackLevel int64) error {
		return shard.UpdateTransferFailoverLevel(
			failoverUUID,
			persistence.TransferFailoverLevel{
				StartTime:    failoverStartTime,
				MinLevel:     minLevel,
				CurrentLevel: ackLevel,
				MaxLevel:     maxLevel,
				DomainIDs:    domainIDs,
			},
		)
	}
	transferQueueShutdown := func() error {
		return shard.DeleteTransferFailoverLevel(failoverUUID)
	}

	processor := &transferQueueActiveProcessorImpl{
		currentClusterName: currentClusterName,
		shard:              shard,
		domainCache:        shard.GetDomainCache(),
		historyService:     historyService,
		options:            options,
		historyClient:      historyClient,
		logger:             logger,
		metricsClient:      historyService.metricsClient,
		cache:              historyService.historyCache,
		transferTaskFilter: transferTaskFilter,
		transferQueueProcessorBase: newTransferQueueProcessorBase(
			shard, options, visibilityMgr, matchingClient,
			maxReadAckLevel, updateTransferAckLevel, transferQueueShutdown, logger,
		),
	}

	queueAckMgr := newQueueFailoverAckMgr(shard, options, processor, minLevel, logger)
	queueProcessorBase := newQueueProcessorBase(currentClusterName, shard, options, processor, queueAckMgr, historyService.historyCache, logger)
	processor.queueAckMgr = queueAckMgr
	processor.queueProcessorBase = queueProcessorBase
	return updateTransferAckLevel, processor
}

func (t *transferQueueActiveProcessorImpl) getTaskFilter() taskFilter {
	return t.transferTaskFilter
}

func (t *transferQueueActiveProcessorImpl) notifyNewTask() {
	t.queueProcessorBase.notifyNewTask()
}

func (t *transferQueueActiveProcessorImpl) complete(
	taskInfo *taskInfo,
) {

	t.queueProcessorBase.complete(taskInfo.task)
}

func (t *transferQueueActiveProcessorImpl) process(
	taskInfo *taskInfo,
) (int, error) {

	task, ok := taskInfo.task.(*persistence.TransferTaskInfo)
	if !ok {
		return metrics.TransferActiveQueueProcessorScope, errUnexpectedQueueTask
	}

	var err error
	switch task.TaskType {
	case persistence.TransferTaskTypeActivityTask:
		if taskInfo.shouldProcessTask {
			err = t.processActivityTask(task)
		}
		return metrics.TransferActiveTaskActivityScope, err

	case persistence.TransferTaskTypeDecisionTask:
		if taskInfo.shouldProcessTask {
			err = t.processDecisionTask(task)
		}
		return metrics.TransferActiveTaskDecisionScope, err

	case persistence.TransferTaskTypeCloseExecution:
		if taskInfo.shouldProcessTask {
			err = t.processCloseExecution(task)
		}
		return metrics.TransferActiveTaskCloseExecutionScope, err

	case persistence.TransferTaskTypeCancelExecution:
		if taskInfo.shouldProcessTask {
			err = t.processCancelExecution(task)
		}
		return metrics.TransferActiveTaskCancelExecutionScope, err

	case persistence.TransferTaskTypeSignalExecution:
		if taskInfo.shouldProcessTask {
			err = t.processSignalExecution(task)
		}
		return metrics.TransferActiveTaskSignalExecutionScope, err

	case persistence.TransferTaskTypeStartChildExecution:
		if taskInfo.shouldProcessTask {
			err = t.processStartChildExecution(task)
		}
		return metrics.TransferActiveTaskStartChildExecutionScope, err

	case persistence.TransferTaskTypeRecordWorkflowStarted:
		if taskInfo.shouldProcessTask {
			err = t.processRecordWorkflowStarted(task)
		}
		return metrics.TransferActiveTaskRecordWorkflowStartedScope, err

	case persistence.TransferTaskTypeResetWorkflow:
		if taskInfo.shouldProcessTask {
			err = t.processResetWorkflow(task)
		}
		return metrics.TransferActiveTaskResetWorkflowScope, err

	case persistence.TransferTaskTypeUpsertWorkflowSearchAttributes:
		if taskInfo.shouldProcessTask {
			err = t.processUpsertWorkflowSearchAttributes(task)
		}
		return metrics.TransferActiveTaskUpsertWorkflowSearchAttributesScope, err

	default:
		return metrics.TransferActiveQueueProcessorScope, errUnknownTransferTask
	}
}

func (t *transferQueueActiveProcessorImpl) processActivityTask(
	task *persistence.TransferTaskInfo,
) (retError error) {

	var err error
	execution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(task.WorkflowID),
		RunId:      common.StringPtr(task.RunID)}

	context, release, err := t.cache.getOrCreateWorkflowExecutionForBackground(task.DomainID, execution)
	if err != nil {
		return err
	}
	defer func() { release(retError) }()

	var mutableState mutableState
	mutableState, err = loadMutableStateForTransferTask(context, task, t.metricsClient, t.logger)
	if err != nil {
		return err
	} else if mutableState == nil || !mutableState.IsWorkflowExecutionRunning() {
		return nil
	}

	ai, ok := mutableState.GetActivityInfo(task.ScheduleID)
	if !ok {
		t.logger.Debug("Potentially duplicate task.", tag.TaskID(task.TaskID), tag.WorkflowScheduleID(task.ScheduleID), tag.TaskType(persistence.TransferTaskTypeActivityTask))
		return nil
	}
	ok, err = verifyTaskVersion(t.shard, t.logger, task.DomainID, ai.Version, task.Version, task)
	if err != nil {
		return err
	} else if !ok {
		return nil
	}

	timeout := common.MinInt32(ai.ScheduleToStartTimeout, common.MaxTaskTimeout)
	// release the context lock since we no longer need mutable state builder and
	// the rest of logic is making RPC call, which takes time.
	release(nil)
	return t.pushActivity(task, timeout)
}

func (t *transferQueueActiveProcessorImpl) processDecisionTask(
	task *persistence.TransferTaskInfo,
) (retError error) {

	var err error
	execution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(task.WorkflowID),
		RunId:      common.StringPtr(task.RunID),
	}
	tasklist := &workflow.TaskList{
		Name: &task.TaskList,
	}

	// get workflow timeout
	context, release, err := t.cache.getOrCreateWorkflowExecutionForBackground(task.DomainID, execution)
	if err != nil {
		return err
	}
	defer func() { release(retError) }()

	var mutableState mutableState
	mutableState, err = loadMutableStateForTransferTask(context, task, t.metricsClient, t.logger)
	if err != nil {
		return err
	} else if mutableState == nil || !mutableState.IsWorkflowExecutionRunning() {
		return nil
	}

	decision, found := mutableState.GetDecisionInfo(task.ScheduleID)
	if !found {
		t.logger.Debug("Potentially duplicate task.", tag.TaskID(task.TaskID), tag.WorkflowScheduleID(task.ScheduleID), tag.TaskType(persistence.TransferTaskTypeDecisionTask))
		return nil
	}
	ok, err := verifyTaskVersion(t.shard, t.logger, task.DomainID, decision.Version, task.Version, task)
	if err != nil {
		return err
	} else if !ok {
		return nil
	}

	executionInfo := mutableState.GetExecutionInfo()
	workflowTimeout := executionInfo.WorkflowTimeout
	decisionTimeout := common.MinInt32(workflowTimeout, common.MaxTaskTimeout)

	// NOTE: previously this section check whether mutable state has enabled
	// sticky decision, if so convert the decision to a sticky decision.
	// that logic has a bug which timer task for that sticky decision is not generated
	// the correct logic should check whether the decision task is a sticky decision
	// task or not.
	if mutableState.GetExecutionInfo().TaskList != task.TaskList {
		// this decision is an sticky decision
		// there shall already be an timer set
		tasklist.Kind = common.TaskListKindPtr(workflow.TaskListKindSticky)
		decisionTimeout = executionInfo.StickyScheduleToStartTimeout
	}

	// release the context lock since we no longer need mutable state builder and
	// the rest of logic is making RPC call, which takes time.
	release(nil)
	return t.pushDecision(task, tasklist, decisionTimeout)
}

func (t *transferQueueActiveProcessorImpl) processCloseExecution(
	task *persistence.TransferTaskInfo,
) (retError error) {

	var err error
	domainID := task.DomainID
	execution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(task.WorkflowID),
		RunId:      common.StringPtr(task.RunID),
	}

	context, release, err := t.cache.getOrCreateWorkflowExecutionForBackground(domainID, execution)
	if err != nil {
		return err
	}
	defer func() { release(retError) }()

	var mutableState mutableState
	mutableState, err = loadMutableStateForTransferTask(context, task, t.metricsClient, t.logger)
	if err != nil {
		return err
	} else if mutableState == nil || mutableState.IsWorkflowExecutionRunning() {
		// this can happen if workflow is reset.
		return nil
	}

	lastWriteVersion, err := mutableState.GetLastWriteVersion()
	if err != nil {
		return err
	}
	ok, err := verifyTaskVersion(t.shard, t.logger, domainID, lastWriteVersion, task.Version, task)
	if err != nil {
		return err
	} else if !ok {
		return nil
	}

	executionInfo := mutableState.GetExecutionInfo()
	replyToParentWorkflow := mutableState.HasParentExecution() && executionInfo.CloseStatus != persistence.WorkflowCloseStatusContinuedAsNew
	completionEvent, err := mutableState.GetCompletionEvent()
	if err != nil {
		return err
	}
	wfCloseTime := completionEvent.GetTimestamp()

	parentDomainID := executionInfo.ParentDomainID
	parentWorkflowID := executionInfo.ParentWorkflowID
	parentRunID := executionInfo.ParentRunID
	initiatedID := executionInfo.InitiatedID

	workflowTypeName := executionInfo.WorkflowTypeName
	workflowStartTimestamp := executionInfo.StartTimestamp.UnixNano()
	workflowCloseTimestamp := wfCloseTime
	workflowCloseStatus := persistence.ToThriftWorkflowExecutionCloseStatus(executionInfo.CloseStatus)
	workflowHistoryLength := mutableState.GetNextEventID() - 1

	startEvent, err := mutableState.GetStartEvent()
	if err != nil {
		return err
	}
	workflowExecutionTimestamp := getWorkflowExecutionTimestamp(mutableState, startEvent)
	visibilityMemo := getWorkflowMemo(executionInfo.Memo)
	searchAttr := executionInfo.SearchAttributes
	domainName := mutableState.GetDomainEntry().GetInfo().Name
	children := mutableState.GetPendingChildExecutionInfos()

	// release the context lock since we no longer need mutable state builder and
	// the rest of logic is making RPC call, which takes time.
	release(nil)
	err = t.recordWorkflowClosed(
		domainID,
		execution,
		workflowTypeName,
		workflowStartTimestamp,
		workflowExecutionTimestamp.UnixNano(),
		workflowCloseTimestamp,
		workflowCloseStatus,
		workflowHistoryLength,
		task.GetTaskID(),
		visibilityMemo,
		searchAttr,
	)
	if err != nil {
		return err
	}

	// Communicate the result to parent execution if this is Child Workflow execution
	if replyToParentWorkflow {
		err = t.historyClient.RecordChildExecutionCompleted(nil, &h.RecordChildExecutionCompletedRequest{
			DomainUUID: common.StringPtr(parentDomainID),
			WorkflowExecution: &workflow.WorkflowExecution{
				WorkflowId: common.StringPtr(parentWorkflowID),
				RunId:      common.StringPtr(parentRunID),
			},
			InitiatedId: common.Int64Ptr(initiatedID),
			CompletedExecution: &workflow.WorkflowExecution{
				WorkflowId: common.StringPtr(task.WorkflowID),
				RunId:      common.StringPtr(task.RunID),
			},
			CompletionEvent: completionEvent,
		})

		// Check to see if the error is non-transient, in which case reset the error and continue with processing
		switch err.(type) {
		case *workflow.EntityNotExistsError:
			err = nil
		}
	}

	if err != nil {
		return err
	}

	if len(children) > 0 {
		err = t.processParentClosePolicy(domainName, domainID, children)
	}
	return err
}

func (t *transferQueueActiveProcessorImpl) processParentClosePolicy(domainName, domainUUID string, children map[int64]*persistence.ChildExecutionInfo) error {
	scope := t.metricsClient.Scope(metrics.TransferActiveTaskCloseExecutionScope)

	if t.shard.GetConfig().EnableParentClosePolicyWorker() && len(children) >= t.shard.GetConfig().ParentClosePolicyThreshold(domainName) {
		executions := make([]parentclosepolicy.RequestDetail, 0, len(children))
		for _, ch := range children {
			if ch.ParentClosePolicy == workflow.ParentClosePolicyAbandon {
				continue
			}

			executions = append(executions, parentclosepolicy.RequestDetail{
				WorkflowID: ch.StartedWorkflowID,
				RunID:      ch.StartedRunID,
				Policy:     ch.ParentClosePolicy,
			})
		}

		request := parentclosepolicy.Request{
			DomainName: domainName,
			DomainUUID: domainUUID,
			Executions: executions,
		}
		err := t.parentClosePolicyClient.SendParentClosePolicyRequest(request)
		if err != nil {
			return err
		}
	} else {
		for _, child := range children {
			var err error
			switch child.ParentClosePolicy {
			case workflow.ParentClosePolicyAbandon:
				//no-op
				continue
			case workflow.ParentClosePolicyTerminate:
				err = t.historyClient.TerminateWorkflowExecution(nil, &h.TerminateWorkflowExecutionRequest{
					DomainUUID: common.StringPtr(domainUUID),
					TerminateRequest: &workflow.TerminateWorkflowExecutionRequest{
						Domain: common.StringPtr(domainName),
						WorkflowExecution: &workflow.WorkflowExecution{
							WorkflowId: common.StringPtr(child.StartedWorkflowID),
							RunId:      common.StringPtr(child.StartedRunID),
						},
						Reason:   common.StringPtr("by parent close policy"),
						Identity: common.StringPtr(identityHistoryService),
					},
				})
			case workflow.ParentClosePolicyRequestCancel:
				err = t.historyClient.RequestCancelWorkflowExecution(nil, &h.RequestCancelWorkflowExecutionRequest{
					DomainUUID: common.StringPtr(domainUUID),
					CancelRequest: &workflow.RequestCancelWorkflowExecutionRequest{
						Domain: common.StringPtr(domainName),
						WorkflowExecution: &workflow.WorkflowExecution{
							WorkflowId: common.StringPtr(child.StartedWorkflowID),
							RunId:      common.StringPtr(child.StartedRunID),
						},
						Identity: common.StringPtr(identityHistoryService),
					},
				})
			}

			if err != nil {
				if _, ok := err.(*workflow.EntityNotExistsError); !ok {
					scope.IncCounter(metrics.ParentClosePolicyProcessorFailures)
					return err
				}
			}
			scope.IncCounter(metrics.ParentClosePolicyProcessorSuccess)
		}
	}
	return nil
}

func (t *transferQueueActiveProcessorImpl) processCancelExecution(
	task *persistence.TransferTaskInfo,
) (retError error) {

	var err error
	domainID := task.DomainID
	targetDomainID := task.TargetDomainID
	execution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(task.WorkflowID),
		RunId:      common.StringPtr(task.RunID),
	}

	var context workflowExecutionContext
	var release releaseWorkflowExecutionFunc
	context, release, err = t.cache.getOrCreateWorkflowExecutionForBackground(domainID, execution)
	if err != nil {
		return err
	}
	defer func() { release(retError) }()

	// First load the execution to validate if there is pending request cancellation for this transfer task
	var mutableState mutableState
	mutableState, err = loadMutableStateForTransferTask(context, task, t.metricsClient, t.logger)
	if err != nil {
		return err
	} else if mutableState == nil || !mutableState.IsWorkflowExecutionRunning() {
		return nil
	}

	initiatedEventID := task.ScheduleID
	ri, ok := mutableState.GetRequestCancelInfo(initiatedEventID)
	if !ok {
		return nil
	}
	ok, err = verifyTaskVersion(t.shard, t.logger, domainID, ri.Version, task.Version, task)
	if err != nil {
		return err
	} else if !ok {
		return nil
	}

	targetDomainEntry, err := t.domainCache.GetDomainByID(targetDomainID)
	if err != nil {
		return err
	}
	targetDomain := targetDomainEntry.GetInfo().Name

	// handle workflow cancel itself
	if domainID == targetDomainID && task.WorkflowID == task.TargetWorkflowID {
		// it does not matter if the run ID is a mismatch
		err = t.requestCancelExternalExecutionFailed(task, context, targetDomain, task.TargetWorkflowID, task.TargetRunID)
		if _, ok := err.(*workflow.EntityNotExistsError); ok {
			// this could happen if this is a duplicate processing of the task, and the execution has already completed.
			return nil
		}
		return err
	}

	cancelRequest := &h.RequestCancelWorkflowExecutionRequest{
		DomainUUID: common.StringPtr(targetDomainID),
		CancelRequest: &workflow.RequestCancelWorkflowExecutionRequest{
			Domain: common.StringPtr(targetDomain),
			WorkflowExecution: &workflow.WorkflowExecution{
				WorkflowId: common.StringPtr(task.TargetWorkflowID),
				RunId:      common.StringPtr(task.TargetRunID),
			},
			Identity: common.StringPtr(identityHistoryService),
			// Use the same request ID to dedupe RequestCancelWorkflowExecution calls
			RequestId: common.StringPtr(ri.CancelRequestID),
		},
		ExternalInitiatedEventId: common.Int64Ptr(task.ScheduleID),
		ExternalWorkflowExecution: &workflow.WorkflowExecution{
			WorkflowId: common.StringPtr(task.WorkflowID),
			RunId:      common.StringPtr(task.RunID),
		},
		ChildWorkflowOnly: common.BoolPtr(task.TargetChildWorkflowOnly),
	}

	if err = t.requestCancelExternalExecutionWithRetry(cancelRequest); err != nil {
		t.logger.Debug(fmt.Sprintf("Failed to cancel external workflow execution. Error: %v", err))

		// Check to see if the error is non-transient, in which case add RequestCancelFailed
		// event and complete transfer task by setting the err = nil
		if !common.IsServiceNonRetryableError(err) {
			// for retryable error just return
			return err
		}
		return t.requestCancelExternalExecutionFailed(
			task,
			context,
			targetDomain,
			task.TargetWorkflowID,
			task.TargetRunID,
		)
	}

	t.logger.Debug(fmt.Sprintf(
		"RequestCancel successfully recorded to external workflow execution.  WorkflowID: %v, RunID: %v",
		task.TargetWorkflowID,
		task.TargetRunID,
	))

	// Record ExternalWorkflowExecutionCancelRequested in source execution
	return t.requestCancelExternalExecutionCompleted(
		task,
		context,
		targetDomain,
		task.TargetWorkflowID,
		task.TargetRunID,
	)
}

func (t *transferQueueActiveProcessorImpl) processSignalExecution(
	task *persistence.TransferTaskInfo,
) (retError error) {

	var err error
	domainID := task.DomainID
	targetDomainID := task.TargetDomainID
	execution := workflow.WorkflowExecution{WorkflowId: common.StringPtr(task.WorkflowID),
		RunId: common.StringPtr(task.RunID)}

	var context workflowExecutionContext
	var release releaseWorkflowExecutionFunc
	context, release, err = t.cache.getOrCreateWorkflowExecutionForBackground(domainID, execution)
	if err != nil {
		return err
	}
	defer func() { release(retError) }()

	var mutableState mutableState
	mutableState, err = loadMutableStateForTransferTask(context, task, t.metricsClient, t.logger)
	if err != nil {
		return err
	} else if mutableState == nil || !mutableState.IsWorkflowExecutionRunning() {
		return nil
	}

	initiatedEventID := task.ScheduleID
	si, ok := mutableState.GetSignalInfo(initiatedEventID)
	if !ok {
		// TODO: here we should also RemoveSignalMutableState from target workflow
		// Otherwise, target SignalRequestID still can leak if shard restart after signalExternalExecutionCompleted
		// To do that, probably need to add the SignalRequestID in transfer task.
		return nil
	}
	ok, err = verifyTaskVersion(t.shard, t.logger, domainID, si.Version, task.Version, task)
	if err != nil {
		return err
	} else if !ok {
		return nil
	}

	targetDomainEntry, err := t.domainCache.GetDomainByID(targetDomainID)
	if err != nil {
		return err
	}
	targetDomain := targetDomainEntry.GetInfo().Name

	// handle workflow signal itself
	if domainID == targetDomainID && task.WorkflowID == task.TargetWorkflowID {
		// it does not matter if the run ID is a mismatch
		return t.signalExternalExecutionFailed(
			task,
			context,
			targetDomain,
			task.TargetWorkflowID,
			task.TargetRunID,
			si.Control,
		)
	}

	signalRequest := &h.SignalWorkflowExecutionRequest{
		DomainUUID: common.StringPtr(targetDomainID),
		SignalRequest: &workflow.SignalWorkflowExecutionRequest{
			Domain: common.StringPtr(targetDomain),
			WorkflowExecution: &workflow.WorkflowExecution{
				WorkflowId: common.StringPtr(task.TargetWorkflowID),
				RunId:      common.StringPtr(task.TargetRunID),
			},
			Identity:   common.StringPtr(identityHistoryService),
			SignalName: common.StringPtr(si.SignalName),
			Input:      si.Input,
			// Use same request ID to deduplicate SignalWorkflowExecution calls
			RequestId: common.StringPtr(si.SignalRequestID),
			Control:   si.Control,
		},
		ExternalWorkflowExecution: &workflow.WorkflowExecution{
			WorkflowId: common.StringPtr(task.WorkflowID),
			RunId:      common.StringPtr(task.RunID),
		},
		ChildWorkflowOnly: common.BoolPtr(task.TargetChildWorkflowOnly),
	}

	if err = t.signalExternalExecutionWithRetry(signalRequest); err != nil {
		t.logger.Debug(fmt.Sprintf("Failed to signal external workflow execution. Error: %v", err))

		// Check to see if the error is non-transient, in which case add SignalFailed
		// event and complete transfer task by setting the err = nil
		if !common.IsServiceNonRetryableError(err) {
			// for retryable error just return
			return err
		}
		return t.signalExternalExecutionFailed(
			task,
			context,
			targetDomain,
			task.TargetWorkflowID,
			task.TargetRunID,
			si.Control,
		)
	}

	t.logger.Debug(fmt.Sprintf(
		"Signal successfully recorded to external workflow execution.  WorkflowID: %v, RunID: %v",
		task.TargetWorkflowID,
		task.TargetRunID,
	))

	err = t.signalExternalExecutionCompleted(
		task,
		context,
		targetDomain,
		task.TargetWorkflowID,
		task.TargetRunID,
		si.Control,
	)
	if err != nil {
		return err
	}

	// release the context lock since we no longer need mutable state builder and
	// the rest of logic is making RPC call, which takes time.
	release(retError)
	// remove signalRequestedID from target workflow, after Signal detail is removed from source workflow
	removeRequest := &h.RemoveSignalMutableStateRequest{
		DomainUUID: common.StringPtr(targetDomainID),
		WorkflowExecution: &workflow.WorkflowExecution{
			WorkflowId: common.StringPtr(task.TargetWorkflowID),
			RunId:      common.StringPtr(task.TargetRunID),
		},
		RequestId: common.StringPtr(si.SignalRequestID),
	}
	return t.historyClient.RemoveSignalMutableState(nil, removeRequest)
}

func (t *transferQueueActiveProcessorImpl) processStartChildExecution(
	task *persistence.TransferTaskInfo,
) (retError error) {

	var err error
	domainID := task.DomainID
	targetDomainID := task.TargetDomainID
	execution := workflow.WorkflowExecution{WorkflowId: common.StringPtr(task.WorkflowID),
		RunId: common.StringPtr(task.RunID)}

	var context workflowExecutionContext
	var release releaseWorkflowExecutionFunc
	context, release, err = t.cache.getOrCreateWorkflowExecutionForBackground(domainID, execution)
	if err != nil {
		return err
	}
	defer func() { release(retError) }()

	// First step is to load workflow execution so we can retrieve the initiated event
	var mutableState mutableState
	mutableState, err = loadMutableStateForTransferTask(context, task, t.metricsClient, t.logger)
	if err != nil {
		return err
	} else if mutableState == nil || !mutableState.IsWorkflowExecutionRunning() {
		return nil
	}

	// Get parent domain name
	var domain string
	if domainEntry, err := t.shard.GetDomainCache().GetDomainByID(domainID); err != nil {
		if _, ok := err.(*workflow.EntityNotExistsError); !ok {
			return err
		}
		// it is possible that the domain got deleted. Use domainID instead as this is only needed for the history event
		domain = domainID
	} else {
		domain = domainEntry.GetInfo().Name
	}

	// Get target domain name
	var targetDomain string
	if domainEntry, err := t.shard.GetDomainCache().GetDomainByID(targetDomainID); err != nil {
		if _, ok := err.(*workflow.EntityNotExistsError); !ok {
			return err
		}
		// it is possible that the domain got deleted. Use domainID instead as this is only needed for the history event
		targetDomain = targetDomainID
	} else {
		targetDomain = domainEntry.GetInfo().Name
	}

	initiatedEventID := task.ScheduleID
	ci, ok := mutableState.GetChildExecutionInfo(initiatedEventID)
	if !ok {
		return nil
	}
	ok, err = verifyTaskVersion(t.shard, t.logger, domainID, ci.Version, task.Version, task)
	if err != nil {
		return err
	} else if !ok {
		return nil
	}

	initiatedEvent, err := mutableState.GetChildExecutionInitiatedEvent(initiatedEventID)
	if err != nil {
		return err
	}

	// ChildExecution already started, just create DecisionTask and complete transfer task
	if ci.StartedID != common.EmptyEventID {
		childExecution := &workflow.WorkflowExecution{
			WorkflowId: common.StringPtr(ci.StartedWorkflowID),
			RunId:      common.StringPtr(ci.StartedRunID),
		}
		return t.createFirstDecisionTask(targetDomainID, childExecution)
	}

	attributes := initiatedEvent.StartChildWorkflowExecutionInitiatedEventAttributes
	now := t.timeSource.Now()
	// Found pending child execution and it is not marked as started
	// Let's try and start the child execution
	startRequest := &h.StartWorkflowExecutionRequest{
		DomainUUID: common.StringPtr(targetDomainID),
		StartRequest: &workflow.StartWorkflowExecutionRequest{
			Domain:                              common.StringPtr(targetDomain),
			WorkflowId:                          attributes.WorkflowId,
			WorkflowType:                        attributes.WorkflowType,
			TaskList:                            attributes.TaskList,
			Input:                               attributes.Input,
			Header:                              attributes.Header,
			ExecutionStartToCloseTimeoutSeconds: attributes.ExecutionStartToCloseTimeoutSeconds,
			TaskStartToCloseTimeoutSeconds:      attributes.TaskStartToCloseTimeoutSeconds,
			// Use the same request ID to dedupe StartWorkflowExecution calls
			RequestId:             common.StringPtr(ci.CreateRequestID),
			WorkflowIdReusePolicy: attributes.WorkflowIdReusePolicy,
			RetryPolicy:           attributes.RetryPolicy,
			CronSchedule:          attributes.CronSchedule,
			Memo:                  attributes.Memo,
			SearchAttributes:      attributes.SearchAttributes,
		},
		ParentExecutionInfo: &h.ParentExecutionInfo{
			DomainUUID: common.StringPtr(domainID),
			Domain:     common.StringPtr(domain),
			Execution: &workflow.WorkflowExecution{
				WorkflowId: common.StringPtr(task.WorkflowID),
				RunId:      common.StringPtr(task.RunID),
			},
			InitiatedId: common.Int64Ptr(initiatedEventID),
		},
		FirstDecisionTaskBackoffSeconds: common.Int32Ptr(
			backoff.GetBackoffForNextScheduleInSeconds(
				attributes.GetCronSchedule(),
				now,
				now,
			),
		),
	}

	var startResponse *workflow.StartWorkflowExecutionResponse
	startResponse, err = t.historyClient.StartWorkflowExecution(nil, startRequest)
	if err != nil {
		t.logger.Debug(fmt.Sprintf("Failed to start child workflow execution. Error: %v", err))

		// Check to see if the error is non-transient, in which case add StartChildWorkflowExecutionFailed
		// event and complete transfer task by setting the err = nil
		switch err.(type) {
		case *workflow.WorkflowExecutionAlreadyStartedError:
			err = t.recordStartChildExecutionFailed(task, context, attributes)
		}
		return err
	}

	t.logger.Debug(fmt.Sprintf("Child Execution started successfully.  WorkflowID: %v, RunID: %v",
		*attributes.WorkflowId, *startResponse.RunId))

	// Child execution is successfully started, record ChildExecutionStartedEvent in parent execution
	err = t.recordChildExecutionStarted(task, context, attributes, *startResponse.RunId)

	if err != nil {
		return err
	}
	// Finally create first decision task for Child execution so it is really started
	return t.createFirstDecisionTask(targetDomainID, &workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(task.TargetWorkflowID),
		RunId:      common.StringPtr(*startResponse.RunId),
	})
}

func (t *transferQueueActiveProcessorImpl) processRecordWorkflowStarted(
	task *persistence.TransferTaskInfo,
) (retError error) {

	return t.processRecordWorkflowStartedOrUpsertHelper(task, true)
}

func (t *transferQueueActiveProcessorImpl) processUpsertWorkflowSearchAttributes(
	task *persistence.TransferTaskInfo,
) (retError error) {

	return t.processRecordWorkflowStartedOrUpsertHelper(task, false)
}

func (t *transferQueueActiveProcessorImpl) processRecordWorkflowStartedOrUpsertHelper(
	task *persistence.TransferTaskInfo,
	isRecordStart bool,
) (retError error) {

	var err error
	execution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(task.WorkflowID),
		RunId:      common.StringPtr(task.RunID),
	}

	// get workflow timeout
	context, release, err := t.cache.getOrCreateWorkflowExecutionForBackground(task.DomainID, execution)
	if err != nil {
		return err
	}
	defer func() { release(retError) }()

	var mutableState mutableState
	mutableState, err = loadMutableStateForTransferTask(context, task, t.metricsClient, t.logger)
	if err != nil {
		return err
	} else if mutableState == nil || !mutableState.IsWorkflowExecutionRunning() {
		return nil
	}

	// verify task version for RecordWorkflowStarted.
	// upsert doesn't require verifyTask, because it is just a sync of mutableState.
	if isRecordStart {
		startVersion, err := mutableState.GetStartVersion()
		if err != nil {
			return err
		}
		ok, err := verifyTaskVersion(t.shard, t.logger, task.DomainID, startVersion, task.Version, task)
		if err != nil {
			return err
		} else if !ok {
			return nil
		}
	}

	executionInfo := mutableState.GetExecutionInfo()
	workflowTimeout := executionInfo.WorkflowTimeout
	wfTypeName := executionInfo.WorkflowTypeName
	startTimestamp := executionInfo.StartTimestamp.UnixNano()
	startEvent, err := mutableState.GetStartEvent()
	if err != nil {
		return err
	}
	executionTimestamp := getWorkflowExecutionTimestamp(mutableState, startEvent)
	visibilityMemo := getWorkflowMemo(executionInfo.Memo)
	searchAttr := copySearchAttributes(executionInfo.SearchAttributes)

	// release the context lock since we no longer need mutable state builder and
	// the rest of logic is making RPC call, which takes time.
	release(nil)

	if isRecordStart {
		return t.recordWorkflowStarted(task.DomainID, execution, wfTypeName, startTimestamp, executionTimestamp.UnixNano(),
			workflowTimeout, task.GetTaskID(), visibilityMemo, searchAttr)
	}
	return t.upsertWorkflowExecution(task.DomainID, execution, wfTypeName, startTimestamp, executionTimestamp.UnixNano(),
		workflowTimeout, task.GetTaskID(), visibilityMemo, searchAttr)
}

func copySearchAttributes(
	input map[string][]byte,
) map[string][]byte {

	if input == nil {
		return nil
	}

	result := make(map[string][]byte)
	for k, v := range input {
		val := make([]byte, len(v))
		copy(val, v)
		result[k] = val
	}
	return result
}

func (t *transferQueueActiveProcessorImpl) processResetWorkflow(
	task *persistence.TransferTaskInfo,
) (retError error) {

	var err error
	execution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(task.WorkflowID),
		RunId:      common.StringPtr(task.RunID),
	}

	logger := t.logger.WithTags(
		tag.WorkflowDomainID(task.DomainID),
		tag.WorkflowID(execution.GetWorkflowId()),
		tag.WorkflowRunID(execution.GetRunId()),
	)
	// get workflow timeout
	currContext, currRelease, err := t.cache.getOrCreateWorkflowExecutionForBackground(task.DomainID, execution)
	if err != nil {
		return err
	}
	defer func() { currRelease(retError) }()

	var currMutableState mutableState
	currMutableState, err = loadMutableStateForTransferTask(currContext, task, t.metricsClient, t.logger)
	if err != nil {
		return err
	} else if currMutableState == nil {
		logger.Warn("Auto-Reset is skipped, because current run is deleted.")
		return nil
	}
	if !currMutableState.IsWorkflowExecutionRunning() {
		// it means this this might not be current anymore, we need to check
		var resp *persistence.GetCurrentExecutionResponse
		resp, retError = t.executionManager.GetCurrentExecution(&persistence.GetCurrentExecutionRequest{
			DomainID:   task.DomainID,
			WorkflowID: task.WorkflowID,
		})
		if retError != nil {
			return
		}
		if resp.RunID != task.RunID {
			logger.Warn("Auto-Reset is skipped, because current run is stale.")
			return nil
		}
	}
	// TODO: current reset doesn't allow childWFs, in the future we will release this restriction
	if len(currMutableState.GetPendingChildExecutionInfos()) > 0 {
		logger.Warn("Auto-Reset is skipped, because current run has pending child executions.")
		return nil
	}

	currentStartVersion, err := currMutableState.GetStartVersion()
	if err != nil {
		return err
	}
	ok, err := verifyTaskVersion(t.shard, t.logger, task.DomainID, currentStartVersion, task.Version, task)
	if err != nil {
		return err
	} else if !ok {
		return nil
	}

	executionInfo := currMutableState.GetExecutionInfo()

	domainEntry, err := t.shard.GetDomainCache().GetDomainByID(executionInfo.DomainID)
	if err != nil {
		return err
	}
	logger = logger.WithTags(tag.WorkflowDomainName(domainEntry.GetInfo().Name))

	reason, resetPt := FindAutoResetPoint(t.timeSource, &domainEntry.GetConfig().BadBinaries, executionInfo.AutoResetPoints)
	if resetPt == nil {
		logger.Warn("Auto-Reset is skipped, because reset point is not found.")
		return nil
	}

	var baseExecution workflow.WorkflowExecution
	var baseContext workflowExecutionContext
	var baseMutableState mutableState
	if resetPt.GetRunId() == executionInfo.RunID {
		baseMutableState = currMutableState
		baseContext = currContext
		baseExecution = execution
	} else {
		baseExecution = workflow.WorkflowExecution{
			WorkflowId: common.StringPtr(task.WorkflowID),
			RunId:      common.StringPtr(resetPt.GetRunId()),
		}
		var baseRelease func(err error)
		baseContext, baseRelease, err = t.cache.getOrCreateWorkflowExecutionForBackground(task.DomainID, baseExecution)
		if err != nil {
			return err
		}
		defer func() { baseRelease(retError) }()
		baseMutableState, err = loadMutableStateForTransferTask(baseContext, task, t.metricsClient, t.logger)
		if err != nil {
			return err
		}
		// in case of base already deleted, we skip the reset task
		// TODO in the future we may allow reset with archival history
		if baseMutableState == nil {
			logger.Warn("Auto-Reset is skipped, because the base run has been deleted.", tag.WorkflowResetBaseRunID(resetPt.GetRunId()))
			return nil
		}
	}
	logger = logger.WithTags(
		tag.WorkflowResetBaseRunID(resetPt.GetRunId()),
		tag.WorkflowBinaryChecksum(resetPt.GetBinaryChecksum()),
		tag.WorkflowEventID(resetPt.GetFirstDecisionCompletedId()))

	resp, err := t.historyService.resetor.ResetWorkflowExecution(ctx.Background(), &workflow.ResetWorkflowExecutionRequest{
		Domain:                common.StringPtr(domainEntry.GetInfo().Name),
		WorkflowExecution:     &baseExecution,
		Reason:                common.StringPtr(fmt.Sprintf("auto-reset reason:%v, binaryChecksum:%v ", reason, resetPt.GetBinaryChecksum())),
		DecisionFinishEventId: common.Int64Ptr(resetPt.GetFirstDecisionCompletedId()),
		RequestId:             common.StringPtr(uuid.New()),
	}, baseContext, baseMutableState, currContext, currMutableState)
	if err != nil {
		if _, ok := err.(*workflow.BadRequestError); ok {
			// This means the reset point is corrupted and not retry able.
			// There must be a bug in our system that we must fix.(for example, history is not the same in active/passive)
			t.metricsClient.IncCounter(metrics.TransferQueueProcessorScope, metrics.AutoResetPointCorruptionCounter)
			logger.Error("Auto-Reset workflow failed and not retryable. The reset point is corrupted.", tag.Error(err))
			return nil
		}
		// log this error and retry
		logger.Error("Auto-Reset workflow failed", tag.Error(err))
		return err
	}
	logger.Info("Auto-Reset workflow finished", tag.WorkflowResetNewRunID(resp.GetRunId()))
	return nil
}

func (t *transferQueueActiveProcessorImpl) recordChildExecutionStarted(
	task *persistence.TransferTaskInfo,
	context workflowExecutionContext,
	initiatedAttributes *workflow.StartChildWorkflowExecutionInitiatedEventAttributes,
	runID string,
) error {

	return t.updateWorkflowExecution(context, true,
		func(mutableState mutableState) error {
			if !mutableState.IsWorkflowExecutionRunning() {
				return &workflow.EntityNotExistsError{Message: "Workflow execution already completed."}
			}

			domain := initiatedAttributes.Domain
			initiatedEventID := task.ScheduleID
			ci, ok := mutableState.GetChildExecutionInfo(initiatedEventID)
			if !ok || ci.StartedID != common.EmptyEventID {
				return &workflow.EntityNotExistsError{Message: "Pending child execution not found."}
			}

			_, err := mutableState.AddChildWorkflowExecutionStartedEvent(
				domain,
				&workflow.WorkflowExecution{
					WorkflowId: common.StringPtr(task.TargetWorkflowID),
					RunId:      common.StringPtr(runID),
				},
				initiatedAttributes.WorkflowType,
				initiatedEventID,
				initiatedAttributes.Header,
			)

			return err
		})
}

func (t *transferQueueActiveProcessorImpl) recordStartChildExecutionFailed(
	task *persistence.TransferTaskInfo,
	context workflowExecutionContext,
	initiatedAttributes *workflow.StartChildWorkflowExecutionInitiatedEventAttributes,
) error {

	return t.updateWorkflowExecution(context, true,
		func(mutableState mutableState) error {
			if !mutableState.IsWorkflowExecutionRunning() {
				return &workflow.EntityNotExistsError{Message: "Workflow execution already completed."}
			}

			initiatedEventID := task.ScheduleID
			ci, ok := mutableState.GetChildExecutionInfo(initiatedEventID)
			if !ok || ci.StartedID != common.EmptyEventID {
				return &workflow.EntityNotExistsError{Message: "Pending child execution not found."}
			}

			_, err := mutableState.AddStartChildWorkflowExecutionFailedEvent(initiatedEventID,
				workflow.ChildWorkflowExecutionFailedCauseWorkflowAlreadyRunning, initiatedAttributes)

			return err
		})
}

// createFirstDecisionTask is used by StartChildExecution transfer task to create the first decision task for
// child execution.
func (t *transferQueueActiveProcessorImpl) createFirstDecisionTask(
	domainID string,
	execution *workflow.WorkflowExecution,
) error {

	err := t.historyClient.ScheduleDecisionTask(nil, &h.ScheduleDecisionTaskRequest{
		DomainUUID:        common.StringPtr(domainID),
		WorkflowExecution: execution,
		IsFirstDecision:   common.BoolPtr(true),
	})

	if err != nil {
		if _, ok := err.(*workflow.EntityNotExistsError); ok {
			// Maybe child workflow execution already timedout or terminated
			// Safe to discard the error and complete this transfer task
			return nil
		}
	}

	return err
}

func (t *transferQueueActiveProcessorImpl) requestCancelExternalExecutionCompleted(
	task *persistence.TransferTaskInfo,
	context workflowExecutionContext,
	targetDomain string,
	targetWorkflowID string,
	targetRunID string,
) error {

	err := t.updateWorkflowExecution(context, true,
		func(mutableState mutableState) error {
			if !mutableState.IsWorkflowExecutionRunning() {
				return &workflow.EntityNotExistsError{Message: "Workflow execution already completed."}
			}

			initiatedEventID := task.ScheduleID
			_, ok := mutableState.GetRequestCancelInfo(initiatedEventID)
			if !ok {
				return ErrMissingRequestCancelInfo
			}

			_, err := mutableState.AddExternalWorkflowExecutionCancelRequested(
				initiatedEventID,
				targetDomain,
				targetWorkflowID,
				targetRunID,
			)
			return err
		})

	if _, ok := err.(*workflow.EntityNotExistsError); ok {
		// this could happen if this is a duplicate processing of the task,
		// or the execution has already completed.
		return nil
	}
	return err
}

func (t *transferQueueActiveProcessorImpl) signalExternalExecutionCompleted(
	task *persistence.TransferTaskInfo,
	context workflowExecutionContext,
	targetDomain string,
	targetWorkflowID string,
	targetRunID string,
	control []byte,
) error {

	err := t.updateWorkflowExecution(context, true,
		func(mutableState mutableState) error {
			if !mutableState.IsWorkflowExecutionRunning() {
				return &workflow.EntityNotExistsError{Message: "Workflow execution already completed."}
			}

			initiatedEventID := task.ScheduleID
			_, ok := mutableState.GetSignalInfo(initiatedEventID)
			if !ok {
				return ErrMissingSignalInfo
			}

			_, err := mutableState.AddExternalWorkflowExecutionSignaled(
				initiatedEventID,
				targetDomain,
				targetWorkflowID,
				targetRunID,
				control,
			)
			return err
		})

	if _, ok := err.(*workflow.EntityNotExistsError); ok {
		// this could happen if this is a duplicate processing of the task,
		// or the execution has already completed.
		return nil
	}
	return err
}

func (t *transferQueueActiveProcessorImpl) requestCancelExternalExecutionFailed(
	task *persistence.TransferTaskInfo,
	context workflowExecutionContext,
	targetDomain string,
	targetWorkflowID string,
	targetRunID string,
) error {

	err := t.updateWorkflowExecution(context, true,
		func(mutableState mutableState) error {
			if !mutableState.IsWorkflowExecutionRunning() {
				return &workflow.EntityNotExistsError{Message: "Workflow execution already completed."}
			}

			initiatedEventID := task.ScheduleID
			_, ok := mutableState.GetRequestCancelInfo(initiatedEventID)
			if !ok {
				return ErrMissingRequestCancelInfo
			}

			_, err := mutableState.AddRequestCancelExternalWorkflowExecutionFailedEvent(
				common.EmptyEventID,
				initiatedEventID,
				targetDomain,
				targetWorkflowID,
				targetRunID,
				workflow.CancelExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution,
			)
			return err
		})

	if _, ok := err.(*workflow.EntityNotExistsError); ok {
		// this could happen if this is a duplicate processing of the task,
		// or the execution has already completed.
		return nil
	}
	return err
}

func (t *transferQueueActiveProcessorImpl) signalExternalExecutionFailed(
	task *persistence.TransferTaskInfo,
	context workflowExecutionContext,
	targetDomain string,
	targetWorkflowID string,
	targetRunID string,
	control []byte,
) error {

	err := t.updateWorkflowExecution(context, true,
		func(mutableState mutableState) error {
			if !mutableState.IsWorkflowExecutionRunning() {
				return &workflow.EntityNotExistsError{Message: "Workflow is not running."}
			}

			initiatedEventID := task.ScheduleID
			_, ok := mutableState.GetSignalInfo(initiatedEventID)
			if !ok {
				return ErrMissingSignalInfo
			}

			_, err := mutableState.AddSignalExternalWorkflowExecutionFailedEvent(
				common.EmptyEventID,
				initiatedEventID,
				targetDomain,
				targetWorkflowID,
				targetRunID,
				control,
				workflow.SignalExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution,
			)
			return err
		})

	if _, ok := err.(*workflow.EntityNotExistsError); ok {
		// this could happen if this is a duplicate processing of the task,
		// or the execution has already completed.
		return nil
	}
	return err
}

func (t *transferQueueActiveProcessorImpl) updateWorkflowExecution(
	context workflowExecutionContext,
	createDecisionTask bool,
	action func(builder mutableState) error,
) error {

	mutableState, err := context.loadWorkflowExecution()
	if err != nil {
		return err
	}

	if err := action(mutableState); err != nil {
		return err
	}

	if createDecisionTask {
		// Create a transfer task to schedule a decision task
		err := scheduleDecision(mutableState)
		if err != nil {
			return err
		}
	}

	return context.updateWorkflowExecutionAsActive(t.shard.GetTimeSource().Now())
}

func (t *transferQueueActiveProcessorImpl) requestCancelExternalExecutionWithRetry(
	cancelRequest *h.RequestCancelWorkflowExecutionRequest,
) error {
	op := func() error {
		return t.historyClient.RequestCancelWorkflowExecution(nil, cancelRequest)
	}

	err := backoff.Retry(op, persistenceOperationRetryPolicy, common.IsPersistenceTransientError)

	if _, ok := err.(*workflow.CancellationAlreadyRequestedError); ok {
		// err is CancellationAlreadyRequestedError
		// this could happen if target workflow cancellation is already requested
		// mark is as success
		return nil
	}
	return err
}

func (t *transferQueueActiveProcessorImpl) signalExternalExecutionWithRetry(
	signalRequest *h.SignalWorkflowExecutionRequest,
) error {

	op := func() error {
		return t.historyClient.SignalWorkflowExecution(nil, signalRequest)
	}

	return backoff.Retry(op, persistenceOperationRetryPolicy, common.IsPersistenceTransientError)
}
