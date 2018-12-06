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

package replicator

import (
	"context"

	"github.com/uber-common/bark"
	"github.com/uber/cadence/.gen/go/history"
	internal "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/client/frontend"
	h "github.com/uber/cadence/client/history"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/errors"
	"github.com/uber/cadence/common/logging"
	"github.com/uber/cadence/common/persistence"
	external "go.uber.org/cadence/.gen/go/shared"
)

var (
	// ErrEmptyHistoryRawEventBatch indicate that one single batch of history raw events is of size 0
	ErrEmptyHistoryRawEventBatch = errors.NewInternalFailureError("encounter empty history batch")

	// ErrNoHistoryRawEventBatches indicate that number of batches of history raw events is of size 0
	ErrNoHistoryRawEventBatches = errors.NewInternalFailureError("no history batches are returned")

	// ErrFristHistoryRawEventBatch indicate that first batch of history raw events is malformed
	ErrFristHistoryRawEventBatch = errors.NewInternalFailureError("encounter malformed first history batch")

	// ErrUnknownEncodingType indicate that the encoding type is unknown
	ErrUnknownEncodingType = errors.NewInternalFailureError("unknown encoding type")
)

type (
	// HistoryRereplicator is the interface for resending history events to remote
	HistoryRereplicator interface {
		sendMultiWorkflowHistory(domainID string, workflowID string,
			beginingRunID string, beginingFirstEventID int64, endingRunID string, endingNextEventID int64) error
	}

	// HistoryRereplicatorImpl is the implementation of HistoryRereplicator
	HistoryRereplicatorImpl struct {
		domainCache    cache.DomainCache
		frontendClient frontend.Client
		historyClient  h.Client
		serializer     persistence.HistorySerializer
		logger         bark.Logger
	}
)

// NewHistoryRereplicator create a new HistoryRereplicatorImpl
func NewHistoryRereplicator(domainCache cache.DomainCache, frontendClient frontend.Client, historyClient h.Client,
	serializer persistence.HistorySerializer, logger bark.Logger) *HistoryRereplicatorImpl {

	return &HistoryRereplicatorImpl{
		domainCache:    domainCache,
		frontendClient: frontendClient,
		historyClient:  historyClient,
		serializer:     serializer,
		logger:         logger,
	}
}

func (h *HistoryRereplicatorImpl) sendMultiWorkflowHistory(domainID string, workflowID string,
	beginingRunID string, beginingFirstEventID int64, endingRunID string, endingNextEventID int64) (err error) {
	// NOTE: begining run ID and ending run ID can be different
	// the logic will try to grab events from begining run ID, first event ID to ending run ID, ending next event ID

	if beginingRunID == endingRunID {
		_, err = h.sendSingleWorkflowHistory(domainID, workflowID, beginingRunID, beginingFirstEventID, endingNextEventID)
		return err
	}

	// beginingRunID != endingRunID
	// this can be caused by continue as new, or simply just workflow ID reuse
	eventIDRange := func(currentRunID string, beginingRunID string, beginingFirstEventID int64,
		endingRunID string, endingNextEventID int64) (int64, int64) {

		if beginingRunID == endingRunID {
			return beginingFirstEventID, endingNextEventID
		}

		// beginingRunID != endingRunID

		if currentRunID == beginingRunID {
			// return all events from beginingFirstEventID to the end
			return beginingFirstEventID, common.EndEventID
		}

		if currentRunID == endingRunID {
			// return all events from the begining to endingNextEventID
			return common.FirstEventID, endingNextEventID
		}

		// for everything else, just dump the emtire history
		return common.FirstEventID, common.EndEventID
	}

	// the beginingRunID can be empty, if there is no workflow in DB
	runID := beginingRunID
	for len(runID) != 0 && runID != endingRunID {
		firstEventID, nextEventID := eventIDRange(runID, beginingRunID, beginingFirstEventID, endingRunID, endingNextEventID)
		runID, err = h.sendSingleWorkflowHistory(domainID, workflowID, runID, firstEventID, nextEventID)
		if err != nil {
			return err
		}
	}

	if runID == endingRunID {
		_, err = h.sendSingleWorkflowHistory(domainID, workflowID, runID, common.FirstEventID, endingNextEventID)
		return err
	}

	// we have runID being empty string
	// this means that beginingRunID to endingRunID does not have a continue as new relation ship
	// such as beginingRunID -> runID1, then runID2 (no continued as new), then runID3 -> endingRunID
	// need to work backwords using endingRunID,
	// moreover, in the above case, runID2 cannot be replicated since no one points to it

	// runIDs to be use to resend history, in reverse order
	runIDs := []string{endingRunID}
	runID = endingRunID
	for len(runID) != 0 {
		runID, err = h.getPrevRunID(domainID, workflowID, runID)
		if err != nil {
			return err
		}
		runIDs = append(runIDs, runID)
	}

	// the last runID append in the array is empty
	// for all the runIDs, send the history
	for index := len(runIDs) - 2; index > -1; index-- {
		runID = runIDs[index]
		firstEventID, nextEventID := eventIDRange(runID, beginingRunID, beginingFirstEventID, endingRunID, endingNextEventID)
		_, err = h.sendSingleWorkflowHistory(domainID, workflowID, runID, firstEventID, nextEventID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *HistoryRereplicatorImpl) sendSingleWorkflowHistory(domainID string, workflowID string, runID string,
	firstEventID int64, nextEventID int64) (string, error) {

	if firstEventID == nextEventID {
		// this is handling the case where the first event of a workflow triggers a resend
		return "", nil
	}

	var request *history.ReplicateRawEventsRequest // pending replication request to history, initialized to nil

	// branch token, event store version, replication
	// for each request to history
	var branchToken []byte
	var eventStoreVersion int32
	var replicationInfo map[string]*internal.ReplicationInfo

	pageSize := int32(100)
	var token []byte
	for doPaging := true; doPaging; doPaging = len(token) > 0 {
		response, err := h.getHistory(domainID, workflowID, runID, branchToken, firstEventID, nextEventID, token, pageSize)
		if err != nil {
			return "", err
		}

		branchToken = response.BranchToken
		eventStoreVersion = response.GetEventStoreVersion()
		replicationInfo = h.replicationInfoFromPublic(response.ReplicationInfo)

		for _, externalBatch := range response.HistoryBatches {
			err := h.sendReplicationRawRequest(request)
			if err != nil {
				return "", err
			}
			internalBatch, err := h.dataBlobFromPublic(externalBatch)
			if err != nil {
				return "", err
			}
			request = h.createReplicationRawRequest(domainID, workflowID, runID, internalBatch, eventStoreVersion, replicationInfo)
		}
	}
	// after this for loop, where shall be one request not sent yet
	// this request contains the last event, possible continue as new event
	lastBatch := request.History
	nextRunID, err := h.getNextRunID(lastBatch)
	if err != nil {
		return "", err
	}
	if len(nextRunID) > 0 {
		// last event is continue as new
		// we need to do something special for that
		var token []byte
		pageSize := int32(1)
		var branckToken []byte
		response, err := h.getHistory(domainID, workflowID, nextRunID, branckToken, common.FirstEventID, common.EndEventID, token, pageSize)
		if err != nil {
			return "", err
		}

		externalBatch := response.HistoryBatches[0]
		internalBatch, err := h.dataBlobFromPublic(externalBatch)
		if err != nil {
			return "", err
		}

		request.NewRunHistory = internalBatch
		request.NewRunEventStoreVersion = response.EventStoreVersion
	}

	return nextRunID, h.sendReplicationRawRequest(request)
}

func (h *HistoryRereplicatorImpl) createReplicationRawRequest(
	domainID string, workflowID string, runID string,
	historyBlob *internal.DataBlob,
	eventStoreVersion int32,
	replicationInfo map[string]*internal.ReplicationInfo,
) *history.ReplicateRawEventsRequest {

	request := &history.ReplicateRawEventsRequest{
		DomainUUID: common.StringPtr(domainID),
		WorkflowExecution: &internal.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
		ReplicationInfo:   replicationInfo,
		History:           historyBlob,
		EventStoreVersion: common.Int32Ptr(eventStoreVersion),
		// NewRunHistory this will be handled separately
		// NewRunEventStoreVersion  this will be handled separately
	}

	return request
}

func (h *HistoryRereplicatorImpl) sendReplicationRawRequest(request *history.ReplicateRawEventsRequest) error {

	if request == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), replicationTimeout)
	defer cancel()
	err := h.historyClient.ReplicateRawEvents(ctx, request)
	if err != nil {
		h.logger.WithFields(bark.Fields{
			logging.TagDomainID:            request.GetDomainUUID(),
			logging.TagWorkflowExecutionID: request.WorkflowExecution.GetWorkflowId(),
			logging.TagWorkflowRunID:       request.WorkflowExecution.GetRunId(),
			logging.TagErr:                 err,
		}).Error("error sending history")
	}
	return err
}

func (h *HistoryRereplicatorImpl) getHistory(domainID string, workflowID string, runID string, branchToken []byte,
	firstEventID int64, nextEventID int64, token []byte, pageSize int32) (*external.GetWorkflowExecutionRawHistoryResponse, error) {

	logger := h.logger.WithFields(bark.Fields{
		logging.TagDomainID:            domainID,
		logging.TagWorkflowExecutionID: workflowID,
		logging.TagWorkflowRunID:       runID,
		logging.TagFirstEventID:        firstEventID,
		logging.TagNextEventID:         nextEventID,
	})

	domainEntry, err := h.domainCache.GetDomainByID(domainID)
	if err != nil {
		logger.WithField(logging.TagErr, err).Error("error getting domain")
		return nil, err
	}
	domainName := domainEntry.GetInfo().Name

	ctx, cancel := context.WithTimeout(context.Background(), replicationTimeout)
	defer cancel()
	response, err := h.frontendClient.GetWorkflowExecutionRawHistory(ctx, &external.GetWorkflowExecutionRawHistoryRequest{
		Domain: common.StringPtr(domainName),
		Execution: &external.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
		BranchToken:     branchToken,
		FirstEventId:    common.Int64Ptr(firstEventID),
		NextEventId:     common.Int64Ptr(nextEventID),
		MaximumPageSize: common.Int32Ptr(pageSize),
		NextPageToken:   token,
	})

	if err != nil {
		logger.WithField(logging.TagErr, err).Error("error getting history")
		return nil, err
	}
	if len(response.HistoryBatches) == 0 {
		logger.WithField(logging.TagErr, ErrNoHistoryRawEventBatches).Error(ErrNoHistoryRawEventBatches.Error())
		return nil, ErrNoHistoryRawEventBatches
	}

	return response, nil
}

func (h *HistoryRereplicatorImpl) getPrevRunID(domainID string, workflowID string, runID string) (string, error) {

	var branchToken []byte // use nil meaning that do not care whetehr the branch has changed or not
	var token []byte       // use nil since we are only getting the first event batch, for the start event
	pageSize := int32(1)
	response, err := h.getHistory(domainID, workflowID, runID, branchToken, common.FirstEventID, common.EndEventID, token, pageSize)
	if err != nil {
		return "", err
	}

	blob, err := h.dataBlobFromPublic(response.HistoryBatches[0])
	if err != nil {
		return "", err
	}
	historyEvents, err := h.deserializeBlob(blob)
	if err != nil {
		return "", err
	}

	firstEvent := historyEvents[0]
	attr := firstEvent.WorkflowExecutionStartedEventAttributes
	if attr == nil {
		// malformed first event batch
		return "", ErrFristHistoryRawEventBatch
	}
	return attr.GetContinuedExecutionRunId(), nil
}

func (h *HistoryRereplicatorImpl) getNextRunID(blob *internal.DataBlob) (string, error) {

	historyEvents, err := h.deserializeBlob(blob)
	if err != nil {
		return "", err
	}

	lastEvent := historyEvents[len(historyEvents)-1]
	attr := lastEvent.WorkflowExecutionContinuedAsNewEventAttributes
	if attr == nil {
		// either workflow has not finished, or finished but not continue as new
		return "", nil
	}
	return attr.GetNewExecutionRunId(), nil
}

func (h *HistoryRereplicatorImpl) deserializeBlob(blob *internal.DataBlob) ([]*internal.HistoryEvent, error) {

	if blob.GetEncodingType() != internal.EncodingTypeThriftRW {
		return nil, ErrUnknownEncodingType
	}
	historyEvents, err := h.serializer.DeserializeBatchEvents(&persistence.DataBlob{
		Encoding: common.EncodingTypeThriftRW,
		Data:     blob.Data,
	})
	if err != nil {
		return nil, err
	}
	if len(historyEvents) == 0 {
		return nil, ErrEmptyHistoryRawEventBatch
	}
	return historyEvents, nil
}

// the conversion functions below exists due to the fact that there are 2 (same) API defination,
// one in the service, one on the client. although the definition are the same,
// the name prefix are different

func (h *HistoryRereplicatorImpl) dataBlobFromPublic(blob *external.DataBlob) (*internal.DataBlob, error) {
	switch blob.GetEncodingType() {
	case external.EncodingTypeThriftRW:
		return &internal.DataBlob{
			EncodingType: internal.EncodingTypeThriftRW.Ptr(),
			Data:         blob.Data,
		}, nil
	default:
		return nil, ErrUnknownEncodingType
	}
}

func (h *HistoryRereplicatorImpl) dataBlobToPublic(blob *internal.DataBlob) (*external.DataBlob, error) {
	switch blob.GetEncodingType() {
	case internal.EncodingTypeThriftRW:
		return &external.DataBlob{
			EncodingType: external.EncodingTypeThriftRW.Ptr(),
			Data:         blob.Data,
		}, nil
	default:
		return nil, ErrUnknownEncodingType
	}
}

func (h *HistoryRereplicatorImpl) replicationInfoFromPublic(in map[string]*external.ReplicationInfo) map[string]*internal.ReplicationInfo {
	out := map[string]*internal.ReplicationInfo{}
	for k, v := range in {
		out[k] = &internal.ReplicationInfo{
			Version:     v.Version,
			LastEventId: v.LastEventId,
		}
	}
	return out
}

func (h *HistoryRereplicatorImpl) replicationInfoToPublic(in map[string]*internal.ReplicationInfo) map[string]*external.ReplicationInfo {
	out := map[string]*external.ReplicationInfo{}
	for k, v := range in {
		out[k] = &external.ReplicationInfo{
			Version:     v.Version,
			LastEventId: v.LastEventId,
		}
	}
	return out
}
