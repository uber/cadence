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
	"github.com/pborman/uuid"
	"github.com/uber-common/bark"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
)

type (
	stateBuilder struct {
		shard       ShardContext
		msBuilder   *mutableStateBuilder
		domainCache cache.DomainCache
		logger      bark.Logger
	}
)

func newStateBuilder(shard ShardContext, msBuilder *mutableStateBuilder, logger bark.Logger) *stateBuilder {

	return &stateBuilder{
		shard:       shard,
		msBuilder:   msBuilder,
		domainCache: shard.GetDomainCache(),
		logger:      logger,
	}
}

func (b *stateBuilder) applyEvents(version int64, sourceClusterName string, domainID, requestID string,
	execution shared.WorkflowExecution, history *shared.History, newRunHistory *shared.History) (*shared.HistoryEvent,
	*decisionInfo, *mutableStateBuilder, error) {
	var lastEvent *shared.HistoryEvent
	var lastDecision *decisionInfo
	var newRunStateBuilder *mutableStateBuilder
	for _, event := range history.Events {
		lastEvent = event
		switch event.GetEventType() {
		case shared.EventTypeWorkflowExecutionStarted:
			attributes := event.WorkflowExecutionStartedEventAttributes
			var parentDomainID *string
			if attributes.ParentWorkflowDomain != nil {
				parentDomainEntry, err := b.domainCache.GetDomain(attributes.GetParentWorkflowDomain())
				if err != nil {
					return nil, nil, nil, err
				}
				parentDomainID = &parentDomainEntry.GetInfo().ID
			}
			b.msBuilder.ReplicateWorkflowExecutionStartedEvent(domainID, parentDomainID, execution, requestID, attributes)

		case shared.EventTypeDecisionTaskScheduled:
			attributes := event.DecisionTaskScheduledEventAttributes
			lastDecision = b.msBuilder.ReplicateDecisionTaskScheduledEvent(event.GetVersion(), event.GetEventId(),
				attributes.TaskList.GetName(), attributes.GetStartToCloseTimeoutSeconds())

		case shared.EventTypeDecisionTaskStarted:
			attributes := event.DecisionTaskStartedEventAttributes
			lastDecision = b.msBuilder.ReplicateDecisionTaskStartedEvent(nil, event.GetVersion(), attributes.GetScheduledEventId(), event.GetEventId(),
				attributes.GetRequestId(), event.GetTimestamp())

		case shared.EventTypeDecisionTaskCompleted:
			attributes := event.DecisionTaskCompletedEventAttributes
			b.msBuilder.ReplicateDecisionTaskCompletedEvent(attributes.GetScheduledEventId(),
				attributes.GetStartedEventId())

		case shared.EventTypeDecisionTaskTimedOut:
			attributes := event.DecisionTaskTimedOutEventAttributes
			b.msBuilder.ReplicateDecisionTaskTimedOutEvent(attributes.GetScheduledEventId(),
				attributes.GetStartedEventId())

		case shared.EventTypeDecisionTaskFailed:
			attributes := event.DecisionTaskFailedEventAttributes
			b.msBuilder.ReplicateDecisionTaskFailedEvent(attributes.GetScheduledEventId(),
				attributes.GetStartedEventId())

		case shared.EventTypeActivityTaskScheduled:
			b.msBuilder.ReplicateActivityTaskScheduledEvent(event)

		case shared.EventTypeActivityTaskStarted:
			b.msBuilder.ReplicateActivityTaskStartedEvent(event)

		case shared.EventTypeActivityTaskCompleted:
			if err := b.msBuilder.ReplicateActivityTaskCompletedEvent(event); err != nil {
				return nil, nil, nil, err
			}

		case shared.EventTypeActivityTaskFailed:
			b.msBuilder.ReplicateActivityTaskFailedEvent(event)

		case shared.EventTypeActivityTaskTimedOut:
			b.msBuilder.ReplicateActivityTaskTimedOutEvent(event)

		case shared.EventTypeActivityTaskCancelRequested:
			b.msBuilder.ReplicateActivityTaskCancelRequestedEvent(event)

		case shared.EventTypeActivityTaskCanceled:
			b.msBuilder.ReplicateActivityTaskCanceledEvent(event)

		case shared.EventTypeRequestCancelActivityTaskFailed:
			// No mutable state action is needed

		case shared.EventTypeTimerStarted:
			b.msBuilder.ReplicateTimerStartedEvent(event)

		case shared.EventTypeTimerFired:
			b.msBuilder.ReplicateTimerFiredEvent(event)

		case shared.EventTypeTimerCanceled:
			b.msBuilder.ReplicateTimerCanceledEvent(event)

		case shared.EventTypeCancelTimerFailed:
			// No mutable state action is needed

		case shared.EventTypeStartChildWorkflowExecutionInitiated:
			// Create a new request ID which is used by transfer queue processor if domain is failed over at this point
			createRequestID := uuid.New()
			b.msBuilder.ReplicateStartChildWorkflowExecutionInitiatedEvent(event, createRequestID)

		case shared.EventTypeStartChildWorkflowExecutionFailed:
			b.msBuilder.ReplicateStartChildWorkflowExecutionFailedEvent(event)

		case shared.EventTypeChildWorkflowExecutionStarted:
			b.msBuilder.ReplicateChildWorkflowExecutionStartedEvent(event)

		case shared.EventTypeChildWorkflowExecutionCompleted:
			b.msBuilder.ReplicateChildWorkflowExecutionCompletedEvent(event)

		case shared.EventTypeChildWorkflowExecutionFailed:
			b.msBuilder.ReplicateChildWorkflowExecutionFailedEvent(event)

		case shared.EventTypeChildWorkflowExecutionCanceled:
			b.msBuilder.ReplicateChildWorkflowExecutionCanceledEvent(event)

		case shared.EventTypeChildWorkflowExecutionTimedOut:
			b.msBuilder.ReplicateChildWorkflowExecutionTimedOutEvent(event)

		case shared.EventTypeChildWorkflowExecutionTerminated:
			b.msBuilder.ReplicateChildWorkflowExecutionTerminatedEvent(event)

		case shared.EventTypeRequestCancelExternalWorkflowExecutionInitiated:
			// Create a new request ID which is used by transfer queue processor if domain is failed over at this point
			cancelRequestID := uuid.New()
			b.msBuilder.ReplicateRequestCancelExternalWorkflowExecutionInitiatedEvent(event, cancelRequestID)

		case shared.EventTypeRequestCancelExternalWorkflowExecutionFailed:
			b.msBuilder.ReplicateRequestCancelExternalWorkflowExecutionFailedEvent(event)

		case shared.EventTypeExternalWorkflowExecutionCancelRequested:
			b.msBuilder.ReplicateExternalWorkflowExecutionCancelRequested(event)

		case shared.EventTypeSignalExternalWorkflowExecutionInitiated:
			// Create a new request ID which is used by transfer queue processor if domain is failed over at this point
			signalRequestID := uuid.New()
			b.msBuilder.ReplicateSignalExternalWorkflowExecutionInitiatedEvent(event, signalRequestID)

		case shared.EventTypeSignalExternalWorkflowExecutionFailed:
			b.msBuilder.ReplicateSignalExternalWorkflowExecutionFailedEvent(event)

		case shared.EventTypeExternalWorkflowExecutionSignaled:
			b.msBuilder.ReplicateExternalWorkflowExecutionSignaled(event)

		case shared.EventTypeMarkerRecorded:
			// No mutable state action is needed

		case shared.EventTypeWorkflowExecutionSignaled:
			// No mutable state action is needed

		case shared.EventTypeWorkflowExecutionCancelRequested:
			b.msBuilder.ReplicateWorkflowExecutionCancelRequestedEvent(event)

		case shared.EventTypeWorkflowExecutionCompleted:
			b.msBuilder.ReplicateWorkflowExecutionCompletedEvent(event)

		case shared.EventTypeWorkflowExecutionFailed:
			b.msBuilder.ReplicateWorkflowExecutionFailedEvent(event)

		case shared.EventTypeWorkflowExecutionTimedOut:
			b.msBuilder.ReplicateWorkflowExecutionTimedoutEvent(event)

		case shared.EventTypeWorkflowExecutionCanceled:
			b.msBuilder.ReplicateWorkflowExecutionCanceledEvent(event)

		case shared.EventTypeWorkflowExecutionTerminated:
			b.msBuilder.ReplicateWorkflowExecutionTerminatedEvent(event)

		case shared.EventTypeWorkflowExecutionContinuedAsNew:
			// ContinuedAsNew event also has history for first 2 events for next run as they are created transactionally
			startedEvent := newRunHistory.Events[0]
			startedAttributes := startedEvent.WorkflowExecutionStartedEventAttributes
			dtScheduledEvent := newRunHistory.Events[1]
			domainEntry, err := b.domainCache.GetDomainByID(domainID)
			if err != nil {
				return nil, nil, nil, err
			}
			domainName := domainEntry.GetInfo().Name

			// History event only have the parentDomainName.  Lookup the domain ID from cache
			var parentDomainID *string
			if startedAttributes.ParentWorkflowDomain != nil {
				parentDomainEntry, err := b.domainCache.GetDomain(startedAttributes.GetParentWorkflowDomain())
				if err != nil {
					return nil, nil, nil, err
				}
				parentDomainID = &parentDomainEntry.GetInfo().ID
			}

			newRunID := event.WorkflowExecutionContinuedAsNewEventAttributes.GetNewExecutionRunId()

			newExecution := shared.WorkflowExecution{
				WorkflowId: execution.WorkflowId,
				RunId:      common.StringPtr(newRunID),
			}

			// Create mutable state updates for the new run
			newRunStateBuilder = newMutableStateBuilderWithReplicationState(b.shard.GetConfig(), b.logger, version)
			newRunStateBuilder.ReplicateWorkflowExecutionStartedEvent(domainID, parentDomainID, newExecution, uuid.New(),
				startedAttributes)
			di := newRunStateBuilder.ReplicateDecisionTaskScheduledEvent(
				dtScheduledEvent.GetVersion(),
				dtScheduledEvent.GetEventId(),
				dtScheduledEvent.DecisionTaskScheduledEventAttributes.TaskList.GetName(),
				dtScheduledEvent.DecisionTaskScheduledEventAttributes.GetStartToCloseTimeoutSeconds(),
			)
			nextEventID := di.ScheduleID + 1
			newRunStateBuilder.executionInfo.NextEventID = nextEventID
			newRunStateBuilder.executionInfo.LastFirstEventID = startedEvent.GetEventId()
			// Set the history from replication task on the newStateBuilder
			newRunStateBuilder.hBuilder = newHistoryBuilderFromEvents(newRunHistory.Events, b.logger)

			b.msBuilder.ReplicateWorkflowExecutionContinuedAsNewEvent(sourceClusterName, domainID, domainName, event,
				startedEvent, di, newRunStateBuilder)

		}
	}

	return lastEvent, lastDecision, newRunStateBuilder, nil
}
