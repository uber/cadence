// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package invariants

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
)

func reasonForDecisionTaskTimeouts(event *types.HistoryEvent, allEvents []*types.HistoryEvent) (string, []byte) {
	eventScheduledID := event.GetDecisionTaskTimedOutEventAttributes().GetScheduledEventID()
	attr := event.GetDecisionTaskTimedOutEventAttributes()
	cause := attr.GetCause()
	switch cause {
	case types.DecisionTaskTimedOutCauseTimeout:
		return attr.TimeoutType.String(), timeoutLimitInBytes(getDecisionTaskConfiguredTimeout(eventScheduledID, allEvents))
	case types.DecisionTaskTimedOutCauseReset:
		newRunID := attr.GetNewRunID()
		return attr.Reason, []byte(newRunID)
	default:
		return "valid cause not available for decision task timeout", nil
	}
}

func getWorkflowExecutionConfiguredTimeout(events []*types.HistoryEvent) int32 {
	for _, event := range events {
		if event.ID == 1 { // event 1 is workflow execution started event
			return event.GetWorkflowExecutionStartedEventAttributes().GetExecutionStartToCloseTimeoutSeconds()
		}
	}
	return 0
}

func getActivityTaskConfiguredTimeout(e *types.HistoryEvent, events []*types.HistoryEvent) (int32, error) {
	eventScheduledID := e.GetActivityTaskTimedOutEventAttributes().GetScheduledEventID()
	timeoutType := e.GetActivityTaskTimedOutEventAttributes().GetTimeoutType()
	for _, event := range events {
		if event.ID == eventScheduledID {
			attr := event.GetActivityTaskScheduledEventAttributes()
			switch timeoutType {
			case types.TimeoutTypeHeartbeat:
				return attr.GetHeartbeatTimeoutSeconds(), nil
			case types.TimeoutTypeScheduleToClose:
				return attr.GetScheduleToCloseTimeoutSeconds(), nil
			case types.TimeoutTypeScheduleToStart:
				return attr.GetScheduleToStartTimeoutSeconds(), nil
			case types.TimeoutTypeStartToClose:
				return attr.GetStartToCloseTimeoutSeconds(), nil
			default:
				return 0, fmt.Errorf("unknown timeout type")
			}
		}
	}
	return 0, fmt.Errorf("activity scheduled event not found")
}

func getDecisionTaskConfiguredTimeout(eventScheduledID int64, events []*types.HistoryEvent) int32 {
	for _, event := range events {
		if event.ID == eventScheduledID {
			return event.GetDecisionTaskScheduledEventAttributes().GetStartToCloseTimeoutSeconds()
		}
	}
	return 0
}

func getChildWorkflowExecutionConfiguredTimeout(e *types.HistoryEvent, events []*types.HistoryEvent) int32 {
	wfInitiatedID := e.GetChildWorkflowExecutionTimedOutEventAttributes().GetInitiatedEventID()
	for _, event := range events {
		if event.ID == wfInitiatedID {
			return event.GetStartChildWorkflowExecutionInitiatedEventAttributes().GetExecutionStartToCloseTimeoutSeconds()
		}
	}
	return 0
}

func timeoutLimitInBytes(val int32) []byte {
	valInBytes, _ := json.Marshal(val)
	return valInBytes
}

func getExecutionTime(startID, timeoutID int64, events []*types.HistoryEvent) time.Duration {
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].ID < events[j].ID
	})

	firstEvent := events[startID-1]
	lastEvent := events[timeoutID-1]
	return time.Unix(0, common.Int64Default(lastEvent.Timestamp)).Sub(time.Unix(0, common.Int64Default(firstEvent.Timestamp)))
}

func marshalData(rc any) []byte {
	data, _ := json.Marshal(rc)
	return data
}
