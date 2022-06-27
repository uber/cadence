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
		executionCache   *exec.Cache
		executionManager persistence.ExecutionManager
		historyManager   persistence.HistoryManager
		rateLimiter      *quotas.DynamicRateLimiter
		retryPolicy      backoff.RetryPolicy
		throttleRetry    *backoff.ThrottleRetry

		lastTaskCreationTime atomic.Value
		maxAllowedLatencyFn  dynamicconfig.DurationPropertyFn

		metricsClient metrics.Client
		logger        log.Logger

		// This is the batch size used by pull based RPC replicator.
		fetchTasksBatchSize dynamicconfig.IntPropertyFnWithShardIDFilter
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
		executionCache:   executionCache,
		executionManager: shard.GetExecutionManager(),
		historyManager:   shard.GetHistoryManager(),
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

	switch task.TaskType {
	case persistence.ReplicationTaskTypeFailoverMarker:
		return t.generateFailoverMarkerTask(task), nil
	}

	execution := types.WorkflowExecution{
		WorkflowID: task.WorkflowID,
		RunID:      task.RunID,
	}

	context, release, err := t.executionCache.GetOrCreateWorkflowExecution(ctx, task.DomainID, execution)
	if err != nil {
		return nil, err
	}
	defer func() { release(nil) }()

	ms, err := context.LoadWorkflowExecution(ctx)
	if common.IsEntityNotExistsError(err) {
		return nil, nil
	}
	if err != nil {
		release(err)
		return nil, err
	}

	switch task.TaskType {
	case persistence.ReplicationTaskTypeSyncActivity:
		return t.generateSyncActivityTask(ctx, task, ms)
	case persistence.ReplicationTaskTypeHistory:
		return t.generateHistoryReplicationTask(ctx, task, ms)
	default:
		return nil, errUnknownReplicationTask
	}
}

func (t *taskAckManagerImpl) getEventsBlob(
	ctx context.Context,
	branchToken []byte,
	firstEventID int64,
	nextEventID int64,
) (*types.DataBlob, error) {

	var eventBatchBlobs []*persistence.DataBlob
	var pageToken []byte
	batchSize := t.shard.GetConfig().ReplicationTaskProcessorReadHistoryBatchSize()
	req := &persistence.ReadHistoryBranchRequest{
		BranchToken:   branchToken,
		MinEventID:    firstEventID,
		MaxEventID:    nextEventID,
		PageSize:      batchSize,
		NextPageToken: pageToken,
		ShardID:       common.IntPtr(t.shard.GetShardID()),
	}

	for {
		resp, err := t.historyManager.ReadRawHistoryBranch(ctx, req)
		if err != nil {
			return nil, err
		}

		req.NextPageToken = resp.NextPageToken
		eventBatchBlobs = append(eventBatchBlobs, resp.HistoryEventBlobs...)

		if len(req.NextPageToken) == 0 {
			break
		}
	}

	if len(eventBatchBlobs) != 1 {
		return nil, &types.InternalDataInconsistencyError{
			Message: "replicatorQueueProcessor encounter more than 1 NDC raw event batch",
		}
	}

	return eventBatchBlobs[0].ToInternal(), nil
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

func (t *taskAckManagerImpl) generateFailoverMarkerTask(task *persistence.ReplicationTaskInfo) *types.ReplicationTask {

	return &types.ReplicationTask{
		TaskType:     types.ReplicationTaskType.Ptr(types.ReplicationTaskTypeFailoverMarker),
		SourceTaskID: task.TaskID,
		FailoverMarkerAttributes: &types.FailoverMarkerAttributes{
			DomainID:        task.DomainID,
			FailoverVersion: task.Version,
		},
		CreationTime: common.Int64Ptr(task.CreationTime),
	}
}

func (t *taskAckManagerImpl) generateSyncActivityTask(ctx context.Context, task *persistence.ReplicationTaskInfo, ms exec.MutableState) (*types.ReplicationTask, error) {

	if !ms.IsWorkflowExecutionRunning() {
		// workflow already finished, no need to process the replication task
		return nil, nil
	}

	activityInfo, ok := ms.GetActivityInfo(task.ScheduledID)
	if !ok {
		return nil, nil
	}
	activityInfo = exec.CopyActivityInfo(activityInfo)

	var startedTime *int64
	var heartbeatTime *int64
	scheduledTime := common.Int64Ptr(activityInfo.ScheduledTime.UnixNano())
	if activityInfo.StartedID != common.EmptyEventID {
		startedTime = common.Int64Ptr(activityInfo.StartedTime.UnixNano())
	}
	// LastHeartBeatUpdatedTime must be valid when getting the sync activity replication task
	heartbeatTime = common.Int64Ptr(activityInfo.LastHeartBeatUpdatedTime.UnixNano())

	versionHistories := ms.GetVersionHistories()
	if versionHistories != nil {
		versionHistories = versionHistories.Duplicate()
	}

	//Version history uses when replicate the sync activity task
	var versionHistory *types.VersionHistory
	if versionHistories != nil {
		rawVersionHistory, err := versionHistories.GetCurrentVersionHistory()
		if err != nil {
			return nil, err
		}
		versionHistory = rawVersionHistory.ToInternalType()
	}

	return &types.ReplicationTask{
		TaskType:     types.ReplicationTaskType.Ptr(types.ReplicationTaskTypeSyncActivity),
		SourceTaskID: task.TaskID,
		SyncActivityTaskAttributes: &types.SyncActivityTaskAttributes{
			DomainID:           task.DomainID,
			WorkflowID:         task.WorkflowID,
			RunID:              task.RunID,
			Version:            activityInfo.Version,
			ScheduledID:        activityInfo.ScheduleID,
			ScheduledTime:      scheduledTime,
			StartedID:          activityInfo.StartedID,
			StartedTime:        startedTime,
			LastHeartbeatTime:  heartbeatTime,
			Details:            activityInfo.Details,
			Attempt:            activityInfo.Attempt,
			LastFailureReason:  common.StringPtr(activityInfo.LastFailureReason),
			LastWorkerIdentity: activityInfo.LastWorkerIdentity,
			LastFailureDetails: activityInfo.LastFailureDetails,
			VersionHistory:     versionHistory,
		},
		CreationTime: common.Int64Ptr(task.CreationTime),
	}, nil
}

func (t *taskAckManagerImpl) generateHistoryReplicationTask(ctx context.Context, task *persistence.ReplicationTaskInfo, ms exec.MutableState) (*types.ReplicationTask, error) {

	versionHistories := ms.GetVersionHistories()
	if versionHistories != nil {
		versionHistories = versionHistories.Duplicate()
	}

	if versionHistories == nil {
		t.logger.Error("encounter workflow without version histories",
			tag.WorkflowDomainID(task.DomainID),
			tag.WorkflowID(task.WorkflowID),
			tag.WorkflowRunID(task.RunID))
		return nil, nil
	}

	_, versionHistory, err := versionHistories.FindFirstVersionHistoryByItem(persistence.NewVersionHistoryItem(task.FirstEventID, task.Version))
	if err != nil {
		return nil, err
	}

	// BranchToken will not set in get dlq replication message request
	if len(task.BranchToken) == 0 {
		task.BranchToken = versionHistory.GetBranchToken()
	}

	eventsBlob, err := t.getEventsBlob(ctx, task.BranchToken, task.FirstEventID, task.NextEventID)
	if err != nil {
		return nil, err
	}

	var newRunEventsBlob *types.DataBlob
	if len(task.NewRunBranchToken) != 0 {
		// only get the first batch
		newRunEventsBlob, err = t.getEventsBlob(ctx, task.NewRunBranchToken, common.FirstEventID, common.FirstEventID+1)
		if err != nil {
			return nil, err
		}
	}

	return &types.ReplicationTask{
		TaskType:     types.ReplicationTaskType.Ptr(types.ReplicationTaskTypeHistoryV2),
		SourceTaskID: task.TaskID,
		HistoryTaskV2Attributes: &types.HistoryTaskV2Attributes{
			DomainID:            task.DomainID,
			WorkflowID:          task.WorkflowID,
			RunID:               task.RunID,
			VersionHistoryItems: versionHistory.ToInternalType().Items,
			Events:              eventsBlob,
			NewRunEvents:        newRunEventsBlob,
		},
		CreationTime: common.Int64Ptr(task.CreationTime),
	}, nil
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
