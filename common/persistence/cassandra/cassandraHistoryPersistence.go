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

	"strings"

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
		`WHERE tree_id = ? AND branch_id IN ( ? ) AND row_type = ? AND node_id >= ? AND node_id < ? `

	v2templateUpdateTreeRoot = `UPDATE events_v2 ` +
		`SET txn_id = ? ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id = ? ` +
		`IF txn_id = ? `

	v2templateMarkBranchDeleted = `UPDATE events_v2 ` +
		`SET deleted = true ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id = ? `

	v2templateRangeDeleteNodes = `DELETE FROM events_v2 ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id >= ? `

	v2templateDeleteOneNode = `DELETE FROM events_v2 ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id = ? `

	v2templateDeleteRoot = `DELETE FROM events_v2 ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? AND node_id = ? ` +
		`IF txn_id = ? `

	// Assume that we won't have branches more than that in a single tree
	// This assumption simplifies our code here.
	maxBranchesReturnForOneTree = 1000
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
	resp := &p.NewHistoryBranchResponse{IsNewTree: isNewTree}
	txnID, err := h.prepareTreeTransaction(treeID)
	if err != nil {
		return nil, err
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
		return nil, convertCommonErrors("NewHistoryBranch", err)
	}

	if !applied {
		return nil, h.getInsertHistoryBranchFailure(previous, iter, txnID, treeID, branchID)
	}

	return resp, nil
}

func (h *cassandraHistoryPersistence) getInsertHistoryBranchFailure(previous map[string]interface{}, iter *gocql.Iter, reqTxnID int64, reqTreeID, reqBranchID string) error {
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

		if rowType == rowTypeHistoryNode && treeID == reqTreeID && branchID == rootNodeBranchID && nodeID == rootNodeID {
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
	return &p.UnexpectedConditionFailedError{
		Msg: fmt.Sprintf("Failed to create a new branch. Request txn_id: %v, Actual Value: %v, columns: (%v)",
			reqTxnID, actualTxnID, columns),
	}
}

// return error is branch is already deleted/not exists
func (h *cassandraHistoryPersistence) validateBranchStatus(treeID string, branchID string) error {
	query := h.session.Query(v2templateReadOneNode,
		treeID,
		branchID,
		rowTypeHistoryBranch,
		branchNodeID)

	result := make(map[string]interface{})
	if err := query.MapScan(result); err != nil {
		return convertCommonErrors("validateBranchStatus", err)
	}

	deleted := result["deleted"].(bool)
	if deleted {
		return &workflow.EntityNotExistsError{
			Message: fmt.Sprintf("branch %v of tree %v is already deleted", branchID, treeID),
		}
	}
	return nil
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
func (h *cassandraHistoryPersistence) AppendHistoryNodes(request *p.InternalAppendHistoryNodesRequest) error {
	if request.NextNodeIDToInsert <= 0 || request.NextNodeIDToUpdate <= 0 {
		return &p.InvalidPersistenceRequestError{
			Msg: fmt.Sprintf("NextNodeIDToInsert and NextNodeIDToUpdate must be greater than zero. Actual: %v, %v", request.NextNodeIDToUpdate, request.NextNodeIDToInsert),
		}
	}
	if len(request.Events) == 0 {
		return &p.InvalidPersistenceRequestError{
			Msg: fmt.Sprintf("events to append cannot be empty"),
		}
	}
	if request.NextNodeIDToInsert-request.NextNodeIDToUpdate < int64(len(request.Events)) {
		return &p.InvalidPersistenceRequestError{
			Msg: fmt.Sprintf("no enough events to update. Actual: %v, %v, %v", request.NextNodeIDToUpdate, request.NextNodeIDToInsert, len(request.Events)),
		}
	}
	treeID := request.BranchInfo.TreeID
	branchID := request.BranchInfo.BranchID
	treeTxnID, err := h.prepareTreeTransaction(treeID)
	if err != nil {
		return err
	}
	err = h.validateBranchStatus(treeID, branchID)
	if err != nil {
		return err
	}
	batch := h.session.NewBatch(gocql.LoggedBatch)
	// NOTE, we don't increase treeTxnID here because this operation doesn't change tree status
	h.beginReadTransaction(batch, treeID, treeTxnID)

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
		return convertCommonErrors("AppendHistoryNodes", err)
	}

	if !applied {
		return h.getAppendHistoryNodesFailure(previous, iter, treeTxnID, request.TransactionID, request.NextNodeIDToInsert, treeID, branchID)
	}

	return nil
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

		if rowType == rowTypeHistoryNode && treeID == reqTreeID && branchID == rootNodeBranchID && nodeID == rootNodeID {
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
	return &p.UnexpectedConditionFailedError{
		Msg: fmt.Sprintf("Failed to append history nodes. reqTreeTxnID: %v, Actual Value: %v, Request reqEventTxnID: %v, Actual Value: %v, columns: (%v)",
			reqTreeTxnID, actualTreeTxnID, reqEventTxnID, actualEventTxnID, columns),
	}
}

// ReadHistoryBranch returns history node data for a branch
func (h *cassandraHistoryPersistence) ReadHistoryBranch(request *p.InternalReadHistoryBranchRequest) (*p.InternalReadHistoryBranchResponse, error) {
	if request.MaxNodeID <= request.MinNodeID {
		return nil, &p.InvalidPersistenceRequestError{
			Msg: fmt.Sprintf("MaxNodeID %v must be greater than MinNodeID %v", request.MaxNodeID, request.MinNodeID),
		}
	}
	numOfExpectedEvents := int(request.MaxNodeID - request.MinNodeID)
	branchInfo := request.BranchInfo
	treeID := branchInfo.TreeID
	branchID := branchInfo.BranchID
	allAncestors := branchInfo.Ancestors
	var err error
	if branchInfo.Ancestors == nil {
		allAncestors, err = h.getBranchAncestors(treeID, branchID)
		if err != nil {
			return nil, err
		}
	}
	branchInfo.Ancestors = allAncestors

	branchesToQuery := make([]string, 0)
	// NOTE: theoretically, the following code can be improved by binary search. But we shouldn't have many branch ranges, see maxBranchesReturnForOneTree. So linear search is sufficient here.
	for _, an := range allAncestors {
		// this range won't contain any nodes needed, since the last node(EndNodeID-1) in the range is strictly less than MinNodeID
		if an.EndNodeID <= request.MinNodeID {
			continue
		}
		// similarly, this range won't contain any nodes needed, since the first node(BeginNodeID) in the range is greater than or equal to MaxNodeID, where MaxNodeID is exclusive for the request
		if an.BeginNodeID >= request.MaxNodeID {
			continue
		}
		branchesToQuery = append(branchesToQuery, fmt.Sprintf("'%v'", an.BranchID))
	}

	// If we haven't got enough branch ranges, then also query the current branch
	var lastEndNodeID int64
	if len(allAncestors) == 0 {
		lastEndNodeID = 1
	} else {
		lastEndNodeID = allAncestors[len(allAncestors)-1].EndNodeID
	}
	if lastEndNodeID < request.MaxNodeID {
		branchesToQuery = append(branchesToQuery, fmt.Sprintf("'%v'", branchID))
	}
	branchesInStr := strings.Join(branchesToQuery, ",")

	query := h.session.Query(v2templateReadNodes,
		treeID, branchesInStr, rowTypeHistoryNode, request.MinNodeID, request.MaxNodeID)

	iter := query.PageSize(numOfExpectedEvents).Iter()
	if iter == nil {
		return nil, &workflow.InternalServiceError{
			Message: "ReadHistoryBranch operation failed.  Not able to create query iterator.",
		}
	}

	eventBlob := &p.DataBlob{}
	history := make([]*p.DataBlob, 0)

	for iter.Scan(&eventBlob.ID, &eventBlob.Data, &eventBlob.Encoding) {
		history = append(history, eventBlob)
		eventBlob = &p.DataBlob{}
	}

	if err := iter.Close(); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ReadHistoryBranch. Close operation failed. Error: %v", err),
		}
	}
	if len(history) != numOfExpectedEvents {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ReadHistoryBranch. Expected %v events, but only got %v.", numOfExpectedEvents, len(history)),
		}
	}

	sort.Slice(history, func(i, j int) bool { return history[i].ID < history[j].ID })

	if history[0].ID != request.MinNodeID || history[len(history)-1].ID != request.MaxNodeID-1 {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ReadHistoryBranch. History events corrupted. Expected first/last eventIDs: %v/%v, but got %v/%v", request.MinNodeID, request.MaxNodeID-1, history[0].ID, history[len(history)-1].ID),
		}
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
	ans := make([]p.HistoryBranchRange, 0)
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
// If we fork B4 from B3 at node 25, the new branch is like:
//           B1[1~10]
//           /
//         B2 [11~20]
//        /
//      B3[21~25]
//     /
//   B4[...]
// However, if we fork from node 15, then the new branch should be
//           B1[1~10]
//           /
//         B2 [11~15]
//        /
//      B4[...]
// So even though it fork from B3, its ancestor doesn't have to have B3 in its branch ranges
func (h *cassandraHistoryPersistence) ForkHistoryBranch(request *p.ForkHistoryBranchRequest) (*p.ForkHistoryBranchResponse, error) {
	forkB := request.BranchInfo
	treeID := forkB.TreeID
	newAncestors := make([]p.HistoryBranchRange, 0)

	txnID, err := h.prepareTreeTransaction(treeID)
	if err != nil {
		return nil, err
	}
	err = h.validateBranchStatus(treeID, forkB.BranchID)
	if err != nil {
		return nil, err
	}

	// read Ancestors if needed
	if forkB.Ancestors == nil {
		forkB.Ancestors, err = h.getBranchAncestors(treeID, forkB.BranchID)
		if err != nil {
			return nil, err
		}
	}

	//add the new forking from branch
	lastEndNodeID := int64(1)
	if len(forkB.Ancestors) > 0 {
		lastEndNodeID = forkB.Ancestors[len(forkB.Ancestors)-1].EndNodeID
	}
	if lastEndNodeID > request.ForkFromNodeID {
		// this is the case that new branch's ancestor doesn't have forking branch
		for _, br := range forkB.Ancestors {
			newAncestors = append(newAncestors, br)
			if br.EndNodeID > request.ForkFromNodeID {
				break
			}
		}
	} else {
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
			BeginNodeID: lastEndNodeID,
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

	ancs := []map[string]interface{}{}
	for _, an := range newAncestors {
		value := make(map[string]interface{})
		value["forked_node_id"] = an.EndNodeID - 1
		value["branch_id"] = an.BranchID
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
		return nil, h.getInsertHistoryBranchFailure(previous, iter, txnID, treeID, request.NewBranchID)
	}

	return resp, nil
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

	// this may do a read transaction or a special write transaction(delete the root node)
	return h.deleteBranchAndRootNode(request.BranchInfo)
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
	treeID := branch.TreeID
	var err error
	if branch.Ancestors == nil {
		branch.Ancestors, err = h.getBranchAncestors(treeID, branch.BranchID)
		if err != nil {
			return err
		}
	}
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
			batch.Query(v2templateRangeDeleteNodes,
				treeID, br.BranchID, rowTypeHistoryNode, maxEndNodeID)
			break
		} else {
			// No any branch is using this range, we can delete all of it
			batch.Query(v2templateRangeDeleteNodes,
				treeID, br.BranchID, rowTypeHistoryNode, br.BeginNodeID)
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

func (h *cassandraHistoryPersistence) deleteBranchAndRootNode(branch p.HistoryBranch) error {
	treeID := branch.TreeID
	txnID, err := h.prepareTreeTransaction(treeID)
	if err != nil {
		return err
	}
	batch := h.session.NewBatch(gocql.LoggedBatch)

	batch.Query(v2templateDeleteOneNode,
		treeID, branch.BranchID, rowTypeHistoryBranch, branchNodeID)

	rsp, err := h.GetHistoryTree(&p.GetHistoryTreeRequest{
		TreeID: treeID,
	})
	if err != nil {
		return err
	}
	if len(rsp.Branches) == 0 {
		batch.Query(v2templateDeleteRoot,
			treeID, rootNodeBranchID, rowTypeHistoryNode, rootNodeID, txnID)
	} else {
		h.beginReadTransaction(batch, treeID, txnID)
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
		rowType, ok := previous["row_type"].(int)
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

		allPrevious = append(allPrevious, previous)
		previous = make(map[string]interface{})
		if !iter.MapScan(previous) {
			break GetFailureReasonLoop
		}
	}

	if txnIDUnmatch {
		return &p.ConditionFailedError{
			Msg: fmt.Sprintf("Failed to create a new branch. txnIDUnmatch: %v . Request txn_id: %v, Actual Value: %v",
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
	return &p.UnexpectedConditionFailedError{
		Msg: fmt.Sprintf("Failed to create a new branch. Request txn_id: %v, Actual Value: %v, columns: (%v)",
			reqTxnID, actualTxnID, columns),
	}
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
	ancs := []map[string]interface{}{}
	deleted := false
	branches := make([]p.HistoryBranch, 0)

	for iter.Scan(&branchUUID, &ancs, &deleted) {
		if deleted {
			continue
		}
		b := p.HistoryBranch{
			TreeID:    treeID,
			BranchID:  branchUUID.String(),
			Ancestors: h.parseBranchAncestors(ancs),
		}
		branches = append(branches, b)

		branchUUID = gocql.UUID{}
		ancs = []map[string]interface{}{}
		deleted = false
	}

	if err := iter.Close(); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("ReadHistoryBranch. Close operation failed. Error: %v", err),
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
