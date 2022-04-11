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
	sharedv1 "github.com/uber/cadence/.gen/proto/shared/v1"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

func FromHostInfo(t *types.HostInfo) *sharedv1.HostInfo {
	if t == nil {
		return nil
	}
	return &sharedv1.HostInfo{
		Identity: t.Identity,
	}
}

func ToHostInfo(t *sharedv1.HostInfo) *types.HostInfo {
	if t == nil {
		return nil
	}
	return &types.HostInfo{
		Identity: t.Identity,
	}
}

func FromMembershipInfo(t *types.MembershipInfo) *sharedv1.MembershipInfo {
	if t == nil {
		return nil
	}
	return &sharedv1.MembershipInfo{
		CurrentHost:      FromHostInfo(t.CurrentHost),
		ReachableMembers: t.ReachableMembers,
		Rings:            FromRingInfoArray(t.Rings),
	}
}

func FromPersistenceSettings(t []*types.PersistenceSetting) []*sharedv1.PersistenceSetting {
	if t == nil {
		return nil
	}
	v := make([]*sharedv1.PersistenceSetting, len(t))
	for i := range t {
		v[i] = FromPersistenceSetting(t[i])
	}
	return v
}

func FromPersistenceSetting(t *types.PersistenceSetting) *sharedv1.PersistenceSetting {
	if t == nil {
		return nil
	}
	return &sharedv1.PersistenceSetting{
		Key:   t.Key,
		Value: t.Value,
	}
}

func FromPersistenceFeatures(t []*types.PersistenceFeature) []*sharedv1.PersistenceFeature {
	if t == nil {
		return nil
	}
	v := make([]*sharedv1.PersistenceFeature, len(t))
	for i := range t {
		v[i] = FromPersistenceFeature(t[i])
	}
	return v
}

func FromPersistenceFeature(t *types.PersistenceFeature) *sharedv1.PersistenceFeature {
	if t == nil {
		return nil
	}
	return &sharedv1.PersistenceFeature{
		Key:     t.Key,
		Enabled: t.Enabled,
	}
}

func FromPersistenceInfoMap(t map[string]*types.PersistenceInfo) map[string]*sharedv1.PersistenceInfo {
	if t == nil {
		return nil
	}
	v := make(map[string]*sharedv1.PersistenceInfo, len(t))
	for key := range t {
		v[key] = FromPersistenceInfo(t[key])
	}

	return v
}

func FromPersistenceInfo(t *types.PersistenceInfo) *sharedv1.PersistenceInfo {
	if t == nil {
		return nil
	}

	return &sharedv1.PersistenceInfo{
		Backend:  t.Backend,
		Settings: FromPersistenceSettings(t.Settings),
		Features: FromPersistenceFeatures(t.Features),
	}
}

func ToPersistenceSettings(t []*sharedv1.PersistenceSetting) []*types.PersistenceSetting {
	if t == nil {
		return nil
	}
	v := make([]*types.PersistenceSetting, len(t))
	for i := range t {
		v[i] = ToPersistenceSetting(t[i])
	}
	return v
}

func ToPersistenceSetting(t *sharedv1.PersistenceSetting) *types.PersistenceSetting {
	if t == nil {
		return nil
	}
	return &types.PersistenceSetting{
		Key:   t.Key,
		Value: t.Value,
	}
}

func ToPersistenceFeatures(t []*sharedv1.PersistenceFeature) []*types.PersistenceFeature {
	if t == nil {
		return nil
	}
	v := make([]*types.PersistenceFeature, len(t))
	for i := range t {
		v[i] = ToPersistenceFeature(t[i])
	}
	return v
}

func ToPersistenceFeature(t *sharedv1.PersistenceFeature) *types.PersistenceFeature {
	if t == nil {
		return nil
	}
	return &types.PersistenceFeature{
		Key:     t.Key,
		Enabled: t.Enabled,
	}
}

func ToPersistenceInfoMap(t map[string]*sharedv1.PersistenceInfo) map[string]*types.PersistenceInfo {
	if t == nil {
		return nil
	}
	v := make(map[string]*types.PersistenceInfo, len(t))
	for key := range t {
		v[key] = ToPersistenceInfo(t[key])
	}

	return v
}

func ToPersistenceInfo(t *sharedv1.PersistenceInfo) *types.PersistenceInfo {
	if t == nil {
		return nil
	}

	return &types.PersistenceInfo{
		Backend:  t.Backend,
		Settings: ToPersistenceSettings(t.Settings),
		Features: ToPersistenceFeatures(t.Features),
	}
}

func ToMembershipInfo(t *sharedv1.MembershipInfo) *types.MembershipInfo {
	if t == nil {
		return nil
	}
	return &types.MembershipInfo{
		CurrentHost:      ToHostInfo(t.CurrentHost),
		ReachableMembers: t.ReachableMembers,
		Rings:            ToRingInfoArray(t.Rings),
	}
}

func FromDomainCacheInfo(t *types.DomainCacheInfo) *sharedv1.DomainCacheInfo {
	if t == nil {
		return nil
	}
	return &sharedv1.DomainCacheInfo{
		NumOfItemsInCacheById:   t.NumOfItemsInCacheByID,
		NumOfItemsInCacheByName: t.NumOfItemsInCacheByName,
	}
}

func ToDomainCacheInfo(t *sharedv1.DomainCacheInfo) *types.DomainCacheInfo {
	if t == nil {
		return nil
	}
	return &types.DomainCacheInfo{
		NumOfItemsInCacheByID:   t.NumOfItemsInCacheById,
		NumOfItemsInCacheByName: t.NumOfItemsInCacheByName,
	}
}

func FromRingInfo(t *types.RingInfo) *sharedv1.RingInfo {
	if t == nil {
		return nil
	}
	return &sharedv1.RingInfo{
		Role:        t.Role,
		MemberCount: t.MemberCount,
		Members:     FromHostInfoArray(t.Members),
	}
}

func ToRingInfo(t *sharedv1.RingInfo) *types.RingInfo {
	if t == nil {
		return nil
	}
	return &types.RingInfo{
		Role:        t.Role,
		MemberCount: t.MemberCount,
		Members:     ToHostInfoArray(t.Members),
	}
}

func FromTaskSource(t *types.TaskSource) sharedv1.TaskSource {
	if t == nil {
		return sharedv1.TaskSource_TASK_SOURCE_INVALID
	}
	switch *t {
	case types.TaskSourceHistory:
		return sharedv1.TaskSource_TASK_SOURCE_HISTORY
	case types.TaskSourceDbBacklog:
		return sharedv1.TaskSource_TASK_SOURCE_DB_BACKLOG
	}
	panic("unexpected enum value")
}

func ToTaskSource(t sharedv1.TaskSource) *types.TaskSource {
	switch t {
	case sharedv1.TaskSource_TASK_SOURCE_INVALID:
		return nil
	case sharedv1.TaskSource_TASK_SOURCE_HISTORY:
		return types.TaskSourceHistory.Ptr()
	case sharedv1.TaskSource_TASK_SOURCE_DB_BACKLOG:
		return types.TaskSourceDbBacklog.Ptr()
	}
	panic("unexpected enum value")
}

func FromTransientDecisionInfo(t *types.TransientDecisionInfo) *sharedv1.TransientDecisionInfo {
	if t == nil {
		return nil
	}
	return &sharedv1.TransientDecisionInfo{
		ScheduledEvent: FromHistoryEvent(t.ScheduledEvent),
		StartedEvent:   FromHistoryEvent(t.StartedEvent),
	}
}

func ToTransientDecisionInfo(t *sharedv1.TransientDecisionInfo) *types.TransientDecisionInfo {
	if t == nil {
		return nil
	}
	return &types.TransientDecisionInfo{
		ScheduledEvent: ToHistoryEvent(t.ScheduledEvent),
		StartedEvent:   ToHistoryEvent(t.StartedEvent),
	}
}

func FromVersionHistories(t *types.VersionHistories) *sharedv1.VersionHistories {
	if t == nil {
		return nil
	}
	return &sharedv1.VersionHistories{
		CurrentVersionHistoryIndex: t.CurrentVersionHistoryIndex,
		Histories:                  FromVersionHistoryArray(t.Histories),
	}
}

func ToVersionHistories(t *sharedv1.VersionHistories) *types.VersionHistories {
	if t == nil {
		return nil
	}
	return &types.VersionHistories{
		CurrentVersionHistoryIndex: t.CurrentVersionHistoryIndex,
		Histories:                  ToVersionHistoryArray(t.Histories),
	}
}

func FromVersionHistory(t *types.VersionHistory) *sharedv1.VersionHistory {
	if t == nil {
		return nil
	}
	return &sharedv1.VersionHistory{
		BranchToken: t.BranchToken,
		Items:       FromVersionHistoryItemArray(t.Items),
	}
}

func ToVersionHistory(t *sharedv1.VersionHistory) *types.VersionHistory {
	if t == nil {
		return nil
	}
	return &types.VersionHistory{
		BranchToken: t.BranchToken,
		Items:       ToVersionHistoryItemArray(t.Items),
	}
}

func FromVersionHistoryItem(t *types.VersionHistoryItem) *sharedv1.VersionHistoryItem {
	if t == nil {
		return nil
	}
	return &sharedv1.VersionHistoryItem{
		EventId: t.EventID,
		Version: t.Version,
	}
}

func ToVersionHistoryItem(t *sharedv1.VersionHistoryItem) *types.VersionHistoryItem {
	if t == nil {
		return nil
	}
	return &types.VersionHistoryItem{
		EventID: t.EventId,
		Version: t.Version,
	}
}

func FromWorkflowState(t *int32) sharedv1.WorkflowState {
	if t == nil {
		return sharedv1.WorkflowState_WORKFLOW_STATE_INVALID
	}
	switch *t {
	case persistence.WorkflowStateCreated:
		return sharedv1.WorkflowState_WORKFLOW_STATE_CREATED
	case persistence.WorkflowStateRunning:
		return sharedv1.WorkflowState_WORKFLOW_STATE_RUNNING
	case persistence.WorkflowStateCompleted:
		return sharedv1.WorkflowState_WORKFLOW_STATE_COMPLETED
	case persistence.WorkflowStateZombie:
		return sharedv1.WorkflowState_WORKFLOW_STATE_ZOMBIE
	case persistence.WorkflowStateVoid:
		return sharedv1.WorkflowState_WORKFLOW_STATE_VOID
	case persistence.WorkflowStateCorrupted:
		return sharedv1.WorkflowState_WORKFLOW_STATE_CORRUPTED
	}
	panic("unexpected enum value")
}

func ToWorkflowState(t sharedv1.WorkflowState) *int32 {
	switch t {
	case sharedv1.WorkflowState_WORKFLOW_STATE_INVALID:
		return nil
	case sharedv1.WorkflowState_WORKFLOW_STATE_CREATED:
		v := int32(persistence.WorkflowStateCreated)
		return &v
	case sharedv1.WorkflowState_WORKFLOW_STATE_RUNNING:
		v := int32(persistence.WorkflowStateRunning)
		return &v
	case sharedv1.WorkflowState_WORKFLOW_STATE_COMPLETED:
		v := int32(persistence.WorkflowStateCompleted)
		return &v
	case sharedv1.WorkflowState_WORKFLOW_STATE_ZOMBIE:
		v := int32(persistence.WorkflowStateZombie)
		return &v
	case sharedv1.WorkflowState_WORKFLOW_STATE_VOID:
		v := int32(persistence.WorkflowStateVoid)
		return &v
	case sharedv1.WorkflowState_WORKFLOW_STATE_CORRUPTED:
		v := int32(persistence.WorkflowStateCorrupted)
		return &v
	}
	panic("unexpected enum value")
}

func FromHostInfoArray(t []*types.HostInfo) []*sharedv1.HostInfo {
	if t == nil {
		return nil
	}
	v := make([]*sharedv1.HostInfo, len(t))
	for i := range t {
		v[i] = FromHostInfo(t[i])
	}
	return v
}

func ToHostInfoArray(t []*sharedv1.HostInfo) []*types.HostInfo {
	if t == nil {
		return nil
	}
	v := make([]*types.HostInfo, len(t))
	for i := range t {
		v[i] = ToHostInfo(t[i])
	}
	return v
}

func FromVersionHistoryArray(t []*types.VersionHistory) []*sharedv1.VersionHistory {
	if t == nil {
		return nil
	}
	v := make([]*sharedv1.VersionHistory, len(t))
	for i := range t {
		v[i] = FromVersionHistory(t[i])
	}
	return v
}

func ToVersionHistoryArray(t []*sharedv1.VersionHistory) []*types.VersionHistory {
	if t == nil {
		return nil
	}
	v := make([]*types.VersionHistory, len(t))
	for i := range t {
		v[i] = ToVersionHistory(t[i])
	}
	return v
}

func FromRingInfoArray(t []*types.RingInfo) []*sharedv1.RingInfo {
	if t == nil {
		return nil
	}
	v := make([]*sharedv1.RingInfo, len(t))
	for i := range t {
		v[i] = FromRingInfo(t[i])
	}
	return v
}

func ToRingInfoArray(t []*sharedv1.RingInfo) []*types.RingInfo {
	if t == nil {
		return nil
	}
	v := make([]*types.RingInfo, len(t))
	for i := range t {
		v[i] = ToRingInfo(t[i])
	}
	return v
}

func FromDLQType(t *types.DLQType) sharedv1.DLQType {
	if t == nil {
		return sharedv1.DLQType_DLQ_TYPE_INVALID
	}
	switch *t {
	case types.DLQTypeReplication:
		return sharedv1.DLQType_DLQ_TYPE_REPLICATION
	case types.DLQTypeDomain:
		return sharedv1.DLQType_DLQ_TYPE_DOMAIN
	}
	panic("unexpected enum value")
}

func ToDLQType(t sharedv1.DLQType) *types.DLQType {
	switch t {
	case sharedv1.DLQType_DLQ_TYPE_INVALID:
		return nil
	case sharedv1.DLQType_DLQ_TYPE_REPLICATION:
		return types.DLQTypeReplication.Ptr()
	case sharedv1.DLQType_DLQ_TYPE_DOMAIN:
		return types.DLQTypeDomain.Ptr()
	}
	panic("unexpected enum value")
}

func FromDomainOperation(t *types.DomainOperation) sharedv1.DomainOperation {
	if t == nil {
		return sharedv1.DomainOperation_DOMAIN_OPERATION_INVALID
	}
	switch *t {
	case types.DomainOperationCreate:
		return sharedv1.DomainOperation_DOMAIN_OPERATION_CREATE
	case types.DomainOperationUpdate:
		return sharedv1.DomainOperation_DOMAIN_OPERATION_UPDATE
	}
	panic("unexpected enum value")
}

func ToDomainOperation(t sharedv1.DomainOperation) *types.DomainOperation {
	switch t {
	case sharedv1.DomainOperation_DOMAIN_OPERATION_INVALID:
		return nil
	case sharedv1.DomainOperation_DOMAIN_OPERATION_CREATE:
		return types.DomainOperationCreate.Ptr()
	case sharedv1.DomainOperation_DOMAIN_OPERATION_UPDATE:
		return types.DomainOperationUpdate.Ptr()
	}
	panic("unexpected enum value")
}

func FromDomainTaskAttributes(t *types.DomainTaskAttributes) *sharedv1.DomainTaskAttributes {
	if t == nil {
		return nil
	}
	return &sharedv1.DomainTaskAttributes{
		DomainOperation: FromDomainOperation(t.DomainOperation),
		Id:              t.ID,
		Domain: FromDescribeDomainResponseDomain(&types.DescribeDomainResponse{
			DomainInfo:               t.Info,
			Configuration:            t.Config,
			ReplicationConfiguration: t.ReplicationConfig,
		}),
		ConfigVersion:           t.ConfigVersion,
		FailoverVersion:         t.FailoverVersion,
		PreviousFailoverVersion: t.PreviousFailoverVersion,
	}
}

func ToDomainTaskAttributes(t *sharedv1.DomainTaskAttributes) *types.DomainTaskAttributes {
	if t == nil {
		return nil
	}
	domain := ToDescribeDomainResponseDomain(t.Domain)
	return &types.DomainTaskAttributes{
		DomainOperation:         ToDomainOperation(t.DomainOperation),
		ID:                      t.Id,
		Info:                    domain.DomainInfo,
		Config:                  domain.Configuration,
		ReplicationConfig:       domain.ReplicationConfiguration,
		ConfigVersion:           t.ConfigVersion,
		FailoverVersion:         t.FailoverVersion,
		PreviousFailoverVersion: t.PreviousFailoverVersion,
	}
}

func FromFailoverMarkerAttributes(t *types.FailoverMarkerAttributes) *sharedv1.FailoverMarkerAttributes {
	if t == nil {
		return nil
	}
	return &sharedv1.FailoverMarkerAttributes{
		DomainId:        t.DomainID,
		FailoverVersion: t.FailoverVersion,
		CreationTime:    unixNanoToTime(t.CreationTime),
	}
}

func ToFailoverMarkerAttributes(t *sharedv1.FailoverMarkerAttributes) *types.FailoverMarkerAttributes {
	if t == nil {
		return nil
	}
	return &types.FailoverMarkerAttributes{
		DomainID:        t.DomainId,
		FailoverVersion: t.FailoverVersion,
		CreationTime:    timeToUnixNano(t.CreationTime),
	}
}

func FromFailoverMarkerToken(t *types.FailoverMarkerToken) *sharedv1.FailoverMarkerToken {
	if t == nil {
		return nil
	}
	return &sharedv1.FailoverMarkerToken{
		ShardIds:       t.ShardIDs,
		FailoverMarker: FromFailoverMarkerAttributes(t.FailoverMarker),
	}
}

func ToFailoverMarkerToken(t *sharedv1.FailoverMarkerToken) *types.FailoverMarkerToken {
	if t == nil {
		return nil
	}
	return &types.FailoverMarkerToken{
		ShardIDs:       t.ShardIds,
		FailoverMarker: ToFailoverMarkerAttributes(t.FailoverMarker),
	}
}

func FromHistoryTaskV2Attributes(t *types.HistoryTaskV2Attributes) *sharedv1.HistoryTaskV2Attributes {
	if t == nil {
		return nil
	}
	return &sharedv1.HistoryTaskV2Attributes{
		TaskId:              t.TaskID,
		DomainId:            t.DomainID,
		WorkflowExecution:   FromWorkflowRunPair(t.WorkflowID, t.RunID),
		VersionHistoryItems: FromVersionHistoryItemArray(t.VersionHistoryItems),
		Events:              FromDataBlob(t.Events),
		NewRunEvents:        FromDataBlob(t.NewRunEvents),
	}
}

func ToHistoryTaskV2Attributes(t *sharedv1.HistoryTaskV2Attributes) *types.HistoryTaskV2Attributes {
	if t == nil {
		return nil
	}
	return &types.HistoryTaskV2Attributes{
		TaskID:              t.TaskId,
		DomainID:            t.DomainId,
		WorkflowID:          ToWorkflowID(t.WorkflowExecution),
		RunID:               ToRunID(t.WorkflowExecution),
		VersionHistoryItems: ToVersionHistoryItemArray(t.VersionHistoryItems),
		Events:              ToDataBlob(t.Events),
		NewRunEvents:        ToDataBlob(t.NewRunEvents),
	}
}

func FromReplicationMessages(t *types.ReplicationMessages) *sharedv1.ReplicationMessages {
	if t == nil {
		return nil
	}
	return &sharedv1.ReplicationMessages{
		ReplicationTasks:       FromReplicationTaskArray(t.ReplicationTasks),
		LastRetrievedMessageId: t.LastRetrievedMessageID,
		HasMore:                t.HasMore,
		SyncShardStatus:        FromSyncShardStatus(t.SyncShardStatus),
	}
}

func ToReplicationMessages(t *sharedv1.ReplicationMessages) *types.ReplicationMessages {
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

func FromReplicationTaskInfo(t *types.ReplicationTaskInfo) *sharedv1.ReplicationTaskInfo {
	if t == nil {
		return nil
	}
	return &sharedv1.ReplicationTaskInfo{
		DomainId:          t.DomainID,
		WorkflowExecution: FromWorkflowRunPair(t.WorkflowID, t.RunID),
		TaskType:          int32(t.TaskType),
		TaskId:            t.TaskID,
		Version:           t.Version,
		FirstEventId:      t.FirstEventID,
		NextEventId:       t.NextEventID,
		ScheduledId:       t.ScheduledID,
	}
}

func ToReplicationTaskInfo(t *sharedv1.ReplicationTaskInfo) *types.ReplicationTaskInfo {
	if t == nil {
		return nil
	}
	return &types.ReplicationTaskInfo{
		DomainID:     t.DomainId,
		WorkflowID:   ToWorkflowID(t.WorkflowExecution),
		RunID:        ToRunID(t.WorkflowExecution),
		TaskType:     int16(t.TaskType),
		TaskID:       t.TaskId,
		Version:      t.Version,
		FirstEventID: t.FirstEventId,
		NextEventID:  t.NextEventId,
		ScheduledID:  t.ScheduledId,
	}
}

func FromReplicationTaskType(t *types.ReplicationTaskType) sharedv1.ReplicationTaskType {
	if t == nil {
		return sharedv1.ReplicationTaskType_REPLICATION_TASK_TYPE_INVALID
	}
	switch *t {
	case types.ReplicationTaskTypeDomain:
		return sharedv1.ReplicationTaskType_REPLICATION_TASK_TYPE_DOMAIN
	case types.ReplicationTaskTypeHistory:
		return sharedv1.ReplicationTaskType_REPLICATION_TASK_TYPE_HISTORY
	case types.ReplicationTaskTypeSyncShardStatus:
		return sharedv1.ReplicationTaskType_REPLICATION_TASK_TYPE_SYNC_SHARD_STATUS
	case types.ReplicationTaskTypeSyncActivity:
		return sharedv1.ReplicationTaskType_REPLICATION_TASK_TYPE_SYNC_ACTIVITY
	case types.ReplicationTaskTypeHistoryMetadata:
		return sharedv1.ReplicationTaskType_REPLICATION_TASK_TYPE_HISTORY_METADATA
	case types.ReplicationTaskTypeHistoryV2:
		return sharedv1.ReplicationTaskType_REPLICATION_TASK_TYPE_HISTORY_V2
	case types.ReplicationTaskTypeFailoverMarker:
		return sharedv1.ReplicationTaskType_REPLICATION_TASK_TYPE_FAILOVER_MARKER
	}
	panic("unexpected enum value")
}

func ToReplicationTaskType(t sharedv1.ReplicationTaskType) *types.ReplicationTaskType {
	switch t {
	case sharedv1.ReplicationTaskType_REPLICATION_TASK_TYPE_INVALID:
		return nil
	case sharedv1.ReplicationTaskType_REPLICATION_TASK_TYPE_DOMAIN:
		return types.ReplicationTaskTypeDomain.Ptr()
	case sharedv1.ReplicationTaskType_REPLICATION_TASK_TYPE_HISTORY:
		return types.ReplicationTaskTypeHistory.Ptr()
	case sharedv1.ReplicationTaskType_REPLICATION_TASK_TYPE_SYNC_SHARD_STATUS:
		return types.ReplicationTaskTypeSyncShardStatus.Ptr()
	case sharedv1.ReplicationTaskType_REPLICATION_TASK_TYPE_SYNC_ACTIVITY:
		return types.ReplicationTaskTypeSyncActivity.Ptr()
	case sharedv1.ReplicationTaskType_REPLICATION_TASK_TYPE_HISTORY_METADATA:
		return types.ReplicationTaskTypeHistoryMetadata.Ptr()
	case sharedv1.ReplicationTaskType_REPLICATION_TASK_TYPE_HISTORY_V2:
		return types.ReplicationTaskTypeHistoryV2.Ptr()
	case sharedv1.ReplicationTaskType_REPLICATION_TASK_TYPE_FAILOVER_MARKER:
		return types.ReplicationTaskTypeFailoverMarker.Ptr()
	}
	panic("unexpected enum value")
}

func FromReplicationToken(t *types.ReplicationToken) *sharedv1.ReplicationToken {
	if t == nil {
		return nil
	}
	return &sharedv1.ReplicationToken{
		ShardId:                t.ShardID,
		LastRetrievedMessageId: t.LastRetrievedMessageID,
		LastProcessedMessageId: t.LastProcessedMessageID,
	}
}

func ToReplicationToken(t *sharedv1.ReplicationToken) *types.ReplicationToken {
	if t == nil {
		return nil
	}
	return &types.ReplicationToken{
		ShardID:                t.ShardId,
		LastRetrievedMessageID: t.LastRetrievedMessageId,
		LastProcessedMessageID: t.LastProcessedMessageId,
	}
}

func FromSyncActivityTaskAttributes(t *types.SyncActivityTaskAttributes) *sharedv1.SyncActivityTaskAttributes {
	if t == nil {
		return nil
	}
	return &sharedv1.SyncActivityTaskAttributes{
		DomainId:           t.DomainID,
		WorkflowExecution:  FromWorkflowRunPair(t.WorkflowID, t.RunID),
		Version:            t.Version,
		ScheduledId:        t.ScheduledID,
		ScheduledTime:      unixNanoToTime(t.ScheduledTime),
		StartedId:          t.StartedID,
		StartedTime:        unixNanoToTime(t.StartedTime),
		LastHeartbeatTime:  unixNanoToTime(t.LastHeartbeatTime),
		Details:            FromPayload(t.Details),
		Attempt:            t.Attempt,
		LastFailure:        FromFailure(t.LastFailureReason, t.LastFailureDetails),
		LastWorkerIdentity: t.LastWorkerIdentity,
		VersionHistory:     FromVersionHistory(t.VersionHistory),
	}
}

func ToSyncActivityTaskAttributes(t *sharedv1.SyncActivityTaskAttributes) *types.SyncActivityTaskAttributes {
	if t == nil {
		return nil
	}
	return &types.SyncActivityTaskAttributes{
		DomainID:           t.DomainId,
		WorkflowID:         ToWorkflowID(t.WorkflowExecution),
		RunID:              ToRunID(t.WorkflowExecution),
		Version:            t.Version,
		ScheduledID:        t.ScheduledId,
		ScheduledTime:      timeToUnixNano(t.ScheduledTime),
		StartedID:          t.StartedId,
		StartedTime:        timeToUnixNano(t.StartedTime),
		LastHeartbeatTime:  timeToUnixNano(t.LastHeartbeatTime),
		Details:            ToPayload(t.Details),
		Attempt:            t.Attempt,
		LastFailureReason:  ToFailureReason(t.LastFailure),
		LastFailureDetails: ToFailureDetails(t.LastFailure),
		LastWorkerIdentity: t.LastWorkerIdentity,
		VersionHistory:     ToVersionHistory(t.VersionHistory),
	}
}

func FromSyncShardStatus(t *types.SyncShardStatus) *sharedv1.SyncShardStatus {
	if t == nil {
		return nil
	}
	return &sharedv1.SyncShardStatus{
		Timestamp: unixNanoToTime(t.Timestamp),
	}
}

func ToSyncShardStatus(t *sharedv1.SyncShardStatus) *types.SyncShardStatus {
	if t == nil {
		return nil
	}
	return &types.SyncShardStatus{
		Timestamp: timeToUnixNano(t.Timestamp),
	}
}

func FromSyncShardStatusTaskAttributes(t *types.SyncShardStatusTaskAttributes) *sharedv1.SyncShardStatusTaskAttributes {
	if t == nil {
		return nil
	}
	return &sharedv1.SyncShardStatusTaskAttributes{
		SourceCluster: t.SourceCluster,
		ShardId:       int32(t.ShardID),
		Timestamp:     unixNanoToTime(t.Timestamp),
	}
}

func ToSyncShardStatusTaskAttributes(t *sharedv1.SyncShardStatusTaskAttributes) *types.SyncShardStatusTaskAttributes {
	if t == nil {
		return nil
	}
	return &types.SyncShardStatusTaskAttributes{
		SourceCluster: t.SourceCluster,
		ShardID:       int64(t.ShardId),
		Timestamp:     timeToUnixNano(t.Timestamp),
	}
}

func FromReplicationTaskInfoArray(t []*types.ReplicationTaskInfo) []*sharedv1.ReplicationTaskInfo {
	if t == nil {
		return nil
	}
	v := make([]*sharedv1.ReplicationTaskInfo, len(t))
	for i := range t {
		v[i] = FromReplicationTaskInfo(t[i])
	}
	return v
}

func ToReplicationTaskInfoArray(t []*sharedv1.ReplicationTaskInfo) []*types.ReplicationTaskInfo {
	if t == nil {
		return nil
	}
	v := make([]*types.ReplicationTaskInfo, len(t))
	for i := range t {
		v[i] = ToReplicationTaskInfo(t[i])
	}
	return v
}

func FromReplicationTaskArray(t []*types.ReplicationTask) []*sharedv1.ReplicationTask {
	if t == nil {
		return nil
	}
	v := make([]*sharedv1.ReplicationTask, len(t))
	for i := range t {
		v[i] = FromReplicationTask(t[i])
	}
	return v
}

func ToReplicationTaskArray(t []*sharedv1.ReplicationTask) []*types.ReplicationTask {
	if t == nil {
		return nil
	}
	v := make([]*types.ReplicationTask, len(t))
	for i := range t {
		v[i] = ToReplicationTask(t[i])
	}
	return v
}

func FromReplicationTokenArray(t []*types.ReplicationToken) []*sharedv1.ReplicationToken {
	if t == nil {
		return nil
	}
	v := make([]*sharedv1.ReplicationToken, len(t))
	for i := range t {
		v[i] = FromReplicationToken(t[i])
	}
	return v
}

func ToReplicationTokenArray(t []*sharedv1.ReplicationToken) []*types.ReplicationToken {
	if t == nil {
		return nil
	}
	v := make([]*types.ReplicationToken, len(t))
	for i := range t {
		v[i] = ToReplicationToken(t[i])
	}
	return v
}

func FromReplicationMessagesMap(t map[int32]*types.ReplicationMessages) map[int32]*sharedv1.ReplicationMessages {
	if t == nil {
		return nil
	}
	v := make(map[int32]*sharedv1.ReplicationMessages, len(t))
	for key := range t {
		v[key] = FromReplicationMessages(t[key])
	}
	return v
}

func ToReplicationMessagesMap(t map[int32]*sharedv1.ReplicationMessages) map[int32]*types.ReplicationMessages {
	if t == nil {
		return nil
	}
	v := make(map[int32]*types.ReplicationMessages, len(t))
	for key := range t {
		v[key] = ToReplicationMessages(t[key])
	}
	return v
}

func FromReplicationTask(t *types.ReplicationTask) *sharedv1.ReplicationTask {
	if t == nil {
		return nil
	}
	task := sharedv1.ReplicationTask{
		TaskType:     FromReplicationTaskType(t.TaskType),
		SourceTaskId: t.SourceTaskID,
		CreationTime: unixNanoToTime(t.CreationTime),
	}
	if t.DomainTaskAttributes != nil {
		task.Attributes = &sharedv1.ReplicationTask_DomainTaskAttributes{
			DomainTaskAttributes: FromDomainTaskAttributes(t.DomainTaskAttributes),
		}
	}
	if t.SyncShardStatusTaskAttributes != nil {
		task.Attributes = &sharedv1.ReplicationTask_SyncShardStatusTaskAttributes{
			SyncShardStatusTaskAttributes: FromSyncShardStatusTaskAttributes(t.SyncShardStatusTaskAttributes),
		}
	}
	if t.SyncActivityTaskAttributes != nil {
		task.Attributes = &sharedv1.ReplicationTask_SyncActivityTaskAttributes{
			SyncActivityTaskAttributes: FromSyncActivityTaskAttributes(t.SyncActivityTaskAttributes),
		}
	}
	if t.HistoryTaskV2Attributes != nil {
		task.Attributes = &sharedv1.ReplicationTask_HistoryTaskV2Attributes{
			HistoryTaskV2Attributes: FromHistoryTaskV2Attributes(t.HistoryTaskV2Attributes),
		}
	}
	if t.FailoverMarkerAttributes != nil {
		task.Attributes = &sharedv1.ReplicationTask_FailoverMarkerAttributes{
			FailoverMarkerAttributes: FromFailoverMarkerAttributes(t.FailoverMarkerAttributes),
		}
	}

	return &task
}

func ToReplicationTask(t *sharedv1.ReplicationTask) *types.ReplicationTask {
	if t == nil {
		return nil
	}
	task := types.ReplicationTask{
		TaskType:     ToReplicationTaskType(t.TaskType),
		SourceTaskID: t.SourceTaskId,
		CreationTime: timeToUnixNano(t.CreationTime),
	}

	switch attr := t.Attributes.(type) {
	case *sharedv1.ReplicationTask_DomainTaskAttributes:
		task.DomainTaskAttributes = ToDomainTaskAttributes(attr.DomainTaskAttributes)
	case *sharedv1.ReplicationTask_SyncShardStatusTaskAttributes:
		task.SyncShardStatusTaskAttributes = ToSyncShardStatusTaskAttributes(attr.SyncShardStatusTaskAttributes)
	case *sharedv1.ReplicationTask_SyncActivityTaskAttributes:
		task.SyncActivityTaskAttributes = ToSyncActivityTaskAttributes(attr.SyncActivityTaskAttributes)
	case *sharedv1.ReplicationTask_HistoryTaskV2Attributes:
		task.HistoryTaskV2Attributes = ToHistoryTaskV2Attributes(attr.HistoryTaskV2Attributes)
	case *sharedv1.ReplicationTask_FailoverMarkerAttributes:
		task.FailoverMarkerAttributes = ToFailoverMarkerAttributes(attr.FailoverMarkerAttributes)
	}
	return &task
}

func FromTaskType(t *int32) sharedv1.TaskType {
	if t == nil {
		return sharedv1.TaskType_TASK_TYPE_INVALID
	}
	switch common.TaskType(*t) {
	case common.TaskTypeTransfer:
		return sharedv1.TaskType_TASK_TYPE_TRANSFER
	case common.TaskTypeTimer:
		return sharedv1.TaskType_TASK_TYPE_TIMER
	case common.TaskTypeReplication:
		return sharedv1.TaskType_TASK_TYPE_REPLICATION
	case common.TaskTypeCrossCluster:
		return sharedv1.TaskType_TASK_TYPE_CROSS_CLUSTER
	}
	panic("unexpected enum value")
}

func ToTaskType(t sharedv1.TaskType) *int32 {
	switch t {
	case sharedv1.TaskType_TASK_TYPE_INVALID:
		return nil
	case sharedv1.TaskType_TASK_TYPE_TRANSFER:
		return common.Int32Ptr(int32(common.TaskTypeTransfer))
	case sharedv1.TaskType_TASK_TYPE_TIMER:
		return common.Int32Ptr(int32(common.TaskTypeTimer))
	case sharedv1.TaskType_TASK_TYPE_REPLICATION:
		return common.Int32Ptr(int32(common.TaskTypeReplication))
	case sharedv1.TaskType_TASK_TYPE_CROSS_CLUSTER:
		return common.Int32Ptr(int32(common.TaskTypeCrossCluster))
	}
	panic("unexpected enum value")
}

func FromFailoverMarkerTokenArray(t []*types.FailoverMarkerToken) []*sharedv1.FailoverMarkerToken {
	if t == nil {
		return nil
	}
	v := make([]*sharedv1.FailoverMarkerToken, len(t))
	for i := range t {
		v[i] = FromFailoverMarkerToken(t[i])
	}
	return v
}

func ToFailoverMarkerTokenArray(t []*sharedv1.FailoverMarkerToken) []*types.FailoverMarkerToken {
	if t == nil {
		return nil
	}
	v := make([]*types.FailoverMarkerToken, len(t))
	for i := range t {
		v[i] = ToFailoverMarkerToken(t[i])
	}
	return v
}

func FromVersionHistoryItemArray(t []*types.VersionHistoryItem) []*sharedv1.VersionHistoryItem {
	if t == nil {
		return nil
	}
	v := make([]*sharedv1.VersionHistoryItem, len(t))
	for i := range t {
		v[i] = FromVersionHistoryItem(t[i])
	}
	return v
}

func ToVersionHistoryItemArray(t []*sharedv1.VersionHistoryItem) []*types.VersionHistoryItem {
	if t == nil {
		return nil
	}
	v := make([]*types.VersionHistoryItem, len(t))
	for i := range t {
		v[i] = ToVersionHistoryItem(t[i])
	}
	return v
}

func FromEventIDVersionPair(id, version *int64) *sharedv1.VersionHistoryItem {
	if id == nil || version == nil {
		return nil
	}
	return &sharedv1.VersionHistoryItem{
		EventId: *id,
		Version: *version,
	}
}

func ToEventID(item *sharedv1.VersionHistoryItem) *int64 {
	if item == nil {
		return nil
	}
	return common.Int64Ptr(item.EventId)
}

func ToEventVersion(item *sharedv1.VersionHistoryItem) *int64 {
	if item == nil {
		return nil
	}
	return common.Int64Ptr(item.Version)
}

// FromCrossClusterTaskType converts internal CrossClusterTaskType type to proto
func FromCrossClusterTaskType(t *types.CrossClusterTaskType) sharedv1.CrossClusterTaskType {
	if t == nil {
		return sharedv1.CrossClusterTaskType_CROSS_CLUSTER_TASK_TYPE_INVALID
	}
	switch *t {
	case types.CrossClusterTaskTypeStartChildExecution:
		return sharedv1.CrossClusterTaskType_CROSS_CLUSTER_TASK_TYPE_START_CHILD_EXECUTION
	case types.CrossClusterTaskTypeCancelExecution:
		return sharedv1.CrossClusterTaskType_CROSS_CLUSTER_TASK_TYPE_CANCEL_EXECUTION
	case types.CrossClusterTaskTypeSignalExecution:
		return sharedv1.CrossClusterTaskType_CROSS_CLUSTER_TASK_TYPE_SIGNAL_EXECUTION
	case types.CrossClusterTaskTypeRecordChildWorkflowExeuctionComplete:
		return sharedv1.CrossClusterTaskType_CROSS_CLUSTER_TASK_TYPE_RECORD_CHILD_WORKKLOW_EXECUTION_COMPLETE
	case types.CrossClusterTaskTypeApplyParentPolicy:
		return sharedv1.CrossClusterTaskType_CROSS_CLUSTER_TASK_TYPE_APPLY_PARENT_CLOSE_POLICY
	}
	panic("unexpected enum value")
}

// ToCrossClusterTaskType converts proto CrossClusterTaskType type to internal
func ToCrossClusterTaskType(t sharedv1.CrossClusterTaskType) *types.CrossClusterTaskType {
	switch t {
	case sharedv1.CrossClusterTaskType_CROSS_CLUSTER_TASK_TYPE_INVALID:
		return nil
	case sharedv1.CrossClusterTaskType_CROSS_CLUSTER_TASK_TYPE_START_CHILD_EXECUTION:
		return types.CrossClusterTaskTypeStartChildExecution.Ptr()
	case sharedv1.CrossClusterTaskType_CROSS_CLUSTER_TASK_TYPE_CANCEL_EXECUTION:
		return types.CrossClusterTaskTypeCancelExecution.Ptr()
	case sharedv1.CrossClusterTaskType_CROSS_CLUSTER_TASK_TYPE_SIGNAL_EXECUTION:
		return types.CrossClusterTaskTypeSignalExecution.Ptr()
	case sharedv1.CrossClusterTaskType_CROSS_CLUSTER_TASK_TYPE_RECORD_CHILD_WORKKLOW_EXECUTION_COMPLETE:
		return types.CrossClusterTaskTypeRecordChildWorkflowExeuctionComplete.Ptr()
	case sharedv1.CrossClusterTaskType_CROSS_CLUSTER_TASK_TYPE_APPLY_PARENT_CLOSE_POLICY:
		return types.CrossClusterTaskTypeApplyParentPolicy.Ptr()
	}
	panic("unexpected enum value")
}

// FromCrossClusterTaskFailedCause converts internal CrossClusterTaskFailedCause type to proto
func FromCrossClusterTaskFailedCause(t *types.CrossClusterTaskFailedCause) sharedv1.CrossClusterTaskFailedCause {
	if t == nil {
		return sharedv1.CrossClusterTaskFailedCause_CROSS_CLUSTER_TASK_FAILED_CAUSE_INVALID
	}
	switch *t {
	case types.CrossClusterTaskFailedCauseDomainNotActive:
		return sharedv1.CrossClusterTaskFailedCause_CROSS_CLUSTER_TASK_FAILED_CAUSE_DOMAIN_NOT_ACTIVE
	case types.CrossClusterTaskFailedCauseDomainNotExists:
		return sharedv1.CrossClusterTaskFailedCause_CROSS_CLUSTER_TASK_FAILED_CAUSE_DOMAIN_NOT_EXISTS
	case types.CrossClusterTaskFailedCauseWorkflowAlreadyRunning:
		return sharedv1.CrossClusterTaskFailedCause_CROSS_CLUSTER_TASK_FAILED_CAUSE_WORKFLOW_ALREADY_RUNNING
	case types.CrossClusterTaskFailedCauseWorkflowNotExists:
		return sharedv1.CrossClusterTaskFailedCause_CROSS_CLUSTER_TASK_FAILED_CAUSE_WORKFLOW_NOT_EXISTS
	case types.CrossClusterTaskFailedCauseWorkflowAlreadyCompleted:
		return sharedv1.CrossClusterTaskFailedCause_CROSS_CLUSTER_TASK_FAILED_CAUSE_WORKFLOW_ALREADY_COMPLETED
	case types.CrossClusterTaskFailedCauseUncategorized:
		return sharedv1.CrossClusterTaskFailedCause_CROSS_CLUSTER_TASK_FAILED_CAUSE_UNCATEGORIZED
	}
	panic("unexpected enum value")
}

// ToCrossClusterTaskFailedCause converts proto CrossClusterTaskFailedCause type to internal
func ToCrossClusterTaskFailedCause(t sharedv1.CrossClusterTaskFailedCause) *types.CrossClusterTaskFailedCause {
	switch t {
	case sharedv1.CrossClusterTaskFailedCause_CROSS_CLUSTER_TASK_FAILED_CAUSE_INVALID:
		return nil
	case sharedv1.CrossClusterTaskFailedCause_CROSS_CLUSTER_TASK_FAILED_CAUSE_DOMAIN_NOT_ACTIVE:
		return types.CrossClusterTaskFailedCauseDomainNotActive.Ptr()
	case sharedv1.CrossClusterTaskFailedCause_CROSS_CLUSTER_TASK_FAILED_CAUSE_DOMAIN_NOT_EXISTS:
		return types.CrossClusterTaskFailedCauseDomainNotExists.Ptr()
	case sharedv1.CrossClusterTaskFailedCause_CROSS_CLUSTER_TASK_FAILED_CAUSE_WORKFLOW_ALREADY_RUNNING:
		return types.CrossClusterTaskFailedCauseWorkflowAlreadyRunning.Ptr()
	case sharedv1.CrossClusterTaskFailedCause_CROSS_CLUSTER_TASK_FAILED_CAUSE_WORKFLOW_NOT_EXISTS:
		return types.CrossClusterTaskFailedCauseWorkflowNotExists.Ptr()
	case sharedv1.CrossClusterTaskFailedCause_CROSS_CLUSTER_TASK_FAILED_CAUSE_WORKFLOW_ALREADY_COMPLETED:
		return types.CrossClusterTaskFailedCauseWorkflowAlreadyCompleted.Ptr()
	case sharedv1.CrossClusterTaskFailedCause_CROSS_CLUSTER_TASK_FAILED_CAUSE_UNCATEGORIZED:
		return types.CrossClusterTaskFailedCauseUncategorized.Ptr()

	}
	panic("unexpected enum value")
}

// FromGetTaskFailedCause converts internal GetTaskFailedCause type to proto
func FromGetTaskFailedCause(t *types.GetTaskFailedCause) sharedv1.GetTaskFailedCause {
	if t == nil {
		return sharedv1.GetTaskFailedCause_GET_TASK_FAILED_CAUSE_INVALID
	}
	switch *t {
	case types.GetTaskFailedCauseServiceBusy:
		return sharedv1.GetTaskFailedCause_GET_TASK_FAILED_CAUSE_SERVICE_BUSY
	case types.GetTaskFailedCauseTimeout:
		return sharedv1.GetTaskFailedCause_GET_TASK_FAILED_CAUSE_TIMEOUT
	case types.GetTaskFailedCauseShardOwnershipLost:
		return sharedv1.GetTaskFailedCause_GET_TASK_FAILED_CAUSE_SHARD_OWNERSHIP_LOST
	case types.GetTaskFailedCauseUncategorized:
		return sharedv1.GetTaskFailedCause_GET_TASK_FAILED_CAUSE_UNCATEGORIZED
	}
	panic("unexpected enum value")
}

// ToGetTaskFailedCause converts proto GetTaskFailedCause type to internal
func ToGetTaskFailedCause(t sharedv1.GetTaskFailedCause) *types.GetTaskFailedCause {
	switch t {
	case sharedv1.GetTaskFailedCause_GET_TASK_FAILED_CAUSE_INVALID:
		return nil
	case sharedv1.GetTaskFailedCause_GET_TASK_FAILED_CAUSE_SERVICE_BUSY:
		return types.GetTaskFailedCauseServiceBusy.Ptr()
	case sharedv1.GetTaskFailedCause_GET_TASK_FAILED_CAUSE_TIMEOUT:
		return types.GetTaskFailedCauseTimeout.Ptr()
	case sharedv1.GetTaskFailedCause_GET_TASK_FAILED_CAUSE_SHARD_OWNERSHIP_LOST:
		return types.GetTaskFailedCauseShardOwnershipLost.Ptr()
	case sharedv1.GetTaskFailedCause_GET_TASK_FAILED_CAUSE_UNCATEGORIZED:
		return types.GetTaskFailedCauseUncategorized.Ptr()
	}
	panic("unexpected enum value")
}

// FromCrossClusterTaskInfo converts internal CrossClusterTaskInfo type to proto
func FromCrossClusterTaskInfo(t *types.CrossClusterTaskInfo) *sharedv1.CrossClusterTaskInfo {
	if t == nil {
		return nil
	}
	return &sharedv1.CrossClusterTaskInfo{
		DomainId:            t.DomainID,
		WorkflowExecution:   FromWorkflowRunPair(t.WorkflowID, t.RunID),
		TaskType:            FromCrossClusterTaskType(t.TaskType),
		TaskState:           int32(t.TaskState),
		TaskId:              t.TaskID,
		VisibilityTimestamp: unixNanoToTime(t.VisibilityTimestamp),
	}
}

// ToCrossClusterTaskInfo converts proto CrossClusterTaskInfo type to internal
func ToCrossClusterTaskInfo(t *sharedv1.CrossClusterTaskInfo) *types.CrossClusterTaskInfo {
	if t == nil {
		return nil
	}
	return &types.CrossClusterTaskInfo{
		DomainID:            t.DomainId,
		WorkflowID:          ToWorkflowID(t.WorkflowExecution),
		RunID:               ToRunID(t.WorkflowExecution),
		TaskType:            ToCrossClusterTaskType(t.TaskType),
		TaskState:           int16(t.TaskState),
		TaskID:              t.TaskId,
		VisibilityTimestamp: timeToUnixNano(t.VisibilityTimestamp),
	}
}

// FromCrossClusterStartChildExecutionRequestAttributes converts internal CrossClusterStartChildExecutionRequestAttributes type to proto
func FromCrossClusterStartChildExecutionRequestAttributes(t *types.CrossClusterStartChildExecutionRequestAttributes) *sharedv1.CrossClusterStartChildExecutionRequestAttributes {
	if t == nil {
		return nil
	}
	return &sharedv1.CrossClusterStartChildExecutionRequestAttributes{
		TargetDomainId:           t.TargetDomainID,
		RequestId:                t.RequestID,
		InitiatedEventId:         t.InitiatedEventID,
		InitiatedEventAttributes: FromStartChildWorkflowExecutionInitiatedEventAttributes(t.InitiatedEventAttributes),
		TargetRunId:              t.GetTargetRunID(),
	}
}

// ToCrossClusterStartChildExecutionRequestAttributes converts proto CrossClusterStartChildExecutionRequestAttributes type to internal
func ToCrossClusterStartChildExecutionRequestAttributes(t *sharedv1.CrossClusterStartChildExecutionRequestAttributes) *types.CrossClusterStartChildExecutionRequestAttributes {
	if t == nil {
		return nil
	}
	return &types.CrossClusterStartChildExecutionRequestAttributes{
		TargetDomainID:           t.TargetDomainId,
		RequestID:                t.RequestId,
		InitiatedEventID:         t.InitiatedEventId,
		InitiatedEventAttributes: ToStartChildWorkflowExecutionInitiatedEventAttributes(t.InitiatedEventAttributes),
		TargetRunID:              &t.TargetRunId,
	}
}

// FromCrossClusterStartChildExecutionResponseAttributes converts internal CrossClusterStartChildExecutionResponseAttributes type to proto
func FromCrossClusterStartChildExecutionResponseAttributes(t *types.CrossClusterStartChildExecutionResponseAttributes) *sharedv1.CrossClusterStartChildExecutionResponseAttributes {
	if t == nil {
		return nil
	}
	return &sharedv1.CrossClusterStartChildExecutionResponseAttributes{
		RunId: t.RunID,
	}
}

// ToCrossClusterStartChildExecutionResponseAttributes converts proto CrossClusterStartChildExecutionResponseAttributes type to internal
func ToCrossClusterStartChildExecutionResponseAttributes(t *sharedv1.CrossClusterStartChildExecutionResponseAttributes) *types.CrossClusterStartChildExecutionResponseAttributes {
	if t == nil {
		return nil
	}
	return &types.CrossClusterStartChildExecutionResponseAttributes{
		RunID: t.RunId,
	}
}

// FromCrossClusterCancelExecutionRequestAttributes converts internal CrossClusterCancelExecutionRequestAttributes type to proto
func FromCrossClusterCancelExecutionRequestAttributes(t *types.CrossClusterCancelExecutionRequestAttributes) *sharedv1.CrossClusterCancelExecutionRequestAttributes {
	if t == nil {
		return nil
	}
	return &sharedv1.CrossClusterCancelExecutionRequestAttributes{
		TargetDomainId:          t.TargetDomainID,
		TargetWorkflowExecution: FromWorkflowRunPair(t.TargetWorkflowID, t.TargetRunID),
		RequestId:               t.RequestID,
		InitiatedEventId:        t.InitiatedEventID,
		ChildWorkflowOnly:       t.ChildWorkflowOnly,
	}
}

// ToCrossClusterCancelExecutionRequestAttributes converts proto CrossClusterCancelExecutionRequestAttributes type to internal
func ToCrossClusterCancelExecutionRequestAttributes(t *sharedv1.CrossClusterCancelExecutionRequestAttributes) *types.CrossClusterCancelExecutionRequestAttributes {
	if t == nil {
		return nil
	}
	return &types.CrossClusterCancelExecutionRequestAttributes{
		TargetDomainID:    t.TargetDomainId,
		TargetWorkflowID:  ToWorkflowID(t.TargetWorkflowExecution),
		TargetRunID:       ToRunID(t.TargetWorkflowExecution),
		RequestID:         t.RequestId,
		InitiatedEventID:  t.InitiatedEventId,
		ChildWorkflowOnly: t.ChildWorkflowOnly,
	}
}

// FromCrossClusterCancelExecutionResponseAttributes converts internal CrossClusterCancelExecutionResponseAttributes type to proto
func FromCrossClusterCancelExecutionResponseAttributes(t *types.CrossClusterCancelExecutionResponseAttributes) *sharedv1.CrossClusterCancelExecutionResponseAttributes {
	if t == nil {
		return nil
	}
	return &sharedv1.CrossClusterCancelExecutionResponseAttributes{}
}

// ToCrossClusterCancelExecutionResponseAttributes converts proto CrossClusterCancelExecutionResponseAttributes type to internal
func ToCrossClusterCancelExecutionResponseAttributes(t *sharedv1.CrossClusterCancelExecutionResponseAttributes) *types.CrossClusterCancelExecutionResponseAttributes {
	if t == nil {
		return nil
	}
	return &types.CrossClusterCancelExecutionResponseAttributes{}
}

// FromCrossClusterSignalExecutionRequestAttributes converts internal CrossClusterSignalExecutionRequestAttributes type to proto
func FromCrossClusterSignalExecutionRequestAttributes(t *types.CrossClusterSignalExecutionRequestAttributes) *sharedv1.CrossClusterSignalExecutionRequestAttributes {
	if t == nil {
		return nil
	}
	return &sharedv1.CrossClusterSignalExecutionRequestAttributes{
		TargetDomainId:          t.TargetDomainID,
		TargetWorkflowExecution: FromWorkflowRunPair(t.TargetWorkflowID, t.TargetRunID),
		RequestId:               t.RequestID,
		InitiatedEventId:        t.InitiatedEventID,
		ChildWorkflowOnly:       t.ChildWorkflowOnly,
		SignalName:              t.SignalName,
		SignalInput:             FromPayload(t.SignalInput),
		Control:                 t.Control,
	}
}

// ToCrossClusterSignalExecutionRequestAttributes converts proto CrossClusterSignalExecutionRequestAttributes type to internal
func ToCrossClusterSignalExecutionRequestAttributes(t *sharedv1.CrossClusterSignalExecutionRequestAttributes) *types.CrossClusterSignalExecutionRequestAttributes {
	if t == nil {
		return nil
	}
	return &types.CrossClusterSignalExecutionRequestAttributes{
		TargetDomainID:    t.TargetDomainId,
		TargetWorkflowID:  ToWorkflowID(t.TargetWorkflowExecution),
		TargetRunID:       ToRunID(t.TargetWorkflowExecution),
		RequestID:         t.RequestId,
		InitiatedEventID:  t.InitiatedEventId,
		ChildWorkflowOnly: t.ChildWorkflowOnly,
		SignalName:        t.SignalName,
		SignalInput:       ToPayload(t.SignalInput),
		Control:           t.Control,
	}
}

// FromCrossClusterSignalExecutionResponseAttributes converts internal CrossClusterSignalExecutionResponseAttributes type to proto
func FromCrossClusterSignalExecutionResponseAttributes(t *types.CrossClusterSignalExecutionResponseAttributes) *sharedv1.CrossClusterSignalExecutionResponseAttributes {
	if t == nil {
		return nil
	}
	return &sharedv1.CrossClusterSignalExecutionResponseAttributes{}
}

// ToCrossClusterSignalExecutionResponseAttributes converts proto CrossClusterSignalExecutionResponseAttributes type to internal
func ToCrossClusterSignalExecutionResponseAttributes(t *sharedv1.CrossClusterSignalExecutionResponseAttributes) *types.CrossClusterSignalExecutionResponseAttributes {
	if t == nil {
		return nil
	}
	return &types.CrossClusterSignalExecutionResponseAttributes{}
}

// FromCrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes converts internal CrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes type to proto
func FromCrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes(t *types.CrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes) *sharedv1.CrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes {
	if t == nil {
		return nil
	}
	return &sharedv1.CrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes{
		TargetDomainId:          t.TargetDomainID,
		TargetWorkflowExecution: FromWorkflowRunPair(t.TargetWorkflowID, t.TargetRunID),
		InitiatedEventId:        t.InitiatedEventID,
		CompletionEvent:         FromHistoryEvent(t.CompletionEvent),
	}
}

// ToCrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes converts proto CrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes type to internal
func ToCrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes(t *sharedv1.CrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes) *types.CrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes {
	if t == nil {
		return nil
	}
	return &types.CrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes{
		TargetDomainID:   t.TargetDomainId,
		TargetWorkflowID: ToWorkflowID(t.TargetWorkflowExecution),
		TargetRunID:      ToRunID(t.TargetWorkflowExecution),
		InitiatedEventID: t.InitiatedEventId,
		CompletionEvent:  ToHistoryEvent(t.CompletionEvent),
	}
}

// FromCrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes converts internal CrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes type to proto
func FromCrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes(t *types.CrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes) *sharedv1.CrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes {
	if t == nil {
		return nil
	}
	return &sharedv1.CrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes{}
}

// ToCrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes converts proto CrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes type to internal
func ToCrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes(t *sharedv1.CrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes) *types.CrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes {
	if t == nil {
		return nil
	}
	return &types.CrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes{}
}

// FromApplyParentClosePolicyStatus converts internal ApplyParentClosePolicyStatus type to proto
func FromApplyParentClosePolicyStatus(t *types.ApplyParentClosePolicyStatus) *sharedv1.ApplyParentClosePolicyStatus {
	if t == nil {
		return nil
	}
	return &sharedv1.ApplyParentClosePolicyStatus{
		Completed:   t.Completed,
		FailedCause: FromCrossClusterTaskFailedCause(t.FailedCause),
	}
}

// ToApplyParentClosePolicyStatus converts proto ApplyParentClosePolicyStatus type to internal
func ToApplyParentClosePolicyStatus(t *sharedv1.ApplyParentClosePolicyStatus) *types.ApplyParentClosePolicyStatus {
	if t == nil {
		return nil
	}
	return &types.ApplyParentClosePolicyStatus{
		Completed:   t.Completed,
		FailedCause: ToCrossClusterTaskFailedCause(t.FailedCause),
	}
}

// FromApplyParentClosePolicyAttributes converts internal ApplyParentClosePolicyAttributes type to proto
func FromApplyParentClosePolicyAttributes(t *types.ApplyParentClosePolicyAttributes) *sharedv1.ApplyParentClosePolicyAttributes {
	if t == nil {
		return nil
	}
	return &sharedv1.ApplyParentClosePolicyAttributes{
		ChildDomainId:     t.ChildDomainID,
		ChildWorkflowId:   t.ChildWorkflowID,
		ChildRunId:        t.ChildRunID,
		ParentClosePolicy: FromParentClosePolicy(t.ParentClosePolicy),
	}
}

// ToApplyParentClosePolicyAttributes converts proto ApplyParentClosePolicyAttributes type to internal
func ToApplyParentClosePolicyAttributes(t *sharedv1.ApplyParentClosePolicyAttributes) *types.ApplyParentClosePolicyAttributes {
	if t == nil {
		return nil
	}
	return &types.ApplyParentClosePolicyAttributes{
		ChildDomainID:     t.ChildDomainId,
		ChildWorkflowID:   t.ChildWorkflowId,
		ChildRunID:        t.ChildRunId,
		ParentClosePolicy: ToParentClosePolicy(t.ParentClosePolicy),
	}
}

// FromApplyParentClosePolicyResult converts proto ApplyParentClosePolicyResult type to internal
func FromApplyParentClosePolicyResult(t *types.ApplyParentClosePolicyResult) *sharedv1.ApplyParentClosePolicyResult {
	if t == nil {
		return nil
	}
	return &sharedv1.ApplyParentClosePolicyResult{
		Child:       FromApplyParentClosePolicyAttributes(t.Child),
		FailedCause: FromCrossClusterTaskFailedCause(t.FailedCause),
	}
}

// ToApplyParentClosePolicyResult converts proto ApplyParentClosePolicyResult type to internal
func ToApplyParentClosePolicyResult(t *sharedv1.ApplyParentClosePolicyResult) *types.ApplyParentClosePolicyResult {
	if t == nil {
		return nil
	}
	return &types.ApplyParentClosePolicyResult{
		Child:       ToApplyParentClosePolicyAttributes(t.Child),
		FailedCause: ToCrossClusterTaskFailedCause(t.FailedCause),
	}
}

// FromCrossClusterApplyParentClosePolicyRequestAttributes converts internal CrossClusterApplyParentClosePolicyRequestAttributes type to proto
func FromCrossClusterApplyParentClosePolicyRequestAttributes(t *types.CrossClusterApplyParentClosePolicyRequestAttributes) *sharedv1.CrossClusterApplyParentClosePolicyRequestAttributes {
	if t == nil {
		return nil
	}
	requestAttributes := &sharedv1.CrossClusterApplyParentClosePolicyRequestAttributes{}
	for _, execution := range t.Children {
		requestAttributes.Children = append(
			requestAttributes.Children,
			&sharedv1.ApplyParentClosePolicyRequest{
				Child:  FromApplyParentClosePolicyAttributes(execution.Child),
				Status: FromApplyParentClosePolicyStatus(execution.Status),
			},
		)
	}
	return requestAttributes
}

// ToCrossClusterApplyParentClosePolicyRequestAttributes converts proto CrossClusterApplyParentClosePolicyRequestAttributes type to internal
func ToCrossClusterApplyParentClosePolicyRequestAttributes(t *sharedv1.CrossClusterApplyParentClosePolicyRequestAttributes) *types.CrossClusterApplyParentClosePolicyRequestAttributes {
	if t == nil {
		return nil
	}
	requestAttributes := &types.CrossClusterApplyParentClosePolicyRequestAttributes{}
	for _, execution := range t.Children {
		requestAttributes.Children = append(
			requestAttributes.Children,
			&types.ApplyParentClosePolicyRequest{
				Child:  ToApplyParentClosePolicyAttributes(execution.Child),
				Status: ToApplyParentClosePolicyStatus(execution.Status),
			},
		)
	}
	return requestAttributes
}

// FromCrossClusterApplyParentClosePolicyResponseAttributes converts internal CrossClusterApplyParentClosePolicyResponseAttributes type to proto
func FromCrossClusterApplyParentClosePolicyResponseAttributes(t *types.CrossClusterApplyParentClosePolicyResponseAttributes) *sharedv1.CrossClusterApplyParentClosePolicyResponseAttributes {
	if t == nil {
		return nil
	}
	response := &sharedv1.CrossClusterApplyParentClosePolicyResponseAttributes{}
	for _, childStatus := range t.ChildrenStatus {
		response.ChildrenStatus = append(
			response.ChildrenStatus,
			FromApplyParentClosePolicyResult(childStatus),
		)
	}
	return response
}

// ToCrossClusterApplyParentClosePolicyResponseAttributes converts proto CrossClusterApplyParentClosePolicyResponseAttributes type to internal
func ToCrossClusterApplyParentClosePolicyResponseAttributes(t *sharedv1.CrossClusterApplyParentClosePolicyResponseAttributes) *types.CrossClusterApplyParentClosePolicyResponseAttributes {
	if t == nil {
		return nil
	}
	response := &types.CrossClusterApplyParentClosePolicyResponseAttributes{}
	for _, childStatus := range t.ChildrenStatus {
		response.ChildrenStatus = append(
			response.ChildrenStatus,
			ToApplyParentClosePolicyResult(childStatus),
		)
	}
	return response
}

// FromCrossClusterTaskRequest converts internal CrossClusterTaskRequest type to proto
func FromCrossClusterTaskRequest(t *types.CrossClusterTaskRequest) *sharedv1.CrossClusterTaskRequest {
	if t == nil {
		return nil
	}
	request := sharedv1.CrossClusterTaskRequest{
		TaskInfo: FromCrossClusterTaskInfo(t.TaskInfo),
	}
	if t.StartChildExecutionAttributes != nil {
		request.Attributes = &sharedv1.CrossClusterTaskRequest_StartChildExecutionAttributes{
			StartChildExecutionAttributes: FromCrossClusterStartChildExecutionRequestAttributes(t.StartChildExecutionAttributes),
		}
	}
	if t.CancelExecutionAttributes != nil {
		request.Attributes = &sharedv1.CrossClusterTaskRequest_CancelExecutionAttributes{
			CancelExecutionAttributes: FromCrossClusterCancelExecutionRequestAttributes(t.CancelExecutionAttributes),
		}
	}
	if t.SignalExecutionAttributes != nil {
		request.Attributes = &sharedv1.CrossClusterTaskRequest_SignalExecutionAttributes{
			SignalExecutionAttributes: FromCrossClusterSignalExecutionRequestAttributes(t.SignalExecutionAttributes),
		}
	}
	if t.RecordChildWorkflowExecutionCompleteAttributes != nil {
		request.Attributes = &sharedv1.CrossClusterTaskRequest_RecordChildWorkflowExecutionCompleteRequestAttributes{
			RecordChildWorkflowExecutionCompleteRequestAttributes: FromCrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes(t.RecordChildWorkflowExecutionCompleteAttributes),
		}
	}
	if t.ApplyParentClosePolicyAttributes != nil {
		request.Attributes = &sharedv1.CrossClusterTaskRequest_ApplyParentClosePolicyRequestAttributes{
			ApplyParentClosePolicyRequestAttributes: FromCrossClusterApplyParentClosePolicyRequestAttributes(t.ApplyParentClosePolicyAttributes),
		}
	}
	return &request
}

// ToCrossClusterTaskRequest converts proto CrossClusterTaskRequest type to internal
func ToCrossClusterTaskRequest(t *sharedv1.CrossClusterTaskRequest) *types.CrossClusterTaskRequest {
	if t == nil {
		return nil
	}
	request := types.CrossClusterTaskRequest{
		TaskInfo: ToCrossClusterTaskInfo(t.TaskInfo),
	}
	switch attr := t.Attributes.(type) {
	case *sharedv1.CrossClusterTaskRequest_StartChildExecutionAttributes:
		request.StartChildExecutionAttributes = ToCrossClusterStartChildExecutionRequestAttributes(attr.StartChildExecutionAttributes)
	case *sharedv1.CrossClusterTaskRequest_CancelExecutionAttributes:
		request.CancelExecutionAttributes = ToCrossClusterCancelExecutionRequestAttributes(attr.CancelExecutionAttributes)
	case *sharedv1.CrossClusterTaskRequest_SignalExecutionAttributes:
		request.SignalExecutionAttributes = ToCrossClusterSignalExecutionRequestAttributes(attr.SignalExecutionAttributes)
	case *sharedv1.CrossClusterTaskRequest_RecordChildWorkflowExecutionCompleteRequestAttributes:
		request.RecordChildWorkflowExecutionCompleteAttributes = ToCrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes(attr.RecordChildWorkflowExecutionCompleteRequestAttributes)
	case *sharedv1.CrossClusterTaskRequest_ApplyParentClosePolicyRequestAttributes:
		request.ApplyParentClosePolicyAttributes = ToCrossClusterApplyParentClosePolicyRequestAttributes(attr.ApplyParentClosePolicyRequestAttributes)
	}
	return &request
}

// FromCrossClusterTaskResponse converts internal CrossClusterTaskResponse type to proto
func FromCrossClusterTaskResponse(t *types.CrossClusterTaskResponse) *sharedv1.CrossClusterTaskResponse {
	if t == nil {
		return nil
	}
	response := sharedv1.CrossClusterTaskResponse{
		TaskId:      t.TaskID,
		TaskType:    FromCrossClusterTaskType(t.TaskType),
		TaskState:   int32(t.TaskState),
		FailedCause: FromCrossClusterTaskFailedCause(t.FailedCause),
	}
	if t.StartChildExecutionAttributes != nil {
		response.Attributes = &sharedv1.CrossClusterTaskResponse_StartChildExecutionAttributes{
			StartChildExecutionAttributes: FromCrossClusterStartChildExecutionResponseAttributes(t.StartChildExecutionAttributes),
		}
	}
	if t.CancelExecutionAttributes != nil {
		response.Attributes = &sharedv1.CrossClusterTaskResponse_CancelExecutionAttributes{
			CancelExecutionAttributes: FromCrossClusterCancelExecutionResponseAttributes(t.CancelExecutionAttributes),
		}
	}
	if t.SignalExecutionAttributes != nil {
		response.Attributes = &sharedv1.CrossClusterTaskResponse_SignalExecutionAttributes{
			SignalExecutionAttributes: FromCrossClusterSignalExecutionResponseAttributes(t.SignalExecutionAttributes),
		}
	}
	if t.RecordChildWorkflowExecutionCompleteAttributes != nil {
		response.Attributes = &sharedv1.CrossClusterTaskResponse_RecordChildWorkflowExecutionCompleteRequestAttributes{
			RecordChildWorkflowExecutionCompleteRequestAttributes: FromCrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes(t.RecordChildWorkflowExecutionCompleteAttributes),
		}
	}
	if t.ApplyParentClosePolicyAttributes != nil {
		response.Attributes = &sharedv1.CrossClusterTaskResponse_ApplyParentClosePolicyResponseAttributes{
			ApplyParentClosePolicyResponseAttributes: FromCrossClusterApplyParentClosePolicyResponseAttributes(t.ApplyParentClosePolicyAttributes),
		}
	}
	return &response
}

// ToCrossClusterTaskResponse converts proto CrossClusterTaskResponse type to internal
func ToCrossClusterTaskResponse(t *sharedv1.CrossClusterTaskResponse) *types.CrossClusterTaskResponse {
	if t == nil {
		return nil
	}
	response := types.CrossClusterTaskResponse{
		TaskID:      t.TaskId,
		TaskType:    ToCrossClusterTaskType(t.TaskType),
		TaskState:   int16(t.TaskState),
		FailedCause: ToCrossClusterTaskFailedCause(t.FailedCause),
	}
	switch attr := t.Attributes.(type) {
	case *sharedv1.CrossClusterTaskResponse_StartChildExecutionAttributes:
		response.StartChildExecutionAttributes = ToCrossClusterStartChildExecutionResponseAttributes(attr.StartChildExecutionAttributes)
	case *sharedv1.CrossClusterTaskResponse_CancelExecutionAttributes:
		response.CancelExecutionAttributes = ToCrossClusterCancelExecutionResponseAttributes(attr.CancelExecutionAttributes)
	case *sharedv1.CrossClusterTaskResponse_SignalExecutionAttributes:
		response.SignalExecutionAttributes = ToCrossClusterSignalExecutionResponseAttributes(attr.SignalExecutionAttributes)
	case *sharedv1.CrossClusterTaskResponse_RecordChildWorkflowExecutionCompleteRequestAttributes:
		response.RecordChildWorkflowExecutionCompleteAttributes = ToCrossClusterRecordChildWorkflowExecutionCompleteResponseAttributes(attr.RecordChildWorkflowExecutionCompleteRequestAttributes)
	case *sharedv1.CrossClusterTaskResponse_ApplyParentClosePolicyResponseAttributes:
		response.ApplyParentClosePolicyAttributes = ToCrossClusterApplyParentClosePolicyResponseAttributes(attr.ApplyParentClosePolicyResponseAttributes)
	}
	return &response
}

// FromCrossClusterTaskRequestArray converts internal CrossClusterTaskRequest type array to proto
func FromCrossClusterTaskRequestArray(t []*types.CrossClusterTaskRequest) *sharedv1.CrossClusterTaskRequests {
	if t == nil {
		return nil
	}
	v := make([]*sharedv1.CrossClusterTaskRequest, len(t))
	for i := range t {
		v[i] = FromCrossClusterTaskRequest(t[i])
	}
	return &sharedv1.CrossClusterTaskRequests{
		TaskRequests: v,
	}
}

// ToCrossClusterTaskRequestArray converts proto CrossClusterTaskRequest type array to internal
func ToCrossClusterTaskRequestArray(t *sharedv1.CrossClusterTaskRequests) []*types.CrossClusterTaskRequest {
	if t == nil || t.TaskRequests == nil {
		return nil
	}
	v := make([]*types.CrossClusterTaskRequest, len(t.TaskRequests))
	for i := range t.TaskRequests {
		v[i] = ToCrossClusterTaskRequest(t.TaskRequests[i])
	}
	return v
}

// FromCrossClusterTaskRequestMap converts internal CrossClusterTaskRequest type map to proto
func FromCrossClusterTaskRequestMap(t map[int32][]*types.CrossClusterTaskRequest) map[int32]*sharedv1.CrossClusterTaskRequests {
	if t == nil {
		return nil
	}
	v := make(map[int32]*sharedv1.CrossClusterTaskRequests, len(t))
	for key := range t {
		v[key] = FromCrossClusterTaskRequestArray(t[key])
	}
	return v
}

// ToCrossClusterTaskRequestMap converts proto CrossClusterTaskRequest type map to internal
func ToCrossClusterTaskRequestMap(t map[int32]*sharedv1.CrossClusterTaskRequests) map[int32][]*types.CrossClusterTaskRequest {
	if t == nil {
		return nil
	}
	v := make(map[int32][]*types.CrossClusterTaskRequest, len(t))
	for key := range t {
		value := ToCrossClusterTaskRequestArray(t[key])
		if value == nil {
			// grpc can't differentiate between empty array or nil array
			// our application logic ensure no nil array will be returned
			// for CrossClusterTaskRequest, so always convert to empty array
			// we only need the special handling here as this array is used
			// as a map value in GetCrossClusterTasksResponse,
			// and if the map value is nil, THRIFT won't be able to encode the value
			// this may happen when we are using grpc within a cluster but thrift across cluster
			value = []*types.CrossClusterTaskRequest{}
		}
		v[key] = value
	}
	return v
}

// FromGetTaskFailedCauseMap converts internal GetTaskFailedCause type map to proto
func FromGetTaskFailedCauseMap(t map[int32]types.GetTaskFailedCause) map[int32]sharedv1.GetTaskFailedCause {
	if t == nil {
		return nil
	}
	v := make(map[int32]sharedv1.GetTaskFailedCause, len(t))
	for key, value := range t {
		v[key] = FromGetTaskFailedCause(&value)
	}
	return v
}

// ToGetTaskFailedCauseMap converts proto GetTaskFailedCause type map to internal
func ToGetTaskFailedCauseMap(t map[int32]sharedv1.GetTaskFailedCause) map[int32]types.GetTaskFailedCause {
	if t == nil {
		return nil
	}
	v := make(map[int32]types.GetTaskFailedCause, len(t))
	for key := range t {
		if internalValue := ToGetTaskFailedCause(t[key]); internalValue != nil {
			v[key] = *internalValue
		}
	}
	return v
}

// FromCrossClusterTaskResponseArray converts internal CrossClusterTaskResponse type array to proto
func FromCrossClusterTaskResponseArray(t []*types.CrossClusterTaskResponse) []*sharedv1.CrossClusterTaskResponse {
	if t == nil {
		return nil
	}
	v := make([]*sharedv1.CrossClusterTaskResponse, len(t))
	for i := range t {
		v[i] = FromCrossClusterTaskResponse(t[i])
	}
	return v
}

// ToCrossClusterTaskResponseArray converts proto CrossClusterTaskResponse type array to internal
func ToCrossClusterTaskResponseArray(t []*sharedv1.CrossClusterTaskResponse) []*types.CrossClusterTaskResponse {
	if t == nil {
		return nil
	}
	v := make([]*types.CrossClusterTaskResponse, len(t))
	for i := range t {
		v[i] = ToCrossClusterTaskResponse(t[i])
	}
	return v
}

func FromHistoryDLQCountEntryMap(t map[types.HistoryDLQCountKey]int64) []*sharedv1.HistoryDLQCountEntry {
	if t == nil {
		return nil
	}
	entries := make([]*sharedv1.HistoryDLQCountEntry, 0, len(t))
	for key, count := range t {
		entries = append(entries, &sharedv1.HistoryDLQCountEntry{
			ShardId:       key.ShardID,
			SourceCluster: key.SourceCluster,
			Count:         count,
		})
	}
	return entries
}

func ToHistoryDLQCountEntryMap(t []*sharedv1.HistoryDLQCountEntry) map[types.HistoryDLQCountKey]int64 {
	if t == nil {
		return nil
	}
	entries := make(map[types.HistoryDLQCountKey]int64, len(t))
	for _, entry := range t {
		key := types.HistoryDLQCountKey{ShardID: entry.ShardId, SourceCluster: entry.SourceCluster}
		entries[key] = entry.Count
	}
	return entries
}
