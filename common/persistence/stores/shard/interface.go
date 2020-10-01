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
	"time"

	"github.com/uber/cadence/common/types"
)

type (
	// TransferFailoverLevel contains corresponding start / end level
	TransferFailoverLevel struct {
		StartTime    time.Time
		MinLevel     int64
		CurrentLevel int64
		MaxLevel     int64
		DomainIDs    map[string]struct{}
	}

	// TimerFailoverLevel contains domain IDs and corresponding start / end level
	TimerFailoverLevel struct {
		StartTime    time.Time
		MinLevel     time.Time
		CurrentLevel time.Time
		MaxLevel     time.Time
		DomainIDs    map[string]struct{}
	}

	// Info describes a shard
	Info struct {
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
		TransferProcessingQueueStates *types.DataBlob                  `json:"transfer_processing_queue_states"`
		TimerProcessingQueueStates    *types.DataBlob                  `json:"timer_processing_queue_states"`
		TransferFailoverLevels        map[string]TransferFailoverLevel // uuid -> TransferFailoverLevel
		TimerFailoverLevels           map[string]TimerFailoverLevel    // uuid -> TimerFailoverLevel
		ClusterReplicationLevel       map[string]int64                 `json:"cluster_replication_level"`
		DomainNotificationVersion     int64                            `json:"domain_notification_version"`
		PendingFailoverMarkers        *types.DataBlob                  `json:"pending_failover_markers"`
	}

	// CreateShardRequest is used to create a shard in executions table
	CreateShardRequest struct {
		ShardInfo *Info
	}

	// GetShardRequest is used to get shard information
	GetShardRequest struct {
		ShardID int
	}

	// GetShardResponse is the response to GetShard
	GetShardResponse struct {
		ShardInfo *Info
	}

	// UpdateShardRequest  is used to update shard information
	UpdateShardRequest struct {
		ShardInfo       *Info
		PreviousRangeID int64
	}

	// Store supports persistence backed operations on shards
	Store interface {
		types.Closeable
		GetName() string
		CreateShard(ctx context.Context, request *CreateShardRequest) error
		GetShard(ctx context.Context, request *GetShardRequest) (*GetShardResponse, error)
		UpdateShard(ctx context.Context, request *UpdateShardRequest) error
	}
)
