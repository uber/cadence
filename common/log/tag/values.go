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
	ValueActionWorkflowStarted                 = workflowAction("add-workflowexecution-started-event")
	ValueActionDecisionTaskScheduled           = workflowAction("add-decisiontask-scheduled-event")
	ValueActionDecisionTaskStarted             = workflowAction("add-decisiontask-started-event")
	ValueActionDecisionTaskCompleted           = workflowAction("add-decisiontask-completed-event")
	ValueActionDecisionTaskTimedOut            = workflowAction("add-decisiontask-timedout-event")
	ValueActionDecisionTaskFailed              = workflowAction("add-decisiontask-failed-event")
	ValueActionActivityTaskScheduled           = workflowAction("add-activitytask-scheduled-event")
	ValueActionActivityTaskStarted             = workflowAction("add-activitytask-started-event")
	ValueActionActivityTaskCompleted           = workflowAction("add-activitytask-completed-event")
	ValueActionActivityTaskFailed              = workflowAction("add-activitytask-failed-event")
	ValueActionActivityTaskTimedOut            = workflowAction("add-activitytask-timed-event")
	ValueActionActivityTaskCanceled            = workflowAction("add-activitytask-canceled-event")
	ValueActionActivityTaskCancelRequest       = workflowAction("add-activitytask-cancel-request-event")
	ValueActionActivityTaskCancelRequestFailed = workflowAction("add-activitytask-cancel-request-failed-event")
	ValueActionCompleteWorkflow                = workflowAction("add-complete-workflow-event")
	ValueActionFailWorkflow                    = workflowAction("add-fail-workflow-event")
	ValueActionTimeoutWorkflow                 = workflowAction("add-timeout-workflow-event")
	ValueActionCancelWorkflow                  = workflowAction("add-cancel-workflow-event")
	ValueActionTimerStarted                    = workflowAction("add-timer-started-event")
	ValueActionTimerFired                      = workflowAction("add-timer-fired-event")
	ValueActionTimerCanceled                   = workflowAction("add-timer-Canceled-event")
	ValueActionWorkflowTerminated              = workflowAction("add-workflowexecution-terminated-event")
	ValueActionWorkflowSignaled                = workflowAction("add-workflowexecution-signaled-event")
	ValueActionContinueAsNew                   = workflowAction("add-continue-as-new-event")
	ValueActionWorkflowCanceled                = workflowAction("add-workflowexecution-canceled-event")
	ValueActionChildExecutionStarted           = workflowAction("add-childexecution-started-event")
	ValueActionStartChildExecutionFailed       = workflowAction("add-start-childexecution-failed-event")
	ValueActionChildExecutionCompleted         = workflowAction("add-childexecution-completed-event")
	ValueActionChildExecutionFailed            = workflowAction("add-childexecution-failed-event")
	ValueActionChildExecutionCanceled          = workflowAction("add-childexecution-canceled-event")
	ValueActionChildExecutionTerminated        = workflowAction("add-childexecution-terminated-event")
	ValueActionChildExecutionTimedOut          = workflowAction("add-childexecution-timedout-event")
	ValueActionRequestCancelWorkflow           = workflowAction("add-request-cancel-workflow-event")
	ValueActionWorkflowCancelRequested         = workflowAction("add-workflow-execution-cancel-requested-event")
	ValueActionWorkflowCancelFailed            = workflowAction("add-workflow-execution-cancel-failed-event")
	ValueActionWorkflowSignalRequested         = workflowAction("add-workflow-execution-signal-requested-event")
	ValueActionWorkflowSignalFailed            = workflowAction("add-workflow-execution-signal-failed-event")
	ValueActionUnknownEvent                    = workflowAction("add-unknown-event")
)

// Pre-defined values for TagWorkflowListFilterType
var (
	ValueListWorkflowFilterByID     = workflowListFilterType("WID")
	ValueListWorkflowFilterByType   = workflowListFilterType("WType")
	ValueListWorkflowFilterByStatus = workflowListFilterType("status")
)

// Pre-defined values for TagSysComponent
var (
	ValueComponentTaskList                 = component("tasklist")
	ValueComponentHistoryBuilder           = component("history-builder")
	ValueComponentHistoryEngine            = component("history-engine")
	ValueComponentHistoryCache             = component("history-cache")
	ValueComponentEventsCache              = component("events-cache")
	ValueComponentTransferQueue            = component("transfer-queue-processor")
	ValueComponentTimerQueue               = component("timer-queue-processor")
	ValueComponentReplicatorQueue          = component("replicator-queue-processor")
	ValueComponentShardController          = component("shard-controller")
	ValueComponentShard                    = component("shard")
	ValueComponentShardItem                = component("shard-item")
	ValueComponentShardEngine              = component("shard-engine")
	ValueComponentMatchingEngine           = component("matching-engine")
	ValueComponentReplicator               = component("replicator")
	ValueComponentReplicationTaskProcessor = component("replication-task-processor")
	ValueComponentHistoryReplicator        = component("history-replicator")
	ValueComponentIndexer                  = component("indexer")
	ValueComponentIndexerProcessor         = component("indexer-processor")
	ValueComponentIndexerESProcessor       = component("indexer-es-processor")
	ValueComponentESVisibilityManager      = component("es-visibility-manager")
	ValueComponentArchiver                 = component("archiver")
)

// Pre-defined values for TagSysLifecycle
var (
	ValueLifeCycleStarting         = lifecycle("Starting")
	ValueLifeCycleStarted          = lifecycle("Started")
	ValueLifeCycleStopping         = lifecycle("Stopping")
	ValueLifeCycleStopped          = lifecycle("Stopped")
	ValueLifeCycleStopTimedout     = lifecycle("StopTimedout")
	ValueLifeCycleStartFailed      = lifecycle("StartFailed")
	ValueLifeCycleStopFailed       = lifecycle("StopFailed")
	ValueLifeCycleProcessingFailed = lifecycle("ProcessingFailed")
)

// Pre-defined values for SysErrorType
var (
	ValueInvalidHistoryAction        = errorType("InvalidHistoryAction")
	ValueInvalidQueryTask            = errorType("InvalidQueryTask")
	ValueQueryTaskFailed             = errorType("QueryTaskFailed")
	ValuePersistentStoreError        = errorType("PersistentStoreError")
	ValueHistorySerializationError   = errorType("HistorySerializationError")
	ValueHistoryDeserializationError = errorType("HistoryDeserializationError")
	ValueDuplicateTask               = errorType("DuplicateTask")
	ValueMultipleCompletionDecisions = errorType("MultipleCompletionDecisions")
	ValueDuplicateTransferTask       = errorType("DuplicateTransferTask")
	ValueDecisionFailed              = errorType("DecisionFailed")
	ValueInvalidMutableStateAction   = errorType("InvalidMutableStateAction")
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
	ValueSysOperationFailed   = operationResult("OperationFailed")
	ValueSysOperationStuck    = operationResult("OperationStuck")
	ValueSysOperationCritical = operationResult("OperationCritical")
)

// Pre-defined values for TagSysStoreOperation
var (
	ValueStoreOperationGetTasks                = storeOperation("get-tasks")
	ValueStoreOperationCompleteTask            = storeOperation("complete-task")
	ValueStoreOperationCompleteTasksLessThan   = storeOperation("complete-tasks-less-than")
	ValueStoreOperationCreateWorkflowExecution = storeOperation("create-wf-execution")
	ValueStoreOperationGetWorkflowExecution    = storeOperation("get-wf-execution")
	ValueStoreOperationUpdateWorkflowExecution = storeOperation("update-wf-execution")
	ValueStoreOperationDeleteWorkflowExecution = storeOperation("delete-wf-execution")
	ValueStoreOperationUpdateShard             = storeOperation("update-shard")
	ValueStoreOperationCreateTask              = storeOperation("create-task")
	ValueStoreOperationUpdateTaskList          = storeOperation("update-task-list")
	ValueStoreOperationStopTaskList            = storeOperation("stop-task-list")
)
