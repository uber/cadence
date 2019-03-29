package log

// All logging tags are defined here.
// Tags are categorized and must be prefixed correctly
// We currently have those categories:
//   1. Workflow(wf-): these tags are information that are useful to our customer, like workflow-id/run-id/task-list/...
//   2. System (sys-): these tags are internal information which usually cannot be understood by our customers,
//      If possible, it is recommended to categorize further into sub-systems:
//          2.1 XDC subsystem: xdc-
//          2.2 Archival subsystem: archival-
//   3. There is a preserved prefix for internal logging: logging-

// Reserved prefix for logging system: logging-
var (
	TagLoggingError       = newTagType("logging-error", valueTypeError)
	TagLoggingErrorFields = newTagType("logging-error-fields", valueTypeString)
)

// Workflow related tags, prefix : wf-
var (
	// Tags with pre-define values
	TagWorkflowAction         = newTagType("wf-action", valueTypeWorkflowAction)
	TagWorkflowListFilterType = newTagType("wf-list-filter-type", valueTypeWorkflowListFilterType)

	// general
	TagWorkflowError              = newTagType("wf-error", valueTypeError)
	TagWorkflowTimeoutType        = newTagType("wf-timeout-type", valueTypeInteger)
	TagWorkflowPollContextTimeout = newTagType("wf-poll-context-timeout", valueTypeDuration)
	TagWorkflowHandlerName        = newTagType("wf-handler-name", valueTypeString)
	TagWorkflowID                 = newTagType("wf-id", valueTypeString)
	TagWorkflowType               = newTagType("wf-type", valueTypeString)
	TagWorkflowRunID              = newTagType("wf-run-id", valueTypeString)
	TagWorkflowBeginningRunID     = newTagType("wf-beginning-run-id", valueTypeString)
	TagWorkflowEndingRunID        = newTagType("wf-ending-run-id", valueTypeString)

	// domain related
	TagWorkflowDomainID   = newTagType("wf-domain-id", valueTypeString)
	TagWorkflowDomainName = newTagType("wf-domain-name", valueTypeString)
	TagWorkflowDomainIDs  = newTagType("wf-domain-ids", valueTypeObject)

	// history event ID related
	TagWorkflowEventID               = newTagType("wf-history-event-id", valueTypeInteger)
	TagWorkflowScheduleID            = newTagType("wf-schedule-id", valueTypeInteger)
	TagWorkflowFirstEventID          = newTagType("wf-first-event-id", valueTypeInteger)
	TagWorkflowNextEventID           = newTagType("wf-next-event-id", valueTypeInteger)
	TagWorkflowBeginningFirstEventID = newTagType("wf-begining-first-event-id", valueTypeInteger)
	TagWorkflowEndingNextEventID     = newTagType("wf-ending-next-event-id", valueTypeInteger)
	TagWorkflowResetNextEventID      = newTagType("wf-reset-next-event-id", valueTypeInteger)

	// history tree
	TagWorkflowTreeID   = "wf-tree-id"
	TagWorkflowBranchID = "wf-branch-id"

	// workflow task
	TagWorkflowDecisionType      = newTagType("wf-decision-type", valueTypeInteger)
	TagWorkflowDecisionFailCause = newTagType("wf-decision-fail-cause", valueTypeInteger)
	TagWorkflowTaskListType      = newTagType("wf-task-list-type", valueTypeInteger)
	TagWorkflowTaskListName      = newTagType("wf-task-list-name", valueTypeString)
)

// Generic system related tags, prefix: sys-
var (
	// Tags with pre-define values
	TagSysComponent       = newTagType("sys-component", valueTypeSysComponent)
	TagSysLifecycle       = newTagType("sys-lifecycle", valueTypeSysLifecycle)
	TagSysStoreOperation  = newTagType("sys-store-operation", valueTypeSysStoreOperation)
	TagSysOperationResult = newTagType("sys-error", valueTypeSysOperationResult)
	TagSysErrorType       = newTagType("sys-error", valueTypeSysErrorType)
	TagSysShardupdate     = newTagType("sys-error", valueTypeSysShardUpdate)

	// general
	TagSysError           = newTagType("sys-error", valueTypeError)
	TagSysTimestamp       = newTagType("sys-timestamp", valueTypeTime)
	TagSysCursorTimestamp = newTagType("sys-cursor-timestamp", valueTypeTime)
	TagSysMetricScope     = newTagType("sys-metric-scope", valueTypeInteger)

	// history engine shard
	TagSysShardID             = newTagType("sys-shard-id", valueTypeInteger)
	TagSysShardTime           = newTagType("sys-shard-time", valueTypeTime)
	TagSysShardReplicationAck = newTagType("sys-shard-replication-ack", valueTypeInteger)
	TagSysShardTransferAcks   = newTagType("sys-shard-transfer-acks", valueTypeInteger)
	TagSysShardTimerAcks      = newTagType("sys-shard-timer-acks", valueTypeInteger)

	// task queue processor
	TagSysTaskID          = newTagType("sys-queue-task-id", valueTypeInteger)
	TagSysTaskType        = newTagType("sys-queue-task-type", valueTypeInteger)
	TagSysTimerTaskStatus = newTagType("sys-timer-task-status", valueTypeInteger)

	// retry
	TagSysAttempt         = newTagType("sys-attempt", valueTypeInteger)
	TagSysAttemptCount    = newTagType("sys-attempt-count", valueTypeInteger)
	TagSysAttemptStart    = newTagType("sys-attempt-start", valueTypeTime)
	TagSysAttemptEnd      = newTagType("sys-attempt-end", valueTypeTime)
	TagSysScheduleAttempt = newTagType("sys-schedule-attempt", valueTypeInteger)

	// size limit
	TagSysWorkflowSize     = newTagType("sys-workflow-size", valueTypeInteger)
	TagSysSignalCount      = newTagType("sys-signal-count", valueTypeInteger)
	TagSysHistorySize      = newTagType("sys-history-size", valueTypeInteger)
	TagSysHistorySizeBytes = newTagType("sys-history-size-bytes", valueTypeInteger)
	TagSysEventCount       = newTagType("sys-event-count", valueTypeInteger)

	// ElastiSearch
	TagSysESRequest = newTagType("sys-es-request", valueTypeString)
	TagSysESKey     = newTagType("sys-es-mapping-key", valueTypeString)
	TagSysESField   = newTagType("sys-es-field", valueTypeString)
)

// XDC related tags: prefix: xdc-
var (
	TagXDCClusterName       = newTagType("xdc-cluster-name", valueTypeString)
	TagXDCSourceCluster     = newTagType("xdc-source-cluster", valueTypeString)
	TagXDCPrevActiveCluster = newTagType("xdc-prev-active-cluster", valueTypeString)

	// Kafka related
	TagXDCTopicName    = newTagType("xdc-topic-name", valueTypeString)
	TagXDCConsumerName = newTagType("xdc-consumer-name", valueTypeString)
	TagXDCPartition    = newTagType("xdc-partition", valueTypeInteger)
	TagXDCPartitionKey = newTagType("xdc-partition-key", valueTypeObject)
	TagXDCOffset       = newTagType("xdc-offset", valueTypeInteger)

	TagXDCFailoverMsg     = newTagType("xdc-failover-msg", valueTypeString)
	TagXDCVersion         = newTagType("xdc-version", valueTypeInteger)
	TagXDCCurrentVersion  = newTagType("xdc-current-version", valueTypeInteger)
	TagXDCIncomingVersion = newTagType("xdc-incoming-version", valueTypeInteger)

	TagXDCReplicationInfo  = newTagType("xdc-replication-info", valueTypeObject)
	TagXDCReplicationState = newTagType("xdc-replication-state", valueTypeObject)
)

// Archival subsystem: archival-
var (
	// archival request tags
	TagArchivalRequestDomainID             = newTagType("archival-request-domain-id", valueTypeString)
	TagArchivalRequestWorkflowID           = newTagType("archival-request-workflow-id", valueTypeString)
	TagArchivalRequestRunID                = newTagType("archival-request-run-id", valueTypeString)
	TagArchivalRequestEventStoreVersion    = newTagType("archival-request-event-store-version", valueTypeInteger)
	TagArchivalRequestBranchToken          = newTagType("archival-request-branch-token", valueTypeString)
	TagArchivalRequestNextEventID          = newTagType("archival-request-next-event-id", valueTypeInteger)
	TagArchivalRequestCloseFailoverVersion = newTagType("archival-request-close-failover-version", valueTypeInteger)

	// archival tags (blob tags)
	TagArchivalBucket  = newTagType("archival-bucket", valueTypeString)
	TagArchivalBlobKey = newTagType("archival-blob-key", valueTypeString)

	// archival tags (file blobstore tags)
	TagArchivalFileBlobstoreBlobPath     = newTagType("archival-file-blobstore-blob-path", valueTypeString)
	TagArchivalFileBlobstoreMetadataPath = newTagType("archival-file-blobstore-metadata-path", valueTypeString)

	// archival tags (other tags)
	TagArchivalClusterArchivalStatus = newTagType("archival-cluster-archival-status", valueTypeObject)
	TagArchivalUploadSkipReason      = newTagType("archival-upload-skip-reason", valueTypeString)
	TagArchivalUploadFailReason      = newTagType("archival-upload-fail-reason", valueTypeString)
)
