// Copyright (c) 2017-2020 Uber Technologies Inc.

// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package matching

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"go.uber.org/yarpc"

	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/isolationgroup/defaultisolationgroupstate"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/partition"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/quotas"
	"github.com/uber/cadence/common/types"
)

type (
	matchingEngineSuite struct {
		suite.Suite
		controller         *gomock.Controller
		mockHistoryClient  *history.MockClient
		mockDomainCache    *cache.MockDomainCache
		mockIsolationStore *dynamicconfig.MockClient

		matchingEngine       *matchingEngineImpl
		taskManager          *testTaskManager
		partitioner          partition.Partitioner
		mockExecutionManager *mocks.ExecutionManager
		logger               log.Logger
		handlerContext       *handlerContext
		sync.Mutex
	}
)

const (
	_minBurst              = 10000
	matchingTestDomainName = "matching-test"
	matchingTestTaskList   = "matching-test-tasklist"
)

func TestMatchingEngineSuite(t *testing.T) {
	s := new(matchingEngineSuite)
	suite.Run(t, s)
}

func (s *matchingEngineSuite) SetupSuite() {
	s.logger = testlogger.New(s.Suite.T())
	http.Handle("/test/tasks", http.HandlerFunc(s.TasksHandler))
}

// Renders content of taskManager and matchingEngine when called at http://localhost:6060/test/tasks
// Uncomment HTTP server initialization in SetupSuite method to enable.

func (s *matchingEngineSuite) TasksHandler(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "%v\n", s.taskManager)
	fmt.Fprintf(w, "%v\n", s.matchingEngine)
}

func (s *matchingEngineSuite) TearDownSuite() {
}

func (s *matchingEngineSuite) SetupTest() {
	s.Lock()
	defer s.Unlock()
	tlKindNormal := types.TaskListKindNormal
	s.mockExecutionManager = &mocks.ExecutionManager{}
	s.controller = gomock.NewController(s.T())
	s.mockHistoryClient = history.NewMockClient(s.controller)
	s.taskManager = newTestTaskManager(s.logger)
	s.mockDomainCache = cache.NewMockDomainCache(s.controller)
	s.mockDomainCache.EXPECT().GetDomainByID(gomock.Any()).Return(cache.CreateDomainCacheEntry(matchingTestDomainName), nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomain(gomock.Any()).Return(cache.CreateDomainCacheEntry(matchingTestDomainName), nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomainName(gomock.Any()).Return(matchingTestDomainName, nil).AnyTimes()
	s.mockIsolationStore = dynamicconfig.NewMockClient(s.controller)
	dcClient := dynamicconfig.NewInMemoryClient()
	dcClient.UpdateValue(dynamicconfig.EnableTasklistIsolation, true)
	dcClient.UpdateValue(dynamicconfig.AllIsolationGroups, []interface{}{"datacenterA", "datacenterB"})
	dc := dynamicconfig.NewCollection(dcClient, s.logger)
	isolationGroupState, _ := defaultisolationgroupstate.NewDefaultIsolationGroupStateWatcherWithConfigStoreClient(s.logger, dc, s.mockDomainCache, s.mockIsolationStore, metrics.NewNoopMetricsClient())
	s.partitioner = partition.NewDefaultPartitioner(s.logger, isolationGroupState)
	s.handlerContext = newHandlerContext(
		context.Background(),
		matchingTestDomainName,
		&types.TaskList{Name: matchingTestTaskList, Kind: &tlKindNormal},
		metrics.NewClient(tally.NoopScope, metrics.Matching),
		metrics.MatchingTaskListMgrScope,
		testlogger.New(s.Suite.T()),
	)

	s.matchingEngine = s.newMatchingEngine(defaultTestConfig(), s.taskManager)
	s.matchingEngine.Start()
}

func (s *matchingEngineSuite) TearDownTest() {
	s.mockExecutionManager.AssertExpectations(s.T())
	s.matchingEngine.Stop()
	s.controller.Finish()
}

func (s *matchingEngineSuite) newMatchingEngine(
	config *Config, taskMgr persistence.TaskManager,
) *matchingEngineImpl {
	return newMatchingEngine(config, taskMgr, s.mockHistoryClient, s.logger, s.mockDomainCache, s.partitioner)
}

func newMatchingEngine(
	config *Config, taskMgr persistence.TaskManager, mockHistoryClient history.Client,
	logger log.Logger, mockDomainCache cache.DomainCache, partitioner partition.Partitioner,
) *matchingEngineImpl {
	return &matchingEngineImpl{
		taskManager:     taskMgr,
		clusterMetadata: cluster.GetTestClusterMetadata(true),
		historyService:  mockHistoryClient,
		taskLists:       make(map[taskListID]taskListManager),
		logger:          logger,
		metricsClient:   metrics.NewClient(tally.NoopScope, metrics.Matching),
		tokenSerializer: common.NewJSONTaskTokenSerializer(),
		config:          config,
		domainCache:     mockDomainCache,
		partitioner:     partitioner,
	}
}

func (s *matchingEngineSuite) TestPollForActivityTasksEmptyResult() {
	s.PollForTasksEmptyResultTest(context.Background(), persistence.TaskListTypeActivity)
}

func (s *matchingEngineSuite) TestPollForDecisionTasksEmptyResult() {
	s.PollForTasksEmptyResultTest(context.Background(), persistence.TaskListTypeDecision)
}

func (s *matchingEngineSuite) TestPollForActivityTasksEmptyResultWithShortContext() {
	shortContextTimeout := returnEmptyTaskTimeBudget + 10*time.Millisecond
	callContext, cancel := context.WithTimeout(context.Background(), shortContextTimeout)
	defer cancel()
	s.PollForTasksEmptyResultTest(callContext, persistence.TaskListTypeActivity)
}

func (s *matchingEngineSuite) TestPollForDecisionTasksEmptyResultWithShortContext() {
	shortContextTimeout := returnEmptyTaskTimeBudget + 10*time.Millisecond
	callContext, cancel := context.WithTimeout(context.Background(), shortContextTimeout)
	defer cancel()
	s.PollForTasksEmptyResultTest(callContext, persistence.TaskListTypeDecision)
}

func (s *matchingEngineSuite) TestOnlyUnloadMatchingInstance() {
	taskListID := newTestTaskListID(
		uuid.New(),
		"makeToast",
		persistence.TaskListTypeActivity)
	tlKind := types.TaskListKindNormal
	tlm, err := s.matchingEngine.getTaskListManager(taskListID, &tlKind)
	s.Require().NoError(err)

	tlm2, err := newTaskListManager(
		s.matchingEngine,
		taskListID, // same taskListID as above
		&tlKind,
		s.matchingEngine.config, time.Now())
	s.Require().NoError(err)

	// try to unload a different tlm instance with the same taskListID
	s.matchingEngine.unloadTaskList(tlm2)

	got, err := s.matchingEngine.getTaskListManager(taskListID, &tlKind)
	s.Require().NoError(err)
	s.Require().Same(tlm, got,
		"Unload call with non-matching taskListManager should not cause unload")

	// this time unload the right tlm
	s.matchingEngine.unloadTaskList(tlm)

	got, err = s.matchingEngine.getTaskListManager(taskListID, &tlKind)
	s.Require().NoError(err)
	s.Require().NotSame(tlm, got,
		"Unload call with matching incarnation should have caused unload")
}

func (s *matchingEngineSuite) TestPollForDecisionTasks() {
	s.PollForDecisionTasksResultTest()
}

func (s *matchingEngineSuite) PollForDecisionTasksResultTest() {
	taskType := persistence.TaskListTypeDecision
	domainID := "domainId"
	tl := "makeToast"
	tlKind := types.TaskListKindNormal
	stickyTl := "makeStickyToast"
	stickyTlKind := types.TaskListKindSticky
	identity := "selfDrivingToaster"

	stickyTaskList := &types.TaskList{}
	stickyTaskList.Name = stickyTl
	stickyTaskList.Kind = &stickyTlKind

	s.matchingEngine.config.RangeSize = 2 // to test that range is not updated without tasks
	s.matchingEngine.config.LongPollExpirationInterval = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(10 * time.Millisecond)

	runID := "run1"
	workflowID := "workflow1"
	workflowType := types.WorkflowType{
		Name: "workflow",
	}
	execution := types.WorkflowExecution{RunID: runID, WorkflowID: workflowID}
	scheduleID := int64(0)

	// History service is using mock
	s.mockHistoryClient.EXPECT().RecordDecisionTaskStarted(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, taskRequest *types.RecordDecisionTaskStartedRequest, option ...yarpc.CallOption) (*types.RecordDecisionTaskStartedResponse, error) {
			s.logger.Debug("Mock Received RecordDecisionTaskStartedRequest")
			taskListKindNormal := types.TaskListKindNormal
			response := &types.RecordDecisionTaskStartedResponse{}
			response.WorkflowType = &workflowType
			response.PreviousStartedEventID = common.Int64Ptr(scheduleID)
			response.ScheduledEventID = scheduleID + 1
			response.Attempt = 0
			response.StickyExecutionEnabled = true
			response.WorkflowExecutionTaskList = &types.TaskList{
				Name: tl,
				Kind: &taskListKindNormal,
			}
			return response, nil
		}).AnyTimes()

	addRequest := &addTaskRequest{
		TaskType:                      taskType,
		DomainUUID:                    domainID,
		Execution:                     &execution,
		TaskList:                      stickyTaskList,
		ScheduleToStartTimeoutSeconds: 1,
	}
	_, err := addTask(s.matchingEngine, s.handlerContext, addRequest)
	s.Error(err)
	s.Contains(err.Error(), "sticky worker is unavailable")
	// poll the sticky tasklist, should get no result
	pollReq := &pollTaskRequest{
		TaskType:   taskType,
		DomainUUID: domainID,
		TaskList:   stickyTaskList,
		Identity:   identity,
	}
	resp, err := pollTask(s.matchingEngine, s.handlerContext, pollReq)
	s.NoError(err)
	s.Equal(&pollTaskResponse{}, resp)
	// add task to sticky tasklist again, this time it should pass
	_, err = addTask(s.matchingEngine, s.handlerContext, addRequest)
	s.NoError(err)

	resp, err = pollTask(s.matchingEngine, s.handlerContext, pollReq)

	expectedResp := &pollTaskResponse{
		TaskToken:              resp.TaskToken,
		WorkflowExecution:      &execution,
		WorkflowType:           &workflowType,
		PreviousStartedEventID: common.Int64Ptr(scheduleID),
		Attempt:                0,
		BacklogCountHint:       1,
		StickyExecutionEnabled: true,
		WorkflowExecutionTaskList: &types.TaskList{
			Name: tl,
			Kind: &tlKind,
		},
	}

	s.Nil(err)
	s.Equal(expectedResp, resp)
}

func (s *matchingEngineSuite) PollForTasksEmptyResultTest(callContext context.Context, taskType int) {
	s.matchingEngine.config.RangeSize = 2 // to test that range is not updated without tasks
	if _, ok := callContext.Deadline(); !ok {
		s.matchingEngine.config.LongPollExpirationInterval = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(10 * time.Millisecond)
	}

	domainID := "domainId"
	tl := "makeToast"
	identity := "selfDrivingToaster"

	taskList := &types.TaskList{Name: tl}
	var taskListType types.TaskListType
	tlID := newTestTaskListID(domainID, tl, taskType)
	s.handlerContext.Context = callContext
	const pollCount = 10
	for i := 0; i < pollCount; i++ {
		pollReq := &pollTaskRequest{
			TaskType:   taskType,
			DomainUUID: domainID,
			TaskList:   taskList,
			Identity:   identity,
		}
		pollResp, err := pollTask(s.matchingEngine, s.handlerContext, pollReq)
		s.NoError(err)
		s.Equal(&pollTaskResponse{}, pollResp)

		if taskType == persistence.TaskListTypeActivity {
			taskListType = types.TaskListTypeActivity
		} else {
			taskListType = types.TaskListTypeDecision
		}
		select {
		case <-callContext.Done():
			s.FailNow("Call context has expired.")
		default:
		}
		// check the poller information
		s.handlerContext.Context = context.Background()
		descResp, err := s.matchingEngine.DescribeTaskList(s.handlerContext, &types.MatchingDescribeTaskListRequest{
			DomainUUID: domainID,
			DescRequest: &types.DescribeTaskListRequest{
				TaskList:              taskList,
				TaskListType:          &taskListType,
				IncludeTaskListStatus: false,
			},
		})
		s.NoError(err)
		s.Equal(1, len(descResp.Pollers))
		s.Equal(identity, descResp.Pollers[0].GetIdentity())
		s.NotEmpty(descResp.Pollers[0].GetLastAccessTime())
		s.Nil(descResp.GetTaskListStatus())
	}
	s.EqualValues(1, s.taskManager.taskLists[*tlID].rangeID)
}

func (s *matchingEngineSuite) TestAddActivityTasks() {
	s.AddTasksTest(persistence.TaskListTypeActivity, false)
}

func (s *matchingEngineSuite) TestAddDecisionTasks() {
	s.AddTasksTest(persistence.TaskListTypeDecision, false)
}

func (s *matchingEngineSuite) TestAddActivityTasksForwarded() {
	s.AddTasksTest(persistence.TaskListTypeActivity, true)
}

func (s *matchingEngineSuite) TestAddDecisionTasksForwarded() {
	s.AddTasksTest(persistence.TaskListTypeDecision, true)
}

func (s *matchingEngineSuite) AddTasksTest(taskType int, isForwarded bool) {
	s.matchingEngine.config.RangeSize = 300 // override to low number for the test

	domainID := "domainId"
	tl := "makeToast"
	forwardedFrom := "/__cadence_sys/makeToast/1"

	taskList := &types.TaskList{Name: tl}

	const taskCount = 111

	runID := "run1"
	workflowID := "workflow1"
	execution := types.WorkflowExecution{RunID: runID, WorkflowID: workflowID}

	for i := int64(0); i < taskCount; i++ {
		scheduleID := i * 3
		addRequest := &addTaskRequest{
			TaskType:                      taskType,
			DomainUUID:                    domainID,
			Execution:                     &execution,
			ScheduleID:                    scheduleID,
			TaskList:                      taskList,
			ScheduleToStartTimeoutSeconds: 1,
		}
		if isForwarded {
			addRequest.ForwardedFrom = forwardedFrom
		}
		_, err := addTask(s.matchingEngine, s.handlerContext, addRequest)

		switch isForwarded {
		case false:
			s.NoError(err)
		case true:
			s.Equal(errRemoteSyncMatchFailed, err)
		}
	}

	switch isForwarded {
	case false:
		s.EqualValues(taskCount, s.taskManager.getTaskCount(newTestTaskListID(domainID, tl, taskType)))
	case true:
		s.EqualValues(0, s.taskManager.getTaskCount(newTestTaskListID(domainID, tl, taskType)))
	}
}

func (s *matchingEngineSuite) TestTaskWriterShutdown() {
	s.matchingEngine.config.RangeSize = 300 // override to low number for the test

	taskType := persistence.TaskListTypeActivity
	domainID := "domainId"
	tl := "makeToast"

	taskList := &types.TaskList{Name: tl}

	runID := "run1"
	workflowID := "workflow1"
	execution := types.WorkflowExecution{RunID: runID, WorkflowID: workflowID}

	tlID := newTestTaskListID(domainID, tl, taskType)
	tlKind := types.TaskListKindNormal
	tlm, err := s.matchingEngine.getTaskListManager(tlID, &tlKind)
	s.Nil(err)

	// stop the task writer explicitly
	tlmImpl := tlm.(*taskListManagerImpl)
	tlmImpl.taskWriter.Stop()

	// now attempt to add a task
	addRequest := &addTaskRequest{
		TaskType:                      taskType,
		DomainUUID:                    domainID,
		Execution:                     &execution,
		TaskList:                      taskList,
		ScheduleToStartTimeoutSeconds: 1,
		ScheduleID:                    5,
	}
	_, err = addTask(s.matchingEngine, s.handlerContext, addRequest)
	s.Error(err)

	// test race
	tlmImpl.taskWriter.stopped = 0
	_, err = addTask(s.matchingEngine, s.handlerContext, addRequest)
	s.Error(err)
	tlmImpl.taskWriter.stopped = 1 // reset it back to old value
}

func (s *matchingEngineSuite) TestAddAndPollDecisionTasks() {
	s.AddAndPollTasks(persistence.TaskListTypeDecision, false)
}

func (s *matchingEngineSuite) TestAddAndPollActivityTasks() {
	s.AddAndPollTasks(persistence.TaskListTypeDecision, false)
}

func (s *matchingEngineSuite) TestAddAndPollDecisionTasksIsolation() {
	s.AddAndPollTasks(persistence.TaskListTypeDecision, true)
}

func (s *matchingEngineSuite) TestAddAndPollActivityTasksIsolation() {
	s.AddAndPollTasks(persistence.TaskListTypeDecision, true)
}

func (s *matchingEngineSuite) AddAndPollTasks(taskType int, enableIsolation bool) {
	s.matchingEngine.config.LongPollExpirationInterval = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(10 * time.Millisecond)
	s.matchingEngine.config.EnableTasklistIsolation = dynamicconfig.GetBoolPropertyFnFilteredByDomainID(enableIsolation)

	isolationGroups := s.matchingEngine.config.AllIsolationGroups

	const taskCount = 1000
	const initialRangeID = 102
	// TODO: Understand why publish is low when rangeSize is 3
	const rangeSize = 30

	testParam := newTestParam(taskType)
	s.taskManager.getTaskListManager(testParam.TaskListID).rangeID = initialRangeID
	s.matchingEngine.config.RangeSize = rangeSize // override to low number for the test

	s.setupGetDrainStatus()

	for i := int64(0); i < taskCount; i++ {
		scheduleID := i * 3
		addRequest := &addTaskRequest{
			TaskType:                      taskType,
			DomainUUID:                    testParam.DomainID,
			Execution:                     testParam.WorkflowExecution,
			ScheduleID:                    scheduleID,
			TaskList:                      testParam.TaskList,
			ScheduleToStartTimeoutSeconds: 1,
			PartitionConfig:               map[string]string{partition.IsolationGroupKey: isolationGroups[int(i)%len(isolationGroups)]},
		}
		_, err := addTask(s.matchingEngine, s.handlerContext, addRequest)
		s.NoError(err)
	}
	s.EqualValues(taskCount, s.taskManager.getTaskCount(testParam.TaskListID))

	s.setupRecordTaskStartedMock(taskType, testParam, false)

	for i := int64(0); i < taskCount; {
		scheduleID := i * 3
		pollReq := &pollTaskRequest{
			TaskType:       taskType,
			DomainUUID:     testParam.DomainID,
			TaskList:       testParam.TaskList,
			Identity:       testParam.Identity,
			IsolationGroup: isolationGroups[int(i)%len(isolationGroups)],
		}
		result, err := pollTask(s.matchingEngine, s.handlerContext, pollReq)
		s.NoError(err)
		s.NotNil(result)
		if isEmptyToken(result.TaskToken) {
			s.logger.Debug("empty poll returned")
			continue
		}
		s.assertPollTaskResponse(taskType, testParam, scheduleID, result)
		i++
	}
	s.EqualValues(0, s.taskManager.getTaskCount(testParam.TaskListID))
	expectedRange := getExpectedRange(initialRangeID, taskCount, rangeSize)
	// Due to conflicts some ids are skipped and more real ranges are used.
	s.True(expectedRange <= s.taskManager.getTaskListManager(testParam.TaskListID).rangeID)
}

func (s *matchingEngineSuite) TestSyncMatchActivityTasks() {
	s.SyncMatchTasks(persistence.TaskListTypeActivity, false)
}

func (s *matchingEngineSuite) TestSyncMatchDecisionTasks() {
	s.SyncMatchTasks(persistence.TaskListTypeDecision, false)
}

func (s *matchingEngineSuite) TestSyncMatchActivityTasksIsolation() {
	s.SyncMatchTasks(persistence.TaskListTypeActivity, true)
}

func (s *matchingEngineSuite) TestSyncMatchDecisionTasksIsolation() {
	s.SyncMatchTasks(persistence.TaskListTypeDecision, true)
}

func (s *matchingEngineSuite) SyncMatchTasks(taskType int, enableIsolation bool) {
	s.matchingEngine.config.EnableTasklistIsolation = dynamicconfig.GetBoolPropertyFnFilteredByDomainID(enableIsolation)
	const taskCount = 10
	const initialRangeID = 102
	// TODO: Understand why publish is low when rangeSize is 3
	const rangeSize = 30
	var throttledTaskCount int64
	if taskType == persistence.TaskListTypeActivity {
		throttledTaskCount = 3
	}
	isolationGroups := s.matchingEngine.config.AllIsolationGroups

	// Set a short long poll expiration so we don't have to wait too long for 0 throttling cases
	s.matchingEngine.config.LongPollExpirationInterval = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(50 * time.Millisecond)
	s.matchingEngine.config.RangeSize = rangeSize // override to low number for the test
	// So we can get snapshots
	scope := tally.NewTestScope("test", nil)
	s.matchingEngine.metricsClient = metrics.NewClient(scope, metrics.Matching)

	testParam := newTestParam(taskType)
	dispatchTTL := time.Nanosecond
	dPtr := _defaultTaskDispatchRPS
	tlKind := types.TaskListKindNormal
	mgr, err := newTaskListManager(s.matchingEngine, testParam.TaskListID, &tlKind, s.matchingEngine.config, time.Now())
	s.NoError(err)
	mgrImpl, ok := mgr.(*taskListManagerImpl)
	s.True(ok)
	mgrImpl.matcher.limiter = quotas.NewRateLimiter(&dPtr, dispatchTTL, _minBurst)
	s.matchingEngine.updateTaskList(testParam.TaskListID, mgr)
	s.taskManager.getTaskListManager(testParam.TaskListID).rangeID = initialRangeID
	s.NoError(mgr.Start())

	s.setupGetDrainStatus()
	s.setupRecordTaskStartedMock(taskType, testParam, false)

	pollFunc := func(maxDispatch float64, isolationGroup string) (*pollTaskResponse, error) {
		pollReq := &pollTaskRequest{
			TaskType:         taskType,
			DomainUUID:       testParam.DomainID,
			TaskList:         testParam.TaskList,
			Identity:         testParam.Identity,
			TaskListMetadata: &types.TaskListMetadata{MaxTasksPerSecond: &maxDispatch},
			IsolationGroup:   isolationGroup,
		}
		return pollTask(s.matchingEngine, s.handlerContext, pollReq)
	}

	for i := int64(0); i < taskCount; i++ {
		scheduleID := i * 3
		group := isolationGroups[int(i)%len(isolationGroups)]
		var wg sync.WaitGroup
		var result *pollTaskResponse
		var pollErr error
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, pollErr = pollFunc(_defaultTaskDispatchRPS, group)
		}()
		time.Sleep(200 * time.Millisecond) // Wait for a short period of time to let the poller start so that sync match will happen
		addRequest := &addTaskRequest{
			TaskType:                      taskType,
			DomainUUID:                    testParam.DomainID,
			Execution:                     testParam.WorkflowExecution,
			ScheduleID:                    scheduleID,
			TaskList:                      testParam.TaskList,
			ScheduleToStartTimeoutSeconds: 1,
			PartitionConfig:               map[string]string{partition.IsolationGroupKey: group},
		}
		_, err := addTask(s.matchingEngine, s.handlerContext, addRequest)
		wg.Wait()
		s.NoError(err)
		s.NoError(pollErr)
		s.NotNil(result)
		s.assertPollTaskResponse(taskType, testParam, scheduleID, result)
	}

	// Set the dispatch RPS to 0, to verify that poller will not get any task and task will be persisted into database
	// Revert the dispatch RPS and verify that poller will get the task
	for i := int64(0); i < throttledTaskCount; i++ {
		scheduleID := i * 3
		group := isolationGroups[int(i)%len(isolationGroups)]
		var wg sync.WaitGroup
		var result *pollTaskResponse
		var pollErr error
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, pollErr = pollFunc(0.0, group)
		}()
		time.Sleep(20 * time.Millisecond) // Wait for a short period of time to let the poller start so that sync match will happen
		addRequest := &addTaskRequest{
			TaskType:                      taskType,
			DomainUUID:                    testParam.DomainID,
			Execution:                     testParam.WorkflowExecution,
			ScheduleID:                    scheduleID,
			TaskList:                      testParam.TaskList,
			ScheduleToStartTimeoutSeconds: 1,
			PartitionConfig:               map[string]string{partition.IsolationGroupKey: group},
		}
		_, err := addTask(s.matchingEngine, s.handlerContext, addRequest)
		wg.Wait()
		s.NoError(err)
		s.NoError(pollErr)
		s.NotNil(result)
		// when ratelimit is set to zero, poller is expected to return empty result
		// reset ratelimit, poll again and make sure task is returned this time
		s.True(isEmptyToken(result.TaskToken))
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, pollErr = pollFunc(_defaultTaskDispatchRPS, group)
		}()
		wg.Wait()
		s.NoError(err)
		s.NoError(pollErr)
		s.NotNil(result)
		s.False(isEmptyToken(result.TaskToken))
		s.assertPollTaskResponse(taskType, testParam, scheduleID, result)
	}
	s.EqualValues(throttledTaskCount, s.taskManager.getCreateTaskCount(testParam.TaskListID))
	s.EqualValues(0, s.taskManager.getTaskCount(testParam.TaskListID))
	expectedRange := getExpectedRange(initialRangeID, int(taskCount+throttledTaskCount), rangeSize)
	// Due to conflicts some ids are skipped and more real ranges are used.
	s.True(expectedRange <= s.taskManager.getTaskListManager(testParam.TaskListID).rangeID)

	if throttledTaskCount > 0 {
		syncCtr := scope.Snapshot().Counters()["test.sync_throttle_count_per_tl+domain="+matchingTestDomainName+",operation=TaskListMgr,tasklist="+testParam.TaskList.Name]
		s.EqualValues(throttledTaskCount, int(syncCtr.Value()))
	}

	// check the poller information
	descResp, err := s.matchingEngine.DescribeTaskList(s.handlerContext, &types.MatchingDescribeTaskListRequest{
		DomainUUID: testParam.DomainID,
		DescRequest: &types.DescribeTaskListRequest{
			TaskList:              testParam.TaskList,
			TaskListType:          testParam.TaskListType,
			IncludeTaskListStatus: true,
		},
	})
	s.NoError(err)
	s.Equal(1, len(descResp.Pollers))
	s.Equal(testParam.Identity, descResp.Pollers[0].GetIdentity())
	s.NotEmpty(descResp.Pollers[0].GetLastAccessTime())
	s.Equal(_defaultTaskDispatchRPS, descResp.Pollers[0].GetRatePerSecond())
	s.NotNil(descResp.GetTaskListStatus())
	s.True(descResp.GetTaskListStatus().GetRatePerSecond() >= (_defaultTaskDispatchRPS - 1))
}

func (s *matchingEngineSuite) TestConcurrentAddAndPollActivities() {
	s.ConcurrentAddAndPollTasks(persistence.TaskListTypeActivity, 20, 100, false, false)
}

func (s *matchingEngineSuite) TestConcurrentAddAndPollActivitiesWithZeroDispatch() {
	s.ConcurrentAddAndPollTasks(persistence.TaskListTypeActivity, 20, 100, true, false)
}

func (s *matchingEngineSuite) TestConcurrentAddAndPollDecisions() {
	s.ConcurrentAddAndPollTasks(persistence.TaskListTypeDecision, 20, 100, false, false)
}

func (s *matchingEngineSuite) TestConcurrentAddAndPollActivitiesIsolation() {
	s.ConcurrentAddAndPollTasks(persistence.TaskListTypeActivity, 20, 100, false, true)
}

func (s *matchingEngineSuite) TestConcurrentAddAndPollActivitiesWithZeroDispatchIsolation() {
	s.ConcurrentAddAndPollTasks(persistence.TaskListTypeActivity, 20, 100, true, true)
}

func (s *matchingEngineSuite) TestConcurrentAddAndPollDecisionsIsolation() {
	s.ConcurrentAddAndPollTasks(persistence.TaskListTypeDecision, 20, 100, false, true)
}

func (s *matchingEngineSuite) ConcurrentAddAndPollTasks(taskType int, workerCount int, taskCount int64, throttled, enableIsolation bool) {
	s.matchingEngine.config.EnableTasklistIsolation = dynamicconfig.GetBoolPropertyFnFilteredByDomainID(enableIsolation)
	isolationGroups := s.matchingEngine.config.AllIsolationGroups
	dispatchLimitFn := func(wc int, tc int64) float64 {
		return _defaultTaskDispatchRPS
	}
	if throttled {
		dispatchLimitFn = func(wc int, tc int64) float64 {
			if tc%50 == 0 && wc%5 == 0 { // Gets triggered atleast 20 times
				return 0
			}
			return _defaultTaskDispatchRPS
		}
	}
	scope := tally.NewTestScope("test", nil)
	s.matchingEngine.metricsClient = metrics.NewClient(scope, metrics.Matching)

	const initialRangeID = 0
	const rangeSize = 3
	var scheduleID int64 = 123

	testParam := newTestParam(taskType)
	tlKind := types.TaskListKindNormal
	dispatchTTL := time.Nanosecond
	s.matchingEngine.config.RangeSize = rangeSize // override to low number for the test
	dPtr := _defaultTaskDispatchRPS

	mgr, err := newTaskListManager(s.matchingEngine, testParam.TaskListID, &tlKind, s.matchingEngine.config, time.Now())
	s.NoError(err)

	mgrImpl := mgr.(*taskListManagerImpl)
	mgrImpl.matcher.limiter = quotas.NewRateLimiter(&dPtr, dispatchTTL, _minBurst)
	s.matchingEngine.updateTaskList(testParam.TaskListID, mgr)
	s.taskManager.getTaskListManager(testParam.TaskListID).rangeID = initialRangeID
	s.NoError(mgr.Start())

	s.setupGetDrainStatus()

	var wg sync.WaitGroup
	wg.Add(2 * workerCount)

	for p := 0; p < workerCount; p++ {
		go func() {
			defer wg.Done()
			for i := int64(0); i < taskCount; i++ {
				group := isolationGroups[int(i)%len(isolationGroups)] // let each worker to generate tasks for all isolation groups
				addRequest := &addTaskRequest{
					TaskType:                      taskType,
					DomainUUID:                    testParam.DomainID,
					Execution:                     testParam.WorkflowExecution,
					ScheduleID:                    scheduleID,
					TaskList:                      testParam.TaskList,
					ScheduleToStartTimeoutSeconds: 1,
					PartitionConfig:               map[string]string{partition.IsolationGroupKey: group},
				}
				_, err := addTask(s.matchingEngine, s.handlerContext, addRequest)
				if err != nil {
					s.logger.Info("Failure in AddActivityTask", tag.Error(err))
					i--
				}
			}
		}()
	}

	s.setupRecordTaskStartedMock(taskType, testParam, false)

	for p := 0; p < workerCount; p++ {
		go func(wNum int) {
			defer wg.Done()
			for i := int64(0); i < taskCount; {
				maxDispatch := dispatchLimitFn(wNum, i)
				group := isolationGroups[int(wNum)%len(isolationGroups)] // let each worker only polls from one isolation group
				pollReq := &pollTaskRequest{
					TaskType:         taskType,
					DomainUUID:       testParam.DomainID,
					TaskList:         testParam.TaskList,
					Identity:         testParam.Identity,
					TaskListMetadata: &types.TaskListMetadata{MaxTasksPerSecond: &maxDispatch},
					IsolationGroup:   group,
				}
				result, err := pollTask(s.matchingEngine, s.handlerContext, pollReq)
				s.NoError(err)
				s.NotNil(result)
				if isEmptyToken(result.TaskToken) {
					s.logger.Debug("empty poll returned")
					continue
				}
				s.assertPollTaskResponse(taskType, testParam, scheduleID, result)
				i++
			}
		}(p)
	}
	wg.Wait()
	totalTasks := int(taskCount) * workerCount
	persisted := s.taskManager.getCreateTaskCount(testParam.TaskListID)
	s.True(persisted < totalTasks)
	expectedRange := getExpectedRange(initialRangeID, persisted, rangeSize)
	// Due to conflicts some ids are skipped and more real ranges are used.
	s.True(expectedRange <= s.taskManager.getTaskListManager(testParam.TaskListID).rangeID)
	s.EqualValues(0, s.taskManager.getTaskCount(testParam.TaskListID))

	syncCtr := scope.Snapshot().Counters()["test.sync_throttle_count_per_tl+domain="+matchingTestDomainName+",operation=TaskListMgr,tasklist="+testParam.TaskList.Name]
	bufCtr := scope.Snapshot().Counters()["test.buffer_throttle_count_per_tl+domain="+matchingTestDomainName+",operation=TaskListMgr,tasklist="+testParam.TaskList.Name]
	total := int64(0)
	if syncCtr != nil {
		total += syncCtr.Value()
	}
	if bufCtr != nil {
		total += bufCtr.Value()
	}
	if throttled {
		// atleast once from 0 dispatch poll, and until TTL is hit at which time throttle limit is reset
		// hard to predict exactly how many times, since the atomic.Value load might not have updated.
		s.True(total >= 1)
	} else {
		s.EqualValues(0, total)
	}
}

func (s *matchingEngineSuite) TestPollActivityWithExpiredContext() {
	s.PollWithExpiredContext(persistence.TaskListTypeActivity)
}

func (s *matchingEngineSuite) TestPollDecisionWithExpiredContext() {
	s.PollWithExpiredContext(persistence.TaskListTypeDecision)
}

func (s *matchingEngineSuite) PollWithExpiredContext(taskType int) {
	identity := "nobody"
	domainID := "domainId"
	tl := "makeToast"

	taskList := &types.TaskList{Name: tl}

	// Try with cancelled context
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	cancel()
	s.handlerContext.Context = ctx
	pollReq := &pollTaskRequest{
		TaskType:   taskType,
		DomainUUID: domainID,
		TaskList:   taskList,
		Identity:   identity,
	}
	_, err := pollTask(s.matchingEngine, s.handlerContext, pollReq)
	s.Equal(ctx.Err(), err)

	// Try with expired context
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	s.handlerContext.Context = ctx
	resp, err := pollTask(s.matchingEngine, s.handlerContext, pollReq)
	s.Nil(err)
	s.Equal(&pollTaskResponse{}, resp)
}

func (s *matchingEngineSuite) TestMultipleEnginesActivitiesRangeStealing() {
	s.MultipleEnginesTasksRangeStealing(persistence.TaskListTypeActivity)
}

func (s *matchingEngineSuite) TestMultipleEnginesDecisionsRangeStealing() {
	s.MultipleEnginesTasksRangeStealing(persistence.TaskListTypeDecision)
}

func (s *matchingEngineSuite) MultipleEnginesTasksRangeStealing(taskType int) {
	const engineCount = 2
	const taskCount = 400
	const iterations = 2
	const initialRangeID = 0
	const rangeSize = 10
	var scheduleID int64 = 123

	testParam := newTestParam(taskType)
	s.taskManager.getTaskListManager(testParam.TaskListID).rangeID = initialRangeID
	s.matchingEngine.config.RangeSize = rangeSize // override to low number for the test

	var engines []*matchingEngineImpl

	for p := 0; p < engineCount; p++ {
		e := s.newMatchingEngine(defaultTestConfig(), s.taskManager)
		e.config.RangeSize = rangeSize
		engines = append(engines, e)
		e.Start()
	}

	for j := 0; j < iterations; j++ {
		for p := 0; p < engineCount; p++ {
			engine := engines[p]
			for i := int64(0); i < taskCount; i++ {
				addRequest := &addTaskRequest{
					TaskType:                      taskType,
					DomainUUID:                    testParam.DomainID,
					Execution:                     testParam.WorkflowExecution,
					ScheduleID:                    scheduleID,
					TaskList:                      testParam.TaskList,
					ScheduleToStartTimeoutSeconds: 600,
				}

				_, err := addTask(engine, s.handlerContext, addRequest)
				if err != nil {
					s.Require().IsType(&persistence.ConditionFailedError{}, err)
					i-- // retry adding
				}
			}
		}
	}
	s.EqualValues(iterations*engineCount*taskCount, s.taskManager.getCreateTaskCount(testParam.TaskListID))

	s.setupRecordTaskStartedMock(taskType, testParam, true)

	for j := 0; j < iterations; j++ {
		for p := 0; p < engineCount; p++ {
			engine := engines[p]
			for i := int64(0); i < taskCount; /* incremented explicitly to skip empty polls */ {
				pollReq := &pollTaskRequest{
					TaskType:   taskType,
					DomainUUID: testParam.DomainID,
					TaskList:   testParam.TaskList,
					Identity:   testParam.Identity,
				}
				result, err := pollTask(engine, s.handlerContext, pollReq)
				s.Require().NoError(err)
				s.NotNil(result)
				if isEmptyToken(result.TaskToken) {
					s.logger.Debug("empty poll returned")
					continue
				}
				s.assertPollTaskResponse(taskType, testParam, scheduleID, result)
				i++
			}
		}
	}

	for _, e := range engines {
		e.Stop()
	}

	s.EqualValues(0, s.taskManager.getTaskCount(testParam.TaskListID))
	totalTasks := taskCount * engineCount * iterations
	persisted := s.taskManager.getCreateTaskCount(testParam.TaskListID)
	// No sync matching as all messages are added first
	s.EqualValues(totalTasks, persisted)
	expectedRange := getExpectedRange(initialRangeID, persisted, rangeSize)
	// Due to conflicts some ids are skipped and more real ranges are used.
	s.True(expectedRange <= s.taskManager.getTaskListManager(testParam.TaskListID).rangeID)
}

func (s *matchingEngineSuite) TestAddTaskAfterStartFailure() {
	taskType := persistence.TaskListTypeActivity
	runID := "run1"
	workflowID := "workflow1"
	workflowExecution := types.WorkflowExecution{RunID: runID, WorkflowID: workflowID}

	domainID := "domainId"
	tl := "makeToast"
	tlID := newTestTaskListID(domainID, tl, taskType)
	tlKind := types.TaskListKindNormal

	taskList := &types.TaskList{Name: tl}

	scheduleID := int64(0)
	addRequest := &addTaskRequest{
		TaskType:                      taskType,
		DomainUUID:                    domainID,
		Execution:                     &workflowExecution,
		ScheduleID:                    scheduleID,
		TaskList:                      taskList,
		ScheduleToStartTimeoutSeconds: 1,
	}

	_, err := addTask(s.matchingEngine, s.handlerContext, addRequest)
	s.NoError(err)
	s.EqualValues(1, s.taskManager.getTaskCount(tlID))

	ctx, err := s.matchingEngine.getTask(context.Background(), tlID, nil, &tlKind)
	s.NoError(err)

	ctx.finish(errors.New("test error"))
	s.EqualValues(1, s.taskManager.getTaskCount(tlID))
	ctx2, err := s.matchingEngine.getTask(context.Background(), tlID, nil, &tlKind)
	s.NoError(err)

	s.NotEqual(ctx.event.TaskID, ctx2.event.TaskID)
	s.Equal(ctx.event.WorkflowID, ctx2.event.WorkflowID)
	s.Equal(ctx.event.RunID, ctx2.event.RunID)
	s.Equal(ctx.event.ScheduleID, ctx2.event.ScheduleID)

	ctx2.finish(nil)
	s.EqualValues(0, s.taskManager.getTaskCount(tlID))
}

func (s *matchingEngineSuite) TestTaskListManagerGetTaskBatch() {
	taskType := persistence.TaskListTypeActivity

	const taskCount = 1200
	const rangeSize = 10
	s.matchingEngine.config.RangeSize = rangeSize
	testParam := newTestParam(taskType)

	// add taskCount tasks
	for i := int64(0); i < taskCount; i++ {
		scheduleID := i * 3
		addRequest := &addTaskRequest{
			TaskType:                      taskType,
			DomainUUID:                    testParam.DomainID,
			Execution:                     testParam.WorkflowExecution,
			ScheduleID:                    scheduleID,
			TaskList:                      testParam.TaskList,
			ScheduleToStartTimeoutSeconds: 1,
		}

		_, err := addTask(s.matchingEngine, s.handlerContext, addRequest)
		s.NoError(err)
	}

	tlMgr, ok := s.matchingEngine.taskLists[*testParam.TaskListID].(*taskListManagerImpl)
	s.True(ok, "taskListManger doesn't implement taskListManager interface")
	s.EqualValues(taskCount, s.taskManager.getTaskCount(testParam.TaskListID))

	// wait until all tasks are read by the task pump and enqeued into the in-memory buffer
	// at the end of this step, ackManager readLevel will also be equal to the buffer size
	expectedBufSize := common.MinInt(cap(tlMgr.taskReader.taskBuffers[defaultTaskBufferIsolationGroup]), taskCount)
	s.True(s.awaitCondition(func() bool {
		return len(tlMgr.taskReader.taskBuffers[defaultTaskBufferIsolationGroup]) == expectedBufSize
	}, time.Second))

	// stop all goroutines that read / write tasks in the background
	// remainder of this test works with the in-memory buffer
	tlMgr.Stop()

	// SetReadLevel should NEVER be called without updating ackManager.outstandingTasks
	// This is only for unit test purpose
	tlMgr.taskAckManager.SetReadLevel(tlMgr.taskWriter.GetMaxReadLevel())
	tasks, readLevel, isReadBatchDone, err := tlMgr.taskReader.getTaskBatch()
	s.Nil(err)
	s.EqualValues(0, len(tasks))
	s.EqualValues(tlMgr.taskWriter.GetMaxReadLevel(), readLevel)
	s.True(isReadBatchDone)

	tlMgr.taskAckManager.SetReadLevel(0)
	tasks, readLevel, isReadBatchDone, err = tlMgr.taskReader.getTaskBatch()
	s.Nil(err)
	s.EqualValues(rangeSize, len(tasks))
	s.EqualValues(rangeSize, readLevel)
	s.True(isReadBatchDone)

	s.setupRecordTaskStartedMock(taskType, testParam, false)

	// reset the ackManager readLevel to the buffer size and consume
	// the in-memory tasks by calling Poll API - assert ackMgr state
	// at the end
	tlMgr.taskAckManager.SetReadLevel(int64(expectedBufSize))

	// complete rangeSize events
	for i := int64(0); i < rangeSize; i++ {
		pollReq := &pollTaskRequest{
			TaskType:   taskType,
			DomainUUID: testParam.DomainID,
			TaskList:   testParam.TaskList,
			Identity:   testParam.Identity,
		}
		result, err := pollTask(s.matchingEngine, s.handlerContext, pollReq)
		s.NoError(err)
		s.NotNil(result)
		s.NotEqual(&pollTaskResponse{}, result)
		if isEmptyToken(result.TaskToken) {
			s.logger.Debug("empty poll returned")
			continue
		}
	}
	s.EqualValues(taskCount-rangeSize, s.taskManager.getTaskCount(testParam.TaskListID))
	tasks, _, isReadBatchDone, err = tlMgr.taskReader.getTaskBatch()
	s.Nil(err)
	s.True(0 < len(tasks) && len(tasks) <= rangeSize)
	s.True(isReadBatchDone)
}

func (s *matchingEngineSuite) TestTaskListManagerGetTaskBatch_ReadBatchDone() {
	domainID := "domainId"
	tl := "makeToast"
	tlID := newTestTaskListID(domainID, tl, persistence.TaskListTypeActivity)
	tlNormal := types.TaskListKindNormal

	const rangeSize = 10
	const maxReadLevel = int64(120)
	config := defaultTestConfig()
	config.RangeSize = rangeSize
	tlMgr0, err := newTaskListManager(s.matchingEngine, tlID, &tlNormal, config, time.Now())
	s.NoError(err)

	tlMgr, ok := tlMgr0.(*taskListManagerImpl)
	s.True(ok)

	tlMgr.taskAckManager.SetReadLevel(0)
	atomic.StoreInt64(&tlMgr.taskWriter.maxReadLevel, maxReadLevel)
	tasks, readLevel, isReadBatchDone, err := tlMgr.taskReader.getTaskBatch()
	s.Empty(tasks)
	s.Equal(int64(rangeSize*10), readLevel)
	s.False(isReadBatchDone)
	s.NoError(err)

	tlMgr.taskAckManager.SetReadLevel(readLevel)
	tasks, readLevel, isReadBatchDone, err = tlMgr.taskReader.getTaskBatch()
	s.Empty(tasks)
	s.Equal(maxReadLevel, readLevel)
	s.True(isReadBatchDone)
	s.NoError(err)
}

func (s *matchingEngineSuite) TestActivityExpiryAndCompletion() {
	s.TaskExpiryAndCompletion(persistence.TaskListTypeActivity)
}

func (s *matchingEngineSuite) TestDecisionExpiryAndCompletion() {
	s.TaskExpiryAndCompletion(persistence.TaskListTypeDecision)
}

func (s *matchingEngineSuite) TaskExpiryAndCompletion(taskType int) {
	const taskCount = 20
	const rangeSize = 10
	s.matchingEngine.config.RangeSize = rangeSize
	s.matchingEngine.config.MaxTaskDeleteBatchSize = dynamicconfig.GetIntPropertyFilteredByTaskListInfo(2)
	// set idle timer check to a really small value to assert that we don't accidentally drop tasks while blocking
	// on enqueuing a task to task buffer
	s.matchingEngine.config.IdleTasklistCheckInterval = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(10 * time.Millisecond)

	testParam := newTestParam(taskType)
	testCases := []struct {
		batchSize          int
		maxTimeBtwnDeletes time.Duration
	}{
		{2, time.Minute},       // test taskGC deleting due to size threshold
		{100, time.Nanosecond}, // test taskGC deleting due to time condition
	}

	for _, tc := range testCases {
		for i := int64(0); i < taskCount; i++ {
			scheduleID := i * 3
			addRequest := &addTaskRequest{
				TaskType:                      taskType,
				DomainUUID:                    testParam.DomainID,
				Execution:                     testParam.WorkflowExecution,
				ScheduleID:                    scheduleID,
				TaskList:                      testParam.TaskList,
				ScheduleToStartTimeoutSeconds: 5,
			}
			if i%2 == 0 {
				// simulates creating a task whose scheduledToStartTimeout is already expired
				addRequest.ScheduleToStartTimeoutSeconds = -5
			}

			_, err := addTask(s.matchingEngine, s.handlerContext, addRequest)
			s.NoError(err)
		}

		tlMgr, ok := s.matchingEngine.taskLists[*testParam.TaskListID].(*taskListManagerImpl)
		s.True(ok, "failed to load task list")
		s.EqualValues(taskCount, s.taskManager.getTaskCount(testParam.TaskListID))

		// wait until all tasks are loaded by into in-memory buffers by task list manager
		// the buffer size should be one less than expected because dispatcher will dequeue the head
		s.True(s.awaitCondition(func() bool {
			return len(tlMgr.taskReader.taskBuffers[defaultTaskBufferIsolationGroup]) >= (taskCount/2 - 1)
		}, time.Second))

		maxTimeBetweenTaskDeletes = tc.maxTimeBtwnDeletes
		s.matchingEngine.config.MaxTaskDeleteBatchSize = dynamicconfig.GetIntPropertyFilteredByTaskListInfo(tc.batchSize)

		s.setupRecordTaskStartedMock(taskType, testParam, false)

		pollReq := &pollTaskRequest{
			TaskType:   taskType,
			DomainUUID: testParam.DomainID,
			TaskList:   testParam.TaskList,
			Identity:   testParam.Identity,
		}

		remaining := taskCount
		for i := 0; i < 2; i++ {
			// verify that (1) expired tasks are not returned in poll result (2) taskCleaner deletes tasks correctly
			for i := int64(0); i < taskCount/4; i++ {
				result, err := pollTask(s.matchingEngine, s.handlerContext, pollReq)
				s.NoError(err)
				s.NotNil(result)
			}
			remaining -= taskCount / 2
			// since every other task is expired, we expect half the tasks to be deleted
			// after poll consumed 1/4th of what is available
			s.EqualValues(remaining, s.taskManager.getTaskCount(testParam.TaskListID))
		}
	}
}

func (s *matchingEngineSuite) TestUnloadActivityTasklistOnIsolationConfigChange() {
	s.UnloadTasklistOnIsolationConfigChange(persistence.TaskListTypeActivity)
}

func (s *matchingEngineSuite) TestUnloadDecisionTasklistOnIsolationConfigChange() {
	s.UnloadTasklistOnIsolationConfigChange(persistence.TaskListTypeDecision)
}

func (s *matchingEngineSuite) UnloadTasklistOnIsolationConfigChange(taskType int) {
	s.matchingEngine.config.LongPollExpirationInterval = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(50 * time.Millisecond)
	s.matchingEngine.config.EnableTasklistIsolation = dynamicconfig.GetBoolPropertyFnFilteredByDomainID(false)

	const taskCount = 1000
	const initialRangeID = 102
	// TODO: Understand why publish is low when rangeSize is 3
	const rangeSize = 30

	testParam := newTestParam(taskType)
	s.taskManager.getTaskListManager(testParam.TaskListID).rangeID = initialRangeID
	s.matchingEngine.config.RangeSize = rangeSize // override to low number for the test

	addRequest := &addTaskRequest{
		TaskType:                      taskType,
		DomainUUID:                    testParam.DomainID,
		Execution:                     testParam.WorkflowExecution,
		ScheduleID:                    333,
		TaskList:                      testParam.TaskList,
		ScheduleToStartTimeoutSeconds: 1,
	}
	_, err := addTask(s.matchingEngine, s.handlerContext, addRequest)
	s.NoError(err)

	// enable isolation and verify that poller should not get any task
	s.matchingEngine.config.EnableTasklistIsolation = dynamicconfig.GetBoolPropertyFnFilteredByDomainID(true)
	s.setupGetDrainStatus()
	s.setupRecordTaskStartedMock(taskType, testParam, false)

	pollReq := &pollTaskRequest{
		TaskType:   taskType,
		DomainUUID: testParam.DomainID,
		TaskList:   testParam.TaskList,
		Identity:   testParam.Identity,
	}
	result, err := pollTask(s.matchingEngine, s.handlerContext, pollReq)
	s.NoError(err)
	s.Equal(&pollTaskResponse{}, result)

	result, err = pollTask(s.matchingEngine, s.handlerContext, pollReq)
	s.NoError(err)
	s.NotNil(result)
	s.assertPollTaskResponse(taskType, testParam, 333, result)

	// disable isolation again and verify add tasklist should fail
	s.matchingEngine.config.EnableTasklistIsolation = dynamicconfig.GetBoolPropertyFnFilteredByDomainID(false)
	_, err = addTask(s.matchingEngine, s.handlerContext, addRequest)
	s.Error(err)
	s.Contains(err.Error(), errShutdown.Error())
}

func (s *matchingEngineSuite) TestDrainActivityBacklogNoPollersIsolationGroup() {
	s.DrainBacklogNoPollersIsolationGroup(persistence.TaskListTypeActivity)
}

func (s *matchingEngineSuite) TestDrainDecisionBacklogNoPollersIsolationGroup() {
	s.DrainBacklogNoPollersIsolationGroup(persistence.TaskListTypeDecision)
}

func (s *matchingEngineSuite) DrainBacklogNoPollersIsolationGroup(taskType int) {
	s.matchingEngine.config.LongPollExpirationInterval = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(10 * time.Millisecond)
	s.matchingEngine.config.EnableTasklistIsolation = dynamicconfig.GetBoolPropertyFnFilteredByDomainID(true)
	s.matchingEngine.config.AsyncTaskDispatchTimeout = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(10 * time.Millisecond)

	isolationGroups := s.matchingEngine.config.AllIsolationGroups

	const taskCount = 1000
	const initialRangeID = 102
	// TODO: Understand why publish is low when rangeSize is 3
	const rangeSize = 30
	// use a const scheduleID because we don't know the order of task polled
	const scheduleID = 11111

	testParam := newTestParam(taskType)
	s.taskManager.getTaskListManager(testParam.TaskListID).rangeID = initialRangeID
	s.matchingEngine.config.RangeSize = rangeSize // override to low number for the test
	mgr, err := s.matchingEngine.getTaskListManager(testParam.TaskListID, testParam.TaskList.Kind)
	s.NoError(err)
	tlMgr, ok := mgr.(*taskListManagerImpl)
	s.True(ok)
	tlMgr.createTime = time.Now().Add(-time.Minute)

	s.setupGetDrainStatus()

	for i := int64(0); i < taskCount; i++ {
		addRequest := &addTaskRequest{
			TaskType:                      taskType,
			DomainUUID:                    testParam.DomainID,
			Execution:                     testParam.WorkflowExecution,
			ScheduleID:                    scheduleID,
			TaskList:                      testParam.TaskList,
			ScheduleToStartTimeoutSeconds: 1,
			PartitionConfig:               map[string]string{partition.IsolationGroupKey: isolationGroups[int(i)%len(isolationGroups)]},
		}
		_, err := addTask(s.matchingEngine, s.handlerContext, addRequest)
		s.NoError(err)
	}
	s.EqualValues(taskCount, s.taskManager.getTaskCount(testParam.TaskListID))

	s.setupRecordTaskStartedMock(taskType, testParam, false)

	for i := int64(0); i < taskCount; {
		pollReq := &pollTaskRequest{
			TaskType:       taskType,
			DomainUUID:     testParam.DomainID,
			TaskList:       testParam.TaskList,
			Identity:       testParam.Identity,
			IsolationGroup: isolationGroups[0],
		}
		result, err := pollTask(s.matchingEngine, s.handlerContext, pollReq)
		s.NoError(err)
		s.NotNil(result)
		if isEmptyToken(result.TaskToken) {
			s.logger.Debug("empty poll returned")
			continue
		}
		s.assertPollTaskResponse(taskType, testParam, scheduleID, result)
		i++
	}
	s.EqualValues(0, s.taskManager.getTaskCount(testParam.TaskListID))
	expectedRange := getExpectedRange(initialRangeID, taskCount, rangeSize)
	// Due to conflicts some ids are skipped and more real ranges are used.
	s.True(expectedRange <= s.taskManager.getTaskListManager(testParam.TaskListID).rangeID)
}

func (s *matchingEngineSuite) TestAddStickyDecisionNoPollerIsolation() {
	s.T().Skip("skip test until we re-enable isolation for sticky tasklist")
	taskType := persistence.TaskListTypeDecision
	s.matchingEngine.config.LongPollExpirationInterval = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(10 * time.Millisecond)
	s.matchingEngine.config.EnableTasklistIsolation = dynamicconfig.GetBoolPropertyFnFilteredByDomainID(true)
	s.matchingEngine.config.AsyncTaskDispatchTimeout = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(100 * time.Millisecond)

	isolationGroups := s.matchingEngine.config.AllIsolationGroups

	const taskCount = 10
	const initialRangeID = 102
	// TODO: Understand why publish is low when rangeSize is 3
	const rangeSize = 30

	testParam := newTestParam(taskType)
	stickyKind := types.TaskListKindSticky
	testParam.TaskList.Kind = &stickyKind
	s.taskManager.getTaskListManager(testParam.TaskListID).rangeID = initialRangeID
	s.matchingEngine.config.RangeSize = rangeSize // override to low number for the test

	s.setupGetDrainStatus()

	pollReq := &pollTaskRequest{
		TaskType:       taskType,
		DomainUUID:     testParam.DomainID,
		TaskList:       testParam.TaskList,
		Identity:       testParam.Identity,
		IsolationGroup: isolationGroups[0],
	}
	result, err := pollTask(s.matchingEngine, s.handlerContext, pollReq)
	s.NoError(err)
	s.True(isEmptyToken(result.TaskToken))

	count := int64(0)
	scheduleIDs := []int64{}
	for i := int64(0); i < taskCount; i++ {
		scheduleID := i * 3
		addRequest := &addTaskRequest{
			TaskType:                      taskType,
			DomainUUID:                    testParam.DomainID,
			Execution:                     testParam.WorkflowExecution,
			ScheduleID:                    scheduleID,
			TaskList:                      testParam.TaskList,
			ScheduleToStartTimeoutSeconds: 1,
			PartitionConfig:               map[string]string{partition.IsolationGroupKey: isolationGroups[int(i)%len(isolationGroups)]},
		}
		_, err := addTask(s.matchingEngine, s.handlerContext, addRequest)
		if int(i)%len(isolationGroups) == 0 {
			s.NoError(err)
			count++
			scheduleIDs = append(scheduleIDs, scheduleID)
		} else {
			s.Error(err)
			s.Contains(err.Error(), "sticky worker is unavailable")
		}
	}
	s.EqualValues(count, s.taskManager.getTaskCount(testParam.TaskListID))

	s.setupRecordTaskStartedMock(taskType, testParam, false)

	for i := int64(0); i < count; {
		scheduleID := scheduleIDs[i]
		pollReq := &pollTaskRequest{
			TaskType:       taskType,
			DomainUUID:     testParam.DomainID,
			TaskList:       testParam.TaskList,
			Identity:       testParam.Identity,
			IsolationGroup: isolationGroups[0],
		}
		result, err := pollTask(s.matchingEngine, s.handlerContext, pollReq)
		s.NoError(err)
		s.NotNil(result)
		if isEmptyToken(result.TaskToken) {
			s.logger.Debug("empty poll returned")
			continue
		}
		s.assertPollTaskResponse(taskType, testParam, scheduleID, result)
		i++
	}
	s.EqualValues(0, s.taskManager.getTaskCount(testParam.TaskListID))
	expectedRange := getExpectedRange(initialRangeID, taskCount, rangeSize)
	// Due to conflicts some ids are skipped and more real ranges are used.
	s.True(expectedRange <= s.taskManager.getTaskListManager(testParam.TaskListID).rangeID)
}

func (s *matchingEngineSuite) setupRecordTaskStartedMock(taskType int, param *testParam, checkDuplicate bool) {
	startedTasks := make(map[int64]bool)
	if taskType == persistence.TaskListTypeActivity {
		s.mockHistoryClient.EXPECT().RecordActivityTaskStarted(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, taskRequest *types.RecordActivityTaskStartedRequest, option ...yarpc.CallOption) (*types.RecordActivityTaskStartedResponse, error) {
				s.logger.Debug("Mock Received RecordActivityTaskStartedRequest")
				if checkDuplicate {
					if _, ok := startedTasks[taskRequest.TaskID]; ok {
						return nil, &types.EventAlreadyStartedError{Message: "already started"}
					}
					startedTasks[taskRequest.TaskID] = true
				}
				return &types.RecordActivityTaskStartedResponse{
					ScheduledEvent: newActivityTaskScheduledEvent(taskRequest.ScheduleID, 0,
						&types.ScheduleActivityTaskDecisionAttributes{
							ActivityID:                    param.ActivityID,
							TaskList:                      param.TaskList,
							ActivityType:                  param.ActivityType,
							Input:                         param.ActivityInput,
							Header:                        param.ActivityHeader,
							ScheduleToCloseTimeoutSeconds: common.Int32Ptr(100),
							ScheduleToStartTimeoutSeconds: common.Int32Ptr(50),
							StartToCloseTimeoutSeconds:    common.Int32Ptr(50),
							HeartbeatTimeoutSeconds:       common.Int32Ptr(10),
						}),
				}, nil
			}).AnyTimes()
	} else {
		s.mockHistoryClient.EXPECT().RecordDecisionTaskStarted(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, taskRequest *types.RecordDecisionTaskStartedRequest, option ...yarpc.CallOption) (*types.RecordDecisionTaskStartedResponse, error) {
				s.logger.Debug("Mock Received RecordDecisionTaskStartedRequest")
				if checkDuplicate {
					if _, ok := startedTasks[taskRequest.TaskID]; ok {
						return nil, &types.EventAlreadyStartedError{Message: "already started"}
					}
					startedTasks[taskRequest.TaskID] = true
				}
				return &types.RecordDecisionTaskStartedResponse{
					PreviousStartedEventID: &param.StartedEventID,
					StartedEventID:         param.StartedEventID,
					ScheduledEventID:       taskRequest.ScheduleID,
					WorkflowType:           param.WorkflowType,
				}, nil
			}).AnyTimes()
	}
}

func (s *matchingEngineSuite) setupGetDrainStatus() {
	s.mockIsolationStore.EXPECT().GetListValue(dynamicconfig.DefaultIsolationGroupConfigStoreManagerGlobalMapping, nil).Return(nil, nil).AnyTimes()
}

func (s *matchingEngineSuite) awaitCondition(cond func() bool, timeout time.Duration) bool {
	expiry := time.Now().Add(timeout)
	for !cond() {
		time.Sleep(time.Millisecond * 5)
		if time.Now().After(expiry) {
			return false
		}
	}
	return true
}

func (s *matchingEngineSuite) assertPollTaskResponse(taskType int, param *testParam, scheduleID int64, actual *pollTaskResponse) {
	if taskType == persistence.TaskListTypeActivity {
		token := &common.TaskToken{
			DomainID:     param.DomainID,
			WorkflowID:   param.WorkflowExecution.WorkflowID,
			RunID:        param.WorkflowExecution.RunID,
			ScheduleID:   scheduleID,
			ActivityID:   param.ActivityID,
			ActivityType: param.ActivityType.Name,
		}
		s.EqualValues(token, actual.TaskToken)
		s.EqualValues(param.ActivityID, actual.ActivityID)
		s.EqualValues(param.ActivityType, actual.ActivityType)
		s.EqualValues(param.ActivityInput, actual.Input)
		s.EqualValues(param.ActivityHeader, actual.Header)
		s.EqualValues(param.WorkflowExecution, actual.WorkflowExecution)
	} else {
		token := &common.TaskToken{
			DomainID:   param.DomainID,
			WorkflowID: param.WorkflowExecution.WorkflowID,
			RunID:      param.WorkflowExecution.RunID,
			ScheduleID: scheduleID,
		}
		s.EqualValues(token, actual.TaskToken)
		s.EqualValues(param.WorkflowExecution, actual.WorkflowExecution)
		s.EqualValues(param.WorkflowType, actual.WorkflowType)
		s.EqualValues(param.StartedEventID, actual.StartedEventID)
	}
}

func (s *matchingEngineSuite) TestConfigDefaultHostName() {
	configEmpty := Config{}
	s.NotEqualValues(s.matchingEngine.config.HostName, configEmpty.HostName)
	s.EqualValues(configEmpty.HostName, "")
}

func newActivityTaskScheduledEvent(eventID int64, decisionTaskCompletedEventID int64,
	scheduleAttributes *types.ScheduleActivityTaskDecisionAttributes) *types.HistoryEvent {
	historyEvent := newHistoryEvent(eventID, types.EventTypeActivityTaskScheduled)
	attributes := &types.ActivityTaskScheduledEventAttributes{}
	attributes.ActivityID = scheduleAttributes.ActivityID
	attributes.ActivityType = scheduleAttributes.ActivityType
	attributes.TaskList = scheduleAttributes.TaskList
	attributes.Input = scheduleAttributes.Input
	attributes.Header = scheduleAttributes.Header
	attributes.ScheduleToCloseTimeoutSeconds = scheduleAttributes.ScheduleToCloseTimeoutSeconds
	attributes.ScheduleToStartTimeoutSeconds = scheduleAttributes.ScheduleToStartTimeoutSeconds
	attributes.StartToCloseTimeoutSeconds = scheduleAttributes.StartToCloseTimeoutSeconds
	attributes.HeartbeatTimeoutSeconds = scheduleAttributes.HeartbeatTimeoutSeconds
	attributes.DecisionTaskCompletedEventID = decisionTaskCompletedEventID
	historyEvent.ActivityTaskScheduledEventAttributes = attributes

	return historyEvent
}

func newHistoryEvent(eventID int64, eventType types.EventType) *types.HistoryEvent {
	ts := common.Int64Ptr(time.Now().UnixNano())
	historyEvent := &types.HistoryEvent{}
	historyEvent.ID = eventID
	historyEvent.Timestamp = ts
	historyEvent.EventType = &eventType

	return historyEvent
}

var _ persistence.TaskManager = (*testTaskManager)(nil) // Asserts that interface is indeed implemented

type testTaskManager struct {
	sync.Mutex
	taskLists map[taskListID]*testTaskListManager
	logger    log.Logger
}

func newTestTaskManager(logger log.Logger) *testTaskManager {
	return &testTaskManager{taskLists: make(map[taskListID]*testTaskListManager), logger: logger}
}

func (m *testTaskManager) GetName() string {
	return "test"
}

func (m *testTaskManager) Close() {
}

func (m *testTaskManager) getTaskListManager(id *taskListID) *testTaskListManager {
	m.Lock()
	defer m.Unlock()
	result, ok := m.taskLists[*id]
	if ok {
		return result
	}
	result = newTestTaskListManager()
	m.taskLists[*id] = result
	return result
}

type testTaskListManager struct {
	sync.Mutex
	rangeID         int64
	ackLevel        int64
	createTaskCount int
	tasks           *treemap.Map
}

func Int64Comparator(a, b interface{}) int {
	aAsserted := a.(int64)
	bAsserted := b.(int64)
	switch {
	case aAsserted > bAsserted:
		return 1
	case aAsserted < bAsserted:
		return -1
	default:
		return 0
	}
}

func newTestTaskListManager() *testTaskListManager {
	return &testTaskListManager{tasks: treemap.NewWith(Int64Comparator)}
}

func newTestTaskListID(domainID string, name string, taskType int) *taskListID {
	result, err := newTaskListID(domainID, name, taskType)
	if err != nil {
		panic(fmt.Sprintf("newTaskListID failed with error %v", err))
	}
	return result
}

// LeaseTaskList provides a mock function with given fields: ctx, request
func (m *testTaskManager) LeaseTaskList(
	_ context.Context,
	request *persistence.LeaseTaskListRequest,
) (*persistence.LeaseTaskListResponse, error) {
	tlm := m.getTaskListManager(newTestTaskListID(request.DomainID, request.TaskList, request.TaskType))
	tlm.Lock()
	defer tlm.Unlock()
	tlm.rangeID++
	m.logger.Debug(fmt.Sprintf("LeaseTaskList rangeID=%v", tlm.rangeID))

	return &persistence.LeaseTaskListResponse{
		TaskListInfo: &persistence.TaskListInfo{
			AckLevel: tlm.ackLevel,
			DomainID: request.DomainID,
			Name:     request.TaskList,
			TaskType: request.TaskType,
			RangeID:  tlm.rangeID,
			Kind:     request.TaskListKind,
		},
	}, nil
}

// UpdateTaskList provides a mock function with given fields: ctx, request
func (m *testTaskManager) UpdateTaskList(
	_ context.Context,
	request *persistence.UpdateTaskListRequest,
) (*persistence.UpdateTaskListResponse, error) {
	m.logger.Debug(fmt.Sprintf("UpdateTaskList taskListInfo=%v, ackLevel=%v", request.TaskListInfo, request.TaskListInfo.AckLevel))

	tli := request.TaskListInfo
	tlm := m.getTaskListManager(newTestTaskListID(tli.DomainID, tli.Name, tli.TaskType))

	tlm.Lock()
	defer tlm.Unlock()
	if tlm.rangeID != tli.RangeID {
		return nil, &persistence.ConditionFailedError{
			Msg: fmt.Sprintf("Failed to update task list: name=%v, type=%v", tli.Name, tli.TaskType),
		}
	}
	tlm.ackLevel = tli.AckLevel
	return &persistence.UpdateTaskListResponse{}, nil
}

// CompleteTask provides a mock function with given fields: ctx, request
func (m *testTaskManager) CompleteTask(
	_ context.Context,
	request *persistence.CompleteTaskRequest,
) error {
	m.logger.Debug(fmt.Sprintf("CompleteTask taskID=%v, ackLevel=%v", request.TaskID, request.TaskList.AckLevel))
	if request.TaskID <= 0 {
		panic(fmt.Errorf("Invalid taskID=%v", request.TaskID))
	}

	tli := request.TaskList
	tlm := m.getTaskListManager(newTestTaskListID(tli.DomainID, tli.Name, tli.TaskType))

	tlm.Lock()
	defer tlm.Unlock()

	tlm.tasks.Remove(request.TaskID)
	return nil
}

// CompleteTasksLessThan provides a mock function with given fields: ctx, request
func (m *testTaskManager) CompleteTasksLessThan(
	_ context.Context,
	request *persistence.CompleteTasksLessThanRequest,
) (*persistence.CompleteTasksLessThanResponse, error) {
	tlm := m.getTaskListManager(newTestTaskListID(request.DomainID, request.TaskListName, request.TaskType))
	tlm.Lock()
	defer tlm.Unlock()
	rowsDeleted := 0
	keys := tlm.tasks.Keys()
	for _, key := range keys {
		id := key.(int64)
		if id <= request.TaskID {
			tlm.tasks.Remove(id)
			rowsDeleted++
		}
	}
	return &persistence.CompleteTasksLessThanResponse{TasksCompleted: rowsDeleted}, nil
}

// ListTaskList provides a mock function with given fields: ctx, request
func (m *testTaskManager) ListTaskList(
	_ context.Context,
	request *persistence.ListTaskListRequest,
) (*persistence.ListTaskListResponse, error) {
	return nil, fmt.Errorf("unsupported operation")
}

// DeleteTaskList provides a mock function with given fields: ctx, request
func (m *testTaskManager) DeleteTaskList(
	_ context.Context,
	request *persistence.DeleteTaskListRequest,
) error {
	m.Lock()
	defer m.Unlock()
	key := newTestTaskListID(request.DomainID, request.TaskListName, request.TaskListType)
	delete(m.taskLists, *key)
	return nil
}

// CreateTask provides a mock function with given fields: ctx, request
func (m *testTaskManager) CreateTasks(
	_ context.Context,
	request *persistence.CreateTasksRequest,
) (*persistence.CreateTasksResponse, error) {
	domainID := request.TaskListInfo.DomainID
	taskList := request.TaskListInfo.Name
	taskType := request.TaskListInfo.TaskType
	rangeID := request.TaskListInfo.RangeID

	tlm := m.getTaskListManager(newTestTaskListID(domainID, taskList, taskType))
	tlm.Lock()
	defer tlm.Unlock()

	// First validate the entire batch
	for _, task := range request.Tasks {
		m.logger.Debug(fmt.Sprintf("testTaskManager.CreateTask taskID=%v, rangeID=%v", task.TaskID, rangeID))
		if task.TaskID <= 0 {
			panic(fmt.Errorf("Invalid taskID=%v", task.TaskID))
		}

		if tlm.rangeID != rangeID {
			m.logger.Debug(fmt.Sprintf("testTaskManager.CreateTask ConditionFailedError taskID=%v, rangeID: %v, db rangeID: %v",
				task.TaskID, rangeID, tlm.rangeID))

			return nil, &persistence.ConditionFailedError{
				Msg: fmt.Sprintf("testTaskManager.CreateTask failed. TaskList: %v, taskType: %v, rangeID: %v, db rangeID: %v",
					taskList, taskType, rangeID, tlm.rangeID),
			}
		}
		_, ok := tlm.tasks.Get(task.TaskID)
		if ok {
			panic(fmt.Sprintf("Duplicated TaskID %v", task.TaskID))
		}
	}

	// Then insert all tasks if no errors
	for _, task := range request.Tasks {
		scheduleID := task.Data.ScheduleID
		info := &persistence.TaskInfo{
			DomainID:        domainID,
			RunID:           task.Execution.RunID,
			ScheduleID:      scheduleID,
			TaskID:          task.TaskID,
			WorkflowID:      task.Execution.WorkflowID,
			PartitionConfig: task.Data.PartitionConfig,
		}
		if task.Data.ScheduleToStartTimeout != 0 {
			info.Expiry = time.Now().Add(time.Duration(task.Data.ScheduleToStartTimeout) * time.Second)
		}
		tlm.tasks.Put(task.TaskID, info)
		tlm.createTaskCount++
	}

	return &persistence.CreateTasksResponse{}, nil
}

// GetTasks provides a mock function with given fields: ctx, request
func (m *testTaskManager) GetTasks(
	_ context.Context,
	request *persistence.GetTasksRequest,
) (*persistence.GetTasksResponse, error) {
	m.logger.Debug(fmt.Sprintf("testTaskManager.GetTasks readLevel=%v, maxReadLevel=%v", request.ReadLevel, *request.MaxReadLevel))

	tlm := m.getTaskListManager(newTestTaskListID(request.DomainID, request.TaskList, request.TaskType))
	tlm.Lock()
	defer tlm.Unlock()
	var tasks []*persistence.TaskInfo

	it := tlm.tasks.Iterator()
	for it.Next() {
		taskID := it.Key().(int64)
		if taskID <= request.ReadLevel {
			continue
		}
		if taskID > *request.MaxReadLevel {
			break
		}
		tasks = append(tasks, it.Value().(*persistence.TaskInfo))
	}
	return &persistence.GetTasksResponse{
		Tasks: tasks,
	}, nil
}

func (m *testTaskManager) GetTaskListSize(_ context.Context, request *persistence.GetTaskListSizeRequest) (*persistence.GetTaskListSizeResponse, error) {
	tlm := m.getTaskListManager(newTestTaskListID(request.DomainID, request.TaskListName, request.TaskListType))
	tlm.Lock()
	defer tlm.Unlock()
	count := int64(0)
	it := tlm.tasks.Iterator()
	for it.Next() {
		taskID := it.Key().(int64)
		if taskID <= request.AckLevel {
			continue
		}
		count++
	}
	return &persistence.GetTaskListSizeResponse{Size: count}, nil
}

func (m *testTaskManager) GetOrphanTasks(_ context.Context, request *persistence.GetOrphanTasksRequest) (*persistence.GetOrphanTasksResponse, error) {
	return &persistence.GetOrphanTasksResponse{}, nil
}

// getTaskCount returns number of tasks in a task list
func (m *testTaskManager) getTaskCount(taskList *taskListID) int {
	tlm := m.getTaskListManager(taskList)
	tlm.Lock()
	defer tlm.Unlock()
	return tlm.tasks.Size()
}

// getCreateTaskCount returns how many times CreateTask was called
func (m *testTaskManager) getCreateTaskCount(taskList *taskListID) int {
	tlm := m.getTaskListManager(taskList)
	tlm.Lock()
	defer tlm.Unlock()
	return tlm.createTaskCount
}

func (m *testTaskManager) String() string {
	m.Lock()
	defer m.Unlock()
	var result string
	for id, tl := range m.taskLists {
		tl.Lock()
		if id.taskType == persistence.TaskListTypeActivity {
			result += "Activity"
		} else {
			result += "Decision"
		}
		result += " task list " + id.name
		result += "\n"
		result += fmt.Sprintf("AckLevel=%v\n", tl.ackLevel)
		result += fmt.Sprintf("CreateTaskCount=%v\n", tl.createTaskCount)
		result += fmt.Sprintf("RangeID=%v\n", tl.rangeID)
		result += "Tasks=\n"
		for _, t := range tl.tasks.Values() {
			result += spew.Sdump(t)
			result += "\n"

		}
		tl.Unlock()
	}
	return result
}

func validateTimeRange(t time.Time, expectedDuration time.Duration) bool {
	currentTime := time.Now()
	diff := time.Duration(currentTime.UnixNano() - t.UnixNano())
	if diff > expectedDuration {
		fmt.Printf("Current time: %v, Application time: %v, Difference: %v \n", currentTime, t, diff)
		return false
	}
	return true
}

func defaultTestConfig() *Config {
	config := NewConfig(dynamicconfig.NewNopCollection(), "some random hostname")
	config.LongPollExpirationInterval = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(100 * time.Millisecond)
	config.MaxTaskDeleteBatchSize = dynamicconfig.GetIntPropertyFilteredByTaskListInfo(1)
	config.AllIsolationGroups = []string{"datacenterA", "datacenterB"}
	config.GetTasksBatchSize = dynamicconfig.GetIntPropertyFilteredByTaskListInfo(10)
	config.AsyncTaskDispatchTimeout = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(10 * time.Millisecond)
	return config
}

func getExpectedRange(initialRangeID, taskCount, rangeSize int) int64 {
	expectedRange := int64(initialRangeID + taskCount/rangeSize)
	if taskCount%rangeSize > 0 {
		expectedRange++
	}
	return expectedRange
}

type testParam struct {
	DomainID          string
	WorkflowExecution *types.WorkflowExecution
	TaskList          *types.TaskList
	TaskListID        *taskListID
	TaskListType      *types.TaskListType
	Identity          string
	ActivityID        string
	ActivityType      *types.ActivityType
	ActivityInput     []byte
	ActivityHeader    *types.Header
	WorkflowType      *types.WorkflowType
	StartedEventID    int64
	ScheduledEventID  int64
}

func newTestParam(taskType int) *testParam {
	domainID := uuid.New()
	taskList := &types.TaskList{
		Name: strings.ReplaceAll(uuid.New(), "-", ""), // metric tags are sanitized
	}
	tlID := newTestTaskListID(domainID, taskList.Name, taskType)
	taskListType := types.TaskListType(taskType)
	return &testParam{
		DomainID: domainID,
		WorkflowExecution: &types.WorkflowExecution{
			WorkflowID: uuid.New(),
			RunID:      uuid.New(),
		},
		TaskList:     taskList,
		TaskListID:   tlID,
		TaskListType: &taskListType,
		Identity:     uuid.New(),
		ActivityID:   uuid.New(),
		ActivityType: &types.ActivityType{
			Name: uuid.New(),
		},
		ActivityInput:    []byte(uuid.New()),
		ActivityHeader:   &types.Header{Fields: map[string][]byte{"tracing": []byte("tracing data")}},
		WorkflowType:     &types.WorkflowType{Name: uuid.New()},
		ScheduledEventID: 1412,
	}
}

type addTaskRequest struct {
	TaskType                      int
	DomainUUID                    string
	Execution                     *types.WorkflowExecution
	TaskList                      *types.TaskList
	ScheduleID                    int64
	ScheduleToStartTimeoutSeconds int32
	Source                        *types.TaskSource
	ForwardedFrom                 string
	PartitionConfig               map[string]string
}

func addTask(engine *matchingEngineImpl, hCtx *handlerContext, request *addTaskRequest) (bool, error) {
	if request.TaskType == persistence.TaskListTypeActivity {
		return engine.AddActivityTask(hCtx, &types.AddActivityTaskRequest{
			SourceDomainUUID:              request.DomainUUID,
			DomainUUID:                    request.DomainUUID,
			Execution:                     request.Execution,
			TaskList:                      request.TaskList,
			ScheduleID:                    request.ScheduleID,
			ScheduleToStartTimeoutSeconds: &request.ScheduleToStartTimeoutSeconds,
			Source:                        request.Source,
			ForwardedFrom:                 request.ForwardedFrom,
			PartitionConfig:               request.PartitionConfig,
		})
	}
	return engine.AddDecisionTask(hCtx, &types.AddDecisionTaskRequest{
		DomainUUID:                    request.DomainUUID,
		Execution:                     request.Execution,
		TaskList:                      request.TaskList,
		ScheduleID:                    request.ScheduleID,
		ScheduleToStartTimeoutSeconds: &request.ScheduleToStartTimeoutSeconds,
		Source:                        request.Source,
		ForwardedFrom:                 request.ForwardedFrom,
		PartitionConfig:               request.PartitionConfig,
	})
}

type pollTaskRequest struct {
	TaskType         int
	DomainUUID       string
	PollerID         string
	TaskList         *types.TaskList
	Identity         string
	ForwardedFrom    string
	IsolationGroup   string
	TaskListMetadata *types.TaskListMetadata
	BinaryChecksum   string
}

type pollTaskResponse struct {
	TaskToken                       *common.TaskToken
	WorkflowExecution               *types.WorkflowExecution
	ActivityID                      string
	ActivityType                    *types.ActivityType
	Input                           []byte
	ScheduledTimestamp              *int64
	ScheduleToCloseTimeoutSeconds   *int32
	StartedTimestamp                *int64
	StartToCloseTimeoutSeconds      *int32
	HeartbeatTimeoutSeconds         *int32
	ScheduledTimestampOfThisAttempt *int64
	HeartbeatDetails                []byte
	WorkflowType                    *types.WorkflowType
	WorkflowDomain                  string
	Header                          *types.Header
	PreviousStartedEventID          *int64
	StartedEventID                  int64
	Attempt                         int64
	NextEventID                     int64
	BacklogCountHint                int64
	StickyExecutionEnabled          bool
	Query                           *types.WorkflowQuery
	DecisionInfo                    *types.TransientDecisionInfo
	WorkflowExecutionTaskList       *types.TaskList
	EventStoreVersion               int32
	BranchToken                     []byte
	Queries                         map[string]*types.WorkflowQuery
}

func pollTask(engine *matchingEngineImpl, hCtx *handlerContext, request *pollTaskRequest) (*pollTaskResponse, error) {
	if request.TaskType == persistence.TaskListTypeActivity {
		resp, err := engine.PollForActivityTask(hCtx, &types.MatchingPollForActivityTaskRequest{
			DomainUUID: request.DomainUUID,
			PollerID:   request.PollerID,
			PollRequest: &types.PollForActivityTaskRequest{
				TaskList:         request.TaskList,
				Identity:         request.Identity,
				TaskListMetadata: request.TaskListMetadata,
			},
			IsolationGroup: request.IsolationGroup,
			ForwardedFrom:  request.ForwardedFrom,
		})
		if err != nil {
			return nil, err
		}
		var token *common.TaskToken
		if len(resp.TaskToken) > 0 {
			token, err = engine.tokenSerializer.Deserialize(resp.TaskToken)
			if err != nil {
				return nil, err
			}
		}
		return &pollTaskResponse{
			TaskToken:                       token,
			WorkflowExecution:               resp.WorkflowExecution,
			ActivityID:                      resp.ActivityID,
			ActivityType:                    resp.ActivityType,
			Input:                           resp.Input,
			ScheduledTimestamp:              resp.ScheduledTimestamp,
			ScheduleToCloseTimeoutSeconds:   resp.ScheduleToCloseTimeoutSeconds,
			StartedTimestamp:                resp.StartedTimestamp,
			StartToCloseTimeoutSeconds:      resp.StartToCloseTimeoutSeconds,
			HeartbeatTimeoutSeconds:         resp.HeartbeatTimeoutSeconds,
			Attempt:                         int64(resp.Attempt),
			ScheduledTimestampOfThisAttempt: resp.ScheduledTimestampOfThisAttempt,
			HeartbeatDetails:                resp.HeartbeatDetails,
			WorkflowType:                    resp.WorkflowType,
			WorkflowDomain:                  resp.WorkflowDomain,
			Header:                          resp.Header,
		}, nil
	}
	resp, err := engine.PollForDecisionTask(hCtx, &types.MatchingPollForDecisionTaskRequest{
		DomainUUID: request.DomainUUID,
		PollerID:   request.PollerID,
		PollRequest: &types.PollForDecisionTaskRequest{
			TaskList:       request.TaskList,
			Identity:       request.Identity,
			BinaryChecksum: request.BinaryChecksum,
		},
		IsolationGroup: request.IsolationGroup,
		ForwardedFrom:  request.ForwardedFrom,
	})
	if err != nil {
		return nil, err
	}
	var token *common.TaskToken
	if len(resp.TaskToken) > 0 {
		token, err = engine.tokenSerializer.Deserialize(resp.TaskToken)
		if err != nil {
			return nil, err
		}
	}
	return &pollTaskResponse{
		TaskToken:                 token,
		WorkflowExecution:         resp.WorkflowExecution,
		WorkflowType:              resp.WorkflowType,
		PreviousStartedEventID:    resp.PreviousStartedEventID,
		StartedEventID:            resp.StartedEventID,
		Attempt:                   resp.Attempt,
		NextEventID:               resp.NextEventID,
		BacklogCountHint:          resp.BacklogCountHint,
		StickyExecutionEnabled:    resp.StickyExecutionEnabled,
		Query:                     resp.Query,
		DecisionInfo:              resp.DecisionInfo,
		WorkflowExecutionTaskList: resp.WorkflowExecutionTaskList,
		EventStoreVersion:         resp.EventStoreVersion,
		BranchToken:               resp.BranchToken,
		ScheduledTimestamp:        resp.ScheduledTimestamp,
		StartedTimestamp:          resp.StartedTimestamp,
		Queries:                   resp.Queries,
	}, nil
}

func isEmptyToken(token *common.TaskToken) bool {
	return token == nil || *token == common.TaskToken{}
}
