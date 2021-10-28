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
	"github.com/uber/cadence/common/collection"
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
		*processorBase

		targetCluster   string
		taskInitializer task.Initializer

		processingLock sync.Mutex
		backoffTimer   map[int]*time.Timer
		shouldProcess  map[int]bool

		ackLevel int64

		// map from taskID -> task.CrossClusterTask
		// ordered by insertion time
		readyForPollTasks collection.OrderedMap

		notifyCh         chan struct{}
		processCh        chan struct{}
		failoverNotifyCh chan map[string]struct{}
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
	options := newCrossClusterQueueProcessorOptions(shard.GetConfig())

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

		readyForPollTasks: collection.NewConcurrentOrderedMap(),

		notifyCh:         make(chan struct{}, 1),
		processCh:        make(chan struct{}, 1),
		failoverNotifyCh: make(chan map[string]struct{}, 10),
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
			func(t task.CrossClusterTask) {
				base.readyForPollTasks.Put(t.GetTaskID(), t)
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
					c.backoffAllQueueCollections()
				}
				continue processorPumpLoop
			}

			numReadyForPoll := c.readyForPollTasks.Len()
			c.metricsScope.RecordTimer(metrics.CrossClusterTaskPendingTimer, time.Duration(numReadyForPoll))
			if numReadyForPoll > c.options.MaxPendingTaskSize() {
				c.logger.Warn("too many outstanding ready for poll tasks in cross cluster queue.")
				c.backoffAllQueueCollections()
				continue processorPumpLoop
			}

			c.processQueueCollections()
		case notification := <-c.actionNotifyCh:
			c.handleActionNotification(notification)
		case domains := <-c.failoverNotifyCh:
			c.validateOutstandingTasks(domains)
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

	for {
		pageSize := c.options.DeleteBatchSize()
		resp, err := c.shard.GetExecutionManager().RangeCompleteCrossClusterTask(ctx, &persistence.RangeCompleteCrossClusterTaskRequest{
			TargetCluster:        c.targetCluster,
			ExclusiveBeginTaskID: c.ackLevel,
			InclusiveEndTaskID:   nextAckLevel,
			PageSize:             pageSize, // pageSize may or may not be honored
		})
		if err != nil {
			return c.ackLevel, err
		}
		if !persistence.HasMoreRowsToDelete(resp.TasksCompleted, pageSize) {
			break
		}
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

	throttleRetry := backoff.NewThrottleRetry(
		backoff.WithRetryPolicy(persistenceOperationRetryPolicy),
		backoff.WithRetryableError(persistence.IsBackgroundTransientError),
	)
	err := throttleRetry.Do(context.Background(), op)
	if err != nil {
		return nil, false, err
	}

	return response.Tasks, len(response.NextPageToken) != 0, nil
}

func (c *crossClusterQueueProcessorBase) pollTasks(
	ctx context.Context,
) []*types.CrossClusterTaskRequest {

	batchSize := c.shard.GetConfig().CrossClusterTaskFetchBatchSize(c.shard.GetShardID())
	tasks := make([]task.CrossClusterTask, 0, batchSize)
	iter := c.readyForPollTasks.Iter()
	for entry := range iter.Entries() {
		crossClusterTask, ok := entry.Value.(task.CrossClusterTask)
		if !ok {
			panic("received non cross cluster task")
		}

		tasks = append(tasks, crossClusterTask)
		if len(tasks) >= batchSize {
			break
		}
	}
	iter.Close()

	result := []*types.CrossClusterTaskRequest{}
	for _, task := range tasks {
		if ctx.Err() != nil {
			break
		}
		request, err := task.GetCrossClusterRequest()
		if err != nil {
			if !task.IsValid() {
				c.readyForPollTasks.Remove(task.GetTaskID())
				if _, err := c.submitTask(task); err != nil {
					break
				}
			} else {
				c.logger.Error("Failed to get cross cluster request", tag.Error(err))
			}
		} else if request != nil {
			result = append(result, request)
		} else {
			// if request is nil, nothing need to be done for the task,
			// task already acked in GetCrossClusterRequest()
			c.readyForPollTasks.Remove(task.GetTaskID())
		}
	}
	return result
}

func (c *crossClusterQueueProcessorBase) recordResponse(
	response *types.CrossClusterTaskResponse,
) error {

	taskID := response.GetTaskID()
	queueTask, ok := c.readyForPollTasks.Get(taskID)
	if !ok {
		// if task not found, should not return error, it's expected for duplicated response
		return nil
	}
	crossClusterTask, ok := queueTask.(task.CrossClusterTask)
	if !ok {
		panic("Encountered non cross cluster task.")
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
	c.readyForPollTasks.Remove(taskID)
	_, err := c.submitTask(crossClusterTask)
	return err
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

func (c *crossClusterQueueProcessorBase) backoffAllQueueCollections() {
	c.processingLock.Lock()
	defer c.processingLock.Unlock()

	for _, queueCollection := range c.processingQueueCollections {
		level := queueCollection.Level()
		c.setupBackoffTimerLocked(level)
	}
}

func (c *crossClusterQueueProcessorBase) setupBackoffTimer(level int) {
	c.processingLock.Lock()
	defer c.processingLock.Unlock()

	c.setupBackoffTimerLocked(level)
}

func (c *crossClusterQueueProcessorBase) setupBackoffTimerLocked(level int) {
	c.metricsScope.IncCounter(metrics.ProcessingQueueThrottledCounter)
	c.logger.Info("Throttled processing queue", tag.QueueLevel(level))

	if _, ok := c.backoffTimer[level]; ok {
		// there's an existing backoff timer, no-op
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
		notification.resultNotificationCh <- actionResultNotification{
			result: &ActionResult{
				ActionType: ActionTypeGetTasks,
				GetTasksResult: &GetTasksResult{
					TaskRequests: c.pollTasks(notification.ctx),
				},
			},
		}
		close(notification.resultNotificationCh)
	case ActionTypeUpdateTask:
		actionAttr := notification.action.UpdateTaskAttributes
		var actionErr error
		for _, response := range actionAttr.TaskResponses {
			if err := c.recordResponse(response); err != nil {
				select {
				case <-c.shutdownCh:
					actionErr = errProcessorShutdown
					break
				default:
				}

				// we ignore all other errors here (error indicating the response itself is invalid)
				// logging is already done in the updateTask method. task is sill available
				// for polling and will be dispatched again to target cluster.
			}
			if ctxErr := notification.ctx.Err(); ctxErr != nil {
				actionErr = ctxErr
				break
			}
		}
		notification.resultNotificationCh <- actionResultNotification{
			result: &ActionResult{
				ActionType:       ActionTypeUpdateTask,
				UpdateTaskResult: &UpdateTasksResult{},
			},
			err: actionErr,
		}
		close(notification.resultNotificationCh)
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
				c.setupBackoffTimerLocked(level)
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
		for _, taskInfo := range taskInfos {
			if !domainFilter.Filter(taskInfo.GetDomainID()) {
				continue
			}
			taskID := taskInfo.GetTaskID()
			task := c.taskInitializer(taskInfo).(task.CrossClusterTask)
			tasks[newCrossClusterTaskKey(taskID)] = task
			if task.IsReadyForPoll() {
				c.readyForPollTasks.Put(taskID, task)
			} else {
				// currently all tasks will be ready for poll after loaded,
				// so this path will never be executed
				// but logically this case is possbile
				if _, err := c.submitTask(task); err != nil {
					// queue shutting down, return
					return
				}
			}

			newTaskID = taskID
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

func (c *crossClusterQueueProcessorBase) notifyDomainFailover(
	domains map[string]struct{},
) {
	c.failoverNotifyCh <- domains
}

// validateOutstandingTasks will be triggered when the cluster received a domain failover event
func (c *crossClusterQueueProcessorBase) validateOutstandingTasks(
	failoverDomains map[string]struct{},
) {
	if failoverDomains == nil {
		failoverDomains = make(map[string]struct{})
	}

DrainFailoverChLoop:
	for {
		select {
		case domains := <-c.failoverNotifyCh:
			for domain := range domains {
				failoverDomains[domain] = struct{}{}
			}
		default:
			break DrainFailoverChLoop
		}
	}

	var invalidTasks []task.CrossClusterTask
	iter := c.readyForPollTasks.Iter()
	for entry := range iter.Entries() {
		crossClusterTask, ok := entry.Value.(task.CrossClusterTask)
		if !ok {
			panic("Encountered non cross cluster task.")
		}

		_, sourceDomainFailover := failoverDomains[crossClusterTask.GetDomainID()]
		_, targetDomainFailover := failoverDomains[crossClusterTask.GetInfo().(*persistence.CrossClusterTaskInfo).TargetDomainID]
		if !sourceDomainFailover && !targetDomainFailover {
			continue
		}

		if !crossClusterTask.IsValid() {
			// note we can't modify the ordered map before iter is closed, otherwise it will cause a deadlock
			invalidTasks = append(invalidTasks, crossClusterTask)
		}
	}
	iter.Close()

	for _, invalidTask := range invalidTasks {
		c.readyForPollTasks.Remove(invalidTask.GetTaskID())
		_, err := c.submitTask(invalidTask)
		if err != nil {
			// the queue is shutting down
			return
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
	return &queueProcessorOptions{
		BatchSize:                            config.CrossClusterTaskBatchSize,
		DeleteBatchSize:                      config.CrossClusterTaskDeleteBatchSize,
		MaxPollRPS:                           config.CrossClusterSourceProcessorMaxPollRPS,
		MaxPollInterval:                      config.CrossClusterSourceProcessorMaxPollInterval,
		MaxPollIntervalJitterCoefficient:     config.CrossClusterSourceProcessorMaxPollIntervalJitterCoefficient,
		UpdateAckInterval:                    config.CrossClusterSourceProcessorUpdateAckInterval,
		UpdateAckIntervalJitterCoefficient:   config.CrossClusterSourceProcessorUpdateAckIntervalJitterCoefficient,
		RedispatchInterval:                   config.ActiveTaskRedispatchInterval,
		RedispatchIntervalJitterCoefficient:  config.TaskRedispatchIntervalJitterCoefficient,
		MaxRedispatchQueueSize:               config.CrossClusterSourceProcessorMaxRedispatchQueueSize,
		PollBackoffInterval:                  config.QueueProcessorPollBackoffInterval,
		PollBackoffIntervalJitterCoefficient: config.QueueProcessorPollBackoffIntervalJitterCoefficient,
		EnableValidator:                      dynamicconfig.GetBoolPropertyFn(false),
		EnableSplit:                          dynamicconfig.GetBoolPropertyFn(false),
		EnablePersistQueueStates:             dynamicconfig.GetBoolPropertyFn(true),
		EnableLoadQueueStates:                dynamicconfig.GetBoolPropertyFn(true),
		MaxPendingTaskSize:                   config.CrossClusterSourceProcessorMaxPendingTaskSize,
		MetricScope:                          metrics.CrossClusterQueueProcessorScope,
	}
}
