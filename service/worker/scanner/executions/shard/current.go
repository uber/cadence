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

	"github.com/uber/cadence/common/pagination"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/common"
	"github.com/uber/cadence/common/reconciliation/invariants"
)

// NewCurrentExecutionScanner returns a scanner which scans current executions
func NewCurrentExecutionScanner(
	params ScannerParams,
) common.Scanner {
	id := uuid.New()

	var ivs []common.Invariant
	for _, fn := range CurrentExecutionType.ToInvariants(params.InvariantCollections) {
		ivs = append(ivs, fn(params.Retryer))
	}

	return &scanner{
		shardID:          params.Retryer.GetShardID(),
		itr:              pagination.NewIterator(nil, getCurrentExecutionsPersistenceFetchPageFn(params.Retryer, params.PersistencePageSize)),
		failedWriter:     common.NewBlobstoreWriter(id, common.FailedExtension, params.BlobstoreClient, params.BlobstoreFlushThreshold),
		corruptedWriter:  common.NewBlobstoreWriter(id, common.CorruptedExtension, params.BlobstoreClient, params.BlobstoreFlushThreshold),
		invariantManager: invariants.NewInvariantManager(ivs),
		progressReportFn: params.ProgressReportFn,
	}
}

func getCurrentExecutionsPersistenceFetchPageFn(
	pr common.PersistenceRetryer,
	pageSize int,
) pagination.FetchFn {
	return func(token pagination.PageToken) (pagination.Page, error) {
		req := &persistence.ListCurrentExecutionsRequest{
			PageSize: pageSize,
		}
		if token != nil {
			req.PageToken = token.([]byte)
		}
		resp, err := pr.ListCurrentExecutions(req)
		if err != nil {
			return pagination.Page{}, err
		}
		executions := make([]pagination.Entity, len(resp.Executions), len(resp.Executions))
		for i, e := range resp.Executions {
			currentExec := &common.CurrentExecution{
				CurrentRunID: e.CurrentRunID,
				Execution: common.Execution{
					ShardID:    pr.GetShardID(),
					DomainID:   e.DomainID,
					WorkflowID: e.WorkflowID,
					RunID:      e.RunID,
					State:      e.State,
				},
			}
			if err := currentExec.Validate(); err != nil {
				return pagination.Page{}, err
			}
			executions[i] = currentExec
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
