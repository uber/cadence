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
	"sync"
	"time"

	"github.com/uber-common/bark"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
)

type (
	transferTaskPredicate   func(string) (bool, error)
	maxReadLevelProvider    func() int64
	transferQueueAckMgrImpl struct {
		isFailover    bool
		clusterName   string
		shard         ShardContext
		executionMgr  persistence.ExecutionManager
		logger        bark.Logger
		metricsClient metrics.Client
		lastUpdated   time.Time
		config        *Config
		// task predicate for filtering
		transferTaskPredicate transferTaskPredicate
		// max read level providers the max read level
		maxReadLevel maxReadLevelProvider
		// isReadFinished indicate timer queue ack manager
		// have no more task to send out
		isReadFinished bool
		// finishedChan will send out signal when timer
		// queue ack manager have no more task to send out and all
		// tasks sent are finished
		finishedChan chan struct{}

		sync.Mutex
		// outstanding timer tasks, which cannot be processed right now
		retryTasks []*persistence.TransferTaskInfo
		// timer task read level
		readLevel int64
		// base ack manager
		*queueAckMgrBase
	}
)

func newTransferQueueAckMgr(shard ShardContext, metricsClient metrics.Client, clusterName string, logger bark.Logger) *transferQueueAckMgrImpl {
	transferTaskPredicate := func(transferDomainID string) (bool, error) {
		domainEntry, err := shard.GetDomainCache().GetDomainByID(transferDomainID)
		if err != nil {
			return false, err
		}
		if !domainEntry.GetIsGlobalDomain() &&
			clusterName != shard.GetService().GetClusterMetadata().GetCurrentClusterName() {
			// timer task does not belong to cluster name
			return false, nil
		} else if domainEntry.GetIsGlobalDomain() &&
			domainEntry.GetReplicationConfig().ActiveClusterName != clusterName {
			// timer task does not belong here
			return false, nil
		}
		return true, nil
	}
	ackLevel := shard.GetTransferAckLevel(clusterName)
	maxReadLevel := func() int64 {
		return shard.GetTransferMaxReadLevel()
	}

	transferQueueAckMgr := &transferQueueAckMgrImpl{
		isFailover:            true,
		clusterName:           clusterName,
		shard:                 shard,
		executionMgr:          shard.GetExecutionManager(),
		metricsClient:         metricsClient,
		logger:                logger,
		lastUpdated:           time.Now(), // this has nothing to do with remote cluster, so use the local time
		config:                shard.GetConfig(),
		retryTasks:            []*persistence.TransferTaskInfo{},
		readLevel:             ackLevel,
		maxReadLevel:          maxReadLevel,
		transferTaskPredicate: transferTaskPredicate,
		isReadFinished:        false,
		finishedChan:          make(chan struct{}, 1),
	}

	return transferQueueAckMgr
}

func newTransferQueueFailoverAckMgr(shard ShardContext, metricsClient metrics.Client, domainID string, standbyClusterName string, logger bark.Logger) *transferQueueAckMgrImpl {
	transferTaskPredicate := func(transferDomainID string) (bool, error) {
		return transferDomainID == domainID, nil
	}
	// failover ack manager will start from the standby cluster's ack level to active cluster's ack level
	ackLevel := shard.GetTransferAckLevel(standbyClusterName)
	maxAckLevel := shard.GetTransferAckLevel(shard.GetService().GetClusterMetadata().GetCurrentClusterName())
	maxReadLevel := func() int64 {
		return maxAckLevel // this should be a fix number, not a function call
	}

	transferQueueAckMgr := &transferQueueAckMgrImpl{
		isFailover:            true,
		clusterName:           standbyClusterName,
		shard:                 shard,
		executionMgr:          shard.GetExecutionManager(),
		metricsClient:         metricsClient,
		logger:                logger,
		lastUpdated:           time.Now(), // this has nothing to do with remote cluster, so use the local time
		config:                shard.GetConfig(),
		retryTasks:            []*persistence.TransferTaskInfo{},
		readLevel:             ackLevel,
		maxReadLevel:          maxReadLevel,
		transferTaskPredicate: transferTaskPredicate,
		isReadFinished:        false,
		finishedChan:          make(chan struct{}, 1),
	}

	return transferQueueAckMgr
}

func (t *transferQueueAckMgrImpl) getFinishedChan() <-chan struct{} {
	return t.finishedChan
}

func (t *transferQueueAckMgrImpl) readTransferTasks() ([]*persistence.TransferTaskInfo, bool, error) {
	t.Lock()
	readLevel := t.readLevel
	transferTaskRetrySize := len(t.retryTasks)
	t.Unlock()

	var tasks []*persistence.TransferTaskInfo
	morePage := transferTaskRetrySize > 0
	var err error
	maxReadLevel := t.maxReadLevel()
	batchSize := t.config.TransferTaskBatchSize

	if transferTaskRetrySize < batchSize && readLevel < maxReadLevel {
		var token []byte
		tasks, token, err = t.getTransferTasks(readLevel, maxReadLevel, batchSize)
		if err != nil {
			return nil, false, err
		}
		t.logger.Debugf("readTransferTasks: read level: %v, count: %v, next token: %v", readLevel, len(tasks), token)
		morePage = len(token) > 0
	}

	t.Lock()
	defer t.Unlock()
	if t.isFailover && !morePage {
		t.isReadFinished = true
	}

	filteredTasks := t.retryTasks
	t.retryTasks = []*persistence.TransferTaskInfo{}

TaskFilterLoop:
	for _, task := range tasks {
		if t.readLevel >= task.GetTaskID() {
			t.logger.Fatalf("Next task ID is less than current read level.  TaskID: %v, ReadLevel: %v",
				task.GetTaskID(), t.readLevel)
		}

		ok, err := t.transferTaskPredicate(task.DomainID)
		if err != nil {
			return nil, false, err
		}

		t.readLevel = task.GetTaskID()
		if !ok {
			continue TaskFilterLoop
		}

		t.createTask(task)
		filteredTasks = append(filteredTasks, task)
	}

	return tasks, morePage, nil
}

func (t *transferQueueAckMgrImpl) retryTimerTask(timerTask *persistence.TransferTaskInfo) {
	t.Lock()
	defer t.Unlock()

	t.retryTasks = append(t.retryTasks, timerTask)
}

func (t *transferQueueAckMgrImpl) readRetryTimerTasks() []*persistence.TransferTaskInfo {
	t.Lock()
	defer t.Unlock()

	retryTasks := t.retryTasks
	t.retryTasks = []*persistence.TransferTaskInfo{}
	return retryTasks
}

func (t *transferQueueAckMgrImpl) updateAckLevel() {
	t.Lock()
	defer t.Unlock()

	t.ackLevel = t.queueAckMgrBase.updateAckLevel(t.readLevel)
	t.updateTransferAckLevel(t.ackLevel)
}

func (t *transferQueueAckMgrImpl) getTransferTasks(minReadLevel int64, maxReadLevel int64, batchSize int) ([]*persistence.TransferTaskInfo, []byte, error) {
	response, err := t.executionMgr.GetTransferTasks(&persistence.GetTransferTasksRequest{
		ReadLevel:    minReadLevel,
		MaxReadLevel: maxReadLevel,
		BatchSize:    batchSize,
	})

	if err != nil {
		return nil, nil, err
	}

	return response.Tasks, response.NextPageToken, nil
}

func (t *transferQueueAckMgrImpl) updateTransferAckLevel(ackLevel int64) {
	t.shard.UpdateTransferAckLevel(t.clusterName, ackLevel)
}
