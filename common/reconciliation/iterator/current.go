package iterator

import (
	"github.com/uber/cadence/common/pagination"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/common"
)

func CurrentExecution(retryer persistence.Retryer, pageSize int) pagination.Iterator {
	return pagination.NewIterator(nil, getCurrentExecution(retryer, pageSize))
}

func getCurrentExecution(
	pr persistence.Retryer,
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
