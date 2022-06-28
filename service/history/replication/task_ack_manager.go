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
	"sync/atomic"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/types"
	exec "github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
	"github.com/uber/cadence/service/history/task"
)

var (
	errUnknownQueueTask       = errors.New("unknown task type")
	errUnknownReplicationTask = errors.New("unknown replication task")
	minReadTaskSize           = 20
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
		shard            shard.Context
		executionManager persistence.ExecutionManager
		rateLimiter      *quotas.DynamicRateLimiter
		retryPolicy      backoff.RetryPolicy
		throttleRetry    *backoff.ThrottleRetry

		lastTaskCreationTime atomic.Value
		maxAllowedLatencyFn  dynamicconfig.DurationPropertyFn

		metricsClient metrics.Client
		logger        log.Logger

		// This is the batch size used by pull based RPC replicator.
		fetchTasksBatchSize dynamicconfig.IntPropertyFnWithShardIDFilter

		historyLoader      HistoryLoader
		mutableStateLoader mutableStateLoader
	}
)

var _ TaskAckManager = (*taskAckManagerImpl)(nil)

// NewTaskAckManager initializes a new replication task ack manager
func NewTaskAckManager(
	shard shard.Context,
	executionCache *exec.Cache,
) TaskAckManager {

	config := shard.GetConfig()
	rateLimiter := quotas.NewDynamicRateLimiter(config.ReplicationTaskGenerationQPS.AsFloat64())

	retryPolicy := backoff.NewExponentialRetryPolicy(100 * time.Millisecond)
	retryPolicy.SetMaximumAttempts(config.ReplicatorReadTaskMaxRetryCount())
	retryPolicy.SetBackoffCoefficient(1)

	return &taskAckManagerImpl{
		shard:            shard,
		executionManager: shard.GetExecutionManager(),
		rateLimiter:      rateLimiter,
		retryPolicy:      retryPolicy,
		throttleRetry: backoff.NewThrottleRetry(
			backoff.WithRetryPolicy(retryPolicy),
			backoff.WithRetryableError(persistence.IsTransientError),
		),
		lastTaskCreationTime: atomic.Value{},
		maxAllowedLatencyFn:  config.ReplicatorUpperLatency,
		metricsClient:        shard.GetMetricsClient(),
		logger:               shard.GetLogger().WithTags(tag.ComponentReplicationAckManager),
		fetchTasksBatchSize:  config.ReplicatorProcessorFetchTasksBatchSize,
		historyLoader: NewHistoryLoader(
			shard.GetShardID(),
			shard.GetHistoryManager(),
		),
		mutableStateLoader: NewMutableStateLoader(executionCache),
	}
}

func (t *taskAckManagerImpl) GetTask(
	ctx context.Context,
	taskInfo *types.ReplicationTaskInfo,
) (*types.ReplicationTask, error) {
	task := &persistence.ReplicationTaskInfo{
		DomainID:     taskInfo.GetDomainID(),
		WorkflowID:   taskInfo.GetWorkflowID(),
		RunID:        taskInfo.GetRunID(),
		TaskID:       taskInfo.GetTaskID(),
		TaskType:     int(taskInfo.GetTaskType()),
		FirstEventID: taskInfo.GetFirstEventID(),
		NextEventID:  taskInfo.GetNextEventID(),
		Version:      taskInfo.GetVersion(),
		ScheduledID:  taskInfo.GetScheduledID(),
	}
	return t.toReplicationTask(ctx, task)
}

func (t *taskAckManagerImpl) GetTasks(
	ctx context.Context,
	pollingCluster string,
	lastReadTaskID int64,
) (*types.ReplicationMessages, error) {

	if lastReadTaskID == common.EmptyMessageID {
		lastReadTaskID = t.shard.GetClusterReplicationLevel(pollingCluster)
	}

	shardID := t.shard.GetShardID()
	replicationScope := t.metricsClient.Scope(
		metrics.ReplicatorQueueProcessorScope,
		metrics.InstanceTag(strconv.Itoa(shardID)),
	)
	taskGeneratedTimer := replicationScope.StartTimer(metrics.TaskLatency)
	batchSize := t.getBatchSize()
	taskInfoList, hasMore, err := t.readTasksWithBatchSize(ctx, lastReadTaskID, batchSize)
	if err != nil {
		return nil, err
	}

	var lastTaskCreationTime time.Time
	var replicationTasks []*types.ReplicationTask
	readLevel := lastReadTaskID
TaskInfoLoop:
	for _, taskInfo := range taskInfoList {
		// filter task info by domain clusters.
		domainEntity, err := t.shard.GetDomainCache().GetDomainByID(taskInfo.GetDomainID())
		if err != nil {
			return nil, err
		}
		if skipTask(pollingCluster, domainEntity) {
			readLevel = taskInfo.GetTaskID()
			continue
		}

		// construct replication task from DB
		_ = t.rateLimiter.Wait(ctx)
		var replicationTask *types.ReplicationTask
		op := func() error {
			var err error
			replicationTask, err = t.toReplicationTask(ctx, taskInfo)
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
		readLevel = taskInfo.GetTaskID()
		if replicationTask != nil {
			replicationTasks = append(replicationTasks, replicationTask)
		}
		if taskInfo.GetVisibilityTimestamp().After(lastTaskCreationTime) {
			lastTaskCreationTime = taskInfo.GetVisibilityTimestamp()
		}
	}
	taskGeneratedTimer.Stop()

	replicationScope.RecordTimer(
		metrics.ReplicationTasksLag,
		time.Duration(t.shard.GetTransferMaxReadLevel()-readLevel),
	)
	replicationScope.RecordTimer(
		metrics.ReplicationTasksFetched,
		time.Duration(len(taskInfoList)),
	)
	replicationScope.RecordTimer(
		metrics.ReplicationTasksReturned,
		time.Duration(len(replicationTasks)),
	)
	replicationScope.RecordTimer(
		metrics.ReplicationTasksReturnedDiff,
		time.Duration(len(taskInfoList)-len(replicationTasks)),
	)

	if err := t.shard.UpdateClusterReplicationLevel(
		pollingCluster,
		lastReadTaskID,
	); err != nil {
		t.logger.Error("error updating replication level for shard", tag.Error(err), tag.OperationFailed)
	}
	t.lastTaskCreationTime.Store(lastTaskCreationTime)

	t.logger.Debug("Get replication tasks", tag.SourceCluster(pollingCluster), tag.ShardReplicationAck(lastReadTaskID), tag.ReadLevel(readLevel))
	return &types.ReplicationMessages{
		ReplicationTasks:       replicationTasks,
		HasMore:                hasMore,
		LastRetrievedMessageID: readLevel,
	}, nil
}

func (t *taskAckManagerImpl) toReplicationTask(
	ctx context.Context,
	taskInfo task.Info,
) (*types.ReplicationTask, error) {

	task, ok := taskInfo.(*persistence.ReplicationTaskInfo)
	if !ok {
		return nil, errUnknownQueueTask
	}

	return Hydrate(ctx, *task, t.mutableStateLoader, t.historyLoader)
}

func (t *taskAckManagerImpl) readTasksWithBatchSize(
	ctx context.Context,
	readLevel int64,
	batchSize int,
) ([]task.Info, bool, error) {

	response, err := t.executionManager.GetReplicationTasks(
		ctx,
		&persistence.GetReplicationTasksRequest{
			ReadLevel:    readLevel,
			MaxReadLevel: t.shard.GetTransferMaxReadLevel(),
			BatchSize:    batchSize,
		},
	)

	if err != nil {
		return nil, false, err
	}

	tasks := make([]task.Info, len(response.Tasks))
	for i := range response.Tasks {
		tasks[i] = response.Tasks[i]
	}

	return tasks, len(response.NextPageToken) != 0, nil
}

func (t *taskAckManagerImpl) getBatchSize() int {

	shardID := t.shard.GetShardID()
	defaultBatchSize := t.fetchTasksBatchSize(shardID)
	maxReplicationLatency := t.maxAllowedLatencyFn()
	now := t.shard.GetTimeSource().Now()

	if t.lastTaskCreationTime.Load() == nil {
		return defaultBatchSize
	}
	taskLatency := now.Sub(t.lastTaskCreationTime.Load().(time.Time))
	if taskLatency < 0 {
		taskLatency = 0
	}
	if taskLatency >= maxReplicationLatency {
		return defaultBatchSize
	}
	return minReadTaskSize + int(float64(taskLatency)/float64(maxReplicationLatency)*float64(defaultBatchSize))
}

func skipTask(pollingCluster string, domainEntity *cache.DomainCacheEntry) bool {
	for _, cluster := range domainEntity.GetReplicationConfig().Clusters {
		if cluster.ClusterName == pollingCluster {
			return false
		}
	}
	return true
}
