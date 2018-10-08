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

package cassandra

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/uber-common/bark"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service/config"
)

const (
	templateAppendHistoryEvents = `INSERT INTO events (` +
		`domain_id, workflow_id, run_id, first_event_id, event_batch_version, range_id, tx_id, data, data_encoding) ` +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) IF NOT EXISTS`

	templateOverwriteHistoryEvents = `UPDATE events ` +
		`SET event_batch_version = ?, range_id = ?, tx_id = ?, data = ?, data_encoding = ? ` +
		`WHERE domain_id = ? AND workflow_id = ? AND run_id = ? AND first_event_id = ? ` +
		`IF range_id <= ? AND tx_id < ?`

	templateGetWorkflowExecutionHistory = `SELECT first_event_id, event_batch_version, data, data_encoding FROM events ` +
		`WHERE domain_id = ? ` +
		`AND workflow_id = ? ` +
		`AND run_id = ? ` +
		`AND first_event_id >= ? ` +
		`AND first_event_id < ?`

	templateDeleteWorkflowExecutionHistory = `DELETE FROM events ` +
		`WHERE domain_id = ? ` +
		`AND workflow_id = ? ` +
		`AND run_id = ? `

	//Below are V2 templates
	v2templateInsertNode = `INSERT INTO events_v2 (` +
		`tree_id, branch_id, row_type, ancestors, deleted, node_id, txn_id, data, data_encoding) ` +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) IF NOT EXISTS `

	v2templateOverrideNode = `UPDATE events_v2 ` +
		`SET txn_id = ?, data = ?, data_encoding = ? ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id = ? ` +
		`IF txn_id < ? `

	v2templateReadNode = `SELECT tree_id, branch_id, row_type, ancestors, deleted, node_id, txn_id, data, data_encoding FROM events_v2 ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id >= ? AND node_id < ? `

	v2templateUpdateTreeRoot = `UPDATE events_v2 ` +
		`SET txn_id = ? ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id = ? ` +
		`IF txn_id = ? `

	v2templateUpdateBranch = `UPDATE events_v2 ` +
		`SET deleted = true ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id = ? ` +
		`IF deleted = false `

	v2templateDeleteNodes = `DELETE FROM events_v2 ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id >= ? AND node_id < ?`

	v2templateDeleteRoot = `DELETE FROM events_v2 ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id = ? ` +
		`IF txn_id =? `
)

const (
	// fake nodeID for branch record
	branchNodeID = -1
	// nodeID of the tree root
	rootNodeID = 0
	// constant branchID for the tree root
	rootNodeBranchId = "10000000-0000-f000-f000-000000000000"
	// the initial txn_id of each node(including root node)
	initialTransactionID = 0

	// Row types for table events_v2
	rowTypeHistoryBranch = iota
	rowTypeHistoryNode
)

type (
	cassandraHistoryPersistence struct {
		cassandraStore
	}
)

// newHistoryPersistence is used to create an instance of HistoryManager implementation
func newHistoryPersistence(cfg config.Cassandra, logger bark.Logger) (p.HistoryStore,
	error) {
	cluster := NewCassandraCluster(cfg.Hosts, cfg.Port, cfg.User, cfg.Password, cfg.Datacenter)
	cluster.Keyspace = cfg.Keyspace
	cluster.ProtoVersion = cassandraProtoVersion
	cluster.Consistency = gocql.LocalQuorum
	cluster.SerialConsistency = gocql.LocalSerial
	cluster.Timeout = defaultSessionTimeout
	if cfg.MaxConns > 0 {
		cluster.NumConns = cfg.MaxConns
	}
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return &cassandraHistoryPersistence{cassandraStore: cassandraStore{session: session, logger: logger}}, nil
}

// Close gracefully releases the resources held by this object
func (h *cassandraHistoryPersistence) Close() {
	if h.session != nil {
		h.session.Close()
	}
}

// NewHistoryBranch creates a new branch from tree root. If tree doesn't exist, then create one. Return error if the branch already exists.
func (h *cassandraHistoryPersistence) NewHistoryBranch(request *p.NewHistoryBranchRequest) (*p.NewHistoryBranchResponse, error) {
	treeID := request.BranchInfo.TreeID
	branchID := request.BranchInfo.BranchID
	isNewTree, err := h.createRoot(treeID)
	if err != nil {
		return nil, err
	}
	resp := &p.NewHistoryBranchResponse{IsNewTree: isNewTree}
	txnID, err := h.getTreeTransactionID(treeID)
	if err != nil {
		return nil, err
	}
	batch := h.session.NewBatch(gocql.LoggedBatch)
	h.updateTreeTransactionID(batch, treeID, txnID, txnID+1)

	batch.Query(v2templateInsertNode,
		treeID, branchID, rowTypeHistoryBranch, nil, false, branchNodeID, nil, nil, nil)

	previous := make(map[string]interface{})
	applied, iter, err := h.session.MapExecuteBatchCAS(batch, previous)
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()

	if err != nil {
		if isTimeoutError(err) {
			// Write may have succeeded, but we don't know
			// return this info to the caller so they have the option of trying to find out by executing a read
			return nil, &p.TimeoutError{Msg: fmt.Sprintf("NewHistoryBranch timed out. Error: %v", err)}
		} else if isThrottlingError(err) {
			return nil, &workflow.ServiceBusyError{
				Message: fmt.Sprintf("NewHistoryBranch operation failed. Error: %v", err),
			}
		}
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("NewHistoryBranch operation failed. Error: %v", err),
		}
	}

	if !applied {
		return nil, h.getNewHistoryBranchFailure(previous, iter, txnID, treeID, branchID)
	}

	return resp, nil
}

func (h *cassandraHistoryPersistence) getNewHistoryBranchFailure(previous map[string]interface{}, iter *gocql.Iter, reqTxnID int64, reqTreeID, reqBranchID string) error {
	//if not applied, then there are two possibilities:
	// 1. tree transactionID condition fails
	// 2. the branch already exists

	txnIDUnmatch := false
	actualTxnID := int64(-1)
	branchAlreadyExits := false
	allPrevious := []map[string]interface{}{}

GetFailureReasonLoop:
	for {
		rowType, ok := previous["row_type"].(int)
		if !ok {
			// This should never happen, as all our rows have the type field.
			break GetFailureReasonLoop
		}

		treeID := previous["tree_id"].(gocql.UUID).String()
		branchID := previous["branch_id"].(gocql.UUID).String()
		nodeID := previous["node_id"].(int64)

		if rowType == rowTypeHistoryNode && treeID == reqTreeID && branchID == rootNodeBranchId && nodeID == rootNodeID {
			actualTxnID = previous["txn_id"].(int64)
			if actualTxnID != reqTxnID {
				txnIDUnmatch = true
			}
		} else if rowType == rowTypeHistoryBranch && treeID == reqTreeID && branchID == reqBranchID {
			branchAlreadyExits = true
		}

		allPrevious = append(allPrevious, previous)
		previous = make(map[string]interface{})
		if !iter.MapScan(previous) {
			break GetFailureReasonLoop
		}
	}

	if txnIDUnmatch || branchAlreadyExits {
		return &p.ConditionFailedError{
			Msg: fmt.Sprintf("Failed to create a new branch. txnIDUnmatch: %v, branchAlreadyExits: %v . Request txn_id: %v, Actual Value: %v",
				txnIDUnmatch, branchAlreadyExits, reqTxnID, actualTxnID),
		}
	}

	// At this point we only know that the write was not applied.
	var columns []string
	columnID := 0
	for _, previous := range allPrevious {
		for k, v := range previous {
			columns = append(columns, fmt.Sprintf("%v: %s=%v", columnID, k, v))
		}
		columnID++
	}
	return &p.ConditionFailedError{
		Msg: fmt.Sprintf("Failed to create a new branch. Request txn_id: %v, Actual Value: %v, columns: (%v)",
			reqTxnID, actualTxnID, columns),
	}
}

func (h *cassandraHistoryPersistence) updateTreeTransactionID(batch *gocql.Batch, treeID string, currentTxnID int64, NextTxnID int64) {
	batch.Query(v2templateUpdateTreeRoot,
		NextTxnID, treeID, rootNodeBranchId, rowTypeHistoryNode, rootNodeID, currentTxnID)
}

func (h *cassandraHistoryPersistence) getTreeTransactionID(treeID string) (int64, error) {
	query := h.session.Query(v2templateReadNode,
		treeID,
		rootNodeBranchId,
		rowTypeHistoryNode,
		rootNodeID,
		rootNodeID+1)

	result := make(map[string]interface{})
	if err := query.MapScan(result); err != nil {
		if err == gocql.ErrNotFound {
			return 0, &workflow.EntityNotExistsError{
				Message: fmt.Sprintf("getTreeTransactionID failed.  TreeID: %v Error: %v ", treeID, err),
			}
		} else if isThrottlingError(err) {
			return 0, &workflow.ServiceBusyError{
				Message: fmt.Sprintf("getTreeTransactionID failed.  TreeID: %v Error: %v", treeID, err),
			}
		}

		return 0, &workflow.InternalServiceError{
			Message: fmt.Sprintf("getTreeTransactionID failed.  TreeID: %v Error: %v", treeID, err),
		}
	}

	txnID := result["txn_id"].(int64)
	return txnID, nil
}

func (h *cassandraHistoryPersistence) getBranchStatus(treeID string, branchID string) (bool, error) {
	query := h.session.Query(v2templateReadNode,
		treeID,
		branchID,
		rowTypeHistoryBranch,
		branchNodeID,
		branchNodeID+1)

	result := make(map[string]interface{})
	if err := query.MapScan(result); err != nil {
		if err == gocql.ErrNotFound {
			return false, &workflow.EntityNotExistsError{
				Message: fmt.Sprintf("getBranchStatus failed.  TreeID: %v, branchID: %v Error: %v", treeID, branchID, err),
			}
		} else if isThrottlingError(err) {
			return false, &workflow.ServiceBusyError{
				Message: fmt.Sprintf("getBranchStatus failed.  TreeID: %v, branchID: %v Error: %v", treeID, branchID, err),
			}
		}

		return false, &workflow.InternalServiceError{
			Message: fmt.Sprintf("getBranchStatus failed.  TreeID: %v, branchID: %v Error: %v", treeID, branchID, err),
		}
	}

	deleted := result["deleted"].(bool)
	return deleted, nil
}

func (h *cassandraHistoryPersistence) createRoot(treeID string) (bool, error) {
	var query *gocql.Query

	query = h.session.Query(v2templateInsertNode,
		treeID,
		rootNodeBranchId,
		rowTypeHistoryNode,
		nil, // ancestors
		false,
		rootNodeID,
		initialTransactionID,
		nil, //data
		nil) //data_encoding

	previous := make(map[string]interface{})
	applied, err := query.MapScanCAS(previous)
	if err != nil {
		if isThrottlingError(err) {
			return false, &workflow.ServiceBusyError{
				Message: fmt.Sprintf("createRoot operation failed. Error: %v", err),
			}
		} else if isTimeoutError(err) {
			// Write may have succeeded, but we don't know
			// return this info to the caller so they have the option of trying to find out by executing a read
			return false, &p.TimeoutError{Msg: fmt.Sprintf("createRoot timed out. Error: %v", err)}
		}
		return false, &workflow.InternalServiceError{
			Message: fmt.Sprintf("createRoot operation failed. Error: %v", err),
		}
	}

	return applied, nil
}

// AppendHistoryNodes add(or override) a batch of nodes to a history branch
func (h *cassandraHistoryPersistence) AppendHistoryNodes(request *p.InternalAppendHistoryNodesRequest) error {
	if request.NextNodeIDToInsert <= 0 || request.NextNodeIDToUpdate <= 0 {
		return &p.InvalidPersistenceRequestError{
			Msg: fmt.Sprintf("NextNodeIDToInsert and NextNodeIDToUpdate must be greater than zero. Actual: %v, %v", request.NextNodeIDToUpdate, request.NextNodeIDToInsert),
		}
	}
	if request.NextNodeIDToInsert-request.NextNodeIDToUpdate < int64(len(request.Events)) {
		return &p.InvalidPersistenceRequestError{
			Msg: fmt.Sprintf("No enough events to update. Actual: %v, %v, %v", request.NextNodeIDToUpdate, request.NextNodeIDToInsert, len(request.Events)),
		}
	}
	treeID := request.BranchInfo.TreeID
	branchID := request.BranchInfo.BranchID
	treeTxnID, err := h.getTreeTransactionID(treeID)
	if err != nil {
		return err
	}
	deleted, err := h.getBranchStatus(treeID, branchID)
	if err != nil {
		return err
	}
	if deleted {
		return &p.ConditionFailedError{
			Msg: fmt.Sprintf("Failed to AppendHistoryNodes, the branch is already deleted. TreeID:%v, BranchID:%v",
				treeID, branchID),
		}
	}
	batch := h.session.NewBatch(gocql.LoggedBatch)
	// NOTE, we don't increase treeTxnID here because this operation doesn't change tree status
	h.updateTreeTransactionID(batch, treeID, treeTxnID, treeTxnID)

	currIdx := 0
	// First update/override existing events
	// If NextNodeIDToUpdate == NextNodeIDToInsert, it won't do any update
	for nodeID := request.NextNodeIDToUpdate; nodeID < request.NextNodeIDToInsert; nodeID++ {
		currIdx++
		event := request.Events[currIdx]
		batch.Query(v2templateOverrideNode,
			request.TransactionID, event.Data, event.Encoding, treeID, branchID, rowTypeHistoryNode, nodeID, request.TransactionID)
	}

	// Then insert new events until we reach the last event
	nodeID := request.NextNodeIDToInsert
	for ; currIdx < len(request.Events); currIdx++ {
		event := request.Events[currIdx]
		batch.Query(v2templateInsertNode,
			treeID, branchID, rowTypeHistoryNode, nil, false, nodeID, request.TransactionID, event.Data, event.Encoding)
		nodeID++
	}

	previous := make(map[string]interface{})
	applied, iter, err := h.session.MapExecuteBatchCAS(batch, previous)
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()

	if err != nil {
		return processCommonErrors("AppendHistoryNodes", err)
	}

	if !applied {
		return h.getAppendHistoryNodesFailure(previous, iter, treeTxnID, request.TransactionID, request.NextNodeIDToInsert, treeID, branchID)
	}

	return nil
}

func processCommonErrors(operation string, err error) error {
	if isTimeoutError(err) {
		// Write may have succeeded, but we don't know
		// return this info to the caller so they have the option of trying to find out by executing a read
		return &p.TimeoutError{Msg: fmt.Sprintf("%v timed out. Error: %v", operation, err)}
	} else if isThrottlingError(err) {
		return &workflow.ServiceBusyError{
			Message: fmt.Sprintf("%v operation failed. Error: %v", operation, err),
		}
	}
	return &workflow.InternalServiceError{
		Message: fmt.Sprintf("%v operation failed. Error: %v", operation, err),
	}
}

func (h *cassandraHistoryPersistence) getAppendHistoryNodesFailure(previous map[string]interface{}, iter *gocql.Iter, reqTreeTxnID, reqEventTxnID, nextNodeIDToInsert int64, reqTreeID, reqBranchID string) error {
	//if not applied, then there are three possibilities:
	// 1. tree transactionID condition fails
	// 2. update existing fails because of txn_id
	// 3. insert new events fails because of the event_id(node_id) already exists

	treeTxnIDUnmatch := false
	actualTreeTxnID := int64(-1)
	eventTxnIDUnmatch := false
	actualEventTxnID := int64(-1) // NOTE: there can be multiple conflicting txn_id, for now we only get one of them
	eventAlreadyExits := false
	allPrevious := []map[string]interface{}{}

GetFailureReasonLoop:
	for {
		rowType, ok := previous["row_type"].(int)
		if !ok {
			// This should never happen, as all our rows have the type field.
			break GetFailureReasonLoop
		}

		treeID := previous["tree_id"].(gocql.UUID).String()
		branchID := previous["branch_id"].(gocql.UUID).String()
		nodeID := previous["node_id"].(int64)

		if rowType == rowTypeHistoryNode && treeID == reqTreeID && branchID == rootNodeBranchId && nodeID == rootNodeID {
			actualTreeTxnID = previous["txn_id"].(int64)
			if actualTreeTxnID != reqTreeTxnID {
				treeTxnIDUnmatch = true
			}
		} else if rowType == rowTypeHistoryNode && treeID == reqTreeID && branchID == branchID && nodeID < nextNodeIDToInsert {
			actualEventTxnID = previous["txn_id"].(int64)
			if actualEventTxnID != reqEventTxnID {
				eventTxnIDUnmatch = true
			}
		} else if rowType == rowTypeHistoryNode && treeID == reqTreeID && branchID == branchID && nodeID >= nextNodeIDToInsert {
			eventAlreadyExits = true
		}

		allPrevious = append(allPrevious, previous)
		previous = make(map[string]interface{})
		if !iter.MapScan(previous) {
			break GetFailureReasonLoop
		}
	}

	if treeTxnIDUnmatch || eventTxnIDUnmatch || eventAlreadyExits {
		return &p.ConditionFailedError{
			Msg: fmt.Sprintf("Failed to append history nodes. treeTxnIDUnmatch: %v, eventTxnIDUnmatch: %v, eventAlreadyExits: %v. reqTreeTxnID: %v, Actual Value: %v, Request reqEventTxnID: %v, Actual Value: %v",
				treeTxnIDUnmatch, eventTxnIDUnmatch, eventAlreadyExits, reqTreeTxnID, actualTreeTxnID, reqEventTxnID, actualEventTxnID),
		}
	}

	// At this point we only know that the write was not applied.
	var columns []string
	columnID := 0
	for _, previous := range allPrevious {
		for k, v := range previous {
			columns = append(columns, fmt.Sprintf("%v: %s=%v", columnID, k, v))
		}
		columnID++
	}
	return &p.ConditionFailedError{
		Msg: fmt.Sprintf("Failed to append history nodes. reqTreeTxnID: %v, Actual Value: %v, Request reqEventTxnID: %v, Actual Value: %v, columns: (%v)",
			reqTreeTxnID, actualTreeTxnID, reqEventTxnID, actualEventTxnID, columns),
	}
}

// ReadHistoryBranch returns history node data for a branch
func (h *cassandraHistoryPersistence) ReadHistoryBranch(request *p.InternalReadHistoryBranchRequest) (*p.InternalReadHistoryBranchResponse, error) {
	//TODO
	return nil, nil
}

// ForkHistoryBranch forks a new branch from a old branch
func (h *cassandraHistoryPersistence) ForkHistoryBranch(request *p.ForkHistoryBranchRequest) (*p.ForkHistoryBranchResponse, error) {
	//TODO
	return nil, nil
}

// DeleteHistoryBranch removes a branch
func (h *cassandraHistoryPersistence) DeleteHistoryBranch(request *p.DeleteHistoryBranchRequest) error {
	//TODO
	return nil
}

// GetHistoryTree returns all branch information of a tree
func (h *cassandraHistoryPersistence) GetHistoryTree(request *p.GetHistoryTreeRequest) (*p.GetHistoryTreeResponse, error) {
	//TODO
	return nil, nil
}

func (h *cassandraHistoryPersistence) AppendHistoryEvents(request *p.InternalAppendHistoryEventsRequest) error {
	var query *gocql.Query

	if request.Overwrite {
		query = h.session.Query(templateOverwriteHistoryEvents,
			request.EventBatchVersion,
			request.RangeID,
			request.TransactionID,
			request.Events.Data,
			request.Events.Encoding,
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
			request.EventBatchVersion,
			request.RangeID,
			request.TransactionID,
			request.Events.Data,
			request.Events.Encoding)
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
			return &p.TimeoutError{Msg: fmt.Sprintf("AppendHistoryEvents timed out. Error: %v", err)}
		}
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("AppendHistoryEvents operation failed. Error: %v", err),
		}
	}

	if !applied {
		return &p.ConditionFailedError{
			Msg: "Failed to append history events.",
		}
	}

	return nil
}

func (h *cassandraHistoryPersistence) GetWorkflowExecutionHistory(request *p.InternalGetWorkflowExecutionHistoryRequest) (
	*p.InternalGetWorkflowExecutionHistoryResponse, error) {
	execution := request.Execution
	query := h.session.Query(templateGetWorkflowExecutionHistory,
		request.DomainID,
		*execution.WorkflowId,
		*execution.RunId,
		request.FirstEventID,
		request.NextEventID)

	iter := query.PageSize(request.PageSize).PageState(request.NextPageToken).Iter()
	if iter == nil {
		return nil, &workflow.InternalServiceError{
			Message: "GetWorkflowExecutionHistory operation failed.  Not able to create query iterator.",
		}
	}

	found := false
	nextPageToken := iter.PageState()

	//NOTE: in this method, we need to make sure eventBatchVersion is NOT decreasing(otherwise we skip the events)
	lastEventBatchVersion := request.LastEventBatchVersion

	eventBatchVersionPointer := new(int64)
	eventBatchVersion := common.EmptyVersion

	eventBatch := &p.DataBlob{}
	history := make([]*p.DataBlob, 0)

	for iter.Scan(nil, &eventBatchVersionPointer, &eventBatch.Data, &eventBatch.Encoding) {
		found = true

		if eventBatchVersionPointer != nil {
			eventBatchVersion = *eventBatchVersionPointer
		}
		if eventBatchVersion >= lastEventBatchVersion {
			history = append(history, eventBatch)
			lastEventBatchVersion = eventBatchVersion
		}

		eventBatchVersionPointer = new(int64)
		eventBatchVersion = common.EmptyVersion
		eventBatch = &p.DataBlob{}
	}

	if err := iter.Close(); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("GetWorkflowExecutionHistory operation failed. Error: %v", err),
		}
	}

	if !found && len(request.NextPageToken) == 0 {
		// adding the check of request next token being not nil, since
		// there can be case when found == false at the very end of pagination.
		return nil, &workflow.EntityNotExistsError{
			Message: fmt.Sprintf("Workflow execution history not found.  WorkflowId: %v, RunId: %v",
				*execution.WorkflowId, *execution.RunId),
		}
	}

	response := &p.InternalGetWorkflowExecutionHistoryResponse{
		NextPageToken:         nextPageToken,
		History:               history,
		LastEventBatchVersion: lastEventBatchVersion,
	}

	return response, nil
}

func (h *cassandraHistoryPersistence) DeleteWorkflowExecutionHistory(
	request *p.DeleteWorkflowExecutionHistoryRequest) error {
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
