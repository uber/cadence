// Copyright (c) 2018 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package sql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/uber/cadence/common/persistence/serialization"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/sql/sqlplugin"
	"github.com/uber/cadence/common/types"
)

type sqlShardManager struct {
	sqlStore
	currentClusterName string
}

// newShardPersistence creates an instance of ShardStore
func newShardPersistence(
	db sqlplugin.DB,
	currentClusterName string,
	log log.Logger,
	parser serialization.Parser,
) (persistence.ShardStore, error) {
	return &sqlShardManager{
		sqlStore: sqlStore{
			db:     db,
			logger: log,
			parser: parser,
		},
		currentClusterName: currentClusterName,
	}, nil
}

func (m *sqlShardManager) CreateShard(
	ctx context.Context,
	request *persistence.InternalCreateShardRequest,
) error {
	if _, err := m.GetShard(ctx, &persistence.InternalGetShardRequest{
		ShardID: request.ShardInfo.ShardID,
	}); err == nil {
		return &persistence.ShardAlreadyExistError{
			Msg: fmt.Sprintf("CreateShard operaiton failed. Shard with ID %v already exists.", request.ShardInfo.ShardID),
		}
	}

	row, err := shardInfoToShardsRow(*request.ShardInfo, m.parser)
	if err != nil {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("CreateShard operation failed. Error: %v", err),
		}
	}

	if _, err := m.db.InsertIntoShards(ctx, row); err != nil {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("CreateShard operation failed. Failed to insert into shards table. Error: %v", err),
		}
	}

	return nil
}

func (m *sqlShardManager) GetShard(
	ctx context.Context,
	request *persistence.InternalGetShardRequest,
) (*persistence.InternalGetShardResponse, error) {
	row, err := m.db.SelectFromShards(ctx, &sqlplugin.ShardsFilter{ShardID: int64(request.ShardID)})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &types.EntityNotExistsError{
				Message: fmt.Sprintf("GetShard operation failed. Shard with ID %v not found. Error: %v", request.ShardID, err),
			}
		}
		return nil, &types.InternalServiceError{
			Message: fmt.Sprintf("GetShard operation failed. Failed to get record. ShardId: %v. Error: %v", request.ShardID, err),
		}
	}

	shardInfo, err := m.parser.ShardInfoFromBlob(row.Data, row.DataEncoding)
	if err != nil {
		return nil, err
	}

	if len(shardInfo.ClusterTransferAckLevel) == 0 {
		shardInfo.ClusterTransferAckLevel = map[string]int64{
			m.currentClusterName: shardInfo.GetTransferAckLevel(),
		}
	}

	timerAckLevel := make(map[string]time.Time, len(shardInfo.ClusterTimerAckLevel))
	for k, v := range shardInfo.ClusterTimerAckLevel {
		timerAckLevel[k] = v
	}

	if len(timerAckLevel) == 0 {
		timerAckLevel = map[string]time.Time{
			m.currentClusterName: shardInfo.GetTimerAckLevel(),
		}
	}

	if shardInfo.ClusterReplicationLevel == nil {
		shardInfo.ClusterReplicationLevel = make(map[string]int64)
	}
	if shardInfo.ReplicationDlqAckLevel == nil {
		shardInfo.ReplicationDlqAckLevel = make(map[string]int64)
	}

	var transferPQS *persistence.DataBlob
	if shardInfo.GetTransferProcessingQueueStates() != nil {
		transferPQS = &persistence.DataBlob{
			Encoding: common.EncodingType(shardInfo.GetTransferProcessingQueueStatesEncoding()),
			Data:     shardInfo.GetTransferProcessingQueueStates(),
		}
	}

	var timerPQS *persistence.DataBlob
	if shardInfo.GetTimerProcessingQueueStates() != nil {
		timerPQS = &persistence.DataBlob{
			Encoding: common.EncodingType(shardInfo.GetTimerProcessingQueueStatesEncoding()),
			Data:     shardInfo.GetTimerProcessingQueueStates(),
		}
	}

	resp := &persistence.InternalGetShardResponse{ShardInfo: &persistence.InternalShardInfo{
		ShardID:                       int(row.ShardID),
		RangeID:                       row.RangeID,
		Owner:                         shardInfo.GetOwner(),
		StolenSinceRenew:              int(shardInfo.GetStolenSinceRenew()),
		UpdatedAt:                     shardInfo.GetUpdatedAt(),
		ReplicationAckLevel:           shardInfo.GetReplicationAckLevel(),
		TransferAckLevel:              shardInfo.GetTransferAckLevel(),
		TimerAckLevel:                 shardInfo.GetTimerAckLevel(),
		ClusterTransferAckLevel:       shardInfo.ClusterTransferAckLevel,
		ClusterTimerAckLevel:          timerAckLevel,
		TransferProcessingQueueStates: transferPQS,
		TimerProcessingQueueStates:    timerPQS,
		DomainNotificationVersion:     shardInfo.GetDomainNotificationVersion(),
		ClusterReplicationLevel:       shardInfo.ClusterReplicationLevel,
		ReplicationDLQAckLevel:        shardInfo.ReplicationDlqAckLevel,
	}}

	return resp, nil
}

func (m *sqlShardManager) UpdateShard(
	ctx context.Context,
	request *persistence.InternalUpdateShardRequest,
) error {
	row, err := shardInfoToShardsRow(*request.ShardInfo, m.parser)
	if err != nil {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("UpdateShard operation failed. Error: %v", err),
		}
	}
	return m.txExecute(ctx, "UpdateShard", func(tx sqlplugin.Tx) error {
		if err := lockShard(ctx, tx, request.ShardInfo.ShardID, request.PreviousRangeID); err != nil {
			return err
		}
		result, err := tx.UpdateShards(ctx, row)
		if err != nil {
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("rowsAffected returned error for shardID %v: %v", request.ShardInfo.ShardID, err)
		}
		if rowsAffected != 1 {
			return fmt.Errorf("rowsAffected returned %v shards instead of one", rowsAffected)
		}
		return nil
	})
}

// initiated by the owning shard
func lockShard(ctx context.Context, tx sqlplugin.Tx, shardID int, oldRangeID int64) error {
	rangeID, err := tx.WriteLockShards(ctx, &sqlplugin.ShardsFilter{ShardID: int64(shardID)})
	if err != nil {
		if err == sql.ErrNoRows {
			return &types.InternalServiceError{
				Message: fmt.Sprintf("Failed to lock shard with ID %v that does not exist.", shardID),
			}
		}
		return &types.InternalServiceError{
			Message: fmt.Sprintf("Failed to lock shard with ID: %v. Error: %v", shardID, err),
		}
	}

	if int64(rangeID) != oldRangeID {
		return &persistence.ShardOwnershipLostError{
			ShardID: shardID,
			Msg:     fmt.Sprintf("Failed to update shard. Previous range ID: %v; new range ID: %v", oldRangeID, rangeID),
		}
	}

	return nil
}

// initiated by the owning shard
func readLockShard(ctx context.Context, tx sqlplugin.Tx, shardID int, oldRangeID int64) error {
	rangeID, err := tx.ReadLockShards(ctx, &sqlplugin.ShardsFilter{ShardID: int64(shardID)})
	if err != nil {
		if err == sql.ErrNoRows {
			return &types.InternalServiceError{
				Message: fmt.Sprintf("Failed to lock shard with ID %v that does not exist.", shardID),
			}
		}
		return &types.InternalServiceError{
			Message: fmt.Sprintf("Failed to lock shard with ID: %v. Error: %v", shardID, err),
		}
	}

	if int64(rangeID) != oldRangeID {
		return &persistence.ShardOwnershipLostError{
			ShardID: shardID,
			Msg:     fmt.Sprintf("Failed to lock shard. Previous range ID: %v; new range ID: %v", oldRangeID, rangeID),
		}
	}
	return nil
}

func shardInfoToShardsRow(s persistence.InternalShardInfo, parser serialization.Parser) (*sqlplugin.ShardsRow, error) {
	var markerData []byte
	var markerEncoding string
	if s.PendingFailoverMarkers != nil {
		markerData = s.PendingFailoverMarkers.Data
		markerEncoding = string(s.PendingFailoverMarkers.Encoding)
	}

	var transferPQSData []byte
	var transferPQSEncoding string
	if s.TransferProcessingQueueStates != nil {
		transferPQSData = s.TransferProcessingQueueStates.Data
		transferPQSEncoding = string(s.TransferProcessingQueueStates.Encoding)
	}

	var timerPQSData []byte
	var timerPQSEncoding string
	if s.TimerProcessingQueueStates != nil {
		timerPQSData = s.TimerProcessingQueueStates.Data
		timerPQSEncoding = string(s.TimerProcessingQueueStates.Encoding)
	}

	shardInfo := &serialization.ShardInfo{
		StolenSinceRenew:                      common.Int32Ptr(int32(s.StolenSinceRenew)),
		UpdatedAt:                             &s.UpdatedAt,
		ReplicationAckLevel:                   common.Int64Ptr(s.ReplicationAckLevel),
		TransferAckLevel:                      common.Int64Ptr(s.TransferAckLevel),
		TimerAckLevel:                         &s.TimerAckLevel,
		ClusterTransferAckLevel:               s.ClusterTransferAckLevel,
		ClusterTimerAckLevel:                  s.ClusterTimerAckLevel,
		TransferProcessingQueueStates:         transferPQSData,
		TransferProcessingQueueStatesEncoding: common.StringPtr(transferPQSEncoding),
		TimerProcessingQueueStates:            timerPQSData,
		TimerProcessingQueueStatesEncoding:    common.StringPtr(timerPQSEncoding),
		DomainNotificationVersion:             common.Int64Ptr(s.DomainNotificationVersion),
		Owner:                                 &s.Owner,
		ClusterReplicationLevel:               s.ClusterReplicationLevel,
		ReplicationDlqAckLevel:                s.ReplicationDLQAckLevel,
		PendingFailoverMarkers:                markerData,
		PendingFailoverMarkersEncoding:        common.StringPtr(markerEncoding),
	}

	blob, err := parser.ShardInfoToBlob(shardInfo)
	if err != nil {
		return nil, err
	}
	return &sqlplugin.ShardsRow{
		ShardID:      int64(s.ShardID),
		RangeID:      s.RangeID,
		Data:         blob.Data,
		DataEncoding: string(blob.Encoding),
	}, nil
}
