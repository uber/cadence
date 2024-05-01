// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package decision

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/testdata"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/execution"
)

const (
	testTaskCompletedID = int64(123)
)

func TestHandleDecisionRequestCancelExternalWorkflow(t *testing.T) {
	tests := []struct {
		name            string
		expectMockCalls func(taskHandler *taskHandlerImpl)
		attributes      *types.RequestCancelExternalWorkflowExecutionDecisionAttributes
		asserts         func(t *testing.T, taskHandler *taskHandlerImpl, err error)
	}{
		{
			name: "success",
			attributes: &types.RequestCancelExternalWorkflowExecutionDecisionAttributes{
				Domain:            constants.TestDomainName,
				WorkflowID:        constants.TestWorkflowID,
				RunID:             constants.TestRunID,
				Control:           nil,
				ChildWorkflowOnly: false,
			},
			expectMockCalls: func(taskHandler *taskHandlerImpl) {
				taskHandler.domainCache.(*cache.MockDomainCache).EXPECT().GetDomain(constants.TestDomainName).Return(constants.TestLocalDomainEntry, nil)
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, err error) {
				assert.False(t, taskHandler.failDecision)
				assert.Empty(t, taskHandler.failMessage)
				assert.Nil(t, taskHandler.failDecisionCause)
				assert.Equal(t, nil, err)

			},
		},
		{
			name: "internal service error",
			attributes: &types.RequestCancelExternalWorkflowExecutionDecisionAttributes{
				Domain:            constants.TestDomainName,
				WorkflowID:        constants.TestWorkflowID,
				RunID:             constants.TestRunID,
				Control:           nil,
				ChildWorkflowOnly: false,
			},
			expectMockCalls: func(taskHandler *taskHandlerImpl) {
				taskHandler.domainCache.(*cache.MockDomainCache).EXPECT().GetDomain(constants.TestDomainName).Return(nil, errors.New("some error getting domain cache"))
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, err error) {
				assert.Equal(t, &types.InternalServiceError{Message: "Unable to cancel workflow across domain: some random domain name."}, err)
			},
		},
		{
			name: "attributes validation failure",
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, err error) {
				assert.True(t, taskHandler.failDecision)
				assert.NotNil(t, taskHandler.failMessage)
				assert.Equal(t, func(i int32) *types.DecisionTaskFailedCause {
					cause := new(types.DecisionTaskFailedCause)
					*cause = types.DecisionTaskFailedCause(i)
					return cause
					// 9 is types.DecisionTaskFailedCause "BAD_REQUEST_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_ATTRIBUTES"
				}(9), taskHandler.failDecisionCause)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			taskHandler := newTaskHandlerForTest(t)
			taskHandler.mutableState.(*execution.MockMutableState).EXPECT().GetExecutionInfo().Times(1).Return(&persistence.WorkflowExecutionInfo{
				DomainID:   constants.TestDomainID,
				WorkflowID: constants.TestWorkflowID,
				RunID:      constants.TestRunID,
			})
			taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddRequestCancelExternalWorkflowExecutionInitiatedEvent(testTaskCompletedID, gomock.Any(), test.attributes).AnyTimes()
			if test.expectMockCalls != nil {
				test.expectMockCalls(taskHandler)
			}
			decision := &types.Decision{
				DecisionType: func(i int32) *types.DecisionType {
					decisionType := new(types.DecisionType)
					*decisionType = types.DecisionType(i)
					return decisionType
				}(7), //types.DecisionTypeRequestCancelExternalWorkflowExecution == 7
				RequestCancelExternalWorkflowExecutionDecisionAttributes: test.attributes,
			}
			err := taskHandler.handleDecision(context.Background(), decision)
			test.asserts(t, taskHandler, err)
		})
	}
}

func TestHandleDecisionRequestCancelActivity(t *testing.T) {
	tests := []struct {
		name            string
		expectMockCalls func(taskHandler *taskHandlerImpl)
		attributes      *types.RequestCancelActivityTaskDecisionAttributes
		asserts         func(t *testing.T, taskHandler *taskHandlerImpl, err error)
	}{
		{
			name:       "success",
			attributes: &types.RequestCancelActivityTaskDecisionAttributes{ActivityID: testdata.ActivityID},
			expectMockCalls: func(taskHandler *taskHandlerImpl) {
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddActivityTaskCancelRequestedEvent(
					testTaskCompletedID,
					testdata.ActivityID,
					testdata.Identity,
				).Times(1).Return(&types.HistoryEvent{}, &persistence.ActivityInfo{StartedID: common.EmptyEventID}, nil)
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddActivityTaskCanceledEvent(int64(0), int64(-23), int64(0), []byte(activityCancellationMsgActivityNotStarted), testdata.Identity).Return(nil, nil)
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, err error) {
				assert.False(t, taskHandler.failDecision)
				assert.Empty(t, taskHandler.failMessage)
				assert.Nil(t, taskHandler.failDecisionCause)
				assert.Equal(t, nil, err)

			},
		},
		{
			name:       "AddActivityTaskCanceledEvent failure",
			attributes: &types.RequestCancelActivityTaskDecisionAttributes{ActivityID: testdata.ActivityID},
			expectMockCalls: func(taskHandler *taskHandlerImpl) {
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddActivityTaskCancelRequestedEvent(
					testTaskCompletedID,
					testdata.ActivityID,
					testdata.Identity,
				).Times(1).Return(&types.HistoryEvent{}, &persistence.ActivityInfo{StartedID: common.EmptyEventID}, nil)
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddActivityTaskCanceledEvent(int64(0), int64(-23), int64(0), []byte(activityCancellationMsgActivityNotStarted), testdata.Identity).Return(nil, errors.New("some random error"))
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, err error) {
				assert.Equal(t, errors.New("some random error"), err)
			},
		},
		{
			name:       "AddRequestCancelActivityTaskFailedEvent failure",
			attributes: &types.RequestCancelActivityTaskDecisionAttributes{ActivityID: testdata.ActivityID},
			expectMockCalls: func(taskHandler *taskHandlerImpl) {
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddActivityTaskCancelRequestedEvent(
					testTaskCompletedID,
					testdata.ActivityID,
					testdata.Identity,
				).Times(1).Return(&types.HistoryEvent{}, &persistence.ActivityInfo{StartedID: common.EmptyEventID}, &types.BadRequestError{Message: "some types.BadRequestError error"})
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddRequestCancelActivityTaskFailedEvent(testTaskCompletedID, testdata.ActivityID, activityCancellationMsgActivityIDUnknown).Return(nil, errors.New("some random error"))
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, err error) {
				assert.Equal(t, errors.New("some random error"), err)
			},
		},
		{
			name:       "default switch case error",
			attributes: &types.RequestCancelActivityTaskDecisionAttributes{ActivityID: testdata.ActivityID},
			expectMockCalls: func(taskHandler *taskHandlerImpl) {
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddActivityTaskCancelRequestedEvent(
					testTaskCompletedID,
					testdata.ActivityID,
					testdata.Identity,
				).Times(1).Return(&types.HistoryEvent{}, &persistence.ActivityInfo{StartedID: common.EmptyEventID}, errors.New("some default error"))
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, err error) {
				assert.Equal(t, errors.New("some default error"), err)
			},
		},
		{
			name: "attributes validation failure",
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, err error) {
				assert.True(t, taskHandler.failDecision)
				assert.NotNil(t, taskHandler.failMessage)
				assert.Equal(t, func(i int32) *types.DecisionTaskFailedCause {
					cause := new(types.DecisionTaskFailedCause)
					*cause = types.DecisionTaskFailedCause(i)
					return cause
					// 2 is types.DecisionTaskFailedCause "BAD_REQUEST_CANCEL_ACTIVITY_ATTRIBUTES"
				}(2), taskHandler.failDecisionCause)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			taskHandler := newTaskHandlerForTest(t)
			taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddRequestCancelExternalWorkflowExecutionInitiatedEvent(testTaskCompletedID, gomock.Any(), test.attributes).AnyTimes()
			if test.expectMockCalls != nil {
				test.expectMockCalls(taskHandler)
			}
			decision := &types.Decision{
				DecisionType: func(i int32) *types.DecisionType {
					decisionType := new(types.DecisionType)
					*decisionType = types.DecisionType(i)
					return decisionType
				}(1), //types.DecisionTypeRequestCancelActivityTask == 1
				RequestCancelActivityTaskDecisionAttributes: test.attributes,
			}
			err := taskHandler.handleDecision(context.Background(), decision)
			test.asserts(t, taskHandler, err)
		})
	}
}

func TestHandleDecisionStartChildWorkflow(t *testing.T) {

	tests := []struct {
		name            string
		expectMockCalls func(taskHandler *taskHandlerImpl, attr *types.StartChildWorkflowExecutionDecisionAttributes)
		attributes      *types.StartChildWorkflowExecutionDecisionAttributes
		asserts         func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.StartChildWorkflowExecutionDecisionAttributes, err error)
	}{
		{
			name: "success - ParentClosePolicy enabled",
			attributes: &types.StartChildWorkflowExecutionDecisionAttributes{
				Domain:       constants.TestDomainName,
				WorkflowID:   constants.TestWorkflowID,
				WorkflowType: &types.WorkflowType{Name: testdata.WorkflowTypeName},
				TaskList:     &types.TaskList{Name: testdata.TaskListName},
			},
			expectMockCalls: func(taskHandler *taskHandlerImpl, attr *types.StartChildWorkflowExecutionDecisionAttributes) {
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddStartChildWorkflowExecutionInitiatedEvent(testTaskCompletedID, gomock.Any(), attr).Times(1)
				taskHandler.domainCache.(*cache.MockDomainCache).EXPECT().GetDomain(constants.TestDomainName).Return(constants.TestLocalDomainEntry, nil)
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.StartChildWorkflowExecutionDecisionAttributes, err error) {
				assert.Equal(t, types.ParentClosePolicyTerminate.Ptr(), attr.ParentClosePolicy)
				assert.False(t, taskHandler.failDecision)
				assert.Empty(t, taskHandler.failMessage)
				assert.Nil(t, taskHandler.failDecisionCause)
				assert.Equal(t, nil, err)

			},
		},
		{
			name: "success - ParentClosePolicy disabled",
			attributes: &types.StartChildWorkflowExecutionDecisionAttributes{
				Domain:       constants.TestDomainName,
				WorkflowID:   constants.TestWorkflowID,
				WorkflowType: &types.WorkflowType{Name: testdata.WorkflowTypeName},
				TaskList:     &types.TaskList{Name: testdata.TaskListName},
			},
			expectMockCalls: func(taskHandler *taskHandlerImpl, attr *types.StartChildWorkflowExecutionDecisionAttributes) {
				taskHandler.config.EnableParentClosePolicy = func(domain string) bool { return false }
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddStartChildWorkflowExecutionInitiatedEvent(testTaskCompletedID, gomock.Any(), attr).Times(1)
				taskHandler.domainCache.(*cache.MockDomainCache).EXPECT().GetDomain(constants.TestDomainName).Return(constants.TestLocalDomainEntry, nil)
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.StartChildWorkflowExecutionDecisionAttributes, err error) {
				assert.Equal(t, types.ParentClosePolicyAbandon.Ptr(), attr.ParentClosePolicy)
				assert.False(t, taskHandler.failDecision)
				assert.Empty(t, taskHandler.failMessage)
				assert.Nil(t, taskHandler.failDecisionCause)
				assert.Equal(t, nil, err)

			},
		},
		{
			name: "success - ParentClosePolicy non nil",
			attributes: &types.StartChildWorkflowExecutionDecisionAttributes{
				Domain:            constants.TestDomainName,
				WorkflowID:        constants.TestWorkflowID,
				WorkflowType:      &types.WorkflowType{Name: testdata.WorkflowTypeName},
				TaskList:          &types.TaskList{Name: testdata.TaskListName},
				ParentClosePolicy: new(types.ParentClosePolicy),
			},
			expectMockCalls: func(taskHandler *taskHandlerImpl, attr *types.StartChildWorkflowExecutionDecisionAttributes) {
				taskHandler.config.EnableParentClosePolicy = func(domain string) bool { return false }
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddStartChildWorkflowExecutionInitiatedEvent(testTaskCompletedID, gomock.Any(), attr).Times(1)
				taskHandler.domainCache.(*cache.MockDomainCache).EXPECT().GetDomain(constants.TestDomainName).Return(constants.TestLocalDomainEntry, nil)

			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.StartChildWorkflowExecutionDecisionAttributes, err error) {
				assert.Equal(t, types.ParentClosePolicyAbandon.Ptr(), attr.ParentClosePolicy)
				assert.False(t, taskHandler.failDecision)
				assert.Empty(t, taskHandler.failMessage)
				assert.Nil(t, taskHandler.failDecisionCause)
				assert.Equal(t, nil, err)

			},
		},
		{
			name: "internal service error",
			attributes: &types.StartChildWorkflowExecutionDecisionAttributes{
				Domain:       constants.TestDomainName,
				WorkflowID:   constants.TestWorkflowID,
				WorkflowType: &types.WorkflowType{Name: testdata.WorkflowTypeName},
				TaskList:     &types.TaskList{Name: testdata.TaskListName},
			},
			expectMockCalls: func(taskHandler *taskHandlerImpl, attr *types.StartChildWorkflowExecutionDecisionAttributes) {
				taskHandler.domainCache.(*cache.MockDomainCache).EXPECT().GetDomain(constants.TestDomainName).Return(nil, errors.New("some error getting domain cache"))
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.StartChildWorkflowExecutionDecisionAttributes, err error) {
				assert.Equal(t, &types.InternalServiceError{Message: "Unable to schedule child execution across domain some random domain name."}, err)
			},
		},
		{
			name: "attributes validation failure",
			attributes: &types.StartChildWorkflowExecutionDecisionAttributes{
				Domain:       constants.TestDomainName,
				WorkflowID:   constants.TestWorkflowID,
				WorkflowType: &types.WorkflowType{Name: testdata.WorkflowTypeName},
			},
			expectMockCalls: func(taskHandler *taskHandlerImpl, attr *types.StartChildWorkflowExecutionDecisionAttributes) {
				taskHandler.domainCache.(*cache.MockDomainCache).EXPECT().GetDomain(constants.TestDomainName).Return(constants.TestLocalDomainEntry, nil)
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.StartChildWorkflowExecutionDecisionAttributes, err error) {
				assert.True(t, taskHandler.failDecision)
				assert.Equal(t, "missing task list name", *taskHandler.failMessage)
				assert.Equal(t, func(i int32) *types.DecisionTaskFailedCause {
					cause := new(types.DecisionTaskFailedCause)
					*cause = types.DecisionTaskFailedCause(i)
					return cause
					// 15 is types.DecisionTaskFailedCause "BAD_START_CHILD_EXECUTION_ATTRIBUTES"
				}(15), taskHandler.failDecisionCause)
				assert.Equal(t, nil, err)
			},
		},
		{
			name: "size limit checker failure",
			attributes: &types.StartChildWorkflowExecutionDecisionAttributes{
				Domain:       constants.TestDomainName,
				WorkflowID:   constants.TestWorkflowID,
				WorkflowType: &types.WorkflowType{Name: testdata.WorkflowTypeName},
				TaskList:     &types.TaskList{Name: testdata.TaskListName},
				Input:        []byte("input"),
			},
			expectMockCalls: func(taskHandler *taskHandlerImpl, attr *types.StartChildWorkflowExecutionDecisionAttributes) {
				taskHandler.sizeLimitChecker.blobSizeLimitError = 1
				taskHandler.sizeLimitChecker.blobSizeLimitWarn = 2
				taskHandler.domainCache.(*cache.MockDomainCache).EXPECT().GetDomain(constants.TestDomainName).Return(constants.TestLocalDomainEntry, nil)
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddFailWorkflowEvent(testTaskCompletedID, &types.FailWorkflowExecutionDecisionAttributes{
					Reason:  common.StringPtr(common.FailureReasonDecisionBlobSizeExceedsLimit),
					Details: []byte("StartChildWorkflowExecutionDecisionAttributes.Input exceeds size limit."),
				}).Return(nil, errors.New("some error adding fail workflow event"))
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.StartChildWorkflowExecutionDecisionAttributes, err error) {
				assert.Equal(t, errors.New("some error adding fail workflow event"), err)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			taskHandler := newTaskHandlerForTest(t)
			taskHandler.mutableState.(*execution.MockMutableState).EXPECT().GetExecutionInfo().AnyTimes().Return(&persistence.WorkflowExecutionInfo{
				DomainID:   constants.TestDomainID,
				WorkflowID: constants.TestWorkflowID,
				RunID:      constants.TestRunID,
			})
			if test.expectMockCalls != nil {
				test.expectMockCalls(taskHandler, test.attributes)
			}
			decision := &types.Decision{
				DecisionType: func(i int32) *types.DecisionType {
					decisionType := new(types.DecisionType)
					*decisionType = types.DecisionType(i)
					return decisionType
				}(10), //types.DecisionTypeStartChildWorkflowExecution == 10
				StartChildWorkflowExecutionDecisionAttributes: test.attributes,
			}
			err := taskHandler.handleDecision(context.Background(), decision)
			test.asserts(t, taskHandler, test.attributes, err)
		})
	}
}

func TestHandleDecisionCancelTimer(t *testing.T) {
	tests := []struct {
		name            string
		expectMockCalls func(taskHandler *taskHandlerImpl, attr *types.CancelTimerDecisionAttributes)
		attributes      *types.CancelTimerDecisionAttributes
		asserts         func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.CancelTimerDecisionAttributes, err error)
	}{
		{
			name:       "success",
			attributes: &types.CancelTimerDecisionAttributes{TimerID: "test-timer-id"},
			expectMockCalls: func(taskHandler *taskHandlerImpl, attr *types.CancelTimerDecisionAttributes) {
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddTimerCanceledEvent(testTaskCompletedID, attr, taskHandler.identity)
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().HasBufferedEvents()
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.CancelTimerDecisionAttributes, err error) {
				assert.False(t, taskHandler.failDecision)
				assert.Empty(t, taskHandler.failMessage)
				assert.Nil(t, taskHandler.failDecisionCause)
				assert.Equal(t, nil, err)
			},
		},
		{
			name:       "bad request error",
			attributes: &types.CancelTimerDecisionAttributes{TimerID: "test-timer-id"},
			expectMockCalls: func(taskHandler *taskHandlerImpl, attr *types.CancelTimerDecisionAttributes) {
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddTimerCanceledEvent(testTaskCompletedID, attr, taskHandler.identity).Return(nil, &types.BadRequestError{"some bad request error"})
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddCancelTimerFailedEvent(testTaskCompletedID, attr, taskHandler.identity)
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.CancelTimerDecisionAttributes, err error) {
				assert.False(t, taskHandler.failDecision)
				assert.Empty(t, taskHandler.failMessage)
				assert.Nil(t, taskHandler.failDecisionCause)
				assert.Equal(t, nil, err)
			},
		},
		{
			name:       "default error",
			attributes: &types.CancelTimerDecisionAttributes{TimerID: "test-timer-id"},
			expectMockCalls: func(taskHandler *taskHandlerImpl, attr *types.CancelTimerDecisionAttributes) {
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddTimerCanceledEvent(testTaskCompletedID, attr, taskHandler.identity).Return(nil, errors.New("some random error"))
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.CancelTimerDecisionAttributes, err error) {
				assert.False(t, taskHandler.failDecision)
				assert.Empty(t, taskHandler.failMessage)
				assert.Nil(t, taskHandler.failDecisionCause)
				assert.Equal(t, errors.New("some random error"), err)
			},
		},
		{
			name:       "attributes validation error",
			attributes: &types.CancelTimerDecisionAttributes{},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.CancelTimerDecisionAttributes, err error) {
				assert.True(t, taskHandler.failDecision)
				assert.Equal(t, "TimerId is not set on decision.", *taskHandler.failMessage)
				assert.Equal(t, func(i int32) *types.DecisionTaskFailedCause {
					cause := new(types.DecisionTaskFailedCause)
					*cause = types.DecisionTaskFailedCause(i)
					return cause
				}(4), taskHandler.failDecisionCause)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			taskHandler := newTaskHandlerForTest(t)
			if test.expectMockCalls != nil {
				test.expectMockCalls(taskHandler, test.attributes)
			}
			decision := &types.Decision{
				DecisionType: func(i int32) *types.DecisionType {
					decisionType := new(types.DecisionType)
					*decisionType = types.DecisionType(i)
					return decisionType
				}(5), //types.DecisionTypeCancelTimer
				CancelTimerDecisionAttributes: test.attributes,
			}
			err := taskHandler.handleDecision(context.Background(), decision)
			test.asserts(t, taskHandler, test.attributes, err)
		})
	}
}

func TestHandleDecisionSignalExternalWorkflow(t *testing.T) {
	tests := []struct {
		name            string
		expectMockCalls func(taskHandler *taskHandlerImpl, attr *types.SignalExternalWorkflowExecutionDecisionAttributes)
		attributes      *types.SignalExternalWorkflowExecutionDecisionAttributes
		asserts         func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.SignalExternalWorkflowExecutionDecisionAttributes, err error)
	}{
		{
			name: "success",
			attributes: &types.SignalExternalWorkflowExecutionDecisionAttributes{
				Domain:     constants.TestDomainName,
				Execution:  &types.WorkflowExecution{WorkflowID: constants.TestWorkflowID, RunID: constants.TestRunID},
				SignalName: testdata.SignalName,
				Input:      []byte("some input data"),
			},
			expectMockCalls: func(taskHandler *taskHandlerImpl, attr *types.SignalExternalWorkflowExecutionDecisionAttributes) {
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().GetExecutionInfo().Times(2).Return(&persistence.WorkflowExecutionInfo{
					DomainID:   constants.TestDomainID,
					WorkflowID: constants.TestWorkflowID,
					RunID:      constants.TestRunID,
				})
				taskHandler.domainCache.(*cache.MockDomainCache).EXPECT().GetDomain(constants.TestDomainName).Return(constants.TestLocalDomainEntry, nil)
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddSignalExternalWorkflowExecutionInitiatedEvent(testTaskCompletedID, gomock.Any(), attr)
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.SignalExternalWorkflowExecutionDecisionAttributes, err error) {
				assert.False(t, taskHandler.failDecision)
				assert.Empty(t, taskHandler.failMessage)
				assert.Nil(t, taskHandler.failDecisionCause)
				assert.Equal(t, nil, err)
			},
		},
		{
			name: "internal service error",
			attributes: &types.SignalExternalWorkflowExecutionDecisionAttributes{
				Domain:     constants.TestDomainName,
				Execution:  &types.WorkflowExecution{WorkflowID: constants.TestWorkflowID, RunID: constants.TestRunID},
				SignalName: testdata.SignalName,
				Input:      []byte("some input data"),
			},
			expectMockCalls: func(taskHandler *taskHandlerImpl, attr *types.SignalExternalWorkflowExecutionDecisionAttributes) {
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().GetExecutionInfo().Times(1).Return(&persistence.WorkflowExecutionInfo{
					DomainID:   constants.TestDomainID,
					WorkflowID: constants.TestWorkflowID,
					RunID:      constants.TestRunID,
				})
				taskHandler.domainCache.(*cache.MockDomainCache).EXPECT().GetDomain(constants.TestDomainName).Return(nil, errors.New("some error getting domain cache"))
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.SignalExternalWorkflowExecutionDecisionAttributes, err error) {
				assert.Equal(t, &types.InternalServiceError{Message: "Unable to signal workflow across domain: some random domain name."}, err)
			},
		},
		{
			name: "attributes validation failure",
			attributes: &types.SignalExternalWorkflowExecutionDecisionAttributes{
				Domain:    constants.TestDomainName,
				Execution: &types.WorkflowExecution{WorkflowID: constants.TestWorkflowID, RunID: constants.TestRunID},
				Input:     []byte("some input data"),
			},
			expectMockCalls: func(taskHandler *taskHandlerImpl, attr *types.SignalExternalWorkflowExecutionDecisionAttributes) {
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().GetExecutionInfo().Times(1).Return(&persistence.WorkflowExecutionInfo{
					DomainID:   constants.TestDomainID,
					WorkflowID: constants.TestWorkflowID,
					RunID:      constants.TestRunID,
				})
				taskHandler.domainCache.(*cache.MockDomainCache).EXPECT().GetDomain(constants.TestDomainName).Return(constants.TestLocalDomainEntry, nil)
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.SignalExternalWorkflowExecutionDecisionAttributes, err error) {
				assert.True(t, taskHandler.failDecision)
				assert.Equal(t, "SignalName is not set on decision.", *taskHandler.failMessage)
				assert.Equal(t, func(i int32) *types.DecisionTaskFailedCause {
					cause := new(types.DecisionTaskFailedCause)
					*cause = types.DecisionTaskFailedCause(i)
					return cause
					// 14 is types.DecisionTaskFailedCause "BAD_SIGNAL_WORKFLOW_EXECUTION_ATTRIBUTES"
				}(14), taskHandler.failDecisionCause)
				assert.Equal(t, nil, err)
			},
		},
		{
			name: "size limit checker failure",
			attributes: &types.SignalExternalWorkflowExecutionDecisionAttributes{
				Domain:     constants.TestDomainName,
				Execution:  &types.WorkflowExecution{WorkflowID: constants.TestWorkflowID, RunID: constants.TestRunID},
				SignalName: testdata.SignalName,
				Input:      []byte("some input data"),
			},
			expectMockCalls: func(taskHandler *taskHandlerImpl, attr *types.SignalExternalWorkflowExecutionDecisionAttributes) {
				taskHandler.sizeLimitChecker.blobSizeLimitError = 1
				taskHandler.sizeLimitChecker.blobSizeLimitWarn = 2
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().GetExecutionInfo().Times(2).Return(&persistence.WorkflowExecutionInfo{
					DomainID:   constants.TestDomainID,
					WorkflowID: constants.TestWorkflowID,
					RunID:      constants.TestRunID,
				})
				taskHandler.domainCache.(*cache.MockDomainCache).EXPECT().GetDomain(constants.TestDomainName).Return(constants.TestLocalDomainEntry, nil)
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddFailWorkflowEvent(testTaskCompletedID, &types.FailWorkflowExecutionDecisionAttributes{
					Reason:  common.StringPtr(common.FailureReasonDecisionBlobSizeExceedsLimit),
					Details: []byte("SignalExternalWorkflowExecutionDecisionAttributes.Input exceeds size limit."),
				}).Return(nil, errors.New("some error adding fail workflow event"))
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.SignalExternalWorkflowExecutionDecisionAttributes, err error) {
				assert.Equal(t, errors.New("some error adding fail workflow event"), err)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			taskHandler := newTaskHandlerForTest(t)
			if test.expectMockCalls != nil {
				test.expectMockCalls(taskHandler, test.attributes)
			}
			decision := &types.Decision{
				DecisionType: func(i int32) *types.DecisionType {
					decisionType := new(types.DecisionType)
					*decisionType = types.DecisionType(i)
					return decisionType
				}(11), //types.DecisionTypeSignalExternalWorkflowExecution
				SignalExternalWorkflowExecutionDecisionAttributes: test.attributes,
			}
			err := taskHandler.handleDecision(context.Background(), decision)
			test.asserts(t, taskHandler, test.attributes, err)
		})
	}
}

func TestHandleDecisionUpsertWorkflowSearchAttributes(t *testing.T) {
	tests := []struct {
		name            string
		expectMockCalls func(taskHandler *taskHandlerImpl, attr *types.UpsertWorkflowSearchAttributesDecisionAttributes)
		attributes      *types.UpsertWorkflowSearchAttributesDecisionAttributes
		asserts         func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.UpsertWorkflowSearchAttributesDecisionAttributes, err error)
	}{
		{
			name: "success",
			attributes: &types.UpsertWorkflowSearchAttributesDecisionAttributes{
				SearchAttributes: &types.SearchAttributes{IndexedFields: map[string][]byte{"some-key": []byte(`"some-value"`)}},
			},
			expectMockCalls: func(taskHandler *taskHandlerImpl, attr *types.UpsertWorkflowSearchAttributesDecisionAttributes) {
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().GetExecutionInfo().Times(2).Return(&persistence.WorkflowExecutionInfo{
					DomainID:   constants.TestDomainID,
					WorkflowID: constants.TestWorkflowID,
					RunID:      constants.TestRunID,
				})
				taskHandler.domainCache.(*cache.MockDomainCache).EXPECT().GetDomainName(constants.TestDomainID).Return(constants.TestDomainName, nil)
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddUpsertWorkflowSearchAttributesEvent(testTaskCompletedID, attr)
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.UpsertWorkflowSearchAttributesDecisionAttributes, err error) {
				assert.False(t, taskHandler.failDecision)
				assert.Empty(t, taskHandler.failMessage)
				assert.Nil(t, taskHandler.failDecisionCause)
				assert.Equal(t, nil, err)
			},
		},
		{
			name: "attributes validation failure",
			attributes: &types.UpsertWorkflowSearchAttributesDecisionAttributes{
				SearchAttributes: &types.SearchAttributes{IndexedFields: map[string][]byte{"some-key": []byte("some-value")}},
			},
			expectMockCalls: func(taskHandler *taskHandlerImpl, attr *types.UpsertWorkflowSearchAttributesDecisionAttributes) {
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().GetExecutionInfo().Times(1).Return(&persistence.WorkflowExecutionInfo{
					DomainID:   constants.TestDomainID,
					WorkflowID: constants.TestWorkflowID,
					RunID:      constants.TestRunID,
				})
				taskHandler.domainCache.(*cache.MockDomainCache).EXPECT().GetDomainName(constants.TestDomainID).Return(constants.TestDomainName, nil)
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.UpsertWorkflowSearchAttributesDecisionAttributes, err error) {
				assert.True(t, taskHandler.failDecision)
				assert.Equal(t, "some-value is not a valid search attribute value for key some-key", *taskHandler.failMessage)
				assert.Equal(t, func(i int32) *types.DecisionTaskFailedCause {
					cause := new(types.DecisionTaskFailedCause)
					*cause = types.DecisionTaskFailedCause(i)
					return cause
					// 22 is types.DecisionTaskFailedCause "BAD_SEARCH_ATTRIBUTES"
				}(22), taskHandler.failDecisionCause)
			},
		},
		{
			name: "size limit checker failure",
			attributes: &types.UpsertWorkflowSearchAttributesDecisionAttributes{
				SearchAttributes: &types.SearchAttributes{IndexedFields: map[string][]byte{"some-key": []byte(`"some-value"`)}},
			},
			expectMockCalls: func(taskHandler *taskHandlerImpl, attr *types.UpsertWorkflowSearchAttributesDecisionAttributes) {
				taskHandler.sizeLimitChecker.blobSizeLimitWarn = 2
				taskHandler.sizeLimitChecker.blobSizeLimitError = 1
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().GetExecutionInfo().Times(2).Return(&persistence.WorkflowExecutionInfo{
					DomainID:   constants.TestDomainID,
					WorkflowID: constants.TestWorkflowID,
					RunID:      constants.TestRunID,
				})
				taskHandler.domainCache.(*cache.MockDomainCache).EXPECT().GetDomainName(constants.TestDomainID).Return(constants.TestDomainName, nil)
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddFailWorkflowEvent(testTaskCompletedID, &types.FailWorkflowExecutionDecisionAttributes{
					Reason:  common.StringPtr(common.FailureReasonDecisionBlobSizeExceedsLimit),
					Details: []byte("UpsertWorkflowSearchAttributesDecisionAttributes exceeds size limit."),
				}).Return(nil, errors.New("some random error adding workflow event"))
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.UpsertWorkflowSearchAttributesDecisionAttributes, err error) {
				assert.False(t, taskHandler.failDecision)
				assert.Empty(t, taskHandler.failMessage)
				assert.Nil(t, taskHandler.failDecisionCause)
				assert.Equal(t, errors.New("some random error adding workflow event"), err)
			},
		},
		{
			name: "internal service error",
			attributes: &types.UpsertWorkflowSearchAttributesDecisionAttributes{
				SearchAttributes: &types.SearchAttributes{IndexedFields: map[string][]byte{"some-key": []byte(`"some-value"`)}},
			},
			expectMockCalls: func(taskHandler *taskHandlerImpl, attr *types.UpsertWorkflowSearchAttributesDecisionAttributes) {
				taskHandler.mutableState.(*execution.MockMutableState).EXPECT().GetExecutionInfo().Times(1).Return(&persistence.WorkflowExecutionInfo{
					DomainID:   constants.TestDomainID,
					WorkflowID: constants.TestWorkflowID,
					RunID:      constants.TestRunID,
				})
				taskHandler.domainCache.(*cache.MockDomainCache).EXPECT().GetDomainName(constants.TestDomainID).Return("", errors.New("some random error"))
			},
			asserts: func(t *testing.T, taskHandler *taskHandlerImpl, attr *types.UpsertWorkflowSearchAttributesDecisionAttributes, err error) {
				assert.Equal(t, &types.InternalServiceError{Message: "Unable to get domain for domainID: deadbeef-0123-4567-890a-bcdef0123456."}, err)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			taskHandler := newTaskHandlerForTest(t)
			if test.expectMockCalls != nil {
				test.expectMockCalls(taskHandler, test.attributes)
			}
			decision := &types.Decision{
				DecisionType: func(i int32) *types.DecisionType {
					decisionType := new(types.DecisionType)
					*decisionType = types.DecisionType(i)
					return decisionType
				}(12), //types.DecisionTypeUpsertWorkflowSearchAttributes == 12
				UpsertWorkflowSearchAttributesDecisionAttributes: test.attributes,
			}
			err := taskHandler.handleDecision(context.Background(), decision)
			test.asserts(t, taskHandler, test.attributes, err)
		})
	}
}

func newTaskHandlerForTest(t *testing.T) *taskHandlerImpl {
	ctrl := gomock.NewController(t)
	testLogger := testlogger.New(t)
	testConfig := config.NewForTest()
	testConfig.ValidSearchAttributes = func(opts ...dynamicconfig.FilterOption) map[string]interface{} {
		validSearchAttr := make(map[string]interface{})
		validSearchAttr["some-key"] = types.IndexedValueTypeString
		return validSearchAttr
	}
	mockMutableState := execution.NewMockMutableState(ctrl)
	mockDomainCache := cache.NewMockDomainCache(ctrl)
	workflowSizeChecker := newWorkflowSizeChecker(
		testConfig.BlobSizeLimitWarn(constants.TestDomainName),
		testConfig.BlobSizeLimitError(constants.TestDomainName),
		testConfig.HistorySizeLimitWarn(constants.TestDomainName),
		testConfig.HistorySizeLimitError(constants.TestDomainName),
		testConfig.HistoryCountLimitWarn(constants.TestDomainName),
		testConfig.HistoryCountLimitError(constants.TestDomainName),
		testTaskCompletedID,
		mockMutableState,
		&persistence.ExecutionStats{},
		metrics.NewClient(tally.NoopScope, metrics.History).Scope(metrics.HistoryRespondDecisionTaskCompletedScope, metrics.DomainTag(constants.TestDomainName)),
		testLogger,
	)
	mockMutableState.EXPECT().HasBufferedEvents().Return(false)
	taskHandler := newDecisionTaskHandler(
		testdata.Identity,
		testTaskCompletedID,
		constants.TestLocalDomainEntry,
		mockMutableState,
		newAttrValidator(mockDomainCache, metrics.NewClient(tally.NoopScope, metrics.History), testConfig, testlogger.New(t)),
		workflowSizeChecker,
		common.NewMockTaskTokenSerializer(ctrl),
		testLogger,
		mockDomainCache,
		metrics.NewClient(tally.NoopScope, metrics.History),
		testConfig,
	)
	return taskHandler
}
