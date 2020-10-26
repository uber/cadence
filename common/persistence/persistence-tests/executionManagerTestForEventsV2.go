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

package persistencetests

import (
	"context"
	"os"
	"runtime/debug"
	"testing"
	"time"

	"github.com/uber/cadence/common/checksum"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	gen "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	p "github.com/uber/cadence/common/persistence"
)

type (
	// ExecutionManagerSuiteForEventsV2 contains matching persistence tests
	ExecutionManagerSuiteForEventsV2 struct {
		TestBase
		// override suite.Suite.Assertions with require.Assertions; this means that s.NotNil(nil) will stop the test,
		// not merely log an error
		*require.Assertions
	}
)

func failOnPanic(t *testing.T) {
	r := recover()
	if r != nil {
		t.Errorf("test panicked: %v %s", r, debug.Stack())
		t.FailNow()
	}
}

// SetupSuite implementation
func (s *ExecutionManagerSuiteForEventsV2) SetupSuite() {
	defer failOnPanic(s.T())
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}
}

// TearDownSuite implementation
func (s *ExecutionManagerSuiteForEventsV2) TearDownSuite() {
	defer failOnPanic(s.T())
	s.TearDownWorkflowStore()
}

// SetupTest implementation
func (s *ExecutionManagerSuiteForEventsV2) SetupTest() {
	defer failOnPanic(s.T())
	// Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
	s.Assertions = require.New(s.T())
	s.ClearTasks()
}

func (s *ExecutionManagerSuiteForEventsV2) newRandomChecksum() checksum.Checksum {
	return checksum.Checksum{
		Flavor:  checksum.FlavorIEEECRC32OverThriftBinary,
		Version: 22,
		Value:   []byte(uuid.NewRandom()),
	}
}

func (s *ExecutionManagerSuiteForEventsV2) assertChecksumsEqual(expected checksum.Checksum, actual checksum.Checksum) {
	if !actual.Flavor.IsValid() {
		// not all stores support checksum persistence today
		// if its not supported, assert that everything is zero'd out
		expected = checksum.Checksum{}
	}
	s.EqualValues(expected, actual)
}

// TestWorkflowCreation test
func (s *ExecutionManagerSuiteForEventsV2) TestWorkflowCreation() {
	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	defer failOnPanic(s.T())
	domainID := uuid.New()
	workflowExecution := gen.WorkflowExecution{
		WorkflowId: common.StringPtr("test-eventsv2-workflow"),
		RunId:      common.StringPtr("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
	}

	csum := s.newRandomChecksum()
	decisionScheduleID := int64(2)
	versionHistory := p.NewVersionHistory([]byte{}, []*p.VersionHistoryItem{
		{decisionScheduleID, common.EmptyVersion},
	})
	verisonHistories := p.NewVersionHistories(versionHistory)
	_, err0 := s.ExecutionManager.CreateWorkflowExecution(ctx, &p.CreateWorkflowExecutionRequest{
		NewWorkflowSnapshot: p.WorkflowSnapshot{
			ExecutionInfo: &p.WorkflowExecutionInfo{
				CreateRequestID:             uuid.New(),
				DomainID:                    domainID,
				WorkflowID:                  workflowExecution.GetWorkflowId(),
				RunID:                       workflowExecution.GetRunId(),
				TaskList:                    "taskList",
				WorkflowTypeName:            "wType",
				WorkflowTimeout:             20,
				DecisionStartToCloseTimeout: 13,
				ExecutionContext:            nil,
				State:                       p.WorkflowStateRunning,
				CloseStatus:                 p.WorkflowCloseStatusNone,
				NextEventID:                 3,
				LastProcessedEvent:          0,
				DecisionScheduleID:          decisionScheduleID,
				DecisionStartedID:           common.EmptyEventID,
				DecisionTimeout:             1,
				BranchToken:                 []byte("branchToken1"),
			},
			ExecutionStats: &p.ExecutionStats{},
			TransferTasks: []p.Task{
				&p.DecisionTask{
					TaskID:              s.GetNextSequenceNumber(),
					DomainID:            domainID,
					TaskList:            "taskList",
					ScheduleID:          2,
					VisibilityTimestamp: time.Now(),
				},
			},
			TimerTasks:       nil,
			Checksum:         csum,
			VersionHistories: verisonHistories,
		},
		RangeID: s.ShardInfo.RangeID,
	})

	s.NoError(err0)

	state0, err1 := s.GetWorkflowExecutionInfo(ctx, domainID, workflowExecution)
	s.NoError(err1)
	info0 := state0.ExecutionInfo
	s.NotNil(info0, "Valid Workflow info expected.")
	s.Equal([]byte("branchToken1"), info0.BranchToken)
	s.assertChecksumsEqual(csum, state0.Checksum)

	updatedInfo := copyWorkflowExecutionInfo(info0)
	updatedStats := copyExecutionStats(state0.ExecutionStats)
	updatedInfo.NextEventID = int64(5)
	updatedInfo.LastProcessedEvent = int64(2)
	currentTime := time.Now().UTC()
	timerID := "id_1"
	timerInfos := []*p.TimerInfo{{
		Version:    3345,
		TimerID:    timerID,
		ExpiryTime: currentTime,
		TaskStatus: 2,
		StartedID:  5,
	}}
	updatedInfo.BranchToken = []byte("branchToken2")

	err2 := s.UpdateWorkflowExecution(ctx, updatedInfo, updatedStats, verisonHistories, []int64{int64(4)}, nil, int64(3), nil, nil, nil, timerInfos, nil)
	s.NoError(err2)

	state, err1 := s.GetWorkflowExecutionInfo(ctx, domainID, workflowExecution)
	s.NoError(err1)
	s.NotNil(state, "expected valid state.")
	s.Equal(1, len(state.TimerInfos))
	s.Equal(int64(3345), state.TimerInfos[timerID].Version)
	s.Equal(timerID, state.TimerInfos[timerID].TimerID)
	s.EqualTimesWithPrecision(currentTime, state.TimerInfos[timerID].ExpiryTime, time.Millisecond*500)
	s.Equal(int64(2), state.TimerInfos[timerID].TaskStatus)
	s.Equal(int64(5), state.TimerInfos[timerID].StartedID)
	s.assertChecksumsEqual(testWorkflowChecksum, state.Checksum)

	err2 = s.UpdateWorkflowExecution(ctx, updatedInfo, updatedStats, verisonHistories, nil, nil, int64(5), nil, nil, nil, nil, []string{timerID})
	s.NoError(err2)

	state, err2 = s.GetWorkflowExecutionInfo(ctx, domainID, workflowExecution)
	s.NoError(err2)
	s.NotNil(state, "expected valid state.")
	s.Equal(0, len(state.TimerInfos))
	info1 := state.ExecutionInfo
	s.Equal([]byte("branchToken2"), info1.BranchToken)
	s.assertChecksumsEqual(testWorkflowChecksum, state.Checksum)
}

// TestWorkflowCreationWithVersionHistories test
func (s *ExecutionManagerSuiteForEventsV2) TestWorkflowCreationWithVersionHistories() {
	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	defer failOnPanic(s.T())
	domainID := uuid.New()
	workflowExecution := gen.WorkflowExecution{
		WorkflowId: common.StringPtr("test-eventsv2-workflow-version-history"),
		RunId:      common.StringPtr("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
	}
	versionHistory := p.NewVersionHistory(
		[]byte{1},
		[]*p.VersionHistoryItem{p.NewVersionHistoryItem(1, 0)},
	)
	versionHistories := p.NewVersionHistories(versionHistory)

	csum := s.newRandomChecksum()

	_, err0 := s.ExecutionManager.CreateWorkflowExecution(ctx, &p.CreateWorkflowExecutionRequest{
		RangeID: s.ShardInfo.RangeID,
		NewWorkflowSnapshot: p.WorkflowSnapshot{
			ExecutionInfo: &p.WorkflowExecutionInfo{
				CreateRequestID:             uuid.New(),
				DomainID:                    domainID,
				WorkflowID:                  workflowExecution.GetWorkflowId(),
				RunID:                       workflowExecution.GetRunId(),
				TaskList:                    "taskList",
				WorkflowTypeName:            "wType",
				WorkflowTimeout:             20,
				DecisionStartToCloseTimeout: 13,
				ExecutionContext:            nil,
				State:                       p.WorkflowStateRunning,
				CloseStatus:                 p.WorkflowCloseStatusNone,
				NextEventID:                 common.EmptyEventID,
				LastProcessedEvent:          0,
				DecisionScheduleID:          2,
				DecisionStartedID:           common.EmptyEventID,
				DecisionTimeout:             1,
				BranchToken:                 nil,
			},
			ExecutionStats:   &p.ExecutionStats{},
			VersionHistories: versionHistories,
			TransferTasks: []p.Task{
				&p.DecisionTask{
					TaskID:              s.GetNextSequenceNumber(),
					DomainID:            domainID,
					TaskList:            "taskList",
					ScheduleID:          2,
					VisibilityTimestamp: time.Now(),
				},
			},
			TimerTasks: nil,
			Checksum:   csum,
		},
	})

	s.NoError(err0)

	state0, err1 := s.GetWorkflowExecutionInfo(ctx, domainID, workflowExecution)
	s.NoError(err1)
	info0 := state0.ExecutionInfo
	s.NotNil(info0, "Valid Workflow info expected.")
	s.Equal(versionHistories, state0.VersionHistories)
	s.assertChecksumsEqual(csum, state0.Checksum)

	updatedInfo := copyWorkflowExecutionInfo(info0)
	updatedStats := copyExecutionStats(state0.ExecutionStats)
	updatedInfo.LastProcessedEvent = int64(2)
	currentTime := time.Now().UTC()
	timerID := "id_1"
	timerInfos := []*p.TimerInfo{{
		Version:    3345,
		TimerID:    timerID,
		ExpiryTime: currentTime,
		TaskStatus: 2,
		StartedID:  5,
	}}
	versionHistory, err := versionHistories.GetCurrentVersionHistory()
	s.NoError(err)
	err = versionHistory.AddOrUpdateItem(p.NewVersionHistoryItem(2, 0))
	s.NoError(err)

	err2 := s.UpdateWorkflowExecution(ctx, updatedInfo, updatedStats, versionHistories, []int64{int64(4)}, nil, common.EmptyEventID, nil, nil, nil, timerInfos, nil)
	s.NoError(err2)

	state, err1 := s.GetWorkflowExecutionInfo(ctx, domainID, workflowExecution)
	s.NoError(err1)
	s.NotNil(state, "expected valid state.")
	s.Equal(1, len(state.TimerInfos))
	s.Equal(int64(3345), state.TimerInfos[timerID].Version)
	s.Equal(timerID, state.TimerInfos[timerID].TimerID)
	s.EqualTimesWithPrecision(currentTime, state.TimerInfos[timerID].ExpiryTime, time.Millisecond*500)
	s.Equal(int64(2), state.TimerInfos[timerID].TaskStatus)
	s.Equal(int64(5), state.TimerInfos[timerID].StartedID)
	s.Equal(state.VersionHistories, versionHistories)
	s.assertChecksumsEqual(testWorkflowChecksum, state.Checksum)
}

//TestContinueAsNew test
func (s *ExecutionManagerSuiteForEventsV2) TestContinueAsNew() {
	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	domainID := uuid.New()
	workflowExecution := gen.WorkflowExecution{
		WorkflowId: common.StringPtr("continue-as-new-workflow-test"),
		RunId:      common.StringPtr("551c88d2-d9e6-404f-8131-9eec14f36643"),
	}
	decisionScheduleID := int64(2)
	_, err0 := s.CreateWorkflowExecution(ctx, domainID, workflowExecution, "queue1", "wType", 20, 13, nil, 3, 0, decisionScheduleID, nil)
	s.NoError(err0)

	state0, err1 := s.GetWorkflowExecutionInfo(ctx, domainID, workflowExecution)
	s.NoError(err1)
	info0 := state0.ExecutionInfo
	updatedInfo := copyWorkflowExecutionInfo(info0)
	updatedStats := copyExecutionStats(state0.ExecutionStats)
	updatedInfo.State = p.WorkflowStateCompleted
	updatedInfo.CloseStatus = p.WorkflowCloseStatusCompleted
	updatedInfo.NextEventID = int64(5)
	updatedInfo.LastProcessedEvent = int64(2)
	versionHistory := p.NewVersionHistory([]byte{}, []*p.VersionHistoryItem{
		{decisionScheduleID, common.EmptyVersion},
	})
	verisonHistories := p.NewVersionHistories(versionHistory)

	newWorkflowExecution := gen.WorkflowExecution{
		WorkflowId: common.StringPtr("continue-as-new-workflow-test"),
		RunId:      common.StringPtr("64c7e15a-3fd7-4182-9c6f-6f25a4fa2614"),
	}

	newdecisionTask := &p.DecisionTask{
		TaskID:     s.GetNextSequenceNumber(),
		DomainID:   updatedInfo.DomainID,
		TaskList:   updatedInfo.TaskList,
		ScheduleID: int64(2),
	}

	_, err2 := s.ExecutionManager.UpdateWorkflowExecution(ctx, &p.UpdateWorkflowExecutionRequest{
		UpdateWorkflowMutation: p.WorkflowMutation{
			ExecutionInfo:       updatedInfo,
			ExecutionStats:      updatedStats,
			TransferTasks:       []p.Task{newdecisionTask},
			TimerTasks:          nil,
			Condition:           info0.NextEventID,
			UpsertActivityInfos: nil,
			DeleteActivityInfos: nil,
			UpsertTimerInfos:    nil,
			DeleteTimerInfos:    nil,
			VersionHistories:    verisonHistories,
		},
		NewWorkflowSnapshot: &p.WorkflowSnapshot{
			ExecutionInfo: &p.WorkflowExecutionInfo{
				CreateRequestID:             uuid.New(),
				DomainID:                    updatedInfo.DomainID,
				WorkflowID:                  newWorkflowExecution.GetWorkflowId(),
				RunID:                       newWorkflowExecution.GetRunId(),
				TaskList:                    updatedInfo.TaskList,
				WorkflowTypeName:            updatedInfo.WorkflowTypeName,
				WorkflowTimeout:             updatedInfo.WorkflowTimeout,
				DecisionStartToCloseTimeout: updatedInfo.DecisionStartToCloseTimeout,
				ExecutionContext:            nil,
				State:                       p.WorkflowStateRunning,
				CloseStatus:                 p.WorkflowCloseStatusNone,
				NextEventID:                 info0.NextEventID,
				LastProcessedEvent:          common.EmptyEventID,
				DecisionScheduleID:          int64(2),
				DecisionStartedID:           common.EmptyEventID,
				DecisionTimeout:             1,
				BranchToken:                 []byte("branchToken1"),
			},
			ExecutionStats:   &p.ExecutionStats{},
			TransferTasks:    nil,
			TimerTasks:       nil,
			VersionHistories: verisonHistories,
		},
		RangeID:  s.ShardInfo.RangeID,
		Encoding: pickRandomEncoding(),
	})

	s.NoError(err2)

	prevExecutionState, err3 := s.GetWorkflowExecutionInfo(ctx, domainID, workflowExecution)
	s.NoError(err3)
	prevExecutionInfo := prevExecutionState.ExecutionInfo
	s.Equal(p.WorkflowStateCompleted, prevExecutionInfo.State)
	s.Equal(int64(5), prevExecutionInfo.NextEventID)
	s.Equal(int64(2), prevExecutionInfo.LastProcessedEvent)

	newExecutionState, err4 := s.GetWorkflowExecutionInfo(ctx, domainID, newWorkflowExecution)
	s.NoError(err4)
	newExecutionInfo := newExecutionState.ExecutionInfo
	s.Equal(p.WorkflowStateRunning, newExecutionInfo.State)
	s.Equal(p.WorkflowCloseStatusNone, newExecutionInfo.CloseStatus)
	s.Equal(int64(3), newExecutionInfo.NextEventID)
	s.Equal(common.EmptyEventID, newExecutionInfo.LastProcessedEvent)
	s.Equal(int64(2), newExecutionInfo.DecisionScheduleID)
	s.Equal([]byte("branchToken1"), newExecutionInfo.BranchToken)

	newRunID, err5 := s.GetCurrentWorkflowRunID(ctx, domainID, *workflowExecution.WorkflowId)
	s.NoError(err5)
	s.Equal(*newWorkflowExecution.RunId, newRunID)
}

func (s *ExecutionManagerSuiteForEventsV2) createWorkflowExecution(
	ctx context.Context,
	domainID string,
	workflowExecution gen.WorkflowExecution,
	taskList string,
	wType string,
	wTimeout int32,
	decisionTimeout int32,
	nextEventID int64,
	lastProcessedEventID int64,
	decisionScheduleID int64,
	txTasks []p.Task,
	brToken []byte,
) (*p.CreateWorkflowExecutionResponse, error) {

	var transferTasks []p.Task
	var replicationTasks []p.Task
	var timerTasks []p.Task
	for _, task := range txTasks {
		switch t := task.(type) {
		case *p.DecisionTask, *p.ActivityTask, *p.CloseExecutionTask, *p.CancelExecutionTask, *p.StartChildExecutionTask, *p.SignalExecutionTask, *p.RecordWorkflowStartedTask:
			transferTasks = append(transferTasks, t)
		case *p.HistoryReplicationTask:
			replicationTasks = append(replicationTasks, t)
		case *p.WorkflowTimeoutTask, *p.DeleteHistoryEventTask:
			timerTasks = append(timerTasks, t)
		default:
			panic("Unknown transfer task type.")
		}
	}

	transferTasks = append(transferTasks, &p.DecisionTask{
		TaskID:     s.GetNextSequenceNumber(),
		DomainID:   domainID,
		TaskList:   taskList,
		ScheduleID: decisionScheduleID,
	})
	versionHistory := p.NewVersionHistory([]byte{}, []*p.VersionHistoryItem{
		{decisionScheduleID, common.EmptyVersion},
	})
	verisonHistories := p.NewVersionHistories(versionHistory)
	response, err := s.ExecutionManager.CreateWorkflowExecution(ctx, &p.CreateWorkflowExecutionRequest{
		NewWorkflowSnapshot: p.WorkflowSnapshot{
			ExecutionInfo: &p.WorkflowExecutionInfo{
				CreateRequestID:             uuid.New(),
				DomainID:                    domainID,
				WorkflowID:                  workflowExecution.GetWorkflowId(),
				RunID:                       workflowExecution.GetRunId(),
				TaskList:                    taskList,
				WorkflowTypeName:            wType,
				WorkflowTimeout:             wTimeout,
				DecisionStartToCloseTimeout: decisionTimeout,
				State:                       p.WorkflowStateRunning,
				CloseStatus:                 p.WorkflowCloseStatusNone,
				NextEventID:                 nextEventID,
				LastProcessedEvent:          lastProcessedEventID,
				DecisionScheduleID:          decisionScheduleID,
				DecisionStartedID:           common.EmptyEventID,
				DecisionTimeout:             1,
				BranchToken:                 brToken,
			},
			ExecutionStats:   &p.ExecutionStats{},
			TimerTasks:       timerTasks,
			TransferTasks:    transferTasks,
			ReplicationTasks: replicationTasks,
			Checksum:         testWorkflowChecksum,
			VersionHistories: verisonHistories,
		},
		RangeID: s.ShardInfo.RangeID,
	})

	return response, err
}

// TestWorkflowResetWithCurrWithReplicate test
func (s *ExecutionManagerSuiteForEventsV2) TestWorkflowResetWithCurrWithReplicate() {
	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	domainID := uuid.New()
	runID := uuid.New()
	workflowExecution := gen.WorkflowExecution{
		WorkflowId: common.StringPtr("test-reset-workflow-with-replication-state-test"),
		RunId:      common.StringPtr(runID),
	}

	currentTime := time.Now()
	txTasks := []p.Task{&p.HistoryReplicationTask{
		TaskID:            s.GetNextSequenceNumber(),
		FirstEventID:      int64(1),
		NextEventID:       int64(3),
		Version:           int64(9),
		BranchToken:       []byte("branchToken1"),
		NewRunBranchToken: []byte("branchToken2"),
	},
		&p.WorkflowTimeoutTask{
			TaskID:              s.GetNextSequenceNumber(),
			VisibilityTimestamp: currentTime,
		}}

	task0, err0 := s.createWorkflowExecution(ctx, domainID, workflowExecution, "taskList", "wType", 20, 13, 3,
		0, 2, txTasks, []byte("branchToken1"))
	s.NoError(err0)
	s.NotNil(task0, "Expected non empty task identifier.")

	taskD, err := s.GetTransferTasks(ctx, 2, false)
	s.Equal(1, len(taskD), "Expected 1 decision task.")
	s.Equal(p.TransferTaskTypeDecisionTask, taskD[0].TaskType)
	err = s.CompleteTransferTask(ctx, taskD[0].TaskID)
	s.NoError(err)
	taskD, err = s.GetTransferTasks(ctx, 2, false)
	s.Equal(0, len(taskD), "Expected 0 decision task.")

	taskT, err := s.GetTimerIndexTasks(ctx, 2, false)
	s.Equal(1, len(taskT), "Expected 1 timer task.")
	s.Equal(p.TaskTypeWorkflowTimeout, taskT[0].TaskType)
	err = s.CompleteTimerTask(ctx, taskT[0].VisibilityTimestamp, taskT[0].TaskID)
	s.NoError(err)
	taskT, err = s.GetTimerIndexTasks(ctx, 2, false)
	s.Equal(0, len(taskT), "Expected 0 timer task.")

	taskR, err := s.GetReplicationTasks(ctx, 2, false)
	s.Nil(err)
	s.Equal(1, len(taskR), "Expected 1 replication task.")
	tsk := taskR[0]
	s.Equal(p.ReplicationTaskTypeHistory, tsk.TaskType)
	s.Equal(domainID, tsk.DomainID)
	s.Equal(*workflowExecution.WorkflowId, tsk.WorkflowID)
	s.Equal(*workflowExecution.RunId, tsk.RunID)
	s.Equal(int64(1), tsk.FirstEventID)
	s.Equal(int64(3), tsk.NextEventID)
	s.Equal(int64(9), tsk.Version)
	s.Equal([]byte("branchToken1"), tsk.BranchToken)
	s.Equal([]byte("branchToken2"), tsk.NewRunBranchToken)
	err = s.CompleteReplicationTask(ctx, taskR[0].TaskID)
	s.NoError(err)
	taskR, err = s.GetReplicationTasks(ctx, 2, false)
	s.Nil(err)
	s.Equal(0, len(taskR), "Expected 0 replication task.")

	state0, err1 := s.GetWorkflowExecutionInfo(ctx, domainID, workflowExecution)
	s.NoError(err1)
	info0 := state0.ExecutionInfo
	s.NotNil(info0, "Valid Workflow info expected.")
	s.Equal(domainID, info0.DomainID)
	s.Equal("taskList", info0.TaskList)
	s.Equal("wType", info0.WorkflowTypeName)
	s.Equal(int32(20), info0.WorkflowTimeout)
	s.Equal(int32(13), info0.DecisionStartToCloseTimeout)
	s.Equal(int64(3), info0.NextEventID)
	s.Equal(int64(0), info0.LastProcessedEvent)
	s.Equal(int64(2), info0.DecisionScheduleID)

	updatedInfo := copyWorkflowExecutionInfo(info0)
	updateStats := copyExecutionStats(state0.ExecutionStats)
	updatedInfo.NextEventID = int64(5)
	updatedInfo.LastProcessedEvent = int64(2)
	updatedInfo.BranchToken = []byte("branchToken3")

	currTransTasks := []p.Task{
		&p.CloseExecutionTask{
			TaskID:              s.GetNextSequenceNumber(),
			VisibilityTimestamp: time.Now(),
			Version:             100,
		},
	}

	currTimerTasks := []p.Task{
		&p.DeleteHistoryEventTask{
			TaskID:              20,
			VisibilityTimestamp: time.Now(),
			Version:             101,
		},
	}

	newRunID := uuid.New()
	newExecution := gen.WorkflowExecution{
		WorkflowId: workflowExecution.WorkflowId,
		RunId:      common.StringPtr(newRunID),
	}
	insertInfo := copyWorkflowExecutionInfo(info0)
	insertStats := copyExecutionStats(state0.ExecutionStats)
	insertInfo.RunID = newRunID
	insertInfo.NextEventID = int64(50)
	insertInfo.LastProcessedEvent = int64(20)
	insertInfo.BranchToken = []byte("branchToken4")

	insertTransTasks := []p.Task{
		&p.DecisionTask{
			TaskID:              s.GetNextSequenceNumber(),
			DomainID:            domainID,
			VisibilityTimestamp: time.Now(),
			ScheduleID:          13,
			Version:             200,
		},
	}

	insertTimerTasks := []p.Task{
		&p.WorkflowTimeoutTask{
			TaskID:              s.GetNextSequenceNumber(),
			VisibilityTimestamp: time.Now().Add(time.Minute),
			Version:             201,
		},
	}

	insertReplicationTasks := []p.Task{&p.HistoryReplicationTask{
		TaskID:       s.GetNextSequenceNumber(),
		FirstEventID: int64(10),
		NextEventID:  int64(30),
		Version:      int64(90),
		BranchToken:  []byte("branchToken5"),
	}}

	insertTimerInfos := []*p.TimerInfo{{
		Version:    100,
		TimerID:    "id101",
		ExpiryTime: currentTime,
		TaskStatus: 102,
		StartedID:  103,
	}}

	insertActivityInfos := []*p.ActivityInfo{{
		Version:        110,
		ScheduleID:     111,
		StartedID:      112,
		ActivityID:     uuid.New(),
		ScheduledEvent: &gen.HistoryEvent{EventId: int64Ptr(1)},
	}}

	insertRequestCancelInfos := []*p.RequestCancelInfo{{
		Version:         120,
		InitiatedID:     121,
		CancelRequestID: uuid.New(),
	}}

	insertChildExecutionInfos := []*p.ChildExecutionInfo{{
		Version:         130,
		InitiatedID:     131,
		StartedID:       132,
		CreateRequestID: uuid.New(),
		InitiatedEvent:  &gen.HistoryEvent{EventId: int64Ptr(1)},
	}}

	insertSignalInfos := []*p.SignalInfo{{
		Version:         140,
		InitiatedID:     141,
		SignalName:      "142",
		SignalRequestID: uuid.New(),
	}}

	insertSignalRequests := []string{uuid.New()}

	err = s.ResetWorkflowExecution(
		ctx,
		3,
		insertInfo,
		insertStats,
		insertActivityInfos,
		insertTimerInfos,
		insertChildExecutionInfos,
		insertRequestCancelInfos,
		insertSignalInfos,
		insertSignalRequests,
		insertTransTasks,
		insertTimerTasks,
		insertReplicationTasks,
		true,
		updatedInfo,
		updateStats,
		currTransTasks,
		currTimerTasks,
		info0.RunID,
		-1000,
	)
	s.Nil(err)

	//////////////////////////////
	// start verifying resetWF
	///////////////////////////////

	// transfer tasks
	taskD, err = s.GetTransferTasks(ctx, 3, false)
	s.Equal(2, len(taskD), "Expected 2 decision task.")
	s.Equal(p.TransferTaskTypeCloseExecution, taskD[0].TaskType)
	s.Equal(int64(100), taskD[0].Version)
	err = s.CompleteTransferTask(ctx, taskD[0].TaskID)
	s.NoError(err)
	s.Equal(p.TransferTaskTypeDecisionTask, taskD[1].TaskType)
	s.Equal(int64(200), taskD[1].Version)
	err = s.CompleteTransferTask(ctx, taskD[1].TaskID)
	s.NoError(err)
	taskD, err = s.GetTransferTasks(ctx, 2, false)
	s.Equal(0, len(taskD), "Expected 0 decision task.")

	// timer tasks
	taskT, err = s.GetTimerIndexTasks(ctx, 3, false)
	s.Equal(2, len(taskT), "Expected 2 timer task.")
	s.Equal(p.TaskTypeDeleteHistoryEvent, taskT[0].TaskType)
	s.Equal(int64(101), taskT[0].Version)
	err = s.CompleteTimerTask(ctx, taskT[0].VisibilityTimestamp, taskT[0].TaskID)
	s.NoError(err)
	s.Equal(p.TaskTypeWorkflowTimeout, taskT[1].TaskType)
	s.Equal(int64(201), taskT[1].Version)
	err = s.CompleteTimerTask(ctx, taskT[1].VisibilityTimestamp, taskT[1].TaskID)
	s.NoError(err)
	taskT, err = s.GetTimerIndexTasks(ctx, 2, false)
	s.Equal(0, len(taskT), "Expected 0 timer task.")

	// replicaiton tasks
	taskR, err = s.GetReplicationTasks(ctx, 2, false)
	s.Nil(err)
	s.Equal(1, len(taskR), "Expected 1 replication task.")
	tsk = taskR[0]
	s.Equal(p.ReplicationTaskTypeHistory, tsk.TaskType)
	s.Equal(domainID, tsk.DomainID)
	s.Equal(*workflowExecution.WorkflowId, tsk.WorkflowID)
	s.Equal(insertInfo.RunID, tsk.RunID)
	s.Equal(int64(10), tsk.FirstEventID)
	s.Equal(int64(30), tsk.NextEventID)
	s.Equal(int64(90), tsk.Version)
	s.Equal([]byte("branchToken5"), tsk.BranchToken)
	s.Equal(0, len(tsk.NewRunBranchToken))
	err = s.CompleteReplicationTask(ctx, taskR[0].TaskID)
	s.NoError(err)
	taskR, err = s.GetReplicationTasks(ctx, 2, false)
	s.Nil(err)
	s.Equal(0, len(taskR), "Expected 0 replication task.")

	// check current run
	currRunID, err := s.GetCurrentWorkflowRunID(ctx, domainID, workflowExecution.GetWorkflowId())
	s.Nil(err)
	s.Equal(newExecution.GetRunId(), currRunID)

	// the previous execution
	state1, err1 := s.GetWorkflowExecutionInfo(ctx, domainID, workflowExecution)
	s.NoError(err1)
	info1 := state1.ExecutionInfo
	s.NotNil(info1, "Valid Workflow info expected.")
	s.Equal(int64(5), info1.NextEventID)
	s.Equal(int64(2), info1.LastProcessedEvent)
	s.Equal([]byte("branchToken3"), info1.BranchToken)
	s.Equal(domainID, info1.DomainID)
	s.Equal("taskList", info1.TaskList)
	s.Equal("wType", info1.WorkflowTypeName)
	s.Equal(int32(20), info1.WorkflowTimeout)
	s.Equal(int32(13), info1.DecisionStartToCloseTimeout)
	s.Equal(int64(2), info1.DecisionScheduleID)

	// the current execution
	state2, err2 := s.GetWorkflowExecutionInfo(ctx, domainID, newExecution)
	s.NoError(err2)
	info2 := state2.ExecutionInfo

	s.NotNil(info2, "Valid Workflow info expected.")
	s.Equal(int64(50), info2.NextEventID)
	s.Equal(int64(20), info2.LastProcessedEvent)
	s.Equal([]byte("branchToken4"), info2.BranchToken)
	s.Equal(domainID, info2.DomainID)
	s.Equal("taskList", info2.TaskList)
	s.Equal("wType", info2.WorkflowTypeName)
	s.Equal(int32(20), info2.WorkflowTimeout)
	s.Equal(int32(13), info2.DecisionStartToCloseTimeout)
	s.Equal(int64(2), info2.DecisionScheduleID)

	timerInfos2 := state2.TimerInfos
	actInfos2 := state2.ActivityInfos
	reqCanInfos2 := state2.RequestCancelInfos
	childInfos2 := state2.ChildExecutionInfos
	sigInfos2 := state2.SignalInfos
	sigReqIDs2 := state2.SignalRequestedIDs

	s.Equal(1, len(timerInfos2))
	s.Equal(1, len(actInfos2))
	s.Equal(1, len(reqCanInfos2))
	s.Equal(1, len(childInfos2))
	s.Equal(1, len(sigInfos2))
	s.Equal(1, len(sigReqIDs2))

	s.Equal(int64(100), timerInfos2["id101"].Version)
	s.Equal(int64(110), actInfos2[111].Version)
	s.Equal(int64(120), reqCanInfos2[121].Version)
	s.Equal(int64(130), childInfos2[131].Version)
	s.Equal(int64(140), sigInfos2[141].Version)
}

// TestWorkflowResetNoCurrWithReplicate test
func (s *ExecutionManagerSuiteForEventsV2) TestWorkflowResetNoCurrWithReplicate() {
	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	domainID := uuid.New()
	runID := uuid.New()
	workflowExecution := gen.WorkflowExecution{
		WorkflowId: common.StringPtr("test-reset-workflow-with-replication-state-test"),
		RunId:      common.StringPtr(runID),
	}

	currentTime := time.Now()
	txTasks := []p.Task{&p.HistoryReplicationTask{
		TaskID:            s.GetNextSequenceNumber(),
		FirstEventID:      int64(1),
		NextEventID:       int64(3),
		Version:           int64(9),
		BranchToken:       []byte("branchToken1"),
		NewRunBranchToken: []byte("branchToken2"),
	},
		&p.WorkflowTimeoutTask{
			TaskID:              s.GetNextSequenceNumber(),
			VisibilityTimestamp: currentTime,
		}}

	task0, err0 := s.createWorkflowExecution(
		ctx,
		domainID,
		workflowExecution,
		"taskList",
		"wType",
		20,
		13,
		3,
		0,
		2,
		txTasks,
		[]byte("branchToken1"),
	)
	s.NoError(err0)
	s.NotNil(task0, "Expected non empty task identifier.")

	taskD, err := s.GetTransferTasks(ctx, 2, false)
	s.Equal(1, len(taskD), "Expected 1 decision task.")
	s.Equal(p.TransferTaskTypeDecisionTask, taskD[0].TaskType)
	err = s.CompleteTransferTask(ctx, taskD[0].TaskID)
	s.NoError(err)
	taskD, err = s.GetTransferTasks(ctx, 2, false)
	s.Equal(0, len(taskD), "Expected 0 decision task.")

	taskT, err := s.GetTimerIndexTasks(ctx, 2, false)
	s.Equal(1, len(taskT), "Expected 1 timer task.")
	s.Equal(p.TaskTypeWorkflowTimeout, taskT[0].TaskType)
	err = s.CompleteTimerTask(ctx, taskT[0].VisibilityTimestamp, taskT[0].TaskID)
	s.NoError(err)
	taskT, err = s.GetTimerIndexTasks(ctx, 2, false)
	s.Equal(0, len(taskT), "Expected 0 timer task.")

	taskR, err := s.GetReplicationTasks(ctx, 2, false)
	s.Nil(err)
	s.Equal(1, len(taskR), "Expected 1 replication task.")
	tsk := taskR[0]
	s.Equal(p.ReplicationTaskTypeHistory, tsk.TaskType)
	s.Equal(domainID, tsk.DomainID)
	s.Equal(*workflowExecution.WorkflowId, tsk.WorkflowID)
	s.Equal(*workflowExecution.RunId, tsk.RunID)
	s.Equal(int64(1), tsk.FirstEventID)
	s.Equal(int64(3), tsk.NextEventID)
	s.Equal(int64(9), tsk.Version)
	s.Equal([]byte("branchToken1"), tsk.BranchToken)
	s.Equal([]byte("branchToken2"), tsk.NewRunBranchToken)
	err = s.CompleteReplicationTask(ctx, taskR[0].TaskID)
	s.NoError(err)
	taskR, err = s.GetReplicationTasks(ctx, 2, false)
	s.Nil(err)
	s.Equal(0, len(taskR), "Expected 0 replication task.")

	state0, err1 := s.GetWorkflowExecutionInfo(ctx, domainID, workflowExecution)
	s.NoError(err1)
	info0 := state0.ExecutionInfo
	stats0 := state0.ExecutionStats
	s.NotNil(info0, "Valid Workflow info expected.")
	s.Equal(domainID, info0.DomainID)
	s.Equal("taskList", info0.TaskList)
	s.Equal("wType", info0.WorkflowTypeName)
	s.Equal(int32(20), info0.WorkflowTimeout)
	s.Equal(int32(13), info0.DecisionStartToCloseTimeout)
	s.Equal(int64(3), info0.NextEventID)
	s.Equal(int64(0), info0.LastProcessedEvent)
	s.Equal(int64(2), info0.DecisionScheduleID)
	s.Equal(p.WorkflowStateRunning, info0.State)
	s.Equal(p.WorkflowCloseStatusNone, info0.CloseStatus)

	info0.State = p.WorkflowStateCompleted
	info0.CloseStatus = p.WorkflowCloseStatusCompleted
	versionHistory := p.NewVersionHistory([]byte{}, []*p.VersionHistoryItem{
		{info0.DecisionScheduleID, common.EmptyVersion},
	})
	verisonHistories := p.NewVersionHistories(versionHistory)
	err = s.UpdateWorklowStateAndReplication(ctx, info0, stats0, verisonHistories, info0.NextEventID, nil)
	s.Nil(err)

	updatedInfo := copyWorkflowExecutionInfo(info0)
	updatedStats := copyExecutionStats(state0.ExecutionStats)

	newRunID := uuid.New()
	newExecution := gen.WorkflowExecution{
		WorkflowId: workflowExecution.WorkflowId,
		RunId:      common.StringPtr(newRunID),
	}
	insertInfo := copyWorkflowExecutionInfo(info0)
	insterStats := copyExecutionStats(state0.ExecutionStats)
	insertInfo.State = p.WorkflowStateRunning
	insertInfo.CloseStatus = p.WorkflowCloseStatusNone
	insertInfo.RunID = newRunID
	insertInfo.NextEventID = int64(50)
	insertInfo.LastProcessedEvent = int64(20)
	insertInfo.BranchToken = []byte("branchToken4")

	insertTransTasks := []p.Task{
		&p.DecisionTask{
			TaskID:              s.GetNextSequenceNumber(),
			DomainID:            domainID,
			VisibilityTimestamp: time.Now(),
			ScheduleID:          13,
			Version:             200,
		},
	}

	insertTimerTasks := []p.Task{
		&p.WorkflowTimeoutTask{
			TaskID:              s.GetNextSequenceNumber(),
			VisibilityTimestamp: time.Now().Add(time.Minute),
			Version:             201,
		},
	}

	insertReplicationTasks := []p.Task{&p.HistoryReplicationTask{
		TaskID:       s.GetNextSequenceNumber(),
		FirstEventID: int64(10),
		NextEventID:  int64(30),
		Version:      int64(90),
		BranchToken:  []byte("branchToken5"),
	}}

	insertTimerInfos := []*p.TimerInfo{{
		Version:    100,
		TimerID:    "id101",
		ExpiryTime: currentTime,
		TaskStatus: 102,
		StartedID:  103,
	}}

	insertActivityInfos := []*p.ActivityInfo{{
		Version:        110,
		ScheduleID:     111,
		StartedID:      112,
		ActivityID:     uuid.New(),
		ScheduledEvent: &gen.HistoryEvent{EventId: int64Ptr(1)},
	}}

	insertRequestCancelInfos := []*p.RequestCancelInfo{{
		Version:         120,
		InitiatedID:     121,
		CancelRequestID: uuid.New(),
	}}

	insertChildExecutionInfos := []*p.ChildExecutionInfo{{
		Version:         130,
		InitiatedID:     131,
		StartedID:       132,
		CreateRequestID: uuid.New(),
		InitiatedEvent:  &gen.HistoryEvent{EventId: int64Ptr(1)},
	}}

	insertSignalInfos := []*p.SignalInfo{{
		Version:         140,
		InitiatedID:     141,
		SignalName:      "142",
		SignalRequestID: uuid.New(),
	}}

	insertSignalRequests := []string{uuid.New()}

	err = s.ResetWorkflowExecution(
		ctx,
		3,
		insertInfo,
		insterStats,
		insertActivityInfos,
		insertTimerInfos,
		insertChildExecutionInfos,
		insertRequestCancelInfos,
		insertSignalInfos,
		insertSignalRequests,
		insertTransTasks,
		insertTimerTasks,
		insertReplicationTasks,
		false,
		updatedInfo,
		updatedStats,
		nil,
		nil,
		info0.RunID,
		-1000,
	)
	s.Nil(err)

	//////////////////////////////
	// start verifying resetWF
	///////////////////////////////

	// transfer tasks
	taskD, err = s.GetTransferTasks(ctx, 3, false)
	s.Equal(1, len(taskD), "Expected 1 decision task.")

	s.Equal(p.TransferTaskTypeDecisionTask, taskD[0].TaskType)
	s.Equal(int64(200), taskD[0].Version)
	err = s.CompleteTransferTask(ctx, taskD[0].TaskID)
	s.NoError(err)
	taskD, err = s.GetTransferTasks(ctx, 2, false)
	s.Equal(0, len(taskD), "Expected 0 decision task.")

	// timer tasks
	taskT, err = s.GetTimerIndexTasks(ctx, 3, false)
	s.Equal(1, len(taskT), "Expected 1 timer task.")
	s.Equal(p.TaskTypeWorkflowTimeout, taskT[0].TaskType)
	s.Equal(int64(201), taskT[0].Version)
	err = s.CompleteTimerTask(ctx, taskT[0].VisibilityTimestamp, taskT[0].TaskID)
	s.NoError(err)
	taskT, err = s.GetTimerIndexTasks(ctx, 2, false)
	s.Equal(0, len(taskT), "Expected 0 timer task.")

	// replicaiton tasks
	taskR, err = s.GetReplicationTasks(ctx, 2, false)
	s.Nil(err)
	s.Equal(1, len(taskR), "Expected 1 replication task.")
	tsk = taskR[0]
	s.Equal(p.ReplicationTaskTypeHistory, tsk.TaskType)
	s.Equal(domainID, tsk.DomainID)
	s.Equal(*workflowExecution.WorkflowId, tsk.WorkflowID)
	s.Equal(insertInfo.RunID, tsk.RunID)
	s.Equal(int64(10), tsk.FirstEventID)
	s.Equal(int64(30), tsk.NextEventID)
	s.Equal(int64(90), tsk.Version)
	s.Equal([]byte("branchToken5"), tsk.BranchToken)
	s.Equal(0, len(tsk.NewRunBranchToken))
	err = s.CompleteReplicationTask(ctx, taskR[0].TaskID)
	s.NoError(err)
	taskR, err = s.GetReplicationTasks(ctx, 2, false)
	s.Nil(err)
	s.Equal(0, len(taskR), "Expected 0 replication task.")

	// check current run
	currRunID, err := s.GetCurrentWorkflowRunID(ctx, domainID, workflowExecution.GetWorkflowId())
	s.Nil(err)
	s.Equal(newExecution.GetRunId(), currRunID)

	// the previous execution
	state1, err1 := s.GetWorkflowExecutionInfo(ctx, domainID, workflowExecution)
	s.NoError(err1)
	info1 := state1.ExecutionInfo
	s.NotNil(info1, "Valid Workflow info expected.")
	s.Equal(int64(3), info1.NextEventID)
	s.Equal(int64(0), info1.LastProcessedEvent)
	s.Equal([]byte("branchToken1"), info1.BranchToken)
	s.Equal(domainID, info1.DomainID)
	s.Equal("taskList", info1.TaskList)
	s.Equal("wType", info1.WorkflowTypeName)
	s.Equal(int32(20), info1.WorkflowTimeout)
	s.Equal(int32(13), info1.DecisionStartToCloseTimeout)
	s.Equal(int64(2), info1.DecisionScheduleID)

	// the current execution
	state2, err2 := s.GetWorkflowExecutionInfo(ctx, domainID, newExecution)
	s.NoError(err2)
	info2 := state2.ExecutionInfo

	s.NotNil(info2, "Valid Workflow info expected.")
	s.Equal(int64(50), info2.NextEventID)
	s.Equal(int64(20), info2.LastProcessedEvent)
	s.Equal([]byte("branchToken4"), info2.BranchToken)
	s.Equal(domainID, info2.DomainID)
	s.Equal("taskList", info2.TaskList)
	s.Equal("wType", info2.WorkflowTypeName)
	s.Equal(int32(20), info2.WorkflowTimeout)
	s.Equal(int32(13), info2.DecisionStartToCloseTimeout)
	s.Equal(int64(2), info2.DecisionScheduleID)

	timerInfos2 := state2.TimerInfos
	actInfos2 := state2.ActivityInfos
	reqCanInfos2 := state2.RequestCancelInfos
	childInfos2 := state2.ChildExecutionInfos
	sigInfos2 := state2.SignalInfos
	sigReqIDs2 := state2.SignalRequestedIDs

	s.Equal(1, len(timerInfos2))
	s.Equal(1, len(actInfos2))
	s.Equal(1, len(reqCanInfos2))
	s.Equal(1, len(childInfos2))
	s.Equal(1, len(sigInfos2))
	s.Equal(1, len(sigReqIDs2))

	s.Equal(int64(100), timerInfos2["id101"].Version)
	s.Equal(int64(110), actInfos2[111].Version)
	s.Equal(int64(120), reqCanInfos2[121].Version)
	s.Equal(int64(130), childInfos2[131].Version)
	s.Equal(int64(140), sigInfos2[141].Version)
}

// TestWorkflowResetNoCurrNoReplicate test
func (s *ExecutionManagerSuiteForEventsV2) TestWorkflowResetNoCurrNoReplicate() {
	ctx, cancel := context.WithTimeout(context.Background(), testContextTimeout)
	defer cancel()

	domainID := uuid.New()
	runID := uuid.New()
	workflowExecution := gen.WorkflowExecution{
		WorkflowId: common.StringPtr("test-reset-workflow-with-replication-state-test"),
		RunId:      common.StringPtr(runID),
	}

	currentTime := time.Now()
	txTasks := []p.Task{
		&p.WorkflowTimeoutTask{
			TaskID:              s.GetNextSequenceNumber(),
			VisibilityTimestamp: currentTime,
		}}

	task0, err0 := s.CreateWorkflowExecution(ctx, domainID, workflowExecution, "taskList", "wType",
		20, 13, nil, 3, 0, 2, txTasks)
	s.NoError(err0)
	s.NotNil(task0, "Expected non empty task identifier.")

	taskD, err := s.GetTransferTasks(ctx, 2, false)
	s.Equal(1, len(taskD), "Expected 1 decision task.")
	s.Equal(p.TransferTaskTypeDecisionTask, taskD[0].TaskType)
	err = s.CompleteTransferTask(ctx, taskD[0].TaskID)
	s.NoError(err)
	taskD, err = s.GetTransferTasks(ctx, 2, false)
	s.Equal(0, len(taskD), "Expected 0 decision task.")

	taskT, err := s.GetTimerIndexTasks(ctx, 2, false)
	s.Equal(1, len(taskT), "Expected 1 timer task.")
	s.Equal(p.TaskTypeWorkflowTimeout, taskT[0].TaskType)
	err = s.CompleteTimerTask(ctx, taskT[0].VisibilityTimestamp, taskT[0].TaskID)
	s.NoError(err)
	taskT, err = s.GetTimerIndexTasks(ctx, 2, false)
	s.Equal(0, len(taskT), "Expected 0 timer task.")

	state0, err1 := s.GetWorkflowExecutionInfo(ctx, domainID, workflowExecution)
	s.NoError(err1)
	info0 := state0.ExecutionInfo
	s.NotNil(info0, "Valid Workflow info expected.")
	s.Equal(domainID, info0.DomainID)
	s.Equal("taskList", info0.TaskList)
	s.Equal("wType", info0.WorkflowTypeName)
	s.Equal(int32(20), info0.WorkflowTimeout)
	s.Equal(int32(13), info0.DecisionStartToCloseTimeout)
	s.Equal(int64(3), info0.NextEventID)
	s.Equal(int64(0), info0.LastProcessedEvent)
	s.Equal(int64(2), info0.DecisionScheduleID)

	updatedInfo := copyWorkflowExecutionInfo(info0)
	updatedStats := copyExecutionStats(state0.ExecutionStats)

	newRunID := uuid.New()
	newExecution := gen.WorkflowExecution{
		WorkflowId: workflowExecution.WorkflowId,
		RunId:      common.StringPtr(newRunID),
	}
	insertInfo := copyWorkflowExecutionInfo(info0)
	insertStats := copyExecutionStats(state0.ExecutionStats)
	insertInfo.RunID = newRunID
	insertInfo.NextEventID = int64(50)
	insertInfo.LastProcessedEvent = int64(20)
	insertInfo.BranchToken = []byte("branchToken4")

	insertTransTasks := []p.Task{
		&p.DecisionTask{
			TaskID:              s.GetNextSequenceNumber(),
			DomainID:            domainID,
			VisibilityTimestamp: time.Now(),
			ScheduleID:          13,
			Version:             200,
		},
	}

	insertTimerTasks := []p.Task{
		&p.WorkflowTimeoutTask{
			TaskID:              s.GetNextSequenceNumber(),
			VisibilityTimestamp: time.Now().Add(time.Minute),
			Version:             201,
		},
	}

	insertTimerInfos := []*p.TimerInfo{{
		Version:    100,
		TimerID:    "id101",
		ExpiryTime: currentTime,
		TaskStatus: 102,
		StartedID:  103,
	}}

	insertActivityInfos := []*p.ActivityInfo{{
		Version:        110,
		ScheduleID:     111,
		StartedID:      112,
		ActivityID:     uuid.New(),
		ScheduledEvent: &gen.HistoryEvent{EventId: int64Ptr(1)},
	}}

	insertRequestCancelInfos := []*p.RequestCancelInfo{{
		Version:         120,
		InitiatedID:     121,
		CancelRequestID: uuid.New(),
	}}

	err = s.ResetWorkflowExecution(
		ctx,
		3,
		insertInfo,
		insertStats,
		insertActivityInfos,
		insertTimerInfos,
		nil,
		insertRequestCancelInfos,
		nil,
		nil,
		insertTransTasks,
		insertTimerTasks,
		nil,
		false,
		updatedInfo,
		updatedStats,
		nil,
		nil,
		info0.RunID,
		-1000,
	)
	s.Nil(err)

	//////////////////////////////
	// start verifying resetWF
	///////////////////////////////

	// transfer tasks
	taskD, err = s.GetTransferTasks(ctx, 3, false)
	s.Equal(1, len(taskD), "Expected 1 decision task.")

	s.Equal(p.TransferTaskTypeDecisionTask, taskD[0].TaskType)
	s.Equal(int64(200), taskD[0].Version)
	err = s.CompleteTransferTask(ctx, taskD[0].TaskID)
	s.NoError(err)
	taskD, err = s.GetTransferTasks(ctx, 2, false)
	s.Equal(0, len(taskD), "Expected 0 decision task.")

	// timer tasks
	taskT, err = s.GetTimerIndexTasks(ctx, 3, false)
	s.Equal(1, len(taskT), "Expected 1 timer task.")
	s.Equal(p.TaskTypeWorkflowTimeout, taskT[0].TaskType)
	s.Equal(int64(201), taskT[0].Version)
	err = s.CompleteTimerTask(ctx, taskT[0].VisibilityTimestamp, taskT[0].TaskID)
	s.NoError(err)
	taskT, err = s.GetTimerIndexTasks(ctx, 2, false)
	s.Equal(0, len(taskT), "Expected 0 timer task.")

	// check current run
	currRunID, err := s.GetCurrentWorkflowRunID(ctx, domainID, workflowExecution.GetWorkflowId())
	s.Nil(err)
	s.Equal(newExecution.GetRunId(), currRunID)

	// the previous execution
	state1, err1 := s.GetWorkflowExecutionInfo(ctx, domainID, workflowExecution)
	s.NoError(err1)
	info1 := state1.ExecutionInfo
	s.NotNil(info1, "Valid Workflow info expected.")
	s.Equal(int64(3), info1.NextEventID)
	s.Equal(int64(0), info1.LastProcessedEvent)
	s.Equal(domainID, info1.DomainID)
	s.Equal("taskList", info1.TaskList)
	s.Equal("wType", info1.WorkflowTypeName)
	s.Equal(int32(20), info1.WorkflowTimeout)
	s.Equal(int32(13), info1.DecisionStartToCloseTimeout)
	s.Equal(int64(2), info1.DecisionScheduleID)

	// the current execution
	state2, err2 := s.GetWorkflowExecutionInfo(ctx, domainID, newExecution)
	s.NoError(err2)
	info2 := state2.ExecutionInfo

	s.NotNil(info2, "Valid Workflow info expected.")
	s.Equal(int64(50), info2.NextEventID)
	s.Equal(int64(20), info2.LastProcessedEvent)
	s.Equal([]byte("branchToken4"), info2.BranchToken)
	s.Equal(domainID, info2.DomainID)
	s.Equal("taskList", info2.TaskList)
	s.Equal("wType", info2.WorkflowTypeName)
	s.Equal(int32(20), info2.WorkflowTimeout)
	s.Equal(int32(13), info2.DecisionStartToCloseTimeout)
	s.Equal(int64(2), info2.DecisionScheduleID)

	timerInfos2 := state2.TimerInfos
	actInfos2 := state2.ActivityInfos
	reqCanInfos2 := state2.RequestCancelInfos
	childInfos2 := state2.ChildExecutionInfos
	sigInfos2 := state2.SignalInfos
	sigReqIDs2 := state2.SignalRequestedIDs

	s.Equal(1, len(timerInfos2))
	s.Equal(1, len(actInfos2))
	s.Equal(1, len(reqCanInfos2))
	s.Equal(0, len(childInfos2))
	s.Equal(0, len(sigInfos2))
	s.Equal(0, len(sigReqIDs2))

	s.Equal(int64(100), timerInfos2["id101"].Version)
	s.Equal(int64(110), actInfos2[111].Version)
	s.Equal(int64(120), reqCanInfos2[121].Version)
}
