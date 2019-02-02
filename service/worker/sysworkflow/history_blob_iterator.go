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
	"errors"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
	"strconv"
)

type (
	// HistoryBlobIterator is used to get history blobs
	HistoryBlobIterator interface {
		Next() (*HistoryBlob, error)
		HasNext() bool
	}

	historyBlobIterator struct {
		hbItr      HistoryBatchIterator
		config     *Config
		domainID   string
		workflowID string
		runID      string
		domain     string // only used for dynamic config lookup
		pageToken  int
	}
)

// NewHistoryBlobIterator returns a new HistoryBlobIterator
func NewHistoryBlobIterator(
	historyManager persistence.HistoryManager,
	historyV2Manager persistence.HistoryV2Manager,
	domainID string,
	workflowID string,
	runID string,
	eventStoreVersion int32,
	branchToken []byte,
	lastFirstEventID int64,
	config *Config,
	domain string,
) HistoryBlobIterator {
	return &historyBlobIterator{
		hbItr: NewHistoryBatchIterator(
			historyManager,
			historyV2Manager,
			domainID,
			workflowID,
			runID,
			eventStoreVersion,
			branchToken,
			lastFirstEventID,
			config.HistoryPageSize,
			domain),
		config: config,
	}
}

// Next returns history blob and advances iterator. Returns error is iterator is empty, or if history could not be fetched.
func (i *historyBlobIterator) Next() (*HistoryBlob, error) {
	if !i.HasNext() {
		return nil, errors.New("iterator is empty")
	}
	var events []*shared.HistoryEvent
	var size int
	var firstEvent *shared.HistoryEvent
	var lastEvent *shared.HistoryEvent
	var eventCount int

	for i.hbItr.HasNext() && size < i.config.TargetArchivalBlobSize(i.domain) {
		batch, err := i.hbItr.Next()
		if err != nil {
			return nil, err
		}
		events = append(events, batch.events...)
		size += batch.size
		if firstEvent == nil {
			firstEvent = batch.events[0]
		}
		lastEvent = batch.events[len(batch.events)-1]
		eventCount += len(batch.events)
	}

	header := &HistoryBlobHeader{
		DomainName:           &i.domain,
		DomainID:             &i.domainID,
		WorkflowID:           &i.workflowID,
		RunID:                &i.runID,
		CurrentPageToken:     common.StringPtr(strconv.Itoa(i.pageToken)),
		FirstFailoverVersion: firstEvent.Version,
		LastFailoverVersion:  lastEvent.Version,
		FirstEventID:         firstEvent.EventId,
		LastEventID:          lastEvent.EventId,
		// TODO: figure out how to generate the rest of the header, we need cluster upload cluster, we should have date time be in milisecondson header (refer to chronos) and we need event count which we have.

	}
	if i.hbItr.HasNext() {
		i.pageToken++
		header.NextPageToken = common.StringPtr(strconv.Itoa(i.pageToken))
	}
	return &HistoryBlob{
		Header: header,
		Body: &shared.History{
			Events: events,
		},
	}, nil
}

// HasNext returns true if there are more items to iterate over
func (i *historyBlobIterator) HasNext() bool {
	return i.hbItr.HasNext()
}
