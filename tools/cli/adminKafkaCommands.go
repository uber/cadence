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

package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
	"github.com/uber-common/bark"
	"github.com/uber/cadence/.gen/go/replicator"
	"github.com/uber/cadence/common/messaging"
	"github.com/urfave/cli"
	"go.uber.org/thriftrw/protocol"
	"go.uber.org/thriftrw/wire"
	"gopkg.in/yaml.v2"
)

type filterFn func(*replicator.ReplicationTask) bool

const (
	bufferSize            = 4096
	preambleVersion0 byte = 0x59
	malformedMessage      = "Input was malformed"
	chanBufferSize        = 10000
)

var (
	r = regexp.MustCompile(`Partition: .*?, Offset: .*?, Key: .*?`)
)

// AdminKafkaParse parses the output of k8read and outputs replication tasks
func AdminKafkaParse(c *cli.Context) {
	inputFile := getInputFile(c.String(FlagInputFile))
	outputFile := getOutputFile(c.String(FlagOutputFilename))
	filter := buildFilterFn(c.String(FlagWorkflowID), c.String(FlagRunID))

	defer inputFile.Close()
	defer outputFile.Close()

	readerCh := make(chan []byte, chanBufferSize)
	writerCh := make(chan *replicator.ReplicationTask, chanBufferSize)
	doneCh := make(chan struct{})

	var skippedCount int
	skipErrMode := c.Bool(FlagSkipErrorMode)
	go startReader(inputFile, readerCh)
	go startParser(readerCh, writerCh, skipErrMode, &skippedCount)
	go startWriter(outputFile, writerCh, filter, doneCh)

	<-doneCh

	if skipErrMode {
		fmt.Printf("%v messages were skipped due to errors in parsing", skippedCount)
	}
}

func buildFilterFn(workflowID, runID string) filterFn {
	return func(task *replicator.ReplicationTask) bool {
		if len(workflowID) != 0 && *task.HistoryTaskAttributes.WorkflowId != workflowID {
			return false
		}
		if len(runID) != 0 && *task.HistoryTaskAttributes.RunId != runID {
			return false
		}
		return true
	}
}

func getInputFile(inputFile string) *os.File {
	if len(inputFile) == 0 {
		info, err := os.Stdin.Stat()
		if err != nil {
			ErrorAndExit("Failed to stat stdin file handle", err)
		}
		if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
			fmt.Println("Misuse of pipe mode")
			os.Exit(1)
		}
		return os.Stdin
	}
	f, err := os.Open(inputFile)
	if err != nil {
		ErrorAndExit(fmt.Sprintf("Failed to open input file for reading: %v", inputFile), err)
	}
	return f
}

func getOutputFile(outputFile string) *os.File {
	if len(outputFile) == 0 {
		return os.Stdout
	}
	f, err := os.Create(outputFile)
	if err != nil {
		ErrorAndExit("failed to create output file", err)
	}
	return f
}

func startReader(file *os.File, readerCh chan<- []byte) {
	defer close(readerCh)
	reader := bufio.NewReader(file)

	for {
		buf := make([]byte, bufferSize)
		n, err := reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				ErrorAndExit("Failed to read from reader", err)
			} else {
				break
			}

		}
		buf = buf[:n]
		readerCh <- buf
	}
}

func startParser(readerCh <-chan []byte, writerCh chan<- *replicator.ReplicationTask, skipErrors bool, skippedCount *int) {
	defer close(writerCh)

	var buffer []byte
Loop:
	for {
		select {
		case data, ok := <-readerCh:
			if !ok {
				break Loop
			}
			buffer = append(buffer, data...)
			data, nextBuffer := splitBuffer(buffer)
			buffer = nextBuffer
			parse(data, skipErrors, skippedCount, writerCh)
		}
	}
	parse(buffer, skipErrors, skippedCount, writerCh)
}

func startWriter(output *os.File, writerCh <-chan *replicator.ReplicationTask, filter filterFn, doneCh chan struct{}) {
	defer close(doneCh)

Loop:
	for {
		select {
		case task, ok := <-writerCh:
			if !ok {
				break Loop
			}
			if filter(task) {
				jsonStr, err := json.Marshal(task)
				if err != nil {
					fmt.Printf("failed to encode into json, err: %v", err)
				}
				output.WriteString(fmt.Sprintf("%v\n", string(jsonStr)))
			}
		}
	}
}

func splitBuffer(buffer []byte) ([]byte, []byte) {
	matches := r.FindAllIndex(buffer, -1)
	if len(matches) == 0 {
		ErrorAndExit(malformedMessage, errors.New("header not found, did you generate dump with -v"))
	}
	splitIndex := matches[len(matches)-1][0]
	return buffer[:splitIndex], buffer[splitIndex:]
}

func parse(bytes []byte, skipErrors bool, skippedCount *int, writerCh chan<- *replicator.ReplicationTask) {
	messages, skippedGetMsgCount := getMessages(bytes, skipErrors)
	tasks, skippedDeserializeCount := deserializeMessages(messages, skipErrors)
	*skippedCount += skippedGetMsgCount + skippedDeserializeCount
	for _, t := range tasks {
		writerCh <- t
	}
}

func getMessages(data []byte, skipErrors bool) ([][]byte, int) {
	str := string(data)
	messagesWithHeaders := r.Split(str, -1)
	if len(messagesWithHeaders[0]) != 0 {
		ErrorAndExit(malformedMessage, errors.New("got data chunk to handle that does not start with valid header"))
	}
	messagesWithHeaders = messagesWithHeaders[1:]
	var rawMessages [][]byte
	skipped := 0
	for _, m := range messagesWithHeaders {
		if len(m) == 0 {
			ErrorAndExit(malformedMessage, errors.New("got empty message between valid headers"))
		}
		curr := []byte(m)
		messageStart := bytes.Index(curr, []byte{preambleVersion0})
		if messageStart == -1 {
			if !skipErrors {
				ErrorAndExit(malformedMessage, errors.New("failed to find message preamble"))
			} else {
				skipped++
				continue
			}
		}
		rawMessages = append(rawMessages, curr[messageStart:])
	}
	return rawMessages, skipped
}

func deserializeMessages(messages [][]byte, skipErrors bool) ([]*replicator.ReplicationTask, int) {
	var replicationTasks []*replicator.ReplicationTask
	skipped := 0
	for _, m := range messages {
		var task replicator.ReplicationTask
		err := decode(m, &task)
		if err != nil {
			if !skipErrors {
				ErrorAndExit(malformedMessage, err)
			} else {
				skipped++
				continue
			}
		}
		replicationTasks = append(replicationTasks, &task)
	}
	return replicationTasks, skipped
}

func decode(message []byte, val *replicator.ReplicationTask) error {
	reader := bytes.NewReader(message[1:])
	wireVal, err := protocol.Binary.Decode(reader, wire.TStruct)
	if err != nil {
		return err
	}
	return val.FromWire(wireVal)
}

// ClustersConfig describes the kafka clusters
type ClustersConfig struct {
	Clusters map[string]messaging.ClusterConfig
}

// AdminRereplicate parses will re-publish replication tasks to topic
func AdminRereplicate(c *cli.Context) {
	hf := getRequiredOption(c, FlagHostFile)
	in := getRequiredOption(c, FlagInputFile)
	cl := getRequiredOption(c, FlagCluster)
	tp := getRequiredOption(c, FlagTopic)

	// initialize kafka producer
	brokers, err := loadBrokers(hf, cl)
	if err != nil {
		ErrorAndExit("", err)
	}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	sproducer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		ErrorAndExit("", err)
	}
	logger := bark.NewLoggerFromLogrus(&logrus.Logger{
		Formatter: new(logrus.JSONFormatter),
		Level:     logrus.InfoLevel,
	})

	producer := messaging.NewKafkaProducer(tp, sproducer, logger)

	// parse json input as replicaiton tasks
	tasks, err := parseReplicationTask(in)
	if err != nil {
		ErrorAndExit("", err)
	}

	// publish to topic
	for idx, t := range tasks {
		err := producer.Publish(t)
		if err != nil {
			fmt.Printf("cannot publish task %v to topic \n", idx)
			ErrorAndExit("", err)
		} else {
			fmt.Printf("replication task sent: %v firstID %v, nextID %v \n", idx, t.HistoryTaskAttributes.GetFirstEventId(), t.HistoryTaskAttributes.GetNextEventId())
		}
	}

}

func parseReplicationTask(in string) (tasks []*replicator.ReplicationTask, err error) {
	file, err := os.Open(in)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	idx := 0
	for scanner.Scan() {
		idx++
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			fmt.Printf("line %v is empty, skipped\n", idx)
			continue
		}

		t := &replicator.ReplicationTask{}
		err := json.Unmarshal([]byte(line), t)
		if err != nil {
			fmt.Printf("line %v cannot be deserialized to replicaiton task: %v.\n", idx, line)
			return nil, err
		}
		tasks = append(tasks, t)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func loadBrokers(hostFile string, cluster string) (brokers []string, err error) {
	contents, err := ioutil.ReadFile(hostFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load kafka cluster info from %v., error: %v", hostFile, err)
	}
	clustersConfig := ClustersConfig{}
	if err := yaml.Unmarshal(contents, &clustersConfig); err != nil {
		return nil, err
	}
	if len(clustersConfig.Clusters) != 0 {
		config, ok := clustersConfig.Clusters[cluster]
		if ok {
			brs := config.Brokers
			for i, b := range brs {
				if !strings.Contains(b, ":") {
					b += ":9092"
					brs[i] = b
				}
			}
			return brs, nil
		}
	}
	return nil, fmt.Errorf("failed to load broker for cluster %v", cluster)
}
