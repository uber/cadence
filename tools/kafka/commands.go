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
	"fmt"
	"os"

	"io/ioutil"

	"strings"

	"strconv"

	"time"

	"github.com/Shopify/sarama"
	saramacluster "github.com/bsm/sarama-cluster"
	"github.com/fatih/color"
	"github.com/uber-go/kafka-client"
	"github.com/uber-go/kafka-client/kafka"
	"github.com/uber-go/tally"
	"github.com/uber/cadence/common/messaging"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	localKafka = "127.0.0.1:9092"
	heartBeat  = time.Second * 5
)

type (
	// ClustersConfig describes the kafka clusters
	ClustersConfig struct {
		Clusters map[string]messaging.ClusterConfig
	}
)

var (
	colorRed     = color.New(color.FgRed).SprintFunc()
	colorMagenta = color.New(color.FgMagenta).SprintFunc()
	colorGreen   = color.New(color.FgGreen).SprintFunc()
)

func mergeDLQ(c *cli.Context) {
	if !c.IsSet(FlagDLQTopic) {
		ErrorAndExit("", fmt.Errorf("DLQ topic name must be provided by flag %v", FlagDLQTopic))
	}
	dlqTopic := c.String(FlagDLQTopic)
	if !c.IsSet(FlagDestinationTopic) {
		ErrorAndExit("", fmt.Errorf("destination topic name must be provided by flag %v", FlagDestinationTopic))
	}
	destTopic := c.String(FlagDestinationTopic)

	dlqGroup := c.String(FlagDLQConsumerGroup)
	dlqCluster := c.String(FlagDLQCluster)
	destCluster := c.String(FlagDestinationCluster)
	clusterFile := c.GlobalString(FlagKafkaClusterFile)
	useLocal := c.GlobalBool(FlagLocal)
	concurrency := c.Int64(FlagConcurrency)
	brokers := map[string][]string{}
	var err error
	if useLocal {
		brokers = map[string][]string{
			dlqCluster:  {localKafka},
			destCluster: {localKafka},
		}
	} else {
		// if not using local mode, cluster names and host file must be provided
		if len(clusterFile) == 0 || len(dlqCluster) == 0 || len(destCluster) == 0 {
			ErrorAndExit("", fmt.Errorf("DLQ and destination cluster names and host file must be provided must be provided by flags %v,%v,%v", FlagDLQCluster, FlagDestinationCluster, FlagKafkaClusterFile))
		}
		brokers, err = loadClusterConfig(clusterFile, []string{dlqCluster, destCluster})
		if err != nil {
			ErrorAndExit("failed to load cluster from file", err)
		}
	}

	consumer, err := buildConsumer(brokers, dlqTopic, dlqCluster, dlqGroup, concurrency)
	if err != nil {
		ErrorAndExit("failed to create DLQ consumer", err)
	}

	producer, err := sarama.NewSyncProducer(brokers[destCluster], nil)
	if err != nil {
		ErrorAndExit("failed to create destination topic producer", err)
	}

	if err := consumer.Start(); err != nil {
		ErrorAndExit("", fmt.Errorf("failed to start DLQ consumer"))
	}

	for {
		select {
		case cmsg, ok := <-consumer.Messages():
			if !ok {
				return
			}

			pmsg := &sarama.ProducerMessage{
				Topic: destTopic,
				Key:   getKey(cmsg.Key()),
				Value: sarama.ByteEncoder(cmsg.Value()),
			}
			partition, offset, err := producer.SendMessage(pmsg)
			if err != nil {
				cmsg.Nack()
				fmt.Printf("[Error] Message [%v],[%v] failed to be sent to DLQ\n", cmsg.Partition(), cmsg.Offset())
			} else {
				cmsg.Ack()
				fmt.Printf("Message [%v],[%v] is sent to [%v],[%v] in DLQ\n", cmsg.Partition(), cmsg.Offset(), partition, offset)
			}
		case <-time.After(heartBeat):
			fmt.Println("heartbeat: waiting for messages...")
		}
	}
}

// use the same key as we sent to DLQ
func getKey(originKey []byte) sarama.Encoder {
	if originKey == nil || len(originKey) == 0 {
		return nil
	}
	return sarama.StringEncoder(string(originKey))
}

func purgeTopic(c *cli.Context) {
	if !c.IsSet(FlagOffsets) {
		ErrorAndExit("", fmt.Errorf("target topic name must be provided by flag %v", FlagTopic))
	}

	offsetStr := c.String(FlagOffsets)
	offsets, err := parseOffsetStr(offsetStr)
	if err != nil {
		ErrorAndExit("input offset is not valid", err)
	}

	setTopicOffsets(c, offsets, true)
}

func setTopicOffsets(c *cli.Context, offsetPerPartition map[int32]int64, isForward bool) {
	if !c.IsSet(FlagTopic) {
		ErrorAndExit("", fmt.Errorf("target topic name must be provided by flag %v", FlagTopic))
	}
	topic := c.String(FlagTopic)
	if !c.IsSet(FlagConsumerGroup) {
		ErrorAndExit("", fmt.Errorf("target consumer group name must be provided by flag %v", FlagConsumerGroup))
	}
	group := c.String(FlagConsumerGroup)

	cluster := c.String(FlagCluster)
	clusterFile := c.GlobalString(FlagKafkaClusterFile)
	useLocal := c.GlobalBool(FlagLocal)
	brokers := map[string][]string{}
	var err error
	if useLocal {
		brokers = map[string][]string{
			cluster: {localKafka},
		}
	} else {
		// if not using local mode, cluster names and host file must be provided
		if len(clusterFile) == 0 || len(cluster) == 0 {
			ErrorAndExit("", fmt.Errorf("cluster name and host file must be provided must be provided by flags %v,%v", FlagCluster, FlagKafkaClusterFile))
		}
		brokers, err = loadClusterConfig(clusterFile, []string{cluster})
		if err != nil {
			ErrorAndExit("failed to load cluster from file", err)
		}
	}

	config := saramacluster.NewConfig()
	config.Group.Mode = saramacluster.ConsumerModePartitions
	consumer, err := saramacluster.NewConsumer(brokers[cluster], group, []string{topic}, config)

outloop:
	for {
		select {
		case <-time.After(heartBeat):
			_, ok := consumer.Subscriptions()[topic]
			if !ok {
				fmt.Println("heartbeat: waiting for subscriptions...")
				continue
			}
			fmt.Println("subs:", consumer.Subscriptions())
			break outloop
		}
	}

	for p, off := range offsetPerPartition {
		fmt.Printf("partition %v setting to offset %v \n", p, off)
		if isForward {
			consumer.MarkPartitionOffset(topic, p, off, "")
		} else {
			consumer.ResetPartitionOffset(topic, p, off, "")
		}
	}

	if err := consumer.CommitOffsets(); err != nil {
		ErrorAndExit("failed to commit offsets", err)
	}

	fmt.Printf("mark offset for topic %v group %v is completed\n", topic, group)
}

func resetTopic(c *cli.Context) {
	if !c.IsSet(FlagOffsets) {
		ErrorAndExit("", fmt.Errorf("target topic name must be provided by flag %v", FlagTopic))
	}
	offsetStr := c.String(FlagOffsets)
	offsets, err := parseOffsetStr(offsetStr)
	if err != nil {
		ErrorAndExit("input offset is not valid", err)
	}
	setTopicOffsets(c, offsets, false)
}

func parseOffsetStr(offsetStr string) (map[int32]int64, error) {
	pss := strings.Split(offsetStr, ",")
	ret := map[int32]int64{}
	for _, ps := range pss {
		if !strings.Contains(ps, ":") {
			return nil, fmt.Errorf("offsets should be in format of partition:offset")
		}
		po := strings.Split(ps, ":")
		partition, err := strconv.Atoi(po[0])
		if err != nil {
			return nil, fmt.Errorf("%v is not a valid integer", po[0])
		}
		offset, err := strconv.ParseInt(po[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%v is not a valid integer", po[1])
		}
		ret[int32(partition)] = offset
	}
	return ret, nil
}

func buildConsumer(brokers map[string][]string, topic string, cluster string, group string, concurrency int64) (kafka.Consumer, error) {
	topicClusterAssignment := map[string][]string{
		topic: {cluster},
	}
	client := kafkaclient.New(kafka.NewStaticNameResolver(topicClusterAssignment, brokers), zap.NewNop(), tally.NoopScope)

	dlqConsumerConfig := &kafka.ConsumerConfig{
		TopicList: kafka.ConsumerTopicList{
			kafka.ConsumerTopic{
				Topic: kafka.Topic{
					Name:    topic,
					Cluster: cluster,
				},
			},
		},
		GroupName:   group,
		Concurrency: int(concurrency),
	}

	consumer, err := client.NewConsumer(dlqConsumerConfig)
	return consumer, err
}

// load the kafka cluster config file from host and convert it to a useable config
func loadClusterConfig(clusterFile string, clusterNames []string) (map[string][]string, error) {
	contents, err := ioutil.ReadFile(clusterFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load kafka cluster info from %v., error: %v", clusterFile, err)
	}
	clustersConfig := ClustersConfig{}
	if err := yaml.Unmarshal(contents, &clustersConfig); err != nil {
		return nil, err
	}
	ret := map[string][]string{}
	for _, name := range clusterNames {
		c, ok := clustersConfig.Clusters[name]
		if !ok {
			return nil, fmt.Errorf("failed to load cluster %v from file %v", name, clusterFile)
		}
		brokers := []string{}
		for i := range c.Brokers {
			if !strings.Contains(c.Brokers[i], ":") {
				brokers = append(brokers, c.Brokers[i]+":9092")
			} else {
				brokers = append(brokers, c.Brokers[i])
			}
		}
		ret[name] = brokers
	}
	return ret, nil
}

// ErrorAndExit print easy to understand error msg first then error detail in a new line
func ErrorAndExit(msg string, err error) {
	fmt.Printf("%s %s\n%s %+v\n", colorRed("Error:"), msg, colorMagenta("Error Details:"), err)
	os.Exit(1)
}
