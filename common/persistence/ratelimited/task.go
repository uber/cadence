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

type taskClient struct {
	rateLimiter quotas.Limiter
	persistence persistence.TaskManager
	logger      log.Logger
}

// NewTaskClient creates a client to manage tasks
func NewTaskClient(
	persistence persistence.TaskManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) persistence.TaskManager {
	return &taskClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

func (p *taskClient) GetName() string {
	return p.persistence.GetName()
}

func (p *taskClient) CreateTasks(
	ctx context.Context,
	request *persistence.CreateTasksRequest,
) (*persistence.CreateTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.CreateTasks(ctx, request)
	return response, err
}

func (p *taskClient) GetTasks(
	ctx context.Context,
	request *persistence.GetTasksRequest,
) (*persistence.GetTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetTasks(ctx, request)
	return response, err
}

func (p *taskClient) CompleteTask(
	ctx context.Context,
	request *persistence.CompleteTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CompleteTask(ctx, request)
	return err
}

func (p *taskClient) CompleteTasksLessThan(
	ctx context.Context,
	request *persistence.CompleteTasksLessThanRequest,
) (*persistence.CompleteTasksLessThanResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.CompleteTasksLessThan(ctx, request)
}

func (p *taskClient) GetOrphanTasks(ctx context.Context, request *persistence.GetOrphanTasksRequest) (*persistence.GetOrphanTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.GetOrphanTasks(ctx, request)
}

func (p *taskClient) LeaseTaskList(
	ctx context.Context,
	request *persistence.LeaseTaskListRequest,
) (*persistence.LeaseTaskListResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.LeaseTaskList(ctx, request)
	return response, err
}

func (p *taskClient) UpdateTaskList(
	ctx context.Context,
	request *persistence.UpdateTaskListRequest,
) (*persistence.UpdateTaskListResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.UpdateTaskList(ctx, request)
	return response, err
}

func (p *taskClient) ListTaskList(
	ctx context.Context,
	request *persistence.ListTaskListRequest,
) (*persistence.ListTaskListResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.ListTaskList(ctx, request)
}

func (p *taskClient) DeleteTaskList(
	ctx context.Context,
	request *persistence.DeleteTaskListRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}
	return p.persistence.DeleteTaskList(ctx, request)
}

func (p *taskClient) GetTaskListSize(
	ctx context.Context,
	request *persistence.GetTaskListSizeRequest,
) (*persistence.GetTaskListSizeResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.GetTaskListSize(ctx, request)
}

func (p *taskClient) Close() {
	p.persistence.Close()
}
