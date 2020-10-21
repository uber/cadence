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

// FromHistoryService_CloseShard_Args converts internal HistoryService_CloseShard_Args type to thrift
func FromHistoryService_CloseShard_Args(t *types.HistoryService_CloseShard_Args) *history.HistoryService_CloseShard_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_CloseShard_Args{
		Request: FromCloseShardRequest(t.Request),
	}
}

// ToHistoryService_CloseShard_Args converts thrift HistoryService_CloseShard_Args type to internal
func ToHistoryService_CloseShard_Args(t *history.HistoryService_CloseShard_Args) *types.HistoryService_CloseShard_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_CloseShard_Args{
		Request: ToCloseShardRequest(t.Request),
	}
}

// FromHistoryService_CloseShard_Result converts internal HistoryService_CloseShard_Result type to thrift
func FromHistoryService_CloseShard_Result(t *types.HistoryService_CloseShard_Result) *history.HistoryService_CloseShard_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_CloseShard_Result{
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    FromAccessDeniedError(t.AccessDeniedError),
	}
}

// ToHistoryService_CloseShard_Result converts thrift HistoryService_CloseShard_Result type to internal
func ToHistoryService_CloseShard_Result(t *history.HistoryService_CloseShard_Result) *types.HistoryService_CloseShard_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_CloseShard_Result{
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    ToAccessDeniedError(t.AccessDeniedError),
	}
}

// FromHistoryService_DescribeHistoryHost_Args converts internal HistoryService_DescribeHistoryHost_Args type to thrift
func FromHistoryService_DescribeHistoryHost_Args(t *types.HistoryService_DescribeHistoryHost_Args) *history.HistoryService_DescribeHistoryHost_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_DescribeHistoryHost_Args{
		Request: FromDescribeHistoryHostRequest(t.Request),
	}
}

// ToHistoryService_DescribeHistoryHost_Args converts thrift HistoryService_DescribeHistoryHost_Args type to internal
func ToHistoryService_DescribeHistoryHost_Args(t *history.HistoryService_DescribeHistoryHost_Args) *types.HistoryService_DescribeHistoryHost_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_DescribeHistoryHost_Args{
		Request: ToDescribeHistoryHostRequest(t.Request),
	}
}

// FromHistoryService_DescribeHistoryHost_Result converts internal HistoryService_DescribeHistoryHost_Result type to thrift
func FromHistoryService_DescribeHistoryHost_Result(t *types.HistoryService_DescribeHistoryHost_Result) *history.HistoryService_DescribeHistoryHost_Result {
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

// ToHistoryService_DescribeHistoryHost_Result converts thrift HistoryService_DescribeHistoryHost_Result type to internal
func ToHistoryService_DescribeHistoryHost_Result(t *history.HistoryService_DescribeHistoryHost_Result) *types.HistoryService_DescribeHistoryHost_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_DescribeHistoryHost_Result{
		Success:              ToDescribeHistoryHostResponse(t.Success),
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    ToAccessDeniedError(t.AccessDeniedError),
	}
}

// FromHistoryService_DescribeMutableState_Args converts internal HistoryService_DescribeMutableState_Args type to thrift
func FromHistoryService_DescribeMutableState_Args(t *types.HistoryService_DescribeMutableState_Args) *history.HistoryService_DescribeMutableState_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_DescribeMutableState_Args{
		Request: FromDescribeMutableStateRequest(t.Request),
	}
}

// ToHistoryService_DescribeMutableState_Args converts thrift HistoryService_DescribeMutableState_Args type to internal
func ToHistoryService_DescribeMutableState_Args(t *history.HistoryService_DescribeMutableState_Args) *types.HistoryService_DescribeMutableState_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_DescribeMutableState_Args{
		Request: ToDescribeMutableStateRequest(t.Request),
	}
}

// FromHistoryService_DescribeMutableState_Result converts internal HistoryService_DescribeMutableState_Result type to thrift
func FromHistoryService_DescribeMutableState_Result(t *types.HistoryService_DescribeMutableState_Result) *history.HistoryService_DescribeMutableState_Result {
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

// ToHistoryService_DescribeMutableState_Result converts thrift HistoryService_DescribeMutableState_Result type to internal
func ToHistoryService_DescribeMutableState_Result(t *history.HistoryService_DescribeMutableState_Result) *types.HistoryService_DescribeMutableState_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_DescribeMutableState_Result{
		Success:                 ToDescribeMutableStateResponse(t.Success),
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		AccessDeniedError:       ToAccessDeniedError(t.AccessDeniedError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
	}
}

// FromHistoryService_DescribeQueue_Args converts internal HistoryService_DescribeQueue_Args type to thrift
func FromHistoryService_DescribeQueue_Args(t *types.HistoryService_DescribeQueue_Args) *history.HistoryService_DescribeQueue_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_DescribeQueue_Args{
		Request: FromDescribeQueueRequest(t.Request),
	}
}

// ToHistoryService_DescribeQueue_Args converts thrift HistoryService_DescribeQueue_Args type to internal
func ToHistoryService_DescribeQueue_Args(t *history.HistoryService_DescribeQueue_Args) *types.HistoryService_DescribeQueue_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_DescribeQueue_Args{
		Request: ToDescribeQueueRequest(t.Request),
	}
}

// FromHistoryService_DescribeQueue_Result converts internal HistoryService_DescribeQueue_Result type to thrift
func FromHistoryService_DescribeQueue_Result(t *types.HistoryService_DescribeQueue_Result) *history.HistoryService_DescribeQueue_Result {
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

// ToHistoryService_DescribeQueue_Result converts thrift HistoryService_DescribeQueue_Result type to internal
func ToHistoryService_DescribeQueue_Result(t *history.HistoryService_DescribeQueue_Result) *types.HistoryService_DescribeQueue_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_DescribeQueue_Result{
		Success:              ToDescribeQueueResponse(t.Success),
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    ToAccessDeniedError(t.AccessDeniedError),
	}
}

// FromHistoryService_DescribeWorkflowExecution_Args converts internal HistoryService_DescribeWorkflowExecution_Args type to thrift
func FromHistoryService_DescribeWorkflowExecution_Args(t *types.HistoryService_DescribeWorkflowExecution_Args) *history.HistoryService_DescribeWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_DescribeWorkflowExecution_Args{
		DescribeRequest: FromHistoryDescribeWorkflowExecutionRequest(t.DescribeRequest),
	}
}

// ToHistoryService_DescribeWorkflowExecution_Args converts thrift HistoryService_DescribeWorkflowExecution_Args type to internal
func ToHistoryService_DescribeWorkflowExecution_Args(t *history.HistoryService_DescribeWorkflowExecution_Args) *types.HistoryService_DescribeWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_DescribeWorkflowExecution_Args{
		DescribeRequest: ToHistoryDescribeWorkflowExecutionRequest(t.DescribeRequest),
	}
}

// FromHistoryService_DescribeWorkflowExecution_Result converts internal HistoryService_DescribeWorkflowExecution_Result type to thrift
func FromHistoryService_DescribeWorkflowExecution_Result(t *types.HistoryService_DescribeWorkflowExecution_Result) *history.HistoryService_DescribeWorkflowExecution_Result {
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

// ToHistoryService_DescribeWorkflowExecution_Result converts thrift HistoryService_DescribeWorkflowExecution_Result type to internal
func ToHistoryService_DescribeWorkflowExecution_Result(t *history.HistoryService_DescribeWorkflowExecution_Result) *types.HistoryService_DescribeWorkflowExecution_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_DescribeWorkflowExecution_Result{
		Success:                 ToDescribeWorkflowExecutionResponse(t.Success),
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryService_GetDLQReplicationMessages_Args converts internal HistoryService_GetDLQReplicationMessages_Args type to thrift
func FromHistoryService_GetDLQReplicationMessages_Args(t *types.HistoryService_GetDLQReplicationMessages_Args) *history.HistoryService_GetDLQReplicationMessages_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_GetDLQReplicationMessages_Args{
		Request: FromGetDLQReplicationMessagesRequest(t.Request),
	}
}

// ToHistoryService_GetDLQReplicationMessages_Args converts thrift HistoryService_GetDLQReplicationMessages_Args type to internal
func ToHistoryService_GetDLQReplicationMessages_Args(t *history.HistoryService_GetDLQReplicationMessages_Args) *types.HistoryService_GetDLQReplicationMessages_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_GetDLQReplicationMessages_Args{
		Request: ToGetDLQReplicationMessagesRequest(t.Request),
	}
}

// FromHistoryService_GetDLQReplicationMessages_Result converts internal HistoryService_GetDLQReplicationMessages_Result type to thrift
func FromHistoryService_GetDLQReplicationMessages_Result(t *types.HistoryService_GetDLQReplicationMessages_Result) *history.HistoryService_GetDLQReplicationMessages_Result {
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

// ToHistoryService_GetDLQReplicationMessages_Result converts thrift HistoryService_GetDLQReplicationMessages_Result type to internal
func ToHistoryService_GetDLQReplicationMessages_Result(t *history.HistoryService_GetDLQReplicationMessages_Result) *types.HistoryService_GetDLQReplicationMessages_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_GetDLQReplicationMessages_Result{
		Success:              ToGetDLQReplicationMessagesResponse(t.Success),
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:  ToEntityNotExistsError(t.EntityNotExistError),
	}
}

// FromHistoryService_GetMutableState_Args converts internal HistoryService_GetMutableState_Args type to thrift
func FromHistoryService_GetMutableState_Args(t *types.HistoryService_GetMutableState_Args) *history.HistoryService_GetMutableState_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_GetMutableState_Args{
		GetRequest: FromGetMutableStateRequest(t.GetRequest),
	}
}

// ToHistoryService_GetMutableState_Args converts thrift HistoryService_GetMutableState_Args type to internal
func ToHistoryService_GetMutableState_Args(t *history.HistoryService_GetMutableState_Args) *types.HistoryService_GetMutableState_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_GetMutableState_Args{
		GetRequest: ToGetMutableStateRequest(t.GetRequest),
	}
}

// FromHistoryService_GetMutableState_Result converts internal HistoryService_GetMutableState_Result type to thrift
func FromHistoryService_GetMutableState_Result(t *types.HistoryService_GetMutableState_Result) *history.HistoryService_GetMutableState_Result {
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

// ToHistoryService_GetMutableState_Result converts thrift HistoryService_GetMutableState_Result type to internal
func ToHistoryService_GetMutableState_Result(t *history.HistoryService_GetMutableState_Result) *types.HistoryService_GetMutableState_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_GetMutableState_Result{
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

// FromHistoryService_GetReplicationMessages_Args converts internal HistoryService_GetReplicationMessages_Args type to thrift
func FromHistoryService_GetReplicationMessages_Args(t *types.HistoryService_GetReplicationMessages_Args) *history.HistoryService_GetReplicationMessages_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_GetReplicationMessages_Args{
		Request: FromGetReplicationMessagesRequest(t.Request),
	}
}

// ToHistoryService_GetReplicationMessages_Args converts thrift HistoryService_GetReplicationMessages_Args type to internal
func ToHistoryService_GetReplicationMessages_Args(t *history.HistoryService_GetReplicationMessages_Args) *types.HistoryService_GetReplicationMessages_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_GetReplicationMessages_Args{
		Request: ToGetReplicationMessagesRequest(t.Request),
	}
}

// FromHistoryService_GetReplicationMessages_Result converts internal HistoryService_GetReplicationMessages_Result type to thrift
func FromHistoryService_GetReplicationMessages_Result(t *types.HistoryService_GetReplicationMessages_Result) *history.HistoryService_GetReplicationMessages_Result {
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

// ToHistoryService_GetReplicationMessages_Result converts thrift HistoryService_GetReplicationMessages_Result type to internal
func ToHistoryService_GetReplicationMessages_Result(t *history.HistoryService_GetReplicationMessages_Result) *types.HistoryService_GetReplicationMessages_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_GetReplicationMessages_Result{
		Success:                        ToGetReplicationMessagesResponse(t.Success),
		BadRequestError:                ToBadRequestError(t.BadRequestError),
		InternalServiceError:           ToInternalServiceError(t.InternalServiceError),
		LimitExceededError:             ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:               ToServiceBusyError(t.ServiceBusyError),
		ClientVersionNotSupportedError: ToClientVersionNotSupportedError(t.ClientVersionNotSupportedError),
	}
}

// FromHistoryService_MergeDLQMessages_Args converts internal HistoryService_MergeDLQMessages_Args type to thrift
func FromHistoryService_MergeDLQMessages_Args(t *types.HistoryService_MergeDLQMessages_Args) *history.HistoryService_MergeDLQMessages_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_MergeDLQMessages_Args{
		Request: FromMergeDLQMessagesRequest(t.Request),
	}
}

// ToHistoryService_MergeDLQMessages_Args converts thrift HistoryService_MergeDLQMessages_Args type to internal
func ToHistoryService_MergeDLQMessages_Args(t *history.HistoryService_MergeDLQMessages_Args) *types.HistoryService_MergeDLQMessages_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_MergeDLQMessages_Args{
		Request: ToMergeDLQMessagesRequest(t.Request),
	}
}

// FromHistoryService_MergeDLQMessages_Result converts internal HistoryService_MergeDLQMessages_Result type to thrift
func FromHistoryService_MergeDLQMessages_Result(t *types.HistoryService_MergeDLQMessages_Result) *history.HistoryService_MergeDLQMessages_Result {
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

// ToHistoryService_MergeDLQMessages_Result converts thrift HistoryService_MergeDLQMessages_Result type to internal
func ToHistoryService_MergeDLQMessages_Result(t *history.HistoryService_MergeDLQMessages_Result) *types.HistoryService_MergeDLQMessages_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_MergeDLQMessages_Result{
		Success:                 ToMergeDLQMessagesResponse(t.Success),
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
	}
}

// FromHistoryService_NotifyFailoverMarkers_Args converts internal HistoryService_NotifyFailoverMarkers_Args type to thrift
func FromHistoryService_NotifyFailoverMarkers_Args(t *types.HistoryService_NotifyFailoverMarkers_Args) *history.HistoryService_NotifyFailoverMarkers_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_NotifyFailoverMarkers_Args{
		Request: FromNotifyFailoverMarkersRequest(t.Request),
	}
}

// ToHistoryService_NotifyFailoverMarkers_Args converts thrift HistoryService_NotifyFailoverMarkers_Args type to internal
func ToHistoryService_NotifyFailoverMarkers_Args(t *history.HistoryService_NotifyFailoverMarkers_Args) *types.HistoryService_NotifyFailoverMarkers_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_NotifyFailoverMarkers_Args{
		Request: ToNotifyFailoverMarkersRequest(t.Request),
	}
}

// FromHistoryService_NotifyFailoverMarkers_Result converts internal HistoryService_NotifyFailoverMarkers_Result type to thrift
func FromHistoryService_NotifyFailoverMarkers_Result(t *types.HistoryService_NotifyFailoverMarkers_Result) *history.HistoryService_NotifyFailoverMarkers_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_NotifyFailoverMarkers_Result{
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToHistoryService_NotifyFailoverMarkers_Result converts thrift HistoryService_NotifyFailoverMarkers_Result type to internal
func ToHistoryService_NotifyFailoverMarkers_Result(t *history.HistoryService_NotifyFailoverMarkers_Result) *types.HistoryService_NotifyFailoverMarkers_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_NotifyFailoverMarkers_Result{
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryService_PollMutableState_Args converts internal HistoryService_PollMutableState_Args type to thrift
func FromHistoryService_PollMutableState_Args(t *types.HistoryService_PollMutableState_Args) *history.HistoryService_PollMutableState_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_PollMutableState_Args{
		PollRequest: FromPollMutableStateRequest(t.PollRequest),
	}
}

// ToHistoryService_PollMutableState_Args converts thrift HistoryService_PollMutableState_Args type to internal
func ToHistoryService_PollMutableState_Args(t *history.HistoryService_PollMutableState_Args) *types.HistoryService_PollMutableState_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_PollMutableState_Args{
		PollRequest: ToPollMutableStateRequest(t.PollRequest),
	}
}

// FromHistoryService_PollMutableState_Result converts internal HistoryService_PollMutableState_Result type to thrift
func FromHistoryService_PollMutableState_Result(t *types.HistoryService_PollMutableState_Result) *history.HistoryService_PollMutableState_Result {
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

// ToHistoryService_PollMutableState_Result converts thrift HistoryService_PollMutableState_Result type to internal
func ToHistoryService_PollMutableState_Result(t *history.HistoryService_PollMutableState_Result) *types.HistoryService_PollMutableState_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_PollMutableState_Result{
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

// FromHistoryService_PurgeDLQMessages_Args converts internal HistoryService_PurgeDLQMessages_Args type to thrift
func FromHistoryService_PurgeDLQMessages_Args(t *types.HistoryService_PurgeDLQMessages_Args) *history.HistoryService_PurgeDLQMessages_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_PurgeDLQMessages_Args{
		Request: FromPurgeDLQMessagesRequest(t.Request),
	}
}

// ToHistoryService_PurgeDLQMessages_Args converts thrift HistoryService_PurgeDLQMessages_Args type to internal
func ToHistoryService_PurgeDLQMessages_Args(t *history.HistoryService_PurgeDLQMessages_Args) *types.HistoryService_PurgeDLQMessages_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_PurgeDLQMessages_Args{
		Request: ToPurgeDLQMessagesRequest(t.Request),
	}
}

// FromHistoryService_PurgeDLQMessages_Result converts internal HistoryService_PurgeDLQMessages_Result type to thrift
func FromHistoryService_PurgeDLQMessages_Result(t *types.HistoryService_PurgeDLQMessages_Result) *history.HistoryService_PurgeDLQMessages_Result {
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

// ToHistoryService_PurgeDLQMessages_Result converts thrift HistoryService_PurgeDLQMessages_Result type to internal
func ToHistoryService_PurgeDLQMessages_Result(t *history.HistoryService_PurgeDLQMessages_Result) *types.HistoryService_PurgeDLQMessages_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_PurgeDLQMessages_Result{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
	}
}

// FromHistoryService_QueryWorkflow_Args converts internal HistoryService_QueryWorkflow_Args type to thrift
func FromHistoryService_QueryWorkflow_Args(t *types.HistoryService_QueryWorkflow_Args) *history.HistoryService_QueryWorkflow_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_QueryWorkflow_Args{
		QueryRequest: FromHistoryQueryWorkflowRequest(t.QueryRequest),
	}
}

// ToHistoryService_QueryWorkflow_Args converts thrift HistoryService_QueryWorkflow_Args type to internal
func ToHistoryService_QueryWorkflow_Args(t *history.HistoryService_QueryWorkflow_Args) *types.HistoryService_QueryWorkflow_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_QueryWorkflow_Args{
		QueryRequest: ToHistoryQueryWorkflowRequest(t.QueryRequest),
	}
}

// FromHistoryService_QueryWorkflow_Result converts internal HistoryService_QueryWorkflow_Result type to thrift
func FromHistoryService_QueryWorkflow_Result(t *types.HistoryService_QueryWorkflow_Result) *history.HistoryService_QueryWorkflow_Result {
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

// ToHistoryService_QueryWorkflow_Result converts thrift HistoryService_QueryWorkflow_Result type to internal
func ToHistoryService_QueryWorkflow_Result(t *history.HistoryService_QueryWorkflow_Result) *types.HistoryService_QueryWorkflow_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_QueryWorkflow_Result{
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

// FromHistoryService_ReadDLQMessages_Args converts internal HistoryService_ReadDLQMessages_Args type to thrift
func FromHistoryService_ReadDLQMessages_Args(t *types.HistoryService_ReadDLQMessages_Args) *history.HistoryService_ReadDLQMessages_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ReadDLQMessages_Args{
		Request: FromReadDLQMessagesRequest(t.Request),
	}
}

// ToHistoryService_ReadDLQMessages_Args converts thrift HistoryService_ReadDLQMessages_Args type to internal
func ToHistoryService_ReadDLQMessages_Args(t *history.HistoryService_ReadDLQMessages_Args) *types.HistoryService_ReadDLQMessages_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_ReadDLQMessages_Args{
		Request: ToReadDLQMessagesRequest(t.Request),
	}
}

// FromHistoryService_ReadDLQMessages_Result converts internal HistoryService_ReadDLQMessages_Result type to thrift
func FromHistoryService_ReadDLQMessages_Result(t *types.HistoryService_ReadDLQMessages_Result) *history.HistoryService_ReadDLQMessages_Result {
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

// ToHistoryService_ReadDLQMessages_Result converts thrift HistoryService_ReadDLQMessages_Result type to internal
func ToHistoryService_ReadDLQMessages_Result(t *history.HistoryService_ReadDLQMessages_Result) *types.HistoryService_ReadDLQMessages_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_ReadDLQMessages_Result{
		Success:                 ToReadDLQMessagesResponse(t.Success),
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
	}
}

// FromHistoryService_ReapplyEvents_Args converts internal HistoryService_ReapplyEvents_Args type to thrift
func FromHistoryService_ReapplyEvents_Args(t *types.HistoryService_ReapplyEvents_Args) *history.HistoryService_ReapplyEvents_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ReapplyEvents_Args{
		ReapplyEventsRequest: FromHistoryReapplyEventsRequest(t.ReapplyEventsRequest),
	}
}

// ToHistoryService_ReapplyEvents_Args converts thrift HistoryService_ReapplyEvents_Args type to internal
func ToHistoryService_ReapplyEvents_Args(t *history.HistoryService_ReapplyEvents_Args) *types.HistoryService_ReapplyEvents_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_ReapplyEvents_Args{
		ReapplyEventsRequest: ToHistoryReapplyEventsRequest(t.ReapplyEventsRequest),
	}
}

// FromHistoryService_ReapplyEvents_Result converts internal HistoryService_ReapplyEvents_Result type to thrift
func FromHistoryService_ReapplyEvents_Result(t *types.HistoryService_ReapplyEvents_Result) *history.HistoryService_ReapplyEvents_Result {
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

// ToHistoryService_ReapplyEvents_Result converts thrift HistoryService_ReapplyEvents_Result type to internal
func ToHistoryService_ReapplyEvents_Result(t *history.HistoryService_ReapplyEvents_Result) *types.HistoryService_ReapplyEvents_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_ReapplyEvents_Result{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
	}
}

// FromHistoryService_RecordActivityTaskHeartbeat_Args converts internal HistoryService_RecordActivityTaskHeartbeat_Args type to thrift
func FromHistoryService_RecordActivityTaskHeartbeat_Args(t *types.HistoryService_RecordActivityTaskHeartbeat_Args) *history.HistoryService_RecordActivityTaskHeartbeat_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RecordActivityTaskHeartbeat_Args{
		HeartbeatRequest: FromHistoryRecordActivityTaskHeartbeatRequest(t.HeartbeatRequest),
	}
}

// ToHistoryService_RecordActivityTaskHeartbeat_Args converts thrift HistoryService_RecordActivityTaskHeartbeat_Args type to internal
func ToHistoryService_RecordActivityTaskHeartbeat_Args(t *history.HistoryService_RecordActivityTaskHeartbeat_Args) *types.HistoryService_RecordActivityTaskHeartbeat_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RecordActivityTaskHeartbeat_Args{
		HeartbeatRequest: ToHistoryRecordActivityTaskHeartbeatRequest(t.HeartbeatRequest),
	}
}

// FromHistoryService_RecordActivityTaskHeartbeat_Result converts internal HistoryService_RecordActivityTaskHeartbeat_Result type to thrift
func FromHistoryService_RecordActivityTaskHeartbeat_Result(t *types.HistoryService_RecordActivityTaskHeartbeat_Result) *history.HistoryService_RecordActivityTaskHeartbeat_Result {
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

// ToHistoryService_RecordActivityTaskHeartbeat_Result converts thrift HistoryService_RecordActivityTaskHeartbeat_Result type to internal
func ToHistoryService_RecordActivityTaskHeartbeat_Result(t *history.HistoryService_RecordActivityTaskHeartbeat_Result) *types.HistoryService_RecordActivityTaskHeartbeat_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RecordActivityTaskHeartbeat_Result{
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

// FromHistoryService_RecordActivityTaskStarted_Args converts internal HistoryService_RecordActivityTaskStarted_Args type to thrift
func FromHistoryService_RecordActivityTaskStarted_Args(t *types.HistoryService_RecordActivityTaskStarted_Args) *history.HistoryService_RecordActivityTaskStarted_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RecordActivityTaskStarted_Args{
		AddRequest: FromRecordActivityTaskStartedRequest(t.AddRequest),
	}
}

// ToHistoryService_RecordActivityTaskStarted_Args converts thrift HistoryService_RecordActivityTaskStarted_Args type to internal
func ToHistoryService_RecordActivityTaskStarted_Args(t *history.HistoryService_RecordActivityTaskStarted_Args) *types.HistoryService_RecordActivityTaskStarted_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RecordActivityTaskStarted_Args{
		AddRequest: ToRecordActivityTaskStartedRequest(t.AddRequest),
	}
}

// FromHistoryService_RecordActivityTaskStarted_Result converts internal HistoryService_RecordActivityTaskStarted_Result type to thrift
func FromHistoryService_RecordActivityTaskStarted_Result(t *types.HistoryService_RecordActivityTaskStarted_Result) *history.HistoryService_RecordActivityTaskStarted_Result {
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

// ToHistoryService_RecordActivityTaskStarted_Result converts thrift HistoryService_RecordActivityTaskStarted_Result type to internal
func ToHistoryService_RecordActivityTaskStarted_Result(t *history.HistoryService_RecordActivityTaskStarted_Result) *types.HistoryService_RecordActivityTaskStarted_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RecordActivityTaskStarted_Result{
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

// FromHistoryService_RecordChildExecutionCompleted_Args converts internal HistoryService_RecordChildExecutionCompleted_Args type to thrift
func FromHistoryService_RecordChildExecutionCompleted_Args(t *types.HistoryService_RecordChildExecutionCompleted_Args) *history.HistoryService_RecordChildExecutionCompleted_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RecordChildExecutionCompleted_Args{
		CompletionRequest: FromRecordChildExecutionCompletedRequest(t.CompletionRequest),
	}
}

// ToHistoryService_RecordChildExecutionCompleted_Args converts thrift HistoryService_RecordChildExecutionCompleted_Args type to internal
func ToHistoryService_RecordChildExecutionCompleted_Args(t *history.HistoryService_RecordChildExecutionCompleted_Args) *types.HistoryService_RecordChildExecutionCompleted_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RecordChildExecutionCompleted_Args{
		CompletionRequest: ToRecordChildExecutionCompletedRequest(t.CompletionRequest),
	}
}

// FromHistoryService_RecordChildExecutionCompleted_Result converts internal HistoryService_RecordChildExecutionCompleted_Result type to thrift
func FromHistoryService_RecordChildExecutionCompleted_Result(t *types.HistoryService_RecordChildExecutionCompleted_Result) *history.HistoryService_RecordChildExecutionCompleted_Result {
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

// ToHistoryService_RecordChildExecutionCompleted_Result converts thrift HistoryService_RecordChildExecutionCompleted_Result type to internal
func ToHistoryService_RecordChildExecutionCompleted_Result(t *history.HistoryService_RecordChildExecutionCompleted_Result) *types.HistoryService_RecordChildExecutionCompleted_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RecordChildExecutionCompleted_Result{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryService_RecordDecisionTaskStarted_Args converts internal HistoryService_RecordDecisionTaskStarted_Args type to thrift
func FromHistoryService_RecordDecisionTaskStarted_Args(t *types.HistoryService_RecordDecisionTaskStarted_Args) *history.HistoryService_RecordDecisionTaskStarted_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RecordDecisionTaskStarted_Args{
		AddRequest: FromRecordDecisionTaskStartedRequest(t.AddRequest),
	}
}

// ToHistoryService_RecordDecisionTaskStarted_Args converts thrift HistoryService_RecordDecisionTaskStarted_Args type to internal
func ToHistoryService_RecordDecisionTaskStarted_Args(t *history.HistoryService_RecordDecisionTaskStarted_Args) *types.HistoryService_RecordDecisionTaskStarted_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RecordDecisionTaskStarted_Args{
		AddRequest: ToRecordDecisionTaskStartedRequest(t.AddRequest),
	}
}

// FromHistoryService_RecordDecisionTaskStarted_Result converts internal HistoryService_RecordDecisionTaskStarted_Result type to thrift
func FromHistoryService_RecordDecisionTaskStarted_Result(t *types.HistoryService_RecordDecisionTaskStarted_Result) *history.HistoryService_RecordDecisionTaskStarted_Result {
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

// ToHistoryService_RecordDecisionTaskStarted_Result converts thrift HistoryService_RecordDecisionTaskStarted_Result type to internal
func ToHistoryService_RecordDecisionTaskStarted_Result(t *history.HistoryService_RecordDecisionTaskStarted_Result) *types.HistoryService_RecordDecisionTaskStarted_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RecordDecisionTaskStarted_Result{
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

// FromHistoryService_RefreshWorkflowTasks_Args converts internal HistoryService_RefreshWorkflowTasks_Args type to thrift
func FromHistoryService_RefreshWorkflowTasks_Args(t *types.HistoryService_RefreshWorkflowTasks_Args) *history.HistoryService_RefreshWorkflowTasks_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RefreshWorkflowTasks_Args{
		Request: FromHistoryRefreshWorkflowTasksRequest(t.Request),
	}
}

// ToHistoryService_RefreshWorkflowTasks_Args converts thrift HistoryService_RefreshWorkflowTasks_Args type to internal
func ToHistoryService_RefreshWorkflowTasks_Args(t *history.HistoryService_RefreshWorkflowTasks_Args) *types.HistoryService_RefreshWorkflowTasks_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RefreshWorkflowTasks_Args{
		Request: ToHistoryRefreshWorkflowTasksRequest(t.Request),
	}
}

// FromHistoryService_RefreshWorkflowTasks_Result converts internal HistoryService_RefreshWorkflowTasks_Result type to thrift
func FromHistoryService_RefreshWorkflowTasks_Result(t *types.HistoryService_RefreshWorkflowTasks_Result) *history.HistoryService_RefreshWorkflowTasks_Result {
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

// ToHistoryService_RefreshWorkflowTasks_Result converts thrift HistoryService_RefreshWorkflowTasks_Result type to internal
func ToHistoryService_RefreshWorkflowTasks_Result(t *history.HistoryService_RefreshWorkflowTasks_Result) *types.HistoryService_RefreshWorkflowTasks_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RefreshWorkflowTasks_Result{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
	}
}

// FromHistoryService_RemoveSignalMutableState_Args converts internal HistoryService_RemoveSignalMutableState_Args type to thrift
func FromHistoryService_RemoveSignalMutableState_Args(t *types.HistoryService_RemoveSignalMutableState_Args) *history.HistoryService_RemoveSignalMutableState_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RemoveSignalMutableState_Args{
		RemoveRequest: FromRemoveSignalMutableStateRequest(t.RemoveRequest),
	}
}

// ToHistoryService_RemoveSignalMutableState_Args converts thrift HistoryService_RemoveSignalMutableState_Args type to internal
func ToHistoryService_RemoveSignalMutableState_Args(t *history.HistoryService_RemoveSignalMutableState_Args) *types.HistoryService_RemoveSignalMutableState_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RemoveSignalMutableState_Args{
		RemoveRequest: ToRemoveSignalMutableStateRequest(t.RemoveRequest),
	}
}

// FromHistoryService_RemoveSignalMutableState_Result converts internal HistoryService_RemoveSignalMutableState_Result type to thrift
func FromHistoryService_RemoveSignalMutableState_Result(t *types.HistoryService_RemoveSignalMutableState_Result) *history.HistoryService_RemoveSignalMutableState_Result {
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

// ToHistoryService_RemoveSignalMutableState_Result converts thrift HistoryService_RemoveSignalMutableState_Result type to internal
func ToHistoryService_RemoveSignalMutableState_Result(t *history.HistoryService_RemoveSignalMutableState_Result) *types.HistoryService_RemoveSignalMutableState_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RemoveSignalMutableState_Result{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryService_RemoveTask_Args converts internal HistoryService_RemoveTask_Args type to thrift
func FromHistoryService_RemoveTask_Args(t *types.HistoryService_RemoveTask_Args) *history.HistoryService_RemoveTask_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RemoveTask_Args{
		Request: FromRemoveTaskRequest(t.Request),
	}
}

// ToHistoryService_RemoveTask_Args converts thrift HistoryService_RemoveTask_Args type to internal
func ToHistoryService_RemoveTask_Args(t *history.HistoryService_RemoveTask_Args) *types.HistoryService_RemoveTask_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RemoveTask_Args{
		Request: ToRemoveTaskRequest(t.Request),
	}
}

// FromHistoryService_RemoveTask_Result converts internal HistoryService_RemoveTask_Result type to thrift
func FromHistoryService_RemoveTask_Result(t *types.HistoryService_RemoveTask_Result) *history.HistoryService_RemoveTask_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RemoveTask_Result{
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    FromAccessDeniedError(t.AccessDeniedError),
	}
}

// ToHistoryService_RemoveTask_Result converts thrift HistoryService_RemoveTask_Result type to internal
func ToHistoryService_RemoveTask_Result(t *history.HistoryService_RemoveTask_Result) *types.HistoryService_RemoveTask_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RemoveTask_Result{
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    ToAccessDeniedError(t.AccessDeniedError),
	}
}

// FromHistoryService_ReplicateEventsV2_Args converts internal HistoryService_ReplicateEventsV2_Args type to thrift
func FromHistoryService_ReplicateEventsV2_Args(t *types.HistoryService_ReplicateEventsV2_Args) *history.HistoryService_ReplicateEventsV2_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ReplicateEventsV2_Args{
		ReplicateV2Request: FromReplicateEventsV2Request(t.ReplicateV2Request),
	}
}

// ToHistoryService_ReplicateEventsV2_Args converts thrift HistoryService_ReplicateEventsV2_Args type to internal
func ToHistoryService_ReplicateEventsV2_Args(t *history.HistoryService_ReplicateEventsV2_Args) *types.HistoryService_ReplicateEventsV2_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_ReplicateEventsV2_Args{
		ReplicateV2Request: ToReplicateEventsV2Request(t.ReplicateV2Request),
	}
}

// FromHistoryService_ReplicateEventsV2_Result converts internal HistoryService_ReplicateEventsV2_Result type to thrift
func FromHistoryService_ReplicateEventsV2_Result(t *types.HistoryService_ReplicateEventsV2_Result) *history.HistoryService_ReplicateEventsV2_Result {
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

// ToHistoryService_ReplicateEventsV2_Result converts thrift HistoryService_ReplicateEventsV2_Result type to internal
func ToHistoryService_ReplicateEventsV2_Result(t *history.HistoryService_ReplicateEventsV2_Result) *types.HistoryService_ReplicateEventsV2_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_ReplicateEventsV2_Result{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		RetryTaskError:          ToRetryTaskV2Error(t.RetryTaskError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryService_RequestCancelWorkflowExecution_Args converts internal HistoryService_RequestCancelWorkflowExecution_Args type to thrift
func FromHistoryService_RequestCancelWorkflowExecution_Args(t *types.HistoryService_RequestCancelWorkflowExecution_Args) *history.HistoryService_RequestCancelWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RequestCancelWorkflowExecution_Args{
		CancelRequest: FromHistoryRequestCancelWorkflowExecutionRequest(t.CancelRequest),
	}
}

// ToHistoryService_RequestCancelWorkflowExecution_Args converts thrift HistoryService_RequestCancelWorkflowExecution_Args type to internal
func ToHistoryService_RequestCancelWorkflowExecution_Args(t *history.HistoryService_RequestCancelWorkflowExecution_Args) *types.HistoryService_RequestCancelWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RequestCancelWorkflowExecution_Args{
		CancelRequest: ToHistoryRequestCancelWorkflowExecutionRequest(t.CancelRequest),
	}
}

// FromHistoryService_RequestCancelWorkflowExecution_Result converts internal HistoryService_RequestCancelWorkflowExecution_Result type to thrift
func FromHistoryService_RequestCancelWorkflowExecution_Result(t *types.HistoryService_RequestCancelWorkflowExecution_Result) *history.HistoryService_RequestCancelWorkflowExecution_Result {
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

// ToHistoryService_RequestCancelWorkflowExecution_Result converts thrift HistoryService_RequestCancelWorkflowExecution_Result type to internal
func ToHistoryService_RequestCancelWorkflowExecution_Result(t *history.HistoryService_RequestCancelWorkflowExecution_Result) *types.HistoryService_RequestCancelWorkflowExecution_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RequestCancelWorkflowExecution_Result{
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

// FromHistoryService_ResetQueue_Args converts internal HistoryService_ResetQueue_Args type to thrift
func FromHistoryService_ResetQueue_Args(t *types.HistoryService_ResetQueue_Args) *history.HistoryService_ResetQueue_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ResetQueue_Args{
		Request: FromResetQueueRequest(t.Request),
	}
}

// ToHistoryService_ResetQueue_Args converts thrift HistoryService_ResetQueue_Args type to internal
func ToHistoryService_ResetQueue_Args(t *history.HistoryService_ResetQueue_Args) *types.HistoryService_ResetQueue_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_ResetQueue_Args{
		Request: ToResetQueueRequest(t.Request),
	}
}

// FromHistoryService_ResetQueue_Result converts internal HistoryService_ResetQueue_Result type to thrift
func FromHistoryService_ResetQueue_Result(t *types.HistoryService_ResetQueue_Result) *history.HistoryService_ResetQueue_Result {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ResetQueue_Result{
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    FromAccessDeniedError(t.AccessDeniedError),
	}
}

// ToHistoryService_ResetQueue_Result converts thrift HistoryService_ResetQueue_Result type to internal
func ToHistoryService_ResetQueue_Result(t *history.HistoryService_ResetQueue_Result) *types.HistoryService_ResetQueue_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_ResetQueue_Result{
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    ToAccessDeniedError(t.AccessDeniedError),
	}
}

// FromHistoryService_ResetStickyTaskList_Args converts internal HistoryService_ResetStickyTaskList_Args type to thrift
func FromHistoryService_ResetStickyTaskList_Args(t *types.HistoryService_ResetStickyTaskList_Args) *history.HistoryService_ResetStickyTaskList_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ResetStickyTaskList_Args{
		ResetRequest: FromHistoryResetStickyTaskListRequest(t.ResetRequest),
	}
}

// ToHistoryService_ResetStickyTaskList_Args converts thrift HistoryService_ResetStickyTaskList_Args type to internal
func ToHistoryService_ResetStickyTaskList_Args(t *history.HistoryService_ResetStickyTaskList_Args) *types.HistoryService_ResetStickyTaskList_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_ResetStickyTaskList_Args{
		ResetRequest: ToHistoryResetStickyTaskListRequest(t.ResetRequest),
	}
}

// FromHistoryService_ResetStickyTaskList_Result converts internal HistoryService_ResetStickyTaskList_Result type to thrift
func FromHistoryService_ResetStickyTaskList_Result(t *types.HistoryService_ResetStickyTaskList_Result) *history.HistoryService_ResetStickyTaskList_Result {
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

// ToHistoryService_ResetStickyTaskList_Result converts thrift HistoryService_ResetStickyTaskList_Result type to internal
func ToHistoryService_ResetStickyTaskList_Result(t *history.HistoryService_ResetStickyTaskList_Result) *types.HistoryService_ResetStickyTaskList_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_ResetStickyTaskList_Result{
		Success:                 ToHistoryResetStickyTaskListResponse(t.Success),
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryService_ResetWorkflowExecution_Args converts internal HistoryService_ResetWorkflowExecution_Args type to thrift
func FromHistoryService_ResetWorkflowExecution_Args(t *types.HistoryService_ResetWorkflowExecution_Args) *history.HistoryService_ResetWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ResetWorkflowExecution_Args{
		ResetRequest: FromHistoryResetWorkflowExecutionRequest(t.ResetRequest),
	}
}

// ToHistoryService_ResetWorkflowExecution_Args converts thrift HistoryService_ResetWorkflowExecution_Args type to internal
func ToHistoryService_ResetWorkflowExecution_Args(t *history.HistoryService_ResetWorkflowExecution_Args) *types.HistoryService_ResetWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_ResetWorkflowExecution_Args{
		ResetRequest: ToHistoryResetWorkflowExecutionRequest(t.ResetRequest),
	}
}

// FromHistoryService_ResetWorkflowExecution_Result converts internal HistoryService_ResetWorkflowExecution_Result type to thrift
func FromHistoryService_ResetWorkflowExecution_Result(t *types.HistoryService_ResetWorkflowExecution_Result) *history.HistoryService_ResetWorkflowExecution_Result {
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

// ToHistoryService_ResetWorkflowExecution_Result converts thrift HistoryService_ResetWorkflowExecution_Result type to internal
func ToHistoryService_ResetWorkflowExecution_Result(t *history.HistoryService_ResetWorkflowExecution_Result) *types.HistoryService_ResetWorkflowExecution_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_ResetWorkflowExecution_Result{
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

// FromHistoryService_RespondActivityTaskCanceled_Args converts internal HistoryService_RespondActivityTaskCanceled_Args type to thrift
func FromHistoryService_RespondActivityTaskCanceled_Args(t *types.HistoryService_RespondActivityTaskCanceled_Args) *history.HistoryService_RespondActivityTaskCanceled_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RespondActivityTaskCanceled_Args{
		CanceledRequest: FromHistoryRespondActivityTaskCanceledRequest(t.CanceledRequest),
	}
}

// ToHistoryService_RespondActivityTaskCanceled_Args converts thrift HistoryService_RespondActivityTaskCanceled_Args type to internal
func ToHistoryService_RespondActivityTaskCanceled_Args(t *history.HistoryService_RespondActivityTaskCanceled_Args) *types.HistoryService_RespondActivityTaskCanceled_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RespondActivityTaskCanceled_Args{
		CanceledRequest: ToHistoryRespondActivityTaskCanceledRequest(t.CanceledRequest),
	}
}

// FromHistoryService_RespondActivityTaskCanceled_Result converts internal HistoryService_RespondActivityTaskCanceled_Result type to thrift
func FromHistoryService_RespondActivityTaskCanceled_Result(t *types.HistoryService_RespondActivityTaskCanceled_Result) *history.HistoryService_RespondActivityTaskCanceled_Result {
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

// ToHistoryService_RespondActivityTaskCanceled_Result converts thrift HistoryService_RespondActivityTaskCanceled_Result type to internal
func ToHistoryService_RespondActivityTaskCanceled_Result(t *history.HistoryService_RespondActivityTaskCanceled_Result) *types.HistoryService_RespondActivityTaskCanceled_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RespondActivityTaskCanceled_Result{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryService_RespondActivityTaskCompleted_Args converts internal HistoryService_RespondActivityTaskCompleted_Args type to thrift
func FromHistoryService_RespondActivityTaskCompleted_Args(t *types.HistoryService_RespondActivityTaskCompleted_Args) *history.HistoryService_RespondActivityTaskCompleted_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RespondActivityTaskCompleted_Args{
		CompleteRequest: FromHistoryRespondActivityTaskCompletedRequest(t.CompleteRequest),
	}
}

// ToHistoryService_RespondActivityTaskCompleted_Args converts thrift HistoryService_RespondActivityTaskCompleted_Args type to internal
func ToHistoryService_RespondActivityTaskCompleted_Args(t *history.HistoryService_RespondActivityTaskCompleted_Args) *types.HistoryService_RespondActivityTaskCompleted_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RespondActivityTaskCompleted_Args{
		CompleteRequest: ToHistoryRespondActivityTaskCompletedRequest(t.CompleteRequest),
	}
}

// FromHistoryService_RespondActivityTaskCompleted_Result converts internal HistoryService_RespondActivityTaskCompleted_Result type to thrift
func FromHistoryService_RespondActivityTaskCompleted_Result(t *types.HistoryService_RespondActivityTaskCompleted_Result) *history.HistoryService_RespondActivityTaskCompleted_Result {
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

// ToHistoryService_RespondActivityTaskCompleted_Result converts thrift HistoryService_RespondActivityTaskCompleted_Result type to internal
func ToHistoryService_RespondActivityTaskCompleted_Result(t *history.HistoryService_RespondActivityTaskCompleted_Result) *types.HistoryService_RespondActivityTaskCompleted_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RespondActivityTaskCompleted_Result{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryService_RespondActivityTaskFailed_Args converts internal HistoryService_RespondActivityTaskFailed_Args type to thrift
func FromHistoryService_RespondActivityTaskFailed_Args(t *types.HistoryService_RespondActivityTaskFailed_Args) *history.HistoryService_RespondActivityTaskFailed_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RespondActivityTaskFailed_Args{
		FailRequest: FromHistoryRespondActivityTaskFailedRequest(t.FailRequest),
	}
}

// ToHistoryService_RespondActivityTaskFailed_Args converts thrift HistoryService_RespondActivityTaskFailed_Args type to internal
func ToHistoryService_RespondActivityTaskFailed_Args(t *history.HistoryService_RespondActivityTaskFailed_Args) *types.HistoryService_RespondActivityTaskFailed_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RespondActivityTaskFailed_Args{
		FailRequest: ToHistoryRespondActivityTaskFailedRequest(t.FailRequest),
	}
}

// FromHistoryService_RespondActivityTaskFailed_Result converts internal HistoryService_RespondActivityTaskFailed_Result type to thrift
func FromHistoryService_RespondActivityTaskFailed_Result(t *types.HistoryService_RespondActivityTaskFailed_Result) *history.HistoryService_RespondActivityTaskFailed_Result {
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

// ToHistoryService_RespondActivityTaskFailed_Result converts thrift HistoryService_RespondActivityTaskFailed_Result type to internal
func ToHistoryService_RespondActivityTaskFailed_Result(t *history.HistoryService_RespondActivityTaskFailed_Result) *types.HistoryService_RespondActivityTaskFailed_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RespondActivityTaskFailed_Result{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryService_RespondDecisionTaskCompleted_Args converts internal HistoryService_RespondDecisionTaskCompleted_Args type to thrift
func FromHistoryService_RespondDecisionTaskCompleted_Args(t *types.HistoryService_RespondDecisionTaskCompleted_Args) *history.HistoryService_RespondDecisionTaskCompleted_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RespondDecisionTaskCompleted_Args{
		CompleteRequest: FromHistoryRespondDecisionTaskCompletedRequest(t.CompleteRequest),
	}
}

// ToHistoryService_RespondDecisionTaskCompleted_Args converts thrift HistoryService_RespondDecisionTaskCompleted_Args type to internal
func ToHistoryService_RespondDecisionTaskCompleted_Args(t *history.HistoryService_RespondDecisionTaskCompleted_Args) *types.HistoryService_RespondDecisionTaskCompleted_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RespondDecisionTaskCompleted_Args{
		CompleteRequest: ToHistoryRespondDecisionTaskCompletedRequest(t.CompleteRequest),
	}
}

// FromHistoryService_RespondDecisionTaskCompleted_Result converts internal HistoryService_RespondDecisionTaskCompleted_Result type to thrift
func FromHistoryService_RespondDecisionTaskCompleted_Result(t *types.HistoryService_RespondDecisionTaskCompleted_Result) *history.HistoryService_RespondDecisionTaskCompleted_Result {
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

// ToHistoryService_RespondDecisionTaskCompleted_Result converts thrift HistoryService_RespondDecisionTaskCompleted_Result type to internal
func ToHistoryService_RespondDecisionTaskCompleted_Result(t *history.HistoryService_RespondDecisionTaskCompleted_Result) *types.HistoryService_RespondDecisionTaskCompleted_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RespondDecisionTaskCompleted_Result{
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

// FromHistoryService_RespondDecisionTaskFailed_Args converts internal HistoryService_RespondDecisionTaskFailed_Args type to thrift
func FromHistoryService_RespondDecisionTaskFailed_Args(t *types.HistoryService_RespondDecisionTaskFailed_Args) *history.HistoryService_RespondDecisionTaskFailed_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_RespondDecisionTaskFailed_Args{
		FailedRequest: FromHistoryRespondDecisionTaskFailedRequest(t.FailedRequest),
	}
}

// ToHistoryService_RespondDecisionTaskFailed_Args converts thrift HistoryService_RespondDecisionTaskFailed_Args type to internal
func ToHistoryService_RespondDecisionTaskFailed_Args(t *history.HistoryService_RespondDecisionTaskFailed_Args) *types.HistoryService_RespondDecisionTaskFailed_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RespondDecisionTaskFailed_Args{
		FailedRequest: ToHistoryRespondDecisionTaskFailedRequest(t.FailedRequest),
	}
}

// FromHistoryService_RespondDecisionTaskFailed_Result converts internal HistoryService_RespondDecisionTaskFailed_Result type to thrift
func FromHistoryService_RespondDecisionTaskFailed_Result(t *types.HistoryService_RespondDecisionTaskFailed_Result) *history.HistoryService_RespondDecisionTaskFailed_Result {
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

// ToHistoryService_RespondDecisionTaskFailed_Result converts thrift HistoryService_RespondDecisionTaskFailed_Result type to internal
func ToHistoryService_RespondDecisionTaskFailed_Result(t *history.HistoryService_RespondDecisionTaskFailed_Result) *types.HistoryService_RespondDecisionTaskFailed_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_RespondDecisionTaskFailed_Result{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryService_ScheduleDecisionTask_Args converts internal HistoryService_ScheduleDecisionTask_Args type to thrift
func FromHistoryService_ScheduleDecisionTask_Args(t *types.HistoryService_ScheduleDecisionTask_Args) *history.HistoryService_ScheduleDecisionTask_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_ScheduleDecisionTask_Args{
		ScheduleRequest: FromScheduleDecisionTaskRequest(t.ScheduleRequest),
	}
}

// ToHistoryService_ScheduleDecisionTask_Args converts thrift HistoryService_ScheduleDecisionTask_Args type to internal
func ToHistoryService_ScheduleDecisionTask_Args(t *history.HistoryService_ScheduleDecisionTask_Args) *types.HistoryService_ScheduleDecisionTask_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_ScheduleDecisionTask_Args{
		ScheduleRequest: ToScheduleDecisionTaskRequest(t.ScheduleRequest),
	}
}

// FromHistoryService_ScheduleDecisionTask_Result converts internal HistoryService_ScheduleDecisionTask_Result type to thrift
func FromHistoryService_ScheduleDecisionTask_Result(t *types.HistoryService_ScheduleDecisionTask_Result) *history.HistoryService_ScheduleDecisionTask_Result {
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

// ToHistoryService_ScheduleDecisionTask_Result converts thrift HistoryService_ScheduleDecisionTask_Result type to internal
func ToHistoryService_ScheduleDecisionTask_Result(t *history.HistoryService_ScheduleDecisionTask_Result) *types.HistoryService_ScheduleDecisionTask_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_ScheduleDecisionTask_Result{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryService_SignalWithStartWorkflowExecution_Args converts internal HistoryService_SignalWithStartWorkflowExecution_Args type to thrift
func FromHistoryService_SignalWithStartWorkflowExecution_Args(t *types.HistoryService_SignalWithStartWorkflowExecution_Args) *history.HistoryService_SignalWithStartWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_SignalWithStartWorkflowExecution_Args{
		SignalWithStartRequest: FromHistorySignalWithStartWorkflowExecutionRequest(t.SignalWithStartRequest),
	}
}

// ToHistoryService_SignalWithStartWorkflowExecution_Args converts thrift HistoryService_SignalWithStartWorkflowExecution_Args type to internal
func ToHistoryService_SignalWithStartWorkflowExecution_Args(t *history.HistoryService_SignalWithStartWorkflowExecution_Args) *types.HistoryService_SignalWithStartWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_SignalWithStartWorkflowExecution_Args{
		SignalWithStartRequest: ToHistorySignalWithStartWorkflowExecutionRequest(t.SignalWithStartRequest),
	}
}

// FromHistoryService_SignalWithStartWorkflowExecution_Result converts internal HistoryService_SignalWithStartWorkflowExecution_Result type to thrift
func FromHistoryService_SignalWithStartWorkflowExecution_Result(t *types.HistoryService_SignalWithStartWorkflowExecution_Result) *history.HistoryService_SignalWithStartWorkflowExecution_Result {
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

// ToHistoryService_SignalWithStartWorkflowExecution_Result converts thrift HistoryService_SignalWithStartWorkflowExecution_Result type to internal
func ToHistoryService_SignalWithStartWorkflowExecution_Result(t *history.HistoryService_SignalWithStartWorkflowExecution_Result) *types.HistoryService_SignalWithStartWorkflowExecution_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_SignalWithStartWorkflowExecution_Result{
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

// FromHistoryService_SignalWorkflowExecution_Args converts internal HistoryService_SignalWorkflowExecution_Args type to thrift
func FromHistoryService_SignalWorkflowExecution_Args(t *types.HistoryService_SignalWorkflowExecution_Args) *history.HistoryService_SignalWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_SignalWorkflowExecution_Args{
		SignalRequest: FromHistorySignalWorkflowExecutionRequest(t.SignalRequest),
	}
}

// ToHistoryService_SignalWorkflowExecution_Args converts thrift HistoryService_SignalWorkflowExecution_Args type to internal
func ToHistoryService_SignalWorkflowExecution_Args(t *history.HistoryService_SignalWorkflowExecution_Args) *types.HistoryService_SignalWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_SignalWorkflowExecution_Args{
		SignalRequest: ToHistorySignalWorkflowExecutionRequest(t.SignalRequest),
	}
}

// FromHistoryService_SignalWorkflowExecution_Result converts internal HistoryService_SignalWorkflowExecution_Result type to thrift
func FromHistoryService_SignalWorkflowExecution_Result(t *types.HistoryService_SignalWorkflowExecution_Result) *history.HistoryService_SignalWorkflowExecution_Result {
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

// ToHistoryService_SignalWorkflowExecution_Result converts thrift HistoryService_SignalWorkflowExecution_Result type to internal
func ToHistoryService_SignalWorkflowExecution_Result(t *history.HistoryService_SignalWorkflowExecution_Result) *types.HistoryService_SignalWorkflowExecution_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_SignalWorkflowExecution_Result{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		DomainNotActiveError:    ToDomainNotActiveError(t.DomainNotActiveError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
	}
}

// FromHistoryService_StartWorkflowExecution_Args converts internal HistoryService_StartWorkflowExecution_Args type to thrift
func FromHistoryService_StartWorkflowExecution_Args(t *types.HistoryService_StartWorkflowExecution_Args) *history.HistoryService_StartWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_StartWorkflowExecution_Args{
		StartRequest: FromHistoryStartWorkflowExecutionRequest(t.StartRequest),
	}
}

// ToHistoryService_StartWorkflowExecution_Args converts thrift HistoryService_StartWorkflowExecution_Args type to internal
func ToHistoryService_StartWorkflowExecution_Args(t *history.HistoryService_StartWorkflowExecution_Args) *types.HistoryService_StartWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_StartWorkflowExecution_Args{
		StartRequest: ToHistoryStartWorkflowExecutionRequest(t.StartRequest),
	}
}

// FromHistoryService_StartWorkflowExecution_Result converts internal HistoryService_StartWorkflowExecution_Result type to thrift
func FromHistoryService_StartWorkflowExecution_Result(t *types.HistoryService_StartWorkflowExecution_Result) *history.HistoryService_StartWorkflowExecution_Result {
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

// ToHistoryService_StartWorkflowExecution_Result converts thrift HistoryService_StartWorkflowExecution_Result type to internal
func ToHistoryService_StartWorkflowExecution_Result(t *history.HistoryService_StartWorkflowExecution_Result) *types.HistoryService_StartWorkflowExecution_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_StartWorkflowExecution_Result{
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

// FromHistoryService_SyncActivity_Args converts internal HistoryService_SyncActivity_Args type to thrift
func FromHistoryService_SyncActivity_Args(t *types.HistoryService_SyncActivity_Args) *history.HistoryService_SyncActivity_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_SyncActivity_Args{
		SyncActivityRequest: FromSyncActivityRequest(t.SyncActivityRequest),
	}
}

// ToHistoryService_SyncActivity_Args converts thrift HistoryService_SyncActivity_Args type to internal
func ToHistoryService_SyncActivity_Args(t *history.HistoryService_SyncActivity_Args) *types.HistoryService_SyncActivity_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_SyncActivity_Args{
		SyncActivityRequest: ToSyncActivityRequest(t.SyncActivityRequest),
	}
}

// FromHistoryService_SyncActivity_Result converts internal HistoryService_SyncActivity_Result type to thrift
func FromHistoryService_SyncActivity_Result(t *types.HistoryService_SyncActivity_Result) *history.HistoryService_SyncActivity_Result {
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

// ToHistoryService_SyncActivity_Result converts thrift HistoryService_SyncActivity_Result type to internal
func ToHistoryService_SyncActivity_Result(t *history.HistoryService_SyncActivity_Result) *types.HistoryService_SyncActivity_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_SyncActivity_Result{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:     ToEntityNotExistsError(t.EntityNotExistError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
		RetryTaskV2Error:        ToRetryTaskV2Error(t.RetryTaskV2Error),
	}
}

// FromHistoryService_SyncShardStatus_Args converts internal HistoryService_SyncShardStatus_Args type to thrift
func FromHistoryService_SyncShardStatus_Args(t *types.HistoryService_SyncShardStatus_Args) *history.HistoryService_SyncShardStatus_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_SyncShardStatus_Args{
		SyncShardStatusRequest: FromSyncShardStatusRequest(t.SyncShardStatusRequest),
	}
}

// ToHistoryService_SyncShardStatus_Args converts thrift HistoryService_SyncShardStatus_Args type to internal
func ToHistoryService_SyncShardStatus_Args(t *history.HistoryService_SyncShardStatus_Args) *types.HistoryService_SyncShardStatus_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_SyncShardStatus_Args{
		SyncShardStatusRequest: ToSyncShardStatusRequest(t.SyncShardStatusRequest),
	}
}

// FromHistoryService_SyncShardStatus_Result converts internal HistoryService_SyncShardStatus_Result type to thrift
func FromHistoryService_SyncShardStatus_Result(t *types.HistoryService_SyncShardStatus_Result) *history.HistoryService_SyncShardStatus_Result {
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

// ToHistoryService_SyncShardStatus_Result converts thrift HistoryService_SyncShardStatus_Result type to internal
func ToHistoryService_SyncShardStatus_Result(t *history.HistoryService_SyncShardStatus_Result) *types.HistoryService_SyncShardStatus_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_SyncShardStatus_Result{
		BadRequestError:         ToBadRequestError(t.BadRequestError),
		InternalServiceError:    ToInternalServiceError(t.InternalServiceError),
		ShardOwnershipLostError: ToShardOwnershipLostError(t.ShardOwnershipLostError),
		LimitExceededError:      ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:        ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromHistoryService_TerminateWorkflowExecution_Args converts internal HistoryService_TerminateWorkflowExecution_Args type to thrift
func FromHistoryService_TerminateWorkflowExecution_Args(t *types.HistoryService_TerminateWorkflowExecution_Args) *history.HistoryService_TerminateWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &history.HistoryService_TerminateWorkflowExecution_Args{
		TerminateRequest: FromHistoryTerminateWorkflowExecutionRequest(t.TerminateRequest),
	}
}

// ToHistoryService_TerminateWorkflowExecution_Args converts thrift HistoryService_TerminateWorkflowExecution_Args type to internal
func ToHistoryService_TerminateWorkflowExecution_Args(t *history.HistoryService_TerminateWorkflowExecution_Args) *types.HistoryService_TerminateWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &types.HistoryService_TerminateWorkflowExecution_Args{
		TerminateRequest: ToHistoryTerminateWorkflowExecutionRequest(t.TerminateRequest),
	}
}

// FromHistoryService_TerminateWorkflowExecution_Result converts internal HistoryService_TerminateWorkflowExecution_Result type to thrift
func FromHistoryService_TerminateWorkflowExecution_Result(t *types.HistoryService_TerminateWorkflowExecution_Result) *history.HistoryService_TerminateWorkflowExecution_Result {
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

// ToHistoryService_TerminateWorkflowExecution_Result converts thrift HistoryService_TerminateWorkflowExecution_Result type to internal
func ToHistoryService_TerminateWorkflowExecution_Result(t *history.HistoryService_TerminateWorkflowExecution_Result) *types.HistoryService_TerminateWorkflowExecution_Result {
	if t == nil {
		return nil
	}
	return &types.HistoryService_TerminateWorkflowExecution_Result{
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
