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
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/collection"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
)

const (
	defaultPageSize = 100
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
	startTime := info.StartTimestamp
	eventStoreVersion := info.EventStoreVersion
	branchToken := info.GetCurrentBranch()
	replayNextEventID := replayEventID + 1

	var resetMutableStateBuilder *mutableStateBuilder
	var sBuilder stateBuilder
	var lastEvent *shared.HistoryEvent
	var lastFirstEventID int64

	eventsToApply := replayNextEventID - common.FirstEventID

	iter := collection.NewPagingIterator(func(paginationToken []byte) ([]interface{}, []byte, error) {
		var paginateItems []interface{}
		historyBatch, _, token, err := r.getHistoryBatch(domainID, execution, common.FirstEventID, replayNextEventID, paginationToken, eventStoreVersion, branchToken)
		for _, history := range historyBatch {
			paginateItems = append(paginateItems, history)
		}
		return paginateItems, token, err
	})

	for iter.HasNext() {
		item, err := iter.Next()
		if err != nil {
			r.logger.Error("Conflict resolution err getting history.", tag.Error(err))
			return nil, err
		}

		history := item.(*shared.History).Events
		batchSize := int64(len(history))
		// NextEventID could be in the middle of the batch.  Trim the history events to not have more events then what
		// need to be applied
		if batchSize > eventsToApply {
			history = history[0:eventsToApply]
		}
		eventsToApply -= int64(len(history))

		if len(history) == 0 {
			break
		}

		firstEvent := history[0]
		lastEvent = history[len(history)-1]
		if firstEvent.GetEventId() == common.FirstEventID {
			resetMutableStateBuilder = newMutableStateBuilderWithReplicationState(
				r.clusterMetadata.GetCurrentClusterName(),
				r.shard,
				r.shard.GetEventsCache(),
				r.logger,
				firstEvent.GetVersion(),
			)

			resetMutableStateBuilder.executionInfo.EventStoreVersion = eventStoreVersion
			sBuilder = newStateBuilder(r.shard, resetMutableStateBuilder, r.logger)
		}

		// NOTE: passing 0 as newRunEventStoreVersion is safe here, since we don't need the newMutableState of the new run
		_, _, _, err = sBuilder.applyEvents(domainID, requestID, execution, history, nil, resetMutableStateBuilder.GetEventStoreVersion(), 0)
		if err != nil {
			r.logger.Error("conflict resolution err applying events.", tag.Error(err))
			return nil, err
		}
		resetMutableStateBuilder.executionInfo.SetLastFirstEventID(lastFirstEventID)
		resetMutableStateBuilder.IncrementHistorySize(len(history))
	}

	// reset branchToken to the original one(it has been set to a wrong branchToken in applyEvents for startEvent)
	resetMutableStateBuilder.executionInfo.BranchToken = branchToken
	// similarly, in case of resetWF, the runID in startEvent is incorrect
	resetMutableStateBuilder.executionInfo.RunID = info.RunID
	// Applying events to mutableState does not move the nextEventID.  Explicitly set nextEventID to new value
	resetMutableStateBuilder.executionInfo.SetNextEventID(replayNextEventID)
	resetMutableStateBuilder.executionInfo.StartTimestamp = startTime
	// the last updated time is not important here, since this should be updated with event time afterwards
	resetMutableStateBuilder.executionInfo.LastUpdatedTimestamp = startTime

	sourceCluster := r.clusterMetadata.ClusterNameForFailoverVersion(lastEvent.GetVersion())
	resetMutableStateBuilder.UpdateReplicationStateLastEventID(sourceCluster, lastEvent.GetVersion(), replayEventID)

	r.logger.Info("All events applied for execution.", tag.WorkflowResetNextEventID(resetMutableStateBuilder.GetNextEventID()))
	msBuilder, err := r.context.resetMutableState(prevRunID, resetMutableStateBuilder)
	if err != nil {
		r.logger.Error("conflict resolution err reset workflow.", tag.Error(err))
	}
	//TODO: use LCA on the msBuilder.getVersionHistories() vs localVersionHistories
	return msBuilder, err
}

func (r *conflictResolverV2Impl) getHistoryBatch(domainID string, execution shared.WorkflowExecution, firstEventID,
	nextEventID int64, nextPageToken []byte, eventStoreVersion int32, branchToken []byte) ([]*shared.History, int, []byte, error) {
	_, historyBatches, token, size, err := PaginateHistory(
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
	return historyBatches, size, token, err
}
