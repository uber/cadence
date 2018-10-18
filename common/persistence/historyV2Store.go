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

	"github.com/pborman/uuid"
	"github.com/uber-common/bark"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/codec"
	"github.com/uber/cadence/common/logging"
)

type (

	// historyManagerImpl implements HistoryManager based on HistoryStore and HistorySerializer
	historyV2ManagerImpl struct {
		serializer  HistorySerializer
		persistence HistoryV2Store
		logger      bark.Logger
		encoder     codec.BinaryEncoder
	}

	// historyV2PagingToken is used to serialize/deserialize pagination token for ReadHistoryBranchRequest
	historyV2PagingToken struct {
		LastEventVersion int64
		LastNodeID       int64
	}
)

var _ HistoryV2Manager = (*historyV2ManagerImpl)(nil)

//NewHistoryManagerImpl returns new HistoryManager
func NewHistoryV2ManagerImpl(persistence HistoryV2Store, logger bark.Logger) HistoryV2Manager {
	return &historyV2ManagerImpl{
		serializer:  NewHistorySerializer(),
		persistence: persistence,
		logger:      logger,
		encoder:     codec.NewThriftRWEncoder(),
	}
}

func (m *historyV2ManagerImpl) GetName() string {
	return m.persistence.GetName()
}

// NewHistoryBranch creates a new branch from tree root. If tree doesn't exist, then create one. Return error if the branch already exists.
func (m *historyV2ManagerImpl) NewHistoryBranch(request *NewHistoryBranchRequest) (*NewHistoryBranchResponse, error) {
	req := &InternalNewHistoryBranchRequest{
		TreeID:   request.TreeID,
		BranchID: uuid.New(),
	}
	resp, err := m.persistence.NewHistoryBranch(req)
	token, err := m.encoder.Encode(&resp.BranchInfo)
	if err != nil {
		return nil, err
	}
	return &NewHistoryBranchResponse{
		BranchToken: token,
		IsNewTree:   resp.IsNewTree,
	}, nil
}

// ForkHistoryBranch forks a new branch from a old branch
func (m *historyV2ManagerImpl) ForkHistoryBranch(request *ForkHistoryBranchRequest) (*ForkHistoryBranchResponse, error) {
	return m.persistence.ForkHistoryBranch(request)
}

// DeleteHistoryBranch removes a branch
func (m *historyV2ManagerImpl) DeleteHistoryBranch(request *DeleteHistoryBranchRequest) error {
	return m.persistence.DeleteHistoryBranch(request)
}

// GetHistoryTree returns all branch information of a tree
func (m *historyV2ManagerImpl) GetHistoryTree(request *GetHistoryTreeRequest) (*GetHistoryTreeResponse, error) {
	return m.persistence.GetHistoryTree(request)
}

// AppendHistoryNodes add(or override) a node to a history branch
func (m *historyV2ManagerImpl) AppendHistoryNodes(request *AppendHistoryNodesRequest) (*AppendHistoryNodesResponse, error) {
	var branch workflow.HistoryBranch
	err := m.encoder.Decode(request.BranchToken, &branch)
	if err != nil {
		return nil, err
	}
	if len(request.Events) == 0 {
		return nil, &InvalidPersistenceRequestError{
			Msg: fmt.Sprintf("events to be appended cannot be empty"),
		}
	}
	nodeID := *request.Events[0].EventId
	// nodeID will be the first eventID
	blob, err := m.serializer.SerializeBatchEvents(request.Events, request.Encoding)
	size := len(blob.Data)

	req := &InternalAppendHistoryNodesRequest{
		BranchInfo:    branch,
		NodeID:        nodeID,
		Events:        blob,
		TransactionID: request.TransactionID,
	}

	err = m.persistence.AppendHistoryNodes(req)

	return &AppendHistoryNodesResponse{
		Size: size,
	}, err
}

func (m *historyV2ManagerImpl) deserializePagingToken(request *ReadHistoryBranchRequest) (*historyV2PagingToken, error) {
	// first batch will start from MinEventID
	token := &historyV2PagingToken{
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

func (m *historyV2ManagerImpl) serializePagingToken(lastEventVersion, lastNodeID int64) ([]byte, error) {
	token := &historyV2PagingToken{
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
func (m *historyV2ManagerImpl) ReadHistoryBranch(request *ReadHistoryBranchRequest) (*ReadHistoryBranchResponse, error) {
	if request.PageSize <= 0 {
		return nil, &InvalidPersistenceRequestError{
			Msg: fmt.Sprintf("page size must be greater than 0: %v", request.PageSize),
		}
	}

	token, err := m.deserializePagingToken(request)
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
	response.NextPageToken, err = m.serializePagingToken(lastEventVersion, lastNodeID)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (m *historyV2ManagerImpl) Close() {
	m.persistence.Close()
}
