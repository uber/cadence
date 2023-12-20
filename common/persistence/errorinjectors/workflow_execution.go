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

type workflowExecutionClient struct {
	persistence persistence.ExecutionManager
	errorRate   float64
	logger      log.Logger
}

// NewWorkflowExecutionClient creates an error injection client to manage executions
func NewWorkflowExecutionClient(
	persistence persistence.ExecutionManager,
	errorRate float64,
	logger log.Logger,
) persistence.ExecutionManager {
	return &workflowExecutionClient{
		persistence: persistence,
		errorRate:   errorRate,
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
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.CreateWorkflowExecutionResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.CreateWorkflowExecution(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationCreateWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) GetWorkflowExecution(
	ctx context.Context,
	request *persistence.GetWorkflowExecutionRequest,
) (*persistence.GetWorkflowExecutionResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetWorkflowExecutionResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetWorkflowExecution(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) UpdateWorkflowExecution(
	ctx context.Context,
	request *persistence.UpdateWorkflowExecutionRequest,
) (*persistence.UpdateWorkflowExecutionResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.UpdateWorkflowExecutionResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.UpdateWorkflowExecution(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationUpdateWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) ConflictResolveWorkflowExecution(
	ctx context.Context,
	request *persistence.ConflictResolveWorkflowExecutionRequest,
) (*persistence.ConflictResolveWorkflowExecutionResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ConflictResolveWorkflowExecutionResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ConflictResolveWorkflowExecution(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationConflictResolveWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) DeleteWorkflowExecution(
	ctx context.Context,
	request *persistence.DeleteWorkflowExecutionRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.DeleteWorkflowExecution(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationDeleteWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *workflowExecutionClient) DeleteCurrentWorkflowExecution(
	ctx context.Context,
	request *persistence.DeleteCurrentWorkflowExecutionRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.DeleteCurrentWorkflowExecution(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationDeleteCurrentWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *workflowExecutionClient) GetCurrentExecution(
	ctx context.Context,
	request *persistence.GetCurrentExecutionRequest,
) (*persistence.GetCurrentExecutionResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetCurrentExecutionResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetCurrentExecution(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetCurrentExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) ListCurrentExecutions(
	ctx context.Context,
	request *persistence.ListCurrentExecutionsRequest,
) (*persistence.ListCurrentExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ListCurrentExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListCurrentExecutions(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListCurrentExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) IsWorkflowExecutionExists(
	ctx context.Context,
	request *persistence.IsWorkflowExecutionExistsRequest,
) (*persistence.IsWorkflowExecutionExistsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.IsWorkflowExecutionExistsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.IsWorkflowExecutionExists(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationIsWorkflowExecutionExists,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) ListConcreteExecutions(
	ctx context.Context,
	request *persistence.ListConcreteExecutionsRequest,
) (*persistence.ListConcreteExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ListConcreteExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListConcreteExecutions(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListConcreteExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) GetTransferTasks(
	ctx context.Context,
	request *persistence.GetTransferTasksRequest,
) (*persistence.GetTransferTasksResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetTransferTasksResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetTransferTasks(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetTransferTasks,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) GetCrossClusterTasks(
	ctx context.Context,
	request *persistence.GetCrossClusterTasksRequest,
) (*persistence.GetCrossClusterTasksResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetCrossClusterTasksResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetCrossClusterTasks(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetTransferTasks,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) GetReplicationTasks(
	ctx context.Context,
	request *persistence.GetReplicationTasksRequest,
) (*persistence.GetReplicationTasksResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetReplicationTasksResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetReplicationTasks(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetReplicationTasks,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) CompleteTransferTask(
	ctx context.Context,
	request *persistence.CompleteTransferTaskRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.CompleteTransferTask(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationCompleteTransferTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *workflowExecutionClient) RangeCompleteTransferTask(
	ctx context.Context,
	request *persistence.RangeCompleteTransferTaskRequest,
) (*persistence.RangeCompleteTransferTaskResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.RangeCompleteTransferTaskResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.RangeCompleteTransferTask(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationRangeCompleteTransferTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) CompleteCrossClusterTask(
	ctx context.Context,
	request *persistence.CompleteCrossClusterTaskRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.CompleteCrossClusterTask(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationCompleteCrossClusterTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *workflowExecutionClient) RangeCompleteCrossClusterTask(
	ctx context.Context,
	request *persistence.RangeCompleteCrossClusterTaskRequest,
) (*persistence.RangeCompleteCrossClusterTaskResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.RangeCompleteCrossClusterTaskResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.RangeCompleteCrossClusterTask(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationRangeCompleteCrossClusterTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) CompleteReplicationTask(
	ctx context.Context,
	request *persistence.CompleteReplicationTaskRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.CompleteReplicationTask(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationCompleteReplicationTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *workflowExecutionClient) RangeCompleteReplicationTask(
	ctx context.Context,
	request *persistence.RangeCompleteReplicationTaskRequest,
) (*persistence.RangeCompleteReplicationTaskResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.RangeCompleteReplicationTaskResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.RangeCompleteReplicationTask(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationRangeCompleteReplicationTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) PutReplicationTaskToDLQ(
	ctx context.Context,
	request *persistence.PutReplicationTaskToDLQRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.PutReplicationTaskToDLQ(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationPutReplicationTaskToDLQ,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *workflowExecutionClient) GetReplicationTasksFromDLQ(
	ctx context.Context,
	request *persistence.GetReplicationTasksFromDLQRequest,
) (*persistence.GetReplicationTasksFromDLQResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetReplicationTasksFromDLQResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetReplicationTasksFromDLQ(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetReplicationTasksFromDLQ,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) GetReplicationDLQSize(
	ctx context.Context,
	request *persistence.GetReplicationDLQSizeRequest,
) (*persistence.GetReplicationDLQSizeResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetReplicationDLQSizeResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetReplicationDLQSize(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetReplicationDLQSize,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) DeleteReplicationTaskFromDLQ(
	ctx context.Context,
	request *persistence.DeleteReplicationTaskFromDLQRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.DeleteReplicationTaskFromDLQ(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationDeleteReplicationTaskFromDLQ,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *workflowExecutionClient) RangeDeleteReplicationTaskFromDLQ(
	ctx context.Context,
	request *persistence.RangeDeleteReplicationTaskFromDLQRequest,
) (*persistence.RangeDeleteReplicationTaskFromDLQResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.RangeDeleteReplicationTaskFromDLQResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.RangeDeleteReplicationTaskFromDLQ(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationRangeDeleteReplicationTaskFromDLQ,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) CreateFailoverMarkerTasks(
	ctx context.Context,
	request *persistence.CreateFailoverMarkersRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.CreateFailoverMarkerTasks(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationCreateFailoverMarkerTasks,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *workflowExecutionClient) GetTimerIndexTasks(
	ctx context.Context,
	request *persistence.GetTimerIndexTasksRequest,
) (*persistence.GetTimerIndexTasksResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetTimerIndexTasksResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetTimerIndexTasks(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetTimerIndexTasks,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) CompleteTimerTask(
	ctx context.Context,
	request *persistence.CompleteTimerTaskRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.CompleteTimerTask(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationCompleteTimerTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *workflowExecutionClient) RangeCompleteTimerTask(
	ctx context.Context,
	request *persistence.RangeCompleteTimerTaskRequest,
) (*persistence.RangeCompleteTimerTaskResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.RangeCompleteTimerTaskResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.RangeCompleteTimerTask(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationRangeCompleteTimerTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *workflowExecutionClient) Close() {
	p.persistence.Close()
}
