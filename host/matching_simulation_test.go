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

//go:build !race && matchingsim
// +build !race,matchingsim

/*
To run locally:

1. Change the matchingconfig in host/testdata/matching_simulation.yaml as you wish

2. Run `./scripts/run_matching_simulator.sh`

Full test logs can be found at test.log file. Event json logs can be found at matching-simulator-output.json.
See the run_matching_simulator.sh script for more details about how to parse events.
*/
package host

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/yarpc"
	"golang.org/x/time/rate"

	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/persistence"
	pt "github.com/uber/cadence/common/persistence/persistence-tests"
	"github.com/uber/cadence/common/types"

	_ "github.com/uber/cadence/common/asyncworkflow/queue/kafka" // needed to load kafka asyncworkflow queue
)

type operation string

const (
	operationPollForDecisionTask operation = "PollForDecisionTask"
	defaultTestCase                        = "testdata/matching_simulation_default.yaml"
)

type operationStats struct {
	op  operation
	dur time.Duration
	err error
}

type operationAggStats struct {
	successCnt    int
	failCnt       int
	totalDuration time.Duration
	maxDuration   time.Duration
	lastUpdated   time.Time
}

func TestMatchingSimulationSuite(t *testing.T) {
	flag.Parse()

	confPath := os.Getenv("MATCHING_SIMULATION_CONFIG")
	if confPath == "" {
		confPath = defaultTestCase
	}
	clusterConfig, err := GetTestClusterConfig(confPath)
	if err != nil {
		t.Fatalf("failed creating cluster config from %s, err: %v", confPath, err)
	}

	clusterConfig.MatchingDynamicConfigOverrides = map[dynamicconfig.Key]interface{}{
		dynamicconfig.MatchingNumTasklistWritePartitions:   getPartitions(clusterConfig.MatchingConfig.SimulationConfig.TaskListWritePartitions),
		dynamicconfig.MatchingNumTasklistReadPartitions:    getPartitions(clusterConfig.MatchingConfig.SimulationConfig.TaskListReadPartitions),
		dynamicconfig.MatchingForwarderMaxOutstandingPolls: getForwarderMaxOutstandingPolls(clusterConfig.MatchingConfig.SimulationConfig.ForwarderMaxOutstandingPolls),
		dynamicconfig.MatchingForwarderMaxOutstandingTasks: getForwarderMaxOutstandingTasks(clusterConfig.MatchingConfig.SimulationConfig.ForwarderMaxOutstandingTasks),
		dynamicconfig.MatchingForwarderMaxRatePerSecond:    getForwarderMaxRPS(clusterConfig.MatchingConfig.SimulationConfig.ForwarderMaxRatePerSecond),
		dynamicconfig.MatchingForwarderMaxChildrenPerNode:  getForwarderMaxChildPerNode(clusterConfig.MatchingConfig.SimulationConfig.ForwarderMaxChildrenPerNode),
		dynamicconfig.LocalPollWaitTime:                    clusterConfig.MatchingConfig.SimulationConfig.LocalPollWaitTime,
		dynamicconfig.LocalTaskWaitTime:                    clusterConfig.MatchingConfig.SimulationConfig.LocalTaskWaitTime,
	}

	ctrl := gomock.NewController(t)
	mockHistoryCl := history.NewMockClient(ctrl)
	mockHistoryCl.EXPECT().RecordDecisionTaskStarted(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *types.RecordDecisionTaskStartedRequest, opts ...yarpc.CallOption) (*types.RecordDecisionTaskStartedResponse, error) {
			time.Sleep(getRecordDecisionTaskStartedTime(clusterConfig.MatchingConfig.SimulationConfig.RecordDecisionTaskStartedTime))
			return &types.RecordDecisionTaskStartedResponse{
				ScheduledEventID: req.ScheduleID,
			}, nil
		}).AnyTimes()
	clusterConfig.HistoryConfig.MockClient = mockHistoryCl

	testCluster := NewPersistenceTestCluster(t, clusterConfig)

	s := new(MatchingSimulationSuite)
	params := IntegrationBaseParams{
		DefaultTestCluster:    testCluster,
		VisibilityTestCluster: testCluster,
		TestClusterConfig:     clusterConfig,
	}
	s.IntegrationBase = NewIntegrationBase(params)
	suite.Run(t, s)
}

func (s *MatchingSimulationSuite) SetupSuite() {
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

	time.Sleep(2 * time.Second)
}

func (s *MatchingSimulationSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *MatchingSimulationSuite) TearDownSuite() {
	s.tearDownSuite()
}

func (s *MatchingSimulationSuite) TestMatchingSimulation() {
	matchingClient := s.testCluster.GetMatchingClient()

	ctx, cancel := context.WithCancel(context.Background())

	domainID := s.domainID(ctx)
	tasklist := "my-tasklist"

	// Start stat collector
	statsCh := make(chan *operationStats, 10000)
	aggStats := make(map[operation]*operationAggStats)
	var collectorWG sync.WaitGroup
	collectorWG.Add(1)
	go s.collectStats(statsCh, aggStats, &collectorWG)

	// Start pollers
	numPollers := getNumPollers(s.testClusterConfig.MatchingConfig.SimulationConfig.NumPollers)
	pollDuration := getPollDuration(s.testClusterConfig.MatchingConfig.SimulationConfig.PollTimeout)
	polledTasksCounter := int32(0)
	maxTasksToGenerate := getMaxTaskstoGenerate(s.testClusterConfig.MatchingConfig.SimulationConfig.MaxTaskToGenerate)
	taskProcessTime := getTaskProcessTime(s.testClusterConfig.MatchingConfig.SimulationConfig.TaskProcessTime)
	var tasksToReceive sync.WaitGroup
	tasksToReceive.Add(maxTasksToGenerate)
	var pollerWG sync.WaitGroup
	for i := 0; i < numPollers; i++ {
		pollerWG.Add(1)
		go s.poll(ctx, matchingClient, domainID, tasklist, &polledTasksCounter, &pollerWG, pollDuration, taskProcessTime, statsCh, &tasksToReceive)
	}

	// wait a bit for pollers to start.
	time.Sleep(300 * time.Millisecond)

	startTime := time.Now()
	// Start task generators
	rps := getTaskQPS(s.testClusterConfig.MatchingConfig.SimulationConfig.TasksPerSecond)
	burst := getTaskBurst(s.testClusterConfig.MatchingConfig.SimulationConfig.TasksBurst)
	rateLimiter := rate.NewLimiter(rate.Limit(rps), burst)
	generatedTasksCounter := int32(0)
	lastTaskScheduleID := int32(0)
	numGenerators := getNumGenerators(s.testClusterConfig.MatchingConfig.SimulationConfig.NumTaskGenerators)
	var tasksToGenerate sync.WaitGroup
	tasksToGenerate.Add(maxTasksToGenerate)
	var generatorWG sync.WaitGroup
	for i := 1; i <= numGenerators; i++ {
		generatorWG.Add(1)
		go s.generate(ctx, matchingClient, domainID, tasklist, maxTasksToGenerate, rateLimiter, &generatedTasksCounter, &lastTaskScheduleID, &generatorWG, &tasksToGenerate)
	}

	// Let it run until all tasks have been polled.
	// There's a test timeout configured in docker/buildkite/docker-compose-local-matching-simulation.yml that you
	// can change if your test case needs more time
	s.log("Waiting until all tasks are generated")
	tasksToGenerate.Wait()
	generationTime := time.Now().Sub(startTime)
	s.log("Waiting until all tasks are received")
	tasksToReceive.Wait()
	executionTime := time.Now().Sub(startTime)
	s.log("Completed benchmark in %v", (time.Now().Sub(startTime)))
	s.log("Canceling context to stop pollers and task generators")
	cancel()
	pollerWG.Wait()
	s.log("Pollers stopped")
	generatorWG.Wait()
	s.log("Generators stopped")
	s.log("Stopping stats collector")
	close(statsCh)
	collectorWG.Wait()
	s.log("Stats collector stopped")

	// Print the test summary.
	// Don't change the start/end line format as it is used by scripts to parse the summary info
	testSummary := []string{}
	testSummary = append(testSummary, "Simulation Summary:")
	testSummary = append(testSummary, fmt.Sprintf("Task generate Duration: %v", generationTime))
	testSummary = append(testSummary, fmt.Sprintf("Simulation Duration: %v", executionTime))
	testSummary = append(testSummary, fmt.Sprintf("Num of Pollers: %d", numPollers))
	testSummary = append(testSummary, fmt.Sprintf("Poll Timeout: %v", pollDuration))
	testSummary = append(testSummary, fmt.Sprintf("Num of Task Generators: %d", numGenerators))
	testSummary = append(testSummary, fmt.Sprintf("Task generated QPS: %v", rps))
	testSummary = append(testSummary, fmt.Sprintf("Num of Write Partitions: %d", s.testClusterConfig.MatchingDynamicConfigOverrides[dynamicconfig.MatchingNumTasklistWritePartitions]))
	testSummary = append(testSummary, fmt.Sprintf("Num of Read Partitions: %d", s.testClusterConfig.MatchingDynamicConfigOverrides[dynamicconfig.MatchingNumTasklistReadPartitions]))
	testSummary = append(testSummary, fmt.Sprintf("Forwarder Max Outstanding Polls: %d", s.testClusterConfig.MatchingDynamicConfigOverrides[dynamicconfig.MatchingForwarderMaxOutstandingPolls]))
	testSummary = append(testSummary, fmt.Sprintf("Forwarder Max Outstanding Tasks: %d", s.testClusterConfig.MatchingDynamicConfigOverrides[dynamicconfig.MatchingForwarderMaxOutstandingTasks]))
	testSummary = append(testSummary, fmt.Sprintf("Forwarder Max RPS: %d", s.testClusterConfig.MatchingDynamicConfigOverrides[dynamicconfig.MatchingForwarderMaxRatePerSecond]))
	testSummary = append(testSummary, fmt.Sprintf("Forwarder Max Children per Node: %d", s.testClusterConfig.MatchingDynamicConfigOverrides[dynamicconfig.MatchingForwarderMaxChildrenPerNode]))
	testSummary = append(testSummary, fmt.Sprintf("Local Poll Wait Time: %v", s.testClusterConfig.MatchingDynamicConfigOverrides[dynamicconfig.LocalPollWaitTime]))
	testSummary = append(testSummary, fmt.Sprintf("Local Task Wait Time: %v", s.testClusterConfig.MatchingDynamicConfigOverrides[dynamicconfig.LocalTaskWaitTime]))
	testSummary = append(testSummary, fmt.Sprintf("Tasks generated: %d", generatedTasksCounter))
	testSummary = append(testSummary, fmt.Sprintf("Tasks polled: %d", polledTasksCounter))
	totalPollCnt := 0
	if pollStats, ok := aggStats[operationPollForDecisionTask]; ok {
		totalPollCnt = pollStats.successCnt + pollStats.failCnt
	}

	if totalPollCnt == 0 {
		testSummary = append(testSummary, "No poll requests were made")
	} else {
		testSummary = append(testSummary, fmt.Sprintf("Total poll requests: %d", totalPollCnt))
		testSummary = append(testSummary, fmt.Sprintf("Poll request failure rate %%: %d", 100*aggStats[operationPollForDecisionTask].failCnt/totalPollCnt))
		testSummary = append(testSummary, fmt.Sprintf("Avg Poll latency (ms): %d", (aggStats[operationPollForDecisionTask].totalDuration/time.Duration(totalPollCnt)).Milliseconds()))
		testSummary = append(testSummary, fmt.Sprintf("Max Poll latency (ms): %d", aggStats[operationPollForDecisionTask].maxDuration.Milliseconds()))
	}
	testSummary = append(testSummary, "End of Simulation Summary")
	fmt.Println(strings.Join(testSummary, "\n"))
}

func (s *MatchingSimulationSuite) log(msg string, args ...interface{}) {
	msg = time.Now().Format(time.RFC3339Nano) + "\t" + msg
	s.T().Logf(msg, args...)
}

func (s *MatchingSimulationSuite) generate(
	ctx context.Context,
	matchingClient MatchingClient,
	domainID, tasklist string,
	maxTasksToGenerate int,
	rateLimiter *rate.Limiter,
	generatedTasksCounter *int32,
	lastTaskScheduleID *int32,
	wg *sync.WaitGroup,
	tasksToGenerate *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			s.log("Generator done")
			return
		default:
			if err := rateLimiter.Wait(ctx); err != nil {
				if !errors.Is(err, context.Canceled) {
					s.T().Fatal("Rate limiter failed: ", err)
				}
				return
			}
			scheduleID := int(atomic.AddInt32(lastTaskScheduleID, 1))
			if scheduleID > maxTasksToGenerate {
				s.log("Generated %d tasks so generator will stop", maxTasksToGenerate)
				return
			}
			decisionTask := newDecisionTask(domainID, tasklist, scheduleID)
			reqCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			err := matchingClient.AddDecisionTask(reqCtx, decisionTask)
			cancel()
			if err != nil {
				s.log("Error when adding decision task, err: %v", err)
				continue
			}

			s.log("Decision task %d added", scheduleID)
			atomic.AddInt32(generatedTasksCounter, 1)
			tasksToGenerate.Done()
		}
	}
}

func (s *MatchingSimulationSuite) poll(
	ctx context.Context,
	matchingClient MatchingClient,
	domainID, tasklist string,
	polledTasksCounter *int32,
	wg *sync.WaitGroup,
	pollDuration, taskProcessTime time.Duration,
	statsCh chan *operationStats,
	tasksToReceive *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			s.log("Poller done")
			return
		default:
			s.log("Poller will initiate a poll")
			reqCtx, cancel := context.WithTimeout(ctx, pollDuration)
			start := time.Now()
			resp, err := matchingClient.PollForDecisionTask(reqCtx, &types.MatchingPollForDecisionTaskRequest{
				DomainUUID: domainID,
				PollRequest: &types.PollForDecisionTaskRequest{
					TaskList: &types.TaskList{
						Name: tasklist,
						Kind: types.TaskListKindNormal.Ptr(),
					},
				},
			})
			cancel()

			statsCh <- &operationStats{
				op:  operationPollForDecisionTask,
				dur: time.Since(start),
				err: err,
			}

			if err != nil {
				s.log("PollForDecisionTask failed: %v", err)
				continue
			}

			empty := &types.MatchingPollForDecisionTaskResponse{}

			if reflect.DeepEqual(empty, resp) {
				s.log("PollForDecisionTask response is empty")
				continue
			}

			atomic.AddInt32(polledTasksCounter, 1)
			s.log("PollForDecisionTask got a task with startedid: %d. resp: %+v", resp.StartedEventID, resp)
			tasksToReceive.Done()
			time.Sleep(taskProcessTime)
		}
	}
}

func (s *MatchingSimulationSuite) collectStats(statsCh chan *operationStats, aggStats map[operation]*operationAggStats, wg *sync.WaitGroup) {
	defer wg.Done()
	for stat := range statsCh {
		opAggStats, ok := aggStats[stat.op]
		if !ok {
			opAggStats = &operationAggStats{}
			aggStats[stat.op] = opAggStats
		}

		opAggStats.lastUpdated = time.Now()
		if stat.err != nil {
			opAggStats.failCnt++
		} else {
			opAggStats.successCnt++
		}

		opAggStats.totalDuration += stat.dur
		if stat.dur > opAggStats.maxDuration {
			opAggStats.maxDuration = stat.dur
		}
	}

	s.log("Stats collector done")
}

func (s *MatchingSimulationSuite) domainID(ctx context.Context) string {
	reqCtx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	domainDesc, err := s.testCluster.GetFrontendClient().DescribeDomain(reqCtx, &types.DescribeDomainRequest{
		Name: &s.domainName,
	})
	s.Require().NoError(err, "Error when describing domain")

	domainID := domainDesc.GetDomainInfo().UUID
	s.T().Logf("DomainID: %s", domainID)
	return domainID
}

func newDecisionTask(domainID, tasklist string, i int) *types.AddDecisionTaskRequest {
	return &types.AddDecisionTaskRequest{
		DomainUUID: domainID,
		Execution: &types.WorkflowExecution{
			WorkflowID: "test-workflow-id",
			RunID:      uuid.New(),
		},
		TaskList: &types.TaskList{
			Name: tasklist,
			Kind: types.TaskListKindNormal.Ptr(),
		},
		ScheduleID: int64(i),
	}
}

func getMaxTaskstoGenerate(i int) int {
	if i == 0 {
		return 2000
	}
	return i
}

func getTaskGenerateInterval(i time.Duration) time.Duration {
	if i == 0 {
		return 50 * time.Millisecond
	}
	return i
}

func getNumGenerators(i int) int {
	if i == 0 {
		return 1
	}
	return i
}

func getNumPollers(i int) int {
	if i == 0 {
		return 10
	}
	return i
}

func getPollDuration(d time.Duration) time.Duration {
	if d == 0 {
		return 1 * time.Second
	}
	return d
}

func getPartitions(i int) int {
	if i == 0 {
		return 1
	}
	return i
}

func getForwarderMaxOutstandingPolls(i int) int {
	return i
}

func getForwarderMaxOutstandingTasks(i int) int {
	return i
}

func getForwarderMaxRPS(i int) int {
	if i == 0 {
		return 10
	}
	return i
}

func getForwarderMaxChildPerNode(i int) int {
	if i == 0 {
		return 20
	}
	return i
}

func getTaskQPS(i int) int {
	if i == 0 {
		return 40
	}
	return i
}

func getTaskBurst(i int) int {
	if i == 0 {
		return 1
	}
	return i
}

func getTaskProcessTime(duration time.Duration) time.Duration {
	if duration == 0 {
		return time.Millisecond
	}
	return duration
}

func getRecordDecisionTaskStartedTime(duration time.Duration) time.Duration {
	if duration == 0 {
		return time.Millisecond
	}

	return duration
}
