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

package tasklist

import (
	"time"

	"sync"

	"sync/atomic"

	"github.com/uber-common/bark"
	"github.com/uber/cadence/common/logging"
	"github.com/uber/cadence/common/metrics"
	p "github.com/uber/cadence/common/persistence"
)

type (
	// Scavenger is the type that holds the state for task list scavenger daemon
	Scavenger struct {
		jobQ         *jobQueue       // Queue of jobs for workers to pick up and execute - each job corresponds to task list
		deferredJobQ *threadSafeList // Queue of deferred jobs to be picked up and executed later (for fairness)
		db           p.TaskManager
		metrics      metrics.Client
		logger       bark.Logger
		outstanding  int64
		signalC      chan struct{}
		stats        stats
		started      int64
		stopped      int64
		stopC        chan struct{}
		stopWG       sync.WaitGroup
	}

	taskListKey struct {
		DomainID string
		Name     string
		TaskType int
	}

	taskListState struct {
		rangeID     int64
		lastUpdated time.Time
	}

	stats struct {
		tasklist struct {
			nProcessed int64
			nDeleted   int64
		}
		task struct {
			nProcessed int64
			nDeleted   int64
		}
	}
)

var (
	taskListBatchSize   = 32                // maximum number of task list we process concurrently
	taskBatchSize       = 16                // number of tasks we read from persistence in one call
	maxTasksPerJob      = 256               // maximum number of tasks we process for a task list as part of a single job
	nWorkers            = taskListBatchSize // number of go routines processing task lists
	taskListGracePeriod = 48 * time.Hour    // amount of time a task list has to be idle before it becomes a candidate for deletion
	epochStartTime      = time.Unix(0, 0)
)

// NewScavenger returns an instance of task list scavenger daemon
// The Scavenger can be started by calling the Start() method on the
// returned object. Calling the Start() method will result in one
// complete iteration over all of the task lists in the system. For
// each task list, the scavenger will attempt
//  - deletion of expired tasks in the task lists
//  - deletion of task list itself, if there are no tasks and the task list hasn't been updated for a grace period
//
// The scavenger will retry on all persistence errors infinitely and will only stop under
// two conditions
//  - either all task lists are processed successfully (or)
//  - Stop() method is called to stop the scavenger
func NewScavenger(db p.TaskManager, metrics metrics.Client, logger bark.Logger) *Scavenger {
	stopC := make(chan struct{})
	return &Scavenger{
		db:           db,
		metrics:      metrics,
		logger:       logger,
		signalC:      make(chan struct{}, 1),
		stopC:        stopC,
		jobQ:         newJobQueue(taskListBatchSize, stopC),
		deferredJobQ: newThreadSafeList(),
	}
}

// Start starts the scavenger
func (s *Scavenger) Start() {
	if !atomic.CompareAndSwapInt64(&s.started, 0, 1) {
		return
	}
	s.logger.Info("Tasklist scavenger starting")
	s.stopWG.Add(nWorkers)
	for i := 0; i < nWorkers; i++ {
		go s.worker()
	}
	s.stopWG.Add(1)
	go s.run()
	s.metrics.IncCounter(metrics.TaskListScavengerScope, metrics.StartedCount)
	s.logger.Info("Tasklist scavenger started")
}

// Stop stops the scavenger
func (s *Scavenger) Stop() {
	if !atomic.CompareAndSwapInt64(&s.stopped, 0, 1) {
		return
	}
	s.metrics.IncCounter(metrics.TaskListScavengerScope, metrics.StoppedCount)
	s.logger.Info("Tasklist scavenger stopping")
	close(s.stopC)
	s.stopWG.Wait()
	s.logger.Info("Tasklist scavenger stopped")
}

// Alive returns true if the scavenger is still running
func (s *Scavenger) Alive() bool {
	select {
	case <-s.stopC:
		return false
	default:
		return true
	}
}

// run does a single run over all task lists
func (s *Scavenger) run() {
	defer func() {
		s.emitStats()
		go s.Stop()
		s.stopWG.Done()
	}()

	var pageToken []byte

	// Step #1: process each task list once
	for {
		resp, err := s.listTaskList(taskListBatchSize, pageToken)
		if err != nil {
			s.logger.WithFields(bark.Fields{logging.TagErr: err}).Error("listTaskList error")
			return
		}

		atomic.AddInt64(&s.stats.tasklist.nProcessed, int64(len(resp.Items)))
		for _, item := range resp.Items {
			atomic.AddInt64(&s.outstanding, 1)
			if !s.jobQ.add(newJob(&item)) {
				return
			}
		}

		pageToken = resp.NextPageToken
		if pageToken == nil {
			break
		}
	}

	if !s.waitForSignal() { // wait for atleast one job to finish
		return
	}

	// Step #2: When some task lists have lot of work to do, we avoid doing
	// too much work in one shot (for fairness). After doing some amount of
	// work, the task list processor adds the task list to a defferedQ which
	// is picked up for processing at a later time. Here, we process all
	// entries in the defferedQ
	for atomic.LoadInt64(&s.outstanding) > 0 || s.deferredJobQ.len() > 0 {
		for e := s.deferredJobQ.remove(); e != nil; e = s.deferredJobQ.remove() {
			atomic.AddInt64(&s.outstanding, 1)
			if !s.jobQ.add(e.Value.(*job)) {
				return
			}
		}
		if !s.waitForSignal() { // wait for one job to finish
			return
		}
	}
}

// worker runs in a separate go routine and processes a single task list
func (s *Scavenger) worker() {
	defer func() {
		s.stopWG.Done()
	}()

	for s.Alive() {
		job, ok := s.jobQ.remove()
		if !ok {
			s.signal()
			return
		}
		//fmt.Println("worker got new job", job.input.Name)
		status, err := s.deleteHandler(&job.input.taskListKey, job.input.taskListState)
		if err != nil {
			s.signal()
			return
		}
		if status == handlerStatusDefer {
			// we have more work to do but we intentionally limited the amount of
			// tasks we processed to maxTasksPerJob for fairness among task lists
			// Add this job to the list of deferredJobQ to be picked up later
			s.deferredJobQ.add(job)
		}
		s.signal()
	}
}

func (s *Scavenger) emitStats() {
	s.metrics.UpdateGauge(metrics.TaskListScavengerScope, metrics.TaskProcessedCount, float64(s.stats.task.nProcessed))
	s.metrics.UpdateGauge(metrics.TaskListScavengerScope, metrics.TaskDeletedCount, float64(s.stats.task.nDeleted))
	s.metrics.UpdateGauge(metrics.TaskListScavengerScope, metrics.TaskListProcessedCount, float64(s.stats.tasklist.nProcessed))
	s.metrics.UpdateGauge(metrics.TaskListScavengerScope, metrics.TaskListDeletedCount, float64(s.stats.tasklist.nDeleted))
}

func (s *Scavenger) signal() {
	atomic.AddInt64(&s.outstanding, -1)
	select {
	case s.signalC <- struct{}{}:
	default:
	}
}

func (s *Scavenger) waitForSignal() bool {
	select {
	case <-s.signalC:
		return true
	case <-s.stopC:
		return false
	}
}
