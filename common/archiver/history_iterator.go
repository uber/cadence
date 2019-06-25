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
	HistoryIterator interface {
		Next() ([]*shared.History, error)
		HasNext() bool
		GetState() ([]byte, error)
	}

	historyIteratorState struct {
		NextEventID       int64
		FinishedIteration bool
	}

	HistoryIteratorConfig struct {
		HistoryPageSize          dynamicconfig.IntPropertyFnWithDomainFilter
		TargetHistoryBatchesSize dynamicconfig.IntPropertyFnWithDomainFilter
	}

	historyIterator struct {
		historyIteratorState

		historyManager    persistence.HistoryManager
		historyV2Manager  persistence.HistoryV2Manager
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

func NewHistoryIterator(
	request ArchiveHistoryRequest,
	historyManager persistence.HistoryManager,
	historyV2Manager persistence.HistoryV2Manager,
	config *HistoryIteratorConfig,
	initialState []byte,
) (HistoryIterator, error) {
	it := &historyIterator{
		historyIteratorState: historyIteratorState{
			NextEventID:       common.FirstEventID,
			FinishedIteration: false,
		},
		historyManager:    historyManager,
		historyV2Manager:  historyV2Manager,
		domainID:          request.DomainID,
		domainName:        request.DomainName,
		workflowID:        request.WorkflowID,
		runID:             request.RunID,
		shardID:           request.ShardID,
		eventStoreVersion: request.EventStoreVersion,
		branchToken:       request.BranchToken,
		config:            config,
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
	return nil, historyIteratorState{}, nil
}

func (i *historyIterator) readHistory(firstEventID int64) ([]*shared.History, int, error) {
	if i.eventStoreVersion == persistence.EventStoreVersionV2 {
		req := &persistence.ReadHistoryBranchRequest{
			BranchToken: i.branchToken,
			MinEventID:  firstEventID,
			MaxEventID:  common.EndEventID,
			PageSize:    i.config.HistoryPageSize(i.domainName),
			ShardID:     common.IntPtr(i.shardID),
		}

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
