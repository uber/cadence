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

// DescribeMutableStateRequest is an internal type (TBD...)
type DescribeMutableStateRequest struct {
	DomainUUID string
	Execution  *WorkflowExecution
}

// GetDomainUUID is an internal getter (TBD...)
func (v *DescribeMutableStateRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// DescribeMutableStateResponse is an internal type (TBD...)
type DescribeMutableStateResponse struct {
	MutableStateInCache    string
	MutableStateInDatabase string
}

// HistoryDescribeWorkflowExecutionRequest is an internal type (TBD...)
type HistoryDescribeWorkflowExecutionRequest struct {
	DomainUUID string
	Request    *DescribeWorkflowExecutionRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryDescribeWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// DomainFilter is an internal type (TBD...)
type DomainFilter struct {
	DomainIDs    []string
	ReverseMatch bool
}

// GetDomainIDs is an internal getter (TBD...)
func (v *DomainFilter) GetDomainIDs() (o []string) {
	if v != nil && v.DomainIDs != nil {
		return v.DomainIDs
	}
	return
}

// GetReverseMatch is an internal getter (TBD...)
func (v *DomainFilter) GetReverseMatch() (o bool) {
	if v != nil {
		return v.ReverseMatch
	}
	return
}

// EventAlreadyStartedError is an internal type (TBD...)
type EventAlreadyStartedError struct {
	Message string
}

// FailoverMarkerToken is an internal type (TBD...)
type FailoverMarkerToken struct {
	ShardIDs       []int32
	FailoverMarker *FailoverMarkerAttributes
}

// GetShardIDs is an internal getter (TBD...)
func (v *FailoverMarkerToken) GetShardIDs() (o []int32) {
	if v != nil && v.ShardIDs != nil {
		return v.ShardIDs
	}
	return
}

// GetFailoverMarker is an internal getter (TBD...)
func (v *FailoverMarkerToken) GetFailoverMarker() (o *FailoverMarkerAttributes) {
	if v != nil && v.FailoverMarker != nil {
		return v.FailoverMarker
	}
	return
}

// GetMutableStateRequest is an internal type (TBD...)
type GetMutableStateRequest struct {
	DomainUUID          string
	Execution           *WorkflowExecution
	ExpectedNextEventID int64
	CurrentBranchToken  []byte
}

// GetDomainUUID is an internal getter (TBD...)
func (v *GetMutableStateRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetExpectedNextEventID is an internal getter (TBD...)
func (v *GetMutableStateRequest) GetExpectedNextEventID() (o int64) {
	if v != nil {
		return v.ExpectedNextEventID
	}
	return
}

// GetMutableStateResponse is an internal type (TBD...)
type GetMutableStateResponse struct {
	Execution                            *WorkflowExecution
	WorkflowType                         *WorkflowType
	NextEventID                          int64
	PreviousStartedEventID               *int64
	LastFirstEventID                     int64
	TaskList                             *TaskList
	StickyTaskList                       *TaskList
	ClientLibraryVersion                 string
	ClientFeatureVersion                 string
	ClientImpl                           string
	IsWorkflowRunning                    bool
	StickyTaskListScheduleToStartTimeout *int32
	EventStoreVersion                    int32
	CurrentBranchToken                   []byte
	WorkflowState                        *int32
	WorkflowCloseState                   *int32
	VersionHistories                     *VersionHistories
	IsStickyTaskListEnabled              bool
}

// GetNextEventID is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetNextEventID() (o int64) {
	if v != nil {
		return v.NextEventID
	}
	return
}

// GetPreviousStartedEventID is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetPreviousStartedEventID() (o int64) {
	if v != nil && v.PreviousStartedEventID != nil {
		return *v.PreviousStartedEventID
	}
	return
}

// GetStickyTaskList is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetStickyTaskList() (o *TaskList) {
	if v != nil && v.StickyTaskList != nil {
		return v.StickyTaskList
	}
	return
}

// GetClientFeatureVersion is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetClientFeatureVersion() (o string) {
	if v != nil {
		return v.ClientFeatureVersion
	}
	return
}

// GetClientImpl is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetClientImpl() (o string) {
	if v != nil {
		return v.ClientImpl
	}
	return
}

// GetIsWorkflowRunning is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetIsWorkflowRunning() (o bool) {
	if v != nil {
		return v.IsWorkflowRunning
	}
	return
}

// GetStickyTaskListScheduleToStartTimeout is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetStickyTaskListScheduleToStartTimeout() (o int32) {
	if v != nil && v.StickyTaskListScheduleToStartTimeout != nil {
		return *v.StickyTaskListScheduleToStartTimeout
	}
	return
}

// GetCurrentBranchToken is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetCurrentBranchToken() (o []byte) {
	if v != nil && v.CurrentBranchToken != nil {
		return v.CurrentBranchToken
	}
	return
}

// GetWorkflowCloseState is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetWorkflowCloseState() (o int32) {
	if v != nil && v.WorkflowCloseState != nil {
		return *v.WorkflowCloseState
	}
	return
}

// GetVersionHistories is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetVersionHistories() (o *VersionHistories) {
	if v != nil && v.VersionHistories != nil {
		return v.VersionHistories
	}
	return
}

// GetIsStickyTaskListEnabled is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetIsStickyTaskListEnabled() (o bool) {
	if v != nil {
		return v.IsStickyTaskListEnabled
	}
	return
}

// NotifyFailoverMarkersRequest is an internal type (TBD...)
type NotifyFailoverMarkersRequest struct {
	FailoverMarkerTokens []*FailoverMarkerToken
}

// GetFailoverMarkerTokens is an internal getter (TBD...)
func (v *NotifyFailoverMarkersRequest) GetFailoverMarkerTokens() (o []*FailoverMarkerToken) {
	if v != nil && v.FailoverMarkerTokens != nil {
		return v.FailoverMarkerTokens
	}
	return
}

// ParentExecutionInfo is an internal type (TBD...)
type ParentExecutionInfo struct {
	DomainUUID  string
	Domain      string
	Execution   *WorkflowExecution
	InitiatedID int64
}

// GetDomain is an internal getter (TBD...)
func (v *ParentExecutionInfo) GetDomain() (o string) {
	if v != nil {
		return v.Domain
	}
	return
}

// GetExecution is an internal getter (TBD...)
func (v *ParentExecutionInfo) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
}

// PollMutableStateRequest is an internal type (TBD...)
type PollMutableStateRequest struct {
	DomainUUID          string
	Execution           *WorkflowExecution
	ExpectedNextEventID int64
	CurrentBranchToken  []byte
}

// GetDomainUUID is an internal getter (TBD...)
func (v *PollMutableStateRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// PollMutableStateResponse is an internal type (TBD...)
type PollMutableStateResponse struct {
	Execution                            *WorkflowExecution
	WorkflowType                         *WorkflowType
	NextEventID                          int64
	PreviousStartedEventID               *int64
	LastFirstEventID                     int64
	TaskList                             *TaskList
	StickyTaskList                       *TaskList
	ClientLibraryVersion                 string
	ClientFeatureVersion                 string
	ClientImpl                           string
	StickyTaskListScheduleToStartTimeout *int32
	CurrentBranchToken                   []byte
	VersionHistories                     *VersionHistories
	WorkflowState                        *int32
	WorkflowCloseState                   *int32
}

// GetNextEventID is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetNextEventID() (o int64) {
	if v != nil {
		return v.NextEventID
	}
	return
}

// GetLastFirstEventID is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetLastFirstEventID() (o int64) {
	if v != nil {
		return v.LastFirstEventID
	}
	return
}

// GetWorkflowCloseState is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetWorkflowCloseState() (o int32) {
	if v != nil && v.WorkflowCloseState != nil {
		return *v.WorkflowCloseState
	}
	return
}

// ProcessingQueueState is an internal type (TBD...)
type ProcessingQueueState struct {
	Level        *int32
	AckLevel     *int64
	MaxLevel     *int64
	DomainFilter *DomainFilter
}

// GetLevel is an internal getter (TBD...)
func (v *ProcessingQueueState) GetLevel() (o int32) {
	if v != nil && v.Level != nil {
		return *v.Level
	}
	return
}

// GetAckLevel is an internal getter (TBD...)
func (v *ProcessingQueueState) GetAckLevel() (o int64) {
	if v != nil && v.AckLevel != nil {
		return *v.AckLevel
	}
	return
}

// GetMaxLevel is an internal getter (TBD...)
func (v *ProcessingQueueState) GetMaxLevel() (o int64) {
	if v != nil && v.MaxLevel != nil {
		return *v.MaxLevel
	}
	return
}

// GetDomainFilter is an internal getter (TBD...)
func (v *ProcessingQueueState) GetDomainFilter() (o *DomainFilter) {
	if v != nil && v.DomainFilter != nil {
		return v.DomainFilter
	}
	return
}

// ProcessingQueueStates is an internal type (TBD...)
type ProcessingQueueStates struct {
	StatesByCluster map[string][]*ProcessingQueueState
}

// HistoryQueryWorkflowRequest is an internal type (TBD...)
type HistoryQueryWorkflowRequest struct {
	DomainUUID string
	Request    *QueryWorkflowRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryQueryWorkflowRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetRequest is an internal getter (TBD...)
func (v *HistoryQueryWorkflowRequest) GetRequest() (o *QueryWorkflowRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// HistoryQueryWorkflowResponse is an internal type (TBD...)
type HistoryQueryWorkflowResponse struct {
	Response *QueryWorkflowResponse
}

// GetResponse is an internal getter (TBD...)
func (v *HistoryQueryWorkflowResponse) GetResponse() (o *QueryWorkflowResponse) {
	if v != nil && v.Response != nil {
		return v.Response
	}
	return
}

// HistoryReapplyEventsRequest is an internal type (TBD...)
type HistoryReapplyEventsRequest struct {
	DomainUUID string
	Request    *ReapplyEventsRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryReapplyEventsRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetRequest is an internal getter (TBD...)
func (v *HistoryReapplyEventsRequest) GetRequest() (o *ReapplyEventsRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// HistoryRecordActivityTaskHeartbeatRequest is an internal type (TBD...)
type HistoryRecordActivityTaskHeartbeatRequest struct {
	DomainUUID       string
	HeartbeatRequest *RecordActivityTaskHeartbeatRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRecordActivityTaskHeartbeatRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// RecordActivityTaskStartedRequest is an internal type (TBD...)
type RecordActivityTaskStartedRequest struct {
	DomainUUID        string
	WorkflowExecution *WorkflowExecution
	ScheduleID        int64
	TaskID            int64
	RequestID         string
	PollRequest       *PollForActivityTaskRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *RecordActivityTaskStartedRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetScheduleID is an internal getter (TBD...)
func (v *RecordActivityTaskStartedRequest) GetScheduleID() (o int64) {
	if v != nil {
		return v.ScheduleID
	}
	return
}

// GetTaskID is an internal getter (TBD...)
func (v *RecordActivityTaskStartedRequest) GetTaskID() (o int64) {
	if v != nil {
		return v.TaskID
	}
	return
}

// GetRequestID is an internal getter (TBD...)
func (v *RecordActivityTaskStartedRequest) GetRequestID() (o string) {
	if v != nil {
		return v.RequestID
	}
	return
}

// RecordActivityTaskStartedResponse is an internal type (TBD...)
type RecordActivityTaskStartedResponse struct {
	ScheduledEvent                  *HistoryEvent
	StartedTimestamp                *int64
	Attempt                         int64
	ScheduledTimestampOfThisAttempt *int64
	HeartbeatDetails                []byte
	WorkflowType                    *WorkflowType
	WorkflowDomain                  string
}

// GetAttempt is an internal getter (TBD...)
func (v *RecordActivityTaskStartedResponse) GetAttempt() (o int64) {
	if v != nil {
		return v.Attempt
	}
	return
}

// GetScheduledTimestampOfThisAttempt is an internal getter (TBD...)
func (v *RecordActivityTaskStartedResponse) GetScheduledTimestampOfThisAttempt() (o int64) {
	if v != nil && v.ScheduledTimestampOfThisAttempt != nil {
		return *v.ScheduledTimestampOfThisAttempt
	}
	return
}

// RecordChildExecutionCompletedRequest is an internal type (TBD...)
type RecordChildExecutionCompletedRequest struct {
	DomainUUID         string
	WorkflowExecution  *WorkflowExecution
	InitiatedID        int64
	CompletedExecution *WorkflowExecution
	CompletionEvent    *HistoryEvent
}

// GetDomainUUID is an internal getter (TBD...)
func (v *RecordChildExecutionCompletedRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// RecordDecisionTaskStartedRequest is an internal type (TBD...)
type RecordDecisionTaskStartedRequest struct {
	DomainUUID        string
	WorkflowExecution *WorkflowExecution
	ScheduleID        int64
	TaskID            int64
	RequestID         string
	PollRequest       *PollForDecisionTaskRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetScheduleID is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedRequest) GetScheduleID() (o int64) {
	if v != nil {
		return v.ScheduleID
	}
	return
}

// GetRequestID is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedRequest) GetRequestID() (o string) {
	if v != nil {
		return v.RequestID
	}
	return
}

// RecordDecisionTaskStartedResponse is an internal type (TBD...)
type RecordDecisionTaskStartedResponse struct {
	WorkflowType              *WorkflowType
	PreviousStartedEventID    *int64
	ScheduledEventID          int64
	StartedEventID            int64
	NextEventID               int64
	Attempt                   int64
	StickyExecutionEnabled    bool
	DecisionInfo              *TransientDecisionInfo
	WorkflowExecutionTaskList *TaskList
	EventStoreVersion         int32
	BranchToken               []byte
	ScheduledTimestamp        *int64
	StartedTimestamp          *int64
	Queries                   map[string]*WorkflowQuery
}

// GetPreviousStartedEventID is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetPreviousStartedEventID() (o int64) {
	if v != nil && v.PreviousStartedEventID != nil {
		return *v.PreviousStartedEventID
	}
	return
}

// GetScheduledEventID is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetScheduledEventID() (o int64) {
	if v != nil {
		return v.ScheduledEventID
	}
	return
}

// GetAttempt is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetAttempt() (o int64) {
	if v != nil {
		return v.Attempt
	}
	return
}

// HistoryRefreshWorkflowTasksRequest is an internal type (TBD...)
type HistoryRefreshWorkflowTasksRequest struct {
	DomainUIID string
	Request    *RefreshWorkflowTasksRequest
}

// GetRequest is an internal getter (TBD...)
func (v *HistoryRefreshWorkflowTasksRequest) GetRequest() (o *RefreshWorkflowTasksRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// RemoveSignalMutableStateRequest is an internal type (TBD...)
type RemoveSignalMutableStateRequest struct {
	DomainUUID        string
	WorkflowExecution *WorkflowExecution
	RequestID         string
}

// GetDomainUUID is an internal getter (TBD...)
func (v *RemoveSignalMutableStateRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetRequestID is an internal getter (TBD...)
func (v *RemoveSignalMutableStateRequest) GetRequestID() (o string) {
	if v != nil {
		return v.RequestID
	}
	return
}

// ReplicateEventsV2Request is an internal type (TBD...)
type ReplicateEventsV2Request struct {
	DomainUUID          string
	WorkflowExecution   *WorkflowExecution
	VersionHistoryItems []*VersionHistoryItem
	Events              *DataBlob
	NewRunEvents        *DataBlob
}

// GetDomainUUID is an internal getter (TBD...)
func (v *ReplicateEventsV2Request) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// HistoryRequestCancelWorkflowExecutionRequest is an internal type (TBD...)
type HistoryRequestCancelWorkflowExecutionRequest struct {
	DomainUUID                string
	CancelRequest             *RequestCancelWorkflowExecutionRequest
	ExternalInitiatedEventID  *int64
	ExternalWorkflowExecution *WorkflowExecution
	ChildWorkflowOnly         bool
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRequestCancelWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetCancelRequest is an internal getter (TBD...)
func (v *HistoryRequestCancelWorkflowExecutionRequest) GetCancelRequest() (o *RequestCancelWorkflowExecutionRequest) {
	if v != nil && v.CancelRequest != nil {
		return v.CancelRequest
	}
	return
}

// GetExternalWorkflowExecution is an internal getter (TBD...)
func (v *HistoryRequestCancelWorkflowExecutionRequest) GetExternalWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.ExternalWorkflowExecution != nil {
		return v.ExternalWorkflowExecution
	}
	return
}

// GetChildWorkflowOnly is an internal getter (TBD...)
func (v *HistoryRequestCancelWorkflowExecutionRequest) GetChildWorkflowOnly() (o bool) {
	if v != nil {
		return v.ChildWorkflowOnly
	}
	return
}

// HistoryResetStickyTaskListRequest is an internal type (TBD...)
type HistoryResetStickyTaskListRequest struct {
	DomainUUID string
	Execution  *WorkflowExecution
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryResetStickyTaskListRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// HistoryResetStickyTaskListResponse is an internal type (TBD...)
type HistoryResetStickyTaskListResponse struct {
}

// HistoryResetWorkflowExecutionRequest is an internal type (TBD...)
type HistoryResetWorkflowExecutionRequest struct {
	DomainUUID   string
	ResetRequest *ResetWorkflowExecutionRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryResetWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// HistoryRespondActivityTaskCanceledRequest is an internal type (TBD...)
type HistoryRespondActivityTaskCanceledRequest struct {
	DomainUUID    string
	CancelRequest *RespondActivityTaskCanceledRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRespondActivityTaskCanceledRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// HistoryRespondActivityTaskCompletedRequest is an internal type (TBD...)
type HistoryRespondActivityTaskCompletedRequest struct {
	DomainUUID      string
	CompleteRequest *RespondActivityTaskCompletedRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRespondActivityTaskCompletedRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// HistoryRespondActivityTaskFailedRequest is an internal type (TBD...)
type HistoryRespondActivityTaskFailedRequest struct {
	DomainUUID    string
	FailedRequest *RespondActivityTaskFailedRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRespondActivityTaskFailedRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// HistoryRespondDecisionTaskCompletedRequest is an internal type (TBD...)
type HistoryRespondDecisionTaskCompletedRequest struct {
	DomainUUID      string
	CompleteRequest *RespondDecisionTaskCompletedRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRespondDecisionTaskCompletedRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetCompleteRequest is an internal getter (TBD...)
func (v *HistoryRespondDecisionTaskCompletedRequest) GetCompleteRequest() (o *RespondDecisionTaskCompletedRequest) {
	if v != nil && v.CompleteRequest != nil {
		return v.CompleteRequest
	}
	return
}

// HistoryRespondDecisionTaskCompletedResponse is an internal type (TBD...)
type HistoryRespondDecisionTaskCompletedResponse struct {
	StartedResponse             *RecordDecisionTaskStartedResponse
	ActivitiesToDispatchLocally map[string]*ActivityLocalDispatchInfo
}

// HistoryRespondDecisionTaskFailedRequest is an internal type (TBD...)
type HistoryRespondDecisionTaskFailedRequest struct {
	DomainUUID    string
	FailedRequest *RespondDecisionTaskFailedRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRespondDecisionTaskFailedRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// ScheduleDecisionTaskRequest is an internal type (TBD...)
type ScheduleDecisionTaskRequest struct {
	DomainUUID        string
	WorkflowExecution *WorkflowExecution
	IsFirstDecision   bool
}

// GetDomainUUID is an internal getter (TBD...)
func (v *ScheduleDecisionTaskRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// ShardOwnershipLostError is an internal type (TBD...)
type ShardOwnershipLostError struct {
	Message string
	Owner   string
}

// GetOwner is an internal getter (TBD...)
func (v *ShardOwnershipLostError) GetOwner() (o string) {
	if v != nil {
		return v.Owner
	}
	return
}

// HistorySignalWithStartWorkflowExecutionRequest is an internal type (TBD...)
type HistorySignalWithStartWorkflowExecutionRequest struct {
	DomainUUID             string
	SignalWithStartRequest *SignalWithStartWorkflowExecutionRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistorySignalWithStartWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// HistorySignalWorkflowExecutionRequest is an internal type (TBD...)
type HistorySignalWorkflowExecutionRequest struct {
	DomainUUID                string
	SignalRequest             *SignalWorkflowExecutionRequest
	ExternalWorkflowExecution *WorkflowExecution
	ChildWorkflowOnly         bool
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistorySignalWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetChildWorkflowOnly is an internal getter (TBD...)
func (v *HistorySignalWorkflowExecutionRequest) GetChildWorkflowOnly() (o bool) {
	if v != nil {
		return v.ChildWorkflowOnly
	}
	return
}

// HistoryStartWorkflowExecutionRequest is an internal type (TBD...)
type HistoryStartWorkflowExecutionRequest struct {
	DomainUUID                      string
	StartRequest                    *StartWorkflowExecutionRequest
	ParentExecutionInfo             *ParentExecutionInfo
	Attempt                         int32
	ExpirationTimestamp             *int64
	ContinueAsNewInitiator          *ContinueAsNewInitiator
	ContinuedFailureReason          *string
	ContinuedFailureDetails         []byte
	LastCompletionResult            []byte
	FirstDecisionTaskBackoffSeconds *int32
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryStartWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetAttempt is an internal getter (TBD...)
func (v *HistoryStartWorkflowExecutionRequest) GetAttempt() (o int32) {
	if v != nil {
		return v.Attempt
	}
	return
}

// GetExpirationTimestamp is an internal getter (TBD...)
func (v *HistoryStartWorkflowExecutionRequest) GetExpirationTimestamp() (o int64) {
	if v != nil && v.ExpirationTimestamp != nil {
		return *v.ExpirationTimestamp
	}
	return
}

// GetFirstDecisionTaskBackoffSeconds is an internal getter (TBD...)
func (v *HistoryStartWorkflowExecutionRequest) GetFirstDecisionTaskBackoffSeconds() (o int32) {
	if v != nil && v.FirstDecisionTaskBackoffSeconds != nil {
		return *v.FirstDecisionTaskBackoffSeconds
	}
	return
}

// SyncActivityRequest is an internal type (TBD...)
type SyncActivityRequest struct {
	DomainID           string
	WorkflowID         string
	RunID              string
	Version            int64
	ScheduledID        int64
	ScheduledTime      *int64
	StartedID          int64
	StartedTime        *int64
	LastHeartbeatTime  *int64
	Details            []byte
	Attempt            int32
	LastFailureReason  *string
	LastWorkerIdentity string
	LastFailureDetails []byte
	VersionHistory     *VersionHistory
}

// GetDomainID is an internal getter (TBD...)
func (v *SyncActivityRequest) GetDomainID() (o string) {
	if v != nil {
		return v.DomainID
	}
	return
}

// GetWorkflowID is an internal getter (TBD...)
func (v *SyncActivityRequest) GetWorkflowID() (o string) {
	if v != nil {
		return v.WorkflowID
	}
	return
}

// GetRunID is an internal getter (TBD...)
func (v *SyncActivityRequest) GetRunID() (o string) {
	if v != nil {
		return v.RunID
	}
	return
}

// GetVersion is an internal getter (TBD...)
func (v *SyncActivityRequest) GetVersion() (o int64) {
	if v != nil {
		return v.Version
	}
	return
}

// GetScheduledID is an internal getter (TBD...)
func (v *SyncActivityRequest) GetScheduledID() (o int64) {
	if v != nil {
		return v.ScheduledID
	}
	return
}

// GetScheduledTime is an internal getter (TBD...)
func (v *SyncActivityRequest) GetScheduledTime() (o int64) {
	if v != nil && v.ScheduledTime != nil {
		return *v.ScheduledTime
	}
	return
}

// GetStartedID is an internal getter (TBD...)
func (v *SyncActivityRequest) GetStartedID() (o int64) {
	if v != nil {
		return v.StartedID
	}
	return
}

// GetStartedTime is an internal getter (TBD...)
func (v *SyncActivityRequest) GetStartedTime() (o int64) {
	if v != nil && v.StartedTime != nil {
		return *v.StartedTime
	}
	return
}

// GetLastHeartbeatTime is an internal getter (TBD...)
func (v *SyncActivityRequest) GetLastHeartbeatTime() (o int64) {
	if v != nil && v.LastHeartbeatTime != nil {
		return *v.LastHeartbeatTime
	}
	return
}

// GetDetails is an internal getter (TBD...)
func (v *SyncActivityRequest) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}

// GetAttempt is an internal getter (TBD...)
func (v *SyncActivityRequest) GetAttempt() (o int32) {
	if v != nil {
		return v.Attempt
	}
	return
}

// GetLastFailureReason is an internal getter (TBD...)
func (v *SyncActivityRequest) GetLastFailureReason() (o string) {
	if v != nil && v.LastFailureReason != nil {
		return *v.LastFailureReason
	}
	return
}

// GetLastWorkerIdentity is an internal getter (TBD...)
func (v *SyncActivityRequest) GetLastWorkerIdentity() (o string) {
	if v != nil {
		return v.LastWorkerIdentity
	}
	return
}

// GetLastFailureDetails is an internal getter (TBD...)
func (v *SyncActivityRequest) GetLastFailureDetails() (o []byte) {
	if v != nil && v.LastFailureDetails != nil {
		return v.LastFailureDetails
	}
	return
}

// GetVersionHistory is an internal getter (TBD...)
func (v *SyncActivityRequest) GetVersionHistory() (o *VersionHistory) {
	if v != nil && v.VersionHistory != nil {
		return v.VersionHistory
	}
	return
}

// SyncShardStatusRequest is an internal type (TBD...)
type SyncShardStatusRequest struct {
	SourceCluster string
	ShardID       int64
	Timestamp     *int64
}

// GetSourceCluster is an internal getter (TBD...)
func (v *SyncShardStatusRequest) GetSourceCluster() (o string) {
	if v != nil {
		return v.SourceCluster
	}
	return
}

// GetShardID is an internal getter (TBD...)
func (v *SyncShardStatusRequest) GetShardID() (o int64) {
	if v != nil {
		return v.ShardID
	}
	return
}

// GetTimestamp is an internal getter (TBD...)
func (v *SyncShardStatusRequest) GetTimestamp() (o int64) {
	if v != nil && v.Timestamp != nil {
		return *v.Timestamp
	}
	return
}

// HistoryTerminateWorkflowExecutionRequest is an internal type (TBD...)
type HistoryTerminateWorkflowExecutionRequest struct {
	DomainUUID                string
	TerminateRequest          *TerminateWorkflowExecutionRequest
	ExternalWorkflowExecution *WorkflowExecution
	ChildWorkflowOnly         bool
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryTerminateWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetTerminateRequest is an internal getter (TBD...)
func (v *HistoryTerminateWorkflowExecutionRequest) GetTerminateRequest() (o *TerminateWorkflowExecutionRequest) {
	if v != nil && v.TerminateRequest != nil {
		return v.TerminateRequest
	}
	return
}

// GetExternalWorkflowExecution is an internal getter (TBD...)
func (v *HistoryTerminateWorkflowExecutionRequest) GetExternalWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.ExternalWorkflowExecution != nil {
		return v.ExternalWorkflowExecution
	}
	return
}

// GetChildWorkflowOnly is an internal getter (TBD...)
func (v *HistoryTerminateWorkflowExecutionRequest) GetChildWorkflowOnly() (o bool) {
	if v != nil {
		return v.ChildWorkflowOnly
	}
	return
}

// GetFailoverInfoRequest is an internal type (TBD...)
type GetFailoverInfoRequest struct {
	DomainID string
}

// GetDomainID is an internal getter (TBD...)
func (v *GetFailoverInfoRequest) GetDomainID() (o string) {
	if v != nil {
		return v.DomainID
	}
	return
}

// GetFailoverInfoResponse is an internal type (TBD...)
type GetFailoverInfoResponse struct {
	CompletedShardCount int32
	PendingShards       []int32
}

// GetCompletedShardCount is an internal getter (TBD...)
func (v *GetFailoverInfoResponse) GetCompletedShardCount() (o int32) {
	if v != nil {
		return v.CompletedShardCount
	}
	return
}

// GetPendingShards is an internal getter (TBD...)
func (v *GetFailoverInfoResponse) GetPendingShards() (o []int32) {
	if v != nil {
		return v.PendingShards
	}
	return
}
