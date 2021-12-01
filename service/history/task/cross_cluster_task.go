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
	"fmt"
	"sync"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/future"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	ctask "github.com/uber/cadence/common/task"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
)

// cross cluster task state
// a typically state transition will be:
// Initialized -> Reported -> Recorded -> Reported
// if task has two stages (e.g. first start childworkflow and
// then schedule first decision task)
// or
// Initialized -> Reported
// if task has only one stage (e.g. cancel external workflow)
//
// NOTE: DO NOT change the value for each state
// new states must be added at the end to ensure backward compatibility
// across clusters.
const (
	// processingStateInitialized is the initial state after a task is loaded
	// task is available for poll at this state
	processingStateInitialized processingState = 1
	// processingStateResponseRecorded is the state when a response
	// for processing the task at target cluster is received
	// task is NOT available for polling at this state.
	// task can reach this state from either processingStateInitialized
	// or processingStateResponseRecorded
	processingStateResponseReported processingState = 2
	// processingStateResponseRecorded is the state after target response
	// is recorded in source workflow's history
	// task is available for poll at this state
	processingStateResponseRecorded processingState = 3
	// processingStateInvalidated is the state for a cross-cluster task
	// that is no longer valid and should be either converted to a transfer
	// task or move to the cross-cluster queue for another target cluster.
	// task is NOT available for poll at this state
	processingStateInvalidated processingState = 4
)

var (
	_ CrossClusterTask = (*crossClusterSourceTask)(nil)
)

type (
	// processingState is a more detailed state description
	// when the tasks is being processed and state is TaskStatePending
	processingState int

	crossClusterTargetTask struct {
		*crossClusterTaskBase

		request  *types.CrossClusterTaskRequest
		response *types.CrossClusterTaskResponse
		settable future.Settable
	}

	crossClusterSourceTask struct {
		*crossClusterTaskBase

		// targetCluster is the cluster name for which the task's owning
		// queue processor is responsible for
		// e.g if targetCluster is C, it means the task is loaded by
		// the queue processor responsible for cluster C, and only cluster
		// C is able to poll the task.
		// targetCluster may not match the current active cluster of the
		// targetDomain if the targetDomain performed a failover after
		// the task is loaded.
		targetCluster  string
		executionCache *execution.Cache
		response       *types.CrossClusterTaskResponse
		readyForPollFn func(task CrossClusterTask)
	}

	crossClusterTaskBase struct {
		Info

		sync.Mutex
		// state is used by the general processing queue implementation (the ack manager)
		// to determine if the execution has finished for the task
		state ctask.State
		// processingState is used by cross cluster task's specific task executor
		// to determine should be done next
		// the value of processingState only matters when state = TaskStatePending
		processingState processingState
		priority        int
		attempt         int

		shard         shard.Context
		timeSource    clock.TimeSource
		submitTime    time.Time
		logger        log.Logger
		eventLogger   eventLogger
		scope         metrics.Scope
		taskExecutor  Executor
		taskProcessor Processor
		redispatchFn  func(task Task)
		maxRetryCount int
	}
)

// NewCrossClusterSourceTask creates a cross cluster task
// for the processing it at the source cluster
func NewCrossClusterSourceTask(
	shard shard.Context,
	targetCluster string,
	executionCache *execution.Cache,
	taskInfo Info,
	taskExecutor Executor,
	taskProcessor Processor,
	logger log.Logger,
	redispatchFn func(task Task),
	readyForPollFn func(task CrossClusterTask),
	maxRetryCount dynamicconfig.IntPropertyFn,
) CrossClusterTask {
	return &crossClusterSourceTask{
		targetCluster:  targetCluster,
		executionCache: executionCache,
		readyForPollFn: readyForPollFn,
		crossClusterTaskBase: newCrossClusterTaskBase(
			shard,
			taskInfo,
			processingStateInitialized,
			taskExecutor,
			taskProcessor,
			getCrossClusterTaskMetricsScope(taskInfo.GetTaskType(), true),
			logger,
			shard.GetTimeSource(),
			redispatchFn,
			maxRetryCount,
		),
	}
}

// NewCrossClusterTargetTask is called at the target cluster
// to process the cross cluster task
// the returned the Future will be unblocked after the task
// is processed. The future value has type types.CrossClusterTaskResponse
// and there won't be any error returned for this future. All errors will
// be recorded by the FailedCause field in the response.
func NewCrossClusterTargetTask(
	shard shard.Context,
	taskRequest *types.CrossClusterTaskRequest,
	taskExecutor Executor,
	taskProcessor Processor,
	logger log.Logger,
	redispatchFn func(task Task),
	maxRetryCount dynamicconfig.IntPropertyFn,
) (Task, future.Future) {
	info := &persistence.CrossClusterTaskInfo{
		DomainID:            taskRequest.TaskInfo.DomainID,
		WorkflowID:          taskRequest.TaskInfo.WorkflowID,
		RunID:               taskRequest.TaskInfo.RunID,
		VisibilityTimestamp: time.Unix(0, taskRequest.TaskInfo.GetVisibilityTimestamp()),
		TaskID:              taskRequest.TaskInfo.TaskID,
		Version:             common.EmptyVersion, // we don't need version information at target cluster
	}
	switch taskRequest.TaskInfo.GetTaskType() {
	case types.CrossClusterTaskTypeStartChildExecution:
		info.TaskType = persistence.CrossClusterTaskTypeStartChildExecution
		info.TargetDomainID = taskRequest.StartChildExecutionAttributes.TargetDomainID
		info.TargetWorkflowID = taskRequest.StartChildExecutionAttributes.InitiatedEventAttributes.WorkflowID
		info.TargetRunID = taskRequest.StartChildExecutionAttributes.GetTargetRunID()
		info.ScheduleID = taskRequest.StartChildExecutionAttributes.InitiatedEventID
	case types.CrossClusterTaskTypeCancelExecution:
		info.TaskType = persistence.CrossClusterTaskTypeCancelExecution
		info.TargetDomainID = taskRequest.CancelExecutionAttributes.TargetDomainID
		info.TargetWorkflowID = taskRequest.CancelExecutionAttributes.TargetWorkflowID
		info.TargetRunID = taskRequest.CancelExecutionAttributes.TargetRunID
		info.ScheduleID = taskRequest.CancelExecutionAttributes.InitiatedEventID
		info.TargetChildWorkflowOnly = taskRequest.CancelExecutionAttributes.ChildWorkflowOnly
	case types.CrossClusterTaskTypeSignalExecution:
		info.TaskType = persistence.CrossClusterTaskTypeSignalExecution
		info.TargetDomainID = taskRequest.SignalExecutionAttributes.TargetDomainID
		info.TargetWorkflowID = taskRequest.SignalExecutionAttributes.TargetWorkflowID
		info.TargetRunID = taskRequest.SignalExecutionAttributes.TargetRunID
		info.ScheduleID = taskRequest.SignalExecutionAttributes.InitiatedEventID
		info.TargetChildWorkflowOnly = taskRequest.SignalExecutionAttributes.ChildWorkflowOnly
	case types.CrossClusterTaskTypeRecordChildWorkflowExeuctionComplete:
		info.TaskType = persistence.CrossClusterTaskTypeRecordChildExeuctionCompleted
		info.TargetDomainID = taskRequest.RecordChildWorkflowExecutionCompleteAttributes.TargetDomainID
		info.TargetWorkflowID = taskRequest.RecordChildWorkflowExecutionCompleteAttributes.TargetWorkflowID
		info.TargetRunID = taskRequest.RecordChildWorkflowExecutionCompleteAttributes.TargetRunID
		info.ScheduleID = taskRequest.RecordChildWorkflowExecutionCompleteAttributes.InitiatedEventID
	case types.CrossClusterTaskTypeApplyParentPolicy:
		info.TaskType = persistence.CrossClusterTaskTypeApplyParentClosePolicy
	default:
		panic(fmt.Sprintf("unknown cross cluster task type: %v", taskRequest.TaskInfo.GetTaskType()))
	}

	future, settable := future.NewFuture()
	return &crossClusterTargetTask{
		crossClusterTaskBase: newCrossClusterTaskBase(
			shard,
			info,
			processingState(taskRequest.TaskInfo.TaskState),
			taskExecutor,
			taskProcessor,
			getCrossClusterTaskMetricsScope(info.GetTaskType(), false),
			logger,
			shard.GetTimeSource(),
			redispatchFn,
			maxRetryCount,
		),
		request:  taskRequest,
		settable: settable,
	}, future
}

func newCrossClusterTaskBase(
	shard shard.Context,
	taskInfo Info,
	processingState processingState,
	taskExecutor Executor,
	taskProcessor Processor,
	metricScopeIdx int,
	logger log.Logger,
	timeSource clock.TimeSource,
	redispatchFn func(task Task),
	maxRetryCount dynamicconfig.IntPropertyFn,
) *crossClusterTaskBase {
	var eventLogger eventLogger
	if shard.GetConfig().EnableDebugMode &&
		shard.GetConfig().EnableTaskInfoLogByDomainID(taskInfo.GetDomainID()) {
		eventLogger = newEventLogger(logger, timeSource, defaultTaskEventLoggerSize)
		eventLogger.AddEvent("Created task")
	}

	return &crossClusterTaskBase{
		Info:            taskInfo,
		shard:           shard,
		state:           ctask.TaskStatePending,
		processingState: processingState,
		priority:        ctask.NoPriority,
		attempt:         0,
		timeSource:      timeSource,
		submitTime:      timeSource.Now(),
		logger:          logger,
		eventLogger:     eventLogger,
		scope: getOrCreateDomainTaggedScope(
			shard,
			metricScopeIdx,
			taskInfo.GetDomainID(),
			logger,
		),
		taskExecutor:  taskExecutor,
		taskProcessor: taskProcessor,
		redispatchFn:  redispatchFn,
		maxRetryCount: maxRetryCount(),
	}
}

// cross cluster source task methods

func (t *crossClusterSourceTask) Execute() error {
	executionStartTime := t.timeSource.Now()

	defer func() {
		t.scope.IncCounter(metrics.TaskRequestsPerDomain)
		t.scope.RecordTimer(metrics.TaskProcessingLatencyPerDomain, time.Since(executionStartTime))
	}()

	logEvent(t.eventLogger, "Executing task")
	return t.taskExecutor.Execute(t, true)
}

func (t *crossClusterSourceTask) Ack() {
	// do not set t.state to Acked, which will prevent the task from being fetched again
	// state and processingState will be updated by the task executor
	t.Lock()
	state := t.state
	processingState := t.processingState
	t.Unlock()

	logEvent(t.eventLogger, "executed task", "processing state", processingState)

	if state == ctask.TaskStateAcked {
		t.scope.RecordTimer(metrics.TaskAttemptTimerPerDomain, time.Duration(t.attempt))
		t.scope.RecordTimer(metrics.TaskLatencyPerDomain, time.Since(t.submitTime))
		t.scope.RecordTimer(metrics.TaskQueueLatencyPerDomain, time.Since(t.GetVisibilityTimestamp()))

		if t.eventLogger != nil && t.attempt != 0 {
			// only dump events when the task has been retried
			t.eventLogger.FlushEvents("Task processing events")
		}
	} else {
		if !t.IsReadyForPoll() {
			panic(fmt.Sprintf("Unexpected task state in CrossClusterSourceTask Ack method, state: %v, processing state: %v", t.State(), t.ProcessingState()))
		}
		t.readyForPollFn(t)
	}
}

func (t *crossClusterSourceTask) Nack() {
	// Nack will be called when we are unable to move the next step for processing
	// basically add the task to redispatch queue, so it can be retried later.
	// wether the task state is setting to Nacked or not is not important, since
	// task will be retried forever
	logEvent(t.eventLogger, "Nacked task")

	if t.GetAttempt() < activeTaskResubmitMaxAttempts {
		if submitted, _ := t.taskProcessor.TrySubmit(t); submitted {
			return
		}
	}

	t.redispatchFn(t)
}

func (t *crossClusterSourceTask) HandleErr(
	err error,
) (retErr error) {
	defer func() {
		if retErr != nil {
			logEvent(t.eventLogger, "Failed to handle error", retErr)

			t.Lock()
			t.attempt++
			attempt := t.attempt
			t.Unlock()

			if attempt > t.maxRetryCount {
				t.scope.RecordTimer(metrics.TaskAttemptTimerPerDomain, time.Duration(attempt))
				t.logger.Error("Critical error processing task, retrying.",
					tag.Error(err), tag.OperationCritical, tag.TaskType(t.GetTaskType()))
			}
		}
	}()

	if err == nil {
		return nil
	}

	logEvent(t.eventLogger, "Handling task processing error", err)

	if _, ok := err.(*types.EntityNotExistsError); ok {
		return nil
	}
	if _, ok := err.(*types.WorkflowExecutionAlreadyCompletedError); ok {
		return nil
	}

	if err == errWorkflowBusy {
		t.scope.IncCounter(metrics.TaskWorkflowBusyPerDomain)
		return err
	}
	if err == ErrTaskPendingActive {
		t.scope.IncCounter(metrics.TaskPendingActiveCounterPerDomain)
		return err
	}
	// return domain not active error here so that the cross-cluster task can be
	// convert to a (passive) transfer task

	t.scope.IncCounter(metrics.TaskFailuresPerDomain)
	if _, ok := err.(*persistence.CurrentWorkflowConditionFailedError); ok {
		t.logger.Error("More than 2 workflow are running.", tag.Error(err), tag.LifeCycleProcessingFailed)
		return nil
	}
	t.logger.Error("Fail to process task", tag.Error(err), tag.LifeCycleProcessingFailed)
	return err
}

func (t *crossClusterSourceTask) RetryErr(
	err error,
) bool {
	return err != errWorkflowBusy && err != ErrTaskPendingActive && !common.IsContextTimeoutError(err)
}

func (t *crossClusterSourceTask) IsReadyForPoll() bool {
	t.Lock()
	defer t.Unlock()

	return t.state == ctask.TaskStatePending &&
		(t.processingState == processingStateInitialized ||
			t.processingState == processingStateResponseRecorded)
}

// GetCrossClusterRequest returns a CrossClusterTaskRequest and error if there's any
// If the returned error is not nil:
// - there might be error while loading the request, we can retry the function call
// - task may be invalidated, in which case caller should submit the task for processing
//   so that a new task can be created.
// If both returned error and request are nil:
// - there's nothing need to be done for the task, task already acked and is not available
//   for polling again
// If the returned request is not nil
// - the request can be returned to the target cluster
func (t *crossClusterSourceTask) GetCrossClusterRequest() (request *types.CrossClusterTaskRequest, retError error) {
	t.Lock()
	defer func() {
		if retError == nil && request == nil {
			t.state = ctask.TaskStateAcked
		}
		t.Unlock()
	}()

	if !t.isValidLocked() {
		return nil, errors.New("task invalidated")
	}

	ctx, cancel := context.WithTimeout(context.Background(), taskDefaultTimeout)
	defer cancel()

	taskInfo := t.GetInfo().(*persistence.CrossClusterTaskInfo)
	_, mutableState, release, err := loadWorkflowForCrossClusterTask(ctx, t.executionCache, taskInfo, t.shard.GetMetricsClient(), t.logger)
	if err != nil || mutableState == nil {
		return nil, err
	}
	defer func() { release(retError) }()

	request = &types.CrossClusterTaskRequest{
		TaskInfo: &types.CrossClusterTaskInfo{
			DomainID:            t.GetDomainID(),
			WorkflowID:          t.GetWorkflowID(),
			RunID:               t.GetRunID(),
			TaskID:              t.GetTaskID(),
			VisibilityTimestamp: common.Int64Ptr(t.GetVisibilityTimestamp().UnixNano()),
		},
	}

	var taskState processingState
	switch t.GetTaskType() {
	case persistence.CrossClusterTaskTypeStartChildExecution:
		var attributes *types.CrossClusterStartChildExecutionRequestAttributes
		attributes, taskState, err = t.getRequestForStartChildExecution(ctx, taskInfo, mutableState)
		if err != nil || attributes == nil {
			return nil, err
		}
		request.TaskInfo.TaskType = types.CrossClusterTaskTypeStartChildExecution.Ptr()
		request.StartChildExecutionAttributes = attributes
	case persistence.CrossClusterTaskTypeCancelExecution:
		var attributes *types.CrossClusterCancelExecutionRequestAttributes
		attributes, taskState, err = t.getRequestForCancelExecution(ctx, taskInfo, mutableState)
		if err != nil || attributes == nil {
			return nil, err
		}
		request.TaskInfo.TaskType = types.CrossClusterTaskTypeCancelExecution.Ptr()
		request.CancelExecutionAttributes = attributes
	case persistence.CrossClusterTaskTypeSignalExecution:
		var attributes *types.CrossClusterSignalExecutionRequestAttributes
		attributes, taskState, err = t.getRequestForSignalExecution(ctx, taskInfo, mutableState)
		if err != nil || attributes == nil {
			return nil, err
		}
		request.TaskInfo.TaskType = types.CrossClusterTaskTypeSignalExecution.Ptr()
		request.SignalExecutionAttributes = attributes
	case persistence.CrossClusterTaskTypeRecordChildExeuctionCompleted:
		var attributes *types.CrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes
		attributes, taskState, err = t.getRequestForRecordChildWorkflowCompletion(ctx, taskInfo, mutableState)
		if err != nil || attributes == nil {
			return nil, err
		}
		request.TaskInfo.TaskType = types.CrossClusterTaskTypeRecordChildWorkflowExeuctionComplete.Ptr()
		request.RecordChildWorkflowExecutionCompleteAttributes = attributes
	case persistence.CrossClusterTaskTypeApplyParentClosePolicy:
		var attributes *types.CrossClusterApplyParentClosePolicyRequestAttributes
		attributes, taskState, err = t.getRequestForApplyParentPolicy(ctx, taskInfo, mutableState)
		if err != nil || attributes == nil {
			return nil, err
		}
		request.TaskInfo.TaskType = types.CrossClusterTaskTypeApplyParentPolicy.Ptr()
		request.ApplyParentClosePolicyAttributes = attributes
	default:
		return nil, errUnknownCrossClusterTask
	}

	request.TaskInfo.TaskState = int16(taskState)
	return request, nil
}

func (t *crossClusterSourceTask) VerifyLastWriteVersion(
	mutableState execution.MutableState,
	taskInfo *persistence.CrossClusterTaskInfo,
) (bool, error) {
	lastWriteVersion, err := mutableState.GetLastWriteVersion()
	if err != nil {
		return false, err
	}
	return verifyTaskVersion(t.shard, t.logger, taskInfo.DomainID, lastWriteVersion, taskInfo.Version, taskInfo)
}

func (t *crossClusterSourceTask) getRequestForApplyParentPolicy(
	ctx context.Context,
	taskInfo *persistence.CrossClusterTaskInfo,
	mutableState execution.MutableState,
) (*types.CrossClusterApplyParentClosePolicyRequestAttributes, processingState, error) {
	if mutableState.IsWorkflowExecutionRunning() {
		return nil, t.processingState, nil
	}

	// No need to check the target failovers, only the active cluster should poll tasks
	// if active cluster changes during polling, target should return error to the source
	verified, err := t.VerifyLastWriteVersion(mutableState, taskInfo)
	if err != nil || !verified {
		return nil, t.processingState, err
	}

	domainEntry := mutableState.GetDomainEntry()
	attributes := &types.CrossClusterApplyParentClosePolicyRequestAttributes{}
	children, err := filterPendingChildExecutions(
		taskInfo.TargetDomainIDs,
		mutableState.GetPendingChildExecutionInfos(),
		t.GetShard().GetDomainCache(),
		domainEntry,
	)
	if err != nil {
		return nil, t.processingState, err
	}
	for _, childInfo := range children {
		// we already filtered the children so that child domainID is in task.TargetDomainIDs
		// don't check if child domain is active or not here,
		// we need to send the request even if the child domain is not active in target cluster
		targetDomainID, err := execution.GetChildExecutionDomainID(childInfo, t.shard.GetDomainCache(), domainEntry)
		if err != nil {
			return nil, t.processingState, err
		}

		attributes.Children = append(
			attributes.Children,
			&types.ApplyParentClosePolicyRequest{
				Child: &types.ApplyParentClosePolicyAttributes{
					ChildDomainID:     targetDomainID,
					ChildWorkflowID:   childInfo.StartedWorkflowID,
					ChildRunID:        childInfo.StartedRunID,
					ParentClosePolicy: &childInfo.ParentClosePolicy,
				},
				Status: &types.ApplyParentClosePolicyStatus{
					Completed: false,
				},
			},
		)
	}
	return attributes, t.processingState, nil
}

func (t *crossClusterSourceTask) getRequestForRecordChildWorkflowCompletion(
	ctx context.Context,
	taskInfo *persistence.CrossClusterTaskInfo,
	mutableState execution.MutableState,
) (*types.CrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes, processingState, error) {
	if mutableState.IsWorkflowExecutionRunning() {
		return nil, t.processingState, nil
	}

	verified, err := t.VerifyLastWriteVersion(mutableState, taskInfo)
	if err != nil || !verified {
		return nil, t.processingState, err
	}

	executionInfo := mutableState.GetExecutionInfo()
	completionEvent, err := mutableState.GetCompletionEvent(ctx)
	if err != nil {
		return nil, t.processingState, err
	}

	attributes := &types.CrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes{
		TargetDomainID:   executionInfo.ParentDomainID,
		TargetWorkflowID: executionInfo.ParentWorkflowID,
		TargetRunID:      executionInfo.ParentRunID,
		InitiatedEventID: executionInfo.InitiatedID,
		CompletionEvent:  completionEvent,
	}

	return attributes, t.processingState, nil
}

func (t *crossClusterSourceTask) getRequestForStartChildExecution(
	ctx context.Context,
	taskInfo *persistence.CrossClusterTaskInfo,
	mutableState execution.MutableState,
) (*types.CrossClusterStartChildExecutionRequestAttributes, processingState, error) {
	initiatedEventID := taskInfo.ScheduleID
	childInfo, ok := mutableState.GetChildExecutionInfo(initiatedEventID)
	if !ok {
		return nil, t.processingState, nil
	}
	ok, err := verifyTaskVersion(t.shard, t.logger, taskInfo.DomainID, childInfo.Version, taskInfo.Version, taskInfo)
	if err != nil || !ok {
		return nil, t.processingState, err
	}

	if !mutableState.IsWorkflowExecutionRunning() &&
		(childInfo.StartedID == common.EmptyEventID ||
			childInfo.ParentClosePolicy != types.ParentClosePolicyAbandon) {
		return nil, t.processingState, err
	}

	initiatedEvent, err := mutableState.GetChildExecutionInitiatedEvent(ctx, initiatedEventID)
	if err != nil {
		return nil, t.processingState, err
	}

	attributes := &types.CrossClusterStartChildExecutionRequestAttributes{
		TargetDomainID:           taskInfo.TargetDomainID,
		RequestID:                childInfo.CreateRequestID,
		InitiatedEventID:         initiatedEventID,
		InitiatedEventAttributes: initiatedEvent.StartChildWorkflowExecutionInitiatedEventAttributes,
	}
	if childInfo.StartedID != common.EmptyEventID {
		// childExecution already started, advance to next state
		t.processingState = processingStateResponseRecorded
		attributes.TargetRunID = common.StringPtr(childInfo.StartedRunID)
	}

	return attributes, t.processingState, nil
}

func (t *crossClusterSourceTask) getRequestForCancelExecution(
	ctx context.Context,
	taskInfo *persistence.CrossClusterTaskInfo,
	mutableState execution.MutableState,
) (*types.CrossClusterCancelExecutionRequestAttributes, processingState, error) {
	if !mutableState.IsWorkflowExecutionRunning() {
		return nil, t.processingState, nil
	}

	initiatedEventID := taskInfo.ScheduleID
	requestCancelInfo, ok := mutableState.GetRequestCancelInfo(initiatedEventID)
	if !ok {
		return nil, t.processingState, nil
	}
	ok, err := verifyTaskVersion(t.shard, t.logger, taskInfo.DomainID, requestCancelInfo.Version, taskInfo.Version, taskInfo)
	if err != nil || !ok {
		return nil, t.processingState, err
	}

	return &types.CrossClusterCancelExecutionRequestAttributes{
		TargetDomainID:    taskInfo.TargetDomainID,
		TargetWorkflowID:  taskInfo.TargetWorkflowID,
		TargetRunID:       taskInfo.TargetRunID,
		RequestID:         requestCancelInfo.CancelRequestID,
		InitiatedEventID:  initiatedEventID,
		ChildWorkflowOnly: taskInfo.TargetChildWorkflowOnly,
	}, t.processingState, nil
}

func (t *crossClusterSourceTask) getRequestForSignalExecution(
	ctx context.Context,
	taskInfo *persistence.CrossClusterTaskInfo,
	mutableState execution.MutableState,
) (*types.CrossClusterSignalExecutionRequestAttributes, processingState, error) {
	if !mutableState.IsWorkflowExecutionRunning() {
		return nil, t.processingState, nil
	}

	initiatedEventID := taskInfo.ScheduleID
	signalInfo, ok := mutableState.GetSignalInfo(initiatedEventID)
	if !ok {
		return nil, t.processingState, nil
	}
	ok, err := verifyTaskVersion(t.shard, t.logger, taskInfo.DomainID, signalInfo.Version, taskInfo.Version, taskInfo)
	if err != nil || !ok {
		return nil, t.processingState, err
	}

	return &types.CrossClusterSignalExecutionRequestAttributes{
		TargetDomainID:    taskInfo.TargetDomainID,
		TargetWorkflowID:  taskInfo.TargetWorkflowID,
		TargetRunID:       taskInfo.TargetRunID,
		RequestID:         signalInfo.SignalRequestID,
		InitiatedEventID:  initiatedEventID,
		ChildWorkflowOnly: taskInfo.TargetChildWorkflowOnly,
		SignalName:        signalInfo.SignalName,
		SignalInput:       signalInfo.Input,
		Control:           signalInfo.Control,
	}, t.processingState, nil
}

func (t *crossClusterSourceTask) IsValid() bool {
	t.Lock()
	defer t.Unlock()

	return t.isValidLocked()
}

func (t *crossClusterSourceTask) isValidLocked() bool {
	if t.processingState == processingStateInvalidated {
		return false
	}

	domainCache := t.shard.GetDomainCache()
	sourceEntry, err := domainCache.GetDomainByID(t.GetDomainID())
	if err != nil {
		// unable to tell, treat the task as valid
		t.logger.Error("Failed to load domain entry", tag.Error(err))
		return true
	}

	var targetEntry *cache.DomainCacheEntry
	// for apply parent policy, target workflow infomation is not
	// persisted with the task, so skip this test for target workflow since the check is best effort
	// TODO: we should check the TargetDomainIDs field
	if t.GetTaskType() != persistence.CrossClusterTaskTypeApplyParentClosePolicy {
		targetEntry, err = domainCache.GetDomainByID(t.Info.(*persistence.CrossClusterTaskInfo).TargetDomainID)
		if err != nil {
			return true
		}
	}

	// pending active state is treated as valid
	sourceInvalid := sourceEntry.GetReplicationConfig().ActiveClusterName !=
		t.shard.GetClusterMetadata().GetCurrentClusterName()
	targetInvalid := targetEntry != nil && targetEntry.GetReplicationConfig().ActiveClusterName != t.targetCluster

	if sourceInvalid || targetInvalid {
		t.processingState = processingStateInvalidated
		return false
	}

	return true
}

// RecordResponse records the response of processing cross cluster task at the target cluster
// If an error is returned, the operation is failed, task state will remain unchanged
// and task is still be available for polling
// If not error is returned, operation is successful, task won't be available for polling
// and caller should submit the task for processing.
func (t *crossClusterSourceTask) RecordResponse(response *types.CrossClusterTaskResponse) error {
	t.Lock()
	defer t.Unlock()

	if t.state != ctask.TaskStatePending || t.processingState != processingState(response.TaskState) {
		// this might happen when:
		// 1. duplicated response (shard movement in target cluster)
		// 2. shard movement in source cluster causing task to be re-processed
		// 3. task got invalidated during domain failover callback
		return fmt.Errorf("unexpected task state, expected: %v, actual: %v", t.processingState, response.TaskState)
	}

	if t.GetTaskID() != response.GetTaskID() {
		return fmt.Errorf("unexpected taskID, expected: %v, actual: %v", t.GetTaskID(), response.GetTaskID())
	}

	var emptyResponse bool
	var taskTypeMatch bool
	switch t.GetTaskType() {
	case persistence.CrossClusterTaskTypeStartChildExecution:
		taskTypeMatch = response.GetTaskType() == types.CrossClusterTaskTypeStartChildExecution
		emptyResponse = response.StartChildExecutionAttributes == nil
	case persistence.CrossClusterTaskTypeCancelExecution:
		taskTypeMatch = response.GetTaskType() == types.CrossClusterTaskTypeCancelExecution
		emptyResponse = response.CancelExecutionAttributes == nil
	case persistence.CrossClusterTaskTypeSignalExecution:
		taskTypeMatch = response.GetTaskType() == types.CrossClusterTaskTypeSignalExecution
		emptyResponse = response.SignalExecutionAttributes == nil
	case persistence.CrossClusterTaskTypeRecordChildExeuctionCompleted:
		taskTypeMatch = response.GetTaskType() == types.CrossClusterTaskTypeRecordChildWorkflowExeuctionComplete
		emptyResponse = response.RecordChildWorkflowExecutionCompleteAttributes == nil
	case persistence.CrossClusterTaskTypeApplyParentClosePolicy:
		taskTypeMatch = response.GetTaskType() == types.CrossClusterTaskTypeApplyParentPolicy
		emptyResponse = response.ApplyParentClosePolicyAttributes == nil
	default:
		return fmt.Errorf("unknown task type: %v", t.GetTaskType())
	}

	if !taskTypeMatch {
		return fmt.Errorf("unexpected task type, expected: %v, actual: %v", t.GetTaskType(), response.GetTaskType())
	}

	if emptyResponse && response.FailedCause == nil {
		return fmt.Errorf("empty cross cluster task response, task type: %v", t.GetTaskType())
	}

	if response.FailedCause != nil {
		unexpectedFailedCause := false
		switch response.GetFailedCause() {
		case types.CrossClusterTaskFailedCauseUncategorized:
			unexpectedFailedCause = true
		case types.CrossClusterTaskFailedCauseWorkflowAlreadyRunning:
			unexpectedFailedCause = (t.GetTaskType() == persistence.CrossClusterTaskTypeCancelExecution ||
				t.GetTaskType() == persistence.CrossClusterTaskTypeSignalExecution)
		}
		if unexpectedFailedCause {
			// nothing needs to be done for unexpectedFailedCause
			// leave the task state as is, so it's available for next poll
			// before it's fetched we will also verify if the task still need to be processed
			// if so task will be fetched and executed again in the target cluster
			return fmt.Errorf("unexpected cross cluster task failed cause: %v", response.GetFailedCause())
		}
	}

	// set state to reported so that the task won't be pulled
	// while it's being processed at the source cluster
	t.processingState = processingStateResponseReported
	t.response = response
	return nil
}

// CROSS CLUSTER TARGET TASK METHODS

func (t *crossClusterTargetTask) Execute() error {
	executionStartTime := t.timeSource.Now()

	defer func() {
		t.scope.IncCounter(metrics.TaskRequestsPerDomain)
		t.scope.RecordTimer(metrics.TaskProcessingLatencyPerDomain, time.Since(executionStartTime))
	}()

	logEvent(t.eventLogger, "Executing task")
	return t.taskExecutor.Execute(t, true)
}

func (t *crossClusterTargetTask) Ack() {
	t.Lock()
	t.state = ctask.TaskStateAcked
	t.Unlock()

	logEvent(t.eventLogger, "Acked task")
	t.completeTask()
}

func (t *crossClusterTargetTask) Nack() {
	t.Lock()
	if t.attempt < t.maxRetryCount {
		t.redispatchFn(t)
		t.Unlock()
		return
	}
	t.state = ctask.TaskStateNacked
	t.Unlock()

	logEvent(t.eventLogger, "Nacked task")
	t.completeTask()
}

func (t *crossClusterTargetTask) HandleErr(
	err error,
) (retErr error) {
	logEvent(t.eventLogger, "Failed to handle error", retErr)

	t.Lock()
	t.attempt++
	t.Unlock()

	t.scope.IncCounter(metrics.TaskFailuresPerDomain)
	t.logger.Error("Fail to process task", tag.Error(err), tag.LifeCycleProcessingFailed)
	return err
}

func (t *crossClusterTargetTask) RetryErr(
	err error,
) bool {
	return err != ErrTaskPendingActive
}

func (t *crossClusterTargetTask) completeTask() {
	t.scope.RecordTimer(metrics.TaskAttemptTimerPerDomain, time.Duration(t.attempt))
	t.scope.RecordTimer(metrics.TaskLatencyPerDomain, time.Since(t.submitTime))
	t.scope.RecordTimer(metrics.TaskQueueLatencyPerDomain, time.Since(t.GetVisibilityTimestamp()))

	if t.eventLogger != nil && t.attempt != 0 {
		// only dump events when the task has been retried
		t.eventLogger.FlushEvents("Task processing events")
	}

	t.settable.Set(*t.response, nil)
}

// cross cluster task base method shared by both source and target task impl

func (t *crossClusterTaskBase) GetInfo() Info {
	return t.Info
}

func (t *crossClusterTaskBase) State() ctask.State {
	t.Lock()
	defer t.Unlock()

	return t.state
}

func (t *crossClusterTaskBase) Priority() int {
	t.Lock()
	defer t.Unlock()

	return t.priority
}

func (t *crossClusterTaskBase) SetPriority(
	priority int,
) {
	t.Lock()
	defer t.Unlock()

	t.priority = priority
}

func (t *crossClusterTaskBase) GetShard() shard.Context {
	return t.shard
}

func (t *crossClusterTaskBase) GetAttempt() int {
	t.Lock()
	defer t.Unlock()

	return t.attempt
}

func (t *crossClusterTaskBase) GetQueueType() QueueType {
	return QueueTypeCrossCluster
}

func (t *crossClusterTaskBase) ProcessingState() processingState {
	t.Lock()
	defer t.Unlock()

	return t.processingState
}

func loadWorkflowForCrossClusterTask(
	ctx context.Context,
	executionCache *execution.Cache,
	taskInfo *persistence.CrossClusterTaskInfo,
	metricsClient metrics.Client,
	logger log.Logger,
) (execution.Context, execution.MutableState, execution.ReleaseFunc, error) {
	wfContext, release, err := executionCache.GetOrCreateWorkflowExecutionWithTimeout(
		taskInfo.GetDomainID(),
		getWorkflowExecution(taskInfo),
		taskGetExecutionContextTimeout,
	)
	if err != nil {
		if err == context.DeadlineExceeded {
			return nil, nil, nil, errWorkflowBusy
		}
		return nil, nil, nil, err
	}

	mutableState, err := loadMutableStateForCrossClusterTask(ctx, wfContext, taskInfo, metricsClient, logger)
	if err != nil {
		release(err)
		return nil, nil, nil, err
	}
	if mutableState == nil {
		release(nil)
		return nil, nil, nil, nil
	}

	return wfContext, mutableState, release, nil
}
