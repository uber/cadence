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

//go:generate mockgen -copyright_file ../../../LICENSE -package $GOPACKAGE -source $GOFILE -destination coordinator_mock.go -self_package github.com/uber/cadence/service/history/failover

package failover

import (
	"sync/atomic"
	"time"

	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/domain"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/service/history/config"
)

const (
	notificationChanBufferSize       = 800
	receiveChanBufferSize            = 400
	cleanupMarkerInterval            = 30 * time.Minute
	invalidMarkerDuration            = 1 * time.Hour
	updateDomainRetryInitialInterval = 50 * time.Millisecond
	updateDomainRetryCoefficient     = 2.0
	updateDomainMaxRetry             = 3
)

type (
	// Coordinator manages the failover markers on sending and receiving
	Coordinator interface {
		common.Daemon

		NotifyFailoverMarkers(shardID int, markers []*replicator.FailoverMarkerAttributes) <-chan error
		ReceiveFailoverMarkers(shardIDs []int, marker *replicator.FailoverMarkerAttributes)
	}

	coordinatorImpl struct {
		status           int32
		recorder         map[string]*failoverRecord
		notificationChan chan *notificationRequest
		receiveChan      chan *receiveRequest
		shutdownChan     chan struct{}
		retryPolicy      backoff.RetryPolicy

		metadataMgr   persistence.MetadataManager
		historyClient history.Client
		config        *config.Config
		timeSource    clock.TimeSource
		metrics       metrics.Client
		logger        log.Logger
	}

	notificationRequest struct {
		shardID int
		markers []*replicator.FailoverMarkerAttributes
		respCh  chan error
	}

	receiveRequest struct {
		shardIDs []int
		marker   *replicator.FailoverMarkerAttributes
		respCh   chan error
	}

	failoverRecord struct {
		failoverVersion int64
		shards          map[int]struct{}
		lastUpdatedTime time.Time
	}
)

// NewCoordinator initialize a failover coordinator
func NewCoordinator(
	metadataMgr persistence.MetadataManager,
	historyClient history.Client,
	timeSource clock.TimeSource,
	config *config.Config,
	metrics metrics.Client,
	logger log.Logger,
) Coordinator {

	retryPolicy := &backoff.ExponentialRetryPolicy{}
	retryPolicy.SetInitialInterval(updateDomainRetryInitialInterval)
	retryPolicy.SetBackoffCoefficient(updateDomainRetryCoefficient)
	retryPolicy.SetMaximumAttempts(updateDomainMaxRetry)

	return &coordinatorImpl{
		status:           common.DaemonStatusInitialized,
		recorder:         make(map[string]*failoverRecord),
		notificationChan: make(chan *notificationRequest, notificationChanBufferSize),
		receiveChan:      make(chan *receiveRequest, receiveChanBufferSize),
		shutdownChan:     make(chan struct{}),
		retryPolicy:      retryPolicy,
		metadataMgr:      metadataMgr,
		historyClient:    historyClient,
		timeSource:       timeSource,
		config:           config,
		metrics:          metrics,
		logger:           logger,
	}
}

func (c *coordinatorImpl) Start() {

	if !atomic.CompareAndSwapInt32(&c.status, common.DaemonStatusInitialized, common.DaemonStatusStarted) {
		return
	}

	go c.receiveFailoverMarkersLoop()
	go c.notifyFailoverMarkerLoop()

	c.logger.Info("Failover coordinator started.")
}

func (c *coordinatorImpl) Stop() {

	if !atomic.CompareAndSwapInt32(&c.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}

	close(c.shutdownChan)
	c.logger.Info("Failover coordinator stopped.")
}

func (c *coordinatorImpl) NotifyFailoverMarkers(
	shardID int,
	markers []*replicator.FailoverMarkerAttributes,
) <-chan error {

	respCh := make(chan error, 1)
	c.notificationChan <- &notificationRequest{
		shardID: shardID,
		markers: markers,
		respCh:  respCh,
	}

	return respCh
}

func (c *coordinatorImpl) ReceiveFailoverMarkers(
	shardIDs []int,
	marker *replicator.FailoverMarkerAttributes,
) {

	c.receiveChan <- &receiveRequest{
		shardIDs: shardIDs,
		marker:   marker,
	}
}

func (c *coordinatorImpl) receiveFailoverMarkersLoop() {

	ticker := time.NewTicker(cleanupMarkerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.shutdownChan:
			return
		case <-ticker.C:
			c.cleanupInvalidMarkers()
		case request := <-c.receiveChan:
			c.handleFailoverMarkers(request)
		}
	}
}

func (c *coordinatorImpl) notifyFailoverMarkerLoop() {

	timer := time.NewTimer(backoff.JitDuration(
		c.config.FailoverMarkerHeartbeatInterval(),
		c.config.FailoverMarkerHeartbeatTimerJitterCoefficient(),
	))
	defer timer.Stop()
	requestByMarker := make(map[*replicator.FailoverMarkerAttributes]*receiveRequest)

	for {
		select {
		case <-c.shutdownChan:
			return
		case notificationReq := <-c.notificationChan:
			// if there is a shard movement happen, it is fine to have duplicate shard ID in the request
			// The receiver side will de-dup the shard IDs. See: handleFailoverMarkers
			for _, marker := range notificationReq.markers {
				if _, ok := requestByMarker[marker]; !ok {
					requestByMarker[marker] = &receiveRequest{
						shardIDs: []int{},
						marker:   marker,
						respCh:   notificationReq.respCh,
					}
				}
				req := requestByMarker[marker]
				req.shardIDs = append(req.shardIDs, notificationReq.shardID)
			}
		case <-timer.C:
			if len(requestByMarker) > 0 {
				var err error
				//TODO: Remote calls to send failover markers
				for marker, request := range requestByMarker {
					request.respCh <- err
					close(request.respCh)
					delete(requestByMarker, marker)
				}
			}

			timer.Reset(backoff.JitDuration(
				c.config.FailoverMarkerHeartbeatInterval(),
				c.config.FailoverMarkerHeartbeatTimerJitterCoefficient(),
			))
		}
	}
}

func (c *coordinatorImpl) handleFailoverMarkers(
	request *receiveRequest,
) {

	marker := request.marker
	domainID := marker.GetDomainID()

	if record, ok := c.recorder[domainID]; ok {
		// if the local failover version is smaller than the new received marker,
		// it means there is another failover happened and the local one should be invalid.
		if record.failoverVersion < marker.GetFailoverVersion() {
			delete(c.recorder, domainID)
		}

		// if the local failover version is larger than the new received marker,
		// ignore the incoming marker
		if record.failoverVersion > marker.GetFailoverVersion() {
			return
		}
	}

	if _, ok := c.recorder[domainID]; !ok {
		// initialize the failover record
		c.recorder[marker.GetDomainID()] = &failoverRecord{
			failoverVersion: marker.GetFailoverVersion(),
			shards:          make(map[int]struct{}),
		}
	}

	record := c.recorder[domainID]
	record.lastUpdatedTime = c.timeSource.Now()
	for shardID := range request.shardIDs {
		record.shards[shardID] = struct{}{}
	}

	if len(record.shards) == c.config.NumberOfShards {
		if err := domain.CleanPendingActiveState(
			c.metadataMgr,
			domainID,
			record.failoverVersion,
			c.retryPolicy,
		); err != nil {
			c.logger.Error("Coordinator failed to update domain after receiving all failover markers",
				tag.WorkflowDomainID(domainID))
			c.metrics.IncCounter(metrics.DomainFailoverScope, metrics.CadenceFailures)
			return
		}
		delete(c.recorder, domainID)
	}
}

func (c *coordinatorImpl) cleanupInvalidMarkers() {
	for domainID, record := range c.recorder {
		if c.timeSource.Now().Sub(record.lastUpdatedTime) > invalidMarkerDuration {
			delete(c.recorder, domainID)
		}
	}
}
