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
	"github.com/uber/cadence/common/persistence"
)

type (
	timerQueueProcessorImpl struct {
		shard                  ShardContext
		activeTimerProcessor   *timerQueueProcessorBase
		standbyTimerProcessors map[string]*timerQueueProcessorBase
	}
)

func newTimerQueueProcessor(shard ShardContext, historyService *historyEngineImpl, executionManager persistence.ExecutionManager, logger bark.Logger) timerQueueProcessor {
	standbyTimerProcessors := make(map[string]*timerQueueProcessorBase)
	for clusterName := range shard.GetService().GetClusterMetadata().GetAllClusterFailoverVersions() {
		if clusterName != shard.GetService().GetClusterMetadata().GetCurrentClusterName() {
			standbyTimerProcessors[clusterName] = newTimerQueueStandbyProcessor(shard, historyService, executionManager, clusterName, logger).timerQueueProcessorBase
		}
	}
	return &timerQueueProcessorImpl{
		shard:                  shard,
		activeTimerProcessor:   newTimerQueueActiveProcessor(shard, historyService, executionManager, logger).timerQueueProcessorBase,
		standbyTimerProcessors: standbyTimerProcessors,
	}
}

func (t *timerQueueProcessorImpl) Start() {
	t.activeTimerProcessor.Start()
	for _, standbyTimerProcessor := range t.standbyTimerProcessors {
		standbyTimerProcessor.Start()
	}
}

func (t *timerQueueProcessorImpl) Stop() {
	t.activeTimerProcessor.Stop()
	for _, standbyTimerProcessor := range t.standbyTimerProcessors {
		standbyTimerProcessor.Stop()
	}
}

// NotifyNewTimers - Notify the processor about the new active / standby timer arrival.
// This should be called each time new timer arrives, otherwise timers maybe fired unexpected.
func (t *timerQueueProcessorImpl) NotifyNewTimers(clusterName string, timerTasks []persistence.Task) {
	if clusterName == t.shard.GetService().GetClusterMetadata().GetCurrentClusterName() {
		t.activeTimerProcessor.NotifyNewTimers(timerTasks)
	} else {
		standbyTimerProcessor, ok := t.standbyTimerProcessors[clusterName]
		if !ok {
			panic("")
		}
		standbyTimerProcessor.NotifyNewTimers(timerTasks)
	}
}

func (t *timerQueueProcessorImpl) getTimerFiredCount(clusterName string) uint64 {
	if clusterName == t.shard.GetService().GetClusterMetadata().GetCurrentClusterName() {
		return t.activeTimerProcessor.getTimerFiredCount()
	}

	standbyTimerProcessor, ok := t.standbyTimerProcessors[clusterName]
	if !ok {
		panic("")
	}
	return standbyTimerProcessor.getTimerFiredCount()
}

// // SyncEventTime - Sync the processor with the view of time of an remote cluster.
// // This should be called each time an event is received from remote cluster.
// func (t *timerQueueMultiProcessorImpl) SyncEventTime(clusterName string, eventTime time.Time) {
// 	standbyTimerProcessor, ok := t.standbyTimerProcessors[clusterName]
// 	if !ok {
// 		panic("")
// 	}
// 	standbyTimerProcessor.SyncEventTime(eventTime)
// }
