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
	if k <= unknownKey || int(k) >= len(keys) {
		return keys[unknownKey]
	}
	return keys[k]
}

const (
	_matchingRoot               = "matching."
	_matchingDomainTaskListRoot = _matchingRoot + "domain." + "taskList."
	_historyRoot                = "history."
	_frontendRoot               = "frontend."
)

// These are keys name dynamic config source use.
var keys = []string{
	"unknownKey",
	"testGetPropertyKey",
	"testGetIntPropertyKey",
	"testGetFloat64PropertyKey",
	"testGetDurationPropertyKey",
	"testGetBoolPropertyKey",
	_frontendRoot + "visibilityMaxPageSize",
	_frontendRoot + "historyMaxPageSize",
	_frontendRoot + "rps",
	_frontendRoot + "historyMgrNumConns",
	_matchingDomainTaskListRoot + "minTaskThrottlingBurstSize",
	_matchingDomainTaskListRoot + "maxTaskBatchSize",
	_matchingDomainTaskListRoot + "longPollExpirationInterval",
	_matchingDomainTaskListRoot + "enableSyncMatch",
	_matchingDomainTaskListRoot + "updateAckInterval",
	_matchingDomainTaskListRoot + "idleTasklistCheckInterval",
	_historyRoot + "longPollExpirationInterval",
	_historyRoot + "cacheInitialSize",
	_historyRoot + "cacheMaxSize",
	_historyRoot + "cacheTTL",
	_historyRoot + "acquireShardInterval",
	_historyRoot + "timerTaskBatchSize",
	_historyRoot + "timerTaskWorkerCount",
	_historyRoot + "timerTaskMaxRetryCount",
	_historyRoot + "timerProcessorGetFailureRetryCount",
	_historyRoot + "timerProcessorCompleteTimerFailureRetryCount",
	_historyRoot + "timerProcessorUpdateShardTaskCount",
	_historyRoot + "timerProcessorUpdateAckInterval",
	_historyRoot + "timerProcessorCompleteTimerInterval",
	_historyRoot + "timerProcessorMaxPollInterval",
	_historyRoot + "timerProcessorStandbyTaskDelay",
	_historyRoot + "transferTaskBatchSize",
	_historyRoot + "transferProcessorMaxPollRPS",
	_historyRoot + "transferTaskWorkerCount",
	_historyRoot + "transferTaskMaxRetryCount",
	_historyRoot + "transferProcessorCompleteTransferFailureRetryCount",
	_historyRoot + "transferProcessorUpdateShardTaskCount",
	_historyRoot + "transferProcessorMaxPollInterval",
	_historyRoot + "transferProcessorUpdateAckInterval",
	_historyRoot + "transferProcessorCompleteTransferInterval",
	_historyRoot + "transferProcessorStandbyTaskDelay",
	_historyRoot + "replicatorTaskBatchSize",
	_historyRoot + "replicatorTaskWorkerCount",
	_historyRoot + "replicatorTaskMaxRetryCount",
	_historyRoot + "replicatorProcessorMaxPollRPS",
	_historyRoot + "replicatorProcessorUpdateShardTaskCount",
	_historyRoot + "replicatorProcessorMaxPollInterval",
	_historyRoot + "replicatorProcessorUpdateAckInterval",
	_historyRoot + "executionMgrNumConns",
	_historyRoot + "historyMgrNumConns",
	_historyRoot + "maximumBufferedEventsBatch",
	_historyRoot + "shardUpdateMinInterval",
}

const (
	// The order of constants is important. It should match the order in the keys array above.
	unknownKey Key = iota

	// key for tests
	testGetPropertyKey
	testGetIntPropertyKey
	testGetFloat64PropertyKey
	testGetDurationPropertyKey
	testGetBoolPropertyKey

	// key for frontend

	// FrontendVisibilityMaxPageSize is default max size for ListWorkflowExecutions in one page
	FrontendVisibilityMaxPageSize
	// FrontendHistoryMaxPageSize is default max size for GetWorkflowExecutionHistory in one page
	FrontendHistoryMaxPageSize
	// FrontendRPS is workflow rate limit per second
	FrontendRPS
	// FrontendHistoryMgrNumConns is for persistence cluster.NumConns
	FrontendHistoryMgrNumConns

	// key for matching

	// MatchingMinTaskThrottlingBurstSize is the minimum burst size for task list throttling
	MatchingMinTaskThrottlingBurstSize
	// MatchingMaxTaskBatchSize is the maximum batch size to fetch from the task buffer
	MatchingMaxTaskBatchSize
	// MatchingLongPollExpirationInterval is the long poll expiration interval in the matching service
	MatchingLongPollExpirationInterval
	// MatchingEnableSyncMatch is to enable sync match
	MatchingEnableSyncMatch
	// MatchingUpdateAckInterval is the interval for update ack
	MatchingUpdateAckInterval
	// MatchingIdleTasklistCheckInterval is the IdleTasklistCheckInterval
	MatchingIdleTasklistCheckInterval

	// key for history

	// HistoryLongPollExpirationInterval is the long poll expiration interval in the history service
	HistoryLongPollExpirationInterval
	// HistoryCacheInitialSize is initial size of history cache
	HistoryCacheInitialSize
	// HistoryCacheMaxSize is max size of history cache
	HistoryCacheMaxSize
	// HistoryCacheTTL is TTL of history cache
	HistoryCacheTTL
	// AcquireShardInterval is interval that timer used to acquire shard
	AcquireShardInterval
	// TimerTaskBatchSize is batch size for timer processor to process tasks
	TimerTaskBatchSize
	// TimerTaskWorkerCount is number of task workers for timer processor
	TimerTaskWorkerCount
	// TimerTaskMaxRetryCount is max retry count for timer processor
	TimerTaskMaxRetryCount
	// TimerProcessorGetFailureRetryCount is retry count for timer processor get failure operation
	TimerProcessorGetFailureRetryCount
	// TimerProcessorCompleteTimerFailureRetryCount is retry count for timer processor complete timer operation
	TimerProcessorCompleteTimerFailureRetryCount
	// TimerProcessorUpdateShardTaskCount is update shard count for timer processor
	TimerProcessorUpdateShardTaskCount
	// TimerProcessorUpdateAckInterval is update interval for timer processor
	TimerProcessorUpdateAckInterval
	// TimerProcessorCompleteTimerInterval is complete timer interval for timer processor
	TimerProcessorCompleteTimerInterval
	// TimerProcessorMaxPollInterval is max poll interval for timer processor
	TimerProcessorMaxPollInterval
	// TimerProcessorStandbyTaskDelay is task delay for standby task in timer processor
	TimerProcessorStandbyTaskDelay
	// TransferTaskBatchSize is batch size for transferQueueProcessor
	TransferTaskBatchSize
	// TransferProcessorMaxPollRPS is max poll rate per second for transferQueueProcessor
	TransferProcessorMaxPollRPS
	// TransferTaskWorkerCount is number of worker for transferQueueProcessor
	TransferTaskWorkerCount
	// TransferTaskMaxRetryCount is max times of retry for transferQueueProcessor
	TransferTaskMaxRetryCount
	// TransferProcessorCompleteTransferFailureRetryCount is times of retry for failure
	TransferProcessorCompleteTransferFailureRetryCount
	// TransferProcessorUpdateShardTaskCount is update shard count for transferQueueProcessor
	TransferProcessorUpdateShardTaskCount
	// TransferProcessorMaxPollInterval max poll interval for transferQueueProcessor
	TransferProcessorMaxPollInterval
	// TransferProcessorUpdateAckInterval is update interval for transferQueueProcessor
	TransferProcessorUpdateAckInterval
	// TransferProcessorCompleteTransferInterval is complete timer interval for transferQueueProcessor
	TransferProcessorCompleteTransferInterval
	// TransferProcessorStandbyTaskDelay is delay time for standby task in transferQueueProcessor
	TransferProcessorStandbyTaskDelay
	// ReplicatorTaskBatchSize is batch size for ReplicatorProcessor
	ReplicatorTaskBatchSize
	// ReplicatorTaskWorkerCount is number of worker for ReplicatorProcessor
	ReplicatorTaskWorkerCount
	// ReplicatorTaskMaxRetryCount is max times of retry for ReplicatorProcessor
	ReplicatorTaskMaxRetryCount
	// ReplicatorProcessorMaxPollRPS is max poll rate per second for ReplicatorProcessor
	ReplicatorProcessorMaxPollRPS
	// ReplicatorProcessorUpdateShardTaskCount is update shard count for ReplicatorProcessor
	ReplicatorProcessorUpdateShardTaskCount
	// ReplicatorProcessorMaxPollInterval is max poll interval for ReplicatorProcessor
	ReplicatorProcessorMaxPollInterval
	// ReplicatorProcessorUpdateAckInterval is update interval for ReplicatorProcessor
	ReplicatorProcessorUpdateAckInterval
	// ExecutionMgrNumConns is persistence connections number for ExecutionManager
	ExecutionMgrNumConns
	// HistoryMgrNumConns is persistence connections number for HistoryManager
	HistoryMgrNumConns
	// MaximumBufferedEventsBatch is max number of buffer event in mutable state
	MaximumBufferedEventsBatch
	// ShardUpdateMinInterval is the minimal time interval which the shard info can be updated
	ShardUpdateMinInterval
)

// Filter represents a filter on the dynamic config key
type Filter int

func (f Filter) String() string {
	if f <= unknownFilter || f > TaskListName {
		return filters[unknownFilter]
	}
	return filters[f]
}

var filters = []string{
	"unknownFilter",
	"domainName",
	"taskListName",
}

const (
	unknownFilter Filter = iota
	// DomainName is the domain name
	DomainName
	// TaskListName is the tasklist name
	TaskListName
)

// FilterOption is used to provide filters for dynamic config keys
type FilterOption func(filterMap map[Filter]interface{})

// TaskListFilter filters by task list name
func TaskListFilter(name string) FilterOption {
	return func(filterMap map[Filter]interface{}) {
		filterMap[TaskListName] = name
	}
}

// DomainFilter filters by domain name
func DomainFilter(name string) FilterOption {
	return func(filterMap map[Filter]interface{}) {
		filterMap[DomainName] = name
	}
}
