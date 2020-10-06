// The MIT License (MIT)
//
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

package persistence

import (
	"context"
)

type (
	shardManager struct {
		persistence ShardStore
	}
)

var _ ShardManager = (*shardManager)(nil)

// NewShardManager returns a new ShardManager
func NewShardManager(
	persistence ShardStore,
) ShardManager {
	return &shardManager{
		persistence: persistence,
	}
}

func (m *shardManager) GetName() string {
	return m.persistence.GetName()
}

func (m *shardManager) Close() {
	m.persistence.Close()
}

func (m *shardManager) CreateShard(ctx context.Context, request *CreateShardRequest) error {
	internalRequest := &InternalCreateShardRequest{
		ShardInfo: m.toInternalShardInfo(request.ShardInfo),
	}
	return m.persistence.CreateShard(ctx, internalRequest)
}

func (m *shardManager) GetShard(ctx context.Context, request *GetShardRequest) (*GetShardResponse, error) {
	internalRequest := &InternalGetShardRequest{
		ShardID: request.ShardID,
	}
	internalResult, err := m.persistence.GetShard(ctx, internalRequest)
	if err != nil {
		return nil, err
	}
	result := &GetShardResponse{
		ShardInfo: m.fromInternalShardInfo(internalResult.ShardInfo),
	}
	return result, nil
}

func (m *shardManager) UpdateShard(ctx context.Context, request *UpdateShardRequest) error {
	internalRequest := &InternalUpdateShardRequest{
		ShardInfo:       m.toInternalShardInfo(request.ShardInfo),
		PreviousRangeID: request.PreviousRangeID,
	}
	return m.persistence.UpdateShard(ctx, internalRequest)
}

func (m *shardManager) toInternalShardInfo(shardInfo *ShardInfo) *InternalShardInfo {
	if shardInfo == nil {
		return nil
	}
	return &InternalShardInfo{
		ShardID:                       shardInfo.ShardID,
		Owner:                         shardInfo.Owner,
		RangeID:                       shardInfo.RangeID,
		StolenSinceRenew:              shardInfo.StolenSinceRenew,
		UpdatedAt:                     shardInfo.UpdatedAt,
		ReplicationAckLevel:           shardInfo.ReplicationAckLevel,
		ReplicationDLQAckLevel:        shardInfo.ReplicationDLQAckLevel,
		TransferAckLevel:              shardInfo.TransferAckLevel,
		TimerAckLevel:                 shardInfo.TimerAckLevel,
		ClusterTransferAckLevel:       shardInfo.ClusterTransferAckLevel,
		ClusterTimerAckLevel:          shardInfo.ClusterTimerAckLevel,
		TransferProcessingQueueStates: shardInfo.TransferProcessingQueueStates,
		TimerProcessingQueueStates:    shardInfo.TimerProcessingQueueStates,
		ClusterReplicationLevel:       shardInfo.ClusterReplicationLevel,
		DomainNotificationVersion:     shardInfo.DomainNotificationVersion,
		PendingFailoverMarkers:        shardInfo.PendingFailoverMarkers,
		TransferFailoverLevels:        shardInfo.TransferFailoverLevels,
		TimerFailoverLevels:           shardInfo.TimerFailoverLevels,
	}
}

func (m *shardManager) fromInternalShardInfo(internalShardInfo *InternalShardInfo) *ShardInfo {
	if internalShardInfo == nil {
		return nil
	}
	return &ShardInfo{
		ShardID:                       internalShardInfo.ShardID,
		Owner:                         internalShardInfo.Owner,
		RangeID:                       internalShardInfo.RangeID,
		StolenSinceRenew:              internalShardInfo.StolenSinceRenew,
		UpdatedAt:                     internalShardInfo.UpdatedAt,
		ReplicationAckLevel:           internalShardInfo.ReplicationAckLevel,
		ReplicationDLQAckLevel:        internalShardInfo.ReplicationDLQAckLevel,
		TransferAckLevel:              internalShardInfo.TransferAckLevel,
		TimerAckLevel:                 internalShardInfo.TimerAckLevel,
		ClusterTransferAckLevel:       internalShardInfo.ClusterTransferAckLevel,
		ClusterTimerAckLevel:          internalShardInfo.ClusterTimerAckLevel,
		TransferProcessingQueueStates: internalShardInfo.TransferProcessingQueueStates,
		TimerProcessingQueueStates:    internalShardInfo.TimerProcessingQueueStates,
		ClusterReplicationLevel:       internalShardInfo.ClusterReplicationLevel,
		DomainNotificationVersion:     internalShardInfo.DomainNotificationVersion,
		PendingFailoverMarkers:        internalShardInfo.PendingFailoverMarkers,
		TransferFailoverLevels:        internalShardInfo.TransferFailoverLevels,
		TimerFailoverLevels:           internalShardInfo.TimerFailoverLevels,
	}
}
