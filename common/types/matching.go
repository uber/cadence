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

import (
	"fmt"
	"strconv"
	"strings"
)

// AddActivityTaskRequest is an internal type (TBD...)
type AddActivityTaskRequest struct {
	DomainUUID                    *string            `json:"domainUUID,omitempty"`
	Execution                     *WorkflowExecution `json:"execution,omitempty"`
	SourceDomainUUID              *string            `json:"sourceDomainUUID,omitempty"`
	TaskList                      *TaskList          `json:"taskList,omitempty"`
	ScheduleID                    *int64             `json:"scheduleId,omitempty"`
	ScheduleToStartTimeoutSeconds *int32             `json:"scheduleToStartTimeoutSeconds,omitempty"`
	Source                        *TaskSource        `json:"source,omitempty"`
	ForwardedFrom                 *string            `json:"forwardedFrom,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *AddActivityTaskRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
	}
	return
}

// GetExecution is an internal getter (TBD...)
func (v *AddActivityTaskRequest) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
}

// GetSourceDomainUUID is an internal getter (TBD...)
func (v *AddActivityTaskRequest) GetSourceDomainUUID() (o string) {
	if v != nil && v.SourceDomainUUID != nil {
		return *v.SourceDomainUUID
	}
	return
}

// GetTaskList is an internal getter (TBD...)
func (v *AddActivityTaskRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

// GetScheduleID is an internal getter (TBD...)
func (v *AddActivityTaskRequest) GetScheduleID() (o int64) {
	if v != nil && v.ScheduleID != nil {
		return *v.ScheduleID
	}
	return
}

// GetScheduleToStartTimeoutSeconds is an internal getter (TBD...)
func (v *AddActivityTaskRequest) GetScheduleToStartTimeoutSeconds() (o int32) {
	if v != nil && v.ScheduleToStartTimeoutSeconds != nil {
		return *v.ScheduleToStartTimeoutSeconds
	}
	return
}

// GetSource is an internal getter (TBD...)
func (v *AddActivityTaskRequest) GetSource() (o TaskSource) {
	if v != nil && v.Source != nil {
		return *v.Source
	}
	return
}

// GetForwardedFrom is an internal getter (TBD...)
func (v *AddActivityTaskRequest) GetForwardedFrom() (o string) {
	if v != nil && v.ForwardedFrom != nil {
		return *v.ForwardedFrom
	}
	return
}

// AddDecisionTaskRequest is an internal type (TBD...)
type AddDecisionTaskRequest struct {
	DomainUUID                    *string            `json:"domainUUID,omitempty"`
	Execution                     *WorkflowExecution `json:"execution,omitempty"`
	TaskList                      *TaskList          `json:"taskList,omitempty"`
	ScheduleID                    *int64             `json:"scheduleId,omitempty"`
	ScheduleToStartTimeoutSeconds *int32             `json:"scheduleToStartTimeoutSeconds,omitempty"`
	Source                        *TaskSource        `json:"source,omitempty"`
	ForwardedFrom                 *string            `json:"forwardedFrom,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *AddDecisionTaskRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
	}
	return
}

// GetExecution is an internal getter (TBD...)
func (v *AddDecisionTaskRequest) GetExecution() (o *WorkflowExecution) {
	if v != nil && v.Execution != nil {
		return v.Execution
	}
	return
}

// GetTaskList is an internal getter (TBD...)
func (v *AddDecisionTaskRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

// GetScheduleID is an internal getter (TBD...)
func (v *AddDecisionTaskRequest) GetScheduleID() (o int64) {
	if v != nil && v.ScheduleID != nil {
		return *v.ScheduleID
	}
	return
}

// GetScheduleToStartTimeoutSeconds is an internal getter (TBD...)
func (v *AddDecisionTaskRequest) GetScheduleToStartTimeoutSeconds() (o int32) {
	if v != nil && v.ScheduleToStartTimeoutSeconds != nil {
		return *v.ScheduleToStartTimeoutSeconds
	}
	return
}

// GetSource is an internal getter (TBD...)
func (v *AddDecisionTaskRequest) GetSource() (o TaskSource) {
	if v != nil && v.Source != nil {
		return *v.Source
	}
	return
}

// GetForwardedFrom is an internal getter (TBD...)
func (v *AddDecisionTaskRequest) GetForwardedFrom() (o string) {
	if v != nil && v.ForwardedFrom != nil {
		return *v.ForwardedFrom
	}
	return
}

// CancelOutstandingPollRequest is an internal type (TBD...)
type CancelOutstandingPollRequest struct {
	DomainUUID   *string   `json:"domainUUID,omitempty"`
	TaskListType *int32    `json:"taskListType,omitempty"`
	TaskList     *TaskList `json:"taskList,omitempty"`
	PollerID     *string   `json:"pollerID,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *CancelOutstandingPollRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
	}
	return
}

// GetTaskListType is an internal getter (TBD...)
func (v *CancelOutstandingPollRequest) GetTaskListType() (o int32) {
	if v != nil && v.TaskListType != nil {
		return *v.TaskListType
	}
	return
}

// GetTaskList is an internal getter (TBD...)
func (v *CancelOutstandingPollRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

// GetPollerID is an internal getter (TBD...)
func (v *CancelOutstandingPollRequest) GetPollerID() (o string) {
	if v != nil && v.PollerID != nil {
		return *v.PollerID
	}
	return
}

// MatchingDescribeTaskListRequest is an internal type (TBD...)
type MatchingDescribeTaskListRequest struct {
	DomainUUID  *string                  `json:"domainUUID,omitempty"`
	DescRequest *DescribeTaskListRequest `json:"descRequest,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *MatchingDescribeTaskListRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
	}
	return
}

// GetDescRequest is an internal getter (TBD...)
func (v *MatchingDescribeTaskListRequest) GetDescRequest() (o *DescribeTaskListRequest) {
	if v != nil && v.DescRequest != nil {
		return v.DescRequest
	}
	return
}

// MatchingListTaskListPartitionsRequest is an internal type (TBD...)
type MatchingListTaskListPartitionsRequest struct {
	Domain   *string   `json:"domain,omitempty"`
	TaskList *TaskList `json:"taskList,omitempty"`
}

// GetDomain is an internal getter (TBD...)
func (v *MatchingListTaskListPartitionsRequest) GetDomain() (o string) {
	if v != nil && v.Domain != nil {
		return *v.Domain
	}
	return
}

// GetTaskList is an internal getter (TBD...)
func (v *MatchingListTaskListPartitionsRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

// MatchingServiceAddActivityTaskArgs is an internal type (TBD...)
type MatchingServiceAddActivityTaskArgs struct {
	AddRequest *AddActivityTaskRequest `json:"addRequest,omitempty"`
}

// GetAddRequest is an internal getter (TBD...)
func (v *MatchingServiceAddActivityTaskArgs) GetAddRequest() (o *AddActivityTaskRequest) {
	if v != nil && v.AddRequest != nil {
		return v.AddRequest
	}
	return
}

// MatchingServiceAddActivityTaskResult is an internal type (TBD...)
type MatchingServiceAddActivityTaskResult struct {
	BadRequestError        *BadRequestError        `json:"badRequestError,omitempty"`
	InternalServiceError   *InternalServiceError   `json:"internalServiceError,omitempty"`
	ServiceBusyError       *ServiceBusyError       `json:"serviceBusyError,omitempty"`
	LimitExceededError     *LimitExceededError     `json:"limitExceededError,omitempty"`
	DomainNotActiveError   *DomainNotActiveError   `json:"domainNotActiveError,omitempty"`
	RemoteSyncMatchedError *RemoteSyncMatchedError `json:"remoteSyncMatchedError,omitempty"`
}

// GetBadRequestError is an internal getter (TBD...)
func (v *MatchingServiceAddActivityTaskResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *MatchingServiceAddActivityTaskResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *MatchingServiceAddActivityTaskResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *MatchingServiceAddActivityTaskResult) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *MatchingServiceAddActivityTaskResult) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetRemoteSyncMatchedError is an internal getter (TBD...)
func (v *MatchingServiceAddActivityTaskResult) GetRemoteSyncMatchedError() (o *RemoteSyncMatchedError) {
	if v != nil && v.RemoteSyncMatchedError != nil {
		return v.RemoteSyncMatchedError
	}
	return
}

// MatchingServiceAddDecisionTaskArgs is an internal type (TBD...)
type MatchingServiceAddDecisionTaskArgs struct {
	AddRequest *AddDecisionTaskRequest `json:"addRequest,omitempty"`
}

// GetAddRequest is an internal getter (TBD...)
func (v *MatchingServiceAddDecisionTaskArgs) GetAddRequest() (o *AddDecisionTaskRequest) {
	if v != nil && v.AddRequest != nil {
		return v.AddRequest
	}
	return
}

// MatchingServiceAddDecisionTaskResult is an internal type (TBD...)
type MatchingServiceAddDecisionTaskResult struct {
	BadRequestError        *BadRequestError        `json:"badRequestError,omitempty"`
	InternalServiceError   *InternalServiceError   `json:"internalServiceError,omitempty"`
	ServiceBusyError       *ServiceBusyError       `json:"serviceBusyError,omitempty"`
	LimitExceededError     *LimitExceededError     `json:"limitExceededError,omitempty"`
	DomainNotActiveError   *DomainNotActiveError   `json:"domainNotActiveError,omitempty"`
	RemoteSyncMatchedError *RemoteSyncMatchedError `json:"remoteSyncMatchedError,omitempty"`
}

// GetBadRequestError is an internal getter (TBD...)
func (v *MatchingServiceAddDecisionTaskResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *MatchingServiceAddDecisionTaskResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *MatchingServiceAddDecisionTaskResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *MatchingServiceAddDecisionTaskResult) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetDomainNotActiveError is an internal getter (TBD...)
func (v *MatchingServiceAddDecisionTaskResult) GetDomainNotActiveError() (o *DomainNotActiveError) {
	if v != nil && v.DomainNotActiveError != nil {
		return v.DomainNotActiveError
	}
	return
}

// GetRemoteSyncMatchedError is an internal getter (TBD...)
func (v *MatchingServiceAddDecisionTaskResult) GetRemoteSyncMatchedError() (o *RemoteSyncMatchedError) {
	if v != nil && v.RemoteSyncMatchedError != nil {
		return v.RemoteSyncMatchedError
	}
	return
}

// MatchingServiceCancelOutstandingPollArgs is an internal type (TBD...)
type MatchingServiceCancelOutstandingPollArgs struct {
	Request *CancelOutstandingPollRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *MatchingServiceCancelOutstandingPollArgs) GetRequest() (o *CancelOutstandingPollRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// MatchingServiceCancelOutstandingPollResult is an internal type (TBD...)
type MatchingServiceCancelOutstandingPollResult struct {
	BadRequestError      *BadRequestError      `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError `json:"internalServiceError,omitempty"`
	ServiceBusyError     *ServiceBusyError     `json:"serviceBusyError,omitempty"`
}

// GetBadRequestError is an internal getter (TBD...)
func (v *MatchingServiceCancelOutstandingPollResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *MatchingServiceCancelOutstandingPollResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *MatchingServiceCancelOutstandingPollResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// MatchingServiceDescribeTaskListArgs is an internal type (TBD...)
type MatchingServiceDescribeTaskListArgs struct {
	Request *MatchingDescribeTaskListRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *MatchingServiceDescribeTaskListArgs) GetRequest() (o *MatchingDescribeTaskListRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// MatchingServiceDescribeTaskListResult is an internal type (TBD...)
type MatchingServiceDescribeTaskListResult struct {
	Success              *DescribeTaskListResponse `json:"success,omitempty"`
	BadRequestError      *BadRequestError          `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError     `json:"internalServiceError,omitempty"`
	EntityNotExistError  *EntityNotExistsError     `json:"entityNotExistError,omitempty"`
	ServiceBusyError     *ServiceBusyError         `json:"serviceBusyError,omitempty"`
}

// GetSuccess is an internal getter (TBD...)
func (v *MatchingServiceDescribeTaskListResult) GetSuccess() (o *DescribeTaskListResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *MatchingServiceDescribeTaskListResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *MatchingServiceDescribeTaskListResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *MatchingServiceDescribeTaskListResult) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *MatchingServiceDescribeTaskListResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// MatchingServiceListTaskListPartitionsArgs is an internal type (TBD...)
type MatchingServiceListTaskListPartitionsArgs struct {
	Request *MatchingListTaskListPartitionsRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *MatchingServiceListTaskListPartitionsArgs) GetRequest() (o *MatchingListTaskListPartitionsRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// MatchingServiceListTaskListPartitionsResult is an internal type (TBD...)
type MatchingServiceListTaskListPartitionsResult struct {
	Success              *ListTaskListPartitionsResponse `json:"success,omitempty"`
	BadRequestError      *BadRequestError                `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError           `json:"internalServiceError,omitempty"`
	ServiceBusyError     *ServiceBusyError               `json:"serviceBusyError,omitempty"`
}

// GetSuccess is an internal getter (TBD...)
func (v *MatchingServiceListTaskListPartitionsResult) GetSuccess() (o *ListTaskListPartitionsResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *MatchingServiceListTaskListPartitionsResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *MatchingServiceListTaskListPartitionsResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *MatchingServiceListTaskListPartitionsResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// MatchingServicePollForActivityTaskArgs is an internal type (TBD...)
type MatchingServicePollForActivityTaskArgs struct {
	PollRequest *MatchingPollForActivityTaskRequest `json:"pollRequest,omitempty"`
}

// GetPollRequest is an internal getter (TBD...)
func (v *MatchingServicePollForActivityTaskArgs) GetPollRequest() (o *MatchingPollForActivityTaskRequest) {
	if v != nil && v.PollRequest != nil {
		return v.PollRequest
	}
	return
}

// MatchingServicePollForActivityTaskResult is an internal type (TBD...)
type MatchingServicePollForActivityTaskResult struct {
	Success              *PollForActivityTaskResponse `json:"success,omitempty"`
	BadRequestError      *BadRequestError             `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError        `json:"internalServiceError,omitempty"`
	LimitExceededError   *LimitExceededError          `json:"limitExceededError,omitempty"`
	ServiceBusyError     *ServiceBusyError            `json:"serviceBusyError,omitempty"`
}

// GetSuccess is an internal getter (TBD...)
func (v *MatchingServicePollForActivityTaskResult) GetSuccess() (o *PollForActivityTaskResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *MatchingServicePollForActivityTaskResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *MatchingServicePollForActivityTaskResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *MatchingServicePollForActivityTaskResult) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *MatchingServicePollForActivityTaskResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// MatchingServicePollForDecisionTaskArgs is an internal type (TBD...)
type MatchingServicePollForDecisionTaskArgs struct {
	PollRequest *MatchingPollForDecisionTaskRequest `json:"pollRequest,omitempty"`
}

// GetPollRequest is an internal getter (TBD...)
func (v *MatchingServicePollForDecisionTaskArgs) GetPollRequest() (o *MatchingPollForDecisionTaskRequest) {
	if v != nil && v.PollRequest != nil {
		return v.PollRequest
	}
	return
}

// MatchingServicePollForDecisionTaskResult is an internal type (TBD...)
type MatchingServicePollForDecisionTaskResult struct {
	Success              *MatchingPollForDecisionTaskResponse `json:"success,omitempty"`
	BadRequestError      *BadRequestError                     `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError                `json:"internalServiceError,omitempty"`
	LimitExceededError   *LimitExceededError                  `json:"limitExceededError,omitempty"`
	ServiceBusyError     *ServiceBusyError                    `json:"serviceBusyError,omitempty"`
}

// GetSuccess is an internal getter (TBD...)
func (v *MatchingServicePollForDecisionTaskResult) GetSuccess() (o *MatchingPollForDecisionTaskResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *MatchingServicePollForDecisionTaskResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *MatchingServicePollForDecisionTaskResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *MatchingServicePollForDecisionTaskResult) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *MatchingServicePollForDecisionTaskResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// MatchingServiceQueryWorkflowArgs is an internal type (TBD...)
type MatchingServiceQueryWorkflowArgs struct {
	QueryRequest *MatchingQueryWorkflowRequest `json:"queryRequest,omitempty"`
}

// GetQueryRequest is an internal getter (TBD...)
func (v *MatchingServiceQueryWorkflowArgs) GetQueryRequest() (o *MatchingQueryWorkflowRequest) {
	if v != nil && v.QueryRequest != nil {
		return v.QueryRequest
	}
	return
}

// MatchingServiceQueryWorkflowResult is an internal type (TBD...)
type MatchingServiceQueryWorkflowResult struct {
	Success              *QueryWorkflowResponse `json:"success,omitempty"`
	BadRequestError      *BadRequestError       `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError  `json:"internalServiceError,omitempty"`
	EntityNotExistError  *EntityNotExistsError  `json:"entityNotExistError,omitempty"`
	QueryFailedError     *QueryFailedError      `json:"queryFailedError,omitempty"`
	LimitExceededError   *LimitExceededError    `json:"limitExceededError,omitempty"`
	ServiceBusyError     *ServiceBusyError      `json:"serviceBusyError,omitempty"`
}

// GetSuccess is an internal getter (TBD...)
func (v *MatchingServiceQueryWorkflowResult) GetSuccess() (o *QueryWorkflowResponse) {
	if v != nil && v.Success != nil {
		return v.Success
	}
	return
}

// GetBadRequestError is an internal getter (TBD...)
func (v *MatchingServiceQueryWorkflowResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *MatchingServiceQueryWorkflowResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *MatchingServiceQueryWorkflowResult) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetQueryFailedError is an internal getter (TBD...)
func (v *MatchingServiceQueryWorkflowResult) GetQueryFailedError() (o *QueryFailedError) {
	if v != nil && v.QueryFailedError != nil {
		return v.QueryFailedError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *MatchingServiceQueryWorkflowResult) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *MatchingServiceQueryWorkflowResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// MatchingServiceRespondQueryTaskCompletedArgs is an internal type (TBD...)
type MatchingServiceRespondQueryTaskCompletedArgs struct {
	Request *MatchingRespondQueryTaskCompletedRequest `json:"request,omitempty"`
}

// GetRequest is an internal getter (TBD...)
func (v *MatchingServiceRespondQueryTaskCompletedArgs) GetRequest() (o *MatchingRespondQueryTaskCompletedRequest) {
	if v != nil && v.Request != nil {
		return v.Request
	}
	return
}

// MatchingServiceRespondQueryTaskCompletedResult is an internal type (TBD...)
type MatchingServiceRespondQueryTaskCompletedResult struct {
	BadRequestError      *BadRequestError      `json:"badRequestError,omitempty"`
	InternalServiceError *InternalServiceError `json:"internalServiceError,omitempty"`
	EntityNotExistError  *EntityNotExistsError `json:"entityNotExistError,omitempty"`
	LimitExceededError   *LimitExceededError   `json:"limitExceededError,omitempty"`
	ServiceBusyError     *ServiceBusyError     `json:"serviceBusyError,omitempty"`
}

// GetBadRequestError is an internal getter (TBD...)
func (v *MatchingServiceRespondQueryTaskCompletedResult) GetBadRequestError() (o *BadRequestError) {
	if v != nil && v.BadRequestError != nil {
		return v.BadRequestError
	}
	return
}

// GetInternalServiceError is an internal getter (TBD...)
func (v *MatchingServiceRespondQueryTaskCompletedResult) GetInternalServiceError() (o *InternalServiceError) {
	if v != nil && v.InternalServiceError != nil {
		return v.InternalServiceError
	}
	return
}

// GetEntityNotExistError is an internal getter (TBD...)
func (v *MatchingServiceRespondQueryTaskCompletedResult) GetEntityNotExistError() (o *EntityNotExistsError) {
	if v != nil && v.EntityNotExistError != nil {
		return v.EntityNotExistError
	}
	return
}

// GetLimitExceededError is an internal getter (TBD...)
func (v *MatchingServiceRespondQueryTaskCompletedResult) GetLimitExceededError() (o *LimitExceededError) {
	if v != nil && v.LimitExceededError != nil {
		return v.LimitExceededError
	}
	return
}

// GetServiceBusyError is an internal getter (TBD...)
func (v *MatchingServiceRespondQueryTaskCompletedResult) GetServiceBusyError() (o *ServiceBusyError) {
	if v != nil && v.ServiceBusyError != nil {
		return v.ServiceBusyError
	}
	return
}

// MatchingPollForActivityTaskRequest is an internal type (TBD...)
type MatchingPollForActivityTaskRequest struct {
	DomainUUID    *string                     `json:"domainUUID,omitempty"`
	PollerID      *string                     `json:"pollerID,omitempty"`
	PollRequest   *PollForActivityTaskRequest `json:"pollRequest,omitempty"`
	ForwardedFrom *string                     `json:"forwardedFrom,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *MatchingPollForActivityTaskRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
	}
	return
}

// GetPollerID is an internal getter (TBD...)
func (v *MatchingPollForActivityTaskRequest) GetPollerID() (o string) {
	if v != nil && v.PollerID != nil {
		return *v.PollerID
	}
	return
}

// GetPollRequest is an internal getter (TBD...)
func (v *MatchingPollForActivityTaskRequest) GetPollRequest() (o *PollForActivityTaskRequest) {
	if v != nil && v.PollRequest != nil {
		return v.PollRequest
	}
	return
}

// GetForwardedFrom is an internal getter (TBD...)
func (v *MatchingPollForActivityTaskRequest) GetForwardedFrom() (o string) {
	if v != nil && v.ForwardedFrom != nil {
		return *v.ForwardedFrom
	}
	return
}

// MatchingPollForDecisionTaskRequest is an internal type (TBD...)
type MatchingPollForDecisionTaskRequest struct {
	DomainUUID    *string                     `json:"domainUUID,omitempty"`
	PollerID      *string                     `json:"pollerID,omitempty"`
	PollRequest   *PollForDecisionTaskRequest `json:"pollRequest,omitempty"`
	ForwardedFrom *string                     `json:"forwardedFrom,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
	}
	return
}

// GetPollerID is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskRequest) GetPollerID() (o string) {
	if v != nil && v.PollerID != nil {
		return *v.PollerID
	}
	return
}

// GetPollRequest is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskRequest) GetPollRequest() (o *PollForDecisionTaskRequest) {
	if v != nil && v.PollRequest != nil {
		return v.PollRequest
	}
	return
}

// GetForwardedFrom is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskRequest) GetForwardedFrom() (o string) {
	if v != nil && v.ForwardedFrom != nil {
		return *v.ForwardedFrom
	}
	return
}

// MatchingPollForDecisionTaskResponse is an internal type (TBD...)
type MatchingPollForDecisionTaskResponse struct {
	TaskToken                 []byte                    `json:"taskToken,omitempty"`
	WorkflowExecution         *WorkflowExecution        `json:"workflowExecution,omitempty"`
	WorkflowType              *WorkflowType             `json:"workflowType,omitempty"`
	PreviousStartedEventID    *int64                    `json:"previousStartedEventId,omitempty"`
	StartedEventID            *int64                    `json:"startedEventId,omitempty"`
	Attempt                   *int64                    `json:"attempt,omitempty"`
	NextEventID               *int64                    `json:"nextEventId,omitempty"`
	BacklogCountHint          *int64                    `json:"backlogCountHint,omitempty"`
	StickyExecutionEnabled    *bool                     `json:"stickyExecutionEnabled,omitempty"`
	Query                     *WorkflowQuery            `json:"query,omitempty"`
	DecisionInfo              *TransientDecisionInfo    `json:"decisionInfo,omitempty"`
	WorkflowExecutionTaskList *TaskList                 `json:"WorkflowExecutionTaskList,omitempty"`
	EventStoreVersion         *int32                    `json:"eventStoreVersion,omitempty"`
	BranchToken               []byte                    `json:"branchToken,omitempty"`
	ScheduledTimestamp        *int64                    `json:"scheduledTimestamp,omitempty"`
	StartedTimestamp          *int64                    `json:"startedTimestamp,omitempty"`
	Queries                   map[string]*WorkflowQuery `json:"queries,omitempty"`
}

// GetTaskToken is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetTaskToken() (o []byte) {
	if v != nil && v.TaskToken != nil {
		return v.TaskToken
	}
	return
}

// GetWorkflowExecution is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetWorkflowExecution() (o *WorkflowExecution) {
	if v != nil && v.WorkflowExecution != nil {
		return v.WorkflowExecution
	}
	return
}

// GetWorkflowType is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetWorkflowType() (o *WorkflowType) {
	if v != nil && v.WorkflowType != nil {
		return v.WorkflowType
	}
	return
}

// GetPreviousStartedEventID is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetPreviousStartedEventID() (o int64) {
	if v != nil && v.PreviousStartedEventID != nil {
		return *v.PreviousStartedEventID
	}
	return
}

// GetStartedEventID is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetStartedEventID() (o int64) {
	if v != nil && v.StartedEventID != nil {
		return *v.StartedEventID
	}
	return
}

// GetAttempt is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetAttempt() (o int64) {
	if v != nil && v.Attempt != nil {
		return *v.Attempt
	}
	return
}

// GetNextEventID is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetNextEventID() (o int64) {
	if v != nil && v.NextEventID != nil {
		return *v.NextEventID
	}
	return
}

// GetBacklogCountHint is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetBacklogCountHint() (o int64) {
	if v != nil && v.BacklogCountHint != nil {
		return *v.BacklogCountHint
	}
	return
}

// GetStickyExecutionEnabled is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetStickyExecutionEnabled() (o bool) {
	if v != nil && v.StickyExecutionEnabled != nil {
		return *v.StickyExecutionEnabled
	}
	return
}

// GetQuery is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetQuery() (o *WorkflowQuery) {
	if v != nil && v.Query != nil {
		return v.Query
	}
	return
}

// GetDecisionInfo is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetDecisionInfo() (o *TransientDecisionInfo) {
	if v != nil && v.DecisionInfo != nil {
		return v.DecisionInfo
	}
	return
}

// GetWorkflowExecutionTaskList is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetWorkflowExecutionTaskList() (o *TaskList) {
	if v != nil && v.WorkflowExecutionTaskList != nil {
		return v.WorkflowExecutionTaskList
	}
	return
}

// GetEventStoreVersion is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetEventStoreVersion() (o int32) {
	if v != nil && v.EventStoreVersion != nil {
		return *v.EventStoreVersion
	}
	return
}

// GetBranchToken is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetBranchToken() (o []byte) {
	if v != nil && v.BranchToken != nil {
		return v.BranchToken
	}
	return
}

// GetScheduledTimestamp is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetScheduledTimestamp() (o int64) {
	if v != nil && v.ScheduledTimestamp != nil {
		return *v.ScheduledTimestamp
	}
	return
}

// GetStartedTimestamp is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetStartedTimestamp() (o int64) {
	if v != nil && v.StartedTimestamp != nil {
		return *v.StartedTimestamp
	}
	return
}

// GetQueries is an internal getter (TBD...)
func (v *MatchingPollForDecisionTaskResponse) GetQueries() (o map[string]*WorkflowQuery) {
	if v != nil && v.Queries != nil {
		return v.Queries
	}
	return
}

// MatchingQueryWorkflowRequest is an internal type (TBD...)
type MatchingQueryWorkflowRequest struct {
	DomainUUID    *string               `json:"domainUUID,omitempty"`
	TaskList      *TaskList             `json:"taskList,omitempty"`
	QueryRequest  *QueryWorkflowRequest `json:"queryRequest,omitempty"`
	ForwardedFrom *string               `json:"forwardedFrom,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *MatchingQueryWorkflowRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
	}
	return
}

// GetTaskList is an internal getter (TBD...)
func (v *MatchingQueryWorkflowRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

// GetQueryRequest is an internal getter (TBD...)
func (v *MatchingQueryWorkflowRequest) GetQueryRequest() (o *QueryWorkflowRequest) {
	if v != nil && v.QueryRequest != nil {
		return v.QueryRequest
	}
	return
}

// GetForwardedFrom is an internal getter (TBD...)
func (v *MatchingQueryWorkflowRequest) GetForwardedFrom() (o string) {
	if v != nil && v.ForwardedFrom != nil {
		return *v.ForwardedFrom
	}
	return
}

// MatchingRespondQueryTaskCompletedRequest is an internal type (TBD...)
type MatchingRespondQueryTaskCompletedRequest struct {
	DomainUUID       *string                           `json:"domainUUID,omitempty"`
	TaskList         *TaskList                         `json:"taskList,omitempty"`
	TaskID           *string                           `json:"taskID,omitempty"`
	CompletedRequest *RespondQueryTaskCompletedRequest `json:"completedRequest,omitempty"`
}

// GetDomainUUID is an internal getter (TBD...)
func (v *MatchingRespondQueryTaskCompletedRequest) GetDomainUUID() (o string) {
	if v != nil && v.DomainUUID != nil {
		return *v.DomainUUID
	}
	return
}

// GetTaskList is an internal getter (TBD...)
func (v *MatchingRespondQueryTaskCompletedRequest) GetTaskList() (o *TaskList) {
	if v != nil && v.TaskList != nil {
		return v.TaskList
	}
	return
}

// GetTaskID is an internal getter (TBD...)
func (v *MatchingRespondQueryTaskCompletedRequest) GetTaskID() (o string) {
	if v != nil && v.TaskID != nil {
		return *v.TaskID
	}
	return
}

// GetCompletedRequest is an internal getter (TBD...)
func (v *MatchingRespondQueryTaskCompletedRequest) GetCompletedRequest() (o *RespondQueryTaskCompletedRequest) {
	if v != nil && v.CompletedRequest != nil {
		return v.CompletedRequest
	}
	return
}

// TaskSource is an internal type (TBD...)
type TaskSource int32

// Ptr is a helper function for getting pointer value
func (e TaskSource) Ptr() *TaskSource {
	return &e
}

// String returns a readable string representation of TaskSource.
func (e TaskSource) String() string {
	w := int32(e)
	switch w {
	case 0:
		return "HISTORY"
	case 1:
		return "DB_BACKLOG"
	}
	return fmt.Sprintf("TaskSource(%d)", w)
}

// UnmarshalText parses enum value from string representation
func (e *TaskSource) UnmarshalText(value []byte) error {
	switch s := strings.ToUpper(string(value)); s {
	case "HISTORY":
		*e = TaskSourceHistory
		return nil
	case "DB_BACKLOG":
		*e = TaskSourceDbBacklog
		return nil
	default:
		val, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return fmt.Errorf("unknown enum value %q for %q: %v", s, "TaskSource", err)
		}
		*e = TaskSource(val)
		return nil
	}
}

// MarshalText encodes TaskSource to text.
func (e TaskSource) MarshalText() ([]byte, error) {
	return []byte(e.String()), nil
}

const (
	// TaskSourceHistory is an option for TaskSource
	TaskSourceHistory TaskSource = iota
	// TaskSourceDbBacklog is an option for TaskSource
	TaskSourceDbBacklog
)
