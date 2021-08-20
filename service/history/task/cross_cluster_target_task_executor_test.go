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

package task

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/shard"
)

type (
	crossClusterTargetTaskExecutorSuite struct {
		suite.Suite
		*require.Assertions

		controller        *gomock.Controller
		mockShard         *shard.TestContext
		mockHistoryClient *history.MockClient

		executor Executor
	}
)

func TestCrossClusterTargetTaskExecutorSuite(t *testing.T) {
	s := new(crossClusterTargetTaskExecutorSuite)
	suite.Run(t, s)
}

func (s *crossClusterTargetTaskExecutorSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	config := config.NewForTest()
	s.controller = gomock.NewController(s.T())
	s.mockShard = shard.NewTestContext(
		s.controller,
		&persistence.ShardInfo{
			ShardID: 0,
			RangeID: 1,
		},
		config,
	)
	s.mockHistoryClient = s.mockShard.Resource.HistoryClient
	mockDomainCache := s.mockShard.Resource.DomainCache
	mockDomainCache.EXPECT().GetDomainName(constants.TestDomainID).Return(constants.TestDomainName, nil).AnyTimes()
	mockDomainCache.EXPECT().GetDomainByID(constants.TestTargetDomainID).Return(constants.TestGlobalTargetDomainEntry, nil).AnyTimes()
	mockDomainCache.EXPECT().GetDomainByID(constants.TestRemoteTargetDomainID).Return(constants.TestGlobalRemoteTargetDomainEntry, nil).AnyTimes()
	mockClusterMetadata := s.mockShard.Resource.ClusterMetadata
	mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	mockClusterMetadata.EXPECT().GetAllClusterInfo().Return(cluster.TestAllClusterInfo).AnyTimes()
	mockClusterMetadata.EXPECT().IsGlobalDomainEnabled().Return(true).AnyTimes()

	s.executor = NewCrossClusterTargetTaskExecutor(s.mockShard, s.mockShard.GetLogger(), config)
}

func (s *crossClusterTargetTaskExecutorSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *crossClusterTargetTaskExecutorSuite) TestExecute_UnexpectedTask() {
	timerTask := NewTimerTask(
		s.mockShard,
		&persistence.TimerTaskInfo{},
		QueueTypeActiveTimer,
		nil, nil, nil, nil, nil, nil,
	)

	var err error
	s.NotPanics(func() {
		err = s.executor.Execute(timerTask, true)
	})
	s.Equal(errUnexpectedTask, err)
}

func (s *crossClusterTargetTaskExecutorSuite) TestExecute_InvalidProcessingState() {
	task := s.getTestCancelExecutionTask(processingStateResponseRecorded)

	err := s.executor.Execute(task, true)
	s.NoError(err)

	s.Equal(task.GetTaskID(), task.response.GetTaskID())
	s.Equal(types.CrossClusterTaskTypeCancelExecution, task.response.GetTaskType())
	s.Equal(types.CrossClusterTaskFailedCauseUncategorized, task.response.GetFailedCause())
}

func (s *crossClusterTargetTaskExecutorSuite) TestExecute_InvalidTargetDomain() {
	task := s.getTestStartChildExecutionTask(processingStateInitialized, nil)
	task.request.StartChildExecutionAttributes.TargetDomainID = constants.TestRemoteTargetDomainID

	err := s.executor.Execute(task, true)
	s.NoError(err)

	s.Equal(task.GetTaskID(), task.response.GetTaskID())
	s.Equal(types.CrossClusterTaskTypeStartChildExecution, task.response.GetTaskType())
	s.Equal(types.CrossClusterTaskFailedCauseDomainNotActive, task.response.GetFailedCause())
}

func (s *crossClusterTargetTaskExecutorSuite) TestStartChildExecutionTask_StartChildSuccess() {
	task := s.getTestStartChildExecutionTask(processingStateInitialized, nil)

	historyReq := createTestChildWorkflowExecutionRequest(
		constants.TestDomainName,
		constants.TestTargetDomainName,
		task.GetInfo().(*persistence.CrossClusterTaskInfo),
		task.request.StartChildExecutionAttributes.InitiatedEventAttributes,
		task.request.StartChildExecutionAttributes.GetRequestID(),
		s.mockShard.GetTimeSource().Now(),
	)
	targetRunID := "random target runID"
	s.mockHistoryClient.EXPECT().StartWorkflowExecution(gomock.Any(), historyReq).
		Return(&types.StartWorkflowExecutionResponse{RunID: targetRunID}, nil).Times(1)

	err := s.executor.Execute(task, true)
	s.NoError(err)

	s.Equal(task.GetTaskID(), task.response.GetTaskID())
	s.Equal(types.CrossClusterTaskTypeStartChildExecution, task.response.GetTaskType())
	s.Nil(task.response.FailedCause)
	s.NotNil(task.response.StartChildExecutionAttributes)
	s.Equal(targetRunID, task.response.StartChildExecutionAttributes.GetRunID())
}

func (s *crossClusterTargetTaskExecutorSuite) TestStartChildExecutionTask_StartChildFailed() {
	task := s.getTestStartChildExecutionTask(processingStateInitialized, nil)

	historyReq := createTestChildWorkflowExecutionRequest(
		constants.TestDomainName,
		constants.TestTargetDomainName,
		task.GetInfo().(*persistence.CrossClusterTaskInfo),
		task.request.StartChildExecutionAttributes.InitiatedEventAttributes,
		task.request.StartChildExecutionAttributes.GetRequestID(),
		s.mockShard.GetTimeSource().Now(),
	)
	s.mockHistoryClient.EXPECT().StartWorkflowExecution(gomock.Any(), historyReq).
		Return(nil, &types.WorkflowExecutionAlreadyStartedError{}).Times(1)

	err := s.executor.Execute(task, true)
	s.NoError(err)

	s.Equal(task.GetTaskID(), task.response.GetTaskID())
	s.Equal(types.CrossClusterTaskTypeStartChildExecution, task.response.GetTaskType())
	s.Equal(types.CrossClusterTaskFailedCauseWorkflowAlreadyRunning, task.response.GetFailedCause())
}

func (s *crossClusterTargetTaskExecutorSuite) TestStartChildExecutionTask_ScheduleDecisionSuccess() {
	targetRunID := "random target runID"
	task := s.getTestStartChildExecutionTask(processingStateResponseRecorded, &targetRunID)

	taskInfo := task.GetInfo().(*persistence.CrossClusterTaskInfo)
	s.mockHistoryClient.EXPECT().ScheduleDecisionTask(gomock.Any(), &types.ScheduleDecisionTaskRequest{
		DomainUUID: taskInfo.TargetDomainID,
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: taskInfo.TargetWorkflowID,
			RunID:      targetRunID,
		},
		IsFirstDecision: true,
	}).Return(nil).Times(1)

	err := s.executor.Execute(task, true)
	s.NoError(err)

	s.Equal(task.GetTaskID(), task.response.GetTaskID())
	s.Equal(types.CrossClusterTaskTypeStartChildExecution, task.response.GetTaskType())
	s.Nil(task.response.FailedCause)
	s.NotNil(task.response.StartChildExecutionAttributes)
}

func (s *crossClusterTargetTaskExecutorSuite) TestStartChildExecutionTask_ScheduleDecisionFailed() {
	targetRunID := "random target runID"
	task := s.getTestStartChildExecutionTask(processingStateResponseRecorded, &targetRunID)

	taskInfo := task.GetInfo().(*persistence.CrossClusterTaskInfo)
	s.mockHistoryClient.EXPECT().ScheduleDecisionTask(gomock.Any(), &types.ScheduleDecisionTaskRequest{
		DomainUUID: taskInfo.TargetDomainID,
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: taskInfo.TargetWorkflowID,
			RunID:      targetRunID,
		},
		IsFirstDecision: true,
	}).Return(&types.EntityNotExistsError{}).Times(1)

	err := s.executor.Execute(task, true)
	s.NoError(err)

	s.Equal(task.GetTaskID(), task.response.GetTaskID())
	s.Equal(types.CrossClusterTaskTypeStartChildExecution, task.response.GetTaskType())
	s.Equal(types.CrossClusterTaskFailedCauseWorkflowNotExists, task.response.GetFailedCause())
}

func (s *crossClusterTargetTaskExecutorSuite) TestCancelExecutionTask() {
	task := s.getTestCancelExecutionTask(processingStateInitialized)

	cancelRequest := createTestRequestCancelWorkflowExecutionRequest(
		constants.TestTargetDomainName,
		task.GetInfo().(*persistence.CrossClusterTaskInfo),
		task.request.CancelExecutionAttributes.RequestID,
	)
	s.mockHistoryClient.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), cancelRequest).Return(nil).Times(1)

	err := s.executor.Execute(task, true)
	s.NoError(err)

	s.Equal(task.GetTaskID(), task.response.GetTaskID())
	s.Equal(types.CrossClusterTaskTypeCancelExecution, task.response.GetTaskType())
	s.Nil(task.response.FailedCause)
	s.NotNil(task.response.CancelExecutionAttributes)
}

func (s *crossClusterTargetTaskExecutorSuite) TestSignalExecutionTask_SignalSuccess() {
	task := s.getTestSignalExecutionTask(processingStateInitialized)

	attributes := task.request.SignalExecutionAttributes
	signalRequest := createTestSignalWorkflowExecutionRequest(
		constants.TestTargetDomainName,
		task.GetInfo().(*persistence.CrossClusterTaskInfo),
		&persistence.SignalInfo{
			InitiatedID:     attributes.InitiatedEventID,
			SignalName:      attributes.SignalName,
			Input:           attributes.SignalInput,
			SignalRequestID: attributes.RequestID,
			Control:         attributes.Control,
		},
	)
	s.mockHistoryClient.EXPECT().SignalWorkflowExecution(gomock.Any(), signalRequest).Return(nil).Times(1)

	err := s.executor.Execute(task, true)
	s.NoError(err)

	s.Equal(task.GetTaskID(), task.response.GetTaskID())
	s.Equal(types.CrossClusterTaskTypeSignalExecution, task.response.GetTaskType())
	s.Nil(task.response.FailedCause)
	s.NotNil(task.response.SignalExecutionAttributes)
}

func (s *crossClusterTargetTaskExecutorSuite) TestSignalExecutionTask_SignalFailed() {
	task := s.getTestSignalExecutionTask(processingStateInitialized)

	attributes := task.request.SignalExecutionAttributes
	signalRequest := createTestSignalWorkflowExecutionRequest(
		constants.TestTargetDomainName,
		task.GetInfo().(*persistence.CrossClusterTaskInfo),
		&persistence.SignalInfo{
			InitiatedID:     attributes.InitiatedEventID,
			SignalName:      attributes.SignalName,
			Input:           attributes.SignalInput,
			SignalRequestID: attributes.RequestID,
			Control:         attributes.Control,
		},
	)
	s.mockHistoryClient.EXPECT().SignalWorkflowExecution(gomock.Any(), signalRequest).Return(&types.EntityNotExistsError{}).Times(1)

	err := s.executor.Execute(task, true)
	s.NoError(err)

	s.Equal(task.GetTaskID(), task.response.GetTaskID())
	s.Equal(types.CrossClusterTaskTypeSignalExecution, task.response.GetTaskType())
	s.Equal(types.CrossClusterTaskFailedCauseWorkflowNotExists, task.response.GetFailedCause())
}

func (s *crossClusterTargetTaskExecutorSuite) TestSignalExecutionTask_RemoveSignalMutableStateSuccess() {
	task := s.getTestSignalExecutionTask(processingStateResponseRecorded)

	taskInfo := task.GetInfo().(*persistence.CrossClusterTaskInfo)
	s.mockHistoryClient.EXPECT().RemoveSignalMutableState(gomock.Any(), &types.RemoveSignalMutableStateRequest{
		DomainUUID: taskInfo.TargetDomainID,
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: taskInfo.TargetWorkflowID,
			RunID:      taskInfo.TargetRunID,
		},
		RequestID: task.request.SignalExecutionAttributes.RequestID,
	}).Return(&types.EntityNotExistsError{}).Times(1)

	err := s.executor.Execute(task, true)
	s.NoError(err)

	s.Equal(task.GetTaskID(), task.response.GetTaskID())
	s.Equal(types.CrossClusterTaskTypeSignalExecution, task.response.GetTaskType())
	s.Nil(task.response.FailedCause)
	s.NotNil(task.response.SignalExecutionAttributes)
}

func (s *crossClusterTargetTaskExecutorSuite) getTestStartChildExecutionTask(
	processingState processingState,
	targetRunID *string,
) *crossClusterTargetTask {
	return s.getTestCrossClusterTargetTask(
		types.CrossClusterTaskTypeStartChildExecution,
		processingState,
		&types.CrossClusterStartChildExecutionRequestAttributes{
			TargetDomainID:   constants.TestTargetDomainID,
			RequestID:        "some random request ID",
			InitiatedEventID: int64(123),
			InitiatedEventAttributes: &types.StartChildWorkflowExecutionInitiatedEventAttributes{
				Domain:     constants.TestTargetDomainName,
				WorkflowID: "random target workflowID",
				WorkflowType: &types.WorkflowType{
					Name: "random workflow type",
				},
			},
			TargetRunID: targetRunID,
		},
		nil,
		nil,
	)
}

func (s *crossClusterTargetTaskExecutorSuite) getTestCancelExecutionTask(
	processingState processingState,
) *crossClusterTargetTask {
	return s.getTestCrossClusterTargetTask(
		types.CrossClusterTaskTypeCancelExecution,
		processingState,
		nil,
		&types.CrossClusterCancelExecutionRequestAttributes{
			TargetDomainID:    constants.TestTargetDomainID,
			TargetWorkflowID:  "some random target workflowID",
			TargetRunID:       "some random target runID",
			RequestID:         "some random request ID",
			InitiatedEventID:  int64(123),
			ChildWorkflowOnly: false,
		},
		nil,
	)
}

func (s *crossClusterTargetTaskExecutorSuite) getTestSignalExecutionTask(
	processingState processingState,
) *crossClusterTargetTask {
	return s.getTestCrossClusterTargetTask(
		types.CrossClusterTaskTypeSignalExecution,
		processingState,
		nil,
		nil,
		&types.CrossClusterSignalExecutionRequestAttributes{
			TargetDomainID:    constants.TestTargetDomainID,
			TargetWorkflowID:  "some random target workflowID",
			TargetRunID:       "some random target runID",
			RequestID:         "some random request ID",
			InitiatedEventID:  int64(123),
			ChildWorkflowOnly: false,
			SignalName:        "some random signal name",
			SignalInput:       []byte("some random signal input"),
			Control:           []byte("some random control"),
		},
	)
}

func (s *crossClusterTargetTaskExecutorSuite) getTestCrossClusterTargetTask(
	taskType types.CrossClusterTaskType,
	processingState processingState,
	startChildAttributes *types.CrossClusterStartChildExecutionRequestAttributes,
	cancelAttributes *types.CrossClusterCancelExecutionRequestAttributes,
	signalAttributes *types.CrossClusterSignalExecutionRequestAttributes,
) *crossClusterTargetTask {
	task, _ := NewCrossClusterTargetTask(
		s.mockShard,
		&types.CrossClusterTaskRequest{
			TaskInfo: &types.CrossClusterTaskInfo{
				DomainID:            constants.TestDomainID,
				WorkflowID:          constants.TestWorkflowID,
				RunID:               constants.TestRunID,
				TaskType:            taskType.Ptr(),
				TaskState:           int16(processingState),
				TaskID:              int64(1234),
				VisibilityTimestamp: common.Int64Ptr(time.Now().UnixNano()),
			},
			StartChildExecutionAttributes: startChildAttributes,
			CancelExecutionAttributes:     cancelAttributes,
			SignalExecutionAttributes:     signalAttributes,
		},
		s.executor,
		nil,
		nil,
		nil,
		dynamicconfig.GetIntPropertyFn(1),
	)
	return task.(*crossClusterTargetTask)
}
