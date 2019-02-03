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

package sysworkflow

import (
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service/dynamicconfig"
)

type (
	// HistoryBatchIterator is used to iterate over batches of history from the persistence layer
	HistoryBatchIterator interface {
		Next() (*HistoryBatch, error)
		HasNext() bool
	}

	historyBatchIterator struct {
		historyManager    persistence.HistoryManager
		historyV2Manager  persistence.HistoryV2Manager
		domainID          string
		workflowID        string
		runID             string
		eventStoreVersion int32
		branchToken       []byte
		lastFirstEventID  int64
		nextPageToken     []byte
		finished          bool
		historyPageSize   dynamicconfig.IntPropertyFnWithDomainFilter
		domain            string // only used for dynamic config lookup
	}

	// HistoryBatch contains a page of history events and an estimate of its serialized size
	HistoryBatch struct {
		events []*shared.HistoryEvent
		size   int
	}
)

// NewHistoryBatchIterator returns a new HistoryBatchIterator
func NewHistoryBatchIterator(
	historyManager persistence.HistoryManager,
	historyV2Manager persistence.HistoryV2Manager,
	domainID string,
	workflowID string,
	runID string,
	eventStoreVersion int32,
	branchToken []byte,
	lastFirstEventID int64,
	historyPageSize dynamicconfig.IntPropertyFnWithDomainFilter,
	domain string,
) HistoryBatchIterator {
	return &historyBatchIterator{
		historyManager:    historyManager,
		historyV2Manager:  historyV2Manager,
		domainID:          domainID,
		workflowID:        workflowID,
		runID:             runID,
		eventStoreVersion: eventStoreVersion,
		branchToken:       branchToken,
		lastFirstEventID:  lastFirstEventID,
		nextPageToken:     []byte{},
		historyPageSize:   historyPageSize,
		domain:            domain,
	}
}

// Next returns the next item and advances iterator. Returns error if iterator is empty or if history could not be read.
func (i *historyBatchIterator) Next() (*HistoryBatch, error) {
	if !i.HasNext() {
		return nil, iteratorEmptyErr
	}
	batch, nextPageToken, err := i.readHistory()
	if err != nil {
		return nil, err
	}
	i.nextPageToken = nextPageToken
	if len(i.nextPageToken) == 0 {
		i.finished = true
	}
	return batch, nil
}

// HasNext returns true if there are more items to iterate over
func (i *historyBatchIterator) HasNext() bool {
	return !i.finished
}

func (i *historyBatchIterator) readHistory() (*HistoryBatch, []byte, error) {
	if i.eventStoreVersion == persistence.EventStoreVersionV2 {
		req := &persistence.ReadHistoryBranchRequest{
			BranchToken:   i.branchToken,
			MinEventID:    common.FirstEventID,
			MaxEventID:    i.lastFirstEventID,
			PageSize:      i.historyPageSize(i.domain),
			NextPageToken: i.nextPageToken,
		}
		events, size, nextPageToken, err := persistence.ReadFullPageV2Events(i.historyV2Manager, req)
		if err != nil {
			return nil, nil, err
		}
		return &HistoryBatch{
			events: events,
			size:   size,
		}, nextPageToken, nil
	}
	req := &persistence.GetWorkflowExecutionHistoryRequest{
		DomainID: i.domainID,
		Execution: shared.WorkflowExecution{
			WorkflowId: common.StringPtr(i.workflowID),
			RunId:      common.StringPtr(i.runID),
		},
		FirstEventID:  common.FirstEventID,
		NextEventID:   i.lastFirstEventID,
		PageSize:      i.historyPageSize(i.domain),
		NextPageToken: i.nextPageToken,
	}
	resp, err := i.historyManager.GetWorkflowExecutionHistory(req)
	if err != nil {
		return nil, nil, err
	}
	return &HistoryBatch{
		events: resp.History.Events,
		size:   resp.Size,
	}, resp.NextPageToken, nil
}
