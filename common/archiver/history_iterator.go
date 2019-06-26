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

package archiver

import (
	"encoding/json"
	"errors"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service/dynamicconfig"
)

type (
	// HistoryBlobIterator is used to get history batches
	HistoryIterator interface {
		Next() ([]*shared.History, error)
		HasNext() bool
		GetState() ([]byte, error)
	}

	historyIteratorState struct {
		NextEventID       int64
		FinishedIteration bool
	}

	// HistoryIteratorConfig cofigs the history iterator
	HistoryIteratorConfig struct {
		HistoryPageSize          dynamicconfig.IntPropertyFnWithDomainFilter
		TargetHistoryBatchesSize dynamicconfig.IntPropertyFnWithDomainFilter
	}

	historyIterator struct {
		historyIteratorState

		historyManager    persistence.HistoryManager
		historyV2Manager  persistence.HistoryV2Manager
		sizeEstimator     SizeEstimator
		domainID          string
		domainName        string
		workflowID        string
		runID             string
		shardID           int
		eventStoreVersion int32
		branchToken       []byte
		config            *HistoryIteratorConfig
	}
)

var (
	errIteratorDepleted = errors.New("iterator is depleted")
)

// NewHistoryIterator returns a new HistoryIterator
func NewHistoryIterator(
	request ArchiveHistoryRequest,
	historyManager persistence.HistoryManager,
	historyV2Manager persistence.HistoryV2Manager,
	config *HistoryIteratorConfig,
	initialState []byte,
	sizeEstimator SizeEstimator,
) (HistoryIterator, error) {
	it := &historyIterator{
		historyIteratorState: historyIteratorState{
			NextEventID:       common.FirstEventID,
			FinishedIteration: false,
		},
		historyManager:    historyManager,
		historyV2Manager:  historyV2Manager,
		sizeEstimator:     sizeEstimator,
		domainID:          request.DomainID,
		domainName:        request.DomainName,
		workflowID:        request.WorkflowID,
		runID:             request.RunID,
		shardID:           request.ShardID,
		eventStoreVersion: request.EventStoreVersion,
		branchToken:       request.BranchToken,
		config:            config,
	}
	if it.sizeEstimator == nil {
		it.sizeEstimator = NewJSONSizeEstimator()
	}
	if initialState == nil {
		return it, nil
	}
	if err := it.reset(initialState); err != nil {
		return nil, err
	}
	return it, nil
}

func (i *historyIterator) Next() ([]*shared.History, error) {
	if !i.HasNext() {
		return nil, errIteratorDepleted
	}

	historyBatches, newIterState, err := i.readHistoryBatches(i.NextEventID)
	if err != nil {
		return nil, err
	}

	i.historyIteratorState = newIterState
	return historyBatches, nil
}

// HasNext returns true if there are more items to iterate over.
func (i *historyIterator) HasNext() bool {
	return !i.FinishedIteration
}

// GetState returns the encoded iterator state
func (i *historyIterator) GetState() ([]byte, error) {
	return json.Marshal(i.historyIteratorState)
}

func (i *historyIterator) readHistoryBatches(firstEventID int64) ([]*shared.History, historyIteratorState, error) {
	size := 0
	targetSize := i.config.TargetHistoryBatchesSize(i.domainName)
	var historyBatches []*shared.History
	newIterState := historyIteratorState{}
	nextPageToken := []byte{}
	for size == 0 || (len(nextPageToken) > 0 && size < targetSize) {
		var currHistoryBatches []*shared.History
		var err error
		currHistoryBatches, nextPageToken, err = i.readHistory(firstEventID)
		if err != nil {
			return nil, newIterState, err
		}
		for idx, batch := range currHistoryBatches {
			batchSize, err := i.sizeEstimator.EstimateSize(batch)
			if err != nil {
				return nil, newIterState, err
			}
			size += batchSize
			historyBatches = append(historyBatches, batch)
			firstEventID = *batch.Events[len(batch.Events)-1].EventId + 1

			// If targetSize is meeted after appending the last batch, we are not sure if there's more batches or not,
			// so we need to exclude that case.
			if size >= targetSize && idx != len(currHistoryBatches)-1 {
				newIterState.FinishedIteration = false
				newIterState.NextEventID = firstEventID
				return historyBatches, newIterState, nil
			}
		}
	}

	if len(nextPageToken) == 0 {
		newIterState.FinishedIteration = true
		return historyBatches, newIterState, nil
	}

	// If nextPageToken is not empty it is still possible there are more batches.
	// This occurs if history was read exactly to the last batch.
	// Here we look forward so that we can treat reading exactly to the end of history
	// the same way as reading through the end of history.
	_, _, err := i.readHistory(firstEventID)
	if _, ok := err.(*shared.EntityNotExistsError); ok {
		newIterState.FinishedIteration = true
		return historyBatches, newIterState, nil
	}
	if err != nil {
		return nil, newIterState, err
	}
	newIterState.FinishedIteration = false
	newIterState.NextEventID = firstEventID
	return historyBatches, newIterState, nil
}

func (i *historyIterator) readHistory(firstEventID int64) ([]*shared.History, []byte, error) {
	if i.eventStoreVersion == persistence.EventStoreVersionV2 {
		req := &persistence.ReadHistoryBranchRequest{
			BranchToken: i.branchToken,
			MinEventID:  firstEventID,
			MaxEventID:  common.EndEventID,
			PageSize:    i.config.HistoryPageSize(i.domainName),
			ShardID:     common.IntPtr(i.shardID),
		}
		return i.readFullPageV2EventsByBatch(i.historyV2Manager, req)
	}
	req := &persistence.GetWorkflowExecutionHistoryRequest{
		DomainID: i.domainID,
		Execution: shared.WorkflowExecution{
			WorkflowId: common.StringPtr(i.workflowID),
			RunId:      common.StringPtr(i.runID),
		},
		FirstEventID: firstEventID,
		NextEventID:  common.EndEventID,
		PageSize:     i.config.HistoryPageSize(i.domainName),
	}
	resp, err := i.historyManager.GetWorkflowExecutionHistoryByBatch(req)
	if err != nil {
		return nil, nil, err
	}
	return resp.History, resp.NextPageToken, nil
}

func (i *historyIterator) readFullPageV2EventsByBatch(
	historyV2Mgr persistence.HistoryV2Manager,
	req *persistence.ReadHistoryBranchRequest,
) ([]*shared.History, []byte, error) {
	historyBatches := []*shared.History{}
	eventsRead := 0
	for {
		response, err := historyV2Mgr.ReadHistoryBranchByBatch(req)
		if err != nil {
			return nil, nil, err
		}
		historyBatches = append(historyBatches, response.History...)
		for _, batch := range historyBatches {
			eventsRead += len(batch.Events)
		}
		if eventsRead >= req.PageSize || len(response.NextPageToken) == 0 {
			return historyBatches, response.NextPageToken, nil
		}
		req.NextPageToken = response.NextPageToken
	}
}

// reset resets iterator to a certain state given its encoded representation
// if it returns an error, the operation will have no effect on the iterator
func (i *historyIterator) reset(stateToken []byte) error {
	var iteratorState historyIteratorState
	if err := json.Unmarshal(stateToken, &iteratorState); err != nil {
		return err
	}
	i.historyIteratorState = iteratorState
	return nil
}

type (
	// SizeEstimator is used to estimate the size of any object
	SizeEstimator interface {
		EstimateSize(v interface{}) (int, error)
	}

	jsonSizeEstimator struct{}
)

func (e *jsonSizeEstimator) EstimateSize(v interface{}) (int, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}
	return len(data), nil
}

// NewJSONSizeEstimator returns a new SizeEstimator which uses json encoding to
// estimate size
func NewJSONSizeEstimator() SizeEstimator {
	return &jsonSizeEstimator{}
}
