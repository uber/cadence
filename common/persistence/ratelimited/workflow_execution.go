// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

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

package ratelimited

import (
	"context"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/quotas"
)

type workflowExecutionClient struct {
	rateLimiter quotas.Limiter
	persistence persistence.ExecutionManager
	logger      log.Logger
}

// NewWorkflowExecutionClient creates a client to manage executions
func NewWorkflowExecutionClient(
	persistence persistence.ExecutionManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) persistence.ExecutionManager {
	return &workflowExecutionClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

func (p *workflowExecutionClient) GetName() string {
	return p.persistence.GetName()
}

func (p *workflowExecutionClient) GetShardID() int {
	return p.persistence.GetShardID()
}

func (p *workflowExecutionClient) CreateWorkflowExecution(
	ctx context.Context,
	request *persistence.CreateWorkflowExecutionRequest,
) (*persistence.CreateWorkflowExecutionResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.CreateWorkflowExecution(ctx, request)
	return response, err
}

func (p *workflowExecutionClient) GetWorkflowExecution(
	ctx context.Context,
	request *persistence.GetWorkflowExecutionRequest,
) (*persistence.GetWorkflowExecutionResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetWorkflowExecution(ctx, request)
	return response, err
}

func (p *workflowExecutionClient) UpdateWorkflowExecution(
	ctx context.Context,
	request *persistence.UpdateWorkflowExecutionRequest,
) (*persistence.UpdateWorkflowExecutionResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	resp, err := p.persistence.UpdateWorkflowExecution(ctx, request)
	return resp, err
}

func (p *workflowExecutionClient) ConflictResolveWorkflowExecution(
	ctx context.Context,
	request *persistence.ConflictResolveWorkflowExecutionRequest,
) (*persistence.ConflictResolveWorkflowExecutionResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	resp, err := p.persistence.ConflictResolveWorkflowExecution(ctx, request)
	return resp, err
}

func (p *workflowExecutionClient) DeleteWorkflowExecution(
	ctx context.Context,
	request *persistence.DeleteWorkflowExecutionRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.DeleteWorkflowExecution(ctx, request)
	return err
}

func (p *workflowExecutionClient) DeleteCurrentWorkflowExecution(
	ctx context.Context,
	request *persistence.DeleteCurrentWorkflowExecutionRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.DeleteCurrentWorkflowExecution(ctx, request)
	return err
}

func (p *workflowExecutionClient) GetCurrentExecution(
	ctx context.Context,
	request *persistence.GetCurrentExecutionRequest,
) (*persistence.GetCurrentExecutionResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetCurrentExecution(ctx, request)
	return response, err
}

func (p *workflowExecutionClient) ListCurrentExecutions(
	ctx context.Context,
	request *persistence.ListCurrentExecutionsRequest,
) (*persistence.ListCurrentExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListCurrentExecutions(ctx, request)
	return response, err
}

func (p *workflowExecutionClient) IsWorkflowExecutionExists(
	ctx context.Context,
	request *persistence.IsWorkflowExecutionExistsRequest,
) (*persistence.IsWorkflowExecutionExistsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.IsWorkflowExecutionExists(ctx, request)
	return response, err
}

func (p *workflowExecutionClient) ListConcreteExecutions(
	ctx context.Context,
	request *persistence.ListConcreteExecutionsRequest,
) (*persistence.ListConcreteExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListConcreteExecutions(ctx, request)
	return response, err
}

func (p *workflowExecutionClient) GetTransferTasks(
	ctx context.Context,
	request *persistence.GetTransferTasksRequest,
) (*persistence.GetTransferTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetTransferTasks(ctx, request)
	return response, err
}

func (p *workflowExecutionClient) GetCrossClusterTasks(
	ctx context.Context,
	request *persistence.GetCrossClusterTasksRequest,
) (*persistence.GetCrossClusterTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetCrossClusterTasks(ctx, request)
	return response, err
}

func (p *workflowExecutionClient) GetReplicationTasks(
	ctx context.Context,
	request *persistence.GetReplicationTasksRequest,
) (*persistence.GetReplicationTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetReplicationTasks(ctx, request)
	return response, err
}

func (p *workflowExecutionClient) CompleteTransferTask(
	ctx context.Context,
	request *persistence.CompleteTransferTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CompleteTransferTask(ctx, request)
	return err
}

func (p *workflowExecutionClient) RangeCompleteTransferTask(
	ctx context.Context,
	request *persistence.RangeCompleteTransferTaskRequest,
) (*persistence.RangeCompleteTransferTaskResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.RangeCompleteTransferTask(ctx, request)
}

func (p *workflowExecutionClient) CompleteCrossClusterTask(
	ctx context.Context,
	request *persistence.CompleteCrossClusterTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CompleteCrossClusterTask(ctx, request)
	return err
}

func (p *workflowExecutionClient) RangeCompleteCrossClusterTask(
	ctx context.Context,
	request *persistence.RangeCompleteCrossClusterTaskRequest,
) (*persistence.RangeCompleteCrossClusterTaskResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.RangeCompleteCrossClusterTask(ctx, request)
}

func (p *workflowExecutionClient) CompleteReplicationTask(
	ctx context.Context,
	request *persistence.CompleteReplicationTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CompleteReplicationTask(ctx, request)
	return err
}

func (p *workflowExecutionClient) RangeCompleteReplicationTask(
	ctx context.Context,
	request *persistence.RangeCompleteReplicationTaskRequest,
) (*persistence.RangeCompleteReplicationTaskResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.RangeCompleteReplicationTask(ctx, request)
}

func (p *workflowExecutionClient) PutReplicationTaskToDLQ(
	ctx context.Context,
	request *persistence.PutReplicationTaskToDLQRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.PutReplicationTaskToDLQ(ctx, request)
}

func (p *workflowExecutionClient) GetReplicationTasksFromDLQ(
	ctx context.Context,
	request *persistence.GetReplicationTasksFromDLQRequest,
) (*persistence.GetReplicationTasksFromDLQResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.GetReplicationTasksFromDLQ(ctx, request)
}

func (p *workflowExecutionClient) GetReplicationDLQSize(
	ctx context.Context,
	request *persistence.GetReplicationDLQSizeRequest,
) (*persistence.GetReplicationDLQSizeResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.GetReplicationDLQSize(ctx, request)
}

func (p *workflowExecutionClient) DeleteReplicationTaskFromDLQ(
	ctx context.Context,
	request *persistence.DeleteReplicationTaskFromDLQRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.DeleteReplicationTaskFromDLQ(ctx, request)
}

func (p *workflowExecutionClient) RangeDeleteReplicationTaskFromDLQ(
	ctx context.Context,
	request *persistence.RangeDeleteReplicationTaskFromDLQRequest,
) (*persistence.RangeDeleteReplicationTaskFromDLQResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.RangeDeleteReplicationTaskFromDLQ(ctx, request)
}

func (p *workflowExecutionClient) CreateFailoverMarkerTasks(
	ctx context.Context,
	request *persistence.CreateFailoverMarkersRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CreateFailoverMarkerTasks(ctx, request)
	return err
}

func (p *workflowExecutionClient) GetTimerIndexTasks(
	ctx context.Context,
	request *persistence.GetTimerIndexTasksRequest,
) (*persistence.GetTimerIndexTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetTimerIndexTasks(ctx, request)
	return response, err
}

func (p *workflowExecutionClient) CompleteTimerTask(
	ctx context.Context,
	request *persistence.CompleteTimerTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CompleteTimerTask(ctx, request)
	return err
}

func (p *workflowExecutionClient) RangeCompleteTimerTask(
	ctx context.Context,
	request *persistence.RangeCompleteTimerTaskRequest,
) (*persistence.RangeCompleteTimerTaskResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.RangeCompleteTimerTask(ctx, request)
}

func (p *workflowExecutionClient) Close() {
	p.persistence.Close()
}
