// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package shard

import (
	"github.com/pborman/uuid"

	"github.com/uber/cadence/common/codec"
	"github.com/uber/cadence/common/pagination"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/common"
	"github.com/uber/cadence/common/reconciliation/invariants"
)

// NewConcreteExecutionScanner returns a scanner which scans concrete executions
func NewConcreteExecutionScanner(
	params ScannerParams,
) common.Scanner {
	id := uuid.New()
	pr := params.Retryer

	var ivs []common.Invariant
	for _, fn := range ConcreteExecutionType.ToInvariants(params.InvariantCollections) {
		ivs = append(ivs, fn(pr))
	}

	return &scanner{
		shardID:          pr.GetShardID(),
		itr:              pagination.NewIterator(nil, getConcreteExecutionsPersistenceFetchPageFn(pr, codec.NewThriftRWEncoder(), params.PersistencePageSize)),
		failedWriter:     common.NewBlobstoreWriter(id, common.FailedExtension, params.BlobstoreClient, params.BlobstoreFlushThreshold),
		corruptedWriter:  common.NewBlobstoreWriter(id, common.CorruptedExtension, params.BlobstoreClient, params.BlobstoreFlushThreshold),
		invariantManager: invariants.NewInvariantManager(ivs),
		progressReportFn: params.ProgressReportFn,
	}
}

func getConcreteExecutionsPersistenceFetchPageFn(
	pr common.PersistenceRetryer,
	encoder *codec.ThriftRWEncoder,
	pageSize int,
) pagination.FetchFn {
	return func(token pagination.PageToken) (pagination.Page, error) {
		req := &persistence.ListConcreteExecutionsRequest{
			PageSize: pageSize,
		}
		if token != nil {
			req.PageToken = token.([]byte)
		}
		resp, err := pr.ListConcreteExecutions(req)
		if err != nil {
			return pagination.Page{}, err
		}
		executions := make([]pagination.Entity, len(resp.Executions), len(resp.Executions))
		for i, e := range resp.Executions {
			branchToken, treeID, branchID, err := common.GetBranchToken(e, encoder)
			if err != nil {
				return pagination.Page{}, err
			}
			concreteExec := &common.ConcreteExecution{
				BranchToken: branchToken,
				TreeID:      treeID,
				BranchID:    branchID,
				Execution: common.Execution{
					ShardID:    pr.GetShardID(),
					DomainID:   e.ExecutionInfo.DomainID,
					WorkflowID: e.ExecutionInfo.WorkflowID,
					RunID:      e.ExecutionInfo.RunID,
					State:      e.ExecutionInfo.State,
				},
			}
			if err := concreteExec.Validate(); err != nil {
				return pagination.Page{}, err
			}
			executions[i] = concreteExec
		}
		var nextToken interface{} = resp.PageToken
		if len(resp.PageToken) == 0 {
			nextToken = nil
		}
		page := pagination.Page{
			CurrentToken: token,
			NextToken:    nextToken,
			Entities:     executions,
		}
		return page, nil
	}
}
