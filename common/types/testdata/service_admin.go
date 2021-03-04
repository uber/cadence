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

package testdata

import (
	"fmt"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
)

const QueueType = 2

var (
	AdminAddSearchAttributeRequest = types.AddSearchAttributeRequest{
		SearchAttribute: IndexedValueTypeMap,
		SecurityToken:   SecurityToken,
	}
	AdminCloseShardRequest = types.CloseShardRequest{
		ShardID: ShardID,
	}
	AdminDescribeClusterResponse = types.DescribeClusterResponse{
		SupportedClientVersions: &SupportedClientVersions,
		MembershipInfo:          &MembershipInfo,
	}
	AdminDescribeHistoryHostRequest_ByHost = types.DescribeHistoryHostRequest{
		HostAddress: common.StringPtr(HostName),
	}
	AdminDescribeHistoryHostRequest_ByShard = types.DescribeHistoryHostRequest{
		ShardIDForHost: common.Int32Ptr(ShardID),
	}
	AdminDescribeHistoryHostRequest_ByExecution = types.DescribeHistoryHostRequest{
		ExecutionForHost: &WorkflowExecution,
	}
	AdminDescribeHistoryHostResponse = types.DescribeHistoryHostResponse{
		NumberOfShards:        1,
		ShardIDs:              []int32{ShardID},
		DomainCache:           &DomainCacheInfo,
		ShardControllerStatus: "ShardControllerStatus",
		Address:               HostName,
	}
	AdminDescribeQueueRequest = types.DescribeQueueRequest{
		ShardID:     ShardID,
		ClusterName: ClusterName1,
		Type:        common.Int32Ptr(QueueType),
	}
	AdminDescribeQueueResponse = types.DescribeQueueResponse{
		ProcessingQueueStates: []string{"state1", "state2"},
	}
	AdminDescribeWorkflowExecutionRequest = types.AdminDescribeWorkflowExecutionRequest{
		Domain:    DomainName,
		Execution: &WorkflowExecution,
	}
	AdminDescribeWorkflowExecutionResponse = types.AdminDescribeWorkflowExecutionResponse{
		ShardID:                fmt.Sprint(ShardID),
		HistoryAddr:            HostName,
		MutableStateInCache:    "MutableStateInCache",
		MutableStateInDatabase: "MutableStateInDatabase",
	}
	AdminGetDLQReplicationMessagesRequest = types.GetDLQReplicationMessagesRequest{
		TaskInfos: ReplicationTaskInfoArray,
	}
	AdminGetDLQReplicationMessagesResponse = types.GetDLQReplicationMessagesResponse{
		ReplicationTasks: ReplicationTaskArray,
	}
	AdminGetDomainReplicationMessagesRequest = types.GetDomainReplicationMessagesRequest{
		LastRetrievedMessageID: common.Int64Ptr(MessageID1),
		LastProcessedMessageID: common.Int64Ptr(MessageID2),
		ClusterName:            ClusterName1,
	}
	AdminGetDomainReplicationMessagesResponse = types.GetDomainReplicationMessagesResponse{
		Messages: &ReplicationMessages,
	}
	AdminGetReplicationMessagesRequest = types.GetReplicationMessagesRequest{
		Tokens:      ReplicationTokenArray,
		ClusterName: ClusterName1,
	}
	AdminGetReplicationMessagesResponse = types.GetReplicationMessagesResponse{
		MessagesByShard: ReplicationMessagesMap,
	}
	AdminGetWorkflowExecutionRawHistoryV2Request = types.GetWorkflowExecutionRawHistoryV2Request{
		Domain:            DomainName,
		Execution:         &WorkflowExecution,
		StartEventID:      common.Int64Ptr(EventID1),
		StartEventVersion: common.Int64Ptr(Version1),
		EndEventID:        common.Int64Ptr(EventID2),
		EndEventVersion:   common.Int64Ptr(EventID2),
		MaximumPageSize:   PageSize,
		NextPageToken:     NextPageToken,
	}
	AdminGetWorkflowExecutionRawHistoryV2Response = types.GetWorkflowExecutionRawHistoryV2Response{
		NextPageToken:  NextPageToken,
		HistoryBatches: DataBlobArray,
		VersionHistory: &VersionHistory,
	}
	AdminMergeDLQMessagesRequest = types.MergeDLQMessagesRequest{
		Type:                  types.DLQTypeDomain.Ptr(),
		ShardID:               ShardID,
		SourceCluster:         ClusterName1,
		InclusiveEndMessageID: common.Int64Ptr(MessageID1),
		MaximumPageSize:       PageSize,
		NextPageToken:         NextPageToken,
	}
	AdminMergeDLQMessagesResponse = types.MergeDLQMessagesResponse{
		NextPageToken: NextPageToken,
	}
	AdminPurgeDLQMessagesRequest = types.PurgeDLQMessagesRequest{
		Type:                  types.DLQTypeDomain.Ptr(),
		ShardID:               ShardID,
		SourceCluster:         ClusterName1,
		InclusiveEndMessageID: common.Int64Ptr(MessageID1),
	}
	AdminReadDLQMessagesRequest = types.ReadDLQMessagesRequest{
		Type:                  types.DLQTypeDomain.Ptr(),
		ShardID:               ShardID,
		SourceCluster:         ClusterName1,
		InclusiveEndMessageID: common.Int64Ptr(MessageID1),
		MaximumPageSize:       PageSize,
		NextPageToken:         NextPageToken,
	}
	AdminReadDLQMessagesResponse = types.ReadDLQMessagesResponse{
		Type:                 types.DLQTypeDomain.Ptr(),
		ReplicationTasks:     ReplicationTaskArray,
		ReplicationTasksInfo: ReplicationTaskInfoArray,
		NextPageToken:        NextPageToken,
	}
	AdminReapplyEventsRequest = types.ReapplyEventsRequest{
		DomainName:        DomainName,
		WorkflowExecution: &WorkflowExecution,
		Events:            &DataBlob,
	}
	AdminRefreshWorkflowTasksRequest = types.RefreshWorkflowTasksRequest{
		Domain:    DomainName,
		Execution: &WorkflowExecution,
	}
	AdminRemoveTaskRequest = types.RemoveTaskRequest{
		ShardID:             ShardID,
		Type:                common.Int32Ptr(QueueType),
		TaskID:              TaskID,
		VisibilityTimestamp: &Timestamp1,
	}
	AdminResendReplicationTasksRequest = types.ResendReplicationTasksRequest{
		DomainID:      DomainID,
		WorkflowID:    WorkflowID,
		RunID:         RunID,
		RemoteCluster: ClusterName1,
		StartEventID:  common.Int64Ptr(EventID1),
		StartVersion:  common.Int64Ptr(Version1),
		EndEventID:    common.Int64Ptr(EventID2),
		EndVersion:    common.Int64Ptr(EventID2),
	}
	AdminResetQueueRequest = types.ResetQueueRequest{
		ShardID:     ShardID,
		ClusterName: ClusterName1,
		Type:        common.Int32Ptr(QueueType),
	}
)
