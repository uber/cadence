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

package persistence

import (
	"encoding/json"
	"fmt"

	"github.com/uber-common/bark"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/logging"
)

type (

	// historyManagerImpl implements HistoryManager based on HistoryStore and HistorySerializer
	historyManagerImpl struct {
		serializer  HistorySerializer
		persistence HistoryStore
		logger      bark.Logger
	}

	// historyToken is used to serialize/deserialize pagination token for ReadHistoryBranchRequest
	historyToken struct {
		LastEventBatchVersion int64
		LastEventID           int64
		Data                  []byte
	}
)

var _ HistoryManager = (*historyManagerImpl)(nil)

//NewHistoryManagerImpl returns new HistoryManager
func NewHistoryManagerImpl(persistence HistoryStore, logger bark.Logger) HistoryManager {
	return &historyManagerImpl{
		serializer:  NewHistorySerializer(),
		persistence: persistence,
		logger:      logger,
	}
}

func (m *historyManagerImpl) GetName() string {
	return m.persistence.GetName()
}

// NewHistoryBranch creates a new branch from tree root. If tree doesn't exist, then create one. Return error if the branch already exists.
func (m *historyManagerImpl) NewHistoryBranch(request *NewHistoryBranchRequest) error {
	return m.persistence.NewHistoryBranch(request)
}

// AppendHistoryNode add(or override) a node to a history branch
func (m *historyManagerImpl) AppendHistoryNode(request *AppendHistoryNodeRequest) (*AppendHistoryNodeResponse, error) {
	if len(request.Events) == 0 {
		return nil, fmt.Errorf("events to be appended cannot be empty")
	}
	eventBlobs := []*DataBlob{}
	size := 0
	for _, e := range request.Events {
		b, err := m.serializer.SerializeEvent(e, request.Encoding)
		if err != nil {
			return nil, err
		}
		eventBlobs = append(eventBlobs, b)
		size += len(b.Data)
	}
	resp := &AppendHistoryNodeResponse{Size: size}
	return resp, m.persistence.AppendHistoryNode(&InternalAppendHistoryNodeRequest{
		BranchInfo:    request.BranchInfo,
		NextNodeID:    request.NextNodeID,
		Events:        eventBlobs,
		Overwrite:     request.Overwrite,
		TransactionID: request.TransactionID,
	})
}

// ReadHistoryBranch returns history node data for a branch
func (m *historyManagerImpl) ReadHistoryBranch(request *ReadHistoryBranchRequest) (*ReadHistoryBranchResponse, error) {
	token, err := m.deserializeToken(request.MinNodeID, request.NextPageToken, request.LastEventBatchVersion)
	if err != nil {
		return nil, err
	}

	events := []*workflow.HistoryEvent{}
	size := 0
	ir := &InternalReadHistoryBranchRequest{
		BranchInfo:    request.BranchInfo,
		MinNodeID:     request.MinNodeID,
		MaxNodeID:     request.MaxNodeID,
		PageSize:      request.PageSize,
		NextPageToken: token.Data,
	}
	resp, err := m.persistence.ReadHistoryBranch(ir)
	if err != nil {
		return nil, err
	}

	//NOTE: in this method, we need to make sure eventBatchVersion is NOT decreasing(otherwise we skip the events), eventID should be continuous(otherwise return error)
	lastEventBatchVersion := token.LastEventBatchVersion
	lastEventID := token.LastEventID
	found := false
	for _, b := range resp.History {
		e, err := m.serializer.DeserializeEvent(b)
		if err != nil {
			return nil, err
		}
		// Skip if version is decreasing
		if *e.Version < lastEventBatchVersion {
			continue
		}
		if lastEventID+1 != *e.EventId {
			logger := m.logger.WithFields(bark.Fields{
				logging.TagBranchID: request.BranchInfo.BranchID,
				logging.TagTreeID:   request.BranchInfo.TreeID,
			})
			logger.Error("Unexpected event batch")
			return nil, fmt.Errorf("corrupted history event batch, eventID is not continous")
		}

		found = true
		lastEventBatchVersion = *e.Version
		lastEventID = *e.EventId
		events = append(events, e)
		size += len(b.Data)
	}

	if len(request.NextPageToken) == 0 {
		// checking for the very first page
		if !found {
			return nil, &workflow.EntityNotExistsError{
				Message: fmt.Sprintf("Branch nodes not found.  TreeId: %v, BranchId: %v",
					request.BranchInfo.TreeID, request.BranchInfo.BranchID),
			}
		}
		// our nodeID should be strictly equal to eventID
		if *events[0].EventId != request.MinNodeID {
			logger := m.logger.WithFields(bark.Fields{
				logging.TagBranchID: request.BranchInfo.BranchID,
				logging.TagTreeID:   request.BranchInfo.TreeID,
			})
			logger.Error("Unexpected event batch")
			return nil, fmt.Errorf("corrupted history event batch, eventID does not match with nodeID")
		}
	}

	newResponse := &ReadHistoryBranchResponse{}
	newResponse.Size = size
	newResponse.History = events
	newResponse.LastNodeID = lastEventID
	token.LastEventID = lastEventID
	token.LastEventBatchVersion = lastEventBatchVersion
	token.Data = resp.NextPageToken
	newResponse.NextPageToken, err = m.serializeToken(token)
	if err != nil {
		return nil, err
	}

	return newResponse, nil
}

// ForkHistoryBranch forks a new branch from a old branch
func (m *historyManagerImpl) ForkHistoryBranch(request *ForkHistoryBranchRequest) (*ForkHistoryBranchResponse, error) {
	return m.persistence.ForkHistoryBranch(request)
}

// DeleteHistoryBranch removes a branch
func (m *historyManagerImpl) DeleteHistoryBranch(request *DeleteHistoryBranchRequest) error {
	return m.persistence.DeleteHistoryBranch(request)
}

// GetHistoryTree returns all branch information of a tree
func (m *historyManagerImpl) GetHistoryTree(request *GetHistoryTreeRequest) (*GetHistoryTreeResponse, error) {
	return m.persistence.GetHistoryTree(request)
}

func (m *historyManagerImpl) AppendHistoryEvents(request *AppendHistoryEventsRequest) (*AppendHistoryEventsResponse, error) {
	if len(request.Events) == 0 {
		return nil, fmt.Errorf("events to be appended cannot be empty")
	}
	eventsData, err := m.serializer.SerializeBatchEvents(request.Events, request.Encoding)
	if err != nil {
		return nil, err
	}

	resp := &AppendHistoryEventsResponse{Size: len(eventsData.Data)}
	return resp, m.persistence.AppendHistoryEvents(
		&InternalAppendHistoryEventsRequest{
			DomainID:          request.DomainID,
			Execution:         request.Execution,
			FirstEventID:      request.FirstEventID,
			EventBatchVersion: request.EventBatchVersion,
			RangeID:           request.RangeID,
			TransactionID:     request.TransactionID,
			Events:            eventsData,
			Overwrite:         request.Overwrite,
		})
}

// GetWorkflowExecutionHistory retrieves the paginated list of history events for given execution
func (m *historyManagerImpl) GetWorkflowExecutionHistory(request *GetWorkflowExecutionHistoryRequest) (*GetWorkflowExecutionHistoryResponse, error) {
	token, err := m.deserializeToken(request.FirstEventID, request.NextPageToken, common.EmptyVersion)
	if err != nil {
		return nil, err
	}

	// persistence API expects the actual cassandra paging token
	newRequest := &InternalGetWorkflowExecutionHistoryRequest{
		LastEventBatchVersion: token.LastEventBatchVersion,
		NextPageToken:         token.Data,

		DomainID:     request.DomainID,
		Execution:    request.Execution,
		FirstEventID: request.FirstEventID,
		NextEventID:  request.NextEventID,
		PageSize:     request.PageSize,
	}
	response, err := m.persistence.GetWorkflowExecutionHistory(newRequest)
	if err != nil {
		return nil, err
	}
	// we store LastEventBatchVersion in the token. The reason we do it here is for historic reason.
	token.LastEventBatchVersion = response.LastEventBatchVersion
	token.Data = response.NextPageToken

	newResponse := &GetWorkflowExecutionHistoryResponse{}

	history := &workflow.History{
		Events: make([]*workflow.HistoryEvent, 0),
	}

	// first_event_id of the last batch
	lastFirstEventID := common.EmptyEventID
	size := 0

	for _, b := range response.History {
		size += len(b.Data)
		historyBatch, err := m.serializer.DeserializeBatchEvents(b)
		if err != nil {
			return nil, err
		}

		if len(historyBatch) == 0 || historyBatch[0].GetEventId() > token.LastEventID+1 {
			logger := m.logger.WithFields(bark.Fields{
				logging.TagWorkflowExecutionID: request.Execution.GetWorkflowId(),
				logging.TagWorkflowRunID:       request.Execution.GetRunId(),
				logging.TagDomainID:            request.DomainID,
			})
			logger.Error("Unexpected event batch")
			return nil, fmt.Errorf("corrupted history event batch")
		}

		if historyBatch[0].GetEventId() != token.LastEventID+1 {
			// staled event batch, skip it
			continue
		}

		lastFirstEventID = historyBatch[0].GetEventId()
		history.Events = append(history.Events, historyBatch...)
		token.LastEventID = historyBatch[len(historyBatch)-1].GetEventId()
	}

	newResponse.Size = size
	newResponse.LastFirstEventID = lastFirstEventID
	newResponse.History = history
	newResponse.NextPageToken, err = m.serializeToken(token, request.NextEventID)
	if err != nil {
		return nil, err
	}

	return newResponse, nil
}

func (m *historyManagerImpl) deserializeToken(initialFirstEventID int64, nextPageToken []byte, initialLastEventBatchVersion int64) (*historyToken, error) {
	token := &historyToken{
		LastEventBatchVersion: initialLastEventBatchVersion,
		LastEventID:           initialFirstEventID - 1,
	}

	if len(nextPageToken) == 0 {
		return token, nil
	}

	err := json.Unmarshal(nextPageToken, token)
	if err == nil {
		return token, nil
	}

	// for backward compatible reason, the input data can be raw Cassandra token
	token.Data = nextPageToken
	return token, nil
}

func (m *historyManagerImpl) serializeToken(token *historyToken, nextEventID int64) ([]byte, error) {
	if token.LastEventID+1 >= nextEventID || len(token.Data) == 0 {
		return nil, nil
	}
	data, err := json.Marshal(token)
	if err != nil {
		return nil, &workflow.InternalServiceError{Message: "Error generating history event token."}
	}
	return data, nil
}

func (m *historyManagerImpl) DeleteWorkflowExecutionHistory(request *DeleteWorkflowExecutionHistoryRequest) error {
	return m.persistence.DeleteWorkflowExecutionHistory(request)
}

func (m *historyManagerImpl) Close() {
	m.persistence.Close()
}
