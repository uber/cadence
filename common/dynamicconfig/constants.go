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

package dynamicconfig

// Key represents a key/property stored in dynamic config
type Key int

func (k Key) String() string {
	keyName, ok := keys[k]
	if !ok {
		return keys[unknownKey]
	}
	return keyName
}

/***
* !!!Important!!!
* For developer: Make sure to add/maintain the comment in the right format: usage, keyName, and default value
* So that our go-docs can have the full [documentation](https://pkg.go.dev/github.com/uber/cadence@v0.19.1/common/service/dynamicconfig#Key).
***/
const (
	unknownKey Key = iota

	// key for tests
	testGetPropertyKey
	testGetIntPropertyKey
	testGetFloat64PropertyKey
	testGetDurationPropertyKey
	testGetBoolPropertyKey
	testGetStringPropertyKey
	testGetMapPropertyKey
	testGetIntPropertyFilteredByDomainKey
	testGetDurationPropertyFilteredByDomainKey
	testGetIntPropertyFilteredByTaskListInfoKey
	testGetDurationPropertyFilteredByTaskListInfoKey
	testGetBoolPropertyFilteredByDomainIDKey
	testGetBoolPropertyFilteredByTaskListInfoKey

	// EnableGlobalDomain is key for enable global domain
	// KeyName: system.enableGlobalDomain
	// Default value: false
	EnableGlobalDomain
	// EnableVisibilitySampling is key for enable visibility sampling
	// KeyName: system.enableVisibilitySampling
	// Default value: TRUE
	EnableVisibilitySampling
	// EnableReadFromClosedExecutionV2 is key for enable read from cadence_visibility.closed_executions_v2
	// KeyName: system.enableReadFromClosedExecutionV2
	// Default value: FALSE
	EnableReadFromClosedExecutionV2
	// AdvancedVisibilityWritingMode is key for how to write to advanced visibility
	// KeyName: system.advancedVisibilityWritingMode
	// Default value: based on whether or not advanced visibility persistence is configured (common.GetDefaultAdvancedVisibilityWritingMode(isAdvancedVisConfigExist))
	AdvancedVisibilityWritingMode
	// EmitShardDiffLog is whether emit the shard diff log
	// KeyName: history.emitShardDiffLog
	// Default value: FALSE
	EmitShardDiffLog
	// EnableReadVisibilityFromES is key for enable read from elastic search
	// KeyName: system.enableReadVisibilityFromES
	// Default value: based on whether or not advanced visibility persistence is configured(isAdvancedVisExistInConfig)
	EnableReadVisibilityFromES
	// DisableListVisibilityByFilter is config to disable list open/close workflow using filter
	// KeyName: frontend.disableListVisibilityByFilter
	// Default value: FALSE
	DisableListVisibilityByFilter
	// HistoryArchivalStatus is key for the status of history archival
	// KeyName: system.historyArchivalStatus
	// Default value: the value in static config: common.Config.Archival.History.Status
	HistoryArchivalStatus
	// EnableReadFromHistoryArchival is key for enabling reading history from archival store
	// KeyName: system.enableReadFromHistoryArchival
	// Default value: the value in static config: common.Config.Archival.History.EnableRead
	EnableReadFromHistoryArchival
	// VisibilityArchivalStatus is key for the status of visibility archival
	// KeyName: system.visibilityArchivalStatus
	// Default value: the value in static config: common.Config.Archival.Visibility.Status
	VisibilityArchivalStatus
	// EnableReadFromVisibilityArchival is key for enabling reading visibility from archival store
	// KeyName: system.enableReadFromVisibilityArchival
	// Default value: the value in static config: common.Config.Archival.Visibility.EnableRead
	EnableReadFromVisibilityArchival
	// EnableDomainNotActiveAutoForwarding is whether enabling DC auto forwarding to active cluster for signal / start / signal with start API if domain is not active
	// KeyName: system.enableDomainNotActiveAutoForwarding
	// Default value: TRUE
	EnableDomainNotActiveAutoForwarding
	// EnableGracefulFailover is whether enabling graceful failover
	// KeyName: system.enableGracefulFailover
	// Default value: FALSE
	EnableGracefulFailover
	// TransactionSizeLimit is the largest allowed transaction size to persistence
	// KeyName: system.transactionSizeLimit
	// Default value:  14 * 1024 * 1024 (common.DefaultTransactionSizeLimit)
	TransactionSizeLimit
	// PersistenceErrorInjectionRate is rate for injecting random error in persistence
	// KeyName: system.persistenceErrorInjectionRate
	// Default value: 0
	PersistenceErrorInjectionRate
	// MaxRetentionDays is the maximum allowed retention days for domain
	// KeyName: system.maxRetentionDays
	// Default value:  30(domain.DefaultMaxWorkflowRetentionInDays)
	MaxRetentionDays
	// MinRetentionDays is the minimal allowed retention days for domain
	// KeyName: system.minRetentionDays
	// Default value: domain.MinRetentionDays
	MinRetentionDays
	// MaxDecisionStartToCloseSeconds is the minimal allowed decision start to close timeout in seconds
	// KeyName: system.maxDecisionStartToCloseSeconds
	// Default value: 240
	MaxDecisionStartToCloseSeconds
	// DisallowQuery is the key to disallow query for a domain
	// KeyName: system.disallowQuery
	// Default value: FALSE
	DisallowQuery
	// EnableDebugMode is for enabling debugging components, logs and metrics
	// KeyName: system.enableDebugMode
	// Default value: false
	EnableDebugMode
	// RequiredDomainDataKeys is the key for the list of data keys required in domain registeration
	// KeyName: system.requiredDomainDataKeys
	// Default value: nil
	RequiredDomainDataKeys
	// EnableGRPCOutbound is the key for enabling outbound GRPC traffic
	// KeyName: system.enableGRPCOutbound
	// Default value: false
	EnableGRPCOutbound

	// BlobSizeLimitError is the per event blob size limit
	// KeyName: limit.blobSize.error
	// Default value: 2*1024*1024
	BlobSizeLimitError
	// BlobSizeLimitWarn is the per event blob size limit for warning
	// KeyName: limit.blobSize.warn
	// Default value: 256*1024
	BlobSizeLimitWarn
	// HistorySizeLimitError is the per workflow execution history size limit
	// KeyName: limit.historySize.error
	// Default value: 200*1024*1024
	HistorySizeLimitError
	// HistorySizeLimitWarn is the per workflow execution history size limit for warning
	// KeyName: limit.historySize.warn
	// Default value: 50*1024*1024
	HistorySizeLimitWarn
	// HistoryCountLimitError is the per workflow execution history event count limit
	// KeyName: limit.historyCount.error
	// Default value: 200*1024
	HistoryCountLimitError
	// HistoryCountLimitWarn is the per workflow execution history event count limit for warning
	// KeyName: limit.historyCount.warn
	// Default value: 50*1024
	HistoryCountLimitWarn
	// MaxIDLengthLimit is the length limit for various IDs, including: Domain, TaskList, WorkflowID, ActivityID, TimerID,WorkflowType, ActivityType, SignalName, MarkerName, ErrorReason/FailureReason/CancelCause, Identity, RequestID
	// KeyName: limit.maxIDLength
	// Default value: 1000
	MaxIDLengthLimit
	// MaxIDLengthWarnLimit is the warn length limit for various IDs, including: Domain, TaskList, WorkflowID, ActivityID, TimerID, WorkflowType, ActivityType, SignalName, MarkerName, ErrorReason/FailureReason/CancelCause, Identity, RequestID
	// KeyName: limit.maxIDWarnLength
	// Default value: 150
	MaxIDLengthWarnLimit
	// MaxRawTaskListNameLimit is max length of user provided task list name (non-sticky and non-scalable)
	// KeyName: limit.maxRawTaskListNameLength
	// Default value: 1000
	MaxRawTaskListNameLimit

	// key for admin

	// AdminErrorInjectionRate is the rate for injecting random error in admin client
	// KeyName: admin.errorInjectionRate
	// Default value: 0
	AdminErrorInjectionRate

	// key for frontend

	// FrontendPersistenceMaxQPS is the max qps frontend host can query DB
	// KeyName: frontend.persistenceMaxQPS
	// Default value: 2000
	FrontendPersistenceMaxQPS
	// FrontendPersistenceGlobalMaxQPS is the max qps frontend cluster can query DB
	// KeyName: frontend.persistenceGlobalMaxQPS
	// Default value: 0
	FrontendPersistenceGlobalMaxQPS
	// FrontendVisibilityMaxPageSize is default max size for ListWorkflowExecutions in one page
	// KeyName: frontend.visibilityMaxPageSize
	// Default value: 1000
	FrontendVisibilityMaxPageSize
	// FrontendVisibilityListMaxQPS is max qps frontend can list open/close workflows
	// KeyName: frontend.visibilityListMaxQPS
	// Default value: 1
	FrontendVisibilityListMaxQPS
	// FrontendESVisibilityListMaxQPS is max qps frontend can list open/close workflows from ElasticSearch
	// KeyName: frontend.esVisibilityListMaxQPS
	// Default value: 3
	FrontendESVisibilityListMaxQPS
	// FrontendESIndexMaxResultWindow is ElasticSearch index setting max_result_window
	// KeyName: frontend.esIndexMaxResultWindow
	// Default value: 10000
	FrontendESIndexMaxResultWindow
	// FrontendHistoryMaxPageSize is default max size for GetWorkflowExecutionHistory in one page
	// KeyName: frontend.historyMaxPageSize
	// Default value: common.GetHistoryMaxPageSize
	FrontendHistoryMaxPageSize
	// FrontendRPS is workflow rate limit per second
	// KeyName: frontend.rps
	// Default value: 1200
	FrontendRPS
	// FrontendMaxDomainRPSPerInstance is workflow domain rate limit per second
	// KeyName: frontend.domainrps
	// Default value: 1200
	FrontendMaxDomainRPSPerInstance
	// FrontendGlobalDomainRPS is workflow domain rate limit per second for the whole Cadence cluster
	// KeyName: frontend.globalDomainrps
	// Default value: 0
	FrontendGlobalDomainRPS
	// FrontendHistoryMgrNumConns is for persistence cluster.NumConns
	// KeyName: frontend.historyMgrNumConns
	// Default value: 10
	FrontendHistoryMgrNumConns
	// FrontendThrottledLogRPS is the rate limit on number of log messages emitted per second for throttled logger
	// KeyName: frontend.throttledLogRPS
	// Default value: 20
	FrontendThrottledLogRPS
	// FrontendShutdownDrainDuration is the duration of traffic drain during shutdown
	// KeyName: frontend.shutdownDrainDuration
	// Default value: 0
	FrontendShutdownDrainDuration
	// EnableClientVersionCheck is enables client version check for frontend
	// KeyName: frontend.enableClientVersionCheck
	// Default value: FALSE
	EnableClientVersionCheck
	// FrontendMaxBadBinaries is the max number of bad binaries in domain config
	// KeyName: frontend.maxBadBinaries
	// Default value: domain.MaxBadBinaries
	FrontendMaxBadBinaries
	// FrontendFailoverCoolDown is duration between two domain failvoers
	// KeyName: frontend.failoverCoolDown
	// Default value: time.Minute
	FrontendFailoverCoolDown
	// ValidSearchAttributes is legal indexed keys that can be used in list APIs
	// KeyName: frontend.validSearchAttributes
	// Default value: definition.GetDefaultIndexedKeys()
	ValidSearchAttributes
	// SendRawWorkflowHistory is whether to enable raw history retrieving
	// KeyName: frontend.sendRawWorkflowHistory
	// Default value: sendRawWorkflowHistory
	SendRawWorkflowHistory
	// SearchAttributesNumberOfKeysLimit is the limit of number of keys
	// KeyName: frontend.searchAttributesNumberOfKeysLimit
	// Default value: 100
	SearchAttributesNumberOfKeysLimit
	// SearchAttributesSizeOfValueLimit is the size limit of each value
	// KeyName: frontend.searchAttributesSizeOfValueLimit
	// Default value: 2*1024
	SearchAttributesSizeOfValueLimit
	// SearchAttributesTotalSizeLimit is the size limit of the whole map
	// KeyName: frontend.searchAttributesTotalSizeLimit
	// Default value: 40*1024
	SearchAttributesTotalSizeLimit
	// VisibilityArchivalQueryMaxPageSize is the maximum page size for a visibility archival query
	// KeyName: frontend.visibilityArchivalQueryMaxPageSize
	// Default value: 10000
	VisibilityArchivalQueryMaxPageSize
	// DomainFailoverRefreshInterval is the domain failover refresh timer
	// KeyName: frontend.domainFailoverRefreshInterval
	// Default value: 10*time.Second
	DomainFailoverRefreshInterval
	// DomainFailoverRefreshTimerJitterCoefficient is the jitter for domain failover refresh timer jitter
	// KeyName: frontend.domainFailoverRefreshTimerJitterCoefficient
	// Default value: 0.1
	DomainFailoverRefreshTimerJitterCoefficient
	// FrontendErrorInjectionRate is rate for injecting random error in frontend client
	// KeyName: frontend.errorInjectionRate
	// Default value: 0
	FrontendErrorInjectionRate

	// key for matching

	// MatchingRPS is request rate per second for each matching host
	// KeyName: matching.rps
	// Default value: 1200
	MatchingRPS
	// MatchingPersistenceMaxQPS is the max qps matching host can query DB
	// KeyName: matching.persistenceMaxQPS
	// Default value: 3000
	MatchingPersistenceMaxQPS
	// MatchingPersistenceGlobalMaxQPS is the max qps matching cluster can query DB
	// KeyName: matching.persistenceGlobalMaxQPS
	// Default value: 0
	MatchingPersistenceGlobalMaxQPS
	// MatchingMinTaskThrottlingBurstSize is the minimum burst size for task list throttling
	// KeyName: matching.minTaskThrottlingBurstSize
	// Default value: 1
	MatchingMinTaskThrottlingBurstSize
	// MatchingGetTasksBatchSize is the maximum batch size to fetch from the task buffer
	// KeyName: matching.getTasksBatchSize
	// Default value: 1000
	MatchingGetTasksBatchSize
	// MatchingLongPollExpirationInterval is the long poll expiration interval in the matching service
	// KeyName: matching.longPollExpirationInterval
	// Default value: time.Minute
	MatchingLongPollExpirationInterval
	// MatchingEnableSyncMatch is to enable sync match
	// KeyName: matching.enableSyncMatch
	// Default value: TRUE
	MatchingEnableSyncMatch
	// MatchingUpdateAckInterval is the interval for update ack
	// KeyName: matching.updateAckInterval
	// Default value: 1*time.Minute
	MatchingUpdateAckInterval
	// MatchingIdleTasklistCheckInterval is the IdleTasklistCheckInterval
	// KeyName: matching.idleTasklistCheckInterval
	// Default value: 5*time.Minute
	MatchingIdleTasklistCheckInterval
	// MaxTasklistIdleTime is the max time tasklist being idle
	// KeyName: matching.maxTasklistIdleTime
	// Default value: 5*time.Minute
	MaxTasklistIdleTime
	// MatchingOutstandingTaskAppendsThreshold is the threshold for outstanding task appends
	// KeyName: matching.outstandingTaskAppendsThreshold
	// Default value: 250
	MatchingOutstandingTaskAppendsThreshold
	// MatchingMaxTaskBatchSize is max batch size for task writer
	// KeyName: matching.maxTaskBatchSize
	// Default value: 100
	MatchingMaxTaskBatchSize
	// MatchingMaxTaskDeleteBatchSize is the max batch size for range deletion of tasks
	// KeyName: matching.maxTaskDeleteBatchSize
	// Default value: 100
	MatchingMaxTaskDeleteBatchSize
	// MatchingThrottledLogRPS is the rate limit on number of log messages emitted per second for throttled logger
	// KeyName: matching.throttledLogRPS
	// Default value: 20
	MatchingThrottledLogRPS
	// MatchingNumTasklistWritePartitions is the number of write partitions for a task list
	// KeyName: matching.numTasklistWritePartitions
	// Default value: 1
	MatchingNumTasklistWritePartitions
	// MatchingNumTasklistReadPartitions is the number of read partitions for a task list
	// KeyName: matching.numTasklistReadPartitions
	// Default value: 1
	MatchingNumTasklistReadPartitions
	// MatchingForwarderMaxOutstandingPolls is the max number of inflight polls from the forwarder
	// KeyName: matching.forwarderMaxOutstandingPolls
	// Default value: 1
	MatchingForwarderMaxOutstandingPolls
	// MatchingForwarderMaxOutstandingTasks is the max number of inflight addTask/queryTask from the forwarder
	// KeyName: matching.forwarderMaxOutstandingTasks
	// Default value: 1
	MatchingForwarderMaxOutstandingTasks
	// MatchingForwarderMaxRatePerSecond is the max rate at which add/query can be forwarded
	// KeyName: matching.forwarderMaxRatePerSecond
	// Default value: 10
	MatchingForwarderMaxRatePerSecond
	// MatchingForwarderMaxChildrenPerNode is the max number of children per node in the task list partition tree
	// KeyName: matching.forwarderMaxChildrenPerNode
	// Default value: 20
	MatchingForwarderMaxChildrenPerNode
	// MatchingShutdownDrainDuration is the duration of traffic drain during shutdown
	// KeyName: matching.shutdownDrainDuration
	// Default value: 0
	MatchingShutdownDrainDuration
	// MatchingErrorInjectionRate is rate for injecting random error in matching client
	// KeyName: matching.errorInjectionRate
	// Default value: 0
	MatchingErrorInjectionRate
	// MatchingEnableTaskInfoLogByDomainID is enables info level logs for decision/activity task based on the request domainID
	// KeyName: matching.enableTaskInfoLogByDomainID
	// Default value: false
	MatchingEnableTaskInfoLogByDomainID

	// key for history

	// HistoryRPS is request rate per second for each history host
	// KeyName: history.rps
	// Default value: 3000
	HistoryRPS
	// HistoryPersistenceMaxQPS is the max qps history host can query DB
	// KeyName: history.persistenceMaxQPS
	// Default value: 9000
	HistoryPersistenceMaxQPS
	// HistoryPersistenceGlobalMaxQPS is the max qps history cluster can query DB
	// KeyName: history.persistenceGlobalMaxQPS
	// Default value: 0
	HistoryPersistenceGlobalMaxQPS
	// HistoryVisibilityOpenMaxQPS is max qps one history host can write visibility open_executions
	// KeyName: history.historyVisibilityOpenMaxQPS
	// Default value: 300
	HistoryVisibilityOpenMaxQPS
	// HistoryVisibilityClosedMaxQPS is max qps one history host can write visibility closed_executions
	// KeyName: history.historyVisibilityClosedMaxQPS
	// Default value: 300
	HistoryVisibilityClosedMaxQPS
	// HistoryLongPollExpirationInterval is the long poll expiration interval in the history service
	// KeyName: history.longPollExpirationInterval
	// Default value: time.Second*20
	HistoryLongPollExpirationInterval
	// HistoryCacheInitialSize is initial size of history cache
	// KeyName: history.cacheInitialSize
	// Default value: 128
	HistoryCacheInitialSize
	// HistoryCacheMaxSize is max size of history cache
	// KeyName: history.cacheMaxSize
	// Default value: 512
	HistoryCacheMaxSize
	// HistoryCacheTTL is TTL of history cache
	// KeyName: history.cacheTTL
	// Default value: time.Hour
	HistoryCacheTTL
	// HistoryShutdownDrainDuration is the duration of traffic drain during shutdown
	// KeyName: history.shutdownDrainDuration
	// Default value: 0
	HistoryShutdownDrainDuration
	// EventsCacheInitialCount is initial count of events cache
	// KeyName: history.eventsCacheInitialSize
	// Default value: 128
	EventsCacheInitialCount
	// EventsCacheMaxCount is max count of events cache
	// KeyName: history.eventsCacheMaxSize
	// Default value: 512
	EventsCacheMaxCount
	// EventsCacheMaxSize is max size of events cache in bytes
	// KeyName: history.eventsCacheMaxSizeInBytes
	// Default value: 0
	EventsCacheMaxSize
	// EventsCacheTTL is TTL of events cache
	// KeyName: history.eventsCacheTTL
	// Default value: time.Hour
	EventsCacheTTL
	// EventsCacheGlobalEnable is enables global cache over all history shards
	// KeyName: history.eventsCacheGlobalEnable
	// Default value: FALSE
	EventsCacheGlobalEnable
	// EventsCacheGlobalInitialCount is initial count of global events cache
	// KeyName: history.eventsCacheGlobalInitialSize
	// Default value: 4096
	EventsCacheGlobalInitialCount
	// EventsCacheGlobalMaxCount is max count of global events cache
	// KeyName: history.eventsCacheGlobalMaxSize
	// Default value: 131072
	EventsCacheGlobalMaxCount
	// AcquireShardInterval is interval that timer used to acquire shard
	// KeyName: history.acquireShardInterval
	// Default value: time.Minute
	AcquireShardInterval
	// AcquireShardConcurrency is number of goroutines that can be used to acquire shards in the shard controller.
	// KeyName: history.acquireShardConcurrency
	// Default value: 1
	AcquireShardConcurrency
	// StandbyClusterDelay is the artificial delay added to standby cluster's view of active cluster's time
	// KeyName: history.standbyClusterDelay
	// Default value: 5*time.Minute
	StandbyClusterDelay
	// StandbyTaskMissingEventsResendDelay is the amount of time standby cluster's will wait (if events are missing)before calling remote for missing events
	// KeyName: history.standbyTaskMissingEventsResendDelay
	// Default value: 15*time.Minute
	StandbyTaskMissingEventsResendDelay
	// StandbyTaskMissingEventsDiscardDelay is the amount of time standby cluster's will wait (if events are missing)before discarding the task
	// KeyName: history.standbyTaskMissingEventsDiscardDelay
	// Default value: 25*time.Minute
	StandbyTaskMissingEventsDiscardDelay
	// TaskProcessRPS is the task processing rate per second for each domain
	// KeyName: history.taskProcessRPS
	// Default value: 1000
	TaskProcessRPS
	// TaskSchedulerType is the task scheduler type for priority task processor
	// KeyName: history.taskSchedulerType
	// Default value: int(task.SchedulerTypeWRR)
	TaskSchedulerType
	// TaskSchedulerWorkerCount is the number of workers per host in task scheduler
	// KeyName: history.taskSchedulerWorkerCount
	// Default value: 200
	TaskSchedulerWorkerCount
	// TaskSchedulerShardWorkerCount is the number of worker per shard in task scheduler
	// KeyName: history.taskSchedulerShardWorkerCount
	// Default value: 0
	TaskSchedulerShardWorkerCount
	// TaskSchedulerQueueSize is the size of task channel for host level task scheduler
	// KeyName: history.taskSchedulerQueueSize
	// Default value: 10000
	TaskSchedulerQueueSize
	// TaskSchedulerShardQueueSize is the size of task channel for shard level task scheduler
	// KeyName: history.taskSchedulerShardQueueSize
	// Default value: 200
	TaskSchedulerShardQueueSize
	// TaskSchedulerDispatcherCount is the number of task dispatcher in task scheduler (only applies to host level task scheduler)
	// KeyName: history.taskSchedulerDispatcherCount
	// Default value: 1
	TaskSchedulerDispatcherCount
	// TaskSchedulerRoundRobinWeights is the priority weight for weighted round robin task scheduler
	// KeyName: history.taskSchedulerRoundRobinWeight
	// Default value: common.ConvertIntMapToDynamicConfigMapProperty(DefaultTaskPriorityWeight)
	TaskSchedulerRoundRobinWeights
	// ActiveTaskRedispatchInterval is the active task redispatch interval
	// KeyName: history.activeTaskRedispatchInterval
	// Default value: 5*time.Second
	ActiveTaskRedispatchInterval
	// StandbyTaskRedispatchInterval is the standby task redispatch interval
	// KeyName: history.standbyTaskRedispatchInterval
	// Default value: 30*time.Second
	StandbyTaskRedispatchInterval
	// TaskRedispatchIntervalJitterCoefficient is the task redispatch interval jitter coefficient
	// KeyName: history.taskRedispatchIntervalJitterCoefficient
	// Default value: 0.15
	TaskRedispatchIntervalJitterCoefficient
	// StandbyTaskReReplicationContextTimeout is the context timeout for standby task re-replication
	// KeyName: history.standbyTaskReReplicationContextTimeout
	// Default value: 3*time.Minute
	StandbyTaskReReplicationContextTimeout
	// QueueProcessorEnableSplit is indicates whether processing queue split policy should be enabled
	// KeyName: history.queueProcessorEnableSplit
	// Default value: FALSE
	QueueProcessorEnableSplit
	// QueueProcessorSplitMaxLevel is the max processing queue level
	// KeyName: history.queueProcessorSplitMaxLevel
	// Default value: 2 // 3 levels, start from 0
	QueueProcessorSplitMaxLevel
	// QueueProcessorEnableRandomSplitByDomainID is indicates whether random queue split policy should be enabled for a domain
	// KeyName: history.queueProcessorEnableRandomSplitByDomainID
	// Default value: FALSE
	QueueProcessorEnableRandomSplitByDomainID
	// QueueProcessorRandomSplitProbability is the probability for a domain to be split to a new processing queue
	// KeyName: history.queueProcessorRandomSplitProbability
	// Default value: 0.01
	QueueProcessorRandomSplitProbability
	// QueueProcessorEnablePendingTaskSplitByDomainID is indicates whether pending task split policy should be enabled
	// KeyName: history.queueProcessorEnablePendingTaskSplitByDomainID
	// Default value: FALSE
	QueueProcessorEnablePendingTaskSplitByDomainID
	// QueueProcessorPendingTaskSplitThreshold is the threshold for the number of pending tasks per domain
	// KeyName: history.queueProcessorPendingTaskSplitThreshold
	// Default value: common.ConvertIntMapToDynamicConfigMapProperty(DefaultPendingTaskSplitThreshold)
	QueueProcessorPendingTaskSplitThreshold
	// QueueProcessorEnableStuckTaskSplitByDomainID is indicates whether stuck task split policy should be enabled
	// KeyName: history.queueProcessorEnableStuckTaskSplitByDomainID
	// Default value: FALSE
	QueueProcessorEnableStuckTaskSplitByDomainID
	// QueueProcessorStuckTaskSplitThreshold is the threshold for the number of attempts of a task
	// KeyName: history.queueProcessorStuckTaskSplitThreshold
	// Default value: common.ConvertIntMapToDynamicConfigMapProperty(DefaultStuckTaskSplitThreshold)
	QueueProcessorStuckTaskSplitThreshold
	// QueueProcessorSplitLookAheadDurationByDomainID is the look ahead duration when spliting a domain to a new processing queue
	// KeyName: history.queueProcessorSplitLookAheadDurationByDomainID
	// Default value: 20*time.Minute
	QueueProcessorSplitLookAheadDurationByDomainID
	// QueueProcessorPollBackoffInterval is the backoff duration when queue processor is throttled
	// KeyName: history.queueProcessorPollBackoffInterval
	// Default value: 5*time.Second
	QueueProcessorPollBackoffInterval
	// QueueProcessorPollBackoffIntervalJitterCoefficient is backoff interval jitter coefficient
	// KeyName: history.queueProcessorPollBackoffIntervalJitterCoefficient
	// Default value: 0.15
	QueueProcessorPollBackoffIntervalJitterCoefficient
	// QueueProcessorEnablePersistQueueStates is indicates whether processing queue states should be persisted
	// KeyName: history.queueProcessorEnablePersistQueueStates
	// Default value: FALSE
	QueueProcessorEnablePersistQueueStates
	// QueueProcessorEnableLoadQueueStates is indicates whether processing queue states should be loaded
	// KeyName: history.queueProcessorEnableLoadQueueStates
	// Default value: FALSE
	QueueProcessorEnableLoadQueueStates
	// TimerTaskBatchSize is batch size for timer processor to process tasks
	// KeyName: history.timerTaskBatchSize
	// Default value: 100
	TimerTaskBatchSize
	// TimerTaskWorkerCount is number of task workers for timer processor
	// KeyName: history.timerTaskWorkerCount
	// Default value: 10
	TimerTaskWorkerCount
	// TimerTaskMaxRetryCount is max retry count for timer processor
	// KeyName: history.timerTaskMaxRetryCount
	// Default value: 100
	TimerTaskMaxRetryCount
	// TimerProcessorGetFailureRetryCount is retry count for timer processor get failure operation
	// KeyName: history.timerProcessorGetFailureRetryCount
	// Default value: 5
	TimerProcessorGetFailureRetryCount
	// TimerProcessorCompleteTimerFailureRetryCount is retry count for timer processor complete timer operation
	// KeyName: history.timerProcessorCompleteTimerFailureRetryCount
	// Default value: 10
	TimerProcessorCompleteTimerFailureRetryCount
	// TimerProcessorUpdateAckInterval is update interval for timer processor
	// KeyName: history.timerProcessorUpdateAckInterval
	// Default value: 30*time.Second
	TimerProcessorUpdateAckInterval
	// TimerProcessorUpdateAckIntervalJitterCoefficient is the update interval jitter coefficient
	// KeyName: history.timerProcessorUpdateAckIntervalJitterCoefficient
	// Default value: 0.15
	TimerProcessorUpdateAckIntervalJitterCoefficient
	// TimerProcessorCompleteTimerInterval is complete timer interval for timer processor
	// KeyName: history.timerProcessorCompleteTimerInterval
	// Default value: 60*time.Second
	TimerProcessorCompleteTimerInterval
	// TimerProcessorFailoverMaxPollRPS is max poll rate per second for timer processor
	// KeyName: history.timerProcessorFailoverMaxPollRPS
	// Default value: 1
	TimerProcessorFailoverMaxPollRPS
	// TimerProcessorMaxPollRPS is max poll rate per second for timer processor
	// KeyName: history.timerProcessorMaxPollRPS
	// Default value: 20
	TimerProcessorMaxPollRPS
	// TimerProcessorMaxPollInterval is max poll interval for timer processor
	// KeyName: history.timerProcessorMaxPollInterval
	// Default value: 5*time.Minute
	TimerProcessorMaxPollInterval
	// TimerProcessorMaxPollIntervalJitterCoefficient is the max poll interval jitter coefficient
	// KeyName: history.timerProcessorMaxPollIntervalJitterCoefficient
	// Default value: 0.15
	TimerProcessorMaxPollIntervalJitterCoefficient
	// TimerProcessorSplitQueueInterval is the split processing queue interval for timer processor
	// KeyName: history.timerProcessorSplitQueueInterval
	// Default value: 1*time.Minute
	TimerProcessorSplitQueueInterval
	// TimerProcessorSplitQueueIntervalJitterCoefficient is the split processing queue interval jitter coefficient
	// KeyName: history.timerProcessorSplitQueueIntervalJitterCoefficient
	// Default value: 0.15
	TimerProcessorSplitQueueIntervalJitterCoefficient
	// TimerProcessorMaxRedispatchQueueSize is the threshold of the number of tasks in the redispatch queue for timer processor
	// KeyName: history.timerProcessorMaxRedispatchQueueSize
	// Default value: 10000
	TimerProcessorMaxRedispatchQueueSize
	// TimerProcessorMaxTimeShift is the max shift timer processor can have
	// KeyName: history.timerProcessorMaxTimeShift
	// Default value: 1*time.Second
	TimerProcessorMaxTimeShift
	// TimerProcessorHistoryArchivalSizeLimit is the max history size for inline archival
	// KeyName: history.timerProcessorHistoryArchivalSizeLimit
	// Default value: 500*1024
	TimerProcessorHistoryArchivalSizeLimit
	// TimerProcessorArchivalTimeLimit is the upper time limit for inline history archival
	// KeyName: history.timerProcessorArchivalTimeLimit
	// Default value: 1*time.Second
	TimerProcessorArchivalTimeLimit
	// TransferTaskBatchSize is batch size for transferQueueProcessor
	// KeyName: history.transferTaskBatchSize
	// Default value: 100
	TransferTaskBatchSize
	// TransferProcessorFailoverMaxPollRPS is max poll rate per second for transferQueueProcessor
	// KeyName: history.transferProcessorFailoverMaxPollRPS
	// Default value: 1
	TransferProcessorFailoverMaxPollRPS
	// TransferProcessorMaxPollRPS is max poll rate per second for transferQueueProcessor
	// KeyName: history.transferProcessorMaxPollRPS
	// Default value: 20
	TransferProcessorMaxPollRPS
	// TransferTaskWorkerCount is number of worker for transferQueueProcessor
	// KeyName: history.transferTaskWorkerCount
	// Default value: 10
	TransferTaskWorkerCount
	// TransferTaskMaxRetryCount is max times of retry for transferQueueProcessor
	// KeyName: history.transferTaskMaxRetryCount
	// Default value: 100
	TransferTaskMaxRetryCount
	// TransferProcessorCompleteTransferFailureRetryCount is times of retry for failure
	// KeyName: history.transferProcessorCompleteTransferFailureRetryCount
	// Default value: 10
	TransferProcessorCompleteTransferFailureRetryCount
	// TransferProcessorMaxPollInterval is max poll interval for transferQueueProcessor
	// KeyName: history.transferProcessorMaxPollInterval
	// Default value: 1*time.Minute
	TransferProcessorMaxPollInterval
	// TransferProcessorMaxPollIntervalJitterCoefficient is the max poll interval jitter coefficient
	// KeyName: history.transferProcessorMaxPollIntervalJitterCoefficient
	// Default value: 0.15
	TransferProcessorMaxPollIntervalJitterCoefficient
	// TransferProcessorSplitQueueInterval is the split processing queue interval for transferQueueProcessor
	// KeyName: history.transferProcessorSplitQueueInterval
	// Default value: 1*time.Minute
	TransferProcessorSplitQueueInterval
	// TransferProcessorSplitQueueIntervalJitterCoefficient is the split processing queue interval jitter coefficient
	// KeyName: history.transferProcessorSplitQueueIntervalJitterCoefficient
	// Default value: 0.15
	TransferProcessorSplitQueueIntervalJitterCoefficient
	// TransferProcessorUpdateAckInterval is update interval for transferQueueProcessor
	// KeyName: history.transferProcessorUpdateAckInterval
	// Default value: 30*time.Second
	TransferProcessorUpdateAckInterval
	// TransferProcessorUpdateAckIntervalJitterCoefficient is the update interval jitter coefficient
	// KeyName: history.transferProcessorUpdateAckIntervalJitterCoefficient
	// Default value: 0.15
	TransferProcessorUpdateAckIntervalJitterCoefficient
	// TransferProcessorCompleteTransferInterval is complete timer interval for transferQueueProcessor
	// KeyName: history.transferProcessorCompleteTransferInterval
	// Default value: 60*time.Second
	TransferProcessorCompleteTransferInterval
	// TransferProcessorMaxRedispatchQueueSize is the threshold of the number of tasks in the redispatch queue for transferQueueProcessor
	// KeyName: history.transferProcessorMaxRedispatchQueueSize
	// Default value: 10000
	TransferProcessorMaxRedispatchQueueSize
	// TransferProcessorEnableValidator is whether validator should be enabled for transferQueueProcessor
	// KeyName: history.transferProcessorEnableValidator
	// Default value: FALSE
	TransferProcessorEnableValidator
	// TransferProcessorValidationInterval is interval for performing transfer queue validation
	// KeyName: history.transferProcessorValidationInterval
	// Default value: 30*time.Second
	TransferProcessorValidationInterval
	// TransferProcessorVisibilityArchivalTimeLimit is the upper time limit for archiving visibility records
	// KeyName: history.transferProcessorVisibilityArchivalTimeLimit
	// Default value: 200*time.Millisecond
	TransferProcessorVisibilityArchivalTimeLimit
	// ReplicatorTaskBatchSize is batch size for ReplicatorProcessor
	// KeyName: history.replicatorTaskBatchSize
	// Default value: 100
	ReplicatorTaskBatchSize
	// ReplicatorTaskWorkerCount is number of worker for ReplicatorProcessor
	// KeyName: history.replicatorTaskWorkerCount
	// Default value: 10
	ReplicatorTaskWorkerCount
	// ReplicatorReadTaskMaxRetryCount is the number of read replication task retry time
	// KeyName: history.replicatorReadTaskMaxRetryCount
	// Default value: 3
	ReplicatorReadTaskMaxRetryCount
	// ReplicatorTaskMaxRetryCount is max times of retry for ReplicatorProcessor
	// KeyName: history.replicatorTaskMaxRetryCount
	// Default value: 100
	ReplicatorTaskMaxRetryCount
	// ReplicatorProcessorMaxPollRPS is max poll rate per second for ReplicatorProcessor
	// KeyName: history.replicatorProcessorMaxPollRPS
	// Default value: 20
	ReplicatorProcessorMaxPollRPS
	// ReplicatorProcessorMaxPollInterval is max poll interval for ReplicatorProcessor
	// KeyName: history.replicatorProcessorMaxPollInterval
	// Default value: 1*time.Minute
	ReplicatorProcessorMaxPollInterval
	// ReplicatorProcessorMaxPollIntervalJitterCoefficient is the max poll interval jitter coefficient
	// KeyName: history.replicatorProcessorMaxPollIntervalJitterCoefficient
	// Default value: 0.15
	ReplicatorProcessorMaxPollIntervalJitterCoefficient
	// ReplicatorProcessorUpdateAckInterval is update interval for ReplicatorProcessor
	// KeyName: history.replicatorProcessorUpdateAckInterval
	// Default value: 5*time.Second
	ReplicatorProcessorUpdateAckInterval
	// ReplicatorProcessorUpdateAckIntervalJitterCoefficient is the update interval jitter coefficient
	// KeyName: history.replicatorProcessorUpdateAckIntervalJitterCoefficient
	// Default value: 0.15
	ReplicatorProcessorUpdateAckIntervalJitterCoefficient
	// ReplicatorProcessorMaxRedispatchQueueSize is the threshold of the number of tasks in the redispatch queue for ReplicatorProcessor
	// KeyName: history.replicatorProcessorMaxRedispatchQueueSize
	// Default value: 10000
	ReplicatorProcessorMaxRedispatchQueueSize
	// ReplicatorProcessorEnablePriorityTaskProcessor is indicates whether priority task processor should be used for ReplicatorProcessor
	// KeyName: history.replicatorProcessorEnablePriorityTaskProcessor
	// Default value: FALSE
	ReplicatorProcessorEnablePriorityTaskProcessor
	// ExecutionMgrNumConns is persistence connections number for ExecutionManager
	// KeyName: history.executionMgrNumConns
	// Default value: 50
	ExecutionMgrNumConns
	// HistoryMgrNumConns is persistence connections number for HistoryManager
	// KeyName: history.historyMgrNumConns
	// Default value: 50
	HistoryMgrNumConns
	// MaximumBufferedEventsBatch is max number of buffer event in mutable state
	// KeyName: history.maximumBufferedEventsBatch
	// Default value: 100
	MaximumBufferedEventsBatch
	// MaximumSignalsPerExecution is max number of signals supported by single execution
	// KeyName: history.maximumSignalsPerExecution
	// Default value: 0
	MaximumSignalsPerExecution
	// ShardUpdateMinInterval is the minimal time interval which the shard info can be updated
	// KeyName: history.shardUpdateMinInterval
	// Default value: 5*time.Minute
	ShardUpdateMinInterval
	// ShardSyncMinInterval is the minimal time interval which the shard info should be sync to remote
	// KeyName: history.shardSyncMinInterval
	// Default value: 5*time.Minute
	ShardSyncMinInterval
	// DefaultEventEncoding is the encoding type for history events
	// KeyName: history.defaultEventEncoding
	// Default value: string(common.EncodingTypeThriftRW)
	DefaultEventEncoding
	// NumArchiveSystemWorkflows is key for number of archive system workflows running in total
	// KeyName: history.numArchiveSystemWorkflows
	// Default value: 1000
	NumArchiveSystemWorkflows
	// ArchiveRequestRPS is the rate limit on the number of archive request per second
	// KeyName: history.archiveRequestRPS
	// Default value: 300 // should be much smaller than frontend RPS
	ArchiveRequestRPS
	// EnableAdminProtection is whether to enable admin checking
	// KeyName: history.enableAdminProtection
	// Default value: FALSE
	EnableAdminProtection
	// AdminOperationToken is the token to pass admin checking
	// KeyName: history.adminOperationToken
	// Default value: common.DefaultAdminOperationToken
	AdminOperationToken
	// HistoryMaxAutoResetPoints is the key for max number of auto reset points stored in mutableState
	// KeyName: history.historyMaxAutoResetPoints
	// Default value: DefaultHistoryMaxAutoResetPoints
	HistoryMaxAutoResetPoints
	// EnableParentClosePolicy is whether to  ParentClosePolicy
	// KeyName: history.enableParentClosePolicy
	// Default value: TRUE
	EnableParentClosePolicy
	// ParentClosePolicyThreshold is decides that parent close policy will be processed by sys workers(if enabled) ifthe number of children greater than or equal to this threshold
	// KeyName: history.parentClosePolicyThreshold
	// Default value: 10
	ParentClosePolicyThreshold
	// NumParentClosePolicySystemWorkflows is key for number of parentClosePolicy system workflows running in total
	// KeyName: history.numParentClosePolicySystemWorkflows
	// Default value: 10
	NumParentClosePolicySystemWorkflows
	// HistoryThrottledLogRPS is the rate limit on number of log messages emitted per second for throttled logger
	// KeyName: history.throttledLogRPS
	// Default value: 4
	HistoryThrottledLogRPS
	// StickyTTL is to expire a sticky tasklist if no update more than this duration
	// KeyName: history.stickyTTL
	// Default value: time.Hour*24*365
	StickyTTL
	// DecisionHeartbeatTimeout is for decision heartbeat
	// KeyName: history.decisionHeartbeatTimeout
	// Default value: time.Minute*30
	DecisionHeartbeatTimeout
	// DecisionRetryCriticalAttempts is decision attempt threshold for logging and emiting metrics
	// KeyName: history.decisionRetryCriticalAttempts
	// Default value: 10
	DecisionRetryCriticalAttempts
	// EnableDropStuckTaskByDomainID is whether stuck timer/transfer task should be dropped for a domain
	// KeyName: history.DropStuckTaskByDomain
	// Default value: FALSE
	EnableDropStuckTaskByDomainID
	// EnableConsistentQuery is indicates if consistent query is enabled for the cluster
	// KeyName: history.EnableConsistentQuery
	// Default value: TRUE
	EnableConsistentQuery
	// EnableConsistentQueryByDomain is indicates if consistent query is enabled for a domain
	// KeyName: history.EnableConsistentQueryByDomain
	// Default value: FALSE
	EnableConsistentQueryByDomain
	// MaxBufferedQueryCount is indicates the maximum number of queries which can be buffered at a given time for a single workflow
	// KeyName: history.MaxBufferedQueryCount
	// Default value: 1
	MaxBufferedQueryCount
	// MutableStateChecksumGenProbability is the probability [0-100] that checksum will be generated for mutable state
	// KeyName: history.mutableStateChecksumGenProbability
	// Default value: 0
	MutableStateChecksumGenProbability
	// MutableStateChecksumVerifyProbability is the probability [0-100] that checksum will be verified for mutable state
	// KeyName: history.mutableStateChecksumVerifyProbability
	// Default value: 0
	MutableStateChecksumVerifyProbability
	// MutableStateChecksumInvalidateBefore is the epoch timestamp before which all checksums are to be discarded
	// KeyName: history.mutableStateChecksumInvalidateBefore
	// Default value: 0
	MutableStateChecksumInvalidateBefore
	// ReplicationEventsFromCurrentCluster is a feature flag to allow cross DC replicate events that generated from the current cluster
	// KeyName: history.ReplicationEventsFromCurrentCluster
	// Default value: FALSE
	ReplicationEventsFromCurrentCluster
	// NotifyFailoverMarkerInterval is determines the frequency to notify failover marker
	// KeyName: history.NotifyFailoverMarkerInterval
	// Default value: 5*time.Second
	NotifyFailoverMarkerInterval
	// NotifyFailoverMarkerTimerJitterCoefficient is the jitter for failover marker notifier timer
	// KeyName: history.NotifyFailoverMarkerTimerJitterCoefficient
	// Default value: 0.15
	NotifyFailoverMarkerTimerJitterCoefficient
	// EnableActivityLocalDispatchByDomain is allows worker to dispatch activity tasks through local tunnel after decisions are made. This is an performance optimization to skip activity scheduling efforts
	// KeyName: history.enableActivityLocalDispatchByDomain
	// Default value: FALSE
	EnableActivityLocalDispatchByDomain
	// HistoryErrorInjectionRate is rate for injecting random error in history client
	// KeyName: history.errorInjectionRate
	// Default value: 0
	HistoryErrorInjectionRate
	// HistoryEnableTaskInfoLogByDomainID is enables info level logs for decision/activity task based on the request domainID
	// KeyName: history.enableTaskInfoLogByDomainID
	// Default value: FALSE
	HistoryEnableTaskInfoLogByDomainID
	// ActivityMaxScheduleToStartTimeoutForRetry is maximum value allowed when overwritting the schedule to start timeout for activities with retry policy
	// KeyName: history.activityMaxScheduleToStartTimeoutForRetry
	// Default value: 30*time.Minute
	ActivityMaxScheduleToStartTimeoutForRetry

	// key for history replication

	//ReplicationTaskFetcherParallelism determines how many go routines we spin up for fetching tasks
	// KeyName: history.ReplicationTaskFetcherParallelism
	// Default value: 1
	ReplicationTaskFetcherParallelism
	// ReplicationTaskFetcherAggregationInterval determines how frequently the fetch requests are sent
	// KeyName: history.ReplicationTaskFetcherAggregationInterval
	// Default value: 2 * time.Second
	ReplicationTaskFetcherAggregationInterval
	// ReplicationTaskFetcherTimerJitterCoefficient is the jitter for fetcher timer
	// KeyName: history.ReplicationTaskFetcherTimerJitterCoefficient
	// Default value: 0.15
	ReplicationTaskFetcherTimerJitterCoefficient
	// ReplicationTaskFetcherErrorRetryWait is the wait time when fetcher encounters error
	// KeyName: history.ReplicationTaskFetcherErrorRetryWait
	// Default value: time.Second
	ReplicationTaskFetcherErrorRetryWait
	// ReplicationTaskFetcherServiceBusyWait is the wait time when fetcher encounters service busy error
	// KeyName: history.ReplicationTaskFetcherServiceBusyWait
	// Default value: 60 * time.Second
	ReplicationTaskFetcherServiceBusyWait
	// ReplicationTaskProcessorErrorRetryWait is the initial retry wait when we see errors in applying replication tasks
	// KeyName: history.ReplicationTaskProcessorErrorRetryWait
	// Default value: 50*time.Millisecond
	ReplicationTaskProcessorErrorRetryWait
	// ReplicationTaskProcessorErrorRetryMaxAttempts is the max retry attempts for applying replication tasks
	// KeyName: history.ReplicationTaskProcessorErrorRetryMaxAttempts
	// Default value: 10
	ReplicationTaskProcessorErrorRetryMaxAttempts
	// ReplicationTaskProcessorErrorSecondRetryWait is the initial retry wait for the second phase retry
	// KeyName: history.ReplicationTaskProcessorErrorSecondRetryWait
	// Default value: 5 * time.Second
	ReplicationTaskProcessorErrorSecondRetryWait
	// ReplicationTaskProcessorErrorSecondRetryMaxWait is the max wait time for the second phase retry
	// KeyName: history.ReplicationTaskProcessorErrorSecondRetryMaxWait
	// Default value: 30 * 5 * time.Second
	ReplicationTaskProcessorErrorSecondRetryMaxWait
	// ReplicationTaskProcessorErrorSecondRetryExpiration is the expiration duration for the second phase retry
	// KeyName: history.ReplicationTaskProcessorErrorSecondRetryExpiration
	// Default value: 5 * time.Minute
	ReplicationTaskProcessorErrorSecondRetryExpiration
	// ReplicationTaskProcessorNoTaskInitialWait is the wait time when not ask is returned
	// KeyName: history.ReplicationTaskProcessorNoTaskInitialWait
	// Default value: 2 * time.Second
	ReplicationTaskProcessorNoTaskInitialWait
	// ReplicationTaskProcessorCleanupInterval determines how frequently the cleanup replication queue
	// KeyName: history.ReplicationTaskProcessorCleanupInterval
	// Default value: 1 * time.Minute
	ReplicationTaskProcessorCleanupInterval
	// ReplicationTaskProcessorCleanupJitterCoefficient is the jitter for cleanup timer
	// KeyName: history.ReplicationTaskProcessorCleanupJitterCoefficient
	// Default value: 0.15
	ReplicationTaskProcessorCleanupJitterCoefficient
	// ReplicationTaskProcessorReadHistoryBatchSize is the batch size to read history events
	// KeyName: history.ReplicationTaskProcessorReadHistoryBatchSize
	// Default value: 5
	ReplicationTaskProcessorReadHistoryBatchSize
	// ReplicationTaskProcessorStartWait is the wait time before each task processing batch
	// KeyName: history.ReplicationTaskProcessorStartWait
	// Default value: 5 * time.Second
	ReplicationTaskProcessorStartWait
	// ReplicationTaskProcessorStartWaitJitterCoefficient is the jitter for batch start wait timer
	// KeyName: history.ReplicationTaskProcessorStartWaitJitterCoefficient
	// Default value: 0.9
	ReplicationTaskProcessorStartWaitJitterCoefficient
	// ReplicationTaskProcessorHostQPS is the qps of task processing rate limiter on host level
	// KeyName: history.ReplicationTaskProcessorHostQPS
	// Default value: 1500
	ReplicationTaskProcessorHostQPS
	// ReplicationTaskProcessorShardQPS is the qps of task processing rate limiter on shard level
	// KeyName: history.ReplicationTaskProcessorShardQPS
	// Default value: 5
	ReplicationTaskProcessorShardQPS
	//ReplicationTaskGenerationQPS is the wait time between each replication task generation qps
	// KeyName: history.ReplicationTaskGenerationQPS
	// Default value: 100
	ReplicationTaskGenerationQPS

	// key for worker

	// WorkerPersistenceMaxQPS is the max qps worker host can query DB
	// KeyName: worker.persistenceMaxQPS
	// Default value: 500
	WorkerPersistenceMaxQPS
	// WorkerPersistenceGlobalMaxQPS is the max qps worker cluster can query DB
	// KeyName: worker.persistenceGlobalMaxQPS
	// Default value: 0
	WorkerPersistenceGlobalMaxQPS
	// WorkerReplicationTaskMaxRetryDuration is the max retry duration for any task
	// KeyName: worker.replicationTaskMaxRetryDuration
	// Default value: #N/A
	WorkerReplicationTaskMaxRetryDuration
	// WorkerIndexerConcurrency is the max concurrent messages to be processed at any given time
	// KeyName: worker.indexerConcurrency
	// Default value: 1000
	WorkerIndexerConcurrency
	// WorkerESProcessorNumOfWorkers is num of workers for esProcessor
	// KeyName: worker.ESProcessorNumOfWorkers
	// Default value: 1
	WorkerESProcessorNumOfWorkers
	// WorkerESProcessorBulkActions is max number of requests in bulk for esProcessor
	// KeyName: worker.ESProcessorBulkActions
	// Default value: 1000
	WorkerESProcessorBulkActions
	// WorkerESProcessorBulkSize is max total size of bulk in bytes for esProcessor
	// KeyName: worker.ESProcessorBulkSize
	// Default value: 2<<24 // 16MB
	WorkerESProcessorBulkSize
	// WorkerESProcessorFlushInterval is flush interval for esProcessor
	// KeyName: worker.ESProcessorFlushInterval
	// Default value: 1*time.Second
	WorkerESProcessorFlushInterval
	// WorkerArchiverConcurrency is controls the number of coroutines handling archival work per archival workflow
	// KeyName: worker.ArchiverConcurrency
	// Default value: 50
	WorkerArchiverConcurrency
	// WorkerArchivalsPerIteration is controls the number of archivals handled in each iteration of archival workflow
	// KeyName: worker.ArchivalsPerIteration
	// Default value: 1000
	WorkerArchivalsPerIteration
	// WorkerTimeLimitPerArchivalIteration is controls the time limit of each iteration of archival workflow
	// KeyName: worker.TimeLimitPerArchivalIteration
	// Default value: archiver.MaxArchivalIterationTimeout()
	WorkerTimeLimitPerArchivalIteration
	// WorkerThrottledLogRPS is the rate limit on number of log messages emitted per second for throttled logger
	// KeyName: worker.throttledLogRPS
	// Default value: 20
	WorkerThrottledLogRPS
	// ScannerPersistenceMaxQPS is the maximum rate of persistence calls from worker.Scanner
	// KeyName: worker.scannerPersistenceMaxQPS
	// Default value: 100
	ScannerPersistenceMaxQPS
	// ScannerGetOrphanTasksPageSize is the maximum number of orphans to delete in one batch
	// KeyName: worker.scannerGetOrphanTasksPageSize
	// Default value: 1000
	ScannerGetOrphanTasksPageSize
	// ScannerBatchSizeForTasklistHandler is for:
	//  1. max number of tasks to query per call(get tasks for tasklist) in the scavenger handler.
	//  2. The scavenger then uses the return to decide if a tasklist can be deleted.
	// It's better to keep it a relatively high number to let it be more efficient.
	// KeyName: worker.scannerBatchSizeForTasklistHandler
	// Default value: 16
	ScannerBatchSizeForTasklistHandler
	// EnableCleaningOrphanTaskInTasklistScavenger indicates if enabling the scanner to clean up orphan tasks
	// KeyName: worker.enableCleaningOrphanTaskInTasklistScavenger
	// Default value: false
	EnableCleaningOrphanTaskInTasklistScavenger
	// ScannerMaxTasksProcessedPerTasklistJob is the number of tasks to process for a tasklist in each workflow run
	// KeyName: worker.scannerMaxTasksProcessedPerTasklistJob
	// Default value: 256
	ScannerMaxTasksProcessedPerTasklistJob
	// TaskListScannerEnabled is indicates if task list scanner should be started as part of worker.Scanner
	// KeyName: worker.taskListScannerEnabled
	// Default value: TRUE
	TaskListScannerEnabled
	// HistoryScannerEnabled is indicates if history scanner should be started as part of worker.Scanner
	// KeyName: worker.historyScannerEnabled
	// Default value: TRUE
	HistoryScannerEnabled
	// ConcreteExecutionsScannerEnabled is indicates if executions scanner should be started as part of worker.Scanner
	// KeyName: worker.executionsScannerEnabled
	// Default value: FALSE
	ConcreteExecutionsScannerEnabled
	// ConcreteExecutionsScannerConcurrency is indicates the concurrency of concrete execution scanner
	// KeyName: worker.executionsScannerConcurrency
	// Default value: 25
	ConcreteExecutionsScannerConcurrency
	// ConcreteExecutionsScannerBlobstoreFlushThreshold is indicates the flush threshold of blobstore in concrete execution scanner
	// KeyName: worker.executionsScannerBlobstoreFlushThreshold
	// Default value: 100
	ConcreteExecutionsScannerBlobstoreFlushThreshold
	// ConcreteExecutionsScannerActivityBatchSize is indicates the batch size of scanner activities
	// KeyName: worker.executionsScannerActivityBatchSize
	// Default value: 25
	ConcreteExecutionsScannerActivityBatchSize
	// ConcreteExecutionsScannerPersistencePageSize is indicates the page size of execution persistence fetches in concrete execution scanner
	// KeyName: worker.executionsScannerPersistencePageSize
	// Default value: 1000
	ConcreteExecutionsScannerPersistencePageSize
	// ConcreteExecutionsScannerInvariantCollectionMutableState is indicates if mutable state invariant checks should be run
	// KeyName: worker.executionsScannerInvariantCollectionMutableState
	// Default value: TRUE
	ConcreteExecutionsScannerInvariantCollectionMutableState
	// ConcreteExecutionsScannerInvariantCollectionHistory is indicates if history invariant checks should be run
	// KeyName: worker.executionsScannerInvariantCollectionHistory
	// Default value: TRUE
	ConcreteExecutionsScannerInvariantCollectionHistory
	// CurrentExecutionsScannerEnabled is indicates if current executions scanner should be started as part of worker.Scanner
	// KeyName: worker.currentExecutionsScannerEnabled
	// Default value: FALSE
	CurrentExecutionsScannerEnabled
	// CurrentExecutionsScannerConcurrency is indicates the concurrency of current executions scanner
	// KeyName: worker.currentExecutionsConcurrency
	// Default value: 25
	CurrentExecutionsScannerConcurrency
	// CurrentExecutionsScannerBlobstoreFlushThreshold is indicates the flush threshold of blobstore in current executions scanner
	// KeyName: worker.currentExecutionsBlobstoreFlushThreshold
	// Default value: 100
	CurrentExecutionsScannerBlobstoreFlushThreshold
	// CurrentExecutionsScannerActivityBatchSize is indicates the batch size of scanner activities
	// KeyName: worker.currentExecutionsActivityBatchSize
	// Default value: 25
	CurrentExecutionsScannerActivityBatchSize
	// CurrentExecutionsScannerPersistencePageSize is indicates the page size of execution persistence fetches in current executions scanner
	// KeyName: worker.currentExecutionsPersistencePageSize
	// Default value: 1000
	CurrentExecutionsScannerPersistencePageSize
	// CurrentExecutionsScannerInvariantCollectionHistory is indicates if history invariant checks should be run
	// KeyName: worker.currentExecutionsScannerInvariantCollectionHistory
	// Default value: FALSE
	CurrentExecutionsScannerInvariantCollectionHistory
	// CurrentExecutionsScannerInvariantCollectionMutableState is indicates if mutable state invariant checks should be run
	// KeyName: worker.currentExecutionsInvariantCollectionMutableState
	// Default value: TRUE
	CurrentExecutionsScannerInvariantCollectionMutableState
	// EnableBatcher is decides whether start batcher in our worker
	// KeyName: worker.enableBatcher
	// Default value: FALSE
	EnableBatcher
	// EnableParentClosePolicyWorker is decides whether or not enable system workers for processing parent close policy task
	// KeyName: system.enableParentClosePolicyWorker
	// Default value: TRUE
	EnableParentClosePolicyWorker
	// EnableStickyQuery is indicates if sticky query should be enabled per domain
	// KeyName: system.enableStickyQuery
	// Default value: TRUE
	EnableStickyQuery
	// EnableFailoverManager is indicates if failover manager is enabled
	// KeyName: system.enableFailoverManager
	// Default value: FALSE
	EnableFailoverManager
	// EnableWorkflowShadower indicates if workflow shadower is enabled
	// KeyName: system.enableWorkflowShadower
	// Default value: false
	EnableWorkflowShadower
	// ConcreteExecutionFixerDomainAllow is which domains are allowed to be fixed by concrete fixer workflow
	// KeyName: worker.concreteExecutionFixerDomainAllow
	// Default value: FALSE
	ConcreteExecutionFixerDomainAllow
	// CurrentExecutionFixerDomainAllow is which domains are allowed to be fixed by current fixer workflow
	// KeyName: worker.currentExecutionFixerDomainAllow
	// Default value: FALSE
	CurrentExecutionFixerDomainAllow
	// TimersScannerEnabled is if timers scanner should be started as part of worker.Scanner
	// KeyName: worker.timersScannerEnabled
	// Default value: FALSE
	TimersScannerEnabled
	// TimersFixerEnabled is if timers fixer should be started as part of worker.Scanner
	// KeyName: worker.timersFixerEnabled
	// Default value: FALSE
	TimersFixerEnabled
	// TimersScannerConcurrency is the concurrency of timers scanner
	// KeyName: worker.timersScannerConcurrency
	// Default value: 5
	TimersScannerConcurrency
	// TimersScannerPersistencePageSize is the page size of timers persistence fetches in timers scanner
	// KeyName: worker.timersScannerPersistencePageSize
	// Default value: 1000
	TimersScannerPersistencePageSize
	// TimersScannerBlobstoreFlushThreshold is threshold to flush blob store
	// KeyName: worker.timersScannerBlobstoreFlushThreshold
	// Default value: 100
	TimersScannerBlobstoreFlushThreshold
	// TimersScannerActivityBatchSize is TimersScannerActivityBatchSize
	// KeyName: worker.timersScannerActivityBatchSize
	// Default value: 25
	TimersScannerActivityBatchSize
	// TimersScannerPeriodStart is interval start for fetching scheduled timers
	// KeyName: worker.timersScannerPeriodStart
	// Default value: 24
	TimersScannerPeriodStart
	// TimersScannerPeriodEnd is interval end for fetching scheduled timers
	// KeyName: worker.timersScannerPeriodEnd
	// Default value: 3
	TimersScannerPeriodEnd
	// TimersFixerDomainAllow is which domains are allowed to be fixed by timer fixer workflow
	// KeyName: worker.timersFixerDomainAllow
	// Default value: FALSE
	TimersFixerDomainAllow
	// ConcreteExecutionFixerEnabled is if concrete execution fixer workflow is enabled
	// KeyName: worker.concreteExecutionFixerEnabled
	// Default value: FALSE
	ConcreteExecutionFixerEnabled
	// CurrentExecutionFixerEnabled is if current execution fixer workflow is enabled
	// KeyName: worker.currentExecutionFixerEnabled
	// Default value: FALSE
	CurrentExecutionFixerEnabled

	// EnableAuthorization is the key to enable authorization for a domain, only for extension binary:
	// KeyName: N/A
	// Default value: N/A
	// TODO: https://github.com/uber/cadence/issues/3861
	EnableAuthorization
	// Usage: VisibilityArchivalQueryMaxRangeInDays is the maximum number of days for a visibility archival query
	// KeyName: N/A
	// Default value: N/A
	// TODO: https://github.com/uber/cadence/issues/3861
	VisibilityArchivalQueryMaxRangeInDays
	// Usage: VisibilityArchivalQueryMaxQPS is the timeout for a visibility archival query
	// KeyName: N/A
	// Default value: N/A
	// TODO: https://github.com/uber/cadence/issues/3861
	VisibilityArchivalQueryMaxQPS
	// EnableArchivalCompression indicates whether blobs are compressed before they are archived
	// KeyName: N/A
	// Default value: N/A
	// TODO: https://github.com/uber/cadence/issues/3861
	EnableArchivalCompression
	// WorkerDeterministicConstructionCheckProbability controls the probability of running a deterministic construction check for any given archival
	// KeyName: N/A
	// Default value: N/A
	// TODO: https://github.com/uber/cadence/issues/3861
	WorkerDeterministicConstructionCheckProbability
	// WorkerBlobIntegrityCheckProbability controls the probability of running an integrity check for any given archival
	// KeyName: N/A
	// Default value: N/A
	// TODO: https://github.com/uber/cadence/issues/3861
	WorkerBlobIntegrityCheckProbability

	// lastKeyForTest must be the last one in this const group for testing purpose
	lastKeyForTest
)

// Mapping from Key to keyName, where keyName are used dynamic config source.
var keys = map[Key]string{
	unknownKey: "unknownKey",

	// tests keys
	testGetPropertyKey:                               "testGetPropertyKey",
	testGetIntPropertyKey:                            "testGetIntPropertyKey",
	testGetFloat64PropertyKey:                        "testGetFloat64PropertyKey",
	testGetDurationPropertyKey:                       "testGetDurationPropertyKey",
	testGetBoolPropertyKey:                           "testGetBoolPropertyKey",
	testGetStringPropertyKey:                         "testGetStringPropertyKey",
	testGetMapPropertyKey:                            "testGetMapPropertyKey",
	testGetIntPropertyFilteredByDomainKey:            "testGetIntPropertyFilteredByDomainKey",
	testGetDurationPropertyFilteredByDomainKey:       "testGetDurationPropertyFilteredByDomainKey",
	testGetIntPropertyFilteredByTaskListInfoKey:      "testGetIntPropertyFilteredByTaskListInfoKey",
	testGetDurationPropertyFilteredByTaskListInfoKey: "testGetDurationPropertyFilteredByTaskListInfoKey",
	testGetBoolPropertyFilteredByDomainIDKey:         "testGetBoolPropertyFilteredByDomainIDKey",
	testGetBoolPropertyFilteredByTaskListInfoKey:     "testGetBoolPropertyFilteredByTaskListInfoKey",

	// system settings
	EnableGlobalDomain:                  "system.enableGlobalDomain",
	EnableVisibilitySampling:            "system.enableVisibilitySampling",
	EnableReadFromClosedExecutionV2:     "system.enableReadFromClosedExecutionV2",
	AdvancedVisibilityWritingMode:       "system.advancedVisibilityWritingMode",
	EnableReadVisibilityFromES:          "system.enableReadVisibilityFromES",
	HistoryArchivalStatus:               "system.historyArchivalStatus",
	EnableReadFromHistoryArchival:       "system.enableReadFromHistoryArchival",
	VisibilityArchivalStatus:            "system.visibilityArchivalStatus",
	EnableReadFromVisibilityArchival:    "system.enableReadFromVisibilityArchival",
	EnableDomainNotActiveAutoForwarding: "system.enableDomainNotActiveAutoForwarding",
	EnableGracefulFailover:              "system.enableGracefulFailover",
	TransactionSizeLimit:                "system.transactionSizeLimit",
	PersistenceErrorInjectionRate:       "system.persistenceErrorInjectionRate",
	MaxRetentionDays:                    "system.maxRetentionDays",
	MinRetentionDays:                    "system.minRetentionDays",
	MaxDecisionStartToCloseSeconds:      "system.maxDecisionStartToCloseSeconds",
	DisallowQuery:                       "system.disallowQuery",
	EnableBatcher:                       "worker.enableBatcher",
	EnableParentClosePolicyWorker:       "system.enableParentClosePolicyWorker",
	EnableFailoverManager:               "system.enableFailoverManager",
	EnableWorkflowShadower:              "system.enableWorkflowShadower",
	EnableStickyQuery:                   "system.enableStickyQuery",
	EnableDebugMode:                     "system.enableDebugMode",
	RequiredDomainDataKeys:              "system.requiredDomainDataKeys",
	EnableGRPCOutbound:                  "system.enableGRPCOutbound",

	// size limit
	BlobSizeLimitError:      "limit.blobSize.error",
	BlobSizeLimitWarn:       "limit.blobSize.warn",
	HistorySizeLimitError:   "limit.historySize.error",
	HistorySizeLimitWarn:    "limit.historySize.warn",
	HistoryCountLimitError:  "limit.historyCount.error",
	HistoryCountLimitWarn:   "limit.historyCount.warn",
	MaxIDLengthLimit:        "limit.maxIDLength",
	MaxIDLengthWarnLimit:    "limit.maxIDWarnLength",
	MaxRawTaskListNameLimit: "limit.maxRawTaskListNameLength",

	// admin settings
	AdminErrorInjectionRate: "admin.errorInjectionRate",

	// frontend settings
	FrontendPersistenceMaxQPS:                   "frontend.persistenceMaxQPS",
	FrontendPersistenceGlobalMaxQPS:             "frontend.persistenceGlobalMaxQPS",
	FrontendVisibilityMaxPageSize:               "frontend.visibilityMaxPageSize",
	FrontendVisibilityListMaxQPS:                "frontend.visibilityListMaxQPS",
	FrontendESVisibilityListMaxQPS:              "frontend.esVisibilityListMaxQPS",
	FrontendMaxBadBinaries:                      "frontend.maxBadBinaries",
	FrontendFailoverCoolDown:                    "frontend.failoverCoolDown",
	FrontendESIndexMaxResultWindow:              "frontend.esIndexMaxResultWindow",
	FrontendHistoryMaxPageSize:                  "frontend.historyMaxPageSize",
	FrontendRPS:                                 "frontend.rps",
	FrontendMaxDomainRPSPerInstance:             "frontend.domainrps",
	FrontendGlobalDomainRPS:                     "frontend.globalDomainrps",
	FrontendHistoryMgrNumConns:                  "frontend.historyMgrNumConns",
	FrontendShutdownDrainDuration:               "frontend.shutdownDrainDuration",
	DisableListVisibilityByFilter:               "frontend.disableListVisibilityByFilter",
	FrontendThrottledLogRPS:                     "frontend.throttledLogRPS",
	EnableClientVersionCheck:                    "frontend.enableClientVersionCheck",
	ValidSearchAttributes:                       "frontend.validSearchAttributes",
	SendRawWorkflowHistory:                      "frontend.sendRawWorkflowHistory",
	SearchAttributesNumberOfKeysLimit:           "frontend.searchAttributesNumberOfKeysLimit",
	SearchAttributesSizeOfValueLimit:            "frontend.searchAttributesSizeOfValueLimit",
	SearchAttributesTotalSizeLimit:              "frontend.searchAttributesTotalSizeLimit",
	VisibilityArchivalQueryMaxPageSize:          "frontend.visibilityArchivalQueryMaxPageSize",
	DomainFailoverRefreshInterval:               "frontend.domainFailoverRefreshInterval",
	DomainFailoverRefreshTimerJitterCoefficient: "frontend.domainFailoverRefreshTimerJitterCoefficient",
	FrontendErrorInjectionRate:                  "frontend.errorInjectionRate",

	// matching settings
	MatchingRPS:                             "matching.rps",
	MatchingPersistenceMaxQPS:               "matching.persistenceMaxQPS",
	MatchingPersistenceGlobalMaxQPS:         "matching.persistenceGlobalMaxQPS",
	MatchingMinTaskThrottlingBurstSize:      "matching.minTaskThrottlingBurstSize",
	MatchingGetTasksBatchSize:               "matching.getTasksBatchSize",
	MatchingLongPollExpirationInterval:      "matching.longPollExpirationInterval",
	MatchingEnableSyncMatch:                 "matching.enableSyncMatch",
	MatchingUpdateAckInterval:               "matching.updateAckInterval",
	MatchingIdleTasklistCheckInterval:       "matching.idleTasklistCheckInterval",
	MaxTasklistIdleTime:                     "matching.maxTasklistIdleTime",
	MatchingOutstandingTaskAppendsThreshold: "matching.outstandingTaskAppendsThreshold",
	MatchingMaxTaskBatchSize:                "matching.maxTaskBatchSize",
	MatchingMaxTaskDeleteBatchSize:          "matching.maxTaskDeleteBatchSize",
	MatchingThrottledLogRPS:                 "matching.throttledLogRPS",
	MatchingNumTasklistWritePartitions:      "matching.numTasklistWritePartitions",
	MatchingNumTasklistReadPartitions:       "matching.numTasklistReadPartitions",
	MatchingForwarderMaxOutstandingPolls:    "matching.forwarderMaxOutstandingPolls",
	MatchingForwarderMaxOutstandingTasks:    "matching.forwarderMaxOutstandingTasks",
	MatchingForwarderMaxRatePerSecond:       "matching.forwarderMaxRatePerSecond",
	MatchingForwarderMaxChildrenPerNode:     "matching.forwarderMaxChildrenPerNode",
	MatchingShutdownDrainDuration:           "matching.shutdownDrainDuration",
	MatchingErrorInjectionRate:              "matching.errorInjectionRate",
	MatchingEnableTaskInfoLogByDomainID:     "matching.enableTaskInfoLogByDomainID",

	// history settings
	HistoryRPS:                                            "history.rps",
	HistoryPersistenceMaxQPS:                              "history.persistenceMaxQPS",
	HistoryPersistenceGlobalMaxQPS:                        "history.persistenceGlobalMaxQPS",
	HistoryVisibilityOpenMaxQPS:                           "history.historyVisibilityOpenMaxQPS",
	HistoryVisibilityClosedMaxQPS:                         "history.historyVisibilityClosedMaxQPS",
	HistoryLongPollExpirationInterval:                     "history.longPollExpirationInterval",
	HistoryCacheInitialSize:                               "history.cacheInitialSize",
	HistoryMaxAutoResetPoints:                             "history.historyMaxAutoResetPoints",
	HistoryCacheMaxSize:                                   "history.cacheMaxSize",
	HistoryCacheTTL:                                       "history.cacheTTL",
	HistoryShutdownDrainDuration:                          "history.shutdownDrainDuration",
	EventsCacheInitialCount:                               "history.eventsCacheInitialSize",
	EventsCacheMaxCount:                                   "history.eventsCacheMaxSize",
	EventsCacheMaxSize:                                    "history.eventsCacheMaxSizeInBytes",
	EventsCacheTTL:                                        "history.eventsCacheTTL",
	EventsCacheGlobalEnable:                               "history.eventsCacheGlobalEnable",
	EventsCacheGlobalInitialCount:                         "history.eventsCacheGlobalInitialSize",
	EventsCacheGlobalMaxCount:                             "history.eventsCacheGlobalMaxSize",
	AcquireShardInterval:                                  "history.acquireShardInterval",
	AcquireShardConcurrency:                               "history.acquireShardConcurrency",
	StandbyClusterDelay:                                   "history.standbyClusterDelay",
	StandbyTaskMissingEventsResendDelay:                   "history.standbyTaskMissingEventsResendDelay",
	StandbyTaskMissingEventsDiscardDelay:                  "history.standbyTaskMissingEventsDiscardDelay",
	TaskProcessRPS:                                        "history.taskProcessRPS",
	TaskSchedulerType:                                     "history.taskSchedulerType",
	TaskSchedulerWorkerCount:                              "history.taskSchedulerWorkerCount",
	TaskSchedulerShardWorkerCount:                         "history.taskSchedulerShardWorkerCount",
	TaskSchedulerQueueSize:                                "history.taskSchedulerQueueSize",
	TaskSchedulerShardQueueSize:                           "history.taskSchedulerShardQueueSize",
	TaskSchedulerDispatcherCount:                          "history.taskSchedulerDispatcherCount",
	TaskSchedulerRoundRobinWeights:                        "history.taskSchedulerRoundRobinWeight",
	ActiveTaskRedispatchInterval:                          "history.activeTaskRedispatchInterval",
	StandbyTaskRedispatchInterval:                         "history.standbyTaskRedispatchInterval",
	TaskRedispatchIntervalJitterCoefficient:               "history.taskRedispatchIntervalJitterCoefficient",
	StandbyTaskReReplicationContextTimeout:                "history.standbyTaskReReplicationContextTimeout",
	QueueProcessorEnableSplit:                             "history.queueProcessorEnableSplit",
	QueueProcessorSplitMaxLevel:                           "history.queueProcessorSplitMaxLevel",
	QueueProcessorEnableRandomSplitByDomainID:             "history.queueProcessorEnableRandomSplitByDomainID",
	QueueProcessorRandomSplitProbability:                  "history.queueProcessorRandomSplitProbability",
	QueueProcessorEnablePendingTaskSplitByDomainID:        "history.queueProcessorEnablePendingTaskSplitByDomainID",
	QueueProcessorPendingTaskSplitThreshold:               "history.queueProcessorPendingTaskSplitThreshold",
	QueueProcessorEnableStuckTaskSplitByDomainID:          "history.queueProcessorEnableStuckTaskSplitByDomainID",
	QueueProcessorStuckTaskSplitThreshold:                 "history.queueProcessorStuckTaskSplitThreshold",
	QueueProcessorSplitLookAheadDurationByDomainID:        "history.queueProcessorSplitLookAheadDurationByDomainID",
	QueueProcessorPollBackoffInterval:                     "history.queueProcessorPollBackoffInterval",
	QueueProcessorPollBackoffIntervalJitterCoefficient:    "history.queueProcessorPollBackoffIntervalJitterCoefficient",
	QueueProcessorEnablePersistQueueStates:                "history.queueProcessorEnablePersistQueueStates",
	QueueProcessorEnableLoadQueueStates:                   "history.queueProcessorEnableLoadQueueStates",
	TimerTaskBatchSize:                                    "history.timerTaskBatchSize",
	TimerTaskWorkerCount:                                  "history.timerTaskWorkerCount",
	TimerTaskMaxRetryCount:                                "history.timerTaskMaxRetryCount",
	TimerProcessorGetFailureRetryCount:                    "history.timerProcessorGetFailureRetryCount",
	TimerProcessorCompleteTimerFailureRetryCount:          "history.timerProcessorCompleteTimerFailureRetryCount",
	TimerProcessorUpdateAckInterval:                       "history.timerProcessorUpdateAckInterval",
	TimerProcessorUpdateAckIntervalJitterCoefficient:      "history.timerProcessorUpdateAckIntervalJitterCoefficient",
	TimerProcessorCompleteTimerInterval:                   "history.timerProcessorCompleteTimerInterval",
	TimerProcessorFailoverMaxPollRPS:                      "history.timerProcessorFailoverMaxPollRPS",
	TimerProcessorMaxPollRPS:                              "history.timerProcessorMaxPollRPS",
	TimerProcessorMaxPollInterval:                         "history.timerProcessorMaxPollInterval",
	TimerProcessorMaxPollIntervalJitterCoefficient:        "history.timerProcessorMaxPollIntervalJitterCoefficient",
	TimerProcessorSplitQueueInterval:                      "history.timerProcessorSplitQueueInterval",
	TimerProcessorSplitQueueIntervalJitterCoefficient:     "history.timerProcessorSplitQueueIntervalJitterCoefficient",
	TimerProcessorMaxRedispatchQueueSize:                  "history.timerProcessorMaxRedispatchQueueSize",
	TimerProcessorMaxTimeShift:                            "history.timerProcessorMaxTimeShift",
	TimerProcessorHistoryArchivalSizeLimit:                "history.timerProcessorHistoryArchivalSizeLimit",
	TimerProcessorArchivalTimeLimit:                       "history.timerProcessorArchivalTimeLimit",
	TransferTaskBatchSize:                                 "history.transferTaskBatchSize",
	TransferProcessorFailoverMaxPollRPS:                   "history.transferProcessorFailoverMaxPollRPS",
	TransferProcessorMaxPollRPS:                           "history.transferProcessorMaxPollRPS",
	TransferTaskWorkerCount:                               "history.transferTaskWorkerCount",
	TransferTaskMaxRetryCount:                             "history.transferTaskMaxRetryCount",
	TransferProcessorCompleteTransferFailureRetryCount:    "history.transferProcessorCompleteTransferFailureRetryCount",
	TransferProcessorMaxPollInterval:                      "history.transferProcessorMaxPollInterval",
	TransferProcessorMaxPollIntervalJitterCoefficient:     "history.transferProcessorMaxPollIntervalJitterCoefficient",
	TransferProcessorSplitQueueInterval:                   "history.transferProcessorSplitQueueInterval",
	TransferProcessorSplitQueueIntervalJitterCoefficient:  "history.transferProcessorSplitQueueIntervalJitterCoefficient",
	TransferProcessorUpdateAckInterval:                    "history.transferProcessorUpdateAckInterval",
	TransferProcessorUpdateAckIntervalJitterCoefficient:   "history.transferProcessorUpdateAckIntervalJitterCoefficient",
	TransferProcessorCompleteTransferInterval:             "history.transferProcessorCompleteTransferInterval",
	TransferProcessorMaxRedispatchQueueSize:               "history.transferProcessorMaxRedispatchQueueSize",
	TransferProcessorEnableValidator:                      "history.transferProcessorEnableValidator",
	TransferProcessorValidationInterval:                   "history.transferProcessorValidationInterval",
	TransferProcessorVisibilityArchivalTimeLimit:          "history.transferProcessorVisibilityArchivalTimeLimit",
	ReplicatorTaskBatchSize:                               "history.replicatorTaskBatchSize",
	ReplicatorTaskWorkerCount:                             "history.replicatorTaskWorkerCount",
	ReplicatorReadTaskMaxRetryCount:                       "history.replicatorReadTaskMaxRetryCount",
	ReplicatorTaskMaxRetryCount:                           "history.replicatorTaskMaxRetryCount",
	ReplicatorProcessorMaxPollRPS:                         "history.replicatorProcessorMaxPollRPS",
	ReplicatorProcessorMaxPollInterval:                    "history.replicatorProcessorMaxPollInterval",
	ReplicatorProcessorMaxPollIntervalJitterCoefficient:   "history.replicatorProcessorMaxPollIntervalJitterCoefficient",
	ReplicatorProcessorUpdateAckInterval:                  "history.replicatorProcessorUpdateAckInterval",
	ReplicatorProcessorUpdateAckIntervalJitterCoefficient: "history.replicatorProcessorUpdateAckIntervalJitterCoefficient",
	ReplicatorProcessorMaxRedispatchQueueSize:             "history.replicatorProcessorMaxRedispatchQueueSize",
	ReplicatorProcessorEnablePriorityTaskProcessor:        "history.replicatorProcessorEnablePriorityTaskProcessor",
	ExecutionMgrNumConns:                                  "history.executionMgrNumConns",
	HistoryMgrNumConns:                                    "history.historyMgrNumConns",
	MaximumBufferedEventsBatch:                            "history.maximumBufferedEventsBatch",
	MaximumSignalsPerExecution:                            "history.maximumSignalsPerExecution",
	ShardUpdateMinInterval:                                "history.shardUpdateMinInterval",
	ShardSyncMinInterval:                                  "history.shardSyncMinInterval",
	DefaultEventEncoding:                                  "history.defaultEventEncoding",
	EnableAdminProtection:                                 "history.enableAdminProtection",
	AdminOperationToken:                                   "history.adminOperationToken",
	EnableParentClosePolicy:                               "history.enableParentClosePolicy",
	NumArchiveSystemWorkflows:                             "history.numArchiveSystemWorkflows",
	ArchiveRequestRPS:                                     "history.archiveRequestRPS",
	EmitShardDiffLog:                                      "history.emitShardDiffLog",
	HistoryThrottledLogRPS:                                "history.throttledLogRPS",
	StickyTTL:                                             "history.stickyTTL",
	DecisionHeartbeatTimeout:                              "history.decisionHeartbeatTimeout",
	DecisionRetryCriticalAttempts:                         "history.decisionRetryCriticalAttempts",
	ParentClosePolicyThreshold:                            "history.parentClosePolicyThreshold",
	NumParentClosePolicySystemWorkflows:                   "history.numParentClosePolicySystemWorkflows",
	ReplicationTaskFetcherParallelism:                     "history.ReplicationTaskFetcherParallelism",
	ReplicationTaskFetcherAggregationInterval:             "history.ReplicationTaskFetcherAggregationInterval",
	ReplicationTaskFetcherTimerJitterCoefficient:          "history.ReplicationTaskFetcherTimerJitterCoefficient",
	ReplicationTaskFetcherErrorRetryWait:                  "history.ReplicationTaskFetcherErrorRetryWait",
	ReplicationTaskFetcherServiceBusyWait:                 "history.ReplicationTaskFetcherServiceBusyWait",
	ReplicationTaskProcessorErrorRetryWait:                "history.ReplicationTaskProcessorErrorRetryWait",
	ReplicationTaskProcessorErrorRetryMaxAttempts:         "history.ReplicationTaskProcessorErrorRetryMaxAttempts",
	ReplicationTaskProcessorErrorSecondRetryWait:          "history.ReplicationTaskProcessorErrorSecondRetryWait",
	ReplicationTaskProcessorErrorSecondRetryMaxWait:       "history.ReplicationTaskProcessorErrorSecondRetryMaxWait",
	ReplicationTaskProcessorErrorSecondRetryExpiration:    "history.ReplicationTaskProcessorErrorSecondRetryExpiration",
	ReplicationTaskProcessorNoTaskInitialWait:             "history.ReplicationTaskProcessorNoTaskInitialWait",
	ReplicationTaskProcessorCleanupInterval:               "history.ReplicationTaskProcessorCleanupInterval",
	ReplicationTaskProcessorCleanupJitterCoefficient:      "history.ReplicationTaskProcessorCleanupJitterCoefficient",
	ReplicationTaskProcessorReadHistoryBatchSize:          "history.ReplicationTaskProcessorReadHistoryBatchSize",
	ReplicationTaskProcessorStartWait:                     "history.ReplicationTaskProcessorStartWait",
	ReplicationTaskProcessorStartWaitJitterCoefficient:    "history.ReplicationTaskProcessorStartWaitJitterCoefficient",
	ReplicationTaskProcessorHostQPS:                       "history.ReplicationTaskProcessorHostQPS",
	ReplicationTaskProcessorShardQPS:                      "history.ReplicationTaskProcessorShardQPS",
	ReplicationTaskGenerationQPS:                          "history.ReplicationTaskGenerationQPS",
	EnableConsistentQuery:                                 "history.EnableConsistentQuery",
	EnableConsistentQueryByDomain:                         "history.EnableConsistentQueryByDomain",
	MaxBufferedQueryCount:                                 "history.MaxBufferedQueryCount",
	MutableStateChecksumGenProbability:                    "history.mutableStateChecksumGenProbability",
	MutableStateChecksumVerifyProbability:                 "history.mutableStateChecksumVerifyProbability",
	MutableStateChecksumInvalidateBefore:                  "history.mutableStateChecksumInvalidateBefore",
	ReplicationEventsFromCurrentCluster:                   "history.ReplicationEventsFromCurrentCluster",
	NotifyFailoverMarkerInterval:                          "history.NotifyFailoverMarkerInterval",
	NotifyFailoverMarkerTimerJitterCoefficient:            "history.NotifyFailoverMarkerTimerJitterCoefficient",
	EnableDropStuckTaskByDomainID:                         "history.DropStuckTaskByDomain",
	EnableActivityLocalDispatchByDomain:                   "history.enableActivityLocalDispatchByDomain",
	HistoryErrorInjectionRate:                             "history.errorInjectionRate",
	HistoryEnableTaskInfoLogByDomainID:                    "history.enableTaskInfoLogByDomainID",
	ActivityMaxScheduleToStartTimeoutForRetry:             "history.activityMaxScheduleToStartTimeoutForRetry",

	WorkerPersistenceMaxQPS:                                  "worker.persistenceMaxQPS",
	WorkerPersistenceGlobalMaxQPS:                            "worker.persistenceGlobalMaxQPS",
	WorkerReplicationTaskMaxRetryDuration:                    "worker.replicationTaskMaxRetryDuration",
	WorkerIndexerConcurrency:                                 "worker.indexerConcurrency",
	WorkerESProcessorNumOfWorkers:                            "worker.ESProcessorNumOfWorkers",
	WorkerESProcessorBulkActions:                             "worker.ESProcessorBulkActions",
	WorkerESProcessorBulkSize:                                "worker.ESProcessorBulkSize",
	WorkerESProcessorFlushInterval:                           "worker.ESProcessorFlushInterval",
	WorkerArchiverConcurrency:                                "worker.ArchiverConcurrency",
	WorkerArchivalsPerIteration:                              "worker.ArchivalsPerIteration",
	WorkerTimeLimitPerArchivalIteration:                      "worker.TimeLimitPerArchivalIteration",
	WorkerThrottledLogRPS:                                    "worker.throttledLogRPS",
	ScannerPersistenceMaxQPS:                                 "worker.scannerPersistenceMaxQPS",
	ScannerGetOrphanTasksPageSize:                            "worker.scannerGetOrphanTasksPageSize",
	ScannerBatchSizeForTasklistHandler:                       "worker.scannerBatchSizeForTasklistHandler",
	EnableCleaningOrphanTaskInTasklistScavenger:              "worker.enableCleaningOrphanTaskInTasklistScavenger",
	ScannerMaxTasksProcessedPerTasklistJob:                   "worker.scannerMaxTasksProcessedPerTasklistJob",
	TaskListScannerEnabled:                                   "worker.taskListScannerEnabled",
	HistoryScannerEnabled:                                    "worker.historyScannerEnabled",
	ConcreteExecutionsScannerEnabled:                         "worker.executionsScannerEnabled",
	ConcreteExecutionsScannerBlobstoreFlushThreshold:         "worker.executionsScannerBlobstoreFlushThreshold",
	ConcreteExecutionsScannerActivityBatchSize:               "worker.executionsScannerActivityBatchSize",
	ConcreteExecutionsScannerConcurrency:                     "worker.executionsScannerConcurrency",
	ConcreteExecutionsScannerPersistencePageSize:             "worker.executionsScannerPersistencePageSize",
	ConcreteExecutionsScannerInvariantCollectionHistory:      "worker.executionsScannerInvariantCollectionHistory",
	ConcreteExecutionsScannerInvariantCollectionMutableState: "worker.executionsScannerInvariantCollectionMutableState",
	CurrentExecutionsScannerEnabled:                          "worker.currentExecutionsScannerEnabled",
	CurrentExecutionsScannerBlobstoreFlushThreshold:          "worker.currentExecutionsBlobstoreFlushThreshold",
	CurrentExecutionsScannerActivityBatchSize:                "worker.currentExecutionsActivityBatchSize",
	CurrentExecutionsScannerConcurrency:                      "worker.currentExecutionsConcurrency",
	CurrentExecutionsScannerPersistencePageSize:              "worker.currentExecutionsPersistencePageSize",
	CurrentExecutionsScannerInvariantCollectionHistory:       "worker.currentExecutionsScannerInvariantCollectionHistory",
	CurrentExecutionsScannerInvariantCollectionMutableState:  "worker.currentExecutionsInvariantCollectionMutableState",
	ConcreteExecutionFixerDomainAllow:                        "worker.concreteExecutionFixerDomainAllow",
	CurrentExecutionFixerDomainAllow:                         "worker.currentExecutionFixerDomainAllow",
	ConcreteExecutionFixerEnabled:                            "worker.concreteExecutionFixerEnabled",
	CurrentExecutionFixerEnabled:                             "worker.currentExecutionFixerEnabled",
	TimersScannerEnabled:                                     "worker.timersScannerEnabled",
	TimersFixerEnabled:                                       "worker.timersFixerEnabled",
	TimersScannerConcurrency:                                 "worker.timersScannerConcurrency",
	TimersScannerPersistencePageSize:                         "worker.timersScannerPersistencePageSize",
	TimersScannerBlobstoreFlushThreshold:                     "worker.timersScannerConcurrency",
	TimersScannerActivityBatchSize:                           "worker.timersScannerBlobstoreFlushThreshold",
	TimersScannerPeriodStart:                                 "worker.timersScannerPeriodStart",
	TimersScannerPeriodEnd:                                   "worker.timersScannerPeriodEnd",
	TimersFixerDomainAllow:                                   "worker.timersFixerDomainAllow",

	// used by internal repos, need to moved out of this repo
	// TODO https://github.com/uber/cadence/issues/3861
	EnableAuthorization:                             "system.enableAuthorization",
	VisibilityArchivalQueryMaxRangeInDays:           "frontend.visibilityArchivalQueryMaxRangeInDays",
	VisibilityArchivalQueryMaxQPS:                   "frontend.visibilityArchivalQueryMaxQPS",
	EnableArchivalCompression:                       "worker.EnableArchivalCompression",
	WorkerDeterministicConstructionCheckProbability: "worker.DeterministicConstructionCheckProbability",
	WorkerBlobIntegrityCheckProbability:             "worker.BlobIntegrityCheckProbability",
}
