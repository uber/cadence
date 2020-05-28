// Copyright (c) 2017-2020 Uber Technologies, Inc.
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

//go:generate mockgen -copyright_file ../../LICENSE -package $GOPACKAGE -source $GOFILE -destination failover_watcher_mock.go

package domain

import (
	"sync/atomic"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service/dynamicconfig"
)

type (
	// FailoverWatcher handles failover operation on domain entities
	FailoverWatcher interface {
		common.Daemon
	}

	failoverWatcherImpl struct {
		status          int32
		shutdownChan    chan struct{}
		refreshInterval dynamicconfig.DurationPropertyFn
		refreshJitter   dynamicconfig.FloatPropertyFn

		metadataMgr persistence.MetadataManager
		domainCache cache.DomainCache
		timeSource  clock.TimeSource
		metrics     metrics.Client
		logger      log.Logger
	}
)

var _ FailoverWatcher = (*failoverWatcherImpl)(nil)

// NewFailoverWatcher initializes domain failover processor
func NewFailoverWatcher(
	domainCache cache.DomainCache,
	metadataMgr persistence.MetadataManager,
	timeSource clock.TimeSource,
	refreshInterval dynamicconfig.DurationPropertyFn,
	refreshJitter dynamicconfig.FloatPropertyFn,
	metrics metrics.Client,
	logger log.Logger,
) FailoverWatcher {

	return &failoverWatcherImpl{
		status:          common.DaemonStatusInitialized,
		shutdownChan:    make(chan struct{}),
		refreshInterval: refreshInterval,
		refreshJitter:   refreshJitter,
		domainCache:     domainCache,
		metadataMgr:     metadataMgr,
		timeSource:      timeSource,
		metrics:         metrics,
		logger:          logger,
	}
}

func (p *failoverWatcherImpl) Start() {
	if !atomic.CompareAndSwapInt32(&p.status, common.DaemonStatusInitialized, common.DaemonStatusStarted) {
		return
	}

	go p.refreshDomainLoop()

	p.logger.Info("Domain failover processor started.")
}

func (p *failoverWatcherImpl) Stop() {
	if !atomic.CompareAndSwapInt32(&p.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}

	close(p.shutdownChan)
	p.logger.Info("Domain failover processor stop.")
}

func (p *failoverWatcherImpl) refreshDomainLoop() {

	timer := time.NewTimer(backoff.JitDuration(
		p.refreshInterval(),
		p.refreshJitter(),
	))

	for {
		select {
		case <-p.shutdownChan:
			timer.Stop()
			return
		case <-timer.C:
			domains := p.domainCache.GetAllDomain()
			for _, domain := range domains {
				p.handleFailoverTimeout(domain)
				select {
				case <-p.shutdownChan:
					timer.Stop()
					return
				default:
					p.logger.Debug("Stop refresh domain as the processing is stopping.")
				}
			}

			timer.Reset(backoff.JitDuration(
				p.refreshInterval(),
				p.refreshJitter(),
			))
		}
	}
}

func (p *failoverWatcherImpl) handleFailoverTimeout(
	domain *cache.DomainCacheEntry,
) {

	failoverEndTime := domain.GetDomainFailoverEndTime()
	if failoverEndTime != nil && p.timeSource.Now().After(time.Unix(0, *failoverEndTime)) {
		domainName := domain.GetInfo().Name
		// force failover the domain without setting the failover timeout
		if err := p.cleanPendingActiveState(
			domainName,
			domain.GetFailoverVersion(),
		); err != nil {
			p.metrics.IncCounter(metrics.DomainFailoverScope, metrics.CadenceFailures)
			p.logger.Error("Failed to update pending-active domain to active.", tag.WorkflowDomainID(domainName), tag.Error(err))
		}
	}
}

func (p *failoverWatcherImpl) cleanPendingActiveState(
	domainName string,
	failoverVersion int64,
) error {

	// must get the metadata (notificationVersion) first
	// this version can be regarded as the lock on the v2 domain table
	// and since we do not know which table will return the domain afterwards
	// this call has to be made
	metadata, err := p.metadataMgr.GetMetadata()
	if err != nil {
		return err
	}
	notificationVersion := metadata.NotificationVersion
	getResponse, err := p.metadataMgr.GetDomain(&persistence.GetDomainRequest{Name: domainName})
	if err != nil {
		return err
	}

	info := getResponse.Info
	config := getResponse.Config
	replicationConfig := getResponse.ReplicationConfig
	configVersion := getResponse.ConfigVersion
	localFailoverVersion := getResponse.FailoverVersion
	failoverNotificationVersion := getResponse.FailoverNotificationVersion
	isGlobalDomain := getResponse.IsGlobalDomain
	gracefulFailoverEndTime := getResponse.FailoverEndTime

	if isGlobalDomain && gracefulFailoverEndTime != nil && failoverVersion == localFailoverVersion {
		// if the domain is still pending active and the failover versions are the same, clean the state
		updateReq := &persistence.UpdateDomainRequest{
			Info:                        info,
			Config:                      config,
			ReplicationConfig:           replicationConfig,
			ConfigVersion:               configVersion,
			FailoverVersion:             failoverVersion,
			FailoverNotificationVersion: failoverNotificationVersion,
			FailoverEndTime:             nil,
			NotificationVersion:         notificationVersion,
		}
		err = p.metadataMgr.UpdateDomain(updateReq)
		if err != nil {
			return err
		}
	}
	return nil
}
