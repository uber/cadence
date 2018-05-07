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

/*
import (
	"github.com/uber-common/bark"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
)

type (
	conflictResolver struct {
		shard              ShardContext
		executionMgr       persistence.ExecutionManager
		historyMgr         persistence.HistoryManager
		hSerializerFactory persistence.HistorySerializerFactory
		logger             bark.Logger
	}
)

func newConflictResolver(shard ShardContext, historyCache *historyCache, domainCache cache.DomainCache,
	historyMgr persistence.HistoryManager, logger bark.Logger) *historyReplicator {
	replicator := &historyReplicator{
		shard:             shard,
		historyCache:      historyCache,
		domainCache:       domainCache,
		historyMgr:        historyMgr,
		historySerializer: persistence.NewJSONHistorySerializer(),
		logger:            logger,
	}

	return replicator
}

func (r *conflictResolver) getHistory(domainID, workflowID, runID string, firstEventID, nextEventID int64,
	nextPageToken []byte) (*shared.History, []byte, error) {

	response, err := r.historyMgr.GetWorkflowExecutionHistory(&persistence.GetWorkflowExecutionHistoryRequest{
		DomainID: domainID,
		Execution: shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
		FirstEventID:  firstEventID,
		NextEventID:   nextEventID,
		PageSize:      defaultHistoryPageSize,
		NextPageToken: nextPageToken,
	})

	if err != nil {
		return nil, nil, err
	}

	historyEvents := []*shared.HistoryEvent{}
	for _, e := range response.Events {
		persistence.SetSerializedHistoryDefaults(&e)
		s, _ := r.hSerializerFactory.Get(e.EncodingType)
		history, err1 := s.Deserialize(&e)
		if err1 != nil {
			return nil, nil, err1
		}
		historyEvents = append(historyEvents, history.Events...)
	}

	executionHistory := &shared.History{}
	executionHistory.Events = historyEvents
	return executionHistory, nextPageToken, nil
}
*/
