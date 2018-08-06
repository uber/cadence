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
	"database/sql"
	"fmt"
	"time"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/persistence"

	"github.com/hmgle/sqlx"
)

type (
	sqlShardManager struct {
		db                 *sqlx.DB
		currentClusterName string
	}

	shardsRow struct {
		ShardID                   int64     `db:"shard_id"`
		Owner                     string    `db:"owner"`
		RangeID                   int64     `db:"range_id"`
		StolenSinceRenew          int64     `db:"stolen_since_renew"`
		UpdatedAt                 time.Time `db:"updated_at"`
		ReplicationAckLevel       int64     `db:"replication_ack_level"`
		TransferAckLevel          int64     `db:"transfer_ack_level"`
		TimerAckLevel             time.Time `db:"timer_ack_level"`
		ClusterTransferAckLevel   []byte    `db:"cluster_transfer_ack_level"`
		ClusterTimerAckLevel      []byte    `db:"cluster_timer_ack_level"`
		DomainNotificationVersion int64     `db:"domain_notification_version"`
	}

	updateShardRequest struct {
		shardsRow
		OldRangeID int64 `db:"old_range_id"`
	}
)

const (
	createShardSQLQuery = `INSERT INTO shards 
(shard_id, 
owner, 
range_id,
stolen_since_renew,
updated_at,
replication_ack_level,
transfer_ack_level,
timer_ack_level,
cluster_transfer_ack_level,
cluster_timer_ack_level,
domain_notification_version)
VALUES
(:shard_id, 
:owner, 
:range_id,
:stolen_since_renew,
:updated_at,
:replication_ack_level,
:transfer_ack_level,
:timer_ack_level,
:cluster_transfer_ack_level,
:cluster_timer_ack_level,
:domain_notification_version)`

	getShardSQLQuery = `SELECT
shard_id,
owner,
range_id,
stolen_since_renew,
updated_at,
replication_ack_level,
transfer_ack_level,
timer_ack_level,
cluster_transfer_ack_level,
cluster_timer_ack_level,
domain_notification_version
FROM shards WHERE
shard_id = ?
`

	updateShardSQLQuery = `UPDATE
shards 
SET
shard_id = :shard_id,
owner = :owner,
range_id = :range_id,
stolen_since_renew = :stolen_since_renew,
updated_at = :updated_at,
replication_ack_level = :replication_ack_level,
transfer_ack_level = :transfer_ack_level,
timer_ack_level = :timer_ack_level,
cluster_transfer_ack_level = :cluster_transfer_ack_level,
cluster_timer_ack_level = :cluster_timer_ack_level,
domain_notification_version = :domain_notification_version
WHERE
shard_id = :shard_id AND
range_id = :old_range_id
`

	lockShardSQLQuery = `SELECT range_id FROM shards WHERE shard_id = ? FOR UPDATE`
)

func NewShardPersistence(username, password, host, port, dbName string, currentClusterName string) (persistence.ShardManager, error) {
	var db, err = sqlx.Connect("mysql",
		fmt.Sprintf(Dsn, username, password, host, port, dbName))
	if err != nil {
		return nil, err
	}

	return &sqlShardManager{
		db:                 db,
		currentClusterName: currentClusterName,
	}, nil
}

func (m *sqlShardManager) Close() {
	if m.db != nil {
		m.db.Close()
	}
}

func (m *sqlShardManager) CreateShard(request *persistence.CreateShardRequest) error {
	row, err := shardInfoToShardsRow(*request.ShardInfo)
	if err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateShard operation failed. Error: %v", err),
		}
	}

	if _, err := m.db.NamedExec(createShardSQLQuery, &row); err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateShard operation failed. Failed to insert into shards table. Error: %v", err),
		}
	}

	return nil
}

func (m *sqlShardManager) GetShard(request *persistence.GetShardRequest) (*persistence.GetShardResponse, error) {
	var row shardsRow
	if err := m.db.Get(&row, getShardSQLQuery, request.ShardID); err != nil {
		if err == sql.ErrNoRows {
			return nil, &workflow.EntityNotExistsError{
				Message: fmt.Sprintf("GetShard operation failed. Shard with ID %v not found. Error: %v", request.ShardID, err),
			}
		}
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("GetShard operation failed. Failed to get record. ShardId: %v. Error: %v", request.ShardID, err),
		}
	}

	clusterTransferAckLevel := make(map[string]int64)
	if err := gobDeserialize(row.ClusterTransferAckLevel, &clusterTransferAckLevel); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("GetShard operation failed. Failed to deserialize ShardInfo.ClusterTransferAckLevel. ShardId: %v. Error: %v", request.ShardID, err),
		}
	}
	if len(clusterTransferAckLevel) == 0 {
		clusterTransferAckLevel = map[string]int64{
			m.currentClusterName: row.TransferAckLevel,
		}
	}

	clusterTimerAckLevel := make(map[string]time.Time)
	if err := gobDeserialize(row.ClusterTimerAckLevel, &clusterTimerAckLevel); err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("GetShard operation failed. Failed to deserialize ShardInfo.ClusterTimerAckLevel. ShardId: %v. Error: %v", request.ShardID, err),
		}
	}
	if len(clusterTimerAckLevel) == 0 {
		clusterTimerAckLevel = map[string]time.Time{
			m.currentClusterName: row.TimerAckLevel,
		}
	}

	resp := &persistence.GetShardResponse{ShardInfo: &persistence.ShardInfo{
		ShardID:                   int(row.ShardID),
		Owner:                     row.Owner,
		RangeID:                   row.RangeID,
		StolenSinceRenew:          int(row.StolenSinceRenew),
		UpdatedAt:                 row.UpdatedAt,
		ReplicationAckLevel:       row.ReplicationAckLevel,
		TransferAckLevel:          row.TransferAckLevel,
		TimerAckLevel:             row.TimerAckLevel,
		ClusterTransferAckLevel:   clusterTransferAckLevel,
		ClusterTimerAckLevel:      clusterTimerAckLevel,
		DomainNotificationVersion: row.DomainNotificationVersion,
	}}

	return resp, nil
}

func (m *sqlShardManager) UpdateShard(request *persistence.UpdateShardRequest) error {
	row, err := shardInfoToShardsRow(*request.ShardInfo)
	if err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("UpdateShard operation failed. Error: %v", err),
		}
	}

	result, err := m.db.NamedExec(updateShardSQLQuery, &updateShardRequest{
		*row,
		request.PreviousRangeID,
	})
	if err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("UpdatedShard operation failed. Failed to update shard with ID: %v. Error: %v", request.ShardInfo.ShardID, err),
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("UpdatedShard operation failed. Failed to verify whether we successfully updated shard with ID: %v. Error: %v", request.ShardInfo.ShardID, err),
		}
	}

	switch {
	case rowsAffected == 0:
		return &persistence.ShardOwnershipLostError{
			ShardID: request.ShardInfo.ShardID,
			Msg:     fmt.Sprintf("UpdateShard operation failed. Previous range ID: ", request.PreviousRangeID),
		}
	case rowsAffected > 1:
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("UpdatedShard operation failed. Tried to update %v shards instead of one.", rowsAffected),
		}
	}

	return nil
}

func lockShard(tx *sqlx.Tx, shardID int, oldRangeID int64) error {
	var rangeID int64

	err := tx.Get(&rangeID, lockShardSQLQuery, shardID)

	if err != nil {
		if err == sql.ErrNoRows {
			return &workflow.InternalServiceError{
				Message: fmt.Sprintf("Failed to lock shard with ID %v that does not exist.", shardID),
			}
		}
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("Failed to lock shard with ID: %v. Error: %v", shardID, err),
		}
	}

	if rangeID != oldRangeID {
		return &persistence.ShardOwnershipLostError{
			ShardID: shardID,
			Msg:     fmt.Sprintf("Failed to update shard. Previous range ID: %v; new range ID: %v", oldRangeID, rangeID),
		}
	}

	return nil
}

func shardInfoToShardsRow(s persistence.ShardInfo) (*shardsRow, error) {
	clusterTransferAckLevel, err := gobSerialize(s.ClusterTransferAckLevel)
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateShard operation failed. Failed to serialize ShardInfo.ClusterTransferAckLevel. Error: %v", err),
		}
	}

	clusterTimerAckLevel, err := gobSerialize(s.ClusterTimerAckLevel)
	if err != nil {
		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateShard operation failed. Failed to serialize ShardInfo.ClusterTimerAckLevel. Error: %v", err),
		}
	}

	return &shardsRow{
		ShardID:                   int64(s.ShardID),
		Owner:                     s.Owner,
		RangeID:                   s.RangeID,
		StolenSinceRenew:          int64(s.StolenSinceRenew),
		UpdatedAt:                 s.UpdatedAt,
		ReplicationAckLevel:       s.ReplicationAckLevel,
		TransferAckLevel:          s.TransferAckLevel,
		TimerAckLevel:             s.TimerAckLevel,
		ClusterTransferAckLevel:   clusterTransferAckLevel,
		ClusterTimerAckLevel:      clusterTimerAckLevel,
		DomainNotificationVersion: s.DomainNotificationVersion,
	}, nil
}
