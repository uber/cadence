// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package queue

import (
	"sort"
	"time"

	"github.com/uber/cadence/common/collection"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/service/history/task"
)

func newProcessingQueueCollections(
	processingQueueStates []ProcessingQueueState,
	logger log.Logger,
	metricsClient metrics.Client,
) []ProcessingQueueCollection {
	processingQueuesMap := make(map[int][]ProcessingQueue) // level -> state
	for _, queueState := range processingQueueStates {
		processingQueuesMap[queueState.Level()] = append(processingQueuesMap[queueState.Level()], NewProcessingQueue(
			queueState,
			logger,
			metricsClient,
		))
	}
	processingQueueCollections := make([]ProcessingQueueCollection, 0, len(processingQueuesMap))
	for level, queues := range processingQueuesMap {
		processingQueueCollections = append(processingQueueCollections, NewProcessingQueueCollection(
			level,
			queues,
		))
	}
	sort.Slice(processingQueueCollections, func(i, j int) bool {
		return processingQueueCollections[i].Level() < processingQueueCollections[j].Level()
	})

	return processingQueueCollections
}

// RedispatchTasks should be un-exported after the queue processing logic
// in history package is deprecated.
func RedispatchTasks(
	redispatchQueue collection.Queue,
	taskProcessor task.Processor,
	logger log.Logger,
	metricsScope metrics.Scope,
	shutdownCh <-chan struct{},
) {
	queueLength := redispatchQueue.Len()
	metricsScope.RecordTimer(metrics.TaskRedispatchQueuePendingTasksTimer, time.Duration(queueLength))
	for i := 0; i != queueLength; i++ {
		queueTask := redispatchQueue.Remove().(task.Task)
		submitted, err := taskProcessor.TrySubmit(queueTask)
		if err != nil {
			// the only reason error will be returned here is because
			// task processor has already shutdown. Just return in this case.
			logger.Error("failed to redispatch task", tag.Error(err))
			return
		}
		if !submitted {
			// failed to submit, enqueue again
			redispatchQueue.Add(queueTask)
		}

		select {
		case <-shutdownCh:
			return
		default:
		}
	}
}

func splitProcessingQueueCollection(
	processingQueueCollections []ProcessingQueueCollection,
	splitPolicy ProcessingQueueSplitPolicy,
) []ProcessingQueueCollection {
	if splitPolicy == nil {
		return processingQueueCollections
	}

	newQueuesMap := make(map[int][]ProcessingQueue)
	for _, queueCollection := range processingQueueCollections {
		newQueues := queueCollection.Split(splitPolicy)
		for _, newQueue := range newQueues {
			newQueueLevel := newQueue.State().Level()
			newQueuesMap[newQueueLevel] = append(newQueuesMap[newQueueLevel], newQueue)
		}

		if queuesToMerge, ok := newQueuesMap[queueCollection.Level()]; ok {
			queueCollection.Merge(queuesToMerge)
			delete(newQueuesMap, queueCollection.Level())
		}
	}

	for level, newQueues := range newQueuesMap {
		processingQueueCollections = append(processingQueueCollections, NewProcessingQueueCollection(
			level,
			newQueues,
		))
	}

	sort.Slice(processingQueueCollections, func(i, j int) bool {
		return processingQueueCollections[i].Level() < processingQueueCollections[j].Level()
	})

	return processingQueueCollections
}
