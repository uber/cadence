// Copyright (c) 2021 Uber Technologies Inc.
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

package proto

import (
	adminv1 "github.com/uber/cadence/.gen/proto/admin/v1"
	"github.com/uber/cadence/common/types"
)

func FromAdminAddSearchAttributeRequest(t *types.AddSearchAttributeRequest) *adminv1.AddSearchAttributeRequest {
	if t == nil {
		return nil
	}
	return &adminv1.AddSearchAttributeRequest{
		SearchAttribute: FromIndexedValueTypeMap(t.SearchAttribute),
		SecurityToken:   t.SecurityToken,
	}
}

func ToAdminAddSearchAttributeRequest(t *adminv1.AddSearchAttributeRequest) *types.AddSearchAttributeRequest {
	if t == nil {
		return nil
	}
	return &types.AddSearchAttributeRequest{
		SearchAttribute: ToIndexedValueTypeMap(t.SearchAttribute),
		SecurityToken:   t.SecurityToken,
	}
}

func FromAdminCloseShardRequest(t *types.CloseShardRequest) *adminv1.CloseShardRequest {
	if t == nil {
		return nil
	}
	return &adminv1.CloseShardRequest{
		ShardId: t.ShardID,
	}
}

func ToAdminCloseShardRequest(t *adminv1.CloseShardRequest) *types.CloseShardRequest {
	if t == nil {
		return nil
	}
	return &types.CloseShardRequest{
		ShardID: t.ShardId,
	}
}

func FromAdminDescribeClusterResponse(t *types.DescribeClusterResponse) *adminv1.DescribeClusterResponse {
	if t == nil {
		return nil
	}
	return &adminv1.DescribeClusterResponse{
		SupportedClientVersions: FromSupportedClientVersions(t.SupportedClientVersions),
		MembershipInfo:          FromMembershipInfo(t.MembershipInfo),
		PersistenceInfo:         FromPersistenceInfoMap(t.PersistenceInfo),
	}
}

func ToAdminDescribeClusterResponse(t *adminv1.DescribeClusterResponse) *types.DescribeClusterResponse {
	if t == nil {
		return nil
	}
	return &types.DescribeClusterResponse{
		SupportedClientVersions: ToSupportedClientVersions(t.SupportedClientVersions),
		MembershipInfo:          ToMembershipInfo(t.MembershipInfo),
		PersistenceInfo:         ToPersistenceInfoMap(t.PersistenceInfo),
	}
}

func FromAdminDescribeHistoryHostRequest(t *types.DescribeHistoryHostRequest) *adminv1.DescribeHistoryHostRequest {
	if t == nil {
		return nil
	}
	if t.HostAddress != nil {
		return &adminv1.DescribeHistoryHostRequest{
			DescribeBy: &adminv1.DescribeHistoryHostRequest_HostAddress{HostAddress: *t.HostAddress},
		}
	}
	if t.ShardIDForHost != nil {
		return &adminv1.DescribeHistoryHostRequest{
			DescribeBy: &adminv1.DescribeHistoryHostRequest_ShardId{ShardId: *t.ShardIDForHost},
		}
	}
	if t.ExecutionForHost != nil {
		return &adminv1.DescribeHistoryHostRequest{
			DescribeBy: &adminv1.DescribeHistoryHostRequest_WorkflowExecution{WorkflowExecution: FromWorkflowExecution(t.ExecutionForHost)},
		}
	}
	panic("neither oneof field is set for DescribeHistoryHostRequest")
}

func ToAdminDescribeHistoryHostRequest(t *adminv1.DescribeHistoryHostRequest) *types.DescribeHistoryHostRequest {
	if t == nil {
		return nil
	}
	switch describeBy := t.DescribeBy.(type) {
	case *adminv1.DescribeHistoryHostRequest_HostAddress:
		return &types.DescribeHistoryHostRequest{HostAddress: &describeBy.HostAddress}
	case *adminv1.DescribeHistoryHostRequest_ShardId:
		return &types.DescribeHistoryHostRequest{ShardIDForHost: &describeBy.ShardId}
	case *adminv1.DescribeHistoryHostRequest_WorkflowExecution:
		return &types.DescribeHistoryHostRequest{ExecutionForHost: ToWorkflowExecution(describeBy.WorkflowExecution)}
	}
	panic("neither oneof field is set for DescribeHistoryHostRequest")
}

func FromAdminDescribeHistoryHostResponse(t *types.DescribeHistoryHostResponse) *adminv1.DescribeHistoryHostResponse {
	if t == nil {
		return nil
	}
	return &adminv1.DescribeHistoryHostResponse{
		NumberOfShards:        t.NumberOfShards,
		ShardIds:              t.ShardIDs,
		DomainCache:           FromDomainCacheInfo(t.DomainCache),
		ShardControllerStatus: t.ShardControllerStatus,
		Address:               t.Address,
	}
}

func ToAdminDescribeHistoryHostResponse(t *adminv1.DescribeHistoryHostResponse) *types.DescribeHistoryHostResponse {
	if t == nil {
		return nil
	}
	return &types.DescribeHistoryHostResponse{
		NumberOfShards:        t.NumberOfShards,
		ShardIDs:              t.ShardIds,
		DomainCache:           ToDomainCacheInfo(t.DomainCache),
		ShardControllerStatus: t.ShardControllerStatus,
		Address:               t.Address,
	}
}

func FromAdminDescribeQueueRequest(t *types.DescribeQueueRequest) *adminv1.DescribeQueueRequest {
	if t == nil {
		return nil
	}
	return &adminv1.DescribeQueueRequest{
		ShardId:     t.ShardID,
		ClusterName: t.ClusterName,
		TaskType:    FromTaskType(t.Type),
	}
}

func ToAdminDescribeQueueRequest(t *adminv1.DescribeQueueRequest) *types.DescribeQueueRequest {
	if t == nil {
		return nil
	}
	return &types.DescribeQueueRequest{
		ShardID:     t.ShardId,
		ClusterName: t.ClusterName,
		Type:        ToTaskType(t.TaskType),
	}
}

func FromAdminDescribeQueueResponse(t *types.DescribeQueueResponse) *adminv1.DescribeQueueResponse {
	if t == nil {
		return nil
	}
	return &adminv1.DescribeQueueResponse{
		ProcessingQueueStates: t.ProcessingQueueStates,
	}
}

func ToAdminDescribeQueueResponse(t *adminv1.DescribeQueueResponse) *types.DescribeQueueResponse {
	if t == nil {
		return nil
	}
	return &types.DescribeQueueResponse{
		ProcessingQueueStates: t.ProcessingQueueStates,
	}
}

func FromAdminDescribeWorkflowExecutionRequest(t *types.AdminDescribeWorkflowExecutionRequest) *adminv1.DescribeWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &adminv1.DescribeWorkflowExecutionRequest{
		Domain:            t.Domain,
		WorkflowExecution: FromWorkflowExecution(t.Execution),
	}
}

func ToAdminDescribeWorkflowExecutionRequest(t *adminv1.DescribeWorkflowExecutionRequest) *types.AdminDescribeWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &types.AdminDescribeWorkflowExecutionRequest{
		Domain:    t.Domain,
		Execution: ToWorkflowExecution(t.WorkflowExecution),
	}
}

func FromAdminDescribeWorkflowExecutionResponse(t *types.AdminDescribeWorkflowExecutionResponse) *adminv1.DescribeWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &adminv1.DescribeWorkflowExecutionResponse{
		ShardId:                stringToInt32(t.ShardID),
		HistoryAddr:            t.HistoryAddr,
		MutableStateInCache:    t.MutableStateInCache,
		MutableStateInDatabase: t.MutableStateInDatabase,
	}
}

func ToAdminDescribeWorkflowExecutionResponse(t *adminv1.DescribeWorkflowExecutionResponse) *types.AdminDescribeWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &types.AdminDescribeWorkflowExecutionResponse{
		ShardID:                int32ToString(t.ShardId),
		HistoryAddr:            t.HistoryAddr,
		MutableStateInCache:    t.MutableStateInCache,
		MutableStateInDatabase: t.MutableStateInDatabase,
	}
}

func FromAdminGetDLQReplicationMessagesRequest(t *types.GetDLQReplicationMessagesRequest) *adminv1.GetDLQReplicationMessagesRequest {
	if t == nil {
		return nil
	}
	return &adminv1.GetDLQReplicationMessagesRequest{
		TaskInfos: FromReplicationTaskInfoArray(t.TaskInfos),
	}
}

func ToAdminGetDLQReplicationMessagesRequest(t *adminv1.GetDLQReplicationMessagesRequest) *types.GetDLQReplicationMessagesRequest {
	if t == nil {
		return nil
	}
	return &types.GetDLQReplicationMessagesRequest{
		TaskInfos: ToReplicationTaskInfoArray(t.TaskInfos),
	}
}

func FromAdminGetDLQReplicationMessagesResponse(t *types.GetDLQReplicationMessagesResponse) *adminv1.GetDLQReplicationMessagesResponse {
	if t == nil {
		return nil
	}
	return &adminv1.GetDLQReplicationMessagesResponse{
		ReplicationTasks: FromReplicationTaskArray(t.ReplicationTasks),
	}
}

func ToAdminGetDLQReplicationMessagesResponse(t *adminv1.GetDLQReplicationMessagesResponse) *types.GetDLQReplicationMessagesResponse {
	if t == nil {
		return nil
	}
	return &types.GetDLQReplicationMessagesResponse{
		ReplicationTasks: ToReplicationTaskArray(t.ReplicationTasks),
	}
}

func FromAdminGetDomainReplicationMessagesRequest(t *types.GetDomainReplicationMessagesRequest) *adminv1.GetDomainReplicationMessagesRequest {
	if t == nil {
		return nil
	}
	return &adminv1.GetDomainReplicationMessagesRequest{
		LastRetrievedMessageId: fromInt64Value(t.LastRetrievedMessageID),
		LastProcessedMessageId: fromInt64Value(t.LastProcessedMessageID),
		ClusterName:            t.ClusterName,
	}
}

func ToAdminGetDomainReplicationMessagesRequest(t *adminv1.GetDomainReplicationMessagesRequest) *types.GetDomainReplicationMessagesRequest {
	if t == nil {
		return nil
	}
	return &types.GetDomainReplicationMessagesRequest{
		LastRetrievedMessageID: toInt64Value(t.LastRetrievedMessageId),
		LastProcessedMessageID: toInt64Value(t.LastProcessedMessageId),
		ClusterName:            t.ClusterName,
	}
}

func FromAdminGetDomainReplicationMessagesResponse(t *types.GetDomainReplicationMessagesResponse) *adminv1.GetDomainReplicationMessagesResponse {
	if t == nil {
		return nil
	}
	return &adminv1.GetDomainReplicationMessagesResponse{
		Messages: FromReplicationMessages(t.Messages),
	}
}

func ToAdminGetDomainReplicationMessagesResponse(t *adminv1.GetDomainReplicationMessagesResponse) *types.GetDomainReplicationMessagesResponse {
	if t == nil {
		return nil
	}
	return &types.GetDomainReplicationMessagesResponse{
		Messages: ToReplicationMessages(t.Messages),
	}
}

func FromAdminGetReplicationMessagesRequest(t *types.GetReplicationMessagesRequest) *adminv1.GetReplicationMessagesRequest {
	if t == nil {
		return nil
	}
	return &adminv1.GetReplicationMessagesRequest{
		Tokens:      FromReplicationTokenArray(t.Tokens),
		ClusterName: t.ClusterName,
	}
}

func ToAdminGetReplicationMessagesRequest(t *adminv1.GetReplicationMessagesRequest) *types.GetReplicationMessagesRequest {
	if t == nil {
		return nil
	}
	return &types.GetReplicationMessagesRequest{
		Tokens:      ToReplicationTokenArray(t.Tokens),
		ClusterName: t.ClusterName,
	}
}

func FromAdminGetReplicationMessagesResponse(t *types.GetReplicationMessagesResponse) *adminv1.GetReplicationMessagesResponse {
	if t == nil {
		return nil
	}
	return &adminv1.GetReplicationMessagesResponse{
		ShardMessages: FromReplicationMessagesMap(t.MessagesByShard),
	}
}

func ToAdminGetReplicationMessagesResponse(t *adminv1.GetReplicationMessagesResponse) *types.GetReplicationMessagesResponse {
	if t == nil {
		return nil
	}
	return &types.GetReplicationMessagesResponse{
		MessagesByShard: ToReplicationMessagesMap(t.ShardMessages),
	}
}

func FromAdminGetWorkflowExecutionRawHistoryV2Request(t *types.GetWorkflowExecutionRawHistoryV2Request) *adminv1.GetWorkflowExecutionRawHistoryV2Request {
	if t == nil {
		return nil
	}
	return &adminv1.GetWorkflowExecutionRawHistoryV2Request{
		Domain:            t.Domain,
		WorkflowExecution: FromWorkflowExecution(t.Execution),
		StartEvent:        FromEventIDVersionPair(t.StartEventID, t.StartEventVersion),
		EndEvent:          FromEventIDVersionPair(t.EndEventID, t.EndEventVersion),
		PageSize:          t.MaximumPageSize,
		NextPageToken:     t.NextPageToken,
	}
}

func ToAdminGetWorkflowExecutionRawHistoryV2Request(t *adminv1.GetWorkflowExecutionRawHistoryV2Request) *types.GetWorkflowExecutionRawHistoryV2Request {
	if t == nil {
		return nil
	}
	return &types.GetWorkflowExecutionRawHistoryV2Request{
		Domain:            t.Domain,
		Execution:         ToWorkflowExecution(t.WorkflowExecution),
		StartEventID:      ToEventID(t.StartEvent),
		StartEventVersion: ToEventVersion(t.StartEvent),
		EndEventID:        ToEventID(t.EndEvent),
		EndEventVersion:   ToEventVersion(t.EndEvent),
		MaximumPageSize:   t.PageSize,
		NextPageToken:     t.NextPageToken,
	}
}

func FromAdminGetWorkflowExecutionRawHistoryV2Response(t *types.GetWorkflowExecutionRawHistoryV2Response) *adminv1.GetWorkflowExecutionRawHistoryV2Response {
	if t == nil {
		return nil
	}
	return &adminv1.GetWorkflowExecutionRawHistoryV2Response{
		NextPageToken:  t.NextPageToken,
		HistoryBatches: FromDataBlobArray(t.HistoryBatches),
		VersionHistory: FromVersionHistory(t.VersionHistory),
	}
}

func ToAdminGetWorkflowExecutionRawHistoryV2Response(t *adminv1.GetWorkflowExecutionRawHistoryV2Response) *types.GetWorkflowExecutionRawHistoryV2Response {
	if t == nil {
		return nil
	}
	return &types.GetWorkflowExecutionRawHistoryV2Response{
		NextPageToken:  t.NextPageToken,
		HistoryBatches: ToDataBlobArray(t.HistoryBatches),
		VersionHistory: ToVersionHistory(t.VersionHistory),
	}
}

func FromAdminMergeDLQMessagesRequest(t *types.MergeDLQMessagesRequest) *adminv1.MergeDLQMessagesRequest {
	if t == nil {
		return nil
	}
	return &adminv1.MergeDLQMessagesRequest{
		Type:                  FromDLQType(t.Type),
		ShardId:               t.ShardID,
		SourceCluster:         t.SourceCluster,
		InclusiveEndMessageId: fromInt64Value(t.InclusiveEndMessageID),
		PageSize:              t.MaximumPageSize,
		NextPageToken:         t.NextPageToken,
	}
}

func ToAdminMergeDLQMessagesRequest(t *adminv1.MergeDLQMessagesRequest) *types.MergeDLQMessagesRequest {
	if t == nil {
		return nil
	}
	return &types.MergeDLQMessagesRequest{
		Type:                  ToDLQType(t.Type),
		ShardID:               t.ShardId,
		SourceCluster:         t.SourceCluster,
		InclusiveEndMessageID: toInt64Value(t.InclusiveEndMessageId),
		MaximumPageSize:       t.PageSize,
		NextPageToken:         t.NextPageToken,
	}
}

func FromAdminMergeDLQMessagesResponse(t *types.MergeDLQMessagesResponse) *adminv1.MergeDLQMessagesResponse {
	if t == nil {
		return nil
	}
	return &adminv1.MergeDLQMessagesResponse{
		NextPageToken: t.NextPageToken,
	}
}

func ToAdminMergeDLQMessagesResponse(t *adminv1.MergeDLQMessagesResponse) *types.MergeDLQMessagesResponse {
	if t == nil {
		return nil
	}
	return &types.MergeDLQMessagesResponse{
		NextPageToken: t.NextPageToken,
	}
}

func FromAdminPurgeDLQMessagesRequest(t *types.PurgeDLQMessagesRequest) *adminv1.PurgeDLQMessagesRequest {
	if t == nil {
		return nil
	}
	return &adminv1.PurgeDLQMessagesRequest{
		Type:                  FromDLQType(t.Type),
		ShardId:               t.ShardID,
		SourceCluster:         t.SourceCluster,
		InclusiveEndMessageId: fromInt64Value(t.InclusiveEndMessageID),
	}
}

func ToAdminPurgeDLQMessagesRequest(t *adminv1.PurgeDLQMessagesRequest) *types.PurgeDLQMessagesRequest {
	if t == nil {
		return nil
	}
	return &types.PurgeDLQMessagesRequest{
		Type:                  ToDLQType(t.Type),
		ShardID:               t.ShardId,
		SourceCluster:         t.SourceCluster,
		InclusiveEndMessageID: toInt64Value(t.InclusiveEndMessageId),
	}
}

func FromAdminReadDLQMessagesRequest(t *types.ReadDLQMessagesRequest) *adminv1.ReadDLQMessagesRequest {
	if t == nil {
		return nil
	}
	return &adminv1.ReadDLQMessagesRequest{
		Type:                  FromDLQType(t.Type),
		ShardId:               t.ShardID,
		SourceCluster:         t.SourceCluster,
		InclusiveEndMessageId: fromInt64Value(t.InclusiveEndMessageID),
		PageSize:              t.MaximumPageSize,
		NextPageToken:         t.NextPageToken,
	}
}

func ToAdminReadDLQMessagesRequest(t *adminv1.ReadDLQMessagesRequest) *types.ReadDLQMessagesRequest {
	if t == nil {
		return nil
	}
	return &types.ReadDLQMessagesRequest{
		Type:                  ToDLQType(t.Type),
		ShardID:               t.ShardId,
		SourceCluster:         t.SourceCluster,
		InclusiveEndMessageID: toInt64Value(t.InclusiveEndMessageId),
		MaximumPageSize:       t.PageSize,
		NextPageToken:         t.NextPageToken,
	}
}

func FromAdminReadDLQMessagesResponse(t *types.ReadDLQMessagesResponse) *adminv1.ReadDLQMessagesResponse {
	if t == nil {
		return nil
	}
	return &adminv1.ReadDLQMessagesResponse{
		Type:                 FromDLQType(t.Type),
		ReplicationTasks:     FromReplicationTaskArray(t.ReplicationTasks),
		ReplicationTasksInfo: FromReplicationTaskInfoArray(t.ReplicationTasksInfo),
		NextPageToken:        t.NextPageToken,
	}
}

func ToAdminReadDLQMessagesResponse(t *adminv1.ReadDLQMessagesResponse) *types.ReadDLQMessagesResponse {
	if t == nil {
		return nil
	}
	return &types.ReadDLQMessagesResponse{
		Type:                 ToDLQType(t.Type),
		ReplicationTasks:     ToReplicationTaskArray(t.ReplicationTasks),
		ReplicationTasksInfo: ToReplicationTaskInfoArray(t.ReplicationTasksInfo),
		NextPageToken:        t.NextPageToken,
	}
}

func FromAdminReapplyEventsRequest(t *types.ReapplyEventsRequest) *adminv1.ReapplyEventsRequest {
	if t == nil {
		return nil
	}
	return &adminv1.ReapplyEventsRequest{
		Domain:            t.DomainName,
		WorkflowExecution: FromWorkflowExecution(t.WorkflowExecution),
		Events:            FromDataBlob(t.Events),
	}
}

func ToAdminReapplyEventsRequest(t *adminv1.ReapplyEventsRequest) *types.ReapplyEventsRequest {
	if t == nil {
		return nil
	}
	return &types.ReapplyEventsRequest{
		DomainName:        t.Domain,
		WorkflowExecution: ToWorkflowExecution(t.WorkflowExecution),
		Events:            ToDataBlob(t.Events),
	}
}

func FromAdminRefreshWorkflowTasksRequest(t *types.RefreshWorkflowTasksRequest) *adminv1.RefreshWorkflowTasksRequest {
	if t == nil {
		return nil
	}
	return &adminv1.RefreshWorkflowTasksRequest{
		Domain:            t.Domain,
		WorkflowExecution: FromWorkflowExecution(t.Execution),
	}
}

func ToAdminRefreshWorkflowTasksRequest(t *adminv1.RefreshWorkflowTasksRequest) *types.RefreshWorkflowTasksRequest {
	if t == nil {
		return nil
	}
	return &types.RefreshWorkflowTasksRequest{
		Domain:    t.Domain,
		Execution: ToWorkflowExecution(t.WorkflowExecution),
	}
}

func FromAdminRemoveTaskRequest(t *types.RemoveTaskRequest) *adminv1.RemoveTaskRequest {
	if t == nil {
		return nil
	}
	return &adminv1.RemoveTaskRequest{
		ShardId:        t.ShardID,
		TaskType:       FromTaskType(t.Type),
		TaskId:         t.TaskID,
		VisibilityTime: unixNanoToTime(t.VisibilityTimestamp),
	}
}

func ToAdminRemoveTaskRequest(t *adminv1.RemoveTaskRequest) *types.RemoveTaskRequest {
	if t == nil {
		return nil
	}
	return &types.RemoveTaskRequest{
		ShardID:             t.ShardId,
		Type:                ToTaskType(t.TaskType),
		TaskID:              t.TaskId,
		VisibilityTimestamp: timeToUnixNano(t.VisibilityTime),
	}
}

func FromAdminResendReplicationTasksRequest(t *types.ResendReplicationTasksRequest) *adminv1.ResendReplicationTasksRequest {
	if t == nil {
		return nil
	}
	return &adminv1.ResendReplicationTasksRequest{
		DomainId:          t.DomainID,
		WorkflowExecution: FromWorkflowRunPair(t.WorkflowID, t.RunID),
		RemoteCluster:     t.RemoteCluster,
		StartEvent:        FromEventIDVersionPair(t.StartEventID, t.StartVersion),
		EndEvent:          FromEventIDVersionPair(t.EndEventID, t.EndVersion),
	}
}

func ToAdminResendReplicationTasksRequest(t *adminv1.ResendReplicationTasksRequest) *types.ResendReplicationTasksRequest {
	if t == nil {
		return nil
	}
	return &types.ResendReplicationTasksRequest{
		DomainID:      t.DomainId,
		WorkflowID:    ToWorkflowID(t.WorkflowExecution),
		RunID:         ToRunID(t.WorkflowExecution),
		RemoteCluster: t.RemoteCluster,
		StartEventID:  ToEventID(t.StartEvent),
		StartVersion:  ToEventVersion(t.StartEvent),
		EndEventID:    ToEventID(t.EndEvent),
		EndVersion:    ToEventVersion(t.EndEvent),
	}
}

func FromAdminResetQueueRequest(t *types.ResetQueueRequest) *adminv1.ResetQueueRequest {
	if t == nil {
		return nil
	}
	return &adminv1.ResetQueueRequest{
		ShardId:     t.ShardID,
		ClusterName: t.ClusterName,
		TaskType:    FromTaskType(t.Type),
	}
}

func ToAdminResetQueueRequest(t *adminv1.ResetQueueRequest) *types.ResetQueueRequest {
	if t == nil {
		return nil
	}
	return &types.ResetQueueRequest{
		ShardID:     t.ShardId,
		ClusterName: t.ClusterName,
		Type:        ToTaskType(t.TaskType),
	}
}
