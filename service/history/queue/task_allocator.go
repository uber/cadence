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

package queue

import (
	"errors"
	"sync"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/shard"
	htask "github.com/uber/cadence/service/history/task"
)

type (
	// TaskAllocator verifies if a task should be processed or not
	TaskAllocator interface {
		VerifyActiveTask(taskDomainID string, task interface{}) (bool, error)
		VerifyFailoverActiveTask(targetDomainIDs map[string]struct{}, taskDomainID string, task interface{}) (bool, error)
		VerifyStandbyTask(standbyCluster string, taskDomainID string, task interface{}) (bool, error)
		Lock()
		Unlock()
	}

	taskAllocatorImpl struct {
		currentClusterName string
		shard              shard.Context
		domainCache        cache.DomainCache
		logger             log.Logger

		locker sync.RWMutex
	}
)

// NewTaskAllocator create a new task allocator
func NewTaskAllocator(shard shard.Context) TaskAllocator {
	return &taskAllocatorImpl{
		currentClusterName: shard.GetService().GetClusterMetadata().GetCurrentClusterName(),
		shard:              shard,
		domainCache:        shard.GetDomainCache(),
		logger:             shard.GetLogger(),
	}
}

// VerifyActiveTask, will return true if task activeness check is successful
func (t *taskAllocatorImpl) VerifyActiveTask(taskDomainID string, task interface{}) (bool, error) {
	t.locker.RLock()
	defer t.locker.RUnlock()

	domainEntry, err := t.domainCache.GetDomainByID(taskDomainID)
	if err != nil {
		// it is possible that the domain is deleted
		// we should treat that domain as active
		if _, ok := err.(*types.EntityNotExistsError); !ok {
			t.logger.Warn("Cannot find domain", tag.WorkflowDomainID(taskDomainID))
			return false, err
		}
		t.logger.Warn("Cannot find domain, default to process task.", tag.WorkflowDomainID(taskDomainID), tag.Value(task))
		return true, nil
	}
	if domainEntry.IsGlobalDomain() && t.currentClusterName != domainEntry.GetReplicationConfig().ActiveClusterName {
		// timer task does not belong to cluster name
		t.logger.Debug("Domain is not active, skip task.", tag.WorkflowDomainID(taskDomainID), tag.Value(task))
		return false, nil
	}

	if err := t.checkDomainPendingActive(
		domainEntry,
		taskDomainID,
		task,
	); err != nil {
		return false, err
	}

	t.logger.Debug("Domain is active, process task.", tag.WorkflowDomainID(taskDomainID), tag.Value(task))
	return true, nil
}

// VerifyFailoverActiveTask, will return true if task activeness check is successful
func (t *taskAllocatorImpl) VerifyFailoverActiveTask(targetDomainIDs map[string]struct{}, taskDomainID string, task interface{}) (bool, error) {
	_, ok := targetDomainIDs[taskDomainID]
	if ok {
		t.locker.RLock()
		defer t.locker.RUnlock()

		domainEntry, err := t.domainCache.GetDomainByID(taskDomainID)
		if err != nil {
			// it is possible that the domain is deleted
			// we should treat that domain as not active
			if _, ok := err.(*types.EntityNotExistsError); !ok {
				t.logger.Warn("Cannot find domain", tag.WorkflowDomainID(taskDomainID))
				return false, err
			}
			t.logger.Warn("Cannot find domain, default to not process task.", tag.WorkflowDomainID(taskDomainID), tag.Value(task))
			return false, nil
		}
		if err := t.checkDomainPendingActive(
			domainEntry,
			taskDomainID,
			task,
		); err != nil {
			return false, err
		}

		t.logger.Debug("Failover Domain is active, process task.", tag.WorkflowDomainID(taskDomainID), tag.Value(task))
		return true, nil
	}
	t.logger.Debug("Failover Domain is not active, skip task.", tag.WorkflowDomainID(taskDomainID), tag.Value(task))
	return false, nil
}

// VerifyStandbyTask, will return true if task standbyness check is successful
func (t *taskAllocatorImpl) VerifyStandbyTask(standbyCluster string, taskDomainID string, task interface{}) (bool, error) {
	t.locker.RLock()
	defer t.locker.RUnlock()

	domainEntry, err := t.domainCache.GetDomainByID(taskDomainID)
	if err != nil {
		// it is possible that the domain is deleted
		// we should treat that domain as not active
		if _, ok := err.(*types.EntityNotExistsError); !ok {
			t.logger.Warn("Cannot find domain", tag.WorkflowDomainID(taskDomainID))
			return false, err
		}
		t.logger.Warn("Cannot find domain, default to not process task.", tag.WorkflowDomainID(taskDomainID), tag.Value(task))
		return false, nil
	}
	if !domainEntry.IsGlobalDomain() {
		// non global domain, timer task does not belong here
		t.logger.Debug("Domain is not global, skip task.", tag.WorkflowDomainID(taskDomainID), tag.Value(task))
		return false, nil
	} else if domainEntry.IsGlobalDomain() && domainEntry.GetReplicationConfig().ActiveClusterName != standbyCluster {
		// timer task does not belong here
		t.logger.Debug("Domain is not standby, skip task.", tag.WorkflowDomainID(taskDomainID), tag.Value(task))
		return false, nil
	}

	if err := t.checkDomainPendingActive(
		domainEntry,
		taskDomainID,
		task,
	); err != nil {
		return false, err
	}

	t.logger.Debug("Domain is standby, process task.", tag.WorkflowDomainID(taskDomainID), tag.Value(task))
	return true, nil
}

func (t *taskAllocatorImpl) checkDomainPendingActive(domainEntry *cache.DomainCacheEntry, taskDomainID string, task interface{}) error {

	if domainEntry.IsGlobalDomain() && domainEntry.GetFailoverEndTime() != nil {
		// the domain is pending active, pause on processing this task
		t.logger.Debug("Domain is not in pending active, skip task.", tag.WorkflowDomainID(taskDomainID), tag.Value(task))
		return htask.ErrTaskPendingActive
	}
	return nil
}

// Lock block all task allocation
func (t *taskAllocatorImpl) Lock() {
	t.locker.Lock()
}

// Unlock resume the task allocator
func (t *taskAllocatorImpl) Unlock() {
	t.locker.Unlock()
}

// isDomainNotRegistered checks either if domain does not exist or is in deprecated or deleted status
func isDomainNotRegistered(shard shard.Context, domainID string) (bool, error) {
	domainEntry, err := shard.GetDomainCache().GetDomainByID(domainID)
	if err != nil {
		// error in finding a domain
		return false, err
	}
	info := domainEntry.GetInfo()
	if info == nil {
		return false, errors.New("domain info is nil in cache")
	}
	return info.Status == persistence.DomainStatusDeprecated || info.Status == persistence.DomainStatusDeleted, nil
}
