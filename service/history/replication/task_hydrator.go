// The MIT License (MIT)
//
// Copyright (c) 2017-2022 Uber Technologies Inc.
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

package replication

import (
	"context"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

// TaskHydrator takes raw replication tasks and hydrates them with additional information.
type TaskHydrator struct {
	shardID int
	history HistoryManager

	readHistoryBatchSize dynamicconfig.IntPropertyFn
}

// NewTaskHydrator creates new TaskHydrator.
func NewTaskHydrator(shardID int, history HistoryManager, readHistoryBatchSize dynamicconfig.IntPropertyFn) TaskHydrator {
	return TaskHydrator{shardID, history, readHistoryBatchSize}
}

// Dependencies
type (
	// MutableState is a subset of mutable state needed for hydration purposes
	MutableState interface {
		IsWorkflowExecutionRunning() bool
		GetActivityInfo(int64) (*persistence.ActivityInfo, bool)
		GetVersionHistories() *persistence.VersionHistories
	}
	// HistoryManager is a subset of history manager needed for hydration purposes
	HistoryManager interface {
		ReadRawHistoryBranch(ctx context.Context, request *persistence.ReadHistoryBranchRequest) (*persistence.ReadRawHistoryBranchResponse, error)
	}
)

// HydrateFailoverMarkerTask hydrates failover marker replication task.
func (t TaskHydrator) HydrateFailoverMarkerTask(task persistence.ReplicationTaskInfo) *types.ReplicationTask {
	return &types.ReplicationTask{
		TaskType:     types.ReplicationTaskTypeFailoverMarker.Ptr(),
		SourceTaskID: task.TaskID,
		FailoverMarkerAttributes: &types.FailoverMarkerAttributes{
			DomainID:        task.DomainID,
			FailoverVersion: task.Version,
		},
		CreationTime: common.Int64Ptr(task.CreationTime),
	}
}

// HydrateSyncActivityTask hydrates sync activity replication task.
// It needs loaded mutable state to hydrate fields for this task.
func (t TaskHydrator) HydrateSyncActivityTask(ctx context.Context, task persistence.ReplicationTaskInfo, ms MutableState) (*types.ReplicationTask, error) {
	// Treat nil mutable state as if workflow does not exist (no longer exists)
	if ms == nil {
		return nil, nil
	}

	if !ms.IsWorkflowExecutionRunning() {
		// workflow already finished, no need to process the replication task
		return nil, nil
	}

	activityInfo, ok := ms.GetActivityInfo(task.ScheduledID)
	if !ok {
		return nil, nil
	}

	var startedTime *int64
	if activityInfo.StartedID != common.EmptyEventID {
		startedTime = timeToUnixNano(activityInfo.StartedTime)
	}

	//Version history uses when replicate the sync activity task
	var versionHistory *types.VersionHistory
	if versionHistories := ms.GetVersionHistories(); versionHistories != nil {
		currentVersionHistory, err := versionHistories.GetCurrentVersionHistory()
		if err != nil {
			return nil, err
		}
		versionHistory = currentVersionHistory.ToInternalType()
	}

	return &types.ReplicationTask{
		TaskType:     types.ReplicationTaskTypeSyncActivity.Ptr(),
		SourceTaskID: task.TaskID,
		SyncActivityTaskAttributes: &types.SyncActivityTaskAttributes{
			DomainID:           task.DomainID,
			WorkflowID:         task.WorkflowID,
			RunID:              task.RunID,
			Version:            activityInfo.Version,
			ScheduledID:        activityInfo.ScheduleID,
			ScheduledTime:      timeToUnixNano(activityInfo.ScheduledTime),
			StartedID:          activityInfo.StartedID,
			StartedTime:        startedTime,
			LastHeartbeatTime:  timeToUnixNano(activityInfo.LastHeartBeatUpdatedTime),
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

// HydrateHistoryReplicationTask hydrates history replication task.
// It needs version histories to load history branch from database with events specified in replication task.
func (t TaskHydrator) HydrateHistoryReplicationTask(ctx context.Context, task persistence.ReplicationTaskInfo, versionHistories *persistence.VersionHistories) (*types.ReplicationTask, error) {
	if versionHistories == nil {
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
		TaskType:     types.ReplicationTaskTypeHistoryV2.Ptr(),
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

func (t TaskHydrator) getEventsBlob(ctx context.Context, branchToken []byte, firstEventID int64, nextEventID int64) (*types.DataBlob, error) {

	var eventBatchBlobs []*persistence.DataBlob
	var pageToken []byte
	req := &persistence.ReadHistoryBranchRequest{
		BranchToken:   branchToken,
		MinEventID:    firstEventID,
		MaxEventID:    nextEventID,
		PageSize:      t.readHistoryBatchSize(),
		NextPageToken: pageToken,
		ShardID:       &t.shardID,
	}

	for {
		resp, err := t.history.ReadRawHistoryBranch(ctx, req)
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

func timeToUnixNano(t time.Time) *int64 {
	return common.Int64Ptr(t.UnixNano())
}
