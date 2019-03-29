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
	TagValueHistoryEngineStarting     TagValueTypeSysLifecycle = "HistoryEngineStarting"
	TagValueHistoryEngineStarted      TagValueTypeSysLifecycle = "HistoryEngineStarted"
	TagValueHistoryEngineShuttingDown TagValueTypeSysLifecycle = "HistoryEngineShuttingDown"
	TagValueHistoryEngineShutdown     TagValueTypeSysLifecycle = "HistoryEngineShutdown"

	// Transfer Queue Processor lifecycle
	TagValueTransferQueueProcessorStarting         TagValueTypeSysLifecycle = "TransferQueueProcessorStarting"
	TagValueTransferQueueProcessorStarted          TagValueTypeSysLifecycle = "TransferQueueProcessorStarted"
	TagValueTransferQueueProcessorShuttingDown     TagValueTypeSysLifecycle = "TransferQueueProcessorShuttingDown"
	TagValueTransferQueueProcessorShutdown         TagValueTypeSysLifecycle = "TransferQueueProcessorShutdown"
	TagValueTransferQueueProcessorShutdownTimedout TagValueTypeSysLifecycle = "TransferQueueProcessorShutdownTimedout"
	TagValueTransferTaskProcessingFailed           TagValueTypeSysLifecycle = "TransferTaskProcessingFailed"

	// ShardController lifecycle
	TagValueShardControllerStarted          TagValueTypeSysLifecycle = "ShardControllerStarted"
	TagValueShardControllerShutdown         TagValueTypeSysLifecycle = "ShardControllerShutdown"
	TagValueShardControllerShuttingDown     TagValueTypeSysLifecycle = "ShardControllerShuttingDown"
	TagValueShardControllerShutdownTimedout TagValueTypeSysLifecycle = "ShardControllerShutdownTimedout"
	TagValueRingMembershipChangedEvent      TagValueTypeSysLifecycle = "RingMembershipChangedEvent"
	TagValueShardClosedEvent                TagValueTypeSysLifecycle = "ShardClosedEvent"
	TagValueShardItemCreated                TagValueTypeSysLifecycle = "ShardItemCreated"
	TagValueShardItemRemoved                TagValueTypeSysLifecycle = "ShardItemRemoved"
	TagValueShardEngineCreating             TagValueTypeSysLifecycle = "ShardEngineCreating"
	TagValueShardEngineCreated              TagValueTypeSysLifecycle = "ShardEngineCreated"
	TagValueShardEngineStopping             TagValueTypeSysLifecycle = "ShardEngineStopping"
	TagValueShardEngineStopped              TagValueTypeSysLifecycle = "ShardEngineStopped"

	// Worker lifecycle
	TagValueReplicationTaskProcessorStarting         TagValueTypeSysLifecycle = "ReplicationTaskProcessorStarting"
	TagValueReplicationTaskProcessorStarted          TagValueTypeSysLifecycle = "ReplicationTaskProcessorStarted"
	TagValueReplicationTaskProcessorStartFailed      TagValueTypeSysLifecycle = "ReplicationTaskProcessorStartFailed"
	TagValueReplicationTaskProcessorShuttingDown     TagValueTypeSysLifecycle = "ReplicationTaskProcessorShuttingDown"
	TagValueReplicationTaskProcessorShutdown         TagValueTypeSysLifecycle = "ReplicationTaskProcessorShutdown"
	TagValueReplicationTaskProcessorShutdownTimedout TagValueTypeSysLifecycle = "ReplicationTaskProcessorShutdownTimedout"
	TagValueReplicationTaskProcessingFailed          TagValueTypeSysLifecycle = "ReplicationTaskProcessingFailed"
	TagValueIndexProcessorStarting                   TagValueTypeSysLifecycle = "IndexProcessorStarting"
	TagValueIndexProcessorStarted                    TagValueTypeSysLifecycle = "IndexProcessorStarted"
	TagValueIndexProcessorStartFailed                TagValueTypeSysLifecycle = "IndexProcessorStartFailed"
	TagValueIndexProcessorShuttingDown               TagValueTypeSysLifecycle = "IndexProcessorShuttingDown"
	TagValueIndexProcessorShutDown                   TagValueTypeSysLifecycle = "IndexProcessorShutDown"
	TagValueIndexProcessorShuttingDownTimedout       TagValueTypeSysLifecycle = "IndexProcessorShuttingDownTimedout"
)

// Pre-defined values for TagSysMajorEvent

const (
	// Errors
	TagValueInvalidHistoryAction        TagValueTypeSysMajorEvent = "InvalidHistoryAction"
	TagValueInvalidQueryTask            TagValueTypeSysMajorEvent = "InvalidQueryTask"
	TagValueQueryTaskFailed             TagValueTypeSysMajorEvent = "QueryTaskFailed"
	TagValuePersistentStoreError        TagValueTypeSysMajorEvent = "PersistentStoreError"
	TagValueHistorySerializationError   TagValueTypeSysMajorEvent = "HistorySerializationError"
	TagValueHistoryDeserializationError TagValueTypeSysMajorEvent = "HistoryDeserializationError"
	TagValueDuplicateTask               TagValueTypeSysMajorEvent = "DuplicateTask"
	TagValueMultipleCompletionDecisions TagValueTypeSysMajorEvent = "MultipleCompletionDecisions"
	TagValueDuplicateTransferTask       TagValueTypeSysMajorEvent = "DuplicateTransferTask"
	TagValueDecisionFailed              TagValueTypeSysMajorEvent = "DecisionFailed"
	TagValueInvalidMutableStateAction   TagValueTypeSysMajorEvent = "InvalidMutableStateAction"

	// tasklist
	TagValueTaskListLoading       TagValueTypeSysMajorEvent = "TaskListLoading"
	TagValueTaskListLoaded        TagValueTypeSysMajorEvent = "TaskListLoaded"
	TagValueTaskListUnloading     TagValueTypeSysMajorEvent = "TaskListUnloading"
	TagValueTaskListUnloaded      TagValueTypeSysMajorEvent = "TaskListUnloaded"
	TagValueTaskListLoadingFailed TagValueTypeSysMajorEvent = "TaskListLoadingFailed"

	// Shard context events
	TagValueShardRangeUpdated            TagValueTypeSysMajorEvent = "ShardRangeUpdated"
	TagValueShardAllocateTimerBeforeRead TagValueTypeSysMajorEvent = "ShardAllocateTimerBeforeRead"

	// General purpose events
	TagValueOperationFailed   TagValueTypeSysMajorEvent = "OperationFailed"
	TagValueOperationStuck    TagValueTypeSysMajorEvent = "OperationStuck"
	TagValueOperationCritical TagValueTypeSysMajorEvent = "OperationCritical"
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
