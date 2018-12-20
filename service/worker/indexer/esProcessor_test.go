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
	"github.com/stretchr/testify/suite"
	"github.com/uber-common/bark"
	"github.com/uber/cadence/.gen/go/indexer"
	"github.com/uber/cadence/common/collection"
	msgMocks "github.com/uber/cadence/common/messaging/mocks"
	mmocks "github.com/uber/cadence/common/metrics/mocks"
	"github.com/uber/cadence/common/service/dynamicconfig"
	"github.com/uber/cadence/service/worker/indexer/mocks"
	"log"
	"os"
	"testing"
	"time"
)

type esProcessorSuite struct {
	suite.Suite
	esProcessor       *esProcessorImpl
	mockBulkProcessor *mocks.ElasticBulkProcessor
}

func TestESProcessorSuite(t *testing.T) {
	s := new(esProcessorSuite)
	suite.Run(t, s)
}

func (s *esProcessorSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}

	config := &Config{
		IndexerConcurrency:       dynamicconfig.GetIntPropertyFn(32),
		ESProcessorNumOfWorkers:  dynamicconfig.GetIntPropertyFn(1),
		ESProcessorBulkActions:   dynamicconfig.GetIntPropertyFn(10),
		ESProcessorBulkSize:      dynamicconfig.GetIntPropertyFn(2 << 20),
		ESProcessorFlushInterval: dynamicconfig.GetDurationPropertyFn(1 * time.Minute),
	}
	mockBulkProcessor := &mocks.ElasticBulkProcessor{}
	p := &esProcessorImpl{
		config:        config,
		logger:        bark.NewNopLogger(),
		metricsClient: &mmocks.Client{},
	}
	p.mapToKafkaMsg = collection.NewShardedConcurrentTxMap(1024, p.hashFn)
	p.processor = mockBulkProcessor

	s.mockBulkProcessor = mockBulkProcessor
	s.esProcessor = p
}

func (s *esProcessorSuite) TearDownTest() {
	s.mockBulkProcessor.AssertExpectations(s.T())
}

func (s *esProcessorSuite) TestNewESProcessorAndStart() {
	config := &Config{
		ESProcessorNumOfWorkers:  dynamicconfig.GetIntPropertyFn(1),
		ESProcessorBulkActions:   dynamicconfig.GetIntPropertyFn(10),
		ESProcessorBulkSize:      dynamicconfig.GetIntPropertyFn(2 << 20),
		ESProcessorFlushInterval: dynamicconfig.GetDurationPropertyFn(1 * time.Minute),
	}

	esClient := &elastic.Client{}
	p, err := NewESProcessorAndStart(config, esClient, "test-processor", bark.NewNopLogger(), &mmocks.Client{})
	s.NoError(err)
	s.NotNil(p)

	processor, ok := p.(*esProcessorImpl)
	s.True(ok)
	s.NotNil(processor.mapToKafkaMsg)

	p.Stop()
}

func (s *esProcessorSuite) TestStop() {
	s.mockBulkProcessor.On("Stop").Return(nil).Once()
	s.esProcessor.Stop()
	s.Nil(s.esProcessor.mapToKafkaMsg)
}

func (s *esProcessorSuite) TestAdd() {
	request := elastic.NewBulkIndexRequest()
	mockKafkaMsg := &msgMocks.Message{}
	key := "test-key"
	s.Equal(0, s.esProcessor.mapToKafkaMsg.Size())

	s.mockBulkProcessor.On("Add", request).Return().Once()
	s.esProcessor.Add(request, key, mockKafkaMsg)
	s.Equal(1, s.esProcessor.mapToKafkaMsg.Size())

	// test duplicate add
	mockKafkaMsg.On("Ack").Return(nil).Once()
	s.esProcessor.Add(request, key, mockKafkaMsg)
	s.Equal(1, s.esProcessor.mapToKafkaMsg.Size())
	mockKafkaMsg.AssertExpectations(s.T())
}

//func (s *esProcessorSuite) TestBulkAfterAction() {
//
//}
//
//func (s *esProcessorSuite) TestAckKafkaMsg() {
//
//}
//
//func (s *esProcessorSuite) TestHashFn() {
//
//}
//
//func (s *esProcessorSuite) TestGetESRequest() {
//
//}
//
//func (s *esProcessorSuite) TestGetReqType() {
//
//}

func (s *esProcessorSuite) TestConvertESVersionToVisibilityMsgType() {
	typ := convertESVersionToVisibilityMsgType(versionForOpen)
	s.Equal(indexer.VisibilityMsgTypeOpen.String(), typ)
	typ = convertESVersionToVisibilityMsgType(versionForClose)
	s.Equal(indexer.VisibilityMsgTypeClosed.String(), typ)
	typ = convertESVersionToVisibilityMsgType(versionForDelete)
	s.Equal(indexer.VisibilityMsgTypeDelete.String(), typ)
	typ = convertESVersionToVisibilityMsgType(0)
	s.Equal("", typ)
}
