package stores

import (
	"context"
	"github.com/uber/cadence/common/types"
	"time"
)

type (
	// ShardInfo describes a shard
	ShardInfo struct {
		ShardID                       int                              `json:"shard_id"`
		Owner                         string                           `json:"owner"`
		RangeID                       int64                            `json:"range_id"`
		StolenSinceRenew              int                              `json:"stolen_since_renew"`
		UpdatedAt                     time.Time                        `json:"updated_at"`
		ReplicationAckLevel           int64                            `json:"replication_ack_level"`
		ReplicationDLQAckLevel        map[string]int64                 `json:"replication_dlq_ack_level"`
		TransferAckLevel              int64                            `json:"transfer_ack_level"`
		TimerAckLevel                 time.Time                        `json:"timer_ack_level"`
		ClusterTransferAckLevel       map[string]int64                 `json:"cluster_transfer_ack_level"`
		ClusterTimerAckLevel          map[string]time.Time             `json:"cluster_timer_ack_level"`
		TransferProcessingQueueStates *DataBlob                        `json:"transfer_processing_queue_states"`
		TimerProcessingQueueStates    *DataBlob                        `json:"timer_processing_queue_states"`
		TransferFailoverLevels        map[string]TransferFailoverLevel // uuid -> TransferFailoverLevel
		TimerFailoverLevels           map[string]TimerFailoverLevel    // uuid -> TimerFailoverLevel
		ClusterReplicationLevel       map[string]int64                 `json:"cluster_replication_level"`
		DomainNotificationVersion     int64                            `json:"domain_notification_version"`
		PendingFailoverMarkers        *DataBlob                        `json:"pending_failover_markers"`
	}

	// CreateShardRequest is used to create a shard in executions table
	CreateShardRequest struct {
		ShardInfo *ShardInfo
	}

	ShardStore interface {
		types.Closeable
		GetName() string
		CreateShard(ctx context.Context, request *CreateShardRequest) error
		GetShard(ctx context.Context, request *GetShardRequest) (*GetShardResponse, error)
		UpdateShard(ctx context.Context, request *UpdateShardRequest) error
	}
)
