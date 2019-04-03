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
	"bytes"
	"encoding/binary"
	"strconv"
	"time"

	"github.com/pborman/uuid"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cluster"
)

func (s *integrationSuite) TestArchival() {
	s.Equal(cluster.ArchivalEnabled, s.ClusterMetadata.ArchivalConfig().GetArchivalStatus())

	workflowID := "archival-workflow-id"
	workflowType := "archival-workflow-type"
	taskList := "archival-task-list"
	numActivities := 1
	numRuns := 1
	runID := s.startAndFinishWorkflow(workflowID, workflowType, taskList, s.archivalDomainName, numActivities, numRuns)[0]

	getHistoryReq := &workflow.GetWorkflowExecutionHistoryRequest{
		Domain: common.StringPtr(s.archivalDomainName),
		Execution: &workflow.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
	}
	var getHistoryResp *workflow.GetWorkflowExecutionHistoryResponse
	var err error
	for i := 0; i < 10; i++ {
		getHistoryResp, err = s.engine.GetWorkflowExecutionHistory(createContext(), getHistoryReq)
		if getHistoryResp != nil && getHistoryResp.GetArchived() {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	s.NoError(err)
	s.NotNil(getHistoryResp)
	s.True(getHistoryResp.GetArchived())
}

func (s *integrationSuite) TestArchival_ContinueAsNew() {
	s.Equal(cluster.ArchivalEnabled, s.ClusterMetadata.ArchivalConfig().GetArchivalStatus())

	workflowID := "archival-continueAsNew-workflow-id"
	workflowType := "archival-continueAsNew-workflow-type"
	taskList := "archival-continueAsNew-task-list"
	numActivities := 1
	numRuns := 5
	runIDs := s.startAndFinishWorkflow(workflowID, workflowType, taskList, s.archivalDomainName, numActivities, numRuns)

	getHistory := func(workflowID, runID string) (*workflow.GetWorkflowExecutionHistoryResponse, error) {
		return s.engine.GetWorkflowExecutionHistory(createContext(), &workflow.GetWorkflowExecutionHistoryRequest{
			Domain: common.StringPtr(s.archivalDomainName),
			Execution: &workflow.WorkflowExecution{
				WorkflowId: common.StringPtr(workflowID),
				RunId:      common.StringPtr(runID),
			},
		})
	}

	for _, runID := range runIDs {
		var getHistoryResp *workflow.GetWorkflowExecutionHistoryResponse
		var err error
		for i := 0; i < 10; i++ {
			getHistoryResp, err = getHistory(workflowID, runID)
			if getHistoryResp != nil && getHistoryResp.GetArchived() {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
		s.NoError(err)
		s.NotNil(getHistoryResp)
		s.True(getHistoryResp.GetArchived())
	}
}

func (s *integrationSuite) startAndFinishWorkflow(id string, wt string, tl string, domain string, numActivities int, numRuns int) []string {
	identity := "worker1"
	activityName := "activity_type1"
	workflowType := &workflow.WorkflowType{
		Name: common.StringPtr(wt),
	}
	taskList := &workflow.TaskList{
		Name: common.StringPtr(tl),
	}
	request := &workflow.StartWorkflowExecutionRequest{
		RequestId:                           common.StringPtr(uuid.New()),
		Domain:                              common.StringPtr(domain),
		WorkflowId:                          common.StringPtr(id),
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(100),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		Identity:                            common.StringPtr(identity),
	}
	we, err := s.engine.StartWorkflowExecution(createContext(), request)
	s.Nil(err)
	s.Logger.Infof("StartWorkflowExecution: response: %v \n", *we.RunId)
	var runIDs []string

	workflowComplete := false
	activityCount := int32(numActivities)
	activityCounter := int32(0)
	expectedActivityID := int32(1)
	runCounter := 1

	dtHandler := func(
		execution *workflow.WorkflowExecution,
		wt *workflow.WorkflowType,
		previousStartedEventID int64,
		startedEventID int64,
		history *workflow.History,
	) ([]byte, []*workflow.Decision, error) {
		if activityCounter < activityCount {
			activityCounter++
			buf := new(bytes.Buffer)
			s.Nil(binary.Write(buf, binary.LittleEndian, activityCounter))
			return []byte(strconv.Itoa(int(activityCounter))), []*workflow.Decision{{
				DecisionType: common.DecisionTypePtr(workflow.DecisionTypeScheduleActivityTask),
				ScheduleActivityTaskDecisionAttributes: &workflow.ScheduleActivityTaskDecisionAttributes{
					ActivityId:                    common.StringPtr(strconv.Itoa(int(activityCounter))),
					ActivityType:                  &workflow.ActivityType{Name: common.StringPtr(activityName)},
					TaskList:                      &workflow.TaskList{Name: &tl},
					Input:                         buf.Bytes(),
					ScheduleToCloseTimeoutSeconds: common.Int32Ptr(100),
					ScheduleToStartTimeoutSeconds: common.Int32Ptr(10),
					StartToCloseTimeoutSeconds:    common.Int32Ptr(50),
					HeartbeatTimeoutSeconds:       common.Int32Ptr(5),
				},
			}}, nil
		}
		runIDs = append(runIDs, execution.GetRunId())
		if runCounter < numRuns {
			activityCounter = int32(0)
			expectedActivityID = int32(1)
			runCounter++
			return []byte(strconv.Itoa(int(activityCounter))), []*workflow.Decision{{
				DecisionType: common.DecisionTypePtr(workflow.DecisionTypeContinueAsNewWorkflowExecution),
				ContinueAsNewWorkflowExecutionDecisionAttributes: &workflow.ContinueAsNewWorkflowExecutionDecisionAttributes{
					WorkflowType:                        workflowType,
					TaskList:                            &workflow.TaskList{Name: &tl},
					Input:                               nil,
					ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(100),
					TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
				},
			}}, nil
		}

		workflowComplete = true
		return []byte(strconv.Itoa(int(activityCounter))), []*workflow.Decision{{
			DecisionType: common.DecisionTypePtr(workflow.DecisionTypeCompleteWorkflowExecution),
			CompleteWorkflowExecutionDecisionAttributes: &workflow.CompleteWorkflowExecutionDecisionAttributes{
				Result: []byte("Done."),
			},
		}}, nil
	}

	atHandler := func(
		execution *workflow.WorkflowExecution,
		activityType *workflow.ActivityType,
		activityID string,
		input []byte,
		taskToken []byte,
	) ([]byte, bool, error) {
		s.Equal(id, *execution.WorkflowId)
		s.Equal(activityName, *activityType.Name)
		id, _ := strconv.Atoi(activityID)
		s.Equal(int(expectedActivityID), id)
		buf := bytes.NewReader(input)
		var in int32
		binary.Read(buf, binary.LittleEndian, &in)
		s.Equal(expectedActivityID, in)
		expectedActivityID++
		return []byte("Activity Result."), false, nil
	}

	poller := &TaskPoller{
		Engine:          s.engine,
		Domain:          domain,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		ActivityHandler: atHandler,
		Logger:          s.Logger,
		T:               s.T(),
	}
	for run := 0; run < numRuns; run++ {
		for i := 0; i < numActivities; i++ {
			_, err := poller.PollAndProcessDecisionTask(false, false)
			s.Logger.Infof("PollAndProcessDecisionTask: %v", err)
			s.Nil(err)
			if i%2 == 0 {
				err = poller.PollAndProcessActivityTask(false)
			} else { // just for testing respondActivityTaskCompleteByID
				err = poller.PollAndProcessActivityTaskWithID(false)
			}
			s.Logger.Infof("PollAndProcessActivityTask: %v", err)
			s.Nil(err)
		}

		_, err = poller.PollAndProcessDecisionTask(true, false)
		s.Nil(err)
	}

	s.True(workflowComplete)
	s.Equal(numRuns, len(runIDs))
	return runIDs
}
