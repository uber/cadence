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

	"github.com/uber/cadence/.gen/go/history"
)

// FromDescribeMutableStateRequest converts internal DescribeMutableStateRequest type to thrift
func FromDescribeMutableStateRequest(t *types.DescribeMutableStateRequest) *history.DescribeMutableStateRequest {
	if t == nil {
		return nil
	}
	return &history.DescribeMutableStateRequest{
		DomainUUID: t.DomainUUID,
		Execution:  FromWorkflowExecution(t.Execution),
	}
}

// ToDescribeMutableStateRequest converts thrift DescribeMutableStateRequest type to internal
func ToDescribeMutableStateRequest(t *history.DescribeMutableStateRequest) *types.DescribeMutableStateRequest {
	if t == nil {
		return nil
	}
	return &types.DescribeMutableStateRequest{
		DomainUUID: t.DomainUUID,
		Execution:  ToWorkflowExecution(t.Execution),
	}
}

// FromDescribeMutableStateResponse converts internal DescribeMutableStateResponse type to thrift
func FromDescribeMutableStateResponse(t *types.DescribeMutableStateResponse) *history.DescribeMutableStateResponse {
	if t == nil {
		return nil
	}
	return &history.DescribeMutableStateResponse{
		MutableStateInCache:    t.MutableStateInCache,
		MutableStateInDatabase: t.MutableStateInDatabase,
	}
}

// ToDescribeMutableStateResponse converts thrift DescribeMutableStateResponse type to internal
func ToDescribeMutableStateResponse(t *history.DescribeMutableStateResponse) *types.DescribeMutableStateResponse {
	if t == nil {
		return nil
	}
	return &types.DescribeMutableStateResponse{
		MutableStateInCache:    t.MutableStateInCache,
		MutableStateInDatabase: t.MutableStateInDatabase,
	}
}

// FromHistoryDescribeWorkflowExecutionRequest converts internal DescribeWorkflowExecutionRequest type to thrift
func FromHistoryDescribeWorkflowExecutionRequest(t *types.HistoryDescribeWorkflowExecutionRequest) *history.DescribeWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &history.DescribeWorkflowExecutionRequest{
		DomainUUID: t.DomainUUID,
		Request:    FromDescribeWorkflowExecutionRequest(t.Request),
	}
}

// ToHistoryDescribeWorkflowExecutionRequest converts thrift DescribeWorkflowExecutionRequest type to internal
func ToHistoryDescribeWorkflowExecutionRequest(t *history.DescribeWorkflowExecutionRequest) *types.HistoryDescribeWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryDescribeWorkflowExecutionRequest{
		DomainUUID: t.DomainUUID,
		Request:    ToDescribeWorkflowExecutionRequest(t.Request),
	}
}

// FromDomainFilter converts internal DomainFilter type to thrift
func FromDomainFilter(t *types.DomainFilter) *history.DomainFilter {
	if t == nil {
		return nil
	}
	return &history.DomainFilter{
		DomainIDs:    t.DomainIDs,
		ReverseMatch: t.ReverseMatch,
	}
}

// ToDomainFilter converts thrift DomainFilter type to internal
func ToDomainFilter(t *history.DomainFilter) *types.DomainFilter {
	if t == nil {
		return nil
	}
	return &types.DomainFilter{
		DomainIDs:    t.DomainIDs,
		ReverseMatch: t.ReverseMatch,
	}
}

// FromEventAlreadyStartedError converts internal EventAlreadyStartedError type to thrift
func FromEventAlreadyStartedError(t *types.EventAlreadyStartedError) *history.EventAlreadyStartedError {
	if t == nil {
		return nil
	}
	return &history.EventAlreadyStartedError{
		Message: t.Message,
	}
}

// ToEventAlreadyStartedError converts thrift EventAlreadyStartedError type to internal
func ToEventAlreadyStartedError(t *history.EventAlreadyStartedError) *types.EventAlreadyStartedError {
	if t == nil {
		return nil
	}
	return &types.EventAlreadyStartedError{
		Message: t.Message,
	}
}

// FromFailoverMarkerToken converts internal FailoverMarkerToken type to thrift
func FromFailoverMarkerToken(t *types.FailoverMarkerToken) *history.FailoverMarkerToken {
	if t == nil {
		return nil
	}
	return &history.FailoverMarkerToken{
		ShardIDs:       t.ShardIDs,
		FailoverMarker: FromFailoverMarkerAttributes(t.FailoverMarker),
	}
}

// ToFailoverMarkerToken converts thrift FailoverMarkerToken type to internal
func ToFailoverMarkerToken(t *history.FailoverMarkerToken) *types.FailoverMarkerToken {
	if t == nil {
		return nil
	}
	return &types.FailoverMarkerToken{
		ShardIDs:       t.ShardIDs,
		FailoverMarker: ToFailoverMarkerAttributes(t.FailoverMarker),
	}
}

// FromGetMutableStateRequest converts internal GetMutableStateRequest type to thrift
func FromGetMutableStateRequest(t *types.GetMutableStateRequest) *history.GetMutableStateRequest {
	if t == nil {
		return nil
	}
	return &history.GetMutableStateRequest{
		DomainUUID:          t.DomainUUID,
		Execution:           FromWorkflowExecution(t.Execution),
		ExpectedNextEventId: t.ExpectedNextEventID,
		CurrentBranchToken:  t.CurrentBranchToken,
	}
}

// ToGetMutableStateRequest converts thrift GetMutableStateRequest type to internal
func ToGetMutableStateRequest(t *history.GetMutableStateRequest) *types.GetMutableStateRequest {
	if t == nil {
		return nil
	}
	return &types.GetMutableStateRequest{
		DomainUUID:          t.DomainUUID,
		Execution:           ToWorkflowExecution(t.Execution),
		ExpectedNextEventID: t.ExpectedNextEventId,
		CurrentBranchToken:  t.CurrentBranchToken,
	}
}

// FromGetMutableStateResponse converts internal GetMutableStateResponse type to thrift
func FromGetMutableStateResponse(t *types.GetMutableStateResponse) *history.GetMutableStateResponse {
	if t == nil {
		return nil
	}
	return &history.GetMutableStateResponse{
		Execution:                            FromWorkflowExecution(t.Execution),
		WorkflowType:                         FromWorkflowType(t.WorkflowType),
		NextEventId:                          t.NextEventID,
		PreviousStartedEventId:               t.PreviousStartedEventID,
		LastFirstEventId:                     t.LastFirstEventID,
		TaskList:                             FromTaskList(t.TaskList),
		StickyTaskList:                       FromTaskList(t.StickyTaskList),
		ClientLibraryVersion:                 t.ClientLibraryVersion,
		ClientFeatureVersion:                 t.ClientFeatureVersion,
		ClientImpl:                           t.ClientImpl,
		IsWorkflowRunning:                    t.IsWorkflowRunning,
		StickyTaskListScheduleToStartTimeout: t.StickyTaskListScheduleToStartTimeout,
		EventStoreVersion:                    t.EventStoreVersion,
		CurrentBranchToken:                   t.CurrentBranchToken,
		WorkflowState:                        t.WorkflowState,
		WorkflowCloseState:                   t.WorkflowCloseState,
		VersionHistories:                     FromVersionHistories(t.VersionHistories),
		IsStickyTaskListEnabled:              t.IsStickyTaskListEnabled,
	}
}

// ToGetMutableStateResponse converts thrift GetMutableStateResponse type to internal
func ToGetMutableStateResponse(t *history.GetMutableStateResponse) *types.GetMutableStateResponse {
	if t == nil {
		return nil
	}
	return &types.GetMutableStateResponse{
		Execution:                            ToWorkflowExecution(t.Execution),
		WorkflowType:                         ToWorkflowType(t.WorkflowType),
		NextEventID:                          t.NextEventId,
		PreviousStartedEventID:               t.PreviousStartedEventId,
		LastFirstEventID:                     t.LastFirstEventId,
		TaskList:                             ToTaskList(t.TaskList),
		StickyTaskList:                       ToTaskList(t.StickyTaskList),
		ClientLibraryVersion:                 t.ClientLibraryVersion,
		ClientFeatureVersion:                 t.ClientFeatureVersion,
		ClientImpl:                           t.ClientImpl,
		IsWorkflowRunning:                    t.IsWorkflowRunning,
		StickyTaskListScheduleToStartTimeout: t.StickyTaskListScheduleToStartTimeout,
		EventStoreVersion:                    t.EventStoreVersion,
		CurrentBranchToken:                   t.CurrentBranchToken,
		WorkflowState:                        t.WorkflowState,
		WorkflowCloseState:                   t.WorkflowCloseState,
		VersionHistories:                     ToVersionHistories(t.VersionHistories),
		IsStickyTaskListEnabled:              t.IsStickyTaskListEnabled,
	}
}

// FromHistoryServiceCloseShardArgs converts internal HistoryService_CloseShard_Args type to thrift
func FromHistoryServiceCloseShardArgs(t *types.HistoryServiceCloseShardArgs) *history.HistoryService_CloseShard_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_CloseShard_Args{
		Request: FromCloseShardRequest(t.Request),
	}
}

// ToHistoryServiceCloseShardArgs converts thrift HistoryService_CloseShard_Args type to internal
func ToHistoryServiceCloseShardArgs(t *history.HistoryService_CloseShard_Args) *types.HistoryServiceCloseShardArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceCloseShardArgs{
		Request: ToCloseShardRequest(t.Request),
	}
}

// FromHistoryServiceCloseShardResult converts internal HistoryService_CloseShard_Result type to thrift
func FromHistoryServiceCloseShardResult(t *types.HistoryServiceCloseShardResult) *history.HistoryService_CloseShard_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_CloseShard_Result{
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    FromAccessDeniedError(t.AccessDeniedError),
	}
}

// ToHistoryServiceCloseShardResult converts thrift HistoryService_CloseShard_Result type to internal
func ToHistoryServiceCloseShardResult(t *history.HistoryService_CloseShard_Result) *types.HistoryServiceCloseShardResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceCloseShardResult{
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    ToAccessDeniedError(t.AccessDeniedError),
	}
}

// FromHistoryServiceDescribeHistoryHostArgs converts internal HistoryService_DescribeHistoryHost_Args type to thrift
func FromHistoryServiceDescribeHistoryHostArgs(t *types.HistoryServiceDescribeHistoryHostArgs) *history.HistoryService_DescribeHistoryHost_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_DescribeHistoryHost_Args{
		Request: FromDescribeHistoryHostRequest(t.Request),
	}
}

// ToHistoryServiceDescribeHistoryHostArgs converts thrift HistoryService_DescribeHistoryHost_Args type to internal
func ToHistoryServiceDescribeHistoryHostArgs(t *history.HistoryService_DescribeHistoryHost_Args) *types.HistoryServiceDescribeHistoryHostArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceDescribeHistoryHostArgs{
		Request: ToDescribeHistoryHostRequest(t.Request),
	}
}

// FromHistoryServiceDescribeHistoryHostResult converts internal HistoryService_DescribeHistoryHost_Result type to thrift
func FromHistoryServiceDescribeHistoryHostResult(t *types.HistoryServiceDescribeHistoryHostResult) *history.HistoryService_DescribeHistoryHost_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_DescribeHistoryHost_Result{
		Success:              FromDescribeHistoryHostResponse(t.Success),
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    FromAccessDeniedError(t.AccessDeniedError),
	}
}

// ToHistoryServiceDescribeHistoryHostResult converts thrift HistoryService_DescribeHistoryHost_Result type to internal
func ToHistoryServiceDescribeHistoryHostResult(t *history.HistoryService_DescribeHistoryHost_Result) *types.HistoryServiceDescribeHistoryHostResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceDescribeHistoryHostResult{
		Success:              ToDescribeHistoryHostResponse(t.Success),
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    ToAccessDeniedError(t.AccessDeniedError),
	}
}

// FromHistoryServiceDescribeMutableStateArgs converts internal HistoryService_DescribeMutableState_Args type to thrift
func FromHistoryServiceDescribeMutableStateArgs(t *types.HistoryServiceDescribeMutableStateArgs) *history.HistoryService_DescribeMutableState_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_DescribeMutableState_Args{
		Request: FromDescribeMutableStateRequest(t.Request),
	}
}

// ToHistoryServiceDescribeMutableStateArgs converts thrift HistoryService_DescribeMutableState_Args type to internal
func ToHistoryServiceDescribeMutableStateArgs(t *history.HistoryService_DescribeMutableState_Args) *types.HistoryServiceDescribeMutableStateArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceDescribeMutableStateArgs{
		Request: ToDescribeMutableStateRequest(t.Request),
	}
}

// FromHistoryServiceDescribeMutableStateResult converts internal HistoryService_DescribeMutableState_Result type to thrift
func FromHistoryServiceDescribeMutableStateResult(t *types.HistoryServiceDescribeMutableStateResult) *history.HistoryService_DescribeMutableState_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_DescribeMutableState_Result{
		Success:                 FromDescribeMutableStateResponse(t.Success),
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		AccessDeniedError:       FromAccessDeniedError(t.AccessDeniedError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
	}
}

// ToHistoryServiceDescribeMutableStateResult converts thrift HistoryService_DescribeMutableState_Result type to internal
func ToHistoryServiceDescribeMutableStateResult(t *history.HistoryService_DescribeMutableState_Result) *types.HistoryServiceDescribeMutableStateResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceDescribeMutableStateResult{
		Success:                 ToDescribeMutableStateResponse(t.Success),
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		AccessDeniedError:       ToAccessDeniedError(t.AccessDeniedError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
	}
}

// FromHistoryServiceDescribeQueueArgs converts internal HistoryService_DescribeQueue_Args type to thrift
func FromHistoryServiceDescribeQueueArgs(t *types.HistoryServiceDescribeQueueArgs) *history.HistoryService_DescribeQueue_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_DescribeQueue_Args{
		Request: FromDescribeQueueRequest(t.Request),
	}
}

// ToHistoryServiceDescribeQueueArgs converts thrift HistoryService_DescribeQueue_Args type to internal
func ToHistoryServiceDescribeQueueArgs(t *history.HistoryService_DescribeQueue_Args) *types.HistoryServiceDescribeQueueArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceDescribeQueueArgs{
		Request: ToDescribeQueueRequest(t.Request),
	}
}

// FromHistoryServiceDescribeQueueResult converts internal HistoryService_DescribeQueue_Result type to thrift
func FromHistoryServiceDescribeQueueResult(t *types.HistoryServiceDescribeQueueResult) *history.HistoryService_DescribeQueue_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_DescribeQueue_Result{
		Success:              FromDescribeQueueResponse(t.Success),
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    FromAccessDeniedError(t.AccessDeniedError),
	}
}

// ToHistoryServiceDescribeQueueResult converts thrift HistoryService_DescribeQueue_Result type to internal
func ToHistoryServiceDescribeQueueResult(t *history.HistoryService_DescribeQueue_Result) *types.HistoryServiceDescribeQueueResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceDescribeQueueResult{
		Success:              ToDescribeQueueResponse(t.Success),
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    ToAccessDeniedError(t.AccessDeniedError),
	}
}

// FromHistoryServiceDescribeWorkflowExecutionArgs converts internal HistoryService_DescribeWorkflowExecution_Args type to thrift
func FromHistoryServiceDescribeWorkflowExecutionArgs(t *types.HistoryServiceDescribeWorkflowExecutionArgs) *history.HistoryService_DescribeWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_DescribeWorkflowExecution_Args{
		DescribeRequest: FromHistoryDescribeWorkflowExecutionRequest(t.DescribeRequest),
	}
}

// ToHistoryServiceDescribeWorkflowExecutionArgs converts thrift HistoryService_DescribeWorkflowExecution_Args type to internal
func ToHistoryServiceDescribeWorkflowExecutionArgs(t *history.HistoryService_DescribeWorkflowExecution_Args) *types.HistoryServiceDescribeWorkflowExecutionArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceDescribeWorkflowExecutionArgs{
		DescribeRequest: ToHistoryDescribeWorkflowExecutionRequest(t.DescribeRequest),
	}
}

// FromHistoryServiceDescribeWorkflowExecutionResult converts internal HistoryService_DescribeWorkflowExecution_Result type to thrift
func FromHistoryServiceDescribeWorkflowExecutionResult(t *types.HistoryServiceDescribeWorkflowExecutionResult) *history.HistoryService_DescribeWorkflowExecution_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_DescribeWorkflowExecution_Result{
		Success:                 FromDescribeWorkflowExecutionResponse(t.Success),
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceDescribeWorkflowExecutionResult converts thrift HistoryService_DescribeWorkflowExecution_Result type to internal
func ToHistoryServiceDescribeWorkflowExecutionResult(t *history.HistoryService_DescribeWorkflowExecution_Result) *types.HistoryServiceDescribeWorkflowExecutionResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceDescribeWorkflowExecutionResult{
		Success:                 ToDescribeWorkflowExecutionResponse(t.Success),
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceGetDLQReplicationMessagesArgs converts internal HistoryService_GetDLQReplicationMessages_Args type to thrift
func FromHistoryServiceGetDLQReplicationMessagesArgs(t *types.HistoryServiceGetDLQReplicationMessagesArgs) *history.HistoryService_GetDLQReplicationMessages_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_GetDLQReplicationMessages_Args{
		Request: FromGetDLQReplicationMessagesRequest(t.Request),
	}
}

// ToHistoryServiceGetDLQReplicationMessagesArgs converts thrift HistoryService_GetDLQReplicationMessages_Args type to internal
func ToHistoryServiceGetDLQReplicationMessagesArgs(t *history.HistoryService_GetDLQReplicationMessages_Args) *types.HistoryServiceGetDLQReplicationMessagesArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceGetDLQReplicationMessagesArgs{
		Request: ToGetDLQReplicationMessagesRequest(t.Request),
	}
}

// FromHistoryServiceGetDLQReplicationMessagesResult converts internal HistoryService_GetDLQReplicationMessages_Result type to thrift
func FromHistoryServiceGetDLQReplicationMessagesResult(t *types.HistoryServiceGetDLQReplicationMessagesResult) *history.HistoryService_GetDLQReplicationMessages_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_GetDLQReplicationMessages_Result{
		Success:              FromGetDLQReplicationMessagesResponse(t.Success),
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:  FromEntityNotExistsError(t.EntityNotExistError),
	}
}

// ToHistoryServiceGetDLQReplicationMessagesResult converts thrift HistoryService_GetDLQReplicationMessages_Result type to internal
func ToHistoryServiceGetDLQReplicationMessagesResult(t *history.HistoryService_GetDLQReplicationMessages_Result) *types.HistoryServiceGetDLQReplicationMessagesResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceGetDLQReplicationMessagesResult{
		Success:              ToGetDLQReplicationMessagesResponse(t.Success),
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:  ToEntityNotExistsError(t.EntityNotExistError),
	}
}

// FromHistoryServiceGetMutableStateArgs converts internal HistoryService_GetMutableState_Args type to thrift
func FromHistoryServiceGetMutableStateArgs(t *types.HistoryServiceGetMutableStateArgs) *history.HistoryService_GetMutableState_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_GetMutableState_Args{
		GetRequest: FromGetMutableStateRequest(t.GetRequest),
	}
}

// ToHistoryServiceGetMutableStateArgs converts thrift HistoryService_GetMutableState_Args type to internal
func ToHistoryServiceGetMutableStateArgs(t *history.HistoryService_GetMutableState_Args) *types.HistoryServiceGetMutableStateArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceGetMutableStateArgs{
		GetRequest: ToGetMutableStateRequest(t.GetRequest),
	}
}

// FromHistoryServiceGetMutableStateResult converts internal HistoryService_GetMutableState_Result type to thrift
func FromHistoryServiceGetMutableStateResult(t *types.HistoryServiceGetMutableStateResult) *history.HistoryService_GetMutableState_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_GetMutableState_Result{
		Success:                   FromGetMutableStateResponse(t.Success),
		BadRequestError:           FromBadRequestError(t.BadRequestError),
		InternalServiceError:      FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:       FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError:   FromShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:        FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:          FromServiceBusyError(t.ServiceBusyError),
		CurrentBranchChangedError: FromCurrentBranchChangedError(t.CurrentBranchChangedError),
	}
}

// ToHistoryServiceGetMutableStateResult converts thrift HistoryService_GetMutableState_Result type to internal
func ToHistoryServiceGetMutableStateResult(t *history.HistoryService_GetMutableState_Result) *types.HistoryServiceGetMutableStateResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceGetMutableStateResult{
		Success:                   ToGetMutableStateResponse(t.Success),
		BadRequestError:           ToBadRequestError(t.BadRequestError),
		InternalServiceError:      ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:       ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError:   ToShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:        ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:          ToServiceBusyError(t.ServiceBusyError),
		CurrentBranchChangedError: ToCurrentBranchChangedError(t.CurrentBranchChangedError),
	}
}

// FromHistoryServiceGetReplicationMessagesArgs converts internal HistoryService_GetReplicationMessages_Args type to thrift
func FromHistoryServiceGetReplicationMessagesArgs(t *types.HistoryServiceGetReplicationMessagesArgs) *history.HistoryService_GetReplicationMessages_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_GetReplicationMessages_Args{
		Request: FromGetReplicationMessagesRequest(t.Request),
	}
}

// ToHistoryServiceGetReplicationMessagesArgs converts thrift HistoryService_GetReplicationMessages_Args type to internal
func ToHistoryServiceGetReplicationMessagesArgs(t *history.HistoryService_GetReplicationMessages_Args) *types.HistoryServiceGetReplicationMessagesArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceGetReplicationMessagesArgs{
		Request: ToGetReplicationMessagesRequest(t.Request),
	}
}

// FromHistoryServiceGetReplicationMessagesResult converts internal HistoryService_GetReplicationMessages_Result type to thrift
func FromHistoryServiceGetReplicationMessagesResult(t *types.HistoryServiceGetReplicationMessagesResult) *history.HistoryService_GetReplicationMessages_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_GetReplicationMessages_Result{
		Success:                        FromGetReplicationMessagesResponse(t.Success),
		BadRequestError:                FromBadRequestError(t.BadRequestError),
		InternalServiceError:           FromInternalServiceError(t.InternalServiceError),
		LimitExceededError:             FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:               FromServiceBusyError(t.ServiceBusyError),
		ClientVersionNotSupportedError: FromClientVersionNotSupportedError(t.ClientVersionNotSupportedError),
	}
}

// ToHistoryServiceGetReplicationMessagesResult converts thrift HistoryService_GetReplicationMessages_Result type to internal
func ToHistoryServiceGetReplicationMessagesResult(t *history.HistoryService_GetReplicationMessages_Result) *types.HistoryServiceGetReplicationMessagesResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceGetReplicationMessagesResult{
		Success:                        ToGetReplicationMessagesResponse(t.Success),
		BadRequestError:                ToBadRequestError(t.BadRequestError),
		InternalServiceError:           ToInternalServiceError(t.InternalServiceError),
		LimitExceededError:             ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:               ToServiceBusyError(t.ServiceBusyError),
		ClientVersionNotSupportedError: ToClientVersionNotSupportedError(t.ClientVersionNotSupportedError),
	}
}

// FromHistoryServiceMergeDLQMessagesArgs converts internal HistoryService_MergeDLQMessages_Args type to thrift
func FromHistoryServiceMergeDLQMessagesArgs(t *types.HistoryServiceMergeDLQMessagesArgs) *history.HistoryService_MergeDLQMessages_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_MergeDLQMessages_Args{
		Request: FromMergeDLQMessagesRequest(t.Request),
	}
}

// ToHistoryServiceMergeDLQMessagesArgs converts thrift HistoryService_MergeDLQMessages_Args type to internal
func ToHistoryServiceMergeDLQMessagesArgs(t *history.HistoryService_MergeDLQMessages_Args) *types.HistoryServiceMergeDLQMessagesArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceMergeDLQMessagesArgs{
		Request: ToMergeDLQMessagesRequest(t.Request),
	}
}

// FromHistoryServiceMergeDLQMessagesResult converts internal HistoryService_MergeDLQMessages_Result type to thrift
func FromHistoryServiceMergeDLQMessagesResult(t *types.HistoryServiceMergeDLQMessagesResult) *history.HistoryService_MergeDLQMessages_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_MergeDLQMessages_Result{
		Success:                 FromMergeDLQMessagesResponse(t.Success),
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
	}
}

// ToHistoryServiceMergeDLQMessagesResult converts thrift HistoryService_MergeDLQMessages_Result type to internal
func ToHistoryServiceMergeDLQMessagesResult(t *history.HistoryService_MergeDLQMessages_Result) *types.HistoryServiceMergeDLQMessagesResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceMergeDLQMessagesResult{
		Success:                 ToMergeDLQMessagesResponse(t.Success),
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
	}
}

// FromHistoryServiceNotifyFailoverMarkersArgs converts internal HistoryService_NotifyFailoverMarkers_Args type to thrift
func FromHistoryServiceNotifyFailoverMarkersArgs(t *types.HistoryServiceNotifyFailoverMarkersArgs) *history.HistoryService_NotifyFailoverMarkers_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_NotifyFailoverMarkers_Args{
		Request: FromNotifyFailoverMarkersRequest(t.Request),
	}
}

// ToHistoryServiceNotifyFailoverMarkersArgs converts thrift HistoryService_NotifyFailoverMarkers_Args type to internal
func ToHistoryServiceNotifyFailoverMarkersArgs(t *history.HistoryService_NotifyFailoverMarkers_Args) *types.HistoryServiceNotifyFailoverMarkersArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceNotifyFailoverMarkersArgs{
		Request: ToNotifyFailoverMarkersRequest(t.Request),
	}
}

// FromHistoryServiceNotifyFailoverMarkersResult converts internal HistoryService_NotifyFailoverMarkers_Result type to thrift
func FromHistoryServiceNotifyFailoverMarkersResult(t *types.HistoryServiceNotifyFailoverMarkersResult) *history.HistoryService_NotifyFailoverMarkers_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_NotifyFailoverMarkers_Result{
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceNotifyFailoverMarkersResult converts thrift HistoryService_NotifyFailoverMarkers_Result type to internal
func ToHistoryServiceNotifyFailoverMarkersResult(t *history.HistoryService_NotifyFailoverMarkers_Result) *types.HistoryServiceNotifyFailoverMarkersResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceNotifyFailoverMarkersResult{
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServicePollMutableStateArgs converts internal HistoryService_PollMutableState_Args type to thrift
func FromHistoryServicePollMutableStateArgs(t *types.HistoryServicePollMutableStateArgs) *history.HistoryService_PollMutableState_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_PollMutableState_Args{
		PollRequest: FromPollMutableStateRequest(t.PollRequest),
	}
}

// ToHistoryServicePollMutableStateArgs converts thrift HistoryService_PollMutableState_Args type to internal
func ToHistoryServicePollMutableStateArgs(t *history.HistoryService_PollMutableState_Args) *types.HistoryServicePollMutableStateArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServicePollMutableStateArgs{
		PollRequest: ToPollMutableStateRequest(t.PollRequest),
	}
}

// FromHistoryServicePollMutableStateResult converts internal HistoryService_PollMutableState_Result type to thrift
func FromHistoryServicePollMutableStateResult(t *types.HistoryServicePollMutableStateResult) *history.HistoryService_PollMutableState_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_PollMutableState_Result{
		Success:                   FromPollMutableStateResponse(t.Success),
		BadRequestError:           FromBadRequestError(t.BadRequestError),
		InternalServiceError:      FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:       FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError:   FromShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:        FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:          FromServiceBusyError(t.ServiceBusyError),
		CurrentBranchChangedError: FromCurrentBranchChangedError(t.CurrentBranchChangedError),
	}
}

// ToHistoryServicePollMutableStateResult converts thrift HistoryService_PollMutableState_Result type to internal
func ToHistoryServicePollMutableStateResult(t *history.HistoryService_PollMutableState_Result) *types.HistoryServicePollMutableStateResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServicePollMutableStateResult{
		Success:                   ToPollMutableStateResponse(t.Success),
		BadRequestError:           ToBadRequestError(t.BadRequestError),
		InternalServiceError:      ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:       ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError:   ToShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:        ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:          ToServiceBusyError(t.ServiceBusyError),
		CurrentBranchChangedError: ToCurrentBranchChangedError(t.CurrentBranchChangedError),
	}
}

// FromHistoryServicePurgeDLQMessagesArgs converts internal HistoryService_PurgeDLQMessages_Args type to thrift
func FromHistoryServicePurgeDLQMessagesArgs(t *types.HistoryServicePurgeDLQMessagesArgs) *history.HistoryService_PurgeDLQMessages_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_PurgeDLQMessages_Args{
		Request: FromPurgeDLQMessagesRequest(t.Request),
	}
}

// ToHistoryServicePurgeDLQMessagesArgs converts thrift HistoryService_PurgeDLQMessages_Args type to internal
func ToHistoryServicePurgeDLQMessagesArgs(t *history.HistoryService_PurgeDLQMessages_Args) *types.HistoryServicePurgeDLQMessagesArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServicePurgeDLQMessagesArgs{
		Request: ToPurgeDLQMessagesRequest(t.Request),
	}
}

// FromHistoryServicePurgeDLQMessagesResult converts internal HistoryService_PurgeDLQMessages_Result type to thrift
func FromHistoryServicePurgeDLQMessagesResult(t *types.HistoryServicePurgeDLQMessagesResult) *history.HistoryService_PurgeDLQMessages_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_PurgeDLQMessages_Result{
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
	}
}

// ToHistoryServicePurgeDLQMessagesResult converts thrift HistoryService_PurgeDLQMessages_Result type to internal
func ToHistoryServicePurgeDLQMessagesResult(t *history.HistoryService_PurgeDLQMessages_Result) *types.HistoryServicePurgeDLQMessagesResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServicePurgeDLQMessagesResult{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
	}
}

// FromHistoryServiceQueryWorkflowArgs converts internal HistoryService_QueryWorkflow_Args type to thrift
func FromHistoryServiceQueryWorkflowArgs(t *types.HistoryServiceQueryWorkflowArgs) *history.HistoryService_QueryWorkflow_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_QueryWorkflow_Args{
		QueryRequest: FromHistoryQueryWorkflowRequest(t.QueryRequest),
	}
}

// ToHistoryServiceQueryWorkflowArgs converts thrift HistoryService_QueryWorkflow_Args type to internal
func ToHistoryServiceQueryWorkflowArgs(t *history.HistoryService_QueryWorkflow_Args) *types.HistoryServiceQueryWorkflowArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceQueryWorkflowArgs{
		QueryRequest: ToHistoryQueryWorkflowRequest(t.QueryRequest),
	}
}

// FromHistoryServiceQueryWorkflowResult converts internal HistoryService_QueryWorkflow_Result type to thrift
func FromHistoryServiceQueryWorkflowResult(t *types.HistoryServiceQueryWorkflowResult) *history.HistoryService_QueryWorkflow_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_QueryWorkflow_Result{
		Success:                        FromHistoryQueryWorkflowResponse(t.Success),
		BadRequestError:                FromBadRequestError(t.BadRequestError),
		InternalServiceError:           FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:            FromEntityNotExistsError(t.EntityNotExistError),
		QueryFailedError:               FromQueryFailedError(t.QueryFailedError),
		LimitExceededError:             FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:               FromServiceBusyError(t.ServiceBusyError),
		ClientVersionNotSupportedError: FromClientVersionNotSupportedError(t.ClientVersionNotSupportedError),
	}
}

// ToHistoryServiceQueryWorkflowResult converts thrift HistoryService_QueryWorkflow_Result type to internal
func ToHistoryServiceQueryWorkflowResult(t *history.HistoryService_QueryWorkflow_Result) *types.HistoryServiceQueryWorkflowResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceQueryWorkflowResult{
		Success:                        ToHistoryQueryWorkflowResponse(t.Success),
		BadRequestError:                ToBadRequestError(t.BadRequestError),
		InternalServiceError:           ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:            ToEntityNotExistsError(t.EntityNotExistError),
		QueryFailedError:               ToQueryFailedError(t.QueryFailedError),
		LimitExceededError:             ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:               ToServiceBusyError(t.ServiceBusyError),
		ClientVersionNotSupportedError: ToClientVersionNotSupportedError(t.ClientVersionNotSupportedError),
	}
}

// FromHistoryServiceReadDLQMessagesArgs converts internal HistoryService_ReadDLQMessages_Args type to thrift
func FromHistoryServiceReadDLQMessagesArgs(t *types.HistoryServiceReadDLQMessagesArgs) *history.HistoryService_ReadDLQMessages_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ReadDLQMessages_Args{
		Request: FromReadDLQMessagesRequest(t.Request),
	}
}

// ToHistoryServiceReadDLQMessagesArgs converts thrift HistoryService_ReadDLQMessages_Args type to internal
func ToHistoryServiceReadDLQMessagesArgs(t *history.HistoryService_ReadDLQMessages_Args) *types.HistoryServiceReadDLQMessagesArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceReadDLQMessagesArgs{
		Request: ToReadDLQMessagesRequest(t.Request),
	}
}

// FromHistoryServiceReadDLQMessagesResult converts internal HistoryService_ReadDLQMessages_Result type to thrift
func FromHistoryServiceReadDLQMessagesResult(t *types.HistoryServiceReadDLQMessagesResult) *history.HistoryService_ReadDLQMessages_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ReadDLQMessages_Result{
		Success:                 FromReadDLQMessagesResponse(t.Success),
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
	}
}

// ToHistoryServiceReadDLQMessagesResult converts thrift HistoryService_ReadDLQMessages_Result type to internal
func ToHistoryServiceReadDLQMessagesResult(t *history.HistoryService_ReadDLQMessages_Result) *types.HistoryServiceReadDLQMessagesResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceReadDLQMessagesResult{
		Success:                 ToReadDLQMessagesResponse(t.Success),
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
	}
}

// FromHistoryServiceReapplyEventsArgs converts internal HistoryService_ReapplyEvents_Args type to thrift
func FromHistoryServiceReapplyEventsArgs(t *types.HistoryServiceReapplyEventsArgs) *history.HistoryService_ReapplyEvents_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ReapplyEvents_Args{
		ReapplyEventsRequest: FromHistoryReapplyEventsRequest(t.ReapplyEventsRequest),
	}
}

// ToHistoryServiceReapplyEventsArgs converts thrift HistoryService_ReapplyEvents_Args type to internal
func ToHistoryServiceReapplyEventsArgs(t *history.HistoryService_ReapplyEvents_Args) *types.HistoryServiceReapplyEventsArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceReapplyEventsArgs{
		ReapplyEventsRequest: ToHistoryReapplyEventsRequest(t.ReapplyEventsRequest),
	}
}

// FromHistoryServiceReapplyEventsResult converts internal HistoryService_ReapplyEvents_Result type to thrift
func FromHistoryServiceReapplyEventsResult(t *types.HistoryServiceReapplyEventsResult) *history.HistoryService_ReapplyEvents_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ReapplyEvents_Result{
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		DomainNotActiveError:    FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
	}
}

// ToHistoryServiceReapplyEventsResult converts thrift HistoryService_ReapplyEvents_Result type to internal
func ToHistoryServiceReapplyEventsResult(t *history.HistoryService_ReapplyEvents_Result) *types.HistoryServiceReapplyEventsResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceReapplyEventsResult{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
	}
}

// FromHistoryServiceRecordActivityTaskHeartbeatArgs converts internal HistoryService_RecordActivityTaskHeartbeat_Args type to thrift
func FromHistoryServiceRecordActivityTaskHeartbeatArgs(t *types.HistoryServiceRecordActivityTaskHeartbeatArgs) *history.HistoryService_RecordActivityTaskHeartbeat_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RecordActivityTaskHeartbeat_Args{
		HeartbeatRequest: FromHistoryRecordActivityTaskHeartbeatRequest(t.HeartbeatRequest),
	}
}

// ToHistoryServiceRecordActivityTaskHeartbeatArgs converts thrift HistoryService_RecordActivityTaskHeartbeat_Args type to internal
func ToHistoryServiceRecordActivityTaskHeartbeatArgs(t *history.HistoryService_RecordActivityTaskHeartbeat_Args) *types.HistoryServiceRecordActivityTaskHeartbeatArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRecordActivityTaskHeartbeatArgs{
		HeartbeatRequest: ToHistoryRecordActivityTaskHeartbeatRequest(t.HeartbeatRequest),
	}
}

// FromHistoryServiceRecordActivityTaskHeartbeatResult converts internal HistoryService_RecordActivityTaskHeartbeat_Result type to thrift
func FromHistoryServiceRecordActivityTaskHeartbeatResult(t *types.HistoryServiceRecordActivityTaskHeartbeatResult) *history.HistoryService_RecordActivityTaskHeartbeat_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RecordActivityTaskHeartbeat_Result{
		Success:                 FromRecordActivityTaskHeartbeatResponse(t.Success),
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceRecordActivityTaskHeartbeatResult converts thrift HistoryService_RecordActivityTaskHeartbeat_Result type to internal
func ToHistoryServiceRecordActivityTaskHeartbeatResult(t *history.HistoryService_RecordActivityTaskHeartbeat_Result) *types.HistoryServiceRecordActivityTaskHeartbeatResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRecordActivityTaskHeartbeatResult{
		Success:                 ToRecordActivityTaskHeartbeatResponse(t.Success),
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceRecordActivityTaskStartedArgs converts internal HistoryService_RecordActivityTaskStarted_Args type to thrift
func FromHistoryServiceRecordActivityTaskStartedArgs(t *types.HistoryServiceRecordActivityTaskStartedArgs) *history.HistoryService_RecordActivityTaskStarted_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RecordActivityTaskStarted_Args{
		AddRequest: FromRecordActivityTaskStartedRequest(t.AddRequest),
	}
}

// ToHistoryServiceRecordActivityTaskStartedArgs converts thrift HistoryService_RecordActivityTaskStarted_Args type to internal
func ToHistoryServiceRecordActivityTaskStartedArgs(t *history.HistoryService_RecordActivityTaskStarted_Args) *types.HistoryServiceRecordActivityTaskStartedArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRecordActivityTaskStartedArgs{
		AddRequest: ToRecordActivityTaskStartedRequest(t.AddRequest),
	}
}

// FromHistoryServiceRecordActivityTaskStartedResult converts internal HistoryService_RecordActivityTaskStarted_Result type to thrift
func FromHistoryServiceRecordActivityTaskStartedResult(t *types.HistoryServiceRecordActivityTaskStartedResult) *history.HistoryService_RecordActivityTaskStarted_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RecordActivityTaskStarted_Result{
		Success:                  FromRecordActivityTaskStartedResponse(t.Success),
		BadRequestError:          FromBadRequestError(t.BadRequestError),
		InternalServiceError:     FromInternalServiceError(t.InternalServiceError),
		EventAlreadyStartedError: FromEventAlreadyStartedError(t.EventAlreadyStartedError),
		EntityNotExistError:      FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError:  FromShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:     FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:       FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:         FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceRecordActivityTaskStartedResult converts thrift HistoryService_RecordActivityTaskStarted_Result type to internal
func ToHistoryServiceRecordActivityTaskStartedResult(t *history.HistoryService_RecordActivityTaskStarted_Result) *types.HistoryServiceRecordActivityTaskStartedResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRecordActivityTaskStartedResult{
		Success:                  ToRecordActivityTaskStartedResponse(t.Success),
		BadRequestError:          ToBadRequestError(t.BadRequestError),
		InternalServiceError:     ToInternalServiceError(t.InternalServiceError),
		EventAlreadyStartedError: ToEventAlreadyStartedError(t.EventAlreadyStartedError),
		EntityNotExistError:      ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError:  ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:     ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:       ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:         ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceRecordChildExecutionCompletedArgs converts internal HistoryService_RecordChildExecutionCompleted_Args type to thrift
func FromHistoryServiceRecordChildExecutionCompletedArgs(t *types.HistoryServiceRecordChildExecutionCompletedArgs) *history.HistoryService_RecordChildExecutionCompleted_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RecordChildExecutionCompleted_Args{
		CompletionRequest: FromRecordChildExecutionCompletedRequest(t.CompletionRequest),
	}
}

// ToHistoryServiceRecordChildExecutionCompletedArgs converts thrift HistoryService_RecordChildExecutionCompleted_Args type to internal
func ToHistoryServiceRecordChildExecutionCompletedArgs(t *history.HistoryService_RecordChildExecutionCompleted_Args) *types.HistoryServiceRecordChildExecutionCompletedArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRecordChildExecutionCompletedArgs{
		CompletionRequest: ToRecordChildExecutionCompletedRequest(t.CompletionRequest),
	}
}

// FromHistoryServiceRecordChildExecutionCompletedResult converts internal HistoryService_RecordChildExecutionCompleted_Result type to thrift
func FromHistoryServiceRecordChildExecutionCompletedResult(t *types.HistoryServiceRecordChildExecutionCompletedResult) *history.HistoryService_RecordChildExecutionCompleted_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RecordChildExecutionCompleted_Result{
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceRecordChildExecutionCompletedResult converts thrift HistoryService_RecordChildExecutionCompleted_Result type to internal
func ToHistoryServiceRecordChildExecutionCompletedResult(t *history.HistoryService_RecordChildExecutionCompleted_Result) *types.HistoryServiceRecordChildExecutionCompletedResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRecordChildExecutionCompletedResult{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceRecordDecisionTaskStartedArgs converts internal HistoryService_RecordDecisionTaskStarted_Args type to thrift
func FromHistoryServiceRecordDecisionTaskStartedArgs(t *types.HistoryServiceRecordDecisionTaskStartedArgs) *history.HistoryService_RecordDecisionTaskStarted_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RecordDecisionTaskStarted_Args{
		AddRequest: FromRecordDecisionTaskStartedRequest(t.AddRequest),
	}
}

// ToHistoryServiceRecordDecisionTaskStartedArgs converts thrift HistoryService_RecordDecisionTaskStarted_Args type to internal
func ToHistoryServiceRecordDecisionTaskStartedArgs(t *history.HistoryService_RecordDecisionTaskStarted_Args) *types.HistoryServiceRecordDecisionTaskStartedArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRecordDecisionTaskStartedArgs{
		AddRequest: ToRecordDecisionTaskStartedRequest(t.AddRequest),
	}
}

// FromHistoryServiceRecordDecisionTaskStartedResult converts internal HistoryService_RecordDecisionTaskStarted_Result type to thrift
func FromHistoryServiceRecordDecisionTaskStartedResult(t *types.HistoryServiceRecordDecisionTaskStartedResult) *history.HistoryService_RecordDecisionTaskStarted_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RecordDecisionTaskStarted_Result{
		Success:                  FromRecordDecisionTaskStartedResponse(t.Success),
		BadRequestError:          FromBadRequestError(t.BadRequestError),
		InternalServiceError:     FromInternalServiceError(t.InternalServiceError),
		EventAlreadyStartedError: FromEventAlreadyStartedError(t.EventAlreadyStartedError),
		EntityNotExistError:      FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError:  FromShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:     FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:       FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:         FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceRecordDecisionTaskStartedResult converts thrift HistoryService_RecordDecisionTaskStarted_Result type to internal
func ToHistoryServiceRecordDecisionTaskStartedResult(t *history.HistoryService_RecordDecisionTaskStarted_Result) *types.HistoryServiceRecordDecisionTaskStartedResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRecordDecisionTaskStartedResult{
		Success:                  ToRecordDecisionTaskStartedResponse(t.Success),
		BadRequestError:          ToBadRequestError(t.BadRequestError),
		InternalServiceError:     ToInternalServiceError(t.InternalServiceError),
		EventAlreadyStartedError: ToEventAlreadyStartedError(t.EventAlreadyStartedError),
		EntityNotExistError:      ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError:  ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:     ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:       ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:         ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceRefreshWorkflowTasksArgs converts internal HistoryService_RefreshWorkflowTasks_Args type to thrift
func FromHistoryServiceRefreshWorkflowTasksArgs(t *types.HistoryServiceRefreshWorkflowTasksArgs) *history.HistoryService_RefreshWorkflowTasks_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RefreshWorkflowTasks_Args{
		Request: FromHistoryRefreshWorkflowTasksRequest(t.Request),
	}
}

// ToHistoryServiceRefreshWorkflowTasksArgs converts thrift HistoryService_RefreshWorkflowTasks_Args type to internal
func ToHistoryServiceRefreshWorkflowTasksArgs(t *history.HistoryService_RefreshWorkflowTasks_Args) *types.HistoryServiceRefreshWorkflowTasksArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRefreshWorkflowTasksArgs{
		Request: ToHistoryRefreshWorkflowTasksRequest(t.Request),
	}
}

// FromHistoryServiceRefreshWorkflowTasksResult converts internal HistoryService_RefreshWorkflowTasks_Result type to thrift
func FromHistoryServiceRefreshWorkflowTasksResult(t *types.HistoryServiceRefreshWorkflowTasksResult) *history.HistoryService_RefreshWorkflowTasks_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RefreshWorkflowTasks_Result{
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		DomainNotActiveError:    FromDomainNotActiveError(t.DomainNotActiveError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
	}
}

// ToHistoryServiceRefreshWorkflowTasksResult converts thrift HistoryService_RefreshWorkflowTasks_Result type to internal
func ToHistoryServiceRefreshWorkflowTasksResult(t *history.HistoryService_RefreshWorkflowTasks_Result) *types.HistoryServiceRefreshWorkflowTasksResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRefreshWorkflowTasksResult{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
	}
}

// FromHistoryServiceRemoveSignalMutableStateArgs converts internal HistoryService_RemoveSignalMutableState_Args type to thrift
func FromHistoryServiceRemoveSignalMutableStateArgs(t *types.HistoryServiceRemoveSignalMutableStateArgs) *history.HistoryService_RemoveSignalMutableState_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RemoveSignalMutableState_Args{
		RemoveRequest: FromRemoveSignalMutableStateRequest(t.RemoveRequest),
	}
}

// ToHistoryServiceRemoveSignalMutableStateArgs converts thrift HistoryService_RemoveSignalMutableState_Args type to internal
func ToHistoryServiceRemoveSignalMutableStateArgs(t *history.HistoryService_RemoveSignalMutableState_Args) *types.HistoryServiceRemoveSignalMutableStateArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRemoveSignalMutableStateArgs{
		RemoveRequest: ToRemoveSignalMutableStateRequest(t.RemoveRequest),
	}
}

// FromHistoryServiceRemoveSignalMutableStateResult converts internal HistoryService_RemoveSignalMutableState_Result type to thrift
func FromHistoryServiceRemoveSignalMutableStateResult(t *types.HistoryServiceRemoveSignalMutableStateResult) *history.HistoryService_RemoveSignalMutableState_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RemoveSignalMutableState_Result{
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceRemoveSignalMutableStateResult converts thrift HistoryService_RemoveSignalMutableState_Result type to internal
func ToHistoryServiceRemoveSignalMutableStateResult(t *history.HistoryService_RemoveSignalMutableState_Result) *types.HistoryServiceRemoveSignalMutableStateResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRemoveSignalMutableStateResult{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceRemoveTaskArgs converts internal HistoryService_RemoveTask_Args type to thrift
func FromHistoryServiceRemoveTaskArgs(t *types.HistoryServiceRemoveTaskArgs) *history.HistoryService_RemoveTask_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RemoveTask_Args{
		Request: FromRemoveTaskRequest(t.Request),
	}
}

// ToHistoryServiceRemoveTaskArgs converts thrift HistoryService_RemoveTask_Args type to internal
func ToHistoryServiceRemoveTaskArgs(t *history.HistoryService_RemoveTask_Args) *types.HistoryServiceRemoveTaskArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRemoveTaskArgs{
		Request: ToRemoveTaskRequest(t.Request),
	}
}

// FromHistoryServiceRemoveTaskResult converts internal HistoryService_RemoveTask_Result type to thrift
func FromHistoryServiceRemoveTaskResult(t *types.HistoryServiceRemoveTaskResult) *history.HistoryService_RemoveTask_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RemoveTask_Result{
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    FromAccessDeniedError(t.AccessDeniedError),
	}
}

// ToHistoryServiceRemoveTaskResult converts thrift HistoryService_RemoveTask_Result type to internal
func ToHistoryServiceRemoveTaskResult(t *history.HistoryService_RemoveTask_Result) *types.HistoryServiceRemoveTaskResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRemoveTaskResult{
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    ToAccessDeniedError(t.AccessDeniedError),
	}
}

// FromHistoryServiceReplicateEventsV2Args converts internal HistoryService_ReplicateEventsV2_Args type to thrift
func FromHistoryServiceReplicateEventsV2Args(t *types.HistoryServiceReplicateEventsV2Args) *history.HistoryService_ReplicateEventsV2_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ReplicateEventsV2_Args{
		ReplicateV2Request: FromReplicateEventsV2Request(t.ReplicateV2Request),
	}
}

// ToHistoryServiceReplicateEventsV2Args converts thrift HistoryService_ReplicateEventsV2_Args type to internal
func ToHistoryServiceReplicateEventsV2Args(t *history.HistoryService_ReplicateEventsV2_Args) *types.HistoryServiceReplicateEventsV2Args {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceReplicateEventsV2Args{
		ReplicateV2Request: ToReplicateEventsV2Request(t.ReplicateV2Request),
	}
}

// FromHistoryServiceReplicateEventsV2Result converts internal HistoryService_ReplicateEventsV2_Result type to thrift
func FromHistoryServiceReplicateEventsV2Result(t *types.HistoryServiceReplicateEventsV2Result) *history.HistoryService_ReplicateEventsV2_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ReplicateEventsV2_Result{
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
		RetryTaskError:          FromRetryTaskV2Error(t.RetryTaskError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceReplicateEventsV2Result converts thrift HistoryService_ReplicateEventsV2_Result type to internal
func ToHistoryServiceReplicateEventsV2Result(t *history.HistoryService_ReplicateEventsV2_Result) *types.HistoryServiceReplicateEventsV2Result {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceReplicateEventsV2Result{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		RetryTaskError:          ToRetryTaskV2Error(t.RetryTaskError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceRequestCancelWorkflowExecutionArgs converts internal HistoryService_RequestCancelWorkflowExecution_Args type to thrift
func FromHistoryServiceRequestCancelWorkflowExecutionArgs(t *types.HistoryServiceRequestCancelWorkflowExecutionArgs) *history.HistoryService_RequestCancelWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RequestCancelWorkflowExecution_Args{
		CancelRequest: FromHistoryRequestCancelWorkflowExecutionRequest(t.CancelRequest),
	}
}

// ToHistoryServiceRequestCancelWorkflowExecutionArgs converts thrift HistoryService_RequestCancelWorkflowExecution_Args type to internal
func ToHistoryServiceRequestCancelWorkflowExecutionArgs(t *history.HistoryService_RequestCancelWorkflowExecution_Args) *types.HistoryServiceRequestCancelWorkflowExecutionArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRequestCancelWorkflowExecutionArgs{
		CancelRequest: ToHistoryRequestCancelWorkflowExecutionRequest(t.CancelRequest),
	}
}

// FromHistoryServiceRequestCancelWorkflowExecutionResult converts internal HistoryService_RequestCancelWorkflowExecution_Result type to thrift
func FromHistoryServiceRequestCancelWorkflowExecutionResult(t *types.HistoryServiceRequestCancelWorkflowExecutionResult) *history.HistoryService_RequestCancelWorkflowExecution_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RequestCancelWorkflowExecution_Result{
		BadRequestError:                   FromBadRequestError(t.BadRequestError),
		InternalServiceError:              FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:               FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError:           FromShardOwnershipLostError(t.ShardOwnershipLostError),
		CancellationAlreadyRequestedError: FromCancellationAlreadyRequestedError(t.CancellationAlreadyRequestedError),
		DomainNotActiveError:              FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:                FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:                  FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceRequestCancelWorkflowExecutionResult converts thrift HistoryService_RequestCancelWorkflowExecution_Result type to internal
func ToHistoryServiceRequestCancelWorkflowExecutionResult(t *history.HistoryService_RequestCancelWorkflowExecution_Result) *types.HistoryServiceRequestCancelWorkflowExecutionResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRequestCancelWorkflowExecutionResult{
		BadRequestError:                   ToBadRequestError(t.BadRequestError),
		InternalServiceError:              ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:               ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError:           ToShardOwnershipLostError(t.ShardOwnershipLostError),
		CancellationAlreadyRequestedError: ToCancellationAlreadyRequestedError(t.CancellationAlreadyRequestedError),
		DomainNotActiveError:              ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:                ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:                  ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceResetQueueArgs converts internal HistoryService_ResetQueue_Args type to thrift
func FromHistoryServiceResetQueueArgs(t *types.HistoryServiceResetQueueArgs) *history.HistoryService_ResetQueue_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ResetQueue_Args{
		Request: FromResetQueueRequest(t.Request),
	}
}

// ToHistoryServiceResetQueueArgs converts thrift HistoryService_ResetQueue_Args type to internal
func ToHistoryServiceResetQueueArgs(t *history.HistoryService_ResetQueue_Args) *types.HistoryServiceResetQueueArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceResetQueueArgs{
		Request: ToResetQueueRequest(t.Request),
	}
}

// FromHistoryServiceResetQueueResult converts internal HistoryService_ResetQueue_Result type to thrift
func FromHistoryServiceResetQueueResult(t *types.HistoryServiceResetQueueResult) *history.HistoryService_ResetQueue_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ResetQueue_Result{
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    FromAccessDeniedError(t.AccessDeniedError),
	}
}

// ToHistoryServiceResetQueueResult converts thrift HistoryService_ResetQueue_Result type to internal
func ToHistoryServiceResetQueueResult(t *history.HistoryService_ResetQueue_Result) *types.HistoryServiceResetQueueResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceResetQueueResult{
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    ToAccessDeniedError(t.AccessDeniedError),
	}
}

// FromHistoryServiceResetStickyTaskListArgs converts internal HistoryService_ResetStickyTaskList_Args type to thrift
func FromHistoryServiceResetStickyTaskListArgs(t *types.HistoryServiceResetStickyTaskListArgs) *history.HistoryService_ResetStickyTaskList_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ResetStickyTaskList_Args{
		ResetRequest: FromHistoryResetStickyTaskListRequest(t.ResetRequest),
	}
}

// ToHistoryServiceResetStickyTaskListArgs converts thrift HistoryService_ResetStickyTaskList_Args type to internal
func ToHistoryServiceResetStickyTaskListArgs(t *history.HistoryService_ResetStickyTaskList_Args) *types.HistoryServiceResetStickyTaskListArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceResetStickyTaskListArgs{
		ResetRequest: ToHistoryResetStickyTaskListRequest(t.ResetRequest),
	}
}

// FromHistoryServiceResetStickyTaskListResult converts internal HistoryService_ResetStickyTaskList_Result type to thrift
func FromHistoryServiceResetStickyTaskListResult(t *types.HistoryServiceResetStickyTaskListResult) *history.HistoryService_ResetStickyTaskList_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ResetStickyTaskList_Result{
		Success:                 FromHistoryResetStickyTaskListResponse(t.Success),
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceResetStickyTaskListResult converts thrift HistoryService_ResetStickyTaskList_Result type to internal
func ToHistoryServiceResetStickyTaskListResult(t *history.HistoryService_ResetStickyTaskList_Result) *types.HistoryServiceResetStickyTaskListResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceResetStickyTaskListResult{
		Success:                 ToHistoryResetStickyTaskListResponse(t.Success),
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceResetWorkflowExecutionArgs converts internal HistoryService_ResetWorkflowExecution_Args type to thrift
func FromHistoryServiceResetWorkflowExecutionArgs(t *types.HistoryServiceResetWorkflowExecutionArgs) *history.HistoryService_ResetWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ResetWorkflowExecution_Args{
		ResetRequest: FromHistoryResetWorkflowExecutionRequest(t.ResetRequest),
	}
}

// ToHistoryServiceResetWorkflowExecutionArgs converts thrift HistoryService_ResetWorkflowExecution_Args type to internal
func ToHistoryServiceResetWorkflowExecutionArgs(t *history.HistoryService_ResetWorkflowExecution_Args) *types.HistoryServiceResetWorkflowExecutionArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceResetWorkflowExecutionArgs{
		ResetRequest: ToHistoryResetWorkflowExecutionRequest(t.ResetRequest),
	}
}

// FromHistoryServiceResetWorkflowExecutionResult converts internal HistoryService_ResetWorkflowExecution_Result type to thrift
func FromHistoryServiceResetWorkflowExecutionResult(t *types.HistoryServiceResetWorkflowExecutionResult) *history.HistoryService_ResetWorkflowExecution_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ResetWorkflowExecution_Result{
		Success:                 FromResetWorkflowExecutionResponse(t.Success),
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceResetWorkflowExecutionResult converts thrift HistoryService_ResetWorkflowExecution_Result type to internal
func ToHistoryServiceResetWorkflowExecutionResult(t *history.HistoryService_ResetWorkflowExecution_Result) *types.HistoryServiceResetWorkflowExecutionResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceResetWorkflowExecutionResult{
		Success:                 ToResetWorkflowExecutionResponse(t.Success),
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceRespondActivityTaskCanceledArgs converts internal HistoryService_RespondActivityTaskCanceled_Args type to thrift
func FromHistoryServiceRespondActivityTaskCanceledArgs(t *types.HistoryServiceRespondActivityTaskCanceledArgs) *history.HistoryService_RespondActivityTaskCanceled_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RespondActivityTaskCanceled_Args{
		CanceledRequest: FromHistoryRespondActivityTaskCanceledRequest(t.CanceledRequest),
	}
}

// ToHistoryServiceRespondActivityTaskCanceledArgs converts thrift HistoryService_RespondActivityTaskCanceled_Args type to internal
func ToHistoryServiceRespondActivityTaskCanceledArgs(t *history.HistoryService_RespondActivityTaskCanceled_Args) *types.HistoryServiceRespondActivityTaskCanceledArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRespondActivityTaskCanceledArgs{
		CanceledRequest: ToHistoryRespondActivityTaskCanceledRequest(t.CanceledRequest),
	}
}

// FromHistoryServiceRespondActivityTaskCanceledResult converts internal HistoryService_RespondActivityTaskCanceled_Result type to thrift
func FromHistoryServiceRespondActivityTaskCanceledResult(t *types.HistoryServiceRespondActivityTaskCanceledResult) *history.HistoryService_RespondActivityTaskCanceled_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RespondActivityTaskCanceled_Result{
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceRespondActivityTaskCanceledResult converts thrift HistoryService_RespondActivityTaskCanceled_Result type to internal
func ToHistoryServiceRespondActivityTaskCanceledResult(t *history.HistoryService_RespondActivityTaskCanceled_Result) *types.HistoryServiceRespondActivityTaskCanceledResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRespondActivityTaskCanceledResult{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceRespondActivityTaskCompletedArgs converts internal HistoryService_RespondActivityTaskCompleted_Args type to thrift
func FromHistoryServiceRespondActivityTaskCompletedArgs(t *types.HistoryServiceRespondActivityTaskCompletedArgs) *history.HistoryService_RespondActivityTaskCompleted_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RespondActivityTaskCompleted_Args{
		CompleteRequest: FromHistoryRespondActivityTaskCompletedRequest(t.CompleteRequest),
	}
}

// ToHistoryServiceRespondActivityTaskCompletedArgs converts thrift HistoryService_RespondActivityTaskCompleted_Args type to internal
func ToHistoryServiceRespondActivityTaskCompletedArgs(t *history.HistoryService_RespondActivityTaskCompleted_Args) *types.HistoryServiceRespondActivityTaskCompletedArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRespondActivityTaskCompletedArgs{
		CompleteRequest: ToHistoryRespondActivityTaskCompletedRequest(t.CompleteRequest),
	}
}

// FromHistoryServiceRespondActivityTaskCompletedResult converts internal HistoryService_RespondActivityTaskCompleted_Result type to thrift
func FromHistoryServiceRespondActivityTaskCompletedResult(t *types.HistoryServiceRespondActivityTaskCompletedResult) *history.HistoryService_RespondActivityTaskCompleted_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RespondActivityTaskCompleted_Result{
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceRespondActivityTaskCompletedResult converts thrift HistoryService_RespondActivityTaskCompleted_Result type to internal
func ToHistoryServiceRespondActivityTaskCompletedResult(t *history.HistoryService_RespondActivityTaskCompleted_Result) *types.HistoryServiceRespondActivityTaskCompletedResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRespondActivityTaskCompletedResult{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceRespondActivityTaskFailedArgs converts internal HistoryService_RespondActivityTaskFailed_Args type to thrift
func FromHistoryServiceRespondActivityTaskFailedArgs(t *types.HistoryServiceRespondActivityTaskFailedArgs) *history.HistoryService_RespondActivityTaskFailed_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RespondActivityTaskFailed_Args{
		FailRequest: FromHistoryRespondActivityTaskFailedRequest(t.FailRequest),
	}
}

// ToHistoryServiceRespondActivityTaskFailedArgs converts thrift HistoryService_RespondActivityTaskFailed_Args type to internal
func ToHistoryServiceRespondActivityTaskFailedArgs(t *history.HistoryService_RespondActivityTaskFailed_Args) *types.HistoryServiceRespondActivityTaskFailedArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRespondActivityTaskFailedArgs{
		FailRequest: ToHistoryRespondActivityTaskFailedRequest(t.FailRequest),
	}
}

// FromHistoryServiceRespondActivityTaskFailedResult converts internal HistoryService_RespondActivityTaskFailed_Result type to thrift
func FromHistoryServiceRespondActivityTaskFailedResult(t *types.HistoryServiceRespondActivityTaskFailedResult) *history.HistoryService_RespondActivityTaskFailed_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RespondActivityTaskFailed_Result{
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceRespondActivityTaskFailedResult converts thrift HistoryService_RespondActivityTaskFailed_Result type to internal
func ToHistoryServiceRespondActivityTaskFailedResult(t *history.HistoryService_RespondActivityTaskFailed_Result) *types.HistoryServiceRespondActivityTaskFailedResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRespondActivityTaskFailedResult{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceRespondDecisionTaskCompletedArgs converts internal HistoryService_RespondDecisionTaskCompleted_Args type to thrift
func FromHistoryServiceRespondDecisionTaskCompletedArgs(t *types.HistoryServiceRespondDecisionTaskCompletedArgs) *history.HistoryService_RespondDecisionTaskCompleted_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RespondDecisionTaskCompleted_Args{
		CompleteRequest: FromHistoryRespondDecisionTaskCompletedRequest(t.CompleteRequest),
	}
}

// ToHistoryServiceRespondDecisionTaskCompletedArgs converts thrift HistoryService_RespondDecisionTaskCompleted_Args type to internal
func ToHistoryServiceRespondDecisionTaskCompletedArgs(t *history.HistoryService_RespondDecisionTaskCompleted_Args) *types.HistoryServiceRespondDecisionTaskCompletedArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRespondDecisionTaskCompletedArgs{
		CompleteRequest: ToHistoryRespondDecisionTaskCompletedRequest(t.CompleteRequest),
	}
}

// FromHistoryServiceRespondDecisionTaskCompletedResult converts internal HistoryService_RespondDecisionTaskCompleted_Result type to thrift
func FromHistoryServiceRespondDecisionTaskCompletedResult(t *types.HistoryServiceRespondDecisionTaskCompletedResult) *history.HistoryService_RespondDecisionTaskCompleted_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RespondDecisionTaskCompleted_Result{
		Success:                 FromHistoryRespondDecisionTaskCompletedResponse(t.Success),
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceRespondDecisionTaskCompletedResult converts thrift HistoryService_RespondDecisionTaskCompleted_Result type to internal
func ToHistoryServiceRespondDecisionTaskCompletedResult(t *history.HistoryService_RespondDecisionTaskCompleted_Result) *types.HistoryServiceRespondDecisionTaskCompletedResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRespondDecisionTaskCompletedResult{
		Success:                 ToHistoryRespondDecisionTaskCompletedResponse(t.Success),
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceRespondDecisionTaskFailedArgs converts internal HistoryService_RespondDecisionTaskFailed_Args type to thrift
func FromHistoryServiceRespondDecisionTaskFailedArgs(t *types.HistoryServiceRespondDecisionTaskFailedArgs) *history.HistoryService_RespondDecisionTaskFailed_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RespondDecisionTaskFailed_Args{
		FailedRequest: FromHistoryRespondDecisionTaskFailedRequest(t.FailedRequest),
	}
}

// ToHistoryServiceRespondDecisionTaskFailedArgs converts thrift HistoryService_RespondDecisionTaskFailed_Args type to internal
func ToHistoryServiceRespondDecisionTaskFailedArgs(t *history.HistoryService_RespondDecisionTaskFailed_Args) *types.HistoryServiceRespondDecisionTaskFailedArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRespondDecisionTaskFailedArgs{
		FailedRequest: ToHistoryRespondDecisionTaskFailedRequest(t.FailedRequest),
	}
}

// FromHistoryServiceRespondDecisionTaskFailedResult converts internal HistoryService_RespondDecisionTaskFailed_Result type to thrift
func FromHistoryServiceRespondDecisionTaskFailedResult(t *types.HistoryServiceRespondDecisionTaskFailedResult) *history.HistoryService_RespondDecisionTaskFailed_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RespondDecisionTaskFailed_Result{
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceRespondDecisionTaskFailedResult converts thrift HistoryService_RespondDecisionTaskFailed_Result type to internal
func ToHistoryServiceRespondDecisionTaskFailedResult(t *history.HistoryService_RespondDecisionTaskFailed_Result) *types.HistoryServiceRespondDecisionTaskFailedResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceRespondDecisionTaskFailedResult{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceScheduleDecisionTaskArgs converts internal HistoryService_ScheduleDecisionTask_Args type to thrift
func FromHistoryServiceScheduleDecisionTaskArgs(t *types.HistoryServiceScheduleDecisionTaskArgs) *history.HistoryService_ScheduleDecisionTask_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ScheduleDecisionTask_Args{
		ScheduleRequest: FromScheduleDecisionTaskRequest(t.ScheduleRequest),
	}
}

// ToHistoryServiceScheduleDecisionTaskArgs converts thrift HistoryService_ScheduleDecisionTask_Args type to internal
func ToHistoryServiceScheduleDecisionTaskArgs(t *history.HistoryService_ScheduleDecisionTask_Args) *types.HistoryServiceScheduleDecisionTaskArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceScheduleDecisionTaskArgs{
		ScheduleRequest: ToScheduleDecisionTaskRequest(t.ScheduleRequest),
	}
}

// FromHistoryServiceScheduleDecisionTaskResult converts internal HistoryService_ScheduleDecisionTask_Result type to thrift
func FromHistoryServiceScheduleDecisionTaskResult(t *types.HistoryServiceScheduleDecisionTaskResult) *history.HistoryService_ScheduleDecisionTask_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ScheduleDecisionTask_Result{
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceScheduleDecisionTaskResult converts thrift HistoryService_ScheduleDecisionTask_Result type to internal
func ToHistoryServiceScheduleDecisionTaskResult(t *history.HistoryService_ScheduleDecisionTask_Result) *types.HistoryServiceScheduleDecisionTaskResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceScheduleDecisionTaskResult{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceSignalWithStartWorkflowExecutionArgs converts internal HistoryService_SignalWithStartWorkflowExecution_Args type to thrift
func FromHistoryServiceSignalWithStartWorkflowExecutionArgs(t *types.HistoryServiceSignalWithStartWorkflowExecutionArgs) *history.HistoryService_SignalWithStartWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_SignalWithStartWorkflowExecution_Args{
		SignalWithStartRequest: FromHistorySignalWithStartWorkflowExecutionRequest(t.SignalWithStartRequest),
	}
}

// ToHistoryServiceSignalWithStartWorkflowExecutionArgs converts thrift HistoryService_SignalWithStartWorkflowExecution_Args type to internal
func ToHistoryServiceSignalWithStartWorkflowExecutionArgs(t *history.HistoryService_SignalWithStartWorkflowExecution_Args) *types.HistoryServiceSignalWithStartWorkflowExecutionArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceSignalWithStartWorkflowExecutionArgs{
		SignalWithStartRequest: ToHistorySignalWithStartWorkflowExecutionRequest(t.SignalWithStartRequest),
	}
}

// FromHistoryServiceSignalWithStartWorkflowExecutionResult converts internal HistoryService_SignalWithStartWorkflowExecution_Result type to thrift
func FromHistoryServiceSignalWithStartWorkflowExecutionResult(t *types.HistoryServiceSignalWithStartWorkflowExecutionResult) *history.HistoryService_SignalWithStartWorkflowExecution_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_SignalWithStartWorkflowExecution_Result{
		Success:                     FromStartWorkflowExecutionResponse(t.Success),
		BadRequestError:             FromBadRequestError(t.BadRequestError),
		InternalServiceError:        FromInternalServiceError(t.InternalServiceError),
		ShardOwnershipLostError:     FromShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:        FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:          FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:            FromServiceBusyError(t.ServiceBusyError),
		WorkflowAlreadyStartedError: FromWorkflowExecutionAlreadyStartedError(t.WorkflowAlreadyStartedError),
	}
}

// ToHistoryServiceSignalWithStartWorkflowExecutionResult converts thrift HistoryService_SignalWithStartWorkflowExecution_Result type to internal
func ToHistoryServiceSignalWithStartWorkflowExecutionResult(t *history.HistoryService_SignalWithStartWorkflowExecution_Result) *types.HistoryServiceSignalWithStartWorkflowExecutionResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceSignalWithStartWorkflowExecutionResult{
		Success:                     ToStartWorkflowExecutionResponse(t.Success),
		BadRequestError:             ToBadRequestError(t.BadRequestError),
		InternalServiceError:        ToInternalServiceError(t.InternalServiceError),
		ShardOwnershipLostError:     ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:        ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:          ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:            ToServiceBusyError(t.ServiceBusyError),
		WorkflowAlreadyStartedError: ToWorkflowExecutionAlreadyStartedError(t.WorkflowAlreadyStartedError),
	}
}

// FromHistoryServiceSignalWorkflowExecutionArgs converts internal HistoryService_SignalWorkflowExecution_Args type to thrift
func FromHistoryServiceSignalWorkflowExecutionArgs(t *types.HistoryServiceSignalWorkflowExecutionArgs) *history.HistoryService_SignalWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_SignalWorkflowExecution_Args{
		SignalRequest: FromHistorySignalWorkflowExecutionRequest(t.SignalRequest),
	}
}

// ToHistoryServiceSignalWorkflowExecutionArgs converts thrift HistoryService_SignalWorkflowExecution_Args type to internal
func ToHistoryServiceSignalWorkflowExecutionArgs(t *history.HistoryService_SignalWorkflowExecution_Args) *types.HistoryServiceSignalWorkflowExecutionArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceSignalWorkflowExecutionArgs{
		SignalRequest: ToHistorySignalWorkflowExecutionRequest(t.SignalRequest),
	}
}

// FromHistoryServiceSignalWorkflowExecutionResult converts internal HistoryService_SignalWorkflowExecution_Result type to thrift
func FromHistoryServiceSignalWorkflowExecutionResult(t *types.HistoryServiceSignalWorkflowExecutionResult) *history.HistoryService_SignalWorkflowExecution_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_SignalWorkflowExecution_Result{
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    FromDomainNotActiveError(t.DomainNotActiveError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
	}
}

// ToHistoryServiceSignalWorkflowExecutionResult converts thrift HistoryService_SignalWorkflowExecution_Result type to internal
func ToHistoryServiceSignalWorkflowExecutionResult(t *history.HistoryService_SignalWorkflowExecution_Result) *types.HistoryServiceSignalWorkflowExecutionResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceSignalWorkflowExecutionResult{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
	}
}

// FromHistoryServiceStartWorkflowExecutionArgs converts internal HistoryService_StartWorkflowExecution_Args type to thrift
func FromHistoryServiceStartWorkflowExecutionArgs(t *types.HistoryServiceStartWorkflowExecutionArgs) *history.HistoryService_StartWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_StartWorkflowExecution_Args{
		StartRequest: FromHistoryStartWorkflowExecutionRequest(t.StartRequest),
	}
}

// ToHistoryServiceStartWorkflowExecutionArgs converts thrift HistoryService_StartWorkflowExecution_Args type to internal
func ToHistoryServiceStartWorkflowExecutionArgs(t *history.HistoryService_StartWorkflowExecution_Args) *types.HistoryServiceStartWorkflowExecutionArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceStartWorkflowExecutionArgs{
		StartRequest: ToHistoryStartWorkflowExecutionRequest(t.StartRequest),
	}
}

// FromHistoryServiceStartWorkflowExecutionResult converts internal HistoryService_StartWorkflowExecution_Result type to thrift
func FromHistoryServiceStartWorkflowExecutionResult(t *types.HistoryServiceStartWorkflowExecutionResult) *history.HistoryService_StartWorkflowExecution_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_StartWorkflowExecution_Result{
		Success:                  FromStartWorkflowExecutionResponse(t.Success),
		BadRequestError:          FromBadRequestError(t.BadRequestError),
		InternalServiceError:     FromInternalServiceError(t.InternalServiceError),
		SessionAlreadyExistError: FromWorkflowExecutionAlreadyStartedError(t.SessionAlreadyExistError),
		ShardOwnershipLostError:  FromShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:     FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:       FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:         FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceStartWorkflowExecutionResult converts thrift HistoryService_StartWorkflowExecution_Result type to internal
func ToHistoryServiceStartWorkflowExecutionResult(t *history.HistoryService_StartWorkflowExecution_Result) *types.HistoryServiceStartWorkflowExecutionResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceStartWorkflowExecutionResult{
		Success:                  ToStartWorkflowExecutionResponse(t.Success),
		BadRequestError:          ToBadRequestError(t.BadRequestError),
		InternalServiceError:     ToInternalServiceError(t.InternalServiceError),
		SessionAlreadyExistError: ToWorkflowExecutionAlreadyStartedError(t.SessionAlreadyExistError),
		ShardOwnershipLostError:  ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:     ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:       ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:         ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceSyncActivityArgs converts internal HistoryService_SyncActivity_Args type to thrift
func FromHistoryServiceSyncActivityArgs(t *types.HistoryServiceSyncActivityArgs) *history.HistoryService_SyncActivity_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_SyncActivity_Args{
		SyncActivityRequest: FromSyncActivityRequest(t.SyncActivityRequest),
	}
}

// ToHistoryServiceSyncActivityArgs converts thrift HistoryService_SyncActivity_Args type to internal
func ToHistoryServiceSyncActivityArgs(t *history.HistoryService_SyncActivity_Args) *types.HistoryServiceSyncActivityArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceSyncActivityArgs{
		SyncActivityRequest: ToSyncActivityRequest(t.SyncActivityRequest),
	}
}

// FromHistoryServiceSyncActivityResult converts internal HistoryService_SyncActivity_Result type to thrift
func FromHistoryServiceSyncActivityResult(t *types.HistoryServiceSyncActivityResult) *history.HistoryService_SyncActivity_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_SyncActivity_Result{
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
		RetryTaskV2Error:        FromRetryTaskV2Error(t.RetryTaskV2Error),
	}
}

// ToHistoryServiceSyncActivityResult converts thrift HistoryService_SyncActivity_Result type to internal
func ToHistoryServiceSyncActivityResult(t *history.HistoryService_SyncActivity_Result) *types.HistoryServiceSyncActivityResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceSyncActivityResult{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
		RetryTaskV2Error:        ToRetryTaskV2Error(t.RetryTaskV2Error),
	}
}

// FromHistoryServiceSyncShardStatusArgs converts internal HistoryService_SyncShardStatus_Args type to thrift
func FromHistoryServiceSyncShardStatusArgs(t *types.HistoryServiceSyncShardStatusArgs) *history.HistoryService_SyncShardStatus_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_SyncShardStatus_Args{
		SyncShardStatusRequest: FromSyncShardStatusRequest(t.SyncShardStatusRequest),
	}
}

// ToHistoryServiceSyncShardStatusArgs converts thrift HistoryService_SyncShardStatus_Args type to internal
func ToHistoryServiceSyncShardStatusArgs(t *history.HistoryService_SyncShardStatus_Args) *types.HistoryServiceSyncShardStatusArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceSyncShardStatusArgs{
		SyncShardStatusRequest: ToSyncShardStatusRequest(t.SyncShardStatusRequest),
	}
}

// FromHistoryServiceSyncShardStatusResult converts internal HistoryService_SyncShardStatus_Result type to thrift
func FromHistoryServiceSyncShardStatusResult(t *types.HistoryServiceSyncShardStatusResult) *history.HistoryService_SyncShardStatus_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_SyncShardStatus_Result{
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceSyncShardStatusResult converts thrift HistoryService_SyncShardStatus_Result type to internal
func ToHistoryServiceSyncShardStatusResult(t *history.HistoryService_SyncShardStatus_Result) *types.HistoryServiceSyncShardStatusResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceSyncShardStatusResult{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryServiceTerminateWorkflowExecutionArgs converts internal HistoryService_TerminateWorkflowExecution_Args type to thrift
func FromHistoryServiceTerminateWorkflowExecutionArgs(t *types.HistoryServiceTerminateWorkflowExecutionArgs) *history.HistoryService_TerminateWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_TerminateWorkflowExecution_Args{
		TerminateRequest: FromHistoryTerminateWorkflowExecutionRequest(t.TerminateRequest),
	}
}

// ToHistoryServiceTerminateWorkflowExecutionArgs converts thrift HistoryService_TerminateWorkflowExecution_Args type to internal
func ToHistoryServiceTerminateWorkflowExecutionArgs(t *history.HistoryService_TerminateWorkflowExecution_Args) *types.HistoryServiceTerminateWorkflowExecutionArgs {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceTerminateWorkflowExecutionArgs{
		TerminateRequest: ToHistoryTerminateWorkflowExecutionRequest(t.TerminateRequest),
	}
}

// FromHistoryServiceTerminateWorkflowExecutionResult converts internal HistoryService_TerminateWorkflowExecution_Result type to thrift
func FromHistoryServiceTerminateWorkflowExecutionResult(t *types.HistoryServiceTerminateWorkflowExecutionResult) *history.HistoryService_TerminateWorkflowExecution_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_TerminateWorkflowExecution_Result{
		BadRequestError:         FromBadRequestError(t.BadRequestError),
		InternalServiceError:    FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     FromEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: FromShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryServiceTerminateWorkflowExecutionResult converts thrift HistoryService_TerminateWorkflowExecution_Result type to internal
func ToHistoryServiceTerminateWorkflowExecutionResult(t *history.HistoryService_TerminateWorkflowExecution_Result) *types.HistoryServiceTerminateWorkflowExecutionResult {
	if t == nil {
		return nil
	}
	return &types.HistoryServiceTerminateWorkflowExecutionResult{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromNotifyFailoverMarkersRequest converts internal NotifyFailoverMarkersRequest type to thrift
func FromNotifyFailoverMarkersRequest(t *types.NotifyFailoverMarkersRequest) *history.NotifyFailoverMarkersRequest {
	if t == nil {
		return nil
	}
	return &history.NotifyFailoverMarkersRequest{
		FailoverMarkerTokens: FromFailoverMarkerTokenArray(t.FailoverMarkerTokens),
	}
}

// ToNotifyFailoverMarkersRequest converts thrift NotifyFailoverMarkersRequest type to internal
func ToNotifyFailoverMarkersRequest(t *history.NotifyFailoverMarkersRequest) *types.NotifyFailoverMarkersRequest {
	if t == nil {
		return nil
	}
	return &types.NotifyFailoverMarkersRequest{
		FailoverMarkerTokens: ToFailoverMarkerTokenArray(t.FailoverMarkerTokens),
	}
}

// FromParentExecutionInfo converts internal ParentExecutionInfo type to thrift
func FromParentExecutionInfo(t *types.ParentExecutionInfo) *history.ParentExecutionInfo {
	if t == nil {
		return nil
	}
	return &history.ParentExecutionInfo{
		DomainUUID:  t.DomainUUID,
		Domain:      t.Domain,
		Execution:   FromWorkflowExecution(t.Execution),
		InitiatedId: t.InitiatedID,
	}
}

// ToParentExecutionInfo converts thrift ParentExecutionInfo type to internal
func ToParentExecutionInfo(t *history.ParentExecutionInfo) *types.ParentExecutionInfo {
	if t == nil {
		return nil
	}
	return &types.ParentExecutionInfo{
		DomainUUID:  t.DomainUUID,
		Domain:      t.Domain,
		Execution:   ToWorkflowExecution(t.Execution),
		InitiatedID: t.InitiatedId,
	}
}

// FromPollMutableStateRequest converts internal PollMutableStateRequest type to thrift
func FromPollMutableStateRequest(t *types.PollMutableStateRequest) *history.PollMutableStateRequest {
	if t == nil {
		return nil
	}
	return &history.PollMutableStateRequest{
		DomainUUID:          t.DomainUUID,
		Execution:           FromWorkflowExecution(t.Execution),
		ExpectedNextEventId: t.ExpectedNextEventID,
		CurrentBranchToken:  t.CurrentBranchToken,
	}
}

// ToPollMutableStateRequest converts thrift PollMutableStateRequest type to internal
func ToPollMutableStateRequest(t *history.PollMutableStateRequest) *types.PollMutableStateRequest {
	if t == nil {
		return nil
	}
	return &types.PollMutableStateRequest{
		DomainUUID:          t.DomainUUID,
		Execution:           ToWorkflowExecution(t.Execution),
		ExpectedNextEventID: t.ExpectedNextEventId,
		CurrentBranchToken:  t.CurrentBranchToken,
	}
}

// FromPollMutableStateResponse converts internal PollMutableStateResponse type to thrift
func FromPollMutableStateResponse(t *types.PollMutableStateResponse) *history.PollMutableStateResponse {
	if t == nil {
		return nil
	}
	return &history.PollMutableStateResponse{
		Execution:                            FromWorkflowExecution(t.Execution),
		WorkflowType:                         FromWorkflowType(t.WorkflowType),
		NextEventId:                          t.NextEventID,
		PreviousStartedEventId:               t.PreviousStartedEventID,
		LastFirstEventId:                     t.LastFirstEventID,
		TaskList:                             FromTaskList(t.TaskList),
		StickyTaskList:                       FromTaskList(t.StickyTaskList),
		ClientLibraryVersion:                 t.ClientLibraryVersion,
		ClientFeatureVersion:                 t.ClientFeatureVersion,
		ClientImpl:                           t.ClientImpl,
		StickyTaskListScheduleToStartTimeout: t.StickyTaskListScheduleToStartTimeout,
		CurrentBranchToken:                   t.CurrentBranchToken,
		VersionHistories:                     FromVersionHistories(t.VersionHistories),
		WorkflowState:                        t.WorkflowState,
		WorkflowCloseState:                   t.WorkflowCloseState,
	}
}

// ToPollMutableStateResponse converts thrift PollMutableStateResponse type to internal
func ToPollMutableStateResponse(t *history.PollMutableStateResponse) *types.PollMutableStateResponse {
	if t == nil {
		return nil
	}
	return &types.PollMutableStateResponse{
		Execution:                            ToWorkflowExecution(t.Execution),
		WorkflowType:                         ToWorkflowType(t.WorkflowType),
		NextEventID:                          t.NextEventId,
		PreviousStartedEventID:               t.PreviousStartedEventId,
		LastFirstEventID:                     t.LastFirstEventId,
		TaskList:                             ToTaskList(t.TaskList),
		StickyTaskList:                       ToTaskList(t.StickyTaskList),
		ClientLibraryVersion:                 t.ClientLibraryVersion,
		ClientFeatureVersion:                 t.ClientFeatureVersion,
		ClientImpl:                           t.ClientImpl,
		StickyTaskListScheduleToStartTimeout: t.StickyTaskListScheduleToStartTimeout,
		CurrentBranchToken:                   t.CurrentBranchToken,
		VersionHistories:                     ToVersionHistories(t.VersionHistories),
		WorkflowState:                        t.WorkflowState,
		WorkflowCloseState:                   t.WorkflowCloseState,
	}
}

// FromProcessingQueueState converts internal ProcessingQueueState type to thrift
func FromProcessingQueueState(t *types.ProcessingQueueState) *history.ProcessingQueueState {
	if t == nil {
		return nil
	}
	return &history.ProcessingQueueState{
		Level:        t.Level,
		AckLevel:     t.AckLevel,
		MaxLevel:     t.MaxLevel,
		DomainFilter: FromDomainFilter(t.DomainFilter),
	}
}

// ToProcessingQueueState converts thrift ProcessingQueueState type to internal
func ToProcessingQueueState(t *history.ProcessingQueueState) *types.ProcessingQueueState {
	if t == nil {
		return nil
	}
	return &types.ProcessingQueueState{
		Level:        t.Level,
		AckLevel:     t.AckLevel,
		MaxLevel:     t.MaxLevel,
		DomainFilter: ToDomainFilter(t.DomainFilter),
	}
}

// FromProcessingQueueStates converts internal ProcessingQueueStates type to thrift
func FromProcessingQueueStates(t *types.ProcessingQueueStates) *history.ProcessingQueueStates {
	if t == nil {
		return nil
	}
	return &history.ProcessingQueueStates{
		StatesByCluster: FromProcessingQueueStateArrayMap(t.StatesByCluster),
	}
}

// ToProcessingQueueStates converts thrift ProcessingQueueStates type to internal
func ToProcessingQueueStates(t *history.ProcessingQueueStates) *types.ProcessingQueueStates {
	if t == nil {
		return nil
	}
	return &types.ProcessingQueueStates{
		StatesByCluster: ToProcessingQueueStateArrayMap(t.StatesByCluster),
	}
}

// FromHistoryQueryWorkflowRequest converts internal QueryWorkflowRequest type to thrift
func FromHistoryQueryWorkflowRequest(t *types.HistoryQueryWorkflowRequest) *history.QueryWorkflowRequest {
	if t == nil {
		return nil
	}
	return &history.QueryWorkflowRequest{
		DomainUUID: t.DomainUUID,
		Request:    FromQueryWorkflowRequest(t.Request),
	}
}

// ToHistoryQueryWorkflowRequest converts thrift QueryWorkflowRequest type to internal
func ToHistoryQueryWorkflowRequest(t *history.QueryWorkflowRequest) *types.HistoryQueryWorkflowRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryQueryWorkflowRequest{
		DomainUUID: t.DomainUUID,
		Request:    ToQueryWorkflowRequest(t.Request),
	}
}

// FromHistoryQueryWorkflowResponse converts internal QueryWorkflowResponse type to thrift
func FromHistoryQueryWorkflowResponse(t *types.HistoryQueryWorkflowResponse) *history.QueryWorkflowResponse {
	if t == nil {
		return nil
	}
	return &history.QueryWorkflowResponse{
		Response: FromQueryWorkflowResponse(t.Response),
	}
}

// ToHistoryQueryWorkflowResponse converts thrift QueryWorkflowResponse type to internal
func ToHistoryQueryWorkflowResponse(t *history.QueryWorkflowResponse) *types.HistoryQueryWorkflowResponse {
	if t == nil {
		return nil
	}
	return &types.HistoryQueryWorkflowResponse{
		Response: ToQueryWorkflowResponse(t.Response),
	}
}

// FromHistoryReapplyEventsRequest converts internal ReapplyEventsRequest type to thrift
func FromHistoryReapplyEventsRequest(t *types.HistoryReapplyEventsRequest) *history.ReapplyEventsRequest {
	if t == nil {
		return nil
	}
	return &history.ReapplyEventsRequest{
		DomainUUID: t.DomainUUID,
		Request:    FromReapplyEventsRequest(t.Request),
	}
}

// ToHistoryReapplyEventsRequest converts thrift ReapplyEventsRequest type to internal
func ToHistoryReapplyEventsRequest(t *history.ReapplyEventsRequest) *types.HistoryReapplyEventsRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryReapplyEventsRequest{
		DomainUUID: t.DomainUUID,
		Request:    ToReapplyEventsRequest(t.Request),
	}
}

// FromHistoryRecordActivityTaskHeartbeatRequest converts internal RecordActivityTaskHeartbeatRequest type to thrift
func FromHistoryRecordActivityTaskHeartbeatRequest(t *types.HistoryRecordActivityTaskHeartbeatRequest) *history.RecordActivityTaskHeartbeatRequest {
	if t == nil {
		return nil
	}
	return &history.RecordActivityTaskHeartbeatRequest{
		DomainUUID:       t.DomainUUID,
		HeartbeatRequest: FromRecordActivityTaskHeartbeatRequest(t.HeartbeatRequest),
	}
}

// ToHistoryRecordActivityTaskHeartbeatRequest converts thrift RecordActivityTaskHeartbeatRequest type to internal
func ToHistoryRecordActivityTaskHeartbeatRequest(t *history.RecordActivityTaskHeartbeatRequest) *types.HistoryRecordActivityTaskHeartbeatRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryRecordActivityTaskHeartbeatRequest{
		DomainUUID:       t.DomainUUID,
		HeartbeatRequest: ToRecordActivityTaskHeartbeatRequest(t.HeartbeatRequest),
	}
}

// FromRecordActivityTaskStartedRequest converts internal RecordActivityTaskStartedRequest type to thrift
func FromRecordActivityTaskStartedRequest(t *types.RecordActivityTaskStartedRequest) *history.RecordActivityTaskStartedRequest {
	if t == nil {
		return nil
	}
	return &history.RecordActivityTaskStartedRequest{
		DomainUUID:        t.DomainUUID,
		WorkflowExecution: FromWorkflowExecution(t.WorkflowExecution),
		ScheduleId:        t.ScheduleID,
		TaskId:            t.TaskID,
		RequestId:         t.RequestID,
		PollRequest:       FromPollForActivityTaskRequest(t.PollRequest),
	}
}

// ToRecordActivityTaskStartedRequest converts thrift RecordActivityTaskStartedRequest type to internal
func ToRecordActivityTaskStartedRequest(t *history.RecordActivityTaskStartedRequest) *types.RecordActivityTaskStartedRequest {
	if t == nil {
		return nil
	}
	return &types.RecordActivityTaskStartedRequest{
		DomainUUID:        t.DomainUUID,
		WorkflowExecution: ToWorkflowExecution(t.WorkflowExecution),
		ScheduleID:        t.ScheduleId,
		TaskID:            t.TaskId,
		RequestID:         t.RequestId,
		PollRequest:       ToPollForActivityTaskRequest(t.PollRequest),
	}
}

// FromRecordActivityTaskStartedResponse converts internal RecordActivityTaskStartedResponse type to thrift
func FromRecordActivityTaskStartedResponse(t *types.RecordActivityTaskStartedResponse) *history.RecordActivityTaskStartedResponse {
	if t == nil {
		return nil
	}
	return &history.RecordActivityTaskStartedResponse{
		ScheduledEvent:                  FromHistoryEvent(t.ScheduledEvent),
		StartedTimestamp:                t.StartedTimestamp,
		Attempt:                         t.Attempt,
		ScheduledTimestampOfThisAttempt: t.ScheduledTimestampOfThisAttempt,
		HeartbeatDetails:                t.HeartbeatDetails,
		WorkflowType:                    FromWorkflowType(t.WorkflowType),
		WorkflowDomain:                  t.WorkflowDomain,
	}
}

// ToRecordActivityTaskStartedResponse converts thrift RecordActivityTaskStartedResponse type to internal
func ToRecordActivityTaskStartedResponse(t *history.RecordActivityTaskStartedResponse) *types.RecordActivityTaskStartedResponse {
	if t == nil {
		return nil
	}
	return &types.RecordActivityTaskStartedResponse{
		ScheduledEvent:                  ToHistoryEvent(t.ScheduledEvent),
		StartedTimestamp:                t.StartedTimestamp,
		Attempt:                         t.Attempt,
		ScheduledTimestampOfThisAttempt: t.ScheduledTimestampOfThisAttempt,
		HeartbeatDetails:                t.HeartbeatDetails,
		WorkflowType:                    ToWorkflowType(t.WorkflowType),
		WorkflowDomain:                  t.WorkflowDomain,
	}
}

// FromRecordChildExecutionCompletedRequest converts internal RecordChildExecutionCompletedRequest type to thrift
func FromRecordChildExecutionCompletedRequest(t *types.RecordChildExecutionCompletedRequest) *history.RecordChildExecutionCompletedRequest {
	if t == nil {
		return nil
	}
	return &history.RecordChildExecutionCompletedRequest{
		DomainUUID:         t.DomainUUID,
		WorkflowExecution:  FromWorkflowExecution(t.WorkflowExecution),
		InitiatedId:        t.InitiatedID,
		CompletedExecution: FromWorkflowExecution(t.CompletedExecution),
		CompletionEvent:    FromHistoryEvent(t.CompletionEvent),
	}
}

// ToRecordChildExecutionCompletedRequest converts thrift RecordChildExecutionCompletedRequest type to internal
func ToRecordChildExecutionCompletedRequest(t *history.RecordChildExecutionCompletedRequest) *types.RecordChildExecutionCompletedRequest {
	if t == nil {
		return nil
	}
	return &types.RecordChildExecutionCompletedRequest{
		DomainUUID:         t.DomainUUID,
		WorkflowExecution:  ToWorkflowExecution(t.WorkflowExecution),
		InitiatedID:        t.InitiatedId,
		CompletedExecution: ToWorkflowExecution(t.CompletedExecution),
		CompletionEvent:    ToHistoryEvent(t.CompletionEvent),
	}
}

// FromRecordDecisionTaskStartedRequest converts internal RecordDecisionTaskStartedRequest type to thrift
func FromRecordDecisionTaskStartedRequest(t *types.RecordDecisionTaskStartedRequest) *history.RecordDecisionTaskStartedRequest {
	if t == nil {
		return nil
	}
	return &history.RecordDecisionTaskStartedRequest{
		DomainUUID:        t.DomainUUID,
		WorkflowExecution: FromWorkflowExecution(t.WorkflowExecution),
		ScheduleId:        t.ScheduleID,
		TaskId:            t.TaskID,
		RequestId:         t.RequestID,
		PollRequest:       FromPollForDecisionTaskRequest(t.PollRequest),
	}
}

// ToRecordDecisionTaskStartedRequest converts thrift RecordDecisionTaskStartedRequest type to internal
func ToRecordDecisionTaskStartedRequest(t *history.RecordDecisionTaskStartedRequest) *types.RecordDecisionTaskStartedRequest {
	if t == nil {
		return nil
	}
	return &types.RecordDecisionTaskStartedRequest{
		DomainUUID:        t.DomainUUID,
		WorkflowExecution: ToWorkflowExecution(t.WorkflowExecution),
		ScheduleID:        t.ScheduleId,
		TaskID:            t.TaskId,
		RequestID:         t.RequestId,
		PollRequest:       ToPollForDecisionTaskRequest(t.PollRequest),
	}
}

// FromRecordDecisionTaskStartedResponse converts internal RecordDecisionTaskStartedResponse type to thrift
func FromRecordDecisionTaskStartedResponse(t *types.RecordDecisionTaskStartedResponse) *history.RecordDecisionTaskStartedResponse {
	if t == nil {
		return nil
	}
	return &history.RecordDecisionTaskStartedResponse{
		WorkflowType:              FromWorkflowType(t.WorkflowType),
		PreviousStartedEventId:    t.PreviousStartedEventID,
		ScheduledEventId:          t.ScheduledEventID,
		StartedEventId:            t.StartedEventID,
		NextEventId:               t.NextEventID,
		Attempt:                   t.Attempt,
		StickyExecutionEnabled:    t.StickyExecutionEnabled,
		DecisionInfo:              FromTransientDecisionInfo(t.DecisionInfo),
		WorkflowExecutionTaskList: FromTaskList(t.WorkflowExecutionTaskList),
		EventStoreVersion:         t.EventStoreVersion,
		BranchToken:               t.BranchToken,
		ScheduledTimestamp:        t.ScheduledTimestamp,
		StartedTimestamp:          t.StartedTimestamp,
		Queries:                   FromWorkflowQueryMap(t.Queries),
	}
}

// ToRecordDecisionTaskStartedResponse converts thrift RecordDecisionTaskStartedResponse type to internal
func ToRecordDecisionTaskStartedResponse(t *history.RecordDecisionTaskStartedResponse) *types.RecordDecisionTaskStartedResponse {
	if t == nil {
		return nil
	}
	return &types.RecordDecisionTaskStartedResponse{
		WorkflowType:              ToWorkflowType(t.WorkflowType),
		PreviousStartedEventID:    t.PreviousStartedEventId,
		ScheduledEventID:          t.ScheduledEventId,
		StartedEventID:            t.StartedEventId,
		NextEventID:               t.NextEventId,
		Attempt:                   t.Attempt,
		StickyExecutionEnabled:    t.StickyExecutionEnabled,
		DecisionInfo:              ToTransientDecisionInfo(t.DecisionInfo),
		WorkflowExecutionTaskList: ToTaskList(t.WorkflowExecutionTaskList),
		EventStoreVersion:         t.EventStoreVersion,
		BranchToken:               t.BranchToken,
		ScheduledTimestamp:        t.ScheduledTimestamp,
		StartedTimestamp:          t.StartedTimestamp,
		Queries:                   ToWorkflowQueryMap(t.Queries),
	}
}

// FromHistoryRefreshWorkflowTasksRequest converts internal RefreshWorkflowTasksRequest type to thrift
func FromHistoryRefreshWorkflowTasksRequest(t *types.HistoryRefreshWorkflowTasksRequest) *history.RefreshWorkflowTasksRequest {
	if t == nil {
		return nil
	}
	return &history.RefreshWorkflowTasksRequest{
		DomainUIID: t.DomainUIID,
		Request:    FromRefreshWorkflowTasksRequest(t.Request),
	}
}

// ToHistoryRefreshWorkflowTasksRequest converts thrift RefreshWorkflowTasksRequest type to internal
func ToHistoryRefreshWorkflowTasksRequest(t *history.RefreshWorkflowTasksRequest) *types.HistoryRefreshWorkflowTasksRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryRefreshWorkflowTasksRequest{
		DomainUIID: t.DomainUIID,
		Request:    ToRefreshWorkflowTasksRequest(t.Request),
	}
}

// FromRemoveSignalMutableStateRequest converts internal RemoveSignalMutableStateRequest type to thrift
func FromRemoveSignalMutableStateRequest(t *types.RemoveSignalMutableStateRequest) *history.RemoveSignalMutableStateRequest {
	if t == nil {
		return nil
	}
	return &history.RemoveSignalMutableStateRequest{
		DomainUUID:        t.DomainUUID,
		WorkflowExecution: FromWorkflowExecution(t.WorkflowExecution),
		RequestId:         t.RequestID,
	}
}

// ToRemoveSignalMutableStateRequest converts thrift RemoveSignalMutableStateRequest type to internal
func ToRemoveSignalMutableStateRequest(t *history.RemoveSignalMutableStateRequest) *types.RemoveSignalMutableStateRequest {
	if t == nil {
		return nil
	}
	return &types.RemoveSignalMutableStateRequest{
		DomainUUID:        t.DomainUUID,
		WorkflowExecution: ToWorkflowExecution(t.WorkflowExecution),
		RequestID:         t.RequestId,
	}
}

// FromReplicateEventsV2Request converts internal ReplicateEventsV2Request type to thrift
func FromReplicateEventsV2Request(t *types.ReplicateEventsV2Request) *history.ReplicateEventsV2Request {
	if t == nil {
		return nil
	}
	return &history.ReplicateEventsV2Request{
		DomainUUID:          t.DomainUUID,
		WorkflowExecution:   FromWorkflowExecution(t.WorkflowExecution),
		VersionHistoryItems: FromVersionHistoryItemArray(t.VersionHistoryItems),
		Events:              FromDataBlob(t.Events),
		NewRunEvents:        FromDataBlob(t.NewRunEvents),
	}
}

// ToReplicateEventsV2Request converts thrift ReplicateEventsV2Request type to internal
func ToReplicateEventsV2Request(t *history.ReplicateEventsV2Request) *types.ReplicateEventsV2Request {
	if t == nil {
		return nil
	}
	return &types.ReplicateEventsV2Request{
		DomainUUID:          t.DomainUUID,
		WorkflowExecution:   ToWorkflowExecution(t.WorkflowExecution),
		VersionHistoryItems: ToVersionHistoryItemArray(t.VersionHistoryItems),
		Events:              ToDataBlob(t.Events),
		NewRunEvents:        ToDataBlob(t.NewRunEvents),
	}
}

// FromHistoryRequestCancelWorkflowExecutionRequest converts internal RequestCancelWorkflowExecutionRequest type to thrift
func FromHistoryRequestCancelWorkflowExecutionRequest(t *types.HistoryRequestCancelWorkflowExecutionRequest) *history.RequestCancelWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &history.RequestCancelWorkflowExecutionRequest{
		DomainUUID:                t.DomainUUID,
		CancelRequest:             FromRequestCancelWorkflowExecutionRequest(t.CancelRequest),
		ExternalInitiatedEventId:  t.ExternalInitiatedEventID,
		ExternalWorkflowExecution: FromWorkflowExecution(t.ExternalWorkflowExecution),
		ChildWorkflowOnly:         t.ChildWorkflowOnly,
	}
}

// ToHistoryRequestCancelWorkflowExecutionRequest converts thrift RequestCancelWorkflowExecutionRequest type to internal
func ToHistoryRequestCancelWorkflowExecutionRequest(t *history.RequestCancelWorkflowExecutionRequest) *types.HistoryRequestCancelWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryRequestCancelWorkflowExecutionRequest{
		DomainUUID:                t.DomainUUID,
		CancelRequest:             ToRequestCancelWorkflowExecutionRequest(t.CancelRequest),
		ExternalInitiatedEventID:  t.ExternalInitiatedEventId,
		ExternalWorkflowExecution: ToWorkflowExecution(t.ExternalWorkflowExecution),
		ChildWorkflowOnly:         t.ChildWorkflowOnly,
	}
}

// FromHistoryResetStickyTaskListRequest converts internal ResetStickyTaskListRequest type to thrift
func FromHistoryResetStickyTaskListRequest(t *types.HistoryResetStickyTaskListRequest) *history.ResetStickyTaskListRequest {
	if t == nil {
		return nil
	}
	return &history.ResetStickyTaskListRequest{
		DomainUUID: t.DomainUUID,
		Execution:  FromWorkflowExecution(t.Execution),
	}
}

// ToHistoryResetStickyTaskListRequest converts thrift ResetStickyTaskListRequest type to internal
func ToHistoryResetStickyTaskListRequest(t *history.ResetStickyTaskListRequest) *types.HistoryResetStickyTaskListRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryResetStickyTaskListRequest{
		DomainUUID: t.DomainUUID,
		Execution:  ToWorkflowExecution(t.Execution),
	}
}

// FromHistoryResetStickyTaskListResponse converts internal ResetStickyTaskListResponse type to thrift
func FromHistoryResetStickyTaskListResponse(t *types.HistoryResetStickyTaskListResponse) *history.ResetStickyTaskListResponse {
	if t == nil {
		return nil
	}
	return &history.ResetStickyTaskListResponse{}
}

// ToHistoryResetStickyTaskListResponse converts thrift ResetStickyTaskListResponse type to internal
func ToHistoryResetStickyTaskListResponse(t *history.ResetStickyTaskListResponse) *types.HistoryResetStickyTaskListResponse {
	if t == nil {
		return nil
	}
	return &types.HistoryResetStickyTaskListResponse{}
}

// FromHistoryResetWorkflowExecutionRequest converts internal ResetWorkflowExecutionRequest type to thrift
func FromHistoryResetWorkflowExecutionRequest(t *types.HistoryResetWorkflowExecutionRequest) *history.ResetWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &history.ResetWorkflowExecutionRequest{
		DomainUUID:   t.DomainUUID,
		ResetRequest: FromResetWorkflowExecutionRequest(t.ResetRequest),
	}
}

// ToHistoryResetWorkflowExecutionRequest converts thrift ResetWorkflowExecutionRequest type to internal
func ToHistoryResetWorkflowExecutionRequest(t *history.ResetWorkflowExecutionRequest) *types.HistoryResetWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryResetWorkflowExecutionRequest{
		DomainUUID:   t.DomainUUID,
		ResetRequest: ToResetWorkflowExecutionRequest(t.ResetRequest),
	}
}

// FromHistoryRespondActivityTaskCanceledRequest converts internal RespondActivityTaskCanceledRequest type to thrift
func FromHistoryRespondActivityTaskCanceledRequest(t *types.HistoryRespondActivityTaskCanceledRequest) *history.RespondActivityTaskCanceledRequest {
	if t == nil {
		return nil
	}
	return &history.RespondActivityTaskCanceledRequest{
		DomainUUID:    t.DomainUUID,
		CancelRequest: FromRespondActivityTaskCanceledRequest(t.CancelRequest),
	}
}

// ToHistoryRespondActivityTaskCanceledRequest converts thrift RespondActivityTaskCanceledRequest type to internal
func ToHistoryRespondActivityTaskCanceledRequest(t *history.RespondActivityTaskCanceledRequest) *types.HistoryRespondActivityTaskCanceledRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryRespondActivityTaskCanceledRequest{
		DomainUUID:    t.DomainUUID,
		CancelRequest: ToRespondActivityTaskCanceledRequest(t.CancelRequest),
	}
}

// FromHistoryRespondActivityTaskCompletedRequest converts internal RespondActivityTaskCompletedRequest type to thrift
func FromHistoryRespondActivityTaskCompletedRequest(t *types.HistoryRespondActivityTaskCompletedRequest) *history.RespondActivityTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &history.RespondActivityTaskCompletedRequest{
		DomainUUID:      t.DomainUUID,
		CompleteRequest: FromRespondActivityTaskCompletedRequest(t.CompleteRequest),
	}
}

// ToHistoryRespondActivityTaskCompletedRequest converts thrift RespondActivityTaskCompletedRequest type to internal
func ToHistoryRespondActivityTaskCompletedRequest(t *history.RespondActivityTaskCompletedRequest) *types.HistoryRespondActivityTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryRespondActivityTaskCompletedRequest{
		DomainUUID:      t.DomainUUID,
		CompleteRequest: ToRespondActivityTaskCompletedRequest(t.CompleteRequest),
	}
}

// FromHistoryRespondActivityTaskFailedRequest converts internal RespondActivityTaskFailedRequest type to thrift
func FromHistoryRespondActivityTaskFailedRequest(t *types.HistoryRespondActivityTaskFailedRequest) *history.RespondActivityTaskFailedRequest {
	if t == nil {
		return nil
	}
	return &history.RespondActivityTaskFailedRequest{
		DomainUUID:    t.DomainUUID,
		FailedRequest: FromRespondActivityTaskFailedRequest(t.FailedRequest),
	}
}

// ToHistoryRespondActivityTaskFailedRequest converts thrift RespondActivityTaskFailedRequest type to internal
func ToHistoryRespondActivityTaskFailedRequest(t *history.RespondActivityTaskFailedRequest) *types.HistoryRespondActivityTaskFailedRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryRespondActivityTaskFailedRequest{
		DomainUUID:    t.DomainUUID,
		FailedRequest: ToRespondActivityTaskFailedRequest(t.FailedRequest),
	}
}

// FromHistoryRespondDecisionTaskCompletedRequest converts internal RespondDecisionTaskCompletedRequest type to thrift
func FromHistoryRespondDecisionTaskCompletedRequest(t *types.HistoryRespondDecisionTaskCompletedRequest) *history.RespondDecisionTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &history.RespondDecisionTaskCompletedRequest{
		DomainUUID:      t.DomainUUID,
		CompleteRequest: FromRespondDecisionTaskCompletedRequest(t.CompleteRequest),
	}
}

// ToHistoryRespondDecisionTaskCompletedRequest converts thrift RespondDecisionTaskCompletedRequest type to internal
func ToHistoryRespondDecisionTaskCompletedRequest(t *history.RespondDecisionTaskCompletedRequest) *types.HistoryRespondDecisionTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryRespondDecisionTaskCompletedRequest{
		DomainUUID:      t.DomainUUID,
		CompleteRequest: ToRespondDecisionTaskCompletedRequest(t.CompleteRequest),
	}
}

// FromHistoryRespondDecisionTaskCompletedResponse converts internal RespondDecisionTaskCompletedResponse type to thrift
func FromHistoryRespondDecisionTaskCompletedResponse(t *types.HistoryRespondDecisionTaskCompletedResponse) *history.RespondDecisionTaskCompletedResponse {
	if t == nil {
		return nil
	}
	return &history.RespondDecisionTaskCompletedResponse{
		StartedResponse:             FromRecordDecisionTaskStartedResponse(t.StartedResponse),
		ActivitiesToDispatchLocally: FromActivityLocalDispatchInfoMap(t.ActivitiesToDispatchLocally),
	}
}

// ToHistoryRespondDecisionTaskCompletedResponse converts thrift RespondDecisionTaskCompletedResponse type to internal
func ToHistoryRespondDecisionTaskCompletedResponse(t *history.RespondDecisionTaskCompletedResponse) *types.HistoryRespondDecisionTaskCompletedResponse {
	if t == nil {
		return nil
	}
	return &types.HistoryRespondDecisionTaskCompletedResponse{
		StartedResponse:             ToRecordDecisionTaskStartedResponse(t.StartedResponse),
		ActivitiesToDispatchLocally: ToActivityLocalDispatchInfoMap(t.ActivitiesToDispatchLocally),
	}
}

// FromHistoryRespondDecisionTaskFailedRequest converts internal RespondDecisionTaskFailedRequest type to thrift
func FromHistoryRespondDecisionTaskFailedRequest(t *types.HistoryRespondDecisionTaskFailedRequest) *history.RespondDecisionTaskFailedRequest {
	if t == nil {
		return nil
	}
	return &history.RespondDecisionTaskFailedRequest{
		DomainUUID:    t.DomainUUID,
		FailedRequest: FromRespondDecisionTaskFailedRequest(t.FailedRequest),
	}
}

// ToHistoryRespondDecisionTaskFailedRequest converts thrift RespondDecisionTaskFailedRequest type to internal
func ToHistoryRespondDecisionTaskFailedRequest(t *history.RespondDecisionTaskFailedRequest) *types.HistoryRespondDecisionTaskFailedRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryRespondDecisionTaskFailedRequest{
		DomainUUID:    t.DomainUUID,
		FailedRequest: ToRespondDecisionTaskFailedRequest(t.FailedRequest),
	}
}

// FromScheduleDecisionTaskRequest converts internal ScheduleDecisionTaskRequest type to thrift
func FromScheduleDecisionTaskRequest(t *types.ScheduleDecisionTaskRequest) *history.ScheduleDecisionTaskRequest {
	if t == nil {
		return nil
	}
	return &history.ScheduleDecisionTaskRequest{
		DomainUUID:        t.DomainUUID,
		WorkflowExecution: FromWorkflowExecution(t.WorkflowExecution),
		IsFirstDecision:   t.IsFirstDecision,
	}
}

// ToScheduleDecisionTaskRequest converts thrift ScheduleDecisionTaskRequest type to internal
func ToScheduleDecisionTaskRequest(t *history.ScheduleDecisionTaskRequest) *types.ScheduleDecisionTaskRequest {
	if t == nil {
		return nil
	}
	return &types.ScheduleDecisionTaskRequest{
		DomainUUID:        t.DomainUUID,
		WorkflowExecution: ToWorkflowExecution(t.WorkflowExecution),
		IsFirstDecision:   t.IsFirstDecision,
	}
}

// FromShardOwnershipLostError converts internal ShardOwnershipLostError type to thrift
func FromShardOwnershipLostError(t *types.ShardOwnershipLostError) *history.ShardOwnershipLostError {
	if t == nil {
		return nil
	}
	return &history.ShardOwnershipLostError{
		Message: t.Message,
		Owner:   t.Owner,
	}
}

// ToShardOwnershipLostError converts thrift ShardOwnershipLostError type to internal
func ToShardOwnershipLostError(t *history.ShardOwnershipLostError) *types.ShardOwnershipLostError {
	if t == nil {
		return nil
	}
	return &types.ShardOwnershipLostError{
		Message: t.Message,
		Owner:   t.Owner,
	}
}

// FromHistorySignalWithStartWorkflowExecutionRequest converts internal SignalWithStartWorkflowExecutionRequest type to thrift
func FromHistorySignalWithStartWorkflowExecutionRequest(t *types.HistorySignalWithStartWorkflowExecutionRequest) *history.SignalWithStartWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &history.SignalWithStartWorkflowExecutionRequest{
		DomainUUID:             t.DomainUUID,
		SignalWithStartRequest: FromSignalWithStartWorkflowExecutionRequest(t.SignalWithStartRequest),
	}
}

// ToHistorySignalWithStartWorkflowExecutionRequest converts thrift SignalWithStartWorkflowExecutionRequest type to internal
func ToHistorySignalWithStartWorkflowExecutionRequest(t *history.SignalWithStartWorkflowExecutionRequest) *types.HistorySignalWithStartWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &types.HistorySignalWithStartWorkflowExecutionRequest{
		DomainUUID:             t.DomainUUID,
		SignalWithStartRequest: ToSignalWithStartWorkflowExecutionRequest(t.SignalWithStartRequest),
	}
}

// FromHistorySignalWorkflowExecutionRequest converts internal SignalWorkflowExecutionRequest type to thrift
func FromHistorySignalWorkflowExecutionRequest(t *types.HistorySignalWorkflowExecutionRequest) *history.SignalWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &history.SignalWorkflowExecutionRequest{
		DomainUUID:                t.DomainUUID,
		SignalRequest:             FromSignalWorkflowExecutionRequest(t.SignalRequest),
		ExternalWorkflowExecution: FromWorkflowExecution(t.ExternalWorkflowExecution),
		ChildWorkflowOnly:         t.ChildWorkflowOnly,
	}
}

// ToHistorySignalWorkflowExecutionRequest converts thrift SignalWorkflowExecutionRequest type to internal
func ToHistorySignalWorkflowExecutionRequest(t *history.SignalWorkflowExecutionRequest) *types.HistorySignalWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &types.HistorySignalWorkflowExecutionRequest{
		DomainUUID:                t.DomainUUID,
		SignalRequest:             ToSignalWorkflowExecutionRequest(t.SignalRequest),
		ExternalWorkflowExecution: ToWorkflowExecution(t.ExternalWorkflowExecution),
		ChildWorkflowOnly:         t.ChildWorkflowOnly,
	}
}

// FromHistoryStartWorkflowExecutionRequest converts internal StartWorkflowExecutionRequest type to thrift
func FromHistoryStartWorkflowExecutionRequest(t *types.HistoryStartWorkflowExecutionRequest) *history.StartWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &history.StartWorkflowExecutionRequest{
		DomainUUID:                      t.DomainUUID,
		StartRequest:                    FromStartWorkflowExecutionRequest(t.StartRequest),
		ParentExecutionInfo:             FromParentExecutionInfo(t.ParentExecutionInfo),
		Attempt:                         t.Attempt,
		ExpirationTimestamp:             t.ExpirationTimestamp,
		ContinueAsNewInitiator:          FromContinueAsNewInitiator(t.ContinueAsNewInitiator),
		ContinuedFailureReason:          t.ContinuedFailureReason,
		ContinuedFailureDetails:         t.ContinuedFailureDetails,
		LastCompletionResult:            t.LastCompletionResult,
		FirstDecisionTaskBackoffSeconds: t.FirstDecisionTaskBackoffSeconds,
	}
}

// ToHistoryStartWorkflowExecutionRequest converts thrift StartWorkflowExecutionRequest type to internal
func ToHistoryStartWorkflowExecutionRequest(t *history.StartWorkflowExecutionRequest) *types.HistoryStartWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryStartWorkflowExecutionRequest{
		DomainUUID:                      t.DomainUUID,
		StartRequest:                    ToStartWorkflowExecutionRequest(t.StartRequest),
		ParentExecutionInfo:             ToParentExecutionInfo(t.ParentExecutionInfo),
		Attempt:                         t.Attempt,
		ExpirationTimestamp:             t.ExpirationTimestamp,
		ContinueAsNewInitiator:          ToContinueAsNewInitiator(t.ContinueAsNewInitiator),
		ContinuedFailureReason:          t.ContinuedFailureReason,
		ContinuedFailureDetails:         t.ContinuedFailureDetails,
		LastCompletionResult:            t.LastCompletionResult,
		FirstDecisionTaskBackoffSeconds: t.FirstDecisionTaskBackoffSeconds,
	}
}

// FromSyncActivityRequest converts internal SyncActivityRequest type to thrift
func FromSyncActivityRequest(t *types.SyncActivityRequest) *history.SyncActivityRequest {
	if t == nil {
		return nil
	}
	return &history.SyncActivityRequest{
		DomainId:           t.DomainID,
		WorkflowId:         t.WorkflowID,
		RunId:              t.RunID,
		Version:            t.Version,
		ScheduledId:        t.ScheduledID,
		ScheduledTime:      t.ScheduledTime,
		StartedId:          t.StartedID,
		StartedTime:        t.StartedTime,
		LastHeartbeatTime:  t.LastHeartbeatTime,
		Details:            t.Details,
		Attempt:            t.Attempt,
		LastFailureReason:  t.LastFailureReason,
		LastWorkerIdentity: t.LastWorkerIdentity,
		LastFailureDetails: t.LastFailureDetails,
		VersionHistory:     FromVersionHistory(t.VersionHistory),
	}
}

// ToSyncActivityRequest converts thrift SyncActivityRequest type to internal
func ToSyncActivityRequest(t *history.SyncActivityRequest) *types.SyncActivityRequest {
	if t == nil {
		return nil
	}
	return &types.SyncActivityRequest{
		DomainID:           t.DomainId,
		WorkflowID:         t.WorkflowId,
		RunID:              t.RunId,
		Version:            t.Version,
		ScheduledID:        t.ScheduledId,
		ScheduledTime:      t.ScheduledTime,
		StartedID:          t.StartedId,
		StartedTime:        t.StartedTime,
		LastHeartbeatTime:  t.LastHeartbeatTime,
		Details:            t.Details,
		Attempt:            t.Attempt,
		LastFailureReason:  t.LastFailureReason,
		LastWorkerIdentity: t.LastWorkerIdentity,
		LastFailureDetails: t.LastFailureDetails,
		VersionHistory:     ToVersionHistory(t.VersionHistory),
	}
}

// FromSyncShardStatusRequest converts internal SyncShardStatusRequest type to thrift
func FromSyncShardStatusRequest(t *types.SyncShardStatusRequest) *history.SyncShardStatusRequest {
	if t == nil {
		return nil
	}
	return &history.SyncShardStatusRequest{
		SourceCluster: t.SourceCluster,
		ShardId:       t.ShardID,
		Timestamp:     t.Timestamp,
	}
}

// ToSyncShardStatusRequest converts thrift SyncShardStatusRequest type to internal
func ToSyncShardStatusRequest(t *history.SyncShardStatusRequest) *types.SyncShardStatusRequest {
	if t == nil {
		return nil
	}
	return &types.SyncShardStatusRequest{
		SourceCluster: t.SourceCluster,
		ShardID:       t.ShardId,
		Timestamp:     t.Timestamp,
	}
}

// FromHistoryTerminateWorkflowExecutionRequest converts internal TerminateWorkflowExecutionRequest type to thrift
func FromHistoryTerminateWorkflowExecutionRequest(t *types.HistoryTerminateWorkflowExecutionRequest) *history.TerminateWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &history.TerminateWorkflowExecutionRequest{
		DomainUUID:       t.DomainUUID,
		TerminateRequest: FromTerminateWorkflowExecutionRequest(t.TerminateRequest),
	}
}

// ToHistoryTerminateWorkflowExecutionRequest converts thrift TerminateWorkflowExecutionRequest type to internal
func ToHistoryTerminateWorkflowExecutionRequest(t *history.TerminateWorkflowExecutionRequest) *types.HistoryTerminateWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &types.HistoryTerminateWorkflowExecutionRequest{
		DomainUUID:       t.DomainUUID,
		TerminateRequest: ToTerminateWorkflowExecutionRequest(t.TerminateRequest),
	}
}

// FromFailoverMarkerTokenArray converts internal FailoverMarkerToken type array to thrift
func FromFailoverMarkerTokenArray(t []*types.FailoverMarkerToken) []*history.FailoverMarkerToken {
	if t == nil {
		return nil
	}
	v := make([]*history.FailoverMarkerToken, len(t))
	for i := range t {
		v[i] = FromFailoverMarkerToken(t[i])
	}
	return v
}

// ToFailoverMarkerTokenArray converts thrift FailoverMarkerToken type array to internal
func ToFailoverMarkerTokenArray(t []*history.FailoverMarkerToken) []*types.FailoverMarkerToken {
	if t == nil {
		return nil
	}
	v := make([]*types.FailoverMarkerToken, len(t))
	for i := range t {
		v[i] = ToFailoverMarkerToken(t[i])
	}
	return v
}

// FromProcessingQueueStateArray converts internal ProcessingQueueState type array to thrift
func FromProcessingQueueStateArray(t []*types.ProcessingQueueState) []*history.ProcessingQueueState {
	if t == nil {
		return nil
	}
	v := make([]*history.ProcessingQueueState, len(t))
	for i := range t {
		v[i] = FromProcessingQueueState(t[i])
	}
	return v
}

// ToProcessingQueueStateArray converts thrift ProcessingQueueState type array to internal
func ToProcessingQueueStateArray(t []*history.ProcessingQueueState) []*types.ProcessingQueueState {
	if t == nil {
		return nil
	}
	v := make([]*types.ProcessingQueueState, len(t))
	for i := range t {
		v[i] = ToProcessingQueueState(t[i])
	}
	return v
}

// FromProcessingQueueStateMap converts internal ProcessingQueueState type map to thrift
func FromProcessingQueueStateMap(t map[string]*types.ProcessingQueueState) map[string]*history.ProcessingQueueState {
	if t == nil {
		return nil
	}
	v := make(map[string]*history.ProcessingQueueState, len(t))
	for key := range t {
		v[key] = FromProcessingQueueState(t[key])
	}
	return v
}

// ToProcessingQueueStateMap converts thrift ProcessingQueueState type map to internal
func ToProcessingQueueStateMap(t map[string]*history.ProcessingQueueState) map[string]*types.ProcessingQueueState {
	if t == nil {
		return nil
	}
	v := make(map[string]*types.ProcessingQueueState, len(t))
	for key := range t {
		v[key] = ToProcessingQueueState(t[key])
	}
	return v
}

// FromProcessingQueueStateArrayMap converts internal ProcessingQueueState array map to thrift
func FromProcessingQueueStateArrayMap(t map[string][]*types.ProcessingQueueState) map[string][]*history.ProcessingQueueState {
	if t == nil {
		return nil
	}
	v := make(map[string][]*history.ProcessingQueueState, len(t))
	for key := range t {
		v[key] = FromProcessingQueueStateArray(t[key])
	}
	return v
}

// ToProcessingQueueStateArrayMap converts thrift ProcessingQueueState array map to internal
func ToProcessingQueueStateArrayMap(t map[string][]*history.ProcessingQueueState) map[string][]*types.ProcessingQueueState {
	if t == nil {
		return nil
	}
	v := make(map[string][]*types.ProcessingQueueState, len(t))
	for key := range t {
		v[key] = ToProcessingQueueStateArray(t[key])
	}
	return v
}
