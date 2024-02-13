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

package asyncworkflow

import (
	"context"
	"sync"
	"time"

	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/asyncworkflow/queue"
	"github.com/uber/cadence/common/asyncworkflow/queue/provider"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
)

const (
	defaultRefreshInterval = 5 * time.Minute
	defaultShutdownTimeout = 5 * time.Second
)

type ConsumerManagerOptions func(*ConsumerManager)

func WithTimeSource(timeSrc clock.TimeSource) ConsumerManagerOptions {
	return func(c *ConsumerManager) {
		c.timeSrc = timeSrc
	}
}

func NewConsumerManager(
	logger log.Logger,
	metricsClient metrics.Client,
	domainCache cache.DomainCache,
	queueProvider queue.Provider,
	frontendClient frontend.Client,
	options ...ConsumerManagerOptions,
) *ConsumerManager {
	ctx, cancel := context.WithCancel(context.Background())
	cm := &ConsumerManager{
		logger:          logger.WithTags(tag.ComponentAsyncWFConsumptionManager),
		metricsClient:   metricsClient,
		domainCache:     domainCache,
		queueProvider:   queueProvider,
		frontendClient:  frontendClient,
		refreshInterval: defaultRefreshInterval,
		shutdownTimeout: defaultShutdownTimeout,
		ctx:             ctx,
		cancelFn:        cancel,
		activeConsumers: make(map[string]provider.Consumer),
		timeSrc:         clock.NewRealTimeSource(),
	}

	for _, opt := range options {
		opt(cm)
	}
	return cm
}

type ConsumerManager struct {
	logger          log.Logger
	metricsClient   metrics.Client
	timeSrc         clock.TimeSource
	domainCache     cache.DomainCache
	queueProvider   queue.Provider
	frontendClient  frontend.Client
	refreshInterval time.Duration
	shutdownTimeout time.Duration
	ctx             context.Context
	cancelFn        context.CancelFunc
	wg              sync.WaitGroup
	activeConsumers map[string]provider.Consumer
}

func (c *ConsumerManager) Start() {
	c.logger.Info("Starting ConsumerManager")
	c.wg.Add(1)
	go c.run()
}

func (c *ConsumerManager) Stop() {
	c.logger.Info("Stopping ConsumerManager")
	c.cancelFn()
	c.wg.Wait()
	if !common.AwaitWaitGroup(&c.wg, c.shutdownTimeout) {
		c.logger.Warn("ConsumerManager timed out on shutdown", tag.Dynamic("timeout", c.shutdownTimeout))
		return
	}

	for qID, consumer := range c.activeConsumers {
		consumer.Stop()
		c.logger.Info("Stopped consumer", tag.AsyncWFQueueID(qID))
	}

	c.logger.Info("Stopped ConsumerManager")
}

func (c *ConsumerManager) run() {
	defer c.wg.Done()

	timer := c.timeSrc.NewTimer(c.refreshInterval)
	defer timer.Stop()
	c.logger.Info("ConsumerManager background loop started", tag.Dynamic("refresh-interval", c.refreshInterval))

	c.refreshConsumers()

	for {
		select {
		case <-timer.Chan():
			c.refreshConsumers()
		case <-c.ctx.Done():
			c.logger.Info("ConsumerManager background loop stopped because context is done")
			return
		}
	}
}

func (c *ConsumerManager) refreshConsumers() {
	domains := c.domainCache.GetAllDomain()
	c.logger.Info("Refreshing consumers", tag.Dynamic("domain-count", len(domains)), tag.Dynamic("consumer-count", len(c.activeConsumers)))
	refCounts := make(map[string]int, len(c.activeConsumers))

	for _, domain := range domains {
		select {
		default:
		case <-c.ctx.Done():
			c.logger.Info("refreshConsumers is terminating because context is done")
			return
		}

		// domain config is not set or async workflow config is not set
		if domain.GetConfig() == nil || domain.GetConfig().AsyncWorkflowConfig == (types.AsyncWorkflowConfiguration{}) {
			continue
		}

		cfg := domain.GetConfig().AsyncWorkflowConfig
		queue, err := c.getQueue(cfg)
		if err != nil {
			c.logger.Error("Failed to get queue", tag.Error(err), tag.WorkflowDomainName(domain.GetInfo().Name))
			continue
		}

		if !cfg.Enabled {
			// Already running active consumers for such queues will be stopped in the next loop
			continue
		}

		// async workflow config is enabled. check if consumer is already running
		if c.activeConsumers[queue.ID()] != nil {
			c.logger.Debug("Consumer already running", tag.WorkflowDomainName(domain.GetInfo().Name), tag.AsyncWFQueueID(queue.ID()))
			refCounts[queue.ID()]++
			continue
		}

		c.logger.Info("Starting consumer", tag.WorkflowDomainName(domain.GetInfo().Name), tag.AsyncWFQueueID(queue.ID()))
		consumer, err := queue.CreateConsumer(&provider.Params{
			Logger:         c.logger,
			MetricsClient:  c.metricsClient,
			FrontendClient: c.frontendClient,
		})
		if err != nil {
			c.logger.Error("Failed to create consumer", tag.Error(err), tag.WorkflowDomainName(domain.GetInfo().Name), tag.AsyncWFQueueID(queue.ID()))
			continue
		}

		if err := consumer.Start(); err != nil {
			c.logger.Error("Failed to start consumer", tag.Error(err), tag.WorkflowDomainName(domain.GetInfo().Name), tag.AsyncWFQueueID(queue.ID()))
			continue
		}

		c.activeConsumers[queue.ID()] = consumer
		refCounts[queue.ID()]++
		c.logger.Info("Created and started consumer", tag.WorkflowDomainName(domain.GetInfo().Name), tag.AsyncWFQueueID(queue.ID()))
	}

	// stop consumers that are not needed
	for qID, consumer := range c.activeConsumers {
		if refCounts[qID] > 0 {
			continue
		}

		c.logger.Info("Stopping consumer because it's not needed", tag.AsyncWFQueueID(qID))
		consumer.Stop()
		delete(c.activeConsumers, qID)
		c.logger.Info("Stopped consumer", tag.AsyncWFQueueID(qID))
	}

	c.logger.Info("Refreshed consumers", tag.Dynamic("consumer-count", len(c.activeConsumers)))
	c.metricsClient.Scope(metrics.AsyncWorkflowConsumerScope).UpdateGauge(metrics.AsyncWorkflowConsumerCount, float64(len(c.activeConsumers)))
}

func (c *ConsumerManager) getQueue(cfg types.AsyncWorkflowConfiguration) (provider.Queue, error) {
	if cfg.PredefinedQueueName != "" {
		return c.queueProvider.GetPredefinedQueue(cfg.PredefinedQueueName)
	}

	return c.queueProvider.GetQueue(cfg.QueueType, cfg.QueueConfig)
}
