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
	"math/rand"
	"testing"
	"time"

	"go.uber.org/yarpc"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"

	"github.com/uber/cadence/client"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/future"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/constants"
)

type (
	fetcherSuite struct {
		suite.Suite
		*require.Assertions

		controller *gomock.Controller

		options       *FetcherOptions
		metricsClient metrics.Client
		logger        log.Logger
	}

	testFetchResult struct {
		fetchParams []interface{}
	}
)

func TestFetcherSuite(t *testing.T) {
	s := new(fetcherSuite)
	suite.Run(t, s)
}

func (s *fetcherSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())

	s.options = &FetcherOptions{
		Parallelism:            dynamicconfig.GetIntPropertyFn(3),
		AggregationInterval:    dynamicconfig.GetDurationPropertyFn(time.Millisecond * 100),
		TimerJitterCoefficient: dynamicconfig.GetFloatPropertyFn(0.5),
	}
	s.metricsClient = metrics.NewClient(tally.NoopScope, metrics.History)
	s.logger = loggerimpl.NewLoggerForTest(s.Suite)
}

func (s *fetcherSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *fetcherSuite) TestCrossClusterTaskFetchers() {
	s.options.Parallelism = dynamicconfig.GetIntPropertyFn(1)
	sourceCluster := cluster.TestAlternativeClusterName
	currentCluster := cluster.TestCurrentClusterName

	shardIDs := []int32{1, 10, 123}
	tasksByShard := make(map[int32][]*types.CrossClusterTaskRequest)

	mockResource := resource.NewTest(s.controller, metrics.History)
	mockResource.RemoteAdminClient.EXPECT().GetCrossClusterTasks(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, request *types.GetCrossClusterTasksRequest, option ...yarpc.CallOption) (*types.GetCrossClusterTasksResponse, error) {
			s.Equal(currentCluster, request.GetTargetCluster())
			resp := &types.GetCrossClusterTasksResponse{
				TasksByShard: make(map[int32][]*types.CrossClusterTaskRequest),
			}
			for _, shardID := range request.GetShardIDs() {
				_, ok := tasksByShard[shardID]
				s.False(ok)

				numTasks := rand.Intn(10)
				taskRequests := make([]*types.CrossClusterTaskRequest, numTasks)
				for i := 0; i != numTasks; i++ {
					taskRequests[i] = &types.CrossClusterTaskRequest{
						TaskInfo: &types.CrossClusterTaskInfo{
							TaskID: rand.Int63n(10000),
						},
					}
				}
				resp.TasksByShard[shardID] = taskRequests
				tasksByShard[shardID] = taskRequests
			}
			return resp, nil
		},
	).AnyTimes()

	crossClusterTaskFetchers := NewCrossClusterTaskFetchers(
		constants.TestClusterMetadata,
		mockResource.GetClientBean(),
		s.options,
		s.metricsClient,
		s.logger,
	)
	var fetcher Fetcher
	for _, f := range crossClusterTaskFetchers {
		if f.GetSourceCluster() == sourceCluster {
			s.Nil(fetcher)
			fetcher = f
		}
	}
	s.NotNil(fetcher)

	fetcher.Start()
	futures := make(map[int32]future.Future, len(shardIDs))
	for _, shardID := range shardIDs {
		futures[shardID] = fetcher.Fetch(int(shardID))
	}

	for shardID, future := range futures {
		var taskRequests []*types.CrossClusterTaskRequest
		err := future.Get(context.Background(), &taskRequests)
		s.NoError(err)
		s.Equal(tasksByShard[shardID], taskRequests)
	}

	fetcher.Stop()
	mockResource.Finish(s.T())
}

func (s *fetcherSuite) TestNewTaskFetchers() {
	clusterMetadata := constants.TestClusterMetadata
	fetchers := newTaskFetchers(clusterMetadata, nil, nil, s.options, s.metricsClient, s.logger)

	var sourceClusters []string
	for _, f := range fetchers {
		sourceCluster := f.GetSourceCluster()
		s.NotEqual(clusterMetadata.GetCurrentClusterName(), sourceCluster)
		sourceClusters = append(sourceClusters, sourceCluster)
	}
	s.Len(sourceClusters, 1)

	fetchers.Start()
	fetchers.Stop()
}

func (s *fetcherSuite) TestFetch() {
	fetcher := newTaskFetcher(
		cluster.TestCurrentClusterName,
		cluster.TestAlternativeClusterName,
		nil,
		s.testFetchTaskFn,
		s.options,
		s.metricsClient,
		s.logger,
	)
	fetcher.Start()

	fetchParams := []interface{}{"token", 1, []byte{1, 2, 3}}
	var fetchResult testFetchResult
	err := fetcher.Fetch(0, fetchParams...).Get(context.Background(), &fetchResult)
	s.NoError(err)
	s.Equal(fetchParams, fetchResult.fetchParams)

	fetcher.Stop()
	err = fetcher.Fetch(0, fetchParams...).Get(context.Background(), &fetchResult)
	s.Error(err)
}

func (s *fetcherSuite) TestAggregator() {
	fetcher := newTaskFetcher(
		cluster.TestCurrentClusterName,
		cluster.TestAlternativeClusterName,
		nil,
		s.testFetchTaskFn,
		s.options,
		s.metricsClient,
		s.logger,
	)
	fetcher.Start()

	numShards := 10
	futures := make([]future.Future, 0, numShards)
	for shardID := 0; shardID != numShards; shardID++ {
		time.Sleep(time.Millisecond * 20)
		f := fetcher.Fetch(shardID, shardID)
		futures = append(futures, f)
	}

	for shardID, f := range futures {
		var fetchRequest testFetchResult
		err := f.Get(context.Background(), &fetchRequest)
		s.NoError(err)
		s.Len(fetchRequest.fetchParams, 1)
		s.Equal(shardID, fetchRequest.fetchParams[0])
	}

	fetcher.Stop()
}

func (s *fetcherSuite) testFetchTaskFn(
	ctx context.Context,
	clientBean client.Bean,
	sourceCluster string,
	currentCluster string,
	tokenByShard map[int32]fetchRequest,
) (map[int32]interface{}, error) {
	results := make(map[int32]interface{}, len(tokenByShard))
	for shardID, request := range tokenByShard {
		results[shardID] = testFetchResult{
			fetchParams: request.params,
		}
	}
	return results, nil
}
