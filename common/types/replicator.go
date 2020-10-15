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

package types

// DLQType is an internal type (TBD...)
type DLQType int32

const (
	// DLQTypeDomain is an option for DLQType
	DLQTypeDomain DLQType = iota
	// DLQTypeReplication is an option for DLQType
	DLQTypeReplication
)

// DomainOperation is an internal type (TBD...)
type DomainOperation int32

const (
	// DomainOperationCreate is an option for DomainOperation
	DomainOperationCreate DomainOperation = iota
	// DomainOperationUpdate is an option for DomainOperation
	DomainOperationUpdate
)

// DomainTaskAttributes is an internal type (TBD...)
type DomainTaskAttributes struct {
	DomainOperation         *DomainOperation
	ID                      *string
	Info                    *DomainInfo
	Config                  *DomainConfiguration
	ReplicationConfig       *DomainReplicationConfiguration
	ConfigVersion           *int64
	FailoverVersion         *int64
	PreviousFailoverVersion *int64
}

func (v *DomainTaskAttributes) GetDomainOperation() (o DomainOperation) {
	if v != nil && v.DomainOperation != nil {
		return *v.DomainOperation
	}
	return
}
func (v *DomainTaskAttributes) GetID() (o string) {
	if v != nil && v.ID != nil {
		return *v.ID
	}
	return
}
func (v *DomainTaskAttributes) GetInfo() (o *DomainInfo) {
	if v != nil && v.Info != nil {
		return v.Info
	}
	return
}
func (v *DomainTaskAttributes) GetConfig() (o *DomainConfiguration) {
	if v != nil && v.Config != nil {
		return v.Config
	}
	return
}
func (v *DomainTaskAttributes) GetReplicationConfig() (o *DomainReplicationConfiguration) {
	if v != nil && v.ReplicationConfig != nil {
		return v.ReplicationConfig
	}
	return
}
func (v *DomainTaskAttributes) GetConfigVersion() (o int64) {
	if v != nil && v.ConfigVersion != nil {
		return *v.ConfigVersion
	}
	return
}
func (v *DomainTaskAttributes) GetFailoverVersion() (o int64) {
	if v != nil && v.FailoverVersion != nil {
		return *v.FailoverVersion
	}
	return
}
func (v *DomainTaskAttributes) GetPreviousFailoverVersion() (o int64) {
	if v != nil && v.PreviousFailoverVersion != nil {
		return *v.PreviousFailoverVersion
	}
	return
}

// FailoverMarkerAttributes is an internal type (TBD...)
type FailoverMarkerAttributes struct {
	DomainID        *string
	FailoverVersion *int64
	CreationTime    *int64
}

func (v *FailoverMarkerAttributes) GetDomainID() (o string) {
	if v != nil && v.DomainID != nil {
		return *v.DomainID
	}
	return
}
func (v *FailoverMarkerAttributes) GetFailoverVersion() (o int64) {
	if v != nil && v.FailoverVersion != nil {
		return *v.FailoverVersion
	}
	return
}
func (v *FailoverMarkerAttributes) GetCreationTime() (o int64) {
	if v != nil && v.CreationTime != nil {
		return *v.CreationTime
	}
	return
}

// FailoverMarkers is an internal type (TBD...)
type FailoverMarkers struct {
	FailoverMarkers []*FailoverMarkerAttributes
}

func (v *FailoverMarkers) GetFailoverMarkers() (o []*FailoverMarkerAttributes) {
	if v != nil && v.FailoverMarkers != nil {
		return v.FailoverMarkers
	}
	return
}

// GetDLQReplicationMessagesRequest is an internal type (TBD...)
type GetDLQReplicationMessagesRequest struct {
	TaskInfos []*ReplicationTaskInfo
}

func (v *GetDLQReplicationMessagesRequest) GetTaskInfos() (o []*ReplicationTaskInfo) {
	if v != nil && v.TaskInfos != nil {
		return v.TaskInfos
	}
	return
}

// GetDLQReplicationMessagesResponse is an internal type (TBD...)
type GetDLQReplicationMessagesResponse struct {
	ReplicationTasks []*ReplicationTask
}

func (v *GetDLQReplicationMessagesResponse) GetReplicationTasks() (o []*ReplicationTask) {
	if v != nil && v.ReplicationTasks != nil {
		return v.ReplicationTasks
	}
	return
}

// GetDomainReplicationMessagesRequest is an internal type (TBD...)
type GetDomainReplicationMessagesRequest struct {
	LastRetrievedMessageID *int64
	LastProcessedMessageID *int64
	ClusterName            *string
}

func (v *GetDomainReplicationMessagesRequest) GetLastRetrievedMessageID() (o int64) {
	if v != nil && v.LastRetrievedMessageID != nil {
		return *v.LastRetrievedMessageID
	}
	return
}
func (v *GetDomainReplicationMessagesRequest) GetLastProcessedMessageID() (o int64) {
	if v != nil && v.LastProcessedMessageID != nil {
		return *v.LastProcessedMessageID
	}
	return
}
func (v *GetDomainReplicationMessagesRequest) GetClusterName() (o string) {
	if v != nil && v.ClusterName != nil {
		return *v.ClusterName
	}
	return
}

// GetDomainReplicationMessagesResponse is an internal type (TBD...)
type GetDomainReplicationMessagesResponse struct {
	Messages *ReplicationMessages
}

func (v *GetDomainReplicationMessagesResponse) GetMessages() (o *ReplicationMessages) {
	if v != nil && v.Messages != nil {
		return v.Messages
	}
	return
}

// GetReplicationMessagesRequest is an internal type (TBD...)
type GetReplicationMessagesRequest struct {
	Tokens      []*ReplicationToken
	ClusterName *string
}

func (v *GetReplicationMessagesRequest) GetTokens() (o []*ReplicationToken) {
	if v != nil && v.Tokens != nil {
		return v.Tokens
	}
	return
}
func (v *GetReplicationMessagesRequest) GetClusterName() (o string) {
	if v != nil && v.ClusterName != nil {
		return *v.ClusterName
	}
	return
}

// GetReplicationMessagesResponse is an internal type (TBD...)
type GetReplicationMessagesResponse struct {
	MessagesByShard map[int32]*ReplicationMessages
}

func (v *GetReplicationMessagesResponse) GetMessagesByShard() (o map[int32]*ReplicationMessages) {
	if v != nil && v.MessagesByShard != nil {
		return v.MessagesByShard
	}
	return
}

// HistoryTaskV2Attributes is an internal type (TBD...)
type HistoryTaskV2Attributes struct {
	TaskID              *int64
	DomainID            *string
	WorkflowID          *string
	RunID               *string
	VersionHistoryItems []*VersionHistoryItem
	Events              *DataBlob
	NewRunEvents        *DataBlob
}

func (v *HistoryTaskV2Attributes) GetTaskID() (o int64) {
	if v != nil && v.TaskID != nil {
		return *v.TaskID
	}
	return
}
func (v *HistoryTaskV2Attributes) GetDomainID() (o string) {
	if v != nil && v.DomainID != nil {
		return *v.DomainID
	}
	return
}
func (v *HistoryTaskV2Attributes) GetWorkflowID() (o string) {
	if v != nil && v.WorkflowID != nil {
		return *v.WorkflowID
	}
	return
}
func (v *HistoryTaskV2Attributes) GetRunID() (o string) {
	if v != nil && v.RunID != nil {
		return *v.RunID
	}
	return
}
func (v *HistoryTaskV2Attributes) GetVersionHistoryItems() (o []*VersionHistoryItem) {
	if v != nil && v.VersionHistoryItems != nil {
		return v.VersionHistoryItems
	}
	return
}
func (v *HistoryTaskV2Attributes) GetEvents() (o *DataBlob) {
	if v != nil && v.Events != nil {
		return v.Events
	}
	return
}
func (v *HistoryTaskV2Attributes) GetNewRunEvents() (o *DataBlob) {
	if v != nil && v.NewRunEvents != nil {
		return v.NewRunEvents
	}
	return
}

// MergeDLQMessagesRequest is an internal type (TBD...)
type MergeDLQMessagesRequest struct {
	Type                  *DLQType
	ShardID               *int32
	SourceCluster         *string
	InclusiveEndMessageID *int64
	MaximumPageSize       *int32
	NextPageToken         []byte
}

func (v *MergeDLQMessagesRequest) GetType() (o DLQType) {
	if v != nil && v.Type != nil {
		return *v.Type
	}
	return
}
func (v *MergeDLQMessagesRequest) GetShardID() (o int32) {
	if v != nil && v.ShardID != nil {
		return *v.ShardID
	}
	return
}
func (v *MergeDLQMessagesRequest) GetSourceCluster() (o string) {
	if v != nil && v.SourceCluster != nil {
		return *v.SourceCluster
	}
	return
}
func (v *MergeDLQMessagesRequest) GetInclusiveEndMessageID() (o int64) {
	if v != nil && v.InclusiveEndMessageID != nil {
		return *v.InclusiveEndMessageID
	}
	return
}
func (v *MergeDLQMessagesRequest) GetMaximumPageSize() (o int32) {
	if v != nil && v.MaximumPageSize != nil {
		return *v.MaximumPageSize
	}
	return
}
func (v *MergeDLQMessagesRequest) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

// MergeDLQMessagesResponse is an internal type (TBD...)
type MergeDLQMessagesResponse struct {
	NextPageToken []byte
}

func (v *MergeDLQMessagesResponse) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

// PurgeDLQMessagesRequest is an internal type (TBD...)
type PurgeDLQMessagesRequest struct {
	Type                  *DLQType
	ShardID               *int32
	SourceCluster         *string
	InclusiveEndMessageID *int64
}

func (v *PurgeDLQMessagesRequest) GetType() (o DLQType) {
	if v != nil && v.Type != nil {
		return *v.Type
	}
	return
}
func (v *PurgeDLQMessagesRequest) GetShardID() (o int32) {
	if v != nil && v.ShardID != nil {
		return *v.ShardID
	}
	return
}
func (v *PurgeDLQMessagesRequest) GetSourceCluster() (o string) {
	if v != nil && v.SourceCluster != nil {
		return *v.SourceCluster
	}
	return
}
func (v *PurgeDLQMessagesRequest) GetInclusiveEndMessageID() (o int64) {
	if v != nil && v.InclusiveEndMessageID != nil {
		return *v.InclusiveEndMessageID
	}
	return
}

// ReadDLQMessagesRequest is an internal type (TBD...)
type ReadDLQMessagesRequest struct {
	Type                  *DLQType
	ShardID               *int32
	SourceCluster         *string
	InclusiveEndMessageID *int64
	MaximumPageSize       *int32
	NextPageToken         []byte
}

func (v *ReadDLQMessagesRequest) GetType() (o DLQType) {
	if v != nil && v.Type != nil {
		return *v.Type
	}
	return
}
func (v *ReadDLQMessagesRequest) GetShardID() (o int32) {
	if v != nil && v.ShardID != nil {
		return *v.ShardID
	}
	return
}
func (v *ReadDLQMessagesRequest) GetSourceCluster() (o string) {
	if v != nil && v.SourceCluster != nil {
		return *v.SourceCluster
	}
	return
}
func (v *ReadDLQMessagesRequest) GetInclusiveEndMessageID() (o int64) {
	if v != nil && v.InclusiveEndMessageID != nil {
		return *v.InclusiveEndMessageID
	}
	return
}
func (v *ReadDLQMessagesRequest) GetMaximumPageSize() (o int32) {
	if v != nil && v.MaximumPageSize != nil {
		return *v.MaximumPageSize
	}
	return
}
func (v *ReadDLQMessagesRequest) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

// ReadDLQMessagesResponse is an internal type (TBD...)
type ReadDLQMessagesResponse struct {
	Type             *DLQType
	ReplicationTasks []*ReplicationTask
	NextPageToken    []byte
}

func (v *ReadDLQMessagesResponse) GetType() (o DLQType) {
	if v != nil && v.Type != nil {
		return *v.Type
	}
	return
}
func (v *ReadDLQMessagesResponse) GetReplicationTasks() (o []*ReplicationTask) {
	if v != nil && v.ReplicationTasks != nil {
		return v.ReplicationTasks
	}
	return
}
func (v *ReadDLQMessagesResponse) GetNextPageToken() (o []byte) {
	if v != nil && v.NextPageToken != nil {
		return v.NextPageToken
	}
	return
}

// ReplicationMessages is an internal type (TBD...)
type ReplicationMessages struct {
	ReplicationTasks       []*ReplicationTask
	LastRetrievedMessageID *int64
	HasMore                *bool
	SyncShardStatus        *SyncShardStatus
}

func (v *ReplicationMessages) GetReplicationTasks() (o []*ReplicationTask) {
	if v != nil && v.ReplicationTasks != nil {
		return v.ReplicationTasks
	}
	return
}
func (v *ReplicationMessages) GetLastRetrievedMessageID() (o int64) {
	if v != nil && v.LastRetrievedMessageID != nil {
		return *v.LastRetrievedMessageID
	}
	return
}
func (v *ReplicationMessages) GetHasMore() (o bool) {
	if v != nil && v.HasMore != nil {
		return *v.HasMore
	}
	return
}
func (v *ReplicationMessages) GetSyncShardStatus() (o *SyncShardStatus) {
	if v != nil && v.SyncShardStatus != nil {
		return v.SyncShardStatus
	}
	return
}

// ReplicationTask is an internal type (TBD...)
type ReplicationTask struct {
	TaskType                      *ReplicationTaskType
	SourceTaskID                  *int64
	DomainTaskAttributes          *DomainTaskAttributes
	SyncShardStatusTaskAttributes *SyncShardStatusTaskAttributes
	SyncActivityTaskAttributes    *SyncActivityTaskAttributes
	HistoryTaskV2Attributes       *HistoryTaskV2Attributes
	FailoverMarkerAttributes      *FailoverMarkerAttributes
}

func (v *ReplicationTask) GetTaskType() (o ReplicationTaskType) {
	if v != nil && v.TaskType != nil {
		return *v.TaskType
	}
	return
}
func (v *ReplicationTask) GetSourceTaskID() (o int64) {
	if v != nil && v.SourceTaskID != nil {
		return *v.SourceTaskID
	}
	return
}
func (v *ReplicationTask) GetDomainTaskAttributes() (o *DomainTaskAttributes) {
	if v != nil && v.DomainTaskAttributes != nil {
		return v.DomainTaskAttributes
	}
	return
}
func (v *ReplicationTask) GetSyncShardStatusTaskAttributes() (o *SyncShardStatusTaskAttributes) {
	if v != nil && v.SyncShardStatusTaskAttributes != nil {
		return v.SyncShardStatusTaskAttributes
	}
	return
}
func (v *ReplicationTask) GetSyncActivityTaskAttributes() (o *SyncActivityTaskAttributes) {
	if v != nil && v.SyncActivityTaskAttributes != nil {
		return v.SyncActivityTaskAttributes
	}
	return
}
func (v *ReplicationTask) GetHistoryTaskV2Attributes() (o *HistoryTaskV2Attributes) {
	if v != nil && v.HistoryTaskV2Attributes != nil {
		return v.HistoryTaskV2Attributes
	}
	return
}
func (v *ReplicationTask) GetFailoverMarkerAttributes() (o *FailoverMarkerAttributes) {
	if v != nil && v.FailoverMarkerAttributes != nil {
		return v.FailoverMarkerAttributes
	}
	return
}

// ReplicationTaskInfo is an internal type (TBD...)
type ReplicationTaskInfo struct {
	DomainID     *string
	WorkflowID   *string
	RunID        *string
	TaskType     *int16
	TaskID       *int64
	Version      *int64
	FirstEventID *int64
	NextEventID  *int64
	ScheduledID  *int64
}

func (v *ReplicationTaskInfo) GetDomainID() (o string) {
	if v != nil && v.DomainID != nil {
		return *v.DomainID
	}
	return
}
func (v *ReplicationTaskInfo) GetWorkflowID() (o string) {
	if v != nil && v.WorkflowID != nil {
		return *v.WorkflowID
	}
	return
}
func (v *ReplicationTaskInfo) GetRunID() (o string) {
	if v != nil && v.RunID != nil {
		return *v.RunID
	}
	return
}
func (v *ReplicationTaskInfo) GetTaskType() (o int16) {
	if v != nil && v.TaskType != nil {
		return *v.TaskType
	}
	return
}
func (v *ReplicationTaskInfo) GetTaskID() (o int64) {
	if v != nil && v.TaskID != nil {
		return *v.TaskID
	}
	return
}
func (v *ReplicationTaskInfo) GetVersion() (o int64) {
	if v != nil && v.Version != nil {
		return *v.Version
	}
	return
}
func (v *ReplicationTaskInfo) GetFirstEventID() (o int64) {
	if v != nil && v.FirstEventID != nil {
		return *v.FirstEventID
	}
	return
}
func (v *ReplicationTaskInfo) GetNextEventID() (o int64) {
	if v != nil && v.NextEventID != nil {
		return *v.NextEventID
	}
	return
}
func (v *ReplicationTaskInfo) GetScheduledID() (o int64) {
	if v != nil && v.ScheduledID != nil {
		return *v.ScheduledID
	}
	return
}

// ReplicationTaskType is an internal type (TBD...)
type ReplicationTaskType int32

const (
	// ReplicationTaskTypeDomain is an option for ReplicationTaskType
	ReplicationTaskTypeDomain ReplicationTaskType = iota
	// ReplicationTaskTypeFailoverMarker is an option for ReplicationTaskType
	ReplicationTaskTypeFailoverMarker
	// ReplicationTaskTypeHistory is an option for ReplicationTaskType
	ReplicationTaskTypeHistory
	// ReplicationTaskTypeHistoryMetadata is an option for ReplicationTaskType
	ReplicationTaskTypeHistoryMetadata
	// ReplicationTaskTypeHistoryV2 is an option for ReplicationTaskType
	ReplicationTaskTypeHistoryV2
	// ReplicationTaskTypeSyncActivity is an option for ReplicationTaskType
	ReplicationTaskTypeSyncActivity
	// ReplicationTaskTypeSyncShardStatus is an option for ReplicationTaskType
	ReplicationTaskTypeSyncShardStatus
)

// ReplicationToken is an internal type (TBD...)
type ReplicationToken struct {
	ShardID                *int32
	LastRetrievedMessageID *int64
	LastProcessedMessageID *int64
}

func (v *ReplicationToken) GetShardID() (o int32) {
	if v != nil && v.ShardID != nil {
		return *v.ShardID
	}
	return
}
func (v *ReplicationToken) GetLastRetrievedMessageID() (o int64) {
	if v != nil && v.LastRetrievedMessageID != nil {
		return *v.LastRetrievedMessageID
	}
	return
}
func (v *ReplicationToken) GetLastProcessedMessageID() (o int64) {
	if v != nil && v.LastProcessedMessageID != nil {
		return *v.LastProcessedMessageID
	}
	return
}

// SyncActivityTaskAttributes is an internal type (TBD...)
type SyncActivityTaskAttributes struct {
	DomainID           *string
	WorkflowID         *string
	RunID              *string
	Version            *int64
	ScheduledID        *int64
	ScheduledTime      *int64
	StartedID          *int64
	StartedTime        *int64
	LastHeartbeatTime  *int64
	Details            []byte
	Attempt            *int32
	LastFailureReason  *string
	LastWorkerIdentity *string
	LastFailureDetails []byte
	VersionHistory     *VersionHistory
}

func (v *SyncActivityTaskAttributes) GetDomainID() (o string) {
	if v != nil && v.DomainID != nil {
		return *v.DomainID
	}
	return
}
func (v *SyncActivityTaskAttributes) GetWorkflowID() (o string) {
	if v != nil && v.WorkflowID != nil {
		return *v.WorkflowID
	}
	return
}
func (v *SyncActivityTaskAttributes) GetRunID() (o string) {
	if v != nil && v.RunID != nil {
		return *v.RunID
	}
	return
}
func (v *SyncActivityTaskAttributes) GetVersion() (o int64) {
	if v != nil && v.Version != nil {
		return *v.Version
	}
	return
}
func (v *SyncActivityTaskAttributes) GetScheduledID() (o int64) {
	if v != nil && v.ScheduledID != nil {
		return *v.ScheduledID
	}
	return
}
func (v *SyncActivityTaskAttributes) GetScheduledTime() (o int64) {
	if v != nil && v.ScheduledTime != nil {
		return *v.ScheduledTime
	}
	return
}
func (v *SyncActivityTaskAttributes) GetStartedID() (o int64) {
	if v != nil && v.StartedID != nil {
		return *v.StartedID
	}
	return
}
func (v *SyncActivityTaskAttributes) GetStartedTime() (o int64) {
	if v != nil && v.StartedTime != nil {
		return *v.StartedTime
	}
	return
}
func (v *SyncActivityTaskAttributes) GetLastHeartbeatTime() (o int64) {
	if v != nil && v.LastHeartbeatTime != nil {
		return *v.LastHeartbeatTime
	}
	return
}
func (v *SyncActivityTaskAttributes) GetDetails() (o []byte) {
	if v != nil && v.Details != nil {
		return v.Details
	}
	return
}
func (v *SyncActivityTaskAttributes) GetAttempt() (o int32) {
	if v != nil && v.Attempt != nil {
		return *v.Attempt
	}
	return
}
func (v *SyncActivityTaskAttributes) GetLastFailureReason() (o string) {
	if v != nil && v.LastFailureReason != nil {
		return *v.LastFailureReason
	}
	return
}
func (v *SyncActivityTaskAttributes) GetLastWorkerIdentity() (o string) {
	if v != nil && v.LastWorkerIdentity != nil {
		return *v.LastWorkerIdentity
	}
	return
}
func (v *SyncActivityTaskAttributes) GetLastFailureDetails() (o []byte) {
	if v != nil && v.LastFailureDetails != nil {
		return v.LastFailureDetails
	}
	return
}
func (v *SyncActivityTaskAttributes) GetVersionHistory() (o *VersionHistory) {
	if v != nil && v.VersionHistory != nil {
		return v.VersionHistory
	}
	return
}

// SyncShardStatus is an internal type (TBD...)
type SyncShardStatus struct {
	Timestamp *int64
}

func (v *SyncShardStatus) GetTimestamp() (o int64) {
	if v != nil && v.Timestamp != nil {
		return *v.Timestamp
	}
	return
}

// SyncShardStatusTaskAttributes is an internal type (TBD...)
type SyncShardStatusTaskAttributes struct {
	SourceCluster *string
	ShardID       *int64
	Timestamp     *int64
}

func (v *SyncShardStatusTaskAttributes) GetSourceCluster() (o string) {
	if v != nil && v.SourceCluster != nil {
		return *v.SourceCluster
	}
	return
}
func (v *SyncShardStatusTaskAttributes) GetShardID() (o int64) {
	if v != nil && v.ShardID != nil {
		return *v.ShardID
	}
	return
}
func (v *SyncShardStatusTaskAttributes) GetTimestamp() (o int64) {
	if v != nil && v.Timestamp != nil {
		return *v.Timestamp
	}
	return
}
