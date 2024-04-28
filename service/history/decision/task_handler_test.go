package decision

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/common/types/testdata"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/execution"
	"testing"
)

const (
	testTaskCompletedID = int64(1234)
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
				}(9), taskHandler.failDecisionCause)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			taskHandler := constructTestTaskHandler(t)
			taskHandler.mutableState.(*execution.MockMutableState).EXPECT().GetExecutionInfo().Times(1).Return(&persistence.WorkflowExecutionInfo{
				DomainID:   constants.TestDomainID,
				WorkflowID: constants.TestWorkflowID,
				RunID:      constants.TestRunID,
			})
			taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddRequestCancelExternalWorkflowExecutionInitiatedEvent(testTaskCompletedID, gomock.Any(), test.attributes).AnyTimes()
			if test.expectMockCalls != nil {
				test.expectMockCalls(taskHandler)
			}
			err := taskHandler.handleDecisionRequestCancelExternalWorkflow(context.Background(), test.attributes)
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
				}(2), taskHandler.failDecisionCause)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			taskHandler := constructTestTaskHandler(t)
			taskHandler.mutableState.(*execution.MockMutableState).EXPECT().AddRequestCancelExternalWorkflowExecutionInitiatedEvent(testTaskCompletedID, gomock.Any(), test.attributes).AnyTimes()
			if test.expectMockCalls != nil {
				test.expectMockCalls(taskHandler)
			}
			err := taskHandler.handleDecisionRequestCancelActivity(context.Background(), test.attributes)
			test.asserts(t, taskHandler, err)
		})
	}
}

func TestHandleDecisionStartChildWorkflow(t *testing.T) {

}

func constructTestTaskHandler(t *testing.T) *taskHandlerImpl {
	ctrl := gomock.NewController(t)
	testLogger := testlogger.New(t)
	testConfig := config.NewForTest()
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
		newAttrValidator(mockDomainCache, metrics.NewClient(tally.NoopScope, metrics.History), config.NewForTest(), testlogger.New(t)),
		workflowSizeChecker,
		common.NewMockTaskTokenSerializer(ctrl),
		testLogger,
		mockDomainCache,
		metrics.NewClient(tally.NoopScope, metrics.History),
		testConfig,
	)
	return taskHandler
}
