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
