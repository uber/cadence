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
	"context"
	"fmt"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
	"github.com/uber/cadence/common/persistence/utils"
	"github.com/uber/cadence/common/types"
)

type (
	nosqlHistoryManager struct {
		db     nosqlplugin.DB
		logger log.Logger
	}
)

// NewHistoryV2PersistenceFromSession returns new HistoryStore
func NewHistoryV2PersistenceFromSession(
	client gocql.Client,
	session gocql.Session,
	logger log.Logger,
) p.HistoryStore {
	// TODO hardcoding to Cassandra for now, will switch to dynamically loading later
	db := cassandra.NewCassandraDBFromSession(client, session, logger)

	return &nosqlHistoryManager{db: db, logger: logger}
}

// newHistoryPersistence is used to create an instance of HistoryManager implementation
func newHistoryV2Persistence(
	cfg config.Cassandra,
	logger log.Logger,
) (p.HistoryStore, error) {

	// TODO hardcoding to Cassandra for now, will switch to dynamically loading later
	db, err := cassandra.NewCassandraDB(cfg, logger)
	if err != nil {
		return nil, err
	}

	return &nosqlHistoryManager{db: db, logger: logger}, nil
}

func (h *nosqlHistoryManager) GetName() string {
	return h.db.PluginName()
}

// Close releases the underlying resources held by this object
func (h *nosqlHistoryManager) Close() {
	h.db.Close()
}

// AppendHistoryNodes upsert a batch of events as a single node to a history branch
// Note that it's not allowed to append above the branch's ancestors' nodes, which means nodeID >= ForkNodeID
func (h *nosqlHistoryManager) AppendHistoryNodes(
	ctx context.Context,
	request *p.InternalAppendHistoryNodesRequest,
) error {

	branchInfo := request.BranchInfo
	beginNodeID := utils.GetBeginNodeID(branchInfo)

	if request.NodeID < beginNodeID {
		return &p.InvalidPersistenceRequestError{
			Msg: fmt.Sprintf("cannot append to ancestors' nodes"),
		}
	}

	var err error
	var treeRow *nosqlplugin.HistoryTreeRow
	if request.IsNewBranch {
		var ancestors []*types.HistoryBranchRange
		for _, anc := range branchInfo.Ancestors {
			ancestors = append(ancestors, anc)
		}
		treeRow = &nosqlplugin.HistoryTreeRow{
			ShardID:         request.ShardID,
			TreeID:          branchInfo.GetTreeID(),
			BranchID:        branchInfo.GetBranchID(),
			Ancestors:       ancestors,
			CreateTimestamp: time.Now(),
			Info:            request.Info,
		}
	}
	nodeRow := &nosqlplugin.HistoryNodeRow{
		TreeID:       branchInfo.GetTreeID(),
		BranchID:     branchInfo.GetBranchID(),
		NodeID:       request.NodeID,
		TxnID:        &request.TransactionID,
		Data:         request.Events.Data,
		DataEncoding: string(request.Events.Encoding),
		ShardID:      request.ShardID,
	}
	err = h.db.InsertIntoHistoryTreeAndNode(ctx, treeRow, nodeRow)

	if err != nil {
		return convertCommonErrors(h.db, "AppendHistoryNodes", err)
	}
	return nil
}

// ReadHistoryBranch returns history node data for a branch
// NOTE: For branch that has ancestors, we need to query Cassandra multiple times, because it doesn't support OR/UNION operator
func (h *nosqlHistoryManager) ReadHistoryBranch(
	ctx context.Context,
	request *p.InternalReadHistoryBranchRequest,
) (*p.InternalReadHistoryBranchResponse, error) {
	filter := &nosqlplugin.HistoryNodeFilter{
		ShardID:       request.ShardID,
		TreeID:        request.TreeID,
		BranchID:      request.BranchID,
		MinNodeID:     request.MinNodeID,
		MaxNodeID:     request.MaxNodeID,
		NextPageToken: request.NextPageToken,
		PageSize:      request.PageSize,
	}
	rows, pagingToken, err := h.db.SelectFromHistoryNode(ctx, filter)
	if err != nil {
		return nil, convertCommonErrors(h.db, "SelectFromHistoryNode", err)
	}

	history := make([]*p.DataBlob, 0, int(request.PageSize))

	eventBlob := &p.DataBlob{}
	nodeID := int64(0)
	txnID := int64(0)
	lastNodeID := request.LastNodeID
	lastTxnID := request.LastTransactionID

	for _, row := range rows {
		nodeID = row.NodeID
		txnID = *row.TxnID
		eventBlob.Data = row.Data
		eventBlob.Encoding = common.EncodingType(row.DataEncoding)
		if txnID < lastTxnID {
			// assuming that business logic layer is correct and transaction ID only increase
			// thus, valid event batch will come with increasing transaction ID

			// event batches with smaller node ID
			//  -> should not be possible since records are already sorted
			// event batches with same node ID
			//  -> batch with higher transaction ID is valid
			// event batches with larger node ID
			//  -> batch with lower transaction ID is invalid (happens before)
			//  -> batch with higher transaction ID is valid
			continue
		}

		switch {
		case nodeID < lastNodeID:
			return nil, &types.InternalDataInconsistencyError{
				Message: fmt.Sprintf("corrupted data, nodeID cannot decrease"),
			}
		case nodeID == lastNodeID:
			return nil, &types.InternalDataInconsistencyError{
				Message: fmt.Sprintf("corrupted data, same nodeID must have smaller txnID"),
			}
		default: // row.NodeID > lastNodeID:
			// NOTE: when row.nodeID > lastNodeID, we expect the one with largest txnID comes first
			lastTxnID = txnID
			lastNodeID = nodeID
			history = append(history, eventBlob)
			eventBlob = &p.DataBlob{}
		}
	}

	return &p.InternalReadHistoryBranchResponse{
		History:           history,
		NextPageToken:     pagingToken,
		LastNodeID:        lastNodeID,
		LastTransactionID: lastTxnID,
	}, nil
}

// ForkHistoryBranch forks a new branch from an existing branch
// Note that application must provide a void forking nodeID, it must be a valid nodeID in that branch.
// A valid forking nodeID can be an ancestor from the existing branch.
// For example, we have branch B1 with three nodes(1[1,2], 3[3,4,5] and 6[6,7,8]. 1, 3 and 6 are nodeIDs (first eventID of the batch).
// So B1 looks like this:
//           1[1,2]
//           /
//         3[3,4,5]
//        /
//      6[6,7,8]
//
// Assuming we have branch B2 which contains one ancestor B1 stopping at 6 (exclusive). So B2 inherit nodeID 1 and 3 from B1, and have its own nodeID 6 and 8.
// Branch B2 looks like this:
//           1[1,2]
//           /
//         3[3,4,5]
//          \
//           6[6,7]
//           \
//            8[8]
//
//Now we want to fork a new branch B3 from B2.
// The only valid forking nodeIDs are 3,6 or 8.
// 1 is not valid because we can't fork from first node.
// 2/4/5 is NOT valid either because they are inside a batch.
//
// Case #1: If we fork from nodeID 6, then B3 will have an ancestor B1 which stops at 6(exclusive).
// As we append a batch of events[6,7,8,9] to B3, it will look like :
//           1[1,2]
//           /
//         3[3,4,5]
//          \
//         6[6,7,8,9]
//
// Case #2: If we fork from node 8, then B3 will have two ancestors: B1 stops at 6(exclusive) and ancestor B2 stops at 8(exclusive)
// As we append a batch of events[8,9] to B3, it will look like:
//           1[1,2]
//           /
//         3[3,4,5]
//        /
//      6[6,7]
//       \
//       8[8,9]
//
func (h *nosqlHistoryManager) ForkHistoryBranch(
	ctx context.Context,
	request *p.InternalForkHistoryBranchRequest,
) (*p.InternalForkHistoryBranchResponse, error) {

	forkB := request.ForkBranchInfo
	treeID := *forkB.TreeID
	newAncestors := make([]*types.HistoryBranchRange, 0, len(forkB.Ancestors)+1)

	beginNodeID := utils.GetBeginNodeID(forkB)
	if beginNodeID >= request.ForkNodeID {
		// this is the case that new branch's ancestors doesn't include the forking branch
		for _, br := range forkB.Ancestors {
			if *br.EndNodeID >= request.ForkNodeID {
				newAncestors = append(newAncestors, &types.HistoryBranchRange{
					BranchID:    br.BranchID,
					BeginNodeID: br.BeginNodeID,
					EndNodeID:   common.Int64Ptr(request.ForkNodeID),
				})
				break
			} else {
				newAncestors = append(newAncestors, br)
			}
		}
	} else {
		// this is the case the new branch will inherit all ancestors from forking branch
		newAncestors = forkB.Ancestors
		newAncestors = append(newAncestors, &types.HistoryBranchRange{
			BranchID:    forkB.BranchID,
			BeginNodeID: common.Int64Ptr(beginNodeID),
			EndNodeID:   common.Int64Ptr(request.ForkNodeID),
		})
	}

	resp := &p.InternalForkHistoryBranchResponse{
		NewBranchInfo: types.HistoryBranch{
			TreeID:    &treeID,
			BranchID:  &request.NewBranchID,
			Ancestors: newAncestors,
		}}

	var ancestors []*types.HistoryBranchRange
	for _, an := range newAncestors {
		anc := &types.HistoryBranchRange{
			BranchID:  an.BranchID,
			EndNodeID: an.EndNodeID,
		}
		ancestors = append(ancestors, anc)
	}
	treeRow := &nosqlplugin.HistoryTreeRow{
		ShardID:         request.ShardID,
		TreeID:          treeID,
		BranchID:        request.NewBranchID,
		Ancestors:       ancestors,
		CreateTimestamp: time.Now(),
		Info:            request.Info,
	}

	err := h.db.InsertIntoHistoryTreeAndNode(ctx, treeRow, nil)
	if err != nil {
		return nil, convertCommonErrors(h.db, "ForkHistoryBranch", err)
	}
	return resp, nil
}

// DeleteHistoryBranch removes a branch
func (h *nosqlHistoryManager) DeleteHistoryBranch(
	ctx context.Context,
	request *p.InternalDeleteHistoryBranchRequest,
) error {

	branch := request.BranchInfo
	treeID := *branch.TreeID
	brsToDelete := branch.Ancestors
	beginNodeID := utils.GetBeginNodeID(branch)
	brsToDelete = append(brsToDelete, &types.HistoryBranchRange{
		BranchID:    branch.BranchID,
		BeginNodeID: common.Int64Ptr(beginNodeID),
	})

	rsp, err := h.GetHistoryTree(ctx, &p.InternalGetHistoryTreeRequest{
		TreeID:  treeID,
		ShardID: &request.ShardID,
	})
	if err != nil {
		return err
	}

	treeFilter := &nosqlplugin.HistoryTreeFilter{
		ShardID:  request.ShardID,
		TreeID:   treeID,
		BranchID: branch.BranchID,
	}
	var nodeFilters []*nosqlplugin.HistoryNodeFilter

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
			nodeFilter := &nosqlplugin.HistoryNodeFilter{
				ShardID:   request.ShardID,
				TreeID:    treeID,
				BranchID:  *br.BranchID,
				MinNodeID: maxReferredEndNodeID,
			}
			nodeFilters = append(nodeFilters, nodeFilter)
			break
		} else {
			// No any branch is using this range, we can delete all of it
			nodeFilter := &nosqlplugin.HistoryNodeFilter{
				ShardID:   request.ShardID,
				TreeID:    treeID,
				BranchID:  *br.BranchID,
				MinNodeID: *br.BeginNodeID,
			}
			nodeFilters = append(nodeFilters, nodeFilter)
		}
	}

	err = h.db.DeleteFromHistoryTreeAndNode(ctx, treeFilter, nodeFilters)
	if err != nil {
		return convertCommonErrors(h.db, "DeleteHistoryBranch", err)
	}
	return nil
}

func (h *nosqlHistoryManager) GetAllHistoryTreeBranches(
	ctx context.Context,
	request *p.GetAllHistoryTreeBranchesRequest,
) (*p.GetAllHistoryTreeBranchesResponse, error) {
	dbBranches, pagingToken, err := h.db.SelectAllHistoryTrees(ctx, request.NextPageToken, request.PageSize)
	if err != nil {
		return nil, convertCommonErrors(h.db, "SelectAllHistoryTrees", err)
	}

	branchDetails := make([]p.HistoryBranchDetail, 0, int(request.PageSize))

	for _, branch := range dbBranches {

		branchDetail := p.HistoryBranchDetail{
			TreeID:   branch.TreeID,
			BranchID: branch.BranchID,
			ForkTime: branch.CreateTimestamp,
			Info:     branch.Info,
		}
		branchDetails = append(branchDetails, branchDetail)
	}

	response := &p.GetAllHistoryTreeBranchesResponse{
		Branches:      branchDetails,
		NextPageToken: pagingToken,
	}

	return response, nil
}

// GetHistoryTree returns all branch information of a tree
func (h *nosqlHistoryManager) GetHistoryTree(
	ctx context.Context,
	request *p.InternalGetHistoryTreeRequest,
) (*p.InternalGetHistoryTreeResponse, error) {

	treeID := request.TreeID

	dbBranches, err := h.db.SelectFromHistoryTree(ctx,
		&nosqlplugin.HistoryTreeFilter{
			ShardID: *request.ShardID,
			TreeID:  treeID,
		})
	if err != nil {
		return nil, convertCommonErrors(h.db, "SelectFromHistoryTree", err)
	}

	branches := make([]*types.HistoryBranch, 0)
	for _, dbBr := range dbBranches {
		br := &types.HistoryBranch{
			TreeID:    &treeID,
			BranchID:  &dbBr.BranchID,
			Ancestors: dbBr.Ancestors,
		}
		branches = append(branches, br)
	}
	return &p.InternalGetHistoryTreeResponse{
		Branches: branches,
	}, nil
}
