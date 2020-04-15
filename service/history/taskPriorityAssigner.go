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

//go:generate mockgen -copyright_file ../../LICENSE -package $GOPACKAGE -source $GOFILE -destination taskPriorityAssigner_mock.go -self_package github.com/uber/cadence/service/history

package history

import (
	"sync"

	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/quotas"
	t "github.com/uber/cadence/common/task"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/task"
)

type (
	taskPriorityAssigner interface {
		Assign(task.Task) error
	}

	taskPriorityAssignerImpl struct {
		sync.RWMutex

		currentClusterName string
		domainCache        cache.DomainCache
		config             *config.Config
		logger             log.Logger
		scope              metrics.Scope
		rateLimiters       map[string]quotas.Limiter
	}
)

var _ taskPriorityAssigner = (*taskPriorityAssignerImpl)(nil)

func newTaskPriorityAssigner(
	currentClusterName string,
	domainCache cache.DomainCache,
	logger log.Logger,
	metricClient metrics.Client,
	config *config.Config,
) *taskPriorityAssignerImpl {
	return &taskPriorityAssignerImpl{
		currentClusterName: currentClusterName,
		domainCache:        domainCache,
		config:             config,
		logger:             logger,
		scope:              metricClient.Scope(metrics.TaskPriorityAssignerScope),
		rateLimiters:       make(map[string]quotas.Limiter),
	}
}

func (a *taskPriorityAssignerImpl) Assign(
	queueTask task.Task,
) error {
	if queueTask.GetQueueType() == task.QueueTypeReplication {
		queueTask.SetPriority(t.GetTaskPriority(t.LowPriorityClass, t.DefaultPrioritySubclass))
		return nil
	}

	// timer of transfer task, first check if domain is active or not
	domainName, active, err := a.getDomainInfo(queueTask.GetDomainID())
	if err != nil {
		return err
	}

	if !active {
		queueTask.SetPriority(t.GetTaskPriority(t.LowPriorityClass, t.DefaultPrioritySubclass))
		return nil
	}

	if !a.getRateLimiter(domainName).Allow() {
		queueTask.SetPriority(t.GetTaskPriority(t.DefaultPriorityClass, t.DefaultPrioritySubclass))
		taggedScope := a.scope.Tagged(metrics.DomainTag(domainName))
		if queueTask.GetQueueType() == task.QueueTypeTransfer {
			taggedScope.IncCounter(metrics.TransferTaskThrottledCounter)
		} else {
			taggedScope.IncCounter(metrics.TimerTaskThrottledCounter)
		}
		return nil
	}

	queueTask.SetPriority(t.GetTaskPriority(t.HighPriorityClass, t.DefaultPrioritySubclass))
	return nil
}

// getDomainInfo returns three pieces of information:
//  1. domain name
//  2. if domain is active
//  3. error, if any
func (a *taskPriorityAssignerImpl) getDomainInfo(
	domainID string,
) (string, bool, error) {
	domainEntry, err := a.domainCache.GetDomainByID(domainID)
	if err != nil {
		if _, ok := err.(*workflow.EntityNotExistsError); !ok {
			a.logger.Warn("Cannot find domain", tag.WorkflowDomainID(domainID))
			return "", false, err
		}
		// it is possible that the domain is deleted
		// we should treat that domain as active
		a.logger.Warn("Cannot find domain, treat as active task.", tag.WorkflowDomainID(domainID))
		return "", true, nil
	}

	if domainEntry.IsGlobalDomain() && a.currentClusterName != domainEntry.GetReplicationConfig().ActiveClusterName {
		return domainEntry.GetInfo().Name, false, nil
	}
	return domainEntry.GetInfo().Name, true, nil
}

func (a *taskPriorityAssignerImpl) getRateLimiter(
	domainName string,
) quotas.Limiter {
	a.RLock()
	if limiter, ok := a.rateLimiters[domainName]; ok {
		a.RUnlock()
		return limiter
	}
	a.RUnlock()

	limiter := quotas.NewDynamicRateLimiter(
		func() float64 {
			return float64(a.config.TaskProcessRPS(domainName))
		},
	)

	a.Lock()
	defer a.Unlock()
	if existingLimiter, ok := a.rateLimiters[domainName]; ok {
		return existingLimiter
	}

	a.rateLimiters[domainName] = limiter
	return limiter
}
