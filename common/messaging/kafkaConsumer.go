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
	"sync"

	"github.com/Shopify/sarama"
	"golang.org/x/net/context"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
)

const rcvBufferSize = 2 * 1024

type (
	// a wrapper of sarama consumer group for our consumer interface
	kafkaConsumer struct {
		topics          TopicList
		consumerHandler *consumerHandlerImpl
		consumerGroup   sarama.ConsumerGroup
		logger          log.Logger
		msgChan         <-chan Message
		cancelFunc      context.CancelFunc
	}

	// consumerHandlerImpl represents a Sarama consumer group consumer
	// It's for passing into sarama consumer group API
	consumerHandlerImpl struct {
		sync.RWMutex
		topic          string
		currentSession sarama.ConsumerGroupSession
		msgChan        chan<- Message
		manager        *partitionAckManager
		logger         log.Logger
	}

	messageImpl struct {
		saramaMsg *sarama.ConsumerMessage
		session   sarama.ConsumerGroupSession
		handler   *consumerHandlerImpl
		logger    log.Logger
	}
)

var _ Message = (*messageImpl)(nil)
var _ Consumer = (*kafkaConsumer)(nil)

// TODO add metrics
func newKafkaConsumer(
	kafkaConfig *KafkaConfig,
	topics TopicList,
	consumerName string,
	saramaConfig *sarama.Config,
	logger log.Logger,
) (Consumer, error) {
	clusterName := kafkaConfig.getKafkaClusterForTopic(topics.Topic)
	brokers := kafkaConfig.getBrokersForKafkaCluster(clusterName)
	consumerGroup, err := sarama.NewConsumerGroup(brokers, consumerName, saramaConfig)
	if err != nil {
		return nil, err
	}

	msgChan := make(chan Message, rcvBufferSize)
	consumerHandler := newConsumerHandlerImpl(topics.Topic, msgChan, logger)

	return &kafkaConsumer{
		topics: topics,

		consumerHandler: consumerHandler,
		consumerGroup:   consumerGroup,
		logger:          logger,
		msgChan:         msgChan,
	}, nil
}

func (c *kafkaConsumer) Start() error {

	ctx, cancel := context.WithCancel(context.Background())
	c.cancelFunc = cancel

	// consumer loop
	go func() {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := c.consumerGroup.Consume(ctx, []string{c.topics.Topic}, c.consumerHandler); err != nil {
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
func (c *kafkaConsumer) Stop() {
	c.logger.Info("Stopping consumer")
	c.cancelFunc()
	c.consumerHandler.stop()
}

// Messages return the message channel for this consumer
func (c *kafkaConsumer) Messages() <-chan Message {
	return c.msgChan
}

func newConsumerHandlerImpl(topic string, msgChan chan<- Message, logger log.Logger) *consumerHandlerImpl {
	return &consumerHandlerImpl{
		topic:   topic,
		logger:  logger,
		msgChan: msgChan,
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *consumerHandlerImpl) Setup(session sarama.ConsumerGroupSession) error {
	h.Lock()
	defer h.Unlock()

	h.currentSession = session
	h.manager = newPartitionAckManager(h.logger)
	h.logger.Info("start consumer group session", tag.Name(string(session.GenerationID())))
	return nil
}

func (h *consumerHandlerImpl) getCurrentSession() sarama.ConsumerGroupSession {
	h.RLock()
	defer h.RUnlock()

	return h.currentSession
}

func (h *consumerHandlerImpl) completeMessage(message *messageImpl) {
	h.RLock()
	defer h.RUnlock()

	ackLevel := h.manager.CompleteMessage(message.Partition(), message.Offset())
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
	m.handler.completeMessage(m)
	return nil
}

func (m *messageImpl) Nack() error {
	// TODO publish to DLQ

	if m.isFromPreviousSession() {
		return nil
	}
	m.handler.completeMessage(m)
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
