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

package history

import (
	"context"
	"errors"
	"math/rand"
	"sync/atomic"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/engine"
	"github.com/uber/cadence/service/history/resource"
	"github.com/uber/cadence/service/history/shard"
)

type (
	handlerSuite struct {
		suite.Suite
		*require.Assertions

		controller          *gomock.Controller
		mockResource        *resource.Test
		mockShardController *shard.MockController
		mockEngine          *engine.MockEngine

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
	s.mockResource = resource.NewTest(s.controller, metrics.History)
	s.mockResource.Logger = loggerimpl.NewLoggerForTest(s.Suite)
	s.mockShardController = shard.NewMockController(s.controller)
	s.mockEngine = engine.NewMockEngine(s.controller)
	s.mockShardController.EXPECT().GetEngineForShard(gomock.Any()).Return(s.mockEngine, nil).AnyTimes()

	s.handler = NewHandler(s.mockResource, config.NewForTest()).(*handlerImpl)
	s.handler.controller = s.mockShardController
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
