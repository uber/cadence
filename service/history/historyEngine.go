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

package history

import (
	"errors"
	"fmt"

	"github.com/pborman/uuid"
	"github.com/uber-common/bark"
	h "github.com/uber/cadence/.gen/go/history"
	workflow "github.com/uber/cadence/.gen/go/shared"
	hc "github.com/uber/cadence/client/history"
	"github.com/uber/cadence/client/matching"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/logging"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
)

const (
	conditionalRetryCount                    = 5
	activityCancelationMsgActivityIDUnknown  = "ACTIVITY_ID_UNKNOWN"
	activityCancelationMsgActivityNotStarted = "ACTIVITY_ID_NOT_STARTED"
	timerCancelationMsgTimerIDUnknown        = "TIMER_ID_UNKNOWN"
)

type (
	historyEngineImpl struct {
		shard              ShardContext
		metadataMgr        persistence.MetadataManager
		historyMgr         persistence.HistoryManager
		executionManager   persistence.ExecutionManager
		txProcessor        transferQueueProcessor
		timerProcessor     timerQueueProcessor
		tokenSerializer    common.TaskTokenSerializer
		hSerializerFactory persistence.HistorySerializerFactory
		metricsReporter    metrics.Client
		historyCache       *historyCache
		domainCache        cache.DomainCache
		metricsClient      metrics.Client
		logger             bark.Logger
	}

	// shardContextWrapper wraps ShardContext to notify transferQueueProcessor on new tasks.
	// TODO: use to notify timerQueueProcessor as well.
	shardContextWrapper struct {
		ShardContext
		txProcessor transferQueueProcessor
	}
)

var _ Engine = (*historyEngineImpl)(nil)

var (
	// ErrDuplicate is exported temporarily for integration test
	ErrDuplicate = errors.New("Duplicate task, completing it")
	// ErrConflict is exported temporarily for integration test
	ErrConflict = errors.New("Conditional update failed")
	// ErrMaxAttemptsExceeded is exported temporarily for integration test
	ErrMaxAttemptsExceeded = errors.New("Maximum attempts exceeded to update history")
)

// NewEngineWithShardContext creates an instance of history engine
func NewEngineWithShardContext(shard ShardContext, metadataMgr persistence.MetadataManager,
	visibilityMgr persistence.VisibilityManager, matching matching.Client, historyClient hc.Client) Engine {
	shardWrapper := &shardContextWrapper{ShardContext: shard}
	shard = shardWrapper
	logger := shard.GetLogger()
	executionManager := shard.GetExecutionManager()
	historyManager := shard.GetHistoryManager()
	historyCache := newHistoryCache(historyCacheMaxSize, shard, logger)
	domainCache := cache.NewDomainCache(metadataMgr, logger)
	txProcessor := newTransferQueueProcessor(shard, visibilityMgr, matching, historyClient, historyCache, domainCache)
	historyEngImpl := &historyEngineImpl{
		shard:              shard,
		metadataMgr:        metadataMgr,
		historyMgr:         historyManager,
		executionManager:   executionManager,
		txProcessor:        txProcessor,
		tokenSerializer:    common.NewJSONTaskTokenSerializer(),
		hSerializerFactory: persistence.NewHistorySerializerFactory(),
		historyCache:       historyCache,
		domainCache:        domainCache,
		logger: logger.WithFields(bark.Fields{
			logging.TagWorkflowComponent: logging.TagValueHistoryEngineComponent,
		}),
		metricsClient: shard.GetMetricsClient(),
	}
	historyEngImpl.timerProcessor = newTimerQueueProcessor(shard, historyEngImpl, executionManager, logger)
	shardWrapper.txProcessor = txProcessor
	return historyEngImpl
}

// Start will spin up all the components needed to start serving this shard.
// Make sure all the components are loaded lazily so start can return immediately.  This is important because
// ShardController calls start sequentially for all the shards for a given host during startup.
func (e *historyEngineImpl) Start() {
	logging.LogHistoryEngineStartingEvent(e.logger)
	defer logging.LogHistoryEngineStartedEvent(e.logger)

	e.txProcessor.Start()
	e.timerProcessor.Start()
}

// Stop the service.
func (e *historyEngineImpl) Stop() {
	logging.LogHistoryEngineShuttingDownEvent(e.logger)
	defer logging.LogHistoryEngineShutdownEvent(e.logger)

	e.txProcessor.Stop()
	e.timerProcessor.Stop()
}

// StartWorkflowExecution starts a workflow execution
func (e *historyEngineImpl) StartWorkflowExecution(startRequest *h.StartWorkflowExecutionRequest) (
	*workflow.StartWorkflowExecutionResponse, error) {
	domainID := startRequest.GetDomainUUID()
	request := startRequest.GetStartRequest()
	executionID := request.GetWorkflowId()
	// We generate a new workflow execution run_id on each StartWorkflowExecution call.  This generated run_id is
	// returned back to the caller as the response to StartWorkflowExecution.
	runID := uuid.New()
	workflowExecution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(executionID),
		RunId:      common.StringPtr(runID),
	}

	var parentExecution *workflow.WorkflowExecution
	initiatedID := emptyEventID
	parentDomainID := ""
	parentInfo := startRequest.GetParentExecutionInfo()
	if parentInfo != nil {
		parentDomainID = parentInfo.GetDomainUUID()
		parentExecution = parentInfo.GetExecution()
		initiatedID = parentInfo.GetInitiatedId()
	}

	// Generate first decision task event.
	taskList := request.GetTaskList().GetName()
	msBuilder := newMutableStateBuilder(e.logger)
	startedEvent := msBuilder.AddWorkflowExecutionStartedEvent(domainID, workflowExecution, request)
	if startedEvent == nil {
		return nil, &workflow.InternalServiceError{Message: "Failed to add workflow execution started event."}
	}

	var transferTasks []persistence.Task
	decisionScheduleID := emptyEventID
	decisionStartID := emptyEventID
	decisionTimeout := int32(0)
	if parentInfo == nil {
		// DecisionTask is only created when it is not a Child Workflow Execution
		_, di := msBuilder.AddDecisionTaskScheduledEvent()
		if di == nil {
			return nil, &workflow.InternalServiceError{Message: "Failed to add decision started event."}
		}

		transferTasks = []persistence.Task{&persistence.DecisionTask{
			DomainID: domainID, TaskList: taskList, ScheduleID: di.ScheduleID,
		}}
		decisionScheduleID = di.ScheduleID
		decisionStartID = di.StartedID
		decisionTimeout = di.DecisionTimeout
	}

	// Serialize the history
	serializedHistory, serializedError := msBuilder.hBuilder.Serialize()
	if serializedError != nil {
		logging.LogHistorySerializationErrorEvent(e.logger, serializedError, fmt.Sprintf(
			"HistoryEventBatch serialization error on start workflow.  WorkflowID: %v, RunID: %v", executionID, runID))
		return nil, serializedError
	}

	err1 := e.shard.AppendHistoryEvents(&persistence.AppendHistoryEventsRequest{
		DomainID:  domainID,
		Execution: workflowExecution,
		// It is ok to use 0 for TransactionID because RunID is unique so there are
		// no potential duplicates to override.
		TransactionID: 0,
		FirstEventID:  startedEvent.GetEventId(),
		Events:        serializedHistory,
	})
	if err1 != nil {
		return nil, err1
	}

	_, err := e.shard.CreateWorkflowExecution(&persistence.CreateWorkflowExecutionRequest{
		RequestID:                   request.GetRequestId(),
		DomainID:                    domainID,
		Execution:                   workflowExecution,
		ParentDomainID:              parentDomainID,
		ParentExecution:             parentExecution,
		InitiatedID:                 initiatedID,
		TaskList:                    request.GetTaskList().GetName(),
		WorkflowTypeName:            request.GetWorkflowType().GetName(),
		DecisionTimeoutValue:        request.GetTaskStartToCloseTimeoutSeconds(),
		ExecutionContext:            nil,
		NextEventID:                 msBuilder.GetNextEventID(),
		LastProcessedEvent:          emptyEventID,
		TransferTasks:               transferTasks,
		DecisionScheduleID:          decisionScheduleID,
		DecisionStartedID:           decisionStartID,
		DecisionStartToCloseTimeout: decisionTimeout,
		ContinueAsNew:               false,
	})

	if err != nil {
		switch t := err.(type) {
		case *workflow.WorkflowExecutionAlreadyStartedError:
			// We created the history events but failed to create workflow execution, so cleanup the history which could cause
			// us to leak history events which are never cleaned up.  Cleaning up the events is absolutely safe here as they
			// are always created for a unique run_id which is not visible beyond this call yet.
			// TODO: Handle error on deletion of execution history
			e.historyMgr.DeleteWorkflowExecutionHistory(&persistence.DeleteWorkflowExecutionHistoryRequest{
				DomainID:  domainID,
				Execution: workflowExecution,
			})

			if t.GetStartRequestId() == request.GetRequestId() {
				return &workflow.StartWorkflowExecutionResponse{
					RunId: t.RunId,
				}, nil
			}
		case *persistence.ShardOwnershipLostError:
			// We created the history events but failed to create workflow execution, so cleanup the history which could cause
			// us to leak history events which are never cleaned up. Cleaning up the events is absolutely safe here as they
			// are always created for a unique run_id which is not visible beyond this call yet.
			// TODO: Handle error on deletion of execution history
			e.historyMgr.DeleteWorkflowExecutionHistory(&persistence.DeleteWorkflowExecutionHistoryRequest{
				DomainID:  domainID,
				Execution: workflowExecution,
			})
		}

		logging.LogPersistantStoreErrorEvent(e.logger, logging.TagValueStoreOperationCreateWorkflowExecution, err,
			fmt.Sprintf("{WorkflowID: %v, RunID: %v}", executionID, runID))
		return nil, err
	}

	return &workflow.StartWorkflowExecutionResponse{
		RunId: workflowExecution.RunId,
	}, nil
}

// GetWorkflowExecutionNextEventID retrieves the nextEventId of the workflow execution history
func (e *historyEngineImpl) GetWorkflowExecutionNextEventID(
	request *h.GetWorkflowExecutionNextEventIDRequest) (*h.GetWorkflowExecutionNextEventIDResponse, error) {
	domainID := request.GetDomainUUID()
	execution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(request.GetExecution().GetWorkflowId()),
		RunId:      common.StringPtr(request.GetExecution().GetRunId()),
	}

	context, release, err0 := e.historyCache.getOrCreateWorkflowExecution(domainID, execution)
	if err0 != nil {
		return nil, err0
	}
	defer release()

	msBuilder, err1 := context.loadWorkflowExecution()
	if err1 != nil {
		return nil, err1
	}

	result := h.NewGetWorkflowExecutionNextEventIDResponse()
	result.EventId = common.Int64Ptr(msBuilder.GetNextEventID())
	result.RunId = context.workflowExecution.RunId

	return result, nil
}

func (e *historyEngineImpl) RecordDecisionTaskStarted(
	request *h.RecordDecisionTaskStartedRequest) (*h.RecordDecisionTaskStartedResponse, error) {
	domainID := request.GetDomainUUID()
	context, release, err0 := e.historyCache.getOrCreateWorkflowExecution(domainID, *request.WorkflowExecution)
	if err0 != nil {
		return nil, err0
	}
	defer release()

	scheduleID := request.GetScheduleId()
	requestID := request.GetRequestId()

Update_History_Loop:
	for attempt := 0; attempt < conditionalRetryCount; attempt++ {
		msBuilder, err0 := context.loadWorkflowExecution()
		if err0 != nil {
			return nil, err0
		}

		// First check to see if cache needs to be refreshed as we could potentially have stale workflow execution in
		// some extreme cassandra failure cases.
		if scheduleID >= msBuilder.GetNextEventID() {
			e.metricsClient.IncCounter(metrics.HistoryRecordDecisionTaskStartedScope, metrics.StaleMutableStateCounter)
			// Reload workflow execution history
			context.clear()
			continue Update_History_Loop
		}

		// Check execution state to make sure task is in the list of outstanding tasks and it is not yet started.  If
		// task is not outstanding than it is most probably a duplicate and complete the task.
		di, isRunning := msBuilder.GetPendingDecision(scheduleID)

		if !msBuilder.isWorkflowExecutionRunning() || !isRunning {
			// Looks like DecisionTask already completed as a result of another call.
			// It is OK to drop the task at this point.
			logging.LogDuplicateTaskEvent(context.logger, persistence.TransferTaskTypeDecisionTask, request.GetTaskId(), requestID,
				scheduleID, emptyEventID, isRunning)

			return nil, &workflow.EntityNotExistsError{Message: "Decision task not found."}
		}

		if di.StartedID != emptyEventID {
			// If decision is started as part of the current request scope then return a positive response
			if di.RequestID == requestID {
				return e.createRecordDecisionTaskStartedResponse(domainID, msBuilder, di.StartedID), nil
			}

			// Looks like DecisionTask already started as a result of another call.
			// It is OK to drop the task at this point.
			logging.LogDuplicateTaskEvent(context.logger, persistence.TaskListTypeDecision, request.GetTaskId(), requestID,
				scheduleID, di.StartedID, isRunning)

			return nil, &h.EventAlreadyStartedError{Message: "Decision task already started."}
		}

		event := msBuilder.AddDecisionTaskStartedEvent(scheduleID, requestID, request.PollRequest)
		if event == nil {
			// Unable to add DecisionTaskStarted event to history
			return nil, &workflow.InternalServiceError{Message: "Unable to add DecisionTaskStarted event to history."}
		}

		// Start a timer for the decision task.
		timeOutTask := context.tBuilder.AddDecisionTimoutTask(scheduleID, di.DecisionTimeout)
		timerTasks := []persistence.Task{timeOutTask}
		defer e.timerProcessor.NotifyNewTimer(timerTasks)

		// Generate a transaction ID for appending events to history
		transactionID, err2 := e.shard.GetNextTransferTaskID()
		if err2 != nil {
			return nil, err2
		}

		// We apply the update to execution using optimistic concurrency.  If it fails due to a conflict than reload
		// the history and try the operation again.
		if err3 := context.updateWorkflowExecution(nil, timerTasks, transactionID); err3 != nil {
			if err3 == ErrConflict {
				e.metricsClient.IncCounter(metrics.HistoryRecordDecisionTaskStartedScope,
					metrics.ConcurrencyUpdateFailureCounter)
				continue Update_History_Loop
			}
			return nil, err3
		}

		return e.createRecordDecisionTaskStartedResponse(domainID, msBuilder, event.GetEventId()), nil
	}

	return nil, ErrMaxAttemptsExceeded
}

func (e *historyEngineImpl) RecordActivityTaskStarted(
	request *h.RecordActivityTaskStartedRequest) (*h.RecordActivityTaskStartedResponse, error) {
	domainID := request.GetDomainUUID()
	context, release, err0 := e.historyCache.getOrCreateWorkflowExecution(domainID, *request.WorkflowExecution)
	if err0 != nil {
		return nil, err0
	}
	defer release()

	scheduleID := request.GetScheduleId()
	requestID := request.GetRequestId()

Update_History_Loop:
	for attempt := 0; attempt < conditionalRetryCount; attempt++ {
		msBuilder, err0 := context.loadWorkflowExecution()
		if err0 != nil {
			return nil, err0
		}

		// First check to see if cache needs to be refreshed as we could potentially have stale workflow execution in
		// some extreme cassandra failure cases.
		if scheduleID >= msBuilder.GetNextEventID() {
			e.metricsClient.IncCounter(metrics.HistoryRecordActivityTaskStartedScope, metrics.StaleMutableStateCounter)
			// Reload workflow execution history
			context.clear()
			continue Update_History_Loop
		}

		// Check execution state to make sure task is in the list of outstanding tasks and it is not yet started.  If
		// task is not outstanding than it is most probably a duplicate and complete the task.
		ai, isRunning := msBuilder.GetActivityInfo(scheduleID)
		if !msBuilder.isWorkflowExecutionRunning() || !isRunning {
			// Looks like ActivityTask already completed as a result of another call.
			// It is OK to drop the task at this point.
			logging.LogDuplicateTaskEvent(context.logger, persistence.TransferTaskTypeActivityTask, request.GetTaskId(), requestID,
				scheduleID, emptyEventID, isRunning)

			return nil, &workflow.EntityNotExistsError{Message: "Activity task not found."}
		}

		scheduledEvent, exists := msBuilder.GetActivityScheduledEvent(scheduleID)
		if !exists {
			return nil, &workflow.InternalServiceError{Message: "Corrupted workflow execution state."}
		}

		if ai.StartedID != emptyEventID {
			// If activity is started as part of the current request scope then return a positive response
			if ai.RequestID == requestID {
				response := h.NewRecordActivityTaskStartedResponse()
				startedEvent, exists := msBuilder.GetActivityStartedEvent(scheduleID)
				if !exists {
					return nil, &workflow.InternalServiceError{Message: "Corrupted workflow execution state."}
				}
				response.ScheduledEvent = scheduledEvent
				response.StartedEvent = startedEvent
				return response, nil
			}

			// Looks like ActivityTask already started as a result of another call.
			// It is OK to drop the task at this point.
			logging.LogDuplicateTaskEvent(context.logger, persistence.TransferTaskTypeActivityTask, request.GetTaskId(), requestID,
				scheduleID, ai.StartedID, isRunning)

			return nil, &h.EventAlreadyStartedError{Message: "Activity task already started."}
		}

		startedEvent := msBuilder.AddActivityTaskStartedEvent(ai, scheduleID, requestID, request.PollRequest)
		if startedEvent == nil {
			// Unable to add ActivityTaskStarted event to history
			return nil, &workflow.InternalServiceError{Message: "Unable to add ActivityTaskStarted event to history."}
		}

		// Start a timer for the activity task.
		timerTasks := []persistence.Task{}
		start2CloseTimeoutTask, err := context.tBuilder.AddStartToCloseActivityTimeout(ai)
		if err != nil {
			return nil, err
		}
		timerTasks = append(timerTasks, start2CloseTimeoutTask)

		start2HeartBeatTimeoutTask, err := context.tBuilder.AddHeartBeatActivityTimeout(ai)
		if err != nil {
			return nil, err
		}
		if start2HeartBeatTimeoutTask != nil {
			timerTasks = append(timerTasks, start2HeartBeatTimeoutTask)
		}
		defer e.timerProcessor.NotifyNewTimer(timerTasks)

		// Generate a transaction ID for appending events to history
		transactionID, err2 := e.shard.GetNextTransferTaskID()
		if err2 != nil {
			return nil, err2
		}

		// We apply the update to execution using optimistic concurrency.  If it fails due to a conflict than reload
		// the history and try the operationi again.
		if err3 := context.updateWorkflowExecution(nil, timerTasks, transactionID); err3 != nil {
			if err3 == ErrConflict {
				e.metricsClient.IncCounter(metrics.HistoryRecordActivityTaskStartedScope,
					metrics.ConcurrencyUpdateFailureCounter)
				continue Update_History_Loop
			}

			return nil, err3
		}

		response := h.NewRecordActivityTaskStartedResponse()
		response.ScheduledEvent = scheduledEvent
		response.StartedEvent = startedEvent
		return response, nil
	}

	return nil, ErrMaxAttemptsExceeded
}

// RespondDecisionTaskCompleted completes a decision task
func (e *historyEngineImpl) RespondDecisionTaskCompleted(req *h.RespondDecisionTaskCompletedRequest) error {
	domainID := req.GetDomainUUID()
	request := req.GetCompleteRequest()
	token, err0 := e.tokenSerializer.Deserialize(request.GetTaskToken())
	if err0 != nil {
		return &workflow.BadRequestError{Message: "Error deserializing task token."}
	}

	workflowExecution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(token.WorkflowID),
		RunId:      common.StringPtr(token.RunID),
	}

	context, release, err0 := e.historyCache.getOrCreateWorkflowExecution(domainID, workflowExecution)
	if err0 != nil {
		return err0
	}
	defer release()

Update_History_Loop:
	for attempt := 0; attempt < conditionalRetryCount; attempt++ {
		msBuilder, err1 := context.loadWorkflowExecution()
		if err1 != nil {
			return err1
		}

		scheduleID := token.ScheduleID
		// First check to see if cache needs to be refreshed as we could potentially have stale workflow execution in
		// some extreme cassandra failure cases.
		if scheduleID >= msBuilder.GetNextEventID() {
			e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope, metrics.StaleMutableStateCounter)
			// Reload workflow execution history
			context.clear()
			continue Update_History_Loop
		}

		di, isRunning := msBuilder.GetPendingDecision(scheduleID)
		if !msBuilder.isWorkflowExecutionRunning() || !isRunning || di.StartedID == emptyEventID {
			return &workflow.EntityNotExistsError{Message: "Decision task not found."}
		}

		startedID := di.StartedID
		completedEvent := msBuilder.AddDecisionTaskCompletedEvent(scheduleID, startedID, request)
		if completedEvent == nil {
			return &workflow.InternalServiceError{Message: "Unable to add DecisionTaskCompleted event to history."}
		}

		failDecision := false
		var failCause workflow.DecisionTaskFailedCause
		var err error
		completedID := completedEvent.GetEventId()
		hasUnhandledEvents := ((completedID - startedID) > 1)
		isComplete := false
		transferTasks := []persistence.Task{}
		timerTasks := []persistence.Task{}
		var continueAsNewBuilder *mutableStateBuilder
		userTimersLoaded := false
	Process_Decision_Loop:
		for _, d := range request.Decisions {
			switch d.GetDecisionType() {
			case workflow.DecisionType_ScheduleActivityTask:
				e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope,
					metrics.DecisionTypeScheduleActivityCounter)
				targetDomainID := domainID
				attributes := d.GetScheduleActivityTaskDecisionAttributes()
				// First check if we need to use a different target domain to schedule activity
				if attributes.IsSetDomain() {
					// TODO: Error handling for ActivitySchedule failed when domain lookup fails
					info, _, err := e.domainCache.GetDomain(attributes.GetDomain())
					if err != nil {
						return &workflow.InternalServiceError{Message: "Unable to schedule activity across domain."}
					}
					targetDomainID = info.ID
				}

				if err = validateActivityScheduleAttributes(attributes); err != nil {
					failDecision = true
					failCause = workflow.DecisionTaskFailedCause_BAD_SCHEDULE_ACTIVITY_ATTRIBUTES
					break Process_Decision_Loop
				}

				scheduleEvent, ai := msBuilder.AddActivityTaskScheduledEvent(completedID, attributes)
				transferTasks = append(transferTasks, &persistence.ActivityTask{
					DomainID:   targetDomainID,
					TaskList:   attributes.GetTaskList().GetName(),
					ScheduleID: scheduleEvent.GetEventId(),
				})

				// Create activity timeouts.
				Schedule2StartTimeoutTask := context.tBuilder.AddScheduleToStartActivityTimeout(ai)
				timerTasks = append(timerTasks, Schedule2StartTimeoutTask)

				Schedule2CloseTimeoutTask, err := context.tBuilder.AddScheduleToCloseActivityTimeout(ai)
				if err != nil {
					return err
				}
				timerTasks = append(timerTasks, Schedule2CloseTimeoutTask)

			case workflow.DecisionType_CompleteWorkflowExecution:
				e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope,
					metrics.DecisionTypeCompleteWorkflowCounter)
				if hasUnhandledEvents {
					failDecision = true
					failCause = workflow.DecisionTaskFailedCause_UNHANDLED_DECISION
					break Process_Decision_Loop
				}

				// If the decision has more than one completion event than just pick the first one
				if isComplete {
					e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope,
						metrics.MultipleCompletionDecisionsCounter)
					logging.LogMultipleCompletionDecisionsEvent(e.logger, d.GetDecisionType())
					continue Process_Decision_Loop
				}
				attributes := d.GetCompleteWorkflowExecutionDecisionAttributes()
				if err = validateCompleteWorkflowExecutionAttributes(attributes); err != nil {
					failDecision = true
					failCause = workflow.DecisionTaskFailedCause_BAD_COMPLETE_WORKFLOW_EXECUTION_ATTRIBUTES
					break Process_Decision_Loop
				}
				msBuilder.AddCompletedWorkflowEvent(completedID, attributes)
				isComplete = true
			case workflow.DecisionType_FailWorkflowExecution:
				e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope,
					metrics.DecisionTypeFailWorkflowCounter)
				if hasUnhandledEvents {
					failDecision = true
					failCause = workflow.DecisionTaskFailedCause_UNHANDLED_DECISION
					break Process_Decision_Loop
				}

				// If the decision has more than one completion event than just pick the first one
				if isComplete {
					e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope,
						metrics.MultipleCompletionDecisionsCounter)
					logging.LogMultipleCompletionDecisionsEvent(e.logger, d.GetDecisionType())
					continue Process_Decision_Loop
				}
				attributes := d.GetFailWorkflowExecutionDecisionAttributes()
				if err = validateFailWorkflowExecutionAttributes(attributes); err != nil {
					failDecision = true
					failCause = workflow.DecisionTaskFailedCause_BAD_FAIL_WORKFLOW_EXECUTION_ATTRIBUTES
					break Process_Decision_Loop
				}
				msBuilder.AddFailWorkflowEvent(completedID, attributes)
				isComplete = true
			case workflow.DecisionType_CancelWorkflowExecution:
				e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope,
					metrics.DecisionTypeCancelWorkflowCounter)
				// If new events came while we are processing the decision, we would fail this and give a chance to client
				// to process the new event.
				if hasUnhandledEvents {
					failDecision = true
					failCause = workflow.DecisionTaskFailedCause_UNHANDLED_DECISION
					break Process_Decision_Loop
				}

				// If the decision has more than one completion event than just pick the first one
				if isComplete {
					e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope,
						metrics.MultipleCompletionDecisionsCounter)
					logging.LogMultipleCompletionDecisionsEvent(e.logger, d.GetDecisionType())
					continue Process_Decision_Loop
				}
				attributes := d.GetCancelWorkflowExecutionDecisionAttributes()
				if err = validateCancelWorkflowExecutionAttributes(attributes); err != nil {
					failDecision = true
					failCause = workflow.DecisionTaskFailedCause_BAD_CANCEL_WORKFLOW_EXECUTION_ATTRIBUTES
					break Process_Decision_Loop
				}
				msBuilder.AddWorkflowExecutionCanceledEvent(completedID, attributes)
				isComplete = true

			case workflow.DecisionType_StartTimer:
				e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope,
					metrics.DecisionTypeStartTimerCounter)
				attributes := d.GetStartTimerDecisionAttributes()
				if err = validateTimerScheduleAttributes(attributes); err != nil {
					failDecision = true
					failCause = workflow.DecisionTaskFailedCause_BAD_START_TIMER_ATTRIBUTES
					break Process_Decision_Loop
				}
				if !userTimersLoaded {
					context.tBuilder.LoadUserTimers(msBuilder)
					userTimersLoaded = true
				}
				_, ti := msBuilder.AddTimerStartedEvent(completedID, attributes)
				if ti == nil {
					failDecision = true
					failCause = workflow.DecisionTaskFailedCause_BAD_START_TIMER_DUPLICATE_ID
					break Process_Decision_Loop
				}
				context.tBuilder.AddUserTimer(ti)

			case workflow.DecisionType_RequestCancelActivityTask:
				e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope,
					metrics.DecisionTypeCancelActivityCounter)
				attributes := d.GetRequestCancelActivityTaskDecisionAttributes()
				if err = validateActivityCancelAttributes(attributes); err != nil {
					failDecision = true
					failCause = workflow.DecisionTaskFailedCause_BAD_REQUEST_CANCEL_ACTIVITY_ATTRIBUTES
					break Process_Decision_Loop
				}
				activityID := attributes.GetActivityId()
				actCancelReqEvent, ai, isRunning := msBuilder.AddActivityTaskCancelRequestedEvent(completedID, activityID,
					request.GetIdentity())
				if !isRunning {
					msBuilder.AddRequestCancelActivityTaskFailedEvent(completedID, activityID,
						activityCancelationMsgActivityIDUnknown)
					continue Process_Decision_Loop
				}

				if ai.StartedID == emptyEventID {
					// We haven't started the activity yet, we can cancel the activity right away.
					msBuilder.AddActivityTaskCanceledEvent(ai.ScheduleID, ai.StartedID, actCancelReqEvent.GetEventId(),
						[]byte(activityCancelationMsgActivityNotStarted), request.GetIdentity())
				}

			case workflow.DecisionType_CancelTimer:
				e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope,
					metrics.DecisionTypeCancelTimerCounter)
				attributes := d.GetCancelTimerDecisionAttributes()
				if err = validateTimerCancelAttributes(attributes); err != nil {
					failDecision = true
					failCause = workflow.DecisionTaskFailedCause_BAD_CANCEL_TIMER_ATTRIBUTES
					break Process_Decision_Loop
				}
				if msBuilder.AddTimerCanceledEvent(completedID, attributes, request.GetIdentity()) == nil {
					msBuilder.AddCancelTimerFailedEvent(completedID, attributes, request.GetIdentity())
				}

			case workflow.DecisionType_RecordMarker:
				e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope,
					metrics.DecisionTypeRecordMarkerCounter)
				attributes := d.GetRecordMarkerDecisionAttributes()
				if err = validateRecordMarkerAttributes(attributes); err != nil {
					failDecision = true
					failCause = workflow.DecisionTaskFailedCause_BAD_RECORD_MARKER_ATTRIBUTES
					break Process_Decision_Loop
				}
				msBuilder.AddRecordMarkerEvent(completedID, attributes)

			case workflow.DecisionType_RequestCancelExternalWorkflowExecution:
				e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope,
					metrics.DecisionTypeCancelExternalWorkflowCounter)
				attributes := d.GetRequestCancelExternalWorkflowExecutionDecisionAttributes()
				if err = validateCancelExternalWorkflowExecutionAttributes(attributes); err != nil {
					failDecision = true
					failCause = workflow.DecisionTaskFailedCause_BAD_REQUEST_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_ATTRIBUTES
					break Process_Decision_Loop
				}

				foreignInfo, _, err := e.domainCache.GetDomain(attributes.GetDomain())
				if err != nil {
					return &workflow.InternalServiceError{
						Message: fmt.Sprintf("Unable to schedule activity across domain: %v.",
							attributes.GetDomain())}
				}

				wfCancelReqEvent := msBuilder.AddRequestCancelExternalWorkflowExecutionInitiatedEvent(
					completedID, attributes)
				if wfCancelReqEvent == nil {
					return &workflow.InternalServiceError{Message: "Unable to add external cancel workflow request."}
				}

				transferTasks = append(transferTasks, &persistence.CancelExecutionTask{
					TargetDomainID:   foreignInfo.ID,
					TargetWorkflowID: attributes.GetWorkflowId(),
					TargetRunID:      attributes.GetRunId(),
					ScheduleID:       wfCancelReqEvent.GetEventId(),
				})

			case workflow.DecisionType_ContinueAsNewWorkflowExecution:
				e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope,
					metrics.DecisionTypeContinueAsNewCounter)
				if hasUnhandledEvents {
					failDecision = true
					failCause = workflow.DecisionTaskFailedCause_UNHANDLED_DECISION
					break Process_Decision_Loop
				}

				// If the decision has more than one completion event than just pick the first one
				if isComplete {
					e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope,
						metrics.MultipleCompletionDecisionsCounter)
					logging.LogMultipleCompletionDecisionsEvent(e.logger, d.GetDecisionType())
					continue Process_Decision_Loop
				}
				attributes := d.GetContinueAsNewWorkflowExecutionDecisionAttributes()
				if err = validateContinueAsNewWorkflowExecutionAttributes(attributes); err != nil {
					failDecision = true
					failCause = workflow.DecisionTaskFailedCause_BAD_CONTINUE_AS_NEW_ATTRIBUTES
					break Process_Decision_Loop
				}
				runID := uuid.New()
				_, newStateBuilder, err := msBuilder.AddContinueAsNewEvent(completedID, domainID, runID, attributes)
				if err != nil {
					return nil
				}
				isComplete = true
				continueAsNewBuilder = newStateBuilder

			case workflow.DecisionType_StartChildWorkflowExecution:
				e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope,
					metrics.DecisionTypeChildWorkflowCounter)
				targetDomainID := domainID
				attributes := d.GetStartChildWorkflowExecutionDecisionAttributes()
				// First check if we need to use a different target domain to schedule child execution
				if attributes.IsSetDomain() {
					// TODO: Error handling for DecisionType_StartChildWorkflowExecution failed when domain lookup fails
					info, _, err := e.domainCache.GetDomain(attributes.GetDomain())
					if err != nil {
						return &workflow.InternalServiceError{Message: "Unable to schedule child execution across domain."}
					}
					targetDomainID = info.ID
				}

				requestID := uuid.New()
				initiatedEvent, _ := msBuilder.AddStartChildWorkflowExecutionInitiatedEvent(completedID, requestID, attributes)
				transferTasks = append(transferTasks, &persistence.StartChildExecutionTask{
					TargetDomainID:   targetDomainID,
					TargetWorkflowID: attributes.GetWorkflowId(),
					InitiatedID:      initiatedEvent.GetEventId(),
				})

			default:
				return &workflow.BadRequestError{Message: fmt.Sprintf("Unknown decision type: %v", d.GetDecisionType())}
			}
		}

		if failDecision {
			e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope, metrics.FailedDecisionsCounter)
			logging.LogDecisionFailedEvent(e.logger, domainID, token.WorkflowID, token.RunID, failCause)
			var err1 error
			msBuilder, err1 = e.failDecision(context, scheduleID, startedID, failCause, request)
			if err1 != nil {
				return err1
			}
			isComplete = false
			hasUnhandledEvents = true
			continueAsNewBuilder = nil
		}

		if userTimersLoaded {
			if tt := context.tBuilder.GetUserTimerTaskIfNeeded(msBuilder); tt != nil {
				timerTasks = append(timerTasks, tt)
			}
		}

		// Schedule another decision task if new events came in during this decision
		if hasUnhandledEvents {
			newDecisionEvent, _ := msBuilder.AddDecisionTaskScheduledEvent()
			transferTasks = append(transferTasks, &persistence.DecisionTask{
				DomainID:   domainID,
				TaskList:   newDecisionEvent.GetDecisionTaskScheduledEventAttributes().GetTaskList().GetName(),
				ScheduleID: newDecisionEvent.GetEventId(),
			})
		}

		if isComplete {
			// Generate a transfer task to delete workflow execution
			transferTasks = append(transferTasks, &persistence.DeleteExecutionTask{})
		}

		// Generate a transaction ID for appending events to history
		transactionID, err3 := e.shard.GetNextTransferTaskID()
		if err3 != nil {
			return err3
		}

		// We apply the update to execution using optimistic concurrency.  If it fails due to a conflict then reload
		// the history and try the operation again.
		var updateErr error
		if continueAsNewBuilder != nil {
			updateErr = context.continueAsNewWorkflowExecution(request.GetExecutionContext(), continueAsNewBuilder,
				transferTasks, transactionID)
		} else {
			updateErr = context.updateWorkflowExecutionWithContext(request.GetExecutionContext(), transferTasks, timerTasks,
				transactionID)
		}

		if updateErr != nil {
			if updateErr == ErrConflict {
				e.metricsClient.IncCounter(metrics.HistoryRespondDecisionTaskCompletedScope,
					metrics.ConcurrencyUpdateFailureCounter)
				continue Update_History_Loop
			}

			return updateErr
		}

		// Inform timer about the new ones.
		e.timerProcessor.NotifyNewTimer(timerTasks)

		return err
	}

	return ErrMaxAttemptsExceeded
}

// RespondActivityTaskCompleted completes an activity task.
func (e *historyEngineImpl) RespondActivityTaskCompleted(req *h.RespondActivityTaskCompletedRequest) error {
	domainID := req.GetDomainUUID()
	request := req.GetCompleteRequest()
	token, err0 := e.tokenSerializer.Deserialize(request.GetTaskToken())
	if err0 != nil {
		return &workflow.BadRequestError{Message: "Error deserializing task token."}
	}

	workflowExecution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(token.WorkflowID),
		RunId:      common.StringPtr(token.RunID),
	}

	context, release, err0 := e.historyCache.getOrCreateWorkflowExecution(domainID, workflowExecution)
	if err0 != nil {
		return err0
	}
	defer release()

Update_History_Loop:
	for attempt := 0; attempt < conditionalRetryCount; attempt++ {
		msBuilder, err1 := context.loadWorkflowExecution()
		if err1 != nil {
			return err1
		}

		scheduleID := token.ScheduleID

		// First check to see if cache needs to be refreshed as we could potentially have stale workflow execution in
		// some extreme cassandra failure cases.
		if scheduleID >= msBuilder.GetNextEventID() {
			e.metricsClient.IncCounter(metrics.HistoryRespondActivityTaskCompletedScope, metrics.StaleMutableStateCounter)
			// Reload workflow execution history
			context.clear()
			continue Update_History_Loop
		}

		ai, isRunning := msBuilder.GetActivityInfo(scheduleID)
		if !msBuilder.isWorkflowExecutionRunning() || !isRunning || ai.StartedID == emptyEventID {
			return &workflow.EntityNotExistsError{Message: "Activity task not found."}
		}

		startedID := ai.StartedID
		if msBuilder.AddActivityTaskCompletedEvent(scheduleID, startedID, request) == nil {
			// Unable to add ActivityTaskCompleted event to history
			return &workflow.InternalServiceError{Message: "Unable to add ActivityTaskCompleted event to history."}
		}

		var transferTasks []persistence.Task
		if !msBuilder.HasPendingDecisionTask() {
			newDecisionEvent, _ := msBuilder.AddDecisionTaskScheduledEvent()
			transferTasks = []persistence.Task{&persistence.DecisionTask{
				DomainID:   domainID,
				TaskList:   newDecisionEvent.GetDecisionTaskScheduledEventAttributes().GetTaskList().GetName(),
				ScheduleID: newDecisionEvent.GetEventId(),
			}}
		}

		// Generate a transaction ID for appending events to history
		transactionID, err2 := e.shard.GetNextTransferTaskID()
		if err2 != nil {
			return err2
		}

		// We apply the update to execution using optimistic concurrency.  If it fails due to a conflict than reload
		// the history and try the operation again.
		if err := context.updateWorkflowExecution(transferTasks, nil, transactionID); err != nil {
			if err == ErrConflict {
				e.metricsClient.IncCounter(metrics.HistoryRespondActivityTaskCompletedScope,
					metrics.ConcurrencyUpdateFailureCounter)
				continue Update_History_Loop
			}

			return err
		}

		return nil
	}

	return ErrMaxAttemptsExceeded
}

// RespondActivityTaskFailed completes an activity task failure.
func (e *historyEngineImpl) RespondActivityTaskFailed(req *h.RespondActivityTaskFailedRequest) error {
	domainID := req.GetDomainUUID()
	request := req.GetFailedRequest()
	token, err0 := e.tokenSerializer.Deserialize(request.GetTaskToken())
	if err0 != nil {
		return &workflow.BadRequestError{Message: "Error deserializing task token."}
	}

	workflowExecution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(token.WorkflowID),
		RunId:      common.StringPtr(token.RunID),
	}

	context, release, err0 := e.historyCache.getOrCreateWorkflowExecution(domainID, workflowExecution)
	if err0 != nil {
		return err0
	}
	defer release()

Update_History_Loop:
	for attempt := 0; attempt < conditionalRetryCount; attempt++ {
		msBuilder, err1 := context.loadWorkflowExecution()
		if err1 != nil {
			return err1
		}

		scheduleID := token.ScheduleID

		// First check to see if cache needs to be refreshed as we could potentially have stale workflow execution in
		// some extreme cassandra failure cases.
		if scheduleID >= msBuilder.GetNextEventID() {
			e.metricsClient.IncCounter(metrics.HistoryRespondActivityTaskFailedScope, metrics.StaleMutableStateCounter)
			// Reload workflow execution history
			context.clear()
			continue Update_History_Loop
		}

		ai, isRunning := msBuilder.GetActivityInfo(scheduleID)
		if !msBuilder.isWorkflowExecutionRunning() || !isRunning || ai.StartedID == emptyEventID {
			return &workflow.EntityNotExistsError{Message: "Activity task not found."}
		}

		startedID := ai.StartedID
		if msBuilder.AddActivityTaskFailedEvent(scheduleID, startedID, request) == nil {
			// Unable to add ActivityTaskFailed event to history
			return &workflow.InternalServiceError{Message: "Unable to add ActivityTaskFailed event to history."}
		}

		var transferTasks []persistence.Task
		if !msBuilder.HasPendingDecisionTask() {
			newDecisionEvent, _ := msBuilder.AddDecisionTaskScheduledEvent()
			transferTasks = []persistence.Task{&persistence.DecisionTask{
				DomainID:   domainID,
				TaskList:   newDecisionEvent.GetDecisionTaskScheduledEventAttributes().GetTaskList().GetName(),
				ScheduleID: newDecisionEvent.GetEventId(),
			}}
		}

		// Generate a transaction ID for appending events to history
		transactionID, err3 := e.shard.GetNextTransferTaskID()
		if err3 != nil {
			return err3
		}

		// We apply the update to execution using optimistic concurrency.  If it fails due to a conflict than reload
		// the history and try the operation again.
		if err := context.updateWorkflowExecution(transferTasks, nil, transactionID); err != nil {
			if err == ErrConflict {
				e.metricsClient.IncCounter(metrics.HistoryRespondActivityTaskFailedScope,
					metrics.ConcurrencyUpdateFailureCounter)
				continue Update_History_Loop
			}

			return err
		}

		return nil
	}

	return ErrMaxAttemptsExceeded
}

// RespondActivityTaskCanceled completes an activity task failure.
func (e *historyEngineImpl) RespondActivityTaskCanceled(req *h.RespondActivityTaskCanceledRequest) error {
	domainID := req.GetDomainUUID()
	request := req.GetCancelRequest()
	token, err0 := e.tokenSerializer.Deserialize(request.GetTaskToken())
	if err0 != nil {
		return &workflow.BadRequestError{Message: "Error deserializing task token."}
	}

	workflowExecution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(token.WorkflowID),
		RunId:      common.StringPtr(token.RunID),
	}

	context, release, err0 := e.historyCache.getOrCreateWorkflowExecution(domainID, workflowExecution)
	if err0 != nil {
		return err0
	}
	defer release()

Update_History_Loop:
	for attempt := 0; attempt < conditionalRetryCount; attempt++ {
		msBuilder, err1 := context.loadWorkflowExecution()
		if err1 != nil {
			return err1
		}

		scheduleID := token.ScheduleID
		// First check to see if cache needs to be refreshed as we could potentially have stale workflow execution in
		// some extreme cassandra failure cases.
		if scheduleID >= msBuilder.GetNextEventID() {
			e.metricsClient.IncCounter(metrics.HistoryRespondActivityTaskCanceledScope, metrics.StaleMutableStateCounter)
			// Reload workflow execution history
			context.clear()
			continue Update_History_Loop
		}

		// Check execution state to make sure task is in the list of outstanding tasks and it is not yet started.  If
		// task is not outstanding than it is most probably a duplicate and complete the task.
		ai, isRunning := msBuilder.GetActivityInfo(scheduleID)
		if !msBuilder.isWorkflowExecutionRunning() || !isRunning || ai.StartedID == emptyEventID {
			return &workflow.EntityNotExistsError{Message: "Activity task not found."}
		}

		if msBuilder.AddActivityTaskCanceledEvent(scheduleID, ai.StartedID, ai.CancelRequestID, request.GetDetails(),
			request.GetIdentity()) == nil {
			// Unable to add ActivityTaskCanceled event to history
			return &workflow.InternalServiceError{Message: "Unable to add ActivityTaskCanceled event to history."}
		}

		var transferTasks []persistence.Task
		if !msBuilder.HasPendingDecisionTask() {
			newDecisionEvent, _ := msBuilder.AddDecisionTaskScheduledEvent()
			transferTasks = []persistence.Task{&persistence.DecisionTask{
				DomainID:   domainID,
				TaskList:   newDecisionEvent.GetDecisionTaskScheduledEventAttributes().GetTaskList().GetName(),
				ScheduleID: newDecisionEvent.GetEventId(),
			}}
		}

		// Generate a transaction ID for appending events to history
		transactionID, err3 := e.shard.GetNextTransferTaskID()
		if err3 != nil {
			return err3
		}

		// We apply the update to execution using optimistic concurrency.  If it fails due to a conflict than reload
		// the history and try the operation again.
		if err := context.updateWorkflowExecution(transferTasks, nil, transactionID); err != nil {
			if err == ErrConflict {
				e.metricsClient.IncCounter(metrics.HistoryRespondActivityTaskCanceledScope,
					metrics.ConcurrencyUpdateFailureCounter)
				continue Update_History_Loop
			}

			return err
		}

		return nil
	}

	return ErrMaxAttemptsExceeded
}

// RecordActivityTaskHeartbeat records an hearbeat for a task.
// This method can be used for two purposes.
// - For reporting liveness of the activity.
// - For reporting progress of the activity, this can be done even if the liveness is not configured.
func (e *historyEngineImpl) RecordActivityTaskHeartbeat(
	req *h.RecordActivityTaskHeartbeatRequest) (*workflow.RecordActivityTaskHeartbeatResponse, error) {
	domainID := req.GetDomainUUID()
	request := req.GetHeartbeatRequest()
	token, err0 := e.tokenSerializer.Deserialize(request.GetTaskToken())
	if err0 != nil {
		return nil, &workflow.BadRequestError{Message: "Error deserializing task token."}
	}

	workflowExecution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(token.WorkflowID),
		RunId:      common.StringPtr(token.RunID),
	}

	context, release, err0 := e.historyCache.getOrCreateWorkflowExecution(domainID, workflowExecution)
	if err0 != nil {
		return nil, err0
	}
	defer release()

Update_History_Loop:
	for attempt := 0; attempt < conditionalRetryCount; attempt++ {
		msBuilder, err1 := context.loadWorkflowExecution()
		if err1 != nil {
			return nil, err1
		}

		scheduleID := token.ScheduleID
		// First check to see if cache needs to be refreshed as we could potentially have stale workflow execution in
		// some extreme cassandra failure cases.
		if scheduleID >= msBuilder.GetNextEventID() {
			e.metricsClient.IncCounter(metrics.HistoryRecordActivityTaskHeartbeatScope, metrics.StaleMutableStateCounter)
			// Reload workflow execution history
			context.clear()
			continue Update_History_Loop
		}

		ai, isRunning := msBuilder.GetActivityInfo(scheduleID)
		if !msBuilder.isWorkflowExecutionRunning() || !isRunning || ai.StartedID == emptyEventID {
			e.logger.Debugf("Activity HeartBeat: scheduleEventID: %v, ActivityInfo: %+v, Exist: %v",
				scheduleID, ai, isRunning)
			return nil, &workflow.EntityNotExistsError{Message: "Activity task not found."}
		}

		cancelRequested := ai.CancelRequested

		e.logger.Debugf("Activity HeartBeat: scheduleEventID: %v, ActivityInfo: %+v, CancelRequested: %v",
			scheduleID, ai, cancelRequested)

		// Save progress and last HB reported time.
		msBuilder.updateActivityProgress(ai, request)

		// Generate a transaction ID for appending events to history
		transactionID, err2 := e.shard.GetNextTransferTaskID()
		if err2 != nil {
			return nil, err2
		}

		// We apply the update to execution using optimistic concurrency.  If it fails due to a conflict than reload
		// the history and try the operation again.
		if err := context.updateWorkflowExecution(nil, nil, transactionID); err != nil {
			if err == ErrConflict {
				e.metricsClient.IncCounter(metrics.HistoryRecordActivityTaskHeartbeatScope,
					metrics.ConcurrencyUpdateFailureCounter)
				continue Update_History_Loop
			}

			return nil, err
		}
		return &workflow.RecordActivityTaskHeartbeatResponse{CancelRequested: common.BoolPtr(cancelRequested)}, nil
	}

	return &workflow.RecordActivityTaskHeartbeatResponse{}, ErrMaxAttemptsExceeded
}

// RequestCancelWorkflowExecution
// https://github.com/uber/cadence/issues/145
// TODO: (1) Each external request can result in one cancel requested event. it would be nice
//	 to have dedupe on the server side.
//	(2) if there are multiple calls if one request goes through then can we respond to the other ones with
//       cancellation in progress instead of success.
func (e *historyEngineImpl) RequestCancelWorkflowExecution(
	req *h.RequestCancelWorkflowExecutionRequest) error {
	domainID := req.GetDomainUUID()
	request := req.GetCancelRequest()

	workflowExecution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(request.GetWorkflowExecution().GetWorkflowId()),
		RunId:      common.StringPtr(request.GetWorkflowExecution().GetRunId()),
	}

	return e.updateWorkflowExecution(domainID, workflowExecution, false, true,
		func(msBuilder *mutableStateBuilder) error {
			if !msBuilder.isWorkflowExecutionRunning() {
				return &workflow.EntityNotExistsError{Message: "Workflow execution already completed."}
			}

			if msBuilder.AddWorkflowExecutionCancelRequestedEvent("", req) == nil {
				return &workflow.InternalServiceError{Message: "Unable to cancel workflow execution."}
			}

			return nil
		})
}

func (e *historyEngineImpl) SignalWorkflowExecution(signalRequest *h.SignalWorkflowExecutionRequest) error {
	domainID := signalRequest.GetDomainUUID()
	request := signalRequest.GetSignalRequest()
	execution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(request.GetWorkflowExecution().GetWorkflowId()),
		RunId:      common.StringPtr(request.GetWorkflowExecution().GetRunId()),
	}

	return e.updateWorkflowExecution(domainID, execution, false, true,
		func(msBuilder *mutableStateBuilder) error {
			if !msBuilder.isWorkflowExecutionRunning() {
				return &workflow.EntityNotExistsError{Message: "Workflow execution already completed."}
			}

			if msBuilder.AddWorkflowExecutionSignaled(request) == nil {
				return &workflow.InternalServiceError{Message: "Unable to signal workflow execution."}
			}

			return nil
		})
}

func (e *historyEngineImpl) TerminateWorkflowExecution(terminateRequest *h.TerminateWorkflowExecutionRequest) error {
	domainID := terminateRequest.GetDomainUUID()
	request := terminateRequest.GetTerminateRequest()
	execution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(request.GetWorkflowExecution().GetWorkflowId()),
		RunId:      common.StringPtr(request.GetWorkflowExecution().GetRunId()),
	}

	return e.updateWorkflowExecution(domainID, execution, true, false,
		func(msBuilder *mutableStateBuilder) error {
			if !msBuilder.isWorkflowExecutionRunning() {
				return &workflow.EntityNotExistsError{Message: "Workflow execution already completed."}
			}

			if msBuilder.AddWorkflowExecutionTerminatedEvent(request) == nil {
				return &workflow.InternalServiceError{Message: "Unable to terminate workflow execution."}
			}

			return nil
		})
}

// ScheduleDecisionTask schedules a decision if no outstanding decision found
func (e *historyEngineImpl) ScheduleDecisionTask(scheduleRequest *h.ScheduleDecisionTaskRequest) error {
	domainID := scheduleRequest.GetDomainUUID()
	execution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(scheduleRequest.GetWorkflowExecution().GetWorkflowId()),
		RunId:      common.StringPtr(scheduleRequest.GetWorkflowExecution().GetRunId()),
	}

	return e.updateWorkflowExecution(domainID, execution, false, true,
		func(msBuilder *mutableStateBuilder) error {
			if !msBuilder.isWorkflowExecutionRunning() {
				return &workflow.EntityNotExistsError{Message: "Workflow execution already completed."}
			}

			// Noop

			return nil
		})
}

// RecordChildExecutionCompleted records the completion of child execution into parent execution history
func (e *historyEngineImpl) RecordChildExecutionCompleted(completionRequest *h.RecordChildExecutionCompletedRequest) error {
	domainID := completionRequest.GetDomainUUID()
	execution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(completionRequest.GetWorkflowExecution().GetWorkflowId()),
		RunId:      common.StringPtr(completionRequest.GetWorkflowExecution().GetRunId()),
	}

	return e.updateWorkflowExecution(domainID, execution, false, true,
		func(msBuilder *mutableStateBuilder) error {
			if !msBuilder.isWorkflowExecutionRunning() {
				return &workflow.EntityNotExistsError{Message: "Workflow execution already completed."}
			}

			initiatedID := completionRequest.GetInitiatedId()
			completedExecution := completionRequest.GetCompletedExecution()
			completionEvent := completionRequest.GetCompletionEvent()

			// Check mutable state to make sure child execution is in pending child executions
			ci, isRunning := msBuilder.GetChildExecutionInfo(initiatedID)
			if !isRunning || ci.StartedID == emptyEventID {
				return &workflow.EntityNotExistsError{Message: "Pending child execution not found."}
			}

			switch completionEvent.GetEventType() {
			case workflow.EventType_WorkflowExecutionCompleted:
				attributes := completionEvent.GetWorkflowExecutionCompletedEventAttributes()
				msBuilder.AddChildWorkflowExecutionCompletedEvent(initiatedID, completedExecution, attributes)
			case workflow.EventType_WorkflowExecutionFailed:
				attributes := completionEvent.GetWorkflowExecutionFailedEventAttributes()
				msBuilder.AddChildWorkflowExecutionFailedEvent(initiatedID, completedExecution, attributes)
			case workflow.EventType_WorkflowExecutionCanceled:
				attributes := completionEvent.GetWorkflowExecutionCanceledEventAttributes()
				msBuilder.AddChildWorkflowExecutionCanceledEvent(initiatedID, completedExecution, attributes)
			case workflow.EventType_WorkflowExecutionTerminated:
				attributes := completionEvent.GetWorkflowExecutionTerminatedEventAttributes()
				msBuilder.AddChildWorkflowExecutionTerminatedEvent(initiatedID, completedExecution, attributes)
			}

			return nil
		})
}

func (e *historyEngineImpl) updateWorkflowExecution(domainID string, execution workflow.WorkflowExecution,
	createDeletionTask, createDecisionTask bool,
	action func(builder *mutableStateBuilder) error) error {

	context, release, err0 := e.historyCache.getOrCreateWorkflowExecution(domainID, execution)
	if err0 != nil {
		return err0
	}
	defer release()

Update_History_Loop:
	for attempt := 0; attempt < conditionalRetryCount; attempt++ {
		msBuilder, err1 := context.loadWorkflowExecution()
		if err1 != nil {
			return err1
		}

		var transferTasks []persistence.Task
		if err := action(msBuilder); err != nil {
			return err
		}

		if createDeletionTask {
			// Create a transfer task to delete workflow execution
			transferTasks = append(transferTasks, &persistence.DeleteExecutionTask{})
		}

		if createDecisionTask {
			// Create a transfer task to schedule a decision task
			if !msBuilder.HasPendingDecisionTask() {
				newDecisionEvent, _ := msBuilder.AddDecisionTaskScheduledEvent()
				transferTasks = append(transferTasks, &persistence.DecisionTask{
					DomainID:   domainID,
					TaskList:   newDecisionEvent.GetDecisionTaskScheduledEventAttributes().GetTaskList().GetName(),
					ScheduleID: newDecisionEvent.GetEventId(),
				})
			}
		}

		// Generate a transaction ID for appending events to history
		transactionID, err2 := e.shard.GetNextTransferTaskID()
		if err2 != nil {
			return err2
		}

		// We apply the update to execution using optimistic concurrency.  If it fails due to a conflict then reload
		// the history and try the operation again.
		if err := context.updateWorkflowExecution(transferTasks, nil, transactionID); err != nil {
			if err == ErrConflict {
				continue Update_History_Loop
			}
			return err
		}
		return nil
	}
	return ErrMaxAttemptsExceeded
}

func (e *historyEngineImpl) createRecordDecisionTaskStartedResponse(domainID string, msBuilder *mutableStateBuilder,
	startedEventID int64) *h.RecordDecisionTaskStartedResponse {
	response := h.NewRecordDecisionTaskStartedResponse()
	response.WorkflowType = msBuilder.getWorkflowType()
	if msBuilder.previousDecisionStartedEvent() != emptyEventID {
		response.PreviousStartedEventId = common.Int64Ptr(msBuilder.previousDecisionStartedEvent())
	}
	response.StartedEventId = common.Int64Ptr(startedEventID)

	return response
}

// sets the version and encoding types to defaults if they
// are missing from persistence. This is purely for backwards
// compatibility
func setSerializedHistoryDefaults(history *persistence.SerializedHistoryEventBatch) {
	if history.Version == 0 {
		history.Version = persistence.GetDefaultHistoryVersion()
	}
	if len(history.EncodingType) == 0 {
		history.EncodingType = persistence.DefaultEncodingType
	}
}

func (e *historyEngineImpl) failDecision(context *workflowExecutionContext, scheduleID, startedID int64,
	cause workflow.DecisionTaskFailedCause, request *workflow.RespondDecisionTaskCompletedRequest) (*mutableStateBuilder,
	error) {
	// Clear any updates we have accumulated so far
	context.clear()

	// Reload workflow execution so we can apply the decision task failure event
	msBuilder, err := context.loadWorkflowExecution()
	if err != nil {
		return nil, err
	}

	msBuilder.AddDecisionTaskFailedEvent(scheduleID, startedID, cause, request)

	// Return new builder back to the caller for further updates
	return msBuilder, nil
}

func (s *shardContextWrapper) UpdateWorkflowExecution(request *persistence.UpdateWorkflowExecutionRequest) error {
	err := s.ShardContext.UpdateWorkflowExecution(request)
	if err == nil {
		if len(request.TransferTasks) > 0 {
			s.txProcessor.NotifyNewTask()
		}
	}
	return err
}

func (s *shardContextWrapper) CreateWorkflowExecution(request *persistence.CreateWorkflowExecutionRequest) (
	*persistence.CreateWorkflowExecutionResponse, error) {
	resp, err := s.ShardContext.CreateWorkflowExecution(request)
	if err == nil {
		if len(request.TransferTasks) > 0 {
			s.txProcessor.NotifyNewTask()
		}
	}
	return resp, err
}

func validateActivityScheduleAttributes(attributes *workflow.ScheduleActivityTaskDecisionAttributes) error {
	if attributes == nil {
		return &workflow.BadRequestError{Message: "ScheduleActivityTaskDecisionAttributes is not set on decision."}
	}

	if !attributes.IsSetTaskList() || !attributes.GetTaskList().IsSetName() || attributes.GetTaskList().GetName() == "" {
		return &workflow.BadRequestError{Message: "TaskList is not set on decision."}
	}

	if !attributes.IsSetActivityId() || attributes.GetActivityId() == "" {
		return &workflow.BadRequestError{Message: "ActivityId is not set on decision."}
	}

	if !attributes.IsSetActivityType() || !attributes.GetActivityType().IsSetName() || attributes.GetActivityType().GetName() == "" {
		return &workflow.BadRequestError{Message: "ActivityType is not set on decision."}
	}

	if !attributes.IsSetStartToCloseTimeoutSeconds() || attributes.GetStartToCloseTimeoutSeconds() <= 0 {
		return &workflow.BadRequestError{Message: "A valid StartToCloseTimeoutSeconds is not set on decision."}
	}
	if !attributes.IsSetScheduleToStartTimeoutSeconds() || attributes.GetScheduleToStartTimeoutSeconds() <= 0 {
		return &workflow.BadRequestError{Message: "A valid ScheduleToStartTimeoutSeconds is not set on decision."}
	}
	if !attributes.IsSetScheduleToCloseTimeoutSeconds() || attributes.GetScheduleToCloseTimeoutSeconds() <= 0 {
		return &workflow.BadRequestError{Message: "A valid ScheduleToCloseTimeoutSeconds is not set on decision."}
	}
	if !attributes.IsSetHeartbeatTimeoutSeconds() || attributes.GetHeartbeatTimeoutSeconds() < 0 {
		return &workflow.BadRequestError{Message: "Ac valid HeartbeatTimeoutSeconds is not set on decision."}
	}

	return nil
}

func validateTimerScheduleAttributes(attributes *workflow.StartTimerDecisionAttributes) error {
	if attributes == nil {
		return &workflow.BadRequestError{Message: "StartTimerDecisionAttributes is not set on decision."}
	}
	if !attributes.IsSetTimerId() || attributes.GetTimerId() == "" {
		return &workflow.BadRequestError{Message: "TimerId is not set on decision."}
	}
	if !attributes.IsSetStartToFireTimeoutSeconds() || attributes.GetStartToFireTimeoutSeconds() <= 0 {
		return &workflow.BadRequestError{Message: "A valid StartToFireTimeoutSeconds is not set on decision."}
	}
	return nil
}

func validateActivityCancelAttributes(attributes *workflow.RequestCancelActivityTaskDecisionAttributes) error {
	if attributes == nil {
		return &workflow.BadRequestError{Message: "RequestCancelActivityTaskDecisionAttributes is not set on decision."}
	}
	if !attributes.IsSetActivityId() || attributes.GetActivityId() == "" {
		return &workflow.BadRequestError{Message: "ActivityId is not set on decision."}
	}
	return nil
}

func validateTimerCancelAttributes(attributes *workflow.CancelTimerDecisionAttributes) error {
	if attributes == nil {
		return &workflow.BadRequestError{Message: "CancelTimerDecisionAttributes is not set on decision."}
	}
	if !attributes.IsSetTimerId() || attributes.GetTimerId() == "" {
		return &workflow.BadRequestError{Message: "TimerId is not set on decision."}
	}
	return nil
}

func validateRecordMarkerAttributes(attributes *workflow.RecordMarkerDecisionAttributes) error {
	if attributes == nil {
		return &workflow.BadRequestError{Message: "RecordMarkerDecisionAttributes is not set on decision."}
	}
	if !attributes.IsSetMarkerName() || attributes.GetMarkerName() == "" {
		return &workflow.BadRequestError{Message: "MarkerName is not set on decision."}
	}
	return nil
}

func validateCompleteWorkflowExecutionAttributes(attributes *workflow.CompleteWorkflowExecutionDecisionAttributes) error {
	if attributes == nil {
		return &workflow.BadRequestError{Message: "CompleteWorkflowExecutionDecisionAttributes is not set on decision."}
	}
	return nil
}

func validateFailWorkflowExecutionAttributes(attributes *workflow.FailWorkflowExecutionDecisionAttributes) error {
	if attributes == nil {
		return &workflow.BadRequestError{Message: "FailWorkflowExecutionDecisionAttributes is not set on decision."}
	}
	if !attributes.IsSetReason() {
		return &workflow.BadRequestError{Message: "Reason is not set on decision."}
	}
	return nil
}

func validateCancelWorkflowExecutionAttributes(attributes *workflow.CancelWorkflowExecutionDecisionAttributes) error {
	if attributes == nil {
		return &workflow.BadRequestError{Message: "CancelWorkflowExecutionDecisionAttributes is not set on decision."}
	}
	return nil
}

func validateCancelExternalWorkflowExecutionAttributes(attributes *workflow.RequestCancelExternalWorkflowExecutionDecisionAttributes) error {
	if attributes == nil {
		return &workflow.BadRequestError{Message: "RequestCancelExternalWorkflowExecutionDecisionAttributes is not set on decision."}
	}
	if !attributes.IsSetWorkflowId() {
		return &workflow.BadRequestError{Message: "WorkflowId is not set on decision."}
	}

	if !attributes.IsSetRunId() {
		return &workflow.BadRequestError{Message: "RunId is not set on decision."}
	}

	if uuid.Parse(attributes.GetRunId()) == nil {
		return &workflow.BadRequestError{Message: "Invalid RunId set on decision."}
	}

	return nil
}

func validateContinueAsNewWorkflowExecutionAttributes(attributes *workflow.ContinueAsNewWorkflowExecutionDecisionAttributes) error {
	if attributes == nil {
		return &workflow.BadRequestError{Message: "ContinueAsNewWorkflowExecutionDecisionAttributes is not set on decision."}
	}

	if !attributes.IsSetWorkflowType() || !attributes.GetWorkflowType().IsSetName() || attributes.GetWorkflowType().GetName() == "" {
		return &workflow.BadRequestError{Message: "WorkflowType is not set on decision."}
	}

	if !attributes.IsSetTaskList() || !attributes.GetTaskList().IsSetName() || attributes.GetTaskList().GetName() == "" {
		return &workflow.BadRequestError{Message: "TaskList is not set on decision."}
	}

	if !attributes.IsSetExecutionStartToCloseTimeoutSeconds() || attributes.GetExecutionStartToCloseTimeoutSeconds() <= 0 {
		return &workflow.BadRequestError{Message: "A valid ExecutionStartToCloseTimeoutSeconds is not set on decision."}
	}

	if !attributes.IsSetTaskStartToCloseTimeoutSeconds() || attributes.GetTaskStartToCloseTimeoutSeconds() <= 0 {
		return &workflow.BadRequestError{Message: "A valid TaskStartToCloseTimeoutSeconds is not set on decision."}
	}

	return nil
}
