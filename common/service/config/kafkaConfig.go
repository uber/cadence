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

package config

import (
	"fmt"

	"github.com/uber/cadence/common/auth"
)

type (
	// KafkaConfig describes the configuration needed to connect to all kafka clusters
	KafkaConfig struct {
		TLS      auth.TLS                 `yaml:"tls"`
		SASL     auth.SASL                `yaml:"sasl"`
		Clusters map[string]ClusterConfig `yaml:"clusters"`
		Topics   map[string]TopicConfig   `yaml:"topics"`
		// Applications describes the applications that will use the Kafka topics
		Applications map[string]TopicList `yaml:"applications"`
	}

	// ClusterConfig describes the configuration for a single Kafka cluster
	ClusterConfig struct {
		Brokers []string `yaml:"brokers"`
	}

	// TopicConfig describes the mapping from topic to Kafka cluster
	TopicConfig struct {
		Cluster string `yaml:"cluster"`
	}

	// TopicList describes the topic names for each cluster
	TopicList struct {
		Topic    string `yaml:"topic"`
		DLQTopic string `yaml:"dlq-topic"`
	}
)

// Validate will validate config for kafka
func (k *KafkaConfig) Validate(checkApp bool) {
	if len(k.Clusters) == 0 {
		panic("Empty Kafka Cluster Config")
	}
	if len(k.Topics) == 0 {
		panic("Empty Topics Config")
	}

	validateTopicsFn := func(topic string) {
		if topic == "" {
			panic("Empty Topic Name")
		} else if topicConfig, ok := k.Topics[topic]; !ok {
			panic(fmt.Sprintf("Missing Topic Config for Topic %v", topic))
		} else if clusterConfig, ok := k.Clusters[topicConfig.Cluster]; !ok {
			panic(fmt.Sprintf("Missing Kafka Cluster Config for Cluster %v", topicConfig.Cluster))
		} else if len(clusterConfig.Brokers) == 0 {
			panic(fmt.Sprintf("Missing Kafka Brokers Config for Cluster %v", topicConfig.Cluster))
		}
	}

	if checkApp {
		if len(k.Applications) == 0 {
			panic("Empty Applications Config")
		}
		for _, topics := range k.Applications {
			validateTopicsFn(topics.Topic)
			validateTopicsFn(topics.DLQTopic)
		}
	}
}

// GetKafkaClusterForTopic gets cluster from topic
func (k *KafkaConfig) GetKafkaClusterForTopic(topic string) string {
	return k.Topics[topic].Cluster
}

// GetBrokersForKafkaCluster gets broker from cluster
func (k *KafkaConfig) GetBrokersForKafkaCluster(kafkaCluster string) []string {
	return k.Clusters[kafkaCluster].Brokers
}

// GetTopicsForApplication gets topic from application
func (k *KafkaConfig) GetTopicsForApplication(app string) TopicList {
	return k.Applications[app]
}
