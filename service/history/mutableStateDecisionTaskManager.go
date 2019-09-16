package history

import (
	"fmt"
	"math"
	"time"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/errors"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
)

type (
	MutableStateDecisionTaskManager interface {
		ReplicateDecisionTaskScheduledEvent(
			version int64,
			scheduleID int64,
			taskList string,
			startToCloseTimeoutSeconds int32,
			attempt int64,
			scheduleTimestamp int64,
			originalScheduledTimestamp int64,
		) (*decisionInfo, error)
		ReplicateTransientDecisionTaskScheduled() (*decisionInfo, error)
		ReplicateDecisionTaskFailedEvent() error
		ReplicateDecisionTaskTimedOutEvent(timeoutType workflow.TimeoutType) error
		ReplicateDecisionTaskCompletedEvent(event *workflow.HistoryEvent) error
		ReplicateDecisionTaskStartedEvent(
			decision *decisionInfo,
			version int64,
			scheduleID int64,
			startedID int64,
			requestID string,
			timestamp int64,
		) (*decisionInfo, error)

		AddDecisionTaskFailedEvent(
			scheduleEventID int64,
			startedEventID int64,
			cause workflow.DecisionTaskFailedCause,
			details []byte,
			identity string,
			reason string,
			baseRunID string,
			newRunID string,
			forkEventVersion int64,
		) (*workflow.HistoryEvent, error)
		AddDecisionTaskScheduleToStartTimeoutEvent(scheduleEventID int64) (*workflow.HistoryEvent, error)
		AddDecisionTaskTimedOutEvent(scheduleEventID int64, startedEventID int64) (*workflow.HistoryEvent, error)
		AddDecisionTaskCompletedEvent(
			scheduleEventID int64,
			startedEventID int64,
			request *workflow.RespondDecisionTaskCompletedRequest,
			maxResetPoints int,
		) (*workflow.HistoryEvent, error)
		AddDecisionTaskStartedEvent(
			scheduleEventID int64,
			requestID string,
			request *workflow.PollForDecisionTaskRequest,
		) (*workflow.HistoryEvent, *decisionInfo, error)
		AddDecisionTaskScheduledEventAsHeartbeat(
			bypassTaskGeneration bool,
			originalScheduledTimestamp int64,
		) (*decisionInfo, error)
		AddDecisionTaskScheduledEvent(bypassTaskGeneration bool) (*decisionInfo, error)
		AddFirstDecisionTaskScheduled(startEvent *workflow.HistoryEvent) error
		FailDecision(incrementAttempt bool)
		DeleteDecision()
		UpdateDecision(decision *decisionInfo)

		HasPendingDecision() bool
		GetPendingDecision() (*decisionInfo, bool)
		HasInFlightDecision() bool
		GetInFlightDecision() (*decisionInfo, bool)
		HasProcessedOrPendingDecision() bool
		GetDecisionInfo(scheduleEventID int64) (*decisionInfo, bool)

		CreateTransientDecisionEvents(decision *decisionInfo, identity string) (*workflow.HistoryEvent, *workflow.HistoryEvent)
	}

	mutableStateDecisionTaskManager struct {
		*mutableStateBuilder
	}
)

func NewMutableStateDecisionTaskManager(msb *mutableStateBuilder) MutableStateDecisionTaskManager {
	return &mutableStateDecisionTaskManager{
		mutableStateBuilder: msb,
	}
}

func (e *mutableStateBuilder) ReplicateDecisionTaskScheduledEvent(
	version int64,
	scheduleID int64,
	taskList string,
	startToCloseTimeoutSeconds int32,
	attempt int64,
	scheduleTimestamp int64,
	originalScheduledTimestamp int64,
) (*decisionInfo, error) {
	decision := &decisionInfo{
		Version:                    version,
		ScheduleID:                 scheduleID,
		StartedID:                  common.EmptyEventID,
		RequestID:                  emptyUUID,
		DecisionTimeout:            startToCloseTimeoutSeconds,
		TaskList:                   taskList,
		Attempt:                    attempt,
		ScheduledTimestamp:         scheduleTimestamp,
		StartedTimestamp:           0,
		OriginalScheduledTimestamp: originalScheduledTimestamp,
	}

	e.UpdateDecision(decision)
	return decision, nil
}

func (e *mutableStateBuilder) ReplicateTransientDecisionTaskScheduled() (*decisionInfo, error) {
	if e.HasPendingDecision() || e.GetExecutionInfo().DecisionAttempt == 0 {
		return nil, nil
	}

	// the schedule ID for this decision is guaranteed to be wrong
	// since the next event ID is assigned at the very end of when
	// all events are applied for replication.
	// this is OK
	// 1. if a failover happen just after this transient decision,
	// AddDecisionTaskStartedEvent will handle the correction of schedule ID
	// and set the attempt to 0
	// 2. if no failover happen during the life time of this transient decision
	// then ReplicateDecisionTaskScheduledEvent will overwrite everything
	// including the decision schedule ID
	decision := &decisionInfo{
		Version:            e.GetCurrentVersion(),
		ScheduleID:         e.GetNextEventID(),
		StartedID:          common.EmptyEventID,
		RequestID:          emptyUUID,
		DecisionTimeout:    e.GetExecutionInfo().DecisionTimeoutValue,
		TaskList:           e.GetExecutionInfo().TaskList,
		Attempt:            e.GetExecutionInfo().DecisionAttempt,
		ScheduledTimestamp: e.timeSource.Now().UnixNano(),
		StartedTimestamp:   0,
	}

	e.UpdateDecision(decision)
	return decision, nil
}

func (e *mutableStateBuilder) ReplicateDecisionTaskFailedEvent() error {
	e.FailDecision(true)
	return nil
}

func (e *mutableStateBuilder) ReplicateDecisionTaskTimedOutEvent(timeoutType workflow.TimeoutType) error {

	incrementAttempt := true
	// Do not increment decision attempt in the case of sticky timeout to prevent creating next decision as transient
	if timeoutType == workflow.TimeoutTypeScheduleToStart {
		incrementAttempt = false
	}
	e.FailDecision(incrementAttempt)
	return nil
}

func (e *mutableStateBuilder) ReplicateDecisionTaskCompletedEvent(event *workflow.HistoryEvent) error {

	e.beforeAddDecisionTaskCompletedEvent()
	e.afterAddDecisionTaskCompletedEvent(event, math.MaxInt32)
	return nil
}

func (e *mutableStateBuilder) ReplicateDecisionTaskStartedEvent(
	decision *decisionInfo,
	version int64,
	scheduleID int64,
	startedID int64,
	requestID string,
	timestamp int64,
) (*decisionInfo, error) {
	// Replicator calls it with a nil decision info, and it is safe to always lookup the decision in this case as it
	// does not have to deal with transient decision case.
	var ok bool
	if decision == nil {
		decision, ok = e.GetDecisionInfo(scheduleID)
		if !ok {
			return nil, errors.NewInternalFailureError(fmt.Sprintf("unable to find decision: %v", scheduleID))
		}
		// setting decision attempt to 0 for decision task replication
		// this mainly handles transient decision completion
		// for transient decision, active side will write 2 batch in a "transaction"
		// 1. decision task scheduled & decision task started
		// 2. decision task completed & other events
		// since we need to treat each individual event batch as one transaction
		// certain "magic" needs to be done, i.e. setting attempt to 0 so
		// if first batch is replicated, but not the second one, decision can be correctly timed out
		decision.Attempt = 0
	}

	e.executionInfo.State = persistence.WorkflowStateRunning
	// Update mutable decision state
	decision = &decisionInfo{
		Version:                    version,
		ScheduleID:                 scheduleID,
		StartedID:                  startedID,
		RequestID:                  requestID,
		DecisionTimeout:            decision.DecisionTimeout,
		Attempt:                    decision.Attempt,
		StartedTimestamp:           timestamp,
		ScheduledTimestamp:         decision.ScheduledTimestamp,
		TaskList:                   decision.TaskList,
		OriginalScheduledTimestamp: decision.OriginalScheduledTimestamp,
	}

	e.UpdateDecision(decision)
	return decision, nil
}

func (e *mutableStateBuilder) AddDecisionTaskFailedEvent(
	scheduleEventID int64,
	startedEventID int64,
	cause workflow.DecisionTaskFailedCause,
	details []byte,
	identity string,
	reason string,
	baseRunID string,
	newRunID string,
	forkEventVersion int64,
) (*workflow.HistoryEvent, error) {

	opTag := tag.WorkflowActionDecisionTaskFailed
	if err := e.checkMutability(opTag); err != nil {
		return nil, err
	}

	attr := workflow.DecisionTaskFailedEventAttributes{
		ScheduledEventId: common.Int64Ptr(scheduleEventID),
		StartedEventId:   common.Int64Ptr(startedEventID),
		Cause:            common.DecisionTaskFailedCausePtr(cause),
		Details:          details,
		Identity:         common.StringPtr(identity),
		Reason:           common.StringPtr(reason),
		BaseRunId:        common.StringPtr(baseRunID),
		NewRunId:         common.StringPtr(newRunID),
		ForkEventVersion: common.Int64Ptr(forkEventVersion),
	}

	dt, ok := e.GetDecisionInfo(scheduleEventID)
	if !ok || dt.StartedID != startedEventID {
		e.logger.Warn(mutableStateInvalidHistoryActionMsg, opTag,
			tag.WorkflowEventID(e.GetNextEventID()),
			tag.ErrorTypeInvalidHistoryAction,
			tag.WorkflowScheduleID(scheduleEventID),
			tag.WorkflowStartedID(startedEventID))
		return nil, e.createInternalServerError(opTag)
	}

	var event *workflow.HistoryEvent
	// Only emit DecisionTaskFailedEvent for the very first time
	if dt.Attempt == 0 || cause == workflow.DecisionTaskFailedCauseResetWorkflow {
		event = e.hBuilder.AddDecisionTaskFailedEvent(attr)
	}

	if err := e.ReplicateDecisionTaskFailedEvent(); err != nil {
		return nil, err
	}

	// always clear decision attempt for reset
	if cause == workflow.DecisionTaskFailedCauseResetWorkflow {
		e.executionInfo.DecisionAttempt = 0
	}
	return event, nil
}

func (e *mutableStateBuilder) AddDecisionTaskScheduleToStartTimeoutEvent(scheduleEventID int64) (*workflow.HistoryEvent, error) {

	opTag := tag.WorkflowActionDecisionTaskTimedOut
	if err := e.checkMutability(opTag); err != nil {
		return nil, err
	}

	if e.executionInfo.DecisionScheduleID != scheduleEventID || e.executionInfo.DecisionStartedID > 0 {
		e.logger.Warn(mutableStateInvalidHistoryActionMsg, opTag,
			tag.WorkflowEventID(e.GetNextEventID()),
			tag.ErrorTypeInvalidHistoryAction,
			tag.WorkflowScheduleID(scheduleEventID),
		)
		return nil, e.createInternalServerError(opTag)
	}

	// Clear stickiness whenever decision fails
	e.ClearStickyness()

	event := e.hBuilder.AddDecisionTaskTimedOutEvent(scheduleEventID, 0, workflow.TimeoutTypeScheduleToStart)

	if err := e.ReplicateDecisionTaskTimedOutEvent(workflow.TimeoutTypeScheduleToStart); err != nil {
		return nil, err
	}
	return event, nil
}

func (e *mutableStateBuilder) AddDecisionTaskTimedOutEvent(scheduleEventID int64, startedEventID int64) (*workflow.HistoryEvent, error) {

	opTag := tag.WorkflowActionDecisionTaskTimedOut
	if err := e.checkMutability(opTag); err != nil {
		return nil, err
	}

	dt, ok := e.GetDecisionInfo(scheduleEventID)
	if !ok || dt.StartedID != startedEventID {
		e.logger.Warn(mutableStateInvalidHistoryActionMsg, opTag,
			tag.WorkflowEventID(e.GetNextEventID()),
			tag.ErrorTypeInvalidHistoryAction,
			tag.WorkflowScheduleID(scheduleEventID),
			tag.WorkflowStartedID(startedEventID))
		return nil, e.createInternalServerError(opTag)
	}

	var event *workflow.HistoryEvent
	// Avoid creating new history events when decisions are continuously timing out
	if dt.Attempt == 0 {
		event = e.hBuilder.AddDecisionTaskTimedOutEvent(scheduleEventID, startedEventID, workflow.TimeoutTypeStartToClose)
	}

	if err := e.ReplicateDecisionTaskTimedOutEvent(workflow.TimeoutTypeStartToClose); err != nil {
		return nil, err
	}
	return event, nil
}

func (e *mutableStateBuilder) AddDecisionTaskCompletedEvent(
	scheduleEventID int64,
	startedEventID int64,
	request *workflow.RespondDecisionTaskCompletedRequest,
	maxResetPoints int,
) (*workflow.HistoryEvent, error) {

	opTag := tag.WorkflowActionDecisionTaskCompleted
	if err := e.checkMutability(opTag); err != nil {
		return nil, err
	}

	decision, ok := e.GetDecisionInfo(scheduleEventID)
	if !ok || decision.StartedID != startedEventID {
		e.logger.Warn(mutableStateInvalidHistoryActionMsg, opTag,
			tag.WorkflowEventID(e.GetNextEventID()),
			tag.ErrorTypeInvalidHistoryAction,
			tag.WorkflowScheduleID(scheduleEventID),
			tag.WorkflowStartedID(startedEventID))

		return nil, e.createInternalServerError(opTag)
	}

	e.beforeAddDecisionTaskCompletedEvent()
	if decision.Attempt > 0 {
		// Create corresponding DecisionTaskSchedule and DecisionTaskStarted events for decisions we have been retrying
		scheduledEvent := e.hBuilder.AddTransientDecisionTaskScheduledEvent(e.executionInfo.TaskList, decision.DecisionTimeout,
			decision.Attempt, decision.ScheduledTimestamp)
		startedEvent := e.hBuilder.AddTransientDecisionTaskStartedEvent(scheduledEvent.GetEventId(), decision.RequestID,
			request.GetIdentity(), decision.StartedTimestamp)
		startedEventID = startedEvent.GetEventId()
	}
	// Now write the completed event
	event := e.hBuilder.AddDecisionTaskCompletedEvent(scheduleEventID, startedEventID, request)

	e.afterAddDecisionTaskCompletedEvent(event, maxResetPoints)
	return event, nil
}

func (e *mutableStateBuilder) AddDecisionTaskStartedEvent(
	scheduleEventID int64,
	requestID string,
	request *workflow.PollForDecisionTaskRequest,
) (*workflow.HistoryEvent, *decisionInfo, error) {

	opTag := tag.WorkflowActionDecisionTaskStarted
	if err := e.checkMutability(opTag); err != nil {
		return nil, nil, err
	}

	decision, ok := e.GetDecisionInfo(scheduleEventID)
	if !ok || decision.StartedID != common.EmptyEventID {
		e.logger.Warn(mutableStateInvalidHistoryActionMsg, opTag,
			tag.WorkflowEventID(e.GetNextEventID()),
			tag.ErrorTypeInvalidHistoryAction,
			tag.WorkflowScheduleID(scheduleEventID))
		return nil, nil, e.createInternalServerError(opTag)
	}

	var event *workflow.HistoryEvent
	scheduleID := decision.ScheduleID
	startedID := scheduleID + 1
	tasklist := request.TaskList.GetName()
	startTime := e.timeSource.Now().UnixNano()
	// First check to see if new events came since transient decision was scheduled
	if decision.Attempt > 0 && decision.ScheduleID != e.GetNextEventID() {
		// Also create a new DecisionTaskScheduledEvent since new events came in when it was scheduled
		scheduleEvent := e.hBuilder.AddDecisionTaskScheduledEvent(tasklist, decision.DecisionTimeout, 0)
		scheduleID = scheduleEvent.GetEventId()
		decision.Attempt = 0
	}

	// Avoid creating new history events when decisions are continuously failing
	if decision.Attempt == 0 {
		// Now create DecisionTaskStartedEvent
		event = e.hBuilder.AddDecisionTaskStartedEvent(scheduleID, requestID, request.GetIdentity())
		startedID = event.GetEventId()
		startTime = event.GetTimestamp()
	}

	decision, err := e.ReplicateDecisionTaskStartedEvent(decision, e.GetCurrentVersion(), scheduleID, startedID, requestID, startTime)
	// TODO merge active & passive task generation
	if err := e.taskGenerator.generateDecisionStartTasks(
		e.unixNanoToTime(startTime), // start time is now
		scheduleID,
	); err != nil {
		return nil, nil, err
	}
	return event, decision, err
}

// originalScheduledTimestamp is to record the first scheduled decision during decision heartbeat.
func (e *mutableStateBuilder) AddDecisionTaskScheduledEventAsHeartbeat(
	bypassTaskGeneration bool,
	originalScheduledTimestamp int64,
) (*decisionInfo, error) {

	opTag := tag.WorkflowActionDecisionTaskScheduled
	if err := e.checkMutability(opTag); err != nil {
		return nil, err
	}

	if e.HasPendingDecision() {
		e.logger.Warn(mutableStateInvalidHistoryActionMsg, opTag,
			tag.WorkflowEventID(e.GetNextEventID()),
			tag.ErrorTypeInvalidHistoryAction,
			tag.WorkflowScheduleID(e.executionInfo.DecisionScheduleID))
		return nil, e.createInternalServerError(opTag)
	}

	// set workflow state to running
	// since decision is scheduled
	e.executionInfo.State = persistence.WorkflowStateRunning

	// Tasklist and decision timeout should already be set from workflow execution started event
	taskList := e.executionInfo.TaskList
	if e.IsStickyTaskListEnabled() {
		taskList = e.executionInfo.StickyTaskList
	} else {
		// It can be because stickyness has expired due to StickyTTL config
		// In that case we need to clear stickyness so that the LastUpdateTimestamp is not corrupted.
		// In other cases, clearing stickyness shouldn't hurt anything.
		// TODO: https://github.com/uber/cadence/issues/2357:
		//  if we can use a new field(LastDecisionUpdateTimestamp), then we could get rid of it.
		e.ClearStickyness()
	}
	startToCloseTimeoutSeconds := e.executionInfo.DecisionTimeoutValue

	// Flush any buffered events before creating the decision, otherwise it will result in invalid IDs for transient
	// decision and will cause in timeout processing to not work for transient decisions
	if e.HasBufferedEvents() {
		// if creating a decision and in the mean time events are flushed from buffered events
		// than this decision cannot be a transient decision
		e.executionInfo.DecisionAttempt = 0
		if err := e.FlushBufferedEvents(); err != nil {
			return nil, err
		}
	}

	var newDecisionEvent *workflow.HistoryEvent
	scheduleID := e.GetNextEventID() // we will generate the schedule event later for repeatedly failing decisions
	// Avoid creating new history events when decisions are continuously failing
	scheduleTime := e.timeSource.Now().UnixNano()
	if e.executionInfo.DecisionAttempt == 0 {
		newDecisionEvent = e.hBuilder.AddDecisionTaskScheduledEvent(taskList, startToCloseTimeoutSeconds,
			e.executionInfo.DecisionAttempt)
		scheduleID = newDecisionEvent.GetEventId()
		scheduleTime = newDecisionEvent.GetTimestamp()
	}

	decision, err := e.ReplicateDecisionTaskScheduledEvent(
		e.GetCurrentVersion(),
		scheduleID,
		taskList,
		startToCloseTimeoutSeconds,
		e.executionInfo.DecisionAttempt,
		scheduleTime,
		originalScheduledTimestamp,
	)
	if err != nil {
		return nil, err
	}

	// TODO merge active & passive task generation
	if !bypassTaskGeneration {
		if err := e.taskGenerator.generateDecisionScheduleTasks(
			e.unixNanoToTime(scheduleTime), // schedule time is now
			scheduleID,
		); err != nil {
			return nil, err
		}
	}

	return decision, nil
}

func (e *mutableStateBuilder) AddDecisionTaskScheduledEvent(bypassTaskGeneration bool) (*decisionInfo, error) {
	return e.AddDecisionTaskScheduledEventAsHeartbeat(bypassTaskGeneration, e.timeSource.Now().UnixNano())
}

func (e *mutableStateBuilder) AddFirstDecisionTaskScheduled(startEvent *workflow.HistoryEvent) error {

	// handle first decision case, i.e. possible delayed decision
	//
	// below handles the following cases:
	// 1. if not continue as new & if workflow has no parent
	//   -> schedule decision & schedule delayed decision
	// 2. if not continue as new & if workflow has parent
	//   -> this function should not be called during workflow start, but should be called as
	//      part of schedule decision in 2 phase commit
	//
	// if continue as new
	//  1. whether has parent workflow or not
	//   -> schedule decision & schedule delayed decision
	//
	startAttr := startEvent.WorkflowExecutionStartedEventAttributes
	decisionBackoffDuration := time.Duration(startAttr.GetFirstDecisionTaskBackoffSeconds()) * time.Second

	var err error
	if decisionBackoffDuration != 0 {
		if err = e.taskGenerator.generateDelayedDecisionTasks(
			e.unixNanoToTime(startEvent.GetTimestamp()),
			startEvent,
		); err != nil {
			return err
		}
	} else {
		if _, err = e.AddDecisionTaskScheduledEvent(
			false,
		); err != nil {
			return err
		}
	}

	return nil
}

func (e *mutableStateBuilder) FailDecision(incrementAttempt bool) {
	// Clear stickiness whenever decision fails
	e.ClearStickyness()

	failDecisionInfo := &decisionInfo{
		Version:                    common.EmptyVersion,
		ScheduleID:                 common.EmptyEventID,
		StartedID:                  common.EmptyEventID,
		RequestID:                  emptyUUID,
		DecisionTimeout:            0,
		StartedTimestamp:           0,
		TaskList:                   "",
		OriginalScheduledTimestamp: 0,
	}
	if incrementAttempt {
		failDecisionInfo.Attempt = e.executionInfo.DecisionAttempt + 1
		failDecisionInfo.ScheduledTimestamp = e.timeSource.Now().UnixNano()
	}
	e.UpdateDecision(failDecisionInfo)
}

// DeleteDecision deletes a decision task.
func (e *mutableStateBuilder) DeleteDecision() {
	resetDecisionInfo := &decisionInfo{
		Version:            common.EmptyVersion,
		ScheduleID:         common.EmptyEventID,
		StartedID:          common.EmptyEventID,
		RequestID:          emptyUUID,
		DecisionTimeout:    0,
		Attempt:            0,
		StartedTimestamp:   0,
		ScheduledTimestamp: 0,
		TaskList:           "",
		// Keep the last original scheduled timestamp, so that AddDecisionAsHeartbeat can continue with it.
		OriginalScheduledTimestamp: e.getDecisionInfo().OriginalScheduledTimestamp,
	}
	e.UpdateDecision(resetDecisionInfo)
}

// UpdateDecision updates a decision task.
func (e *mutableStateBuilder) UpdateDecision(decision *decisionInfo) {

	e.executionInfo.DecisionVersion = decision.Version
	e.executionInfo.DecisionScheduleID = decision.ScheduleID
	e.executionInfo.DecisionStartedID = decision.StartedID
	e.executionInfo.DecisionRequestID = decision.RequestID
	e.executionInfo.DecisionTimeout = decision.DecisionTimeout
	e.executionInfo.DecisionAttempt = decision.Attempt
	e.executionInfo.DecisionStartedTimestamp = decision.StartedTimestamp
	e.executionInfo.DecisionScheduledTimestamp = decision.ScheduledTimestamp
	e.executionInfo.DecisionOriginalScheduledTimestamp = decision.OriginalScheduledTimestamp

	// NOTE: do not update tasklist in execution info

	e.logger.Debug(fmt.Sprintf(
		"Decision Updated: {Schedule: %v, Started: %v, ID: %v, Timeout: %v, Attempt: %v, Timestamp: %v}",
		decision.ScheduleID,
		decision.StartedID,
		decision.RequestID,
		decision.DecisionTimeout,
		decision.Attempt,
		decision.StartedTimestamp,
	))
}

func (e *mutableStateBuilder) HasPendingDecision() bool {
	return e.executionInfo.DecisionScheduleID != common.EmptyEventID
}

func (e *mutableStateBuilder) GetPendingDecision() (*decisionInfo, bool) {
	if e.executionInfo.DecisionScheduleID == common.EmptyEventID {
		return nil, false
	}

	decision := e.getDecisionInfo()
	return decision, true
}

func (e *mutableStateBuilder) HasInFlightDecision() bool {
	return e.executionInfo.DecisionStartedID > 0
}

func (e *mutableStateBuilder) GetInFlightDecision() (*decisionInfo, bool) {
	if e.executionInfo.DecisionScheduleID == common.EmptyEventID ||
		e.executionInfo.DecisionStartedID == common.EmptyEventID {
		return nil, false
	}

	decision := e.getDecisionInfo()
	return decision, true
}

func (e *mutableStateBuilder) HasProcessedOrPendingDecision() bool {
	return e.HasPendingDecision() || e.GetPreviousStartedEventID() != common.EmptyEventID
}

// GetDecisionInfo returns details about the in-progress decision task
func (e *mutableStateBuilder) GetDecisionInfo(scheduleEventID int64) (*decisionInfo, bool) {
	decision := e.getDecisionInfo()
	if scheduleEventID == decision.ScheduleID {
		return decision, true
	}
	return nil, false
}

func (e *mutableStateBuilder) CreateTransientDecisionEvents(decision *decisionInfo, identity string) (*workflow.HistoryEvent, *workflow.HistoryEvent) {
	tasklist := e.executionInfo.TaskList
	scheduledEvent := newDecisionTaskScheduledEventWithInfo(
		decision.ScheduleID,
		decision.ScheduledTimestamp,
		tasklist,
		decision.DecisionTimeout,
		decision.Attempt,
	)

	startedEvent := newDecisionTaskStartedEventWithInfo(
		decision.StartedID,
		decision.StartedTimestamp,
		decision.ScheduleID,
		decision.RequestID,
		identity,
	)

	return scheduledEvent, startedEvent
}


func (e *mutableStateBuilder) getDecisionInfo() *decisionInfo {
	taskList := e.executionInfo.TaskList
	if e.IsStickyTaskListEnabled() {
		taskList = e.executionInfo.StickyTaskList
	}
	return &decisionInfo{
		Version:                    e.executionInfo.DecisionVersion,
		ScheduleID:                 e.executionInfo.DecisionScheduleID,
		StartedID:                  e.executionInfo.DecisionStartedID,
		RequestID:                  e.executionInfo.DecisionRequestID,
		DecisionTimeout:            e.executionInfo.DecisionTimeout,
		Attempt:                    e.executionInfo.DecisionAttempt,
		StartedTimestamp:           e.executionInfo.DecisionStartedTimestamp,
		ScheduledTimestamp:         e.executionInfo.DecisionScheduledTimestamp,
		TaskList:                   taskList,
		OriginalScheduledTimestamp: e.executionInfo.DecisionOriginalScheduledTimestamp,
	}
}

func (e *mutableStateBuilder) beforeAddDecisionTaskCompletedEvent() {
	// Make sure to delete decision before adding events.  Otherwise they are buffered rather than getting appended
	e.DeleteDecision()
}

func (e *mutableStateBuilder) afterAddDecisionTaskCompletedEvent(
	event *workflow.HistoryEvent,
	maxResetPoints int,
) {
	e.executionInfo.LastProcessedEvent = event.GetDecisionTaskCompletedEventAttributes().GetStartedEventId()
	e.addBinaryCheckSumIfNotExists(event, maxResetPoints)
}


