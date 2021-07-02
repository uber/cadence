// Copyright (c) 2021 Uber Technologies, Inc.
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
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/future"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/shard"
)

const (
	respondCrossClusterTaskTimeout = 5 * time.Second
)

type (
	// CrossClusterTaskProcessorOptions configures crossClusterTaskProcessor
	CrossClusterTaskProcessorOptions struct {
		MaxPendingTasks            dynamicconfig.IntPropertyFn
		TaskMaxRetryCount          dynamicconfig.IntPropertyFn
		TaskRedispatchInterval     dynamicconfig.DurationPropertyFn
		TaskWaitInterval           dynamicconfig.DurationPropertyFn
		ServiceBusyBackoffInterval dynamicconfig.DurationPropertyFn
		TimerJitterCoefficient     dynamicconfig.FloatPropertyFn
	}

	crossClusterTaskProcessors []*crossClusterTaskProcessor

	crossClusterTaskProcessor struct {
		shard         shard.Context
		taskProcessor Processor
		redispatcher  Redispatcher
		taskFetcher   Fetcher
		options       *CrossClusterTaskProcessorOptions
		retryPolicy   backoff.RetryPolicy
		logger        log.Logger
		metricsScope  metrics.Scope

		status     int32
		shutdownCh chan struct{}
		shutdownWG sync.WaitGroup

		taskLock     sync.Mutex
		pendingTasks map[int64]future.Future
	}
)

// NewCrossClusterTaskProcessors creates a list of crossClusterTaskProcessors
// for processing cross cluster tasks at target cluster.
// One processor per source cluster per shard
func NewCrossClusterTaskProcessors(
	shard shard.Context,
	taskProcessor Processor,
	taskFetchers Fetchers,
	options *CrossClusterTaskProcessorOptions,
) common.Daemon {
	processors := make(crossClusterTaskProcessors, 0, len(taskFetchers))
	for _, fetcher := range taskFetchers {
		processor := newCrossClusterTaskProcessor(
			shard,
			taskProcessor,
			fetcher,
			options,
		)
		processors = append(processors, processor)
	}
	return processors
}

func (processors crossClusterTaskProcessors) Start() {
	for _, processor := range processors {
		processor.Start()
	}
}

func (processors crossClusterTaskProcessors) Stop() {
	for _, processor := range processors {
		processor.Stop()
	}
}

func newCrossClusterTaskProcessor(
	shard shard.Context,
	taskProcessor Processor,
	taskFetcher Fetcher,
	options *CrossClusterTaskProcessorOptions,
) *crossClusterTaskProcessor {
	logger := shard.GetLogger().WithTags(
		tag.ComponentCrossClusterTaskProcessor,
		tag.SourceCluster(taskFetcher.GetSourceCluster()),
	)
	metricsScope := shard.GetMetricsClient().Scope(metrics.CrossClusterTaskProcessorScope)
	retryPolicy := backoff.NewExponentialRetryPolicy(time.Millisecond * 100)
	retryPolicy.SetMaximumInterval(time.Second)
	retryPolicy.SetExpirationInterval(options.TaskWaitInterval())
	return &crossClusterTaskProcessor{
		shard:         shard,
		taskProcessor: taskProcessor,
		taskFetcher:   taskFetcher,
		redispatcher: NewRedispatcher(
			taskProcessor,
			&RedispatcherOptions{
				TaskRedispatchInterval:                  options.TaskRedispatchInterval,
				TaskRedispatchIntervalJitterCoefficient: options.TimerJitterCoefficient,
			},
			logger,
			metricsScope,
		),
		options:      options,
		retryPolicy:  retryPolicy,
		logger:       logger,
		metricsScope: metricsScope,

		status:     common.DaemonStatusInitialized,
		shutdownCh: make(chan struct{}),

		pendingTasks: make(map[int64]future.Future),
	}
}

func (p *crossClusterTaskProcessor) Start() {
	if !atomic.CompareAndSwapInt32(&p.status, common.DaemonStatusInitialized, common.DaemonStatusStarted) {
		return
	}

	p.redispatcher.Start()

	p.shutdownWG.Add(2)
	go p.processLoop()
	go p.respondPendingTaskLoop()

	p.logger.Info("Task processor started.", tag.LifeCycleStarted)
}

func (p *crossClusterTaskProcessor) Stop() {
	if !atomic.CompareAndSwapInt32(&p.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}

	close(p.shutdownCh)
	p.redispatcher.Stop()

	if success := common.AwaitWaitGroup(&p.shutdownWG, time.Minute); !success {
		p.logger.Warn("Task processor timedout on shutdown.", tag.LifeCycleStopTimedout)
	}
	p.logger.Info("Task processor stopped.", tag.LifeCycleStopped)
}

func (p *crossClusterTaskProcessor) processLoop() {
	defer p.shutdownWG.Done()

	for {
		if p.hasShutdown() {
			return
		}

		if p.numPendingTasks() > p.options.MaxPendingTasks() {
			time.Sleep(backoff.JitDuration(
				p.options.TaskWaitInterval(),
				p.options.TimerJitterCoefficient(),
			))
			continue
		}

		// this will submit the fetching request to the host level task fetcher for batching
		fetchFuture := p.taskFetcher.Fetch(p.shard.GetShardID())

		var taskRequests []*types.CrossClusterTaskRequest
		if err := fetchFuture.Get(context.Background(), &taskRequests); err != nil {
			p.logger.Error("Unable to fetch cross cluster tasks", tag.Error(err))
			if common.IsServiceBusyError(err) {
				time.Sleep(backoff.JitDuration(
					p.options.ServiceBusyBackoffInterval(),
					p.options.TimerJitterCoefficient(),
				))
			}
			continue
		}

		p.processTaskRequests(taskRequests)
	}
}

func (p *crossClusterTaskProcessor) processTaskRequests(
	taskRequests []*types.CrossClusterTaskRequest,
) {
	taskRequests = p.dedupTaskRequests(taskRequests)
	// it's ok to drop task requests,
	// the same request will be sent by the source cluster again upon next fetch
	for len(taskRequests) != 0 && !p.hasShutdown() && p.numPendingTasks() < p.options.MaxPendingTasks() {

		taskFutures := make([]future.Future, len(taskRequests))
		for idx, taskRequest := range taskRequests {
			crossClusterTask, future := NewCrossClusterTaskForTargetCluster(
				p.shard,
				taskRequest,
				p.logger,
				p.options.TaskMaxRetryCount,
			)
			taskFutures[idx] = future

			if err := p.submitTask(crossClusterTask); err != nil {
				return
			}
		}

		respondRequest := &types.RespondCrossClusterTasksCompletedRequest{
			ShardID:       int32(p.shard.GetShardID()),
			TargetCluster: p.shard.GetClusterMetadata().GetCurrentClusterName(),
			// TODO: uncomment the following line when the field is added
			// FetchNewTasks: common.BoolPtr(p.numPendingTasks() < p.options.MaxPendingTasks()),
		}
		taskWaitContext, cancel := context.WithTimeout(context.Background(), p.options.TaskWaitInterval())
		deadlineExceeded := false
		for idx, taskFuture := range taskFutures {
			if deadlineExceeded && !taskFuture.IsReady() {
				continue
			}

			var taskResponse types.CrossClusterTaskResponse
			if err := taskFuture.Get(taskWaitContext, &taskResponse); err != nil {
				if err == context.DeadlineExceeded {
					// switch to a valid context here, otherwise Get() will also return an error
					// this is fine since we will only be calling Get() with background context
					// when the future is ready
					taskWaitContext = context.Background()
					deadlineExceeded = true
					continue
				}

				// this case should not happen,
				// task failure should be converted to FailCause in the response by the processing logic
				taskResponse = types.CrossClusterTaskResponse{
					TaskID: taskRequests[idx].TaskInfo.GetTaskID(),
					// TODO: uncomment this following line when the Unknown cause is added
					// FailedCause: types.CrossClusterTaskFailedCauseUnknown.Ptr(),
				}
			} else {
				respondRequest.TaskResponses = append(respondRequest.TaskResponses, &taskResponse)
			}
		}
		cancel()

		successfullyRespondedTaskIDs := make(map[int64]struct{})
		var respondResponse *types.RespondCrossClusterTasksCompletedResponse
		var respondErr error
		respondResponse, respondErr = p.respondTaskCompletedWithRetry(respondRequest)
		if respondErr == nil {
			for _, response := range respondRequest.TaskResponses {
				successfullyRespondedTaskIDs[response.GetTaskID()] = struct{}{}
			}
		}

		// move tasks that are still running or failed to respond to pendingTasks map
		// so that the respond can be done later
		p.taskLock.Lock()
		for idx, request := range taskRequests {
			taskID := request.TaskInfo.GetTaskID()
			if _, ok := successfullyRespondedTaskIDs[taskID]; ok {
				continue
			}
			p.pendingTasks[taskID] = taskFutures[idx]
		}
		p.taskLock.Unlock()

		if respondErr != nil {
			return
		}
		taskRequests = p.dedupTaskRequests(respondResponse.Tasks)
	}
}

func (p *crossClusterTaskProcessor) respondPendingTaskLoop() {
	defer p.shutdownWG.Done()

	respondTimer := time.NewTimer(backoff.JitDuration(
		p.options.TaskWaitInterval(),
		p.options.TimerJitterCoefficient(),
	))

	for {
		select {
		case <-p.shutdownCh:
			return
		case <-respondTimer.C:
			// reset the timer first so that if respond task retried for some time
			// we won't an additional TaskWaitInterval before checking the status of
			// pending tasks
			respondTimer.Reset(backoff.JitDuration(
				p.options.TaskWaitInterval(),
				p.options.TimerJitterCoefficient(),
			))
			p.taskLock.Lock()
			respondRequest := &types.RespondCrossClusterTasksCompletedRequest{
				ShardID:       int32(p.shard.GetShardID()),
				TargetCluster: p.shard.GetClusterMetadata().GetCurrentClusterName(),
				// TODO: uncomment this following line when the field is added
				// FetchNewTasks: common.BoolPtr(false),
			}
			for taskID, taskFuture := range p.pendingTasks {
				if taskFuture.IsReady() {
					var taskResponse types.CrossClusterTaskResponse
					if err := taskFuture.Get(context.Background(), &taskResponse); err != nil {
						// this case should not happen,
						// task failure should be converted to FailCause in the response by the processing logic
						taskResponse = types.CrossClusterTaskResponse{
							TaskID: taskID,
							// TODO: uncomment this following line when the Unknown cause is added
							// FailedCause: types.CrossClusterTaskFailedCauseUnknown.Ptr(),
						}
						p.logger.Error("Encountered unknown error from cross cluster task future", tag.Error(err))
					}
					respondRequest.TaskResponses = append(respondRequest.TaskResponses, &taskResponse)
				}
			}
			p.taskLock.Unlock()
			if len(respondRequest.TaskResponses) == 0 {
				continue
			}

			_, err := p.respondTaskCompletedWithRetry(respondRequest)
			if err == nil {
				// we can be sure that source cluster has received the response
				p.taskLock.Lock()
				for _, response := range respondRequest.TaskResponses {
					taskID := response.GetTaskID()
					delete(p.pendingTasks, taskID)
				}
				p.taskLock.Unlock()
			}

			if common.IsServiceBusyError(err) {
				respondTimer.Reset(backoff.JitDuration(
					p.options.ServiceBusyBackoffInterval(),
					p.options.TimerJitterCoefficient(),
				))
			}
		}
	}
}

func (p *crossClusterTaskProcessor) dedupTaskRequests(
	taskRequests []*types.CrossClusterTaskRequest,
) []*types.CrossClusterTaskRequest {
	// NOTE: this is only best effort dedup for reducing the number unnecessary task executions.
	// it's possible that a task is removed from the pendingTasks maps before this dedup logic
	// is executed for that task. In that case, that task will be executed multiple times. This
	// is fine as all task processing logic is supposed to be idempotent.
	dedupedRequests := make([]*types.CrossClusterTaskRequest, 0, len(taskRequests))

	p.taskLock.Lock()
	defer p.taskLock.Unlock()

	for _, taskRequest := range taskRequests {
		taskID := taskRequest.TaskInfo.GetTaskID()
		if _, ok := p.pendingTasks[taskID]; ok {
			continue
		}
		dedupedRequests = append(dedupedRequests, taskRequest)
	}

	return dedupedRequests
}

func (p *crossClusterTaskProcessor) respondTaskCompletedWithRetry(
	request *types.RespondCrossClusterTasksCompletedRequest,
) (*types.RespondCrossClusterTasksCompletedResponse, error) {
	var response *types.RespondCrossClusterTasksCompletedResponse
	err := backoff.Retry(
		func() error {
			ctx, cancel := context.WithTimeout(context.Background(), respondCrossClusterTaskTimeout)
			defer cancel()

			var err error
			response, err = p.shard.GetService().GetHistoryRawClient().RespondCrossClusterTasksCompleted(ctx, request)
			return err
		},
		p.retryPolicy,
		func(err error) bool {
			if common.IsServiceBusyError(err) {
				return false
			}
			return common.IsServiceTransientError(err)
		},
	)

	return response, err
}

// submitTask submits the task to the host level task processor
// or to the redispatch queue if failed to submit (task ch full, or other errors)
// so that the submission can be retried later
// error will be returned by this function only when the shard has been shutdown
func (p *crossClusterTaskProcessor) submitTask(
	task CrossClusterTask,
) error {
	submitted, err := p.taskProcessor.TrySubmit(task)
	if err != nil {
		if p.hasShutdown() {
			return err
		}
	}

	// p.logger.Error("Failed to submit task", tag.Error(err))
	if err != nil || !submitted {
		p.redispatcher.AddTask(task)
	}
	return nil
}

func (p *crossClusterTaskProcessor) numPendingTasks() int {
	p.taskLock.Lock()
	defer p.taskLock.Unlock()

	return len(p.pendingTasks)
}

func (p *crossClusterTaskProcessor) hasShutdown() bool {
	select {
	case <-p.shutdownCh:
		return true
	default:
		return false
	}
}
