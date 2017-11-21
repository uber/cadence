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

package frontend

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	gen "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
)

const (
	workflowType = "dummy workflow type"
	workflowID   = "dummy workflow ID"

	activityType = "dummy activity type"
	activityID   = "dummy activity ID"
	eventID      = 59

	signalName = "dummy signal"
	markerName = "dummy marker"
)

var (
	/*
	 * The order of event type is the same as history events defined in shared.thrift HistoryEvent type
	 */
	historyEventMapping = map[gen.EventType]*gen.HistoryEvent{
		gen.EventTypeWorkflowExecutionStarted:   &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeWorkflowExecutionStarted)},
		gen.EventTypeWorkflowExecutionCompleted: &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeWorkflowExecutionCompleted)},
		gen.EventTypeWorkflowExecutionFailed:    &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeWorkflowExecutionFailed)},
		gen.EventTypeWorkflowExecutionTimedOut:  &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeWorkflowExecutionTimedOut)},
		gen.EventTypeDecisionTaskScheduled:      &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeDecisionTaskScheduled)},
		gen.EventTypeDecisionTaskStarted:        &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeDecisionTaskStarted)},
		gen.EventTypeDecisionTaskCompleted:      &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeDecisionTaskCompleted)},
		gen.EventTypeDecisionTaskTimedOut:       &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeDecisionTaskTimedOut)},
		gen.EventTypeDecisionTaskFailed:         &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeDecisionTaskFailed)},

		gen.EventTypeActivityTaskScheduled: &gen.HistoryEvent{
			EventId:   common.Int64Ptr(eventID),
			EventType: common.EventTypePtr(gen.EventTypeActivityTaskScheduled),
			ActivityTaskScheduledEventAttributes: &gen.ActivityTaskScheduledEventAttributes{
				ActivityType: &gen.ActivityType{
					Name: common.StringPtr(activityType),
				},
				ActivityId: common.StringPtr(activityID),
			},
		},

		gen.EventTypeActivityTaskStarted: &gen.HistoryEvent{
			EventType: common.EventTypePtr(gen.EventTypeActivityTaskStarted),
			ActivityTaskStartedEventAttributes: &gen.ActivityTaskStartedEventAttributes{
				ScheduledEventId: common.Int64Ptr(eventID),
			},
		},
		gen.EventTypeActivityTaskCompleted: &gen.HistoryEvent{
			EventType: common.EventTypePtr(gen.EventTypeActivityTaskCompleted),
			ActivityTaskCompletedEventAttributes: &gen.ActivityTaskCompletedEventAttributes{
				ScheduledEventId: common.Int64Ptr(eventID),
			},
		},
		gen.EventTypeActivityTaskFailed: &gen.HistoryEvent{
			EventType: common.EventTypePtr(gen.EventTypeActivityTaskFailed),
			ActivityTaskFailedEventAttributes: &gen.ActivityTaskFailedEventAttributes{
				ScheduledEventId: common.Int64Ptr(eventID),
			},
		},
		gen.EventTypeActivityTaskTimedOut: &gen.HistoryEvent{
			EventType: common.EventTypePtr(gen.EventTypeActivityTaskTimedOut),
			ActivityTaskTimedOutEventAttributes: &gen.ActivityTaskTimedOutEventAttributes{
				ScheduledEventId: common.Int64Ptr(eventID),
			},
		},
		gen.EventTypeTimerStarted:                    &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeTimerStarted)},
		gen.EventTypeTimerFired:                      &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeTimerFired)},
		gen.EventTypeActivityTaskCancelRequested:     &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeActivityTaskCancelRequested)},
		gen.EventTypeRequestCancelActivityTaskFailed: &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeRequestCancelActivityTaskFailed)},
		gen.EventTypeActivityTaskCanceled: &gen.HistoryEvent{
			EventType: common.EventTypePtr(gen.EventTypeActivityTaskCanceled),
			ActivityTaskCanceledEventAttributes: &gen.ActivityTaskCanceledEventAttributes{
				ScheduledEventId: common.Int64Ptr(eventID),
			},
		},
		gen.EventTypeTimerCanceled:     &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeTimerCanceled)},
		gen.EventTypeCancelTimerFailed: &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeCancelTimerFailed)},

		gen.EventTypeMarkerRecorded: &gen.HistoryEvent{
			EventType: common.EventTypePtr(gen.EventTypeMarkerRecorded),
			MarkerRecordedEventAttributes: &gen.MarkerRecordedEventAttributes{
				MarkerName: common.StringPtr(markerName),
			},
		},
		gen.EventTypeWorkflowExecutionSignaled: &gen.HistoryEvent{
			EventType: common.EventTypePtr(gen.EventTypeWorkflowExecutionSignaled),
			WorkflowExecutionSignaledEventAttributes: &gen.WorkflowExecutionSignaledEventAttributes{
				SignalName: common.StringPtr(signalName),
			},
		},

		gen.EventTypeWorkflowExecutionTerminated:                     &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeWorkflowExecutionTerminated)},
		gen.EventTypeWorkflowExecutionCancelRequested:                &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeWorkflowExecutionCancelRequested)},
		gen.EventTypeWorkflowExecutionCanceled:                       &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeWorkflowExecutionCanceled)},
		gen.EventTypeRequestCancelExternalWorkflowExecutionInitiated: &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeRequestCancelExternalWorkflowExecutionInitiated)},
		gen.EventTypeRequestCancelExternalWorkflowExecutionFailed:    &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeRequestCancelExternalWorkflowExecutionFailed)},
		gen.EventTypeExternalWorkflowExecutionCancelRequested:        &gen.HistoryEvent{EventType: common.EventTypePtr(gen.EventTypeExternalWorkflowExecutionCancelRequested)},

		gen.EventTypeStartChildWorkflowExecutionInitiated: &gen.HistoryEvent{
			EventType: common.EventTypePtr(gen.EventTypeStartChildWorkflowExecutionInitiated),
			StartChildWorkflowExecutionInitiatedEventAttributes: &gen.StartChildWorkflowExecutionInitiatedEventAttributes{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr(workflowType),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
		gen.EventTypeStartChildWorkflowExecutionFailed: &gen.HistoryEvent{
			EventType: common.EventTypePtr(gen.EventTypeStartChildWorkflowExecutionFailed),
			StartChildWorkflowExecutionFailedEventAttributes: &gen.StartChildWorkflowExecutionFailedEventAttributes{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr(workflowType),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
		gen.EventTypeChildWorkflowExecutionStarted: &gen.HistoryEvent{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionStarted),
			ChildWorkflowExecutionStartedEventAttributes: &gen.ChildWorkflowExecutionStartedEventAttributes{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr(workflowType),
				},
				WorkflowExecution: &gen.WorkflowExecution{
					WorkflowId: common.StringPtr(workflowID),
				},
			},
		},
		gen.EventTypeChildWorkflowExecutionCompleted: &gen.HistoryEvent{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionCompleted),
			ChildWorkflowExecutionCompletedEventAttributes: &gen.ChildWorkflowExecutionCompletedEventAttributes{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr(workflowType),
				},
				WorkflowExecution: &gen.WorkflowExecution{
					WorkflowId: common.StringPtr(workflowID),
				},
			},
		},
		gen.EventTypeChildWorkflowExecutionFailed: &gen.HistoryEvent{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionFailed),
			ChildWorkflowExecutionFailedEventAttributes: &gen.ChildWorkflowExecutionFailedEventAttributes{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr(workflowType),
				},
				WorkflowExecution: &gen.WorkflowExecution{
					WorkflowId: common.StringPtr(workflowID),
				},
			},
		},
		gen.EventTypeChildWorkflowExecutionCanceled: &gen.HistoryEvent{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionCanceled),
			ChildWorkflowExecutionCanceledEventAttributes: &gen.ChildWorkflowExecutionCanceledEventAttributes{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr(workflowType),
				},
				WorkflowExecution: &gen.WorkflowExecution{
					WorkflowId: common.StringPtr(workflowID),
				},
			},
		},
		gen.EventTypeChildWorkflowExecutionTimedOut: &gen.HistoryEvent{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionTimedOut),
			ChildWorkflowExecutionTimedOutEventAttributes: &gen.ChildWorkflowExecutionTimedOutEventAttributes{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr(workflowType),
				},
				WorkflowExecution: &gen.WorkflowExecution{
					WorkflowId: common.StringPtr(workflowID),
				},
			},
		},
		gen.EventTypeChildWorkflowExecutionTerminated: &gen.HistoryEvent{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionTerminated),
			ChildWorkflowExecutionTerminatedEventAttributes: &gen.ChildWorkflowExecutionTerminatedEventAttributes{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr(workflowType),
				},
				WorkflowExecution: &gen.WorkflowExecution{
					WorkflowId: common.StringPtr(workflowID),
				},
			},
		},
	}
)

type (
	historyEventFilteringSuite struct {
		suite.Suite
	}
)

func TestHistoryEventFilteringSuite(t *testing.T) {
	s := new(historyEventFilteringSuite)
	suite.Run(t, s)
}

func (s *historyEventFilteringSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}

}

func (s *historyEventFilteringSuite) TearDownSuite() {
}

func (s *historyEventFilteringSuite) SetupTest() {
}

func (s *historyEventFilteringSuite) TearDownTest() {
}

func (s *historyEventFilteringSuite) TestIsActivityType() {
	expectedActivityType := map[gen.EventType]bool{
		gen.EventTypeActivityTaskScheduled: true,
		gen.EventTypeActivityTaskStarted:   true,
		gen.EventTypeActivityTaskCompleted: true,
		gen.EventTypeActivityTaskFailed:    true,
		gen.EventTypeActivityTaskTimedOut:  true,
		gen.EventTypeActivityTaskCanceled:  true,
	}

	s.Equal(len(activityEventTypes), len(expectedActivityType))

	for eventType := range expectedActivityType {
		s.True(isActivityType(generateDummyHistoryEvent(eventType)))
	}
}

func (s *historyEventFilteringSuite) TestIsChildWorkflowTypeType() {
	expectedChildWorkflowEventTypes := map[gen.EventType]bool{
		gen.EventTypeStartChildWorkflowExecutionInitiated: true,
		gen.EventTypeStartChildWorkflowExecutionFailed:    true,
		gen.EventTypeChildWorkflowExecutionStarted:        true,
		gen.EventTypeChildWorkflowExecutionCompleted:      true,
		gen.EventTypeChildWorkflowExecutionFailed:         true,
		gen.EventTypeChildWorkflowExecutionTimedOut:       true,
		gen.EventTypeChildWorkflowExecutionTerminated:     true,
		gen.EventTypeChildWorkflowExecutionCanceled:       true,
	}

	s.Equal(len(childWorkflowEventTypes), len(expectedChildWorkflowEventTypes))

	for eventType := range expectedChildWorkflowEventTypes {
		s.True(isChildWorkflowType(generateDummyHistoryEvent(eventType)))
	}
}

func (s *historyEventFilteringSuite) TestIsSignType() {
	expectedSignalEventTypes := map[gen.EventType]bool{
		gen.EventTypeWorkflowExecutionSignaled: true,
	}

	s.Equal(len(signalEventTypes), len(expectedSignalEventTypes))

	for eventType := range expectedSignalEventTypes {
		s.True(isSignType(generateDummyHistoryEvent(eventType)))
	}
}

func (s *historyEventFilteringSuite) TestIsMarkerType() {
	expectedMarkerEventTypes := map[gen.EventType]bool{
		gen.EventTypeMarkerRecorded: true,
	}

	s.Equal(len(markerEventTypes), len(expectedMarkerEventTypes))

	for eventType := range expectedMarkerEventTypes {
		s.True(isMarkerType(generateDummyHistoryEvent(eventType)))
	}
}

func (s *historyEventFilteringSuite) TestIndexHistoryEventFilters() {
	/*
	 * The order of event type is the same as history events defined in shared.thrift HistoryEvent type
	 */
	expectedMapping := map[gen.EventType][]*historyEventFilter{
		gen.EventTypeWorkflowExecutionStarted:                        []*historyEventFilter{generateDummyHistoryEventFilter(0, gen.EventTypeWorkflowExecutionStarted)},
		gen.EventTypeWorkflowExecutionCompleted:                      []*historyEventFilter{generateDummyHistoryEventFilter(1, gen.EventTypeWorkflowExecutionCompleted)},
		gen.EventTypeWorkflowExecutionFailed:                         []*historyEventFilter{generateDummyHistoryEventFilter(2, gen.EventTypeWorkflowExecutionFailed)},
		gen.EventTypeWorkflowExecutionTimedOut:                       []*historyEventFilter{generateDummyHistoryEventFilter(3, gen.EventTypeWorkflowExecutionTimedOut)},
		gen.EventTypeDecisionTaskScheduled:                           []*historyEventFilter{generateDummyHistoryEventFilter(4, gen.EventTypeDecisionTaskScheduled)},
		gen.EventTypeDecisionTaskStarted:                             []*historyEventFilter{generateDummyHistoryEventFilter(5, gen.EventTypeDecisionTaskStarted)},
		gen.EventTypeDecisionTaskCompleted:                           []*historyEventFilter{generateDummyHistoryEventFilter(6, gen.EventTypeDecisionTaskCompleted)},
		gen.EventTypeDecisionTaskTimedOut:                            []*historyEventFilter{generateDummyHistoryEventFilter(7, gen.EventTypeDecisionTaskTimedOut)},
		gen.EventTypeDecisionTaskFailed:                              []*historyEventFilter{generateDummyHistoryEventFilter(8, gen.EventTypeDecisionTaskFailed)},
		gen.EventTypeActivityTaskScheduled:                           []*historyEventFilter{generateDummyHistoryEventFilter(9, gen.EventTypeActivityTaskScheduled)},
		gen.EventTypeActivityTaskStarted:                             []*historyEventFilter{generateDummyHistoryEventFilter(10, gen.EventTypeActivityTaskStarted)},
		gen.EventTypeActivityTaskCompleted:                           []*historyEventFilter{generateDummyHistoryEventFilter(11, gen.EventTypeActivityTaskCompleted)},
		gen.EventTypeActivityTaskFailed:                              []*historyEventFilter{generateDummyHistoryEventFilter(12, gen.EventTypeActivityTaskFailed)},
		gen.EventTypeActivityTaskTimedOut:                            []*historyEventFilter{generateDummyHistoryEventFilter(13, gen.EventTypeActivityTaskTimedOut)},
		gen.EventTypeTimerStarted:                                    []*historyEventFilter{generateDummyHistoryEventFilter(14, gen.EventTypeTimerStarted)},
		gen.EventTypeTimerFired:                                      []*historyEventFilter{generateDummyHistoryEventFilter(15, gen.EventTypeTimerFired)},
		gen.EventTypeActivityTaskCancelRequested:                     []*historyEventFilter{generateDummyHistoryEventFilter(16, gen.EventTypeActivityTaskCancelRequested)},
		gen.EventTypeRequestCancelActivityTaskFailed:                 []*historyEventFilter{generateDummyHistoryEventFilter(17, gen.EventTypeRequestCancelActivityTaskFailed)},
		gen.EventTypeActivityTaskCanceled:                            []*historyEventFilter{generateDummyHistoryEventFilter(18, gen.EventTypeActivityTaskCanceled)},
		gen.EventTypeTimerCanceled:                                   []*historyEventFilter{generateDummyHistoryEventFilter(19, gen.EventTypeTimerCanceled)},
		gen.EventTypeCancelTimerFailed:                               []*historyEventFilter{generateDummyHistoryEventFilter(20, gen.EventTypeCancelTimerFailed)},
		gen.EventTypeMarkerRecorded:                                  []*historyEventFilter{generateDummyHistoryEventFilter(21, gen.EventTypeMarkerRecorded)},
		gen.EventTypeWorkflowExecutionSignaled:                       []*historyEventFilter{generateDummyHistoryEventFilter(22, gen.EventTypeWorkflowExecutionSignaled)},
		gen.EventTypeWorkflowExecutionTerminated:                     []*historyEventFilter{generateDummyHistoryEventFilter(23, gen.EventTypeWorkflowExecutionTerminated)},
		gen.EventTypeWorkflowExecutionCancelRequested:                []*historyEventFilter{generateDummyHistoryEventFilter(24, gen.EventTypeWorkflowExecutionCancelRequested)},
		gen.EventTypeWorkflowExecutionCanceled:                       []*historyEventFilter{generateDummyHistoryEventFilter(25, gen.EventTypeWorkflowExecutionCanceled)},
		gen.EventTypeRequestCancelExternalWorkflowExecutionInitiated: []*historyEventFilter{generateDummyHistoryEventFilter(26, gen.EventTypeRequestCancelExternalWorkflowExecutionInitiated)},
		gen.EventTypeRequestCancelExternalWorkflowExecutionFailed:    []*historyEventFilter{generateDummyHistoryEventFilter(27, gen.EventTypeRequestCancelExternalWorkflowExecutionFailed)},
		gen.EventTypeExternalWorkflowExecutionCancelRequested:        []*historyEventFilter{generateDummyHistoryEventFilter(28, gen.EventTypeExternalWorkflowExecutionCancelRequested)},
		gen.EventTypeStartChildWorkflowExecutionInitiated:            []*historyEventFilter{generateDummyHistoryEventFilter(29, gen.EventTypeStartChildWorkflowExecutionInitiated)},
		gen.EventTypeStartChildWorkflowExecutionFailed:               []*historyEventFilter{generateDummyHistoryEventFilter(30, gen.EventTypeStartChildWorkflowExecutionFailed)},
		gen.EventTypeChildWorkflowExecutionStarted:                   []*historyEventFilter{generateDummyHistoryEventFilter(31, gen.EventTypeChildWorkflowExecutionStarted)},
		gen.EventTypeChildWorkflowExecutionCompleted:                 []*historyEventFilter{generateDummyHistoryEventFilter(32, gen.EventTypeChildWorkflowExecutionCompleted)},
		gen.EventTypeChildWorkflowExecutionFailed:                    []*historyEventFilter{generateDummyHistoryEventFilter(33, gen.EventTypeChildWorkflowExecutionFailed)},
		gen.EventTypeChildWorkflowExecutionCanceled:                  []*historyEventFilter{generateDummyHistoryEventFilter(34, gen.EventTypeChildWorkflowExecutionCanceled)},
		gen.EventTypeChildWorkflowExecutionTimedOut:                  []*historyEventFilter{generateDummyHistoryEventFilter(35, gen.EventTypeChildWorkflowExecutionTimedOut)},
		gen.EventTypeChildWorkflowExecutionTerminated:                []*historyEventFilter{generateDummyHistoryEventFilter(36, gen.EventTypeChildWorkflowExecutionTerminated)},
	}

	inputFilters := make([]*gen.HistoryEventFilter, len(expectedMapping))
	for _, filters := range expectedMapping {
		for _, filter := range filters {
			inputFilters[filter.ID] = filter.HistoryEventFilter
		}
	}

	actualMapping := indexHistoryEventFilters(inputFilters)

	s.Equal(len(expectedMapping), len(actualMapping))

	for index := range expectedMapping {
		// since there is only one item in the slice
		s.Equal(*(expectedMapping[index][0]), *(actualMapping[index][0]))
	}
}

func (s *historyEventFilteringSuite) TestFilterHistoryEvents() {

}

func (s *historyEventFilteringSuite) TestTestActivityHistoryEvent_ActivityTaskScheduled() {
	eventType := gen.EventTypeActivityTaskScheduled
	historyEvent := historyEventMapping[eventType]

	historyEventFilterPass := &historyEventFilter{
		ID: 0,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(eventType),
			ActivityTaskScheduledEventFilter: &gen.ActivityTaskScheduledEventFilter{
				ActivityType: &gen.ActivityType{
					Name: common.StringPtr(activityType),
				},
				ActivityId: common.StringPtr(activityID),
			},
		},
	}

	historyEventFilterFail := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(eventType),
			ActivityTaskScheduledEventFilter: &gen.ActivityTaskScheduledEventFilter{
				ActivityType: &gen.ActivityType{
					Name: common.StringPtr("another activity type"),
				},
				ActivityId: common.StringPtr(activityID),
			},
		},
	}

	token := generateDummyEmptyToken()
	filters := []*historyEventFilter{historyEventFilterPass}
	eventTypeToFilters := map[gen.EventType][]*historyEventFilter{
		eventType: filters,
	}
	s.True(testActivityHistoryEvent(eventTypeToFilters, filters, historyEvent, token))

	token = generateDummyEmptyToken()
	filters = []*historyEventFilter{historyEventFilterFail}
	eventTypeToFilters = map[gen.EventType][]*historyEventFilter{
		eventType: filters,
	}
	s.False(testActivityHistoryEvent(eventTypeToFilters, filters, historyEvent, token))

	// or logic, should pass
	token = generateDummyEmptyToken()
	filters = []*historyEventFilter{historyEventFilterPass, historyEventFilterFail}
	eventTypeToFilters = map[gen.EventType][]*historyEventFilter{
		eventType: filters,
	}
	s.True(testActivityHistoryEvent(eventTypeToFilters, filters, historyEvent, token))
}

func (s *historyEventFilteringSuite) TestTestActivityHistoryEvent_ActivityTaskStarted() {
	eventType := gen.EventTypeActivityTaskStarted
	historyEventFilterPass := &historyEventFilter{
		ID: 0,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(eventType),
			ActivityTaskStartedEventFilter: &gen.ActivityTaskStartedEventFilter{
				ActivityType: &gen.ActivityType{
					Name: common.StringPtr(activityType),
				},
				ActivityId: common.StringPtr(activityID),
			},
		},
	}

	historyEventFilterFail := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(eventType),
			ActivityTaskStartedEventFilter: &gen.ActivityTaskStartedEventFilter{
				ActivityType: &gen.ActivityType{
					Name: common.StringPtr("another activity type"),
				},
				ActivityId: common.StringPtr(activityID),
			},
		},
	}

	doTestActivityHistoryEvent(s, eventType, historyEventFilterPass, historyEventFilterFail)
}

func (s *historyEventFilteringSuite) TestTestActivityHistoryEvent_ActivityTaskCompleted() {
	eventType := gen.EventTypeActivityTaskCompleted
	historyEventFilterPass := &historyEventFilter{
		ID: 0,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(eventType),
			ActivityTaskCompletedEventFilter: &gen.ActivityTaskCompletedEventFilter{
				ActivityType: &gen.ActivityType{
					Name: common.StringPtr(activityType),
				},
				ActivityId: common.StringPtr(activityID),
			},
		},
	}

	historyEventFilterFail := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(eventType),
			ActivityTaskCompletedEventFilter: &gen.ActivityTaskCompletedEventFilter{
				ActivityType: &gen.ActivityType{
					Name: common.StringPtr("another activity type"),
				},
				ActivityId: common.StringPtr(activityID),
			},
		},
	}

	doTestActivityHistoryEvent(s, eventType, historyEventFilterPass, historyEventFilterFail)
}

func (s *historyEventFilteringSuite) TestTestActivityHistoryEvent_ActivityTaskFailed() {
	eventType := gen.EventTypeActivityTaskFailed
	historyEventFilterPass := &historyEventFilter{
		ID: 0,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(eventType),
			ActivityTaskFailedEventFilter: &gen.ActivityTaskFailedEventFilter{
				ActivityType: &gen.ActivityType{
					Name: common.StringPtr(activityType),
				},
				ActivityId: common.StringPtr(activityID),
			},
		},
	}

	historyEventFilterFail := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(eventType),
			ActivityTaskFailedEventFilter: &gen.ActivityTaskFailedEventFilter{
				ActivityType: &gen.ActivityType{
					Name: common.StringPtr("another activity type"),
				},
				ActivityId: common.StringPtr(activityID),
			},
		},
	}

	doTestActivityHistoryEvent(s, eventType, historyEventFilterPass, historyEventFilterFail)
}

func (s *historyEventFilteringSuite) TestTestActivityHistoryEvent_ActivityTaskTimedOut() {
	eventType := gen.EventTypeActivityTaskTimedOut
	historyEventFilterPass := &historyEventFilter{
		ID: 0,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(eventType),
			ActivityTaskTimedOutEventFilter: &gen.ActivityTaskTimedOutEventFilter{
				ActivityType: &gen.ActivityType{
					Name: common.StringPtr(activityType),
				},
				ActivityId: common.StringPtr(activityID),
			},
		},
	}

	historyEventFilterFail := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(eventType),
			ActivityTaskTimedOutEventFilter: &gen.ActivityTaskTimedOutEventFilter{
				ActivityType: &gen.ActivityType{
					Name: common.StringPtr("another activity type"),
				},
				ActivityId: common.StringPtr(activityID),
			},
		},
	}

	doTestActivityHistoryEvent(s, eventType, historyEventFilterPass, historyEventFilterFail)
}

func (s *historyEventFilteringSuite) TestTestActivityHistoryEvent_ActivityTaskCanceled() {
	eventType := gen.EventTypeActivityTaskCanceled
	historyEventFilterPass := &historyEventFilter{
		ID: 0,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(eventType),
			ActivityTaskCanceledEventFilter: &gen.ActivityTaskCanceledEventFilter{
				ActivityType: &gen.ActivityType{
					Name: common.StringPtr(activityType),
				},
				ActivityId: common.StringPtr(activityID),
			},
		},
	}

	historyEventFilterFail := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(eventType),
			ActivityTaskCanceledEventFilter: &gen.ActivityTaskCanceledEventFilter{
				ActivityType: &gen.ActivityType{
					Name: common.StringPtr("another activity type"),
				},
				ActivityId: common.StringPtr(activityID),
			},
		},
	}

	doTestActivityHistoryEvent(s, eventType, historyEventFilterPass, historyEventFilterFail)
}

func doTestActivityHistoryEvent(s *historyEventFilteringSuite, eventType gen.EventType,
	historyEventFilterPass *historyEventFilter, historyEventFilterFail *historyEventFilter) {

	scheduleHistoryEvent := historyEventMapping[gen.EventTypeActivityTaskScheduled]
	testHistoryEvent := historyEventMapping[eventType]

	token := generateDummyEmptyToken()
	filters := []*historyEventFilter{historyEventFilterPass}
	eventTypeToFilters := map[gen.EventType][]*historyEventFilter{eventType: filters}
	testActivityHistoryEvent(eventTypeToFilters, []*historyEventFilter{}, scheduleHistoryEvent, token)
	s.True(testActivityHistoryEvent(eventTypeToFilters, filters, testHistoryEvent, token))

	token = generateDummyEmptyToken()
	filters = []*historyEventFilter{historyEventFilterFail}
	eventTypeToFilters = map[gen.EventType][]*historyEventFilter{eventType: filters}
	testActivityHistoryEvent(eventTypeToFilters, []*historyEventFilter{}, scheduleHistoryEvent, token)
	s.False(testActivityHistoryEvent(eventTypeToFilters, filters, testHistoryEvent, token))

	// or logic, should pass
	token = generateDummyEmptyToken()
	filters = []*historyEventFilter{historyEventFilterPass, historyEventFilterFail}
	eventTypeToFilters = map[gen.EventType][]*historyEventFilter{eventType: filters}
	testActivityHistoryEvent(eventTypeToFilters, []*historyEventFilter{}, scheduleHistoryEvent, token)
	s.True(testActivityHistoryEvent(eventTypeToFilters, filters, testHistoryEvent, token))
}

func (s *historyEventFilteringSuite) TestTestChildWorkflowHistoryEvent_StartChildWorkflowExecutionInitiated() {
	historyEvent := historyEventMapping[gen.EventTypeStartChildWorkflowExecutionInitiated]

	historyEventFilterPass := &historyEventFilter{
		ID: 0,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeStartChildWorkflowExecutionInitiated),
			StartChildWorkflowExecutionInitiatedEventFilter: &gen.StartChildWorkflowExecutionInitiatedEventFilter{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr(workflowType),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
	}
	s.True(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterPass}, historyEvent))

	historyEventFilterFail := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeStartChildWorkflowExecutionInitiated),
			StartChildWorkflowExecutionInitiatedEventFilter: &gen.StartChildWorkflowExecutionInitiatedEventFilter{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr("another workflow type"),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
	}
	s.False(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterFail}, historyEvent))

	// or logic, should pass
	s.True(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterPass, historyEventFilterFail}, historyEvent))
}

func (s *historyEventFilteringSuite) TestTestChildWorkflowHistoryEvent_StartChildWorkflowExecutionFailed() {
	historyEvent := historyEventMapping[gen.EventTypeStartChildWorkflowExecutionFailed]

	historyEventFilterPass := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeStartChildWorkflowExecutionFailed),
			StartChildWorkflowExecutionFailedEventFilter: &gen.StartChildWorkflowExecutionFailedEventFilter{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr(workflowType),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
	}
	s.True(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterPass}, historyEvent))

	historyEventFilterFail := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeStartChildWorkflowExecutionFailed),
			StartChildWorkflowExecutionFailedEventFilter: &gen.StartChildWorkflowExecutionFailedEventFilter{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr("another workflow type"),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
	}
	s.False(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterFail}, historyEvent))

	// or logic, should pass
	s.True(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterPass, historyEventFilterFail}, historyEvent))
}

func (s *historyEventFilteringSuite) TestTestChildWorkflowHistoryEvent_ChildWorkflowExecutionStarted() {
	historyEvent := historyEventMapping[gen.EventTypeChildWorkflowExecutionStarted]

	historyEventFilterPass := &historyEventFilter{
		ID: 0,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionStarted),
			ChildWorkflowExecutionStartedEventFilter: &gen.ChildWorkflowExecutionStartedEventFilter{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr(workflowType),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
	}
	s.True(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterPass}, historyEvent))

	historyEventFilterFail := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionStarted),
			ChildWorkflowExecutionStartedEventFilter: &gen.ChildWorkflowExecutionStartedEventFilter{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr("another workflow type"),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
	}
	s.False(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterFail}, historyEvent))

	// or logic, should pass
	s.True(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterPass, historyEventFilterFail}, historyEvent))
}

func (s *historyEventFilteringSuite) TestTestChildWorkflowHistoryEvent_ChildWorkflowExecutionCompleted() {
	historyEvent := historyEventMapping[gen.EventTypeChildWorkflowExecutionCompleted]

	historyEventFilterPass := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionCompleted),
			ChildWorkflowExecutionCompletedEventFilter: &gen.ChildWorkflowExecutionCompletedEventFilter{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr(workflowType),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
	}
	s.True(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterPass}, historyEvent))

	historyEventFilterFail := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionCompleted),
			ChildWorkflowExecutionCompletedEventFilter: &gen.ChildWorkflowExecutionCompletedEventFilter{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr("another workflow type"),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
	}
	s.False(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterFail}, historyEvent))

	// or logic, should pass
	s.True(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterPass, historyEventFilterFail}, historyEvent))
}

func (s *historyEventFilteringSuite) TestTestChildWorkflowHistoryEvent_ChildWorkflowExecutionFailed() {
	historyEvent := historyEventMapping[gen.EventTypeChildWorkflowExecutionFailed]

	historyEventFilterPass := &historyEventFilter{
		ID: 0,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionFailed),
			ChildWorkflowExecutionFailedEventFilter: &gen.ChildWorkflowExecutionFailedEventFilter{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr(workflowType),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
	}
	s.True(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterPass}, historyEvent))

	historyEventFilterFail := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionFailed),
			ChildWorkflowExecutionFailedEventFilter: &gen.ChildWorkflowExecutionFailedEventFilter{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr("another workflow type"),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
	}
	s.False(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterFail}, historyEvent))

	// or logic, should pass
	s.True(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterPass, historyEventFilterFail}, historyEvent))
}

func (s *historyEventFilteringSuite) TestTestChildWorkflowHistoryEvent_ChildWorkflowExecutionTimedOut() {
	historyEvent := historyEventMapping[gen.EventTypeChildWorkflowExecutionTimedOut]

	historyEventFilterPass := &historyEventFilter{
		ID: 0,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionTimedOut),
			ChildWorkflowExecutionTimedOutEventFilter: &gen.ChildWorkflowExecutionTimedOutEventFilter{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr(workflowType),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
	}
	s.True(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterPass}, historyEvent))

	historyEventFilterFail := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionTimedOut),
			ChildWorkflowExecutionTimedOutEventFilter: &gen.ChildWorkflowExecutionTimedOutEventFilter{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr("another workflow type"),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
	}
	s.False(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterFail}, historyEvent))

	// or logic, should pass
	s.True(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterPass, historyEventFilterFail}, historyEvent))
}

func (s *historyEventFilteringSuite) TestTestChildWorkflowHistoryEvent_ChildWorkflowExecutionTerminated() {
	historyEvent := historyEventMapping[gen.EventTypeChildWorkflowExecutionTerminated]

	historyEventFilterPass := &historyEventFilter{
		ID: 0,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionTerminated),
			ChildWorkflowExecutionTerminatedEventFilter: &gen.ChildWorkflowExecutionTerminatedEventFilter{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr(workflowType),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
	}
	s.True(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterPass}, historyEvent))

	historyEventFilterFail := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionTerminated),
			ChildWorkflowExecutionTerminatedEventFilter: &gen.ChildWorkflowExecutionTerminatedEventFilter{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr("another workflow type"),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
	}
	s.False(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterFail}, historyEvent))

	// or logic, should pass
	s.True(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterPass, historyEventFilterFail}, historyEvent))
}

func (s *historyEventFilteringSuite) TestTestChildWorkflowHistoryEvent_ChildWorkflowExecutionCanceled() {
	historyEvent := historyEventMapping[gen.EventTypeChildWorkflowExecutionCanceled]

	historyEventFilterPass := &historyEventFilter{
		ID: 0,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionCanceled),
			ChildWorkflowExecutionCanceledEventFilter: &gen.ChildWorkflowExecutionCanceledEventFilter{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr(workflowType),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
	}
	s.True(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterPass}, historyEvent))

	historyEventFilterFail := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeChildWorkflowExecutionCanceled),
			ChildWorkflowExecutionCanceledEventFilter: &gen.ChildWorkflowExecutionCanceledEventFilter{
				WorkflowType: &gen.WorkflowType{
					Name: common.StringPtr("another workflow type"),
				},
				WorkflowId: common.StringPtr(workflowID),
			},
		},
	}
	s.False(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterFail}, historyEvent))

	// or logic, should pass
	s.True(testChildWorkflowHistoryEvent([]*historyEventFilter{historyEventFilterPass, historyEventFilterFail}, historyEvent))
}

func (s *historyEventFilteringSuite) TestTestSignalHistoryEvent() {
	historyEvent := historyEventMapping[gen.EventTypeWorkflowExecutionSignaled]

	historyEventFilterPass := &historyEventFilter{
		ID: 0,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeWorkflowExecutionSignaled),
			WorkflowExecutionSignaledEventFilter: &gen.WorkflowExecutionSignaledEventFilter{
				SignalName: common.StringPtr(signalName),
			},
		},
	}
	s.True(testSignalHistoryEvent([]*historyEventFilter{historyEventFilterPass}, historyEvent))

	historyEventFilterFail := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeWorkflowExecutionSignaled),
			WorkflowExecutionSignaledEventFilter: &gen.WorkflowExecutionSignaledEventFilter{
				SignalName: common.StringPtr("some other dummy signal"),
			},
		},
	}
	s.False(testSignalHistoryEvent([]*historyEventFilter{historyEventFilterFail}, historyEvent))

	// or logic, should pass
	s.True(testSignalHistoryEvent([]*historyEventFilter{historyEventFilterPass, historyEventFilterFail}, historyEvent))
}

func (s *historyEventFilteringSuite) TestTestMarkerHistoryEvent() {
	historyEvent := historyEventMapping[gen.EventTypeMarkerRecorded]

	historyEventFilterPass := &historyEventFilter{
		ID: 0,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeMarkerRecorded),
			MarkerRecordedEventFilter: &gen.MarkerRecordedEventFilter{
				MarkerName: common.StringPtr(markerName),
			},
		},
	}
	s.True(testMarkerHistoryEvent([]*historyEventFilter{historyEventFilterPass}, historyEvent))

	historyEventFilterFail := &historyEventFilter{
		ID: 1,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(gen.EventTypeMarkerRecorded),
			MarkerRecordedEventFilter: &gen.MarkerRecordedEventFilter{
				MarkerName: common.StringPtr("some other dummy marker"),
			},
		},
	}
	s.False(testMarkerHistoryEvent([]*historyEventFilter{historyEventFilterFail}, historyEvent))

	// or logic, should pass
	s.True(testMarkerHistoryEvent([]*historyEventFilter{historyEventFilterPass, historyEventFilterFail}, historyEvent))
}

func generateDummyHistoryEvent(eventType gen.EventType) *gen.HistoryEvent {
	return &gen.HistoryEvent{
		EventType: common.EventTypePtr(eventType),
	}
}

func generateDummyHistoryEventFilter(ID int, eventType gen.EventType) *historyEventFilter {
	return &historyEventFilter{
		ID: ID,
		HistoryEventFilter: &gen.HistoryEventFilter{
			EventType: common.EventTypePtr(eventType),
		},
	}
}

func generateDummyEmptyToken() *getHistoryContinuationToken {
	return &getHistoryContinuationToken{
		FilterIndexToEventIDs: make(map[int]map[int64]bool),
	}
}
