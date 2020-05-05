package common

import (
	"github.com/uber/cadence/common/pagination"
	"github.com/uber/cadence/common/persistence"
)

type (
	persistenceIterator struct {
		itr pagination.Iterator
	}
)

// NewPersistenceIterator returns a new paginated iterator over persistence
func NewPersistenceIterator(
	pr PersistenceRetryer,
	pageSize int,
	shardID int,
) ExecutionIterator {
	return &persistenceIterator{
		itr: pagination.NewIterator(nil, getPersistenceFetchPageFn(pr, pageSize, shardID)),
	}
}

// Next returns the next execution
func (i *persistenceIterator) Next() (*Execution, error) {
	exec, err := i.itr.Next()
	return exec.(*Execution), err
}

// HasNext returns true if there is another execution, false otherwise.
func (i *persistenceIterator) HasNext() bool {
	return i.itr.HasNext()
}

func getPersistenceFetchPageFn(
	pr PersistenceRetryer,
	pageSize int,
	shardID int,
) pagination.FetchFn {
	return func(token pagination.PageToken) ([]pagination.Entity, pagination.PageToken, error) {
		req := &persistence.ListConcreteExecutionsRequest{
			PageSize:  pageSize,
			PageToken: token.([]byte),
		}
		resp, err := pr.ListConcreteExecutions(req)
		if err != nil {
			return nil, nil, err
		}
		executions := make([]pagination.Entity, len(resp.Executions), len(resp.Executions))
		for i, e := range resp.Executions {
			branchToken := e.ExecutionInfo.BranchToken
			if e.VersionHistories != nil {
				branchToken = e.VersionHistories.Histories[e.VersionHistories.CurrentVersionHistoryIndex].BranchToken
			}
			executions[i] = &Execution{
				ShardID:     shardID,
				DomainID:    e.ExecutionInfo.DomainID,
				WorkflowID:  e.ExecutionInfo.WorkflowID,
				RunID:       e.ExecutionInfo.RunID,
				BranchToken: branchToken,
				State:       e.ExecutionInfo.State,
			}
		}
		return executions, resp.PageToken, nil
	}
}
