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
	"sync/atomic"
	"time"

	"strings"

	"github.com/uber-common/bark"
	"github.com/uber/cadence/common/logging"
	p "github.com/uber/cadence/common/persistence"
)

// handlerStatus is the return code from a task list handler
type handlerStatus int

const (
	handlerStatusDone handlerStatus = iota
	handlerStatusDefer
	handlerStatusErr
)

const scannerTaskListPrefix = "cadence-sys-scanner"

// deleteHandler handles deletions for a given task list
func (s *Scavenger) deleteHandler(key *taskListKey, state taskListState) (handlerStatus, error) {
	var err error
	var nProcessed, nDeleted int

	defer func() { s.deleteHandlerLog(key, state, nProcessed, nDeleted, err) }()

	for nProcessed < maxTasksPerJob {
		resp, err1 := s.getTasks(key, taskBatchSize)
		if err1 != nil {
			err = err1
			return handlerStatusErr, err
		}

		nTasks := len(resp.Tasks)
		if nTasks == 0 {
			s.tryDeleteTaskList(key, state)
			return handlerStatusDone, nil
		}

		for _, task := range resp.Tasks {
			nProcessed++
			if !s.isTaskExpired(task) {
				return handlerStatusDone, nil
			}
		}

		taskID := resp.Tasks[nTasks-1].TaskID
		_, err = s.completeTasks(key, taskID, nTasks)
		if err != nil {
			return handlerStatusErr, err
		}

		nDeleted += nTasks
		if nTasks < taskBatchSize {
			s.tryDeleteTaskList(key, state)
			return handlerStatusDone, nil
		}
	}

	return handlerStatusDefer, nil
}

func (s *Scavenger) tryDeleteTaskList(key *taskListKey, state taskListState) {
	if strings.HasPrefix(key.Name, scannerTaskListPrefix) {
		return // avoid deleting our own task list
	}
	delta := time.Now().Sub(state.lastUpdated)
	if delta < taskListGracePeriod {
		return
	}
	err := s.deleteTaskList(key, state.rangeID)
	if err != nil {
		s.logger.Errorf("deleteTaskList error: %v", err)
		return
	}
	atomic.AddInt64(&s.stats.tasklist.nDeleted, 1)
	s.logger.WithFields(bark.Fields{
		logging.TagDomainID:     key.DomainID,
		logging.TagTaskListName: key.Name,
		logging.TagTaskType:     key.TaskType,
	}).Info("tasklist deleted")
}

func (s *Scavenger) deleteHandlerLog(key *taskListKey, state taskListState, nProcessed int, nDeleted int, err error) {
	atomic.AddInt64(&s.stats.task.nDeleted, int64(nDeleted))
	atomic.AddInt64(&s.stats.task.nProcessed, int64(nProcessed))
	if err != nil {
		s.logger.WithFields(bark.Fields{
			logging.TagDomainID:     key.DomainID,
			logging.TagTaskListName: key.Name,
			logging.TagTaskType:     key.TaskType,
			logging.TagErr:          err.Error(),
		}).Errorf("scavenger.deleteHandler: processed:%v, deleted:%v", nProcessed, nDeleted)
		return
	}
	if nProcessed > 0 {
		s.logger.WithFields(bark.Fields{
			logging.TagDomainID:     key.DomainID,
			logging.TagTaskListName: key.Name,
			logging.TagTaskType:     key.TaskType,
		}).Infof("scavenger.deleteHandler: processed:%v, deleted:%v", nProcessed, nDeleted)
	}
}

func (s *Scavenger) isTaskExpired(t *p.TaskInfo) bool {
	return t.Expiry.After(epochStartTime) && time.Now().After(t.Expiry)
}
