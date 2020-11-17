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

package thrift

import (
	"github.com/uber/cadence/common/types"

	"github.com/uber/cadence/.gen/go/matching"
)

// FromAddActivityTaskRequest converts internal AddActivityTaskRequest type to thrift
func FromAddActivityTaskRequest(t *types.AddActivityTaskRequest) *matching.AddActivityTaskRequest {
	if t == nil {
		return nil
	}
	return &matching.AddActivityTaskRequest{
		DomainUUID:                    t.DomainUUID,
		Execution:                     FromWorkflowExecution(t.Execution),
		SourceDomainUUID:              t.SourceDomainUUID,
		TaskList:                      FromTaskList(t.TaskList),
		ScheduleId:                    t.ScheduleID,
		ScheduleToStartTimeoutSeconds: t.ScheduleToStartTimeoutSeconds,
		Source:                        FromTaskSource(t.Source),
		ForwardedFrom:                 t.ForwardedFrom,
	}
}

// ToAddActivityTaskRequest converts thrift AddActivityTaskRequest type to internal
func ToAddActivityTaskRequest(t *matching.AddActivityTaskRequest) *types.AddActivityTaskRequest {
	if t == nil {
		return nil
	}
	return &types.AddActivityTaskRequest{
		DomainUUID:                    t.DomainUUID,
		Execution:                     ToWorkflowExecution(t.Execution),
		SourceDomainUUID:              t.SourceDomainUUID,
		TaskList:                      ToTaskList(t.TaskList),
		ScheduleID:                    t.ScheduleId,
		ScheduleToStartTimeoutSeconds: t.ScheduleToStartTimeoutSeconds,
		Source:                        ToTaskSource(t.Source),
		ForwardedFrom:                 t.ForwardedFrom,
	}
}

// FromAddDecisionTaskRequest converts internal AddDecisionTaskRequest type to thrift
func FromAddDecisionTaskRequest(t *types.AddDecisionTaskRequest) *matching.AddDecisionTaskRequest {
	if t == nil {
		return nil
	}
	return &matching.AddDecisionTaskRequest{
		DomainUUID:                    t.DomainUUID,
		Execution:                     FromWorkflowExecution(t.Execution),
		TaskList:                      FromTaskList(t.TaskList),
		ScheduleId:                    t.ScheduleID,
		ScheduleToStartTimeoutSeconds: t.ScheduleToStartTimeoutSeconds,
		Source:                        FromTaskSource(t.Source),
		ForwardedFrom:                 t.ForwardedFrom,
	}
}

// ToAddDecisionTaskRequest converts thrift AddDecisionTaskRequest type to internal
func ToAddDecisionTaskRequest(t *matching.AddDecisionTaskRequest) *types.AddDecisionTaskRequest {
	if t == nil {
		return nil
	}
	return &types.AddDecisionTaskRequest{
		DomainUUID:                    t.DomainUUID,
		Execution:                     ToWorkflowExecution(t.Execution),
		TaskList:                      ToTaskList(t.TaskList),
		ScheduleID:                    t.ScheduleId,
		ScheduleToStartTimeoutSeconds: t.ScheduleToStartTimeoutSeconds,
		Source:                        ToTaskSource(t.Source),
		ForwardedFrom:                 t.ForwardedFrom,
	}
}

// FromCancelOutstandingPollRequest converts internal CancelOutstandingPollRequest type to thrift
func FromCancelOutstandingPollRequest(t *types.CancelOutstandingPollRequest) *matching.CancelOutstandingPollRequest {
	if t == nil {
		return nil
	}
	return &matching.CancelOutstandingPollRequest{
		DomainUUID:   t.DomainUUID,
		TaskListType: t.TaskListType,
		TaskList:     FromTaskList(t.TaskList),
		PollerID:     t.PollerID,
	}
}

// ToCancelOutstandingPollRequest converts thrift CancelOutstandingPollRequest type to internal
func ToCancelOutstandingPollRequest(t *matching.CancelOutstandingPollRequest) *types.CancelOutstandingPollRequest {
	if t == nil {
		return nil
	}
	return &types.CancelOutstandingPollRequest{
		DomainUUID:   t.DomainUUID,
		TaskListType: t.TaskListType,
		TaskList:     ToTaskList(t.TaskList),
		PollerID:     t.PollerID,
	}
}

// FromMatchingDescribeTaskListRequest converts internal DescribeTaskListRequest type to thrift
func FromMatchingDescribeTaskListRequest(t *types.MatchingDescribeTaskListRequest) *matching.DescribeTaskListRequest {
	if t == nil {
		return nil
	}
	return &matching.DescribeTaskListRequest{
		DomainUUID:  t.DomainUUID,
		DescRequest: FromDescribeTaskListRequest(t.DescRequest),
	}
}

// ToMatchingDescribeTaskListRequest converts thrift DescribeTaskListRequest type to internal
func ToMatchingDescribeTaskListRequest(t *matching.DescribeTaskListRequest) *types.MatchingDescribeTaskListRequest {
	if t == nil {
		return nil
	}
	return &types.MatchingDescribeTaskListRequest{
		DomainUUID:  t.DomainUUID,
		DescRequest: ToDescribeTaskListRequest(t.DescRequest),
	}
}

// FromMatchingListTaskListPartitionsRequest converts internal ListTaskListPartitionsRequest type to thrift
func FromMatchingListTaskListPartitionsRequest(t *types.MatchingListTaskListPartitionsRequest) *matching.ListTaskListPartitionsRequest {
	if t == nil {
		return nil
	}
	return &matching.ListTaskListPartitionsRequest{
		Domain:   t.Domain,
		TaskList: FromTaskList(t.TaskList),
	}
}

// ToMatchingListTaskListPartitionsRequest converts thrift ListTaskListPartitionsRequest type to internal
func ToMatchingListTaskListPartitionsRequest(t *matching.ListTaskListPartitionsRequest) *types.MatchingListTaskListPartitionsRequest {
	if t == nil {
		return nil
	}
	return &types.MatchingListTaskListPartitionsRequest{
		Domain:   t.Domain,
		TaskList: ToTaskList(t.TaskList),
	}
}

// FromMatchingServiceAddActivityTaskArgs converts internal MatchingService_AddActivityTask_Args type to thrift
func FromMatchingServiceAddActivityTaskArgs(t *types.MatchingServiceAddActivityTaskArgs) *matching.MatchingService_AddActivityTask_Args {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_AddActivityTask_Args{
		AddRequest: FromAddActivityTaskRequest(t.AddRequest),
	}
}

// ToMatchingServiceAddActivityTaskArgs converts thrift MatchingService_AddActivityTask_Args type to internal
func ToMatchingServiceAddActivityTaskArgs(t *matching.MatchingService_AddActivityTask_Args) *types.MatchingServiceAddActivityTaskArgs {
	if t == nil {
		return nil
	}
	return &types.MatchingServiceAddActivityTaskArgs{
		AddRequest: ToAddActivityTaskRequest(t.AddRequest),
	}
}

// FromMatchingServiceAddActivityTaskResult converts internal MatchingService_AddActivityTask_Result type to thrift
func FromMatchingServiceAddActivityTaskResult(t *types.MatchingServiceAddActivityTaskResult) *matching.MatchingService_AddActivityTask_Result {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_AddActivityTask_Result{
		BadRequestError:        FromBadRequestError(t.BadRequestError),
		InternalServiceError:   FromInternalServiceError(t.InternalServiceError),
		ServiceBusyError:       FromServiceBusyError(t.ServiceBusyError),
		LimitExceededError:     FromLimitExceededError(t.LimitExceededError),
		DomainNotActiveError:   FromDomainNotActiveError(t.DomainNotActiveError),
		RemoteSyncMatchedError: FromRemoteSyncMatchedError(t.RemoteSyncMatchedError),
	}
}

// ToMatchingServiceAddActivityTaskResult converts thrift MatchingService_AddActivityTask_Result type to internal
func ToMatchingServiceAddActivityTaskResult(t *matching.MatchingService_AddActivityTask_Result) *types.MatchingServiceAddActivityTaskResult {
	if t == nil {
		return nil
	}
	return &types.MatchingServiceAddActivityTaskResult{
		BadRequestError:        ToBadRequestError(t.BadRequestError),
		InternalServiceError:   ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:       ToServiceBusyError(t.ServiceBusyError),
		LimitExceededError:     ToLimitExceededError(t.LimitExceededError),
		DomainNotActiveError:   ToDomainNotActiveError(t.DomainNotActiveError),
		RemoteSyncMatchedError: ToRemoteSyncMatchedError(t.RemoteSyncMatchedError),
	}
}

// FromMatchingServiceAddDecisionTaskArgs converts internal MatchingService_AddDecisionTask_Args type to thrift
func FromMatchingServiceAddDecisionTaskArgs(t *types.MatchingServiceAddDecisionTaskArgs) *matching.MatchingService_AddDecisionTask_Args {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_AddDecisionTask_Args{
		AddRequest: FromAddDecisionTaskRequest(t.AddRequest),
	}
}

// ToMatchingServiceAddDecisionTaskArgs converts thrift MatchingService_AddDecisionTask_Args type to internal
func ToMatchingServiceAddDecisionTaskArgs(t *matching.MatchingService_AddDecisionTask_Args) *types.MatchingServiceAddDecisionTaskArgs {
	if t == nil {
		return nil
	}
	return &types.MatchingServiceAddDecisionTaskArgs{
		AddRequest: ToAddDecisionTaskRequest(t.AddRequest),
	}
}

// FromMatchingServiceAddDecisionTaskResult converts internal MatchingService_AddDecisionTask_Result type to thrift
func FromMatchingServiceAddDecisionTaskResult(t *types.MatchingServiceAddDecisionTaskResult) *matching.MatchingService_AddDecisionTask_Result {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_AddDecisionTask_Result{
		BadRequestError:        FromBadRequestError(t.BadRequestError),
		InternalServiceError:   FromInternalServiceError(t.InternalServiceError),
		ServiceBusyError:       FromServiceBusyError(t.ServiceBusyError),
		LimitExceededError:     FromLimitExceededError(t.LimitExceededError),
		DomainNotActiveError:   FromDomainNotActiveError(t.DomainNotActiveError),
		RemoteSyncMatchedError: FromRemoteSyncMatchedError(t.RemoteSyncMatchedError),
	}
}

// ToMatchingServiceAddDecisionTaskResult converts thrift MatchingService_AddDecisionTask_Result type to internal
func ToMatchingServiceAddDecisionTaskResult(t *matching.MatchingService_AddDecisionTask_Result) *types.MatchingServiceAddDecisionTaskResult {
	if t == nil {
		return nil
	}
	return &types.MatchingServiceAddDecisionTaskResult{
		BadRequestError:        ToBadRequestError(t.BadRequestError),
		InternalServiceError:   ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:       ToServiceBusyError(t.ServiceBusyError),
		LimitExceededError:     ToLimitExceededError(t.LimitExceededError),
		DomainNotActiveError:   ToDomainNotActiveError(t.DomainNotActiveError),
		RemoteSyncMatchedError: ToRemoteSyncMatchedError(t.RemoteSyncMatchedError),
	}
}

// FromMatchingServiceCancelOutstandingPollArgs converts internal MatchingService_CancelOutstandingPoll_Args type to thrift
func FromMatchingServiceCancelOutstandingPollArgs(t *types.MatchingServiceCancelOutstandingPollArgs) *matching.MatchingService_CancelOutstandingPoll_Args {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_CancelOutstandingPoll_Args{
		Request: FromCancelOutstandingPollRequest(t.Request),
	}
}

// ToMatchingServiceCancelOutstandingPollArgs converts thrift MatchingService_CancelOutstandingPoll_Args type to internal
func ToMatchingServiceCancelOutstandingPollArgs(t *matching.MatchingService_CancelOutstandingPoll_Args) *types.MatchingServiceCancelOutstandingPollArgs {
	if t == nil {
		return nil
	}
	return &types.MatchingServiceCancelOutstandingPollArgs{
		Request: ToCancelOutstandingPollRequest(t.Request),
	}
}

// FromMatchingServiceCancelOutstandingPollResult converts internal MatchingService_CancelOutstandingPoll_Result type to thrift
func FromMatchingServiceCancelOutstandingPollResult(t *types.MatchingServiceCancelOutstandingPollResult) *matching.MatchingService_CancelOutstandingPoll_Result {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_CancelOutstandingPoll_Result{
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToMatchingServiceCancelOutstandingPollResult converts thrift MatchingService_CancelOutstandingPoll_Result type to internal
func ToMatchingServiceCancelOutstandingPollResult(t *matching.MatchingService_CancelOutstandingPoll_Result) *types.MatchingServiceCancelOutstandingPollResult {
	if t == nil {
		return nil
	}
	return &types.MatchingServiceCancelOutstandingPollResult{
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromMatchingServiceDescribeTaskListArgs converts internal MatchingService_DescribeTaskList_Args type to thrift
func FromMatchingServiceDescribeTaskListArgs(t *types.MatchingServiceDescribeTaskListArgs) *matching.MatchingService_DescribeTaskList_Args {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_DescribeTaskList_Args{
		Request: FromMatchingDescribeTaskListRequest(t.Request),
	}
}

// ToMatchingServiceDescribeTaskListArgs converts thrift MatchingService_DescribeTaskList_Args type to internal
func ToMatchingServiceDescribeTaskListArgs(t *matching.MatchingService_DescribeTaskList_Args) *types.MatchingServiceDescribeTaskListArgs {
	if t == nil {
		return nil
	}
	return &types.MatchingServiceDescribeTaskListArgs{
		Request: ToMatchingDescribeTaskListRequest(t.Request),
	}
}

// FromMatchingServiceDescribeTaskListResult converts internal MatchingService_DescribeTaskList_Result type to thrift
func FromMatchingServiceDescribeTaskListResult(t *types.MatchingServiceDescribeTaskListResult) *matching.MatchingService_DescribeTaskList_Result {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_DescribeTaskList_Result{
		Success:              FromDescribeTaskListResponse(t.Success),
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:  FromEntityNotExistsError(t.EntityNotExistError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToMatchingServiceDescribeTaskListResult converts thrift MatchingService_DescribeTaskList_Result type to internal
func ToMatchingServiceDescribeTaskListResult(t *matching.MatchingService_DescribeTaskList_Result) *types.MatchingServiceDescribeTaskListResult {
	if t == nil {
		return nil
	}
	return &types.MatchingServiceDescribeTaskListResult{
		Success:              ToDescribeTaskListResponse(t.Success),
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:  ToEntityNotExistsError(t.EntityNotExistError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromMatchingServiceListTaskListPartitionsArgs converts internal MatchingService_ListTaskListPartitions_Args type to thrift
func FromMatchingServiceListTaskListPartitionsArgs(t *types.MatchingServiceListTaskListPartitionsArgs) *matching.MatchingService_ListTaskListPartitions_Args {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_ListTaskListPartitions_Args{
		Request: FromMatchingListTaskListPartitionsRequest(t.Request),
	}
}

// ToMatchingServiceListTaskListPartitionsArgs converts thrift MatchingService_ListTaskListPartitions_Args type to internal
func ToMatchingServiceListTaskListPartitionsArgs(t *matching.MatchingService_ListTaskListPartitions_Args) *types.MatchingServiceListTaskListPartitionsArgs {
	if t == nil {
		return nil
	}
	return &types.MatchingServiceListTaskListPartitionsArgs{
		Request: ToMatchingListTaskListPartitionsRequest(t.Request),
	}
}

// FromMatchingServiceListTaskListPartitionsResult converts internal MatchingService_ListTaskListPartitions_Result type to thrift
func FromMatchingServiceListTaskListPartitionsResult(t *types.MatchingServiceListTaskListPartitionsResult) *matching.MatchingService_ListTaskListPartitions_Result {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_ListTaskListPartitions_Result{
		Success:              FromListTaskListPartitionsResponse(t.Success),
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToMatchingServiceListTaskListPartitionsResult converts thrift MatchingService_ListTaskListPartitions_Result type to internal
func ToMatchingServiceListTaskListPartitionsResult(t *matching.MatchingService_ListTaskListPartitions_Result) *types.MatchingServiceListTaskListPartitionsResult {
	if t == nil {
		return nil
	}
	return &types.MatchingServiceListTaskListPartitionsResult{
		Success:              ToListTaskListPartitionsResponse(t.Success),
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromMatchingServicePollForActivityTaskArgs converts internal MatchingService_PollForActivityTask_Args type to thrift
func FromMatchingServicePollForActivityTaskArgs(t *types.MatchingServicePollForActivityTaskArgs) *matching.MatchingService_PollForActivityTask_Args {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_PollForActivityTask_Args{
		PollRequest: FromMatchingPollForActivityTaskRequest(t.PollRequest),
	}
}

// ToMatchingServicePollForActivityTaskArgs converts thrift MatchingService_PollForActivityTask_Args type to internal
func ToMatchingServicePollForActivityTaskArgs(t *matching.MatchingService_PollForActivityTask_Args) *types.MatchingServicePollForActivityTaskArgs {
	if t == nil {
		return nil
	}
	return &types.MatchingServicePollForActivityTaskArgs{
		PollRequest: ToMatchingPollForActivityTaskRequest(t.PollRequest),
	}
}

// FromMatchingServicePollForActivityTaskResult converts internal MatchingService_PollForActivityTask_Result type to thrift
func FromMatchingServicePollForActivityTaskResult(t *types.MatchingServicePollForActivityTaskResult) *matching.MatchingService_PollForActivityTask_Result {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_PollForActivityTask_Result{
		Success:              FromPollForActivityTaskResponse(t.Success),
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		LimitExceededError:   FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToMatchingServicePollForActivityTaskResult converts thrift MatchingService_PollForActivityTask_Result type to internal
func ToMatchingServicePollForActivityTaskResult(t *matching.MatchingService_PollForActivityTask_Result) *types.MatchingServicePollForActivityTaskResult {
	if t == nil {
		return nil
	}
	return &types.MatchingServicePollForActivityTaskResult{
		Success:              ToPollForActivityTaskResponse(t.Success),
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		LimitExceededError:   ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromMatchingServicePollForDecisionTaskArgs converts internal MatchingService_PollForDecisionTask_Args type to thrift
func FromMatchingServicePollForDecisionTaskArgs(t *types.MatchingServicePollForDecisionTaskArgs) *matching.MatchingService_PollForDecisionTask_Args {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_PollForDecisionTask_Args{
		PollRequest: FromMatchingPollForDecisionTaskRequest(t.PollRequest),
	}
}

// ToMatchingServicePollForDecisionTaskArgs converts thrift MatchingService_PollForDecisionTask_Args type to internal
func ToMatchingServicePollForDecisionTaskArgs(t *matching.MatchingService_PollForDecisionTask_Args) *types.MatchingServicePollForDecisionTaskArgs {
	if t == nil {
		return nil
	}
	return &types.MatchingServicePollForDecisionTaskArgs{
		PollRequest: ToMatchingPollForDecisionTaskRequest(t.PollRequest),
	}
}

// FromMatchingServicePollForDecisionTaskResult converts internal MatchingService_PollForDecisionTask_Result type to thrift
func FromMatchingServicePollForDecisionTaskResult(t *types.MatchingServicePollForDecisionTaskResult) *matching.MatchingService_PollForDecisionTask_Result {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_PollForDecisionTask_Result{
		Success:              FromMatchingPollForDecisionTaskResponse(t.Success),
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		LimitExceededError:   FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToMatchingServicePollForDecisionTaskResult converts thrift MatchingService_PollForDecisionTask_Result type to internal
func ToMatchingServicePollForDecisionTaskResult(t *matching.MatchingService_PollForDecisionTask_Result) *types.MatchingServicePollForDecisionTaskResult {
	if t == nil {
		return nil
	}
	return &types.MatchingServicePollForDecisionTaskResult{
		Success:              ToMatchingPollForDecisionTaskResponse(t.Success),
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		LimitExceededError:   ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromMatchingServiceQueryWorkflowArgs converts internal MatchingService_QueryWorkflow_Args type to thrift
func FromMatchingServiceQueryWorkflowArgs(t *types.MatchingServiceQueryWorkflowArgs) *matching.MatchingService_QueryWorkflow_Args {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_QueryWorkflow_Args{
		QueryRequest: FromMatchingQueryWorkflowRequest(t.QueryRequest),
	}
}

// ToMatchingServiceQueryWorkflowArgs converts thrift MatchingService_QueryWorkflow_Args type to internal
func ToMatchingServiceQueryWorkflowArgs(t *matching.MatchingService_QueryWorkflow_Args) *types.MatchingServiceQueryWorkflowArgs {
	if t == nil {
		return nil
	}
	return &types.MatchingServiceQueryWorkflowArgs{
		QueryRequest: ToMatchingQueryWorkflowRequest(t.QueryRequest),
	}
}

// FromMatchingServiceQueryWorkflowResult converts internal MatchingService_QueryWorkflow_Result type to thrift
func FromMatchingServiceQueryWorkflowResult(t *types.MatchingServiceQueryWorkflowResult) *matching.MatchingService_QueryWorkflow_Result {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_QueryWorkflow_Result{
		Success:              FromQueryWorkflowResponse(t.Success),
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:  FromEntityNotExistsError(t.EntityNotExistError),
		QueryFailedError:     FromQueryFailedError(t.QueryFailedError),
		LimitExceededError:   FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToMatchingServiceQueryWorkflowResult converts thrift MatchingService_QueryWorkflow_Result type to internal
func ToMatchingServiceQueryWorkflowResult(t *matching.MatchingService_QueryWorkflow_Result) *types.MatchingServiceQueryWorkflowResult {
	if t == nil {
		return nil
	}
	return &types.MatchingServiceQueryWorkflowResult{
		Success:              ToQueryWorkflowResponse(t.Success),
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:  ToEntityNotExistsError(t.EntityNotExistError),
		QueryFailedError:     ToQueryFailedError(t.QueryFailedError),
		LimitExceededError:   ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromMatchingServiceRespondQueryTaskCompletedArgs converts internal MatchingService_RespondQueryTaskCompleted_Args type to thrift
func FromMatchingServiceRespondQueryTaskCompletedArgs(t *types.MatchingServiceRespondQueryTaskCompletedArgs) *matching.MatchingService_RespondQueryTaskCompleted_Args {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_RespondQueryTaskCompleted_Args{
		Request: FromMatchingRespondQueryTaskCompletedRequest(t.Request),
	}
}

// ToMatchingServiceRespondQueryTaskCompletedArgs converts thrift MatchingService_RespondQueryTaskCompleted_Args type to internal
func ToMatchingServiceRespondQueryTaskCompletedArgs(t *matching.MatchingService_RespondQueryTaskCompleted_Args) *types.MatchingServiceRespondQueryTaskCompletedArgs {
	if t == nil {
		return nil
	}
	return &types.MatchingServiceRespondQueryTaskCompletedArgs{
		Request: ToMatchingRespondQueryTaskCompletedRequest(t.Request),
	}
}

// FromMatchingServiceRespondQueryTaskCompletedResult converts internal MatchingService_RespondQueryTaskCompleted_Result type to thrift
func FromMatchingServiceRespondQueryTaskCompletedResult(t *types.MatchingServiceRespondQueryTaskCompletedResult) *matching.MatchingService_RespondQueryTaskCompleted_Result {
	if t == nil {
		return nil
	}
	return &matching.MatchingService_RespondQueryTaskCompleted_Result{
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:  FromEntityNotExistsError(t.EntityNotExistError),
		LimitExceededError:   FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToMatchingServiceRespondQueryTaskCompletedResult converts thrift MatchingService_RespondQueryTaskCompleted_Result type to internal
func ToMatchingServiceRespondQueryTaskCompletedResult(t *matching.MatchingService_RespondQueryTaskCompleted_Result) *types.MatchingServiceRespondQueryTaskCompletedResult {
	if t == nil {
		return nil
	}
	return &types.MatchingServiceRespondQueryTaskCompletedResult{
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:  ToEntityNotExistsError(t.EntityNotExistError),
		LimitExceededError:   ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromMatchingPollForActivityTaskRequest converts internal PollForActivityTaskRequest type to thrift
func FromMatchingPollForActivityTaskRequest(t *types.MatchingPollForActivityTaskRequest) *matching.PollForActivityTaskRequest {
	if t == nil {
		return nil
	}
	return &matching.PollForActivityTaskRequest{
		DomainUUID:    t.DomainUUID,
		PollerID:      t.PollerID,
		PollRequest:   FromPollForActivityTaskRequest(t.PollRequest),
		ForwardedFrom: t.ForwardedFrom,
	}
}

// ToMatchingPollForActivityTaskRequest converts thrift PollForActivityTaskRequest type to internal
func ToMatchingPollForActivityTaskRequest(t *matching.PollForActivityTaskRequest) *types.MatchingPollForActivityTaskRequest {
	if t == nil {
		return nil
	}
	return &types.MatchingPollForActivityTaskRequest{
		DomainUUID:    t.DomainUUID,
		PollerID:      t.PollerID,
		PollRequest:   ToPollForActivityTaskRequest(t.PollRequest),
		ForwardedFrom: t.ForwardedFrom,
	}
}

// FromMatchingPollForDecisionTaskRequest converts internal PollForDecisionTaskRequest type to thrift
func FromMatchingPollForDecisionTaskRequest(t *types.MatchingPollForDecisionTaskRequest) *matching.PollForDecisionTaskRequest {
	if t == nil {
		return nil
	}
	return &matching.PollForDecisionTaskRequest{
		DomainUUID:    t.DomainUUID,
		PollerID:      t.PollerID,
		PollRequest:   FromPollForDecisionTaskRequest(t.PollRequest),
		ForwardedFrom: t.ForwardedFrom,
	}
}

// ToMatchingPollForDecisionTaskRequest converts thrift PollForDecisionTaskRequest type to internal
func ToMatchingPollForDecisionTaskRequest(t *matching.PollForDecisionTaskRequest) *types.MatchingPollForDecisionTaskRequest {
	if t == nil {
		return nil
	}
	return &types.MatchingPollForDecisionTaskRequest{
		DomainUUID:    t.DomainUUID,
		PollerID:      t.PollerID,
		PollRequest:   ToPollForDecisionTaskRequest(t.PollRequest),
		ForwardedFrom: t.ForwardedFrom,
	}
}

// FromMatchingPollForDecisionTaskResponse converts internal PollForDecisionTaskResponse type to thrift
func FromMatchingPollForDecisionTaskResponse(t *types.MatchingPollForDecisionTaskResponse) *matching.PollForDecisionTaskResponse {
	if t == nil {
		return nil
	}
	return &matching.PollForDecisionTaskResponse{
		TaskToken:                 t.TaskToken,
		WorkflowExecution:         FromWorkflowExecution(t.WorkflowExecution),
		WorkflowType:              FromWorkflowType(t.WorkflowType),
		PreviousStartedEventId:    t.PreviousStartedEventID,
		StartedEventId:            t.StartedEventID,
		Attempt:                   t.Attempt,
		NextEventId:               t.NextEventID,
		BacklogCountHint:          t.BacklogCountHint,
		StickyExecutionEnabled:    t.StickyExecutionEnabled,
		Query:                     FromWorkflowQuery(t.Query),
		DecisionInfo:              FromTransientDecisionInfo(t.DecisionInfo),
		WorkflowExecutionTaskList: FromTaskList(t.WorkflowExecutionTaskList),
		EventStoreVersion:         t.EventStoreVersion,
		BranchToken:               t.BranchToken,
		ScheduledTimestamp:        t.ScheduledTimestamp,
		StartedTimestamp:          t.StartedTimestamp,
		Queries:                   FromWorkflowQueryMap(t.Queries),
	}
}

// ToMatchingPollForDecisionTaskResponse converts thrift PollForDecisionTaskResponse type to internal
func ToMatchingPollForDecisionTaskResponse(t *matching.PollForDecisionTaskResponse) *types.MatchingPollForDecisionTaskResponse {
	if t == nil {
		return nil
	}
	return &types.MatchingPollForDecisionTaskResponse{
		TaskToken:                 t.TaskToken,
		WorkflowExecution:         ToWorkflowExecution(t.WorkflowExecution),
		WorkflowType:              ToWorkflowType(t.WorkflowType),
		PreviousStartedEventID:    t.PreviousStartedEventId,
		StartedEventID:            t.StartedEventId,
		Attempt:                   t.Attempt,
		NextEventID:               t.NextEventId,
		BacklogCountHint:          t.BacklogCountHint,
		StickyExecutionEnabled:    t.StickyExecutionEnabled,
		Query:                     ToWorkflowQuery(t.Query),
		DecisionInfo:              ToTransientDecisionInfo(t.DecisionInfo),
		WorkflowExecutionTaskList: ToTaskList(t.WorkflowExecutionTaskList),
		EventStoreVersion:         t.EventStoreVersion,
		BranchToken:               t.BranchToken,
		ScheduledTimestamp:        t.ScheduledTimestamp,
		StartedTimestamp:          t.StartedTimestamp,
		Queries:                   ToWorkflowQueryMap(t.Queries),
	}
}

// FromMatchingQueryWorkflowRequest converts internal QueryWorkflowRequest type to thrift
func FromMatchingQueryWorkflowRequest(t *types.MatchingQueryWorkflowRequest) *matching.QueryWorkflowRequest {
	if t == nil {
		return nil
	}
	return &matching.QueryWorkflowRequest{
		DomainUUID:    t.DomainUUID,
		TaskList:      FromTaskList(t.TaskList),
		QueryRequest:  FromQueryWorkflowRequest(t.QueryRequest),
		ForwardedFrom: t.ForwardedFrom,
	}
}

// ToMatchingQueryWorkflowRequest converts thrift QueryWorkflowRequest type to internal
func ToMatchingQueryWorkflowRequest(t *matching.QueryWorkflowRequest) *types.MatchingQueryWorkflowRequest {
	if t == nil {
		return nil
	}
	return &types.MatchingQueryWorkflowRequest{
		DomainUUID:    t.DomainUUID,
		TaskList:      ToTaskList(t.TaskList),
		QueryRequest:  ToQueryWorkflowRequest(t.QueryRequest),
		ForwardedFrom: t.ForwardedFrom,
	}
}

// FromMatchingRespondQueryTaskCompletedRequest converts internal RespondQueryTaskCompletedRequest type to thrift
func FromMatchingRespondQueryTaskCompletedRequest(t *types.MatchingRespondQueryTaskCompletedRequest) *matching.RespondQueryTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &matching.RespondQueryTaskCompletedRequest{
		DomainUUID:       t.DomainUUID,
		TaskList:         FromTaskList(t.TaskList),
		TaskID:           t.TaskID,
		CompletedRequest: FromRespondQueryTaskCompletedRequest(t.CompletedRequest),
	}
}

// ToMatchingRespondQueryTaskCompletedRequest converts thrift RespondQueryTaskCompletedRequest type to internal
func ToMatchingRespondQueryTaskCompletedRequest(t *matching.RespondQueryTaskCompletedRequest) *types.MatchingRespondQueryTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &types.MatchingRespondQueryTaskCompletedRequest{
		DomainUUID:       t.DomainUUID,
		TaskList:         ToTaskList(t.TaskList),
		TaskID:           t.TaskID,
		CompletedRequest: ToRespondQueryTaskCompletedRequest(t.CompletedRequest),
	}
}

// FromTaskSource converts internal TaskSource type to thrift
func FromTaskSource(t *types.TaskSource) *matching.TaskSource {
	if t == nil {
		return nil
	}
	switch *t {
	case types.TaskSourceHistory:
		v := matching.TaskSourceHistory
		return &v
	case types.TaskSourceDbBacklog:
		v := matching.TaskSourceDbBacklog
		return &v
	}
	panic("unexpected enum value")
}

// ToTaskSource converts thrift TaskSource type to internal
func ToTaskSource(t *matching.TaskSource) *types.TaskSource {
	if t == nil {
		return nil
	}
	switch *t {
	case matching.TaskSourceHistory:
		v := types.TaskSourceHistory
		return &v
	case matching.TaskSourceDbBacklog:
		v := types.TaskSourceDbBacklog
		return &v
	}
	panic("unexpected enum value")
}
