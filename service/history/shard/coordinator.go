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

//go:generate mockgen -copyright_file ../../../LICENSE -package $GOPACKAGE -source $GOFILE -destination coordinator_mock.go -self_package github.com/uber/cadence/service/history/shard

package shard

import (
	"sync/atomic"
	"time"

	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/service/history/config"
)

const (
	sendChanBufferSize    = 800
	receiveChanBufferSize = 1000
)

type (
	// Coordinator manages the failover markers on sending and receiving
	Coordinator interface {
		common.Daemon

		HeartbeatFailoverMarkers(shardID int, markers []*replicator.FailoverMarkerAttributes) <-chan error
		ReceiveFailoverMarkers(shardID int, markers []*replicator.FailoverMarkerAttributes)
	}

	coordinatorImpl struct {
		status        int32
		recorder      map[string]*failoverRecord
		heartbeatChan chan *request
		receiveChan   chan *request
		shutdownChan  chan struct{}

		metadataMgr   persistence.MetadataManager
		historyClient history.Client
		config        *config.Config
		metrics       metrics.Client
		logger        log.Logger
	}

	request struct {
		shardID int
		markers []*replicator.FailoverMarkerAttributes
		respCh  chan error
	}

	failoverRecord struct {
		failoverVersion int64
		shards          map[int]struct{}
	}
)

// NewCoordinator initialize a failover coordinator
func NewCoordinator(
	metadataMgr persistence.MetadataManager,
	historyClient history.Client,
	config *config.Config,
	metrics metrics.Client,
	logger log.Logger,
) Coordinator {
	return &coordinatorImpl{
		status:        common.DaemonStatusInitialized,
		recorder:      make(map[string]*failoverRecord),
		heartbeatChan: make(chan *request, sendChanBufferSize),
		receiveChan:   make(chan *request, receiveChanBufferSize),
		shutdownChan:  make(chan struct{}),
		metadataMgr:   metadataMgr,
		historyClient: historyClient,
		config:        config,
		metrics:       metrics,
		logger:        logger,
	}
}

func (c *coordinatorImpl) Start() {

	if !atomic.CompareAndSwapInt32(&c.status, common.DaemonStatusInitialized, common.DaemonStatusStarted) {
		return
	}

	go c.receiveFailoverMarkersLoop()
	go c.heartbeatFailoverMarkerLoop()

	c.logger.Info("Failover coordinator started.")
}

func (c *coordinatorImpl) Stop() {

	if !atomic.CompareAndSwapInt32(&c.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}

	close(c.shutdownChan)
	c.logger.Info("Failover coordinator stopped.")
}

func (c *coordinatorImpl) HeartbeatFailoverMarkers(
	shardID int,
	markers []*replicator.FailoverMarkerAttributes,
) <-chan error {

	respCh := make(chan error, 1)
	c.heartbeatChan <- &request{
		shardID: shardID,
		markers: markers,
		respCh:  respCh,
	}

	return respCh
}

func (c *coordinatorImpl) ReceiveFailoverMarkers(
	shardID int,
	markers []*replicator.FailoverMarkerAttributes,
) {

	c.receiveChan <- &request{
		shardID: shardID,
		markers: markers,
	}
}

func (c *coordinatorImpl) receiveFailoverMarkersLoop() {

	for {
		select {
		case <-c.shutdownChan:
			return
		case request := <-c.receiveChan:
			c.handleFailoverMarkers(request)
		}
	}
}

func (c *coordinatorImpl) heartbeatFailoverMarkerLoop() {

	timer := time.NewTimer(backoff.JitDuration(
		c.config.FailoverMarkerHeartbeatInterval(),
		c.config.FailoverMarkerHeartbeatTimerJitterCoefficient(),
	))
	defer timer.Stop()
	requestByShard := make(map[int]*request)

	for {
		select {
		case <-c.shutdownChan:
			return
		case request := <-c.heartbeatChan:
			// Here we only add the request to map. We will wait until timer fires to send the request to remote.
			if req, ok := requestByShard[request.shardID]; ok && req != request {
				// during shard movement, duplicated requests can appear
				// if shard moved from this host, to this host.
				c.logger.Warn("Failover marker heartbeat request already exist for shard.")
				if req.respCh != nil {
					close(req.respCh)
				}
			}
			requestByShard[request.shardID] = request
		case <-timer.C:
			if len(requestByShard) > 0 {
				var err error
				//TODO: Remote calls to send failover markers
				for shardID, request := range requestByShard {
					request.respCh <- err
					close(request.respCh)
					delete(requestByShard, shardID)
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
	request *request,
) {
	for _, marker := range request.markers {
		domainID := marker.GetDomainID()

		if record, ok := c.recorder[domainID]; ok {
			// if the local failover version is smaller than the new received marker,
			// it means there is another failover happened and the local one should be invalid.
			if record.failoverVersion < marker.GetFailoverVersion() {
				delete(c.recorder, domainID)
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
		record.shards[request.shardID] = struct{}{}
		if len(record.shards) == c.config.NumberOfShards {
			// TODO: update domain from pend_active to active
			delete(c.recorder, domainID)
		}
	}
}
