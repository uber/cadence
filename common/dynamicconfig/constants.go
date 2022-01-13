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
	keyName, ok := Keys[k]
	if !ok {
		return Keys[UnknownKey]
	}
	return keyName
}

/***
* !!!Important!!!
* For developer: Make sure to add/maintain the comment in the right format: usage, keyName, and default value
* So that our go-docs can have the full [documentation](https://pkg.go.dev/github.com/uber/cadence@v0.19.1/common/service/dynamicconfig#Key).
***/
const (
	UnknownKey Key = iota

	// key for tests
	TestGetPropertyKey
	TestGetIntPropertyKey
	TestGetFloat64PropertyKey
	TestGetDurationPropertyKey
	TestGetBoolPropertyKey
	TestGetStringPropertyKey
	TestGetMapPropertyKey
	TestGetIntPropertyFilteredByDomainKey
	TestGetDurationPropertyFilteredByDomainKey
	TestGetIntPropertyFilteredByTaskListInfoKey
	TestGetDurationPropertyFilteredByTaskListInfoKey
	TestGetBoolPropertyFilteredByDomainIDKey
	TestGetBoolPropertyFilteredByTaskListInfoKey

	// key for common & admin

	// EnableGlobalDomain is key for enable global domain
	// KeyName: system.enableGlobalDomain
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	EnableGlobalDomain
	// EnableVisibilitySampling is key for enable visibility sampling for basic(DB based) visibility
	// KeyName: system.enableVisibilitySampling
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	EnableVisibilitySampling
	// EnableReadFromClosedExecutionV2 is key for enable read from cadence_visibility.closed_executions_v2
	// KeyName: system.enableReadFromClosedExecutionV2
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	EnableReadFromClosedExecutionV2
	// AdvancedVisibilityWritingMode is key for how to write to advanced visibility. The most useful option is "dual", which can be used for seamless migration from db visibility to advanced visibility, usually using with EnableReadVisibilityFromES
	// KeyName: system.advancedVisibilityWritingMode
	// Value type: String enum: "on"(means writing to advancedVisibility only, "off" (means writing to db visibility only), or "dual" (means writing to both)
	// Default value: "on" if advanced visibility persistence is configured, otherwise "off" (see common.GetDefaultAdvancedVisibilityWritingMode(isAdvancedVisConfigExist))
	// Allowed filters: N/A
	AdvancedVisibilityWritingMode
	// EnableReadVisibilityFromES is key for enable read from elastic search or db visibility, usually using with AdvancedVisibilityWritingMode for seamless migration from db visibility to advanced visibility
	// KeyName: system.enableReadVisibilityFromES
	// Value type: Bool
	// Default value: true if advanced visibility persistence is configured, otherwise false
	// Allowed filters: DomainName
	EnableReadVisibilityFromES
	// EmitShardDiffLog is whether emit the shard diff log
	// KeyName: history.emitShardDiffLog
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	EmitShardDiffLog
	// DisableListVisibilityByFilter is config to disable list open/close workflow using filter
	// KeyName: frontend.disableListVisibilityByFilter
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainName
	DisableListVisibilityByFilter
	// HistoryArchivalStatus is key for the status of history archival to override the value from static config.
	// KeyName: system.historyArchivalStatus
	// Value type: string enum: "enabled" or "disabled"
	// Default value: the value in static config: common.Config.Archival.History.Status
	// Allowed filters: N/A
	HistoryArchivalStatus
	// EnableReadFromHistoryArchival is key for enabling reading history from archival store
	// KeyName: system.enableReadFromHistoryArchival
	// Value type: string enum: "enabled" or "disabled"
	// Default value: the value in static config: common.Config.Archival.History.EnableRead
	// Allowed filters: N/A
	EnableReadFromHistoryArchival
	// VisibilityArchivalStatus is key for the status of visibility archival to override the value from static config.
	// KeyName: system.visibilityArchivalStatus
	// Value type: string enum: "enabled" or "disabled"
	// Default value: the value in static config: common.Config.Archival.Visibility.Status
	// Allowed filters: N/A
	VisibilityArchivalStatus
	// EnableReadFromVisibilityArchival is key for enabling reading visibility from archival store to override the value from static config.
	// KeyName: system.enableReadFromVisibilityArchival
	// Value type: string enum: "enabled" or "disabled"
	// Default value: the value in static config: common.Config.Archival.Visibility.EnableRead
	// Allowed filters: N/A
	EnableReadFromVisibilityArchival
	// EnableDomainNotActiveAutoForwarding decides requests form which domain will be forwarded to active cluster if domain is not active in current cluster.
	// Only when "selected-api-forwarding" or "all-domain-apis-forwarding" is the policy in ClusterRedirectionPolicy(in static config).
	// If the policy is "noop"(default) this flag is not doing anything.
	// KeyName: system.enableDomainNotActiveAutoForwarding
	// Value type: Bool
	// Default value: true (meaning all domains are allowed to use the policy specified in static config)
	// Allowed filters: DomainName
	EnableDomainNotActiveAutoForwarding
	// EnableGracefulFailover is whether enabling graceful failover
	// KeyName: system.enableGracefulFailover
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	EnableGracefulFailover
	// TransactionSizeLimit is the largest allowed transaction size to persistence
	// KeyName: system.transactionSizeLimit
	// Value type: Int
	// Default value: 14680064 (from common.DefaultTransactionSizeLimit : 14 * 1024 * 1024)
	// Allowed filters: N/A
	TransactionSizeLimit
	// PersistenceErrorInjectionRate is rate for injecting random error in persistence
	// KeyName: system.persistenceErrorInjectionRate
	// Value type: Float64
	// Default value: 0
	// Allowed filters: N/A
	PersistenceErrorInjectionRate
	// MaxRetentionDays is the maximum allowed retention days for domain
	// KeyName: system.maxRetentionDays
	// Value type: Int
	// Default value: 30 (see domain.DefaultMaxWorkflowRetentionInDays)
	// Allowed filters: N/A
	MaxRetentionDays
	// MinRetentionDays is the minimal allowed retention days for domain
	// KeyName: system.minRetentionDays
	// Value type: Int
	// Default value: 1 (see domain.MinRetentionDays)
	// Allowed filters: N/A
	MinRetentionDays
	// MaxDecisionStartToCloseSeconds is the maximum allowed value for decision start to close timeout in seconds
	// KeyName: system.maxDecisionStartToCloseSeconds
	// Value type: Int
	// Default value: 240
	// Allowed filters: DomainName
	MaxDecisionStartToCloseSeconds
	// DisallowQuery is the key to disallow query for a domain
	// KeyName: system.disallowQuery
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainName
	DisallowQuery
	// EnableDebugMode is for enabling debugging components, logs and metrics
	// KeyName: system.enableDebugMode
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	EnableDebugMode
	// RequiredDomainDataKeys is the key for the list of data keys required in domain registration
	// KeyName: system.requiredDomainDataKeys
	// Value type: Map
	// Default value: nil
	// Allowed filters: N/A
	RequiredDomainDataKeys
	// EnableGRPCOutbound is the key for enabling outbound GRPC traffic
	// KeyName: system.enableGRPCOutbound
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	EnableGRPCOutbound
	// GRPCMaxSizeInByte is the key for config GRPC response size
	// KeyName: system.grpcMaxSizeInByte
	// Value type: Int
	// Default value: 4194304 (4*1024*1024)
	// Allowed filters: N/A
	GRPCMaxSizeInByte
	// BlobSizeLimitError is the per event blob size limit
	// KeyName: limit.blobSize.error
	// Value type: Int
	// Default value: 2097152 (2*1024*1024)
	// Allowed filters: DomainName
	BlobSizeLimitError
	// BlobSizeLimitWarn is the per event blob size limit for warning
	// KeyName: limit.blobSize.warn
	// Value type: Int
	// Default value: 262144 (256*1024)
	// Allowed filters: DomainName
	BlobSizeLimitWarn
	// HistorySizeLimitError is the per workflow execution history size limit
	// KeyName: limit.historySize.error
	// Value type: Int
	// Default value: 209715200 (200*1024*1024)
	// Allowed filters: DomainName
	HistorySizeLimitError
	// HistorySizeLimitWarn is the per workflow execution history size limit for warning
	// KeyName: limit.historySize.warn
	// Value type: Int
	// Default value: 52428800 (50*1024*1024)
	// Allowed filters: DomainName
	HistorySizeLimitWarn
	// HistoryCountLimitError is the per workflow execution history event count limit
	// KeyName: limit.historyCount.error
	// Value type: Int
	// Default value: 204800 (200*1024)
	// Allowed filters: DomainName
	HistoryCountLimitError
	// HistoryCountLimitWarn is the per workflow execution history event count limit for warning
	// KeyName: limit.historyCount.warn
	// Value type: Int
	// Default value: 51200 (50*1024)
	// Allowed filters: DomainName
	HistoryCountLimitWarn
	// DomainNameMaxLength is the length limit for domain name
	// KeyName: limit.domainNameLength
	// Value type: Int
	// Default value: common.DefaultIDLengthErrorLimit (1000)
	// Allowed filters: DomainName
	DomainNameMaxLength
	// IdentityMaxLength is the length limit for identity
	// KeyName: limit.identityLength
	// Value type: Int
	// Default value: 1000 ( see common.DefaultIDLengthErrorLimit)
	// Allowed filters: DomainName
	IdentityMaxLength
	// WorkflowIDMaxLength is the length limit for workflowID
	// KeyName: limit.workflowIDLength
	// Value type: Int
	// Default value: 1000 (see common.DefaultIDLengthErrorLimit)
	// Allowed filters: DomainName
	WorkflowIDMaxLength
	// SignalNameMaxLength is the length limit for signal name
	// KeyName: limit.signalNameLength
	// Value type: Int
	// Default value: 1000 (see common.DefaultIDLengthErrorLimit)
	// Allowed filters: DomainName
	SignalNameMaxLength
	// WorkflowTypeMaxLength is the length limit for workflow type
	// KeyName: limit.workflowTypeLength
	// Value type: Int
	// Default value: 1000 (see common.DefaultIDLengthErrorLimit)
	// Allowed filters: DomainName
	WorkflowTypeMaxLength
	// RequestIDMaxLength is the length limit for requestID
	// KeyName: limit.requestIDLength
	// Value type: Int
	// Default value: 1000 (see common.DefaultIDLengthErrorLimit)
	// Allowed filters: DomainName
	RequestIDMaxLength
	// TaskListNameMaxLength is the length limit for task list name
	// KeyName: limit.taskListNameLength
	// Value type: Int
	// Default value: 1000 (see common.DefaultIDLengthErrorLimit)
	// Allowed filters: DomainName
	TaskListNameMaxLength
	// ActivityIDMaxLength is the length limit for activityID
	// KeyName: limit.activityIDLength
	// Value type: Int
	// Default value: 1000 (see common.DefaultIDLengthErrorLimit)
	// Allowed filters: DomainName
	ActivityIDMaxLength
	// ActivityTypeMaxLength is the length limit for activity type
	// KeyName: limit.activityTypeLength
	// Value type: Int
	// Default value: 1000 (see common.DefaultIDLengthErrorLimit)
	// Allowed filters: DomainName
	ActivityTypeMaxLength
	// MarkerNameMaxLength is the length limit for marker name
	// KeyName: limit.markerNameLength
	// Value type: Int
	// Default value: 1000 (see common.DefaultIDLengthErrorLimit)
	// Allowed filters: DomainName
	MarkerNameMaxLength
	// TimerIDMaxLength is the length limit for timerID
	// KeyName: limit.timerIDLength
	// Value type: Int
	// Default value: 1000 (see common.DefaultIDLengthErrorLimit)
	// Allowed filters: DomainName
	TimerIDMaxLength
	// MaxIDLengthWarnLimit is the warn length limit for various IDs, including: Domain, TaskList, WorkflowID, ActivityID, TimerID, WorkflowType, ActivityType, SignalName, MarkerName, ErrorReason/FailureReason/CancelCause, Identity, RequestID
	// KeyName: limit.maxIDWarnLength
	// Value type: Int
	// Default value: 128 (see common.DefaultIDLengthWarnLimit)
	// Allowed filters: N/A
	MaxIDLengthWarnLimit
	// AdminErrorInjectionRate is the rate for injecting random error in admin client
	// KeyName: admin.errorInjectionRate
	// Value type: Float64
	// Default value: 0
	// Allowed filters: N/A
	AdminErrorInjectionRate

	// key for frontend

	// FrontendPersistenceMaxQPS is the max qps frontend host can query DB
	// KeyName: frontend.persistenceMaxQPS
	// Value type: Int
	// Default value: 2000
	// Allowed filters: N/A
	FrontendPersistenceMaxQPS
	// FrontendPersistenceGlobalMaxQPS is the max qps frontend cluster can query DB
	// KeyName: frontend.persistenceGlobalMaxQPS
	// Value type: Int
	// Default value: 0
	// Allowed filters: N/A
	FrontendPersistenceGlobalMaxQPS
	// FrontendVisibilityMaxPageSize is default max size for ListWorkflowExecutions in one page
	// KeyName: frontend.visibilityMaxPageSize
	// Value type: Int
	// Default value: 1000
	// Allowed filters: DomainName
	FrontendVisibilityMaxPageSize
	// FrontendVisibilityListMaxQPS is max qps frontend can list open/close workflows
	// KeyName: frontend.visibilityListMaxQPS
	// Value type: Int
	// Default value: 10
	// Allowed filters: DomainName
	FrontendVisibilityListMaxQPS
	// FrontendESVisibilityListMaxQPS is max qps frontend can list open/close workflows from ElasticSearch
	// KeyName: frontend.esVisibilityListMaxQPS
	// Value type: Int
	// Default value: 3
	// Allowed filters: DomainName
	FrontendESVisibilityListMaxQPS
	// FrontendESIndexMaxResultWindow is ElasticSearch index setting max_result_window
	// KeyName: frontend.esIndexMaxResultWindow
	// Value type: Int
	// Default value: 10000
	// Allowed filters: N/A
	FrontendESIndexMaxResultWindow
	// FrontendHistoryMaxPageSize is default max size for GetWorkflowExecutionHistory in one page
	// KeyName: frontend.historyMaxPageSize
	// Value type: Int
	// Default value: 1000 (see common.GetHistoryMaxPageSize)
	// Allowed filters: DomainName
	FrontendHistoryMaxPageSize
	// FrontendRPS is workflow rate limit per second
	// KeyName: frontend.rps
	// Value type: Int
	// Default value: 1200
	// Allowed filters: N/A
	FrontendRPS
	// FrontendMaxDomainRPSPerInstance is workflow domain rate limit per second
	// KeyName: frontend.domainrps
	// Value type: Int
	// Default value: 1200
	// Allowed filters: DomainName
	FrontendMaxDomainRPSPerInstance
	// FrontendGlobalDomainRPS is workflow domain rate limit per second for the whole Cadence cluster
	// KeyName: frontend.globalDomainrps
	// Value type: Int
	// Default value: 0
	// Allowed filters: DomainName
	FrontendGlobalDomainRPS
	// FrontendDecisionResultCountLimit is max number of decisions per RespondDecisionTaskCompleted request
	// KeyName: frontend.decisionResultCountLimit
	// Value type: Int
	// Default value: 0
	// Allowed filters: DomainName
	FrontendDecisionResultCountLimit
	// FrontendHistoryMgrNumConns is for persistence cluster.NumConns
	// KeyName: frontend.historyMgrNumConns
	// Value type: Int
	// Default value: 10
	// Allowed filters: N/A
	FrontendHistoryMgrNumConns
	// FrontendThrottledLogRPS is the rate limit on number of log messages emitted per second for throttled logger
	// KeyName: frontend.throttledLogRPS
	// Value type: Int
	// Default value: 20
	// Allowed filters: N/A
	FrontendThrottledLogRPS
	// FrontendShutdownDrainDuration is the duration of traffic drain during shutdown
	// KeyName: frontend.shutdownDrainDuration
	// Value type: Duration
	// Default value: 0
	// Allowed filters: N/A
	FrontendShutdownDrainDuration
	// EnableClientVersionCheck is enables client version check for frontend
	// KeyName: frontend.enableClientVersionCheck
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	EnableClientVersionCheck
	// FrontendMaxBadBinaries is the max number of bad binaries in domain config
	// KeyName: frontend.maxBadBinaries
	// Value type: Int
	// Default value: 10 (see domain.MaxBadBinaries)
	// Allowed filters: DomainName
	FrontendMaxBadBinaries
	// FrontendFailoverCoolDown is duration between two domain failvoers
	// KeyName: frontend.failoverCoolDown
	// Value type: Duration
	// Default value: 1m (one minute, see domain.FailoverCoolDown)
	// Allowed filters: DomainName
	FrontendFailoverCoolDown
	// ValidSearchAttributes is legal indexed keys that can be used in list APIs. When overriding, ensure to include the existing default attributes of the current release
	// KeyName: frontend.validSearchAttributes
	// Value type: Map
	// Default value: the default attributes of this release version, see definition.GetDefaultIndexedKeys()
	// Allowed filters: N/A
	ValidSearchAttributes
	// SendRawWorkflowHistory is whether to enable raw history retrieving
	// KeyName: frontend.sendRawWorkflowHistory
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainName
	SendRawWorkflowHistory
	// SearchAttributesNumberOfKeysLimit is the limit of number of keys
	// KeyName: frontend.searchAttributesNumberOfKeysLimit
	// Value type: Int
	// Default value: 100
	// Allowed filters: DomainName
	SearchAttributesNumberOfKeysLimit
	// SearchAttributesSizeOfValueLimit is the size limit of each value
	// KeyName: frontend.searchAttributesSizeOfValueLimit
	// Value type: Int
	// Default value: 2048 (2*1024)
	// Allowed filters: DomainName
	SearchAttributesSizeOfValueLimit
	// SearchAttributesTotalSizeLimit is the size limit of the whole map
	// KeyName: frontend.searchAttributesTotalSizeLimit
	// Value type: Int
	// Default value: 40960 (40*1024)
	// Allowed filters: DomainName
	SearchAttributesTotalSizeLimit
	// VisibilityArchivalQueryMaxPageSize is the maximum page size for a visibility archival query
	// KeyName: frontend.visibilityArchivalQueryMaxPageSize
	// Value type: Int
	// Default value: 10000
	// Allowed filters: N/A
	VisibilityArchivalQueryMaxPageSize
	// DomainFailoverRefreshInterval is the domain failover refresh timer
	// KeyName: frontend.domainFailoverRefreshInterval
	// Value type: Duration
	// Default value: 10s (10*time.Second)
	// Allowed filters: N/A
	DomainFailoverRefreshInterval
	// DomainFailoverRefreshTimerJitterCoefficient is the jitter for domain failover refresh timer jitter
	// KeyName: frontend.domainFailoverRefreshTimerJitterCoefficient
	// Value type: Float64
	// Default value: 0.1
	// Allowed filters: N/A
	DomainFailoverRefreshTimerJitterCoefficient
	// FrontendErrorInjectionRate is rate for injecting random error in frontend client
	// KeyName: frontend.errorInjectionRate
	// Value type: Float64
	// Default value: 0
	// Allowed filters: N/A
	FrontendErrorInjectionRate
	// FrontendEmitSignalNameMetricsTag enables emitting signal name tag in metrics in frontend client
	// KeyName: frontend.emitSignalNameMetricsTag
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainName
	FrontendEmitSignalNameMetricsTag

	// key for matching

	// MatchingRPS is request rate per second for each matching host
	// KeyName: matching.rps
	// Value type: Int
	// Default value: 1200
	// Allowed filters: N/A
	MatchingRPS
	// MatchingDomainRPS is request rate per domain per second for each matching host
	// KeyName: matching.domainrps
	// Value type: Int
	// Default value: 1200
	// Allowed filters: N/A
	MatchingDomainRPS
	// MatchingPersistenceMaxQPS is the max qps matching host can query DB
	// KeyName: matching.persistenceMaxQPS
	// Value type: Int
	// Default value: 3000
	// Allowed filters: N/A
	MatchingPersistenceMaxQPS
	// MatchingPersistenceGlobalMaxQPS is the max qps matching cluster can query DB
	// KeyName: matching.persistenceGlobalMaxQPS
	// Value type: Int
	// Default value: 0
	// Allowed filters: N/A
	MatchingPersistenceGlobalMaxQPS
	// MatchingMinTaskThrottlingBurstSize is the minimum burst size for task list throttling
	// KeyName: matching.minTaskThrottlingBurstSize
	// Value type: Int
	// Default value: 1
	// Allowed filters: DomainName,TasklistName,TasklistType
	MatchingMinTaskThrottlingBurstSize
	// MatchingGetTasksBatchSize is the maximum batch size to fetch from the task buffer
	// KeyName: matching.getTasksBatchSize
	// Value type: Int
	// Default value: 1000
	// Allowed filters: DomainName,TasklistName,TasklistType
	MatchingGetTasksBatchSize
	// MatchingLongPollExpirationInterval is the long poll expiration interval in the matching service
	// KeyName: matching.longPollExpirationInterval
	// Value type: Duration
	// Default value: time.Minute
	// Allowed filters: DomainName,TasklistName,TasklistType
	MatchingLongPollExpirationInterval
	// MatchingEnableSyncMatch is to enable sync match
	// KeyName: matching.enableSyncMatch
	// Value type: Bool
	// Default value: true
	// Allowed filters: DomainName,TasklistName,TasklistType
	MatchingEnableSyncMatch
	// MatchingUpdateAckInterval is the interval for update ack
	// KeyName: matching.updateAckInterval
	// Value type: Duration
	// Default value: 1m (1*time.Minute)
	// Allowed filters: DomainName,TasklistName,TasklistType
	MatchingUpdateAckInterval
	// MatchingIdleTasklistCheckInterval is the IdleTasklistCheckInterval
	// KeyName: matching.idleTasklistCheckInterval
	// Value type: Duration
	// Default value: 5m (5*time.Minute)
	// Allowed filters: DomainName,TasklistName,TasklistType
	MatchingIdleTasklistCheckInterval
	// MaxTasklistIdleTime is the max time tasklist being idle
	// KeyName: matching.maxTasklistIdleTime
	// Value type: Duration
	// Default value: 5m (5*time.Minute)
	// Allowed filters: DomainName,TasklistName,TasklistType
	MaxTasklistIdleTime
	// MatchingOutstandingTaskAppendsThreshold is the threshold for outstanding task appends
	// KeyName: matching.outstandingTaskAppendsThreshold
	// Value type: Int
	// Default value: 250
	// Allowed filters: DomainName,TasklistName,TasklistType
	MatchingOutstandingTaskAppendsThreshold
	// MatchingMaxTaskBatchSize is max batch size for task writer
	// KeyName: matching.maxTaskBatchSize
	// Value type: Int
	// Default value: 100
	// Allowed filters: DomainName,TasklistName,TasklistType
	MatchingMaxTaskBatchSize
	// MatchingMaxTaskDeleteBatchSize is the max batch size for range deletion of tasks
	// KeyName: matching.maxTaskDeleteBatchSize
	// Value type: Int
	// Default value: 100
	// Allowed filters: DomainName,TasklistName,TasklistType
	MatchingMaxTaskDeleteBatchSize
	// MatchingThrottledLogRPS is the rate limit on number of log messages emitted per second for throttled logger
	// KeyName: matching.throttledLogRPS
	// Value type: Int
	// Default value: 20
	// Allowed filters: N/A
	MatchingThrottledLogRPS
	// MatchingNumTasklistWritePartitions is the number of write partitions for a task list
	// KeyName: matching.numTasklistWritePartitions
	// Value type: Int
	// Default value: 1
	// Allowed filters: DomainName,TasklistName,TasklistType
	MatchingNumTasklistWritePartitions
	// MatchingNumTasklistReadPartitions is the number of read partitions for a task list
	// KeyName: matching.numTasklistReadPartitions
	// Value type: Int
	// Default value: 1
	// Allowed filters: DomainName,TasklistName,TasklistType
	MatchingNumTasklistReadPartitions
	// MatchingForwarderMaxOutstandingPolls is the max number of inflight polls from the forwarder
	// KeyName: matching.forwarderMaxOutstandingPolls
	// Value type: Int
	// Default value: 1
	// Allowed filters: DomainName,TasklistName,TasklistType
	MatchingForwarderMaxOutstandingPolls
	// MatchingForwarderMaxOutstandingTasks is the max number of inflight addTask/queryTask from the forwarder
	// KeyName: matching.forwarderMaxOutstandingTasks
	// Value type: Int
	// Default value: 1
	// Allowed filters: DomainName,TasklistName,TasklistType
	MatchingForwarderMaxOutstandingTasks
	// MatchingForwarderMaxRatePerSecond is the max rate at which add/query can be forwarded
	// KeyName: matching.forwarderMaxRatePerSecond
	// Value type: Int
	// Default value: 10
	// Allowed filters: DomainName,TasklistName,TasklistType
	MatchingForwarderMaxRatePerSecond
	// MatchingForwarderMaxChildrenPerNode is the max number of children per node in the task list partition tree
	// KeyName: matching.forwarderMaxChildrenPerNode
	// Value type: Int
	// Default value: 20
	// Allowed filters: DomainName,TasklistName,TasklistType
	MatchingForwarderMaxChildrenPerNode
	// MatchingShutdownDrainDuration is the duration of traffic drain during shutdown
	// KeyName: matching.shutdownDrainDuration
	// Value type: Duration
	// Default value: 0
	// Allowed filters: N/A
	MatchingShutdownDrainDuration
	// MatchingErrorInjectionRate is rate for injecting random error in matching client
	// KeyName: matching.errorInjectionRate
	// Value type: Float64
	// Default value: 0
	// Allowed filters: N/A
	MatchingErrorInjectionRate
	// MatchingEnableTaskInfoLogByDomainID is enables info level logs for decision/activity task based on the request domainID
	// KeyName: matching.enableTaskInfoLogByDomainID
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainID
	MatchingEnableTaskInfoLogByDomainID

	// key for history

	// HistoryRPS is request rate per second for each history host
	// KeyName: history.rps
	// Value type: Int
	// Default value: 3000
	// Allowed filters: N/A
	HistoryRPS
	// HistoryPersistenceMaxQPS is the max qps history host can query DB
	// KeyName: history.persistenceMaxQPS
	// Value type: Int
	// Default value: 9000
	// Allowed filters: N/A
	HistoryPersistenceMaxQPS
	// HistoryPersistenceGlobalMaxQPS is the max qps history cluster can query DB
	// KeyName: history.persistenceGlobalMaxQPS
	// Value type: Int
	// Default value: 0
	// Allowed filters: N/A
	HistoryPersistenceGlobalMaxQPS
	// HistoryVisibilityOpenMaxQPS is max qps one history host can write visibility open_executions
	// KeyName: history.historyVisibilityOpenMaxQPS
	// Value type: Int
	// Default value: 300
	// Allowed filters: DomainName
	HistoryVisibilityOpenMaxQPS
	// HistoryVisibilityClosedMaxQPS is max qps one history host can write visibility closed_executions
	// KeyName: history.historyVisibilityClosedMaxQPS
	// Value type: Int
	// Default value: 300
	// Allowed filters: DomainName
	HistoryVisibilityClosedMaxQPS
	// HistoryLongPollExpirationInterval is the long poll expiration interval in the history service
	// KeyName: history.longPollExpirationInterval
	// Value type: Duration
	// Default value: 20s( time.Second*20)
	// Allowed filters: DomainName
	HistoryLongPollExpirationInterval
	// HistoryCacheInitialSize is initial size of history cache
	// KeyName: history.cacheInitialSize
	// Value type: Int
	// Default value: 128
	// Allowed filters: N/A
	HistoryCacheInitialSize
	// HistoryCacheMaxSize is max size of history cache
	// KeyName: history.cacheMaxSize
	// Value type: Int
	// Default value: 512
	// Allowed filters: N/A
	HistoryCacheMaxSize
	// HistoryCacheTTL is TTL of history cache
	// KeyName: history.cacheTTL
	// Value type: Duration
	// Default value: 1h (time.Hour)
	// Allowed filters: N/A
	HistoryCacheTTL
	// HistoryShutdownDrainDuration is the duration of traffic drain during shutdown
	// KeyName: history.shutdownDrainDuration
	// Value type: Duration
	// Default value: 0
	// Allowed filters: N/A
	HistoryShutdownDrainDuration
	// EventsCacheInitialCount is initial count of events cache
	// KeyName: history.eventsCacheInitialSize
	// Value type: Int
	// Default value: 128
	// Allowed filters: N/A
	EventsCacheInitialCount
	// EventsCacheMaxCount is max count of events cache
	// KeyName: history.eventsCacheMaxSize
	// Value type: Int
	// Default value: 512
	// Allowed filters: N/A
	EventsCacheMaxCount
	// EventsCacheMaxSize is max size of events cache in bytes
	// KeyName: history.eventsCacheMaxSizeInBytes
	// Value type: Int
	// Default value: 0
	// Allowed filters: N/A
	EventsCacheMaxSize
	// EventsCacheTTL is TTL of events cache
	// KeyName: history.eventsCacheTTL
	// Value type: Duration
	// Default value: 1h (time.Hour)
	// Allowed filters: N/A
	EventsCacheTTL
	// EventsCacheGlobalEnable is enables global cache over all history shards
	// KeyName: history.eventsCacheGlobalEnable
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	EventsCacheGlobalEnable
	// EventsCacheGlobalInitialCount is initial count of global events cache
	// KeyName: history.eventsCacheGlobalInitialSize
	// Value type: Int
	// Default value: 4096
	// Allowed filters: N/A
	EventsCacheGlobalInitialCount
	// EventsCacheGlobalMaxCount is max count of global events cache
	// KeyName: history.eventsCacheGlobalMaxSize
	// Value type: Int
	// Default value: 131072
	// Allowed filters: N/A
	EventsCacheGlobalMaxCount
	// AcquireShardInterval is interval that timer used to acquire shard
	// KeyName: history.acquireShardInterval
	// Value type: Duration
	// Default value: 1m (time.Minute)
	// Allowed filters: N/A
	AcquireShardInterval
	// AcquireShardConcurrency is number of goroutines that can be used to acquire shards in the shard controller.
	// KeyName: history.acquireShardConcurrency
	// Value type: Int
	// Default value: 1
	// Allowed filters: N/A
	AcquireShardConcurrency
	// StandbyClusterDelay is the artificial delay added to standby cluster's view of active cluster's time
	// KeyName: history.standbyClusterDelay
	// Value type: Duration
	// Default value: 5m (5*time.Minute)
	// Allowed filters: N/A
	StandbyClusterDelay
	// StandbyTaskMissingEventsResendDelay is the amount of time standby cluster's will wait (if events are missing)before calling remote for missing events
	// KeyName: history.standbyTaskMissingEventsResendDelay
	// Value type: Duration
	// Default value: 15m (15*time.Minute)
	// Allowed filters: N/A
	StandbyTaskMissingEventsResendDelay
	// StandbyTaskMissingEventsDiscardDelay is the amount of time standby cluster's will wait (if events are missing)before discarding the task
	// KeyName: history.standbyTaskMissingEventsDiscardDelay
	// Value type: Duration
	// Default value: 25m (25*time.Minute)
	// Allowed filters: N/A
	StandbyTaskMissingEventsDiscardDelay
	// TaskProcessRPS is the task processing rate per second for each domain
	// KeyName: history.taskProcessRPS
	// Value type: Int
	// Default value: 1000
	// Allowed filters: DomainName
	TaskProcessRPS
	// TaskSchedulerType is the task scheduler type for priority task processor
	// KeyName: history.taskSchedulerType
	// Value type: Int enum(1 for SchedulerTypeFIFO, 2 for SchedulerTypeWRR(weighted round robin scheduler implementation))
	// Default value: 2 (task.SchedulerTypeWRR)
	// Allowed filters: N/A
	TaskSchedulerType
	// TaskSchedulerWorkerCount is the number of workers per host in task scheduler
	// KeyName: history.taskSchedulerWorkerCount
	// Value type: Int
	// Default value: 200
	// Allowed filters: N/A
	TaskSchedulerWorkerCount
	// TaskSchedulerShardWorkerCount is the number of worker per shard in task scheduler
	// KeyName: history.taskSchedulerShardWorkerCount
	// Value type: Int
	// Default value: 0
	// Allowed filters: N/A
	TaskSchedulerShardWorkerCount
	// TaskSchedulerQueueSize is the size of task channel for host level task scheduler
	// KeyName: history.taskSchedulerQueueSize
	// Value type: Int
	// Default value: 10000
	// Allowed filters: N/A
	TaskSchedulerQueueSize
	// TaskSchedulerShardQueueSize is the size of task channel for shard level task scheduler
	// KeyName: history.taskSchedulerShardQueueSize
	// Value type: Int
	// Default value: 200
	// Allowed filters: N/A
	TaskSchedulerShardQueueSize
	// TaskSchedulerDispatcherCount is the number of task dispatcher in task scheduler (only applies to host level task scheduler)
	// KeyName: history.taskSchedulerDispatcherCount
	// Value type: Int
	// Default value: 1
	// Allowed filters: N/A
	TaskSchedulerDispatcherCount
	// TaskSchedulerRoundRobinWeights is the priority weight for weighted round robin task scheduler
	// KeyName: history.taskSchedulerRoundRobinWeight
	// Value type: Map
	// Default value: please see common.ConvertIntMapToDynamicConfigMapProperty(DefaultTaskPriorityWeight) in code base
	// Allowed filters: N/A
	TaskSchedulerRoundRobinWeights
	// TaskCriticalRetryCount is the critical retry count for background tasks
	// when task attempt exceeds this threshold:
	// - task attempt metrics and additional error logs will be emitted
	// - task priority will be lowered
	// KeyName: history.taskCriticalRetryCount
	// Value type: Int
	// Default value: 20
	// Allowed filters: N/A
	TaskCriticalRetryCount
	// ActiveTaskRedispatchInterval is the active task redispatch interval
	// KeyName: history.activeTaskRedispatchInterval
	// Value type: Duration
	// Default value: 5s (5*time.Second)
	// Allowed filters: N/A
	ActiveTaskRedispatchInterval
	// StandbyTaskRedispatchInterval is the standby task redispatch interval
	// KeyName: history.standbyTaskRedispatchInterval
	// Value type: Duration
	// Default value: 30s (30*time.Second)
	// Allowed filters: N/A
	StandbyTaskRedispatchInterval
	// TaskRedispatchIntervalJitterCoefficient is the task redispatch interval jitter coefficient
	// KeyName: history.taskRedispatchIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	TaskRedispatchIntervalJitterCoefficient
	// StandbyTaskReReplicationContextTimeout is the context timeout for standby task re-replication
	// KeyName: history.standbyTaskReReplicationContextTimeout
	// Value type: Duration
	// Default value: 3m (3*time.Minute)
	// Allowed filters: DomainID
	StandbyTaskReReplicationContextTimeout
	// ResurrectionCheckMinDelay is the minimal timer processing delay before scanning history to see
	// if there's a resurrected timer/activity
	// KeyName: history.resurrectionCheckMinDelay
	// Value type: Duration
	// Default value: 24h (24*time.Hour)
	// Allowed filters: DomainName
	ResurrectionCheckMinDelay
	// QueueProcessorEnableSplit is indicates whether processing queue split policy should be enabled
	// KeyName: history.queueProcessorEnableSplit
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	QueueProcessorEnableSplit
	// QueueProcessorSplitMaxLevel is the max processing queue level
	// KeyName: history.queueProcessorSplitMaxLevel
	// Value type: Int
	// Default value: 2 // 3 levels, start from 0
	// Allowed filters: N/A
	QueueProcessorSplitMaxLevel
	// QueueProcessorEnableRandomSplitByDomainID is indicates whether random queue split policy should be enabled for a domain
	// KeyName: history.queueProcessorEnableRandomSplitByDomainID
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainID
	QueueProcessorEnableRandomSplitByDomainID
	// QueueProcessorRandomSplitProbability is the probability for a domain to be split to a new processing queue
	// KeyName: history.queueProcessorRandomSplitProbability
	// Value type: Float64
	// Default value: 0.01
	// Allowed filters: N/A
	QueueProcessorRandomSplitProbability
	// QueueProcessorEnablePendingTaskSplitByDomainID is indicates whether pending task split policy should be enabled
	// KeyName: history.queueProcessorEnablePendingTaskSplitByDomainID
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainID
	QueueProcessorEnablePendingTaskSplitByDomainID
	// QueueProcessorPendingTaskSplitThreshold is the threshold for the number of pending tasks per domain
	// KeyName: history.queueProcessorPendingTaskSplitThreshold
	// Value type: Map
	// Default value: see common.ConvertIntMapToDynamicConfigMapProperty(DefaultPendingTaskSplitThreshold) in code base
	// Allowed filters: N/A
	QueueProcessorPendingTaskSplitThreshold
	// QueueProcessorEnableStuckTaskSplitByDomainID is indicates whether stuck task split policy should be enabled
	// KeyName: history.queueProcessorEnableStuckTaskSplitByDomainID
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainID
	QueueProcessorEnableStuckTaskSplitByDomainID
	// QueueProcessorStuckTaskSplitThreshold is the threshold for the number of attempts of a task
	// KeyName: history.queueProcessorStuckTaskSplitThreshold
	// Value type: Map
	// Default value: see common.ConvertIntMapToDynamicConfigMapProperty(DefaultStuckTaskSplitThreshold) in code base
	// Allowed filters: N/A
	QueueProcessorStuckTaskSplitThreshold
	// QueueProcessorSplitLookAheadDurationByDomainID is the look ahead duration when spliting a domain to a new processing queue
	// KeyName: history.queueProcessorSplitLookAheadDurationByDomainID
	// Value type: Duration
	// Default value: 20m (20*time.Minute)
	// Allowed filters: DomainID
	QueueProcessorSplitLookAheadDurationByDomainID
	// QueueProcessorPollBackoffInterval is the backoff duration when queue processor is throttled
	// KeyName: history.queueProcessorPollBackoffInterval
	// Value type: Duration
	// Default value: 5s (5*time.Second)
	// Allowed filters: N/A
	QueueProcessorPollBackoffInterval
	// QueueProcessorPollBackoffIntervalJitterCoefficient is backoff interval jitter coefficient
	// KeyName: history.queueProcessorPollBackoffIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	QueueProcessorPollBackoffIntervalJitterCoefficient
	// QueueProcessorEnablePersistQueueStates is indicates whether processing queue states should be persisted
	// KeyName: history.queueProcessorEnablePersistQueueStates
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	QueueProcessorEnablePersistQueueStates
	// QueueProcessorEnableLoadQueueStates is indicates whether processing queue states should be loaded
	// KeyName: history.queueProcessorEnableLoadQueueStates
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	QueueProcessorEnableLoadQueueStates

	// TimerTaskBatchSize is batch size for timer processor to process tasks
	// KeyName: history.timerTaskBatchSize
	// Value type: Int
	// Default value: 100
	// Allowed filters: N/A
	TimerTaskBatchSize
	// TimerTaskDeleteBatchSize is batch size for timer processor to delete timer tasks
	// KeyName: history.timerTaskDeleteBatchSize
	// Value type: Int
	// Default value: 4000
	// Allowed filters: N/A
	TimerTaskDeleteBatchSize
	// TimerProcessorGetFailureRetryCount is retry count for timer processor get failure operation
	// KeyName: history.timerProcessorGetFailureRetryCount
	// Value type: Int
	// Default value: 5
	// Allowed filters: N/A
	TimerProcessorGetFailureRetryCount
	// TimerProcessorCompleteTimerFailureRetryCount is retry count for timer processor complete timer operation
	// KeyName: history.timerProcessorCompleteTimerFailureRetryCount
	// Value type: Int
	// Default value: 10
	// Allowed filters: N/A
	TimerProcessorCompleteTimerFailureRetryCount
	// TimerProcessorUpdateAckInterval is update interval for timer processor
	// KeyName: history.timerProcessorUpdateAckInterval
	// Value type: Duration
	// Default value: 30s (30*time.Second)
	// Allowed filters: N/A
	TimerProcessorUpdateAckInterval
	// TimerProcessorUpdateAckIntervalJitterCoefficient is the update interval jitter coefficient
	// KeyName: history.timerProcessorUpdateAckIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	TimerProcessorUpdateAckIntervalJitterCoefficient
	// TimerProcessorCompleteTimerInterval is complete timer interval for timer processor
	// KeyName: history.timerProcessorCompleteTimerInterval
	// Value type: Duration
	// Default value: 60s (60*time.Second)
	// Allowed filters: N/A
	TimerProcessorCompleteTimerInterval
	// TimerProcessorFailoverMaxStartJitterInterval is the max jitter interval for starting timer
	// failover queue processing. The actual jitter interval used will be a random duration between
	// 0 and the max interval so that timer failover queue across different shards won't start at
	// the same time
	// KeyName: history.timerProcessorFailoverMaxStartJitterInterval
	// Value type: Duration
	// Default value: 0s (0*time.Second)
	// Allowed filters: N/A
	TimerProcessorFailoverMaxStartJitterInterval
	// TimerProcessorFailoverMaxPollRPS is max poll rate per second for timer processor
	// KeyName: history.timerProcessorFailoverMaxPollRPS
	// Value type: Int
	// Default value: 1
	// Allowed filters: N/A
	TimerProcessorFailoverMaxPollRPS
	// TimerProcessorMaxPollRPS is max poll rate per second for timer processor
	// KeyName: history.timerProcessorMaxPollRPS
	// Value type: Int
	// Default value: 20
	// Allowed filters: N/A
	TimerProcessorMaxPollRPS
	// TimerProcessorMaxPollInterval is max poll interval for timer processor
	// KeyName: history.timerProcessorMaxPollInterval
	// Value type: Duration
	// Default value: 5m (5*time.Minute)
	// Allowed filters: N/A
	TimerProcessorMaxPollInterval
	// TimerProcessorMaxPollIntervalJitterCoefficient is the max poll interval jitter coefficient
	// KeyName: history.timerProcessorMaxPollIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	TimerProcessorMaxPollIntervalJitterCoefficient
	// TimerProcessorSplitQueueInterval is the split processing queue interval for timer processor
	// KeyName: history.timerProcessorSplitQueueInterval
	// Value type: Duration
	// Default value: 1m (1*time.Minute)
	// Allowed filters: N/A
	TimerProcessorSplitQueueInterval
	// TimerProcessorSplitQueueIntervalJitterCoefficient is the split processing queue interval jitter coefficient
	// KeyName: history.timerProcessorSplitQueueIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	TimerProcessorSplitQueueIntervalJitterCoefficient
	// TimerProcessorMaxRedispatchQueueSize is the threshold of the number of tasks in the redispatch queue for timer processor
	// KeyName: history.timerProcessorMaxRedispatchQueueSize
	// Value type: Int
	// Default value: 10000
	// Allowed filters: N/A
	TimerProcessorMaxRedispatchQueueSize
	// TimerProcessorMaxTimeShift is the max shift timer processor can have
	// KeyName: history.timerProcessorMaxTimeShift
	// Value type: Duration
	// Default value: 1s (1*time.Second)
	// Allowed filters: N/A
	TimerProcessorMaxTimeShift
	// TimerProcessorHistoryArchivalSizeLimit is the max history size for inline archival
	// KeyName: history.timerProcessorHistoryArchivalSizeLimit
	// Value type: Int
	// Default value: 500*1024
	// Allowed filters: N/A
	TimerProcessorHistoryArchivalSizeLimit
	// TimerProcessorArchivalTimeLimit is the upper time limit for inline history archival
	// KeyName: history.timerProcessorArchivalTimeLimit
	// Value type: Duration
	// Default value: 1s (1*time.Second)
	// Allowed filters: N/A
	TimerProcessorArchivalTimeLimit

	// TransferTaskBatchSize is batch size for transferQueueProcessor
	// KeyName: history.transferTaskBatchSize
	// Value type: Int
	// Default value: 100
	// Allowed filters: N/A
	TransferTaskBatchSize
	// TransferTaskDeleteBatchSize is batch size for transferQueueProcessor to delete transfer tasks
	// KeyName: history.transferTaskDeleteBatchSize
	// Value type: Int
	// Default value: 4000
	// Allowed filters: N/A
	TransferTaskDeleteBatchSize
	// TransferProcessorFailoverMaxStartJitterInterval is the max jitter interval for starting transfer
	// failover queue processing. The actual jitter interval used will be a random duration between
	// 0 and the max interval so that timer failover queue across different shards won't start at
	// the same time
	// KeyName: history.transferProcessorFailoverMaxStartJitterInterval
	// Value type: Duration
	// Default value: 0s (0*time.Second)
	// Allowed filters: N/A
	TransferProcessorFailoverMaxStartJitterInterval
	// TransferProcessorFailoverMaxPollRPS is max poll rate per second for transferQueueProcessor
	// KeyName: history.transferProcessorFailoverMaxPollRPS
	// Value type: Int
	// Default value: 1
	// Allowed filters: N/A
	TransferProcessorFailoverMaxPollRPS
	// TransferProcessorMaxPollRPS is max poll rate per second for transferQueueProcessor
	// KeyName: history.transferProcessorMaxPollRPS
	// Value type: Int
	// Default value: 20
	// Allowed filters: N/A
	TransferProcessorMaxPollRPS
	// TransferProcessorCompleteTransferFailureRetryCount is times of retry for failure
	// KeyName: history.transferProcessorCompleteTransferFailureRetryCount
	// Value type: Int
	// Default value: 10
	// Allowed filters: N/A
	TransferProcessorCompleteTransferFailureRetryCount
	// TransferProcessorMaxPollInterval is max poll interval for transferQueueProcessor
	// KeyName: history.transferProcessorMaxPollInterval
	// Value type: Duration
	// Default value: 1m (1*time.Minute)
	// Allowed filters: N/A
	TransferProcessorMaxPollInterval
	// TransferProcessorMaxPollIntervalJitterCoefficient is the max poll interval jitter coefficient
	// KeyName: history.transferProcessorMaxPollIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	TransferProcessorMaxPollIntervalJitterCoefficient
	// TransferProcessorSplitQueueInterval is the split processing queue interval for transferQueueProcessor
	// KeyName: history.transferProcessorSplitQueueInterval
	// Value type: Duration
	// Default value: 1m (1*time.Minute)
	// Allowed filters: N/A
	TransferProcessorSplitQueueInterval
	// TransferProcessorSplitQueueIntervalJitterCoefficient is the split processing queue interval jitter coefficient
	// KeyName: history.transferProcessorSplitQueueIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	TransferProcessorSplitQueueIntervalJitterCoefficient
	// TransferProcessorUpdateAckInterval is update interval for transferQueueProcessor
	// KeyName: history.transferProcessorUpdateAckInterval
	// Value type: Duration
	// Default value: 30s (30*time.Second)
	// Allowed filters: N/A
	TransferProcessorUpdateAckInterval
	// TransferProcessorUpdateAckIntervalJitterCoefficient is the update interval jitter coefficient
	// KeyName: history.transferProcessorUpdateAckIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	TransferProcessorUpdateAckIntervalJitterCoefficient
	// TransferProcessorCompleteTransferInterval is complete timer interval for transferQueueProcessor
	// KeyName: history.transferProcessorCompleteTransferInterval
	// Value type: Duration
	// Default value: 60s (60*time.Second)
	// Allowed filters: N/A
	TransferProcessorCompleteTransferInterval
	// TransferProcessorMaxRedispatchQueueSize is the threshold of the number of tasks in the redispatch queue for transferQueueProcessor
	// KeyName: history.transferProcessorMaxRedispatchQueueSize
	// Value type: Int
	// Default value: 10000
	// Allowed filters: N/A
	TransferProcessorMaxRedispatchQueueSize
	// TransferProcessorEnableValidator is whether validator should be enabled for transferQueueProcessor
	// KeyName: history.transferProcessorEnableValidator
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	TransferProcessorEnableValidator
	// TransferProcessorValidationInterval is interval for performing transfer queue validation
	// KeyName: history.transferProcessorValidationInterval
	// Value type: Duration
	// Default value: 30s (30*time.Second)
	// Allowed filters: N/A
	TransferProcessorValidationInterval
	// TransferProcessorVisibilityArchivalTimeLimit is the upper time limit for archiving visibility records
	// KeyName: history.transferProcessorVisibilityArchivalTimeLimit
	// Value type: Duration
	// Default value: 200ms (200*time.Millisecond)
	// Allowed filters: N/A
	TransferProcessorVisibilityArchivalTimeLimit

	// CrossClusterTaskBatchSize is the batch size for loading cross cluster tasks from persistence in crossClusterQueueProcessor
	// KeyName: history.crossClusterTaskBatchSize
	// Value type: Int
	// Default value: 100
	// Allowed filters: N/A
	CrossClusterTaskBatchSize
	// CrossClusterTaskDeleteBatchSize is the batch size for deleting cross cluster tasks from persistence in crossClusterQueueProcessor
	// KeyName: history.crossClusterTaskDeleteBatchSize
	// Value type: Int
	// Default value: 4000
	// Allowed filters: N/A
	CrossClusterTaskDeleteBatchSize
	// CrossClusterTaskFetchBatchSize is batch size for dispatching cross cluster tasks to target cluster in crossClusterQueueProcessor
	// KeyName: history.crossClusterTaskFetchBatchSize
	// Value type: Int
	// Default value: 100
	// Allowed filters: ShardID
	CrossClusterTaskFetchBatchSize
	// CrossClusterSourceProcessorMaxPollRPS is max poll rate per second for crossClusterQueueProcessor
	// KeyName: history.crossClusterProcessorMaxPollRPS
	// Value type: Int
	// Default value: 20
	// Allowed filters: N/A
	CrossClusterSourceProcessorMaxPollRPS
	// CrossClusterSourceProcessorCompleteTaskFailureRetryCount is times of retry for failure
	// KeyName: history.crossClusterProcessorCompleteTaskFailureRetryCount
	// Value type: Int
	// Default value: 10
	// Allowed filters: N/A
	CrossClusterSourceProcessorCompleteTaskFailureRetryCount // TODO
	// CrossClusterSourceProcessorMaxPollInterval is max poll interval for crossClusterQueueProcessor
	// KeyName: history.crossClusterProcessorMaxPollInterval
	// Value type: Duration
	// Default value: 1m (1*time.Minute)
	// Allowed filters: N/A
	CrossClusterSourceProcessorMaxPollInterval
	// CrossClusterSourceProcessorMaxPollIntervalJitterCoefficient is the max poll interval jitter coefficient
	// KeyName: history.crossClusterProcessorMaxPollIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	CrossClusterSourceProcessorMaxPollIntervalJitterCoefficient
	// CrossClusterSourceProcessorUpdateAckInterval is update interval for crossClusterQueueProcessor
	// KeyName: history.crossClusterProcessorUpdateAckInterval
	// Value type: Duration
	// Default value: 30s (30*time.Second)
	// Allowed filters: N/A
	CrossClusterSourceProcessorUpdateAckInterval
	// CrossClusterSourceProcessorUpdateAckIntervalJitterCoefficient is the update interval jitter coefficient
	// KeyName: history.crossClusterProcessorUpdateAckIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	CrossClusterSourceProcessorUpdateAckIntervalJitterCoefficient
	// CrossClusterSourceProcessorMaxRedispatchQueueSize is the threshold of the number of tasks in the redispatch queue for crossClusterQueueProcessor
	// KeyName: history.crossClusterProcessorMaxRedispatchQueueSize
	// Value type: Int
	// Default value: 10000
	// Allowed filters: N/A
	CrossClusterSourceProcessorMaxRedispatchQueueSize
	// CrossClusterSourceProcessorMaxPendingTaskSize is the threshold of the number of ready for polling tasks in crossClusterQueueProcessor,
	// task loading will be stopped when the number is reached
	// KeyName: history.crossClusterSourceProcessorMaxPendingTaskSize
	// Value type: Int
	// Default value: 500
	// Allowed filters: N/A
	CrossClusterSourceProcessorMaxPendingTaskSize

	// CrossClusterTargetProcessorMaxPendingTasks is the max number of pending tasks in cross cluster task processor
	// note there's one cross cluster task processor per shard per source cluster
	// KeyName: history.crossClusterTargetProcessorMaxPendingTasks
	// Value type: Int
	// Default value: 200
	// Allowed filters: N/A
	CrossClusterTargetProcessorMaxPendingTasks
	// CrossClusterTargetProcessorMaxRetryCount is the max number of retries when executing a cross-cluster task in target cluster
	// KeyName: history.crossClusterTargetProcessorMaxRetryCount
	// Value type: Int
	// Default value: 20
	// Allowed filters: N/A
	CrossClusterTargetProcessorMaxRetryCount
	// CrossClusterTargetProcessorTaskWaitInterval is the duration for waiting a cross-cluster task response before responding to source
	// KeyName: history.crossClusterTargetProcessorTaskWaitInterval
	// Value type: Duration
	// Default value: 3s (3*time.Second)
	// Allowed filters: N/A
	CrossClusterTargetProcessorTaskWaitInterval
	// CrossClusterTargetProcessorServiceBusyBackoffInterval is the backoff duration for cross cluster task processor when getting
	// a service busy error when calling source cluster
	// KeyName: history.crossClusterTargetProcessorServiceBusyBackoffInterval
	// Value type: Duration
	// Default value: 5s (5*time.Second)
	// Allowed filters: N/A
	CrossClusterTargetProcessorServiceBusyBackoffInterval
	// CrossClusterTargetProcessorJitterCoefficient is the jitter coefficient used in cross cluster task processor
	// KeyName: history.crossClusterTargetProcessorJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	CrossClusterTargetProcessorJitterCoefficient

	// CrossClusterFetcherParallelism is the number of go routines each cross cluster fetcher use
	// note there's one cross cluster task fetcher per host per source cluster
	// KeyName: history.crossClusterFetcherParallelism
	// Value type: Int
	// Default value: 1
	// Allowed filters: N/A
	CrossClusterFetcherParallelism
	// CrossClusterFetcherAggregationInterval determines how frequently the fetch requests are sent
	// KeyName: history.crossClusterFetcherAggregationInterval
	// Value type: Duration
	// Default value: 2s (2*time.Second)
	// Allowed filters: N/A
	CrossClusterFetcherAggregationInterval
	// CrossClusterFetcherServiceBusyBackoffInterval is the backoff duration for cross cluster task fetcher when getting
	// a service busy error when calling source cluster
	// KeyName: history.crossClusterFetcherServiceBusyBackoffInterval
	// Value type: Duration
	// Default value: 5s (5*time.Second)
	// Allowed filters: N/A
	CrossClusterFetcherServiceBusyBackoffInterval
	// CrossClusterFetcherServiceBusyBackoffInterval is the backoff duration for cross cluster task fetcher when getting
	// a non-service busy error when calling source cluster
	// KeyName: history.crossClusterFetcherErrorBackoffInterval
	// Value type: Duration
	// Default value: 1s (time.Second)
	// Allowed filters: N/A
	CrossClusterFetcherErrorBackoffInterval
	// CrossClusterFetcherJitterCoefficient is the jitter coefficient used in cross cluster task fetcher
	// KeyName: history.crossClusterFetcherJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	CrossClusterFetcherJitterCoefficient

	// ReplicatorTaskBatchSize is batch size for ReplicatorProcessor
	// KeyName: history.replicatorTaskBatchSize
	// Value type: Int
	// Default value: 100
	// Allowed filters: N/A
	ReplicatorTaskBatchSize
	// ReplicatorTaskDeleteBatchSize is batch size for ReplicatorProcessor to delete replication tasks
	// KeyName: history.replicatorTaskDeleteBatchSize
	// Value type: Int
	// Default value: 4000
	// Allowed filters: N/A
	ReplicatorTaskDeleteBatchSize
	// ReplicatorTaskWorkerCount is number of worker for ReplicatorProcessor
	// KeyName: history.replicatorTaskWorkerCount
	// Value type: Int
	// Default value: 10
	// Allowed filters: N/A
	ReplicatorTaskWorkerCount
	// ReplicatorReadTaskMaxRetryCount is the number of read replication task retry time
	// KeyName: history.replicatorReadTaskMaxRetryCount
	// Value type: Int
	// Default value: 3
	// Allowed filters: N/A
	ReplicatorReadTaskMaxRetryCount
	// ReplicatorProcessorMaxPollRPS is max poll rate per second for ReplicatorProcessor
	// KeyName: history.replicatorProcessorMaxPollRPS
	// Value type: Int
	// Default value: 20
	// Allowed filters: N/A
	ReplicatorProcessorMaxPollRPS
	// ReplicatorProcessorMaxPollInterval is max poll interval for ReplicatorProcessor
	// KeyName: history.replicatorProcessorMaxPollInterval
	// Value type: Duration
	// Default value: 1m (1*time.Minute)
	// Allowed filters: N/A
	ReplicatorProcessorMaxPollInterval
	// ReplicatorProcessorMaxPollIntervalJitterCoefficient is the max poll interval jitter coefficient
	// KeyName: history.replicatorProcessorMaxPollIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	ReplicatorProcessorMaxPollIntervalJitterCoefficient
	// ReplicatorProcessorUpdateAckInterval is update interval for ReplicatorProcessor
	// KeyName: history.replicatorProcessorUpdateAckInterval
	// Value type: Duration
	// Default value: 5s (5*time.Second)
	// Allowed filters: N/A
	ReplicatorProcessorUpdateAckInterval
	// ReplicatorProcessorUpdateAckIntervalJitterCoefficient is the update interval jitter coefficient
	// KeyName: history.replicatorProcessorUpdateAckIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	ReplicatorProcessorUpdateAckIntervalJitterCoefficient
	// ReplicatorProcessorMaxRedispatchQueueSize is the threshold of the number of tasks in the redispatch queue for ReplicatorProcessor
	// KeyName: history.replicatorProcessorMaxRedispatchQueueSize
	// Value type: Int
	// Default value: 10000
	// Allowed filters: N/A
	ReplicatorProcessorMaxRedispatchQueueSize
	// ReplicatorProcessorEnablePriorityTaskProcessor is indicates whether priority task processor should be used for ReplicatorProcessor
	// KeyName: history.replicatorProcessorEnablePriorityTaskProcessor
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	ReplicatorProcessorEnablePriorityTaskProcessor
	// ReplicatorUpperLatency indicates the max allowed replication latency between clusters
	// KeyName: history.replicatorUpperLatency
	// Value type: Duration
	// Default value: 40s (40 * time.Second)
	// Allowed filters: N/A
	ReplicatorUpperLatency

	// ExecutionMgrNumConns is persistence connections number for ExecutionManager
	// KeyName: history.executionMgrNumConns
	// Value type: Int
	// Default value: 50
	// Allowed filters: N/A
	ExecutionMgrNumConns
	// HistoryMgrNumConns is persistence connections number for HistoryManager
	// KeyName: history.historyMgrNumConns
	// Value type: Int
	// Default value: 50
	// Allowed filters: N/A
	HistoryMgrNumConns
	// MaximumBufferedEventsBatch is max number of buffer event in mutable state
	// KeyName: history.maximumBufferedEventsBatch
	// Value type: Int
	// Default value: 100
	// Allowed filters: N/A
	MaximumBufferedEventsBatch
	// MaximumSignalsPerExecution is max number of signals supported by single execution
	// KeyName: history.maximumSignalsPerExecution
	// Value type: Int
	// Default value: 10000
	// Allowed filters: DomainName
	MaximumSignalsPerExecution
	// ShardUpdateMinInterval is the minimal time interval which the shard info can be updated
	// KeyName: history.shardUpdateMinInterval
	// Value type: Duration
	// Default value: 5m (5*time.Minute)
	// Allowed filters: N/A
	ShardUpdateMinInterval
	// ShardSyncMinInterval is the minimal time interval which the shard info should be sync to remote
	// KeyName: history.shardSyncMinInterval
	// Value type: Duration
	// Default value: 5m (5*time.Minute)
	// Allowed filters: N/A
	ShardSyncMinInterval
	// DefaultEventEncoding is the encoding type for history events
	// KeyName: history.defaultEventEncoding
	// Value type: String
	// Default value: string(common.EncodingTypeThriftRW)
	// Allowed filters: DomainName
	DefaultEventEncoding
	// NumArchiveSystemWorkflows is key for number of archive system workflows running in total
	// KeyName: history.numArchiveSystemWorkflows
	// Value type: Int
	// Default value: 1000
	// Allowed filters: N/A
	NumArchiveSystemWorkflows
	// ArchiveRequestRPS is the rate limit on the number of archive request per second
	// KeyName: history.archiveRequestRPS
	// Value type: Int
	// Default value: 300 // should be much smaller than frontend RPS
	// Allowed filters: N/A
	ArchiveRequestRPS
	// EnableAdminProtection is whether to enable admin checking
	// KeyName: history.enableAdminProtection
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	EnableAdminProtection
	// AdminOperationToken is the token to pass admin checking
	// KeyName: history.adminOperationToken
	// Value type: String
	// Default value: common.DefaultAdminOperationToken
	// Allowed filters: N/A
	AdminOperationToken
	// HistoryMaxAutoResetPoints is the key for max number of auto reset points stored in mutableState
	// KeyName: history.historyMaxAutoResetPoints
	// Value type: Int
	// Default value: DefaultHistoryMaxAutoResetPoints
	// Allowed filters: DomainName
	HistoryMaxAutoResetPoints
	// EnableParentClosePolicy is whether to  ParentClosePolicy
	// KeyName: history.enableParentClosePolicy
	// Value type: Bool
	// Default value: true
	// Allowed filters: DomainName
	EnableParentClosePolicy
	// ParentClosePolicyThreshold is decides that parent close policy will be processed by sys workers(if enabled) ifthe number of children greater than or equal to this threshold
	// KeyName: history.parentClosePolicyThreshold
	// Value type: Int
	// Default value: 10
	// Allowed filters: DomainName
	ParentClosePolicyThreshold
	// NumParentClosePolicySystemWorkflows is key for number of parentClosePolicy system workflows running in total
	// KeyName: history.numParentClosePolicySystemWorkflows
	// Value type: Int
	// Default value: 10
	// Allowed filters: N/A
	NumParentClosePolicySystemWorkflows
	// HistoryThrottledLogRPS is the rate limit on number of log messages emitted per second for throttled logger
	// KeyName: history.throttledLogRPS
	// Value type: Int
	// Default value: 4
	// Allowed filters: N/A
	HistoryThrottledLogRPS
	// StickyTTL is to expire a sticky tasklist if no update more than this duration
	// KeyName: history.stickyTTL
	// Value type: Duration
	// Default value: time.Hour*24*365
	// Allowed filters: DomainName
	StickyTTL
	// DecisionHeartbeatTimeout is for decision heartbeat
	// KeyName: history.decisionHeartbeatTimeout
	// Value type: Duration
	// Default value: 30m (time.Minute*30)
	// Allowed filters: DomainName
	DecisionHeartbeatTimeout
	// DecisionRetryCriticalAttempts is decision attempt threshold for logging and emiting metrics
	// KeyName: history.decisionRetryCriticalAttempts
	// Value type: Int
	// Default value: 10
	// Allowed filters: N/A
	DecisionRetryCriticalAttempts
	// DecisionRetryMaxAttempts is the max limit for decision retry attempts. 0 indicates infinite number of attempts.
	// KeyName: history.decisionRetryMaxAttempts
	// Value type: Int
	// Default value: 1000
	// Allowed filters: DomainName
	DecisionRetryMaxAttempts
	// NormalDecisionScheduleToStartMaxAttempts is the maximum decision attempt for creating a scheduleToStart timeout
	// timer for normal (non-sticky) decision
	// KeyName: history.normalDecisionScheduleToStartMaxAttempts
	// Value type: Int
	// Default value: 0
	// Allowed filters: DomainName
	NormalDecisionScheduleToStartMaxAttempts
	// NormalDecisionScheduleToStartTimeout is scheduleToStart timeout duration for normal (non-sticky) decision task
	// KeyName: history.normalDecisionScheduleToStartTimeout
	// Value type: Duration
	// Default value: time.Minute*5
	// Allowed filters: DomainName
	NormalDecisionScheduleToStartTimeout
	// EnableDropStuckTaskByDomainID is whether stuck timer/transfer task should be dropped for a domain
	// KeyName: history.DropStuckTaskByDomain
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainID
	EnableDropStuckTaskByDomainID
	// EnableConsistentQuery indicates if consistent query is enabled for the cluster
	// KeyName: history.EnableConsistentQuery
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	EnableConsistentQuery
	// EnableConsistentQueryByDomain indicates if consistent query is enabled for a domain
	// KeyName: history.EnableConsistentQueryByDomain
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainName
	EnableConsistentQueryByDomain
	// EnableCrossClusterOperations indicates if cross cluster operations can be scheduled for a domain
	// KeyName: history.enableCrossClusterOperations
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainName
	EnableCrossClusterOperations
	// MaxBufferedQueryCount indicates the maximum number of queries which can be buffered at a given time for a single workflow
	// KeyName: history.MaxBufferedQueryCount
	// Value type: Int
	// Default value: 1
	// Allowed filters: N/A
	MaxBufferedQueryCount
	// MutableStateChecksumGenProbability is the probability [0-100] that checksum will be generated for mutable state
	// KeyName: history.mutableStateChecksumGenProbability
	// Value type: Int
	// Default value: 0
	// Allowed filters: DomainName
	MutableStateChecksumGenProbability
	// MutableStateChecksumVerifyProbability is the probability [0-100] that checksum will be verified for mutable state
	// KeyName: history.mutableStateChecksumVerifyProbability
	// Value type: Int
	// Default value: 0
	// Allowed filters: DomainName
	MutableStateChecksumVerifyProbability
	// MutableStateChecksumInvalidateBefore is the epoch timestamp before which all checksums are to be discarded
	// KeyName: history.mutableStateChecksumInvalidateBefore
	// Value type: Float64
	// Default value: 0
	// Allowed filters: N/A
	MutableStateChecksumInvalidateBefore
	// NotifyFailoverMarkerInterval is determines the frequency to notify failover marker
	// KeyName: history.NotifyFailoverMarkerInterval
	// Value type: Duration
	// Default value: 5s (5*time.Second)
	// Allowed filters: N/A
	NotifyFailoverMarkerInterval
	// NotifyFailoverMarkerTimerJitterCoefficient is the jitter for failover marker notifier timer
	// KeyName: history.NotifyFailoverMarkerTimerJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	NotifyFailoverMarkerTimerJitterCoefficient
	// EnableActivityLocalDispatchByDomain is allows worker to dispatch activity tasks through local tunnel after decisions are made. This is an performance optimization to skip activity scheduling efforts
	// KeyName: history.enableActivityLocalDispatchByDomain
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainName
	EnableActivityLocalDispatchByDomain
	// HistoryErrorInjectionRate is rate for injecting random error in history client
	// KeyName: history.errorInjectionRate
	// Value type: Float64
	// Default value: 0
	// Allowed filters: N/A
	HistoryErrorInjectionRate
	// HistoryEnableTaskInfoLogByDomainID is enables info level logs for decision/activity task based on the request domainID
	// KeyName: history.enableTaskInfoLogByDomainID
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainID
	HistoryEnableTaskInfoLogByDomainID
	// ActivityMaxScheduleToStartTimeoutForRetry is maximum value allowed when overwritting the schedule to start timeout for activities with retry policy
	// KeyName: history.activityMaxScheduleToStartTimeoutForRetry
	// Value type: Duration
	// Default value: 30m (30*time.Minute)
	// Allowed filters: DomainName
	ActivityMaxScheduleToStartTimeoutForRetry

	// key for history replication

	// ReplicationTaskFetcherParallelism determines how many go routines we spin up for fetching tasks
	// KeyName: history.ReplicationTaskFetcherParallelism
	// Value type: Int
	// Default value: 1
	// Allowed filters: N/A
	ReplicationTaskFetcherParallelism
	// ReplicationTaskFetcherAggregationInterval determines how frequently the fetch requests are sent
	// KeyName: history.ReplicationTaskFetcherAggregationInterval
	// Value type: Duration
	// Default value: 2s (2 * time.Second)
	// Allowed filters: N/A
	ReplicationTaskFetcherAggregationInterval
	// ReplicationTaskFetcherTimerJitterCoefficient is the jitter for fetcher timer
	// KeyName: history.ReplicationTaskFetcherTimerJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	ReplicationTaskFetcherTimerJitterCoefficient
	// ReplicationTaskFetcherErrorRetryWait is the wait time when fetcher encounters error
	// KeyName: history.ReplicationTaskFetcherErrorRetryWait
	// Value type: Duration
	// Default value: time.Second
	// Allowed filters: N/A
	ReplicationTaskFetcherErrorRetryWait
	// ReplicationTaskFetcherServiceBusyWait is the wait time when fetcher encounters service busy error
	// KeyName: history.ReplicationTaskFetcherServiceBusyWait
	// Value type: Duration
	// Default value: 60s (60 * time.Second)
	// Allowed filters: N/A
	ReplicationTaskFetcherServiceBusyWait
	// ReplicationTaskProcessorErrorRetryWait is the initial retry wait when we see errors in applying replication tasks
	// KeyName: history.ReplicationTaskProcessorErrorRetryWait
	// Value type: Duration
	// Default value: 50ms (50*time.Millisecond)
	// Allowed filters: ShardID
	ReplicationTaskProcessorErrorRetryWait
	// ReplicationTaskProcessorErrorRetryMaxAttempts is the max retry attempts for applying replication tasks
	// KeyName: history.ReplicationTaskProcessorErrorRetryMaxAttempts
	// Value type: Int
	// Default value: 10
	// Allowed filters: ShardID
	ReplicationTaskProcessorErrorRetryMaxAttempts
	// ReplicationTaskProcessorErrorSecondRetryWait is the initial retry wait for the second phase retry
	// KeyName: history.ReplicationTaskProcessorErrorSecondRetryWait
	// Value type: Duration
	// Default value: 5s (5* time.Second)
	// Allowed filters: ShardID
	ReplicationTaskProcessorErrorSecondRetryWait
	// ReplicationTaskProcessorErrorSecondRetryMaxWait is the max wait time for the second phase retry
	// KeyName: history.ReplicationTaskProcessorErrorSecondRetryMaxWait
	// Value type: Duration
	// Default value: 150s (30 * 5 * time.Second)
	// Allowed filters: ShardID
	ReplicationTaskProcessorErrorSecondRetryMaxWait
	// ReplicationTaskProcessorErrorSecondRetryExpiration is the expiration duration for the second phase retry
	// KeyName: history.ReplicationTaskProcessorErrorSecondRetryExpiration
	// Value type: Duration
	// Default value: 5m (5* time.Minute)
	// Allowed filters: ShardID
	ReplicationTaskProcessorErrorSecondRetryExpiration
	// ReplicationTaskProcessorNoTaskInitialWait is the wait time when not ask is returned
	// KeyName: history.ReplicationTaskProcessorNoTaskInitialWait
	// Value type: Duration
	// Default value: 2s (2* time.Second)
	// Allowed filters: ShardID
	ReplicationTaskProcessorNoTaskInitialWait
	// ReplicationTaskProcessorCleanupInterval determines how frequently the cleanup replication queue
	// KeyName: history.ReplicationTaskProcessorCleanupInterval
	// Value type: Duration
	// Default value: 1m (1* time.Minute)
	// Allowed filters: ShardID
	ReplicationTaskProcessorCleanupInterval
	// ReplicationTaskProcessorCleanupJitterCoefficient is the jitter for cleanup timer
	// KeyName: history.ReplicationTaskProcessorCleanupJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: ShardID
	ReplicationTaskProcessorCleanupJitterCoefficient
	// ReplicationTaskProcessorReadHistoryBatchSize is the batch size to read history events
	// KeyName: history.ReplicationTaskProcessorReadHistoryBatchSize
	// Value type: Int
	// Default value: 5
	// Allowed filters: N/A
	ReplicationTaskProcessorReadHistoryBatchSize
	// ReplicationTaskProcessorStartWait is the wait time before each task processing batch
	// KeyName: history.ReplicationTaskProcessorStartWait
	// Value type: Duration
	// Default value: 5s (5* time.Second)
	// Allowed filters: ShardID
	ReplicationTaskProcessorStartWait
	// ReplicationTaskProcessorStartWaitJitterCoefficient is the jitter for batch start wait timer
	// KeyName: history.ReplicationTaskProcessorStartWaitJitterCoefficient
	// Value type: Float64
	// Default value: 0.9
	// Allowed filters: ShardID
	ReplicationTaskProcessorStartWaitJitterCoefficient
	// ReplicationTaskProcessorHostQPS is the qps of task processing rate limiter on host level
	// KeyName: history.ReplicationTaskProcessorHostQPS
	// Value type: Float64
	// Default value: 1500
	// Allowed filters: N/A
	ReplicationTaskProcessorHostQPS
	// ReplicationTaskProcessorShardQPS is the qps of task processing rate limiter on shard level
	// KeyName: history.ReplicationTaskProcessorShardQPS
	// Value type: Float64
	// Default value: 5
	// Allowed filters: N/A
	ReplicationTaskProcessorShardQPS
	// ReplicationTaskGenerationQPS is the wait time between each replication task generation qps
	// KeyName: history.ReplicationTaskGenerationQPS
	// Value type: Float64
	// Default value: 100
	// Allowed filters: N/A
	ReplicationTaskGenerationQPS
	// EnableReplicationTaskGeneration is the flag to control replication generation
	// KeyName: history.enableReplicationTaskGeneration
	// Value type: Bool
	// Default value: true
	// Allowed filters: DomainID, WorkflowID
	EnableReplicationTaskGeneration

	// key for worker

	// WorkerPersistenceMaxQPS is the max qps worker host can query DB
	// KeyName: worker.persistenceMaxQPS
	// Value type: Int
	// Default value: 500
	// Allowed filters: N/A
	WorkerPersistenceMaxQPS
	// WorkerPersistenceGlobalMaxQPS is the max qps worker cluster can query DB
	// KeyName: worker.persistenceGlobalMaxQPS
	// Value type: Int
	// Default value: 0
	// Allowed filters: N/A
	WorkerPersistenceGlobalMaxQPS
	// WorkerReplicationTaskMaxRetryDuration is the max retry duration for any task
	// KeyName: worker.replicationTaskMaxRetryDuration
	// Value type: Duration
	// Default value: #N/A
	// Allowed filters: N/A
	WorkerReplicationTaskMaxRetryDuration
	// WorkerIndexerConcurrency is the max concurrent messages to be processed at any given time
	// KeyName: worker.indexerConcurrency
	// Value type: Int
	// Default value: 1000
	// Allowed filters: N/A
	WorkerIndexerConcurrency
	// WorkerESProcessorNumOfWorkers is num of workers for esProcessor
	// KeyName: worker.ESProcessorNumOfWorkers
	// Value type: Int
	// Default value: 1
	// Allowed filters: N/A
	WorkerESProcessorNumOfWorkers
	// WorkerESProcessorBulkActions is max number of requests in bulk for esProcessor
	// KeyName: worker.ESProcessorBulkActions
	// Value type: Int
	// Default value: 1000
	// Allowed filters: N/A
	WorkerESProcessorBulkActions
	// WorkerESProcessorBulkSize is max total size of bulk in bytes for esProcessor
	// KeyName: worker.ESProcessorBulkSize
	// Value type: Int
	// Default value: 2<<24 // 16MB
	// Allowed filters: N/A
	WorkerESProcessorBulkSize
	// WorkerESProcessorFlushInterval is flush interval for esProcessor
	// KeyName: worker.ESProcessorFlushInterval
	// Value type: Duration
	// Default value: 1s (1*time.Second)
	// Allowed filters: N/A
	WorkerESProcessorFlushInterval
	// WorkerArchiverConcurrency is controls the number of coroutines handling archival work per archival workflow
	// KeyName: worker.ArchiverConcurrency
	// Value type: Int
	// Default value: 50
	// Allowed filters: N/A
	WorkerArchiverConcurrency
	// WorkerArchivalsPerIteration is controls the number of archivals handled in each iteration of archival workflow
	// KeyName: worker.ArchivalsPerIteration
	// Value type: Int
	// Default value: 1000
	// Allowed filters: N/A
	WorkerArchivalsPerIteration
	// WorkerTimeLimitPerArchivalIteration is controls the time limit of each iteration of archival workflow
	// KeyName: worker.TimeLimitPerArchivalIteration
	// Value type: Duration
	// Default value: archiver.MaxArchivalIterationTimeout()
	// Allowed filters: N/A
	WorkerTimeLimitPerArchivalIteration
	// AllowArchivingIncompleteHistory will continue on when seeing some error like history mutated(usually caused by database consistency issues)
	// KeyName: worker.AllowArchivingIncompleteHistory
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	AllowArchivingIncompleteHistory
	// WorkerThrottledLogRPS is the rate limit on number of log messages emitted per second for throttled logger
	// KeyName: worker.throttledLogRPS
	// Value type: Int
	// Default value: 20
	// Allowed filters: N/A
	WorkerThrottledLogRPS
	// ScannerPersistenceMaxQPS is the maximum rate of persistence calls from worker.Scanner
	// KeyName: worker.scannerPersistenceMaxQPS
	// Value type: Int
	// Default value: 100
	// Allowed filters: N/A
	ScannerPersistenceMaxQPS
	// ScannerGetOrphanTasksPageSize is the maximum number of orphans to delete in one batch
	// KeyName: worker.scannerGetOrphanTasksPageSize
	// Value type: Int
	// Default value: 1000
	// Allowed filters: N/A
	ScannerGetOrphanTasksPageSize
	// ScannerBatchSizeForTasklistHandler is for: 1. max number of tasks to query per call(get tasks for tasklist) in the scavenger handler. 2. The scavenger then uses the return to decide if a tasklist can be deleted. It's better to keep it a relatively high number to let it be more efficient.
	// KeyName: worker.scannerBatchSizeForTasklistHandler
	// Value type: Int
	// Default value: 16
	// Allowed filters: N/A
	ScannerBatchSizeForTasklistHandler
	// EnableCleaningOrphanTaskInTasklistScavenger indicates if enabling the scanner to clean up orphan tasks
	// Only implemented for single SQL database. TODO https://github.com/uber/cadence/issues/4064 for supporting multiple/sharded SQL database and NoSQL
	// KeyName: worker.enableCleaningOrphanTaskInTasklistScavenger
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	EnableCleaningOrphanTaskInTasklistScavenger
	// ScannerMaxTasksProcessedPerTasklistJob is the number of tasks to process for a tasklist in each workflow run
	// KeyName: worker.scannerMaxTasksProcessedPerTasklistJob
	// Value type: Int
	// Default value: 256
	// Allowed filters: N/A
	ScannerMaxTasksProcessedPerTasklistJob
	// TaskListScannerEnabled is indicates if task list scanner should be started as part of worker.Scanner
	// KeyName: worker.taskListScannerEnabled
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	TaskListScannerEnabled
	// HistoryScannerEnabled is indicates if history scanner should be started as part of worker.Scanner
	// KeyName: worker.historyScannerEnabled
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	HistoryScannerEnabled
	// ConcreteExecutionsScannerEnabled is indicates if executions scanner should be started as part of worker.Scanner
	// KeyName: worker.executionsScannerEnabled
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	ConcreteExecutionsScannerEnabled
	// ConcreteExecutionsScannerConcurrency is indicates the concurrency of concrete execution scanner
	// KeyName: worker.executionsScannerConcurrency
	// Value type: Int
	// Default value: 25
	// Allowed filters: N/A
	ConcreteExecutionsScannerConcurrency
	// ConcreteExecutionsScannerBlobstoreFlushThreshold is indicates the flush threshold of blobstore in concrete execution scanner
	// KeyName: worker.executionsScannerBlobstoreFlushThreshold
	// Value type: Int
	// Default value: 100
	// Allowed filters: N/A
	ConcreteExecutionsScannerBlobstoreFlushThreshold
	// ConcreteExecutionsScannerActivityBatchSize is indicates the batch size of scanner activities
	// KeyName: worker.executionsScannerActivityBatchSize
	// Value type: Int
	// Default value: 25
	// Allowed filters: N/A
	ConcreteExecutionsScannerActivityBatchSize
	// ConcreteExecutionsScannerPersistencePageSize is indicates the page size of execution persistence fetches in concrete execution scanner
	// KeyName: worker.executionsScannerPersistencePageSize
	// Value type: Int
	// Default value: 1000
	// Allowed filters: N/A
	ConcreteExecutionsScannerPersistencePageSize
	// ConcreteExecutionsScannerInvariantCollectionMutableState is indicates if mutable state invariant checks should be run
	// KeyName: worker.executionsScannerInvariantCollectionMutableState
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	ConcreteExecutionsScannerInvariantCollectionMutableState
	// ConcreteExecutionsScannerInvariantCollectionHistory is indicates if history invariant checks should be run
	// KeyName: worker.executionsScannerInvariantCollectionHistory
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	ConcreteExecutionsScannerInvariantCollectionHistory
	// CurrentExecutionsScannerEnabled is indicates if current executions scanner should be started as part of worker.Scanner
	// KeyName: worker.currentExecutionsScannerEnabled
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	CurrentExecutionsScannerEnabled
	// CurrentExecutionsScannerConcurrency is indicates the concurrency of current executions scanner
	// KeyName: worker.currentExecutionsConcurrency
	// Value type: Int
	// Default value: 25
	// Allowed filters: N/A
	CurrentExecutionsScannerConcurrency
	// CurrentExecutionsScannerBlobstoreFlushThreshold is indicates the flush threshold of blobstore in current executions scanner
	// KeyName: worker.currentExecutionsBlobstoreFlushThreshold
	// Value type: Int
	// Default value: 100
	// Allowed filters: N/A
	CurrentExecutionsScannerBlobstoreFlushThreshold
	// CurrentExecutionsScannerActivityBatchSize is indicates the batch size of scanner activities
	// KeyName: worker.currentExecutionsActivityBatchSize
	// Value type: Int
	// Default value: 25
	// Allowed filters: N/A
	CurrentExecutionsScannerActivityBatchSize
	// CurrentExecutionsScannerPersistencePageSize is indicates the page size of execution persistence fetches in current executions scanner
	// KeyName: worker.currentExecutionsPersistencePageSize
	// Value type: INt
	// Default value: 1000
	// Allowed filters: N/A
	CurrentExecutionsScannerPersistencePageSize
	// CurrentExecutionsScannerInvariantCollectionHistory is indicates if history invariant checks should be run
	// KeyName: worker.currentExecutionsScannerInvariantCollectionHistory
	// Value type: Int
	// Default value: false
	// Allowed filters: N/A
	CurrentExecutionsScannerInvariantCollectionHistory
	// CurrentExecutionsScannerInvariantCollectionMutableState is indicates if mutable state invariant checks should be run
	// KeyName: worker.currentExecutionsInvariantCollectionMutableState
	// Value type: Int
	// Default value: true
	// Allowed filters: N/A
	CurrentExecutionsScannerInvariantCollectionMutableState
	// EnableBatcher is decides whether start batcher in our worker
	// KeyName: worker.enableBatcher
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	EnableBatcher
	// EnableParentClosePolicyWorker decides whether or not enable system workers for processing parent close policy task
	// KeyName: system.enableParentClosePolicyWorker
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	EnableParentClosePolicyWorker
	// EnableESAnalyzer decides whether to enable system workers for processing ElasticSearch Analyzer
	// KeyName: system.enableESAnalyzer
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	EnableESAnalyzer
	// EnableStickyQuery is indicates if sticky query should be enabled per domain
	// KeyName: system.enableStickyQuery
	// Value type: Bool
	// Default value: true
	// Allowed filters: DomainName
	EnableStickyQuery
	// EnableFailoverManager is indicates if failover manager is enabled
	// KeyName: system.enableFailoverManager
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	EnableFailoverManager
	// EnableWorkflowShadower indicates if workflow shadower is enabled
	// KeyName: system.enableWorkflowShadower
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	EnableWorkflowShadower
	// ConcreteExecutionFixerDomainAllow is which domains are allowed to be fixed by concrete fixer workflow
	// KeyName: worker.concreteExecutionFixerDomainAllow
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainName
	ConcreteExecutionFixerDomainAllow
	// CurrentExecutionFixerDomainAllow is which domains are allowed to be fixed by current fixer workflow
	// KeyName: worker.currentExecutionFixerDomainAllow
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainName
	CurrentExecutionFixerDomainAllow
	// TimersScannerEnabled is if timers scanner should be started as part of worker.Scanner
	// KeyName: worker.timersScannerEnabled
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	TimersScannerEnabled
	// TimersFixerEnabled is if timers fixer should be started as part of worker.Scanner
	// KeyName: worker.timersFixerEnabled
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	TimersFixerEnabled
	// TimersScannerConcurrency is the concurrency of timers scanner
	// KeyName: worker.timersScannerConcurrency
	// Value type: Int
	// Default value: 5
	// Allowed filters: N/A
	TimersScannerConcurrency
	// TimersScannerPersistencePageSize is the page size of timers persistence fetches in timers scanner
	// KeyName: worker.timersScannerPersistencePageSize
	// Value type: Int
	// Default value: 1000
	// Allowed filters: N/A
	TimersScannerPersistencePageSize
	// TimersScannerBlobstoreFlushThreshold is threshold to flush blob store
	// KeyName: worker.timersScannerBlobstoreFlushThreshold
	// Value type: Int
	// Default value: 100
	// Allowed filters: N/A
	TimersScannerBlobstoreFlushThreshold
	// TimersScannerActivityBatchSize is TimersScannerActivityBatchSize
	// KeyName: worker.timersScannerActivityBatchSize
	// Value type: Int
	// Default value: 25
	// Allowed filters: N/A
	TimersScannerActivityBatchSize
	// TimersScannerPeriodStart is interval start for fetching scheduled timers
	// KeyName: worker.timersScannerPeriodStart
	// Value type: Int
	// Default value: 24
	// Allowed filters: N/A
	TimersScannerPeriodStart
	// TimersScannerPeriodEnd is interval end for fetching scheduled timers
	// KeyName: worker.timersScannerPeriodEnd
	// Value type: Int
	// Default value: 3
	// Allowed filters: N/A
	TimersScannerPeriodEnd
	// TimersFixerDomainAllow is which domains are allowed to be fixed by timer fixer workflow
	// KeyName: worker.timersFixerDomainAllow
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainName
	TimersFixerDomainAllow
	// ConcreteExecutionFixerEnabled is if concrete execution fixer workflow is enabled
	// KeyName: worker.concreteExecutionFixerEnabled
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	ConcreteExecutionFixerEnabled
	// CurrentExecutionFixerEnabled is if current execution fixer workflow is enabled
	// KeyName: worker.currentExecutionFixerEnabled
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	CurrentExecutionFixerEnabled

	// EnableAuthorization is the key to enable authorization for a domain, only for extension binary:
	// KeyName: N/A
	// Default value: N/A
	// TODO: https://github.com/uber/cadence/issues/3861
	EnableAuthorization
	// EnableServiceAuthorization is the key to enable authorization for a service, only for extension binary:
	// KeyName: N/A
	// Default value: N/A
	// TODO: https://github.com/uber/cadence/issues/3861
	EnableServiceAuthorization
	// EnableServiceAuthorizationLogOnly is the key to enable authorization logging for a service, only for extension binary:
	// KeyName: N/A
	// Default value: N/A
	// TODO: https://github.com/uber/cadence/issues/3861
	EnableServiceAuthorizationLogOnly
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

	// ESAnalyzerPause defines if we want to dynamically pause the analyzer workflow
	// KeyName: worker.ESAnalyzerPause
	// Value type: bool
	// Default value: false
	ESAnalyzerPause
	// ESAnalyzerTimeWindow defines the time window ElasticSearch Analyzer will consider while taking workflow averages
	// KeyName: worker.ESAnalyzerTimeWindow
	// Value type: Duration
	// Default value: 30 days
	ESAnalyzerTimeWindow
	// ESAnalyzerMaxNumDomains defines how many domains to check
	// KeyName: worker.ESAnalyzerMaxNumDomains
	// Value type: int
	// Default value: 500
	ESAnalyzerMaxNumDomains
	// ESAnalyzerMaxNumWorkflowTypes defines how many workflow types to check per domain
	// KeyName: worker.ESAnalyzerMaxNumWorkflowTypes
	// Value type: int
	// Default value: 100
	ESAnalyzerMaxNumWorkflowTypes
	// ESAnalyzerNumWorkflowsToRefresh controls how many workflows per workflow type should be refreshed per workflow type
	// KeyName: worker.ESAnalyzerNumWorkflowsToRefresh
	// Value type: Int
	// Default value: 100
	ESAnalyzerNumWorkflowsToRefresh
	// ESAnalyzerBufferWaitTime controls min time required to consider a worklow stuck
	// KeyName: worker.ESAnalyzerBufferWaitTime
	// Value type: Duration
	// Default value: 30 minutes
	ESAnalyzerBufferWaitTime
	// ESAnalyzerMinNumWorkflowsForAvg controls how many workflows to have at least to rely on workflow run time avg per type
	// KeyName: worker.ESAnalyzerMinNumWorkflowsForAvg
	// Value type: Int
	// Default value: 100
	ESAnalyzerMinNumWorkflowsForAvg
	// ESAnalyzerLimitToTypes controls if we want to limit ESAnalyzer only to some workflow types
	// KeyName: worker.ESAnalyzerLimitToTypes
	// Value type: Int
	// Default value: "" => means no limitation
	ESAnalyzerLimitToTypes
	// ESAnalyzerLimitToDomains controls if we want to limit ESAnalyzer only to some domains
	// KeyName: worker.ESAnalyzerLimitToDomains
	// Value type: Int
	// Default value: "" => means no limitation
	ESAnalyzerLimitToDomains
	// ESAnalyzerWorkflowDurationWarnThresholds defines the warning execution thresholds for workflow types
	// KeyName: worker.ESAnalyzerWorkflowDurationWarnThresholds
	// Value type: string (json of a dictionary {"<domainName>/<workflowType>":<value>,...})
	// Default value: ""
	ESAnalyzerWorkflowDurationWarnThresholds

	// LastKeyForTest must be the last one in this const group for testing purpose
	LastKeyForTest
)

// Keys maps Key to keyName, where keyName are used dynamic config source.
var Keys = map[Key]string{
	UnknownKey: "unknownKey",

	// tests keys
	TestGetPropertyKey:                               "testGetPropertyKey",
	TestGetIntPropertyKey:                            "testGetIntPropertyKey",
	TestGetFloat64PropertyKey:                        "testGetFloat64PropertyKey",
	TestGetDurationPropertyKey:                       "testGetDurationPropertyKey",
	TestGetBoolPropertyKey:                           "testGetBoolPropertyKey",
	TestGetStringPropertyKey:                         "testGetStringPropertyKey",
	TestGetMapPropertyKey:                            "testGetMapPropertyKey",
	TestGetIntPropertyFilteredByDomainKey:            "testGetIntPropertyFilteredByDomainKey",
	TestGetDurationPropertyFilteredByDomainKey:       "testGetDurationPropertyFilteredByDomainKey",
	TestGetIntPropertyFilteredByTaskListInfoKey:      "testGetIntPropertyFilteredByTaskListInfoKey",
	TestGetDurationPropertyFilteredByTaskListInfoKey: "testGetDurationPropertyFilteredByTaskListInfoKey",
	TestGetBoolPropertyFilteredByDomainIDKey:         "testGetBoolPropertyFilteredByDomainIDKey",
	TestGetBoolPropertyFilteredByTaskListInfoKey:     "testGetBoolPropertyFilteredByTaskListInfoKey",

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
	EnableESAnalyzer:                    "system.enableESAnalyzer",
	EnableFailoverManager:               "system.enableFailoverManager",
	EnableWorkflowShadower:              "system.enableWorkflowShadower",
	EnableStickyQuery:                   "system.enableStickyQuery",
	EnableDebugMode:                     "system.enableDebugMode",
	RequiredDomainDataKeys:              "system.requiredDomainDataKeys",
	EnableGRPCOutbound:                  "system.enableGRPCOutbound",
	GRPCMaxSizeInByte:                   "system.grpcMaxSizeInByte",

	// size limit
	BlobSizeLimitError:     "limit.blobSize.error",
	BlobSizeLimitWarn:      "limit.blobSize.warn",
	HistorySizeLimitError:  "limit.historySize.error",
	HistorySizeLimitWarn:   "limit.historySize.warn",
	HistoryCountLimitError: "limit.historyCount.error",
	HistoryCountLimitWarn:  "limit.historyCount.warn",

	// id length limits
	MaxIDLengthWarnLimit:  "limit.maxIDWarnLength",
	DomainNameMaxLength:   "limit.domainNameLength",
	IdentityMaxLength:     "limit.identityLength",
	WorkflowIDMaxLength:   "limit.workflowIDLength",
	SignalNameMaxLength:   "limit.signalNameLength",
	WorkflowTypeMaxLength: "limit.workflowTypeLength",
	RequestIDMaxLength:    "limit.requestIDLength",
	TaskListNameMaxLength: "limit.taskListNameLength",
	ActivityIDMaxLength:   "limit.activityIDLength",
	ActivityTypeMaxLength: "limit.activityTypeLength",
	MarkerNameMaxLength:   "limit.markerNameLength",
	TimerIDMaxLength:      "limit.timerIDLength",

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
	FrontendDecisionResultCountLimit:            "frontend.decisionResultCountLimit",
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
	FrontendEmitSignalNameMetricsTag:            "frontend.emitSignalNameMetricsTag",
	// matching settings
	MatchingRPS:                             "matching.rps",
	MatchingDomainRPS:                       "matching.domainrps",
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
	HistoryRPS:                                         "history.rps",
	HistoryPersistenceMaxQPS:                           "history.persistenceMaxQPS",
	HistoryPersistenceGlobalMaxQPS:                     "history.persistenceGlobalMaxQPS",
	HistoryVisibilityOpenMaxQPS:                        "history.historyVisibilityOpenMaxQPS",
	HistoryVisibilityClosedMaxQPS:                      "history.historyVisibilityClosedMaxQPS",
	HistoryLongPollExpirationInterval:                  "history.longPollExpirationInterval",
	HistoryCacheInitialSize:                            "history.cacheInitialSize",
	HistoryMaxAutoResetPoints:                          "history.historyMaxAutoResetPoints",
	HistoryCacheMaxSize:                                "history.cacheMaxSize",
	HistoryCacheTTL:                                    "history.cacheTTL",
	HistoryShutdownDrainDuration:                       "history.shutdownDrainDuration",
	EventsCacheInitialCount:                            "history.eventsCacheInitialSize",
	EventsCacheMaxCount:                                "history.eventsCacheMaxSize",
	EventsCacheMaxSize:                                 "history.eventsCacheMaxSizeInBytes",
	EventsCacheTTL:                                     "history.eventsCacheTTL",
	EventsCacheGlobalEnable:                            "history.eventsCacheGlobalEnable",
	EventsCacheGlobalInitialCount:                      "history.eventsCacheGlobalInitialSize",
	EventsCacheGlobalMaxCount:                          "history.eventsCacheGlobalMaxSize",
	AcquireShardInterval:                               "history.acquireShardInterval",
	AcquireShardConcurrency:                            "history.acquireShardConcurrency",
	StandbyClusterDelay:                                "history.standbyClusterDelay",
	StandbyTaskMissingEventsResendDelay:                "history.standbyTaskMissingEventsResendDelay",
	StandbyTaskMissingEventsDiscardDelay:               "history.standbyTaskMissingEventsDiscardDelay",
	TaskProcessRPS:                                     "history.taskProcessRPS",
	TaskSchedulerType:                                  "history.taskSchedulerType",
	TaskSchedulerWorkerCount:                           "history.taskSchedulerWorkerCount",
	TaskSchedulerShardWorkerCount:                      "history.taskSchedulerShardWorkerCount",
	TaskSchedulerQueueSize:                             "history.taskSchedulerQueueSize",
	TaskSchedulerShardQueueSize:                        "history.taskSchedulerShardQueueSize",
	TaskSchedulerDispatcherCount:                       "history.taskSchedulerDispatcherCount",
	TaskSchedulerRoundRobinWeights:                     "history.taskSchedulerRoundRobinWeight",
	TaskCriticalRetryCount:                             "history.taskCriticalRetryCount",
	ActiveTaskRedispatchInterval:                       "history.activeTaskRedispatchInterval",
	StandbyTaskRedispatchInterval:                      "history.standbyTaskRedispatchInterval",
	TaskRedispatchIntervalJitterCoefficient:            "history.taskRedispatchIntervalJitterCoefficient",
	StandbyTaskReReplicationContextTimeout:             "history.standbyTaskReReplicationContextTimeout",
	ResurrectionCheckMinDelay:                          "history.resurrectionCheckMinDelay",
	QueueProcessorEnableSplit:                          "history.queueProcessorEnableSplit",
	QueueProcessorSplitMaxLevel:                        "history.queueProcessorSplitMaxLevel",
	QueueProcessorEnableRandomSplitByDomainID:          "history.queueProcessorEnableRandomSplitByDomainID",
	QueueProcessorRandomSplitProbability:               "history.queueProcessorRandomSplitProbability",
	QueueProcessorEnablePendingTaskSplitByDomainID:     "history.queueProcessorEnablePendingTaskSplitByDomainID",
	QueueProcessorPendingTaskSplitThreshold:            "history.queueProcessorPendingTaskSplitThreshold",
	QueueProcessorEnableStuckTaskSplitByDomainID:       "history.queueProcessorEnableStuckTaskSplitByDomainID",
	QueueProcessorStuckTaskSplitThreshold:              "history.queueProcessorStuckTaskSplitThreshold",
	QueueProcessorSplitLookAheadDurationByDomainID:     "history.queueProcessorSplitLookAheadDurationByDomainID",
	QueueProcessorPollBackoffInterval:                  "history.queueProcessorPollBackoffInterval",
	QueueProcessorPollBackoffIntervalJitterCoefficient: "history.queueProcessorPollBackoffIntervalJitterCoefficient",
	QueueProcessorEnablePersistQueueStates:             "history.queueProcessorEnablePersistQueueStates",
	QueueProcessorEnableLoadQueueStates:                "history.queueProcessorEnableLoadQueueStates",

	TimerTaskBatchSize:                                "history.timerTaskBatchSize",
	TimerTaskDeleteBatchSize:                          "history.timerTaskDeleteBatchSize",
	TimerProcessorGetFailureRetryCount:                "history.timerProcessorGetFailureRetryCount",
	TimerProcessorCompleteTimerFailureRetryCount:      "history.timerProcessorCompleteTimerFailureRetryCount",
	TimerProcessorUpdateAckInterval:                   "history.timerProcessorUpdateAckInterval",
	TimerProcessorUpdateAckIntervalJitterCoefficient:  "history.timerProcessorUpdateAckIntervalJitterCoefficient",
	TimerProcessorCompleteTimerInterval:               "history.timerProcessorCompleteTimerInterval",
	TimerProcessorFailoverMaxStartJitterInterval:      "history.timerProcessorFailoverMaxStartJitterInterval",
	TimerProcessorFailoverMaxPollRPS:                  "history.timerProcessorFailoverMaxPollRPS",
	TimerProcessorMaxPollRPS:                          "history.timerProcessorMaxPollRPS",
	TimerProcessorMaxPollInterval:                     "history.timerProcessorMaxPollInterval",
	TimerProcessorMaxPollIntervalJitterCoefficient:    "history.timerProcessorMaxPollIntervalJitterCoefficient",
	TimerProcessorSplitQueueInterval:                  "history.timerProcessorSplitQueueInterval",
	TimerProcessorSplitQueueIntervalJitterCoefficient: "history.timerProcessorSplitQueueIntervalJitterCoefficient",
	TimerProcessorMaxRedispatchQueueSize:              "history.timerProcessorMaxRedispatchQueueSize",
	TimerProcessorMaxTimeShift:                        "history.timerProcessorMaxTimeShift",
	TimerProcessorHistoryArchivalSizeLimit:            "history.timerProcessorHistoryArchivalSizeLimit",
	TimerProcessorArchivalTimeLimit:                   "history.timerProcessorArchivalTimeLimit",

	TransferTaskBatchSize:                                "history.transferTaskBatchSize",
	TransferTaskDeleteBatchSize:                          "history.transferTaskDeleteBatchSize",
	TransferProcessorFailoverMaxStartJitterInterval:      "history.transferProcessorFailoverMaxStartJitterInterval",
	TransferProcessorFailoverMaxPollRPS:                  "history.transferProcessorFailoverMaxPollRPS",
	TransferProcessorMaxPollRPS:                          "history.transferProcessorMaxPollRPS",
	TransferProcessorCompleteTransferFailureRetryCount:   "history.transferProcessorCompleteTransferFailureRetryCount",
	TransferProcessorMaxPollInterval:                     "history.transferProcessorMaxPollInterval",
	TransferProcessorMaxPollIntervalJitterCoefficient:    "history.transferProcessorMaxPollIntervalJitterCoefficient",
	TransferProcessorSplitQueueInterval:                  "history.transferProcessorSplitQueueInterval",
	TransferProcessorSplitQueueIntervalJitterCoefficient: "history.transferProcessorSplitQueueIntervalJitterCoefficient",
	TransferProcessorUpdateAckInterval:                   "history.transferProcessorUpdateAckInterval",
	TransferProcessorUpdateAckIntervalJitterCoefficient:  "history.transferProcessorUpdateAckIntervalJitterCoefficient",
	TransferProcessorCompleteTransferInterval:            "history.transferProcessorCompleteTransferInterval",
	TransferProcessorMaxRedispatchQueueSize:              "history.transferProcessorMaxRedispatchQueueSize",
	TransferProcessorEnableValidator:                     "history.transferProcessorEnableValidator",
	TransferProcessorValidationInterval:                  "history.transferProcessorValidationInterval",
	TransferProcessorVisibilityArchivalTimeLimit:         "history.transferProcessorVisibilityArchivalTimeLimit",

	CrossClusterTaskBatchSize:                                     "history.crossClusterTaskBatchSize",
	CrossClusterTaskDeleteBatchSize:                               "history.crossClusterTaskDeleteBatchSize",
	CrossClusterTaskFetchBatchSize:                                "history.crossClusterTaskFetchBatchSize",
	CrossClusterSourceProcessorMaxPollRPS:                         "history.crossClusterSourceProcessorMaxPollRPS",
	CrossClusterSourceProcessorCompleteTaskFailureRetryCount:      "history.crossClusterSourceProcessorCompleteTaskFailureRetryCount",
	CrossClusterSourceProcessorMaxPollInterval:                    "history.crossClusterSourceProcessorMaxPollInterval",
	CrossClusterSourceProcessorMaxPollIntervalJitterCoefficient:   "history.crossClusterSourceProcessorMaxPollIntervalJitterCoefficient",
	CrossClusterSourceProcessorUpdateAckInterval:                  "history.crossClusterSourceProcessorUpdateAckInterval",
	CrossClusterSourceProcessorUpdateAckIntervalJitterCoefficient: "history.crossClusterSourceProcessorUpdateAckIntervalJitterCoefficient",
	CrossClusterSourceProcessorMaxRedispatchQueueSize:             "history.crossClusterSourceProcessorMaxRedispatchQueueSize",
	CrossClusterSourceProcessorMaxPendingTaskSize:                 "history.crossClusterSourceProcessorMaxPendingTaskSize",

	CrossClusterTargetProcessorMaxPendingTasks:            "history.crossClusterTargetProcessorMaxPendingTasks",
	CrossClusterTargetProcessorMaxRetryCount:              "history.crossClusterTargetProcessorMaxRetryCount",
	CrossClusterTargetProcessorTaskWaitInterval:           "history.crossClusterTargetProcessorTaskWaitInterval",
	CrossClusterTargetProcessorServiceBusyBackoffInterval: "history.crossClusterTargetProcessorServiceBusyBackoffInterval",
	CrossClusterTargetProcessorJitterCoefficient:          "history.crossClusterTargetProcessorJitterCoefficient",

	CrossClusterFetcherParallelism:                "history.crossClusterFetcherParallelism",
	CrossClusterFetcherAggregationInterval:        "history.crossClusterFetcherAggregationInterval",
	CrossClusterFetcherServiceBusyBackoffInterval: "history.crossClusterFetcherServiceBusyBackoffInterval",
	CrossClusterFetcherErrorBackoffInterval:       "history.crossClusterFetcherErrorBackoffInterval",
	CrossClusterFetcherJitterCoefficient:          "history.crossClusterFetcherJitterCoefficient",

	ReplicatorTaskBatchSize:                               "history.replicatorTaskBatchSize",
	ReplicatorTaskDeleteBatchSize:                         "history.replicatorTaskDeleteBatchSize",
	ReplicatorTaskWorkerCount:                             "history.replicatorTaskWorkerCount",
	ReplicatorReadTaskMaxRetryCount:                       "history.replicatorReadTaskMaxRetryCount",
	ReplicatorProcessorMaxPollRPS:                         "history.replicatorProcessorMaxPollRPS",
	ReplicatorProcessorMaxPollInterval:                    "history.replicatorProcessorMaxPollInterval",
	ReplicatorProcessorMaxPollIntervalJitterCoefficient:   "history.replicatorProcessorMaxPollIntervalJitterCoefficient",
	ReplicatorProcessorUpdateAckInterval:                  "history.replicatorProcessorUpdateAckInterval",
	ReplicatorProcessorUpdateAckIntervalJitterCoefficient: "history.replicatorProcessorUpdateAckIntervalJitterCoefficient",
	ReplicatorProcessorMaxRedispatchQueueSize:             "history.replicatorProcessorMaxRedispatchQueueSize",
	ReplicatorProcessorEnablePriorityTaskProcessor:        "history.replicatorProcessorEnablePriorityTaskProcessor",
	ReplicatorUpperLatency:                                "history.replicatorUpperLatency",

	ExecutionMgrNumConns:                               "history.executionMgrNumConns",
	HistoryMgrNumConns:                                 "history.historyMgrNumConns",
	MaximumBufferedEventsBatch:                         "history.maximumBufferedEventsBatch",
	MaximumSignalsPerExecution:                         "history.maximumSignalsPerExecution",
	ShardUpdateMinInterval:                             "history.shardUpdateMinInterval",
	ShardSyncMinInterval:                               "history.shardSyncMinInterval",
	DefaultEventEncoding:                               "history.defaultEventEncoding",
	EnableAdminProtection:                              "history.enableAdminProtection",
	AdminOperationToken:                                "history.adminOperationToken",
	EnableParentClosePolicy:                            "history.enableParentClosePolicy",
	NumArchiveSystemWorkflows:                          "history.numArchiveSystemWorkflows",
	ArchiveRequestRPS:                                  "history.archiveRequestRPS",
	EmitShardDiffLog:                                   "history.emitShardDiffLog",
	HistoryThrottledLogRPS:                             "history.throttledLogRPS",
	StickyTTL:                                          "history.stickyTTL",
	DecisionHeartbeatTimeout:                           "history.decisionHeartbeatTimeout",
	DecisionRetryCriticalAttempts:                      "history.decisionRetryCriticalAttempts",
	DecisionRetryMaxAttempts:                           "history.decisionRetryMaxAttempts",
	NormalDecisionScheduleToStartMaxAttempts:           "history.normalDecisionScheduleToStartMaxAttempts",
	NormalDecisionScheduleToStartTimeout:               "history.normalDecisionScheduleToStartTimeout",
	ParentClosePolicyThreshold:                         "history.parentClosePolicyThreshold",
	NumParentClosePolicySystemWorkflows:                "history.numParentClosePolicySystemWorkflows",
	ReplicationTaskFetcherParallelism:                  "history.ReplicationTaskFetcherParallelism",
	ReplicationTaskFetcherAggregationInterval:          "history.ReplicationTaskFetcherAggregationInterval",
	ReplicationTaskFetcherTimerJitterCoefficient:       "history.ReplicationTaskFetcherTimerJitterCoefficient",
	ReplicationTaskFetcherErrorRetryWait:               "history.ReplicationTaskFetcherErrorRetryWait",
	ReplicationTaskFetcherServiceBusyWait:              "history.ReplicationTaskFetcherServiceBusyWait",
	ReplicationTaskProcessorErrorRetryWait:             "history.ReplicationTaskProcessorErrorRetryWait",
	ReplicationTaskProcessorErrorRetryMaxAttempts:      "history.ReplicationTaskProcessorErrorRetryMaxAttempts",
	ReplicationTaskProcessorErrorSecondRetryWait:       "history.ReplicationTaskProcessorErrorSecondRetryWait",
	ReplicationTaskProcessorErrorSecondRetryMaxWait:    "history.ReplicationTaskProcessorErrorSecondRetryMaxWait",
	ReplicationTaskProcessorErrorSecondRetryExpiration: "history.ReplicationTaskProcessorErrorSecondRetryExpiration",
	ReplicationTaskProcessorNoTaskInitialWait:          "history.ReplicationTaskProcessorNoTaskInitialWait",
	ReplicationTaskProcessorCleanupInterval:            "history.ReplicationTaskProcessorCleanupInterval",
	ReplicationTaskProcessorCleanupJitterCoefficient:   "history.ReplicationTaskProcessorCleanupJitterCoefficient",
	ReplicationTaskProcessorReadHistoryBatchSize:       "history.ReplicationTaskProcessorReadHistoryBatchSize",
	ReplicationTaskProcessorStartWait:                  "history.ReplicationTaskProcessorStartWait",
	ReplicationTaskProcessorStartWaitJitterCoefficient: "history.ReplicationTaskProcessorStartWaitJitterCoefficient",
	ReplicationTaskProcessorHostQPS:                    "history.ReplicationTaskProcessorHostQPS",
	ReplicationTaskProcessorShardQPS:                   "history.ReplicationTaskProcessorShardQPS",
	EnableReplicationTaskGeneration:                    "history.enableReplicationTaskGeneration",
	ReplicationTaskGenerationQPS:                       "history.ReplicationTaskGenerationQPS",
	EnableConsistentQuery:                              "history.EnableConsistentQuery",
	EnableConsistentQueryByDomain:                      "history.EnableConsistentQueryByDomain",
	EnableCrossClusterOperations:                       "history.enableCrossClusterOperations",
	MaxBufferedQueryCount:                              "history.MaxBufferedQueryCount",
	MutableStateChecksumGenProbability:                 "history.mutableStateChecksumGenProbability",
	MutableStateChecksumVerifyProbability:              "history.mutableStateChecksumVerifyProbability",
	MutableStateChecksumInvalidateBefore:               "history.mutableStateChecksumInvalidateBefore",
	NotifyFailoverMarkerInterval:                       "history.NotifyFailoverMarkerInterval",
	NotifyFailoverMarkerTimerJitterCoefficient:         "history.NotifyFailoverMarkerTimerJitterCoefficient",
	EnableDropStuckTaskByDomainID:                      "history.DropStuckTaskByDomain",
	EnableActivityLocalDispatchByDomain:                "history.enableActivityLocalDispatchByDomain",
	HistoryErrorInjectionRate:                          "history.errorInjectionRate",
	HistoryEnableTaskInfoLogByDomainID:                 "history.enableTaskInfoLogByDomainID",
	ActivityMaxScheduleToStartTimeoutForRetry:          "history.activityMaxScheduleToStartTimeoutForRetry",

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
	AllowArchivingIncompleteHistory:                          "worker.AllowArchivingIncompleteHistory",
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
	EnableServiceAuthorization:                      "system.enableServiceAuthorization",
	EnableServiceAuthorizationLogOnly:               "system.enableServiceAuthorizationLogOnly",
	VisibilityArchivalQueryMaxRangeInDays:           "frontend.visibilityArchivalQueryMaxRangeInDays",
	VisibilityArchivalQueryMaxQPS:                   "frontend.visibilityArchivalQueryMaxQPS",
	EnableArchivalCompression:                       "worker.EnableArchivalCompression",
	WorkerDeterministicConstructionCheckProbability: "worker.DeterministicConstructionCheckProbability",
	WorkerBlobIntegrityCheckProbability:             "worker.BlobIntegrityCheckProbability",

	ESAnalyzerPause:                          "worker.ESAnalyzerPause",
	ESAnalyzerTimeWindow:                     "worker.ESAnalyzerTimeWindow",
	ESAnalyzerMaxNumDomains:                  "worker.ESAnalyzerMaxNumDomains",
	ESAnalyzerMaxNumWorkflowTypes:            "worker.ESAnalyzerMaxNumWorkflowTypes",
	ESAnalyzerNumWorkflowsToRefresh:          "worker.ESAnalyzerNumWorkflowsToRefresh",
	ESAnalyzerBufferWaitTime:                 "worker.ESAnalyzerBufferWaitTime",
	ESAnalyzerMinNumWorkflowsForAvg:          "worker.ESAnalyzerMinNumWorkflowsForAvg",
	ESAnalyzerLimitToTypes:                   "worker.ESAnalyzerLimitToTypes",
	ESAnalyzerLimitToDomains:                 "worker.ESAnalyzerLimitToDomains",
	ESAnalyzerWorkflowDurationWarnThresholds: "worker.ESAnalyzerWorkflowDurationWarnThresholds",
}

var KeyNames map[string]Key

func init() {
	KeyNames = make(map[string]Key)
	for k, v := range Keys {
		KeyNames[v] = k
	}
}
