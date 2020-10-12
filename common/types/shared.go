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

func (v *AccessDeniedError) GetMessage() (o string) {
	if v != nil {
		return v.Message
	}
	return
}

type ActivityTaskCancelRequestedEventAttributes struct {
	ActivityId                   *string
	DecisionTaskCompletedEventId *int64
}

func (v *ActivityTaskCancelRequestedEventAttributes) GetActivityId() (o string) {
	if v != nil && v.ActivityId != nil {
		return *v.ActivityId
	}
	return
}

func (v *ActivityTaskCancelRequestedEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
}

type ActivityTaskCanceledEventAttributes struct {
	Details                      []byte
	LatestCancelRequestedEventId *int64
	ScheduledEventId             *int64
	StartedEventId               *int64
	Identity                     *string
}

func (v *ActivityTaskCanceledEventAttributes) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *ActivityTaskCanceledEventAttributes) GetLatestCancelRequestedEventId() (o int64) {
	if v != nil && v.LatestCancelRequestedEventId != nil {
		return *v.LatestCancelRequestedEventId
	}
	return
}

func (v *ActivityTaskCanceledEventAttributes) GetScheduledEventId() (o int64) {
	if v != nil && v.ScheduledEventId != nil {
		return *v.ScheduledEventId
	}
	return
}

func (v *ActivityTaskCanceledEventAttributes) GetStartedEventId() (o int64) {
	if v != nil && v.StartedEventId != nil {
		return *v.StartedEventId
	}
	return
}

func (v *ActivityTaskCanceledEventAttributes) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

type ActivityTaskCompletedEventAttributes struct {
	Result           []byte
	ScheduledEventId *int64
	StartedEventId   *int64
	Identity         *string
}

func (v *ActivityTaskCompletedEventAttributes) GetResult() (o []byte) {
	if v != nil && v.Result != nil {
		return v.Result
	}
	return
}

func (v *ActivityTaskCompletedEventAttributes) GetScheduledEventId() (o int64) {
	if v != nil && v.ScheduledEventId != nil {
		return *v.ScheduledEventId
	}
	return
}

func (v *ActivityTaskCompletedEventAttributes) GetStartedEventId() (o int64) {
	if v != nil && v.StartedEventId != nil {
		return *v.StartedEventId
	}
	return
}

func (v *ActivityTaskCompletedEventAttributes) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

type ActivityTaskFailedEventAttributes struct {
	Reason           *string
	Details          []byte
	ScheduledEventId *int64
	StartedEventId   *int64
	Identity         *string
}

func (v *ActivityTaskFailedEventAttributes) GetReason() (o string) {
	if v != nil && v.Reason != nil {
		return *v.Reason
	}
	return
}

func (v *ActivityTaskFailedEventAttributes) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *ActivityTaskFailedEventAttributes) GetScheduledEventId() (o int64) {
	if v != nil && v.ScheduledEventId != nil {
		return *v.ScheduledEventId
	}
	return
}

func (v *ActivityTaskFailedEventAttributes) GetStartedEventId() (o int64) {
	if v != nil && v.StartedEventId != nil {
		return *v.StartedEventId
	}
	return
}

func (v *ActivityTaskFailedEventAttributes) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
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

func (v *ActivityTaskScheduledEventAttributes) GetActivityId() (o string) {
	if v != nil && v.ActivityId != nil {
		return *v.ActivityId
	}
	return
}

func (v *ActivityTaskScheduledEventAttributes) GetActivityType() (o *ActivityType) {
	if v != nil && v.ActivityType != nil {
		return v.ActivityType
	}
	return
}

func (v *ActivityTaskScheduledEventAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *ActivityTaskScheduledEventAttributes) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

func (v *ActivityTaskScheduledEventAttributes) GetInput() (o []byte) {
	if v != nil && v.Input != nil {
		return v.Input
	}
	return
}

func (v *ActivityTaskScheduledEventAttributes) GetScheduleToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.ScheduleToCloseTimeoutSeconds != nil {
		return *v.ScheduleToCloseTimeoutSeconds
	}
	return
}

func (v *ActivityTaskScheduledEventAttributes) GetScheduleToStartTimeoutSeconds() (o int32) {
	if v != nil && v.ScheduleToStartTimeoutSeconds != nil {
		return *v.ScheduleToStartTimeoutSeconds
	}
	return
}

func (v *ActivityTaskScheduledEventAttributes) GetStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.StartToCloseTimeoutSeconds != nil {
		return *v.StartToCloseTimeoutSeconds
	}
	return
}

func (v *ActivityTaskScheduledEventAttributes) GetHeartbeatTimeoutSeconds() (o int32) {
	if v != nil && v.HeartbeatTimeoutSeconds != nil {
		return *v.HeartbeatTimeoutSeconds
	}
	return
}

func (v *ActivityTaskScheduledEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
}

func (v *ActivityTaskScheduledEventAttributes) GetRetryPolicy() (o *RetryPolicy) {
	if v != nil && v.RetryPolicy != nil {
		return v.RetryPolicy
	}
	return
}

func (v *ActivityTaskScheduledEventAttributes) GetHeader() (o *Header) {
	if v != nil && v.Header != nil {
		return v.Header
	}
	return
}

type ActivityTaskStartedEventAttributes struct {
	ScheduledEventId   *int64
	Identity           *string
	RequestId          *string
	Attempt            *int32
	LastFailureReason  *string
	LastFailureDetails []byte
}

func (v *ActivityTaskStartedEventAttributes) GetScheduledEventId() (o int64) {
	if v != nil && v.ScheduledEventId != nil {
		return *v.ScheduledEventId
	}
	return
}

func (v *ActivityTaskStartedEventAttributes) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

func (v *ActivityTaskStartedEventAttributes) GetRequestId() (o string) {
	if v != nil && v.RequestId != nil {
		return *v.RequestId
	}
	return
}

func (v *ActivityTaskStartedEventAttributes) GetAttempt() (o int32) {
	if v != nil && v.Attempt != nil {
		return *v.Attempt
	}
	return
}

func (v *ActivityTaskStartedEventAttributes) GetLastFailureReason() (o string) {
	if v != nil && v.LastFailureReason != nil {
		return *v.LastFailureReason
	}
	return
}

func (v *ActivityTaskStartedEventAttributes) GetLastFailureDetails() (o []byte) {
	if v != nil && v.LastFailureDetails != nil {
		return v.LastFailureDetails
	}
	return
}

type ActivityTaskTimedOutEventAttributes struct {
	Details            []byte
	ScheduledEventId   *int64
	StartedEventId     *int64
	TimeoutType        *TimeoutType
	LastFailureReason  *string
	LastFailureDetails []byte
}

func (v *ActivityTaskTimedOutEventAttributes) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *ActivityTaskTimedOutEventAttributes) GetScheduledEventId() (o int64) {
	if v != nil && v.ScheduledEventId != nil {
		return *v.ScheduledEventId
	}
	return
}

func (v *ActivityTaskTimedOutEventAttributes) GetStartedEventId() (o int64) {
	if v != nil && v.StartedEventId != nil {
		return *v.StartedEventId
	}
	return
}

func (v *ActivityTaskTimedOutEventAttributes) GetTimeoutType() (o TimeoutType) {
	if v != nil && v.TimeoutType != nil {
		return *v.TimeoutType
	}
	return
}

func (v *ActivityTaskTimedOutEventAttributes) GetLastFailureReason() (o string) {
	if v != nil && v.LastFailureReason != nil {
		return *v.LastFailureReason
	}
	return
}

func (v *ActivityTaskTimedOutEventAttributes) GetLastFailureDetails() (o []byte) {
	if v != nil && v.LastFailureDetails != nil {
		return v.LastFailureDetails
	}
	return
}

type ActivityType struct {
	Name *string
}

func (v *ActivityType) GetName() (o string) {
	if v != nil && v.Name != nil {
		return *v.Name
	}
	return
}

type ArchivalStatus int32

const (
	ArchivalStatusDisabled ArchivalStatus = iota
	ArchivalStatusEnabled
)

type BadBinaries struct {
	Binaries map[string]*BadBinaryInfo
}

func (v *BadBinaries) GetBinaries() (o map[string]*BadBinaryInfo) {
	if v != nil && v.Binaries != nil {
		return v.Binaries
	}
	return
}

type BadBinaryInfo struct {
	Reason          *string
	Operator        *string
	CreatedTimeNano *int64
}

func (v *BadBinaryInfo) GetReason() (o string) {
	if v != nil && v.Reason != nil {
		return *v.Reason
	}
	return
}

func (v *BadBinaryInfo) GetOperator() (o string) {
	if v != nil && v.Operator != nil {
		return *v.Operator
	}
	return
}

func (v *BadBinaryInfo) GetCreatedTimeNano() (o int64) {
	if v != nil && v.CreatedTimeNano != nil {
		return *v.CreatedTimeNano
	}
	return
}

type BadRequestError struct {
	Message string
}

func (v *BadRequestError) GetMessage() (o string) {
	if v != nil {
		return v.Message
	}
	return
}

type CancelExternalWorkflowExecutionFailedCause int32

const (
	CancelExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution CancelExternalWorkflowExecutionFailedCause = iota
)

type CancelTimerDecisionAttributes struct {
	TimerId *string
}

func (v *CancelTimerDecisionAttributes) GetTimerId() (o string) {
	if v != nil && v.TimerId != nil {
		return *v.TimerId
	}
	return
}

type CancelTimerFailedEventAttributes struct {
	TimerId                      *string
	Cause                        *string
	DecisionTaskCompletedEventId *int64
	Identity                     *string
}

func (v *CancelTimerFailedEventAttributes) GetTimerId() (o string) {
	if v != nil && v.TimerId != nil {
		return *v.TimerId
	}
	return
}

func (v *CancelTimerFailedEventAttributes) GetCause() (o string) {
	if v != nil && v.Cause != nil {
		return *v.Cause
	}
	return
}

func (v *CancelTimerFailedEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
}

func (v *CancelTimerFailedEventAttributes) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

type CancelWorkflowExecutionDecisionAttributes struct {
	Details []byte
}

func (v *CancelWorkflowExecutionDecisionAttributes) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

type CancellationAlreadyRequestedError struct {
	Message string
}

func (v *CancellationAlreadyRequestedError) GetMessage() (o string) {
	if v != nil {
		return v.Message
	}
	return
}

type ChildWorkflowExecutionCanceledEventAttributes struct {
	Details           []byte
	Domain            *string
	WorkflowExecution *WorkflowExecution
	WorkflowType      *WorkflowType
	InitiatedEventId  *int64
	StartedEventId    *int64
}

func (v *ChildWorkflowExecutionCanceledEventAttributes) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *ChildWorkflowExecutionCanceledEventAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *ChildWorkflowExecutionCanceledEventAttributes) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *ChildWorkflowExecutionCanceledEventAttributes) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

func (v *ChildWorkflowExecutionCanceledEventAttributes) GetInitiatedEventId() (o int64) {
	if v != nil && v.InitiatedEventId != nil {
		return *v.InitiatedEventId
	}
	return
}

func (v *ChildWorkflowExecutionCanceledEventAttributes) GetStartedEventId() (o int64) {
	if v != nil && v.StartedEventId != nil {
		return *v.StartedEventId
	}
	return
}

type ChildWorkflowExecutionCompletedEventAttributes struct {
	Result            []byte
	Domain            *string
	WorkflowExecution *WorkflowExecution
	WorkflowType      *WorkflowType
	InitiatedEventId  *int64
	StartedEventId    *int64
}

func (v *ChildWorkflowExecutionCompletedEventAttributes) GetResult() (o []byte) {
	if v != nil && v.Result != nil {
		return v.Result
	}
	return
}

func (v *ChildWorkflowExecutionCompletedEventAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *ChildWorkflowExecutionCompletedEventAttributes) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *ChildWorkflowExecutionCompletedEventAttributes) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

func (v *ChildWorkflowExecutionCompletedEventAttributes) GetInitiatedEventId() (o int64) {
	if v != nil && v.InitiatedEventId != nil {
		return *v.InitiatedEventId
	}
	return
}

func (v *ChildWorkflowExecutionCompletedEventAttributes) GetStartedEventId() (o int64) {
	if v != nil && v.StartedEventId != nil {
		return *v.StartedEventId
	}
	return
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

func (v *ChildWorkflowExecutionFailedEventAttributes) GetReason() (o string) {
	if v != nil && v.Reason != nil {
		return *v.Reason
	}
	return
}

func (v *ChildWorkflowExecutionFailedEventAttributes) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *ChildWorkflowExecutionFailedEventAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *ChildWorkflowExecutionFailedEventAttributes) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *ChildWorkflowExecutionFailedEventAttributes) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

func (v *ChildWorkflowExecutionFailedEventAttributes) GetInitiatedEventId() (o int64) {
	if v != nil && v.InitiatedEventId != nil {
		return *v.InitiatedEventId
	}
	return
}

func (v *ChildWorkflowExecutionFailedEventAttributes) GetStartedEventId() (o int64) {
	if v != nil && v.StartedEventId != nil {
		return *v.StartedEventId
	}
	return
}

type ChildWorkflowExecutionStartedEventAttributes struct {
	Domain            *string
	InitiatedEventId  *int64
	WorkflowExecution *WorkflowExecution
	WorkflowType      *WorkflowType
	Header            *Header
}

func (v *ChildWorkflowExecutionStartedEventAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *ChildWorkflowExecutionStartedEventAttributes) GetInitiatedEventId() (o int64) {
	if v != nil && v.InitiatedEventId != nil {
		return *v.InitiatedEventId
	}
	return
}

func (v *ChildWorkflowExecutionStartedEventAttributes) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *ChildWorkflowExecutionStartedEventAttributes) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

func (v *ChildWorkflowExecutionStartedEventAttributes) GetHeader() (o *Header) {
	if v != nil && v.Header != nil {
		return v.Header
	}
	return
}

type ChildWorkflowExecutionTerminatedEventAttributes struct {
	Domain            *string
	WorkflowExecution *WorkflowExecution
	WorkflowType      *WorkflowType
	InitiatedEventId  *int64
	StartedEventId    *int64
}

func (v *ChildWorkflowExecutionTerminatedEventAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *ChildWorkflowExecutionTerminatedEventAttributes) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *ChildWorkflowExecutionTerminatedEventAttributes) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

func (v *ChildWorkflowExecutionTerminatedEventAttributes) GetInitiatedEventId() (o int64) {
	if v != nil && v.InitiatedEventId != nil {
		return *v.InitiatedEventId
	}
	return
}

func (v *ChildWorkflowExecutionTerminatedEventAttributes) GetStartedEventId() (o int64) {
	if v != nil && v.StartedEventId != nil {
		return *v.StartedEventId
	}
	return
}

type ChildWorkflowExecutionTimedOutEventAttributes struct {
	TimeoutType       *TimeoutType
	Domain            *string
	WorkflowExecution *WorkflowExecution
	WorkflowType      *WorkflowType
	InitiatedEventId  *int64
	StartedEventId    *int64
}

func (v *ChildWorkflowExecutionTimedOutEventAttributes) GetTimeoutType() (o TimeoutType) {
	if v != nil && v.TimeoutType != nil {
		return *v.TimeoutType
	}
	return
}

func (v *ChildWorkflowExecutionTimedOutEventAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *ChildWorkflowExecutionTimedOutEventAttributes) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *ChildWorkflowExecutionTimedOutEventAttributes) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

func (v *ChildWorkflowExecutionTimedOutEventAttributes) GetInitiatedEventId() (o int64) {
	if v != nil && v.InitiatedEventId != nil {
		return *v.InitiatedEventId
	}
	return
}

func (v *ChildWorkflowExecutionTimedOutEventAttributes) GetStartedEventId() (o int64) {
	if v != nil && v.StartedEventId != nil {
		return *v.StartedEventId
	}
	return
}

type ClientVersionNotSupportedError struct {
	FeatureVersion    string
	ClientImpl        string
	SupportedVersions string
}

func (v *ClientVersionNotSupportedError) GetFeatureVersion() (o string) {
	if v != nil {
		return v.FeatureVersion
	}
	return
}

func (v *ClientVersionNotSupportedError) GetClientImpl() (o string) {
	if v != nil {
		return v.ClientImpl
	}
	return
}

func (v *ClientVersionNotSupportedError) GetSupportedVersions() (o string) {
	if v != nil {
		return v.SupportedVersions
	}
	return
}

type CloseShardRequest struct {
	ShardID *int32
}

func (v *CloseShardRequest) GetShardID() (o int32) {
	if v != nil && v.ShardID != nil {
		return *v.ShardID
	}
	return
}

type ClusterInfo struct {
	SupportedClientVersions *SupportedClientVersions
}

func (v *ClusterInfo) GetSupportedClientVersions() (o *SupportedClientVersions) {
	if v != nil && v.SupportedClientVersions != nil {
		return v.SupportedClientVersions
	}
	return
}

type ClusterReplicationConfiguration struct {
	ClusterName *string
}

func (v *ClusterReplicationConfiguration) GetClusterName() (o string) {
	if v != nil && v.ClusterName != nil {
		return *v.ClusterName
	}
	return
}

type CompleteWorkflowExecutionDecisionAttributes struct {
	Result []byte
}

func (v *CompleteWorkflowExecutionDecisionAttributes) GetResult() (o []byte) {
	if v != nil && v.Result != nil {
		return v.Result
	}
	return
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

func (v *ContinueAsNewWorkflowExecutionDecisionAttributes) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

func (v *ContinueAsNewWorkflowExecutionDecisionAttributes) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

func (v *ContinueAsNewWorkflowExecutionDecisionAttributes) GetInput() (o []byte) {
	if v != nil && v.Input != nil {
		return v.Input
	}
	return
}

func (v *ContinueAsNewWorkflowExecutionDecisionAttributes) GetExecutionStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.ExecutionStartToCloseTimeoutSeconds != nil {
		return *v.ExecutionStartToCloseTimeoutSeconds
	}
	return
}

func (v *ContinueAsNewWorkflowExecutionDecisionAttributes) GetTaskStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.TaskStartToCloseTimeoutSeconds != nil {
		return *v.TaskStartToCloseTimeoutSeconds
	}
	return
}

func (v *ContinueAsNewWorkflowExecutionDecisionAttributes) GetBackoffStartIntervalInSeconds() (o int32) {
	if v != nil && v.BackoffStartIntervalInSeconds != nil {
		return *v.BackoffStartIntervalInSeconds
	}
	return
}

func (v *ContinueAsNewWorkflowExecutionDecisionAttributes) GetRetryPolicy() (o *RetryPolicy) {
	if v != nil && v.RetryPolicy != nil {
		return v.RetryPolicy
	}
	return
}

func (v *ContinueAsNewWorkflowExecutionDecisionAttributes) GetInitiator() (o ContinueAsNewInitiator) {
	if v != nil && v.Initiator != nil {
		return *v.Initiator
	}
	return
}

func (v *ContinueAsNewWorkflowExecutionDecisionAttributes) GetFailureReason() (o string) {
	if v != nil && v.FailureReason != nil {
		return *v.FailureReason
	}
	return
}

func (v *ContinueAsNewWorkflowExecutionDecisionAttributes) GetFailureDetails() (o []byte) {
	if v != nil && v.FailureDetails != nil {
		return v.FailureDetails
	}
	return
}

func (v *ContinueAsNewWorkflowExecutionDecisionAttributes) GetLastCompletionResult() (o []byte) {
	if v != nil && v.LastCompletionResult != nil {
		return v.LastCompletionResult
	}
	return
}

func (v *ContinueAsNewWorkflowExecutionDecisionAttributes) GetCronSchedule() (o string) {
	if v != nil && v.CronSchedule != nil {
		return *v.CronSchedule
	}
	return
}

func (v *ContinueAsNewWorkflowExecutionDecisionAttributes) GetHeader() (o *Header) {
	if v != nil && v.Header != nil {
		return v.Header
	}
	return
}

func (v *ContinueAsNewWorkflowExecutionDecisionAttributes) GetMemo() (o *Memo) {
	if v != nil && v.Memo != nil {
		return v.Memo
	}
	return
}

func (v *ContinueAsNewWorkflowExecutionDecisionAttributes) GetSearchAttributes() (o *SearchAttributes) {
	if v != nil && v.SearchAttributes != nil {
		return v.SearchAttributes
	}
	return
}

type CountWorkflowExecutionsRequest struct {
	Domain *string
	Query  *string
}

func (v *CountWorkflowExecutionsRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *CountWorkflowExecutionsRequest) GetQuery() (o string) {
	if v != nil && v.Query != nil {
		return *v.Query
	}
	return
}

type CountWorkflowExecutionsResponse struct {
	Count *int64
}

func (v *CountWorkflowExecutionsResponse) GetCount() (o int64) {
	if v != nil && v.Count != nil {
		return *v.Count
	}
	return
}

type CurrentBranchChangedError struct {
	Message            string
	CurrentBranchToken []byte
}

func (v *CurrentBranchChangedError) GetMessage() (o string) {
	if v != nil {
		return v.Message
	}
	return
}

func (v *CurrentBranchChangedError) GetCurrentBranchToken() (o []byte) {
	if v != nil && v.CurrentBranchToken != nil {
		return v.CurrentBranchToken
	}
	return
}

type DataBlob struct {
	EncodingType *EncodingType
	Data         []byte
}

func (v *DataBlob) GetEncodingType() (o EncodingType) {
	if v != nil && v.EncodingType != nil {
		return *v.EncodingType
	}
	return
}

func (v *DataBlob) GetData() (o []byte) {
	if v != nil && v.Data != nil {
		return v.Data
	}
	return
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

func (v *Decision) GetDecisionType() (o DecisionType) {
	if v != nil && v.DecisionType != nil {
		return *v.DecisionType
	}
	return
}

func (v *Decision) GetScheduleActivityTaskDecisionAttributes() (o *ScheduleActivityTaskDecisionAttributes) {
	if v != nil && v.ScheduleActivityTaskDecisionAttributes != nil {
		return v.ScheduleActivityTaskDecisionAttributes
	}
	return
}

func (v *Decision) GetStartTimerDecisionAttributes() (o *StartTimerDecisionAttributes) {
	if v != nil && v.StartTimerDecisionAttributes != nil {
		return v.StartTimerDecisionAttributes
	}
	return
}

func (v *Decision) GetCompleteWorkflowExecutionDecisionAttributes() (o *CompleteWorkflowExecutionDecisionAttributes) {
	if v != nil && v.CompleteWorkflowExecutionDecisionAttributes != nil {
		return v.CompleteWorkflowExecutionDecisionAttributes
	}
	return
}

func (v *Decision) GetFailWorkflowExecutionDecisionAttributes() (o *FailWorkflowExecutionDecisionAttributes) {
	if v != nil && v.FailWorkflowExecutionDecisionAttributes != nil {
		return v.FailWorkflowExecutionDecisionAttributes
	}
	return
}

func (v *Decision) GetRequestCancelActivityTaskDecisionAttributes() (o *RequestCancelActivityTaskDecisionAttributes) {
	if v != nil && v.RequestCancelActivityTaskDecisionAttributes != nil {
		return v.RequestCancelActivityTaskDecisionAttributes
	}
	return
}

func (v *Decision) GetCancelTimerDecisionAttributes() (o *CancelTimerDecisionAttributes) {
	if v != nil && v.CancelTimerDecisionAttributes != nil {
		return v.CancelTimerDecisionAttributes
	}
	return
}

func (v *Decision) GetCancelWorkflowExecutionDecisionAttributes() (o *CancelWorkflowExecutionDecisionAttributes) {
	if v != nil && v.CancelWorkflowExecutionDecisionAttributes != nil {
		return v.CancelWorkflowExecutionDecisionAttributes
	}
	return
}

func (v *Decision) GetRequestCancelExternalWorkflowExecutionDecisionAttributes() (o *RequestCancelExternalWorkflowExecutionDecisionAttributes) {
	if v != nil && v.RequestCancelExternalWorkflowExecutionDecisionAttributes != nil {
		return v.RequestCancelExternalWorkflowExecutionDecisionAttributes
	}
	return
}

func (v *Decision) GetRecordMarkerDecisionAttributes() (o *RecordMarkerDecisionAttributes) {
	if v != nil && v.RecordMarkerDecisionAttributes != nil {
		return v.RecordMarkerDecisionAttributes
	}
	return
}

func (v *Decision) GetContinueAsNewWorkflowExecutionDecisionAttributes() (o *ContinueAsNewWorkflowExecutionDecisionAttributes) {
	if v != nil && v.ContinueAsNewWorkflowExecutionDecisionAttributes != nil {
		return v.ContinueAsNewWorkflowExecutionDecisionAttributes
	}
	return
}

func (v *Decision) GetStartChildWorkflowExecutionDecisionAttributes() (o *StartChildWorkflowExecutionDecisionAttributes) {
	if v != nil && v.StartChildWorkflowExecutionDecisionAttributes != nil {
		return v.StartChildWorkflowExecutionDecisionAttributes
	}
	return
}

func (v *Decision) GetSignalExternalWorkflowExecutionDecisionAttributes() (o *SignalExternalWorkflowExecutionDecisionAttributes) {
	if v != nil && v.SignalExternalWorkflowExecutionDecisionAttributes != nil {
		return v.SignalExternalWorkflowExecutionDecisionAttributes
	}
	return
}

func (v *Decision) GetUpsertWorkflowSearchAttributesDecisionAttributes() (o *UpsertWorkflowSearchAttributesDecisionAttributes) {
	if v != nil && v.UpsertWorkflowSearchAttributesDecisionAttributes != nil {
		return v.UpsertWorkflowSearchAttributesDecisionAttributes
	}
	return
}

type DecisionTaskCompletedEventAttributes struct {
	ExecutionContext []byte
	ScheduledEventId *int64
	StartedEventId   *int64
	Identity         *string
	BinaryChecksum   *string
}

func (v *DecisionTaskCompletedEventAttributes) GetExecutionContext() (o []byte) {
	if v != nil && v.ExecutionContext != nil {
		return v.ExecutionContext
	}
	return
}

func (v *DecisionTaskCompletedEventAttributes) GetScheduledEventId() (o int64) {
	if v != nil && v.ScheduledEventId != nil {
		return *v.ScheduledEventId
	}
	return
}

func (v *DecisionTaskCompletedEventAttributes) GetStartedEventId() (o int64) {
	if v != nil && v.StartedEventId != nil {
		return *v.StartedEventId
	}
	return
}

func (v *DecisionTaskCompletedEventAttributes) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

func (v *DecisionTaskCompletedEventAttributes) GetBinaryChecksum() (o string) {
	if v != nil && v.BinaryChecksum != nil {
		return *v.BinaryChecksum
	}
	return
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

func (v *DecisionTaskFailedEventAttributes) GetScheduledEventId() (o int64) {
	if v != nil && v.ScheduledEventId != nil {
		return *v.ScheduledEventId
	}
	return
}

func (v *DecisionTaskFailedEventAttributes) GetStartedEventId() (o int64) {
	if v != nil && v.StartedEventId != nil {
		return *v.StartedEventId
	}
	return
}

func (v *DecisionTaskFailedEventAttributes) GetCause() (o DecisionTaskFailedCause) {
	if v != nil && v.Cause != nil {
		return *v.Cause
	}
	return
}

func (v *DecisionTaskFailedEventAttributes) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *DecisionTaskFailedEventAttributes) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

func (v *DecisionTaskFailedEventAttributes) GetReason() (o string) {
	if v != nil && v.Reason != nil {
		return *v.Reason
	}
	return
}

func (v *DecisionTaskFailedEventAttributes) GetBaseRunId() (o string) {
	if v != nil && v.BaseRunId != nil {
		return *v.BaseRunId
	}
	return
}

func (v *DecisionTaskFailedEventAttributes) GetNewRunId() (o string) {
	if v != nil && v.NewRunId != nil {
		return *v.NewRunId
	}
	return
}

func (v *DecisionTaskFailedEventAttributes) GetForkEventVersion() (o int64) {
	if v != nil && v.ForkEventVersion != nil {
		return *v.ForkEventVersion
	}
	return
}

func (v *DecisionTaskFailedEventAttributes) GetBinaryChecksum() (o string) {
	if v != nil && v.BinaryChecksum != nil {
		return *v.BinaryChecksum
	}
	return
}

type DecisionTaskScheduledEventAttributes struct {
	TaskList                   *TaskList
	StartToCloseTimeoutSeconds *int32
	Attempt                    *int64
}

func (v *DecisionTaskScheduledEventAttributes) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

func (v *DecisionTaskScheduledEventAttributes) GetStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.StartToCloseTimeoutSeconds != nil {
		return *v.StartToCloseTimeoutSeconds
	}
	return
}

func (v *DecisionTaskScheduledEventAttributes) GetAttempt() (o int64) {
	if v != nil && v.Attempt != nil {
		return *v.Attempt
	}
	return
}

type DecisionTaskStartedEventAttributes struct {
	ScheduledEventId *int64
	Identity         *string
	RequestId        *string
}

func (v *DecisionTaskStartedEventAttributes) GetScheduledEventId() (o int64) {
	if v != nil && v.ScheduledEventId != nil {
		return *v.ScheduledEventId
	}
	return
}

func (v *DecisionTaskStartedEventAttributes) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

func (v *DecisionTaskStartedEventAttributes) GetRequestId() (o string) {
	if v != nil && v.RequestId != nil {
		return *v.RequestId
	}
	return
}

type DecisionTaskTimedOutEventAttributes struct {
	ScheduledEventId *int64
	StartedEventId   *int64
	TimeoutType      *TimeoutType
}

func (v *DecisionTaskTimedOutEventAttributes) GetScheduledEventId() (o int64) {
	if v != nil && v.ScheduledEventId != nil {
		return *v.ScheduledEventId
	}
	return
}

func (v *DecisionTaskTimedOutEventAttributes) GetStartedEventId() (o int64) {
	if v != nil && v.StartedEventId != nil {
		return *v.StartedEventId
	}
	return
}

func (v *DecisionTaskTimedOutEventAttributes) GetTimeoutType() (o TimeoutType) {
	if v != nil && v.TimeoutType != nil {
		return *v.TimeoutType
	}
	return
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

func (v *DeprecateDomainRequest) GetName() (o string) {
	if v != nil && v.Name != nil {
		return *v.Name
	}
	return
}

func (v *DeprecateDomainRequest) GetSecurityToken() (o string) {
	if v != nil && v.SecurityToken != nil {
		return *v.SecurityToken
	}
	return
}

type DescribeDomainRequest struct {
	Name *string
	UUID *string
}

func (v *DescribeDomainRequest) GetName() (o string) {
	if v != nil && v.Name != nil {
		return *v.Name
	}
	return
}

func (v *DescribeDomainRequest) GetUUID() (o string) {
	if v != nil && v.UUID != nil {
		return *v.UUID
	}
	return
}

type DescribeDomainResponse struct {
	DomainInfo               *DomainInfo
	Configuration            *DomainConfiguration
	ReplicationConfiguration *DomainReplicationConfiguration
	FailoverVersion          *int64
	IsGlobalDomain           *bool
}

func (v *DescribeDomainResponse) GetDomainInfo() (o *DomainInfo) {
	if v != nil && v.DomainInfo != nil {
		return v.DomainInfo
	}
	return
}

func (v *DescribeDomainResponse) GetConfiguration() (o *DomainConfiguration) {
	if v != nil && v.Configuration != nil {
		return v.Configuration
	}
	return
}

func (v *DescribeDomainResponse) GetReplicationConfiguration() (o *DomainReplicationConfiguration) {
	if v != nil && v.ReplicationConfiguration != nil {
		return v.ReplicationConfiguration
	}
	return
}

func (v *DescribeDomainResponse) GetFailoverVersion() (o int64) {
	if v != nil && v.FailoverVersion != nil {
		return *v.FailoverVersion
	}
	return
}

func (v *DescribeDomainResponse) GetIsGlobalDomain() (o bool) {
	if v != nil && v.IsGlobalDomain != nil {
		return *v.IsGlobalDomain
	}
	return
}

type DescribeHistoryHostRequest struct {
	HostAddress      *string
	ShardIdForHost   *int32
	ExecutionForHost *WorkflowExecution
}

func (v *DescribeHistoryHostRequest) GetHostAddress() (o string) {
	if v != nil && v.HostAddress != nil {
		return *v.HostAddress
	}
	return
}

func (v *DescribeHistoryHostRequest) GetShardIdForHost() (o int32) {
	if v != nil && v.ShardIdForHost != nil {
		return *v.ShardIdForHost
	}
	return
}

func (v *DescribeHistoryHostRequest) GetExecutionForHost() (o *WorkflowExecution) {
	if v != nil && v.ExecutionForHost != nil {
		return v.ExecutionForHost
	}
	return
}

type DescribeHistoryHostResponse struct {
	NumberOfShards        *int32
	ShardIDs              []int32
	DomainCache           *DomainCacheInfo
	ShardControllerStatus *string
	Address               *string
}

func (v *DescribeHistoryHostResponse) GetNumberOfShards() (o int32) {
	if v != nil && v.NumberOfShards != nil {
		return *v.NumberOfShards
	}
	return
}

func (v *DescribeHistoryHostResponse) GetShardIDs() (o []int32) {
	if v != nil && v.ShardIDs != nil {
		return v.ShardIDs
	}
	return
}

func (v *DescribeHistoryHostResponse) GetDomainCache() (o *DomainCacheInfo) {
	if v != nil && v.DomainCache != nil {
		return v.DomainCache
	}
	return
}

func (v *DescribeHistoryHostResponse) GetShardControllerStatus() (o string) {
	if v != nil && v.ShardControllerStatus != nil {
		return *v.ShardControllerStatus
	}
	return
}

func (v *DescribeHistoryHostResponse) GetAddress() (o string) {
	if v != nil && v.Address != nil {
		return *v.Address
	}
	return
}

type DescribeQueueRequest struct {
	ShardID     *int32
	ClusterName *string
	Type        *int32
}

func (v *DescribeQueueRequest) GetShardID() (o int32) {
	if v != nil && v.ShardID != nil {
		return *v.ShardID
	}
	return
}

func (v *DescribeQueueRequest) GetClusterName() (o string) {
	if v != nil && v.ClusterName != nil {
		return *v.ClusterName
	}
	return
}

func (v *DescribeQueueRequest) GetType() (o int32) {
	if v != nil && v.Type != nil {
		return *v.Type
	}
	return
}

type DescribeQueueResponse struct {
	ProcessingQueueStates []string
}

func (v *DescribeQueueResponse) GetProcessingQueueStates() (o []string) {
	if v != nil && v.ProcessingQueueStates != nil {
		return v.ProcessingQueueStates
	}
	return
}

type DescribeTaskListRequest struct {
	Domain                *string
	TaskList              *TaskList
	TaskListType          *TaskListType
	IncludeTaskListStatus *bool
}

func (v *DescribeTaskListRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *DescribeTaskListRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

func (v *DescribeTaskListRequest) GetTaskListType() (o TaskListType) {
	if v != nil && v.TaskListType != nil {
		return *v.TaskListType
	}
	return
}

func (v *DescribeTaskListRequest) GetIncludeTaskListStatus() (o bool) {
	if v != nil && v.IncludeTaskListStatus != nil {
		return *v.IncludeTaskListStatus
	}
	return
}

type DescribeTaskListResponse struct {
	Pollers        []*PollerInfo
	TaskListStatus *TaskListStatus
}

func (v *DescribeTaskListResponse) GetPollers() (o []*PollerInfo) {
	if v != nil && v.Pollers != nil {
		return v.Pollers
	}
	return
}

func (v *DescribeTaskListResponse) GetTaskListStatus() (o *TaskListStatus) {
	if v != nil && v.TaskListStatus != nil {
		return v.TaskListStatus
	}
	return
}

type DescribeWorkflowExecutionRequest struct {
	Domain    *string
	Execution *WorkflowExecution
}

func (v *DescribeWorkflowExecutionRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *DescribeWorkflowExecutionRequest) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
}

type DescribeWorkflowExecutionResponse struct {
	ExecutionConfiguration *WorkflowExecutionConfiguration
	WorkflowExecutionInfo  *WorkflowExecutionInfo
	PendingActivities      []*PendingActivityInfo
	PendingChildren        []*PendingChildExecutionInfo
}

func (v *DescribeWorkflowExecutionResponse) GetExecutionConfiguration() (o *WorkflowExecutionConfiguration) {
	if v != nil && v.ExecutionConfiguration != nil {
		return v.ExecutionConfiguration
	}
	return
}

func (v *DescribeWorkflowExecutionResponse) GetWorkflowExecutionInfo() (o *WorkflowExecutionInfo) {
	if v != nil && v.WorkflowExecutionInfo != nil {
		return v.WorkflowExecutionInfo
	}
	return
}

func (v *DescribeWorkflowExecutionResponse) GetPendingActivities() (o []*PendingActivityInfo) {
	if v != nil && v.PendingActivities != nil {
		return v.PendingActivities
	}
	return
}

func (v *DescribeWorkflowExecutionResponse) GetPendingChildren() (o []*PendingChildExecutionInfo) {
	if v != nil && v.PendingChildren != nil {
		return v.PendingChildren
	}
	return
}

type DomainAlreadyExistsError struct {
	Message string
}

func (v *DomainAlreadyExistsError) GetMessage() (o string) {
	if v != nil {
		return v.Message
	}
	return
}

type DomainCacheInfo struct {
	NumOfItemsInCacheByID   *int64
	NumOfItemsInCacheByName *int64
}

func (v *DomainCacheInfo) GetNumOfItemsInCacheByID() (o int64) {
	if v != nil && v.NumOfItemsInCacheByID != nil {
		return *v.NumOfItemsInCacheByID
	}
	return
}

func (v *DomainCacheInfo) GetNumOfItemsInCacheByName() (o int64) {
	if v != nil && v.NumOfItemsInCacheByName != nil {
		return *v.NumOfItemsInCacheByName
	}
	return
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

func (v *DomainConfiguration) GetWorkflowExecutionRetentionPeriodInDays() (o int32) {
	if v != nil && v.WorkflowExecutionRetentionPeriodInDays != nil {
		return *v.WorkflowExecutionRetentionPeriodInDays
	}
	return
}

func (v *DomainConfiguration) GetEmitMetric() (o bool) {
	if v != nil && v.EmitMetric != nil {
		return *v.EmitMetric
	}
	return
}

func (v *DomainConfiguration) GetBadBinaries() (o *BadBinaries) {
	if v != nil && v.BadBinaries != nil {
		return v.BadBinaries
	}
	return
}

func (v *DomainConfiguration) GetHistoryArchivalStatus() (o ArchivalStatus) {
	if v != nil && v.HistoryArchivalStatus != nil {
		return *v.HistoryArchivalStatus
	}
	return
}

func (v *DomainConfiguration) GetHistoryArchivalURI() (o string) {
	if v != nil && v.HistoryArchivalURI != nil {
		return *v.HistoryArchivalURI
	}
	return
}

func (v *DomainConfiguration) GetVisibilityArchivalStatus() (o ArchivalStatus) {
	if v != nil && v.VisibilityArchivalStatus != nil {
		return *v.VisibilityArchivalStatus
	}
	return
}

func (v *DomainConfiguration) GetVisibilityArchivalURI() (o string) {
	if v != nil && v.VisibilityArchivalURI != nil {
		return *v.VisibilityArchivalURI
	}
	return
}

type DomainInfo struct {
	Name        *string
	Status      *DomainStatus
	Description *string
	OwnerEmail  *string
	Data        map[string]string
	UUID        *string
}

func (v *DomainInfo) GetName() (o string) {
	if v != nil && v.Name != nil {
		return *v.Name
	}
	return
}

func (v *DomainInfo) GetStatus() (o DomainStatus) {
	if v != nil && v.Status != nil {
		return *v.Status
	}
	return
}

func (v *DomainInfo) GetDescription() (o string) {
	if v != nil && v.Description != nil {
		return *v.Description
	}
	return
}

func (v *DomainInfo) GetOwnerEmail() (o string) {
	if v != nil && v.OwnerEmail != nil {
		return *v.OwnerEmail
	}
	return
}

func (v *DomainInfo) GetData() (o map[string]string) {
	if v != nil && v.Data != nil {
		return v.Data
	}
	return
}

func (v *DomainInfo) GetUUID() (o string) {
	if v != nil && v.UUID != nil {
		return *v.UUID
	}
	return
}

type DomainNotActiveError struct {
	Message        string
	DomainName     string
	CurrentCluster string
	ActiveCluster  string
}

func (v *DomainNotActiveError) GetMessage() (o string) {
	if v != nil {
		return v.Message
	}
	return
}

func (v *DomainNotActiveError) GetDomainName() (o string) {
	if v != nil {
		return v.DomainName
	}
	return
}

func (v *DomainNotActiveError) GetCurrentCluster() (o string) {
	if v != nil {
		return v.CurrentCluster
	}
	return
}

func (v *DomainNotActiveError) GetActiveCluster() (o string) {
	if v != nil {
		return v.ActiveCluster
	}
	return
}

type DomainReplicationConfiguration struct {
	ActiveClusterName *string
	Clusters          []*ClusterReplicationConfiguration
}

func (v *DomainReplicationConfiguration) GetActiveClusterName() (o string) {
	if v != nil && v.ActiveClusterName != nil {
		return *v.ActiveClusterName
	}
	return
}

func (v *DomainReplicationConfiguration) GetClusters() (o []*ClusterReplicationConfiguration) {
	if v != nil && v.Clusters != nil {
		return v.Clusters
	}
	return
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

func (v *EntityNotExistsError) GetMessage() (o string) {
	if v != nil {
		return v.Message
	}
	return
}

func (v *EntityNotExistsError) GetCurrentCluster() (o string) {
	if v != nil && v.CurrentCluster != nil {
		return *v.CurrentCluster
	}
	return
}

func (v *EntityNotExistsError) GetActiveCluster() (o string) {
	if v != nil && v.ActiveCluster != nil {
		return *v.ActiveCluster
	}
	return
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

func (v *ExternalWorkflowExecutionCancelRequestedEventAttributes) GetInitiatedEventId() (o int64) {
	if v != nil && v.InitiatedEventId != nil {
		return *v.InitiatedEventId
	}
	return
}

func (v *ExternalWorkflowExecutionCancelRequestedEventAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *ExternalWorkflowExecutionCancelRequestedEventAttributes) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

type ExternalWorkflowExecutionSignaledEventAttributes struct {
	InitiatedEventId  *int64
	Domain            *string
	WorkflowExecution *WorkflowExecution
	Control           []byte
}

func (v *ExternalWorkflowExecutionSignaledEventAttributes) GetInitiatedEventId() (o int64) {
	if v != nil && v.InitiatedEventId != nil {
		return *v.InitiatedEventId
	}
	return
}

func (v *ExternalWorkflowExecutionSignaledEventAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *ExternalWorkflowExecutionSignaledEventAttributes) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *ExternalWorkflowExecutionSignaledEventAttributes) GetControl() (o []byte) {
	if v != nil && v.Control != nil {
		return v.Control
	}
	return
}

type FailWorkflowExecutionDecisionAttributes struct {
	Reason  *string
	Details []byte
}

func (v *FailWorkflowExecutionDecisionAttributes) GetReason() (o string) {
	if v != nil && v.Reason != nil {
		return *v.Reason
	}
	return
}

func (v *FailWorkflowExecutionDecisionAttributes) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

type GetSearchAttributesResponse struct {
	Keys map[string]IndexedValueType
}

func (v *GetSearchAttributesResponse) GetKeys() (o map[string]IndexedValueType) {
	if v != nil && v.Keys != nil {
		return v.Keys
	}
	return
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

func (v *GetWorkflowExecutionHistoryRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *GetWorkflowExecutionHistoryRequest) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
}

func (v *GetWorkflowExecutionHistoryRequest) GetMaximumPageSize() (o int32) {
	if v != nil && v.MaximumPageSize != nil {
		return *v.MaximumPageSize
	}
	return
}

func (v *GetWorkflowExecutionHistoryRequest) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

func (v *GetWorkflowExecutionHistoryRequest) GetWaitForNewEvent() (o bool) {
	if v != nil && v.WaitForNewEvent != nil {
		return *v.WaitForNewEvent
	}
	return
}

func (v *GetWorkflowExecutionHistoryRequest) GetHistoryEventFilterType() (o HistoryEventFilterType) {
	if v != nil && v.HistoryEventFilterType != nil {
		return *v.HistoryEventFilterType
	}
	return
}

func (v *GetWorkflowExecutionHistoryRequest) GetSkipArchival() (o bool) {
	if v != nil && v.SkipArchival != nil {
		return *v.SkipArchival
	}
	return
}

type GetWorkflowExecutionHistoryResponse struct {
	History       *History
	RawHistory    []*DataBlob
	NextPageToken []byte
	Archived      *bool
}

func (v *GetWorkflowExecutionHistoryResponse) GetHistory() (o *History) {
	if v != nil && v.History != nil {
		return v.History
	}
	return
}

func (v *GetWorkflowExecutionHistoryResponse) GetRawHistory() (o []*DataBlob) {
	if v != nil && v.RawHistory != nil {
		return v.RawHistory
	}
	return
}

func (v *GetWorkflowExecutionHistoryResponse) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

func (v *GetWorkflowExecutionHistoryResponse) GetArchived() (o bool) {
	if v != nil && v.Archived != nil {
		return *v.Archived
	}
	return
}

type Header struct {
	Fields map[string][]byte
}

func (v *Header) GetFields() (o map[string][]byte) {
	if v != nil && v.Fields != nil {
		return v.Fields
	}
	return
}

type History struct {
	Events []*HistoryEvent
}

func (v *History) GetEvents() (o []*HistoryEvent) {
	if v != nil && v.Events != nil {
		return v.Events
	}
	return
}

type HistoryBranch struct {
	TreeID    *string
	BranchID  *string
	Ancestors []*HistoryBranchRange
}

func (v *HistoryBranch) GetTreeID() (o string) {
	if v != nil && v.TreeID != nil {
		return *v.TreeID
	}
	return
}

func (v *HistoryBranch) GetBranchID() (o string) {
	if v != nil && v.BranchID != nil {
		return *v.BranchID
	}
	return
}

func (v *HistoryBranch) GetAncestors() (o []*HistoryBranchRange) {
	if v != nil && v.Ancestors != nil {
		return v.Ancestors
	}
	return
}

type HistoryBranchRange struct {
	BranchID    *string
	BeginNodeID *int64
	EndNodeID   *int64
}

func (v *HistoryBranchRange) GetBranchID() (o string) {
	if v != nil && v.BranchID != nil {
		return *v.BranchID
	}
	return
}

func (v *HistoryBranchRange) GetBeginNodeID() (o int64) {
	if v != nil && v.BeginNodeID != nil {
		return *v.BeginNodeID
	}
	return
}

func (v *HistoryBranchRange) GetEndNodeID() (o int64) {
	if v != nil && v.EndNodeID != nil {
		return *v.EndNodeID
	}
	return
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

func (v *HistoryEvent) GetEventId() (o int64) {
	if v != nil && v.EventId != nil {
		return *v.EventId
	}
	return
}

func (v *HistoryEvent) GetTimestamp() (o int64) {
	if v != nil && v.Timestamp != nil {
		return *v.Timestamp
	}
	return
}

func (v *HistoryEvent) GetEventType() (o EventType) {
	if v != nil && v.EventType != nil {
		return *v.EventType
	}
	return
}

func (v *HistoryEvent) GetVersion() (o int64) {
	if v != nil && v.Version != nil {
		return *v.Version
	}
	return
}

func (v *HistoryEvent) GetTaskId() (o int64) {
	if v != nil && v.TaskId != nil {
		return *v.TaskId
	}
	return
}

func (v *HistoryEvent) GetWorkflowExecutionStartedEventAttributes() (o *WorkflowExecutionStartedEventAttributes) {
	if v != nil && v.WorkflowExecutionStartedEventAttributes != nil {
		return v.WorkflowExecutionStartedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetWorkflowExecutionCompletedEventAttributes() (o *WorkflowExecutionCompletedEventAttributes) {
	if v != nil && v.WorkflowExecutionCompletedEventAttributes != nil {
		return v.WorkflowExecutionCompletedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetWorkflowExecutionFailedEventAttributes() (o *WorkflowExecutionFailedEventAttributes) {
	if v != nil && v.WorkflowExecutionFailedEventAttributes != nil {
		return v.WorkflowExecutionFailedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetWorkflowExecutionTimedOutEventAttributes() (o *WorkflowExecutionTimedOutEventAttributes) {
	if v != nil && v.WorkflowExecutionTimedOutEventAttributes != nil {
		return v.WorkflowExecutionTimedOutEventAttributes
	}
	return
}

func (v *HistoryEvent) GetDecisionTaskScheduledEventAttributes() (o *DecisionTaskScheduledEventAttributes) {
	if v != nil && v.DecisionTaskScheduledEventAttributes != nil {
		return v.DecisionTaskScheduledEventAttributes
	}
	return
}

func (v *HistoryEvent) GetDecisionTaskStartedEventAttributes() (o *DecisionTaskStartedEventAttributes) {
	if v != nil && v.DecisionTaskStartedEventAttributes != nil {
		return v.DecisionTaskStartedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetDecisionTaskCompletedEventAttributes() (o *DecisionTaskCompletedEventAttributes) {
	if v != nil && v.DecisionTaskCompletedEventAttributes != nil {
		return v.DecisionTaskCompletedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetDecisionTaskTimedOutEventAttributes() (o *DecisionTaskTimedOutEventAttributes) {
	if v != nil && v.DecisionTaskTimedOutEventAttributes != nil {
		return v.DecisionTaskTimedOutEventAttributes
	}
	return
}

func (v *HistoryEvent) GetDecisionTaskFailedEventAttributes() (o *DecisionTaskFailedEventAttributes) {
	if v != nil && v.DecisionTaskFailedEventAttributes != nil {
		return v.DecisionTaskFailedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetActivityTaskScheduledEventAttributes() (o *ActivityTaskScheduledEventAttributes) {
	if v != nil && v.ActivityTaskScheduledEventAttributes != nil {
		return v.ActivityTaskScheduledEventAttributes
	}
	return
}

func (v *HistoryEvent) GetActivityTaskStartedEventAttributes() (o *ActivityTaskStartedEventAttributes) {
	if v != nil && v.ActivityTaskStartedEventAttributes != nil {
		return v.ActivityTaskStartedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetActivityTaskCompletedEventAttributes() (o *ActivityTaskCompletedEventAttributes) {
	if v != nil && v.ActivityTaskCompletedEventAttributes != nil {
		return v.ActivityTaskCompletedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetActivityTaskFailedEventAttributes() (o *ActivityTaskFailedEventAttributes) {
	if v != nil && v.ActivityTaskFailedEventAttributes != nil {
		return v.ActivityTaskFailedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetActivityTaskTimedOutEventAttributes() (o *ActivityTaskTimedOutEventAttributes) {
	if v != nil && v.ActivityTaskTimedOutEventAttributes != nil {
		return v.ActivityTaskTimedOutEventAttributes
	}
	return
}

func (v *HistoryEvent) GetTimerStartedEventAttributes() (o *TimerStartedEventAttributes) {
	if v != nil && v.TimerStartedEventAttributes != nil {
		return v.TimerStartedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetTimerFiredEventAttributes() (o *TimerFiredEventAttributes) {
	if v != nil && v.TimerFiredEventAttributes != nil {
		return v.TimerFiredEventAttributes
	}
	return
}

func (v *HistoryEvent) GetActivityTaskCancelRequestedEventAttributes() (o *ActivityTaskCancelRequestedEventAttributes) {
	if v != nil && v.ActivityTaskCancelRequestedEventAttributes != nil {
		return v.ActivityTaskCancelRequestedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetRequestCancelActivityTaskFailedEventAttributes() (o *RequestCancelActivityTaskFailedEventAttributes) {
	if v != nil && v.RequestCancelActivityTaskFailedEventAttributes != nil {
		return v.RequestCancelActivityTaskFailedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetActivityTaskCanceledEventAttributes() (o *ActivityTaskCanceledEventAttributes) {
	if v != nil && v.ActivityTaskCanceledEventAttributes != nil {
		return v.ActivityTaskCanceledEventAttributes
	}
	return
}

func (v *HistoryEvent) GetTimerCanceledEventAttributes() (o *TimerCanceledEventAttributes) {
	if v != nil && v.TimerCanceledEventAttributes != nil {
		return v.TimerCanceledEventAttributes
	}
	return
}

func (v *HistoryEvent) GetCancelTimerFailedEventAttributes() (o *CancelTimerFailedEventAttributes) {
	if v != nil && v.CancelTimerFailedEventAttributes != nil {
		return v.CancelTimerFailedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetMarkerRecordedEventAttributes() (o *MarkerRecordedEventAttributes) {
	if v != nil && v.MarkerRecordedEventAttributes != nil {
		return v.MarkerRecordedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetWorkflowExecutionSignaledEventAttributes() (o *WorkflowExecutionSignaledEventAttributes) {
	if v != nil && v.WorkflowExecutionSignaledEventAttributes != nil {
		return v.WorkflowExecutionSignaledEventAttributes
	}
	return
}

func (v *HistoryEvent) GetWorkflowExecutionTerminatedEventAttributes() (o *WorkflowExecutionTerminatedEventAttributes) {
	if v != nil && v.WorkflowExecutionTerminatedEventAttributes != nil {
		return v.WorkflowExecutionTerminatedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetWorkflowExecutionCancelRequestedEventAttributes() (o *WorkflowExecutionCancelRequestedEventAttributes) {
	if v != nil && v.WorkflowExecutionCancelRequestedEventAttributes != nil {
		return v.WorkflowExecutionCancelRequestedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetWorkflowExecutionCanceledEventAttributes() (o *WorkflowExecutionCanceledEventAttributes) {
	if v != nil && v.WorkflowExecutionCanceledEventAttributes != nil {
		return v.WorkflowExecutionCanceledEventAttributes
	}
	return
}

func (v *HistoryEvent) GetRequestCancelExternalWorkflowExecutionInitiatedEventAttributes() (o *RequestCancelExternalWorkflowExecutionInitiatedEventAttributes) {
	if v != nil && v.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes != nil {
		return v.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetRequestCancelExternalWorkflowExecutionFailedEventAttributes() (o *RequestCancelExternalWorkflowExecutionFailedEventAttributes) {
	if v != nil && v.RequestCancelExternalWorkflowExecutionFailedEventAttributes != nil {
		return v.RequestCancelExternalWorkflowExecutionFailedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetExternalWorkflowExecutionCancelRequestedEventAttributes() (o *ExternalWorkflowExecutionCancelRequestedEventAttributes) {
	if v != nil && v.ExternalWorkflowExecutionCancelRequestedEventAttributes != nil {
		return v.ExternalWorkflowExecutionCancelRequestedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetWorkflowExecutionContinuedAsNewEventAttributes() (o *WorkflowExecutionContinuedAsNewEventAttributes) {
	if v != nil && v.WorkflowExecutionContinuedAsNewEventAttributes != nil {
		return v.WorkflowExecutionContinuedAsNewEventAttributes
	}
	return
}

func (v *HistoryEvent) GetStartChildWorkflowExecutionInitiatedEventAttributes() (o *StartChildWorkflowExecutionInitiatedEventAttributes) {
	if v != nil && v.StartChildWorkflowExecutionInitiatedEventAttributes != nil {
		return v.StartChildWorkflowExecutionInitiatedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetStartChildWorkflowExecutionFailedEventAttributes() (o *StartChildWorkflowExecutionFailedEventAttributes) {
	if v != nil && v.StartChildWorkflowExecutionFailedEventAttributes != nil {
		return v.StartChildWorkflowExecutionFailedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetChildWorkflowExecutionStartedEventAttributes() (o *ChildWorkflowExecutionStartedEventAttributes) {
	if v != nil && v.ChildWorkflowExecutionStartedEventAttributes != nil {
		return v.ChildWorkflowExecutionStartedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetChildWorkflowExecutionCompletedEventAttributes() (o *ChildWorkflowExecutionCompletedEventAttributes) {
	if v != nil && v.ChildWorkflowExecutionCompletedEventAttributes != nil {
		return v.ChildWorkflowExecutionCompletedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetChildWorkflowExecutionFailedEventAttributes() (o *ChildWorkflowExecutionFailedEventAttributes) {
	if v != nil && v.ChildWorkflowExecutionFailedEventAttributes != nil {
		return v.ChildWorkflowExecutionFailedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetChildWorkflowExecutionCanceledEventAttributes() (o *ChildWorkflowExecutionCanceledEventAttributes) {
	if v != nil && v.ChildWorkflowExecutionCanceledEventAttributes != nil {
		return v.ChildWorkflowExecutionCanceledEventAttributes
	}
	return
}

func (v *HistoryEvent) GetChildWorkflowExecutionTimedOutEventAttributes() (o *ChildWorkflowExecutionTimedOutEventAttributes) {
	if v != nil && v.ChildWorkflowExecutionTimedOutEventAttributes != nil {
		return v.ChildWorkflowExecutionTimedOutEventAttributes
	}
	return
}

func (v *HistoryEvent) GetChildWorkflowExecutionTerminatedEventAttributes() (o *ChildWorkflowExecutionTerminatedEventAttributes) {
	if v != nil && v.ChildWorkflowExecutionTerminatedEventAttributes != nil {
		return v.ChildWorkflowExecutionTerminatedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetSignalExternalWorkflowExecutionInitiatedEventAttributes() (o *SignalExternalWorkflowExecutionInitiatedEventAttributes) {
	if v != nil && v.SignalExternalWorkflowExecutionInitiatedEventAttributes != nil {
		return v.SignalExternalWorkflowExecutionInitiatedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetSignalExternalWorkflowExecutionFailedEventAttributes() (o *SignalExternalWorkflowExecutionFailedEventAttributes) {
	if v != nil && v.SignalExternalWorkflowExecutionFailedEventAttributes != nil {
		return v.SignalExternalWorkflowExecutionFailedEventAttributes
	}
	return
}

func (v *HistoryEvent) GetExternalWorkflowExecutionSignaledEventAttributes() (o *ExternalWorkflowExecutionSignaledEventAttributes) {
	if v != nil && v.ExternalWorkflowExecutionSignaledEventAttributes != nil {
		return v.ExternalWorkflowExecutionSignaledEventAttributes
	}
	return
}

func (v *HistoryEvent) GetUpsertWorkflowSearchAttributesEventAttributes() (o *UpsertWorkflowSearchAttributesEventAttributes) {
	if v != nil && v.UpsertWorkflowSearchAttributesEventAttributes != nil {
		return v.UpsertWorkflowSearchAttributesEventAttributes
	}
	return
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

func (v *InternalDataInconsistencyError) GetMessage() (o string) {
	if v != nil {
		return v.Message
	}
	return
}

type InternalServiceError struct {
	Message string
}

func (v *InternalServiceError) GetMessage() (o string) {
	if v != nil {
		return v.Message
	}
	return
}

type LimitExceededError struct {
	Message string
}

func (v *LimitExceededError) GetMessage() (o string) {
	if v != nil {
		return v.Message
	}
	return
}

type ListArchivedWorkflowExecutionsRequest struct {
	Domain        *string
	PageSize      *int32
	NextPageToken []byte
	Query         *string
}

func (v *ListArchivedWorkflowExecutionsRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *ListArchivedWorkflowExecutionsRequest) GetPageSize() (o int32) {
	if v != nil && v.PageSize != nil {
		return *v.PageSize
	}
	return
}

func (v *ListArchivedWorkflowExecutionsRequest) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

func (v *ListArchivedWorkflowExecutionsRequest) GetQuery() (o string) {
	if v != nil && v.Query != nil {
		return *v.Query
	}
	return
}

type ListArchivedWorkflowExecutionsResponse struct {
	Executions    []*WorkflowExecutionInfo
	NextPageToken []byte
}

func (v *ListArchivedWorkflowExecutionsResponse) GetExecutions() (o []*WorkflowExecutionInfo) {
	if v != nil && v.Executions != nil {
		return v.Executions
	}
	return
}

func (v *ListArchivedWorkflowExecutionsResponse) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
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

func (v *ListClosedWorkflowExecutionsRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *ListClosedWorkflowExecutionsRequest) GetMaximumPageSize() (o int32) {
	if v != nil && v.MaximumPageSize != nil {
		return *v.MaximumPageSize
	}
	return
}

func (v *ListClosedWorkflowExecutionsRequest) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

func (v *ListClosedWorkflowExecutionsRequest) GetStartTimeFilter() (o *StartTimeFilter) {
	if v != nil && v.StartTimeFilter != nil {
		return v.StartTimeFilter
	}
	return
}

func (v *ListClosedWorkflowExecutionsRequest) GetExecutionFilter() (o *WorkflowExecutionFilter) {
	if v != nil && v.ExecutionFilter != nil {
		return v.ExecutionFilter
	}
	return
}

func (v *ListClosedWorkflowExecutionsRequest) GetTypeFilter() (o *WorkflowTypeFilter) {
	if v != nil && v.TypeFilter != nil {
		return v.TypeFilter
	}
	return
}

func (v *ListClosedWorkflowExecutionsRequest) GetStatusFilter() (o WorkflowExecutionCloseStatus) {
	if v != nil && v.StatusFilter != nil {
		return *v.StatusFilter
	}
	return
}

type ListClosedWorkflowExecutionsResponse struct {
	Executions    []*WorkflowExecutionInfo
	NextPageToken []byte
}

func (v *ListClosedWorkflowExecutionsResponse) GetExecutions() (o []*WorkflowExecutionInfo) {
	if v != nil && v.Executions != nil {
		return v.Executions
	}
	return
}

func (v *ListClosedWorkflowExecutionsResponse) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

type ListDomainsRequest struct {
	PageSize      *int32
	NextPageToken []byte
}

func (v *ListDomainsRequest) GetPageSize() (o int32) {
	if v != nil && v.PageSize != nil {
		return *v.PageSize
	}
	return
}

func (v *ListDomainsRequest) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

type ListDomainsResponse struct {
	Domains       []*DescribeDomainResponse
	NextPageToken []byte
}

func (v *ListDomainsResponse) GetDomains() (o []*DescribeDomainResponse) {
	if v != nil && v.Domains != nil {
		return v.Domains
	}
	return
}

func (v *ListDomainsResponse) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

type ListOpenWorkflowExecutionsRequest struct {
	Domain          *string
	MaximumPageSize *int32
	NextPageToken   []byte
	StartTimeFilter *StartTimeFilter
	ExecutionFilter *WorkflowExecutionFilter
	TypeFilter      *WorkflowTypeFilter
}

func (v *ListOpenWorkflowExecutionsRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *ListOpenWorkflowExecutionsRequest) GetMaximumPageSize() (o int32) {
	if v != nil && v.MaximumPageSize != nil {
		return *v.MaximumPageSize
	}
	return
}

func (v *ListOpenWorkflowExecutionsRequest) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

func (v *ListOpenWorkflowExecutionsRequest) GetStartTimeFilter() (o *StartTimeFilter) {
	if v != nil && v.StartTimeFilter != nil {
		return v.StartTimeFilter
	}
	return
}

func (v *ListOpenWorkflowExecutionsRequest) GetExecutionFilter() (o *WorkflowExecutionFilter) {
	if v != nil && v.ExecutionFilter != nil {
		return v.ExecutionFilter
	}
	return
}

func (v *ListOpenWorkflowExecutionsRequest) GetTypeFilter() (o *WorkflowTypeFilter) {
	if v != nil && v.TypeFilter != nil {
		return v.TypeFilter
	}
	return
}

type ListOpenWorkflowExecutionsResponse struct {
	Executions    []*WorkflowExecutionInfo
	NextPageToken []byte
}

func (v *ListOpenWorkflowExecutionsResponse) GetExecutions() (o []*WorkflowExecutionInfo) {
	if v != nil && v.Executions != nil {
		return v.Executions
	}
	return
}

func (v *ListOpenWorkflowExecutionsResponse) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

type ListTaskListPartitionsRequest struct {
	Domain   *string
	TaskList *TaskList
}

func (v *ListTaskListPartitionsRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *ListTaskListPartitionsRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

type ListTaskListPartitionsResponse struct {
	ActivityTaskListPartitions []*TaskListPartitionMetadata
	DecisionTaskListPartitions []*TaskListPartitionMetadata
}

func (v *ListTaskListPartitionsResponse) GetActivityTaskListPartitions() (o []*TaskListPartitionMetadata) {
	if v != nil && v.ActivityTaskListPartitions != nil {
		return v.ActivityTaskListPartitions
	}
	return
}

func (v *ListTaskListPartitionsResponse) GetDecisionTaskListPartitions() (o []*TaskListPartitionMetadata) {
	if v != nil && v.DecisionTaskListPartitions != nil {
		return v.DecisionTaskListPartitions
	}
	return
}

type ListWorkflowExecutionsRequest struct {
	Domain        *string
	PageSize      *int32
	NextPageToken []byte
	Query         *string
}

func (v *ListWorkflowExecutionsRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *ListWorkflowExecutionsRequest) GetPageSize() (o int32) {
	if v != nil && v.PageSize != nil {
		return *v.PageSize
	}
	return
}

func (v *ListWorkflowExecutionsRequest) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

func (v *ListWorkflowExecutionsRequest) GetQuery() (o string) {
	if v != nil && v.Query != nil {
		return *v.Query
	}
	return
}

type ListWorkflowExecutionsResponse struct {
	Executions    []*WorkflowExecutionInfo
	NextPageToken []byte
}

func (v *ListWorkflowExecutionsResponse) GetExecutions() (o []*WorkflowExecutionInfo) {
	if v != nil && v.Executions != nil {
		return v.Executions
	}
	return
}

func (v *ListWorkflowExecutionsResponse) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

type MarkerRecordedEventAttributes struct {
	MarkerName                   *string
	Details                      []byte
	DecisionTaskCompletedEventId *int64
	Header                       *Header
}

func (v *MarkerRecordedEventAttributes) GetMarkerName() (o string) {
	if v != nil && v.MarkerName != nil {
		return *v.MarkerName
	}
	return
}

func (v *MarkerRecordedEventAttributes) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *MarkerRecordedEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
}

func (v *MarkerRecordedEventAttributes) GetHeader() (o *Header) {
	if v != nil && v.Header != nil {
		return v.Header
	}
	return
}

type Memo struct {
	Fields map[string][]byte
}

func (v *Memo) GetFields() (o map[string][]byte) {
	if v != nil && v.Fields != nil {
		return v.Fields
	}
	return
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

func (v *PendingActivityInfo) GetActivityID() (o string) {
	if v != nil && v.ActivityID != nil {
		return *v.ActivityID
	}
	return
}

func (v *PendingActivityInfo) GetActivityType() (o *ActivityType) {
	if v != nil && v.ActivityType != nil {
		return v.ActivityType
	}
	return
}

func (v *PendingActivityInfo) GetState() (o PendingActivityState) {
	if v != nil && v.State != nil {
		return *v.State
	}
	return
}

func (v *PendingActivityInfo) GetHeartbeatDetails() (o []byte) {
	if v != nil && v.HeartbeatDetails != nil {
		return v.HeartbeatDetails
	}
	return
}

func (v *PendingActivityInfo) GetLastHeartbeatTimestamp() (o int64) {
	if v != nil && v.LastHeartbeatTimestamp != nil {
		return *v.LastHeartbeatTimestamp
	}
	return
}

func (v *PendingActivityInfo) GetLastStartedTimestamp() (o int64) {
	if v != nil && v.LastStartedTimestamp != nil {
		return *v.LastStartedTimestamp
	}
	return
}

func (v *PendingActivityInfo) GetAttempt() (o int32) {
	if v != nil && v.Attempt != nil {
		return *v.Attempt
	}
	return
}

func (v *PendingActivityInfo) GetMaximumAttempts() (o int32) {
	if v != nil && v.MaximumAttempts != nil {
		return *v.MaximumAttempts
	}
	return
}

func (v *PendingActivityInfo) GetScheduledTimestamp() (o int64) {
	if v != nil && v.ScheduledTimestamp != nil {
		return *v.ScheduledTimestamp
	}
	return
}

func (v *PendingActivityInfo) GetExpirationTimestamp() (o int64) {
	if v != nil && v.ExpirationTimestamp != nil {
		return *v.ExpirationTimestamp
	}
	return
}

func (v *PendingActivityInfo) GetLastFailureReason() (o string) {
	if v != nil && v.LastFailureReason != nil {
		return *v.LastFailureReason
	}
	return
}

func (v *PendingActivityInfo) GetLastWorkerIdentity() (o string) {
	if v != nil && v.LastWorkerIdentity != nil {
		return *v.LastWorkerIdentity
	}
	return
}

func (v *PendingActivityInfo) GetLastFailureDetails() (o []byte) {
	if v != nil && v.LastFailureDetails != nil {
		return v.LastFailureDetails
	}
	return
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

func (v *PendingChildExecutionInfo) GetWorkflowID() (o string) {
	if v != nil && v.WorkflowID != nil {
		return *v.WorkflowID
	}
	return
}

func (v *PendingChildExecutionInfo) GetRunID() (o string) {
	if v != nil && v.RunID != nil {
		return *v.RunID
	}
	return
}

func (v *PendingChildExecutionInfo) GetWorkflowTypName() (o string) {
	if v != nil && v.WorkflowTypName != nil {
		return *v.WorkflowTypName
	}
	return
}

func (v *PendingChildExecutionInfo) GetInitiatedID() (o int64) {
	if v != nil && v.InitiatedID != nil {
		return *v.InitiatedID
	}
	return
}

func (v *PendingChildExecutionInfo) GetParentClosePolicy() (o ParentClosePolicy) {
	if v != nil && v.ParentClosePolicy != nil {
		return *v.ParentClosePolicy
	}
	return
}

type PollForActivityTaskRequest struct {
	Domain           *string
	TaskList         *TaskList
	Identity         *string
	TaskListMetadata *TaskListMetadata
}

func (v *PollForActivityTaskRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *PollForActivityTaskRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

func (v *PollForActivityTaskRequest) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

func (v *PollForActivityTaskRequest) GetTaskListMetadata() (o *TaskListMetadata) {
	if v != nil && v.TaskListMetadata != nil {
		return v.TaskListMetadata
	}
	return
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

func (v *PollForActivityTaskResponse) GetTaskToken() (o []byte) {
	if v != nil && v.TaskToken != nil {
		return v.TaskToken
	}
	return
}

func (v *PollForActivityTaskResponse) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *PollForActivityTaskResponse) GetActivityId() (o string) {
	if v != nil && v.ActivityId != nil {
		return *v.ActivityId
	}
	return
}

func (v *PollForActivityTaskResponse) GetActivityType() (o *ActivityType) {
	if v != nil && v.ActivityType != nil {
		return v.ActivityType
	}
	return
}

func (v *PollForActivityTaskResponse) GetInput() (o []byte) {
	if v != nil && v.Input != nil {
		return v.Input
	}
	return
}

func (v *PollForActivityTaskResponse) GetScheduledTimestamp() (o int64) {
	if v != nil && v.ScheduledTimestamp != nil {
		return *v.ScheduledTimestamp
	}
	return
}

func (v *PollForActivityTaskResponse) GetScheduleToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.ScheduleToCloseTimeoutSeconds != nil {
		return *v.ScheduleToCloseTimeoutSeconds
	}
	return
}

func (v *PollForActivityTaskResponse) GetStartedTimestamp() (o int64) {
	if v != nil && v.StartedTimestamp != nil {
		return *v.StartedTimestamp
	}
	return
}

func (v *PollForActivityTaskResponse) GetStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.StartToCloseTimeoutSeconds != nil {
		return *v.StartToCloseTimeoutSeconds
	}
	return
}

func (v *PollForActivityTaskResponse) GetHeartbeatTimeoutSeconds() (o int32) {
	if v != nil && v.HeartbeatTimeoutSeconds != nil {
		return *v.HeartbeatTimeoutSeconds
	}
	return
}

func (v *PollForActivityTaskResponse) GetAttempt() (o int32) {
	if v != nil && v.Attempt != nil {
		return *v.Attempt
	}
	return
}

func (v *PollForActivityTaskResponse) GetScheduledTimestampOfThisAttempt() (o int64) {
	if v != nil && v.ScheduledTimestampOfThisAttempt != nil {
		return *v.ScheduledTimestampOfThisAttempt
	}
	return
}

func (v *PollForActivityTaskResponse) GetHeartbeatDetails() (o []byte) {
	if v != nil && v.HeartbeatDetails != nil {
		return v.HeartbeatDetails
	}
	return
}

func (v *PollForActivityTaskResponse) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

func (v *PollForActivityTaskResponse) GetWorkflowDomain() (o string) {
	if v != nil && v.WorkflowDomain != nil {
		return *v.WorkflowDomain
	}
	return
}

func (v *PollForActivityTaskResponse) GetHeader() (o *Header) {
	if v != nil && v.Header != nil {
		return v.Header
	}
	return
}

type PollForDecisionTaskRequest struct {
	Domain         *string
	TaskList       *TaskList
	Identity       *string
	BinaryChecksum *string
}

func (v *PollForDecisionTaskRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *PollForDecisionTaskRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

func (v *PollForDecisionTaskRequest) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

func (v *PollForDecisionTaskRequest) GetBinaryChecksum() (o string) {
	if v != nil && v.BinaryChecksum != nil {
		return *v.BinaryChecksum
	}
	return
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

func (v *PollForDecisionTaskResponse) GetTaskToken() (o []byte) {
	if v != nil && v.TaskToken != nil {
		return v.TaskToken
	}
	return
}

func (v *PollForDecisionTaskResponse) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *PollForDecisionTaskResponse) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

func (v *PollForDecisionTaskResponse) GetPreviousStartedEventId() (o int64) {
	if v != nil && v.PreviousStartedEventId != nil {
		return *v.PreviousStartedEventId
	}
	return
}

func (v *PollForDecisionTaskResponse) GetStartedEventId() (o int64) {
	if v != nil && v.StartedEventId != nil {
		return *v.StartedEventId
	}
	return
}

func (v *PollForDecisionTaskResponse) GetAttempt() (o int64) {
	if v != nil && v.Attempt != nil {
		return *v.Attempt
	}
	return
}

func (v *PollForDecisionTaskResponse) GetBacklogCountHint() (o int64) {
	if v != nil && v.BacklogCountHint != nil {
		return *v.BacklogCountHint
	}
	return
}

func (v *PollForDecisionTaskResponse) GetHistory() (o *History) {
	if v != nil && v.History != nil {
		return v.History
	}
	return
}

func (v *PollForDecisionTaskResponse) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

func (v *PollForDecisionTaskResponse) GetQuery() (o *WorkflowQuery) {
	if v != nil && v.Query != nil {
		return v.Query
	}
	return
}

func (v *PollForDecisionTaskResponse) GetWorkflowExecutionTaskList() (o *TaskList) {
	if v != nil && v.WorkflowExecutionTaskList != nil {
		return v.WorkflowExecutionTaskList
	}
	return
}

func (v *PollForDecisionTaskResponse) GetScheduledTimestamp() (o int64) {
	if v != nil && v.ScheduledTimestamp != nil {
		return *v.ScheduledTimestamp
	}
	return
}

func (v *PollForDecisionTaskResponse) GetStartedTimestamp() (o int64) {
	if v != nil && v.StartedTimestamp != nil {
		return *v.StartedTimestamp
	}
	return
}

func (v *PollForDecisionTaskResponse) GetQueries() (o map[string]*WorkflowQuery) {
	if v != nil && v.Queries != nil {
		return v.Queries
	}
	return
}

type PollerInfo struct {
	LastAccessTime *int64
	Identity       *string
	RatePerSecond  *float64
}

func (v *PollerInfo) GetLastAccessTime() (o int64) {
	if v != nil && v.LastAccessTime != nil {
		return *v.LastAccessTime
	}
	return
}

func (v *PollerInfo) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

func (v *PollerInfo) GetRatePerSecond() (o float64) {
	if v != nil && v.RatePerSecond != nil {
		return *v.RatePerSecond
	}
	return
}

type QueryConsistencyLevel int32

const (
	QueryConsistencyLevelEventual QueryConsistencyLevel = iota
	QueryConsistencyLevelStrong
)

type QueryFailedError struct {
	Message string
}

func (v *QueryFailedError) GetMessage() (o string) {
	if v != nil {
		return v.Message
	}
	return
}

type QueryRejectCondition int32

const (
	QueryRejectConditionNotCompletedCleanly QueryRejectCondition = iota
	QueryRejectConditionNotOpen
)

type QueryRejected struct {
	CloseStatus *WorkflowExecutionCloseStatus
}

func (v *QueryRejected) GetCloseStatus() (o WorkflowExecutionCloseStatus) {
	if v != nil && v.CloseStatus != nil {
		return *v.CloseStatus
	}
	return
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

func (v *QueryWorkflowRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *QueryWorkflowRequest) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
}

func (v *QueryWorkflowRequest) GetQuery() (o *WorkflowQuery) {
	if v != nil && v.Query != nil {
		return v.Query
	}
	return
}

func (v *QueryWorkflowRequest) GetQueryRejectCondition() (o QueryRejectCondition) {
	if v != nil && v.QueryRejectCondition != nil {
		return *v.QueryRejectCondition
	}
	return
}

func (v *QueryWorkflowRequest) GetQueryConsistencyLevel() (o QueryConsistencyLevel) {
	if v != nil && v.QueryConsistencyLevel != nil {
		return *v.QueryConsistencyLevel
	}
	return
}

type QueryWorkflowResponse struct {
	QueryResult   []byte
	QueryRejected *QueryRejected
}

func (v *QueryWorkflowResponse) GetQueryResult() (o []byte) {
	if v != nil && v.QueryResult != nil {
		return v.QueryResult
	}
	return
}

func (v *QueryWorkflowResponse) GetQueryRejected() (o *QueryRejected) {
	if v != nil && v.QueryRejected != nil {
		return v.QueryRejected
	}
	return
}

type ReapplyEventsRequest struct {
	DomainName        *string
	WorkflowExecution *WorkflowExecution
	Events            *DataBlob
}

func (v *ReapplyEventsRequest) GetDomainName() (o string) {
	if v != nil && v.DomainName != nil {
		return *v.DomainName
	}
	return
}

func (v *ReapplyEventsRequest) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *ReapplyEventsRequest) GetEvents() (o *DataBlob) {
	if v != nil && v.Events != nil {
		return v.Events
	}
	return
}

type RecordActivityTaskHeartbeatByIDRequest struct {
	Domain     *string
	WorkflowID *string
	RunID      *string
	ActivityID *string
	Details    []byte
	Identity   *string
}

func (v *RecordActivityTaskHeartbeatByIDRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *RecordActivityTaskHeartbeatByIDRequest) GetWorkflowID() (o string) {
	if v != nil && v.WorkflowID != nil {
		return *v.WorkflowID
	}
	return
}

func (v *RecordActivityTaskHeartbeatByIDRequest) GetRunID() (o string) {
	if v != nil && v.RunID != nil {
		return *v.RunID
	}
	return
}

func (v *RecordActivityTaskHeartbeatByIDRequest) GetActivityID() (o string) {
	if v != nil && v.ActivityID != nil {
		return *v.ActivityID
	}
	return
}

func (v *RecordActivityTaskHeartbeatByIDRequest) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *RecordActivityTaskHeartbeatByIDRequest) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

type RecordActivityTaskHeartbeatRequest struct {
	TaskToken []byte
	Details   []byte
	Identity  *string
}

func (v *RecordActivityTaskHeartbeatRequest) GetTaskToken() (o []byte) {
	if v != nil && v.TaskToken != nil {
		return v.TaskToken
	}
	return
}

func (v *RecordActivityTaskHeartbeatRequest) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *RecordActivityTaskHeartbeatRequest) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

type RecordActivityTaskHeartbeatResponse struct {
	CancelRequested *bool
}

func (v *RecordActivityTaskHeartbeatResponse) GetCancelRequested() (o bool) {
	if v != nil && v.CancelRequested != nil {
		return *v.CancelRequested
	}
	return
}

type RecordMarkerDecisionAttributes struct {
	MarkerName *string
	Details    []byte
	Header     *Header
}

func (v *RecordMarkerDecisionAttributes) GetMarkerName() (o string) {
	if v != nil && v.MarkerName != nil {
		return *v.MarkerName
	}
	return
}

func (v *RecordMarkerDecisionAttributes) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *RecordMarkerDecisionAttributes) GetHeader() (o *Header) {
	if v != nil && v.Header != nil {
		return v.Header
	}
	return
}

type RefreshWorkflowTasksRequest struct {
	Domain    *string
	Execution *WorkflowExecution
}

func (v *RefreshWorkflowTasksRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *RefreshWorkflowTasksRequest) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
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

func (v *RegisterDomainRequest) GetName() (o string) {
	if v != nil && v.Name != nil {
		return *v.Name
	}
	return
}

func (v *RegisterDomainRequest) GetDescription() (o string) {
	if v != nil && v.Description != nil {
		return *v.Description
	}
	return
}

func (v *RegisterDomainRequest) GetOwnerEmail() (o string) {
	if v != nil && v.OwnerEmail != nil {
		return *v.OwnerEmail
	}
	return
}

func (v *RegisterDomainRequest) GetWorkflowExecutionRetentionPeriodInDays() (o int32) {
	if v != nil && v.WorkflowExecutionRetentionPeriodInDays != nil {
		return *v.WorkflowExecutionRetentionPeriodInDays
	}
	return
}

func (v *RegisterDomainRequest) GetEmitMetric() (o bool) {
	if v != nil && v.EmitMetric != nil {
		return *v.EmitMetric
	}
	return
}

func (v *RegisterDomainRequest) GetClusters() (o []*ClusterReplicationConfiguration) {
	if v != nil && v.Clusters != nil {
		return v.Clusters
	}
	return
}

func (v *RegisterDomainRequest) GetActiveClusterName() (o string) {
	if v != nil && v.ActiveClusterName != nil {
		return *v.ActiveClusterName
	}
	return
}

func (v *RegisterDomainRequest) GetData() (o map[string]string) {
	if v != nil && v.Data != nil {
		return v.Data
	}
	return
}

func (v *RegisterDomainRequest) GetSecurityToken() (o string) {
	if v != nil && v.SecurityToken != nil {
		return *v.SecurityToken
	}
	return
}

func (v *RegisterDomainRequest) GetIsGlobalDomain() (o bool) {
	if v != nil && v.IsGlobalDomain != nil {
		return *v.IsGlobalDomain
	}
	return
}

func (v *RegisterDomainRequest) GetHistoryArchivalStatus() (o ArchivalStatus) {
	if v != nil && v.HistoryArchivalStatus != nil {
		return *v.HistoryArchivalStatus
	}
	return
}

func (v *RegisterDomainRequest) GetHistoryArchivalURI() (o string) {
	if v != nil && v.HistoryArchivalURI != nil {
		return *v.HistoryArchivalURI
	}
	return
}

func (v *RegisterDomainRequest) GetVisibilityArchivalStatus() (o ArchivalStatus) {
	if v != nil && v.VisibilityArchivalStatus != nil {
		return *v.VisibilityArchivalStatus
	}
	return
}

func (v *RegisterDomainRequest) GetVisibilityArchivalURI() (o string) {
	if v != nil && v.VisibilityArchivalURI != nil {
		return *v.VisibilityArchivalURI
	}
	return
}

type RemoteSyncMatchedError struct {
	Message string
}

func (v *RemoteSyncMatchedError) GetMessage() (o string) {
	if v != nil {
		return v.Message
	}
	return
}

type RemoveTaskRequest struct {
	ShardID             *int32
	Type                *int32
	TaskID              *int64
	VisibilityTimestamp *int64
}

func (v *RemoveTaskRequest) GetShardID() (o int32) {
	if v != nil && v.ShardID != nil {
		return *v.ShardID
	}
	return
}

func (v *RemoveTaskRequest) GetType() (o int32) {
	if v != nil && v.Type != nil {
		return *v.Type
	}
	return
}

func (v *RemoveTaskRequest) GetTaskID() (o int64) {
	if v != nil && v.TaskID != nil {
		return *v.TaskID
	}
	return
}

func (v *RemoveTaskRequest) GetVisibilityTimestamp() (o int64) {
	if v != nil && v.VisibilityTimestamp != nil {
		return *v.VisibilityTimestamp
	}
	return
}

type RequestCancelActivityTaskDecisionAttributes struct {
	ActivityId *string
}

func (v *RequestCancelActivityTaskDecisionAttributes) GetActivityId() (o string) {
	if v != nil && v.ActivityId != nil {
		return *v.ActivityId
	}
	return
}

type RequestCancelActivityTaskFailedEventAttributes struct {
	ActivityId                   *string
	Cause                        *string
	DecisionTaskCompletedEventId *int64
}

func (v *RequestCancelActivityTaskFailedEventAttributes) GetActivityId() (o string) {
	if v != nil && v.ActivityId != nil {
		return *v.ActivityId
	}
	return
}

func (v *RequestCancelActivityTaskFailedEventAttributes) GetCause() (o string) {
	if v != nil && v.Cause != nil {
		return *v.Cause
	}
	return
}

func (v *RequestCancelActivityTaskFailedEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
}

type RequestCancelExternalWorkflowExecutionDecisionAttributes struct {
	Domain            *string
	WorkflowId        *string
	RunId             *string
	Control           []byte
	ChildWorkflowOnly *bool
}

func (v *RequestCancelExternalWorkflowExecutionDecisionAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *RequestCancelExternalWorkflowExecutionDecisionAttributes) GetWorkflowId() (o string) {
	if v != nil && v.WorkflowId != nil {
		return *v.WorkflowId
	}
	return
}

func (v *RequestCancelExternalWorkflowExecutionDecisionAttributes) GetRunId() (o string) {
	if v != nil && v.RunId != nil {
		return *v.RunId
	}
	return
}

func (v *RequestCancelExternalWorkflowExecutionDecisionAttributes) GetControl() (o []byte) {
	if v != nil && v.Control != nil {
		return v.Control
	}
	return
}

func (v *RequestCancelExternalWorkflowExecutionDecisionAttributes) GetChildWorkflowOnly() (o bool) {
	if v != nil && v.ChildWorkflowOnly != nil {
		return *v.ChildWorkflowOnly
	}
	return
}

type RequestCancelExternalWorkflowExecutionFailedEventAttributes struct {
	Cause                        *CancelExternalWorkflowExecutionFailedCause
	DecisionTaskCompletedEventId *int64
	Domain                       *string
	WorkflowExecution            *WorkflowExecution
	InitiatedEventId             *int64
	Control                      []byte
}

func (v *RequestCancelExternalWorkflowExecutionFailedEventAttributes) GetCause() (o CancelExternalWorkflowExecutionFailedCause) {
	if v != nil && v.Cause != nil {
		return *v.Cause
	}
	return
}

func (v *RequestCancelExternalWorkflowExecutionFailedEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
}

func (v *RequestCancelExternalWorkflowExecutionFailedEventAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *RequestCancelExternalWorkflowExecutionFailedEventAttributes) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *RequestCancelExternalWorkflowExecutionFailedEventAttributes) GetInitiatedEventId() (o int64) {
	if v != nil && v.InitiatedEventId != nil {
		return *v.InitiatedEventId
	}
	return
}

func (v *RequestCancelExternalWorkflowExecutionFailedEventAttributes) GetControl() (o []byte) {
	if v != nil && v.Control != nil {
		return v.Control
	}
	return
}

type RequestCancelExternalWorkflowExecutionInitiatedEventAttributes struct {
	DecisionTaskCompletedEventId *int64
	Domain                       *string
	WorkflowExecution            *WorkflowExecution
	Control                      []byte
	ChildWorkflowOnly            *bool
}

func (v *RequestCancelExternalWorkflowExecutionInitiatedEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
}

func (v *RequestCancelExternalWorkflowExecutionInitiatedEventAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *RequestCancelExternalWorkflowExecutionInitiatedEventAttributes) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *RequestCancelExternalWorkflowExecutionInitiatedEventAttributes) GetControl() (o []byte) {
	if v != nil && v.Control != nil {
		return v.Control
	}
	return
}

func (v *RequestCancelExternalWorkflowExecutionInitiatedEventAttributes) GetChildWorkflowOnly() (o bool) {
	if v != nil && v.ChildWorkflowOnly != nil {
		return *v.ChildWorkflowOnly
	}
	return
}

type RequestCancelWorkflowExecutionRequest struct {
	Domain            *string
	WorkflowExecution *WorkflowExecution
	Identity          *string
	RequestId         *string
}

func (v *RequestCancelWorkflowExecutionRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *RequestCancelWorkflowExecutionRequest) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *RequestCancelWorkflowExecutionRequest) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

func (v *RequestCancelWorkflowExecutionRequest) GetRequestId() (o string) {
	if v != nil && v.RequestId != nil {
		return *v.RequestId
	}
	return
}

type ResetPointInfo struct {
	BinaryChecksum           *string
	RunId                    *string
	FirstDecisionCompletedId *int64
	CreatedTimeNano          *int64
	ExpiringTimeNano         *int64
	Resettable               *bool
}

func (v *ResetPointInfo) GetBinaryChecksum() (o string) {
	if v != nil && v.BinaryChecksum != nil {
		return *v.BinaryChecksum
	}
	return
}

func (v *ResetPointInfo) GetRunId() (o string) {
	if v != nil && v.RunId != nil {
		return *v.RunId
	}
	return
}

func (v *ResetPointInfo) GetFirstDecisionCompletedId() (o int64) {
	if v != nil && v.FirstDecisionCompletedId != nil {
		return *v.FirstDecisionCompletedId
	}
	return
}

func (v *ResetPointInfo) GetCreatedTimeNano() (o int64) {
	if v != nil && v.CreatedTimeNano != nil {
		return *v.CreatedTimeNano
	}
	return
}

func (v *ResetPointInfo) GetExpiringTimeNano() (o int64) {
	if v != nil && v.ExpiringTimeNano != nil {
		return *v.ExpiringTimeNano
	}
	return
}

func (v *ResetPointInfo) GetResettable() (o bool) {
	if v != nil && v.Resettable != nil {
		return *v.Resettable
	}
	return
}

type ResetPoints struct {
	Points []*ResetPointInfo
}

func (v *ResetPoints) GetPoints() (o []*ResetPointInfo) {
	if v != nil && v.Points != nil {
		return v.Points
	}
	return
}

type ResetQueueRequest struct {
	ShardID     *int32
	ClusterName *string
	Type        *int32
}

func (v *ResetQueueRequest) GetShardID() (o int32) {
	if v != nil && v.ShardID != nil {
		return *v.ShardID
	}
	return
}

func (v *ResetQueueRequest) GetClusterName() (o string) {
	if v != nil && v.ClusterName != nil {
		return *v.ClusterName
	}
	return
}

func (v *ResetQueueRequest) GetType() (o int32) {
	if v != nil && v.Type != nil {
		return *v.Type
	}
	return
}

type ResetStickyTaskListRequest struct {
	Domain    *string
	Execution *WorkflowExecution
}

func (v *ResetStickyTaskListRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *ResetStickyTaskListRequest) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
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

func (v *ResetWorkflowExecutionRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *ResetWorkflowExecutionRequest) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *ResetWorkflowExecutionRequest) GetReason() (o string) {
	if v != nil && v.Reason != nil {
		return *v.Reason
	}
	return
}

func (v *ResetWorkflowExecutionRequest) GetDecisionFinishEventId() (o int64) {
	if v != nil && v.DecisionFinishEventId != nil {
		return *v.DecisionFinishEventId
	}
	return
}

func (v *ResetWorkflowExecutionRequest) GetRequestId() (o string) {
	if v != nil && v.RequestId != nil {
		return *v.RequestId
	}
	return
}

type ResetWorkflowExecutionResponse struct {
	RunId *string
}

func (v *ResetWorkflowExecutionResponse) GetRunId() (o string) {
	if v != nil && v.RunId != nil {
		return *v.RunId
	}
	return
}

type RespondActivityTaskCanceledByIDRequest struct {
	Domain     *string
	WorkflowID *string
	RunID      *string
	ActivityID *string
	Details    []byte
	Identity   *string
}

func (v *RespondActivityTaskCanceledByIDRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *RespondActivityTaskCanceledByIDRequest) GetWorkflowID() (o string) {
	if v != nil && v.WorkflowID != nil {
		return *v.WorkflowID
	}
	return
}

func (v *RespondActivityTaskCanceledByIDRequest) GetRunID() (o string) {
	if v != nil && v.RunID != nil {
		return *v.RunID
	}
	return
}

func (v *RespondActivityTaskCanceledByIDRequest) GetActivityID() (o string) {
	if v != nil && v.ActivityID != nil {
		return *v.ActivityID
	}
	return
}

func (v *RespondActivityTaskCanceledByIDRequest) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *RespondActivityTaskCanceledByIDRequest) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

type RespondActivityTaskCanceledRequest struct {
	TaskToken []byte
	Details   []byte
	Identity  *string
}

func (v *RespondActivityTaskCanceledRequest) GetTaskToken() (o []byte) {
	if v != nil && v.TaskToken != nil {
		return v.TaskToken
	}
	return
}

func (v *RespondActivityTaskCanceledRequest) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *RespondActivityTaskCanceledRequest) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

type RespondActivityTaskCompletedByIDRequest struct {
	Domain     *string
	WorkflowID *string
	RunID      *string
	ActivityID *string
	Result     []byte
	Identity   *string
}

func (v *RespondActivityTaskCompletedByIDRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *RespondActivityTaskCompletedByIDRequest) GetWorkflowID() (o string) {
	if v != nil && v.WorkflowID != nil {
		return *v.WorkflowID
	}
	return
}

func (v *RespondActivityTaskCompletedByIDRequest) GetRunID() (o string) {
	if v != nil && v.RunID != nil {
		return *v.RunID
	}
	return
}

func (v *RespondActivityTaskCompletedByIDRequest) GetActivityID() (o string) {
	if v != nil && v.ActivityID != nil {
		return *v.ActivityID
	}
	return
}

func (v *RespondActivityTaskCompletedByIDRequest) GetResult() (o []byte) {
	if v != nil && v.Result != nil {
		return v.Result
	}
	return
}

func (v *RespondActivityTaskCompletedByIDRequest) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

type RespondActivityTaskCompletedRequest struct {
	TaskToken []byte
	Result    []byte
	Identity  *string
}

func (v *RespondActivityTaskCompletedRequest) GetTaskToken() (o []byte) {
	if v != nil && v.TaskToken != nil {
		return v.TaskToken
	}
	return
}

func (v *RespondActivityTaskCompletedRequest) GetResult() (o []byte) {
	if v != nil && v.Result != nil {
		return v.Result
	}
	return
}

func (v *RespondActivityTaskCompletedRequest) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
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

func (v *RespondActivityTaskFailedByIDRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *RespondActivityTaskFailedByIDRequest) GetWorkflowID() (o string) {
	if v != nil && v.WorkflowID != nil {
		return *v.WorkflowID
	}
	return
}

func (v *RespondActivityTaskFailedByIDRequest) GetRunID() (o string) {
	if v != nil && v.RunID != nil {
		return *v.RunID
	}
	return
}

func (v *RespondActivityTaskFailedByIDRequest) GetActivityID() (o string) {
	if v != nil && v.ActivityID != nil {
		return *v.ActivityID
	}
	return
}

func (v *RespondActivityTaskFailedByIDRequest) GetReason() (o string) {
	if v != nil && v.Reason != nil {
		return *v.Reason
	}
	return
}

func (v *RespondActivityTaskFailedByIDRequest) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *RespondActivityTaskFailedByIDRequest) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

type RespondActivityTaskFailedRequest struct {
	TaskToken []byte
	Reason    *string
	Details   []byte
	Identity  *string
}

func (v *RespondActivityTaskFailedRequest) GetTaskToken() (o []byte) {
	if v != nil && v.TaskToken != nil {
		return v.TaskToken
	}
	return
}

func (v *RespondActivityTaskFailedRequest) GetReason() (o string) {
	if v != nil && v.Reason != nil {
		return *v.Reason
	}
	return
}

func (v *RespondActivityTaskFailedRequest) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *RespondActivityTaskFailedRequest) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
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

func (v *RespondDecisionTaskCompletedRequest) GetTaskToken() (o []byte) {
	if v != nil && v.TaskToken != nil {
		return v.TaskToken
	}
	return
}

func (v *RespondDecisionTaskCompletedRequest) GetDecisions() (o []*Decision) {
	if v != nil && v.Decisions != nil {
		return v.Decisions
	}
	return
}

func (v *RespondDecisionTaskCompletedRequest) GetExecutionContext() (o []byte) {
	if v != nil && v.ExecutionContext != nil {
		return v.ExecutionContext
	}
	return
}

func (v *RespondDecisionTaskCompletedRequest) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

func (v *RespondDecisionTaskCompletedRequest) GetStickyAttributes() (o *StickyExecutionAttributes) {
	if v != nil && v.StickyAttributes != nil {
		return v.StickyAttributes
	}
	return
}

func (v *RespondDecisionTaskCompletedRequest) GetReturnNewDecisionTask() (o bool) {
	if v != nil && v.ReturnNewDecisionTask != nil {
		return *v.ReturnNewDecisionTask
	}
	return
}

func (v *RespondDecisionTaskCompletedRequest) GetForceCreateNewDecisionTask() (o bool) {
	if v != nil && v.ForceCreateNewDecisionTask != nil {
		return *v.ForceCreateNewDecisionTask
	}
	return
}

func (v *RespondDecisionTaskCompletedRequest) GetBinaryChecksum() (o string) {
	if v != nil && v.BinaryChecksum != nil {
		return *v.BinaryChecksum
	}
	return
}

func (v *RespondDecisionTaskCompletedRequest) GetQueryResults() (o map[string]*WorkflowQueryResult) {
	if v != nil && v.QueryResults != nil {
		return v.QueryResults
	}
	return
}

type RespondDecisionTaskCompletedResponse struct {
	DecisionTask *PollForDecisionTaskResponse
}

func (v *RespondDecisionTaskCompletedResponse) GetDecisionTask() (o *PollForDecisionTaskResponse) {
	if v != nil && v.DecisionTask != nil {
		return v.DecisionTask
	}
	return
}

type RespondDecisionTaskFailedRequest struct {
	TaskToken      []byte
	Cause          *DecisionTaskFailedCause
	Details        []byte
	Identity       *string
	BinaryChecksum *string
}

func (v *RespondDecisionTaskFailedRequest) GetTaskToken() (o []byte) {
	if v != nil && v.TaskToken != nil {
		return v.TaskToken
	}
	return
}

func (v *RespondDecisionTaskFailedRequest) GetCause() (o DecisionTaskFailedCause) {
	if v != nil && v.Cause != nil {
		return *v.Cause
	}
	return
}

func (v *RespondDecisionTaskFailedRequest) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *RespondDecisionTaskFailedRequest) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

func (v *RespondDecisionTaskFailedRequest) GetBinaryChecksum() (o string) {
	if v != nil && v.BinaryChecksum != nil {
		return *v.BinaryChecksum
	}
	return
}

type RespondQueryTaskCompletedRequest struct {
	TaskToken         []byte
	CompletedType     *QueryTaskCompletedType
	QueryResult       []byte
	ErrorMessage      *string
	WorkerVersionInfo *WorkerVersionInfo
}

func (v *RespondQueryTaskCompletedRequest) GetTaskToken() (o []byte) {
	if v != nil && v.TaskToken != nil {
		return v.TaskToken
	}
	return
}

func (v *RespondQueryTaskCompletedRequest) GetCompletedType() (o QueryTaskCompletedType) {
	if v != nil && v.CompletedType != nil {
		return *v.CompletedType
	}
	return
}

func (v *RespondQueryTaskCompletedRequest) GetQueryResult() (o []byte) {
	if v != nil && v.QueryResult != nil {
		return v.QueryResult
	}
	return
}

func (v *RespondQueryTaskCompletedRequest) GetErrorMessage() (o string) {
	if v != nil && v.ErrorMessage != nil {
		return *v.ErrorMessage
	}
	return
}

func (v *RespondQueryTaskCompletedRequest) GetWorkerVersionInfo() (o *WorkerVersionInfo) {
	if v != nil && v.WorkerVersionInfo != nil {
		return v.WorkerVersionInfo
	}
	return
}

type RetryPolicy struct {
	InitialIntervalInSeconds    *int32
	BackoffCoefficient          *float64
	MaximumIntervalInSeconds    *int32
	MaximumAttempts             *int32
	NonRetriableErrorReasons    []string
	ExpirationIntervalInSeconds *int32
}

func (v *RetryPolicy) GetInitialIntervalInSeconds() (o int32) {
	if v != nil && v.InitialIntervalInSeconds != nil {
		return *v.InitialIntervalInSeconds
	}
	return
}

func (v *RetryPolicy) GetBackoffCoefficient() (o float64) {
	if v != nil && v.BackoffCoefficient != nil {
		return *v.BackoffCoefficient
	}
	return
}

func (v *RetryPolicy) GetMaximumIntervalInSeconds() (o int32) {
	if v != nil && v.MaximumIntervalInSeconds != nil {
		return *v.MaximumIntervalInSeconds
	}
	return
}

func (v *RetryPolicy) GetMaximumAttempts() (o int32) {
	if v != nil && v.MaximumAttempts != nil {
		return *v.MaximumAttempts
	}
	return
}

func (v *RetryPolicy) GetNonRetriableErrorReasons() (o []string) {
	if v != nil && v.NonRetriableErrorReasons != nil {
		return v.NonRetriableErrorReasons
	}
	return
}

func (v *RetryPolicy) GetExpirationIntervalInSeconds() (o int32) {
	if v != nil && v.ExpirationIntervalInSeconds != nil {
		return *v.ExpirationIntervalInSeconds
	}
	return
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

func (v *RetryTaskV2Error) GetMessage() (o string) {
	if v != nil {
		return v.Message
	}
	return
}

func (v *RetryTaskV2Error) GetDomainId() (o string) {
	if v != nil && v.DomainId != nil {
		return *v.DomainId
	}
	return
}

func (v *RetryTaskV2Error) GetWorkflowId() (o string) {
	if v != nil && v.WorkflowId != nil {
		return *v.WorkflowId
	}
	return
}

func (v *RetryTaskV2Error) GetRunId() (o string) {
	if v != nil && v.RunId != nil {
		return *v.RunId
	}
	return
}

func (v *RetryTaskV2Error) GetStartEventId() (o int64) {
	if v != nil && v.StartEventId != nil {
		return *v.StartEventId
	}
	return
}

func (v *RetryTaskV2Error) GetStartEventVersion() (o int64) {
	if v != nil && v.StartEventVersion != nil {
		return *v.StartEventVersion
	}
	return
}

func (v *RetryTaskV2Error) GetEndEventId() (o int64) {
	if v != nil && v.EndEventId != nil {
		return *v.EndEventId
	}
	return
}

func (v *RetryTaskV2Error) GetEndEventVersion() (o int64) {
	if v != nil && v.EndEventVersion != nil {
		return *v.EndEventVersion
	}
	return
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

func (v *ScheduleActivityTaskDecisionAttributes) GetActivityId() (o string) {
	if v != nil && v.ActivityId != nil {
		return *v.ActivityId
	}
	return
}

func (v *ScheduleActivityTaskDecisionAttributes) GetActivityType() (o *ActivityType) {
	if v != nil && v.ActivityType != nil {
		return v.ActivityType
	}
	return
}

func (v *ScheduleActivityTaskDecisionAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *ScheduleActivityTaskDecisionAttributes) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

func (v *ScheduleActivityTaskDecisionAttributes) GetInput() (o []byte) {
	if v != nil && v.Input != nil {
		return v.Input
	}
	return
}

func (v *ScheduleActivityTaskDecisionAttributes) GetScheduleToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.ScheduleToCloseTimeoutSeconds != nil {
		return *v.ScheduleToCloseTimeoutSeconds
	}
	return
}

func (v *ScheduleActivityTaskDecisionAttributes) GetScheduleToStartTimeoutSeconds() (o int32) {
	if v != nil && v.ScheduleToStartTimeoutSeconds != nil {
		return *v.ScheduleToStartTimeoutSeconds
	}
	return
}

func (v *ScheduleActivityTaskDecisionAttributes) GetStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.StartToCloseTimeoutSeconds != nil {
		return *v.StartToCloseTimeoutSeconds
	}
	return
}

func (v *ScheduleActivityTaskDecisionAttributes) GetHeartbeatTimeoutSeconds() (o int32) {
	if v != nil && v.HeartbeatTimeoutSeconds != nil {
		return *v.HeartbeatTimeoutSeconds
	}
	return
}

func (v *ScheduleActivityTaskDecisionAttributes) GetRetryPolicy() (o *RetryPolicy) {
	if v != nil && v.RetryPolicy != nil {
		return v.RetryPolicy
	}
	return
}

func (v *ScheduleActivityTaskDecisionAttributes) GetHeader() (o *Header) {
	if v != nil && v.Header != nil {
		return v.Header
	}
	return
}

type SearchAttributes struct {
	IndexedFields map[string][]byte
}

func (v *SearchAttributes) GetIndexedFields() (o map[string][]byte) {
	if v != nil && v.IndexedFields != nil {
		return v.IndexedFields
	}
	return
}

type ServiceBusyError struct {
	Message string
}

func (v *ServiceBusyError) GetMessage() (o string) {
	if v != nil {
		return v.Message
	}
	return
}

type SignalExternalWorkflowExecutionDecisionAttributes struct {
	Domain            *string
	Execution         *WorkflowExecution
	SignalName        *string
	Input             []byte
	Control           []byte
	ChildWorkflowOnly *bool
}

func (v *SignalExternalWorkflowExecutionDecisionAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *SignalExternalWorkflowExecutionDecisionAttributes) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
}

func (v *SignalExternalWorkflowExecutionDecisionAttributes) GetSignalName() (o string) {
	if v != nil && v.SignalName != nil {
		return *v.SignalName
	}
	return
}

func (v *SignalExternalWorkflowExecutionDecisionAttributes) GetInput() (o []byte) {
	if v != nil && v.Input != nil {
		return v.Input
	}
	return
}

func (v *SignalExternalWorkflowExecutionDecisionAttributes) GetControl() (o []byte) {
	if v != nil && v.Control != nil {
		return v.Control
	}
	return
}

func (v *SignalExternalWorkflowExecutionDecisionAttributes) GetChildWorkflowOnly() (o bool) {
	if v != nil && v.ChildWorkflowOnly != nil {
		return *v.ChildWorkflowOnly
	}
	return
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

func (v *SignalExternalWorkflowExecutionFailedEventAttributes) GetCause() (o SignalExternalWorkflowExecutionFailedCause) {
	if v != nil && v.Cause != nil {
		return *v.Cause
	}
	return
}

func (v *SignalExternalWorkflowExecutionFailedEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
}

func (v *SignalExternalWorkflowExecutionFailedEventAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *SignalExternalWorkflowExecutionFailedEventAttributes) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *SignalExternalWorkflowExecutionFailedEventAttributes) GetInitiatedEventId() (o int64) {
	if v != nil && v.InitiatedEventId != nil {
		return *v.InitiatedEventId
	}
	return
}

func (v *SignalExternalWorkflowExecutionFailedEventAttributes) GetControl() (o []byte) {
	if v != nil && v.Control != nil {
		return v.Control
	}
	return
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

func (v *SignalExternalWorkflowExecutionInitiatedEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
}

func (v *SignalExternalWorkflowExecutionInitiatedEventAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *SignalExternalWorkflowExecutionInitiatedEventAttributes) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *SignalExternalWorkflowExecutionInitiatedEventAttributes) GetSignalName() (o string) {
	if v != nil && v.SignalName != nil {
		return *v.SignalName
	}
	return
}

func (v *SignalExternalWorkflowExecutionInitiatedEventAttributes) GetInput() (o []byte) {
	if v != nil && v.Input != nil {
		return v.Input
	}
	return
}

func (v *SignalExternalWorkflowExecutionInitiatedEventAttributes) GetControl() (o []byte) {
	if v != nil && v.Control != nil {
		return v.Control
	}
	return
}

func (v *SignalExternalWorkflowExecutionInitiatedEventAttributes) GetChildWorkflowOnly() (o bool) {
	if v != nil && v.ChildWorkflowOnly != nil {
		return *v.ChildWorkflowOnly
	}
	return
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

func (v *SignalWithStartWorkflowExecutionRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *SignalWithStartWorkflowExecutionRequest) GetWorkflowId() (o string) {
	if v != nil && v.WorkflowId != nil {
		return *v.WorkflowId
	}
	return
}

func (v *SignalWithStartWorkflowExecutionRequest) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

func (v *SignalWithStartWorkflowExecutionRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

func (v *SignalWithStartWorkflowExecutionRequest) GetInput() (o []byte) {
	if v != nil && v.Input != nil {
		return v.Input
	}
	return
}

func (v *SignalWithStartWorkflowExecutionRequest) GetExecutionStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.ExecutionStartToCloseTimeoutSeconds != nil {
		return *v.ExecutionStartToCloseTimeoutSeconds
	}
	return
}

func (v *SignalWithStartWorkflowExecutionRequest) GetTaskStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.TaskStartToCloseTimeoutSeconds != nil {
		return *v.TaskStartToCloseTimeoutSeconds
	}
	return
}

func (v *SignalWithStartWorkflowExecutionRequest) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

func (v *SignalWithStartWorkflowExecutionRequest) GetRequestId() (o string) {
	if v != nil && v.RequestId != nil {
		return *v.RequestId
	}
	return
}

func (v *SignalWithStartWorkflowExecutionRequest) GetWorkflowIdReusePolicy() (o WorkflowIdReusePolicy) {
	if v != nil && v.WorkflowIdReusePolicy != nil {
		return *v.WorkflowIdReusePolicy
	}
	return
}

func (v *SignalWithStartWorkflowExecutionRequest) GetSignalName() (o string) {
	if v != nil && v.SignalName != nil {
		return *v.SignalName
	}
	return
}

func (v *SignalWithStartWorkflowExecutionRequest) GetSignalInput() (o []byte) {
	if v != nil && v.SignalInput != nil {
		return v.SignalInput
	}
	return
}

func (v *SignalWithStartWorkflowExecutionRequest) GetControl() (o []byte) {
	if v != nil && v.Control != nil {
		return v.Control
	}
	return
}

func (v *SignalWithStartWorkflowExecutionRequest) GetRetryPolicy() (o *RetryPolicy) {
	if v != nil && v.RetryPolicy != nil {
		return v.RetryPolicy
	}
	return
}

func (v *SignalWithStartWorkflowExecutionRequest) GetCronSchedule() (o string) {
	if v != nil && v.CronSchedule != nil {
		return *v.CronSchedule
	}
	return
}

func (v *SignalWithStartWorkflowExecutionRequest) GetMemo() (o *Memo) {
	if v != nil && v.Memo != nil {
		return v.Memo
	}
	return
}

func (v *SignalWithStartWorkflowExecutionRequest) GetSearchAttributes() (o *SearchAttributes) {
	if v != nil && v.SearchAttributes != nil {
		return v.SearchAttributes
	}
	return
}

func (v *SignalWithStartWorkflowExecutionRequest) GetHeader() (o *Header) {
	if v != nil && v.Header != nil {
		return v.Header
	}
	return
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

func (v *SignalWorkflowExecutionRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *SignalWorkflowExecutionRequest) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *SignalWorkflowExecutionRequest) GetSignalName() (o string) {
	if v != nil && v.SignalName != nil {
		return *v.SignalName
	}
	return
}

func (v *SignalWorkflowExecutionRequest) GetInput() (o []byte) {
	if v != nil && v.Input != nil {
		return v.Input
	}
	return
}

func (v *SignalWorkflowExecutionRequest) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

func (v *SignalWorkflowExecutionRequest) GetRequestId() (o string) {
	if v != nil && v.RequestId != nil {
		return *v.RequestId
	}
	return
}

func (v *SignalWorkflowExecutionRequest) GetControl() (o []byte) {
	if v != nil && v.Control != nil {
		return v.Control
	}
	return
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

func (v *StartChildWorkflowExecutionDecisionAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *StartChildWorkflowExecutionDecisionAttributes) GetWorkflowId() (o string) {
	if v != nil && v.WorkflowId != nil {
		return *v.WorkflowId
	}
	return
}

func (v *StartChildWorkflowExecutionDecisionAttributes) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

func (v *StartChildWorkflowExecutionDecisionAttributes) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

func (v *StartChildWorkflowExecutionDecisionAttributes) GetInput() (o []byte) {
	if v != nil && v.Input != nil {
		return v.Input
	}
	return
}

func (v *StartChildWorkflowExecutionDecisionAttributes) GetExecutionStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.ExecutionStartToCloseTimeoutSeconds != nil {
		return *v.ExecutionStartToCloseTimeoutSeconds
	}
	return
}

func (v *StartChildWorkflowExecutionDecisionAttributes) GetTaskStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.TaskStartToCloseTimeoutSeconds != nil {
		return *v.TaskStartToCloseTimeoutSeconds
	}
	return
}

func (v *StartChildWorkflowExecutionDecisionAttributes) GetParentClosePolicy() (o ParentClosePolicy) {
	if v != nil && v.ParentClosePolicy != nil {
		return *v.ParentClosePolicy
	}
	return
}

func (v *StartChildWorkflowExecutionDecisionAttributes) GetControl() (o []byte) {
	if v != nil && v.Control != nil {
		return v.Control
	}
	return
}

func (v *StartChildWorkflowExecutionDecisionAttributes) GetWorkflowIdReusePolicy() (o WorkflowIdReusePolicy) {
	if v != nil && v.WorkflowIdReusePolicy != nil {
		return *v.WorkflowIdReusePolicy
	}
	return
}

func (v *StartChildWorkflowExecutionDecisionAttributes) GetRetryPolicy() (o *RetryPolicy) {
	if v != nil && v.RetryPolicy != nil {
		return v.RetryPolicy
	}
	return
}

func (v *StartChildWorkflowExecutionDecisionAttributes) GetCronSchedule() (o string) {
	if v != nil && v.CronSchedule != nil {
		return *v.CronSchedule
	}
	return
}

func (v *StartChildWorkflowExecutionDecisionAttributes) GetHeader() (o *Header) {
	if v != nil && v.Header != nil {
		return v.Header
	}
	return
}

func (v *StartChildWorkflowExecutionDecisionAttributes) GetMemo() (o *Memo) {
	if v != nil && v.Memo != nil {
		return v.Memo
	}
	return
}

func (v *StartChildWorkflowExecutionDecisionAttributes) GetSearchAttributes() (o *SearchAttributes) {
	if v != nil && v.SearchAttributes != nil {
		return v.SearchAttributes
	}
	return
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

func (v *StartChildWorkflowExecutionFailedEventAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *StartChildWorkflowExecutionFailedEventAttributes) GetWorkflowId() (o string) {
	if v != nil && v.WorkflowId != nil {
		return *v.WorkflowId
	}
	return
}

func (v *StartChildWorkflowExecutionFailedEventAttributes) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

func (v *StartChildWorkflowExecutionFailedEventAttributes) GetCause() (o ChildWorkflowExecutionFailedCause) {
	if v != nil && v.Cause != nil {
		return *v.Cause
	}
	return
}

func (v *StartChildWorkflowExecutionFailedEventAttributes) GetControl() (o []byte) {
	if v != nil && v.Control != nil {
		return v.Control
	}
	return
}

func (v *StartChildWorkflowExecutionFailedEventAttributes) GetInitiatedEventId() (o int64) {
	if v != nil && v.InitiatedEventId != nil {
		return *v.InitiatedEventId
	}
	return
}

func (v *StartChildWorkflowExecutionFailedEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
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

func (v *StartChildWorkflowExecutionInitiatedEventAttributes) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *StartChildWorkflowExecutionInitiatedEventAttributes) GetWorkflowId() (o string) {
	if v != nil && v.WorkflowId != nil {
		return *v.WorkflowId
	}
	return
}

func (v *StartChildWorkflowExecutionInitiatedEventAttributes) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

func (v *StartChildWorkflowExecutionInitiatedEventAttributes) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

func (v *StartChildWorkflowExecutionInitiatedEventAttributes) GetInput() (o []byte) {
	if v != nil && v.Input != nil {
		return v.Input
	}
	return
}

func (v *StartChildWorkflowExecutionInitiatedEventAttributes) GetExecutionStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.ExecutionStartToCloseTimeoutSeconds != nil {
		return *v.ExecutionStartToCloseTimeoutSeconds
	}
	return
}

func (v *StartChildWorkflowExecutionInitiatedEventAttributes) GetTaskStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.TaskStartToCloseTimeoutSeconds != nil {
		return *v.TaskStartToCloseTimeoutSeconds
	}
	return
}

func (v *StartChildWorkflowExecutionInitiatedEventAttributes) GetParentClosePolicy() (o ParentClosePolicy) {
	if v != nil && v.ParentClosePolicy != nil {
		return *v.ParentClosePolicy
	}
	return
}

func (v *StartChildWorkflowExecutionInitiatedEventAttributes) GetControl() (o []byte) {
	if v != nil && v.Control != nil {
		return v.Control
	}
	return
}

func (v *StartChildWorkflowExecutionInitiatedEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
}

func (v *StartChildWorkflowExecutionInitiatedEventAttributes) GetWorkflowIdReusePolicy() (o WorkflowIdReusePolicy) {
	if v != nil && v.WorkflowIdReusePolicy != nil {
		return *v.WorkflowIdReusePolicy
	}
	return
}

func (v *StartChildWorkflowExecutionInitiatedEventAttributes) GetRetryPolicy() (o *RetryPolicy) {
	if v != nil && v.RetryPolicy != nil {
		return v.RetryPolicy
	}
	return
}

func (v *StartChildWorkflowExecutionInitiatedEventAttributes) GetCronSchedule() (o string) {
	if v != nil && v.CronSchedule != nil {
		return *v.CronSchedule
	}
	return
}

func (v *StartChildWorkflowExecutionInitiatedEventAttributes) GetHeader() (o *Header) {
	if v != nil && v.Header != nil {
		return v.Header
	}
	return
}

func (v *StartChildWorkflowExecutionInitiatedEventAttributes) GetMemo() (o *Memo) {
	if v != nil && v.Memo != nil {
		return v.Memo
	}
	return
}

func (v *StartChildWorkflowExecutionInitiatedEventAttributes) GetSearchAttributes() (o *SearchAttributes) {
	if v != nil && v.SearchAttributes != nil {
		return v.SearchAttributes
	}
	return
}

type StartTimeFilter struct {
	EarliestTime *int64
	LatestTime   *int64
}

func (v *StartTimeFilter) GetEarliestTime() (o int64) {
	if v != nil && v.EarliestTime != nil {
		return *v.EarliestTime
	}
	return
}

func (v *StartTimeFilter) GetLatestTime() (o int64) {
	if v != nil && v.LatestTime != nil {
		return *v.LatestTime
	}
	return
}

type StartTimerDecisionAttributes struct {
	TimerId                   *string
	StartToFireTimeoutSeconds *int64
}

func (v *StartTimerDecisionAttributes) GetTimerId() (o string) {
	if v != nil && v.TimerId != nil {
		return *v.TimerId
	}
	return
}

func (v *StartTimerDecisionAttributes) GetStartToFireTimeoutSeconds() (o int64) {
	if v != nil && v.StartToFireTimeoutSeconds != nil {
		return *v.StartToFireTimeoutSeconds
	}
	return
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

func (v *StartWorkflowExecutionRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *StartWorkflowExecutionRequest) GetWorkflowId() (o string) {
	if v != nil && v.WorkflowId != nil {
		return *v.WorkflowId
	}
	return
}

func (v *StartWorkflowExecutionRequest) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

func (v *StartWorkflowExecutionRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

func (v *StartWorkflowExecutionRequest) GetInput() (o []byte) {
	if v != nil && v.Input != nil {
		return v.Input
	}
	return
}

func (v *StartWorkflowExecutionRequest) GetExecutionStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.ExecutionStartToCloseTimeoutSeconds != nil {
		return *v.ExecutionStartToCloseTimeoutSeconds
	}
	return
}

func (v *StartWorkflowExecutionRequest) GetTaskStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.TaskStartToCloseTimeoutSeconds != nil {
		return *v.TaskStartToCloseTimeoutSeconds
	}
	return
}

func (v *StartWorkflowExecutionRequest) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

func (v *StartWorkflowExecutionRequest) GetRequestId() (o string) {
	if v != nil && v.RequestId != nil {
		return *v.RequestId
	}
	return
}

func (v *StartWorkflowExecutionRequest) GetWorkflowIdReusePolicy() (o WorkflowIdReusePolicy) {
	if v != nil && v.WorkflowIdReusePolicy != nil {
		return *v.WorkflowIdReusePolicy
	}
	return
}

func (v *StartWorkflowExecutionRequest) GetRetryPolicy() (o *RetryPolicy) {
	if v != nil && v.RetryPolicy != nil {
		return v.RetryPolicy
	}
	return
}

func (v *StartWorkflowExecutionRequest) GetCronSchedule() (o string) {
	if v != nil && v.CronSchedule != nil {
		return *v.CronSchedule
	}
	return
}

func (v *StartWorkflowExecutionRequest) GetMemo() (o *Memo) {
	if v != nil && v.Memo != nil {
		return v.Memo
	}
	return
}

func (v *StartWorkflowExecutionRequest) GetSearchAttributes() (o *SearchAttributes) {
	if v != nil && v.SearchAttributes != nil {
		return v.SearchAttributes
	}
	return
}

func (v *StartWorkflowExecutionRequest) GetHeader() (o *Header) {
	if v != nil && v.Header != nil {
		return v.Header
	}
	return
}

type StartWorkflowExecutionResponse struct {
	RunId *string
}

func (v *StartWorkflowExecutionResponse) GetRunId() (o string) {
	if v != nil && v.RunId != nil {
		return *v.RunId
	}
	return
}

type StickyExecutionAttributes struct {
	WorkerTaskList                *TaskList
	ScheduleToStartTimeoutSeconds *int32
}

func (v *StickyExecutionAttributes) GetWorkerTaskList() (o *TaskList) {
	if v != nil && v.WorkerTaskList != nil {
		return v.WorkerTaskList
	}
	return
}

func (v *StickyExecutionAttributes) GetScheduleToStartTimeoutSeconds() (o int32) {
	if v != nil && v.ScheduleToStartTimeoutSeconds != nil {
		return *v.ScheduleToStartTimeoutSeconds
	}
	return
}

type SupportedClientVersions struct {
	GoSdk   *string
	JavaSdk *string
}

func (v *SupportedClientVersions) GetGoSdk() (o string) {
	if v != nil && v.GoSdk != nil {
		return *v.GoSdk
	}
	return
}

func (v *SupportedClientVersions) GetJavaSdk() (o string) {
	if v != nil && v.JavaSdk != nil {
		return *v.JavaSdk
	}
	return
}

type TaskIDBlock struct {
	StartID *int64
	EndID   *int64
}

func (v *TaskIDBlock) GetStartID() (o int64) {
	if v != nil && v.StartID != nil {
		return *v.StartID
	}
	return
}

func (v *TaskIDBlock) GetEndID() (o int64) {
	if v != nil && v.EndID != nil {
		return *v.EndID
	}
	return
}

type TaskList struct {
	Name *string
	Kind *TaskListKind
}

func (v *TaskList) GetName() (o string) {
	if v != nil && v.Name != nil {
		return *v.Name
	}
	return
}

func (v *TaskList) GetKind() (o TaskListKind) {
	if v != nil && v.Kind != nil {
		return *v.Kind
	}
	return
}

type TaskListKind int32

const (
	TaskListKindNormal TaskListKind = iota
	TaskListKindSticky
)

type TaskListMetadata struct {
	MaxTasksPerSecond *float64
}

func (v *TaskListMetadata) GetMaxTasksPerSecond() (o float64) {
	if v != nil && v.MaxTasksPerSecond != nil {
		return *v.MaxTasksPerSecond
	}
	return
}

type TaskListPartitionMetadata struct {
	Key           *string
	OwnerHostName *string
}

func (v *TaskListPartitionMetadata) GetKey() (o string) {
	if v != nil && v.Key != nil {
		return *v.Key
	}
	return
}

func (v *TaskListPartitionMetadata) GetOwnerHostName() (o string) {
	if v != nil && v.OwnerHostName != nil {
		return *v.OwnerHostName
	}
	return
}

type TaskListStatus struct {
	BacklogCountHint *int64
	ReadLevel        *int64
	AckLevel         *int64
	RatePerSecond    *float64
	TaskIDBlock      *TaskIDBlock
}

func (v *TaskListStatus) GetBacklogCountHint() (o int64) {
	if v != nil && v.BacklogCountHint != nil {
		return *v.BacklogCountHint
	}
	return
}

func (v *TaskListStatus) GetReadLevel() (o int64) {
	if v != nil && v.ReadLevel != nil {
		return *v.ReadLevel
	}
	return
}

func (v *TaskListStatus) GetAckLevel() (o int64) {
	if v != nil && v.AckLevel != nil {
		return *v.AckLevel
	}
	return
}

func (v *TaskListStatus) GetRatePerSecond() (o float64) {
	if v != nil && v.RatePerSecond != nil {
		return *v.RatePerSecond
	}
	return
}

func (v *TaskListStatus) GetTaskIDBlock() (o *TaskIDBlock) {
	if v != nil && v.TaskIDBlock != nil {
		return v.TaskIDBlock
	}
	return
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

func (v *TerminateWorkflowExecutionRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

func (v *TerminateWorkflowExecutionRequest) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

func (v *TerminateWorkflowExecutionRequest) GetReason() (o string) {
	if v != nil && v.Reason != nil {
		return *v.Reason
	}
	return
}

func (v *TerminateWorkflowExecutionRequest) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *TerminateWorkflowExecutionRequest) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
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

func (v *TimerCanceledEventAttributes) GetTimerId() (o string) {
	if v != nil && v.TimerId != nil {
		return *v.TimerId
	}
	return
}

func (v *TimerCanceledEventAttributes) GetStartedEventId() (o int64) {
	if v != nil && v.StartedEventId != nil {
		return *v.StartedEventId
	}
	return
}

func (v *TimerCanceledEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
}

func (v *TimerCanceledEventAttributes) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

type TimerFiredEventAttributes struct {
	TimerId        *string
	StartedEventId *int64
}

func (v *TimerFiredEventAttributes) GetTimerId() (o string) {
	if v != nil && v.TimerId != nil {
		return *v.TimerId
	}
	return
}

func (v *TimerFiredEventAttributes) GetStartedEventId() (o int64) {
	if v != nil && v.StartedEventId != nil {
		return *v.StartedEventId
	}
	return
}

type TimerStartedEventAttributes struct {
	TimerId                      *string
	StartToFireTimeoutSeconds    *int64
	DecisionTaskCompletedEventId *int64
}

func (v *TimerStartedEventAttributes) GetTimerId() (o string) {
	if v != nil && v.TimerId != nil {
		return *v.TimerId
	}
	return
}

func (v *TimerStartedEventAttributes) GetStartToFireTimeoutSeconds() (o int64) {
	if v != nil && v.StartToFireTimeoutSeconds != nil {
		return *v.StartToFireTimeoutSeconds
	}
	return
}

func (v *TimerStartedEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
}

type TransientDecisionInfo struct {
	ScheduledEvent *HistoryEvent
	StartedEvent   *HistoryEvent
}

func (v *TransientDecisionInfo) GetScheduledEvent() (o *HistoryEvent) {
	if v != nil && v.ScheduledEvent != nil {
		return v.ScheduledEvent
	}
	return
}

func (v *TransientDecisionInfo) GetStartedEvent() (o *HistoryEvent) {
	if v != nil && v.StartedEvent != nil {
		return v.StartedEvent
	}
	return
}

type UpdateDomainInfo struct {
	Description *string
	OwnerEmail  *string
	Data        map[string]string
}

func (v *UpdateDomainInfo) GetDescription() (o string) {
	if v != nil && v.Description != nil {
		return *v.Description
	}
	return
}

func (v *UpdateDomainInfo) GetOwnerEmail() (o string) {
	if v != nil && v.OwnerEmail != nil {
		return *v.OwnerEmail
	}
	return
}

func (v *UpdateDomainInfo) GetData() (o map[string]string) {
	if v != nil && v.Data != nil {
		return v.Data
	}
	return
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

func (v *UpdateDomainRequest) GetName() (o string) {
	if v != nil && v.Name != nil {
		return *v.Name
	}
	return
}

func (v *UpdateDomainRequest) GetUpdatedInfo() (o *UpdateDomainInfo) {
	if v != nil && v.UpdatedInfo != nil {
		return v.UpdatedInfo
	}
	return
}

func (v *UpdateDomainRequest) GetConfiguration() (o *DomainConfiguration) {
	if v != nil && v.Configuration != nil {
		return v.Configuration
	}
	return
}

func (v *UpdateDomainRequest) GetReplicationConfiguration() (o *DomainReplicationConfiguration) {
	if v != nil && v.ReplicationConfiguration != nil {
		return v.ReplicationConfiguration
	}
	return
}

func (v *UpdateDomainRequest) GetSecurityToken() (o string) {
	if v != nil && v.SecurityToken != nil {
		return *v.SecurityToken
	}
	return
}

func (v *UpdateDomainRequest) GetDeleteBadBinary() (o string) {
	if v != nil && v.DeleteBadBinary != nil {
		return *v.DeleteBadBinary
	}
	return
}

func (v *UpdateDomainRequest) GetFailoverTimeoutInSeconds() (o int32) {
	if v != nil && v.FailoverTimeoutInSeconds != nil {
		return *v.FailoverTimeoutInSeconds
	}
	return
}

type UpdateDomainResponse struct {
	DomainInfo               *DomainInfo
	Configuration            *DomainConfiguration
	ReplicationConfiguration *DomainReplicationConfiguration
	FailoverVersion          *int64
	IsGlobalDomain           *bool
}

func (v *UpdateDomainResponse) GetDomainInfo() (o *DomainInfo) {
	if v != nil && v.DomainInfo != nil {
		return v.DomainInfo
	}
	return
}

func (v *UpdateDomainResponse) GetConfiguration() (o *DomainConfiguration) {
	if v != nil && v.Configuration != nil {
		return v.Configuration
	}
	return
}

func (v *UpdateDomainResponse) GetReplicationConfiguration() (o *DomainReplicationConfiguration) {
	if v != nil && v.ReplicationConfiguration != nil {
		return v.ReplicationConfiguration
	}
	return
}

func (v *UpdateDomainResponse) GetFailoverVersion() (o int64) {
	if v != nil && v.FailoverVersion != nil {
		return *v.FailoverVersion
	}
	return
}

func (v *UpdateDomainResponse) GetIsGlobalDomain() (o bool) {
	if v != nil && v.IsGlobalDomain != nil {
		return *v.IsGlobalDomain
	}
	return
}

type UpsertWorkflowSearchAttributesDecisionAttributes struct {
	SearchAttributes *SearchAttributes
}

func (v *UpsertWorkflowSearchAttributesDecisionAttributes) GetSearchAttributes() (o *SearchAttributes) {
	if v != nil && v.SearchAttributes != nil {
		return v.SearchAttributes
	}
	return
}

type UpsertWorkflowSearchAttributesEventAttributes struct {
	DecisionTaskCompletedEventId *int64
	SearchAttributes             *SearchAttributes
}

func (v *UpsertWorkflowSearchAttributesEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
}

func (v *UpsertWorkflowSearchAttributesEventAttributes) GetSearchAttributes() (o *SearchAttributes) {
	if v != nil && v.SearchAttributes != nil {
		return v.SearchAttributes
	}
	return
}

type VersionHistories struct {
	CurrentVersionHistoryIndex *int32
	Histories                  []*VersionHistory
}

func (v *VersionHistories) GetCurrentVersionHistoryIndex() (o int32) {
	if v != nil && v.CurrentVersionHistoryIndex != nil {
		return *v.CurrentVersionHistoryIndex
	}
	return
}

func (v *VersionHistories) GetHistories() (o []*VersionHistory) {
	if v != nil && v.Histories != nil {
		return v.Histories
	}
	return
}

type VersionHistory struct {
	BranchToken []byte
	Items       []*VersionHistoryItem
}

func (v *VersionHistory) GetBranchToken() (o []byte) {
	if v != nil && v.BranchToken != nil {
		return v.BranchToken
	}
	return
}

func (v *VersionHistory) GetItems() (o []*VersionHistoryItem) {
	if v != nil && v.Items != nil {
		return v.Items
	}
	return
}

type VersionHistoryItem struct {
	EventID *int64
	Version *int64
}

func (v *VersionHistoryItem) GetEventID() (o int64) {
	if v != nil && v.EventID != nil {
		return *v.EventID
	}
	return
}

func (v *VersionHistoryItem) GetVersion() (o int64) {
	if v != nil && v.Version != nil {
		return *v.Version
	}
	return
}

type WorkerVersionInfo struct {
	Impl           *string
	FeatureVersion *string
}

func (v *WorkerVersionInfo) GetImpl() (o string) {
	if v != nil && v.Impl != nil {
		return *v.Impl
	}
	return
}

func (v *WorkerVersionInfo) GetFeatureVersion() (o string) {
	if v != nil && v.FeatureVersion != nil {
		return *v.FeatureVersion
	}
	return
}

type WorkflowExecution struct {
	WorkflowId *string
	RunId      *string
}

func (v *WorkflowExecution) GetWorkflowId() (o string) {
	if v != nil && v.WorkflowId != nil {
		return *v.WorkflowId
	}
	return
}

func (v *WorkflowExecution) GetRunId() (o string) {
	if v != nil && v.RunId != nil {
		return *v.RunId
	}
	return
}

type WorkflowExecutionAlreadyStartedError struct {
	Message        *string
	StartRequestId *string
	RunId          *string
}

func (v *WorkflowExecutionAlreadyStartedError) GetMessage() (o string) {
	if v != nil && v.Message != nil {
		return *v.Message
	}
	return
}

func (v *WorkflowExecutionAlreadyStartedError) GetStartRequestId() (o string) {
	if v != nil && v.StartRequestId != nil {
		return *v.StartRequestId
	}
	return
}

func (v *WorkflowExecutionAlreadyStartedError) GetRunId() (o string) {
	if v != nil && v.RunId != nil {
		return *v.RunId
	}
	return
}

type WorkflowExecutionCancelRequestedEventAttributes struct {
	Cause                     *string
	ExternalInitiatedEventId  *int64
	ExternalWorkflowExecution *WorkflowExecution
	Identity                  *string
}

func (v *WorkflowExecutionCancelRequestedEventAttributes) GetCause() (o string) {
	if v != nil && v.Cause != nil {
		return *v.Cause
	}
	return
}

func (v *WorkflowExecutionCancelRequestedEventAttributes) GetExternalInitiatedEventId() (o int64) {
	if v != nil && v.ExternalInitiatedEventId != nil {
		return *v.ExternalInitiatedEventId
	}
	return
}

func (v *WorkflowExecutionCancelRequestedEventAttributes) GetExternalWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.ExternalWorkflowExecution != nil {
		return v.ExternalWorkflowExecution
	}
	return
}

func (v *WorkflowExecutionCancelRequestedEventAttributes) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

type WorkflowExecutionCanceledEventAttributes struct {
	DecisionTaskCompletedEventId *int64
	Details                      []byte
}

func (v *WorkflowExecutionCanceledEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
}

func (v *WorkflowExecutionCanceledEventAttributes) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
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

func (v *WorkflowExecutionCompletedEventAttributes) GetResult() (o []byte) {
	if v != nil && v.Result != nil {
		return v.Result
	}
	return
}

func (v *WorkflowExecutionCompletedEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
}

type WorkflowExecutionConfiguration struct {
	TaskList                            *TaskList
	ExecutionStartToCloseTimeoutSeconds *int32
	TaskStartToCloseTimeoutSeconds      *int32
}

func (v *WorkflowExecutionConfiguration) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

func (v *WorkflowExecutionConfiguration) GetExecutionStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.ExecutionStartToCloseTimeoutSeconds != nil {
		return *v.ExecutionStartToCloseTimeoutSeconds
	}
	return
}

func (v *WorkflowExecutionConfiguration) GetTaskStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.TaskStartToCloseTimeoutSeconds != nil {
		return *v.TaskStartToCloseTimeoutSeconds
	}
	return
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

func (v *WorkflowExecutionContinuedAsNewEventAttributes) GetNewExecutionRunId() (o string) {
	if v != nil && v.NewExecutionRunId != nil {
		return *v.NewExecutionRunId
	}
	return
}

func (v *WorkflowExecutionContinuedAsNewEventAttributes) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

func (v *WorkflowExecutionContinuedAsNewEventAttributes) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

func (v *WorkflowExecutionContinuedAsNewEventAttributes) GetInput() (o []byte) {
	if v != nil && v.Input != nil {
		return v.Input
	}
	return
}

func (v *WorkflowExecutionContinuedAsNewEventAttributes) GetExecutionStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.ExecutionStartToCloseTimeoutSeconds != nil {
		return *v.ExecutionStartToCloseTimeoutSeconds
	}
	return
}

func (v *WorkflowExecutionContinuedAsNewEventAttributes) GetTaskStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.TaskStartToCloseTimeoutSeconds != nil {
		return *v.TaskStartToCloseTimeoutSeconds
	}
	return
}

func (v *WorkflowExecutionContinuedAsNewEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
}

func (v *WorkflowExecutionContinuedAsNewEventAttributes) GetBackoffStartIntervalInSeconds() (o int32) {
	if v != nil && v.BackoffStartIntervalInSeconds != nil {
		return *v.BackoffStartIntervalInSeconds
	}
	return
}

func (v *WorkflowExecutionContinuedAsNewEventAttributes) GetInitiator() (o ContinueAsNewInitiator) {
	if v != nil && v.Initiator != nil {
		return *v.Initiator
	}
	return
}

func (v *WorkflowExecutionContinuedAsNewEventAttributes) GetFailureReason() (o string) {
	if v != nil && v.FailureReason != nil {
		return *v.FailureReason
	}
	return
}

func (v *WorkflowExecutionContinuedAsNewEventAttributes) GetFailureDetails() (o []byte) {
	if v != nil && v.FailureDetails != nil {
		return v.FailureDetails
	}
	return
}

func (v *WorkflowExecutionContinuedAsNewEventAttributes) GetLastCompletionResult() (o []byte) {
	if v != nil && v.LastCompletionResult != nil {
		return v.LastCompletionResult
	}
	return
}

func (v *WorkflowExecutionContinuedAsNewEventAttributes) GetHeader() (o *Header) {
	if v != nil && v.Header != nil {
		return v.Header
	}
	return
}

func (v *WorkflowExecutionContinuedAsNewEventAttributes) GetMemo() (o *Memo) {
	if v != nil && v.Memo != nil {
		return v.Memo
	}
	return
}

func (v *WorkflowExecutionContinuedAsNewEventAttributes) GetSearchAttributes() (o *SearchAttributes) {
	if v != nil && v.SearchAttributes != nil {
		return v.SearchAttributes
	}
	return
}

type WorkflowExecutionFailedEventAttributes struct {
	Reason                       *string
	Details                      []byte
	DecisionTaskCompletedEventId *int64
}

func (v *WorkflowExecutionFailedEventAttributes) GetReason() (o string) {
	if v != nil && v.Reason != nil {
		return *v.Reason
	}
	return
}

func (v *WorkflowExecutionFailedEventAttributes) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *WorkflowExecutionFailedEventAttributes) GetDecisionTaskCompletedEventId() (o int64) {
	if v != nil && v.DecisionTaskCompletedEventId != nil {
		return *v.DecisionTaskCompletedEventId
	}
	return
}

type WorkflowExecutionFilter struct {
	WorkflowId *string
	RunId      *string
}

func (v *WorkflowExecutionFilter) GetWorkflowId() (o string) {
	if v != nil && v.WorkflowId != nil {
		return *v.WorkflowId
	}
	return
}

func (v *WorkflowExecutionFilter) GetRunId() (o string) {
	if v != nil && v.RunId != nil {
		return *v.RunId
	}
	return
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

func (v *WorkflowExecutionInfo) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
}

func (v *WorkflowExecutionInfo) GetType() (o *WorkflowType) {
	if v != nil && v.Type != nil {
		return v.Type
	}
	return
}

func (v *WorkflowExecutionInfo) GetStartTime() (o int64) {
	if v != nil && v.StartTime != nil {
		return *v.StartTime
	}
	return
}

func (v *WorkflowExecutionInfo) GetCloseTime() (o int64) {
	if v != nil && v.CloseTime != nil {
		return *v.CloseTime
	}
	return
}

func (v *WorkflowExecutionInfo) GetCloseStatus() (o WorkflowExecutionCloseStatus) {
	if v != nil && v.CloseStatus != nil {
		return *v.CloseStatus
	}
	return
}

func (v *WorkflowExecutionInfo) GetHistoryLength() (o int64) {
	if v != nil && v.HistoryLength != nil {
		return *v.HistoryLength
	}
	return
}

func (v *WorkflowExecutionInfo) GetParentDomainId() (o string) {
	if v != nil && v.ParentDomainId != nil {
		return *v.ParentDomainId
	}
	return
}

func (v *WorkflowExecutionInfo) GetParentExecution() (o *WorkflowExecution) {
	if v != nil && v.ParentExecution != nil {
		return v.ParentExecution
	}
	return
}

func (v *WorkflowExecutionInfo) GetExecutionTime() (o int64) {
	if v != nil && v.ExecutionTime != nil {
		return *v.ExecutionTime
	}
	return
}

func (v *WorkflowExecutionInfo) GetMemo() (o *Memo) {
	if v != nil && v.Memo != nil {
		return v.Memo
	}
	return
}

func (v *WorkflowExecutionInfo) GetSearchAttributes() (o *SearchAttributes) {
	if v != nil && v.SearchAttributes != nil {
		return v.SearchAttributes
	}
	return
}

func (v *WorkflowExecutionInfo) GetAutoResetPoints() (o *ResetPoints) {
	if v != nil && v.AutoResetPoints != nil {
		return v.AutoResetPoints
	}
	return
}

func (v *WorkflowExecutionInfo) GetTaskList() (o string) {
	if v != nil && v.TaskList != nil {
		return *v.TaskList
	}
	return
}

type WorkflowExecutionSignaledEventAttributes struct {
	SignalName *string
	Input      []byte
	Identity   *string
}

func (v *WorkflowExecutionSignaledEventAttributes) GetSignalName() (o string) {
	if v != nil && v.SignalName != nil {
		return *v.SignalName
	}
	return
}

func (v *WorkflowExecutionSignaledEventAttributes) GetInput() (o []byte) {
	if v != nil && v.Input != nil {
		return v.Input
	}
	return
}

func (v *WorkflowExecutionSignaledEventAttributes) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
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

func (v *WorkflowExecutionStartedEventAttributes) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetParentWorkflowDomain() (o string) {
	if v != nil && v.ParentWorkflowDomain != nil {
		return *v.ParentWorkflowDomain
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetParentWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.ParentWorkflowExecution != nil {
		return v.ParentWorkflowExecution
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetParentInitiatedEventId() (o int64) {
	if v != nil && v.ParentInitiatedEventId != nil {
		return *v.ParentInitiatedEventId
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetInput() (o []byte) {
	if v != nil && v.Input != nil {
		return v.Input
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetExecutionStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.ExecutionStartToCloseTimeoutSeconds != nil {
		return *v.ExecutionStartToCloseTimeoutSeconds
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetTaskStartToCloseTimeoutSeconds() (o int32) {
	if v != nil && v.TaskStartToCloseTimeoutSeconds != nil {
		return *v.TaskStartToCloseTimeoutSeconds
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetContinuedExecutionRunId() (o string) {
	if v != nil && v.ContinuedExecutionRunId != nil {
		return *v.ContinuedExecutionRunId
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetInitiator() (o ContinueAsNewInitiator) {
	if v != nil && v.Initiator != nil {
		return *v.Initiator
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetContinuedFailureReason() (o string) {
	if v != nil && v.ContinuedFailureReason != nil {
		return *v.ContinuedFailureReason
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetContinuedFailureDetails() (o []byte) {
	if v != nil && v.ContinuedFailureDetails != nil {
		return v.ContinuedFailureDetails
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetLastCompletionResult() (o []byte) {
	if v != nil && v.LastCompletionResult != nil {
		return v.LastCompletionResult
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetOriginalExecutionRunId() (o string) {
	if v != nil && v.OriginalExecutionRunId != nil {
		return *v.OriginalExecutionRunId
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetFirstExecutionRunId() (o string) {
	if v != nil && v.FirstExecutionRunId != nil {
		return *v.FirstExecutionRunId
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetRetryPolicy() (o *RetryPolicy) {
	if v != nil && v.RetryPolicy != nil {
		return v.RetryPolicy
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetAttempt() (o int32) {
	if v != nil && v.Attempt != nil {
		return *v.Attempt
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetExpirationTimestamp() (o int64) {
	if v != nil && v.ExpirationTimestamp != nil {
		return *v.ExpirationTimestamp
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetCronSchedule() (o string) {
	if v != nil && v.CronSchedule != nil {
		return *v.CronSchedule
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetFirstDecisionTaskBackoffSeconds() (o int32) {
	if v != nil && v.FirstDecisionTaskBackoffSeconds != nil {
		return *v.FirstDecisionTaskBackoffSeconds
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetMemo() (o *Memo) {
	if v != nil && v.Memo != nil {
		return v.Memo
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetSearchAttributes() (o *SearchAttributes) {
	if v != nil && v.SearchAttributes != nil {
		return v.SearchAttributes
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetPrevAutoResetPoints() (o *ResetPoints) {
	if v != nil && v.PrevAutoResetPoints != nil {
		return v.PrevAutoResetPoints
	}
	return
}

func (v *WorkflowExecutionStartedEventAttributes) GetHeader() (o *Header) {
	if v != nil && v.Header != nil {
		return v.Header
	}
	return
}

type WorkflowExecutionTerminatedEventAttributes struct {
	Reason   *string
	Details  []byte
	Identity *string
}

func (v *WorkflowExecutionTerminatedEventAttributes) GetReason() (o string) {
	if v != nil && v.Reason != nil {
		return *v.Reason
	}
	return
}

func (v *WorkflowExecutionTerminatedEventAttributes) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

func (v *WorkflowExecutionTerminatedEventAttributes) GetIdentity() (o string) {
	if v != nil && v.Identity != nil {
		return *v.Identity
	}
	return
}

type WorkflowExecutionTimedOutEventAttributes struct {
	TimeoutType *TimeoutType
}

func (v *WorkflowExecutionTimedOutEventAttributes) GetTimeoutType() (o TimeoutType) {
	if v != nil && v.TimeoutType != nil {
		return *v.TimeoutType
	}
	return
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

func (v *WorkflowQuery) GetQueryType() (o string) {
	if v != nil && v.QueryType != nil {
		return *v.QueryType
	}
	return
}

func (v *WorkflowQuery) GetQueryArgs() (o []byte) {
	if v != nil && v.QueryArgs != nil {
		return v.QueryArgs
	}
	return
}

type WorkflowQueryResult struct {
	ResultType   *QueryResultType
	Answer       []byte
	ErrorMessage *string
}

func (v *WorkflowQueryResult) GetResultType() (o QueryResultType) {
	if v != nil && v.ResultType != nil {
		return *v.ResultType
	}
	return
}

func (v *WorkflowQueryResult) GetAnswer() (o []byte) {
	if v != nil && v.Answer != nil {
		return v.Answer
	}
	return
}

func (v *WorkflowQueryResult) GetErrorMessage() (o string) {
	if v != nil && v.ErrorMessage != nil {
		return *v.ErrorMessage
	}
	return
}

type WorkflowType struct {
	Name *string
}

func (v *WorkflowType) GetName() (o string) {
	if v != nil && v.Name != nil {
		return *v.Name
	}
	return
}

type WorkflowTypeFilter struct {
	Name *string
}

func (v *WorkflowTypeFilter) GetName() (o string) {
	if v != nil && v.Name != nil {
		return *v.Name
	}
	return
}
