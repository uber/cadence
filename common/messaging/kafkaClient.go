// Copyright (c) 2017 Uber Technologies, Inc.
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

package messaging

import (
	"encoding/json"

	"github.com/Shopify/sarama"
	"github.com/uber-common/bark"
	uberKafkaClient "github.com/uber-go/kafka-client"
	uberKafka "github.com/uber-go/kafka-client/kafka"
	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/common/metrics"
)

const rcvBufferSize = 2 * 1024

type (
	// This is a default implementation of Client interface which makes use of uber-go/kafka-client as consumer
	kafkaClient struct {
		config *KafkaConfig
		client uberKafkaClient.Client
		logger bark.Logger
	}

	kafkaConsumer struct {
		uConsumer uberKafka.Consumer
		logger    bark.Logger
		msgC      chan Message
		doneC     chan struct{}
	}

	kafkaMessage struct {
		uMsg uberKafka.Message
	}
)

var _ Client = (*kafkaClient)(nil)
var _ Consumer = (*kafkaConsumer)(nil)
var _ Message = (*kafkaMessage)(nil)

func newKafkaMessage(uMsg uberKafka.Message) Message {
	return &kafkaMessage{
		uMsg: uMsg,
	}
}

func newKafkaConsumer(uConsumer uberKafka.Consumer, metricsClient metrics.Client, logger bark.Logger) Consumer {
	return &kafkaConsumer{
		uConsumer: uConsumer,
		logger:    logger,
		msgC:      make(chan Message, rcvBufferSize),
		doneC:     make(chan struct{}),
	}
}

func (m *kafkaMessage) Value() []byte {
	return m.uMsg.Value()
}

// Partition is the ID of the partition from which the message was read.
func (m *kafkaMessage) Partition() int32 {
	return m.uMsg.Partition()
}

// Offset is the message's offset.
func (m *kafkaMessage) Offset() int64 {
	return m.uMsg.Offset()
}

// Ack marks the message as successfully processed.
func (m *kafkaMessage) Ack() error {
	return m.uMsg.Ack()
}

// Nack marks the message processing as failed and the message will be retried or sent to DLQ.
func (m *kafkaMessage) Nack() error {
	return m.uMsg.Nack()
}

func (c *kafkaConsumer) Start() error {
	if err := c.uConsumer.Start(); err != nil {
		return err
	}
	for {
		select {
		case <-c.doneC:
			return nil
		case uMsg := <-c.uConsumer.Messages():
			c.msgC <- uMsg
		}
	}
}

// Stop stops the consumer
func (c *kafkaConsumer) Stop() {
	close(c.doneC)
	c.uConsumer.Stop()
}

// Messages return the message channel for this consumer
func (c *kafkaConsumer) Messages() <-chan Message {
	return c.msgC
}

func Deserialize(payload []byte) (*replicator.ReplicationTask, error) {
	var task replicator.ReplicationTask
	if err := json.Unmarshal(payload, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// NewConsumer is used to create a Kafka consumer
func (c *kafkaClient) NewConsumer(currentCluster, sourceCluster, consumerName string, concurrency int) (Consumer, error) {
	currentTopics := c.config.getTopicsForCadenceCluster(currentCluster)
	sourceTopics := c.config.getTopicsForCadenceCluster(sourceCluster)

	topicKafkaCluster := c.config.getKafkaClusterForTopic(sourceTopics.Topic)
	dqlTopicKafkaCluster := c.config.getKafkaClusterForTopic(currentTopics.DLQTopic)
	topicList := uberKafka.ConsumerTopicList{
		uberKafka.ConsumerTopic{
			Topic: uberKafka.Topic{
				Name:       sourceTopics.Topic,
				Cluster:    topicKafkaCluster,
				BrokerList: c.config.getBrokersForKafkaCluster(topicKafkaCluster),
			},
			DLQ: uberKafka.Topic{
				Name:       currentTopics.DLQTopic,
				Cluster:    dqlTopicKafkaCluster,
				BrokerList: c.config.getBrokersForKafkaCluster(dqlTopicKafkaCluster),
			},
		},
	}

	consumerConfig := uberKafka.NewConsumerConfig(consumerName, topicList)
	consumerConfig.Concurrency = concurrency
	consumerConfig.Offsets.Initial.Offset = uberKafka.OffsetNewest

	uConsumer, err := c.client.NewConsumer(consumerConfig)
	if err != nil {
		return nil, err
	}
	return newKafkaConsumer(uConsumer, c.metricsClient, c.logger), nil
}

// NewProducer is used to create a Kafka producer for shipping replication tasks
func (c *kafkaClient) NewProducer(sourceCluster string) (Producer, error) {
	topics := c.config.getTopicsForCadenceCluster(sourceCluster)
	kafkaClusterName := c.config.getKafkaClusterForTopic(topics.Topic)
	brokers := c.config.getBrokersForKafkaCluster(kafkaClusterName)

	producer, err := sarama.NewSyncProducer(brokers, nil)
	if err != nil {
		return nil, err
	}

	return NewKafkaProducer(topics.Topic, producer, c.logger), nil
}
