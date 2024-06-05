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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"

	"github.com/uber/cadence/client"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/future"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
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
	s.logger = testlogger.New(s.Suite.T())
}

func (s *fetcherSuite) TearDownTest() {
	s.controller.Finish()
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
