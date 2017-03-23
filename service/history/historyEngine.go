package history

import (
	"errors"
	"fmt"
	"sync"

	"github.com/pborman/uuid"
	"github.com/uber-common/bark"
	h "github.com/uber/cadence/.gen/go/history"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/client/matching"
	"github.com/uber/cadence/common"
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
		shard            ShardContext
		historyMgr       persistence.HistoryManager
		executionManager persistence.ExecutionManager
		txProcessor      transferQueueProcessor
		timerProcessor   timerQueueProcessor
		tokenSerializer  common.TaskTokenSerializer
		hSerializer      historySerializer
		tracker          *pendingTaskTracker
		metricsReporter  metrics.Client
		cache            *historyCache
		logger           bark.Logger
	}

	pendingTaskTracker struct {
		shard        ShardContext
		txProcessor  transferQueueProcessor
		logger       bark.Logger
		lk           sync.RWMutex
		pendingTasks map[int64]bool
		minID        int64
		maxID        int64
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

func newPendingTaskTracker(shard ShardContext, txProcessor transferQueueProcessor,
	logger bark.Logger) *pendingTaskTracker {
	return &pendingTaskTracker{
		shard:        shard,
		txProcessor:  txProcessor,
		pendingTasks: make(map[int64]bool),
		minID:        shard.GetTransferSequenceNumber(),
		maxID:        shard.GetTransferSequenceNumber(),
		logger:       logger,
	}
}

// NewEngineWithShardContext creates an instance of history engine
func NewEngineWithShardContext(shard ShardContext, matching matching.Client) Engine {
	logger := shard.GetLogger()
	executionManager := shard.GetExecutionManager()
	historyManager := shard.GetHistoryManager()
	cache := newHistoryCache(shard, logger)
	txProcessor := newTransferQueueProcessor(shard, matching, cache)
	tracker := newPendingTaskTracker(shard, txProcessor, logger)
	historyEngImpl := &historyEngineImpl{
		shard:            shard,
		historyMgr:       historyManager,
		executionManager: executionManager,
		txProcessor:      txProcessor,
		tokenSerializer:  common.NewJSONTaskTokenSerializer(),
		tracker:          tracker,
		cache:            cache,
		logger: logger.WithFields(bark.Fields{
			tagWorkflowComponent: tagValueHistoryEngineComponent,
		}),
	}
	historyEngImpl.timerProcessor = newTimerQueueProcessor(historyEngImpl, executionManager, logger)
	return historyEngImpl
}

// Start will spin up all the components needed to start serving this shard.
// Make sure all the components are loaded lazily so start can return immediately.  This is important because
// ShardController calls start sequentially for all the shards for a given host during startup.
func (e *historyEngineImpl) Start() {
	logHistoryEngineStartingEvent(e.logger)
	defer logHistoryEngineStartedEvent(e.logger)

	e.txProcessor.Start()
	e.timerProcessor.Start()
}

// Stop the service.
func (e *historyEngineImpl) Stop() {
	logHistoryEngineShuttingDownEvent(e.logger)
	defer logHistoryEngineShutdownEvent(e.logger)

	e.txProcessor.Stop()
	e.timerProcessor.Stop()
}

// StartWorkflowExecution starts a workflow execution
func (e *historyEngineImpl) StartWorkflowExecution(request *workflow.StartWorkflowExecutionRequest) (
	*workflow.StartWorkflowExecutionResponse, error) {
	executionID := request.GetWorkflowId()
	runID := uuid.New()
	workflowExecution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(executionID),
		RunId:      common.StringPtr(runID),
	}

	// Generate first decision task event.
	taskList := request.GetTaskList().GetName()
	msBuilder := newMutableStateBuilder(e.logger)
	startedEvent := msBuilder.AddWorkflowExecutionStartedEvent(workflowExecution, request)
	if startedEvent == nil {
		return nil, &workflow.InternalServiceError{Message: "Failed to add workflow execution started event."}
	}

	_, di := msBuilder.AddDecisionTaskScheduledEvent()
	if di == nil {
		return nil, &workflow.InternalServiceError{Message: "Failed to add decision started event."}
	}

	// Serialize the history
	events, serializedError := msBuilder.hBuilder.Serialize()
	if serializedError != nil {
		logHistorySerializationErrorEvent(e.logger, serializedError, fmt.Sprintf(
			"History serialization error on start workflow.  WorkflowID: %v, RunID: %v", executionID, runID))
		return nil, serializedError
	}

	err0 := e.shard.AppendHistoryEvents(&persistence.AppendHistoryEventsRequest{
		Execution:    workflowExecution,
		FirstEventID: startedEvent.GetEventId(),
		Events:       events,
	})
	if err0 != nil {
		return nil, err0
	}

	id, err1 := e.tracker.getNextTaskID()
	if err1 != nil {
		return nil, err1
	}
	defer e.tracker.completeTask(id)
	_, err := e.shard.CreateWorkflowExecution(&persistence.CreateWorkflowExecutionRequest{
		RequestID:            request.GetRequestId(),
		Execution:            workflowExecution,
		TaskList:             request.GetTaskList().GetName(),
		DecisionTimeoutValue: request.GetTaskStartToCloseTimeoutSeconds(),
		ExecutionContext:     nil,
		NextEventID:          msBuilder.GetNextEventID(),
		LastProcessedEvent:   emptyEventID,
		TransferTasks: []persistence.Task{&persistence.DecisionTask{
			TaskID:   id,
			TaskList: taskList, ScheduleID: di.ScheduleID,
		}},
		DecisionScheduleID:          di.ScheduleID,
		DecisionStartedID:           di.StartedID,
		DecisionStartToCloseTimeout: di.DecisionTimeout,
	})

	if err != nil {
		// We created the history events but failed to create workflow execution, so cleanup the history which could cause
		// us to leak history events which are never cleaned up
		// TODO: Handle error on deletion of execution history
		e.historyMgr.DeleteWorkflowExecutionHistory(&persistence.DeleteWorkflowExecutionHistoryRequest{
			Execution: workflowExecution,
		})

		if t, ok := err.(*workflow.WorkflowExecutionAlreadyStartedError); ok {
			if t.GetStartRequestId() == request.GetRequestId() {
				return &workflow.StartWorkflowExecutionResponse{
					RunId: t.RunId,
				}, nil
			}
		}
		logPersistantStoreErrorEvent(e.logger, tagValueStoreOperationCreateWorkflowExecution, err,
			fmt.Sprintf("{WorkflowID: %v, RunID: %v}", executionID, runID))
		return nil, err
	}

	return &workflow.StartWorkflowExecutionResponse{
		RunId: workflowExecution.RunId,
	}, nil
}

// GetWorkflowExecutionHistory retrieves the history for given workflow execution
func (e *historyEngineImpl) GetWorkflowExecutionHistory(
	request *workflow.GetWorkflowExecutionHistoryRequest) (*workflow.GetWorkflowExecutionHistoryResponse, error) {
	execution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(request.GetExecution().GetWorkflowId()),
		RunId:      common.StringPtr(request.GetExecution().GetRunId()),
	}

	context, err0 := e.cache.getOrCreateWorkflowExecution(execution)
	if err0 != nil {
		return nil, err0
	}

	context.Lock()
	defer context.Unlock()
	msBuilder, err1 := context.loadWorkflowExecution()
	if err1 != nil {
		return nil, err1
	}

	nextPageToken := []byte{}
	historyEvents := []*workflow.HistoryEvent{}
Pagination_Loop:
	for {
		response, _ := e.historyMgr.GetWorkflowExecutionHistory(&persistence.GetWorkflowExecutionHistoryRequest{
			Execution:     execution,
			NextEventID:   msBuilder.GetNextEventID(),
			PageSize:      100,
			NextPageToken: nextPageToken,
		})

		for _, data := range response.Events {
			events, _ := e.hSerializer.Deserialize(data)
			historyEvents = append(historyEvents, events...)
		}

		if len(response.NextPageToken) == 0 {
			break Pagination_Loop
		}

		nextPageToken = response.NextPageToken
	}

	history := workflow.NewHistory()
	history.Events = historyEvents
	result := workflow.NewGetWorkflowExecutionHistoryResponse()
	result.History = history

	return result, nil
}

func (e *historyEngineImpl) RecordDecisionTaskStarted(
	request *h.RecordDecisionTaskStartedRequest) (*h.RecordDecisionTaskStartedResponse, error) {
	context, err0 := e.cache.getOrCreateWorkflowExecution(*request.WorkflowExecution)
	if err0 != nil {
		return nil, err0
	}

	context.Lock()
	defer context.Unlock()
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
			// Reload workflow execution history
			context.clear()
			continue Update_History_Loop
		}

		// Check execution state to make sure task is in the list of outstanding tasks and it is not yet started.  If
		// task is not outstanding than it is most probably a duplicate and complete the task.
		di, isRunning := msBuilder.GetPendingDecision(scheduleID)

		if !isRunning {
			// Looks like DecisionTask already completed as a result of another call.
			// It is OK to drop the task at this point.
			logDuplicateTaskEvent(context.logger, persistence.TransferTaskTypeDecisionTask, request.GetTaskId(), requestID,
				scheduleID, emptyEventID, isRunning)

			return nil, &workflow.EntityNotExistsError{Message: "Decision task not found."}
		}

		if di.StartedID != emptyEventID {
			// If decision is started as part of the current request scope then return a positive response
			if di.RequestID == requestID {
				return e.createRecordDecisionTaskStartedResponse(msBuilder, di.StartedID), nil
			}

			// Looks like DecisionTask already started as a result of another call.
			// It is OK to drop the task at this point.
			logDuplicateTaskEvent(context.logger, persistence.TaskListTypeDecision, request.GetTaskId(), requestID,
				scheduleID, di.StartedID, isRunning)

			return nil, &workflow.EntityNotExistsError{Message: "Decision task already started."}
		}

		event := msBuilder.AddDecisionTaskStartedEvent(scheduleID, requestID, request.PollRequest)
		if event == nil {
			// Unable to add DecisionTaskStarted event to history
			return nil, &workflow.InternalServiceError{Message: "Unable to add DecisionTaskStarted event to history."}
		}

		// Start a timer for the decision task.
		timeOutTask := context.tBuilder.AddDecisionTimoutTask(scheduleID, di.DecisionTimeout)
		timerTasks := []persistence.Task{timeOutTask}
		defer e.timerProcessor.NotifyNewTimer(timeOutTask.GetTaskID())

		// We apply the update to execution using optimistic concurrency.  If it fails due to a conflict than reload
		// the history and try the operation again.
		if err2 := context.updateWorkflowExecution(nil, timerTasks); err2 != nil {
			if err2 == ErrConflict {
				continue Update_History_Loop
			}

			return nil, err2
		}

		return e.createRecordDecisionTaskStartedResponse(msBuilder, event.GetEventId()), nil
	}

	return nil, ErrMaxAttemptsExceeded
}

func (e *historyEngineImpl) RecordActivityTaskStarted(
	request *h.RecordActivityTaskStartedRequest) (*h.RecordActivityTaskStartedResponse, error) {
	context, err0 := e.cache.getOrCreateWorkflowExecution(*request.WorkflowExecution)
	if err0 != nil {
		return nil, err0
	}

	context.Lock()
	defer context.Unlock()
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
			// Reload workflow execution history
			context.clear()
			continue Update_History_Loop
		}

		// Check execution state to make sure task is in the list of outstanding tasks and it is not yet started.  If
		// task is not outstanding than it is most probably a duplicate and complete the task.
		ai, isRunning := msBuilder.GetActivityInfo(scheduleID)
		if !isRunning {
			// Looks like ActivityTask already completed as a result of another call.
			// It is OK to drop the task at this point.
			logDuplicateTaskEvent(context.logger, persistence.TransferTaskTypeActivityTask, request.GetTaskId(), requestID,
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
			logDuplicateTaskEvent(context.logger, persistence.TransferTaskTypeActivityTask, request.GetTaskId(), requestID,
				scheduleID, ai.StartedID, isRunning)

			return nil, &workflow.EntityNotExistsError{Message: "Activity task already started."}
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
		defer e.timerProcessor.NotifyNewTimer(start2CloseTimeoutTask.GetTaskID())

		start2HeartBeatTimeoutTask, err := context.tBuilder.AddHeartBeatActivityTimeout(ai)
		if err != nil {
			return nil, err
		}
		if start2HeartBeatTimeoutTask != nil {
			timerTasks = append(timerTasks, start2HeartBeatTimeoutTask)
			defer e.timerProcessor.NotifyNewTimer(start2HeartBeatTimeoutTask.GetTaskID())
		}

		// We apply the update to execution using optimistic concurrency.  If it fails due to a conflict than reload
		// the history and try the operationi again.
		if err2 := context.updateWorkflowExecution(nil, timerTasks); err2 != nil {
			if err2 == ErrConflict {
				continue Update_History_Loop
			}

			return nil, err2
		}

		response := h.NewRecordActivityTaskStartedResponse()
		response.ScheduledEvent = scheduledEvent
		response.StartedEvent = startedEvent
		return response, nil
	}

	return nil, ErrMaxAttemptsExceeded
}

// RespondDecisionTaskCompleted completes a decision task
func (e *historyEngineImpl) RespondDecisionTaskCompleted(request *workflow.RespondDecisionTaskCompletedRequest) error {
	token, err0 := e.tokenSerializer.Deserialize(request.GetTaskToken())
	if err0 != nil {
		return &workflow.BadRequestError{Message: "Error deserializing task token."}
	}

	workflowExecution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(token.WorkflowID),
		RunId:      common.StringPtr(token.RunID),
	}

	context, err0 := e.cache.getOrCreateWorkflowExecution(workflowExecution)
	if err0 != nil {
		return err0
	}

	context.Lock()
	defer context.Unlock()

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
			// Reload workflow execution history
			context.clear()
			continue Update_History_Loop
		}

		di, isRunning := msBuilder.GetPendingDecision(scheduleID)
		if !isRunning || di.StartedID == emptyEventID {
			return &workflow.EntityNotExistsError{Message: "Decision task not found."}
		}

		startedID := di.StartedID
		completedEvent := msBuilder.AddDecisionTaskCompletedEvent(scheduleID, startedID, request)
		if completedEvent == nil {
			return &workflow.InternalServiceError{Message: "Unable to add DecisionTaskCompleted event to history."}
		}

		completedID := completedEvent.GetEventId()
		isComplete := false
		transferTasks := []persistence.Task{}
		timerTasks := []persistence.Task{}

	Process_Decision_Loop:
		for _, d := range request.Decisions {
			switch d.GetDecisionType() {
			case workflow.DecisionType_ScheduleActivityTask:
				attributes := d.GetScheduleActivityTaskDecisionAttributes()
				// TODO: We cannot fail the decision.  Append ActivityTaskScheduledFailed and continue processing
				if attributes.GetStartToCloseTimeoutSeconds() <= 0 {
					return &workflow.BadRequestError{Message: "Missing StartToCloseTimeoutSeconds in the activity scheduling parameters."}
				}
				if attributes.GetScheduleToStartTimeoutSeconds() <= 0 {
					return &workflow.BadRequestError{Message: "Missing ScheduleToStartTimeoutSeconds in the activity scheduling parameters."}
				}
				if attributes.GetScheduleToCloseTimeoutSeconds() <= 0 {
					return &workflow.BadRequestError{Message: "Missing ScheduleToCloseTimeoutSeconds in the activity scheduling parameters."}
				}
				if attributes.GetHeartbeatTimeoutSeconds() < 0 {
					// Sanity check on server. HeartBeat of Zero is allowed.
					return &workflow.BadRequestError{Message: "Invalid HeartbeatTimeoutSeconds value in the activity scheduling parameters."}
				}

				scheduleEvent, ai := msBuilder.AddActivityTaskScheduledEvent(completedID, attributes)
				id, err2 := e.tracker.getNextTaskID()
				if err2 != nil {
					return err2
				}
				defer e.tracker.completeTask(id)
				transferTasks = append(transferTasks, &persistence.ActivityTask{
					TaskID:     id,
					TaskList:   attributes.GetTaskList().GetName(),
					ScheduleID: scheduleEvent.GetEventId(),
				})

				// Create activity timeouts.
				Schedule2StartTimeoutTask := context.tBuilder.AddScheduleToStartActivityTimeout(ai)
				timerTasks = append(timerTasks, Schedule2StartTimeoutTask)
				defer e.timerProcessor.NotifyNewTimer(Schedule2StartTimeoutTask.GetTaskID())

				Schedule2CloseTimeoutTask, err := context.tBuilder.AddScheduleToCloseActivityTimeout(ai)
				if err != nil {
					return err
				}
				timerTasks = append(timerTasks, Schedule2CloseTimeoutTask)
				defer e.timerProcessor.NotifyNewTimer(Schedule2CloseTimeoutTask.GetTaskID())

			case workflow.DecisionType_CompleteWorkflowExecution:
				if isComplete || msBuilder.hasPendingTasks() {
					msBuilder.AddCompleteWorkflowExecutionFailedEvent(completedID,
						workflow.WorkflowCompleteFailedCause_UNHANDLED_DECISION)
					continue Process_Decision_Loop
				}
				attributes := d.GetCompleteWorkflowExecutionDecisionAttributes()
				msBuilder.AddCompletedWorkflowEvent(completedID, attributes)
				isComplete = true
			case workflow.DecisionType_FailWorkflowExecution:
				if isComplete || msBuilder.hasPendingTasks() {
					msBuilder.AddCompleteWorkflowExecutionFailedEvent(completedID,
						workflow.WorkflowCompleteFailedCause_UNHANDLED_DECISION)
					continue Process_Decision_Loop
				}
				attributes := d.GetFailWorkflowExecutionDecisionAttributes()
				msBuilder.AddFailWorkflowEvent(completedID, attributes)
				isComplete = true
			case workflow.DecisionType_StartTimer:
				attributes := d.GetStartTimerDecisionAttributes()
				_, ti := msBuilder.AddTimerStartedEvent(completedID, attributes)
				nextTimerTask := context.tBuilder.AddUserTimer(ti, msBuilder)
				if nextTimerTask != nil {
					timerTasks = append(timerTasks, nextTimerTask)
					defer e.timerProcessor.NotifyNewTimer(nextTimerTask.GetTaskID())
				}
			case workflow.DecisionType_RequestCancelActivityTask:
				attributes := d.GetRequestCancelActivityTaskDecisionAttributes()
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
				attributes := d.GetCancelTimerDecisionAttributes()
				if msBuilder.AddTimerCanceledEvent(completedID, attributes, request.GetIdentity()) == nil {
					msBuilder.AddCancelTimerFailedEvent(completedID, attributes, request.GetIdentity())
				}

			default:
				return &workflow.BadRequestError{Message: fmt.Sprintf("Unknown decision type: %v", d.GetDecisionType())}
			}
		}

		// Schedule another decision task if new events came in during this decision
		if (completedID - startedID) > 1 {
			newDecisionEvent, _ := msBuilder.AddDecisionTaskScheduledEvent()
			id, err2 := e.tracker.getNextTaskID()
			if err2 != nil {
				return err2
			}
			defer e.tracker.completeTask(id)
			transferTasks = append(transferTasks, &persistence.DecisionTask{
				TaskID:     id,
				TaskList:   newDecisionEvent.GetDecisionTaskScheduledEventAttributes().GetTaskList().GetName(),
				ScheduleID: newDecisionEvent.GetEventId(),
			})
		}

		if isComplete {
			// Generate a transfer task to delete workflow execution
			id, err2 := e.tracker.getNextTaskID()
			if err2 != nil {
				return err2
			}
			defer e.tracker.completeTask(id)
			transferTasks = append(transferTasks, &persistence.DeleteExecutionTask{
				TaskID: id,
			})
		}

		// We apply the update to execution using optimistic concurrency.  If it fails due to a conflict then reload
		// the history and try the operation again.
		if err := context.updateWorkflowExecutionWithContext(request.GetExecutionContext(), transferTasks, timerTasks); err != nil {
			if err == ErrConflict {
				continue Update_History_Loop
			}

			return err
		}

		return nil
	}

	return ErrMaxAttemptsExceeded
}

// RespondActivityTaskCompleted completes an activity task.
func (e *historyEngineImpl) RespondActivityTaskCompleted(request *workflow.RespondActivityTaskCompletedRequest) error {
	token, err0 := e.tokenSerializer.Deserialize(request.GetTaskToken())
	if err0 != nil {
		return &workflow.BadRequestError{Message: "Error deserializing task token."}
	}

	workflowExecution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(token.WorkflowID),
		RunId:      common.StringPtr(token.RunID),
	}

	context, err0 := e.cache.getOrCreateWorkflowExecution(workflowExecution)
	if err0 != nil {
		return err0
	}

	context.Lock()
	defer context.Unlock()

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
			// Reload workflow execution history
			context.clear()
			continue Update_History_Loop
		}

		ai, isRunning := msBuilder.GetActivityInfo(scheduleID)
		if !isRunning || ai.StartedID == emptyEventID {
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
			id, err2 := e.tracker.getNextTaskID()
			if err2 != nil {
				return err2
			}
			defer e.tracker.completeTask(id)
			transferTasks = []persistence.Task{&persistence.DecisionTask{
				TaskID:     id,
				TaskList:   newDecisionEvent.GetDecisionTaskScheduledEventAttributes().GetTaskList().GetName(),
				ScheduleID: newDecisionEvent.GetEventId(),
			}}
		}

		// We apply the update to execution using optimistic concurrency.  If it fails due to a conflict than reload
		// the history and try the operation again.
		if err := context.updateWorkflowExecution(transferTasks, nil); err != nil {
			if err == ErrConflict {
				continue Update_History_Loop
			}

			return err
		}

		return nil
	}

	return ErrMaxAttemptsExceeded
}

// RespondActivityTaskFailed completes an activity task failure.
func (e *historyEngineImpl) RespondActivityTaskFailed(request *workflow.RespondActivityTaskFailedRequest) error {
	token, err0 := e.tokenSerializer.Deserialize(request.GetTaskToken())
	if err0 != nil {
		return &workflow.BadRequestError{Message: "Error deserializing task token."}
	}

	workflowExecution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(token.WorkflowID),
		RunId:      common.StringPtr(token.RunID),
	}

	context, err0 := e.cache.getOrCreateWorkflowExecution(workflowExecution)
	if err0 != nil {
		return err0
	}

	context.Lock()
	defer context.Unlock()

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
			// Reload workflow execution history
			context.clear()
			continue Update_History_Loop
		}

		ai, isRunning := msBuilder.GetActivityInfo(scheduleID)
		if !isRunning || ai.StartedID == emptyEventID {
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
			id, err2 := e.tracker.getNextTaskID()
			if err2 != nil {
				return err2
			}
			defer e.tracker.completeTask(id)
			transferTasks = []persistence.Task{&persistence.DecisionTask{
				TaskID:     id,
				TaskList:   newDecisionEvent.GetDecisionTaskScheduledEventAttributes().GetTaskList().GetName(),
				ScheduleID: newDecisionEvent.GetEventId(),
			}}
		}

		// We apply the update to execution using optimistic concurrency.  If it fails due to a conflict than reload
		// the history and try the operation again.
		if err := context.updateWorkflowExecution(transferTasks, nil); err != nil {
			if err == ErrConflict {
				continue Update_History_Loop
			}

			return err
		}

		return nil
	}

	return ErrMaxAttemptsExceeded
}

// RespondActivityTaskCanceled completes an activity task failure.
func (e *historyEngineImpl) RespondActivityTaskCanceled(request *workflow.RespondActivityTaskCanceledRequest) error {
	token, err0 := e.tokenSerializer.Deserialize(request.GetTaskToken())
	if err0 != nil {
		return &workflow.BadRequestError{Message: "Error deserializing task token."}
	}

	workflowExecution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(token.WorkflowID),
		RunId:      common.StringPtr(token.RunID),
	}

	context, err0 := e.cache.getOrCreateWorkflowExecution(workflowExecution)
	if err0 != nil {
		return err0
	}

	context.Lock()
	defer context.Unlock()

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
			// Reload workflow execution history
			context.clear()
			continue Update_History_Loop
		}

		// Check execution state to make sure task is in the list of outstanding tasks and it is not yet started.  If
		// task is not outstanding than it is most probably a duplicate and complete the task.
		ai, isRunning := msBuilder.GetActivityInfo(scheduleID)
		if !isRunning || ai.StartedID == emptyEventID {
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
			id, err2 := e.tracker.getNextTaskID()
			if err2 != nil {
				return err2
			}

			defer e.tracker.completeTask(id)
			transferTasks = []persistence.Task{&persistence.DecisionTask{
				TaskID:     id,
				TaskList:   newDecisionEvent.GetDecisionTaskScheduledEventAttributes().GetTaskList().GetName(),
				ScheduleID: newDecisionEvent.GetEventId(),
			}}
		}

		// We apply the update to execution using optimistic concurrency.  If it fails due to a conflict than reload
		// the history and try the operation again.
		if err := context.updateWorkflowExecution(transferTasks, nil); err != nil {
			if err == ErrConflict {
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
	request *workflow.RecordActivityTaskHeartbeatRequest) (*workflow.RecordActivityTaskHeartbeatResponse, error) {
	token, err0 := e.tokenSerializer.Deserialize(request.GetTaskToken())
	if err0 != nil {
		return nil, &workflow.BadRequestError{Message: "Error deserializing task token."}
	}

	workflowExecution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(token.WorkflowID),
		RunId:      common.StringPtr(token.RunID),
	}

	context, err0 := e.cache.getOrCreateWorkflowExecution(workflowExecution)
	if err0 != nil {
		return nil, err0
	}

	context.Lock()
	defer context.Unlock()

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
			// Reload workflow execution history
			context.clear()
			continue Update_History_Loop
		}

		ai, isRunning := msBuilder.GetActivityInfo(scheduleID)
		if !isRunning || ai.StartedID == emptyEventID {
			e.logger.Debugf("Activity HeartBeat: scheduleEventID: %v, ActivityInfo: %+v, Exist: %v",
				scheduleID, ai, isRunning)
			return nil, &workflow.EntityNotExistsError{Message: "Activity task not found."}
		}

		cancelRequested := ai.CancelRequested

		var timerTasks []persistence.Task
		var transferTasks []persistence.Task

		e.logger.Debugf("Activity HeartBeat: scheduleEventID: %v, ActivityInfo: %+v, CancelRequested: %v",
			scheduleID, ai, cancelRequested)

		// Re-schedule next heartbeat.
		start2HeartBeatTimeoutTask, _ := context.tBuilder.AddHeartBeatActivityTimeout(ai)
		if start2HeartBeatTimeoutTask != nil {
			timerTasks = append(timerTasks, start2HeartBeatTimeoutTask)
			defer e.timerProcessor.NotifyNewTimer(start2HeartBeatTimeoutTask.GetTaskID())
		}

		// Save progress reported.
		msBuilder.updateActivityProgress(ai, request.GetDetails())

		// We apply the update to execution using optimistic concurrency.  If it fails due to a conflict than reload
		// the history and try the operation again.
		if err := context.updateWorkflowExecution(transferTasks, timerTasks); err != nil {
			if err == ErrConflict {
				continue Update_History_Loop
			}

			return nil, err
		}

		return &workflow.RecordActivityTaskHeartbeatResponse{CancelRequested: common.BoolPtr(cancelRequested)}, nil
	}

	return &workflow.RecordActivityTaskHeartbeatResponse{}, ErrMaxAttemptsExceeded
}

func (e *historyEngineImpl) createRecordDecisionTaskStartedResponse(msBuilder *mutableStateBuilder,
	startedEventID int64) *h.RecordDecisionTaskStartedResponse {
	response := h.NewRecordDecisionTaskStartedResponse()
	response.WorkflowType = msBuilder.getWorkflowType()
	if msBuilder.previousDecisionStartedEvent() != emptyEventID {
		response.PreviousStartedEventId = common.Int64Ptr(msBuilder.previousDecisionStartedEvent())
	}
	response.StartedEventId = common.Int64Ptr(startedEventID)

	return response
}

func (t *pendingTaskTracker) getNextTaskID() (int64, error) {
	t.lk.Lock()
	defer t.lk.Unlock()

	nextID, err := t.shard.GetNextTransferTaskID()
	if err != nil {
		t.logger.Debugf("Error generating next taskID: %v", err)
		return -1, err
	}

	if nextID != t.maxID+1 {
		t.logger.Fatalf("No holes allowed for nextID.  nextID: %v, MaxID: %v", nextID, t.maxID)
	}
	t.pendingTasks[nextID] = false
	t.maxID = nextID

	t.logger.Debugf("Generated new transfer task ID: %v", nextID)
	return nextID, nil
}

func (t *pendingTaskTracker) completeTask(taskID int64) {
	t.lk.Lock()
	updatedMin := int64(-1)
	if _, ok := t.pendingTasks[taskID]; ok {
		t.logger.Debugf("Completing transfer task ID: %v, minID: %v, maxID: %v", taskID, t.minID, t.maxID)
		t.pendingTasks[taskID] = true

	UpdateMinLoop:
		for newMin := t.minID + 1; newMin <= t.maxID; newMin++ {
			t.logger.Debugf("minID: %v, maxID: %v", newMin, t.maxID)
			if done, ok := t.pendingTasks[newMin]; ok && done {
				t.minID = newMin
				updatedMin = newMin
				delete(t.pendingTasks, newMin)
			} else {
				break UpdateMinLoop
			}
		}
	}

	t.lk.Unlock()

	if updatedMin != -1 {
		t.txProcessor.UpdateMaxAllowedReadLevel(updatedMin)
	}
}
