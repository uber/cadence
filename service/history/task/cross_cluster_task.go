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
	"sync"
	"time"

	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/future"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	ctask "github.com/uber/cadence/common/task"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/shard"
)

// cross cluster task state
const (
	processingStateInitialed processingState = iota + 1
	processingStateResponseReported
	processingStateResponseRecorded
)

type (
	processingState int

	crossClusterSignalWorkflowTask struct {
		*crossClusterTaskBase
	}

	crossClusterCancelWorkflowTask struct {
		*crossClusterTaskBase
	}

	crossClusterStartChildWorkflowTask struct {
		*crossClusterTaskBase
	}

	crossClusterTaskBase struct {
		sync.Mutex
		Info //TODO: provide cross cluster task info detail

		shard           shard.Context
		state           ctask.State
		processingState processingState
		priority        int
		attempt         int
		timeSource      clock.TimeSource
		submitTime      time.Time
		logger          log.Logger
		eventLogger     eventLogger
		scopeIdx        int
		scope           metrics.Scope
		taskExecutor    Executor        //TODO: wire up the dependency
		taskProcessor   Processor       //TODO: wire up the dependency
		redispatchFn    func(task Task) //TODO: wire up the dependency
		maxRetryCount   dynamicconfig.IntPropertyFn
		settable        future.Settable
	}
)

// NewCrossClusterTaskForTargetCluster creates a CrossClusterTask
// for the processing logic at target cluster
// the returned the Future will be unblocked when after the task
// is processed. The future value has type types.CrossClusterTaskResponse
// and there will be not error returned for this future. All errors will
// be recorded by the FailedCause field in the response.
func NewCrossClusterTaskForTargetCluster(
	shard shard.Context,
	taskRequest *types.CrossClusterTaskRequest,
	logger log.Logger,
	maxRetryCount dynamicconfig.IntPropertyFn,
) (CrossClusterTask, future.Future) {
	// TODO: create CrossClusterTasks based on request
	// for now create a dummy CrossClusterTask for testing
	future, settable := future.NewFuture()
	return &crossClusterSignalWorkflowTask{
		crossClusterTaskBase: &crossClusterTaskBase{
			Info: &persistence.CrossClusterTaskInfo{
				TaskID: taskRequest.TaskInfo.TaskID,
			},
			settable: settable,
		},
	}, future
}

// NewCrossClusterSignalWorkflowTask initialize cross cluster signal workflow task and task future
func NewCrossClusterSignalWorkflowTask(
	shard shard.Context,
	taskInfo Info,
	logger log.Logger,
	timeSource clock.TimeSource,
	maxRetryCount dynamicconfig.IntPropertyFn,
) CrossClusterTask {
	crossClusterTask := newCrossClusterTask(
		shard,
		taskInfo,
		logger,
		timeSource,
		maxRetryCount)
	return &crossClusterSignalWorkflowTask{
		crossClusterTask,
	}
}

// NewCrossClusterCancelWorkflowTask initialize cross cluster cancel workflow task and task future
func NewCrossClusterCancelWorkflowTask(
	shard shard.Context,
	taskInfo Info,
	logger log.Logger,
	timeSource clock.TimeSource,
	maxRetryCount dynamicconfig.IntPropertyFn,
) CrossClusterTask {
	crossClusterTask := newCrossClusterTask(
		shard,
		taskInfo,
		logger,
		timeSource,
		maxRetryCount)
	return &crossClusterCancelWorkflowTask{
		crossClusterTask,
	}
}

// NewCrossClusterStartChildWorkflowTask initialize cross cluster start child workflow task and task future
func NewCrossClusterStartChildWorkflowTask(
	shard shard.Context,
	taskInfo Info,
	logger log.Logger,
	timeSource clock.TimeSource,
	maxRetryCount dynamicconfig.IntPropertyFn,
) CrossClusterTask {
	crossClusterTask := newCrossClusterTask(
		shard,
		taskInfo,
		logger,
		timeSource,
		maxRetryCount)
	return &crossClusterStartChildWorkflowTask{
		crossClusterTask,
	}
}

func newCrossClusterTask(
	shard shard.Context,
	taskInfo Info,
	logger log.Logger,
	timeSource clock.TimeSource,
	maxRetryCount dynamicconfig.IntPropertyFn,
) *crossClusterTaskBase {
	var eventLogger eventLogger
	if shard.GetConfig().EnableDebugMode &&
		shard.GetConfig().EnableTaskInfoLogByDomainID(taskInfo.GetDomainID()) {
		eventLogger = newEventLogger(logger, timeSource, defaultTaskEventLoggerSize)
		eventLogger.AddEvent("Created task")
	}

	return &crossClusterTaskBase{
		Info:          taskInfo,
		shard:         shard,
		priority:      ctask.NoPriority,
		attempt:       0,
		timeSource:    timeSource,
		submitTime:    timeSource.Now(),
		logger:        logger,
		eventLogger:   eventLogger,
		scopeIdx:      GetCrossClusterTaskMetricsScope(taskInfo.GetTaskType()),
		scope:         nil,
		maxRetryCount: maxRetryCount,
	}
}

// Cross cluster signal workflow task

func (c *crossClusterSignalWorkflowTask) Execute() error {
	panic("Not implement")
}

func (c *crossClusterSignalWorkflowTask) Ack() {
	// TODO: rewrite the implementation
	// current impl is just for testing purpose
	if c.settable != nil {
		c.settable.Set(types.CrossClusterTaskResponse{
			TaskID: c.Info.GetTaskID(),
		}, nil)
	}
}

func (c *crossClusterSignalWorkflowTask) Nack() {
	panic("Not implement")
}

func (c *crossClusterSignalWorkflowTask) HandleErr(
	err error,
) (retErr error) {
	panic("Not implement")
}

func (c *crossClusterSignalWorkflowTask) RetryErr(
	err error,
) bool {
	panic("Not implement")
}

func (c *crossClusterSignalWorkflowTask) Update(interface{}) error {
	panic("Not implement")
}

// Cross cluster cancel workflow task

func (c *crossClusterCancelWorkflowTask) Execute() error {
	panic("Not implement")
}

func (c *crossClusterCancelWorkflowTask) Ack() {
	panic("Not implement")
}

func (c *crossClusterCancelWorkflowTask) Nack() {
	panic("Not implement")

}

func (c *crossClusterCancelWorkflowTask) HandleErr(
	err error,
) (retErr error) {
	panic("Not implement")
}

func (c *crossClusterCancelWorkflowTask) RetryErr(
	err error,
) bool {
	panic("Not implement")
}

func (c *crossClusterCancelWorkflowTask) Update(interface{}) error {
	panic("Not implement")
}

// Cross cluster start child workflow task

func (c *crossClusterStartChildWorkflowTask) Execute() error {
	panic("Not implement")
}

func (c *crossClusterStartChildWorkflowTask) Ack() {
	panic("Not implement")

}

func (c *crossClusterStartChildWorkflowTask) Nack() {
	panic("Not implement")

}

func (c *crossClusterStartChildWorkflowTask) HandleErr(
	err error,
) (retErr error) {
	panic("Not implement")
}

func (c *crossClusterStartChildWorkflowTask) RetryErr(
	err error,
) bool {
	panic("Not implement")
}

func (c *crossClusterStartChildWorkflowTask) IsReadyForPoll() bool {
	panic("Not implement")
}

func (c *crossClusterStartChildWorkflowTask) Update(interface{}) error {
	panic("Not implement")
}

func (c *crossClusterTaskBase) State() ctask.State {
	c.Lock()
	defer c.Unlock()

	return c.state
}

func (c *crossClusterTaskBase) Priority() int {
	c.Lock()
	defer c.Unlock()

	return c.priority
}

func (c *crossClusterTaskBase) SetPriority(
	priority int,
) {
	c.Lock()
	defer c.Unlock()

	c.priority = priority
}

func (c *crossClusterTaskBase) GetShard() shard.Context {
	return c.shard
}

func (c *crossClusterTaskBase) GetAttempt() int {
	c.Lock()
	defer c.Unlock()

	return c.attempt
}

func (c *crossClusterTaskBase) GetQueueType() QueueType {
	return QueueTypeCrossCluster
}

func (c *crossClusterTaskBase) IsReadyForPoll() bool {
	c.Lock()
	defer c.Unlock()

	return c.state == ctask.TaskStatePending &&
		(c.processingState == processingStateInitialed || c.processingState == processingStateResponseRecorded)
}
