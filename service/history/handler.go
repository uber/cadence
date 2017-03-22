package history

import (
	"fmt"
	"log"
	"sync"

	hist "github.com/uber/cadence/.gen/go/history"
	gen "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/client/matching"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/membership"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service"
	"github.com/uber/tchannel-go/thrift"
)

// Handler - Thrift handler inteface for history service
type Handler struct {
	numberOfShards        int
	shardManager          persistence.ShardManager
	executionMgrFactory   persistence.ExecutionManagerFactory
	matchingServiceClient matching.Client
	hServiceResolver      membership.ServiceResolver
	controller            *shardController
	tokenSerializer       common.TaskTokenSerializer
	startWG               sync.WaitGroup
	metricsClient         metrics.Client
	service.Service
}

var _ hist.TChanHistoryService = (*Handler)(nil)
var _ EngineFactory = (*Handler)(nil)

// NewHandler creates a thrift handler for the history service
func NewHandler(sVice service.Service, shardManager persistence.ShardManager,
	executionMgrFactory persistence.ExecutionManagerFactory, numberOfShards int) (*Handler, []thrift.TChanServer) {
	handler := &Handler{
		Service:             sVice,
		shardManager:        shardManager,
		executionMgrFactory: executionMgrFactory,
		numberOfShards:      numberOfShards,
		tokenSerializer:     common.NewJSONTaskTokenSerializer(),
	}
	// prevent us from trying to serve requests before shard controller is started and ready
	handler.startWG.Add(1)
	return handler, []thrift.TChanServer{hist.NewTChanHistoryServiceServer(handler)}
}

// Start starts the handler
func (h *Handler) Start(thriftService []thrift.TChanServer) error {
	h.Service.Start(thriftService)
	matchingServiceClient, err0 := h.Service.GetClientFactory().NewMatchingClient()
	if err0 != nil {
		return err0
	}
	h.matchingServiceClient = matchingServiceClient

	hServiceResolver, err1 := h.GetMembershipMonitor().GetResolver(common.HistoryServiceName)
	if err1 != nil {
		h.Service.GetLogger().Fatalf("Unable to get history service resolver.")
	}
	h.hServiceResolver = hServiceResolver
	h.controller = newShardController(h.numberOfShards, h.GetHostInfo(), hServiceResolver, h.shardManager,
		h.executionMgrFactory, h, h.GetLogger(), h.GetMetricsClient())
	h.controller.Start()
	h.metricsClient = h.GetMetricsClient()
	h.startWG.Done()
	return nil
}

// Stop stops the handler
func (h *Handler) Stop() {
	h.controller.Stop()
	h.Service.Stop()
}

// CreateEngine is implementation for HistoryEngineFactory used for creating the engine instance for shard
func (h *Handler) CreateEngine(context ShardContext) Engine {
	return NewEngineWithShardContext(context, h.matchingServiceClient)
}

// IsHealthy - Health endpoint.
func (h *Handler) IsHealthy(ctx thrift.Context) (bool, error) {
	log.Println("Workflow Health endpoint reached.")
	return true, nil
}

// RecordActivityTaskHeartbeat - Record Activity Task Heart beat.
func (h *Handler) RecordActivityTaskHeartbeat(ctx thrift.Context,
	heartbeatRequest *gen.RecordActivityTaskHeartbeatRequest) (*gen.RecordActivityTaskHeartbeatResponse, error) {
	h.startWG.Wait()

	h.metricsClient.IncCounter(metrics.HistoryRecordActivityTaskHeartbeatScope, metrics.CadenceRequests)
	sw := h.metricsClient.StartTimer(metrics.HistoryRecordActivityTaskHeartbeatScope, metrics.CadenceLatency)
	defer sw.Stop()

	token, err0 := h.tokenSerializer.Deserialize(heartbeatRequest.GetTaskToken())
	if err0 != nil {
		err0 = &gen.BadRequestError{Message: fmt.Sprintf("Error deserializing task token. Error: %v", err0)}
		h.updateErrorMetric(metrics.HistoryRecordActivityTaskHeartbeatScope, err0)
		return nil, err0
	}

	engine, err1 := h.controller.GetEngine(token.WorkflowID)
	if err1 != nil {
		h.updateErrorMetric(metrics.HistoryRecordActivityTaskHeartbeatScope, err1)
		return nil, err1
	}

	response, err2 := engine.RecordActivityTaskHeartbeat(heartbeatRequest)
	if err2 != nil {
		h.updateErrorMetric(metrics.HistoryRecordActivityTaskHeartbeatScope, h.convertError(err2))
		return nil, h.convertError(err2)
	}

	return response, nil
}

// RecordActivityTaskStarted - Record Activity Task started.
func (h *Handler) RecordActivityTaskStarted(ctx thrift.Context,
	recordRequest *hist.RecordActivityTaskStartedRequest) (*hist.RecordActivityTaskStartedResponse, error) {
	h.startWG.Wait()

	h.metricsClient.IncCounter(metrics.HistoryRecordActivityTaskStartedScope, metrics.CadenceRequests)
	sw := h.metricsClient.StartTimer(metrics.HistoryRecordActivityTaskStartedScope, metrics.CadenceLatency)
	defer sw.Stop()

	workflowExecution := recordRequest.GetWorkflowExecution()
	engine, err1 := h.controller.GetEngine(workflowExecution.GetWorkflowId())
	if err1 != nil {
		h.updateErrorMetric(metrics.HistoryRecordActivityTaskStartedScope, err1)
		return nil, err1
	}

	response, err2 := engine.RecordActivityTaskStarted(recordRequest)
	if err2 != nil {
		h.updateErrorMetric(metrics.HistoryRecordActivityTaskStartedScope, h.convertError(err2))
		return nil, h.convertError(err2)
	}

	return response, nil
}

// RecordDecisionTaskStarted - Record Decision Task started.
func (h *Handler) RecordDecisionTaskStarted(ctx thrift.Context,
	recordRequest *hist.RecordDecisionTaskStartedRequest) (*hist.RecordDecisionTaskStartedResponse, error) {
	h.startWG.Wait()
	h.Service.GetLogger().Debugf("RecordDecisionTaskStarted. WorkflowID: %v, RunID: %v, ScheduleID: %v",
		recordRequest.GetWorkflowExecution().GetWorkflowId(),
		recordRequest.GetWorkflowExecution().GetRunId(),
		recordRequest.GetScheduleId())

	h.metricsClient.IncCounter(metrics.HistoryRecordDecisionTaskStartedScope, metrics.CadenceRequests)
	sw := h.metricsClient.StartTimer(metrics.HistoryRecordDecisionTaskStartedScope, metrics.CadenceLatency)
	defer sw.Stop()

	workflowExecution := recordRequest.GetWorkflowExecution()
	engine, err1 := h.controller.GetEngine(workflowExecution.GetWorkflowId())
	if err1 != nil {
		h.Service.GetLogger().Errorf("RecordDecisionTaskStarted failed. Error: %v. WorkflowID: %v, RunID: %v, ScheduleID: %v",
			err1,
			recordRequest.GetWorkflowExecution().GetWorkflowId(),
			recordRequest.GetWorkflowExecution().GetRunId(),
			recordRequest.GetScheduleId())
		h.updateErrorMetric(metrics.HistoryRecordDecisionTaskStartedScope, err1)
		return nil, err1
	}

	response, err2 := engine.RecordDecisionTaskStarted(recordRequest)
	if err2 != nil {
		h.updateErrorMetric(metrics.HistoryRecordDecisionTaskStartedScope, h.convertError(err2))
		return nil, h.convertError(err2)
	}

	return response, nil
}

// RespondActivityTaskCompleted - records completion of an activity task
func (h *Handler) RespondActivityTaskCompleted(ctx thrift.Context,
	completeRequest *gen.RespondActivityTaskCompletedRequest) error {
	h.startWG.Wait()

	h.metricsClient.IncCounter(metrics.HistoryRespondActivityTaskCompletedScope, metrics.CadenceRequests)
	sw := h.metricsClient.StartTimer(metrics.HistoryRespondActivityTaskCompletedScope, metrics.CadenceLatency)
	defer sw.Stop()

	token, err0 := h.tokenSerializer.Deserialize(completeRequest.GetTaskToken())
	if err0 != nil {
		err0 = &gen.BadRequestError{Message: fmt.Sprintf("Error deserializing task token. Error: %v", err0)}
		h.updateErrorMetric(metrics.HistoryRespondActivityTaskCompletedScope, err0)
		return err0
	}

	engine, err1 := h.controller.GetEngine(token.WorkflowID)
	if err1 != nil {
		h.updateErrorMetric(metrics.HistoryRespondActivityTaskCompletedScope, err1)
		return err1
	}

	err2 := engine.RespondActivityTaskCompleted(completeRequest)
	if err2 != nil {
		h.updateErrorMetric(metrics.HistoryRespondActivityTaskCompletedScope, h.convertError(err2))
		return h.convertError(err2)
	}

	return nil
}

// RespondActivityTaskFailed - records failure of an activity task
func (h *Handler) RespondActivityTaskFailed(ctx thrift.Context,
	failRequest *gen.RespondActivityTaskFailedRequest) error {
	h.startWG.Wait()

	h.metricsClient.IncCounter(metrics.HistoryRespondActivityTaskFailedScope, metrics.CadenceRequests)
	sw := h.metricsClient.StartTimer(metrics.HistoryRespondActivityTaskFailedScope, metrics.CadenceLatency)
	defer sw.Stop()

	token, err0 := h.tokenSerializer.Deserialize(failRequest.GetTaskToken())
	if err0 != nil {
		err0 = &gen.BadRequestError{Message: fmt.Sprintf("Error deserializing task token. Error: %v", err0)}
		h.updateErrorMetric(metrics.HistoryRespondActivityTaskFailedScope, err0)
		return err0
	}

	engine, err1 := h.controller.GetEngine(token.WorkflowID)
	if err1 != nil {
		h.updateErrorMetric(metrics.HistoryRespondActivityTaskFailedScope, err1)
		return err1
	}

	err2 := engine.RespondActivityTaskFailed(failRequest)
	if err2 != nil {
		h.updateErrorMetric(metrics.HistoryRespondActivityTaskFailedScope, h.convertError(err2))
		return h.convertError(err2)
	}

	return nil
}

// RespondActivityTaskCanceled - records failure of an activity task
func (h *Handler) RespondActivityTaskCanceled(ctx thrift.Context,
	cancelRequest *gen.RespondActivityTaskCanceledRequest) error {
	h.startWG.Wait()

	h.metricsClient.IncCounter(metrics.HistoryRespondActivityTaskCanceledScope, metrics.CadenceRequests)
	sw := h.metricsClient.StartTimer(metrics.HistoryRespondActivityTaskCanceledScope, metrics.CadenceLatency)
	defer sw.Stop()

	token, err0 := h.tokenSerializer.Deserialize(cancelRequest.GetTaskToken())
	if err0 != nil {
		err0 = &gen.BadRequestError{Message: fmt.Sprintf("Error deserializing task token. Error: %v", err0)}
		h.updateErrorMetric(metrics.HistoryRespondActivityTaskCanceledScope, err0)
		return err0
	}

	engine, err1 := h.controller.GetEngine(token.WorkflowID)
	if err1 != nil {
		h.updateErrorMetric(metrics.HistoryRespondActivityTaskCanceledScope, err1)
		return err1
	}

	err2 := engine.RespondActivityTaskCanceled(cancelRequest)
	if err2 != nil {
		h.updateErrorMetric(metrics.HistoryRespondActivityTaskCanceledScope, h.convertError(err2))
		return h.convertError(err2)
	}

	return nil
}

// RespondDecisionTaskCompleted - records completion of a decision task
func (h *Handler) RespondDecisionTaskCompleted(ctx thrift.Context,
	completeRequest *gen.RespondDecisionTaskCompletedRequest) error {
	h.startWG.Wait()

	h.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope, metrics.CadenceRequests)
	sw := h.metricsClient.StartTimer(metrics.HistoryRespondDecisionTaskCompletedScope, metrics.CadenceLatency)
	defer sw.Stop()

	token, err0 := h.tokenSerializer.Deserialize(completeRequest.GetTaskToken())
	if err0 != nil {
		err0 = &gen.BadRequestError{Message: fmt.Sprintf("Error deserializing task token. Error: %v", err0)}
		h.updateErrorMetric(metrics.HistoryRespondDecisionTaskCompletedScope, err0)
		return err0
	}

	h.Service.GetLogger().Debugf("RespondDecisionTaskCompleted. WorkflowID: %v, RunID: %v, ScheduleID: %v",
		token.WorkflowID,
		token.RunID,
		token.ScheduleID)

	engine, err1 := h.controller.GetEngine(token.WorkflowID)
	if err1 != nil {
		h.updateErrorMetric(metrics.HistoryRespondDecisionTaskCompletedScope, err1)
		return err1
	}

	err2 := engine.RespondDecisionTaskCompleted(completeRequest)
	if err2 != nil {
		h.updateErrorMetric(metrics.HistoryRespondDecisionTaskCompletedScope, h.convertError(err2))
		return h.convertError(err2)
	}

	return nil
}

// StartWorkflowExecution - creates a new workflow execution
func (h *Handler) StartWorkflowExecution(ctx thrift.Context,
	startRequest *gen.StartWorkflowExecutionRequest) (*gen.StartWorkflowExecutionResponse, error) {
	h.startWG.Wait()

	h.metricsClient.IncCounter(metrics.HistoryStartWorkflowExecutionScope, metrics.CadenceRequests)
	sw := h.metricsClient.StartTimer(metrics.HistoryStartWorkflowExecutionScope, metrics.CadenceLatency)
	defer sw.Stop()

	engine, err1 := h.controller.GetEngine(startRequest.GetWorkflowId())
	if err1 != nil {
		h.updateErrorMetric(metrics.HistoryStartWorkflowExecutionScope, err1)
		return nil, err1
	}

	response, err2 := engine.StartWorkflowExecution(startRequest)
	if err2 != nil {
		h.updateErrorMetric(metrics.HistoryStartWorkflowExecutionScope, h.convertError(err2))
		return nil, h.convertError(err2)
	}

	return response, nil
}

// GetWorkflowExecutionHistory - returns the complete history of a workflow execution
func (h *Handler) GetWorkflowExecutionHistory(ctx thrift.Context,
	getRequest *gen.GetWorkflowExecutionHistoryRequest) (*gen.GetWorkflowExecutionHistoryResponse, error) {
	h.startWG.Wait()

	h.metricsClient.IncCounter(metrics.HistoryGetWorkflowExecutionHistoryScope, metrics.CadenceRequests)
	sw := h.metricsClient.StartTimer(metrics.HistoryGetWorkflowExecutionHistoryScope, metrics.CadenceLatency)
	defer sw.Stop()

	workflowExecution := getRequest.GetExecution()
	engine, err1 := h.controller.GetEngine(workflowExecution.GetWorkflowId())
	if err1 != nil {
		h.updateErrorMetric(metrics.HistoryGetWorkflowExecutionHistoryScope, err1)
		return nil, err1
	}

	resp, err2 := engine.GetWorkflowExecutionHistory(getRequest)
	if err2 != nil {
		h.updateErrorMetric(metrics.HistoryGetWorkflowExecutionHistoryScope, h.convertError(err2))
		return nil, h.convertError(err2)
	}
	return resp, nil
}

// RequestCancelWorkflowExecution - requests cancellation of a workflow
func (h *Handler) RequestCancelWorkflowExecution(ctx thrift.Context,
	request *gen.RequestCancelWorkflowExecutionRequest) error {
	h.startWG.Wait()

	h.metricsClient.IncCounter(metrics.HistoryRequestCancelWorkflowExecutionScope, metrics.CadenceRequests)
	sw := h.metricsClient.StartTimer(metrics.HistoryRequestCancelWorkflowExecutionScope, metrics.CadenceLatency)
	defer sw.Stop()

	h.Service.GetLogger().Debugf("RequestCancelWorkflowExecution. WorkflowID: %v, RunID: %v.",
		request.GetWorkflowId(),
		request.GetRunId())

	engine, err1 := h.controller.GetEngine(request.GetWorkflowId())
	if err1 != nil {
		h.updateErrorMetric(metrics.HistoryRequestCancelWorkflowExecutionScope, err1)
		return err1
	}

	err2 := engine.RequestCancelWorkflowExecution(request)
	if err2 != nil {
		h.updateErrorMetric(metrics.HistoryRequestCancelWorkflowExecutionScope, h.convertError(err2))
		return h.convertError(err2)
	}

	return nil
}


// convertError is a helper method to convert ShardOwnershipLostError from persistence layer returned by various
// HistoryEngine API calls to ShardOwnershipLost error return by HistoryService for client to be redirected to the
// correct shard.
func (h *Handler) convertError(err error) error {
	switch err.(type) {
	case *persistence.ShardOwnershipLostError:
		shardID := err.(*persistence.ShardOwnershipLostError).ShardID
		info, err := h.hServiceResolver.Lookup(string(shardID))
		if err != nil {
			return createShardOwnershipLostError(h.GetHostInfo().GetAddress(), info.GetAddress())
		}
		return createShardOwnershipLostError(h.GetHostInfo().GetAddress(), "")
	}

	return err
}

func (h *Handler) updateErrorMetric(scope int, err error) {
	switch err.(type) {
	case *hist.ShardOwnershipLostError:
		h.metricsClient.IncCounter(scope, metrics.CadenceErrShardOwnershipLostCounter)
	case *hist.EventAlreadyStartedError:
		h.metricsClient.IncCounter(scope, metrics.CadenceErrEventAlreadyStartedCounter)
	case *gen.BadRequestError:
		h.metricsClient.IncCounter(scope, metrics.CadenceErrBadRequestCounter)
	case *gen.EntityNotExistsError:
		h.metricsClient.IncCounter(scope, metrics.CadenceErrEntityNotExistsCounter)
	default:
		h.metricsClient.IncCounter(scope, metrics.CadenceFailures)
	}
}

func createShardOwnershipLostError(currentHost, ownerHost string) *hist.ShardOwnershipLostError {
	shardLostErr := hist.NewShardOwnershipLostError()
	shardLostErr.Message = common.StringPtr(fmt.Sprintf("Shard is not owned by host: %v", currentHost))
	shardLostErr.Owner = common.StringPtr(ownerHost)

	return shardLostErr
}
