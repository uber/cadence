// The MIT License (MIT)
//
// Copyright (c) 2020 Uber Technologies, Inc.
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

package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/quotas"
)

const maxDBRetries = 10

var (
	persistenceOperationRetryPolicy = common.CreatePersistanceRetryPolicy()
)

func retryListConcreteExecutions(
	limiter *quotas.DynamicRateLimiter,
	totalDBRequests *int64,
	execStore persistence.ExecutionStore,
	req *persistence.ListConcreteExecutionsRequest,
	requestNum int,
) (*persistence.InternalListConcreteExecutionsResponse, error) {
	fmt.Printf("called retryListConcreteExecutions: %v, %+v\n", requestNum, *req)
	var resp *persistence.InternalListConcreteExecutionsResponse
	op := func() error {
		var err error
		preconditionForDBCall(totalDBRequests, limiter)
		resp, err = execStore.ListConcreteExecutions(req)
		return err
	}

	var err error
	for i := 0; i < maxDBRetries; i++ {
		<-time.After(time.Second)
		fmt.Printf("retryListConcreteExecutions top of loop: %v, %v\n", requestNum, i)
		err = backoff.Retry(op, persistenceOperationRetryPolicy, common.IsPersistenceTransientError)
		if err != nil {
			fmt.Printf("retryListConcreteExecutions got error: %v, %v, %v\n", requestNum, i, err)
			if !common.IsPersistenceTransientError(err) {
				fmt.Printf("retryListConcreteExecutions got non-tranisent error: %v, %v, %v\n", requestNum, i, err)
				return nil, err
			}
			continue
		}

		fmt.Printf("retryListConcreteExecutions got no error: %v, %v\n", requestNum, i)
		if len(resp.Executions) == 0 {
			fmt.Printf("retryListConcreteExecutions got empty executions: %v, %v, %v\n", requestNum, i, len(resp.NextPageToken))
			continue
		}
		fmt.Printf("retryListConcreteExecutions got non-empty executions: %v, %v, %v\n", requestNum, i, len(resp.Executions))
		return resp, nil
	}
	fmt.Printf("retryListConcreteExecutions failed all retriers: %v, %v\n", requestNum, err)
	return nil, err
}

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
