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

package testing

import (
	"time"

	"github.com/pborman/uuid"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
)

const (
	timeout              = int32(10000)
	signal               = "NDC signal"
	checksum             = "NDC checksum"
	childWorkflowPrefix  = "child-"
	reason               = "NDC reason"
	workflowType         = "test-workflow-type"
	taskList             = "taskList"
	identity             = "identity"
	decisionTaskAttempts = 0
	childWorkflowID      = "child-WorkflowID"
	externalWorkflowID   = "external-WorkflowID"
)

var (
	globalTaskID int64 = 1
)

// InitializeHistoryEventGenerator initializes the history event generator
func InitializeHistoryEventGenerator(
	domain string,
	defaultVersion int64,
) Generator {

	generator := NewEventGenerator(time.Now().UnixNano())
	generator.SetVersion(defaultVersion)
	// Functions
	notPendingDecisionTask := func(input ...interface{}) bool {
		count := 0
		history := input[0].([]Vertex)
		for _, e := range history {
			switch e.GetName() {
			case types.EventTypeDecisionTaskScheduled.String():
				count++
			case types.EventTypeDecisionTaskCompleted.String(),
				types.EventTypeDecisionTaskFailed.String(),
				types.EventTypeDecisionTaskTimedOut.String():
				count--
			}
		}
		return count <= 0
	}
	containActivityComplete := func(input ...interface{}) bool {
		history := input[0].([]Vertex)
		for _, e := range history {
			if e.GetName() == types.EventTypeActivityTaskCompleted.String() {
				return true
			}
		}
		return false
	}
	hasPendingActivity := func(input ...interface{}) bool {
		count := 0
		history := input[0].([]Vertex)
		for _, e := range history {
			switch e.GetName() {
			case types.EventTypeActivityTaskScheduled.String():
				count++
			case types.EventTypeActivityTaskCanceled.String(),
				types.EventTypeActivityTaskFailed.String(),
				types.EventTypeActivityTaskTimedOut.String(),
				types.EventTypeActivityTaskCompleted.String():
				count--
			}
		}
		return count > 0
	}
	canDoBatch := func(currentBatch []Vertex, history []Vertex) bool {
		if len(currentBatch) == 0 {
			return true
		}

		hasPendingDecisionTask := false
		for _, event := range history {
			switch event.GetName() {
			case types.EventTypeDecisionTaskScheduled.String():
				hasPendingDecisionTask = true
			case types.EventTypeDecisionTaskCompleted.String(),
				types.EventTypeDecisionTaskFailed.String(),
				types.EventTypeDecisionTaskTimedOut.String():
				hasPendingDecisionTask = false
			}
		}
		if hasPendingDecisionTask {
			return false
		}
		if currentBatch[len(currentBatch)-1].GetName() == types.EventTypeDecisionTaskScheduled.String() {
			return false
		}
		if currentBatch[0].GetName() == types.EventTypeDecisionTaskCompleted.String() {
			return len(currentBatch) == 1
		}
		return true
	}

	// Setup decision task model
	decisionModel := NewHistoryEventModel()
	decisionSchedule := NewHistoryEventVertex(types.EventTypeDecisionTaskScheduled.String())
	decisionSchedule.SetDataFunc(func(input ...interface{}) interface{} {
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeDecisionTaskScheduled.Ptr()
		historyEvent.DecisionTaskScheduledEventAttributes = &types.DecisionTaskScheduledEventAttributes{
			TaskList: &types.TaskList{
				Name: taskList,
				Kind: types.TaskListKindNormal.Ptr(),
			},
			StartToCloseTimeoutSeconds: common.Int32Ptr(timeout),
			Attempt:                    decisionTaskAttempts,
		}
		return historyEvent
	})
	decisionStart := NewHistoryEventVertex(types.EventTypeDecisionTaskStarted.String())
	decisionStart.SetIsStrictOnNextVertex(true)
	decisionStart.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeDecisionTaskStarted.Ptr()
		historyEvent.DecisionTaskStartedEventAttributes = &types.DecisionTaskStartedEventAttributes{
			ScheduledEventID: lastEvent.EventID,
			Identity:         identity,
			RequestID:        uuid.New(),
		}
		return historyEvent
	})
	decisionFail := NewHistoryEventVertex(types.EventTypeDecisionTaskFailed.String())
	decisionFail.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeDecisionTaskFailed.Ptr()
		historyEvent.DecisionTaskFailedEventAttributes = &types.DecisionTaskFailedEventAttributes{
			ScheduledEventID: lastEvent.GetDecisionTaskStartedEventAttributes().ScheduledEventID,
			StartedEventID:   lastEvent.EventID,
			Cause:            types.DecisionTaskFailedCauseUnhandledDecision.Ptr(),
			Identity:         identity,
			ForkEventVersion: version,
		}
		return historyEvent
	})
	decisionTimedOut := NewHistoryEventVertex(types.EventTypeDecisionTaskTimedOut.String())
	decisionTimedOut.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeDecisionTaskTimedOut.Ptr()
		historyEvent.DecisionTaskTimedOutEventAttributes = &types.DecisionTaskTimedOutEventAttributes{
			ScheduledEventID: lastEvent.GetDecisionTaskStartedEventAttributes().ScheduledEventID,
			StartedEventID:   lastEvent.EventID,
			TimeoutType:      types.TimeoutTypeScheduleToStart.Ptr(),
		}
		return historyEvent
	})
	decisionComplete := NewHistoryEventVertex(types.EventTypeDecisionTaskCompleted.String())
	decisionComplete.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeDecisionTaskCompleted.Ptr()
		historyEvent.DecisionTaskCompletedEventAttributes = &types.DecisionTaskCompletedEventAttributes{
			ScheduledEventID: lastEvent.GetDecisionTaskStartedEventAttributes().ScheduledEventID,
			StartedEventID:   lastEvent.EventID,
			Identity:         identity,
			BinaryChecksum:   checksum,
		}
		return historyEvent
	})
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

	// Setup workflow model
	workflowModel := NewHistoryEventModel()

	workflowStart := NewHistoryEventVertex(types.EventTypeWorkflowExecutionStarted.String())
	workflowStart.SetDataFunc(func(input ...interface{}) interface{} {
		historyEvent := getDefaultHistoryEvent(1, defaultVersion)
		historyEvent.EventType = types.EventTypeWorkflowExecutionStarted.Ptr()
		historyEvent.WorkflowExecutionStartedEventAttributes = &types.WorkflowExecutionStartedEventAttributes{
			WorkflowType: &types.WorkflowType{
				Name: workflowType,
			},
			TaskList: &types.TaskList{
				Name: taskList,
				Kind: types.TaskListKindNormal.Ptr(),
			},
			ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(timeout),
			TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(timeout),
			Identity:                            identity,
			FirstExecutionRunID:                 uuid.New(),
		}
		return historyEvent
	})
	workflowSignal := NewHistoryEventVertex(types.EventTypeWorkflowExecutionSignaled.String())
	workflowSignal.SetDataFunc(func(input ...interface{}) interface{} {
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeWorkflowExecutionSignaled.Ptr()
		historyEvent.WorkflowExecutionSignaledEventAttributes = &types.WorkflowExecutionSignaledEventAttributes{
			SignalName: signal,
			Identity:   identity,
		}
		return historyEvent
	})
	workflowComplete := NewHistoryEventVertex(types.EventTypeWorkflowExecutionCompleted.String())
	workflowComplete.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		EventID := lastEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeWorkflowExecutionCompleted.Ptr()
		historyEvent.WorkflowExecutionCompletedEventAttributes = &types.WorkflowExecutionCompletedEventAttributes{
			DecisionTaskCompletedEventID: lastEvent.EventID,
		}
		return historyEvent
	})
	continueAsNew := NewHistoryEventVertex(types.EventTypeWorkflowExecutionContinuedAsNew.String())
	continueAsNew.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		EventID := lastEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeWorkflowExecutionContinuedAsNew.Ptr()
		historyEvent.WorkflowExecutionContinuedAsNewEventAttributes = &types.WorkflowExecutionContinuedAsNewEventAttributes{
			NewExecutionRunID: uuid.New(),
			WorkflowType: &types.WorkflowType{
				Name: workflowType,
			},
			TaskList: &types.TaskList{
				Name: taskList,
				Kind: types.TaskListKindNormal.Ptr(),
			},
			ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(timeout),
			TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(timeout),
			DecisionTaskCompletedEventID:        EventID - 1,
			Initiator:                           types.ContinueAsNewInitiatorDecider.Ptr(),
		}
		return historyEvent
	})
	workflowFail := NewHistoryEventVertex(types.EventTypeWorkflowExecutionFailed.String())
	workflowFail.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		EventID := lastEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeWorkflowExecutionFailed.Ptr()
		historyEvent.WorkflowExecutionFailedEventAttributes = &types.WorkflowExecutionFailedEventAttributes{
			DecisionTaskCompletedEventID: lastEvent.EventID,
		}
		return historyEvent
	})
	workflowCancel := NewHistoryEventVertex(types.EventTypeWorkflowExecutionCanceled.String())
	workflowCancel.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeWorkflowExecutionCanceled.Ptr()
		historyEvent.WorkflowExecutionCanceledEventAttributes = &types.WorkflowExecutionCanceledEventAttributes{
			DecisionTaskCompletedEventID: lastEvent.EventID,
		}
		return historyEvent
	})
	workflowCancelRequest := NewHistoryEventVertex(types.EventTypeWorkflowExecutionCancelRequested.String())
	workflowCancelRequest.SetDataFunc(func(input ...interface{}) interface{} {
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeWorkflowExecutionCancelRequested.Ptr()
		historyEvent.WorkflowExecutionCancelRequestedEventAttributes = &types.WorkflowExecutionCancelRequestedEventAttributes{
			Cause:                    "",
			ExternalInitiatedEventID: common.Int64Ptr(1),
			ExternalWorkflowExecution: &types.WorkflowExecution{
				WorkflowID: externalWorkflowID,
				RunID:      uuid.New(),
			},
			Identity: identity,
		}
		return historyEvent
	})
	workflowTerminate := NewHistoryEventVertex(types.EventTypeWorkflowExecutionTerminated.String())
	workflowTerminate.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		EventID := lastEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeWorkflowExecutionTerminated.Ptr()
		historyEvent.WorkflowExecutionTerminatedEventAttributes = &types.WorkflowExecutionTerminatedEventAttributes{
			Identity: identity,
			Reason:   reason,
		}
		return historyEvent
	})
	workflowTimedOut := NewHistoryEventVertex(types.EventTypeWorkflowExecutionTimedOut.String())
	workflowTimedOut.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		EventID := lastEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeWorkflowExecutionTimedOut.Ptr()
		historyEvent.WorkflowExecutionTimedOutEventAttributes = &types.WorkflowExecutionTimedOutEventAttributes{
			TimeoutType: types.TimeoutTypeStartToClose.Ptr(),
		}
		return historyEvent
	})
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

	// Setup activity model
	activityModel := NewHistoryEventModel()
	activitySchedule := NewHistoryEventVertex(types.EventTypeActivityTaskScheduled.String())
	activitySchedule.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeActivityTaskScheduled.Ptr()
		historyEvent.ActivityTaskScheduledEventAttributes = &types.ActivityTaskScheduledEventAttributes{
			ActivityID: uuid.New(),
			ActivityType: &types.ActivityType{
				Name: "activity",
			},
			Domain: common.StringPtr(domain),
			TaskList: &types.TaskList{
				Name: taskList,
				Kind: types.TaskListKindNormal.Ptr(),
			},
			ScheduleToCloseTimeoutSeconds: common.Int32Ptr(timeout),
			ScheduleToStartTimeoutSeconds: common.Int32Ptr(timeout),
			StartToCloseTimeoutSeconds:    common.Int32Ptr(timeout),
			DecisionTaskCompletedEventID:  lastEvent.EventID,
		}
		return historyEvent
	})
	activityStart := NewHistoryEventVertex(types.EventTypeActivityTaskStarted.String())
	activityStart.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeActivityTaskStarted.Ptr()
		historyEvent.ActivityTaskStartedEventAttributes = &types.ActivityTaskStartedEventAttributes{
			ScheduledEventID: lastEvent.EventID,
			Identity:         identity,
			RequestID:        uuid.New(),
			Attempt:          0,
		}
		return historyEvent
	})
	activityComplete := NewHistoryEventVertex(types.EventTypeActivityTaskCompleted.String())
	activityComplete.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeActivityTaskCompleted.Ptr()
		historyEvent.ActivityTaskCompletedEventAttributes = &types.ActivityTaskCompletedEventAttributes{
			ScheduledEventID: lastEvent.GetActivityTaskStartedEventAttributes().ScheduledEventID,
			StartedEventID:   lastEvent.EventID,
			Identity:         identity,
		}
		return historyEvent
	})
	activityFail := NewHistoryEventVertex(types.EventTypeActivityTaskFailed.String())
	activityFail.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeActivityTaskFailed.Ptr()
		historyEvent.ActivityTaskFailedEventAttributes = &types.ActivityTaskFailedEventAttributes{
			ScheduledEventID: lastEvent.GetActivityTaskStartedEventAttributes().ScheduledEventID,
			StartedEventID:   lastEvent.EventID,
			Identity:         identity,
			Reason:           common.StringPtr(reason),
		}
		return historyEvent
	})
	activityTimedOut := NewHistoryEventVertex(types.EventTypeActivityTaskTimedOut.String())
	activityTimedOut.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeActivityTaskTimedOut.Ptr()
		historyEvent.ActivityTaskTimedOutEventAttributes = &types.ActivityTaskTimedOutEventAttributes{
			ScheduledEventID: lastEvent.GetActivityTaskStartedEventAttributes().ScheduledEventID,
			StartedEventID:   lastEvent.EventID,
			TimeoutType:      types.TimeoutTypeScheduleToClose.Ptr(),
		}
		return historyEvent
	})
	activityCancelRequest := NewHistoryEventVertex(types.EventTypeActivityTaskCancelRequested.String())
	activityCancelRequest.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeActivityTaskCancelRequested.Ptr()
		historyEvent.ActivityTaskCancelRequestedEventAttributes = &types.ActivityTaskCancelRequestedEventAttributes{
			DecisionTaskCompletedEventID: lastEvent.GetActivityTaskScheduledEventAttributes().DecisionTaskCompletedEventID,
			ActivityID:                   lastEvent.GetActivityTaskScheduledEventAttributes().ActivityID,
		}
		return historyEvent
	})
	activityCancel := NewHistoryEventVertex(types.EventTypeActivityTaskCanceled.String())
	activityCancel.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeActivityTaskCanceled.Ptr()
		historyEvent.ActivityTaskCanceledEventAttributes = &types.ActivityTaskCanceledEventAttributes{
			LatestCancelRequestedEventID: lastEvent.EventID,
			ScheduledEventID:             lastEvent.EventID,
			StartedEventID:               lastEvent.EventID,
			Identity:                     identity,
		}
		return historyEvent
	})
	activityCancelRequestFail := NewHistoryEventVertex(types.EventTypeRequestCancelActivityTaskFailed.String())
	activityCancelRequestFail.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		versionBump := input[2].(int64)
		subVersion := input[3].(int64)
		version := lastGeneratedEvent.GetVersion() + versionBump + subVersion
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeRequestCancelActivityTaskFailed.Ptr()
		historyEvent.RequestCancelActivityTaskFailedEventAttributes = &types.RequestCancelActivityTaskFailedEventAttributes{
			ActivityID:                   uuid.New(),
			DecisionTaskCompletedEventID: lastEvent.GetActivityTaskCancelRequestedEventAttributes().DecisionTaskCompletedEventID,
		}
		return historyEvent
	})
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

	// TODO: bypass activity cancel request event. Support this event later.
	//activityScheduleToActivityCancelRequest := NewHistoryEventEdge(activitySchedule, activityCancelRequest)
	//activityScheduleToActivityCancelRequest.SetCondition(hasPendingActivity)
	activityCancelReqToCancel := NewHistoryEventEdge(activityCancelRequest, activityCancel)
	activityCancelReqToCancel.SetCondition(hasPendingActivity)

	activityCancelReqToCancelFail := NewHistoryEventEdge(activityCancelRequest, activityCancelRequestFail)
	activityCancelRequestFailToDecisionSchedule := NewHistoryEventEdge(activityCancelRequestFail, decisionSchedule)
	activityCancelRequestFailToDecisionSchedule.SetCondition(notPendingDecisionTask)

	activityModel.AddEdge(decisionCompleteToATSchedule, activityScheduleToStart, activityStartToComplete,
		activityStartToFail, activityStartToTimedOut, decisionCompleteToATSchedule, activityCompleteToDecisionSchedule,
		activityFailToDecisionSchedule, activityTimedOutToDecisionSchedule, activityCancelReqToCancel,
		activityCancelReqToCancelFail, activityCancelToDecisionSchedule, activityCancelRequestFailToDecisionSchedule)

	// Setup timer model
	timerModel := NewHistoryEventModel()
	timerStart := NewHistoryEventVertex(types.EventTypeTimerStarted.String())
	timerStart.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeTimerStarted.Ptr()
		historyEvent.TimerStartedEventAttributes = &types.TimerStartedEventAttributes{
			TimerID:                      uuid.New(),
			StartToFireTimeoutSeconds:    common.Int64Ptr(10),
			DecisionTaskCompletedEventID: lastEvent.EventID,
		}
		return historyEvent
	})
	timerFired := NewHistoryEventVertex(types.EventTypeTimerFired.String())
	timerFired.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeTimerFired.Ptr()
		historyEvent.TimerFiredEventAttributes = &types.TimerFiredEventAttributes{
			TimerID:        lastEvent.GetTimerStartedEventAttributes().TimerID,
			StartedEventID: lastEvent.EventID,
		}
		return historyEvent
	})
	timerCancel := NewHistoryEventVertex(types.EventTypeTimerCanceled.String())
	timerCancel.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeTimerCanceled.Ptr()
		historyEvent.TimerCanceledEventAttributes = &types.TimerCanceledEventAttributes{
			TimerID:                      lastEvent.GetTimerStartedEventAttributes().TimerID,
			StartedEventID:               lastEvent.EventID,
			DecisionTaskCompletedEventID: lastEvent.GetTimerStartedEventAttributes().DecisionTaskCompletedEventID,
			Identity:                     identity,
		}
		return historyEvent
	})
	timerStartToFire := NewHistoryEventEdge(timerStart, timerFired)
	timerStartToCancel := NewHistoryEventEdge(timerStart, timerCancel)

	decisionCompleteToTimerStart := NewHistoryEventEdge(decisionComplete, timerStart)
	timerFiredToDecisionSchedule := NewHistoryEventEdge(timerFired, decisionSchedule)
	timerFiredToDecisionSchedule.SetCondition(notPendingDecisionTask)
	timerCancelToDecisionSchedule := NewHistoryEventEdge(timerCancel, decisionSchedule)
	timerCancelToDecisionSchedule.SetCondition(notPendingDecisionTask)
	timerModel.AddEdge(timerStartToFire, timerStartToCancel, decisionCompleteToTimerStart, timerFiredToDecisionSchedule, timerCancelToDecisionSchedule)

	// Setup child workflow model
	childWorkflowModel := NewHistoryEventModel()
	childWorkflowInitial := NewHistoryEventVertex(types.EventTypeStartChildWorkflowExecutionInitiated.String())
	childWorkflowInitial.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeStartChildWorkflowExecutionInitiated.Ptr()
		historyEvent.StartChildWorkflowExecutionInitiatedEventAttributes = &types.StartChildWorkflowExecutionInitiatedEventAttributes{
			Domain:     domain,
			WorkflowID: childWorkflowID,
			WorkflowType: &types.WorkflowType{
				Name: childWorkflowPrefix + workflowType,
			},
			TaskList: &types.TaskList{
				Name: taskList,
				Kind: types.TaskListKindNormal.Ptr(),
			},
			ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(timeout),
			TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(timeout),
			DecisionTaskCompletedEventID:        lastEvent.EventID,
			WorkflowIDReusePolicy:               types.WorkflowIDReusePolicyRejectDuplicate.Ptr(),
		}
		return historyEvent
	})
	childWorkflowInitialFail := NewHistoryEventVertex(types.EventTypeStartChildWorkflowExecutionFailed.String())
	childWorkflowInitialFail.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeStartChildWorkflowExecutionFailed.Ptr()
		historyEvent.StartChildWorkflowExecutionFailedEventAttributes = &types.StartChildWorkflowExecutionFailedEventAttributes{
			Domain:     domain,
			WorkflowID: childWorkflowID,
			WorkflowType: &types.WorkflowType{
				Name: childWorkflowPrefix + workflowType,
			},
			Cause:                        types.ChildWorkflowExecutionFailedCauseWorkflowAlreadyRunning.Ptr(),
			InitiatedEventID:             lastEvent.EventID,
			DecisionTaskCompletedEventID: lastEvent.GetStartChildWorkflowExecutionInitiatedEventAttributes().DecisionTaskCompletedEventID,
		}
		return historyEvent
	})
	childWorkflowStart := NewHistoryEventVertex(types.EventTypeChildWorkflowExecutionStarted.String())
	childWorkflowStart.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeChildWorkflowExecutionStarted.Ptr()
		historyEvent.ChildWorkflowExecutionStartedEventAttributes = &types.ChildWorkflowExecutionStartedEventAttributes{
			Domain: domain,
			WorkflowType: &types.WorkflowType{
				Name: childWorkflowPrefix + workflowType,
			},
			InitiatedEventID: lastEvent.EventID,
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: childWorkflowID,
				RunID:      uuid.New(),
			},
		}
		return historyEvent
	})
	childWorkflowCancel := NewHistoryEventVertex(types.EventTypeChildWorkflowExecutionCanceled.String())
	childWorkflowCancel.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeChildWorkflowExecutionCanceled.Ptr()
		historyEvent.ChildWorkflowExecutionCanceledEventAttributes = &types.ChildWorkflowExecutionCanceledEventAttributes{
			Domain: domain,
			WorkflowType: &types.WorkflowType{
				Name: childWorkflowPrefix + workflowType,
			},
			InitiatedEventID: lastEvent.GetChildWorkflowExecutionStartedEventAttributes().InitiatedEventID,
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: childWorkflowID,
				RunID:      lastEvent.GetChildWorkflowExecutionStartedEventAttributes().GetWorkflowExecution().RunID,
			},
			StartedEventID: lastEvent.EventID,
		}
		return historyEvent
	})
	childWorkflowComplete := NewHistoryEventVertex(types.EventTypeChildWorkflowExecutionCompleted.String())
	childWorkflowComplete.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeChildWorkflowExecutionCompleted.Ptr()
		historyEvent.ChildWorkflowExecutionCompletedEventAttributes = &types.ChildWorkflowExecutionCompletedEventAttributes{
			Domain: domain,
			WorkflowType: &types.WorkflowType{
				Name: childWorkflowPrefix + workflowType,
			},
			InitiatedEventID: lastEvent.GetChildWorkflowExecutionStartedEventAttributes().InitiatedEventID,
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: childWorkflowID,
				RunID:      lastEvent.GetChildWorkflowExecutionStartedEventAttributes().GetWorkflowExecution().RunID,
			},
			StartedEventID: lastEvent.EventID,
		}
		return historyEvent
	})
	childWorkflowFail := NewHistoryEventVertex(types.EventTypeChildWorkflowExecutionFailed.String())
	childWorkflowFail.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeChildWorkflowExecutionFailed.Ptr()
		historyEvent.ChildWorkflowExecutionFailedEventAttributes = &types.ChildWorkflowExecutionFailedEventAttributes{
			Domain: domain,
			WorkflowType: &types.WorkflowType{
				Name: childWorkflowPrefix + workflowType,
			},
			InitiatedEventID: lastEvent.GetChildWorkflowExecutionStartedEventAttributes().InitiatedEventID,
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: childWorkflowID,
				RunID:      lastEvent.GetChildWorkflowExecutionStartedEventAttributes().GetWorkflowExecution().RunID,
			},
			StartedEventID: lastEvent.EventID,
		}
		return historyEvent
	})
	childWorkflowTerminate := NewHistoryEventVertex(types.EventTypeChildWorkflowExecutionTerminated.String())
	childWorkflowTerminate.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeChildWorkflowExecutionTerminated.Ptr()
		historyEvent.ChildWorkflowExecutionTerminatedEventAttributes = &types.ChildWorkflowExecutionTerminatedEventAttributes{
			Domain: domain,
			WorkflowType: &types.WorkflowType{
				Name: childWorkflowPrefix + workflowType,
			},
			InitiatedEventID: lastEvent.GetChildWorkflowExecutionStartedEventAttributes().InitiatedEventID,
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: childWorkflowID,
				RunID:      lastEvent.GetChildWorkflowExecutionStartedEventAttributes().GetWorkflowExecution().RunID,
			},
			StartedEventID: lastEvent.EventID,
		}
		return historyEvent
	})
	childWorkflowTimedOut := NewHistoryEventVertex(types.EventTypeChildWorkflowExecutionTimedOut.String())
	childWorkflowTimedOut.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeChildWorkflowExecutionTimedOut.Ptr()
		historyEvent.ChildWorkflowExecutionTimedOutEventAttributes = &types.ChildWorkflowExecutionTimedOutEventAttributes{
			Domain: domain,
			WorkflowType: &types.WorkflowType{
				Name: childWorkflowPrefix + workflowType,
			},
			InitiatedEventID: lastEvent.GetChildWorkflowExecutionStartedEventAttributes().InitiatedEventID,
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: childWorkflowID,
				RunID:      lastEvent.GetChildWorkflowExecutionStartedEventAttributes().GetWorkflowExecution().RunID,
			},
			StartedEventID: lastEvent.EventID,
			TimeoutType:    types.TimeoutTypeScheduleToClose.Ptr(),
		}
		return historyEvent
	})
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

	// Setup external workflow model
	externalWorkflowModel := NewHistoryEventModel()
	externalWorkflowSignal := NewHistoryEventVertex(types.EventTypeSignalExternalWorkflowExecutionInitiated.String())
	externalWorkflowSignal.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeSignalExternalWorkflowExecutionInitiated.Ptr()
		historyEvent.SignalExternalWorkflowExecutionInitiatedEventAttributes = &types.SignalExternalWorkflowExecutionInitiatedEventAttributes{
			DecisionTaskCompletedEventID: lastEvent.EventID,
			Domain:                       domain,
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: externalWorkflowID,
				RunID:      uuid.New(),
			},
			SignalName:        "signal",
			ChildWorkflowOnly: false,
		}
		return historyEvent
	})
	externalWorkflowSignalFailed := NewHistoryEventVertex(types.EventTypeSignalExternalWorkflowExecutionFailed.String())
	externalWorkflowSignalFailed.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeSignalExternalWorkflowExecutionFailed.Ptr()
		historyEvent.SignalExternalWorkflowExecutionFailedEventAttributes = &types.SignalExternalWorkflowExecutionFailedEventAttributes{
			Cause:                        types.SignalExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution.Ptr(),
			DecisionTaskCompletedEventID: lastEvent.GetSignalExternalWorkflowExecutionInitiatedEventAttributes().DecisionTaskCompletedEventID,
			Domain:                       domain,
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: lastEvent.GetSignalExternalWorkflowExecutionInitiatedEventAttributes().GetWorkflowExecution().WorkflowID,
				RunID:      lastEvent.GetSignalExternalWorkflowExecutionInitiatedEventAttributes().GetWorkflowExecution().RunID,
			},
			InitiatedEventID: lastEvent.EventID,
		}
		return historyEvent
	})
	externalWorkflowSignaled := NewHistoryEventVertex(types.EventTypeExternalWorkflowExecutionSignaled.String())
	externalWorkflowSignaled.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeExternalWorkflowExecutionSignaled.Ptr()
		historyEvent.ExternalWorkflowExecutionSignaledEventAttributes = &types.ExternalWorkflowExecutionSignaledEventAttributes{
			InitiatedEventID: lastEvent.EventID,
			Domain:           domain,
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: lastEvent.GetSignalExternalWorkflowExecutionInitiatedEventAttributes().GetWorkflowExecution().WorkflowID,
				RunID:      lastEvent.GetSignalExternalWorkflowExecutionInitiatedEventAttributes().GetWorkflowExecution().RunID,
			},
		}
		return historyEvent
	})
	externalWorkflowCancel := NewHistoryEventVertex(types.EventTypeRequestCancelExternalWorkflowExecutionInitiated.String())
	externalWorkflowCancel.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeRequestCancelExternalWorkflowExecutionInitiated.Ptr()
		historyEvent.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes =
			&types.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{
				DecisionTaskCompletedEventID: lastEvent.EventID,
				Domain:                       domain,
				WorkflowExecution: &types.WorkflowExecution{
					WorkflowID: externalWorkflowID,
					RunID:      uuid.New(),
				},
				ChildWorkflowOnly: false,
			}
		return historyEvent
	})
	externalWorkflowCancelFail := NewHistoryEventVertex(types.EventTypeRequestCancelExternalWorkflowExecutionFailed.String())
	externalWorkflowCancelFail.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeRequestCancelExternalWorkflowExecutionFailed.Ptr()
		historyEvent.RequestCancelExternalWorkflowExecutionFailedEventAttributes = &types.RequestCancelExternalWorkflowExecutionFailedEventAttributes{
			Cause:                        types.CancelExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution.Ptr(),
			DecisionTaskCompletedEventID: lastEvent.GetRequestCancelExternalWorkflowExecutionInitiatedEventAttributes().DecisionTaskCompletedEventID,
			Domain:                       domain,
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: lastEvent.GetRequestCancelExternalWorkflowExecutionInitiatedEventAttributes().GetWorkflowExecution().WorkflowID,
				RunID:      lastEvent.GetRequestCancelExternalWorkflowExecutionInitiatedEventAttributes().GetWorkflowExecution().RunID,
			},
			InitiatedEventID: lastEvent.EventID,
		}
		return historyEvent
	})
	externalWorkflowCanceled := NewHistoryEventVertex(types.EventTypeExternalWorkflowExecutionCancelRequested.String())
	externalWorkflowCanceled.SetDataFunc(func(input ...interface{}) interface{} {
		lastEvent := input[0].(*types.HistoryEvent)
		lastGeneratedEvent := input[1].(*types.HistoryEvent)
		EventID := lastGeneratedEvent.GetEventID() + 1
		version := input[2].(int64)
		historyEvent := getDefaultHistoryEvent(EventID, version)
		historyEvent.EventType = types.EventTypeExternalWorkflowExecutionCancelRequested.Ptr()
		historyEvent.ExternalWorkflowExecutionCancelRequestedEventAttributes = &types.ExternalWorkflowExecutionCancelRequestedEventAttributes{
			InitiatedEventID: lastEvent.EventID,
			Domain:           domain,
			WorkflowExecution: &types.WorkflowExecution{
				WorkflowID: lastEvent.GetRequestCancelExternalWorkflowExecutionInitiatedEventAttributes().GetWorkflowExecution().WorkflowID,
				RunID:      lastEvent.GetRequestCancelExternalWorkflowExecutionInitiatedEventAttributes().GetWorkflowExecution().RunID,
			},
		}
		return historyEvent
	})
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

	// Config event generator
	generator.SetBatchGenerationRule(canDoBatch)
	generator.AddInitialEntryVertex(workflowStart)
	generator.AddExitVertex(workflowComplete, workflowFail, workflowTerminate, workflowTimedOut, continueAsNew)
	// generator.AddRandomEntryVertex(workflowSignal, workflowTerminate, workflowTimedOut)
	generator.AddModel(decisionModel)
	generator.AddModel(workflowModel)
	generator.AddModel(activityModel)
	generator.AddModel(timerModel)
	generator.AddModel(childWorkflowModel)
	generator.AddModel(externalWorkflowModel)
	return generator
}

func getDefaultHistoryEvent(
	EventID int64,
	version int64,
) *types.HistoryEvent {

	globalTaskID++
	return &types.HistoryEvent{
		EventID:   EventID,
		Timestamp: common.Int64Ptr(time.Now().UnixNano()),
		TaskID:    globalTaskID,
		Version:   version,
	}
}

func copyConnections(
	originalMap map[string][]Edge,
) map[string][]Edge {

	newMap := make(map[string][]Edge)
	for key, value := range originalMap {
		newMap[key] = copyEdges(value)
	}
	return newMap
}

func copyExitVertices(
	originalMap map[string]bool,
) map[string]bool {

	newMap := make(map[string]bool)
	for key, value := range originalMap {
		newMap[key] = value
	}
	return newMap
}

func copyVertex(vertex []Vertex) []Vertex {
	newVertex := make([]Vertex, len(vertex))
	for idx, v := range vertex {
		newVertex[idx] = v.DeepCopy()
	}
	return newVertex
}

func copyEdges(edges []Edge) []Edge {
	newEdges := make([]Edge, len(edges))
	for idx, e := range edges {
		newEdges[idx] = e.DeepCopy()
	}
	return newEdges
}
