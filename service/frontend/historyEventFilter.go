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
	gen "github.com/uber/cadence/.gen/go/shared"
)

var signalEventTypes map[gen.EventType]bool
var markerEventTypes map[gen.EventType]bool
var activityEventTypes map[gen.EventType]bool
var childWorkflowEventTypes map[gen.EventType]bool

func init() {
	signalEventTypes = map[gen.EventType]bool{
		gen.EventTypeWorkflowExecutionSignaled: true,
	}

	markerEventTypes = map[gen.EventType]bool{
		gen.EventTypeMarkerRecorded: true,
	}

	activityEventTypes = map[gen.EventType]bool{
		gen.EventTypeActivityTaskScheduled: true,
		gen.EventTypeActivityTaskStarted:   true,
		gen.EventTypeActivityTaskCompleted: true,
		gen.EventTypeActivityTaskFailed:    true,
		gen.EventTypeActivityTaskTimedOut:  true,
		gen.EventTypeActivityTaskCanceled:  true,
	}

	childWorkflowEventTypes = map[gen.EventType]bool{
		gen.EventTypeStartChildWorkflowExecutionInitiated: true,
		gen.EventTypeStartChildWorkflowExecutionFailed:    true,
		gen.EventTypeChildWorkflowExecutionStarted:        true,
		gen.EventTypeChildWorkflowExecutionCompleted:      true,
		gen.EventTypeChildWorkflowExecutionFailed:         true,
		gen.EventTypeChildWorkflowExecutionTimedOut:       true,
		gen.EventTypeChildWorkflowExecutionTerminated:     true,
		gen.EventTypeChildWorkflowExecutionCanceled:       true,
	}
}

type (
	activityInfo struct {
		ActivityType string
		ActivityID   string
		EventID      int64
	}

	childWorkflowInfo struct {
		WorkflowType string
		WorkflowID   string
	}

	historyEventFilter struct {
		ID int
		*gen.HistoryEventFilter
	}
)

// FilterHistoryEvents filter on history events for desired results
func FilterHistoryEvents(inputFilters []*gen.HistoryEventFilter, events []*gen.HistoryEvent, token *getHistoryContinuationToken) []*gen.HistoryEvent {

	eventTypeToFilters := indexHistoryEventFilters(inputFilters)
	results := []*gen.HistoryEvent{}
	if token.FilterIndexToEventIDs == nil {
		token.FilterIndexToEventIDs = make(map[int]map[int64]bool)
	}

	for index := range events {
		event := events[index]
		filters, ok := eventTypeToFilters[event.GetEventType()]
		if !ok {
			// meaning no filters found, not interested
			continue
		}

		if isActivityType(event) {
			if testActivityHistoryEvent(eventTypeToFilters, filters, event, token) {
				results = append(results, event)
			}
		} else if isChildWorkflowType(event) {
			if testChildWorkflowHistoryEvent(filters, event) {
				results = append(results, event)
			}
		} else if isSignType(event) {
			if testSignalHistoryEvent(filters, event) {
				results = append(results, event)
			}
		} else if isMarkerType(event) {
			if testMarkerHistoryEvent(filters, event) {
				results = append(results, event)
			}
		} else {
			// event types remaining do NOT contain detailed information used for filtering
			// so as long as we found corresponding filter(s), the event is interested.
			results = append(results, event)
		}
	}

	return results
}

func isActivityType(event *gen.HistoryEvent) bool {
	_, ok := activityEventTypes[event.GetEventType()]
	return ok
}

func isChildWorkflowType(event *gen.HistoryEvent) bool {
	_, ok := childWorkflowEventTypes[event.GetEventType()]
	return ok
}

func isSignType(event *gen.HistoryEvent) bool {
	_, ok := signalEventTypes[event.GetEventType()]
	return ok
}

func isMarkerType(event *gen.HistoryEvent) bool {
	_, ok := markerEventTypes[event.GetEventType()]
	return ok
}

func indexHistoryEventFilters(filters []*gen.HistoryEventFilter) map[gen.EventType][]*historyEventFilter {

	mapping := make(map[gen.EventType][]*historyEventFilter)

	for index := range filters {
		filter := &historyEventFilter{
			ID:                 index,
			HistoryEventFilter: filters[index],
		}

		values := mapping[filter.GetEventType()]
		values = append(values, filter)
		mapping[filter.GetEventType()] = values
	}

	return mapping
}

func testActivityHistoryEvent(eventTypeToFilters map[gen.EventType][]*historyEventFilter, filters []*historyEventFilter, event *gen.HistoryEvent, token *getHistoryContinuationToken) bool {
	var filterActivityType *string
	var filterActicityID *string

	var eventID int64

	getActivityType := func(activityType *gen.ActivityType) *string {
		if activityType == nil {
			return nil
		}

		return activityType.Name
	}

	switch event.GetEventType() {
	case gen.EventTypeActivityTaskScheduled:
		// find all relevant filters which has corresponding activity ID && activity type
		// register the filter index -> schedule Event ID (which is used in all activity events)
		eventActivityType := event.ActivityTaskScheduledEventAttributes.ActivityType.GetName()
		eventActivityID := event.ActivityTaskScheduledEventAttributes.GetActivityId()
		putActivityInfo(token, eventTypeToFilters, eventActivityType, eventActivityID, event.GetEventId())
	}

	for _, filter := range filters {
		switch event.GetEventType() {
		case gen.EventTypeActivityTaskScheduled:
			if filter.ActivityTaskScheduledEventFilter != nil {
				filterActivityType = getActivityType(filter.ActivityTaskScheduledEventFilter.ActivityType)
				filterActicityID = filter.ActivityTaskScheduledEventFilter.ActivityId
			}
			eventID = event.GetEventId()

		case gen.EventTypeActivityTaskStarted:
			if filter.ActivityTaskStartedEventFilter != nil {
				filterActivityType = getActivityType(filter.ActivityTaskStartedEventFilter.ActivityType)
				filterActicityID = filter.ActivityTaskStartedEventFilter.ActivityId
			}
			eventID = event.ActivityTaskStartedEventAttributes.GetScheduledEventId()

		case gen.EventTypeActivityTaskCompleted:
			if filter.ActivityTaskCompletedEventFilter != nil {
				filterActivityType = getActivityType(filter.ActivityTaskCompletedEventFilter.ActivityType)
				filterActicityID = filter.ActivityTaskCompletedEventFilter.ActivityId
			}
			eventID = event.ActivityTaskCompletedEventAttributes.GetScheduledEventId()

		case gen.EventTypeActivityTaskFailed:
			if filter.ActivityTaskFailedEventFilter != nil {
				filterActivityType = getActivityType(filter.ActivityTaskFailedEventFilter.ActivityType)
				filterActicityID = filter.ActivityTaskFailedEventFilter.ActivityId
			}
			eventID = event.ActivityTaskFailedEventAttributes.GetScheduledEventId()

		case gen.EventTypeActivityTaskTimedOut:
			if filter.ActivityTaskTimedOutEventFilter != nil {
				filterActivityType = getActivityType(filter.ActivityTaskTimedOutEventFilter.ActivityType)
				filterActicityID = filter.ActivityTaskTimedOutEventFilter.ActivityId
			}
			eventID = event.ActivityTaskTimedOutEventAttributes.GetScheduledEventId()

		case gen.EventTypeActivityTaskCanceled:
			if filter.ActivityTaskCanceledEventFilter != nil {
				filterActivityType = getActivityType(filter.ActivityTaskCanceledEventFilter.ActivityType)
				filterActicityID = filter.ActivityTaskCanceledEventFilter.ActivityId
			}
			eventID = event.ActivityTaskCanceledEventAttributes.GetScheduledEventId()
		}

		if filterActivityType == nil && filterActicityID == nil {
			// filter is essentially empty
			return true
		}
	}

	return testActivityInfo(token, filters, eventID)
}

func testChildWorkflowHistoryEvent(filters []*historyEventFilter, event *gen.HistoryEvent) bool {
	var filterChildWokflowType *string
	var filterChildWokflowID *string
	var eventChildWokflowType string
	var eventChildWokflowID string

	getWorkflowType := func(workflowType *gen.WorkflowType) *string {
		if workflowType == nil {
			return nil
		}

		return workflowType.Name
	}

	for _, filter := range filters {

		switch filter.GetEventType() {
		case gen.EventTypeStartChildWorkflowExecutionInitiated:
			if filter.StartChildWorkflowExecutionInitiatedEventFilter != nil {
				filterChildWokflowType = getWorkflowType(filter.StartChildWorkflowExecutionInitiatedEventFilter.WorkflowType)
				filterChildWokflowID = filter.StartChildWorkflowExecutionInitiatedEventFilter.WorkflowId
			}

			eventChildWokflowType = event.StartChildWorkflowExecutionInitiatedEventAttributes.WorkflowType.GetName()
			eventChildWokflowID = event.StartChildWorkflowExecutionInitiatedEventAttributes.GetWorkflowId()

		case gen.EventTypeStartChildWorkflowExecutionFailed:
			if filter.StartChildWorkflowExecutionFailedEventFilter != nil {
				filterChildWokflowType = getWorkflowType(filter.StartChildWorkflowExecutionFailedEventFilter.WorkflowType)
				filterChildWokflowID = filter.StartChildWorkflowExecutionFailedEventFilter.WorkflowId
			}

			eventChildWokflowType = event.StartChildWorkflowExecutionFailedEventAttributes.WorkflowType.GetName()
			eventChildWokflowID = event.StartChildWorkflowExecutionFailedEventAttributes.GetWorkflowId()

		case gen.EventTypeChildWorkflowExecutionStarted:
			if filter.ChildWorkflowExecutionStartedEventFilter != nil {
				filterChildWokflowType = getWorkflowType(filter.ChildWorkflowExecutionStartedEventFilter.WorkflowType)
				filterChildWokflowID = filter.ChildWorkflowExecutionStartedEventFilter.WorkflowId
			}

			eventChildWokflowType = event.ChildWorkflowExecutionStartedEventAttributes.WorkflowType.GetName()
			eventChildWokflowID = event.ChildWorkflowExecutionStartedEventAttributes.WorkflowExecution.GetWorkflowId()

		case gen.EventTypeChildWorkflowExecutionCompleted:
			if filter.ChildWorkflowExecutionCompletedEventFilter != nil {
				filterChildWokflowType = getWorkflowType(filter.ChildWorkflowExecutionCompletedEventFilter.WorkflowType)
				filterChildWokflowID = filter.ChildWorkflowExecutionCompletedEventFilter.WorkflowId
			}

			eventChildWokflowType = event.ChildWorkflowExecutionCompletedEventAttributes.WorkflowType.GetName()
			eventChildWokflowID = event.ChildWorkflowExecutionCompletedEventAttributes.WorkflowExecution.GetWorkflowId()

		case gen.EventTypeChildWorkflowExecutionFailed:
			if filter.ChildWorkflowExecutionFailedEventFilter != nil {
				filterChildWokflowType = getWorkflowType(filter.ChildWorkflowExecutionFailedEventFilter.WorkflowType)
				filterChildWokflowID = filter.ChildWorkflowExecutionFailedEventFilter.WorkflowId
			}

			eventChildWokflowType = event.ChildWorkflowExecutionFailedEventAttributes.WorkflowType.GetName()
			eventChildWokflowID = event.ChildWorkflowExecutionFailedEventAttributes.WorkflowExecution.GetWorkflowId()

		case gen.EventTypeChildWorkflowExecutionTimedOut:
			if filter.ChildWorkflowExecutionTimedOutEventFilter != nil {
				filterChildWokflowType = getWorkflowType(filter.ChildWorkflowExecutionTimedOutEventFilter.WorkflowType)
				filterChildWokflowID = filter.ChildWorkflowExecutionTimedOutEventFilter.WorkflowId
			}

			eventChildWokflowType = event.ChildWorkflowExecutionTimedOutEventAttributes.WorkflowType.GetName()
			eventChildWokflowID = event.ChildWorkflowExecutionTimedOutEventAttributes.WorkflowExecution.GetWorkflowId()

		case gen.EventTypeChildWorkflowExecutionTerminated:
			if filter.ChildWorkflowExecutionTerminatedEventFilter != nil {
				filterChildWokflowType = getWorkflowType(filter.ChildWorkflowExecutionTerminatedEventFilter.WorkflowType)
				filterChildWokflowID = filter.ChildWorkflowExecutionTerminatedEventFilter.WorkflowId
			}

			eventChildWokflowType = event.ChildWorkflowExecutionTerminatedEventAttributes.WorkflowType.GetName()
			eventChildWokflowID = event.ChildWorkflowExecutionTerminatedEventAttributes.WorkflowExecution.GetWorkflowId()

		case gen.EventTypeChildWorkflowExecutionCanceled:
			if filter.ChildWorkflowExecutionCanceledEventFilter != nil {
				filterChildWokflowType = getWorkflowType(filter.ChildWorkflowExecutionCanceledEventFilter.WorkflowType)
				filterChildWokflowID = filter.ChildWorkflowExecutionCanceledEventFilter.WorkflowId
			}

			eventChildWokflowType = event.ChildWorkflowExecutionCanceledEventAttributes.WorkflowType.GetName()
			eventChildWokflowID = event.ChildWorkflowExecutionCanceledEventAttributes.WorkflowExecution.GetWorkflowId()
		}

		if filterChildWokflowType != nil {
			if eventChildWokflowType != *filterChildWokflowType {
				continue
			}
		}
		if filterChildWokflowID != nil {
			if eventChildWokflowID != *filterChildWokflowID {
				continue
			}
		}

		return true
	}

	return false
}

func testSignalHistoryEvent(filters []*historyEventFilter, event *gen.HistoryEvent) bool {

	var signalName *string

	for index := range filters {
		filter := filters[index]

		switch filter.GetEventType() {
		case gen.EventTypeWorkflowExecutionSignaled:
			signalName = filter.WorkflowExecutionSignaledEventFilter.SignalName
		}

		if signalName != nil {
			if event.WorkflowExecutionSignaledEventAttributes.GetSignalName() != *signalName {
				continue
			}
		}

		return true
	}

	return false
}

func testMarkerHistoryEvent(filters []*historyEventFilter, event *gen.HistoryEvent) bool {

	var markerName *string

	for index := range filters {
		filter := filters[index]

		switch filter.GetEventType() {
		case gen.EventTypeMarkerRecorded:
			markerName = filter.MarkerRecordedEventFilter.MarkerName
		}

		if markerName != nil {
			if event.MarkerRecordedEventAttributes.GetMarkerName() != *markerName {
				continue
			}
		}

		return true
	}

	return false
}

func putActivityInfo(token *getHistoryContinuationToken, eventTypeToFilters map[gen.EventType][]*historyEventFilter,
	eventActivityType string, eventActivityID string, eventID int64) {

	var filterActivityType *string
	var filterActicityID *string

	getActivityType := func(activityType *gen.ActivityType) *string {
		if activityType == nil {
			return nil
		}

		return activityType.Name
	}

	for eventType := range activityEventTypes {
		for _, filter := range eventTypeToFilters[eventType] {
			switch filter.GetEventType() {
			case gen.EventTypeActivityTaskScheduled:
				if filter.ActivityTaskScheduledEventFilter != nil {
					filterActivityType = getActivityType(filter.ActivityTaskScheduledEventFilter.ActivityType)
					filterActicityID = filter.ActivityTaskScheduledEventFilter.ActivityId
				}

			case gen.EventTypeActivityTaskStarted:
				if filter.ActivityTaskStartedEventFilter != nil {
					filterActivityType = getActivityType(filter.ActivityTaskStartedEventFilter.ActivityType)
					filterActicityID = filter.ActivityTaskStartedEventFilter.ActivityId
				}

			case gen.EventTypeActivityTaskCompleted:
				if filter.ActivityTaskCompletedEventFilter != nil {
					filterActivityType = getActivityType(filter.ActivityTaskCompletedEventFilter.ActivityType)
					filterActicityID = filter.ActivityTaskCompletedEventFilter.ActivityId
				}

			case gen.EventTypeActivityTaskFailed:
				if filter.ActivityTaskFailedEventFilter != nil {
					filterActivityType = getActivityType(filter.ActivityTaskFailedEventFilter.ActivityType)
					filterActicityID = filter.ActivityTaskFailedEventFilter.ActivityId
				}

			case gen.EventTypeActivityTaskTimedOut:
				if filter.ActivityTaskTimedOutEventFilter != nil {
					filterActivityType = getActivityType(filter.ActivityTaskTimedOutEventFilter.ActivityType)
					filterActicityID = filter.ActivityTaskTimedOutEventFilter.ActivityId
				}

			case gen.EventTypeActivityTaskCanceled:
				if filter.ActivityTaskCanceledEventFilter != nil {
					filterActivityType = getActivityType(filter.ActivityTaskCanceledEventFilter.ActivityType)
					filterActicityID = filter.ActivityTaskCanceledEventFilter.ActivityId
				}
			}

			if filterActivityType != nil {
				if eventActivityType != *filterActivityType {
					continue
				}
			}
			if filterActicityID != nil {
				if eventActivityID != *filterActicityID {
					continue
				}
			}

			if filterActivityType != nil || filterActicityID != nil {
				// meaning there is at least one attribute set, either activity type or activity ID
				// and the attribute(s) set in the filter match given activity info
				// we need to store this relation to token
				eventIDs, ok := token.FilterIndexToEventIDs[filter.ID]
				if !ok {
					eventIDs = make(map[int64]bool)
				}
				eventIDs[eventID] = true
				token.FilterIndexToEventIDs[filter.ID] = eventIDs
			}
		}
	}
}

func testActivityInfo(token *getHistoryContinuationToken, filters []*historyEventFilter, eventID int64) bool {

	result := false

	for _, filter := range filters {
		eventIDs, ok := token.FilterIndexToEventIDs[filter.ID]
		if !ok {
			continue
		}

		for id := range eventIDs {
			if eventID == id {
				delete(eventIDs, id)
				if len(eventIDs) == 0 {
					delete(token.FilterIndexToEventIDs, filter.ID)
				}
				result = true
			}
		}
	}

	return result
}
