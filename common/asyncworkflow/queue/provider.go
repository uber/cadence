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

package queue

import (
	"fmt"

	"github.com/uber/cadence/common/asyncworkflow/queue/provider"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/messaging"
)

type (
	providerImpl struct {
		domainCache cache.DomainCache

		// predefined queues in static config
		producers map[string]messaging.Producer
		consumers map[string]messaging.Consumer

		// cache for producers created on the fly for domains not using predefined queues
		// key is the domain name
		producerCache cache.Cache
	}
)

// NewAsyncQueueProvider returns a new async queue provider
func NewAsyncQueueProvider(cfg map[string]config.AsyncWorkflowQueueProvider, params *provider.Params) (Provider, error) {
	p := &providerImpl{
		producers: make(map[string]messaging.Producer),
		consumers: make(map[string]messaging.Consumer),
	}
	for queueName, queueCfg := range cfg {
		producerConstructor, ok := provider.GetAsyncQueueProducerProvider(queueCfg.Type)
		if !ok {
			return nil, fmt.Errorf("queue type %v not registered", queueCfg.Type)
		}
		producer, err := producerConstructor(queueCfg.Config, params)
		if err != nil {
			return nil, err
		}
		p.producers[queueName] = producer

		consumerConstructor, ok := provider.GetAsyncQueueConsumerProvider(queueCfg.Type)
		if !ok {
			return nil, fmt.Errorf("queue type %v not registered", queueCfg.Type)
		}
		consumer, err := consumerConstructor(queueCfg.Config, params)
		if err != nil {
			return nil, err
		}
		p.consumers[queueName] = consumer
	}
	return p, nil
}

func (p *providerImpl) GetAsyncQueueProducer(domain string) (messaging.Producer, error) {
	domainEntry, err := p.domainCache.GetDomain(domain)
	if err != nil {
		return nil, err
	}
	queue := domainEntry.GetConfig().AsyncWorkflowConfig.PredefinedQueueName
	if queue != "" {
		producer, ok := p.producers[queue]
		if !ok {
			return nil, fmt.Errorf("queue %v not found", queue)
		}
		return producer, nil
	}

	val := p.producerCache.Get(domain)
	if val != nil {
		return val.(messaging.Producer), nil
	}
	// TODO: create producer on the fly and add to cache
	return nil, fmt.Errorf("to be implemented")
}

func (p *providerImpl) GetAsyncQueueConsumer(domain string) (messaging.Consumer, error) {
	domainEntry, err := p.domainCache.GetDomain(domain)
	if err != nil {
		return nil, err
	}
	queue := domainEntry.GetConfig().AsyncWorkflowConfig.PredefinedQueueName
	if queue != "" {
		consumer, ok := p.consumers[queue]
		if !ok {
			return nil, fmt.Errorf("queue %v not found", queue)
		}
		return consumer, nil
	}
	// TODO: create consumer on the fly, we don't need to cache it
	return nil, fmt.Errorf("to be implemented")
}
