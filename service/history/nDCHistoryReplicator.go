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

package history

import (
	ctx "context"
	"time"

	"github.com/pborman/uuid"

	h "github.com/uber/cadence/.gen/go/history"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
)

type (
	nDCBranchMgrProvider func(
		context workflowExecutionContext,
		mutableState mutableState,
		logger log.Logger,
	) nDCBranchMgr

	nDCConflictResolverProvider func(
		context workflowExecutionContext,
		mutableState mutableState,
		logger log.Logger,
	) nDCConflictResolver

	nDCWorkflowResetterProvider func(
		domainID string,
		workflowID string,
		baseRunID string,
		newContext workflowExecutionContext,
		newRunID string,
		logger log.Logger,
	) nDCWorkflowResetter

	nDCHistoryReplicator struct {
		shard           ShardContext
		clusterMetadata cluster.Metadata
		historyV2Mgr    persistence.HistoryV2Manager
		metricsClient   metrics.Client
		domainCache     cache.DomainCache
		historyCache    *historyCache
		transactionMgr  nDCTransactionMgr
		logger          log.Logger

		newBranchMgr        nDCBranchMgrProvider
		newConflictResolver nDCConflictResolverProvider
		newWorkflowResetter nDCWorkflowResetterProvider
		newStateBuilder     stateBuilderProvider
		newMutableState     mutableStateProvider
	}
)

func newNDCHistoryReplicator(
	shard ShardContext,
	historyCache *historyCache,
	logger log.Logger,
) *nDCHistoryReplicator {

	transactionMgr := newNDCTransactionMgr(shard, historyCache, logger)
	replicator := &nDCHistoryReplicator{
		shard:           shard,
		clusterMetadata: shard.GetService().GetClusterMetadata(),
		historyV2Mgr:    shard.GetHistoryV2Manager(),
		metricsClient:   shard.GetMetricsClient(),
		domainCache:     shard.GetDomainCache(),
		historyCache:    historyCache,
		transactionMgr:  transactionMgr,
		logger:          logger.WithTags(tag.ComponentHistoryReplicator),

		newBranchMgr: func(
			context workflowExecutionContext,
			mutableState mutableState,
			logger log.Logger,
		) nDCBranchMgr {
			return newNDCBranchMgr(shard, context, mutableState, logger)
		},
		newConflictResolver: func(
			context workflowExecutionContext,
			mutableState mutableState,
			logger log.Logger,
		) nDCConflictResolver {
			return newNDCConflictResolver(shard, context, mutableState, logger)
		},
		newWorkflowResetter: func(
			domainID string,
			workflowID string,
			baseRunID string,
			newContext workflowExecutionContext,
			newRunID string,
			logger log.Logger,
		) nDCWorkflowResetter {
			return newNDCWorkflowResetter(shard, transactionMgr, domainID, workflowID, baseRunID, newContext, newRunID, logger)
		},
		newStateBuilder: func(
			msBuilder mutableState,
			logger log.Logger,
		) stateBuilder {
			return newStateBuilder(shard, msBuilder, logger)
		},
		newMutableState: func(
			version int64,
			domainName string,
			logger log.Logger,
		) mutableState {
			return newMutableStateBuilderWithVersionHistories(
				shard,
				shard.GetEventsCache(),
				logger,
				version,
				// if can see replication task, meaning that domain is
				// global domain with > 1 target clusters
				cache.ReplicationPolicyMultiCluster,
				domainName,
			)
		},
	}

	return replicator
}

func (r *nDCHistoryReplicator) ApplyEvents(
	ctx ctx.Context,
	request *h.ReplicateEventsRequest,
) (retError error) {

	startTime := time.Now()
	task, err := newNDCReplicationTask(r.clusterMetadata, startTime, r.logger, request)
	if err != nil {
		return err
	}

	return r.applyEvents(ctx, task)
}

func (r *nDCHistoryReplicator) applyEvents(
	ctx ctx.Context,
	task nDCReplicationTask,
) (retError error) {

	context, releaseFn, err := r.historyCache.getOrCreateWorkflowExecution(
		ctx,
		task.getDomainID(),
		*task.getExecution(),
	)
	if err != nil {
		// for get workflow execution context, with valid run id
		// err will not be of type EntityNotExistsError
		return err
	}
	defer func() { releaseFn(retError) }()

	switch task.getFirstEvent().GetEventType() {
	case shared.EventTypeWorkflowExecutionStarted:
		return r.applyStartEvents(ctx, context, releaseFn, task)

	default:
		// apply events, other than simple start workflow execution
		// the continue as new + start workflow execution combination will also be processed here
		mutableState, err := context.loadWorkflowExecution()
		switch err.(type) {
		case nil:
			branchIndex, err := r.applyNonStartEventsPrepareBranch(ctx, context, mutableState, task)
			if err != nil {
				return err
			}

			doContinue, err := r.applyNonStartEventsPrepareReorder(ctx, context, mutableState, branchIndex, task)
			if err != nil || !doContinue {
				return err
			}

			mutableState, isRebuilt, err := r.applyNonStartEventsPrepareMutableState(ctx, context, mutableState, branchIndex, task)
			if err != nil {
				return err
			}

			if mutableState.GetVersionHistories().GetCurrentVersionHistoryIndex() == branchIndex {
				return r.applyNonStartEventsToCurrentBranch(ctx, context, mutableState, isRebuilt, releaseFn, task)
			}
			return r.applyNonStartEventsToNoneCurrentBranch(ctx, context, mutableState, branchIndex, releaseFn, task)

		case *shared.EntityNotExistsError:
			// mutable state not created, check if is workflow reset
			mutableState, completeFn, err := r.applyNonStartEventsMissingMutableState(ctx, context, task)
			if err != nil {
				return err
			}

			return r.applyNonStartEventsResetWorkflow(ctx, context, mutableState, completeFn, task)

		default:
			// unable to get mutable state, return err so we can retry the task later
			return err
		}
	}
}

func (r *nDCHistoryReplicator) applyStartEvents(
	ctx ctx.Context,
	context workflowExecutionContext,
	releaseFn releaseWorkflowExecutionFunc,
	task nDCReplicationTask,
) (retError error) {

	requestID := uuid.New() // requestID used for start workflow execution request.  This is not on the history event.
	mutableState := r.newMutableState(task.getVersion(), context.getDomainName(), task.getLogger())
	stateBuilder := r.newStateBuilder(mutableState, task.getLogger())

	// use state builder for workflow mutable state mutation
	_, _, _, err := stateBuilder.applyEvents(
		task.getDomainID(),
		requestID,
		*task.getExecution(),
		task.getEvents(),
		task.getNewEvents(),
		nDCMutableStateEventStoreVersion,
		nDCMutableStateEventStoreVersion,
	)
	if err != nil {
		return err
	}

	err = r.transactionMgr.createWorkflow(
		ctx,
		task.getEventTime(),
		newNDCWorkflow(
			ctx,
			r.domainCache,
			r.clusterMetadata,
			context,
			mutableState,
			releaseFn,
		),
	)
	if err == nil {
		r.notify(task.getSourceCluster(), task.getEventTime())
	}
	return err
}

func (r *nDCHistoryReplicator) applyNonStartEventsPrepareBranch(
	ctx ctx.Context,
	context workflowExecutionContext,
	mutableState mutableState,
	task nDCReplicationTask,
) (int, error) {

	incomingVersionHistory := task.getVersionHistory()
	branchMgr := r.newBranchMgr(context, mutableState, task.getLogger())
	versionHistoryIndex, err := branchMgr.prepareVersionHistory(
		ctx,
		incomingVersionHistory,
	)
	if err != nil {
		return 0, err
	}
	return versionHistoryIndex, nil

}

func (r *nDCHistoryReplicator) applyNonStartEventsPrepareReorder(
	ctx ctx.Context,
	context workflowExecutionContext,
	mutableState mutableState,
	branchIndex int,
	task nDCReplicationTask,
) (bool, error) {

	versionHistories := mutableState.GetVersionHistories()
	versionHistory, err := versionHistories.GetVersionHistory(branchIndex)
	if err != nil {
		return false, err
	}
	lastVersionHistoryItem, err := versionHistory.GetLastItem()
	if err != nil {
		return false, err
	}
	nextEventID := lastVersionHistoryItem.GetEventID() + 1

	if task.getFirstEvent().GetEventId() < nextEventID {
		// duplicate replication task
		r.metricsClient.IncCounter(metrics.ReplicateHistoryEventsScope, metrics.DuplicateReplicationEventsCounter)
		return false, nil
	}
	if task.getFirstEvent().GetEventId() > nextEventID {
		// TODO we should use a new retry error for 3+DC
		return false, newRetryTaskErrorWithHint(
			ErrRetryBufferEventsMsg,
			task.getDomainID(),
			task.getWorkflowID(),
			task.getRunID(),
			lastVersionHistoryItem.GetEventID()+1,
		)
	}
	// task.getFirstEvent().GetEventId() == nextEventID
	return true, nil
}

func (r *nDCHistoryReplicator) applyNonStartEventsPrepareMutableState(
	ctx ctx.Context,
	context workflowExecutionContext,
	mutableState mutableState,
	branchIndex int,
	task nDCReplicationTask,
) (mutableState, bool, error) {

	incomingVersion := task.getVersion()
	conflictResolver := r.newConflictResolver(context, mutableState, task.getLogger())
	return conflictResolver.prepareMutableState(
		ctx,
		branchIndex,
		incomingVersion,
	)
}

func (r *nDCHistoryReplicator) applyNonStartEventsToCurrentBranch(
	ctx ctx.Context,
	context workflowExecutionContext,
	mutableState mutableState,
	isRebuilt bool,
	releaseFn releaseWorkflowExecutionFunc,
	task nDCReplicationTask,
) error {

	requestID := uuid.New() // requestID used for start workflow execution request.  This is not on the history event.
	stateBuilder := r.newStateBuilder(mutableState, task.getLogger())
	_, _, newMutableState, err := stateBuilder.applyEvents(
		task.getDomainID(),
		requestID,
		*task.getExecution(),
		task.getEvents(),
		task.getNewEvents(),
		nDCMutableStateEventStoreVersion,
		nDCMutableStateEventStoreVersion,
	)
	if err != nil {
		return err
	}

	targetWorkflow := newNDCWorkflow(
		ctx,
		r.domainCache,
		r.clusterMetadata,
		context,
		mutableState,
		releaseFn,
	)

	var newWorkflow nDCWorkflow
	if newMutableState != nil {
		newExecutionInfo := newMutableState.GetExecutionInfo()
		newContext := newWorkflowExecutionContext(
			newExecutionInfo.DomainID,
			shared.WorkflowExecution{
				WorkflowId: common.StringPtr(newExecutionInfo.WorkflowID),
				RunId:      common.StringPtr(newExecutionInfo.RunID),
			},
			r.shard,
			r.shard.GetExecutionManager(),
			r.logger,
		)

		newWorkflow = newNDCWorkflow(
			ctx,
			r.domainCache,
			r.clusterMetadata,
			newContext,
			newMutableState,
			noopReleaseFn,
		)
	}

	err = r.transactionMgr.updateWorkflow(
		ctx,
		task.getEventTime(),
		isRebuilt,
		targetWorkflow,
		newWorkflow,
	)
	if err == nil {
		r.notify(task.getSourceCluster(), task.getEventTime())
	}
	return err
}

func (r *nDCHistoryReplicator) applyNonStartEventsToNoneCurrentBranch(
	ctx ctx.Context,
	context workflowExecutionContext,
	mutableState mutableState,
	branchIndex int,
	releaseFn releaseWorkflowExecutionFunc,
	task nDCReplicationTask,
) error {

	// workflow backfill to non current branch
	// if encounter backfill with continue as new
	// first, create the new workflow as zombie
	if len(task.getNewEvents()) != 0 {
		startTime := time.Now()
		newTask, err := task.generateNewRunTask(startTime)
		if err != nil {
			return err
		}
		if err := r.applyEvents(ctx, newTask); err != nil {
			return err
		}
	}

	versionHistoryItem := persistence.NewVersionHistoryItem(
		task.getLastEvent().GetEventId(),
		task.getLastEvent().GetVersion(),
	)
	versionHistory, err := mutableState.GetVersionHistories().GetVersionHistory(branchIndex)
	if err != nil {
		return err
	}
	if err = versionHistory.AddOrUpdateItem(versionHistoryItem); err != nil {
		return err
	}

	return r.transactionMgr.backfillWorkflow(
		ctx,
		task.getEventTime(),
		newNDCWorkflow(
			ctx,
			r.domainCache,
			r.clusterMetadata,
			context,
			mutableState,
			releaseFn,
		),
		&persistence.WorkflowEvents{
			DomainID:    task.getDomainID(),
			WorkflowID:  task.getExecution().GetWorkflowId(),
			RunID:       task.getExecution().GetRunId(),
			BranchToken: versionHistory.GetBranchToken(),
			Events:      task.getEvents(),
		},
	)
}

func (r *nDCHistoryReplicator) applyNonStartEventsMissingMutableState(
	ctx ctx.Context,
	newContext workflowExecutionContext,
	task nDCReplicationTask,
) (mutableState, nDCWorkflowResetterCompleteFn, error) {

	// for non reset workflow execution replication task, just do re-application
	if !task.getRequest().GetResetWorkflow() {
		// TODO we should use a new retry error for 3+DC
		return nil, nil, newRetryTaskErrorWithHint(
			ErrWorkflowNotFoundMsg,
			task.getDomainID(),
			task.getWorkflowID(),
			task.getRunID(),
			common.FirstEventID,
		)
	}

	decisionTaskFailedEvent := task.getFirstEvent()
	attr := decisionTaskFailedEvent.DecisionTaskFailedEventAttributes
	baseRunID := attr.GetBaseRunId()
	baseEventID := decisionTaskFailedEvent.GetEventId() - 1
	baseEventVersion := attr.GetForkEventVersion()
	newRunID := attr.GetNewRunId()

	workflowResetter := r.newWorkflowResetter(
		task.getDomainID(),
		task.getWorkflowID(),
		baseRunID,
		newContext,
		newRunID,
		task.getLogger(),
	)

	return workflowResetter.resetWorkflow(ctx, task.getEventTime(), baseEventID, baseEventVersion)
}

func (r *nDCHistoryReplicator) applyNonStartEventsResetWorkflow(
	ctx ctx.Context,
	context workflowExecutionContext,
	mutableState mutableState,
	completeFn nDCWorkflowResetterCompleteFn,
	task nDCReplicationTask,
) error {

	defer completeFn()

	requestID := uuid.New() // requestID used for start workflow execution request.  This is not on the history event.
	stateBuilder := r.newStateBuilder(mutableState, task.getLogger())
	_, _, _, err := stateBuilder.applyEvents(
		task.getDomainID(),
		requestID,
		*task.getExecution(),
		task.getEvents(),
		task.getNewEvents(),
		nDCMutableStateEventStoreVersion,
		nDCMutableStateEventStoreVersion,
	)
	if err != nil {
		return err
	}

	targetWorkflow := newNDCWorkflow(
		ctx,
		r.domainCache,
		r.clusterMetadata,
		context,
		mutableState,
		noopReleaseFn,
	)

	err = r.transactionMgr.createWorkflow(
		ctx,
		task.getEventTime(),
		targetWorkflow,
	)
	if err == nil {
		r.notify(task.getSourceCluster(), task.getEventTime())
	}
	return err
}

func (r *nDCHistoryReplicator) notify(
	clusterName string,
	now time.Time,
) {

	now = now.Add(-r.shard.GetConfig().StandbyClusterDelay())
	r.shard.SetCurrentTime(clusterName, now)
}
