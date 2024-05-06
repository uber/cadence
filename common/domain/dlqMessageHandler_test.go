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

package domain

import (
	"context"
	"errors"
	"fmt"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log/tag"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
)

type (
	dlqMessageHandlerSuite struct {
		suite.Suite

		*require.Assertions
		controller *gomock.Controller

		mockReplicationTaskExecutor *MockReplicationTaskExecutor
		mockReplicationQueue        *MockReplicationQueue
		dlqMessageHandler           *dlqMessageHandlerImpl
	}
)

func TestDLQMessageHandlerSuite(t *testing.T) {
	s := new(dlqMessageHandlerSuite)
	suite.Run(t, s)
}

func (s *dlqMessageHandlerSuite) SetupSuite() {
}

func (s *dlqMessageHandlerSuite) TearDownSuite() {

}

func (s *dlqMessageHandlerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.controller = gomock.NewController(s.T())

	s.mockReplicationTaskExecutor = NewMockReplicationTaskExecutor(s.controller)
	s.mockReplicationQueue = NewMockReplicationQueue(s.controller)

	logger := testlogger.New(s.Suite.T())
	s.dlqMessageHandler = NewDLQMessageHandler(
		s.mockReplicationTaskExecutor,
		s.mockReplicationQueue,
		logger,
		metrics.NewNoopMetricsClient(),
	).(*dlqMessageHandlerImpl)
}

func (s *dlqMessageHandlerSuite) TearDownTest() {
}

func (s *dlqMessageHandlerSuite) TestReadMessages() {
	ackLevel := int64(10)
	lastMessageID := int64(20)
	pageSize := 100
	pageToken := []byte{}

	tasks := []*types.ReplicationTask{
		{
			TaskType:     types.ReplicationTaskTypeDomain.Ptr(),
			SourceTaskID: 1,
		},
	}
	s.mockReplicationQueue.EXPECT().GetDLQAckLevel(gomock.Any()).Return(ackLevel, nil).Times(1)
	s.mockReplicationQueue.EXPECT().GetMessagesFromDLQ(gomock.Any(), ackLevel, lastMessageID, pageSize, pageToken).
		Return(tasks, nil, nil).Times(1)

	resp, token, err := s.dlqMessageHandler.Read(context.Background(), lastMessageID, pageSize, pageToken)

	s.NoError(err)
	s.Equal(tasks, resp)
	s.Nil(token)
}

func (s *dlqMessageHandlerSuite) TestStart() {
	tests := []struct {
		name           string
		initialStatus  int32
		expectedStatus int32
		shouldStart    bool
	}{
		{
			name:           "Should start when initialized",
			initialStatus:  common.DaemonStatusInitialized,
			expectedStatus: common.DaemonStatusStarted,
			shouldStart:    true,
		},
		{
			name:           "Should not start when already started",
			initialStatus:  common.DaemonStatusStarted,
			expectedStatus: common.DaemonStatusStarted,
			shouldStart:    false,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			atomic.StoreInt32(&s.dlqMessageHandler.status, test.initialStatus)
			s.dlqMessageHandler.Start()
			if test.shouldStart {
				s.dlqMessageHandler.logger.Info("Domain DLQ handler started.")
			}
			s.Equal(test.expectedStatus, atomic.LoadInt32(&s.dlqMessageHandler.status))
		})
	}
}

func (s *dlqMessageHandlerSuite) TestStop() {
	tests := []struct {
		name           string
		initialStatus  int32
		expectedStatus int32
		shouldStop     bool
	}{
		{
			name:           "Should stop when started",
			initialStatus:  common.DaemonStatusStarted,
			expectedStatus: common.DaemonStatusStopped,
			shouldStop:     true,
		},
		{
			name:           "Should not stop when not started",
			initialStatus:  common.DaemonStatusInitialized,
			expectedStatus: common.DaemonStatusInitialized,
			shouldStop:     false,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			atomic.StoreInt32(&s.dlqMessageHandler.status, test.initialStatus)
			s.dlqMessageHandler.Stop()
			if test.shouldStop {
				s.dlqMessageHandler.logger.Info("Domain DLQ handler shutting down.")
			}
			s.Equal(test.expectedStatus, atomic.LoadInt32(&s.dlqMessageHandler.status))
		})
	}
}

func (s *dlqMessageHandlerSuite) TestCount() {
	tests := []struct {
		name          string
		forceFetch    bool
		lastCount     int64
		fetchSize     int64
		fetchError    error
		expectedCount int64
		expectedError error
	}{
		{
			name:          "Force fetch with error",
			forceFetch:    true,
			lastCount:     10,
			fetchSize:     0,
			fetchError:    fmt.Errorf("fetch error"),
			expectedCount: 0,
			expectedError: fmt.Errorf("fetch error"),
		},
		{
			name:          "Force fetch with success",
			forceFetch:    true,
			lastCount:     10,
			fetchSize:     20,
			fetchError:    nil,
			expectedCount: 20,
			expectedError: nil,
		},
		{
			name:          "No fetch needed",
			forceFetch:    false,
			lastCount:     30,
			expectedCount: 30,
			expectedError: nil,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.mockReplicationQueue.EXPECT().GetDLQSize(gomock.Any()).Return(test.fetchSize, test.fetchError).MaxTimes(1)
			s.dlqMessageHandler.lastCount = test.lastCount
			count, err := s.dlqMessageHandler.Count(context.Background(), test.forceFetch)
			s.Equal(test.expectedCount, count)
			if test.expectedError != nil {
				s.Equal(test.expectedError, err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *dlqMessageHandlerSuite) TestReadMessages_ThrowErrorOnGetDLQAckLevel() {
	lastMessageID := int64(20)
	pageSize := 100
	pageToken := []byte{}

	tasks := []*types.ReplicationTask{
		{
			TaskType:     types.ReplicationTaskTypeDomain.Ptr(),
			SourceTaskID: 1,
		},
	}
	testError := fmt.Errorf("test")
	s.mockReplicationQueue.EXPECT().GetDLQAckLevel(gomock.Any()).Return(int64(-1), testError).Times(1)
	s.mockReplicationQueue.EXPECT().GetMessagesFromDLQ(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(tasks, nil, nil).Times(0)

	_, _, err := s.dlqMessageHandler.Read(context.Background(), lastMessageID, pageSize, pageToken)

	s.Equal(testError, err)
}

func (s *dlqMessageHandlerSuite) TestReadMessages_ThrowErrorOnReadMessages() {
	ackLevel := int64(10)
	lastMessageID := int64(20)
	pageSize := 100
	pageToken := []byte{}

	testError := fmt.Errorf("test")
	s.mockReplicationQueue.EXPECT().GetDLQAckLevel(gomock.Any()).Return(ackLevel, nil).Times(1)
	s.mockReplicationQueue.EXPECT().GetMessagesFromDLQ(gomock.Any(), ackLevel, lastMessageID, pageSize, pageToken).
		Return(nil, nil, testError).Times(1)

	_, _, err := s.dlqMessageHandler.Read(context.Background(), lastMessageID, pageSize, pageToken)

	s.Equal(testError, err)
}

func (s *dlqMessageHandlerSuite) TestPurgeMessages() {
	ackLevel := int64(10)
	lastMessageID := int64(20)

	s.mockReplicationQueue.EXPECT().GetDLQAckLevel(gomock.Any()).Return(ackLevel, nil).Times(1)
	s.mockReplicationQueue.EXPECT().RangeDeleteMessagesFromDLQ(gomock.Any(), ackLevel, lastMessageID).Return(nil).Times(1)
	s.mockReplicationQueue.EXPECT().UpdateDLQAckLevel(gomock.Any(), lastMessageID).Return(nil).Times(1)
	err := s.dlqMessageHandler.Purge(context.Background(), lastMessageID)

	s.NoError(err)
}

func (s *dlqMessageHandlerSuite) TestPurgeMessages_ThrowErrorOnGetDLQAckLevel() {
	lastMessageID := int64(20)
	testError := fmt.Errorf("test")

	s.mockReplicationQueue.EXPECT().GetDLQAckLevel(gomock.Any()).Return(int64(-1), testError).Times(1)
	s.mockReplicationQueue.EXPECT().RangeDeleteMessagesFromDLQ(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
	s.mockReplicationQueue.EXPECT().UpdateDLQAckLevel(gomock.Any(), gomock.Any()).Times(0)
	err := s.dlqMessageHandler.Purge(context.Background(), lastMessageID)

	s.Equal(testError, err)
}

func (s *dlqMessageHandlerSuite) TestPurgeMessages_ThrowErrorOnPurgeMessages() {
	ackLevel := int64(10)
	lastMessageID := int64(20)
	testError := fmt.Errorf("test")

	s.mockReplicationQueue.EXPECT().GetDLQAckLevel(gomock.Any()).Return(ackLevel, nil).Times(1)
	s.mockReplicationQueue.EXPECT().RangeDeleteMessagesFromDLQ(gomock.Any(), ackLevel, lastMessageID).Return(testError).Times(1)
	s.mockReplicationQueue.EXPECT().UpdateDLQAckLevel(gomock.Any(), gomock.Any()).Times(0)
	err := s.dlqMessageHandler.Purge(context.Background(), lastMessageID)

	s.Equal(testError, err)
}

func (s *dlqMessageHandlerSuite) TestMergeMessages() {
	ackLevel := int64(10)
	lastMessageID := int64(20)
	pageSize := 100
	pageToken := []byte{}
	messageID := int64(11)

	domainAttribute := &types.DomainTaskAttributes{
		ID: uuid.New(),
	}

	tasks := []*types.ReplicationTask{
		{
			TaskType:             types.ReplicationTaskTypeDomain.Ptr(),
			SourceTaskID:         messageID,
			DomainTaskAttributes: domainAttribute,
		},
	}
	s.mockReplicationQueue.EXPECT().GetDLQAckLevel(gomock.Any()).Return(ackLevel, nil).Times(1)
	s.mockReplicationQueue.EXPECT().GetMessagesFromDLQ(gomock.Any(), ackLevel, lastMessageID, pageSize, pageToken).
		Return(tasks, nil, nil).Times(1)
	s.mockReplicationTaskExecutor.EXPECT().Execute(domainAttribute).Return(nil).Times(1)
	s.mockReplicationQueue.EXPECT().UpdateDLQAckLevel(gomock.Any(), messageID).Return(nil).Times(1)
	s.mockReplicationQueue.EXPECT().RangeDeleteMessagesFromDLQ(gomock.Any(), ackLevel, messageID).Return(nil).Times(1)

	token, err := s.dlqMessageHandler.Merge(context.Background(), lastMessageID, pageSize, pageToken)
	s.NoError(err)
	s.Nil(token)
}

func (s *dlqMessageHandlerSuite) TestMergeMessages_ThrowErrorOnGetDLQAckLevel() {
	lastMessageID := int64(20)
	pageSize := 100
	pageToken := []byte{}
	messageID := int64(11)
	testError := fmt.Errorf("test")
	domainAttribute := &types.DomainTaskAttributes{
		ID: uuid.New(),
	}

	tasks := []*types.ReplicationTask{
		{
			TaskType:             types.ReplicationTaskTypeDomain.Ptr(),
			SourceTaskID:         messageID,
			DomainTaskAttributes: domainAttribute,
		},
	}
	s.mockReplicationQueue.EXPECT().GetDLQAckLevel(gomock.Any()).Return(int64(-1), testError).Times(1)
	s.mockReplicationQueue.EXPECT().GetMessagesFromDLQ(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(tasks, nil, nil).Times(0)
	s.mockReplicationTaskExecutor.EXPECT().Execute(gomock.Any()).Times(0)
	s.mockReplicationQueue.EXPECT().DeleteMessageFromDLQ(gomock.Any(), gomock.Any()).Times(0)
	s.mockReplicationQueue.EXPECT().UpdateDLQAckLevel(gomock.Any(), gomock.Any()).Times(0)

	token, err := s.dlqMessageHandler.Merge(context.Background(), lastMessageID, pageSize, pageToken)
	s.Equal(testError, err)
	s.Nil(token)
}

func (s *dlqMessageHandlerSuite) TestMergeMessages_ThrowErrorOnGetDLQMessages() {
	ackLevel := int64(10)
	lastMessageID := int64(20)
	pageSize := 100
	pageToken := []byte{}
	testError := fmt.Errorf("test")

	s.mockReplicationQueue.EXPECT().GetDLQAckLevel(gomock.Any()).Return(ackLevel, nil).Times(1)
	s.mockReplicationQueue.EXPECT().GetMessagesFromDLQ(gomock.Any(), ackLevel, lastMessageID, pageSize, pageToken).
		Return(nil, nil, testError).Times(1)
	s.mockReplicationTaskExecutor.EXPECT().Execute(gomock.Any()).Times(0)
	s.mockReplicationQueue.EXPECT().DeleteMessageFromDLQ(gomock.Any(), gomock.Any()).Times(0)
	s.mockReplicationQueue.EXPECT().UpdateDLQAckLevel(gomock.Any(), gomock.Any()).Times(0)

	token, err := s.dlqMessageHandler.Merge(context.Background(), lastMessageID, pageSize, pageToken)
	s.Equal(testError, err)
	s.Nil(token)
}

func (s *dlqMessageHandlerSuite) TestMergeMessages_ThrowErrorOnHandleReceivingTask() {
	ackLevel := int64(10)
	lastMessageID := int64(20)
	pageSize := 100
	pageToken := []byte{}
	messageID1 := int64(11)
	messageID2 := int64(12)
	testError := fmt.Errorf("test")
	domainAttribute1 := &types.DomainTaskAttributes{
		ID: uuid.New(),
	}
	domainAttribute2 := &types.DomainTaskAttributes{
		ID: uuid.New(),
	}
	tasks := []*types.ReplicationTask{
		{
			TaskType:             types.ReplicationTaskTypeDomain.Ptr(),
			SourceTaskID:         messageID1,
			DomainTaskAttributes: domainAttribute1,
		},
		{
			TaskType:             types.ReplicationTaskTypeDomain.Ptr(),
			SourceTaskID:         messageID2,
			DomainTaskAttributes: domainAttribute2,
		},
	}
	s.mockReplicationQueue.EXPECT().GetDLQAckLevel(gomock.Any()).Return(ackLevel, nil).Times(1)
	s.mockReplicationQueue.EXPECT().GetMessagesFromDLQ(gomock.Any(), ackLevel, lastMessageID, pageSize, pageToken).
		Return(tasks, nil, nil).Times(1)
	s.mockReplicationTaskExecutor.EXPECT().Execute(domainAttribute1).Return(nil).Times(1)
	s.mockReplicationTaskExecutor.EXPECT().Execute(domainAttribute2).Return(testError).Times(1)
	s.mockReplicationQueue.EXPECT().DeleteMessageFromDLQ(gomock.Any(), messageID2).Times(0)

	token, err := s.dlqMessageHandler.Merge(context.Background(), lastMessageID, pageSize, pageToken)
	s.Equal(testError, err)
	s.Nil(token)
}

func (s *dlqMessageHandlerSuite) TestMergeMessages_ThrowErrorOnDeleteMessages() {
	ackLevel := int64(10)
	lastMessageID := int64(20)
	pageSize := 100
	pageToken := []byte{}
	messageID1 := int64(11)
	messageID2 := int64(12)
	testError := fmt.Errorf("test")
	domainAttribute1 := &types.DomainTaskAttributes{
		ID: uuid.New(),
	}
	domainAttribute2 := &types.DomainTaskAttributes{
		ID: uuid.New(),
	}
	tasks := []*types.ReplicationTask{
		{
			TaskType:             types.ReplicationTaskTypeDomain.Ptr(),
			SourceTaskID:         messageID1,
			DomainTaskAttributes: domainAttribute1,
		},
		{
			TaskType:             types.ReplicationTaskTypeDomain.Ptr(),
			SourceTaskID:         messageID2,
			DomainTaskAttributes: domainAttribute2,
		},
	}
	s.mockReplicationQueue.EXPECT().GetDLQAckLevel(gomock.Any()).Return(ackLevel, nil).Times(1)
	s.mockReplicationQueue.EXPECT().GetMessagesFromDLQ(gomock.Any(), ackLevel, lastMessageID, pageSize, pageToken).
		Return(tasks, nil, nil).Times(1)
	s.mockReplicationTaskExecutor.EXPECT().Execute(domainAttribute1).Return(nil).Times(1)
	s.mockReplicationTaskExecutor.EXPECT().Execute(domainAttribute2).Return(nil).Times(1)
	s.mockReplicationQueue.EXPECT().RangeDeleteMessagesFromDLQ(gomock.Any(), ackLevel, messageID2).Return(testError).Times(1)

	token, err := s.dlqMessageHandler.Merge(context.Background(), lastMessageID, pageSize, pageToken)
	s.Error(err)
	s.Nil(token)
}

func (s *dlqMessageHandlerSuite) TestMergeMessages_IgnoreErrorOnUpdateDLQAckLevel() {
	ackLevel := int64(10)
	lastMessageID := int64(20)
	pageSize := 100
	pageToken := []byte{}
	messageID := int64(11)
	testError := fmt.Errorf("test")
	domainAttribute := &types.DomainTaskAttributes{
		ID: uuid.New(),
	}

	tasks := []*types.ReplicationTask{
		{
			TaskType:             types.ReplicationTaskTypeDomain.Ptr(),
			SourceTaskID:         messageID,
			DomainTaskAttributes: domainAttribute,
		},
	}
	s.mockReplicationQueue.EXPECT().GetDLQAckLevel(gomock.Any()).Return(ackLevel, nil).Times(1)
	s.mockReplicationQueue.EXPECT().GetMessagesFromDLQ(gomock.Any(), ackLevel, lastMessageID, pageSize, pageToken).
		Return(tasks, nil, nil).Times(1)
	s.mockReplicationTaskExecutor.EXPECT().Execute(domainAttribute).Return(nil).Times(1)
	s.mockReplicationQueue.EXPECT().RangeDeleteMessagesFromDLQ(gomock.Any(), ackLevel, messageID).Return(nil).Times(1)
	s.mockReplicationQueue.EXPECT().UpdateDLQAckLevel(gomock.Any(), messageID).Return(testError).Times(1)

	token, err := s.dlqMessageHandler.Merge(context.Background(), lastMessageID, pageSize, pageToken)
	s.NoError(err)
	s.Nil(token)
}

func (s *dlqMessageHandlerSuite) TestMergeMessages_NonDomainTask() {
	ackLevel := int64(10)
	lastMessageID := int64(20)
	pageSize := 100
	pageToken := []byte{}

	// Create a task that mimics a non-domain replication task by not setting DomainTaskAttributes
	tasks := []*types.ReplicationTask{
		{
			TaskType:     types.ReplicationTaskTypeDomain.Ptr(), // Still set to domain but no attributes
			SourceTaskID: 1,
		},
	}

	s.mockReplicationQueue.EXPECT().GetDLQAckLevel(gomock.Any()).Return(ackLevel, nil).Times(1)
	s.mockReplicationQueue.EXPECT().GetMessagesFromDLQ(gomock.Any(), ackLevel, lastMessageID, pageSize, pageToken).
		Return(tasks, nil, nil).Times(1)

	token, err := s.dlqMessageHandler.Merge(context.Background(), lastMessageID, pageSize, pageToken)

	s.NotNil(err)
	s.IsType(&types.InternalServiceError{}, err)
	s.Equal("Encounter non domain replication task in domain replication queue.", err.Error())
	s.Nil(token)
}

func (s *dlqMessageHandlerSuite) TestEmitDLQSizeMetricsLoop_ErrorHandling() {
	expectedError := fmt.Errorf("error fetching DLQ size")
	s.mockReplicationQueue.EXPECT().GetDLQSize(gomock.Any()).Return(int64(0), expectedError).AnyTimes()

	// Start the metrics loop in a goroutine
	go s.dlqMessageHandler.emitDLQSizeMetricsLoop()

	// Allow some time for the goroutine to run and tick at least once
	time.Sleep(100 * time.Millisecond)

	// Close the done channel to signal the loop to stop
	close(s.dlqMessageHandler.done)

	// Wait a bit to ensure the loop exits
	time.Sleep(100 * time.Millisecond)

	s.dlqMessageHandler.logger.Warn("Failed to get DLQ size.", tag.Error(errors.New("DomainReplicationQueueSizeLimit")))

}
