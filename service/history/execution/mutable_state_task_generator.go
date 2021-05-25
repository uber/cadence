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
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

type (
	// MutableStateTaskGenerator generates workflow transfer and timer tasks
	MutableStateTaskGenerator interface {
		GenerateWorkflowStartTasks(
			now time.Time,
			startEvent *types.HistoryEvent,
		) error
		GenerateWorkflowCloseTasks(
			now time.Time,
		) error
		GenerateRecordWorkflowStartedTasks(
			startEvent *types.HistoryEvent,
		) error
		GenerateDelayedDecisionTasks(
			now time.Time,
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

		// these 2 APIs should only be called when mutable state transaction is being closed
		GenerateActivityTimerTasks(
			now time.Time,
		) error
		GenerateUserTimerTasks(
			now time.Time,
		) error
	}

	mutableStateTaskGeneratorImpl struct {
		domainCache cache.DomainCache
		logger      log.Logger

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
	domainCache cache.DomainCache,
	logger log.Logger,
	mutableState MutableState,
) MutableStateTaskGenerator {

	return &mutableStateTaskGeneratorImpl{
		domainCache: domainCache,
		logger:      logger,

		mutableState: mutableState,
	}
}

func (r *mutableStateTaskGeneratorImpl) GenerateWorkflowStartTasks(
	now time.Time,
	startEvent *types.HistoryEvent,
) error {

	attr := startEvent.WorkflowExecutionStartedEventAttributes
	firstDecisionDelayDuration := time.Duration(attr.GetFirstDecisionTaskBackoffSeconds()) * time.Second

	executionInfo := r.mutableState.GetExecutionInfo()
	startVersion := startEvent.GetVersion()

	workflowTimeoutDuration := time.Duration(executionInfo.WorkflowTimeout) * time.Second
	workflowTimeoutDuration = workflowTimeoutDuration + firstDecisionDelayDuration
	workflowTimeoutTimestamp := now.Add(workflowTimeoutDuration)
	if !executionInfo.ExpirationTime.IsZero() && workflowTimeoutTimestamp.After(executionInfo.ExpirationTime) {
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
	now time.Time,
) error {

	currentVersion := r.mutableState.GetCurrentVersion()
	executionInfo := r.mutableState.GetExecutionInfo()

	r.mutableState.AddTransferTasks(&persistence.CloseExecutionTask{
		// TaskID and VisibilityTimestamp are set by shard context
		Version: currentVersion,
	})

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

	retentionDuration := time.Duration(retentionInDays) * time.Hour * 24
	r.mutableState.AddTimerTasks(&persistence.DeleteHistoryEventTask{
		// TaskID is set by shard
		VisibilityTimestamp: now.Add(retentionDuration),
		Version:             currentVersion,
	})

	return nil
}

func (r *mutableStateTaskGeneratorImpl) GenerateDelayedDecisionTasks(
	now time.Time,
	startEvent *types.HistoryEvent,
) error {

	startVersion := startEvent.GetVersion()

	startAttr := startEvent.WorkflowExecutionStartedEventAttributes
	decisionBackoffDuration := time.Duration(startAttr.GetFirstDecisionTaskBackoffSeconds()) * time.Second
	executionTimestamp := now.Add(decisionBackoffDuration)

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

	if r.mutableState.IsStickyTaskListEnabled() {
		scheduledTime := time.Unix(0, decision.ScheduledTimestamp)
		scheduleToStartTimeout := time.Duration(
			r.mutableState.GetExecutionInfo().StickyScheduleToStartTimeout,
		) * time.Second

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

	targetDomainID, err := r.getTargetDomainID(childWorkflowTargetDomain)
	if err != nil {
		return err
	}

	r.mutableState.AddTransferTasks(&persistence.StartChildExecutionTask{
		// TaskID and VisibilityTimestamp are set by shard context
		TargetDomainID:   targetDomainID,
		TargetWorkflowID: childWorkflowInfo.StartedWorkflowID,
		InitiatedID:      childWorkflowInfo.InitiatedID,
		Version:          childWorkflowInfo.Version,
	})

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

	r.mutableState.AddTransferTasks(&persistence.CancelExecutionTask{
		// TaskID and VisibilityTimestamp are set by shard context
		TargetDomainID:          targetDomainID,
		TargetWorkflowID:        targetWorkflowID,
		TargetRunID:             targetRunID,
		TargetChildWorkflowOnly: targetChildOnly,
		InitiatedID:             scheduleID,
		Version:                 version,
	})

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

	r.mutableState.AddTransferTasks(&persistence.SignalExecutionTask{
		// TaskID and VisibilityTimestamp are set by shard context
		TargetDomainID:          targetDomainID,
		TargetWorkflowID:        targetWorkflowID,
		TargetRunID:             targetRunID,
		TargetChildWorkflowOnly: targetChildOnly,
		InitiatedID:             scheduleID,
		Version:                 version,
	})

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

func (r *mutableStateTaskGeneratorImpl) GenerateActivityTimerTasks(
	now time.Time,
) error {

	_, err := r.getTimerSequence(now).CreateNextActivityTimer()
	return err
}

func (r *mutableStateTaskGeneratorImpl) GenerateUserTimerTasks(
	now time.Time,
) error {

	_, err := r.getTimerSequence(now).CreateNextUserTimer()
	return err
}

func (r *mutableStateTaskGeneratorImpl) getTimerSequence(now time.Time) TimerSequence {
	timeSource := clock.NewEventTimeSource()
	timeSource.Update(now)
	return NewTimerSequence(timeSource, r.mutableState)
}

func (r *mutableStateTaskGeneratorImpl) getTargetDomainID(
	targetDomainName string,
) (string, error) {

	targetDomainID := r.mutableState.GetExecutionInfo().DomainID
	if targetDomainName != "" {
		return r.domainCache.GetDomainID(targetDomainName)
	}

	return targetDomainID, nil
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
