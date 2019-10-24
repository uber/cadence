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

package task

import (
	"fmt"
	"sort"
	"sync"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
)

type (
	// AckMgr is the interface for reading/acknowledging Tasks
	AckMgr interface {
		AddTasks([]Info)
		CompleteTask(Info)
		AckLevel() Info
	}

	ackMgrImpl struct {
		sync.RWMutex
		outstandingTasks map[Info]bool
		ackLevel         Info

		logger       log.Logger
		metricsScope metrics.Scope
	}
)

// NewAckMgr creates a new AckMgr
func NewAckMgr(
	logger log.Logger,
	metricsScope metrics.Scope,
) AckMgr {
	return &ackMgrImpl{
		logger:           logger,
		metricsScope:     metricsScope,
		outstandingTasks: make(map[Info]bool),
	}
}

func (a *ackMgrImpl) AddTasks(tasks []Info) {
	a.Lock()
	defer a.Unlock()
	for _, task := range tasks {
		if _, isLoaded := a.outstandingTasks[task]; isLoaded {
			a.logger.Debug(fmt.Sprintf("Skipping task: TaskID: %v, VisibilityTimestamp: %v, WorkflowID: %v, RunID: %v, Type: %v",
				task.GetTaskID(), task.GetVisibilityTimestamp(), task.GetWorkflowID(), task.GetRunID(), task.GetTaskType()))
			continue
		}
		a.outstandingTasks[task] = false
	}
}

func (a *ackMgrImpl) CompleteTask(task Info) {
	a.Lock()
	defer a.Unlock()
	if _, ok := a.outstandingTasks[task]; ok {
		a.outstandingTasks[task] = true
	}
}

func (a *ackMgrImpl) AckLevel() Info {
	a.metricsScope.IncCounter(metrics.AckLevelUpdateCounter)
	a.Lock()
	defer a.Unlock()

	var tasks []Info
	for task := range a.outstandingTasks {
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return compareTaskInfoLess(tasks[i], tasks[j])
	})
	// TODO: setup new metrics for # of pending tasks and emit such metrics here.
	// may need to pass in addition parameters for the metrics when create the ackMgr.

	for _, current := range tasks {
		if a.outstandingTasks[current] {
			a.ackLevel = current
			delete(a.outstandingTasks, current)
			a.logger.Debug(fmt.Sprintf("Moving task ack level to %v.", a.ackLevel))
		} else {
			break
		}
	}
	return a.ackLevel
}

func compareTaskInfoLess(
	first Info,
	second Info,
) bool {
	if first.GetVisibilityTimestamp().Before(second.GetVisibilityTimestamp()) {
		return true
	}
	if first.GetVisibilityTimestamp().Equal(second.GetVisibilityTimestamp()) {
		return first.GetTaskID() < second.GetTaskID()
	}
	return false
}
