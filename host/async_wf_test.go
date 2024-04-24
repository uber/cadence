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

//go:build !race && asyncwfintegration
// +build !race,asyncwfintegration

/*
To run locally:

1. Stop the previous run if any

	docker-compose -f docker/buildkite/docker-compose-local-async-wf.yml down

2. Build the integration-test-async-wf image

	docker-compose -f docker/buildkite/docker-compose-local-async-wf.yml build integration-test-async-wf

3. Run the test in the docker container

	docker-compose -f docker/buildkite/docker-compose-local-async-wf.yml run --rm integration-test-async-wf

4. Full test run logs can be found at test.log file
*/
package host

import (
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/persistence"
	pt "github.com/uber/cadence/common/persistence/persistence-tests"
	"github.com/uber/cadence/common/types"

	_ "github.com/uber/cadence/common/asyncworkflow/queue/kafka" // needed to load kafka asyncworkflow queue
)

func TestAsyncWFIntegrationSuite(t *testing.T) {
	flag.Parse()

	confPath := "testdata/integration_async_wf_with_kafka_cluster.yaml"
	clusterConfig, err := GetTestClusterConfig(confPath)
	if err != nil {
		t.Fatalf("failed creating cluster config from %s, err: %v", confPath, err)
	}

	clusterConfig.TimeSource = clock.NewMockedTimeSource()
	clusterConfig.FrontendDynamicConfigOverrides = map[dynamicconfig.Key]interface{}{
		dynamicconfig.FrontendFailoverCoolDown:        time.Duration(0),
		dynamicconfig.EnableReadFromClosedExecutionV2: true,
	}

	testCluster := NewPersistenceTestCluster(t, clusterConfig)

	s := new(AsyncWFIntegrationSuite)
	params := IntegrationBaseParams{
		DefaultTestCluster:    testCluster,
		VisibilityTestCluster: testCluster,
		TestClusterConfig:     clusterConfig,
	}
	s.IntegrationBase = NewIntegrationBase(params)
	suite.Run(t, s)
}

func (s *AsyncWFIntegrationSuite) SetupSuite() {
	s.setupLogger()

	s.Logger.Info("Running integration test against test cluster")
	clusterMetadata := NewClusterMetadata(s.T(), s.testClusterConfig)
	dc := persistence.DynamicConfiguration{
		EnableCassandraAllConsistencyLevelDelete: dynamicconfig.GetBoolPropertyFn(true),
		PersistenceSampleLoggingRate:             dynamicconfig.GetIntPropertyFn(100),
		EnableShardIDMetrics:                     dynamicconfig.GetBoolPropertyFn(true),
	}
	params := pt.TestBaseParams{
		DefaultTestCluster:    s.defaultTestCluster,
		VisibilityTestCluster: s.visibilityTestCluster,
		ClusterMetadata:       clusterMetadata,
		DynamicConfiguration:  dc,
	}
	cluster, err := NewCluster(s.T(), s.testClusterConfig, s.Logger, params)
	s.Require().NoError(err)
	s.testCluster = cluster
	s.engine = s.testCluster.GetFrontendClient()
	s.adminClient = s.testCluster.GetAdminClient()

	s.domainName = s.randomizeStr("integration-test-domain")
	s.Require().NoError(s.registerDomain(s.domainName, 1, types.ArchivalStatusDisabled, "", types.ArchivalStatusDisabled, ""))
	s.secondaryDomainName = s.randomizeStr("unused-test-domain")
	s.Require().NoError(s.registerDomain(s.secondaryDomainName, 1, types.ArchivalStatusDisabled, "", types.ArchivalStatusDisabled, ""))

	s.domainCacheRefresh()
}

func (s *AsyncWFIntegrationSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *AsyncWFIntegrationSuite) TearDownSuite() {
	s.tearDownSuite()
}

func (s *AsyncWFIntegrationSuite) TestStartWorkflowExecutionAsync() {
	tests := []struct {
		name             string
		wantStartFailure bool
		asyncWFCfg       *types.AsyncWorkflowConfiguration
		secondaryCfg     *types.AsyncWorkflowConfiguration
	}{
		{
			name:             "start workflow execution async fails because domain missing async queue",
			wantStartFailure: true,
			secondaryCfg: &types.AsyncWorkflowConfiguration{
				Enabled:             true,
				PredefinedQueueName: "test-async-wf-queue",
			},
		},
		{
			name: "start workflow execution async fails because async queue is disabled",
			asyncWFCfg: &types.AsyncWorkflowConfiguration{
				Enabled: false,
			},
			wantStartFailure: true,
			secondaryCfg: &types.AsyncWorkflowConfiguration{
				Enabled:             true,
				PredefinedQueueName: "test-async-wf-queue",
			},
		},
		{
			name: "start workflow execution async succeeds and workflow starts",
			asyncWFCfg: &types.AsyncWorkflowConfiguration{
				Enabled:             true,
				PredefinedQueueName: "test-async-wf-queue",
			},
			secondaryCfg: &types.AsyncWorkflowConfiguration{
				Enabled:             false,
				PredefinedQueueName: "test-async-wf-queue",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			// advance the time so each test has a unique start time
			s.testClusterConfig.TimeSource.Advance(time.Second)

			ctx, cancel := createContext()
			defer cancel()

			if tc.asyncWFCfg != nil {
				_, err := s.adminClient.UpdateDomainAsyncWorkflowConfiguraton(ctx, &types.UpdateDomainAsyncWorkflowConfiguratonRequest{
					Domain:        s.domainName,
					Configuration: tc.asyncWFCfg,
				})
				if err != nil {
					t.Fatalf("UpdateDomainAsyncWorkflowConfiguraton() failed: %v", err)
				}

				s.domainCacheRefresh()
			}

			if tc.secondaryCfg != nil {
				_, err := s.adminClient.UpdateDomainAsyncWorkflowConfiguraton(ctx, &types.UpdateDomainAsyncWorkflowConfiguratonRequest{
					Domain:        s.secondaryDomainName,
					Configuration: tc.secondaryCfg,
				})
				if err != nil {
					t.Fatalf("UpdateDomainAsyncWorkflowConfiguraton() failed: %v", err)
				}

				s.domainCacheRefresh()
			}

			startTime := s.testClusterConfig.TimeSource.Now().UnixNano()
			wfID := fmt.Sprintf("async-wf-integration-start-workflow-test-%d", startTime)
			wfType := "async-wf-integration-start-workflow-test-type"
			taskList := "async-wf-integration-start-workflow-test-tasklist"
			identity := "worker1"

			asyncReq := &types.StartWorkflowExecutionAsyncRequest{
				StartWorkflowExecutionRequest: &types.StartWorkflowExecutionRequest{
					RequestID:  uuid.New(),
					Domain:     s.domainName,
					WorkflowID: wfID,
					WorkflowType: &types.WorkflowType{
						Name: wfType,
					},
					TaskList: &types.TaskList{
						Name: taskList,
					},
					Input:                               nil,
					ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(100),
					TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
					Identity:                            identity,
				},
			}

			_, err := s.engine.StartWorkflowExecutionAsync(ctx, asyncReq)
			if tc.wantStartFailure != (err != nil) {
				t.Errorf("StartWorkflowExecutionAsync() failed: %v, wantStartFailure: %v", err, tc.wantStartFailure)
			}

			if err != nil || tc.wantStartFailure {
				return
			}

			// there's no worker or poller for async workflow, so we just validate whether it started.
			// this is sufficient to verify the async workflow start path.
			for i := 0; i < 30; i++ {
				resp, err := s.engine.DescribeWorkflowExecution(ctx, &types.DescribeWorkflowExecutionRequest{
					Domain: s.domainName,
					Execution: &types.WorkflowExecution{
						WorkflowID: wfID,
					},
				})

				if err != nil {
					t.Logf("Workflow execution not found yet. DescribeWorkflowExecution() returned err: %v", err)
					time.Sleep(time.Second)
					s.testClusterConfig.TimeSource.Advance(time.Second)
					continue
				}
				if resp.GetWorkflowExecutionInfo() != nil {
					t.Logf("DescribeWorkflowExecution() found the execution: %#v", resp.GetWorkflowExecutionInfo())
					return
				}
			}

			t.Fatal("Async started workflow not found")
		})
	}
}
