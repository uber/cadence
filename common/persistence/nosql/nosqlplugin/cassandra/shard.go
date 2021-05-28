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

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
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
		`cross_cluster_processing_queue_states: ?, ` +
		`cross_cluster_processing_queue_states_encoding: ?, ` +
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

// InsertShard creates a new shard, return error is there is any.
// When error is nil, return applied=true if there is a conflict, and return the conflicted row as previous
func (db *cdb) InsertShard(ctx context.Context, row *nosqlplugin.ShardRow) (*nosqlplugin.ConflictedShardRow, error) {
	cqlNowTimestamp := persistence.UnixNanoToDBTimestamp(time.Now().UnixNano())
	markerData, markerEncoding := persistence.FromDataBlob(row.PendingFailoverMarkers)
	transferPQS, transferPQSEncoding := persistence.FromDataBlob(row.TransferProcessingQueueStates)
	crossClusterPQS, crossClusterPQSEncoding := persistence.FromDataBlob(row.CrossClusterProcessingQueueStates)
	timerPQS, timerPQSEncoding := persistence.FromDataBlob(row.TimerProcessingQueueStates)
	query := db.session.Query(templateCreateShardQuery,
		row.ShardID,
		rowTypeShard,
		rowTypeShardDomainID,
		rowTypeShardWorkflowID,
		rowTypeShardRunID,
		defaultVisibilityTimestamp,
		rowTypeShardTaskID,
		row.ShardID,
		row.Owner,
		row.RangeID,
		row.StolenSinceRenew,
		cqlNowTimestamp,
		row.ReplicationAckLevel,
		row.TransferAckLevel,
		row.TimerAckLevel,
		row.ClusterTransferAckLevel,
		row.ClusterTimerAckLevel,
		transferPQS,
		transferPQSEncoding,
		crossClusterPQS,
		crossClusterPQSEncoding,
		timerPQS,
		timerPQSEncoding,
		row.DomainNotificationVersion,
		row.ClusterReplicationLevel,
		row.ReplicationDLQAckLevel,
		markerData,
		markerEncoding,
		row.RangeID,
	).WithContext(ctx)

	previous := make(map[string]interface{})
	applied, err := query.MapScanCAS(previous)
	if err != nil {
		return nil, err
	}

	if !applied {
		return convertToConflictedShardRow(row.ShardID, row.RangeID, previous), errConditionFailed
	}

	return nil, nil
}

func convertToConflictedShardRow(shardID int, previousRangeID int64, previous map[string]interface{}) *nosqlplugin.ConflictedShardRow {
	var columns []string
	for k, v := range previous {
		columns = append(columns, fmt.Sprintf("%s=%v", k, v))
	}
	return &nosqlplugin.ConflictedShardRow{
		ShardID:         shardID,
		PreviousRangeID: previousRangeID,
		Details:         strings.Join(columns, ","),
	}
}

// SelectShard gets a shard
func (db *cdb) SelectShard(ctx context.Context, shardID int, currentClusterName string) (int64, *nosqlplugin.ShardRow, error) {
	query := db.session.Query(templateGetShardQuery,
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
		return 0, nil, err
	}

	rangeID := result["range_id"].(int64)
	shard := result["shard"].(map[string]interface{})
	shardInfoRangeID := shard["range_id"].(int64)
	return rangeID, convertToShardInfo(currentClusterName, shardInfoRangeID, shard), nil
}

func convertToShardInfo(
	currentCluster string,
	rangeID int64,
	shard map[string]interface{},
) *nosqlplugin.ShardRow {

	var pendingFailoverMarkersRawData []byte
	var pendingFailoverMarkersEncoding string
	var transferProcessingQueueStatesRawData []byte
	var transferProcessingQueueStatesEncoding string
	var crossClusterProcessingQueueStatesRawData []byte
	var crossClusterProcessingQueueStatesEncoding string
	var timerProcessingQueueStatesRawData []byte
	var timerProcessingQueueStatesEncoding string
	info := &persistence.InternalShardInfo{}
	info.RangeID = rangeID
	for k, v := range shard {
		switch k {
		case "shard_id":
			info.ShardID = v.(int)
		case "owner":
			info.Owner = v.(string)
		case "stolen_since_renew":
			info.StolenSinceRenew = v.(int)
		case "updated_at":
			info.UpdatedAt = v.(time.Time)
		case "replication_ack_level":
			info.ReplicationAckLevel = v.(int64)
		case "transfer_ack_level":
			info.TransferAckLevel = v.(int64)
		case "timer_ack_level":
			info.TimerAckLevel = v.(time.Time)
		case "cluster_transfer_ack_level":
			info.ClusterTransferAckLevel = v.(map[string]int64)
		case "cluster_timer_ack_level":
			info.ClusterTimerAckLevel = v.(map[string]time.Time)
		case "transfer_processing_queue_states":
			transferProcessingQueueStatesRawData = v.([]byte)
		case "transfer_processing_queue_states_encoding":
			transferProcessingQueueStatesEncoding = v.(string)
		case "cross_cluster_processing_queue_states":
			crossClusterProcessingQueueStatesRawData = v.([]byte)
		case "cross_cluster_processing_queue_states_encoding":
			crossClusterProcessingQueueStatesEncoding = v.(string)
		case "timer_processing_queue_states":
			timerProcessingQueueStatesRawData = v.([]byte)
		case "timer_processing_queue_states_encoding":
			timerProcessingQueueStatesEncoding = v.(string)
		case "domain_notification_version":
			info.DomainNotificationVersion = v.(int64)
		case "cluster_replication_level":
			info.ClusterReplicationLevel = v.(map[string]int64)
		case "replication_dlq_ack_level":
			info.ReplicationDLQAckLevel = v.(map[string]int64)
		case "pending_failover_markers":
			pendingFailoverMarkersRawData = v.([]byte)
		case "pending_failover_markers_encoding":
			pendingFailoverMarkersEncoding = v.(string)
		}
	}

	if info.ClusterTransferAckLevel == nil {
		info.ClusterTransferAckLevel = map[string]int64{
			currentCluster: info.TransferAckLevel,
		}
	}
	if info.ClusterTimerAckLevel == nil {
		info.ClusterTimerAckLevel = map[string]time.Time{
			currentCluster: info.TimerAckLevel,
		}
	}
	if info.ClusterReplicationLevel == nil {
		info.ClusterReplicationLevel = make(map[string]int64)
	}
	if info.ReplicationDLQAckLevel == nil {
		info.ReplicationDLQAckLevel = make(map[string]int64)
	}
	info.PendingFailoverMarkers = persistence.NewDataBlob(
		pendingFailoverMarkersRawData,
		common.EncodingType(pendingFailoverMarkersEncoding),
	)
	info.TransferProcessingQueueStates = persistence.NewDataBlob(
		transferProcessingQueueStatesRawData,
		common.EncodingType(transferProcessingQueueStatesEncoding),
	)
	info.CrossClusterProcessingQueueStates = persistence.NewDataBlob(
		crossClusterProcessingQueueStatesRawData,
		common.EncodingType(crossClusterProcessingQueueStatesEncoding),
	)
	info.TimerProcessingQueueStates = persistence.NewDataBlob(
		timerProcessingQueueStatesRawData,
		common.EncodingType(timerProcessingQueueStatesEncoding),
	)

	return info
}

// UpdateRangeID updates the rangeID, return error is there is any
// When error is nil, return applied=true if there is a conflict, and return the conflicted row as previous
func (db *cdb) UpdateRangeID(ctx context.Context, shardID int, rangeID int64, previousRangeID int64) (*nosqlplugin.ConflictedShardRow, error) {
	query := db.session.Query(templateUpdateRangeIDQuery,
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
		return nil, err
	}

	if !applied {
		return convertToConflictedShardRow(shardID, previousRangeID, previous), errConditionFailed
	}

	return nil, nil
}

// UpdateShard updates a shard, return error is there is any.
// When error is nil, return applied=true if there is a conflict, and return the conflicted row as previous
func (db *cdb) UpdateShard(ctx context.Context, row *nosqlplugin.ShardRow, previousRangeID int64) (*nosqlplugin.ConflictedShardRow, error) {
	cqlNowTimestamp := persistence.UnixNanoToDBTimestamp(time.Now().UnixNano())
	markerData, markerEncoding := persistence.FromDataBlob(row.PendingFailoverMarkers)
	transferPQS, transferPQSEncoding := persistence.FromDataBlob(row.TransferProcessingQueueStates)
	crossClusterPQS, crossClusterPQSEncoding := persistence.FromDataBlob(row.CrossClusterProcessingQueueStates)
	timerPQS, timerPQSEncoding := persistence.FromDataBlob(row.TimerProcessingQueueStates)

	query := db.session.Query(templateUpdateShardQuery,
		row.ShardID,
		row.Owner,
		row.RangeID,
		row.StolenSinceRenew,
		cqlNowTimestamp,
		row.ReplicationAckLevel,
		row.TransferAckLevel,
		row.TimerAckLevel,
		row.ClusterTransferAckLevel,
		row.ClusterTimerAckLevel,
		transferPQS,
		transferPQSEncoding,
		crossClusterPQS,
		crossClusterPQSEncoding,
		timerPQS,
		timerPQSEncoding,
		row.DomainNotificationVersion,
		row.ClusterReplicationLevel,
		row.ReplicationDLQAckLevel,
		markerData,
		markerEncoding,
		row.RangeID,
		row.ShardID,
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
		return nil, err
	}

	if !applied {
		return convertToConflictedShardRow(row.ShardID, previousRangeID, previous), errConditionFailed
	}

	return nil, nil
}
