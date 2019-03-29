package log

// Pre-defined values for TagWorkflowAction
const (
	TagValueActionWorkflowStarted                 tagValueTypeWorkflowAction = "add-workflowexecution-started-event"
	TagValueActionDecisionTaskScheduled           tagValueTypeWorkflowAction = "add-decisiontask-scheduled-event"
	TagValueActionDecisionTaskStarted             tagValueTypeWorkflowAction = "add-decisiontask-started-event"
	TagValueActionDecisionTaskCompleted           tagValueTypeWorkflowAction = "add-decisiontask-completed-event"
	TagValueActionDecisionTaskTimedOut            tagValueTypeWorkflowAction = "add-decisiontask-timedout-event"
	TagValueActionDecisionTaskFailed              tagValueTypeWorkflowAction = "add-decisiontask-failed-event"
	TagValueActionActivityTaskScheduled           tagValueTypeWorkflowAction = "add-activitytask-scheduled-event"
	TagValueActionActivityTaskStarted             tagValueTypeWorkflowAction = "add-activitytask-started-event"
	TagValueActionActivityTaskCompleted           tagValueTypeWorkflowAction = "add-activitytask-completed-event"
	TagValueActionActivityTaskFailed              tagValueTypeWorkflowAction = "add-activitytask-failed-event"
	TagValueActionActivityTaskTimedOut            tagValueTypeWorkflowAction = "add-activitytask-timed-event"
	TagValueActionActivityTaskCanceled            tagValueTypeWorkflowAction = "add-activitytask-canceled-event"
	TagValueActionActivityTaskCancelRequest       tagValueTypeWorkflowAction = "add-activitytask-cancel-request-event"
	TagValueActionActivityTaskCancelRequestFailed tagValueTypeWorkflowAction = "add-activitytask-cancel-request-failed-event"
	TagValueActionCompleteWorkflow                tagValueTypeWorkflowAction = "add-complete-workflow-event"
	TagValueActionFailWorkflow                    tagValueTypeWorkflowAction = "add-fail-workflow-event"
	TagValueActionTimeoutWorkflow                 tagValueTypeWorkflowAction = "add-timeout-workflow-event"
	TagValueActionCancelWorkflow                  tagValueTypeWorkflowAction = "add-cancel-workflow-event"
	TagValueActionTimerStarted                    tagValueTypeWorkflowAction = "add-timer-started-event"
	TagValueActionTimerFired                      tagValueTypeWorkflowAction = "add-timer-fired-event"
	TagValueActionTimerCanceled                   tagValueTypeWorkflowAction = "add-timer-Canceled-event"
	TagValueActionWorkflowTerminated              tagValueTypeWorkflowAction = "add-workflowexecution-terminated-event"
	TagValueActionWorkflowSignaled                tagValueTypeWorkflowAction = "add-workflowexecution-signaled-event"
	TagValueActionContinueAsNew                   tagValueTypeWorkflowAction = "add-continue-as-new-event"
	TagValueActionWorkflowCanceled                tagValueTypeWorkflowAction = "add-workflowexecution-canceled-event"
	TagValueActionChildExecutionStarted           tagValueTypeWorkflowAction = "add-childexecution-started-event"
	TagValueActionStartChildExecutionFailed       tagValueTypeWorkflowAction = "add-start-childexecution-failed-event"
	TagValueActionChildExecutionCompleted         tagValueTypeWorkflowAction = "add-childexecution-completed-event"
	TagValueActionChildExecutionFailed            tagValueTypeWorkflowAction = "add-childexecution-failed-event"
	TagValueActionChildExecutionCanceled          tagValueTypeWorkflowAction = "add-childexecution-canceled-event"
	TagValueActionChildExecutionTerminated        tagValueTypeWorkflowAction = "add-childexecution-terminated-event"
	TagValueActionChildExecutionTimedOut          tagValueTypeWorkflowAction = "add-childexecution-timedout-event"
	TagValueActionRequestCancelWorkflow           tagValueTypeWorkflowAction = "add-request-cancel-workflow-event"
	TagValueActionWorkflowCancelRequested         tagValueTypeWorkflowAction = "add-workflow-execution-cancel-requested-event"
	TagValueActionWorkflowCancelFailed            tagValueTypeWorkflowAction = "add-workflow-execution-cancel-failed-event"
	TagValueActionWorkflowSignalRequested         tagValueTypeWorkflowAction = "add-workflow-execution-signal-requested-event"
	TagValueActionWorkflowSignalFailed            tagValueTypeWorkflowAction = "add-workflow-execution-signal-failed-event"
	TagValueActionUnknownEvent                    tagValueTypeWorkflowAction = "add-unknown-event"
)

// Pre-defined values for TagWorkflowListFilterType
const (
	TagValueListWorkflowFilterByID     tagValueTypeWorkflowListFilterType = "WID"
	TagValueListWorkflowFilterByType   tagValueTypeWorkflowListFilterType = "WType"
	TagValueListWorkflowFilterByStatus tagValueTypeWorkflowListFilterType = "status"
)

// Pre-defined values for TagSysComponent
const (
	TagValueComponentTaskList                 tagValueTypeSysComponent = "tasklist"
	TagValueComponentHistoryBuilder           tagValueTypeSysComponent = "history-builder"
	TagValueComponentHistoryEngine            tagValueTypeSysComponent = "history-engine"
	TagValueComponentHistoryCache             tagValueTypeSysComponent = "history-cache"
	TagValueComponentEventsCache              tagValueTypeSysComponent = "events-cache"
	TagValueComponentTransferQueue            tagValueTypeSysComponent = "transfer-queue-processor"
	TagValueComponentTimerQueue               tagValueTypeSysComponent = "timer-queue-processor"
	TagValueComponentReplicatorQueue          tagValueTypeSysComponent = "replicator-queue-processor"
	TagValueComponentShardController          tagValueTypeSysComponent = "shard-controller"
	TagValueComponentShard                    tagValueTypeSysComponent = "shard"
	TagValueComponentShardItem                tagValueTypeSysComponent = "shard-item"
	TagValueComponentShardEngine              tagValueTypeSysComponent = "shard-engine"
	TagValueComponentMatchingEngine           tagValueTypeSysComponent = "matching-engine"
	TagValueComponentReplicator               tagValueTypeSysComponent = "replicator"
	TagValueComponentReplicationTaskProcessor tagValueTypeSysComponent = "replication-task-processor"
	TagValueComponentHistoryReplicator        tagValueTypeSysComponent = "history-replicator"
	TagValueComponentIndexer                  tagValueTypeSysComponent = "indexer"
	TagValueComponentIndexerProcessor         tagValueTypeSysComponent = "indexer-processor"
	TagValueComponentIndexerESProcessor       tagValueTypeSysComponent = "indexer-es-processor"
	TagValueComponentESVisibilityManager      tagValueTypeSysComponent = "es-visibility-manager"
	TagValueComponentArchiver                 tagValueTypeSysComponent = "archiver"
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
