package cli

import (
	"context"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/quotas"
)

const maxDBRetries    = 10

var (
	persistenceOperationRetryPolicy = common.CreatePersistanceRetryPolicy()
)

func retryGetWorkflowExecution(
	limiter *quotas.DynamicRateLimiter,
	totalDBRequests *int64,
	execStore persistence.ExecutionStore,
	req *persistence.GetWorkflowExecutionRequest,
) (*persistence.InternalGetWorkflowExecutionResponse, error) {
	var resp *persistence.InternalGetWorkflowExecutionResponse
	op := func() error {
		var err error
		preconditionForDBCall(totalDBRequests, limiter)
		resp, err = execStore.GetWorkflowExecution(req)
		return err
	}

	err := backoff.Retry(op, persistenceOperationRetryPolicy, common.IsPersistenceTransientError)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func retryGetCurrentExecution(
	limiter *quotas.DynamicRateLimiter,
	totalDBRequests *int64,
	execStore persistence.ExecutionStore,
	req *persistence.GetCurrentExecutionRequest,
) (*persistence.GetCurrentExecutionResponse, error) {
	var resp *persistence.GetCurrentExecutionResponse
	op := func() error {
		var err error
		preconditionForDBCall(totalDBRequests, limiter)
		resp, err = execStore.GetCurrentExecution(req)
		return err
	}

	err := backoff.Retry(op, persistenceOperationRetryPolicy, common.IsPersistenceTransientError)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func retryReadHistoryBranch(
	limiter *quotas.DynamicRateLimiter,
	totalDBRequests *int64,
	historyStore persistence.HistoryStore,
	req *persistence.InternalReadHistoryBranchRequest,
) (*persistence.InternalReadHistoryBranchResponse, error) {
	var resp *persistence.InternalReadHistoryBranchResponse
	op := func() error {
		var err error
		preconditionForDBCall(totalDBRequests, limiter)
		resp, err = historyStore.ReadHistoryBranch(req)
		return err
	}

	err := backoff.Retry(op, persistenceOperationRetryPolicy, common.IsPersistenceTransientError)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func retryDeleteWorkflowExecution(
	limiter *quotas.DynamicRateLimiter,
	totalDBRequests *int64,
	execStore persistence.ExecutionStore,
	req *persistence.DeleteWorkflowExecutionRequest,
) error {
	op := func() error {
		preconditionForDBCall(totalDBRequests, limiter)
		return execStore.DeleteWorkflowExecution(req)
	}

	err := backoff.Retry(op, persistenceOperationRetryPolicy, common.IsPersistenceTransientError)
	if err != nil {
		return err
	}
	return nil
}

func retryDeleteCurrentWorkflowExecution(
	limiter *quotas.DynamicRateLimiter,
	totalDBRequests *int64,
	execStore persistence.ExecutionStore,
	req *persistence.DeleteCurrentWorkflowExecutionRequest,
) error {
	op := func() error {
		preconditionForDBCall(totalDBRequests, limiter)
		return execStore.DeleteCurrentWorkflowExecution(req)
	}

	err := backoff.Retry(op, persistenceOperationRetryPolicy, common.IsPersistenceTransientError)
	if err != nil {
		return err
	}
	return nil
}

func preconditionForDBCall(totalDBRequests *int64, limiter *quotas.DynamicRateLimiter) {
	*totalDBRequests = *totalDBRequests + 1
	limiter.Wait(context.Background())
}