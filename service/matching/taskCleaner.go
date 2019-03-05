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

package matching

import (
	"sync/atomic"
	"time"
)

type taskCleaner struct {
	lock           int64
	db             *taskListDB
	ackLevel       int64
	lastDeleteTime time.Time
	config         *taskListConfig
}

var maxTimeBetweenTaskDeletes = time.Second

// newTaskCleaner returns an instance of a taskCleaner object
// The returned taskCleaner attempts to delete a batch of completed tasks
// from persistence everytime Run() method is called. Internally, cleaner
// maintains the current delete cursor and updates this after successful deletes
//
// In order for the cleaner to actually delete tasks when Run() is called, one of
// two conditions must be met
//  - Size Threshold: More than MaxDeleteBatchSize tasks are pending to be deleted (rough estimation)
//  - Time Threshold: Time since previous delete was attempted exceeds maxTimeBetweenTaskDeletes
//
// Finally, the Run() method is safe to be called from multiple threads. The underlying
// implementation will make sure only one caller executes Run() and others simply bail out
func newTaskCleaner(db *taskListDB, config *taskListConfig) *taskCleaner {
	return &taskCleaner{db: db, config: config}
}

// Run deletes a batch of completed tasks, if its possible to do so
// Only attempts deletion if size or time thresholds are met
func (tc *taskCleaner) Run(ackLevel int64) {
	tc.tryDeleteNextBatch(ackLevel, false)
}

// RunNow deletes a batch of completed tasks if its possible to do so
// This method attempts deletions without waiting for size/time threshold to be met
func (tc *taskCleaner) RunNow(ackLevel int64) {
	tc.tryDeleteNextBatch(ackLevel, true)
}

func (tc *taskCleaner) tryDeleteNextBatch(ackLevel int64, ignoreTimeCond bool) {
	if !tc.tryLock() {
		return
	}
	defer tc.unlock()
	batchSize := tc.config.MaxTaskDeleteBatchSize()
	if !tc.checkPrecond(ackLevel, batchSize, ignoreTimeCond) {
		return
	}
	tc.lastDeleteTime = time.Now()
	n, err := tc.db.CompleteTasksLessThan(ackLevel, batchSize)
	switch {
	case err != nil:
		return
	case n < batchSize:
		tc.ackLevel = ackLevel
	}
}

func (tc *taskCleaner) checkPrecond(ackLevel int64, batchSize int, ignoreTimeCond bool) bool {
	backlog := ackLevel - tc.ackLevel
	if backlog >= int64(batchSize) {
		return true
	}
	return backlog > 0 && (ignoreTimeCond || time.Now().Sub(tc.lastDeleteTime) > maxTimeBetweenTaskDeletes)
}

func (tc *taskCleaner) tryLock() bool {
	return atomic.CompareAndSwapInt64(&tc.lock, 0, 1)
}

func (tc *taskCleaner) unlock() {
	atomic.StoreInt64(&tc.lock, 0)
}
