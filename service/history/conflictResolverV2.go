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
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/collection"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
)

const (
	defaultPageSize   = 100
	eventStoreVersion = persistence.EventStoreVersionV2
)

type (
	conflictResolverV2 interface {
		reset(prevRunID string, requestID string, replayEventID int64, info *persistence.WorkflowExecutionInfo) (mutableState, error)
	}

	conflictResolverV2Impl struct {
		shard           ShardContext
		clusterMetadata cluster.Metadata
		context         workflowExecutionContext
		historyMgr      persistence.HistoryManager
		historyV2Mgr    persistence.HistoryV2Manager
		logger          log.Logger
	}
)

func newConflictResolverV2(shard ShardContext, context workflowExecutionContext, historyMgr persistence.HistoryManager, historyV2Mgr persistence.HistoryV2Manager,
	logger log.Logger) *conflictResolverImpl {

	return &conflictResolverImpl{
		shard:           shard,
		clusterMetadata: shard.GetService().GetClusterMetadata(),
		context:         context,
		historyMgr:      historyMgr,
		historyV2Mgr:    historyV2Mgr,
		logger:          logger,
	}
}

func (r *conflictResolverV2Impl) reset(prevRunID string, requestID string, replayEventID int64, info *persistence.WorkflowExecutionInfo, localVersionHistories persistence.VersionHistories) (mutableState, error) {
	domainID := r.context.getDomainID()
	execution := *r.context.getExecution()
	branchToken := info.GetCurrentBranch()
	replayNextEventID := replayEventID + 1

	var resetMutableStateBuilder *mutableStateBuilder
	var sBuilder stateBuilder
	var lastEvent *shared.HistoryEvent

	iter := collection.NewPagingIterator(r.getHistoryBatch(domainID, execution, common.FirstEventID, replayNextEventID, nil, eventStoreVersion, branchToken))
	for iter.HasNext() {
		batch, err := iter.Next()
		if err != nil {
			r.logger.Error("Conflict resolution err getting history.", tag.Error(err))
			return nil, err
		}

		history := batch.(*shared.History).Events
		firstEvent := history[0]
		lastEvent = history[len(history)-1]
		if firstEvent.GetEventId() == common.FirstEventID {
			resetMutableStateBuilder, sBuilder = r.initializeBuilders(firstEvent)
		}
		_, _, _, err = sBuilder.applyEvents(domainID, requestID, execution, history, nil, eventStoreVersion, eventStoreVersion)
		if err != nil {
			r.logger.Error("conflict resolution err applying events.", tag.Error(err))
			return nil, err
		}
		resetMutableStateBuilder.IncrementHistorySize(len(history))
	}
	//Sanity check on last event id of the last history batch
	if *lastEvent.EventId != replayEventID {
		return resetMutableStateBuilder, &shared.BadRequestError{
			Message: fmt.Sprintf("failed to resolve conflict as the last even id %v and replay event id %v are not matched", lastEvent, replayNextEventID),
		}
	}

	msBuilder, err := r.resetMutableState(resetMutableStateBuilder, info, lastEvent, replayEventID, prevRunID)
	if err != nil {
		r.logger.Error("conflict resolution err reset workflow.", tag.Error(err))
	}
	//TODO: use LCA on the msBuilder.getVersionHistories() vs localVersionHistories
	//r.resolveVersionHistoryConflict(localVersionHistories, msBuilder.GetVersionHistories())
	return msBuilder, err
}

func (r *conflictResolverV2Impl) resolveVersionHistoryConflict(local persistence.VersionHistories, remote persistence.VersionHistories) error {

	for _, newHistory := range remote.GetHistories() {
		commonEventId, history, err := local.FindLowestCommonVersionHistory(newHistory)
		if err != nil {
			return err
		}

		if err := local.AddHistory(commonEventId, history, newHistory); err != nil {
			return err
		}
	}
	return nil
}

func (r *conflictResolverV2Impl) getHistoryBatch(domainID string, execution shared.WorkflowExecution, firstEventID,
	nextEventID int64, nextPageToken []byte, eventStoreVersion int32, branchToken []byte) collection.PaginationFn {
	return func(paginationToken []byte) ([]interface{}, []byte, error) {
		var paginateItems []interface{}
		_, historyBatches, token, _, err := PaginateHistory(
			r.historyMgr,
			r.historyV2Mgr,
			nil,
			r.logger,
			true,
			domainID,
			*execution.WorkflowId,
			*execution.RunId,
			firstEventID,
			nextEventID,
			nextPageToken,
			eventStoreVersion,
			branchToken,
			defaultPageSize,
			common.IntPtr(r.shard.GetShardID()))
		for _, history := range historyBatches {
			paginateItems = append(paginateItems, history)
		}
		return paginateItems, token, err
	}
}

func (r *conflictResolverV2Impl) initializeBuilders(firstEvent *shared.HistoryEvent) (*mutableStateBuilder, *stateBuilderImpl) {
	resetMutableStateBuilder := newMutableStateBuilderWithReplicationState(
		r.clusterMetadata.GetCurrentClusterName(),
		r.shard,
		r.shard.GetEventsCache(),
		r.logger,
		firstEvent.GetVersion(),
	)
	resetMutableStateBuilder.executionInfo.EventStoreVersion = eventStoreVersion
	sBuilder := newStateBuilder(r.shard, resetMutableStateBuilder, r.logger)
	return resetMutableStateBuilder, sBuilder
}

func (r *conflictResolverV2Impl) resetMutableState(msBuilder *mutableStateBuilder,
	execution *persistence.WorkflowExecutionInfo,
	lastEvent *shared.HistoryEvent,
	replayEventID int64,
	prevRunID string) (mutableState, error) {
	// reset branchToken to the original one(it has been set to a wrong branchToken in applyEvents for startEvent)
	msBuilder.executionInfo.BranchToken = execution.GetCurrentBranch()
	// similarly, in case of resetWF, the runID in startEvent is incorrect
	msBuilder.executionInfo.RunID = execution.RunID
	msBuilder.executionInfo.StartTimestamp = execution.StartTimestamp
	// the last updated time is not important here, since this should be updated with event time afterwards
	msBuilder.executionInfo.LastUpdatedTimestamp = execution.StartTimestamp
	sourceCluster := r.clusterMetadata.ClusterNameForFailoverVersion(lastEvent.GetVersion())
	msBuilder.UpdateReplicationStateLastEventID(sourceCluster, lastEvent.GetVersion(), replayEventID)

	r.logger.Info("All events applied for execution.", tag.WorkflowResetNextEventID(msBuilder.GetNextEventID()))
	return r.context.resetMutableState(prevRunID, msBuilder)
}
