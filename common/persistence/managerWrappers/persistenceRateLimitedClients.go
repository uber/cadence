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

package managerWrappers

import (
	"context"

	persistence2 "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/visibility"

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
		persistence persistence2.ShardManager
		logger      log.Logger
	}

	workflowExecutionRateLimitedPersistenceClient struct {
		rateLimiter quotas.Limiter
		persistence persistence2.ExecutionManager
		logger      log.Logger
	}

	taskRateLimitedPersistenceClient struct {
		rateLimiter quotas.Limiter
		persistence persistence2.TaskManager
		logger      log.Logger
	}

	historyRateLimitedPersistenceClient struct {
		rateLimiter quotas.Limiter
		persistence persistence2.HistoryManager
		logger      log.Logger
	}

	metadataRateLimitedPersistenceClient struct {
		rateLimiter quotas.Limiter
		persistence persistence2.DomainManager
		logger      log.Logger
	}

	visibilityRateLimitedPersistenceClient struct {
		rateLimiter quotas.Limiter
		persistence visibility.VisibilityManager
		logger      log.Logger
	}

	queueRateLimitedPersistenceClient struct {
		rateLimiter quotas.Limiter
		persistence persistence2.QueueManager
		logger      log.Logger
	}
)

var _ persistence2.ShardManager = (*shardRateLimitedPersistenceClient)(nil)
var _ persistence2.ExecutionManager = (*workflowExecutionRateLimitedPersistenceClient)(nil)
var _ persistence2.TaskManager = (*taskRateLimitedPersistenceClient)(nil)
var _ persistence2.HistoryManager = (*historyRateLimitedPersistenceClient)(nil)
var _ persistence2.DomainManager = (*metadataRateLimitedPersistenceClient)(nil)
var _ visibility.VisibilityManager = (*visibilityRateLimitedPersistenceClient)(nil)
var _ persistence2.QueueManager = (*queueRateLimitedPersistenceClient)(nil)

// NewShardPersistenceRateLimitedClient creates a client to manage shards
func NewShardPersistenceRateLimitedClient(
	persistence persistence2.ShardManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) persistence2.ShardManager {
	return &shardRateLimitedPersistenceClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

// NewWorkflowExecutionPersistenceRateLimitedClient creates a client to manage executions
func NewWorkflowExecutionPersistenceRateLimitedClient(
	persistence persistence2.ExecutionManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) persistence2.ExecutionManager {
	return &workflowExecutionRateLimitedPersistenceClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

// NewTaskPersistenceRateLimitedClient creates a client to manage tasks
func NewTaskPersistenceRateLimitedClient(
	persistence persistence2.TaskManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) persistence2.TaskManager {
	return &taskRateLimitedPersistenceClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

// NewHistoryPersistenceRateLimitedClient creates a HistoryManager client to manage workflow execution history
func NewHistoryPersistenceRateLimitedClient(
	persistence persistence2.HistoryManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) persistence2.HistoryManager {
	return &historyRateLimitedPersistenceClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

// NewMetadataPersistenceRateLimitedClient creates a DomainManager client to manage metadata
func NewMetadataPersistenceRateLimitedClient(
	persistence persistence2.DomainManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) persistence2.DomainManager {
	return &metadataRateLimitedPersistenceClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

// NewVisibilityPersistenceRateLimitedClient creates a client to manage visibility
func NewVisibilityPersistenceRateLimitedClient(
	persistence visibility.VisibilityManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) visibility.VisibilityManager {
	return &visibilityRateLimitedPersistenceClient{
		persistence: persistence,
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

// NewQueuePersistenceRateLimitedClient creates a client to manage queue
func NewQueuePersistenceRateLimitedClient(
	persistence persistence2.QueueManager,
	rateLimiter quotas.Limiter,
	logger log.Logger,
) persistence2.QueueManager {
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
	request *persistence2.CreateShardRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CreateShard(ctx, request)
	return err
}

func (p *shardRateLimitedPersistenceClient) GetShard(
	ctx context.Context,
	request *persistence2.GetShardRequest,
) (*persistence2.GetShardResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetShard(ctx, request)
	return response, err
}

func (p *shardRateLimitedPersistenceClient) UpdateShard(
	ctx context.Context,
	request *persistence2.UpdateShardRequest,
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
	request *persistence2.CreateWorkflowExecutionRequest,
) (*persistence2.CreateWorkflowExecutionResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.CreateWorkflowExecution(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) GetWorkflowExecution(
	ctx context.Context,
	request *persistence2.GetWorkflowExecutionRequest,
) (*persistence2.GetWorkflowExecutionResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetWorkflowExecution(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) UpdateWorkflowExecution(
	ctx context.Context,
	request *persistence2.UpdateWorkflowExecutionRequest,
) (*persistence2.UpdateWorkflowExecutionResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	resp, err := p.persistence.UpdateWorkflowExecution(ctx, request)
	return resp, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) ConflictResolveWorkflowExecution(
	ctx context.Context,
	request *persistence2.ConflictResolveWorkflowExecutionRequest,
) (*persistence2.ConflictResolveWorkflowExecutionResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	resp, err := p.persistence.ConflictResolveWorkflowExecution(ctx, request)
	return resp, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) ResetWorkflowExecution(
	ctx context.Context,
	request *persistence2.ResetWorkflowExecutionRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.ResetWorkflowExecution(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) DeleteWorkflowExecution(
	ctx context.Context,
	request *persistence2.DeleteWorkflowExecutionRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.DeleteWorkflowExecution(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) DeleteCurrentWorkflowExecution(
	ctx context.Context,
	request *persistence2.DeleteCurrentWorkflowExecutionRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.DeleteCurrentWorkflowExecution(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) GetCurrentExecution(
	ctx context.Context,
	request *persistence2.GetCurrentExecutionRequest,
) (*persistence2.GetCurrentExecutionResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetCurrentExecution(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) ListCurrentExecutions(
	ctx context.Context,
	request *persistence2.ListCurrentExecutionsRequest,
) (*persistence2.ListCurrentExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListCurrentExecutions(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) IsWorkflowExecutionExists(
	ctx context.Context,
	request *persistence2.IsWorkflowExecutionExistsRequest,
) (*persistence2.IsWorkflowExecutionExistsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.IsWorkflowExecutionExists(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) ListConcreteExecutions(
	ctx context.Context,
	request *persistence2.ListConcreteExecutionsRequest,
) (*persistence2.ListConcreteExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListConcreteExecutions(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) GetTransferTasks(
	ctx context.Context,
	request *persistence2.GetTransferTasksRequest,
) (*persistence2.GetTransferTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetTransferTasks(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) GetCrossClusterTasks(
	ctx context.Context,
	request *persistence2.GetCrossClusterTasksRequest,
) (*persistence2.GetCrossClusterTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetCrossClusterTasks(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) GetReplicationTasks(
	ctx context.Context,
	request *persistence2.GetReplicationTasksRequest,
) (*persistence2.GetReplicationTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetReplicationTasks(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) CompleteTransferTask(
	ctx context.Context,
	request *persistence2.CompleteTransferTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CompleteTransferTask(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) RangeCompleteTransferTask(
	ctx context.Context,
	request *persistence2.RangeCompleteTransferTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.RangeCompleteTransferTask(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) CompleteCrossClusterTask(
	ctx context.Context,
	request *persistence2.CompleteCrossClusterTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CompleteCrossClusterTask(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) RangeCompleteCrossClusterTask(
	ctx context.Context,
	request *persistence2.RangeCompleteCrossClusterTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.RangeCompleteCrossClusterTask(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) CompleteReplicationTask(
	ctx context.Context,
	request *persistence2.CompleteReplicationTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CompleteReplicationTask(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) RangeCompleteReplicationTask(
	ctx context.Context,
	request *persistence2.RangeCompleteReplicationTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.RangeCompleteReplicationTask(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) PutReplicationTaskToDLQ(
	ctx context.Context,
	request *persistence2.PutReplicationTaskToDLQRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.PutReplicationTaskToDLQ(ctx, request)
}

func (p *workflowExecutionRateLimitedPersistenceClient) GetReplicationTasksFromDLQ(
	ctx context.Context,
	request *persistence2.GetReplicationTasksFromDLQRequest,
) (*persistence2.GetReplicationTasksFromDLQResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.GetReplicationTasksFromDLQ(ctx, request)
}

func (p *workflowExecutionRateLimitedPersistenceClient) GetReplicationDLQSize(
	ctx context.Context,
	request *persistence2.GetReplicationDLQSizeRequest,
) (*persistence2.GetReplicationDLQSizeResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	return p.persistence.GetReplicationDLQSize(ctx, request)
}

func (p *workflowExecutionRateLimitedPersistenceClient) DeleteReplicationTaskFromDLQ(
	ctx context.Context,
	request *persistence2.DeleteReplicationTaskFromDLQRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.DeleteReplicationTaskFromDLQ(ctx, request)
}

func (p *workflowExecutionRateLimitedPersistenceClient) RangeDeleteReplicationTaskFromDLQ(
	ctx context.Context,
	request *persistence2.RangeDeleteReplicationTaskFromDLQRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	return p.persistence.RangeDeleteReplicationTaskFromDLQ(ctx, request)
}

func (p *workflowExecutionRateLimitedPersistenceClient) CreateFailoverMarkerTasks(
	ctx context.Context,
	request *persistence2.CreateFailoverMarkersRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CreateFailoverMarkerTasks(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) GetTimerIndexTasks(
	ctx context.Context,
	request *persistence2.GetTimerIndexTasksRequest,
) (*persistence2.GetTimerIndexTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetTimerIndexTasks(ctx, request)
	return response, err
}

func (p *workflowExecutionRateLimitedPersistenceClient) CompleteTimerTask(
	ctx context.Context,
	request *persistence2.CompleteTimerTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CompleteTimerTask(ctx, request)
	return err
}

func (p *workflowExecutionRateLimitedPersistenceClient) RangeCompleteTimerTask(
	ctx context.Context,
	request *persistence2.RangeCompleteTimerTaskRequest,
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
	request *persistence2.CreateTasksRequest,
) (*persistence2.CreateTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.CreateTasks(ctx, request)
	return response, err
}

func (p *taskRateLimitedPersistenceClient) GetTasks(
	ctx context.Context,
	request *persistence2.GetTasksRequest,
) (*persistence2.GetTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetTasks(ctx, request)
	return response, err
}

func (p *taskRateLimitedPersistenceClient) CompleteTask(
	ctx context.Context,
	request *persistence2.CompleteTaskRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.CompleteTask(ctx, request)
	return err
}

func (p *taskRateLimitedPersistenceClient) CompleteTasksLessThan(
	ctx context.Context,
	request *persistence2.CompleteTasksLessThanRequest,
) (int, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return 0, ErrPersistenceLimitExceeded
	}
	return p.persistence.CompleteTasksLessThan(ctx, request)
}

func (p *taskRateLimitedPersistenceClient) GetOrphanTasks(ctx context.Context, request *persistence2.GetOrphanTasksRequest) (*persistence2.GetOrphanTasksResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.GetOrphanTasks(ctx, request)
}

func (p *taskRateLimitedPersistenceClient) LeaseTaskList(
	ctx context.Context,
	request *persistence2.LeaseTaskListRequest,
) (*persistence2.LeaseTaskListResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.LeaseTaskList(ctx, request)
	return response, err
}

func (p *taskRateLimitedPersistenceClient) UpdateTaskList(
	ctx context.Context,
	request *persistence2.UpdateTaskListRequest,
) (*persistence2.UpdateTaskListResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.UpdateTaskList(ctx, request)
	return response, err
}

func (p *taskRateLimitedPersistenceClient) ListTaskList(
	ctx context.Context,
	request *persistence2.ListTaskListRequest,
) (*persistence2.ListTaskListResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.ListTaskList(ctx, request)
}

func (p *taskRateLimitedPersistenceClient) DeleteTaskList(
	ctx context.Context,
	request *persistence2.DeleteTaskListRequest,
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
	request *persistence2.CreateDomainRequest,
) (*persistence2.CreateDomainResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.CreateDomain(ctx, request)
	return response, err
}

func (p *metadataRateLimitedPersistenceClient) GetDomain(
	ctx context.Context,
	request *persistence2.GetDomainRequest,
) (*persistence2.GetDomainResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetDomain(ctx, request)
	return response, err
}

func (p *metadataRateLimitedPersistenceClient) UpdateDomain(
	ctx context.Context,
	request *persistence2.UpdateDomainRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.UpdateDomain(ctx, request)
	return err
}

func (p *metadataRateLimitedPersistenceClient) DeleteDomain(
	ctx context.Context,
	request *persistence2.DeleteDomainRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.DeleteDomain(ctx, request)
	return err
}

func (p *metadataRateLimitedPersistenceClient) DeleteDomainByName(
	ctx context.Context,
	request *persistence2.DeleteDomainByNameRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.DeleteDomainByName(ctx, request)
	return err
}

func (p *metadataRateLimitedPersistenceClient) ListDomains(
	ctx context.Context,
	request *persistence2.ListDomainsRequest,
) (*persistence2.ListDomainsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListDomains(ctx, request)
	return response, err
}

func (p *metadataRateLimitedPersistenceClient) GetMetadata(
	ctx context.Context,
) (*persistence2.GetMetadataResponse, error) {
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
	request *visibility.RecordWorkflowExecutionStartedRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.RecordWorkflowExecutionStarted(ctx, request)
	return err
}

func (p *visibilityRateLimitedPersistenceClient) RecordWorkflowExecutionClosed(
	ctx context.Context,
	request *visibility.RecordWorkflowExecutionClosedRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.RecordWorkflowExecutionClosed(ctx, request)
	return err
}

func (p *visibilityRateLimitedPersistenceClient) UpsertWorkflowExecution(
	ctx context.Context,
	request *visibility.UpsertWorkflowExecutionRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}

	err := p.persistence.UpsertWorkflowExecution(ctx, request)
	return err
}

func (p *visibilityRateLimitedPersistenceClient) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *visibility.ListWorkflowExecutionsRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListOpenWorkflowExecutions(ctx, request)
	return response, err
}

func (p *visibilityRateLimitedPersistenceClient) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *visibility.ListWorkflowExecutionsRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListClosedWorkflowExecutions(ctx, request)
	return response, err
}

func (p *visibilityRateLimitedPersistenceClient) ListOpenWorkflowExecutionsByType(
	ctx context.Context,
	request *visibility.ListWorkflowExecutionsByTypeRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListOpenWorkflowExecutionsByType(ctx, request)
	return response, err
}

func (p *visibilityRateLimitedPersistenceClient) ListClosedWorkflowExecutionsByType(
	ctx context.Context,
	request *visibility.ListWorkflowExecutionsByTypeRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListClosedWorkflowExecutionsByType(ctx, request)
	return response, err
}

func (p *visibilityRateLimitedPersistenceClient) ListOpenWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *visibility.ListWorkflowExecutionsByWorkflowIDRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListOpenWorkflowExecutionsByWorkflowID(ctx, request)
	return response, err
}

func (p *visibilityRateLimitedPersistenceClient) ListClosedWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *visibility.ListWorkflowExecutionsByWorkflowIDRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListClosedWorkflowExecutionsByWorkflowID(ctx, request)
	return response, err
}

func (p *visibilityRateLimitedPersistenceClient) ListClosedWorkflowExecutionsByStatus(
	ctx context.Context,
	request *visibility.ListClosedWorkflowExecutionsByStatusRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.ListClosedWorkflowExecutionsByStatus(ctx, request)
	return response, err
}

func (p *visibilityRateLimitedPersistenceClient) GetClosedWorkflowExecution(
	ctx context.Context,
	request *visibility.GetClosedWorkflowExecutionRequest,
) (*visibility.GetClosedWorkflowExecutionResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}

	response, err := p.persistence.GetClosedWorkflowExecution(ctx, request)
	return response, err
}

func (p *visibilityRateLimitedPersistenceClient) DeleteWorkflowExecution(
	ctx context.Context,
	request *visibility.VisibilityDeleteWorkflowExecutionRequest,
) error {
	if ok := p.rateLimiter.Allow(); !ok {
		return ErrPersistenceLimitExceeded
	}
	return p.persistence.DeleteWorkflowExecution(ctx, request)
}

func (p *visibilityRateLimitedPersistenceClient) ListWorkflowExecutions(
	ctx context.Context,
	request *visibility.ListWorkflowExecutionsByQueryRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.ListWorkflowExecutions(ctx, request)
}

func (p *visibilityRateLimitedPersistenceClient) ScanWorkflowExecutions(
	ctx context.Context,
	request *visibility.ListWorkflowExecutionsByQueryRequest,
) (*visibility.ListWorkflowExecutionsResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.ScanWorkflowExecutions(ctx, request)
}

func (p *visibilityRateLimitedPersistenceClient) CountWorkflowExecutions(
	ctx context.Context,
	request *visibility.CountWorkflowExecutionsRequest,
) (*visibility.CountWorkflowExecutionsResponse, error) {
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
	request *persistence2.AppendHistoryNodesRequest,
) (*persistence2.AppendHistoryNodesResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	return p.persistence.AppendHistoryNodes(ctx, request)
}

// ReadHistoryBranch returns history node data for a branch
func (p *historyRateLimitedPersistenceClient) ReadHistoryBranch(
	ctx context.Context,
	request *persistence2.ReadHistoryBranchRequest,
) (*persistence2.ReadHistoryBranchResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	response, err := p.persistence.ReadHistoryBranch(ctx, request)
	return response, err
}

// ReadHistoryBranchByBatch returns history node data for a branch
func (p *historyRateLimitedPersistenceClient) ReadHistoryBranchByBatch(
	ctx context.Context,
	request *persistence2.ReadHistoryBranchRequest,
) (*persistence2.ReadHistoryBranchByBatchResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	response, err := p.persistence.ReadHistoryBranchByBatch(ctx, request)
	return response, err
}

// ReadHistoryBranchByBatch returns history node data for a branch
func (p *historyRateLimitedPersistenceClient) ReadRawHistoryBranch(
	ctx context.Context,
	request *persistence2.ReadHistoryBranchRequest,
) (*persistence2.ReadRawHistoryBranchResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	response, err := p.persistence.ReadRawHistoryBranch(ctx, request)
	return response, err
}

// ForkHistoryBranch forks a new branch from a old branch
func (p *historyRateLimitedPersistenceClient) ForkHistoryBranch(
	ctx context.Context,
	request *persistence2.ForkHistoryBranchRequest,
) (*persistence2.ForkHistoryBranchResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	response, err := p.persistence.ForkHistoryBranch(ctx, request)
	return response, err
}

// DeleteHistoryBranch removes a branch
func (p *historyRateLimitedPersistenceClient) DeleteHistoryBranch(
	ctx context.Context,
	request *persistence2.DeleteHistoryBranchRequest,
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
	request *persistence2.GetHistoryTreeRequest,
) (*persistence2.GetHistoryTreeResponse, error) {
	if ok := p.rateLimiter.Allow(); !ok {
		return nil, ErrPersistenceLimitExceeded
	}
	response, err := p.persistence.GetHistoryTree(ctx, request)
	return response, err
}

func (p *historyRateLimitedPersistenceClient) GetAllHistoryTreeBranches(
	ctx context.Context,
	request *persistence2.GetAllHistoryTreeBranchesRequest,
) (*persistence2.GetAllHistoryTreeBranchesResponse, error) {
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
) ([]*persistence2.QueueMessage, error) {
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
) ([]*persistence2.QueueMessage, []byte, error) {
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
