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
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/shard"
	"github.com/uber/cadence/service/history/task"
)

type (
	crossClusterQueueProcessorBase struct {
		sync.Mutex

		*processorBase
		targetCluster   string
		taskInitializer task.Initializer

		processingLock sync.Mutex
		backoffTimer   map[int]*time.Timer
		shouldProcess  map[int]bool

		outstandingTaskCount int

		notifyCh  chan struct{}
		processCh chan struct{}
	}
)

// TODO: implement initialize taskInitializer this intializer should have a callback to update outstandingTaskCount
func newCrossClusterQueueProcessorBase(
	shard shard.Context,
	targetCluster string,
	processingQueueStates []ProcessingQueueState,
	taskProcessor task.Processor,
	options *queueProcessorOptions,
	updateProcessingQueueStates updateProcessingQueueStatesFn,
	queueShutdown queueShutdownFn,
	taskExecutor task.Executor,
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
		processorBase: processorBase,
		targetCluster: targetCluster,
		// taskInitializer: // TODO

		backoffTimer:  make(map[int]*time.Timer),
		shouldProcess: make(map[int]bool),

		notifyCh:  make(chan struct{}, 1),
		processCh: make(chan struct{}, 1),
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

	// trigger an initial load of tasks
	c.notifyAllQueueCollections()

	c.shutdownWG.Add(1)
	go c.processorPump()
}

func (c *crossClusterQueueProcessorBase) Stop() {
	if !atomic.CompareAndSwapInt32(&c.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}

	c.logger.Info("Cross cluster queue processor state changed", tag.LifeCycleStopping)
	defer c.logger.Info("Cross cluster queue processor state changed", tag.LifeCycleStopped)

	close(c.shutdownCh)
	c.processingLock.Lock()
	for _, timer := range c.backoffTimer {
		timer.Stop()
	}
	for level := range c.shouldProcess {
		c.shouldProcess[level] = false
	}
	c.processingLock.Unlock()

	if success := common.AwaitWaitGroup(&c.shutdownWG, time.Minute); !success {
		c.logger.Warn("", tag.LifeCycleStopTimedout)
	}

	c.redispatcher.Stop()
}

func (c *crossClusterQueueProcessorBase) processorPump() {
	defer c.shutdownWG.Done()

	updateAckTimer := time.NewTimer(backoff.JitDuration(
		c.options.UpdateAckInterval(),
		c.options.UpdateAckIntervalJitterCoefficient(),
	))
	defer updateAckTimer.Stop()

	maxPollTimer := time.NewTimer(backoff.JitDuration(
		c.options.MaxPollInterval(),
		c.options.MaxPollIntervalJitterCoefficient(),
	))
	defer maxPollTimer.Stop()

processorPumpLoop:
	for {
		select {
		case <-c.shutdownCh:
			break processorPumpLoop
		case <-c.notifyCh:
			c.notifyAllQueueCollections()
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
		case <-maxPollTimer.C:
			c.notifyAllQueueCollections()
			maxPollTimer.Reset(backoff.JitDuration(
				c.options.MaxPollInterval(),
				c.options.MaxPollIntervalJitterCoefficient(),
			))
		case <-c.processCh:
			maxRedispatchQueueSize := c.options.MaxRedispatchQueueSize()
			if c.redispatcher.Size() > maxRedispatchQueueSize {
				// has too many pending tasks in re-dispatch queue, block loading tasks from persistence
				c.redispatcher.Redispatch(maxRedispatchQueueSize)
				if c.redispatcher.Size() > maxRedispatchQueueSize {
					// if redispatcher still has a large number of tasks
					// this only happens when system is under very high load
					// we should backoff here instead of keeping submitting tasks to task processor
					time.Sleep(backoff.JitDuration(
						c.options.PollBackoffInterval(),
						c.options.PollBackoffIntervalJitterCoefficient(),
					))
				}
				// re-enqueue the event to see if we need keep re-dispatching or load new tasks from persistence
				select {
				case c.processCh <- struct{}{}:
				default:
				}
				continue processorPumpLoop
			}
			c.processQueueCollections()
		case notification := <-c.actionNotifyCh:
			c.handleActionNotification(notification)
		}
	}
}

func (c *crossClusterQueueProcessorBase) readTasks(
	readLevel task.Key,
	maxReadLevel task.Key,
) ([]*persistence.CrossClusterTaskInfo, bool, error) {

	var response *persistence.GetCrossClusterTasksResponse
	op := func() error {
		var err error
		response, err = c.shard.GetExecutionManager().GetCrossClusterTasks(
			context.Background(),
			&persistence.GetCrossClusterTasksRequest{
				TargetCluster: c.targetCluster,
				ReadLevel:     readLevel.(transferTaskKey).taskID,
				MaxReadLevel:  maxReadLevel.(transferTaskKey).taskID,
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

	var result []task.Task
	for _, queueTask := range c.getTasks() {
		crossClusterTask, ok := queueTask.(task.CrossClusterTask)
		if !ok {
			panic("received non cross cluster task")
		}
		if crossClusterTask.IsReadyForPoll() {
			result = append(result, crossClusterTask)
		}
	}
	return result
}

func (c *crossClusterQueueProcessorBase) getTask(key task.Key) (task.Task, error) {
	for _, collection := range c.processingQueueCollections {
		if task, err := collection.GetTask(key); err == nil {
			return task, nil
		}
	}
	return nil, errTaskNotFound
}

func (c *crossClusterQueueProcessorBase) getTasks() []task.Task {
	var outstandingTask []task.Task
	for _, collection := range c.processingQueueCollections {
		outstandingTask = append(outstandingTask, collection.GetTasks()...)
	}
	return outstandingTask
}

func (c *crossClusterQueueProcessorBase) updateTask(
	taskKey task.Key,
	remoteResponse interface{},
) error {

	queueTask, err := c.getTask(taskKey)
	if err != nil {
		return err
	}
	crossClusterTask, ok := queueTask.(task.CrossClusterTask)
	if ok {
		if err := crossClusterTask.Update(remoteResponse); err != nil {
			return err
		}
	} else {
		panic("Received non cross cluster task.")
	}

	// if the task fails to submit to main queue, it will use redispatch queue
	_, err = c.submitTask(crossClusterTask)
	return err
}

func (c *crossClusterQueueProcessorBase) addOutstandingTaskCountLocked(count int) {
	c.outstandingTaskCount += count
}

// TODO: this will be called in cross cluster task Acked function.
func (c *crossClusterQueueProcessorBase) removeOutstandingTaskCountLocked(count int) {
	c.outstandingTaskCount -= count
}

func (c *crossClusterQueueProcessorBase) notifyAllQueueCollections() {
	for _, queueCollection := range c.processingQueueCollections {
		c.readyForProcess(queueCollection.Level())
	}
}

func (c *crossClusterQueueProcessorBase) readyForProcess(level int) {
	c.processingLock.Lock()
	defer c.processingLock.Unlock()

	if _, ok := c.backoffTimer[level]; ok {
		// current level is being throttled
		return
	}

	c.shouldProcess[level] = true

	// trigger the actual processing
	select {
	case c.processCh <- struct{}{}:
	default:
	}
}

func (c *crossClusterQueueProcessorBase) setupBackoffTimer(level int) {
	c.processingLock.Lock()
	defer c.processingLock.Unlock()

	c.metricsScope.IncCounter(metrics.ProcessingQueueThrottledCounter)
	c.logger.Info("Throttled processing queue", tag.QueueLevel(level))

	if _, ok := c.backoffTimer[level]; ok {
		// there's an existing backoff timer, no-op
		// this case should not happen
		return
	}

	backoffDuration := backoff.JitDuration(
		c.options.PollBackoffInterval(),
		c.options.PollBackoffIntervalJitterCoefficient(),
	)
	c.backoffTimer[level] = time.AfterFunc(backoffDuration, func() {
		select {
		case <-c.shutdownCh:
			return
		default:
		}

		c.processingLock.Lock()
		defer c.processingLock.Unlock()

		c.shouldProcess[level] = true
		delete(c.backoffTimer, level)

		// trigger the actual processing
		select {
		case c.processCh <- struct{}{}:
		default:
		}
	})
}

func (c *crossClusterQueueProcessorBase) handleActionNotification(notification actionNotification) {
	c.processorBase.handleActionNotification(notification, func() {
		switch notification.action.ActionType {
		case ActionTypeReset:
			c.readyForProcess(defaultProcessingQueueLevel)
		case ActionTypeGetTasks:
			tasks := c.getTasks()
			notification.resultNotificationCh <- actionResultNotification{
				result: &ActionResult{
					ActionType: ActionTypeGetTasks,
					GetTasksResult: &GetTasksResult{
						tasks: tasks,
					},
				},
			}
		case ActionTypeUpdateTask:
			action := notification.action.UpdateTaskAttributes
			err := c.updateTask(newTransferTaskKey(action.taskID), action.result)
			notification.resultNotificationCh <- actionResultNotification{
				result: &ActionResult{
					ActionType:       ActionTypeUpdateTask,
					UpdateTaskResult: &UpdateTaskResult{},
				},
				err: err,
			}
		}
	})
}

func (c *crossClusterQueueProcessorBase) processQueueCollections() {
	for _, queueCollection := range c.processingQueueCollections {
		level := queueCollection.Level()
		c.processingLock.Lock()
		if shouldProcess, ok := c.shouldProcess[level]; !ok || !shouldProcess {
			continue
		}
		c.shouldProcess[level] = false
		c.processingLock.Unlock()

		activeQueue := queueCollection.ActiveQueue()
		if activeQueue == nil {
			// process for this queue collection has finished
			// it's possible that new queue will be added to this collection later though,
			// pollTime will be updated after split/merge
			continue
		}

		readLevel := activeQueue.State().ReadLevel()
		maxReadLevel := minTaskKey(activeQueue.State().MaxLevel(), c.updateMaxReadLevel())
		domainFilter := activeQueue.State().DomainFilter()

		if !readLevel.Less(maxReadLevel) {
			// no task need to be processed for now, wait for new task notification
			// note that if taskID for new task is still less than readLevel, the notification
			// will just be a no-op and there's no DB requests.
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), loadQueueTaskThrottleRetryDelay)
		if err := c.rateLimiter.Wait(ctx); err != nil {
			cancel()
			if level != defaultProcessingQueueLevel {
				c.setupBackoffTimer(level)
			} else {
				c.readyForProcess(level)
			}
			continue
		}
		cancel()

		taskInfos, more, err := c.readTasks(readLevel, maxReadLevel)
		if err != nil {
			c.logger.Error("Processor unable to retrieve tasks", tag.Error(err))
			c.readyForProcess(level) // re-enqueue the event
			continue
		}

		tasks := make(map[task.Key]task.Task)
		newTaskID := readLevel.(transferTaskKey).taskID

		c.Lock()
		for _, taskInfo := range taskInfos {
			if !domainFilter.Filter(taskInfo.GetDomainID()) {
				continue
			}

			if c.outstandingTaskCount > c.options.MaxPendingTaskSize() {
				more = true
				break
			}
			task := c.taskInitializer(taskInfo)
			tasks[newTransferTaskKey(taskInfo.GetTaskID())] = task
			newTaskID = task.GetTaskID()
		}
		c.addOutstandingTaskCountLocked(len(tasks))
		c.Unlock()

		var newReadLevel task.Key
		if !more {
			newReadLevel = maxReadLevel
		} else {
			newReadLevel = newTransferTaskKey(newTaskID)
		}

		queueCollection.AddTasks(tasks, newReadLevel)
		newActiveQueue := queueCollection.ActiveQueue()
		if more || (newActiveQueue != nil && newActiveQueue != activeQueue) {
			// more tasks for the current active queue or the active queue has changed
			c.readyForProcess(level)
		}
	}
}

func (c *crossClusterQueueProcessorBase) notifyNewTask() {
	select {
	case c.notifyCh <- struct{}{}:
	default:
	}
}

func newCrossClusterQueueProcessorOptions(
	config *config.Config,
) *queueProcessorOptions {
	options := &queueProcessorOptions{
		BatchSize:                            config.CrossClusterTaskBatchSize,
		MaxPollRPS:                           config.CrossClusterProcessorMaxPollRPS,
		MaxPollInterval:                      config.CrossClusterProcessorMaxPollInterval,
		MaxPollIntervalJitterCoefficient:     config.CrossClusterProcessorMaxPollIntervalJitterCoefficient,
		UpdateAckInterval:                    config.CrossClusterProcessorUpdateAckInterval,
		UpdateAckIntervalJitterCoefficient:   config.CrossClusterProcessorUpdateAckIntervalJitterCoefficient,
		RedispatchIntervalJitterCoefficient:  config.TaskRedispatchIntervalJitterCoefficient,
		MaxRedispatchQueueSize:               config.CrossClusterProcessorMaxRedispatchQueueSize,
		SplitQueueInterval:                   config.CrossClusterProcessorSplitQueueInterval,
		SplitQueueIntervalJitterCoefficient:  config.CrossClusterProcessorSplitQueueIntervalJitterCoefficient,
		PollBackoffInterval:                  config.QueueProcessorPollBackoffInterval,
		PollBackoffIntervalJitterCoefficient: config.QueueProcessorPollBackoffIntervalJitterCoefficient,
		EnableValidator:                      config.CrossClusterProcessorEnableValidator,
		ValidationInterval:                   config.CrossClusterProcessorValidationInterval,
	}

	options.EnableSplit = config.QueueProcessorEnableSplit
	options.SplitMaxLevel = config.QueueProcessorSplitMaxLevel
	options.EnableRandomSplitByDomainID = config.QueueProcessorEnableRandomSplitByDomainID
	options.RandomSplitProbability = config.QueueProcessorRandomSplitProbability
	options.EnablePendingTaskSplitByDomainID = config.QueueProcessorEnablePendingTaskSplitByDomainID
	options.PendingTaskSplitThreshold = config.QueueProcessorPendingTaskSplitThreshold
	options.EnableStuckTaskSplitByDomainID = config.QueueProcessorEnableStuckTaskSplitByDomainID
	options.StuckTaskSplitThreshold = config.QueueProcessorStuckTaskSplitThreshold
	options.SplitLookAheadDurationByDomainID = config.QueueProcessorSplitLookAheadDurationByDomainID

	options.EnablePersistQueueStates = config.QueueProcessorEnablePersistQueueStates
	options.EnableLoadQueueStates = config.QueueProcessorEnableLoadQueueStates

	options.MetricScope = metrics.CrossClusterQueueProcessorScope
	options.RedispatchInterval = config.ActiveTaskRedispatchInterval
	return options
}
