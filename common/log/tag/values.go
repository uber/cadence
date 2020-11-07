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

package tag

// Pre-defined values for TagWorkflowAction
var (
	// workflow start / finish
	WorkflowActionWorkflowStarted       = workflowAction("add-workflow-started-event")
	WorkflowActionWorkflowCanceled      = workflowAction("add-workflow-canceled-event")
	WorkflowActionWorkflowCompleted     = workflowAction("add-workflow-completed--event")
	WorkflowActionWorkflowFailed        = workflowAction("add-workflow-failed-event")
	WorkflowActionWorkflowTimeout       = workflowAction("add-workflow-timeout-event")
	WorkflowActionWorkflowTerminated    = workflowAction("add-workflow-terminated-event")
	WorkflowActionWorkflowContinueAsNew = workflowAction("add-workflow-continue-as-new-event")

	// workflow cancellation / sign
	WorkflowActionWorkflowCancelRequested        = workflowAction("add-workflow-cancel-requested-event")
	WorkflowActionWorkflowSignaled               = workflowAction("add-workflow-signaled-event")
	WorkflowActionWorkflowRecordMarker           = workflowAction("add-workflow-marker-record-event")
	WorkflowActionUpsertWorkflowSearchAttributes = workflowAction("add-workflow-upsert-search-attributes-event")

	// decision
	WorkflowActionDecisionTaskScheduled = workflowAction("add-decisiontask-scheduled-event")
	WorkflowActionDecisionTaskStarted   = workflowAction("add-decisiontask-started-event")
	WorkflowActionDecisionTaskCompleted = workflowAction("add-decisiontask-completed-event")
	WorkflowActionDecisionTaskTimedOut  = workflowAction("add-decisiontask-timedout-event")
	WorkflowActionDecisionTaskFailed    = workflowAction("add-decisiontask-failed-event")

	// in memory decision
	WorkflowActionInMemoryDecisionTaskScheduled = workflowAction("add-in-memory-decisiontask-scheduled")
	WorkflowActionInMemoryDecisionTaskStarted   = workflowAction("add-in-memory-decisiontask-started")

	// activity
	WorkflowActionActivityTaskScheduled       = workflowAction("add-activitytask-scheduled-event")
	WorkflowActionActivityTaskStarted         = workflowAction("add-activitytask-started-event")
	WorkflowActionActivityTaskCompleted       = workflowAction("add-activitytask-completed-event")
	WorkflowActionActivityTaskFailed          = workflowAction("add-activitytask-failed-event")
	WorkflowActionActivityTaskTimedOut        = workflowAction("add-activitytask-timed-event")
	WorkflowActionActivityTaskCanceled        = workflowAction("add-activitytask-canceled-event")
	WorkflowActionActivityTaskCancelRequested = workflowAction("add-activitytask-cancel-requested-event")
	WorkflowActionActivityTaskCancelFailed    = workflowAction("add-activitytask-cancel-failed-event")
	WorkflowActionActivityTaskRetry           = workflowAction("add-activitytask-retry-event")

	// timer
	WorkflowActionTimerStarted      = workflowAction("add-timer-started-event")
	WorkflowActionTimerFired        = workflowAction("add-timer-fired-event")
	WorkflowActionTimerCanceled     = workflowAction("add-timer-canceled-event")
	WorkflowActionTimerCancelFailed = workflowAction("add-timer-cancel-failed-event")

	// child workflow start / finish
	WorkflowActionChildWorkflowInitiated        = workflowAction("add-childworkflow-initiated-event")
	WorkflowActionChildWorkflowStarted          = workflowAction("add-childworkflow-started-event")
	WorkflowActionChildWorkflowInitiationFailed = workflowAction("add-childworkflow-initiation-failed-event")
	WorkflowActionChildWorkflowCanceled         = workflowAction("add-childworkflow-canceled-event")
	WorkflowActionChildWorkflowCompleted        = workflowAction("add-childworkflow-completed-event")
	WorkflowActionChildWorkflowFailed           = workflowAction("add-childworkflow-failed-event")
	WorkflowActionChildWorkflowTerminated       = workflowAction("add-childworkflow-terminated-event")
	WorkflowActionChildWorkflowTimedOut         = workflowAction("add-childworkflow-timedout-event")

	// external workflow cancellation
	WorkflowActionExternalWorkflowCancelInitiated = workflowAction("add-externalworkflow-cancel-initiated-event")
	WorkflowActionExternalWorkflowCancelRequested = workflowAction("add-externalworkflow-cancel-requested-event")
	WorkflowActionExternalWorkflowCancelFailed    = workflowAction("add-externalworkflow-cancel-failed-event")

	// external workflow signal
	WorkflowActionExternalWorkflowSignalInitiated = workflowAction("add-externalworkflow-signal-initiated-event")
	WorkflowActionExternalWorkflowSignalRequested = workflowAction("add-externalworkflow-signal-requested-event")
	WorkflowActionExternalWorkflowSignalFailed    = workflowAction("add-externalworkflow-signal-failed-event")

	WorkflowActionUnknown = workflowAction("add-unknown-event")
)

// Pre-defined values for TagWorkflowListFilterType
var (
	WorkflowListWorkflowFilterByID     = workflowListFilterType("WID")
	WorkflowListWorkflowFilterByType   = workflowListFilterType("WType")
	WorkflowListWorkflowFilterByStatus = workflowListFilterType("status")
)

// Pre-defined values for TagSysComponent
var (
	ComponentTaskList                 = component("tasklist")
	ComponentHistoryEngine            = component("history-engine")
	ComponentHistoryCache             = component("history-cache")
	ComponentEventsCache              = component("events-cache")
	ComponentTransferQueue            = component("transfer-queue-processor")
	ComponentTimerQueue               = component("timer-queue-processor")
	ComponentTimerBuilder             = component("timer-builder")
	ComponentReplicatorQueue          = component("replicator-queue-processor")
	ComponentShardController          = component("shard-controller")
	ComponentShard                    = component("shard")
	ComponentShardItem                = component("shard-item")
	ComponentShardEngine              = component("shard-engine")
	ComponentMatchingEngine           = component("matching-engine")
	ComponentReplicator               = component("replicator")
	ComponentReplicationTaskProcessor = component("replication-task-processor")
	ComponentReplicationAckManager    = component("replication-ack-manager")
	ComponentHistoryReplicator        = component("history-replicator")
	ComponentHistoryResender          = component("history-resender")
	ComponentIndexer                  = component("indexer")
	ComponentIndexerProcessor         = component("indexer-processor")
	ComponentIndexerESProcessor       = component("indexer-es-processor")
	ComponentESVisibilityManager      = component("es-visibility-manager")
	ComponentArchiver                 = component("archiver")
	ComponentBatcher                  = component("batcher")
	ComponentWorker                   = component("worker")
	ComponentServiceResolver          = component("service-resolver")
	ComponentFailoverCoordinator      = component("failover-coordinator")
	ComponentFailoverMarkerNotifier   = component("failover-marker-notifier")
)

// Pre-defined values for TagSysLifecycle
var (
	LifeCycleStarting         = lifecycle("Starting")
	LifeCycleStarted          = lifecycle("Started")
	LifeCycleStopping         = lifecycle("Stopping")
	LifeCycleStopped          = lifecycle("Stopped")
	LifeCycleStopTimedout     = lifecycle("StopTimedout")
	LifeCycleStartFailed      = lifecycle("StartFailed")
	LifeCycleStopFailed       = lifecycle("StopFailed")
	LifeCycleProcessingFailed = lifecycle("ProcessingFailed")
)

// Pre-defined values for SysErrorType
var (
	ErrorTypeInvalidHistoryAction         = errorType("InvalidHistoryAction")
	ErrorTypeInvalidQueryTask             = errorType("InvalidQueryTask")
	ErrorTypeQueryTaskFailed              = errorType("QueryTaskFailed")
	ErrorTypePersistentStoreError         = errorType("PersistentStoreError")
	ErrorTypeHistorySerializationError    = errorType("HistorySerializationError")
	ErrorTypeHistoryDeserializationError  = errorType("HistoryDeserializationError")
	ErrorTypeDuplicateTask                = errorType("DuplicateTask")
	ErrorTypeMultipleCompletionDecisions  = errorType("MultipleCompletionDecisions")
	ErrorTypeDuplicateTransferTask        = errorType("DuplicateTransferTask")
	ErrorTypeDecisionFailed               = errorType("DecisionFailed")
	ErrorTypeInvalidMutableStateAction    = errorType("InvalidMutableStateAction")
	ErrorTypeInvalidMemDecisionTaskAction = errorType("InvalidMemDecisionTaskAction")
)

// Pre-defined values for SysShardUpdate
var (
	// Shard context events
	ValueShardRangeUpdated            = shardupdate("ShardRangeUpdated")
	ValueShardAllocateTimerBeforeRead = shardupdate("ShardAllocateTimerBeforeRead")
	ValueRingMembershipChangedEvent   = shardupdate("RingMembershipChangedEvent")
)

// Pre-defined values for OperationResult
var (
	OperationFailed   = operationResult("OperationFailed")
	OperationStuck    = operationResult("OperationStuck")
	OperationCritical = operationResult("OperationCritical")
)

// Pre-defined values for TagSysStoreOperation
var (
	StoreOperationCreateShard = storeOperation("create-shard")
	StoreOperationGetShard    = storeOperation("get-shard")
	StoreOperationUpdateShard = storeOperation("update-shard")

	StoreOperationCreateWorkflowExecution           = storeOperation("create-wf-execution")
	StoreOperationGetWorkflowExecution              = storeOperation("get-wf-execution")
	StoreOperationUpdateWorkflowExecution           = storeOperation("update-wf-execution")
	StoreOperationConflictResolveWorkflowExecution  = storeOperation("conflict-resolve-wf-execution")
	StoreOperationResetWorkflowExecution            = storeOperation("reset-wf-execution")
	StoreOperationDeleteWorkflowExecution           = storeOperation("delete-wf-execution")
	StoreOperationDeleteCurrentWorkflowExecution    = storeOperation("delete-current-wf-execution")
	StoreOperationGetCurrentExecution               = storeOperation("get-current-execution")
	StoreOperationListCurrentExecution              = storeOperation("list-current-execution")
	StoreOperationIsWorkflowExecutionExists         = storeOperation("is-wf-execution-exists")
	StoreOperationListConcreteExecution             = storeOperation("list-concrete-execution")
	StoreOperationGetTransferTasks                  = storeOperation("get-transfer-tasks")
	StoreOperationGetReplicationTasks               = storeOperation("get-replication-tasks")
	StoreOperationCompleteTransferTask              = storeOperation("complete-transfer-task")
	StoreOperationRangeCompleteTransferTask         = storeOperation("range-complete-transfer-task")
	StoreOperationCompleteReplicationTask           = storeOperation("complete-replication-task")
	StoreOperationRangeCompleteReplicationTask      = storeOperation("range-complete-replication-task")
	StoreOperationPutReplicationTaskToDLQ           = storeOperation("put-replication-task-to-dlq")
	StoreOperationGetReplicationTasksFromDLQ        = storeOperation("get-replication-tasks-from-dlq")
	StoreOperationGetReplicationDLQSize             = storeOperation("get-replication-dlq-size")
	StoreOperationDeleteReplicationTaskFromDLQ      = storeOperation("delete-replication-task-from-dlq")
	StoreOperationRangeDeleteReplicationTaskFromDLQ = storeOperation("range-delete-replication-task-from-dlq")
	StoreOperationCreateFailoverMarkerTasks         = storeOperation("createFailoverMarkerTasks")
	StoreOperationGetTimerIndexTasks                = storeOperation("get-timer-index-tasks")
	StoreOperationCompleteTimerTask                 = storeOperation("complete-timer-task")
	StoreOperationRangeCompleteTimerTask            = storeOperation("range-complete-timer-task")

	StoreOperationCreateTasks           = storeOperation("create-tasks")
	StoreOperationGetTasks              = storeOperation("get-tasks")
	StoreOperationCompleteTask          = storeOperation("complete-task")
	StoreOperationCompleteTasksLessThan = storeOperation("complete-tasks-less-than")
	StoreOperationLeaseTaskList         = storeOperation("lease-task-list")
	StoreOperationUpdateTaskList        = storeOperation("update-task-list")
	StoreOperationListTaskList          = storeOperation("list-task-list")
	StoreOperationDeleteTaskList        = storeOperation("delete-task-list")
	StoreOperationStopTaskList          = storeOperation("stop-task-list")

	StoreOperationCreateDomain       = storeOperation("create-domain")
	StoreOperationGetDomain          = storeOperation("get-domain")
	StoreOperationUpdateDomain       = storeOperation("update-domain")
	StoreOperationDeleteDomain       = storeOperation("delete-domain")
	StoreOperationDeleteDomainByName = storeOperation("delete-domain-by-name")
	StoreOperationListDomains        = storeOperation("list-domains")
	StoreOperationGetMetadata        = storeOperation("get-metadata")

	StoreOperationRecordWorkflowExecutionStarted           = storeOperation("record-wf-execution-started")
	StoreOperationRecordWorkflowExecutionClosed            = storeOperation("record-wf-execution-closed")
	StoreOperationUpsertWorkflowExecution                  = storeOperation("upsert-wf-execution")
	StoreOperationListOpenWorkflowExecutions               = storeOperation("list-open-wf-executions")
	StoreOperationListClosedWorkflowExecutions             = storeOperation("list-closed-wf-executions")
	StoreOperationListOpenWorkflowExecutionsByType         = storeOperation("list-open-wf-executions-by-type")
	StoreOperationListClosedWorkflowExecutionsByType       = storeOperation("list-closed-wf-executions-by-type")
	StoreOperationListOpenWorkflowExecutionsByWorkflowID   = storeOperation("list-open-wf-executions-by-wfID")
	StoreOperationListClosedWorkflowExecutionsByWorkflowID = storeOperation("list-closed-wf-executions-by-wfID")
	StoreOperationListClosedWorkflowExecutionsByStatus     = storeOperation("list-closed-wf-executions-by-status")
	StoreOperationGetClosedWorkflowExecution               = storeOperation("get-closed-wf-execution")
	StoreOperationVisibilityDeleteWorkflowExecution        = storeOperation("vis-delete-wf-execution")
	StoreOperationListWorkflowExecutions                   = storeOperation("list-wf-executions")
	StoreOperationScanWorkflowExecutions                   = storeOperation("scan-wf-executions")
	StoreOperationCountWorkflowExecutions                  = storeOperation("count-wf-executions")

	StoreOperationAppendHistoryNodes        = storeOperation("append-history-nodes")
	StoreOperationReadHistoryBranch         = storeOperation("read-history-branch")
	StoreOperationReadHistoryBranchByBatch  = storeOperation("read-history-branch-by-batch")
	StoreOperationReadRawHistoryBranch      = storeOperation("read-raw-history-branch")
	StoreOperationForkHistoryBranch         = storeOperation("fork-history-branch")
	StoreOperationDeleteHistoryBranch       = storeOperation("delete-history-branch")
	StoreOperationGetHistoryTree            = storeOperation("get-history-tree")
	StoreOperationGetAllHistoryTreeBranches = storeOperation("get-all-history-tree-branches")

	StoreOperationEnqueueMessage             = storeOperation("enqueue-message")
	StoreOperationReadMessages               = storeOperation("read-messages")
	StoreOperationUpdateAckLevel             = storeOperation("update-ack-level")
	StoreOperationGetAckLevels               = storeOperation("get-ack-levels")
	StoreOperationDeleteMessagesBefore       = storeOperation("delete-messages-before")
	StoreOperationEnqueueMessageToDLQ        = storeOperation("enqueue-message-to-dlq")
	StoreOperationReadMessagesFromDLQ        = storeOperation("read-messages-from-dlq")
	StoreOperationRangeDeleteMessagesFromDLQ = storeOperation("range-delete-messages-from-dlq")
	StoreOperationUpdateDLQAckLevel          = storeOperation("update-dlq-ack-level")
	StoreOperationGetDLQAckLevels            = storeOperation("get-dlq-ack-levels")
	StoreOperationGetDLQSize                 = storeOperation("get-dlq-size")
	StoreOperationDeleteMessageFromDLQ       = storeOperation("delete-message-from-dlq")
)

// Pre-defined values for TagSysClientOperation
var (
	MatchingClientOperationAddActivityTask        = clientOperation("matching-add-activity-task")
	MatchingClientOperationAddDecisionTask        = clientOperation("matching-add-decision-task")
	MatchingClientOperationPollForActivityTask    = clientOperation("matching-poll-for-activity-task")
	MatchingClientOperationPollForDecisionTask    = clientOperation("matching-poll-for-decision-task")
	MatchingClientOperationQueryWorkflow          = clientOperation("matching-query-wf")
	MatchingClientOperationQueryTaskCompleted     = clientOperation("matching-query-task-completed")
	MatchingClientOperationCancelOutstandingPoll  = clientOperation("matching-cancel-outstanding-poll")
	MatchingClientOperationDescribeTaskList       = clientOperation("matching-describe-task-list")
	MatchingClientOperationListTaskListPartitions = clientOperation("matching-list-task-list-partitions")
)
