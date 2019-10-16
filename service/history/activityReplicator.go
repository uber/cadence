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

//go:generate mockgen -copyright_file ../../LICENSE -package $GOPACKAGE -source $GOFILE -destination activityReplicator_mock.go

package history

import (
	ctx "context"
	"time"

	"go.uber.org/cadence/.gen/go/shared"

	h "github.com/uber/cadence/.gen/go/history"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
)

type (
	activityReplicator interface {
		SyncActivity(
			ctx ctx.Context,
			request *h.SyncActivityRequest,
		) error
	}

	activityReplicatorImpl struct {
		historyCache    *historyCache
		clusterMetadata cluster.Metadata
		logger          log.Logger
	}
)

func newActivityReplicator(
	shard ShardContext,
	historyCache *historyCache,
	logger log.Logger,
) *activityReplicatorImpl {

	return &activityReplicatorImpl{
		historyCache:    historyCache,
		clusterMetadata: shard.GetService().GetClusterMetadata(),
		logger:          logger.WithTags(tag.ComponentHistoryReplicator),
	}
}

func (r *activityReplicatorImpl) SyncActivity(
	ctx ctx.Context,
	request *h.SyncActivityRequest,
) (retError error) {

	// sync activity info will only be sent from active side, when
	// 1. activity has retry policy and activity got started
	// 2. activity heart beat
	// no sync activity task will be sent when active side fail / timeout activity,
	// since standby side does not have activity retry timer

	domainID := request.GetDomainId()
	execution := workflow.WorkflowExecution{
		WorkflowId: request.WorkflowId,
		RunId:      request.RunId,
	}

	context, release, err := r.historyCache.getOrCreateWorkflowExecution(ctx, domainID, execution)
	if err != nil {
		// for get workflow execution context, with valid run id
		// err will not be of type EntityNotExistsError
		return err
	}
	defer func() { release(retError) }()

	msBuilder, err := context.loadWorkflowExecution()
	if err != nil {
		if _, ok := err.(*workflow.EntityNotExistsError); !ok {
			return err
		}

		// this can happen if the workflow start event and this sync activity task are out of order
		// or the target workflow is long gone
		// the safe solution to this is to throw away the sync activity task
		// or otherwise, worker attempt will exceeds limit and put this message to DLQ
		return nil
	}

	if !msBuilder.IsWorkflowExecutionRunning() {
		// perhaps conflict resolution force termination
		return nil
	}

	version := request.GetVersion()
	scheduleID := request.GetScheduledId()
	if scheduleID >= msBuilder.GetNextEventID() {
		lastWriteVersion, err := msBuilder.GetLastWriteVersion()
		if err != nil {
			return err
		}
		if version < lastWriteVersion {
			// activity version < workflow last write version
			// this can happen if target workflow has
			return nil
		}

		// TODO when 2DC is deprecated, remove this block
		if msBuilder.GetReplicationState() != nil {
			// version >= last write version
			// this can happen if out of order delivery happens
			return newRetryTaskErrorWithHint(
				ErrRetrySyncActivityMsg,
				domainID,
				execution.GetWorkflowId(),
				execution.GetRunId(),
				msBuilder.GetNextEventID(),
			)
		}

		if request.VersionHistory == nil {
			// sanity check on 2DC or 3DC
			r.logger.Error(
				"The workflow is not 2DC or 3DC enabled.",
				tag.WorkflowID(execution.GetWorkflowId()),
				tag.WorkflowRunID(execution.GetRunId()),
			)
			return &shared.InternalServiceError{Message: "The workflow is not 2DC or 3DC enabled."}
		}
		// compare LCA of the incoming version history with local version history
		currentVersionHistory, err := msBuilder.GetVersionHistories().GetCurrentVersionHistory()
		if err != nil {
			return err
		}
		incomingVersionHistory := persistence.NewVersionHistoryFromThrift(request.GetVersionHistory())
		lcaItem, err := incomingVersionHistory.FindLCAItem(currentVersionHistory)
		if err != nil {
			return err
		}
		lastIncomingItem, err := incomingVersionHistory.GetLastItem()
		if err != nil {
			return err
		}

		// Send history from the LCA item to the last item on the incoming branch
		return newNDCRetryTaskErrorWithHint(
			domainID,
			execution.GetWorkflowId(),
			execution.GetRunId(),
			common.Int64Ptr(lcaItem.GetEventID()),
			common.Int64Ptr(lcaItem.GetVersion()),
			common.Int64Ptr(lastIncomingItem.GetEventID()),
			common.Int64Ptr(lastIncomingItem.GetVersion()),
		)
	}

	ai, ok := msBuilder.GetActivityInfo(scheduleID)
	if !ok {
		// this should not retry, can be caused by out of order delivery
		// since the activity is already finished
		return nil
	}

	if ai.Version > request.GetVersion() {
		// this should not retry, can be caused by failover or reset
		return nil
	}

	if ai.Version == request.GetVersion() {
		if ai.Attempt > request.GetAttempt() {
			// this should not retry, can be caused by failover or reset
			return nil
		}
		if ai.Attempt == request.GetAttempt() {
			lastHeartbeatTime := time.Unix(0, request.GetLastHeartbeatTime())
			if ai.LastHeartBeatUpdatedTime.After(lastHeartbeatTime) {
				// this should not retry, can be caused by out of order delivery
				return nil
			}
			// version equal & attempt equal & last heartbeat after existing heartbeat
			// should update activity
		}
		// version equal & attempt larger then existing, should update activity
	}
	// version larger then existing, should update activity

	// calculate whether to reset the activity timer task status bits
	// reset timer task status bits if
	// 1. same source cluster & attempt changes
	// 2. different source cluster
	resetActivityTimerTaskStatus := false
	if !r.clusterMetadata.IsVersionFromSameCluster(request.GetVersion(), ai.Version) {
		resetActivityTimerTaskStatus = true
	} else if ai.Attempt < request.GetAttempt() {
		resetActivityTimerTaskStatus = true
	}
	err = msBuilder.ReplicateActivityInfo(request, resetActivityTimerTaskStatus)
	if err != nil {
		return err
	}

	// see whether we need to refresh the activity timer
	eventTime := request.GetScheduledTime()
	if eventTime < request.GetStartedTime() {
		eventTime = request.GetStartedTime()
	}
	if eventTime < request.GetLastHeartbeatTime() {
		eventTime = request.GetLastHeartbeatTime()
	}
	now := time.Unix(0, eventTime)
	timerTasks := []persistence.Task{}
	timeSource := clock.NewEventTimeSource()
	timeSource.Update(now)
	timerBuilder := newTimerBuilder(timeSource)
	if tt := timerBuilder.GetActivityTimerTaskIfNeeded(msBuilder); tt != nil {
		timerTasks = append(timerTasks, tt)
	}

	msBuilder.AddTimerTasks(timerTasks...)
	return context.updateWorkflowExecutionAsPassive(now)
}
