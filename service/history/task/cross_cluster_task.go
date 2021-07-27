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
	"fmt"
	"sync"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/future"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	ctask "github.com/uber/cadence/common/task"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/shard"
)

// cross cluster task state
const (
	processingStateInitialized processingState = iota + 1
	processingStateResponseReported
	processingStateResponseRecorded
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

		// TODO: add more fields here for processing tasks
		// in the source cluster
	}

	crossClusterTaskBase struct {
		Info

		sync.Mutex
		state           ctask.State
		processingState processingState
		priority        int
		attempt         int

		shard          shard.Context
		timeSource     clock.TimeSource
		submitTime     time.Time
		logger         log.Logger
		eventLogger    eventLogger
		scope          metrics.Scope
		taskExecutor   Executor
		resubmitTaskFn func(task Task)
		maxRetryCount  int
	}
)

// NewCrossClusterSourceTask creates a cross cluster task
// for the processing it at the source cluster
func NewCrossClusterSourceTask(
	shard shard.Context,
	taskInfo Info,
	taskExecutor Executor,
	logger log.Logger,
	resubmitTaskFn func(task Task),
	maxRetryCount dynamicconfig.IntPropertyFn,
) Task {
	return &crossClusterSourceTask{
		crossClusterTaskBase: newCrossClusterTaskBase(
			shard,
			taskInfo,
			processingStateInitialized,
			taskExecutor,
			logger,
			shard.GetTimeSource(),
			resubmitTaskFn,
			maxRetryCount,
		),
	}
}

// NewCrossClusterTargetTask is called at the target cluster
// to process the cross cluster task
// the returned the Future will be unblocked when after the task
// is processed. The future value has type types.CrossClusterTaskResponse
// and there will be not error returned for this future. All errors will
// be recorded by the FailedCause field in the response.
func NewCrossClusterTargetTask(
	shard shard.Context,
	taskRequest *types.CrossClusterTaskRequest,
	taskExecutor Executor,
	logger log.Logger,
	resubmitTaskFn func(task Task),
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
			logger,
			shard.GetTimeSource(),
			resubmitTaskFn,
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
	logger log.Logger,
	timeSource clock.TimeSource,
	resubmitTaskFn func(task Task),
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
			GetCrossClusterTaskMetricsScope(taskInfo.GetTaskType()), taskInfo.GetDomainID(),
			logger,
		),
		resubmitTaskFn: resubmitTaskFn,
		maxRetryCount:  maxRetryCount(),
	}
}

// cross cluster source task methods

func (t *crossClusterSourceTask) Execute() error {
	panic("Not implement")
}

func (t *crossClusterSourceTask) Ack() {
	panic("Not implement")
}

func (t *crossClusterSourceTask) Nack() {
	panic("Not implement")
}

func (t *crossClusterSourceTask) HandleErr(
	err error,
) (retErr error) {
	panic("Not implement")
}

func (t *crossClusterSourceTask) RetryErr(
	err error,
) bool {
	panic("Not implement")
}

func (t *crossClusterSourceTask) IsReadyForPoll() bool {
	t.Lock()
	defer t.Unlock()

	return t.state == ctask.TaskStatePending &&
		(t.processingState == processingStateInitialized || t.processingState == processingStateResponseRecorded)
}

func (t *crossClusterSourceTask) GetCrossClusterRequest() *types.CrossClusterTaskRequest {
	panic("Not implement")
}

// cross cluster target task methods

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
		t.resubmitTaskFn(t)
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
	if err == ErrTaskPendingActive {
		return false
	}

	return true
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
