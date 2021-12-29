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
	"fmt"
	"time"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
)

type (
	persistenceMetricsClientBase struct {
		metricClient                  metrics.Client
		logger                        log.Logger
		throttledLogger               log.Logger
		enableLatencyHistogramMetrics bool
		maxExpectedLatency            dynamicconfig.DurationPropertyFnWithOperationFilter
	}

	shardPersistenceClient struct {
		persistenceMetricsClientBase
		persistence ShardManager
	}

	workflowExecutionPersistenceClient struct {
		persistenceMetricsClientBase
		persistence ExecutionManager
	}

	taskPersistenceClient struct {
		persistenceMetricsClientBase
		persistence TaskManager
	}

	historyPersistenceClient struct {
		persistenceMetricsClientBase
		persistence HistoryManager
	}

	metadataPersistenceClient struct {
		persistenceMetricsClientBase
		persistence DomainManager
	}

	visibilityPersistenceClient struct {
		persistenceMetricsClientBase
		persistence VisibilityManager
	}

	queuePersistenceClient struct {
		persistenceMetricsClientBase
		persistence QueueManager
	}

	configStorePersistenceClient struct {
		persistenceMetricsClientBase
		persistence ConfigStoreManager
	}
)

var _ ShardManager = (*shardPersistenceClient)(nil)
var _ ExecutionManager = (*workflowExecutionPersistenceClient)(nil)
var _ TaskManager = (*taskPersistenceClient)(nil)
var _ HistoryManager = (*historyPersistenceClient)(nil)
var _ DomainManager = (*metadataPersistenceClient)(nil)
var _ VisibilityManager = (*visibilityPersistenceClient)(nil)
var _ QueueManager = (*queuePersistenceClient)(nil)
var _ ConfigStoreManager = (*configStorePersistenceClient)(nil)

var emptyTags = func() []tag.Tag {
	return []tag.Tag{}
}

func newPersistenceMetricsClientBase(
	metricClient metrics.Client,
	logger log.Logger,
	throttledLogger log.Logger,
	enableLatencyHistogramMetrics bool,
	maxExpectedLatency dynamicconfig.DurationPropertyFnWithOperationFilter,
) persistenceMetricsClientBase {
	return persistenceMetricsClientBase{
		metricClient:                  metricClient,
		logger:                        logger,
		throttledLogger:               throttledLogger.WithTags(tag.ComponentPersistence),
		enableLatencyHistogramMetrics: enableLatencyHistogramMetrics,
		maxExpectedLatency:            maxExpectedLatency,
	}
}

// NewShardPersistenceMetricsClient creates a client to manage shards
func NewShardPersistenceMetricsClient(
	persistence ShardManager,
	metricClient metrics.Client,
	logger log.Logger,
	throttledLogger log.Logger,
	cfg *config.Persistence,
	maxExpectedLatency dynamicconfig.DurationPropertyFnWithOperationFilter,
) ShardManager {
	return &shardPersistenceClient{
		persistence: persistence,
		persistenceMetricsClientBase: newPersistenceMetricsClientBase(
			metricClient,
			logger,
			throttledLogger,
			cfg.EnablePersistenceLatencyHistogramMetrics,
			maxExpectedLatency,
		),
	}
}

// NewWorkflowExecutionPersistenceMetricsClient creates a client to manage executions
func NewWorkflowExecutionPersistenceMetricsClient(
	persistence ExecutionManager,
	metricClient metrics.Client,
	logger log.Logger,
	throttledLogger log.Logger,
	cfg *config.Persistence,
	maxExpectedLatency dynamicconfig.DurationPropertyFnWithOperationFilter,
) ExecutionManager {
	return &workflowExecutionPersistenceClient{
		persistence: persistence,
		persistenceMetricsClientBase: newPersistenceMetricsClientBase(
			metricClient,
			logger.WithTags(tag.ShardID(persistence.GetShardID())),
			throttledLogger,
			cfg.EnablePersistenceLatencyHistogramMetrics,
			maxExpectedLatency,
		),
	}
}

// NewTaskPersistenceMetricsClient creates a client to manage tasks
func NewTaskPersistenceMetricsClient(
	persistence TaskManager,
	metricClient metrics.Client,
	logger log.Logger,
	throttledLogger log.Logger,
	cfg *config.Persistence,
	maxExpectedLatency dynamicconfig.DurationPropertyFnWithOperationFilter,
) TaskManager {
	return &taskPersistenceClient{
		persistence: persistence,
		persistenceMetricsClientBase: newPersistenceMetricsClientBase(
			metricClient,
			logger,
			throttledLogger,
			cfg.EnablePersistenceLatencyHistogramMetrics,
			maxExpectedLatency,
		),
	}
}

// NewHistoryPersistenceMetricsClient creates a HistoryManager client to manage workflow execution history
func NewHistoryPersistenceMetricsClient(
	persistence HistoryManager,
	metricClient metrics.Client,
	logger log.Logger,
	throttledLogger log.Logger,
	cfg *config.Persistence,
	maxExpectedLatency dynamicconfig.DurationPropertyFnWithOperationFilter,
) HistoryManager {
	return &historyPersistenceClient{
		persistence: persistence,
		persistenceMetricsClientBase: newPersistenceMetricsClientBase(
			metricClient,
			logger,
			throttledLogger,
			cfg.EnablePersistenceLatencyHistogramMetrics,
			maxExpectedLatency,
		),
	}
}

// NewDomainPersistenceMetricsClient creates a DomainManager client to manage metadata
func NewDomainPersistenceMetricsClient(
	persistence DomainManager,
	metricClient metrics.Client,
	logger log.Logger,
	throttledLogger log.Logger,
	cfg *config.Persistence,
	maxExpectedLatency dynamicconfig.DurationPropertyFnWithOperationFilter,
) DomainManager {
	return &metadataPersistenceClient{
		persistence: persistence,
		persistenceMetricsClientBase: newPersistenceMetricsClientBase(
			metricClient,
			logger,
			throttledLogger,
			cfg.EnablePersistenceLatencyHistogramMetrics,
			maxExpectedLatency,
		),
	}
}

// NewVisibilityPersistenceMetricsClient creates a client to manage visibility
func NewVisibilityPersistenceMetricsClient(
	persistence VisibilityManager,
	metricClient metrics.Client,
	logger log.Logger,
	throttledLogger log.Logger,
	cfg *config.Persistence,
	maxExpectedLatency dynamicconfig.DurationPropertyFnWithOperationFilter,
) VisibilityManager {
	return &visibilityPersistenceClient{
		persistence: persistence,
		persistenceMetricsClientBase: newPersistenceMetricsClientBase(
			metricClient,
			logger,
			throttledLogger,
			cfg.EnablePersistenceLatencyHistogramMetrics,
			maxExpectedLatency,
		),
	}
}

// NewQueuePersistenceMetricsClient creates a client to manage queue
func NewQueuePersistenceMetricsClient(
	persistence QueueManager,
	metricClient metrics.Client,
	logger log.Logger,
	throttledLogger log.Logger,
	cfg *config.Persistence,
	maxExpectedLatency dynamicconfig.DurationPropertyFnWithOperationFilter,
) QueueManager {
	return &queuePersistenceClient{
		persistence: persistence,
		persistenceMetricsClientBase: newPersistenceMetricsClientBase(
			metricClient,
			logger,
			throttledLogger,
			cfg.EnablePersistenceLatencyHistogramMetrics,
			maxExpectedLatency,
		),
	}
}

// NewConfigStorePersistenceMetricsClient creates a client to manage config store
func NewConfigStorePersistenceMetricsClient(
	persistence ConfigStoreManager,
	metricClient metrics.Client,
	logger log.Logger,
	throttledLogger log.Logger,
	cfg *config.Persistence,
	maxExpectedLatency dynamicconfig.DurationPropertyFnWithOperationFilter,
) ConfigStoreManager {
	return &configStorePersistenceClient{
		persistence: persistence,
		persistenceMetricsClientBase: newPersistenceMetricsClientBase(
			metricClient,
			logger,
			throttledLogger,
			cfg.EnablePersistenceLatencyHistogramMetrics,
			maxExpectedLatency,
		),
	}
}

func (p *persistenceMetricsClientBase) updateErrorMetric(scope int, err error) {
	switch err.(type) {
	case *types.DomainAlreadyExistsError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrDomainAlreadyExistsCounter)
	case *types.BadRequestError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrBadRequestCounter)
	case *WorkflowExecutionAlreadyStartedError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrExecutionAlreadyStartedCounter)
	case *ConditionFailedError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrConditionFailedCounter)
	case *CurrentWorkflowConditionFailedError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrCurrentWorkflowConditionFailedCounter)
	case *ShardAlreadyExistError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrShardExistsCounter)
	case *ShardOwnershipLostError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrShardOwnershipLostCounter)
	case *types.EntityNotExistsError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrEntityNotExistsCounter)
	case *TimeoutError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrTimeoutCounter)
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	case *types.ServiceBusyError:
		p.metricClient.IncCounter(scope, metrics.PersistenceErrBusyCounter)
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	default:
		p.logger.Error("Operation failed with internal error.", tag.Error(err), tag.MetricScope(scope))
		p.metricClient.IncCounter(scope, metrics.PersistenceFailures)
	}
}

func (p *persistenceMetricsClientBase) call(scope int, op func() error, tagsF func() []tag.Tag) error {
	p.metricClient.IncCounter(scope, metrics.PersistenceRequests)
	before := time.Now()
	err := op()
	duration := time.Now().Sub(before)
	p.metricClient.RecordTimer(scope, metrics.PersistenceLatency, duration)
	if p.maxExpectedLatency != nil {
		operation := p.metricClient.GetOperation(scope)
		if operation != nil {
			maxExpectedLatency := p.maxExpectedLatency(*operation)
			if duration > maxExpectedLatency {
				tags := append(tagsF(), tag.OperationName(*operation))
				p.throttledLogger.Warn(
					fmt.Sprintf("Operation took longer (%v ms) than expected (%v ms)",
						duration.Milliseconds(), maxExpectedLatency.Milliseconds()),
					tags...)
			}
		}
	}

	if p.enableLatencyHistogramMetrics {
		p.metricClient.RecordHistogramDuration(scope, metrics.PersistenceLatencyHistogram, duration)
	}

	if err != nil {
		p.updateErrorMetric(scope, err)
	}
	return err
}

func (p *shardPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *shardPersistenceClient) CreateShard(
	ctx context.Context,
	request *CreateShardRequest,
) error {
	op := func() error {
		return p.persistence.CreateShard(ctx, request)
	}
	tags := func() []tag.Tag {
		tags := []tag.Tag{}
		if request != nil && request.ShardInfo != nil {
			tags = append(tags, tag.ShardID(request.ShardInfo.ShardID))
		}
		return tags
	}
	return p.call(metrics.PersistenceCreateShardScope, op, tags)
}

func (p *shardPersistenceClient) GetShard(
	ctx context.Context,
	request *GetShardRequest,
) (*GetShardResponse, error) {
	var resp *GetShardResponse
	op := func() error {
		var err error
		resp, err = p.persistence.GetShard(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request != nil {
			return []tag.Tag{
				tag.ShardID(request.ShardID),
			}
		}
		return []tag.Tag{}
	}
	err := p.call(metrics.PersistenceGetShardScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *shardPersistenceClient) UpdateShard(
	ctx context.Context,
	request *UpdateShardRequest,
) error {
	op := func() error {
		return p.persistence.UpdateShard(ctx, request)
	}
	tags := func() []tag.Tag {
		tags := []tag.Tag{}
		if request != nil && request.ShardInfo != nil {
			tags = append(tags, tag.ShardID(request.ShardInfo.ShardID))
		}
		return tags
	}
	return p.call(metrics.PersistenceUpdateShardScope, op, tags)
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
	var resp *CreateWorkflowExecutionResponse
	op := func() error {
		var err error
		resp, err = p.persistence.CreateWorkflowExecution(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		tags := []tag.Tag{
			tag.WorkflowEndingRunID(request.PreviousRunID),
		}
		if request != nil && request.NewWorkflowSnapshot.ExecutionInfo != nil {
			tags = append(tags, tag.WorkflowDomainID(request.NewWorkflowSnapshot.ExecutionInfo.DomainID))
			tags = append(tags, tag.WorkflowID(request.NewWorkflowSnapshot.ExecutionInfo.WorkflowID))
			tags = append(tags, tag.WorkflowRunID(request.NewWorkflowSnapshot.ExecutionInfo.RunID))
		}
		return tags
	}
	err := p.call(metrics.PersistenceCreateWorkflowExecutionScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) GetWorkflowExecution(
	ctx context.Context,
	request *GetWorkflowExecutionRequest,
) (*GetWorkflowExecutionResponse, error) {
	var resp *GetWorkflowExecutionResponse
	op := func() error {
		var err error
		resp, err = p.persistence.GetWorkflowExecution(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainID),
			tag.WorkflowID(request.Execution.WorkflowID),
			tag.WorkflowRunID(request.Execution.RunID),
		}
	}
	err := p.call(metrics.PersistenceGetWorkflowExecutionScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) UpdateWorkflowExecution(
	ctx context.Context,
	request *UpdateWorkflowExecutionRequest,
) (*UpdateWorkflowExecutionResponse, error) {
	var resp *UpdateWorkflowExecutionResponse
	op := func() error {
		var err error
		resp, err = p.persistence.UpdateWorkflowExecution(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request != nil && request.UpdateWorkflowMutation.ExecutionInfo != nil {
			return []tag.Tag{
				tag.WorkflowDomainID(request.UpdateWorkflowMutation.ExecutionInfo.DomainID),
				tag.WorkflowID(request.UpdateWorkflowMutation.ExecutionInfo.WorkflowID),
				tag.WorkflowRunID(request.UpdateWorkflowMutation.ExecutionInfo.RunID),
			}
		}
		return []tag.Tag{}
	}
	err := p.call(metrics.PersistenceUpdateWorkflowExecutionScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) ConflictResolveWorkflowExecution(
	ctx context.Context,
	request *ConflictResolveWorkflowExecutionRequest,
) (*ConflictResolveWorkflowExecutionResponse, error) {
	var resp *ConflictResolveWorkflowExecutionResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ConflictResolveWorkflowExecution(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request != nil &&
			request.CurrentWorkflowMutation != nil &&
			request.CurrentWorkflowMutation.ExecutionInfo != nil {
			return []tag.Tag{
				tag.WorkflowDomainID(request.CurrentWorkflowMutation.ExecutionInfo.DomainID),
				tag.WorkflowID(request.CurrentWorkflowMutation.ExecutionInfo.WorkflowID),
				tag.WorkflowRunID(request.CurrentWorkflowMutation.ExecutionInfo.RunID),
			}
		}
		return []tag.Tag{}
	}
	err := p.call(metrics.PersistenceConflictResolveWorkflowExecutionScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) DeleteWorkflowExecution(
	ctx context.Context,
	request *DeleteWorkflowExecutionRequest,
) error {
	op := func() error {
		return p.persistence.DeleteWorkflowExecution(ctx, request)
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainID),
			tag.WorkflowID(request.WorkflowID),
			tag.WorkflowRunID(request.RunID),
		}
	}
	return p.call(metrics.PersistenceDeleteWorkflowExecutionScope, op, tags)
}

func (p *workflowExecutionPersistenceClient) DeleteCurrentWorkflowExecution(
	ctx context.Context,
	request *DeleteCurrentWorkflowExecutionRequest,
) error {
	op := func() error {
		return p.persistence.DeleteCurrentWorkflowExecution(ctx, request)
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainID),
			tag.WorkflowID(request.WorkflowID),
			tag.WorkflowRunID(request.RunID),
		}
	}
	return p.call(metrics.PersistenceDeleteCurrentWorkflowExecutionScope, op, tags)
}

func (p *workflowExecutionPersistenceClient) GetCurrentExecution(
	ctx context.Context,
	request *GetCurrentExecutionRequest,
) (*GetCurrentExecutionResponse, error) {
	var resp *GetCurrentExecutionResponse
	op := func() error {
		var err error
		resp, err = p.persistence.GetCurrentExecution(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainID),
			tag.WorkflowID(request.WorkflowID),
		}
	}
	err := p.call(metrics.PersistenceGetCurrentExecutionScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) ListCurrentExecutions(
	ctx context.Context,
	request *ListCurrentExecutionsRequest,
) (*ListCurrentExecutionsResponse, error) {
	var resp *ListCurrentExecutionsResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ListCurrentExecutions(ctx, request)
		return err
	}
	err := p.call(metrics.PersistenceListCurrentExecutionsScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) IsWorkflowExecutionExists(
	ctx context.Context,
	request *IsWorkflowExecutionExistsRequest,
) (*IsWorkflowExecutionExistsResponse, error) {
	var resp *IsWorkflowExecutionExistsResponse
	op := func() error {
		var err error
		resp, err = p.persistence.IsWorkflowExecutionExists(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainID),
			tag.WorkflowID(request.WorkflowID),
			tag.WorkflowRunID(request.RunID),
		}
	}
	err := p.call(metrics.PersistenceIsWorkflowExecutionExistsScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) ListConcreteExecutions(
	ctx context.Context,
	request *ListConcreteExecutionsRequest,
) (*ListConcreteExecutionsResponse, error) {
	var resp *ListConcreteExecutionsResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ListConcreteExecutions(ctx, request)
		return err
	}
	err := p.call(metrics.PersistenceListConcreteExecutionsScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) GetTransferTasks(
	ctx context.Context,
	request *GetTransferTasksRequest,
) (*GetTransferTasksResponse, error) {
	var resp *GetTransferTasksResponse
	op := func() error {
		var err error
		resp, err = p.persistence.GetTransferTasks(ctx, request)
		return err
	}
	err := p.call(metrics.PersistenceGetTransferTasksScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) GetCrossClusterTasks(
	ctx context.Context,
	request *GetCrossClusterTasksRequest,
) (*GetCrossClusterTasksResponse, error) {
	var resp *GetCrossClusterTasksResponse
	op := func() error {
		var err error
		resp, err = p.persistence.GetCrossClusterTasks(ctx, request)
		return err
	}
	err := p.call(metrics.PersistenceGetCrossClusterTasksScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) GetReplicationTasks(
	ctx context.Context,
	request *GetReplicationTasksRequest,
) (*GetReplicationTasksResponse, error) {
	var resp *GetReplicationTasksResponse
	op := func() error {
		var err error
		resp, err = p.persistence.GetReplicationTasks(ctx, request)
		return err
	}
	err := p.call(metrics.PersistenceGetReplicationTasksScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) CompleteTransferTask(
	ctx context.Context,
	request *CompleteTransferTaskRequest,
) error {
	op := func() error {
		return p.persistence.CompleteTransferTask(ctx, request)
	}
	return p.call(metrics.PersistenceCompleteTransferTaskScope, op, emptyTags)
}

func (p *workflowExecutionPersistenceClient) RangeCompleteTransferTask(
	ctx context.Context,
	request *RangeCompleteTransferTaskRequest,
) (*RangeCompleteTransferTaskResponse, error) {
	var resp *RangeCompleteTransferTaskResponse
	op := func() error {
		var err error
		resp, err = p.persistence.RangeCompleteTransferTask(ctx, request)
		return err
	}
	err := p.call(metrics.PersistenceRangeCompleteTransferTaskScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) CompleteCrossClusterTask(
	ctx context.Context,
	request *CompleteCrossClusterTaskRequest,
) error {
	op := func() error {
		return p.persistence.CompleteCrossClusterTask(ctx, request)
	}
	return p.call(metrics.PersistenceCompleteCrossClusterTaskScope, op, emptyTags)
}

func (p *workflowExecutionPersistenceClient) RangeCompleteCrossClusterTask(
	ctx context.Context,
	request *RangeCompleteCrossClusterTaskRequest,
) (*RangeCompleteCrossClusterTaskResponse, error) {
	var resp *RangeCompleteCrossClusterTaskResponse
	op := func() error {
		var err error
		resp, err = p.persistence.RangeCompleteCrossClusterTask(ctx, request)
		return err
	}
	err := p.call(metrics.PersistenceRangeCompleteCrossClusterTaskScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) CompleteReplicationTask(
	ctx context.Context,
	request *CompleteReplicationTaskRequest,
) error {
	op := func() error {
		return p.persistence.CompleteReplicationTask(ctx, request)
	}
	return p.call(metrics.PersistenceCompleteReplicationTaskScope, op, emptyTags)
}

func (p *workflowExecutionPersistenceClient) RangeCompleteReplicationTask(
	ctx context.Context,
	request *RangeCompleteReplicationTaskRequest,
) (*RangeCompleteReplicationTaskResponse, error) {
	var resp *RangeCompleteReplicationTaskResponse
	op := func() error {
		var err error
		resp, err = p.persistence.RangeCompleteReplicationTask(ctx, request)
		return err
	}
	err := p.call(metrics.PersistenceRangeCompleteReplicationTaskScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) PutReplicationTaskToDLQ(
	ctx context.Context,
	request *PutReplicationTaskToDLQRequest,
) error {
	op := func() error {
		return p.persistence.PutReplicationTaskToDLQ(ctx, request)
	}
	tags := func() []tag.Tag {
		if request != nil && request.TaskInfo != nil {
			return []tag.Tag{
				tag.WorkflowDomainID(request.TaskInfo.DomainID),
				tag.WorkflowID(request.TaskInfo.WorkflowID),
				tag.WorkflowRunID(request.TaskInfo.RunID),
			}
		}
		return []tag.Tag{}
	}
	return p.call(metrics.PersistencePutReplicationTaskToDLQScope, op, tags)
}

func (p *workflowExecutionPersistenceClient) GetReplicationTasksFromDLQ(
	ctx context.Context,
	request *GetReplicationTasksFromDLQRequest,
) (*GetReplicationTasksFromDLQResponse, error) {
	var resp *GetReplicationTasksFromDLQResponse
	op := func() error {
		var err error
		resp, err = p.persistence.GetReplicationTasksFromDLQ(ctx, request)
		return err
	}
	err := p.call(metrics.PersistenceGetReplicationTasksFromDLQScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) GetReplicationDLQSize(
	ctx context.Context,
	request *GetReplicationDLQSizeRequest,
) (*GetReplicationDLQSizeResponse, error) {
	var resp *GetReplicationDLQSizeResponse
	op := func() error {
		var err error
		resp, err = p.persistence.GetReplicationDLQSize(ctx, request)
		return err
	}
	err := p.call(metrics.PersistenceGetReplicationDLQSizeScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) DeleteReplicationTaskFromDLQ(
	ctx context.Context,
	request *DeleteReplicationTaskFromDLQRequest,
) error {
	op := func() error {
		return p.persistence.DeleteReplicationTaskFromDLQ(ctx, request)
	}
	return p.call(metrics.PersistenceDeleteReplicationTaskFromDLQScope, op, emptyTags)
}

func (p *workflowExecutionPersistenceClient) RangeDeleteReplicationTaskFromDLQ(
	ctx context.Context,
	request *RangeDeleteReplicationTaskFromDLQRequest,
) (*RangeDeleteReplicationTaskFromDLQResponse, error) {
	var resp *RangeDeleteReplicationTaskFromDLQResponse
	op := func() error {
		var err error
		resp, err = p.persistence.RangeDeleteReplicationTaskFromDLQ(ctx, request)
		return err
	}
	err := p.call(metrics.PersistenceRangeDeleteReplicationTaskFromDLQScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) CreateFailoverMarkerTasks(
	ctx context.Context,
	request *CreateFailoverMarkersRequest,
) error {
	op := func() error {
		return p.persistence.CreateFailoverMarkerTasks(ctx, request)
	}
	return p.call(metrics.PersistenceCreateFailoverMarkerTasksScope, op, emptyTags)
}

func (p *workflowExecutionPersistenceClient) GetTimerIndexTasks(
	ctx context.Context,
	request *GetTimerIndexTasksRequest,
) (*GetTimerIndexTasksResponse, error) {
	var resp *GetTimerIndexTasksResponse
	op := func() error {
		var err error
		resp, err = p.persistence.GetTimerIndexTasks(ctx, request)
		return err
	}
	err := p.call(metrics.PersistenceGetTimerIndexTasksScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *workflowExecutionPersistenceClient) CompleteTimerTask(
	ctx context.Context,
	request *CompleteTimerTaskRequest,
) error {
	op := func() error {
		return p.persistence.CompleteTimerTask(ctx, request)
	}
	return p.call(metrics.PersistenceCompleteTimerTaskScope, op, emptyTags)
}

func (p *workflowExecutionPersistenceClient) RangeCompleteTimerTask(
	ctx context.Context,
	request *RangeCompleteTimerTaskRequest,
) (*RangeCompleteTimerTaskResponse, error) {
	var resp *RangeCompleteTimerTaskResponse
	op := func() error {
		var err error
		resp, err = p.persistence.RangeCompleteTimerTask(ctx, request)
		return err
	}
	err := p.call(metrics.PersistenceRangeCompleteTimerTaskScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
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
	var resp *CreateTasksResponse
	op := func() error {
		var err error
		resp, err = p.persistence.CreateTasks(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request != nil && request.TaskListInfo != nil {
			return []tag.Tag{
				tag.WorkflowDomainID(request.TaskListInfo.DomainID),
				tag.NumberProcessed(len(request.Tasks)),
			}
		}
		return []tag.Tag{}
	}
	err := p.call(metrics.PersistenceCreateTaskScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *taskPersistenceClient) GetTasks(
	ctx context.Context,
	request *GetTasksRequest,
) (*GetTasksResponse, error) {
	var resp *GetTasksResponse
	op := func() error {
		var err error
		resp, err = p.persistence.GetTasks(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainID),
			tag.ReadLevel(request.ReadLevel),
			tag.MaxLevel(*request.MaxReadLevel),
			tag.NumberRequested(request.BatchSize),
		}
	}
	err := p.call(metrics.PersistenceGetTasksScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *taskPersistenceClient) CompleteTask(
	ctx context.Context,
	request *CompleteTaskRequest,
) error {
	op := func() error {
		return p.persistence.CompleteTask(ctx, request)
	}
	tags := func() []tag.Tag {
		if request != nil && request.TaskList != nil {
			return []tag.Tag{
				tag.WorkflowDomainID(request.TaskList.DomainID),
			}
		}
		return []tag.Tag{}
	}
	return p.call(metrics.PersistenceCompleteTaskScope, op, tags)
}

func (p *taskPersistenceClient) CompleteTasksLessThan(
	ctx context.Context,
	request *CompleteTasksLessThanRequest,
) (*CompleteTasksLessThanResponse, error) {
	var resp *CompleteTasksLessThanResponse
	op := func() error {
		var err error
		resp, err = p.persistence.CompleteTasksLessThan(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainID),
		}
	}
	err := p.call(metrics.PersistenceCompleteTasksLessThanScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *taskPersistenceClient) GetOrphanTasks(ctx context.Context, request *GetOrphanTasksRequest) (*GetOrphanTasksResponse, error) {
	var resp *GetOrphanTasksResponse
	op := func() error {
		var err error
		resp, err = p.persistence.GetOrphanTasks(ctx, request)
		return err
	}
	err := p.call(metrics.PersistenceGetOrphanTasksScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *taskPersistenceClient) LeaseTaskList(
	ctx context.Context,
	request *LeaseTaskListRequest,
) (*LeaseTaskListResponse, error) {
	var resp *LeaseTaskListResponse
	op := func() error {
		var err error
		resp, err = p.persistence.LeaseTaskList(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainID),
		}
	}
	err := p.call(metrics.PersistenceLeaseTaskListScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *taskPersistenceClient) ListTaskList(
	ctx context.Context,
	request *ListTaskListRequest,
) (*ListTaskListResponse, error) {
	var resp *ListTaskListResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ListTaskList(ctx, request)
		return err
	}
	err := p.call(metrics.PersistenceListTaskListScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *taskPersistenceClient) DeleteTaskList(
	ctx context.Context,
	request *DeleteTaskListRequest,
) error {
	op := func() error {
		return p.persistence.DeleteTaskList(ctx, request)
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainID),
		}
	}
	return p.call(metrics.PersistenceDeleteTaskListScope, op, tags)
}

func (p *taskPersistenceClient) UpdateTaskList(
	ctx context.Context,
	request *UpdateTaskListRequest,
) (*UpdateTaskListResponse, error) {
	var resp *UpdateTaskListResponse
	op := func() error {
		var err error
		resp, err = p.persistence.UpdateTaskList(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request != nil && request.TaskListInfo != nil {
			return []tag.Tag{
				tag.WorkflowDomainID(request.TaskListInfo.DomainID),
			}
		}
		return []tag.Tag{}
	}
	err := p.call(metrics.PersistenceUpdateTaskListScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
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
	var resp *CreateDomainResponse
	op := func() error {
		var err error
		resp, err = p.persistence.CreateDomain(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request != nil && request.Info != nil {
			return []tag.Tag{
				tag.WorkflowDomainID(request.Info.ID),
			}
		}
		return []tag.Tag{}
	}
	err := p.call(metrics.PersistenceCreateDomainScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *metadataPersistenceClient) GetDomain(
	ctx context.Context,
	request *GetDomainRequest,
) (*GetDomainResponse, error) {
	var resp *GetDomainResponse
	op := func() error {
		var err error
		resp, err = p.persistence.GetDomain(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.ID),
			tag.WorkflowDomainName(request.Name),
		}
	}
	err := p.call(metrics.PersistenceGetDomainScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *metadataPersistenceClient) UpdateDomain(
	ctx context.Context,
	request *UpdateDomainRequest,
) error {
	op := func() error {
		return p.persistence.UpdateDomain(ctx, request)
	}
	tags := func() []tag.Tag {
		if request != nil && request.Info != nil {
			return []tag.Tag{
				tag.WorkflowDomainID(request.Info.ID),
				tag.WorkflowDomainName(request.Info.Name),
			}
		}
		return []tag.Tag{}
	}
	return p.call(metrics.PersistenceUpdateDomainScope, op, tags)
}

func (p *metadataPersistenceClient) DeleteDomain(
	ctx context.Context,
	request *DeleteDomainRequest,
) error {
	op := func() error {
		return p.persistence.DeleteDomain(ctx, request)
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.ID),
		}
	}
	return p.call(metrics.PersistenceDeleteDomainScope, op, tags)
}

func (p *metadataPersistenceClient) DeleteDomainByName(
	ctx context.Context,
	request *DeleteDomainByNameRequest,
) error {
	op := func() error {
		return p.persistence.DeleteDomainByName(ctx, request)
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainName(request.Name),
		}
	}
	return p.call(metrics.PersistenceDeleteDomainByNameScope, op, tags)
}

func (p *metadataPersistenceClient) ListDomains(
	ctx context.Context,
	request *ListDomainsRequest,
) (*ListDomainsResponse, error) {
	var resp *ListDomainsResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ListDomains(ctx, request)
		return err
	}
	err := p.call(metrics.PersistenceListDomainScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *metadataPersistenceClient) GetMetadata(
	ctx context.Context,
) (*GetMetadataResponse, error) {
	var resp *GetMetadataResponse
	op := func() error {
		var err error
		resp, err = p.persistence.GetMetadata(ctx)
		return err
	}
	err := p.call(metrics.PersistenceGetMetadataScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *metadataPersistenceClient) Close() {
	p.persistence.Close()
}

func (p *visibilityPersistenceClient) GetName() string {
	return p.persistence.GetName()
}

func (p *visibilityPersistenceClient) RecordWorkflowExecutionStarted(
	ctx context.Context,
	request *RecordWorkflowExecutionStartedRequest,
) error {
	op := func() error {
		return p.persistence.RecordWorkflowExecutionStarted(ctx, request)
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainUUID),
			tag.WorkflowID(request.Execution.WorkflowID),
			tag.WorkflowRunID(request.Execution.RunID),
		}
	}
	return p.call(metrics.PersistenceRecordWorkflowExecutionStartedScope, op, tags)
}

func (p *visibilityPersistenceClient) RecordWorkflowExecutionClosed(
	ctx context.Context,
	request *RecordWorkflowExecutionClosedRequest,
) error {
	op := func() error {
		return p.persistence.RecordWorkflowExecutionClosed(ctx, request)
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainUUID),
			tag.WorkflowID(request.Execution.WorkflowID),
			tag.WorkflowRunID(request.Execution.RunID),
		}
	}
	return p.call(metrics.PersistenceRecordWorkflowExecutionClosedScope, op, tags)
}

func (p *visibilityPersistenceClient) UpsertWorkflowExecution(
	ctx context.Context,
	request *UpsertWorkflowExecutionRequest,
) error {
	op := func() error {
		return p.persistence.UpsertWorkflowExecution(ctx, request)
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainUUID),
			tag.WorkflowID(request.Execution.WorkflowID),
			tag.WorkflowRunID(request.Execution.RunID),
		}
	}
	return p.call(metrics.PersistenceUpsertWorkflowExecutionScope, op, tags)
}

func (p *visibilityPersistenceClient) ListOpenWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsRequest,
) (*ListWorkflowExecutionsResponse, error) {
	var resp *ListWorkflowExecutionsResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ListOpenWorkflowExecutions(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainUUID),
		}
	}
	err := p.call(metrics.PersistenceListOpenWorkflowExecutionsScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *visibilityPersistenceClient) ListClosedWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsRequest,
) (*ListWorkflowExecutionsResponse, error) {
	var resp *ListWorkflowExecutionsResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ListClosedWorkflowExecutions(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainUUID),
		}
	}
	err := p.call(metrics.PersistenceListClosedWorkflowExecutionsScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *visibilityPersistenceClient) ListOpenWorkflowExecutionsByType(
	ctx context.Context,
	request *ListWorkflowExecutionsByTypeRequest,
) (*ListWorkflowExecutionsResponse, error) {
	var resp *ListWorkflowExecutionsResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ListOpenWorkflowExecutionsByType(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainUUID),
		}
	}
	err := p.call(metrics.PersistenceListOpenWorkflowExecutionsByTypeScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *visibilityPersistenceClient) ListClosedWorkflowExecutionsByType(
	ctx context.Context,
	request *ListWorkflowExecutionsByTypeRequest,
) (*ListWorkflowExecutionsResponse, error) {
	var resp *ListWorkflowExecutionsResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ListClosedWorkflowExecutionsByType(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainUUID),
		}
	}
	err := p.call(metrics.PersistenceListClosedWorkflowExecutionsByTypeScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *visibilityPersistenceClient) ListOpenWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *ListWorkflowExecutionsByWorkflowIDRequest,
) (*ListWorkflowExecutionsResponse, error) {
	var resp *ListWorkflowExecutionsResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ListOpenWorkflowExecutionsByWorkflowID(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainUUID),
		}
	}
	err := p.call(metrics.PersistenceListOpenWorkflowExecutionsByWorkflowIDScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *visibilityPersistenceClient) ListClosedWorkflowExecutionsByWorkflowID(
	ctx context.Context,
	request *ListWorkflowExecutionsByWorkflowIDRequest,
) (*ListWorkflowExecutionsResponse, error) {
	var resp *ListWorkflowExecutionsResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ListClosedWorkflowExecutionsByWorkflowID(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainUUID),
		}
	}
	err := p.call(metrics.PersistenceListClosedWorkflowExecutionsByWorkflowIDScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *visibilityPersistenceClient) ListClosedWorkflowExecutionsByStatus(
	ctx context.Context,
	request *ListClosedWorkflowExecutionsByStatusRequest,
) (*ListWorkflowExecutionsResponse, error) {
	var resp *ListWorkflowExecutionsResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ListClosedWorkflowExecutionsByStatus(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainUUID),
		}
	}
	err := p.call(metrics.PersistenceListClosedWorkflowExecutionsByStatusScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *visibilityPersistenceClient) GetClosedWorkflowExecution(
	ctx context.Context,
	request *GetClosedWorkflowExecutionRequest,
) (*GetClosedWorkflowExecutionResponse, error) {
	var resp *GetClosedWorkflowExecutionResponse
	op := func() error {
		var err error
		resp, err = p.persistence.GetClosedWorkflowExecution(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainUUID),
			tag.WorkflowID(request.Execution.WorkflowID),
			tag.WorkflowRunID(request.Execution.RunID),
		}
	}
	err := p.call(metrics.PersistenceGetClosedWorkflowExecutionScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *visibilityPersistenceClient) DeleteWorkflowExecution(
	ctx context.Context,
	request *VisibilityDeleteWorkflowExecutionRequest,
) error {
	op := func() error {
		return p.persistence.DeleteWorkflowExecution(ctx, request)
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainID),
			tag.WorkflowID(request.WorkflowID),
			tag.WorkflowRunID(request.RunID),
		}
	}
	return p.call(metrics.PersistenceVisibilityDeleteWorkflowExecutionScope, op, tags)
}

func (p *visibilityPersistenceClient) ListWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsByQueryRequest,
) (*ListWorkflowExecutionsResponse, error) {
	var resp *ListWorkflowExecutionsResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ListWorkflowExecutions(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainUUID),
		}
	}
	err := p.call(metrics.PersistenceListWorkflowExecutionsScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *visibilityPersistenceClient) ScanWorkflowExecutions(
	ctx context.Context,
	request *ListWorkflowExecutionsByQueryRequest,
) (*ListWorkflowExecutionsResponse, error) {
	var resp *ListWorkflowExecutionsResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ScanWorkflowExecutions(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainUUID),
		}
	}
	err := p.call(metrics.PersistenceScanWorkflowExecutionsScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *visibilityPersistenceClient) CountWorkflowExecutions(
	ctx context.Context,
	request *CountWorkflowExecutionsRequest,
) (*CountWorkflowExecutionsResponse, error) {
	var resp *CountWorkflowExecutionsResponse
	op := func() error {
		var err error
		resp, err = p.persistence.CountWorkflowExecutions(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request == nil {
			return []tag.Tag{}
		}
		return []tag.Tag{
			tag.WorkflowDomainID(request.DomainUUID),
		}
	}
	err := p.call(metrics.PersistenceCountWorkflowExecutionsScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
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
	var resp *AppendHistoryNodesResponse
	op := func() error {
		var err error
		resp, err = p.persistence.AppendHistoryNodes(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request != nil && request.ShardID != nil {
			return []tag.Tag{
				tag.ShardID(*request.ShardID),
			}
		}
		return []tag.Tag{}
	}
	err := p.call(metrics.PersistenceAppendHistoryNodesScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ReadHistoryBranch returns history node data for a branch
func (p *historyPersistenceClient) ReadHistoryBranch(
	ctx context.Context,
	request *ReadHistoryBranchRequest,
) (*ReadHistoryBranchResponse, error) {
	var resp *ReadHistoryBranchResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ReadHistoryBranch(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request != nil && request.ShardID != nil {
			return []tag.Tag{
				tag.ShardID(*request.ShardID),
			}
		}
		return []tag.Tag{}
	}
	err := p.call(metrics.PersistenceReadHistoryBranchScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ReadHistoryBranchByBatch returns history node data for a branch ByBatch
func (p *historyPersistenceClient) ReadHistoryBranchByBatch(
	ctx context.Context,
	request *ReadHistoryBranchRequest,
) (*ReadHistoryBranchByBatchResponse, error) {
	var resp *ReadHistoryBranchByBatchResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ReadHistoryBranchByBatch(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request != nil && request.ShardID != nil {
			return []tag.Tag{
				tag.ShardID(*request.ShardID),
			}
		}
		return []tag.Tag{}
	}
	err := p.call(metrics.PersistenceReadHistoryBranchScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ReadRawHistoryBranch returns history node raw data for a branch ByBatch
func (p *historyPersistenceClient) ReadRawHistoryBranch(
	ctx context.Context,
	request *ReadHistoryBranchRequest,
) (*ReadRawHistoryBranchResponse, error) {
	var resp *ReadRawHistoryBranchResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ReadRawHistoryBranch(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request != nil && request.ShardID != nil {
			return []tag.Tag{
				tag.ShardID(*request.ShardID),
			}
		}
		return []tag.Tag{}
	}
	err := p.call(metrics.PersistenceReadHistoryBranchScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ForkHistoryBranch forks a new branch from a old branch
func (p *historyPersistenceClient) ForkHistoryBranch(
	ctx context.Context,
	request *ForkHistoryBranchRequest,
) (*ForkHistoryBranchResponse, error) {
	var resp *ForkHistoryBranchResponse
	op := func() error {
		var err error
		resp, err = p.persistence.ForkHistoryBranch(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request != nil && request.ShardID != nil {
			return []tag.Tag{
				tag.ShardID(*request.ShardID),
			}
		}
		return []tag.Tag{}
	}
	err := p.call(metrics.PersistenceForkHistoryBranchScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteHistoryBranch removes a branch
func (p *historyPersistenceClient) DeleteHistoryBranch(
	ctx context.Context,
	request *DeleteHistoryBranchRequest,
) error {
	op := func() error {
		return p.persistence.DeleteHistoryBranch(ctx, request)
	}
	tags := func() []tag.Tag {
		if request != nil && request.ShardID != nil {
			return []tag.Tag{
				tag.ShardID(*request.ShardID),
			}
		}
		return []tag.Tag{}
	}
	return p.call(metrics.PersistenceDeleteHistoryBranchScope, op, tags)
}

func (p *historyPersistenceClient) GetAllHistoryTreeBranches(
	ctx context.Context,
	request *GetAllHistoryTreeBranchesRequest,
) (*GetAllHistoryTreeBranchesResponse, error) {
	var resp *GetAllHistoryTreeBranchesResponse
	op := func() error {
		var err error
		resp, err = p.persistence.GetAllHistoryTreeBranches(ctx, request)
		return err
	}
	err := p.call(metrics.PersistenceGetAllHistoryTreeBranchesScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetHistoryTree returns all branch information of a tree
func (p *historyPersistenceClient) GetHistoryTree(
	ctx context.Context,
	request *GetHistoryTreeRequest,
) (*GetHistoryTreeResponse, error) {
	var resp *GetHistoryTreeResponse
	op := func() error {
		var err error
		resp, err = p.persistence.GetHistoryTree(ctx, request)
		return err
	}
	tags := func() []tag.Tag {
		if request != nil && request.ShardID != nil {
			return []tag.Tag{
				tag.ShardID(*request.ShardID),
			}
		}
		return []tag.Tag{}
	}
	err := p.call(metrics.PersistenceGetHistoryTreeScope, op, tags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *queuePersistenceClient) EnqueueMessage(
	ctx context.Context,
	message []byte,
) error {
	op := func() error {
		return p.persistence.EnqueueMessage(ctx, message)
	}
	return p.call(metrics.PersistenceEnqueueMessageScope, op, emptyTags)
}

func (p *queuePersistenceClient) ReadMessages(
	ctx context.Context,
	lastMessageID int64,
	maxCount int,
) ([]*QueueMessage, error) {
	var resp []*QueueMessage
	op := func() error {
		var err error
		resp, err = p.persistence.ReadMessages(ctx, lastMessageID, maxCount)
		return err
	}
	err := p.call(metrics.PersistenceReadQueueMessagesScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *queuePersistenceClient) UpdateAckLevel(
	ctx context.Context,
	messageID int64,
	clusterName string,
) error {
	op := func() error {
		return p.persistence.UpdateAckLevel(ctx, messageID, clusterName)
	}
	return p.call(metrics.PersistenceUpdateAckLevelScope, op, emptyTags)
}

func (p *queuePersistenceClient) GetAckLevels(
	ctx context.Context,
) (map[string]int64, error) {
	var resp map[string]int64
	op := func() error {
		var err error
		resp, err = p.persistence.GetAckLevels(ctx)
		return err
	}
	err := p.call(metrics.PersistenceGetAckLevelScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *queuePersistenceClient) DeleteMessagesBefore(
	ctx context.Context,
	messageID int64,
) error {
	op := func() error {
		return p.persistence.DeleteMessagesBefore(ctx, messageID)
	}
	return p.call(metrics.PersistenceDeleteQueueMessagesScope, op, emptyTags)
}

func (p *queuePersistenceClient) EnqueueMessageToDLQ(
	ctx context.Context,
	message []byte,
) error {
	op := func() error {
		return p.persistence.EnqueueMessageToDLQ(ctx, message)
	}
	return p.call(metrics.PersistenceEnqueueMessageToDLQScope, op, emptyTags)
}

func (p *queuePersistenceClient) ReadMessagesFromDLQ(
	ctx context.Context,
	firstMessageID int64,
	lastMessageID int64,
	pageSize int,
	pageToken []byte,
) ([]*QueueMessage, []byte, error) {
	var result []*QueueMessage
	var token []byte
	op := func() error {
		var err error
		result, token, err = p.persistence.ReadMessagesFromDLQ(ctx, firstMessageID, lastMessageID, pageSize, pageToken)
		return err
	}
	err := p.call(metrics.PersistenceReadQueueMessagesFromDLQScope, op, emptyTags)
	if err != nil {
		return nil, nil, err
	}
	return result, token, nil
}

func (p *queuePersistenceClient) DeleteMessageFromDLQ(
	ctx context.Context,
	messageID int64,
) error {
	op := func() error {
		return p.persistence.DeleteMessageFromDLQ(ctx, messageID)
	}
	return p.call(metrics.PersistenceDeleteQueueMessageFromDLQScope, op, emptyTags)
}

func (p *queuePersistenceClient) RangeDeleteMessagesFromDLQ(
	ctx context.Context,
	firstMessageID int64,
	lastMessageID int64,
) error {
	op := func() error {
		return p.persistence.RangeDeleteMessagesFromDLQ(ctx, firstMessageID, lastMessageID)
	}
	return p.call(metrics.PersistenceRangeDeleteMessagesFromDLQScope, op, emptyTags)
}

func (p *queuePersistenceClient) UpdateDLQAckLevel(
	ctx context.Context,
	messageID int64,
	clusterName string,
) error {
	op := func() error {
		return p.persistence.UpdateDLQAckLevel(ctx, messageID, clusterName)
	}
	return p.call(metrics.PersistenceUpdateDLQAckLevelScope, op, emptyTags)
}

func (p *queuePersistenceClient) GetDLQAckLevels(
	ctx context.Context,
) (map[string]int64, error) {
	var resp map[string]int64
	op := func() error {
		var err error
		resp, err = p.persistence.GetDLQAckLevels(ctx)
		return err
	}
	err := p.call(metrics.PersistenceGetDLQAckLevelScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *queuePersistenceClient) GetDLQSize(
	ctx context.Context,
) (int64, error) {
	var resp int64
	op := func() error {
		var err error
		resp, err = p.persistence.GetDLQSize(ctx)
		return err
	}
	err := p.call(metrics.PersistenceGetDLQSizeScope, op, emptyTags)
	if err != nil {
		return 0, err
	}
	return resp, nil
}

func (p *queuePersistenceClient) Close() {
	p.persistence.Close()
}

func (p *configStorePersistenceClient) FetchDynamicConfig(ctx context.Context) (*FetchDynamicConfigResponse, error) {
	var resp *FetchDynamicConfigResponse
	op := func() error {
		var err error
		resp, err = p.persistence.FetchDynamicConfig(ctx)
		return err
	}
	err := p.call(metrics.PersistenceFetchDynamicConfigScope, op, emptyTags)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *configStorePersistenceClient) UpdateDynamicConfig(ctx context.Context, request *UpdateDynamicConfigRequest) error {
	op := func() error {
		return p.persistence.UpdateDynamicConfig(ctx, request)
	}
	return p.call(metrics.PersistenceUpdateDynamicConfigScope, op, emptyTags)
}

func (p *configStorePersistenceClient) Close() {
	p.persistence.Close()
}
