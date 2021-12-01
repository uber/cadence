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

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination mutable_state_task_generator_mock.go -self_package github.com/uber/cadence/service/history/execution

package execution

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

type (
	// MutableStateTaskGenerator generates workflow transfer and timer tasks
	MutableStateTaskGenerator interface {
		// for workflow reset startTime should be the reset time instead of
		// the startEvent time, so that workflow timeout timestamp can be
		// re-calculated.
		GenerateWorkflowStartTasks(
			startTime time.Time,
			startEvent *types.HistoryEvent,
		) error
		GenerateWorkflowCloseTasks(
			closeEvent *types.HistoryEvent,
		) error
		GenerateRecordWorkflowStartedTasks(
			startEvent *types.HistoryEvent,
		) error
		GenerateDelayedDecisionTasks(
			startEvent *types.HistoryEvent,
		) error
		GenerateDecisionScheduleTasks(
			decisionScheduleID int64,
		) error
		GenerateDecisionStartTasks(
			decisionScheduleID int64,
		) error
		GenerateActivityTransferTasks(
			event *types.HistoryEvent,
		) error
		GenerateActivityRetryTasks(
			activityScheduleID int64,
		) error
		GenerateChildWorkflowTasks(
			event *types.HistoryEvent,
		) error
		GenerateRequestCancelExternalTasks(
			event *types.HistoryEvent,
		) error
		GenerateSignalExternalTasks(
			event *types.HistoryEvent,
		) error
		GenerateWorkflowSearchAttrTasks() error
		GenerateWorkflowResetTasks() error
		GenerateFromTransferTask(
			transferTask *persistence.TransferTaskInfo,
			targetCluster string,
		) error
		// NOTE: CloseExecution task may generate both RecordChildCompletion
		// and ApplyParentPolicy tasks. That's why we currently have separate
		// functions for generating tasks from CloseExecution task
		GenerateCrossClusterRecordChildCompletedTask(
			transferTask *persistence.TransferTaskInfo,
			targetCluster string,
			parentInfo *types.ParentExecutionInfo,
		) error
		GenerateCrossClusterApplyParentClosePolicyTask(
			transferTask *persistence.TransferTaskInfo,
			targetCluster string,
			childDomainIDs map[string]struct{},
		) error
		GenerateFromCrossClusterTask(
			crossClusterTask *persistence.CrossClusterTaskInfo,
		) error

		// these 2 APIs should only be called when mutable state transaction is being closed
		GenerateActivityTimerTasks() error
		GenerateUserTimerTasks() error
	}

	mutableStateTaskGeneratorImpl struct {
		clusterMetadata cluster.Metadata
		domainCache     cache.DomainCache
		logger          log.Logger

		mutableState MutableState
	}
)

const (
	defaultWorkflowRetentionInDays      int32 = 1
	defaultInitIntervalForDecisionRetry       = 1 * time.Minute
	defaultMaxIntervalForDecisionRetry        = 5 * time.Minute
	defaultJitterCoefficient                  = 0.2
)

var _ MutableStateTaskGenerator = (*mutableStateTaskGeneratorImpl)(nil)

// NewMutableStateTaskGenerator creates a new task generator for mutable state
func NewMutableStateTaskGenerator(
	clusterMetadata cluster.Metadata,
	domainCache cache.DomainCache,
	logger log.Logger,
	mutableState MutableState,
) MutableStateTaskGenerator {

	return &mutableStateTaskGeneratorImpl{
		clusterMetadata: clusterMetadata,
		domainCache:     domainCache,
		logger:          logger,

		mutableState: mutableState,
	}
}

func (r *mutableStateTaskGeneratorImpl) GenerateWorkflowStartTasks(
	startTime time.Time,
	startEvent *types.HistoryEvent,
) error {
	attr := startEvent.WorkflowExecutionStartedEventAttributes
	firstDecisionDelayDuration := time.Duration(attr.GetFirstDecisionTaskBackoffSeconds()) * time.Second

	executionInfo := r.mutableState.GetExecutionInfo()
	startVersion := startEvent.GetVersion()

	workflowTimeoutDuration := time.Duration(executionInfo.WorkflowTimeout) * time.Second
	workflowTimeoutTimestamp := startTime.Add(workflowTimeoutDuration + firstDecisionDelayDuration)
	// ensure that the first attempt does not time out early based on retry policy timeout
	if attr.Attempt > 0 && !executionInfo.ExpirationTime.IsZero() && workflowTimeoutTimestamp.After(executionInfo.ExpirationTime) {
		workflowTimeoutTimestamp = executionInfo.ExpirationTime
	}
	r.mutableState.AddTimerTasks(&persistence.WorkflowTimeoutTask{
		// TaskID is set by shard
		VisibilityTimestamp: workflowTimeoutTimestamp,
		Version:             startVersion,
	})

	return nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateWorkflowCloseTasks(
	closeEvent *types.HistoryEvent,
) error {

	executionInfo := r.mutableState.GetExecutionInfo()
	transferTasks := []persistence.Task{}
	crossClusterTasks := []persistence.Task{}
	_, isActive, err := getTargetCluster(executionInfo.DomainID, r.domainCache)
	if err != nil {
		return err
	}

	if !isActive {
		// source domain passive, generate only one close execution task
		transferTasks = append(transferTasks, &persistence.CloseExecutionTask{
			// TaskID and VisibilityTimestamp are set by shard context
			Version: closeEvent.GetVersion(),
		})
	} else {
		// 1. check if parent is cross cluster
		parentTargetCluster, isActive, err := getParentCluster(r.mutableState, r.domainCache)
		if err != nil {
			return err
		}

		if parentTargetCluster != "" {
			recordChildCompletionTask := &persistence.RecordChildExecutionCompletedTask{
				TargetDomainID:   executionInfo.ParentDomainID,
				TargetWorkflowID: executionInfo.ParentWorkflowID,
				TargetRunID:      executionInfo.ParentRunID,
				Version:          closeEvent.GetVersion(),
			}

			if !isActive {
				crossClusterTasks = append(crossClusterTasks, &persistence.CrossClusterRecordChildExecutionCompletedTask{
					TargetCluster:                     parentTargetCluster,
					RecordChildExecutionCompletedTask: *recordChildCompletionTask,
				})
			} else {
				transferTasks = append(transferTasks, recordChildCompletionTask)
			}
		}

		// 2. check if child domains are cross cluster
		parentCloseTransferTask,
			parentCloseCrossClusterTask,
			err := r.generateApplyParentCloseTasks(nil, closeEvent.GetVersion(), time.Time{}, false)
		if err != nil {
			return err
		}
		transferTasks = append(transferTasks, parentCloseTransferTask...)
		crossClusterTasks = append(crossClusterTasks, parentCloseCrossClusterTask...)

		// 3. add record workflow closed task
		if len(crossClusterTasks) != 0 {
			transferTasks = append(transferTasks, &persistence.RecordWorkflowClosedTask{
				Version: closeEvent.GetVersion(),
			})
		} else {
			transferTasks = []persistence.Task{
				&persistence.CloseExecutionTask{
					Version: closeEvent.GetVersion(),
				},
			}
		}
	}

	r.mutableState.AddTransferTasks(transferTasks...)
	r.mutableState.AddCrossClusterTasks(crossClusterTasks...)

	retentionInDays := defaultWorkflowRetentionInDays
	domainEntry, err := r.domainCache.GetDomainByID(executionInfo.DomainID)
	switch err.(type) {
	case nil:
		retentionInDays = domainEntry.GetRetentionDays(executionInfo.WorkflowID)
	case *types.EntityNotExistsError:
		// domain is not accessible, use default value above
	default:
		return err
	}

	closeTimestamp := time.Unix(0, closeEvent.GetTimestamp())
	retentionDuration := time.Duration(retentionInDays) * time.Hour * 24
	r.mutableState.AddTimerTasks(&persistence.DeleteHistoryEventTask{
		// TaskID is set by shard
		VisibilityTimestamp: closeTimestamp.Add(retentionDuration),
		Version:             closeEvent.GetVersion(),
	})

	return nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateDelayedDecisionTasks(
	startEvent *types.HistoryEvent,
) error {

	startVersion := startEvent.GetVersion()
	startTimestamp := time.Unix(0, startEvent.GetTimestamp())
	startAttr := startEvent.WorkflowExecutionStartedEventAttributes
	decisionBackoffDuration := time.Duration(startAttr.GetFirstDecisionTaskBackoffSeconds()) * time.Second
	executionTimestamp := startTimestamp.Add(decisionBackoffDuration)

	// noParentWorkflow case
	firstDecisionDelayType := persistence.WorkflowBackoffTimeoutTypeCron
	// continue as new case
	if startAttr.Initiator != nil {
		switch startAttr.GetInitiator() {
		case types.ContinueAsNewInitiatorRetryPolicy:
			firstDecisionDelayType = persistence.WorkflowBackoffTimeoutTypeRetry
		case types.ContinueAsNewInitiatorCronSchedule:
			firstDecisionDelayType = persistence.WorkflowBackoffTimeoutTypeCron
		case types.ContinueAsNewInitiatorDecider:
			return &types.InternalServiceError{
				Message: "encounter continue as new iterator & first decision delay not 0",
			}
		default:
			return &types.InternalServiceError{
				Message: fmt.Sprintf("unknown iterator retry policy: %v", startAttr.GetInitiator()),
			}
		}
	}

	r.mutableState.AddTimerTasks(&persistence.WorkflowBackoffTimerTask{
		// TaskID is set by shard
		// TODO EventID seems not used at all
		VisibilityTimestamp: executionTimestamp,
		TimeoutType:         firstDecisionDelayType,
		Version:             startVersion,
	})

	return nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateRecordWorkflowStartedTasks(
	startEvent *types.HistoryEvent,
) error {

	startVersion := startEvent.GetVersion()

	r.mutableState.AddTransferTasks(&persistence.RecordWorkflowStartedTask{
		// TaskID and VisibilityTimestamp are set by shard context
		Version: startVersion,
	})

	return nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateDecisionScheduleTasks(
	decisionScheduleID int64,
) error {

	executionInfo := r.mutableState.GetExecutionInfo()
	decision, ok := r.mutableState.GetDecisionInfo(
		decisionScheduleID,
	)
	if !ok {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("it could be a bug, cannot get pending decision: %v", decisionScheduleID),
		}
	}

	r.mutableState.AddTransferTasks(&persistence.DecisionTask{
		// TaskID and VisibilityTimestamp are set by shard context
		DomainID:   executionInfo.DomainID,
		TaskList:   decision.TaskList,
		ScheduleID: decision.ScheduleID,
		Version:    decision.Version,
	})

	if scheduleToStartTimeout := r.mutableState.GetDecisionScheduleToStartTimeout(); scheduleToStartTimeout != 0 {
		scheduledTime := time.Unix(0, decision.ScheduledTimestamp)
		r.mutableState.AddTimerTasks(&persistence.DecisionTimeoutTask{
			// TaskID is set by shard
			VisibilityTimestamp: scheduledTime.Add(scheduleToStartTimeout),
			TimeoutType:         int(TimerTypeScheduleToStart),
			EventID:             decision.ScheduleID,
			ScheduleAttempt:     decision.Attempt,
			Version:             decision.Version,
		})
	}

	return nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateDecisionStartTasks(
	decisionScheduleID int64,
) error {

	decision, ok := r.mutableState.GetDecisionInfo(
		decisionScheduleID,
	)
	if !ok {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("it could be a bug, cannot get pending decision: %v", decisionScheduleID),
		}
	}

	startedTime := time.Unix(0, decision.StartedTimestamp)
	startToCloseTimeout := time.Duration(
		decision.DecisionTimeout,
	) * time.Second

	// schedule timer exponentially if decision keeps failing
	if decision.Attempt > 1 {
		defaultStartToCloseTimeout := r.mutableState.GetExecutionInfo().DecisionStartToCloseTimeout
		startToCloseTimeout = getNextDecisionTimeout(decision.Attempt, time.Duration(defaultStartToCloseTimeout)*time.Second)
		decision.DecisionTimeout = int32(startToCloseTimeout.Seconds()) // override decision timeout
		r.mutableState.UpdateDecision(decision)
	}

	r.mutableState.AddTimerTasks(&persistence.DecisionTimeoutTask{
		// TaskID is set by shard
		VisibilityTimestamp: startedTime.Add(startToCloseTimeout),
		TimeoutType:         int(TimerTypeStartToClose),
		EventID:             decision.ScheduleID,
		ScheduleAttempt:     decision.Attempt,
		Version:             decision.Version,
	})

	return nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateActivityTransferTasks(
	event *types.HistoryEvent,
) error {

	attr := event.ActivityTaskScheduledEventAttributes
	activityScheduleID := event.GetEventID()

	activityInfo, ok := r.mutableState.GetActivityInfo(activityScheduleID)
	if !ok {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("it could be a bug, cannot get pending activity: %v", activityScheduleID),
		}
	}

	var targetDomainID string
	var err error
	if activityInfo.DomainID != "" {
		targetDomainID = activityInfo.DomainID
	} else {
		// TODO remove this block after Mar, 1th, 2020
		//  previously, DomainID in activity info is not used, so need to get
		//  schedule event from DB checking whether activity to be scheduled
		//  belongs to this domain
		targetDomainID, err = r.getTargetDomainID(attr.GetDomain())
		if err != nil {
			return err
		}
	}

	r.mutableState.AddTransferTasks(&persistence.ActivityTask{
		// TaskID and VisibilityTimestamp are set by shard context
		DomainID:   targetDomainID,
		TaskList:   activityInfo.TaskList,
		ScheduleID: activityInfo.ScheduleID,
		Version:    activityInfo.Version,
	})

	return nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateActivityRetryTasks(
	activityScheduleID int64,
) error {

	ai, ok := r.mutableState.GetActivityInfo(activityScheduleID)
	if !ok {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("it could be a bug, cannot get pending activity: %v", activityScheduleID),
		}
	}

	r.mutableState.AddTimerTasks(&persistence.ActivityRetryTimerTask{
		// TaskID is set by shard
		Version:             ai.Version,
		VisibilityTimestamp: ai.ScheduledTime,
		EventID:             ai.ScheduleID,
		Attempt:             ai.Attempt,
	})
	return nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateChildWorkflowTasks(
	event *types.HistoryEvent,
) error {

	attr := event.StartChildWorkflowExecutionInitiatedEventAttributes
	childWorkflowScheduleID := event.GetEventID()
	childWorkflowTargetDomain := attr.GetDomain()

	childWorkflowInfo, ok := r.mutableState.GetChildExecutionInfo(childWorkflowScheduleID)
	if !ok {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("it could be a bug, cannot get pending child workflow: %v", childWorkflowScheduleID),
		}
	}

	targetDomainID := childWorkflowInfo.DomainID
	if targetDomainID == "" {
		var err error
		targetDomainID, err = r.getTargetDomainID(childWorkflowTargetDomain)
		if err != nil {
			return err
		}
	}

	targetCluster, isCrossClusterTask, err := r.isCrossClusterTask(targetDomainID)
	if err != nil {
		return err
	}

	startChildExecutionTask := &persistence.StartChildExecutionTask{
		// TaskID and VisibilityTimestamp are set by shard context
		TargetDomainID:   targetDomainID,
		TargetWorkflowID: childWorkflowInfo.StartedWorkflowID,
		InitiatedID:      childWorkflowInfo.InitiatedID,
		Version:          childWorkflowInfo.Version,
	}

	if !isCrossClusterTask {
		r.mutableState.AddTransferTasks(startChildExecutionTask)
	} else {
		r.mutableState.AddCrossClusterTasks(&persistence.CrossClusterStartChildExecutionTask{
			TargetCluster:           targetCluster,
			StartChildExecutionTask: *startChildExecutionTask,
		})
	}

	return nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateRequestCancelExternalTasks(
	event *types.HistoryEvent,
) error {

	attr := event.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes
	scheduleID := event.GetEventID()
	version := event.GetVersion()
	targetDomainName := attr.GetDomain()
	targetWorkflowID := attr.GetWorkflowExecution().GetWorkflowID()
	targetRunID := attr.GetWorkflowExecution().GetRunID()
	targetChildOnly := attr.GetChildWorkflowOnly()

	_, ok := r.mutableState.GetRequestCancelInfo(scheduleID)
	if !ok {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("it could be a bug, cannot get pending request cancel external workflow: %v", scheduleID),
		}
	}

	targetDomainID, err := r.getTargetDomainID(targetDomainName)
	if err != nil {
		return err
	}

	targetCluster, isCrossClusterTask, err := r.isCrossClusterTask(targetDomainID)
	if err != nil {
		return err
	}

	cancelExecutionTask := &persistence.CancelExecutionTask{
		// TaskID and VisibilityTimestamp are set by shard context
		TargetDomainID:          targetDomainID,
		TargetWorkflowID:        targetWorkflowID,
		TargetRunID:             targetRunID,
		TargetChildWorkflowOnly: targetChildOnly,
		InitiatedID:             scheduleID,
		Version:                 version,
	}

	if !isCrossClusterTask {
		r.mutableState.AddTransferTasks(cancelExecutionTask)
	} else {
		r.mutableState.AddCrossClusterTasks(&persistence.CrossClusterCancelExecutionTask{
			TargetCluster:       targetCluster,
			CancelExecutionTask: *cancelExecutionTask,
		})
	}

	return nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateSignalExternalTasks(
	event *types.HistoryEvent,
) error {

	attr := event.SignalExternalWorkflowExecutionInitiatedEventAttributes
	scheduleID := event.GetEventID()
	version := event.GetVersion()
	targetDomainName := attr.GetDomain()
	targetWorkflowID := attr.GetWorkflowExecution().GetWorkflowID()
	targetRunID := attr.GetWorkflowExecution().GetRunID()
	targetChildOnly := attr.GetChildWorkflowOnly()

	_, ok := r.mutableState.GetSignalInfo(scheduleID)
	if !ok {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("it could be a bug, cannot get pending signal external workflow: %v", scheduleID),
		}
	}

	targetDomainID, err := r.getTargetDomainID(targetDomainName)
	if err != nil {
		return err
	}

	targetCluster, isCrossClusterTask, err := r.isCrossClusterTask(targetDomainID)
	if err != nil {
		return err
	}

	signalExecutionTask := &persistence.SignalExecutionTask{
		// TaskID and VisibilityTimestamp are set by shard context
		TargetDomainID:          targetDomainID,
		TargetWorkflowID:        targetWorkflowID,
		TargetRunID:             targetRunID,
		TargetChildWorkflowOnly: targetChildOnly,
		InitiatedID:             scheduleID,
		Version:                 version,
	}

	if !isCrossClusterTask {
		r.mutableState.AddTransferTasks(signalExecutionTask)
	} else {
		r.mutableState.AddCrossClusterTasks(&persistence.CrossClusterSignalExecutionTask{
			TargetCluster:       targetCluster,
			SignalExecutionTask: *signalExecutionTask,
		})
	}

	return nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateWorkflowSearchAttrTasks() error {

	currentVersion := r.mutableState.GetCurrentVersion()

	r.mutableState.AddTransferTasks(&persistence.UpsertWorkflowSearchAttributesTask{
		// TaskID and VisibilityTimestamp are set by shard context
		Version: currentVersion, // task processing does not check this version
	})

	return nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateWorkflowResetTasks() error {

	currentVersion := r.mutableState.GetCurrentVersion()

	r.mutableState.AddTransferTasks(&persistence.ResetWorkflowTask{
		// TaskID and VisibilityTimestamp are set by shard context
		Version: currentVersion,
	})

	return nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateCrossClusterRecordChildCompletedTask(
	task *persistence.TransferTaskInfo,
	targetCluster string,
	parentInfo *types.ParentExecutionInfo,
) error {
	if targetCluster == r.clusterMetadata.GetCurrentClusterName() {
		// this should not happen
		return errors.New("unable to create cross-cluster task for current cluster")
	}

	if parentInfo != nil {
		r.mutableState.AddCrossClusterTasks(&persistence.CrossClusterRecordChildExecutionCompletedTask{
			TargetCluster: targetCluster,
			RecordChildExecutionCompletedTask: persistence.RecordChildExecutionCompletedTask{
				VisibilityTimestamp: task.VisibilityTimestamp,
				TargetDomainID:      parentInfo.DomainUUID,
				TargetWorkflowID:    parentInfo.GetExecution().GetWorkflowID(),
				TargetRunID:         parentInfo.GetExecution().GetRunID(),
				Version:             task.Version,
			},
		})
	}

	return nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateCrossClusterApplyParentClosePolicyTask(
	task *persistence.TransferTaskInfo,
	targetCluster string,
	childDomainIDs map[string]struct{},
) error {
	if targetCluster == r.clusterMetadata.GetCurrentClusterName() {
		// this should not happen
		return errors.New("unable to create cross-cluster task for current cluster")
	}

	if len(childDomainIDs) != 0 {
		r.mutableState.AddCrossClusterTasks(&persistence.CrossClusterApplyParentClosePolicyTask{
			TargetCluster: targetCluster,
			ApplyParentClosePolicyTask: persistence.ApplyParentClosePolicyTask{
				TargetDomainIDs: childDomainIDs,
				// TaskID is set by shard context
				// Domain, workflow and run ids will be collected from mutableState
				// when processing the apply parent policy tasks.
				Version:             task.Version,
				VisibilityTimestamp: task.VisibilityTimestamp,
			},
		})
	}

	return nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateFromTransferTask(
	task *persistence.TransferTaskInfo,
	targetCluster string,
) error {
	if targetCluster == r.clusterMetadata.GetCurrentClusterName() {
		// this should not happen
		return errors.New("unable to create cross-cluster task for current cluster")
	}

	var crossClusterTask persistence.Task
	switch task.TaskType {
	case persistence.TransferTaskTypeCancelExecution:
		crossClusterTask = &persistence.CrossClusterCancelExecutionTask{
			TargetCluster: targetCluster,
			CancelExecutionTask: persistence.CancelExecutionTask{
				// TaskID is set by shard context
				TargetDomainID:          task.TargetDomainID,
				TargetWorkflowID:        task.TargetWorkflowID,
				TargetRunID:             task.TargetRunID,
				TargetChildWorkflowOnly: task.TargetChildWorkflowOnly,
				InitiatedID:             task.ScheduleID,
				Version:                 task.Version,
			},
		}
	case persistence.TransferTaskTypeSignalExecution:
		crossClusterTask = &persistence.CrossClusterSignalExecutionTask{
			TargetCluster: targetCluster,
			SignalExecutionTask: persistence.SignalExecutionTask{
				// TaskID is set by shard context
				TargetDomainID:          task.TargetDomainID,
				TargetWorkflowID:        task.TargetWorkflowID,
				TargetRunID:             task.TargetRunID,
				TargetChildWorkflowOnly: task.TargetChildWorkflowOnly,
				InitiatedID:             task.ScheduleID,
				Version:                 task.Version,
			},
		}
	case persistence.TransferTaskTypeStartChildExecution:
		crossClusterTask = &persistence.CrossClusterStartChildExecutionTask{
			TargetCluster: targetCluster,
			StartChildExecutionTask: persistence.StartChildExecutionTask{
				// TaskID is set by shard context
				TargetDomainID:   task.TargetDomainID,
				TargetWorkflowID: task.TargetWorkflowID,
				InitiatedID:      task.ScheduleID,
				Version:          task.Version,
			},
		}
	// TransferTaskTypeCloseExecution,
	// TransferTaskTypeRecordChildExecutionCompleted,
	// TransferTaskTypeApplyParentClosePolicy are handled by GenerateCrossClusterRecordChildCompletedTask
	default:
		return fmt.Errorf("unable to convert transfer task of type %v to cross-cluster task", task.TaskType)
	}

	// set visibility timestamp here so we the metric for task latency
	// can include the latency for the original transfer task.
	crossClusterTask.SetVisibilityTimestamp(task.VisibilityTimestamp)
	r.mutableState.AddCrossClusterTasks(crossClusterTask)

	return nil
}

func (r *mutableStateTaskGeneratorImpl) generateApplyParentCloseTasks(
	childDomainIDs map[string]struct{},
	version int64,
	visibilityTimestamp time.Time,
	isPassive bool,
) ([]persistence.Task, []persistence.Task, error) {
	transferTasks := []persistence.Task{}
	crossClusterTasks := []persistence.Task{}

	if isPassive {
		transferTasks = []persistence.Task{
			&persistence.ApplyParentClosePolicyTask{
				// TaskID is set by shard context
				VisibilityTimestamp: visibilityTimestamp,
				TargetDomainIDs:     childDomainIDs,
				Version:             version,
			},
		}
		return transferTasks, crossClusterTasks, nil
	}

	sameClusterDomainIDs, remoteClusterDomainIDs, err := getChildrenClusters(childDomainIDs, r.mutableState, r.domainCache)
	if err != nil {
		return nil, nil, err
	}

	if len(sameClusterDomainIDs) != 0 {
		transferTasks = append(transferTasks, &persistence.ApplyParentClosePolicyTask{
			VisibilityTimestamp: visibilityTimestamp,
			TargetDomainIDs:     sameClusterDomainIDs,
			Version:             version,
		})
	}
	for remoteCluster, domainIDs := range remoteClusterDomainIDs {
		crossClusterTasks = append(crossClusterTasks, &persistence.CrossClusterApplyParentClosePolicyTask{
			TargetCluster: remoteCluster,
			ApplyParentClosePolicyTask: persistence.ApplyParentClosePolicyTask{
				VisibilityTimestamp: visibilityTimestamp,
				TargetDomainIDs:     domainIDs,
				Version:             version,
			},
		})
	}

	return transferTasks, crossClusterTasks, nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateFromCrossClusterTask(
	task *persistence.CrossClusterTaskInfo,
) error {
	generateTransferTask := false
	var targetCluster string

	sourceDomainEntry := r.mutableState.GetDomainEntry()
	if !sourceDomainEntry.IsDomainActive() && !sourceDomainEntry.IsDomainPendingActive() {
		// domain is passive, generate (passive) transfer task
		generateTransferTask = true
	}

	// ApplyParentClosePolicy is different than the others because it doesn't have a single
	// target domain, workflow or run id.
	if task.GetTaskType() == persistence.CrossClusterTaskTypeApplyParentClosePolicy {
		transferTasks, crossClusterTasks, err := r.generateApplyParentCloseTasks(task.TargetDomainIDs, task.GetVersion(), task.GetVisibilityTimestamp(), generateTransferTask)
		if err != nil {
			return err
		}
		r.mutableState.AddTransferTasks(transferTasks...)
		r.mutableState.AddCrossClusterTasks(crossClusterTasks...)
		return nil
	}

	if !generateTransferTask {
		targetDomainEntry, err := r.domainCache.GetDomainByID(task.TargetDomainID)
		if err != nil {
			return err
		}
		targetCluster = targetDomainEntry.GetReplicationConfig().ActiveClusterName
		if targetCluster == r.clusterMetadata.GetCurrentClusterName() {
			generateTransferTask = true
		}
	}

	var newTask persistence.Task
	switch task.GetTaskType() {
	case persistence.CrossClusterTaskTypeCancelExecution:
		cancelExecutionTask := &persistence.CancelExecutionTask{
			// TaskID is set by shard context
			TargetDomainID:          task.TargetDomainID,
			TargetWorkflowID:        task.TargetWorkflowID,
			TargetRunID:             task.TargetRunID,
			TargetChildWorkflowOnly: task.TargetChildWorkflowOnly,
			InitiatedID:             task.ScheduleID,
			Version:                 task.Version,
		}
		if generateTransferTask {
			newTask = cancelExecutionTask
		} else {
			newTask = &persistence.CrossClusterCancelExecutionTask{
				TargetCluster:       targetCluster,
				CancelExecutionTask: *cancelExecutionTask,
			}
		}
	case persistence.CrossClusterTaskTypeSignalExecution:
		signalExecutionTask := &persistence.SignalExecutionTask{
			// TaskID is set by shard context
			TargetDomainID:          task.TargetDomainID,
			TargetWorkflowID:        task.TargetWorkflowID,
			TargetRunID:             task.TargetRunID,
			TargetChildWorkflowOnly: task.TargetChildWorkflowOnly,
			InitiatedID:             task.ScheduleID,
			Version:                 task.Version,
		}
		if generateTransferTask {
			newTask = signalExecutionTask
		} else {
			newTask = &persistence.CrossClusterSignalExecutionTask{
				TargetCluster:       targetCluster,
				SignalExecutionTask: *signalExecutionTask,
			}
		}
	case persistence.CrossClusterTaskTypeStartChildExecution:
		startChildExecutionTask := &persistence.StartChildExecutionTask{
			// TaskID is set by shard context
			TargetDomainID:   task.TargetDomainID,
			TargetWorkflowID: task.TargetWorkflowID,
			InitiatedID:      task.ScheduleID,
			Version:          task.Version,
		}
		if generateTransferTask {
			newTask = startChildExecutionTask
		} else {
			newTask = &persistence.CrossClusterStartChildExecutionTask{
				TargetCluster:           targetCluster,
				StartChildExecutionTask: *startChildExecutionTask,
			}
		}
	case persistence.CrossClusterTaskTypeRecordChildExeuctionCompleted:
		recordChildExecutionCompletedTask := &persistence.RecordChildExecutionCompletedTask{
			// TaskID is set by shard context
			Version:          task.Version,
			TargetDomainID:   task.TargetDomainID,
			TargetWorkflowID: task.TargetWorkflowID,
			TargetRunID:      task.TargetRunID,
		}
		if generateTransferTask {
			newTask = recordChildExecutionCompletedTask
		} else {
			newTask = &persistence.CrossClusterRecordChildExecutionCompletedTask{
				TargetCluster:                     targetCluster,
				RecordChildExecutionCompletedTask: *recordChildExecutionCompletedTask,
			}
		}
	// persistence.CrossClusterTaskTypeApplyParentPolicy is handled by generateFromApplyParentCloseCrossClusterTask above
	default:
		return fmt.Errorf("unable to convert cross-cluster task of type %v", task.TaskType)
	}

	// set visibility timestamp here so we the metric for task latency
	// can include the latency for the original transfer task.
	newTask.SetVisibilityTimestamp(task.VisibilityTimestamp)
	if generateTransferTask {
		r.mutableState.AddTransferTasks(newTask)
	} else {
		r.mutableState.AddCrossClusterTasks(newTask)
	}

	return nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateActivityTimerTasks() error {

	_, err := NewTimerSequence(r.mutableState).CreateNextActivityTimer()
	return err
}

func (r *mutableStateTaskGeneratorImpl) GenerateUserTimerTasks() error {

	_, err := NewTimerSequence(r.mutableState).CreateNextUserTimer()
	return err
}

func (r *mutableStateTaskGeneratorImpl) getTargetDomainID(
	targetDomainName string,
) (string, error) {
	if targetDomainName != "" {
		return r.domainCache.GetDomainID(targetDomainName)
	}

	return r.mutableState.GetExecutionInfo().DomainID, nil
}

// isCrossClusterTask determines if the task belongs to the cross-cluster queue
// this is only an best effort check
// even if the task ended up in the wrong queue, the actual processing logic
// will detect it and create a new task in the right queue.
func (r *mutableStateTaskGeneratorImpl) isCrossClusterTask(
	targetDomainID string,
) (string, bool, error) {
	sourceDomainID := r.mutableState.GetExecutionInfo().DomainID

	// case 1: not cross domain task
	if sourceDomainID == targetDomainID {
		return "", false, nil
	}

	sourceDomainEntry, err := r.domainCache.GetDomainByID(sourceDomainID)
	if err != nil {
		return "", false, err
	}

	// case 2: source domain is not active in the current cluster
	if !sourceDomainEntry.IsDomainActive() {
		return "", false, nil
	}

	targetDomainEntry, err := r.domainCache.GetDomainByID(targetDomainID)
	if err != nil {
		return "", false, err
	}
	targetCluster := targetDomainEntry.GetReplicationConfig().ActiveClusterName

	// case 3: target cluster is the same as source domain active cluster
	// which is current cluster since source domain is active
	if targetCluster == r.clusterMetadata.GetCurrentClusterName() {
		return "", false, nil
	}

	return targetCluster, true, nil
}

func getTargetCluster(
	domainID string,
	domainCache cache.DomainCache,
) (string, bool, error) {
	domainEntry, err := domainCache.GetDomainByID(domainID)
	if err != nil {
		return "", false, err
	}

	isActive := domainEntry.IsDomainActive()
	if !isActive {
		// treat pending active as active
		isActive = domainEntry.IsDomainPendingActive()
	}

	activeCluster := domainEntry.GetReplicationConfig().ActiveClusterName
	return activeCluster, isActive, nil
}

func getParentCluster(
	mutableState MutableState,
	domainCache cache.DomainCache,
) (string, bool, error) {
	executionInfo := mutableState.GetExecutionInfo()
	if !mutableState.HasParentExecution() ||
		executionInfo.CloseStatus == persistence.WorkflowCloseStatusContinuedAsNew {
		// we don't need to reply to parent
		return "", false, nil
	}

	return getTargetCluster(executionInfo.ParentDomainID, domainCache)
}

func getChildrenClusters(
	childDomainIDs map[string]struct{},
	mutableState MutableState,
	domainCache cache.DomainCache,
) (map[string]struct{}, map[string]map[string]struct{}, error) {

	if len(childDomainIDs) == 0 {
		childDomainIDs = make(map[string]struct{})
		children := mutableState.GetPendingChildExecutionInfos()
		for _, childInfo := range children {
			if childInfo.ParentClosePolicy == types.ParentClosePolicyAbandon {
				continue
			}

			childDomainID, err := GetChildExecutionDomainID(childInfo, domainCache, mutableState.GetDomainEntry())
			if err != nil {
				if common.IsEntityNotExistsError(err) {
					continue // ignore deleted domain
				}
				return nil, nil, err
			}
			childDomainIDs[childDomainID] = struct{}{}
		}
	}

	sameClusterDomainIDs := make(map[string]struct{})
	remoteClusterDomainIDs := make(map[string]map[string]struct{})
	for childDomainID := range childDomainIDs {
		childCluster, isActive, err := getTargetCluster(childDomainID, domainCache)
		if err != nil {
			return nil, nil, err
		}

		if isActive {
			sameClusterDomainIDs[childDomainID] = struct{}{}
		} else {
			if _, ok := remoteClusterDomainIDs[childCluster]; !ok {
				remoteClusterDomainIDs[childCluster] = make(map[string]struct{})
			}
			remoteClusterDomainIDs[childCluster][childDomainID] = struct{}{}
		}
	}

	return sameClusterDomainIDs, remoteClusterDomainIDs, nil
}

func getNextDecisionTimeout(attempt int64, defaultStartToCloseTimeout time.Duration) time.Duration {
	if attempt <= 1 {
		return defaultStartToCloseTimeout
	}

	nextInterval := float64(defaultInitIntervalForDecisionRetry) * math.Pow(2, float64(attempt-2))
	nextInterval = math.Min(nextInterval, float64(defaultMaxIntervalForDecisionRetry))
	jitterPortion := int(defaultJitterCoefficient * nextInterval)
	if jitterPortion < 1 {
		jitterPortion = 1
	}
	nextInterval = nextInterval*(1-defaultJitterCoefficient) + float64(rand.Intn(jitterPortion))
	return time.Duration(nextInterval)
}
