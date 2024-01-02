package kafka

import (
	"github.com/Shopify/sarama/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	"github.com/uber/cadence/common/config"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/metrics"
	"testing"
)

func TestNewConsumer(t *testing.T) {
	mockProducer := mocks.NewSyncProducer(t, nil)
	kafkaConfig := &config.KafkaConfig{
		Clusters: map[string]config.ClusterConfig{
			"test-cluster": {
				Brokers: []string{"test-brokers"},
			},
		},
		Topics: map[string]config.TopicConfig{
			"test-topic": {
				Cluster: "test-cluster",
			},
			"test-topic-dlq": {
				Cluster: "test-cluster",
			},
		},
		Applications: map[string]config.TopicList{
			"test-app": {
				Topic:    "test-topic",
				DLQTopic: "test-topic-dlq",
			},
		},
	}
	topic := "test-topic"
	consumerName := "test-consumer"
	metricsClient := metrics.NewClient(tally.NoopScope, metrics.History)
	logger := testlogger.New(t)
	kafkaProducer := NewKafkaProducer(topic, mockProducer, logger)
	_, err := newKafkaConsumer(kafkaProducer, kafkaConfig, topic, consumerName,
		nil, metricsClient, logger)
	// test will fail at the sarama.NewConsumerGroup which requires to connect to actual broker
	// actual functionality will be tested in the integration test
	assert.EqualError(t, err, "kafka: client has run out of available brokers to talk to: dial tcp: address test-brokers: missing port in address")
}

func TestNewConsumerHandlerImpl(t *testing.T) {
	topic := "test-topic"
	metricsClient := metrics.NewClient(tally.NoopScope, metrics.History)
	logger := testlogger.New(t)
	mockProducer := mocks.NewSyncProducer(t, nil)
	kafkaProducer := NewKafkaProducer(topic, mockProducer, logger)
	msgChan := make(chan messaging.Message, 1)
	consumerHandler := newConsumerHandlerImpl(kafkaProducer, topic, msgChan, metricsClient, logger)

	assert.NotNil(t, consumerHandler)
	assert.Equal(t, kafkaProducer, consumerHandler.dlqProducer)
	assert.Equal(t, topic, consumerHandler.topic)
	// Close the channel at the end of the test
	close(msgChan)
}
