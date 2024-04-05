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

package handler

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/uber/cadence/common/metrics/mocks"
	"github.com/uber/cadence/service/history/constants"
	"go.uber.org/yarpc/yarpcerrors"
	"math/rand"
	"sync/atomic"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/engine"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/resource"
	"github.com/uber/cadence/service/history/shard"
	"github.com/uber/cadence/service/history/task"
	"github.com/uber/cadence/service/history/workflowcache"
)

const (
	testWorkflowID    = "test-workflow-id"
	testWorkflowRunID = "test-workflow-run-id"
	testDomainID      = "BF80C53A-ED56-4DD9-84EB-BE9AD4E45867"
	testValidUUID     = "FCD00931-EBD4-4028-B67E-4DE624641255"
)

type (
	handlerSuite struct {
		suite.Suite
		*require.Assertions

		controller                   *gomock.Controller
		mockResource                 *resource.Test
		mockShardController          *shard.MockController
		mockEngine                   *engine.MockEngine
		mockWFCache                  *workflowcache.MockWFCache
		mockTokenSerializer          *common.MockTaskTokenSerializer
		mockHistoryEventNotifier     *events.MockNotifier
		mockRatelimiter              *quotas.MockLimiter
		mockCrossClusterTaskFetchers *task.MockFetcher

		handler *handlerImpl
	}
)

func TestHandlerSuite(t *testing.T) {
	s := new(handlerSuite)
	suite.Run(t, s)
}

func (s *handlerSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockResource = resource.NewTest(s.T(), s.controller, metrics.History)
	s.mockResource.Logger = testlogger.New(s.Suite.T())
	s.mockShardController = shard.NewMockController(s.controller)
	s.mockEngine = engine.NewMockEngine(s.controller)
	s.mockShardController.EXPECT().GetEngineForShard(gomock.Any()).Return(s.mockEngine, nil).AnyTimes()
	s.mockWFCache = workflowcache.NewMockWFCache(s.controller)
	internalRequestRateLimitingEnabledConfig := func(domainName string) bool { return false }
	s.handler = NewHandler(s.mockResource, config.NewForTest(), s.mockWFCache, internalRequestRateLimitingEnabledConfig).(*handlerImpl)
	s.handler.controller = s.mockShardController
	s.mockTokenSerializer = common.NewMockTaskTokenSerializer(s.controller)
	s.mockRatelimiter = quotas.NewMockLimiter(s.controller)
	s.handler.rateLimiter = s.mockRatelimiter
	s.handler.tokenSerializer = s.mockTokenSerializer
	s.handler.startWG.Done()
}

func (s *handlerSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *handlerSuite) TestHealth() {
	hs, err := s.handler.Health(context.Background())
	s.NoError(err)
	s.Equal(&types.HealthStatus{Ok: true, Msg: "OK"}, hs)
}

func (s *handlerSuite) TestRecordActivityTaskHeartbeat() {
	testInput := map[string]struct {
		caseName      string
		input         *types.HistoryRecordActivityTaskHeartbeatRequest
		expected      *types.RecordActivityTaskHeartbeatResponse
		expectedError bool
	}{
		"valid input": {
			caseName: "valid input",
			input: &types.HistoryRecordActivityTaskHeartbeatRequest{
				DomainUUID: testDomainID,
				HeartbeatRequest: &types.RecordActivityTaskHeartbeatRequest{
					TaskToken: []byte("task-token"),
				},
			},
			expected:      &types.RecordActivityTaskHeartbeatResponse{CancelRequested: false},
			expectedError: false,
		},
		"empty domainID": {
			caseName: "empty domainID",
			input: &types.HistoryRecordActivityTaskHeartbeatRequest{
				DomainUUID: "",
			},
			expected:      nil,
			expectedError: true,
		},
		"ratelimit exceeded": {
			caseName: "ratelimit exceeded",
			input: &types.HistoryRecordActivityTaskHeartbeatRequest{
				DomainUUID: testDomainID,
			},
			expected:      nil,
			expectedError: true,
		},
		"token deserialization error": {
			caseName: "token deserialization error",
			input: &types.HistoryRecordActivityTaskHeartbeatRequest{
				DomainUUID: testDomainID,
				HeartbeatRequest: &types.RecordActivityTaskHeartbeatRequest{
					TaskToken: []byte("task-token"),
				},
			},
			expected:      nil,
			expectedError: true,
		},
		"invalid task token": {
			caseName: "invalid task token",
			input: &types.HistoryRecordActivityTaskHeartbeatRequest{
				DomainUUID: testDomainID,
				HeartbeatRequest: &types.RecordActivityTaskHeartbeatRequest{
					TaskToken: []byte("task-token"),
				},
			},
			expected:      nil,
			expectedError: true,
		},
		"get engine error": {
			caseName: "get engine error",
			input: &types.HistoryRecordActivityTaskHeartbeatRequest{
				DomainUUID: testDomainID,
				HeartbeatRequest: &types.RecordActivityTaskHeartbeatRequest{
					TaskToken: []byte("task-token"),
				},
			},
			expected:      nil,
			expectedError: true,
		},
		"engine error": {
			caseName: "engine error",
			input: &types.HistoryRecordActivityTaskHeartbeatRequest{
				DomainUUID: testDomainID,
				HeartbeatRequest: &types.RecordActivityTaskHeartbeatRequest{
					TaskToken: []byte("task-token"),
				},
			},
			expected:      nil,
			expectedError: true,
		},
	}

	for name, input := range testInput {
		s.Run(name, func() {
			switch input.caseName {
			case "valid input":
				s.mockTokenSerializer.EXPECT().Deserialize(gomock.Any()).Return(&common.TaskToken{
					WorkflowID: testWorkflowID,
					RunID:      testValidUUID,
				}, nil).Times(1)
				s.mockShardController.EXPECT().GetEngine(testWorkflowID).Return(s.mockEngine, nil).Times(1)
				s.mockEngine.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), input.input).Return(input.expected, nil).Times(1)
				s.mockRatelimiter.EXPECT().Allow().Return(true).Times(1)
			case "empty domainID":
			case "ratelimit exceeded":
				s.mockRatelimiter.EXPECT().Allow().Return(false).Times(1)
			case "token deserialization error":
				s.mockRatelimiter.EXPECT().Allow().Return(true).Times(1)
				s.mockTokenSerializer.EXPECT().Deserialize(gomock.Any()).Return(nil, errors.New("some random error")).Times(1)
			case "invalid task token":
				s.mockRatelimiter.EXPECT().Allow().Return(true).Times(1)
				s.mockTokenSerializer.EXPECT().Deserialize(gomock.Any()).Return(&common.TaskToken{
					WorkflowID: "",
					RunID:      "",
				}, nil).Times(1)
			case "get engine error":
				s.mockRatelimiter.EXPECT().Allow().Return(true).Times(1)
				s.mockTokenSerializer.EXPECT().Deserialize(gomock.Any()).Return(&common.TaskToken{
					WorkflowID: testWorkflowID,
					RunID:      testValidUUID,
				}, nil).Times(1)
				s.mockShardController.EXPECT().GetEngine(testWorkflowID).Return(nil, errors.New("error")).Times(1)
			case "engine error":
				s.mockTokenSerializer.EXPECT().Deserialize(gomock.Any()).Return(&common.TaskToken{
					WorkflowID: testWorkflowID,
					RunID:      testValidUUID,
				}, nil).Times(1)
				s.mockShardController.EXPECT().GetEngine(testWorkflowID).Return(s.mockEngine, nil).Times(1)
				s.mockEngine.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), input.input).Return(nil, errors.New("error")).Times(1)
				s.mockRatelimiter.EXPECT().Allow().Return(true).Times(1)
			}
			response, err := s.handler.RecordActivityTaskHeartbeat(context.Background(), input.input)
			s.Equal(input.expected, response)
			if input.expectedError {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *handlerSuite) TestGetCrossClusterTasks() {
	numShards := 10
	targetCluster := cluster.TestAlternativeClusterName
	var shardIDs []int32
	numSucceeded := int32(0)
	numTasksPerShard := rand.Intn(10)
	s.mockEngine.EXPECT().GetCrossClusterTasks(gomock.Any(), targetCluster).DoAndReturn(
		func(_ context.Context, _ string) ([]*types.CrossClusterTaskRequest, error) {
			succeeded := rand.Intn(2) == 0
			if succeeded {
				atomic.AddInt32(&numSucceeded, 1)
				return make([]*types.CrossClusterTaskRequest, numTasksPerShard), nil
			}
			return nil, errors.New("some random error")
		},
	).MaxTimes(numShards)
	for i := 0; i != numShards; i++ {
		shardIDs = append(shardIDs, int32(i))
	}
	request := &types.GetCrossClusterTasksRequest{
		ShardIDs:      shardIDs,
		TargetCluster: targetCluster,
	}

	response, err := s.handler.GetCrossClusterTasks(context.Background(), request)
	s.NoError(err)
	s.NotNil(response)

	s.Len(response.TasksByShard, int(numSucceeded))
	s.Len(response.FailedCauseByShard, numShards-int(numSucceeded))
	for _, tasksRequests := range response.GetTasksByShard() {
		s.Len(tasksRequests, numTasksPerShard)
	}
}

func (s *handlerSuite) TestRespondCrossClusterTaskCompleted_FetchNewTask() {
	s.testRespondCrossClusterTaskCompleted(true)
}

func (s *handlerSuite) TestRespondCrossClusterTaskCompleted_NoNewTask() {
	s.testRespondCrossClusterTaskCompleted(false)
}

func (s *handlerSuite) testRespondCrossClusterTaskCompleted(
	fetchNewTask bool,
) {
	numTasks := 10
	targetCluster := cluster.TestAlternativeClusterName
	request := &types.RespondCrossClusterTasksCompletedRequest{
		ShardID:       0,
		TargetCluster: targetCluster,
		TaskResponses: make([]*types.CrossClusterTaskResponse, numTasks),
		FetchNewTasks: fetchNewTask,
	}
	s.mockEngine.EXPECT().RespondCrossClusterTasksCompleted(gomock.Any(), targetCluster, request.TaskResponses).Return(nil).Times(1)
	if fetchNewTask {
		s.mockEngine.EXPECT().GetCrossClusterTasks(gomock.Any(), targetCluster).Return(make([]*types.CrossClusterTaskRequest, numTasks), nil).Times(1)
	}

	response, err := s.handler.RespondCrossClusterTasksCompleted(context.Background(), request)
	s.NoError(err)
	s.NotNil(response)

	if !fetchNewTask {
		s.Empty(response.Tasks)
	} else {
		s.Len(response.Tasks, numTasks)
	}
}

func (s *handlerSuite) TestStartWorkflowExecution() {

	request := &types.HistoryStartWorkflowExecutionRequest{
		DomainUUID: testDomainID,
		StartRequest: &types.StartWorkflowExecutionRequest{
			WorkflowID: testWorkflowID,
		},
	}

	expectedResponse := &types.StartWorkflowExecutionResponse{
		RunID: testWorkflowRunID,
	}

	s.mockShardController.EXPECT().GetEngine(testWorkflowID).Return(s.mockEngine, nil).AnyTimes()
	s.mockRatelimiter.EXPECT().Allow().Return(true).AnyTimes()
	s.mockEngine.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any()).Return(expectedResponse, nil).AnyTimes()

	response, err := s.handler.StartWorkflowExecution(context.Background(), request)
	s.Equal(expectedResponse, response)
	s.Nil(err)
}

func (s *handlerSuite) TestEmitInfoOrDebugLog() {
	// test emitInfoOrDebugLog
	s.mockResource.Logger = testlogger.New(s.Suite.T())
	s.handler.emitInfoOrDebugLog("domain1", "test log")
}

func (s *handlerSuite) TestValidateTaskToken() {
	testInput := map[string]struct {
		taskToken     *common.TaskToken
		expectedError error
	}{
		"valid task token": {
			taskToken: &common.TaskToken{
				WorkflowID: testWorkflowID,
				RunID:      testValidUUID,
			},
			expectedError: nil,
		},
		"empty workflow id": {
			taskToken: &common.TaskToken{
				WorkflowID: "",
			},
			expectedError: constants.ErrWorkflowIDNotSet,
		},
		"invalid run id": {
			taskToken: &common.TaskToken{
				WorkflowID: testWorkflowID,
				RunID:      "invalid",
			},
			expectedError: constants.ErrRunIDNotValid,
		},
	}

	for name, input := range testInput {
		s.Run(name, func() {
			err := validateTaskToken(input.taskToken)
			s.Equal(input.expectedError, err)
		})
	}
}

func (s *handlerSuite) TestCorrectUseOfErrorHandling() {

	tests := map[string]struct {
		input       error
		expectation func(scope *mocks.Scope)
	}{
		"A deadline exceeded error": {
			input: context.DeadlineExceeded,
			expectation: func(scope *mocks.Scope) {
				scope.Mock.On("IncCounter", metrics.CadenceErrContextTimeoutCounter).Once()
			},
		},
		"A cancelled error": {
			input: context.Canceled,
			expectation: func(scope *mocks.Scope) {
				scope.Mock.On("IncCounter", metrics.CadenceErrContextTimeoutCounter).Once()
			},
		},
		"A shard ownership lost error": {
			input: &types.ShardOwnershipLostError{
				Message: "something is lost",
				Owner:   "owner",
			},
			expectation: func(scope *mocks.Scope) {
				scope.Mock.On("IncCounter", metrics.CadenceErrShardOwnershipLostCounter).Once()
			},
		},
		"a workflow is already started": {
			input: &types.EventAlreadyStartedError{
				Message: "workflow already running",
			},
			expectation: func(scope *mocks.Scope) {
				scope.Mock.On("IncCounter", metrics.CadenceErrEventAlreadyStartedCounter).Once()
			},
		},
		"a bad request": {
			input: &types.BadRequestError{
				Message: "bad request",
			},
			expectation: func(scope *mocks.Scope) {
				scope.Mock.On("IncCounter", metrics.CadenceErrBadRequestCounter).Once()
			},
		},
		"domain is not active": {
			input: &types.DomainNotActiveError{
				Message: "domain not active",
			},
			expectation: func(scope *mocks.Scope) {
				scope.Mock.On("IncCounter", metrics.CadenceErrDomainNotActiveCounter).Once()
			},
		},
		"workflow is already started err": {
			input: &types.WorkflowExecutionAlreadyStartedError{
				Message: "bad already started",
			},
			expectation: func(scope *mocks.Scope) {
				scope.Mock.On("IncCounter", metrics.CadenceErrExecutionAlreadyStartedCounter).Once()
			},
		},
		"does not exist": {
			input: &types.EntityNotExistsError{
				Message: "the workflow doesn't exist",
			},
			expectation: func(scope *mocks.Scope) {
				scope.Mock.On("IncCounter", metrics.CadenceErrEntityNotExistsCounter).Once()
			},
		},
		"already completed": {
			input: &types.WorkflowExecutionAlreadyCompletedError{
				Message: "the workflow is done",
			},
			expectation: func(scope *mocks.Scope) {
				scope.Mock.On("IncCounter", metrics.CadenceErrWorkflowExecutionAlreadyCompletedCounter).Once()
			},
		},
		"Cancellation already requested": {
			input: &types.CancellationAlreadyRequestedError{
				Message: "the workflow is cancelled already",
			},
			expectation: func(scope *mocks.Scope) {
				scope.Mock.On("IncCounter", metrics.CadenceErrCancellationAlreadyRequestedCounter).Once()
			},
		},
		"rate-limits": {
			input: &types.LimitExceededError{
				Message: "limits exceeded",
			},
			expectation: func(scope *mocks.Scope) {
				scope.Mock.On("IncCounter", metrics.CadenceErrLimitExceededCounter).Once()
			},
		},
		"retry tasks": {
			input: &types.RetryTaskV2Error{
				Message: "limits exceeded",
			},
			expectation: func(scope *mocks.Scope) {
				scope.Mock.On("IncCounter", metrics.CadenceErrRetryTaskCounter).Once()
			},
		},
		"service busy error": {
			input: &types.ServiceBusyError{
				Message: "limits exceeded - service is busy",
			},
			expectation: func(scope *mocks.Scope) {
				scope.Mock.On("IncCounter", metrics.CadenceErrServiceBusyCounter).Once()
			},
		},
		"deadline exceeded": {
			input: yarpcerrors.DeadlineExceededErrorf("some deadline exceeded err"),
			expectation: func(scope *mocks.Scope) {
				scope.Mock.On("IncCounter", metrics.CadenceErrContextTimeoutCounter).Once()
				scope.Mock.On("IncCounter", metrics.CadenceFailures).Once()
			},
		},
		"internal error": {
			input: types.InternalServiceError{Message: "internal error"},
			expectation: func(scope *mocks.Scope) {
				scope.Mock.On("IncCounter", metrics.CadenceFailures).Once()
			},
		},
		"uncategorized error": {
			input: errors.New("some random error"),
			expectation: func(scope *mocks.Scope) {
				scope.Mock.On("IncCounter", metrics.CadenceFailures).Once()
			},
		},
	}

	for name, td := range tests {
		s.Run(name, func() {
			scope := mocks.Scope{}
			td.expectation(&scope)
			h := handlerImpl{
				Resource: resource.NewTest(s.T(), gomock.NewController(s.T()), 0),
			}
			h.error(td.input, &scope, "some-domain", "some-wf", "some-run")
			// we're doing the args assertion in the On, so using mock.Anything to avoid having to duplicate this
			// a wrong metric being emitted will fail the mock.On() expectation. This will catch missing calls
			scope.Mock.AssertCalled(s.T(), "IncCounter", mock.Anything)
		})
	}
}
