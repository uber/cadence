package kafka

import (
	"fmt"
	"os"

	"io/ioutil"

	"github.com/uber-go/kafka-client"
	"github.com/uber-go/kafka-client/kafka"
	"github.com/uber-go/tally"
	"go.uber.org/zap"

	"strings"

	"os/signal"

	"github.com/Shopify/sarama"
	"github.com/fatih/color"
	kafkaclient "github.com/uber-go/kafka-client"
	"github.com/uber-go/kafka-client/kafka"
	"github.com/uber-go/tally"
	"github.com/uber/cadence/common/messaging"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	yaml "gopkg.in/yaml.v2"
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
	clusterFile := c.String(FlagKafkaClusterFile)
	useLocal := c.Bool(FlagLocal)
	concurrency := c.Int64(FlagConcurrency)
	brokers := map[string][]string{}
	var err error
	if useLocal {
		brokers = map[string][]string{
			dlqCluster:  {"127.0.0.1:9092"},
			destCluster: {"127.0.0.1:9092"},
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

	topicClusterAssignment := map[string][]string{
		dlqTopic:  {dlqCluster},
		destTopic: {destCluster},
	}
	client := kafkaclient.New(kafka.NewStaticNameResolver(topicClusterAssignment, brokers), zap.NewNop(), tally.NoopScope)

	dlqConsumerConfig := &kafka.ConsumerConfig{
		TopicList: kafka.ConsumerTopicList{
			kafka.ConsumerTopic{
				Topic: kafka.Topic{
					Name:    dlqTopic,
					Cluster: dlqCluster,
				},
			},
		},
		GroupName:   dlqGroup,
		Concurrency: int(concurrency),
	}

	consumer, err := client.NewConsumer(dlqConsumerConfig)
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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

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
			} else {
				cmsg.Ack()
				fmt.Println("Message [%v],[%v] is sent to [%v],[%v] in DLQ", cmsg.Partition(), cmsg.Offset(), partition, offset)
			}
		case <-sigCh:
			consumer.Stop()
			<-consumer.Closed()
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
	if !c.IsSet(FlagTopic) {
		ErrorAndExit("", fmt.Errorf("target topic name must be provided by flag %v", FlagTopic))
	}
	if !c.IsSet(FlagConsumerGroup) {
		ErrorAndExit("", fmt.Errorf("target consumer group name must be provided by flag %v", FlagConsumerGroup))
	}
}

func resetTopic(c *cli.Context) {
	if !c.IsSet(FlagTopic) {
		ErrorAndExit("", fmt.Errorf("target topic name must be provided by flag %v", FlagTopic))
	}
	if !c.IsSet(FlagConsumerGroup) {
		ErrorAndExit("", fmt.Errorf("target consumer group name must be provided by flag %v", FlagConsumerGroup))
	}
}

func buildConsumer() {

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
