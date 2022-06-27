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
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/execution"
)

type TaskHydrator struct {
	shardID int
	history HistoryManager

	logger               log.Logger
	readHistoryBatchSize dynamicconfig.IntPropertyFn
}

func NewTaskHydrator(shardID int, history HistoryManager, logger log.Logger, readHistoryBatchSize dynamicconfig.IntPropertyFn) TaskHydrator {
	return TaskHydrator{shardID, history, logger, readHistoryBatchSize}
}

// Dependencies
type (
	MutableState interface {
		IsWorkflowExecutionRunning() bool
		GetActivityInfo(int64) (*persistence.ActivityInfo, bool)
		GetVersionHistories() *persistence.VersionHistories
	}
	HistoryManager interface {
		ReadRawHistoryBranch(ctx context.Context, request *persistence.ReadHistoryBranchRequest) (*persistence.ReadRawHistoryBranchResponse, error)
	}
)

func (t TaskHydrator) HydrateFailoverMarkerTask(task *persistence.ReplicationTaskInfo) *types.ReplicationTask {
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

func (t TaskHydrator) HydrateSyncActivityTask(ctx context.Context, task *persistence.ReplicationTaskInfo, ms MutableState) (*types.ReplicationTask, error) {
	if !ms.IsWorkflowExecutionRunning() {
		// workflow already finished, no need to process the replication task
		return nil, nil
	}

	activityInfo, ok := ms.GetActivityInfo(task.ScheduledID)
	if !ok {
		return nil, nil
	}
	activityInfo = execution.CopyActivityInfo(activityInfo)

	var startedTime *int64
	var heartbeatTime *int64
	scheduledTime := timeToUnixNano(activityInfo.ScheduledTime)
	if activityInfo.StartedID != common.EmptyEventID {
		startedTime = timeToUnixNano(activityInfo.StartedTime)
	}
	// LastHeartBeatUpdatedTime must be valid when getting the sync activity replication task
	heartbeatTime = timeToUnixNano(activityInfo.LastHeartBeatUpdatedTime)

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
		TaskType:     types.ReplicationTaskTypeSyncActivity.Ptr(),
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

func (t TaskHydrator) HydrateHistoryReplicationTask(ctx context.Context, task *persistence.ReplicationTaskInfo, ms MutableState) (*types.ReplicationTask, error) {
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
