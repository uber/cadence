// Copyright (c) 2017 Uber Technologies, Inc.
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

package history

import (
	"fmt"
	"sync"

	"github.com/uber-common/bark"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/persistence"
)

// NOTE: this whole file is used by standby cluster dealing with
// workflow timer which got fired.

type (
	// outstandingTimers keeps all the timer which is fired for a workflow
	// in standby cluster, and when we do domain failover, all fired timer
	// will be redirected to timer queue processor for processing.
	outstandingTimers struct {
		// workflow timer
		workflowStartToClose *persistence.TimerTaskInfo

		// decision timers
		decisionScheduleToStartTimer *persistence.TimerTaskInfo
		decisionStartToCloseTimer    *persistence.TimerTaskInfo

		// activity timers
		// timer builder will only create one timer for all activity timers at a given time,
		// i.e. sort the fire time of all activity timers and use the earlist fire time
		// as the timer task for activity timer.
		// what need to be done here is tracking the latest timer task, and that is all.
		activityTimer *persistence.TimerTaskInfo

		// user timers
		// timer builder will only create one timer for all user timers at a given time,
		// i.e. sort the fire time of all user timers and use the earlist fire time
		// as the timer task for user timer.
		// what need to be done here is tracking the latest timer task, and that is all.
		userTimer *persistence.TimerTaskInfo
	}

	timerQueueStandbyProcessor struct {
		logger           bark.Logger
		timerQueueAckMgr timerQueueAckMgr

		sync.Mutex
		workflowOutstandingTimers map[workflowIdentifier]*outstandingTimers
	}
)

// newTimerQueueStandbyProcessor create a new standby timer processor
func newTimerQueueStandbyProcessor(logger bark.Logger, timerQueueAckMgr timerQueueAckMgr) *timerQueueStandbyProcessor {
	return &timerQueueStandbyProcessor{
		logger:                    logger,
		timerQueueAckMgr:          timerQueueAckMgr,
		workflowOutstandingTimers: make(map[workflowIdentifier]*outstandingTimers),
	}
}

// newOutstandingTimers create a new empty outstandingTimers for a a workflow
func newOutstandingTimers() *outstandingTimers {
	return &outstandingTimers{}
}

func (processor *timerQueueStandbyProcessor) AddTimers(timers []*persistence.TimerTaskInfo) {
	if len(timers) == 0 {
		return
	}

	processor.Lock()
	defer processor.Unlock()

	for _, timer := range timers {
		identifier := workflowIdentifier{
			domainID:   timer.DomainID,
			workflowID: timer.WorkflowID,
			runID:      timer.RunID,
		}
		outstandingTimers, ok := processor.workflowOutstandingTimers[identifier]
		if !ok {
			outstandingTimers = newOutstandingTimers()
			processor.workflowOutstandingTimers[identifier] = outstandingTimers
		}

		switch timer.TaskType {
		case persistence.TaskTypeWorkflowTimeout:
			// there can only be one workflow start to close timer
			// so making the check if this timer is set does not make much sense
			outstandingTimers.workflowStartToClose = timer

		case persistence.TaskTypeDecisionTimeout:
			switch timer.TimeoutType {
			case int(workflow.TimeoutTypeStartToClose):
				if outstandingTimers.decisionStartToCloseTimer != nil {
					// this should not happen, since there can only be one decision
					// on the fly and one corresponding decision start to close timer
					panic(fmt.Sprintf("Decision start to close timer is not cleared, current: %v, incoming %v",
						outstandingTimers.decisionStartToCloseTimer, timer))
				}
				outstandingTimers.decisionStartToCloseTimer = timer
			case int(workflow.TimeoutTypeScheduleToStart):
				if outstandingTimers.decisionScheduleToStartTimer != nil {
					// this should not happen, since there can only be one decision
					// on the fly and one corresponding decision schedule to start timer
					panic(fmt.Sprintf("Decision schedule to start timer is not cleared, current: %v, incoming %v",
						outstandingTimers.decisionScheduleToStartTimer, timer))
				}
				outstandingTimers.decisionScheduleToStartTimer = timer
			default:
				panic(fmt.Sprintf("Unknown decision task type: %v.", timer.TimeoutType))
			}

		case persistence.TaskTypeActivityTimeout:
			if outstandingTimers.activityTimer == nil {
				outstandingTimers.activityTimer = timer
			} else {
				activityTimerSequenceID := &TimerSequenceID{
					VisibilityTimestamp: outstandingTimers.activityTimer.VisibilityTimestamp,
					TaskID:              outstandingTimers.activityTimer.TaskID,
				}
				timerSequenceID := &TimerSequenceID{
					VisibilityTimestamp: timer.VisibilityTimestamp,
					TaskID:              timer.TaskID,
				}
				if compareTimerIDLess(activityTimerSequenceID, timerSequenceID) {
					// the current user timer is processed, update timer queue ack manager
					processor.ackTimer(outstandingTimers.activityTimer)

					outstandingTimers.activityTimer = timer
				}
			}

		case persistence.TaskTypeUserTimer:
			if outstandingTimers.userTimer == nil {
				outstandingTimers.userTimer = timer
			} else {
				userTimerSequenceID := &TimerSequenceID{
					VisibilityTimestamp: outstandingTimers.userTimer.VisibilityTimestamp,
					TaskID:              outstandingTimers.userTimer.TaskID,
				}
				timerSequenceID := &TimerSequenceID{
					VisibilityTimestamp: timer.VisibilityTimestamp,
					TaskID:              timer.TaskID,
				}
				if compareTimerIDLess(userTimerSequenceID, timerSequenceID) {
					// the current user timer is processed, update timer queue ack manager
					processor.ackTimer(outstandingTimers.userTimer)
					outstandingTimers.userTimer = timer
				}
			}
		case persistence.TaskTypeDeleteHistoryEvent:
			panic("Timer task delete history event should not be buffered.")
		default:
			panic(fmt.Sprintf("Unknown timer task type: %v.", timer.TaskType))
		}

	}
}

func (processor *timerQueueStandbyProcessor) CompleteTimers(identifier workflowIdentifier, events []*workflow.HistoryEvent) {
	if len(events) == 0 {
		return
	}

	// There are 4 types of timer we need to take care of
	// 1. decision timer
	// 2. activity timer
	// 3. workflow timer
	// 4. user timer
	//
	// The delete history timer should not be processed here,
	// since there is no event representing history deletion,
	// moreover this timer is to honor the retention policy
	// and has nothing to do with timer business logic.
	for _, event := range events {
		switch event.GetEventType() {
		case
			// below are all workflow state change history event
			// only put the started event here for reference
			// workflow.EventTypeWorkflowExecutionStarted
			workflow.EventTypeWorkflowExecutionCompleted,
			workflow.EventTypeWorkflowExecutionFailed,
			workflow.EventTypeWorkflowExecutionTimedOut,
			workflow.EventTypeWorkflowExecutionCanceled,
			workflow.EventTypeWorkflowExecutionTerminated,
			workflow.EventTypeWorkflowExecutionContinuedAsNew:
			processor.completeWorkflowTimer(identifier, event)
		case
			// below are all decision state change history event
			// only put the schedule event here for reference
			// workflow.EventTypeDecisionTaskScheduled
			workflow.EventTypeDecisionTaskStarted,
			workflow.EventTypeDecisionTaskCompleted,
			workflow.EventTypeDecisionTaskTimedOut,
			workflow.EventTypeDecisionTaskFailed:
			processor.completeDecisionTimer(identifier, event)
		case
			// below are all activity state change history event
			// only put the schedule event here for reference
			// workflow.EventTypeActivityTaskScheduled
			workflow.EventTypeActivityTaskStarted,
			workflow.EventTypeActivityTaskCompleted,
			workflow.EventTypeActivityTaskFailed,
			workflow.EventTypeActivityTaskTimedOut,
			workflow.EventTypeActivityTaskCanceled:
			// we do not do anything to activity timer event here,
			// this is because activity timer is triggered by the earlist
			// time of all activity timers, not a specific activity ID.
			// NOTE:
			// the current activity timer is considered completed
			// when a new activity timer is received
		case
			workflow.EventTypeTimerStarted,
			workflow.EventTypeTimerFired,
			workflow.EventTypeTimerCanceled:
			// we do not do anything to user timer event here,
			// this is because user timer is triggered by the earlist
			// time of all user timers, not a specific timer ID.
			// NOTE:
			// the current user timer is considered completed
			// when a new user timer is received
		default:
			// other event types do not have any timer relations, just skip
		}
	}
}

func (processor *timerQueueStandbyProcessor) GetAndRemoveTimersByDomainID(domainID string) []*persistence.TimerTaskInfo {
	// TODO pending implementation
	return nil
}

func (processor *timerQueueStandbyProcessor) completeWorkflowTimer(identifier workflowIdentifier, event *workflow.HistoryEvent) {

	processor.Lock()
	defer processor.Unlock()

	outstandingTimers, ok := processor.workflowOutstandingTimers[identifier]
	if !ok {
		processor.logger.Debugf("Event %v being processed is an duplicate.", event)
		return
	}

	// for each timer within this workflow, update the timer ack manager
	processor.ackTimer(outstandingTimers.workflowStartToClose)
	processor.ackTimer(outstandingTimers.decisionScheduleToStartTimer)
	processor.ackTimer(outstandingTimers.decisionStartToCloseTimer)
	processor.ackTimer(outstandingTimers.activityTimer)
	processor.ackTimer(outstandingTimers.userTimer)

	// for workflow timer, there are only 1 type
	// 1. start to close.
	// one can use regex `WorkflowTimeoutTask\{` to find out the usage
	//
	// we should set the workflowStartToClose to nil, however, since the workflow is finished
	// just clear all timer related to this workflow
	delete(processor.workflowOutstandingTimers, identifier)
}

func (processor *timerQueueStandbyProcessor) completeDecisionTimer(identifier workflowIdentifier, event *workflow.HistoryEvent) {

	processor.Lock()
	defer processor.Unlock()

	outstandingTimers, ok := processor.workflowOutstandingTimers[identifier]
	if !ok {
		processor.logger.Debugf("Event %v being processed is an duplicate.", event)
		return
	}

	var scheduleID int64
	switch event.GetEventType() {
	case workflow.EventTypeDecisionTaskStarted:
		scheduleID = event.DecisionTaskStartedEventAttributes.GetScheduledEventId()
	case workflow.EventTypeDecisionTaskCompleted:
		scheduleID = event.DecisionTaskCompletedEventAttributes.GetScheduledEventId()
	case workflow.EventTypeDecisionTaskTimedOut:
		scheduleID = event.DecisionTaskTimedOutEventAttributes.GetScheduledEventId()
	case workflow.EventTypeDecisionTaskFailed:
		scheduleID = event.DecisionTaskFailedEventAttributes.GetScheduledEventId()
	default:
		panic(fmt.Sprintf("Unknown decision event type: %v.", event.GetEventType()))
	}

	// for decision timer, there are 2 types
	// 1. schedule to start. only used by sticky decision as of Mar 16, 2018
	// 2. start to close.
	// one can use regex `\.Add(ScheduleToStart){0,1}DecisionTimoutTask\(` to find out the usage
	switch event.GetEventType() {
	case workflow.EventTypeDecisionTaskStarted:
		if outstandingTimers.decisionScheduleToStartTimer != nil {
			if outstandingTimers.decisionScheduleToStartTimer.EventID == scheduleID {
				processor.ackTimer(outstandingTimers.decisionScheduleToStartTimer)
				outstandingTimers.decisionScheduleToStartTimer = nil
			} else {
				panic(fmt.Sprintf("Decision schedule to start timer mismatch, current: %v, incoming %v",
					outstandingTimers.decisionScheduleToStartTimer, event))
			}
		}
	case workflow.EventTypeDecisionTaskTimedOut:
		if outstandingTimers.decisionScheduleToStartTimer != nil {
			if outstandingTimers.decisionScheduleToStartTimer.EventID == scheduleID {
				processor.ackTimer(outstandingTimers.decisionScheduleToStartTimer)
				outstandingTimers.decisionScheduleToStartTimer = nil
			} else {
				panic(fmt.Sprintf("Decision schedule to start timer mismatch, current: %v, incoming %v",
					outstandingTimers.decisionScheduleToStartTimer, event))
			}
		}
		if outstandingTimers.decisionStartToCloseTimer != nil {
			if outstandingTimers.decisionStartToCloseTimer.EventID == scheduleID {
				processor.ackTimer(outstandingTimers.decisionStartToCloseTimer)
				outstandingTimers.decisionStartToCloseTimer = nil
			} else {
				panic(fmt.Sprintf("Decision start to close timer mismatch, current: %v, incoming %v",
					outstandingTimers.decisionScheduleToStartTimer, event))
			}
		}
	case workflow.EventTypeDecisionTaskCompleted,
		workflow.EventTypeDecisionTaskFailed:

		// check start to close timer
		if outstandingTimers.decisionStartToCloseTimer.EventID == scheduleID {
			processor.ackTimer(outstandingTimers.decisionStartToCloseTimer)
			outstandingTimers.decisionStartToCloseTimer = nil
		} else {
			panic(fmt.Sprintf("Decision start to close timer mismatch, current: %v, incoming %v",
				outstandingTimers.decisionScheduleToStartTimer, event))
		}

	}
}

func (processor *timerQueueStandbyProcessor) ackTimer(timer *persistence.TimerTaskInfo) {
	if timer == nil {
		return
	}
	timerSequenceID := TimerSequenceID{
		VisibilityTimestamp: timer.VisibilityTimestamp,
		TaskID:              timer.TaskID,
	}

	processor.timerQueueAckMgr.completeTimerTask(timerSequenceID)
}
