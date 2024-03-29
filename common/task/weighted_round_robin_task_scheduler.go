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
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
)

type weightedRoundRobinTaskSchedulerImpl struct {
	sync.RWMutex

	status       int32
	weights      atomic.Value // store the currently used weights
	taskChs      map[int]chan PriorityTask
	shutdownCh   chan struct{}
	notifyCh     chan struct{}
	dispatcherWG sync.WaitGroup
	logger       log.Logger
	metricsScope metrics.Scope
	options      *WeightedRoundRobinTaskSchedulerOptions

	processor Processor
}

const (
	wRRTaskProcessorQueueSize    = 1
	defaultUpdateWeightsInterval = 5 * time.Second
)

var (
	// ErrTaskSchedulerClosed is the error returned when submitting task to a stopped scheduler
	ErrTaskSchedulerClosed = errors.New("task scheduler has already shutdown")
)

// NewWeightedRoundRobinTaskScheduler creates a new WRR task scheduler
func NewWeightedRoundRobinTaskScheduler(
	logger log.Logger,
	metricsClient metrics.Client,
	options *WeightedRoundRobinTaskSchedulerOptions,
) (Scheduler, error) {
	weights, err := common.ConvertDynamicConfigMapPropertyToIntMap(options.Weights())
	if err != nil {
		return nil, err
	}

	if len(weights) == 0 {
		return nil, errors.New("weight is not specified in the scheduler option")
	}

	scheduler := &weightedRoundRobinTaskSchedulerImpl{
		status:       common.DaemonStatusInitialized,
		taskChs:      make(map[int]chan PriorityTask),
		shutdownCh:   make(chan struct{}),
		notifyCh:     make(chan struct{}, 1),
		logger:       logger,
		metricsScope: metricsClient.Scope(metrics.TaskSchedulerScope),
		options:      options,
		processor: NewParallelTaskProcessor(
			logger,
			metricsClient,
			&ParallelTaskProcessorOptions{
				QueueSize:   wRRTaskProcessorQueueSize,
				WorkerCount: options.WorkerCount,
				RetryPolicy: options.RetryPolicy,
			},
		),
	}
	scheduler.weights.Store(weights)

	return scheduler, nil
}

func (w *weightedRoundRobinTaskSchedulerImpl) Start() {
	if !atomic.CompareAndSwapInt32(&w.status, common.DaemonStatusInitialized, common.DaemonStatusStarted) {
		return
	}

	w.processor.Start()

	w.dispatcherWG.Add(w.options.DispatcherCount)
	for i := 0; i != w.options.DispatcherCount; i++ {
		go w.dispatcher()
	}
	go w.updateWeights()

	w.logger.Info("Weighted round robin task scheduler started.")
}

func (w *weightedRoundRobinTaskSchedulerImpl) Stop() {
	if !atomic.CompareAndSwapInt32(&w.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}

	close(w.shutdownCh)

	w.processor.Stop()

	w.RLock()
	for _, taskCh := range w.taskChs {
		drainAndNackPriorityTask(taskCh)
	}
	w.RUnlock()

	if success := common.AwaitWaitGroup(&w.dispatcherWG, time.Minute); !success {
		w.logger.Warn("Weighted round robin task scheduler timedout on shutdown.")
	}

	w.logger.Info("Weighted round robin task scheduler shutdown.")
}

func (w *weightedRoundRobinTaskSchedulerImpl) Submit(task PriorityTask) error {
	w.metricsScope.IncCounter(metrics.PriorityTaskSubmitRequest)
	sw := w.metricsScope.StartTimer(metrics.PriorityTaskSubmitLatency)
	defer sw.Stop()

	if w.isStopped() {
		return ErrTaskSchedulerClosed
	}

	taskCh, err := w.getOrCreateTaskChan(task.Priority())
	if err != nil {
		return err
	}

	select {
	case taskCh <- task:
		w.notifyDispatcher()
		if w.isStopped() {
			drainAndNackPriorityTask(taskCh)
		}
		return nil
	case <-w.shutdownCh:
		return ErrTaskSchedulerClosed
	}
}

func (w *weightedRoundRobinTaskSchedulerImpl) TrySubmit(
	task PriorityTask,
) (bool, error) {
	if w.isStopped() {
		return false, ErrTaskSchedulerClosed
	}

	taskCh, err := w.getOrCreateTaskChan(task.Priority())
	if err != nil {
		return false, err
	}

	select {
	case taskCh <- task:
		w.metricsScope.IncCounter(metrics.PriorityTaskSubmitRequest)
		if w.isStopped() {
			drainAndNackPriorityTask(taskCh)
		} else {
			w.notifyDispatcher()
		}
		return true, nil
	case <-w.shutdownCh:
		return false, ErrTaskSchedulerClosed
	default:
		return false, nil
	}
}

func (w *weightedRoundRobinTaskSchedulerImpl) dispatcher() {
	defer w.dispatcherWG.Done()

	outstandingTasks := false
	taskChs := make(map[int]chan PriorityTask)

	for {
		if !outstandingTasks {
			// if no task is dispatched in the last round,
			// wait for a notification
			w.logger.Debug("Weighted round robin task scheduler is waiting for new task notification because there was no task dispatched in the last round.")
			select {
			case <-w.notifyCh:
				// block until there's a new task
				w.logger.Debug("Weighted round robin task scheduler got notification so will check for new tasks.")
			case <-w.shutdownCh:
				return
			}
		}

		outstandingTasks = false
		w.updateTaskChs(taskChs)
		weights := w.getWeights()
		for priority, taskCh := range taskChs {
			count, ok := weights[priority]
			if !ok {
				w.logger.Error("weights not found for task priority", tag.Dynamic("priority", priority), tag.Dynamic("weights", weights))
				continue
			}
		Submit_Loop:
			for i := 0; i < count; i++ {
				select {
				case task := <-taskCh:
					// dispatched at least one task in this round
					outstandingTasks = true

					if err := w.processor.Submit(task); err != nil {
						w.logger.Error("fail to submit task to processor", tag.Error(err))
						task.Nack()
					}
				case <-w.shutdownCh:
					return
				default:
					// if no task, don't block. Skip to next priority
					break Submit_Loop
				}
			}
		}
	}
}

func (w *weightedRoundRobinTaskSchedulerImpl) getOrCreateTaskChan(priority int) (chan PriorityTask, error) {
	if _, ok := w.getWeights()[priority]; !ok {
		return nil, fmt.Errorf("unknown task priority: %v", priority)
	}

	w.RLock()
	if taskCh, ok := w.taskChs[priority]; ok {
		w.RUnlock()
		return taskCh, nil
	}
	w.RUnlock()

	w.Lock()
	defer w.Unlock()
	if taskCh, ok := w.taskChs[priority]; ok {
		return taskCh, nil
	}
	taskCh := make(chan PriorityTask, w.options.QueueSize)
	w.taskChs[priority] = taskCh
	return taskCh, nil
}

func (w *weightedRoundRobinTaskSchedulerImpl) updateTaskChs(taskChs map[int]chan PriorityTask) {
	w.RLock()
	defer w.RUnlock()

	for priority, taskCh := range w.taskChs {
		if _, ok := taskChs[priority]; !ok {
			taskChs[priority] = taskCh
		}
	}
}

func (w *weightedRoundRobinTaskSchedulerImpl) notifyDispatcher() {
	select {
	case w.notifyCh <- struct{}{}:
		// sent a notification to the dispatcher
	default:
		// do not block if there's already a notification
	}
}

func (w *weightedRoundRobinTaskSchedulerImpl) getWeights() map[int]int {
	return w.weights.Load().(map[int]int)
}

func (w *weightedRoundRobinTaskSchedulerImpl) updateWeights() {
	ticker := time.NewTicker(defaultUpdateWeightsInterval)
	for {
		select {
		case <-ticker.C:
			weights, err := common.ConvertDynamicConfigMapPropertyToIntMap(w.options.Weights())
			if err != nil {
				w.logger.Error("failed to update weight for round robin task scheduler", tag.Error(err))
			} else {
				w.weights.Store(weights)
			}
		case <-w.shutdownCh:
			ticker.Stop()
			return
		}
	}
}

func (w *weightedRoundRobinTaskSchedulerImpl) isStopped() bool {
	return atomic.LoadInt32(&w.status) == common.DaemonStatusStopped
}

func drainAndNackPriorityTask(taskCh <-chan PriorityTask) {
	for {
		select {
		case task := <-taskCh:
			task.Nack()
		default:
			return
		}
	}
}
