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

package errorinjectors

import (
	"context"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
)

type taskClient struct {
	persistence persistence.TaskManager
	errorRate   float64
	logger      log.Logger
}

// NewTaskClient creates an error injection client to manage tasks
func NewTaskClient(
	persistence persistence.TaskManager,
	errorRate float64,
	logger log.Logger,
) persistence.TaskManager {
	return &taskClient{
		persistence: persistence,
		errorRate:   errorRate,
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
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.CreateTasksResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.CreateTasks(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationCreateTasks,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *taskClient) GetTasks(
	ctx context.Context,
	request *persistence.GetTasksRequest,
) (*persistence.GetTasksResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetTasksResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetTasks(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetTasks,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *taskClient) CompleteTask(
	ctx context.Context,
	request *persistence.CompleteTaskRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.CompleteTask(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationCompleteTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *taskClient) CompleteTasksLessThan(
	ctx context.Context,
	request *persistence.CompleteTasksLessThanRequest,
) (*persistence.CompleteTasksLessThanResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.CompleteTasksLessThanResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.CompleteTasksLessThan(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationCompleteTasksLessThan,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *taskClient) GetOrphanTasks(
	ctx context.Context,
	request *persistence.GetOrphanTasksRequest,
) (*persistence.GetOrphanTasksResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetOrphanTasksResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetOrphanTasks(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationCompleteTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *taskClient) LeaseTaskList(
	ctx context.Context,
	request *persistence.LeaseTaskListRequest,
) (*persistence.LeaseTaskListResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.LeaseTaskListResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.LeaseTaskList(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationLeaseTaskList,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *taskClient) UpdateTaskList(
	ctx context.Context,
	request *persistence.UpdateTaskListRequest,
) (*persistence.UpdateTaskListResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.UpdateTaskListResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.UpdateTaskList(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationUpdateTaskList,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *taskClient) ListTaskList(
	ctx context.Context,
	request *persistence.ListTaskListRequest,
) (*persistence.ListTaskListResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ListTaskListResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListTaskList(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListTaskList,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *taskClient) DeleteTaskList(
	ctx context.Context,
	request *persistence.DeleteTaskListRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.DeleteTaskList(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationDeleteTaskList,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *taskClient) GetTaskListSize(
	ctx context.Context,
	request *persistence.GetTaskListSizeRequest,
) (*persistence.GetTaskListSizeResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var resp *persistence.GetTaskListSizeResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		resp, persistenceErr = p.persistence.GetTaskListSize(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetTaskListSize,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return resp, persistenceErr
}

func (p *taskClient) Close() {
	p.persistence.Close()
}
