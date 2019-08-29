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

package common

import (
	"github.com/pborman/uuid"
	"github.com/uber/cadence/.gen/go/shared"
	"reflect"
	"time"
)

const (
	timeout                = int32(1)
	cause                  = "NDC test"
	signal                 = "NDC signal"
	checksum               = "NDC checksum"
	childWorkflowPrefix    = "child-"
	externalWorkflowPrefix = "external-"
	reason                 = "NDC reason"
)

type (
	// HistoryAttributesGenerator is to generator history attribute in history events
	HistoryAttributesGenerator interface {
		GenerateHistoryEvents([]NDCTestBatch, int64, int64) []*shared.History
	}

	// HistoryAttributesGeneratorImpl is to generator history attribute in history events
	HistoryAttributesGeneratorImpl struct {
		decisionTaskAttempts                    int64
		timerIDs                                map[string]bool
		timerStartEventIDs                      map[int64]bool
		childWorkflowInitialEventIDs            map[int64]int64
		childWorkflowStartEventIDs              map[int64]int64
		childWorkflowRunIDs                     map[int64]string
		signalExternalWorkflowEventIDs          map[int64]int64
		requestExternalWorkflowCanceledEventIDs map[int64]int64
		activityScheduleEventIDs                map[int64]bool
		activityStartEventIDs                   map[int64]int64
		activityCancelRequestEventIDs           map[int64]int64
		decisionTaskScheduleEventID             int64
		decisionTaskStartEventID                int64
		decisionTaskCompleteEventID             int64

		identity     string
		workflowID   string
		runID        string
		taskList     string
		workflowType string
		domainID     string
		domain       string
	}
)

// NewHistoryAttributesGenerator is initial a generator
func NewHistoryAttributesGenerator(
	workflowID,
	runID,
	taskList,
	workflowType,
	domainID,
	domain,
	identity string,
) HistoryAttributesGenerator {

	return &HistoryAttributesGeneratorImpl{
		decisionTaskAttempts:                    int64(0),
		timerIDs:                                make(map[string]bool),
		timerStartEventIDs:                      make(map[int64]bool),
		childWorkflowInitialEventIDs:            make(map[int64]int64),
		childWorkflowStartEventIDs:              make(map[int64]int64),
		childWorkflowRunIDs:                     make(map[int64]string),
		signalExternalWorkflowEventIDs:          make(map[int64]int64),
		requestExternalWorkflowCanceledEventIDs: make(map[int64]int64),
		activityScheduleEventIDs:                make(map[int64]bool),
		activityStartEventIDs:                   make(map[int64]int64),
		activityCancelRequestEventIDs:           make(map[int64]int64),
		decisionTaskScheduleEventID:             -1,
		decisionTaskStartEventID:                -1,
		decisionTaskCompleteEventID:             -1,
		identity:                                identity,
		workflowID:                              workflowID,
		runID:                                   runID,
		taskList:                                taskList,
		workflowType:                            workflowType,
		domainID:                                domainID,
		domain:                                  domain,
	}
}

// GenerateHistoryEvents is to generator batches of history events
func (h *HistoryAttributesGeneratorImpl) GenerateHistoryEvents(
	batches []NDCTestBatch,
	startEventID int64,
	version int64,
) []*shared.History {

	history := make([]*shared.History, 0, len(batches))
	eventID := startEventID

	//TODO: Marker and EventTypeUpsertWorkflowSearchAttributes need to be added to the model and also to generate event attributes
	for _, batch := range batches {
		historyEvents := make([]*shared.HistoryEvent, 0)
		for _, event := range batch.Events {
			historyEvent := h.generateEventAttribute(event, eventID, version)
			eventID++
			historyEvents = append(historyEvents, historyEvent)
		}
		nextHistoryBatch := &shared.History{
			Events: historyEvents,
		}
		history = append(history, nextHistoryBatch)
	}
	return history
}

func (h *HistoryAttributesGeneratorImpl) generateEventAttribute(vertex Vertex, eventID, version int64) *shared.HistoryEvent {
	historyEvent := getDefaultHistoryEvent(eventID, version)
	switch vertex.GetName() {
	case shared.EventTypeWorkflowExecutionStarted.String():
		historyEvent.EventType = shared.EventTypeWorkflowExecutionStarted.Ptr()
		historyEvent.WorkflowExecutionStartedEventAttributes = &shared.WorkflowExecutionStartedEventAttributes{
			WorkflowType: &shared.WorkflowType{
				Name: StringPtr(h.workflowType),
			},
			TaskList: &shared.TaskList{
				Name: StringPtr(h.taskList),
				Kind: shared.TaskListKindNormal.Ptr(),
			},
			ExecutionStartToCloseTimeoutSeconds: Int32Ptr(timeout),
			TaskStartToCloseTimeoutSeconds:      Int32Ptr(timeout),
			Identity:                            StringPtr(h.identity),
			FirstExecutionRunId:                 StringPtr(h.runID),
		}
	case shared.EventTypeWorkflowExecutionCompleted.String():
		historyEvent.EventType = shared.EventTypeWorkflowExecutionCompleted.Ptr()
		historyEvent.WorkflowExecutionCompletedEventAttributes = &shared.WorkflowExecutionCompletedEventAttributes{
			DecisionTaskCompletedEventId: Int64Ptr(h.decisionTaskCompleteEventID),
		}
	case shared.EventTypeWorkflowExecutionFailed.String():
		historyEvent.EventType = shared.EventTypeWorkflowExecutionFailed.Ptr()
		historyEvent.WorkflowExecutionFailedEventAttributes = &shared.WorkflowExecutionFailedEventAttributes{
			DecisionTaskCompletedEventId: Int64Ptr(h.decisionTaskCompleteEventID),
		}
	case shared.EventTypeWorkflowExecutionTerminated.String():
		historyEvent.EventType = shared.EventTypeWorkflowExecutionTerminated.Ptr()
		historyEvent.WorkflowExecutionTerminatedEventAttributes = &shared.WorkflowExecutionTerminatedEventAttributes{
			Identity: StringPtr(h.identity),
		}
	case shared.EventTypeWorkflowExecutionTimedOut.String():
		historyEvent.EventType = shared.EventTypeWorkflowExecutionTimedOut.Ptr()
		historyEvent.WorkflowExecutionTimedOutEventAttributes = &shared.WorkflowExecutionTimedOutEventAttributes{
			TimeoutType: shared.TimeoutTypeStartToClose.Ptr(),
		}
	case shared.EventTypeWorkflowExecutionCancelRequested.String():
		historyEvent.EventType = shared.EventTypeWorkflowExecutionCancelRequested.Ptr()
		historyEvent.WorkflowExecutionCancelRequestedEventAttributes = &shared.WorkflowExecutionCancelRequestedEventAttributes{
			Cause:                    StringPtr(cause),
			ExternalInitiatedEventId: Int64Ptr(10),
			ExternalWorkflowExecution: &shared.WorkflowExecution{
				WorkflowId: StringPtr(h.workflowID),
				RunId:      StringPtr(h.runID),
			},
			Identity: StringPtr(h.identity),
		}
	case shared.EventTypeWorkflowExecutionCanceled.String():
		historyEvent.EventType = shared.EventTypeWorkflowExecutionCanceled.Ptr()
		historyEvent.WorkflowExecutionCanceledEventAttributes = &shared.WorkflowExecutionCanceledEventAttributes{
			DecisionTaskCompletedEventId: Int64Ptr(eventID - 1),
		}
	case shared.EventTypeWorkflowExecutionContinuedAsNew.String():
		historyEvent.EventType = shared.EventTypeWorkflowExecutionContinuedAsNew.Ptr()
		historyEvent.WorkflowExecutionContinuedAsNewEventAttributes = &shared.WorkflowExecutionContinuedAsNewEventAttributes{
			NewExecutionRunId: StringPtr(h.runID),
			WorkflowType: &shared.WorkflowType{
				Name: StringPtr(h.workflowType),
			},
			TaskList: &shared.TaskList{
				Name: StringPtr(h.taskList),
				Kind: shared.TaskListKindNormal.Ptr(),
			},
			ExecutionStartToCloseTimeoutSeconds: Int32Ptr(timeout),
			TaskStartToCloseTimeoutSeconds:      Int32Ptr(timeout),
			DecisionTaskCompletedEventId:        Int64Ptr(eventID - 1),
			Initiator:                           shared.ContinueAsNewInitiatorDecider.Ptr(),
		}
	case shared.EventTypeWorkflowExecutionSignaled.String():
		historyEvent.EventType = shared.EventTypeWorkflowExecutionSignaled.Ptr()
		historyEvent.WorkflowExecutionSignaledEventAttributes = &shared.WorkflowExecutionSignaledEventAttributes{
			SignalName: StringPtr(signal),
			Identity:   StringPtr(h.identity),
		}
	case shared.EventTypeDecisionTaskScheduled.String():
		historyEvent.EventType = shared.EventTypeDecisionTaskScheduled.Ptr()
		historyEvent.DecisionTaskScheduledEventAttributes = &shared.DecisionTaskScheduledEventAttributes{
			TaskList: &shared.TaskList{
				Name: StringPtr(h.taskList),
				Kind: shared.TaskListKindNormal.Ptr(),
			},
			StartToCloseTimeoutSeconds: Int32Ptr(timeout),
			Attempt:                    Int64Ptr(h.decisionTaskAttempts),
		}
		h.decisionTaskScheduleEventID = eventID
	case shared.EventTypeDecisionTaskStarted.String():
		historyEvent.EventType = shared.EventTypeDecisionTaskStarted.Ptr()
		historyEvent.DecisionTaskStartedEventAttributes = &shared.DecisionTaskStartedEventAttributes{
			ScheduledEventId: Int64Ptr(h.decisionTaskScheduleEventID),
			Identity:         StringPtr(h.identity),
			RequestId:        StringPtr(uuid.New()),
		}
		h.decisionTaskStartEventID = eventID
	case shared.EventTypeDecisionTaskTimedOut.String():
		historyEvent.EventType = shared.EventTypeDecisionTaskTimedOut.Ptr()
		historyEvent.DecisionTaskTimedOutEventAttributes = &shared.DecisionTaskTimedOutEventAttributes{
			ScheduledEventId: Int64Ptr(h.decisionTaskScheduleEventID),
			StartedEventId:   Int64Ptr(h.decisionTaskStartEventID),
			TimeoutType:      shared.TimeoutTypeScheduleToStart.Ptr(),
		}
		h.decisionTaskAttempts++
	case shared.EventTypeDecisionTaskFailed.String():
		historyEvent.EventType = shared.EventTypeDecisionTaskFailed.Ptr()
		historyEvent.DecisionTaskFailedEventAttributes = &shared.DecisionTaskFailedEventAttributes{
			ScheduledEventId: Int64Ptr(h.decisionTaskScheduleEventID),
			StartedEventId:   Int64Ptr(h.decisionTaskStartEventID),
			Cause:            DecisionTaskFailedCausePtr(shared.DecisionTaskFailedCauseUnhandledDecision),
			Identity:         StringPtr(h.identity),
			BaseRunId:        StringPtr(h.runID),
		}
		h.decisionTaskAttempts++
	case shared.EventTypeDecisionTaskCompleted.String():
		historyEvent.EventType = shared.EventTypeDecisionTaskCompleted.Ptr()
		h.decisionTaskAttempts = 0
		historyEvent.DecisionTaskCompletedEventAttributes = &shared.DecisionTaskCompletedEventAttributes{
			ScheduledEventId: Int64Ptr(h.decisionTaskScheduleEventID),
			StartedEventId:   Int64Ptr(h.decisionTaskStartEventID),
			Identity:         StringPtr(h.identity),
			BinaryChecksum:   StringPtr(checksum),
		}
		h.decisionTaskCompleteEventID = eventID
	case shared.EventTypeActivityTaskScheduled.String():
		historyEvent.EventType = shared.EventTypeActivityTaskScheduled.Ptr()
		historyEvent.ActivityTaskScheduledEventAttributes = &shared.ActivityTaskScheduledEventAttributes{
			ActivityId:                    StringPtr(uuid.New()),
			ActivityType:                  ActivityTypePtr(shared.ActivityType{Name: StringPtr("activity")}),
			Domain:                        StringPtr(h.domain),
			TaskList:                      TaskListPtr(shared.TaskList{Name: StringPtr(h.taskList), Kind: TaskListKindPtr(shared.TaskListKindNormal)}),
			ScheduleToCloseTimeoutSeconds: Int32Ptr(timeout),
			ScheduleToStartTimeoutSeconds: Int32Ptr(timeout),
			StartToCloseTimeoutSeconds:    Int32Ptr(timeout),
			DecisionTaskCompletedEventId:  Int64Ptr(h.decisionTaskCompleteEventID),
		}
		h.activityScheduleEventIDs[eventID] = true
	case shared.EventTypeActivityTaskStarted.String():
		if len(h.activityScheduleEventIDs) == 0 {
			panic("No activity scheduled")
		}
		activityScheduleEventID := reflect.ValueOf(h.activityScheduleEventIDs).MapKeys()[0].Int()
		delete(h.activityScheduleEventIDs, activityScheduleEventID)

		historyEvent.EventType = shared.EventTypeActivityTaskStarted.Ptr()
		historyEvent.ActivityTaskStartedEventAttributes = &shared.ActivityTaskStartedEventAttributes{
			ScheduledEventId: Int64Ptr(activityScheduleEventID),
			Identity:         StringPtr(h.identity),
			RequestId:        StringPtr(uuid.New()),
			Attempt:          Int32Ptr(0),
		}
		h.activityStartEventIDs[eventID] = activityScheduleEventID
	case shared.EventTypeActivityTaskCompleted.String():
		if len(h.activityStartEventIDs) == 0 {
			panic("No activity started before complete")
		}
		activityStartEventID := reflect.ValueOf(h.activityStartEventIDs).MapKeys()[0].Int()
		activityScheduleEventID := h.activityStartEventIDs[activityStartEventID]
		delete(h.activityStartEventIDs, activityStartEventID)

		historyEvent.EventType = shared.EventTypeActivityTaskCompleted.Ptr()
		historyEvent.ActivityTaskCompletedEventAttributes = &shared.ActivityTaskCompletedEventAttributes{
			ScheduledEventId: Int64Ptr(activityScheduleEventID),
			StartedEventId:   Int64Ptr(activityStartEventID),
			Identity:         StringPtr(h.identity),
		}
	case shared.EventTypeActivityTaskTimedOut.String():
		if len(h.activityStartEventIDs) == 0 {
			panic("No activity started before timeout")
		}
		activityStartEventID := reflect.ValueOf(h.activityStartEventIDs).MapKeys()[0].Int()
		activityScheduleEventID := h.activityStartEventIDs[activityStartEventID]
		delete(h.activityStartEventIDs, activityStartEventID)

		historyEvent.EventType = shared.EventTypeActivityTaskTimedOut.Ptr()
		historyEvent.ActivityTaskTimedOutEventAttributes = &shared.ActivityTaskTimedOutEventAttributes{
			ScheduledEventId: Int64Ptr(activityScheduleEventID),
			StartedEventId:   Int64Ptr(activityStartEventID),
			TimeoutType:      TimeoutTypePtr(shared.TimeoutTypeScheduleToClose),
		}
	case shared.EventTypeActivityTaskFailed.String():
		if len(h.activityStartEventIDs) == 0 {
			panic("No activity started before fail")
		}
		activityStartEventID := reflect.ValueOf(h.activityStartEventIDs).MapKeys()[0].Int()
		activityScheduleEventID := h.activityStartEventIDs[activityStartEventID]
		delete(h.activityStartEventIDs, activityStartEventID)

		historyEvent.EventType = shared.EventTypeActivityTaskFailed.Ptr()
		historyEvent.ActivityTaskFailedEventAttributes = &shared.ActivityTaskFailedEventAttributes{
			ScheduledEventId: Int64Ptr(activityScheduleEventID),
			StartedEventId:   Int64Ptr(activityStartEventID),
			Identity:         StringPtr(h.identity),
			Reason:           StringPtr(reason),
		}
	case shared.EventTypeActivityTaskCancelRequested.String():
		historyEvent.EventType = shared.EventTypeActivityTaskCancelRequested.Ptr()
		historyEvent.ActivityTaskCancelRequestedEventAttributes = &shared.ActivityTaskCancelRequestedEventAttributes{
			DecisionTaskCompletedEventId: Int64Ptr(h.decisionTaskCompleteEventID),
			ActivityId:                   StringPtr(uuid.New()),
		}
		h.activityCancelRequestEventIDs[eventID] = h.decisionTaskCompleteEventID
	case shared.EventTypeActivityTaskCanceled.String():
		if len(h.activityStartEventIDs) == 0 {
			panic("No activity started before canceled")
		}
		activityStartEventID := reflect.ValueOf(h.activityStartEventIDs).MapKeys()[0].Int()
		activityScheduleEventID := h.activityStartEventIDs[activityStartEventID]
		delete(h.activityStartEventIDs, activityStartEventID)
		delete(h.activityScheduleEventIDs, activityScheduleEventID)
		if len(h.activityCancelRequestEventIDs) == 0 {
			panic("No activity cancel requested before canceled")
		}
		activityCancelRequestEventID := reflect.ValueOf(h.activityCancelRequestEventIDs).MapKeys()[0].Int()
		delete(h.activityCancelRequestEventIDs, activityCancelRequestEventID)

		historyEvent.EventType = shared.EventTypeActivityTaskCanceled.Ptr()
		historyEvent.ActivityTaskCanceledEventAttributes = &shared.ActivityTaskCanceledEventAttributes{
			LatestCancelRequestedEventId: Int64Ptr(activityCancelRequestEventID),
			ScheduledEventId:             Int64Ptr(activityScheduleEventID),
			StartedEventId:               Int64Ptr(activityStartEventID),
			Identity:                     StringPtr(h.identity),
		}
	case shared.EventTypeRequestCancelActivityTaskFailed.String():
		if len(h.activityCancelRequestEventIDs) == 0 {
			panic("No activity cancel requested before failed")
		}
		activityCancelRequestEventID := reflect.ValueOf(h.activityCancelRequestEventIDs).MapKeys()[0].Int()
		completeEventID := h.activityCancelRequestEventIDs[activityCancelRequestEventID]
		delete(h.activityCancelRequestEventIDs, activityCancelRequestEventID)

		historyEvent.EventType = shared.EventTypeRequestCancelActivityTaskFailed.Ptr()
		historyEvent.RequestCancelActivityTaskFailedEventAttributes = &shared.RequestCancelActivityTaskFailedEventAttributes{
			ActivityId:                   StringPtr(uuid.New()),
			DecisionTaskCompletedEventId: Int64Ptr(completeEventID),
		}
	case shared.EventTypeTimerStarted.String():
		historyEvent.EventType = shared.EventTypeTimerStarted.Ptr()
		timerID := uuid.New()
		historyEvent.TimerStartedEventAttributes = &shared.TimerStartedEventAttributes{
			TimerId:                      StringPtr(timerID),
			StartToFireTimeoutSeconds:    Int64Ptr(10),
			DecisionTaskCompletedEventId: Int64Ptr(h.decisionTaskCompleteEventID),
		}
		h.timerStartEventIDs[eventID] = true
		h.timerIDs[timerID] = true
	case shared.EventTypeTimerFired.String():
		historyEvent.EventType = shared.EventTypeTimerFired.Ptr()
		if len(h.timerIDs) == 0 {
			panic("Timer fired before timer started")
		}
		timerID := reflect.ValueOf(h.timerIDs).MapKeys()[0].String()
		delete(h.timerIDs, timerID)

		if len(h.timerStartEventIDs) == 0 {
			panic("Timer fired before timer started")
		}
		timerStartEventID := reflect.ValueOf(h.timerStartEventIDs).MapKeys()[0].Int()
		delete(h.timerStartEventIDs, timerStartEventID)

		historyEvent.TimerFiredEventAttributes = &shared.TimerFiredEventAttributes{
			TimerId:        StringPtr(timerID),
			StartedEventId: Int64Ptr(timerStartEventID),
		}
	case shared.EventTypeTimerCanceled.String():
		historyEvent.EventType = shared.EventTypeTimerCanceled.Ptr()
		if len(h.timerIDs) == 0 {
			panic("Timer fired before timer started")
		}
		timerID := reflect.ValueOf(h.timerIDs).MapKeys()[0].String()
		delete(h.timerIDs, timerID)
		if len(h.timerStartEventIDs) == 0 {
			panic("Timer fired before timer started")
		}
		timerStartEventID := reflect.ValueOf(h.timerStartEventIDs).MapKeys()[0].Int()
		delete(h.timerStartEventIDs, timerStartEventID)

		historyEvent.TimerCanceledEventAttributes = &shared.TimerCanceledEventAttributes{
			TimerId:                      StringPtr(timerID),
			StartedEventId:               Int64Ptr(timerStartEventID),
			DecisionTaskCompletedEventId: Int64Ptr(h.decisionTaskCompleteEventID),
			Identity:                     StringPtr(h.identity),
		}
	case shared.EventTypeStartChildWorkflowExecutionInitiated.String():
		historyEvent.EventType = shared.EventTypeStartChildWorkflowExecutionInitiated.Ptr()
		historyEvent.StartChildWorkflowExecutionInitiatedEventAttributes = &shared.StartChildWorkflowExecutionInitiatedEventAttributes{
			Domain:                              StringPtr(h.domain),
			WorkflowId:                          StringPtr(childWorkflowPrefix + h.workflowID),
			WorkflowType:                        WorkflowTypePtr(shared.WorkflowType{Name: StringPtr(childWorkflowPrefix + h.workflowType)}),
			TaskList:                            TaskListPtr(shared.TaskList{Name: StringPtr(h.taskList), Kind: TaskListKindPtr(shared.TaskListKindNormal)}),
			ExecutionStartToCloseTimeoutSeconds: Int32Ptr(timeout),
			TaskStartToCloseTimeoutSeconds:      Int32Ptr(timeout),
			DecisionTaskCompletedEventId:        Int64Ptr(h.decisionTaskCompleteEventID),
			WorkflowIdReusePolicy:               shared.WorkflowIdReusePolicyRejectDuplicate.Ptr(),
		}
		h.childWorkflowInitialEventIDs[eventID] = h.decisionTaskCompleteEventID
	case shared.EventTypeStartChildWorkflowExecutionFailed.String():
		if len(h.childWorkflowInitialEventIDs) == 0 {
			panic("Child workflow did not initial before failed")
		}
		childWorkflowInitialEventID := reflect.ValueOf(h.childWorkflowInitialEventIDs).MapKeys()[0].Int()
		completeEventID := h.childWorkflowInitialEventIDs[childWorkflowInitialEventID]
		delete(h.childWorkflowInitialEventIDs, childWorkflowInitialEventID)

		historyEvent.EventType = shared.EventTypeStartChildWorkflowExecutionFailed.Ptr()
		historyEvent.StartChildWorkflowExecutionFailedEventAttributes = &shared.StartChildWorkflowExecutionFailedEventAttributes{
			Domain:                       StringPtr(h.domain),
			WorkflowId:                   StringPtr(childWorkflowPrefix + h.workflowID),
			WorkflowType:                 WorkflowTypePtr(shared.WorkflowType{Name: StringPtr(childWorkflowPrefix + h.workflowType)}),
			Cause:                        shared.ChildWorkflowExecutionFailedCauseWorkflowAlreadyRunning.Ptr(),
			InitiatedEventId:             Int64Ptr(childWorkflowInitialEventID),
			DecisionTaskCompletedEventId: Int64Ptr(completeEventID),
		}
	case shared.EventTypeChildWorkflowExecutionStarted.String():
		if len(h.childWorkflowInitialEventIDs) == 0 {
			panic("Child workflow did not initial before start")
		}
		childWorkflowInitialEventID := reflect.ValueOf(h.childWorkflowInitialEventIDs).MapKeys()[0].Int()
		delete(h.childWorkflowInitialEventIDs, childWorkflowInitialEventID)

		runID := uuid.New()
		historyEvent.EventType = shared.EventTypeChildWorkflowExecutionStarted.Ptr()
		historyEvent.ChildWorkflowExecutionStartedEventAttributes = &shared.ChildWorkflowExecutionStartedEventAttributes{
			Domain:           StringPtr(h.domain),
			WorkflowType:     WorkflowTypePtr(shared.WorkflowType{Name: StringPtr(childWorkflowPrefix + h.workflowType)}),
			InitiatedEventId: Int64Ptr(childWorkflowInitialEventID),
			WorkflowExecution: &shared.WorkflowExecution{
				WorkflowId: StringPtr(childWorkflowPrefix + h.workflowID),
				RunId:      StringPtr(runID),
			},
		}
		h.childWorkflowStartEventIDs[eventID] = childWorkflowInitialEventID
		h.childWorkflowRunIDs[eventID] = runID
	case shared.EventTypeChildWorkflowExecutionCompleted.String():
		if len(h.childWorkflowStartEventIDs) == 0 {
			panic("Child workflow did not start before complete")
		}
		childWorkflowStartEventID := reflect.ValueOf(h.childWorkflowStartEventIDs).MapKeys()[0].Int()
		childWorkflowInitialEventID := h.childWorkflowStartEventIDs[childWorkflowStartEventID]
		delete(h.childWorkflowStartEventIDs, childWorkflowStartEventID)

		runID := h.childWorkflowRunIDs[childWorkflowStartEventID]
		if runID == "" {
			panic("child run id is not set with the start event")
		}
		delete(h.childWorkflowRunIDs, childWorkflowStartEventID)

		historyEvent.EventType = shared.EventTypeChildWorkflowExecutionCompleted.Ptr()
		historyEvent.ChildWorkflowExecutionCompletedEventAttributes = &shared.ChildWorkflowExecutionCompletedEventAttributes{
			Domain:           StringPtr(h.domain),
			WorkflowType:     WorkflowTypePtr(shared.WorkflowType{Name: StringPtr(childWorkflowPrefix + h.workflowType)}),
			InitiatedEventId: Int64Ptr(childWorkflowInitialEventID),
			WorkflowExecution: &shared.WorkflowExecution{
				WorkflowId: StringPtr(childWorkflowPrefix + h.workflowID),
				RunId:      StringPtr(runID),
			},
			StartedEventId: Int64Ptr(childWorkflowStartEventID),
		}
	case shared.EventTypeChildWorkflowExecutionTimedOut.String():
		if len(h.childWorkflowStartEventIDs) == 0 {
			panic("Child workflow did not start before timeout")
		}
		childWorkflowStartEventID := reflect.ValueOf(h.childWorkflowStartEventIDs).MapKeys()[0].Int()
		childWorkflowInitialEventID := h.childWorkflowStartEventIDs[childWorkflowStartEventID]
		delete(h.childWorkflowStartEventIDs, childWorkflowStartEventID)

		runID := h.childWorkflowRunIDs[childWorkflowStartEventID]
		if runID == "" {
			panic("child run id is not set with the start event")
		}
		delete(h.childWorkflowRunIDs, childWorkflowStartEventID)

		historyEvent.EventType = shared.EventTypeChildWorkflowExecutionTimedOut.Ptr()
		historyEvent.ChildWorkflowExecutionTimedOutEventAttributes = &shared.ChildWorkflowExecutionTimedOutEventAttributes{
			Domain:           StringPtr(h.domain),
			WorkflowType:     WorkflowTypePtr(shared.WorkflowType{Name: StringPtr(childWorkflowPrefix + h.workflowType)}),
			InitiatedEventId: Int64Ptr(childWorkflowInitialEventID),
			WorkflowExecution: &shared.WorkflowExecution{
				WorkflowId: StringPtr(childWorkflowPrefix + h.workflowID),
				RunId:      StringPtr(runID),
			},
			StartedEventId: Int64Ptr(childWorkflowStartEventID),
			TimeoutType:    TimeoutTypePtr(shared.TimeoutTypeScheduleToClose),
		}
	case shared.EventTypeChildWorkflowExecutionTerminated.String():

		if len(h.childWorkflowStartEventIDs) == 0 {
			panic("Child workflow did not start before terminate ")
		}
		childWorkflowStartEventID := reflect.ValueOf(h.childWorkflowStartEventIDs).MapKeys()[0].Int()
		childWorkflowInitialEventID := h.childWorkflowStartEventIDs[childWorkflowStartEventID]
		delete(h.childWorkflowStartEventIDs, childWorkflowStartEventID)

		runID := h.childWorkflowRunIDs[childWorkflowStartEventID]
		if runID == "" {
			panic("child run id is not set with the start event")
		}
		delete(h.childWorkflowRunIDs, childWorkflowStartEventID)

		historyEvent.EventType = shared.EventTypeChildWorkflowExecutionTerminated.Ptr()
		historyEvent.ChildWorkflowExecutionTerminatedEventAttributes = &shared.ChildWorkflowExecutionTerminatedEventAttributes{
			Domain:           StringPtr(h.domain),
			WorkflowType:     WorkflowTypePtr(shared.WorkflowType{Name: StringPtr(childWorkflowPrefix + h.workflowType)}),
			InitiatedEventId: Int64Ptr(childWorkflowInitialEventID),
			WorkflowExecution: &shared.WorkflowExecution{
				WorkflowId: StringPtr(childWorkflowPrefix + h.workflowID),
				RunId:      StringPtr(runID),
			},
			StartedEventId: Int64Ptr(childWorkflowStartEventID),
		}
	case shared.EventTypeChildWorkflowExecutionFailed.String():
		if len(h.childWorkflowStartEventIDs) == 0 {
			panic("Child workflow did not start before fail")
		}
		childWorkflowStartEventID := reflect.ValueOf(h.childWorkflowStartEventIDs).MapKeys()[0].Int()
		childWorkflowInitialEventID := h.childWorkflowStartEventIDs[childWorkflowStartEventID]
		delete(h.childWorkflowStartEventIDs, childWorkflowStartEventID)

		runID := h.childWorkflowRunIDs[childWorkflowStartEventID]
		if runID == "" {
			panic("child run id is not set with the start event")
		}
		delete(h.childWorkflowRunIDs, childWorkflowStartEventID)

		historyEvent.EventType = shared.EventTypeChildWorkflowExecutionFailed.Ptr()
		historyEvent.ChildWorkflowExecutionFailedEventAttributes = &shared.ChildWorkflowExecutionFailedEventAttributes{
			Domain:           StringPtr(h.domain),
			WorkflowType:     WorkflowTypePtr(shared.WorkflowType{Name: StringPtr(childWorkflowPrefix + h.workflowType)}),
			InitiatedEventId: Int64Ptr(childWorkflowInitialEventID),
			WorkflowExecution: &shared.WorkflowExecution{
				WorkflowId: StringPtr(childWorkflowPrefix + h.workflowID),
				RunId:      StringPtr(runID),
			},
			StartedEventId: Int64Ptr(childWorkflowStartEventID),
		}
	case shared.EventTypeChildWorkflowExecutionCanceled.String():
		if len(h.childWorkflowStartEventIDs) == 0 {
			panic("Child workflow did not start before cancel")
		}
		childWorkflowStartEventID := reflect.ValueOf(h.childWorkflowStartEventIDs).MapKeys()[0].Int()
		childWorkflowInitialEventID := h.childWorkflowStartEventIDs[childWorkflowStartEventID]
		delete(h.childWorkflowStartEventIDs, childWorkflowStartEventID)

		runID := h.childWorkflowRunIDs[childWorkflowStartEventID]
		if runID == "" {
			panic("child run id is not set with the start event")
		}
		delete(h.childWorkflowRunIDs, childWorkflowStartEventID)

		historyEvent.EventType = shared.EventTypeChildWorkflowExecutionCanceled.Ptr()
		historyEvent.ChildWorkflowExecutionCanceledEventAttributes = &shared.ChildWorkflowExecutionCanceledEventAttributes{
			Domain:           StringPtr(h.domain),
			WorkflowType:     WorkflowTypePtr(shared.WorkflowType{Name: StringPtr(childWorkflowPrefix + h.workflowType)}),
			InitiatedEventId: Int64Ptr(childWorkflowInitialEventID),
			WorkflowExecution: &shared.WorkflowExecution{
				WorkflowId: StringPtr(childWorkflowPrefix + h.workflowID),
				RunId:      StringPtr(runID),
			},
			StartedEventId: Int64Ptr(childWorkflowStartEventID),
		}
	case shared.EventTypeSignalExternalWorkflowExecutionInitiated.String():
		historyEvent.EventType = shared.EventTypeSignalExternalWorkflowExecutionInitiated.Ptr()
		historyEvent.SignalExternalWorkflowExecutionInitiatedEventAttributes = &shared.SignalExternalWorkflowExecutionInitiatedEventAttributes{
			DecisionTaskCompletedEventId: Int64Ptr(h.decisionTaskCompleteEventID),
			Domain:                       StringPtr(h.domain),
			WorkflowExecution: &shared.WorkflowExecution{
				WorkflowId: StringPtr(externalWorkflowPrefix + h.workflowID),
				RunId:      StringPtr(uuid.New()),
			},
			SignalName:        StringPtr("signal"),
			ChildWorkflowOnly: BoolPtr(false),
		}
		h.signalExternalWorkflowEventIDs[eventID] = h.decisionTaskCompleteEventID
	case shared.EventTypeSignalExternalWorkflowExecutionFailed.String():
		if len(h.signalExternalWorkflowEventIDs) == 0 {
			panic("No external workflow signaled")
		}
		signalExternalWorkflowEventID := reflect.ValueOf(h.signalExternalWorkflowEventIDs).MapKeys()[0].Int()
		completeEventID := h.signalExternalWorkflowEventIDs[signalExternalWorkflowEventID]
		delete(h.signalExternalWorkflowEventIDs, signalExternalWorkflowEventID)

		historyEvent.EventType = shared.EventTypeSignalExternalWorkflowExecutionFailed.Ptr()
		historyEvent.SignalExternalWorkflowExecutionFailedEventAttributes = &shared.SignalExternalWorkflowExecutionFailedEventAttributes{
			Cause:                        SignalExternalWorkflowExecutionFailedCausePtr(shared.SignalExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution),
			DecisionTaskCompletedEventId: Int64Ptr(completeEventID),
			Domain:                       StringPtr(h.domain),
			WorkflowExecution: &shared.WorkflowExecution{
				WorkflowId: StringPtr(externalWorkflowPrefix + h.workflowID),
				RunId:      StringPtr(uuid.New()),
			},
			InitiatedEventId: Int64Ptr(signalExternalWorkflowEventID),
		}
	case shared.EventTypeExternalWorkflowExecutionSignaled.String():
		if len(h.signalExternalWorkflowEventIDs) == 0 {
			panic("No external workflow signaled")
		}
		signalExternalWorkflowEventID := reflect.ValueOf(h.signalExternalWorkflowEventIDs).MapKeys()[0].Int()
		delete(h.signalExternalWorkflowEventIDs, signalExternalWorkflowEventID)

		historyEvent.EventType = shared.EventTypeExternalWorkflowExecutionSignaled.Ptr()
		historyEvent.ExternalWorkflowExecutionSignaledEventAttributes = &shared.ExternalWorkflowExecutionSignaledEventAttributes{
			InitiatedEventId: Int64Ptr(signalExternalWorkflowEventID),
			Domain:           StringPtr(h.domain),
			WorkflowExecution: &shared.WorkflowExecution{
				WorkflowId: StringPtr(externalWorkflowPrefix + h.workflowID),
				RunId:      StringPtr(uuid.New()),
			},
		}
	case shared.EventTypeRequestCancelExternalWorkflowExecutionInitiated.String():
		historyEvent.EventType = shared.EventTypeRequestCancelExternalWorkflowExecutionInitiated.Ptr()
		historyEvent.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes = &shared.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{
			DecisionTaskCompletedEventId: Int64Ptr(h.decisionTaskCompleteEventID),
			Domain:                       StringPtr(h.domain),
			WorkflowExecution: &shared.WorkflowExecution{
				WorkflowId: StringPtr(externalWorkflowPrefix + h.workflowID),
				RunId:      StringPtr(uuid.New()),
			},
			ChildWorkflowOnly: BoolPtr(false),
		}
		h.requestExternalWorkflowCanceledEventIDs[eventID] = h.decisionTaskCompleteEventID
	case shared.EventTypeRequestCancelExternalWorkflowExecutionFailed.String():
		if len(h.requestExternalWorkflowCanceledEventIDs) == 0 {
			panic("No cancel request external workflow")
		}
		requestExternalWorkflowCanceledEventID := reflect.ValueOf(h.requestExternalWorkflowCanceledEventIDs).MapKeys()[0].Int()
		completeEventID := h.requestExternalWorkflowCanceledEventIDs[requestExternalWorkflowCanceledEventID]
		delete(h.requestExternalWorkflowCanceledEventIDs, requestExternalWorkflowCanceledEventID)

		historyEvent.EventType = shared.EventTypeRequestCancelExternalWorkflowExecutionFailed.Ptr()
		historyEvent.RequestCancelExternalWorkflowExecutionFailedEventAttributes = &shared.RequestCancelExternalWorkflowExecutionFailedEventAttributes{
			Cause:                        CancelExternalWorkflowExecutionFailedCausePtr(shared.CancelExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution),
			DecisionTaskCompletedEventId: Int64Ptr(completeEventID),
			Domain:                       StringPtr(h.domain),
			WorkflowExecution: &shared.WorkflowExecution{
				WorkflowId: StringPtr(externalWorkflowPrefix + h.workflowID),
				RunId:      StringPtr(uuid.New()),
			},
			InitiatedEventId: Int64Ptr(requestExternalWorkflowCanceledEventID),
		}
	case shared.EventTypeExternalWorkflowExecutionCancelRequested.String():
		if len(h.requestExternalWorkflowCanceledEventIDs) == 0 {
			panic("No cancel request external workflow")
		}
		requestExternalWorkflowCanceledEventID := reflect.ValueOf(h.requestExternalWorkflowCanceledEventIDs).MapKeys()[0].Int()
		delete(h.requestExternalWorkflowCanceledEventIDs, requestExternalWorkflowCanceledEventID)

		historyEvent.EventType = shared.EventTypeExternalWorkflowExecutionCancelRequested.Ptr()
		historyEvent.ExternalWorkflowExecutionCancelRequestedEventAttributes = &shared.ExternalWorkflowExecutionCancelRequestedEventAttributes{
			InitiatedEventId: Int64Ptr(requestExternalWorkflowCanceledEventID),
			Domain:           StringPtr(h.domain),
			WorkflowExecution: &shared.WorkflowExecution{
				WorkflowId: StringPtr(externalWorkflowPrefix + h.workflowID),
				RunId:      StringPtr(uuid.New()),
			},
		}
	}
	return historyEvent
}

func getDefaultHistoryEvent(eventID, version int64) *shared.HistoryEvent {
	return &shared.HistoryEvent{
		EventId:   Int64Ptr(eventID),
		Timestamp: Int64Ptr(time.Now().Unix()),
		TaskId:    Int64Ptr(EmptyEventTaskID),
		Version:   Int64Ptr(version),
	}
}

// InitializeHistoryEventGenerator initializes the history event generator
func InitializeHistoryEventGenerator() Generator {
	generator := NewEventGenerator()

	//Functions
	notPendingDecisionTask := func() bool {
		count := 0
		for _, e := range generator.ListGeneratedVertices() {
			switch e.GetName() {
			case shared.EventTypeDecisionTaskScheduled.String():
				count++
			case shared.EventTypeDecisionTaskCompleted.String(),
				shared.EventTypeDecisionTaskFailed.String(),
				shared.EventTypeDecisionTaskTimedOut.String():
				count--
			}
		}
		return count <= 0
	}
	containActivityComplete := func() bool {
		for _, e := range generator.ListGeneratedVertices() {
			if e.GetName() == shared.EventTypeActivityTaskCompleted.String() {
				return true
			}
		}
		return false
	}
	hasPendingTimer := func() bool {
		count := 0
		for _, e := range generator.ListGeneratedVertices() {
			switch e.GetName() {
			case shared.EventTypeTimerStarted.String():
				count++
			case shared.EventTypeTimerFired.String(),
				shared.EventTypeTimerCanceled.String():
				count--
			}
		}
		return count > 0
	}
	hasPendingActivity := func() bool {
		count := 0
		for _, e := range generator.ListGeneratedVertices() {
			switch e.GetName() {
			case shared.EventTypeActivityTaskScheduled.String():
				count++
			case shared.EventTypeActivityTaskCanceled.String(),
				shared.EventTypeActivityTaskFailed.String(),
				shared.EventTypeActivityTaskTimedOut.String(),
				shared.EventTypeActivityTaskCompleted.String():
				count--
			}
		}
		return count > 0
	}
	hasPendingStartActivity := func() bool {
		count := 0
		for _, e := range generator.ListGeneratedVertices() {
			switch e.GetName() {
			case shared.EventTypeActivityTaskStarted.String():
				count++
			case shared.EventTypeActivityTaskCanceled.String(),
				shared.EventTypeActivityTaskFailed.String(),
				shared.EventTypeActivityTaskTimedOut.String(),
				shared.EventTypeActivityTaskCompleted.String():
				count--
			}
		}
		return count > 0
	}
	canDoBatch := func(history []Vertex) bool {
		if len(history) == 0 {
			return true
		}

		hasPendingDecisionTask := false
		for _, event := range generator.ListGeneratedVertices() {
			switch event.GetName() {
			case shared.EventTypeDecisionTaskScheduled.String():
				hasPendingDecisionTask = true
			case shared.EventTypeDecisionTaskCompleted.String(),
				shared.EventTypeDecisionTaskFailed.String(),
				shared.EventTypeDecisionTaskTimedOut.String():
				hasPendingDecisionTask = false
			}
		}
		if hasPendingDecisionTask {
			return false
		}
		if history[len(history)-1].GetName() == shared.EventTypeDecisionTaskScheduled.String() {
			return false
		}
		if history[0].GetName() == shared.EventTypeDecisionTaskCompleted.String() {
			return len(history) == 1
		}
		return true
	}

	//Setup decision task model
	decisionModel := NewHistoryEventModel()
	decisionSchedule := NewHistoryEventVertex(shared.EventTypeDecisionTaskScheduled.String())
	decisionStart := NewHistoryEventVertex(shared.EventTypeDecisionTaskStarted.String())
	decisionStart.SetIsStrictOnNextVertex(true)
	decisionFail := NewHistoryEventVertex(shared.EventTypeDecisionTaskFailed.String())
	decisionTimedOut := NewHistoryEventVertex(shared.EventTypeDecisionTaskTimedOut.String())
	decisionComplete := NewHistoryEventVertex(shared.EventTypeDecisionTaskCompleted.String())
	decisionComplete.SetIsStrictOnNextVertex(true)
	decisionComplete.SetMaxNextVertex(2)
	decisionScheduleToStart := NewHistoryEventEdge(decisionSchedule, decisionStart)
	decisionStartToComplete := NewHistoryEventEdge(decisionStart, decisionComplete)
	decisionStartToFail := NewHistoryEventEdge(decisionStart, decisionFail)
	decisionStartToTimedOut := NewHistoryEventEdge(decisionStart, decisionTimedOut)
	decisionFailToSchedule := NewHistoryEventEdge(decisionFail, decisionSchedule)
	decisionFailToSchedule.SetCondition(notPendingDecisionTask)
	decisionTimedOutToSchedule := NewHistoryEventEdge(decisionTimedOut, decisionSchedule)
	decisionTimedOutToSchedule.SetCondition(notPendingDecisionTask)
	decisionModel.AddEdge(decisionScheduleToStart, decisionStartToComplete, decisionStartToFail, decisionStartToTimedOut,
		decisionFailToSchedule, decisionTimedOutToSchedule)

	//Setup workflow model
	workflowModel := NewHistoryEventModel()
	workflowStart := NewHistoryEventVertex(shared.EventTypeWorkflowExecutionStarted.String())
	workflowSignal := NewHistoryEventVertex(shared.EventTypeWorkflowExecutionSignaled.String())
	workflowComplete := NewHistoryEventVertex(shared.EventTypeWorkflowExecutionCompleted.String())
	continueAsNew := NewHistoryEventVertex(shared.EventTypeWorkflowExecutionContinuedAsNew.String())
	workflowFail := NewHistoryEventVertex(shared.EventTypeWorkflowExecutionFailed.String())
	workflowCancel := NewHistoryEventVertex(shared.EventTypeWorkflowExecutionCanceled.String())
	workflowCancelRequest := NewHistoryEventVertex(shared.EventTypeWorkflowExecutionCancelRequested.String()) //?
	workflowTerminate := NewHistoryEventVertex(shared.EventTypeWorkflowExecutionTerminated.String())
	workflowTimedOut := NewHistoryEventVertex(shared.EventTypeWorkflowExecutionTimedOut.String())
	workflowStartToSignal := NewHistoryEventEdge(workflowStart, workflowSignal)
	workflowStartToDecisionSchedule := NewHistoryEventEdge(workflowStart, decisionSchedule)
	workflowStartToDecisionSchedule.SetCondition(notPendingDecisionTask)
	workflowSignalToDecisionSchedule := NewHistoryEventEdge(workflowSignal, decisionSchedule)
	workflowSignalToDecisionSchedule.SetCondition(notPendingDecisionTask)
	decisionCompleteToWorkflowComplete := NewHistoryEventEdge(decisionComplete, workflowComplete)
	decisionCompleteToWorkflowComplete.SetCondition(containActivityComplete)
	decisionCompleteToWorkflowFailed := NewHistoryEventEdge(decisionComplete, workflowFail)
	decisionCompleteToWorkflowFailed.SetCondition(containActivityComplete)
	decisionCompleteToCAN := NewHistoryEventEdge(decisionComplete, continueAsNew)
	decisionCompleteToCAN.SetCondition(containActivityComplete)
	workflowCancelRequestToCancel := NewHistoryEventEdge(workflowCancelRequest, workflowCancel)
	workflowModel.AddEdge(workflowStartToSignal, workflowStartToDecisionSchedule, workflowSignalToDecisionSchedule,
		decisionCompleteToCAN, decisionCompleteToWorkflowComplete, decisionCompleteToWorkflowFailed, workflowCancelRequestToCancel)

	//Setup activity model
	activityModel := NewHistoryEventModel()
	activitySchedule := NewHistoryEventVertex(shared.EventTypeActivityTaskScheduled.String())
	activityStart := NewHistoryEventVertex(shared.EventTypeActivityTaskStarted.String())
	activityComplete := NewHistoryEventVertex(shared.EventTypeActivityTaskCompleted.String())
	activityFail := NewHistoryEventVertex(shared.EventTypeActivityTaskFailed.String())
	activityTimedOut := NewHistoryEventVertex(shared.EventTypeActivityTaskTimedOut.String())
	activityCancelRequest := NewHistoryEventVertex(shared.EventTypeActivityTaskCancelRequested.String()) //?
	activityCancel := NewHistoryEventVertex(shared.EventTypeActivityTaskCanceled.String())
	activityCancelRequestFail := NewHistoryEventVertex(shared.EventTypeRequestCancelActivityTaskFailed.String())
	decisionCompleteToATSchedule := NewHistoryEventEdge(decisionComplete, activitySchedule)

	activityScheduleToStart := NewHistoryEventEdge(activitySchedule, activityStart)
	activityScheduleToStart.SetCondition(hasPendingActivity)

	activityStartToComplete := NewHistoryEventEdge(activityStart, activityComplete)
	activityStartToComplete.SetCondition(hasPendingActivity)

	activityStartToFail := NewHistoryEventEdge(activityStart, activityFail)
	activityStartToFail.SetCondition(hasPendingActivity)

	activityStartToTimedOut := NewHistoryEventEdge(activityStart, activityTimedOut)
	activityStartToTimedOut.SetCondition(hasPendingActivity)

	activityCompleteToDecisionSchedule := NewHistoryEventEdge(activityComplete, decisionSchedule)
	activityCompleteToDecisionSchedule.SetCondition(notPendingDecisionTask)
	activityFailToDecisionSchedule := NewHistoryEventEdge(activityFail, decisionSchedule)
	activityFailToDecisionSchedule.SetCondition(notPendingDecisionTask)
	activityTimedOutToDecisionSchedule := NewHistoryEventEdge(activityTimedOut, decisionSchedule)
	activityTimedOutToDecisionSchedule.SetCondition(notPendingDecisionTask)
	activityCancelToDecisionSchedule := NewHistoryEventEdge(activityCancel, decisionSchedule)
	activityCancelToDecisionSchedule.SetCondition(notPendingDecisionTask)

	activityCancelReqToCancel := NewHistoryEventEdge(activityCancelRequest, activityCancel)
	activityCancelReqToCancel.SetCondition(hasPendingActivity)

	activityCancelReqToCancelFail := NewHistoryEventEdge(activityCancelRequest, activityCancelRequestFail)
	activityCancelRequestFailToDecisionSchedule := NewHistoryEventEdge(activityCancelRequestFail, decisionSchedule)
	activityCancelRequestFailToDecisionSchedule.SetCondition(notPendingDecisionTask)
	decisionCompleteToActivityCancelRequest := NewHistoryEventEdge(decisionComplete, activityCancelRequest)
	decisionCompleteToActivityCancelRequest.SetCondition(hasPendingStartActivity)

	activityModel.AddEdge(decisionCompleteToATSchedule, activityScheduleToStart, activityStartToComplete,
		activityStartToFail, activityStartToTimedOut, decisionCompleteToATSchedule, activityCompleteToDecisionSchedule,
		activityFailToDecisionSchedule, activityTimedOutToDecisionSchedule, activityCancelReqToCancel, activityCancelReqToCancelFail,
		activityCancelToDecisionSchedule, decisionCompleteToActivityCancelRequest, activityCancelRequestFailToDecisionSchedule)

	//Setup timer model
	timerModel := NewHistoryEventModel()
	timerStart := NewHistoryEventVertex(shared.EventTypeTimerStarted.String())
	timerFired := NewHistoryEventVertex(shared.EventTypeTimerFired.String())
	timerCancel := NewHistoryEventVertex(shared.EventTypeTimerCanceled.String())
	timerStartToFire := NewHistoryEventEdge(timerStart, timerFired)
	timerStartToFire.SetCondition(hasPendingTimer)
	decisionCompleteToCancel := NewHistoryEventEdge(decisionComplete, timerCancel)
	decisionCompleteToCancel.SetCondition(hasPendingTimer)

	decisionCompleteToTimerStart := NewHistoryEventEdge(decisionComplete, timerStart)
	timerFiredToDecisionSchedule := NewHistoryEventEdge(timerFired, decisionSchedule)
	timerFiredToDecisionSchedule.SetCondition(notPendingDecisionTask)
	timerCancelToDecisionSchedule := NewHistoryEventEdge(timerCancel, decisionSchedule)
	timerCancelToDecisionSchedule.SetCondition(notPendingDecisionTask)
	timerModel.AddEdge(timerStartToFire, decisionCompleteToCancel, decisionCompleteToTimerStart, timerFiredToDecisionSchedule, timerCancelToDecisionSchedule)

	//Setup child workflow model
	childWorkflowModel := NewHistoryEventModel()
	childWorkflowInitial := NewHistoryEventVertex(shared.EventTypeStartChildWorkflowExecutionInitiated.String())
	childWorkflowInitialFail := NewHistoryEventVertex(shared.EventTypeStartChildWorkflowExecutionFailed.String())
	childWorkflowStart := NewHistoryEventVertex(shared.EventTypeChildWorkflowExecutionStarted.String())
	childWorkflowCancel := NewHistoryEventVertex(shared.EventTypeChildWorkflowExecutionCanceled.String())
	childWorkflowComplete := NewHistoryEventVertex(shared.EventTypeChildWorkflowExecutionCompleted.String())
	childWorkflowFail := NewHistoryEventVertex(shared.EventTypeChildWorkflowExecutionFailed.String())
	childWorkflowTerminate := NewHistoryEventVertex(shared.EventTypeChildWorkflowExecutionTerminated.String())
	childWorkflowTimedOut := NewHistoryEventVertex(shared.EventTypeChildWorkflowExecutionTimedOut.String())
	decisionCompleteToChildWorkflowInitial := NewHistoryEventEdge(decisionComplete, childWorkflowInitial)
	childWorkflowInitialToFail := NewHistoryEventEdge(childWorkflowInitial, childWorkflowInitialFail)
	childWorkflowInitialToStart := NewHistoryEventEdge(childWorkflowInitial, childWorkflowStart)
	childWorkflowStartToCancel := NewHistoryEventEdge(childWorkflowStart, childWorkflowCancel)
	childWorkflowStartToFail := NewHistoryEventEdge(childWorkflowStart, childWorkflowFail)
	childWorkflowStartToComplete := NewHistoryEventEdge(childWorkflowStart, childWorkflowComplete)
	childWorkflowStartToTerminate := NewHistoryEventEdge(childWorkflowStart, childWorkflowTerminate)
	childWorkflowStartToTimedOut := NewHistoryEventEdge(childWorkflowStart, childWorkflowTimedOut)
	childWorkflowCancelToDecisionSchedule := NewHistoryEventEdge(childWorkflowCancel, decisionSchedule)
	childWorkflowCancelToDecisionSchedule.SetCondition(notPendingDecisionTask)
	childWorkflowFailToDecisionSchedule := NewHistoryEventEdge(childWorkflowFail, decisionSchedule)
	childWorkflowFailToDecisionSchedule.SetCondition(notPendingDecisionTask)
	childWorkflowCompleteToDecisionSchedule := NewHistoryEventEdge(childWorkflowComplete, decisionSchedule)
	childWorkflowCompleteToDecisionSchedule.SetCondition(notPendingDecisionTask)
	childWorkflowTerminateToDecisionSchedule := NewHistoryEventEdge(childWorkflowTerminate, decisionSchedule)
	childWorkflowTerminateToDecisionSchedule.SetCondition(notPendingDecisionTask)
	childWorkflowTimedOutToDecisionSchedule := NewHistoryEventEdge(childWorkflowTimedOut, decisionSchedule)
	childWorkflowTimedOutToDecisionSchedule.SetCondition(notPendingDecisionTask)
	childWorkflowInitialFailToDecisionSchedule := NewHistoryEventEdge(childWorkflowInitialFail, decisionSchedule)
	childWorkflowInitialFailToDecisionSchedule.SetCondition(notPendingDecisionTask)
	childWorkflowModel.AddEdge(decisionCompleteToChildWorkflowInitial, childWorkflowInitialToFail, childWorkflowInitialToStart,
		childWorkflowStartToCancel, childWorkflowStartToFail, childWorkflowStartToComplete, childWorkflowStartToTerminate,
		childWorkflowStartToTimedOut, childWorkflowCancelToDecisionSchedule, childWorkflowFailToDecisionSchedule,
		childWorkflowCompleteToDecisionSchedule, childWorkflowTerminateToDecisionSchedule, childWorkflowTimedOutToDecisionSchedule,
		childWorkflowInitialFailToDecisionSchedule)

	//Setup external workflow model
	externalWorkflowModel := NewHistoryEventModel()
	externalWorkflowSignal := NewHistoryEventVertex(shared.EventTypeSignalExternalWorkflowExecutionInitiated.String())
	externalWorkflowSignalFailed := NewHistoryEventVertex(shared.EventTypeSignalExternalWorkflowExecutionFailed.String())
	externalWorkflowSignaled := NewHistoryEventVertex(shared.EventTypeExternalWorkflowExecutionSignaled.String())
	externalWorkflowCancel := NewHistoryEventVertex(shared.EventTypeRequestCancelExternalWorkflowExecutionInitiated.String())
	externalWorkflowCancelFail := NewHistoryEventVertex(shared.EventTypeRequestCancelExternalWorkflowExecutionFailed.String())
	externalWorkflowCanceled := NewHistoryEventVertex(shared.EventTypeExternalWorkflowExecutionCancelRequested.String())
	decisionCompleteToExternalWorkflowSignal := NewHistoryEventEdge(decisionComplete, externalWorkflowSignal)
	decisionCompleteToExternalWorkflowCancel := NewHistoryEventEdge(decisionComplete, externalWorkflowCancel)
	externalWorkflowSignalToFail := NewHistoryEventEdge(externalWorkflowSignal, externalWorkflowSignalFailed)
	externalWorkflowSignalToSignaled := NewHistoryEventEdge(externalWorkflowSignal, externalWorkflowSignaled)
	externalWorkflowCancelToFail := NewHistoryEventEdge(externalWorkflowCancel, externalWorkflowCancelFail)
	externalWorkflowCancelToCanceled := NewHistoryEventEdge(externalWorkflowCancel, externalWorkflowCanceled)
	externalWorkflowSignaledToDecisionSchedule := NewHistoryEventEdge(externalWorkflowSignaled, decisionSchedule)
	externalWorkflowSignaledToDecisionSchedule.SetCondition(notPendingDecisionTask)
	externalWorkflowSignalFailedToDecisionSchedule := NewHistoryEventEdge(externalWorkflowSignalFailed, decisionSchedule)
	externalWorkflowSignalFailedToDecisionSchedule.SetCondition(notPendingDecisionTask)
	externalWorkflowCanceledToDecisionSchedule := NewHistoryEventEdge(externalWorkflowCanceled, decisionSchedule)
	externalWorkflowCanceledToDecisionSchedule.SetCondition(notPendingDecisionTask)
	externalWorkflowCancelFailToDecisionSchedule := NewHistoryEventEdge(externalWorkflowCancelFail, decisionSchedule)
	externalWorkflowCancelFailToDecisionSchedule.SetCondition(notPendingDecisionTask)
	externalWorkflowModel.AddEdge(decisionCompleteToExternalWorkflowSignal, decisionCompleteToExternalWorkflowCancel,
		externalWorkflowSignalToFail, externalWorkflowSignalToSignaled, externalWorkflowCancelToFail, externalWorkflowCancelToCanceled,
		externalWorkflowSignaledToDecisionSchedule, externalWorkflowSignalFailedToDecisionSchedule,
		externalWorkflowCanceledToDecisionSchedule, externalWorkflowCancelFailToDecisionSchedule)

	//Config event generator
	generator.SetBatchGenerationRule(canDoBatch)
	generator.AddInitialEntryVertex(workflowStart)
	generator.AddExitVertex(workflowComplete, workflowFail, workflowTerminate, workflowTimedOut, continueAsNew)
	//generator.AddRandomEntryVertex(workflowSignal, workflowTerminate, workflowTimedOut)
	generator.AddModel(decisionModel)
	generator.AddModel(workflowModel)
	generator.AddModel(activityModel)
	generator.AddModel(timerModel)
	generator.AddModel(childWorkflowModel)
	generator.AddModel(externalWorkflowModel)
	return generator
}
