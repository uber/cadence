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
	"github.com/urfave/cli"
)

const (
	FlagKafkaClusterFile          = "clusterfile"
	FlagKafkaClusterFileWithAlias = FlagKafkaClusterFile + ", f"

	FlagLocal          = "local"
	FlagLocalWithAlias = FlagLocal + ", loc"

	FlagCluster       = "cluster"
	FlagTopic         = "topic"
	FlagConsumerGroup = "consumerGroup"

	FlagDLQCluster       = "dlqCluster"
	FlagDLQTopic         = "dlqTopic"
	FlagDLQConsumerGroup = "dlqConsumerGroup"

	FlagDestinationCluster = "destCluster"
	FlagDestinationTopic   = "destTopic"

	FlagOffsets          = "offsets"
	FlagOffsetsWithAlias = FlagOffsets + ", off"

	FlagConcurrency = "concurrency"
)

// RunTool runs the cadence-cassandra-tool command line tool
func RunTool(args []string) error {
	app := buildCLIOptions()
	return app.Run(args)
}

func buildCLIOptions() *cli.App {

	app := cli.NewApp()
	app.Name = "cadence-kafka-tool"
	app.Usage = "Command line tool for cadence kafka operations."
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   FlagKafkaClusterFileWithAlias,
			Usage:  "A file that contains ip:port of brokers that are connecting to. See kafka_clusters.yaml as an example.",
			EnvVar: "KAFKA_CLUSTERS_FILE",
		},
		cli.BoolFlag{
			Name:  FlagLocalWithAlias,
			Usage: "Use local kafka broker: 127.0.0.1:9092 as the broker. This will ignore Kafka cluster file (" + FlagKafkaClusterFileWithAlias + ") if it is set",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "mergeDLQ",
			Usage: "Merge DLQ topic into another topic",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagDLQCluster,
					Usage: "The DLQ cluster that is operated on.",
				},
				cli.StringFlag{
					Name:  FlagDLQTopic,
					Usage: "The DLQ topic that is operated on.",
				},
				cli.StringFlag{
					Name:  FlagDLQConsumerGroup,
					Usage: "The DLQ consumer group.",
					Value: "DLQConsumerGroup",
				},

				cli.StringFlag{
					Name:  FlagDestinationCluster,
					Usage: "The Destination cluster that is operated on.",
				},
				cli.StringFlag{
					Name:  FlagDestinationTopic,
					Usage: "The Destination topic that is operated on.",
				},
				cli.Int64Flag{
					Name:  FlagConcurrency,
					Usage: "number of go routines processing messages in parallel",
					Value: 10,
				},
			},
			Action: func(c *cli.Context) {
				mergeDLQ(c)
			},
		},
		{
			Name:  "purge",
			Usage: "Purge a kafka topic",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagCluster,
					Usage: "The cluster that is operated on.",
				},
				cli.StringFlag{
					Name:  FlagTopic,
					Usage: "The topic that is operated on.",
				},
				cli.StringFlag{
					Name:  FlagConsumerGroup,
					Usage: "The consumer group.",
				},
			},
			Action: func(c *cli.Context) {
				purgeTopic(c)
			},
		},
		{
			Name:  "reset",
			Usage: "Reset offsets of a kafka topic.",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  FlagCluster,
					Usage: "The cluster that is operated on.",
				},
				cli.StringFlag{
					Name:  FlagTopic,
					Usage: "The topic that is operated on.",
				},
				cli.StringFlag{
					Name:  FlagConsumerGroup,
					Usage: "The consumer group.",
				},
				cli.StringFlag{
					Name:  FlagOffsetsWithAlias,
					Usage: "The offsets for each partition in format of p0:off0,p1:off1. Or '*:off' will set an offset for all partitions ",
				},
			},
			Action: func(c *cli.Context) {
				resetTopic(c)
			},
		},
	}
	return app
}
