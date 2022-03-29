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

func FromAdminDescribeShardDistributionRequest(t *types.DescribeShardDistributionRequest) *adminv1.DescribeShardDistributionRequest {
	if t == nil {
		return nil
	}
	return &adminv1.DescribeShardDistributionRequest{
		PageSize: t.PageSize,
		PageId:   t.PageID,
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

func ToAdminDescribeShardDistributionRequest(t *adminv1.DescribeShardDistributionRequest) *types.DescribeShardDistributionRequest {
	if t == nil {
		return nil
	}
	return &types.DescribeShardDistributionRequest{
		PageSize: t.PageSize,
		PageID:   t.PageId,
	}
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

func FromAdminDescribeShardDistributionResponse(t *types.DescribeShardDistributionResponse) *adminv1.DescribeShardDistributionResponse {
	if t == nil {
		return nil
	}
	return &adminv1.DescribeShardDistributionResponse{
		NumberOfShards: t.NumberOfShards,
		Shards:         t.Shards,
	}
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

func ToAdminDescribeShardDistributionResponse(t *adminv1.DescribeShardDistributionResponse) *types.DescribeShardDistributionResponse {
	if t == nil {
		return nil
	}
	return &types.DescribeShardDistributionResponse{
		NumberOfShards: t.NumberOfShards,
		Shards:         t.Shards,
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

func FromAdminCountDLQMessagesRequest(t *types.CountDLQMessagesRequest) *adminv1.CountDLQMessagesRequest {
	if t == nil {
		return nil
	}
	return &adminv1.CountDLQMessagesRequest{
		ForceFetch: t.ForceFetch,
	}
}

func ToAdminCountDLQMessagesRequest(t *adminv1.CountDLQMessagesRequest) *types.CountDLQMessagesRequest {
	if t == nil {
		return nil
	}
	return &types.CountDLQMessagesRequest{
		ForceFetch: t.ForceFetch,
	}
}

func FromAdminCountDLQMessagesResponse(t *types.CountDLQMessagesResponse) *adminv1.CountDLQMessagesResponse {
	if t == nil {
		return nil
	}
	return &adminv1.CountDLQMessagesResponse{
		History: FromHistoryDLQCountEntryMap(t.History),
		Domain:  t.Domain,
	}
}

func ToAdminCountDLQMessagesResponse(t *adminv1.CountDLQMessagesResponse) *types.CountDLQMessagesResponse {
	if t == nil {
		return nil
	}
	return &types.CountDLQMessagesResponse{
		History: ToHistoryDLQCountEntryMap(t.History),
		Domain:  t.Domain,
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
		ClusterName:    t.ClusterName,
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
		ClusterName:         t.ClusterName,
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

// FromAdminGetCrossClusterTasksRequest converts internal GetCrossClusterTasksRequest type to proto
func FromAdminGetCrossClusterTasksRequest(t *types.GetCrossClusterTasksRequest) *adminv1.GetCrossClusterTasksRequest {
	if t == nil {
		return nil
	}
	return &adminv1.GetCrossClusterTasksRequest{
		ShardIds:      t.ShardIDs,
		TargetCluster: t.TargetCluster,
	}
}

// ToAdminGetCrossClusterTasksRequest converts proto GetCrossClusterTasksRequest type to internal
func ToAdminGetCrossClusterTasksRequest(t *adminv1.GetCrossClusterTasksRequest) *types.GetCrossClusterTasksRequest {
	if t == nil {
		return nil
	}
	return &types.GetCrossClusterTasksRequest{
		ShardIDs:      t.ShardIds,
		TargetCluster: t.TargetCluster,
	}
}

// FromAdminGetCrossClusterTasksResponse converts internal GetCrossClusterTasksResponse type to proto
func FromAdminGetCrossClusterTasksResponse(t *types.GetCrossClusterTasksResponse) *adminv1.GetCrossClusterTasksResponse {
	if t == nil {
		return nil
	}
	return &adminv1.GetCrossClusterTasksResponse{
		TasksByShard:       FromCrossClusterTaskRequestMap(t.TasksByShard),
		FailedCauseByShard: FromGetTaskFailedCauseMap(t.FailedCauseByShard),
	}
}

// ToAdminGetCrossClusterTasksResponse converts proto GetCrossClusterTasksResponse type to internal
func ToAdminGetCrossClusterTasksResponse(t *adminv1.GetCrossClusterTasksResponse) *types.GetCrossClusterTasksResponse {
	if t == nil {
		return nil
	}
	return &types.GetCrossClusterTasksResponse{
		TasksByShard:       ToCrossClusterTaskRequestMap(t.TasksByShard),
		FailedCauseByShard: ToGetTaskFailedCauseMap(t.FailedCauseByShard),
	}
}

// FromAdminRespondCrossClusterTasksCompletedRequest converts internal RespondCrossClusterTasksCompletedRequest type to thrift
func FromAdminRespondCrossClusterTasksCompletedRequest(t *types.RespondCrossClusterTasksCompletedRequest) *adminv1.RespondCrossClusterTasksCompletedRequest {
	if t == nil {
		return nil
	}
	return &adminv1.RespondCrossClusterTasksCompletedRequest{
		ShardId:       t.ShardID,
		TargetCluster: t.TargetCluster,
		TaskResponses: FromCrossClusterTaskResponseArray(t.TaskResponses),
		FetchNewTasks: t.FetchNewTasks,
	}
}

// ToAdminRespondCrossClusterTasksCompletedRequest converts thrift RespondCrossClusterTasksCompletedRequest type to internal
func ToAdminRespondCrossClusterTasksCompletedRequest(t *adminv1.RespondCrossClusterTasksCompletedRequest) *types.RespondCrossClusterTasksCompletedRequest {
	if t == nil {
		return nil
	}
	return &types.RespondCrossClusterTasksCompletedRequest{
		ShardID:       t.ShardId,
		TargetCluster: t.TargetCluster,
		TaskResponses: ToCrossClusterTaskResponseArray(t.TaskResponses),
		FetchNewTasks: t.FetchNewTasks,
	}
}

// FromAdminRespondCrossClusterTasksCompletedResponse converts internal RespondCrossClusterTasksCompletedResponse type to thrift
func FromAdminRespondCrossClusterTasksCompletedResponse(t *types.RespondCrossClusterTasksCompletedResponse) *adminv1.RespondCrossClusterTasksCompletedResponse {
	if t == nil {
		return nil
	}
	return &adminv1.RespondCrossClusterTasksCompletedResponse{
		Tasks: FromCrossClusterTaskRequestArray(t.Tasks),
	}
}

// ToAdminRespondCrossClusterTasksCompletedResponse converts thrift RespondCrossClusterTasksCompletedResponse type to internal
func ToAdminRespondCrossClusterTasksCompletedResponse(t *adminv1.RespondCrossClusterTasksCompletedResponse) *types.RespondCrossClusterTasksCompletedResponse {
	if t == nil {
		return nil
	}
	return &types.RespondCrossClusterTasksCompletedResponse{
		Tasks: ToCrossClusterTaskRequestArray(t.Tasks),
	}
}

//FromGetDynamicConfigRequest converts internal GetDynamicConfigRequest type to proto
func FromGetDynamicConfigRequest(t *types.GetDynamicConfigRequest) *adminv1.GetDynamicConfigRequest {
	if t == nil {
		return nil
	}
	return &adminv1.GetDynamicConfigRequest{
		ConfigName: t.ConfigName,
		Filters:    FromDynamicConfigFilterArray(t.Filters),
	}
}

//ToGetDynamicConfigRequest converts proto GetDynamicConfigRequest type to internal
func ToGetDynamicConfigRequest(t *adminv1.GetDynamicConfigRequest) *types.GetDynamicConfigRequest {
	if t == nil {
		return nil
	}
	return &types.GetDynamicConfigRequest{
		ConfigName: t.ConfigName,
		Filters:    ToDynamicConfigFilterArray(t.Filters),
	}
}

//FromGetDynamicConfigResponse converts internal GetDynamicConfigResponse type to proto
func FromGetDynamicConfigResponse(t *types.GetDynamicConfigResponse) *adminv1.GetDynamicConfigResponse {
	if t == nil {
		return nil
	}
	return &adminv1.GetDynamicConfigResponse{
		Value: FromDataBlob(t.Value),
	}
}

//ToGetDynamicConfigResponse converts proto GetDynamicConfigResponse type to internal
func ToGetDynamicConfigResponse(t *adminv1.GetDynamicConfigResponse) *types.GetDynamicConfigResponse {
	if t == nil {
		return nil
	}
	return &types.GetDynamicConfigResponse{
		Value: ToDataBlob(t.Value),
	}
}

//FromUpdateDynamicConfigRequest converts internal UpdateDynamicConfigRequest type to proto
func FromUpdateDynamicConfigRequest(t *types.UpdateDynamicConfigRequest) *adminv1.UpdateDynamicConfigRequest {
	if t == nil {
		return nil
	}
	return &adminv1.UpdateDynamicConfigRequest{
		ConfigName:   t.ConfigName,
		ConfigValues: FromDynamicConfigValueArray(t.ConfigValues),
	}
}

//ToUpdateDynamicConfigRequest converts proto UpdateDynamicConfigRequest type to internal
func ToUpdateDynamicConfigRequest(t *adminv1.UpdateDynamicConfigRequest) *types.UpdateDynamicConfigRequest {
	if t == nil {
		return nil
	}
	return &types.UpdateDynamicConfigRequest{
		ConfigName:   t.ConfigName,
		ConfigValues: ToDynamicConfigValueArray(t.ConfigValues),
	}
}

//FromRestoreDynamicConfigRequest converts internal RestoreDynamicConfigRequest type to proto
func FromRestoreDynamicConfigRequest(t *types.RestoreDynamicConfigRequest) *adminv1.RestoreDynamicConfigRequest {
	if t == nil {
		return nil
	}
	return &adminv1.RestoreDynamicConfigRequest{
		ConfigName: t.ConfigName,
		Filters:    FromDynamicConfigFilterArray(t.Filters),
	}
}

//ToRestoreDynamicConfigRequest converts proto RestoreDynamicConfigRequest type to internal
func ToRestoreDynamicConfigRequest(t *adminv1.RestoreDynamicConfigRequest) *types.RestoreDynamicConfigRequest {
	if t == nil {
		return nil
	}
	return &types.RestoreDynamicConfigRequest{
		ConfigName: t.ConfigName,
		Filters:    ToDynamicConfigFilterArray(t.Filters),
	}
}

//FromAdminDeleteWorkflowRequest converts internal AdminDeleteWorkflowRequest type to proto
func FromAdminDeleteWorkflowRequest(t *types.AdminDeleteWorkflowRequest) *adminv1.AdminDeleteWorkflowRequest {
	if t == nil {
		return nil
	}
	return &adminv1.AdminDeleteWorkflowRequest{
		Domain:            t.Domain,
		WorkflowExecution: FromWorkflowExecution(t.Execution),
	}
}

//ToAdminDeleteWorkflowRequest converts proto AdminDeleteWorkflowRequest type to internal
func ToAdminDeleteWorkflowRequest(t *adminv1.AdminDeleteWorkflowRequest) *types.AdminDeleteWorkflowRequest {
	if t == nil {
		return nil
	}
	return &types.AdminDeleteWorkflowRequest{
		Domain:    t.Domain,
		Execution: ToWorkflowExecution(t.WorkflowExecution),
	}
}

//FromAdminDeleteWorkflowResponse converts internal AdminDeleteWorkflowRequest type to proto
func FromAdminDeleteWorkflowResponse(t *types.AdminDeleteWorkflowResponse) *adminv1.AdminDeleteWorkflowResponse {
	if t == nil {
		return nil
	}
	return &adminv1.AdminDeleteWorkflowResponse{
		HistoryDeleted:    t.HistoryDeleted,
		ExecutionsDeleted: t.ExecutionsDeleted,
		VisibilityDeleted: t.VisibilityDeleted,
	}
}

//ToAdminDeleteWorkflowResponse converts proto AdminDeleteWorkflowResponse type to internal
func ToAdminDeleteWorkflowResponse(t *adminv1.AdminDeleteWorkflowResponse) *types.AdminDeleteWorkflowResponse {
	if t == nil {
		return nil
	}
	return &types.AdminDeleteWorkflowResponse{
		HistoryDeleted:    t.HistoryDeleted,
		ExecutionsDeleted: t.ExecutionsDeleted,
		VisibilityDeleted: t.VisibilityDeleted,
	}
}

//FromAdminMaintainWorkflowRequest converts internal AdminMaintainWorkflowRequest type to proto
func FromAdminMaintainWorkflowRequest(t *types.AdminMaintainWorkflowRequest) *adminv1.AdminMaintainWorkflowRequest {
	if t == nil {
		return nil
	}
	return &adminv1.AdminMaintainWorkflowRequest{
		Domain:            t.Domain,
		WorkflowExecution: FromWorkflowExecution(t.Execution),
	}
}

//ToAdminMaintainWorkflowRequest converts proto AdminMaintainWorkflowRequest type to internal
func ToAdminMaintainWorkflowRequest(t *adminv1.AdminMaintainWorkflowRequest) *types.AdminMaintainWorkflowRequest {
	if t == nil {
		return nil
	}
	return &types.AdminMaintainWorkflowRequest{
		Domain:    t.Domain,
		Execution: ToWorkflowExecution(t.WorkflowExecution),
	}
}

//FromAdminMaintainWorkflowResponse converts internal AdminMaintainWorkflowResponse type to proto
func FromAdminMaintainWorkflowResponse(t *types.AdminMaintainWorkflowResponse) *adminv1.AdminMaintainWorkflowResponse {
	if t == nil {
		return nil
	}
	return &adminv1.AdminMaintainWorkflowResponse{
		HistoryDeleted:    t.HistoryDeleted,
		ExecutionsDeleted: t.ExecutionsDeleted,
		VisibilityDeleted: t.VisibilityDeleted,
	}
}

//ToAdminMaintainWorkflowResponse converts proto AdminMaintainWorkflowResponse type to internal
func ToAdminMaintainWorkflowResponse(t *adminv1.AdminMaintainWorkflowResponse) *types.AdminMaintainWorkflowResponse {
	if t == nil {
		return nil
	}
	return &types.AdminMaintainWorkflowResponse{
		HistoryDeleted:    t.HistoryDeleted,
		ExecutionsDeleted: t.ExecutionsDeleted,
		VisibilityDeleted: t.VisibilityDeleted,
	}
}

//FromListDynamicConfigRequest converts internal ListDynamicConfigRequest type to proto
func FromListDynamicConfigRequest(t *types.ListDynamicConfigRequest) *adminv1.ListDynamicConfigRequest {
	if t == nil {
		return nil
	}
	return &adminv1.ListDynamicConfigRequest{
		ConfigName: t.ConfigName,
	}
}

//ToListDynamicConfigRequest converts proto ListDynamicConfigRequest type to internal
func ToListDynamicConfigRequest(t *adminv1.ListDynamicConfigRequest) *types.ListDynamicConfigRequest {
	if t == nil {
		return nil
	}
	return &types.ListDynamicConfigRequest{
		ConfigName: t.ConfigName,
	}
}

//FromListDynamicConfigResponse converts internal ListDynamicConfigResponse type to proto
func FromListDynamicConfigResponse(t *types.ListDynamicConfigResponse) *adminv1.ListDynamicConfigResponse {
	if t == nil {
		return nil
	}
	return &adminv1.ListDynamicConfigResponse{
		Entries: FromDynamicConfigEntryArray(t.Entries),
	}
}

//ToListDynamicConfigResponse converts proto ListDynamicConfigResponse type to internal
func ToListDynamicConfigResponse(t *adminv1.ListDynamicConfigResponse) *types.ListDynamicConfigResponse {
	if t == nil {
		return nil
	}
	return &types.ListDynamicConfigResponse{
		Entries: ToDynamicConfigEntryArray(t.Entries),
	}
}

//FromDynamicConfigEntryArray converts internal DynamicConfigEntry array type to proto
func FromDynamicConfigEntryArray(t []*types.DynamicConfigEntry) []*adminv1.DynamicConfigEntry {
	if t == nil {
		return nil
	}
	v := make([]*adminv1.DynamicConfigEntry, len(t))
	for i := range t {
		v[i] = FromDynamicConfigEntry(t[i])
	}
	return v
}

//ToDynamicConfigEntryArray converts proto DynamicConfigEntry array type to internal
func ToDynamicConfigEntryArray(t []*adminv1.DynamicConfigEntry) []*types.DynamicConfigEntry {
	if t == nil {
		return nil
	}
	v := make([]*types.DynamicConfigEntry, len(t))
	for i := range t {
		v[i] = ToDynamicConfigEntry(t[i])
	}
	return v
}

//FromDynamicConfigEntry converts internal DynamicConfigEntry type to proto
func FromDynamicConfigEntry(t *types.DynamicConfigEntry) *adminv1.DynamicConfigEntry {
	if t == nil {
		return nil
	}
	return &adminv1.DynamicConfigEntry{
		Name:   t.Name,
		Values: FromDynamicConfigValueArray(t.Values),
	}
}

//ToDynamicConfigEntry converts proto DynamicConfigEntry type to internal
func ToDynamicConfigEntry(t *adminv1.DynamicConfigEntry) *types.DynamicConfigEntry {
	if t == nil {
		return nil
	}
	return &types.DynamicConfigEntry{
		Name:   t.Name,
		Values: ToDynamicConfigValueArray(t.Values),
	}
}

//FromDynamicConfigValueArray converts internal DynamicConfigValue array type to proto
func FromDynamicConfigValueArray(t []*types.DynamicConfigValue) []*adminv1.DynamicConfigValue {
	if t == nil {
		return nil
	}
	v := make([]*adminv1.DynamicConfigValue, len(t))
	for i := range t {
		v[i] = FromDynamicConfigValue(t[i])
	}
	return v
}

//ToDynamicConfigValueArray converts proto DynamicConfigValue array type to internal
func ToDynamicConfigValueArray(t []*adminv1.DynamicConfigValue) []*types.DynamicConfigValue {
	if t == nil {
		return nil
	}
	v := make([]*types.DynamicConfigValue, len(t))
	for i := range t {
		v[i] = ToDynamicConfigValue(t[i])
	}
	return v
}

//FromDynamicConfigValue converts internal DynamicConfigValue type to proto
func FromDynamicConfigValue(t *types.DynamicConfigValue) *adminv1.DynamicConfigValue {
	if t == nil {
		return nil
	}
	return &adminv1.DynamicConfigValue{
		Value:   FromDataBlob(t.Value),
		Filters: FromDynamicConfigFilterArray(t.Filters),
	}
}

//ToDynamicConfigValue converts proto DynamicConfigValue type to internal
func ToDynamicConfigValue(t *adminv1.DynamicConfigValue) *types.DynamicConfigValue {
	if t == nil {
		return nil
	}
	return &types.DynamicConfigValue{
		Value:   ToDataBlob(t.Value),
		Filters: ToDynamicConfigFilterArray(t.Filters),
	}
}

//FromDynamicConfigFilterArray converts internal DynamicConfigFilter array type to proto
func FromDynamicConfigFilterArray(t []*types.DynamicConfigFilter) []*adminv1.DynamicConfigFilter {
	if t == nil {
		return nil
	}
	v := make([]*adminv1.DynamicConfigFilter, len(t))
	for i := range t {
		v[i] = FromDynamicConfigFilter(t[i])
	}
	return v
}

//ToDynamicConfigFilterArray converts proto DynamicConfigFilter array type to internal
func ToDynamicConfigFilterArray(t []*adminv1.DynamicConfigFilter) []*types.DynamicConfigFilter {
	if t == nil {
		return nil
	}
	v := make([]*types.DynamicConfigFilter, len(t))
	for i := range t {
		v[i] = ToDynamicConfigFilter(t[i])
	}
	return v
}

//FromDynamicConfigFilter converts internal DynamicConfigFilter type to proto
func FromDynamicConfigFilter(t *types.DynamicConfigFilter) *adminv1.DynamicConfigFilter {
	if t == nil {
		return nil
	}
	return &adminv1.DynamicConfigFilter{
		Name:  t.Name,
		Value: FromDataBlob(t.Value),
	}
}

//ToDynamicConfigFilter converts thrift DynamicConfigFilter type to internal
func ToDynamicConfigFilter(t *adminv1.DynamicConfigFilter) *types.DynamicConfigFilter {
	if t == nil {
		return nil
	}
	return &types.DynamicConfigFilter{
		Name:  t.Name,
		Value: ToDataBlob(t.Value),
	}
}
