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
	"github.com/uber-common/bark"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/logging"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
)

type (
	eventsCache struct {
		cache.Cache
		eventsMgr     persistence.HistoryManager
		eventsV2Mgr   persistence.HistoryV2Manager
		logger        bark.Logger
		metricsClient metrics.Client
	}

	eventKey struct {
		domainID   string
		workflowID string
		runID      string
		eventID    int64
	}
)

func newEventsCache(shardCtx ShardContext) *eventsCache {
	opts := &cache.Options{}
	config := shardCtx.GetConfig()
	opts.InitialCapacity = config.EventsCacheInitialSize()
	opts.TTL = config.EventsCacheTTL()

	return &eventsCache{
		Cache:       cache.New(config.EventsCacheMaxSize(), opts),
		eventsMgr:   shardCtx.GetHistoryManager(),
		eventsV2Mgr: shardCtx.GetHistoryV2Manager(),
		logger: shardCtx.GetLogger().WithFields(bark.Fields{
			logging.TagWorkflowComponent: logging.TagValueEventsCacheComponent,
		}),
		metricsClient: shardCtx.GetMetricsClient(),
	}
}

func newEventKey(domainID, workflowID, runID string, eventID int64) eventKey {
	return eventKey{
		domainID:   domainID,
		workflowID: workflowID,
		runID:      runID,
		eventID:    eventID,
	}
}

func (e *eventsCache) getEvent(domainID, workflowID, runID string, eventID int64, eventStoreVersion int32,
	branchToken []byte) (*shared.HistoryEvent, error) {
		key := newEventKey(domainID, workflowID, runID, eventID)
		event, cacheHit := e.Cache.Get(key).(*shared.HistoryEvent)
		if cacheHit {
			return event, nil
		}

		event, err := e.getHistoryEventFromStore(domainID, workflowID, runID, eventID, eventStoreVersion, branchToken)
		if err != nil {
			return nil, err
		}

		e.Put(key, event)
		return event, nil
}

func (e *eventsCache) putEvent(domainID, workflowID, runID string, eventID int64, event *shared.HistoryEvent) {
	key := newEventKey(domainID, workflowID, runID, eventID)
	e.Put(key, event)
}

func (e *eventsCache) deleteEvent(domainID, workflowID, runID string, eventID int64) {
	key := newEventKey(domainID, workflowID, runID, eventID)
	e.Delete(key)
}

func (e *eventsCache) getHistoryEventFromStore(domainID, workflowID, runID string, eventID int64,
	eventStoreVersion int32, branchToken []byte) (*shared.HistoryEvent, error) {

	var historyEvent *shared.HistoryEvent
	if eventStoreVersion == persistence.EventStoreVersionV2 {
		response, err := e.eventsV2Mgr.ReadHistoryBranch(&persistence.ReadHistoryBranchRequest{
			BranchToken:   branchToken,
			MinEventID:    eventID,
			MaxEventID:    eventID + 1,
			PageSize:      1,
			NextPageToken: nil,
		})

		if err != nil {
			return nil, err
		}

		if len(response.History) > 0 {
			historyEvent = response.History[0]
		}
	} else {
		response, err := e.eventsMgr.GetWorkflowExecutionHistory(&persistence.GetWorkflowExecutionHistoryRequest{
			DomainID: domainID,
			Execution: shared.WorkflowExecution{
				WorkflowId: common.StringPtr(workflowID),
				RunId:      common.StringPtr(runID),
			},
			FirstEventID:  eventID,
			NextEventID:   eventID + 1,
			PageSize:      1,
			NextPageToken: nil,
		})

		if err != nil {
			return nil, err
		}

		if response.History != nil && len(response.History.Events) > 0 {
			historyEvent = response.History.Events[0]
		}
	}

	return historyEvent, nil
}
