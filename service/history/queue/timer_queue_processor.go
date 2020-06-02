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
	"context"
	"sync"
	"time"

	h "github.com/uber/cadence/.gen/go/history"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/xdc"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/engine"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
	"github.com/uber/cadence/service/history/task"
	"github.com/uber/cadence/service/worker/archiver"
)

type (
	timerQueueProcessor struct {
		shard         shard.Context
		historyEngine engine.Engine
		taskProcessor task.Processor

		config                *config.Config
		isGlobalDomainEnabled bool
		currentClusterName    string

		metricsClient metrics.Client
		logger        log.Logger

		status       int32
		shutdownChan chan struct{}
		shutdownWG   sync.WaitGroup

		ackLevel               time.Time
		taskAllocator          TaskAllocator
		activeTaskExecutor     task.Executor
		activeQueueProcessor   *timerQueueProcessorBase
		standbyQueueProcessors map[string]*timerQueueProcessorBase
	}
)

func NewTimerQueueProcessor(
	shard shard.Context,
	historyEngine engine.Engine,
	taskProcessor task.Processor,
	executionCache *execution.Cache,
	archivalClient archiver.Client,
) QueueProcessor {
	logger := shard.GetLogger().WithTags(tag.ComponentTimerQueue)
	currentClusterName := shard.GetClusterMetadata().GetCurrentClusterName()
	taskAllocator := NewTaskAllocator(shard)

	activeTaskExecutor := task.NewTimerActiveTaskExecutor(
		shard,
		archivalClient,
		executionCache,
		logger,
		shard.GetMetricsClient(),
		shard.GetConfig(),
	)

	activeQueueProcessor := newTimerQueueActiveProcessor(
		shard,
		historyEngine,
		taskProcessor,
		taskAllocator,
		activeTaskExecutor,
		logger,
	)

	standbyQueueProcessors := make(map[string]*timerQueueProcessorBase)
	rereplicatorLogger := shard.GetLogger().WithTags(tag.ComponentHistoryReplicator)
	resenderLogger := shard.GetLogger().WithTags(tag.ComponentHistoryResender)
	for clusterName, info := range shard.GetClusterMetadata().GetAllClusterInfo() {
		if !info.Enabled || clusterName == currentClusterName {
			continue
		}

		historyRereplicator := xdc.NewHistoryRereplicator(
			currentClusterName,
			shard.GetDomainCache(),
			shard.GetService().GetClientBean().GetRemoteAdminClient(clusterName),
			func(ctx context.Context, request *h.ReplicateRawEventsRequest) error {
				return historyEngine.ReplicateRawEvents(ctx, request)
			},
			shard.GetService().GetPayloadSerializer(),
			historyRereplicationTimeout,
			nil,
			rereplicatorLogger,
		)
		nDCHistoryResender := xdc.NewNDCHistoryResender(
			shard.GetDomainCache(),
			shard.GetService().GetClientBean().GetRemoteAdminClient(clusterName),
			func(ctx context.Context, request *h.ReplicateEventsV2Request) error {
				return historyEngine.ReplicateEventsV2(ctx, request)
			},
			shard.GetService().GetPayloadSerializer(),
			nil,
			resenderLogger,
		)
		standbyTaskExecutor := task.NewTimerStandbyTaskExecutor(
			shard,
			archivalClient,
			executionCache,
			historyRereplicator,
			nDCHistoryResender,
			logger,
			shard.GetMetricsClient(),
			clusterName,
			shard.GetConfig(),
		)
		standbyQueueProcessors[clusterName] = newTimerQueueStandbyProcessor(
			clusterName,
			shard,
			historyEngine,
			taskProcessor,
			taskAllocator,
			standbyTaskExecutor,
			historyRereplicator,
			nDCHistoryResender,
			logger,
		)
	}

	return &timerQueueProcessor{
		shard:         shard,
		historyEngine: historyEngine,
		taskProcessor: taskProcessor,

		config:                shard.GetConfig(),
		isGlobalDomainEnabled: shard.GetClusterMetadata().IsGlobalDomainEnabled(),
		currentClusterName:    currentClusterName,

		metricsClient: shard.GetMetricsClient(),
		logger:        logger,

		status:       common.DaemonStatusInitialized,
		shutdownChan: make(chan struct{}),

		ackLevel:               shard.GetTimerAckLevel(),
		taskAllocator:          taskAllocator,
		activeTaskExecutor:     activeTaskExecutor,
		activeQueueProcessor:   activeQueueProcessor,
		standbyQueueProcessors: standbyQueueProcessors,
	}
}

func (t *timerQueueProcessor) Start() {

}

func (t *timerQueueProcessor) Stop() {}

func (t *timerQueueProcessor) NotifyNewTask(
	clusterName string,
	transferTasks []persistence.Task,
) {
}

func (t *timerQueueProcessor) FailoverDomain(
	domainIDs map[string]struct{},
) {
}

func (t *timerQueueProcessor) LockTaskProcessing() {
	t.taskAllocator.Lock()
}

func (t *timerQueueProcessor) UnlockTaskProcessing() {
	t.taskAllocator.Unlock()
}

func newTimerQueueActiveProcessor(
	shard shard.Context,
	historyEngine engine.Engine,
	taskProcessor task.Processor,
	taskAllocator TaskAllocator,
	taskExecutor task.Executor,
	logger log.Logger,
) *timerQueueProcessorBase {
	return nil
}

func newTimerQueueStandbyProcessor(
	clusterName string,
	shard shard.Context,
	historyEngine engine.Engine,
	taskProcessor task.Processor,
	taskAllocator TaskAllocator,
	taskExecutor task.Executor,
	historyRereplicator xdc.HistoryRereplicator,
	nDCHistoryResender xdc.NDCHistoryResender,
	logger log.Logger,
) *timerQueueProcessorBase {
	return nil
}
