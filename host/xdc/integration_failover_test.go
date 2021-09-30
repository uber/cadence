// Copyright (c) 2017 Uber Technologies, Inc.
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

//go:build !race
// +build !race

// need to run xdc tests with race detector off because of ringpop bug causing data race issue

package xdc

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/environment"
	"github.com/uber/cadence/host"
)

type (
	integrationClustersTestSuite struct {
		// override suite.Suite.Assertions with require.Assertions; this means that s.NotNil(nil) will stop the test,
		// not merely log an error
		*require.Assertions
		suite.Suite
		cluster1 *host.TestCluster
		cluster2 *host.TestCluster
		logger   log.Logger
	}
)

const (
	cacheRefreshInterval = cache.DomainCacheRefreshInterval + 5*time.Second
)

var (
	clusterName              = []string{"active", "standby"}
	clusterReplicationConfig = []*types.ClusterReplicationConfiguration{
		{
			ClusterName: clusterName[0],
		},
		{
			ClusterName: clusterName[1],
		},
	}
)

func createContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 90*time.Second)
	return ctx
}

func TestIntegrationClustersTestSuite(t *testing.T) {
	flag.Parse()
	// TODO: Suite is disabled since it was intented to be deprecated.
	// Until we have a certain decision, disabling the suite run
	// suite.Run(t, new(integrationClustersTestSuite))
}

func (s *integrationClustersTestSuite) SetupSuite() {
	zapLogger, err := zap.NewDevelopment()
	// cannot use s.Nil since it is not initialized
	s.Require().NoError(err)
	s.logger = loggerimpl.NewLogger(zapLogger)

	fileName := "../testdata/xdc_integration_test_clusters.yaml"
	if host.TestFlags.TestClusterConfigFile != "" {
		fileName = host.TestFlags.TestClusterConfigFile
	}
	environment.SetupEnv()

	confContent, err := ioutil.ReadFile(fileName)
	s.Require().NoError(err)
	confContent = []byte(os.ExpandEnv(string(confContent)))

	var clusterConfigs []*host.TestClusterConfig
	s.Require().NoError(yaml.Unmarshal(confContent, &clusterConfigs))

	/*
		// TODO: following lines are failing build; it was introduced after integration tests refactor.
		// Looks like this test is deprecated, decide if we want to delete the whole test.
		// Commenting the build-failing parts until we have a decision
		c, err := host.NewCluster(clusterConfigs[0], s.logger.WithTags(tag.ClusterName(clusterName[0])))
		s.Require().NoError(err)
		s.cluster1 = c

		c, err = host.NewCluster(clusterConfigs[1], s.logger.WithTags(tag.ClusterName(clusterName[1])))
		s.Require().NoError(err)
		s.cluster2 = c
	*/
}

func (s *integrationClustersTestSuite) SetupTest() {
	// Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
	s.Assertions = require.New(s.T())
}

func (s *integrationClustersTestSuite) TearDownSuite() {
	s.cluster1.TearDownCluster()
	s.cluster2.TearDownCluster()
}

func (s *integrationClustersTestSuite) TestDomainFailover() {
	domainName := "test-domain-for-fail-over-" + common.GenerateRandomString(5)
	client1 := s.cluster1.GetFrontendClient() // active
	regReq := &types.RegisterDomainRequest{
		Name:                                   domainName,
		IsGlobalDomain:                         true,
		Clusters:                               clusterReplicationConfig,
		ActiveClusterName:                      clusterName[0],
		WorkflowExecutionRetentionPeriodInDays: 7,
	}
	err := client1.RegisterDomain(createContext(), regReq)
	s.NoError(err)

	descReq := &types.DescribeDomainRequest{
		Name: common.StringPtr(domainName),
	}
	resp, err := client1.DescribeDomain(createContext(), descReq)
	s.NoError(err)
	s.NotNil(resp)
	// Wait for domain cache to pick the change
	time.Sleep(cacheRefreshInterval)

	client2 := s.cluster2.GetFrontendClient() // standby
	resp2, err := client2.DescribeDomain(createContext(), descReq)
	s.NoError(err)
	s.NotNil(resp2)
	s.Equal(resp, resp2)

	// update domain to fail over
	updateReq := &types.UpdateDomainRequest{
		Name:              domainName,
		ActiveClusterName: common.StringPtr(clusterName[1]),
	}
	updateResp, err := client1.UpdateDomain(createContext(), updateReq)
	s.NoError(err)
	s.NotNil(updateResp)
	s.Equal(clusterName[1], updateResp.ReplicationConfiguration.GetActiveClusterName())
	s.Equal(int64(1), updateResp.GetFailoverVersion())

	updated := false
	var resp3 *types.DescribeDomainResponse
	for i := 0; i < 30; i++ {
		resp3, err = client2.DescribeDomain(createContext(), descReq)
		s.NoError(err)
		if resp3.ReplicationConfiguration.GetActiveClusterName() == clusterName[1] {
			updated = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.True(updated)
	s.NotNil(resp3)
	s.Equal(int64(1), resp3.GetFailoverVersion())

	// start workflow in new cluster
	id := "integration-domain-failover-test"
	wt := "integration-domain-failover-test-type"
	tl := "integration-domain-failover-test-tasklist"
	identity := "worker1"
	workflowType := &types.WorkflowType{Name: wt}
	taskList := &types.TaskList{Name: tl}
	startReq := &types.StartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              domainName,
		WorkflowID:                          id,
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(100),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		Identity:                            identity,
	}
	var we *types.StartWorkflowExecutionResponse
	for i := 0; i < 30; i++ {
		we, err = client2.StartWorkflowExecution(createContext(), startReq)
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	s.NoError(err)
	s.NotNil(we.GetRunID())
}

func (s *integrationClustersTestSuite) TestSimpleWorkflowFailover() {
	domainName := "test-simple-workflow-failover-" + common.GenerateRandomString(5)
	client1 := s.cluster1.GetFrontendClient() // active
	regReq := &types.RegisterDomainRequest{
		Name:                                   domainName,
		IsGlobalDomain:                         true,
		Clusters:                               clusterReplicationConfig,
		ActiveClusterName:                      clusterName[0],
		WorkflowExecutionRetentionPeriodInDays: 1,
	}
	err := client1.RegisterDomain(createContext(), regReq)
	s.NoError(err)

	descReq := &types.DescribeDomainRequest{
		Name: common.StringPtr(domainName),
	}
	resp, err := client1.DescribeDomain(createContext(), descReq)
	s.NoError(err)
	s.NotNil(resp)
	// Wait for domain cache to pick the change
	time.Sleep(cache.DomainCacheRefreshInterval)

	client2 := s.cluster2.GetFrontendClient() // standby
	resp2, err := client2.DescribeDomain(createContext(), descReq)
	s.NoError(err)
	s.NotNil(resp2)
	s.Equal(resp, resp2)

	// start a workflow
	id := "integration-simple-workflow-failover-test"
	wt := "integration-simple-workflow-failover-test-type"
	tl := "integration-simple-workflow-failover-test-tasklist"
	identity := "worker1"
	workflowType := &types.WorkflowType{Name: wt}
	taskList := &types.TaskList{Name: tl}
	startReq := &types.StartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              domainName,
		WorkflowID:                          id,
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(100),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		Identity:                            identity,
	}
	we, err := client1.StartWorkflowExecution(createContext(), startReq)
	s.Nil(err)
	s.NotNil(we.GetRunID())
	rid := we.GetRunID()

	s.logger.Info("StartWorkflowExecution \n", tag.WorkflowRunID(we.GetRunID()))

	workflowComplete := false
	activityName := "activity_type1"
	activityCount := int32(1)
	activityCounter := int32(0)
	dtHandler := func(execution *types.WorkflowExecution, wt *types.WorkflowType,
		previousStartedEventID, startedEventID int64, history *types.History) ([]byte, []*types.Decision, error) {
		if activityCounter < activityCount {
			activityCounter++
			buf := new(bytes.Buffer)
			s.Nil(binary.Write(buf, binary.LittleEndian, activityCounter))

			return []byte(strconv.Itoa(int(activityCounter))), []*types.Decision{{
				DecisionType: types.DecisionTypeScheduleActivityTask.Ptr(),
				ScheduleActivityTaskDecisionAttributes: &types.ScheduleActivityTaskDecisionAttributes{
					ActivityID:                    strconv.Itoa(int(activityCounter)),
					ActivityType:                  &types.ActivityType{Name: activityName},
					TaskList:                      &types.TaskList{Name: tl},
					Input:                         buf.Bytes(),
					ScheduleToCloseTimeoutSeconds: common.Int32Ptr(100),
					ScheduleToStartTimeoutSeconds: common.Int32Ptr(30),
					StartToCloseTimeoutSeconds:    common.Int32Ptr(50),
					HeartbeatTimeoutSeconds:       common.Int32Ptr(20),
				},
			}}, nil
		}

		workflowComplete = true
		return []byte(strconv.Itoa(int(activityCounter))), []*types.Decision{{
			DecisionType: types.DecisionTypeCompleteWorkflowExecution.Ptr(),
			CompleteWorkflowExecutionDecisionAttributes: &types.CompleteWorkflowExecutionDecisionAttributes{
				Result: []byte("Done."),
			},
		}}, nil
	}

	atHandler := func(execution *types.WorkflowExecution, activityType *types.ActivityType,
		activityID string, input []byte, taskToken []byte) ([]byte, bool, error) {

		return []byte("Activity Result."), false, nil
	}

	queryType := "test-query"
	queryHandler := func(task *types.PollForDecisionTaskResponse) ([]byte, error) {
		s.NotNil(task.Query)
		s.NotNil(task.Query.QueryType)
		if task.Query.QueryType == queryType {
			return []byte("query-result"), nil
		}

		return nil, errors.New("unknown-query-type")
	}

	poller := host.TaskPoller{
		Engine:          client1,
		Domain:          domainName,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		ActivityHandler: atHandler,
		QueryHandler:    queryHandler,
		Logger:          s.logger,
		T:               s.T(),
	}

	poller2 := host.TaskPoller{
		Engine:          client2,
		Domain:          domainName,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		ActivityHandler: atHandler,
		QueryHandler:    queryHandler,
		Logger:          s.logger,
		T:               s.T(),
	}

	// make some progress in cluster 1
	_, err = poller.PollAndProcessDecisionTask(false, false)
	s.logger.Info("PollAndProcessDecisionTask", tag.Error(err))
	s.Nil(err)

	type QueryResult struct {
		Resp *types.QueryWorkflowResponse
		Err  error
	}
	queryResultCh := make(chan QueryResult)
	queryWorkflowFn := func(client frontend.Client, queryType string) {
		queryResp, err := client.QueryWorkflow(createContext(), &types.QueryWorkflowRequest{
			Domain: domainName,
			Execution: &types.WorkflowExecution{
				WorkflowID: id,
				RunID:      we.RunID,
			},
			Query: &types.WorkflowQuery{
				QueryType: queryType,
			},
		})
		queryResultCh <- QueryResult{Resp: queryResp, Err: err}
	}

	// call QueryWorkflow in separate goroutinue (because it is blocking). That will generate a query task
	go queryWorkflowFn(client1, queryType)
	// process that query task, which should respond via RespondQueryTaskCompleted
	for {
		// loop until process the query task
		isQueryTask, errInner := poller.PollAndProcessDecisionTask(false, false)
		s.logger.Info("PollAndProcessQueryTask", tag.Error(err))
		s.Nil(errInner)
		if isQueryTask {
			break
		}
	}
	// wait until query result is ready
	queryResult := <-queryResultCh
	s.NoError(queryResult.Err)
	s.NotNil(queryResult.Resp)
	s.NotNil(queryResult.Resp.QueryResult)
	queryResultString := string(queryResult.Resp.QueryResult)
	s.Equal("query-result", queryResultString)

	// Wait a while so the events are replicated.
	time.Sleep(5 * time.Second)

	// call QueryWorkflow in separate goroutinue (because it is blocking). That will generate a query task
	go queryWorkflowFn(client2, queryType)
	// process that query task, which should respond via RespondQueryTaskCompleted
	for {
		// loop until process the query task
		isQueryTask, errInner := poller2.PollAndProcessDecisionTask(false, false)
		s.logger.Info("PollAndProcessQueryTask", tag.Error(err))
		s.Nil(errInner)
		if isQueryTask {
			break
		}
	}
	// wait until query result is ready
	queryResult = <-queryResultCh
	s.NoError(queryResult.Err)
	s.NotNil(queryResult.Resp)
	s.NotNil(queryResult.Resp.QueryResult)
	queryResultString = string(queryResult.Resp.QueryResult)
	s.Equal("query-result", queryResultString)

	// update domain to fail over
	updateReq := &types.UpdateDomainRequest{
		Name:              domainName,
		ActiveClusterName: common.StringPtr(clusterName[1]),
	}
	updateResp, err := client1.UpdateDomain(createContext(), updateReq)
	s.NoError(err)
	s.NotNil(updateResp)
	s.Equal(clusterName[1], updateResp.ReplicationConfiguration.GetActiveClusterName())
	s.Equal(int64(1), updateResp.GetFailoverVersion())

	// wait till failover completed
	time.Sleep(cacheRefreshInterval)

	// check history matched
	getHistoryReq := &types.GetWorkflowExecutionHistoryRequest{
		Domain: domainName,
		Execution: &types.WorkflowExecution{
			WorkflowID: id,
			RunID:      rid,
		},
	}
	var historyResponse *types.GetWorkflowExecutionHistoryResponse
	eventsReplicated := false
	for i := 0; i < 15; i++ {
		historyResponse, err = client2.GetWorkflowExecutionHistory(createContext(), getHistoryReq)
		if err == nil && len(historyResponse.History.Events) == 5 {
			eventsReplicated = true
			break
		}
		time.Sleep(1 * time.Second)
	}
	s.Nil(err)
	s.True(eventsReplicated)

	// Make sure query is still working after failover
	// call QueryWorkflow in separate goroutinue (because it is blocking). That will generate a query task
	go queryWorkflowFn(client1, queryType)
	// process that query task, which should respond via RespondQueryTaskCompleted
	for {
		// loop until process the query task
		isQueryTask, errInner := poller.PollAndProcessDecisionTask(false, false)
		s.logger.Info("PollAndProcessDecisionTask", tag.Error(err))
		s.Nil(errInner)
		if isQueryTask {
			break
		}
	}
	// wait until query result is ready
	queryResult = <-queryResultCh
	s.NoError(queryResult.Err)
	s.NotNil(queryResult.Resp)
	s.NotNil(queryResult.Resp.QueryResult)
	queryResultString = string(queryResult.Resp.QueryResult)
	s.Equal("query-result", queryResultString)

	// call QueryWorkflow in separate goroutinue (because it is blocking). That will generate a query task
	go queryWorkflowFn(client2, queryType)
	// process that query task, which should respond via RespondQueryTaskCompleted
	for {
		// loop until process the query task
		isQueryTask, errInner := poller2.PollAndProcessDecisionTask(false, false)
		s.logger.Info("PollAndProcessDecisionTask", tag.Error(err))
		s.Nil(errInner)
		if isQueryTask {
			break
		}
	}
	// wait until query result is ready
	queryResult = <-queryResultCh
	s.NoError(queryResult.Err)
	s.NotNil(queryResult.Resp)
	s.NotNil(queryResult.Resp.QueryResult)
	queryResultString = string(queryResult.Resp.QueryResult)
	s.Equal("query-result", queryResultString)

	// make process in cluster 2
	err = poller2.PollAndProcessActivityTask(false)
	s.logger.Info("PollAndProcessActivityTask 2", tag.Error(err))
	s.Nil(err)

	s.False(workflowComplete)
	_, err = poller2.PollAndProcessDecisionTask(false, false)
	s.logger.Info("PollAndProcessDecisionTask 2", tag.Error(err))
	s.Nil(err)
	s.True(workflowComplete)

	// check history replicated in cluster 1
	eventsReplicated = false
	for i := 0; i < 15; i++ {
		historyResponse, err = client1.GetWorkflowExecutionHistory(createContext(), getHistoryReq)
		if err == nil && len(historyResponse.History.Events) == 11 {
			eventsReplicated = true
			break
		}
		time.Sleep(1 * time.Second)
	}
	s.Nil(err)
	s.True(eventsReplicated)
}

func (s *integrationClustersTestSuite) TestStickyDecisionFailover() {
	domainName := "test-sticky-decision-workflow-failover-" + common.GenerateRandomString(5)
	client1 := s.cluster1.GetFrontendClient() // active
	regReq := &types.RegisterDomainRequest{
		Name:                                   domainName,
		IsGlobalDomain:                         true,
		Clusters:                               clusterReplicationConfig,
		ActiveClusterName:                      clusterName[0],
		WorkflowExecutionRetentionPeriodInDays: 1,
	}
	err := client1.RegisterDomain(createContext(), regReq)
	s.NoError(err)

	descReq := &types.DescribeDomainRequest{
		Name: common.StringPtr(domainName),
	}
	resp, err := client1.DescribeDomain(createContext(), descReq)
	s.NoError(err)
	s.NotNil(resp)
	// Wait for domain cache to pick the change
	time.Sleep(cacheRefreshInterval)

	client2 := s.cluster2.GetFrontendClient() // standby

	// Start a workflow
	id := "integration-sticky-decision-workflow-failover-test"
	wt := "integration-sticky-decision-workflow-failover-test-type"
	tl := "integration-sticky-decision-workflow-failover-test-tasklist"
	stl1 := "integration-sticky-decision-workflow-failover-test-tasklist-sticky1"
	stl2 := "integration-sticky-decision-workflow-failover-test-tasklist-sticky2"
	identity1 := "worker1"
	identity2 := "worker2"

	workflowType := &types.WorkflowType{Name: wt}
	taskList := &types.TaskList{Name: tl}
	stickyTaskList1 := &types.TaskList{Name: stl1}
	stickyTaskList2 := &types.TaskList{Name: stl2}
	stickyTaskTimeout := common.Int32Ptr(100)
	startReq := &types.StartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              domainName,
		WorkflowID:                          id,
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2592000),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(60),
		Identity:                            identity1,
	}
	we, err := client1.StartWorkflowExecution(createContext(), startReq)
	s.NoError(err)
	s.NotNil(we.GetRunID())

	s.logger.Info("StartWorkflowExecution", tag.WorkflowRunID(we.GetRunID()))

	firstDecisionMade := false
	secondDecisionMade := false
	workflowCompleted := false
	dtHandler := func(execution *types.WorkflowExecution, wt *types.WorkflowType,
		previousStartedEventID, startedEventID int64, history *types.History) ([]byte, []*types.Decision, error) {
		if !firstDecisionMade {
			firstDecisionMade = true
			return nil, []*types.Decision{}, nil
		}

		if !secondDecisionMade {
			secondDecisionMade = true
			return nil, []*types.Decision{}, nil
		}

		workflowCompleted = true
		return nil, []*types.Decision{{
			DecisionType: types.DecisionTypeCompleteWorkflowExecution.Ptr(),
			CompleteWorkflowExecutionDecisionAttributes: &types.CompleteWorkflowExecutionDecisionAttributes{
				Result: []byte("Done."),
			},
		}}, nil
	}

	poller1 := &host.TaskPoller{
		Engine:                              client1,
		Domain:                              domainName,
		TaskList:                            taskList,
		StickyTaskList:                      stickyTaskList1,
		StickyScheduleToStartTimeoutSeconds: stickyTaskTimeout,
		Identity:                            identity1,
		DecisionHandler:                     dtHandler,
		Logger:                              s.logger,
		T:                                   s.T(),
	}

	poller2 := &host.TaskPoller{
		Engine:                              client2,
		Domain:                              domainName,
		TaskList:                            taskList,
		StickyTaskList:                      stickyTaskList2,
		StickyScheduleToStartTimeoutSeconds: stickyTaskTimeout,
		Identity:                            identity2,
		DecisionHandler:                     dtHandler,
		Logger:                              s.logger,
		T:                                   s.T(),
	}

	_, err = poller1.PollAndProcessDecisionTaskWithAttemptAndRetry(false, false, false, true, 0, 5)
	s.logger.Info("PollAndProcessDecisionTask", tag.Error(err))
	s.Nil(err)
	s.True(firstDecisionMade)

	// Send a signal in cluster
	signalName := "my signal"
	signalInput := []byte("my signal input.")
	err = client1.SignalWorkflowExecution(createContext(), &types.SignalWorkflowExecutionRequest{
		Domain: domainName,
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: id,
			RunID:      we.GetRunID(),
		},
		SignalName: signalName,
		Input:      signalInput,
		Identity:   identity1,
	})
	s.Nil(err)

	// Update domain to fail over
	updateReq := &types.UpdateDomainRequest{
		Name:              domainName,
		ActiveClusterName: common.StringPtr(clusterName[1]),
	}
	updateResp, err := client1.UpdateDomain(createContext(), updateReq)
	s.NoError(err)
	s.NotNil(updateResp)
	s.Equal(clusterName[1], updateResp.ReplicationConfiguration.GetActiveClusterName())
	s.Equal(int64(1), updateResp.GetFailoverVersion())

	// Wait for domain cache to pick the change
	time.Sleep(cacheRefreshInterval)

	_, err = poller2.PollAndProcessDecisionTaskWithAttemptAndRetry(false, false, false, true, 0, 5)
	s.logger.Info("PollAndProcessDecisionTask", tag.Error(err))
	s.Nil(err)
	s.True(secondDecisionMade)

	err = client2.SignalWorkflowExecution(createContext(), &types.SignalWorkflowExecutionRequest{
		Domain: domainName,
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: id,
			RunID:      we.GetRunID(),
		},
		SignalName: signalName,
		Input:      signalInput,
		Identity:   identity2,
	})
	s.Nil(err)

	// Update domain to fail over back
	updateReq = &types.UpdateDomainRequest{
		Name:              domainName,
		ActiveClusterName: common.StringPtr(clusterName[0]),
	}
	updateResp, err = client2.UpdateDomain(createContext(), updateReq)
	s.NoError(err)
	s.NotNil(updateResp)
	s.Equal(clusterName[0], updateResp.ReplicationConfiguration.GetActiveClusterName())
	s.Equal(int64(10), updateResp.GetFailoverVersion())

	_, err = poller1.PollAndProcessDecisionTask(false, false)
	s.logger.Info("PollAndProcessDecisionTask", tag.Error(err))
	s.Nil(err)
	s.True(workflowCompleted)
}

func (s *integrationClustersTestSuite) TestStartWorkflowExecution_Failover_WorkflowIDReusePolicy() {
	domainName := "test-start-workflow-failover-ID-reuse-policy" + common.GenerateRandomString(5)
	client1 := s.cluster1.GetFrontendClient() // active
	regReq := &types.RegisterDomainRequest{
		Name:                                   domainName,
		IsGlobalDomain:                         true,
		Clusters:                               clusterReplicationConfig,
		ActiveClusterName:                      clusterName[0],
		WorkflowExecutionRetentionPeriodInDays: 1,
	}
	err := client1.RegisterDomain(createContext(), regReq)
	s.NoError(err)

	descReq := &types.DescribeDomainRequest{
		Name: common.StringPtr(domainName),
	}
	resp, err := client1.DescribeDomain(createContext(), descReq)
	s.NoError(err)
	s.NotNil(resp)
	// Wait for domain cache to pick the change
	time.Sleep(cache.DomainCacheRefreshInterval)

	client2 := s.cluster2.GetFrontendClient() // standby
	resp2, err := client2.DescribeDomain(createContext(), descReq)
	s.NoError(err)
	s.NotNil(resp2)
	s.Equal(resp, resp2)

	// start a workflow
	id := "integration-start-workflow-failover-ID-reuse-policy-test"
	wt := "integration-start-workflow-failover-ID-reuse-policy-test-type"
	tl := "integration-start-workflow-failover-ID-reuse-policy-test-tasklist"
	identity := "worker1"
	workflowType := &types.WorkflowType{Name: wt}
	taskList := &types.TaskList{Name: tl}
	startReq := &types.StartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              domainName,
		WorkflowID:                          id,
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(100),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		Identity:                            identity,
		WorkflowIDReusePolicy:               types.WorkflowIDReusePolicyAllowDuplicate.Ptr(),
	}
	we, err := client1.StartWorkflowExecution(createContext(), startReq)
	s.Nil(err)
	s.NotNil(we.GetRunID())
	s.logger.Info("StartWorkflowExecution in cluster 1: ", tag.WorkflowRunID(we.GetRunID()))

	workflowCompleteTimes := 0
	dtHandler := func(execution *types.WorkflowExecution, wt *types.WorkflowType,
		previousStartedEventID, startedEventID int64, history *types.History) ([]byte, []*types.Decision, error) {

		workflowCompleteTimes++
		return nil, []*types.Decision{{
			DecisionType: types.DecisionTypeCompleteWorkflowExecution.Ptr(),
			CompleteWorkflowExecutionDecisionAttributes: &types.CompleteWorkflowExecutionDecisionAttributes{
				Result: []byte("Done."),
			},
		}}, nil
	}

	poller := host.TaskPoller{
		Engine:          client1,
		Domain:          domainName,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		ActivityHandler: nil,
		Logger:          s.logger,
		T:               s.T(),
	}

	poller2 := host.TaskPoller{
		Engine:          client2,
		Domain:          domainName,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		ActivityHandler: nil,
		Logger:          s.logger,
		T:               s.T(),
	}

	// Complete the workflow in cluster 1
	_, err = poller.PollAndProcessDecisionTask(false, false)
	s.logger.Info("PollAndProcessDecisionTask", tag.Error(err))
	s.Nil(err)
	s.Equal(1, workflowCompleteTimes)

	// update domain to fail over
	updateReq := &types.UpdateDomainRequest{
		Name:              domainName,
		ActiveClusterName: common.StringPtr(clusterName[1]),
	}
	updateResp, err := client1.UpdateDomain(createContext(), updateReq)
	s.NoError(err)
	s.NotNil(updateResp)
	s.Equal(clusterName[1], updateResp.ReplicationConfiguration.GetActiveClusterName())
	s.Equal(int64(1), updateResp.GetFailoverVersion())

	// wait till failover completed
	time.Sleep(cacheRefreshInterval)

	// start the same workflow in cluster 2 is not allowed if policy is AllowDuplicateFailedOnly
	startReq.RequestID = uuid.New()
	startReq.WorkflowIDReusePolicy = types.WorkflowIDReusePolicyAllowDuplicateFailedOnly.Ptr()
	we, err = client2.StartWorkflowExecution(createContext(), startReq)
	s.IsType(&types.WorkflowExecutionAlreadyStartedError{}, err)
	s.Nil(we)

	// start the same workflow in cluster 2 is not allowed if policy is RejectDuplicate
	startReq.RequestID = uuid.New()
	startReq.WorkflowIDReusePolicy = types.WorkflowIDReusePolicyRejectDuplicate.Ptr()
	we, err = client2.StartWorkflowExecution(createContext(), startReq)
	s.IsType(&types.WorkflowExecutionAlreadyStartedError{}, err)
	s.Nil(we)

	// start the workflow in cluster 2
	startReq.RequestID = uuid.New()
	startReq.WorkflowIDReusePolicy = types.WorkflowIDReusePolicyAllowDuplicate.Ptr()
	we, err = client2.StartWorkflowExecution(createContext(), startReq)
	s.Nil(err)
	s.NotNil(we.GetRunID())
	s.logger.Info("StartWorkflowExecution in cluster 2: ", tag.WorkflowRunID(we.GetRunID()))

	_, err = poller2.PollAndProcessDecisionTask(false, false)
	s.logger.Info("PollAndProcessDecisionTask 2", tag.Error(err))
	s.Nil(err)
	s.Equal(2, workflowCompleteTimes)
}

func (s *integrationClustersTestSuite) TestTerminateFailover() {
	domainName := "test-terminate-workflow-failover-" + common.GenerateRandomString(5)
	client1 := s.cluster1.GetFrontendClient() // active
	regReq := &types.RegisterDomainRequest{
		Name:                                   domainName,
		IsGlobalDomain:                         true,
		Clusters:                               clusterReplicationConfig,
		ActiveClusterName:                      clusterName[0],
		WorkflowExecutionRetentionPeriodInDays: 1,
	}
	err := client1.RegisterDomain(createContext(), regReq)
	s.NoError(err)

	descReq := &types.DescribeDomainRequest{
		Name: common.StringPtr(domainName),
	}
	resp, err := client1.DescribeDomain(createContext(), descReq)
	s.NoError(err)
	s.NotNil(resp)
	// Wait for domain cache to pick the change
	time.Sleep(cacheRefreshInterval)

	client2 := s.cluster2.GetFrontendClient() // standby

	// start a workflow
	id := "integration-terminate-workflow-failover-test"
	wt := "integration-terminate-workflow-failover-test-type"
	tl := "integration-terminate-workflow-failover-test-tasklist"
	identity := "worker1"
	workflowType := &types.WorkflowType{Name: wt}
	taskList := &types.TaskList{Name: tl}
	startReq := &types.StartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              domainName,
		WorkflowID:                          id,
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(100),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		Identity:                            identity,
	}
	we, err := client1.StartWorkflowExecution(createContext(), startReq)
	s.NoError(err)
	s.NotNil(we.GetRunID())

	activityName := "activity_type1"
	activityCount := int32(1)
	activityCounter := int32(0)
	dtHandler := func(execution *types.WorkflowExecution, wt *types.WorkflowType,
		previousStartedEventID, startedEventID int64, history *types.History) ([]byte, []*types.Decision, error) {
		if activityCounter < activityCount {
			activityCounter++
			buf := new(bytes.Buffer)
			s.Nil(binary.Write(buf, binary.LittleEndian, activityCounter))

			return []byte(strconv.Itoa(int(activityCounter))), []*types.Decision{{
				DecisionType: types.DecisionTypeScheduleActivityTask.Ptr(),
				ScheduleActivityTaskDecisionAttributes: &types.ScheduleActivityTaskDecisionAttributes{
					ActivityID:                    strconv.Itoa(int(activityCounter)),
					ActivityType:                  &types.ActivityType{Name: activityName},
					TaskList:                      &types.TaskList{Name: tl},
					Input:                         buf.Bytes(),
					ScheduleToCloseTimeoutSeconds: common.Int32Ptr(100),
					ScheduleToStartTimeoutSeconds: common.Int32Ptr(10),
					StartToCloseTimeoutSeconds:    common.Int32Ptr(50),
					HeartbeatTimeoutSeconds:       common.Int32Ptr(5),
				},
			}}, nil
		}

		return []byte(strconv.Itoa(int(activityCounter))), []*types.Decision{{
			DecisionType: types.DecisionTypeCompleteWorkflowExecution.Ptr(),
			CompleteWorkflowExecutionDecisionAttributes: &types.CompleteWorkflowExecutionDecisionAttributes{
				Result: []byte("Done."),
			},
		}}, nil
	}

	atHandler := func(execution *types.WorkflowExecution, activityType *types.ActivityType,
		activityID string, input []byte, taskToken []byte) ([]byte, bool, error) {

		return []byte("Activity Result."), false, nil
	}

	poller := &host.TaskPoller{
		Engine:          client1,
		Domain:          domainName,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		ActivityHandler: atHandler,
		Logger:          s.logger,
		T:               s.T(),
	}

	// make some progress in cluster 1
	_, err = poller.PollAndProcessDecisionTask(false, false)
	s.logger.Info("PollAndProcessDecisionTask", tag.Error(err))
	s.Nil(err)

	// update domain to fail over
	updateReq := &types.UpdateDomainRequest{
		Name:              domainName,
		ActiveClusterName: common.StringPtr(clusterName[1]),
	}
	updateResp, err := client1.UpdateDomain(createContext(), updateReq)
	s.NoError(err)
	s.NotNil(updateResp)
	s.Equal(clusterName[1], updateResp.ReplicationConfiguration.GetActiveClusterName())
	s.Equal(int64(1), updateResp.GetFailoverVersion())

	// wait till failover completed
	time.Sleep(cacheRefreshInterval)

	// terminate workflow at cluster 2
	terminateReason := "terminate reason."
	terminateDetails := []byte("terminate details.")
	err = client2.TerminateWorkflowExecution(createContext(), &types.TerminateWorkflowExecutionRequest{
		Domain: domainName,
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: id,
		},
		Reason:   terminateReason,
		Details:  terminateDetails,
		Identity: identity,
	})
	s.Nil(err)

	// check terminate done
	executionTerminated := false
	getHistoryReq := &types.GetWorkflowExecutionHistoryRequest{
		Domain: domainName,
		Execution: &types.WorkflowExecution{
			WorkflowID: id,
		},
	}
GetHistoryLoop:
	for i := 0; i < 10; i++ {
		historyResponse, err := client2.GetWorkflowExecutionHistory(createContext(), getHistoryReq)
		s.Nil(err)
		history := historyResponse.History

		lastEvent := history.Events[len(history.Events)-1]
		if *lastEvent.EventType != types.EventTypeWorkflowExecutionTerminated {
			s.logger.Warn("Execution not terminated yet.")
			time.Sleep(100 * time.Millisecond)
			continue GetHistoryLoop
		}

		terminateEventAttributes := lastEvent.WorkflowExecutionTerminatedEventAttributes
		s.Equal(terminateReason, terminateEventAttributes.Reason)
		s.Equal(terminateDetails, terminateEventAttributes.Details)
		s.Equal(identity, terminateEventAttributes.Identity)
		executionTerminated = true
		break GetHistoryLoop
	}
	s.True(executionTerminated)

	// check history replicated to the other cluster
	var historyResponse *types.GetWorkflowExecutionHistoryResponse
	eventsReplicated := false
GetHistoryLoop2:
	for i := 0; i < 15; i++ {
		historyResponse, err = client1.GetWorkflowExecutionHistory(createContext(), getHistoryReq)
		if err == nil {
			history := historyResponse.History
			lastEvent := history.Events[len(history.Events)-1]
			if *lastEvent.EventType == types.EventTypeWorkflowExecutionTerminated {
				terminateEventAttributes := lastEvent.WorkflowExecutionTerminatedEventAttributes
				s.Equal(terminateReason, terminateEventAttributes.Reason)
				s.Equal(terminateDetails, terminateEventAttributes.Details)
				s.Equal(identity, terminateEventAttributes.Identity)
				eventsReplicated = true
				break GetHistoryLoop2
			}
		}
		time.Sleep(1 * time.Second)
	}
	s.Nil(err)
	s.True(eventsReplicated)
}

func (s *integrationClustersTestSuite) TestContinueAsNewFailover() {
	domainName := "test-continueAsNew-workflow-failover-" + common.GenerateRandomString(5)
	client1 := s.cluster1.GetFrontendClient() // active
	regReq := &types.RegisterDomainRequest{
		Name:                                   domainName,
		IsGlobalDomain:                         true,
		Clusters:                               clusterReplicationConfig,
		ActiveClusterName:                      clusterName[0],
		WorkflowExecutionRetentionPeriodInDays: 1,
	}
	err := client1.RegisterDomain(createContext(), regReq)
	s.NoError(err)

	descReq := &types.DescribeDomainRequest{
		Name: common.StringPtr(domainName),
	}
	resp, err := client1.DescribeDomain(createContext(), descReq)
	s.NoError(err)
	s.NotNil(resp)
	// Wait for domain cache to pick the change
	time.Sleep(cacheRefreshInterval)

	client2 := s.cluster2.GetFrontendClient() // standby

	// start a workflow
	id := "integration-continueAsNew-workflow-failover-test"
	wt := "integration-continueAsNew-workflow-failover-test-type"
	tl := "integration-continueAsNew-workflow-failover-test-tasklist"
	identity := "worker1"
	workflowType := &types.WorkflowType{Name: wt}
	taskList := &types.TaskList{Name: tl}
	startReq := &types.StartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              domainName,
		WorkflowID:                          id,
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(100),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		Identity:                            identity,
	}
	we, err := client1.StartWorkflowExecution(createContext(), startReq)
	s.NoError(err)
	s.NotNil(we.GetRunID())

	workflowComplete := false
	continueAsNewCount := int32(5)
	continueAsNewCounter := int32(0)
	var previousRunID string
	var lastRunStartedEvent *types.HistoryEvent
	dtHandler := func(execution *types.WorkflowExecution, wt *types.WorkflowType,
		previousStartedEventID, startedEventID int64, history *types.History) ([]byte, []*types.Decision, error) {
		if continueAsNewCounter < continueAsNewCount {
			previousRunID = execution.GetRunID()
			continueAsNewCounter++
			buf := new(bytes.Buffer)
			s.Nil(binary.Write(buf, binary.LittleEndian, continueAsNewCounter))

			return []byte(strconv.Itoa(int(continueAsNewCounter))), []*types.Decision{{
				DecisionType: types.DecisionTypeContinueAsNewWorkflowExecution.Ptr(),
				ContinueAsNewWorkflowExecutionDecisionAttributes: &types.ContinueAsNewWorkflowExecutionDecisionAttributes{
					WorkflowType:                        workflowType,
					TaskList:                            &types.TaskList{Name: tl},
					Input:                               buf.Bytes(),
					ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(100),
					TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(10),
				},
			}}, nil
		}

		lastRunStartedEvent = history.Events[0]
		workflowComplete = true
		return []byte(strconv.Itoa(int(continueAsNewCounter))), []*types.Decision{{
			DecisionType: types.DecisionTypeCompleteWorkflowExecution.Ptr(),
			CompleteWorkflowExecutionDecisionAttributes: &types.CompleteWorkflowExecutionDecisionAttributes{
				Result: []byte("Done."),
			},
		}}, nil
	}

	poller := &host.TaskPoller{
		Engine:          client1,
		Domain:          domainName,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		Logger:          s.logger,
		T:               s.T(),
	}

	poller2 := host.TaskPoller{
		Engine:          client2,
		Domain:          domainName,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		Logger:          s.logger,
		T:               s.T(),
	}

	// make some progress in cluster 1 and did some continueAsNew
	for i := 0; i < 3; i++ {
		_, err := poller.PollAndProcessDecisionTask(false, false)
		s.logger.Info("PollAndProcessDecisionTask", tag.Error(err))
		s.Nil(err, strconv.Itoa(i))
	}

	// update domain to fail over
	updateReq := &types.UpdateDomainRequest{
		Name:              domainName,
		ActiveClusterName: common.StringPtr(clusterName[1]),
	}
	updateResp, err := client1.UpdateDomain(createContext(), updateReq)
	s.NoError(err)
	s.NotNil(updateResp)
	s.Equal(clusterName[1], updateResp.ReplicationConfiguration.GetActiveClusterName())
	s.Equal(int64(1), updateResp.GetFailoverVersion())

	// wait till failover completed
	time.Sleep(cacheRefreshInterval)

	// finish the rest in cluster 2
	for i := 0; i < 2; i++ {
		_, err := poller2.PollAndProcessDecisionTask(false, false)
		s.logger.Info("PollAndProcessDecisionTask", tag.Error(err))
		s.Nil(err, strconv.Itoa(i))
	}

	s.False(workflowComplete)
	_, err = poller2.PollAndProcessDecisionTask(false, false)
	s.Nil(err)
	s.True(workflowComplete)
	s.Equal(previousRunID, lastRunStartedEvent.WorkflowExecutionStartedEventAttributes.GetContinuedExecutionRunID())
}

func (s *integrationClustersTestSuite) TestSignalFailover() {
	domainName := "test-signal-workflow-failover-" + common.GenerateRandomString(5)
	client1 := s.cluster1.GetFrontendClient() // active
	regReq := &types.RegisterDomainRequest{
		Name:                                   domainName,
		IsGlobalDomain:                         true,
		Clusters:                               clusterReplicationConfig,
		ActiveClusterName:                      clusterName[0],
		WorkflowExecutionRetentionPeriodInDays: 1,
	}
	err := client1.RegisterDomain(createContext(), regReq)
	s.NoError(err)

	descReq := &types.DescribeDomainRequest{
		Name: common.StringPtr(domainName),
	}
	resp, err := client1.DescribeDomain(createContext(), descReq)
	s.NoError(err)
	s.NotNil(resp)
	// Wait for domain cache to pick the change
	time.Sleep(cacheRefreshInterval)

	client2 := s.cluster2.GetFrontendClient() // standby

	// Start a workflow
	id := "integration-signal-workflow-failover-test"
	wt := "integration-signal-workflow-failover-test-type"
	tl := "integration-signal-workflow-failover-test-tasklist"
	identity := "worker1"
	workflowType := &types.WorkflowType{Name: wt}
	taskList := &types.TaskList{Name: tl}
	startReq := &types.StartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              domainName,
		WorkflowID:                          id,
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(300),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		Identity:                            identity,
	}
	we, err := client1.StartWorkflowExecution(createContext(), startReq)
	s.NoError(err)
	s.NotNil(we.GetRunID())

	s.logger.Info("StartWorkflowExecution", tag.WorkflowRunID(we.GetRunID()))

	eventSignaled := false
	dtHandler := func(execution *types.WorkflowExecution, wt *types.WorkflowType,
		previousStartedEventID, startedEventID int64, history *types.History) ([]byte, []*types.Decision, error) {
		if !eventSignaled {
			for _, event := range history.Events[previousStartedEventID:] {
				if *event.EventType == types.EventTypeWorkflowExecutionSignaled {
					eventSignaled = true
					return nil, []*types.Decision{}, nil
				}
			}
		}

		return nil, []*types.Decision{{
			DecisionType: types.DecisionTypeCompleteWorkflowExecution.Ptr(),
			CompleteWorkflowExecutionDecisionAttributes: &types.CompleteWorkflowExecutionDecisionAttributes{
				Result: []byte("Done."),
			},
		}}, nil
	}

	poller := &host.TaskPoller{
		Engine:          client1,
		Domain:          domainName,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		Logger:          s.logger,
		T:               s.T(),
	}

	poller2 := &host.TaskPoller{
		Engine:          client2,
		Domain:          domainName,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		Logger:          s.logger,
		T:               s.T(),
	}

	// Send a signal in cluster 1
	signalName := "my signal"
	signalInput := []byte("my signal input.")
	err = client1.SignalWorkflowExecution(createContext(), &types.SignalWorkflowExecutionRequest{
		Domain: domainName,
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: id,
			RunID:      we.GetRunID(),
		},
		SignalName: signalName,
		Input:      signalInput,
		Identity:   identity,
	})
	s.Nil(err)

	// Process signal in cluster 1
	s.False(eventSignaled)
	_, err = poller.PollAndProcessDecisionTask(false, false)
	s.logger.Info("PollAndProcessDecisionTask", tag.Error(err))
	s.Nil(err)
	s.True(eventSignaled)

	// Update domain to fail over
	updateReq := &types.UpdateDomainRequest{
		Name:              domainName,
		ActiveClusterName: common.StringPtr(clusterName[1]),
	}
	updateResp, err := client1.UpdateDomain(createContext(), updateReq)
	s.NoError(err)
	s.NotNil(updateResp)
	s.Equal(clusterName[1], updateResp.ReplicationConfiguration.GetActiveClusterName())
	s.Equal(int64(1), updateResp.GetFailoverVersion())

	// Wait for domain cache to pick the change
	time.Sleep(cacheRefreshInterval)

	// check history matched
	getHistoryReq := &types.GetWorkflowExecutionHistoryRequest{
		Domain: domainName,
		Execution: &types.WorkflowExecution{
			WorkflowID: id,
		},
	}
	var historyResponse *types.GetWorkflowExecutionHistoryResponse
	eventsReplicated := false
	for i := 0; i < 15; i++ {
		historyResponse, err = client2.GetWorkflowExecutionHistory(createContext(), getHistoryReq)
		if err == nil && len(historyResponse.History.Events) == 5 {
			eventsReplicated = true
			break
		}
		time.Sleep(1 * time.Second)
	}
	s.Nil(err)
	s.True(eventsReplicated)

	// Send another signal in cluster 2
	signalName2 := "my signal 2"
	signalInput2 := []byte("my signal input 2.")
	err = client2.SignalWorkflowExecution(createContext(), &types.SignalWorkflowExecutionRequest{
		Domain: domainName,
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: id,
		},
		SignalName: signalName2,
		Input:      signalInput2,
		Identity:   identity,
	})
	s.Nil(err)

	// Process signal in cluster 2
	eventSignaled = false
	_, err = poller2.PollAndProcessDecisionTask(false, false)
	s.logger.Info("PollAndProcessDecisionTask 2", tag.Error(err))
	s.Nil(err)
	s.True(eventSignaled)

	// check history matched
	eventsReplicated = false
	for i := 0; i < 15; i++ {
		historyResponse, err = client2.GetWorkflowExecutionHistory(createContext(), getHistoryReq)
		if err == nil && len(historyResponse.History.Events) == 9 {
			eventsReplicated = true
			break
		}
		time.Sleep(1 * time.Second)
	}
	s.Nil(err)
	s.True(eventsReplicated)
}

func (s *integrationClustersTestSuite) TestUserTimerFailover() {
	domainName := "test-user-timer-workflow-failover-" + common.GenerateRandomString(5)
	client1 := s.cluster1.GetFrontendClient() // active
	regReq := &types.RegisterDomainRequest{
		Name:                                   domainName,
		IsGlobalDomain:                         true,
		Clusters:                               clusterReplicationConfig,
		ActiveClusterName:                      clusterName[0],
		WorkflowExecutionRetentionPeriodInDays: 1,
	}
	err := client1.RegisterDomain(createContext(), regReq)
	s.NoError(err)

	descReq := &types.DescribeDomainRequest{
		Name: common.StringPtr(domainName),
	}
	resp, err := client1.DescribeDomain(createContext(), descReq)
	s.NoError(err)
	s.NotNil(resp)
	// Wait for domain cache to pick the change
	time.Sleep(cacheRefreshInterval)

	client2 := s.cluster2.GetFrontendClient() // standby

	// Start a workflow
	id := "integration-user-timer-workflow-failover-test"
	wt := "integration-user-timer-workflow-failover-test-type"
	tl := "integration-user-timer-workflow-failover-test-tasklist"
	identity := "worker1"
	workflowType := &types.WorkflowType{Name: wt}
	taskList := &types.TaskList{Name: tl}
	startReq := &types.StartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              domainName,
		WorkflowID:                          id,
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(300),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(10),
		Identity:                            identity,
	}
	var we *types.StartWorkflowExecutionResponse
	for i := 0; i < 10; i++ {
		we, err = client1.StartWorkflowExecution(createContext(), startReq)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	s.NoError(err)
	s.NotNil(we.GetRunID())

	s.logger.Info("StartWorkflowExecution", tag.WorkflowRunID(we.GetRunID()))

	timerCreated := false
	timerFired := false
	workflowCompleted := false
	dtHandler := func(execution *types.WorkflowExecution, wt *types.WorkflowType,
		previousStartedEventID, startedEventID int64, history *types.History) ([]byte, []*types.Decision, error) {

		if !timerCreated {
			timerCreated = true

			// Send a signal in cluster
			signalName := "my signal"
			signalInput := []byte("my signal input.")
			err = client1.SignalWorkflowExecution(createContext(), &types.SignalWorkflowExecutionRequest{
				Domain: domainName,
				WorkflowExecution: &types.WorkflowExecution{
					WorkflowID: id,
					RunID:      we.GetRunID(),
				},
				SignalName: signalName,
				Input:      signalInput,
				Identity:   "",
			})
			s.Nil(err)
			return nil, []*types.Decision{{
				DecisionType: types.DecisionTypeStartTimer.Ptr(),
				StartTimerDecisionAttributes: &types.StartTimerDecisionAttributes{
					TimerID:                   "timer-id",
					StartToFireTimeoutSeconds: common.Int64Ptr(2),
				},
			}}, nil
		}

		if !timerFired {
			resp, err := client2.GetWorkflowExecutionHistory(createContext(), &types.GetWorkflowExecutionHistoryRequest{
				Domain: domainName,
				Execution: &types.WorkflowExecution{
					WorkflowID: id,
					RunID:      we.GetRunID(),
				},
			})
			s.Nil(err)
			for _, event := range resp.History.Events {
				if event.GetEventType() == types.EventTypeTimerFired {
					timerFired = true
				}
			}
			if !timerFired {
				return nil, []*types.Decision{}, nil
			}
		}

		workflowCompleted = true
		return nil, []*types.Decision{{
			DecisionType: types.DecisionTypeCompleteWorkflowExecution.Ptr(),
			CompleteWorkflowExecutionDecisionAttributes: &types.CompleteWorkflowExecutionDecisionAttributes{
				Result: []byte("Done."),
			},
		}}, nil
	}

	poller1 := &host.TaskPoller{
		Engine:          client1,
		Domain:          domainName,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		Logger:          s.logger,
		T:               s.T(),
	}

	poller2 := &host.TaskPoller{
		Engine:          client2,
		Domain:          domainName,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		Logger:          s.logger,
		T:               s.T(),
	}

	for i := 0; i < 2; i++ {
		_, err = poller1.PollAndProcessDecisionTask(false, false)
		if err != nil {
			timerCreated = false
			continue
		}
		if timerCreated {
			break
		}
	}
	s.True(timerCreated)

	// Update domain to fail over
	updateReq := &types.UpdateDomainRequest{
		Name:              domainName,
		ActiveClusterName: common.StringPtr(clusterName[1]),
	}
	updateResp, err := client1.UpdateDomain(createContext(), updateReq)
	s.NoError(err)
	s.NotNil(updateResp)
	s.Equal(clusterName[1], updateResp.ReplicationConfiguration.GetActiveClusterName())
	s.Equal(int64(1), updateResp.GetFailoverVersion())

	// Wait for domain cache to pick the change
	time.Sleep(cacheRefreshInterval)

	for i := 1; i < 20; i++ {
		if !workflowCompleted {
			_, err = poller2.PollAndProcessDecisionTask(false, false)
			s.Nil(err)
			time.Sleep(time.Second)
		}
	}
}

func (s *integrationClustersTestSuite) TestActivityHeartbeatFailover() {
	domainName := "test-activity-heartbeat-workflow-failover-" + common.GenerateRandomString(5)
	client1 := s.cluster1.GetFrontendClient() // active
	regReq := &types.RegisterDomainRequest{
		Name:                                   domainName,
		IsGlobalDomain:                         true,
		Clusters:                               clusterReplicationConfig,
		ActiveClusterName:                      clusterName[0],
		WorkflowExecutionRetentionPeriodInDays: 1,
	}
	err := client1.RegisterDomain(createContext(), regReq)
	s.NoError(err)

	descReq := &types.DescribeDomainRequest{
		Name: common.StringPtr(domainName),
	}
	resp, err := client1.DescribeDomain(createContext(), descReq)
	s.NoError(err)
	s.NotNil(resp)
	// Wait for domain cache to pick the change
	time.Sleep(cacheRefreshInterval)

	client2 := s.cluster2.GetFrontendClient() // standby

	// Start a workflow
	id := "integration-activity-heartbeat-workflow-failover-test"
	wt := "integration-activity-heartbeat-workflow-failover-test-type"
	tl := "integration-activity-heartbeat-workflow-failover-test-tasklist"
	identity1 := "worker1"
	identity2 := "worker2"
	workflowType := &types.WorkflowType{Name: wt}
	taskList := &types.TaskList{Name: tl}
	startReq := &types.StartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              domainName,
		WorkflowID:                          id,
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(300),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(10),
		Identity:                            identity1,
	}
	var we *types.StartWorkflowExecutionResponse
	for i := 0; i < 10; i++ {
		we, err = client1.StartWorkflowExecution(createContext(), startReq)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	s.NoError(err)
	s.NotNil(we.GetRunID())

	s.logger.Info("StartWorkflowExecution", tag.WorkflowRunID(we.GetRunID()))

	activitySent := false
	dtHandler := func(execution *types.WorkflowExecution, wt *types.WorkflowType,
		previousStartedEventID, startedEventID int64, history *types.History) ([]byte, []*types.Decision, error) {
		if !activitySent {
			activitySent = true
			return nil, []*types.Decision{{
				DecisionType: types.DecisionTypeScheduleActivityTask.Ptr(),
				ScheduleActivityTaskDecisionAttributes: &types.ScheduleActivityTaskDecisionAttributes{
					ActivityID:                    "1",
					ActivityType:                  &types.ActivityType{Name: "some random activity type"},
					TaskList:                      &types.TaskList{Name: tl},
					Input:                         []byte("some random input"),
					ScheduleToCloseTimeoutSeconds: common.Int32Ptr(1000),
					ScheduleToStartTimeoutSeconds: common.Int32Ptr(1000),
					StartToCloseTimeoutSeconds:    common.Int32Ptr(1000),
					HeartbeatTimeoutSeconds:       common.Int32Ptr(3),
					RetryPolicy: &types.RetryPolicy{
						InitialIntervalInSeconds:    1,
						MaximumAttempts:             3,
						MaximumIntervalInSeconds:    1,
						NonRetriableErrorReasons:    []string{"bad-bug"},
						BackoffCoefficient:          1,
						ExpirationIntervalInSeconds: 100,
					},
				},
			}}, nil
		}

		return nil, []*types.Decision{{
			DecisionType: types.DecisionTypeCompleteWorkflowExecution.Ptr(),
			CompleteWorkflowExecutionDecisionAttributes: &types.CompleteWorkflowExecutionDecisionAttributes{
				Result: []byte("Done."),
			},
		}}, nil
	}

	// activity handler
	activity1Called := false
	heartbeatDetails := []byte("details")
	atHandler1 := func(execution *types.WorkflowExecution, activityType *types.ActivityType,
		activityID string, input []byte, taskToken []byte) ([]byte, bool, error) {
		activity1Called = true
		_, err = client1.RecordActivityTaskHeartbeat(createContext(), &types.RecordActivityTaskHeartbeatRequest{
			TaskToken: taskToken, Details: heartbeatDetails})
		s.Nil(err)
		time.Sleep(5 * time.Second)
		return []byte("Activity Result."), false, nil
	}

	// activity handler
	activity2Called := false
	atHandler2 := func(execution *types.WorkflowExecution, activityType *types.ActivityType,
		activityID string, input []byte, taskToken []byte) ([]byte, bool, error) {
		activity2Called = true
		return []byte("Activity Result."), false, nil
	}

	poller1 := &host.TaskPoller{
		Engine:          client1,
		Domain:          domainName,
		TaskList:        taskList,
		Identity:        identity1,
		DecisionHandler: dtHandler,
		ActivityHandler: atHandler1,
		Logger:          s.logger,
		T:               s.T(),
	}

	poller2 := &host.TaskPoller{
		Engine:          client2,
		Domain:          domainName,
		TaskList:        taskList,
		Identity:        identity2,
		DecisionHandler: dtHandler,
		ActivityHandler: atHandler2,
		Logger:          s.logger,
		T:               s.T(),
	}

	describeWorkflowExecution := func(client frontend.Client) (*types.DescribeWorkflowExecutionResponse, error) {
		return client.DescribeWorkflowExecution(createContext(), &types.DescribeWorkflowExecutionRequest{
			Domain: domainName,
			Execution: &types.WorkflowExecution{
				WorkflowID: id,
				RunID:      we.RunID,
			},
		})
	}

	_, err = poller1.PollAndProcessDecisionTask(false, false)
	s.Nil(err)
	err = poller1.PollAndProcessActivityTask(false)
	s.IsType(&types.EntityNotExistsError{}, err)

	// Update domain to fail over
	updateReq := &types.UpdateDomainRequest{
		Name:              domainName,
		ActiveClusterName: common.StringPtr(clusterName[1]),
	}
	updateResp, err := client1.UpdateDomain(createContext(), updateReq)
	s.NoError(err)
	s.NotNil(updateResp)
	s.Equal(clusterName[1], updateResp.ReplicationConfiguration.GetActiveClusterName())
	s.Equal(int64(1), updateResp.GetFailoverVersion())

	// Wait for domain cache to pick the change
	time.Sleep(cacheRefreshInterval)

	// Make sure the heartbeat details are sent to cluster2 even when the activity at cluster1
	// has heartbeat timeout. Also make sure the information is recorded when the activity state
	// is "Scheduled"
	dweResponse, err := describeWorkflowExecution(client2)
	s.Nil(err)
	pendingActivities := dweResponse.GetPendingActivities()
	s.Equal(1, len(pendingActivities))
	s.Equal(types.PendingActivityStateScheduled, pendingActivities[0].GetState())
	s.Equal(heartbeatDetails, pendingActivities[0].GetHeartbeatDetails())
	s.Equal("cadenceInternal:Timeout HEARTBEAT", pendingActivities[0].GetLastFailureReason())
	s.Equal(identity1, pendingActivities[0].GetLastWorkerIdentity())

	for i := 0; i < 10; i++ {
		poller2.PollAndProcessActivityTask(false)
		if activity2Called {
			break
		} else {
			time.Sleep(1 * time.Second)
		}
	}

	s.True(activity1Called)
	s.True(activity2Called)

	historyResponse, err := client2.GetWorkflowExecutionHistory(createContext(), &types.GetWorkflowExecutionHistoryRequest{
		Domain: domainName,
		Execution: &types.WorkflowExecution{
			WorkflowID: id,
		},
	})
	s.Nil(err)
	history := historyResponse.History

	activityRetryFound := false
	for _, event := range history.Events {
		if event.GetEventType() == types.EventTypeActivityTaskStarted {
			attribute := event.ActivityTaskStartedEventAttributes
			s.True(attribute.GetAttempt() > 0)
			activityRetryFound = true
		}
	}
	s.True(activityRetryFound)
}

func (s *integrationClustersTestSuite) TestTransientDecisionFailover() {
	domainName := "test-transient-decision-workflow-failover-" + common.GenerateRandomString(5)
	client1 := s.cluster1.GetFrontendClient() // active
	regReq := &types.RegisterDomainRequest{
		Name:                                   domainName,
		IsGlobalDomain:                         true,
		Clusters:                               clusterReplicationConfig,
		ActiveClusterName:                      clusterName[0],
		WorkflowExecutionRetentionPeriodInDays: 1,
	}
	err := client1.RegisterDomain(createContext(), regReq)
	s.NoError(err)

	descReq := &types.DescribeDomainRequest{
		Name: common.StringPtr(domainName),
	}
	resp, err := client1.DescribeDomain(createContext(), descReq)
	s.NoError(err)
	s.NotNil(resp)
	// Wait for domain cache to pick the change
	time.Sleep(cacheRefreshInterval)

	client2 := s.cluster2.GetFrontendClient() // standby

	// Start a workflow
	id := "integration-transient-decision-workflow-failover-test"
	wt := "integration-transient-decision-workflow-failover-test-type"
	tl := "integration-transient-decision-workflow-failover-test-tasklist"
	identity := "worker1"
	workflowType := &types.WorkflowType{Name: wt}
	taskList := &types.TaskList{Name: tl}
	startReq := &types.StartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              domainName,
		WorkflowID:                          id,
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(300),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(8),
		Identity:                            identity,
	}
	var we *types.StartWorkflowExecutionResponse
	for i := 0; i < 10; i++ {
		we, err = client1.StartWorkflowExecution(createContext(), startReq)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	s.NoError(err)
	s.NotNil(we.GetRunID())

	s.logger.Info("StartWorkflowExecution", tag.WorkflowRunID(we.GetRunID()))

	decisionFailed := false
	workflowFinished := false
	dtHandler := func(execution *types.WorkflowExecution, wt *types.WorkflowType,
		previousStartedEventID, startedEventID int64, history *types.History) ([]byte, []*types.Decision, error) {
		if !decisionFailed {
			decisionFailed = true
			return nil, nil, errors.New("random fail decision reason")
		}

		workflowFinished = true
		return nil, []*types.Decision{{
			DecisionType: types.DecisionTypeCompleteWorkflowExecution.Ptr(),
			CompleteWorkflowExecutionDecisionAttributes: &types.CompleteWorkflowExecutionDecisionAttributes{
				Result: []byte("Done."),
			},
		}}, nil
	}

	poller1 := &host.TaskPoller{
		Engine:          client1,
		Domain:          domainName,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		Logger:          s.logger,
		T:               s.T(),
	}

	poller2 := &host.TaskPoller{
		Engine:          client2,
		Domain:          domainName,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		Logger:          s.logger,
		T:               s.T(),
	}

	// this will fail the decision
	_, err = poller1.PollAndProcessDecisionTask(false, false)
	s.Nil(err)

	// Update domain to fail over
	updateReq := &types.UpdateDomainRequest{
		Name:              domainName,
		ActiveClusterName: common.StringPtr(clusterName[1]),
	}
	updateResp, err := client1.UpdateDomain(createContext(), updateReq)
	s.NoError(err)
	s.NotNil(updateResp)
	s.Equal(clusterName[1], updateResp.ReplicationConfiguration.GetActiveClusterName())
	s.Equal(int64(1), updateResp.GetFailoverVersion())

	// Wait for domain cache to pick the change
	time.Sleep(cacheRefreshInterval)

	// for failover transient decision, it is guaranteed that the transient decision
	// after the failover has attempt 0
	// for details see ReplicateTransientDecisionTaskScheduled
	_, err = poller2.PollAndProcessDecisionTaskWithAttempt(false, false, false, false, 0)
	s.Nil(err)
	s.True(workflowFinished)
}

func (s *integrationClustersTestSuite) TestCronWorkflowFailover() {
	domainName := "test-cron-workflow-failover-" + common.GenerateRandomString(5)
	client1 := s.cluster1.GetFrontendClient() // active
	regReq := &types.RegisterDomainRequest{
		Name:                                   domainName,
		IsGlobalDomain:                         true,
		Clusters:                               clusterReplicationConfig,
		ActiveClusterName:                      clusterName[0],
		WorkflowExecutionRetentionPeriodInDays: 1,
	}
	err := client1.RegisterDomain(createContext(), regReq)
	s.NoError(err)

	descReq := &types.DescribeDomainRequest{
		Name: common.StringPtr(domainName),
	}
	resp, err := client1.DescribeDomain(createContext(), descReq)
	s.NoError(err)
	s.NotNil(resp)
	// Wait for domain cache to pick the change
	time.Sleep(cacheRefreshInterval)

	client2 := s.cluster2.GetFrontendClient() // standby

	// start a workflow
	id := "integration-cron-workflow-failover-test"
	wt := "integration-cron-workflow-failover-test-type"
	tl := "integration-cron-workflow-failover-test-tasklist"
	identity := "worker1"
	workflowType := &types.WorkflowType{Name: wt}
	taskList := &types.TaskList{Name: tl}
	startReq := &types.StartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              domainName,
		WorkflowID:                          id,
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(100),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		Identity:                            identity,
		CronSchedule:                        "@every 5s",
	}
	we, err := client1.StartWorkflowExecution(createContext(), startReq)
	s.NoError(err)
	s.NotNil(we.GetRunID())

	dtHandler := func(execution *types.WorkflowExecution, wt *types.WorkflowType,
		previousStartedEventID, startedEventID int64, history *types.History) ([]byte, []*types.Decision, error) {
		return nil, []*types.Decision{
			{
				DecisionType: types.DecisionTypeCompleteWorkflowExecution.Ptr(),
				CompleteWorkflowExecutionDecisionAttributes: &types.CompleteWorkflowExecutionDecisionAttributes{
					Result: []byte("cron-test-result"),
				},
			}}, nil
	}

	poller2 := host.TaskPoller{
		Engine:          client2,
		Domain:          domainName,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		Logger:          s.logger,
		T:               s.T(),
	}

	// Failover during backoff
	// Update domain to fail over
	updateReq := &types.UpdateDomainRequest{
		Name:              domainName,
		ActiveClusterName: common.StringPtr(clusterName[1]),
	}
	updateResp, err := client1.UpdateDomain(createContext(), updateReq)
	s.NoError(err)
	s.NotNil(updateResp)
	s.Equal(clusterName[1], updateResp.ReplicationConfiguration.GetActiveClusterName())
	s.Equal(int64(1), updateResp.GetFailoverVersion())

	// Wait for domain cache to pick the change
	time.Sleep(cacheRefreshInterval)

	// Run twice to make sure cron schedule is passed to standby.
	for i := 0; i < 2; i++ {
		_, err = poller2.PollAndProcessDecisionTask(false, false)
		s.Nil(err)
	}

	err = client2.TerminateWorkflowExecution(createContext(), &types.TerminateWorkflowExecutionRequest{
		Domain: domainName,
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: id,
		},
	})
	s.Nil(err)
}

func (s *integrationClustersTestSuite) TestWorkflowRetryFailover() {
	domainName := "test-workflow-retry-failover-" + common.GenerateRandomString(5)
	client1 := s.cluster1.GetFrontendClient() // active
	regReq := &types.RegisterDomainRequest{
		Name:                                   domainName,
		IsGlobalDomain:                         true,
		Clusters:                               clusterReplicationConfig,
		ActiveClusterName:                      clusterName[0],
		WorkflowExecutionRetentionPeriodInDays: 1,
	}
	err := client1.RegisterDomain(createContext(), regReq)
	s.NoError(err)

	descReq := &types.DescribeDomainRequest{
		Name: common.StringPtr(domainName),
	}
	resp, err := client1.DescribeDomain(createContext(), descReq)
	s.NoError(err)
	s.NotNil(resp)
	// Wait for domain cache to pick the change
	time.Sleep(cacheRefreshInterval)

	client2 := s.cluster2.GetFrontendClient() // standby

	// start a workflow
	id := "integration-workflow-retry-failover-test"
	wt := "integration-workflow-retry-failover-test-type"
	tl := "integration-workflow-retry-failover-test-tasklist"
	identity := "worker1"
	workflowType := &types.WorkflowType{Name: wt}
	taskList := &types.TaskList{Name: tl}
	startReq := &types.StartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              domainName,
		WorkflowID:                          id,
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(100),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		Identity:                            identity,
		RetryPolicy: &types.RetryPolicy{
			InitialIntervalInSeconds:    1,
			MaximumAttempts:             3,
			MaximumIntervalInSeconds:    1,
			NonRetriableErrorReasons:    []string{"bad-bug"},
			BackoffCoefficient:          1,
			ExpirationIntervalInSeconds: 100,
		},
	}
	we, err := client1.StartWorkflowExecution(createContext(), startReq)
	s.NoError(err)
	s.NotNil(we.GetRunID())

	executions := []*types.WorkflowExecution{}
	dtHandler := func(execution *types.WorkflowExecution, wt *types.WorkflowType,
		previousStartedEventID, startedEventID int64, history *types.History) ([]byte, []*types.Decision, error) {
		executions = append(executions, execution)
		return nil, []*types.Decision{
			{
				DecisionType: types.DecisionTypeFailWorkflowExecution.Ptr(),
				FailWorkflowExecutionDecisionAttributes: &types.FailWorkflowExecutionDecisionAttributes{
					Reason:  common.StringPtr("retryable-error"),
					Details: nil,
				},
			}}, nil
	}

	poller2 := host.TaskPoller{
		Engine:          client2,
		Domain:          domainName,
		TaskList:        taskList,
		Identity:        identity,
		DecisionHandler: dtHandler,
		Logger:          s.logger,
		T:               s.T(),
	}

	// Update domain to fail over
	updateReq := &types.UpdateDomainRequest{
		Name:              domainName,
		ActiveClusterName: common.StringPtr(clusterName[1]),
	}
	updateResp, err := client1.UpdateDomain(createContext(), updateReq)
	s.NoError(err)
	s.NotNil(updateResp)
	s.Equal(clusterName[1], updateResp.ReplicationConfiguration.GetActiveClusterName())
	s.Equal(int64(1), updateResp.GetFailoverVersion())

	// Wait for domain cache to pick the change
	time.Sleep(cacheRefreshInterval)

	// First attempt
	_, err = poller2.PollAndProcessDecisionTask(false, false)
	s.Nil(err)
	events := s.getHistory(client2, domainName, executions[0])
	s.Equal(types.EventTypeWorkflowExecutionContinuedAsNew, events[len(events)-1].GetEventType())
	s.Equal(int32(0), events[0].GetWorkflowExecutionStartedEventAttributes().GetAttempt())

	// second attempt
	_, err = poller2.PollAndProcessDecisionTask(false, false)
	s.Nil(err)
	events = s.getHistory(client2, domainName, executions[1])
	s.Equal(types.EventTypeWorkflowExecutionContinuedAsNew, events[len(events)-1].GetEventType())
	s.Equal(int32(1), events[0].GetWorkflowExecutionStartedEventAttributes().GetAttempt())

	// third attempt. Still failing, should stop retry.
	_, err = poller2.PollAndProcessDecisionTask(false, false)
	s.Nil(err)
	events = s.getHistory(client2, domainName, executions[2])
	s.Equal(types.EventTypeWorkflowExecutionFailed, events[len(events)-1].GetEventType())
	s.Equal(int32(2), events[0].GetWorkflowExecutionStartedEventAttributes().GetAttempt())
}

func (s *integrationClustersTestSuite) getHistory(client host.FrontendClient, domain string, execution *types.WorkflowExecution) []*types.HistoryEvent {
	historyResponse, err := client.GetWorkflowExecutionHistory(createContext(), &types.GetWorkflowExecutionHistoryRequest{
		Domain:          domain,
		Execution:       execution,
		MaximumPageSize: 5, // Use small page size to force pagination code path
	})
	s.Nil(err)

	events := historyResponse.History.Events
	for historyResponse.NextPageToken != nil {
		historyResponse, err = client.GetWorkflowExecutionHistory(createContext(), &types.GetWorkflowExecutionHistoryRequest{
			Domain:        domain,
			Execution:     execution,
			NextPageToken: historyResponse.NextPageToken,
		})
		s.Nil(err)
		events = append(events, historyResponse.History.Events...)
	}

	return events
}
