// Copyright (c) 2020 Uber Technologies, Inc.
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

package task

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/metrics"
)

type (
	priorityTaskProcessorSuite struct {
		*require.Assertions
		suite.Suite

		controller    *gomock.Controller
		mockScheduler *MockPriorityScheduler

		processor *priorityTaskProcessorImpl
	}
)

const (
	testWorkerCount = 3
)

var (
	testRetryableError    = errors.New("retryable error")
	testNonRetryableError = errors.New("non-retryable error")
)

func TestPriorityTaskProcessorSuite(t *testing.T) {
	s := new(priorityTaskProcessorSuite)
	suite.Run(t, s)
}

func (s *priorityTaskProcessorSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockScheduler = NewMockPriorityScheduler(s.controller)

	s.processor = NewPriorityTaskProcessor(
		s.mockScheduler,
		loggerimpl.NewDevelopmentForTest(s.Suite),
		metrics.NewClient(tally.NoopScope, metrics.Common),
		&PriorityTaskProcessorOptions{
			WorkerCount: testWorkerCount,
			RetryPolicy: backoff.NewExponentialRetryPolicy(time.Millisecond),
		},
	).(*priorityTaskProcessorImpl)
}

func (s *priorityTaskProcessorSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *priorityTaskProcessorSuite) TestSubmit() {
	priority := 0

	mockTask := NewMockTask(s.controller)
	s.mockScheduler.EXPECT().Schedule(mockTask, priority).Return(nil).Times(1)

	err := s.processor.Submit(mockTask, priority)
	s.NoError(err)
}

func (s *priorityTaskProcessorSuite) TestTaskWorker() {
	numTasks := 5

	mockTask := NewMockTask(s.controller)
	mockTask.EXPECT().Execute().Return(nil).Times(numTasks)
	mockTask.EXPECT().Ack().Times(numTasks)

	gomock.InOrder(
		s.mockScheduler.EXPECT().Consume().Return(mockTask).Times(numTasks),
		s.mockScheduler.EXPECT().Consume().Return(nil).Times(testWorkerCount),
	)
	gomock.InOrder(
		s.mockScheduler.EXPECT().Start().Times(1),
		s.mockScheduler.EXPECT().Stop().Times(1),
	)

	s.processor.Start()
	s.processor.Stop()
}

func (s *priorityTaskProcessorSuite) TestExecuteTask_RetryableError() {
	mockTask := NewMockTask(s.controller)
	gomock.InOrder(
		mockTask.EXPECT().Execute().Return(testRetryableError),
		mockTask.EXPECT().HandleErr(testRetryableError).Return(testRetryableError),
		mockTask.EXPECT().RetryErr(testRetryableError).Return(true),
		mockTask.EXPECT().Execute().Return(testRetryableError),
		mockTask.EXPECT().HandleErr(testRetryableError).Return(testRetryableError),
		mockTask.EXPECT().RetryErr(testRetryableError).Return(true),
		mockTask.EXPECT().Execute().Return(nil),
		mockTask.EXPECT().Ack(),
	)

	s.processor.executeTask(mockTask)
}

func (s *priorityTaskProcessorSuite) TestExecuteTask_NonRetryableError() {
	mockTask := NewMockTask(s.controller)
	gomock.InOrder(
		mockTask.EXPECT().Execute().Return(testNonRetryableError),
		mockTask.EXPECT().HandleErr(testNonRetryableError).Return(testNonRetryableError),
		mockTask.EXPECT().RetryErr(testNonRetryableError).Return(false).AnyTimes(),
		mockTask.EXPECT().Nack(),
	)

	s.processor.executeTask(mockTask)
}

func (s *priorityTaskProcessorSuite) TestExecuteTask_ProcessorStopped() {
	mockTask := NewMockTask(s.controller)
	mockTask.EXPECT().Execute().Return(testRetryableError).AnyTimes()
	mockTask.EXPECT().HandleErr(testRetryableError).Return(testRetryableError).AnyTimes()
	mockTask.EXPECT().RetryErr(testRetryableError).Return(true).AnyTimes()

	done := make(chan struct{})
	go func() {
		s.processor.executeTask(mockTask)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	atomic.StoreInt32(&s.processor.status, common.DaemonStatusStopped)
	<-done
}
