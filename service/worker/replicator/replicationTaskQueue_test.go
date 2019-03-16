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

package replicator

import (
	"log"
	"os"
	"testing"

	"github.com/dgryski/go-farm"

	"github.com/uber/cadence/common/definition"

	"github.com/stretchr/testify/suite"
)

type (
	replicationSequentialTaskQueueSuite struct {
		suite.Suite

		queueID definition.WorkflowIdentifier
		queue   *replicationSequentialTaskQueue
	}
)

func TestReplicationSequentialTaskQueueSuite(t *testing.T) {
	s := new(replicationSequentialTaskQueueSuite)
	suite.Run(t, s)
}

func (s *replicationSequentialTaskQueueSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}
}

func (s *replicationSequentialTaskQueueSuite) TearDownSuite() {

}

func (s *replicationSequentialTaskQueueSuite) SetupTest() {
	s.queueID = definition.NewWorkflowIdentifier(
		"some random domain ID",
		"some random workflow ID",
		"some random run ID",
	)
	s.queue = newReplicationSequentialTaskQueue(s.queueID)
}

func (s *replicationSequentialTaskQueueSuite) TearDownTest() {

}

func (s *replicationSequentialTaskQueueSuite) TestQueueID() {
	s.Equal(s.queueID, s.queue.QueueID())
}

func (s *replicationSequentialTaskQueueSuite) TestOfferPollIsEmptySize() {
	taskID := int64(0)

	s.Equal(0, s.queue.Size())
	s.True(s.queue.IsEmpty())

	testTask1 := s.generateActivityTask(taskID)
	taskID++

	s.queue.Offer(testTask1)
	s.Equal(1, s.queue.Size())
	s.False(s.queue.IsEmpty())

	testTask2 := s.generateHistoryTask(taskID)
	taskID++

	s.queue.Offer(testTask2)
	s.Equal(2, s.queue.Size())
	s.False(s.queue.IsEmpty())

	testTask := s.queue.Poll()
	s.Equal(1, s.queue.Size())
	s.False(s.queue.IsEmpty())
	s.Equal(testTask1, testTask)

	testTask3 := s.generateHistoryTask(taskID)
	taskID++

	s.queue.Offer(testTask3)
	s.Equal(2, s.queue.Size())
	s.False(s.queue.IsEmpty())

	testTask = s.queue.Poll()
	s.Equal(1, s.queue.Size())
	s.False(s.queue.IsEmpty())
	s.Equal(testTask2, testTask)

	testTask = s.queue.Poll()
	s.Equal(0, s.queue.Size())
	s.True(s.queue.IsEmpty())
	s.Equal(testTask3, testTask)

	testTask4 := s.generateActivityTask(taskID)
	taskID++

	s.queue.Offer(testTask4)
	s.Equal(1, s.queue.Size())
	s.False(s.queue.IsEmpty())

	testTask = s.queue.Poll()
	s.Equal(0, s.queue.Size())
	s.True(s.queue.IsEmpty())
	s.Equal(testTask4, testTask)
}

func (s *replicationSequentialTaskQueueSuite) TestHashFn() {
	s.Equal(
		farm.Fingerprint32([]byte(s.queueID.WorkflowID)),
		replicationSequentialTaskQueueHashFn(s.queue),
	)
}

func (s *replicationSequentialTaskQueueSuite) TestCompareLess() {
	s.True(replicationSequentialTaskQueueCompareLess(
		s.generateActivityTask(1),
		s.generateActivityTask(2),
	))

	s.True(replicationSequentialTaskQueueCompareLess(
		s.generateActivityTask(1),
		s.generateHistoryTask(2),
	))

	s.True(replicationSequentialTaskQueueCompareLess(
		s.generateHistoryTask(1),
		s.generateActivityTask(2),
	))

	s.True(replicationSequentialTaskQueueCompareLess(
		s.generateHistoryTask(1),
		s.generateHistoryTask(2),
	))

	s.False(replicationSequentialTaskQueueCompareLess(
		s.generateActivityTask(10),
		s.generateActivityTask(2),
	))

	s.False(replicationSequentialTaskQueueCompareLess(
		s.generateActivityTask(10),
		s.generateHistoryTask(2),
	))

	s.False(replicationSequentialTaskQueueCompareLess(
		s.generateHistoryTask(10),
		s.generateActivityTask(2),
	))

	s.False(replicationSequentialTaskQueueCompareLess(
		s.generateHistoryTask(10),
		s.generateHistoryTask(2),
	))
}

func (s *replicationSequentialTaskQueueSuite) generateActivityTask(taskID int64) *activityReplicationTask {
	return &activityReplicationTask{
		workflowReplicationTask: workflowReplicationTask{queueID: s.queueID,
			taskID: taskID,
		},
	}
}

func (s *replicationSequentialTaskQueueSuite) generateHistoryTask(taskID int64) *historyReplicationTask {
	return &historyReplicationTask{
		workflowReplicationTask: workflowReplicationTask{queueID: s.queueID,
			taskID: taskID,
		},
	}
}
