// Copyright (c) 2017-2020 Uber Technologies Inc.

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
	"sync"
	"sync/atomic"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/collection"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/service/history/shard"
	"github.com/uber/cadence/service/history/task"
)

type (
	timerTaskKey struct {
		visibilityTimeStamp time.Time
		taskID              int64
	}

	timeTaskReadProgress struct {
		currentQueue  ProcessingQueue
		readLevel     task.Key
		maxReadLevel  task.Key
		nextPageToken []byte
	}

	timerQueueProcessorBase struct {
		shard           shard.Context
		taskProcessor   task.Processor
		redispatchQueue collection.Queue
		timerGate       TimerGate

		options                *queueProcessorOptions
		maxReadLevel           maxReadLevel
		updateTransferAckLevel updateTransferAckLevel
		transferQueueShutdown  transferQueueShutdown
		taskInitializer        task.Initializer

		logger        log.Logger
		metricsClient metrics.Client
		metricsScope  metrics.Scope

		rateLimiter  quotas.Limiter
		lastPollTime time.Time
		status       int32
		shutdownWG   sync.WaitGroup
		shutdownCh   chan struct{}

		// timer notification
		newTimerCh  chan struct{}
		newTimeLock sync.Mutex
		newTime     time.Time

		queueCollectionsLock        sync.RWMutex
		processingQueueCollections  []ProcessingQueueCollection
		processingQueueReadProgress map[int]timeTaskReadProgress
	}
)

func newTimeQueueProcessorBase(
	shard shard.Context,
	processingQueueStates []ProcessingQueueState,
	taskProcessor task.Processor,
	redispatchQueue collection.Queue,
	timerGate TimerGate,
	options *queueProcessorOptions,
	maxReadLevel maxReadLevel,
	updateTransferAckLevel updateTransferAckLevel,
	transferQueueShutdown transferQueueShutdown,
	taskInitializer task.Initializer,
	logger log.Logger,
	metricsClient metrics.Client,
) *timerQueueProcessorBase {
	return &timerQueueProcessorBase{
		shard:           shard,
		taskProcessor:   taskProcessor,
		redispatchQueue: redispatchQueue,
		timerGate:       timerGate,

		options:                options,
		maxReadLevel:           maxReadLevel,
		updateTransferAckLevel: updateTransferAckLevel,
		transferQueueShutdown:  transferQueueShutdown,
		taskInitializer:        taskInitializer,

		logger:        logger.WithTags(tag.ComponentTimerQueue),
		metricsClient: metricsClient,
		metricsScope:  metricsClient.Scope(options.MetricScope),

		rateLimiter: quotas.NewDynamicRateLimiter(
			func() float64 {
				return float64(options.MaxPollRPS())
			},
		),
		lastPollTime: time.Time{},
		status:       common.DaemonStatusInitialized,
		shutdownCh:   make(chan struct{}),
		newTimerCh:   make(chan struct{}, 1),

		processingQueueCollections: newProcessingQueueCollections(
			processingQueueStates,
			logger,
			metricsClient,
		),
		processingQueueReadProgress: make(map[int]timeTaskReadProgress),
	}
}

func (t *timerQueueProcessorBase) Start() {
	if !atomic.CompareAndSwapInt32(&t.status, common.DaemonStatusInitialized, common.DaemonStatusStarted) {
		return
	}

	t.logger.Info("", tag.LifeCycleStarting)
	defer t.logger.Info("", tag.LifeCycleStarted)

	t.notifyNewTimer(time.Time{})

	t.shutdownWG.Add(1)
	go t.processorPump()
}

func (t *timerQueueProcessorBase) Stop() {
	if !atomic.CompareAndSwapInt32(&t.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}

	t.logger.Info("", tag.LifeCycleStopping)
	defer t.logger.Info("", tag.LifeCycleStopped)

	t.timerGate.Close()
	close(t.shutdownCh)

	if success := common.AwaitWaitGroup(&t.shutdownWG, time.Minute); !success {
		t.logger.Warn("", tag.LifeCycleStopTimedout)
	}
}

func (t *timerQueueProcessorBase) processorPump() {
	defer t.shutdownWG.Done()

	pollTimer := time.NewTimer(backoff.JitDuration(
		t.options.MaxPollInterval(),
		t.options.MaxPollIntervalJitterCoefficient(),
	))
	defer pollTimer.Stop()

	updateAckTimer := time.NewTimer(backoff.JitDuration(
		t.options.UpdateAckInterval(),
		t.options.UpdateAckIntervalJitterCoefficient(),
	))
	defer updateAckTimer.Stop()

	redispatchTimer := time.NewTimer(backoff.JitDuration(
		t.options.RedispatchInterval(),
		t.options.RedispatchIntervalJitterCoefficient(),
	))
	defer redispatchTimer.Stop()

	splitQueueTimer := time.NewTimer(backoff.JitDuration(
		t.options.SplitQueueInterval(),
		t.options.SplitQueueIntervalJitterCoefficient(),
	))
	defer splitQueueTimer.Stop()

processorPumpLoop:
	for {
		select {
		case <-t.shutdownCh:
			break processorPumpLoop
		case <-t.timerGate.FireChan():
			if t.redispatchQueue.Len() <= t.options.MaxRedispatchQueueSize() {
				if lookAheadTimer, _ := t.processBatch(); lookAheadTimer != nil {
					t.timerGate.Update(lookAheadTimer.VisibilityTimestamp)
				}
				continue
			}

			RedispatchTasks(
				t.redispatchQueue,
				t.taskProcessor,
				t.logger,
				t.metricsScope,
				t.shutdownCh,
			)
			t.notifyNewTimer(time.Time{})
		case <-pollTimer.C:
			pollTimer.Reset(backoff.JitDuration(
				t.options.MaxPollInterval(),
				t.options.MaxPollIntervalJitterCoefficient(),
			))
			if t.lastPollTime.Add(t.options.MaxPollInterval()).Before(t.shard.GetTimeSource().Now()) {
				if lookAheadTimer, _ := t.processBatch(); lookAheadTimer != nil {
					t.timerGate.Update(lookAheadTimer.VisibilityTimestamp)
				}
			}
		case <-updateAckTimer.C:
			updateAckTimer.Reset(backoff.JitDuration(
				t.options.UpdateAckInterval(),
				t.options.UpdateAckIntervalJitterCoefficient(),
			))
			processFinished, err := t.updateAckLevel()
			if err == shard.ErrShardClosed || (err == nil && processFinished) {
				go t.Stop()
				break processorPumpLoop
			}
		case <-t.newTimerCh:
			t.newTimeLock.Lock()
			newTime := t.newTime
			t.newTime = time.Time{}
			t.newTimeLock.Unlock()

			// New Timer has arrived.
			t.metricsScope.IncCounter(metrics.NewTimerNotifyCounter)
			t.timerGate.Update(newTime)
		case <-redispatchTimer.C:
			redispatchTimer.Reset(backoff.JitDuration(
				t.options.RedispatchInterval(),
				t.options.RedispatchIntervalJitterCoefficient(),
			))
			RedispatchTasks(
				t.redispatchQueue,
				t.taskProcessor,
				t.logger,
				t.metricsScope,
				t.shutdownCh,
			)
		case <-splitQueueTimer.C:
			splitQueueTimer.Reset(backoff.JitDuration(
				t.options.SplitQueueInterval(),
				t.options.SplitQueueIntervalJitterCoefficient(),
			))
			t.splitQueue()
		}
	}
}

func (t *timerQueueProcessorBase) processBatch() (*persistence.TimerTaskInfo, error) {
	// TODO
	return nil, nil
}

func (t *timerQueueProcessorBase) updateAckLevel() (bool, error) {
	t.queueCollectionsLock.Lock()
	defer t.queueCollectionsLock.Unlock()

	// TODO
	return false, nil
}

func (t *timerQueueProcessorBase) splitQueue() {
	t.queueCollectionsLock.Lock()
	defer t.queueCollectionsLock.Unlock()

	t.processingQueueCollections = splitProcessingQueueCollection(
		t.processingQueueCollections,
		t.options.QueueSplitPolicy,
	)
}

func (t *timerQueueProcessorBase) notifyNewTimers(
	timerTasks []persistence.Task,
) {
	if len(timerTasks) == 0 {
		return
	}

	isActive := t.options.MetricScope == metrics.TimerActiveQueueProcessorScope

	minNewTime := timerTasks[0].GetVisibilityTimestamp()
	for _, timerTask := range timerTasks {
		ts := timerTask.GetVisibilityTimestamp()
		if ts.Before(minNewTime) {
			minNewTime = ts
		}

		taskScopeIdx := task.GetTimerTaskMetricScope(
			timerTask.GetType(),
			isActive,
		)
		t.metricsClient.IncCounter(taskScopeIdx, metrics.NewTimerCounter)
	}

	t.notifyNewTimer(minNewTime)
}

func (t *timerQueueProcessorBase) notifyNewTimer(
	newTime time.Time,
) {
	t.newTimeLock.Lock()
	defer t.newTimeLock.Unlock()

	if t.newTime.IsZero() || newTime.Before(t.newTime) {
		t.newTime = newTime
		select {
		case t.newTimerCh <- struct{}{}:
			// Notified about new time.
		default:
			// Channel "full" -> drop and move on, this will happen only if service is in high load.
		}
	}
}

func newTimerTaskKey(
	visibilityTimestamp time.Time,
	taskID int64,
) task.Key {
	return &timerTaskKey{
		visibilityTimeStamp: visibilityTimestamp,
		taskID:              taskID,
	}
}

func (k *timerTaskKey) Less(
	key task.Key,
) bool {
	timerKey := key.(*timerTaskKey)
	if k.visibilityTimeStamp.Equal(timerKey.visibilityTimeStamp) {
		return k.taskID < timerKey.taskID
	}
	return k.visibilityTimeStamp.Before(timerKey.visibilityTimeStamp)
}
