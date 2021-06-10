// Copyright (c) 2020 Uber Technologies, Inc.
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
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	ctask "github.com/uber/cadence/common/task"
	"github.com/uber/cadence/service/history/shard"
)

type (
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
		processingState ctask.CrossClusterTaskState
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
	}
)

// NewCrossClusterSignalWorkflowTask initialize cross cluster signal workflow task and task future
func NewCrossClusterSignalWorkflowTask(
	shard shard.Context,
	taskInfo Info,
	logger log.Logger,
	timeSource clock.TimeSource,
	maxRetryCount dynamicconfig.IntPropertyFn,
) Task {
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
) Task {
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
) Task {
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

func (c *crossClusterSignalWorkflowTask) Execute() error {
	panic("Not implement")
}

func (c *crossClusterSignalWorkflowTask) Ack() {
	panic("Not implement")

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

func (c *crossClusterSignalWorkflowTask) State() ctask.State {
	c.Lock()
	defer c.Unlock()

	return c.state
}

func (c *crossClusterSignalWorkflowTask) Priority() int {
	c.Lock()
	defer c.Unlock()

	return c.priority
}

func (c *crossClusterSignalWorkflowTask) SetPriority(
	priority int,
) {
	c.Lock()
	defer c.Unlock()

	c.priority = priority
}

func (c *crossClusterSignalWorkflowTask) GetShard() shard.Context {
	return c.shard
}

func (c *crossClusterSignalWorkflowTask) GetAttempt() int {
	c.Lock()
	defer c.Unlock()

	return c.attempt
}

func (c *crossClusterSignalWorkflowTask) GetQueueType() QueueType {
	return QueueTypeCrossCluster
}

func (c *crossClusterSignalWorkflowTask) IsReadyForPickup() bool {
	panic("Not implement")
}

func (c *crossClusterSignalWorkflowTask) Update(interface{}) error {
	panic("Not implement")
}

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

func (c *crossClusterCancelWorkflowTask) State() ctask.State {
	c.Lock()
	defer c.Unlock()

	return c.state
}

func (c *crossClusterCancelWorkflowTask) Priority() int {
	c.Lock()
	defer c.Unlock()

	return c.priority
}

func (c *crossClusterCancelWorkflowTask) SetPriority(
	priority int,
) {
	c.Lock()
	defer c.Unlock()

	c.priority = priority
}

func (c *crossClusterCancelWorkflowTask) GetShard() shard.Context {
	return c.shard
}

func (c *crossClusterCancelWorkflowTask) GetAttempt() int {
	c.Lock()
	defer c.Unlock()

	return c.attempt
}

func (c *crossClusterCancelWorkflowTask) GetQueueType() QueueType {
	return QueueTypeCrossCluster
}

func (c *crossClusterCancelWorkflowTask) IsReadyForPickup() bool {
	panic("Not implement")
}

func (c *crossClusterCancelWorkflowTask) Update(interface{}) error {
	panic("Not implement")
}

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

func (c *crossClusterStartChildWorkflowTask) State() ctask.State {
	c.Lock()
	defer c.Unlock()

	return c.state
}

func (c *crossClusterStartChildWorkflowTask) Priority() int {
	c.Lock()
	defer c.Unlock()

	return c.priority
}

func (c *crossClusterStartChildWorkflowTask) SetPriority(
	priority int,
) {
	c.Lock()
	defer c.Unlock()

	c.priority = priority
}

func (c *crossClusterStartChildWorkflowTask) GetShard() shard.Context {
	return c.shard
}

func (c *crossClusterStartChildWorkflowTask) GetAttempt() int {
	c.Lock()
	defer c.Unlock()

	return c.attempt
}

func (c *crossClusterStartChildWorkflowTask) GetQueueType() QueueType {
	return QueueTypeCrossCluster
}

func (c *crossClusterStartChildWorkflowTask) IsReadyForPickup() bool {
	panic("Not implement")
}

func (c *crossClusterStartChildWorkflowTask) Update(interface{}) error {
	panic("Not implement")
}
