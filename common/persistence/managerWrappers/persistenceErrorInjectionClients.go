// Copyright (c) 2017-2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package managerWrappers

import (
	"context"
	"math/rand"

	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/visibility"

	"github.com/uber/cadence/common/errors"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
)

var (
	// ErrFakeTimeout is a fake persistence timeout error.
	ErrFakeTimeout = &persistence.TimeoutError{Msg: "Fake Persistence Timeout Error."}
)

var (
	fakeErrors = []error{
		errors.ErrFakeServiceBusy,
		errors.ErrFakeInternalService,
		ErrFakeTimeout,
		errors.ErrFakeUnhandled,
	}
)

const (
	msgInjectedFakeErr = "Injected fake persistence error"
)

type (
	shardErrorInjectionPersistenceClient struct {
		persistence persistence.ShardManager
		errorRate   float64
		logger      log.Logger
	}

	workflowExecutionErrorInjectionPersistenceClient struct {
		persistence persistence.ExecutionManager
		errorRate   float64
		logger      log.Logger
	}

	taskErrorInjectionPersistenceClient struct {
		persistence persistence.TaskManager
		errorRate   float64
		logger      log.Logger
	}

	historyErrorInjectionPersistenceClient struct {
		persistence persistence.HistoryManager
		errorRate   float64
		logger      log.Logger
	}

	metadataErrorInjectionPersistenceClient struct {
		persistence persistence.DomainManager
		errorRate   float64
		logger      log.Logger
	}

	visibilityErrorInjectionPersistenceClient struct {
		persistence visibility.VisibilityManager
		errorRate   float64
		logger      log.Logger
	}

	queueErrorInjectionPersistenceClient struct {
		persistence persistence.QueueManager
		errorRate   float64
		logger      log.Logger
	}
)

var _ persistence.ShardManager = (*shardErrorInjectionPersistenceClient)(nil)
var _ persistence.ExecutionManager = (*workflowExecutionErrorInjectionPersistenceClient)(nil)
var _ persistence.TaskManager = (*taskErrorInjectionPersistenceClient)(nil)
var _ persistence.HistoryManager = (*historyErrorInjectionPersistenceClient)(nil)
var _ persistence.DomainManager = (*metadataErrorInjectionPersistenceClient)(nil)
var _ visibility.VisibilityManager = (*visibilityErrorInjectionPersistenceClient)(nil)
var _ persistence.QueueManager = (*queueErrorInjectionPersistenceClient)(nil)

// NewShardPersistenceErrorInjectionClient creates an error injection client to manage shards
func NewShardPersistenceErrorInjectionClient(
	persistence persistence.ShardManager,
	errorRate float64,
	logger log.Logger,
) persistence.ShardManager {
	return &shardErrorInjectionPersistenceClient{
		persistence: persistence,
		errorRate:   errorRate,
		logger:      logger,
	}
}

// NewWorkflowExecutionPersistenceErrorInjectionClient creates an error injection client to manage executions
func NewWorkflowExecutionPersistenceErrorInjectionClient(
	persistence persistence.ExecutionManager,
	errorRate float64,
	logger log.Logger,
) persistence.ExecutionManager {
	return &workflowExecutionErrorInjectionPersistenceClient{
		persistence: persistence,
		errorRate:   errorRate,
		logger:      logger,
	}
}

// NewTaskPersistenceErrorInjectionClient creates an error injection client to manage tasks
func NewTaskPersistenceErrorInjectionClient(
	persistence persistence.TaskManager,
	errorRate float64,
	logger log.Logger,
) persistence.TaskManager {
	return &taskErrorInjectionPersistenceClient{
		persistence: persistence,
		errorRate:   errorRate,
		logger:      logger,
	}
}

// NewHistoryPersistenceErrorInjectionClient creates an error injection HistoryManager client to manage workflow execution history
func NewHistoryPersistenceErrorInjectionClient(
	persistence persistence.HistoryManager,
	errorRate float64,
	logger log.Logger,
) persistence.HistoryManager {
	return &historyErrorInjectionPersistenceClient{
		persistence: persistence,
		errorRate:   errorRate,
		logger:      logger,
	}
}

// NewMetadataPersistenceErrorInjectionClient creates an error injection DomainManager client to manage metadata
func NewMetadataPersistenceErrorInjectionClient(
	persistence persistence.DomainManager,
	errorRate float64,
	logger log.Logger,
) persistence.DomainManager {
	return &metadataErrorInjectionPersistenceClient{
		persistence: persistence,
		errorRate:   errorRate,
		logger:      logger,
	}
}

// NewVisibilityPersistenceErrorInjectionClient creates an error injection client to manage visibility
func NewVisibilityPersistenceErrorInjectionClient(
	persistence visibility.VisibilityManager,
	errorRate float64,
	logger log.Logger,
) visibility.VisibilityManager {
	return &visibilityErrorInjectionPersistenceClient{
		persistence: persistence,
		errorRate:   errorRate,
		logger:      logger,
	}
}

// NewQueuePersistenceErrorInjectionClient creates an error injection client to manage queue
func NewQueuePersistenceErrorInjectionClient(
	persistence persistence.QueueManager,
	errorRate float64,
	logger log.Logger,
) persistence.QueueManager {
	return &queueErrorInjectionPersistenceClient{
		persistence: persistence,
		errorRate:   errorRate,
		logger:      logger,
	}
}

func (p *shardErrorInjectionPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *shardErrorInjectionPersistenceClient) CreateShard(
	ctx context.Context,
	request *persistence.CreateShardRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.CreateShard(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr, tag.StoreOperationCreateShard, tag.Error(fakeErr), tag.Bool(forwardCall), tag.StoreError(persistenceErr))
		return fakeErr
	}
	return persistenceErr
}

func (p *shardErrorInjectionPersistenceClient) GetShard(
	ctx context.Context,
	request *persistence.GetShardRequest,
) (*persistence.GetShardResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetShardResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetShard(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr, tag.StoreOperationGetShard, tag.Error(fakeErr), tag.Bool(forwardCall), tag.StoreError(persistenceErr))
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *shardErrorInjectionPersistenceClient) UpdateShard(
	ctx context.Context,
	request *persistence.UpdateShardRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.UpdateShard(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr, tag.StoreOperationUpdateShard, tag.Error(fakeErr), tag.Bool(forwardCall), tag.StoreError(persistenceErr))
		return fakeErr
	}
	return persistenceErr
}

func (p *shardErrorInjectionPersistenceClient) Close() {
	p.persistence.Close()
}

func (p *workflowExecutionErrorInjectionPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *workflowExecutionErrorInjectionPersistenceClient) GetShardID() int {
	return p.persistence.GetShardID()
}

func (p *workflowExecutionErrorInjectionPersistenceClient) CreateWorkflowExecution(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) GetWorkflowExecution(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) UpdateWorkflowExecution(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) ConflictResolveWorkflowExecution(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) ResetWorkflowExecution(
	ctx context.Context,
	request *persistence.ResetWorkflowExecutionRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.ResetWorkflowExecution(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationResetWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *workflowExecutionErrorInjectionPersistenceClient) DeleteWorkflowExecution(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) DeleteCurrentWorkflowExecution(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) GetCurrentExecution(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) ListCurrentExecutions(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) IsWorkflowExecutionExists(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) ListConcreteExecutions(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) GetTransferTasks(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) GetCrossClusterTasks(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) GetReplicationTasks(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) CompleteTransferTask(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) RangeCompleteTransferTask(
	ctx context.Context,
	request *persistence.RangeCompleteTransferTaskRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.RangeCompleteTransferTask(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationRangeCompleteTransferTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *workflowExecutionErrorInjectionPersistenceClient) CompleteCrossClusterTask(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) RangeCompleteCrossClusterTask(
	ctx context.Context,
	request *persistence.RangeCompleteCrossClusterTaskRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.RangeCompleteCrossClusterTask(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationRangeCompleteCrossClusterTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *workflowExecutionErrorInjectionPersistenceClient) CompleteReplicationTask(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) RangeCompleteReplicationTask(
	ctx context.Context,
	request *persistence.RangeCompleteReplicationTaskRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.RangeCompleteReplicationTask(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationRangeCompleteReplicationTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *workflowExecutionErrorInjectionPersistenceClient) PutReplicationTaskToDLQ(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) GetReplicationTasksFromDLQ(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) GetReplicationDLQSize(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) DeleteReplicationTaskFromDLQ(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) RangeDeleteReplicationTaskFromDLQ(
	ctx context.Context,
	request *persistence.RangeDeleteReplicationTaskFromDLQRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.RangeDeleteReplicationTaskFromDLQ(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationRangeDeleteReplicationTaskFromDLQ,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *workflowExecutionErrorInjectionPersistenceClient) CreateFailoverMarkerTasks(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) GetTimerIndexTasks(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) CompleteTimerTask(
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

func (p *workflowExecutionErrorInjectionPersistenceClient) RangeCompleteTimerTask(
	ctx context.Context,
	request *persistence.RangeCompleteTimerTaskRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.RangeCompleteTimerTask(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationRangeCompleteTimerTask,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *workflowExecutionErrorInjectionPersistenceClient) Close() {
	p.persistence.Close()
}

func (p *taskErrorInjectionPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *taskErrorInjectionPersistenceClient) CreateTasks(
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

func (p *taskErrorInjectionPersistenceClient) GetTasks(
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

func (p *taskErrorInjectionPersistenceClient) CompleteTask(
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

func (p *taskErrorInjectionPersistenceClient) CompleteTasksLessThan(
	ctx context.Context,
	request *persistence.CompleteTasksLessThanRequest,
) (int, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response int
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
		return 0, fakeErr
	}
	return response, persistenceErr
}

func (p *taskErrorInjectionPersistenceClient) GetOrphanTasks(
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

func (p *taskErrorInjectionPersistenceClient) LeaseTaskList(
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

func (p *taskErrorInjectionPersistenceClient) UpdateTaskList(
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

func (p *taskErrorInjectionPersistenceClient) ListTaskList(
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

func (p *taskErrorInjectionPersistenceClient) DeleteTaskList(
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

func (p *taskErrorInjectionPersistenceClient) Close() {
	p.persistence.Close()
}

func (p *metadataErrorInjectionPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *metadataErrorInjectionPersistenceClient) CreateDomain(
	ctx context.Context,
	request *persistence.CreateDomainRequest,
) (*persistence.CreateDomainResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.CreateDomainResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.CreateDomain(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationCreateDomain,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *metadataErrorInjectionPersistenceClient) GetDomain(
	ctx context.Context,
	request *persistence.GetDomainRequest,
) (*persistence.GetDomainResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetDomainResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetDomain(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetDomain,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *metadataErrorInjectionPersistenceClient) UpdateDomain(
	ctx context.Context,
	request *persistence.UpdateDomainRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.UpdateDomain(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationUpdateDomain,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *metadataErrorInjectionPersistenceClient) DeleteDomain(
	ctx context.Context,
	request *persistence.DeleteDomainRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.DeleteDomain(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationDeleteDomain,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *metadataErrorInjectionPersistenceClient) DeleteDomainByName(
	ctx context.Context,
	request *persistence.DeleteDomainByNameRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.DeleteDomainByName(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationDeleteDomainByName,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *metadataErrorInjectionPersistenceClient) ListDomains(
	ctx context.Context,
	request *persistence.ListDomainsRequest,
) (*persistence.ListDomainsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ListDomainsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListDomains(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListDomains,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *metadataErrorInjectionPersistenceClient) GetMetadata(
	ctx context.Context,
) (*persistence.GetMetadataResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetMetadataResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetMetadata(ctx)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetMetadata,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *metadataErrorInjectionPersistenceClient) Close() {
	p.persistence.Close()
}

func (p *visibilityErrorInjectionPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *visibilityErrorInjectionPersistenceClient) RecordWorkflowExecutionStarted(
	ctx context.Context,
	request *visibility.RecordWorkflowExecutionStartedRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.RecordWorkflowExecutionStarted(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationRecordWorkflowExecutionStarted,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *visibilityErrorInjectionPersistenceClient) RecordWorkflowExecutionClosed(
	ctx context.Context,
	request *visibility.RecordWorkflowExecutionClosedRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.RecordWorkflowExecutionClosed(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationRecordWorkflowExecutionClosed,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *visibilityErrorInjectionPersistenceClient) UpsertWorkflowExecution(
	ctx context.Context,
	request *visibility.UpsertWorkflowExecutionRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.UpsertWorkflowExecution(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationUpsertWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *visibilityErrorInjectionPersistenceClient) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *visibility.ListWorkflowExecutionsRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *visibility.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListOpenWorkflowExecutions(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListOpenWorkflowExecutions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityErrorInjectionPersistenceClient) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *visibility.ListWorkflowExecutionsRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *visibility.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListClosedWorkflowExecutions(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListClosedWorkflowExecutions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityErrorInjectionPersistenceClient) ListOpenWorkflowExecutionsByType(
	ctx context.Context,
	request *visibility.ListWorkflowExecutionsByTypeRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *visibility.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListOpenWorkflowExecutionsByType(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListOpenWorkflowExecutionsByType,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityErrorInjectionPersistenceClient) ListClosedWorkflowExecutionsByType(
	ctx context.Context,
	request *visibility.ListWorkflowExecutionsByTypeRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *visibility.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListClosedWorkflowExecutionsByType(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListClosedWorkflowExecutionsByType,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityErrorInjectionPersistenceClient) ListOpenWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *visibility.ListWorkflowExecutionsByWorkflowIDRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *visibility.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListOpenWorkflowExecutionsByWorkflowID(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListOpenWorkflowExecutionsByWorkflowID,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityErrorInjectionPersistenceClient) ListClosedWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *visibility.ListWorkflowExecutionsByWorkflowIDRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *visibility.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListClosedWorkflowExecutionsByWorkflowID(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListClosedWorkflowExecutionsByWorkflowID,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityErrorInjectionPersistenceClient) ListClosedWorkflowExecutionsByStatus(
	ctx context.Context,
	request *visibility.ListClosedWorkflowExecutionsByStatusRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *visibility.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListClosedWorkflowExecutionsByStatus(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListClosedWorkflowExecutionsByStatus,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityErrorInjectionPersistenceClient) GetClosedWorkflowExecution(
	ctx context.Context,
	request *visibility.GetClosedWorkflowExecutionRequest,
) (*visibility.GetClosedWorkflowExecutionResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *visibility.GetClosedWorkflowExecutionResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetClosedWorkflowExecution(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetClosedWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityErrorInjectionPersistenceClient) DeleteWorkflowExecution(
	ctx context.Context,
	request *visibility.VisibilityDeleteWorkflowExecutionRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.DeleteWorkflowExecution(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationVisibilityDeleteWorkflowExecution,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *visibilityErrorInjectionPersistenceClient) ListWorkflowExecutions(
	ctx context.Context,
	request *visibility.ListWorkflowExecutionsByQueryRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *visibility.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ListWorkflowExecutions(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationListWorkflowExecutions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityErrorInjectionPersistenceClient) ScanWorkflowExecutions(
	ctx context.Context,
	request *visibility.ListWorkflowExecutionsByQueryRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *visibility.ListWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ScanWorkflowExecutions(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationScanWorkflowExecutions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityErrorInjectionPersistenceClient) CountWorkflowExecutions(
	ctx context.Context,
	request *visibility.CountWorkflowExecutionsRequest,
) (*visibility.CountWorkflowExecutionsResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *visibility.CountWorkflowExecutionsResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.CountWorkflowExecutions(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationCountWorkflowExecutions,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *visibilityErrorInjectionPersistenceClient) Close() {
	p.persistence.Close()
}

func (p *historyErrorInjectionPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

// AppendHistoryNodes add(or override) a node to a history branch
func (p *historyErrorInjectionPersistenceClient) AppendHistoryNodes(
	ctx context.Context,
	request *persistence.AppendHistoryNodesRequest,
) (*persistence.AppendHistoryNodesResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.AppendHistoryNodesResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.AppendHistoryNodes(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationAppendHistoryNodes,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

// ReadHistoryBranch returns history node data for a branch
func (p *historyErrorInjectionPersistenceClient) ReadHistoryBranch(
	ctx context.Context,
	request *persistence.ReadHistoryBranchRequest,
) (*persistence.ReadHistoryBranchResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ReadHistoryBranchResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ReadHistoryBranch(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationReadHistoryBranch,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

// ReadHistoryBranchByBatch returns history node data for a branch
func (p *historyErrorInjectionPersistenceClient) ReadHistoryBranchByBatch(
	ctx context.Context,
	request *persistence.ReadHistoryBranchRequest,
) (*persistence.ReadHistoryBranchByBatchResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ReadHistoryBranchByBatchResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ReadHistoryBranchByBatch(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationReadHistoryBranchByBatch,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

// ReadHistoryBranchByBatch returns history node data for a branch
func (p *historyErrorInjectionPersistenceClient) ReadRawHistoryBranch(
	ctx context.Context,
	request *persistence.ReadHistoryBranchRequest,
) (*persistence.ReadRawHistoryBranchResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ReadRawHistoryBranchResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ReadRawHistoryBranch(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationReadRawHistoryBranch,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

// ForkHistoryBranch forks a new branch from a old branch
func (p *historyErrorInjectionPersistenceClient) ForkHistoryBranch(
	ctx context.Context,
	request *persistence.ForkHistoryBranchRequest,
) (*persistence.ForkHistoryBranchResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.ForkHistoryBranchResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ForkHistoryBranch(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationForkHistoryBranch,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

// DeleteHistoryBranch removes a branch
func (p *historyErrorInjectionPersistenceClient) DeleteHistoryBranch(
	ctx context.Context,
	request *persistence.DeleteHistoryBranchRequest,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.DeleteHistoryBranch(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationDeleteHistoryBranch,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

// GetHistoryTree returns all branch information of a tree
func (p *historyErrorInjectionPersistenceClient) GetHistoryTree(
	ctx context.Context,
	request *persistence.GetHistoryTreeRequest,
) (*persistence.GetHistoryTreeResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetHistoryTreeResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetHistoryTree(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetHistoryTree,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *historyErrorInjectionPersistenceClient) GetAllHistoryTreeBranches(
	ctx context.Context,
	request *persistence.GetAllHistoryTreeBranchesRequest,
) (*persistence.GetAllHistoryTreeBranchesResponse, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response *persistence.GetAllHistoryTreeBranchesResponse
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetAllHistoryTreeBranches(ctx, request)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetAllHistoryTreeBranches,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *historyErrorInjectionPersistenceClient) Close() {
	p.persistence.Close()
}

func (p *queueErrorInjectionPersistenceClient) EnqueueMessage(
	ctx context.Context,
	message []byte,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.EnqueueMessage(ctx, message)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationEnqueueMessage,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *queueErrorInjectionPersistenceClient) ReadMessages(
	ctx context.Context,
	lastMessageID int64,
	maxCount int,
) ([]*persistence.QueueMessage, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response []*persistence.QueueMessage
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.ReadMessages(ctx, lastMessageID, maxCount)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationReadMessages,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *queueErrorInjectionPersistenceClient) UpdateAckLevel(
	ctx context.Context,
	messageID int64,
	clusterName string,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.UpdateAckLevel(ctx, messageID, clusterName)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationUpdateAckLevel,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *queueErrorInjectionPersistenceClient) GetAckLevels(
	ctx context.Context,
) (map[string]int64, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response map[string]int64
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetAckLevels(ctx)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetAckLevels,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *queueErrorInjectionPersistenceClient) DeleteMessagesBefore(
	ctx context.Context,
	messageID int64,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.DeleteMessagesBefore(ctx, messageID)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationDeleteMessagesBefore,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *queueErrorInjectionPersistenceClient) EnqueueMessageToDLQ(
	ctx context.Context,
	message []byte,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.EnqueueMessageToDLQ(ctx, message)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationEnqueueMessageToDLQ,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *queueErrorInjectionPersistenceClient) ReadMessagesFromDLQ(
	ctx context.Context,
	firstMessageID int64,
	lastMessageID int64,
	pageSize int,
	pageToken []byte,
) ([]*persistence.QueueMessage, []byte, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response []*persistence.QueueMessage
	var token []byte
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, token, persistenceErr = p.persistence.ReadMessagesFromDLQ(ctx, firstMessageID, lastMessageID, pageSize, pageToken)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationReadMessagesFromDLQ,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, nil, fakeErr
	}
	return response, token, persistenceErr
}

func (p *queueErrorInjectionPersistenceClient) RangeDeleteMessagesFromDLQ(
	ctx context.Context,
	firstMessageID int64,
	lastMessageID int64,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.RangeDeleteMessagesFromDLQ(ctx, firstMessageID, lastMessageID)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationRangeDeleteMessagesFromDLQ,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *queueErrorInjectionPersistenceClient) UpdateDLQAckLevel(
	ctx context.Context,
	messageID int64,
	clusterName string,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.UpdateDLQAckLevel(ctx, messageID, clusterName)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationUpdateDLQAckLevel,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *queueErrorInjectionPersistenceClient) GetDLQAckLevels(
	ctx context.Context,
) (map[string]int64, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response map[string]int64
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetDLQAckLevels(ctx)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetDLQAckLevels,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return nil, fakeErr
	}
	return response, persistenceErr
}

func (p *queueErrorInjectionPersistenceClient) GetDLQSize(
	ctx context.Context,
) (int64, error) {
	fakeErr := generateFakeError(p.errorRate)

	var response int64
	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		response, persistenceErr = p.persistence.GetDLQSize(ctx)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationGetDLQSize,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return 0, fakeErr
	}
	return response, persistenceErr
}

func (p *queueErrorInjectionPersistenceClient) DeleteMessageFromDLQ(
	ctx context.Context,
	messageID int64,
) error {
	fakeErr := generateFakeError(p.errorRate)

	var persistenceErr error
	var forwardCall bool
	if forwardCall = shouldForwardCallToPersistence(fakeErr); forwardCall {
		persistenceErr = p.persistence.DeleteMessageFromDLQ(ctx, messageID)
	}

	if fakeErr != nil {
		p.logger.Error(msgInjectedFakeErr,
			tag.StoreOperationDeleteMessageFromDLQ,
			tag.Error(fakeErr),
			tag.Bool(forwardCall),
			tag.StoreError(persistenceErr),
		)
		return fakeErr
	}
	return persistenceErr
}

func (p *queueErrorInjectionPersistenceClient) Close() {
	p.persistence.Close()
}

func shouldForwardCallToPersistence(
	err error,
) bool {
	if err == nil {
		return true
	}

	if err == ErrFakeTimeout || err == errors.ErrFakeUnhandled {
		// forward the call with 50% chance
		return rand.Intn(2) == 0
	}

	return false
}

func generateFakeError(
	errorRate float64,
) error {
	if rand.Float64() < errorRate {
		return fakeErrors[rand.Intn(len(fakeErrors))]
	}

	return nil
}
