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

package tasklist

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
	"go.uber.org/goleak"
	"go.uber.org/mock/gomock"
	"golang.org/x/sync/errgroup"

	"github.com/uber/cadence/client/history"
	"github.com/uber/cadence/client/matching"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/isolationgroup"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/log/testlogger"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/partition"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/stats"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/matching/config"
	"github.com/uber/cadence/service/matching/poller"
)

type mockDeps struct {
	mockDomainCache    *cache.MockDomainCache
	mockTaskManager    *persistence.MockTaskManager
	mockPartitioner    *partition.MockPartitioner
	mockMatchingClient *matching.MockClient
	mockTimeSource     clock.MockedTimeSource
	dynamicClient      dynamicconfig.Client
}

var (
	testIsolationGroups = []string{"datacenterA", "datacenterB"}
)

func setupMocksForTaskListManager(t *testing.T, taskListID *Identifier, taskListKind types.TaskListKind) (*taskListManagerImpl, *mockDeps) {
	ctrl := gomock.NewController(t)
	dynamicClient := dynamicconfig.NewInMemoryClient()
	logger := testlogger.New(t)
	metricsClient := metrics.NewNoopMetricsClient()
	clusterMetadata := cluster.GetTestClusterMetadata(true)
	deps := &mockDeps{
		mockDomainCache:    cache.NewMockDomainCache(ctrl),
		mockTaskManager:    persistence.NewMockTaskManager(ctrl),
		mockPartitioner:    partition.NewMockPartitioner(ctrl),
		mockMatchingClient: matching.NewMockClient(ctrl),
		mockTimeSource:     clock.NewMockedTimeSource(),
		dynamicClient:      dynamicClient,
	}
	deps.mockDomainCache.EXPECT().GetDomainName(gomock.Any()).Return("domainName", nil).Times(1)
	config := config.NewConfig(dynamicconfig.NewCollection(dynamicClient, logger), "hostname", getIsolationgroupsHelper)
	mockHistoryService := history.NewMockClient(ctrl)

	tlm, err := NewManager(
		deps.mockDomainCache,
		logger,
		metricsClient,
		deps.mockTaskManager,
		clusterMetadata,
		deps.mockPartitioner,
		deps.mockMatchingClient,
		func(Manager) {},
		taskListID,
		&taskListKind,
		config,
		deps.mockTimeSource,
		deps.mockTimeSource.Now(),
		mockHistoryService,
	)
	require.NoError(t, err)
	return tlm.(*taskListManagerImpl), deps
}

func defaultTestConfig() *config.Config {
	config := config.NewConfig(dynamicconfig.NewNopCollection(), "some random hostname", getIsolationgroupsHelper)
	config.LongPollExpirationInterval = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(100 * time.Millisecond)
	config.MaxTaskDeleteBatchSize = dynamicconfig.GetIntPropertyFilteredByTaskListInfo(1)
	config.AllIsolationGroups = getIsolationgroupsHelper
	config.GetTasksBatchSize = dynamicconfig.GetIntPropertyFilteredByTaskListInfo(10)
	config.AsyncTaskDispatchTimeout = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(10 * time.Millisecond)
	config.LocalTaskWaitTime = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(time.Millisecond)
	return config
}

func TestDeliverBufferTasks(t *testing.T) {
	controller := gomock.NewController(t)
	logger := testlogger.New(t)

	tests := []func(tlm *taskListManagerImpl){
		func(tlm *taskListManagerImpl) { tlm.taskReader.cancelFunc() },
		func(tlm *taskListManagerImpl) {
			rps := 0.1
			tlm.matcher.UpdateRatelimit(&rps)
			tlm.taskReader.taskBuffers[defaultTaskBufferIsolationGroup] <- &persistence.TaskInfo{}
			err := tlm.matcher.(*taskMatcherImpl).ratelimit(context.Background()) // consume the token
			assert.NoError(t, err)
			tlm.taskReader.cancelFunc()
		},
	}
	for _, test := range tests {
		tlm := createTestTaskListManager(t, logger, controller)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			tlm.taskReader.dispatchBufferedTasks(defaultTaskBufferIsolationGroup)
		}()
		test(tlm)
		// dispatchBufferedTasks should stop after invocation of the test function
		wg.Wait()
	}
}

func TestTaskListString(t *testing.T) {
	controller := gomock.NewController(t)
	logger := testlogger.New(t)
	tlm := createTestTaskListManager(t, logger, controller)
	got := tlm.String()
	want := "Activity task list tl\nRangeID=0\nTaskIDBlock={start:-99999 end:0}\nAckLevel=-1\nMaxReadLevel=-1\n"
	assert.Equal(t, want, got)
}

func TestDeliverBufferTasks_NoPollers(t *testing.T) {
	controller := gomock.NewController(t)
	logger := testlogger.New(t)

	tlm := createTestTaskListManager(t, logger, controller)
	tlm.taskReader.taskBuffers[defaultTaskBufferIsolationGroup] <- &persistence.TaskInfo{}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		tlm.taskReader.dispatchBufferedTasks("")
		wg.Done()
	}()
	time.Sleep(100 * time.Millisecond) // let go routine run first and block on tasksForPoll
	tlm.taskReader.cancelFunc()
	wg.Wait()
}

func TestReadLevelForAllExpiredTasksInBatch(t *testing.T) {
	controller := gomock.NewController(t)
	logger := testlogger.New(t)

	tlm := createTestTaskListManager(t, logger, controller)
	tlm.db.rangeID = int64(1)
	tlm.taskAckManager.SetAckLevel(0)
	tlm.taskAckManager.SetReadLevel(0)
	require.Equal(t, int64(0), tlm.taskAckManager.GetAckLevel())
	require.Equal(t, int64(0), tlm.taskAckManager.GetReadLevel())

	// Add all expired tasks
	tasks := []*persistence.TaskInfo{
		{
			TaskID:      11,
			Expiry:      time.Now().Add(-time.Minute),
			CreatedTime: time.Now().Add(-time.Hour),
		},
		{
			TaskID:      12,
			Expiry:      time.Now().Add(-time.Minute),
			CreatedTime: time.Now().Add(-time.Hour),
		},
	}

	require.True(t, tlm.taskReader.addTasksToBuffer(tasks))
	require.Equal(t, int64(0), tlm.taskAckManager.GetAckLevel())
	require.Equal(t, int64(12), tlm.taskAckManager.GetReadLevel())

	// Now add a mix of valid and expired tasks
	require.True(t, tlm.taskReader.addTasksToBuffer([]*persistence.TaskInfo{
		{
			TaskID:      13,
			Expiry:      time.Now().Add(-time.Minute),
			CreatedTime: time.Now().Add(-time.Hour),
		},
		{
			TaskID:      14,
			Expiry:      time.Now().Add(time.Hour),
			CreatedTime: time.Now().Add(time.Minute),
		},
	}))
	require.Equal(t, int64(0), tlm.taskAckManager.GetAckLevel())
	require.Equal(t, int64(14), tlm.taskAckManager.GetReadLevel())
}

func createTestTaskListManager(t *testing.T, logger log.Logger, controller *gomock.Controller) *taskListManagerImpl {
	return createTestTaskListManagerWithConfig(t, logger, controller, defaultTestConfig(), clock.NewMockedTimeSource())
}

func createTestTaskListManagerWithConfig(t *testing.T, logger log.Logger, controller *gomock.Controller, cfg *config.Config, timeSource clock.TimeSource) *taskListManagerImpl {
	tm := NewTestTaskManager(t, logger, timeSource)
	mockPartitioner := partition.NewMockPartitioner(controller)
	mockPartitioner.EXPECT().GetIsolationGroupByDomainID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
	mockDomainCache := cache.NewMockDomainCache(controller)
	mockDomainCache.EXPECT().GetDomainByID(gomock.Any()).Return(cache.CreateDomainCacheEntry("domainName"), nil).AnyTimes()
	mockDomainCache.EXPECT().GetDomainName(gomock.Any()).Return("domainName", nil).AnyTimes()
	mockHistoryService := history.NewMockClient(controller)
	tl := "tl"
	dID := "domain"
	tlID, err := NewIdentifier(dID, tl, persistence.TaskListTypeActivity)
	if err != nil {
		panic(err)
	}
	tlKind := types.TaskListKindNormal
	tlMgr, err := NewManager(mockDomainCache, logger, metrics.NewClient(tally.NoopScope, metrics.Matching), tm, cluster.GetTestClusterMetadata(true), mockPartitioner, nil, func(Manager) {}, tlID, &tlKind, cfg, timeSource, timeSource.Now(), mockHistoryService)
	if err != nil {
		logger.Fatal("error when createTestTaskListManager", tag.Error(err))
	}
	return tlMgr.(*taskListManagerImpl)
}

func TestDescribeTaskList(t *testing.T) {
	// Magic values hardcoded in matching/config.go. Not much of a config :(
	defaultRps := 100000.0
	defaultRangeSize := 100000
	startedID := int64(1)
	firstIDBlock := &types.TaskIDBlock{
		StartID: startedID,
		EndID:   int64(defaultRangeSize),
	}
	cases := []struct {
		name           string
		includeStatus  bool
		pollers        map[string]poller.Info
		allowance      func(ctrl *gomock.Controller, impl *taskListManagerImpl)
		expectedStatus *types.TaskListStatus
		expectedConfig *types.TaskListPartitionConfig
	}{
		{
			name: "no status, pollers, or config",
		},
		{
			name: "no status, with config",
			allowance: func(_ *gomock.Controller, impl *taskListManagerImpl) {
				err := impl.RefreshTaskListPartitionConfig(context.Background(), &types.TaskListPartitionConfig{
					Version:            1,
					NumReadPartitions:  3,
					NumWritePartitions: 2,
				})
				require.NoError(t, err)
			},
			expectedConfig: &types.TaskListPartitionConfig{
				Version:            1,
				NumReadPartitions:  3,
				NumWritePartitions: 2,
			},
		},
		{
			name: "no status, with pollers",
			pollers: map[string]poller.Info{
				"pollerID": {
					RatePerSecond:  1.0,
					IsolationGroup: "a",
				},
			},
		},
		{
			name:          "with status",
			includeStatus: true,
			expectedStatus: &types.TaskListStatus{
				RatePerSecond: defaultRps,
				TaskIDBlock:   firstIDBlock,
				IsolationGroupMetrics: map[string]*types.IsolationGroupMetrics{
					"datacenterA": {},
					"datacenterB": {},
				},
			},
		},
		{
			name:          "with status, tasks completed",
			includeStatus: true,
			allowance: func(_ *gomock.Controller, tlm *taskListManagerImpl) {
				tlm.taskAckManager.SetAckLevel(5)
				tlm.taskAckManager.SetReadLevel(10)
			},
			expectedStatus: &types.TaskListStatus{
				BacklogCountHint: 0,
				ReadLevel:        10,
				AckLevel:         5,
				RatePerSecond:    defaultRps,
				TaskIDBlock:      firstIDBlock,
				IsolationGroupMetrics: map[string]*types.IsolationGroupMetrics{
					"datacenterA": {},
					"datacenterB": {},
				},
			},
		},
		{
			name:          "with status, pollers and metrics",
			includeStatus: true,
			pollers: map[string]poller.Info{
				"a-1": {
					RatePerSecond:  1.0,
					IsolationGroup: "datacenterA",
				},
			},
			allowance: func(ctrl *gomock.Controller, impl *taskListManagerImpl) {
				mockQPS := stats.NewMockQPSTrackerGroup(ctrl)
				mockQPS.EXPECT().GroupQPS("datacenterA").Return(float64(75.0))
				mockQPS.EXPECT().GroupQPS("datacenterB").Return(float64(25.0))
				mockQPS.EXPECT().QPS().Return(float64(100.0))
				impl.qpsTracker = mockQPS
				impl.matcher.UpdateRatelimit(common.Float64Ptr(1.0))
			},
			expectedStatus: &types.TaskListStatus{
				RatePerSecond:     1.0, // From poller
				TaskIDBlock:       firstIDBlock,
				NewTasksPerSecond: 100,
				IsolationGroupMetrics: map[string]*types.IsolationGroupMetrics{
					"datacenterA": {
						PollerCount:       1,
						NewTasksPerSecond: 75.0,
					},
					"datacenterB": {
						PollerCount:       0,
						NewTasksPerSecond: 25.0,
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			logger := testlogger.New(t)
			tlm := createTestTaskListManager(t, logger, controller)
			tlm.db.rangeID = int64(1)
			tlm.taskAckManager.SetAckLevel(0)
			tlm.startWG.Done()

			expectedPollers := make([]*types.PollerInfo, 0, len(tc.pollers))
			for id, info := range tc.pollers {
				tlm.pollerHistory.UpdatePollerInfo(poller.Identity(id), info)
				expectedPollers = append(expectedPollers, &types.PollerInfo{
					LastAccessTime: common.Int64Ptr(tlm.timeSource.Now().UnixNano()),
					Identity:       id,
					RatePerSecond:  info.RatePerSecond,
				})
			}
			if tc.allowance != nil {
				tc.allowance(controller, tlm)
			}
			result := tlm.DescribeTaskList(tc.includeStatus)
			assert.Equal(t, tc.expectedStatus, result.TaskListStatus)
			assert.Equal(t, tc.expectedConfig, result.PartitionConfig)
			assert.ElementsMatch(t, expectedPollers, result.Pollers)
		})
	}
}

func TestCheckIdleTaskList(t *testing.T) {
	defer goleak.VerifyNone(t)
	cfg := config.NewConfig(dynamicconfig.NewNopCollection(), "some random hostname", getIsolationgroupsHelper)
	cfg.IdleTasklistCheckInterval = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(10 * time.Millisecond)

	t.Run("Idle task-list", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		tlm := createTestTaskListManagerWithConfig(t, testlogger.New(t), ctrl, cfg, clock.NewRealTimeSource())
		require.NoError(t, tlm.Start())

		require.EqualValues(t, 0, atomic.LoadInt32(&tlm.stopped), "idle check interval had not passed yet")
		time.Sleep(20 * time.Millisecond)
		require.EqualValues(t, 1, atomic.LoadInt32(&tlm.stopped), "idle check interval should have pass")
	})

	t.Run("Active poll-er", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		tlm := createTestTaskListManagerWithConfig(t, testlogger.New(t), ctrl, cfg, clock.NewRealTimeSource())
		require.NoError(t, tlm.Start())

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, _ = tlm.GetTask(ctx, nil)
		cancel()

		// task list manager should have been stopped,
		// but GetTask extends auto-stop until the next check-idle-task-list-interval
		time.Sleep(6 * time.Millisecond)
		require.EqualValues(t, 0, atomic.LoadInt32(&tlm.stopped))

		time.Sleep(20 * time.Millisecond)
		require.EqualValues(t, 1, atomic.LoadInt32(&tlm.stopped), "idle check interval should have pass")
	})

	t.Run("Active adding task", func(t *testing.T) {
		domainID := uuid.New()
		workflowID := uuid.New()
		runID := uuid.New()

		addTaskParam := AddTaskParams{
			TaskInfo: &persistence.TaskInfo{
				DomainID:                      domainID,
				WorkflowID:                    workflowID,
				RunID:                         runID,
				ScheduleID:                    2,
				ScheduleToStartTimeoutSeconds: 5,
				CreatedTime:                   time.Now(),
			},
		}

		ctrl := gomock.NewController(t)
		tlm := createTestTaskListManagerWithConfig(t, testlogger.New(t), ctrl, cfg, clock.NewRealTimeSource())
		require.NoError(t, tlm.Start())

		time.Sleep(8 * time.Millisecond)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err := tlm.AddTask(ctx, addTaskParam)
		require.NoError(t, err)
		cancel()

		// task list manager should have been stopped,
		// but AddTask extends auto-stop until the next check-idle-task-list-interval
		time.Sleep(6 * time.Millisecond)
		require.EqualValues(t, 0, atomic.LoadInt32(&tlm.stopped))

		time.Sleep(20 * time.Millisecond)
		require.EqualValues(t, 1, atomic.LoadInt32(&tlm.stopped), "idle check interval should have pass")
	})
}

func TestAddTaskStandby(t *testing.T) {
	controller := gomock.NewController(t)
	logger := testlogger.New(t)

	cfg := config.NewConfig(dynamicconfig.NewNopCollection(), "some random hostname", getIsolationgroupsHelper)
	cfg.IdleTasklistCheckInterval = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(10 * time.Millisecond)

	tlm := createTestTaskListManagerWithConfig(t, logger, controller, cfg, clock.NewMockedTimeSource())
	require.NoError(t, tlm.Start())

	// stop taskWriter so that we can check if there's any call to it
	// otherwise the task persist process is async and hard to test
	tlm.taskWriter.Stop()

	domainID := uuid.New()
	workflowID := "some random workflowID"
	runID := "some random runID"

	addTaskParam := AddTaskParams{
		TaskInfo: &persistence.TaskInfo{
			DomainID:                      domainID,
			WorkflowID:                    workflowID,
			RunID:                         runID,
			ScheduleID:                    2,
			ScheduleToStartTimeoutSeconds: 5,
			CreatedTime:                   time.Now(),
		},
	}

	testStandbyDomainEntry := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{ID: domainID, Name: "some random domain name"},
		&persistence.DomainConfig{Retention: 1},
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestAlternativeClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
				{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		1234,
	)
	mockDomainCache := tlm.domainCache.(*cache.MockDomainCache)
	mockDomainCache.EXPECT().GetDomainByID(domainID).Return(testStandbyDomainEntry, nil).AnyTimes()

	syncMatch, err := tlm.AddTask(context.Background(), addTaskParam)
	require.Equal(t, errShutdown, err) // task writer was stopped above
	require.False(t, syncMatch)

	addTaskParam.ForwardedFrom = "from child partition"
	syncMatch, err = tlm.AddTask(context.Background(), addTaskParam)
	require.Error(t, err) // should not persist the task
	require.False(t, syncMatch)
}

func TestGetPollerIsolationGroup(t *testing.T) {
	controller := gomock.NewController(t)
	logger := testlogger.New(t)

	config := defaultTestConfig()
	mockClock := clock.NewMockedTimeSource()
	config.LongPollExpirationInterval = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(30 * time.Second)
	config.EnableTasklistIsolation = dynamicconfig.GetBoolPropertyFnFilteredByDomainID(true)
	tlm := createTestTaskListManagerWithConfig(t, logger, controller, config, mockClock)

	bgCtx := ContextWithPollerID(context.Background(), "poller0")
	bgCtx = ContextWithIdentity(bgCtx, "id0")
	bgCtx = ContextWithIsolationGroup(bgCtx, getIsolationgroupsHelper()[0])
	ctx, cancel := context.WithTimeout(bgCtx, time.Millisecond)
	_, err := tlm.GetTask(ctx, nil)
	cancel()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrNoTasks.Error())

	// we should get isolation groups that showed up within last isolationDuration
	groups := tlm.getPollerIsolationGroups()
	assert.Equal(t, 1, len(groups))
	assert.Equal(t, getIsolationgroupsHelper()[0], groups[0])

	// after isolation duration, the poller from that isolation group are cleared from the poller history
	mockClock.Advance((10 * time.Second) + time.Nanosecond)
	groups = tlm.getPollerIsolationGroups()
	assert.Equal(t, 0, len(groups))

	// we should get isolation groups of outstanding pollers
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		tlm.GetTask(ctx, nil)
		cancel()
	}()
	awaitCondition(func() bool {
		return len(tlm.getPollerIsolationGroups()) > 0
	}, time.Second)
	groups = tlm.getPollerIsolationGroups()
	cancel()
	wg.Wait()
	assert.Equal(t, 1, len(groups))
	assert.Equal(t, getIsolationgroupsHelper()[0], groups[0])
}

// return a client side tasklist throttle error from the rate limiter.
// The expected behaviour is to retry
func TestRateLimitErrorsFromTasklistDispatch(t *testing.T) {
	controller := gomock.NewController(t)
	logger := testlogger.New(t)

	config := defaultTestConfig()
	config.EnableTasklistIsolation = func(domain string) bool { return true }
	tlm := createTestTaskListManagerWithConfig(t, logger, controller, config, clock.NewMockedTimeSource())

	tlm.taskReader.dispatchTask = func(ctx context.Context, task *InternalTask) error {
		return ErrTasklistThrottled
	}
	tlm.taskReader.getIsolationGroupForTask = func(ctx context.Context, info *persistence.TaskInfo) (string, time.Duration, error) {
		return "datacenterA", -1, nil
	}

	breakDispatcher, breakRetryLoop := tlm.taskReader.dispatchSingleTaskFromBuffer(&persistence.TaskInfo{})
	assert.False(t, breakDispatcher)
	assert.False(t, breakRetryLoop)
}

// This is a bit of a strange unit-test as it's
// ensuring that invalid behaviour is handled defensively.
// It *should never be the case* that the isolation group tries to
// dispatch to a buffer that isn't there, however, if it does, we want to just
// log this, emit a metric and fallback to the default isolation group.
func TestMisconfiguredZoneDoesNotBlock(t *testing.T) {
	controller := gomock.NewController(t)
	logger := testlogger.New(t)

	config := defaultTestConfig()
	config.EnableTasklistIsolation = func(domain string) bool { return true }
	tlm := createTestTaskListManagerWithConfig(t, logger, controller, config, clock.NewMockedTimeSource())
	invalidIsolationGroup := "invalid"
	dispatched := make(map[string]int)

	// record dispatched isolation group
	tlm.taskReader.dispatchTask = func(ctx context.Context, task *InternalTask) error {
		dispatched[task.isolationGroup] = dispatched[task.isolationGroup] + 1
		return nil
	}
	tlm.taskReader.getIsolationGroupForTask = func(ctx context.Context, info *persistence.TaskInfo) (string, time.Duration, error) {
		return invalidIsolationGroup, -1, nil
	}

	maxBufferSize := config.GetTasksBatchSize("", "", 0) - 1

	for i := 0; i < maxBufferSize; i++ {
		breakDispatcher, breakRetryLoop := tlm.taskReader.dispatchSingleTaskFromBuffer(&persistence.TaskInfo{})
		assert.False(t, breakDispatcher, "dispatch isn't shutting down")
		assert.True(t, breakRetryLoop, "should be able to successfully dispatch all these tasks to the default isolation group")
	}
	// We should see them all being redirected
	assert.Equal(t, dispatched[""], maxBufferSize)
	// we should see none in the returned isolation group
	assert.Equal(t, dispatched[invalidIsolationGroup], 0)

	// ok, and here we try and ensure that this *does not block
	// and instead complains and live-retries
	breakDispatcher, breakRetryLoop := tlm.taskReader.dispatchSingleTaskFromBuffer(&persistence.TaskInfo{})
	assert.False(t, breakDispatcher, "dispatch isn't shutting down")
	assert.True(t, breakRetryLoop, "task should be dispatched to default channel")
}

func TestGetIsolationGroupForTask(t *testing.T) {
	defaultAvailableIsolationGroups := []string{
		"a", "b", "c",
	}
	taskIsolationPollerWindow := time.Second * 10
	testCases := []struct {
		name                     string
		taskIsolationGroup       string
		taskIsolationDuration    time.Duration
		taskLatency              time.Duration
		availableIsolationGroups []string
		recentPollers            []string
		expectedGroup            string
		expectedDuration         time.Duration
		disableTaskIsolation     bool
	}{
		{
			name:                     "success - recent poller allows group",
			taskIsolationGroup:       "a",
			availableIsolationGroups: defaultAvailableIsolationGroups,
			expectedGroup:            "a",
			expectedDuration:         0,
			recentPollers:            []string{"a"},
		},
		{
			name:                     "success - with isolation duration",
			taskIsolationGroup:       "b",
			taskIsolationDuration:    time.Second,
			availableIsolationGroups: defaultAvailableIsolationGroups,
			expectedGroup:            "b",
			expectedDuration:         time.Second,
			recentPollers:            []string{"b"},
		},
		{
			name:                     "success - low task latency",
			taskIsolationGroup:       "a",
			taskIsolationDuration:    time.Second,
			taskLatency:              time.Millisecond * 300,
			availableIsolationGroups: defaultAvailableIsolationGroups,
			expectedGroup:            "a",
			expectedDuration:         time.Second - (time.Millisecond * 300),
			recentPollers:            []string{"a"},
		},
		{
			name:                     "leak - no recent pollers",
			taskIsolationGroup:       "a",
			availableIsolationGroups: defaultAvailableIsolationGroups,
			expectedGroup:            "",
			expectedDuration:         0,
			recentPollers:            nil,
		},
		{
			name:                     "leak - no matching recent poller",
			taskIsolationGroup:       "a",
			taskIsolationDuration:    time.Second,
			availableIsolationGroups: defaultAvailableIsolationGroups,
			expectedGroup:            "",
			expectedDuration:         0,
			recentPollers:            []string{"b"},
		},
		{
			name:                     "leak - task latency",
			taskIsolationGroup:       "a",
			taskIsolationDuration:    time.Second,
			taskLatency:              time.Second,
			availableIsolationGroups: defaultAvailableIsolationGroups,
			expectedGroup:            "",
			expectedDuration:         0,
		},
		{
			name:                     "leak - task latency close to taskIsolationDuration",
			taskIsolationGroup:       "a",
			taskIsolationDuration:    time.Second,
			taskLatency:              time.Second - minimumIsolationDuration,
			availableIsolationGroups: defaultAvailableIsolationGroups,
			expectedGroup:            "",
			expectedDuration:         0,
		},
		{
			name:                  "leak - partitioner error",
			taskIsolationGroup:    "a",
			taskIsolationDuration: time.Second,
			// No isolation groups causes an error
			// availableIsolationGroups: defaultAvailableIsolationGroups,
			expectedGroup:    "",
			expectedDuration: 0,
		},
		{
			name:                     "leak - task isolation disabled",
			taskIsolationGroup:       "a",
			taskIsolationDuration:    time.Second,
			availableIsolationGroups: defaultAvailableIsolationGroups,
			expectedGroup:            "",
			expectedDuration:         0,
			disableTaskIsolation:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			logger := testlogger.New(t)

			config := defaultTestConfig()
			config.EnableTasklistIsolation = func(domain string) bool { return !tc.disableTaskIsolation }
			config.TaskIsolationDuration = func(domain string, taskList string, taskType int) time.Duration {
				return tc.taskIsolationDuration
			}
			config.TaskIsolationPollerWindow = func(domain string, taskList string, taskType int) time.Duration {
				return taskIsolationPollerWindow
			}
			config.AllIsolationGroups = func() []string {
				return tc.availableIsolationGroups
			}
			mockClock := clock.NewMockedTimeSource()
			tlm := createTestTaskListManagerWithConfig(t, logger, controller, config, mockClock)

			mockIsolationGroupState := isolationgroup.NewMockState(controller)
			mockIsolationGroupState.EXPECT().IsDrained(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
			mockIsolationGroupState.EXPECT().IsDrainedByDomainID(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
			mockIsolationGroupState.EXPECT().IsolationGroupsByDomainID(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, domainId string) (types.IsolationGroupConfiguration, error) {
				// Report all available isolation groups as healthy
				isolationGroupStates := make(types.IsolationGroupConfiguration, len(tc.availableIsolationGroups))
				for _, availableGroup := range tc.availableIsolationGroups {
					isolationGroupStates[availableGroup] = types.IsolationGroupPartition{
						Name:  availableGroup,
						State: types.IsolationGroupStateHealthy,
					}
				}
				return isolationGroupStates, nil
			}).AnyTimes()
			partitioner := partition.NewDefaultPartitioner(log.NewNoop(), mockIsolationGroupState)
			tlm.partitioner = partitioner

			for _, pollerGroup := range tc.recentPollers {
				tlm.pollerHistory.UpdatePollerInfo(poller.Identity(pollerGroup), poller.Info{IsolationGroup: pollerGroup})
			}

			taskInfo := &persistence.TaskInfo{
				DomainID:                      "domainId",
				RunID:                         "run1",
				WorkflowID:                    "workflow1",
				ScheduleID:                    5,
				ScheduleToStartTimeoutSeconds: 1,
				PartitionConfig: map[string]string{
					partition.IsolationGroupKey: tc.taskIsolationGroup,
					partition.WorkflowIDKey:     "workflow1",
				},
				CreatedTime: mockClock.Now().Add(-1 * tc.taskLatency),
			}

			actual, actualDuration, err := tlm.getIsolationGroupForTask(context.Background(), taskInfo)

			assert.Equal(t, tc.expectedGroup, actual)
			assert.Equal(t, tc.expectedDuration, actualDuration)
			// There are no longer any error cases
			assert.Nil(t, err)
		})
	}
}

func TestTaskWriterShutdown(t *testing.T) {
	controller := gomock.NewController(t)
	logger := testlogger.New(t)
	tlm := createTestTaskListManager(t, logger, controller)
	err := tlm.Start()
	assert.NoError(t, err)

	// stop the task writer explicitly
	tlm.taskWriter.Stop()

	// now attempt to add a task
	addParams := AddTaskParams{
		TaskInfo: &persistence.TaskInfo{
			DomainID:                      "domainId",
			RunID:                         "run1",
			WorkflowID:                    "workflow1",
			ScheduleID:                    5,
			ScheduleToStartTimeoutSeconds: 1,
		},
	}
	_, err = tlm.AddTask(context.Background(), addParams)
	assert.Error(t, err)

	// test race
	tlm.taskWriter.stopped = 0
	_, err = tlm.AddTask(context.Background(), addParams)
	assert.Error(t, err)
	tlm.taskWriter.stopped = 1 // reset it back to old value
	tlm.Stop()
}

func TestTaskListManagerGetTaskBatch(t *testing.T) {
	const taskCount = 1200
	const rangeSize = 10
	controller := gomock.NewController(t)
	mockPartitioner := partition.NewMockPartitioner(controller)
	mockPartitioner.EXPECT().GetIsolationGroupByDomainID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
	mockDomainCache := cache.NewMockDomainCache(controller)
	mockDomainCache.EXPECT().GetDomainByID(gomock.Any()).Return(cache.CreateDomainCacheEntry("domainName"), nil).AnyTimes()
	mockDomainCache.EXPECT().GetDomainName(gomock.Any()).Return("domainName", nil).AnyTimes()
	mockHistoryService := history.NewMockClient(controller)
	logger := testlogger.New(t)
	timeSource := clock.NewRealTimeSource()
	tm := NewTestTaskManager(t, logger, timeSource)
	taskListID := NewTestTaskListID(t, "domainId", "tl", 0)
	cfg := defaultTestConfig()
	cfg.RangeSize = rangeSize
	tlMgr, err := NewManager(
		mockDomainCache,
		logger,
		metrics.NewClient(tally.NoopScope, metrics.Matching),
		tm,
		cluster.GetTestClusterMetadata(true),
		mockPartitioner,
		nil,
		func(Manager) {},
		taskListID,
		types.TaskListKindNormal.Ptr(),
		cfg,
		timeSource,
		timeSource.Now(),
		mockHistoryService,
	)
	assert.NoError(t, err)
	tlm := tlMgr.(*taskListManagerImpl)
	err = tlm.Start()
	assert.NoError(t, err)

	// add taskCount tasks
	for i := int64(0); i < taskCount; i++ {
		scheduleID := i * 3
		addParams := AddTaskParams{
			TaskInfo: &persistence.TaskInfo{
				DomainID:                      "domainId",
				RunID:                         "run1",
				WorkflowID:                    "workflow1",
				ScheduleID:                    scheduleID,
				ScheduleToStartTimeoutSeconds: 100,
			},
		}
		_, err = tlm.AddTask(context.Background(), addParams)
		assert.NoError(t, err)
	}
	assert.Equal(t, taskCount, tm.GetTaskCount(taskListID))

	// wait until all tasks are read by the task pump and enqeued into the in-memory buffer
	// at the end of this step, ackManager readLevel will also be equal to the buffer size
	expectedBufSize := common.MinInt(cap(tlm.taskReader.taskBuffers[defaultTaskBufferIsolationGroup]), taskCount)
	assert.True(t, awaitCondition(func() bool {
		return len(tlm.taskReader.taskBuffers[defaultTaskBufferIsolationGroup]) == expectedBufSize
	}, 10*time.Second))

	// stop all goroutines that read / write tasks in the background
	// remainder of this test works with the in-memory buffer
	tlm.Stop()

	// SetReadLevel should NEVER be called without updating ackManager.outstandingTasks
	// This is only for unit test purpose
	tlm.taskAckManager.SetReadLevel(tlm.taskWriter.GetMaxReadLevel())
	tasks, readLevel, isReadBatchDone, err := tlm.taskReader.getTaskBatch(tlm.taskAckManager.GetReadLevel(), tlm.taskWriter.GetMaxReadLevel())
	assert.NoError(t, err)
	assert.Equal(t, 0, len(tasks))
	assert.Equal(t, readLevel, tlm.taskWriter.GetMaxReadLevel())
	assert.True(t, isReadBatchDone)

	tlm.taskAckManager.SetReadLevel(0)
	tasks, readLevel, isReadBatchDone, err = tlm.taskReader.getTaskBatch(tlm.taskAckManager.GetReadLevel(), tlm.taskWriter.GetMaxReadLevel())
	assert.NoError(t, err)
	assert.Equal(t, rangeSize, len(tasks))
	assert.Equal(t, rangeSize, int(readLevel))
	assert.True(t, isReadBatchDone)

	// reset the ackManager readLevel to the buffer size and consume
	// the in-memory tasks by calling Poll API - assert ackMgr state
	// at the end
	tlm.taskAckManager.SetReadLevel(int64(expectedBufSize))

	// complete rangeSize events
	tlMgr, err = NewManager(
		mockDomainCache,
		logger,
		metrics.NewClient(tally.NoopScope, metrics.Matching),
		tm,
		cluster.GetTestClusterMetadata(true),
		mockPartitioner,
		nil,
		func(Manager) {},
		taskListID,
		types.TaskListKindNormal.Ptr(),
		cfg,
		timeSource,
		timeSource.Now(),
		mockHistoryService,
	)
	assert.NoError(t, err)
	tlm = tlMgr.(*taskListManagerImpl)
	err = tlm.Start()
	assert.NoError(t, err)
	for i := int64(0); i < rangeSize; i++ {
		task, err := tlm.GetTask(context.Background(), nil)
		if errors.Is(err, ErrNoTasks) {
			continue
		}
		assert.NotNil(t, task.AutoConfigHint)
		assert.NoError(t, err)
		assert.NotNil(t, task)
		task.Finish(nil)
	}
	assert.Equal(t, taskCount-rangeSize, tm.GetTaskCount(taskListID))
	tasks, _, isReadBatchDone, err = tlm.taskReader.getTaskBatch(tlm.taskAckManager.GetReadLevel(), tlm.taskWriter.GetMaxReadLevel())
	assert.NoError(t, err)
	assert.True(t, 0 < len(tasks) && len(tasks) <= rangeSize)
	assert.True(t, isReadBatchDone)
	tlm.Stop()
}

func TestTaskListReaderPumpAdvancesAckLevelAfterEmptyReads(t *testing.T) {
	const taskCount = 5
	const rangeSize = 10
	const nLeaseRenewals = 15

	controller := gomock.NewController(t)
	mockPartitioner := partition.NewMockPartitioner(controller)
	mockPartitioner.EXPECT().GetIsolationGroupByDomainID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
	mockDomainCache := cache.NewMockDomainCache(controller)
	mockDomainCache.EXPECT().GetDomainByID(gomock.Any()).Return(cache.CreateDomainCacheEntry("domainName"), nil).AnyTimes()
	mockDomainCache.EXPECT().GetDomainName(gomock.Any()).Return("domainName", nil).AnyTimes()
	mockHistoryService := history.NewMockClient(controller)

	logger := testlogger.New(t)
	timeSource := clock.NewRealTimeSource()
	tm := NewTestTaskManager(t, logger, timeSource)
	taskListID := NewTestTaskListID(t, "domainId", "tl", 0)
	cfg := defaultTestConfig()
	cfg.RangeSize = rangeSize

	tlMgr, err := NewManager(
		mockDomainCache,
		logger,
		metrics.NewClient(tally.NoopScope, metrics.Matching),
		tm,
		cluster.GetTestClusterMetadata(true),
		mockPartitioner,
		nil,
		func(Manager) {},
		taskListID,
		types.TaskListKindNormal.Ptr(),
		cfg,
		timeSource,
		timeSource.Now(),
		mockHistoryService,
	)
	require.NoError(t, err)
	tlm := tlMgr.(*taskListManagerImpl)

	// simulate lease renewal multiple times without writing any tasks
	for i := 0; i < nLeaseRenewals; i++ {
		tlm.taskWriter.renewLeaseWithRetry()
	}

	err = tlm.Start() // this call will also renew lease
	require.NoError(t, err)
	defer tlm.Stop()

	// we expect AckLevel to advance and skip all the previously leased ranges
	expectedAckLevel := int64(rangeSize) * nLeaseRenewals

	// wait until task pump will read batches of empty ranges
	assert.True(t, awaitCondition(func() bool {
		return tlm.taskAckManager.GetAckLevel() == expectedAckLevel
	}, 10*time.Second))

	assert.Equal(
		t,
		expectedAckLevel,
		tlm.taskAckManager.GetAckLevel(),
		"we should ack ranges of all the previously acquired leases",
	)

	assert.Equal(
		t,
		tlm.taskWriter.GetMaxReadLevel(),
		tlm.taskAckManager.GetAckLevel(),
		"we should have been acked everything possible",
	)

	maxReadLevelBeforeAddingTasks := tlm.taskWriter.GetMaxReadLevel()

	// verify new task writes go beyond the MaxReadLevel/AckLevel
	for i := int64(0); i < taskCount; i++ {
		addParams := AddTaskParams{
			TaskInfo: &persistence.TaskInfo{
				DomainID:   "domainId",
				RunID:      "run1",
				WorkflowID: "workflow1",
				ScheduleID: i,
			},
		}
		_, err = tlm.AddTask(context.Background(), addParams)
		require.NoError(t, err)
		assert.Equal(t, maxReadLevelBeforeAddingTasks+i+1, tlm.taskWriter.GetMaxReadLevel())
	}
}

func TestTaskListManagerGetTaskBatch_ReadBatchDone(t *testing.T) {
	const rangeSize = 10
	const maxReadLevel = int64(120)
	config := defaultTestConfig()
	config.RangeSize = rangeSize
	controller := gomock.NewController(t)
	logger := testlogger.New(t)
	tlm := createTestTaskListManagerWithConfig(t, logger, controller, config, clock.NewMockedTimeSource())

	tlm.taskAckManager.SetReadLevel(0)
	atomic.StoreInt64(&tlm.taskWriter.maxReadLevel, maxReadLevel)
	tasks, readLevel, isReadBatchDone, err := tlm.taskReader.getTaskBatch(tlm.taskAckManager.GetReadLevel(), tlm.taskWriter.GetMaxReadLevel())
	assert.Empty(t, tasks)
	assert.Equal(t, int64(rangeSize*10), readLevel)
	assert.False(t, isReadBatchDone)
	assert.NoError(t, err)

	tlm.taskAckManager.SetReadLevel(readLevel)
	tasks, readLevel, isReadBatchDone, err = tlm.taskReader.getTaskBatch(tlm.taskAckManager.GetReadLevel(), tlm.taskWriter.GetMaxReadLevel())
	assert.Empty(t, tasks)
	assert.Equal(t, maxReadLevel, readLevel)
	assert.True(t, isReadBatchDone)
	assert.NoError(t, err)
}

func awaitCondition(cond func() bool, timeout time.Duration) bool {
	expiry := time.Now().Add(timeout)
	for !cond() {
		time.Sleep(time.Millisecond * 5)
		if time.Now().After(expiry) {
			return false
		}
	}
	return true
}

func TestTaskExpiryAndCompletion(t *testing.T) {
	const taskCount = 20
	const rangeSize = 10

	testCases := []struct {
		name               string
		batchSize          int
		maxTimeBtwnDeletes time.Duration
	}{
		{"test taskGC deleting due to size threshold", 2, time.Minute},
		{"test taskGC deleting due to time condition", 100, time.Nanosecond},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			controller := gomock.NewController(t)
			mockPartitioner := partition.NewMockPartitioner(controller)
			mockPartitioner.EXPECT().GetIsolationGroupByDomainID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
			mockDomainCache := cache.NewMockDomainCache(controller)
			mockDomainCache.EXPECT().GetDomainByID(gomock.Any()).Return(cache.CreateDomainCacheEntry("domainName"), nil).AnyTimes()
			mockDomainCache.EXPECT().GetDomainName(gomock.Any()).Return("domainName", nil).AnyTimes()
			mockHistoryService := history.NewMockClient(controller)
			logger := testlogger.New(t)
			timeSource := clock.NewRealTimeSource()
			tm := NewTestTaskManager(t, logger, timeSource)
			taskListID := NewTestTaskListID(t, "domainId", "tl", 0)
			cfg := defaultTestConfig()
			cfg.RangeSize = rangeSize
			cfg.MaxTaskDeleteBatchSize = dynamicconfig.GetIntPropertyFilteredByTaskListInfo(tc.batchSize)
			cfg.MaxTimeBetweenTaskDeletes = tc.maxTimeBtwnDeletes
			// set idle timer check to a really small value to assert that we don't accidentally drop tasks while blocking
			// on enqueuing a task to task buffer
			cfg.IdleTasklistCheckInterval = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(20 * time.Millisecond)
			tlMgr, err := NewManager(
				mockDomainCache,
				logger,
				metrics.NewClient(tally.NoopScope, metrics.Matching),
				tm,
				cluster.GetTestClusterMetadata(true),
				mockPartitioner,
				nil,
				func(Manager) {},
				taskListID,
				types.TaskListKindNormal.Ptr(),
				cfg,
				timeSource,
				timeSource.Now(),
				mockHistoryService,
			)
			assert.NoError(t, err)
			tlm := tlMgr.(*taskListManagerImpl)
			err = tlm.Start()
			assert.NoError(t, err)
			for i := int64(0); i < taskCount; i++ {
				scheduleID := i * 3
				addParams := AddTaskParams{
					TaskInfo: &persistence.TaskInfo{
						DomainID:                      "domainId",
						RunID:                         "run1",
						WorkflowID:                    "workflow1",
						ScheduleID:                    scheduleID,
						ScheduleToStartTimeoutSeconds: 100,
					},
				}
				if i%2 == 0 {
					// simulates creating a task whose scheduledToStartTimeout is already expired
					addParams.TaskInfo.ScheduleToStartTimeoutSeconds = -5
				}
				_, err = tlm.AddTask(context.Background(), addParams)
				assert.NoError(t, err)
			}
			assert.Equal(t, taskCount, tm.GetTaskCount(taskListID))

			// wait until all tasks are loaded by into in-memory buffers by task list manager
			// the buffer size should be one less than expected because dispatcher will dequeue the head
			assert.True(t, awaitCondition(func() bool {
				return len(tlm.taskReader.taskBuffers[defaultTaskBufferIsolationGroup]) >= (taskCount/2 - 1)
			}, time.Second))

			remaining := taskCount
			for i := 0; i < 2; i++ {
				// verify that (1) expired tasks are not returned in poll result (2) taskCleaner deletes tasks correctly
				for i := int64(0); i < taskCount/4; i++ {
					task, err := tlm.GetTask(context.Background(), nil)
					assert.NoError(t, err)
					assert.NotNil(t, task)
					task.Finish(nil)
				}
				remaining -= taskCount / 2
				// since every other task is expired, we expect half the tasks to be deleted
				// after poll consumed 1/4th of what is available
				assert.Equal(t, remaining, tm.GetTaskCount(taskListID))
			}
			tlm.Stop()
		})
	}
}

func TestTaskListManagerImpl_HasPollerAfter(t *testing.T) {
	for name, tc := range map[string]struct {
		outstandingPollers []string
		prepareManager     func(*taskListManagerImpl)
	}{
		"has_outstanding_pollers": {
			prepareManager: func(tlm *taskListManagerImpl) {
				tlm.addOutstandingPoller("poller1", "group1", func() {})
				tlm.addOutstandingPoller("poller2", "group2", func() {})
			},
		},
		"no_outstanding_pollers": {
			prepareManager: func(tlm *taskListManagerImpl) {
				tlm.pollerHistory.UpdatePollerInfo("identity", poller.Info{RatePerSecond: 1.0, IsolationGroup: "isolationGroup"})
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			controller := gomock.NewController(t)
			logger := testlogger.New(t)
			tlm := createTestTaskListManager(t, logger, controller)
			err := tlm.Start()
			assert.NoError(t, err)

			if tc.prepareManager != nil {
				tc.prepareManager(tlm)
			}

			assert.True(t, tlm.HasPollerAfter(time.Time{}))
		})
	}
}

func getIsolationgroupsHelper() []string {
	return testIsolationGroups
}

func TestRefreshTaskListPartitionConfig(t *testing.T) {
	testCases := []struct {
		name           string
		req            *types.TaskListPartitionConfig
		originalConfig *types.TaskListPartitionConfig
		setupMocks     func(*mockDeps)
		expectedConfig *types.TaskListPartitionConfig
		expectError    bool
		expectedError  string
	}{
		{
			name: "success - refresh from request",
			req: &types.TaskListPartitionConfig{
				Version:            2,
				NumReadPartitions:  3,
				NumWritePartitions: 3,
			},
			setupMocks: func(m *mockDeps) {},
			expectedConfig: &types.TaskListPartitionConfig{
				Version:            2,
				NumReadPartitions:  3,
				NumWritePartitions: 3,
			},
		},
		{
			name: "success - ignore older version",
			req: &types.TaskListPartitionConfig{
				Version:            2,
				NumReadPartitions:  3,
				NumWritePartitions: 3,
			},
			originalConfig: &types.TaskListPartitionConfig{
				Version:            3,
				NumReadPartitions:  2,
				NumWritePartitions: 2,
			},
			setupMocks: func(m *mockDeps) {},
			expectedConfig: &types.TaskListPartitionConfig{
				Version:            3,
				NumReadPartitions:  2,
				NumWritePartitions: 2,
			},
		},
		{
			name: "success - refresh from database",
			originalConfig: &types.TaskListPartitionConfig{
				Version:            3,
				NumReadPartitions:  2,
				NumWritePartitions: 2,
			},
			setupMocks: func(deps *mockDeps) {
				deps.mockTaskManager.EXPECT().GetTaskList(gomock.Any(), &persistence.GetTaskListRequest{
					DomainID:   "domain-id",
					DomainName: "domainName",
					TaskList:   "tl",
					TaskType:   persistence.TaskListTypeDecision,
				}).Return(&persistence.GetTaskListResponse{
					TaskListInfo: &persistence.TaskListInfo{
						AdaptivePartitionConfig: &persistence.TaskListPartitionConfig{
							Version:            4,
							NumReadPartitions:  10,
							NumWritePartitions: 10,
						},
					},
				}, nil)
			},
			expectedConfig: &types.TaskListPartitionConfig{
				Version:            4,
				NumReadPartitions:  10,
				NumWritePartitions: 10,
			},
		},
		{
			name: "failed to refresh from database",
			originalConfig: &types.TaskListPartitionConfig{
				Version:            3,
				NumReadPartitions:  2,
				NumWritePartitions: 2,
			},
			setupMocks: func(deps *mockDeps) {
				deps.mockTaskManager.EXPECT().GetTaskList(gomock.Any(), &persistence.GetTaskListRequest{
					DomainID:   "domain-id",
					DomainName: "domainName",
					TaskList:   "tl",
					TaskType:   persistence.TaskListTypeDecision,
				}).Return(nil, errors.New("some error"))
			},
			expectedConfig: &types.TaskListPartitionConfig{
				Version:            3,
				NumReadPartitions:  2,
				NumWritePartitions: 2,
			},
			expectError:   true,
			expectedError: "some error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tlID, err := NewIdentifier("domain-id", "tl", persistence.TaskListTypeDecision)
			require.NoError(t, err)
			tlm, deps := setupMocksForTaskListManager(t, tlID, types.TaskListKindNormal)
			tc.setupMocks(deps)
			tlm.partitionConfig = tc.originalConfig
			tlm.startWG.Done()

			err = tlm.RefreshTaskListPartitionConfig(context.Background(), tc.req)
			if tc.expectError {
				assert.ErrorContains(t, err, tc.expectedError)
				assert.Equal(t, tc.expectedConfig, tlm.TaskListPartitionConfig())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedConfig, tlm.TaskListPartitionConfig())
			}
		})
	}
}

func TestUpdateTaskListPartitionConfig(t *testing.T) {
	testCases := []struct {
		name           string
		req            *types.TaskListPartitionConfig
		originalConfig *types.TaskListPartitionConfig
		setupMocks     func(*mockDeps)
		expectedConfig *types.TaskListPartitionConfig
		expectError    bool
		expectedError  string
	}{
		{
			name: "success - no op",
			req: &types.TaskListPartitionConfig{
				NumReadPartitions:  3,
				NumWritePartitions: 3,
			},
			originalConfig: &types.TaskListPartitionConfig{
				Version:            1,
				NumReadPartitions:  3,
				NumWritePartitions: 3,
			},
			setupMocks: func(m *mockDeps) {},
			expectedConfig: &types.TaskListPartitionConfig{
				Version:            1,
				NumReadPartitions:  3,
				NumWritePartitions: 3,
			},
		},
		{
			name: "success - no op, nil pointer",
			req: &types.TaskListPartitionConfig{
				NumReadPartitions:  1,
				NumWritePartitions: 1,
			},
			originalConfig: nil,
			setupMocks:     func(m *mockDeps) {},
			expectedConfig: nil,
		},
		{
			name: "success - update",
			req: &types.TaskListPartitionConfig{
				NumReadPartitions:  3,
				NumWritePartitions: 1,
			},
			originalConfig: &types.TaskListPartitionConfig{
				Version:            1,
				NumReadPartitions:  3,
				NumWritePartitions: 3,
			},
			setupMocks: func(deps *mockDeps) {
				deps.mockTaskManager.EXPECT().UpdateTaskList(gomock.Any(), &persistence.UpdateTaskListRequest{
					DomainName: "domainName",
					TaskListInfo: &persistence.TaskListInfo{
						DomainID: "domain-id",
						Name:     "tl",
						AckLevel: 0,
						RangeID:  0,
						Kind:     persistence.TaskListKindNormal,
						AdaptivePartitionConfig: &persistence.TaskListPartitionConfig{
							Version:            2,
							NumReadPartitions:  3,
							NumWritePartitions: 1,
						},
					},
				}).Return(&persistence.UpdateTaskListResponse{}, nil)
				deps.mockMatchingClient.EXPECT().RefreshTaskListPartitionConfig(gomock.Any(), &types.MatchingRefreshTaskListPartitionConfigRequest{
					DomainUUID:   "domain-id",
					TaskList:     &types.TaskList{Name: "/__cadence_sys/tl/1", Kind: types.TaskListKindNormal.Ptr()},
					TaskListType: types.TaskListTypeDecision.Ptr(),
					PartitionConfig: &types.TaskListPartitionConfig{
						Version:            2,
						NumReadPartitions:  3,
						NumWritePartitions: 1,
					},
				}).Return(&types.MatchingRefreshTaskListPartitionConfigResponse{}, nil)
				deps.mockMatchingClient.EXPECT().RefreshTaskListPartitionConfig(gomock.Any(), &types.MatchingRefreshTaskListPartitionConfigRequest{
					DomainUUID:   "domain-id",
					TaskList:     &types.TaskList{Name: "/__cadence_sys/tl/2", Kind: types.TaskListKindNormal.Ptr()},
					TaskListType: types.TaskListTypeDecision.Ptr(),
					PartitionConfig: &types.TaskListPartitionConfig{
						Version:            2,
						NumReadPartitions:  3,
						NumWritePartitions: 1,
					},
				}).Return(&types.MatchingRefreshTaskListPartitionConfigResponse{}, nil)
			},
			expectedConfig: &types.TaskListPartitionConfig{
				Version:            2,
				NumReadPartitions:  3,
				NumWritePartitions: 1,
			},
		},
		{
			name: "success - push failures are ignored",
			req: &types.TaskListPartitionConfig{
				NumReadPartitions:  3,
				NumWritePartitions: 1,
			},
			originalConfig: &types.TaskListPartitionConfig{
				Version:            1,
				NumReadPartitions:  3,
				NumWritePartitions: 3,
			},
			setupMocks: func(deps *mockDeps) {
				deps.mockTaskManager.EXPECT().UpdateTaskList(gomock.Any(), &persistence.UpdateTaskListRequest{
					DomainName: "domainName",
					TaskListInfo: &persistence.TaskListInfo{
						DomainID: "domain-id",
						Name:     "tl",
						AckLevel: 0,
						RangeID:  0,
						Kind:     persistence.TaskListKindNormal,
						AdaptivePartitionConfig: &persistence.TaskListPartitionConfig{
							Version:            2,
							NumReadPartitions:  3,
							NumWritePartitions: 1,
						},
					},
				}).Return(&persistence.UpdateTaskListResponse{}, nil)
				deps.mockMatchingClient.EXPECT().RefreshTaskListPartitionConfig(gomock.Any(), &types.MatchingRefreshTaskListPartitionConfigRequest{
					DomainUUID:   "domain-id",
					TaskList:     &types.TaskList{Name: "/__cadence_sys/tl/1", Kind: types.TaskListKindNormal.Ptr()},
					TaskListType: types.TaskListTypeDecision.Ptr(),
					PartitionConfig: &types.TaskListPartitionConfig{
						Version:            2,
						NumReadPartitions:  3,
						NumWritePartitions: 1,
					},
				}).Return(nil, errors.New("matching client error"))
				deps.mockMatchingClient.EXPECT().RefreshTaskListPartitionConfig(gomock.Any(), &types.MatchingRefreshTaskListPartitionConfigRequest{
					DomainUUID:   "domain-id",
					TaskList:     &types.TaskList{Name: "/__cadence_sys/tl/2", Kind: types.TaskListKindNormal.Ptr()},
					TaskListType: types.TaskListTypeDecision.Ptr(),
					PartitionConfig: &types.TaskListPartitionConfig{
						Version:            2,
						NumReadPartitions:  3,
						NumWritePartitions: 1,
					},
				}).Return(nil, errors.New("matching client error"))
			},
			expectedConfig: &types.TaskListPartitionConfig{
				Version:            2,
				NumReadPartitions:  3,
				NumWritePartitions: 1,
			},
		},
		{
			name: "failed to update",
			req: &types.TaskListPartitionConfig{
				NumReadPartitions:  3,
				NumWritePartitions: 1,
			},
			originalConfig: &types.TaskListPartitionConfig{
				Version:            1,
				NumReadPartitions:  3,
				NumWritePartitions: 3,
			},
			setupMocks: func(deps *mockDeps) {
				deps.mockTaskManager.EXPECT().UpdateTaskList(gomock.Any(), &persistence.UpdateTaskListRequest{
					DomainName: "domainName",
					TaskListInfo: &persistence.TaskListInfo{
						DomainID: "domain-id",
						Name:     "tl",
						AckLevel: 0,
						RangeID:  0,
						Kind:     persistence.TaskListKindNormal,
						AdaptivePartitionConfig: &persistence.TaskListPartitionConfig{
							Version:            2,
							NumReadPartitions:  3,
							NumWritePartitions: 1,
						},
					},
				}).Return(nil, errors.New("some error"))
			},
			expectedConfig: &types.TaskListPartitionConfig{
				Version:            1,
				NumReadPartitions:  3,
				NumWritePartitions: 3,
			},
			expectError:   true,
			expectedError: "some error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tlID, err := NewIdentifier("domain-id", "tl", persistence.TaskListTypeDecision)
			require.NoError(t, err)
			tlm, deps := setupMocksForTaskListManager(t, tlID, types.TaskListKindNormal)
			tc.setupMocks(deps)
			tlm.partitionConfig = tc.originalConfig
			tlm.startWG.Done()

			err = tlm.UpdateTaskListPartitionConfig(context.Background(), tc.req)
			if tc.expectError {
				assert.ErrorContains(t, err, tc.expectedError)
				assert.Equal(t, int32(1), tlm.stopped)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedConfig, tlm.TaskListPartitionConfig())
		})
	}
}

func TestRefreshTaskListPartitionConfigConcurrency(t *testing.T) {
	tlID, err := NewIdentifier("domain-id", "/__cadence_sys/tl/1", persistence.TaskListTypeDecision)
	require.NoError(t, err)
	tlm, _ := setupMocksForTaskListManager(t, tlID, types.TaskListKindNormal)
	tlm.startWG.Done()

	var g errgroup.Group
	for i := 0; i < 100; i++ {
		v := i
		g.Go(func() error {
			return tlm.RefreshTaskListPartitionConfig(context.Background(), &types.TaskListPartitionConfig{Version: int64(v), NumReadPartitions: int32(v), NumWritePartitions: int32(v)})
		})
	}
	require.NoError(t, g.Wait())
	assert.Equal(t, int64(99), tlm.TaskListPartitionConfig().Version)
}

func TestUpdateTaskListPartitionConfigConcurrency(t *testing.T) {
	tlID, err := NewIdentifier("domain-id", "/__cadence_sys/tl/1", persistence.TaskListTypeDecision)
	require.NoError(t, err)
	tlm, deps := setupMocksForTaskListManager(t, tlID, types.TaskListKindNormal)
	deps.mockTaskManager.EXPECT().UpdateTaskList(gomock.Any(), gomock.Any()).Return(&persistence.UpdateTaskListResponse{}, nil).AnyTimes()
	deps.mockMatchingClient.EXPECT().RefreshTaskListPartitionConfig(gomock.Any(), gomock.Any()).Return(&types.MatchingRefreshTaskListPartitionConfigResponse{}, nil).AnyTimes()
	tlm.startWG.Done()

	var g errgroup.Group
	for i := 2; i < 102; i++ {
		v := i
		g.Go(func() error {
			return tlm.UpdateTaskListPartitionConfig(context.Background(), &types.TaskListPartitionConfig{NumReadPartitions: int32(v), NumWritePartitions: int32(v)})
		})
	}
	require.NoError(t, g.Wait())
	assert.Equal(t, int64(100), tlm.TaskListPartitionConfig().Version)
}

func TestManagerStart_RootPartition(t *testing.T) {
	tlID, err := NewIdentifier("domain-id", "tl", persistence.TaskListTypeDecision)
	require.NoError(t, err)
	tlm, deps := setupMocksForTaskListManager(t, tlID, types.TaskListKindNormal)
	deps.mockTaskManager.EXPECT().LeaseTaskList(gomock.Any(), &persistence.LeaseTaskListRequest{
		DomainID:   "domain-id",
		DomainName: "domainName",
		TaskList:   "tl",
		TaskType:   persistence.TaskListTypeDecision,
	}).Return(&persistence.LeaseTaskListResponse{
		TaskListInfo: &persistence.TaskListInfo{
			DomainID: "domain-id",
			Name:     "tl",
			Kind:     persistence.TaskListKindNormal,
			AckLevel: 0,
			RangeID:  0,
			AdaptivePartitionConfig: &persistence.TaskListPartitionConfig{
				Version:            1,
				NumReadPartitions:  2,
				NumWritePartitions: 2,
			},
		},
	}, nil)
	deps.mockMatchingClient.EXPECT().RefreshTaskListPartitionConfig(gomock.Any(), &types.MatchingRefreshTaskListPartitionConfigRequest{
		DomainUUID:   "domain-id",
		TaskList:     &types.TaskList{Name: "/__cadence_sys/tl/1", Kind: types.TaskListKindNormal.Ptr()},
		TaskListType: types.TaskListTypeDecision.Ptr(),
		PartitionConfig: &types.TaskListPartitionConfig{
			Version:            1,
			NumReadPartitions:  2,
			NumWritePartitions: 2,
		},
	}).Return(&types.MatchingRefreshTaskListPartitionConfigResponse{}, nil)
	assert.NoError(t, tlm.Start())
	assert.Equal(t, &types.TaskListPartitionConfig{Version: 1, NumReadPartitions: 2, NumWritePartitions: 2}, tlm.TaskListPartitionConfig())
	tlm.stopWG.Wait()
}

func TestManagerStart_NonRootPartition(t *testing.T) {
	tlID, err := NewIdentifier("domain-id", "/__cadence_sys/tl/1", persistence.TaskListTypeDecision)
	require.NoError(t, err)
	tlm, deps := setupMocksForTaskListManager(t, tlID, types.TaskListKindNormal)
	deps.mockTaskManager.EXPECT().GetTaskList(gomock.Any(), &persistence.GetTaskListRequest{
		DomainID:   "domain-id",
		DomainName: "domainName",
		TaskList:   "tl",
		TaskType:   persistence.TaskListTypeDecision,
	}).Return(&persistence.GetTaskListResponse{
		TaskListInfo: &persistence.TaskListInfo{
			DomainID: "domain-id",
			Name:     "tl",
			Kind:     persistence.TaskListKindNormal,
			AckLevel: 0,
			RangeID:  0,
			AdaptivePartitionConfig: &persistence.TaskListPartitionConfig{
				Version:            1,
				NumReadPartitions:  3,
				NumWritePartitions: 3,
			},
		},
	}, nil)
	deps.mockTaskManager.EXPECT().LeaseTaskList(gomock.Any(), &persistence.LeaseTaskListRequest{
		DomainID:   "domain-id",
		DomainName: "domainName",
		TaskList:   "/__cadence_sys/tl/1",
		TaskType:   persistence.TaskListTypeDecision,
	}).Return(&persistence.LeaseTaskListResponse{
		TaskListInfo: &persistence.TaskListInfo{
			DomainID: "domain-id",
			Name:     "/__cadence_sys/tl/1",
			Kind:     persistence.TaskListKindNormal,
			AckLevel: 0,
			RangeID:  0,
		},
	}, nil)
	assert.NoError(t, tlm.Start())
	assert.Equal(t, &types.TaskListPartitionConfig{
		Version:            1,
		NumReadPartitions:  3,
		NumWritePartitions: 3,
	}, tlm.TaskListPartitionConfig())
}

func TestDispatchTask(t *testing.T) {
	testCases := []struct {
		name                        string
		mockSetup                   func(matcher *MockTaskMatcher, taskCompleter *MockTaskCompleter, ctx context.Context, task *InternalTask)
		enableStandByTaskCompletion bool
		activeClusterName           string
		err                         error
	}{
		{
			name: "active cluster - disabled StandByTaskCompletion - task sent to MustOffer and no error returned",
			mockSetup: func(matcher *MockTaskMatcher, taskCompleter *MockTaskCompleter, ctx context.Context, task *InternalTask) {
				matcher.EXPECT().MustOffer(ctx, task).Return(nil).Times(1)
			},
			activeClusterName: cluster.TestCurrentClusterName,
			err:               nil,
		},
		{
			name: "active cluster - disabled StandByTaskCompletion - task sent to MustOffer and error returned",
			mockSetup: func(matcher *MockTaskMatcher, taskCompleter *MockTaskCompleter, ctx context.Context, task *InternalTask) {
				matcher.EXPECT().MustOffer(ctx, task).Return(errors.New("no-task-completion-must-offer-error")).Times(1)
			},
			activeClusterName: cluster.TestCurrentClusterName,
			err:               errors.New("no-task-completion-must-offer-error"),
		},
		{
			name: "active cluster - enabled StandByTaskCompletion - task sent to MustOffer and no error returned",
			mockSetup: func(matcher *MockTaskMatcher, taskCompleter *MockTaskCompleter, ctx context.Context, task *InternalTask) {
				matcher.EXPECT().MustOffer(ctx, task).Return(nil).Times(1)
			},
			enableStandByTaskCompletion: true,
			activeClusterName:           cluster.TestCurrentClusterName,
			err:                         nil,
		},
		{
			name: "active cluster - enabled StandByTaskCompletion - task sent to MustOffer and error returned",
			mockSetup: func(matcher *MockTaskMatcher, taskCompleter *MockTaskCompleter, ctx context.Context, task *InternalTask) {
				matcher.EXPECT().MustOffer(ctx, task).Return(errors.New("task-completion-must-offer-error")).Times(1)
			},
			enableStandByTaskCompletion: true,
			activeClusterName:           cluster.TestCurrentClusterName,
			err:                         errors.New("task-completion-must-offer-error"),
		},
		{
			name: "standby cluster - disabled StandByTaskCompletion - task sent to MustOffer and no error returned",
			mockSetup: func(matcher *MockTaskMatcher, taskCompleter *MockTaskCompleter, ctx context.Context, task *InternalTask) {
				matcher.EXPECT().MustOffer(ctx, task).Return(nil).Times(1)
			},
			activeClusterName: cluster.TestAlternativeClusterName,
			err:               nil,
		},
		{
			name: "standby cluster - disabled StandByTaskCompletion - task sent to MustOffer and error returned",
			mockSetup: func(matcher *MockTaskMatcher, taskCompleter *MockTaskCompleter, ctx context.Context, task *InternalTask) {
				matcher.EXPECT().MustOffer(ctx, task).Return(errors.New("no-task-completion-must-offer-error")).Times(1)
			},
			activeClusterName: cluster.TestAlternativeClusterName,
			err:               errors.New("no-task-completion-must-offer-error"),
		},
		{
			name: "standby cluster - enabled StandByTaskCompletion - task completed",
			mockSetup: func(matcher *MockTaskMatcher, taskCompleter *MockTaskCompleter, ctx context.Context, task *InternalTask) {
				taskCompleter.EXPECT().CompleteTaskIfStarted(ctx, task).Return(nil).Times(1)
			},
			enableStandByTaskCompletion: true,
			activeClusterName:           cluster.TestAlternativeClusterName,
			err:                         nil,
		},
		{
			name: "standby cluster - enabled StandByTaskCompletion - task completion error",
			mockSetup: func(matcher *MockTaskMatcher, taskCompleter *MockTaskCompleter, ctx context.Context, task *InternalTask) {
				taskCompleter.EXPECT().CompleteTaskIfStarted(ctx, task).Return(errTaskNotStarted).Times(1)
			},
			enableStandByTaskCompletion: true,
			activeClusterName:           cluster.TestAlternativeClusterName,
			err:                         errTaskNotStarted,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			logger := testlogger.New(t)
			tlm := createTestTaskListManager(t, logger, controller)

			task := &InternalTask{
				Event: &genericTaskInfo{
					TaskInfo: &persistence.TaskInfo{
						DomainID:   constants.TestDomainID,
						WorkflowID: constants.TestWorkflowID,
						RunID:      constants.TestRunID,
					},
				},
			}

			taskMatcher := NewMockTaskMatcher(controller)
			taskCompleter := NewMockTaskCompleter(controller)
			tlm.matcher = taskMatcher
			tlm.taskCompleter = taskCompleter
			tlm.config.EnableStandbyTaskCompletion = func() bool {
				return tc.enableStandByTaskCompletion
			}

			mockDomainCache := cache.NewMockDomainCache(controller)
			cacheEntry := cache.NewGlobalDomainCacheEntryForTest(
				&persistence.DomainInfo{ID: constants.TestDomainID, Name: constants.TestDomainName},
				&persistence.DomainConfig{Retention: 1},
				&persistence.DomainReplicationConfig{
					ActiveClusterName: tc.activeClusterName,
					Clusters:          []*persistence.ClusterReplicationConfig{{ClusterName: cluster.TestCurrentClusterName}, {ClusterName: cluster.TestAlternativeClusterName}},
				},
				1,
			)

			mockDomainCache.EXPECT().GetDomainByID(constants.TestDomainID).Return(cacheEntry, nil).AnyTimes()
			tlm.domainCache = mockDomainCache

			ctx := context.Background()
			tc.mockSetup(taskMatcher, taskCompleter, ctx, task)

			err := tlm.DispatchTask(ctx, task)

			if tc.err != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.err, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetNumPartitions(t *testing.T) {
	tlID, err := NewIdentifier("domain-id", "tl", persistence.TaskListTypeDecision)
	require.NoError(t, err)
	tlm, deps := setupMocksForTaskListManager(t, tlID, types.TaskListKindNormal)
	require.NoError(t, deps.dynamicClient.UpdateValue(dynamicconfig.MatchingEnableGetNumberOfPartitionsFromCache, true))
	assert.NotPanics(t, func() { tlm.matcher.UpdateRatelimit(common.Ptr(float64(100))) })
}
