package executor

import (
	"testing"

	"time"

	"sync/atomic"

	"github.com/stretchr/testify/suite"
)

type (
	ExecutorTestSuite struct {
		suite.Suite
	}
	testTask struct {
		next    TaskStatus
		counter *int64
	}
)

func TestExecutionTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutorTestSuite))
}

func (s *ExecutorTestSuite) TestStartStop() {
	e := NewFixedSizePoolExecutor(4)
	e.Start()
	e.Stop()
}

func (s *ExecutorTestSuite) TestTaskExecution() {
	e := NewFixedSizePoolExecutor(32)
	e.Start()
	var runCounter int64
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			e.Submit(&testTask{TaskStatusDefer, &runCounter})
			continue
		}
		e.Submit(&testTask{TaskStatusDone, &runCounter})
	}
	s.True(s.awaitCompletion(e))
	s.Equal(int64(150), runCounter)
	e.Stop()
}

func (s *ExecutorTestSuite) awaitCompletion(e Executor) bool {
	expiry := time.Now().Add(time.Second * 10)
	for time.Now().Before(expiry) {
		if e.TaskCount() == 0 {
			return true
		}
		time.Sleep(time.Millisecond * 50)
	}
	return false
}

func (tt *testTask) Run() TaskStatus {
	atomic.AddInt64(tt.counter, 1)
	status := tt.next
	if status == TaskStatusDefer {
		tt.next = TaskStatusDone
	}
	return status
}
