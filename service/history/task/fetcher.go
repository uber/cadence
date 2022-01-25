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
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/uber/cadence/client"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/future"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
)

type (
	// FetcherOptions configures a Fetcher
	FetcherOptions struct {
		Parallelism                dynamicconfig.IntPropertyFn
		AggregationInterval        dynamicconfig.DurationPropertyFn
		ServiceBusyBackoffInterval dynamicconfig.DurationPropertyFn
		ErrorRetryInterval         dynamicconfig.DurationPropertyFn
		TimerJitterCoefficient     dynamicconfig.FloatPropertyFn
	}

	fetchRequest struct {
		shardID  int32
		params   []interface{}
		settable future.Settable
	}

	fetchTaskFunc func(
		ctx context.Context,
		clientBean client.Bean,
		sourceCluster string,
		currentCluster string,
		requestByShard map[int32]fetchRequest,
	) (map[int32]interface{}, error)

	fetcherImpl struct {
		status         int32
		currentCluster string
		sourceCluster  string
		clientBean     client.Bean

		options      *FetcherOptions
		metricsScope metrics.Scope
		logger       log.Logger

		shutdownWG  sync.WaitGroup
		shutdownCh  chan struct{}
		requestChan chan fetchRequest

		fetchCtx       context.Context
		fetchCtxCancel context.CancelFunc
		fetchTaskFunc  fetchTaskFunc
	}
)

const (
	defaultFetchTimeout          = 30 * time.Second
	defaultRequestChanBufferSize = 1000
)

var (
	errTaskFetcherShutdown    = errors.New("task fetcher has already shutdown")
	errDuplicatedFetchRequest = errors.New("duplicated task fetch request")
)

// NewCrossClusterTaskFetchers creates a set of task fetchers,
// one for each source cluster
// The future returned by Fetcher.Get() will have value type []*types.CrossClusterTaskRequest
func NewCrossClusterTaskFetchers(
	clusterMetadata cluster.Metadata,
	clientBean client.Bean,
	options *FetcherOptions,
	metricsClient metrics.Client,
	logger log.Logger,
) Fetchers {
	return newTaskFetchers(
		clusterMetadata,
		clientBean,
		crossClusterTaskFetchFn,
		options,
		metricsClient,
		logger,
	)
}

func crossClusterTaskFetchFn(
	ctx context.Context,
	clientBean client.Bean,
	sourceCluster string,
	currentCluster string,
	requestByShard map[int32]fetchRequest,
) (map[int32]interface{}, error) {
	adminClient := clientBean.GetRemoteAdminClient(sourceCluster)
	shardIDs := make([]int32, 0, len(requestByShard))
	for shardID := range requestByShard {
		shardIDs = append(shardIDs, shardID)
	}
	// number of tasks returned will be controlled by source cluster.
	// if there are lots of tasks in the source cluster, they will be
	// returned in batches.
	request := &types.GetCrossClusterTasksRequest{
		ShardIDs:      shardIDs,
		TargetCluster: currentCluster,
	}
	ctx, cancel := context.WithTimeout(ctx, defaultFetchTimeout)
	defer cancel()
	resp, err := adminClient.GetCrossClusterTasks(ctx, request)
	if err != nil {
		return nil, err
	}

	responseByShard := make(map[int32]interface{}, len(resp.TasksByShard))
	for shardID, tasks := range resp.TasksByShard {
		responseByShard[shardID] = tasks
	}
	for shardID, failedCause := range resp.FailedCauseByShard {
		responseByShard[shardID] = common.ConvertGetTaskFailedCauseToErr(failedCause)
	}
	return responseByShard, nil
}

func newTaskFetchers(
	clusterMetadata cluster.Metadata,
	clientBean client.Bean,
	fetchTaskFunc fetchTaskFunc,
	options *FetcherOptions,
	metricsClient metrics.Client,
	logger log.Logger,
) Fetchers {
	currentClusterName := clusterMetadata.GetCurrentClusterName()
	clusterInfos := clusterMetadata.GetAllClusterInfo()
	fetchers := make([]Fetcher, 0, len(clusterInfos))

	for clusterName, info := range clusterInfos {
		if !info.Enabled || clusterName == currentClusterName {
			continue
		}

		fetchers = append(fetchers, newTaskFetcher(
			currentClusterName,
			clusterName,
			clientBean,
			fetchTaskFunc,
			options,
			metricsClient,
			logger,
		))
	}

	return fetchers
}

// Start is a util method for starting a group of fetchers
func (fetchers Fetchers) Start() {
	for _, fetcher := range fetchers {
		fetcher.Start()
	}
}

// Stop is a util method for stopping a group of fetchers
func (fetchers Fetchers) Stop() {
	for _, fetcher := range fetchers {
		fetcher.Stop()
	}
}

func newTaskFetcher(
	currentCluster string,
	sourceCluster string,
	clientBean client.Bean,
	fetchTaskFunc fetchTaskFunc,
	options *FetcherOptions,
	metricsClient metrics.Client,
	logger log.Logger,
) *fetcherImpl {
	fetchCtx, fetchCtxCancel := context.WithCancel(context.Background())
	return &fetcherImpl{
		status:         common.DaemonStatusInitialized,
		currentCluster: currentCluster,
		sourceCluster:  sourceCluster,
		clientBean:     clientBean,
		options:        options,
		metricsScope: metricsClient.Scope(
			metrics.CrossClusterTaskFetcherScope,
			metrics.ActiveClusterTag(sourceCluster),
		),
		logger: logger.WithTags(
			tag.ComponentCrossClusterTaskFetcher,
			tag.SourceCluster(sourceCluster),
		),
		shutdownCh:     make(chan struct{}),
		requestChan:    make(chan fetchRequest, defaultRequestChanBufferSize),
		fetchCtx:       fetchCtx,
		fetchCtxCancel: fetchCtxCancel,
		fetchTaskFunc:  fetchTaskFunc,
	}
}

func (f *fetcherImpl) Start() {
	if !atomic.CompareAndSwapInt32(&f.status, common.DaemonStatusInitialized, common.DaemonStatusStarted) {
		return
	}

	parallelism := f.options.Parallelism()
	f.shutdownWG.Add(parallelism)
	for i := 0; i != parallelism; i++ {
		go f.aggregator()
	}

	f.logger.Info("Task fetcher started.", tag.LifeCycleStarted, tag.Counter(parallelism))
}

func (f *fetcherImpl) Stop() {
	if !atomic.CompareAndSwapInt32(&f.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}

	close(f.shutdownCh)
	f.fetchCtxCancel()
	if success := common.AwaitWaitGroup(&f.shutdownWG, time.Minute); !success {
		f.logger.Warn("Task fetcher timedout on shutdown.", tag.LifeCycleStopTimedout)
	}

	f.logger.Info("Task fetcher stopped.", tag.LifeCycleStopped)
}

func (f *fetcherImpl) GetSourceCluster() string {
	return f.sourceCluster
}

func (f *fetcherImpl) Fetch(
	shardID int,
	fetchParams ...interface{},
) future.Future {
	future, settable := future.NewFuture()

	select {
	case f.requestChan <- fetchRequest{
		shardID:  int32(shardID),
		params:   fetchParams,
		settable: settable,
	}:
		// no-op
	case <-f.shutdownCh:
		settable.Set(nil, errTaskFetcherShutdown)
		return future
	}

	// we need to check again here
	// since even if shutdownCh is closed, the above select may still
	// push the request to requestChan
	select {
	case <-f.shutdownCh:
		f.drainRequestCh()
	default:
	}

	return future
}

func (f *fetcherImpl) aggregator() {
	defer f.shutdownWG.Done()

	fetchTimer := time.NewTimer(backoff.JitDuration(
		f.options.AggregationInterval(),
		f.options.TimerJitterCoefficient(),
	))

	outstandingRequests := make(map[int32]fetchRequest)

	for {
		select {
		case <-f.shutdownCh:
			fetchTimer.Stop()
			f.drainRequestCh()
			for _, request := range outstandingRequests {
				request.settable.Set(nil, errTaskFetcherShutdown)
			}
			return
		case request := <-f.requestChan:
			if existingRequest, ok := outstandingRequests[request.shardID]; ok {
				existingRequest.settable.Set(nil, errDuplicatedFetchRequest)
			}
			outstandingRequests[request.shardID] = request
		case <-fetchTimer.C:
			var nextFetchInterval time.Duration
			if err := f.fetch(outstandingRequests); err != nil {
				f.logger.Error("Failed to fetch cross cluster tasks", tag.Error(err))
				if common.IsServiceBusyError(err) {
					nextFetchInterval = f.options.ServiceBusyBackoffInterval()
					f.metricsScope.IncCounter(metrics.CrossClusterFetchServiceBusyFailures)
				} else {
					nextFetchInterval = f.options.ErrorRetryInterval()
					f.metricsScope.IncCounter(metrics.CrossClusterFetchFailures)
				}
			} else {
				nextFetchInterval = f.options.AggregationInterval()
			}

			fetchTimer.Reset(backoff.JitDuration(
				nextFetchInterval,
				f.options.TimerJitterCoefficient(),
			))
		}
	}
}

func (f *fetcherImpl) fetch(
	outstandingRequests map[int32]fetchRequest,
) error {
	if len(outstandingRequests) == 0 {
		return nil
	}

	f.metricsScope.IncCounter(metrics.CrossClusterFetchRequests)
	sw := f.metricsScope.StartTimer(metrics.CrossClusterFetchLatency)
	defer sw.Stop()

	responseByShard, err := f.fetchTaskFunc(
		f.fetchCtx,
		f.clientBean,
		f.sourceCluster,
		f.currentCluster,
		outstandingRequests,
	)
	if err != nil {
		return err
	}

	for shardID, response := range responseByShard {
		if request, ok := outstandingRequests[shardID]; ok {
			if fetchErr, ok := response.(error); ok {
				request.settable.Set(nil, fetchErr)
			} else {
				request.settable.Set(response, nil)
			}

			delete(outstandingRequests, shardID)
		}
	}

	return nil
}

func (f *fetcherImpl) drainRequestCh() {
	for {
		select {
		case request := <-f.requestChan:
			request.settable.Set(nil, errTaskFetcherShutdown)
		default:
			return
		}
	}
}
