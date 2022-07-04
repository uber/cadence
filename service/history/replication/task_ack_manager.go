// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination task_ack_manager_mock.go

package replication

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/shard"
)

var (
	errUnknownReplicationTask = errors.New("unknown replication task")
)

type (
	// TaskAckManager is the ack manager for replication tasks
	TaskAckManager interface {
		GetTask(
			ctx context.Context,
			taskInfo *types.ReplicationTaskInfo,
		) (*types.ReplicationTask, error)

		GetTasks(
			ctx context.Context,
			pollingCluster string,
			lastReadTaskID int64,
		) (*types.ReplicationMessages, error)
	}

	taskAckManagerImpl struct {
		shard         shard.Context
		ackLevels     ackLevelStore
		domains       cache.DomainCache
		rateLimiter   *quotas.DynamicRateLimiter
		retryPolicy   backoff.RetryPolicy
		throttleRetry *backoff.ThrottleRetry

		scope  metrics.Scope
		logger log.Logger

		taskReader   *DynamicTaskReader
		taskHydrator TaskHydrator
	}

	ackLevelStore interface {
		GetTransferMaxReadLevel() int64

		GetClusterReplicationLevel(cluster string) int64
		UpdateClusterReplicationLevel(cluster string, lastTaskID int64) error
	}
)

var _ TaskAckManager = (*taskAckManagerImpl)(nil)

// NewTaskAckManager initializes a new replication task ack manager
func NewTaskAckManager(
	shard shard.Context,
	taskReader *DynamicTaskReader,
	taskHydrator TaskHydrator,
) TaskAckManager {

	config := shard.GetConfig()
	rateLimiter := quotas.NewDynamicRateLimiter(config.ReplicationTaskGenerationQPS.AsFloat64())

	retryPolicy := backoff.NewExponentialRetryPolicy(100 * time.Millisecond)
	retryPolicy.SetMaximumAttempts(config.ReplicatorReadTaskMaxRetryCount())
	retryPolicy.SetBackoffCoefficient(1)

	return &taskAckManagerImpl{
		shard:       shard,
		ackLevels:   shard,
		domains:     shard.GetDomainCache(),
		rateLimiter: rateLimiter,
		retryPolicy: retryPolicy,
		throttleRetry: backoff.NewThrottleRetry(
			backoff.WithRetryPolicy(retryPolicy),
			backoff.WithRetryableError(persistence.IsTransientError),
		),
		scope: shard.GetMetricsClient().Scope(
			metrics.ReplicatorQueueProcessorScope,
			metrics.InstanceTag(strconv.Itoa(shard.GetShardID())),
		),
		logger:       shard.GetLogger().WithTags(tag.ComponentReplicationAckManager),
		taskReader:   taskReader,
		taskHydrator: taskHydrator,
	}
}

func (t *taskAckManagerImpl) GetTask(ctx context.Context, taskInfo *types.ReplicationTaskInfo) (*types.ReplicationTask, error) {
	task := persistence.ReplicationTaskInfo{
		DomainID:     taskInfo.DomainID,
		WorkflowID:   taskInfo.WorkflowID,
		RunID:        taskInfo.RunID,
		TaskID:       taskInfo.TaskID,
		TaskType:     int(taskInfo.TaskType),
		FirstEventID: taskInfo.FirstEventID,
		NextEventID:  taskInfo.NextEventID,
		Version:      taskInfo.Version,
		ScheduledID:  taskInfo.ScheduledID,
	}

	return t.taskHydrator.Hydrate(ctx, task)
}

func (t *taskAckManagerImpl) GetTasks(
	ctx context.Context,
	pollingCluster string,
	lastReadTaskID int64,
) (*types.ReplicationMessages, error) {

	if lastReadTaskID == common.EmptyMessageID {
		lastReadTaskID = t.shard.GetClusterReplicationLevel(pollingCluster)
	}

	taskGeneratedTimer := t.scope.StartTimer(metrics.TaskLatency)

	tasks, hasMore, err := t.taskReader.Read(ctx, lastReadTaskID, t.shard.GetTransferMaxReadLevel())
	if err != nil {
		return nil, err
	}

	var replicationTasks []*types.ReplicationTask
	readLevel := lastReadTaskID
TaskInfoLoop:
	for _, task := range tasks {
		// filter task info by domain clusters.
		domainEntity, err := t.domains.GetDomainByID(task.DomainID)
		if err != nil {
			return nil, err
		}
		if skipTask(pollingCluster, domainEntity) {
			readLevel = task.TaskID
			continue
		}

		// construct replication task from DB
		_ = t.rateLimiter.Wait(ctx)
		var replicationTask *types.ReplicationTask
		op := func() error {
			var err error
			replicationTask, err = t.taskHydrator.Hydrate(ctx, *task)
			return err
		}
		err = t.throttleRetry.Do(ctx, op)
		switch err.(type) {
		case nil:
			// No action
		case *types.BadRequestError, *types.InternalDataInconsistencyError, *types.EntityNotExistsError:
			t.logger.Warn("Failed to get replication task.", tag.Error(err))
		default:
			t.logger.Error("Failed to get replication task. Return what we have so far.", tag.Error(err))
			hasMore = true
			break TaskInfoLoop
		}
		readLevel = task.TaskID
		if replicationTask != nil {
			replicationTasks = append(replicationTasks, replicationTask)
		}
	}
	taskGeneratedTimer.Stop()

	t.scope.RecordTimer(
		metrics.ReplicationTasksLag,
		time.Duration(t.shard.GetTransferMaxReadLevel()-readLevel),
	)
	t.scope.RecordTimer(
		metrics.ReplicationTasksFetched,
		time.Duration(len(tasks)),
	)
	t.scope.RecordTimer(
		metrics.ReplicationTasksReturned,
		time.Duration(len(replicationTasks)),
	)
	t.scope.RecordTimer(
		metrics.ReplicationTasksReturnedDiff,
		time.Duration(len(tasks)-len(replicationTasks)),
	)

	if err := t.shard.UpdateClusterReplicationLevel(
		pollingCluster,
		lastReadTaskID,
	); err != nil {
		t.logger.Error("error updating replication level for shard", tag.Error(err), tag.OperationFailed)
	}

	t.logger.Debug("Get replication tasks", tag.SourceCluster(pollingCluster), tag.ShardReplicationAck(lastReadTaskID), tag.ReadLevel(readLevel))
	return &types.ReplicationMessages{
		ReplicationTasks:       replicationTasks,
		HasMore:                hasMore,
		LastRetrievedMessageID: readLevel,
	}, nil
}

func skipTask(pollingCluster string, domainEntity *cache.DomainCacheEntry) bool {
	for _, cluster := range domainEntity.GetReplicationConfig().Clusters {
		if cluster.ClusterName == pollingCluster {
			return false
		}
	}
	return true
}
