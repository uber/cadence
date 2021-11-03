// Copyright (c) 2020 Uber Technologies, Inc.
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

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination mutable_state_task_refresher_mock.go -self_package github.com/uber/cadence/service/history/execution

package execution

import (
	"context"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/events"
)

var emptyTasks = []persistence.Task{}

type (
	// MutableStateTaskRefresher refreshes workflow transfer and timer tasks
	MutableStateTaskRefresher interface {
		RefreshTasks(ctx context.Context, startTime time.Time, mutableState MutableState) error
	}

	mutableStateTaskRefresherImpl struct {
		config          *config.Config
		clusterMetadata cluster.Metadata
		domainCache     cache.DomainCache
		eventsCache     events.Cache
		logger          log.Logger
		shardID         int
	}
)

// NewMutableStateTaskRefresher creates a new task refresher for mutable state
func NewMutableStateTaskRefresher(
	config *config.Config,
	clusterMetadata cluster.Metadata,
	domainCache cache.DomainCache,
	eventsCache events.Cache,
	logger log.Logger,
	shardID int,
) MutableStateTaskRefresher {

	return &mutableStateTaskRefresherImpl{
		config:          config,
		clusterMetadata: clusterMetadata,
		domainCache:     domainCache,
		eventsCache:     eventsCache,
		logger:          logger,
		shardID:         shardID,
	}
}

func (r *mutableStateTaskRefresherImpl) RefreshTasks(
	ctx context.Context,
	startTime time.Time,
	mutableState MutableState,
) error {

	taskGenerator := NewMutableStateTaskGenerator(
		r.clusterMetadata,
		r.domainCache,
		r.logger,
		mutableState,
	)

	if err := r.refreshTasksForWorkflowStart(
		ctx,
		startTime,
		mutableState,
		taskGenerator,
	); err != nil {
		return err
	}

	if err := r.refreshTasksForWorkflowClose(
		ctx,
		mutableState,
		taskGenerator,
	); err != nil {
		return err
	}

	if err := r.refreshTasksForRecordWorkflowStarted(
		ctx,
		mutableState,
		taskGenerator,
	); err != nil {
		return err
	}

	if err := r.refreshTasksForDecision(
		ctx,
		mutableState,
		taskGenerator,
	); err != nil {
		return err
	}

	if err := r.refreshTasksForActivity(
		ctx,
		mutableState,
		taskGenerator,
	); err != nil {
		return err
	}

	if err := r.refreshTasksForTimer(
		ctx,
		mutableState,
		taskGenerator,
	); err != nil {
		return err
	}

	if err := r.refreshTasksForChildWorkflow(
		ctx,
		mutableState,
		taskGenerator,
	); err != nil {
		return err
	}

	if err := r.refreshTasksForRequestCancelExternalWorkflow(
		ctx,
		mutableState,
		taskGenerator,
	); err != nil {
		return err
	}

	if err := r.refreshTasksForSignalExternalWorkflow(
		ctx,
		mutableState,
		taskGenerator,
	); err != nil {
		return err
	}

	if r.config.AdvancedVisibilityWritingMode() != common.AdvancedVisibilityWritingModeOff {
		if err := r.refreshTasksForWorkflowSearchAttr(
			ctx,
			mutableState,
			taskGenerator,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *mutableStateTaskRefresherImpl) refreshTasksForWorkflowStart(
	ctx context.Context,
	startTime time.Time,
	mutableState MutableState,
	taskGenerator MutableStateTaskGenerator,
) error {

	startEvent, err := mutableState.GetStartEvent(ctx)
	if err != nil {
		return err
	}

	if err := taskGenerator.GenerateWorkflowStartTasks(
		startTime,
		startEvent,
	); err != nil {
		return err
	}

	startAttr := startEvent.WorkflowExecutionStartedEventAttributes
	if !mutableState.HasProcessedOrPendingDecision() && startAttr.GetFirstDecisionTaskBackoffSeconds() > 0 {
		if err := taskGenerator.GenerateDelayedDecisionTasks(
			startEvent,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *mutableStateTaskRefresherImpl) refreshTasksForWorkflowClose(
	ctx context.Context,
	mutableState MutableState,
	taskGenerator MutableStateTaskGenerator,
) error {

	executionInfo := mutableState.GetExecutionInfo()
	if executionInfo.CloseStatus != persistence.WorkflowCloseStatusNone {
		closeEvent, err := mutableState.GetCompletionEvent(ctx)
		if err != nil {
			return err
		}
		return taskGenerator.GenerateWorkflowCloseTasks(
			closeEvent,
		)
	}

	return nil
}

func (r *mutableStateTaskRefresherImpl) refreshTasksForRecordWorkflowStarted(
	ctx context.Context,
	mutableState MutableState,
	taskGenerator MutableStateTaskGenerator,
) error {

	startEvent, err := mutableState.GetStartEvent(ctx)
	if err != nil {
		return err
	}

	executionInfo := mutableState.GetExecutionInfo()

	if executionInfo.CloseStatus == persistence.WorkflowCloseStatusNone {
		return taskGenerator.GenerateRecordWorkflowStartedTasks(
			startEvent,
		)
	}

	return nil
}

func (r *mutableStateTaskRefresherImpl) refreshTasksForDecision(
	ctx context.Context,
	mutableState MutableState,
	taskGenerator MutableStateTaskGenerator,
) error {

	if !mutableState.HasPendingDecision() {
		// no decision task at all
		return nil
	}

	decision, ok := mutableState.GetPendingDecision()
	if !ok {
		return &types.InternalServiceError{Message: "it could be a bug, cannot get pending decision"}
	}

	// decision already started
	if decision.StartedID != common.EmptyEventID {
		return taskGenerator.GenerateDecisionStartTasks(
			decision.ScheduleID,
		)
	}

	// decision only scheduled
	return taskGenerator.GenerateDecisionScheduleTasks(
		decision.ScheduleID,
	)
}

func (r *mutableStateTaskRefresherImpl) refreshTasksForActivity(
	ctx context.Context,
	mutableState MutableState,
	taskGenerator MutableStateTaskGenerator,
) error {

	executionInfo := mutableState.GetExecutionInfo()
	pendingActivityInfos := mutableState.GetPendingActivityInfos()

	currentBranchToken, err := mutableState.GetCurrentBranchToken()
	if err != nil {
		return err
	}

Loop:
	for _, activityInfo := range pendingActivityInfos {
		// clear all activity timer task mask for later activity timer task re-generation
		activityInfo.TimerTaskStatus = TimerTaskStatusNone

		// need to update activity timer task mask for which task is generated
		if err := mutableState.UpdateActivity(
			activityInfo,
		); err != nil {
			return err
		}

		if activityInfo.StartedID != common.EmptyEventID {
			continue Loop
		}

		scheduleEvent, err := r.eventsCache.GetEvent(
			ctx,
			r.shardID,
			executionInfo.DomainID,
			executionInfo.WorkflowID,
			executionInfo.RunID,
			activityInfo.ScheduledEventBatchID,
			activityInfo.ScheduleID,
			currentBranchToken,
		)
		if err != nil {
			return err
		}

		if err := taskGenerator.GenerateActivityTransferTasks(
			scheduleEvent,
		); err != nil {
			return err
		}
	}

	if _, err := NewTimerSequence(
		mutableState,
	).CreateNextActivityTimer(); err != nil {
		return err
	}

	return nil
}

func (r *mutableStateTaskRefresherImpl) refreshTasksForTimer(
	ctx context.Context,
	mutableState MutableState,
	taskGenerator MutableStateTaskGenerator,
) error {

	pendingTimerInfos := mutableState.GetPendingTimerInfos()

	for _, timerInfo := range pendingTimerInfos {
		// clear all timer task mask for later timer task re-generation
		timerInfo.TaskStatus = TimerTaskStatusNone

		// need to update user timer task mask for which task is generated
		if err := mutableState.UpdateUserTimer(
			timerInfo,
		); err != nil {
			return err
		}
	}

	if _, err := NewTimerSequence(
		mutableState,
	).CreateNextUserTimer(); err != nil {
		return err
	}

	return nil
}

func (r *mutableStateTaskRefresherImpl) refreshTasksForChildWorkflow(
	ctx context.Context,
	mutableState MutableState,
	taskGenerator MutableStateTaskGenerator,
) error {

	executionInfo := mutableState.GetExecutionInfo()
	pendingChildWorkflowInfos := mutableState.GetPendingChildExecutionInfos()

	currentBranchToken, err := mutableState.GetCurrentBranchToken()
	if err != nil {
		return err
	}

Loop:
	for _, childWorkflowInfo := range pendingChildWorkflowInfos {
		if childWorkflowInfo.StartedID != common.EmptyEventID {
			continue Loop
		}

		scheduleEvent, err := r.eventsCache.GetEvent(
			ctx,
			r.shardID,
			executionInfo.DomainID,
			executionInfo.WorkflowID,
			executionInfo.RunID,
			childWorkflowInfo.InitiatedEventBatchID,
			childWorkflowInfo.InitiatedID,
			currentBranchToken,
		)
		if err != nil {
			return err
		}

		if err := taskGenerator.GenerateChildWorkflowTasks(
			scheduleEvent,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *mutableStateTaskRefresherImpl) refreshTasksForRequestCancelExternalWorkflow(
	ctx context.Context,
	mutableState MutableState,
	taskGenerator MutableStateTaskGenerator,
) error {

	executionInfo := mutableState.GetExecutionInfo()
	pendingRequestCancelInfos := mutableState.GetPendingRequestCancelExternalInfos()

	currentBranchToken, err := mutableState.GetCurrentBranchToken()
	if err != nil {
		return err
	}

	for _, requestCancelInfo := range pendingRequestCancelInfos {
		initiateEvent, err := r.eventsCache.GetEvent(
			ctx,
			r.shardID,
			executionInfo.DomainID,
			executionInfo.WorkflowID,
			executionInfo.RunID,
			requestCancelInfo.InitiatedEventBatchID,
			requestCancelInfo.InitiatedID,
			currentBranchToken,
		)
		if err != nil {
			return err
		}

		if err := taskGenerator.GenerateRequestCancelExternalTasks(
			initiateEvent,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *mutableStateTaskRefresherImpl) refreshTasksForSignalExternalWorkflow(
	ctx context.Context,
	mutableState MutableState,
	taskGenerator MutableStateTaskGenerator,
) error {

	executionInfo := mutableState.GetExecutionInfo()
	pendingSignalInfos := mutableState.GetPendingSignalExternalInfos()

	currentBranchToken, err := mutableState.GetCurrentBranchToken()
	if err != nil {
		return err
	}

	for _, signalInfo := range pendingSignalInfos {
		initiateEvent, err := r.eventsCache.GetEvent(
			ctx,
			r.shardID,
			executionInfo.DomainID,
			executionInfo.WorkflowID,
			executionInfo.RunID,
			signalInfo.InitiatedEventBatchID,
			signalInfo.InitiatedID,
			currentBranchToken,
		)
		if err != nil {
			return err
		}

		if err := taskGenerator.GenerateSignalExternalTasks(
			initiateEvent,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *mutableStateTaskRefresherImpl) refreshTasksForWorkflowSearchAttr(
	ctx context.Context,
	mutableState MutableState,
	taskGenerator MutableStateTaskGenerator,
) error {

	return taskGenerator.GenerateWorkflowSearchAttrTasks()
}
