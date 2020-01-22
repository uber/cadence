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
	"sync/atomic"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
)

type (
	// PriorityTaskProcessorOptions configs PriorityTaskProcessor
	PriorityTaskProcessorOptions struct {
		WorkerCount int
		RetryPolicy backoff.RetryPolicy
	}

	priorityTaskProcessorImpl struct {
		scheduler PriorityScheduler

		status       int32
		workerWG     sync.WaitGroup
		logger       log.Logger
		metricsScope metrics.Scope
		options      *PriorityTaskProcessorOptions
	}
)

// NewPriorityTaskProcessor creates a new PriorityTaskProcessor
func NewPriorityTaskProcessor(
	scheduler PriorityScheduler,
	logger log.Logger,
	metricsClient metrics.Client,
	options *PriorityTaskProcessorOptions,
) PriorityTaskProcessor {
	return &priorityTaskProcessorImpl{
		scheduler:    scheduler,
		status:       common.DaemonStatusInitialized,
		logger:       logger,
		metricsScope: metricsClient.Scope(metrics.PriorityTaskProcessingScope),
		options:      options,
	}
}

func (p *priorityTaskProcessorImpl) Start() {
	if !atomic.CompareAndSwapInt32(&p.status, common.DaemonStatusInitialized, common.DaemonStatusStarted) {
		return
	}

	p.workerWG.Add(p.options.WorkerCount)
	for i := 0; i < p.options.WorkerCount; i++ {
		go p.taskWorker()
	}
	p.scheduler.Start()
	p.logger.Info("Priority task processor started.")
}

func (p *priorityTaskProcessorImpl) Stop() {
	if !atomic.CompareAndSwapInt32(&p.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}

	p.scheduler.Stop()
	if success := common.AwaitWaitGroup(&p.workerWG, time.Minute); !success {
		p.logger.Warn("Priority task processor timedout on shutdown.")
	}
	p.logger.Info("Priority task processor shutdown.")
}

func (p *priorityTaskProcessorImpl) Submit(task Task, priority int) error {
	p.metricsScope.IncCounter(metrics.PriorityTaskSubmitRequest)
	sw := p.metricsScope.StartTimer(metrics.PriorityTaskSubmitLatency)
	defer sw.Stop()

	return p.scheduler.Schedule(task, priority)
}

func (p *priorityTaskProcessorImpl) taskWorker() {
	defer p.workerWG.Done()

	for {
		task := p.scheduler.Consume()
		if task == nil {
			return
		}

		p.executeTask(task.(Task))
	}
}

func (p *priorityTaskProcessorImpl) executeTask(task Task) {
	sw := p.metricsScope.StartTimer(metrics.PriorityTaskTaskProcessingLatency)
	defer sw.Stop()

	op := func() error {
		if err := task.Execute(); err != nil {
			return task.HandleErr(err)
		}
		return nil
	}

	isRetryable := func(err error) bool {
		if p.isStopped() {
			return false
		}
		return task.RetryErr(err)
	}

	for {
		if p.isStopped() {
			// neither ack or nack here
			return
		}

		if err := backoff.Retry(op, p.options.RetryPolicy, isRetryable); err != nil {
			if !task.RetryErr(err) {
				// encounter an non-retryable error
				task.Nack()
				return
			}
			// retryable error, continue the loop
		} else {
			// no error
			task.Ack()
			return
		}
	}
}

func (p *priorityTaskProcessorImpl) isStopped() bool {
	return atomic.LoadInt32(&p.status) == common.DaemonStatusStopped
}
