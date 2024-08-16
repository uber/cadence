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
	"context"
	"time"

	"github.com/uber/cadence/common/types"
)

type Timeout Invariant

type timeout struct {
	workflowExecutionHistory *types.GetWorkflowExecutionHistoryResponse
}

func NewTimeout(wfHistory *types.GetWorkflowExecutionHistoryResponse) Invariant {
	return &timeout{
		workflowExecutionHistory: wfHistory,
	}
}

func (t *timeout) Check(context.Context) ([]InvariantCheckResult, error) {
	result := make([]InvariantCheckResult, 0)
	events := t.workflowExecutionHistory.GetHistory().GetEvents()
	for _, event := range events {
		if event.WorkflowExecutionTimedOutEventAttributes != nil {
			timeoutLimit := getWorkflowExecutionConfiguredTimeout(events)
			data := ExecutionTimeoutMetadata{
				ExecutionTime:     getExecutionTime(1, event.ID, events),
				ConfiguredTimeout: time.Duration(timeoutLimit) * time.Second,
				LastOngoingEvent:  events[len(events)-2],
			}
			result = append(result, InvariantCheckResult{
				InvariantType: TimeoutTypeExecution.String(),
				Reason:        event.GetWorkflowExecutionTimedOutEventAttributes().GetTimeoutType().String(),
				Metadata:      marshalData(data),
			})
		}
		if event.ActivityTaskTimedOutEventAttributes != nil {
			metadata, err := getActivityTaskMetadata(event, events)
			if err != nil {
				return nil, err
			}
			result = append(result, InvariantCheckResult{
				InvariantType: TimeoutTypeActivity.String(),
				Reason:        event.GetActivityTaskTimedOutEventAttributes().GetTimeoutType().String(),
				Metadata:      marshalData(metadata),
			})
		}
		if event.DecisionTaskTimedOutEventAttributes != nil {
			reason, metadata := reasonForDecisionTaskTimeouts(event, events)
			result = append(result, InvariantCheckResult{
				InvariantType: TimeoutTypeDecision.String(),
				Reason:        reason,
				Metadata:      metadata,
			})
		}
		if event.ChildWorkflowExecutionTimedOutEventAttributes != nil {
			timeoutLimit := getChildWorkflowExecutionConfiguredTimeout(event, events)
			data := ChildWfTimeoutMetadata{
				ExecutionTime:     getExecutionTime(event.GetChildWorkflowExecutionTimedOutEventAttributes().StartedEventID, event.ID, events),
				ConfiguredTimeout: time.Duration(timeoutLimit) * time.Second,
				Execution:         event.GetChildWorkflowExecutionTimedOutEventAttributes().WorkflowExecution,
			}
			result = append(result, InvariantCheckResult{
				InvariantType: TimeoutTypeChildWorkflow.String(),
				Reason:        event.GetChildWorkflowExecutionTimedOutEventAttributes().TimeoutType.String(),
				Metadata:      marshalData(data),
			})
		}
	}
	return result, nil
}
