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
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/quotas"
)


type timerQueueTaskProcessor struct {
	scope            int
	shard            ShardContext
	historyService   *historyEngineImpl
	cache            *historyCache
	executionManager persistence.ExecutionManager
	status           int32
	shutdownWG       sync.WaitGroup
	shutdownCh       chan struct{}
	tasksCh          chan *persistence.TimerTaskInfo
	config           *Config
	logger           log.Logger
	metricsClient    metrics.Client
	timerFiredCount  uint64
	timerProcessor   timerProcessor
	timerGate        TimerGate
	timeSource       clock.TimeSource
	rateLimiter      quotas.Limiter
	retryPolicy      backoff.RetryPolicy

	// worker coroutines notification
	workerNotificationChans []chan struct{}
	// duplicate numOfWorker from config.TimerTaskWorkerCount for dynamic config works correctly
	numOfWorker int

	lastPollTime time.Time

	// timer notification
	newTimerCh  chan struct{}
	newTimeLock sync.Mutex
	newTime     time.Time
}

func (t *timerQueueTaskProcessor) start (){
	var workerWG sync.WaitGroup
	for i := 0; i < t.numOfWorker; i++ {
		workerWG.Add(1)
		notificationChan := t.workerNotificationChans[i]
		go t.taskWorker(&workerWG, notificationChan)
	}
}

func (t *timerQueueTaskProcessor) taskWorker(
	workerWG *sync.WaitGroup,
	notificationChan chan struct{},
) {
	defer workerWG.Done()

	for {
		select {
		case <-t.shutdownCh:
			return
		case task, ok := <-t.tasksCh:
			if !ok {
				return
			}
			t.processTaskAndAck(notificationChan, task)
		}
	}
}

func (t *timerQueueTaskProcessor) processTaskAndAck(
	notificationChan <-chan struct{},
	task *persistence.TimerTaskInfo,
) {

	var scope int
	var shouldProcessTask bool
	var err error
	startTime := t.timeSource.Now()
	logger := t.initializeLoggerForTask(task)
	attempt := 0
	incAttempt := func() {
		attempt++
		if attempt >= t.config.TimerTaskMaxRetryCount() {
			t.metricsClient.RecordTimer(scope, metrics.TaskAttemptTimer, time.Duration(attempt))
			logger.Error("Critical error processing timer task, retrying.", tag.Error(err), tag.OperationCritical)
		}
	}

FilterLoop:
	for {
		select {
		case <-t.shutdownCh:
			// this must return without ack
			return
		default:
			shouldProcessTask, err = t.timerProcessor.getTaskFilter()(task)
			if err == nil {
				break FilterLoop
			}
			incAttempt()
			time.Sleep(loadDomainEntryForTimerTaskRetryDelay)
		}
	}

	op := func() error {
		scope, err = t.processTaskOnce(notificationChan, task, shouldProcessTask, logger)
		return t.handleTaskError(scope, startTime, notificationChan, err, logger)
	}
	retryCondition := func(err error) bool {
		select {
		case <-t.shutdownCh:
			return false
		default:
			return true
		}
	}

	for {
		select {
		case <-t.shutdownCh:
			// this must return without ack
			return
		default:
			err = backoff.Retry(op, t.retryPolicy, retryCondition)
			if err == nil {
				t.ackTaskOnce(task, scope, shouldProcessTask, startTime, attempt)
				return
			}
			incAttempt()
		}
	}
}

func (t *timerQueueTaskProcessor) processTaskOnce(
	notificationChan <-chan struct{},
	task *persistence.TimerTaskInfo,
	shouldProcessTask bool,
	logger log.Logger,
) (int, error) {

	select {
	case <-notificationChan:
	default:
	}

	startTime := t.timeSource.Now()
	scope, err := t.timerProcessor.process(task, shouldProcessTask)
	if shouldProcessTask {
		t.metricsClient.IncCounter(scope, metrics.TaskRequests)
		t.metricsClient.RecordTimer(scope, metrics.TaskProcessingLatency, time.Since(startTime))
	}

	return scope, err
}

func (t *timerQueueTaskProcessor) handleTaskError(
	scope int,
	startTime time.Time,
	notificationChan <-chan struct{},
	err error,
	logger log.Logger,
) error {

	if err == nil {
		return nil
	}

	if _, ok := err.(*workflow.EntityNotExistsError); ok {
		return nil
	}

	// this is a transient error
	if err == ErrTaskRetry {
		t.metricsClient.IncCounter(scope, metrics.TaskStandbyRetryCounter)
		<-notificationChan
		return err
	}

	if err == ErrTaskDiscarded {
		t.metricsClient.IncCounter(scope, metrics.TaskDiscarded)
		err = nil
	}

	// this is a transient error
	if _, ok := err.(*workflow.DomainNotActiveError); ok {
		if t.timeSource.Now().Sub(startTime) > cache.DomainCacheRefreshInterval {
			t.metricsClient.IncCounter(scope, metrics.TaskNotActiveCounter)
			return nil
		}

		return err
	}

	t.metricsClient.IncCounter(scope, metrics.TaskFailures)

	if _, ok := err.(*persistence.CurrentWorkflowConditionFailedError); ok {
		logger.Error("More than 2 workflow are running.", tag.Error(err), tag.LifeCycleProcessingFailed)
		return nil
	}

	logger.Error("Fail to process task", tag.Error(err), tag.LifeCycleProcessingFailed)
	return err
}

func (t *timerQueueTaskProcessor) ackTaskOnce(
	task *persistence.TimerTaskInfo,
	scope int,
	reportMetrics bool,
	startTime time.Time,
	attempt int,
) {

	t.timerQueueAckMgr.completeTimerTask(task)
	if reportMetrics {
		t.metricsClient.RecordTimer(scope, metrics.TaskAttemptTimer, time.Duration(attempt))
		t.metricsClient.RecordTimer(scope, metrics.TaskLatency, time.Since(startTime))
		t.metricsClient.RecordTimer(
			scope,
			metrics.TaskQueueLatency,
			time.Since(task.GetVisibilityTimestamp()),
		)
	}
	atomic.AddUint64(&t.timerFiredCount, 1)
}

func (t *timerQueueTaskProcessor) initializeLoggerForTask(
	task *persistence.TimerTaskInfo,
) log.Logger {

	logger := t.logger.WithTags(
		tag.ShardID(t.shard.GetShardID()),
		tag.WorkflowDomainID(task.DomainID),
		tag.WorkflowID(task.WorkflowID),
		tag.WorkflowRunID(task.RunID),
		tag.TaskID(task.GetTaskID()),
		tag.FailoverVersion(task.GetVersion()),
		tag.TaskType(task.GetTaskType()),
		tag.WorkflowTimeoutType(int64(task.TimeoutType)),
	)
	logger.Debug(fmt.Sprintf("Processing timer task: %v, type: %v", task.GetTaskID(), task.GetTaskType()))
	return logger
}

