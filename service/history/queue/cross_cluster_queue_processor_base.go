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
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
	"github.com/uber/cadence/service/history/task"
)

const (
	taskCleanupTimeout = time.Second * 5
)

type (
	crossClusterTaskKey = transferTaskKey

	crossClusterQueueProcessorBase struct {
		sync.Mutex // TODO: rename this to outstandingTaskCount lock

		*processorBase
		targetCluster   string
		taskInitializer task.Initializer

		processingLock sync.Mutex
		backoffTimer   map[int]*time.Timer
		shouldProcess  map[int]bool

		outstandingTaskCount int // maybe use atomic.Value?
		ackLevel             int64

		notifyCh         chan struct{}
		processCh        chan struct{}
		failoverNotifyCh chan struct{}
	}
)

func newCrossClusterQueueProcessorBase(
	shard shard.Context,
	clusterName string,
	executionCache *execution.Cache,
	taskProcessor task.Processor,
	taskExecutor task.Executor,
	logger log.Logger,
) *crossClusterQueueProcessorBase {
	config := shard.GetConfig()
	options := newCrossClusterQueueProcessorOptions(config)

	logger = logger.WithTags(tag.ClusterName(clusterName))

	updateMaxReadLevel := func() task.Key {
		return newCrossClusterTaskKey(shard.GetTransferMaxReadLevel())
	}

	updateProcessingQueueStates := func(states []ProcessingQueueState) error {
		pStates := convertToPersistenceTransferProcessingQueueStates(states)
		return shard.UpdateCrossClusterProcessingQueueStates(clusterName, pStates)
	}

	queueShutdown := func() error {
		return nil
	}

	return newCrossClusterQueueProcessorBaseHelper(
		shard,
		clusterName,
		executionCache,
		convertFromPersistenceTransferProcessingQueueStates(shard.GetCrossClusterProcessingQueueStates(clusterName)),
		taskProcessor,
		options,
		updateMaxReadLevel,
		updateProcessingQueueStates,
		queueShutdown,
		taskExecutor,
		logger,
		shard.GetMetricsClient(),
	)
}

func newCrossClusterQueueProcessorBaseHelper(
	shard shard.Context,
	targetCluster string,
	executionCache *execution.Cache,
	processingQueueStates []ProcessingQueueState,
	taskProcessor task.Processor,
	options *queueProcessorOptions,
	updateMaxReadLevel updateMaxReadLevelFn,
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
		updateMaxReadLevel,
		nil,
		updateProcessingQueueStates,
		queueShutdown,
		logger.WithTags(tag.ComponentTransferQueue),
		metricsClient,
	)

	base := &crossClusterQueueProcessorBase{
		processorBase: processorBase,
		targetCluster: targetCluster,

		backoffTimer:  make(map[int]*time.Timer),
		shouldProcess: make(map[int]bool),

		notifyCh:         make(chan struct{}, 1),
		processCh:        make(chan struct{}, 1),
		failoverNotifyCh: make(chan struct{}, 1),
	}
	base.taskInitializer = func(taskInfo task.Info) task.Task {
		return task.NewCrossClusterSourceTask(
			shard,
			targetCluster,
			executionCache,
			taskInfo,
			taskExecutor,
			taskProcessor,
			task.InitializeLoggerForTask(shard.GetShardID(), taskInfo, logger),
			func(t task.Task) {
				_, _ = base.submitTask(t)
			},
			shard.GetConfig().TaskCriticalRetryCount,
		)
	}
	return base
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
	go c.processLoop()
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

func (c *crossClusterQueueProcessorBase) processLoop() {
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
			processFinished, ackLevel, err := c.updateAckLevel()
			if err == shard.ErrShardClosed || (err == nil && processFinished) {
				go c.Stop()
				break processorPumpLoop
			}
			if ackLevel != nil {
				c.ackLevel, err = c.completeAckTasks(ackLevel.(crossClusterTaskKey).taskID)
				if err != nil {
					c.logger.Error("failed to complete ack tasks", tag.Error(err))
					c.metricsScope.IncCounter(metrics.TaskBatchCompleteFailure)
				}
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
					c.resetBackoffTimer()
				}
				// re-enqueue the event to see if we need keep re-dispatching or load new tasks from persistence
				select {
				case c.processCh <- struct{}{}:
				default:
				}
				continue processorPumpLoop
			}

			c.Lock()
			pendingTaskCount := c.outstandingTaskCount
			c.Unlock()
			if pendingTaskCount > c.options.MaxPendingTaskSize() {
				c.resetBackoffTimer()

				select {
				case c.processCh <- struct{}{}:
				default:
				}
				continue processorPumpLoop
			}

			c.processQueueCollections()
		case notification := <-c.actionNotifyCh:
			c.handleActionNotification(notification)
		case <-c.failoverNotifyCh:
			c.validateOutstandingTasks()
		}
	}
}
func (c *crossClusterQueueProcessorBase) completeAckTasks(nextAckLevel int64) (int64, error) {

	c.logger.Debug(fmt.Sprintf("Start completing cross cluster task from: %v, to %v.", c.ackLevel, nextAckLevel))
	c.metricsScope.IncCounter(metrics.TaskBatchCompleteCounter)

	if c.ackLevel >= nextAckLevel {
		return c.ackLevel, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), taskCleanupTimeout)
	defer cancel()

	if err := c.shard.GetExecutionManager().RangeCompleteCrossClusterTask(ctx, &persistence.RangeCompleteCrossClusterTaskRequest{
		TargetCluster:        c.targetCluster,
		ExclusiveBeginTaskID: c.ackLevel,
		InclusiveEndTaskID:   nextAckLevel,
	}); err != nil {
		return c.ackLevel, err
	}

	return nextAckLevel, nil
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

	err := backoff.Retry(op, persistenceOperationRetryPolicy, persistence.IsBackgroundTransientError)
	if err != nil {
		return nil, false, err
	}

	return response.Tasks, len(response.NextPageToken) != 0, nil
}

func (c *crossClusterQueueProcessorBase) pollTasks() []*types.CrossClusterTaskRequest {

	// TODO: control the number of tasks returned
	var result []*types.CrossClusterTaskRequest
	for _, queueTask := range c.getTasks() {
		crossClusterTask, ok := queueTask.(task.CrossClusterTask)
		if !ok {
			panic("received non cross cluster task")
		}
		if crossClusterTask.IsReadyForPoll() {
			request, err := crossClusterTask.GetCrossClusterRequest()
			if err != nil {
				if !crossClusterTask.IsValid() {
					if _, err := c.submitTask(crossClusterTask); err != nil {
						break
					}
				}
				c.logger.Error("Failed to get cross cluster request", tag.Error(err))
			}
			if request != nil {
				// if request is nil, nothing need to be done for the task,
				// task already acked in GetCrossClusterRequest()
				result = append(result, request)
			}
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
	response *types.CrossClusterTaskResponse,
) error {

	queueTask, err := c.getTask(newCrossClusterTaskKey(response.GetTaskID()))
	// TODO: if task not found, should not return error, it's expected
	if err != nil {
		return err
	}
	crossClusterTask, ok := queueTask.(task.CrossClusterTask)
	if !ok {
		panic("Received non cross cluster task.")
	}

	if err := crossClusterTask.RecordResponse(response); err != nil {
		c.logger.Error("failed to update cross cluster task",
			tag.TaskID(crossClusterTask.GetTaskID()),
			tag.WorkflowDomainID(crossClusterTask.GetDomainID()),
			tag.WorkflowID(crossClusterTask.GetWorkflowID()),
			tag.WorkflowRunID(crossClusterTask.GetRunID()),
			tag.Error(err))
		return err
	}

	// if the task fails to submit to main queue, it will use redispatch queue
	// note the only possible error here is shutdown
	_, err = c.submitTask(crossClusterTask)
	return err
}

func (c *crossClusterQueueProcessorBase) addOutstandingTaskCountLocked(count int) {
	c.outstandingTaskCount += count
}

// TODO: this will be called in cross cluster task Acked function.
func (c *crossClusterQueueProcessorBase) removeOutstandingTaskCount(count int) {
	c.Lock()
	defer c.Unlock()

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

func (c *crossClusterQueueProcessorBase) resetBackoffTimer() {
	c.processingLock.Lock()
	defer c.processingLock.Unlock()

	for _, queueCollection := range c.processingQueueCollections {
		level := queueCollection.Level()
		c.setupBackoffTimer(level)
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
	switch notification.action.ActionType {
	case ActionTypeGetTasks:
		tasks := c.pollTasks()
		notification.resultNotificationCh <- actionResultNotification{
			result: &ActionResult{
				ActionType: ActionTypeGetTasks,
				GetTasksResult: &GetTasksResult{
					TaskRequests: tasks,
				},
			},
		}
	case ActionTypeUpdateTask:
		actionAttr := notification.action.UpdateTaskAttributes
		var err error
		for _, task := range actionAttr.TaskResponses {
			if err = c.updateTask(task); err != nil {
				// TODO: log the error and keep updating the rest of the tasks
				break
			}
		}
		notification.resultNotificationCh <- actionResultNotification{
			result: &ActionResult{
				ActionType:       ActionTypeUpdateTask,
				UpdateTaskResult: &UpdateTasksResult{},
			},
			err: err,
		}
	default:
		c.processorBase.handleActionNotification(notification, func() {
			switch notification.action.ActionType {
			case ActionTypeReset:
				c.readyForProcess(defaultProcessingQueueLevel)
			}
		})
	}
}

func (c *crossClusterQueueProcessorBase) processQueueCollections() {
	for _, queueCollection := range c.processingQueueCollections {
		level := queueCollection.Level()
		c.processingLock.Lock()
		if shouldProcess, ok := c.shouldProcess[level]; !ok || !shouldProcess {
			c.processingLock.Unlock()
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

		c.Lock()
		if c.outstandingTaskCount > c.options.MaxPendingTaskSize() {
			c.logger.Warn("too many outstanding tasks in cross cluster queue.")
			c.Unlock()
			return
		}
		c.addOutstandingTaskCountLocked(len(taskInfos))
		c.Unlock()

		tasks := make(map[task.Key]task.Task)
		newTaskID := readLevel.(transferTaskKey).taskID
		for _, taskInfo := range taskInfos {
			if !domainFilter.Filter(taskInfo.GetDomainID()) {
				continue
			}
			task := c.taskInitializer(taskInfo)
			tasks[newCrossClusterTaskKey(taskInfo.GetTaskID())] = task
			newTaskID = task.GetTaskID()
		}

		var newReadLevel task.Key
		if !more {
			newReadLevel = maxReadLevel
		} else {
			newReadLevel = newCrossClusterTaskKey(newTaskID)
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

func (c *crossClusterQueueProcessorBase) notifyDomainFailover() {
	select {
	case c.failoverNotifyCh <- struct{}{}:
	default:
	}
}

// validateOutstandingTasks will be triggered when the cluster received a domain fail over event
// TODO: this function should only check tasks that are not yet submitted to the taskProcessor
// (or in the redispatch queue), i.e. the task should currently be available for poll
// otherwise, we may end up submitting the task twice
func (c *crossClusterQueueProcessorBase) validateOutstandingTasks() {
	tasks := c.getTasks()
	for _, t := range tasks {
		if crossClusterTask, ok := t.(task.CrossClusterTask); ok && !crossClusterTask.IsValid() {
			_, err := c.submitTask(t)
			if err != nil {
				// the queue is shutting down
				return
			}
		}
	}
}

func newCrossClusterTaskKey(
	taskID int64,
) task.Key {
	return crossClusterTaskKey{
		taskID: taskID,
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

	options.EnableSplit = dynamicconfig.GetBoolPropertyFn(false)
	options.SplitMaxLevel = config.QueueProcessorSplitMaxLevel
	options.EnableRandomSplitByDomainID = config.QueueProcessorEnableRandomSplitByDomainID
	options.RandomSplitProbability = config.QueueProcessorRandomSplitProbability
	options.EnablePendingTaskSplitByDomainID = config.QueueProcessorEnablePendingTaskSplitByDomainID
	options.PendingTaskSplitThreshold = config.QueueProcessorPendingTaskSplitThreshold
	options.EnableStuckTaskSplitByDomainID = config.QueueProcessorEnableStuckTaskSplitByDomainID
	options.StuckTaskSplitThreshold = config.QueueProcessorStuckTaskSplitThreshold
	options.SplitLookAheadDurationByDomainID = config.QueueProcessorSplitLookAheadDurationByDomainID

	options.EnablePersistQueueStates = dynamicconfig.GetBoolPropertyFn(true)
	options.EnableLoadQueueStates = dynamicconfig.GetBoolPropertyFn(true)

	options.MetricScope = metrics.CrossClusterQueueProcessorScope
	options.RedispatchInterval = config.ActiveTaskRedispatchInterval
	return options
}
