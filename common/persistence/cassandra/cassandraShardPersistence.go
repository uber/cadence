// Copyright (c) 2020 Uber Technologies, Inc.
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

package cassandra

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
	"github.com/uber/cadence/common/types"
)

const (
	templateShardType = `{` +
		`shard_id: ?, ` +
		`owner: ?, ` +
		`range_id: ?, ` +
		`stolen_since_renew: ?, ` +
		`updated_at: ?, ` +
		`replication_ack_level: ?, ` +
		`transfer_ack_level: ?, ` +
		`timer_ack_level: ?, ` +
		`cluster_transfer_ack_level: ?, ` +
		`cluster_timer_ack_level: ?, ` +
		`transfer_processing_queue_states: ?, ` +
		`transfer_processing_queue_states_encoding: ?, ` +
		`timer_processing_queue_states: ?, ` +
		`timer_processing_queue_states_encoding: ?, ` +
		`domain_notification_version: ?, ` +
		`cluster_replication_level: ?, ` +
		`replication_dlq_ack_level: ?, ` +
		`pending_failover_markers: ?, ` +
		`pending_failover_markers_encoding: ? ` +
		`}`

	templateCreateShardQuery = `INSERT INTO executions (` +
		`shard_id, type, domain_id, workflow_id, run_id, visibility_ts, task_id, shard, range_id)` +
		`VALUES(?, ?, ?, ?, ?, ?, ?, ` + templateShardType + `, ?) IF NOT EXISTS`

	templateGetShardQuery = `SELECT shard, range_id ` +
		`FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ?`

	templateUpdateShardQuery = `UPDATE executions ` +
		`SET shard = ` + templateShardType + `, range_id = ? ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? ` +
		`IF range_id = ?`

	templateUpdateRangeIDQuery = `UPDATE executions ` +
		`SET range_id = ? ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ? ` +
		`IF range_id = ?`
)

type (
	// Implements ShardManager
	cassandraShardPersistence struct {
		cassandraStore
		shardID            int
		currentClusterName string
	}
)

var _ p.ShardStore = (*cassandraShardPersistence)(nil)

// newShardPersistence is used to create an instance of ShardManager implementation
func newShardPersistence(
	cfg config.Cassandra,
	clusterName string,
	logger log.Logger,
) (p.ShardStore, error) {
	session, err := cassandra.CreateSession(cfg)
	if err != nil {
		return nil, err
	}

	return &cassandraShardPersistence{
		cassandraStore: cassandraStore{
			client:  gocql.NewClient(),
			session: session,
			logger:  logger,
		},
		shardID:            -1,
		currentClusterName: clusterName,
	}, nil
}

// NewShardPersistenceFromSession is used to create an instance of ShardManager implementation
// It is being used by some admin toolings
func NewShardPersistenceFromSession(
	client gocql.Client,
	session gocql.Session,
	clusterName string,
	logger log.Logger,
) p.ShardStore {
	return &cassandraShardPersistence{
		cassandraStore: cassandraStore{
			client:  client,
			session: session,
			logger:  logger,
		},
		shardID:            -1,
		currentClusterName: clusterName,
	}
}

func (d *cassandraShardPersistence) CreateShard(
	ctx context.Context,
	request *p.InternalCreateShardRequest,
) error {
	cqlNowTimestamp := p.UnixNanoToDBTimestamp(time.Now().UnixNano())
	shardInfo := request.ShardInfo
	markerData, markerEncoding := p.FromDataBlob(shardInfo.PendingFailoverMarkers)
	transferPQS, transferPQSEncoding := p.FromDataBlob(shardInfo.TransferProcessingQueueStates)
	timerPQS, timerPQSEncoding := p.FromDataBlob(shardInfo.TimerProcessingQueueStates)
	query := d.session.Query(templateCreateShardQuery,
		shardInfo.ShardID,
		rowTypeShard,
		rowTypeShardDomainID,
		rowTypeShardWorkflowID,
		rowTypeShardRunID,
		defaultVisibilityTimestamp,
		rowTypeShardTaskID,
		shardInfo.ShardID,
		shardInfo.Owner,
		shardInfo.RangeID,
		shardInfo.StolenSinceRenew,
		cqlNowTimestamp,
		shardInfo.ReplicationAckLevel,
		shardInfo.TransferAckLevel,
		shardInfo.TimerAckLevel,
		shardInfo.ClusterTransferAckLevel,
		shardInfo.ClusterTimerAckLevel,
		transferPQS,
		transferPQSEncoding,
		timerPQS,
		timerPQSEncoding,
		shardInfo.DomainNotificationVersion,
		shardInfo.ClusterReplicationLevel,
		shardInfo.ReplicationDLQAckLevel,
		markerData,
		markerEncoding,
		shardInfo.RangeID,
	).WithContext(ctx)

	previous := make(map[string]interface{})
	applied, err := query.MapScanCAS(previous)
	if err != nil {
		return convertCommonErrors(d.client, "CreateShard", err)
	}

	if !applied {
		shard := previous["shard"].(map[string]interface{})
		return &p.ShardAlreadyExistError{
			Msg: fmt.Sprintf("Shard already exists in executions table.  ShardId: %v, RangeId: %v",
				shard["shard_id"], shard["range_id"]),
		}
	}

	return nil
}

func (d *cassandraShardPersistence) GetShard(
	ctx context.Context,
	request *p.InternalGetShardRequest,
) (*p.InternalGetShardResponse, error) {
	shardID := request.ShardID
	query := d.session.Query(templateGetShardQuery,
		shardID,
		rowTypeShard,
		rowTypeShardDomainID,
		rowTypeShardWorkflowID,
		rowTypeShardRunID,
		defaultVisibilityTimestamp,
		rowTypeShardTaskID,
	).WithContext(ctx)

	result := make(map[string]interface{})
	if err := query.MapScan(result); err != nil {
		if d.client.IsNotFoundError(err) {
			return nil, &types.EntityNotExistsError{
				Message: fmt.Sprintf("Shard not found.  ShardId: %v", shardID),
			}
		}

		return nil, convertCommonErrors(d.client, "GetShard", err)
	}

	rangeID := result["range_id"].(int64)
	shard := result["shard"].(map[string]interface{})
	shardInfoRangeID := shard["range_id"].(int64)

	// check if rangeID column and rangeID field in shard column matches, if not we need to pick the larger
	// rangeID.
	if shardInfoRangeID > rangeID {
		// In this case we need to fix the rangeID column before returning the result as:
		// 1. if we return shardInfoRangeID, then later shard CAS operation will fail
		// 2. if we still return rangeID, CAS will work but rangeID will move backward which
		// result in lost tasks, corrupted workflow history, etc.

		d.logger.Warn("Corrupted shard rangeID", tag.ShardID(shardID), tag.ShardRangeID(shardInfoRangeID), tag.PreviousShardRangeID(rangeID))
		if err := d.updateRangeID(ctx, shardID, shardInfoRangeID, rangeID); err != nil {
			return nil, err
		}

		// now we know rangeID column has the same value as shardInfoRangeID
		rangeID = shardInfoRangeID
	} else {
		// no-op
		//
		// If shardInfoRangeID = rangeID, no corruption, so no action needed.
		//
		// If shardInfoRangeID < rangeID, we also don't need to do anything here as createShardInfo will ignore
		// shardInfoRangeID and return rangeID instead. Later when updating the shard, CAS can still succeed
		// as the value from rangeID columns is returned, shardInfoRangeID will also be updated to the correct value.
	}

	info := createShardInfo(d.currentClusterName, rangeID, shard)

	return &p.InternalGetShardResponse{ShardInfo: info}, nil
}

func (d *cassandraShardPersistence) updateRangeID(
	ctx context.Context,
	shardID int,
	rangeID int64,
	previousRangeID int64,
) error {
	query := d.session.Query(templateUpdateRangeIDQuery,
		rangeID,
		shardID,
		rowTypeShard,
		rowTypeShardDomainID,
		rowTypeShardWorkflowID,
		rowTypeShardRunID,
		defaultVisibilityTimestamp,
		rowTypeShardTaskID,
		previousRangeID,
	).WithContext(ctx)

	previous := make(map[string]interface{})
	applied, err := query.MapScanCAS(previous)
	if err != nil {
		return convertCommonErrors(d.client, "UpdateRangeID", err)
	}

	if !applied {
		var columns []string
		for k, v := range previous {
			columns = append(columns, fmt.Sprintf("%s=%v", k, v))
		}

		return &p.ShardOwnershipLostError{
			ShardID: d.shardID,
			Msg: fmt.Sprintf("Failed to update shard rangeID.  previous_range_id: %v, columns: (%v)",
				previousRangeID, strings.Join(columns, ",")),
		}
	}

	return nil
}

func (d *cassandraShardPersistence) UpdateShard(
	ctx context.Context,
	request *p.InternalUpdateShardRequest,
) error {
	cqlNowTimestamp := p.UnixNanoToDBTimestamp(time.Now().UnixNano())
	shardInfo := request.ShardInfo
	markerData, markerEncoding := p.FromDataBlob(shardInfo.PendingFailoverMarkers)
	transferPQS, transferPQSEncoding := p.FromDataBlob(shardInfo.TransferProcessingQueueStates)
	timerPQS, timerPQSEncoding := p.FromDataBlob(shardInfo.TimerProcessingQueueStates)

	query := d.session.Query(templateUpdateShardQuery,
		shardInfo.ShardID,
		shardInfo.Owner,
		shardInfo.RangeID,
		shardInfo.StolenSinceRenew,
		cqlNowTimestamp,
		shardInfo.ReplicationAckLevel,
		shardInfo.TransferAckLevel,
		shardInfo.TimerAckLevel,
		shardInfo.ClusterTransferAckLevel,
		shardInfo.ClusterTimerAckLevel,
		transferPQS,
		transferPQSEncoding,
		timerPQS,
		timerPQSEncoding,
		shardInfo.DomainNotificationVersion,
		shardInfo.ClusterReplicationLevel,
		shardInfo.ReplicationDLQAckLevel,
		markerData,
		markerEncoding,
		shardInfo.RangeID,
		shardInfo.ShardID,
		rowTypeShard,
		rowTypeShardDomainID,
		rowTypeShardWorkflowID,
		rowTypeShardRunID,
		defaultVisibilityTimestamp,
		rowTypeShardTaskID,
		request.PreviousRangeID,
	).WithContext(ctx)

	previous := make(map[string]interface{})
	applied, err := query.MapScanCAS(previous)
	if err != nil {
		return convertCommonErrors(d.client, "UpdateShard", err)
	}

	if !applied {
		var columns []string
		for k, v := range previous {
			columns = append(columns, fmt.Sprintf("%s=%v", k, v))
		}

		return &p.ShardOwnershipLostError{
			ShardID: d.shardID,
			Msg: fmt.Sprintf("Failed to update shard.  previous_range_id: %v, columns: (%v)",
				request.PreviousRangeID, strings.Join(columns, ",")),
		}
	}

	return nil
}
