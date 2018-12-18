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

package indexer

import (
	"github.com/olivere/elastic"
	"github.com/uber-common/bark"
	"github.com/uber/cadence/.gen/go/indexer"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/codec"
	"github.com/uber/cadence/common/logging"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/metrics"
	"sync"
	"sync/atomic"
	"time"
)

type indexProcessor struct {
	appName         string
	consumerName    string
	kafkaClient     messaging.Client
	consumer        messaging.Consumer
	esClient        *elastic.Client
	esProcessor     ESProcessor
	esProcessorName string
	config          *Config
	logger          bark.Logger
	metricsClient   metrics.Client
	isStarted       int32
	isStopped       int32
	shutdownWG      sync.WaitGroup
	shutdownCh      chan struct{}
	msgEncoder      codec.BinaryEncoder
}

const (
	esDocIDDelimiter      = "~"
	esDocType             = "_doc"
	esVisibilityIndexName = "test-vis" // TODO: configuration index name
	versionTypeExternal   = "external"
	versionForOpen        = 10
	versionForClose       = 20
	versionForDelete      = 30
)

func newIndexProcessor(appName, consumerName string, kafkaClient messaging.Client, esClient *elastic.Client, esProcessorName string,
	config *Config, logger bark.Logger, metricsClient metrics.Client) *indexProcessor {
	return &indexProcessor{
		appName:         appName,
		consumerName:    consumerName,
		kafkaClient:     kafkaClient,
		esClient:        esClient,
		esProcessorName: esProcessorName,
		config:          config,
		logger: logger.WithFields(bark.Fields{
			logging.TagWorkflowComponent: logging.TagValueIndexerProcessorComponent,
		}),
		metricsClient: metricsClient,
		shutdownCh:    make(chan struct{}),
		msgEncoder:    codec.NewThriftRWEncoder(),
	}
}

func (p *indexProcessor) Start() error {
	if !atomic.CompareAndSwapInt32(&p.isStarted, 0, 1) {
		return nil
	}

	logging.LogIndexProcessorStartingEvent(p.logger)
	consumer, err := p.kafkaClient.NewConsumer(p.appName, p.consumerName, p.config.IndexerConcurrency())
	if err != nil {
		logging.LogIndexProcessorStartFailedEvent(p.logger, err)
		return err
	}

	if err := consumer.Start(); err != nil {
		logging.LogIndexProcessorStartFailedEvent(p.logger, err)
		return err
	}

	esProcessor, err := NewESProcessorAndStart(p.config, p.esClient, p.esProcessorName, p.logger)
	if err != nil {
		logging.LogIndexProcessorStartFailedEvent(p.logger, err)
		return err
	}

	p.consumer = consumer
	p.esProcessor = esProcessor
	p.shutdownWG.Add(1)
	go p.processorPump()

	logging.LogIndexProcessorStartedEvent(p.logger)
	return nil
}

func (p *indexProcessor) Stop() {
	if !atomic.CompareAndSwapInt32(&p.isStopped, 0, 1) {
		return
	}

	logging.LogIndexProcessorShuttingDownEvent(p.logger)
	defer logging.LogIndexProcessorShutDownEvent(p.logger)

	if atomic.LoadInt32(&p.isStarted) == 1 {
		close(p.shutdownCh)
	}

	if success := common.AwaitWaitGroup(&p.shutdownWG, time.Minute); !success {
		logging.LogIndexProcessorShutDownTimedoutEvent(p.logger)
	}
}

func (p *indexProcessor) processorPump() {
	defer p.shutdownWG.Done()

	var workerWG sync.WaitGroup
	for workerID := 0; workerID < p.config.IndexerConcurrency(); workerID++ {
		workerWG.Add(1)
		go p.messageProcessLoop(&workerWG, workerID)
	}

	select {
	case <-p.shutdownCh:
		// Processor is shutting down, close the underlying consumer and esProcessor
		p.consumer.Stop()
		p.esProcessor.Stop()
	}

	p.logger.Info("Index processor pump shutting down.")
	if success := common.AwaitWaitGroup(&workerWG, 10*time.Second); !success {
		p.logger.Warn("Index processor timed out on worker shutdown.")
	}
}

func (p *indexProcessor) messageProcessLoop(workerWG *sync.WaitGroup, workerID int) {
	defer workerWG.Done()

	for {
		select {
		case msg, ok := <-p.consumer.Messages():
			if !ok {
				p.logger.Info("Worker for index processor shutting down.")
				return // channel closed
			}
			logger := p.logger.WithFields(bark.Fields{
				logging.TagPartitionKey: msg.Partition(),
				logging.TagOffset:       msg.Offset(),
				logging.TagAttemptStart: time.Now(),
			})
			p.process(msg, logger) // what if power off here?
		}
	}
}

func (p *indexProcessor) process(msg messaging.Message, logger bark.Logger) (bark.Logger, error) {
	record, err := p.deserialize(msg.Value())
	if err != nil {
		logger.WithFields(bark.Fields{
			logging.TagErr: err,
		}).Error("Failed to deserialize index messages.")

		return logger, err
	}
	switch record.GetMsgType() {
	case indexer.VisibilityMsgTypeOpen:
		docID := record.GetWorkflowID() + esDocIDDelimiter + record.GetRunID()
		req := elastic.NewBulkIndexRequest().
			Index(esVisibilityIndexName).
			Type(esDocType).
			Id(docID).
			VersionType(versionTypeExternal).
			Version(versionForOpen).
			Doc(*record)
		key := docID + indexer.VisibilityMsgTypeOpen.String()
		p.esProcessor.Add(req, key, msg)
	case indexer.VisibilityMsgTypeClosed:
		docID := record.GetWorkflowID() + esDocIDDelimiter + record.GetRunID()
		req := elastic.NewBulkIndexRequest().
			Index(esVisibilityIndexName).
			Type(esDocType).
			Id(docID).
			VersionType(versionTypeExternal).
			Version(versionForClose).
			Doc(*record)
		key := docID + indexer.VisibilityMsgTypeClosed.String()
		p.esProcessor.Add(req, key, msg)
	case indexer.VisibilityMsgTypeDelete:
		//logger.Infof("vance in process, delete record is: %v", record)
	}
	return logger, nil
}

func (p *indexProcessor) deserialize(payload []byte) (*indexer.VisibilityMsg, error) {
	var msg indexer.VisibilityMsg
	if err := p.msgEncoder.Decode(payload, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
