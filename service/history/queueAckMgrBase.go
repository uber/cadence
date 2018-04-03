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

package history

import (
	"github.com/uber-common/bark"
	"github.com/uber/cadence/common/metrics"
)

type (
	queueAckMgrBase struct {
		logger        bark.Logger
		metricsClient metrics.Client

		// outstanding timer task -> finished (true)
		outstandingTasks map[int64]bool
		// timer task ack level
		ackLevel int64
	}
)

func newQueueAckMgrBase(ackLevel int64, metricsClient metrics.Client, logger bark.Logger) *queueAckMgrBase {
	queueAckMgrBase := &queueAckMgrBase{
		metricsClient:    metricsClient,
		logger:           logger,
		outstandingTasks: make(map[int64]bool),
		ackLevel:         ackLevel,
	}

	return queueAckMgrBase
}

func (t *queueAckMgrBase) createTask(queueTask queueTaskInfo) {
	_, ok := t.outstandingTasks[queueTask.GetTaskID()]
	if !ok {
		t.outstandingTasks[queueTask.GetTaskID()] = false
	}
}

func (t *queueAckMgrBase) completeTask(queueTask queueTaskInfo) {
	t.outstandingTasks[queueTask.GetTaskID()] = true
}

func (t *queueAckMgrBase) updateAckLevel(maxAckLevel int64) int64 {
	t.metricsClient.IncCounter(metrics.TimerQueueProcessorScope, metrics.AckLevelUpdateCounter)

	initialAckLevel := t.ackLevel
	updatedAckLevel := t.ackLevel

MoveAckLevelLoop:
	for current := initialAckLevel + 1; current <= maxAckLevel; current++ {
		acked := t.outstandingTasks[current]
		if acked {
			updatedAckLevel = current
			delete(t.outstandingTasks, current)
		} else {
			break MoveAckLevelLoop
		}
	}
	t.ackLevel = updatedAckLevel

	return updatedAckLevel
}
