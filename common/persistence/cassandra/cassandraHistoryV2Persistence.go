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

	"sort"

	"github.com/gocql/gocql"
	"github.com/uber-common/bark"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service/config"
)

const (
	// Belows are V2 templates
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

	v2templateReadRowType = `SELECT branch_id, ancestors, deleted FROM events_v2 WHERE tree_id = ? AND row_type = ? `

	v2templateReadOneNode = `SELECT ancestors, deleted, txn_id FROM events_v2 ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id = ? `

	v2templateReadNodes = `SELECT node_id, data, data_encoding FROM events_v2 ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id >= ? AND node_id < ? `

	v2templateUpdateTreeRoot = `UPDATE events_v2 ` +
		`SET txn_id = ? ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id = ? ` +
		`IF txn_id = ? `

	v2templateValidateBranchStatus = `UPDATE events_v2 ` +
		`SET deleted = false ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id = ? if deleted = false `

	v2templateMarkBranchDeleted = `UPDATE events_v2 ` +
		`SET deleted = true ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id = ? `

	// NOTE: Range deletion in batch is not supported on our production version  of Cassandra, we have to workaround by deleting one by one in a batch
	//v2templateRangeDeleteNodes = `DELETE FROM events_v2 ` +
	//	`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id >= ? `

	// to workaround the above issue, we need to know what is the maxium node_id in a branch
	v2templateGetMaxNodeID = `select max(node_id) AS max_id FROM events_v2 ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id >= ? `

	v2templateDeleteOneNode = `DELETE FROM events_v2 ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id = ? `

	v2templateDeleteRoot = `DELETE FROM events_v2 ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id = ? ` +
		`IF txn_id = ? `

	// Assume that we won't have branches more than that in a single tree
	// This assumption simplifies our code here.
	maxBranchesReturnForOneTree = 100
)

const (
	// fake nodeID for branch record
	branchNodeID = -1
	// nodeID of the tree root
	rootNodeID = 0
	// constant branchID for the tree root
	rootNodeBranchID = "10000000-0000-f000-f000-000000000000"
	// the initial txn_id of each node(including root node)
	initialTransactionID = 0

	// Row types for table events_v2
	rowTypeHistoryBranch = 0
	rowTypeHistoryNode   = 1
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

// We have two types of transactions for eventsV2 APIs: read/write trasaction.
// Like RWLock, read will not conflict with any other read, but write will conflict with other read/write:
// 		When Read/Write happens concurrently, write will succeed but read will fail if write committed before read.
// 		When Write/Write happens concurrently, only one of them will succeed.
// WriteTransaction is used by: NewBranch/ForkBranch/MarkBranchAsDeleted
// ReadTransaction is used by: AppendNodes/DeleteDataNode
// ReadHistoryBranch doesn't do any transaction

// write transaction will increase the txn_id of tree(root node) by one
func (h *cassandraHistoryPersistence) beginWriteTransaction(batch *gocql.Batch, treeID string, currentTxnID int64) {
	batch.Query(v2templateUpdateTreeRoot,
		currentTxnID+1, treeID, rootNodeBranchID, rowTypeHistoryNode, rootNodeID, currentTxnID)
}

// read transaction will not change the txn_id of tree(root node)
func (h *cassandraHistoryPersistence) beginReadTransaction(batch *gocql.Batch, treeID string, currentTxnID int64) {
	batch.Query(v2templateUpdateTreeRoot,
		currentTxnID, treeID, rootNodeBranchID, rowTypeHistoryNode, rootNodeID, currentTxnID)
}

// prepareTreeTransaction is an operation required for any read/write transaction
func (h *cassandraHistoryPersistence) prepareTreeTransaction(treeID string) (int64, error) {
	query := h.session.Query(v2templateReadOneNode,
		treeID,
		rootNodeBranchID,
		rowTypeHistoryNode,
		rootNodeID)

	result := make(map[string]interface{})
	if err := query.MapScan(result); err != nil {
		return 0, convertCommonErrors("prepareTreeTransaction", err)
	}

	txnID := result["txn_id"].(int64)
	return txnID, nil
}

// Close gracefully releases the resources held by this object
func (h *cassandraHistoryPersistence) Close() {
	if h.session != nil {
		h.session.Close()
	}
}

func convertCommonErrors(operation string, err error) error {
	if err == gocql.ErrNotFound {
		return &workflow.EntityNotExistsError{
			Message: fmt.Sprintf("%v failed, %v. Error: %v ", operation, err),
		}
	} else if isTimeoutError(err) {
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

// NewHistoryBranch creates a new branch from tree root. If tree doesn't exist, then create one. Return error if the branch already exists.
func (h *cassandraHistoryPersistence) NewHistoryBranch(request *p.NewHistoryBranchRequest) (*p.NewHistoryBranchResponse, error) {
	treeID := request.BranchInfo.TreeID
	branchID := request.BranchInfo.BranchID
	isNewTree, err := h.createRoot(treeID)
	if err != nil {
		return nil, err
	}
	resp := &p.NewHistoryBranchResponse{
		IsNewTree: isNewTree,
		BranchInfo: p.HistoryBranch{
			TreeID:    treeID,
			BranchID:  branchID,
			Ancestors: []p.HistoryBranchRange{},
		},
	}
	txnID, err := h.prepareTreeTransaction(treeID)
	if err != nil {
		return resp, err
	}
	batch := h.session.NewBatch(gocql.LoggedBatch)
	h.beginWriteTransaction(batch, treeID, txnID)

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
		return resp, convertCommonErrors("NewHistoryBranch", err)
	}

	if !applied {
		return resp, h.getNewHistoryBranchFailure(previous, iter, txnID, treeID, branchID)
	}

	return resp, nil
}

func (h *cassandraHistoryPersistence) getNewHistoryBranchFailure(previous map[string]interface{}, iter *gocql.Iter, reqTxnID int64, reqTreeID, reqBranchID string) error {
	//if not applied, then there are 4 possibilities:
	// 1. tree transactionID condition fails
	// 2. the branch already exists

	txnIDUnmatch := false
	actualTxnID := int64(-1)
	branchAlreadyExits := false
	allPrevious := []map[string]interface{}{}

GetFailureReasonLoop:
	for {
		allPrevious = append(allPrevious, previous)
		rowType, ok := previous["row_type"].(int64)
		if !ok {
			// This should never happen, as all our rows have the type field.
			break GetFailureReasonLoop
		}

		treeID := previous["tree_id"].(gocql.UUID).String()
		branchID := previous["branch_id"].(gocql.UUID).String()
		nodeID := previous["node_id"].(int64)

		if rowType == rowTypeHistoryNode && treeID == reqTreeID && branchID == rootNodeBranchID && nodeID == rootNodeID {
			actualTxnID = previous["txn_id"].(int64)
			if actualTxnID != reqTxnID {
				txnIDUnmatch = true
			}
		} else if rowType == rowTypeHistoryBranch && treeID == reqTreeID && branchID == reqBranchID {
			branchAlreadyExits = true
		}

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

// return error is branch is already deleted/not exists
func (h *cassandraHistoryPersistence) validateBranchStatus(batch *gocql.Batch, treeID string, branchID string) {
	batch.Query(v2templateValidateBranchStatus,
		treeID,
		branchID,
		rowTypeHistoryBranch,
		branchNodeID)
}

func (h *cassandraHistoryPersistence) createRoot(treeID string) (bool, error) {
	var query *gocql.Query

	query = h.session.Query(v2templateInsertNode,
		treeID,
		rootNodeBranchID,
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
		return false, convertCommonErrors("createRoot", err)
	}

	return applied, nil
}

// AppendHistoryNodes add(or override) a batch of nodes to a history branch
// Note that it's not allowed to override the ancestors' nodes, which means NextNodeIDToUpdate > forking point
func (h *cassandraHistoryPersistence) AppendHistoryNodes(request *p.InternalAppendHistoryNodesRequest) (*p.InternalAppendHistoryNodesResponse, error) {
	branchInfo, err := h.refillAncestors(request.BranchInfo)
	if err != nil {
		return nil, err
	}
	forkingNodeID, err := h.getForkingNode(branchInfo)
	if err != nil {
		return nil, err
	}

	if request.NextNodeIDToUpdate <= forkingNodeID || request.NextNodeIDToInsert < request.NextNodeIDToUpdate {
		return nil, &p.InvalidPersistenceRequestError{
			Msg: fmt.Sprintf("NextNodeIDToUpdate,NextNodeIDToInsert must be: forkingNodeID < NextNodeIDToUpdate <= NextNodeIDToInsert. Actual:%v, %v, %v for branch %v", forkingNodeID, request.NextNodeIDToUpdate, request.NextNodeIDToInsert, branchInfo.BranchID),
		}
	}
	if len(request.Events) == 0 {
		return nil, &p.InvalidPersistenceRequestError{
			Msg: fmt.Sprintf("events to append cannot be empty"),
		}
	}
	if request.NextNodeIDToInsert-request.NextNodeIDToUpdate > int64(len(request.Events)) {
		return nil, &p.InvalidPersistenceRequestError{
			Msg: fmt.Sprintf("no enough events to update. Actual: %v, %v, %v", request.NextNodeIDToUpdate, request.NextNodeIDToInsert, len(request.Events)),
		}
	}
	treeID := branchInfo.TreeID
	branchID := branchInfo.BranchID
	treeTxnID, err := h.prepareTreeTransaction(treeID)
	if err != nil {
		return nil, err
	}

	batch := h.session.NewBatch(gocql.LoggedBatch)
	h.beginReadTransaction(batch, treeID, treeTxnID)
	h.validateBranchStatus(batch, treeID, branchID)

	currIdx := 0
	// First update/override existing events
	// If NextNodeIDToUpdate == NextNodeIDToInsert, it won't do any update
	for nodeID := request.NextNodeIDToUpdate; nodeID < request.NextNodeIDToInsert; nodeID++ {
		event := request.Events[currIdx]
		batch.Query(v2templateOverrideNode,
			request.TransactionID, event.Data, event.Encoding, treeID, branchID, rowTypeHistoryNode, nodeID, request.TransactionID)
		currIdx++
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
		return nil, convertCommonErrors("AppendHistoryNodes", err)
	}

	if !applied {
		return nil, h.getAppendHistoryNodesFailure(previous, iter, treeTxnID, request.TransactionID, request.NextNodeIDToInsert, treeID, branchID)
	}

	return &p.InternalAppendHistoryNodesResponse{
		BranchInfo: branchInfo,
	}, nil
}

func (h *cassandraHistoryPersistence) getAppendHistoryNodesFailure(previous map[string]interface{}, iter *gocql.Iter, reqTreeTxnID, reqEventTxnID, nextNodeIDToInsert int64, reqTreeID, reqBranchID string) error {
	//if not applied, then there are 5 possibilities:
	// 1. tree transactionID condition fails
	// 2. update existing fails because of txn_id
	// 3. insert new events fails because of the event_id(node_id) already exists
	// 4. the branch is marked as deleted
	// 5. the branch doesn't exist(truly deleted)

	treeTxnIDUnmatch := false
	actualTreeTxnID := int64(-1)
	eventTxnIDUnmatch := false
	actualEventTxnID := int64(-1) // NOTE: there can be multiple conflicting txn_id, for now we only get one of them
	eventAlreadyExits := false
	allPrevious := []map[string]interface{}{}
	branchMarkedAsDeleted := false

GetFailureReasonLoop:
	for {
		allPrevious = append(allPrevious, previous)
		rowType, ok := previous["row_type"].(int64)
		if !ok {
			// This should never happen, as all our rows have the type field.
			break GetFailureReasonLoop
		}

		treeID := previous["tree_id"].(gocql.UUID).String()
		branchID := previous["branch_id"].(gocql.UUID).String()
		nodeID := previous["node_id"].(int64)
		deleted := previous["deleted"].(bool)

		if rowType == rowTypeHistoryNode && treeID == reqTreeID && branchID == rootNodeBranchID && nodeID == rootNodeID {
			actualTreeTxnID = previous["txn_id"].(int64)
			if actualTreeTxnID != reqTreeTxnID {
				treeTxnIDUnmatch = true
			}
		} else if rowType == rowTypeHistoryNode && treeID == reqTreeID && branchID == reqBranchID {
			actualEventTxnID = previous["txn_id"].(int64)
			if nodeID < nextNodeIDToInsert && actualEventTxnID >= reqEventTxnID {
				eventTxnIDUnmatch = true
			} else {
				eventAlreadyExits = true
			}
		} else if rowType == rowTypeHistoryBranch && treeID == reqTreeID && branchID == reqBranchID {
			if deleted {
				branchMarkedAsDeleted = true
			}
		}

		previous = make(map[string]interface{})
		if !iter.MapScan(previous) {
			break GetFailureReasonLoop
		}
	}

	if treeTxnIDUnmatch || eventTxnIDUnmatch || eventAlreadyExits || branchMarkedAsDeleted {
		return &p.ConditionFailedError{
			Msg: fmt.Sprintf("Failed to append history nodes. treeTxnIDUnmatch: %v, eventTxnIDUnmatch: %v, eventAlreadyExits: %v, branchMarkedAsDeleted:%v . reqTreeTxnID: %v, Actual Value: %v, Request reqEventTxnID: %v, Actual Value: %v",
				treeTxnIDUnmatch, eventTxnIDUnmatch, eventAlreadyExits, branchMarkedAsDeleted, reqTreeTxnID, actualTreeTxnID, reqEventTxnID, actualEventTxnID),
		}
	}

	// At this point we only know that the write was not applied. It is likely that the branch record doesn't exist
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

// refill the ancestors if needed(nil)
func (h *cassandraHistoryPersistence) refillAncestors(bi p.HistoryBranch) (p.HistoryBranch, error) {
	if bi.Ancestors == nil {
		allAncestors, err := h.getBranchAncestors(bi.TreeID, bi.BranchID)
		if err != nil {
			return bi, err
		}
		bi.Ancestors = allAncestors
	}
	return bi, nil
}

func (h *cassandraHistoryPersistence) getForkingNode(bi p.HistoryBranch) (int64, error) {
	bi, err := h.refillAncestors(bi)
	if err != nil {
		return 0, err
	}

	if len(bi.Ancestors) == 0 {
		// root branch
		return 0, nil
	} else {
		return bi.Ancestors[len(bi.Ancestors)-1].EndNodeID - 1, nil
	}
}

func (h *cassandraHistoryPersistence) readBranchRange(treeID, branchID string, minID, maxID int64) ([]*p.DataBlob, error) {
	bs := make([]*p.DataBlob, 0, int(maxID-minID))

	query := h.session.Query(v2templateReadNodes,
		treeID, branchID, rowTypeHistoryNode, minID, maxID)

	iter := query.PageSize(int(maxID - minID)).Iter()
	if iter == nil {
		return nil, &workflow.InternalServiceError{
			Message: "ReadHistoryBranch operation failed.  Not able to create query iterator.",
		}
	}

	eventBlob := &p.DataBlob{}

	for iter.Scan(nil, &eventBlob.Data, &eventBlob.Encoding) {
		bs = append(bs, eventBlob)
		eventBlob = &p.DataBlob{}
	}

	if err := iter.Close(); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ReadHistoryBranch. Close operation failed. Error: %v", err),
		}
	}
	return bs, nil
}

// ReadHistoryBranch returns history node data for a branch
// NOTE: For branch that has ancestors, we need to query Cassandra multiple times, because it doesn't support OR/UNION operator
func (h *cassandraHistoryPersistence) ReadHistoryBranch(request *p.InternalReadHistoryBranchRequest) (*p.InternalReadHistoryBranchResponse, error) {
	if request.MaxNodeID <= request.MinNodeID {
		return nil, &p.InvalidPersistenceRequestError{
			Msg: fmt.Sprintf("MaxNodeID %v must be greater than MinNodeID %v", request.MaxNodeID, request.MinNodeID),
		}
	}
	branchInfo, err := h.refillAncestors(request.BranchInfo)
	if err != nil {
		return nil, err
	}

	treeID := branchInfo.TreeID
	branchID := branchInfo.BranchID
	// We also query the current branch
	forkingNodeID, err := h.getForkingNode(branchInfo)
	if err != nil {
		return nil, err
	}

	allBRs := branchInfo.Ancestors
	allBRs = append(allBRs, p.HistoryBranchRange{
		BranchID:    branchID,
		BeginNodeID: forkingNodeID + 1,
		EndNodeID:   request.MaxNodeID,
	})

	history := make([]*p.DataBlob, 0, request.MaxNodeID-request.MinNodeID)
	for _, br := range allBRs {
		// this range won't contain any nodes needed, since the last node(EndNodeID-1) in the range is strictly less than MinNodeID
		if br.EndNodeID <= request.MinNodeID {
			continue
		}
		// similarly, this range won't contain any nodes needed, since the first node(BeginNodeID) in the range is greater than or equal to MaxNodeID, where MaxNodeID is exclusive for the request
		if br.BeginNodeID >= request.MaxNodeID {
			// since they are sorted, the rest branch range can be skipped
			break
		}

		minID := br.BeginNodeID
		if request.MinNodeID > minID {
			minID = request.MinNodeID
		}
		maxID := br.EndNodeID
		if request.MaxNodeID < maxID {
			maxID = request.MaxNodeID
		}
		bs, err := h.readBranchRange(treeID, br.BranchID, minID, maxID)
		if err != nil {
			return nil, err
		}
		history = append(history, bs...)
	}

	response := &p.InternalReadHistoryBranchResponse{
		History:    history,
		BranchInfo: branchInfo,
	}

	return response, nil
}

func (h *cassandraHistoryPersistence) checkIfNodeExists(treeID, branchID string, nodeID int64) (bool, error) {
	query := h.session.Query(v2templateReadOneNode,
		treeID,
		branchID,
		rowTypeHistoryNode,
		nodeID)

	result := make(map[string]interface{})
	if err := query.MapScan(result); err != nil {
		if err == gocql.ErrNotFound {
			return false, nil
		}
		return false, convertCommonErrors("checkIfNodeExists", err)

	}

	return true, nil
}

func (h *cassandraHistoryPersistence) getBranchAncestors(treeID, branchID string) ([]p.HistoryBranchRange, error) {
	query := h.session.Query(v2templateReadOneNode,
		treeID,
		branchID,
		rowTypeHistoryBranch,
		branchNodeID)

	result := make(map[string]interface{})
	if err := query.MapScan(result); err != nil {
		return nil, convertCommonErrors("getBranchAncestors", err)
	}

	eList := result["ancestors"].([]map[string]interface{})
	return h.parseBranchAncestors(eList), nil
}

func (h *cassandraHistoryPersistence) parseBranchAncestors(ancestors []map[string]interface{}) []p.HistoryBranchRange {
	ans := make([]p.HistoryBranchRange, 0, len(ancestors))
	for _, e := range ancestors {
		an := p.HistoryBranchRange{}
		for k, v := range e {
			switch k {
			case "branch_id":
				an.BranchID = v.(gocql.UUID).String()
			case "forked_node_id":
				an.EndNodeID = v.(int64) + 1
			}
		}
		ans = append(ans, an)
	}

	if len(ans) > 0 {
		// sort ans based onf EndNodeID so that we can set BeginNodeID
		sort.Slice(ans, func(i, j int) bool { return ans[i].EndNodeID < ans[j].EndNodeID })
		ans[0].BeginNodeID = 1
		for i := 1; i < len(ans); i++ {
			ans[i].BeginNodeID = ans[i-1].EndNodeID
		}
	}
	return ans
}

// ForkHistoryBranch forks a new branch from a old branch
// We will only take the ancestors up to the forking point. For example:
// Assume we have B2 forked from B1 at 10, B3 forked from B2 at 20.
//           B1 [1~10]
//           /
//         B2 [11~20]
//        /
//      B3[21~30]
// If we fork B4 from B3 at node 22/25, the new branch is like:
//           B1[1~10]
//           /
//         B2 [11~20]
//        /
//      B3[21~22/25]
//     /
//   B4[...]
// However, if we fork from node 12/15, then the new branch should be
//           B1[1~10]
//           /
//         B2 [11~12/15]
//        /
//      B4[...]
// So even though it fork from B3, its ancestor doesn't have to have B3 in its branch ranges
func (h *cassandraHistoryPersistence) ForkHistoryBranch(request *p.ForkHistoryBranchRequest) (*p.ForkHistoryBranchResponse, error) {
	forkB, err := h.refillAncestors(request.BranchInfo)
	if err != nil {
		return nil, err
	}
	treeID := forkB.TreeID
	newAncestors := make([]p.HistoryBranchRange, 0, len(forkB.Ancestors)+1)

	txnID, err := h.prepareTreeTransaction(treeID)
	if err != nil {
		return nil, err
	}

	lastForkingNodeID, err := h.getForkingNode(forkB)
	if err != nil {
		return nil, err
	}

	if lastForkingNodeID >= request.ForkFromNodeID {
		// this is the case that new branch's ancestors doesn't include the forking branch
		for _, br := range forkB.Ancestors {
			if br.EndNodeID > request.ForkFromNodeID {
				newAncestors = append(newAncestors, p.HistoryBranchRange{
					BranchID:    br.BranchID,
					BeginNodeID: br.BeginNodeID,
					EndNodeID:   request.ForkFromNodeID + 1,
				})
				break
			} else {
				newAncestors = append(newAncestors, br)
			}
		}
	} else {
		// this is the case the new branch will inherit all ancestors from forking branch

		// first we need to check if the branch has a valid nodeID to fork from
		exits, err := h.checkIfNodeExists(treeID, forkB.BranchID, request.ForkFromNodeID)
		if err != nil {
			return nil, err
		}
		if !exits {
			return nil, &p.InvalidPersistenceRequestError{
				Msg: fmt.Sprintf("forking branch %v doesn't have node %v", forkB.BranchID, request.ForkFromNodeID),
			}
		}

		newAncestors = append(newAncestors, forkB.Ancestors...)
		newAncestors = append(newAncestors, p.HistoryBranchRange{
			BranchID:    forkB.BranchID,
			BeginNodeID: lastForkingNodeID + 1,
			EndNodeID:   request.ForkFromNodeID + 1,
		})
	}

	resp := &p.ForkHistoryBranchResponse{BranchInfo: p.HistoryBranch{
		TreeID:    treeID,
		BranchID:  request.NewBranchID,
		Ancestors: newAncestors,
	}}

	batch := h.session.NewBatch(gocql.LoggedBatch)
	h.beginWriteTransaction(batch, treeID, txnID)
	h.validateBranchStatus(batch, treeID, forkB.BranchID)

	ancs := []map[string]interface{}{}
	for _, an := range newAncestors {
		value := make(map[string]interface{})
		value["forked_node_id"] = an.EndNodeID - 1
		value["branch_id"] = an.BranchID
		ancs = append(ancs, value)
	}

	batch.Query(v2templateInsertNode,
		treeID, request.NewBranchID, rowTypeHistoryBranch, ancs, false, branchNodeID, nil, nil, nil)

	previous := make(map[string]interface{})
	applied, iter, err := h.session.MapExecuteBatchCAS(batch, previous)
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()

	if err != nil {
		return nil, convertCommonErrors("ForkHistoryBranch", err)
	}

	if !applied {
		return nil, h.getForkHistoryBranchFailure(previous, iter, txnID, treeID, request.NewBranchID, forkB.BranchID)
	}

	return resp, nil
}

func (h *cassandraHistoryPersistence) getForkHistoryBranchFailure(previous map[string]interface{}, iter *gocql.Iter, reqTxnID int64, reqTreeID, newBranchID, forkBranchID string) error {
	//if not applied, then there are 4 possibilities:
	// 1. tree transactionID condition fails
	// 2. the branch already exists
	// 3. forking branch is marked as deleted
	// 4. forking branch doesn't exist(truly deleted)

	txnIDUnmatch := false
	actualTxnID := int64(-1)
	branchAlreadyExits := false
	allPrevious := []map[string]interface{}{}
	forkBranchMarkedDeleted := false

GetFailureReasonLoop:
	for {
		allPrevious = append(allPrevious, previous)
		rowType, ok := previous["row_type"].(int64)
		if !ok {
			// This should never happen, as all our rows have the type field.
			break GetFailureReasonLoop
		}

		treeID := previous["tree_id"].(gocql.UUID).String()
		branchID := previous["branch_id"].(gocql.UUID).String()
		nodeID := previous["node_id"].(int64)
		deleted := previous["deleted"].(bool)

		if rowType == rowTypeHistoryNode && treeID == reqTreeID && branchID == rootNodeBranchID && nodeID == rootNodeID {
			actualTxnID = previous["txn_id"].(int64)
			if actualTxnID != reqTxnID {
				txnIDUnmatch = true
			}
		} else if rowType == rowTypeHistoryBranch && treeID == reqTreeID && branchID == newBranchID {
			branchAlreadyExits = true
		} else if rowType == rowTypeHistoryBranch && treeID == reqTreeID && branchID == forkBranchID {
			if deleted {
				forkBranchMarkedDeleted = true
			}
		}

		previous = make(map[string]interface{})
		if !iter.MapScan(previous) {
			break GetFailureReasonLoop
		}
	}

	if txnIDUnmatch || branchAlreadyExits || forkBranchMarkedDeleted {
		return &p.ConditionFailedError{
			Msg: fmt.Sprintf("Failed to create a new branch. txnIDUnmatch: %v, branchAlreadyExits: %v , forkBranchMarkedDeleted %v. Request txn_id: %v, Actual Value: %v",
				txnIDUnmatch, branchAlreadyExits, forkBranchMarkedDeleted, reqTxnID, actualTxnID),
		}
	}

	// At this point we only know that the write was not applied. It is mostly likely that the forking branch doesn't exist
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

// DeleteHistoryBranch removes a branch
func (h *cassandraHistoryPersistence) DeleteHistoryBranch(request *p.DeleteHistoryBranchRequest) error {
	// We break the operation into three phases:
	// 1. Mark the branch as deleted
	// 2. Delete data nodes
	// 3. Delete the branch node, also delete root node if the branch is the last branch

	// this would do a write transaction
	err := h.markBranchAsDeleted(request.BranchInfo)
	if err != nil {
		return err
	}

	// this would do a read transaction
	err = h.deleteDataNodes(request.BranchInfo)
	if err != nil {
		return err
	}

	// this may do a write transaction or a special write transaction(delete the root node)
	err = h.deleteBranchAndRootNode(request.BranchInfo)
	return err
}

func (h *cassandraHistoryPersistence) markBranchAsDeleted(branch p.HistoryBranch) error {
	treeID := branch.TreeID
	txnID, err := h.prepareTreeTransaction(treeID)
	if err != nil {
		return err
	}
	batch := h.session.NewBatch(gocql.LoggedBatch)
	h.beginWriteTransaction(batch, treeID, txnID)

	batch.Query(v2templateMarkBranchDeleted,
		treeID, branch.BranchID, rowTypeHistoryBranch, branchNodeID)

	previous := make(map[string]interface{})
	applied, iter, err := h.session.MapExecuteBatchCAS(batch, previous)
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()

	if err != nil {
		return convertCommonErrors("markBranchAsDeleted", err)
	}

	if !applied {
		return h.getTreeTrasactionFailure(previous, iter, txnID, treeID, branch.BranchID)
	}
	return nil
}

func (h *cassandraHistoryPersistence) deleteDataNodes(branch p.HistoryBranch) error {
	branch, err := h.refillAncestors(branch)
	if err != nil {
		return err
	}

	treeID := branch.TreeID
	brsToDelete := branch.Ancestors
	lastEndNodeID := int64(1)
	if len(brsToDelete) > 0 {
		lastEndNodeID = brsToDelete[len(brsToDelete)-1].EndNodeID
	}
	brsToDelete = append(brsToDelete, p.HistoryBranchRange{
		BranchID:    branch.BranchID,
		BeginNodeID: lastEndNodeID,
	})

	txnID, err := h.prepareTreeTransaction(treeID)
	if err != nil {
		return err
	}
	batch := h.session.NewBatch(gocql.LoggedBatch)
	h.beginReadTransaction(batch, treeID, txnID)

	rsp, err := h.GetHistoryTree(&p.GetHistoryTreeRequest{
		TreeID: treeID,
	})
	if err != nil {
		return err
	}
	// for each branch range that is being used, we want to know what is the max nodeID referred
	validBRsEndNode := map[string]int64{}
	for _, b := range rsp.Branches {
		for _, br := range b.Ancestors {
			curr, ok := validBRsEndNode[br.BranchID]
			if !ok || curr < br.EndNodeID {
				validBRsEndNode[br.BranchID] = br.EndNodeID
			}
		}
	}

	for i := len(brsToDelete) - 1; i >= 0; i-- {
		br := brsToDelete[i]
		maxEndNodeID, ok := validBRsEndNode[br.BranchID]
		if ok {
			// we can only delete from the maxEndNode
			h.doRangeDeleteBranch(batch, treeID, br.BranchID, maxEndNodeID)
			break
		} else {
			// No any branch is using this range, we can delete all of it
			h.doRangeDeleteBranch(batch, treeID, br.BranchID, br.BeginNodeID)
		}
	}

	previous := make(map[string]interface{})
	applied, iter, err := h.session.MapExecuteBatchCAS(batch, previous)
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()

	if err != nil {
		return convertCommonErrors("deleteDataNodes", err)
	}

	if !applied {
		return h.getTreeTrasactionFailure(previous, iter, txnID, treeID, branch.BranchID)
	}
	return nil
}

// NOTE: ideally we should do
// batch.Query(v2templateRangeDeleteNodes,
//				treeID, br.BranchID, rowTypeHistoryNode, maxEndNodeID)
// However Cassandra doesn't support range deletion in our production version
func (h *cassandraHistoryPersistence) doRangeDeleteBranch(batch *gocql.Batch, treeID, branchID string, beginNodeID int64) error {
	// first we need to get the last nodeID to delete
	query := h.session.Query(v2templateGetMaxNodeID,
		treeID,
		branchID,
		rowTypeHistoryNode,
		beginNodeID)

	result := make(map[string]interface{})
	if err := query.MapScan(result); err != nil {
		if err == gocql.ErrNotFound {
			// return if the branch doesn't have any node to delete
			return nil
		}
		return convertCommonErrors("doRangeDeleteBranch", err)
	}

	lastNodeID := result["max_id"].(int64)
	for nodeID := beginNodeID; nodeID <= lastNodeID; nodeID++ {
		batch.Query(v2templateDeleteOneNode,
			treeID, branchID, rowTypeHistoryNode, nodeID)
	}
	return nil
}

func (h *cassandraHistoryPersistence) deleteBranchAndRootNode(branch p.HistoryBranch) error {
	treeID := branch.TreeID
	txnID, err := h.prepareTreeTransaction(treeID)
	if err != nil {
		return err
	}
	batch := h.session.NewBatch(gocql.LoggedBatch)

	batch.Query(v2templateDeleteOneNode,
		treeID, branch.BranchID, rowTypeHistoryBranch, branchNodeID)

	isLast, err := h.isLastBranch(branch.TreeID)
	if err != nil {
		return err
	}
	if isLast {
		batch.Query(v2templateDeleteRoot,
			treeID, rootNodeBranchID, rowTypeHistoryNode, rootNodeID, txnID)
	} else {
		h.beginWriteTransaction(batch, treeID, txnID)
	}

	previous := make(map[string]interface{})
	applied, iter, err := h.session.MapExecuteBatchCAS(batch, previous)
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()

	if err != nil {
		return convertCommonErrors("deleteBranchAndRootNode", err)
	}

	if !applied {
		return h.getTreeTrasactionFailure(previous, iter, txnID, treeID, branch.BranchID)
	}
	return nil
}

func (h *cassandraHistoryPersistence) getTreeTrasactionFailure(previous map[string]interface{}, iter *gocql.Iter, reqTxnID int64, reqTreeID, reqBranchID string) error {
	//if not applied then it is only possible that the tree's txn_id has changed

	txnIDUnmatch := false
	actualTxnID := int64(-1)
	allPrevious := []map[string]interface{}{}

GetFailureReasonLoop:
	for {
		allPrevious = append(allPrevious, previous)
		rowType, ok := previous["row_type"].(int64)
		if !ok {
			// This should never happen, as all our rows have the type field.
			break GetFailureReasonLoop
		}

		treeID := previous["tree_id"].(gocql.UUID).String()
		branchID := previous["branch_id"].(gocql.UUID).String()
		nodeID := previous["node_id"].(int64)

		if rowType == rowTypeHistoryNode && treeID == reqTreeID && branchID == rootNodeBranchID && nodeID == rootNodeID {
			actualTxnID = previous["txn_id"].(int64)
			if actualTxnID != reqTxnID {
				txnIDUnmatch = true
			}
		}

		previous = make(map[string]interface{})
		if !iter.MapScan(previous) {
			break GetFailureReasonLoop
		}
	}

	if txnIDUnmatch {
		return &p.ConditionFailedError{
			Msg: fmt.Sprintf("Failed to do a tree transation. txnIDUnmatch: %v . Request txn_id: %v, Actual Value: %v",
				txnIDUnmatch, reqTxnID, actualTxnID),
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
		Msg: fmt.Sprintf("Failed to do a tree transation. Request txn_id: %v, Actual Value: %v, columns: (%v)",
			reqTxnID, actualTxnID, columns),
	}
}

func (h *cassandraHistoryPersistence) isLastBranch(treeID string) (bool, error) {
	query := h.session.Query(v2templateReadRowType, treeID, rowTypeHistoryBranch)

	iter := query.PageSize(2).Iter()
	if iter == nil {
		return false, &workflow.InternalServiceError{
			Message: "isLastBranch operation failed.  Not able to create query iterator.",
		}
	}

	brCount := 0
	for iter.Scan(nil, nil, nil) {
		brCount++
	}

	if err := iter.Close(); err != nil {
		return false, &workflow.InternalServiceError{
			Message: fmt.Sprintf("isLastBranch. Close operation failed. Error: %v", err),
		}
	}

	return brCount == 1, nil
}

// GetHistoryTree returns all branch information of a tree
func (h *cassandraHistoryPersistence) GetHistoryTree(request *p.GetHistoryTreeRequest) (*p.GetHistoryTreeResponse, error) {
	treeID := request.TreeID

	query := h.session.Query(v2templateReadRowType, treeID, rowTypeHistoryBranch)

	iter := query.PageSize(maxBranchesReturnForOneTree).Iter()
	if iter == nil {
		return nil, &workflow.InternalServiceError{
			Message: "GetHistoryTree operation failed.  Not able to create query iterator.",
		}
	}

	branchUUID := gocql.UUID{}
	ancsResult := []map[string]interface{}{}
	deleted := false
	branches := make([]p.HistoryBranch, 0)

	for iter.Scan(&branchUUID, &ancsResult, &deleted) {
		if deleted {
			continue
		}
		ancs := h.parseBranchAncestors(ancsResult)
		b := p.HistoryBranch{
			TreeID:    treeID,
			BranchID:  branchUUID.String(),
			Ancestors: ancs,
		}
		branches = append(branches, b)

		branchUUID = gocql.UUID{}
		ancsResult = []map[string]interface{}{}
		deleted = false
	}

	if err := iter.Close(); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("GetHistoryTree. Close operation failed. Error: %v", err),
		}
	}
	if len(branches) >= maxBranchesReturnForOneTree {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("Too many branches in a tree"),
		}
	}

	return &p.GetHistoryTreeResponse{
		Branches: branches,
	}, nil
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
	history := make([]*p.DataBlob, 0, request.PageSize)

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
