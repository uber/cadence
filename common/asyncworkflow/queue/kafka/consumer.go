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
	"time"

	"github.com/Shopify/sarama"

	"github.com/uber/cadence/common/asyncworkflow/queue/provider"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/messaging/kafka"
	"github.com/uber/cadence/common/metrics"
)

// ConsumerConstructor is a function that constructs a queue consumer
func ConsumerConstructor(cfg *config.YamlNode, params *provider.Params) (messaging.Consumer, error) {
	return newConsumerConstructor(params.MetricsClient, params.Logger)(cfg)

}
func newConsumerConstructor(metricsClient metrics.Client, logger log.Logger) func(cfg *config.YamlNode) (messaging.Consumer, error) {
	return func(cfg *config.YamlNode) (messaging.Consumer, error) {
		var out *QueueConfig
		if err := cfg.Decode(&out); err != nil {
			return nil, fmt.Errorf("bad config: %w", err)
		}
		consumerGroup := fmt.Sprintf("%s-consumer", out.Topic)
		dlqTopic := fmt.Sprintf("%s-dlq", out.Topic)
		saramaConfig, err := newSaramaConfigWithAuth(&out.Connection.TLS, &out.Connection.SASL)
		if err != nil {
			return nil, err
		}
		saramaConfig.Consumer.Fetch.Default = 30 * 1024 * 1024 // 30MB.
		saramaConfig.Consumer.Return.Errors = true
		saramaConfig.Consumer.Offsets.CommitInterval = time.Second
		saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
		saramaConfig.Consumer.MaxProcessingTime = 250 * time.Millisecond

		dlqConfig, err := newSaramaConfigWithAuth(&out.Connection.TLS, &out.Connection.SASL)
		if err != nil {
			return nil, err
		}
		dlqConfig.Producer.Return.Successes = true
		dlqProducer, err := newProducer(dlqTopic, out.Connection.Brokers, saramaConfig, metricsClient, logger)
		return kafka.NewKafkaConsumer(dlqProducer, out.Connection.Brokers, out.Topic, consumerGroup, saramaConfig, metricsClient, logger)
	}
}
