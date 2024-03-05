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
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/types"
)

type (
	providerImpl struct {
		queues map[string]provider.Queue
	}
)

// NewAsyncQueueProvider returns a new async queue provider
func NewAsyncQueueProvider(cfg map[string]config.AsyncWorkflowQueueProvider) (Provider, error) {
	p := &providerImpl{
		queues: make(map[string]provider.Queue),
	}
	for queueName, queueCfg := range cfg {
		queueConstructor, ok := provider.GetQueueProvider(queueCfg.Type)
		if !ok {
			return nil, fmt.Errorf("queue type %v not registered", queueCfg.Type)
		}
		queue, err := queueConstructor(queueCfg.Config)
		if err != nil {
			return nil, err
		}
		p.queues[queueName] = queue
	}
	return p, nil
}

func (p *providerImpl) GetPredefinedQueue(name string) (provider.Queue, error) {
	queue, ok := p.queues[name]
	if !ok {
		return nil, fmt.Errorf("queue %v not found", name)
	}
	return queue, nil
}

func (p *providerImpl) GetQueue(queueType string, queueConfig *types.DataBlob) (provider.Queue, error) {
	queueConfigDecoder, ok := provider.GetDecoder(queueType)
	if !ok {
		return nil, fmt.Errorf("queue type %v not registered", queueType)
	}
	decoder := queueConfigDecoder(queueConfig)
	queueConstructor, ok := provider.GetQueueProvider(queueType)
	if !ok {
		return nil, fmt.Errorf("queue type %v not registered", queueType)
	}
	return queueConstructor(decoder)
}
