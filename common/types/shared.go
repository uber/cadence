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

type AccessDeniedError struct {
	Message string
}

type ActivityTaskCancelRequestedEventAttributes struct {
	ActivityId                   *string
	DecisionTaskCompletedEventId *int64
}

type ActivityTaskCanceledEventAttributes struct {
	Details                      []byte
	LatestCancelRequestedEventId *int64
	ScheduledEventId             *int64
	StartedEventId               *int64
	Identity                     *string
}

type ActivityTaskCompletedEventAttributes struct {
	Result           []byte
	ScheduledEventId *int64
	StartedEventId   *int64
	Identity         *string
}

type ActivityTaskFailedEventAttributes struct {
	Reason           *string
	Details          []byte
	ScheduledEventId *int64
	StartedEventId   *int64
	Identity         *string
}

type ActivityTaskScheduledEventAttributes struct {
	ActivityId                    *string
	ActivityType                  *ActivityType
	Domain                        *string
	TaskList                      *TaskList
	Input                         []byte
	ScheduleToCloseTimeoutSeconds *int32
	ScheduleToStartTimeoutSeconds *int32
	StartToCloseTimeoutSeconds    *int32
	HeartbeatTimeoutSeconds       *int32
	DecisionTaskCompletedEventId  *int64
	RetryPolicy                   *RetryPolicy
	Header                        *Header
}

type ActivityTaskStartedEventAttributes struct {
	ScheduledEventId   *int64
	Identity           *string
	RequestId          *string
	Attempt            *int32
	LastFailureReason  *string
	LastFailureDetails []byte
}

type ActivityTaskTimedOutEventAttributes struct {
	Details            []byte
	ScheduledEventId   *int64
	StartedEventId     *int64
	TimeoutType        *TimeoutType
	LastFailureReason  *string
	LastFailureDetails []byte
}

type ActivityType struct {
	Name *string
}

type ArchivalStatus int32

const (
	ArchivalStatusDisabled ArchivalStatus = iota
	ArchivalStatusEnabled
)

type BadBinaries struct {
	Binaries map[string]*BadBinaryInfo
}

type BadBinaryInfo struct {
	Reason          *string
	Operator        *string
	CreatedTimeNano *int64
}

type BadRequestError struct {
	Message string
}

type CancelExternalWorkflowExecutionFailedCause int32

const (
	CancelExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution CancelExternalWorkflowExecutionFailedCause = iota
)

type CancelTimerDecisionAttributes struct {
	TimerId *string
}

type CancelTimerFailedEventAttributes struct {
	TimerId                      *string
	Cause                        *string
	DecisionTaskCompletedEventId *int64
	Identity                     *string
}

type CancelWorkflowExecutionDecisionAttributes struct {
	Details []byte
}

type CancellationAlreadyRequestedError struct {
	Message string
}

type ChildWorkflowExecutionCanceledEventAttributes struct {
	Details           []byte
	Domain            *string
	WorkflowExecution *WorkflowExecution
	WorkflowType      *WorkflowType
	InitiatedEventId  *int64
	StartedEventId    *int64
}

type ChildWorkflowExecutionCompletedEventAttributes struct {
	Result            []byte
	Domain            *string
	WorkflowExecution *WorkflowExecution
	WorkflowType      *WorkflowType
	InitiatedEventId  *int64
	StartedEventId    *int64
}

type ChildWorkflowExecutionFailedCause int32

const (
	ChildWorkflowExecutionFailedCauseWorkflowAlreadyRunning ChildWorkflowExecutionFailedCause = iota
)

type ChildWorkflowExecutionFailedEventAttributes struct {
	Reason            *string
	Details           []byte
	Domain            *string
	WorkflowExecution *WorkflowExecution
	WorkflowType      *WorkflowType
	InitiatedEventId  *int64
	StartedEventId    *int64
}

type ChildWorkflowExecutionStartedEventAttributes struct {
	Domain            *string
	InitiatedEventId  *int64
	WorkflowExecution *WorkflowExecution
	WorkflowType      *WorkflowType
	Header            *Header
}

type ChildWorkflowExecutionTerminatedEventAttributes struct {
	Domain            *string
	WorkflowExecution *WorkflowExecution
	WorkflowType      *WorkflowType
	InitiatedEventId  *int64
	StartedEventId    *int64
}

type ChildWorkflowExecutionTimedOutEventAttributes struct {
	TimeoutType       *TimeoutType
	Domain            *string
	WorkflowExecution *WorkflowExecution
	WorkflowType      *WorkflowType
	InitiatedEventId  *int64
	StartedEventId    *int64
}

type ClientVersionNotSupportedError struct {
	FeatureVersion    string
	ClientImpl        string
	SupportedVersions string
}

type CloseShardRequest struct {
	ShardID *int32
}

type ClusterInfo struct {
	SupportedClientVersions *SupportedClientVersions
}

type ClusterReplicationConfiguration struct {
	ClusterName *string
}

type CompleteWorkflowExecutionDecisionAttributes struct {
	Result []byte
}

type ContinueAsNewInitiator int32

const (
	ContinueAsNewInitiatorCronSchedule ContinueAsNewInitiator = iota
	ContinueAsNewInitiatorDecider
	ContinueAsNewInitiatorRetryPolicy
)

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

type CountWorkflowExecutionsRequest struct {
	Domain *string
	Query  *string
}

type CountWorkflowExecutionsResponse struct {
	Count *int64
}

type CurrentBranchChangedError struct {
	Message            string
	CurrentBranchToken []byte
}

type DataBlob struct {
	EncodingType *EncodingType
	Data         []byte
}

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

type DecisionTaskCompletedEventAttributes struct {
	ExecutionContext []byte
	ScheduledEventId *int64
	StartedEventId   *int64
	Identity         *string
	BinaryChecksum   *string
}

type DecisionTaskFailedCause int32

const (
	DecisionTaskFailedCauseBadBinary DecisionTaskFailedCause = iota
	DecisionTaskFailedCauseBadCancelTimerAttributes
	DecisionTaskFailedCauseBadCancelWorkflowExecutionAttributes
	DecisionTaskFailedCauseBadCompleteWorkflowExecutionAttributes
	DecisionTaskFailedCauseBadContinueAsNewAttributes
	DecisionTaskFailedCauseBadFailWorkflowExecutionAttributes
	DecisionTaskFailedCauseBadRecordMarkerAttributes
	DecisionTaskFailedCauseBadRequestCancelActivityAttributes
	DecisionTaskFailedCauseBadRequestCancelExternalWorkflowExecutionAttributes
	DecisionTaskFailedCauseBadScheduleActivityAttributes
	DecisionTaskFailedCauseBadSearchAttributes
	DecisionTaskFailedCauseBadSignalInputSize
	DecisionTaskFailedCauseBadSignalWorkflowExecutionAttributes
	DecisionTaskFailedCauseBadStartChildExecutionAttributes
	DecisionTaskFailedCauseBadStartTimerAttributes
	DecisionTaskFailedCauseFailoverCloseDecision
	DecisionTaskFailedCauseForceCloseDecision
	DecisionTaskFailedCauseResetStickyTasklist
	DecisionTaskFailedCauseResetWorkflow
	DecisionTaskFailedCauseScheduleActivityDuplicateID
	DecisionTaskFailedCauseStartTimerDuplicateID
	DecisionTaskFailedCauseUnhandledDecision
	DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure
)

type DecisionTaskFailedEventAttributes struct {
	ScheduledEventId *int64
	StartedEventId   *int64
	Cause            *DecisionTaskFailedCause
	Details          []byte
	Identity         *string
	Reason           *string
	BaseRunId        *string
	NewRunId         *string
	ForkEventVersion *int64
	BinaryChecksum   *string
}

type DecisionTaskScheduledEventAttributes struct {
	TaskList                   *TaskList
	StartToCloseTimeoutSeconds *int32
	Attempt                    *int64
}

type DecisionTaskStartedEventAttributes struct {
	ScheduledEventId *int64
	Identity         *string
	RequestId        *string
}

type DecisionTaskTimedOutEventAttributes struct {
	ScheduledEventId *int64
	StartedEventId   *int64
	TimeoutType      *TimeoutType
}

type DecisionType int32

const (
	DecisionTypeCancelTimer DecisionType = iota
	DecisionTypeCancelWorkflowExecution
	DecisionTypeCompleteWorkflowExecution
	DecisionTypeContinueAsNewWorkflowExecution
	DecisionTypeFailWorkflowExecution
	DecisionTypeRecordMarker
	DecisionTypeRequestCancelActivityTask
	DecisionTypeRequestCancelExternalWorkflowExecution
	DecisionTypeScheduleActivityTask
	DecisionTypeSignalExternalWorkflowExecution
	DecisionTypeStartChildWorkflowExecution
	DecisionTypeStartTimer
	DecisionTypeUpsertWorkflowSearchAttributes
)

type DeprecateDomainRequest struct {
	Name          *string
	SecurityToken *string
}

type DescribeDomainRequest struct {
	Name *string
	UUID *string
}

type DescribeDomainResponse struct {
	DomainInfo               *DomainInfo
	Configuration            *DomainConfiguration
	ReplicationConfiguration *DomainReplicationConfiguration
	FailoverVersion          *int64
	IsGlobalDomain           *bool
}

type DescribeHistoryHostRequest struct {
	HostAddress      *string
	ShardIdForHost   *int32
	ExecutionForHost *WorkflowExecution
}

type DescribeHistoryHostResponse struct {
	NumberOfShards        *int32
	ShardIDs              []int32
	DomainCache           *DomainCacheInfo
	ShardControllerStatus *string
	Address               *string
}

type DescribeQueueRequest struct {
	ShardID     *int32
	ClusterName *string
	Type        *int32
}

type DescribeQueueResponse struct {
	ProcessingQueueStates []string
}

type DescribeTaskListRequest struct {
	Domain                *string
	TaskList              *TaskList
	TaskListType          *TaskListType
	IncludeTaskListStatus *bool
}

type DescribeTaskListResponse struct {
	Pollers        []*PollerInfo
	TaskListStatus *TaskListStatus
}

type DescribeWorkflowExecutionRequest struct {
	Domain    *string
	Execution *WorkflowExecution
}

type DescribeWorkflowExecutionResponse struct {
	ExecutionConfiguration *WorkflowExecutionConfiguration
	WorkflowExecutionInfo  *WorkflowExecutionInfo
	PendingActivities      []*PendingActivityInfo
	PendingChildren        []*PendingChildExecutionInfo
}

type DomainAlreadyExistsError struct {
	Message string
}

type DomainCacheInfo struct {
	NumOfItemsInCacheByID   *int64
	NumOfItemsInCacheByName *int64
}

type DomainConfiguration struct {
	WorkflowExecutionRetentionPeriodInDays *int32
	EmitMetric                             *bool
	BadBinaries                            *BadBinaries
	HistoryArchivalStatus                  *ArchivalStatus
	HistoryArchivalURI                     *string
	VisibilityArchivalStatus               *ArchivalStatus
	VisibilityArchivalURI                  *string
}

type DomainInfo struct {
	Name        *string
	Status      *DomainStatus
	Description *string
	OwnerEmail  *string
	Data        map[string]string
	UUID        *string
}

type DomainNotActiveError struct {
	Message        string
	DomainName     string
	CurrentCluster string
	ActiveCluster  string
}

type DomainReplicationConfiguration struct {
	ActiveClusterName *string
	Clusters          []*ClusterReplicationConfiguration
}

type DomainStatus int32

const (
	DomainStatusDeleted DomainStatus = iota
	DomainStatusDeprecated
	DomainStatusRegistered
)

type EncodingType int32

const (
	EncodingTypeJSON EncodingType = iota
	EncodingTypeThriftRW
)

type EntityNotExistsError struct {
	Message        string
	CurrentCluster *string
	ActiveCluster  *string
}

type EventType int32

const (
	EventTypeActivityTaskCancelRequested EventType = iota
	EventTypeActivityTaskCanceled
	EventTypeActivityTaskCompleted
	EventTypeActivityTaskFailed
	EventTypeActivityTaskScheduled
	EventTypeActivityTaskStarted
	EventTypeActivityTaskTimedOut
	EventTypeCancelTimerFailed
	EventTypeChildWorkflowExecutionCanceled
	EventTypeChildWorkflowExecutionCompleted
	EventTypeChildWorkflowExecutionFailed
	EventTypeChildWorkflowExecutionStarted
	EventTypeChildWorkflowExecutionTerminated
	EventTypeChildWorkflowExecutionTimedOut
	EventTypeDecisionTaskCompleted
	EventTypeDecisionTaskFailed
	EventTypeDecisionTaskScheduled
	EventTypeDecisionTaskStarted
	EventTypeDecisionTaskTimedOut
	EventTypeExternalWorkflowExecutionCancelRequested
	EventTypeExternalWorkflowExecutionSignaled
	EventTypeMarkerRecorded
	EventTypeRequestCancelActivityTaskFailed
	EventTypeRequestCancelExternalWorkflowExecutionFailed
	EventTypeRequestCancelExternalWorkflowExecutionInitiated
	EventTypeSignalExternalWorkflowExecutionFailed
	EventTypeSignalExternalWorkflowExecutionInitiated
	EventTypeStartChildWorkflowExecutionFailed
	EventTypeStartChildWorkflowExecutionInitiated
	EventTypeTimerCanceled
	EventTypeTimerFired
	EventTypeTimerStarted
	EventTypeUpsertWorkflowSearchAttributes
	EventTypeWorkflowExecutionCancelRequested
	EventTypeWorkflowExecutionCanceled
	EventTypeWorkflowExecutionCompleted
	EventTypeWorkflowExecutionContinuedAsNew
	EventTypeWorkflowExecutionFailed
	EventTypeWorkflowExecutionSignaled
	EventTypeWorkflowExecutionStarted
	EventTypeWorkflowExecutionTerminated
	EventTypeWorkflowExecutionTimedOut
)

type ExternalWorkflowExecutionCancelRequestedEventAttributes struct {
	InitiatedEventId  *int64
	Domain            *string
	WorkflowExecution *WorkflowExecution
}

type ExternalWorkflowExecutionSignaledEventAttributes struct {
	InitiatedEventId  *int64
	Domain            *string
	WorkflowExecution *WorkflowExecution
	Control           []byte
}

type FailWorkflowExecutionDecisionAttributes struct {
	Reason  *string
	Details []byte
}

type GetSearchAttributesResponse struct {
	Keys map[string]IndexedValueType
}

type GetWorkflowExecutionHistoryRequest struct {
	Domain                 *string
	Execution              *WorkflowExecution
	MaximumPageSize        *int32
	NextPageToken          []byte
	WaitForNewEvent        *bool
	HistoryEventFilterType *HistoryEventFilterType
	SkipArchival           *bool
}

type GetWorkflowExecutionHistoryResponse struct {
	History       *History
	RawHistory    []*DataBlob
	NextPageToken []byte
	Archived      *bool
}

type Header struct {
	Fields map[string][]byte
}

type History struct {
	Events []*HistoryEvent
}

type HistoryBranch struct {
	TreeID    *string
	BranchID  *string
	Ancestors []*HistoryBranchRange
}

type HistoryBranchRange struct {
	BranchID    *string
	BeginNodeID *int64
	EndNodeID   *int64
}

type HistoryEvent struct {
	EventId                                                        *int64
	Timestamp                                                      *int64
	EventType                                                      *EventType
	Version                                                        *int64
	TaskId                                                         *int64
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

type HistoryEventFilterType int32

const (
	HistoryEventFilterTypeAllEvent HistoryEventFilterType = iota
	HistoryEventFilterTypeCloseEvent
)

type IndexedValueType int32

const (
	IndexedValueTypeBool IndexedValueType = iota
	IndexedValueTypeDatetime
	IndexedValueTypeDouble
	IndexedValueTypeInt
	IndexedValueTypeKeyword
	IndexedValueTypeString
)

type InternalDataInconsistencyError struct {
	Message string
}

type InternalServiceError struct {
	Message string
}

type LimitExceededError struct {
	Message string
}

type ListArchivedWorkflowExecutionsRequest struct {
	Domain        *string
	PageSize      *int32
	NextPageToken []byte
	Query         *string
}

type ListArchivedWorkflowExecutionsResponse struct {
	Executions    []*WorkflowExecutionInfo
	NextPageToken []byte
}

type ListClosedWorkflowExecutionsRequest struct {
	Domain          *string
	MaximumPageSize *int32
	NextPageToken   []byte
	StartTimeFilter *StartTimeFilter
	ExecutionFilter *WorkflowExecutionFilter
	TypeFilter      *WorkflowTypeFilter
	StatusFilter    *WorkflowExecutionCloseStatus
}

type ListClosedWorkflowExecutionsResponse struct {
	Executions    []*WorkflowExecutionInfo
	NextPageToken []byte
}

type ListDomainsRequest struct {
	PageSize      *int32
	NextPageToken []byte
}

type ListDomainsResponse struct {
	Domains       []*DescribeDomainResponse
	NextPageToken []byte
}

type ListOpenWorkflowExecutionsRequest struct {
	Domain          *string
	MaximumPageSize *int32
	NextPageToken   []byte
	StartTimeFilter *StartTimeFilter
	ExecutionFilter *WorkflowExecutionFilter
	TypeFilter      *WorkflowTypeFilter
}

type ListOpenWorkflowExecutionsResponse struct {
	Executions    []*WorkflowExecutionInfo
	NextPageToken []byte
}

type ListTaskListPartitionsRequest struct {
	Domain   *string
	TaskList *TaskList
}

type ListTaskListPartitionsResponse struct {
	ActivityTaskListPartitions []*TaskListPartitionMetadata
	DecisionTaskListPartitions []*TaskListPartitionMetadata
}

type ListWorkflowExecutionsRequest struct {
	Domain        *string
	PageSize      *int32
	NextPageToken []byte
	Query         *string
}

type ListWorkflowExecutionsResponse struct {
	Executions    []*WorkflowExecutionInfo
	NextPageToken []byte
}

type MarkerRecordedEventAttributes struct {
	MarkerName                   *string
	Details                      []byte
	DecisionTaskCompletedEventId *int64
	Header                       *Header
}

type Memo struct {
	Fields map[string][]byte
}

type ParentClosePolicy int32

const (
	ParentClosePolicyAbandon ParentClosePolicy = iota
	ParentClosePolicyRequestCancel
	ParentClosePolicyTerminate
)

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

type PendingActivityState int32

const (
	PendingActivityStateCancelRequested PendingActivityState = iota
	PendingActivityStateScheduled
	PendingActivityStateStarted
)

type PendingChildExecutionInfo struct {
	WorkflowID        *string
	RunID             *string
	WorkflowTypName   *string
	InitiatedID       *int64
	ParentClosePolicy *ParentClosePolicy
}

type PollForActivityTaskRequest struct {
	Domain           *string
	TaskList         *TaskList
	Identity         *string
	TaskListMetadata *TaskListMetadata
}

type PollForActivityTaskResponse struct {
	TaskToken                       []byte
	WorkflowExecution               *WorkflowExecution
	ActivityId                      *string
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

type PollForDecisionTaskRequest struct {
	Domain         *string
	TaskList       *TaskList
	Identity       *string
	BinaryChecksum *string
}

type PollForDecisionTaskResponse struct {
	TaskToken                 []byte
	WorkflowExecution         *WorkflowExecution
	WorkflowType              *WorkflowType
	PreviousStartedEventId    *int64
	StartedEventId            *int64
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

type PollerInfo struct {
	LastAccessTime *int64
	Identity       *string
	RatePerSecond  *float64
}

type QueryConsistencyLevel int32

const (
	QueryConsistencyLevelEventual QueryConsistencyLevel = iota
	QueryConsistencyLevelStrong
)

type QueryFailedError struct {
	Message string
}

type QueryRejectCondition int32

const (
	QueryRejectConditionNotCompletedCleanly QueryRejectCondition = iota
	QueryRejectConditionNotOpen
)

type QueryRejected struct {
	CloseStatus *WorkflowExecutionCloseStatus
}

type QueryResultType int32

const (
	QueryResultTypeAnswered QueryResultType = iota
	QueryResultTypeFailed
)

type QueryTaskCompletedType int32

const (
	QueryTaskCompletedTypeCompleted QueryTaskCompletedType = iota
	QueryTaskCompletedTypeFailed
)

type QueryWorkflowRequest struct {
	Domain                *string
	Execution             *WorkflowExecution
	Query                 *WorkflowQuery
	QueryRejectCondition  *QueryRejectCondition
	QueryConsistencyLevel *QueryConsistencyLevel
}

type QueryWorkflowResponse struct {
	QueryResult   []byte
	QueryRejected *QueryRejected
}

type ReapplyEventsRequest struct {
	DomainName        *string
	WorkflowExecution *WorkflowExecution
	Events            *DataBlob
}

type RecordActivityTaskHeartbeatByIDRequest struct {
	Domain     *string
	WorkflowID *string
	RunID      *string
	ActivityID *string
	Details    []byte
	Identity   *string
}

type RecordActivityTaskHeartbeatRequest struct {
	TaskToken []byte
	Details   []byte
	Identity  *string
}

type RecordActivityTaskHeartbeatResponse struct {
	CancelRequested *bool
}

type RecordMarkerDecisionAttributes struct {
	MarkerName *string
	Details    []byte
	Header     *Header
}

type RefreshWorkflowTasksRequest struct {
	Domain    *string
	Execution *WorkflowExecution
}

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

type RemoteSyncMatchedError struct {
	Message string
}

type RemoveTaskRequest struct {
	ShardID             *int32
	Type                *int32
	TaskID              *int64
	VisibilityTimestamp *int64
}

type RequestCancelActivityTaskDecisionAttributes struct {
	ActivityId *string
}

type RequestCancelActivityTaskFailedEventAttributes struct {
	ActivityId                   *string
	Cause                        *string
	DecisionTaskCompletedEventId *int64
}

type RequestCancelExternalWorkflowExecutionDecisionAttributes struct {
	Domain            *string
	WorkflowId        *string
	RunId             *string
	Control           []byte
	ChildWorkflowOnly *bool
}

type RequestCancelExternalWorkflowExecutionFailedEventAttributes struct {
	Cause                        *CancelExternalWorkflowExecutionFailedCause
	DecisionTaskCompletedEventId *int64
	Domain                       *string
	WorkflowExecution            *WorkflowExecution
	InitiatedEventId             *int64
	Control                      []byte
}

type RequestCancelExternalWorkflowExecutionInitiatedEventAttributes struct {
	DecisionTaskCompletedEventId *int64
	Domain                       *string
	WorkflowExecution            *WorkflowExecution
	Control                      []byte
	ChildWorkflowOnly            *bool
}

type RequestCancelWorkflowExecutionRequest struct {
	Domain            *string
	WorkflowExecution *WorkflowExecution
	Identity          *string
	RequestId         *string
}

type ResetPointInfo struct {
	BinaryChecksum           *string
	RunId                    *string
	FirstDecisionCompletedId *int64
	CreatedTimeNano          *int64
	ExpiringTimeNano         *int64
	Resettable               *bool
}

type ResetPoints struct {
	Points []*ResetPointInfo
}

type ResetQueueRequest struct {
	ShardID     *int32
	ClusterName *string
	Type        *int32
}

type ResetStickyTaskListRequest struct {
	Domain    *string
	Execution *WorkflowExecution
}

type ResetStickyTaskListResponse struct {
}

type ResetWorkflowExecutionRequest struct {
	Domain                *string
	WorkflowExecution     *WorkflowExecution
	Reason                *string
	DecisionFinishEventId *int64
	RequestId             *string
}

type ResetWorkflowExecutionResponse struct {
	RunId *string
}

type RespondActivityTaskCanceledByIDRequest struct {
	Domain     *string
	WorkflowID *string
	RunID      *string
	ActivityID *string
	Details    []byte
	Identity   *string
}

type RespondActivityTaskCanceledRequest struct {
	TaskToken []byte
	Details   []byte
	Identity  *string
}

type RespondActivityTaskCompletedByIDRequest struct {
	Domain     *string
	WorkflowID *string
	RunID      *string
	ActivityID *string
	Result     []byte
	Identity   *string
}

type RespondActivityTaskCompletedRequest struct {
	TaskToken []byte
	Result    []byte
	Identity  *string
}

type RespondActivityTaskFailedByIDRequest struct {
	Domain     *string
	WorkflowID *string
	RunID      *string
	ActivityID *string
	Reason     *string
	Details    []byte
	Identity   *string
}

type RespondActivityTaskFailedRequest struct {
	TaskToken []byte
	Reason    *string
	Details   []byte
	Identity  *string
}

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

type RespondDecisionTaskCompletedResponse struct {
	DecisionTask *PollForDecisionTaskResponse
}

type RespondDecisionTaskFailedRequest struct {
	TaskToken      []byte
	Cause          *DecisionTaskFailedCause
	Details        []byte
	Identity       *string
	BinaryChecksum *string
}

type RespondQueryTaskCompletedRequest struct {
	TaskToken         []byte
	CompletedType     *QueryTaskCompletedType
	QueryResult       []byte
	ErrorMessage      *string
	WorkerVersionInfo *WorkerVersionInfo
}

type RetryPolicy struct {
	InitialIntervalInSeconds    *int32
	BackoffCoefficient          *float64
	MaximumIntervalInSeconds    *int32
	MaximumAttempts             *int32
	NonRetriableErrorReasons    []string
	ExpirationIntervalInSeconds *int32
}

type RetryTaskV2Error struct {
	Message           string
	DomainId          *string
	WorkflowId        *string
	RunId             *string
	StartEventId      *int64
	StartEventVersion *int64
	EndEventId        *int64
	EndEventVersion   *int64
}

type ScheduleActivityTaskDecisionAttributes struct {
	ActivityId                    *string
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

type SearchAttributes struct {
	IndexedFields map[string][]byte
}

type ServiceBusyError struct {
	Message string
}

type SignalExternalWorkflowExecutionDecisionAttributes struct {
	Domain            *string
	Execution         *WorkflowExecution
	SignalName        *string
	Input             []byte
	Control           []byte
	ChildWorkflowOnly *bool
}

type SignalExternalWorkflowExecutionFailedCause int32

const (
	SignalExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution SignalExternalWorkflowExecutionFailedCause = iota
)

type SignalExternalWorkflowExecutionFailedEventAttributes struct {
	Cause                        *SignalExternalWorkflowExecutionFailedCause
	DecisionTaskCompletedEventId *int64
	Domain                       *string
	WorkflowExecution            *WorkflowExecution
	InitiatedEventId             *int64
	Control                      []byte
}

type SignalExternalWorkflowExecutionInitiatedEventAttributes struct {
	DecisionTaskCompletedEventId *int64
	Domain                       *string
	WorkflowExecution            *WorkflowExecution
	SignalName                   *string
	Input                        []byte
	Control                      []byte
	ChildWorkflowOnly            *bool
}

type SignalWithStartWorkflowExecutionRequest struct {
	Domain                              *string
	WorkflowId                          *string
	WorkflowType                        *WorkflowType
	TaskList                            *TaskList
	Input                               []byte
	ExecutionStartToCloseTimeoutSeconds *int32
	TaskStartToCloseTimeoutSeconds      *int32
	Identity                            *string
	RequestId                           *string
	WorkflowIdReusePolicy               *WorkflowIdReusePolicy
	SignalName                          *string
	SignalInput                         []byte
	Control                             []byte
	RetryPolicy                         *RetryPolicy
	CronSchedule                        *string
	Memo                                *Memo
	SearchAttributes                    *SearchAttributes
	Header                              *Header
}

type SignalWorkflowExecutionRequest struct {
	Domain            *string
	WorkflowExecution *WorkflowExecution
	SignalName        *string
	Input             []byte
	Identity          *string
	RequestId         *string
	Control           []byte
}

type StartChildWorkflowExecutionDecisionAttributes struct {
	Domain                              *string
	WorkflowId                          *string
	WorkflowType                        *WorkflowType
	TaskList                            *TaskList
	Input                               []byte
	ExecutionStartToCloseTimeoutSeconds *int32
	TaskStartToCloseTimeoutSeconds      *int32
	ParentClosePolicy                   *ParentClosePolicy
	Control                             []byte
	WorkflowIdReusePolicy               *WorkflowIdReusePolicy
	RetryPolicy                         *RetryPolicy
	CronSchedule                        *string
	Header                              *Header
	Memo                                *Memo
	SearchAttributes                    *SearchAttributes
}

type StartChildWorkflowExecutionFailedEventAttributes struct {
	Domain                       *string
	WorkflowId                   *string
	WorkflowType                 *WorkflowType
	Cause                        *ChildWorkflowExecutionFailedCause
	Control                      []byte
	InitiatedEventId             *int64
	DecisionTaskCompletedEventId *int64
}

type StartChildWorkflowExecutionInitiatedEventAttributes struct {
	Domain                              *string
	WorkflowId                          *string
	WorkflowType                        *WorkflowType
	TaskList                            *TaskList
	Input                               []byte
	ExecutionStartToCloseTimeoutSeconds *int32
	TaskStartToCloseTimeoutSeconds      *int32
	ParentClosePolicy                   *ParentClosePolicy
	Control                             []byte
	DecisionTaskCompletedEventId        *int64
	WorkflowIdReusePolicy               *WorkflowIdReusePolicy
	RetryPolicy                         *RetryPolicy
	CronSchedule                        *string
	Header                              *Header
	Memo                                *Memo
	SearchAttributes                    *SearchAttributes
}

type StartTimeFilter struct {
	EarliestTime *int64
	LatestTime   *int64
}

type StartTimerDecisionAttributes struct {
	TimerId                   *string
	StartToFireTimeoutSeconds *int64
}

type StartWorkflowExecutionRequest struct {
	Domain                              *string
	WorkflowId                          *string
	WorkflowType                        *WorkflowType
	TaskList                            *TaskList
	Input                               []byte
	ExecutionStartToCloseTimeoutSeconds *int32
	TaskStartToCloseTimeoutSeconds      *int32
	Identity                            *string
	RequestId                           *string
	WorkflowIdReusePolicy               *WorkflowIdReusePolicy
	RetryPolicy                         *RetryPolicy
	CronSchedule                        *string
	Memo                                *Memo
	SearchAttributes                    *SearchAttributes
	Header                              *Header
}

type StartWorkflowExecutionResponse struct {
	RunId *string
}

type StickyExecutionAttributes struct {
	WorkerTaskList                *TaskList
	ScheduleToStartTimeoutSeconds *int32
}

type SupportedClientVersions struct {
	GoSdk   *string
	JavaSdk *string
}

type TaskIDBlock struct {
	StartID *int64
	EndID   *int64
}

type TaskList struct {
	Name *string
	Kind *TaskListKind
}

type TaskListKind int32

const (
	TaskListKindNormal TaskListKind = iota
	TaskListKindSticky
)

type TaskListMetadata struct {
	MaxTasksPerSecond *float64
}

type TaskListPartitionMetadata struct {
	Key           *string
	OwnerHostName *string
}

type TaskListStatus struct {
	BacklogCountHint *int64
	ReadLevel        *int64
	AckLevel         *int64
	RatePerSecond    *float64
	TaskIDBlock      *TaskIDBlock
}

type TaskListType int32

const (
	TaskListTypeActivity TaskListType = iota
	TaskListTypeDecision
)

type TerminateWorkflowExecutionRequest struct {
	Domain            *string
	WorkflowExecution *WorkflowExecution
	Reason            *string
	Details           []byte
	Identity          *string
}

type TimeoutType int32

const (
	TimeoutTypeHeartbeat TimeoutType = iota
	TimeoutTypeScheduleToClose
	TimeoutTypeScheduleToStart
	TimeoutTypeStartToClose
)

type TimerCanceledEventAttributes struct {
	TimerId                      *string
	StartedEventId               *int64
	DecisionTaskCompletedEventId *int64
	Identity                     *string
}

type TimerFiredEventAttributes struct {
	TimerId        *string
	StartedEventId *int64
}

type TimerStartedEventAttributes struct {
	TimerId                      *string
	StartToFireTimeoutSeconds    *int64
	DecisionTaskCompletedEventId *int64
}

type TransientDecisionInfo struct {
	ScheduledEvent *HistoryEvent
	StartedEvent   *HistoryEvent
}

type UpdateDomainInfo struct {
	Description *string
	OwnerEmail  *string
	Data        map[string]string
}

type UpdateDomainRequest struct {
	Name                     *string
	UpdatedInfo              *UpdateDomainInfo
	Configuration            *DomainConfiguration
	ReplicationConfiguration *DomainReplicationConfiguration
	SecurityToken            *string
	DeleteBadBinary          *string
	FailoverTimeoutInSeconds *int32
}

type UpdateDomainResponse struct {
	DomainInfo               *DomainInfo
	Configuration            *DomainConfiguration
	ReplicationConfiguration *DomainReplicationConfiguration
	FailoverVersion          *int64
	IsGlobalDomain           *bool
}

type UpsertWorkflowSearchAttributesDecisionAttributes struct {
	SearchAttributes *SearchAttributes
}

type UpsertWorkflowSearchAttributesEventAttributes struct {
	DecisionTaskCompletedEventId *int64
	SearchAttributes             *SearchAttributes
}

type VersionHistories struct {
	CurrentVersionHistoryIndex *int32
	Histories                  []*VersionHistory
}

type VersionHistory struct {
	BranchToken []byte
	Items       []*VersionHistoryItem
}

type VersionHistoryItem struct {
	EventID *int64
	Version *int64
}

type WorkerVersionInfo struct {
	Impl           *string
	FeatureVersion *string
}

type WorkflowExecution struct {
	WorkflowId *string
	RunId      *string
}

type WorkflowExecutionAlreadyStartedError struct {
	Message        *string
	StartRequestId *string
	RunId          *string
}

type WorkflowExecutionCancelRequestedEventAttributes struct {
	Cause                     *string
	ExternalInitiatedEventId  *int64
	ExternalWorkflowExecution *WorkflowExecution
	Identity                  *string
}

type WorkflowExecutionCanceledEventAttributes struct {
	DecisionTaskCompletedEventId *int64
	Details                      []byte
}

type WorkflowExecutionCloseStatus int32

const (
	WorkflowExecutionCloseStatusCanceled WorkflowExecutionCloseStatus = iota
	WorkflowExecutionCloseStatusCompleted
	WorkflowExecutionCloseStatusContinuedAsNew
	WorkflowExecutionCloseStatusFailed
	WorkflowExecutionCloseStatusTerminated
	WorkflowExecutionCloseStatusTimedOut
)

type WorkflowExecutionCompletedEventAttributes struct {
	Result                       []byte
	DecisionTaskCompletedEventId *int64
}

type WorkflowExecutionConfiguration struct {
	TaskList                            *TaskList
	ExecutionStartToCloseTimeoutSeconds *int32
	TaskStartToCloseTimeoutSeconds      *int32
}

type WorkflowExecutionContinuedAsNewEventAttributes struct {
	NewExecutionRunId                   *string
	WorkflowType                        *WorkflowType
	TaskList                            *TaskList
	Input                               []byte
	ExecutionStartToCloseTimeoutSeconds *int32
	TaskStartToCloseTimeoutSeconds      *int32
	DecisionTaskCompletedEventId        *int64
	BackoffStartIntervalInSeconds       *int32
	Initiator                           *ContinueAsNewInitiator
	FailureReason                       *string
	FailureDetails                      []byte
	LastCompletionResult                []byte
	Header                              *Header
	Memo                                *Memo
	SearchAttributes                    *SearchAttributes
}

type WorkflowExecutionFailedEventAttributes struct {
	Reason                       *string
	Details                      []byte
	DecisionTaskCompletedEventId *int64
}

type WorkflowExecutionFilter struct {
	WorkflowId *string
	RunId      *string
}

type WorkflowExecutionInfo struct {
	Execution        *WorkflowExecution
	Type             *WorkflowType
	StartTime        *int64
	CloseTime        *int64
	CloseStatus      *WorkflowExecutionCloseStatus
	HistoryLength    *int64
	ParentDomainId   *string
	ParentExecution  *WorkflowExecution
	ExecutionTime    *int64
	Memo             *Memo
	SearchAttributes *SearchAttributes
	AutoResetPoints  *ResetPoints
	TaskList         *string
}

type WorkflowExecutionSignaledEventAttributes struct {
	SignalName *string
	Input      []byte
	Identity   *string
}

type WorkflowExecutionStartedEventAttributes struct {
	WorkflowType                        *WorkflowType
	ParentWorkflowDomain                *string
	ParentWorkflowExecution             *WorkflowExecution
	ParentInitiatedEventId              *int64
	TaskList                            *TaskList
	Input                               []byte
	ExecutionStartToCloseTimeoutSeconds *int32
	TaskStartToCloseTimeoutSeconds      *int32
	ContinuedExecutionRunId             *string
	Initiator                           *ContinueAsNewInitiator
	ContinuedFailureReason              *string
	ContinuedFailureDetails             []byte
	LastCompletionResult                []byte
	OriginalExecutionRunId              *string
	Identity                            *string
	FirstExecutionRunId                 *string
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

type WorkflowExecutionTerminatedEventAttributes struct {
	Reason   *string
	Details  []byte
	Identity *string
}

type WorkflowExecutionTimedOutEventAttributes struct {
	TimeoutType *TimeoutType
}

type WorkflowIdReusePolicy int32

const (
	WorkflowIdReusePolicyAllowDuplicate WorkflowIdReusePolicy = iota
	WorkflowIdReusePolicyAllowDuplicateFailedOnly
	WorkflowIdReusePolicyRejectDuplicate
	WorkflowIdReusePolicyTerminateIfRunning
)

type WorkflowQuery struct {
	QueryType *string
	QueryArgs []byte
}

type WorkflowQueryResult struct {
	ResultType   *QueryResultType
	Answer       []byte
	ErrorMessage *string
}

type WorkflowType struct {
	Name *string
}

type WorkflowTypeFilter struct {
	Name *string
}
