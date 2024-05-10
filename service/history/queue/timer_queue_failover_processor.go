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
	"time"

	"github.com/pborman/uuid"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/service/history/shard"
	"github.com/uber/cadence/service/history/task"
)

func newTimerQueueFailoverProcessor(
	standbyClusterName string,
	shardContext shard.Context,
	taskProcessor task.Processor,
	taskAllocator TaskAllocator,
	taskExecutor task.Executor,
	logger log.Logger,
	minLevel, maxLevel time.Time,
	domainIDs map[string]struct{},
) (updateClusterAckLevelFn, *timerQueueProcessorBase) {
	config := shardContext.GetConfig()
	options := newTimerQueueProcessorOptions(config, true, true)

	currentClusterName := shardContext.GetService().GetClusterMetadata().GetCurrentClusterName()
	failoverStartTime := shardContext.GetTimeSource().Now()
	failoverUUID := uuid.New()
	logger = logger.WithTags(
		tag.ClusterName(currentClusterName),
		tag.WorkflowDomainIDs(domainIDs),
		tag.FailoverMsg("from: "+standbyClusterName),
	)

	taskFilter := func(taskInfo task.Info) (bool, error) {
		timer, ok := taskInfo.(*persistence.TimerTaskInfo)
		if !ok {
			return false, errUnexpectedQueueTask
		}
		if notRegistered, err := isDomainNotRegistered(shardContext, timer.DomainID); notRegistered && err == nil {
			logger.Info("Domain is not in registered status, skip task in failover timer queue.", tag.WorkflowDomainID(timer.DomainID), tag.Value(taskInfo))
			return false, nil
		}
		return taskAllocator.VerifyFailoverActiveTask(domainIDs, timer.DomainID, timer)
	}

	maxReadLevelTaskKey := newTimerTaskKey(maxLevel, 0)
	updateMaxReadLevel := func() task.Key {
		return maxReadLevelTaskKey // this is a const
	}

	updateClusterAckLevel := func(ackLevel task.Key) error {
		return shardContext.UpdateTimerFailoverLevel(
			failoverUUID,
			shard.TimerFailoverLevel{
				StartTime:    failoverStartTime,
				MinLevel:     minLevel,
				CurrentLevel: ackLevel.(timerTaskKey).visibilityTimestamp,
				MaxLevel:     maxLevel,
				DomainIDs:    domainIDs,
			},
		)
	}

	queueShutdown := func() error {
		return shardContext.DeleteTimerFailoverLevel(failoverUUID)
	}

	processingQueueStates := []ProcessingQueueState{
		NewProcessingQueueState(
			defaultProcessingQueueLevel,
			newTimerTaskKey(minLevel, 0),
			maxReadLevelTaskKey,
			NewDomainFilter(domainIDs, false),
		),
	}

	return updateClusterAckLevel, newTimerQueueProcessorBase(
		currentClusterName, // should use current cluster's time when doing domain failover
		shardContext,
		processingQueueStates,
		taskProcessor,
		NewLocalTimerGate(shardContext.GetTimeSource()),
		options,
		updateMaxReadLevel,
		updateClusterAckLevel,
		nil,
		queueShutdown,
		taskFilter,
		taskExecutor,
		logger,
		shardContext.GetMetricsClient(),
	)
}
