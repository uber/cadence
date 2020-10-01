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

package shard

import (
	"context"
	"fmt"
	"github.com/gocql/gocql"
	workflow "github.com/uber/cadence/.gen/go/shared"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra"
	"github.com/uber/cadence/common/persistence/stores"
	"github.com/uber/cadence/common/service/config"
	"github.com/uber/cadence/common/types"
	"log"
	"strings"
	"time"
)

const (
	templateGetShardQuery = `SELECT shard ` +
		`FROM executions ` +
		`WHERE shard_id = ? ` +
		`and type = ? ` +
		`and domain_id = ? ` +
		`and workflow_id = ? ` +
		`and run_id = ? ` +
		`and visibility_ts = ? ` +
		`and task_id = ?`

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
)

type (
	store struct {
		session *gocql.Session
		logger  log.Logger
		currentClusterName string
	}
)

func NewShardStoreFromConfig(cfg config.Cassandra, clusterName string, logger log.Logger) (Store, error) {
	cluster := cassandra.NewCassandraCluster(cfg)
	cluster.ProtoVersion = stores.CassandraProtoVersion
	cluster.Consistency = gocql.LocalQuorum
	cluster.SerialConsistency = gocql.LocalSerial
	cluster.Timeout = stores.CassandraDefaultSessionTimeout

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return &store{
		session: session,
		logger: logger,
		currentClusterName: clusterName,
	}, nil
}

func NewShardStoreFromSession(session *gocql.Session, clusterName string, logger log.Logger) Store {
	return &store{
		session: session,
		logger: logger,
		currentClusterName: clusterName,
	}
}

func (s *store) CreateShard(
	_ context.Context,
	request *CreateShardRequest,
) error {
	cqlNowTimestamp := stores.CassandraUnixNanoToDBTimestamp(time.Now().UnixNano())
	shardInfo := request.ShardInfo
	markerData, markerEncoding := types.FromDataBlob(shardInfo.PendingFailoverMarkers)
	transferPQS, transferPQSEncoding := types.FromDataBlob(shardInfo.TransferProcessingQueueStates)
	timerPQS, timerPQSEncoding := types.FromDataBlob(shardInfo.TimerProcessingQueueStates)
	query := s.session.Query(templateCreateShardQuery,
		shardInfo.ShardID,
		stores.CassandraRowTypeShard,
		stores.CassandraRowTypeShardDomainID,
		stores.CassandraRowTypeShardWorkflowID,
		stores.CassandraRowTypeShardRunID,
		stores.CassandraDefaultVisibilityTimestamp,
		stores.CassandraRowTypeShardTaskID,
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
		shardInfo.RangeID)

	previous := make(map[string]interface{})
	applied, err := query.MapScanCAS(previous)
	if err != nil {
		if stores.CassandraIsThrottlingError(err) {
			return &workflow.ServiceBusyError{
				Message: fmt.Sprintf("CreateShard operation failed. Error: %v", err),
			}
		}
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("CreateShard operation failed. Error: %v", err),
		}
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

func (s *store) GetShard(
	_ context.Context,
	request *GetShardRequest,
) (*GetShardResponse, error) {
	shardID := request.ShardID
	query := s.session.Query(templateGetShardQuery,
		shardID,
		stores.CassandraRowTypeShard,
		stores.CassandraRowTypeShardDomainID,
		stores.CassandraRowTypeShardWorkflowID,
		stores.CassandraRowTypeShardRunID,
		stores.CassandraDefaultVisibilityTimestamp,
		stores.CassandraRowTypeShardTaskID)

	result := make(map[string]interface{})
	if err := query.MapScan(result); err != nil {
		if err == gocql.ErrNotFound {
			return nil, &workflow.EntityNotExistsError{
				Message: fmt.Sprintf("Shard not found.  ShardId: %v", shardID),
			}
		} else if stores.CassandraIsThrottlingError(err) {
			return nil, &workflow.ServiceBusyError{
				Message: fmt.Sprintf("GetShard operation failed. Error: %v", err),
			}
		}

		return nil, &workflow.InternalServiceError{
			Message: fmt.Sprintf("GetShard operation failed. Error: %v", err),
		}
	}

	info := createShardInfo(s.currentClusterName, result["shard"].(map[string]interface{}))

	return &GetShardResponse{ShardInfo: info}, nil
}

func createShardInfo(
	currentCluster string,
	result map[string]interface{},
) *Info {

	var pendingFailoverMarkersRawData []byte
	var pendingFailoverMarkersEncoding string
	var transferProcessingQueueStatesRawData []byte
	var transferProcessingQueueStatesEncoding string
	var timerProcessingQueueStatesRawData []byte
	var timerProcessingQueueStatesEncoding string
	info := &Info{}
	for k, v := range result {
		switch k {
		case "shard_id":
			info.ShardID = v.(int)
		case "owner":
			info.Owner = v.(string)
		case "range_id":
			info.RangeID = v.(int64)
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
	info.PendingFailoverMarkers = types.NewDataBlob(
		pendingFailoverMarkersRawData,
		types.EncodingType(pendingFailoverMarkersEncoding),
	)
	info.TransferProcessingQueueStates = types.NewDataBlob(
		transferProcessingQueueStatesRawData,
		types.EncodingType(transferProcessingQueueStatesEncoding),
	)
	info.TimerProcessingQueueStates = types.NewDataBlob(
		timerProcessingQueueStatesRawData,
		types.EncodingType(timerProcessingQueueStatesEncoding),
	)

	return info
}

func (s *store) UpdateShard(
	_ context.Context,
	request *UpdateShardRequest,
) error {
	cqlNowTimestamp := stores.CassandraUnixNanoToDBTimestamp(time.Now().UnixNano())
	shardInfo := request.ShardInfo
	markerData, markerEncoding := types.FromDataBlob(shardInfo.PendingFailoverMarkers)
	transferPQS, transferPQSEncoding := types.FromDataBlob(shardInfo.TransferProcessingQueueStates)
	timerPQS, timerPQSEncoding := types.FromDataBlob(shardInfo.TimerProcessingQueueStates)

	query := s.session.Query(templateUpdateShardQuery,
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
		stores.CassandraRowTypeShard,
		stores.CassandraRowTypeShardDomainID,
		stores.CassandraRowTypeShardWorkflowID,
		stores.CassandraRowTypeShardRunID,
		stores.CassandraDefaultVisibilityTimestamp,
		stores.CassandraRowTypeShardTaskID,
		request.PreviousRangeID)

	previous := make(map[string]interface{})
	applied, err := query.MapScanCAS(previous)
	if err != nil {
		if stores.CassandraIsThrottlingError(err) {
			return &workflow.ServiceBusyError{
				Message: fmt.Sprintf("UpdateShard operation failed. Error: %v", err),
			}
		}
		return &workflow.InternalServiceError{
			Message: fmt.Sprintf("UpdateShard operation failed. Error: %v", err),
		}
	}

	if !applied {
		var columns []string
		for k, v := range previous {
			columns = append(columns, fmt.Sprintf("%s=%v", k, v))
		}

		return &p.ShardOwnershipLostError{
			ShardID: request.ShardInfo.ShardID,
			Msg: fmt.Sprintf("Failed to update shard.  previous_range_id: %v, columns: (%v)",
				request.PreviousRangeID, strings.Join(columns, ",")),
		}
	}

	return nil
}

func (s *store) GetName() string {
	return stores.CassandraPersistenceName
}

// Close releases the underlying resources held by this object
func (s *store) Close() {
	if s.session != nil {
		s.session.Close()
	}
}