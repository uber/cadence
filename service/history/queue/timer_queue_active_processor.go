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
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/service/history/engine"
	"github.com/uber/cadence/service/history/shard"
	"github.com/uber/cadence/service/history/task"
)

func newTimerQueueActiveProcessor(
	clusterName string,
	shard shard.Context,
	historyEngine engine.Engine,
	taskProcessor task.Processor,
	taskAllocator TaskAllocator,
	taskExecutor task.Executor,
	logger log.Logger,
) *timerQueueProcessorBase {
	config := shard.GetConfig()
	options := newTimerQueueProcessorOptions(config, true, false)

	logger = logger.WithTags(tag.ClusterName(clusterName))

	taskFilter := func(taskInfo task.Info) (bool, error) {
		timer, ok := taskInfo.(*persistence.TimerTaskInfo)
		if !ok {
			return false, errUnexpectedQueueTask
		}
		if notRegistered, err := isDomainNotRegistered(shard, timer.DomainID); notRegistered && err == nil {
			logger.Info("Domain is not in registered status, skip task in active timer queue.", tag.WorkflowDomainID(timer.DomainID), tag.Value(taskInfo))
			return false, nil
		}

		return taskAllocator.VerifyActiveTask(timer.DomainID, timer)
	}

	updateMaxReadLevel := func() task.Key {
		return newTimerTaskKey(shard.UpdateTimerMaxReadLevel(clusterName), 0)
	}

	updateClusterAckLevel := func(ackLevel task.Key) error {
		return shard.UpdateTimerClusterAckLevel(clusterName, ackLevel.(timerTaskKey).visibilityTimestamp)
	}

	updateProcessingQueueStates := func(states []ProcessingQueueState) error {
		pStates := convertToPersistenceTimerProcessingQueueStates(states)
		return shard.UpdateTimerProcessingQueueStates(clusterName, pStates)
	}

	queueShutdown := func() error {
		return nil
	}

	return newTimerQueueProcessorBase(
		clusterName,
		shard,
		loadTimerProcessingQueueStates(clusterName, shard, options, logger),
		taskProcessor,
		NewLocalTimerGate(shard.GetTimeSource()),
		options,
		updateMaxReadLevel,
		updateClusterAckLevel,
		updateProcessingQueueStates,
		queueShutdown,
		taskFilter,
		taskExecutor,
		logger,
		shard.GetMetricsClient(),
	)
}
