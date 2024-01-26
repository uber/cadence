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

package kafka

import (
	"fmt"

	"github.com/Shopify/sarama"

	"github.com/uber/cadence/common/asyncworkflow/queue/provider"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/messaging/kafka"
	"github.com/uber/cadence/common/metrics"
)

// ProducerConstructor is a function that constructs a queue producer
func ProducerConstructor(cfg *config.YamlNode, params *provider.Params) (messaging.Producer, error) {
	return newProducerConstructor(params.MetricsClient, params.Logger)(cfg)
}

// newProducerConstructor returns a new kafka producer constructor for async queue
func newProducerConstructor(metricsClient metrics.Client, logger log.Logger) func(cfg *config.YamlNode) (messaging.Producer, error) {
	return func(cfg *config.YamlNode) (messaging.Producer, error) {
		var out *QueueConfig
		if err := cfg.Decode(&out); err != nil {
			return nil, fmt.Errorf("bad config: %w", err)
		}

		config, err := newSaramaConfigWithAuth(&out.Connection.TLS, &out.Connection.SASL)
		if err != nil {
			return nil, err
		}
		config.Producer.Return.Successes = true
		return newProducer(out.Topic, out.Connection.Brokers, config, metricsClient, logger)
	}
}

func newProducer(topic string, brokers []string, saramaConfig *sarama.Config, metricsClient metrics.Client, logger log.Logger) (messaging.Producer, error) {
	p, err := sarama.NewSyncProducer(brokers, saramaConfig)
	if err != nil {
		return nil, err
	}
	return messaging.NewMetricProducer(kafka.NewKafkaProducer(topic, p, logger), metricsClient), nil
}
