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

package matching

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/partition"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

func TestDeliverBufferTasks(t *testing.T) {
	controller := gomock.NewController(t)

	tests := []func(tlm *taskListManagerImpl){
		func(tlm *taskListManagerImpl) { tlm.taskReader.cancelFunc() },
		func(tlm *taskListManagerImpl) {
			rps := 0.1
			tlm.matcher.UpdateRatelimit(&rps)
			tlm.taskReader.taskBuffers[defaultTaskBufferIsolationGroup] <- &persistence.TaskInfo{}
			_, err := tlm.matcher.ratelimit(context.Background()) // consume the token
			assert.NoError(t, err)
			tlm.taskReader.cancelFunc()
		},
	}
	for _, test := range tests {
		tlm := createTestTaskListManager(controller)
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

func TestDeliverBufferTasks_NoPollers(t *testing.T) {
	controller := gomock.NewController(t)

	tlm := createTestTaskListManager(controller)
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

	tlm := createTestTaskListManager(controller)
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

func createTestTaskListManager(controller *gomock.Controller) *taskListManagerImpl {
	return createTestTaskListManagerWithConfig(controller, defaultTestConfig())
}

func createTestTaskListManagerWithConfig(controller *gomock.Controller, cfg *Config) *taskListManagerImpl {
	logger, err := loggerimpl.NewDevelopment()
	if err != nil {
		panic(err)
	}
	tm := newTestTaskManager(logger)
	mockPartitioner := partition.NewMockPartitioner(controller)
	mockPartitioner.EXPECT().GetIsolationGroupByDomainID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
	mockDomainCache := cache.NewMockDomainCache(controller)
	mockDomainCache.EXPECT().GetDomainByID(gomock.Any()).Return(cache.CreateDomainCacheEntry("domainName"), nil).AnyTimes()
	mockDomainCache.EXPECT().GetDomainName(gomock.Any()).Return("domainName", nil).AnyTimes()
	me := newMatchingEngine(
		cfg, tm, nil, logger, mockDomainCache, mockPartitioner,
	)
	tl := "tl"
	dID := "domain"
	tlID := newTestTaskListID(dID, tl, persistence.TaskListTypeActivity)
	tlKind := types.TaskListKindNormal
	tlMgr, err := newTaskListManager(me, tlID, &tlKind, cfg, time.Now())
	if err != nil {
		logger.Fatal("error when createTestTaskListManager", tag.Error(err))
	}
	return tlMgr.(*taskListManagerImpl)
}

func TestDescribeTaskList(t *testing.T) {
	controller := gomock.NewController(t)

	startTaskID := int64(1)
	taskCount := int64(3)
	PollerIdentity := "test-poll"

	// Create taskList Manager and set taskList state
	tlm := createTestTaskListManager(controller)
	tlm.db.rangeID = int64(1)
	tlm.taskAckManager.SetAckLevel(0)

	for i := int64(0); i < taskCount; i++ {
		err := tlm.taskAckManager.ReadItem(startTaskID + i)
		assert.Nil(t, err)
	}

	includeTaskStatus := false
	descResp := tlm.DescribeTaskList(includeTaskStatus)
	require.Equal(t, 0, len(descResp.GetPollers()))
	require.Nil(t, descResp.GetTaskListStatus())

	includeTaskStatus = true
	taskListStatus := tlm.DescribeTaskList(includeTaskStatus).GetTaskListStatus()
	require.NotNil(t, taskListStatus)
	require.Zero(t, taskListStatus.GetAckLevel())
	require.Equal(t, taskCount, taskListStatus.GetReadLevel())
	require.Equal(t, int64(0), taskListStatus.GetBacklogCountHint())
	require.True(t, taskListStatus.GetRatePerSecond() > (_defaultTaskDispatchRPS-1))
	require.True(t, taskListStatus.GetRatePerSecond() < (_defaultTaskDispatchRPS+1))
	taskIDBlock := taskListStatus.GetTaskIDBlock()
	require.Equal(t, int64(1), taskIDBlock.GetStartID())
	require.Equal(t, tlm.config.RangeSize, taskIDBlock.GetEndID())

	// Add a poller and complete all tasks
	tlm.pollerHistory.updatePollerInfo(pollerIdentity(PollerIdentity), pollerInfo{})
	for i := int64(0); i < taskCount; i++ {
		tlm.taskAckManager.AckItem(startTaskID + i)
	}

	descResp = tlm.DescribeTaskList(includeTaskStatus)
	require.Equal(t, 1, len(descResp.GetPollers()))
	require.Equal(t, PollerIdentity, descResp.Pollers[0].GetIdentity())
	require.NotEmpty(t, descResp.Pollers[0].GetLastAccessTime())
	require.True(t, descResp.Pollers[0].GetRatePerSecond() > (_defaultTaskDispatchRPS-1))

	rps := 5.0
	tlm.pollerHistory.updatePollerInfo(pollerIdentity(PollerIdentity), pollerInfo{ratePerSecond: &rps})
	descResp = tlm.DescribeTaskList(includeTaskStatus)
	require.Equal(t, 1, len(descResp.GetPollers()))
	require.Equal(t, PollerIdentity, descResp.Pollers[0].GetIdentity())
	require.True(t, descResp.Pollers[0].GetRatePerSecond() > 4.0 && descResp.Pollers[0].GetRatePerSecond() < 6.0)

	taskListStatus = descResp.GetTaskListStatus()
	require.NotNil(t, taskListStatus)
	require.Equal(t, taskCount, taskListStatus.GetAckLevel())
	require.Zero(t, taskListStatus.GetBacklogCountHint())
}

func tlMgrStartWithoutNotifyEvent(tlm *taskListManagerImpl) {
	// mimic tlm.Start() but avoid calling notifyEvent
	tlm.liveness.Start()
	tlm.startWG.Done()
	go tlm.taskReader.dispatchBufferedTasks(defaultTaskBufferIsolationGroup)
	go tlm.taskReader.getTasksPump()
}

func TestCheckIdleTaskList(t *testing.T) {
	controller := gomock.NewController(t)

	cfg := NewConfig(dynamicconfig.NewNopCollection(), "some random hostname")
	cfg.IdleTasklistCheckInterval = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(10 * time.Millisecond)

	// Idle
	tlm := createTestTaskListManagerWithConfig(controller, cfg)
	tlMgrStartWithoutNotifyEvent(tlm)
	time.Sleep(20 * time.Millisecond)
	require.False(t, atomic.CompareAndSwapInt32(&tlm.stopped, 0, 1))

	// Active poll-er
	tlm = createTestTaskListManagerWithConfig(controller, cfg)
	tlMgrStartWithoutNotifyEvent(tlm)
	time.Sleep(8 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond)
	_, _ = tlm.GetTask(ctx, nil)
	cancel()
	time.Sleep(6 * time.Millisecond)
	require.Equal(t, int32(0), tlm.stopped)
	tlm.Stop()
	require.Equal(t, int32(1), tlm.stopped)

	// Active adding task
	domainID := uuid.New()
	workflowID := "some random workflowID"
	runID := "some random runID"

	addTaskParam := addTaskParams{
		execution: &types.WorkflowExecution{
			WorkflowID: workflowID,
			RunID:      runID,
		},
		taskInfo: &persistence.TaskInfo{
			DomainID:               domainID,
			WorkflowID:             workflowID,
			RunID:                  runID,
			ScheduleID:             2,
			ScheduleToStartTimeout: 5,
			CreatedTime:            time.Now(),
		},
	}
	tlm = createTestTaskListManagerWithConfig(controller, cfg)
	tlMgrStartWithoutNotifyEvent(tlm)
	time.Sleep(8 * time.Millisecond)
	ctx, cancel = context.WithTimeout(context.Background(), time.Microsecond)
	_, _ = tlm.AddTask(ctx, addTaskParam)
	cancel()
	time.Sleep(6 * time.Millisecond)
	require.Equal(t, int32(0), tlm.stopped)
	tlm.Stop()
	require.Equal(t, int32(1), tlm.stopped)
}

func TestAddTaskStandby(t *testing.T) {
	controller := gomock.NewController(t)

	cfg := NewConfig(dynamicconfig.NewNopCollection(), "some random hostname")
	cfg.IdleTasklistCheckInterval = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(10 * time.Millisecond)

	tlm := createTestTaskListManagerWithConfig(controller, cfg)
	tlMgrStartWithoutNotifyEvent(tlm)
	// stop taskWriter so that we can check if there's any call to it
	// otherwise the task persist process is async and hard to test
	tlm.taskWriter.Stop()

	domainID := uuid.New()
	workflowID := "some random workflowID"
	runID := "some random runID"

	addTaskParam := addTaskParams{
		execution: &types.WorkflowExecution{
			WorkflowID: workflowID,
			RunID:      runID,
		},
		taskInfo: &persistence.TaskInfo{
			DomainID:               domainID,
			WorkflowID:             workflowID,
			RunID:                  runID,
			ScheduleID:             2,
			ScheduleToStartTimeout: 5,
			CreatedTime:            time.Now(),
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

	addTaskParam.forwardedFrom = "from child partition"
	syncMatch, err = tlm.AddTask(context.Background(), addTaskParam)
	require.Error(t, err) // should not persist the task
	require.False(t, syncMatch)
}

func TestGetPollerIsolationGroup(t *testing.T) {
	controller := gomock.NewController(t)

	config := defaultTestConfig()
	config.LongPollExpirationInterval = dynamicconfig.GetDurationPropertyFnFilteredByTaskListInfo(30 * time.Second)
	tlm := createTestTaskListManagerWithConfig(controller, config)

	bgCtx := context.WithValue(context.Background(), pollerIDKey, "poller0")
	bgCtx = context.WithValue(bgCtx, identityKey, "id0")
	bgCtx = context.WithValue(bgCtx, _isolationGroupKey, config.AllIsolationGroups[0])
	ctx, cancel := context.WithTimeout(bgCtx, time.Second)
	_, err := tlm.GetTask(ctx, nil)
	cancel()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrNoTasks.Error())

	// we should get isolation groups that showed up within last 10 seconds
	groups := tlm.getPollerIsolationGroups()
	assert.Equal(t, 1, len(groups))
	assert.Equal(t, config.AllIsolationGroups[0], groups[0])

	// after 10s, the poller from that isolation group are cleared from the poller history
	time.Sleep(10 * time.Second)
	groups = tlm.getPollerIsolationGroups()
	assert.Equal(t, 0, len(groups))

	// we should get isolation groups of outstanding pollers
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(bgCtx, time.Second*20)
		_, err := tlm.GetTask(ctx, nil)
		cancel()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ErrNoTasks.Error())
	}()
	time.Sleep(11 * time.Second)
	groups = tlm.getPollerIsolationGroups()
	wg.Wait()
	assert.Equal(t, 1, len(groups))
	assert.Equal(t, config.AllIsolationGroups[0], groups[0])
}

// return a client side tasklist throttle error from the rate limiter.
// The expected behaviour is to retry
func TestRateLimitErrorsFromTasklistDispatch(t *testing.T) {
	controller := gomock.NewController(t)

	config := defaultTestConfig()
	config.EnableTasklistIsolation = func(domain string) bool { return true }
	tlm := createTestTaskListManagerWithConfig(controller, config)

	tlm.taskReader.dispatchTask = func(ctx context.Context, task *InternalTask) error {
		return ErrTasklistThrottled
	}
	tlm.taskReader.getIsolationGroupForTask = func(ctx context.Context, info *persistence.TaskInfo) (string, error) { return "datacenterA", nil }

	breakDispatcher, breakRetryLoop := tlm.taskReader.dispatchSingleTaskFromBuffer("datacenterA", &persistence.TaskInfo{})
	assert.False(t, breakDispatcher)
	assert.False(t, breakRetryLoop)
}

// The intent of this test: SingleTaskDispatch is one of two places where tasks are written to
// the taskreader.taskBuffers channels. As such, it needs to take care to not accidentally
// hit the channel when it's full, as it'll block, causing a deadlock (due to both this dispatch
// and the async task pump blocking trying to read to the channel)
func TestSingleTaskDispatchDoesNotDeadlock(t *testing.T) {
	controller := gomock.NewController(t)

	config := defaultTestConfig()
	config.EnableTasklistIsolation = func(domain string) bool { return true }
	tlm := createTestTaskListManagerWithConfig(controller, config)

	// mock a timeout to cause tasks to be rerouted
	tlm.taskReader.dispatchTask = func(ctx context.Context, task *InternalTask) error {
		return context.DeadlineExceeded
	}
	tlm.taskReader.getIsolationGroupForTask = func(ctx context.Context, info *persistence.TaskInfo) (string, error) {
		// force this, on the second read, to always reroute to the other DC
		return "datacenterB", nil
	}

	maxBufferSize := config.GetTasksBatchSize("", "", 0) - 1

	tlm.taskReader.logger = loggerimpl.NewNopLogger()

	for i := 0; i < maxBufferSize; i++ {
		breakDispatcher, breakRetryLoop := tlm.taskReader.dispatchSingleTaskFromBuffer("datacenterA", &persistence.TaskInfo{})
		assert.False(t, breakDispatcher, "dispatch isn't shutting down")
		assert.True(t, breakRetryLoop, "should be able to successfully dispatch all these tasks to the other isolation group")
	}
	// We should see them all being redirected
	assert.Len(t, tlm.taskReader.taskBuffers["datacenterB"], maxBufferSize)
	// we should see none in DC A (they're not being redirected here)
	assert.Len(t, tlm.taskReader.taskBuffers["datacenterA"], 0)

	// ok, and here we try and ensure that this *does not block
	// and instead complains and live-retries
	breakDispatcher, breakRetryLoop := tlm.taskReader.dispatchSingleTaskFromBuffer("datacenterA", &persistence.TaskInfo{})
	assert.False(t, breakDispatcher, "dispatch isn't shutting down")
	assert.False(t, breakRetryLoop, "in the event of a task being unable to be cross-dispatched, the expectation is that it'll keep retrying")

	// and to be certain and avoid any accidental off-by-one errors, one more time to just be sure
	breakDispatcher, breakRetryLoop = tlm.taskReader.dispatchSingleTaskFromBuffer("datacenterA", &persistence.TaskInfo{})
	assert.False(t, breakDispatcher, "dispatch isn't shutting down")
	assert.False(t, breakRetryLoop, "in the event of a task being unable to be cross-dispatched, the expectation is that it'll keep retrying")
}

// This is a bit of a strange unit-test as it's
// ensuring that invalid behaviour is handled defensively.
// It *should never be the case* that the isolation group tries to
// dispatch to a buffer that isn't there, however, if it does, we want to just
// log this, emit a metric and fallback to the default isolation group.
func TestMisconfiguredZoneDoesNotBlock(t *testing.T) {
	controller := gomock.NewController(t)

	config := defaultTestConfig()
	config.EnableTasklistIsolation = func(domain string) bool { return true }
	tlm := createTestTaskListManagerWithConfig(controller, config)

	// mock a timeout to cause tasks to be rerouted
	tlm.taskReader.dispatchTask = func(ctx context.Context, task *InternalTask) error {
		return context.DeadlineExceeded
	}
	tlm.taskReader.getIsolationGroupForTask = func(ctx context.Context, info *persistence.TaskInfo) (string, error) {
		return "a misconfigured isolation group", nil
	}

	maxBufferSize := config.GetTasksBatchSize("", "", 0) - 1
	tlm.taskReader.logger = loggerimpl.NewNopLogger()

	for i := 0; i < maxBufferSize; i++ {
		breakDispatcher, breakRetryLoop := tlm.taskReader.dispatchSingleTaskFromBuffer("datacenterA", &persistence.TaskInfo{})
		assert.False(t, breakDispatcher, "dispatch isn't shutting down")
		assert.True(t, breakRetryLoop, "should be able to successfully dispatch all these tasks to the default isolation group")
	}
	// We should see them all being redirected
	assert.Len(t, tlm.taskReader.taskBuffers[defaultTaskBufferIsolationGroup], maxBufferSize)
	// we should see none in DC A (they're not being redirected here)
	assert.Len(t, tlm.taskReader.taskBuffers["datacenterA"], 0)

	// ok, and here we try and ensure that this *does not block
	// and instead complains and live-retries
	breakDispatcher, breakRetryLoop := tlm.taskReader.dispatchSingleTaskFromBuffer("datacenterA", &persistence.TaskInfo{})
	assert.False(t, breakDispatcher, "dispatch isn't shutting down")
	assert.False(t, breakRetryLoop, "in the event of a task being unable to be cross-dispatched, the expectation is that it'll keep retrying")
}
