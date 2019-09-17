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

//go:generate mockgen -copyright_file ../../LICENSE -package $GOPACKAGE -source $GOFILE -destination newNDCEventsReapplication_mock.go

package history

import (
	ctx "context"
	"fmt"

	"github.com/pborman/uuid"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
)

type (
	nDCEventsReapplication struct {
		timeSource      clock.TimeSource
		historyCache    *historyCache
		domainCache     cache.DomainCache
		clusterMetadata cluster.Metadata
		resetor         workflowResetor
		metricsClient   metrics.Client
		logger          log.Logger
	}
)

func newNDCEventsReapplication(
	timeSource clock.TimeSource,
	resetor workflowResetor,
	historyCache *historyCache,
	domainCache cache.DomainCache,
	metricsClient metrics.Client,
	logger log.Logger,
	clusterMetadata cluster.Metadata,
) *nDCEventsReapplication {

	return &nDCEventsReapplication{
		timeSource:      timeSource,
		historyCache:    historyCache,
		domainCache:     domainCache,
		metricsClient:   metricsClient,
		logger:          logger,
		clusterMetadata: clusterMetadata,
		resetor:         resetor,
	}
}
func (r *nDCEventsReapplication) reapplyEvents(
	ctx ctx.Context,
	domainID string,
	workflowID string,
	historyEvents []*shared.HistoryEvent,
) (retError error) {
	reapplyEvents := []*shared.HistoryEvent{}
	for _, event := range historyEvents {
		switch event.GetEventType() {
		case shared.EventTypeWorkflowExecutionSignaled:
			reapplyEvents = append(reapplyEvents, event)
		}
	}

	if len(reapplyEvents) == 0 {
		return nil
	}
	context, msBuilder, releaseFunc, err := r.getCurrentWorkflowMutableState(ctx, domainID, workflowID)
	defer func() { releaseFunc(retError) }()
	if err != nil {
		// for get workflow execution context, with valid run id
		// err will not be of type EntityNotExistsError
		return err
	}

	if msBuilder.IsWorkflowExecutionRunning() {
		return r.reapplyEventsToCurrentRunningWorkflow(ctx, context, msBuilder, reapplyEvents)
	}
	return r.reapplyEventsToCurrentClosedWorkflow(ctx, context, msBuilder, reapplyEvents)
}

func (r *nDCEventsReapplication) reapplyEventsToCurrentRunningWorkflow(
	ctx ctx.Context,
	context workflowExecutionContext,
	msBuilder mutableState,
	events []*shared.HistoryEvent,
) error {

	canMutateWorkflow, err := r.prepareWorkflowMutation(msBuilder)
	if err != nil || !canMutateWorkflow {
		return err
	}

	numSignals := 0
	for _, event := range events {
		switch event.GetEventType() {
		case shared.EventTypeWorkflowExecutionSignaled:
			attr := event.WorkflowExecutionSignaledEventAttributes
			if _, err := msBuilder.AddWorkflowExecutionSignaled(
				attr.GetSignalName(),
				attr.Input,
				attr.GetIdentity()); err != nil {
				return ErrWorkflowMutationSignal
			}
			numSignals++

		default:
			return ErrUnreappliableEvent
		}
	}

	r.logger.Info(fmt.Sprintf("reapplying %v signals", numSignals))
	now := r.timeSource.Now()
	return context.updateWorkflowExecutionAsActive(now)
}

func (r *nDCEventsReapplication) reapplyEventsToCurrentClosedWorkflow(
	ctx ctx.Context,
	context workflowExecutionContext,
	msBuilder mutableState,
	events []*shared.HistoryEvent,
) (retError error) {

	domainID := msBuilder.GetExecutionInfo().DomainID
	workflowID := msBuilder.GetExecutionInfo().WorkflowID

	domainEntry, err := r.domainCache.GetDomainByID(domainID)
	if err != nil {
		return err
	}

	resetRequestID := uuid.New()
	// workflow event buffer guarantee that the event immediately
	// after the decision task started is decision task finished event
	lastDecisionTaskStartEventID := msBuilder.GetPreviousStartedEventID()
	if lastDecisionTaskStartEventID == common.EmptyEventID {
		// TODO when https://github.com/uber/cadence/issues/2420 is finished
		//  reset to workflow finish event
		errStr := "cannot reapply signal due to workflow missing decision"
		r.logger.Error(errStr)
		return &shared.BadRequestError{Message: errStr}

	}

	resetDecisionFinishID := lastDecisionTaskStartEventID + 1

	baseContext := context
	baseMutableState := msBuilder
	currContext := context
	currMutableState := msBuilder
	_, err = r.resetor.ResetWorkflowExecution(
		ctx,
		&shared.ResetWorkflowExecutionRequest{
			Domain:                common.StringPtr(domainEntry.GetInfo().Name),
			WorkflowExecution:     context.getExecution(),
			Reason:                common.StringPtr(workflowResetReason),
			DecisionFinishEventId: common.Int64Ptr(resetDecisionFinishID),
			RequestId:             common.StringPtr(resetRequestID),
		},
		baseContext,
		baseMutableState,
		currContext,
		currMutableState,
	)
	if err != nil {
		if _, ok := err.(*shared.DomainNotActiveError); ok {
			return nil
		}
		return err
	}

	resetNewContext, resetNewMsBuilder, resetNewRelease, err := r.getCurrentWorkflowMutableState(ctx, domainID, workflowID)
	defer func() { resetNewRelease(retError) }()
	if err != nil {
		return err
	}
	if resetNewMsBuilder.IsWorkflowExecutionRunning() {
		return ErrRetryRaceCondition
	}

	return r.reapplyEventsToCurrentRunningWorkflow(ctx, resetNewContext, resetNewMsBuilder, events)
}

func (r *nDCEventsReapplication) getCurrentWorkflowMutableState(
	ctx ctx.Context,
	domainID string,
	workflowID string,
) (workflowExecutionContext, mutableState, releaseWorkflowExecutionFunc, error) {
	// we need to check the current workflow execution
	context, release, err := r.historyCache.getOrCreateWorkflowExecution(ctx,
		domainID,
		// only use the workflow ID, to get the current running one
		shared.WorkflowExecution{WorkflowId: common.StringPtr(workflowID)},
	)
	if err != nil {
		return nil, nil, nil, err
	}

	msBuilder, err := context.loadWorkflowExecution()
	if err != nil {
		// no matter what error happen, we need to retry
		release(err)
		return nil, nil, nil, err
	}
	return context, msBuilder, release, nil
}

func (r *nDCEventsReapplication) prepareWorkflowMutation(
	msBuilder mutableState,
) (bool, error) {

	// for replication stack to modify workflow re-applying events
	// we need to check 2 things
	// 1. if the workflow's last write version indicates that workflow is active here
	// 2. if the domain entry says this domain is active and failover version in the domain entry >= workflow's last write version
	// if either of the above is true, then the workflow can be mutated

	lastWriteVersion, err := msBuilder.GetLastWriteVersion()
	if err != nil {
		return false, err
	}
	lastWriteVersionActive := r.clusterMetadata.ClusterNameForFailoverVersion(lastWriteVersion) == r.clusterMetadata.GetCurrentClusterName()
	if lastWriteVersionActive {
		msBuilder.UpdateReplicationStateVersion(lastWriteVersion, true)
		return true, nil
	}

	domainEntry, err := r.domainCache.GetDomainByID(msBuilder.GetExecutionInfo().DomainID)
	if err != nil {
		return false, err
	}

	domainFailoverVersion := domainEntry.GetFailoverVersion()
	domainActive := domainEntry.GetReplicationConfig().ActiveClusterName == r.clusterMetadata.GetCurrentClusterName() &&
		domainFailoverVersion >= lastWriteVersion

	if domainActive {
		msBuilder.UpdateReplicationStateVersion(domainFailoverVersion, true)
		return true, nil
	}
	return false, nil
}
