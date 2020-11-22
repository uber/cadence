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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/uber-go/tally"

	"github.com/uber/cadence/common/auth"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
)

type (
	// This is a default implementation of Client interface which makes use of uber-go/kafka-client as consumer
	kafkaClient struct {
		config        *KafkaConfig
		metricsClient metrics.Client
		logger        log.Logger
	}
)

var _ Client = (*kafkaClient)(nil)

// NewKafkaClient is used to create an instance of KafkaClient
func NewKafkaClient(
	kc *KafkaConfig,
	metricsClient metrics.Client,
	logger log.Logger,
	_ tally.Scope,
	checkApp bool,
) Client {
	kc.Validate(checkApp)

	// mapping from cluster name to list of broker ip addresses
	brokers := map[string][]string{}
	for cluster, cfg := range kc.Clusters {
		brokers[cluster] = cfg.Brokers
		for i := range brokers[cluster] {
			if !strings.Contains(cfg.Brokers[i], ":") {
				cfg.Brokers[i] += ":9092"
			}
		}
	}

	// mapping from topic name to cluster that has that topic
	topicClusterAssignment := map[string][]string{}
	for topic, cfg := range kc.Topics {
		topicClusterAssignment[topic] = []string{cfg.Cluster}
	}

	return &kafkaClient{
		config:        kc,
		metricsClient: metricsClient,
		logger:        logger,
	}
}

// NewConsumer is used to create a Kafka consumer
func (c *kafkaClient) NewConsumer(app, consumerName string, _ int) (Consumer, error) {
	topics := c.config.getTopicsForApplication(app)
	saramaConfig := sarama.NewConfig()
	// bellow config is copied from uber/kafka-client bo keep the same behavior
	saramaConfig.Version = sarama.V0_10_2_0
	saramaConfig.Consumer.Fetch.Default = 30 * 1024 * 1024 // 30MB.
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.CommitInterval = time.Second
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.MaxProcessingTime = 250 * time.Millisecond

	err := c.initAuth(saramaConfig)
	if err != nil {
		return nil, err
	}
	return newKafkaConsumer(c.config, topics, consumerName, saramaConfig, c.logger)
}

// NewProducer is used to create a Kafka producer
func (c *kafkaClient) NewProducer(app string) (Producer, error) {
	topics := c.config.getTopicsForApplication(app)
	return c.newProducerHelper(topics.Topic)
}

func (c *kafkaClient) newProducerHelper(topic string) (Producer, error) {
	kafkaClusterName := c.config.getKafkaClusterForTopic(topic)
	brokers := c.config.getBrokersForKafkaCluster(kafkaClusterName)

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	err := c.initAuth(config)
	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	if c.metricsClient != nil {
		c.logger.Info("Create producer with metricsClient")
		return NewMetricProducer(NewKafkaProducer(topic, producer, c.logger), c.metricsClient), nil
	}
	return NewKafkaProducer(topic, producer, c.logger), nil
}

func (c *kafkaClient) initAuth(saramaConfig *sarama.Config) error {
	tlsConfig, err := convertTLSConfig(c.config.TLS)
	if err != nil {
		panic(fmt.Sprintf("Error creating Kafka TLS config %v", err))
	}

	// TLS support
	saramaConfig.Net.TLS.Enable = tlsConfig != nil
	saramaConfig.Net.TLS.Config = tlsConfig

	// SASL support
	saramaConfig.Net.SASL.Enable = c.config.SASL.Enabled
	saramaConfig.Net.SASL.User = c.config.SASL.User
	saramaConfig.Net.SASL.Password = c.config.SASL.Password
	saramaConfig.Net.SASL.Handshake = true

	if c.config.SASL.Algorithm == "sha512" {
		saramaConfig.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &auth.XDGSCRAMClient{HashGeneratorFcn: auth.SHA512} }
		saramaConfig.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
	} else if c.config.SASL.Algorithm == "sha256" {
		saramaConfig.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &auth.XDGSCRAMClient{HashGeneratorFcn: auth.SHA256} }
		saramaConfig.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
	} else {
		return fmt.Errorf("invalid SHA algorithm %s: can be either sha256 or sha512", c.config.SASL.Algorithm)
	}
	return nil
}

// convertTLSConfig convert tls config
func convertTLSConfig(tlsConfig auth.TLS) (*tls.Config, error) {
	if !tlsConfig.Enabled {
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(tlsConfig.CertFile, tlsConfig.KeyFile)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	pemData, err := ioutil.ReadFile(tlsConfig.CaFile)
	if err != nil {
		return nil, err
	}
	caCertPool.AppendCertsFromPEM(pemData)

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}, nil
}
