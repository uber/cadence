// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

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

package replication

import (
	ctx "context"
	"fmt"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
)

const (
	metricsEmissionInterval = time.Minute
)

type (
	// MetricsEmitter is responsible for emitting replication metrics occasionally.
	// TODO: Should this be merged into task_processor or is this a good split?
	MetricsEmitter interface {
		common.Daemon
	}

	MetricsEmitterShardData interface {
		GetShardID() int
		GetLogger() log.Logger
		GetClusterMetadata() cluster.Metadata
		GetTransferMaxReadLevel() int64
		GetClusterReplicationLevel(cluster string) int64
		GetTimeSource() clock.TimeSource
	}

	metricsEmitterImpl struct {
		currentCluster     string
		remoteClusterNames []string
		shardData          MetricsEmitterShardData
		reader             taskReader
		scope              metrics.Scope
		logger             log.Logger
		status             int32
		done               chan struct{}
	}
)

var _ MetricsEmitter = (*metricsEmitterImpl)(nil)

// NewMetricsEmitter creates a new metrics emitter, which starts a goroutine to emit replication metrics occasionally.
func NewMetricsEmitter(
	shardData MetricsEmitterShardData,
	reader taskReader,
	metricsClient metrics.Client,
) *metricsEmitterImpl {
	shardID := shardData.GetShardID()
	currentCluster := shardData.GetClusterMetadata().GetCurrentClusterName()

	remoteClusters := shardData.GetClusterMetadata().GetRemoteClusterInfo()
	remoteClusterNames := make([]string, len(remoteClusters))
	for remoteCluster := range remoteClusters {
		remoteClusterNames = append(remoteClusterNames, remoteCluster)
	}

	scope := metricsClient.Scope(
		metrics.ReplicationMetricEmitterScope,
		metrics.ActiveClusterTag(currentCluster),
		metrics.InstanceTag(strconv.Itoa(shardID)),
	)
	logger := shardData.GetLogger().WithTags(
		tag.ClusterName(currentCluster),
		tag.ShardID(shardID))

	return &metricsEmitterImpl{
		currentCluster:     currentCluster,
		remoteClusterNames: remoteClusterNames,
		status:             common.DaemonStatusInitialized,
		shardData:          shardData,
		reader:             reader,
		scope:              scope,
		logger:             logger,
		done:               make(chan struct{}),
	}
}

func (m *metricsEmitterImpl) Start() {
	if !atomic.CompareAndSwapInt32(&m.status, common.DaemonStatusInitialized, common.DaemonStatusStarted) {
		return
	}

	go m.emitMetricsLoop()
	m.logger.Info("ReplicationMetricsEmitter started.")
}

func (m *metricsEmitterImpl) Stop() {
	if !atomic.CompareAndSwapInt32(&m.status, common.DaemonStatusStarted, common.DaemonStatusStopped) {
		return
	}

	m.logger.Info("ReplicationMetricsEmitter shutting down.")
	close(m.done)
}

func (m *metricsEmitterImpl) emitMetricsLoop() {
	ticker := time.NewTicker(metricsEmissionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.done:
			return
		case <-ticker.C:
			m.emitMetrics()
		}
	}
}

func (m *metricsEmitterImpl) emitMetrics() {
	for _, remoteClusterName := range m.remoteClusterNames {
		logger := m.logger.WithTags(tag.RemoteCluster(remoteClusterName))
		scope := m.scope.Tagged(metrics.TargetClusterTag(remoteClusterName))

		replicationLatency, err := m.determineReplicationLatency(remoteClusterName)
		if err != nil {
			return
		}

		scope.UpdateGauge(metrics.ReplicationLatency, float64(replicationLatency.Nanoseconds()))
		logger.Debug(fmt.Sprintf("ReplicationLatency metric emitted: %v", float64(replicationLatency.Nanoseconds())))
	}
}

func (m *metricsEmitterImpl) determineReplicationLatency(remoteClusterName string) (time.Duration, error) {
	logger := m.logger.WithTags(tag.RemoteCluster(remoteClusterName))
	lastReadTaskID := m.shardData.GetClusterReplicationLevel(remoteClusterName)
	maxReadLevel := m.shardData.GetTransferMaxReadLevel()

	// TODO I should probably use task_store instead to get from cache, even though I don't need it hydrated
	// TODO Do I need to use a different context? Or add a deadline / cancellation?
	tasks, _, err := m.reader.Read(ctx.Background(), lastReadTaskID, maxReadLevel)
	if err != nil {
		logger.Error("Error reading", tag.Error(err))
		return 0, err
	}
	logger.Debug("Number of tasks retrieved", tag.Number(int64(len(tasks))))

	var replicationLatency time.Duration
	if len(tasks) > 0 {
		creationTime := time.Unix(0, tasks[0].CreationTime)
		replicationLatency = m.shardData.GetTimeSource().Now().Sub(creationTime)
	}

	return replicationLatency, nil
}
