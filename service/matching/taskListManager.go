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
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/messaging"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/partition"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
)

const (
	// Time budget for empty task to propagate through the function stack and be returned to
	// pollForActivityTask or pollForDecisionTask handler.
	returnEmptyTaskTimeBudget time.Duration = time.Second
)

var (
	taskListActivityTypeTag = metrics.TaskListTypeTag("activity")
	taskListDecisionTypeTag = metrics.TaskListTypeTag("decision")
)

type (
	addTaskParams struct {
		execution                *types.WorkflowExecution
		taskInfo                 *persistence.TaskInfo
		source                   types.TaskSource
		forwardedFrom            string
		activityTaskDispatchInfo *types.ActivityTaskDispatchInfo
	}

	taskListManager interface {
		Start() error
		Stop()
		// AddTask adds a task to the task list. This method will first attempt a synchronous
		// match with a poller. When that fails, task will be written to database and later
		// asynchronously matched with a poller
		AddTask(ctx context.Context, params addTaskParams) (syncMatch bool, err error)
		// GetTask blocks waiting for a task Returns error when context deadline is exceeded
		// maxDispatchPerSecond is the max rate at which tasks are allowed to be dispatched
		// from this task list to pollers
		GetTask(ctx context.Context, maxDispatchPerSecond *float64) (*InternalTask, error)
		// DispatchTask dispatches a task to a poller. When there are no pollers to pick
		// up the task, this method will return error. Task will not be persisted to db
		DispatchTask(ctx context.Context, task *InternalTask) error
		// DispatchQueryTask will dispatch query to local or remote poller. If forwarded then result or error is returned,
		// if dispatched to local poller then nil and nil is returned.
		DispatchQueryTask(ctx context.Context, taskID string, request *types.MatchingQueryWorkflowRequest) (*types.QueryWorkflowResponse, error)
		CancelPoller(pollerID string)
		GetAllPollerInfo() []*types.PollerInfo
		HasPollerAfter(accessTime time.Time) bool
		// DescribeTaskList returns information about the target tasklist
		DescribeTaskList(includeTaskListStatus bool) *types.DescribeTaskListResponse
		String() string
		GetTaskListKind() types.TaskListKind
		TaskListID() *taskListID
	}

	outstandingPollerInfo struct {
		isolationGroup string
		cancel         context.CancelFunc
	}

	// Single task list in memory state
	taskListManagerImpl struct {
		createTime      time.Time
		enableIsolation bool
		taskListID      *taskListID
		taskListKind    types.TaskListKind // sticky taskList has different process in persistence
		config          *taskListConfig
		db              *taskListDB
		taskWriter      *taskWriter
		taskReader      *taskReader // reads tasks from db and async matches it with poller
		liveness        *liveness
		taskGC          *taskGC
		taskAckManager  messaging.AckManager // tracks ackLevel for delivered messages
		matcher         *TaskMatcher         // for matching a task producer with a poller
		clusterMetadata cluster.Metadata
		domainCache     cache.DomainCache
		partitioner     partition.Partitioner
		logger          log.Logger
		scope           metrics.Scope
		domainName      string
		// pollerHistory stores poller which poll from this tasklist in last few minutes
		pollerHistory *pollerHistory
		// outstandingPollsMap is needed to keep track of all outstanding pollers for a
		// particular tasklist.  PollerID generated by frontend is used as the key and
		// CancelFunc is the value.  This is used to cancel the context to unblock any
		// outstanding poller when the frontend detects client connection is closed to
		// prevent tasks being dispatched to zombie pollers.
		outstandingPollsLock sync.Mutex
		outstandingPollsMap  map[string]outstandingPollerInfo
		startWG              sync.WaitGroup // ensures that background processes do not start until setup is ready
		stopped              int32
		closeCallback        func(taskListManager)
	}
)

const (
	// maxSyncMatchWaitTime is the max amount of time that we are willing to wait for a sync match to happen
	maxSyncMatchWaitTime = 200 * time.Millisecond
)

var _ taskListManager = (*taskListManagerImpl)(nil)

var errRemoteSyncMatchFailed = &types.RemoteSyncMatchedError{Message: "remote sync match failed"}

func newTaskListManager(
	e *matchingEngineImpl,
	taskList *taskListID,
	taskListKind *types.TaskListKind,
	config *Config,
	createTime time.Time,
) (taskListManager, error) {

	taskListConfig, err := newTaskListConfig(taskList, config, e.domainCache)
	if err != nil {
		return nil, err
	}

	if taskListKind == nil {
		normalTaskListKind := types.TaskListKindNormal
		taskListKind = &normalTaskListKind
	}
	domainName, err := e.domainCache.GetDomainName(taskList.domainID)
	if err != nil {
		return nil, err
	}
	scope := newPerTaskListScope(domainName, taskList.name, *taskListKind, e.metricsClient, metrics.MatchingTaskListMgrScope)
	db := newTaskListDB(e.taskManager, taskList.domainID, domainName, taskList.name, taskList.taskType, int(*taskListKind), e.logger)

	tlMgr := &taskListManagerImpl{
		createTime:          createTime,
		enableIsolation:     taskListConfig.EnableTasklistIsolation(),
		domainCache:         e.domainCache,
		clusterMetadata:     e.clusterMetadata,
		partitioner:         e.partitioner,
		taskListID:          taskList,
		taskListKind:        *taskListKind,
		logger:              e.logger.WithTags(tag.WorkflowDomainName(domainName), tag.WorkflowTaskListName(taskList.name), tag.WorkflowTaskListType(taskList.taskType)),
		db:                  db,
		taskAckManager:      messaging.NewAckManager(e.logger),
		taskGC:              newTaskGC(db, taskListConfig),
		config:              taskListConfig,
		outstandingPollsMap: make(map[string]outstandingPollerInfo),
		domainName:          domainName,
		scope:               scope,
		closeCallback:       e.removeTaskListManager,
	}

	taskListTypeMetricScope := tlMgr.scope.Tagged(
		getTaskListTypeTag(taskList.taskType),
	)
	tlMgr.pollerHistory = newPollerHistory(func() {
		taskListTypeMetricScope.UpdateGauge(metrics.PollerPerTaskListCounter,
			float64(len(tlMgr.pollerHistory.getPollerInfo(time.Time{}))))
	})
	tlMgr.liveness = newLiveness(clock.NewRealTimeSource(), taskListConfig.IdleTasklistCheckInterval(), tlMgr.Stop)
	var isolationGroups []string
	if tlMgr.isIsolationMatcherEnabled() {
		isolationGroups = config.AllIsolationGroups
	}
	var fwdr *Forwarder
	if tlMgr.isFowardingAllowed(taskList, *taskListKind) {
		fwdr = newForwarder(&taskListConfig.forwarderConfig, taskList, *taskListKind, e.matchingClient, isolationGroups)
	}
	tlMgr.matcher = newTaskMatcher(taskListConfig, fwdr, tlMgr.scope, isolationGroups)
	tlMgr.taskWriter = newTaskWriter(tlMgr)
	tlMgr.taskReader = newTaskReader(tlMgr, isolationGroups)
	tlMgr.startWG.Add(1)
	return tlMgr, nil
}

// Starts reading pump for the given task list.
// The pump fills up taskBuffer from persistence.
func (c *taskListManagerImpl) Start() error {
	defer c.startWG.Done()

	c.liveness.Start()
	if err := c.taskWriter.Start(); err != nil {
		c.Stop()
		return err
	}
	c.taskReader.Start()

	return nil
}

// Stops pump that fills up taskBuffer from persistence.
func (c *taskListManagerImpl) Stop() {
	if !atomic.CompareAndSwapInt32(&c.stopped, 0, 1) {
		return
	}
	c.closeCallback(c)
	c.liveness.Stop()
	c.taskWriter.Stop()
	c.taskReader.Stop()
	c.logger.Info("Task list manager state changed", tag.LifeCycleStopped)
}

func (c *taskListManagerImpl) handleErr(err error) error {
	var e *persistence.ConditionFailedError
	if errors.As(err, &e) {
		// This indicates the task list may have moved to another host.
		c.scope.IncCounter(metrics.ConditionFailedErrorPerTaskListCounter)
		c.logger.Debug("Stopping task list due to persistence condition failure.", tag.Error(err))
		c.Stop()
		if c.taskListKind == types.TaskListKindSticky {
			// TODO: we don't see this error in our logs, we might be able to remove this error
			err = &types.InternalServiceError{Message: common.StickyTaskConditionFailedErrorMsg}
		}
	}
	return err
}

// AddTask adds a task to the task list. This method will first attempt a synchronous
// match with a poller. When there are no pollers or if rate limit is exceeded, task will
// be written to database and later asynchronously matched with a poller
func (c *taskListManagerImpl) AddTask(ctx context.Context, params addTaskParams) (bool, error) {
	c.startWG.Wait()
	if c.shouldReload() {
		c.Stop()
		return false, errShutdown
	}
	if params.forwardedFrom == "" {
		// request sent by history service
		c.liveness.markAlive(time.Now())
	}
	var syncMatch bool
	_, err := c.executeWithRetry(func() (interface{}, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		domainEntry, err := c.domainCache.GetDomainByID(params.taskInfo.DomainID)
		if err != nil {
			return nil, err
		}

		isForwarded := params.forwardedFrom != ""

		if _, err := domainEntry.IsActiveIn(c.clusterMetadata.GetCurrentClusterName()); err != nil {
			// standby task, only persist when task is not forwarded from child partition
			syncMatch = false
			if isForwarded {
				return &persistence.CreateTasksResponse{}, errRemoteSyncMatchFailed
			}

			r, err := c.taskWriter.appendTask(params.execution, params.taskInfo)
			return r, err
		}

		isolationGroup, err := c.getIsolationGroupForTask(ctx, params.taskInfo)
		if err != nil {
			return false, err
		}
		// active task, try sync match first
		syncMatch, err = c.trySyncMatch(ctx, params, isolationGroup)
		if syncMatch {
			return &persistence.CreateTasksResponse{}, err
		}
		if params.activityTaskDispatchInfo != nil {
			return false, errRemoteSyncMatchFailed
		}

		if isForwarded {
			// forwarded from child partition - only do sync match
			// child partition will persist the task when sync match fails
			return &persistence.CreateTasksResponse{}, errRemoteSyncMatchFailed
		}

		return c.taskWriter.appendTask(params.execution, params.taskInfo)
	})

	if err == nil && !syncMatch {
		c.taskReader.Signal()
	}

	return syncMatch, err
}

// DispatchTask dispatches a task to a poller. When there are no pollers to pick
// up the task or if rate limit is exceeded, this method will return error. Task
// *will not* be persisted to db
func (c *taskListManagerImpl) DispatchTask(ctx context.Context, task *InternalTask) error {
	return c.matcher.MustOffer(ctx, task)
}

// DispatchQueryTask will dispatch query to local or remote poller. If forwarded then result or error is returned,
// if dispatched to local poller then nil and nil is returned.
func (c *taskListManagerImpl) DispatchQueryTask(
	ctx context.Context,
	taskID string,
	request *types.MatchingQueryWorkflowRequest,
) (*types.QueryWorkflowResponse, error) {
	c.startWG.Wait()
	task := newInternalQueryTask(taskID, request)
	return c.matcher.OfferQuery(ctx, task)
}

// GetTask blocks waiting for a task.
// Returns error when context deadline is exceeded
// maxDispatchPerSecond is the max rate at which tasks are allowed
// to be dispatched from this task list to pollers
func (c *taskListManagerImpl) GetTask(
	ctx context.Context,
	maxDispatchPerSecond *float64,
) (*InternalTask, error) {
	if c.shouldReload() {
		c.Stop()
		return nil, ErrNoTasks
	}
	c.liveness.markAlive(time.Now())
	task, err := c.getTask(ctx, maxDispatchPerSecond)
	if err != nil {
		return nil, err
	}
	task.domainName = c.domainName
	task.backlogCountHint = c.taskAckManager.GetBacklogCount()
	return task, nil
}

func (c *taskListManagerImpl) getTask(ctx context.Context, maxDispatchPerSecond *float64) (*InternalTask, error) {
	// We need to set a shorter timeout than the original ctx; otherwise, by the time ctx deadline is
	// reached, instead of emptyTask, context timeout error is returned to the frontend by the rpc stack,
	// which counts against our SLO. By shortening the timeout by a very small amount, the emptyTask can be
	// returned to the handler before a context timeout error is generated.
	childCtx, cancel := c.newChildContext(ctx, c.config.LongPollExpirationInterval(), returnEmptyTaskTimeBudget)
	defer cancel()

	isolationGroup, _ := ctx.Value(_isolationGroupKey).(string)
	pollerID, ok := ctx.Value(pollerIDKey).(string)
	if ok && pollerID != "" {
		// Found pollerID on context, add it to the map to allow it to be canceled in
		// response to CancelPoller call
		c.outstandingPollsLock.Lock()
		c.outstandingPollsMap[pollerID] = outstandingPollerInfo{isolationGroup: isolationGroup, cancel: cancel}
		c.outstandingPollsLock.Unlock()
		defer func() {
			c.outstandingPollsLock.Lock()
			delete(c.outstandingPollsMap, pollerID)
			c.outstandingPollsLock.Unlock()
		}()
	}

	identity, ok := ctx.Value(identityKey).(string)
	if ok && identity != "" {
		c.pollerHistory.updatePollerInfo(pollerIdentity(identity), pollerInfo{ratePerSecond: maxDispatchPerSecond, isolationGroup: isolationGroup})
		defer func() {
			// to update timestamp of this poller when long poll ends
			c.pollerHistory.updatePollerInfo(pollerIdentity(identity), pollerInfo{ratePerSecond: maxDispatchPerSecond, isolationGroup: isolationGroup})
		}()
	}

	domainEntry, err := c.domainCache.GetDomainByID(c.taskListID.domainID)
	if err != nil {
		return nil, err
	}

	// the desired global rate limit for the task list comes from the
	// poller, which lives inside the client side worker. There is
	// one rateLimiter for this entire task list and as we get polls,
	// we update the ratelimiter rps if it has changed from the last
	// value. Last poller wins if different pollers provide different values
	c.matcher.UpdateRatelimit(maxDispatchPerSecond)

	if _, err := domainEntry.IsActiveIn(c.clusterMetadata.GetCurrentClusterName()); err != nil {
		return c.matcher.PollForQuery(childCtx)
	}

	if c.isIsolationMatcherEnabled() {
		return c.matcher.Poll(childCtx, isolationGroup)
	}
	return c.matcher.Poll(childCtx, "")
}

// GetAllPollerInfo returns all pollers that polled from this tasklist in last few minutes
func (c *taskListManagerImpl) GetAllPollerInfo() []*types.PollerInfo {
	return c.pollerHistory.getPollerInfo(time.Time{})
}

// HasPollerAfter checks if there is any poller after a timestamp
func (c *taskListManagerImpl) HasPollerAfter(accessTime time.Time) bool {
	inflightPollerCount := 0
	c.outstandingPollsLock.Lock()
	inflightPollerCount = len(c.outstandingPollsMap)
	c.outstandingPollsLock.Unlock()
	if inflightPollerCount > 0 {
		return true
	}
	recentPollers := c.pollerHistory.getPollerInfo(accessTime)
	return len(recentPollers) > 0
}

func (c *taskListManagerImpl) CancelPoller(pollerID string) {
	c.outstandingPollsLock.Lock()
	info, ok := c.outstandingPollsMap[pollerID]
	c.outstandingPollsLock.Unlock()

	if ok && info.cancel != nil {
		info.cancel()
		c.logger.Info("canceled outstanding poller", tag.WorkflowDomainName(c.domainName))
	}
}

// DescribeTaskList returns information about the target tasklist, right now this API returns the
// pollers which polled this tasklist in last few minutes and status of tasklist's ackManager
// (readLevel, ackLevel, backlogCountHint and taskIDBlock).
func (c *taskListManagerImpl) DescribeTaskList(includeTaskListStatus bool) *types.DescribeTaskListResponse {
	response := &types.DescribeTaskListResponse{Pollers: c.GetAllPollerInfo()}
	if !includeTaskListStatus {
		return response
	}

	taskIDBlock := rangeIDToTaskIDBlock(c.db.RangeID(), c.config.RangeSize)
	response.TaskListStatus = &types.TaskListStatus{
		ReadLevel:        c.taskAckManager.GetReadLevel(),
		AckLevel:         c.taskAckManager.GetAckLevel(),
		BacklogCountHint: c.taskAckManager.GetBacklogCount(),
		RatePerSecond:    c.matcher.Rate(),
		TaskIDBlock: &types.TaskIDBlock{
			StartID: taskIDBlock.start,
			EndID:   taskIDBlock.end,
		},
	}

	return response
}

func (c *taskListManagerImpl) String() string {
	buf := new(bytes.Buffer)
	if c.taskListID.taskType == persistence.TaskListTypeActivity {
		buf.WriteString("Activity")
	} else {
		buf.WriteString("Decision")
	}
	rangeID := c.db.RangeID()
	fmt.Fprintf(buf, " task list %v\n", c.taskListID.name)
	fmt.Fprintf(buf, "RangeID=%v\n", rangeID)
	fmt.Fprintf(buf, "TaskIDBlock=%+v\n", rangeIDToTaskIDBlock(rangeID, c.config.RangeSize))
	fmt.Fprintf(buf, "AckLevel=%v\n", c.taskAckManager.GetAckLevel())
	fmt.Fprintf(buf, "MaxReadLevel=%v\n", c.taskAckManager.GetReadLevel())

	return buf.String()
}

func (c *taskListManagerImpl) GetTaskListKind() types.TaskListKind {
	return c.taskListKind
}

func (c *taskListManagerImpl) TaskListID() *taskListID {
	return c.taskListID
}

// Retry operation on transient error. On rangeID update by another process calls c.Stop().
func (c *taskListManagerImpl) executeWithRetry(
	operation func() (interface{}, error),
) (result interface{}, err error) {

	op := func() error {
		result, err = operation()
		return err
	}

	throttleRetry := backoff.NewThrottleRetry(
		backoff.WithRetryPolicy(persistenceOperationRetryPolicy),
		backoff.WithRetryableError(persistence.IsTransientError),
	)
	err = c.handleErr(throttleRetry.Do(context.Background(), op))
	return
}

func (c *taskListManagerImpl) trySyncMatch(ctx context.Context, params addTaskParams, isolationGroup string) (bool, error) {
	task := newInternalTask(params.taskInfo, nil, params.source, params.forwardedFrom, true, params.activityTaskDispatchInfo, isolationGroup)
	childCtx := ctx
	cancel := func() {}
	waitTime := maxSyncMatchWaitTime
	if params.activityTaskDispatchInfo != nil {
		waitTime = c.config.ActivityTaskSyncMatchWaitTime(params.activityTaskDispatchInfo.WorkflowDomain)
	}
	if !task.isForwarded() {
		// when task is forwarded from another matching host, we trust the context as is
		// otherwise, we override to limit the amount of time we can block on sync match
		childCtx, cancel = c.newChildContext(ctx, waitTime, time.Second)
	}
	var matched bool
	var err error
	if params.activityTaskDispatchInfo != nil {
		matched, err = c.matcher.offerOrTimeout(childCtx, task)
	} else {
		matched, err = c.matcher.Offer(childCtx, task)
	}
	cancel()
	return matched, err
}

// newChildContext creates a child context with desired timeout.
// if tailroom is non-zero, then child context timeout will be
// the minOf(parentCtx.Deadline()-tailroom, timeout). Use this
// method to create child context when childContext cannot use
// all of parent's deadline but instead there is a need to leave
// some time for parent to do some post-work
func (c *taskListManagerImpl) newChildContext(
	parent context.Context,
	timeout time.Duration,
	tailroom time.Duration,
) (context.Context, context.CancelFunc) {
	select {
	case <-parent.Done():
		return parent, func() {}
	default:
	}
	deadline, ok := parent.Deadline()
	if !ok {
		return context.WithTimeout(parent, timeout)
	}
	remaining := time.Until(deadline) - tailroom
	if remaining < timeout {
		timeout = time.Duration(common.MaxInt64(0, int64(remaining)))
	}
	return context.WithTimeout(parent, timeout)
}

func (c *taskListManagerImpl) isFowardingAllowed(taskList *taskListID, kind types.TaskListKind) bool {
	return !taskList.IsRoot() && kind != types.TaskListKindSticky
}

func (c *taskListManagerImpl) isIsolationMatcherEnabled() bool {
	return c.taskListKind != types.TaskListKindSticky && c.enableIsolation
}

func (c *taskListManagerImpl) shouldReload() bool {
	return c.config.EnableTasklistIsolation() != c.enableIsolation
}

func (c *taskListManagerImpl) getIsolationGroupForTask(ctx context.Context, taskInfo *persistence.TaskInfo) (string, error) {
	if c.enableIsolation && len(taskInfo.PartitionConfig) > 0 && c.taskListKind != types.TaskListKindSticky {
		partitionConfig := make(map[string]string)
		for k, v := range taskInfo.PartitionConfig {
			partitionConfig[k] = v
		}
		partitionConfig[partition.WorkflowIDKey] = taskInfo.WorkflowID
		pollerIsolationGroups := c.config.AllIsolationGroups
		// Not all poller information are available at the time of task list manager creation,
		// because we don't persist poller information in database, so in the first minute, we always assume
		// pollers are available in all isolation groups to avoid the risk of leaking a task to another isolation group.
		// Besides, for sticky and scalable tasklists, not all poller information are available, we also use all isolation group.
		if time.Now().Sub(c.createTime) > time.Minute && c.taskListKind != types.TaskListKindSticky && c.taskListID.IsRoot() {
			pollerIsolationGroups = c.getPollerIsolationGroups()
			if len(pollerIsolationGroups) == 0 {
				// we don't have any pollers, use all isolation groups and wait for pollers' arriving
				pollerIsolationGroups = c.config.AllIsolationGroups
			}
		}
		group, err := c.partitioner.GetIsolationGroupByDomainID(ctx, taskInfo.DomainID, partitionConfig, pollerIsolationGroups)
		if err != nil {
			// For a sticky tasklist, return StickyUnavailableError to let it be added to the non-sticky tasklist.
			if err == partition.ErrNoIsolationGroupsAvailable && c.taskListKind == types.TaskListKindSticky {
				return "", _stickyPollerUnavailableError
			}
			// if we're unable to get the isolation group, log the error and fallback to no isolation
			c.logger.Error("Failed to get isolation group from partition library", tag.WorkflowID(taskInfo.WorkflowID), tag.WorkflowRunID(taskInfo.RunID), tag.TaskID(taskInfo.TaskID), tag.Error(err))
			return defaultTaskBufferIsolationGroup, nil
		}
		c.logger.Debug("get isolation group", tag.PollerGroups(pollerIsolationGroups), tag.IsolationGroup(group), tag.PartitionConfig(partitionConfig))
		// For a sticky tasklist, it is possible that when an isolation group is undrained, the tasks from one workflow is reassigned
		// to the isolation group undrained. If there is no poller from the isolation group, we should return StickyUnavailableError
		// to let the task to be re-enqueued to the non-sticky tasklist. If there is poller, just return an empty isolation group, because
		// there is at most one isolation group for sticky tasklist and we could just use empty isolation group for matching.
		if c.taskListKind == types.TaskListKindSticky {
			pollerIsolationGroups = c.getPollerIsolationGroups()
			for _, pollerGroup := range pollerIsolationGroups {
				if group == pollerGroup {
					return "", nil
				}
			}
			return "", _stickyPollerUnavailableError
		}
		return group, nil
	}
	return defaultTaskBufferIsolationGroup, nil
}

func (c *taskListManagerImpl) getPollerIsolationGroups() []string {
	groupSet := c.pollerHistory.getPollerIsolationGroups(time.Now().Add(-10 * time.Second))
	c.outstandingPollsLock.Lock()
	for _, poller := range c.outstandingPollsMap {
		groupSet[poller.isolationGroup] = struct{}{}
	}
	c.outstandingPollsLock.Unlock()
	result := make([]string, 0, len(groupSet))
	for k := range groupSet {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

func getTaskListTypeTag(taskListType int) metrics.Tag {
	switch taskListType {
	case persistence.TaskListTypeActivity:
		return taskListActivityTypeTag
	case persistence.TaskListTypeDecision:
		return taskListDecisionTypeTag
	default:
		return metrics.TaskListTypeTag("")
	}
}

func createServiceBusyError(msg string) *types.ServiceBusyError {
	return &types.ServiceBusyError{Message: msg}
}

func rangeIDToTaskIDBlock(rangeID, rangeSize int64) taskIDBlock {
	return taskIDBlock{
		start: (rangeID-1)*rangeSize + 1,
		end:   rangeID * rangeSize,
	}
}
