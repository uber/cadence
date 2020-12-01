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
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
)

type (
	shardPersistenceClient struct {
		metricClient metrics.Client
		persistence  ShardManager
		logger       log.Logger
	}

	workflowExecutionPersistenceClient struct {
		metricClient metrics.Client
		persistence  ExecutionManager
		logger       log.Logger
	}

	taskPersistenceClient struct {
		metricClient metrics.Client
		persistence  TaskManager
		logger       log.Logger
	}

	historyPersistenceClient struct {
		metricClient metrics.Client
		persistence  HistoryManager
		logger       log.Logger
	}

	metadataPersistenceClient struct {
		metricClient metrics.Client
		persistence  MetadataManager
		logger       log.Logger
	}

	visibilityPersistenceClient struct {
		metricClient metrics.Client
		persistence  VisibilityManager
		logger       log.Logger
	}

	queuePersistenceClient struct {
		metricClient metrics.Client
		persistence  QueueManager
		logger       log.Logger
	}
)

var _ ShardManager = (*shardPersistenceClient)(nil)
var _ ExecutionManager = (*workflowExecutionPersistenceClient)(nil)
var _ TaskManager = (*taskPersistenceClient)(nil)
var _ HistoryManager = (*historyPersistenceClient)(nil)
var _ MetadataManager = (*metadataPersistenceClient)(nil)
var _ VisibilityManager = (*visibilityPersistenceClient)(nil)
var _ QueueManager = (*queuePersistenceClient)(nil)

// NewShardPersistenceMetricsClient creates a client to manage shards
func NewShardPersistenceMetricsClient(
	persistence ShardManager,
	metricClient metrics.Client,
	logger log.Logger,
) ShardManager {
	return &shardPersistenceClient{
		persistence:  persistence,
		metricClient: metricClient,
		logger:       logger,
	}
}

// NewWorkflowExecutionPersistenceMetricsClient creates a client to manage executions
func NewWorkflowExecutionPersistenceMetricsClient(
	persistence ExecutionManager,
	metricClient metrics.Client,
	logger log.Logger,
) ExecutionManager {
	return &workflowExecutionPersistenceClient{
		persistence:  persistence,
		metricClient: metricClient,
		logger:       logger,
	}
}

// NewTaskPersistenceMetricsClient creates a client to manage tasks
func NewTaskPersistenceMetricsClient(
	persistence TaskManager,
	metricClient metrics.Client,
	logger log.Logger,
) TaskManager {
	return &taskPersistenceClient{
		persistence:  persistence,
		metricClient: metricClient,
		logger:       logger,
	}
}

// NewHistoryPersistenceMetricsClient creates a HistoryManager client to manage workflow execution history
func NewHistoryPersistenceMetricsClient(
	persistence HistoryManager,
	metricClient metrics.Client,
	logger log.Logger,
) HistoryManager {
	return &historyPersistenceClient{
		persistence:  persistence,
		metricClient: metricClient,
		logger:       logger,
	}
}

// NewMetadataPersistenceMetricsClient creates a MetadataManager client to manage metadata
func NewMetadataPersistenceMetricsClient(
	persistence MetadataManager,
	metricClient metrics.Client,
	logger log.Logger,
) MetadataManager {
	return &metadataPersistenceClient{
		persistence:  persistence,
		metricClient: metricClient,
		logger:       logger,
	}
}

// NewVisibilityPersistenceMetricsClient creates a client to manage visibility
func NewVisibilityPersistenceMetricsClient(
	persistence VisibilityManager,
	metricClient metrics.Client,
	logger log.Logger,
) VisibilityManager {
	return &visibilityPersistenceClient{
		persistence:  persistence,
		metricClient: metricClient,
		logger:       logger,
	}
}

// NewQueuePersistenceMetricsClient creates a client to manage queue
func NewQueuePersistenceMetricsClient(
	persistence QueueManager,
	metricClient metrics.Client,
	logger log.Logger,
) QueueManager {
	return &queuePersistenceClient{
		persistence:  persistence,
		metricClient: metricClient,
		logger:       logger,
	}
}

func (p *shardPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *shardPersistenceClient) CreateShard(
	ctx context.Context,
	request *CreateShardRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceCreateShardScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceCreateShardScope, metrics.PersistenceLatency)
	err := p.persistence.CreateShard(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceCreateShardScope, err)
	}

	return err
}

func (p *shardPersistenceClient) GetShard(
	ctx context.Context,
	request *GetShardRequest,
) (*GetShardResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceGetShardScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceGetShardScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetShard(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceGetShardScope, err)
	}

	return response, err
}

func (p *shardPersistenceClient) UpdateShard(
	ctx context.Context,
	request *UpdateShardRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceUpdateShardScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceUpdateShardScope, metrics.PersistenceLatency)
	err := p.persistence.UpdateShard(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceUpdateShardScope, err)
	}

	return err
}

func (p *shardPersistenceClient) updateErrorMetric(scope int, err error) {
	switch err.(type) {
	case *ShardAlreadyExistError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrShardExistsCounter)
	case *ShardOwnershipLostError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrShardOwnershipLostCounter)
	case *types.EntityNotExistsError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrEntityNotExistsCounter)
	case *types.ServiceBusyError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrBusyCounter)
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	default:
		p.logger.Error("Operation failed with internal error.", tag.Error(err), tag.MetricScope(scope))
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	}
}

func (p *shardPersistenceClient) Close() {
	p.persistence.Close()
}

func (p *workflowExecutionPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *workflowExecutionPersistenceClient) GetShardID() int {
	return p.persistence.GetShardID()
}

func (p *workflowExecutionPersistenceClient) CreateWorkflowExecution(
	ctx context.Context,
	request *CreateWorkflowExecutionRequest,
) (*CreateWorkflowExecutionResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceCreateWorkflowExecutionScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceCreateWorkflowExecutionScope, metrics.PersistenceLatency)
	response, err := p.persistence.CreateWorkflowExecution(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceCreateWorkflowExecutionScope, err)
	}

	return response, err
}

func (p *workflowExecutionPersistenceClient) GetWorkflowExecution(
	ctx context.Context,
	request *GetWorkflowExecutionRequest,
) (*GetWorkflowExecutionResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceGetWorkflowExecutionScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceGetWorkflowExecutionScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetWorkflowExecution(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceGetWorkflowExecutionScope, err)
	}

	return response, err
}

func (p *workflowExecutionPersistenceClient) UpdateWorkflowExecution(
	ctx context.Context,
	request *UpdateWorkflowExecutionRequest,
) (*UpdateWorkflowExecutionResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceUpdateWorkflowExecutionScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceUpdateWorkflowExecutionScope, metrics.PersistenceLatency)
	resp, err := p.persistence.UpdateWorkflowExecution(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceUpdateWorkflowExecutionScope, err)
	}

	return resp, err
}

func (p *workflowExecutionPersistenceClient) ConflictResolveWorkflowExecution(
	ctx context.Context,
	request *ConflictResolveWorkflowExecutionRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceConflictResolveWorkflowExecutionScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceConflictResolveWorkflowExecutionScope, metrics.PersistenceLatency)
	err := p.persistence.ConflictResolveWorkflowExecution(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceConflictResolveWorkflowExecutionScope, err)
	}

	return err
}

func (p *workflowExecutionPersistenceClient) ResetWorkflowExecution(
	ctx context.Context,
	request *ResetWorkflowExecutionRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceResetWorkflowExecutionScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceResetWorkflowExecutionScope, metrics.PersistenceLatency)
	err := p.persistence.ResetWorkflowExecution(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceResetWorkflowExecutionScope, err)
	}

	return err
}

func (p *workflowExecutionPersistenceClient) DeleteWorkflowExecution(
	ctx context.Context,
	request *DeleteWorkflowExecutionRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceDeleteWorkflowExecutionScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceDeleteWorkflowExecutionScope, metrics.PersistenceLatency)
	err := p.persistence.DeleteWorkflowExecution(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceDeleteWorkflowExecutionScope, err)
	}

	return err
}

func (p *workflowExecutionPersistenceClient) DeleteCurrentWorkflowExecution(
	ctx context.Context,
	request *DeleteCurrentWorkflowExecutionRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceDeleteCurrentWorkflowExecutionScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceDeleteCurrentWorkflowExecutionScope, metrics.PersistenceLatency)
	err := p.persistence.DeleteCurrentWorkflowExecution(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceDeleteCurrentWorkflowExecutionScope, err)
	}

	return err
}

func (p *workflowExecutionPersistenceClient) GetCurrentExecution(
	ctx context.Context,
	request *GetCurrentExecutionRequest,
) (*GetCurrentExecutionResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceGetCurrentExecutionScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceGetCurrentExecutionScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetCurrentExecution(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceGetCurrentExecutionScope, err)
	}

	return response, err
}

func (p *workflowExecutionPersistenceClient) ListCurrentExecutions(
	ctx context.Context,
	request *ListCurrentExecutionsRequest,
) (*ListCurrentExecutionsResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceListCurrentExecutionsScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceListCurrentExecutionsScope, metrics.PersistenceLatency)
	response, err := p.persistence.ListCurrentExecutions(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceListCurrentExecutionsScope, err)
	}

	return response, err
}

func (p *workflowExecutionPersistenceClient) IsWorkflowExecutionExists(
	ctx context.Context,
	request *IsWorkflowExecutionExistsRequest,
) (*IsWorkflowExecutionExistsResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceIsWorkflowExecutionExistsScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceIsWorkflowExecutionExistsScope, metrics.PersistenceLatency)
	response, err := p.persistence.IsWorkflowExecutionExists(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceIsWorkflowExecutionExistsScope, err)
	}

	return response, err
}

func (p *workflowExecutionPersistenceClient) ListConcreteExecutions(
	ctx context.Context,
	request *ListConcreteExecutionsRequest,
) (*ListConcreteExecutionsResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceListConcreteExecutionsScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceListConcreteExecutionsScope, metrics.PersistenceLatency)
	response, err := p.persistence.ListConcreteExecutions(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceListConcreteExecutionsScope, err)
	}

	return response, err
}

func (p *workflowExecutionPersistenceClient) GetTransferTasks(
	ctx context.Context,
	request *GetTransferTasksRequest,
) (*GetTransferTasksResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceGetTransferTasksScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceGetTransferTasksScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetTransferTasks(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceGetTransferTasksScope, err)
	}

	return response, err
}

func (p *workflowExecutionPersistenceClient) GetReplicationTasks(
	ctx context.Context,
	request *GetReplicationTasksRequest,
) (*GetReplicationTasksResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceGetReplicationTasksScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceGetReplicationTasksScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetReplicationTasks(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceGetReplicationTasksScope, err)
	}

	return response, err
}

func (p *workflowExecutionPersistenceClient) CompleteTransferTask(
	ctx context.Context,
	request *CompleteTransferTaskRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceCompleteTransferTaskScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceCompleteTransferTaskScope, metrics.PersistenceLatency)
	err := p.persistence.CompleteTransferTask(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceCompleteTransferTaskScope, err)
	}

	return err
}

func (p *workflowExecutionPersistenceClient) RangeCompleteTransferTask(
	ctx context.Context,
	request *RangeCompleteTransferTaskRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceRangeCompleteTransferTaskScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceRangeCompleteTransferTaskScope, metrics.PersistenceLatency)
	err := p.persistence.RangeCompleteTransferTask(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceRangeCompleteTransferTaskScope, err)
	}

	return err
}

func (p *workflowExecutionPersistenceClient) CompleteReplicationTask(
	ctx context.Context,
	request *CompleteReplicationTaskRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceCompleteReplicationTaskScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceCompleteReplicationTaskScope, metrics.PersistenceLatency)
	err := p.persistence.CompleteReplicationTask(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceCompleteReplicationTaskScope, err)
	}

	return err
}

func (p *workflowExecutionPersistenceClient) RangeCompleteReplicationTask(
	ctx context.Context,
	request *RangeCompleteReplicationTaskRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceRangeCompleteReplicationTaskScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceRangeCompleteReplicationTaskScope, metrics.PersistenceLatency)
	err := p.persistence.RangeCompleteReplicationTask(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceRangeCompleteReplicationTaskScope, err)
	}

	return err
}

func (p *workflowExecutionPersistenceClient) PutReplicationTaskToDLQ(
	ctx context.Context,
	request *PutReplicationTaskToDLQRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistencePutReplicationTaskToDLQScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistencePutReplicationTaskToDLQScope, metrics.PersistenceLatency)
	err := p.persistence.PutReplicationTaskToDLQ(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistencePutReplicationTaskToDLQScope, err)
	}

	return err
}

func (p *workflowExecutionPersistenceClient) GetReplicationTasksFromDLQ(
	ctx context.Context,
	request *GetReplicationTasksFromDLQRequest,
) (*GetReplicationTasksFromDLQResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceGetReplicationTasksFromDLQScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceGetReplicationTasksFromDLQScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetReplicationTasksFromDLQ(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceGetReplicationTasksFromDLQScope, err)
	}

	return response, err
}

func (p *workflowExecutionPersistenceClient) GetReplicationDLQSize(
	ctx context.Context,
	request *GetReplicationDLQSizeRequest,
) (*GetReplicationDLQSizeResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceGetReplicationDLQSizeScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceGetReplicationDLQSizeScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetReplicationDLQSize(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceGetReplicationDLQSizeScope, err)
	}

	return response, err
}

func (p *workflowExecutionPersistenceClient) DeleteReplicationTaskFromDLQ(
	ctx context.Context,
	request *DeleteReplicationTaskFromDLQRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceDeleteReplicationTaskFromDLQScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceDeleteReplicationTaskFromDLQScope, metrics.PersistenceLatency)
	err := p.persistence.DeleteReplicationTaskFromDLQ(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceDeleteReplicationTaskFromDLQScope, err)
	}

	return nil
}

func (p *workflowExecutionPersistenceClient) RangeDeleteReplicationTaskFromDLQ(
	ctx context.Context,
	request *RangeDeleteReplicationTaskFromDLQRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceRangeDeleteReplicationTaskFromDLQScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceRangeDeleteReplicationTaskFromDLQScope, metrics.PersistenceLatency)
	err := p.persistence.RangeDeleteReplicationTaskFromDLQ(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceRangeDeleteReplicationTaskFromDLQScope, err)
	}

	return nil
}

func (p *workflowExecutionPersistenceClient) CreateFailoverMarkerTasks(
	ctx context.Context,
	request *CreateFailoverMarkersRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceCreateFailoverMarkerTasksScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceCreateFailoverMarkerTasksScope, metrics.PersistenceLatency)
	err := p.persistence.CreateFailoverMarkerTasks(ctx, request)
	sw.Stop()

	if err != nil {
		p.metricClient.IncCounter(metrics.PersistenceCreateFailoverMarkerTasksScope, metrics.PersistenceFailures)
	}

	return err
}

func (p *workflowExecutionPersistenceClient) GetTimerIndexTasks(
	ctx context.Context,
	request *GetTimerIndexTasksRequest,
) (*GetTimerIndexTasksResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceGetTimerIndexTasksScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceGetTimerIndexTasksScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetTimerIndexTasks(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceGetTimerIndexTasksScope, err)
	}

	return response, err
}

func (p *workflowExecutionPersistenceClient) CompleteTimerTask(
	ctx context.Context,
	request *CompleteTimerTaskRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceCompleteTimerTaskScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceCompleteTimerTaskScope, metrics.PersistenceLatency)
	err := p.persistence.CompleteTimerTask(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceCompleteTimerTaskScope, err)
	}

	return err
}

func (p *workflowExecutionPersistenceClient) RangeCompleteTimerTask(
	ctx context.Context,
	request *RangeCompleteTimerTaskRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceRangeCompleteTimerTaskScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceRangeCompleteTimerTaskScope, metrics.PersistenceLatency)
	err := p.persistence.RangeCompleteTimerTask(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceRangeCompleteTimerTaskScope, err)
	}

	return err
}

func (p *workflowExecutionPersistenceClient) updateErrorMetric(scope int, err error) {
	switch err.(type) {
	case *WorkflowExecutionAlreadyStartedError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrExecutionAlreadyStartedCounter)
	case *types.EntityNotExistsError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrEntityNotExistsCounter)
	case *ShardOwnershipLostError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrShardOwnershipLostCounter)
	case *ConditionFailedError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrConditionFailedCounter)
	case *CurrentWorkflowConditionFailedError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrCurrentWorkflowConditionFailedCounter)
	case *TimeoutError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrTimeoutCounter)
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	case *types.ServiceBusyError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrBusyCounter)
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	default:
		p.logger.Error("Operation failed with internal error.",
			tag.Error(err), tag.MetricScope(scope), tag.ShardID(p.GetShardID()))
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	}
}

func (p *workflowExecutionPersistenceClient) Close() {
	p.persistence.Close()
}

func (p *taskPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *taskPersistenceClient) CreateTasks(
	ctx context.Context,
	request *CreateTasksRequest,
) (*CreateTasksResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceCreateTaskScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceCreateTaskScope, metrics.PersistenceLatency)
	response, err := p.persistence.CreateTasks(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceCreateTaskScope, err)
	}

	return response, err
}

func (p *taskPersistenceClient) GetTasks(
	ctx context.Context,
	request *GetTasksRequest,
) (*GetTasksResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceGetTasksScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceGetTasksScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetTasks(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceGetTasksScope, err)
	}

	return response, err
}

func (p *taskPersistenceClient) CompleteTask(
	ctx context.Context,
	request *CompleteTaskRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceCompleteTaskScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceCompleteTaskScope, metrics.PersistenceLatency)
	err := p.persistence.CompleteTask(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceCompleteTaskScope, err)
	}

	return err
}

func (p *taskPersistenceClient) CompleteTasksLessThan(
	ctx context.Context,
	request *CompleteTasksLessThanRequest,
) (int, error) {
	p.metricClient.IncCounter(metrics.PersistenceCompleteTasksLessThanScope, metrics.PersistenceRequests)
	sw := p.metricClient.StartTimer(metrics.PersistenceCompleteTasksLessThanScope, metrics.PersistenceLatency)
	result, err := p.persistence.CompleteTasksLessThan(ctx, request)
	sw.Stop()
	if err != nil {
		p.updateErrorMetric(metrics.PersistenceCompleteTasksLessThanScope, err)
	}
	return result, err
}

func (p *taskPersistenceClient) LeaseTaskList(
	ctx context.Context,
	request *LeaseTaskListRequest,
) (*LeaseTaskListResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceLeaseTaskListScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceLeaseTaskListScope, metrics.PersistenceLatency)
	response, err := p.persistence.LeaseTaskList(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceLeaseTaskListScope, err)
	}

	return response, err
}

func (p *taskPersistenceClient) ListTaskList(
	ctx context.Context,
	request *ListTaskListRequest,
) (*ListTaskListResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceListTaskListScope, metrics.PersistenceRequests)
	sw := p.metricClient.StartTimer(metrics.PersistenceListTaskListScope, metrics.PersistenceLatency)
	response, err := p.persistence.ListTaskList(ctx, request)
	sw.Stop()
	if err != nil {
		p.updateErrorMetric(metrics.PersistenceListTaskListScope, err)
	}
	return response, err
}

func (p *taskPersistenceClient) DeleteTaskList(
	ctx context.Context,
	request *DeleteTaskListRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceDeleteTaskListScope, metrics.PersistenceRequests)
	sw := p.metricClient.StartTimer(metrics.PersistenceDeleteTaskListScope, metrics.PersistenceLatency)
	err := p.persistence.DeleteTaskList(ctx, request)
	sw.Stop()
	if err != nil {
		p.updateErrorMetric(metrics.PersistenceDeleteTaskListScope, err)
	}
	return err
}

func (p *taskPersistenceClient) UpdateTaskList(
	ctx context.Context,
	request *UpdateTaskListRequest,
) (*UpdateTaskListResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceUpdateTaskListScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceUpdateTaskListScope, metrics.PersistenceLatency)
	response, err := p.persistence.UpdateTaskList(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceUpdateTaskListScope, err)
	}

	return response, err
}

func (p *taskPersistenceClient) updateErrorMetric(scope int, err error) {
	switch err.(type) {
	case *ConditionFailedError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrConditionFailedCounter)
	case *TimeoutError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrTimeoutCounter)
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	case *types.ServiceBusyError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrBusyCounter)
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	default:
		p.logger.Error("Operation failed with internal error.",
			tag.Error(err), tag.MetricScope(scope))
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	}
}

func (p *taskPersistenceClient) Close() {
	p.persistence.Close()
}

func (p *metadataPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *metadataPersistenceClient) CreateDomain(
	ctx context.Context,
	request *CreateDomainRequest,
) (*CreateDomainResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceCreateDomainScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceCreateDomainScope, metrics.PersistenceLatency)
	response, err := p.persistence.CreateDomain(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceCreateDomainScope, err)
	}

	return response, err
}

func (p *metadataPersistenceClient) GetDomain(
	ctx context.Context,
	request *GetDomainRequest,
) (*GetDomainResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceGetDomainScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceGetDomainScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetDomain(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceGetDomainScope, err)
	}

	return response, err
}

func (p *metadataPersistenceClient) UpdateDomain(
	ctx context.Context,
	request *UpdateDomainRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceUpdateDomainScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceUpdateDomainScope, metrics.PersistenceLatency)
	err := p.persistence.UpdateDomain(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceUpdateDomainScope, err)
	}

	return err
}

func (p *metadataPersistenceClient) DeleteDomain(
	ctx context.Context,
	request *DeleteDomainRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceDeleteDomainScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceDeleteDomainScope, metrics.PersistenceLatency)
	err := p.persistence.DeleteDomain(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceDeleteDomainScope, err)
	}

	return err
}

func (p *metadataPersistenceClient) DeleteDomainByName(
	ctx context.Context,
	request *DeleteDomainByNameRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceDeleteDomainByNameScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceDeleteDomainByNameScope, metrics.PersistenceLatency)
	err := p.persistence.DeleteDomainByName(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceDeleteDomainByNameScope, err)
	}

	return err
}

func (p *metadataPersistenceClient) ListDomains(
	ctx context.Context,
	request *ListDomainsRequest,
) (*ListDomainsResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceListDomainScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceListDomainScope, metrics.PersistenceLatency)
	response, err := p.persistence.ListDomains(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceListDomainScope, err)
	}

	return response, err
}

func (p *metadataPersistenceClient) GetMetadata(
	ctx context.Context,
) (*GetMetadataResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceGetMetadataScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceGetMetadataScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetMetadata(ctx)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceGetMetadataScope, err)
	}

	return response, err
}

func (p *metadataPersistenceClient) Close() {
	p.persistence.Close()
}

func (p *metadataPersistenceClient) updateErrorMetric(scope int, err error) {
	switch err.(type) {
	case *types.DomainAlreadyExistsError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrDomainAlreadyExistsCounter)
	case *types.EntityNotExistsError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrEntityNotExistsCounter)
	case *types.BadRequestError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrBadRequestCounter)
	case *types.ServiceBusyError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrBusyCounter)
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	default:
		p.logger.Error("Operation failed with internal error.",
			tag.Error(err), tag.MetricScope(scope))
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	}
}

func (p *visibilityPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *visibilityPersistenceClient) RecordWorkflowExecutionStarted(
	ctx context.Context,
	request *RecordWorkflowExecutionStartedRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceRecordWorkflowExecutionStartedScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceRecordWorkflowExecutionStartedScope, metrics.PersistenceLatency)
	err := p.persistence.RecordWorkflowExecutionStarted(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceRecordWorkflowExecutionStartedScope, err)
	}

	return err
}

func (p *visibilityPersistenceClient) RecordWorkflowExecutionClosed(
	ctx context.Context,
	request *RecordWorkflowExecutionClosedRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceRecordWorkflowExecutionClosedScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceRecordWorkflowExecutionClosedScope, metrics.PersistenceLatency)
	err := p.persistence.RecordWorkflowExecutionClosed(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceRecordWorkflowExecutionClosedScope, err)
	}

	return err
}

func (p *visibilityPersistenceClient) UpsertWorkflowExecution(
	ctx context.Context,
	request *UpsertWorkflowExecutionRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceUpsertWorkflowExecutionScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceUpsertWorkflowExecutionScope, metrics.PersistenceLatency)
	err := p.persistence.UpsertWorkflowExecution(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceUpsertWorkflowExecutionScope, err)
	}

	return err
}

func (p *visibilityPersistenceClient) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsRequest,
) (*ListWorkflowExecutionsResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceListOpenWorkflowExecutionsScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceListOpenWorkflowExecutionsScope, metrics.PersistenceLatency)
	response, err := p.persistence.ListOpenWorkflowExecutions(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceListOpenWorkflowExecutionsScope, err)
	}

	return response, err
}

func (p *visibilityPersistenceClient) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsRequest,
) (*ListWorkflowExecutionsResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceListClosedWorkflowExecutionsScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceListClosedWorkflowExecutionsScope, metrics.PersistenceLatency)
	response, err := p.persistence.ListClosedWorkflowExecutions(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceListClosedWorkflowExecutionsScope, err)
	}

	return response, err
}

func (p *visibilityPersistenceClient) ListOpenWorkflowExecutionsByType(
	ctx context.Context,
	request *ListWorkflowExecutionsByTypeRequest,
) (*ListWorkflowExecutionsResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceListOpenWorkflowExecutionsByTypeScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceListOpenWorkflowExecutionsByTypeScope, metrics.PersistenceLatency)
	response, err := p.persistence.ListOpenWorkflowExecutionsByType(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceListOpenWorkflowExecutionsByTypeScope, err)
	}

	return response, err
}

func (p *visibilityPersistenceClient) ListClosedWorkflowExecutionsByType(
	ctx context.Context,
	request *ListWorkflowExecutionsByTypeRequest,
) (*ListWorkflowExecutionsResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceListClosedWorkflowExecutionsByTypeScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceListClosedWorkflowExecutionsByTypeScope, metrics.PersistenceLatency)
	response, err := p.persistence.ListClosedWorkflowExecutionsByType(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceListClosedWorkflowExecutionsByTypeScope, err)
	}

	return response, err
}

func (p *visibilityPersistenceClient) ListOpenWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *ListWorkflowExecutionsByWorkflowIDRequest,
) (*ListWorkflowExecutionsResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceListOpenWorkflowExecutionsByWorkflowIDScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceListOpenWorkflowExecutionsByWorkflowIDScope, metrics.PersistenceLatency)
	response, err := p.persistence.ListOpenWorkflowExecutionsByWorkflowID(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceListOpenWorkflowExecutionsByWorkflowIDScope, err)
	}

	return response, err
}

func (p *visibilityPersistenceClient) ListClosedWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *ListWorkflowExecutionsByWorkflowIDRequest,
) (*ListWorkflowExecutionsResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceListClosedWorkflowExecutionsByWorkflowIDScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceListClosedWorkflowExecutionsByWorkflowIDScope, metrics.PersistenceLatency)
	response, err := p.persistence.ListClosedWorkflowExecutionsByWorkflowID(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceListClosedWorkflowExecutionsByWorkflowIDScope, err)
	}

	return response, err
}

func (p *visibilityPersistenceClient) ListClosedWorkflowExecutionsByStatus(
	ctx context.Context,
	request *ListClosedWorkflowExecutionsByStatusRequest,
) (*ListWorkflowExecutionsResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceListClosedWorkflowExecutionsByStatusScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceListClosedWorkflowExecutionsByStatusScope, metrics.PersistenceLatency)
	response, err := p.persistence.ListClosedWorkflowExecutionsByStatus(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceListClosedWorkflowExecutionsByStatusScope, err)
	}

	return response, err
}

func (p *visibilityPersistenceClient) GetClosedWorkflowExecution(
	ctx context.Context,
	request *GetClosedWorkflowExecutionRequest,
) (*GetClosedWorkflowExecutionResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceGetClosedWorkflowExecutionScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceGetClosedWorkflowExecutionScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetClosedWorkflowExecution(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceGetClosedWorkflowExecutionScope, err)
	}

	return response, err
}

func (p *visibilityPersistenceClient) DeleteWorkflowExecution(
	ctx context.Context,
	request *VisibilityDeleteWorkflowExecutionRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceVisibilityDeleteWorkflowExecutionScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceVisibilityDeleteWorkflowExecutionScope, metrics.PersistenceLatency)
	err := p.persistence.DeleteWorkflowExecution(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceVisibilityDeleteWorkflowExecutionScope, err)
	}

	return err
}

func (p *visibilityPersistenceClient) ListWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsByQueryRequest,
) (*ListWorkflowExecutionsResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceListWorkflowExecutionsScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceListWorkflowExecutionsScope, metrics.PersistenceLatency)
	response, err := p.persistence.ListWorkflowExecutions(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceListWorkflowExecutionsScope, err)
	}

	return response, err
}

func (p *visibilityPersistenceClient) ScanWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsByQueryRequest,
) (*ListWorkflowExecutionsResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceScanWorkflowExecutionsScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceScanWorkflowExecutionsScope, metrics.PersistenceLatency)
	response, err := p.persistence.ScanWorkflowExecutions(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceScanWorkflowExecutionsScope, err)
	}

	return response, err
}

func (p *visibilityPersistenceClient) CountWorkflowExecutions(
	ctx context.Context,
	request *CountWorkflowExecutionsRequest,
) (*CountWorkflowExecutionsResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceCountWorkflowExecutionsScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceCountWorkflowExecutionsScope, metrics.PersistenceLatency)
	response, err := p.persistence.CountWorkflowExecutions(ctx, request)
	sw.Stop()

	if err != nil {
		p.updateErrorMetric(metrics.PersistenceCountWorkflowExecutionsScope, err)
	}

	return response, err
}

func (p *visibilityPersistenceClient) updateErrorMetric(scope int, err error) {
	switch err.(type) {
	case *ConditionFailedError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrConditionFailedCounter)
	case *TimeoutError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrTimeoutCounter)
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	case *types.EntityNotExistsError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrEntityNotExistsCounter)
	case *types.ServiceBusyError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrBusyCounter)
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	default:
		p.logger.Error("Operation failed with internal error.",
			tag.Error(err), tag.MetricScope(scope))
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	}
}

func (p *visibilityPersistenceClient) Close() {
	p.persistence.Close()
}

func (p *historyPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *historyPersistenceClient) Close() {
	p.persistence.Close()
}

// AppendHistoryNodes add(or override) a node to a history branch
func (p *historyPersistenceClient) AppendHistoryNodes(
	ctx context.Context,
	request *AppendHistoryNodesRequest,
) (*AppendHistoryNodesResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceAppendHistoryNodesScope, metrics.PersistenceRequests)
	sw := p.metricClient.StartTimer(metrics.PersistenceAppendHistoryNodesScope, metrics.PersistenceLatency)
	resp, err := p.persistence.AppendHistoryNodes(ctx, request)
	sw.Stop()
	if err != nil {
		p.updateErrorMetric(metrics.PersistenceAppendHistoryNodesScope, err)
	}
	return resp, err
}

// ReadHistoryBranch returns history node data for a branch
func (p *historyPersistenceClient) ReadHistoryBranch(
	ctx context.Context,
	request *ReadHistoryBranchRequest,
) (*ReadHistoryBranchResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceReadHistoryBranchScope, metrics.PersistenceRequests)
	sw := p.metricClient.StartTimer(metrics.PersistenceReadHistoryBranchScope, metrics.PersistenceLatency)
	response, err := p.persistence.ReadHistoryBranch(ctx, request)
	sw.Stop()
	if err != nil {
		p.updateErrorMetric(metrics.PersistenceReadHistoryBranchScope, err)
	}
	return response, err
}

// ReadHistoryBranchByBatch returns history node data for a branch ByBatch
func (p *historyPersistenceClient) ReadHistoryBranchByBatch(
	ctx context.Context,
	request *ReadHistoryBranchRequest,
) (*ReadHistoryBranchByBatchResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceReadHistoryBranchScope, metrics.PersistenceRequests)
	sw := p.metricClient.StartTimer(metrics.PersistenceReadHistoryBranchScope, metrics.PersistenceLatency)
	response, err := p.persistence.ReadHistoryBranchByBatch(ctx, request)
	sw.Stop()
	if err != nil {
		p.updateErrorMetric(metrics.PersistenceReadHistoryBranchScope, err)
	}
	return response, err
}

// ReadRawHistoryBranch returns history node raw data for a branch ByBatch
func (p *historyPersistenceClient) ReadRawHistoryBranch(
	ctx context.Context,
	request *ReadHistoryBranchRequest,
) (*ReadRawHistoryBranchResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceReadHistoryBranchScope, metrics.PersistenceRequests)
	sw := p.metricClient.StartTimer(metrics.PersistenceReadHistoryBranchScope, metrics.PersistenceLatency)
	response, err := p.persistence.ReadRawHistoryBranch(ctx, request)
	sw.Stop()
	if err != nil {
		p.updateErrorMetric(metrics.PersistenceReadHistoryBranchScope, err)
	}
	return response, err
}

// ForkHistoryBranch forks a new branch from a old branch
func (p *historyPersistenceClient) ForkHistoryBranch(
	ctx context.Context,
	request *ForkHistoryBranchRequest,
) (*ForkHistoryBranchResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceForkHistoryBranchScope, metrics.PersistenceRequests)
	sw := p.metricClient.StartTimer(metrics.PersistenceForkHistoryBranchScope, metrics.PersistenceLatency)
	response, err := p.persistence.ForkHistoryBranch(ctx, request)
	sw.Stop()
	if err != nil {
		p.updateErrorMetric(metrics.PersistenceForkHistoryBranchScope, err)
	}
	return response, err
}

// DeleteHistoryBranch removes a branch
func (p *historyPersistenceClient) DeleteHistoryBranch(
	ctx context.Context,
	request *DeleteHistoryBranchRequest,
) error {
	p.metricClient.IncCounter(metrics.PersistenceDeleteHistoryBranchScope, metrics.PersistenceRequests)
	sw := p.metricClient.StartTimer(metrics.PersistenceDeleteHistoryBranchScope, metrics.PersistenceLatency)
	err := p.persistence.DeleteHistoryBranch(ctx, request)
	sw.Stop()
	if err != nil {
		p.updateErrorMetric(metrics.PersistenceDeleteHistoryBranchScope, err)
	}
	return err
}

func (p *historyPersistenceClient) GetAllHistoryTreeBranches(
	ctx context.Context,
	request *GetAllHistoryTreeBranchesRequest,
) (*GetAllHistoryTreeBranchesResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceGetAllHistoryTreeBranchesScope, metrics.PersistenceRequests)
	sw := p.metricClient.StartTimer(metrics.PersistenceGetAllHistoryTreeBranchesScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetAllHistoryTreeBranches(ctx, request)
	sw.Stop()
	if err != nil {
		p.updateErrorMetric(metrics.PersistenceGetAllHistoryTreeBranchesScope, err)
	}
	return response, err
}

// GetHistoryTree returns all branch information of a tree
func (p *historyPersistenceClient) GetHistoryTree(
	ctx context.Context,
	request *GetHistoryTreeRequest,
) (*GetHistoryTreeResponse, error) {
	p.metricClient.IncCounter(metrics.PersistenceGetHistoryTreeScope, metrics.PersistenceRequests)
	sw := p.metricClient.StartTimer(metrics.PersistenceGetHistoryTreeScope, metrics.PersistenceLatency)
	response, err := p.persistence.GetHistoryTree(ctx, request)
	sw.Stop()
	if err != nil {
		p.updateErrorMetric(metrics.PersistenceGetHistoryTreeScope, err)
	}
	return response, err
}

func (p *historyPersistenceClient) updateErrorMetric(scope int, err error) {
	switch err.(type) {
	case *types.EntityNotExistsError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrEntityNotExistsCounter)
	case *ConditionFailedError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrConditionFailedCounter)
	case *TimeoutError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrTimeoutCounter)
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	case *types.ServiceBusyError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrBusyCounter)
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	default:
		p.logger.Error("Operation failed with internal error.",
			tag.Error(err), tag.MetricScope(scope))
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	}
}

func (p *queuePersistenceClient) EnqueueMessage(
	ctx context.Context,
	message []byte,
) error {
	p.metricClient.IncCounter(metrics.PersistenceEnqueueMessageScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceEnqueueMessageScope, metrics.PersistenceLatency)
	err := p.persistence.EnqueueMessage(ctx, message)
	sw.Stop()

	if err != nil {
		p.metricClient.IncCounter(metrics.PersistenceEnqueueMessageScope, metrics.PersistenceFailures)
	}

	return err
}

func (p *queuePersistenceClient) ReadMessages(
	ctx context.Context,
	lastMessageID int64,
	maxCount int,
) ([]*QueueMessage, error) {
	p.metricClient.IncCounter(metrics.PersistenceReadQueueMessagesScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceReadQueueMessagesScope, metrics.PersistenceLatency)
	result, err := p.persistence.ReadMessages(ctx, lastMessageID, maxCount)
	sw.Stop()

	if err != nil {
		p.metricClient.IncCounter(metrics.PersistenceReadQueueMessagesScope, metrics.PersistenceFailures)
	}

	return result, err
}

func (p *queuePersistenceClient) UpdateAckLevel(
	ctx context.Context,
	messageID int64,
	clusterName string,
) error {
	p.metricClient.IncCounter(metrics.PersistenceUpdateAckLevelScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceUpdateAckLevelScope, metrics.PersistenceLatency)
	err := p.persistence.UpdateAckLevel(ctx, messageID, clusterName)
	sw.Stop()

	if err != nil {
		p.metricClient.IncCounter(metrics.PersistenceUpdateAckLevelScope, metrics.PersistenceFailures)
	}

	return err
}

func (p *queuePersistenceClient) GetAckLevels(
	ctx context.Context,
) (map[string]int64, error) {
	p.metricClient.IncCounter(metrics.PersistenceGetAckLevelScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceGetAckLevelScope, metrics.PersistenceLatency)
	result, err := p.persistence.GetAckLevels(ctx)
	sw.Stop()

	if err != nil {
		p.metricClient.IncCounter(metrics.PersistenceGetAckLevelScope, metrics.PersistenceFailures)
	}

	return result, err
}

func (p *queuePersistenceClient) DeleteMessagesBefore(
	ctx context.Context,
	messageID int64,
) error {
	p.metricClient.IncCounter(metrics.PersistenceDeleteQueueMessagesScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceDeleteQueueMessagesScope, metrics.PersistenceLatency)
	err := p.persistence.DeleteMessagesBefore(ctx, messageID)
	sw.Stop()

	if err != nil {
		p.metricClient.IncCounter(metrics.PersistenceDeleteQueueMessagesScope, metrics.PersistenceFailures)
	}

	return err
}

func (p *queuePersistenceClient) EnqueueMessageToDLQ(
	ctx context.Context,
	message []byte,
) error {
	p.metricClient.IncCounter(metrics.PersistenceEnqueueMessageToDLQScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceEnqueueMessageToDLQScope, metrics.PersistenceLatency)
	err := p.persistence.EnqueueMessageToDLQ(ctx, message)
	sw.Stop()

	if err != nil {
		p.metricClient.IncCounter(metrics.PersistenceEnqueueMessageToDLQScope, metrics.PersistenceFailures)
	}

	return err
}

func (p *queuePersistenceClient) ReadMessagesFromDLQ(
	ctx context.Context,
	firstMessageID int64,
	lastMessageID int64,
	pageSize int,
	pageToken []byte,
) ([]*QueueMessage, []byte, error) {
	p.metricClient.IncCounter(metrics.PersistenceReadQueueMessagesFromDLQScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceReadQueueMessagesFromDLQScope, metrics.PersistenceLatency)
	result, token, err := p.persistence.ReadMessagesFromDLQ(ctx, firstMessageID, lastMessageID, pageSize, pageToken)
	sw.Stop()

	if err != nil {
		p.metricClient.IncCounter(metrics.PersistenceReadQueueMessagesFromDLQScope, metrics.PersistenceFailures)
	}

	return result, token, err
}

func (p *queuePersistenceClient) DeleteMessageFromDLQ(
	ctx context.Context,
	messageID int64,
) error {
	p.metricClient.IncCounter(metrics.PersistenceDeleteQueueMessageFromDLQScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceDeleteQueueMessageFromDLQScope, metrics.PersistenceLatency)
	err := p.persistence.DeleteMessageFromDLQ(ctx, messageID)
	sw.Stop()

	if err != nil {
		p.metricClient.IncCounter(metrics.PersistenceDeleteQueueMessageFromDLQScope, metrics.PersistenceFailures)
	}

	return err
}

func (p *queuePersistenceClient) RangeDeleteMessagesFromDLQ(
	ctx context.Context,
	firstMessageID int64,
	lastMessageID int64,
) error {
	p.metricClient.IncCounter(metrics.PersistenceRangeDeleteMessagesFromDLQScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceRangeDeleteMessagesFromDLQScope, metrics.PersistenceLatency)
	err := p.persistence.RangeDeleteMessagesFromDLQ(ctx, firstMessageID, lastMessageID)
	sw.Stop()

	if err != nil {
		p.metricClient.IncCounter(metrics.PersistenceRangeDeleteMessagesFromDLQScope, metrics.PersistenceFailures)
	}

	return err
}

func (p *queuePersistenceClient) UpdateDLQAckLevel(
	ctx context.Context,
	messageID int64,
	clusterName string,
) error {
	p.metricClient.IncCounter(metrics.PersistenceUpdateDLQAckLevelScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceUpdateDLQAckLevelScope, metrics.PersistenceLatency)
	err := p.persistence.UpdateDLQAckLevel(ctx, messageID, clusterName)
	sw.Stop()

	if err != nil {
		p.metricClient.IncCounter(metrics.PersistenceUpdateDLQAckLevelScope, metrics.PersistenceFailures)
	}

	return err
}

func (p *queuePersistenceClient) GetDLQAckLevels(
	ctx context.Context,
) (map[string]int64, error) {
	p.metricClient.IncCounter(metrics.PersistenceGetDLQAckLevelScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceGetDLQAckLevelScope, metrics.PersistenceLatency)
	result, err := p.persistence.GetDLQAckLevels(ctx)
	sw.Stop()

	if err != nil {
		p.metricClient.IncCounter(metrics.PersistenceGetDLQAckLevelScope, metrics.PersistenceFailures)
	}

	return result, err
}

func (p *queuePersistenceClient) GetDLQSize(
	ctx context.Context,
) (int64, error) {
	p.metricClient.IncCounter(metrics.PersistenceGetDLQSizeScope, metrics.PersistenceRequests)

	sw := p.metricClient.StartTimer(metrics.PersistenceGetDLQSizeScope, metrics.PersistenceLatency)
	result, err := p.persistence.GetDLQSize(ctx)
	sw.Stop()

	if err != nil {
		p.metricClient.IncCounter(metrics.PersistenceGetDLQSizeScope, metrics.PersistenceFailures)
	}

	return result, err
}

func (p *queuePersistenceClient) Close() {
	p.persistence.Close()
}
