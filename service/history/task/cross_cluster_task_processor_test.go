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
	"errors"
	"math/rand"
	"testing"
	"time"

	"go.uber.org/yarpc"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/future"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/config"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/shard"
)

type (
	crossClusterTaskProcessorSuite struct {
		suite.Suite
		*require.Assertions

		controller    *gomock.Controller
		mockShard     *shard.TestContext
		mockProcessor *MockProcessor

		processorOptions *CrossClusterTaskProcessorOptions
		fetcherOptions   *FetcherOptions
		retryPolicy      *backoff.ExponentialRetryPolicy
	}
)

func TestCrossClusterTaskProcessSuite(t *testing.T) {
	s := new(crossClusterTaskProcessorSuite)
	suite.Run(t, s)
}

func (s *crossClusterTaskProcessorSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockShard = shard.NewTestContext(
		s.controller,
		&persistence.ShardInfo{
			ShardID: 0,
			RangeID: 1,
		},
		config.NewForTest(),
	)
	s.mockProcessor = NewMockProcessor(s.controller)
	s.mockShard.Resource.ClusterMetadata.EXPECT().GetCurrentClusterName().Return(cluster.TestCurrentClusterName).AnyTimes()
	s.mockShard.Resource.DomainCache.EXPECT().GetDomainName(constants.TestDomainID).Return(constants.TestDomainName, nil).AnyTimes()

	s.processorOptions = &CrossClusterTaskProcessorOptions{
		MaxPendingTasks:            dynamicconfig.GetIntPropertyFn(100),
		TaskMaxRetryCount:          dynamicconfig.GetIntPropertyFn(100),
		TaskRedispatchInterval:     dynamicconfig.GetDurationPropertyFn(time.Hour),
		TaskWaitInterval:           dynamicconfig.GetDurationPropertyFn(time.Millisecond * 100),
		ServiceBusyBackoffInterval: dynamicconfig.GetDurationPropertyFn(time.Millisecond * 200),
		TimerJitterCoefficient:     dynamicconfig.GetFloatPropertyFn(0.1),
	}
	s.fetcherOptions = &FetcherOptions{
		Parallelism:            dynamicconfig.GetIntPropertyFn(3),
		AggregationInterval:    dynamicconfig.GetDurationPropertyFn(time.Millisecond * 100),
		TimerJitterCoefficient: dynamicconfig.GetFloatPropertyFn(0.5),
	}
	s.retryPolicy = backoff.NewExponentialRetryPolicy(time.Millisecond * 50)
	s.retryPolicy.SetMaximumInterval(time.Millisecond * 100)
	s.retryPolicy.SetMaximumAttempts(3)
}

func (s *crossClusterTaskProcessorSuite) TestCrossClusterTaskProcessorStartStop() {
	taskFetchers := NewCrossClusterTaskFetchers(
		constants.TestClusterMetadata,
		s.mockShard.GetClientBean(),
		s.fetcherOptions,
		s.mockShard.GetMetricsClient(),
		s.mockShard.GetLogger(),
	)
	taskProcessors := NewCrossClusterTaskProcessors(s.mockShard, s.mockProcessor, taskFetchers, s.processorOptions)
	s.Len(taskProcessors, len(taskFetchers))

	s.mockShard.Resource.RemoteAdminClient.EXPECT().GetCrossClusterTasks(gomock.Any(), gomock.Any()).Return(
		&types.GetCrossClusterTasksResponse{
			TasksByShard: map[int32][]*types.CrossClusterTaskRequest{int32(s.mockShard.GetShardID()): {}},
		}, nil,
	).AnyTimes()

	taskFetchers.Start()
	taskProcessors.Start()
	taskProcessors.Stop()
	taskFetchers.Stop()
}

func (s *crossClusterTaskProcessorSuite) TestRespondPendingTasks_Failed() {
	s.testRespondPendingTasks(true)
}

func (s *crossClusterTaskProcessorSuite) TestRespondPendingTasks_Success() {
	s.testRespondPendingTasks(false)
}

func (s *crossClusterTaskProcessorSuite) testRespondPendingTasks(failedToRespond bool) {
	fetcher := newTaskFetcher(
		cluster.TestAlternativeClusterName,
		cluster.TestCurrentClusterName,
		nil, nil, nil,
		s.mockShard.GetMetricsClient(),
		s.mockShard.GetLogger(),
	)
	processor := newCrossClusterTaskProcessor(s.mockShard, s.mockProcessor, fetcher, s.processorOptions)
	numPendingTasks := 10
	completedTasks := 6
	futureSettables := make([]future.Settable, numPendingTasks)
	for i := 0; i != numPendingTasks; i++ {
		taskFuture, settable := future.NewFuture()
		processor.pendingTasks[int64(i)] = taskFuture
		futureSettables[i] = settable
	}

	for idx := range rand.Perm(numPendingTasks)[:completedTasks] {
		futureSettables[idx].Set(types.CrossClusterTaskResponse{TaskID: int64(idx)}, nil)
	}

	s.mockShard.Resource.RemoteAdminClient.EXPECT().RespondCrossClusterTasksCompleted(gomock.Any(), gomock.Any()).DoAndReturn(
		func(
			_ context.Context,
			request *types.RespondCrossClusterTasksCompletedRequest,
			option ...yarpc.CallOption,
		) (*types.RespondCrossClusterTasksCompletedResponse, error) {
			select {
			case <-processor.shutdownCh:
			default:
				close(processor.shutdownCh)
			}

			s.Len(request.TaskResponses, completedTasks)
			s.Equal(s.mockShard.Resource.GetClusterMetadata().GetCurrentClusterName(), request.TargetCluster)
			s.Equal(s.mockShard.GetShardID(), int(request.GetShardID()))
			s.False(request.GetFetchNewTasks())
			if failedToRespond {
				return nil, errors.New("some random error")
			}
			return &types.RespondCrossClusterTasksCompletedResponse{}, nil
		},
	).AnyTimes()

	processor.shutdownWG.Add(1)
	go processor.respondPendingTaskLoop()
	processor.shutdownWG.Wait()

	if failedToRespond {
		s.Len(processor.pendingTasks, numPendingTasks)
	} else {
		s.Len(processor.pendingTasks, numPendingTasks-completedTasks)
	}
}

func (s *crossClusterTaskProcessorSuite) TestProcessTaskRequests_RespondFailed() {
	s.testProcessTaskRequests(true)
}

func (s *crossClusterTaskProcessorSuite) TestProcessTaskRequests_RespondSuccess() {
	s.testProcessTaskRequests(false)
}

func (s *crossClusterTaskProcessorSuite) testProcessTaskRequests(failedToRespond bool) {
	fetcher := newTaskFetcher(
		cluster.TestAlternativeClusterName,
		cluster.TestCurrentClusterName,
		nil, nil, nil,
		s.mockShard.GetMetricsClient(),
		s.mockShard.GetLogger(),
	)
	processor := newCrossClusterTaskProcessor(s.mockShard, s.mockProcessor, fetcher, s.processorOptions)
	pendingTaskID := 0
	future, _ := future.NewFuture()
	processor.pendingTasks[int64(pendingTaskID)] = future

	numTasks := 10
	var tasksRequests []*types.CrossClusterTaskRequest
	for id := pendingTaskID; id != pendingTaskID+numTasks; id++ {
		tasksRequests = append(tasksRequests, &types.CrossClusterTaskRequest{
			TaskInfo: &types.CrossClusterTaskInfo{
				TaskID:   int64(id),
				DomainID: constants.TestDomainID,
				TaskType: types.CrossClusterTaskTypeCancelExecution.Ptr(),
			},
			CancelExecutionAttributes: &types.CrossClusterCancelExecutionRequestAttributes{},
		})
	}

	completedTasks := 0
	s.mockProcessor.EXPECT().TrySubmit(gomock.Any()).DoAndReturn(
		func(t Task) (bool, error) {
			submitted := rand.Intn(2) == 0
			if submitted {
				completedTasks++
				t.(*crossClusterTargetTask).response = &types.CrossClusterTaskResponse{
					TaskID: t.GetTaskID(),
				}
				t.Ack()
			}
			// redispatcher interval is set to 1hr, basically disabled
			// so that it won't re-submit tasks and we can easily count
			// how many task submits are attempted
			return submitted, nil
		},
	).Times(numTasks - 1) // -1 since there's a duplicate task

	s.mockShard.Resource.RemoteAdminClient.EXPECT().RespondCrossClusterTasksCompleted(gomock.Any(), gomock.Any()).DoAndReturn(
		func(
			_ context.Context,
			request *types.RespondCrossClusterTasksCompletedRequest,
			option ...yarpc.CallOption,
		) (*types.RespondCrossClusterTasksCompletedResponse, error) {
			s.Len(request.TaskResponses, completedTasks)
			s.Equal(s.mockShard.Resource.GetClusterMetadata().GetCurrentClusterName(), request.TargetCluster)
			s.Equal(s.mockShard.GetShardID(), int(request.GetShardID()))
			s.True(request.GetFetchNewTasks())
			if failedToRespond {
				return nil, errors.New("some random error")
			}
			return &types.RespondCrossClusterTasksCompletedResponse{
				Tasks: []*types.CrossClusterTaskRequest{
					{
						TaskInfo: &types.CrossClusterTaskInfo{TaskID: int64(pendingTaskID)},
					},
				},
			}, nil
		},
	).AnyTimes()

	processor.processTaskRequests(tasksRequests)

	if failedToRespond {
		s.Len(processor.pendingTasks, numTasks)
	} else {
		s.Len(processor.pendingTasks, numTasks-completedTasks)
	}
}

func (s *crossClusterTaskProcessorSuite) TestProcessLoop() {
	crossClusterTaskFetchers := NewCrossClusterTaskFetchers(
		constants.TestClusterMetadata,
		s.mockShard.GetClientBean(),
		&FetcherOptions{
			Parallelism:            dynamicconfig.GetIntPropertyFn(3),
			AggregationInterval:    dynamicconfig.GetDurationPropertyFn(time.Millisecond * 100),
			TimerJitterCoefficient: dynamicconfig.GetFloatPropertyFn(0.5),
		},
		s.mockShard.GetMetricsClient(),
		s.mockShard.GetLogger(),
	)
	var fetcher Fetcher
	for _, f := range crossClusterTaskFetchers {
		if f.GetSourceCluster() == cluster.TestAlternativeClusterName {
			s.Nil(fetcher)
			fetcher = f
		}
	}
	s.NotNil(fetcher)
	processor := newCrossClusterTaskProcessor(s.mockShard, s.mockProcessor, fetcher, s.processorOptions)

	totalGetRequests := 3
	numGetRequests := 0
	s.mockShard.Resource.RemoteAdminClient.EXPECT().GetCrossClusterTasks(gomock.Any(), gomock.Any()).DoAndReturn(
		func(
			_ context.Context,
			request *types.GetCrossClusterTasksRequest,
			option ...yarpc.CallOption,
		) (*types.GetCrossClusterTasksResponse, error) {
			numGetRequests++
			if numGetRequests == totalGetRequests {
				close(processor.shutdownCh)
			}

			s.Len(request.ShardIDs, 1)
			s.Equal(int32(s.mockShard.GetShardID()), request.ShardIDs[0])
			s.Equal(cluster.TestCurrentClusterName, request.GetTargetCluster())
			return &types.GetCrossClusterTasksResponse{
				TasksByShard: map[int32][]*types.CrossClusterTaskRequest{int32(s.mockShard.GetShardID()): {}},
			}, nil
		},
	).AnyTimes()

	fetcher.Start()
	processor.shutdownWG.Add(1)
	go processor.processLoop()
	processor.shutdownWG.Wait()
	fetcher.Stop()
}
