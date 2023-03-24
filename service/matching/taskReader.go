// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package matching

import (
	"context"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

var epochStartTime = time.Unix(0, 0)

type (
	taskReader struct {
		taskListID      *taskListID
		config          *taskListConfig
		db              *taskListDB
		taskWriter      *taskWriter
		taskGC          *taskGC
		taskAckManagers map[string]messaging.AckManager
		taskBuffer      map[string]chan *persistence.TaskInfo
		notifyC         map[string]chan struct{}
		// The cancel objects are to cancel the ratelimiter Wait in dispatchBufferedTasks. The ideal
		// approach is to use request-scoped contexts and use a unique one for each call to Wait. However
		// in order to cancel it on shutdown, we need a new goroutine for each call that would wait on
		// the shutdown channel. To optimize on efficiency, we instead create one and tag it on the struct
		// so the cancel can be called directly on shutdown.
		cancelCtx     context.Context
		cancelFunc    context.CancelFunc
		status        int32
		logger        log.Logger
		scope         metrics.Scope
		throttleRetry *backoff.ThrottleRetry
		handleErr     func(error) error
		onFatalErr    func()
		dispatchTask  func(context.Context, *InternalTask) error
		isVersion2    bool
	}
)

func newTaskReader(
	isolationGroups []string,
	taskListID *taskListID,
	config *taskListConfig,
	db *taskListDB,
	taskWriter *taskWriter,
	taskAckManagers map[string]messaging.AckManager,
	logger log.Logger,
	scope metrics.Scope,
	tlMgr *taskListManagerImpl,
) *taskReader {
	ctx, cancel := context.WithCancel(context.Background())
	taskBuffer := make(map[string]chan *persistence.TaskInfo)
	notifyC := make(map[string]chan struct{})
	for _, isolationGroup := range isolationGroups {
		taskBuffer[isolationGroup] = make(chan *persistence.TaskInfo, config.GetTasksBatchSize()-1)
		notifyC[isolationGroup] = make(chan struct{}, 1)
	}
	taskBuffer[""] = make(chan *persistence.TaskInfo, config.GetTasksBatchSize()-1)
	notifyC[""] = make(chan struct{}, 1)

	return &taskReader{
		taskListID:      taskListID,
		config:          config,
		db:              db,
		taskWriter:      taskWriter,
		taskGC:          newTaskGC(db, config),
		taskAckManagers: taskAckManagers,
		taskBuffer:      taskBuffer,
		notifyC:         notifyC,
		cancelCtx:       ctx,
		cancelFunc:      cancel,
		logger:          logger,
		scope:           scope,
		handleErr:       tlMgr.handleErr,
		onFatalErr:      tlMgr.Stop,
		dispatchTask:    tlMgr.DispatchTask,
		throttleRetry: backoff.NewThrottleRetry(
			backoff.WithRetryPolicy(persistenceOperationRetryPolicy),
			backoff.WithRetryableError(persistence.IsTransientError),
		),
		isVersion2: len(isolationGroups) > 0,
	}
}

func (tr *taskReader) Start() {
	if !atomic.CompareAndSwapInt32(&tr.status, common.DaemonStatusInitialized, common.DaemonStatusStarted) {
		return
	}
	tr.SignalAll()
	for isolationGroup := range tr.notifyC {
		go tr.dispatchBufferedTasks(isolationGroup)
		go tr.getTasksPump(isolationGroup)
	}
	go tr.updateAck()
}

func (tr *taskReader) Stop() {
	if !atomic.CompareAndSwapInt32(&tr.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}
	tr.cancelFunc()
	if err := tr.persistAckLevel(); err != nil {
		tr.logger.Error("Persistent store operation failure",
			tag.StoreOperationUpdateTaskList,
			tag.Error(err))
	}
	for isolationGroup, taskAckManager := range tr.taskAckManagers {
		tr.taskGC.RunNow(isolationGroup, taskAckManager.GetAckLevel())
	}
}

func (tr *taskReader) Signal(isolationGroup string) {
	notifyC, ok := tr.notifyC[isolationGroup]
	if !ok {
		tr.logger.Fatal("")
		return
	}
	select {
	case notifyC <- struct{}{}:
	default: // channel already has an event, don't block
	}
}

func (tr *taskReader) SignalAll() {
	for _, notifyC := range tr.notifyC {
		select {
		case notifyC <- struct{}{}:
		default: // channel already has an event, don't block
		}
	}
}

func (tr *taskReader) updateAck() {
	updateAckTimer := time.NewTimer(tr.config.UpdateAckInterval())
	defer updateAckTimer.Stop()
	for {
		select {
		case <-tr.cancelCtx.Done():
			break
		case <-updateAckTimer.C:
			{
				if err := tr.handleErr(tr.persistAckLevel()); err != nil {
					tr.logger.Error("Persistent store operation failure",
						tag.StoreOperationUpdateTaskList,
						tag.Error(err))
					// keep going as saving ack is not critical
				}
				tr.SignalAll() // periodically signal pump to check persistence for tasks
				updateAckTimer.Reset(tr.config.UpdateAckInterval())
			}
		}
	}
}

func (tr *taskReader) dispatchBufferedTasks(isolationGroup string) {
dispatchLoop:
	for {
		select {
		case taskInfo, ok := <-tr.taskBuffer[isolationGroup]:
			if !ok { // Task list getTasks pump is shutdown
				break dispatchLoop
			}
			task := newInternalTask(taskInfo, tr.completeTask, types.TaskSourceDbBacklog, isolationGroup, "", false, nil)
			for {
				err := tr.dispatchTask(tr.cancelCtx, task)
				if err == nil {
					break
				}
				if err == context.Canceled {
					tr.logger.Info("Tasklist manager context is cancelled, shutting down")
					break dispatchLoop
				}
				// this should never happen unless there is a bug - don't drop the task
				tr.scope.IncCounter(metrics.BufferThrottlePerTaskListCounter)
				tr.logger.Error("taskReader: unexpected error dispatching task", tag.Error(err))
				runtime.Gosched()
			}
		case <-tr.cancelCtx.Done():
			break dispatchLoop
		}
	}
}

func (tr *taskReader) getTasksPump(isolationGroup string) {
getTasksPumpLoop:
	for {
		select {
		case <-tr.cancelCtx.Done():
			break getTasksPumpLoop
		case <-tr.notifyC[isolationGroup]:
			{
				tasks, readLevel, isReadBatchDone, err := tr.getTaskBatch(isolationGroup)
				if err != nil {
					tr.Signal(isolationGroup) // re-enqueue the event
					// TODO: Should we ever stop retrying on db errors?
					continue getTasksPumpLoop
				}

				if len(tasks) == 0 {
					tr.taskAckManagers[isolationGroup].SetReadLevel(readLevel)
					if !isReadBatchDone {
						tr.Signal(isolationGroup)
					}
					continue getTasksPumpLoop
				}

				if !tr.addTasksToBuffer(isolationGroup, tasks) {
					break getTasksPumpLoop
				}
				// There maybe more tasks. We yield now, but signal pump to check again later.
				tr.Signal(isolationGroup)
			}
		}
	}
}

func (tr *taskReader) getTaskBatchWithRange(isolationGroup string, readLevel int64, maxReadLevel int64) ([]*persistence.TaskInfo, error) {
	var response *persistence.GetTasksResponse
	op := func() (err error) {
		response, err = tr.db.GetTasks(isolationGroup, readLevel, maxReadLevel, tr.config.GetTasksBatchSize())
		return
	}
	err := tr.throttleRetry.Do(context.Background(), op)
	if err != nil {
		tr.logger.Error("Persistent store operation failure",
			tag.StoreOperationGetTasks,
			tag.Error(err),
			tag.WorkflowTaskListName(tr.taskListID.name),
			tag.WorkflowTaskListType(tr.taskListID.taskType))
		return nil, err
	}
	return response.Tasks, nil
}

func (tr *taskReader) getTaskBatch(isolationGroup string) ([]*persistence.TaskInfo, int64, bool, error) {
	taskAckManager := tr.taskAckManagers[isolationGroup]
	readLevel := taskAckManager.GetReadLevel()
	maxReadLevel := tr.taskWriter.GetMaxReadLevel()

	var tasks []*persistence.TaskInfo
	// counter i is used to break and let caller check whether tasklist is still alive and need resume read.
	for i := 0; i < 10 && readLevel < maxReadLevel; i++ {
		upper := readLevel + tr.config.RangeSize
		if upper > maxReadLevel {
			upper = maxReadLevel
		}
		tasks, err := tr.getTaskBatchWithRange(isolationGroup, readLevel, upper)
		if err != nil {
			return nil, readLevel, true, err
		}
		// return as long as it grabs any tasks
		if len(tasks) > 0 {
			return tasks, upper, true, nil
		}
		readLevel = upper
	}
	return tasks, readLevel, readLevel == maxReadLevel, nil // caller will update readLevel when no task grabbed
}

func (tr *taskReader) addTasksToBuffer(isolationGroup string, tasks []*persistence.TaskInfo) bool {
	taskAckManager := tr.taskAckManagers[isolationGroup]
	now := time.Now()
	for _, t := range tasks {
		if isTaskExpired(t, now) {
			tr.scope.IncCounter(metrics.ExpiredTasksPerTaskListCounter)
			// Also increment readLevel for expired tasks otherwise it could result in
			// looping over the same tasks if all tasks read in the batch are expired
			taskAckManager.SetReadLevel(t.TaskID)
			continue
		}
		if !tr.addSingleTaskToBuffer(isolationGroup, t) {
			return false // we are shutting down the task list
		}
	}
	return true
}

func (tr *taskReader) addSingleTaskToBuffer(isolationGroup string, task *persistence.TaskInfo) bool {
	taskAckManager := tr.taskAckManagers[isolationGroup]
	taskBuffer := tr.taskBuffer[isolationGroup] // TODO: consider re-calculating the isolation group of the task
	err := taskAckManager.ReadItem(task.TaskID)
	if err != nil {
		tr.logger.Fatal("critical bug when adding item to ackManager", tag.Error(err))
	}
	select {
	case taskBuffer <- task:
		return true
	case <-tr.cancelCtx.Done():
		return false
	}
}

func (tr *taskReader) persistAckLevel() error {
	ackLevel := int64(0)
	ackLevels := make(map[string]int64)
	for isolationGroup, taskAckManager := range tr.taskAckManagers {
		ackLevels[isolationGroup] = taskAckManager.GetAckLevel()
	}
	maxReadLevel := tr.taskWriter.GetMaxReadLevel()
	scope := tr.scope.Tagged(getTaskListTypeTag(tr.taskListID.taskType))
	// note: this metrics is only an estimation for the lag. taskID in DB may not be continuous,
	// especially when task list ownership changes.
	scope.UpdateGauge(metrics.TaskLagPerTaskListGauge, float64(maxReadLevel-ackLevel))

	if !tr.isVersion2 {
		ackLevel = ackLevels[""]
		ackLevels = nil
	}
	return tr.db.UpdateState(ackLevel, ackLevels)
}

func (tr *taskReader) completeTask(sourceGroup string, task *persistence.TaskInfo, err error) {
	if err != nil {
		// failed to start the task.
		// We cannot just remove it from persistence because then it will be lost.
		// We handle this by writing the task back to persistence with a higher taskID.
		// This will allow subsequent tasks to make progress, and hopefully by the time this task is picked-up
		// again the underlying reason for failing to start will be resolved.
		// Note that RecordTaskStarted only fails after retrying for a long time, so a single task will not be
		// re-written to persistence frequently.
		targetGroup := sourceGroup
		op := func() error {
			wf := &types.WorkflowExecution{WorkflowID: task.WorkflowID, RunID: task.RunID}
			_, err := tr.taskWriter.appendTask(wf, task, targetGroup)
			return err
		}
		err = tr.throttleRetry.Do(context.Background(), op)
		if err != nil {
			// OK, we also failed to write to persistence.
			// This should only happen in very extreme cases where persistence is completely down.
			// We still can't lose the old task so we just unload the entire task list
			tr.logger.Error("Failed to complete task", tag.Error(err))
			tr.onFatalErr()
			return
		}
		tr.Signal(targetGroup)
	}
	ackLevel := tr.taskAckManagers[sourceGroup].AckItem(task.TaskID)
	tr.taskGC.Run(sourceGroup, ackLevel)
}

func isTaskExpired(t *persistence.TaskInfo, now time.Time) bool {
	return t.Expiry.After(epochStartTime) && time.Now().After(t.Expiry)
}
