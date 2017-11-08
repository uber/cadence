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

	"github.com/gocql/gocql"
	"github.com/uber-common/bark"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
)

const (
	templateAppendHistoryEvents = `INSERT INTO events (` +
		`domain_id, workflow_id, run_id, first_event_id, range_id, tx_id, data, data_encoding, data_version) ` +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) IF NOT EXISTS`

	templateOverwriteHistoryEvents = `UPDATE events ` +
		`SET range_id = ?, tx_id = ?, data = ?, data_encoding = ?, data_version = ? ` +
		`WHERE domain_id = ? AND workflow_id = ? AND run_id = ? AND first_event_id = ? ` +
		`IF range_id <= ? AND tx_id < ?`

	templateGetWorkflowExecutionHistory = `SELECT first_event_id, data, data_encoding, data_version FROM events ` +
		`WHERE domain_id = ? ` +
		`AND workflow_id = ? ` +
		`AND run_id = ? ` +
		`AND first_event_id < ?`

	templateListWorkflowExecutionHistory = `SELECT first_event_id, data, data_encoding, data_version FROM events ` +
		`WHERE domain_id = ? ` +
		`AND workflow_id = ? ` +
		`AND run_id = ? ` +
		`AND first_event_id >= ? ` +
		`AND first_event_id < ?`

	templateDeleteWorkflowExecutionHistory = `DELETE FROM events ` +
		`WHERE domain_id = ? ` +
		`AND workflow_id = ? ` +
		`AND run_id = ? `
)

type (
	cassandraHistoryPersistence struct {
		session                  *gocql.Session
		logger                   bark.Logger
		historySerializerFactory HistorySerializerFactory
	}

	historyPaginationToken struct {
		// Query the history events begining from this event ID. Inclusive.
		QueryEventIDStart int64
		// Indicates that we should return events which has ID >= this ID. Inclusive.
		FilterEventIDStart int64
		// QueryEventIDStart <= FilterEventIDStart
	}
)

// NewCassandraHistoryPersistence is used to create an instance of HistoryManager implementation
func NewCassandraHistoryPersistence(hosts string, port int, user, password, dc string, keyspace string,
	numConns int, logger bark.Logger) (HistoryManager,
	error) {
	cluster := common.NewCassandraCluster(hosts, port, user, password, dc)
	cluster.Keyspace = keyspace
	cluster.ProtoVersion = cassandraProtoVersion
	cluster.Consistency = gocql.LocalQuorum
	cluster.SerialConsistency = gocql.LocalSerial
	cluster.Timeout = defaultSessionTimeout
	cluster.NumConns = numConns

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return &cassandraHistoryPersistence{
		session: session,
		logger:  logger,
		historySerializerFactory: NewHistorySerializerFactory(),
	}, nil
}

// Close gracefully releases the resources held by this object
func (h *cassandraHistoryPersistence) Close() {
	if h.session != nil {
		h.session.Close()
	}
}

func (h *cassandraHistoryPersistence) AppendHistoryEvents(request *AppendHistoryEventsRequest) error {
	var query *gocql.Query
	if request.Overwrite {
		query = h.session.Query(templateOverwriteHistoryEvents,
			request.RangeID,
			request.TransactionID,
			request.Events.Data,
			request.Events.EncodingType,
			request.Events.Version,
			request.DomainID,
			*request.Execution.WorkflowId,
			*request.Execution.RunId,
			request.FirstEventID,
			request.RangeID,
			request.TransactionID)
	} else {
		query = h.session.Query(templateAppendHistoryEvents,
			request.DomainID,
			*request.Execution.WorkflowId,
			*request.Execution.RunId,
			request.FirstEventID,
			request.RangeID,
			request.TransactionID,
			request.Events.Data,
			request.Events.EncodingType,
			request.Events.Version)
	}

	previous := make(map[string]interface{})
	applied, err := query.MapScanCAS(previous)
	if err != nil {
		if isThrottlingError(err) {
			return &workflow.ServiceBusyError{
				Message: fmt.Sprintf("AppendHistoryEvents operation failed. Error: %v", err),
			}
		} else if isTimeoutError(err) {
			// Write may have succeeded, but we don't know
			// return this info to the caller so they have the option of trying to find out by executing a read
			return &TimeoutError{Msg: fmt.Sprintf("AppendHistoryEvents timed out. Error: %v", err)}
		}
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("AppendHistoryEvents operation failed. Error: %v", err),
		}
	}

	if !applied {
		return &ConditionFailedError{
			Msg: "Failed to append history events.",
		}
	}

	return nil
}

func (h *cassandraHistoryPersistence) GetWorkflowExecutionHistory(request *GetWorkflowExecutionHistoryRequest) (
	*GetWorkflowExecutionHistoryResponse, error) {
	execution := request.Execution
	query := h.session.Query(templateGetWorkflowExecutionHistory,
		request.DomainID,
		*execution.WorkflowId,
		*execution.RunId,
		request.NextEventID)

	iter := query.PageSize(request.PageSize).PageState(request.NextPageToken).Iter()
	if iter == nil {
		return nil, &workflow.InternalServiceError{
			Message: "GetWorkflowExecutionHistory operation failed.  Not able to create query iterator.",
		}
	}

	var firstEventID int64
	var history SerializedHistoryEventBatch
	response := &GetWorkflowExecutionHistoryResponse{}
	found := false
	for iter.Scan(&firstEventID, &history.Data, &history.EncodingType, &history.Version) {
		found = true
		response.Events = append(response.Events, history)
		history = SerializedHistoryEventBatch{}
	}

	nextPageToken := iter.PageState()
	response.NextPageToken = make([]byte, len(nextPageToken))
	copy(response.NextPageToken, nextPageToken)
	if err := iter.Close(); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("GetWorkflowExecutionHistory operation failed. Error: %v", err),
		}
	}

	if !found {
		return nil, &workflow.EntityNotExistsError{
			Message: fmt.Sprintf("Workflow execution history not found.  WorkflowId: %v, RunId: %v",
				*execution.WorkflowId, *execution.RunId),
		}
	}

	return response, nil
}

func (h *cassandraHistoryPersistence) ListWorkflowExecutionHistory(request *ListWorkflowExecutionHistoryRequest) (
	*ListWorkflowExecutionHistoryResponse, error) {

	// this API requires the followings to be ALWAYS true
	// 1. token.QueryEventIDStart <= token.FilterEventIDStart < request.NextEventID
	// 2. page size > 0
	// 1 && 2 will guarantee the number of ruturned events in the response to be > 0

	execution := request.Execution
	token, err := deserializeHistoryPaginationToken(request.NextPageToken)
	if err != nil {
		return nil, err
	}
	queryEventIDStart := token.QueryEventIDStart
	filterEventIDStart := token.FilterEventIDStart
	queryEventIDEnd := request.NextEventID
	query := h.session.Query(templateListWorkflowExecutionHistory,
		request.DomainID,
		*execution.WorkflowId,
		*execution.RunId,
		queryEventIDStart,
		queryEventIDEnd)

	// here the page size means number of rows to be fetched, i.e. number of batch of events tabel rows
	iter := query.PageSize(request.PageSize + 1).PageState([]byte{}).Iter()
	if iter == nil {
		return nil, &workflow.InternalServiceError{
			Message: "GetWorkflowExecutionHistory operation failed.  Not able to create query iterator.",
		}
	}

	// the first event ID within a batch of events
	var serializedFirstEventID int64
	// serialized batch of events
	var serializedEventBatch SerializedHistoryEventBatch
	var eventIdToSerializedFirstEventIdMap = make(map[int64]int64)
	var found = false
	var events []*workflow.HistoryEvent

	for iter.Scan(&serializedFirstEventID, &serializedEventBatch.Data, &serializedEventBatch.EncodingType, &serializedEventBatch.Version) {
		found = true
		setSerializedHistoryDefaults(&serializedEventBatch)
		historySerializer, _ := h.historySerializerFactory.Get(serializedEventBatch.EncodingType)
		deserializedEvents, err1 := historySerializer.Deserialize(&serializedEventBatch)
		if err1 != nil {
			return nil, err1
		}

		for _, event := range deserializedEvents.Events {
			eventIdToSerializedFirstEventIdMap[*(event.EventId)] = serializedFirstEventID
			if *event.EventId >= filterEventIDStart {
				events = append(events, event)
			}
		}
	}

	if err := iter.Close(); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("GetWorkflowExecutionHistory2 operation failed. Error: %v", err),
		}
	}

	if !found {
		return nil, &workflow.EntityNotExistsError{
			Message: fmt.Sprintf("Workflow execution history not found.  WorkflowId: %v, RunId: %v",
				*execution.WorkflowId, *execution.RunId),
		}
	}

	if len(iter.PageState()) == 0 && queryEventIDEnd > *(events[len(events)-1].EventId)+1 {
		// this happen meaning we potentially need to update the next event ID,
		// since there is a chance that next event ID provided is > actual next event ID
		queryEventIDEnd = *(events[len(events)-1].EventId) + 1
	}

	response := &ListWorkflowExecutionHistoryResponse{}
	for _, event := range events {
		// it is possible that caller will provide a next event ID
		// which is not the event ID of any first event in a patch of events
		if *event.EventId < filterEventIDStart {
			continue
		} else if *event.EventId >= queryEventIDEnd {
			break
		} else if len(response.Events) < request.PageSize {
			response.Events = append(response.Events, event)
		} else {
			// len(response.Events) >= request.PageSize
			break
		}
	}

	token = &historyPaginationToken{}
	nextFilterEventIDStart := *(response.Events[len(response.Events)-1].EventId) + 1
	nextQueryEventIDStart, ok := eventIdToSerializedFirstEventIdMap[nextFilterEventIDStart]
	if !ok {
		nextQueryEventIDStart = nextFilterEventIDStart
	}

	if nextFilterEventIDStart >= queryEventIDEnd {
		token = nil
	} else {
		token.QueryEventIDStart = nextQueryEventIDStart
		token.FilterEventIDStart = nextFilterEventIDStart
	}
	response.NextPageToken, err = serializeHistoryPaginationToken(token)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (h *cassandraHistoryPersistence) DeleteWorkflowExecutionHistory(
	request *DeleteWorkflowExecutionHistoryRequest) error {
	execution := request.Execution
	query := h.session.Query(templateDeleteWorkflowExecutionHistory,
		request.DomainID,
		*execution.WorkflowId,
		*execution.RunId)

	err := query.Exec()
	if err != nil {
		if isThrottlingError(err) {
			return &workflow.ServiceBusyError{
				Message: fmt.Sprintf("DeleteWorkflowExecutionHistory operation failed. Error: %v", err),
			}
		}
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("DeleteWorkflowExecutionHistory operation failed. Error: %v", err),
		}
	}

	return nil
}

// sets the version and encoding types to defaults if they
// are missing from persistence. This is purely for backwards
// compatibility
func setSerializedHistoryDefaults(history *SerializedHistoryEventBatch) {
	if history.Version == 0 {
		history.Version = GetDefaultHistoryVersion()
	}
	if len(history.EncodingType) == 0 {
		history.EncodingType = DefaultEncodingType
	}
}

func deserializeHistoryPaginationToken(bytes []byte) (*historyPaginationToken, error) {
	token := &historyPaginationToken{}
	if bytes == nil {
		token.QueryEventIDStart = 1
		token.FilterEventIDStart = 1
		return token, nil
	}

	err := json.Unmarshal(bytes, token)
	// TODO make this invalid next token err
	return token, err
}

func serializeHistoryPaginationToken(token *historyPaginationToken) ([]byte, error) {
	if token == nil {
		return nil, nil
	}
	bytes, err := json.Marshal(token)
	// TODO make this invalid next token err
	return bytes, err
}
