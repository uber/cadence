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

package history

import (
	"github.com/golang/mock/gomock"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/resource"
)

type MockShardContext struct {
	*shardContextImpl

	resource *resource.Test
}

var _ ShardContext = (*MockShardContext)(nil)

func NewMockShardContext(ctrl *gomock.Controller, shardInfo *persistence.ShardInfo) *MockShardContext {
	mockResource := resource.NewTest(ctrl, metrics.History)
	mockShard := &shardContextImpl{
		Resource:                  mockResource,
		shardInfo:                 shardInfo,
		transferSequenceNumber:    1,
		maxTransferSequenceNumber: 100000,
		executionManager:          mockResource.ExecutionMgr,
		closeCh:                   make(chan int, 100),
		config:                    NewDynamicConfigForTest(),
		logger:                    mockResource.GetLogger(),
		throttledLogger:           mockResource.GetThrottledLogger(),
	}
	mockShard.eventsCache = newEventsCache(mockShard)
	return &MockShardContext{
		resource:         mockResource,
		shardContextImpl: mockShard,
	}
}
