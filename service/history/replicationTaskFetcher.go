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
	"fmt"
	"github.com/uber/cadence/common/backoff"
	"time"

	"github.com/uber/cadence/.gen/go/cadence/workflowserviceclient"
	h "github.com/uber/cadence/.gen/go/history"
	r "github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/service/worker/replicator"
	"go.uber.org/yarpc/yarpcerrors"
)

const (
	dropSyncShardTaskTimeThreshold = 10 * time.Minute
	replicationTimeout             = 30 * time.Second
	timerFireInterval              = 5 * time.Second
	timerRetryInterval             = 1 * time.Second
	timerJitter                    = 0.15
)

var (
	// ErrUnknownReplicationTask is the error to indicate unknown replication task type
	ErrUnknownReplicationTask = &shared.BadRequestError{Message: "unknown replication task"}
)

type (
	replicationTaskProcessor struct {
		shard            ShardContext
		readLevel        int64
		historyEngine    Engine
		sourceCluster    string
		domainReplicator replicator.DomainReplicator
		metricsClient    metrics.Client
		logger           log.Logger

		requestChan chan<- *request
	}

	request struct {
		token    *r.ReplicationToken
		respChan chan<- *r.ReplicationTasksInfo
	}
)

func NewReplicationTaskProcessor(
	shard ShardContext,
	historyEngine Engine,
	domainReplicator replicator.DomainReplicator,
	metricsClient metrics.Client,
	replicationTaskFetcher *replicationTaskFetcher,
) *replicationTaskProcessor {
	return &replicationTaskProcessor{
		shard:            shard,
		historyEngine:    historyEngine,
		sourceCluster:    replicationTaskFetcher.GetSourceCluster(),
		domainReplicator: domainReplicator,
		metricsClient:    metricsClient,
		logger:           shard.GetLogger(),
		requestChan:      replicationTaskFetcher.GetRequestChan(),
	}
}

func (p *replicationTaskProcessor) Start() {
	go func() {
		p.readLevel = p.shard.GetClusterReplicationLevel(p.sourceCluster)

		for {
			respChan := make(chan *r.ReplicationTasksInfo)
			p.requestChan <- &request{
				token:    &r.ReplicationToken{ShardID: common.Int32Ptr(int32(p.shard.GetShardID())), TaskID: common.Int64Ptr(p.readLevel)},
				respChan: respChan,
			}
			response := <-respChan

			p.logger.Info("Got fetch replication tasks response.",
				tag.ReadLevel(response.GetReadLevel()),
				tag.Bool(response.GetHasMore()),
				tag.Counter(len(response.GetReplicationTasks())),
			)

			for _, replicationTask := range response.ReplicationTasks {
				p.processTask(replicationTask)
			}

			p.readLevel = response.GetReadLevel()
			err := p.shard.UpdateClusterReplicationLevel(p.sourceCluster, p.readLevel)
			if err != nil {
				p.logger.Error("Error updating replication level for shard", tag.Error(err), tag.OperationFailed)
			}
		}
	}()

	p.logger.Info("ReplicationTaskProcessor started.")
}

func (p *replicationTaskProcessor) processTask(replicationTask *r.ReplicationTask) {
	var err error
SubmitLoop:
	for {
		var scope int
		switch replicationTask.GetTaskType() {
		case r.ReplicationTaskTypeDomain:
			scope = metrics.DomainReplicationTaskScope
			err = p.handleDomainReplicationTask(replicationTask)
		case r.ReplicationTaskTypeSyncShardStatus:
			scope = metrics.SyncShardTaskScope
			err = p.handleSyncShardTask(replicationTask)
		case r.ReplicationTaskTypeSyncActivity:
			scope = metrics.SyncActivityTaskScope
			err = p.handleActivityTask(replicationTask)
		case r.ReplicationTaskTypeHistory:
			scope = metrics.HistoryReplicationTaskScope
			err = p.handleHistoryReplicationTask(replicationTask)
		case r.ReplicationTaskTypeHistoryMetadata:
			// Without kafka we should not have size limits so we don't necessary need this in the new replication scheme.
		default:
			p.logger.Error("Unknown task type.")
			scope = metrics.ReplicatorScope
			err = ErrUnknownReplicationTask
		}

		if err != nil {
			p.updateFailureMetric(scope, err)
			if !isTransientRetryableError(err) {
				break SubmitLoop
			}
		} else {
			break SubmitLoop
		}
	}

	if err != nil {
		// TODO: insert into our own dlq in cadence persistence?
		// p.nackMsg(msg, err, logger)
		p.logger.Error("Failed to apply replication task")
	}
}

func isTransientRetryableError(err error) bool {
	switch err.(type) {
	case *shared.BadRequestError:
		return false
	default:
		return true
	}
}

func (p *replicationTaskProcessor) updateFailureMetric(scope int, err error) {
	// Always update failure counter for all replicator errors
	p.metricsClient.IncCounter(scope, metrics.ReplicatorFailures)

	// Also update counter to distinguish between type of failures
	switch err := err.(type) {
	case *h.ShardOwnershipLostError:
		p.metricsClient.IncCounter(scope, metrics.CadenceErrShardOwnershipLostCounter)
	case *shared.BadRequestError:
		p.metricsClient.IncCounter(scope, metrics.CadenceErrBadRequestCounter)
	case *shared.DomainNotActiveError:
		p.metricsClient.IncCounter(scope, metrics.CadenceErrDomainNotActiveCounter)
	case *shared.WorkflowExecutionAlreadyStartedError:
		p.metricsClient.IncCounter(scope, metrics.CadenceErrExecutionAlreadyStartedCounter)
	case *shared.EntityNotExistsError:
		p.metricsClient.IncCounter(scope, metrics.CadenceErrEntityNotExistsCounter)
	case *shared.LimitExceededError:
		p.metricsClient.IncCounter(scope, metrics.CadenceErrLimitExceededCounter)
	case *shared.RetryTaskError:
		p.metricsClient.IncCounter(scope, metrics.CadenceErrRetryTaskCounter)
	case *yarpcerrors.Status:
		if err.Code() == yarpcerrors.CodeDeadlineExceeded {
			p.metricsClient.IncCounter(scope, metrics.CadenceErrContextTimeoutCounter)
		}
	}
}

func (p *replicationTaskProcessor) handleActivityTask(task *r.ReplicationTask) error {
	attr := task.SyncActicvityTaskAttributes
	request := &h.SyncActivityRequest{
		DomainId:           attr.DomainId,
		WorkflowId:         attr.WorkflowId,
		RunId:              attr.RunId,
		Version:            attr.Version,
		ScheduledId:        attr.ScheduledId,
		ScheduledTime:      attr.ScheduledTime,
		StartedId:          attr.StartedId,
		StartedTime:        attr.StartedTime,
		LastHeartbeatTime:  attr.LastHeartbeatTime,
		Details:            attr.Details,
		Attempt:            attr.Attempt,
		LastFailureReason:  attr.LastFailureReason,
		LastWorkerIdentity: attr.LastWorkerIdentity,
	}
	ctx, cancel := context.WithTimeout(context.Background(), replicationTimeout)
	defer cancel()
	return p.historyEngine.SyncActivity(ctx, request)
}

func (p *replicationTaskProcessor) handleHistoryReplicationTask(task *r.ReplicationTask) error {
	attr := task.HistoryTaskAttributes
	request := &h.ReplicateEventsRequest{
		SourceCluster: common.StringPtr(p.sourceCluster),
		DomainUUID:    attr.DomainId,
		WorkflowExecution: &shared.WorkflowExecution{
			WorkflowId: attr.WorkflowId,
			RunId:      attr.RunId,
		},
		FirstEventId:            attr.FirstEventId,
		NextEventId:             attr.NextEventId,
		Version:                 attr.Version,
		ReplicationInfo:         attr.ReplicationInfo,
		History:                 attr.History,
		NewRunHistory:           attr.NewRunHistory,
		ForceBufferEvents:       common.BoolPtr(false),
		EventStoreVersion:       attr.EventStoreVersion,
		NewRunEventStoreVersion: attr.NewRunEventStoreVersion,
		ResetWorkflow:           attr.ResetWorkflow,
	}
	ctx, cancel := context.WithTimeout(context.Background(), replicationTimeout)
	defer cancel()
	return p.historyEngine.ReplicateEvents(ctx, request)
}

func (p *replicationTaskProcessor) handleSyncShardTask(task *r.ReplicationTask) error {
	attr := task.SyncShardStatusTaskAttributes
	if time.Now().Sub(time.Unix(0, attr.GetTimestamp())) > dropSyncShardTaskTimeThreshold {
		return nil
	}

	req := &h.SyncShardStatusRequest{
		SourceCluster: attr.SourceCluster,
		ShardId:       attr.ShardId,
		Timestamp:     attr.Timestamp,
	}
	ctx, cancel := context.WithTimeout(context.Background(), replicationTimeout)
	defer cancel()
	return p.historyEngine.SyncShardStatus(ctx, req)
}

func (p *replicationTaskProcessor) handleDomainReplicationTask(task *r.ReplicationTask) error {
	p.metricsClient.IncCounter(metrics.DomainReplicationTaskScope, metrics.ReplicatorMessages)
	sw := p.metricsClient.StartTimer(metrics.DomainReplicationTaskScope, metrics.ReplicatorLatency)
	defer sw.Stop()

	return p.domainReplicator.HandleReceivingTask(task.DomainTaskAttributes)
}

type replicationTaskFetcher struct {
	sourceCluster string
	numFetchers   int
	logger        log.Logger
	remotePeer    workflowserviceclient.Interface
	requestChan   chan *request
	done          chan struct{}
}

func NewReplicationTaskFetcher(logger log.Logger, sourceCluster string, sourceFrontend workflowserviceclient.Interface) *replicationTaskFetcher {
	return &replicationTaskFetcher{
		numFetchers:   1,
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
	fmt.Println("Fetcher started...")
}

func (f *replicationTaskFetcher) fetchTasks() {
	jitter := backoff.NewJitter()
	timer := time.NewTimer(jitter.JitDuration(timerFireInterval, timerJitter))
	defer timer.Stop()

	requestByShard := make(map[int32]*request)
	for {
		select {
		case request := <-f.requestChan:
			fmt.Printf("Fetch task request received token:%v\n", request.token)
			requestByShard[*request.token.ShardID] = request

			// TODO: fetch directly if # tokens > threshold

			// 	if !t.Stop() {
			// 		<-t.C
			// 	}
			// 	t.Reset(d)

		case <-timer.C:
			var tokens []*r.ReplicationToken
			for _, request := range requestByShard {
				tokens = append(tokens, request.token)
			}

			request := &r.GetReplicationTasksRequest{Tokens: tokens}
			response, err := f.remotePeer.GetReplicationTasks(context.Background(), request)
			if err != nil {
				f.logger.Error("Failed to get replication tasks", tag.Error(err))
				timer.Reset(jitter.JitDuration(timerRetryInterval, timerJitter))
				continue
			}

			f.logger.Info("Successfully fetched replication tasks.", tag.Counter(len(response.TasksByShard)))

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
