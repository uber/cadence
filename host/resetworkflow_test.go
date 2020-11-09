// Copyright (c) 2017-2020 Uber Technologies Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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

package host

import (
	"bytes"
	"encoding/binary"
	"strconv"

	"github.com/pborman/uuid"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log/tag"
)

func (s *integrationSuite) TestResetWorkflow() {
	id := "integration-reset-workflow-test"
	wt := "integration-reset-workflow-test-type"
	tl := "integration-reset-workflow-test-taskqueue"
	identity := "worker1"

	workflowType := &shared.WorkflowType{Name: common.StringPtr(wt)}

	tasklist := &shared.TaskList{Name: common.StringPtr(tl)}

	// Start workflow execution
	request := &shared.StartWorkflowExecutionRequest{
		RequestId:                           common.StringPtr(uuid.New()),
		Domain:                              common.StringPtr(s.domainName),
		WorkflowId:                          common.StringPtr(id),
		WorkflowType:                        workflowType,
		TaskList:                            tasklist,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(100),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(2),
		Identity:                            common.StringPtr(identity),
	}

	we, err0 := s.engine.StartWorkflowExecution(createContext(), request)
	s.NoError(err0)

	s.Logger.Info("StartWorkflowExecution", tag.WorkflowRunID(we.GetRunId()))

	// workflow logic
	workflowComplete := false
	activityData := int32(1)
	activityCount := 3
	isFirstTaskProcessed := false
	isSecondTaskProcessed := false
	var firstActivityCompletionEvent *shared.HistoryEvent
	wtHandler := func(execution *shared.WorkflowExecution, wt *shared.WorkflowType,
		previousStartedEventID, startedEventID int64, history *shared.History) ([]byte, []*shared.Decision, error) {

		if !isFirstTaskProcessed {
			// Schedule 3 activities on first workflow task
			isFirstTaskProcessed = true
			buf := new(bytes.Buffer)
			s.Nil(binary.Write(buf, binary.LittleEndian, activityData))

			var scheduleActivityCommands []*shared.Decision
			for i := 1; i <= activityCount; i++ {
				scheduleActivityCommands = append(scheduleActivityCommands, &shared.Decision{
					DecisionType: common.DecisionTypePtr(shared.DecisionTypeScheduleActivityTask),
					ScheduleActivityTaskDecisionAttributes: &shared.ScheduleActivityTaskDecisionAttributes{
						ActivityId:                    common.StringPtr(strconv.Itoa(i)),
						ActivityType:                  &shared.ActivityType{Name: common.StringPtr("ResetActivity")},
						TaskList:                      tasklist,
						Input:                         buf.Bytes(),
						ScheduleToCloseTimeoutSeconds: common.Int32Ptr(100),
						ScheduleToStartTimeoutSeconds: common.Int32Ptr(100),
						StartToCloseTimeoutSeconds:    common.Int32Ptr(50),
						HeartbeatTimeoutSeconds:       common.Int32Ptr(5),
					},
				})
			}

			return nil, scheduleActivityCommands, nil
		} else if !isSecondTaskProcessed {
			// Confirm one activity completion on second workflow task
			isSecondTaskProcessed = true
			for _, event := range history.Events[previousStartedEventID:] {
				if event.GetEventType() == shared.EventTypeActivityTaskCompleted {
					firstActivityCompletionEvent = event
					return nil, []*shared.Decision{}, nil
				}
			}
		}

		// Complete workflow after reset
		workflowComplete = true
		return nil, []*shared.Decision{{
			DecisionType: common.DecisionTypePtr(shared.DecisionTypeCompleteWorkflowExecution),
			CompleteWorkflowExecutionDecisionAttributes: &shared.CompleteWorkflowExecutionDecisionAttributes{
				Result: []byte("Done."),
			},
		}}, nil

	}

	// activity handler
	atHandler := func(execution *shared.WorkflowExecution, activityType *shared.ActivityType,
		activityID string, input []byte, taskToken []byte) ([]byte, bool, error) {

		return []byte("Activity Result."), false, nil
	}

	poller := &TaskPoller{
		Engine:          s.engine,
		Domain:          s.domainName,
		TaskList:        tasklist,
		Identity:        identity,
		DecisionHandler: wtHandler,
		ActivityHandler: atHandler,
		Logger:          s.Logger,
		T:               s.T(),
	}

	// Process first workflow decision task to schedule activities
	_, err := poller.PollAndProcessDecisionTask(false, false)
	s.Logger.Info("PollAndProcessWorkflowTask", tag.Error(err))
	s.NoError(err)

	// Process one activity task which also creates second workflow task
	err = poller.PollAndProcessActivityTask(false)
	s.Logger.Info("Poll and process first activity", tag.Error(err))
	s.NoError(err)

	// Process second workflow task which checks activity completion
	_, err = poller.PollAndProcessDecisionTask(false, false)
	s.Logger.Info("Poll and process second workflow task", tag.Error(err))
	s.NoError(err)

	// Find reset point (last completed decision task)
	events := s.getHistory(s.domainName, &shared.WorkflowExecution{
		WorkflowId: common.StringPtr(id),
		RunId:      common.StringPtr(we.GetRunId()),
	})
	var lastDecisionCompleted *shared.HistoryEvent
	for _, event := range events {
		if event.GetEventType() == shared.EventTypeDecisionTaskCompleted {
			lastDecisionCompleted = event
		}
	}

	// FIRST reset: Reset workflow execution, current is open
	resp, err := s.engine.ResetWorkflowExecution(createContext(), &shared.ResetWorkflowExecutionRequest{
		Domain: common.StringPtr(s.domainName),
		WorkflowExecution: &shared.WorkflowExecution{
			WorkflowId: common.StringPtr(id),
			RunId:      we.RunId,
		},
		Reason:                common.StringPtr("reset execution from test"),
		DecisionFinishEventId: common.Int64Ptr(lastDecisionCompleted.GetEventId()),
		RequestId:             common.StringPtr(uuid.New()),
	})
	s.NoError(err)

	err = poller.PollAndProcessActivityTask(false)
	s.Logger.Info("Poll and process second activity", tag.Error(err))
	s.NoError(err)

	err = poller.PollAndProcessActivityTask(false)
	s.Logger.Info("Poll and process third activity", tag.Error(err))
	s.NoError(err)

	s.NotNil(firstActivityCompletionEvent)
	s.False(workflowComplete)

	// get the history of the first run again
	events = s.getHistory(s.domainName, &shared.WorkflowExecution{
		WorkflowId: common.StringPtr(id),
		RunId:      common.StringPtr(we.GetRunId()),
	})
	var lastEvent *shared.HistoryEvent
	for _, event := range events {
		if event.GetEventType() == shared.EventTypeDecisionTaskCompleted {
			lastDecisionCompleted = event
		}
		lastEvent = event
	}
	// assert the first run is closed, terminated by the previous reset
	s.Equal(shared.EventTypeWorkflowExecutionTerminated, lastEvent.GetEventType())
	// SECOND reset: reset the first run again, to exercise the code path of resetting closed workflow
	resp, err = s.engine.ResetWorkflowExecution(createContext(), &shared.ResetWorkflowExecutionRequest{
		Domain: common.StringPtr(s.domainName),
		WorkflowExecution: &shared.WorkflowExecution{
			WorkflowId: common.StringPtr(id),
			RunId:      common.StringPtr(we.GetRunId()),
		},
		Reason:                common.StringPtr("reset execution from test"),
		DecisionFinishEventId: common.Int64Ptr(lastDecisionCompleted.GetEventId()),
		RequestId:             common.StringPtr(uuid.New()),
	})
	s.NoError(err)
	newRunID := resp.GetRunId()

	_, err = poller.PollAndProcessDecisionTask(false, false)
	s.Logger.Info("Poll and process final decision task", tag.Error(err))
	s.NoError(err)
	s.True(workflowComplete)

	// get the history of the newRunID
	events = s.getHistory(s.domainName, &shared.WorkflowExecution{
		WorkflowId: common.StringPtr(id),
		RunId:      common.StringPtr(newRunID),
	})
	for _, event := range events {
		if event.GetEventType() == shared.EventTypeDecisionTaskCompleted {
			lastDecisionCompleted = event
		}
		lastEvent = event
	}
	// assert the new run is closed, completed by decision task
	s.Equal(shared.EventTypeWorkflowExecutionCompleted, lastEvent.GetEventType())

	// THIRD reset: reset the workflow run that is after a reset
	_, err = s.engine.ResetWorkflowExecution(createContext(), &shared.ResetWorkflowExecutionRequest{
		Domain: common.StringPtr(s.domainName),
		WorkflowExecution: &shared.WorkflowExecution{
			WorkflowId: common.StringPtr(id),
			RunId:      common.StringPtr(newRunID),
		},
		Reason:                common.StringPtr("reset execution from test"),
		DecisionFinishEventId: common.Int64Ptr(lastDecisionCompleted.GetEventId()),
		RequestId:             common.StringPtr(uuid.New()),
	})
	s.NoError(err)
}
