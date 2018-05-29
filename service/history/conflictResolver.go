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
	"time"

	"github.com/pborman/uuid"
	"github.com/uber-common/bark"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
)

type (
	conflictResolver struct {
		shard              ShardContext
		context            *workflowExecutionContext
		historyMgr         persistence.HistoryManager
		hSerializerFactory persistence.HistorySerializerFactory
		logger             bark.Logger
	}
)

func newConflictResolver(shard ShardContext, context *workflowExecutionContext, historyMgr persistence.HistoryManager,
	logger bark.Logger) *conflictResolver {

	return &conflictResolver{
		shard:              shard,
		context:            context,
		historyMgr:         historyMgr,
		hSerializerFactory: persistence.NewHistorySerializerFactory(),
		logger:             logger,
	}
}

func (r *conflictResolver) reset(replayEventID int64, startTime time.Time) (*mutableStateBuilder, error) {
	domainID := r.context.domainID
	execution := r.context.workflowExecution
	replayNextEventID := replayEventID + 1
	var nextPageToken []byte
	var history *shared.History
	var err error
	var resetMutableStateBuilder *mutableStateBuilder
	var sBuilder *stateBuilder
	requestID := uuid.New()

	var lastFirstEventID int64
	for remainingHistorySize := replayNextEventID - common.FirstEventID; remainingHistorySize > 0; {
		history, nextPageToken, lastFirstEventID, err = r.getHistory(domainID, execution, common.FirstEventID, replayNextEventID,
			nextPageToken)
		if err != nil {
			return nil, err
		}

		if int64(len(history.Events)) <= remainingHistorySize {
			remainingHistorySize -= int64(len(history.Events))
		} else {
			history.Events = history.Events[0:remainingHistorySize]
			remainingHistorySize = 0
		}

		if history.Events[0].GetEventId() == common.FirstEventID {
			resetMutableStateBuilder = newMutableStateBuilderWithReplicationState(r.shard.GetConfig(), r.logger,
				history.Events[0].GetVersion())

			sBuilder = newStateBuilder(r.shard, resetMutableStateBuilder, r.logger)
		}
		resetMutableStateBuilder.executionInfo.LastFirstEventID = lastFirstEventID
		_, _, _, err = sBuilder.applyEvents(common.EmptyVersion, "", domainID, requestID, execution, history, nil)
		if err != nil {
			return nil, err
		}
	}
	resetMutableStateBuilder.executionInfo.NextEventID = replayNextEventID
	resetMutableStateBuilder.executionInfo.StartTimestamp = startTime
	// the last updated time is not important here, since this should be updated with event time afterwards
	resetMutableStateBuilder.executionInfo.LastUpdatedTimestamp = startTime

	return r.context.resetWorkflowExecution(resetMutableStateBuilder)
}

func (r *conflictResolver) getHistory(domainID string, execution shared.WorkflowExecution, firstEventID,
	nextEventID int64, nextPageToken []byte) (*shared.History, []byte, int64, error) {

	response, err := r.historyMgr.GetWorkflowExecutionHistory(&persistence.GetWorkflowExecutionHistoryRequest{
		DomainID:      domainID,
		Execution:     execution,
		FirstEventID:  firstEventID,
		NextEventID:   nextEventID,
		PageSize:      defaultHistoryPageSize,
		NextPageToken: nextPageToken,
	})

	if err != nil {
		return nil, nil, 0, err
	}

	var lastFirstEventID int64
	historyEvents := []*shared.HistoryEvent{}
	for _, e := range response.Events {
		persistence.SetSerializedHistoryDefaults(&e)
		s, _ := r.hSerializerFactory.Get(e.EncodingType)
		history, err1 := s.Deserialize(&e)
		if err1 != nil {
			return nil, nil, 0, err1
		}
		lastFirstEventID = history.Events[0].GetEventId()
		historyEvents = append(historyEvents, history.Events...)
	}

	executionHistory := &shared.History{}
	executionHistory.Events = historyEvents
	return executionHistory, nextPageToken, lastFirstEventID, nil
}
