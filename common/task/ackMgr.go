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

	"github.com/uber/cadence/common/collection"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
)

type (
	// AckMgr is the interface for reading/acknowledging Tasks
	AckMgr interface {
		AddTasks([]Info)
		CompleteTask(Info)
		UpdateAckLevel() SequenceID
		PendingTasks() int64
	}

	ackMgrImpl struct {
		logger       log.Logger
		metricsScope metrics.Scope
		iterator     collection.Iterator

		sync.RWMutex
		outstandingTasks map[SequenceID]bool
		ackLevel         SequenceID
	}
)

// NewAckMgr creates a new AckMgr
func NewAckMgr(
	logger log.Logger,
	metricsScope metrics.Scope,
	ackLevel SequenceID,
) AckMgr {
	return &ackMgrImpl{
		logger:           logger,
		metricsScope:     metricsScope,
		outstandingTasks: make(map[SequenceID]bool),
		ackLevel:         ackLevel,
	}
}

func (a *ackMgrImpl) AddTasks(tasks []Info) {
	a.Lock()
	defer a.Unlock()
	for _, task := range tasks {
		SequenceID := ToSequenceID(task)
		if _, isLoaded := a.outstandingTasks[SequenceID]; isLoaded {
			a.logger.Debug(fmt.Sprintf("Skipping task: %v. WorkflowID: %v, RunID: %v, Type: %v",
				SequenceID, task.GetWorkflowID(), task.GetRunID(), task.GetTaskType()))
			continue
		}
		a.outstandingTasks[SequenceID] = false
	}
}

func (a *ackMgrImpl) CompleteTask(taskInfo Info) {
	a.Lock()
	defer a.Unlock()
	ID := ToSequenceID(taskInfo)
	if _, ok := a.outstandingTasks[ID]; ok {
		a.outstandingTasks[ID] = true
	}
}

func (a *ackMgrImpl) UpdateAckLevel() SequenceID {
	a.metricsScope.IncCounter(metrics.AckLevelUpdateCounter)
	a.Lock()
	defer a.Unlock()

	var SequenceIDs SequenceIDs
	for ID := range a.outstandingTasks {
		SequenceIDs = append(SequenceIDs, ID)
	}
	sort.Sort(SequenceIDs)

	for _, currentLevel := range SequenceIDs {
		if a.outstandingTasks[currentLevel] {
			a.ackLevel = currentLevel
			delete(a.outstandingTasks, currentLevel)
			a.logger.Debug(fmt.Sprintf("Moving task ack level to %v.", a.ackLevel))
		} else {
			break
		}
	}
	return a.ackLevel
}

func (a *ackMgrImpl) PendingTasks() int64 {
	a.RLock()
	defer a.RUnlock()
	return int64(len(a.outstandingTasks))
}
