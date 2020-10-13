// Copyright (c) 2017-2020 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package types

// AccessDeniedError is an internal type (TBD...)
type AccessDeniedError struct {
	Message string
}

// ActivityTaskCancelRequestedEventAttributes is an internal type (TBD...)
type ActivityTaskCancelRequestedEventAttributes struct {
	ActivityID                   *string
	DecisionTaskCompletedEventID *int64
}

// ActivityTaskCanceledEventAttributes is an internal type (TBD...)
type ActivityTaskCanceledEventAttributes struct {
	Details                      []byte
	LatestCancelRequestedEventID *int64
	ScheduledEventID             *int64
	StartedEventID               *int64
	Identity                     *string
}

// ActivityTaskCompletedEventAttributes is an internal type (TBD...)
type ActivityTaskCompletedEventAttributes struct {
	Result           []byte
	ScheduledEventID *int64
	StartedEventID   *int64
	Identity         *string
}

// ActivityTaskFailedEventAttributes is an internal type (TBD...)
type ActivityTaskFailedEventAttributes struct {
	Reason           *string
	Details          []byte
	ScheduledEventID *int64
	StartedEventID   *int64
	Identity         *string
}

// ActivityTaskScheduledEventAttributes is an internal type (TBD...)
type ActivityTaskScheduledEventAttributes struct {
	ActivityID                    *string
	ActivityType                  *ActivityType
	Domain                        *string
	TaskList                      *TaskList
	Input                         []byte
	ScheduleToCloseTimeoutSeconds *int32
	ScheduleToStartTimeoutSeconds *int32
	StartToCloseTimeoutSeconds    *int32
	HeartbeatTimeoutSeconds       *int32
	DecisionTaskCompletedEventID  *int64
	RetryPolicy                   *RetryPolicy
	Header                        *Header
}

// ActivityTaskStartedEventAttributes is an internal type (TBD...)
type ActivityTaskStartedEventAttributes struct {
	ScheduledEventID   *int64
	Identity           *string
	RequestID          *string
	Attempt            *int32
	LastFailureReason  *string
	LastFailureDetails []byte
}

// ActivityTaskTimedOutEventAttributes is an internal type (TBD...)
type ActivityTaskTimedOutEventAttributes struct {
	Details            []byte
	ScheduledEventID   *int64
	StartedEventID     *int64
	TimeoutType        *TimeoutType
	LastFailureReason  *string
	LastFailureDetails []byte
}

// ActivityType is an internal type (TBD...)
type ActivityType struct {
	Name *string
}

// ArchivalStatus is an internal type (TBD...)
type ArchivalStatus int32

const (
	// ArchivalStatusDisabled is an option for ArchivalStatus
	ArchivalStatusDisabled ArchivalStatus = iota
	// ArchivalStatusEnabled is an option for ArchivalStatus
	ArchivalStatusEnabled
)

// BadBinaries is an internal type (TBD...)
type BadBinaries struct {
	Binaries map[string]*BadBinaryInfo
}

// BadBinaryInfo is an internal type (TBD...)
type BadBinaryInfo struct {
	Reason          *string
	Operator        *string
	CreatedTimeNano *int64
}

// BadRequestError is an internal type (TBD...)
type BadRequestError struct {
	Message string
}

// CancelExternalWorkflowExecutionFailedCause is an internal type (TBD...)
type CancelExternalWorkflowExecutionFailedCause int32

const (
	// CancelExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution is an option for CancelExternalWorkflowExecutionFailedCause
	CancelExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution CancelExternalWorkflowExecutionFailedCause = iota
)

// CancelTimerDecisionAttributes is an internal type (TBD...)
type CancelTimerDecisionAttributes struct {
	TimerID *string
}

// CancelTimerFailedEventAttributes is an internal type (TBD...)
type CancelTimerFailedEventAttributes struct {
	TimerID                      *string
	Cause                        *string
	DecisionTaskCompletedEventID *int64
	Identity                     *string
}

// CancelWorkflowExecutionDecisionAttributes is an internal type (TBD...)
type CancelWorkflowExecutionDecisionAttributes struct {
	Details []byte
}

// CancellationAlreadyRequestedError is an internal type (TBD...)
type CancellationAlreadyRequestedError struct {
	Message string
}

// ChildWorkflowExecutionCanceledEventAttributes is an internal type (TBD...)
type ChildWorkflowExecutionCanceledEventAttributes struct {
	Details           []byte
	Domain            *string
	WorkflowExecution *WorkflowExecution
	WorkflowType      *WorkflowType
	InitiatedEventID  *int64
	StartedEventID    *int64
}

// ChildWorkflowExecutionCompletedEventAttributes is an internal type (TBD...)
type ChildWorkflowExecutionCompletedEventAttributes struct {
	Result            []byte
	Domain            *string
	WorkflowExecution *WorkflowExecution
	WorkflowType      *WorkflowType
	InitiatedEventID  *int64
	StartedEventID    *int64
}

// ChildWorkflowExecutionFailedCause is an internal type (TBD...)
type ChildWorkflowExecutionFailedCause int32

const (
	// ChildWorkflowExecutionFailedCauseWorkflowAlreadyRunning is an option for ChildWorkflowExecutionFailedCause
	ChildWorkflowExecutionFailedCauseWorkflowAlreadyRunning ChildWorkflowExecutionFailedCause = iota
)

// ChildWorkflowExecutionFailedEventAttributes is an internal type (TBD...)
type ChildWorkflowExecutionFailedEventAttributes struct {
	Reason            *string
	Details           []byte
	Domain            *string
	WorkflowExecution *WorkflowExecution
	WorkflowType      *WorkflowType
	InitiatedEventID  *int64
	StartedEventID    *int64
}

// ChildWorkflowExecutionStartedEventAttributes is an internal type (TBD...)
type ChildWorkflowExecutionStartedEventAttributes struct {
	Domain            *string
	InitiatedEventID  *int64
	WorkflowExecution *WorkflowExecution
	WorkflowType      *WorkflowType
	Header            *Header
}

// ChildWorkflowExecutionTerminatedEventAttributes is an internal type (TBD...)
type ChildWorkflowExecutionTerminatedEventAttributes struct {
	Domain            *string
	WorkflowExecution *WorkflowExecution
	WorkflowType      *WorkflowType
	InitiatedEventID  *int64
	StartedEventID    *int64
}

// ChildWorkflowExecutionTimedOutEventAttributes is an internal type (TBD...)
type ChildWorkflowExecutionTimedOutEventAttributes struct {
	TimeoutType       *TimeoutType
	Domain            *string
	WorkflowExecution *WorkflowExecution
	WorkflowType      *WorkflowType
	InitiatedEventID  *int64
	StartedEventID    *int64
}

// ClientVersionNotSupportedError is an internal type (TBD...)
type ClientVersionNotSupportedError struct {
	FeatureVersion    string
	ClientImpl        string
	SupportedVersions string
}

// CloseShardRequest is an internal type (TBD...)
type CloseShardRequest struct {
	ShardID *int32
}

// ClusterInfo is an internal type (TBD...)
type ClusterInfo struct {
	SupportedClientVersions *SupportedClientVersions
}

// ClusterReplicationConfiguration is an internal type (TBD...)
type ClusterReplicationConfiguration struct {
	ClusterName *string
}

// CompleteWorkflowExecutionDecisionAttributes is an internal type (TBD...)
type CompleteWorkflowExecutionDecisionAttributes struct {
	Result []byte
}

// ContinueAsNewInitiator is an internal type (TBD...)
type ContinueAsNewInitiator int32

const (
	// ContinueAsNewInitiatorCronSchedule is an option for ContinueAsNewInitiator
	ContinueAsNewInitiatorCronSchedule ContinueAsNewInitiator = iota
	// ContinueAsNewInitiatorDecider is an option for ContinueAsNewInitiator
	ContinueAsNewInitiatorDecider
	// ContinueAsNewInitiatorRetryPolicy is an option for ContinueAsNewInitiator
	ContinueAsNewInitiatorRetryPolicy
)

// ContinueAsNewWorkflowExecutionDecisionAttributes is an internal type (TBD...)
type ContinueAsNewWorkflowExecutionDecisionAttributes struct {
	WorkflowType                        *WorkflowType
	TaskList                            *TaskList
	Input                               []byte
	ExecutionStartToCloseTimeoutSeconds *int32
	TaskStartToCloseTimeoutSeconds      *int32
	BackoffStartIntervalInSeconds       *int32
	RetryPolicy                         *RetryPolicy
	Initiator                           *ContinueAsNewInitiator
	FailureReason                       *string
	FailureDetails                      []byte
	LastCompletionResult                []byte
	CronSchedule                        *string
	Header                              *Header
	Memo                                *Memo
	SearchAttributes                    *SearchAttributes
}

// CountWorkflowExecutionsRequest is an internal type (TBD...)
type CountWorkflowExecutionsRequest struct {
	Domain *string
	Query  *string
}

// CountWorkflowExecutionsResponse is an internal type (TBD...)
type CountWorkflowExecutionsResponse struct {
	Count *int64
}

// CurrentBranchChangedError is an internal type (TBD...)
type CurrentBranchChangedError struct {
	Message            string
	CurrentBranchToken []byte
}

// DataBlob is an internal type (TBD...)
type DataBlob struct {
	EncodingType *EncodingType
	Data         []byte
}

// Decision is an internal type (TBD...)
type Decision struct {
	DecisionType                                             *DecisionType
	ScheduleActivityTaskDecisionAttributes                   *ScheduleActivityTaskDecisionAttributes
	StartTimerDecisionAttributes                             *StartTimerDecisionAttributes
	CompleteWorkflowExecutionDecisionAttributes              *CompleteWorkflowExecutionDecisionAttributes
	FailWorkflowExecutionDecisionAttributes                  *FailWorkflowExecutionDecisionAttributes
	RequestCancelActivityTaskDecisionAttributes              *RequestCancelActivityTaskDecisionAttributes
	CancelTimerDecisionAttributes                            *CancelTimerDecisionAttributes
	CancelWorkflowExecutionDecisionAttributes                *CancelWorkflowExecutionDecisionAttributes
	RequestCancelExternalWorkflowExecutionDecisionAttributes *RequestCancelExternalWorkflowExecutionDecisionAttributes
	RecordMarkerDecisionAttributes                           *RecordMarkerDecisionAttributes
	ContinueAsNewWorkflowExecutionDecisionAttributes         *ContinueAsNewWorkflowExecutionDecisionAttributes
	StartChildWorkflowExecutionDecisionAttributes            *StartChildWorkflowExecutionDecisionAttributes
	SignalExternalWorkflowExecutionDecisionAttributes        *SignalExternalWorkflowExecutionDecisionAttributes
	UpsertWorkflowSearchAttributesDecisionAttributes         *UpsertWorkflowSearchAttributesDecisionAttributes
}

// DecisionTaskCompletedEventAttributes is an internal type (TBD...)
type DecisionTaskCompletedEventAttributes struct {
	ExecutionContext []byte
	ScheduledEventID *int64
	StartedEventID   *int64
	Identity         *string
	BinaryChecksum   *string
}

// DecisionTaskFailedCause is an internal type (TBD...)
type DecisionTaskFailedCause int32

const (
	// DecisionTaskFailedCauseBadBinary is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseBadBinary DecisionTaskFailedCause = iota
	// DecisionTaskFailedCauseBadCancelTimerAttributes is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseBadCancelTimerAttributes
	// DecisionTaskFailedCauseBadCancelWorkflowExecutionAttributes is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseBadCancelWorkflowExecutionAttributes
	// DecisionTaskFailedCauseBadCompleteWorkflowExecutionAttributes is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseBadCompleteWorkflowExecutionAttributes
	// DecisionTaskFailedCauseBadContinueAsNewAttributes is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseBadContinueAsNewAttributes
	// DecisionTaskFailedCauseBadFailWorkflowExecutionAttributes is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseBadFailWorkflowExecutionAttributes
	// DecisionTaskFailedCauseBadRecordMarkerAttributes is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseBadRecordMarkerAttributes
	// DecisionTaskFailedCauseBadRequestCancelActivityAttributes is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseBadRequestCancelActivityAttributes
	// DecisionTaskFailedCauseBadRequestCancelExternalWorkflowExecutionAttributes is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseBadRequestCancelExternalWorkflowExecutionAttributes
	// DecisionTaskFailedCauseBadScheduleActivityAttributes is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseBadScheduleActivityAttributes
	// DecisionTaskFailedCauseBadSearchAttributes is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseBadSearchAttributes
	// DecisionTaskFailedCauseBadSignalInputSize is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseBadSignalInputSize
	// DecisionTaskFailedCauseBadSignalWorkflowExecutionAttributes is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseBadSignalWorkflowExecutionAttributes
	// DecisionTaskFailedCauseBadStartChildExecutionAttributes is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseBadStartChildExecutionAttributes
	// DecisionTaskFailedCauseBadStartTimerAttributes is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseBadStartTimerAttributes
	// DecisionTaskFailedCauseFailoverCloseDecision is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseFailoverCloseDecision
	// DecisionTaskFailedCauseForceCloseDecision is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseForceCloseDecision
	// DecisionTaskFailedCauseResetStickyTasklist is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseResetStickyTasklist
	// DecisionTaskFailedCauseResetWorkflow is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseResetWorkflow
	// DecisionTaskFailedCauseScheduleActivityDuplicateID is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseScheduleActivityDuplicateID
	// DecisionTaskFailedCauseStartTimerDuplicateID is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseStartTimerDuplicateID
	// DecisionTaskFailedCauseUnhandledDecision is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseUnhandledDecision
	// DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure is an option for DecisionTaskFailedCause
	DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure
)

// DecisionTaskFailedEventAttributes is an internal type (TBD...)
type DecisionTaskFailedEventAttributes struct {
	ScheduledEventID *int64
	StartedEventID   *int64
	Cause            *DecisionTaskFailedCause
	Details          []byte
	Identity         *string
	Reason           *string
	BaseRunID        *string
	NewRunID         *string
	ForkEventVersion *int64
	BinaryChecksum   *string
}

// DecisionTaskScheduledEventAttributes is an internal type (TBD...)
type DecisionTaskScheduledEventAttributes struct {
	TaskList                   *TaskList
	StartToCloseTimeoutSeconds *int32
	Attempt                    *int64
}

// DecisionTaskStartedEventAttributes is an internal type (TBD...)
type DecisionTaskStartedEventAttributes struct {
	ScheduledEventID *int64
	Identity         *string
	RequestID        *string
}

// DecisionTaskTimedOutEventAttributes is an internal type (TBD...)
type DecisionTaskTimedOutEventAttributes struct {
	ScheduledEventID *int64
	StartedEventID   *int64
	TimeoutType      *TimeoutType
}

// DecisionType is an internal type (TBD...)
type DecisionType int32

const (
	// DecisionTypeCancelTimer is an option for DecisionType
	DecisionTypeCancelTimer DecisionType = iota
	// DecisionTypeCancelWorkflowExecution is an option for DecisionType
	DecisionTypeCancelWorkflowExecution
	// DecisionTypeCompleteWorkflowExecution is an option for DecisionType
	DecisionTypeCompleteWorkflowExecution
	// DecisionTypeContinueAsNewWorkflowExecution is an option for DecisionType
	DecisionTypeContinueAsNewWorkflowExecution
	// DecisionTypeFailWorkflowExecution is an option for DecisionType
	DecisionTypeFailWorkflowExecution
	// DecisionTypeRecordMarker is an option for DecisionType
	DecisionTypeRecordMarker
	// DecisionTypeRequestCancelActivityTask is an option for DecisionType
	DecisionTypeRequestCancelActivityTask
	// DecisionTypeRequestCancelExternalWorkflowExecution is an option for DecisionType
	DecisionTypeRequestCancelExternalWorkflowExecution
	// DecisionTypeScheduleActivityTask is an option for DecisionType
	DecisionTypeScheduleActivityTask
	// DecisionTypeSignalExternalWorkflowExecution is an option for DecisionType
	DecisionTypeSignalExternalWorkflowExecution
	// DecisionTypeStartChildWorkflowExecution is an option for DecisionType
	DecisionTypeStartChildWorkflowExecution
	// DecisionTypeStartTimer is an option for DecisionType
	DecisionTypeStartTimer
	// DecisionTypeUpsertWorkflowSearchAttributes is an option for DecisionType
	DecisionTypeUpsertWorkflowSearchAttributes
)

// DeprecateDomainRequest is an internal type (TBD...)
type DeprecateDomainRequest struct {
	Name          *string
	SecurityToken *string
}

// DescribeDomainRequest is an internal type (TBD...)
type DescribeDomainRequest struct {
	Name *string
	UUID *string
}

// DescribeDomainResponse is an internal type (TBD...)
type DescribeDomainResponse struct {
	DomainInfo               *DomainInfo
	Configuration            *DomainConfiguration
	ReplicationConfiguration *DomainReplicationConfiguration
	FailoverVersion          *int64
	IsGlobalDomain           *bool
}

// DescribeHistoryHostRequest is an internal type (TBD...)
type DescribeHistoryHostRequest struct {
	HostAddress      *string
	ShardIDForHost   *int32
	ExecutionForHost *WorkflowExecution
}

// DescribeHistoryHostResponse is an internal type (TBD...)
type DescribeHistoryHostResponse struct {
	NumberOfShards        *int32
	ShardIDs              []int32
	DomainCache           *DomainCacheInfo
	ShardControllerStatus *string
	Address               *string
}

// DescribeQueueRequest is an internal type (TBD...)
type DescribeQueueRequest struct {
	ShardID     *int32
	ClusterName *string
	Type        *int32
}

// DescribeQueueResponse is an internal type (TBD...)
type DescribeQueueResponse struct {
	ProcessingQueueStates []string
}

// DescribeTaskListRequest is an internal type (TBD...)
type DescribeTaskListRequest struct {
	Domain                *string
	TaskList              *TaskList
	TaskListType          *TaskListType
	IncludeTaskListStatus *bool
}

// DescribeTaskListResponse is an internal type (TBD...)
type DescribeTaskListResponse struct {
	Pollers        []*PollerInfo
	TaskListStatus *TaskListStatus
}

// DescribeWorkflowExecutionRequest is an internal type (TBD...)
type DescribeWorkflowExecutionRequest struct {
	Domain    *string
	Execution *WorkflowExecution
}

// DescribeWorkflowExecutionResponse is an internal type (TBD...)
type DescribeWorkflowExecutionResponse struct {
	ExecutionConfiguration *WorkflowExecutionConfiguration
	WorkflowExecutionInfo  *WorkflowExecutionInfo
	PendingActivities      []*PendingActivityInfo
	PendingChildren        []*PendingChildExecutionInfo
}

// DomainAlreadyExistsError is an internal type (TBD...)
type DomainAlreadyExistsError struct {
	Message string
}

// DomainCacheInfo is an internal type (TBD...)
type DomainCacheInfo struct {
	NumOfItemsInCacheByID   *int64
	NumOfItemsInCacheByName *int64
}

// DomainConfiguration is an internal type (TBD...)
type DomainConfiguration struct {
	WorkflowExecutionRetentionPeriodInDays *int32
	EmitMetric                             *bool
	BadBinaries                            *BadBinaries
	HistoryArchivalStatus                  *ArchivalStatus
	HistoryArchivalURI                     *string
	VisibilityArchivalStatus               *ArchivalStatus
	VisibilityArchivalURI                  *string
}

// DomainInfo is an internal type (TBD...)
type DomainInfo struct {
	Name        *string
	Status      *DomainStatus
	Description *string
	OwnerEmail  *string
	Data        map[string]string
	UUID        *string
}

// DomainNotActiveError is an internal type (TBD...)
type DomainNotActiveError struct {
	Message        string
	DomainName     string
	CurrentCluster string
	ActiveCluster  string
}

// DomainReplicationConfiguration is an internal type (TBD...)
type DomainReplicationConfiguration struct {
	ActiveClusterName *string
	Clusters          []*ClusterReplicationConfiguration
}

// DomainStatus is an internal type (TBD...)
type DomainStatus int32

const (
	// DomainStatusDeleted is an option for DomainStatus
	DomainStatusDeleted DomainStatus = iota
	// DomainStatusDeprecated is an option for DomainStatus
	DomainStatusDeprecated
	// DomainStatusRegistered is an option for DomainStatus
	DomainStatusRegistered
)

// EncodingType is an internal type (TBD...)
type EncodingType int32

const (
	// EncodingTypeJSON is an option for EncodingType
	EncodingTypeJSON EncodingType = iota
	// EncodingTypeThriftRW is an option for EncodingType
	EncodingTypeThriftRW
)

// EntityNotExistsError is an internal type (TBD...)
type EntityNotExistsError struct {
	Message        string
	CurrentCluster *string
	ActiveCluster  *string
}

// EventType is an internal type (TBD...)
type EventType int32

const (
	// EventTypeActivityTaskCancelRequested is an option for EventType
	EventTypeActivityTaskCancelRequested EventType = iota
	// EventTypeActivityTaskCanceled is an option for EventType
	EventTypeActivityTaskCanceled
	// EventTypeActivityTaskCompleted is an option for EventType
	EventTypeActivityTaskCompleted
	// EventTypeActivityTaskFailed is an option for EventType
	EventTypeActivityTaskFailed
	// EventTypeActivityTaskScheduled is an option for EventType
	EventTypeActivityTaskScheduled
	// EventTypeActivityTaskStarted is an option for EventType
	EventTypeActivityTaskStarted
	// EventTypeActivityTaskTimedOut is an option for EventType
	EventTypeActivityTaskTimedOut
	// EventTypeCancelTimerFailed is an option for EventType
	EventTypeCancelTimerFailed
	// EventTypeChildWorkflowExecutionCanceled is an option for EventType
	EventTypeChildWorkflowExecutionCanceled
	// EventTypeChildWorkflowExecutionCompleted is an option for EventType
	EventTypeChildWorkflowExecutionCompleted
	// EventTypeChildWorkflowExecutionFailed is an option for EventType
	EventTypeChildWorkflowExecutionFailed
	// EventTypeChildWorkflowExecutionStarted is an option for EventType
	EventTypeChildWorkflowExecutionStarted
	// EventTypeChildWorkflowExecutionTerminated is an option for EventType
	EventTypeChildWorkflowExecutionTerminated
	// EventTypeChildWorkflowExecutionTimedOut is an option for EventType
	EventTypeChildWorkflowExecutionTimedOut
	// EventTypeDecisionTaskCompleted is an option for EventType
	EventTypeDecisionTaskCompleted
	// EventTypeDecisionTaskFailed is an option for EventType
	EventTypeDecisionTaskFailed
	// EventTypeDecisionTaskScheduled is an option for EventType
	EventTypeDecisionTaskScheduled
	// EventTypeDecisionTaskStarted is an option for EventType
	EventTypeDecisionTaskStarted
	// EventTypeDecisionTaskTimedOut is an option for EventType
	EventTypeDecisionTaskTimedOut
	// EventTypeExternalWorkflowExecutionCancelRequested is an option for EventType
	EventTypeExternalWorkflowExecutionCancelRequested
	// EventTypeExternalWorkflowExecutionSignaled is an option for EventType
	EventTypeExternalWorkflowExecutionSignaled
	// EventTypeMarkerRecorded is an option for EventType
	EventTypeMarkerRecorded
	// EventTypeRequestCancelActivityTaskFailed is an option for EventType
	EventTypeRequestCancelActivityTaskFailed
	// EventTypeRequestCancelExternalWorkflowExecutionFailed is an option for EventType
	EventTypeRequestCancelExternalWorkflowExecutionFailed
	// EventTypeRequestCancelExternalWorkflowExecutionInitiated is an option for EventType
	EventTypeRequestCancelExternalWorkflowExecutionInitiated
	// EventTypeSignalExternalWorkflowExecutionFailed is an option for EventType
	EventTypeSignalExternalWorkflowExecutionFailed
	// EventTypeSignalExternalWorkflowExecutionInitiated is an option for EventType
	EventTypeSignalExternalWorkflowExecutionInitiated
	// EventTypeStartChildWorkflowExecutionFailed is an option for EventType
	EventTypeStartChildWorkflowExecutionFailed
	// EventTypeStartChildWorkflowExecutionInitiated is an option for EventType
	EventTypeStartChildWorkflowExecutionInitiated
	// EventTypeTimerCanceled is an option for EventType
	EventTypeTimerCanceled
	// EventTypeTimerFired is an option for EventType
	EventTypeTimerFired
	// EventTypeTimerStarted is an option for EventType
	EventTypeTimerStarted
	// EventTypeUpsertWorkflowSearchAttributes is an option for EventType
	EventTypeUpsertWorkflowSearchAttributes
	// EventTypeWorkflowExecutionCancelRequested is an option for EventType
	EventTypeWorkflowExecutionCancelRequested
	// EventTypeWorkflowExecutionCanceled is an option for EventType
	EventTypeWorkflowExecutionCanceled
	// EventTypeWorkflowExecutionCompleted is an option for EventType
	EventTypeWorkflowExecutionCompleted
	// EventTypeWorkflowExecutionContinuedAsNew is an option for EventType
	EventTypeWorkflowExecutionContinuedAsNew
	// EventTypeWorkflowExecutionFailed is an option for EventType
	EventTypeWorkflowExecutionFailed
	// EventTypeWorkflowExecutionSignaled is an option for EventType
	EventTypeWorkflowExecutionSignaled
	// EventTypeWorkflowExecutionStarted is an option for EventType
	EventTypeWorkflowExecutionStarted
	// EventTypeWorkflowExecutionTerminated is an option for EventType
	EventTypeWorkflowExecutionTerminated
	// EventTypeWorkflowExecutionTimedOut is an option for EventType
	EventTypeWorkflowExecutionTimedOut
)

// ExternalWorkflowExecutionCancelRequestedEventAttributes is an internal type (TBD...)
type ExternalWorkflowExecutionCancelRequestedEventAttributes struct {
	InitiatedEventID  *int64
	Domain            *string
	WorkflowExecution *WorkflowExecution
}

// ExternalWorkflowExecutionSignaledEventAttributes is an internal type (TBD...)
type ExternalWorkflowExecutionSignaledEventAttributes struct {
	InitiatedEventID  *int64
	Domain            *string
	WorkflowExecution *WorkflowExecution
	Control           []byte
}

// FailWorkflowExecutionDecisionAttributes is an internal type (TBD...)
type FailWorkflowExecutionDecisionAttributes struct {
	Reason  *string
	Details []byte
}

// GetSearchAttributesResponse is an internal type (TBD...)
type GetSearchAttributesResponse struct {
	Keys map[string]IndexedValueType
}

// GetWorkflowExecutionHistoryRequest is an internal type (TBD...)
type GetWorkflowExecutionHistoryRequest struct {
	Domain                 *string
	Execution              *WorkflowExecution
	MaximumPageSize        *int32
	NextPageToken          []byte
	WaitForNewEvent        *bool
	HistoryEventFilterType *HistoryEventFilterType
	SkipArchival           *bool
}

// GetWorkflowExecutionHistoryResponse is an internal type (TBD...)
type GetWorkflowExecutionHistoryResponse struct {
	History       *History
	RawHistory    []*DataBlob
	NextPageToken []byte
	Archived      *bool
}

// Header is an internal type (TBD...)
type Header struct {
	Fields map[string][]byte
}

// History is an internal type (TBD...)
type History struct {
	Events []*HistoryEvent
}

// HistoryBranch is an internal type (TBD...)
type HistoryBranch struct {
	TreeID    *string
	BranchID  *string
	Ancestors []*HistoryBranchRange
}

// HistoryBranchRange is an internal type (TBD...)
type HistoryBranchRange struct {
	BranchID    *string
	BeginNodeID *int64
	EndNodeID   *int64
}

// HistoryEvent is an internal type (TBD...)
type HistoryEvent struct {
	EventID                                                        *int64
	Timestamp                                                      *int64
	EventType                                                      *EventType
	Version                                                        *int64
	TaskID                                                         *int64
	WorkflowExecutionStartedEventAttributes                        *WorkflowExecutionStartedEventAttributes
	WorkflowExecutionCompletedEventAttributes                      *WorkflowExecutionCompletedEventAttributes
	WorkflowExecutionFailedEventAttributes                         *WorkflowExecutionFailedEventAttributes
	WorkflowExecutionTimedOutEventAttributes                       *WorkflowExecutionTimedOutEventAttributes
	DecisionTaskScheduledEventAttributes                           *DecisionTaskScheduledEventAttributes
	DecisionTaskStartedEventAttributes                             *DecisionTaskStartedEventAttributes
	DecisionTaskCompletedEventAttributes                           *DecisionTaskCompletedEventAttributes
	DecisionTaskTimedOutEventAttributes                            *DecisionTaskTimedOutEventAttributes
	DecisionTaskFailedEventAttributes                              *DecisionTaskFailedEventAttributes
	ActivityTaskScheduledEventAttributes                           *ActivityTaskScheduledEventAttributes
	ActivityTaskStartedEventAttributes                             *ActivityTaskStartedEventAttributes
	ActivityTaskCompletedEventAttributes                           *ActivityTaskCompletedEventAttributes
	ActivityTaskFailedEventAttributes                              *ActivityTaskFailedEventAttributes
	ActivityTaskTimedOutEventAttributes                            *ActivityTaskTimedOutEventAttributes
	TimerStartedEventAttributes                                    *TimerStartedEventAttributes
	TimerFiredEventAttributes                                      *TimerFiredEventAttributes
	ActivityTaskCancelRequestedEventAttributes                     *ActivityTaskCancelRequestedEventAttributes
	RequestCancelActivityTaskFailedEventAttributes                 *RequestCancelActivityTaskFailedEventAttributes
	ActivityTaskCanceledEventAttributes                            *ActivityTaskCanceledEventAttributes
	TimerCanceledEventAttributes                                   *TimerCanceledEventAttributes
	CancelTimerFailedEventAttributes                               *CancelTimerFailedEventAttributes
	MarkerRecordedEventAttributes                                  *MarkerRecordedEventAttributes
	WorkflowExecutionSignaledEventAttributes                       *WorkflowExecutionSignaledEventAttributes
	WorkflowExecutionTerminatedEventAttributes                     *WorkflowExecutionTerminatedEventAttributes
	WorkflowExecutionCancelRequestedEventAttributes                *WorkflowExecutionCancelRequestedEventAttributes
	WorkflowExecutionCanceledEventAttributes                       *WorkflowExecutionCanceledEventAttributes
	RequestCancelExternalWorkflowExecutionInitiatedEventAttributes *RequestCancelExternalWorkflowExecutionInitiatedEventAttributes
	RequestCancelExternalWorkflowExecutionFailedEventAttributes    *RequestCancelExternalWorkflowExecutionFailedEventAttributes
	ExternalWorkflowExecutionCancelRequestedEventAttributes        *ExternalWorkflowExecutionCancelRequestedEventAttributes
	WorkflowExecutionContinuedAsNewEventAttributes                 *WorkflowExecutionContinuedAsNewEventAttributes
	StartChildWorkflowExecutionInitiatedEventAttributes            *StartChildWorkflowExecutionInitiatedEventAttributes
	StartChildWorkflowExecutionFailedEventAttributes               *StartChildWorkflowExecutionFailedEventAttributes
	ChildWorkflowExecutionStartedEventAttributes                   *ChildWorkflowExecutionStartedEventAttributes
	ChildWorkflowExecutionCompletedEventAttributes                 *ChildWorkflowExecutionCompletedEventAttributes
	ChildWorkflowExecutionFailedEventAttributes                    *ChildWorkflowExecutionFailedEventAttributes
	ChildWorkflowExecutionCanceledEventAttributes                  *ChildWorkflowExecutionCanceledEventAttributes
	ChildWorkflowExecutionTimedOutEventAttributes                  *ChildWorkflowExecutionTimedOutEventAttributes
	ChildWorkflowExecutionTerminatedEventAttributes                *ChildWorkflowExecutionTerminatedEventAttributes
	SignalExternalWorkflowExecutionInitiatedEventAttributes        *SignalExternalWorkflowExecutionInitiatedEventAttributes
	SignalExternalWorkflowExecutionFailedEventAttributes           *SignalExternalWorkflowExecutionFailedEventAttributes
	ExternalWorkflowExecutionSignaledEventAttributes               *ExternalWorkflowExecutionSignaledEventAttributes
	UpsertWorkflowSearchAttributesEventAttributes                  *UpsertWorkflowSearchAttributesEventAttributes
}

// HistoryEventFilterType is an internal type (TBD...)
type HistoryEventFilterType int32

const (
	// HistoryEventFilterTypeAllEvent is an option for HistoryEventFilterType
	HistoryEventFilterTypeAllEvent HistoryEventFilterType = iota
	// HistoryEventFilterTypeCloseEvent is an option for HistoryEventFilterType
	HistoryEventFilterTypeCloseEvent
)

// IndexedValueType is an internal type (TBD...)
type IndexedValueType int32

const (
	// IndexedValueTypeBool is an option for IndexedValueType
	IndexedValueTypeBool IndexedValueType = iota
	// IndexedValueTypeDatetime is an option for IndexedValueType
	IndexedValueTypeDatetime
	// IndexedValueTypeDouble is an option for IndexedValueType
	IndexedValueTypeDouble
	// IndexedValueTypeInt is an option for IndexedValueType
	IndexedValueTypeInt
	// IndexedValueTypeKeyword is an option for IndexedValueType
	IndexedValueTypeKeyword
	// IndexedValueTypeString is an option for IndexedValueType
	IndexedValueTypeString
)

// InternalDataInconsistencyError is an internal type (TBD...)
type InternalDataInconsistencyError struct {
	Message string
}

// InternalServiceError is an internal type (TBD...)
type InternalServiceError struct {
	Message string
}

// LimitExceededError is an internal type (TBD...)
type LimitExceededError struct {
	Message string
}

// ListArchivedWorkflowExecutionsRequest is an internal type (TBD...)
type ListArchivedWorkflowExecutionsRequest struct {
	Domain        *string
	PageSize      *int32
	NextPageToken []byte
	Query         *string
}

// ListArchivedWorkflowExecutionsResponse is an internal type (TBD...)
type ListArchivedWorkflowExecutionsResponse struct {
	Executions    []*WorkflowExecutionInfo
	NextPageToken []byte
}

// ListClosedWorkflowExecutionsRequest is an internal type (TBD...)
type ListClosedWorkflowExecutionsRequest struct {
	Domain          *string
	MaximumPageSize *int32
	NextPageToken   []byte
	StartTimeFilter *StartTimeFilter
	ExecutionFilter *WorkflowExecutionFilter
	TypeFilter      *WorkflowTypeFilter
	StatusFilter    *WorkflowExecutionCloseStatus
}

// ListClosedWorkflowExecutionsResponse is an internal type (TBD...)
type ListClosedWorkflowExecutionsResponse struct {
	Executions    []*WorkflowExecutionInfo
	NextPageToken []byte
}

// ListDomainsRequest is an internal type (TBD...)
type ListDomainsRequest struct {
	PageSize      *int32
	NextPageToken []byte
}

// ListDomainsResponse is an internal type (TBD...)
type ListDomainsResponse struct {
	Domains       []*DescribeDomainResponse
	NextPageToken []byte
}

// ListOpenWorkflowExecutionsRequest is an internal type (TBD...)
type ListOpenWorkflowExecutionsRequest struct {
	Domain          *string
	MaximumPageSize *int32
	NextPageToken   []byte
	StartTimeFilter *StartTimeFilter
	ExecutionFilter *WorkflowExecutionFilter
	TypeFilter      *WorkflowTypeFilter
}

// ListOpenWorkflowExecutionsResponse is an internal type (TBD...)
type ListOpenWorkflowExecutionsResponse struct {
	Executions    []*WorkflowExecutionInfo
	NextPageToken []byte
}

// ListTaskListPartitionsRequest is an internal type (TBD...)
type ListTaskListPartitionsRequest struct {
	Domain   *string
	TaskList *TaskList
}

// ListTaskListPartitionsResponse is an internal type (TBD...)
type ListTaskListPartitionsResponse struct {
	ActivityTaskListPartitions []*TaskListPartitionMetadata
	DecisionTaskListPartitions []*TaskListPartitionMetadata
}

// ListWorkflowExecutionsRequest is an internal type (TBD...)
type ListWorkflowExecutionsRequest struct {
	Domain        *string
	PageSize      *int32
	NextPageToken []byte
	Query         *string
}

// ListWorkflowExecutionsResponse is an internal type (TBD...)
type ListWorkflowExecutionsResponse struct {
	Executions    []*WorkflowExecutionInfo
	NextPageToken []byte
}

// MarkerRecordedEventAttributes is an internal type (TBD...)
type MarkerRecordedEventAttributes struct {
	MarkerName                   *string
	Details                      []byte
	DecisionTaskCompletedEventID *int64
	Header                       *Header
}

// Memo is an internal type (TBD...)
type Memo struct {
	Fields map[string][]byte
}

// ParentClosePolicy is an internal type (TBD...)
type ParentClosePolicy int32

const (
	// ParentClosePolicyAbandon is an option for ParentClosePolicy
	ParentClosePolicyAbandon ParentClosePolicy = iota
	// ParentClosePolicyRequestCancel is an option for ParentClosePolicy
	ParentClosePolicyRequestCancel
	// ParentClosePolicyTerminate is an option for ParentClosePolicy
	ParentClosePolicyTerminate
)

// PendingActivityInfo is an internal type (TBD...)
type PendingActivityInfo struct {
	ActivityID             *string
	ActivityType           *ActivityType
	State                  *PendingActivityState
	HeartbeatDetails       []byte
	LastHeartbeatTimestamp *int64
	LastStartedTimestamp   *int64
	Attempt                *int32
	MaximumAttempts        *int32
	ScheduledTimestamp     *int64
	ExpirationTimestamp    *int64
	LastFailureReason      *string
	LastWorkerIdentity     *string
	LastFailureDetails     []byte
}

// PendingActivityState is an internal type (TBD...)
type PendingActivityState int32

const (
	// PendingActivityStateCancelRequested is an option for PendingActivityState
	PendingActivityStateCancelRequested PendingActivityState = iota
	// PendingActivityStateScheduled is an option for PendingActivityState
	PendingActivityStateScheduled
	// PendingActivityStateStarted is an option for PendingActivityState
	PendingActivityStateStarted
)

// PendingChildExecutionInfo is an internal type (TBD...)
type PendingChildExecutionInfo struct {
	WorkflowID        *string
	RunID             *string
	WorkflowTypName   *string
	InitiatedID       *int64
	ParentClosePolicy *ParentClosePolicy
}

// PollForActivityTaskRequest is an internal type (TBD...)
type PollForActivityTaskRequest struct {
	Domain           *string
	TaskList         *TaskList
	Identity         *string
	TaskListMetadata *TaskListMetadata
}

// PollForActivityTaskResponse is an internal type (TBD...)
type PollForActivityTaskResponse struct {
	TaskToken                       []byte
	WorkflowExecution               *WorkflowExecution
	ActivityID                      *string
	ActivityType                    *ActivityType
	Input                           []byte
	ScheduledTimestamp              *int64
	ScheduleToCloseTimeoutSeconds   *int32
	StartedTimestamp                *int64
	StartToCloseTimeoutSeconds      *int32
	HeartbeatTimeoutSeconds         *int32
	Attempt                         *int32
	ScheduledTimestampOfThisAttempt *int64
	HeartbeatDetails                []byte
	WorkflowType                    *WorkflowType
	WorkflowDomain                  *string
	Header                          *Header
}

// PollForDecisionTaskRequest is an internal type (TBD...)
type PollForDecisionTaskRequest struct {
	Domain         *string
	TaskList       *TaskList
	Identity       *string
	BinaryChecksum *string
}

// PollForDecisionTaskResponse is an internal type (TBD...)
type PollForDecisionTaskResponse struct {
	TaskToken                 []byte
	WorkflowExecution         *WorkflowExecution
	WorkflowType              *WorkflowType
	PreviousStartedEventID    *int64
	StartedEventID            *int64
	Attempt                   *int64
	BacklogCountHint          *int64
	History                   *History
	NextPageToken             []byte
	Query                     *WorkflowQuery
	WorkflowExecutionTaskList *TaskList
	ScheduledTimestamp        *int64
	StartedTimestamp          *int64
	Queries                   map[string]*WorkflowQuery
}

// PollerInfo is an internal type (TBD...)
type PollerInfo struct {
	LastAccessTime *int64
	Identity       *string
	RatePerSecond  *float64
}

// QueryConsistencyLevel is an internal type (TBD...)
type QueryConsistencyLevel int32

const (
	// QueryConsistencyLevelEventual is an option for QueryConsistencyLevel
	QueryConsistencyLevelEventual QueryConsistencyLevel = iota
	// QueryConsistencyLevelStrong is an option for QueryConsistencyLevel
	QueryConsistencyLevelStrong
)

// QueryFailedError is an internal type (TBD...)
type QueryFailedError struct {
	Message string
}

// QueryRejectCondition is an internal type (TBD...)
type QueryRejectCondition int32

const (
	// QueryRejectConditionNotCompletedCleanly is an option for QueryRejectCondition
	QueryRejectConditionNotCompletedCleanly QueryRejectCondition = iota
	// QueryRejectConditionNotOpen is an option for QueryRejectCondition
	QueryRejectConditionNotOpen
)

// QueryRejected is an internal type (TBD...)
type QueryRejected struct {
	CloseStatus *WorkflowExecutionCloseStatus
}

// QueryResultType is an internal type (TBD...)
type QueryResultType int32

const (
	// QueryResultTypeAnswered is an option for QueryResultType
	QueryResultTypeAnswered QueryResultType = iota
	// QueryResultTypeFailed is an option for QueryResultType
	QueryResultTypeFailed
)

// QueryTaskCompletedType is an internal type (TBD...)
type QueryTaskCompletedType int32

const (
	// QueryTaskCompletedTypeCompleted is an option for QueryTaskCompletedType
	QueryTaskCompletedTypeCompleted QueryTaskCompletedType = iota
	// QueryTaskCompletedTypeFailed is an option for QueryTaskCompletedType
	QueryTaskCompletedTypeFailed
)

// QueryWorkflowRequest is an internal type (TBD...)
type QueryWorkflowRequest struct {
	Domain                *string
	Execution             *WorkflowExecution
	Query                 *WorkflowQuery
	QueryRejectCondition  *QueryRejectCondition
	QueryConsistencyLevel *QueryConsistencyLevel
}

// QueryWorkflowResponse is an internal type (TBD...)
type QueryWorkflowResponse struct {
	QueryResult   []byte
	QueryRejected *QueryRejected
}

// ReapplyEventsRequest is an internal type (TBD...)
type ReapplyEventsRequest struct {
	DomainName        *string
	WorkflowExecution *WorkflowExecution
	Events            *DataBlob
}

// RecordActivityTaskHeartbeatByIDRequest is an internal type (TBD...)
type RecordActivityTaskHeartbeatByIDRequest struct {
	Domain     *string
	WorkflowID *string
	RunID      *string
	ActivityID *string
	Details    []byte
	Identity   *string
}

// RecordActivityTaskHeartbeatRequest is an internal type (TBD...)
type RecordActivityTaskHeartbeatRequest struct {
	TaskToken []byte
	Details   []byte
	Identity  *string
}

// RecordActivityTaskHeartbeatResponse is an internal type (TBD...)
type RecordActivityTaskHeartbeatResponse struct {
	CancelRequested *bool
}

// RecordMarkerDecisionAttributes is an internal type (TBD...)
type RecordMarkerDecisionAttributes struct {
	MarkerName *string
	Details    []byte
	Header     *Header
}

// RefreshWorkflowTasksRequest is an internal type (TBD...)
type RefreshWorkflowTasksRequest struct {
	Domain    *string
	Execution *WorkflowExecution
}

// RegisterDomainRequest is an internal type (TBD...)
type RegisterDomainRequest struct {
	Name                                   *string
	Description                            *string
	OwnerEmail                             *string
	WorkflowExecutionRetentionPeriodInDays *int32
	EmitMetric                             *bool
	Clusters                               []*ClusterReplicationConfiguration
	ActiveClusterName                      *string
	Data                                   map[string]string
	SecurityToken                          *string
	IsGlobalDomain                         *bool
	HistoryArchivalStatus                  *ArchivalStatus
	HistoryArchivalURI                     *string
	VisibilityArchivalStatus               *ArchivalStatus
	VisibilityArchivalURI                  *string
}

// RemoteSyncMatchedError is an internal type (TBD...)
type RemoteSyncMatchedError struct {
	Message string
}

// RemoveTaskRequest is an internal type (TBD...)
type RemoveTaskRequest struct {
	ShardID             *int32
	Type                *int32
	TaskID              *int64
	VisibilityTimestamp *int64
}

// RequestCancelActivityTaskDecisionAttributes is an internal type (TBD...)
type RequestCancelActivityTaskDecisionAttributes struct {
	ActivityID *string
}

// RequestCancelActivityTaskFailedEventAttributes is an internal type (TBD...)
type RequestCancelActivityTaskFailedEventAttributes struct {
	ActivityID                   *string
	Cause                        *string
	DecisionTaskCompletedEventID *int64
}

// RequestCancelExternalWorkflowExecutionDecisionAttributes is an internal type (TBD...)
type RequestCancelExternalWorkflowExecutionDecisionAttributes struct {
	Domain            *string
	WorkflowID        *string
	RunID             *string
	Control           []byte
	ChildWorkflowOnly *bool
}

// RequestCancelExternalWorkflowExecutionFailedEventAttributes is an internal type (TBD...)
type RequestCancelExternalWorkflowExecutionFailedEventAttributes struct {
	Cause                        *CancelExternalWorkflowExecutionFailedCause
	DecisionTaskCompletedEventID *int64
	Domain                       *string
	WorkflowExecution            *WorkflowExecution
	InitiatedEventID             *int64
	Control                      []byte
}

// RequestCancelExternalWorkflowExecutionInitiatedEventAttributes is an internal type (TBD...)
type RequestCancelExternalWorkflowExecutionInitiatedEventAttributes struct {
	DecisionTaskCompletedEventID *int64
	Domain                       *string
	WorkflowExecution            *WorkflowExecution
	Control                      []byte
	ChildWorkflowOnly            *bool
}

// RequestCancelWorkflowExecutionRequest is an internal type (TBD...)
type RequestCancelWorkflowExecutionRequest struct {
	Domain            *string
	WorkflowExecution *WorkflowExecution
	Identity          *string
	RequestID         *string
}

// ResetPointInfo is an internal type (TBD...)
type ResetPointInfo struct {
	BinaryChecksum           *string
	RunID                    *string
	FirstDecisionCompletedID *int64
	CreatedTimeNano          *int64
	ExpiringTimeNano         *int64
	Resettable               *bool
}

// ResetPoints is an internal type (TBD...)
type ResetPoints struct {
	Points []*ResetPointInfo
}

// ResetQueueRequest is an internal type (TBD...)
type ResetQueueRequest struct {
	ShardID     *int32
	ClusterName *string
	Type        *int32
}

// ResetStickyTaskListRequest is an internal type (TBD...)
type ResetStickyTaskListRequest struct {
	Domain    *string
	Execution *WorkflowExecution
}

// ResetStickyTaskListResponse is an internal type (TBD...)
type ResetStickyTaskListResponse struct {
}

// ResetWorkflowExecutionRequest is an internal type (TBD...)
type ResetWorkflowExecutionRequest struct {
	Domain                *string
	WorkflowExecution     *WorkflowExecution
	Reason                *string
	DecisionFinishEventID *int64
	RequestID             *string
}

// ResetWorkflowExecutionResponse is an internal type (TBD...)
type ResetWorkflowExecutionResponse struct {
	RunID *string
}

// RespondActivityTaskCanceledByIDRequest is an internal type (TBD...)
type RespondActivityTaskCanceledByIDRequest struct {
	Domain     *string
	WorkflowID *string
	RunID      *string
	ActivityID *string
	Details    []byte
	Identity   *string
}

// RespondActivityTaskCanceledRequest is an internal type (TBD...)
type RespondActivityTaskCanceledRequest struct {
	TaskToken []byte
	Details   []byte
	Identity  *string
}

// RespondActivityTaskCompletedByIDRequest is an internal type (TBD...)
type RespondActivityTaskCompletedByIDRequest struct {
	Domain     *string
	WorkflowID *string
	RunID      *string
	ActivityID *string
	Result     []byte
	Identity   *string
}

// RespondActivityTaskCompletedRequest is an internal type (TBD...)
type RespondActivityTaskCompletedRequest struct {
	TaskToken []byte
	Result    []byte
	Identity  *string
}

// RespondActivityTaskFailedByIDRequest is an internal type (TBD...)
type RespondActivityTaskFailedByIDRequest struct {
	Domain     *string
	WorkflowID *string
	RunID      *string
	ActivityID *string
	Reason     *string
	Details    []byte
	Identity   *string
}

// RespondActivityTaskFailedRequest is an internal type (TBD...)
type RespondActivityTaskFailedRequest struct {
	TaskToken []byte
	Reason    *string
	Details   []byte
	Identity  *string
}

// RespondDecisionTaskCompletedRequest is an internal type (TBD...)
type RespondDecisionTaskCompletedRequest struct {
	TaskToken                  []byte
	Decisions                  []*Decision
	ExecutionContext           []byte
	Identity                   *string
	StickyAttributes           *StickyExecutionAttributes
	ReturnNewDecisionTask      *bool
	ForceCreateNewDecisionTask *bool
	BinaryChecksum             *string
	QueryResults               map[string]*WorkflowQueryResult
}

// RespondDecisionTaskCompletedResponse is an internal type (TBD...)
type RespondDecisionTaskCompletedResponse struct {
	DecisionTask *PollForDecisionTaskResponse
}

// RespondDecisionTaskFailedRequest is an internal type (TBD...)
type RespondDecisionTaskFailedRequest struct {
	TaskToken      []byte
	Cause          *DecisionTaskFailedCause
	Details        []byte
	Identity       *string
	BinaryChecksum *string
}

// RespondQueryTaskCompletedRequest is an internal type (TBD...)
type RespondQueryTaskCompletedRequest struct {
	TaskToken         []byte
	CompletedType     *QueryTaskCompletedType
	QueryResult       []byte
	ErrorMessage      *string
	WorkerVersionInfo *WorkerVersionInfo
}

// RetryPolicy is an internal type (TBD...)
type RetryPolicy struct {
	InitialIntervalInSeconds    *int32
	BackoffCoefficient          *float64
	MaximumIntervalInSeconds    *int32
	MaximumAttempts             *int32
	NonRetriableErrorReasons    []string
	ExpirationIntervalInSeconds *int32
}

// RetryTaskV2Error is an internal type (TBD...)
type RetryTaskV2Error struct {
	Message           string
	DomainID          *string
	WorkflowID        *string
	RunID             *string
	StartEventID      *int64
	StartEventVersion *int64
	EndEventID        *int64
	EndEventVersion   *int64
}

// ScheduleActivityTaskDecisionAttributes is an internal type (TBD...)
type ScheduleActivityTaskDecisionAttributes struct {
	ActivityID                    *string
	ActivityType                  *ActivityType
	Domain                        *string
	TaskList                      *TaskList
	Input                         []byte
	ScheduleToCloseTimeoutSeconds *int32
	ScheduleToStartTimeoutSeconds *int32
	StartToCloseTimeoutSeconds    *int32
	HeartbeatTimeoutSeconds       *int32
	RetryPolicy                   *RetryPolicy
	Header                        *Header
}

// SearchAttributes is an internal type (TBD...)
type SearchAttributes struct {
	IndexedFields map[string][]byte
}

// ServiceBusyError is an internal type (TBD...)
type ServiceBusyError struct {
	Message string
}

// SignalExternalWorkflowExecutionDecisionAttributes is an internal type (TBD...)
type SignalExternalWorkflowExecutionDecisionAttributes struct {
	Domain            *string
	Execution         *WorkflowExecution
	SignalName        *string
	Input             []byte
	Control           []byte
	ChildWorkflowOnly *bool
}

// SignalExternalWorkflowExecutionFailedCause is an internal type (TBD...)
type SignalExternalWorkflowExecutionFailedCause int32

const (
	// SignalExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution is an option for SignalExternalWorkflowExecutionFailedCause
	SignalExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution SignalExternalWorkflowExecutionFailedCause = iota
)

// SignalExternalWorkflowExecutionFailedEventAttributes is an internal type (TBD...)
type SignalExternalWorkflowExecutionFailedEventAttributes struct {
	Cause                        *SignalExternalWorkflowExecutionFailedCause
	DecisionTaskCompletedEventID *int64
	Domain                       *string
	WorkflowExecution            *WorkflowExecution
	InitiatedEventID             *int64
	Control                      []byte
}

// SignalExternalWorkflowExecutionInitiatedEventAttributes is an internal type (TBD...)
type SignalExternalWorkflowExecutionInitiatedEventAttributes struct {
	DecisionTaskCompletedEventID *int64
	Domain                       *string
	WorkflowExecution            *WorkflowExecution
	SignalName                   *string
	Input                        []byte
	Control                      []byte
	ChildWorkflowOnly            *bool
}

// SignalWithStartWorkflowExecutionRequest is an internal type (TBD...)
type SignalWithStartWorkflowExecutionRequest struct {
	Domain                              *string
	WorkflowID                          *string
	WorkflowType                        *WorkflowType
	TaskList                            *TaskList
	Input                               []byte
	ExecutionStartToCloseTimeoutSeconds *int32
	TaskStartToCloseTimeoutSeconds      *int32
	Identity                            *string
	RequestID                           *string
	WorkflowIDReusePolicy               *WorkflowIDReusePolicy
	SignalName                          *string
	SignalInput                         []byte
	Control                             []byte
	RetryPolicy                         *RetryPolicy
	CronSchedule                        *string
	Memo                                *Memo
	SearchAttributes                    *SearchAttributes
	Header                              *Header
}

// SignalWorkflowExecutionRequest is an internal type (TBD...)
type SignalWorkflowExecutionRequest struct {
	Domain            *string
	WorkflowExecution *WorkflowExecution
	SignalName        *string
	Input             []byte
	Identity          *string
	RequestID         *string
	Control           []byte
}

// StartChildWorkflowExecutionDecisionAttributes is an internal type (TBD...)
type StartChildWorkflowExecutionDecisionAttributes struct {
	Domain                              *string
	WorkflowID                          *string
	WorkflowType                        *WorkflowType
	TaskList                            *TaskList
	Input                               []byte
	ExecutionStartToCloseTimeoutSeconds *int32
	TaskStartToCloseTimeoutSeconds      *int32
	ParentClosePolicy                   *ParentClosePolicy
	Control                             []byte
	WorkflowIDReusePolicy               *WorkflowIDReusePolicy
	RetryPolicy                         *RetryPolicy
	CronSchedule                        *string
	Header                              *Header
	Memo                                *Memo
	SearchAttributes                    *SearchAttributes
}

// StartChildWorkflowExecutionFailedEventAttributes is an internal type (TBD...)
type StartChildWorkflowExecutionFailedEventAttributes struct {
	Domain                       *string
	WorkflowID                   *string
	WorkflowType                 *WorkflowType
	Cause                        *ChildWorkflowExecutionFailedCause
	Control                      []byte
	InitiatedEventID             *int64
	DecisionTaskCompletedEventID *int64
}

// StartChildWorkflowExecutionInitiatedEventAttributes is an internal type (TBD...)
type StartChildWorkflowExecutionInitiatedEventAttributes struct {
	Domain                              *string
	WorkflowID                          *string
	WorkflowType                        *WorkflowType
	TaskList                            *TaskList
	Input                               []byte
	ExecutionStartToCloseTimeoutSeconds *int32
	TaskStartToCloseTimeoutSeconds      *int32
	ParentClosePolicy                   *ParentClosePolicy
	Control                             []byte
	DecisionTaskCompletedEventID        *int64
	WorkflowIDReusePolicy               *WorkflowIDReusePolicy
	RetryPolicy                         *RetryPolicy
	CronSchedule                        *string
	Header                              *Header
	Memo                                *Memo
	SearchAttributes                    *SearchAttributes
}

// StartTimeFilter is an internal type (TBD...)
type StartTimeFilter struct {
	EarliestTime *int64
	LatestTime   *int64
}

// StartTimerDecisionAttributes is an internal type (TBD...)
type StartTimerDecisionAttributes struct {
	TimerID                   *string
	StartToFireTimeoutSeconds *int64
}

// StartWorkflowExecutionRequest is an internal type (TBD...)
type StartWorkflowExecutionRequest struct {
	Domain                              *string
	WorkflowID                          *string
	WorkflowType                        *WorkflowType
	TaskList                            *TaskList
	Input                               []byte
	ExecutionStartToCloseTimeoutSeconds *int32
	TaskStartToCloseTimeoutSeconds      *int32
	Identity                            *string
	RequestID                           *string
	WorkflowIDReusePolicy               *WorkflowIDReusePolicy
	RetryPolicy                         *RetryPolicy
	CronSchedule                        *string
	Memo                                *Memo
	SearchAttributes                    *SearchAttributes
	Header                              *Header
}

// StartWorkflowExecutionResponse is an internal type (TBD...)
type StartWorkflowExecutionResponse struct {
	RunID *string
}

// StickyExecutionAttributes is an internal type (TBD...)
type StickyExecutionAttributes struct {
	WorkerTaskList                *TaskList
	ScheduleToStartTimeoutSeconds *int32
}

// SupportedClientVersions is an internal type (TBD...)
type SupportedClientVersions struct {
	GoSdk   *string
	JavaSdk *string
}

// TaskIDBlock is an internal type (TBD...)
type TaskIDBlock struct {
	StartID *int64
	EndID   *int64
}

// TaskList is an internal type (TBD...)
type TaskList struct {
	Name *string
	Kind *TaskListKind
}

// TaskListKind is an internal type (TBD...)
type TaskListKind int32

const (
	// TaskListKindNormal is an option for TaskListKind
	TaskListKindNormal TaskListKind = iota
	// TaskListKindSticky is an option for TaskListKind
	TaskListKindSticky
)

// TaskListMetadata is an internal type (TBD...)
type TaskListMetadata struct {
	MaxTasksPerSecond *float64
}

// TaskListPartitionMetadata is an internal type (TBD...)
type TaskListPartitionMetadata struct {
	Key           *string
	OwnerHostName *string
}

// TaskListStatus is an internal type (TBD...)
type TaskListStatus struct {
	BacklogCountHint *int64
	ReadLevel        *int64
	AckLevel         *int64
	RatePerSecond    *float64
	TaskIDBlock      *TaskIDBlock
}

// TaskListType is an internal type (TBD...)
type TaskListType int32

const (
	// TaskListTypeActivity is an option for TaskListType
	TaskListTypeActivity TaskListType = iota
	// TaskListTypeDecision is an option for TaskListType
	TaskListTypeDecision
)

// TerminateWorkflowExecutionRequest is an internal type (TBD...)
type TerminateWorkflowExecutionRequest struct {
	Domain            *string
	WorkflowExecution *WorkflowExecution
	Reason            *string
	Details           []byte
	Identity          *string
}

// TimeoutType is an internal type (TBD...)
type TimeoutType int32

const (
	// TimeoutTypeHeartbeat is an option for TimeoutType
	TimeoutTypeHeartbeat TimeoutType = iota
	// TimeoutTypeScheduleToClose is an option for TimeoutType
	TimeoutTypeScheduleToClose
	// TimeoutTypeScheduleToStart is an option for TimeoutType
	TimeoutTypeScheduleToStart
	// TimeoutTypeStartToClose is an option for TimeoutType
	TimeoutTypeStartToClose
)

// TimerCanceledEventAttributes is an internal type (TBD...)
type TimerCanceledEventAttributes struct {
	TimerID                      *string
	StartedEventID               *int64
	DecisionTaskCompletedEventID *int64
	Identity                     *string
}

// TimerFiredEventAttributes is an internal type (TBD...)
type TimerFiredEventAttributes struct {
	TimerID        *string
	StartedEventID *int64
}

// TimerStartedEventAttributes is an internal type (TBD...)
type TimerStartedEventAttributes struct {
	TimerID                      *string
	StartToFireTimeoutSeconds    *int64
	DecisionTaskCompletedEventID *int64
}

// TransientDecisionInfo is an internal type (TBD...)
type TransientDecisionInfo struct {
	ScheduledEvent *HistoryEvent
	StartedEvent   *HistoryEvent
}

// UpdateDomainInfo is an internal type (TBD...)
type UpdateDomainInfo struct {
	Description *string
	OwnerEmail  *string
	Data        map[string]string
}

// UpdateDomainRequest is an internal type (TBD...)
type UpdateDomainRequest struct {
	Name                     *string
	UpdatedInfo              *UpdateDomainInfo
	Configuration            *DomainConfiguration
	ReplicationConfiguration *DomainReplicationConfiguration
	SecurityToken            *string
	DeleteBadBinary          *string
	FailoverTimeoutInSeconds *int32
}

// UpdateDomainResponse is an internal type (TBD...)
type UpdateDomainResponse struct {
	DomainInfo               *DomainInfo
	Configuration            *DomainConfiguration
	ReplicationConfiguration *DomainReplicationConfiguration
	FailoverVersion          *int64
	IsGlobalDomain           *bool
}

// UpsertWorkflowSearchAttributesDecisionAttributes is an internal type (TBD...)
type UpsertWorkflowSearchAttributesDecisionAttributes struct {
	SearchAttributes *SearchAttributes
}

// UpsertWorkflowSearchAttributesEventAttributes is an internal type (TBD...)
type UpsertWorkflowSearchAttributesEventAttributes struct {
	DecisionTaskCompletedEventID *int64
	SearchAttributes             *SearchAttributes
}

// VersionHistories is an internal type (TBD...)
type VersionHistories struct {
	CurrentVersionHistoryIndex *int32
	Histories                  []*VersionHistory
}

// VersionHistory is an internal type (TBD...)
type VersionHistory struct {
	BranchToken []byte
	Items       []*VersionHistoryItem
}

// VersionHistoryItem is an internal type (TBD...)
type VersionHistoryItem struct {
	EventID *int64
	Version *int64
}

// WorkerVersionInfo is an internal type (TBD...)
type WorkerVersionInfo struct {
	Impl           *string
	FeatureVersion *string
}

// WorkflowExecution is an internal type (TBD...)
type WorkflowExecution struct {
	WorkflowID *string
	RunID      *string
}

// WorkflowExecutionAlreadyStartedError is an internal type (TBD...)
type WorkflowExecutionAlreadyStartedError struct {
	Message        *string
	StartRequestID *string
	RunID          *string
}

// WorkflowExecutionCancelRequestedEventAttributes is an internal type (TBD...)
type WorkflowExecutionCancelRequestedEventAttributes struct {
	Cause                     *string
	ExternalInitiatedEventID  *int64
	ExternalWorkflowExecution *WorkflowExecution
	Identity                  *string
}

// WorkflowExecutionCanceledEventAttributes is an internal type (TBD...)
type WorkflowExecutionCanceledEventAttributes struct {
	DecisionTaskCompletedEventID *int64
	Details                      []byte
}

// WorkflowExecutionCloseStatus is an internal type (TBD...)
type WorkflowExecutionCloseStatus int32

const (
	// WorkflowExecutionCloseStatusCanceled is an option for WorkflowExecutionCloseStatus
	WorkflowExecutionCloseStatusCanceled WorkflowExecutionCloseStatus = iota
	// WorkflowExecutionCloseStatusCompleted is an option for WorkflowExecutionCloseStatus
	WorkflowExecutionCloseStatusCompleted
	// WorkflowExecutionCloseStatusContinuedAsNew is an option for WorkflowExecutionCloseStatus
	WorkflowExecutionCloseStatusContinuedAsNew
	// WorkflowExecutionCloseStatusFailed is an option for WorkflowExecutionCloseStatus
	WorkflowExecutionCloseStatusFailed
	// WorkflowExecutionCloseStatusTerminated is an option for WorkflowExecutionCloseStatus
	WorkflowExecutionCloseStatusTerminated
	// WorkflowExecutionCloseStatusTimedOut is an option for WorkflowExecutionCloseStatus
	WorkflowExecutionCloseStatusTimedOut
)

// WorkflowExecutionCompletedEventAttributes is an internal type (TBD...)
type WorkflowExecutionCompletedEventAttributes struct {
	Result                       []byte
	DecisionTaskCompletedEventID *int64
}

// WorkflowExecutionConfiguration is an internal type (TBD...)
type WorkflowExecutionConfiguration struct {
	TaskList                            *TaskList
	ExecutionStartToCloseTimeoutSeconds *int32
	TaskStartToCloseTimeoutSeconds      *int32
}

// WorkflowExecutionContinuedAsNewEventAttributes is an internal type (TBD...)
type WorkflowExecutionContinuedAsNewEventAttributes struct {
	NewExecutionRunID                   *string
	WorkflowType                        *WorkflowType
	TaskList                            *TaskList
	Input                               []byte
	ExecutionStartToCloseTimeoutSeconds *int32
	TaskStartToCloseTimeoutSeconds      *int32
	DecisionTaskCompletedEventID        *int64
	BackoffStartIntervalInSeconds       *int32
	Initiator                           *ContinueAsNewInitiator
	FailureReason                       *string
	FailureDetails                      []byte
	LastCompletionResult                []byte
	Header                              *Header
	Memo                                *Memo
	SearchAttributes                    *SearchAttributes
}

// WorkflowExecutionFailedEventAttributes is an internal type (TBD...)
type WorkflowExecutionFailedEventAttributes struct {
	Reason                       *string
	Details                      []byte
	DecisionTaskCompletedEventID *int64
}

// WorkflowExecutionFilter is an internal type (TBD...)
type WorkflowExecutionFilter struct {
	WorkflowID *string
	RunID      *string
}

// WorkflowExecutionInfo is an internal type (TBD...)
type WorkflowExecutionInfo struct {
	Execution        *WorkflowExecution
	Type             *WorkflowType
	StartTime        *int64
	CloseTime        *int64
	CloseStatus      *WorkflowExecutionCloseStatus
	HistoryLength    *int64
	ParentDomainID   *string
	ParentExecution  *WorkflowExecution
	ExecutionTime    *int64
	Memo             *Memo
	SearchAttributes *SearchAttributes
	AutoResetPoints  *ResetPoints
	TaskList         *string
}

// WorkflowExecutionSignaledEventAttributes is an internal type (TBD...)
type WorkflowExecutionSignaledEventAttributes struct {
	SignalName *string
	Input      []byte
	Identity   *string
}

// WorkflowExecutionStartedEventAttributes is an internal type (TBD...)
type WorkflowExecutionStartedEventAttributes struct {
	WorkflowType                        *WorkflowType
	ParentWorkflowDomain                *string
	ParentWorkflowExecution             *WorkflowExecution
	ParentInitiatedEventID              *int64
	TaskList                            *TaskList
	Input                               []byte
	ExecutionStartToCloseTimeoutSeconds *int32
	TaskStartToCloseTimeoutSeconds      *int32
	ContinuedExecutionRunID             *string
	Initiator                           *ContinueAsNewInitiator
	ContinuedFailureReason              *string
	ContinuedFailureDetails             []byte
	LastCompletionResult                []byte
	OriginalExecutionRunID              *string
	Identity                            *string
	FirstExecutionRunID                 *string
	RetryPolicy                         *RetryPolicy
	Attempt                             *int32
	ExpirationTimestamp                 *int64
	CronSchedule                        *string
	FirstDecisionTaskBackoffSeconds     *int32
	Memo                                *Memo
	SearchAttributes                    *SearchAttributes
	PrevAutoResetPoints                 *ResetPoints
	Header                              *Header
}

// WorkflowExecutionTerminatedEventAttributes is an internal type (TBD...)
type WorkflowExecutionTerminatedEventAttributes struct {
	Reason   *string
	Details  []byte
	Identity *string
}

// WorkflowExecutionTimedOutEventAttributes is an internal type (TBD...)
type WorkflowExecutionTimedOutEventAttributes struct {
	TimeoutType *TimeoutType
}

// WorkflowIDReusePolicy is an internal type (TBD...)
type WorkflowIDReusePolicy int32

const (
	// WorkflowIDReusePolicyAllowDuplicate is an option for WorkflowIDReusePolicy
	WorkflowIDReusePolicyAllowDuplicate WorkflowIDReusePolicy = iota
	// WorkflowIDReusePolicyAllowDuplicateFailedOnly is an option for WorkflowIDReusePolicy
	WorkflowIDReusePolicyAllowDuplicateFailedOnly
	// WorkflowIDReusePolicyRejectDuplicate is an option for WorkflowIDReusePolicy
	WorkflowIDReusePolicyRejectDuplicate
	// WorkflowIDReusePolicyTerminateIfRunning is an option for WorkflowIDReusePolicy
	WorkflowIDReusePolicyTerminateIfRunning
)

// WorkflowQuery is an internal type (TBD...)
type WorkflowQuery struct {
	QueryType *string
	QueryArgs []byte
}

// WorkflowQueryResult is an internal type (TBD...)
type WorkflowQueryResult struct {
	ResultType   *QueryResultType
	Answer       []byte
	ErrorMessage *string
}

// WorkflowType is an internal type (TBD...)
type WorkflowType struct {
	Name *string
}

// WorkflowTypeFilter is an internal type (TBD...)
type WorkflowTypeFilter struct {
	Name *string
}
