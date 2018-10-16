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

	// historyV2Token is used to serialize/deserialize pagination token for ReadHistoryBranchRequest
	historyV2Token struct {
		LastEventVersion int64
		LastNodeID       int64
	}

	// historyToken is used to serialize/deserialize pagination token for GetWorkflowExecutionHistory
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
func (m *historyManagerImpl) NewHistoryBranch(request *NewHistoryBranchRequest) (*NewHistoryBranchResponse, error) {
	return m.persistence.NewHistoryBranch(request)
}

// AppendHistoryNodes add(or override) a node to a history branch
func (m *historyManagerImpl) AppendHistoryNodes(request *AppendHistoryNodesRequest) (*AppendHistoryNodesResponse, error) {
	if len(request.Events) == 0 {
		return nil, &InvalidPersistenceRequestError{
			Msg: fmt.Sprintf("events to be appended cannot be empty"),
		}
	}
	events := request.Events
	nextNodeID := *request.Events[0].EventId
	lastNodeID := *events[len(events)-1].EventId
	if lastNodeID-nextNodeID != int64(len(events))-1 {
		return nil, &InvalidPersistenceRequestError{
			Msg: fmt.Sprintf("eventIDs must be continuous"),
		}
	}
	eventBlobs := []*DataBlob{}
	size := 0
	lastEventVersion := int64(0)
	lastEventID := nextNodeID - 1
	for _, e := range events {
		if *e.Version < lastEventVersion {
			return nil, &InvalidPersistenceRequestError{
				Msg: fmt.Sprintf("eventVersion cannot be decreasing"),
			}
		} else {
			lastEventVersion = *e.Version
		}

		if *e.EventId != lastEventID+1 {
			return nil, &InvalidPersistenceRequestError{
				Msg: fmt.Sprintf("eventID must be continuous"),
			}
		} else {
			lastEventID = *e.EventId
		}

		b, err := m.serializer.SerializeEvent(e, request.Encoding)
		if err != nil {
			return nil, err
		}
		eventBlobs = append(eventBlobs, b)
		size += len(b.Data)
	}
	resp := &AppendHistoryNodesResponse{Size: size}

	// first try to do purely insert if not exist
	intResp, err := m.persistence.AppendHistoryNodes(&InternalAppendHistoryNodesRequest{
		BranchInfo:         request.BranchInfo,
		NextNodeIDToUpdate: nextNodeID,
		NextNodeIDToInsert: nextNodeID,
		Events:             eventBlobs,
		TransactionID:      request.TransactionID,
	})
	if err == nil {
		resp.BranchInfo = intResp.BranchInfo
		return resp, nil
	}
	if _, ok := err.(*ConditionFailedError); ok {
		// if condition fails, than try to do update on the existing nodes

		// first get the existing nodes
		readReq := &InternalReadHistoryBranchRequest{
			BranchInfo: request.BranchInfo,
			MinNodeID:  nextNodeID,
			MaxNodeID:  nextNodeID + int64(len(events)),
		}
		readResp, err := m.persistence.ReadHistoryBranch(readReq)
		if err != nil {
			return nil, err
		}
		resp.OverrideCount = len(readResp.History)

		// append with override
		intResp, err = m.persistence.AppendHistoryNodes(&InternalAppendHistoryNodesRequest{
			BranchInfo:         request.BranchInfo,
			NextNodeIDToUpdate: nextNodeID,
			NextNodeIDToInsert: nextNodeID + int64(resp.OverrideCount),
			Events:             eventBlobs,
			TransactionID:      request.TransactionID,
		})
		if err == nil {
			resp.BranchInfo = intResp.BranchInfo
		}
		return resp, err
	}
	return nil, err

}

func (m *historyManagerImpl) deserializeV2Token(request *ReadHistoryBranchRequest) (*historyV2Token, error) {
	// first batch will start from MinEventID
	token := &historyV2Token{
		LastEventVersion: request.LastEventVersion,
		LastNodeID:       request.MinEventID - 1,
	}

	if len(request.NextPageToken) == 0 {
		return token, nil
	}

	err := json.Unmarshal(request.NextPageToken, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (m *historyManagerImpl) serializeV2Token(lastEventVersion, lastNodeID int64) ([]byte, error) {
	token := &historyV2Token{
		LastEventVersion: lastEventVersion,
		LastNodeID:       lastNodeID,
	}
	data, err := json.Marshal(token)
	if err != nil {
		return nil, &workflow.InternalServiceError{Message: "Error generating history event token."}
	}
	return data, nil
}

// ReadHistoryBranch returns history node data for a branch
// Pagination is implemented here, the actual minNodeID passing to persistence layer is calculated along with token's LastNodeID
func (m *historyManagerImpl) ReadHistoryBranch(request *ReadHistoryBranchRequest) (*ReadHistoryBranchResponse, error) {
	if request.PageSize <= 0 {
		return nil, &InvalidPersistenceRequestError{
			Msg: fmt.Sprintf("page size must be greater than 0: %v", request.PageSize),
		}
	}

	token, err := m.deserializeV2Token(request)
	if err != nil {
		return nil, err
	}

	events := make([]*workflow.HistoryEvent, 0, request.PageSize)
	dataSize := 0

	minNodeID := token.LastNodeID + 1
	maxNodeID := minNodeID + int64(request.PageSize)
	if request.MaxEventID < maxNodeID {
		maxNodeID = request.MaxEventID
	}

	response := &ReadHistoryBranchResponse{
		BranchInfo:    request.BranchInfo,
		History:       events,
		NextPageToken: request.NextPageToken,
		Size:          0,
	}

	// this means that we had reached the final page
	if minNodeID >= maxNodeID {
		return response, nil
	}

	ir := &InternalReadHistoryBranchRequest{
		BranchInfo: request.BranchInfo,
		MinNodeID:  minNodeID,
		MaxNodeID:  maxNodeID,
	}
	resp, err := m.persistence.ReadHistoryBranch(ir)
	if err != nil {
		return nil, err
	}

	//NOTE: in this method, we need to make sure eventVersion is NOT decreasing(otherwise we skip the events), eventID should be continuous(otherwise return error)
	lastEventVersion := token.LastEventVersion
	lastNodeID := token.LastNodeID

	for _, b := range resp.History {
		e, err := m.serializer.DeserializeEvent(b)
		if err != nil {
			return nil, err
		}

		if *e.Version < lastEventVersion {
			//version decrease means the rest are all stale events
			break
		}
		if lastNodeID+1 != *e.EventId {
			logger := m.logger.WithFields(bark.Fields{
				logging.TagBranchID: request.BranchInfo.BranchID,
				logging.TagTreeID:   request.BranchInfo.TreeID,
			})
			logger.Error("Unexpected event ID")
			return nil, fmt.Errorf("corrupted history event batch, unexpected eventID")
		}

		lastEventVersion = *e.Version
		lastNodeID = *e.EventId
		events = append(events, e)
		dataSize += len(b.Data)
	}

	// this also means that we had reached the final page
	if len(events) == 0 {
		return response, nil
	}

	// our first nodeID should be strictly equal to first eventID
	if *events[0].EventId != minNodeID {
		logger := m.logger.WithFields(bark.Fields{
			logging.TagBranchID: request.BranchInfo.BranchID,
			logging.TagTreeID:   request.BranchInfo.TreeID,
		})
		logger.Error("Unexpected event batch")
		return nil, fmt.Errorf("corrupted history event batch, eventID does not match with nodeID")
	}

	response.Size = dataSize
	response.History = events
	response.NextPageToken, err = m.serializeV2Token(lastEventVersion, lastNodeID)
	if err != nil {
		return nil, err
	}

	return response, nil
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
	token, err := m.deserializeToken(request)
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
		Events: make([]*workflow.HistoryEvent, 0, request.PageSize),
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

func (m *historyManagerImpl) deserializeToken(request *GetWorkflowExecutionHistoryRequest) (*historyToken, error) {
	token := &historyToken{
		LastEventBatchVersion: common.EmptyVersion,
		LastEventID:           request.FirstEventID - 1,
	}

	if len(request.NextPageToken) == 0 {
		return token, nil
	}

	err := json.Unmarshal(request.NextPageToken, token)
	if err == nil {
		return token, nil
	}

	// for backward compatible reason, the input data can be raw Cassandra token
	token.Data = request.NextPageToken
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
