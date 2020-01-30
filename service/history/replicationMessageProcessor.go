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

package history

import (
	"context"

	h "github.com/uber/cadence/.gen/go/history"
	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/xdc"
)

type (
	// DLQTaskHandler is the interface handles domain DLQ messages
	replicationMessageHandler interface {
		readMessages(
			ctx context.Context,
			sourceCluster string,
			lastMessageID int64,
			pageSize int,
			pageToken []byte,
		) ([]*replicator.ReplicationTask, []byte, error)
		purgeMessages(
			sourceCluster string,
			lastMessageID int64,
		) error
		mergeMessages(
			ctx context.Context,
			sourceCluster string,
			lastMessageID int64,
			pageSize int,
			pageToken []byte,
		) ([]byte, error)
	}

	replicationMessageHandlerImpl struct {
		replicationTaskHandler replicationTaskHandler
		shard                  ShardContext
		logger                 log.Logger
	}
)

func newReplicationMessageHandler(
	shard ShardContext,
	historyEngine Engine,
) replicationMessageHandler {

	currentCluster := shard.GetClusterMetadata().GetCurrentClusterName()
	nDCHistoryResender := xdc.NewNDCHistoryResender(
		shard.GetDomainCache(),
		shard.GetService().GetClientBean().GetRemoteAdminClient(currentCluster),
		func(ctx context.Context, request *h.ReplicateEventsV2Request) error {
			return shard.GetService().GetHistoryClient().ReplicateEventsV2(ctx, request)
		},
		shard.GetService().GetPayloadSerializer(),
		shard.GetLogger(),
	)
	historyRereplicator := xdc.NewHistoryRereplicator(
		currentCluster,
		shard.GetDomainCache(),
		shard.GetService().GetClientBean().GetRemoteAdminClient(currentCluster),
		func(ctx context.Context, request *h.ReplicateRawEventsRequest) error {
			return shard.GetService().GetHistoryClient().ReplicateRawEvents(ctx, request)
		},
		shard.GetService().GetPayloadSerializer(),
		replicationTimeout,
		shard.GetLogger(),
	)
	replicationTaskHandler := newReplicationTaskHandler(
		shard.GetClusterMetadata().GetCurrentClusterName(),
		shard.GetDomainCache(),
		nDCHistoryResender,
		historyRereplicator,
		historyEngine,
		shard.GetMetricsClient(),
		shard.GetLogger(),
	)
	return &replicationMessageHandlerImpl{
		shard:                  shard,
		replicationTaskHandler: replicationTaskHandler,
		logger:                 shard.GetLogger(),
	}
}

func (r *replicationMessageHandlerImpl) readMessages(
	ctx context.Context,
	sourceCluster string,
	lastMessageID int64,
	pageSize int,
	pageToken []byte,
) ([]*replicator.ReplicationTask, []byte, error) {

	ackLevel := r.shard.GetReplicatorDLQAckLevel()
	resp, err := r.shard.GetExecutionManager().GetReplicationTasksFromDLQ(&persistence.GetReplicationTasksFromDLQRequest{
		SourceClusterName: sourceCluster,
		GetReplicationTasksRequest: persistence.GetReplicationTasksRequest{
			ReadLevel:     ackLevel,
			MaxReadLevel:  lastMessageID,
			BatchSize:     pageSize,
			NextPageToken: pageToken,
		},
	})
	if err != nil {
		return nil, nil, err
	}

	remoteAdminClient := r.shard.GetService().GetClientBean().GetRemoteAdminClient(sourceCluster)
	taskInfo := make([]*replicator.ReplicationTaskInfo, len(resp.Tasks))
	for _, task := range resp.Tasks {
		taskInfo = append(taskInfo, &replicator.ReplicationTaskInfo{
			DomainID:     common.StringPtr(task.GetDomainID()),
			WorkflowID:   common.StringPtr(task.GetWorkflowID()),
			RunID:        common.StringPtr(task.GetRunID()),
			TaskType:     common.Int16Ptr(int16(task.GetTaskType())),
			TaskID:       common.Int64Ptr(task.GetTaskID()),
			Version:      common.Int64Ptr(task.GetVersion()),
			FirstEventID: common.Int64Ptr(task.FirstEventID),
			NextEventID:  common.Int64Ptr(task.NextEventID),
			ScheduledID:  common.Int64Ptr(task.ScheduledID),
		})
	}
	dlqResponse, err := remoteAdminClient.GetDLQReplicationMessages(
		ctx,
		&replicator.GetDLQReplicationMessagesRequest{
			TaskInfos: taskInfo,
		},
	)
	if err != nil {
		return nil, nil, err
	}
	return dlqResponse.ReplicationTasks, resp.NextPageToken, nil
}

func (r *replicationMessageHandlerImpl) purgeMessages(
	sourceCluster string,
	lastMessageID int64,
) error {

	ackLevel := r.shard.GetReplicatorDLQAckLevel()
	err := r.shard.GetExecutionManager().RangeDeleteReplicationTaskFromDLQ(
		&persistence.RangeDeleteReplicationTaskFromDLQRequest{
			SourceClusterName:    sourceCluster,
			ExclusiveBeginTaskID: ackLevel,
			InclusiveEndTaskID:   lastMessageID,
		},
	)
	if err != nil {
		return err
	}

	if err = r.shard.UpdateReplicatorDLQAckLevel(lastMessageID); err != nil {
		r.logger.Error("Failed to purge history replication message", tag.Error(err))
	}
	return nil
}

func (r *replicationMessageHandlerImpl) mergeMessages(
	ctx context.Context,
	sourceCluster string,
	lastMessageID int64,
	pageSize int,
	pageToken []byte,
) ([]byte, error) {

	ackLevel := r.shard.GetReplicatorDLQAckLevel()
	resp, err := r.shard.GetExecutionManager().GetReplicationTasksFromDLQ(&persistence.GetReplicationTasksFromDLQRequest{
		SourceClusterName: sourceCluster,
		GetReplicationTasksRequest: persistence.GetReplicationTasksRequest{
			ReadLevel:     ackLevel,
			MaxReadLevel:  lastMessageID,
			BatchSize:     pageSize,
			NextPageToken: pageToken,
		},
	})
	if err != nil {
		return nil, err
	}

	remoteAdminClient := r.shard.GetService().GetClientBean().GetRemoteAdminClient(sourceCluster)
	taskInfo := make([]*replicator.ReplicationTaskInfo, len(resp.Tasks))
	for _, task := range resp.Tasks {
		taskInfo = append(taskInfo, &replicator.ReplicationTaskInfo{
			DomainID:     common.StringPtr(task.GetDomainID()),
			WorkflowID:   common.StringPtr(task.GetWorkflowID()),
			RunID:        common.StringPtr(task.GetRunID()),
			TaskType:     common.Int16Ptr(int16(task.GetTaskType())),
			TaskID:       common.Int64Ptr(task.GetTaskID()),
			Version:      common.Int64Ptr(task.GetVersion()),
			FirstEventID: common.Int64Ptr(task.FirstEventID),
			NextEventID:  common.Int64Ptr(task.NextEventID),
			ScheduledID:  common.Int64Ptr(task.ScheduledID),
		})
	}
	dlqResponse, err := remoteAdminClient.GetDLQReplicationMessages(
		ctx,
		&replicator.GetDLQReplicationMessagesRequest{
			TaskInfos: taskInfo,
		},
	)
	if err != nil {
		return nil, err
	}

	for _, task := range dlqResponse.GetReplicationTasks() {
		r.replicationTaskHandler.process(
			sourceCluster,
			task,
			true,
		)
	}

	err = r.shard.GetExecutionManager().RangeDeleteReplicationTaskFromDLQ(
		&persistence.RangeDeleteReplicationTaskFromDLQRequest{
			SourceClusterName:    sourceCluster,
			ExclusiveBeginTaskID: ackLevel,
			InclusiveEndTaskID:   lastMessageID,
		},
	)
	if err != nil {
		return nil, err
	}

	if err = r.shard.UpdateReplicatorDLQAckLevel(lastMessageID); err != nil {
		r.logger.Error("Failed to purge history replication message", tag.Error(err))
	}
	return resp.NextPageToken, nil
}
