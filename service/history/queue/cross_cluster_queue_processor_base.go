// Copyright (c) 2017-2021 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package queue

import (
	"context"
	"errors"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/service/history/shard"
	"github.com/uber/cadence/service/history/task"
)

var (
	errTaskCallbackNotFound = errors.New("task callback not found")
)

type (
	updateCallback func(interface{}) error

	crossClusterQueueProcessorBase struct {
		sync.Mutex

		*processorBase
		targetCluster    string
		maxTaskCount     int
		maxTaskID        *int64
		initialReadLevel task.Key
		taskInitializer  task.Initializer
		outstandingTasks map[task.Key]updateCallback
	}
)

// TODO: add taskInitializer
func newCrossClusterQueueProcessor(
	shard shard.Context,
	targetCluster string,
	maxQueueLength int,
	processingQueueStates []ProcessingQueueState,
	taskProcessor task.Processor,
	options *queueProcessorOptions,
	updateProcessingQueueStates updateProcessingQueueStatesFn,
	queueShutdown queueShutdownFn,
	logger log.Logger,
	metricsClient metrics.Client,
) *crossClusterQueueProcessorBase {
	processorBase := newProcessorBase(
		shard,
		processingQueueStates,
		taskProcessor,
		options,
		nil,
		nil,
		updateProcessingQueueStates,
		queueShutdown,
		logger.WithTags(tag.ComponentTransferQueue),
		metricsClient,
	)

	ackLevel := newTransferTaskKey(math.MaxInt64)
	for _, state := range processingQueueStates {
		if state.AckLevel().Less(ackLevel) {
			ackLevel = state.AckLevel()
		}
	}
	crossClusterQueueProcessorBase := &crossClusterQueueProcessorBase{
		processorBase:    processorBase,
		targetCluster:    targetCluster,
		maxTaskCount:     maxQueueLength,
		initialReadLevel: ackLevel,
		outstandingTasks: make(map[task.Key]updateCallback),
	}

	return crossClusterQueueProcessorBase
}

func (c *crossClusterQueueProcessorBase) Start() {
	if !atomic.CompareAndSwapInt32(&c.status, common.DaemonStatusInitialized, common.DaemonStatusStarted) {
		return
	}

	c.logger.Info("Cross cluster queue processor state changed", tag.LifeCycleStarting)
	defer c.logger.Info("Cross cluster queue processor state changed", tag.LifeCycleStarted)

	c.redispatcher.Start()

	c.shutdownWG.Add(1)
	go c.processorPump(c.initialReadLevel)
}

func (c *crossClusterQueueProcessorBase) Stop() {
	if !atomic.CompareAndSwapInt32(&c.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}

	c.logger.Info("Cross cluster queue processor state changed", tag.LifeCycleStopping)
	defer c.logger.Info("Cross cluster queue processor state changed", tag.LifeCycleStopped)

	close(c.shutdownCh)

	if success := common.AwaitWaitGroup(&c.shutdownWG, time.Minute); !success {
		c.logger.Warn("", tag.LifeCycleStopTimedout)
	}

	c.redispatcher.Stop()
}

func (c *crossClusterQueueProcessorBase) processorPump(readLevel task.Key) {
	defer c.shutdownWG.Done()

	updateAckTimer := time.NewTimer(backoff.JitDuration(
		c.options.UpdateAckInterval(),
		c.options.UpdateAckIntervalJitterCoefficient(),
	))
	nextPollTimer := time.NewTimer(backoff.JitDuration(
		c.options.MaxPollInterval(),
		c.options.MaxPollIntervalJitterCoefficient(),
	))
	defer updateAckTimer.Stop()
	defer nextPollTimer.Stop()

processorPumpLoop:
	for {
		select {
		case <-c.shutdownCh:
			break processorPumpLoop
		case <-updateAckTimer.C:
			processFinished, err := c.updateAckLevel()
			if err == shard.ErrShardClosed || (err == nil && processFinished) {
				go c.Stop()
				break processorPumpLoop
			}
			updateAckTimer.Reset(backoff.JitDuration(
				c.options.UpdateAckInterval(),
				c.options.UpdateAckIntervalJitterCoefficient(),
			))
		case <-nextPollTimer.C:
			var err error
			readLevel, err = c.refillOutstandingTasks(readLevel)
			if err != nil {
				//retry if the read tasks failed
				nextPollTimer.Reset(backoff.JitDuration(
					c.options.PollBackoffInterval(),
					c.options.PollBackoffIntervalJitterCoefficient(),
				))
			}
			nextPollTimer.Reset(backoff.JitDuration(
				c.options.MaxPollInterval(),
				c.options.MaxPollIntervalJitterCoefficient(),
			))
		}
	}
}

func (c *crossClusterQueueProcessorBase) readTasks(
	readLevel task.Key,
) ([]*persistence.CrossClusterTaskInfo, bool, error) {

	var response *persistence.GetCrossClusterTasksResponse
	op := func() error {
		var err error
		response, err = c.shard.GetExecutionManager().GetCrossClusterTasks(
			context.Background(),
			&persistence.GetCrossClusterTasksRequest{
				TargetCluster: c.targetCluster,
				ReadLevel:     readLevel.(transferTaskKey).taskID,
				MaxReadLevel:  math.MaxInt64,
				BatchSize:     c.options.BatchSize(),
			})
		return err
	}

	err := backoff.Retry(op, persistenceOperationRetryPolicy, persistence.IsTransientError)
	if err != nil {
		return nil, false, err
	}

	return response.Tasks, len(response.NextPageToken) != 0, nil
}

func (c *crossClusterQueueProcessorBase) pollTasks() []task.Task {

	c.Lock()
	defer c.Unlock()

	var result []task.Task
	for key, _ := range c.outstandingTasks {
		//TODO: uncomment the code when IsReadyForPoll is ready
		_, err := c.getTask(key)
		if err != nil {
			continue
		}
		//if task.IsReadyForPoll() {
		//	result = append(result, task)
		//}
	}
	return result
}

func (c *crossClusterQueueProcessorBase) getTask(key task.Key) (task.Task, error) {
	for _, queue := range c.processingQueueCollections {
		if task, err := queue.GetTask(key); err == nil {
			return task, nil
		}
	}
	return nil, errTaskNotFound
}

func (c *crossClusterQueueProcessorBase) updateTask(
	taskKey task.Key,
	remoteResponse interface{},
) error {

	c.Lock()
	taskCallback, ok := c.outstandingTasks[taskKey]
	c.Unlock()
	if !ok {
		return errTaskCallbackNotFound
	}

	task, err := c.getTask(taskKey)
	if err != nil {
		return err
	}

	err = taskCallback(remoteResponse)
	if err != nil {
		return err
	}

	// if the task fails to submit to main queue, it will use redispatch queue
	_, err = c.submitTask(task)
	return err
}

func (c *crossClusterQueueProcessorBase) refillOutstandingTasks(
	readLevel task.Key,
) (task.Key, error) {
	c.Lock()
	if !c.hasMoreTaskLocked(readLevel) || c.getAvailableSpaceLocked() <= 0 {
		c.Unlock()
		return readLevel, nil
	}
	c.Unlock()

	tasks, _, err := c.readTasks(readLevel)
	if err != nil {
		return readLevel, err
	}

	if len(tasks) == 0 {
		return readLevel, nil
	}

	c.Lock()
	defer c.Unlock()
	availableSpace := c.getAvailableSpaceLocked()
	maxReadLevel := readLevel
	for _, task := range tasks {
		currentTaskKey := transferTaskKey{taskID: task.TaskID}
		if _, ok := c.outstandingTasks[currentTaskKey]; ok {
			continue
		}
		if availableSpace > 0 {
			//newTask, cb := c.taskInitializer(task)
			//taskPair := taskCallback{
			//	task: newTask,
			//	updateCallback: cb,
			//}
			fakeCallback := func(interface{}) error {
				return nil
			}
			c.outstandingTasks[currentTaskKey] = fakeCallback
			availableSpace--
			taskID := newTransferTaskKey(task.TaskID)
			if maxReadLevel.Less(taskID) {
				maxReadLevel = taskID
			}
		} else {
			return maxReadLevel, nil
		}
	}
	return maxReadLevel, nil
}

func (c *crossClusterQueueProcessorBase) notifyNewTask(
	tasks []persistence.Task,
) {

	if len(tasks) == 0 {
		return
	}
	maxTaskID := tasks[0].GetTaskID()
	for _, task := range tasks {
		if maxTaskID < task.GetTaskID() {
			maxTaskID = task.GetTaskID()
		}
	}

	c.Lock()
	defer c.Unlock()
	if c.maxTaskID == nil || *c.maxTaskID < maxTaskID {
		c.maxTaskID = &maxTaskID
	}
}

func (c *crossClusterQueueProcessorBase) hasMoreTaskLocked(readLevel task.Key) bool {
	return c.maxTaskID == nil || readLevel.Less(newTransferTaskKey(*c.maxTaskID))
}

func (c *crossClusterQueueProcessorBase) getAvailableSpaceLocked() int {
	return len(c.outstandingTasks)
}
