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
	DomainUUID *string
	Execution  *WorkflowExecution
}

// GetDomainUUID is an internal getter (TBD...)
func (v *DescribeMutableStateRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	MutableStateInCache    *string
	MutableStateInDatabase *string
}

// GetMutableStateInCache is an internal getter (TBD...)
func (v *DescribeMutableStateResponse) GetMutableStateInCache() (o string) {
	if v != nil && v.MutableStateInCache != nil {
		return *v.MutableStateInCache
	}
	return
}

// GetMutableStateInDatabase is an internal getter (TBD...)
func (v *DescribeMutableStateResponse) GetMutableStateInDatabase() (o string) {
	if v != nil && v.MutableStateInDatabase != nil {
		return *v.MutableStateInDatabase
	}
	return
}

// HistoryDescribeWorkflowExecutionRequest is an internal type (TBD...)
type HistoryDescribeWorkflowExecutionRequest struct {
	DomainUUID *string
	Request    *DescribeWorkflowExecutionRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryDescribeWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	DomainIDs    []string
	ReverseMatch *bool
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
	if v != nil && v.ReverseMatch != nil {
		return *v.ReverseMatch
	}
	return
}

// EventAlreadyStartedError is an internal type (TBD...)
type EventAlreadyStartedError struct {
	Message string
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
	DomainUUID          *string
	Execution           *WorkflowExecution
	ExpectedNextEventID *int64
	CurrentBranchToken  []byte
}

// GetDomainUUID is an internal getter (TBD...)
func (v *GetMutableStateRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	if v != nil && v.ExpectedNextEventID != nil {
		return *v.ExpectedNextEventID
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
	Execution                            *WorkflowExecution
	WorkflowType                         *WorkflowType
	NextEventID                          *int64
	PreviousStartedEventID               *int64
	LastFirstEventID                     *int64
	TaskList                             *TaskList
	StickyTaskList                       *TaskList
	ClientLibraryVersion                 *string
	ClientFeatureVersion                 *string
	ClientImpl                           *string
	IsWorkflowRunning                    *bool
	StickyTaskListScheduleToStartTimeout *int32
	EventStoreVersion                    *int32
	CurrentBranchToken                   []byte
	WorkflowState                        *int32
	WorkflowCloseState                   *int32
	VersionHistories                     *VersionHistories
	IsStickyTaskListEnabled              *bool
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
	if v != nil && v.NextEventID != nil {
		return *v.NextEventID
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
	if v != nil && v.LastFirstEventID != nil {
		return *v.LastFirstEventID
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
	if v != nil && v.ClientLibraryVersion != nil {
		return *v.ClientLibraryVersion
	}
	return
}

// GetClientFeatureVersion is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetClientFeatureVersion() (o string) {
	if v != nil && v.ClientFeatureVersion != nil {
		return *v.ClientFeatureVersion
	}
	return
}

// GetClientImpl is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetClientImpl() (o string) {
	if v != nil && v.ClientImpl != nil {
		return *v.ClientImpl
	}
	return
}

// GetIsWorkflowRunning is an internal getter (TBD...)
func (v *GetMutableStateResponse) GetIsWorkflowRunning() (o bool) {
	if v != nil && v.IsWorkflowRunning != nil {
		return *v.IsWorkflowRunning
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
	if v != nil && v.EventStoreVersion != nil {
		return *v.EventStoreVersion
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
	if v != nil && v.IsStickyTaskListEnabled != nil {
		return *v.IsStickyTaskListEnabled
	}
	return
}

// HistoryService_CloseShard_Args is an internal type (TBD...)
type HistoryService_CloseShard_Args struct {
	Request *CloseShardRequest
}

// GetRequest is an internal getter (TBD...)
func (v *HistoryService_CloseShard_Args) GetRequest() (o *CloseShardRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// HistoryService_CloseShard_Result is an internal type (TBD...)
type HistoryService_CloseShard_Result struct {
	BadRequestError      *BadRequestError
	InternalServiceError *InternalServiceError
	AccessDeniedError    *AccessDeniedError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_CloseShard_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_CloseShard_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetAccessDeniedError is an internal getter (TBD...)
func (v *HistoryService_CloseShard_Result) GetAccessDeniedError() (o *AccessDeniedError) {
	if v != nil && v.AccessDeniedError != nil {
		return v.AccessDeniedError
	}
	return
}

// HistoryService_DescribeHistoryHost_Args is an internal type (TBD...)
type HistoryService_DescribeHistoryHost_Args struct {
	Request *DescribeHistoryHostRequest
}

// GetRequest is an internal getter (TBD...)
func (v *HistoryService_DescribeHistoryHost_Args) GetRequest() (o *DescribeHistoryHostRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// HistoryService_DescribeHistoryHost_Result is an internal type (TBD...)
type HistoryService_DescribeHistoryHost_Result struct {
	Success              *DescribeHistoryHostResponse
	BadRequestError      *BadRequestError
	InternalServiceError *InternalServiceError
	AccessDeniedError    *AccessDeniedError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_DescribeHistoryHost_Result) GetSuccess() (o *DescribeHistoryHostResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_DescribeHistoryHost_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_DescribeHistoryHost_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetAccessDeniedError is an internal getter (TBD...)
func (v *HistoryService_DescribeHistoryHost_Result) GetAccessDeniedError() (o *AccessDeniedError) {
	if v != nil && v.AccessDeniedError != nil {
		return v.AccessDeniedError
	}
	return
}

// HistoryService_DescribeMutableState_Args is an internal type (TBD...)
type HistoryService_DescribeMutableState_Args struct {
	Request *DescribeMutableStateRequest
}

// GetRequest is an internal getter (TBD...)
func (v *HistoryService_DescribeMutableState_Args) GetRequest() (o *DescribeMutableStateRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// HistoryService_DescribeMutableState_Result is an internal type (TBD...)
type HistoryService_DescribeMutableState_Result struct {
	Success                 *DescribeMutableStateResponse
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	EntityNotExistError     *EntityNotExistsError
	AccessDeniedError       *AccessDeniedError
	ShardOwnershipLostError *ShardOwnershipLostError
	LimitExceededError      *LimitExceededError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_DescribeMutableState_Result) GetSuccess() (o *DescribeMutableStateResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_DescribeMutableState_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_DescribeMutableState_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_DescribeMutableState_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetAccessDeniedError is an internal getter (TBD...)
func (v *HistoryService_DescribeMutableState_Result) GetAccessDeniedError() (o *AccessDeniedError) {
	if v != nil && v.AccessDeniedError != nil {
		return v.AccessDeniedError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_DescribeMutableState_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_DescribeMutableState_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// HistoryService_DescribeQueue_Args is an internal type (TBD...)
type HistoryService_DescribeQueue_Args struct {
	Request *DescribeQueueRequest
}

// GetRequest is an internal getter (TBD...)
func (v *HistoryService_DescribeQueue_Args) GetRequest() (o *DescribeQueueRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// HistoryService_DescribeQueue_Result is an internal type (TBD...)
type HistoryService_DescribeQueue_Result struct {
	Success              *DescribeQueueResponse
	BadRequestError      *BadRequestError
	InternalServiceError *InternalServiceError
	AccessDeniedError    *AccessDeniedError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_DescribeQueue_Result) GetSuccess() (o *DescribeQueueResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_DescribeQueue_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_DescribeQueue_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetAccessDeniedError is an internal getter (TBD...)
func (v *HistoryService_DescribeQueue_Result) GetAccessDeniedError() (o *AccessDeniedError) {
	if v != nil && v.AccessDeniedError != nil {
		return v.AccessDeniedError
	}
	return
}

// HistoryService_DescribeWorkflowExecution_Args is an internal type (TBD...)
type HistoryService_DescribeWorkflowExecution_Args struct {
	DescribeRequest *HistoryDescribeWorkflowExecutionRequest
}

// GetDescribeRequest is an internal getter (TBD...)
func (v *HistoryService_DescribeWorkflowExecution_Args) GetDescribeRequest() (o *HistoryDescribeWorkflowExecutionRequest) {
	if v != nil && v.DescribeRequest != nil {
		return v.DescribeRequest
	}
	return
}

// HistoryService_DescribeWorkflowExecution_Result is an internal type (TBD...)
type HistoryService_DescribeWorkflowExecution_Result struct {
	Success                 *DescribeWorkflowExecutionResponse
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
	LimitExceededError      *LimitExceededError
	ServiceBusyError        *ServiceBusyError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_DescribeWorkflowExecution_Result) GetSuccess() (o *DescribeWorkflowExecutionResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_DescribeWorkflowExecution_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_DescribeWorkflowExecution_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_DescribeWorkflowExecution_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_DescribeWorkflowExecution_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_DescribeWorkflowExecution_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_DescribeWorkflowExecution_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_GetDLQReplicationMessages_Args is an internal type (TBD...)
type HistoryService_GetDLQReplicationMessages_Args struct {
	Request *GetDLQReplicationMessagesRequest
}

// GetRequest is an internal getter (TBD...)
func (v *HistoryService_GetDLQReplicationMessages_Args) GetRequest() (o *GetDLQReplicationMessagesRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// HistoryService_GetDLQReplicationMessages_Result is an internal type (TBD...)
type HistoryService_GetDLQReplicationMessages_Result struct {
	Success              *GetDLQReplicationMessagesResponse
	BadRequestError      *BadRequestError
	InternalServiceError *InternalServiceError
	ServiceBusyError     *ServiceBusyError
	EntityNotExistError  *EntityNotExistsError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_GetDLQReplicationMessages_Result) GetSuccess() (o *GetDLQReplicationMessagesResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_GetDLQReplicationMessages_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_GetDLQReplicationMessages_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_GetDLQReplicationMessages_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_GetDLQReplicationMessages_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// HistoryService_GetMutableState_Args is an internal type (TBD...)
type HistoryService_GetMutableState_Args struct {
	GetRequest *GetMutableStateRequest
}

// GetGetRequest is an internal getter (TBD...)
func (v *HistoryService_GetMutableState_Args) GetGetRequest() (o *GetMutableStateRequest) {
	if v != nil && v.GetRequest != nil {
		return v.GetRequest
	}
	return
}

// HistoryService_GetMutableState_Result is an internal type (TBD...)
type HistoryService_GetMutableState_Result struct {
	Success                   *GetMutableStateResponse
	BadRequestError           *BadRequestError
	InternalServiceError      *InternalServiceError
	EntityNotExistError       *EntityNotExistsError
	ShardOwnershipLostError   *ShardOwnershipLostError
	LimitExceededError        *LimitExceededError
	ServiceBusyError          *ServiceBusyError
	CurrentBranchChangedError *CurrentBranchChangedError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_GetMutableState_Result) GetSuccess() (o *GetMutableStateResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_GetMutableState_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_GetMutableState_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_GetMutableState_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_GetMutableState_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_GetMutableState_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_GetMutableState_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetCurrentBranchChangedError is an internal getter (TBD...)
func (v *HistoryService_GetMutableState_Result) GetCurrentBranchChangedError() (o *CurrentBranchChangedError) {
	if v != nil && v.CurrentBranchChangedError != nil {
		return v.CurrentBranchChangedError
	}
	return
}

// HistoryService_GetReplicationMessages_Args is an internal type (TBD...)
type HistoryService_GetReplicationMessages_Args struct {
	Request *GetReplicationMessagesRequest
}

// GetRequest is an internal getter (TBD...)
func (v *HistoryService_GetReplicationMessages_Args) GetRequest() (o *GetReplicationMessagesRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// HistoryService_GetReplicationMessages_Result is an internal type (TBD...)
type HistoryService_GetReplicationMessages_Result struct {
	Success                        *GetReplicationMessagesResponse
	BadRequestError                *BadRequestError
	InternalServiceError           *InternalServiceError
	LimitExceededError             *LimitExceededError
	ServiceBusyError               *ServiceBusyError
	ClientVersionNotSupportedError *ClientVersionNotSupportedError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_GetReplicationMessages_Result) GetSuccess() (o *GetReplicationMessagesResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_GetReplicationMessages_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_GetReplicationMessages_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_GetReplicationMessages_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_GetReplicationMessages_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetClientVersionNotSupportedError is an internal getter (TBD...)
func (v *HistoryService_GetReplicationMessages_Result) GetClientVersionNotSupportedError() (o *ClientVersionNotSupportedError) {
	if v != nil && v.ClientVersionNotSupportedError != nil {
		return v.ClientVersionNotSupportedError
	}
	return
}

// HistoryService_MergeDLQMessages_Args is an internal type (TBD...)
type HistoryService_MergeDLQMessages_Args struct {
	Request *MergeDLQMessagesRequest
}

// GetRequest is an internal getter (TBD...)
func (v *HistoryService_MergeDLQMessages_Args) GetRequest() (o *MergeDLQMessagesRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// HistoryService_MergeDLQMessages_Result is an internal type (TBD...)
type HistoryService_MergeDLQMessages_Result struct {
	Success                 *MergeDLQMessagesResponse
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	ServiceBusyError        *ServiceBusyError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_MergeDLQMessages_Result) GetSuccess() (o *MergeDLQMessagesResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_MergeDLQMessages_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_MergeDLQMessages_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_MergeDLQMessages_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_MergeDLQMessages_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_MergeDLQMessages_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// HistoryService_NotifyFailoverMarkers_Args is an internal type (TBD...)
type HistoryService_NotifyFailoverMarkers_Args struct {
	Request *NotifyFailoverMarkersRequest
}

// GetRequest is an internal getter (TBD...)
func (v *HistoryService_NotifyFailoverMarkers_Args) GetRequest() (o *NotifyFailoverMarkersRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// HistoryService_NotifyFailoverMarkers_Result is an internal type (TBD...)
type HistoryService_NotifyFailoverMarkers_Result struct {
	BadRequestError      *BadRequestError
	InternalServiceError *InternalServiceError
	ServiceBusyError     *ServiceBusyError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_NotifyFailoverMarkers_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_NotifyFailoverMarkers_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_NotifyFailoverMarkers_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_PollMutableState_Args is an internal type (TBD...)
type HistoryService_PollMutableState_Args struct {
	PollRequest *PollMutableStateRequest
}

// GetPollRequest is an internal getter (TBD...)
func (v *HistoryService_PollMutableState_Args) GetPollRequest() (o *PollMutableStateRequest) {
	if v != nil && v.PollRequest != nil {
		return v.PollRequest
	}
	return
}

// HistoryService_PollMutableState_Result is an internal type (TBD...)
type HistoryService_PollMutableState_Result struct {
	Success                   *PollMutableStateResponse
	BadRequestError           *BadRequestError
	InternalServiceError      *InternalServiceError
	EntityNotExistError       *EntityNotExistsError
	ShardOwnershipLostError   *ShardOwnershipLostError
	LimitExceededError        *LimitExceededError
	ServiceBusyError          *ServiceBusyError
	CurrentBranchChangedError *CurrentBranchChangedError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_PollMutableState_Result) GetSuccess() (o *PollMutableStateResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_PollMutableState_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_PollMutableState_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_PollMutableState_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_PollMutableState_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_PollMutableState_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_PollMutableState_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetCurrentBranchChangedError is an internal getter (TBD...)
func (v *HistoryService_PollMutableState_Result) GetCurrentBranchChangedError() (o *CurrentBranchChangedError) {
	if v != nil && v.CurrentBranchChangedError != nil {
		return v.CurrentBranchChangedError
	}
	return
}

// HistoryService_PurgeDLQMessages_Args is an internal type (TBD...)
type HistoryService_PurgeDLQMessages_Args struct {
	Request *PurgeDLQMessagesRequest
}

// GetRequest is an internal getter (TBD...)
func (v *HistoryService_PurgeDLQMessages_Args) GetRequest() (o *PurgeDLQMessagesRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// HistoryService_PurgeDLQMessages_Result is an internal type (TBD...)
type HistoryService_PurgeDLQMessages_Result struct {
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	ServiceBusyError        *ServiceBusyError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_PurgeDLQMessages_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_PurgeDLQMessages_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_PurgeDLQMessages_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_PurgeDLQMessages_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_PurgeDLQMessages_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// HistoryService_QueryWorkflow_Args is an internal type (TBD...)
type HistoryService_QueryWorkflow_Args struct {
	QueryRequest *HistoryQueryWorkflowRequest
}

// GetQueryRequest is an internal getter (TBD...)
func (v *HistoryService_QueryWorkflow_Args) GetQueryRequest() (o *HistoryQueryWorkflowRequest) {
	if v != nil && v.QueryRequest != nil {
		return v.QueryRequest
	}
	return
}

// HistoryService_QueryWorkflow_Result is an internal type (TBD...)
type HistoryService_QueryWorkflow_Result struct {
	Success                        *HistoryQueryWorkflowResponse
	BadRequestError                *BadRequestError
	InternalServiceError           *InternalServiceError
	EntityNotExistError            *EntityNotExistsError
	QueryFailedError               *QueryFailedError
	LimitExceededError             *LimitExceededError
	ServiceBusyError               *ServiceBusyError
	ClientVersionNotSupportedError *ClientVersionNotSupportedError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_QueryWorkflow_Result) GetSuccess() (o *HistoryQueryWorkflowResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_QueryWorkflow_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_QueryWorkflow_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_QueryWorkflow_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetQueryFailedError is an internal getter (TBD...)
func (v *HistoryService_QueryWorkflow_Result) GetQueryFailedError() (o *QueryFailedError) {
	if v != nil && v.QueryFailedError != nil {
		return v.QueryFailedError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_QueryWorkflow_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_QueryWorkflow_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetClientVersionNotSupportedError is an internal getter (TBD...)
func (v *HistoryService_QueryWorkflow_Result) GetClientVersionNotSupportedError() (o *ClientVersionNotSupportedError) {
	if v != nil && v.ClientVersionNotSupportedError != nil {
		return v.ClientVersionNotSupportedError
	}
	return
}

// HistoryService_ReadDLQMessages_Args is an internal type (TBD...)
type HistoryService_ReadDLQMessages_Args struct {
	Request *ReadDLQMessagesRequest
}

// GetRequest is an internal getter (TBD...)
func (v *HistoryService_ReadDLQMessages_Args) GetRequest() (o *ReadDLQMessagesRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// HistoryService_ReadDLQMessages_Result is an internal type (TBD...)
type HistoryService_ReadDLQMessages_Result struct {
	Success                 *ReadDLQMessagesResponse
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	ServiceBusyError        *ServiceBusyError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_ReadDLQMessages_Result) GetSuccess() (o *ReadDLQMessagesResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_ReadDLQMessages_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_ReadDLQMessages_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_ReadDLQMessages_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_ReadDLQMessages_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_ReadDLQMessages_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// HistoryService_ReapplyEvents_Args is an internal type (TBD...)
type HistoryService_ReapplyEvents_Args struct {
	ReapplyEventsRequest *HistoryReapplyEventsRequest
}

// GetReapplyEventsRequest is an internal getter (TBD...)
func (v *HistoryService_ReapplyEvents_Args) GetReapplyEventsRequest() (o *HistoryReapplyEventsRequest) {
	if v != nil && v.ReapplyEventsRequest != nil {
		return v.ReapplyEventsRequest
	}
	return
}

// HistoryService_ReapplyEvents_Result is an internal type (TBD...)
type HistoryService_ReapplyEvents_Result struct {
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	DomainNotActiveError    *DomainNotActiveError
	LimitExceededError      *LimitExceededError
	ServiceBusyError        *ServiceBusyError
	ShardOwnershipLostError *ShardOwnershipLostError
	EntityNotExistError     *EntityNotExistsError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_ReapplyEvents_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_ReapplyEvents_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_ReapplyEvents_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_ReapplyEvents_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_ReapplyEvents_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_ReapplyEvents_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_ReapplyEvents_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// HistoryService_RecordActivityTaskHeartbeat_Args is an internal type (TBD...)
type HistoryService_RecordActivityTaskHeartbeat_Args struct {
	HeartbeatRequest *HistoryRecordActivityTaskHeartbeatRequest
}

// GetHeartbeatRequest is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskHeartbeat_Args) GetHeartbeatRequest() (o *HistoryRecordActivityTaskHeartbeatRequest) {
	if v != nil && v.HeartbeatRequest != nil {
		return v.HeartbeatRequest
	}
	return
}

// HistoryService_RecordActivityTaskHeartbeat_Result is an internal type (TBD...)
type HistoryService_RecordActivityTaskHeartbeat_Result struct {
	Success                 *RecordActivityTaskHeartbeatResponse
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
	DomainNotActiveError    *DomainNotActiveError
	LimitExceededError      *LimitExceededError
	ServiceBusyError        *ServiceBusyError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskHeartbeat_Result) GetSuccess() (o *RecordActivityTaskHeartbeatResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskHeartbeat_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskHeartbeat_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskHeartbeat_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskHeartbeat_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskHeartbeat_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskHeartbeat_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskHeartbeat_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_RecordActivityTaskStarted_Args is an internal type (TBD...)
type HistoryService_RecordActivityTaskStarted_Args struct {
	AddRequest *RecordActivityTaskStartedRequest
}

// GetAddRequest is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskStarted_Args) GetAddRequest() (o *RecordActivityTaskStartedRequest) {
	if v != nil && v.AddRequest != nil {
		return v.AddRequest
	}
	return
}

// HistoryService_RecordActivityTaskStarted_Result is an internal type (TBD...)
type HistoryService_RecordActivityTaskStarted_Result struct {
	Success                  *RecordActivityTaskStartedResponse
	BadRequestError          *BadRequestError
	InternalServiceError     *InternalServiceError
	EventAlreadyStartedError *EventAlreadyStartedError
	EntityNotExistError      *EntityNotExistsError
	ShardOwnershipLostError  *ShardOwnershipLostError
	DomainNotActiveError     *DomainNotActiveError
	LimitExceededError       *LimitExceededError
	ServiceBusyError         *ServiceBusyError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskStarted_Result) GetSuccess() (o *RecordActivityTaskStartedResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskStarted_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskStarted_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEventAlreadyStartedError is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskStarted_Result) GetEventAlreadyStartedError() (o *EventAlreadyStartedError) {
	if v != nil && v.EventAlreadyStartedError != nil {
		return v.EventAlreadyStartedError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskStarted_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskStarted_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskStarted_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskStarted_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_RecordActivityTaskStarted_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_RecordChildExecutionCompleted_Args is an internal type (TBD...)
type HistoryService_RecordChildExecutionCompleted_Args struct {
	CompletionRequest *RecordChildExecutionCompletedRequest
}

// GetCompletionRequest is an internal getter (TBD...)
func (v *HistoryService_RecordChildExecutionCompleted_Args) GetCompletionRequest() (o *RecordChildExecutionCompletedRequest) {
	if v != nil && v.CompletionRequest != nil {
		return v.CompletionRequest
	}
	return
}

// HistoryService_RecordChildExecutionCompleted_Result is an internal type (TBD...)
type HistoryService_RecordChildExecutionCompleted_Result struct {
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
	DomainNotActiveError    *DomainNotActiveError
	LimitExceededError      *LimitExceededError
	ServiceBusyError        *ServiceBusyError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_RecordChildExecutionCompleted_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_RecordChildExecutionCompleted_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_RecordChildExecutionCompleted_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_RecordChildExecutionCompleted_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_RecordChildExecutionCompleted_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_RecordChildExecutionCompleted_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_RecordChildExecutionCompleted_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_RecordDecisionTaskStarted_Args is an internal type (TBD...)
type HistoryService_RecordDecisionTaskStarted_Args struct {
	AddRequest *RecordDecisionTaskStartedRequest
}

// GetAddRequest is an internal getter (TBD...)
func (v *HistoryService_RecordDecisionTaskStarted_Args) GetAddRequest() (o *RecordDecisionTaskStartedRequest) {
	if v != nil && v.AddRequest != nil {
		return v.AddRequest
	}
	return
}

// HistoryService_RecordDecisionTaskStarted_Result is an internal type (TBD...)
type HistoryService_RecordDecisionTaskStarted_Result struct {
	Success                  *RecordDecisionTaskStartedResponse
	BadRequestError          *BadRequestError
	InternalServiceError     *InternalServiceError
	EventAlreadyStartedError *EventAlreadyStartedError
	EntityNotExistError      *EntityNotExistsError
	ShardOwnershipLostError  *ShardOwnershipLostError
	DomainNotActiveError     *DomainNotActiveError
	LimitExceededError       *LimitExceededError
	ServiceBusyError         *ServiceBusyError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_RecordDecisionTaskStarted_Result) GetSuccess() (o *RecordDecisionTaskStartedResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_RecordDecisionTaskStarted_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_RecordDecisionTaskStarted_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEventAlreadyStartedError is an internal getter (TBD...)
func (v *HistoryService_RecordDecisionTaskStarted_Result) GetEventAlreadyStartedError() (o *EventAlreadyStartedError) {
	if v != nil && v.EventAlreadyStartedError != nil {
		return v.EventAlreadyStartedError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_RecordDecisionTaskStarted_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_RecordDecisionTaskStarted_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_RecordDecisionTaskStarted_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_RecordDecisionTaskStarted_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_RecordDecisionTaskStarted_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_RefreshWorkflowTasks_Args is an internal type (TBD...)
type HistoryService_RefreshWorkflowTasks_Args struct {
	Request *HistoryRefreshWorkflowTasksRequest
}

// GetRequest is an internal getter (TBD...)
func (v *HistoryService_RefreshWorkflowTasks_Args) GetRequest() (o *HistoryRefreshWorkflowTasksRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// HistoryService_RefreshWorkflowTasks_Result is an internal type (TBD...)
type HistoryService_RefreshWorkflowTasks_Result struct {
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	DomainNotActiveError    *DomainNotActiveError
	ShardOwnershipLostError *ShardOwnershipLostError
	ServiceBusyError        *ServiceBusyError
	EntityNotExistError     *EntityNotExistsError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_RefreshWorkflowTasks_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_RefreshWorkflowTasks_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_RefreshWorkflowTasks_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_RefreshWorkflowTasks_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_RefreshWorkflowTasks_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_RefreshWorkflowTasks_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// HistoryService_RemoveSignalMutableState_Args is an internal type (TBD...)
type HistoryService_RemoveSignalMutableState_Args struct {
	RemoveRequest *RemoveSignalMutableStateRequest
}

// GetRemoveRequest is an internal getter (TBD...)
func (v *HistoryService_RemoveSignalMutableState_Args) GetRemoveRequest() (o *RemoveSignalMutableStateRequest) {
	if v != nil && v.RemoveRequest != nil {
		return v.RemoveRequest
	}
	return
}

// HistoryService_RemoveSignalMutableState_Result is an internal type (TBD...)
type HistoryService_RemoveSignalMutableState_Result struct {
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
	DomainNotActiveError    *DomainNotActiveError
	LimitExceededError      *LimitExceededError
	ServiceBusyError        *ServiceBusyError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_RemoveSignalMutableState_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_RemoveSignalMutableState_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_RemoveSignalMutableState_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_RemoveSignalMutableState_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_RemoveSignalMutableState_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_RemoveSignalMutableState_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_RemoveSignalMutableState_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_RemoveTask_Args is an internal type (TBD...)
type HistoryService_RemoveTask_Args struct {
	Request *RemoveTaskRequest
}

// GetRequest is an internal getter (TBD...)
func (v *HistoryService_RemoveTask_Args) GetRequest() (o *RemoveTaskRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// HistoryService_RemoveTask_Result is an internal type (TBD...)
type HistoryService_RemoveTask_Result struct {
	BadRequestError      *BadRequestError
	InternalServiceError *InternalServiceError
	AccessDeniedError    *AccessDeniedError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_RemoveTask_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_RemoveTask_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetAccessDeniedError is an internal getter (TBD...)
func (v *HistoryService_RemoveTask_Result) GetAccessDeniedError() (o *AccessDeniedError) {
	if v != nil && v.AccessDeniedError != nil {
		return v.AccessDeniedError
	}
	return
}

// HistoryService_ReplicateEventsV2_Args is an internal type (TBD...)
type HistoryService_ReplicateEventsV2_Args struct {
	ReplicateV2Request *ReplicateEventsV2Request
}

// GetReplicateV2Request is an internal getter (TBD...)
func (v *HistoryService_ReplicateEventsV2_Args) GetReplicateV2Request() (o *ReplicateEventsV2Request) {
	if v != nil && v.ReplicateV2Request != nil {
		return v.ReplicateV2Request
	}
	return
}

// HistoryService_ReplicateEventsV2_Result is an internal type (TBD...)
type HistoryService_ReplicateEventsV2_Result struct {
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
	LimitExceededError      *LimitExceededError
	RetryTaskError          *RetryTaskV2Error
	ServiceBusyError        *ServiceBusyError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_ReplicateEventsV2_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_ReplicateEventsV2_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_ReplicateEventsV2_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_ReplicateEventsV2_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_ReplicateEventsV2_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetRetryTaskError is an internal getter (TBD...)
func (v *HistoryService_ReplicateEventsV2_Result) GetRetryTaskError() (o *RetryTaskV2Error) {
	if v != nil && v.RetryTaskError != nil {
		return v.RetryTaskError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_ReplicateEventsV2_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_RequestCancelWorkflowExecution_Args is an internal type (TBD...)
type HistoryService_RequestCancelWorkflowExecution_Args struct {
	CancelRequest *HistoryRequestCancelWorkflowExecutionRequest
}

// GetCancelRequest is an internal getter (TBD...)
func (v *HistoryService_RequestCancelWorkflowExecution_Args) GetCancelRequest() (o *HistoryRequestCancelWorkflowExecutionRequest) {
	if v != nil && v.CancelRequest != nil {
		return v.CancelRequest
	}
	return
}

// HistoryService_RequestCancelWorkflowExecution_Result is an internal type (TBD...)
type HistoryService_RequestCancelWorkflowExecution_Result struct {
	BadRequestError                   *BadRequestError
	InternalServiceError              *InternalServiceError
	EntityNotExistError               *EntityNotExistsError
	ShardOwnershipLostError           *ShardOwnershipLostError
	CancellationAlreadyRequestedError *CancellationAlreadyRequestedError
	DomainNotActiveError              *DomainNotActiveError
	LimitExceededError                *LimitExceededError
	ServiceBusyError                  *ServiceBusyError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_RequestCancelWorkflowExecution_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_RequestCancelWorkflowExecution_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_RequestCancelWorkflowExecution_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_RequestCancelWorkflowExecution_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetCancellationAlreadyRequestedError is an internal getter (TBD...)
func (v *HistoryService_RequestCancelWorkflowExecution_Result) GetCancellationAlreadyRequestedError() (o *CancellationAlreadyRequestedError) {
	if v != nil && v.CancellationAlreadyRequestedError != nil {
		return v.CancellationAlreadyRequestedError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_RequestCancelWorkflowExecution_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_RequestCancelWorkflowExecution_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_RequestCancelWorkflowExecution_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_ResetQueue_Args is an internal type (TBD...)
type HistoryService_ResetQueue_Args struct {
	Request *ResetQueueRequest
}

// GetRequest is an internal getter (TBD...)
func (v *HistoryService_ResetQueue_Args) GetRequest() (o *ResetQueueRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// HistoryService_ResetQueue_Result is an internal type (TBD...)
type HistoryService_ResetQueue_Result struct {
	BadRequestError      *BadRequestError
	InternalServiceError *InternalServiceError
	AccessDeniedError    *AccessDeniedError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_ResetQueue_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_ResetQueue_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetAccessDeniedError is an internal getter (TBD...)
func (v *HistoryService_ResetQueue_Result) GetAccessDeniedError() (o *AccessDeniedError) {
	if v != nil && v.AccessDeniedError != nil {
		return v.AccessDeniedError
	}
	return
}

// HistoryService_ResetStickyTaskList_Args is an internal type (TBD...)
type HistoryService_ResetStickyTaskList_Args struct {
	ResetRequest *HistoryResetStickyTaskListRequest
}

// GetResetRequest is an internal getter (TBD...)
func (v *HistoryService_ResetStickyTaskList_Args) GetResetRequest() (o *HistoryResetStickyTaskListRequest) {
	if v != nil && v.ResetRequest != nil {
		return v.ResetRequest
	}
	return
}

// HistoryService_ResetStickyTaskList_Result is an internal type (TBD...)
type HistoryService_ResetStickyTaskList_Result struct {
	Success                 *HistoryResetStickyTaskListResponse
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
	LimitExceededError      *LimitExceededError
	ServiceBusyError        *ServiceBusyError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_ResetStickyTaskList_Result) GetSuccess() (o *HistoryResetStickyTaskListResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_ResetStickyTaskList_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_ResetStickyTaskList_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_ResetStickyTaskList_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_ResetStickyTaskList_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_ResetStickyTaskList_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_ResetStickyTaskList_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_ResetWorkflowExecution_Args is an internal type (TBD...)
type HistoryService_ResetWorkflowExecution_Args struct {
	ResetRequest *HistoryResetWorkflowExecutionRequest
}

// GetResetRequest is an internal getter (TBD...)
func (v *HistoryService_ResetWorkflowExecution_Args) GetResetRequest() (o *HistoryResetWorkflowExecutionRequest) {
	if v != nil && v.ResetRequest != nil {
		return v.ResetRequest
	}
	return
}

// HistoryService_ResetWorkflowExecution_Result is an internal type (TBD...)
type HistoryService_ResetWorkflowExecution_Result struct {
	Success                 *ResetWorkflowExecutionResponse
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
	DomainNotActiveError    *DomainNotActiveError
	LimitExceededError      *LimitExceededError
	ServiceBusyError        *ServiceBusyError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_ResetWorkflowExecution_Result) GetSuccess() (o *ResetWorkflowExecutionResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_ResetWorkflowExecution_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_ResetWorkflowExecution_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_ResetWorkflowExecution_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_ResetWorkflowExecution_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_ResetWorkflowExecution_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_ResetWorkflowExecution_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_ResetWorkflowExecution_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_RespondActivityTaskCanceled_Args is an internal type (TBD...)
type HistoryService_RespondActivityTaskCanceled_Args struct {
	CanceledRequest *HistoryRespondActivityTaskCanceledRequest
}

// GetCanceledRequest is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskCanceled_Args) GetCanceledRequest() (o *HistoryRespondActivityTaskCanceledRequest) {
	if v != nil && v.CanceledRequest != nil {
		return v.CanceledRequest
	}
	return
}

// HistoryService_RespondActivityTaskCanceled_Result is an internal type (TBD...)
type HistoryService_RespondActivityTaskCanceled_Result struct {
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
	DomainNotActiveError    *DomainNotActiveError
	LimitExceededError      *LimitExceededError
	ServiceBusyError        *ServiceBusyError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskCanceled_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskCanceled_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskCanceled_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskCanceled_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskCanceled_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskCanceled_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskCanceled_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_RespondActivityTaskCompleted_Args is an internal type (TBD...)
type HistoryService_RespondActivityTaskCompleted_Args struct {
	CompleteRequest *HistoryRespondActivityTaskCompletedRequest
}

// GetCompleteRequest is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskCompleted_Args) GetCompleteRequest() (o *HistoryRespondActivityTaskCompletedRequest) {
	if v != nil && v.CompleteRequest != nil {
		return v.CompleteRequest
	}
	return
}

// HistoryService_RespondActivityTaskCompleted_Result is an internal type (TBD...)
type HistoryService_RespondActivityTaskCompleted_Result struct {
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
	DomainNotActiveError    *DomainNotActiveError
	LimitExceededError      *LimitExceededError
	ServiceBusyError        *ServiceBusyError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskCompleted_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskCompleted_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskCompleted_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskCompleted_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskCompleted_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskCompleted_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskCompleted_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_RespondActivityTaskFailed_Args is an internal type (TBD...)
type HistoryService_RespondActivityTaskFailed_Args struct {
	FailRequest *HistoryRespondActivityTaskFailedRequest
}

// GetFailRequest is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskFailed_Args) GetFailRequest() (o *HistoryRespondActivityTaskFailedRequest) {
	if v != nil && v.FailRequest != nil {
		return v.FailRequest
	}
	return
}

// HistoryService_RespondActivityTaskFailed_Result is an internal type (TBD...)
type HistoryService_RespondActivityTaskFailed_Result struct {
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
	DomainNotActiveError    *DomainNotActiveError
	LimitExceededError      *LimitExceededError
	ServiceBusyError        *ServiceBusyError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskFailed_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskFailed_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskFailed_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskFailed_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskFailed_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskFailed_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_RespondActivityTaskFailed_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_RespondDecisionTaskCompleted_Args is an internal type (TBD...)
type HistoryService_RespondDecisionTaskCompleted_Args struct {
	CompleteRequest *HistoryRespondDecisionTaskCompletedRequest
}

// GetCompleteRequest is an internal getter (TBD...)
func (v *HistoryService_RespondDecisionTaskCompleted_Args) GetCompleteRequest() (o *HistoryRespondDecisionTaskCompletedRequest) {
	if v != nil && v.CompleteRequest != nil {
		return v.CompleteRequest
	}
	return
}

// HistoryService_RespondDecisionTaskCompleted_Result is an internal type (TBD...)
type HistoryService_RespondDecisionTaskCompleted_Result struct {
	Success                 *HistoryRespondDecisionTaskCompletedResponse
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
	DomainNotActiveError    *DomainNotActiveError
	LimitExceededError      *LimitExceededError
	ServiceBusyError        *ServiceBusyError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_RespondDecisionTaskCompleted_Result) GetSuccess() (o *HistoryRespondDecisionTaskCompletedResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_RespondDecisionTaskCompleted_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_RespondDecisionTaskCompleted_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_RespondDecisionTaskCompleted_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_RespondDecisionTaskCompleted_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_RespondDecisionTaskCompleted_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_RespondDecisionTaskCompleted_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_RespondDecisionTaskCompleted_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_RespondDecisionTaskFailed_Args is an internal type (TBD...)
type HistoryService_RespondDecisionTaskFailed_Args struct {
	FailedRequest *HistoryRespondDecisionTaskFailedRequest
}

// GetFailedRequest is an internal getter (TBD...)
func (v *HistoryService_RespondDecisionTaskFailed_Args) GetFailedRequest() (o *HistoryRespondDecisionTaskFailedRequest) {
	if v != nil && v.FailedRequest != nil {
		return v.FailedRequest
	}
	return
}

// HistoryService_RespondDecisionTaskFailed_Result is an internal type (TBD...)
type HistoryService_RespondDecisionTaskFailed_Result struct {
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
	DomainNotActiveError    *DomainNotActiveError
	LimitExceededError      *LimitExceededError
	ServiceBusyError        *ServiceBusyError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_RespondDecisionTaskFailed_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_RespondDecisionTaskFailed_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_RespondDecisionTaskFailed_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_RespondDecisionTaskFailed_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_RespondDecisionTaskFailed_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_RespondDecisionTaskFailed_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_RespondDecisionTaskFailed_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_ScheduleDecisionTask_Args is an internal type (TBD...)
type HistoryService_ScheduleDecisionTask_Args struct {
	ScheduleRequest *ScheduleDecisionTaskRequest
}

// GetScheduleRequest is an internal getter (TBD...)
func (v *HistoryService_ScheduleDecisionTask_Args) GetScheduleRequest() (o *ScheduleDecisionTaskRequest) {
	if v != nil && v.ScheduleRequest != nil {
		return v.ScheduleRequest
	}
	return
}

// HistoryService_ScheduleDecisionTask_Result is an internal type (TBD...)
type HistoryService_ScheduleDecisionTask_Result struct {
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
	DomainNotActiveError    *DomainNotActiveError
	LimitExceededError      *LimitExceededError
	ServiceBusyError        *ServiceBusyError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_ScheduleDecisionTask_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_ScheduleDecisionTask_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_ScheduleDecisionTask_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_ScheduleDecisionTask_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_ScheduleDecisionTask_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_ScheduleDecisionTask_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_ScheduleDecisionTask_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_SignalWithStartWorkflowExecution_Args is an internal type (TBD...)
type HistoryService_SignalWithStartWorkflowExecution_Args struct {
	SignalWithStartRequest *HistorySignalWithStartWorkflowExecutionRequest
}

// GetSignalWithStartRequest is an internal getter (TBD...)
func (v *HistoryService_SignalWithStartWorkflowExecution_Args) GetSignalWithStartRequest() (o *HistorySignalWithStartWorkflowExecutionRequest) {
	if v != nil && v.SignalWithStartRequest != nil {
		return v.SignalWithStartRequest
	}
	return
}

// HistoryService_SignalWithStartWorkflowExecution_Result is an internal type (TBD...)
type HistoryService_SignalWithStartWorkflowExecution_Result struct {
	Success                     *StartWorkflowExecutionResponse
	BadRequestError             *BadRequestError
	InternalServiceError        *InternalServiceError
	ShardOwnershipLostError     *ShardOwnershipLostError
	DomainNotActiveError        *DomainNotActiveError
	LimitExceededError          *LimitExceededError
	ServiceBusyError            *ServiceBusyError
	WorkflowAlreadyStartedError *WorkflowExecutionAlreadyStartedError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) GetSuccess() (o *StartWorkflowExecutionResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetWorkflowAlreadyStartedError is an internal getter (TBD...)
func (v *HistoryService_SignalWithStartWorkflowExecution_Result) GetWorkflowAlreadyStartedError() (o *WorkflowExecutionAlreadyStartedError) {
	if v != nil && v.WorkflowAlreadyStartedError != nil {
		return v.WorkflowAlreadyStartedError
	}
	return
}

// HistoryService_SignalWorkflowExecution_Args is an internal type (TBD...)
type HistoryService_SignalWorkflowExecution_Args struct {
	SignalRequest *HistorySignalWorkflowExecutionRequest
}

// GetSignalRequest is an internal getter (TBD...)
func (v *HistoryService_SignalWorkflowExecution_Args) GetSignalRequest() (o *HistorySignalWorkflowExecutionRequest) {
	if v != nil && v.SignalRequest != nil {
		return v.SignalRequest
	}
	return
}

// HistoryService_SignalWorkflowExecution_Result is an internal type (TBD...)
type HistoryService_SignalWorkflowExecution_Result struct {
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
	DomainNotActiveError    *DomainNotActiveError
	ServiceBusyError        *ServiceBusyError
	LimitExceededError      *LimitExceededError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_SignalWorkflowExecution_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_SignalWorkflowExecution_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_SignalWorkflowExecution_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_SignalWorkflowExecution_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_SignalWorkflowExecution_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_SignalWorkflowExecution_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_SignalWorkflowExecution_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// HistoryService_StartWorkflowExecution_Args is an internal type (TBD...)
type HistoryService_StartWorkflowExecution_Args struct {
	StartRequest *HistoryStartWorkflowExecutionRequest
}

// GetStartRequest is an internal getter (TBD...)
func (v *HistoryService_StartWorkflowExecution_Args) GetStartRequest() (o *HistoryStartWorkflowExecutionRequest) {
	if v != nil && v.StartRequest != nil {
		return v.StartRequest
	}
	return
}

// HistoryService_StartWorkflowExecution_Result is an internal type (TBD...)
type HistoryService_StartWorkflowExecution_Result struct {
	Success                  *StartWorkflowExecutionResponse
	BadRequestError          *BadRequestError
	InternalServiceError     *InternalServiceError
	SessionAlreadyExistError *WorkflowExecutionAlreadyStartedError
	ShardOwnershipLostError  *ShardOwnershipLostError
	DomainNotActiveError     *DomainNotActiveError
	LimitExceededError       *LimitExceededError
	ServiceBusyError         *ServiceBusyError
}

// GetSuccess is an internal getter (TBD...)
func (v *HistoryService_StartWorkflowExecution_Result) GetSuccess() (o *StartWorkflowExecutionResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_StartWorkflowExecution_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_StartWorkflowExecution_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetSessionAlreadyExistError is an internal getter (TBD...)
func (v *HistoryService_StartWorkflowExecution_Result) GetSessionAlreadyExistError() (o *WorkflowExecutionAlreadyStartedError) {
	if v != nil && v.SessionAlreadyExistError != nil {
		return v.SessionAlreadyExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_StartWorkflowExecution_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_StartWorkflowExecution_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_StartWorkflowExecution_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_StartWorkflowExecution_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_SyncActivity_Args is an internal type (TBD...)
type HistoryService_SyncActivity_Args struct {
	SyncActivityRequest *SyncActivityRequest
}

// GetSyncActivityRequest is an internal getter (TBD...)
func (v *HistoryService_SyncActivity_Args) GetSyncActivityRequest() (o *SyncActivityRequest) {
	if v != nil && v.SyncActivityRequest != nil {
		return v.SyncActivityRequest
	}
	return
}

// HistoryService_SyncActivity_Result is an internal type (TBD...)
type HistoryService_SyncActivity_Result struct {
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
	ServiceBusyError        *ServiceBusyError
	RetryTaskV2Error        *RetryTaskV2Error
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_SyncActivity_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_SyncActivity_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_SyncActivity_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_SyncActivity_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_SyncActivity_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetRetryTaskV2Error is an internal getter (TBD...)
func (v *HistoryService_SyncActivity_Result) GetRetryTaskV2Error() (o *RetryTaskV2Error) {
	if v != nil && v.RetryTaskV2Error != nil {
		return v.RetryTaskV2Error
	}
	return
}

// HistoryService_SyncShardStatus_Args is an internal type (TBD...)
type HistoryService_SyncShardStatus_Args struct {
	SyncShardStatusRequest *SyncShardStatusRequest
}

// GetSyncShardStatusRequest is an internal getter (TBD...)
func (v *HistoryService_SyncShardStatus_Args) GetSyncShardStatusRequest() (o *SyncShardStatusRequest) {
	if v != nil && v.SyncShardStatusRequest != nil {
		return v.SyncShardStatusRequest
	}
	return
}

// HistoryService_SyncShardStatus_Result is an internal type (TBD...)
type HistoryService_SyncShardStatus_Result struct {
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	ShardOwnershipLostError *ShardOwnershipLostError
	LimitExceededError      *LimitExceededError
	ServiceBusyError        *ServiceBusyError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_SyncShardStatus_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_SyncShardStatus_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_SyncShardStatus_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_SyncShardStatus_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_SyncShardStatus_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// HistoryService_TerminateWorkflowExecution_Args is an internal type (TBD...)
type HistoryService_TerminateWorkflowExecution_Args struct {
	TerminateRequest *HistoryTerminateWorkflowExecutionRequest
}

// GetTerminateRequest is an internal getter (TBD...)
func (v *HistoryService_TerminateWorkflowExecution_Args) GetTerminateRequest() (o *HistoryTerminateWorkflowExecutionRequest) {
	if v != nil && v.TerminateRequest != nil {
		return v.TerminateRequest
	}
	return
}

// HistoryService_TerminateWorkflowExecution_Result is an internal type (TBD...)
type HistoryService_TerminateWorkflowExecution_Result struct {
	BadRequestError         *BadRequestError
	InternalServiceError    *InternalServiceError
	EntityNotExistError     *EntityNotExistsError
	ShardOwnershipLostError *ShardOwnershipLostError
	DomainNotActiveError    *DomainNotActiveError
	LimitExceededError      *LimitExceededError
	ServiceBusyError        *ServiceBusyError
}

// GetBadRequestError is an internal getter (TBD...)
func (v *HistoryService_TerminateWorkflowExecution_Result) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *HistoryService_TerminateWorkflowExecution_Result) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *HistoryService_TerminateWorkflowExecution_Result) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetShardOwnershipLostError is an internal getter (TBD...)
func (v *HistoryService_TerminateWorkflowExecution_Result) GetShardOwnershipLostError() (o *ShardOwnershipLostError) {
	if v != nil && v.ShardOwnershipLostError != nil {
		return v.ShardOwnershipLostError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *HistoryService_TerminateWorkflowExecution_Result) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *HistoryService_TerminateWorkflowExecution_Result) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *HistoryService_TerminateWorkflowExecution_Result) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
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
	DomainUUID  *string
	Domain      *string
	Execution   *WorkflowExecution
	InitiatedID *int64
}

// GetDomainUUID is an internal getter (TBD...)
func (v *ParentExecutionInfo) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
	}
	return
}

// GetDomain is an internal getter (TBD...)
func (v *ParentExecutionInfo) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
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
	if v != nil && v.InitiatedID != nil {
		return *v.InitiatedID
	}
	return
}

// PollMutableStateRequest is an internal type (TBD...)
type PollMutableStateRequest struct {
	DomainUUID          *string
	Execution           *WorkflowExecution
	ExpectedNextEventID *int64
	CurrentBranchToken  []byte
}

// GetDomainUUID is an internal getter (TBD...)
func (v *PollMutableStateRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	if v != nil && v.ExpectedNextEventID != nil {
		return *v.ExpectedNextEventID
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
	Execution                            *WorkflowExecution
	WorkflowType                         *WorkflowType
	NextEventID                          *int64
	PreviousStartedEventID               *int64
	LastFirstEventID                     *int64
	TaskList                             *TaskList
	StickyTaskList                       *TaskList
	ClientLibraryVersion                 *string
	ClientFeatureVersion                 *string
	ClientImpl                           *string
	StickyTaskListScheduleToStartTimeout *int32
	CurrentBranchToken                   []byte
	VersionHistories                     *VersionHistories
	WorkflowState                        *int32
	WorkflowCloseState                   *int32
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
	if v != nil && v.NextEventID != nil {
		return *v.NextEventID
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
	if v != nil && v.LastFirstEventID != nil {
		return *v.LastFirstEventID
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
	if v != nil && v.ClientLibraryVersion != nil {
		return *v.ClientLibraryVersion
	}
	return
}

// GetClientFeatureVersion is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetClientFeatureVersion() (o string) {
	if v != nil && v.ClientFeatureVersion != nil {
		return *v.ClientFeatureVersion
	}
	return
}

// GetClientImpl is an internal getter (TBD...)
func (v *PollMutableStateResponse) GetClientImpl() (o string) {
	if v != nil && v.ClientImpl != nil {
		return *v.ClientImpl
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

// GetStatesByCluster is an internal getter (TBD...)
func (v *ProcessingQueueStates) GetStatesByCluster() (o map[string][]*ProcessingQueueState) {
	if v != nil && v.StatesByCluster != nil {
		return v.StatesByCluster
	}
	return
}

// HistoryQueryWorkflowRequest is an internal type (TBD...)
type HistoryQueryWorkflowRequest struct {
	DomainUUID *string
	Request    *QueryWorkflowRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryQueryWorkflowRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	DomainUUID *string
	Request    *ReapplyEventsRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryReapplyEventsRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	DomainUUID       *string
	HeartbeatRequest *RecordActivityTaskHeartbeatRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRecordActivityTaskHeartbeatRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	DomainUUID        *string
	WorkflowExecution *WorkflowExecution
	ScheduleID        *int64
	TaskID            *int64
	RequestID         *string
	PollRequest       *PollForActivityTaskRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *RecordActivityTaskStartedRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	if v != nil && v.ScheduleID != nil {
		return *v.ScheduleID
	}
	return
}

// GetTaskID is an internal getter (TBD...)
func (v *RecordActivityTaskStartedRequest) GetTaskID() (o int64) {
	if v != nil && v.TaskID != nil {
		return *v.TaskID
	}
	return
}

// GetRequestID is an internal getter (TBD...)
func (v *RecordActivityTaskStartedRequest) GetRequestID() (o string) {
	if v != nil && v.RequestID != nil {
		return *v.RequestID
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
	ScheduledEvent                  *HistoryEvent
	StartedTimestamp                *int64
	Attempt                         *int64
	ScheduledTimestampOfThisAttempt *int64
	HeartbeatDetails                []byte
	WorkflowType                    *WorkflowType
	WorkflowDomain                  *string
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
	if v != nil && v.Attempt != nil {
		return *v.Attempt
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
	if v != nil && v.WorkflowDomain != nil {
		return *v.WorkflowDomain
	}
	return
}

// RecordChildExecutionCompletedRequest is an internal type (TBD...)
type RecordChildExecutionCompletedRequest struct {
	DomainUUID         *string
	WorkflowExecution  *WorkflowExecution
	InitiatedID        *int64
	CompletedExecution *WorkflowExecution
	CompletionEvent    *HistoryEvent
}

// GetDomainUUID is an internal getter (TBD...)
func (v *RecordChildExecutionCompletedRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	if v != nil && v.InitiatedID != nil {
		return *v.InitiatedID
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
	DomainUUID        *string
	WorkflowExecution *WorkflowExecution
	ScheduleID        *int64
	TaskID            *int64
	RequestID         *string
	PollRequest       *PollForDecisionTaskRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	if v != nil && v.ScheduleID != nil {
		return *v.ScheduleID
	}
	return
}

// GetTaskID is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedRequest) GetTaskID() (o int64) {
	if v != nil && v.TaskID != nil {
		return *v.TaskID
	}
	return
}

// GetRequestID is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedRequest) GetRequestID() (o string) {
	if v != nil && v.RequestID != nil {
		return *v.RequestID
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
	WorkflowType              *WorkflowType
	PreviousStartedEventID    *int64
	ScheduledEventID          *int64
	StartedEventID            *int64
	NextEventID               *int64
	Attempt                   *int64
	StickyExecutionEnabled    *bool
	DecisionInfo              *TransientDecisionInfo
	WorkflowExecutionTaskList *TaskList
	EventStoreVersion         *int32
	BranchToken               []byte
	ScheduledTimestamp        *int64
	StartedTimestamp          *int64
	Queries                   map[string]*WorkflowQuery
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
	if v != nil && v.ScheduledEventID != nil {
		return *v.ScheduledEventID
	}
	return
}

// GetStartedEventID is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetStartedEventID() (o int64) {
	if v != nil && v.StartedEventID != nil {
		return *v.StartedEventID
	}
	return
}

// GetNextEventID is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetNextEventID() (o int64) {
	if v != nil && v.NextEventID != nil {
		return *v.NextEventID
	}
	return
}

// GetAttempt is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetAttempt() (o int64) {
	if v != nil && v.Attempt != nil {
		return *v.Attempt
	}
	return
}

// GetStickyExecutionEnabled is an internal getter (TBD...)
func (v *RecordDecisionTaskStartedResponse) GetStickyExecutionEnabled() (o bool) {
	if v != nil && v.StickyExecutionEnabled != nil {
		return *v.StickyExecutionEnabled
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
	if v != nil && v.EventStoreVersion != nil {
		return *v.EventStoreVersion
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
	DomainUIID *string
	Request    *RefreshWorkflowTasksRequest
}

// GetDomainUIID is an internal getter (TBD...)
func (v *HistoryRefreshWorkflowTasksRequest) GetDomainUIID() (o string) {
	if v != nil && v.DomainUIID != nil {
		return *v.DomainUIID
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
	DomainUUID        *string
	WorkflowExecution *WorkflowExecution
	RequestID         *string
}

// GetDomainUUID is an internal getter (TBD...)
func (v *RemoveSignalMutableStateRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	if v != nil && v.RequestID != nil {
		return *v.RequestID
	}
	return
}

// ReplicateEventsV2Request is an internal type (TBD...)
type ReplicateEventsV2Request struct {
	DomainUUID          *string
	WorkflowExecution   *WorkflowExecution
	VersionHistoryItems []*VersionHistoryItem
	Events              *DataBlob
	NewRunEvents        *DataBlob
}

// GetDomainUUID is an internal getter (TBD...)
func (v *ReplicateEventsV2Request) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	DomainUUID                *string
	CancelRequest             *RequestCancelWorkflowExecutionRequest
	ExternalInitiatedEventID  *int64
	ExternalWorkflowExecution *WorkflowExecution
	ChildWorkflowOnly         *bool
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRequestCancelWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	if v != nil && v.ChildWorkflowOnly != nil {
		return *v.ChildWorkflowOnly
	}
	return
}

// HistoryResetStickyTaskListRequest is an internal type (TBD...)
type HistoryResetStickyTaskListRequest struct {
	DomainUUID *string
	Execution  *WorkflowExecution
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryResetStickyTaskListRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	DomainUUID   *string
	ResetRequest *ResetWorkflowExecutionRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryResetWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	DomainUUID    *string
	CancelRequest *RespondActivityTaskCanceledRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRespondActivityTaskCanceledRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	DomainUUID      *string
	CompleteRequest *RespondActivityTaskCompletedRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRespondActivityTaskCompletedRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	DomainUUID    *string
	FailedRequest *RespondActivityTaskFailedRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRespondActivityTaskFailedRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	DomainUUID      *string
	CompleteRequest *RespondDecisionTaskCompletedRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRespondDecisionTaskCompletedRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	DomainUUID    *string
	FailedRequest *RespondDecisionTaskFailedRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryRespondDecisionTaskFailedRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	DomainUUID        *string
	WorkflowExecution *WorkflowExecution
	IsFirstDecision   *bool
}

// GetDomainUUID is an internal getter (TBD...)
func (v *ScheduleDecisionTaskRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	if v != nil && v.IsFirstDecision != nil {
		return *v.IsFirstDecision
	}
	return
}

// ShardOwnershipLostError is an internal type (TBD...)
type ShardOwnershipLostError struct {
	Message *string
	Owner   *string
}

// GetMessage is an internal getter (TBD...)
func (v *ShardOwnershipLostError) GetMessage() (o string) {
	if v != nil && v.Message != nil {
		return *v.Message
	}
	return
}

// GetOwner is an internal getter (TBD...)
func (v *ShardOwnershipLostError) GetOwner() (o string) {
	if v != nil && v.Owner != nil {
		return *v.Owner
	}
	return
}

// HistorySignalWithStartWorkflowExecutionRequest is an internal type (TBD...)
type HistorySignalWithStartWorkflowExecutionRequest struct {
	DomainUUID             *string
	SignalWithStartRequest *SignalWithStartWorkflowExecutionRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistorySignalWithStartWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	DomainUUID                *string
	SignalRequest             *SignalWorkflowExecutionRequest
	ExternalWorkflowExecution *WorkflowExecution
	ChildWorkflowOnly         *bool
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistorySignalWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	if v != nil && v.ChildWorkflowOnly != nil {
		return *v.ChildWorkflowOnly
	}
	return
}

// HistoryStartWorkflowExecutionRequest is an internal type (TBD...)
type HistoryStartWorkflowExecutionRequest struct {
	DomainUUID                      *string
	StartRequest                    *StartWorkflowExecutionRequest
	ParentExecutionInfo             *ParentExecutionInfo
	Attempt                         *int32
	ExpirationTimestamp             *int64
	ContinueAsNewInitiator          *ContinueAsNewInitiator
	ContinuedFailureReason          *string
	ContinuedFailureDetails         []byte
	LastCompletionResult            []byte
	FirstDecisionTaskBackoffSeconds *int32
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryStartWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
	if v != nil && v.Attempt != nil {
		return *v.Attempt
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
	DomainID           *string
	WorkflowID         *string
	RunID              *string
	Version            *int64
	ScheduledID        *int64
	ScheduledTime      *int64
	StartedID          *int64
	StartedTime        *int64
	LastHeartbeatTime  *int64
	Details            []byte
	Attempt            *int32
	LastFailureReason  *string
	LastWorkerIdentity *string
	LastFailureDetails []byte
	VersionHistory     *VersionHistory
}

// GetDomainID is an internal getter (TBD...)
func (v *SyncActivityRequest) GetDomainID() (o string) {
	if v != nil && v.DomainID != nil {
		return *v.DomainID
	}
	return
}

// GetWorkflowID is an internal getter (TBD...)
func (v *SyncActivityRequest) GetWorkflowID() (o string) {
	if v != nil && v.WorkflowID != nil {
		return *v.WorkflowID
	}
	return
}

// GetRunID is an internal getter (TBD...)
func (v *SyncActivityRequest) GetRunID() (o string) {
	if v != nil && v.RunID != nil {
		return *v.RunID
	}
	return
}

// GetVersion is an internal getter (TBD...)
func (v *SyncActivityRequest) GetVersion() (o int64) {
	if v != nil && v.Version != nil {
		return *v.Version
	}
	return
}

// GetScheduledID is an internal getter (TBD...)
func (v *SyncActivityRequest) GetScheduledID() (o int64) {
	if v != nil && v.ScheduledID != nil {
		return *v.ScheduledID
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
	if v != nil && v.StartedID != nil {
		return *v.StartedID
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
	if v != nil && v.Attempt != nil {
		return *v.Attempt
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
	if v != nil && v.LastWorkerIdentity != nil {
		return *v.LastWorkerIdentity
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
	SourceCluster *string
	ShardID       *int64
	Timestamp     *int64
}

// GetSourceCluster is an internal getter (TBD...)
func (v *SyncShardStatusRequest) GetSourceCluster() (o string) {
	if v != nil && v.SourceCluster != nil {
		return *v.SourceCluster
	}
	return
}

// GetShardID is an internal getter (TBD...)
func (v *SyncShardStatusRequest) GetShardID() (o int64) {
	if v != nil && v.ShardID != nil {
		return *v.ShardID
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
	DomainUUID       *string
	TerminateRequest *TerminateWorkflowExecutionRequest
}

// GetDomainUUID is an internal getter (TBD...)
func (v *HistoryTerminateWorkflowExecutionRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
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
