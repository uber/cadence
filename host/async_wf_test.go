// Copyright (c) 2018 Uber Technologies, Inc.
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

// To run locally,
// 1. Run Kafka and Cassandra
// 	docker-compose -f docker-compose.yml up
// 2. Run the test
// 	CASSANDRA=1 go test -v ./host -run TestAsyncWFIntegrationSuite

package host

import (
	"flag"
	"testing"
	"time"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/types"
)

type AsyncWFIntegrationSuite struct {
	*require.Assertions
	*IntegrationBase
	logger log.Logger
}

func TestAsyncWFIntegrationSuite(t *testing.T) {
	flag.Parse()
	clusterConfig, err := GetTestClusterConfig("testdata/integration_async_wf_with_kafka_cluster.yaml")
	if err != nil {
		panic(err)
	}
	testCluster := NewPersistenceTestCluster(t, clusterConfig)

	s := new(AsyncWFIntegrationSuite)
	params := IntegrationBaseParams{
		DefaultTestCluster: testCluster,
		TestClusterConfig:  clusterConfig,
	}
	s.IntegrationBase = NewIntegrationBase(params)
	suite.Run(t, s)
}

func (s *AsyncWFIntegrationSuite) SetupSuite() {
	s.setupSuite()
}

func (s *AsyncWFIntegrationSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *AsyncWFIntegrationSuite) TearDownSuite() {
	s.tearDownSuite()

}

func (s *AsyncWFIntegrationSuite) TestStartWorkflowExecutionAsync() {
	id := "async-wf-integration-start-workflow-test"
	wt := "async-wf-integration-start-workflow-test-type"
	tl := "async-wf-integration-start-workflow-test-tasklist"

	identity := "worker1"
	workflowType := &types.WorkflowType{
		Name: wt,
	}

	taskList := &types.TaskList{}
	taskList.Name = tl

	asyncReq := &types.StartWorkflowExecutionAsyncRequest{
		StartWorkflowExecutionRequest: &types.StartWorkflowExecutionRequest{
			RequestID:                           uuid.New(),
			Domain:                              s.domainName,
			WorkflowID:                          id,
			WorkflowType:                        workflowType,
			TaskList:                            taskList,
			Input:                               nil,
			ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(100),
			TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			Identity:                            identity,
		},
	}

	startTime := time.Now().UnixNano()
	ctx, cancel := createContext()
	defer cancel()
	_, err := s.engine.StartWorkflowExecutionAsync(ctx, asyncReq)
	s.Nil(err)

	for i := 0; i < 10; i++ {
		resp, err := s.engine.ListOpenWorkflowExecutions(ctx, &types.ListOpenWorkflowExecutionsRequest{
			Domain:          s.domainName,
			MaximumPageSize: 10,
			StartTimeFilter: &types.StartTimeFilter{
				EarliestTime: common.Int64Ptr(startTime),
				LatestTime:   common.Int64Ptr(time.Now().UnixNano()),
			},
			ExecutionFilter: &types.WorkflowExecutionFilter{
				WorkflowID: id,
			},
		})
		s.Nil(err)

		if len(resp.GetExecutions()) > 0 {
			execInfo := resp.GetExecutions()[0]
			s.NotNil(execInfo)
			s.logger.Info("Got execInfo", tag.Value(execInfo))
			break
		}

		s.logger.Info("No open workflow yet", tag.WorkflowID(id))
		time.Sleep(time.Second)
	}
}
