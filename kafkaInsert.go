package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pborman/uuid"
	"github.com/uber/cadence/.gen/go/indexer"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/codec"
	es "github.com/uber/cadence/common/elasticsearch"
)

const maxMessageBytes = 2000000

var (
	brokers      = []string{"kloak280-dca1:9092", "kloak159-dca1:9092", "kloak277-dca1:9092", "kloak276-dca1:9092", "kloak160-dca1:9092", "kloak282-dca1:9092", "kloak158-dca1:9092", "kloak527-dca1:9092", "kloak281-dca1:9092", "kloak157-dca1:9092", "kloak278-dca1:9092", "kloak279-dca1:9092", "kloak161-dca:9092"}
	topic        = "cadence-visibility-dev-dca1a"
	totalMessage = 100000
	concurrency  = 8
)

func main() {
	defer elapsed("main")()

	producer := buildSaramaProducer()
	var wg sync.WaitGroup
	wg.Add(concurrency)
	encoder := codec.NewThriftRWEncoder()
	for i := 0; i < concurrency; i++ {
		go sendMessage(producer, i, &wg, encoder)
	}
	wg.Wait()
}

func sendMessage(producer sarama.SyncProducer, i int, wg *sync.WaitGroup, encoder *codec.ThriftRWEncoder) {
	defer elapsed(strconv.Itoa(i))()
	defer wg.Done()
	for i := 0; i < totalMessage/concurrency; i++ {
		indexMsg := getIndexMsg()
		_, _, err := producer.SendMessage(&sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder(indexMsg.GetWorkflowID()),
			Value: sarama.ByteEncoder(getPayload(encoder, indexMsg)),
		})
		if err != nil {
			panic(fmt.Sprintf("failed to produce message: %v", err))
		}
	}
}

func elapsed(what string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", what, time.Since(start))
	}
}

func constructPayload(size int64) []byte {
	var i int64
	var result []byte
	var b byte
	b = 0x56
	for i = 0; i < size; i++ {
		result = append(result, b)
	}
	return result
}

func getPayload(encoder *codec.ThriftRWEncoder, indexMsg *indexer.Message) []byte {
	payload, err := serializeThrift(encoder, indexMsg)
	if err != nil {
		fmt.Println(err)
	}
	return payload
}

var (
	workflowTypeName  = "code.uber.internal/devexp/cadence-bench/load/basic.stressWorkflowExecute"
	startTimeUnixNano = int64(1547591727348000000)
	endTimeUnixNano   = int64(1547591727787289000)
	closeStatus       = 0
	historyLength     = int64(0)
	domainID          = uuid.New()
	taskID            = int64(123454321)
)

func getIndexMsg() *indexer.Message {
	msgType := indexer.MessageTypeIndex
	fields := map[string]*indexer.Field{
		es.WorkflowType:  {Type: &es.FieldTypeString, StringData: common.StringPtr(workflowTypeName)},
		es.StartTime:     {Type: &es.FieldTypeInt, IntData: common.Int64Ptr(startTimeUnixNano)},
		es.CloseTime:     {Type: &es.FieldTypeInt, IntData: common.Int64Ptr(endTimeUnixNano)},
		es.CloseStatus:   {Type: &es.FieldTypeInt, IntData: common.Int64Ptr(int64(closeStatus))},
		es.HistoryLength: {Type: &es.FieldTypeInt, IntData: common.Int64Ptr(historyLength)},
	}

	msg := &indexer.Message{
		MessageType: &msgType,
		DomainID:    common.StringPtr(domainID),
		WorkflowID:  common.StringPtr(uuid.New()),
		RunID:       common.StringPtr(uuid.New()),
		Version:     common.Int64Ptr(taskID),
		IndexAttributes: &indexer.IndexAttributes{
			Fields: fields,
		},
	}
	return msg
}

func serializeThrift(encoder *codec.ThriftRWEncoder, input codec.ThriftObject) ([]byte, error) {
	payload, err := encoder.Encode(input)
	if err != nil {
		panic("Failed to serialize thrift object")
		return nil, err
	}
	return payload, nil
}

func buildSaramaProducer() sarama.SyncProducer {
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Return.Successes = true
	cfg.Producer.MaxMessageBytes = maxMessageBytes

	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		panic(fmt.Sprintf("got error constructing producer: %v", err))
		return nil
	}
	return producer
}
