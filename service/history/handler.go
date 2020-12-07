// Copyright (c) 2017-2020 Uber Technologies Inc.
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

package history

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pborman/uuid"
	"go.uber.org/yarpc/yarpcerrors"

	"github.com/uber/cadence/.gen/go/health"
	hist "github.com/uber/cadence/.gen/go/history"
	r "github.com/uber/cadence/.gen/go/replicator"
	gen "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/definition"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/mapper/thrift"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/engine"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/failover"
	"github.com/uber/cadence/service/history/replication"
	"github.com/uber/cadence/service/history/resource"
	"github.com/uber/cadence/service/history/shard"
	"github.com/uber/cadence/service/history/task"
)

//go:generate mockgen -copyright_file=../../LICENSE -package $GOPACKAGE -source $GOFILE -destination handler_mock.go -package history github.com/uber/cadence/service/history Handler

type (
	// Handler interface for history service
	Handler interface {
		Health(context.Context) (*health.HealthStatus, error)
		CloseShard(context.Context, *gen.CloseShardRequest) error
		DescribeHistoryHost(context.Context, *gen.DescribeHistoryHostRequest) (*gen.DescribeHistoryHostResponse, error)
		DescribeMutableState(context.Context, *hist.DescribeMutableStateRequest) (*hist.DescribeMutableStateResponse, error)
		DescribeQueue(context.Context, *gen.DescribeQueueRequest) (*gen.DescribeQueueResponse, error)
		DescribeWorkflowExecution(context.Context, *hist.DescribeWorkflowExecutionRequest) (*gen.DescribeWorkflowExecutionResponse, error)
		GetDLQReplicationMessages(context.Context, *r.GetDLQReplicationMessagesRequest) (*r.GetDLQReplicationMessagesResponse, error)
		GetMutableState(context.Context, *hist.GetMutableStateRequest) (*hist.GetMutableStateResponse, error)
		GetReplicationMessages(context.Context, *r.GetReplicationMessagesRequest) (*r.GetReplicationMessagesResponse, error)
		MergeDLQMessages(context.Context, *r.MergeDLQMessagesRequest) (*r.MergeDLQMessagesResponse, error)
		NotifyFailoverMarkers(context.Context, *types.NotifyFailoverMarkersRequest) error
		PollMutableState(context.Context, *hist.PollMutableStateRequest) (*hist.PollMutableStateResponse, error)
		PurgeDLQMessages(context.Context, *r.PurgeDLQMessagesRequest) error
		QueryWorkflow(context.Context, *hist.QueryWorkflowRequest) (*hist.QueryWorkflowResponse, error)
		ReadDLQMessages(context.Context, *r.ReadDLQMessagesRequest) (*r.ReadDLQMessagesResponse, error)
		ReapplyEvents(context.Context, *hist.ReapplyEventsRequest) error
		RecordActivityTaskHeartbeat(context.Context, *hist.RecordActivityTaskHeartbeatRequest) (*gen.RecordActivityTaskHeartbeatResponse, error)
		RecordActivityTaskStarted(context.Context, *hist.RecordActivityTaskStartedRequest) (*hist.RecordActivityTaskStartedResponse, error)
		RecordChildExecutionCompleted(context.Context, *hist.RecordChildExecutionCompletedRequest) error
		RecordDecisionTaskStarted(context.Context, *hist.RecordDecisionTaskStartedRequest) (*hist.RecordDecisionTaskStartedResponse, error)
		RefreshWorkflowTasks(context.Context, *hist.RefreshWorkflowTasksRequest) error
		RemoveSignalMutableState(context.Context, *hist.RemoveSignalMutableStateRequest) error
		RemoveTask(context.Context, *gen.RemoveTaskRequest) error
		ReplicateEventsV2(context.Context, *hist.ReplicateEventsV2Request) error
		RequestCancelWorkflowExecution(context.Context, *hist.RequestCancelWorkflowExecutionRequest) error
		ResetQueue(context.Context, *gen.ResetQueueRequest) error
		ResetStickyTaskList(context.Context, *hist.ResetStickyTaskListRequest) (*hist.ResetStickyTaskListResponse, error)
		ResetWorkflowExecution(context.Context, *hist.ResetWorkflowExecutionRequest) (*gen.ResetWorkflowExecutionResponse, error)
		RespondActivityTaskCanceled(context.Context, *hist.RespondActivityTaskCanceledRequest) error
		RespondActivityTaskCompleted(context.Context, *hist.RespondActivityTaskCompletedRequest) error
		RespondActivityTaskFailed(context.Context, *hist.RespondActivityTaskFailedRequest) error
		RespondDecisionTaskCompleted(context.Context, *hist.RespondDecisionTaskCompletedRequest) (*hist.RespondDecisionTaskCompletedResponse, error)
		RespondDecisionTaskFailed(context.Context, *hist.RespondDecisionTaskFailedRequest) error
		ScheduleDecisionTask(context.Context, *hist.ScheduleDecisionTaskRequest) error
		SignalWithStartWorkflowExecution(context.Context, *hist.SignalWithStartWorkflowExecutionRequest) (*gen.StartWorkflowExecutionResponse, error)
		SignalWorkflowExecution(context.Context, *hist.SignalWorkflowExecutionRequest) error
		StartWorkflowExecution(context.Context, *hist.StartWorkflowExecutionRequest) (*gen.StartWorkflowExecutionResponse, error)
		SyncActivity(context.Context, *hist.SyncActivityRequest) error
		SyncShardStatus(context.Context, *hist.SyncShardStatusRequest) error
		TerminateWorkflowExecution(context.Context, *hist.TerminateWorkflowExecutionRequest) error
	}

	// handlerImpl is an implementation for history service independent of wire protocol
	handlerImpl struct {
		resource.Resource

		shuttingDown            int32
		controller              shard.Controller
		tokenSerializer         common.TaskTokenSerializer
		startWG                 sync.WaitGroup
		config                  *config.Config
		historyEventNotifier    events.Notifier
		rateLimiter             quotas.Limiter
		replicationTaskFetchers replication.TaskFetchers
		queueTaskProcessor      task.Processor
		failoverCoordinator     failover.Coordinator
	}
)

var _ Handler = (*handlerImpl)(nil)
var _ shard.EngineFactory = (*handlerImpl)(nil)

var (
	errDomainNotSet            = &types.BadRequestError{Message: "Domain not set on request."}
	errWorkflowExecutionNotSet = &types.BadRequestError{Message: "WorkflowExecution not set on request."}
	errTaskListNotSet          = &types.BadRequestError{Message: "Tasklist not set."}
	errWorkflowIDNotSet        = &types.BadRequestError{Message: "WorkflowId is not set on request."}
	errRunIDNotValid           = &types.BadRequestError{Message: "RunID is not valid UUID."}
	errSourceClusterNotSet     = &types.BadRequestError{Message: "Source Cluster not set on request."}
	errShardIDNotSet           = &types.BadRequestError{Message: "Shard ID not set on request."}
	errTimestampNotSet         = &types.BadRequestError{Message: "Timestamp not set on request."}
	errInvalidTaskType         = &types.BadRequestError{Message: "Invalid task type"}
	errHistoryHostThrottle     = &types.ServiceBusyError{Message: "History host rps exceeded"}
	errShuttingDown            = &types.InternalServiceError{Message: "Shutting down"}
)

// NewHandler creates a thrift handler for the history service
func NewHandler(
	resource resource.Resource,
	config *config.Config,
) *handlerImpl {
	handler := &handlerImpl{
		Resource:        resource,
		config:          config,
		tokenSerializer: common.NewJSONTaskTokenSerializer(),
		rateLimiter: quotas.NewDynamicRateLimiter(
			func() float64 {
				return float64(config.RPS())
			},
		),
	}

	// prevent us from trying to serve requests before shard controller is started and ready
	handler.startWG.Add(1)
	return handler
}

// Start starts the handler
func (h *handlerImpl) Start() {

	h.replicationTaskFetchers = replication.NewTaskFetchers(
		h.GetLogger(),
		h.config,
		h.GetClusterMetadata(),
		h.GetClientBean(),
	)

	h.replicationTaskFetchers.Start()

	if h.config.EnablePriorityTaskProcessor() {
		var err error
		taskPriorityAssigner := task.NewPriorityAssigner(
			h.GetClusterMetadata().GetCurrentClusterName(),
			h.GetDomainCache(),
			h.GetLogger(),
			h.GetMetricsClient(),
			h.config,
		)

		h.queueTaskProcessor, err = task.NewProcessor(
			taskPriorityAssigner,
			h.config,
			h.GetLogger(),
			h.GetMetricsClient(),
		)
		if err != nil {
			h.GetLogger().Fatal("Creating priority task processor failed", tag.Error(err))
		}
		h.queueTaskProcessor.Start()
	}

	h.controller = shard.NewShardController(
		h.Resource,
		h,
		h.config,
	)
	h.historyEventNotifier = events.NewNotifier(h.GetTimeSource(), h.GetMetricsClient(), h.config.GetShardID)
	// events notifier must starts before controller
	h.historyEventNotifier.Start()

	h.failoverCoordinator = failover.NewCoordinator(
		h.GetMetadataManager(),
		h.GetHistoryClient(),
		h.GetTimeSource(),
		h.config,
		h.GetMetricsClient(),
		h.GetLogger(),
	)
	if h.config.EnableGracefulFailover() {
		h.failoverCoordinator.Start()
	}

	h.controller.Start()

	h.startWG.Done()
}

// Stop stops the handler
func (h *handlerImpl) Stop() {
	h.PrepareToStop()
	h.replicationTaskFetchers.Stop()
	if h.queueTaskProcessor != nil {
		h.queueTaskProcessor.Stop()
	}
	h.controller.Stop()
	h.historyEventNotifier.Stop()
	h.failoverCoordinator.Stop()
}

// PrepareToStop starts graceful traffic drain in preparation for shutdown
func (h *handlerImpl) PrepareToStop() {
	atomic.StoreInt32(&h.shuttingDown, 1)
}

func (h *handlerImpl) isShuttingDown() bool {
	return atomic.LoadInt32(&h.shuttingDown) != 0
}

// CreateEngine is implementation for HistoryEngineFactory used for creating the engine instance for shard
func (h *handlerImpl) CreateEngine(
	shardContext shard.Context,
) engine.Engine {
	return NewEngineWithShardContext(
		shardContext,
		h.GetVisibilityManager(),
		h.GetMatchingClient(),
		h.GetHistoryClient(),
		h.GetSDKClient(),
		h.historyEventNotifier,
		h.config,
		h.replicationTaskFetchers,
		h.GetMatchingRawClient(),
		h.queueTaskProcessor,
		h.failoverCoordinator,
	)
}

// Health is for health check
func (h *handlerImpl) Health(ctx context.Context) (*health.HealthStatus, error) {
	h.startWG.Wait()
	h.GetLogger().Debug("History health check endpoint reached.")
	hs := &health.HealthStatus{Ok: true, Msg: common.StringPtr("OK")}
	return hs, nil
}

// RecordActivityTaskHeartbeat - Record Activity Task Heart beat.
func (h *handlerImpl) RecordActivityTaskHeartbeat(
	ctx context.Context,
	wrappedRequest *hist.RecordActivityTaskHeartbeatRequest,
) (resp *gen.RecordActivityTaskHeartbeatResponse, retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryRecordActivityTaskHeartbeatScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	domainID := wrappedRequest.GetDomainUUID()
	if domainID == "" {
		return nil, h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return nil, h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	heartbeatRequest := wrappedRequest.HeartbeatRequest
	token, err0 := h.tokenSerializer.Deserialize(heartbeatRequest.TaskToken)
	if err0 != nil {
		err0 = &types.BadRequestError{Message: fmt.Sprintf("Error deserializing task token. Error: %v", err0)}
		return nil, h.error(err0, scope, domainID, "")
	}

	err0 = validateTaskToken(token)
	if err0 != nil {
		return nil, h.error(err0, scope, domainID, "")
	}
	workflowID := token.WorkflowID

	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return nil, h.error(err1, scope, domainID, workflowID)
	}

	response, err2 := engine.RecordActivityTaskHeartbeat(ctx, wrappedRequest)
	if err2 != nil {
		return nil, h.error(err2, scope, domainID, workflowID)
	}

	return response, nil
}

// RecordActivityTaskStarted - Record Activity Task started.
func (h *handlerImpl) RecordActivityTaskStarted(
	ctx context.Context,
	recordRequest *hist.RecordActivityTaskStartedRequest,
) (resp *hist.RecordActivityTaskStartedResponse, retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryRecordActivityTaskStartedScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	domainID := recordRequest.GetDomainUUID()
	workflowExecution := recordRequest.WorkflowExecution
	workflowID := workflowExecution.GetWorkflowId()
	if recordRequest.GetDomainUUID() == "" {
		return nil, h.error(errDomainNotSet, scope, domainID, workflowID)
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return nil, h.error(errHistoryHostThrottle, scope, domainID, workflowID)
	}

	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return nil, h.error(err1, scope, domainID, workflowID)
	}

	response, err2 := engine.RecordActivityTaskStarted(ctx, recordRequest)
	if err2 != nil {
		return nil, h.error(err2, scope, domainID, workflowID)
	}

	return response, nil
}

// RecordDecisionTaskStarted - Record Decision Task started.
func (h *handlerImpl) RecordDecisionTaskStarted(
	ctx context.Context,
	recordRequest *hist.RecordDecisionTaskStartedRequest,
) (resp *hist.RecordDecisionTaskStartedResponse, retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()
	h.GetLogger().Debug(fmt.Sprintf("RecordDecisionTaskStarted. DomainID: %v, WorkflowID: %v, RunID: %v, ScheduleID: %v",
		recordRequest.GetDomainUUID(),
		recordRequest.WorkflowExecution.GetWorkflowId(),
		common.StringDefault(recordRequest.WorkflowExecution.RunId),
		recordRequest.GetScheduleId()))

	scope := metrics.HistoryRecordDecisionTaskStartedScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	domainID := recordRequest.GetDomainUUID()
	workflowExecution := recordRequest.WorkflowExecution
	workflowID := workflowExecution.GetWorkflowId()
	if domainID == "" {
		return nil, h.error(errDomainNotSet, scope, domainID, workflowID)
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return nil, h.error(errHistoryHostThrottle, scope, domainID, workflowID)
	}

	if recordRequest.PollRequest == nil || recordRequest.PollRequest.TaskList.GetName() == "" {
		return nil, h.error(errTaskListNotSet, scope, domainID, workflowID)
	}

	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		h.GetLogger().Error("RecordDecisionTaskStarted failed.",
			tag.Error(err1),
			tag.WorkflowID(recordRequest.WorkflowExecution.GetWorkflowId()),
			tag.WorkflowRunID(recordRequest.WorkflowExecution.GetRunId()),
			tag.WorkflowScheduleID(recordRequest.GetScheduleId()),
		)
		return nil, h.error(err1, scope, domainID, workflowID)
	}

	response, err2 := engine.RecordDecisionTaskStarted(ctx, recordRequest)
	if err2 != nil {
		return nil, h.error(err2, scope, domainID, workflowID)
	}

	return response, nil
}

// RespondActivityTaskCompleted - records completion of an activity task
func (h *handlerImpl) RespondActivityTaskCompleted(
	ctx context.Context,
	wrappedRequest *hist.RespondActivityTaskCompletedRequest,
) (retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryRespondActivityTaskCompletedScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	domainID := wrappedRequest.GetDomainUUID()
	if domainID == "" {
		return h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	completeRequest := wrappedRequest.CompleteRequest
	token, err0 := h.tokenSerializer.Deserialize(completeRequest.TaskToken)
	if err0 != nil {
		err0 = &types.BadRequestError{Message: fmt.Sprintf("Error deserializing task token. Error: %v", err0)}
		return h.error(err0, scope, domainID, "")
	}

	err0 = validateTaskToken(token)
	if err0 != nil {
		return h.error(err0, scope, domainID, "")
	}
	workflowID := token.WorkflowID

	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return h.error(err1, scope, domainID, workflowID)
	}

	err2 := engine.RespondActivityTaskCompleted(ctx, wrappedRequest)
	if err2 != nil {
		return h.error(err2, scope, domainID, workflowID)
	}

	return nil
}

// RespondActivityTaskFailed - records failure of an activity task
func (h *handlerImpl) RespondActivityTaskFailed(
	ctx context.Context,
	wrappedRequest *hist.RespondActivityTaskFailedRequest,
) (retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryRespondActivityTaskFailedScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	domainID := wrappedRequest.GetDomainUUID()
	if domainID == "" {
		return h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	failRequest := wrappedRequest.FailedRequest
	token, err0 := h.tokenSerializer.Deserialize(failRequest.TaskToken)
	if err0 != nil {
		err0 = &types.BadRequestError{Message: fmt.Sprintf("Error deserializing task token. Error: %v", err0)}
		return h.error(err0, scope, domainID, "")
	}

	err0 = validateTaskToken(token)
	if err0 != nil {
		return h.error(err0, scope, domainID, "")
	}
	workflowID := token.WorkflowID

	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return h.error(err1, scope, domainID, workflowID)
	}

	err2 := engine.RespondActivityTaskFailed(ctx, wrappedRequest)
	if err2 != nil {
		return h.error(err2, scope, domainID, workflowID)
	}

	return nil
}

// RespondActivityTaskCanceled - records failure of an activity task
func (h *handlerImpl) RespondActivityTaskCanceled(
	ctx context.Context,
	wrappedRequest *hist.RespondActivityTaskCanceledRequest,
) (retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryRespondActivityTaskCanceledScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	domainID := wrappedRequest.GetDomainUUID()
	if domainID == "" {
		return h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	cancelRequest := wrappedRequest.CancelRequest
	token, err0 := h.tokenSerializer.Deserialize(cancelRequest.TaskToken)
	if err0 != nil {
		err0 = &types.BadRequestError{Message: fmt.Sprintf("Error deserializing task token. Error: %v", err0)}
		return h.error(err0, scope, domainID, "")
	}

	err0 = validateTaskToken(token)
	if err0 != nil {
		return h.error(err0, scope, domainID, "")
	}
	workflowID := token.WorkflowID

	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return h.error(err1, scope, domainID, workflowID)
	}

	err2 := engine.RespondActivityTaskCanceled(ctx, wrappedRequest)
	if err2 != nil {
		return h.error(err2, scope, domainID, workflowID)
	}

	return nil
}

// RespondDecisionTaskCompleted - records completion of a decision task
func (h *handlerImpl) RespondDecisionTaskCompleted(
	ctx context.Context,
	wrappedRequest *hist.RespondDecisionTaskCompletedRequest,
) (resp *hist.RespondDecisionTaskCompletedResponse, retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryRespondDecisionTaskCompletedScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	domainID := wrappedRequest.GetDomainUUID()
	if domainID == "" {
		return nil, h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return nil, h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	completeRequest := wrappedRequest.CompleteRequest
	if len(completeRequest.Decisions) == 0 {
		h.GetMetricsClient().IncCounter(scope, metrics.EmptyCompletionDecisionsCounter)
	}
	token, err0 := h.tokenSerializer.Deserialize(completeRequest.TaskToken)
	if err0 != nil {
		err0 = &types.BadRequestError{Message: fmt.Sprintf("Error deserializing task token. Error: %v", err0)}
		return nil, h.error(err0, scope, domainID, "")
	}

	h.GetLogger().Debug(fmt.Sprintf("RespondDecisionTaskCompleted. DomainID: %v, WorkflowID: %v, RunID: %v, ScheduleID: %v",
		token.DomainID,
		token.WorkflowID,
		token.RunID,
		token.ScheduleID))

	err0 = validateTaskToken(token)
	if err0 != nil {
		return nil, h.error(err0, scope, domainID, "")
	}
	workflowID := token.WorkflowID

	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return nil, h.error(err1, scope, domainID, workflowID)
	}

	response, err2 := engine.RespondDecisionTaskCompleted(ctx, wrappedRequest)
	if err2 != nil {
		return nil, h.error(err2, scope, domainID, workflowID)
	}

	return response, nil
}

// RespondDecisionTaskFailed - failed response to decision task
func (h *handlerImpl) RespondDecisionTaskFailed(
	ctx context.Context,
	wrappedRequest *hist.RespondDecisionTaskFailedRequest,
) (retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryRespondDecisionTaskFailedScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	domainID := wrappedRequest.GetDomainUUID()
	if domainID == "" {
		return h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	failedRequest := wrappedRequest.FailedRequest
	token, err0 := h.tokenSerializer.Deserialize(failedRequest.TaskToken)
	if err0 != nil {
		err0 = &types.BadRequestError{Message: fmt.Sprintf("Error deserializing task token. Error: %v", err0)}
		return h.error(err0, scope, domainID, "")
	}

	h.GetLogger().Debug(fmt.Sprintf("RespondDecisionTaskFailed. DomainID: %v, WorkflowID: %v, RunID: %v, ScheduleID: %v",
		token.DomainID,
		token.WorkflowID,
		token.RunID,
		token.ScheduleID))

	if failedRequest != nil && failedRequest.GetCause() == gen.DecisionTaskFailedCauseUnhandledDecision {
		h.GetLogger().Info("Non-Deterministic Error", tag.WorkflowDomainID(token.DomainID), tag.WorkflowID(token.WorkflowID), tag.WorkflowRunID(token.RunID))
		domainName, err := h.GetDomainCache().GetDomainName(token.DomainID)
		var domainTag metrics.Tag

		if err == nil {
			domainTag = metrics.DomainTag(domainName)
		} else {
			domainTag = metrics.DomainUnknownTag()
		}

		h.GetMetricsClient().Scope(scope, domainTag).IncCounter(metrics.CadenceErrNonDeterministicCounter)
	}
	err0 = validateTaskToken(token)
	if err0 != nil {
		return h.error(err0, scope, domainID, "")
	}
	workflowID := token.WorkflowID

	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return h.error(err1, scope, domainID, workflowID)
	}

	err2 := engine.RespondDecisionTaskFailed(ctx, wrappedRequest)
	if err2 != nil {
		return h.error(err2, scope, domainID, workflowID)
	}

	return nil
}

// StartWorkflowExecution - creates a new workflow execution
func (h *handlerImpl) StartWorkflowExecution(
	ctx context.Context,
	wrappedRequest *hist.StartWorkflowExecutionRequest,
) (resp *gen.StartWorkflowExecutionResponse, retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryStartWorkflowExecutionScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	domainID := wrappedRequest.GetDomainUUID()
	if domainID == "" {
		return nil, h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return nil, h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	startRequest := wrappedRequest.StartRequest
	workflowID := startRequest.GetWorkflowId()
	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return nil, h.error(err1, scope, domainID, workflowID)
	}

	response, err2 := engine.StartWorkflowExecution(ctx, wrappedRequest)
	if err2 != nil {
		return nil, h.error(err2, scope, domainID, workflowID)
	}

	return response, nil
}

// DescribeHistoryHost returns information about the internal states of a history host
func (h *handlerImpl) DescribeHistoryHost(
	ctx context.Context,
	request *gen.DescribeHistoryHostRequest,
) (resp *gen.DescribeHistoryHostResponse, retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	numOfItemsInCacheByID, numOfItemsInCacheByName := h.GetDomainCache().GetCacheSize()
	status := ""
	switch h.controller.Status() {
	case common.DaemonStatusInitialized:
		status = "initialized"
	case common.DaemonStatusStarted:
		status = "started"
	case common.DaemonStatusStopped:
		status = "stopped"
	}

	resp = &gen.DescribeHistoryHostResponse{
		NumberOfShards: common.Int32Ptr(int32(h.controller.NumShards())),
		ShardIDs:       h.controller.ShardIDs(),
		DomainCache: &gen.DomainCacheInfo{
			NumOfItemsInCacheByID:   &numOfItemsInCacheByID,
			NumOfItemsInCacheByName: &numOfItemsInCacheByName,
		},
		ShardControllerStatus: &status,
		Address:               common.StringPtr(h.GetHostInfo().GetAddress()),
	}
	return resp, nil
}

// RemoveTask returns information about the internal states of a history host
func (h *handlerImpl) RemoveTask(
	ctx context.Context,
	request *gen.RemoveTaskRequest,
) (retError error) {
	executionMgr, err := h.GetExecutionManager(int(request.GetShardID()))
	if err != nil {
		return err
	}

	switch taskType := common.TaskType(request.GetType()); taskType {
	case common.TaskTypeTransfer:
		return executionMgr.CompleteTransferTask(ctx, &persistence.CompleteTransferTaskRequest{
			TaskID: request.GetTaskID(),
		})
	case common.TaskTypeTimer:
		return executionMgr.CompleteTimerTask(ctx, &persistence.CompleteTimerTaskRequest{
			VisibilityTimestamp: time.Unix(0, request.GetVisibilityTimestamp()),
			TaskID:              request.GetTaskID(),
		})
	case common.TaskTypeReplication:
		return executionMgr.CompleteReplicationTask(ctx, &persistence.CompleteReplicationTaskRequest{
			TaskID: request.GetTaskID(),
		})
	default:
		return errInvalidTaskType
	}
}

// CloseShard closes a shard hosted by this instance
func (h *handlerImpl) CloseShard(
	ctx context.Context,
	request *gen.CloseShardRequest,
) (retError error) {
	h.controller.RemoveEngineForShard(int(request.GetShardID()))
	return nil
}

// ResetQueue resets processing queue states
func (h *handlerImpl) ResetQueue(
	ctx context.Context,
	request *gen.ResetQueueRequest,
) (retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryResetQueueScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	engine, err := h.controller.GetEngineForShard(int(request.GetShardID()))
	if err != nil {
		return h.error(err, scope, "", "")
	}

	switch taskType := common.TaskType(request.GetType()); taskType {
	case common.TaskTypeTransfer:
		err = engine.ResetTransferQueue(ctx, request.GetClusterName())
	case common.TaskTypeTimer:
		err = engine.ResetTimerQueue(ctx, request.GetClusterName())
	default:
		err = errInvalidTaskType
	}

	if err != nil {
		return h.error(err, scope, "", "")
	}
	return nil
}

// DescribeQueue describes processing queue states
func (h *handlerImpl) DescribeQueue(
	ctx context.Context,
	request *gen.DescribeQueueRequest,
) (resp *gen.DescribeQueueResponse, retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryDescribeQueueScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	engine, err := h.controller.GetEngineForShard(int(request.GetShardID()))
	if err != nil {
		return nil, h.error(err, scope, "", "")
	}

	switch taskType := common.TaskType(request.GetType()); taskType {
	case common.TaskTypeTransfer:
		resp, err = engine.DescribeTransferQueue(ctx, request.GetClusterName())
	case common.TaskTypeTimer:
		resp, err = engine.DescribeTimerQueue(ctx, request.GetClusterName())
	default:
		err = errInvalidTaskType
	}

	if err != nil {
		return nil, h.error(err, scope, "", "")
	}
	return resp, nil
}

// DescribeMutableState - returns the internal analysis of workflow execution state
func (h *handlerImpl) DescribeMutableState(
	ctx context.Context,
	request *hist.DescribeMutableStateRequest,
) (resp *hist.DescribeMutableStateResponse, retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryDescribeMutabelStateScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	domainID := request.GetDomainUUID()
	if domainID == "" {
		return nil, h.error(errDomainNotSet, scope, domainID, "")
	}

	workflowExecution := request.Execution
	workflowID := workflowExecution.GetWorkflowId()
	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return nil, h.error(err1, scope, domainID, workflowID)
	}

	resp, err2 := engine.DescribeMutableState(ctx, request)
	if err2 != nil {
		return nil, h.error(err2, scope, domainID, workflowID)
	}
	return resp, nil
}

// GetMutableState - returns the id of the next event in the execution's history
func (h *handlerImpl) GetMutableState(
	ctx context.Context,
	getRequest *hist.GetMutableStateRequest,
) (resp *hist.GetMutableStateResponse, retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryGetMutableStateScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	domainID := getRequest.GetDomainUUID()
	if domainID == "" {
		return nil, h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return nil, h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	workflowExecution := getRequest.Execution
	workflowID := workflowExecution.GetWorkflowId()
	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return nil, h.error(err1, scope, domainID, workflowID)
	}

	resp, err2 := engine.GetMutableState(ctx, getRequest)
	if err2 != nil {
		return nil, h.error(err2, scope, domainID, workflowID)
	}
	return resp, nil
}

// PollMutableState - returns the id of the next event in the execution's history
func (h *handlerImpl) PollMutableState(
	ctx context.Context,
	getRequest *hist.PollMutableStateRequest,
) (resp *hist.PollMutableStateResponse, retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryPollMutableStateScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	domainID := getRequest.GetDomainUUID()
	if domainID == "" {
		return nil, h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return nil, h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	workflowExecution := getRequest.Execution
	workflowID := workflowExecution.GetWorkflowId()
	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return nil, h.error(err1, scope, domainID, workflowID)
	}

	resp, err2 := engine.PollMutableState(ctx, getRequest)
	if err2 != nil {
		return nil, h.error(err2, scope, domainID, workflowID)
	}
	return resp, nil
}

// DescribeWorkflowExecution returns information about the specified workflow execution.
func (h *handlerImpl) DescribeWorkflowExecution(
	ctx context.Context,
	request *hist.DescribeWorkflowExecutionRequest,
) (resp *gen.DescribeWorkflowExecutionResponse, retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryDescribeWorkflowExecutionScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	domainID := request.GetDomainUUID()
	if domainID == "" {
		return nil, h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return nil, h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	workflowExecution := request.Request.Execution
	workflowID := workflowExecution.GetWorkflowId()
	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return nil, h.error(err1, scope, domainID, workflowID)
	}

	resp, err2 := engine.DescribeWorkflowExecution(ctx, request)
	if err2 != nil {
		return nil, h.error(err2, scope, domainID, workflowID)
	}
	return resp, nil
}

// RequestCancelWorkflowExecution - requests cancellation of a workflow
func (h *handlerImpl) RequestCancelWorkflowExecution(
	ctx context.Context,
	request *hist.RequestCancelWorkflowExecutionRequest,
) (retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryRequestCancelWorkflowExecutionScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return errShuttingDown
	}

	domainID := request.GetDomainUUID()
	if domainID == "" || request.CancelRequest.GetDomain() == "" {
		return h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	cancelRequest := request.CancelRequest
	h.GetLogger().Debug(fmt.Sprintf("RequestCancelWorkflowExecution. DomainID: %v/%v, WorkflowID: %v, RunID: %v.",
		cancelRequest.GetDomain(),
		request.GetDomainUUID(),
		cancelRequest.WorkflowExecution.GetWorkflowId(),
		cancelRequest.WorkflowExecution.GetRunId()))

	workflowID := cancelRequest.WorkflowExecution.GetWorkflowId()
	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return h.error(err1, scope, domainID, workflowID)
	}

	err2 := engine.RequestCancelWorkflowExecution(ctx, request)
	if err2 != nil {
		return h.error(err2, scope, domainID, workflowID)
	}

	return nil
}

// SignalWorkflowExecution is used to send a signal event to running workflow execution.  This results in
// WorkflowExecutionSignaled event recorded in the history and a decision task being created for the execution.
func (h *handlerImpl) SignalWorkflowExecution(
	ctx context.Context,
	wrappedRequest *hist.SignalWorkflowExecutionRequest,
) (retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistorySignalWorkflowExecutionScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return errShuttingDown
	}

	domainID := wrappedRequest.GetDomainUUID()
	if domainID == "" {
		return h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	workflowExecution := wrappedRequest.SignalRequest.WorkflowExecution
	workflowID := workflowExecution.GetWorkflowId()
	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return h.error(err1, scope, domainID, workflowID)
	}

	err2 := engine.SignalWorkflowExecution(ctx, wrappedRequest)
	if err2 != nil {
		return h.error(err2, scope, domainID, workflowID)
	}

	return nil
}

// SignalWithStartWorkflowExecution is used to ensure sending a signal event to a workflow execution.
// If workflow is running, this results in WorkflowExecutionSignaled event recorded in the history
// and a decision task being created for the execution.
// If workflow is not running or not found, this results in WorkflowExecutionStarted and WorkflowExecutionSignaled
// event recorded in history, and a decision task being created for the execution
func (h *handlerImpl) SignalWithStartWorkflowExecution(
	ctx context.Context,
	wrappedRequest *hist.SignalWithStartWorkflowExecutionRequest,
) (resp *gen.StartWorkflowExecutionResponse, retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistorySignalWithStartWorkflowExecutionScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return nil, errShuttingDown
	}

	domainID := wrappedRequest.GetDomainUUID()
	if domainID == "" {
		return nil, h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return nil, h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	signalWithStartRequest := wrappedRequest.SignalWithStartRequest
	workflowID := signalWithStartRequest.GetWorkflowId()
	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return nil, h.error(err1, scope, domainID, workflowID)
	}

	resp, err2 := engine.SignalWithStartWorkflowExecution(ctx, wrappedRequest)
	if err2 != nil {
		return nil, h.error(err2, scope, domainID, workflowID)
	}

	return resp, nil
}

// RemoveSignalMutableState is used to remove a signal request ID that was previously recorded.  This is currently
// used to clean execution info when signal decision finished.
func (h *handlerImpl) RemoveSignalMutableState(
	ctx context.Context,
	wrappedRequest *hist.RemoveSignalMutableStateRequest,
) (retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryRemoveSignalMutableStateScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return errShuttingDown
	}

	domainID := wrappedRequest.GetDomainUUID()
	if domainID == "" {
		return h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	workflowExecution := wrappedRequest.WorkflowExecution
	workflowID := workflowExecution.GetWorkflowId()
	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return h.error(err1, scope, domainID, workflowID)
	}

	err2 := engine.RemoveSignalMutableState(ctx, wrappedRequest)
	if err2 != nil {
		return h.error(err2, scope, domainID, workflowID)
	}

	return nil
}

// TerminateWorkflowExecution terminates an existing workflow execution by recording WorkflowExecutionTerminated event
// in the history and immediately terminating the execution instance.
func (h *handlerImpl) TerminateWorkflowExecution(
	ctx context.Context,
	wrappedRequest *hist.TerminateWorkflowExecutionRequest,
) (retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryTerminateWorkflowExecutionScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return errShuttingDown
	}

	domainID := wrappedRequest.GetDomainUUID()
	if domainID == "" {
		return h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	workflowExecution := wrappedRequest.TerminateRequest.WorkflowExecution
	workflowID := workflowExecution.GetWorkflowId()
	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return h.error(err1, scope, domainID, workflowID)
	}

	err2 := engine.TerminateWorkflowExecution(ctx, wrappedRequest)
	if err2 != nil {
		return h.error(err2, scope, domainID, workflowID)
	}

	return nil
}

// ResetWorkflowExecution reset an existing workflow execution
// in the history and immediately terminating the execution instance.
func (h *handlerImpl) ResetWorkflowExecution(
	ctx context.Context,
	wrappedRequest *hist.ResetWorkflowExecutionRequest,
) (resp *gen.ResetWorkflowExecutionResponse, retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryResetWorkflowExecutionScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return nil, errShuttingDown
	}

	domainID := wrappedRequest.GetDomainUUID()
	if domainID == "" {
		return nil, h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return nil, h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	workflowExecution := wrappedRequest.ResetRequest.WorkflowExecution
	workflowID := workflowExecution.GetWorkflowId()
	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return nil, h.error(err1, scope, domainID, workflowID)
	}

	resp, err2 := engine.ResetWorkflowExecution(ctx, wrappedRequest)
	if err2 != nil {
		return nil, h.error(err2, scope, domainID, workflowID)
	}

	return resp, nil
}

// QueryWorkflow queries a workflow.
func (h *handlerImpl) QueryWorkflow(
	ctx context.Context,
	request *hist.QueryWorkflowRequest,
) (resp *hist.QueryWorkflowResponse, retError error) {
	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryQueryWorkflowScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return nil, errShuttingDown
	}

	domainID := request.GetDomainUUID()
	if domainID == "" {
		return nil, h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return nil, h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	workflowID := request.GetRequest().GetExecution().GetWorkflowId()
	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return nil, h.error(err1, scope, domainID, workflowID)
	}

	resp, err2 := engine.QueryWorkflow(ctx, request)
	if err2 != nil {
		return nil, h.error(err2, scope, domainID, workflowID)
	}

	return resp, nil
}

// ScheduleDecisionTask is used for creating a decision task for already started workflow execution.  This is mainly
// used by transfer queue processor during the processing of StartChildWorkflowExecution task, where it first starts
// child execution without creating the decision task and then calls this API after updating the mutable state of
// parent execution.
func (h *handlerImpl) ScheduleDecisionTask(
	ctx context.Context,
	request *hist.ScheduleDecisionTaskRequest,
) (retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryScheduleDecisionTaskScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return errShuttingDown
	}

	domainID := request.GetDomainUUID()
	if domainID == "" {
		return h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	if request.WorkflowExecution == nil {
		return h.error(errWorkflowExecutionNotSet, scope, domainID, "")
	}

	workflowExecution := request.WorkflowExecution
	workflowID := workflowExecution.GetWorkflowId()
	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return h.error(err1, scope, domainID, workflowID)
	}

	err2 := engine.ScheduleDecisionTask(ctx, request)
	if err2 != nil {
		return h.error(err2, scope, domainID, workflowID)
	}

	return nil
}

// RecordChildExecutionCompleted is used for reporting the completion of child workflow execution to parent.
// This is mainly called by transfer queue processor during the processing of DeleteExecution task.
func (h *handlerImpl) RecordChildExecutionCompleted(
	ctx context.Context,
	request *hist.RecordChildExecutionCompletedRequest,
) (retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryRecordChildExecutionCompletedScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return errShuttingDown
	}

	domainID := request.GetDomainUUID()
	if domainID == "" {
		return h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	if request.WorkflowExecution == nil {
		return h.error(errWorkflowExecutionNotSet, scope, domainID, "")
	}

	workflowExecution := request.WorkflowExecution
	workflowID := workflowExecution.GetWorkflowId()
	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return h.error(err1, scope, domainID, workflowID)
	}

	err2 := engine.RecordChildExecutionCompleted(ctx, request)
	if err2 != nil {
		return h.error(err2, scope, domainID, workflowID)
	}

	return nil
}

// ResetStickyTaskList reset the volatile information in mutable state of a given workflow.
// Volatile information are the information related to client, such as:
// 1. StickyTaskList
// 2. StickyScheduleToStartTimeout
// 3. ClientLibraryVersion
// 4. ClientFeatureVersion
// 5. ClientImpl
func (h *handlerImpl) ResetStickyTaskList(
	ctx context.Context,
	resetRequest *hist.ResetStickyTaskListRequest,
) (resp *hist.ResetStickyTaskListResponse, retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryResetStickyTaskListScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return nil, errShuttingDown
	}

	domainID := resetRequest.GetDomainUUID()
	if domainID == "" {
		return nil, h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return nil, h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	workflowID := resetRequest.Execution.GetWorkflowId()
	engine, err := h.controller.GetEngine(workflowID)
	if err != nil {
		return nil, h.error(err, scope, domainID, workflowID)
	}

	resp, err = engine.ResetStickyTaskList(ctx, resetRequest)
	if err != nil {
		return nil, h.error(err, scope, domainID, workflowID)
	}

	return resp, nil
}

// ReplicateEventsV2 is called by processor to replicate history events for passive domains
func (h *handlerImpl) ReplicateEventsV2(
	ctx context.Context,
	replicateRequest *hist.ReplicateEventsV2Request,
) (retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	if h.isShuttingDown() {
		return errShuttingDown
	}

	scope := metrics.HistoryReplicateEventsV2Scope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	domainID := replicateRequest.GetDomainUUID()
	if domainID == "" {
		return h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	workflowExecution := replicateRequest.WorkflowExecution
	workflowID := workflowExecution.GetWorkflowId()
	engine, err1 := h.controller.GetEngine(workflowID)
	if err1 != nil {
		return h.error(err1, scope, domainID, workflowID)
	}

	err2 := engine.ReplicateEventsV2(ctx, replicateRequest)
	if err2 != nil {
		return h.error(err2, scope, domainID, workflowID)
	}

	return nil
}

// SyncShardStatus is called by processor to sync history shard information from another cluster
func (h *handlerImpl) SyncShardStatus(
	ctx context.Context,
	syncShardStatusRequest *hist.SyncShardStatusRequest,
) (retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistorySyncShardStatusScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return errShuttingDown
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return h.error(errHistoryHostThrottle, scope, "", "")
	}

	if syncShardStatusRequest.SourceCluster == nil {
		return h.error(errSourceClusterNotSet, scope, "", "")
	}

	if syncShardStatusRequest.ShardId == nil {
		return h.error(errShardIDNotSet, scope, "", "")
	}

	if syncShardStatusRequest.Timestamp == nil {
		return h.error(errTimestampNotSet, scope, "", "")
	}

	// shard ID is already provided in the request
	engine, err := h.controller.GetEngineForShard(int(syncShardStatusRequest.GetShardId()))
	if err != nil {
		return h.error(err, scope, "", "")
	}

	err = engine.SyncShardStatus(ctx, syncShardStatusRequest)
	if err != nil {
		return h.error(err, scope, "", "")
	}

	return nil
}

// SyncActivity is called by processor to sync activity
func (h *handlerImpl) SyncActivity(
	ctx context.Context,
	syncActivityRequest *hist.SyncActivityRequest,
) (retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistorySyncActivityScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return errShuttingDown
	}

	domainID := syncActivityRequest.GetDomainId()
	if syncActivityRequest.DomainId == nil || uuid.Parse(syncActivityRequest.GetDomainId()) == nil {
		return h.error(errDomainNotSet, scope, domainID, "")
	}

	if ok := h.rateLimiter.Allow(); !ok {
		return h.error(errHistoryHostThrottle, scope, domainID, "")
	}

	if syncActivityRequest.WorkflowId == nil {
		return h.error(errWorkflowIDNotSet, scope, domainID, "")
	}

	if syncActivityRequest.RunId == nil || uuid.Parse(syncActivityRequest.GetRunId()) == nil {
		return h.error(errRunIDNotValid, scope, domainID, "")
	}

	workflowID := syncActivityRequest.GetWorkflowId()
	engine, err := h.controller.GetEngine(workflowID)
	if err != nil {
		return h.error(err, scope, domainID, workflowID)
	}

	err = engine.SyncActivity(ctx, syncActivityRequest)
	if err != nil {
		return h.error(err, scope, domainID, workflowID)
	}

	return nil
}

// GetReplicationMessages is called by remote peers to get replicated messages for cross DC replication
func (h *handlerImpl) GetReplicationMessages(
	ctx context.Context,
	request *r.GetReplicationMessagesRequest,
) (resp *r.GetReplicationMessagesResponse, retError error) {
	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	h.GetLogger().Debug("Received GetReplicationMessages call.")

	scope := metrics.HistoryGetReplicationMessagesScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return nil, errShuttingDown
	}

	var wg sync.WaitGroup
	wg.Add(len(request.Tokens))
	result := new(sync.Map)

	for _, token := range request.Tokens {
		go func(token *r.ReplicationToken) {
			defer wg.Done()

			engine, err := h.controller.GetEngineForShard(int(token.GetShardID()))
			if err != nil {
				h.GetLogger().Warn("History engine not found for shard", tag.Error(err))
				return
			}
			tasks, err := engine.GetReplicationMessages(
				ctx,
				request.GetClusterName(),
				token.GetLastRetrievedMessageId(),
			)
			if err != nil {
				h.GetLogger().Warn("Failed to get replication tasks for shard", tag.Error(err))
				return
			}

			result.Store(token.GetShardID(), tasks)
		}(token)
	}

	wg.Wait()

	messagesByShard := make(map[int32]*r.ReplicationMessages)
	result.Range(func(key, value interface{}) bool {
		shardID := key.(int32)
		tasks := value.(*r.ReplicationMessages)
		messagesByShard[shardID] = tasks
		return true
	})

	h.GetLogger().Debug("GetReplicationMessages succeeded.")

	return &r.GetReplicationMessagesResponse{MessagesByShard: messagesByShard}, nil
}

// GetDLQReplicationMessages is called by remote peers to get replicated messages for DLQ merging
func (h *handlerImpl) GetDLQReplicationMessages(
	ctx context.Context,
	request *r.GetDLQReplicationMessagesRequest,
) (resp *r.GetDLQReplicationMessagesResponse, retError error) {
	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryGetDLQReplicationMessagesScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return nil, errShuttingDown
	}

	taskInfoPerExecution := map[definition.WorkflowIdentifier][]*r.ReplicationTaskInfo{}
	// do batch based on workflow ID and run ID
	for _, taskInfo := range request.GetTaskInfos() {
		identity := definition.NewWorkflowIdentifier(
			taskInfo.GetDomainID(),
			taskInfo.GetWorkflowID(),
			taskInfo.GetRunID(),
		)
		if _, ok := taskInfoPerExecution[identity]; !ok {
			taskInfoPerExecution[identity] = []*r.ReplicationTaskInfo{}
		}
		taskInfoPerExecution[identity] = append(taskInfoPerExecution[identity], taskInfo)
	}

	var wg sync.WaitGroup
	wg.Add(len(taskInfoPerExecution))
	tasksChan := make(chan *r.ReplicationTask, len(request.GetTaskInfos()))
	handleTaskInfoPerExecution := func(taskInfos []*r.ReplicationTaskInfo) {
		defer wg.Done()
		if len(taskInfos) == 0 {
			return
		}

		engine, err := h.controller.GetEngine(
			taskInfos[0].GetWorkflowID(),
		)
		if err != nil {
			h.GetLogger().Warn("History engine not found for workflow ID.", tag.Error(err))
			return
		}

		tasks, err := engine.GetDLQReplicationMessages(
			ctx,
			taskInfos,
		)
		if err != nil {
			h.GetLogger().Error("Failed to get dlq replication tasks.", tag.Error(err))
			return
		}

		for _, task := range tasks {
			tasksChan <- task
		}
	}

	for _, replicationTaskInfos := range taskInfoPerExecution {
		go handleTaskInfoPerExecution(replicationTaskInfos)
	}
	wg.Wait()
	close(tasksChan)

	replicationTasks := make([]*r.ReplicationTask, 0, len(tasksChan))
	for task := range tasksChan {
		replicationTasks = append(replicationTasks, task)
	}
	return &r.GetDLQReplicationMessagesResponse{
		ReplicationTasks: replicationTasks,
	}, nil
}

// ReapplyEvents applies stale events to the current workflow and the current run
func (h *handlerImpl) ReapplyEvents(
	ctx context.Context,
	request *hist.ReapplyEventsRequest,
) (retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryReapplyEventsScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return errShuttingDown
	}

	domainID := request.GetDomainUUID()
	workflowID := request.GetRequest().GetWorkflowExecution().GetWorkflowId()
	engine, err := h.controller.GetEngine(workflowID)
	if err != nil {
		return h.error(err, scope, domainID, workflowID)
	}
	// deserialize history event object
	historyEvents, err := h.GetPayloadSerializer().DeserializeBatchEvents(&persistence.DataBlob{
		Encoding: common.EncodingTypeThriftRW,
		Data:     request.GetRequest().GetEvents().GetData(),
	})
	if err != nil {
		return h.error(err, scope, domainID, workflowID)
	}

	execution := request.GetRequest().GetWorkflowExecution()
	if err := engine.ReapplyEvents(
		ctx,
		request.GetDomainUUID(),
		execution.GetWorkflowId(),
		execution.GetRunId(),
		thrift.FromHistoryEventArray(historyEvents),
	); err != nil {
		return h.error(err, scope, domainID, workflowID)
	}
	return nil
}

// ReadDLQMessages reads replication DLQ messages
func (h *handlerImpl) ReadDLQMessages(
	ctx context.Context,
	request *r.ReadDLQMessagesRequest,
) (resp *r.ReadDLQMessagesResponse, retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryReadDLQMessagesScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return nil, errShuttingDown
	}

	engine, err := h.controller.GetEngineForShard(int(request.GetShardID()))
	if err != nil {
		return nil, h.error(err, scope, "", "")
	}

	return engine.ReadDLQMessages(ctx, request)
}

// PurgeDLQMessages deletes replication DLQ messages
func (h *handlerImpl) PurgeDLQMessages(
	ctx context.Context,
	request *r.PurgeDLQMessagesRequest,
) (retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	scope := metrics.HistoryPurgeDLQMessagesScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return errShuttingDown
	}

	engine, err := h.controller.GetEngineForShard(int(request.GetShardID()))
	if err != nil {
		return h.error(err, scope, "", "")
	}

	return engine.PurgeDLQMessages(ctx, request)
}

// MergeDLQMessages reads and applies replication DLQ messages
func (h *handlerImpl) MergeDLQMessages(
	ctx context.Context,
	request *r.MergeDLQMessagesRequest,
) (resp *r.MergeDLQMessagesResponse, retError error) {

	defer log.CapturePanic(h.GetLogger(), &retError)
	h.startWG.Wait()

	if h.isShuttingDown() {
		return nil, errShuttingDown
	}

	scope := metrics.HistoryMergeDLQMessagesScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	engine, err := h.controller.GetEngineForShard(int(request.GetShardID()))
	if err != nil {
		return nil, h.error(err, scope, "", "")
	}

	return engine.MergeDLQMessages(ctx, request)
}

// RefreshWorkflowTasks refreshes all the tasks of a workflow
func (h *handlerImpl) RefreshWorkflowTasks(
	ctx context.Context,
	request *hist.RefreshWorkflowTasksRequest) (retError error) {

	scope := metrics.HistoryRefreshWorkflowTasksScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	if h.isShuttingDown() {
		return errShuttingDown
	}

	domainID := request.GetDomainUIID()
	execution := request.GetRequest().GetExecution()
	workflowID := execution.GetWorkflowId()
	engine, err := h.controller.GetEngine(workflowID)
	if err != nil {
		return h.error(err, scope, domainID, workflowID)
	}

	err = engine.RefreshWorkflowTasks(
		ctx,
		domainID,
		types.WorkflowExecution{
			WorkflowID: execution.WorkflowId,
			RunID:      execution.RunId,
		},
	)

	if err != nil {
		return h.error(err, scope, domainID, workflowID)
	}

	return nil
}

// NotifyFailoverMarkers sends the failover markers to failover coordinator.
// The coordinator decides when the failover finishes based on received failover marker.
func (h *handlerImpl) NotifyFailoverMarkers(
	ctx context.Context,
	request *types.NotifyFailoverMarkersRequest,
) (retError error) {

	scope := metrics.HistoryNotifyFailoverMarkersScope
	h.GetMetricsClient().IncCounter(scope, metrics.CadenceRequests)
	sw := h.GetMetricsClient().StartTimer(scope, metrics.CadenceLatency)
	defer sw.Stop()

	for _, token := range request.GetFailoverMarkerTokens() {
		marker := token.GetFailoverMarker()
		h.GetLogger().Debug("Handling failover maker", tag.WorkflowDomainID(marker.GetDomainID()))
		h.failoverCoordinator.ReceiveFailoverMarkers(token.GetShardIDs(), token.GetFailoverMarker())
	}
	return nil
}

// convertError is a helper method to convert ShardOwnershipLostError from persistence layer returned by various
// HistoryEngine API calls to ShardOwnershipLost error return by HistoryService for client to be redirected to the
// correct shard.
func (h *handlerImpl) convertError(err error) error {
	switch err.(type) {
	case *persistence.ShardOwnershipLostError:
		shardID := err.(*persistence.ShardOwnershipLostError).ShardID
		info, err := h.GetHistoryServiceResolver().Lookup(string(shardID))
		if err == nil {
			return shard.CreateShardOwnershipLostError(h.GetHostInfo().GetAddress(), info.GetAddress())
		}
		return shard.CreateShardOwnershipLostError(h.GetHostInfo().GetAddress(), "")
	case *persistence.WorkflowExecutionAlreadyStartedError:
		err := err.(*persistence.WorkflowExecutionAlreadyStartedError)
		return &types.InternalServiceError{Message: err.Msg}
	case *persistence.CurrentWorkflowConditionFailedError:
		err := err.(*persistence.CurrentWorkflowConditionFailedError)
		return &types.InternalServiceError{Message: err.Msg}
	case *persistence.TransactionSizeLimitError:
		err := err.(*persistence.TransactionSizeLimitError)
		return &types.BadRequestError{Message: err.Msg}
	}

	return err
}

func (h *handlerImpl) updateErrorMetric(
	scope int,
	domainID string,
	workflowID string,
	err error,
) {

	if err == context.DeadlineExceeded || err == context.Canceled {
		h.GetMetricsClient().IncCounter(scope, metrics.CadenceErrContextTimeoutCounter)
		return
	}

	switch err := err.(type) {
	case *types.ShardOwnershipLostError:
		h.GetMetricsClient().IncCounter(scope, metrics.CadenceErrShardOwnershipLostCounter)
	case *types.EventAlreadyStartedError:
		h.GetMetricsClient().IncCounter(scope, metrics.CadenceErrEventAlreadyStartedCounter)
	case *types.BadRequestError:
		h.GetMetricsClient().IncCounter(scope, metrics.CadenceErrBadRequestCounter)
	case *types.DomainNotActiveError:
		h.GetMetricsClient().IncCounter(scope, metrics.CadenceErrBadRequestCounter)
	case *types.WorkflowExecutionAlreadyStartedError:
		h.GetMetricsClient().IncCounter(scope, metrics.CadenceErrExecutionAlreadyStartedCounter)
	case *types.EntityNotExistsError:
		h.GetMetricsClient().IncCounter(scope, metrics.CadenceErrEntityNotExistsCounter)
	case *types.CancellationAlreadyRequestedError:
		h.GetMetricsClient().IncCounter(scope, metrics.CadenceErrCancellationAlreadyRequestedCounter)
	case *types.LimitExceededError:
		h.GetMetricsClient().IncCounter(scope, metrics.CadenceErrLimitExceededCounter)
	case *types.RetryTaskV2Error:
		h.GetMetricsClient().IncCounter(scope, metrics.CadenceErrRetryTaskCounter)
	case *types.ServiceBusyError:
		h.GetMetricsClient().IncCounter(scope, metrics.CadenceErrServiceBusyCounter)
	case *yarpcerrors.Status:
		if err.Code() == yarpcerrors.CodeDeadlineExceeded {
			h.GetMetricsClient().IncCounter(scope, metrics.CadenceErrContextTimeoutCounter)
		}
		h.GetMetricsClient().IncCounter(scope, metrics.CadenceFailures)
	case *types.InternalServiceError:
		h.GetMetricsClient().IncCounter(scope, metrics.CadenceFailures)
		h.GetLogger().Error("Internal service error",
			tag.Error(err),
			tag.WorkflowID(workflowID),
			tag.WorkflowDomainID(domainID))
	default:
		h.GetMetricsClient().IncCounter(scope, metrics.CadenceFailures)
		h.getLoggerWithTags(domainID, workflowID).Error("Uncategorized error", tag.Error(err))
	}
}

func (h *handlerImpl) error(
	err error,
	scope int,
	domainID string,
	workflowID string,
) error {

	err = h.convertError(err)
	h.updateErrorMetric(scope, domainID, workflowID, err)

	return err
}

func (h *handlerImpl) getLoggerWithTags(
	domainID string,
	workflowID string,
) log.Logger {

	logger := h.GetLogger()
	if domainID != "" {
		logger = logger.WithTags(tag.WorkflowDomainID(domainID))
	}

	if workflowID != "" {
		logger = logger.WithTags(tag.WorkflowID(workflowID))
	}

	return logger
}

func validateTaskToken(token *common.TaskToken) error {
	if token.WorkflowID == "" {
		return errWorkflowIDNotSet
	}
	if token.RunID != "" && uuid.Parse(token.RunID) == nil {
		return errRunIDNotValid
	}
	return nil
}
