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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/elasticsearch/bulk"
	mocks2 "github.com/uber/cadence/common/elasticsearch/bulk/mocks"
	esMocks "github.com/uber/cadence/common/elasticsearch/mocks"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/metrics"
)

func TestNewDualIndexer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &Config{
		ESProcessorNumOfWorkers:  dynamicconfig.GetIntPropertyFn(1),
		ESProcessorBulkActions:   dynamicconfig.GetIntPropertyFn(10),
		ESProcessorBulkSize:      dynamicconfig.GetIntPropertyFn(2 << 20),
		ESProcessorFlushInterval: dynamicconfig.GetDurationPropertyFn(1 * time.Minute),
	}
	processorName := "test-bulkProcessor"
	mockESClient := &esMocks.GenericClient{}
	mockESClient.On("RunBulkProcessor", mock.Anything, mock.MatchedBy(func(input *bulk.BulkProcessorParameters) bool {
		return true
	})).Return(&mocks2.GenericBulkProcessor{}, nil).Times(2)

	mockMessagingClient := messaging.NewMockClient(ctrl)
	mockMessagingClient.EXPECT().NewConsumer("visibility", "test-bulkProcessor-consumer").Return(nil, nil).Times(1)

	indexer := NewMigrationIndexer(config, mockMessagingClient, mockESClient, mockESClient, processorName, testlogger.New(t), metrics.NewNoopMetricsClient())
	assert.NotNil(t, indexer)
}

func TestNewIndexer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	config := &Config{
		ESProcessorNumOfWorkers:  dynamicconfig.GetIntPropertyFn(1),
		ESProcessorBulkActions:   dynamicconfig.GetIntPropertyFn(10),
		ESProcessorBulkSize:      dynamicconfig.GetIntPropertyFn(2 << 20),
		ESProcessorFlushInterval: dynamicconfig.GetDurationPropertyFn(1 * time.Minute),
	}
	processorName := "test-bulkProcessor"
	mockESClient := &esMocks.GenericClient{}
	mockESClient.On("RunBulkProcessor", mock.Anything, mock.MatchedBy(func(input *bulk.BulkProcessorParameters) bool {
		return true
	})).Return(&mocks2.GenericBulkProcessor{}, nil).Times(2)

	mockMessagingClient := messaging.NewMockClient(ctrl)
	mockMessagingClient.EXPECT().NewConsumer("visibility", "test-bulkProcessor-consumer").Return(nil, nil).Times(1)

	indexer := NewIndexer(config, mockMessagingClient, mockESClient, processorName, testlogger.New(t), metrics.NewNoopMetricsClient())
	assert.NotNil(t, indexer)
}
