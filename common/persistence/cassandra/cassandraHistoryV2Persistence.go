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
	// below are templates for history_node table
	v2templateUpsertData = `INSERT INTO history_node (` +
		`tree_id, branch_id, node_id, txn_id, data, data_encoding) ` +
		`VALUES (?, ?, ?, ?, ?, ?) `

	v2templateReadData = `SELECT node_id, txn_id, data, data_encoding FROM history_node ` +
		`WHERE tree_id = ? AND branch_id = ? AND node_id >= ? AND node_id < ? `

	v2templateCheckNodeExists = `SELECT node_id FROM history_node ` +
		`WHERE tree_id = ? AND branch_id = ? AND node_id = ? LIMIT 1`

	v2templateRangeDeleteData = `DELETE FROM history_node ` +
		`WHERE tree_id = ? AND branch_id = ? AND node_id >= ? `

	// below are templates for history_tree table
	v2templateInsertTree = `INSERT INTO history_tree (` +
		`tree_id, branch_id, row_type, ancestors, deleted, txn_id) ` +
		`VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) IF NOT EXISTS `

	v2templateReadTwoBranches = `SELECT branch_id FROM history_tree WHERE tree_id = ? AND row_type = ? LIMIT 2 `

	v2templateReadAllBranches = `SELECT branch_id, ancestors, deleted FROM history_tree WHERE tree_id = ? AND row_type = ? `

	v2templateReadTreeRoot = `SELECT txn_id FROM history_tree WHERE tree_id = ? AND branch_id = ? AND row_type = ? `

	v2templateUpdateTreeRoot = `UPDATE history_tree SET txn_id = ? ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? IF txn_id = ? `

	v2templateValidateBranchStatus = `UPDATE history_tree SET deleted = false ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? if deleted = false `

	v2templateMarkBranchDeleted = `UPDATE history_tree SET deleted = true ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? `

	v2templateDeleteBranch = `DELETE FROM history_tree ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? `

	v2templateDeleteRoot = `DELETE FROM history_tree ` +
		`WHERE tree_id = ? AND branch_id = ? AND row_type = ? ` +
		`IF txn_id = ? `
)

const (
	// fake branchID for the tree root
	rootNodeFakeBranchID = "10000000-0000-f000-f000-000000000000"
	// the initial txn_id of for root node
	initialTransactionID = 0
	// a fake txn_id for branch rows
	fakeTransactionID = -1

	// Row types for history_tree table
	rowTypeHistoryBranch = 0
	rowTypeHistoryRoot   = 1

	// assume we don't have branches more than this number, otherwise CLI won't get all branches with GetHistoryTree() API
	maxBranchesReturnForOneTree = 10000
)

type (
	cassandraHistoryV2Persistence struct {
		cassandraStore
	}
)

// newHistoryPersistence is used to create an instance of HistoryManager implementation
func newHistoryV2Persistence(cfg config.Cassandra, logger bark.Logger) (p.HistoryV2Store,
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

	return &cassandraHistoryV2Persistence{cassandraStore: cassandraStore{session: session, logger: logger}}, nil
}

// We have two types of transactions for eventsV2 APIs: read/write trasaction.
// Like RWLock, read will not conflict with any other read, but write will conflict with other read/write:
// 		When Read/Write happens concurrently, write will succeed but read will fail if write committed before read.
// 		When Write/Write happens concurrently, only one of them will succeed.
// WriteTransaction is used by: NewBranch/ForkBranch/MarkBranchAsDeleted
// ReadTransaction is used by: AppendNodes/DeleteDataNode
// ReadHistoryBranch doesn't do any transaction

// write transaction will increase the txn_id of tree(root node) by one
func (h *cassandraHistoryV2Persistence) beginWriteTransaction(treeID string) (int64, *gocql.Batch, error) {
	currentTxnID, err := h.prepareTreeTransaction(treeID)
	if err != nil {
		return 0, nil, err
	}

	batch := h.session.NewBatch(gocql.LoggedBatch)
	batch.Query(v2templateUpdateTreeRoot,
		currentTxnID+1, treeID, rootNodeFakeBranchID, rowTypeHistoryRoot, currentTxnID)
	return currentTxnID, batch, nil
}

// read transaction will not change the txn_id of tree(root node)
func (h *cassandraHistoryV2Persistence) beginReadTransaction(treeID string) (int64, *gocql.Batch, error) {
	currentTxnID, err := h.prepareTreeTransaction(treeID)
	if err != nil {
		return 0, nil, err
	}

	batch := h.session.NewBatch(gocql.LoggedBatch)
	batch.Query(v2templateUpdateTreeRoot,
		currentTxnID, treeID, rootNodeFakeBranchID, rowTypeHistoryRoot, currentTxnID)
	return currentTxnID, batch, nil
}

// prepareTreeTransaction is an operation required for any read/write transaction
func (h *cassandraHistoryV2Persistence) prepareTreeTransaction(treeID string) (int64, error) {
	query := h.session.Query(v2templateReadTreeRoot,
		treeID,
		rootNodeFakeBranchID,
		rowTypeHistoryRoot)

	result := make(map[string]interface{})
	if err := query.MapScan(result); err != nil {
		return 0, convertCommonErrors("prepareTreeTransaction", err)
	}

	txnID := result["txn_id"].(int64)
	return txnID, nil
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
func (h *cassandraHistoryV2Persistence) NewHistoryBranch(request *p.InternalNewHistoryBranchRequest) (*p.InternalNewHistoryBranchResponse, error) {
	treeID := request.TreeID
	branchID := request.BranchID
	isNewTree, err := h.createRoot(treeID)
	if err != nil {
		return nil, err
	}
	resp := &p.InternalNewHistoryBranchResponse{
		IsNewTree:  isNewTree,
		BranchInfo: nil,
	}
	txnID, batch, err := h.beginWriteTransaction(treeID)
	if err != nil {
		return resp, err
	}

	batch.Query(v2templateInsertTree,
		treeID, branchID, rowTypeHistoryBranch, nil, false, fakeTransactionID)

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

	resp.BranchInfo = workflow.HistoryBranch{
		TreeID:    &treeID,
		BranchID:  &branchID,
		Ancestors: []*workflow.HistoryBranchRange{},
	}
	return resp, nil
}

func (h *cassandraHistoryV2Persistence) getNewHistoryBranchFailure(previous map[string]interface{}, iter *gocql.Iter, reqTxnID int64, reqTreeID, reqBranchID string) error {
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

		if rowType == rowTypeHistoryRoot && treeID == reqTreeID && branchID == rootNodeFakeBranchID {
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

func (h *cassandraHistoryV2Persistence) createRoot(treeID string) (bool, error) {
	var query *gocql.Query

	query = h.session.Query(v2templateInsertTree,
		treeID,
		rootNodeFakeBranchID,
		rowTypeHistoryRoot,
		nil, // ancestors
		false,
		initialTransactionID)

	previous := make(map[string]interface{})
	applied, err := query.MapScanCAS(previous)
	if err != nil {
		return false, convertCommonErrors("createRoot", err)
	}

	return applied, nil
}

// return error is branch is already deleted/not exists
func (h *cassandraHistoryV2Persistence) validateBranchStatus(batch *gocql.Batch, treeID string, branchID string) {
	batch.Query(v2templateValidateBranchStatus,
		treeID,
		branchID,
		rowTypeHistoryBranch)
}

// AppendHistoryNodes upsert a batch of events as a single node to a history branch
// Note that it's not allowed to append above the branch's ancestors' nodes, which means nodeID > forking point
func (h *cassandraHistoryV2Persistence) AppendHistoryNodes(request *p.InternalAppendHistoryNodesRequest) error {
	branchInfo := request.BranchInfo
	forkingNodeID := h.getForkingNode(branchInfo)

	if request.NodeID <= forkingNodeID {
		return &p.InvalidPersistenceRequestError{
			Msg: fmt.Sprintf("cannot append to ancestors' nodes"),
		}
	}

	query := h.session.Query(v2templateUpsertData,
		branchInfo.TreeID, branchInfo.BranchID, request.NodeID, request.TransactionID, request.Events.Data, request.Events.Encoding)
	return query.Exec()
}

func (h *cassandraHistoryV2Persistence) getForkingNode(bi workflow.HistoryBranch) int64 {
	if len(bi.Ancestors) == 0 {
		// root branch
		return 0
	} else {
		idx := len(bi.Ancestors) - 1
		return *bi.Ancestors[idx].EndNodeID - 1
	}
}

func (h *cassandraHistoryV2Persistence) readBranchRange(treeID, branchID string, minID, maxID int64) ([]*p.DataBlob, error) {
	bs := make([]*p.DataBlob, 0, int(maxID-minID))

	query := h.session.Query(v2templateReadData,
		treeID, branchID, minID, maxID)

	iter := query.PageSize(int(maxID - minID)).Iter()
	if iter == nil {
		return nil, &workflow.InternalServiceError{
			Message: "readBranchRange operation failed.  Not able to create query iterator.",
		}
	}

	lastNodeID := int64(-1)
	eventBlob := &p.DataBlob{}
	nodeID := int64(0)

	for iter.Scan(&nodeID, nil, &eventBlob.Data, &eventBlob.Encoding) {
		if nodeID == lastNodeID {
			// skip the nodes with smaller txn_id
			continue
		}
		lastNodeID = nodeID
		bs = append(bs, eventBlob)
		eventBlob = &p.DataBlob{}
	}

	if err := iter.Close(); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("readBranchRange. Close operation failed. Error: %v", err),
		}
	}
	return bs, nil
}

// ReadHistoryBranch returns history node data for a branch
// NOTE: For branch that has ancestors, we need to query Cassandra multiple times, because it doesn't support OR/UNION operator
func (h *cassandraHistoryV2Persistence) ReadHistoryBranch(request *p.InternalReadHistoryBranchRequest) (*p.InternalReadHistoryBranchResponse, error) {
	treeID := *request.BranchInfo.TreeID
	branchID := request.BranchInfo.BranchID

	allBRs := request.BranchInfo.Ancestors
	// We may also query the current branch from forkingNodeID
	forkingNodeID := h.getForkingNode(request.BranchInfo)
	allBRs = append(allBRs, &workflow.HistoryBranchRange{
		BranchID:    branchID,
		BeginNodeID: common.Int64Ptr(forkingNodeID + 1),
		EndNodeID:   common.Int64Ptr(request.MaxNodeID),
	})

	history := make([]*p.DataBlob, 0, request.MaxNodeID-request.MinNodeID)
	for _, br := range allBRs {
		// this range won't contain any nodes needed, since the last node(EndNodeID-1) in the range is strictly less than MinNodeID
		if *br.EndNodeID <= request.MinNodeID {
			continue
		}
		// similarly, this range won't contain any nodes needed, since the first node(BeginNodeID) in the range is greater than or equal to MaxNodeID, where MaxNodeID is exclusive for the request
		if *br.BeginNodeID >= request.MaxNodeID {
			// since they are sorted, the rest branch range can be skipped
			break
		}

		minID := *br.BeginNodeID
		if request.MinNodeID > minID {
			minID = request.MinNodeID
		}
		maxID := *br.EndNodeID
		if request.MaxNodeID < maxID {
			maxID = request.MaxNodeID
		}

		var err error
		history, err = h.readBranchRange(treeID, *br.BranchID, minID, maxID)
		if err != nil {
			return nil, err
		}
		// NOTE: we will break here for simplicity. Since we do batching, the pagination is already broken anyway for PageSize
		break
	}

	response := &p.InternalReadHistoryBranchResponse{
		History: history,
	}

	return response, nil
}

func (h *cassandraHistoryV2Persistence) checkIfNodeExists(treeID, branchID string, nodeID int64) (bool, error) {
	query := h.session.Query(v2templateCheckNodeExists,
		treeID,
		branchID,
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
func (h *cassandraHistoryV2Persistence) ForkHistoryBranch(request *p.InternalForkHistoryBranchRequest) (*p.InternalForkHistoryBranchResponse, error) {
	forkB := request.ForkBranchInfo
	treeID := *forkB.TreeID
	newAncestors := make([]*workflow.HistoryBranchRange, 0, len(forkB.Ancestors)+1)

	lastForkingNodeID := h.getForkingNode(forkB)
	if lastForkingNodeID >= request.ForkFromNodeID {
		// this is the case that new branch's ancestors doesn't include the forking branch
		for _, br := range forkB.Ancestors {
			if *br.EndNodeID > request.ForkFromNodeID {
				newAncestors = append(newAncestors, &workflow.HistoryBranchRange{
					BranchID:    br.BranchID,
					BeginNodeID: br.BeginNodeID,
					EndNodeID:   common.Int64Ptr(request.ForkFromNodeID + 1),
				})
				break
			} else {
				newAncestors = append(newAncestors, br)
			}
		}
	} else {
		// this is the case the new branch will inherit all ancestors from forking branch

		// first we need to check if the branch has a valid nodeID to fork from
		exits, err := h.checkIfNodeExists(treeID, *forkB.BranchID, request.ForkFromNodeID)
		if err != nil {
			return nil, err
		}
		if !exits {
			return nil, &p.InvalidPersistenceRequestError{
				Msg: fmt.Sprintf("forking branch %v doesn't have node %v", *forkB.BranchID, request.ForkFromNodeID),
			}
		}

		newAncestors = forkB.Ancestors
		newAncestors = append(newAncestors, &workflow.HistoryBranchRange{
			BranchID:    forkB.BranchID,
			BeginNodeID: common.Int64Ptr(lastForkingNodeID + 1),
			EndNodeID:   common.Int64Ptr(request.ForkFromNodeID + 1),
		})
	}

	resp := &p.InternalForkHistoryBranchResponse{
		NewBranchInfo: workflow.HistoryBranch{
			TreeID:    &treeID,
			BranchID:  &request.NewBranchID,
			Ancestors: newAncestors,
		}}

	txnID, batch, err := h.beginWriteTransaction(treeID)
	if err != nil {
		return nil, err
	}

	h.validateBranchStatus(batch, treeID, *forkB.BranchID)

	ancs := []map[string]interface{}{}
	for _, an := range newAncestors {
		value := make(map[string]interface{})
		value["forked_node_id"] = *an.EndNodeID - 1
		value["branch_id"] = an.BranchID
		ancs = append(ancs, value)
	}

	batch.Query(v2templateInsertTree,
		treeID, request.NewBranchID, rowTypeHistoryBranch, ancs, false, fakeTransactionID)

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
		return nil, h.getForkHistoryBranchFailure(previous, iter, txnID, treeID, request.NewBranchID, *forkB.BranchID)
	}

	return resp, nil
}

func (h *cassandraHistoryV2Persistence) getForkHistoryBranchFailure(previous map[string]interface{}, iter *gocql.Iter, reqTxnID int64, reqTreeID, newBranchID, forkBranchID string) error {
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
		deleted := previous["deleted"].(bool)

		if rowType == rowTypeHistoryRoot && treeID == reqTreeID && branchID == rootNodeFakeBranchID {
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
func (h *cassandraHistoryV2Persistence) DeleteHistoryBranch(request *p.InternalDeleteHistoryBranchRequest) error {
	// We break the operation into three phases:
	// 1. Mark the branch as deleted
	// 2. Delete data nodes
	// 3. Delete the branch node, also delete root node if the branch is the last branch

	// this would do a write transaction
	err := h.markBranchAsDeleted(request.BranchInfo)
	if err != nil {
		return err
	}

	// this would not do transaction, there can be race condition if we are appending events to the branch, we might leak the events
	err = h.deleteDataNodes(request.BranchInfo)
	if err != nil {
		return err
	}

	// this may do a write transaction or a special write transaction(delete the root node)
	err = h.deleteBranchAndRootNode(request.BranchInfo)
	return err
}

func (h *cassandraHistoryV2Persistence) markBranchAsDeleted(branch workflow.HistoryBranch) error {
	batch := h.session.NewBatch(gocql.LoggedBatch)
	txnID, batch, err := h.beginWriteTransaction(*branch.TreeID)
	if err != nil {
		return err
	}

	treeID := *branch.TreeID
	branchID := *branch.BranchID
	batch.Query(v2templateMarkBranchDeleted,
		treeID, branchID, rowTypeHistoryBranch)

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
		return h.getTreeTrasactionFailure(previous, iter, txnID, treeID, branchID)
	}
	return nil
}

func (h *cassandraHistoryV2Persistence) deleteDataNodes(branch workflow.HistoryBranch) error {
	treeID := *branch.TreeID
	brsToDelete := branch.Ancestors
	forkingNodeID := h.getForkingNode(branch)
	brsToDelete = append(brsToDelete, &workflow.HistoryBranchRange{
		BranchID:    branch.BranchID,
		BeginNodeID: common.Int64Ptr(forkingNodeID + 1),
	})

	rsp, err := h.GetHistoryTree(&p.GetHistoryTreeRequest{
		TreeID: treeID,
	})
	if err != nil {
		return err
	}

	// validBRsMaxEndNode is to know each branch range that is being used, we want to know what is the max nodeID referred by other valid branch
	validBRsMaxEndNode := map[string]int64{}
	for _, b := range rsp.Branches {
		for _, br := range b.Ancestors {
			curr, ok := validBRsMaxEndNode[*br.BranchID]
			if !ok || curr < *br.EndNodeID {
				validBRsMaxEndNode[*br.BranchID] = *br.EndNodeID
			}
		}
	}

	// for each branch range to delete, we iterate from bottom to up, and delete up to the point according to validBRsEndNode
	for i := len(brsToDelete) - 1; i >= 0; i-- {
		br := brsToDelete[i]
		maxReferredEndNodeID, ok := validBRsMaxEndNode[*br.BranchID]
		if ok {
			// we can only delete from the maxEndNode and stop here
			err := h.deleteBranchRangeNodes(treeID, *br.BranchID, maxReferredEndNodeID)
			if err != nil {
				return err
			}
			break
		} else {
			// No any branch is using this range, we can delete all of it
			err := h.deleteBranchRangeNodes(treeID, *br.BranchID, *br.BeginNodeID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *cassandraHistoryV2Persistence) deleteBranchRangeNodes(treeID, branchID string, beginNodeID int64) error {
	query := h.session.Query(v2templateRangeDeleteData,
		treeID,
		branchID,
		beginNodeID)

	err := query.Exec()
	if err != nil {
		return convertCommonErrors("deleteBranchRangeNodes", err)
	}
	return nil
}

func (h *cassandraHistoryV2Persistence) deleteBranchAndRootNode(branch workflow.HistoryBranch) error {
	treeID := *branch.TreeID
	branchID := *branch.BranchID
	txnID, err := h.prepareTreeTransaction(treeID)
	if err != nil {
		return err
	}
	batch := h.session.NewBatch(gocql.LoggedBatch)

	batch.Query(v2templateDeleteBranch,
		treeID, branchID, rowTypeHistoryBranch)

	isLast, err := h.isLastBranch(treeID)
	if err != nil {
		return err
	}
	if isLast {
		batch.Query(v2templateDeleteRoot,
			treeID, rootNodeFakeBranchID, rowTypeHistoryRoot, txnID)
	} else {
		batch.Query(v2templateUpdateTreeRoot,
			txnID+1, treeID, rootNodeFakeBranchID, rowTypeHistoryRoot, txnID)
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
		return h.getTreeTrasactionFailure(previous, iter, txnID, treeID, branchID)
	}
	return nil
}

func (h *cassandraHistoryV2Persistence) getTreeTrasactionFailure(previous map[string]interface{}, iter *gocql.Iter, reqTxnID int64, reqTreeID, reqBranchID string) error {
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

		if rowType == rowTypeHistoryRoot && treeID == reqTreeID && branchID == rootNodeFakeBranchID {
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

func (h *cassandraHistoryV2Persistence) isLastBranch(treeID string) (bool, error) {
	query := h.session.Query(v2templateReadTwoBranches, treeID, rowTypeHistoryBranch)

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
func (h *cassandraHistoryV2Persistence) GetHistoryTree(request *p.GetHistoryTreeRequest) (*p.GetHistoryTreeResponse, error) {
	treeID := request.TreeID

	query := h.session.Query(v2templateReadAllBranches, treeID, rowTypeHistoryBranch)

	iter := query.PageSize(maxBranchesReturnForOneTree).Iter()
	if iter == nil {
		return nil, &workflow.InternalServiceError{
			Message: "GetHistoryTree operation failed.  Not able to create query iterator.",
		}
	}

	branchUUID := gocql.UUID{}
	ancsResult := []map[string]interface{}{}
	deleted := false
	branches := make([]workflow.HistoryBranch, 0)

	for iter.Scan(&branchUUID, &ancsResult, &deleted) {
		if deleted {
			continue
		}
		ancs := h.parseBranchAncestors(ancsResult)
		b := workflow.HistoryBranch{
			TreeID:    &treeID,
			BranchID:  common.StringPtr(branchUUID.String()),
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

func (h *cassandraHistoryV2Persistence) parseBranchAncestors(ancestors []map[string]interface{}) []*workflow.HistoryBranchRange {
	ans := make([]*workflow.HistoryBranchRange, 0, len(ancestors))
	for _, e := range ancestors {
		an := &workflow.HistoryBranchRange{}
		for k, v := range e {
			switch k {
			case "branch_id":
				an.BranchID = common.StringPtr(v.(gocql.UUID).String())
			case "forked_node_id":
				an.EndNodeID = common.Int64Ptr(v.(int64) + 1)
			}
		}
		ans = append(ans, an)
	}

	if len(ans) > 0 {
		// sort ans based onf EndNodeID so that we can set BeginNodeID
		sort.Slice(ans, func(i, j int) bool { return *ans[i].EndNodeID < *ans[j].EndNodeID })
		ans[0].BeginNodeID = common.Int64Ptr(int64(1))
		for i := 1; i < len(ans); i++ {
			ans[i].BeginNodeID = ans[i-1].EndNodeID
		}
	}
	return ans
}
