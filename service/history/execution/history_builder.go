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

package execution

import (
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

type (
	// HistoryBuilder builds and stores workflow history events
	HistoryBuilder struct {
		transientHistory []*types.HistoryEvent
		history          []*types.HistoryEvent
		msBuilder        MutableState
	}
)

// NewHistoryBuilder creates a new history builder
func NewHistoryBuilder(msBuilder MutableState, logger log.Logger) *HistoryBuilder {
	return &HistoryBuilder{
		transientHistory: []*types.HistoryEvent{},
		history:          []*types.HistoryEvent{},
		msBuilder:        msBuilder,
	}
}

// NewHistoryBuilderFromEvents creates a new history builder based on the given workflow history events
func NewHistoryBuilderFromEvents(history []*types.HistoryEvent, logger log.Logger) *HistoryBuilder {
	return &HistoryBuilder{
		history: history,
	}
}

// GetFirstEvent gets the first event in workflow history
// it returns the first transient history event if exists
func (b *HistoryBuilder) GetFirstEvent() *types.HistoryEvent {
	// Transient decision events are always written before other events
	if b.transientHistory != nil && len(b.transientHistory) > 0 {
		return b.transientHistory[0]
	}

	if b.history != nil && len(b.history) > 0 {
		return b.history[0]
	}

	return nil
}

// HasTransientEvents returns true if there are transient history events
func (b *HistoryBuilder) HasTransientEvents() bool {
	return b.transientHistory != nil && len(b.transientHistory) > 0
}

// AddWorkflowExecutionStartedEvent adds WorkflowExecutionStarted event to history
// originalRunID is the runID when the WorkflowExecutionStarted event is written
// firstRunID is the very first runID along the chain of ContinueAsNew and Reset
func (b *HistoryBuilder) AddWorkflowExecutionStartedEvent(request *types.HistoryStartWorkflowExecutionRequest,
	previousExecution *persistence.WorkflowExecutionInfo, firstRunID, originalRunID string) *types.HistoryEvent {
	event := b.newWorkflowExecutionStartedEvent(request, previousExecution, firstRunID, originalRunID)

	return b.addEventToHistory(event)
}

// AddDecisionTaskScheduledEvent adds DecisionTaskScheduled event to history
func (b *HistoryBuilder) AddDecisionTaskScheduledEvent(taskList string,
	startToCloseTimeoutSeconds int32, attempt int64) *types.HistoryEvent {
	event := b.newDecisionTaskScheduledEvent(taskList, startToCloseTimeoutSeconds, attempt)

	return b.addEventToHistory(event)
}

// AddTransientDecisionTaskScheduledEvent adds transient DecisionTaskScheduled event
func (b *HistoryBuilder) AddTransientDecisionTaskScheduledEvent(taskList string,
	startToCloseTimeoutSeconds int32, attempt int64, timestamp int64) *types.HistoryEvent {
	event := b.newTransientDecisionTaskScheduledEvent(taskList, startToCloseTimeoutSeconds, attempt, timestamp)

	return b.addTransientEvent(event)
}

// AddDecisionTaskStartedEvent adds DecisionTaskStarted event to history
func (b *HistoryBuilder) AddDecisionTaskStartedEvent(scheduleEventID int64, requestID string,
	identity string) *types.HistoryEvent {
	event := b.newDecisionTaskStartedEvent(scheduleEventID, requestID, identity)

	return b.addEventToHistory(event)
}

// AddTransientDecisionTaskStartedEvent adds transient DecisionTaskStarted event
func (b *HistoryBuilder) AddTransientDecisionTaskStartedEvent(scheduleEventID int64, requestID string,
	identity string, timestamp int64) *types.HistoryEvent {
	event := b.newTransientDecisionTaskStartedEvent(scheduleEventID, requestID, identity, timestamp)

	return b.addTransientEvent(event)
}

// AddDecisionTaskCompletedEvent adds DecisionTaskCompleted event to history
func (b *HistoryBuilder) AddDecisionTaskCompletedEvent(scheduleEventID, StartedEventID int64,
	request *types.RespondDecisionTaskCompletedRequest) *types.HistoryEvent {
	event := b.newDecisionTaskCompletedEvent(scheduleEventID, StartedEventID, request)

	return b.addEventToHistory(event)
}

// AddDecisionTaskTimedOutEvent adds DecisionTaskTimedOut event to history
func (b *HistoryBuilder) AddDecisionTaskTimedOutEvent(
	scheduleEventID int64,
	startedEventID int64,
	timeoutType types.TimeoutType,
	baseRunID string,
	newRunID string,
	forkEventVersion int64,
	reason string,
	cause types.DecisionTaskTimedOutCause,
) *types.HistoryEvent {

	event := b.newDecisionTaskTimedOutEvent(
		scheduleEventID,
		startedEventID,
		timeoutType,
		baseRunID,
		newRunID,
		forkEventVersion,
		reason,
		cause,
	)
	return b.addEventToHistory(event)
}

// AddDecisionTaskFailedEvent adds DecisionTaskFailed event to history
func (b *HistoryBuilder) AddDecisionTaskFailedEvent(attr types.DecisionTaskFailedEventAttributes) *types.HistoryEvent {
	event := b.newDecisionTaskFailedEvent(attr)
	return b.addEventToHistory(event)
}

// AddActivityTaskScheduledEvent adds ActivityTaskScheduled event to history
func (b *HistoryBuilder) AddActivityTaskScheduledEvent(decisionCompletedEventID int64,
	attributes *types.ScheduleActivityTaskDecisionAttributes) *types.HistoryEvent {
	event := b.newActivityTaskScheduledEvent(decisionCompletedEventID, attributes)

	return b.addEventToHistory(event)
}

// AddActivityTaskStartedEvent adds ActivityTaskStarted event to history
func (b *HistoryBuilder) AddActivityTaskStartedEvent(
	scheduleEventID int64,
	attempt int32,
	requestID string,
	identity string,
	lastFailureReason string,
	lastFailureDetails []byte,
) *types.HistoryEvent {
	event := b.newActivityTaskStartedEvent(scheduleEventID, attempt, requestID, identity, lastFailureReason,
		lastFailureDetails)

	return b.addEventToHistory(event)
}

// AddActivityTaskCompletedEvent adds ActivityTaskCompleted event to history
func (b *HistoryBuilder) AddActivityTaskCompletedEvent(scheduleEventID, StartedEventID int64,
	request *types.RespondActivityTaskCompletedRequest) *types.HistoryEvent {
	event := b.newActivityTaskCompletedEvent(scheduleEventID, StartedEventID, request)

	return b.addEventToHistory(event)
}

// AddActivityTaskFailedEvent adds ActivityTaskFailed event to history
func (b *HistoryBuilder) AddActivityTaskFailedEvent(scheduleEventID, StartedEventID int64,
	request *types.RespondActivityTaskFailedRequest) *types.HistoryEvent {
	event := b.newActivityTaskFailedEvent(scheduleEventID, StartedEventID, request)

	return b.addEventToHistory(event)
}

// AddActivityTaskTimedOutEvent adds ActivityTaskTimedOut event to history
func (b *HistoryBuilder) AddActivityTaskTimedOutEvent(
	scheduleEventID,
	StartedEventID int64,
	timeoutType types.TimeoutType,
	lastHeartBeatDetails []byte,
	lastFailureReason string,
	lastFailureDetail []byte,
) *types.HistoryEvent {
	event := b.newActivityTaskTimedOutEvent(scheduleEventID, StartedEventID, timeoutType, lastHeartBeatDetails,
		lastFailureReason, lastFailureDetail)

	return b.addEventToHistory(event)
}

// AddCompletedWorkflowEvent adds WorkflowExecutionCompleted event to history
func (b *HistoryBuilder) AddCompletedWorkflowEvent(decisionCompletedEventID int64,
	attributes *types.CompleteWorkflowExecutionDecisionAttributes) *types.HistoryEvent {
	event := b.newCompleteWorkflowExecutionEvent(decisionCompletedEventID, attributes)

	return b.addEventToHistory(event)
}

// AddFailWorkflowEvent adds WorkflowExecutionFailed event to history
func (b *HistoryBuilder) AddFailWorkflowEvent(decisionCompletedEventID int64,
	attributes *types.FailWorkflowExecutionDecisionAttributes) *types.HistoryEvent {
	event := b.newFailWorkflowExecutionEvent(decisionCompletedEventID, attributes)

	return b.addEventToHistory(event)
}

// AddTimeoutWorkflowEvent adds WorkflowExecutionTimedout event to history
func (b *HistoryBuilder) AddTimeoutWorkflowEvent() *types.HistoryEvent {
	event := b.newTimeoutWorkflowExecutionEvent()

	return b.addEventToHistory(event)
}

// AddWorkflowExecutionTerminatedEvent add WorkflowExecutionTerminated event to history
func (b *HistoryBuilder) AddWorkflowExecutionTerminatedEvent(
	reason string,
	details []byte,
	identity string,
) *types.HistoryEvent {
	event := b.newWorkflowExecutionTerminatedEvent(reason, details, identity)
	return b.addEventToHistory(event)
}

// AddContinuedAsNewEvent adds WorkflowExecutionContinuedAsNew event to history
func (b *HistoryBuilder) AddContinuedAsNewEvent(decisionCompletedEventID int64, newRunID string,
	attributes *types.ContinueAsNewWorkflowExecutionDecisionAttributes) *types.HistoryEvent {
	event := b.newWorkflowExecutionContinuedAsNewEvent(decisionCompletedEventID, newRunID, attributes)

	return b.addEventToHistory(event)
}

// AddTimerStartedEvent adds TimerStart event to history
func (b *HistoryBuilder) AddTimerStartedEvent(decisionCompletedEventID int64,
	request *types.StartTimerDecisionAttributes) *types.HistoryEvent {

	attributes := &types.TimerStartedEventAttributes{}
	attributes.TimerID = common.StringPtr(*request.TimerID)
	attributes.StartToFireTimeoutSeconds = common.Int64Ptr(*request.StartToFireTimeoutSeconds)
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(decisionCompletedEventID)

	event := b.msBuilder.CreateNewHistoryEvent(types.EventTypeTimerStarted)
	event.TimerStartedEventAttributes = attributes

	return b.addEventToHistory(event)
}

// AddTimerFiredEvent adds TimerFired event to history
func (b *HistoryBuilder) AddTimerFiredEvent(
	StartedEventID int64,
	TimerID string,
) *types.HistoryEvent {

	attributes := &types.TimerFiredEventAttributes{}
	attributes.TimerID = common.StringPtr(TimerID)
	attributes.StartedEventID = common.Int64Ptr(StartedEventID)

	event := b.msBuilder.CreateNewHistoryEvent(types.EventTypeTimerFired)
	event.TimerFiredEventAttributes = attributes

	return b.addEventToHistory(event)
}

// AddActivityTaskCancelRequestedEvent add ActivityTaskCancelRequested event to history
func (b *HistoryBuilder) AddActivityTaskCancelRequestedEvent(decisionCompletedEventID int64,
	ActivityID string) *types.HistoryEvent {

	attributes := &types.ActivityTaskCancelRequestedEventAttributes{}
	attributes.ActivityID = common.StringPtr(ActivityID)
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(decisionCompletedEventID)

	event := b.msBuilder.CreateNewHistoryEvent(types.EventTypeActivityTaskCancelRequested)
	event.ActivityTaskCancelRequestedEventAttributes = attributes

	return b.addEventToHistory(event)
}

// AddRequestCancelActivityTaskFailedEvent add RequestCancelActivityTaskFailed event to history
func (b *HistoryBuilder) AddRequestCancelActivityTaskFailedEvent(decisionCompletedEventID int64,
	ActivityID string, cause string) *types.HistoryEvent {

	attributes := &types.RequestCancelActivityTaskFailedEventAttributes{}
	attributes.ActivityID = common.StringPtr(ActivityID)
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(decisionCompletedEventID)
	attributes.Cause = common.StringPtr(cause)

	event := b.msBuilder.CreateNewHistoryEvent(types.EventTypeRequestCancelActivityTaskFailed)
	event.RequestCancelActivityTaskFailedEventAttributes = attributes

	return b.addEventToHistory(event)
}

// AddActivityTaskCanceledEvent adds ActivityTaskCanceled event to history
func (b *HistoryBuilder) AddActivityTaskCanceledEvent(scheduleEventID, StartedEventID int64,
	LatestCancelRequestedEventID int64, details []byte, identity string) *types.HistoryEvent {

	attributes := &types.ActivityTaskCanceledEventAttributes{}
	attributes.ScheduledEventID = common.Int64Ptr(scheduleEventID)
	attributes.StartedEventID = common.Int64Ptr(StartedEventID)
	attributes.LatestCancelRequestedEventID = common.Int64Ptr(LatestCancelRequestedEventID)
	attributes.Details = details
	attributes.Identity = common.StringPtr(identity)

	event := b.msBuilder.CreateNewHistoryEvent(types.EventTypeActivityTaskCanceled)
	event.ActivityTaskCanceledEventAttributes = attributes

	return b.addEventToHistory(event)
}

// AddTimerCanceledEvent adds TimerCanceled event to history
func (b *HistoryBuilder) AddTimerCanceledEvent(StartedEventID int64,
	DecisionTaskCompletedEventID int64, TimerID string, identity string) *types.HistoryEvent {

	attributes := &types.TimerCanceledEventAttributes{}
	attributes.StartedEventID = common.Int64Ptr(StartedEventID)
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(DecisionTaskCompletedEventID)
	attributes.TimerID = common.StringPtr(TimerID)
	attributes.Identity = common.StringPtr(identity)

	event := b.msBuilder.CreateNewHistoryEvent(types.EventTypeTimerCanceled)
	event.TimerCanceledEventAttributes = attributes

	return b.addEventToHistory(event)
}

// AddCancelTimerFailedEvent adds CancelTimerFailed event to history
func (b *HistoryBuilder) AddCancelTimerFailedEvent(TimerID string, DecisionTaskCompletedEventID int64,
	cause string, identity string) *types.HistoryEvent {

	attributes := &types.CancelTimerFailedEventAttributes{}
	attributes.TimerID = common.StringPtr(TimerID)
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(DecisionTaskCompletedEventID)
	attributes.Cause = common.StringPtr(cause)
	attributes.Identity = common.StringPtr(identity)

	event := b.msBuilder.CreateNewHistoryEvent(types.EventTypeCancelTimerFailed)
	event.CancelTimerFailedEventAttributes = attributes

	return b.addEventToHistory(event)
}

// AddWorkflowExecutionCancelRequestedEvent adds WorkflowExecutionCancelRequested event to history
func (b *HistoryBuilder) AddWorkflowExecutionCancelRequestedEvent(cause string,
	request *types.HistoryRequestCancelWorkflowExecutionRequest) *types.HistoryEvent {
	event := b.newWorkflowExecutionCancelRequestedEvent(cause, request)

	return b.addEventToHistory(event)
}

// AddWorkflowExecutionCanceledEvent adds WorkflowExecutionCanceled event to history
func (b *HistoryBuilder) AddWorkflowExecutionCanceledEvent(DecisionTaskCompletedEventID int64,
	attributes *types.CancelWorkflowExecutionDecisionAttributes) *types.HistoryEvent {
	event := b.newWorkflowExecutionCanceledEvent(DecisionTaskCompletedEventID, attributes)

	return b.addEventToHistory(event)
}

// AddRequestCancelExternalWorkflowExecutionInitiatedEvent adds RequestCancelExternalWorkflowExecutionInitiated event to history
func (b *HistoryBuilder) AddRequestCancelExternalWorkflowExecutionInitiatedEvent(DecisionTaskCompletedEventID int64,
	request *types.RequestCancelExternalWorkflowExecutionDecisionAttributes) *types.HistoryEvent {
	event := b.newRequestCancelExternalWorkflowExecutionInitiatedEvent(DecisionTaskCompletedEventID, request)

	return b.addEventToHistory(event)
}

// AddRequestCancelExternalWorkflowExecutionFailedEvent adds RequestCancelExternalWorkflowExecutionFailed event to history
func (b *HistoryBuilder) AddRequestCancelExternalWorkflowExecutionFailedEvent(DecisionTaskCompletedEventID, initiatedEventID int64,
	domain, workflowID, runID string, cause types.CancelExternalWorkflowExecutionFailedCause) *types.HistoryEvent {
	event := b.newRequestCancelExternalWorkflowExecutionFailedEvent(DecisionTaskCompletedEventID, initiatedEventID,
		domain, workflowID, runID, cause)

	return b.addEventToHistory(event)
}

// AddExternalWorkflowExecutionCancelRequested adds ExternalWorkflowExecutionCancelRequested event to history
func (b *HistoryBuilder) AddExternalWorkflowExecutionCancelRequested(initiatedEventID int64,
	domain, workflowID, runID string) *types.HistoryEvent {
	event := b.newExternalWorkflowExecutionCancelRequestedEvent(initiatedEventID,
		domain, workflowID, runID)

	return b.addEventToHistory(event)
}

// AddSignalExternalWorkflowExecutionInitiatedEvent adds SignalExternalWorkflowExecutionInitiated event to history
func (b *HistoryBuilder) AddSignalExternalWorkflowExecutionInitiatedEvent(DecisionTaskCompletedEventID int64,
	attributes *types.SignalExternalWorkflowExecutionDecisionAttributes) *types.HistoryEvent {
	event := b.newSignalExternalWorkflowExecutionInitiatedEvent(DecisionTaskCompletedEventID, attributes)

	return b.addEventToHistory(event)
}

// AddUpsertWorkflowSearchAttributesEvent adds UpsertWorkflowSearchAttributes event to history
func (b *HistoryBuilder) AddUpsertWorkflowSearchAttributesEvent(
	DecisionTaskCompletedEventID int64,
	attributes *types.UpsertWorkflowSearchAttributesDecisionAttributes) *types.HistoryEvent {
	event := b.newUpsertWorkflowSearchAttributesEvent(DecisionTaskCompletedEventID, attributes)

	return b.addEventToHistory(event)
}

// AddSignalExternalWorkflowExecutionFailedEvent adds SignalExternalWorkflowExecutionFailed event to history
func (b *HistoryBuilder) AddSignalExternalWorkflowExecutionFailedEvent(DecisionTaskCompletedEventID, initiatedEventID int64,
	domain, workflowID, runID string, control []byte, cause types.SignalExternalWorkflowExecutionFailedCause) *types.HistoryEvent {
	event := b.newSignalExternalWorkflowExecutionFailedEvent(DecisionTaskCompletedEventID, initiatedEventID,
		domain, workflowID, runID, control, cause)

	return b.addEventToHistory(event)
}

// AddExternalWorkflowExecutionSignaled adds ExternalWorkflowExecutionSignaled event to history
func (b *HistoryBuilder) AddExternalWorkflowExecutionSignaled(initiatedEventID int64,
	domain, workflowID, runID string, control []byte) *types.HistoryEvent {
	event := b.newExternalWorkflowExecutionSignaledEvent(initiatedEventID,
		domain, workflowID, runID, control)

	return b.addEventToHistory(event)
}

// AddMarkerRecordedEvent adds MarkerRecorded event to history
func (b *HistoryBuilder) AddMarkerRecordedEvent(decisionCompletedEventID int64,
	attributes *types.RecordMarkerDecisionAttributes) *types.HistoryEvent {
	event := b.newMarkerRecordedEventAttributes(decisionCompletedEventID, attributes)

	return b.addEventToHistory(event)
}

// AddWorkflowExecutionSignaledEvent adds WorkflowExecutionSignaled event to history
func (b *HistoryBuilder) AddWorkflowExecutionSignaledEvent(
	signalName string, input []byte, identity string) *types.HistoryEvent {
	event := b.newWorkflowExecutionSignaledEvent(signalName, input, identity)

	return b.addEventToHistory(event)
}

// AddStartChildWorkflowExecutionInitiatedEvent adds ChildWorkflowExecutionInitiated event to history
func (b *HistoryBuilder) AddStartChildWorkflowExecutionInitiatedEvent(decisionCompletedEventID int64,
	attributes *types.StartChildWorkflowExecutionDecisionAttributes) *types.HistoryEvent {
	event := b.newStartChildWorkflowExecutionInitiatedEvent(decisionCompletedEventID, attributes)

	return b.addEventToHistory(event)
}

// AddChildWorkflowExecutionStartedEvent adds ChildWorkflowExecutionStarted event to history
func (b *HistoryBuilder) AddChildWorkflowExecutionStartedEvent(
	domain *string,
	execution *types.WorkflowExecution,
	workflowType *types.WorkflowType,
	initiatedID int64,
	header *types.Header,
) *types.HistoryEvent {
	event := b.newChildWorkflowExecutionStartedEvent(domain, execution, workflowType, initiatedID, header)

	return b.addEventToHistory(event)
}

// AddStartChildWorkflowExecutionFailedEvent adds ChildWorkflowExecutionFailed event to history
func (b *HistoryBuilder) AddStartChildWorkflowExecutionFailedEvent(initiatedID int64,
	cause types.ChildWorkflowExecutionFailedCause,
	initiatedEventAttributes *types.StartChildWorkflowExecutionInitiatedEventAttributes) *types.HistoryEvent {
	event := b.newStartChildWorkflowExecutionFailedEvent(initiatedID, cause, initiatedEventAttributes)

	return b.addEventToHistory(event)
}

// AddChildWorkflowExecutionCompletedEvent adds ChildWorkflowExecutionCompleted event to history
func (b *HistoryBuilder) AddChildWorkflowExecutionCompletedEvent(domain *string, execution *types.WorkflowExecution,
	workflowType *types.WorkflowType, initiatedID, startedID int64,
	completedAttributes *types.WorkflowExecutionCompletedEventAttributes) *types.HistoryEvent {
	event := b.newChildWorkflowExecutionCompletedEvent(domain, execution, workflowType, initiatedID, startedID,
		completedAttributes)

	return b.addEventToHistory(event)
}

// AddChildWorkflowExecutionFailedEvent adds ChildWorkflowExecutionFailed event to history
func (b *HistoryBuilder) AddChildWorkflowExecutionFailedEvent(domain *string, execution *types.WorkflowExecution,
	workflowType *types.WorkflowType, initiatedID, startedID int64,
	failedAttributes *types.WorkflowExecutionFailedEventAttributes) *types.HistoryEvent {
	event := b.newChildWorkflowExecutionFailedEvent(domain, execution, workflowType, initiatedID, startedID,
		failedAttributes)

	return b.addEventToHistory(event)
}

// AddChildWorkflowExecutionCanceledEvent adds ChildWorkflowExecutionCanceled event to history
func (b *HistoryBuilder) AddChildWorkflowExecutionCanceledEvent(domain *string, execution *types.WorkflowExecution,
	workflowType *types.WorkflowType, initiatedID, startedID int64,
	canceledAttributes *types.WorkflowExecutionCanceledEventAttributes) *types.HistoryEvent {
	event := b.newChildWorkflowExecutionCanceledEvent(domain, execution, workflowType, initiatedID, startedID,
		canceledAttributes)

	return b.addEventToHistory(event)
}

// AddChildWorkflowExecutionTerminatedEvent adds ChildWorkflowExecutionTerminated event to history
func (b *HistoryBuilder) AddChildWorkflowExecutionTerminatedEvent(domain *string, execution *types.WorkflowExecution,
	workflowType *types.WorkflowType, initiatedID, startedID int64,
	terminatedAttributes *types.WorkflowExecutionTerminatedEventAttributes) *types.HistoryEvent {
	event := b.newChildWorkflowExecutionTerminatedEvent(domain, execution, workflowType, initiatedID, startedID,
		terminatedAttributes)

	return b.addEventToHistory(event)
}

// AddChildWorkflowExecutionTimedOutEvent adds ChildWorkflowExecutionTimedOut event to history
func (b *HistoryBuilder) AddChildWorkflowExecutionTimedOutEvent(domain *string, execution *types.WorkflowExecution,
	workflowType *types.WorkflowType, initiatedID, startedID int64,
	timedOutAttributes *types.WorkflowExecutionTimedOutEventAttributes) *types.HistoryEvent {
	event := b.newChildWorkflowExecutionTimedOutEvent(domain, execution, workflowType, initiatedID, startedID,
		timedOutAttributes)

	return b.addEventToHistory(event)
}

func (b *HistoryBuilder) addEventToHistory(event *types.HistoryEvent) *types.HistoryEvent {
	b.history = append(b.history, event)
	return event
}

func (b *HistoryBuilder) addTransientEvent(event *types.HistoryEvent) *types.HistoryEvent {
	b.transientHistory = append(b.transientHistory, event)
	return event
}

func (b *HistoryBuilder) newWorkflowExecutionStartedEvent(
	startRequest *types.HistoryStartWorkflowExecutionRequest, previousExecution *persistence.WorkflowExecutionInfo, firstRunID, originalRunID string) *types.HistoryEvent {
	var prevRunID *string
	var resetPoints *types.ResetPoints
	if previousExecution != nil {
		prevRunID = common.StringPtr(previousExecution.RunID)
		resetPoints = previousExecution.AutoResetPoints
	}
	request := startRequest.StartRequest
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeWorkflowExecutionStarted)
	attributes := &types.WorkflowExecutionStartedEventAttributes{}
	attributes.WorkflowType = request.WorkflowType
	attributes.TaskList = request.TaskList
	attributes.Header = request.Header
	attributes.Input = request.Input
	attributes.ExecutionStartToCloseTimeoutSeconds = common.Int32Ptr(*request.ExecutionStartToCloseTimeoutSeconds)
	attributes.TaskStartToCloseTimeoutSeconds = common.Int32Ptr(*request.TaskStartToCloseTimeoutSeconds)
	attributes.ContinuedExecutionRunID = prevRunID
	attributes.PrevAutoResetPoints = resetPoints
	attributes.Identity = common.StringPtr(common.StringDefault(request.Identity))
	attributes.RetryPolicy = request.RetryPolicy
	attributes.Attempt = common.Int32Ptr(startRequest.GetAttempt())
	attributes.ExpirationTimestamp = startRequest.ExpirationTimestamp
	attributes.CronSchedule = request.CronSchedule
	attributes.LastCompletionResult = startRequest.LastCompletionResult
	attributes.ContinuedFailureReason = startRequest.ContinuedFailureReason
	attributes.ContinuedFailureDetails = startRequest.ContinuedFailureDetails
	attributes.Initiator = startRequest.ContinueAsNewInitiator
	attributes.FirstDecisionTaskBackoffSeconds = startRequest.FirstDecisionTaskBackoffSeconds
	attributes.FirstExecutionRunID = common.StringPtr(firstRunID)
	attributes.OriginalExecutionRunID = common.StringPtr(originalRunID)
	attributes.Memo = request.Memo
	attributes.SearchAttributes = request.SearchAttributes

	parentInfo := startRequest.ParentExecutionInfo
	if parentInfo != nil {
		attributes.ParentWorkflowDomain = parentInfo.Domain
		attributes.ParentWorkflowExecution = parentInfo.Execution
		attributes.ParentInitiatedEventID = parentInfo.InitiatedID
	}
	historyEvent.WorkflowExecutionStartedEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newDecisionTaskScheduledEvent(taskList string, startToCloseTimeoutSeconds int32,
	attempt int64) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeDecisionTaskScheduled)

	return setDecisionTaskScheduledEventInfo(historyEvent, taskList, startToCloseTimeoutSeconds, attempt)
}

func (b *HistoryBuilder) newTransientDecisionTaskScheduledEvent(taskList string, startToCloseTimeoutSeconds int32,
	attempt int64, timestamp int64) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEventWithTimestamp(types.EventTypeDecisionTaskScheduled, timestamp)

	return setDecisionTaskScheduledEventInfo(historyEvent, taskList, startToCloseTimeoutSeconds, attempt)
}

func (b *HistoryBuilder) newDecisionTaskStartedEvent(ScheduledEventID int64, requestID string,
	identity string) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeDecisionTaskStarted)

	return setDecisionTaskStartedEventInfo(historyEvent, ScheduledEventID, requestID, identity)
}

func (b *HistoryBuilder) newTransientDecisionTaskStartedEvent(ScheduledEventID int64, requestID string,
	identity string, timestamp int64) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEventWithTimestamp(types.EventTypeDecisionTaskStarted, timestamp)

	return setDecisionTaskStartedEventInfo(historyEvent, ScheduledEventID, requestID, identity)
}

func (b *HistoryBuilder) newDecisionTaskCompletedEvent(scheduleEventID, StartedEventID int64,
	request *types.RespondDecisionTaskCompletedRequest) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeDecisionTaskCompleted)
	attributes := &types.DecisionTaskCompletedEventAttributes{}
	attributes.ExecutionContext = request.ExecutionContext
	attributes.ScheduledEventID = common.Int64Ptr(scheduleEventID)
	attributes.StartedEventID = common.Int64Ptr(StartedEventID)
	attributes.Identity = common.StringPtr(common.StringDefault(request.Identity))
	attributes.BinaryChecksum = request.BinaryChecksum
	historyEvent.DecisionTaskCompletedEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newDecisionTaskTimedOutEvent(
	scheduleEventID int64,
	startedEventID int64,
	timeoutType types.TimeoutType,
	baseRunID string,
	newRunID string,
	forkEventVersion int64,
	reason string,
	cause types.DecisionTaskTimedOutCause,
) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeDecisionTaskTimedOut)
	attributes := &types.DecisionTaskTimedOutEventAttributes{}
	attributes.ScheduledEventID = common.Int64Ptr(scheduleEventID)
	attributes.StartedEventID = common.Int64Ptr(startedEventID)
	attributes.TimeoutType = timeoutType.Ptr()
	attributes.BaseRunID = common.StringPtr(baseRunID)
	attributes.NewRunID = common.StringPtr(newRunID)
	attributes.ForkEventVersion = common.Int64Ptr(forkEventVersion)
	attributes.Reason = common.StringPtr(reason)
	attributes.Cause = cause.Ptr()
	historyEvent.DecisionTaskTimedOutEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newDecisionTaskFailedEvent(attr types.DecisionTaskFailedEventAttributes) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeDecisionTaskFailed)
	historyEvent.DecisionTaskFailedEventAttributes = &attr
	return historyEvent
}

func (b *HistoryBuilder) newActivityTaskScheduledEvent(DecisionTaskCompletedEventID int64,
	scheduleAttributes *types.ScheduleActivityTaskDecisionAttributes) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeActivityTaskScheduled)
	attributes := &types.ActivityTaskScheduledEventAttributes{}
	attributes.ActivityID = common.StringPtr(common.StringDefault(scheduleAttributes.ActivityID))
	attributes.ActivityType = scheduleAttributes.ActivityType
	attributes.TaskList = scheduleAttributes.TaskList
	attributes.Header = scheduleAttributes.Header
	attributes.Input = scheduleAttributes.Input
	attributes.ScheduleToCloseTimeoutSeconds = common.Int32Ptr(common.Int32Default(scheduleAttributes.ScheduleToCloseTimeoutSeconds))
	attributes.ScheduleToStartTimeoutSeconds = common.Int32Ptr(common.Int32Default(scheduleAttributes.ScheduleToStartTimeoutSeconds))
	attributes.StartToCloseTimeoutSeconds = common.Int32Ptr(common.Int32Default(scheduleAttributes.StartToCloseTimeoutSeconds))
	attributes.HeartbeatTimeoutSeconds = common.Int32Ptr(common.Int32Default(scheduleAttributes.HeartbeatTimeoutSeconds))
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(DecisionTaskCompletedEventID)
	attributes.RetryPolicy = scheduleAttributes.RetryPolicy
	historyEvent.ActivityTaskScheduledEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newActivityTaskStartedEvent(
	ScheduledEventID int64,
	attempt int32,
	requestID string,
	identity string,
	lastFailureReason string,
	lastFailureDetails []byte,
) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeActivityTaskStarted)
	attributes := &types.ActivityTaskStartedEventAttributes{}
	attributes.ScheduledEventID = common.Int64Ptr(ScheduledEventID)
	attributes.Attempt = common.Int32Ptr(attempt)
	attributes.Identity = common.StringPtr(identity)
	attributes.RequestID = common.StringPtr(requestID)
	attributes.LastFailureReason = common.StringPtr(lastFailureReason)
	attributes.LastFailureDetails = lastFailureDetails
	historyEvent.ActivityTaskStartedEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newActivityTaskCompletedEvent(scheduleEventID, StartedEventID int64,
	request *types.RespondActivityTaskCompletedRequest) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeActivityTaskCompleted)
	attributes := &types.ActivityTaskCompletedEventAttributes{}
	attributes.Result = request.Result
	attributes.ScheduledEventID = common.Int64Ptr(scheduleEventID)
	attributes.StartedEventID = common.Int64Ptr(StartedEventID)
	attributes.Identity = common.StringPtr(common.StringDefault(request.Identity))
	historyEvent.ActivityTaskCompletedEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newActivityTaskTimedOutEvent(
	scheduleEventID, StartedEventID int64,
	timeoutType types.TimeoutType,
	lastHeartBeatDetails []byte,
	lastFailureReason string,
	lastFailureDetail []byte,
) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeActivityTaskTimedOut)
	attributes := &types.ActivityTaskTimedOutEventAttributes{}
	attributes.ScheduledEventID = common.Int64Ptr(scheduleEventID)
	attributes.StartedEventID = common.Int64Ptr(StartedEventID)
	attributes.TimeoutType = &timeoutType
	attributes.Details = lastHeartBeatDetails
	attributes.LastFailureReason = common.StringPtr(lastFailureReason)
	attributes.LastFailureDetails = lastFailureDetail

	historyEvent.ActivityTaskTimedOutEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newActivityTaskFailedEvent(scheduleEventID, StartedEventID int64,
	request *types.RespondActivityTaskFailedRequest) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeActivityTaskFailed)
	attributes := &types.ActivityTaskFailedEventAttributes{}
	attributes.Reason = common.StringPtr(common.StringDefault(request.Reason))
	attributes.Details = request.Details
	attributes.ScheduledEventID = common.Int64Ptr(scheduleEventID)
	attributes.StartedEventID = common.Int64Ptr(StartedEventID)
	attributes.Identity = common.StringPtr(common.StringDefault(request.Identity))
	historyEvent.ActivityTaskFailedEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newCompleteWorkflowExecutionEvent(decisionTaskCompletedEventID int64,
	request *types.CompleteWorkflowExecutionDecisionAttributes) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeWorkflowExecutionCompleted)
	attributes := &types.WorkflowExecutionCompletedEventAttributes{}
	attributes.Result = request.Result
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(decisionTaskCompletedEventID)
	historyEvent.WorkflowExecutionCompletedEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newFailWorkflowExecutionEvent(decisionTaskCompletedEventID int64,
	request *types.FailWorkflowExecutionDecisionAttributes) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeWorkflowExecutionFailed)
	attributes := &types.WorkflowExecutionFailedEventAttributes{}
	attributes.Reason = common.StringPtr(common.StringDefault(request.Reason))
	attributes.Details = request.Details
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(decisionTaskCompletedEventID)
	historyEvent.WorkflowExecutionFailedEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newTimeoutWorkflowExecutionEvent() *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeWorkflowExecutionTimedOut)
	attributes := &types.WorkflowExecutionTimedOutEventAttributes{}
	attributes.TimeoutType = types.TimeoutTypeStartToClose.Ptr()
	historyEvent.WorkflowExecutionTimedOutEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newWorkflowExecutionSignaledEvent(
	signalName string, input []byte, identity string) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeWorkflowExecutionSignaled)
	attributes := &types.WorkflowExecutionSignaledEventAttributes{}
	attributes.SignalName = common.StringPtr(signalName)
	attributes.Input = input
	attributes.Identity = common.StringPtr(identity)
	historyEvent.WorkflowExecutionSignaledEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newWorkflowExecutionTerminatedEvent(
	reason string, details []byte, identity string) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeWorkflowExecutionTerminated)
	attributes := &types.WorkflowExecutionTerminatedEventAttributes{}
	attributes.Reason = common.StringPtr(reason)
	attributes.Details = details
	attributes.Identity = common.StringPtr(identity)
	historyEvent.WorkflowExecutionTerminatedEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newMarkerRecordedEventAttributes(DecisionTaskCompletedEventID int64,
	request *types.RecordMarkerDecisionAttributes) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeMarkerRecorded)
	attributes := &types.MarkerRecordedEventAttributes{}
	attributes.MarkerName = common.StringPtr(common.StringDefault(request.MarkerName))
	attributes.Details = request.Details
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(DecisionTaskCompletedEventID)
	attributes.Header = request.Header
	historyEvent.MarkerRecordedEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newWorkflowExecutionCancelRequestedEvent(cause string,
	request *types.HistoryRequestCancelWorkflowExecutionRequest) *types.HistoryEvent {
	event := b.msBuilder.CreateNewHistoryEvent(types.EventTypeWorkflowExecutionCancelRequested)
	attributes := &types.WorkflowExecutionCancelRequestedEventAttributes{}
	attributes.Cause = common.StringPtr(cause)
	attributes.Identity = common.StringPtr(common.StringDefault(request.CancelRequest.Identity))
	if request.ExternalInitiatedEventID != nil {
		attributes.ExternalInitiatedEventID = common.Int64Ptr(*request.ExternalInitiatedEventID)
	}
	if request.ExternalWorkflowExecution != nil {
		attributes.ExternalWorkflowExecution = request.ExternalWorkflowExecution
	}
	event.WorkflowExecutionCancelRequestedEventAttributes = attributes

	return event
}

func (b *HistoryBuilder) newWorkflowExecutionCanceledEvent(DecisionTaskCompletedEventID int64,
	request *types.CancelWorkflowExecutionDecisionAttributes) *types.HistoryEvent {
	event := b.msBuilder.CreateNewHistoryEvent(types.EventTypeWorkflowExecutionCanceled)
	attributes := &types.WorkflowExecutionCanceledEventAttributes{}
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(DecisionTaskCompletedEventID)
	attributes.Details = request.Details
	event.WorkflowExecutionCanceledEventAttributes = attributes

	return event
}

func (b *HistoryBuilder) newRequestCancelExternalWorkflowExecutionInitiatedEvent(DecisionTaskCompletedEventID int64,
	request *types.RequestCancelExternalWorkflowExecutionDecisionAttributes) *types.HistoryEvent {
	event := b.msBuilder.CreateNewHistoryEvent(types.EventTypeRequestCancelExternalWorkflowExecutionInitiated)
	attributes := &types.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{}
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(DecisionTaskCompletedEventID)
	attributes.Domain = request.Domain
	attributes.WorkflowExecution = &types.WorkflowExecution{
		WorkflowID: request.WorkflowID,
		RunID:      request.RunID,
	}
	attributes.Control = request.Control
	attributes.ChildWorkflowOnly = request.ChildWorkflowOnly
	event.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes = attributes

	return event
}

func (b *HistoryBuilder) newRequestCancelExternalWorkflowExecutionFailedEvent(decisionTaskCompletedEventID, initiatedEventID int64,
	domain, workflowID, runID string, cause types.CancelExternalWorkflowExecutionFailedCause) *types.HistoryEvent {
	event := b.msBuilder.CreateNewHistoryEvent(types.EventTypeRequestCancelExternalWorkflowExecutionFailed)
	attributes := &types.RequestCancelExternalWorkflowExecutionFailedEventAttributes{}
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(decisionTaskCompletedEventID)
	attributes.InitiatedEventID = common.Int64Ptr(initiatedEventID)
	attributes.Domain = common.StringPtr(domain)
	attributes.WorkflowExecution = &types.WorkflowExecution{
		WorkflowID: common.StringPtr(workflowID),
		RunID:      common.StringPtr(runID),
	}
	attributes.Cause = cause.Ptr()
	event.RequestCancelExternalWorkflowExecutionFailedEventAttributes = attributes

	return event
}

func (b *HistoryBuilder) newExternalWorkflowExecutionCancelRequestedEvent(initiatedEventID int64,
	domain, workflowID, runID string) *types.HistoryEvent {
	event := b.msBuilder.CreateNewHistoryEvent(types.EventTypeExternalWorkflowExecutionCancelRequested)
	attributes := &types.ExternalWorkflowExecutionCancelRequestedEventAttributes{}
	attributes.InitiatedEventID = common.Int64Ptr(initiatedEventID)
	attributes.Domain = common.StringPtr(domain)
	attributes.WorkflowExecution = &types.WorkflowExecution{
		WorkflowID: common.StringPtr(workflowID),
		RunID:      common.StringPtr(runID),
	}
	event.ExternalWorkflowExecutionCancelRequestedEventAttributes = attributes

	return event
}

func (b *HistoryBuilder) newSignalExternalWorkflowExecutionInitiatedEvent(decisionTaskCompletedEventID int64,
	request *types.SignalExternalWorkflowExecutionDecisionAttributes) *types.HistoryEvent {
	event := b.msBuilder.CreateNewHistoryEvent(types.EventTypeSignalExternalWorkflowExecutionInitiated)
	attributes := &types.SignalExternalWorkflowExecutionInitiatedEventAttributes{}
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(decisionTaskCompletedEventID)
	attributes.Domain = request.Domain
	attributes.WorkflowExecution = &types.WorkflowExecution{
		WorkflowID: request.Execution.WorkflowID,
		RunID:      request.Execution.RunID,
	}
	attributes.SignalName = common.StringPtr(request.GetSignalName())
	attributes.Input = request.Input
	attributes.Control = request.Control
	attributes.ChildWorkflowOnly = request.ChildWorkflowOnly
	event.SignalExternalWorkflowExecutionInitiatedEventAttributes = attributes

	return event
}

func (b *HistoryBuilder) newUpsertWorkflowSearchAttributesEvent(decisionTaskCompletedEventID int64,
	request *types.UpsertWorkflowSearchAttributesDecisionAttributes) *types.HistoryEvent {
	event := b.msBuilder.CreateNewHistoryEvent(types.EventTypeUpsertWorkflowSearchAttributes)
	attributes := &types.UpsertWorkflowSearchAttributesEventAttributes{}
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(decisionTaskCompletedEventID)
	attributes.SearchAttributes = request.GetSearchAttributes()
	event.UpsertWorkflowSearchAttributesEventAttributes = attributes

	return event
}

func (b *HistoryBuilder) newSignalExternalWorkflowExecutionFailedEvent(decisionTaskCompletedEventID, initiatedEventID int64,
	domain, workflowID, runID string, control []byte, cause types.SignalExternalWorkflowExecutionFailedCause) *types.HistoryEvent {
	event := b.msBuilder.CreateNewHistoryEvent(types.EventTypeSignalExternalWorkflowExecutionFailed)
	attributes := &types.SignalExternalWorkflowExecutionFailedEventAttributes{}
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(decisionTaskCompletedEventID)
	attributes.InitiatedEventID = common.Int64Ptr(initiatedEventID)
	attributes.Domain = common.StringPtr(domain)
	attributes.WorkflowExecution = &types.WorkflowExecution{
		WorkflowID: common.StringPtr(workflowID),
		RunID:      common.StringPtr(runID),
	}
	attributes.Cause = cause.Ptr()
	attributes.Control = control
	event.SignalExternalWorkflowExecutionFailedEventAttributes = attributes

	return event
}

func (b *HistoryBuilder) newExternalWorkflowExecutionSignaledEvent(initiatedEventID int64,
	domain, workflowID, runID string, control []byte) *types.HistoryEvent {
	event := b.msBuilder.CreateNewHistoryEvent(types.EventTypeExternalWorkflowExecutionSignaled)
	attributes := &types.ExternalWorkflowExecutionSignaledEventAttributes{}
	attributes.InitiatedEventID = common.Int64Ptr(initiatedEventID)
	attributes.Domain = common.StringPtr(domain)
	attributes.WorkflowExecution = &types.WorkflowExecution{
		WorkflowID: common.StringPtr(workflowID),
		RunID:      common.StringPtr(runID),
	}
	attributes.Control = control
	event.ExternalWorkflowExecutionSignaledEventAttributes = attributes

	return event
}

func (b *HistoryBuilder) newWorkflowExecutionContinuedAsNewEvent(decisionTaskCompletedEventID int64,
	newRunID string, request *types.ContinueAsNewWorkflowExecutionDecisionAttributes) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeWorkflowExecutionContinuedAsNew)
	attributes := &types.WorkflowExecutionContinuedAsNewEventAttributes{}
	attributes.NewExecutionRunID = common.StringPtr(newRunID)
	attributes.WorkflowType = request.WorkflowType
	attributes.TaskList = request.TaskList
	attributes.Header = request.Header
	attributes.Input = request.Input
	attributes.ExecutionStartToCloseTimeoutSeconds = common.Int32Ptr(*request.ExecutionStartToCloseTimeoutSeconds)
	attributes.TaskStartToCloseTimeoutSeconds = common.Int32Ptr(*request.TaskStartToCloseTimeoutSeconds)
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(decisionTaskCompletedEventID)
	attributes.BackoffStartIntervalInSeconds = common.Int32Ptr(request.GetBackoffStartIntervalInSeconds())
	attributes.Initiator = request.Initiator
	if attributes.Initiator == nil {
		attributes.Initiator = types.ContinueAsNewInitiatorDecider.Ptr()
	}
	attributes.FailureReason = request.FailureReason
	attributes.FailureDetails = request.FailureDetails
	attributes.LastCompletionResult = request.LastCompletionResult
	attributes.Memo = request.Memo
	attributes.SearchAttributes = request.SearchAttributes
	historyEvent.WorkflowExecutionContinuedAsNewEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newStartChildWorkflowExecutionInitiatedEvent(decisionTaskCompletedEventID int64,
	startAttributes *types.StartChildWorkflowExecutionDecisionAttributes) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeStartChildWorkflowExecutionInitiated)
	attributes := &types.StartChildWorkflowExecutionInitiatedEventAttributes{}
	attributes.Domain = startAttributes.Domain
	attributes.WorkflowID = startAttributes.WorkflowID
	attributes.WorkflowType = startAttributes.WorkflowType
	attributes.TaskList = startAttributes.TaskList
	attributes.Header = startAttributes.Header
	attributes.Input = startAttributes.Input
	attributes.ExecutionStartToCloseTimeoutSeconds = startAttributes.ExecutionStartToCloseTimeoutSeconds
	attributes.TaskStartToCloseTimeoutSeconds = startAttributes.TaskStartToCloseTimeoutSeconds
	attributes.Control = startAttributes.Control
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(decisionTaskCompletedEventID)
	attributes.WorkflowIDReusePolicy = startAttributes.WorkflowIDReusePolicy
	attributes.RetryPolicy = startAttributes.RetryPolicy
	attributes.CronSchedule = startAttributes.CronSchedule
	attributes.Memo = startAttributes.Memo
	attributes.SearchAttributes = startAttributes.SearchAttributes
	attributes.ParentClosePolicy = startAttributes.GetParentClosePolicy().Ptr()
	historyEvent.StartChildWorkflowExecutionInitiatedEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newChildWorkflowExecutionStartedEvent(
	domain *string,
	execution *types.WorkflowExecution,
	workflowType *types.WorkflowType,
	initiatedID int64,
	header *types.Header,
) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeChildWorkflowExecutionStarted)
	attributes := &types.ChildWorkflowExecutionStartedEventAttributes{}
	attributes.Domain = domain
	attributes.WorkflowExecution = execution
	attributes.WorkflowType = workflowType
	attributes.InitiatedEventID = common.Int64Ptr(initiatedID)
	attributes.Header = header
	historyEvent.ChildWorkflowExecutionStartedEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newStartChildWorkflowExecutionFailedEvent(initiatedID int64,
	cause types.ChildWorkflowExecutionFailedCause,
	initiatedEventAttributes *types.StartChildWorkflowExecutionInitiatedEventAttributes) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeStartChildWorkflowExecutionFailed)
	attributes := &types.StartChildWorkflowExecutionFailedEventAttributes{}
	attributes.Domain = common.StringPtr(*initiatedEventAttributes.Domain)
	attributes.WorkflowID = common.StringPtr(*initiatedEventAttributes.WorkflowID)
	attributes.WorkflowType = initiatedEventAttributes.WorkflowType
	attributes.InitiatedEventID = common.Int64Ptr(initiatedID)
	attributes.DecisionTaskCompletedEventID = common.Int64Ptr(*initiatedEventAttributes.DecisionTaskCompletedEventID)
	attributes.Control = initiatedEventAttributes.Control
	attributes.Cause = cause.Ptr()
	historyEvent.StartChildWorkflowExecutionFailedEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newChildWorkflowExecutionCompletedEvent(domain *string, execution *types.WorkflowExecution,
	workflowType *types.WorkflowType, initiatedID, startedID int64,
	completedAttributes *types.WorkflowExecutionCompletedEventAttributes) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeChildWorkflowExecutionCompleted)
	attributes := &types.ChildWorkflowExecutionCompletedEventAttributes{}
	attributes.Domain = domain
	attributes.WorkflowExecution = execution
	attributes.WorkflowType = workflowType
	attributes.InitiatedEventID = common.Int64Ptr(initiatedID)
	attributes.StartedEventID = common.Int64Ptr(startedID)
	attributes.Result = completedAttributes.Result
	historyEvent.ChildWorkflowExecutionCompletedEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newChildWorkflowExecutionFailedEvent(domain *string, execution *types.WorkflowExecution,
	workflowType *types.WorkflowType, initiatedID, startedID int64,
	failedAttributes *types.WorkflowExecutionFailedEventAttributes) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeChildWorkflowExecutionFailed)
	attributes := &types.ChildWorkflowExecutionFailedEventAttributes{}
	attributes.Domain = domain
	attributes.WorkflowExecution = execution
	attributes.WorkflowType = workflowType
	attributes.InitiatedEventID = common.Int64Ptr(initiatedID)
	attributes.StartedEventID = common.Int64Ptr(startedID)
	attributes.Reason = common.StringPtr(common.StringDefault(failedAttributes.Reason))
	attributes.Details = failedAttributes.Details
	historyEvent.ChildWorkflowExecutionFailedEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newChildWorkflowExecutionCanceledEvent(domain *string, execution *types.WorkflowExecution,
	workflowType *types.WorkflowType, initiatedID, startedID int64,
	canceledAttributes *types.WorkflowExecutionCanceledEventAttributes) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeChildWorkflowExecutionCanceled)
	attributes := &types.ChildWorkflowExecutionCanceledEventAttributes{}
	attributes.Domain = domain
	attributes.WorkflowExecution = execution
	attributes.WorkflowType = workflowType
	attributes.InitiatedEventID = common.Int64Ptr(initiatedID)
	attributes.StartedEventID = common.Int64Ptr(startedID)
	attributes.Details = canceledAttributes.Details
	historyEvent.ChildWorkflowExecutionCanceledEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newChildWorkflowExecutionTerminatedEvent(domain *string, execution *types.WorkflowExecution,
	workflowType *types.WorkflowType, initiatedID, startedID int64,
	terminatedAttributes *types.WorkflowExecutionTerminatedEventAttributes) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeChildWorkflowExecutionTerminated)
	attributes := &types.ChildWorkflowExecutionTerminatedEventAttributes{}
	attributes.Domain = domain
	attributes.WorkflowExecution = execution
	attributes.WorkflowType = workflowType
	attributes.InitiatedEventID = common.Int64Ptr(initiatedID)
	attributes.StartedEventID = common.Int64Ptr(startedID)
	historyEvent.ChildWorkflowExecutionTerminatedEventAttributes = attributes

	return historyEvent
}

func (b *HistoryBuilder) newChildWorkflowExecutionTimedOutEvent(domain *string, execution *types.WorkflowExecution,
	workflowType *types.WorkflowType, initiatedID, startedID int64,
	timedOutAttributes *types.WorkflowExecutionTimedOutEventAttributes) *types.HistoryEvent {
	historyEvent := b.msBuilder.CreateNewHistoryEvent(types.EventTypeChildWorkflowExecutionTimedOut)
	attributes := &types.ChildWorkflowExecutionTimedOutEventAttributes{}
	attributes.Domain = domain
	attributes.TimeoutType = timedOutAttributes.TimeoutType
	attributes.WorkflowExecution = execution
	attributes.WorkflowType = workflowType
	attributes.InitiatedEventID = common.Int64Ptr(initiatedID)
	attributes.StartedEventID = common.Int64Ptr(startedID)
	historyEvent.ChildWorkflowExecutionTimedOutEventAttributes = attributes

	return historyEvent
}

func newDecisionTaskScheduledEventWithInfo(eventID, timestamp int64, taskList string, startToCloseTimeoutSeconds int32,
	attempt int64) *types.HistoryEvent {
	historyEvent := createNewHistoryEvent(eventID, types.EventTypeDecisionTaskScheduled, timestamp)

	return setDecisionTaskScheduledEventInfo(historyEvent, taskList, startToCloseTimeoutSeconds, attempt)
}

func newDecisionTaskStartedEventWithInfo(eventID, timestamp int64, ScheduledEventID int64, requestID string,
	identity string) *types.HistoryEvent {
	historyEvent := createNewHistoryEvent(eventID, types.EventTypeDecisionTaskStarted, timestamp)

	return setDecisionTaskStartedEventInfo(historyEvent, ScheduledEventID, requestID, identity)
}

func createNewHistoryEvent(eventID int64, eventType types.EventType, timestamp int64) *types.HistoryEvent {
	historyEvent := &types.HistoryEvent{}
	historyEvent.EventID = common.Int64Ptr(eventID)
	historyEvent.Timestamp = common.Int64Ptr(timestamp)
	historyEvent.EventType = eventType.Ptr()

	return historyEvent
}

func setDecisionTaskScheduledEventInfo(historyEvent *types.HistoryEvent, taskList string,
	startToCloseTimeoutSeconds int32, attempt int64) *types.HistoryEvent {
	attributes := &types.DecisionTaskScheduledEventAttributes{}
	attributes.TaskList = &types.TaskList{}
	attributes.TaskList.Name = common.StringPtr(taskList)
	attributes.StartToCloseTimeoutSeconds = common.Int32Ptr(startToCloseTimeoutSeconds)
	attributes.Attempt = common.Int64Ptr(attempt)
	historyEvent.DecisionTaskScheduledEventAttributes = attributes

	return historyEvent
}

func setDecisionTaskStartedEventInfo(historyEvent *types.HistoryEvent, ScheduledEventID int64, requestID string,
	identity string) *types.HistoryEvent {
	attributes := &types.DecisionTaskStartedEventAttributes{}
	attributes.ScheduledEventID = common.Int64Ptr(ScheduledEventID)
	attributes.Identity = common.StringPtr(identity)
	attributes.RequestID = common.StringPtr(requestID)
	historyEvent.DecisionTaskStartedEventAttributes = attributes

	return historyEvent
}

// GetHistory gets workflow history stored inside history builder
func (b *HistoryBuilder) GetHistory() *types.History {
	history := types.History{Events: b.history}
	return &history
}

// SetHistory sets workflow history inside history builder
func (b *HistoryBuilder) SetHistory(history *types.History) {
	b.history = history.Events
}
