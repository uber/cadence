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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service"
)

type (
	timerQueueTaskProcessorSuite struct {
		clusterName     string
		logger          log.Logger
		mockService     service.Service
		mockShard       ShardContext
		mockMetadataMgr *mocks.MetadataManager
		mockProcessor   *MockTimerProcessor
		mockQueueAckMgr *MockTimerQueueAckMgr

		scope            int
		notificationChan chan struct{}

		suite.Suite
		timerQueueTaskProcessor *timerQueueTaskProcessor
	}
)

func TestTimerQueueTaskProcessorSuite(t *testing.T) {
	s := new(timerQueueTaskProcessorSuite)
	suite.Run(t, s)
}

func (s *timerQueueTaskProcessorSuite) SetupSuite() {

}

func (s *timerQueueTaskProcessorSuite) TearDownSuite() {

}

func (s *timerQueueTaskProcessorSuite) SetupTest() {
	shardID := 0
	metricsClient := metrics.NewClient(tally.NoopScope, metrics.History)
	s.clusterName = cluster.TestAlternativeClusterName
	s.logger = loggerimpl.NewDevelopmentForTest(s.Suite)
	s.mockProcessor = &MockTimerProcessor{}
	s.mockQueueAckMgr = &MockTimerQueueAckMgr{}
	s.mockMetadataMgr = &mocks.MetadataManager{}
	s.mockService = service.NewTestService(nil, nil, metricsClient, nil, nil, nil)
	s.mockShard = &shardContextImpl{
		service:                   s.mockService,
		shardInfo:                 &persistence.ShardInfo{ShardID: shardID, RangeID: 1, TransferAckLevel: 0},
		transferSequenceNumber:    1,
		maxTransferSequenceNumber: 100000,
		closeCh:                   make(chan int, 100),
		config:                    NewDynamicConfigForTest(),
		logger:                    s.logger,
		domainCache:               cache.NewDomainCache(s.mockMetadataMgr, nil, metricsClient, s.logger),
		metricsClient:             metricsClient,
		standbyClusterCurrentTime: make(map[string]time.Time),
		timeSource:                clock.NewRealTimeSource(),
	}

	s.scope = 0
	s.notificationChan = make(chan struct{})
	h := &historyEngineImpl{
		shard:         s.mockShard,
		logger:        s.logger,
		metricsClient: metricsClient,
	}
	s.timerQueueTaskProcessor = newTimerQueueTaskProcessor(s.mockShard, h, s.logger)
}

func (s *timerQueueTaskProcessorSuite) TearDownTest() {
	s.mockMetadataMgr.AssertExpectations(s.T())
	s.mockProcessor.AssertExpectations(s.T())
	s.mockQueueAckMgr.AssertExpectations(s.T())
}

func (s *timerQueueTaskProcessorSuite) TestProcessTaskAndAck_ShutDown() {
	close(s.timerQueueTaskProcessor.shutdownCh)
	s.timerQueueTaskProcessor.processTaskAndAck(
		s.notificationChan,
		&timerTask{
			processor: s.mockProcessor,
			task:      &persistence.TimerTaskInfo{},
		},
	)
}

func (s *timerQueueTaskProcessorSuite) TestProcessTaskAndAck_DomainErrRetry_ProcessNoErr() {
	task := &persistence.TimerTaskInfo{TaskID: 12345, VisibilityTimestamp: time.Now()}
	var taskFilterErr timerTaskFilter = func(timer *persistence.TimerTaskInfo) (bool, error) {
		return false, errors.New("some random error")
	}
	var taskFilter timerTaskFilter = func(timer *persistence.TimerTaskInfo) (bool, error) {
		return true, nil
	}
	s.mockProcessor.On("getTaskFilter").Return(taskFilterErr).Once()
	s.mockProcessor.On("getTaskFilter").Return(taskFilter).Once()
	s.mockProcessor.On("process", task, true).Return(s.scope, nil).Once()
	s.mockProcessor.On("complete", task).Once()
	s.timerQueueTaskProcessor.processTaskAndAck(
		s.notificationChan,
		&timerTask{
			processor: s.mockProcessor,
			task:      task,
		},
	)
}

func (s *timerQueueTaskProcessorSuite) TestProcessTaskAndAck_DomainFalse_ProcessNoErr() {
	task := &persistence.TimerTaskInfo{TaskID: 12345, VisibilityTimestamp: time.Now()}
	var taskFilter timerTaskFilter = func(timer *persistence.TimerTaskInfo) (bool, error) {
		return false, nil
	}
	s.mockProcessor.On("getTaskFilter").Return(taskFilter).Once()
	s.mockProcessor.On("process", task, false).Return(s.scope, nil).Once()
	s.mockProcessor.On("complete", task).Once()
	s.timerQueueTaskProcessor.processTaskAndAck(
		s.notificationChan,
		&timerTask{
			processor: s.mockProcessor,
			task:      task,
		},
	)
}

func (s *timerQueueTaskProcessorSuite) TestProcessTaskAndAck_DomainTrue_ProcessNoErr() {
	task := &persistence.TimerTaskInfo{TaskID: 12345, VisibilityTimestamp: time.Now()}
	var taskFilter timerTaskFilter = func(timer *persistence.TimerTaskInfo) (bool, error) {
		return true, nil
	}
	s.mockProcessor.On("getTaskFilter").Return(taskFilter).Once()
	s.mockProcessor.On("process", task, true).Return(s.scope, nil).Once()
	s.mockProcessor.On("complete", task).Once()
	s.timerQueueTaskProcessor.processTaskAndAck(
		s.notificationChan,
		&timerTask{
			processor: s.mockProcessor,
			task:      task,
		},
	)
}

func (s *timerQueueTaskProcessorSuite) TestProcessTaskAndAck_DomainTrue_ProcessErrNoErr() {
	err := errors.New("some random err")
	task := &persistence.TimerTaskInfo{TaskID: 12345, VisibilityTimestamp: time.Now()}
	var taskFilter timerTaskFilter = func(timer *persistence.TimerTaskInfo) (bool, error) {
		return true, nil
	}
	s.mockProcessor.On("getTaskFilter").Return(taskFilter).Once()
	s.mockProcessor.On("process", task, true).Return(s.scope, err).Once()
	s.mockProcessor.On("process", task, true).Return(s.scope, nil).Once()
	s.mockProcessor.On("complete", task).Once()
	s.timerQueueTaskProcessor.processTaskAndAck(
		s.notificationChan,
		&timerTask{
			processor: s.mockProcessor,
			task:      task,
		},
	)
}

func (s *timerQueueTaskProcessorSuite) TestHandleTaskError_EntityNotExists() {
	err := &workflow.EntityNotExistsError{}
	s.Nil(s.timerQueueTaskProcessor.handleTaskError(s.scope, time.Now(), s.notificationChan, err, s.logger))
}

func (s *timerQueueTaskProcessorSuite) TestHandleTaskError_ErrTaskRetry() {
	err := ErrTaskRetry
	delay := time.Second

	startTime := time.Now()
	go func() {
		time.Sleep(delay)
		s.notificationChan <- struct{}{}
	}()

	err = s.timerQueueTaskProcessor.handleTaskError(s.scope, time.Now(), s.notificationChan, err, s.logger)
	duration := time.Since(startTime)
	s.True(duration >= delay)
	s.Equal(ErrTaskRetry, err)
}

func (s *timerQueueTaskProcessorSuite) TestHandleTaskError_ErrTaskDiscarded() {
	err := ErrTaskDiscarded
	s.Nil(s.timerQueueTaskProcessor.handleTaskError(s.scope, time.Now(), s.notificationChan, err, s.logger))
}

func (s *timerQueueTaskProcessorSuite) TestHandleTaskError_DomainNotActiveError() {
	err := &workflow.DomainNotActiveError{}

	startTime := time.Now().Add(-cache.DomainCacheRefreshInterval * time.Duration(2))
	s.Nil(s.timerQueueTaskProcessor.handleTaskError(s.scope, startTime, s.notificationChan, err, s.logger))

	startTime = time.Now()
	s.Equal(err, s.timerQueueTaskProcessor.handleTaskError(s.scope, startTime, s.notificationChan, err, s.logger))
}

func (s *timerQueueTaskProcessorSuite) TestHandleTaskError_CurrentWorkflowConditionFailedError() {
	err := &persistence.CurrentWorkflowConditionFailedError{}
	s.Nil(s.timerQueueTaskProcessor.handleTaskError(s.scope, time.Now(), s.notificationChan, err, s.logger))
}

func (s *timerQueueTaskProcessorSuite) TestHandleTaskError_RandomErr() {
	err := errors.New("random error")
	s.Equal(err, s.timerQueueTaskProcessor.handleTaskError(s.scope, time.Now(), s.notificationChan, err, s.logger))
}
