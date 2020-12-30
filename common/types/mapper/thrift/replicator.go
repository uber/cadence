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

	"github.com/uber/cadence/.gen/go/replicator"
)

// FromDLQType converts internal DLQType type to thrift
func FromDLQType(t *types.DLQType) *replicator.DLQType {
	if t == nil {
		return nil
	}
	switch *t {
	case types.DLQTypeReplication:
		v := replicator.DLQTypeReplication
		return &v
	case types.DLQTypeDomain:
		v := replicator.DLQTypeDomain
		return &v
	}
	panic("unexpected enum value")
}

// ToDLQType converts thrift DLQType type to internal
func ToDLQType(t *replicator.DLQType) *types.DLQType {
	if t == nil {
		return nil
	}
	switch *t {
	case replicator.DLQTypeReplication:
		v := types.DLQTypeReplication
		return &v
	case replicator.DLQTypeDomain:
		v := types.DLQTypeDomain
		return &v
	}
	panic("unexpected enum value")
}

// FromDomainOperation converts internal DomainOperation type to thrift
func FromDomainOperation(t *types.DomainOperation) *replicator.DomainOperation {
	if t == nil {
		return nil
	}
	switch *t {
	case types.DomainOperationCreate:
		v := replicator.DomainOperationCreate
		return &v
	case types.DomainOperationUpdate:
		v := replicator.DomainOperationUpdate
		return &v
	}
	panic("unexpected enum value")
}

// ToDomainOperation converts thrift DomainOperation type to internal
func ToDomainOperation(t *replicator.DomainOperation) *types.DomainOperation {
	if t == nil {
		return nil
	}
	switch *t {
	case replicator.DomainOperationCreate:
		v := types.DomainOperationCreate
		return &v
	case replicator.DomainOperationUpdate:
		v := types.DomainOperationUpdate
		return &v
	}
	panic("unexpected enum value")
}

// FromDomainTaskAttributes converts internal DomainTaskAttributes type to thrift
func FromDomainTaskAttributes(t *types.DomainTaskAttributes) *replicator.DomainTaskAttributes {
	if t == nil {
		return nil
	}
	return &replicator.DomainTaskAttributes{
		DomainOperation:         FromDomainOperation(t.DomainOperation),
		ID:                      t.ID,
		Info:                    FromDomainInfo(t.Info),
		Config:                  FromDomainConfiguration(t.Config),
		ReplicationConfig:       FromDomainReplicationConfiguration(t.ReplicationConfig),
		ConfigVersion:           t.ConfigVersion,
		FailoverVersion:         t.FailoverVersion,
		PreviousFailoverVersion: t.PreviousFailoverVersion,
	}
}

// ToDomainTaskAttributes converts thrift DomainTaskAttributes type to internal
func ToDomainTaskAttributes(t *replicator.DomainTaskAttributes) *types.DomainTaskAttributes {
	if t == nil {
		return nil
	}
	return &types.DomainTaskAttributes{
		DomainOperation:         ToDomainOperation(t.DomainOperation),
		ID:                      t.ID,
		Info:                    ToDomainInfo(t.Info),
		Config:                  ToDomainConfiguration(t.Config),
		ReplicationConfig:       ToDomainReplicationConfiguration(t.ReplicationConfig),
		ConfigVersion:           t.ConfigVersion,
		FailoverVersion:         t.FailoverVersion,
		PreviousFailoverVersion: t.PreviousFailoverVersion,
	}
}

// FromFailoverMarkerAttributes converts internal FailoverMarkerAttributes type to thrift
func FromFailoverMarkerAttributes(t *types.FailoverMarkerAttributes) *replicator.FailoverMarkerAttributes {
	if t == nil {
		return nil
	}
	return &replicator.FailoverMarkerAttributes{
		DomainID:        t.DomainID,
		FailoverVersion: t.FailoverVersion,
		CreationTime:    t.CreationTime,
	}
}

// ToFailoverMarkerAttributes converts thrift FailoverMarkerAttributes type to internal
func ToFailoverMarkerAttributes(t *replicator.FailoverMarkerAttributes) *types.FailoverMarkerAttributes {
	if t == nil {
		return nil
	}
	return &types.FailoverMarkerAttributes{
		DomainID:        t.DomainID,
		FailoverVersion: t.FailoverVersion,
		CreationTime:    t.CreationTime,
	}
}

// FromFailoverMarkers converts internal FailoverMarkers type to thrift
func FromFailoverMarkers(t *types.FailoverMarkers) *replicator.FailoverMarkers {
	if t == nil {
		return nil
	}
	return &replicator.FailoverMarkers{
		FailoverMarkers: FromFailoverMarkerAttributesArray(t.FailoverMarkers),
	}
}

// ToFailoverMarkers converts thrift FailoverMarkers type to internal
func ToFailoverMarkers(t *replicator.FailoverMarkers) *types.FailoverMarkers {
	if t == nil {
		return nil
	}
	return &types.FailoverMarkers{
		FailoverMarkers: ToFailoverMarkerAttributesArray(t.FailoverMarkers),
	}
}

// FromGetDLQReplicationMessagesRequest converts internal GetDLQReplicationMessagesRequest type to thrift
func FromGetDLQReplicationMessagesRequest(t *types.GetDLQReplicationMessagesRequest) *replicator.GetDLQReplicationMessagesRequest {
	if t == nil {
		return nil
	}
	return &replicator.GetDLQReplicationMessagesRequest{
		TaskInfos: FromReplicationTaskInfoArray(t.TaskInfos),
	}
}

// ToGetDLQReplicationMessagesRequest converts thrift GetDLQReplicationMessagesRequest type to internal
func ToGetDLQReplicationMessagesRequest(t *replicator.GetDLQReplicationMessagesRequest) *types.GetDLQReplicationMessagesRequest {
	if t == nil {
		return nil
	}
	return &types.GetDLQReplicationMessagesRequest{
		TaskInfos: ToReplicationTaskInfoArray(t.TaskInfos),
	}
}

// FromGetDLQReplicationMessagesResponse converts internal GetDLQReplicationMessagesResponse type to thrift
func FromGetDLQReplicationMessagesResponse(t *types.GetDLQReplicationMessagesResponse) *replicator.GetDLQReplicationMessagesResponse {
	if t == nil {
		return nil
	}
	return &replicator.GetDLQReplicationMessagesResponse{
		ReplicationTasks: FromReplicationTaskArray(t.ReplicationTasks),
	}
}

// ToGetDLQReplicationMessagesResponse converts thrift GetDLQReplicationMessagesResponse type to internal
func ToGetDLQReplicationMessagesResponse(t *replicator.GetDLQReplicationMessagesResponse) *types.GetDLQReplicationMessagesResponse {
	if t == nil {
		return nil
	}
	return &types.GetDLQReplicationMessagesResponse{
		ReplicationTasks: ToReplicationTaskArray(t.ReplicationTasks),
	}
}

// FromGetDomainReplicationMessagesRequest converts internal GetDomainReplicationMessagesRequest type to thrift
func FromGetDomainReplicationMessagesRequest(t *types.GetDomainReplicationMessagesRequest) *replicator.GetDomainReplicationMessagesRequest {
	if t == nil {
		return nil
	}
	return &replicator.GetDomainReplicationMessagesRequest{
		LastRetrievedMessageId: t.LastRetrievedMessageID,
		LastProcessedMessageId: t.LastProcessedMessageID,
		ClusterName:            t.ClusterName,
	}
}

// ToGetDomainReplicationMessagesRequest converts thrift GetDomainReplicationMessagesRequest type to internal
func ToGetDomainReplicationMessagesRequest(t *replicator.GetDomainReplicationMessagesRequest) *types.GetDomainReplicationMessagesRequest {
	if t == nil {
		return nil
	}
	return &types.GetDomainReplicationMessagesRequest{
		LastRetrievedMessageID: t.LastRetrievedMessageId,
		LastProcessedMessageID: t.LastProcessedMessageId,
		ClusterName:            t.ClusterName,
	}
}

// FromGetDomainReplicationMessagesResponse converts internal GetDomainReplicationMessagesResponse type to thrift
func FromGetDomainReplicationMessagesResponse(t *types.GetDomainReplicationMessagesResponse) *replicator.GetDomainReplicationMessagesResponse {
	if t == nil {
		return nil
	}
	return &replicator.GetDomainReplicationMessagesResponse{
		Messages: FromReplicationMessages(t.Messages),
	}
}

// ToGetDomainReplicationMessagesResponse converts thrift GetDomainReplicationMessagesResponse type to internal
func ToGetDomainReplicationMessagesResponse(t *replicator.GetDomainReplicationMessagesResponse) *types.GetDomainReplicationMessagesResponse {
	if t == nil {
		return nil
	}
	return &types.GetDomainReplicationMessagesResponse{
		Messages: ToReplicationMessages(t.Messages),
	}
}

// FromGetReplicationMessagesRequest converts internal GetReplicationMessagesRequest type to thrift
func FromGetReplicationMessagesRequest(t *types.GetReplicationMessagesRequest) *replicator.GetReplicationMessagesRequest {
	if t == nil {
		return nil
	}
	return &replicator.GetReplicationMessagesRequest{
		Tokens:      FromReplicationTokenArray(t.Tokens),
		ClusterName: t.ClusterName,
	}
}

// ToGetReplicationMessagesRequest converts thrift GetReplicationMessagesRequest type to internal
func ToGetReplicationMessagesRequest(t *replicator.GetReplicationMessagesRequest) *types.GetReplicationMessagesRequest {
	if t == nil {
		return nil
	}
	return &types.GetReplicationMessagesRequest{
		Tokens:      ToReplicationTokenArray(t.Tokens),
		ClusterName: t.ClusterName,
	}
}

// FromGetReplicationMessagesResponse converts internal GetReplicationMessagesResponse type to thrift
func FromGetReplicationMessagesResponse(t *types.GetReplicationMessagesResponse) *replicator.GetReplicationMessagesResponse {
	if t == nil {
		return nil
	}
	return &replicator.GetReplicationMessagesResponse{
		MessagesByShard: FromReplicationMessagesMap(t.MessagesByShard),
	}
}

// ToGetReplicationMessagesResponse converts thrift GetReplicationMessagesResponse type to internal
func ToGetReplicationMessagesResponse(t *replicator.GetReplicationMessagesResponse) *types.GetReplicationMessagesResponse {
	if t == nil {
		return nil
	}
	return &types.GetReplicationMessagesResponse{
		MessagesByShard: ToReplicationMessagesMap(t.MessagesByShard),
	}
}

// FromHistoryTaskV2Attributes converts internal HistoryTaskV2Attributes type to thrift
func FromHistoryTaskV2Attributes(t *types.HistoryTaskV2Attributes) *replicator.HistoryTaskV2Attributes {
	if t == nil {
		return nil
	}
	return &replicator.HistoryTaskV2Attributes{
		TaskId:              t.TaskID,
		DomainId:            t.DomainID,
		WorkflowId:          t.WorkflowID,
		RunId:               t.RunID,
		VersionHistoryItems: FromVersionHistoryItemArray(t.VersionHistoryItems),
		Events:              FromDataBlob(t.Events),
		NewRunEvents:        FromDataBlob(t.NewRunEvents),
	}
}

// ToHistoryTaskV2Attributes converts thrift HistoryTaskV2Attributes type to internal
func ToHistoryTaskV2Attributes(t *replicator.HistoryTaskV2Attributes) *types.HistoryTaskV2Attributes {
	if t == nil {
		return nil
	}
	return &types.HistoryTaskV2Attributes{
		TaskID:              t.TaskId,
		DomainID:            t.DomainId,
		WorkflowID:          t.WorkflowId,
		RunID:               t.RunId,
		VersionHistoryItems: ToVersionHistoryItemArray(t.VersionHistoryItems),
		Events:              ToDataBlob(t.Events),
		NewRunEvents:        ToDataBlob(t.NewRunEvents),
	}
}

// FromMergeDLQMessagesRequest converts internal MergeDLQMessagesRequest type to thrift
func FromMergeDLQMessagesRequest(t *types.MergeDLQMessagesRequest) *replicator.MergeDLQMessagesRequest {
	if t == nil {
		return nil
	}
	return &replicator.MergeDLQMessagesRequest{
		Type:                  FromDLQType(t.Type),
		ShardID:               t.ShardID,
		SourceCluster:         t.SourceCluster,
		InclusiveEndMessageID: t.InclusiveEndMessageID,
		MaximumPageSize:       t.MaximumPageSize,
		NextPageToken:         t.NextPageToken,
	}
}

// ToMergeDLQMessagesRequest converts thrift MergeDLQMessagesRequest type to internal
func ToMergeDLQMessagesRequest(t *replicator.MergeDLQMessagesRequest) *types.MergeDLQMessagesRequest {
	if t == nil {
		return nil
	}
	return &types.MergeDLQMessagesRequest{
		Type:                  ToDLQType(t.Type),
		ShardID:               t.ShardID,
		SourceCluster:         t.SourceCluster,
		InclusiveEndMessageID: t.InclusiveEndMessageID,
		MaximumPageSize:       t.MaximumPageSize,
		NextPageToken:         t.NextPageToken,
	}
}

// FromMergeDLQMessagesResponse converts internal MergeDLQMessagesResponse type to thrift
func FromMergeDLQMessagesResponse(t *types.MergeDLQMessagesResponse) *replicator.MergeDLQMessagesResponse {
	if t == nil {
		return nil
	}
	return &replicator.MergeDLQMessagesResponse{
		NextPageToken: t.NextPageToken,
	}
}

// ToMergeDLQMessagesResponse converts thrift MergeDLQMessagesResponse type to internal
func ToMergeDLQMessagesResponse(t *replicator.MergeDLQMessagesResponse) *types.MergeDLQMessagesResponse {
	if t == nil {
		return nil
	}
	return &types.MergeDLQMessagesResponse{
		NextPageToken: t.NextPageToken,
	}
}

// FromPurgeDLQMessagesRequest converts internal PurgeDLQMessagesRequest type to thrift
func FromPurgeDLQMessagesRequest(t *types.PurgeDLQMessagesRequest) *replicator.PurgeDLQMessagesRequest {
	if t == nil {
		return nil
	}
	return &replicator.PurgeDLQMessagesRequest{
		Type:                  FromDLQType(t.Type),
		ShardID:               t.ShardID,
		SourceCluster:         t.SourceCluster,
		InclusiveEndMessageID: t.InclusiveEndMessageID,
	}
}

// ToPurgeDLQMessagesRequest converts thrift PurgeDLQMessagesRequest type to internal
func ToPurgeDLQMessagesRequest(t *replicator.PurgeDLQMessagesRequest) *types.PurgeDLQMessagesRequest {
	if t == nil {
		return nil
	}
	return &types.PurgeDLQMessagesRequest{
		Type:                  ToDLQType(t.Type),
		ShardID:               t.ShardID,
		SourceCluster:         t.SourceCluster,
		InclusiveEndMessageID: t.InclusiveEndMessageID,
	}
}

// FromReadDLQMessagesRequest converts internal ReadDLQMessagesRequest type to thrift
func FromReadDLQMessagesRequest(t *types.ReadDLQMessagesRequest) *replicator.ReadDLQMessagesRequest {
	if t == nil {
		return nil
	}
	return &replicator.ReadDLQMessagesRequest{
		Type:                  FromDLQType(t.Type),
		ShardID:               t.ShardID,
		SourceCluster:         t.SourceCluster,
		InclusiveEndMessageID: t.InclusiveEndMessageID,
		MaximumPageSize:       t.MaximumPageSize,
		NextPageToken:         t.NextPageToken,
	}
}

// ToReadDLQMessagesRequest converts thrift ReadDLQMessagesRequest type to internal
func ToReadDLQMessagesRequest(t *replicator.ReadDLQMessagesRequest) *types.ReadDLQMessagesRequest {
	if t == nil {
		return nil
	}
	return &types.ReadDLQMessagesRequest{
		Type:                  ToDLQType(t.Type),
		ShardID:               t.ShardID,
		SourceCluster:         t.SourceCluster,
		InclusiveEndMessageID: t.InclusiveEndMessageID,
		MaximumPageSize:       t.MaximumPageSize,
		NextPageToken:         t.NextPageToken,
	}
}

// FromReadDLQMessagesResponse converts internal ReadDLQMessagesResponse type to thrift
func FromReadDLQMessagesResponse(t *types.ReadDLQMessagesResponse) *replicator.ReadDLQMessagesResponse {
	if t == nil {
		return nil
	}
	return &replicator.ReadDLQMessagesResponse{
		Type:                 FromDLQType(t.Type),
		ReplicationTasks:     FromReplicationTaskArray(t.ReplicationTasks),
		ReplicationTasksInfo: FromReplicationTaskInfoArray(t.ReplicationTasksInfo),
		NextPageToken:        t.NextPageToken,
	}
}

// ToReadDLQMessagesResponse converts thrift ReadDLQMessagesResponse type to internal
func ToReadDLQMessagesResponse(t *replicator.ReadDLQMessagesResponse) *types.ReadDLQMessagesResponse {
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

// FromReplicationMessages converts internal ReplicationMessages type to thrift
func FromReplicationMessages(t *types.ReplicationMessages) *replicator.ReplicationMessages {
	if t == nil {
		return nil
	}
	return &replicator.ReplicationMessages{
		ReplicationTasks:       FromReplicationTaskArray(t.ReplicationTasks),
		LastRetrievedMessageId: t.LastRetrievedMessageID,
		HasMore:                t.HasMore,
		SyncShardStatus:        FromSyncShardStatus(t.SyncShardStatus),
	}
}

// ToReplicationMessages converts thrift ReplicationMessages type to internal
func ToReplicationMessages(t *replicator.ReplicationMessages) *types.ReplicationMessages {
	if t == nil {
		return nil
	}
	return &types.ReplicationMessages{
		ReplicationTasks:       ToReplicationTaskArray(t.ReplicationTasks),
		LastRetrievedMessageID: t.LastRetrievedMessageId,
		HasMore:                t.HasMore,
		SyncShardStatus:        ToSyncShardStatus(t.SyncShardStatus),
	}
}

// FromReplicationTask converts internal ReplicationTask type to thrift
func FromReplicationTask(t *types.ReplicationTask) *replicator.ReplicationTask {
	if t == nil {
		return nil
	}
	return &replicator.ReplicationTask{
		TaskType:                      FromReplicationTaskType(t.TaskType),
		SourceTaskId:                  t.SourceTaskID,
		DomainTaskAttributes:          FromDomainTaskAttributes(t.DomainTaskAttributes),
		SyncShardStatusTaskAttributes: FromSyncShardStatusTaskAttributes(t.SyncShardStatusTaskAttributes),
		SyncActivityTaskAttributes:    FromSyncActivityTaskAttributes(t.SyncActivityTaskAttributes),
		HistoryTaskV2Attributes:       FromHistoryTaskV2Attributes(t.HistoryTaskV2Attributes),
		FailoverMarkerAttributes:      FromFailoverMarkerAttributes(t.FailoverMarkerAttributes),
		CreationTime:                  t.CreationTime,
	}
}

// ToReplicationTask converts thrift ReplicationTask type to internal
func ToReplicationTask(t *replicator.ReplicationTask) *types.ReplicationTask {
	if t == nil {
		return nil
	}
	return &types.ReplicationTask{
		TaskType:                      ToReplicationTaskType(t.TaskType),
		SourceTaskID:                  t.SourceTaskId,
		DomainTaskAttributes:          ToDomainTaskAttributes(t.DomainTaskAttributes),
		SyncShardStatusTaskAttributes: ToSyncShardStatusTaskAttributes(t.SyncShardStatusTaskAttributes),
		SyncActivityTaskAttributes:    ToSyncActivityTaskAttributes(t.SyncActivityTaskAttributes),
		HistoryTaskV2Attributes:       ToHistoryTaskV2Attributes(t.HistoryTaskV2Attributes),
		FailoverMarkerAttributes:      ToFailoverMarkerAttributes(t.FailoverMarkerAttributes),
		CreationTime:                  t.CreationTime,
	}
}

// FromReplicationTaskInfo converts internal ReplicationTaskInfo type to thrift
func FromReplicationTaskInfo(t *types.ReplicationTaskInfo) *replicator.ReplicationTaskInfo {
	if t == nil {
		return nil
	}
	return &replicator.ReplicationTaskInfo{
		DomainID:     t.DomainID,
		WorkflowID:   t.WorkflowID,
		RunID:        t.RunID,
		TaskType:     t.TaskType,
		TaskID:       t.TaskID,
		Version:      t.Version,
		FirstEventID: t.FirstEventID,
		NextEventID:  t.NextEventID,
		ScheduledID:  t.ScheduledID,
	}
}

// ToReplicationTaskInfo converts thrift ReplicationTaskInfo type to internal
func ToReplicationTaskInfo(t *replicator.ReplicationTaskInfo) *types.ReplicationTaskInfo {
	if t == nil {
		return nil
	}
	return &types.ReplicationTaskInfo{
		DomainID:     t.DomainID,
		WorkflowID:   t.WorkflowID,
		RunID:        t.RunID,
		TaskType:     t.TaskType,
		TaskID:       t.TaskID,
		Version:      t.Version,
		FirstEventID: t.FirstEventID,
		NextEventID:  t.NextEventID,
		ScheduledID:  t.ScheduledID,
	}
}

// FromReplicationTaskType converts internal ReplicationTaskType type to thrift
func FromReplicationTaskType(t *types.ReplicationTaskType) *replicator.ReplicationTaskType {
	if t == nil {
		return nil
	}
	switch *t {
	case types.ReplicationTaskTypeDomain:
		v := replicator.ReplicationTaskTypeDomain
		return &v
	case types.ReplicationTaskTypeHistory:
		v := replicator.ReplicationTaskTypeHistory
		return &v
	case types.ReplicationTaskTypeSyncShardStatus:
		v := replicator.ReplicationTaskTypeSyncShardStatus
		return &v
	case types.ReplicationTaskTypeSyncActivity:
		v := replicator.ReplicationTaskTypeSyncActivity
		return &v
	case types.ReplicationTaskTypeHistoryMetadata:
		v := replicator.ReplicationTaskTypeHistoryMetadata
		return &v
	case types.ReplicationTaskTypeHistoryV2:
		v := replicator.ReplicationTaskTypeHistoryV2
		return &v
	case types.ReplicationTaskTypeFailoverMarker:
		v := replicator.ReplicationTaskTypeFailoverMarker
		return &v
	}
	panic("unexpected enum value")
}

// ToReplicationTaskType converts thrift ReplicationTaskType type to internal
func ToReplicationTaskType(t *replicator.ReplicationTaskType) *types.ReplicationTaskType {
	if t == nil {
		return nil
	}
	switch *t {
	case replicator.ReplicationTaskTypeDomain:
		v := types.ReplicationTaskTypeDomain
		return &v
	case replicator.ReplicationTaskTypeHistory:
		v := types.ReplicationTaskTypeHistory
		return &v
	case replicator.ReplicationTaskTypeSyncShardStatus:
		v := types.ReplicationTaskTypeSyncShardStatus
		return &v
	case replicator.ReplicationTaskTypeSyncActivity:
		v := types.ReplicationTaskTypeSyncActivity
		return &v
	case replicator.ReplicationTaskTypeHistoryMetadata:
		v := types.ReplicationTaskTypeHistoryMetadata
		return &v
	case replicator.ReplicationTaskTypeHistoryV2:
		v := types.ReplicationTaskTypeHistoryV2
		return &v
	case replicator.ReplicationTaskTypeFailoverMarker:
		v := types.ReplicationTaskTypeFailoverMarker
		return &v
	}
	panic("unexpected enum value")
}

// FromReplicationToken converts internal ReplicationToken type to thrift
func FromReplicationToken(t *types.ReplicationToken) *replicator.ReplicationToken {
	if t == nil {
		return nil
	}
	return &replicator.ReplicationToken{
		ShardID:                t.ShardID,
		LastRetrievedMessageId: t.LastRetrievedMessageID,
		LastProcessedMessageId: t.LastProcessedMessageID,
	}
}

// ToReplicationToken converts thrift ReplicationToken type to internal
func ToReplicationToken(t *replicator.ReplicationToken) *types.ReplicationToken {
	if t == nil {
		return nil
	}
	return &types.ReplicationToken{
		ShardID:                t.ShardID,
		LastRetrievedMessageID: t.LastRetrievedMessageId,
		LastProcessedMessageID: t.LastProcessedMessageId,
	}
}

// FromSyncActivityTaskAttributes converts internal SyncActivityTaskAttributes type to thrift
func FromSyncActivityTaskAttributes(t *types.SyncActivityTaskAttributes) *replicator.SyncActivityTaskAttributes {
	if t == nil {
		return nil
	}
	return &replicator.SyncActivityTaskAttributes{
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

// ToSyncActivityTaskAttributes converts thrift SyncActivityTaskAttributes type to internal
func ToSyncActivityTaskAttributes(t *replicator.SyncActivityTaskAttributes) *types.SyncActivityTaskAttributes {
	if t == nil {
		return nil
	}
	return &types.SyncActivityTaskAttributes{
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

// FromSyncShardStatus converts internal SyncShardStatus type to thrift
func FromSyncShardStatus(t *types.SyncShardStatus) *replicator.SyncShardStatus {
	if t == nil {
		return nil
	}
	return &replicator.SyncShardStatus{
		Timestamp: t.Timestamp,
	}
}

// ToSyncShardStatus converts thrift SyncShardStatus type to internal
func ToSyncShardStatus(t *replicator.SyncShardStatus) *types.SyncShardStatus {
	if t == nil {
		return nil
	}
	return &types.SyncShardStatus{
		Timestamp: t.Timestamp,
	}
}

// FromSyncShardStatusTaskAttributes converts internal SyncShardStatusTaskAttributes type to thrift
func FromSyncShardStatusTaskAttributes(t *types.SyncShardStatusTaskAttributes) *replicator.SyncShardStatusTaskAttributes {
	if t == nil {
		return nil
	}
	return &replicator.SyncShardStatusTaskAttributes{
		SourceCluster: t.SourceCluster,
		ShardId:       t.ShardID,
		Timestamp:     t.Timestamp,
	}
}

// ToSyncShardStatusTaskAttributes converts thrift SyncShardStatusTaskAttributes type to internal
func ToSyncShardStatusTaskAttributes(t *replicator.SyncShardStatusTaskAttributes) *types.SyncShardStatusTaskAttributes {
	if t == nil {
		return nil
	}
	return &types.SyncShardStatusTaskAttributes{
		SourceCluster: t.SourceCluster,
		ShardID:       t.ShardId,
		Timestamp:     t.Timestamp,
	}
}

// FromFailoverMarkerAttributesArray converts internal FailoverMarkerAttributes type array to thrift
func FromFailoverMarkerAttributesArray(t []*types.FailoverMarkerAttributes) []*replicator.FailoverMarkerAttributes {
	if t == nil {
		return nil
	}
	v := make([]*replicator.FailoverMarkerAttributes, len(t))
	for i := range t {
		v[i] = FromFailoverMarkerAttributes(t[i])
	}
	return v
}

// ToFailoverMarkerAttributesArray converts thrift FailoverMarkerAttributes type array to internal
func ToFailoverMarkerAttributesArray(t []*replicator.FailoverMarkerAttributes) []*types.FailoverMarkerAttributes {
	if t == nil {
		return nil
	}
	v := make([]*types.FailoverMarkerAttributes, len(t))
	for i := range t {
		v[i] = ToFailoverMarkerAttributes(t[i])
	}
	return v
}

// FromReplicationTaskInfoArray converts internal ReplicationTaskInfo type array to thrift
func FromReplicationTaskInfoArray(t []*types.ReplicationTaskInfo) []*replicator.ReplicationTaskInfo {
	if t == nil {
		return nil
	}
	v := make([]*replicator.ReplicationTaskInfo, len(t))
	for i := range t {
		v[i] = FromReplicationTaskInfo(t[i])
	}
	return v
}

// ToReplicationTaskInfoArray converts thrift ReplicationTaskInfo type array to internal
func ToReplicationTaskInfoArray(t []*replicator.ReplicationTaskInfo) []*types.ReplicationTaskInfo {
	if t == nil {
		return nil
	}
	v := make([]*types.ReplicationTaskInfo, len(t))
	for i := range t {
		v[i] = ToReplicationTaskInfo(t[i])
	}
	return v
}

// FromReplicationTaskArray converts internal ReplicationTask type array to thrift
func FromReplicationTaskArray(t []*types.ReplicationTask) []*replicator.ReplicationTask {
	if t == nil {
		return nil
	}
	v := make([]*replicator.ReplicationTask, len(t))
	for i := range t {
		v[i] = FromReplicationTask(t[i])
	}
	return v
}

// ToReplicationTaskArray converts thrift ReplicationTask type array to internal
func ToReplicationTaskArray(t []*replicator.ReplicationTask) []*types.ReplicationTask {
	if t == nil {
		return nil
	}
	v := make([]*types.ReplicationTask, len(t))
	for i := range t {
		v[i] = ToReplicationTask(t[i])
	}
	return v
}

// FromReplicationTokenArray converts internal ReplicationToken type array to thrift
func FromReplicationTokenArray(t []*types.ReplicationToken) []*replicator.ReplicationToken {
	if t == nil {
		return nil
	}
	v := make([]*replicator.ReplicationToken, len(t))
	for i := range t {
		v[i] = FromReplicationToken(t[i])
	}
	return v
}

// ToReplicationTokenArray converts thrift ReplicationToken type array to internal
func ToReplicationTokenArray(t []*replicator.ReplicationToken) []*types.ReplicationToken {
	if t == nil {
		return nil
	}
	v := make([]*types.ReplicationToken, len(t))
	for i := range t {
		v[i] = ToReplicationToken(t[i])
	}
	return v
}

// FromReplicationMessagesMap converts internal ReplicationMessages type map to thrift
func FromReplicationMessagesMap(t map[int32]*types.ReplicationMessages) map[int32]*replicator.ReplicationMessages {
	if t == nil {
		return nil
	}
	v := make(map[int32]*replicator.ReplicationMessages, len(t))
	for key := range t {
		v[key] = FromReplicationMessages(t[key])
	}
	return v
}

// ToReplicationMessagesMap converts thrift ReplicationMessages type map to internal
func ToReplicationMessagesMap(t map[int32]*replicator.ReplicationMessages) map[int32]*types.ReplicationMessages {
	if t == nil {
		return nil
	}
	v := make(map[int32]*types.ReplicationMessages, len(t))
	for key := range t {
		v[key] = ToReplicationMessages(t[key])
	}
	return v
}
