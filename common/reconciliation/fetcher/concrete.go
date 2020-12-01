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

package fetcher

import (
	"context"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/codec"
	"github.com/uber/cadence/common/pagination"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/entity"
	"github.com/uber/cadence/common/types"
)

// ConcreteExecutionIterator is used to retrieve Concrete executions.
func ConcreteExecutionIterator(
	ctx context.Context,
	retryer persistence.Retryer,
	pageSize int,
) pagination.Iterator {
	return pagination.NewIterator(ctx, nil, getConcreteExecutions(retryer, pageSize, codec.NewThriftRWEncoder()))
}

// ConcreteExecution returns a single ConcreteExecution from persistence
func ConcreteExecution(
	ctx context.Context,
	retryer persistence.Retryer,
	request ExecutionRequest,
) (entity.Entity, error) {

	req := persistence.GetWorkflowExecutionRequest{
		DomainID: request.DomainID,
		Execution: types.WorkflowExecution{
			WorkflowID: common.StringPtr(request.WorkflowID),
			RunID:      common.StringPtr(request.RunID),
		},
	}
	e, err := retryer.GetWorkflowExecution(ctx, &req)
	if err != nil {
		return nil, err
	}

	branchToken, branch, err := getBranchToken(e.State.ExecutionInfo.BranchToken, e.State.VersionHistories, codec.NewThriftRWEncoder())

	return &entity.ConcreteExecution{
		BranchToken: branchToken,
		TreeID:      branch.GetTreeID(),
		BranchID:    branch.GetBranchID(),
		Execution: entity.Execution{
			ShardID:    retryer.GetShardID(),
			DomainID:   e.State.ExecutionInfo.DomainID,
			WorkflowID: e.State.ExecutionInfo.WorkflowID,
			RunID:      e.State.ExecutionInfo.RunID,
			State:      e.State.ExecutionInfo.State,
		},
	}, nil
}

func getConcreteExecutions(
	pr persistence.Retryer,
	pageSize int,
	encoder *codec.ThriftRWEncoder,
) pagination.FetchFn {
	return func(ctx context.Context, token pagination.PageToken) (pagination.Page, error) {
		req := &persistence.ListConcreteExecutionsRequest{
			PageSize: pageSize,
		}
		if token != nil {
			req.PageToken = token.([]byte)
		}
		resp, err := pr.ListConcreteExecutions(ctx, req)
		if err != nil {
			return pagination.Page{}, err
		}
		executions := make([]pagination.Entity, len(resp.Executions), len(resp.Executions))
		for i, e := range resp.Executions {
			branchToken, branch, err := getBranchToken(e.ExecutionInfo.BranchToken, e.VersionHistories, encoder)
			if err != nil {
				return pagination.Page{}, err
			}
			concreteExec := &entity.ConcreteExecution{
				BranchToken: branchToken,
				TreeID:      branch.GetTreeID(),
				BranchID:    branch.GetBranchID(),
				Execution: entity.Execution{
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

// getBranchToken returns the branchToken and historyBranch, error on failure.
func getBranchToken(
	branchToken []byte,
	histories *persistence.VersionHistories,
	decoder *codec.ThriftRWEncoder,
) ([]byte, shared.HistoryBranch, error) {
	var branch shared.HistoryBranch
	bt := branchToken
	if histories != nil {
		versionHistory, err := histories.GetCurrentVersionHistory()
		if err != nil {
			return nil, branch, err
		}
		bt = versionHistory.GetBranchToken()
	}

	if err := decoder.Decode(bt, &branch); err != nil {
		return nil, branch, err
	}

	return bt, branch, nil
}
