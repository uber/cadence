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
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/uber/cadence/common/log/tag"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/service/worker/scanner/executor"
)

type handlerStatus = executor.TaskStatus

const (
	handlerStatusDone  = executor.TaskStatusDone
	handlerStatusErr   = executor.TaskStatusErr
	handlerStatusDefer = executor.TaskStatusDefer
)

const scannerTaskListPrefix = "cadence-sys-tl-scanner"

// deleteHandler handles deletions for a given task list
// this handler limits the amount of tasks deleted to maxTasksPerJob
// for fairness among all the task-list in the system - when there
// is more work to do subsequently, this handler will return StatusDefer
// with the assumption that the executor will schedule this task later
//
// Each loop of the handler proceeds as follows
//    - Attempt to delete tasks up to the persisted ACK level.
//    - If there are 0 tasks for this task-list, try deleting the task-list if its idle
//    - If the number of tasks completed is less than the batchSize, there are no more tasks in the task-list
//      Try deleting the task-list if its idle
func (s *Scavenger) deleteHandler(info *p.TaskListInfo) handlerStatus {
	var err error
	var nProcessed, nDeleted int

	taskBatchSize := s.taskBatchSizeFn()
	max := s.maxTasksPerJobFn()

	defer func() { s.deleteHandlerLog(info, nProcessed, nDeleted, err) }()

	for nProcessed < max {
		// First we complete tasks up to the persisted acklevel regardless of the task expiration
		nTasks, err := s.completeTasks(info, info.AckLevel, taskBatchSize)
		if err == p.ErrPersistenceLimitExceeded {
			s.logger.Info("scavenger.deleteHandler query was ratelimited; will retry")
			return handlerStatusDefer
		}
		if err != nil {
			s.logger.Error(fmt.Sprintf("Scavenger error completing tasks: %v", err))
			return handlerStatusErr
		}
		nProcessed += nTasks
		nDeleted += nTasks
		if nTasks < taskBatchSize {
			break
		}
	}
	for nProcessed < max {
		// if we finished completing tasks below the ack level, but still have budget, then focus on expired tasks
		resp, err1 := s.getTasks(info, taskBatchSize)
		if err1 != nil {
			err = err1
			s.logger.Error(fmt.Sprintf("Scavenger error getting tasks: %v", err))
			return handlerStatusErr
		}

		nTasks := 0
		unexpiredTaskFound := false
		for _, task := range resp.Tasks {
			nProcessed++
			if !s.isTaskExpired(task) {
				s.logger.Info(fmt.Sprintf("Scavenger stopping at an unexpired task. (Expires %v)", task.Expiry), tag.WorkflowTaskListName(info.Name), tag.TaskType(info.TaskType), tag.WorkflowDomainID(task.DomainID), tag.WorkflowID(task.WorkflowID), tag.WorkflowRunID(task.RunID), tag.TaskID(task.TaskID))
				unexpiredTaskFound = true
				break
			}
			nTasks++
		}

		if nTasks > 0 {
			taskID := resp.Tasks[nTasks-1].TaskID
			if _, err = s.completeTasks(info, taskID, nTasks); err != nil {
				return handlerStatusErr
			}
			nDeleted += nTasks
		}

		if unexpiredTaskFound {
			return handlerStatusDone
		}

		if nTasks < taskBatchSize {
			// with no unexpired tasks left, it's safe to try and delete the task list itself
			s.tryDeleteTaskList(info)
			return handlerStatusDone
		}
	}

	return handlerStatusDefer
}

func (s *Scavenger) tryDeleteTaskList(info *p.TaskListInfo) {
	if strings.HasPrefix(info.Name, scannerTaskListPrefix) {
		return // avoid deleting our own task list
	}
	delta := time.Now().Sub(info.LastUpdated)
	if delta < taskListGracePeriod {
		return
	}
	// usually, matching engine is the authoritative owner of a tasklist
	// and its incorrect for any other entity to mutate executorTask lists (including deleting it)
	// the delete here is safe because of two reasons:
	//   - we delete the executorTask list only if the lastUpdated is > 48H. If a executorTask list is idle for
	//     this amount of time, it will no longer be owned by any host in matching engine (because
	//     of idle timeout). If any new host has to take ownership of this at this time, it can only
	//     do so by updating the rangeID
	//   - deleteTaskList is a conditional delete where condition is the rangeID
	if err := s.deleteTaskList(info); err != nil {
		s.logger.Error("deleteTaskList error", tag.Error(err))
		return
	}
	atomic.AddInt64(&s.stats.tasklist.nDeleted, 1)
	s.logger.Info("tasklist deleted", tag.WorkflowDomainID(info.DomainID), tag.WorkflowTaskListName(info.Name), tag.TaskType(info.TaskType))
}

func (s *Scavenger) deleteHandlerLog(info *p.TaskListInfo, nProcessed int, nDeleted int, err error) {
	atomic.AddInt64(&s.stats.task.nDeleted, int64(nDeleted))
	atomic.AddInt64(&s.stats.task.nProcessed, int64(nProcessed))
	if err != nil {
		s.logger.Error("scavenger.deleteHandler processed.",
			tag.Error(err), tag.WorkflowDomainID(info.DomainID), tag.WorkflowTaskListName(info.Name), tag.TaskType(info.TaskType), tag.NumberProcessed(nProcessed), tag.NumberDeleted(nDeleted))
		return
	}
	if nProcessed > 0 {
		s.logger.Info("scavenger.deleteHandler processed.",
			tag.WorkflowDomainID(info.DomainID), tag.WorkflowTaskListName(info.Name), tag.TaskType(info.TaskType), tag.NumberProcessed(nProcessed), tag.NumberDeleted(nDeleted))
	}
}

func (s *Scavenger) isTaskExpired(t *p.TaskInfo) bool {
	return t.Expiry.After(epochStartTime) && time.Now().After(t.Expiry)
}

func (s *Scavenger) completeOrphanTasksHandler() handlerStatus {
	var nDeleted int
	batchSize := s.getOrphanTasksPageSizeFn()
	resp, err := s.getOrphanTasks(batchSize)
	if err == p.ErrPersistenceLimitExceeded {
		s.logger.Info("scavenger.completeOrphanTasksHandler query was ratelimited; will retry")
		return handlerStatusDefer
	}
	if err != nil {
		s.logger.Error("scavenger.completeOrphanTasksHandler error getting orphan tasks")
		return handlerStatusErr
	}
	for _, taskKey := range resp.Tasks {
		err = s.completeTask(&p.TaskListInfo{
			DomainID: taskKey.DomainID,
			Name:     taskKey.TaskListName,
			TaskType: taskKey.TaskType,
		}, taskKey.TaskID)
		if err == p.ErrPersistenceLimitExceeded {
			s.logger.Info("scavenger.completeOrphanTasksHandler query was ratelimited; will retry")
			return handlerStatusDefer
		}
		if err != nil {
			s.logger.Error("scavenger.completeOrphanTasksHandler error getting orphan tasks")
			return handlerStatusErr
		}
		nDeleted++
		atomic.AddInt64(&s.stats.task.nDeleted, 1)
		atomic.AddInt64(&s.stats.task.nProcessed, 1)
	}
	s.logger.Info("scavenger.completeOrphanTasksHandler deleted.", tag.NumberDeleted(nDeleted))
	if len(resp.Tasks) < batchSize {
		return handlerStatusDone
	}
	return handlerStatusDefer
}
