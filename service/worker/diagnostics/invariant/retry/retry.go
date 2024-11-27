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

package retry

import (
	"context"
	"github.com/uber/cadence/service/worker/diagnostics/invariant/failure"

	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/worker/diagnostics/invariant"
)

// Retry is an invariant that will be used to identify the issues regarding retries in the workflow execution history
type Retry invariant.Invariant

type retry struct {
	workflowExecutionHistory *types.GetWorkflowExecutionHistoryResponse
}

type Params struct {
	WorkflowExecutionHistory *types.GetWorkflowExecutionHistoryResponse
}

func NewInvariant(p Params) Retry {
	return &retry{
		workflowExecutionHistory: p.WorkflowExecutionHistory,
	}
}

func (r *retry) Check(context.Context) ([]invariant.InvariantCheckResult, error) {
	result := make([]invariant.InvariantCheckResult, 0)
	events := r.workflowExecutionHistory.GetHistory().GetEvents()
	lastEvent := fetchContinuedAsNewEvent(events)
	startedEvent := fetchStartedEvent(events)
	if lastEvent != nil && startedEvent != nil {
		if startedEvent.RetryPolicy != nil {
			result = append(result, invariant.InvariantCheckResult{
				InvariantType: WorkflowRetry.String(),
				Reason:        failure.ErrorTypeFromReason(*lastEvent.FailureReason).String(),
				Metadata: invariant.MarshalData(RetryMetadata{
					RetryPolicy: startedEvent.RetryPolicy,
					Attempt:     startedEvent.Attempt,
				}),
			})
		}
	}
	return result, nil
}

func fetchContinuedAsNewEvent(events []*types.HistoryEvent) *types.WorkflowExecutionContinuedAsNewEventAttributes {
	for _, event := range events {
		if event.GetWorkflowExecutionContinuedAsNewEventAttributes() != nil {
			return event.GetWorkflowExecutionContinuedAsNewEventAttributes()
		}
	}
	return nil
}

func fetchStartedEvent(events []*types.HistoryEvent) *types.WorkflowExecutionStartedEventAttributes {
	for _, event := range events {
		if event.GetWorkflowExecutionStartedEventAttributes() != nil {
			return event.GetWorkflowExecutionStartedEventAttributes()
		}
	}
	return nil
}

func (r *retry) RootCause(ctx context.Context, issues []invariant.InvariantCheckResult) ([]invariant.InvariantRootCauseResult, error) {
	// Not implemented since this invariant does not have any root cause.
	// Issue identified in Check() is a workflow retry which is essentially handled by failure invariant
	result := make([]invariant.InvariantRootCauseResult, 0)
	return result, nil
}
