// Copyright (c) 2017-2021 Uber Technologies Inc.

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
	"fmt"
	"sync/atomic"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/engine"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
	"github.com/uber/cadence/service/history/task"
)

type (
	crossClusterQueueProcessor struct {
		shard         shard.Context
		historyEngine engine.Engine
		taskProcessor task.Processor

		metricsClient metrics.Client
		logger        log.Logger

		status       int32
		shutdownChan chan struct{}

		queueProcessors map[string]*crossClusterQueueProcessorBase
	}
)

// NewCrossClusterQueueProcessor creates a new cross cluster QueueProcessor
func NewCrossClusterQueueProcessor(
	shard shard.Context,
	historyEngine engine.Engine,
	executionCache *execution.Cache,
	taskProcessor task.Processor,
) Processor {
	logger := shard.GetLogger().WithTags(tag.ComponentCrossClusterQueueProcessor)
	currentClusterName := shard.GetClusterMetadata().GetCurrentClusterName()

	queueProcessors := make(map[string]*crossClusterQueueProcessorBase)
	for clusterName, info := range shard.GetClusterMetadata().GetAllClusterInfo() {
		if !info.Enabled || clusterName == currentClusterName {
			continue
		}

		taskExecutor := task.NewCrossClusterSourceTaskExecutor(
			shard,
			executionCache,
			logger,
		)

		queueProcessor := newCrossClusterQueueProcessorBase(
			shard,
			clusterName,
			executionCache,
			taskProcessor,
			taskExecutor,
			logger,
		)
		queueProcessors[clusterName] = queueProcessor
	}

	return &crossClusterQueueProcessor{
		shard:           shard,
		historyEngine:   historyEngine,
		taskProcessor:   taskProcessor,
		metricsClient:   shard.GetMetricsClient(),
		logger:          logger,
		status:          common.DaemonStatusInitialized,
		shutdownChan:    make(chan struct{}),
		queueProcessors: queueProcessors,
	}
}

func (c *crossClusterQueueProcessor) Start() {
	if !atomic.CompareAndSwapInt32(&c.status, common.DaemonStatusInitialized, common.DaemonStatusStarted) {
		return
	}

	if c.shard.GetClusterMetadata().IsGlobalDomainEnabled() {
		for _, queueProcessor := range c.queueProcessors {
			queueProcessor.Start()
		}
	}
}

func (c *crossClusterQueueProcessor) Stop() {
	if !atomic.CompareAndSwapInt32(&c.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}

	for _, queueProcessor := range c.queueProcessors {
		queueProcessor.Stop()
	}

	close(c.shutdownChan)
}

func (c *crossClusterQueueProcessor) NotifyNewTask(
	clusterName string,
	executionInfo *persistence.WorkflowExecutionInfo,
	tasks []persistence.Task,
) {
	if len(tasks) == 0 {
		return
	}

	queueProcessor, ok := c.queueProcessors[clusterName]
	if !ok {
		panic(fmt.Sprintf("Cannot find cross cluster processor for %s.", clusterName))
	}
	queueProcessor.notifyNewTask()
}

func (c *crossClusterQueueProcessor) HandleAction(
	ctx context.Context,
	clusterName string,
	action *Action,
) (*ActionResult, error) {

	queueProcessor, ok := c.queueProcessors[clusterName]
	if !ok {
		return nil, &types.BadRequestError{
			Message: fmt.Sprintf("failed to find the cross cluster queue with cluster name: %v", clusterName),
		}
	}

	resultNotificationCh, added := queueProcessor.addAction(ctx, action)
	if !added {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, errProcessorShutdown
	}

	select {
	case resultNotification := <-resultNotificationCh:
		return resultNotification.result, resultNotification.err
	case <-c.shutdownChan:
		return nil, errProcessorShutdown
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *crossClusterQueueProcessor) FailoverDomain(
	domains map[string]struct{},
) {
	for _, queueBase := range c.queueProcessors {
		queueBase.notifyDomainFailover(domains)
	}
}

func (c *crossClusterQueueProcessor) LockTaskProcessing() {
	panic("cross cluster queue doesn't provide locking")
}

func (c *crossClusterQueueProcessor) UnlockTaskProcessing() {
	panic("cross cluster queue doesn't provide locking")
}
