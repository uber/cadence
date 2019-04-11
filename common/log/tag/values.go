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
const (
	ValueActionWorkflowStarted                 valueTypeWorkflowAction = "add-workflowexecution-started-event"
	ValueActionDecisionTaskScheduled           valueTypeWorkflowAction = "add-decisiontask-scheduled-event"
	ValueActionDecisionTaskStarted             valueTypeWorkflowAction = "add-decisiontask-started-event"
	ValueActionDecisionTaskCompleted           valueTypeWorkflowAction = "add-decisiontask-completed-event"
	ValueActionDecisionTaskTimedOut            valueTypeWorkflowAction = "add-decisiontask-timedout-event"
	ValueActionDecisionTaskFailed              valueTypeWorkflowAction = "add-decisiontask-failed-event"
	ValueActionActivityTaskScheduled           valueTypeWorkflowAction = "add-activitytask-scheduled-event"
	ValueActionActivityTaskStarted             valueTypeWorkflowAction = "add-activitytask-started-event"
	ValueActionActivityTaskCompleted           valueTypeWorkflowAction = "add-activitytask-completed-event"
	ValueActionActivityTaskFailed              valueTypeWorkflowAction = "add-activitytask-failed-event"
	ValueActionActivityTaskTimedOut            valueTypeWorkflowAction = "add-activitytask-timed-event"
	ValueActionActivityTaskCanceled            valueTypeWorkflowAction = "add-activitytask-canceled-event"
	ValueActionActivityTaskCancelRequest       valueTypeWorkflowAction = "add-activitytask-cancel-request-event"
	ValueActionActivityTaskCancelRequestFailed valueTypeWorkflowAction = "add-activitytask-cancel-request-failed-event"
	ValueActionCompleteWorkflow                valueTypeWorkflowAction = "add-complete-workflow-event"
	ValueActionFailWorkflow                    valueTypeWorkflowAction = "add-fail-workflow-event"
	ValueActionTimeoutWorkflow                 valueTypeWorkflowAction = "add-timeout-workflow-event"
	ValueActionCancelWorkflow                  valueTypeWorkflowAction = "add-cancel-workflow-event"
	ValueActionTimerStarted                    valueTypeWorkflowAction = "add-timer-started-event"
	ValueActionTimerFired                      valueTypeWorkflowAction = "add-timer-fired-event"
	ValueActionTimerCanceled                   valueTypeWorkflowAction = "add-timer-Canceled-event"
	ValueActionWorkflowTerminated              valueTypeWorkflowAction = "add-workflowexecution-terminated-event"
	ValueActionWorkflowSignaled                valueTypeWorkflowAction = "add-workflowexecution-signaled-event"
	ValueActionContinueAsNew                   valueTypeWorkflowAction = "add-continue-as-new-event"
	ValueActionWorkflowCanceled                valueTypeWorkflowAction = "add-workflowexecution-canceled-event"
	ValueActionChildExecutionStarted           valueTypeWorkflowAction = "add-childexecution-started-event"
	ValueActionStartChildExecutionFailed       valueTypeWorkflowAction = "add-start-childexecution-failed-event"
	ValueActionChildExecutionCompleted         valueTypeWorkflowAction = "add-childexecution-completed-event"
	ValueActionChildExecutionFailed            valueTypeWorkflowAction = "add-childexecution-failed-event"
	ValueActionChildExecutionCanceled          valueTypeWorkflowAction = "add-childexecution-canceled-event"
	ValueActionChildExecutionTerminated        valueTypeWorkflowAction = "add-childexecution-terminated-event"
	ValueActionChildExecutionTimedOut          valueTypeWorkflowAction = "add-childexecution-timedout-event"
	ValueActionRequestCancelWorkflow           valueTypeWorkflowAction = "add-request-cancel-workflow-event"
	ValueActionWorkflowCancelRequested         valueTypeWorkflowAction = "add-workflow-execution-cancel-requested-event"
	ValueActionWorkflowCancelFailed            valueTypeWorkflowAction = "add-workflow-execution-cancel-failed-event"
	ValueActionWorkflowSignalRequested         valueTypeWorkflowAction = "add-workflow-execution-signal-requested-event"
	ValueActionWorkflowSignalFailed            valueTypeWorkflowAction = "add-workflow-execution-signal-failed-event"
	ValueActionUnknownEvent                    valueTypeWorkflowAction = "add-unknown-event"
)

// Pre-defined values for TagWorkflowListFilterType
const (
	ValueListWorkflowFilterByID     valueTypeWorkflowListFilterType = "WID"
	ValueListWorkflowFilterByType   valueTypeWorkflowListFilterType = "WType"
	ValueListWorkflowFilterByStatus valueTypeWorkflowListFilterType = "status"
)

// Pre-defined values for TagSysComponent
const (
	ValueComponentTaskList                 valueTypeSysComponent = "tasklist"
	ValueComponentHistoryBuilder           valueTypeSysComponent = "history-builder"
	ValueComponentHistoryEngine            valueTypeSysComponent = "history-engine"
	ValueComponentHistoryCache             valueTypeSysComponent = "history-cache"
	ValueComponentEventsCache              valueTypeSysComponent = "events-cache"
	ValueComponentTransferQueue            valueTypeSysComponent = "transfer-queue-processor"
	ValueComponentTimerQueue               valueTypeSysComponent = "timer-queue-processor"
	ValueComponentReplicatorQueue          valueTypeSysComponent = "replicator-queue-processor"
	ValueComponentShardController          valueTypeSysComponent = "shard-controller"
	ValueComponentShard                    valueTypeSysComponent = "shard"
	ValueComponentShardItem                valueTypeSysComponent = "shard-item"
	ValueComponentShardEngine              valueTypeSysComponent = "shard-engine"
	ValueComponentMatchingEngine           valueTypeSysComponent = "matching-engine"
	ValueComponentReplicator               valueTypeSysComponent = "replicator"
	ValueComponentReplicationTaskProcessor valueTypeSysComponent = "replication-task-processor"
	ValueComponentHistoryReplicator        valueTypeSysComponent = "history-replicator"
	ValueComponentIndexer                  valueTypeSysComponent = "indexer"
	ValueComponentIndexerProcessor         valueTypeSysComponent = "indexer-processor"
	ValueComponentIndexerESProcessor       valueTypeSysComponent = "indexer-es-processor"
	ValueComponentESVisibilityManager      valueTypeSysComponent = "es-visibility-manager"
	ValueComponentArchiver                 valueTypeSysComponent = "archiver"
)

// Pre-defined values for TagSysLifecycle
const (
	ValueLifeCycleStarting         valueTypeSysLifecycle = "Starting"
	ValueLifeCycleStarted          valueTypeSysLifecycle = "Started"
	ValueLifeCycleStopping         valueTypeSysLifecycle = "Stopping"
	ValueLifeCycleStopped          valueTypeSysLifecycle = "Stopped"
	ValueLifeCycleStopTimedout     valueTypeSysLifecycle = "StopTimedout"
	ValueLifeCycleStartFailed      valueTypeSysLifecycle = "StartFailed"
	ValueLifeCycleStopFailed       valueTypeSysLifecycle = "StopFailed"
	ValueLifeCycleProcessingFailed valueTypeSysLifecycle = "ProcessingFailed"
)

// Pre-defined values for SysErrorType
const (
	ValueInvalidHistoryAction        valueTypeSysErrorType = "InvalidHistoryAction"
	ValueInvalidQueryTask            valueTypeSysErrorType = "InvalidQueryTask"
	ValueQueryTaskFailed             valueTypeSysErrorType = "QueryTaskFailed"
	ValuePersistentStoreError        valueTypeSysErrorType = "PersistentStoreError"
	ValueHistorySerializationError   valueTypeSysErrorType = "HistorySerializationError"
	ValueHistoryDeserializationError valueTypeSysErrorType = "HistoryDeserializationError"
	ValueDuplicateTask               valueTypeSysErrorType = "DuplicateTask"
	ValueMultipleCompletionDecisions valueTypeSysErrorType = "MultipleCompletionDecisions"
	ValueDuplicateTransferTask       valueTypeSysErrorType = "DuplicateTransferTask"
	ValueDecisionFailed              valueTypeSysErrorType = "DecisionFailed"
	ValueInvalidMutableStateAction   valueTypeSysErrorType = "InvalidMutableStateAction"
)

// Pre-defined values for SysShardUpdate
const (
	// Shard context events
	ValueShardRangeUpdated            valueTypeSysShardUpdate = "ShardRangeUpdated"
	ValueShardAllocateTimerBeforeRead valueTypeSysShardUpdate = "ShardAllocateTimerBeforeRead"
	ValueRingMembershipChangedEvent   valueTypeSysShardUpdate = "RingMembershipChangedEvent"
)

// Pre-defined values for OperationResult
const (
	ValueSysOperationFailed   valueTypeSysOperationResult = "OperationFailed"
	ValueSysOperationStuck    valueTypeSysOperationResult = "OperationStuck"
	ValueSysOperationCritical valueTypeSysOperationResult = "OperationCritical"
)

// Pre-defined values for TagSysStoreOperation
const (
	ValueStoreOperationGetTasks                valueTypeSysStoreOperation = "get-tasks"
	ValueStoreOperationCompleteTask            valueTypeSysStoreOperation = "complete-task"
	ValueStoreOperationCompleteTasksLessThan   valueTypeSysStoreOperation = "complete-tasks-less-than"
	ValueStoreOperationCreateWorkflowExecution valueTypeSysStoreOperation = "create-wf-execution"
	ValueStoreOperationGetWorkflowExecution    valueTypeSysStoreOperation = "get-wf-execution"
	ValueStoreOperationUpdateWorkflowExecution valueTypeSysStoreOperation = "update-wf-execution"
	ValueStoreOperationDeleteWorkflowExecution valueTypeSysStoreOperation = "delete-wf-execution"
	ValueStoreOperationUpdateShard             valueTypeSysStoreOperation = "update-shard"
	ValueStoreOperationCreateTask              valueTypeSysStoreOperation = "create-task"
	ValueStoreOperationUpdateTaskList          valueTypeSysStoreOperation = "update-task-list"
	ValueStoreOperationStopTaskList            valueTypeSysStoreOperation = "stop-task-list"
)
