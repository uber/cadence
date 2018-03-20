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
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	"github.com/uber-common/bark"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
)

type (
	timerQueueStandbyProcessorSuite struct {
		suite.Suite
		logger                     bark.Logger
		timerQueueAckMgr           *MockTimerQueueAckMgr
		timerQueueStandbyProcessor *timerQueueStandbyProcessor
	}
)

func TestTimerQueueStandbyProcessorSuite(t *testing.T) {
	s := new(timerQueueStandbyProcessorSuite)
	suite.Run(t, s)
}

func (s *timerQueueStandbyProcessorSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}

}

func (s *timerQueueStandbyProcessorSuite) TearDownSuite() {

}

func (s *timerQueueStandbyProcessorSuite) SetupTest() {
	log2 := log.New()
	log2.Level = log.DebugLevel
	s.logger = bark.NewLoggerFromLogrus(log2)
	s.timerQueueAckMgr = &MockTimerQueueAckMgr{}
	s.timerQueueStandbyProcessor = newTimerQueueStandbyProcessor(s.logger, s.timerQueueAckMgr)
}

func (s *timerQueueStandbyProcessorSuite) TearDownTest() {

}

func (s *timerQueueStandbyProcessorSuite) TestWorkflowTimer() {
	domainID := "some random domain ID"
	workflowID := "some random workflow ID"
	runID := "some random run ID"
	timestamp := time.Now()
	taskID := int64(59)
	taskType := persistence.TaskTypeWorkflowTimeout
	timeoutType := 0             // for workflow timeout this is not used
	eventID := int64(28)         // for workflow timeout this is not used
	scheduleAttempt := int64(12) // for workflow timeout this is not used
	identifier := workflowIdentifier{
		domainID:   domainID,
		workflowID: workflowID,
		runID:      runID,
	}

	timers := []*persistence.TimerTaskInfo{
		&persistence.TimerTaskInfo{
			DomainID:            domainID,
			WorkflowID:          workflowID,
			RunID:               runID,
			VisibilityTimestamp: timestamp,
			TaskID:              taskID,
			TaskType:            taskType,
			TimeoutType:         timeoutType,
			EventID:             eventID,
			ScheduleAttempt:     scheduleAttempt,
		},
	}

	workflowEventTypes := []workflow.EventType{
		workflow.EventTypeWorkflowExecutionCompleted,
		workflow.EventTypeWorkflowExecutionFailed,
		workflow.EventTypeWorkflowExecutionTimedOut,
		workflow.EventTypeWorkflowExecutionTerminated,
		workflow.EventTypeWorkflowExecutionCanceled,
	}

	for _, eventType := range workflowEventTypes {
		events := []*workflow.HistoryEvent{
			&workflow.HistoryEvent{
				EventId:   common.Int64Ptr(11),
				Timestamp: common.Int64Ptr(time.Now().Unix()),
				EventType: &eventType,
				// we are not going to use the workflow end event
			},
		}
		s.timerQueueStandbyProcessor.AddTimers(timers)
		s.timerQueueAckMgr.On("completeTimerTask", TimerSequenceID{
			VisibilityTimestamp: timestamp,
			TaskID:              taskID,
		}).Once()
		s.timerQueueStandbyProcessor.CompleteTimers(identifier, events)
	}
}

func (s *timerQueueStandbyProcessorSuite) TestDecisionTimer_StartToCloseOnly() {
	domainID := "some random domain ID"
	workflowID := "some random workflow ID"
	runID := "some random run ID"
	timestamp := time.Now()
	taskType := persistence.TaskTypeDecisionTimeout
	eventID := int64(28)
	scheduleAttempt := int64(12) // for decision timeout this is not used
	identifier := workflowIdentifier{
		domainID:   domainID,
		workflowID: workflowID,
		runID:      runID,
	}

	prepareTimers := func(taskID int64) {
		timers := []*persistence.TimerTaskInfo{
			&persistence.TimerTaskInfo{
				DomainID:            domainID,
				WorkflowID:          workflowID,
				RunID:               runID,
				VisibilityTimestamp: timestamp,
				TaskID:              taskID,
				TaskType:            taskType,
				TimeoutType:         int(workflow.TimeoutTypeStartToClose),
				EventID:             eventID,
				ScheduleAttempt:     scheduleAttempt,
			},
		}
		s.timerQueueStandbyProcessor.AddTimers(timers)
	}

	taskID := int64(60)
	eventType := workflow.EventTypeDecisionTaskCompleted
	events := []*workflow.HistoryEvent{
		&workflow.HistoryEvent{
			EventId:   common.Int64Ptr(11),
			Timestamp: common.Int64Ptr(time.Now().Unix()),
			EventType: &eventType,
			DecisionTaskCompletedEventAttributes: &workflow.DecisionTaskCompletedEventAttributes{
				ScheduledEventId: &eventID,
			},
		},
	}
	prepareTimers(taskID)
	s.timerQueueAckMgr.On("completeTimerTask", TimerSequenceID{
		VisibilityTimestamp: timestamp,
		TaskID:              taskID,
	}).Once()
	s.timerQueueStandbyProcessor.CompleteTimers(identifier, events)

	taskID = int64(61)
	eventType = workflow.EventTypeDecisionTaskFailed
	events = []*workflow.HistoryEvent{
		&workflow.HistoryEvent{
			EventId:   common.Int64Ptr(11),
			Timestamp: common.Int64Ptr(time.Now().Unix()),
			EventType: &eventType,
			DecisionTaskFailedEventAttributes: &workflow.DecisionTaskFailedEventAttributes{
				ScheduledEventId: &eventID,
			},
		},
	}
	prepareTimers(taskID)
	s.timerQueueAckMgr.On("completeTimerTask", TimerSequenceID{
		VisibilityTimestamp: timestamp,
		TaskID:              taskID,
	}).Once()
	s.timerQueueStandbyProcessor.CompleteTimers(identifier, events)

	taskID = int64(62)
	eventType = workflow.EventTypeDecisionTaskTimedOut
	events = []*workflow.HistoryEvent{
		&workflow.HistoryEvent{
			EventId:   common.Int64Ptr(11),
			Timestamp: common.Int64Ptr(time.Now().Unix()),
			EventType: &eventType,
			DecisionTaskTimedOutEventAttributes: &workflow.DecisionTaskTimedOutEventAttributes{
				ScheduledEventId: &eventID,
			},
		},
	}
	prepareTimers(taskID)
	s.timerQueueAckMgr.On("completeTimerTask", TimerSequenceID{
		VisibilityTimestamp: timestamp,
		TaskID:              taskID,
	}).Once()
	s.timerQueueStandbyProcessor.CompleteTimers(identifier, events)
}

func (s *timerQueueStandbyProcessorSuite) TestDecisionTimer_ScheduleToStart() {
	domainID := "some random domain ID"
	workflowID := "some random workflow ID"
	runID := "some random run ID"
	timestamp := time.Now()
	taskType := persistence.TaskTypeDecisionTimeout
	eventID := int64(28)
	scheduleAttempt := int64(12) // for decision timeout this is not used
	identifier := workflowIdentifier{
		domainID:   domainID,
		workflowID: workflowID,
		runID:      runID,
	}

	prepareTimers := func(taskID int64) {
		timers := []*persistence.TimerTaskInfo{
			&persistence.TimerTaskInfo{
				DomainID:            domainID,
				WorkflowID:          workflowID,
				RunID:               runID,
				VisibilityTimestamp: timestamp,
				TaskID:              taskID,
				TaskType:            taskType,
				TimeoutType:         int(workflow.TimeoutTypeScheduleToStart),
				EventID:             eventID,
				ScheduleAttempt:     scheduleAttempt,
			},
		}
		s.timerQueueStandbyProcessor.AddTimers(timers)
	}

	taskID := int64(59)
	eventType := workflow.EventTypeDecisionTaskStarted
	events := []*workflow.HistoryEvent{
		&workflow.HistoryEvent{
			EventId:   common.Int64Ptr(11),
			Timestamp: common.Int64Ptr(time.Now().Unix()),
			EventType: &eventType,
			DecisionTaskStartedEventAttributes: &workflow.DecisionTaskStartedEventAttributes{
				ScheduledEventId: &eventID,
			},
		},
	}
	prepareTimers(taskID)
	s.timerQueueAckMgr.On("completeTimerTask", TimerSequenceID{
		VisibilityTimestamp: timestamp,
		TaskID:              taskID,
	}).Once()
	s.timerQueueStandbyProcessor.CompleteTimers(identifier, events)

	taskID = int64(62)
	eventType = workflow.EventTypeDecisionTaskTimedOut
	events = []*workflow.HistoryEvent{
		&workflow.HistoryEvent{
			EventId:   common.Int64Ptr(11),
			Timestamp: common.Int64Ptr(time.Now().Unix()),
			EventType: &eventType,
			DecisionTaskTimedOutEventAttributes: &workflow.DecisionTaskTimedOutEventAttributes{
				ScheduledEventId: &eventID,
			},
		},
	}
	prepareTimers(taskID)
	s.timerQueueAckMgr.On("completeTimerTask", TimerSequenceID{
		VisibilityTimestamp: timestamp,
		TaskID:              taskID,
	}).Once()
	s.timerQueueStandbyProcessor.CompleteTimers(identifier, events)
}

func (s *timerQueueStandbyProcessorSuite) TestActivityTimers() {
	domainID := "some random domain ID"
	workflowID := "some random workflow ID"
	runID := "some random run ID"
	timestamp := time.Now()
	taskID := int64(59)
	taskType := persistence.TaskTypeActivityTimeout
	timeoutType := 0             // for activity timeout this is not used
	eventID := int64(28)         // for activity timeout this is not used
	scheduleAttempt := int64(12) // for activity timeout this is not used

	timers := []*persistence.TimerTaskInfo{
		&persistence.TimerTaskInfo{
			DomainID:            domainID,
			WorkflowID:          workflowID,
			RunID:               runID,
			VisibilityTimestamp: timestamp,
			TaskID:              taskID,
			TaskType:            taskType,
			TimeoutType:         timeoutType,
			EventID:             eventID,
			ScheduleAttempt:     scheduleAttempt,
		},
	}
	s.timerQueueStandbyProcessor.AddTimers(timers)
	s.timerQueueAckMgr.On("completeTimerTask", TimerSequenceID{
		VisibilityTimestamp: timestamp,
		TaskID:              taskID,
	}).Once()

	timers = []*persistence.TimerTaskInfo{
		&persistence.TimerTaskInfo{
			DomainID:            domainID,
			WorkflowID:          workflowID,
			RunID:               runID,
			VisibilityTimestamp: timestamp.Add(1 * time.Second),
			TaskID:              taskID,
			TaskType:            taskType,
			TimeoutType:         timeoutType,
			EventID:             eventID,
			ScheduleAttempt:     scheduleAttempt,
		},
	}
	s.timerQueueStandbyProcessor.AddTimers(timers)
}

func (s *timerQueueStandbyProcessorSuite) TestUserTimers() {
	domainID := "some random domain ID"
	workflowID := "some random workflow ID"
	runID := "some random run ID"
	timestamp := time.Now()
	taskID := int64(59)
	taskType := persistence.TaskTypeUserTimer
	timeoutType := 0             // for activity timeout this is not used
	eventID := int64(28)         // for activity timeout this is not used
	scheduleAttempt := int64(12) // for activity timeout this is not used

	timers := []*persistence.TimerTaskInfo{
		&persistence.TimerTaskInfo{
			DomainID:            domainID,
			WorkflowID:          workflowID,
			RunID:               runID,
			VisibilityTimestamp: timestamp,
			TaskID:              taskID,
			TaskType:            taskType,
			TimeoutType:         timeoutType,
			EventID:             eventID,
			ScheduleAttempt:     scheduleAttempt,
		},
	}
	s.timerQueueStandbyProcessor.AddTimers(timers)
	s.timerQueueAckMgr.On("completeTimerTask", TimerSequenceID{
		VisibilityTimestamp: timestamp,
		TaskID:              taskID,
	}).Once()

	timers = []*persistence.TimerTaskInfo{
		&persistence.TimerTaskInfo{
			DomainID:            domainID,
			WorkflowID:          workflowID,
			RunID:               runID,
			VisibilityTimestamp: timestamp,
			TaskID:              taskID + 1,
			TaskType:            taskType,
			TimeoutType:         timeoutType,
			EventID:             eventID,
			ScheduleAttempt:     scheduleAttempt,
		},
	}
	s.timerQueueStandbyProcessor.AddTimers(timers)
}
