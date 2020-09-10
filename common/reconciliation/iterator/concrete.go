package iterator

import (
	"github.com/uber/cadence/common/codec"
	"github.com/uber/cadence/common/pagination"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/common"
)

func ConcreteExecution(retryer persistence.Retryer, pageSize int) pagination.Iterator {
	return pagination.NewIterator(nil, GetConcreteExecutions(retryer, pageSize, codec.NewThriftRWEncoder()))
}

func GetConcreteExecutions(
	pr persistence.Retryer,
	pageSize int,
	encoder *codec.ThriftRWEncoder,
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
