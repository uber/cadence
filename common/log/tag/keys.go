package tag

import (
	"time"

	"github.com/uber/cadence/common/persistence"
)

// All logging tags are defined in this file.
// To help finding available tags, we recommend that all tags to be categorized and placed in the corresponding section.
// We currently have those categories:
//   0. Common tags that can't be categorized(or belong to more than one)
//   1. Workflow: these tags are information that are useful to our customer, like workflow-id/run-id/task-list/...
//   2. System : these tags are internal information which usually cannot be understood by our customers,

///////////////////  Common tags defined here ///////////////////
// Error returns tag for Error
func Error(err error) Tag {
	return newTag("error", err, ValueTypeError)
}

// ClusterName returns tag for ClusterName
func ClusterName(clusterName string) Tag {
	return newTag("cluster-name", clusterName, ValueTypeString)
}

// Timestamp returns tag for Timestamp
func Timestamp(timestamp time.Time) Tag {
	return newTag("timestamp", timestamp, ValueTypeTime)
}

///////////////////  Workflow tags defined here: ( wf is short for workflow) ///////////////////
// WorkflowAction returns tag for WorkflowAction
func WorkflowAction(action valueTypeWorkflowAction) Tag {
	return newTag("wf-action", action, ValueTypeWorkflowAction)
}

// WorkflowListFilterType returns tag for WorkflowListFilterType
func WorkflowListFilterType(listFilterType valueTypeWorkflowListFilterType) Tag {
	return newTag("wf-list-filter-type", listFilterType, ValueTypeWorkflowListFilterType)
}

// general
// WorkflowError returns tag for WorkflowError
func WorkflowError(error error) Tag { return newTag("wf-error", error, ValueTypeError) }

// WorkflowTimeoutType returns tag for WorkflowTimeoutType
func WorkflowTimeoutType(timeoutType int64) Tag {
	return newTag("wf-timeout-type", timeoutType, ValueTypeInteger)
}

// WorkflowPollContextTimeout returns tag for WorkflowPollContextTimeout
func WorkflowPollContextTimeout(pollContextTimeout time.Duration) Tag {
	return newTag("wf-poll-context-timeout", pollContextTimeout, ValueTypeDuration)
}

// WorkflowHandlerName returns tag for WorkflowHandlerName
func WorkflowHandlerName(handlerName string) Tag {
	return newTag("wf-handler-name", handlerName, ValueTypeString)
}

// WorkflowID returns tag for WorkflowID
func WorkflowID(workflowID string) Tag {
	return newTag("wf-id", workflowID, ValueTypeString)
}

// WorkflowType returns tag for WorkflowType
func WorkflowType(wfType string) Tag {
	return newTag("wf-type", wfType, ValueTypeString)
}

// WorkflowRunID returns tag for WorkflowRunID
func WorkflowRunID(runID string) Tag {
	return newTag("wf-run-id", runID, ValueTypeString)
}

// WorkflowBeginningRunID returns tag for WorkflowBeginningRunID
func WorkflowBeginningRunID(beginningRunID string) Tag {
	return newTag("wf-beginning-run-id", beginningRunID, ValueTypeString)
}

// WorkflowEndingRunID returns tag for WorkflowEndingRunID
func WorkflowEndingRunID(endingRunID string) Tag {
	return newTag("wf-ending-run-id", endingRunID, ValueTypeString)
}

// domain related
// WorkflowDomainID returns tag for WorkflowDomainID
func WorkflowDomainID(domainID string) Tag {
	return newTag("wf-domain-id", domainID, ValueTypeString)
}

// WorkflowDomainName returns tag for WorkflowDomainName
func WorkflowDomainName(domainName string) Tag {
	return newTag("wf-domain-name", domainName, ValueTypeString)
}

// WorkflowDomainIDs returns tag for WorkflowDomainIDs
func WorkflowDomainIDs(domainIDs []string) Tag {
	return newTag("wf-domain-ids", domainIDs, ValueTypeObject)
}

// history event ID related
// WorkflowEventID returns tag for WorkflowEventID
func WorkflowEventID(eventID int64) Tag {
	return newTag("wf-history-event-id", eventID, ValueTypeInteger)
}

// WorkflowScheduleID returns tag for WorkflowScheduleID
func WorkflowScheduleID(scheduleID int64) Tag {
	return newTag("wf-schedule-id", scheduleID, ValueTypeInteger)
}

// WorkflowFirstEventID returns tag for WorkflowFirstEventID
func WorkflowFirstEventID(firstEventID int64) Tag {
	return newTag("wf-first-event-id", firstEventID, ValueTypeInteger)
}

// WorkflowNextEventID returns tag for WorkflowNextEventID
func WorkflowNextEventID(nextEventID int64) Tag {
	return newTag("wf-next-event-id", nextEventID, ValueTypeInteger)
}

// WorkflowBeginningFirstEventID returns tag for WorkflowBeginningFirstEventID
func WorkflowBeginningFirstEventID(beginningFirstEventID int64) Tag {
	return newTag("wf-begining-first-event-id", beginningFirstEventID, ValueTypeInteger)
}

// WorkflowEndingNextEventID returns tag for WorkflowEndingNextEventID
func WorkflowEndingNextEventID(endingNextEventID int64) Tag {
	return newTag("wf-ending-next-event-id", endingNextEventID, ValueTypeInteger)
}

// WorkflowResetNextEventID returns tag for WorkflowResetNextEventID
func WorkflowResetNextEventID(resetNextEventID int64) Tag {
	return newTag("wf-reset-next-event-id", resetNextEventID, ValueTypeInteger)
}

// history tree
// WorkflowTreeID returns tag for WorkflowTreeID
func WorkflowTreeID(treeID string) Tag {
	return newTag("wf-tree-id", treeID, ValueTypeString)
}

// WorkflowBranchID returns tag for WorkflowBranchID
func WorkflowBranchID(branchID string) Tag {
	return newTag("wf-branch-id", branchID, ValueTypeString)
}

// workflow task
// WorkflowDecisionType returns tag for WorkflowDecisionType
func WorkflowDecisionType(decisionType int64) Tag {
	return newTag("wf-decision-type", decisionType, ValueTypeInteger)
}

// WorkflowDecisionFailCause returns tag for WorkflowDecisionFailCause
func WorkflowDecisionFailCause(decisionFailCause int64) Tag {
	return newTag("wf-decision-fail-cause", decisionFailCause, ValueTypeInteger)
}

// WorkflowTaskListType returns tag for WorkflowTaskListType
func WorkflowTaskListType(taskListType int64) Tag {
	return newTag("wf-task-list-type", taskListType, ValueTypeInteger)
}

// WorkflowTaskListName returns tag for WorkflowTaskListName
func WorkflowTaskListName(taskListName string) Tag {
	return newTag("wf-task-list-name", taskListName, ValueTypeString)
}

// size limit
// WorkflowSize returns tag for WorkflowSize
func WorkflowSize(workflowSize int64) Tag {
	return newTag("wf-size", workflowSize, ValueTypeInteger)
}

// WorkflowSignalCount returns tag for SignalCount
func WorkflowSignalCount(signalCount int64) Tag {
	return newTag("wf-signal-count", signalCount, ValueTypeInteger)
}

// WorkflowHistorySize returns tag for HistorySize
func WorkflowHistorySize(historySize int64) Tag {
	return newTag("wf-history-size", historySize, ValueTypeInteger)
}

// WorkflowHistorySizeBytes returns tag for HistorySizeBytes
func WorkflowHistorySizeBytes(historySizeBytes int64) Tag {
	return newTag("wf-history-size-bytes", historySizeBytes, ValueTypeInteger)
}

// WorkflowEventCount returns tag for EventCount
func WorkflowEventCount(eventCount int64) Tag {
	return newTag("wf-event-count", eventCount, ValueTypeInteger)
}

///////////////////  System tags defined here:  ///////////////////
// Tags with pre-define values
// Component returns tag for Component
func Component(component valueTypeSysComponent) Tag {
	return newTag("component", component, ValueTypeSysComponent)
}

// Lifecycle returns tag for Lifecycle
func Lifecycle(lifecycle valueTypeSysLifecycle) Tag {
	return newTag("lifecycle", lifecycle, ValueTypeSysLifecycle)
}

// StoreOperation returns tag for StoreOperation
func StoreOperation(storeOperation valueTypeSysStoreOperation) Tag {
	return newTag("store-operation", storeOperation, ValueTypeSysStoreOperation)
}

// OperationResult returns tag for OperationResult
func OperationResult(operationResult valueTypeSysOperationResult) Tag {
	return newTag("operation-result", operationResult, ValueTypeSysOperationResult)
}

// ErrorType returns tag for ErrorType
func ErrorType(errorType valueTypeSysErrorType) Tag {
	return newTag("error", errorType, ValueTypeSysErrorType)
}

// Shardupdate returns tag for Shardupdate
func Shardupdate(shardupdate valueTypeSysShardUpdate) Tag {
	return newTag("shard-update", shardupdate, ValueTypeSysShardUpdate)
}

// general
// CursorTimestamp returns tag for CursorTimestamp
func CursorTimestamp(timestamp time.Time) Tag {
	return newTag("cursor-timestamp", timestamp, ValueTypeTime)
}

// MetricScope returns tag for MetricScope
func MetricScope(metricScope int64) Tag {
	return newTag("metric-scope", metricScope, ValueTypeInteger)
}

// history engine shard
// ShardID returns tag for ShardID
func ShardID(shardID int64) Tag {
	return newTag("shard-id", shardID, ValueTypeInteger)
}

// ShardTime returns tag for ShardTime
func ShardTime(shardTime time.Time) Tag {
	return newTag("shard-time", shardTime, ValueTypeTime)
}

// ShardReplicationAck returns tag for ShardReplicationAck
func ShardReplicationAck(shardReplicationAck int64) Tag {
	return newTag("shard-replication-ack", shardReplicationAck, ValueTypeInteger)
}

// ShardTransferAcks returns tag for ShardTransferAcks
func ShardTransferAcks(shardTransferAcks int64) Tag {
	return newTag("shard-transfer-acks", shardTransferAcks, ValueTypeInteger)
}

// ShardTimerAcks returns tag for ShardTimerAcks
func ShardTimerAcks(shardTimerAcks int64) Tag {
	return newTag("shard-timer-acks", shardTimerAcks, ValueTypeInteger)
}

// task queue processor
// TaskID returns tag for TaskID
func TaskID(taskID int64) Tag {
	return newTag("queue-task-id", taskID, ValueTypeInteger)
}

// TaskType returns tag for TaskType for queue processor
func TaskType(taskType int64) Tag {
	return newTag("queue-task-type", taskType, ValueTypeInteger)
}

// TimerTaskStatus returns tag for TimerTaskStatus
func TimerTaskStatus(timerTaskStatus int64) Tag {
	return newTag("timer-task-status", timerTaskStatus, ValueTypeInteger)
}

// retry
// Attempt returns tag for Attempt
func Attempt(attempt int64) Tag {
	return newTag("attempt", attempt, ValueTypeInteger)
}

// AttemptCount returns tag for AttemptCount
func AttemptCount(attemptCount int64) Tag {
	return newTag("attempt-count", attemptCount, ValueTypeInteger)
}

// AttemptStart returns tag for AttemptStart
func AttemptStart(attemptStart time.Time) Tag {
	return newTag("attempt-start", attemptStart, ValueTypeTime)
}

// AttemptEnd returns tag for AttemptEnd
func AttemptEnd(attemptEnd time.Time) Tag {
	return newTag("attempt-end", attemptEnd, ValueTypeTime)
}

// ScheduleAttempt returns tag for ScheduleAttempt
func ScheduleAttempt(scheduleAttempt int64) Tag {
	return newTag("schedule-attempt", scheduleAttempt, ValueTypeInteger)
}

// ElastiSearch
// ESRequest returns tag for ESRequest
func ESRequest(ESRequest string) Tag {
	return newTag("es-request", ESRequest, ValueTypeString)
}

// ESKey returns tag for ESKey
func ESKey(ESKey string) Tag {
	return newTag("es-mapping-key", ESKey, ValueTypeString)
}

// ESField returns tag for ESField
func ESField(ESField string) Tag {
	return newTag("es-field", ESField, ValueTypeString)
}

// CallAt returns a tag for LoggingError
func LoggingCallAt(position string) Tag {
	return newTag("logging-call-at", position, ValueTypeString)
}

// LoggingError returns a tag for LoggingError
func LoggingError(err error) Tag {
	return newTag("longging-err", err, ValueTypeError)
}

// LoggingErrorFields returns a tag for LoggingErrorFields
func LoggingErrorFields(fields []string) Tag {
	return newTag("longging-err-fields", fields, ValueTypeObject)
}

// Kafka related
// KafkaTopicName returns tag for TopicName
func KafkaTopicName(topicName string) Tag {
	return newTag("kafka-topic-name", topicName, ValueTypeString)
}

// KafkaConsumerName returns tag for ConsumerName
func ConsumerName(consumerName string) Tag {
	return newTag("kafka-consumer-name", consumerName, ValueTypeString)
}

// KafkaPartition returns tag for Partition
func Partition(partition int64) Tag {
	return newTag("kafka-partition", partition, ValueTypeInteger)
}

// KafkaPartitionKey returns tag for PartitionKey
func PartitionKey(partitionKey interface{}) Tag {
	return newTag("kafka-partition-key", partitionKey, ValueTypeObject)
}

// KafkaOffset returns tag for Offset
func KafkaOffset(offset int64) Tag {
	return newTag("kafka-offset", offset, ValueTypeInteger)
}

///////////////////  XDC tags defined here: xdc- ///////////////////
// SourceCluster returns tag for SourceCluster
func SourceCluster(sourceCluster string) Tag {
	return newTag("source-cluster", sourceCluster, ValueTypeString)
}

// PrevActiveCluster returns tag for PrevActiveCluster
func PrevActiveCluster(prevActiveCluster string) Tag {
	return newTag("prev-active-cluster", prevActiveCluster, ValueTypeString)
}

// FailoverMsg returns tag for FailoverMsg
func FailoverMsg(failoverMsg string) Tag {
	return newTag("xdc-failover-msg", failoverMsg, ValueTypeString)
}

// FailoverVersion returns tag for Version
func FailoverVersion(version int64) Tag {
	return newTag("xdc-failover-version", version, ValueTypeInteger)
}

// CurrentVersion returns tag for CurrentVersion
func CurrentVersion(currentVersion int64) Tag {
	return newTag("xdc-current-version", currentVersion, ValueTypeInteger)
}

// IncomingVersion returns tag for IncomingVersion
func IncomingVersion(incomingVersion int64) Tag {
	return newTag("xdc-incoming-version", incomingVersion, ValueTypeInteger)
}

// ReplicationInfo returns tag for ReplicationInfo
func ReplicationInfo(replicationInfo *persistence.ReplicationInfo) Tag {
	return newTag("xdc-replication-info", replicationInfo, ValueTypeObject)
}

// ReplicationState returns tag for ReplicationState
func ReplicationState(replicationState *persistence.ReplicationState) Tag {
	return newTag("xdc-replication-state", replicationState, ValueTypeObject)
}

///////////////////  Archival tags defined here: archival- ///////////////////
// archival request tags
// ArchivalRequestDomainID returns tag for RequestDomainID
func ArchivalRequestDomainID(requestDomainID string) Tag {
	return newTag("archival-request-domain-id", requestDomainID, ValueTypeString)
}

// ArchivalRequestWorkflowID returns tag for RequestWorkflowID
func ArchivalRequestWorkflowID(requestWorkflowID string) Tag {
	return newTag("archival-request-workflow-id", requestWorkflowID, ValueTypeString)
}

// ArchivalRequestRunID returns tag for RequestRunID
func ArchivalRequestRunID(requestRunID string) Tag {
	return newTag("archival-request-run-id", requestRunID, ValueTypeString)
}

// ArchivalRequestEventStoreVersion returns tag for RequestEventStoreVersion
func ArchivalRequestEventStoreVersion(requestEventStoreVersion int64) Tag {
	return newTag("archival-request-event-store-version", requestEventStoreVersion, ValueTypeInteger)
}

// ArchivalRequestBranchToken returns tag for RequestBranchToken
func ArchivalRequestBranchToken(requestBranchToken string) Tag {
	return newTag("archival-request-branch-token", requestBranchToken, ValueTypeString)
}

// ArchivalRequestNextEventID returns tag for RequestNextEventID
func ArchivalRequestNextEventID(requestNextEventID int64) Tag {
	return newTag("archival-request-next-event-id", requestNextEventID, ValueTypeInteger)
}

// ArchivalRequestCloseFailoverVersion returns tag for RequestCloseFailoverVersion
func ArchivalRequestCloseFailoverVersion(requestCloseFailoverVersion int64) Tag {
	return newTag("archival-request-close-failover-version", requestCloseFailoverVersion, ValueTypeInteger)
}

// archival tags (blob tags)
// ArchivalBucket returns tag for Bucket
func ArchivalArchivalBucket(bucket string) Tag {
	return newTag("archival-bucket", bucket, ValueTypeString)
}

// ArchivalBlobKey returns tag for BlobKey
func ArchivalBlobKey(blobKey string) Tag {
	return newTag("archival-blob-key", blobKey, ValueTypeString)
}

// archival tags (file blobstore tags)
// ArchivalFileBlobstoreBlobPath returns tag for FileBlobstoreBlobPath
func ArchivalArchivalFileBlobstoreBlobPath(fileBlobstoreBlobPath string) Tag {
	return newTag("archival-file-blobstore-blob-path", fileBlobstoreBlobPath, ValueTypeString)
}

// ArchivalFileBlobstoreMetadataPath returns tag for FileBlobstoreMetadataPath
func ArchivalFileBlobstoreMetadataPath(fileBlobstoreMetadataPath string) Tag {
	return newTag("archival-file-blobstore-metadata-path", fileBlobstoreMetadataPath, ValueTypeString)
}

// ArchivalClusterArchivalStatus returns tag for ClusterArchivalStatus
func ArchivalClusterArchivalStatus(clusterArchivalStatus interface{}) Tag {
	return newTag("archival-cluster-archival-status", clusterArchivalStatus, ValueTypeObject)
}

// ArchivalUploadSkipReason returns tag for UploadSkipReason
func ArchivalUploadSkipReason(uploadSkipReason string) Tag {
	return newTag("archival-upload-skip-reason", uploadSkipReason, ValueTypeString)
}

// UploadFailReason returns tag for UploadFailReason
func UploadFailReason(uploadFailReason string) Tag {
	return newTag("archival-upload-fail-reason", uploadFailReason, ValueTypeString)
}
