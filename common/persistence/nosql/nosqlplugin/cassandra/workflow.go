// Copyright (c) 2021 Uber Technologies, Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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

	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin"
	"github.com/uber/cadence/common/persistence/nosql/nosqlplugin/cassandra/gocql"
)

func (db *cdb) InsertWorkflowExecutionWithTasks(
	ctx context.Context,
	currentWorkflowRequest *nosqlplugin.CurrentWorkflowWriteRequest,
	execution *nosqlplugin.WorkflowExecutionRow,
	transferTasks []*nosqlplugin.TransferTask,
	crossClusterTasks []*nosqlplugin.CrossClusterTask,
	replicationTasks []*nosqlplugin.ReplicationTask,
	timerTasks []*nosqlplugin.TimerTask,
	activityInfoMap map[int64]*persistence.InternalActivityInfo,
	timerInfoMap map[string]*persistence.TimerInfo,
	childWorkflowInfoMap map[int64]*persistence.InternalChildExecutionInfo,
	requestCancelInfoMap map[int64]*persistence.RequestCancelInfo,
	signalInfoMap map[int64]*persistence.SignalInfo,
	signalRequestedIDs []string,
	shardCondition *nosqlplugin.ShardCondition,
) (*nosqlplugin.ConditionFailureReason, error) {
	shardID := shardCondition.ShardID
	domainID := execution.DomainID
	workflowID := execution.WorkflowID
	runID := execution.RunID

	batch := db.session.NewBatch(gocql.LoggedBatch).WithContext(ctx)

	err := db.createOrUpdateCurrentWorkflow(batch, shardID, domainID, workflowID, currentWorkflowRequest)
	if err != nil {
		return nil, err
	}

	err = db.createWorkflowExecution(batch, shardID, domainID, workflowID, runID, execution)
	if err != nil {
		return nil, err
	}

	err = db.updateActivityInfos(batch, shardID, domainID, workflowID, runID, activityInfoMap, nil)
	if err != nil {
		return nil, err
	}
	err = db.updateTimerInfos(batch, shardID, domainID, workflowID, runID, timerInfoMap, nil)
	if err != nil {
		return nil, err
	}
	err = db.updateChildExecutionInfos(batch, shardID, domainID, workflowID, runID, childWorkflowInfoMap, nil)
	if err != nil {
		return nil, err
	}
	err = db.updateRequestCancelInfos(batch, shardID, domainID, workflowID, runID, requestCancelInfoMap, nil)
	if err != nil {
		return nil, err
	}
	err = db.updateSignalInfos(batch, shardID, domainID, workflowID, runID, signalInfoMap, nil)
	if err != nil {
		return nil, err
	}
	err = db.updateSignalsRequested(batch, shardID, domainID, workflowID, runID, signalRequestedIDs, nil)
	if err != nil {
		return nil, err
	}

	err = db.createTransferTasks(batch, shardID, domainID, workflowID, runID, transferTasks)
	if err != nil {
		return nil, err
	}
	err = db.createReplicationTasks(batch, shardID, domainID, workflowID, runID, replicationTasks)
	if err != nil {
		return nil, err
	}
	err = db.createCrossClusterTasks(batch, shardID, domainID, workflowID, runID, crossClusterTasks)
	if err != nil {
		return nil, err
	}
	err = db.createTimerTasks(batch, shardID, domainID, workflowID, runID, timerTasks)
	if err != nil {
		return nil, err
	}

	err = db.assertShardRangeID(batch, shardID, shardCondition.RangeID)
	if err != nil {
		return nil, err
	}

	return db.executeCreateWorkflowBatchTransaction(batch, currentWorkflowRequest, execution, shardCondition)
}
