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
	DomainUUID string             `json:"domainUUID,omitempty"`
	Execution  *WorkflowExecution `json:"execution,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *DescribeMutableStateRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetExecution is an internal getter (TBD...)
func (v *DescribeMutableStateRequest) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
}

// DescribeMutableStateResponse is an internal type (TBD...)
type DescribeMutableStateResponse struct {
	MutableStateInCache    string `json:"mutableStateInCache,omitempty"`
	MutableStateInDatabase string `json:"mutableStateInDatabase,omitempty"`
}

// GetMutableStateInCache is an internal getter (TBD...)
func (v *DescribeMutableStateResponse) GetMutableStateInCache() (o string) {
	if v != nil {
		return v.MutableStateInCache
	}
	return
}

// GetMutableStateInDatabase is an internal getter (TBD...)
func (v *DescribeMutableStateResponse) GetMutableStateInDatabase() (o string) {
	if v != nil {
		return v.MutableStateInDatabase
	}
	return
}

// HistoryDescribeWorkflowExecutionRequest is an internal type (TBD...)
type HistoryDescribeWorkflowExecutionRequest struct {
	DomainUUID string                            `json:"domainUUID,omitempty"`
	Request    *DescribeWorkflowExecutionRequest `json:"request,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryDescribeWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetRequest is an internal getter (TBD...)
func (v *HistoryDescribeWorkflowExecutionRequest) GetRequest() (o *DescribeWorkflowExecutionRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// DomainFilter is an internal type (TBD...)
type DomainFilter struct {
	DomainIDs    []string `json:"domainIDs,omitempty"`
	ReverseMatch bool     `json:"reverseMatch,omitempty"`
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
	Message string `json:"message,required"`
}

// GetMessage is an internal getter (TBD...)
func (v *EventAlreadyStartedError) GetMessage() (o string) {
	if v != nil {
		return v.Message
	}
	return
}

// FailoverMarkerToken is an internal type (TBD...)
type FailoverMarkerToken struct {
	ShardIDs       []int32                   `json:"shardIDs,omitempty"`
	FailoverMarker *FailoverMarkerAttributes `json:"failoverMarker,omitempty"`
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
	DomainUUID          string             `json:"domainUUID,omitempty"`
	Execution           *WorkflowExecution `json:"execution,omitempty"`
	ExpectedNextEventID int64              `json:"expectedNextEventId,omitempty"`
	CurrentBranchToken  []byte             `json:"currentBranchToken,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *GetMutableStateRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetExecution is an internal getter (TBD...)
func (v *GetMutableStateRequest) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
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

// GetCurrentBranchToken is an internal getter (TBD...)
func (v *GetMutableStateRequest) GetCurrentBranchToken() (o []byte) {
	if v != nil && v.CurrentBranchToken != nil {
		return v.CurrentBranchToken
	}
	return
}

// GetMutableStateResponse is an internal type (TBD...)
type GetMutableStateResponse struct {
	Execution                            *WorkflowExecution `json:"execution,omitempty"`
	WorkflowType                         *WorkflowType      `json:"workflowType,omitempty"`
	NextEventID                          int64              `json:"NextEventId,omitempty"`
	PreviousStartedEventID               *int64             `json:"PreviousStartedEventId,omitempty"`
	LastFirstEventID                     int64              `json:"LastFirstEventId,omitempty"`
	TaskList                             *TaskList          `json:"taskList,omitempty"`
	StickyTaskList                       *TaskList          `json:"stickyTaskList,omitempty"`
	ClientLibraryVersion                 string             `json:"clientLibraryVersion,omitempty"`
	ClientFeatureVersion                 string             `json:"clientFeatureVersion,omitempty"`
	ClientImpl                           string             `json:"clientImpl,omitempty"`
	IsWorkflowRunning                    bool               `json:"isWorkflowRunning,omitempty"`
	StickyTaskListScheduleToStartTimeout *int32             `json:"stickyTaskListScheduleToStartTimeout,omitempty"`
	EventStoreVersion                    int32              `json:"eventStoreVersion,omitempty"`
	CurrentBranchToken                   []byte             `json:"currentBranchToken,omitempty"`
	WorkflowState                        *int32             `json:"workflowState,omitempty"`
	WorkflowCloseState                   *int32             `json:"workflowCloseState,omitempty"`
	VersionHistories                     *VersionHistories  `json:"versionHistories,omitempty"`
	IsStickyTaskListEnabled              bool               `json:"isStickyTaskListEnabled,omitempty"`
}

// GetExecution is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
}

// GetWorkflowType is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
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

// GetLastFirstEventID is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetLastFirstEventID() (o int64) {
	if v != nil {
		return v.LastFirstEventID
	}
	return
}

// GetTaskList is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
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

// GetClientLibraryVersion is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetClientLibraryVersion() (o string) {
	if v != nil {
		return v.ClientLibraryVersion
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

// GetEventStoreVersion is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetEventStoreVersion() (o int32) {
	if v != nil {
		return v.EventStoreVersion
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

// GetWorkflowState is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetWorkflowState() (o int32) {
	if v != nil && v.WorkflowState != nil {
		return *v.WorkflowState
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
	FailoverMarkerTokens []*FailoverMarkerToken `json:"failoverMarkerTokens,omitempty"`
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
	DomainUUID  string             `json:"domainUUID,omitempty"`
	Domain      string             `json:"domain,omitempty"`
	Execution   *WorkflowExecution `json:"execution,omitempty"`
	InitiatedID int64              `json:"initiatedId,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *ParentExecutionInfo) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
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

// GetInitiatedID is an internal getter (TBD...)
func (v *ParentExecutionInfo) GetInitiatedID() (o int64) {
	if v != nil {
		return v.InitiatedID
	}
	return
}

// PollMutableStateRequest is an internal type (TBD...)
type PollMutableStateRequest struct {
	DomainUUID          string             `json:"domainUUID,omitempty"`
	Execution           *WorkflowExecution `json:"execution,omitempty"`
	ExpectedNextEventID int64              `json:"expectedNextEventId,omitempty"`
	CurrentBranchToken  []byte             `json:"currentBranchToken,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *PollMutableStateRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetExecution is an internal getter (TBD...)
func (v *PollMutableStateRequest) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
}

// GetExpectedNextEventID is an internal getter (TBD...)
func (v *PollMutableStateRequest) GetExpectedNextEventID() (o int64) {
	if v != nil {
		return v.ExpectedNextEventID
	}
	return
}

// GetCurrentBranchToken is an internal getter (TBD...)
func (v *PollMutableStateRequest) GetCurrentBranchToken() (o []byte) {
	if v != nil && v.CurrentBranchToken != nil {
		return v.CurrentBranchToken
	}
	return
}

// PollMutableStateResponse is an internal type (TBD...)
type PollMutableStateResponse struct {
	Execution                            *WorkflowExecution `json:"execution,omitempty"`
	WorkflowType                         *WorkflowType      `json:"workflowType,omitempty"`
	NextEventID                          int64              `json:"NextEventId,omitempty"`
	PreviousStartedEventID               *int64             `json:"PreviousStartedEventId,omitempty"`
	LastFirstEventID                     int64              `json:"LastFirstEventId,omitempty"`
	TaskList                             *TaskList          `json:"taskList,omitempty"`
	StickyTaskList                       *TaskList          `json:"stickyTaskList,omitempty"`
	ClientLibraryVersion                 string             `json:"clientLibraryVersion,omitempty"`
	ClientFeatureVersion                 string             `json:"clientFeatureVersion,omitempty"`
	ClientImpl                           string             `json:"clientImpl,omitempty"`
	StickyTaskListScheduleToStartTimeout *int32             `json:"stickyTaskListScheduleToStartTimeout,omitempty"`
	CurrentBranchToken                   []byte             `json:"currentBranchToken,omitempty"`
	VersionHistories                     *VersionHistories  `json:"versionHistories,omitempty"`
	WorkflowState                        *int32             `json:"workflowState,omitempty"`
	WorkflowCloseState                   *int32             `json:"workflowCloseState,omitempty"`
}

// GetExecution is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
}

// GetWorkflowType is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

// GetNextEventID is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetNextEventID() (o int64) {
	if v != nil {
		return v.NextEventID
	}
	return
}

// GetPreviousStartedEventID is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetPreviousStartedEventID() (o int64) {
	if v != nil && v.PreviousStartedEventID != nil {
		return *v.PreviousStartedEventID
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

// GetTaskList is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

// GetStickyTaskList is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetStickyTaskList() (o *TaskList) {
	if v != nil && v.StickyTaskList != nil {
		return v.StickyTaskList
	}
	return
}

// GetClientLibraryVersion is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetClientLibraryVersion() (o string) {
	if v != nil {
		return v.ClientLibraryVersion
	}
	return
}

// GetClientFeatureVersion is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetClientFeatureVersion() (o string) {
	if v != nil {
		return v.ClientFeatureVersion
	}
	return
}

// GetClientImpl is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetClientImpl() (o string) {
	if v != nil {
		return v.ClientImpl
	}
	return
}

// GetStickyTaskListScheduleToStartTimeout is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetStickyTaskListScheduleToStartTimeout() (o int32) {
	if v != nil && v.StickyTaskListScheduleToStartTimeout != nil {
		return *v.StickyTaskListScheduleToStartTimeout
	}
	return
}

// GetCurrentBranchToken is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetCurrentBranchToken() (o []byte) {
	if v != nil && v.CurrentBranchToken != nil {
		return v.CurrentBranchToken
	}
	return
}

// GetVersionHistories is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetVersionHistories() (o *VersionHistories) {
	if v != nil && v.VersionHistories != nil {
		return v.VersionHistories
	}
	return
}

// GetWorkflowState is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetWorkflowState() (o int32) {
	if v != nil && v.WorkflowState != nil {
		return *v.WorkflowState
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
	Level        *int32        `json:"level,omitempty"`
	AckLevel     *int64        `json:"ackLevel,omitempty"`
	MaxLevel     *int64        `json:"maxLevel,omitempty"`
	DomainFilter *DomainFilter `json:"domainFilter,omitempty"`
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
	StatesByCluster map[string][]*ProcessingQueueState `json:"statesByCluster,omitempty"`
}

// GetStatesByCluster is an internal getter (TBD...)
func (v *ProcessingQueueStates) GetStatesByCluster() (o map[string][]*ProcessingQueueState) {
	if v != nil && v.StatesByCluster != nil {
		return v.StatesByCluster
	}
	return
}

// HistoryQueryWorkflowRequest is an internal type (TBD...)
type HistoryQueryWorkflowRequest struct {
	DomainUUID string                `json:"domainUUID,omitempty"`
	Request    *QueryWorkflowRequest `json:"request,omitempty"`
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
	Response *QueryWorkflowResponse `json:"response,omitempty"`
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
	DomainUUID string                `json:"domainUUID,omitempty"`
	Request    *ReapplyEventsRequest `json:"request,omitempty"`
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
	DomainUUID       string                              `json:"domainUUID,omitempty"`
	HeartbeatRequest *RecordActivityTaskHeartbeatRequest `json:"heartbeatRequest,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRecordActivityTaskHeartbeatRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetHeartbeatRequest is an internal getter (TBD...)
func (v *HistoryRecordActivityTaskHeartbeatRequest) GetHeartbeatRequest() (o *RecordActivityTaskHeartbeatRequest) {
	if v != nil && v.HeartbeatRequest != nil {
		return v.HeartbeatRequest
	}
	return
}

// RecordActivityTaskStartedRequest is an internal type (TBD...)
type RecordActivityTaskStartedRequest struct {
	DomainUUID        string                      `json:"domainUUID,omitempty"`
	WorkflowExecution *WorkflowExecution          `json:"workflowExecution,omitempty"`
	ScheduleID        int64                       `json:"scheduleId,omitempty"`
	TaskID            int64                       `json:"taskId,omitempty"`
	RequestID         string                      `json:"requestId,omitempty"`
	PollRequest       *PollForActivityTaskRequest `json:"pollRequest,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *RecordActivityTaskStartedRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetWorkflowExecution is an internal getter (TBD...)
func (v *RecordActivityTaskStartedRequest) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
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

// GetPollRequest is an internal getter (TBD...)
func (v *RecordActivityTaskStartedRequest) GetPollRequest() (o *PollForActivityTaskRequest) {
	if v != nil && v.PollRequest != nil {
		return v.PollRequest
	}
	return
}

// RecordActivityTaskStartedResponse is an internal type (TBD...)
type RecordActivityTaskStartedResponse struct {
	ScheduledEvent                  *HistoryEvent `json:"scheduledEvent,omitempty"`
	StartedTimestamp                *int64        `json:"startedTimestamp,omitempty"`
	Attempt                         int64         `json:"attempt,omitempty"`
	ScheduledTimestampOfThisAttempt *int64        `json:"scheduledTimestampOfThisAttempt,omitempty"`
	HeartbeatDetails                []byte        `json:"heartbeatDetails,omitempty"`
	WorkflowType                    *WorkflowType `json:"workflowType,omitempty"`
	WorkflowDomain                  string        `json:"workflowDomain,omitempty"`
}

// GetScheduledEvent is an internal getter (TBD...)
func (v *RecordActivityTaskStartedResponse) GetScheduledEvent() (o *HistoryEvent) {
	if v != nil && v.ScheduledEvent != nil {
		return v.ScheduledEvent
	}
	return
}

// GetStartedTimestamp is an internal getter (TBD...)
func (v *RecordActivityTaskStartedResponse) GetStartedTimestamp() (o int64) {
	if v != nil && v.StartedTimestamp != nil {
		return *v.StartedTimestamp
	}
	return
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

// GetHeartbeatDetails is an internal getter (TBD...)
func (v *RecordActivityTaskStartedResponse) GetHeartbeatDetails() (o []byte) {
	if v != nil && v.HeartbeatDetails != nil {
		return v.HeartbeatDetails
	}
	return
}

// GetWorkflowType is an internal getter (TBD...)
func (v *RecordActivityTaskStartedResponse) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

// GetWorkflowDomain is an internal getter (TBD...)
func (v *RecordActivityTaskStartedResponse) GetWorkflowDomain() (o string) {
	if v != nil {
		return v.WorkflowDomain
	}
	return
}

// RecordChildExecutionCompletedRequest is an internal type (TBD...)
type RecordChildExecutionCompletedRequest struct {
	DomainUUID         string             `json:"domainUUID,omitempty"`
	WorkflowExecution  *WorkflowExecution `json:"workflowExecution,omitempty"`
	InitiatedID        int64              `json:"initiatedId,omitempty"`
	CompletedExecution *WorkflowExecution `json:"completedExecution,omitempty"`
	CompletionEvent    *HistoryEvent      `json:"completionEvent,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *RecordChildExecutionCompletedRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetWorkflowExecution is an internal getter (TBD...)
func (v *RecordChildExecutionCompletedRequest) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

// GetInitiatedID is an internal getter (TBD...)
func (v *RecordChildExecutionCompletedRequest) GetInitiatedID() (o int64) {
	if v != nil {
		return v.InitiatedID
	}
	return
}

// GetCompletedExecution is an internal getter (TBD...)
func (v *RecordChildExecutionCompletedRequest) GetCompletedExecution() (o *WorkflowExecution) {
	if v != nil && v.CompletedExecution != nil {
		return v.CompletedExecution
	}
	return
}

// GetCompletionEvent is an internal getter (TBD...)
func (v *RecordChildExecutionCompletedRequest) GetCompletionEvent() (o *HistoryEvent) {
	if v != nil && v.CompletionEvent != nil {
		return v.CompletionEvent
	}
	return
}

// RecordDecisionTaskStartedRequest is an internal type (TBD...)
type RecordDecisionTaskStartedRequest struct {
	DomainUUID        string                      `json:"domainUUID,omitempty"`
	WorkflowExecution *WorkflowExecution          `json:"workflowExecution,omitempty"`
	ScheduleID        int64                       `json:"scheduleId,omitempty"`
	TaskID            int64                       `json:"taskId,omitempty"`
	RequestID         string                      `json:"requestId,omitempty"`
	PollRequest       *PollForDecisionTaskRequest `json:"pollRequest,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetWorkflowExecution is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedRequest) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
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

// GetTaskID is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedRequest) GetTaskID() (o int64) {
	if v != nil {
		return v.TaskID
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

// GetPollRequest is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedRequest) GetPollRequest() (o *PollForDecisionTaskRequest) {
	if v != nil && v.PollRequest != nil {
		return v.PollRequest
	}
	return
}

// RecordDecisionTaskStartedResponse is an internal type (TBD...)
type RecordDecisionTaskStartedResponse struct {
	WorkflowType              *WorkflowType             `json:"workflowType,omitempty"`
	PreviousStartedEventID    *int64                    `json:"previousStartedEventId,omitempty"`
	ScheduledEventID          int64                     `json:"scheduledEventId,omitempty"`
	StartedEventID            int64                     `json:"startedEventId,omitempty"`
	NextEventID               int64                     `json:"nextEventId,omitempty"`
	Attempt                   int64                     `json:"attempt,omitempty"`
	StickyExecutionEnabled    bool                      `json:"stickyExecutionEnabled,omitempty"`
	DecisionInfo              *TransientDecisionInfo    `json:"decisionInfo,omitempty"`
	WorkflowExecutionTaskList *TaskList                 `json:"WorkflowExecutionTaskList,omitempty"`
	EventStoreVersion         int32                     `json:"eventStoreVersion,omitempty"`
	BranchToken               []byte                    `json:"branchToken,omitempty"`
	ScheduledTimestamp        *int64                    `json:"scheduledTimestamp,omitempty"`
	StartedTimestamp          *int64                    `json:"startedTimestamp,omitempty"`
	Queries                   map[string]*WorkflowQuery `json:"queries,omitempty"`
}

// GetWorkflowType is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
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

// GetStartedEventID is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetStartedEventID() (o int64) {
	if v != nil {
		return v.StartedEventID
	}
	return
}

// GetNextEventID is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetNextEventID() (o int64) {
	if v != nil {
		return v.NextEventID
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

// GetStickyExecutionEnabled is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetStickyExecutionEnabled() (o bool) {
	if v != nil {
		return v.StickyExecutionEnabled
	}
	return
}

// GetDecisionInfo is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetDecisionInfo() (o *TransientDecisionInfo) {
	if v != nil && v.DecisionInfo != nil {
		return v.DecisionInfo
	}
	return
}

// GetWorkflowExecutionTaskList is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetWorkflowExecutionTaskList() (o *TaskList) {
	if v != nil && v.WorkflowExecutionTaskList != nil {
		return v.WorkflowExecutionTaskList
	}
	return
}

// GetEventStoreVersion is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetEventStoreVersion() (o int32) {
	if v != nil {
		return v.EventStoreVersion
	}
	return
}

// GetBranchToken is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetBranchToken() (o []byte) {
	if v != nil && v.BranchToken != nil {
		return v.BranchToken
	}
	return
}

// GetScheduledTimestamp is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetScheduledTimestamp() (o int64) {
	if v != nil && v.ScheduledTimestamp != nil {
		return *v.ScheduledTimestamp
	}
	return
}

// GetStartedTimestamp is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetStartedTimestamp() (o int64) {
	if v != nil && v.StartedTimestamp != nil {
		return *v.StartedTimestamp
	}
	return
}

// GetQueries is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetQueries() (o map[string]*WorkflowQuery) {
	if v != nil && v.Queries != nil {
		return v.Queries
	}
	return
}

// HistoryRefreshWorkflowTasksRequest is an internal type (TBD...)
type HistoryRefreshWorkflowTasksRequest struct {
	DomainUIID string                       `json:"domainUIID,omitempty"`
	Request    *RefreshWorkflowTasksRequest `json:"request,omitempty"`
}

// GetDomainUIID is an internal getter (TBD...)
func (v *HistoryRefreshWorkflowTasksRequest) GetDomainUIID() (o string) {
	if v != nil {
		return v.DomainUIID
	}
	return
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
	DomainUUID        string             `json:"domainUUID,omitempty"`
	WorkflowExecution *WorkflowExecution `json:"workflowExecution,omitempty"`
	RequestID         string             `json:"requestId,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *RemoveSignalMutableStateRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetWorkflowExecution is an internal getter (TBD...)
func (v *RemoveSignalMutableStateRequest) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
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
	DomainUUID          string                `json:"domainUUID,omitempty"`
	WorkflowExecution   *WorkflowExecution    `json:"workflowExecution,omitempty"`
	VersionHistoryItems []*VersionHistoryItem `json:"versionHistoryItems,omitempty"`
	Events              *DataBlob             `json:"events,omitempty"`
	NewRunEvents        *DataBlob             `json:"newRunEvents,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *ReplicateEventsV2Request) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetWorkflowExecution is an internal getter (TBD...)
func (v *ReplicateEventsV2Request) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

// GetVersionHistoryItems is an internal getter (TBD...)
func (v *ReplicateEventsV2Request) GetVersionHistoryItems() (o []*VersionHistoryItem) {
	if v != nil && v.VersionHistoryItems != nil {
		return v.VersionHistoryItems
	}
	return
}

// GetEvents is an internal getter (TBD...)
func (v *ReplicateEventsV2Request) GetEvents() (o *DataBlob) {
	if v != nil && v.Events != nil {
		return v.Events
	}
	return
}

// GetNewRunEvents is an internal getter (TBD...)
func (v *ReplicateEventsV2Request) GetNewRunEvents() (o *DataBlob) {
	if v != nil && v.NewRunEvents != nil {
		return v.NewRunEvents
	}
	return
}

// HistoryRequestCancelWorkflowExecutionRequest is an internal type (TBD...)
type HistoryRequestCancelWorkflowExecutionRequest struct {
	DomainUUID                string                                 `json:"domainUUID,omitempty"`
	CancelRequest             *RequestCancelWorkflowExecutionRequest `json:"cancelRequest,omitempty"`
	ExternalInitiatedEventID  *int64                                 `json:"externalInitiatedEventId,omitempty"`
	ExternalWorkflowExecution *WorkflowExecution                     `json:"externalWorkflowExecution,omitempty"`
	ChildWorkflowOnly         bool                                   `json:"childWorkflowOnly,omitempty"`
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

// GetExternalInitiatedEventID is an internal getter (TBD...)
func (v *HistoryRequestCancelWorkflowExecutionRequest) GetExternalInitiatedEventID() (o int64) {
	if v != nil && v.ExternalInitiatedEventID != nil {
		return *v.ExternalInitiatedEventID
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
	DomainUUID string             `json:"domainUUID,omitempty"`
	Execution  *WorkflowExecution `json:"execution,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryResetStickyTaskListRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetExecution is an internal getter (TBD...)
func (v *HistoryResetStickyTaskListRequest) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
}

// HistoryResetStickyTaskListResponse is an internal type (TBD...)
type HistoryResetStickyTaskListResponse struct {
}

// HistoryResetWorkflowExecutionRequest is an internal type (TBD...)
type HistoryResetWorkflowExecutionRequest struct {
	DomainUUID   string                         `json:"domainUUID,omitempty"`
	ResetRequest *ResetWorkflowExecutionRequest `json:"resetRequest,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryResetWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetResetRequest is an internal getter (TBD...)
func (v *HistoryResetWorkflowExecutionRequest) GetResetRequest() (o *ResetWorkflowExecutionRequest) {
	if v != nil && v.ResetRequest != nil {
		return v.ResetRequest
	}
	return
}

// HistoryRespondActivityTaskCanceledRequest is an internal type (TBD...)
type HistoryRespondActivityTaskCanceledRequest struct {
	DomainUUID    string                              `json:"domainUUID,omitempty"`
	CancelRequest *RespondActivityTaskCanceledRequest `json:"cancelRequest,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRespondActivityTaskCanceledRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetCancelRequest is an internal getter (TBD...)
func (v *HistoryRespondActivityTaskCanceledRequest) GetCancelRequest() (o *RespondActivityTaskCanceledRequest) {
	if v != nil && v.CancelRequest != nil {
		return v.CancelRequest
	}
	return
}

// HistoryRespondActivityTaskCompletedRequest is an internal type (TBD...)
type HistoryRespondActivityTaskCompletedRequest struct {
	DomainUUID      string                               `json:"domainUUID,omitempty"`
	CompleteRequest *RespondActivityTaskCompletedRequest `json:"completeRequest,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRespondActivityTaskCompletedRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetCompleteRequest is an internal getter (TBD...)
func (v *HistoryRespondActivityTaskCompletedRequest) GetCompleteRequest() (o *RespondActivityTaskCompletedRequest) {
	if v != nil && v.CompleteRequest != nil {
		return v.CompleteRequest
	}
	return
}

// HistoryRespondActivityTaskFailedRequest is an internal type (TBD...)
type HistoryRespondActivityTaskFailedRequest struct {
	DomainUUID    string                            `json:"domainUUID,omitempty"`
	FailedRequest *RespondActivityTaskFailedRequest `json:"failedRequest,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRespondActivityTaskFailedRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetFailedRequest is an internal getter (TBD...)
func (v *HistoryRespondActivityTaskFailedRequest) GetFailedRequest() (o *RespondActivityTaskFailedRequest) {
	if v != nil && v.FailedRequest != nil {
		return v.FailedRequest
	}
	return
}

// HistoryRespondDecisionTaskCompletedRequest is an internal type (TBD...)
type HistoryRespondDecisionTaskCompletedRequest struct {
	DomainUUID      string                               `json:"domainUUID,omitempty"`
	CompleteRequest *RespondDecisionTaskCompletedRequest `json:"completeRequest,omitempty"`
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
	StartedResponse             *RecordDecisionTaskStartedResponse    `json:"startedResponse,omitempty"`
	ActivitiesToDispatchLocally map[string]*ActivityLocalDispatchInfo `json:"activitiesToDispatchLocally,omitempty"`
}

// GetStartedResponse is an internal getter (TBD...)
func (v *HistoryRespondDecisionTaskCompletedResponse) GetStartedResponse() (o *RecordDecisionTaskStartedResponse) {
	if v != nil && v.StartedResponse != nil {
		return v.StartedResponse
	}
	return
}

// GetActivitiesToDispatchLocally is an internal getter (TBD...)
func (v *HistoryRespondDecisionTaskCompletedResponse) GetActivitiesToDispatchLocally() (o map[string]*ActivityLocalDispatchInfo) {
	if v != nil && v.ActivitiesToDispatchLocally != nil {
		return v.ActivitiesToDispatchLocally
	}
	return
}

// HistoryRespondDecisionTaskFailedRequest is an internal type (TBD...)
type HistoryRespondDecisionTaskFailedRequest struct {
	DomainUUID    string                            `json:"domainUUID,omitempty"`
	FailedRequest *RespondDecisionTaskFailedRequest `json:"failedRequest,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRespondDecisionTaskFailedRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetFailedRequest is an internal getter (TBD...)
func (v *HistoryRespondDecisionTaskFailedRequest) GetFailedRequest() (o *RespondDecisionTaskFailedRequest) {
	if v != nil && v.FailedRequest != nil {
		return v.FailedRequest
	}
	return
}

// ScheduleDecisionTaskRequest is an internal type (TBD...)
type ScheduleDecisionTaskRequest struct {
	DomainUUID        string             `json:"domainUUID,omitempty"`
	WorkflowExecution *WorkflowExecution `json:"workflowExecution,omitempty"`
	IsFirstDecision   bool               `json:"isFirstDecision,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *ScheduleDecisionTaskRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetWorkflowExecution is an internal getter (TBD...)
func (v *ScheduleDecisionTaskRequest) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

// GetIsFirstDecision is an internal getter (TBD...)
func (v *ScheduleDecisionTaskRequest) GetIsFirstDecision() (o bool) {
	if v != nil {
		return v.IsFirstDecision
	}
	return
}

// ShardOwnershipLostError is an internal type (TBD...)
type ShardOwnershipLostError struct {
	Message string `json:"message,omitempty"`
	Owner   string `json:"owner,omitempty"`
}

// GetMessage is an internal getter (TBD...)
func (v *ShardOwnershipLostError) GetMessage() (o string) {
	if v != nil {
		return v.Message
	}
	return
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
	DomainUUID             string                                   `json:"domainUUID,omitempty"`
	SignalWithStartRequest *SignalWithStartWorkflowExecutionRequest `json:"signalWithStartRequest,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistorySignalWithStartWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetSignalWithStartRequest is an internal getter (TBD...)
func (v *HistorySignalWithStartWorkflowExecutionRequest) GetSignalWithStartRequest() (o *SignalWithStartWorkflowExecutionRequest) {
	if v != nil && v.SignalWithStartRequest != nil {
		return v.SignalWithStartRequest
	}
	return
}

// HistorySignalWorkflowExecutionRequest is an internal type (TBD...)
type HistorySignalWorkflowExecutionRequest struct {
	DomainUUID                string                          `json:"domainUUID,omitempty"`
	SignalRequest             *SignalWorkflowExecutionRequest `json:"signalRequest,omitempty"`
	ExternalWorkflowExecution *WorkflowExecution              `json:"externalWorkflowExecution,omitempty"`
	ChildWorkflowOnly         bool                            `json:"childWorkflowOnly,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistorySignalWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetSignalRequest is an internal getter (TBD...)
func (v *HistorySignalWorkflowExecutionRequest) GetSignalRequest() (o *SignalWorkflowExecutionRequest) {
	if v != nil && v.SignalRequest != nil {
		return v.SignalRequest
	}
	return
}

// GetExternalWorkflowExecution is an internal getter (TBD...)
func (v *HistorySignalWorkflowExecutionRequest) GetExternalWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.ExternalWorkflowExecution != nil {
		return v.ExternalWorkflowExecution
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
	DomainUUID                      string                         `json:"domainUUID,omitempty"`
	StartRequest                    *StartWorkflowExecutionRequest `json:"startRequest,omitempty"`
	ParentExecutionInfo             *ParentExecutionInfo           `json:"parentExecutionInfo,omitempty"`
	Attempt                         int32                          `json:"attempt,omitempty"`
	ExpirationTimestamp             *int64                         `json:"expirationTimestamp,omitempty"`
	ContinueAsNewInitiator          *ContinueAsNewInitiator        `json:"continueAsNewInitiator,omitempty"`
	ContinuedFailureReason          *string                        `json:"continuedFailureReason,omitempty"`
	ContinuedFailureDetails         []byte                         `json:"continuedFailureDetails,omitempty"`
	LastCompletionResult            []byte                         `json:"lastCompletionResult,omitempty"`
	FirstDecisionTaskBackoffSeconds *int32                         `json:"firstDecisionTaskBackoffSeconds,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryStartWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil {
		return v.DomainUUID
	}
	return
}

// GetStartRequest is an internal getter (TBD...)
func (v *HistoryStartWorkflowExecutionRequest) GetStartRequest() (o *StartWorkflowExecutionRequest) {
	if v != nil && v.StartRequest != nil {
		return v.StartRequest
	}
	return
}

// GetParentExecutionInfo is an internal getter (TBD...)
func (v *HistoryStartWorkflowExecutionRequest) GetParentExecutionInfo() (o *ParentExecutionInfo) {
	if v != nil && v.ParentExecutionInfo != nil {
		return v.ParentExecutionInfo
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

// GetContinueAsNewInitiator is an internal getter (TBD...)
func (v *HistoryStartWorkflowExecutionRequest) GetContinueAsNewInitiator() (o *ContinueAsNewInitiator) {
	if v != nil && v.ContinueAsNewInitiator != nil {
		return v.ContinueAsNewInitiator
	}
	return
}

// GetContinuedFailureReason is an internal getter (TBD...)
func (v *HistoryStartWorkflowExecutionRequest) GetContinuedFailureReason() (o string) {
	if v != nil && v.ContinuedFailureReason != nil {
		return *v.ContinuedFailureReason
	}
	return
}

// GetContinuedFailureDetails is an internal getter (TBD...)
func (v *HistoryStartWorkflowExecutionRequest) GetContinuedFailureDetails() (o []byte) {
	if v != nil && v.ContinuedFailureDetails != nil {
		return v.ContinuedFailureDetails
	}
	return
}

// GetLastCompletionResult is an internal getter (TBD...)
func (v *HistoryStartWorkflowExecutionRequest) GetLastCompletionResult() (o []byte) {
	if v != nil && v.LastCompletionResult != nil {
		return v.LastCompletionResult
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
	DomainID           string          `json:"domainId,omitempty"`
	WorkflowID         string          `json:"workflowId,omitempty"`
	RunID              string          `json:"runId,omitempty"`
	Version            int64           `json:"version,omitempty"`
	ScheduledID        int64           `json:"scheduledId,omitempty"`
	ScheduledTime      *int64          `json:"scheduledTime,omitempty"`
	StartedID          int64           `json:"startedId,omitempty"`
	StartedTime        *int64          `json:"startedTime,omitempty"`
	LastHeartbeatTime  *int64          `json:"lastHeartbeatTime,omitempty"`
	Details            []byte          `json:"details,omitempty"`
	Attempt            int32           `json:"attempt,omitempty"`
	LastFailureReason  *string         `json:"lastFailureReason,omitempty"`
	LastWorkerIdentity string          `json:"lastWorkerIdentity,omitempty"`
	LastFailureDetails []byte          `json:"lastFailureDetails,omitempty"`
	VersionHistory     *VersionHistory `json:"versionHistory,omitempty"`
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
	SourceCluster string `json:"sourceCluster,omitempty"`
	ShardID       int64  `json:"shardId,omitempty"`
	Timestamp     *int64 `json:"timestamp,omitempty"`
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
	DomainUUID                string                             `json:"domainUUID,omitempty"`
	TerminateRequest          *TerminateWorkflowExecutionRequest `json:"terminateRequest,omitempty"`
	ExternalWorkflowExecution *WorkflowExecution                 `json:"externalWorkflowExecution,omitempty"`
	ChildWorkflowOnly         bool                               `json:"childWorkflowOnly,omitempty"`
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
	DomainID string `json:"domainID,omitempty"`
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
	CompletedShardCount int32   `json:"completedShardCount,omitempty"`
	PendingShards       []int32 `json:"pendingShards,omitempty"`
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
