// Copyright (c) 2017 Uber Technologies, Inc.
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

package persistence

import (
	"context"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/types"
)

var (
	// ErrPersistenceLimitExceeded is the error indicating QPS limit reached.
	ErrPersistenceLimitExceeded = &types.ServiceBusyError{Message: "Persistence Max QPS Reached."}
)

type (
	shardRateLimitedPersistenceClient struct {
		rateLimiter quotas.Limiter
		persistence ShardManager
		logger      log.Logger
	}

	workflowExecutionRateLimitedPersistenceClient struct {
		rateLimiter quotas.Limiter
		persistence ExecutionManager
		logger      log.Logger
	}

	taskRateLimitedPersistenceClient struct {
		rateLimiter quotas.Limiter
		persistence TaskManager
		logger      log.Logger
	}

	historyRateLimitedPersistenceClient struct {
		rateLimiter quotas.Limiter
		persistence HistoryManager
		logger      log.Logger
	}

	metadataRateLimitedPersistenceClient struct {
		rateLimiter quotas.Limiter
		persistence MetadataManager
		logger      log.Logger
	}

	visibilityRateLimitedPersistenceClient struct {
		rateLimiter quotas.Limiter
		persistence VisibilityManager
		logger      log.Logger
	}

	queueRateLimitedPersistenceClient struct {
		rateLimiter quotas.Limiter
		persistence QueueManager
		logger      log.Logger
	}
)

var _ ShardManager = (*shardRateLimitedPersistenceClient)(nil)
var _ ExecutionManager = (*workflowExecutionRateLimitedPersistenceClient)(nil)
var _ TaskManager = (*taskRateLimitedPersistenceClient)(nil)
var _ HistoryManager = (*historyRateLimitedPersistenceClient)(nil)
var _ MetadataManager = (*metadataRateLimitedPersistenceClient)(nil)
var _ VisibilityManager = (*visibilityRateLimitedPersistenceClient)(nil)
var _ QueueManager = (*queueRateLimitedPersistenceClient)(nil)

// NewShardPersistenceRateLimitedClient creates a client to manage shards
func NewShardPersistenceRateLimitedClient(
	persistence ShardManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) ShardManager {
	return &shardRateLimitedPersistenceClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

// NewWorkflowExecutionPersistenceRateLimitedClient creates a client to manage executions
func NewWorkflowExecutionPersistenceRateLimitedClient(
	persistence ExecutionManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) ExecutionManager {
	return &workflowExecutionRateLimitedPersistenceClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

// NewTaskPersistenceRateLimitedClient creates a client to manage tasks
func NewTaskPersistenceRateLimitedClient(
	persistence TaskManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) TaskManager {
	return &taskRateLimitedPersistenceClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

// NewHistoryPersistenceRateLimitedClient creates a HistoryManager client to manage workflow execution history
func NewHistoryPersistenceRateLimitedClient(
	persistence HistoryManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) HistoryManager {
	return &historyRateLimitedPersistenceClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

// NewMetadataPersistenceRateLimitedClient creates a MetadataManager client to manage metadata
func NewMetadataPersistenceRateLimitedClient(
	persistence MetadataManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) MetadataManager {
	return &metadataRateLimitedPersistenceClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

// NewVisibilityPersistenceRateLimitedClient creates a client to manage visibility
func NewVisibilityPersistenceRateLimitedClient(
	persistence VisibilityManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) VisibilityManager {
	return &visibilityRateLimitedPersistenceClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

// NewQueuePersistenceRateLimitedClient creates a client to manage queue
func NewQueuePersistenceRateLimitedClient(
	persistence QueueManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) QueueManager {
	return &queueRateLimitedPersistenceClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

func (p *shardRateLimitedPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *shardRateLimitedPersistenceClient) CreateShard(
	ctx context.Context,
	request *CreateShardRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CreateShard(ctx, request)
	return err
}

func (p *shardRateLimitedPersistenceClient) GetShard(
	ctx context.Context,
	request *GetShardRequest,
) (*GetShardResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetShard(ctx, request)
	return response, err
}

func (p *shardRateLimitedPersistenceClient) UpdateShard(
	ctx context.Context,
	request *UpdateShardRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.UpdateShard(ctx, request)
	return err
}

func (p *shardRateLimitedPersistenceClient) Close() {
	p.persistence.Close()
}

func (p *workflowExecutionRateLimitedPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *workflowExecutionRateLimitedPersistenceClient) GetShardID() int {
	return p.persistence.GetShardID()
}

func (p *workflowExecutionRateLimitedPersistenceClient) CreateWorkflowExecution(
	ctx context.Context,
	request *CreateWorkflowExecutionRequest,
) (*CreateWorkflowExecutionResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.CreateWorkflowExecution(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) GetWorkflowExecution(
	ctx context.Context,
	request *GetWorkflowExecutionRequest,
) (*GetWorkflowExecutionResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetWorkflowExecution(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) UpdateWorkflowExecution(
	ctx context.Context,
	request *UpdateWorkflowExecutionRequest,
) (*UpdateWorkflowExecutionResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	resp, err := p.persistence.UpdateWorkflowExecution(ctx, request)
	return resp, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) ConflictResolveWorkflowExecution(
	ctx context.Context,
	request *ConflictResolveWorkflowExecutionRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.ConflictResolveWorkflowExecution(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) ResetWorkflowExecution(
	ctx context.Context,
	request *ResetWorkflowExecutionRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.ResetWorkflowExecution(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) DeleteWorkflowExecution(
	ctx context.Context,
	request *DeleteWorkflowExecutionRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.DeleteWorkflowExecution(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) DeleteCurrentWorkflowExecution(
	ctx context.Context,
	request *DeleteCurrentWorkflowExecutionRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.DeleteCurrentWorkflowExecution(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) GetCurrentExecution(
	ctx context.Context,
	request *GetCurrentExecutionRequest,
) (*GetCurrentExecutionResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetCurrentExecution(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) ListCurrentExecutions(
	ctx context.Context,
	request *ListCurrentExecutionsRequest,
) (*ListCurrentExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListCurrentExecutions(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) IsWorkflowExecutionExists(
	ctx context.Context,
	request *IsWorkflowExecutionExistsRequest,
) (*IsWorkflowExecutionExistsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.IsWorkflowExecutionExists(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) ListConcreteExecutions(
	ctx context.Context,
	request *ListConcreteExecutionsRequest,
) (*ListConcreteExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListConcreteExecutions(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) GetTransferTasks(
	ctx context.Context,
	request *GetTransferTasksRequest,
) (*GetTransferTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetTransferTasks(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) GetReplicationTasks(
	ctx context.Context,
	request *GetReplicationTasksRequest,
) (*GetReplicationTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetReplicationTasks(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) CompleteTransferTask(
	ctx context.Context,
	request *CompleteTransferTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CompleteTransferTask(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) RangeCompleteTransferTask(
	ctx context.Context,
	request *RangeCompleteTransferTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.RangeCompleteTransferTask(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) CompleteReplicationTask(
	ctx context.Context,
	request *CompleteReplicationTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CompleteReplicationTask(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) RangeCompleteReplicationTask(
	ctx context.Context,
	request *RangeCompleteReplicationTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.RangeCompleteReplicationTask(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) PutReplicationTaskToDLQ(
	ctx context.Context,
	request *PutReplicationTaskToDLQRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.PutReplicationTaskToDLQ(ctx, request)
}

func (p *workflowExecutionRateLimitedPersistenceClient) GetReplicationTasksFromDLQ(
	ctx context.Context,
	request *GetReplicationTasksFromDLQRequest,
) (*GetReplicationTasksFromDLQResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.GetReplicationTasksFromDLQ(ctx, request)
}

func (p *workflowExecutionRateLimitedPersistenceClient) GetReplicationDLQSize(
	ctx context.Context,
	request *GetReplicationDLQSizeRequest,
) (*GetReplicationDLQSizeResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.GetReplicationDLQSize(ctx, request)
}

func (p *workflowExecutionRateLimitedPersistenceClient) DeleteReplicationTaskFromDLQ(
	ctx context.Context,
	request *DeleteReplicationTaskFromDLQRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.DeleteReplicationTaskFromDLQ(ctx, request)
}

func (p *workflowExecutionRateLimitedPersistenceClient) RangeDeleteReplicationTaskFromDLQ(
	ctx context.Context,
	request *RangeDeleteReplicationTaskFromDLQRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.RangeDeleteReplicationTaskFromDLQ(ctx, request)
}

func (p *workflowExecutionRateLimitedPersistenceClient) CreateFailoverMarkerTasks(
	ctx context.Context,
	request *CreateFailoverMarkersRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CreateFailoverMarkerTasks(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) GetTimerIndexTasks(
	ctx context.Context,
	request *GetTimerIndexTasksRequest,
) (*GetTimerIndexTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetTimerIndexTasks(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) CompleteTimerTask(
	ctx context.Context,
	request *CompleteTimerTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CompleteTimerTask(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) RangeCompleteTimerTask(
	ctx context.Context,
	request *RangeCompleteTimerTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.RangeCompleteTimerTask(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) Close() {
	p.persistence.Close()
}

func (p *taskRateLimitedPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *taskRateLimitedPersistenceClient) CreateTasks(
	ctx context.Context,
	request *CreateTasksRequest,
) (*CreateTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.CreateTasks(ctx, request)
	return response, err
}

func (p *taskRateLimitedPersistenceClient) GetTasks(
	ctx context.Context,
	request *GetTasksRequest,
) (*GetTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetTasks(ctx, request)
	return response, err
}

func (p *taskRateLimitedPersistenceClient) CompleteTask(
	ctx context.Context,
	request *CompleteTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CompleteTask(ctx, request)
	return err
}

func (p *taskRateLimitedPersistenceClient) CompleteTasksLessThan(
	ctx context.Context,
	request *CompleteTasksLessThanRequest,
) (int, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return 0, ErrPersistenceLimitExceeded
	}
	return p.persistence.CompleteTasksLessThan(ctx, request)
}

func (p *taskRateLimitedPersistenceClient) LeaseTaskList(
	ctx context.Context,
	request *LeaseTaskListRequest,
) (*LeaseTaskListResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.LeaseTaskList(ctx, request)
	return response, err
}

func (p *taskRateLimitedPersistenceClient) UpdateTaskList(
	ctx context.Context,
	request *UpdateTaskListRequest,
) (*UpdateTaskListResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.UpdateTaskList(ctx, request)
	return response, err
}

func (p *taskRateLimitedPersistenceClient) ListTaskList(
	ctx context.Context,
	request *ListTaskListRequest,
) (*ListTaskListResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.ListTaskList(ctx, request)
}

func (p *taskRateLimitedPersistenceClient) DeleteTaskList(
	ctx context.Context,
	request *DeleteTaskListRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}
	return p.persistence.DeleteTaskList(ctx, request)
}

func (p *taskRateLimitedPersistenceClient) Close() {
	p.persistence.Close()
}

func (p *metadataRateLimitedPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *metadataRateLimitedPersistenceClient) CreateDomain(
	ctx context.Context,
	request *CreateDomainRequest,
) (*CreateDomainResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.CreateDomain(ctx, request)
	return response, err
}

func (p *metadataRateLimitedPersistenceClient) GetDomain(
	ctx context.Context,
	request *GetDomainRequest,
) (*GetDomainResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetDomain(ctx, request)
	return response, err
}

func (p *metadataRateLimitedPersistenceClient) UpdateDomain(
	ctx context.Context,
	request *UpdateDomainRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.UpdateDomain(ctx, request)
	return err
}

func (p *metadataRateLimitedPersistenceClient) DeleteDomain(
	ctx context.Context,
	request *DeleteDomainRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.DeleteDomain(ctx, request)
	return err
}

func (p *metadataRateLimitedPersistenceClient) DeleteDomainByName(
	ctx context.Context,
	request *DeleteDomainByNameRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.DeleteDomainByName(ctx, request)
	return err
}

func (p *metadataRateLimitedPersistenceClient) ListDomains(
	ctx context.Context,
	request *ListDomainsRequest,
) (*ListDomainsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListDomains(ctx, request)
	return response, err
}

func (p *metadataRateLimitedPersistenceClient) GetMetadata(
	ctx context.Context,
) (*GetMetadataResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetMetadata(ctx)
	return response, err
}

func (p *metadataRateLimitedPersistenceClient) Close() {
	p.persistence.Close()
}

func (p *visibilityRateLimitedPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *visibilityRateLimitedPersistenceClient) RecordWorkflowExecutionStarted(
	ctx context.Context,
	request *RecordWorkflowExecutionStartedRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.RecordWorkflowExecutionStarted(ctx, request)
	return err
}

func (p *visibilityRateLimitedPersistenceClient) RecordWorkflowExecutionClosed(
	ctx context.Context,
	request *RecordWorkflowExecutionClosedRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.RecordWorkflowExecutionClosed(ctx, request)
	return err
}

func (p *visibilityRateLimitedPersistenceClient) UpsertWorkflowExecution(
	ctx context.Context,
	request *UpsertWorkflowExecutionRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.UpsertWorkflowExecution(ctx, request)
	return err
}

func (p *visibilityRateLimitedPersistenceClient) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsRequest,
) (*ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListOpenWorkflowExecutions(ctx, request)
	return response, err
}

func (p *visibilityRateLimitedPersistenceClient) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsRequest,
) (*ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListClosedWorkflowExecutions(ctx, request)
	return response, err
}

func (p *visibilityRateLimitedPersistenceClient) ListOpenWorkflowExecutionsByType(
	ctx context.Context,
	request *ListWorkflowExecutionsByTypeRequest,
) (*ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListOpenWorkflowExecutionsByType(ctx, request)
	return response, err
}

func (p *visibilityRateLimitedPersistenceClient) ListClosedWorkflowExecutionsByType(
	ctx context.Context,
	request *ListWorkflowExecutionsByTypeRequest,
) (*ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListClosedWorkflowExecutionsByType(ctx, request)
	return response, err
}

func (p *visibilityRateLimitedPersistenceClient) ListOpenWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *ListWorkflowExecutionsByWorkflowIDRequest,
) (*ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListOpenWorkflowExecutionsByWorkflowID(ctx, request)
	return response, err
}

func (p *visibilityRateLimitedPersistenceClient) ListClosedWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *ListWorkflowExecutionsByWorkflowIDRequest,
) (*ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListClosedWorkflowExecutionsByWorkflowID(ctx, request)
	return response, err
}

func (p *visibilityRateLimitedPersistenceClient) ListClosedWorkflowExecutionsByStatus(
	ctx context.Context,
	request *ListClosedWorkflowExecutionsByStatusRequest,
) (*ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListClosedWorkflowExecutionsByStatus(ctx, request)
	return response, err
}

func (p *visibilityRateLimitedPersistenceClient) GetClosedWorkflowExecution(
	ctx context.Context,
	request *GetClosedWorkflowExecutionRequest,
) (*GetClosedWorkflowExecutionResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetClosedWorkflowExecution(ctx, request)
	return response, err
}

func (p *visibilityRateLimitedPersistenceClient) DeleteWorkflowExecution(
	ctx context.Context,
	request *VisibilityDeleteWorkflowExecutionRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}
	return p.persistence.DeleteWorkflowExecution(ctx, request)
}

func (p *visibilityRateLimitedPersistenceClient) ListWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsByQueryRequest,
) (*ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.ListWorkflowExecutions(ctx, request)
}

func (p *visibilityRateLimitedPersistenceClient) ScanWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsByQueryRequest,
) (*ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.ScanWorkflowExecutions(ctx, request)
}

func (p *visibilityRateLimitedPersistenceClient) CountWorkflowExecutions(
	ctx context.Context,
	request *CountWorkflowExecutionsRequest,
) (*CountWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.CountWorkflowExecutions(ctx, request)
}

func (p *visibilityRateLimitedPersistenceClient) Close() {
	p.persistence.Close()
}

func (p *historyRateLimitedPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *historyRateLimitedPersistenceClient) Close() {
	p.persistence.Close()
}

// AppendHistoryNodes add(or override) a node to a history branch
func (p *historyRateLimitedPersistenceClient) AppendHistoryNodes(
	ctx context.Context,
	request *AppendHistoryNodesRequest,
) (*AppendHistoryNodesResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.AppendHistoryNodes(ctx, request)
}

// ReadHistoryBranch returns history node data for a branch
func (p *historyRateLimitedPersistenceClient) ReadHistoryBranch(
	ctx context.Context,
	request *ReadHistoryBranchRequest,
) (*ReadHistoryBranchResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	response, err := p.persistence.ReadHistoryBranch(ctx, request)
	return response, err
}

// ReadHistoryBranchByBatch returns history node data for a branch
func (p *historyRateLimitedPersistenceClient) ReadHistoryBranchByBatch(
	ctx context.Context,
	request *ReadHistoryBranchRequest,
) (*ReadHistoryBranchByBatchResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	response, err := p.persistence.ReadHistoryBranchByBatch(ctx, request)
	return response, err
}

// ReadHistoryBranchByBatch returns history node data for a branch
func (p *historyRateLimitedPersistenceClient) ReadRawHistoryBranch(
	ctx context.Context,
	request *ReadHistoryBranchRequest,
) (*ReadRawHistoryBranchResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	response, err := p.persistence.ReadRawHistoryBranch(ctx, request)
	return response, err
}

// ForkHistoryBranch forks a new branch from a old branch
func (p *historyRateLimitedPersistenceClient) ForkHistoryBranch(
	ctx context.Context,
	request *ForkHistoryBranchRequest,
) (*ForkHistoryBranchResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	response, err := p.persistence.ForkHistoryBranch(ctx, request)
	return response, err
}

// DeleteHistoryBranch removes a branch
func (p *historyRateLimitedPersistenceClient) DeleteHistoryBranch(
	ctx context.Context,
	request *DeleteHistoryBranchRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}
	err := p.persistence.DeleteHistoryBranch(ctx, request)
	return err
}

// GetHistoryTree returns all branch information of a tree
func (p *historyRateLimitedPersistenceClient) GetHistoryTree(
	ctx context.Context,
	request *GetHistoryTreeRequest,
) (*GetHistoryTreeResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	response, err := p.persistence.GetHistoryTree(ctx, request)
	return response, err
}

func (p *historyRateLimitedPersistenceClient) GetAllHistoryTreeBranches(
	ctx context.Context,
	request *GetAllHistoryTreeBranchesRequest,
) (*GetAllHistoryTreeBranchesResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	response, err := p.persistence.GetAllHistoryTreeBranches(ctx, request)
	return response, err
}

func (p *queueRateLimitedPersistenceClient) EnqueueMessage(
	ctx context.Context,
	message []byte,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.EnqueueMessage(ctx, message)
}

func (p *queueRateLimitedPersistenceClient) ReadMessages(
	ctx context.Context,
	lastMessageID int64,
	maxCount int,
) ([]*QueueMessage, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.ReadMessages(ctx, lastMessageID, maxCount)
}

func (p *queueRateLimitedPersistenceClient) UpdateAckLevel(
	ctx context.Context,
	messageID int64,
	clusterName string,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.UpdateAckLevel(ctx, messageID, clusterName)
}

func (p *queueRateLimitedPersistenceClient) GetAckLevels(
	ctx context.Context,
) (map[string]int64, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.GetAckLevels(ctx)
}

func (p *queueRateLimitedPersistenceClient) DeleteMessagesBefore(
	ctx context.Context,
	messageID int64,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.DeleteMessagesBefore(ctx, messageID)
}

func (p *queueRateLimitedPersistenceClient) EnqueueMessageToDLQ(
	ctx context.Context,
	message []byte,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.EnqueueMessageToDLQ(ctx, message)
}

func (p *queueRateLimitedPersistenceClient) ReadMessagesFromDLQ(
	ctx context.Context,
	firstMessageID int64,
	lastMessageID int64,
	pageSize int,
	pageToken []byte,
) ([]*QueueMessage, []byte, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.ReadMessagesFromDLQ(ctx, firstMessageID, lastMessageID, pageSize, pageToken)
}

func (p *queueRateLimitedPersistenceClient) RangeDeleteMessagesFromDLQ(
	ctx context.Context,
	firstMessageID int64,
	lastMessageID int64,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.RangeDeleteMessagesFromDLQ(ctx, firstMessageID, lastMessageID)
}

func (p *queueRateLimitedPersistenceClient) UpdateDLQAckLevel(
	ctx context.Context,
	messageID int64,
	clusterName string,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.UpdateDLQAckLevel(ctx, messageID, clusterName)
}

func (p *queueRateLimitedPersistenceClient) GetDLQAckLevels(
	ctx context.Context,
) (map[string]int64, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.GetDLQAckLevels(ctx)
}

func (p *queueRateLimitedPersistenceClient) GetDLQSize(
	ctx context.Context,
) (int64, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return 0, ErrPersistenceLimitExceeded
	}

	return p.persistence.GetDLQSize(ctx)
}

func (p *queueRateLimitedPersistenceClient) DeleteMessageFromDLQ(
	ctx context.Context,
	messageID int64,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.DeleteMessageFromDLQ(ctx, messageID)
}

func (p *queueRateLimitedPersistenceClient) Close() {
	p.persistence.Close()
}
