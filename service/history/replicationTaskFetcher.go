// Copyright (c) 2019 Uber Technologies, Inc.
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
	"context"
	"time"

	"github.com/uber/cadence/.gen/go/cadence/workflowserviceclient"
	r "github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
)

const (
	timerFireInterval       = 2 * time.Second
	timerRetryInterval      = 1 * time.Second
	timerJitter             = 0.15
	fetchTaskRequestTimeout = 10 * time.Second
)

type replicationTaskFetcher struct {
	sourceCluster string
	numFetchers   int
	logger        log.Logger
	remotePeer    workflowserviceclient.Interface
	requestChan   chan *request
	done          chan struct{}
}

func newReplicationTaskFetcher(logger log.Logger, sourceCluster string, numFetchers int, sourceFrontend workflowserviceclient.Interface) *replicationTaskFetcher {
	return &replicationTaskFetcher{
		numFetchers:   numFetchers,
		logger:        logger,
		remotePeer:    sourceFrontend,
		sourceCluster: sourceCluster,
		requestChan:   make(chan *request, 1000),
		done:          make(chan struct{}),
	}
}

func (f *replicationTaskFetcher) Start() {
	for i := 0; i < f.numFetchers; i++ {
		go f.fetchTasks()
	}
	f.logger.Info("Replication task fetcher started.", tag.ClusterName(f.sourceCluster), tag.Counter(f.numFetchers))
}

func (f *replicationTaskFetcher) Stop() {
	close(f.done)
	f.logger.Info("Replication task fetcher stopped.", tag.ClusterName(f.sourceCluster))
}

func (f *replicationTaskFetcher) fetchTasks() {
	jitter := backoff.NewJitter()
	timer := time.NewTimer(jitter.JitDuration(timerFireInterval, timerJitter))
	defer timer.Stop()

	requestByShard := make(map[int32]*request)
	for {
		select {
		case request := <-f.requestChan:
			requestByShard[*request.token.ShardID] = request
		case <-timer.C:
			var tokens []*r.ReplicationToken
			for _, request := range requestByShard {
				tokens = append(tokens, request.token)
			}

			ctx, cancel := context.WithTimeout(context.Background(), fetchTaskRequestTimeout)
			request := &r.GetReplicationTasksRequest{Tokens: tokens}
			response, err := f.remotePeer.GetReplicationTasks(ctx, request)
			cancel()
			if err != nil {
				f.logger.Error("Failed to get replication tasks", tag.Error(err))
				timer.Reset(jitter.JitDuration(timerRetryInterval, timerJitter))
				continue
			}

			f.logger.Debug("Successfully fetched replication tasks.", tag.Counter(len(response.TasksByShard)))

			for shardID, tasks := range response.TasksByShard {
				request := requestByShard[shardID]
				request.respChan <- tasks
				close(request.respChan)
				delete(requestByShard, shardID)
			}

			timer.Reset(jitter.JitDuration(timerFireInterval, timerJitter))
		case <-f.done:
			timer.Stop()
			return
		}
	}
}

func (f *replicationTaskFetcher) GetSourceCluster() string {
	return f.sourceCluster
}

func (f *replicationTaskFetcher) GetRequestChan() chan<- *request {
	return f.requestChan
}
