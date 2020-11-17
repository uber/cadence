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

	"github.com/uber/cadence/.gen/go/admin"
)

// FromAddSearchAttributeRequest converts internal AddSearchAttributeRequest type to thrift
func FromAddSearchAttributeRequest(t *types.AddSearchAttributeRequest) *admin.AddSearchAttributeRequest {
	if t == nil {
		return nil
	}
	return &admin.AddSearchAttributeRequest{
		SearchAttribute: FromIndexedValueTypeMap(t.SearchAttribute),
		SecurityToken:   t.SecurityToken,
	}
}

// ToAddSearchAttributeRequest converts thrift AddSearchAttributeRequest type to internal
func ToAddSearchAttributeRequest(t *admin.AddSearchAttributeRequest) *types.AddSearchAttributeRequest {
	if t == nil {
		return nil
	}
	return &types.AddSearchAttributeRequest{
		SearchAttribute: ToIndexedValueTypeMap(t.SearchAttribute),
		SecurityToken:   t.SecurityToken,
	}
}

// FromAdminServiceAddSearchAttributeArgs converts internal AdminService_AddSearchAttribute_Args type to thrift
func FromAdminServiceAddSearchAttributeArgs(t *types.AdminServiceAddSearchAttributeArgs) *admin.AdminService_AddSearchAttribute_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_AddSearchAttribute_Args{
		Request: FromAddSearchAttributeRequest(t.Request),
	}
}

// ToAdminServiceAddSearchAttributeArgs converts thrift AdminService_AddSearchAttribute_Args type to internal
func ToAdminServiceAddSearchAttributeArgs(t *admin.AdminService_AddSearchAttribute_Args) *types.AdminServiceAddSearchAttributeArgs {
	if t == nil {
		return nil
	}
	return &types.AdminServiceAddSearchAttributeArgs{
		Request: ToAddSearchAttributeRequest(t.Request),
	}
}

// FromAdminServiceAddSearchAttributeResult converts internal AdminService_AddSearchAttribute_Result type to thrift
func FromAdminServiceAddSearchAttributeResult(t *types.AdminServiceAddSearchAttributeResult) *admin.AdminService_AddSearchAttribute_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_AddSearchAttribute_Result{
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToAdminServiceAddSearchAttributeResult converts thrift AdminService_AddSearchAttribute_Result type to internal
func ToAdminServiceAddSearchAttributeResult(t *admin.AdminService_AddSearchAttribute_Result) *types.AdminServiceAddSearchAttributeResult {
	if t == nil {
		return nil
	}
	return &types.AdminServiceAddSearchAttributeResult{
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromAdminServiceCloseShardArgs converts internal AdminService_CloseShard_Args type to thrift
func FromAdminServiceCloseShardArgs(t *types.AdminServiceCloseShardArgs) *admin.AdminService_CloseShard_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_CloseShard_Args{
		Request: FromCloseShardRequest(t.Request),
	}
}

// ToAdminServiceCloseShardArgs converts thrift AdminService_CloseShard_Args type to internal
func ToAdminServiceCloseShardArgs(t *admin.AdminService_CloseShard_Args) *types.AdminServiceCloseShardArgs {
	if t == nil {
		return nil
	}
	return &types.AdminServiceCloseShardArgs{
		Request: ToCloseShardRequest(t.Request),
	}
}

// FromAdminServiceCloseShardResult converts internal AdminService_CloseShard_Result type to thrift
func FromAdminServiceCloseShardResult(t *types.AdminServiceCloseShardResult) *admin.AdminService_CloseShard_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_CloseShard_Result{
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    FromAccessDeniedError(t.AccessDeniedError),
	}
}

// ToAdminServiceCloseShardResult converts thrift AdminService_CloseShard_Result type to internal
func ToAdminServiceCloseShardResult(t *admin.AdminService_CloseShard_Result) *types.AdminServiceCloseShardResult {
	if t == nil {
		return nil
	}
	return &types.AdminServiceCloseShardResult{
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    ToAccessDeniedError(t.AccessDeniedError),
	}
}

// FromAdminServiceDescribeClusterArgs converts internal AdminService_DescribeCluster_Args type to thrift
func FromAdminServiceDescribeClusterArgs(t *types.AdminServiceDescribeClusterArgs) *admin.AdminService_DescribeCluster_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_DescribeCluster_Args{}
}

// ToAdminServiceDescribeClusterArgs converts thrift AdminService_DescribeCluster_Args type to internal
func ToAdminServiceDescribeClusterArgs(t *admin.AdminService_DescribeCluster_Args) *types.AdminServiceDescribeClusterArgs {
	if t == nil {
		return nil
	}
	return &types.AdminServiceDescribeClusterArgs{}
}

// FromAdminServiceDescribeClusterResult converts internal AdminService_DescribeCluster_Result type to thrift
func FromAdminServiceDescribeClusterResult(t *types.AdminServiceDescribeClusterResult) *admin.AdminService_DescribeCluster_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_DescribeCluster_Result{
		Success:              FromDescribeClusterResponse(t.Success),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToAdminServiceDescribeClusterResult converts thrift AdminService_DescribeCluster_Result type to internal
func ToAdminServiceDescribeClusterResult(t *admin.AdminService_DescribeCluster_Result) *types.AdminServiceDescribeClusterResult {
	if t == nil {
		return nil
	}
	return &types.AdminServiceDescribeClusterResult{
		Success:              ToDescribeClusterResponse(t.Success),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromAdminServiceDescribeHistoryHostArgs converts internal AdminService_DescribeHistoryHost_Args type to thrift
func FromAdminServiceDescribeHistoryHostArgs(t *types.AdminServiceDescribeHistoryHostArgs) *admin.AdminService_DescribeHistoryHost_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_DescribeHistoryHost_Args{
		Request: FromDescribeHistoryHostRequest(t.Request),
	}
}

// ToAdminServiceDescribeHistoryHostArgs converts thrift AdminService_DescribeHistoryHost_Args type to internal
func ToAdminServiceDescribeHistoryHostArgs(t *admin.AdminService_DescribeHistoryHost_Args) *types.AdminServiceDescribeHistoryHostArgs {
	if t == nil {
		return nil
	}
	return &types.AdminServiceDescribeHistoryHostArgs{
		Request: ToDescribeHistoryHostRequest(t.Request),
	}
}

// FromAdminServiceDescribeHistoryHostResult converts internal AdminService_DescribeHistoryHost_Result type to thrift
func FromAdminServiceDescribeHistoryHostResult(t *types.AdminServiceDescribeHistoryHostResult) *admin.AdminService_DescribeHistoryHost_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_DescribeHistoryHost_Result{
		Success:              FromDescribeHistoryHostResponse(t.Success),
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    FromAccessDeniedError(t.AccessDeniedError),
	}
}

// ToAdminServiceDescribeHistoryHostResult converts thrift AdminService_DescribeHistoryHost_Result type to internal
func ToAdminServiceDescribeHistoryHostResult(t *admin.AdminService_DescribeHistoryHost_Result) *types.AdminServiceDescribeHistoryHostResult {
	if t == nil {
		return nil
	}
	return &types.AdminServiceDescribeHistoryHostResult{
		Success:              ToDescribeHistoryHostResponse(t.Success),
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    ToAccessDeniedError(t.AccessDeniedError),
	}
}

// FromAdminServiceDescribeQueueArgs converts internal AdminService_DescribeQueue_Args type to thrift
func FromAdminServiceDescribeQueueArgs(t *types.AdminServiceDescribeQueueArgs) *admin.AdminService_DescribeQueue_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_DescribeQueue_Args{
		Request: FromDescribeQueueRequest(t.Request),
	}
}

// ToAdminServiceDescribeQueueArgs converts thrift AdminService_DescribeQueue_Args type to internal
func ToAdminServiceDescribeQueueArgs(t *admin.AdminService_DescribeQueue_Args) *types.AdminServiceDescribeQueueArgs {
	if t == nil {
		return nil
	}
	return &types.AdminServiceDescribeQueueArgs{
		Request: ToDescribeQueueRequest(t.Request),
	}
}

// FromAdminServiceDescribeQueueResult converts internal AdminService_DescribeQueue_Result type to thrift
func FromAdminServiceDescribeQueueResult(t *types.AdminServiceDescribeQueueResult) *admin.AdminService_DescribeQueue_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_DescribeQueue_Result{
		Success:              FromDescribeQueueResponse(t.Success),
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    FromAccessDeniedError(t.AccessDeniedError),
	}
}

// ToAdminServiceDescribeQueueResult converts thrift AdminService_DescribeQueue_Result type to internal
func ToAdminServiceDescribeQueueResult(t *admin.AdminService_DescribeQueue_Result) *types.AdminServiceDescribeQueueResult {
	if t == nil {
		return nil
	}
	return &types.AdminServiceDescribeQueueResult{
		Success:              ToDescribeQueueResponse(t.Success),
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    ToAccessDeniedError(t.AccessDeniedError),
	}
}

// FromAdminServiceDescribeWorkflowExecutionArgs converts internal AdminService_DescribeWorkflowExecution_Args type to thrift
func FromAdminServiceDescribeWorkflowExecutionArgs(t *types.AdminServiceDescribeWorkflowExecutionArgs) *admin.AdminService_DescribeWorkflowExecution_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_DescribeWorkflowExecution_Args{
		Request: FromAdminDescribeWorkflowExecutionRequest(t.Request),
	}
}

// ToAdminServiceDescribeWorkflowExecutionArgs converts thrift AdminService_DescribeWorkflowExecution_Args type to internal
func ToAdminServiceDescribeWorkflowExecutionArgs(t *admin.AdminService_DescribeWorkflowExecution_Args) *types.AdminServiceDescribeWorkflowExecutionArgs {
	if t == nil {
		return nil
	}
	return &types.AdminServiceDescribeWorkflowExecutionArgs{
		Request: ToAdminDescribeWorkflowExecutionRequest(t.Request),
	}
}

// FromAdminServiceDescribeWorkflowExecutionResult converts internal AdminService_DescribeWorkflowExecution_Result type to thrift
func FromAdminServiceDescribeWorkflowExecutionResult(t *types.AdminServiceDescribeWorkflowExecutionResult) *admin.AdminService_DescribeWorkflowExecution_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_DescribeWorkflowExecution_Result{
		Success:              FromAdminDescribeWorkflowExecutionResponse(t.Success),
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:  FromEntityNotExistsError(t.EntityNotExistError),
		AccessDeniedError:    FromAccessDeniedError(t.AccessDeniedError),
	}
}

// ToAdminServiceDescribeWorkflowExecutionResult converts thrift AdminService_DescribeWorkflowExecution_Result type to internal
func ToAdminServiceDescribeWorkflowExecutionResult(t *admin.AdminService_DescribeWorkflowExecution_Result) *types.AdminServiceDescribeWorkflowExecutionResult {
	if t == nil {
		return nil
	}
	return &types.AdminServiceDescribeWorkflowExecutionResult{
		Success:              ToAdminDescribeWorkflowExecutionResponse(t.Success),
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:  ToEntityNotExistsError(t.EntityNotExistError),
		AccessDeniedError:    ToAccessDeniedError(t.AccessDeniedError),
	}
}

// FromAdminServiceGetDLQReplicationMessagesArgs converts internal AdminService_GetDLQReplicationMessages_Args type to thrift
func FromAdminServiceGetDLQReplicationMessagesArgs(t *types.AdminServiceGetDLQReplicationMessagesArgs) *admin.AdminService_GetDLQReplicationMessages_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_GetDLQReplicationMessages_Args{
		Request: FromGetDLQReplicationMessagesRequest(t.Request),
	}
}

// ToAdminServiceGetDLQReplicationMessagesArgs converts thrift AdminService_GetDLQReplicationMessages_Args type to internal
func ToAdminServiceGetDLQReplicationMessagesArgs(t *admin.AdminService_GetDLQReplicationMessages_Args) *types.AdminServiceGetDLQReplicationMessagesArgs {
	if t == nil {
		return nil
	}
	return &types.AdminServiceGetDLQReplicationMessagesArgs{
		Request: ToGetDLQReplicationMessagesRequest(t.Request),
	}
}

// FromAdminServiceGetDLQReplicationMessagesResult converts internal AdminService_GetDLQReplicationMessages_Result type to thrift
func FromAdminServiceGetDLQReplicationMessagesResult(t *types.AdminServiceGetDLQReplicationMessagesResult) *admin.AdminService_GetDLQReplicationMessages_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_GetDLQReplicationMessages_Result{
		Success:          FromGetDLQReplicationMessagesResponse(t.Success),
		BadRequestError:  FromBadRequestError(t.BadRequestError),
		ServiceBusyError: FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToAdminServiceGetDLQReplicationMessagesResult converts thrift AdminService_GetDLQReplicationMessages_Result type to internal
func ToAdminServiceGetDLQReplicationMessagesResult(t *admin.AdminService_GetDLQReplicationMessages_Result) *types.AdminServiceGetDLQReplicationMessagesResult {
	if t == nil {
		return nil
	}
	return &types.AdminServiceGetDLQReplicationMessagesResult{
		Success:          ToGetDLQReplicationMessagesResponse(t.Success),
		BadRequestError:  ToBadRequestError(t.BadRequestError),
		ServiceBusyError: ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromAdminServiceGetDomainReplicationMessagesArgs converts internal AdminService_GetDomainReplicationMessages_Args type to thrift
func FromAdminServiceGetDomainReplicationMessagesArgs(t *types.AdminServiceGetDomainReplicationMessagesArgs) *admin.AdminService_GetDomainReplicationMessages_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_GetDomainReplicationMessages_Args{
		Request: FromGetDomainReplicationMessagesRequest(t.Request),
	}
}

// ToAdminServiceGetDomainReplicationMessagesArgs converts thrift AdminService_GetDomainReplicationMessages_Args type to internal
func ToAdminServiceGetDomainReplicationMessagesArgs(t *admin.AdminService_GetDomainReplicationMessages_Args) *types.AdminServiceGetDomainReplicationMessagesArgs {
	if t == nil {
		return nil
	}
	return &types.AdminServiceGetDomainReplicationMessagesArgs{
		Request: ToGetDomainReplicationMessagesRequest(t.Request),
	}
}

// FromAdminServiceGetDomainReplicationMessagesResult converts internal AdminService_GetDomainReplicationMessages_Result type to thrift
func FromAdminServiceGetDomainReplicationMessagesResult(t *types.AdminServiceGetDomainReplicationMessagesResult) *admin.AdminService_GetDomainReplicationMessages_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_GetDomainReplicationMessages_Result{
		Success:                        FromGetDomainReplicationMessagesResponse(t.Success),
		BadRequestError:                FromBadRequestError(t.BadRequestError),
		LimitExceededError:             FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:               FromServiceBusyError(t.ServiceBusyError),
		ClientVersionNotSupportedError: FromClientVersionNotSupportedError(t.ClientVersionNotSupportedError),
	}
}

// ToAdminServiceGetDomainReplicationMessagesResult converts thrift AdminService_GetDomainReplicationMessages_Result type to internal
func ToAdminServiceGetDomainReplicationMessagesResult(t *admin.AdminService_GetDomainReplicationMessages_Result) *types.AdminServiceGetDomainReplicationMessagesResult {
	if t == nil {
		return nil
	}
	return &types.AdminServiceGetDomainReplicationMessagesResult{
		Success:                        ToGetDomainReplicationMessagesResponse(t.Success),
		BadRequestError:                ToBadRequestError(t.BadRequestError),
		LimitExceededError:             ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:               ToServiceBusyError(t.ServiceBusyError),
		ClientVersionNotSupportedError: ToClientVersionNotSupportedError(t.ClientVersionNotSupportedError),
	}
}

// FromAdminServiceGetReplicationMessagesArgs converts internal AdminService_GetReplicationMessages_Args type to thrift
func FromAdminServiceGetReplicationMessagesArgs(t *types.AdminServiceGetReplicationMessagesArgs) *admin.AdminService_GetReplicationMessages_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_GetReplicationMessages_Args{
		Request: FromGetReplicationMessagesRequest(t.Request),
	}
}

// ToAdminServiceGetReplicationMessagesArgs converts thrift AdminService_GetReplicationMessages_Args type to internal
func ToAdminServiceGetReplicationMessagesArgs(t *admin.AdminService_GetReplicationMessages_Args) *types.AdminServiceGetReplicationMessagesArgs {
	if t == nil {
		return nil
	}
	return &types.AdminServiceGetReplicationMessagesArgs{
		Request: ToGetReplicationMessagesRequest(t.Request),
	}
}

// FromAdminServiceGetReplicationMessagesResult converts internal AdminService_GetReplicationMessages_Result type to thrift
func FromAdminServiceGetReplicationMessagesResult(t *types.AdminServiceGetReplicationMessagesResult) *admin.AdminService_GetReplicationMessages_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_GetReplicationMessages_Result{
		Success:                        FromGetReplicationMessagesResponse(t.Success),
		BadRequestError:                FromBadRequestError(t.BadRequestError),
		LimitExceededError:             FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:               FromServiceBusyError(t.ServiceBusyError),
		ClientVersionNotSupportedError: FromClientVersionNotSupportedError(t.ClientVersionNotSupportedError),
	}
}

// ToAdminServiceGetReplicationMessagesResult converts thrift AdminService_GetReplicationMessages_Result type to internal
func ToAdminServiceGetReplicationMessagesResult(t *admin.AdminService_GetReplicationMessages_Result) *types.AdminServiceGetReplicationMessagesResult {
	if t == nil {
		return nil
	}
	return &types.AdminServiceGetReplicationMessagesResult{
		Success:                        ToGetReplicationMessagesResponse(t.Success),
		BadRequestError:                ToBadRequestError(t.BadRequestError),
		LimitExceededError:             ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:               ToServiceBusyError(t.ServiceBusyError),
		ClientVersionNotSupportedError: ToClientVersionNotSupportedError(t.ClientVersionNotSupportedError),
	}
}

// FromAdminServiceGetWorkflowExecutionRawHistoryV2Args converts internal AdminService_GetWorkflowExecutionRawHistoryV2_Args type to thrift
func FromAdminServiceGetWorkflowExecutionRawHistoryV2Args(t *types.AdminServiceGetWorkflowExecutionRawHistoryV2Args) *admin.AdminService_GetWorkflowExecutionRawHistoryV2_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_GetWorkflowExecutionRawHistoryV2_Args{
		GetRequest: FromGetWorkflowExecutionRawHistoryV2Request(t.GetRequest),
	}
}

// ToAdminServiceGetWorkflowExecutionRawHistoryV2Args converts thrift AdminService_GetWorkflowExecutionRawHistoryV2_Args type to internal
func ToAdminServiceGetWorkflowExecutionRawHistoryV2Args(t *admin.AdminService_GetWorkflowExecutionRawHistoryV2_Args) *types.AdminServiceGetWorkflowExecutionRawHistoryV2Args {
	if t == nil {
		return nil
	}
	return &types.AdminServiceGetWorkflowExecutionRawHistoryV2Args{
		GetRequest: ToGetWorkflowExecutionRawHistoryV2Request(t.GetRequest),
	}
}

// FromAdminServiceGetWorkflowExecutionRawHistoryV2Result converts internal AdminService_GetWorkflowExecutionRawHistoryV2_Result type to thrift
func FromAdminServiceGetWorkflowExecutionRawHistoryV2Result(t *types.AdminServiceGetWorkflowExecutionRawHistoryV2Result) *admin.AdminService_GetWorkflowExecutionRawHistoryV2_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_GetWorkflowExecutionRawHistoryV2_Result{
		Success:              FromGetWorkflowExecutionRawHistoryV2Response(t.Success),
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		EntityNotExistError:  FromEntityNotExistsError(t.EntityNotExistError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
	}
}

// ToAdminServiceGetWorkflowExecutionRawHistoryV2Result converts thrift AdminService_GetWorkflowExecutionRawHistoryV2_Result type to internal
func ToAdminServiceGetWorkflowExecutionRawHistoryV2Result(t *admin.AdminService_GetWorkflowExecutionRawHistoryV2_Result) *types.AdminServiceGetWorkflowExecutionRawHistoryV2Result {
	if t == nil {
		return nil
	}
	return &types.AdminServiceGetWorkflowExecutionRawHistoryV2Result{
		Success:              ToGetWorkflowExecutionRawHistoryV2Response(t.Success),
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		EntityNotExistError:  ToEntityNotExistsError(t.EntityNotExistError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
	}
}

// FromAdminServiceMergeDLQMessagesArgs converts internal AdminService_MergeDLQMessages_Args type to thrift
func FromAdminServiceMergeDLQMessagesArgs(t *types.AdminServiceMergeDLQMessagesArgs) *admin.AdminService_MergeDLQMessages_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_MergeDLQMessages_Args{
		Request: FromMergeDLQMessagesRequest(t.Request),
	}
}

// ToAdminServiceMergeDLQMessagesArgs converts thrift AdminService_MergeDLQMessages_Args type to internal
func ToAdminServiceMergeDLQMessagesArgs(t *admin.AdminService_MergeDLQMessages_Args) *types.AdminServiceMergeDLQMessagesArgs {
	if t == nil {
		return nil
	}
	return &types.AdminServiceMergeDLQMessagesArgs{
		Request: ToMergeDLQMessagesRequest(t.Request),
	}
}

// FromAdminServiceMergeDLQMessagesResult converts internal AdminService_MergeDLQMessages_Result type to thrift
func FromAdminServiceMergeDLQMessagesResult(t *types.AdminServiceMergeDLQMessagesResult) *admin.AdminService_MergeDLQMessages_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_MergeDLQMessages_Result{
		Success:              FromMergeDLQMessagesResponse(t.Success),
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:  FromEntityNotExistsError(t.EntityNotExistError),
	}
}

// ToAdminServiceMergeDLQMessagesResult converts thrift AdminService_MergeDLQMessages_Result type to internal
func ToAdminServiceMergeDLQMessagesResult(t *admin.AdminService_MergeDLQMessages_Result) *types.AdminServiceMergeDLQMessagesResult {
	if t == nil {
		return nil
	}
	return &types.AdminServiceMergeDLQMessagesResult{
		Success:              ToMergeDLQMessagesResponse(t.Success),
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:  ToEntityNotExistsError(t.EntityNotExistError),
	}
}

// FromAdminServicePurgeDLQMessagesArgs converts internal AdminService_PurgeDLQMessages_Args type to thrift
func FromAdminServicePurgeDLQMessagesArgs(t *types.AdminServicePurgeDLQMessagesArgs) *admin.AdminService_PurgeDLQMessages_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_PurgeDLQMessages_Args{
		Request: FromPurgeDLQMessagesRequest(t.Request),
	}
}

// ToAdminServicePurgeDLQMessagesArgs converts thrift AdminService_PurgeDLQMessages_Args type to internal
func ToAdminServicePurgeDLQMessagesArgs(t *admin.AdminService_PurgeDLQMessages_Args) *types.AdminServicePurgeDLQMessagesArgs {
	if t == nil {
		return nil
	}
	return &types.AdminServicePurgeDLQMessagesArgs{
		Request: ToPurgeDLQMessagesRequest(t.Request),
	}
}

// FromAdminServicePurgeDLQMessagesResult converts internal AdminService_PurgeDLQMessages_Result type to thrift
func FromAdminServicePurgeDLQMessagesResult(t *types.AdminServicePurgeDLQMessagesResult) *admin.AdminService_PurgeDLQMessages_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_PurgeDLQMessages_Result{
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:  FromEntityNotExistsError(t.EntityNotExistError),
	}
}

// ToAdminServicePurgeDLQMessagesResult converts thrift AdminService_PurgeDLQMessages_Result type to internal
func ToAdminServicePurgeDLQMessagesResult(t *admin.AdminService_PurgeDLQMessages_Result) *types.AdminServicePurgeDLQMessagesResult {
	if t == nil {
		return nil
	}
	return &types.AdminServicePurgeDLQMessagesResult{
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:  ToEntityNotExistsError(t.EntityNotExistError),
	}
}

// FromAdminServiceReadDLQMessagesArgs converts internal AdminService_ReadDLQMessages_Args type to thrift
func FromAdminServiceReadDLQMessagesArgs(t *types.AdminServiceReadDLQMessagesArgs) *admin.AdminService_ReadDLQMessages_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_ReadDLQMessages_Args{
		Request: FromReadDLQMessagesRequest(t.Request),
	}
}

// ToAdminServiceReadDLQMessagesArgs converts thrift AdminService_ReadDLQMessages_Args type to internal
func ToAdminServiceReadDLQMessagesArgs(t *admin.AdminService_ReadDLQMessages_Args) *types.AdminServiceReadDLQMessagesArgs {
	if t == nil {
		return nil
	}
	return &types.AdminServiceReadDLQMessagesArgs{
		Request: ToReadDLQMessagesRequest(t.Request),
	}
}

// FromAdminServiceReadDLQMessagesResult converts internal AdminService_ReadDLQMessages_Result type to thrift
func FromAdminServiceReadDLQMessagesResult(t *types.AdminServiceReadDLQMessagesResult) *admin.AdminService_ReadDLQMessages_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_ReadDLQMessages_Result{
		Success:              FromReadDLQMessagesResponse(t.Success),
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:  FromEntityNotExistsError(t.EntityNotExistError),
	}
}

// ToAdminServiceReadDLQMessagesResult converts thrift AdminService_ReadDLQMessages_Result type to internal
func ToAdminServiceReadDLQMessagesResult(t *admin.AdminService_ReadDLQMessages_Result) *types.AdminServiceReadDLQMessagesResult {
	if t == nil {
		return nil
	}
	return &types.AdminServiceReadDLQMessagesResult{
		Success:              ToReadDLQMessagesResponse(t.Success),
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:  ToEntityNotExistsError(t.EntityNotExistError),
	}
}

// FromAdminServiceReapplyEventsArgs converts internal AdminService_ReapplyEvents_Args type to thrift
func FromAdminServiceReapplyEventsArgs(t *types.AdminServiceReapplyEventsArgs) *admin.AdminService_ReapplyEvents_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_ReapplyEvents_Args{
		ReapplyEventsRequest: FromReapplyEventsRequest(t.ReapplyEventsRequest),
	}
}

// ToAdminServiceReapplyEventsArgs converts thrift AdminService_ReapplyEvents_Args type to internal
func ToAdminServiceReapplyEventsArgs(t *admin.AdminService_ReapplyEvents_Args) *types.AdminServiceReapplyEventsArgs {
	if t == nil {
		return nil
	}
	return &types.AdminServiceReapplyEventsArgs{
		ReapplyEventsRequest: ToReapplyEventsRequest(t.ReapplyEventsRequest),
	}
}

// FromAdminServiceReapplyEventsResult converts internal AdminService_ReapplyEvents_Result type to thrift
func FromAdminServiceReapplyEventsResult(t *types.AdminServiceReapplyEventsResult) *admin.AdminService_ReapplyEvents_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_ReapplyEvents_Result{
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		DomainNotActiveError: FromDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:   FromLimitExceededError(t.LimitExceededError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:  FromEntityNotExistsError(t.EntityNotExistError),
	}
}

// ToAdminServiceReapplyEventsResult converts thrift AdminService_ReapplyEvents_Result type to internal
func ToAdminServiceReapplyEventsResult(t *admin.AdminService_ReapplyEvents_Result) *types.AdminServiceReapplyEventsResult {
	if t == nil {
		return nil
	}
	return &types.AdminServiceReapplyEventsResult{
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		DomainNotActiveError: ToDomainNotActiveError(t.DomainNotActiveError),
		LimitExceededError:   ToLimitExceededError(t.LimitExceededError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:  ToEntityNotExistsError(t.EntityNotExistError),
	}
}

// FromAdminServiceRefreshWorkflowTasksArgs converts internal AdminService_RefreshWorkflowTasks_Args type to thrift
func FromAdminServiceRefreshWorkflowTasksArgs(t *types.AdminServiceRefreshWorkflowTasksArgs) *admin.AdminService_RefreshWorkflowTasks_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_RefreshWorkflowTasks_Args{
		Request: FromRefreshWorkflowTasksRequest(t.Request),
	}
}

// ToAdminServiceRefreshWorkflowTasksArgs converts thrift AdminService_RefreshWorkflowTasks_Args type to internal
func ToAdminServiceRefreshWorkflowTasksArgs(t *admin.AdminService_RefreshWorkflowTasks_Args) *types.AdminServiceRefreshWorkflowTasksArgs {
	if t == nil {
		return nil
	}
	return &types.AdminServiceRefreshWorkflowTasksArgs{
		Request: ToRefreshWorkflowTasksRequest(t.Request),
	}
}

// FromAdminServiceRefreshWorkflowTasksResult converts internal AdminService_RefreshWorkflowTasks_Result type to thrift
func FromAdminServiceRefreshWorkflowTasksResult(t *types.AdminServiceRefreshWorkflowTasksResult) *admin.AdminService_RefreshWorkflowTasks_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_RefreshWorkflowTasks_Result{
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		DomainNotActiveError: FromDomainNotActiveError(t.DomainNotActiveError),
		ServiceBusyError:     FromServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:  FromEntityNotExistsError(t.EntityNotExistError),
	}
}

// ToAdminServiceRefreshWorkflowTasksResult converts thrift AdminService_RefreshWorkflowTasks_Result type to internal
func ToAdminServiceRefreshWorkflowTasksResult(t *admin.AdminService_RefreshWorkflowTasks_Result) *types.AdminServiceRefreshWorkflowTasksResult {
	if t == nil {
		return nil
	}
	return &types.AdminServiceRefreshWorkflowTasksResult{
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		DomainNotActiveError: ToDomainNotActiveError(t.DomainNotActiveError),
		ServiceBusyError:     ToServiceBusyError(t.ServiceBusyError),
		EntityNotExistError:  ToEntityNotExistsError(t.EntityNotExistError),
	}
}

// FromAdminServiceRemoveTaskArgs converts internal AdminService_RemoveTask_Args type to thrift
func FromAdminServiceRemoveTaskArgs(t *types.AdminServiceRemoveTaskArgs) *admin.AdminService_RemoveTask_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_RemoveTask_Args{
		Request: FromRemoveTaskRequest(t.Request),
	}
}

// ToAdminServiceRemoveTaskArgs converts thrift AdminService_RemoveTask_Args type to internal
func ToAdminServiceRemoveTaskArgs(t *admin.AdminService_RemoveTask_Args) *types.AdminServiceRemoveTaskArgs {
	if t == nil {
		return nil
	}
	return &types.AdminServiceRemoveTaskArgs{
		Request: ToRemoveTaskRequest(t.Request),
	}
}

// FromAdminServiceRemoveTaskResult converts internal AdminService_RemoveTask_Result type to thrift
func FromAdminServiceRemoveTaskResult(t *types.AdminServiceRemoveTaskResult) *admin.AdminService_RemoveTask_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_RemoveTask_Result{
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    FromAccessDeniedError(t.AccessDeniedError),
	}
}

// ToAdminServiceRemoveTaskResult converts thrift AdminService_RemoveTask_Result type to internal
func ToAdminServiceRemoveTaskResult(t *admin.AdminService_RemoveTask_Result) *types.AdminServiceRemoveTaskResult {
	if t == nil {
		return nil
	}
	return &types.AdminServiceRemoveTaskResult{
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    ToAccessDeniedError(t.AccessDeniedError),
	}
}

// FromAdminServiceResendReplicationTasksArgs converts internal AdminService_ResendReplicationTasks_Args type to thrift
func FromAdminServiceResendReplicationTasksArgs(t *types.AdminServiceResendReplicationTasksArgs) *admin.AdminService_ResendReplicationTasks_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_ResendReplicationTasks_Args{
		Request: FromResendReplicationTasksRequest(t.Request),
	}
}

// ToAdminServiceResendReplicationTasksArgs converts thrift AdminService_ResendReplicationTasks_Args type to internal
func ToAdminServiceResendReplicationTasksArgs(t *admin.AdminService_ResendReplicationTasks_Args) *types.AdminServiceResendReplicationTasksArgs {
	if t == nil {
		return nil
	}
	return &types.AdminServiceResendReplicationTasksArgs{
		Request: ToResendReplicationTasksRequest(t.Request),
	}
}

// FromAdminServiceResendReplicationTasksResult converts internal AdminService_ResendReplicationTasks_Result type to thrift
func FromAdminServiceResendReplicationTasksResult(t *types.AdminServiceResendReplicationTasksResult) *admin.AdminService_ResendReplicationTasks_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_ResendReplicationTasks_Result{
		BadRequestError:     FromBadRequestError(t.BadRequestError),
		ServiceBusyError:    FromServiceBusyError(t.ServiceBusyError),
		EntityNotExistError: FromEntityNotExistsError(t.EntityNotExistError),
	}
}

// ToAdminServiceResendReplicationTasksResult converts thrift AdminService_ResendReplicationTasks_Result type to internal
func ToAdminServiceResendReplicationTasksResult(t *admin.AdminService_ResendReplicationTasks_Result) *types.AdminServiceResendReplicationTasksResult {
	if t == nil {
		return nil
	}
	return &types.AdminServiceResendReplicationTasksResult{
		BadRequestError:     ToBadRequestError(t.BadRequestError),
		ServiceBusyError:    ToServiceBusyError(t.ServiceBusyError),
		EntityNotExistError: ToEntityNotExistsError(t.EntityNotExistError),
	}
}

// FromAdminServiceResetQueueArgs converts internal AdminService_ResetQueue_Args type to thrift
func FromAdminServiceResetQueueArgs(t *types.AdminServiceResetQueueArgs) *admin.AdminService_ResetQueue_Args {
	if t == nil {
		return nil
	}
	return &admin.AdminService_ResetQueue_Args{
		Request: FromResetQueueRequest(t.Request),
	}
}

// ToAdminServiceResetQueueArgs converts thrift AdminService_ResetQueue_Args type to internal
func ToAdminServiceResetQueueArgs(t *admin.AdminService_ResetQueue_Args) *types.AdminServiceResetQueueArgs {
	if t == nil {
		return nil
	}
	return &types.AdminServiceResetQueueArgs{
		Request: ToResetQueueRequest(t.Request),
	}
}

// FromAdminServiceResetQueueResult converts internal AdminService_ResetQueue_Result type to thrift
func FromAdminServiceResetQueueResult(t *types.AdminServiceResetQueueResult) *admin.AdminService_ResetQueue_Result {
	if t == nil {
		return nil
	}
	return &admin.AdminService_ResetQueue_Result{
		BadRequestError:      FromBadRequestError(t.BadRequestError),
		InternalServiceError: FromInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    FromAccessDeniedError(t.AccessDeniedError),
	}
}

// ToAdminServiceResetQueueResult converts thrift AdminService_ResetQueue_Result type to internal
func ToAdminServiceResetQueueResult(t *admin.AdminService_ResetQueue_Result) *types.AdminServiceResetQueueResult {
	if t == nil {
		return nil
	}
	return &types.AdminServiceResetQueueResult{
		BadRequestError:      ToBadRequestError(t.BadRequestError),
		InternalServiceError: ToInternalServiceError(t.InternalServiceError),
		AccessDeniedError:    ToAccessDeniedError(t.AccessDeniedError),
	}
}

// FromDescribeClusterResponse converts internal DescribeClusterResponse type to thrift
func FromDescribeClusterResponse(t *types.DescribeClusterResponse) *admin.DescribeClusterResponse {
	if t == nil {
		return nil
	}
	return &admin.DescribeClusterResponse{
		SupportedClientVersions: FromSupportedClientVersions(t.SupportedClientVersions),
		MembershipInfo:          FromMembershipInfo(t.MembershipInfo),
	}
}

// ToDescribeClusterResponse converts thrift DescribeClusterResponse type to internal
func ToDescribeClusterResponse(t *admin.DescribeClusterResponse) *types.DescribeClusterResponse {
	if t == nil {
		return nil
	}
	return &types.DescribeClusterResponse{
		SupportedClientVersions: ToSupportedClientVersions(t.SupportedClientVersions),
		MembershipInfo:          ToMembershipInfo(t.MembershipInfo),
	}
}

// FromAdminDescribeWorkflowExecutionRequest converts internal DescribeWorkflowExecutionRequest type to thrift
func FromAdminDescribeWorkflowExecutionRequest(t *types.AdminDescribeWorkflowExecutionRequest) *admin.DescribeWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &admin.DescribeWorkflowExecutionRequest{
		Domain:    t.Domain,
		Execution: FromWorkflowExecution(t.Execution),
	}
}

// ToAdminDescribeWorkflowExecutionRequest converts thrift DescribeWorkflowExecutionRequest type to internal
func ToAdminDescribeWorkflowExecutionRequest(t *admin.DescribeWorkflowExecutionRequest) *types.AdminDescribeWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &types.AdminDescribeWorkflowExecutionRequest{
		Domain:    t.Domain,
		Execution: ToWorkflowExecution(t.Execution),
	}
}

// FromAdminDescribeWorkflowExecutionResponse converts internal DescribeWorkflowExecutionResponse type to thrift
func FromAdminDescribeWorkflowExecutionResponse(t *types.AdminDescribeWorkflowExecutionResponse) *admin.DescribeWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &admin.DescribeWorkflowExecutionResponse{
		ShardId:                t.ShardID,
		HistoryAddr:            t.HistoryAddr,
		MutableStateInCache:    t.MutableStateInCache,
		MutableStateInDatabase: t.MutableStateInDatabase,
	}
}

// ToAdminDescribeWorkflowExecutionResponse converts thrift DescribeWorkflowExecutionResponse type to internal
func ToAdminDescribeWorkflowExecutionResponse(t *admin.DescribeWorkflowExecutionResponse) *types.AdminDescribeWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &types.AdminDescribeWorkflowExecutionResponse{
		ShardID:                t.ShardId,
		HistoryAddr:            t.HistoryAddr,
		MutableStateInCache:    t.MutableStateInCache,
		MutableStateInDatabase: t.MutableStateInDatabase,
	}
}

// FromGetWorkflowExecutionRawHistoryV2Request converts internal GetWorkflowExecutionRawHistoryV2Request type to thrift
func FromGetWorkflowExecutionRawHistoryV2Request(t *types.GetWorkflowExecutionRawHistoryV2Request) *admin.GetWorkflowExecutionRawHistoryV2Request {
	if t == nil {
		return nil
	}
	return &admin.GetWorkflowExecutionRawHistoryV2Request{
		Domain:            t.Domain,
		Execution:         FromWorkflowExecution(t.Execution),
		StartEventId:      t.StartEventID,
		StartEventVersion: t.StartEventVersion,
		EndEventId:        t.EndEventID,
		EndEventVersion:   t.EndEventVersion,
		MaximumPageSize:   t.MaximumPageSize,
		NextPageToken:     t.NextPageToken,
	}
}

// ToGetWorkflowExecutionRawHistoryV2Request converts thrift GetWorkflowExecutionRawHistoryV2Request type to internal
func ToGetWorkflowExecutionRawHistoryV2Request(t *admin.GetWorkflowExecutionRawHistoryV2Request) *types.GetWorkflowExecutionRawHistoryV2Request {
	if t == nil {
		return nil
	}
	return &types.GetWorkflowExecutionRawHistoryV2Request{
		Domain:            t.Domain,
		Execution:         ToWorkflowExecution(t.Execution),
		StartEventID:      t.StartEventId,
		StartEventVersion: t.StartEventVersion,
		EndEventID:        t.EndEventId,
		EndEventVersion:   t.EndEventVersion,
		MaximumPageSize:   t.MaximumPageSize,
		NextPageToken:     t.NextPageToken,
	}
}

// FromGetWorkflowExecutionRawHistoryV2Response converts internal GetWorkflowExecutionRawHistoryV2Response type to thrift
func FromGetWorkflowExecutionRawHistoryV2Response(t *types.GetWorkflowExecutionRawHistoryV2Response) *admin.GetWorkflowExecutionRawHistoryV2Response {
	if t == nil {
		return nil
	}
	return &admin.GetWorkflowExecutionRawHistoryV2Response{
		NextPageToken:  t.NextPageToken,
		HistoryBatches: FromDataBlobArray(t.HistoryBatches),
		VersionHistory: FromVersionHistory(t.VersionHistory),
	}
}

// ToGetWorkflowExecutionRawHistoryV2Response converts thrift GetWorkflowExecutionRawHistoryV2Response type to internal
func ToGetWorkflowExecutionRawHistoryV2Response(t *admin.GetWorkflowExecutionRawHistoryV2Response) *types.GetWorkflowExecutionRawHistoryV2Response {
	if t == nil {
		return nil
	}
	return &types.GetWorkflowExecutionRawHistoryV2Response{
		NextPageToken:  t.NextPageToken,
		HistoryBatches: ToDataBlobArray(t.HistoryBatches),
		VersionHistory: ToVersionHistory(t.VersionHistory),
	}
}

// FromHostInfo converts internal HostInfo type to thrift
func FromHostInfo(t *types.HostInfo) *admin.HostInfo {
	if t == nil {
		return nil
	}
	return &admin.HostInfo{
		Identity: t.Identity,
	}
}

// ToHostInfo converts thrift HostInfo type to internal
func ToHostInfo(t *admin.HostInfo) *types.HostInfo {
	if t == nil {
		return nil
	}
	return &types.HostInfo{
		Identity: t.Identity,
	}
}

// FromMembershipInfo converts internal MembershipInfo type to thrift
func FromMembershipInfo(t *types.MembershipInfo) *admin.MembershipInfo {
	if t == nil {
		return nil
	}
	return &admin.MembershipInfo{
		CurrentHost:      FromHostInfo(t.CurrentHost),
		ReachableMembers: t.ReachableMembers,
		Rings:            FromRingInfoArray(t.Rings),
	}
}

// ToMembershipInfo converts thrift MembershipInfo type to internal
func ToMembershipInfo(t *admin.MembershipInfo) *types.MembershipInfo {
	if t == nil {
		return nil
	}
	return &types.MembershipInfo{
		CurrentHost:      ToHostInfo(t.CurrentHost),
		ReachableMembers: t.ReachableMembers,
		Rings:            ToRingInfoArray(t.Rings),
	}
}

// FromResendReplicationTasksRequest converts internal ResendReplicationTasksRequest type to thrift
func FromResendReplicationTasksRequest(t *types.ResendReplicationTasksRequest) *admin.ResendReplicationTasksRequest {
	if t == nil {
		return nil
	}
	return &admin.ResendReplicationTasksRequest{
		DomainID:      t.DomainID,
		WorkflowID:    t.WorkflowID,
		RunID:         t.RunID,
		RemoteCluster: t.RemoteCluster,
		StartEventID:  t.StartEventID,
		StartVersion:  t.StartVersion,
		EndEventID:    t.EndEventID,
		EndVersion:    t.EndVersion,
	}
}

// ToResendReplicationTasksRequest converts thrift ResendReplicationTasksRequest type to internal
func ToResendReplicationTasksRequest(t *admin.ResendReplicationTasksRequest) *types.ResendReplicationTasksRequest {
	if t == nil {
		return nil
	}
	return &types.ResendReplicationTasksRequest{
		DomainID:      t.DomainID,
		WorkflowID:    t.WorkflowID,
		RunID:         t.RunID,
		RemoteCluster: t.RemoteCluster,
		StartEventID:  t.StartEventID,
		StartVersion:  t.StartVersion,
		EndEventID:    t.EndEventID,
		EndVersion:    t.EndVersion,
	}
}

// FromRingInfo converts internal RingInfo type to thrift
func FromRingInfo(t *types.RingInfo) *admin.RingInfo {
	if t == nil {
		return nil
	}
	return &admin.RingInfo{
		Role:        t.Role,
		MemberCount: t.MemberCount,
		Members:     FromHostInfoArray(t.Members),
	}
}

// ToRingInfo converts thrift RingInfo type to internal
func ToRingInfo(t *admin.RingInfo) *types.RingInfo {
	if t == nil {
		return nil
	}
	return &types.RingInfo{
		Role:        t.Role,
		MemberCount: t.MemberCount,
		Members:     ToHostInfoArray(t.Members),
	}
}

// FromRingInfoArray converts internal RingInfo type array to thrift
func FromRingInfoArray(t []*types.RingInfo) []*admin.RingInfo {
	if t == nil {
		return nil
	}
	v := make([]*admin.RingInfo, len(t))
	for i := range t {
		v[i] = FromRingInfo(t[i])
	}
	return v
}

// ToRingInfoArray converts thrift RingInfo type array to internal
func ToRingInfoArray(t []*admin.RingInfo) []*types.RingInfo {
	if t == nil {
		return nil
	}
	v := make([]*types.RingInfo, len(t))
	for i := range t {
		v[i] = ToRingInfo(t[i])
	}
	return v
}

// FromHostInfoArray converts internal HostInfo type array to thrift
func FromHostInfoArray(t []*types.HostInfo) []*admin.HostInfo {
	if t == nil {
		return nil
	}
	v := make([]*admin.HostInfo, len(t))
	for i := range t {
		v[i] = FromHostInfo(t[i])
	}
	return v
}

// ToHostInfoArray converts thrift HostInfo type array to internal
func ToHostInfoArray(t []*admin.HostInfo) []*types.HostInfo {
	if t == nil {
		return nil
	}
	v := make([]*types.HostInfo, len(t))
	for i := range t {
		v[i] = ToHostInfo(t[i])
	}
	return v
}
