// Copyright (c) 2016 Uber Technologies, Inc.
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
	"encoding/json"
	"flag"
	"strconv"
	"testing"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
)

type (
	decisionIntegrationSuite struct {
		// override suite.Suite.Assertions with require.Assertions; this means that s.NotNil(nil) will stop the test,
		// not merely log an error
		*require.Assertions
		IntegrationBase
	}
)

func (s *decisionIntegrationSuite) SetupSuite() {
	s.setupSuite("testdata/integrationtestcluster.yaml")
}

func (s *decisionIntegrationSuite) TearDownSuite() {
	s.tearDownSuite()
}

func (s *decisionIntegrationSuite) SetupTest() {
	// Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
	s.Assertions = require.New(s.T())
}

func TestIntegrationSuite2(t *testing.T) {
	flag.Parse()
	suite.Run(t, new(decisionIntegrationSuite))
}

func (s *decisionIntegrationSuite) TestWorkflowTerminationSignalBeforeRegularDecisionStarted() {
	id := uuid.New()
	wt := "interation-workflow-transient-decision-test-type"
	tl := id
	identity := "worker1"

	workflowType := &workflow.WorkflowType{}
	workflowType.Name = common.StringPtr(wt)

	taskList := &workflow.TaskList{}
	taskList.Name = common.StringPtr(tl)

	request := &workflow.StartWorkflowExecutionRequest{
		RequestId:                           common.StringPtr(uuid.New()),
		Domain:                              common.StringPtr(s.domainName),
		WorkflowId:                          common.StringPtr(id),
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(3),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(10),
		Identity:                            common.StringPtr(identity),
	}

	resp0, err0 := s.engine.StartWorkflowExecution(createContext(), request)
	s.Nil(err0)

	we := &workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(id),
		RunId:      resp0.RunId,
	}

	s.assertLastHistoryEvent(we, 2, workflow.EventTypeDecisionTaskScheduled)

	err0 = s.engine.SignalWorkflowExecution(createContext(), &workflow.SignalWorkflowExecutionRequest{
		Domain:            common.StringPtr(s.domainName),
		WorkflowExecution: we,
		SignalName:        common.StringPtr("sig-for-integ-test"),
		Input:             []byte(""),
		Identity:          common.StringPtr("integ test"),
		RequestId:         common.StringPtr(uuid.New()),
	})
	s.Nil(err0)
	s.assertLastHistoryEvent(we, 3, workflow.EventTypeWorkflowExecutionSignaled)

	// start this transient decision, the attempt should be cleared and it becomes again a regular decision
	resp1, err1 := s.engine.PollForDecisionTask(createContext(), &workflow.PollForDecisionTaskRequest{
		Domain:   common.StringPtr(s.domainName),
		TaskList: taskList,
		Identity: common.StringPtr(identity),
	})
	s.Nil(err1)

	s.Equal(int64(0), resp1.GetAttempt())
	s.assertLastHistoryEvent(we, 4, workflow.EventTypeDecisionTaskStarted)

	// then terminate the worklfow
	err := s.engine.TerminateWorkflowExecution(createContext(), &workflow.TerminateWorkflowExecutionRequest{
		Domain:            common.StringPtr(s.domainName),
		WorkflowExecution: we,
		Reason:            common.StringPtr("test-reason"),
	})
	s.Nil(err)

	expectedHistory := []workflow.EventType{
		workflow.EventTypeWorkflowExecutionStarted,
		workflow.EventTypeDecisionTaskScheduled,
		workflow.EventTypeWorkflowExecutionSignaled,
		workflow.EventTypeDecisionTaskStarted,
		workflow.EventTypeWorkflowExecutionTerminated,
	}
	s.assertHistory(we, expectedHistory)
}

// TODO signals are left in buffer in this case, which will make reset losing signal
func (s *decisionIntegrationSuite) TestWorkflowTerminationSignalAfterRegularDecisionStarted() {
	id := uuid.New()
	wt := "interation-workflow-transient-decision-test-type"
	tl := id
	identity := "worker1"

	workflowType := &workflow.WorkflowType{}
	workflowType.Name = common.StringPtr(wt)

	taskList := &workflow.TaskList{}
	taskList.Name = common.StringPtr(tl)

	request := &workflow.StartWorkflowExecutionRequest{
		RequestId:                           common.StringPtr(uuid.New()),
		Domain:                              common.StringPtr(s.domainName),
		WorkflowId:                          common.StringPtr(id),
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(3),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(10),
		Identity:                            common.StringPtr(identity),
	}

	resp0, err0 := s.engine.StartWorkflowExecution(createContext(), request)
	s.Nil(err0)

	we := &workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(id),
		RunId:      resp0.RunId,
	}

	s.assertLastHistoryEvent(we, 2, workflow.EventTypeDecisionTaskScheduled)

	// start decision to make signals into bufferedEvents
	_, err1 := s.engine.PollForDecisionTask(createContext(), &workflow.PollForDecisionTaskRequest{
		Domain:   common.StringPtr(s.domainName),
		TaskList: taskList,
		Identity: common.StringPtr(identity),
	})
	s.Nil(err1)

	s.assertLastHistoryEvent(we, 3, workflow.EventTypeDecisionTaskStarted)

	// this signal should be buffered
	err0 = s.engine.SignalWorkflowExecution(createContext(), &workflow.SignalWorkflowExecutionRequest{
		Domain:            common.StringPtr(s.domainName),
		WorkflowExecution: we,
		SignalName:        common.StringPtr("sig-for-integ-test"),
		Input:             []byte(""),
		Identity:          common.StringPtr("integ test"),
		RequestId:         common.StringPtr(uuid.New()),
	})
	s.Nil(err0)
	s.assertLastHistoryEvent(we, 3, workflow.EventTypeDecisionTaskStarted)

	// then terminate the worklfow
	err := s.engine.TerminateWorkflowExecution(createContext(), &workflow.TerminateWorkflowExecutionRequest{
		Domain:            common.StringPtr(s.domainName),
		WorkflowExecution: we,
		Reason:            common.StringPtr("test-reason"),
	})
	s.Nil(err)

	expectedHistory := []workflow.EventType{
		workflow.EventTypeWorkflowExecutionStarted,
		workflow.EventTypeDecisionTaskScheduled,
		workflow.EventTypeDecisionTaskStarted,
		workflow.EventTypeWorkflowExecutionTerminated,
	}
	s.assertHistory(we, expectedHistory)
}

func (s *decisionIntegrationSuite) TestWorkflowTerminationSignalAfterRegularDecisionStartedAndFailDecision() {
	id := uuid.New()
	wt := "interation-workflow-transient-decision-test-type"
	tl := id
	identity := "worker1"

	workflowType := &workflow.WorkflowType{}
	workflowType.Name = common.StringPtr(wt)

	taskList := &workflow.TaskList{}
	taskList.Name = common.StringPtr(tl)

	request := &workflow.StartWorkflowExecutionRequest{
		RequestId:                           common.StringPtr(uuid.New()),
		Domain:                              common.StringPtr(s.domainName),
		WorkflowId:                          common.StringPtr(id),
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(3),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(10),
		Identity:                            common.StringPtr(identity),
	}

	resp0, err0 := s.engine.StartWorkflowExecution(createContext(), request)
	s.Nil(err0)

	we := &workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(id),
		RunId:      resp0.RunId,
	}

	s.assertLastHistoryEvent(we, 2, workflow.EventTypeDecisionTaskScheduled)

	cause := workflow.DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure

	// start decision to make signals into bufferedEvents
	resp1, err1 := s.engine.PollForDecisionTask(createContext(), &workflow.PollForDecisionTaskRequest{
		Domain:   common.StringPtr(s.domainName),
		TaskList: taskList,
		Identity: common.StringPtr(identity),
	})
	s.Nil(err1)

	s.assertLastHistoryEvent(we, 3, workflow.EventTypeDecisionTaskStarted)

	// this signal should be buffered
	err0 = s.engine.SignalWorkflowExecution(createContext(), &workflow.SignalWorkflowExecutionRequest{
		Domain:            common.StringPtr(s.domainName),
		WorkflowExecution: we,
		SignalName:        common.StringPtr("sig-for-integ-test"),
		Input:             []byte(""),
		Identity:          common.StringPtr("integ test"),
		RequestId:         common.StringPtr(uuid.New()),
	})
	s.Nil(err0)
	s.assertLastHistoryEvent(we, 3, workflow.EventTypeDecisionTaskStarted)

	// fail this decision to flush buffer, and then another decision will be scheduled
	err2 := s.engine.RespondDecisionTaskFailed(createContext(), &workflow.RespondDecisionTaskFailedRequest{
		TaskToken: resp1.GetTaskToken(),
		Cause:     &cause,
		Identity:  common.StringPtr("integ test"),
	})
	s.Nil(err2)
	s.assertLastHistoryEvent(we, 6, workflow.EventTypeDecisionTaskScheduled)

	// then terminate the worklfow
	err := s.engine.TerminateWorkflowExecution(createContext(), &workflow.TerminateWorkflowExecutionRequest{
		Domain:            common.StringPtr(s.domainName),
		WorkflowExecution: we,
		Reason:            common.StringPtr("test-reason"),
	})
	s.Nil(err)

	expectedHistory := []workflow.EventType{
		workflow.EventTypeWorkflowExecutionStarted,
		workflow.EventTypeDecisionTaskScheduled,
		workflow.EventTypeDecisionTaskStarted,
		workflow.EventTypeDecisionTaskFailed,
		workflow.EventTypeWorkflowExecutionSignaled,
		workflow.EventTypeDecisionTaskScheduled,
		workflow.EventTypeWorkflowExecutionTerminated,
	}
	s.assertHistory(we, expectedHistory)
}

func (s *decisionIntegrationSuite) TestWorkflowTerminationSignalBeforeTransientDecisionStarted() {
	id := uuid.New()
	wt := "interation-workflow-transient-decision-test-type"
	tl := id
	identity := "worker1"

	workflowType := &workflow.WorkflowType{}
	workflowType.Name = common.StringPtr(wt)

	taskList := &workflow.TaskList{}
	taskList.Name = common.StringPtr(tl)

	request := &workflow.StartWorkflowExecutionRequest{
		RequestId:                           common.StringPtr(uuid.New()),
		Domain:                              common.StringPtr(s.domainName),
		WorkflowId:                          common.StringPtr(id),
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(3),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(10),
		Identity:                            common.StringPtr(identity),
	}

	resp0, err0 := s.engine.StartWorkflowExecution(createContext(), request)
	s.Nil(err0)

	we := &workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(id),
		RunId:      resp0.RunId,
	}

	s.assertLastHistoryEvent(we, 2, workflow.EventTypeDecisionTaskScheduled)

	cause := workflow.DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure
	for i := 0; i < 10; i++ {
		resp1, err1 := s.engine.PollForDecisionTask(createContext(), &workflow.PollForDecisionTaskRequest{
			Domain:   common.StringPtr(s.domainName),
			TaskList: taskList,
			Identity: common.StringPtr(identity),
		})
		s.Nil(err1)
		s.Equal(int64(i), resp1.GetAttempt())
		if i == 0 {
			// first time is regular decision
			s.Equal(int64(3), resp1.GetStartedEventId())
		} else {
			// the rest is transient decision
			s.Equal(int64(6), resp1.GetStartedEventId())
		}

		err2 := s.engine.RespondDecisionTaskFailed(createContext(), &workflow.RespondDecisionTaskFailedRequest{
			TaskToken: resp1.GetTaskToken(),
			Cause:     &cause,
			Identity:  common.StringPtr("integ test"),
		})
		s.Nil(err2)
	}

	s.assertLastHistoryEvent(we, 4, workflow.EventTypeDecisionTaskFailed)

	err0 = s.engine.SignalWorkflowExecution(createContext(), &workflow.SignalWorkflowExecutionRequest{
		Domain:            common.StringPtr(s.domainName),
		WorkflowExecution: we,
		SignalName:        common.StringPtr("sig-for-integ-test"),
		Input:             []byte(""),
		Identity:          common.StringPtr("integ test"),
		RequestId:         common.StringPtr(uuid.New()),
	})
	s.Nil(err0)
	s.assertLastHistoryEvent(we, 5, workflow.EventTypeWorkflowExecutionSignaled)

	// start this transient decision, the attempt should be cleared and it becomes again a regular decision
	resp1, err1 := s.engine.PollForDecisionTask(createContext(), &workflow.PollForDecisionTaskRequest{
		Domain:   common.StringPtr(s.domainName),
		TaskList: taskList,
		Identity: common.StringPtr(identity),
	})
	s.Nil(err1)

	s.Equal(int64(0), resp1.GetAttempt())
	s.assertLastHistoryEvent(we, 7, workflow.EventTypeDecisionTaskStarted)

	// then terminate the worklfow
	err := s.engine.TerminateWorkflowExecution(createContext(), &workflow.TerminateWorkflowExecutionRequest{
		Domain:            common.StringPtr(s.domainName),
		WorkflowExecution: we,
		Reason:            common.StringPtr("test-reason"),
	})
	s.Nil(err)

	expectedHistory := []workflow.EventType{
		workflow.EventTypeWorkflowExecutionStarted,
		workflow.EventTypeDecisionTaskScheduled,
		workflow.EventTypeDecisionTaskStarted,
		workflow.EventTypeDecisionTaskFailed,
		workflow.EventTypeWorkflowExecutionSignaled,
		workflow.EventTypeDecisionTaskScheduled,
		workflow.EventTypeDecisionTaskStarted,
		workflow.EventTypeWorkflowExecutionTerminated,
	}
	s.assertHistory(we, expectedHistory)
}

// TODO signals are left in buffer in this case, which will make reset losing signal
func (s *decisionIntegrationSuite) TestWorkflowTerminationSignalAfterTransientDecisionStarted() {
	id := uuid.New()
	wt := "interation-workflow-transient-decision-test-type"
	tl := id
	identity := "worker1"

	workflowType := &workflow.WorkflowType{}
	workflowType.Name = common.StringPtr(wt)

	taskList := &workflow.TaskList{}
	taskList.Name = common.StringPtr(tl)

	request := &workflow.StartWorkflowExecutionRequest{
		RequestId:                           common.StringPtr(uuid.New()),
		Domain:                              common.StringPtr(s.domainName),
		WorkflowId:                          common.StringPtr(id),
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(3),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(10),
		Identity:                            common.StringPtr(identity),
	}

	resp0, err0 := s.engine.StartWorkflowExecution(createContext(), request)
	s.Nil(err0)

	we := &workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(id),
		RunId:      resp0.RunId,
	}

	s.assertLastHistoryEvent(we, 2, workflow.EventTypeDecisionTaskScheduled)

	cause := workflow.DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure
	for i := 0; i < 10; i++ {
		resp1, err1 := s.engine.PollForDecisionTask(createContext(), &workflow.PollForDecisionTaskRequest{
			Domain:   common.StringPtr(s.domainName),
			TaskList: taskList,
			Identity: common.StringPtr(identity),
		})
		s.Nil(err1)
		s.Equal(int64(i), resp1.GetAttempt())
		if i == 0 {
			// first time is regular decision
			s.Equal(int64(3), resp1.GetStartedEventId())
		} else {
			// the rest is transient decision
			s.Equal(int64(6), resp1.GetStartedEventId())
		}

		err2 := s.engine.RespondDecisionTaskFailed(createContext(), &workflow.RespondDecisionTaskFailedRequest{
			TaskToken: resp1.GetTaskToken(),
			Cause:     &cause,
			Identity:  common.StringPtr("integ test"),
		})
		s.Nil(err2)
	}

	s.assertLastHistoryEvent(we, 4, workflow.EventTypeDecisionTaskFailed)

	// start decision to make signals into bufferedEvents
	_, err1 := s.engine.PollForDecisionTask(createContext(), &workflow.PollForDecisionTaskRequest{
		Domain:   common.StringPtr(s.domainName),
		TaskList: taskList,
		Identity: common.StringPtr(identity),
	})
	s.Nil(err1)

	s.assertLastHistoryEvent(we, 4, workflow.EventTypeDecisionTaskFailed)

	// this signal should be buffered
	err0 = s.engine.SignalWorkflowExecution(createContext(), &workflow.SignalWorkflowExecutionRequest{
		Domain:            common.StringPtr(s.domainName),
		WorkflowExecution: we,
		SignalName:        common.StringPtr("sig-for-integ-test"),
		Input:             []byte(""),
		Identity:          common.StringPtr("integ test"),
		RequestId:         common.StringPtr(uuid.New()),
	})
	s.Nil(err0)
	s.assertLastHistoryEvent(we, 4, workflow.EventTypeDecisionTaskFailed)

	// then terminate the worklfow
	err := s.engine.TerminateWorkflowExecution(createContext(), &workflow.TerminateWorkflowExecutionRequest{
		Domain:            common.StringPtr(s.domainName),
		WorkflowExecution: we,
		Reason:            common.StringPtr("test-reason"),
	})
	s.Nil(err)

	expectedHistory := []workflow.EventType{
		workflow.EventTypeWorkflowExecutionStarted,
		workflow.EventTypeDecisionTaskScheduled,
		workflow.EventTypeDecisionTaskStarted,
		workflow.EventTypeDecisionTaskFailed,
		workflow.EventTypeWorkflowExecutionTerminated,
	}
	s.assertHistory(we, expectedHistory)
}

func (s *decisionIntegrationSuite) TestWorkflowTerminationSignalAfterTransientDecisionStartedAndFailDecision() {
	id := uuid.New()
	wt := "interation-workflow-transient-decision-test-type"
	tl := id
	identity := "worker1"

	workflowType := &workflow.WorkflowType{}
	workflowType.Name = common.StringPtr(wt)

	taskList := &workflow.TaskList{}
	taskList.Name = common.StringPtr(tl)

	request := &workflow.StartWorkflowExecutionRequest{
		RequestId:                           common.StringPtr(uuid.New()),
		Domain:                              common.StringPtr(s.domainName),
		WorkflowId:                          common.StringPtr(id),
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(3),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(10),
		Identity:                            common.StringPtr(identity),
	}

	resp0, err0 := s.engine.StartWorkflowExecution(createContext(), request)
	s.Nil(err0)

	we := &workflow.WorkflowExecution{
		WorkflowId: common.StringPtr(id),
		RunId:      resp0.RunId,
	}

	s.assertLastHistoryEvent(we, 2, workflow.EventTypeDecisionTaskScheduled)

	cause := workflow.DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure
	for i := 0; i < 10; i++ {
		resp1, err1 := s.engine.PollForDecisionTask(createContext(), &workflow.PollForDecisionTaskRequest{
			Domain:   common.StringPtr(s.domainName),
			TaskList: taskList,
			Identity: common.StringPtr(identity),
		})
		s.Nil(err1)
		s.Equal(int64(i), resp1.GetAttempt())
		if i == 0 {
			// first time is regular decision
			s.Equal(int64(3), resp1.GetStartedEventId())
		} else {
			// the rest is transient decision
			s.Equal(int64(6), resp1.GetStartedEventId())
		}

		err2 := s.engine.RespondDecisionTaskFailed(createContext(), &workflow.RespondDecisionTaskFailedRequest{
			TaskToken: resp1.GetTaskToken(),
			Cause:     &cause,
			Identity:  common.StringPtr("integ test"),
		})
		s.Nil(err2)
	}

	s.assertLastHistoryEvent(we, 4, workflow.EventTypeDecisionTaskFailed)

	// start decision to make signals into bufferedEvents
	resp1, err1 := s.engine.PollForDecisionTask(createContext(), &workflow.PollForDecisionTaskRequest{
		Domain:   common.StringPtr(s.domainName),
		TaskList: taskList,
		Identity: common.StringPtr(identity),
	})
	s.Nil(err1)

	s.assertLastHistoryEvent(we, 4, workflow.EventTypeDecisionTaskFailed)

	// this signal should be buffered
	err0 = s.engine.SignalWorkflowExecution(createContext(), &workflow.SignalWorkflowExecutionRequest{
		Domain:            common.StringPtr(s.domainName),
		WorkflowExecution: we,
		SignalName:        common.StringPtr("sig-for-integ-test"),
		Input:             []byte(""),
		Identity:          common.StringPtr("integ test"),
		RequestId:         common.StringPtr(uuid.New()),
	})
	s.Nil(err0)
	s.assertLastHistoryEvent(we, 4, workflow.EventTypeDecisionTaskFailed)

	// fail this decision to flush buffer
	err2 := s.engine.RespondDecisionTaskFailed(createContext(), &workflow.RespondDecisionTaskFailedRequest{
		TaskToken: resp1.GetTaskToken(),
		Cause:     &cause,
		Identity:  common.StringPtr("integ test"),
	})
	s.Nil(err2)
	s.assertLastHistoryEvent(we, 6, workflow.EventTypeDecisionTaskScheduled)

	// then terminate the worklfow
	err := s.engine.TerminateWorkflowExecution(createContext(), &workflow.TerminateWorkflowExecutionRequest{
		Domain:            common.StringPtr(s.domainName),
		WorkflowExecution: we,
		Reason:            common.StringPtr("test-reason"),
	})
	s.Nil(err)

	expectedHistory := []workflow.EventType{
		workflow.EventTypeWorkflowExecutionStarted,
		workflow.EventTypeDecisionTaskScheduled,
		workflow.EventTypeDecisionTaskStarted,
		workflow.EventTypeDecisionTaskFailed,
		workflow.EventTypeWorkflowExecutionSignaled,
		workflow.EventTypeDecisionTaskScheduled,
		workflow.EventTypeWorkflowExecutionTerminated,
	}
	s.assertHistory(we, expectedHistory)
}

func (s *decisionIntegrationSuite) assertHistory(we *workflow.WorkflowExecution, expectedHistory []workflow.EventType) {
	historyResponse, err := s.engine.GetWorkflowExecutionHistory(createContext(), &workflow.GetWorkflowExecutionHistoryRequest{
		Domain:    common.StringPtr(s.domainName),
		Execution: we,
	})
	s.Nil(err)
	history := historyResponse.History
	data, err := json.MarshalIndent(history, "", "    ")
	s.Nil(err)
	s.Equal(len(expectedHistory), len(history.Events), string(data))
	for i, e := range history.Events {
		s.Equal(expectedHistory[i], e.GetEventType(), "%v, %v, %v", strconv.Itoa(i), e.GetEventType().String(), string(data))
	}
}

func (s *decisionIntegrationSuite) assertLastHistoryEvent(we *workflow.WorkflowExecution, count int, eventType workflow.EventType) {
	historyResponse, err := s.engine.GetWorkflowExecutionHistory(createContext(), &workflow.GetWorkflowExecutionHistoryRequest{
		Domain:    common.StringPtr(s.domainName),
		Execution: we,
	})
	s.Nil(err)
	history := historyResponse.History
	data, err := json.MarshalIndent(history, "", "    ")
	s.Nil(err)
	s.Equal(count, len(history.Events), string(data))
	s.Equal(eventType, history.Events[len(history.Events)-1].GetEventType(), string(data))
}
