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
	"math/rand"
	"sync/atomic"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/yarpc/yarpcerrors"

	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/metrics/mocks"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/engine"
	"github.com/uber/cadence/service/history/resource"
	"github.com/uber/cadence/service/history/shard"
	"github.com/uber/cadence/service/history/workflowcache"
)

const (
	testWorkflowID    = "test-workflow-id"
	testWorkflowRunID = "test-workflow-run-id"
	testDomainID      = "BF80C53A-ED56-4DD9-84EB-BE9AD4E45867"
)

type (
	handlerSuite struct {
		suite.Suite
		*require.Assertions

		controller          *gomock.Controller
		mockResource        *resource.Test
		mockShardController *shard.MockController
		mockEngine          *engine.MockEngine
		mockWFCache         *workflowcache.MockWFCache

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

	s.handler = NewHandler(s.mockResource, config.NewForTest()).(*handlerImpl)
	s.handler.controller = s.mockShardController
	s.handler.workflowIDCache = s.mockWFCache
	s.handler.startWG.Done()
}

func (s *handlerSuite) TearDownTest() {
	s.controller.Finish()
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

func TestCorrectUseOfErrorHandling(t *testing.T) {

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
		t.Run(name, func(t *testing.T) {
			scope := mocks.Scope{}
			td.expectation(&scope)
			h := handlerImpl{
				Resource: resource.NewTest(t, gomock.NewController(t), 0),
			}
			h.error(td.input, &scope, "some-domain", "some-wf", "some-run")
			// we're doing the args assertion in the On, so using mock.Anything to avoid having to duplicate this
			// a wrong metric being emitted will fail the mock.On() expectation. This will catch missing calls
			scope.Mock.AssertCalled(t, "IncCounter", mock.Anything)
		})
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

	// We should _always_ see the startworkflowexecution call no matter if allow external is true or false
	// as we are in shadow mode
	tests := map[string]struct{ allowExternal bool }{
		"allow external":    {allowExternal: true},
		"disallow external": {allowExternal: false},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.mockWFCache.EXPECT().AllowExternal(gomock.Any(), gomock.Any()).Return(test.allowExternal).Times(1)
			s.mockShardController.EXPECT().GetEngine(testWorkflowID).Return(s.mockEngine, nil).AnyTimes()
			s.mockEngine.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any()).Return(expectedResponse, nil).Times(1)

			response, err := s.handler.StartWorkflowExecution(context.Background(), request)
			s.Equal(expectedResponse, response)
			s.Nil(err)
		})
	}
}
