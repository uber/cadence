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

package provider

import (
	"fmt"

	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/syncmap"
)

type (
	Params struct {
		Logger        log.Logger
		MetricsClient metrics.Client
	}

	// ProducerConstructor is a function that constructs a queue producer
	ProducerConstructor func(*config.YamlNode, *Params) (messaging.Producer, error)

	// ConsumerConstructor is a function that constructs a queue consumer
	ConsumerConstructor func(*config.YamlNode, *Params) (messaging.Consumer, error)
)

var (
	asyncQueueProducerConstructors = syncmap.New[string, ProducerConstructor]()
	asyncQueueConsumerConstructors = syncmap.New[string, ConsumerConstructor]()
)

// RegisterAsyncQueueProducerProvider registers a queue producer constructor for a given queue type
func RegisterAsyncQueueProducerProvider(queueType string, producerConstructor ProducerConstructor) error {
	inserted := asyncQueueProducerConstructors.Put(queueType, producerConstructor)
	if !inserted {
		return fmt.Errorf("queue type %v already registered", queueType)
	}
	return nil
}

// GetAsyncQueueProducerProvider returns a queue producer constructor for a given queue type
func GetAsyncQueueProducerProvider(queueType string) (ProducerConstructor, bool) {
	return asyncQueueProducerConstructors.Get(queueType)
}

// RegisterAsyncQueueConsumerProvider registers a queue consumer constructor for a given queue type
func RegisterAsyncQueueConsumerProvider(queueType string, consumerConstructor ConsumerConstructor) error {
	inserted := asyncQueueConsumerConstructors.Put(queueType, consumerConstructor)
	if !inserted {
		return fmt.Errorf("queue type %v already registered", queueType)
	}
	return nil
}

// GetAsyncQueueConsumerProvider returns a queue consumer constructor for a given queue type
func GetAsyncQueueConsumerProvider(queueType string) (ConsumerConstructor, bool) {
	return asyncQueueConsumerConstructors.Get(queueType)
}
