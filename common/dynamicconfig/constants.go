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

import (
	"fmt"
	"math"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/definition"
)

type (
	// DynamicInt defines the properties for a dynamic config with int value type
	DynamicInt struct {
		KeyName      string
		Description  string
		DefaultValue int
		Filters      []Filter
	}

	DynamicBool struct {
		KeyName      string
		Description  string
		DefaultValue bool
		Filters      []Filter
	}

	DynamicFloat struct {
		KeyName      string
		Description  string
		DefaultValue float64
		Filters      []Filter
	}

	DynamicString struct {
		KeyName      string
		Description  string
		DefaultValue string
		Filters      []Filter
	}

	DynamicDuration struct {
		KeyName      string
		Description  string
		DefaultValue time.Duration
		Filters      []Filter
	}

	DynamicMap struct {
		KeyName      string
		Description  string
		DefaultValue map[string]interface{}
		Filters      []Filter
	}

	DynamicList struct {
		KeyName      string
		Description  string
		DefaultValue []interface{}
		Filters      []Filter
	}

	IntKey      int
	BoolKey     int
	FloatKey    int
	StringKey   int
	DurationKey int
	MapKey      int
	ListKey     int

	Key interface {
		String() string
		Description() string
		DefaultValue() interface{}
		// Filters is used to identify what filters a DynamicConfig key may have.
		// For example, CLI tool uses this to figure out all domain specific configurations for migration validation.
		Filters() []Filter
	}
)

// ListAllProductionKeys returns all key used in production
func ListAllProductionKeys() []Key {
	result := make([]Key, 0, len(IntKeys)+len(BoolKeys)+len(FloatKeys)+len(StringKeys)+len(DurationKeys)+len(MapKeys))
	for i := TestGetIntPropertyFilteredByTaskListInfoKey + 1; i < LastIntKey; i++ {
		result = append(result, i)
	}
	for i := TestGetBoolPropertyFilteredByTaskListInfoKey + 1; i < LastBoolKey; i++ {
		result = append(result, i)
	}
	for i := TestGetFloat64PropertyKey + 1; i < LastFloatKey; i++ {
		result = append(result, i)
	}
	for i := TestGetStringPropertyKey + 1; i < LastStringKey; i++ {
		result = append(result, i)
	}
	for i := TestGetDurationPropertyFilteredByTaskListInfoKey + 1; i < LastDurationKey; i++ {
		result = append(result, i)
	}
	for i := TestGetMapPropertyKey + 1; i < LastMapKey; i++ {
		result = append(result, i)
	}
	for i := TestGetListPropertyKey + 1; i < LastListKey; i++ {
		result = append(result, i)
	}
	return result
}

func GetKeyFromKeyName(keyName string) (Key, error) {
	keyVal, ok := _keyNames[keyName]
	if !ok {
		return nil, fmt.Errorf("invalid dynamic config key name: %s", keyName)
	}
	return keyVal, nil
}

// GetAllKeys returns a copy of all configuration keys with all details
func GetAllKeys() map[string]Key {
	result := make(map[string]Key, len(_keyNames))
	for k, v := range _keyNames {
		result[k] = v
	}
	return result
}

func ValidateKeyValuePair(key Key, value interface{}) error {
	err := fmt.Errorf("key value pair mismatch, key type: %T, value type: %T", key, value)
	switch key.(type) {
	case IntKey:
		if _, ok := value.(int); !ok {
			return err
		}
	case BoolKey:
		if _, ok := value.(bool); !ok {
			return err
		}
	case FloatKey:
		if _, ok := value.(float64); !ok {
			return err
		}
	case StringKey:
		if _, ok := value.(string); !ok {
			return err
		}
	case DurationKey:
		if _, ok := value.(time.Duration); !ok {
			return err
		}
	case MapKey:
		if _, ok := value.(map[string]interface{}); !ok {
			return err
		}
	case ListKey:
		if _, ok := value.([]interface{}); !ok {
			return err
		}
	default:
		return fmt.Errorf("unknown key type: %T", key)
	}
	return nil
}

func (k IntKey) String() string {
	return IntKeys[k].KeyName
}

func (k IntKey) Description() string {
	return IntKeys[k].Description
}

func (k IntKey) DefaultValue() interface{} {
	return IntKeys[k].DefaultValue
}

func (k IntKey) DefaultInt() int {
	return IntKeys[k].DefaultValue
}

func (k IntKey) Filters() []Filter {
	return IntKeys[k].Filters
}

func (k BoolKey) String() string {
	return BoolKeys[k].KeyName
}

func (k BoolKey) Description() string {
	return BoolKeys[k].Description
}

func (k BoolKey) DefaultValue() interface{} {
	return BoolKeys[k].DefaultValue
}

func (k BoolKey) DefaultBool() bool {
	return BoolKeys[k].DefaultValue
}

func (k BoolKey) Filters() []Filter {
	return BoolKeys[k].Filters
}

func (k FloatKey) String() string {
	return FloatKeys[k].KeyName
}

func (k FloatKey) Description() string {
	return FloatKeys[k].Description
}

func (k FloatKey) DefaultValue() interface{} {
	return FloatKeys[k].DefaultValue
}

func (k FloatKey) DefaultFloat() float64 {
	return FloatKeys[k].DefaultValue
}

func (k FloatKey) Filters() []Filter {
	return FloatKeys[k].Filters
}

func (k StringKey) String() string {
	return StringKeys[k].KeyName
}

func (k StringKey) Description() string {
	return StringKeys[k].Description
}

func (k StringKey) DefaultValue() interface{} {
	return StringKeys[k].DefaultValue
}

func (k StringKey) DefaultString() string {
	return StringKeys[k].DefaultValue
}

func (k StringKey) Filters() []Filter {
	return StringKeys[k].Filters
}

func (k DurationKey) String() string {
	return DurationKeys[k].KeyName
}

func (k DurationKey) Description() string {
	return DurationKeys[k].Description
}

func (k DurationKey) DefaultValue() interface{} {
	return DurationKeys[k].DefaultValue
}

func (k DurationKey) DefaultDuration() time.Duration {
	return DurationKeys[k].DefaultValue
}

func (k DurationKey) Filters() []Filter {
	return DurationKeys[k].Filters
}

func (k MapKey) String() string {
	return MapKeys[k].KeyName
}

func (k MapKey) Description() string {
	return MapKeys[k].Description
}

func (k MapKey) DefaultValue() interface{} {
	return MapKeys[k].DefaultValue
}

func (k MapKey) DefaultMap() map[string]interface{} {
	return MapKeys[k].DefaultValue
}

func (k MapKey) Filters() []Filter {
	return MapKeys[k].Filters
}

func (k ListKey) String() string {
	return ListKeys[k].KeyName
}

func (k ListKey) Description() string {
	return ListKeys[k].Description
}

func (k ListKey) DefaultValue() interface{} {
	return ListKeys[k].DefaultValue
}

func (k ListKey) DefaultList() []interface{} {
	return ListKeys[k].DefaultValue
}

func (k ListKey) Filters() []Filter {
	return ListKeys[k].Filters
}

// UnlimitedRPS represents an integer to use for "unlimited" RPS values.
//
// Since our ratelimiters do int/float conversions, and zero or negative values
// result in not allowing any requests, math.MaxInt is unsafe:
//
//	int(float64(math.MaxInt)) // -9223372036854775808
//
// Much higher values are possible, but we can't handle 2 billion RPS, this is good enough.
const UnlimitedRPS = math.MaxInt32

/***
* !!!Important!!!
* For developer: Make sure to add/maintain the comment in the right format: usage, keyName, and default value
* So that our go-docs can have the full [documentation](https://pkg.go.dev/github.com/uber/cadence@v0.19.1/common/service/dynamicconfig#Key).
***/
const (
	UnknownIntKey IntKey = iota

	// key for tests
	TestGetIntPropertyKey
	TestGetIntPropertyFilteredByDomainKey
	TestGetIntPropertyFilteredByTaskListInfoKey

	// key for common & admin

	TransactionSizeLimit
	MaxRetentionDays
	MinRetentionDays
	MaxDecisionStartToCloseSeconds
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
	// PendingActivitiesCountLimitError is the limit of how many pending activities a workflow can have at a point in time
	// KeyName: limit.pendingActivityCount.error
	// Value type: Int
	// Default value: 1024
	PendingActivitiesCountLimitError
	// PendingActivitiesCountLimitWarn is the limit of how many activities a workflow can have before a warning is logged
	// KeyName: limit.pendingActivityCount.warn
	// Value type: Int
	// Default value: 512
	PendingActivitiesCountLimitWarn
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
	// deprecated: never used for ratelimiting, only sampling-based failure injection, and only on database-based visibility
	FrontendVisibilityListMaxQPS
	// FrontendESVisibilityListMaxQPS is max qps frontend can list open/close workflows from ElasticSearch
	// KeyName: frontend.esVisibilityListMaxQPS
	// Value type: Int
	// Default value: 30
	// Allowed filters: DomainName
	// deprecated: never read from, all ES reads and writes erroneously use PersistenceMaxQPS
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
	// FrontendUserRPS is workflow rate limit per second
	// KeyName: frontend.rps
	// Value type: Int
	// Default value: 1200
	// Allowed filters: N/A
	FrontendUserRPS
	// FrontendWorkerRPS is background-processing workflow rate limit per second
	// KeyName: frontend.workerrps
	// Value type: Int
	// Default value: UnlimitedRPS
	// Allowed filters: N/A
	FrontendWorkerRPS
	// FrontendVisibilityRPS is the global workflow List*WorkflowExecutions request rate limit per second
	// KeyName: frontend.visibilityrps
	// Value type: Int
	// Default value: UnlimitedRPS
	// Allowed filters: N/A
	FrontendVisibilityRPS
	// FrontendMaxDomainUserRPSPerInstance is workflow domain rate limit per second
	// KeyName: frontend.domainrps
	// Value type: Int
	// Default value: 1200
	// Allowed filters: DomainName
	FrontendMaxDomainUserRPSPerInstance
	// FrontendMaxDomainWorkerRPSPerInstance is background-processing workflow domain rate limit per second
	// KeyName: frontend.domainworkerrps
	// Value type: Int
	// Default value: UnlimitedRPS
	// Allowed filters: DomainName
	FrontendMaxDomainWorkerRPSPerInstance
	// FrontendMaxDomainVisibilityRPSPerInstance is the per-instance List*WorkflowExecutions request rate limit per second
	// KeyName: frontend.domainvisibilityrps
	// Value type: Int
	// Default value: UnlimitedRPS
	// Allowed filters: DomainName
	FrontendMaxDomainVisibilityRPSPerInstance
	// FrontendGlobalDomainUserRPS is workflow domain rate limit per second for the whole Cadence cluster
	// KeyName: frontend.globalDomainrps
	// Value type: Int
	// Default value: 0
	// Allowed filters: DomainName
	FrontendGlobalDomainUserRPS
	// FrontendGlobalDomainWorkerRPS is background-processing workflow domain rate limit per second for the whole Cadence cluster
	// KeyName: frontend.globalDomainWorkerrps
	// Value type: Int
	// Default value: UnlimitedRPS
	// Allowed filters: DomainName
	FrontendGlobalDomainWorkerRPS
	// FrontendGlobalDomainVisibilityRPS is the per-domain List*WorkflowExecutions request rate limit per second
	// KeyName: frontend.globalDomainVisibilityrps
	// Value type: Int
	// Default value: UnlimitedRPS
	// Allowed filters: DomainName
	FrontendGlobalDomainVisibilityRPS
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
	// FrontendMaxBadBinaries is the max number of bad binaries in domain config
	// KeyName: frontend.maxBadBinaries
	// Value type: Int
	// Default value: 10 (see domain.MaxBadBinaries)
	// Allowed filters: DomainName
	FrontendMaxBadBinaries
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

	// key for matching

	// MatchingUserRPS is request rate per second for each matching host
	// KeyName: matching.rps
	// Value type: Int
	// Default value: 1200
	// Allowed filters: N/A
	MatchingUserRPS
	// MatchingWorkerRPS is background-processing request rate per second for each matching host
	// KeyName: matching.workerrps
	// Value type: Int
	// Default value: UnlimitedRPS
	// Allowed filters: N/A
	MatchingWorkerRPS
	// MatchingDomainUserRPS is request rate per domain per second for each matching host
	// KeyName: matching.domainrps
	// Value type: Int
	// Default value: 0
	// Allowed filters: N/A
	MatchingDomainUserRPS
	// MatchingDomainWorkerRPS is background-processing request rate per domain per second for each matching host
	// KeyName: matching.domainworkerrps
	// Value type: Int
	// Default value: UnlimitedRPS
	// Allowed filters: N/A
	MatchingDomainWorkerRPS
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
	// AcquireShardConcurrency is number of goroutines that can be used to acquire shards in the shard controller.
	// KeyName: history.acquireShardConcurrency
	// Value type: Int
	// Default value: 1
	// Allowed filters: N/A
	AcquireShardConcurrency
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
	// TaskCriticalRetryCount is the critical retry count for background tasks
	// when task attempt exceeds this threshold:
	// - task attempt metrics and additional error logs will be emitted
	// - task priority will be lowered
	// KeyName: history.taskCriticalRetryCount
	// Value type: Int
	// Default value: 50
	// Allowed filters: N/A
	TaskCriticalRetryCount
	// QueueProcessorSplitMaxLevel is the max processing queue level
	// KeyName: history.queueProcessorSplitMaxLevel
	// Value type: Int
	// Default value: 2 // 3 levels, start from 0
	// Allowed filters: N/A
	QueueProcessorSplitMaxLevel
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
	// TimerProcessorMaxRedispatchQueueSize is the threshold of the number of tasks in the redispatch queue for timer processor
	// KeyName: history.timerProcessorMaxRedispatchQueueSize
	// Value type: Int
	// Default value: 10000
	// Allowed filters: N/A
	TimerProcessorMaxRedispatchQueueSize
	// TimerProcessorHistoryArchivalSizeLimit is the max history size for inline archival
	// KeyName: history.timerProcessorHistoryArchivalSizeLimit
	// Value type: Int
	// Default value: 500*1024
	// Allowed filters: N/A
	TimerProcessorHistoryArchivalSizeLimit

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
	// TransferProcessorMaxRedispatchQueueSize is the threshold of the number of tasks in the redispatch queue for transferQueueProcessor
	// KeyName: history.transferProcessorMaxRedispatchQueueSize
	// Value type: Int
	// Default value: 10000
	// Allowed filters: N/A
	TransferProcessorMaxRedispatchQueueSize
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

	// CrossClusterFetcherParallelism is the number of go routines each cross cluster fetcher use
	// note there's one cross cluster task fetcher per host per source cluster
	// KeyName: history.crossClusterFetcherParallelism
	// Value type: Int
	// Default value: 1
	// Allowed filters: N/A
	CrossClusterFetcherParallelism

	// ReplicatorTaskBatchSize is batch size for ReplicatorProcessor
	// KeyName: history.replicatorTaskBatchSize
	// Value type: Int
	// Default value: 25
	// Allowed filters: N/A
	ReplicatorTaskBatchSize
	// ReplicatorTaskDeleteBatchSize is batch size for ReplicatorProcessor to delete replication tasks
	// KeyName: history.replicatorTaskDeleteBatchSize
	// Value type: Int
	// Default value: 4000
	// Allowed filters: N/A
	ReplicatorTaskDeleteBatchSize
	// ReplicatorReadTaskMaxRetryCount is the number of read replication task retry time
	// KeyName: history.replicatorReadTaskMaxRetryCount
	// Value type: Int
	// Default value: 3
	// Allowed filters: N/A
	ReplicatorReadTaskMaxRetryCount
	// ReplicatorCacheCapacity is the capacity of replication cache in number of tasks
	// KeyName: history.replicatorCacheCapacity
	// Value type: Int
	// Default value: 10000
	// Allowed filters: N/A
	ReplicatorCacheCapacity

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
	// ArchiveInlineHistoryRPS is the (per instance) rate limit on the number of inline history archival attempts per second
	// KeyName: history.archiveInlineHistoryRPS
	// Value type: Int
	// Default value: 1000
	// Allowed filters: N/A
	ArchiveInlineHistoryRPS
	// ArchiveInlineHistoryGlobalRPS is the global rate limit on the number of inline history archival attempts per second
	// KeyName: history.archiveInlineHistoryGlobalRPS
	// Value type: Int
	// Default value: 10000
	// Allowed filters: N/A
	ArchiveInlineHistoryGlobalRPS
	// ArchiveInlineVisibilityRPS is the (per instance) rate limit on the number of inline visibility archival attempts per second
	// KeyName: history.archiveInlineVisibilityRPS
	// Value type: Int
	// Default value: 1000
	// Allowed filters: N/A
	ArchiveInlineVisibilityRPS
	// ArchiveInlineVisibilityGlobalRPS is the global rate limit on the number of inline visibility archival attempts per second
	// KeyName: history.archiveInlineVisibilityGlobalRPS
	// Value type: Int
	// Default value: 10000
	// Allowed filters: N/A
	ArchiveInlineVisibilityGlobalRPS
	// HistoryMaxAutoResetPoints is the key for max number of auto reset points stored in mutableState
	// KeyName: history.historyMaxAutoResetPoints
	// Value type: Int
	// Default value: DefaultHistoryMaxAutoResetPoints
	// Allowed filters: DomainName
	HistoryMaxAutoResetPoints
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
	// MaxActivityCountDispatchByDomain max # of activity tasks to dispatch to matching before creating transfer tasks. This is an performance optimization to skip activity scheduling efforts.
	// KeyName: history.activityDispatchForSyncMatchCountByDomain
	// Value type: Int
	// Default value: 0
	// Allowed filters: DomainName
	MaxActivityCountDispatchByDomain

	// key for history replication

	// ReplicationTaskFetcherParallelism determines how many go routines we spin up for fetching tasks
	// KeyName: history.ReplicationTaskFetcherParallelism
	// Value type: Int
	// Default value: 1
	// Allowed filters: N/A
	ReplicationTaskFetcherParallelism
	// ReplicationTaskProcessorErrorRetryMaxAttempts is the max retry attempts for applying replication tasks
	// KeyName: history.ReplicationTaskProcessorErrorRetryMaxAttempts
	// Value type: Int
	// Default value: 10
	// Allowed filters: ShardID
	ReplicationTaskProcessorErrorRetryMaxAttempts

	// WorkflowIDExternalRPS is the rate limit per workflowID for external calls
	// KeyName: history.workflowIDExternalRPS
	// Value type: Int
	// Default value: UnlimitedRPS
	// Allowed filters: DomainName
	WorkflowIDExternalRPS

	// WorkflowIDInternalRPS is the rate limit per workflowID for internal calls
	// KeyName: history.workflowIDInternalRPS
	// Value type: Int
	// Default value: UnlimitedRPS
	// Allowed filters: DomainName
	WorkflowIDInternalRPS

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
	// WorkerThrottledLogRPS is the rate limit on number of log messages emitted per second for throttled logger
	// KeyName: worker.throttledLogRPS
	// Value type: Int
	// Default value: 20
	// Allowed filters: N/A
	WorkerThrottledLogRPS
	// ScannerPersistenceMaxQPS is the maximum rate of persistence calls from worker.Scanner
	// KeyName: worker.scannerPersistenceMaxQPS
	// Value type: Int
	// Default value: 5
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
	// Default value: 1000
	// Allowed filters: N/A
	ScannerBatchSizeForTasklistHandler
	// ScannerMaxTasksProcessedPerTasklistJob is the number of tasks to process for a tasklist in each workflow run
	// KeyName: worker.scannerMaxTasksProcessedPerTasklistJob
	// Value type: Int
	// Default value: 256
	// Allowed filters: N/A
	ScannerMaxTasksProcessedPerTasklistJob
	// ConcreteExecutionsScannerConcurrency indicates the concurrency of concrete execution scanner
	// KeyName: worker.executionsScannerConcurrency
	// Value type: Int
	// Default value: 25
	// Allowed filters: N/A
	ConcreteExecutionsScannerConcurrency
	// ConcreteExecutionsScannerBlobstoreFlushThreshold indicates the flush threshold of blobstore in concrete execution scanner
	// KeyName: worker.executionsScannerBlobstoreFlushThreshold
	// Value type: Int
	// Default value: 100
	// Allowed filters: N/A
	ConcreteExecutionsScannerBlobstoreFlushThreshold
	// ConcreteExecutionsScannerActivityBatchSize indicates the batch size of scanner activities
	// KeyName: worker.executionsScannerActivityBatchSize
	// Value type: Int
	// Default value: 25
	// Allowed filters: N/A
	ConcreteExecutionsScannerActivityBatchSize
	// ConcreteExecutionsScannerPersistencePageSize indicates the page size of execution persistence fetches in concrete execution scanner
	// KeyName: worker.executionsScannerPersistencePageSize
	// Value type: Int
	// Default value: 1000
	// Allowed filters: N/A
	ConcreteExecutionsScannerPersistencePageSize
	// CurrentExecutionsScannerConcurrency indicates the concurrency of current executions scanner
	// KeyName: worker.currentExecutionsConcurrency
	// Value type: Int
	// Default value: 25
	// Allowed filters: N/A
	CurrentExecutionsScannerConcurrency
	// CurrentExecutionsScannerBlobstoreFlushThreshold indicates the flush threshold of blobstore in current executions scanner
	// KeyName: worker.currentExecutionsBlobstoreFlushThreshold
	// Value type: Int
	// Default value: 100
	// Allowed filters: N/A
	CurrentExecutionsScannerBlobstoreFlushThreshold
	// CurrentExecutionsScannerActivityBatchSize indicates the batch size of scanner activities
	// KeyName: worker.currentExecutionsActivityBatchSize
	// Value type: Int
	// Default value: 25
	// Allowed filters: N/A
	CurrentExecutionsScannerActivityBatchSize
	// CurrentExecutionsScannerPersistencePageSize indicates the page size of execution persistence fetches in current executions scanner
	// KeyName: worker.currentExecutionsPersistencePageSize
	// Value type: INt
	// Default value: 1000
	// Allowed filters: N/A
	CurrentExecutionsScannerPersistencePageSize
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
	// ESAnalyzerMinNumWorkflowsForAvg controls how many workflows to have at least to rely on workflow run time avg per type
	// KeyName: worker.ESAnalyzerMinNumWorkflowsForAvg
	// Value type: Int
	// Default value: 100
	ESAnalyzerMinNumWorkflowsForAvg
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

	// WorkflowDeletionJitterRange defines the duration in minutes for workflow close tasks jittering
	// KeyName: system.workflowDeletionJitterRange
	// Value type: Int
	// Default value: 1 (no jittering)
	WorkflowDeletionJitterRange

	// SampleLoggingRate defines the rate we want sampled logs to be logged at
	// KeyName: system.sampleLoggingRate
	// Value type: Int
	// Default value: 100
	SampleLoggingRate
	// LargeShardHistorySizeMetricThreshold defines the threshold for what consititutes a large history storage size to alert on
	// KeyName: system.largeShardHistorySizeMetricThreshold
	// Value type: Int
	// Default value: 10485760 (10mb)
	LargeShardHistorySizeMetricThreshold
	// LargeShardHistoryEventMetricThreshold defines the threshold for what consititutes a large history event size to alert on
	// KeyName: system.largeShardHistoryEventMetricThreshold
	// Value type: Int
	// Default value: 50 * 1024
	LargeShardHistoryEventMetricThreshold
	// LargeShardHistoryBlobMetricThreshold defines the threshold for what consititutes a large history blob size to alert on
	// KeyName: system.largeShardHistoryBlobMetricThreshold
	// Value type: Int
	// Default value: 262144 (1/4mb)

	// IsolationGroupStateUpdateRetryAttempts
	// KeyName: system.isolationGroupStateUpdateRetryAttempts
	// Value type: int
	// Default value: 2
	IsolationGroupStateUpdateRetryAttempts

	LargeShardHistoryBlobMetricThreshold
	// LastIntKey must be the last one in this const group
	LastIntKey
)

const (
	UnknownBoolKey BoolKey = iota

	// key for tests
	TestGetBoolPropertyKey
	TestGetBoolPropertyFilteredByDomainIDKey
	TestGetBoolPropertyFilteredByTaskListInfoKey

	// key for common & admin

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
	// EnableReadVisibilityFromES is key for enable read from elastic search or db visibility, usually using with AdvancedVisibilityWritingMode for seamless migration from db visibility to advanced visibility
	// KeyName: system.enableReadVisibilityFromES
	// Value type: Bool
	// Default value: true
	// Allowed filters: DomainName
	EnableReadVisibilityFromES
	// EnableReadVisibilityFromPinot is key for enable read from pinot or db visibility, usually using with AdvancedVisibilityWritingMode for seamless migration from db visibility to advanced visibility
	// KeyName: system.enableReadVisibilityFromPinot
	// Value type: Bool
	// Default value: true
	// Allowed filters: DomainName
	EnableReadVisibilityFromPinot
	// EnableLogCustomerQueryParameter is key for enable log customer query parameters
	// KeyName: system.enableLogCustomerQueryParameter
	// Value type: Bool
	// Default value: false
	EnableLogCustomerQueryParameter
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
	// EnableReadFromHistoryArchival is key for enabling reading history from archival store
	// KeyName: system.enableReadFromHistoryArchival
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	EnableReadFromHistoryArchival
	// EnableReadFromVisibilityArchival is key for enabling reading visibility from archival store to override the value from static config.
	// KeyName: system.enableReadFromVisibilityArchival
	// Value type: Bool
	// Default value: true
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
	// Default value: true
	// Allowed filters: N/A
	EnableGracefulFailover
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
	// EnableGRPCOutbound is the key for enabling outbound GRPC traffic
	// KeyName: system.enableGRPCOutbound
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	EnableGRPCOutbound
	// EnableSQLAsyncTransaction is the key for enabling async transaction
	// KeyName: system.enableSQLAsyncTransaction
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	EnableSQLAsyncTransaction

	// key for frontend

	// EnableClientVersionCheck is enables client version check for frontend
	// KeyName: frontend.enableClientVersionCheck
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	EnableClientVersionCheck
	// SendRawWorkflowHistory is whether to enable raw history retrieving
	// KeyName: frontend.sendRawWorkflowHistory
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainName
	SendRawWorkflowHistory
	// FrontendEmitSignalNameMetricsTag enables emitting signal name tag in metrics in frontend client
	// KeyName: frontend.emitSignalNameMetricsTag
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainName
	FrontendEmitSignalNameMetricsTag
	// EnableQueryAttributeValidation enables validation of queries' search attributes against the dynamic config whitelist
	// Keyname: frontend.enableQueryAttributeValidation
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	EnableQueryAttributeValidation

	// key for matching

	// MatchingEnableSyncMatch is to enable sync match
	// KeyName: matching.enableSyncMatch
	// Value type: Bool
	// Default value: true
	// Allowed filters: DomainName,TasklistName,TasklistType
	MatchingEnableSyncMatch
	// MatchingEnableTaskInfoLogByDomainID is enables info level logs for decision/activity task based on the request domainID
	// KeyName: matching.enableTaskInfoLogByDomainID
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainID
	MatchingEnableTaskInfoLogByDomainID

	// key for history

	// EventsCacheGlobalEnable is enables global cache over all history shards
	// KeyName: history.eventsCacheGlobalEnable
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	EventsCacheGlobalEnable
	// QueueProcessorEnableSplit indicates whether processing queue split policy should be enabled
	// KeyName: history.queueProcessorEnableSplit
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	QueueProcessorEnableSplit
	// QueueProcessorEnableRandomSplitByDomainID indicates whether random queue split policy should be enabled for a domain
	// KeyName: history.queueProcessorEnableRandomSplitByDomainID
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainID
	QueueProcessorEnableRandomSplitByDomainID
	// QueueProcessorEnablePendingTaskSplitByDomainID indicates whether pending task split policy should be enabled
	// KeyName: history.queueProcessorEnablePendingTaskSplitByDomainID
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainID
	QueueProcessorEnablePendingTaskSplitByDomainID
	// QueueProcessorEnableStuckTaskSplitByDomainID indicates whether stuck task split policy should be enabled
	// KeyName: history.queueProcessorEnableStuckTaskSplitByDomainID
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainID
	QueueProcessorEnableStuckTaskSplitByDomainID
	// QueueProcessorEnablePersistQueueStates indicates whether processing queue states should be persisted
	// KeyName: history.queueProcessorEnablePersistQueueStates
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	QueueProcessorEnablePersistQueueStates
	// QueueProcessorEnableLoadQueueStates indicates whether processing queue states should be loaded
	// KeyName: history.queueProcessorEnableLoadQueueStates
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	QueueProcessorEnableLoadQueueStates
	// QueueProcessorEnableGracefulSyncShutdown indicates whether processing queue should be shutdown gracefully & synchronously
	// KeyName: history.queueProcessorEnableGracefulSyncShutdown
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	QueueProcessorEnableGracefulSyncShutdown
	// ReplicationTaskFetcherEnableGracefulSyncShutdown indicates whether task fetcher should be shutdown gracefully & synchronously
	// KeyName: history.replicationTaskFetcherEnableGracefulSyncShutdown
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	ReplicationTaskFetcherEnableGracefulSyncShutdown
	// TransferProcessorEnableValidator is whether validator should be enabled for transferQueueProcessor
	// KeyName: history.transferProcessorEnableValidator
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	TransferProcessorEnableValidator
	// EnableAdminProtection is whether to enable admin checking
	// KeyName: history.enableAdminProtection
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	EnableAdminProtection
	// EnableParentClosePolicy is whether to  ParentClosePolicy
	// KeyName: history.enableParentClosePolicy
	// Value type: Bool
	// Default value: true
	// Allowed filters: DomainName
	EnableParentClosePolicy
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
	// EnableCrossClusterEngine is used as an overall switch for the cross-cluster feature, a feature which, if not enabled
	// can be quite expensive in terms of resources
	// KeyName: history.enableCrossClusterEngine
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainName
	EnableCrossClusterEngine
	// EnableCrossClusterOperationsForDomain indicates if cross cluster operations can be scheduled for a domain
	// KeyName: history.enableCrossClusterOperations
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainName
	EnableCrossClusterOperationsForDomain
	// EnableHistoryCorruptionCheck enables additional sanity check for corrupted history. This allows early catches of DB corruptions but potiantally increased latency.
	// KeyName: history.enableHistoryCorruptionCheck
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainName
	EnableHistoryCorruptionCheck
	// EnableActivityLocalDispatchByDomain is allows worker to dispatch activity tasks through local tunnel after decisions are made. This is an performance optimization to skip activity scheduling efforts
	// KeyName: history.enableActivityLocalDispatchByDomain
	// Value type: Bool
	// Default value: true
	// Allowed filters: DomainName
	EnableActivityLocalDispatchByDomain
	// HistoryEnableTaskInfoLogByDomainID is enables info level logs for decision/activity task based on the request domainID
	// KeyName: history.enableTaskInfoLogByDomainID
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainID
	HistoryEnableTaskInfoLogByDomainID
	// EnableReplicationTaskGeneration is the flag to control replication generation
	// KeyName: history.enableReplicationTaskGeneration
	// Value type: Bool
	// Default value: true
	// Allowed filters: DomainID, WorkflowID
	EnableReplicationTaskGeneration
	// UseNewInitialFailoverVersion is a switch to issue a failover version based on the minFailoverVersion
	// rather than the default initialFailoverVersion. USed as a per-domain migration switch
	// KeyName: history.useNewInitialFailoverVersion
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	UseNewInitialFailoverVersion
	// EnableRecordWorkflowExecutionUninitialized enables record workflow execution uninitialized state in ElasticSearch
	// KeyName: history.EnableRecordWorkflowExecutionUninitialized
	// Value type: Bool
	// Default value: true
	// Allowed filters: DomainName
	EnableRecordWorkflowExecutionUninitialized
	// WorkflowIDCacheEnabled is the key to enable/disable caching of workflowID specific information
	// KeyName: history.workflowIDCacheEnabled
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainName
	WorkflowIDCacheEnabled
	// AllowArchivingIncompleteHistory will continue on when seeing some error like history mutated(usually caused by database consistency issues)
	// KeyName: worker.AllowArchivingIncompleteHistory
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	AllowArchivingIncompleteHistory
	// EnableCleaningOrphanTaskInTasklistScavenger indicates if enabling the scanner to clean up orphan tasks
	// Only implemented for single SQL database. TODO https://github.com/uber/cadence/issues/4064 for supporting multiple/sharded SQL database and NoSQL
	// KeyName: worker.enableCleaningOrphanTaskInTasklistScavenger
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	EnableCleaningOrphanTaskInTasklistScavenger
	// TaskListScannerEnabled indicates if task list scanner should be started as part of worker.Scanner
	// KeyName: worker.taskListScannerEnabled
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	TaskListScannerEnabled
	// HistoryScannerEnabled indicates if history scanner should be started as part of worker.Scanner
	// KeyName: worker.historyScannerEnabled
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	HistoryScannerEnabled
	// ConcreteExecutionsScannerEnabled indicates if executions scanner should be started as part of worker.Scanner
	// KeyName: worker.executionsScannerEnabled
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	ConcreteExecutionsScannerEnabled
	// ConcreteExecutionsScannerInvariantCollectionMutableState indicates if mutable state invariant checks should be run
	// KeyName: worker.executionsScannerInvariantCollectionMutableState
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	ConcreteExecutionsScannerInvariantCollectionMutableState
	// ConcreteExecutionsScannerInvariantCollectionHistory indicates if history invariant checks should be run
	// KeyName: worker.executionsScannerInvariantCollectionHistory
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	ConcreteExecutionsScannerInvariantCollectionHistory
	// ConcreteExecutionsFixerInvariantCollectionStale indicates if the stale workflow invariant should be run
	// KeyName: worker.executionsFixerInvariantCollectionStale
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	ConcreteExecutionsFixerInvariantCollectionStale
	// ConcreteExecutionsFixerInvariantCollectionMutableState indicates if mutable state invariant checks should be run
	// KeyName: worker.executionsFixerInvariantCollectionMutableState
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	ConcreteExecutionsFixerInvariantCollectionMutableState
	// ConcreteExecutionsFixerInvariantCollectionHistory indicates if history invariant checks should be run
	// KeyName: worker.executionsFixerInvariantCollectionHistory
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	ConcreteExecutionsFixerInvariantCollectionHistory
	// ConcreteExecutionsScannerInvariantCollectionStale indicates if the stale workflow invariant should be run
	// KeyName: worker.executionsScannerInvariantCollectionStale
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	ConcreteExecutionsScannerInvariantCollectionStale
	// CurrentExecutionsScannerEnabled indicates if current executions scanner should be started as part of worker.Scanner
	// KeyName: worker.currentExecutionsScannerEnabled
	// Value type: Bool
	// Default value: false
	// Allowed filters: N/A
	CurrentExecutionsScannerEnabled
	// CurrentExecutionsScannerInvariantCollectionHistory indicates if history invariant checks should be run
	// KeyName: worker.currentExecutionsScannerInvariantCollectionHistory
	// Value type: Bool
	// Default value: true
	// Allowed filters: N/A
	CurrentExecutionsScannerInvariantCollectionHistory
	// CurrentExecutionsScannerInvariantCollectionMutableState indicates if mutable state invariant checks should be run
	// KeyName: worker.currentExecutionsInvariantCollectionMutableState
	// Value type: Bool
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

	// EnableStickyQuery indicates if sticky query should be enabled per domain
	// KeyName: system.enableStickyQuery
	// Value type: Bool
	// Default value: true
	// Allowed filters: DomainName
	EnableStickyQuery
	// EnableFailoverManager indicates if failover manager is enabled
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
	// ESAnalyzerPause defines if we want to dynamically pause the analyzer workflow
	// KeyName: worker.ESAnalyzerPause
	// Value type: bool
	// Default value: false
	ESAnalyzerPause
	// EnableArchivalCompression indicates whether blobs are compressed before they are archived
	// KeyName: N/A
	// Default value: N/A
	// TODO: https://github.com/uber/cadence/issues/3861
	EnableArchivalCompression
	// ESAnalyzerEnableAvgDurationBasedChecks controls if we want to enable avg duration based task refreshes
	// KeyName: worker.ESAnalyzerEnableAvgDurationBasedChecks
	// Value type: Bool
	// Default value: false
	ESAnalyzerEnableAvgDurationBasedChecks

	// Lockdown defines if we want to allow failovers of domains to this cluster
	// KeyName: system.Lockdown
	// Value type: bool
	// Default value: false
	Lockdown

	// PendingActivityValidationEnabled is feature flag if pending activity count validation is enabled
	// KeyName: limit.pendingActivityCount.enabled
	// Value type: bool
	// Default value: false
	EnablePendingActivityValidation

	EnableCassandraAllConsistencyLevelDelete

	// EnableTasklistIsolation Is a feature to enable subdivision of workflows by units called 'isolation-groups'
	// and to control their movement and blast radius. This has some nontrivial operational overheads in management
	// and a good understanding of poller distribution, so probably not worth enabling unless it's well understood.
	// KeyName: system.enableTasklistIsolation
	// Value type: bool
	// Default value: false
	EnableTasklistIsolation

	// EnableShardIDMetrics turns on or off shardId metrics
	// KeyName: system.enableShardIDMetrics
	// Value type: Bool
	// Default value: true
	EnableShardIDMetrics

	EnableTimerDebugLogByDomainID

	// EnableTaskVal is which allows the taskvalidation to be enabled.
	// KeyName: system.enableTaskVal
	// Value type: Bool
	// Default value: false
	// Allowed filters: DomainID
	EnableTaskVal

	// LastBoolKey must be the last one in this const group
	LastBoolKey
)

const (
	UnknownFloatKey FloatKey = iota

	// key for tests
	TestGetFloat64PropertyKey

	// key for common & admin

	// PersistenceErrorInjectionRate is rate for injecting random error in persistence
	// KeyName: system.persistenceErrorInjectionRate
	// Value type: Float64
	// Default value: 0
	// Allowed filters: N/A
	PersistenceErrorInjectionRate
	// AdminErrorInjectionRate is the rate for injecting random error in admin client
	// KeyName: admin.errorInjectionRate
	// Value type: Float64
	// Default value: 0
	// Allowed filters: N/A
	AdminErrorInjectionRate

	// key for frontend

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

	// key for matching

	// MatchingErrorInjectionRate is rate for injecting random error in matching client
	// KeyName: matching.errorInjectionRate
	// Value type: Float64
	// Default value: 0
	// Allowed filters: N/A
	MatchingErrorInjectionRate

	// key for history

	// TaskRedispatchIntervalJitterCoefficient is the task redispatch interval jitter coefficient
	// KeyName: history.taskRedispatchIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	TaskRedispatchIntervalJitterCoefficient
	// QueueProcessorRandomSplitProbability is the probability for a domain to be split to a new processing queue
	// KeyName: history.queueProcessorRandomSplitProbability
	// Value type: Float64
	// Default value: 0.01
	// Allowed filters: N/A
	QueueProcessorRandomSplitProbability
	// QueueProcessorPollBackoffIntervalJitterCoefficient is backoff interval jitter coefficient
	// KeyName: history.queueProcessorPollBackoffIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	QueueProcessorPollBackoffIntervalJitterCoefficient
	// TimerProcessorUpdateAckIntervalJitterCoefficient is the update interval jitter coefficient
	// KeyName: history.timerProcessorUpdateAckIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	TimerProcessorUpdateAckIntervalJitterCoefficient
	// TimerProcessorMaxPollIntervalJitterCoefficient is the max poll interval jitter coefficient
	// KeyName: history.timerProcessorMaxPollIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	TimerProcessorMaxPollIntervalJitterCoefficient
	// TimerProcessorSplitQueueIntervalJitterCoefficient is the split processing queue interval jitter coefficient
	// KeyName: history.timerProcessorSplitQueueIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	TimerProcessorSplitQueueIntervalJitterCoefficient
	// TransferProcessorMaxPollIntervalJitterCoefficient is the max poll interval jitter coefficient
	// KeyName: history.transferProcessorMaxPollIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	TransferProcessorMaxPollIntervalJitterCoefficient
	// TransferProcessorSplitQueueIntervalJitterCoefficient is the split processing queue interval jitter coefficient
	// KeyName: history.transferProcessorSplitQueueIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	TransferProcessorSplitQueueIntervalJitterCoefficient
	// TransferProcessorUpdateAckIntervalJitterCoefficient is the update interval jitter coefficient
	// KeyName: history.transferProcessorUpdateAckIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	TransferProcessorUpdateAckIntervalJitterCoefficient
	// CrossClusterSourceProcessorMaxPollIntervalJitterCoefficient is the max poll interval jitter coefficient
	// KeyName: history.crossClusterProcessorMaxPollIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	CrossClusterSourceProcessorMaxPollIntervalJitterCoefficient
	// CrossClusterSourceProcessorUpdateAckIntervalJitterCoefficient is the update interval jitter coefficient
	// KeyName: history.crossClusterProcessorUpdateAckIntervalJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	CrossClusterSourceProcessorUpdateAckIntervalJitterCoefficient
	// CrossClusterTargetProcessorJitterCoefficient is the jitter coefficient used in cross cluster task processor
	// KeyName: history.crossClusterTargetProcessorJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	CrossClusterTargetProcessorJitterCoefficient
	// CrossClusterFetcherJitterCoefficient is the jitter coefficient used in cross cluster task fetcher
	// KeyName: history.crossClusterFetcherJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	CrossClusterFetcherJitterCoefficient
	// ReplicationTaskProcessorCleanupJitterCoefficient is the jitter for cleanup timer
	// KeyName: history.ReplicationTaskProcessorCleanupJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: ShardID
	ReplicationTaskProcessorCleanupJitterCoefficient
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
	// MutableStateChecksumInvalidateBefore is the epoch timestamp before which all checksums are to be discarded
	// KeyName: history.mutableStateChecksumInvalidateBefore
	// Value type: Float64
	// Default value: 0
	// Allowed filters: N/A
	MutableStateChecksumInvalidateBefore
	// NotifyFailoverMarkerTimerJitterCoefficient is the jitter for failover marker notifier timer
	// KeyName: history.NotifyFailoverMarkerTimerJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	NotifyFailoverMarkerTimerJitterCoefficient
	// HistoryErrorInjectionRate is rate for injecting random error in history client
	// KeyName: history.errorInjectionRate
	// Value type: Float64
	// Default value: 0
	// Allowed filters: N/A
	HistoryErrorInjectionRate
	// ReplicationTaskFetcherTimerJitterCoefficient is the jitter for fetcher timer
	// KeyName: history.ReplicationTaskFetcherTimerJitterCoefficient
	// Value type: Float64
	// Default value: 0.15
	// Allowed filters: N/A
	ReplicationTaskFetcherTimerJitterCoefficient
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

	// LastFloatKey must be the last one in this const group
	LastFloatKey
)

const (
	UnknownStringKey StringKey = iota

	// key for tests
	TestGetStringPropertyKey

	// key for common & admin

	// AdvancedVisibilityWritingMode is key for how to write to advanced visibility. The most useful option is "dual", which can be used for seamless migration from db visibility to advanced visibility, usually using with EnableReadVisibilityFromES
	// KeyName: system.advancedVisibilityWritingMode
	// Value type: String enum: "on"(means writing to advancedVisibility only, "off" (means writing to db visibility only), or "dual" (means writing to both)
	// Default value: "on"
	// Allowed filters: N/A
	AdvancedVisibilityWritingMode
	// HistoryArchivalStatus is key for the status of history archival to override the value from static config.
	// KeyName: system.historyArchivalStatus
	// Value type: string enum: "enabled" or "disabled"
	// Default value: "enabled"
	// Allowed filters: N/A
	HistoryArchivalStatus
	// VisibilityArchivalStatus is key for the status of visibility archival to override the value from static config.
	// KeyName: system.visibilityArchivalStatus
	// Value type: string enum: "enabled" or "disabled"
	// Default value: "enabled"
	// Allowed filters: N/A
	VisibilityArchivalStatus
	// DefaultEventEncoding is the encoding type for history events
	// KeyName: history.defaultEventEncoding
	// Value type: String
	// Default value: string(common.EncodingTypeThriftRW)
	// Allowed filters: DomainName
	DefaultEventEncoding
	// AdminOperationToken is the token to pass admin checking
	// KeyName: history.adminOperationToken
	// Value type: String
	// Default value: common.DefaultAdminOperationToken
	// Allowed filters: N/A
	AdminOperationToken
	// ESAnalyzerLimitToTypes controls if we want to limit ESAnalyzer only to some workflow types
	// KeyName: worker.ESAnalyzerLimitToTypes
	// Value type: String
	// Default value: "" => means no limitation
	ESAnalyzerLimitToTypes
	// ESAnalyzerLimitToDomains controls if we want to limit ESAnalyzer only to some domains
	// KeyName: worker.ESAnalyzerLimitToDomains
	// Value type: String
	// Default value: "" => means no limitation
	ESAnalyzerLimitToDomains
	// ESAnalyzerWorkflowDurationWarnThresholds defines the warning execution thresholds for workflow types
	// KeyName: worker.ESAnalyzerWorkflowDurationWarnThresholds
	// Value type: string [{"DomainName":"<domain>", "WorkflowType":"<workflowType>", "Threshold":"<duration>", "Refresh":<shouldRefresh>, "MaxNumWorkflows":<maxNumber>}]
	// Default value: ""
	ESAnalyzerWorkflowDurationWarnThresholds
	// ESAnalyzerWorkflowVersionMetricDomains defines the domains we want to emit wf version metrics on
	// KeyName: worker.ESAnalyzerWorkflowVersionMetricDomains
	// Value type: string ["test-domain","test-domain2"]
	// Default value: ""
	ESAnalyzerWorkflowVersionMetricDomains
	// ESAnalyzerWorkflowTypeMetricDomains defines the domains we want to emit wf type metrics on
	// KeyName: worker.ESAnalyzerWorkflowTypeMetricDomains
	// Value type: string ["test-domain","test-domain2"]
	// Default value: ""
	ESAnalyzerWorkflowTypeMetricDomains

	// LastStringKey must be the last one in this const group
	LastStringKey
)

const (
	UnknownDurationKey DurationKey = iota

	// key for tests
	TestGetDurationPropertyKey
	TestGetDurationPropertyFilteredByDomainKey
	TestGetDurationPropertyFilteredByTaskListInfoKey

	// FrontendShutdownDrainDuration is the duration of traffic drain during shutdown
	// KeyName: frontend.shutdownDrainDuration
	// Value type: Duration
	// Default value: 0
	// Allowed filters: N/A
	FrontendShutdownDrainDuration
	// FrontendFailoverCoolDown is duration between two domain failvoers
	// KeyName: frontend.failoverCoolDown
	// Value type: Duration
	// Default value: 1m (one minute, see domain.FailoverCoolDown)
	// Allowed filters: DomainName
	FrontendFailoverCoolDown
	// DomainFailoverRefreshInterval is the domain failover refresh timer
	// KeyName: frontend.domainFailoverRefreshInterval
	// Value type: Duration
	// Default value: 10s (10*time.Second)
	// Allowed filters: N/A
	DomainFailoverRefreshInterval

	// MatchingLongPollExpirationInterval is the long poll expiration interval in the matching service
	// KeyName: matching.longPollExpirationInterval
	// Value type: Duration
	// Default value: time.Minute
	// Allowed filters: DomainName,TasklistName,TasklistType
	MatchingLongPollExpirationInterval
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
	// MatchingShutdownDrainDuration is the duration of traffic drain during shutdown
	// KeyName: matching.shutdownDrainDuration
	// Value type: Duration
	// Default value: 0
	// Allowed filters: N/A
	MatchingShutdownDrainDuration
	// MatchingActivityTaskSyncMatchWaitTime is the amount of time activity task will wait to be sync matched
	// KeyName: matching.activityTaskSyncMatchWaitTime
	// Value type: Duration
	// Default value: 100ms
	// Allowed filters: DomainName
	MatchingActivityTaskSyncMatchWaitTime

	// HistoryLongPollExpirationInterval is the long poll expiration interval in the history service
	// KeyName: history.longPollExpirationInterval
	// Value type: Duration
	// Default value: 20s( time.Second*20)
	// Allowed filters: DomainName
	HistoryLongPollExpirationInterval
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
	// EventsCacheTTL is TTL of events cache
	// KeyName: history.eventsCacheTTL
	// Value type: Duration
	// Default value: 1h (time.Hour)
	// Allowed filters: N/A
	EventsCacheTTL
	// AcquireShardInterval is interval that timer used to acquire shard
	// KeyName: history.acquireShardInterval
	// Value type: Duration
	// Default value: 1m (time.Minute)
	// Allowed filters: N/A
	AcquireShardInterval
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
	// TimerProcessorUpdateAckInterval is update interval for timer processor
	// KeyName: history.timerProcessorUpdateAckInterval
	// Value type: Duration
	// Default value: 30s (30*time.Second)
	// Allowed filters: N/A
	TimerProcessorUpdateAckInterval
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
	// TimerProcessorMaxPollInterval is max poll interval for timer processor
	// KeyName: history.timerProcessorMaxPollInterval
	// Value type: Duration
	// Default value: 5m (5*time.Minute)
	// Allowed filters: N/A
	TimerProcessorMaxPollInterval
	// TimerProcessorSplitQueueInterval is the split processing queue interval for timer processor
	// KeyName: history.timerProcessorSplitQueueInterval
	// Value type: Duration
	// Default value: 1m (1*time.Minute)
	// Allowed filters: N/A
	TimerProcessorSplitQueueInterval
	// TimerProcessorArchivalTimeLimit is the upper time limit for inline history archival
	// KeyName: history.timerProcessorArchivalTimeLimit
	// Value type: Duration
	// Default value: 2s (2*time.Second)
	// Allowed filters: N/A
	TimerProcessorArchivalTimeLimit
	// TimerProcessorMaxTimeShift is the max shift timer processor can have
	// KeyName: history.timerProcessorMaxTimeShift
	// Value type: Duration
	// Default value: 1s (1*time.Second)
	// Allowed filters: N/A
	TimerProcessorMaxTimeShift
	// TransferProcessorFailoverMaxStartJitterInterval is the max jitter interval for starting transfer
	// failover queue processing. The actual jitter interval used will be a random duration between
	// 0 and the max interval so that timer failover queue across different shards won't start at
	// the same time
	// KeyName: history.transferProcessorFailoverMaxStartJitterInterval
	// Value type: Duration
	// Default value: 0s (0*time.Second)
	// Allowed filters: N/A
	TransferProcessorFailoverMaxStartJitterInterval
	// TransferProcessorMaxPollInterval is max poll interval for transferQueueProcessor
	// KeyName: history.transferProcessorMaxPollInterval
	// Value type: Duration
	// Default value: 1m (1*time.Minute)
	// Allowed filters: N/A
	TransferProcessorMaxPollInterval
	// TransferProcessorSplitQueueInterval is the split processing queue interval for transferQueueProcessor
	// KeyName: history.transferProcessorSplitQueueInterval
	// Value type: Duration
	// Default value: 1m (1*time.Minute)
	// Allowed filters: N/A
	TransferProcessorSplitQueueInterval
	// TransferProcessorUpdateAckInterval is update interval for transferQueueProcessor
	// KeyName: history.transferProcessorUpdateAckInterval
	// Value type: Duration
	// Default value: 30s (30*time.Second)
	// Allowed filters: N/A
	TransferProcessorUpdateAckInterval
	// TransferProcessorCompleteTransferInterval is complete timer interval for transferQueueProcessor
	// KeyName: history.transferProcessorCompleteTransferInterval
	// Value type: Duration
	// Default value: 60s (60*time.Second)
	// Allowed filters: N/A
	TransferProcessorCompleteTransferInterval
	// TransferProcessorValidationInterval is interval for performing transfer queue validation
	// KeyName: history.transferProcessorValidationInterval
	// Value type: Duration
	// Default value: 30s (30*time.Second)
	// Allowed filters: N/A
	TransferProcessorValidationInterval
	// TransferProcessorVisibilityArchivalTimeLimit is the upper time limit for archiving visibility records
	// KeyName: history.transferProcessorVisibilityArchivalTimeLimit
	// Value type: Duration
	// Default value: 400ms (400*time.Millisecond)
	// Allowed filters: N/A
	TransferProcessorVisibilityArchivalTimeLimit
	// CrossClusterSourceProcessorMaxPollInterval is max poll interval for crossClusterQueueProcessor
	// KeyName: history.crossClusterProcessorMaxPollInterval
	// Value type: Duration
	// Default value: 1m (1*time.Minute)
	// Allowed filters: N/A
	CrossClusterSourceProcessorMaxPollInterval
	// CrossClusterSourceProcessorUpdateAckInterval is update interval for crossClusterQueueProcessor
	// KeyName: history.crossClusterProcessorUpdateAckInterval
	// Value type: Duration
	// Default value: 30s (30*time.Second)
	// Allowed filters: N/A
	CrossClusterSourceProcessorUpdateAckInterval
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
	// ReplicatorUpperLatency indicates the max allowed replication latency between clusters
	// KeyName: history.replicatorUpperLatency
	// Value type: Duration
	// Default value: 40s (40 * time.Second)
	// Allowed filters: N/A
	ReplicatorUpperLatency
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
	// NormalDecisionScheduleToStartTimeout is scheduleToStart timeout duration for normal (non-sticky) decision task
	// KeyName: history.normalDecisionScheduleToStartTimeout
	// Value type: Duration
	// Default value: time.Minute*5
	// Allowed filters: DomainName
	NormalDecisionScheduleToStartTimeout
	// NotifyFailoverMarkerInterval is determines the frequency to notify failover marker
	// KeyName: history.NotifyFailoverMarkerInterval
	// Value type: Duration
	// Default value: 5s (5*time.Second)
	// Allowed filters: N/A
	NotifyFailoverMarkerInterval
	// ActivityMaxScheduleToStartTimeoutForRetry is maximum value allowed when overwritting the schedule to start timeout for activities with retry policy
	// KeyName: history.activityMaxScheduleToStartTimeoutForRetry
	// Value type: Duration
	// Default value: 30m (30*time.Minute)
	// Allowed filters: DomainName
	ActivityMaxScheduleToStartTimeoutForRetry
	// ReplicationTaskFetcherAggregationInterval determines how frequently the fetch requests are sent
	// KeyName: history.ReplicationTaskFetcherAggregationInterval
	// Value type: Duration
	// Default value: 2s (2 * time.Second)
	// Allowed filters: N/A
	ReplicationTaskFetcherAggregationInterval
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
	// ReplicationTaskProcessorErrorSecondRetryWait is the initial retry wait for the second phase retry
	// KeyName: history.ReplicationTaskProcessorErrorSecondRetryWait
	// Value type: Duration
	// Default value: 5s (5* time.Second)
	// Allowed filters: ShardID
	ReplicationTaskProcessorErrorSecondRetryWait
	// ReplicationTaskProcessorErrorSecondRetryMaxWait is the max wait time for the second phase retry
	// KeyName: history.ReplicationTaskProcessorErrorSecondRetryMaxWait
	// Value type: Duration
	// Default value: 30s (30 * time.Second)
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
	// ReplicationTaskProcessorStartWait is the wait time before each task processing batch
	// KeyName: history.ReplicationTaskProcessorStartWait
	// Value type: Duration
	// Default value: 5s (5* time.Second)
	// Allowed filters: ShardID
	ReplicationTaskProcessorStartWait
	// WorkerESProcessorFlushInterval is flush interval for esProcessor
	// KeyName: worker.ESProcessorFlushInterval
	// Value type: Duration
	// Default value: 1s (1*time.Second)
	// Allowed filters: N/A
	WorkerESProcessorFlushInterval
	// WorkerTimeLimitPerArchivalIteration is controls the time limit of each iteration of archival workflow
	// KeyName: worker.TimeLimitPerArchivalIteration
	// Value type: Duration
	// Default value: archiver.MaxArchivalIterationTimeout()
	// Allowed filters: N/A
	WorkerTimeLimitPerArchivalIteration
	// WorkerReplicationTaskMaxRetryDuration is the max retry duration for any task
	// KeyName: worker.replicationTaskMaxRetryDuration
	// Value type: Duration
	// Default value: 10m (time.Minute*10)
	// Allowed filters: N/A
	WorkerReplicationTaskMaxRetryDuration
	// ESAnalyzerTimeWindow defines the time window ElasticSearch Analyzer will consider while taking workflow averages
	// KeyName: worker.ESAnalyzerTimeWindow
	// Value type: Duration
	// Default value: 30 days
	ESAnalyzerTimeWindow
	// ESAnalyzerBufferWaitTime controls min time required to consider a worklow stuck
	// KeyName: worker.ESAnalyzerBufferWaitTime
	// Value type: Duration
	// Default value: 30 minutes
	ESAnalyzerBufferWaitTime
	// IsolationGroupStateRefreshInterval
	// KeyName: system.isolationGroupStateRefreshInterval
	// Value type: Duration
	// Default value: 30 seconds
	IsolationGroupStateRefreshInterval
	// IsolationGroupStateFetchTimeout is the dynamic config DB fetch timeout value
	// KeyName: system.isolationGroupStateFetchTimeout
	// Value type: Duration
	// Default value: 30 seconds
	IsolationGroupStateFetchTimeout
	// IsolationGroupStateUpdateTimeout is the dynamic config DB update timeout value
	// KeyName: system.isolationGroupStateUpdateTimeout
	// Value type: Duration
	// Default value: 30 seconds
	IsolationGroupStateUpdateTimeout

	// AsyncTaskDispatchTimeout is the timeout of dispatching tasks for async match
	// KeyName: matching.asyncTaskDispatchTimeout
	// Value type: Duration
	// Default value: 3 seconds
	// Allowed filters: domainName, taskListName, taskListType
	AsyncTaskDispatchTimeout

	// LastDurationKey must be the last one in this const group
	LastDurationKey
)

const (
	UnknownMapKey MapKey = iota

	// key for tests
	TestGetMapPropertyKey

	// key for common & admin

	// RequiredDomainDataKeys is the key for the list of data keys required in domain registration
	// KeyName: system.requiredDomainDataKeys
	// Value type: Map
	// Default value: nil
	// Allowed filters: N/A
	RequiredDomainDataKeys

	// key for frontend

	// ValidSearchAttributes is legal indexed keys that can be used in list APIs. When overriding, ensure to include the existing default attributes of the current release
	// KeyName: frontend.validSearchAttributes
	// Value type: Map
	// Default value: the default attributes of this release version, see definition.GetDefaultIndexedKeys()
	// Allowed filters: N/A
	ValidSearchAttributes

	// key for history

	// TaskSchedulerRoundRobinWeights is the priority weight for weighted round robin task scheduler
	// KeyName: history.taskSchedulerRoundRobinWeight
	// Value type: Map
	// Default value: please see common.ConvertIntMapToDynamicConfigMapProperty(DefaultTaskPriorityWeight) in code base
	// Allowed filters: N/A
	TaskSchedulerRoundRobinWeights
	// QueueProcessorPendingTaskSplitThreshold is the threshold for the number of pending tasks per domain
	// KeyName: history.queueProcessorPendingTaskSplitThreshold
	// Value type: Map
	// Default value: see common.ConvertIntMapToDynamicConfigMapProperty(DefaultPendingTaskSplitThreshold) in code base
	// Allowed filters: N/A
	QueueProcessorPendingTaskSplitThreshold
	// QueueProcessorStuckTaskSplitThreshold is the threshold for the number of attempts of a task
	// KeyName: history.queueProcessorStuckTaskSplitThreshold
	// Value type: Map
	// Default value: see common.ConvertIntMapToDynamicConfigMapProperty(DefaultStuckTaskSplitThreshold) in code base
	// Allowed filters: N/A
	QueueProcessorStuckTaskSplitThreshold

	// LastMapKey must be the last one in this const group
	LastMapKey
)

const (
	UnknownListKey ListKey = iota
	TestGetListPropertyKey

	// HeaderForwardingRules defines which headers are forwarded from inbound calls to outbound.
	// This value is only loaded at startup.
	//
	// Regexes and header names are used as-is, you are strongly encouraged to use `(?i)` to make your regex case-insensitive.
	//
	// KeyName: admin.HeaderForwardingRules
	// Value type: []rpc.HeaderRule or an []interface{} containing `map[string]interface{}{"Add":bool,"Match":string}` values.
	// Default value: forward all headers.  (this is a problematic value, and it will be changing as we reduce to a list of known values)
	HeaderForwardingRules
	// AllIsolationGroups is the list of all possible isolation groups in a service
	// KeyName: system.allIsolationGroups
	// Value type: []string
	// Default value: N/A
	// Allowed filters: N/A
	AllIsolationGroups

	LastListKey
)

// DefaultIsolationGroupConfigStoreManagerGlobalMapping is the dynamic config value for isolation groups
// Note: This is not typically used for normal dynamic config (type 0), but instead
// it's used only for IsolationGroup config (type 1).
// KeyName: system.defaultIsolationGroupConfigStoreManagerGlobalMapping
const DefaultIsolationGroupConfigStoreManagerGlobalMapping ListKey = -1 // This is a hack to put it in a different list due to it being a different config type

var IntKeys = map[IntKey]DynamicInt{
	TestGetIntPropertyKey: DynamicInt{
		KeyName:      "testGetIntPropertyKey",
		Description:  "",
		DefaultValue: 0,
	},
	TestGetIntPropertyFilteredByDomainKey: DynamicInt{
		KeyName:      "testGetIntPropertyFilteredByDomainKey",
		Description:  "",
		DefaultValue: 0,
	},
	TestGetIntPropertyFilteredByTaskListInfoKey: DynamicInt{
		KeyName:      "testGetIntPropertyFilteredByTaskListInfoKey",
		Description:  "",
		DefaultValue: 0,
	},
	TransactionSizeLimit: DynamicInt{
		KeyName:      "system.transactionSizeLimit",
		Description:  "TransactionSizeLimit is the largest allowed transaction size to persistence",
		DefaultValue: 14680064,
	},
	MaxRetentionDays: DynamicInt{
		KeyName:      "system.maxRetentionDays",
		Description:  "MaxRetentionDays is the maximum allowed retention days for domain",
		DefaultValue: 30,
	},
	MinRetentionDays: DynamicInt{
		KeyName:      "system.minRetentionDays",
		Description:  "MinRetentionDays is the minimal allowed retention days for domain",
		DefaultValue: 1,
	},
	MaxDecisionStartToCloseSeconds: DynamicInt{
		KeyName:      "system.maxDecisionStartToCloseSeconds",
		Description:  "MaxDecisionStartToCloseSeconds is the maximum allowed value for decision start to close timeout in seconds",
		DefaultValue: 240,
	},
	BlobSizeLimitError: DynamicInt{
		KeyName:      "limit.blobSize.error",
		Description:  "BlobSizeLimitError is the per event blob size limit",
		DefaultValue: 2 * 1024 * 1024,
	},
	BlobSizeLimitWarn: DynamicInt{
		KeyName:      "limit.blobSize.warn",
		Filters:      []Filter{DomainName},
		Description:  "BlobSizeLimitWarn is the per event blob size limit for warning",
		DefaultValue: 256 * 1024,
	},
	HistorySizeLimitError: DynamicInt{
		KeyName:      "limit.historySize.error",
		Filters:      []Filter{DomainName},
		Description:  "HistorySizeLimitError is the per workflow execution history size limit",
		DefaultValue: 200 * 1024 * 1024,
	},
	HistorySizeLimitWarn: DynamicInt{
		KeyName:      "limit.historySize.warn",
		Filters:      []Filter{DomainName},
		Description:  "HistorySizeLimitWarn is the per workflow execution history size limit for warning",
		DefaultValue: 50 * 1024 * 1024,
	},
	HistoryCountLimitError: DynamicInt{
		KeyName:      "limit.historyCount.error",
		Filters:      []Filter{DomainName},
		Description:  "HistoryCountLimitError is the per workflow execution history event count limit",
		DefaultValue: 200 * 1024,
	},
	HistoryCountLimitWarn: DynamicInt{
		KeyName:      "limit.historyCount.warn",
		Filters:      []Filter{DomainName},
		Description:  "HistoryCountLimitWarn is the per workflow execution history event count limit for warning",
		DefaultValue: 50 * 1024,
	},
	PendingActivitiesCountLimitError: DynamicInt{
		KeyName:      "limit.pendingActivityCount.error",
		Description:  "PendingActivitiesCountLimitError is the limit of how many pending activities a workflow can have at a point in time",
		DefaultValue: 1024,
	},
	PendingActivitiesCountLimitWarn: DynamicInt{
		KeyName:      "limit.pendingActivityCount.warn",
		Description:  "PendingActivitiesCountLimitWarn is the limit of how many activities a workflow can have before a warning is logged",
		DefaultValue: 512,
	},
	DomainNameMaxLength: DynamicInt{
		KeyName:      "limit.domainNameLength",
		Filters:      []Filter{DomainName},
		Description:  "DomainNameMaxLength is the length limit for domain name",
		DefaultValue: 1000,
	},
	IdentityMaxLength: DynamicInt{
		KeyName:      "limit.identityLength",
		Filters:      []Filter{DomainName},
		Description:  "IdentityMaxLength is the length limit for identity",
		DefaultValue: 1000,
	},
	WorkflowIDMaxLength: DynamicInt{
		KeyName:      "limit.workflowIDLength",
		Filters:      []Filter{DomainName},
		Description:  "WorkflowIDMaxLength is the length limit for workflowID",
		DefaultValue: 1000,
	},
	SignalNameMaxLength: DynamicInt{
		KeyName:      "limit.signalNameLength",
		Filters:      []Filter{DomainName},
		Description:  "SignalNameMaxLength is the length limit for signal name",
		DefaultValue: 1000,
	},
	WorkflowTypeMaxLength: DynamicInt{
		KeyName:      "limit.workflowTypeLength",
		Filters:      []Filter{DomainName},
		Description:  "WorkflowTypeMaxLength is the length limit for workflow type",
		DefaultValue: 1000,
	},
	RequestIDMaxLength: DynamicInt{
		KeyName:      "limit.requestIDLength",
		Filters:      []Filter{DomainName},
		Description:  "RequestIDMaxLength is the length limit for requestID",
		DefaultValue: 1000,
	},
	TaskListNameMaxLength: DynamicInt{
		KeyName:      "limit.taskListNameLength",
		Filters:      []Filter{DomainName},
		Description:  "TaskListNameMaxLength is the length limit for task list name",
		DefaultValue: 1000,
	},
	ActivityIDMaxLength: DynamicInt{
		KeyName:      "limit.activityIDLength",
		Filters:      []Filter{DomainName},
		Description:  "ActivityIDMaxLength is the length limit for activityID",
		DefaultValue: 1000,
	},
	ActivityTypeMaxLength: DynamicInt{
		KeyName:      "limit.activityTypeLength",
		Filters:      []Filter{DomainName},
		Description:  "ActivityTypeMaxLength is the length limit for activity type",
		DefaultValue: 1000,
	},
	MarkerNameMaxLength: DynamicInt{
		KeyName:      "limit.markerNameLength",
		Filters:      []Filter{DomainName},
		Description:  "MarkerNameMaxLength is the length limit for marker name",
		DefaultValue: 1000,
	},
	TimerIDMaxLength: DynamicInt{
		KeyName:      "limit.timerIDLength",
		Filters:      []Filter{DomainName},
		Description:  "TimerIDMaxLength is the length limit for timerID",
		DefaultValue: 1000,
	},
	MaxIDLengthWarnLimit: DynamicInt{
		KeyName:      "limit.maxIDWarnLength",
		Description:  "MaxIDLengthWarnLimit is the warn length limit for various IDs, including: Domain, TaskList, WorkflowID, ActivityID, TimerID, WorkflowType, ActivityType, SignalName, MarkerName, ErrorReason/FailureReason/CancelCause, Identity, RequestID",
		DefaultValue: 128,
	},
	FrontendPersistenceMaxQPS: DynamicInt{
		KeyName:      "frontend.persistenceMaxQPS",
		Description:  "FrontendPersistenceMaxQPS is the max qps frontend host can query DB",
		DefaultValue: 2000,
	},
	FrontendPersistenceGlobalMaxQPS: DynamicInt{
		KeyName:      "frontend.persistenceGlobalMaxQPS",
		Description:  "FrontendPersistenceGlobalMaxQPS is the max qps frontend cluster can query DB",
		DefaultValue: 0,
	},
	FrontendVisibilityMaxPageSize: DynamicInt{
		KeyName:      "frontend.visibilityMaxPageSize",
		Filters:      []Filter{DomainName},
		Description:  "FrontendVisibilityMaxPageSize is default max size for ListWorkflowExecutions in one page",
		DefaultValue: 1000,
	},
	FrontendVisibilityListMaxQPS: DynamicInt{
		KeyName: "frontend.visibilityListMaxQPS",
		Description: "deprecated: never used for ratelimiting, only sampling-based failure injection, and only on database-based visibility.\n" +
			"FrontendVisibilityListMaxQPS is max qps frontend can list open/close workflows",
		DefaultValue: 10,
	},
	FrontendESVisibilityListMaxQPS: DynamicInt{
		KeyName: "frontend.esVisibilityListMaxQPS",
		Description: "deprecated: never read from, all ES reads and writes erroneously use PersistenceMaxQPS.\n" +
			"FrontendESVisibilityListMaxQPS is max qps frontend can list open/close workflows from ElasticSearch",
		DefaultValue: 30,
	},
	FrontendESIndexMaxResultWindow: DynamicInt{
		KeyName:      "frontend.esIndexMaxResultWindow",
		Description:  "FrontendESIndexMaxResultWindow is ElasticSearch index setting max_result_window",
		DefaultValue: 10000,
	},
	FrontendHistoryMaxPageSize: DynamicInt{
		KeyName:      "frontend.historyMaxPageSize",
		Filters:      []Filter{DomainName},
		Description:  "FrontendHistoryMaxPageSize is default max size for GetWorkflowExecutionHistory in one page",
		DefaultValue: 1000,
	},
	FrontendUserRPS: DynamicInt{
		KeyName:      "frontend.rps",
		Description:  "FrontendUserRPS is workflow rate limit per second",
		DefaultValue: 1200,
	},
	FrontendWorkerRPS: DynamicInt{
		KeyName:      "frontend.workerrps",
		Description:  "FrontendWorkerRPS is background-processing workflow rate limit per second",
		DefaultValue: UnlimitedRPS,
	},
	FrontendVisibilityRPS: DynamicInt{
		KeyName:      "frontend.visibilityrps",
		Description:  "FrontendVisibilityRPS is the global workflow List*WorkflowExecutions request rate limit per second",
		DefaultValue: UnlimitedRPS,
	},
	FrontendMaxDomainUserRPSPerInstance: DynamicInt{
		KeyName:      "frontend.domainrps",
		Filters:      []Filter{DomainName},
		Description:  "FrontendMaxDomainUserRPSPerInstance is workflow domain rate limit per second",
		DefaultValue: 1200,
	},
	FrontendMaxDomainWorkerRPSPerInstance: DynamicInt{
		KeyName:      "frontend.domainworkerrps",
		Filters:      []Filter{DomainName},
		Description:  "FrontendMaxDomainWorkerRPSPerInstance is background-processing workflow domain rate limit per second",
		DefaultValue: UnlimitedRPS,
	},
	FrontendMaxDomainVisibilityRPSPerInstance: DynamicInt{
		KeyName:      "frontend.domainvisibilityrps",
		Filters:      []Filter{DomainName},
		Description:  "FrontendMaxDomainVisibilityRPSPerInstance is the per-instance List*WorkflowExecutions request rate limit per second",
		DefaultValue: UnlimitedRPS,
	},
	FrontendGlobalDomainUserRPS: DynamicInt{
		KeyName:      "frontend.globalDomainrps",
		Filters:      []Filter{DomainName},
		Description:  "FrontendGlobalDomainUserRPS is workflow domain rate limit per second for the whole Cadence cluster",
		DefaultValue: 0,
	},
	FrontendGlobalDomainWorkerRPS: DynamicInt{
		KeyName:      "frontend.globalDomainWorkerrps",
		Filters:      []Filter{DomainName},
		Description:  "FrontendGlobalDomainWorkerRPS is background-processing workflow domain rate limit per second for the whole Cadence cluster",
		DefaultValue: UnlimitedRPS,
	},
	FrontendGlobalDomainVisibilityRPS: DynamicInt{
		KeyName:      "frontend.globalDomainVisibilityrps",
		Filters:      []Filter{DomainName},
		Description:  "FrontendGlobalDomainVisibilityRPS is the per-domain List*WorkflowExecutions request rate limit per second",
		DefaultValue: UnlimitedRPS,
	},
	FrontendDecisionResultCountLimit: DynamicInt{
		KeyName:      "frontend.decisionResultCountLimit",
		Filters:      []Filter{DomainName},
		Description:  "FrontendDecisionResultCountLimit is max number of decisions per RespondDecisionTaskCompleted request",
		DefaultValue: 0,
	},
	FrontendHistoryMgrNumConns: DynamicInt{
		KeyName:      "frontend.historyMgrNumConns",
		Description:  "FrontendHistoryMgrNumConns is for persistence cluster.NumConns",
		DefaultValue: 10,
	},
	FrontendThrottledLogRPS: DynamicInt{
		KeyName:      "frontend.throttledLogRPS",
		Description:  "FrontendThrottledLogRPS is the rate limit on number of log messages emitted per second for throttled logger",
		DefaultValue: 20,
	},
	FrontendMaxBadBinaries: DynamicInt{
		KeyName:      "frontend.maxBadBinaries",
		Filters:      []Filter{DomainName},
		Description:  "FrontendMaxBadBinaries is the max number of bad binaries in domain config",
		DefaultValue: 10,
	},
	SearchAttributesNumberOfKeysLimit: DynamicInt{
		KeyName:      "frontend.searchAttributesNumberOfKeysLimit",
		Filters:      []Filter{DomainName},
		Description:  "SearchAttributesNumberOfKeysLimit is the limit of number of keys",
		DefaultValue: 100,
	},
	SearchAttributesSizeOfValueLimit: DynamicInt{
		KeyName:      "frontend.searchAttributesSizeOfValueLimit",
		Filters:      []Filter{DomainName},
		Description:  "SearchAttributesSizeOfValueLimit is the size limit of each value",
		DefaultValue: 2048,
	},
	SearchAttributesTotalSizeLimit: DynamicInt{
		KeyName:      "frontend.searchAttributesTotalSizeLimit",
		Filters:      []Filter{DomainName},
		Description:  "SearchAttributesTotalSizeLimit is the size limit of the whole map",
		DefaultValue: 40 * 1024,
	},
	VisibilityArchivalQueryMaxPageSize: DynamicInt{
		KeyName:      "frontend.visibilityArchivalQueryMaxPageSize",
		Description:  "VisibilityArchivalQueryMaxPageSize is the maximum page size for a visibility archival query",
		DefaultValue: 10000,
	},
	MatchingUserRPS: DynamicInt{
		KeyName:      "matching.rps",
		Description:  "MatchingUserRPS is request rate per second for each matching host",
		DefaultValue: 1200,
	},
	MatchingWorkerRPS: DynamicInt{
		KeyName:      "matching.workerrps",
		Description:  "MatchingWorkerRPS is background-processing request rate per second for each matching host",
		DefaultValue: UnlimitedRPS,
	},
	MatchingDomainUserRPS: DynamicInt{
		KeyName:      "matching.domainrps",
		Description:  "MatchingDomainUserRPS is request rate per domain per second for each matching host",
		DefaultValue: 0,
	},
	MatchingDomainWorkerRPS: DynamicInt{
		KeyName:      "matching.domainworkerrps",
		Description:  "MatchingDomainWorkerRPS is background-processing request rate per domain per second for each matching host",
		DefaultValue: UnlimitedRPS,
	},
	MatchingPersistenceMaxQPS: DynamicInt{
		KeyName:      "matching.persistenceMaxQPS",
		Description:  "MatchingPersistenceMaxQPS is the max qps matching host can query DB",
		DefaultValue: 3000,
	},
	MatchingPersistenceGlobalMaxQPS: DynamicInt{
		KeyName:      "matching.persistenceGlobalMaxQPS",
		Description:  "MatchingPersistenceGlobalMaxQPS is the max qps matching cluster can query DB",
		DefaultValue: 0,
	},
	MatchingMinTaskThrottlingBurstSize: DynamicInt{
		KeyName:      "matching.minTaskThrottlingBurstSize",
		Filters:      []Filter{DomainName, TaskListName, TaskType},
		Description:  "MatchingMinTaskThrottlingBurstSize is the minimum burst size for task list throttling",
		DefaultValue: 1,
	},
	MatchingGetTasksBatchSize: DynamicInt{
		KeyName:      "matching.getTasksBatchSize",
		Filters:      []Filter{DomainName, TaskListName, TaskType},
		Description:  "MatchingGetTasksBatchSize is the maximum batch size to fetch from the task buffer",
		DefaultValue: 1000,
	},
	MatchingOutstandingTaskAppendsThreshold: DynamicInt{
		KeyName:      "matching.outstandingTaskAppendsThreshold",
		Filters:      []Filter{DomainName, TaskListName, TaskType},
		Description:  "MatchingOutstandingTaskAppendsThreshold is the threshold for outstanding task appends",
		DefaultValue: 250,
	},
	MatchingMaxTaskBatchSize: DynamicInt{
		KeyName:      "matching.maxTaskBatchSize",
		Filters:      []Filter{DomainName, TaskListName, TaskType},
		Description:  "MatchingMaxTaskBatchSize is max batch size for task writer",
		DefaultValue: 100,
	},
	MatchingMaxTaskDeleteBatchSize: DynamicInt{
		KeyName:      "matching.maxTaskDeleteBatchSize",
		Filters:      []Filter{DomainName, TaskListName, TaskType},
		Description:  "MatchingMaxTaskDeleteBatchSize is the max batch size for range deletion of tasks",
		DefaultValue: 100,
	},
	MatchingThrottledLogRPS: DynamicInt{
		KeyName:      "matching.throttledLogRPS",
		Description:  "MatchingThrottledLogRPS is the rate limit on number of log messages emitted per second for throttled logger",
		DefaultValue: 20,
	},
	MatchingNumTasklistWritePartitions: DynamicInt{
		KeyName:      "matching.numTasklistWritePartitions",
		Filters:      []Filter{DomainName, TaskListName, TaskType},
		Description:  "MatchingNumTasklistWritePartitions is the number of write partitions for a task list",
		DefaultValue: 1,
	},
	MatchingNumTasklistReadPartitions: DynamicInt{
		KeyName:      "matching.numTasklistReadPartitions",
		Filters:      []Filter{DomainName, TaskListName, TaskType},
		Description:  "MatchingNumTasklistReadPartitions is the number of read partitions for a task list",
		DefaultValue: 1,
	},
	MatchingForwarderMaxOutstandingPolls: DynamicInt{
		KeyName:      "matching.forwarderMaxOutstandingPolls",
		Filters:      []Filter{DomainName, TaskListName, TaskType},
		Description:  "MatchingForwarderMaxOutstandingPolls is the max number of inflight polls from the forwarder",
		DefaultValue: 1,
	},
	MatchingForwarderMaxOutstandingTasks: DynamicInt{
		KeyName:      "matching.forwarderMaxOutstandingTasks",
		Filters:      []Filter{DomainName, TaskListName, TaskType},
		Description:  "MatchingForwarderMaxOutstandingTasks is the max number of inflight addTask/queryTask from the forwarder",
		DefaultValue: 1,
	},
	MatchingForwarderMaxRatePerSecond: DynamicInt{
		KeyName:      "matching.forwarderMaxRatePerSecond",
		Filters:      []Filter{DomainName, TaskListName, TaskType},
		Description:  "MatchingForwarderMaxRatePerSecond is the max rate at which add/query can be forwarded",
		DefaultValue: 10,
	},
	MatchingForwarderMaxChildrenPerNode: DynamicInt{
		KeyName:      "matching.forwarderMaxChildrenPerNode",
		Filters:      []Filter{DomainName, TaskListName, TaskType},
		Description:  "MatchingForwarderMaxChildrenPerNode is the max number of children per node in the task list partition tree",
		DefaultValue: 20,
	},
	HistoryRPS: DynamicInt{
		KeyName:      "history.rps",
		Description:  "HistoryRPS is request rate per second for each history host",
		DefaultValue: 3000,
	},
	HistoryPersistenceMaxQPS: DynamicInt{
		KeyName:      "history.persistenceMaxQPS",
		Description:  "HistoryPersistenceMaxQPS is the max qps history host can query DB",
		DefaultValue: 9000,
	},
	HistoryPersistenceGlobalMaxQPS: DynamicInt{
		KeyName:      "history.persistenceGlobalMaxQPS",
		Description:  "HistoryPersistenceGlobalMaxQPS is the max qps history cluster can query DB",
		DefaultValue: 0,
	},
	HistoryVisibilityOpenMaxQPS: DynamicInt{
		KeyName:      "history.historyVisibilityOpenMaxQPS",
		Filters:      []Filter{DomainName},
		Description:  "HistoryVisibilityOpenMaxQPS is max qps one history host can write visibility open_executions",
		DefaultValue: 300,
	},
	HistoryVisibilityClosedMaxQPS: DynamicInt{
		KeyName:      "history.historyVisibilityClosedMaxQPS",
		Filters:      []Filter{DomainName},
		Description:  "HistoryVisibilityClosedMaxQPS is max qps one history host can write visibility closed_executions",
		DefaultValue: 300,
	},
	HistoryCacheInitialSize: DynamicInt{
		KeyName:      "history.cacheInitialSize",
		Description:  "HistoryCacheInitialSize is initial size of history cache",
		DefaultValue: 128,
	},
	HistoryCacheMaxSize: DynamicInt{
		KeyName:      "history.cacheMaxSize",
		Description:  "HistoryCacheMaxSize is max size of history cache",
		DefaultValue: 512,
	},
	EventsCacheInitialCount: DynamicInt{
		KeyName:      "history.eventsCacheInitialSize",
		Description:  "EventsCacheInitialCount is initial count of events cache",
		DefaultValue: 128,
	},
	EventsCacheMaxCount: DynamicInt{
		KeyName:      "history.eventsCacheMaxSize",
		Description:  "EventsCacheMaxCount is max count of events cache",
		DefaultValue: 512,
	},
	EventsCacheMaxSize: DynamicInt{
		KeyName:      "history.eventsCacheMaxSizeInBytes",
		Description:  "EventsCacheMaxSize is max size of events cache in bytes",
		DefaultValue: 0,
	},
	EventsCacheGlobalInitialCount: DynamicInt{
		KeyName:      "history.eventsCacheGlobalInitialSize",
		Description:  "EventsCacheGlobalInitialCount is initial count of global events cache",
		DefaultValue: 4096,
	},
	EventsCacheGlobalMaxCount: DynamicInt{
		KeyName:      "history.eventsCacheGlobalMaxSize",
		Description:  "EventsCacheGlobalMaxCount is max count of global events cache",
		DefaultValue: 131072,
	},
	AcquireShardConcurrency: DynamicInt{
		KeyName:      "history.acquireShardConcurrency",
		Description:  "AcquireShardConcurrency is number of goroutines that can be used to acquire shards in the shard controller.",
		DefaultValue: 1,
	},
	TaskProcessRPS: DynamicInt{
		KeyName:      "history.taskProcessRPS",
		Filters:      []Filter{DomainName},
		Description:  "TaskProcessRPS is the task processing rate per second for each domain",
		DefaultValue: 1000,
	},
	TaskSchedulerType: DynamicInt{
		KeyName:      "history.taskSchedulerType",
		Description:  "TaskSchedulerType is the task scheduler type for priority task processor",
		DefaultValue: 2, // int(task.SchedulerTypeWRR),
	},
	TaskSchedulerWorkerCount: DynamicInt{
		KeyName:      "history.taskSchedulerWorkerCount",
		Description:  "TaskSchedulerWorkerCount is the number of workers per host in task scheduler",
		DefaultValue: 200,
	},
	TaskSchedulerShardWorkerCount: DynamicInt{
		KeyName:      "history.taskSchedulerShardWorkerCount",
		Description:  "TaskSchedulerShardWorkerCount is the number of worker per shard in task scheduler",
		DefaultValue: 0,
	},
	TaskSchedulerQueueSize: DynamicInt{
		KeyName:      "history.taskSchedulerQueueSize",
		Description:  "TaskSchedulerQueueSize is the size of task channel for host level task scheduler",
		DefaultValue: 10000,
	},
	TaskSchedulerShardQueueSize: DynamicInt{
		KeyName:      "history.taskSchedulerShardQueueSize",
		Description:  "TaskSchedulerShardQueueSize is the size of task channel for shard level task scheduler",
		DefaultValue: 200,
	},
	TaskSchedulerDispatcherCount: DynamicInt{
		KeyName:      "history.taskSchedulerDispatcherCount",
		Description:  "TaskSchedulerDispatcherCount is the number of task dispatcher in task scheduler (only applies to host level task scheduler)",
		DefaultValue: 1,
	},
	TaskCriticalRetryCount: DynamicInt{
		KeyName:      "history.taskCriticalRetryCount",
		Description:  "TaskCriticalRetryCount is the critical retry count for background tasks, when task attempt exceeds this threshold:- task attempt metrics and additional error logs will be emitted- task priority will be lowered",
		DefaultValue: 50,
	},
	QueueProcessorSplitMaxLevel: DynamicInt{
		KeyName:      "history.queueProcessorSplitMaxLevel",
		Description:  "QueueProcessorSplitMaxLevel is the max processing queue level",
		DefaultValue: 2, // 3 levels, start from 0
	},
	TimerTaskBatchSize: DynamicInt{
		KeyName:      "history.timerTaskBatchSize",
		Description:  "TimerTaskBatchSize is batch size for timer processor to process tasks",
		DefaultValue: 100,
	},
	TimerTaskDeleteBatchSize: DynamicInt{
		KeyName:      "history.timerTaskDeleteBatchSize",
		Description:  "TimerTaskDeleteBatchSize is batch size for timer processor to delete timer tasks",
		DefaultValue: 4000,
	},
	TimerProcessorGetFailureRetryCount: DynamicInt{
		KeyName:      "history.timerProcessorGetFailureRetryCount",
		Description:  "TimerProcessorGetFailureRetryCount is retry count for timer processor get failure operation",
		DefaultValue: 5,
	},
	TimerProcessorCompleteTimerFailureRetryCount: DynamicInt{
		KeyName:      "history.timerProcessorCompleteTimerFailureRetryCount",
		Description:  "TimerProcessorCompleteTimerFailureRetryCount is retry count for timer processor complete timer operation",
		DefaultValue: 10,
	},
	TimerProcessorFailoverMaxPollRPS: DynamicInt{
		KeyName:      "history.timerProcessorFailoverMaxPollRPS",
		Description:  "TimerProcessorFailoverMaxPollRPS is max poll rate per second for timer processor",
		DefaultValue: 1,
	},
	TimerProcessorMaxPollRPS: DynamicInt{
		KeyName:      "history.timerProcessorMaxPollRPS",
		Description:  "TimerProcessorMaxPollRPS is max poll rate per second for timer processor",
		DefaultValue: 20,
	},
	TimerProcessorMaxRedispatchQueueSize: DynamicInt{
		KeyName:      "history.timerProcessorMaxRedispatchQueueSize",
		Description:  "TimerProcessorMaxRedispatchQueueSize is the threshold of the number of tasks in the redispatch queue for timer processor",
		DefaultValue: 10000,
	},
	TimerProcessorHistoryArchivalSizeLimit: DynamicInt{
		KeyName:      "history.timerProcessorHistoryArchivalSizeLimit",
		Description:  "TimerProcessorHistoryArchivalSizeLimit is the max history size for inline archival",
		DefaultValue: 500 * 1024,
	},
	TransferTaskBatchSize: DynamicInt{
		KeyName:      "history.transferTaskBatchSize",
		Description:  "TransferTaskBatchSize is batch size for transferQueueProcessor",
		DefaultValue: 100,
	},
	TransferTaskDeleteBatchSize: DynamicInt{
		KeyName:      "history.transferTaskDeleteBatchSize",
		Description:  "TransferTaskDeleteBatchSize is batch size for transferQueueProcessor to delete transfer tasks",
		DefaultValue: 4000,
	},
	TransferProcessorFailoverMaxPollRPS: DynamicInt{
		KeyName:      "history.transferProcessorFailoverMaxPollRPS",
		Description:  "TransferProcessorFailoverMaxPollRPS is max poll rate per second for transferQueueProcessor",
		DefaultValue: 1,
	},
	TransferProcessorMaxPollRPS: DynamicInt{
		KeyName:      "history.transferProcessorMaxPollRPS",
		Description:  "TransferProcessorMaxPollRPS is max poll rate per second for transferQueueProcessor",
		DefaultValue: 20,
	},
	TransferProcessorCompleteTransferFailureRetryCount: DynamicInt{
		KeyName:      "history.transferProcessorCompleteTransferFailureRetryCount",
		Description:  "TransferProcessorCompleteTransferFailureRetryCount is times of retry for failure",
		DefaultValue: 10,
	},
	TransferProcessorMaxRedispatchQueueSize: DynamicInt{
		KeyName:      "history.transferProcessorMaxRedispatchQueueSize",
		Description:  "TransferProcessorMaxRedispatchQueueSize is the threshold of the number of tasks in the redispatch queue for transferQueueProcessor",
		DefaultValue: 10000,
	},
	CrossClusterTaskBatchSize: DynamicInt{
		KeyName:      "history.crossClusterTaskBatchSize",
		Description:  "CrossClusterTaskBatchSize is the batch size for loading cross cluster tasks from persistence in crossClusterQueueProcessor",
		DefaultValue: 100,
	},
	CrossClusterTaskDeleteBatchSize: DynamicInt{
		KeyName:      "history.crossClusterTaskDeleteBatchSize",
		Description:  "CrossClusterTaskDeleteBatchSize is the batch size for deleting cross cluster tasks from persistence in crossClusterQueueProcessor",
		DefaultValue: 4000,
	},
	CrossClusterTaskFetchBatchSize: DynamicInt{
		KeyName:      "history.crossClusterTaskFetchBatchSize",
		Filters:      []Filter{ShardID},
		Description:  "CrossClusterTaskFetchBatchSize is batch size for dispatching cross cluster tasks to target cluster in crossClusterQueueProcessor",
		DefaultValue: 100,
	},
	CrossClusterSourceProcessorMaxPollRPS: DynamicInt{
		KeyName:      "history.crossClusterSourceProcessorMaxPollRPS",
		Description:  "CrossClusterSourceProcessorMaxPollRPS is max poll rate per second for crossClusterQueueProcessor",
		DefaultValue: 20,
	},
	CrossClusterSourceProcessorCompleteTaskFailureRetryCount: DynamicInt{
		KeyName:      "history.crossClusterSourceProcessorCompleteTaskFailureRetryCount",
		Description:  "CrossClusterSourceProcessorCompleteTaskFailureRetryCount is times of retry for failure",
		DefaultValue: 10,
	},
	CrossClusterSourceProcessorMaxRedispatchQueueSize: DynamicInt{
		KeyName:      "history.crossClusterSourceProcessorMaxRedispatchQueueSize",
		Description:  "CrossClusterSourceProcessorMaxRedispatchQueueSize is the threshold of the number of tasks in the redispatch queue for crossClusterQueueProcessor",
		DefaultValue: 10000,
	},
	CrossClusterSourceProcessorMaxPendingTaskSize: DynamicInt{
		KeyName:      "history.crossClusterSourceProcessorMaxPendingTaskSize",
		Description:  "CrossClusterSourceProcessorMaxPendingTaskSize is the threshold of the number of ready for polling tasks in crossClusterQueueProcessor, task loading will be stopped when the number is reached",
		DefaultValue: 500,
	},
	CrossClusterTargetProcessorMaxPendingTasks: DynamicInt{
		KeyName:      "history.crossClusterTargetProcessorMaxPendingTasks",
		Description:  "CrossClusterTargetProcessorMaxPendingTasks is the max number of pending tasks in cross cluster task processor",
		DefaultValue: 200,
	},
	CrossClusterTargetProcessorMaxRetryCount: DynamicInt{
		KeyName:      "history.crossClusterTargetProcessorMaxRetryCount",
		Description:  "CrossClusterTargetProcessorMaxRetryCount is the max number of retries when executing a cross-cluster task in target cluster",
		DefaultValue: 20,
	},
	CrossClusterFetcherParallelism: DynamicInt{
		KeyName:      "history.crossClusterFetcherParallelism",
		Description:  "CrossClusterFetcherParallelism is the number of go routines each cross cluster fetcher use, note there's one cross cluster task fetcher per host per source cluster",
		DefaultValue: 1,
	},
	ReplicatorTaskBatchSize: DynamicInt{
		KeyName:      "history.replicatorTaskBatchSize",
		Description:  "ReplicatorTaskBatchSize is batch size for ReplicatorProcessor",
		DefaultValue: 25,
	},
	ReplicatorTaskDeleteBatchSize: DynamicInt{
		KeyName:      "history.replicatorTaskDeleteBatchSize",
		Description:  "ReplicatorTaskDeleteBatchSize is batch size for ReplicatorProcessor to delete replication tasks",
		DefaultValue: 4000,
	},
	ReplicatorReadTaskMaxRetryCount: DynamicInt{
		KeyName:      "history.replicatorReadTaskMaxRetryCount",
		Description:  "ReplicatorReadTaskMaxRetryCount is the number of read replication task retry time",
		DefaultValue: 3,
	},
	ReplicatorCacheCapacity: DynamicInt{
		KeyName:      "history.replicatorCacheCapacity",
		Description:  "ReplicatorCacheCapacity is the capacity of replication cache in number of tasks",
		DefaultValue: 0,
	},
	ExecutionMgrNumConns: DynamicInt{
		KeyName:      "history.executionMgrNumConns",
		Description:  "ExecutionMgrNumConns is persistence connections number for ExecutionManager",
		DefaultValue: 50,
	},
	HistoryMgrNumConns: DynamicInt{
		KeyName:      "history.historyMgrNumConns",
		Description:  "HistoryMgrNumConns is persistence connections number for HistoryManager",
		DefaultValue: 50,
	},
	MaximumBufferedEventsBatch: DynamicInt{
		KeyName:      "history.maximumBufferedEventsBatch",
		Description:  "MaximumBufferedEventsBatch is max number of buffer event in mutable state",
		DefaultValue: 100,
	},
	MaximumSignalsPerExecution: DynamicInt{
		KeyName:      "history.maximumSignalsPerExecution",
		Filters:      []Filter{DomainName},
		Description:  "MaximumSignalsPerExecution is max number of signals supported by single execution",
		DefaultValue: 10000, // 10K signals should big enough given workflow execution has 200K history lengh limit. It needs to be non-zero to protect continueAsNew from infinit loop
	},
	NumArchiveSystemWorkflows: DynamicInt{
		KeyName:      "history.numArchiveSystemWorkflows",
		Description:  "NumArchiveSystemWorkflows is key for number of archive system workflows running in total",
		DefaultValue: 1000,
	},
	ArchiveRequestRPS: DynamicInt{
		KeyName:      "history.archiveRequestRPS",
		Description:  "ArchiveRequestRPS is the rate limit on the number of archive request per second",
		DefaultValue: 300, // should be much smaller than frontend RPS
	},
	ArchiveInlineHistoryRPS: DynamicInt{
		KeyName:      "history.archiveInlineHistoryRPS",
		Description:  "ArchiveInlineHistoryRPS is the (per instance) rate limit on the number of inline history archival attempts per second",
		DefaultValue: 1000,
	},
	ArchiveInlineHistoryGlobalRPS: DynamicInt{
		KeyName:      "history.archiveInlineHistoryGlobalRPS",
		Description:  "ArchiveInlineHistoryGlobalRPS is the global rate limit on the number of inline history archival attempts per second",
		DefaultValue: 10000,
	},
	ArchiveInlineVisibilityRPS: DynamicInt{
		KeyName:      "history.archiveInlineVisibilityRPS",
		Description:  "ArchiveInlineVisibilityRPS is the (per instance) rate limit on the number of inline visibility archival attempts per second",
		DefaultValue: 1000,
	},
	ArchiveInlineVisibilityGlobalRPS: DynamicInt{
		KeyName:      "history.archiveInlineVisibilityGlobalRPS",
		Description:  "ArchiveInlineVisibilityGlobalRPS is the global rate limit on the number of inline visibility archival attempts per second",
		DefaultValue: 10000,
	},
	HistoryMaxAutoResetPoints: DynamicInt{
		KeyName:      "history.historyMaxAutoResetPoints",
		Filters:      []Filter{DomainName},
		Description:  "HistoryMaxAutoResetPoints is the key for max number of auto reset points stored in mutableState",
		DefaultValue: 20,
	},
	ParentClosePolicyThreshold: DynamicInt{
		KeyName:      "history.parentClosePolicyThreshold",
		Filters:      []Filter{DomainName},
		Description:  "ParentClosePolicyThreshold is decides that parent close policy will be processed by sys workers(if enabled) ifthe number of children greater than or equal to this threshold",
		DefaultValue: 10,
	},
	NumParentClosePolicySystemWorkflows: DynamicInt{
		KeyName:      "history.numParentClosePolicySystemWorkflows",
		Description:  "NumParentClosePolicySystemWorkflows is key for number of parentClosePolicy system workflows running in total",
		DefaultValue: 10,
	},
	HistoryThrottledLogRPS: DynamicInt{
		KeyName:      "history.throttledLogRPS",
		Description:  "HistoryThrottledLogRPS is the rate limit on number of log messages emitted per second for throttled logger",
		DefaultValue: 4,
	},
	DecisionRetryCriticalAttempts: DynamicInt{
		KeyName:      "history.decisionRetryCriticalAttempts",
		Description:  "DecisionRetryCriticalAttempts is decision attempt threshold for logging and emiting metrics",
		DefaultValue: 10,
	},
	DecisionRetryMaxAttempts: DynamicInt{
		KeyName:      "history.decisionRetryMaxAttempts",
		Filters:      []Filter{DomainName},
		Description:  "DecisionRetryMaxAttempts is the max limit for decision retry attempts. 0 indicates infinite number of attempts.",
		DefaultValue: 1000,
	},
	NormalDecisionScheduleToStartMaxAttempts: DynamicInt{
		KeyName:      "history.normalDecisionScheduleToStartMaxAttempts",
		Filters:      []Filter{DomainName},
		Description:  "NormalDecisionScheduleToStartMaxAttempts is the maximum decision attempt for creating a scheduleToStart timeout timer for normal (non-sticky) decision",
		DefaultValue: 0,
	},
	MaxBufferedQueryCount: DynamicInt{
		KeyName:      "history.MaxBufferedQueryCount",
		Description:  "MaxBufferedQueryCount indicates the maximum number of queries which can be buffered at a given time for a single workflow",
		DefaultValue: 1,
	},
	MutableStateChecksumGenProbability: DynamicInt{
		KeyName:      "history.mutableStateChecksumGenProbability",
		Filters:      []Filter{DomainName},
		Description:  "MutableStateChecksumGenProbability is the probability [0-100] that checksum will be generated for mutable state",
		DefaultValue: 0,
	},
	MutableStateChecksumVerifyProbability: DynamicInt{
		KeyName:      "history.mutableStateChecksumVerifyProbability",
		Filters:      []Filter{DomainName},
		Description:  "MutableStateChecksumVerifyProbability is the probability [0-100] that checksum will be verified for mutable state",
		DefaultValue: 0,
	},
	MaxActivityCountDispatchByDomain: DynamicInt{
		KeyName:      "history.maxActivityCountDispatchByDomain",
		Description:  "MaxActivityCountDispatchByDomain max # of activity tasks to dispatch to matching before creating transfer tasks. This is an performance optimization to skip activity scheduling efforts.",
		DefaultValue: 0,
	},
	ReplicationTaskFetcherParallelism: DynamicInt{
		KeyName:      "history.ReplicationTaskFetcherParallelism",
		Description:  "ReplicationTaskFetcherParallelism determines how many go routines we spin up for fetching tasks",
		DefaultValue: 1,
	},
	ReplicationTaskProcessorErrorRetryMaxAttempts: DynamicInt{
		KeyName:      "history.ReplicationTaskProcessorErrorRetryMaxAttempts",
		Filters:      []Filter{ShardID},
		Description:  "ReplicationTaskProcessorErrorRetryMaxAttempts is the max retry attempts for applying replication tasks",
		DefaultValue: 10,
	},
	WorkflowIDExternalRPS: DynamicInt{
		KeyName:      "history.workflowIDExternalRPS",
		Filters:      []Filter{DomainName},
		Description:  "WorkflowIDExternalRPS is the rate limit per workflowID for external calls",
		DefaultValue: UnlimitedRPS,
	},
	WorkflowIDInternalRPS: DynamicInt{
		KeyName:      "history.workflowIDInternalRPS",
		Filters:      []Filter{DomainName},
		Description:  "WorkflowIDInternalRPS is the rate limit per workflowID for internal calls",
		DefaultValue: UnlimitedRPS,
	},
	WorkerPersistenceMaxQPS: DynamicInt{
		KeyName:      "worker.persistenceMaxQPS",
		Description:  "WorkerPersistenceMaxQPS is the max qps worker host can query DB",
		DefaultValue: 500,
	},
	WorkerPersistenceGlobalMaxQPS: DynamicInt{
		KeyName:      "worker.persistenceGlobalMaxQPS",
		Description:  "WorkerPersistenceGlobalMaxQPS is the max qps worker cluster can query DB",
		DefaultValue: 0,
	},
	WorkerIndexerConcurrency: DynamicInt{
		KeyName:      "worker.indexerConcurrency",
		Description:  "WorkerIndexerConcurrency is the max concurrent messages to be processed at any given time",
		DefaultValue: 1000,
	},
	WorkerESProcessorNumOfWorkers: DynamicInt{
		KeyName:      "worker.ESProcessorNumOfWorkers",
		Description:  "WorkerESProcessorNumOfWorkers is num of workers for esProcessor",
		DefaultValue: 1,
	},
	WorkerESProcessorBulkActions: DynamicInt{
		KeyName:      "worker.ESProcessorBulkActions",
		Description:  "WorkerESProcessorBulkActions is max number of requests in bulk for esProcessor",
		DefaultValue: 1000,
	},
	WorkerESProcessorBulkSize: DynamicInt{
		KeyName:      "worker.ESProcessorBulkSize",
		Description:  "WorkerESProcessorBulkSize is max total size of bulk in bytes for esProcessor",
		DefaultValue: 2 << 24, // 16MB
	},
	WorkerArchiverConcurrency: DynamicInt{
		KeyName:      "worker.ArchiverConcurrency",
		Description:  "WorkerArchiverConcurrency is controls the number of coroutines handling archival work per archival workflow",
		DefaultValue: 50,
	},
	WorkerArchivalsPerIteration: DynamicInt{
		KeyName:      "worker.ArchivalsPerIteration",
		Description:  "WorkerArchivalsPerIteration is controls the number of archivals handled in each iteration of archival workflow",
		DefaultValue: 1000,
	},
	WorkerThrottledLogRPS: DynamicInt{
		KeyName:      "worker.throttledLogRPS",
		Description:  "WorkerThrottledLogRPS is the rate limit on number of log messages emitted per second for throttled logger",
		DefaultValue: 20,
	},
	ScannerPersistenceMaxQPS: DynamicInt{
		KeyName:      "worker.scannerPersistenceMaxQPS",
		Description:  "ScannerPersistenceMaxQPS is the maximum rate of persistence calls from worker.Scanner",
		DefaultValue: 5,
	},
	ScannerGetOrphanTasksPageSize: DynamicInt{
		KeyName:      "worker.scannerGetOrphanTasksPageSize",
		Description:  "ScannerGetOrphanTasksPageSize is the maximum number of orphans to delete in one batch",
		DefaultValue: 1000,
	},
	ScannerBatchSizeForTasklistHandler: DynamicInt{
		KeyName:      "worker.scannerBatchSizeForTasklistHandler",
		Description:  "ScannerBatchSizeForTasklistHandler is for: 1. max number of tasks to query per call(get tasks for tasklist) in the scavenger handler. 2. The scavenger then uses the return to decide if a tasklist can be deleted. It's better to keep it a relatively high number to let it be more efficient.",
		DefaultValue: 1000,
	},
	ScannerMaxTasksProcessedPerTasklistJob: DynamicInt{
		KeyName:      "worker.scannerMaxTasksProcessedPerTasklistJob",
		Description:  "ScannerMaxTasksProcessedPerTasklistJob is the number of tasks to process for a tasklist in each workflow run",
		DefaultValue: 256,
	},
	ConcreteExecutionsScannerConcurrency: DynamicInt{
		KeyName:      "worker.executionsScannerConcurrency",
		Description:  "ConcreteExecutionsScannerConcurrency indicates the concurrency of concrete execution scanner",
		DefaultValue: 25,
	},
	ConcreteExecutionsScannerBlobstoreFlushThreshold: DynamicInt{
		KeyName:      "worker.executionsScannerBlobstoreFlushThreshold",
		Description:  "ConcreteExecutionsScannerBlobstoreFlushThreshold indicates the flush threshold of blobstore in concrete execution scanner",
		DefaultValue: 100,
	},
	ConcreteExecutionsScannerActivityBatchSize: DynamicInt{
		KeyName:      "worker.executionsScannerActivityBatchSize",
		Description:  "ConcreteExecutionsScannerActivityBatchSize indicates the batch size of scanner activities",
		DefaultValue: 25,
	},
	ConcreteExecutionsScannerPersistencePageSize: DynamicInt{
		KeyName:      "worker.executionsScannerPersistencePageSize",
		Description:  "ConcreteExecutionsScannerPersistencePageSize indicates the page size of execution persistence fetches in concrete execution scanner",
		DefaultValue: 1000,
	},
	CurrentExecutionsScannerConcurrency: DynamicInt{
		KeyName:      "worker.currentExecutionsConcurrency",
		Description:  "CurrentExecutionsScannerConcurrency indicates the concurrency of current executions scanner",
		DefaultValue: 25,
	},
	CurrentExecutionsScannerBlobstoreFlushThreshold: DynamicInt{
		KeyName:      "worker.currentExecutionsBlobstoreFlushThreshold",
		Description:  "CurrentExecutionsScannerBlobstoreFlushThreshold indicates the flush threshold of blobstore in current executions scanner",
		DefaultValue: 100,
	},
	CurrentExecutionsScannerActivityBatchSize: DynamicInt{
		KeyName:      "worker.currentExecutionsActivityBatchSize",
		Description:  "CurrentExecutionsScannerActivityBatchSize indicates the batch size of scanner activities",
		DefaultValue: 25,
	},
	CurrentExecutionsScannerPersistencePageSize: DynamicInt{
		KeyName:      "worker.currentExecutionsPersistencePageSize",
		Description:  "CurrentExecutionsScannerPersistencePageSize indicates the page size of execution persistence fetches in current executions scanner",
		DefaultValue: 1000,
	},
	TimersScannerConcurrency: DynamicInt{
		KeyName:      "worker.timersScannerConcurrency",
		Description:  "TimersScannerConcurrency is the concurrency of timers scanner",
		DefaultValue: 5,
	},
	TimersScannerPersistencePageSize: DynamicInt{
		KeyName:      "worker.timersScannerPersistencePageSize",
		Description:  "TimersScannerPersistencePageSize is the page size of timers persistence fetches in timers scanner",
		DefaultValue: 1000,
	},
	TimersScannerBlobstoreFlushThreshold: DynamicInt{
		KeyName:      "worker.timersScannerBlobstoreFlushThreshold",
		Description:  "TimersScannerBlobstoreFlushThreshold is threshold to flush blob store",
		DefaultValue: 100,
	},
	TimersScannerActivityBatchSize: DynamicInt{
		KeyName:      "worker.timersScannerActivityBatchSize",
		Description:  "TimersScannerActivityBatchSize is TimersScannerActivityBatchSize",
		DefaultValue: 25,
	},
	TimersScannerPeriodStart: DynamicInt{
		KeyName:      "worker.timersScannerPeriodStart",
		Description:  "TimersScannerPeriodStart is interval start for fetching scheduled timers",
		DefaultValue: 24,
	},
	TimersScannerPeriodEnd: DynamicInt{
		KeyName:      "worker.timersScannerPeriodEnd",
		Description:  "TimersScannerPeriodEnd is interval end for fetching scheduled timers",
		DefaultValue: 3,
	},
	ESAnalyzerMaxNumDomains: DynamicInt{
		KeyName:      "worker.ESAnalyzerMaxNumDomains",
		Description:  "ESAnalyzerMaxNumDomains defines how many domains to check",
		DefaultValue: 500,
	},
	ESAnalyzerMaxNumWorkflowTypes: DynamicInt{
		KeyName:      "worker.ESAnalyzerMaxNumWorkflowTypes",
		Description:  "ESAnalyzerMaxNumWorkflowTypes defines how many workflow types to check per domain",
		DefaultValue: 100,
	},
	ESAnalyzerNumWorkflowsToRefresh: DynamicInt{
		KeyName:      "worker.ESAnalyzerNumWorkflowsToRefresh",
		Description:  "ESAnalyzerNumWorkflowsToRefresh controls how many workflows per workflow type should be refreshed per workflow type",
		DefaultValue: 100,
	},
	ESAnalyzerMinNumWorkflowsForAvg: DynamicInt{
		KeyName:      "worker.ESAnalyzerMinNumWorkflowsForAvg",
		Description:  "ESAnalyzerMinNumWorkflowsForAvg controls how many workflows to have at least to rely on workflow run time avg per type",
		DefaultValue: 100,
	},
	VisibilityArchivalQueryMaxRangeInDays: DynamicInt{
		KeyName:      "frontend.visibilityArchivalQueryMaxRangeInDays",
		Description:  "VisibilityArchivalQueryMaxRangeInDays is the maximum number of days for a visibility archival query",
		DefaultValue: 60,
	},
	VisibilityArchivalQueryMaxQPS: DynamicInt{
		KeyName:      "frontend.visibilityArchivalQueryMaxQPS",
		Description:  "VisibilityArchivalQueryMaxQPS is the timeout for a visibility archival query",
		DefaultValue: 1,
	},
	WorkflowDeletionJitterRange: DynamicInt{
		KeyName:      "system.workflowDeletionJitterRange",
		Description:  "WorkflowDeletionJitterRange defines the duration in minutes for workflow close tasks jittering",
		DefaultValue: 60,
	},
	SampleLoggingRate: DynamicInt{
		KeyName:      "system.sampleLoggingRate",
		Description:  "The rate for which sampled logs are logged at. 100 means 1/100 is logged",
		DefaultValue: 100,
	},
	LargeShardHistorySizeMetricThreshold: DynamicInt{
		KeyName:      "system.largeShardHistorySizeMetricThreshold",
		Description:  "defines the threshold for what consititutes a large history size to alert on, default is 10mb",
		DefaultValue: 10485760,
	},
	LargeShardHistoryEventMetricThreshold: DynamicInt{
		KeyName:      "system.largeShardHistoryEventMetricThreshold",
		Description:  "defines the threshold for what consititutes a large history event length to alert on, default is 50k",
		DefaultValue: 50 * 1024,
	},
	LargeShardHistoryBlobMetricThreshold: DynamicInt{
		KeyName:      "system.largeShardHistoryBlobMetricThreshold",
		Description:  "defines the threshold for what consititutes a large history blob write to alert on, default is 1/4mb",
		DefaultValue: 262144,
	},
	IsolationGroupStateUpdateRetryAttempts: DynamicInt{
		KeyName:      "system.isolationGroupStateUpdateRetryAttempts",
		Description:  "The number of attempts to push Isolation group configuration to the config store",
		DefaultValue: 2,
	},
}

var BoolKeys = map[BoolKey]DynamicBool{
	TestGetBoolPropertyKey: DynamicBool{
		KeyName:      "testGetBoolPropertyKey",
		Description:  "",
		DefaultValue: false,
	},
	TestGetBoolPropertyFilteredByDomainIDKey: DynamicBool{
		KeyName:      "testGetBoolPropertyFilteredByDomainIDKey",
		Description:  "",
		DefaultValue: false,
	},
	TestGetBoolPropertyFilteredByTaskListInfoKey: DynamicBool{
		KeyName:      "testGetBoolPropertyFilteredByTaskListInfoKey",
		Description:  "",
		DefaultValue: false,
	},
	EnableVisibilitySampling: DynamicBool{
		KeyName:      "system.enableVisibilitySampling",
		Description:  "EnableVisibilitySampling is key for enable visibility sampling for basic(DB based) visibility",
		DefaultValue: false, // ...
		Filters:      nil,
	},
	EnableReadFromClosedExecutionV2: DynamicBool{
		KeyName:      "system.enableReadFromClosedExecutionV2",
		Description:  "EnableReadFromClosedExecutionV2 is key for enable read from cadence_visibility.closed_executions_v2",
		DefaultValue: false,
	},
	EnableReadVisibilityFromES: DynamicBool{
		KeyName:      "system.enableReadVisibilityFromES",
		Filters:      []Filter{DomainName},
		Description:  "EnableReadVisibilityFromES is key for enable read from elastic search or db visibility, usually using with AdvancedVisibilityWritingMode for seamless migration from db visibility to advanced visibility",
		DefaultValue: true,
	},
	EnableReadVisibilityFromPinot: DynamicBool{
		KeyName:      "system.enableReadVisibilityFromPinot",
		Filters:      []Filter{DomainName},
		Description:  "EnableReadVisibilityFromPinot is key for enable read from pinot or db visibility, usually using with AdvancedVisibilityWritingMode for seamless migration from db visibility to advanced visibility",
		DefaultValue: true,
	},
	EnableLogCustomerQueryParameter: DynamicBool{
		KeyName:      "system.enableLogCustomerQueryParameter",
		Filters:      []Filter{DomainName},
		Description:  "EnableLogCustomerQueryParameter is key for enable log customer query parameters",
		DefaultValue: false,
	},
	EmitShardDiffLog: DynamicBool{
		KeyName:      "history.emitShardDiffLog",
		Description:  "EmitShardDiffLog is whether emit the shard diff log",
		DefaultValue: false,
	},
	EnableRecordWorkflowExecutionUninitialized: DynamicBool{
		KeyName:      "history.enableRecordWorkflowExecutionUninitialized",
		Description:  "EnableRecordWorkflowExecutionUninitialized enables record workflow execution uninitialized state in ElasticSearch",
		DefaultValue: false,
	},
	DisableListVisibilityByFilter: DynamicBool{
		KeyName:      "frontend.disableListVisibilityByFilter",
		Filters:      []Filter{DomainName},
		Description:  "DisableListVisibilityByFilter is config to disable list open/close workflow using filter",
		DefaultValue: false,
	},
	EnableReadFromHistoryArchival: DynamicBool{
		KeyName:      "system.enableReadFromHistoryArchival",
		Description:  "EnableReadFromHistoryArchival is key for enabling reading history from archival store",
		DefaultValue: true,
	},
	EnableReadFromVisibilityArchival: DynamicBool{
		KeyName:      "system.enableReadFromVisibilityArchival",
		Description:  "EnableReadFromVisibilityArchival is key for enabling reading visibility from archival store to override the value from static config.",
		DefaultValue: true,
	},
	EnableDomainNotActiveAutoForwarding: DynamicBool{
		KeyName:      "system.enableDomainNotActiveAutoForwarding",
		Filters:      []Filter{DomainName},
		Description:  "EnableDomainNotActiveAutoForwarding decides requests form which domain will be forwarded to active cluster if domain is not active in current cluster. Only when selected-api-forwarding or all-domain-apis-forwarding is the policy in ClusterRedirectionPolicy(in static config). If the policy is noop(default) this flag is not doing anything.",
		DefaultValue: true,
	},
	EnableGracefulFailover: DynamicBool{
		KeyName:      "system.enableGracefulFailover",
		Description:  "EnableGracefulFailover is whether enabling graceful failover",
		DefaultValue: true,
	},
	DisallowQuery: DynamicBool{
		KeyName:      "system.disallowQuery",
		Filters:      []Filter{DomainName},
		Description:  "DisallowQuery is the key to disallow query for a domain",
		DefaultValue: false,
	},
	EnableDebugMode: DynamicBool{
		KeyName:      "system.enableDebugMode",
		Description:  "EnableDebugMode is for enabling debugging components, logs and metrics",
		DefaultValue: false,
	},
	EnableGRPCOutbound: DynamicBool{
		KeyName:      "system.enableGRPCOutbound",
		Description:  "EnableGRPCOutbound is the key for enabling outbound GRPC traffic",
		DefaultValue: true,
	},
	EnableSQLAsyncTransaction: DynamicBool{
		KeyName:      "system.enableSQLAsyncTransaction",
		Description:  "EnableSQLAsyncTransaction is the key for enabling async transaction",
		DefaultValue: false,
	},
	EnableClientVersionCheck: DynamicBool{
		KeyName:      "frontend.enableClientVersionCheck",
		Description:  "EnableClientVersionCheck is enables client version check for frontend",
		DefaultValue: false,
	},
	SendRawWorkflowHistory: DynamicBool{
		KeyName:      "frontend.sendRawWorkflowHistory",
		Filters:      []Filter{DomainName},
		Description:  "SendRawWorkflowHistory is whether to enable raw history retrieving",
		DefaultValue: false,
	},
	FrontendEmitSignalNameMetricsTag: DynamicBool{
		KeyName:      "frontend.emitSignalNameMetricsTag",
		Filters:      []Filter{DomainName},
		Description:  "FrontendEmitSignalNameMetricsTag enables emitting signal name tag in metrics in frontend client",
		DefaultValue: false,
	},
	EnableQueryAttributeValidation: DynamicBool{
		KeyName:      "frontend.enableQueryAttributeValidation",
		Description:  "EnableQueryAttributeValidation enables validation of queries' search attributes against the dynamic config whitelist",
		DefaultValue: true,
	},
	MatchingEnableSyncMatch: DynamicBool{
		KeyName:      "matching.enableSyncMatch",
		Filters:      []Filter{DomainName, TaskListName, TaskType},
		Description:  "MatchingEnableSyncMatch is to enable sync match",
		DefaultValue: true,
	},
	MatchingEnableTaskInfoLogByDomainID: DynamicBool{
		KeyName:      "matching.enableTaskInfoLogByDomainID",
		Filters:      []Filter{DomainID},
		Description:  "MatchingEnableTaskInfoLogByDomainID is enables info level logs for decision/activity task based on the request domainID",
		DefaultValue: false,
	},
	EventsCacheGlobalEnable: DynamicBool{
		KeyName:      "history.eventsCacheGlobalEnable",
		Description:  "EventsCacheGlobalEnable is enables global cache over all history shards",
		DefaultValue: false,
	},
	QueueProcessorEnableSplit: DynamicBool{
		KeyName:      "history.queueProcessorEnableSplit",
		Description:  "QueueProcessorEnableSplit indicates whether processing queue split policy should be enabled",
		DefaultValue: false,
	},
	QueueProcessorEnableRandomSplitByDomainID: DynamicBool{
		KeyName:      "history.queueProcessorEnableRandomSplitByDomainID",
		Filters:      []Filter{DomainID},
		Description:  "QueueProcessorEnableRandomSplitByDomainID indicates whether random queue split policy should be enabled for a domain",
		DefaultValue: false,
	},
	QueueProcessorEnablePendingTaskSplitByDomainID: DynamicBool{
		KeyName:      "history.queueProcessorEnablePendingTaskSplitByDomainID",
		Filters:      []Filter{DomainID},
		Description:  "ueueProcessorEnablePendingTaskSplitByDomainID indicates whether pending task split policy should be enabled",
		DefaultValue: false,
	},
	QueueProcessorEnableStuckTaskSplitByDomainID: DynamicBool{
		KeyName:      "history.queueProcessorEnableStuckTaskSplitByDomainID",
		Filters:      []Filter{DomainID},
		Description:  "QueueProcessorEnableStuckTaskSplitByDomainID indicates whether stuck task split policy should be enabled",
		DefaultValue: false,
	},
	QueueProcessorEnablePersistQueueStates: DynamicBool{
		KeyName:      "history.queueProcessorEnablePersistQueueStates",
		Description:  "QueueProcessorEnablePersistQueueStates indicates whether processing queue states should be persisted",
		DefaultValue: true,
	},
	QueueProcessorEnableLoadQueueStates: DynamicBool{
		KeyName:      "history.queueProcessorEnableLoadQueueStates",
		Description:  "QueueProcessorEnableLoadQueueStates indicates whether processing queue states should be loaded",
		DefaultValue: true,
	},
	QueueProcessorEnableGracefulSyncShutdown: DynamicBool{
		KeyName:      "history.queueProcessorEnableGracefulSyncShutdown",
		Description:  "QueueProcessorEnableGracefulSyncShutdown indicates whether processing queue should be shutdown gracefully & synchronously",
		DefaultValue: false,
	},
	ReplicationTaskFetcherEnableGracefulSyncShutdown: DynamicBool{
		KeyName:      "history.replicationTaskFetcherEnableGracefulSyncShutdown",
		Description:  "ReplicationTaskFetcherEnableGracefulSyncShutdown is whether we should gracefully drain replication task fetcher on shutdown",
		DefaultValue: false,
	},
	TransferProcessorEnableValidator: DynamicBool{
		KeyName:      "history.transferProcessorEnableValidator",
		Description:  "TransferProcessorEnableValidator is whether validator should be enabled for transferQueueProcessor",
		DefaultValue: false,
	},
	EnableAdminProtection: DynamicBool{
		KeyName:      "history.enableAdminProtection",
		Description:  "EnableAdminProtection is whether to enable admin checking",
		DefaultValue: false,
	},
	EnableParentClosePolicy: DynamicBool{
		KeyName:      "history.enableParentClosePolicy",
		Filters:      []Filter{DomainName},
		Description:  "EnableParentClosePolicy is whether to  ParentClosePolicy",
		DefaultValue: true,
	},
	EnableDropStuckTaskByDomainID: DynamicBool{
		KeyName:      "history.DropStuckTaskByDomain",
		Filters:      []Filter{DomainID},
		Description:  "EnableDropStuckTaskByDomainID is whether stuck timer/transfer task should be dropped for a domain",
		DefaultValue: false,
	},
	EnableConsistentQuery: DynamicBool{
		KeyName:      "history.EnableConsistentQuery",
		Description:  "EnableConsistentQuery indicates if consistent query is enabled for the cluster",
		DefaultValue: true,
	},
	EnableConsistentQueryByDomain: DynamicBool{
		KeyName:      "history.EnableConsistentQueryByDomain",
		Filters:      []Filter{DomainName},
		Description:  "EnableConsistentQueryByDomain indicates if consistent query is enabled for a domain",
		DefaultValue: false,
	},
	EnableCrossClusterEngine: DynamicBool{
		KeyName:      "history.enableCrossClusterEngine",
		Description:  "an overall toggle for the cross-cluster domain feature",
		DefaultValue: false,
	},
	EnableCrossClusterOperationsForDomain: DynamicBool{
		KeyName:      "history.enableCrossClusterOperations",
		Filters:      []Filter{DomainName},
		Description:  "EnableCrossClusterOperationsForDomain indicates if cross cluster operations can be scheduled for a domain",
		DefaultValue: false,
	},
	EnableHistoryCorruptionCheck: DynamicBool{
		KeyName:      "history.enableHistoryCorruptionCheck",
		Filters:      []Filter{DomainName},
		Description:  "EnableHistoryCorruptionCheck enables additional sanity check for corrupted history. This allows early catches of DB corruptions but potiantally increased latency.",
		DefaultValue: false,
	},
	EnableActivityLocalDispatchByDomain: DynamicBool{
		KeyName:      "history.enableActivityLocalDispatchByDomain",
		Filters:      []Filter{DomainName},
		Description:  "EnableActivityLocalDispatchByDomain is allows worker to dispatch activity tasks through local tunnel after decisions are made. This is an performance optimization to skip activity scheduling efforts",
		DefaultValue: true,
	},
	HistoryEnableTaskInfoLogByDomainID: DynamicBool{
		KeyName:      "history.enableTaskInfoLogByDomainID",
		Filters:      []Filter{DomainID},
		Description:  "HistoryEnableTaskInfoLogByDomainID is enables info level logs for decision/activity task based on the request domainID",
		DefaultValue: false,
	},
	EnableReplicationTaskGeneration: DynamicBool{
		KeyName:      "history.enableReplicationTaskGeneration",
		Filters:      []Filter{DomainID, WorkflowID},
		Description:  "EnableReplicationTaskGeneration is the flag to control replication generation",
		DefaultValue: true,
	},
	UseNewInitialFailoverVersion: DynamicBool{
		KeyName:      "history.useNewInitialFailoverVersion",
		Description:  "use the minInitialFailover version",
		DefaultValue: false,
	},
	AllowArchivingIncompleteHistory: DynamicBool{
		KeyName:      "worker.AllowArchivingIncompleteHistory",
		Description:  "AllowArchivingIncompleteHistory will continue on when seeing some error like history mutated(usually caused by database consistency issues)",
		DefaultValue: false,
	},
	EnableCleaningOrphanTaskInTasklistScavenger: DynamicBool{
		KeyName:      "worker.enableCleaningOrphanTaskInTasklistScavenger",
		Description:  "EnableCleaningOrphanTaskInTasklistScavenger indicates if enabling the scanner to clean up orphan tasks",
		DefaultValue: false,
	},
	TaskListScannerEnabled: DynamicBool{
		KeyName:      "worker.taskListScannerEnabled",
		Description:  "TaskListScannerEnabled indicates if task list scanner should be started as part of worker.Scanner",
		DefaultValue: true,
	},
	HistoryScannerEnabled: DynamicBool{
		KeyName:      "worker.historyScannerEnabled",
		Description:  "HistoryScannerEnabled indicates if history scanner should be started as part of worker.Scanner",
		DefaultValue: false,
	},
	ConcreteExecutionsScannerEnabled: DynamicBool{
		KeyName:      "worker.executionsScannerEnabled",
		Description:  "ConcreteExecutionsScannerEnabled indicates if executions scanner should be started as part of worker.Scanner",
		DefaultValue: false,
	},
	ConcreteExecutionsScannerInvariantCollectionMutableState: DynamicBool{
		KeyName:      "worker.executionsScannerInvariantCollectionMutableState",
		Description:  "ConcreteExecutionsScannerInvariantCollectionMutableState indicates if mutable state invariant checks should be run",
		DefaultValue: true,
	},
	ConcreteExecutionsScannerInvariantCollectionHistory: DynamicBool{
		KeyName:      "worker.executionsScannerInvariantCollectionHistory",
		Description:  "ConcreteExecutionsScannerInvariantCollectionHistory indicates if history invariant checks should be run",
		DefaultValue: true,
	},
	ConcreteExecutionsScannerInvariantCollectionStale: DynamicBool{
		KeyName:      "worker.executionsScannerInvariantCollectionStale",
		Description:  "ConcreteExecutionsScannerInvariantCollectionStale indicates if the stale-workflow invariant should be run",
		DefaultValue: false, // may be enabled after further verification, but for now it's a bit too risky to enable by default
	},
	ConcreteExecutionsFixerInvariantCollectionMutableState: DynamicBool{
		KeyName:      "worker.executionsFixerInvariantCollectionMutableState",
		Description:  "ConcreteExecutionsFixerInvariantCollectionMutableState indicates if mutable state invariant checks should be run",
		DefaultValue: true,
	},
	ConcreteExecutionsFixerInvariantCollectionHistory: DynamicBool{
		KeyName:      "worker.executionsFixerInvariantCollectionHistory",
		Description:  "ConcreteExecutionsFixerInvariantCollectionHistory indicates if history invariant checks should be run",
		DefaultValue: true,
	},
	ConcreteExecutionsFixerInvariantCollectionStale: DynamicBool{
		KeyName:      "worker.executionsFixerInvariantCollectionStale",
		Description:  "ConcreteExecutionsFixerInvariantCollectionStale indicates if the stale-workflow invariant should be run",
		DefaultValue: false, // may be enabled after further verification, but for now it's a bit too risky to enable by default
	},
	CurrentExecutionsScannerEnabled: DynamicBool{
		KeyName:      "worker.currentExecutionsScannerEnabled",
		Description:  "CurrentExecutionsScannerEnabled indicates if current executions scanner should be started as part of worker.Scanner",
		DefaultValue: false,
	},
	CurrentExecutionsScannerInvariantCollectionHistory: DynamicBool{
		KeyName:      "worker.currentExecutionsScannerInvariantCollectionHistory",
		Description:  "CurrentExecutionsScannerInvariantCollectionHistory indicates if history invariant checks should be run",
		DefaultValue: true,
	},
	CurrentExecutionsScannerInvariantCollectionMutableState: DynamicBool{
		KeyName:      "worker.currentExecutionsInvariantCollectionMutableState",
		Description:  "CurrentExecutionsScannerInvariantCollectionMutableState indicates if mutable state invariant checks should be run",
		DefaultValue: true,
	},
	EnableBatcher: DynamicBool{
		KeyName:      "worker.enableBatcher",
		Description:  "EnableBatcher is decides whether start batcher in our worker",
		DefaultValue: true,
	},
	EnableParentClosePolicyWorker: DynamicBool{
		KeyName:      "system.enableParentClosePolicyWorker",
		Description:  "EnableParentClosePolicyWorker decides whether or not enable system workers for processing parent close policy task",
		DefaultValue: true,
	},
	EnableESAnalyzer: DynamicBool{
		KeyName:      "system.enableESAnalyzer",
		Description:  "EnableESAnalyzer decides whether to enable system workers for processing ElasticSearch Analyzer",
		DefaultValue: false,
	},
	EnableStickyQuery: DynamicBool{
		KeyName:      "system.enableStickyQuery",
		Filters:      []Filter{DomainName},
		Description:  "EnableStickyQuery indicates if sticky query should be enabled per domain",
		DefaultValue: true,
	},
	EnableFailoverManager: DynamicBool{
		KeyName:      "system.enableFailoverManager",
		Description:  "EnableFailoverManager indicates if failover manager is enabled",
		DefaultValue: true,
	},
	EnableWorkflowShadower: DynamicBool{
		KeyName:      "system.enableWorkflowShadower",
		Description:  "EnableWorkflowShadower indicates if workflow shadower is enabled",
		DefaultValue: true,
	},
	ConcreteExecutionFixerDomainAllow: DynamicBool{
		KeyName:      "worker.concreteExecutionFixerDomainAllow",
		Filters:      []Filter{DomainName},
		Description:  "ConcreteExecutionFixerDomainAllow is which domains are allowed to be fixed by concrete fixer workflow",
		DefaultValue: false,
	},
	CurrentExecutionFixerDomainAllow: DynamicBool{
		KeyName:      "worker.currentExecutionFixerDomainAllow",
		Filters:      []Filter{DomainName},
		Description:  "CurrentExecutionFixerDomainAllow is which domains are allowed to be fixed by current fixer workflow",
		DefaultValue: false,
	},
	TimersScannerEnabled: DynamicBool{
		KeyName:      "worker.timersScannerEnabled",
		Description:  "TimersScannerEnabled is if timers scanner should be started as part of worker.Scanner",
		DefaultValue: false,
	},
	TimersFixerEnabled: DynamicBool{
		KeyName:      "worker.timersFixerEnabled",
		Description:  "TimersFixerEnabled is if timers fixer should be started as part of worker.Scanner",
		DefaultValue: false,
	},
	TimersFixerDomainAllow: DynamicBool{
		KeyName:      "worker.timersFixerDomainAllow",
		Filters:      []Filter{DomainName},
		Description:  "TimersFixerDomainAllow is which domains are allowed to be fixed by timer fixer workflow",
		DefaultValue: false,
	},
	ConcreteExecutionFixerEnabled: DynamicBool{
		KeyName:      "worker.concreteExecutionFixerEnabled",
		Description:  "ConcreteExecutionFixerEnabled is if concrete execution fixer workflow is enabled",
		DefaultValue: false,
	},
	CurrentExecutionFixerEnabled: DynamicBool{
		KeyName:      "worker.currentExecutionFixerEnabled",
		Description:  "CurrentExecutionFixerEnabled is if current execution fixer workflow is enabled",
		DefaultValue: false,
	},
	EnableAuthorization: DynamicBool{
		KeyName:      "system.enableAuthorization",
		Description:  "EnableAuthorization is the key to enable authorization for a domain, only for extension binary:",
		DefaultValue: false,
	},
	EnableTasklistIsolation: DynamicBool{
		KeyName:      "system.enableTasklistIsolation",
		Description:  "EnableTasklistIsolation is a feature to enable isolation-groups for a domain. Should not be enabled without a deep understanding of this feature",
		DefaultValue: false,
	},
	EnableServiceAuthorization: DynamicBool{
		KeyName:      "system.enableServiceAuthorization",
		Description:  "EnableServiceAuthorization is the key to enable authorization for a service, only for extension binary:",
		DefaultValue: false,
	},
	EnableServiceAuthorizationLogOnly: DynamicBool{
		KeyName:      "system.enableServiceAuthorizationLogOnly",
		Description:  "EnableServiceAuthorizationLogOnly is the key to enable authorization logging for a service, only for extension binary:",
		DefaultValue: false,
	},
	ESAnalyzerPause: DynamicBool{
		KeyName:      "worker.ESAnalyzerPause",
		Description:  "ESAnalyzerPause defines if we want to dynamically pause the analyzer workflow",
		DefaultValue: false,
	},
	EnableArchivalCompression: DynamicBool{
		KeyName:      "worker.EnableArchivalCompression",
		Description:  "EnableArchivalCompression indicates whether blobs are compressed before they are archived",
		DefaultValue: false,
	},
	ESAnalyzerEnableAvgDurationBasedChecks: DynamicBool{
		KeyName:      "worker.ESAnalyzerEnableAvgDurationBasedChecks",
		Description:  "ESAnalyzerEnableAvgDurationBasedChecks controls if we want to enable avg duration based task refreshes",
		DefaultValue: false,
	},
	Lockdown: DynamicBool{
		KeyName:      "system.Lockdown",
		Description:  "Lockdown defines if we want to allow failovers of domains to this cluster",
		DefaultValue: false,
	},
	EnablePendingActivityValidation: DynamicBool{
		KeyName:      "limit.pendingActivityCount.enabled",
		Description:  "Enables pending activity count limiting/validation",
		DefaultValue: false,
	},
	EnableCassandraAllConsistencyLevelDelete: DynamicBool{
		KeyName:      "system.enableCassandraAllConsistencyLevelDelete",
		Description:  "Uses all consistency level for Cassandra delete operations",
		DefaultValue: false,
	},
	EnableShardIDMetrics: DynamicBool{
		KeyName:      "system.enableShardIDMetrics",
		Description:  "Enable shardId metrics in persistence client",
		DefaultValue: true,
	},
	EnableTimerDebugLogByDomainID: DynamicBool{
		KeyName:      "history.enableTimerDebugLogByDomainID",
		Filters:      []Filter{DomainID},
		Description:  "Enable log for debugging timer task issue by domain",
		DefaultValue: false,
	},
	EnableTaskVal: DynamicBool{
		KeyName:      "system.enableTaskVal",
		Description:  "Enable TaskValidation",
		DefaultValue: false,
	},
	WorkflowIDCacheEnabled: DynamicBool{
		KeyName:      "history.workflowIDCacheEnabled",
		Filters:      []Filter{DomainName},
		Description:  "WorkflowIDCacheEnabled is the key to enable/disable caching of workflowID specific information",
		DefaultValue: false,
	},
}

var FloatKeys = map[FloatKey]DynamicFloat{
	TestGetFloat64PropertyKey: DynamicFloat{
		KeyName:      "testGetFloat64PropertyKey",
		Description:  "",
		DefaultValue: 0,
	},
	PersistenceErrorInjectionRate: DynamicFloat{
		KeyName:      "system.persistenceErrorInjectionRate",
		Description:  "PersistenceErrorInjectionRate is rate for injecting random error in persistence",
		DefaultValue: 0,
	},
	AdminErrorInjectionRate: DynamicFloat{
		KeyName:      "admin.errorInjectionRate",
		Description:  "dminErrorInjectionRate is the rate for injecting random error in admin client",
		DefaultValue: 0,
	},
	DomainFailoverRefreshTimerJitterCoefficient: DynamicFloat{
		KeyName:      "frontend.domainFailoverRefreshTimerJitterCoefficient",
		Description:  "DomainFailoverRefreshTimerJitterCoefficient is the jitter for domain failover refresh timer jitter",
		DefaultValue: 0.1,
	},
	FrontendErrorInjectionRate: DynamicFloat{
		KeyName:      "frontend.errorInjectionRate",
		Description:  "FrontendErrorInjectionRate is rate for injecting random error in frontend client",
		DefaultValue: 0,
	},
	MatchingErrorInjectionRate: DynamicFloat{
		KeyName:      "matching.errorInjectionRate",
		Description:  "MatchingErrorInjectionRate is rate for injecting random error in matching client",
		DefaultValue: 0,
	},
	TaskRedispatchIntervalJitterCoefficient: DynamicFloat{
		KeyName:      "history.taskRedispatchIntervalJitterCoefficient",
		Description:  "TaskRedispatchIntervalJitterCoefficient is the task redispatch interval jitter coefficient",
		DefaultValue: 0.15,
	},
	QueueProcessorRandomSplitProbability: DynamicFloat{
		KeyName:      "history.queueProcessorRandomSplitProbability",
		Description:  "QueueProcessorRandomSplitProbability is the probability for a domain to be split to a new processing queue",
		DefaultValue: 0.01,
	},
	QueueProcessorPollBackoffIntervalJitterCoefficient: DynamicFloat{
		KeyName:      "history.queueProcessorPollBackoffIntervalJitterCoefficient",
		Description:  "QueueProcessorPollBackoffIntervalJitterCoefficient is backoff interval jitter coefficient",
		DefaultValue: 0.15,
	},
	TimerProcessorUpdateAckIntervalJitterCoefficient: DynamicFloat{
		KeyName:      "history.timerProcessorUpdateAckIntervalJitterCoefficient",
		Description:  "TimerProcessorUpdateAckIntervalJitterCoefficient is the update interval jitter coefficient",
		DefaultValue: 0.15,
	},
	TimerProcessorMaxPollIntervalJitterCoefficient: DynamicFloat{
		KeyName:      "history.timerProcessorMaxPollIntervalJitterCoefficient",
		Description:  "TimerProcessorMaxPollIntervalJitterCoefficient is the max poll interval jitter coefficient",
		DefaultValue: 0.15,
	},
	TimerProcessorSplitQueueIntervalJitterCoefficient: DynamicFloat{
		KeyName:      "history.timerProcessorSplitQueueIntervalJitterCoefficient",
		Description:  "TimerProcessorSplitQueueIntervalJitterCoefficient is the split processing queue interval jitter coefficient",
		DefaultValue: 0.15,
	},
	TransferProcessorMaxPollIntervalJitterCoefficient: DynamicFloat{
		KeyName:      "history.transferProcessorMaxPollIntervalJitterCoefficient",
		Description:  "TransferProcessorMaxPollIntervalJitterCoefficient is the max poll interval jitter coefficient",
		DefaultValue: 0.15,
	},
	TransferProcessorSplitQueueIntervalJitterCoefficient: DynamicFloat{
		KeyName:      "history.transferProcessorSplitQueueIntervalJitterCoefficient",
		Description:  "TransferProcessorSplitQueueIntervalJitterCoefficient is the split processing queue interval jitter coefficient",
		DefaultValue: 0.15,
	},
	TransferProcessorUpdateAckIntervalJitterCoefficient: DynamicFloat{
		KeyName:      "history.transferProcessorUpdateAckIntervalJitterCoefficient",
		Description:  "TransferProcessorUpdateAckIntervalJitterCoefficient is the update interval jitter coefficient",
		DefaultValue: 0.15,
	},
	CrossClusterSourceProcessorMaxPollIntervalJitterCoefficient: DynamicFloat{
		KeyName:      "history.crossClusterSourceProcessorMaxPollIntervalJitterCoefficient",
		Description:  "CrossClusterSourceProcessorMaxPollIntervalJitterCoefficient is the max poll interval jitter coefficient",
		DefaultValue: 0.15,
	},
	CrossClusterSourceProcessorUpdateAckIntervalJitterCoefficient: DynamicFloat{
		KeyName:      "history.crossClusterSourceProcessorUpdateAckIntervalJitterCoefficient",
		Description:  "CrossClusterSourceProcessorUpdateAckIntervalJitterCoefficient is the update interval jitter coefficient",
		DefaultValue: 0.15,
	},
	CrossClusterTargetProcessorJitterCoefficient: DynamicFloat{
		KeyName:      "history.crossClusterTargetProcessorJitterCoefficient",
		Description:  "CrossClusterTargetProcessorJitterCoefficient is the jitter coefficient used in cross cluster task processor",
		DefaultValue: 0.15,
	},
	CrossClusterFetcherJitterCoefficient: DynamicFloat{
		KeyName:      "history.crossClusterFetcherJitterCoefficient",
		Description:  "CrossClusterFetcherJitterCoefficient is the jitter coefficient used in cross cluster task fetcher",
		DefaultValue: 0.15,
	},
	ReplicationTaskProcessorCleanupJitterCoefficient: DynamicFloat{
		KeyName:      "history.ReplicationTaskProcessorCleanupJitterCoefficient",
		Filters:      []Filter{ShardID},
		Description:  "ReplicationTaskProcessorCleanupJitterCoefficient is the jitter for cleanup timer",
		DefaultValue: 0.15,
	},
	ReplicationTaskProcessorStartWaitJitterCoefficient: DynamicFloat{
		KeyName:      "history.ReplicationTaskProcessorStartWaitJitterCoefficient",
		Filters:      []Filter{ShardID},
		Description:  "ReplicationTaskProcessorStartWaitJitterCoefficient is the jitter for batch start wait timer",
		DefaultValue: 0.9,
	},
	ReplicationTaskProcessorHostQPS: DynamicFloat{
		KeyName:      "history.ReplicationTaskProcessorHostQPS",
		Description:  "ReplicationTaskProcessorHostQPS is the qps of task processing rate limiter on host level",
		DefaultValue: 1500,
	},
	ReplicationTaskProcessorShardQPS: DynamicFloat{
		KeyName:      "history.ReplicationTaskProcessorShardQPS",
		Description:  "ReplicationTaskProcessorShardQPS is the qps of task processing rate limiter on shard level",
		DefaultValue: 5,
	},
	ReplicationTaskGenerationQPS: DynamicFloat{
		KeyName:      "history.ReplicationTaskGenerationQPS",
		Description:  "ReplicationTaskGenerationQPS is the wait time between each replication task generation qps",
		DefaultValue: 100,
	},
	MutableStateChecksumInvalidateBefore: DynamicFloat{
		KeyName:      "history.mutableStateChecksumInvalidateBefore",
		Description:  "MutableStateChecksumInvalidateBefore is the epoch timestamp before which all checksums are to be discarded",
		DefaultValue: 0,
	},
	NotifyFailoverMarkerTimerJitterCoefficient: DynamicFloat{
		KeyName:      "history.NotifyFailoverMarkerTimerJitterCoefficient",
		Description:  "NotifyFailoverMarkerTimerJitterCoefficient is the jitter for failover marker notifier timer",
		DefaultValue: 0.15,
	},
	HistoryErrorInjectionRate: DynamicFloat{
		KeyName:      "history.errorInjectionRate",
		Description:  "HistoryErrorInjectionRate is rate for injecting random error in history client",
		DefaultValue: 0,
	},
	ReplicationTaskFetcherTimerJitterCoefficient: DynamicFloat{
		KeyName:      "history.ReplicationTaskFetcherTimerJitterCoefficient",
		Description:  "ReplicationTaskFetcherTimerJitterCoefficient is the jitter for fetcher timer",
		DefaultValue: 0.15,
	},
	WorkerDeterministicConstructionCheckProbability: DynamicFloat{
		KeyName:      "worker.DeterministicConstructionCheckProbability",
		Description:  "WorkerDeterministicConstructionCheckProbability controls the probability of running a deterministic construction check for any given archival",
		DefaultValue: 0.002,
	},
	WorkerBlobIntegrityCheckProbability: DynamicFloat{
		KeyName:      "worker.BlobIntegrityCheckProbability",
		Description:  "WorkerBlobIntegrityCheckProbability controls the probability of running an integrity check for any given archival",
		DefaultValue: 0.002,
	},
}

var StringKeys = map[StringKey]DynamicString{
	TestGetStringPropertyKey: DynamicString{
		KeyName:      "testGetStringPropertyKey",
		Description:  "",
		DefaultValue: "",
	},
	AdvancedVisibilityWritingMode: DynamicString{
		KeyName:      "system.advancedVisibilityWritingMode",
		Description:  "AdvancedVisibilityWritingMode is key for how to write to advanced visibility. The most useful option is dual, which can be used for seamless migration from db visibility to advanced visibility, usually using with EnableReadVisibilityFromES",
		DefaultValue: "on",
	},
	HistoryArchivalStatus: DynamicString{
		KeyName:      "system.historyArchivalStatus",
		Description:  "HistoryArchivalStatus is key for the status of history archival to override the value from static config.",
		DefaultValue: "enabled",
	},
	VisibilityArchivalStatus: DynamicString{
		KeyName:      "system.visibilityArchivalStatus",
		Description:  "VisibilityArchivalStatus is key for the status of visibility archival to override the value from static config.",
		DefaultValue: "enabled",
	},
	DefaultEventEncoding: DynamicString{
		KeyName:      "history.defaultEventEncoding",
		Filters:      []Filter{DomainName},
		Description:  "DefaultEventEncoding is the encoding type for history events",
		DefaultValue: string(common.EncodingTypeThriftRW),
	},
	AdminOperationToken: DynamicString{
		KeyName:      "history.adminOperationToken",
		Description:  "AdminOperationToken is the token to pass admin checking",
		DefaultValue: "CadenceTeamONLY",
	},
	ESAnalyzerLimitToTypes: DynamicString{
		KeyName:      "worker.ESAnalyzerLimitToTypes",
		Description:  "ESAnalyzerLimitToTypes controls if we want to limit ESAnalyzer only to some workflow types",
		DefaultValue: "",
	},
	ESAnalyzerLimitToDomains: DynamicString{
		KeyName:      "worker.ESAnalyzerLimitToDomains",
		Description:  "ESAnalyzerLimitToDomains controls if we want to limit ESAnalyzer only to some domains",
		DefaultValue: "",
	},
	ESAnalyzerWorkflowDurationWarnThresholds: DynamicString{
		KeyName:      "worker.ESAnalyzerWorkflowDurationWarnThresholds",
		Description:  "ESAnalyzerWorkflowDurationWarnThresholds defines the warning execution thresholds for workflow types",
		DefaultValue: "",
	},
	ESAnalyzerWorkflowVersionMetricDomains: DynamicString{
		KeyName:      "worker.ESAnalyzerWorkflowVersionMetricDomains",
		Description:  "ESAnalyzerWorkflowDurationWarnThresholds defines the domains we want to emit wf version metrics on",
		DefaultValue: "",
	},
	ESAnalyzerWorkflowTypeMetricDomains: DynamicString{
		KeyName:      "worker.ESAnalyzerWorkflowTypeMetricDomains",
		Description:  "ESAnalyzerWorkflowDurationWarnThresholds defines the domains we want to emit wf version metrics on",
		DefaultValue: "",
	},
}

var DurationKeys = map[DurationKey]DynamicDuration{
	TestGetDurationPropertyKey: DynamicDuration{
		KeyName:      "testGetDurationPropertyKey",
		Description:  "",
		DefaultValue: 0,
	},
	TestGetDurationPropertyFilteredByDomainKey: DynamicDuration{
		KeyName:      "testGetDurationPropertyFilteredByDomainKey",
		Description:  "",
		DefaultValue: 0,
	},
	TestGetDurationPropertyFilteredByTaskListInfoKey: DynamicDuration{
		KeyName:      "testGetDurationPropertyFilteredByTaskListInfoKey",
		Description:  "",
		DefaultValue: 0,
	},
	FrontendShutdownDrainDuration: DynamicDuration{
		KeyName:      "frontend.shutdownDrainDuration",
		Description:  "FrontendShutdownDrainDuration is the duration of traffic drain during shutdown",
		DefaultValue: 0,
	},
	FrontendFailoverCoolDown: DynamicDuration{
		KeyName:      "frontend.failoverCoolDown",
		Filters:      []Filter{DomainName},
		Description:  "FrontendFailoverCoolDown is duration between two domain failvoers",
		DefaultValue: time.Minute,
	},
	DomainFailoverRefreshInterval: DynamicDuration{
		KeyName:      "frontend.domainFailoverRefreshInterval",
		Description:  "DomainFailoverRefreshInterval is the domain failover refresh timer",
		DefaultValue: time.Second * 10,
	},
	MatchingLongPollExpirationInterval: DynamicDuration{
		KeyName:      "matching.longPollExpirationInterval",
		Filters:      []Filter{DomainName, TaskListName, TaskType},
		Description:  "MatchingLongPollExpirationInterval is the long poll expiration interval in the matching service",
		DefaultValue: time.Minute,
	},
	MatchingUpdateAckInterval: DynamicDuration{
		KeyName:      "matching.updateAckInterval",
		Filters:      []Filter{DomainName, TaskListName, TaskType},
		Description:  "MatchingUpdateAckInterval is the interval for update ack",
		DefaultValue: time.Minute,
	},
	MatchingIdleTasklistCheckInterval: DynamicDuration{
		KeyName:      "matching.idleTasklistCheckInterval",
		Filters:      []Filter{DomainName, TaskListName, TaskType},
		Description:  "MatchingIdleTasklistCheckInterval is the IdleTasklistCheckInterval",
		DefaultValue: time.Minute * 5,
	},
	MaxTasklistIdleTime: DynamicDuration{
		KeyName:      "matching.maxTasklistIdleTime",
		Filters:      []Filter{DomainName, TaskListName, TaskType},
		Description:  "MaxTasklistIdleTime is the max time tasklist being idle",
		DefaultValue: time.Minute * 5,
	},
	MatchingShutdownDrainDuration: DynamicDuration{
		KeyName:      "matching.shutdownDrainDuration",
		Description:  "MatchingShutdownDrainDuration is the duration of traffic drain during shutdown",
		DefaultValue: 0,
	},
	MatchingActivityTaskSyncMatchWaitTime: DynamicDuration{
		KeyName:      "matching.activityTaskSyncMatchWaitTime",
		Filters:      []Filter{DomainName},
		Description:  "MatchingActivityTaskSyncMatchWaitTime is the amount of time activity task will wait to be sync matched",
		DefaultValue: time.Millisecond * 50,
	},
	HistoryLongPollExpirationInterval: DynamicDuration{
		KeyName:      "history.longPollExpirationInterval",
		Filters:      []Filter{DomainName},
		Description:  "HistoryLongPollExpirationInterval is the long poll expiration interval in the history service",
		DefaultValue: time.Second * 20, // history client: client/history/client.go set the client timeout 20s
	},
	HistoryCacheTTL: DynamicDuration{
		KeyName:      "history.cacheTTL",
		Description:  "HistoryCacheTTL is TTL of history cache",
		DefaultValue: time.Hour,
	},
	HistoryShutdownDrainDuration: DynamicDuration{
		KeyName:      "history.shutdownDrainDuration",
		Description:  "HistoryShutdownDrainDuration is the duration of traffic drain during shutdown",
		DefaultValue: 0,
	},
	EventsCacheTTL: DynamicDuration{
		KeyName:      "history.eventsCacheTTL",
		Description:  "EventsCacheTTL is TTL of events cache",
		DefaultValue: time.Hour,
	},
	AcquireShardInterval: DynamicDuration{
		KeyName:      "history.acquireShardInterval",
		Description:  "AcquireShardInterval is interval that timer used to acquire shard",
		DefaultValue: time.Minute,
	},
	StandbyClusterDelay: DynamicDuration{
		KeyName:      "history.standbyClusterDelay",
		Description:  "StandbyClusterDelay is the artificial delay added to standby cluster's view of active cluster's time",
		DefaultValue: time.Minute * 5,
	},
	StandbyTaskMissingEventsResendDelay: DynamicDuration{
		KeyName:      "history.standbyTaskMissingEventsResendDelay",
		Description:  "StandbyTaskMissingEventsResendDelay is the amount of time standby cluster's will wait (if events are missing)before calling remote for missing events",
		DefaultValue: time.Minute * 15,
	},
	StandbyTaskMissingEventsDiscardDelay: DynamicDuration{
		KeyName:      "history.standbyTaskMissingEventsDiscardDelay",
		Description:  "StandbyTaskMissingEventsDiscardDelay is the amount of time standby cluster's will wait (if events are missing)before discarding the task",
		DefaultValue: time.Minute * 25,
	},
	ActiveTaskRedispatchInterval: DynamicDuration{
		KeyName:      "history.activeTaskRedispatchInterval",
		Description:  "ActiveTaskRedispatchInterval is the active task redispatch interval",
		DefaultValue: time.Second * 5,
	},
	StandbyTaskRedispatchInterval: DynamicDuration{
		KeyName:      "history.standbyTaskRedispatchInterval",
		Description:  "StandbyTaskRedispatchInterval is the standby task redispatch interval",
		DefaultValue: time.Second * 30,
	},
	StandbyTaskReReplicationContextTimeout: DynamicDuration{
		KeyName:      "history.standbyTaskReReplicationContextTimeout",
		Filters:      []Filter{DomainID},
		Description:  "StandbyTaskReReplicationContextTimeout is the context timeout for standby task re-replication",
		DefaultValue: time.Minute * 3,
	},
	ResurrectionCheckMinDelay: DynamicDuration{
		KeyName:      "history.resurrectionCheckMinDelay",
		Filters:      []Filter{DomainName},
		Description:  "ResurrectionCheckMinDelay is the minimal timer processing delay before scanning history to see if there's a resurrected timer/activity",
		DefaultValue: time.Hour * 24,
	},
	QueueProcessorSplitLookAheadDurationByDomainID: DynamicDuration{
		KeyName:      "history.queueProcessorSplitLookAheadDurationByDomainID",
		Filters:      []Filter{DomainID},
		Description:  "QueueProcessorSplitLookAheadDurationByDomainID is the look ahead duration when spliting a domain to a new processing queue",
		DefaultValue: time.Minute * 20,
	},
	QueueProcessorPollBackoffInterval: DynamicDuration{
		KeyName:      "history.queueProcessorPollBackoffInterval",
		Description:  "QueueProcessorPollBackoffInterval is the backoff duration when queue processor is throttled",
		DefaultValue: time.Second * 5,
	},
	TimerProcessorUpdateAckInterval: DynamicDuration{
		KeyName:      "history.timerProcessorUpdateAckInterval",
		Description:  "TimerProcessorUpdateAckInterval is update interval for timer processor",
		DefaultValue: time.Second * 30,
	},
	TimerProcessorCompleteTimerInterval: DynamicDuration{
		KeyName:      "history.timerProcessorCompleteTimerInterval",
		Description:  "TimerProcessorCompleteTimerInterval is complete timer interval for timer processor",
		DefaultValue: time.Minute,
	},
	TimerProcessorFailoverMaxStartJitterInterval: DynamicDuration{
		KeyName:      "history.timerProcessorFailoverMaxStartJitterInterval",
		Description:  "TimerProcessorFailoverMaxStartJitterInterval is the max jitter interval for starting timer failover queue processing. The actual jitter interval used will be a random duration between 0 and the max interval so that timer failover queue across different shards won't start at the same time",
		DefaultValue: 0,
	},
	TimerProcessorMaxPollInterval: DynamicDuration{
		KeyName:      "history.timerProcessorMaxPollInterval",
		Description:  "TimerProcessorMaxPollInterval is max poll interval for timer processor",
		DefaultValue: time.Minute * 5,
	},
	TimerProcessorSplitQueueInterval: DynamicDuration{
		KeyName:      "history.timerProcessorSplitQueueInterval",
		Description:  "TimerProcessorSplitQueueInterval is the split processing queue interval for timer processor",
		DefaultValue: time.Minute,
	},
	TimerProcessorArchivalTimeLimit: DynamicDuration{
		KeyName:      "history.timerProcessorArchivalTimeLimit",
		Description:  "TimerProcessorArchivalTimeLimit is the upper time limit for inline history archival",
		DefaultValue: time.Second * 2,
	},
	TimerProcessorMaxTimeShift: DynamicDuration{
		KeyName:      "history.timerProcessorMaxTimeShift",
		Description:  "TimerProcessorMaxTimeShift is the max shift timer processor can have",
		DefaultValue: time.Second,
	},
	TransferProcessorFailoverMaxStartJitterInterval: DynamicDuration{
		KeyName:      "history.transferProcessorFailoverMaxStartJitterInterval",
		Description:  "TransferProcessorFailoverMaxStartJitterInterval is the max jitter interval for starting transfer failover queue processing. The actual jitter interval used will be a random duration between 0 and the max interval so that timer failover queue across different shards won't start at the same time",
		DefaultValue: 0,
	},
	TransferProcessorMaxPollInterval: DynamicDuration{
		KeyName:      "history.transferProcessorMaxPollInterval",
		Description:  "TransferProcessorMaxPollInterval is max poll interval for transferQueueProcessor",
		DefaultValue: time.Minute,
	},
	TransferProcessorSplitQueueInterval: DynamicDuration{
		KeyName:      "history.transferProcessorSplitQueueInterval",
		Description:  "TransferProcessorSplitQueueInterval is the split processing queue interval for transferQueueProcessor",
		DefaultValue: time.Minute,
	},
	TransferProcessorUpdateAckInterval: DynamicDuration{
		KeyName:      "history.transferProcessorUpdateAckInterval",
		Description:  "TransferProcessorUpdateAckInterval is update interval for transferQueueProcessor",
		DefaultValue: time.Second * 30,
	},
	TransferProcessorCompleteTransferInterval: DynamicDuration{
		KeyName:      "history.transferProcessorCompleteTransferInterval",
		Description:  "TransferProcessorCompleteTransferInterval is complete timer interval for transferQueueProcessor",
		DefaultValue: time.Minute,
	},
	TransferProcessorValidationInterval: DynamicDuration{
		KeyName:      "history.transferProcessorValidationInterval",
		Description:  "TransferProcessorValidationInterval is interval for performing transfer queue validation",
		DefaultValue: time.Second * 30,
	},
	TransferProcessorVisibilityArchivalTimeLimit: DynamicDuration{
		KeyName:      "history.transferProcessorVisibilityArchivalTimeLimit",
		Description:  "TransferProcessorVisibilityArchivalTimeLimit is the upper time limit for archiving visibility records",
		DefaultValue: time.Millisecond * 400,
	},
	CrossClusterSourceProcessorMaxPollInterval: DynamicDuration{
		KeyName:      "history.crossClusterSourceProcessorMaxPollInterval",
		Description:  "CrossClusterSourceProcessorMaxPollInterval is max poll interval for crossClusterQueueProcessor",
		DefaultValue: time.Minute,
	},
	CrossClusterSourceProcessorUpdateAckInterval: DynamicDuration{
		KeyName:      "history.crossClusterSourceProcessorUpdateAckInterval",
		Description:  "CrossClusterSourceProcessorUpdateAckInterval is update interval for crossClusterQueueProcessor",
		DefaultValue: time.Second * 30,
	},
	CrossClusterTargetProcessorTaskWaitInterval: DynamicDuration{
		KeyName:      "history.crossClusterTargetProcessorTaskWaitInterval",
		Description:  "CrossClusterTargetProcessorTaskWaitInterval is the duration for waiting a cross-cluster task response before responding to source",
		DefaultValue: time.Second * 3,
	},
	CrossClusterTargetProcessorServiceBusyBackoffInterval: DynamicDuration{
		KeyName:      "history.crossClusterTargetProcessorServiceBusyBackoffInterval",
		Description:  "CrossClusterTargetProcessorServiceBusyBackoffInterval is the backoff duration for cross cluster task processor when getting a service busy error when calling source cluster",
		DefaultValue: time.Second * 5,
	},
	CrossClusterFetcherAggregationInterval: DynamicDuration{
		KeyName:      "history.crossClusterFetcherAggregationInterval",
		Description:  "CrossClusterFetcherAggregationInterval determines how frequently the fetch requests are sent",
		DefaultValue: time.Second * 2,
	},
	CrossClusterFetcherServiceBusyBackoffInterval: DynamicDuration{
		KeyName:      "history.crossClusterFetcherServiceBusyBackoffInterval",
		Description:  "CrossClusterFetcherServiceBusyBackoffInterval is the backoff duration for cross cluster task fetcher when getting",
		DefaultValue: time.Second * 5,
	},
	CrossClusterFetcherErrorBackoffInterval: DynamicDuration{
		KeyName:      "history.crossClusterFetcherErrorBackoffInterval",
		Description:  "",
		DefaultValue: time.Second,
	},
	ReplicatorUpperLatency: DynamicDuration{
		KeyName:      "history.replicatorUpperLatency",
		Description:  "ReplicatorUpperLatency indicates the max allowed replication latency between clusters",
		DefaultValue: time.Second * 40,
	},
	ShardUpdateMinInterval: DynamicDuration{
		KeyName:      "history.shardUpdateMinInterval",
		Description:  "ShardUpdateMinInterval is the minimal time interval which the shard info can be updated",
		DefaultValue: time.Minute * 5,
	},
	ShardSyncMinInterval: DynamicDuration{
		KeyName:      "history.shardSyncMinInterval",
		Description:  "ShardSyncMinInterval is the minimal time interval which the shard info should be sync to remote",
		DefaultValue: time.Minute * 5,
	},
	StickyTTL: DynamicDuration{
		KeyName:      "history.stickyTTL",
		Filters:      []Filter{DomainName},
		Description:  "StickyTTL is to expire a sticky tasklist if no update more than this duration",
		DefaultValue: time.Hour * 24 * 365,
	},
	DecisionHeartbeatTimeout: DynamicDuration{
		KeyName:      "history.decisionHeartbeatTimeout",
		Filters:      []Filter{DomainName},
		Description:  "DecisionHeartbeatTimeout is for decision heartbeat",
		DefaultValue: time.Minute * 30, // about 30m
	},
	NormalDecisionScheduleToStartTimeout: DynamicDuration{
		KeyName:      "history.normalDecisionScheduleToStartTimeout",
		Filters:      []Filter{DomainName},
		Description:  "NormalDecisionScheduleToStartTimeout is scheduleToStart timeout duration for normal (non-sticky) decision task",
		DefaultValue: time.Minute * 5,
	},
	NotifyFailoverMarkerInterval: DynamicDuration{
		KeyName:      "history.NotifyFailoverMarkerInterval",
		Description:  "NotifyFailoverMarkerInterval is determines the frequency to notify failover marker",
		DefaultValue: time.Second * 5,
	},
	ActivityMaxScheduleToStartTimeoutForRetry: DynamicDuration{
		KeyName:      "history.activityMaxScheduleToStartTimeoutForRetry",
		Filters:      []Filter{DomainName},
		Description:  "ActivityMaxScheduleToStartTimeoutForRetry is maximum value allowed when overwritting the schedule to start timeout for activities with retry policy",
		DefaultValue: time.Minute * 30,
	},
	ReplicationTaskFetcherAggregationInterval: DynamicDuration{
		KeyName:      "history.ReplicationTaskFetcherAggregationInterval",
		Description:  "ReplicationTaskFetcherAggregationInterval determines how frequently the fetch requests are sent",
		DefaultValue: time.Second * 2,
	},
	ReplicationTaskFetcherErrorRetryWait: DynamicDuration{
		KeyName:      "history.ReplicationTaskFetcherErrorRetryWait",
		Description:  "ReplicationTaskFetcherErrorRetryWait is the wait time when fetcher encounters error",
		DefaultValue: time.Second,
	},
	ReplicationTaskFetcherServiceBusyWait: DynamicDuration{
		KeyName:      "history.ReplicationTaskFetcherServiceBusyWait",
		Description:  "ReplicationTaskFetcherServiceBusyWait is the wait time when fetcher encounters service busy error",
		DefaultValue: time.Minute,
	},
	ReplicationTaskProcessorErrorRetryWait: DynamicDuration{
		KeyName:      "history.ReplicationTaskProcessorErrorRetryWait",
		Filters:      []Filter{ShardID},
		Description:  "ReplicationTaskProcessorErrorRetryWait is the initial retry wait when we see errors in applying replication tasks",
		DefaultValue: time.Millisecond * 50,
	},
	ReplicationTaskProcessorErrorSecondRetryWait: DynamicDuration{
		KeyName:      "history.ReplicationTaskProcessorErrorSecondRetryWait",
		Filters:      []Filter{ShardID},
		Description:  "ReplicationTaskProcessorErrorSecondRetryWait is the initial retry wait for the second phase retry",
		DefaultValue: time.Second * 5,
	},
	ReplicationTaskProcessorErrorSecondRetryMaxWait: DynamicDuration{
		KeyName:      "history.ReplicationTaskProcessorErrorSecondRetryMaxWait",
		Filters:      []Filter{ShardID},
		Description:  "ReplicationTaskProcessorErrorSecondRetryMaxWait is the max wait time for the second phase retry",
		DefaultValue: time.Second * 30,
	},
	ReplicationTaskProcessorErrorSecondRetryExpiration: DynamicDuration{
		KeyName:      "history.ReplicationTaskProcessorErrorSecondRetryExpiration",
		Filters:      []Filter{ShardID},
		Description:  "ReplicationTaskProcessorErrorSecondRetryExpiration is the expiration duration for the second phase retry",
		DefaultValue: time.Minute * 5,
	},
	ReplicationTaskProcessorNoTaskInitialWait: DynamicDuration{
		KeyName:      "history.ReplicationTaskProcessorNoTaskInitialWait",
		Filters:      []Filter{ShardID},
		Description:  "ReplicationTaskProcessorNoTaskInitialWait is the wait time when not ask is returned",
		DefaultValue: time.Second * 2,
	},
	ReplicationTaskProcessorCleanupInterval: DynamicDuration{
		KeyName:      "history.ReplicationTaskProcessorCleanupInterval",
		Filters:      []Filter{ShardID},
		Description:  "ReplicationTaskProcessorCleanupInterval determines how frequently the cleanup replication queue",
		DefaultValue: time.Minute,
	},
	ReplicationTaskProcessorStartWait: DynamicDuration{
		KeyName:      "history.ReplicationTaskProcessorStartWait",
		Filters:      []Filter{ShardID},
		Description:  "ReplicationTaskProcessorStartWait is the wait time before each task processing batch",
		DefaultValue: time.Second * 5,
	},
	WorkerESProcessorFlushInterval: DynamicDuration{
		KeyName:      "worker.ESProcessorFlushInterval",
		Description:  "WorkerESProcessorFlushInterval is flush interval for esProcessor",
		DefaultValue: time.Second,
	},
	WorkerTimeLimitPerArchivalIteration: DynamicDuration{
		KeyName:      "worker.TimeLimitPerArchivalIteration",
		Description:  "WorkerTimeLimitPerArchivalIteration is controls the time limit of each iteration of archival workflow",
		DefaultValue: time.Hour * 24 * 15,
	},
	WorkerReplicationTaskMaxRetryDuration: DynamicDuration{
		KeyName:      "worker.replicationTaskMaxRetryDuration",
		Description:  "WorkerReplicationTaskMaxRetryDuration is the max retry duration for any task",
		DefaultValue: time.Minute * 10,
	},
	ESAnalyzerTimeWindow: DynamicDuration{
		KeyName:      "worker.ESAnalyzerTimeWindow",
		Description:  "ESAnalyzerTimeWindow defines the time window ElasticSearch Analyzer will consider while taking workflow averages",
		DefaultValue: time.Hour * 24 * 30,
	},
	IsolationGroupStateRefreshInterval: DynamicDuration{
		KeyName:      "system.isolationGroupStateRefreshInterval",
		Description:  "the frequency by which the IsolationGroupState handler will poll configuration",
		DefaultValue: time.Second * 30,
	},
	IsolationGroupStateFetchTimeout: DynamicDuration{
		KeyName:      "system.IsolationGroupStateFetchTimeout",
		Description:  "IsolationGroupStateFetchTimeout is the dynamic config DB fetch timeout value",
		DefaultValue: time.Second * 30,
	},
	IsolationGroupStateUpdateTimeout: DynamicDuration{
		KeyName:      "system.IsolationGroupStateUpdateTimeout",
		Description:  "IsolationGroupStateFetchTimeout is the dynamic config DB update timeout value",
		DefaultValue: time.Second * 30,
	},
	ESAnalyzerBufferWaitTime: DynamicDuration{
		KeyName:      "worker.ESAnalyzerBufferWaitTime",
		Description:  "ESAnalyzerBufferWaitTime controls min time required to consider a worklow stuck",
		DefaultValue: time.Minute * 30,
	},
	AsyncTaskDispatchTimeout: DynamicDuration{
		KeyName:      "matching.asyncTaskDispatchTimeout",
		Filters:      []Filter{DomainName, TaskListName, TaskType},
		Description:  "AsyncTaskDispatchTimeout is the timeout of dispatching tasks for async match",
		DefaultValue: time.Second * 3,
	},
}

var MapKeys = map[MapKey]DynamicMap{
	TestGetMapPropertyKey: DynamicMap{
		KeyName:      "testGetMapPropertyKey",
		Description:  "",
		DefaultValue: nil,
	},
	RequiredDomainDataKeys: DynamicMap{
		KeyName:      "system.requiredDomainDataKeys",
		Description:  "RequiredDomainDataKeys is the key for the list of data keys required in domain registration",
		DefaultValue: nil,
	},
	ValidSearchAttributes: DynamicMap{
		KeyName:      "frontend.validSearchAttributes",
		Description:  "ValidSearchAttributes is legal indexed keys that can be used in list APIs. When overriding, ensure to include the existing default attributes of the current release",
		DefaultValue: definition.GetDefaultIndexedKeys(),
	},
	TaskSchedulerRoundRobinWeights: DynamicMap{
		KeyName:     "history.taskSchedulerRoundRobinWeight",
		Description: "TaskSchedulerRoundRobinWeights is the priority weight for weighted round robin task scheduler",
		DefaultValue: common.ConvertIntMapToDynamicConfigMapProperty(map[int]int{
			common.GetTaskPriority(common.HighPriorityClass, common.DefaultPrioritySubclass):    500,
			common.GetTaskPriority(common.DefaultPriorityClass, common.DefaultPrioritySubclass): 20,
			common.GetTaskPriority(common.LowPriorityClass, common.DefaultPrioritySubclass):     5,
		}),
	},
	QueueProcessorPendingTaskSplitThreshold: DynamicMap{
		KeyName:      "history.queueProcessorPendingTaskSplitThreshold",
		Description:  "QueueProcessorPendingTaskSplitThreshold is the threshold for the number of pending tasks per domain",
		DefaultValue: common.ConvertIntMapToDynamicConfigMapProperty(map[int]int{0: 1000, 1: 10000}),
	},
	QueueProcessorStuckTaskSplitThreshold: DynamicMap{
		KeyName:      "history.queueProcessorStuckTaskSplitThreshold",
		Description:  "QueueProcessorStuckTaskSplitThreshold is the threshold for the number of attempts of a task",
		DefaultValue: common.ConvertIntMapToDynamicConfigMapProperty(map[int]int{0: 100, 1: 10000}),
	},
}

var ListKeys = map[ListKey]DynamicList{
	AllIsolationGroups: {
		KeyName:     "system.allIsolationGroups",
		Description: "A list of all the isolation groups in a system",
	},
	DefaultIsolationGroupConfigStoreManagerGlobalMapping: {
		KeyName: "system.defaultIsolationGroupConfigStoreManagerGlobalMapping",
		Description: "A configuration store for global isolation groups - used in isolation-group config only, not normal dynamic config." +
			"Not intended for use in normal dynamic config",
	},
	HeaderForwardingRules: {
		KeyName: "admin.HeaderForwardingRules", // make a new scope for global?
		Description: "Only loaded at startup.  " +
			"A list of rpc.HeaderRule values that define which headers to include or exclude for all requests, applied in order.  " +
			"Regexes and header names are used as-is, you are strongly encouraged to use `(?i)` to make your regex case-insensitive.",
		DefaultValue: []interface{}{
			// historical behavior: include literally everything.
			// this alone is quite problematic, and is strongly recommended against.
			map[string]interface{}{ // config imports dynamicconfig, sadly
				"Add":   true,
				"Match": "",
			},
		},
	},
}

var _keyNames map[string]Key

func init() {
	panicIfKeyInvalid := func(name string, key Key) {
		if name == "" {
			panic(fmt.Sprintf("empty keyName: %T, %v", key, key))
		}
		if _, ok := _keyNames[name]; ok {
			panic(fmt.Sprintf("duplicate keyName: %v", name))
		}
	}
	_keyNames = make(map[string]Key)
	for k, v := range IntKeys {
		panicIfKeyInvalid(v.KeyName, k)
		_keyNames[v.KeyName] = k
	}
	for k, v := range BoolKeys {
		panicIfKeyInvalid(v.KeyName, k)
		_keyNames[v.KeyName] = k
	}
	for k, v := range FloatKeys {
		panicIfKeyInvalid(v.KeyName, k)
		_keyNames[v.KeyName] = k
	}
	for k, v := range StringKeys {
		panicIfKeyInvalid(v.KeyName, k)
		_keyNames[v.KeyName] = k
	}
	for k, v := range DurationKeys {
		panicIfKeyInvalid(v.KeyName, k)
		_keyNames[v.KeyName] = k
	}
	for k, v := range MapKeys {
		panicIfKeyInvalid(v.KeyName, k)
		_keyNames[v.KeyName] = k
	}
	for k, v := range ListKeys {
		panicIfKeyInvalid(v.KeyName, k)
		_keyNames[v.KeyName] = k
	}
}
