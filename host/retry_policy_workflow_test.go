// Copyright (c) 2021 Uber Technologies, Inc.
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
	"strconv"
	"time"

	"github.com/pborman/uuid"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/types"
)

func (s *IntegrationSuite) TestWorkflowRetryPolicyTimeout() {
	id := "integration-workflow-retry-policy-timeout-test"
	wt := "integration-workflow-retry-policy-timeout-test-type"
	tl := "integration-workflow-retry-policy-timeout-test-tasklist"
	identity := "worker1"

	workflowType := &types.WorkflowType{}
	workflowType.Name = wt

	taskList := &types.TaskList{}
	taskList.Name = tl

	retryPolicy := &types.RetryPolicy{
		InitialIntervalInSeconds:    1,
		BackoffCoefficient:          2.0,
		MaximumIntervalInSeconds:    5,
		ExpirationIntervalInSeconds: 5,
		MaximumAttempts:             0,
		NonRetriableErrorReasons:    []string{"bad-error"},
	}
	request := &types.StartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              s.domainName,
		WorkflowID:                          id,
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(10),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(10),
		Identity:                            identity,
		RetryPolicy:                         retryPolicy,
	}
	now := time.Now()
	we, err0 := s.engine.StartWorkflowExecution(createContext(), request)
	s.Nil(err0)

	s.Logger.Info("StartWorkflowExecution", tag.WorkflowRunID(we.RunID))
	time.Sleep(11 * time.Second) // wait 1 second for timeout

	historyResponse, err := s.engine.GetWorkflowExecutionHistory(createContext(), &types.GetWorkflowExecutionHistoryRequest{
		Domain: s.domainName,
		Execution: &types.WorkflowExecution{
			WorkflowID: id,
		},
	})
	s.Nil(err)
	history := historyResponse.History
	s.NotNil(history.Events[0])
	s.Equal(history.Events[0].EventType, types.EventTypeWorkflowExecutionStarted.Ptr())
	expirationTime := history.Events[0].WorkflowExecutionStartedEventAttributes.ExpirationTimestamp
	s.NotNil(expirationTime)
	delta := time.Unix(0, *expirationTime).Sub(now)
	s.True(delta > 5*time.Second)
	s.True(delta < 6*time.Second)

	wfExecution, err := s.engine.DescribeWorkflowExecution(createContext(), &types.DescribeWorkflowExecutionRequest{
		Domain: s.domainName,
		Execution: &types.WorkflowExecution{
			WorkflowID: id,
		},
	})
	s.NotNil(wfExecution)
	delta = time.Unix(0, *wfExecution.WorkflowExecutionInfo.CloseTime).Sub(time.Unix(0, *wfExecution.WorkflowExecutionInfo.StartTime))
	s.True(delta > 9*time.Second)
	s.True(delta < 11*time.Second)
}

func (s *IntegrationSuite) TestWorkflowRetryPolicyContinueAsNewAsRetry() {
	id := "integration-workflow-retry-policy-continue-as-new-retry-test"
	wt := "integration-workflow-retry-policy-continue-as-new-retry-test-type"
	tl := "integration-workflow-retry-policy-continue-as-new-retry-test-tasklist"
	identity := "worker1"

	workflowType := &types.WorkflowType{}
	workflowType.Name = wt

	taskList := &types.TaskList{}
	taskList.Name = tl

	retryPolicy := &types.RetryPolicy{
		InitialIntervalInSeconds:    1,
		BackoffCoefficient:          2.0,
		MaximumIntervalInSeconds:    5,
		ExpirationIntervalInSeconds: 5,
		MaximumAttempts:             0,
		NonRetriableErrorReasons:    []string{"bad-error"},
	}
	request := &types.StartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              s.domainName,
		WorkflowID:                          id,
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(10),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(10),
		Identity:                            identity,
		RetryPolicy:                         retryPolicy,
	}

	we, err0 := s.engine.StartWorkflowExecution(createContext(), request)
	s.Nil(err0)

	s.Logger.Info("StartWorkflowExecution", tag.WorkflowRunID(we.RunID))
	dtHandler := func(execution *types.WorkflowExecution, wt *types.WorkflowType,
		previousStartedEventID, startedEventID int64, history *types.History) ([]byte, []*types.Decision, error) {
		buf := new(bytes.Buffer)
		return []byte(strconv.Itoa(0)), []*types.Decision{{
			DecisionType: types.DecisionTypeContinueAsNewWorkflowExecution.Ptr(),
			ContinueAsNewWorkflowExecutionDecisionAttributes: &types.ContinueAsNewWorkflowExecutionDecisionAttributes{
				WorkflowType:                        workflowType,
				TaskList:                            &types.TaskList{Name: tl},
				Input:                               buf.Bytes(),
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(10),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(10),
				RetryPolicy:                         retryPolicy,
				Initiator:                           types.ContinueAsNewInitiatorRetryPolicy.Ptr(),
			},
		}}, nil
	}

	poller := &TaskPoller{
		Engine:          s.engine,
		Domain:          s.domainName,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		Logger:          s.Logger,
		T:               s.T(),
	}

	_, err := poller.PollAndProcessDecisionTask(false, false)
	s.Logger.Info("PollAndProcessDecisionTask", tag.Error(err))
	s.Nil(err)
	time.Sleep(10 * time.Second) // wait 1 second for timeout

GetHistoryLoop:
	for i := 0; i < 20; i++ {
		historyResponse, err := s.engine.GetWorkflowExecutionHistory(createContext(), &types.GetWorkflowExecutionHistoryRequest{
			Domain: s.domainName,
			Execution: &types.WorkflowExecution{
				WorkflowID: id,
			},
		})
		s.Nil(err)
		history := historyResponse.History

		lastEvent := history.Events[len(history.Events)-1]
		if *lastEvent.EventType != types.EventTypeWorkflowExecutionTimedOut {
			s.Logger.Warn("Execution not timedout yet.")
			time.Sleep(200 * time.Millisecond)
			continue GetHistoryLoop
		}

		timeoutEventAttributes := lastEvent.WorkflowExecutionTimedOutEventAttributes
		s.Equal(types.TimeoutTypeStartToClose, *timeoutEventAttributes.TimeoutType)
		break GetHistoryLoop
	}
	historyResponse, err := s.engine.GetWorkflowExecutionHistory(createContext(), &types.GetWorkflowExecutionHistoryRequest{
		Domain: s.domainName,
		Execution: &types.WorkflowExecution{
			WorkflowID: id,
		},
	})
	s.Nil(err)
	history := historyResponse.History
	s.NotNil(history.Events[0])
	s.Equal(history.Events[0].EventType, types.EventTypeWorkflowExecutionStarted.Ptr())
	numAttempts := history.Events[0].WorkflowExecutionStartedEventAttributes.Attempt
	s.Equal(numAttempts, int32(1))

	wfExecution, err := s.engine.DescribeWorkflowExecution(createContext(), &types.DescribeWorkflowExecutionRequest{
		Domain: s.domainName,
		Execution: &types.WorkflowExecution{
			WorkflowID: id,
		},
	})
	s.NotNil(wfExecution)
	delta := time.Unix(0, *wfExecution.WorkflowExecutionInfo.CloseTime).Sub(time.Unix(0, *wfExecution.WorkflowExecutionInfo.StartTime))
	s.True(delta > 4*time.Second)
	s.True(delta < 6*time.Second)
}

func (s *IntegrationSuite) TestWorkflowRetryPolicyContinueAsNewAsCron() {
	id := "integration-workflow-retry-policy-continue-as-new-cron-test"
	wt := "integration-workflow-retry-policy-continue-as-new-cron-test-type"
	tl := "integration-workflow-retry-policy-continue-as-new-cron-test-tasklist"
	identity := "worker1"

	workflowType := &types.WorkflowType{}
	workflowType.Name = wt

	taskList := &types.TaskList{}
	taskList.Name = tl

	retryPolicy := &types.RetryPolicy{
		InitialIntervalInSeconds:    1,
		BackoffCoefficient:          2.0,
		MaximumIntervalInSeconds:    5,
		ExpirationIntervalInSeconds: 5,
		MaximumAttempts:             0,
		NonRetriableErrorReasons:    []string{"bad-error"},
	}
	request := &types.StartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              s.domainName,
		WorkflowID:                          id,
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(10),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(10),
		Identity:                            identity,
		RetryPolicy:                         retryPolicy,
	}

	we, err0 := s.engine.StartWorkflowExecution(createContext(), request)
	s.Nil(err0)

	s.Logger.Info("StartWorkflowExecution", tag.WorkflowRunID(we.RunID))
	dtHandler := func(execution *types.WorkflowExecution, wt *types.WorkflowType,
		previousStartedEventID, startedEventID int64, history *types.History) ([]byte, []*types.Decision, error) {
		buf := new(bytes.Buffer)
		return []byte(strconv.Itoa(0)), []*types.Decision{{
			DecisionType: types.DecisionTypeContinueAsNewWorkflowExecution.Ptr(),
			ContinueAsNewWorkflowExecutionDecisionAttributes: &types.ContinueAsNewWorkflowExecutionDecisionAttributes{
				WorkflowType:                        workflowType,
				TaskList:                            &types.TaskList{Name: tl},
				Input:                               buf.Bytes(),
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(15),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(15),
				RetryPolicy:                         retryPolicy,
				Initiator:                           types.ContinueAsNewInitiatorCronSchedule.Ptr(),
			},
		}}, nil
	}

	poller := &TaskPoller{
		Engine:          s.engine,
		Domain:          s.domainName,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		Logger:          s.Logger,
		T:               s.T(),
	}

	_, err := poller.PollAndProcessDecisionTask(false, false)
	s.Logger.Info("PollAndProcessDecisionTask", tag.Error(err))
	s.Nil(err)
	time.Sleep(15 * time.Second) // wait 1 second for timeout

GetHistoryLoop:
	for i := 0; i < 20; i++ {
		historyResponse, err := s.engine.GetWorkflowExecutionHistory(createContext(), &types.GetWorkflowExecutionHistoryRequest{
			Domain: s.domainName,
			Execution: &types.WorkflowExecution{
				WorkflowID: id,
			},
		})
		s.Nil(err)
		history := historyResponse.History

		lastEvent := history.Events[len(history.Events)-1]
		if *lastEvent.EventType != types.EventTypeWorkflowExecutionTimedOut {
			s.Logger.Warn("Execution not timedout yet.")
			time.Sleep(200 * time.Millisecond)
			continue GetHistoryLoop
		}

		timeoutEventAttributes := lastEvent.WorkflowExecutionTimedOutEventAttributes
		s.Equal(types.TimeoutTypeStartToClose, *timeoutEventAttributes.TimeoutType)
		break GetHistoryLoop
	}

	wfExecution, err := s.engine.DescribeWorkflowExecution(createContext(), &types.DescribeWorkflowExecutionRequest{
		Domain: s.domainName,
		Execution: &types.WorkflowExecution{
			WorkflowID: id,
		},
	})
	s.NotNil(wfExecution)
	delta := time.Unix(0, *wfExecution.WorkflowExecutionInfo.CloseTime).Sub(time.Unix(0, *wfExecution.WorkflowExecutionInfo.StartTime))
	s.True(delta > 14*time.Second)
	s.True(delta < 16*time.Second)
}
