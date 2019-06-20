// Copyright (c) 2017 Uber Technologies, Inc.
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
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
)

type (
	nDCTransactionMgrImpl struct {
		shard        ShardContext
		context      workflowExecutionContext
		mutableState mutableState
		task         nDCReplicationTask

		currentContext      workflowExecutionContext
		currentMutableState mutableState

		metricsClient metrics.Client
		logger        log.Logger
	}
)

func newNDCTransactionMgr(
	shard ShardContext,
	context workflowExecutionContext,
	mutableState mutableState,
	task nDCReplicationTask,
	metricsClient metrics.Client,
	logger log.Logger,
) *nDCTransactionMgrImpl {
	return &nDCTransactionMgrImpl{
		shard:         shard,
		context:       context,
		mutableState:  mutableState,
		task:          task,
		metricsClient: metricsClient,
		logger:        logger,
	}
}

func (r *nDCTransactionMgrImpl) appendFirstEventBatch() error {
	if _, _, err := r.context.appendFirstBatchEventsForStandby(
		r.mutableState,
		r.task.getEvents(),
	); err != nil {
		return err
	}
	return nil
}

func (r *nDCTransactionMgrImpl) appendNonFirstEventBatch() error {
	transactionID, err := r.shard.GetNextTransferTaskID()
	if err != nil {
		return err
	}
	historySize, _, err := r.context.appendHistoryEvents(
		r.task.getEvents(),
		transactionID,
		true,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	r.mutableState.IncrementHistorySize(historySize)
	panic("implement this")
}

func (r *nDCTransactionMgrImpl) createWorkflowAsZombie() {
	// create the workflow as zombie
	// i.e. current workflow's version is higher
	// or current workflow's version is the same, but last task ID is higher
}

func (r *nDCTransactionMgrImpl) createWorkflowWithConflict() {
	// create the workflow as current workflow
	// i.e. current workflow's version is lower
	// or current workflow's version is the same, but last task ID is lower
}

//func (r *nDCTransactionMgrImpl) createWorkflow() {
//	// ??
//}

func (r nDCTransactionMgrImpl) updateWorkflowNonCurrentBranch() {
	// append history to branch
	// update the workflow as is
}

func (r nDCTransactionMgrImpl) updateWorkflowCurrentBranchWithoutRebuildAsZombie() {
	// append history to branch
	// update the workflow as is
}

func (r nDCTransactionMgrImpl) updateWorkflowCurrentBranchWithoutRebuildWithConflict() {
	// append history to branch
	// update the workflow as current workflow
	// transactionally change the current workflow with conflict resolution
}

func (r nDCTransactionMgrImpl) updateWorkflowCurrentBranchWithRebuildAsZombie() {
	// append history to branch
	// rebuild the workflow as zombie
}

func (r nDCTransactionMgrImpl) updateWorkflowCurrentBranchWithRebuildWithConflict() {
	// append history to branch
	// rebuild the workflow as current workflow
	// transactionally change the current workflow with conflict resolution
}
