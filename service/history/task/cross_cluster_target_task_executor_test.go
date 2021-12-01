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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/yarpc"

	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
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
		mockDomainCache   *cache.MockDomainCache

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
	s.mockDomainCache = s.mockShard.Resource.DomainCache
	s.mockDomainCache.EXPECT().GetDomainName(constants.TestDomainID).Return(constants.TestDomainName, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestTargetDomainID).Return(constants.TestGlobalTargetDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestRemoteTargetDomainID).Return(constants.TestGlobalRemoteTargetDomainEntry, nil).AnyTimes()
	mockClusterMetadata := s.mockShard.Resource.ClusterMetadata
	mockClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	mockClusterMetadata.EXPECT().GetAllClusterInfo().Return(cluster.TestAllClusterInfo).AnyTimes()
	mockClusterMetadata.EXPECT().IsGlobalDomainEnabled().Return(true).AnyTimes()

	s.executor = NewCrossClusterTargetTaskExecutor(s.mockShard, s.mockShard.GetLogger(), config)
}

func (s *crossClusterTargetTaskExecutorSuite) TearDownTest() {
	s.controller.Finish()
	s.mockShard.Finish(s.T())
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

func (s *crossClusterTargetTaskExecutorSuite) TestCancelExecutionTask_CancelSuccess() {
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

func (s *crossClusterTargetTaskExecutorSuite) TestCancelExecutionTask_CancelFailed() {
	task := s.getTestCancelExecutionTask(processingStateInitialized)

	cancelRequest := createTestRequestCancelWorkflowExecutionRequest(
		constants.TestTargetDomainName,
		task.GetInfo().(*persistence.CrossClusterTaskInfo),
		task.request.CancelExecutionAttributes.RequestID,
	)
	// domain failovered after the domain active check,
	// but before we make the requestCancel call
	s.mockHistoryClient.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), cancelRequest).Return(&types.DomainNotActiveError{}).Times(1)

	err := s.executor.Execute(task, true)
	s.NoError(err)

	s.Equal(task.GetTaskID(), task.response.GetTaskID())
	s.Equal(types.CrossClusterTaskTypeCancelExecution, task.response.GetTaskType())
	s.Equal(types.CrossClusterTaskFailedCauseDomainNotActive, task.response.GetFailedCause())
	s.Nil(task.response.CancelExecutionAttributes)
}

func (s *crossClusterTargetTaskExecutorSuite) TestApplyParentPolicyTask() {
	task := s.getTestApplyParentPolicyTask(processingStateInitialized, types.ParentClosePolicyRequestCancel)

	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestDomainID).Return(constants.TestGlobalDomainEntry, nil).Times(1)
	s.mockHistoryClient.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), gomock.Any()).Do(func(
		ctx context.Context,
		request *types.HistoryRequestCancelWorkflowExecutionRequest,
		opts ...yarpc.CallOption,
	) {
		s.Equal(request.DomainUUID, constants.TestDomainID)
	}).Return(nil).Times(1)

	err := s.executor.Execute(task, true)
	s.NoError(err)
	s.Equal(task.GetTaskID(), task.response.GetTaskID())
	s.Equal(types.CrossClusterTaskTypeApplyParentPolicy, task.response.GetTaskType())
	s.Nil(task.response.FailedCause)
	s.NotNil(task.response.ApplyParentClosePolicyAttributes)
}

func (s *crossClusterTargetTaskExecutorSuite) TestApplyParentPolicyTaskWithRetriableFailures() {
	expectedFailedCause := types.CrossClusterTaskFailedCauseDomainNotActive.Ptr()
	s.testApplyParentPolicyTaskWithFailures(ErrTaskPendingActive, expectedFailedCause, true, 2)
}
func (s *crossClusterTargetTaskExecutorSuite) TestApplyParentPolicyTaskWithNonRetriableFailures() {
	expectedFailedCause := types.CrossClusterTaskFailedCauseDomainNotExists.Ptr()
	s.testApplyParentPolicyTaskWithFailures(errDomainNotExists, expectedFailedCause, false, 2)
}

func (s *crossClusterTargetTaskExecutorSuite) testApplyParentPolicyTaskWithFailures(
	applyPolicyError error,
	expectedFailedCause *types.CrossClusterTaskFailedCause,
	retriable bool,
	numExpectedChildrenInRequest int,
) {
	task := s.getTestApplyParentPolicyTask(processingStateInitialized, types.ParentClosePolicyRequestCancel)
	otherDomainID := fmt.Sprintf("%s-%d", constants.TestDomainID, 1)
	policy := types.ParentClosePolicyRequestCancel
	task.request.ApplyParentClosePolicyAttributes.Children = append(
		task.request.ApplyParentClosePolicyAttributes.Children,
		&types.ApplyParentClosePolicyRequest{
			Child: &types.ApplyParentClosePolicyAttributes{
				ChildDomainID:     otherDomainID,
				ChildWorkflowID:   "some random workflow id",
				ChildRunID:        "some random run id",
				ParentClosePolicy: &policy,
			},
			Status: &types.ApplyParentClosePolicyStatus{
				Completed: false,
			},
		},
	)

	s.mockDomainCache.EXPECT().GetDomainByID(otherDomainID).Return(constants.TestGlobalDomainEntry, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainByID(constants.TestDomainID).Return(constants.TestGlobalDomainEntry, nil).AnyTimes()

	cancelRequest := func(domainID string) *types.HistoryRequestCancelWorkflowExecutionRequest {
		return &types.HistoryRequestCancelWorkflowExecutionRequest{
			DomainUUID:               domainID,
			ExternalInitiatedEventID: nil,
			ExternalWorkflowExecution: &types.WorkflowExecution{
				WorkflowID: constants.TestWorkflowID,
				RunID:      constants.TestRunID,
			},
			ChildWorkflowOnly: true,
			CancelRequest: &types.RequestCancelWorkflowExecutionRequest{
				Domain: "some random domain name",
				WorkflowExecution: &types.WorkflowExecution{
					WorkflowID: "some random workflow id",
					RunID:      "some random run id",
				},
				Identity:  "history-service",
				RequestID: "",
			},
		}
	}
	s.mockHistoryClient.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), cancelRequest(constants.TestDomainID)).Return(nil).Times(1)
	s.mockHistoryClient.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), cancelRequest(otherDomainID)).Return(applyPolicyError).Times(1)

	err := s.executor.Execute(task, true)
	if retriable {
		s.Equal(applyPolicyError, err)
	} else {
		s.Nil(err)
	}

	s.Equal(task.GetTaskID(), task.response.GetTaskID())
	s.Equal(types.CrossClusterTaskTypeApplyParentPolicy, task.response.GetTaskType())
	s.NotNil(task.response.ApplyParentClosePolicyAttributes)
	s.Equal(2, len(task.response.ApplyParentClosePolicyAttributes.ChildrenStatus))
	for _, childStatus := range task.response.ApplyParentClosePolicyAttributes.ChildrenStatus {
		if childStatus.Child.ChildDomainID == constants.TestDomainID {
			s.Nil(childStatus.FailedCause)
		} else if childStatus.Child.ChildDomainID == otherDomainID {
			s.Equal(expectedFailedCause, childStatus.FailedCause)
		} else {
			panic(fmt.Sprintf("unexpected domain id: %v", childStatus.Child.ChildDomainID))
		}
	}

	// only one of those tasks failed, so we only have one of them in the retry request
	s.Equal(
		numExpectedChildrenInRequest,
		len(task.request.ApplyParentClosePolicyAttributes.Children),
	)
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

func (s *crossClusterTargetTaskExecutorSuite) TestApplyParentClosePolicyTask_Success() {
	task := s.getTestApplyParentClosePolicyTask(processingStateInitialized)

	taskInfo := task.GetInfo().(*persistence.CrossClusterTaskInfo)
	for _, child := range task.request.ApplyParentClosePolicyAttributes.Children {
		childAttr := child.Child
		switch *childAttr.GetParentClosePolicy() {
		case types.ParentClosePolicyRequestCancel:
			s.mockHistoryClient.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, request *types.HistoryRequestCancelWorkflowExecutionRequest, option ...yarpc.CallOption) error {
					s.Equal(childAttr.ChildDomainID, request.GetDomainUUID())
					s.Equal(childAttr.ChildWorkflowID, request.GetCancelRequest().GetWorkflowExecution().GetWorkflowID())
					s.Equal(childAttr.ChildRunID, request.GetCancelRequest().GetWorkflowExecution().GetRunID())
					s.True(request.GetChildWorkflowOnly())
					s.Equal(taskInfo.GetWorkflowID(), request.GetExternalWorkflowExecution().GetWorkflowID())
					s.Equal(taskInfo.GetRunID(), request.GetExternalWorkflowExecution().GetRunID())
					return nil
				}).Times(1)
		case types.ParentClosePolicyTerminate:
			s.mockHistoryClient.EXPECT().TerminateWorkflowExecution(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, request *types.HistoryTerminateWorkflowExecutionRequest, option ...yarpc.CallOption) error {
					s.Equal(childAttr.ChildDomainID, request.GetDomainUUID())
					s.Equal(childAttr.ChildWorkflowID, request.GetTerminateRequest().GetWorkflowExecution().GetWorkflowID())
					s.Equal(childAttr.ChildRunID, request.GetTerminateRequest().GetWorkflowExecution().GetRunID())
					s.True(request.GetChildWorkflowOnly())
					s.Equal(taskInfo.GetWorkflowID(), request.GetExternalWorkflowExecution().GetWorkflowID())
					s.Equal(taskInfo.GetRunID(), request.GetExternalWorkflowExecution().GetRunID())
					return nil
				}).Times(1)
		}
	}

	err := s.executor.Execute(task, true)
	s.NoError(err)

	s.Equal(task.GetTaskID(), task.response.GetTaskID())
	s.Equal(types.CrossClusterTaskTypeApplyParentPolicy, task.response.GetTaskType())
	s.Nil(task.response.FailedCause)
	s.NotNil(task.response.ApplyParentClosePolicyAttributes)
}

func (s *crossClusterTargetTaskExecutorSuite) TestApplyParentClosePolicyTask_Failed() {
	task := s.getTestApplyParentClosePolicyTask(processingStateInitialized)

	s.mockHistoryClient.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), gomock.Any()).Return(&types.DomainNotActiveError{}).AnyTimes()
	s.mockHistoryClient.EXPECT().TerminateWorkflowExecution(gomock.Any(), gomock.Any()).Return(&types.DomainNotActiveError{}).AnyTimes()

	err := s.executor.Execute(task, true)
	s.NoError(err)

	s.Equal(task.GetTaskID(), task.response.GetTaskID())
	s.Equal(types.CrossClusterTaskTypeApplyParentPolicy, task.response.GetTaskType())
	s.Equal(types.CrossClusterTaskFailedCauseDomainNotActive, task.response.GetFailedCause())
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
		nil,
		nil,
	)
}

func (s *crossClusterTargetTaskExecutorSuite) getTestApplyParentPolicyTask(
	processingState processingState,
	policy types.ParentClosePolicy,
) *crossClusterTargetTask {
	return s.getTestCrossClusterTargetTask(
		types.CrossClusterTaskTypeApplyParentPolicy,
		processingState,
		nil,
		nil,
		nil,
		nil,
		&types.CrossClusterApplyParentClosePolicyRequestAttributes{
			Children: []*types.ApplyParentClosePolicyRequest{
				{
					Child: &types.ApplyParentClosePolicyAttributes{
						ChildDomainID:     constants.TestDomainID,
						ChildWorkflowID:   "some random workflow id",
						ChildRunID:        "some random run id",
						ParentClosePolicy: &policy,
					},
					Status: &types.ApplyParentClosePolicyStatus{
						Completed: false,
					},
				},
			},
		},
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
		nil,
		nil,
	)
}

func (s *crossClusterTargetTaskExecutorSuite) getTestApplyParentClosePolicyTask(
	processingState processingState,
) *crossClusterTargetTask {
	return s.getTestCrossClusterTargetTask(
		types.CrossClusterTaskTypeApplyParentPolicy,
		processingState,
		nil,
		nil,
		nil,
		nil,
		&types.CrossClusterApplyParentClosePolicyRequestAttributes{
			Children: []*types.ApplyParentClosePolicyRequest{
				{
					Child: &types.ApplyParentClosePolicyAttributes{
						ChildDomainID:     constants.TestTargetDomainID,
						ChildWorkflowID:   "some random target workflowID",
						ChildRunID:        "some random target runID",
						ParentClosePolicy: types.ParentClosePolicyAbandon.Ptr(),
					},
					Status: &types.ApplyParentClosePolicyStatus{
						Completed: false,
					},
				},
				{
					Child: &types.ApplyParentClosePolicyAttributes{
						ChildDomainID:     constants.TestTargetDomainID,
						ChildWorkflowID:   "some random target workflowID",
						ChildRunID:        "some random target runID",
						ParentClosePolicy: types.ParentClosePolicyRequestCancel.Ptr(),
					},
					Status: &types.ApplyParentClosePolicyStatus{
						Completed: false,
					},
				},
				{
					Child: &types.ApplyParentClosePolicyAttributes{
						ChildDomainID:     constants.TestTargetDomainID,
						ChildWorkflowID:   "some random target workflowID",
						ChildRunID:        "some random target runID",
						ParentClosePolicy: types.ParentClosePolicyTerminate.Ptr(),
					},
					Status: &types.ApplyParentClosePolicyStatus{
						Completed: false,
					},
				},
			},
		},
	)
}

func (s *crossClusterTargetTaskExecutorSuite) getTestCrossClusterTargetTask(
	taskType types.CrossClusterTaskType,
	processingState processingState,
	startChildAttributes *types.CrossClusterStartChildExecutionRequestAttributes,
	cancelAttributes *types.CrossClusterCancelExecutionRequestAttributes,
	signalAttributes *types.CrossClusterSignalExecutionRequestAttributes,
	recordChildCompletionAttributes *types.CrossClusterRecordChildWorkflowExecutionCompleteRequestAttributes,
	applyParentPolicyAttributes *types.CrossClusterApplyParentClosePolicyRequestAttributes,
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
			StartChildExecutionAttributes:                  startChildAttributes,
			CancelExecutionAttributes:                      cancelAttributes,
			SignalExecutionAttributes:                      signalAttributes,
			RecordChildWorkflowExecutionCompleteAttributes: recordChildCompletionAttributes,
			ApplyParentClosePolicyAttributes:               applyParentPolicyAttributes,
		},
		s.executor,
		nil,
		nil,
		nil,
		dynamicconfig.GetIntPropertyFn(1),
	)
	return task.(*crossClusterTargetTask)
}
