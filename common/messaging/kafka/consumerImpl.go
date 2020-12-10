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

package kafka

import (
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"golang.org/x/net/context"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/service/config"
)

const rcvBufferSize = 2 * 1024
const dlqPublishTimeout = time.Minute

type (
	// a wrapper of sarama consumer group for our consumer interface
	consumerImpl struct {
		topic           string
		consumerHandler *consumerHandlerImpl
		consumerGroup   sarama.ConsumerGroup
		msgChan         <-chan messaging.Message
		cancelFunc      context.CancelFunc

		logger log.Logger
	}

	// consumerHandlerImpl represents a Sarama consumer group consumer
	// It's for passing into sarama consumer group API
	consumerHandlerImpl struct {
		sync.RWMutex
		dlqProducer messaging.Producer

		topic          string
		currentSession sarama.ConsumerGroupSession
		msgChan        chan<- messaging.Message
		manager        *partitionAckManager

		metricsClient metrics.Client
		logger        log.Logger
	}

	messageImpl struct {
		saramaMsg *sarama.ConsumerMessage
		session   sarama.ConsumerGroupSession
		handler   *consumerHandlerImpl
		logger    log.Logger
	}
)

var _ messaging.Message = (*messageImpl)(nil)
var _ messaging.Consumer = (*consumerImpl)(nil)

func newKafkaConsumer(
	dlqProducer messaging.Producer,
	kafkaConfig *config.KafkaConfig,
	topic string,
	consumerName string,
	saramaConfig *sarama.Config,
	metricsClient metrics.Client,
	logger log.Logger,
) (messaging.Consumer, error) {
	clusterName := kafkaConfig.GetKafkaClusterForTopic(topic)
	brokers := kafkaConfig.GetBrokersForKafkaCluster(clusterName)
	consumerGroup, err := sarama.NewConsumerGroup(brokers, consumerName, saramaConfig)
	if err != nil {
		return nil, err
	}

	msgChan := make(chan messaging.Message, rcvBufferSize)
	consumerHandler := newConsumerHandlerImpl(dlqProducer, topic, msgChan, metricsClient, logger)

	return &consumerImpl{
		topic: topic,

		consumerHandler: consumerHandler,
		consumerGroup:   consumerGroup,
		msgChan:         msgChan,

		logger: logger,
	}, nil
}

func (c *consumerImpl) Start() error {

	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFunc = cancel

	// consumer loop
	go func() {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := c.consumerGroup.Consume(ctx, []string{c.topic}, c.consumerHandler); err != nil {
				c.logger.Error("Error from consumer: %v", tag.Error(err))
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				c.logger.Info("context was cancel, stopping consumer loop")
				return
			}
		}
	}()
	return nil
}

// Stop stops the consumer
func (c *consumerImpl) Stop() {
	c.logger.Info("Stopping consumer")
	c.cancelFunc()
	c.consumerHandler.stop()
}

// Messages return the message channel for this consumer
func (c *consumerImpl) Messages() <-chan messaging.Message {
	return c.msgChan
}

func newConsumerHandlerImpl(
	dlqProducer messaging.Producer,
	topic string,
	msgChan chan<- messaging.Message,
	metricsClient metrics.Client,
	logger log.Logger,
) *consumerHandlerImpl {
	return &consumerHandlerImpl{
		dlqProducer: dlqProducer,
		topic:       topic,
		msgChan:     msgChan,

		metricsClient: metricsClient,
		logger:        logger,
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *consumerHandlerImpl) Setup(session sarama.ConsumerGroupSession) error {
	h.Lock()
	defer h.Unlock()

	h.currentSession = session
	h.manager = newPartitionAckManager(h.metricsClient, h.logger)
	h.logger.Info("start consumer group session", tag.Name(string(session.GenerationID())))
	h.metricsClient.IncCounter(metrics.MessagingClientConsumerScope, metrics.KafkaConsumerSessionStart)

	return nil
}

func (h *consumerHandlerImpl) getCurrentSession() sarama.ConsumerGroupSession {
	h.RLock()
	defer h.RUnlock()

	return h.currentSession
}

func (h *consumerHandlerImpl) completeMessage(message *messageImpl, isAck bool) {
	h.RLock()
	defer h.RUnlock()

	if !isAck {
		op := func() error {
			// NOTE: current KafkaProducer is not taking use the this context, because saramaProducer doesn't support it
			// https://github.com/Shopify/sarama/issues/1849
			ctx, cancel := context.WithTimeout(context.Background(), dlqPublishTimeout)
			err := h.dlqProducer.Publish(ctx, message.saramaMsg)
			cancel()
			return err
		}
		err := backoff.Retry(op, common.CreateDlqPublishRetryPolicy(), nil)
		if err != nil {
			h.metricsClient.IncCounter(metrics.MessagingClientConsumerScope, metrics.KafkaConsumerMessageNackDlqErr)
			h.logger.Error("Fail to publish message to DLQ when nacking message, please take action!!",
				tag.KafkaPartition(message.Partition()),
				tag.KafkaOffset(message.Offset()))
		} else {
			h.logger.Warn("nack message and publish to DLQ",
				tag.KafkaPartition(message.Partition()),
				tag.KafkaOffset(message.Offset()))
		}
	}
	ackLevel := h.manager.CompleteMessage(message.Partition(), message.Offset(), isAck)
	h.currentSession.MarkOffset(h.topic, message.Partition(), ackLevel+1, "")
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *consumerHandlerImpl) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (h *consumerHandlerImpl) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	h.RLock()
	defer h.RUnlock()

	// NOTE: Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine:
	for message := range claim.Messages() {
		h.manager.AddMessage(message.Partition, message.Offset)
		h.msgChan <- &messageImpl{
			saramaMsg: message,
			session:   session,
			handler:   h,
			logger:    h.logger,
		}
	}

	return nil
}

func (h *consumerHandlerImpl) stop() {
	close(h.msgChan)
}

func (m *messageImpl) Value() []byte {
	return m.saramaMsg.Value
}

func (m *messageImpl) Partition() int32 {
	return m.saramaMsg.Partition
}

func (m *messageImpl) Offset() int64 {
	return m.saramaMsg.Offset
}

func (m *messageImpl) Ack() error {
	if m.isFromPreviousSession() {
		return nil
	}
	m.handler.completeMessage(m, true)
	return nil
}

func (m *messageImpl) Nack() error {
	if m.isFromPreviousSession() {
		return nil
	}
	m.handler.completeMessage(m, false)
	return nil
}

func (m *messageImpl) isFromPreviousSession() bool {
	if m.session.GenerationID() != m.handler.getCurrentSession().GenerationID() {
		m.logger.Warn("Skip message that is from a previous session",
			tag.KafkaPartition(m.saramaMsg.Partition),
			tag.KafkaOffset(m.saramaMsg.Offset),
			tag.Name(string(m.session.GenerationID())),
			tag.Value(m.handler.getCurrentSession().GenerationID()),
		)
		return true
	}
	return false
}
