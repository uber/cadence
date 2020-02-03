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

//go:generate mockgen -copyright_file ../../LICENSE -package $GOPACKAGE -source $GOFILE -destination replicationMessageHandler_mock.go -self_package github.com/uber/cadence/service/history

package history

import (
	"context"

	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
)

type (
	// replicationMessageHandler is the interface handles replication DLQ messages
	replicationDLQHandler interface {
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

	replicationDLQHandlerImpl struct {
		replicationTaskExecutor replicationTaskExecutor
		shard                   ShardContext
		logger                  log.Logger
	}
)

func newReplicationDLQHandler(
	shard ShardContext,
	replicationTaskExecutor replicationTaskExecutor,
) replicationDLQHandler {

	return &replicationDLQHandlerImpl{
		shard:                   shard,
		replicationTaskExecutor: replicationTaskExecutor,
		logger:                  shard.GetLogger(),
	}
}

func (r *replicationDLQHandlerImpl) readMessages(
	ctx context.Context,
	sourceCluster string,
	lastMessageID int64,
	pageSize int,
	pageToken []byte,
) ([]*replicator.ReplicationTask, []byte, error) {

	ackLevel := r.shard.GetReplicatorDLQAckLevel(sourceCluster)
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

func (r *replicationDLQHandlerImpl) purgeMessages(
	sourceCluster string,
	lastMessageID int64,
) error {

	ackLevel := r.shard.GetReplicatorDLQAckLevel(sourceCluster)
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

	if err = r.shard.UpdateReplicatorDLQAckLevel(
		sourceCluster,
		lastMessageID,
	); err != nil {
		r.logger.Error("Failed to purge history replication message", tag.Error(err))
	}
	return nil
}

func (r *replicationDLQHandlerImpl) mergeMessages(
	ctx context.Context,
	sourceCluster string,
	lastMessageID int64,
	pageSize int,
	pageToken []byte,
) ([]byte, error) {

	ackLevel := r.shard.GetReplicatorDLQAckLevel(sourceCluster)
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
		if _, err := r.replicationTaskExecutor.execute(
			sourceCluster,
			task,
			true,
		); err != nil {
			return nil, err
		}
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

	if err = r.shard.UpdateReplicatorDLQAckLevel(
		sourceCluster,
		lastMessageID,
	); err != nil {
		r.logger.Error("Failed to purge history replication message", tag.Error(err))
	}
	return resp.NextPageToken, nil
}
