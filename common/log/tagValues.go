package log

// Pre-defined values for TagWorkflowListFilterType
const (
	TagValueListWorkflowFilterByID     = "WID"
	TagValueListWorkflowFilterByType   = "WType"
	TagValueListWorkflowFilterByStatus = "status"
)

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

// Pre-defined values for TagSysLifecycle
const (
	// History Engine lifecycle
	TagValueHistoryEngineStarting     = "HistoryEngineStarting"
	TagValueHistoryEngineStarted      = "HistoryEngineStarted"
	TagValueHistoryEngineShuttingDown = "HistoryEngineShuttingDown"
	TagValueHistoryEngineShutdown     = "HistoryEngineShutdown"

	// Transfer Queue Processor lifecycle
	TagValueTransferQueueProcessorStarting         = "TransferQueueProcessorStarting"
	TagValueTransferQueueProcessorStarted          = "TransferQueueProcessorStarted"
	TagValueTransferQueueProcessorShuttingDown     = "TransferQueueProcessorShuttingDown"
	TagValueTransferQueueProcessorShutdown         = "TransferQueueProcessorShutdown"
	TagValueTransferQueueProcessorShutdownTimedout = "TransferQueueProcessorShutdownTimedout"
	TagValueTransferTaskProcessingFailed           = "TransferTaskProcessingFailed"

	// ShardController lifecycle
	TagValueShardControllerStarted          = "ShardControllerStarted"
	TagValueShardControllerShutdown         = "ShardControllerShutdown"
	TagValueShardControllerShuttingDown     = "ShardControllerShuttingDown"
	TagValueShardControllerShutdownTimedout = "ShardControllerShutdownTimedout"
	TagValueRingMembershipChangedEvent      = "RingMembershipChangedEvent"
	TagValueShardClosedEvent                = "ShardClosedEvent"
	TagValueShardItemCreated                = "ShardItemCreated"
	TagValueShardItemRemoved                = "ShardItemRemoved"
	TagValueShardEngineCreating             = "ShardEngineCreating"
	TagValueShardEngineCreated              = "ShardEngineCreated"
	TagValueShardEngineStopping             = "ShardEngineStopping"
	TagValueShardEngineStopped              = "ShardEngineStopped"

	// Worker lifecycle
	TagValueReplicationTaskProcessorStarting         = "ReplicationTaskProcessorStarting"
	TagValueReplicationTaskProcessorStarted          = "ReplicationTaskProcessorStarted"
	TagValueReplicationTaskProcessorStartFailed      = "ReplicationTaskProcessorStartFailed"
	TagValueReplicationTaskProcessorShuttingDown     = "ReplicationTaskProcessorShuttingDown"
	TagValueReplicationTaskProcessorShutdown         = "ReplicationTaskProcessorShutdown"
	TagValueReplicationTaskProcessorShutdownTimedout = "ReplicationTaskProcessorShutdownTimedout"
	TagValueReplicationTaskProcessingFailed          = "ReplicationTaskProcessingFailed"
	TagValueIndexProcessorStarting                   = "IndexProcessorStarting"
	TagValueIndexProcessorStarted                    = "IndexProcessorStarted"
	TagValueIndexProcessorStartFailed                = "IndexProcessorStartFailed"
	TagValueIndexProcessorShuttingDown               = "IndexProcessorShuttingDown"
	TagValueIndexProcessorShutDown                   = "IndexProcessorShutDown"
	TagValueIndexProcessorShuttingDownTimedout       = "IndexProcessorShuttingDownTimedout"
)

// Pre-defined values for TagSysMajorEvent
const (
	// Errors
	TagValueInvalidHistoryAction        = "InvalidHistoryAction"
	TagValueInvalidQueryTask            = "InvalidQueryTask"
	TagValueQueryTaskFailed             = "QueryTaskFailed"
	TagValuePersistentStoreError        = "PersistentStoreError"
	TagValueHistorySerializationError   = "HistorySerializationError"
	TagValueHistoryDeserializationError = "HistoryDeserializationError"
	TagValueDuplicateTask               = "DuplicateTask"
	TagValueMultipleCompletionDecisions = "MultipleCompletionDecisions"
	TagValueDuplicateTransferTask       = "DuplicateTransferTask"
	TagValueDecisionFailed              = "DecisionFailed"
	TagValueInvalidMutableStateAction   = "InvalidMutableStateAction"

	// tasklist
	TagValueTaskListLoading       = "TaskListLoading"
	TagValueTaskListLoaded        = "TaskListLoaded"
	TagValueTaskListUnloading     = "TaskListUnloading"
	TagValueTaskListUnloaded      = "TaskListUnloaded"
	TagValueTaskListLoadingFailed = "TaskListLoadingFailed"

	// Shard context events
	TagValueShardRangeUpdated            = "ShardRangeUpdated"
	TagValueShardAllocateTimerBeforeRead = "ShardAllocateTimerBeforeRead"

	// General purpose events
	TagValueOperationFailed   = "OperationFailed"
	TagValueOperationStuck    = "OperationStuck"
	TagValueOperationCritical = "OperationCritical"
)

// Pre-defined values for TagSysComponent
const (
	TagValueHistoryBuilderComponent           = "history-builder"
	TagValueHistoryEngineComponent            = "history-engine"
	TagValueHistoryCacheComponent             = "history-cache"
	TagValueEventsCacheComponent              = "events-cache"
	TagValueTransferQueueComponent            = "transfer-queue-processor"
	TagValueTimerQueueComponent               = "timer-queue-processor"
	TagValueReplicatorQueueComponent          = "replicator-queue-processor"
	TagValueShardController                   = "shard-controller"
	TagValueMatchingEngineComponent           = "matching-engine"
	TagValueReplicatorComponent               = "replicator"
	TagValueReplicationTaskProcessorComponent = "replication-task-processor"
	TagValueHistoryReplicatorComponent        = "history-replicator"
	TagValueIndexerComponent                  = "indexer"
	TagValueIndexerProcessorComponent         = "indexer-processor"
	TagValueIndexerESProcessorComponent       = "indexer-es-processor"
	TagValueESVisibilityManager               = "es-visibility-manager"
	TagValueArchiverComponent                 = "archiver"
)

// Pre-defined values for TagSysStoreOperation
const (
	TagValueStoreOperationGetTasks                = "get-tasks"
	TagValueStoreOperationCompleteTask            = "complete-task"
	TagValueStoreOperationCompleteTasksLessThan   = "complete-tasks-less-than"
	TagValueStoreOperationCreateWorkflowExecution = "create-wf-execution"
	TagValueStoreOperationGetWorkflowExecution    = "get-wf-execution"
	TagValueStoreOperationUpdateWorkflowExecution = "update-wf-execution"
	TagValueStoreOperationDeleteWorkflowExecution = "delete-wf-execution"
	TagValueStoreOperationUpdateShard             = "update-shard"
	TagValueStoreOperationCreateTask              = "create-task"
	TagValueStoreOperationUpdateTaskList          = "update-task-list"
	TagValueStoreOperationStopTaskList            = "stop-task-list"
)
