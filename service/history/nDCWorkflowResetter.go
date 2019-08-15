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

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
)

type (
	nDCWorkflowResetter interface {
		resetMutableState(
			ctx ctx.Context,
			baseEventID int64,
			baseVersion int64,
		) (mutableState, error)
	}

	nDCWorkflowResetterImpl struct {
		shard           ShardContext
		domainCache     cache.DomainCache
		clusterMetadata cluster.Metadata
		historyV2Mgr    persistence.HistoryV2Manager

		baseContext      workflowExecutionContext
		baseMutableState mutableState
		resetContext     workflowExecutionContext

		resetHistorySize int64
		logger           log.Logger
	}
)

var _ nDCWorkflowResetter = (*nDCWorkflowResetterImpl)(nil)

func newNDCWorkflowResetter(
	shard ShardContext,

	baseContext workflowExecutionContext,
	baseMutableState mutableState,
	resetContext workflowExecutionContext,
	logger log.Logger,
) *nDCWorkflowResetterImpl {

	return &nDCWorkflowResetterImpl{
		shard:           shard,
		domainCache:     shard.GetDomainCache(),
		clusterMetadata: shard.GetService().GetClusterMetadata(),
		historyV2Mgr:    shard.GetHistoryV2Manager(),

		baseContext:      baseContext,
		baseMutableState: baseMutableState,
		resetContext:     resetContext,
		resetHistorySize: 0,
		logger:           logger,
	}
}

func (r *nDCWorkflowResetterImpl) resetMutableState(
	ctx ctx.Context,
	baseEventID int64,
	baseVersion int64,
) (mutableState, error) {

	// baseVersionHistories := r.baseMutableState.GetVersionHistories()
	// index, err := baseVersionHistories.FindFirstVersionHistoryIndexByItem(
	// 	persistence.NewVersionHistoryItem(baseEventID, baseVersion),
	// )
	// if err != nil {
	// 	// TODO we should use a new retry error for 3+DC
	// 	baseExecutionInfo := r.baseMutableState.GetExecutionInfo()
	// 	return nil, newRetryTaskErrorWithHint(
	// 		ErrRetryBufferEventsMsg,
	// 		baseExecutionInfo.DomainID,
	// 		baseExecutionInfo.WorkflowID,
	// 		baseExecutionInfo.RunID,
	// 		r.baseMutableState.GetNextEventID(), // especially here
	// 	)
	// }
	//
	// baseVersionHistory, err := baseVersionHistories.GetVersionHistory(index)
	// if err != nil {
	// 	return nil, err
	// }
	// baseBranchToken := baseVersionHistory.GetBranchToken()

	panic("implement this")
}
