package log

// Pre-defined values for TagWorkflowAction
const (
	TagValueActionWorkflowStarted                 = "add-workflowexecution-started-event"
	TagValueActionDecisionTaskScheduled           = "add-decisiontask-scheduled-event"
	TagValueActionDecisionTaskStarted             = "add-decisiontask-started-event"
	TagValueActionDecisionTaskCompleted           = "add-decisiontask-completed-event"
	TagValueActionDecisionTaskTimedOut            = "add-decisiontask-timedout-event"
	TagValueActionDecisionTaskFailed              = "add-decisiontask-failed-event"
	TagValueActionActivityTaskScheduled           = "add-activitytask-scheduled-event"
	TagValueActionActivityTaskStarted             = "add-activitytask-started-event"
	TagValueActionActivityTaskCompleted           = "add-activitytask-completed-event"
	TagValueActionActivityTaskFailed              = "add-activitytask-failed-event"
	TagValueActionActivityTaskTimedOut            = "add-activitytask-timed-event"
	TagValueActionActivityTaskCanceled            = "add-activitytask-canceled-event"
	TagValueActionActivityTaskCancelRequest       = "add-activitytask-cancel-request-event"
	TagValueActionActivityTaskCancelRequestFailed = "add-activitytask-cancel-request-failed-event"
	TagValueActionCompleteWorkflow                = "add-complete-workflow-event"
	TagValueActionFailWorkflow                    = "add-fail-workflow-event"
	TagValueActionTimeoutWorkflow                 = "add-timeout-workflow-event"
	TagValueActionCancelWorkflow                  = "add-cancel-workflow-event"
	TagValueActionTimerStarted                    = "add-timer-started-event"
	TagValueActionTimerFired                      = "add-timer-fired-event"
	TagValueActionTimerCanceled                   = "add-timer-Canceled-event"
	TagValueActionWorkflowTerminated              = "add-workflowexecution-terminated-event"
	TagValueActionWorkflowSignaled                = "add-workflowexecution-signaled-event"
	TagValueActionContinueAsNew                   = "add-continue-as-new-event"
	TagValueActionWorkflowCanceled                = "add-workflowexecution-canceled-event"
	TagValueActionChildExecutionStarted           = "add-childexecution-started-event"
	TagValueActionStartChildExecutionFailed       = "add-start-childexecution-failed-event"
	TagValueActionChildExecutionCompleted         = "add-childexecution-completed-event"
	TagValueActionChildExecutionFailed            = "add-childexecution-failed-event"
	TagValueActionChildExecutionCanceled          = "add-childexecution-canceled-event"
	TagValueActionChildExecutionTerminated        = "add-childexecution-terminated-event"
	TagValueActionChildExecutionTimedOut          = "add-childexecution-timedout-event"
	TagValueActionRequestCancelWorkflow           = "add-request-cancel-workflow-event"
	TagValueActionWorkflowCancelRequested         = "add-workflow-execution-cancel-requested-event"
	TagValueActionWorkflowCancelFailed            = "add-workflow-execution-cancel-failed-event"
	TagValueActionWorkflowSignalRequested         = "add-workflow-execution-signal-requested-event"
	TagValueActionWorkflowSignalFailed            = "add-workflow-execution-signal-failed-event"
	TagValueActionUnknownEvent                    = "add-unknown-event"
)

// Pre-defined values for TagWorkflowListFilterType
const (
	TagValueListWorkflowFilterByID     = "WID"
	TagValueListWorkflowFilterByType   = "WType"
	TagValueListWorkflowFilterByStatus = "status"
)

// Pre-defined values for TagSysComponent
const (
	TagValueComponentTaskList                 = "tasklist"
	TagValueComponentHistoryBuilder           = "history-builder"
	TagValueComponentHistoryEngine            = "history-engine"
	TagValueComponentHistoryCache             = "history-cache"
	TagValueComponentEventsCache              = "events-cache"
	TagValueComponentTransferQueue            = "transfer-queue-processor"
	TagValueComponentTimerQueue               = "timer-queue-processor"
	TagValueComponentReplicatorQueue          = "replicator-queue-processor"
	TagValueComponentShardController          = "shard-controller"
	TagValueComponentShard                    = "shard"
	TagValueComponentShardItem                = "shard-item"
	TagValueComponentShardEngine              = "shard-engine"
	TagValueComponentMatchingEngine           = "matching-engine"
	TagValueComponentReplicator               = "replicator"
	TagValueComponentReplicationTaskProcessor = "replication-task-processor"
	TagValueComponentHistoryReplicator        = "history-replicator"
	TagValueComponentIndexer                  = "indexer"
	TagValueComponentIndexerProcessor         = "indexer-processor"
	TagValueComponentIndexerESProcessor       = "indexer-es-processor"
	TagValueComponentESVisibilityManager      = "es-visibility-manager"
	TagValueComponentArchiver                 = "archiver"
)

// Pre-defined values for TagSysLifecycle
const (
	TagValueLifeCycleStarting         tagValueTypeSysLifecycle = "Starting"
	TagValueLifeCycleStarted          tagValueTypeSysLifecycle = "Started"
	TagValueLifeCycleStopping         tagValueTypeSysLifecycle = "Stopping"
	TagValueLifeCycleStopped          tagValueTypeSysLifecycle = "Stopped"
	TagValueLifeCycleStopTimedout     tagValueTypeSysLifecycle = "StopTimedout"
	TagValueLifeCycleStartFailed      tagValueTypeSysLifecycle = "StartFailed"
	TagValueLifeCycleStopFailed       tagValueTypeSysLifecycle = "StopFailed"
	TagValueLifeCycleProcessingFailed tagValueTypeSysLifecycle = "ProcessingFailed"
)

// Pre-defined values for SysErrorType
const (
	TagValueInvalidHistoryAction        tagValueTypeSysErrorType = "InvalidHistoryAction"
	TagValueInvalidQueryTask            tagValueTypeSysErrorType = "InvalidQueryTask"
	TagValueQueryTaskFailed             tagValueTypeSysErrorType = "QueryTaskFailed"
	TagValuePersistentStoreError        tagValueTypeSysErrorType = "PersistentStoreError"
	TagValueHistorySerializationError   tagValueTypeSysErrorType = "HistorySerializationError"
	TagValueHistoryDeserializationError tagValueTypeSysErrorType = "HistoryDeserializationError"
	TagValueDuplicateTask               tagValueTypeSysErrorType = "DuplicateTask"
	TagValueMultipleCompletionDecisions tagValueTypeSysErrorType = "MultipleCompletionDecisions"
	TagValueDuplicateTransferTask       tagValueTypeSysErrorType = "DuplicateTransferTask"
	TagValueDecisionFailed              tagValueTypeSysErrorType = "DecisionFailed"
	TagValueInvalidMutableStateAction   tagValueTypeSysErrorType = "InvalidMutableStateAction"
)

// Pre-defined values for SysShardUpdate
const (
	// Shard context events
	TagValueShardRangeUpdated            tagValueTypeSysShardUpdate = "ShardRangeUpdated"
	TagValueShardAllocateTimerBeforeRead tagValueTypeSysShardUpdate = "ShardAllocateTimerBeforeRead"
	TagValueRingMembershipChangedEvent   tagValueTypeSysShardUpdate = "RingMembershipChangedEvent"
)

// Pre-defined values for OperationResult
const (
	TagValueOperationFailed   tagValueTypeOperationResult = "OperationFailed"
	TagValueOperationStuck    tagValueTypeOperationResult = "OperationStuck"
	TagValueOperationCritical tagValueTypeOperationResult = "OperationCritical"
)

// Pre-defined values for TagSysStoreOperation
const (
	TagValueStoreOperationGetTasks                tagValueTypeSysStoreOperation = "get-tasks"
	TagValueStoreOperationCompleteTask            tagValueTypeSysStoreOperation = "complete-task"
	TagValueStoreOperationCompleteTasksLessThan   tagValueTypeSysStoreOperation = "complete-tasks-less-than"
	TagValueStoreOperationCreateWorkflowExecution tagValueTypeSysStoreOperation = "create-wf-execution"
	TagValueStoreOperationGetWorkflowExecution    tagValueTypeSysStoreOperation = "get-wf-execution"
	TagValueStoreOperationUpdateWorkflowExecution tagValueTypeSysStoreOperation = "update-wf-execution"
	TagValueStoreOperationDeleteWorkflowExecution tagValueTypeSysStoreOperation = "delete-wf-execution"
	TagValueStoreOperationUpdateShard             tagValueTypeSysStoreOperation = "update-shard"
	TagValueStoreOperationCreateTask              tagValueTypeSysStoreOperation = "create-task"
	TagValueStoreOperationUpdateTaskList          tagValueTypeSysStoreOperation = "update-task-list"
	TagValueStoreOperationStopTaskList            tagValueTypeSysStoreOperation = "stop-task-list"
)
