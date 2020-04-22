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

package queue

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	t "github.com/uber/cadence/common/task"
	"github.com/uber/cadence/service/history/task"
)

type (
	processingQueueSuite struct {
		suite.Suite
		*require.Assertions

		controller *gomock.Controller
	}

	testKey struct {
		ID int
	}
)

func TestProcessingQueueSuite(t *testing.T) {
	s := new(processingQueueSuite)
	suite.Run(t, s)
}

func (s *processingQueueSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
}

func (s *processingQueueSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *processingQueueSuite) TestNewProcessingQueue_NoOutstandingTasks() {
	level := 0
	ackLevel := &testKey{ID: 1}
	maxLevel := &testKey{ID: 10}

	queue := NewProcessingQueue(
		NewProcessingQueueState(
			level,
			ackLevel,
			maxLevel,
			DomainFilter{ReverseMatch: true},
		),
		nil,
		nil,
	)
	s.Equal(level, queue.State().Level())
	s.Equal(ackLevel, queue.State().AckLevel())
	s.Equal(ackLevel, queue.State().ReadLevel())
	s.Equal(maxLevel, queue.State().MaxLevel())
}

func (s *processingQueueSuite) TestNewProcessingQueue_WithOutstandingTasks() {
	ackLevel := &testKey{ID: 1}
	maxLevel := &testKey{ID: 10}

	taskKeys := []task.Key{
		&testKey{ID: 2},
		&testKey{ID: 3},
		&testKey{ID: 10},
	}
	taskStates := []t.State{
		t.TaskStateAcked,
		t.TaskStateAcked,
		t.TaskStatePending,
	}
	tasks := make(map[task.Key]task.Task)
	for i, key := range taskKeys {
		task := task.NewMockTask(s.controller)
		task.EXPECT().State().Return(taskStates[i]).MaxTimes(1)
		tasks[key] = task
	}

	queue := newProcessingQueue(
		NewProcessingQueueState(
			0,
			ackLevel,
			maxLevel,
			DomainFilter{ReverseMatch: true},
		),
		tasks,
		nil,
		nil,
	)
	// make sure ack level is updated
	s.Equal(&testKey{ID: 3}, queue.State().AckLevel())
	// make sure read level is updated
	s.Equal(&testKey{ID: 10}, queue.State().ReadLevel())
	s.Equal(maxLevel, queue.State().MaxLevel())
}

func (s *processingQueueSuite) TestAddTasks() {
	ackLevel := &testKey{ID: 1}
	maxLevel := &testKey{ID: 10}

	taskKeys := []task.Key{
		&testKey{ID: 2},
		&testKey{ID: 3},
		&testKey{ID: 5},
		&testKey{ID: 10},
	}
	tasks := make(map[task.Key]task.Task)
	for _, key := range taskKeys {
		tasks[key] = task.NewMockTask(s.controller)
	}

	queue := s.newTestProcessingQueue(
		0,
		ackLevel,
		ackLevel,
		maxLevel,
		DomainFilter{ReverseMatch: true},
		make(map[task.Key]task.Task),
	)

	queue.AddTasks(tasks)
	s.Len(queue.outstandingTasks, len(taskKeys))
	s.Equal(&testKey{ID: 10}, queue.state.readLevel)

	// add the same set of tasks again, should have no effect
	queue.AddTasks(tasks)
	s.Len(queue.outstandingTasks, len(taskKeys))
	s.Equal(&testKey{ID: 10}, queue.state.readLevel)
}

func (s *processingQueueSuite) TestUpdateAckLevel() {
	ackLevel := &testKey{ID: 1}
	maxLevel := &testKey{ID: 10}

	taskKeys := []task.Key{
		&testKey{ID: 2},
		&testKey{ID: 3},
		&testKey{ID: 5},
		&testKey{ID: 8},
		&testKey{ID: 10},
	}
	taskStates := []t.State{
		t.TaskStateAcked,
		t.TaskStateAcked,
		t.TaskStateNacked,
		t.TaskStateAcked,
		t.TaskStatePending,
	}
	tasks := make(map[task.Key]task.Task)
	for i, key := range taskKeys {
		task := task.NewMockTask(s.controller)
		task.EXPECT().State().Return(taskStates[i]).MaxTimes(1)
		tasks[key] = task
	}

	queue := s.newTestProcessingQueue(
		0,
		ackLevel,
		ackLevel,
		maxLevel,
		DomainFilter{ReverseMatch: true},
		tasks,
	)

	queue.UpdateAckLevel()
	s.Equal(&testKey{ID: 3}, queue.state.ackLevel)
}

func (s *processingQueueSuite) newTestProcessingQueue(
	level int,
	ackLevel task.Key,
	readLevel task.Key,
	maxLevel task.Key,
	domainFilter DomainFilter,
	outstandingTasks map[task.Key]task.Task,
) *processingQueueImpl {
	return &processingQueueImpl{
		state: &processingQueueStateImpl{
			level:        level,
			ackLevel:     ackLevel,
			readLevel:    readLevel,
			maxLevel:     maxLevel,
			domainFilter: domainFilter,
		},
		outstandingTasks: outstandingTasks,
	}
}

func (k *testKey) Less(key task.Key) bool {
	return k.ID < key.(*testKey).ID
}
