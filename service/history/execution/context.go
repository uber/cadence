// Copyright (c) 2020 Uber Technologies, Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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

//go:generate mockgen -package $GOPACKAGE -source $GOFILE -destination context_mock.go -self_package github.com/uber/cadence/service/history/execution

package execution

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/locks"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/shard"
)

const (
	defaultRemoteCallTimeout = 30 * time.Second
)

var (
	// ErrConflict is exported temporarily for integration test
	ErrConflict = errors.New("conditional update failed")
)

type (
	// Context is the processing context for all operations on workflow execution
	Context interface {
		GetDomainName() string
		GetDomainID() string
		GetExecution() *types.WorkflowExecution

		GetWorkflowExecution() MutableState
		SetWorkflowExecution(mutableState MutableState)
		LoadWorkflowExecution(ctx context.Context) (MutableState, error)
		LoadWorkflowExecutionForReplication(ctx context.Context, incomingVersion int64) (MutableState, error)
		LoadExecutionStats(ctx context.Context) (*persistence.ExecutionStats, error)
		Clear()

		Lock(ctx context.Context) error
		Unlock()

		GetHistorySize() int64
		SetHistorySize(size int64)

		ReapplyEvents(
			eventBatches []*persistence.WorkflowEvents,
		) error

		PersistFirstWorkflowEvents(
			ctx context.Context,
			workflowEvents *persistence.WorkflowEvents,
		) (int64, error)
		PersistNonFirstWorkflowEvents(
			ctx context.Context,
			workflowEvents *persistence.WorkflowEvents,
		) (int64, error)

		CreateWorkflowExecution(
			ctx context.Context,
			newWorkflow *persistence.WorkflowSnapshot,
			historySize int64,
			now time.Time,
			createMode persistence.CreateWorkflowMode,
			prevRunID string,
			prevLastWriteVersion int64,
		) error
		ConflictResolveWorkflowExecution(
			ctx context.Context,
			now time.Time,
			conflictResolveMode persistence.ConflictResolveWorkflowMode,
			resetMutableState MutableState,
			newContext Context,
			newMutableState MutableState,
			currentContext Context,
			currentMutableState MutableState,
			currentTransactionPolicy *TransactionPolicy,
		) error
		UpdateWorkflowExecutionAsActive(
			ctx context.Context,
			now time.Time,
		) error
		UpdateWorkflowExecutionWithNewAsActive(
			ctx context.Context,
			now time.Time,
			newContext Context,
			newMutableState MutableState,
		) error
		UpdateWorkflowExecutionAsPassive(
			ctx context.Context,
			now time.Time,
		) error
		UpdateWorkflowExecutionWithNewAsPassive(
			ctx context.Context,
			now time.Time,
			newContext Context,
			newMutableState MutableState,
		) error
		UpdateWorkflowExecutionWithNew(
			ctx context.Context,
			now time.Time,
			updateMode persistence.UpdateWorkflowMode,
			newContext Context,
			newMutableState MutableState,
			currentWorkflowTransactionPolicy TransactionPolicy,
			newWorkflowTransactionPolicy *TransactionPolicy,
		) error
	}
)

type (
	contextImpl struct {
		domainID          string
		workflowExecution types.WorkflowExecution
		shard             shard.Context
		executionManager  persistence.ExecutionManager
		logger            log.Logger
		metricsClient     metrics.Client

		mutex           locks.Mutex
		mutableState    MutableState
		stats           *persistence.ExecutionStats
		updateCondition int64
	}
)

var _ Context = (*contextImpl)(nil)

var (
	persistenceOperationRetryPolicy = common.CreatePersistenceRetryPolicy()
)

// NewContext creates a new workflow execution context
func NewContext(
	domainID string,
	execution types.WorkflowExecution,
	shard shard.Context,
	executionManager persistence.ExecutionManager,
	logger log.Logger,
) Context {
	return &contextImpl{
		domainID:          domainID,
		workflowExecution: execution,
		shard:             shard,
		executionManager:  executionManager,
		logger:            logger,
		metricsClient:     shard.GetMetricsClient(),
		mutex:             locks.NewMutex(),
		stats: &persistence.ExecutionStats{
			HistorySize: 0,
		},
	}
}

func (c *contextImpl) Lock(ctx context.Context) error {
	return c.mutex.Lock(ctx)
}

func (c *contextImpl) Unlock() {
	c.mutex.Unlock()
}

func (c *contextImpl) Clear() {
	c.metricsClient.IncCounter(metrics.WorkflowContextScope, metrics.WorkflowContextCleared)
	c.mutableState = nil
	c.stats = &persistence.ExecutionStats{
		HistorySize: 0,
	}
}

func (c *contextImpl) GetDomainID() string {
	return c.domainID
}

func (c *contextImpl) GetExecution() *types.WorkflowExecution {
	return &c.workflowExecution
}

func (c *contextImpl) GetDomainName() string {
	domainEntry, err := c.shard.GetDomainCache().GetDomainByID(c.domainID)
	if err != nil {
		return ""
	}
	return domainEntry.GetInfo().Name
}

func (c *contextImpl) GetHistorySize() int64 {
	return c.stats.HistorySize
}

func (c *contextImpl) SetHistorySize(size int64) {
	c.stats.HistorySize = size
}

func (c *contextImpl) LoadExecutionStats(
	ctx context.Context,
) (*persistence.ExecutionStats, error) {
	_, err := c.LoadWorkflowExecution(ctx)
	if err != nil {
		return nil, err
	}
	return c.stats, nil
}

func (c *contextImpl) LoadWorkflowExecutionForReplication(
	ctx context.Context,
	incomingVersion int64,
) (MutableState, error) {

	domainEntry, err := c.shard.GetDomainCache().GetDomainByID(c.domainID)
	if err != nil {
		return nil, err
	}

	if c.mutableState == nil {
		response, err := c.getWorkflowExecutionWithRetry(ctx, &persistence.GetWorkflowExecutionRequest{
			DomainID:  c.domainID,
			Execution: c.workflowExecution,
		})
		if err != nil {
			return nil, err
		}

		c.mutableState = NewMutableStateBuilder(
			c.shard,
			c.logger,
			domainEntry,
		)

		c.mutableState.Load(response.State)

		c.stats = response.State.ExecutionStats
		c.updateCondition = response.State.ExecutionInfo.NextEventID

		// finally emit execution and session stats
		emitWorkflowExecutionStats(
			c.metricsClient,
			c.GetDomainName(),
			response.MutableStateStats,
			c.stats.HistorySize,
		)
	}

	lastWriteVersion, err := c.mutableState.GetLastWriteVersion()
	if err != nil {
		return nil, err
	}

	if lastWriteVersion == incomingVersion {
		err = c.mutableState.StartTransactionSkipDecisionFail(domainEntry)
		if err != nil {
			return nil, err
		}
	} else {
		flushBeforeReady, err := c.mutableState.StartTransaction(domainEntry)
		if err != nil {
			return nil, err
		}
		if !flushBeforeReady {
			return c.mutableState, nil
		}

		if err = c.UpdateWorkflowExecutionAsActive(
			ctx,
			c.shard.GetTimeSource().Now(),
		); err != nil {
			return nil, err
		}

		flushBeforeReady, err = c.mutableState.StartTransaction(domainEntry)
		if err != nil {
			return nil, err
		}
		if flushBeforeReady {
			return nil, &types.InternalServiceError{
				Message: "workflowExecutionContext counter flushBeforeReady status after loading mutable state from DB",
			}
		}
	}
	return c.mutableState, nil
}

// GetWorkflowExecution should only be used in tests
func (c *contextImpl) GetWorkflowExecution() MutableState {
	return c.mutableState
}

// SetWorkflowExecution should only be used in tests
func (c *contextImpl) SetWorkflowExecution(mutableState MutableState) {
	c.mutableState = mutableState
}

func (c *contextImpl) LoadWorkflowExecution(
	ctx context.Context,
) (MutableState, error) {

	domainEntry, err := c.shard.GetDomainCache().GetDomainByID(c.domainID)
	if err != nil {
		return nil, err
	}

	if c.mutableState == nil {
		response, err := c.getWorkflowExecutionWithRetry(ctx, &persistence.GetWorkflowExecutionRequest{
			DomainID:  c.domainID,
			Execution: c.workflowExecution,
		})
		if err != nil {
			return nil, err
		}

		c.mutableState = NewMutableStateBuilder(
			c.shard,
			c.logger,
			domainEntry,
		)

		c.mutableState.Load(response.State)

		c.stats = response.State.ExecutionStats
		c.updateCondition = response.State.ExecutionInfo.NextEventID

		// finally emit execution and session stats
		emitWorkflowExecutionStats(
			c.metricsClient,
			c.GetDomainName(),
			response.MutableStateStats,
			c.stats.HistorySize,
		)
	}

	flushBeforeReady, err := c.mutableState.StartTransaction(domainEntry)
	if err != nil {
		return nil, err
	}
	if !flushBeforeReady {
		return c.mutableState, nil
	}

	if err = c.UpdateWorkflowExecutionAsActive(
		ctx,
		c.shard.GetTimeSource().Now(),
	); err != nil {
		return nil, err
	}

	flushBeforeReady, err = c.mutableState.StartTransaction(domainEntry)
	if err != nil {
		return nil, err
	}
	if flushBeforeReady {
		return nil, &types.InternalServiceError{
			Message: "workflowExecutionContext counter flushBeforeReady status after loading mutable state from DB",
		}
	}

	return c.mutableState, nil
}

func (c *contextImpl) CreateWorkflowExecution(
	ctx context.Context,
	newWorkflow *persistence.WorkflowSnapshot,
	historySize int64,
	now time.Time,
	createMode persistence.CreateWorkflowMode,
	prevRunID string,
	prevLastWriteVersion int64,
) (retError error) {

	defer func() {
		if retError != nil {
			c.Clear()
		}
	}()

	createRequest := &persistence.CreateWorkflowExecutionRequest{
		// workflow create mode & prev run ID & version
		Mode:                     createMode,
		PreviousRunID:            prevRunID,
		PreviousLastWriteVersion: prevLastWriteVersion,

		NewWorkflowSnapshot: *newWorkflow,
	}

	historySize += c.GetHistorySize()
	c.SetHistorySize(historySize)
	createRequest.NewWorkflowSnapshot.ExecutionStats = &persistence.ExecutionStats{
		HistorySize: historySize,
	}

	_, err := c.createWorkflowExecutionWithRetry(ctx, createRequest)
	if err != nil {
		if c.isPersistenceTimeoutError(err) {
			c.notifyTasksFromWorkflowSnapshot(newWorkflow)
		}
		return err
	}

	c.notifyTasksFromWorkflowSnapshot(newWorkflow)

	return nil
}

func (c *contextImpl) ConflictResolveWorkflowExecution(
	ctx context.Context,
	now time.Time,
	conflictResolveMode persistence.ConflictResolveWorkflowMode,
	resetMutableState MutableState,
	newContext Context,
	newMutableState MutableState,
	currentContext Context,
	currentMutableState MutableState,
	currentTransactionPolicy *TransactionPolicy,
) (retError error) {

	defer func() {
		if retError != nil {
			c.Clear()
		}
	}()

	resetWorkflow, resetWorkflowEventsSeq, err := resetMutableState.CloseTransactionAsSnapshot(
		now,
		TransactionPolicyPassive,
	)
	if err != nil {
		return err
	}
	resetHistorySize := c.GetHistorySize()
	for _, workflowEvents := range resetWorkflowEventsSeq {
		eventsSize, err := c.PersistNonFirstWorkflowEvents(ctx, workflowEvents)
		if err != nil {
			return err
		}
		resetHistorySize += eventsSize
	}
	c.SetHistorySize(resetHistorySize)
	resetWorkflow.ExecutionStats = &persistence.ExecutionStats{
		HistorySize: resetHistorySize,
	}

	var newWorkflow *persistence.WorkflowSnapshot
	var newWorkflowEventsSeq []*persistence.WorkflowEvents
	if newContext != nil && newMutableState != nil {

		defer func() {
			if retError != nil {
				newContext.Clear()
			}
		}()

		newWorkflow, newWorkflowEventsSeq, err = newMutableState.CloseTransactionAsSnapshot(
			now,
			TransactionPolicyPassive,
		)
		if err != nil {
			return err
		}
		newWorkflowSizeSize := newContext.GetHistorySize()
		startEvents := newWorkflowEventsSeq[0]
		eventsSize, err := c.PersistFirstWorkflowEvents(ctx, startEvents)
		if err != nil {
			return err
		}
		newWorkflowSizeSize += eventsSize
		newContext.SetHistorySize(newWorkflowSizeSize)
		newWorkflow.ExecutionStats = &persistence.ExecutionStats{
			HistorySize: newWorkflowSizeSize,
		}
	}

	var currentWorkflow *persistence.WorkflowMutation
	var currentWorkflowEventsSeq []*persistence.WorkflowEvents
	if currentContext != nil && currentMutableState != nil && currentTransactionPolicy != nil {

		defer func() {
			if retError != nil {
				currentContext.Clear()
			}
		}()

		currentWorkflow, currentWorkflowEventsSeq, err = currentMutableState.CloseTransactionAsMutation(
			now,
			*currentTransactionPolicy,
		)
		if err != nil {
			return err
		}
		currentWorkflowSize := currentContext.GetHistorySize()
		for _, workflowEvents := range currentWorkflowEventsSeq {
			eventsSize, err := c.PersistNonFirstWorkflowEvents(ctx, workflowEvents)
			if err != nil {
				return err
			}
			currentWorkflowSize += eventsSize
		}
		currentContext.SetHistorySize(currentWorkflowSize)
		currentWorkflow.ExecutionStats = &persistence.ExecutionStats{
			HistorySize: currentWorkflowSize,
		}
	}

	if err := c.conflictResolveEventReapply(
		conflictResolveMode,
		resetWorkflowEventsSeq,
		newWorkflowEventsSeq,
		// current workflow events will not participate in the events reapplication
	); err != nil {
		return err
	}

	if err := c.shard.ConflictResolveWorkflowExecution(ctx, &persistence.ConflictResolveWorkflowExecutionRequest{
		// RangeID , this is set by shard context
		Mode:                    conflictResolveMode,
		ResetWorkflowSnapshot:   *resetWorkflow,
		NewWorkflowSnapshot:     newWorkflow,
		CurrentWorkflowMutation: currentWorkflow,
		// Encoding, this is set by shard context
	}); err != nil {
		if c.isPersistenceTimeoutError(err) {
			c.notifyTasksFromWorkflowSnapshot(resetWorkflow)
			c.notifyTasksFromWorkflowSnapshot(newWorkflow)
			c.notifyTasksFromWorkflowMutation(currentWorkflow)
		}
		return err
	}

	currentBranchToken, err := resetMutableState.GetCurrentBranchToken()
	if err != nil {
		return err
	}

	workflowState, workflowCloseState := resetMutableState.GetWorkflowStateCloseStatus()
	// Current branch changed and notify the watchers
	c.shard.GetEngine().NotifyNewHistoryEvent(events.NewNotification(
		c.domainID,
		&c.workflowExecution,
		resetMutableState.GetLastFirstEventID(),
		resetMutableState.GetNextEventID(),
		resetMutableState.GetPreviousStartedEventID(),
		currentBranchToken,
		workflowState,
		workflowCloseState,
	))

	c.notifyTasksFromWorkflowSnapshot(resetWorkflow)
	c.notifyTasksFromWorkflowSnapshot(newWorkflow)
	c.notifyTasksFromWorkflowMutation(currentWorkflow)

	return nil
}

func (c *contextImpl) UpdateWorkflowExecutionAsActive(
	ctx context.Context,
	now time.Time,
) error {

	return c.UpdateWorkflowExecutionWithNew(
		ctx,
		now,
		persistence.UpdateWorkflowModeUpdateCurrent,
		nil,
		nil,
		TransactionPolicyActive,
		nil,
	)
}

func (c *contextImpl) UpdateWorkflowExecutionWithNewAsActive(
	ctx context.Context,
	now time.Time,
	newContext Context,
	newMutableState MutableState,
) error {

	return c.UpdateWorkflowExecutionWithNew(
		ctx,
		now,
		persistence.UpdateWorkflowModeUpdateCurrent,
		newContext,
		newMutableState,
		TransactionPolicyActive,
		TransactionPolicyActive.Ptr(),
	)
}

func (c *contextImpl) UpdateWorkflowExecutionAsPassive(
	ctx context.Context,
	now time.Time,
) error {

	return c.UpdateWorkflowExecutionWithNew(
		ctx,
		now,
		persistence.UpdateWorkflowModeUpdateCurrent,
		nil,
		nil,
		TransactionPolicyPassive,
		nil,
	)
}

func (c *contextImpl) UpdateWorkflowExecutionWithNewAsPassive(
	ctx context.Context,
	now time.Time,
	newContext Context,
	newMutableState MutableState,
) error {

	return c.UpdateWorkflowExecutionWithNew(
		ctx,
		now,
		persistence.UpdateWorkflowModeUpdateCurrent,
		newContext,
		newMutableState,
		TransactionPolicyPassive,
		TransactionPolicyPassive.Ptr(),
	)
}

func (c *contextImpl) UpdateWorkflowExecutionWithNew(
	ctx context.Context,
	now time.Time,
	updateMode persistence.UpdateWorkflowMode,
	newContext Context,
	newMutableState MutableState,
	currentWorkflowTransactionPolicy TransactionPolicy,
	newWorkflowTransactionPolicy *TransactionPolicy,
) (retError error) {

	defer func() {
		if retError != nil {
			c.Clear()
		}
	}()

	currentWorkflow, currentWorkflowEventsSeq, err := c.mutableState.CloseTransactionAsMutation(
		now,
		currentWorkflowTransactionPolicy,
	)
	if err != nil {
		return err
	}

	currentWorkflowSize := c.GetHistorySize()
	for _, workflowEvents := range currentWorkflowEventsSeq {
		eventsSize, err := c.PersistNonFirstWorkflowEvents(ctx, workflowEvents)
		if err != nil {
			return err
		}
		currentWorkflowSize += eventsSize
	}
	c.SetHistorySize(currentWorkflowSize)
	currentWorkflow.ExecutionStats = &persistence.ExecutionStats{
		HistorySize: currentWorkflowSize,
	}

	var newWorkflow *persistence.WorkflowSnapshot
	var newWorkflowEventsSeq []*persistence.WorkflowEvents
	if newContext != nil && newMutableState != nil && newWorkflowTransactionPolicy != nil {

		defer func() {
			if retError != nil {
				newContext.Clear()
			}
		}()

		newWorkflow, newWorkflowEventsSeq, err = newMutableState.CloseTransactionAsSnapshot(
			now,
			*newWorkflowTransactionPolicy,
		)
		if err != nil {
			return err
		}
		newWorkflowSizeSize := newContext.GetHistorySize()
		startEvents := newWorkflowEventsSeq[0]
		firstEventID := startEvents.Events[0].GetEventID()
		var eventsSize int64
		if firstEventID == common.FirstEventID {
			eventsSize, err = c.PersistFirstWorkflowEvents(ctx, startEvents)
			if err != nil {
				return err
			}
		} else {
			// NOTE: This is the case for reset workflow, reset workflow already inserted a branch record
			eventsSize, err = c.PersistNonFirstWorkflowEvents(ctx, startEvents)
			if err != nil {
				return err
			}
		}

		newWorkflowSizeSize += eventsSize
		newContext.SetHistorySize(newWorkflowSizeSize)
		newWorkflow.ExecutionStats = &persistence.ExecutionStats{
			HistorySize: newWorkflowSizeSize,
		}
	}

	if err := c.mergeContinueAsNewReplicationTasks(
		updateMode,
		currentWorkflow,
		newWorkflow,
	); err != nil {
		return err
	}

	if err := c.updateWorkflowExecutionEventReapply(
		updateMode,
		currentWorkflowEventsSeq,
		newWorkflowEventsSeq,
	); err != nil {
		return err
	}

	resp, err := c.updateWorkflowExecutionWithRetry(ctx, &persistence.UpdateWorkflowExecutionRequest{
		// RangeID , this is set by shard context
		Mode:                   updateMode,
		UpdateWorkflowMutation: *currentWorkflow,
		NewWorkflowSnapshot:    newWorkflow,
		// Encoding, this is set by shard context
	})
	if err != nil {
		if c.isPersistenceTimeoutError(err) {
			c.notifyTasksFromWorkflowMutation(currentWorkflow)
			c.notifyTasksFromWorkflowSnapshot(newWorkflow)
		}
		return err
	}

	// TODO remove updateCondition in favor of condition in mutable state
	c.updateCondition = currentWorkflow.ExecutionInfo.NextEventID

	// for any change in the workflow, send a event
	currentBranchToken, err := c.mutableState.GetCurrentBranchToken()
	if err != nil {
		return err
	}
	workflowState, workflowCloseState := c.mutableState.GetWorkflowStateCloseStatus()
	c.shard.GetEngine().NotifyNewHistoryEvent(events.NewNotification(
		c.domainID,
		&c.workflowExecution,
		c.mutableState.GetLastFirstEventID(),
		c.mutableState.GetNextEventID(),
		c.mutableState.GetPreviousStartedEventID(),
		currentBranchToken,
		workflowState,
		workflowCloseState,
	))

	// notify current workflow tasks
	c.notifyTasksFromWorkflowMutation(currentWorkflow)

	// notify new workflow tasks
	c.notifyTasksFromWorkflowSnapshot(newWorkflow)

	// finally emit session stats
	domainName := c.GetDomainName()
	emitWorkflowHistoryStats(
		c.metricsClient,
		domainName,
		int(c.stats.HistorySize),
		int(c.mutableState.GetNextEventID()-1),
	)
	emitSessionUpdateStats(
		c.metricsClient,
		domainName,
		resp.MutableStateUpdateSessionStats,
	)
	// emit workflow completion stats if any
	if currentWorkflow.ExecutionInfo.State == persistence.WorkflowStateCompleted {
		if event, err := c.mutableState.GetCompletionEvent(ctx); err == nil {
			workflowType := currentWorkflow.ExecutionInfo.WorkflowTypeName
			taskList := currentWorkflow.ExecutionInfo.TaskList
			emitWorkflowCompletionStats(c.metricsClient, domainName, workflowType, taskList, event)
		}
	}

	return nil
}

func (c *contextImpl) notifyTasksFromWorkflowSnapshot(
	workflowSnapShot *persistence.WorkflowSnapshot,
) {
	if workflowSnapShot == nil {
		return
	}

	c.notifyTasks(
		workflowSnapShot.ExecutionInfo,
		workflowSnapShot.TransferTasks,
		workflowSnapShot.TimerTasks,
	)
}

func (c *contextImpl) notifyTasksFromWorkflowMutation(
	workflowMutation *persistence.WorkflowMutation,
) {
	if workflowMutation == nil {
		return
	}

	c.notifyTasks(
		workflowMutation.ExecutionInfo,
		workflowMutation.TransferTasks,
		workflowMutation.TimerTasks,
	)
}

func (c *contextImpl) notifyTasks(
	executionInfo *persistence.WorkflowExecutionInfo,
	transferTasks []persistence.Task,
	timerTasks []persistence.Task,
) {
	c.shard.GetEngine().NotifyNewTransferTasks(executionInfo, transferTasks)
	c.shard.GetEngine().NotifyNewTimerTasks(executionInfo, timerTasks)
}

func (c *contextImpl) mergeContinueAsNewReplicationTasks(
	updateMode persistence.UpdateWorkflowMode,
	currentWorkflowMutation *persistence.WorkflowMutation,
	newWorkflowSnapshot *persistence.WorkflowSnapshot,
) error {

	if currentWorkflowMutation.ExecutionInfo.CloseStatus != persistence.WorkflowCloseStatusContinuedAsNew {
		return nil
	} else if updateMode == persistence.UpdateWorkflowModeBypassCurrent && newWorkflowSnapshot == nil {
		// update current workflow as zombie & continue as new without new zombie workflow
		// this case can be valid if new workflow is already created by resend
		return nil
	}

	// current workflow is doing continue as new

	// it is possible that continue as new is done as part of passive logic
	if len(currentWorkflowMutation.ReplicationTasks) == 0 {
		return nil
	}

	if newWorkflowSnapshot == nil || len(newWorkflowSnapshot.ReplicationTasks) != 1 {
		return &types.InternalServiceError{
			Message: "unable to find replication task from new workflow for continue as new replication",
		}
	}

	// merge the new run first event batch replication task
	// to current event batch replication task
	newRunTask := newWorkflowSnapshot.ReplicationTasks[0].(*persistence.HistoryReplicationTask)
	newWorkflowSnapshot.ReplicationTasks = nil

	newRunBranchToken := newRunTask.BranchToken
	taskUpdated := false
	for _, replicationTask := range currentWorkflowMutation.ReplicationTasks {
		if task, ok := replicationTask.(*persistence.HistoryReplicationTask); ok {
			taskUpdated = true
			task.NewRunBranchToken = newRunBranchToken
		}
	}
	if !taskUpdated {
		return &types.InternalServiceError{
			Message: "unable to find replication task from current workflow for continue as new replication",
		}
	}
	return nil
}

func (c *contextImpl) PersistFirstWorkflowEvents(
	ctx context.Context,
	workflowEvents *persistence.WorkflowEvents,
) (int64, error) {

	if len(workflowEvents.Events) == 0 {
		return 0, &types.InternalServiceError{
			Message: "cannot persist first workflow events with empty events",
		}
	}

	domainID := workflowEvents.DomainID
	workflowID := workflowEvents.WorkflowID
	runID := workflowEvents.RunID
	execution := types.WorkflowExecution{
		WorkflowID: common.StringPtr(workflowEvents.WorkflowID),
		RunID:      common.StringPtr(workflowEvents.RunID),
	}
	branchToken := workflowEvents.BranchToken
	events := workflowEvents.Events

	size, err := c.appendHistoryV2EventsWithRetry(
		ctx,
		domainID,
		execution,
		&persistence.AppendHistoryNodesRequest{
			IsNewBranch: true,
			Info:        persistence.BuildHistoryGarbageCleanupInfo(domainID, workflowID, runID),
			BranchToken: branchToken,
			Events:      events,
			// TransactionID is set by shard context
		},
	)
	return int64(size), err
}

func (c *contextImpl) PersistNonFirstWorkflowEvents(
	ctx context.Context,
	workflowEvents *persistence.WorkflowEvents,
) (int64, error) {

	if len(workflowEvents.Events) == 0 {
		return 0, nil // allow update workflow without events
	}

	domainID := workflowEvents.DomainID
	execution := types.WorkflowExecution{
		WorkflowID: common.StringPtr(workflowEvents.WorkflowID),
		RunID:      common.StringPtr(workflowEvents.RunID),
	}
	branchToken := workflowEvents.BranchToken
	events := workflowEvents.Events

	size, err := c.appendHistoryV2EventsWithRetry(
		ctx,
		domainID,
		execution,
		&persistence.AppendHistoryNodesRequest{
			IsNewBranch: false,
			BranchToken: branchToken,
			Events:      events,
			// TransactionID is set by shard context
		},
	)
	return size, err
}

func (c *contextImpl) appendHistoryV2EventsWithRetry(
	ctx context.Context,
	domainID string,
	execution types.WorkflowExecution,
	request *persistence.AppendHistoryNodesRequest,
) (int64, error) {

	resp := 0
	op := func() error {
		var err error
		resp, err = c.shard.AppendHistoryV2Events(ctx, request, domainID, execution)
		return err
	}

	err := backoff.Retry(
		op,
		persistenceOperationRetryPolicy,
		common.IsPersistenceTransientError,
	)
	return int64(resp), err
}

func (c *contextImpl) createWorkflowExecutionWithRetry(
	ctx context.Context,
	request *persistence.CreateWorkflowExecutionRequest,
) (*persistence.CreateWorkflowExecutionResponse, error) {

	var resp *persistence.CreateWorkflowExecutionResponse
	op := func() error {
		var err error
		resp, err = c.shard.CreateWorkflowExecution(ctx, request)
		return err
	}

	err := backoff.Retry(
		op,
		persistenceOperationRetryPolicy,
		common.IsPersistenceTransientError,
	)
	switch err.(type) {
	case nil:
		return resp, nil
	case *persistence.WorkflowExecutionAlreadyStartedError:
		// it is possible that workflow already exists and caller need to apply
		// workflow ID reuse policy
		return nil, err
	default:
		c.logger.Error(
			"Persistent store operation failure",
			tag.WorkflowID(c.workflowExecution.GetWorkflowID()),
			tag.WorkflowRunID(c.workflowExecution.GetRunID()),
			tag.WorkflowDomainID(c.domainID),
			tag.StoreOperationCreateWorkflowExecution,
			tag.Error(err),
		)
		return nil, err
	}
}

func (c *contextImpl) getWorkflowExecutionWithRetry(
	ctx context.Context,
	request *persistence.GetWorkflowExecutionRequest,
) (*persistence.GetWorkflowExecutionResponse, error) {

	var resp *persistence.GetWorkflowExecutionResponse
	op := func() error {
		var err error
		resp, err = c.executionManager.GetWorkflowExecution(ctx, request)

		return err
	}

	err := backoff.Retry(
		op,
		persistenceOperationRetryPolicy,
		common.IsPersistenceTransientError,
	)
	switch err.(type) {
	case nil:
		return resp, nil
	case *types.EntityNotExistsError:
		// it is possible that workflow does not exists
		return nil, err
	default:
		c.logger.Error(
			"Persistent fetch operation failure",
			tag.WorkflowID(c.workflowExecution.GetWorkflowID()),
			tag.WorkflowRunID(c.workflowExecution.GetRunID()),
			tag.WorkflowDomainID(c.domainID),
			tag.StoreOperationGetWorkflowExecution,
			tag.Error(err),
		)
		return nil, err
	}
}

func (c *contextImpl) updateWorkflowExecutionWithRetry(
	ctx context.Context,
	request *persistence.UpdateWorkflowExecutionRequest,
) (*persistence.UpdateWorkflowExecutionResponse, error) {

	var resp *persistence.UpdateWorkflowExecutionResponse
	op := func() error {
		var err error
		resp, err = c.shard.UpdateWorkflowExecution(ctx, request)
		return err
	}

	err := backoff.Retry(
		op, persistenceOperationRetryPolicy,
		common.IsPersistenceTransientError,
	)
	switch err.(type) {
	case nil:
		return resp, nil
	case *persistence.ConditionFailedError:
		// TODO get rid of ErrConflict
		return nil, ErrConflict
	default:
		c.logger.Error(
			"Persistent store operation failure",
			tag.WorkflowID(c.workflowExecution.GetWorkflowID()),
			tag.WorkflowRunID(c.workflowExecution.GetRunID()),
			tag.WorkflowDomainID(c.domainID),
			tag.StoreOperationUpdateWorkflowExecution,
			tag.Error(err),
			tag.Number(c.updateCondition),
		)
		return nil, err
	}
}

func (c *contextImpl) updateWorkflowExecutionEventReapply(
	updateMode persistence.UpdateWorkflowMode,
	eventBatch1 []*persistence.WorkflowEvents,
	eventBatch2 []*persistence.WorkflowEvents,
) error {

	if updateMode != persistence.UpdateWorkflowModeBypassCurrent {
		return nil
	}

	var eventBatches []*persistence.WorkflowEvents
	eventBatches = append(eventBatches, eventBatch1...)
	eventBatches = append(eventBatches, eventBatch2...)
	return c.ReapplyEvents(eventBatches)
}

func (c *contextImpl) conflictResolveEventReapply(
	conflictResolveMode persistence.ConflictResolveWorkflowMode,
	eventBatch1 []*persistence.WorkflowEvents,
	eventBatch2 []*persistence.WorkflowEvents,
) error {

	if conflictResolveMode != persistence.ConflictResolveWorkflowModeBypassCurrent {
		return nil
	}

	var eventBatches []*persistence.WorkflowEvents
	eventBatches = append(eventBatches, eventBatch1...)
	eventBatches = append(eventBatches, eventBatch2...)
	return c.ReapplyEvents(eventBatches)
}

func (c *contextImpl) ReapplyEvents(
	eventBatches []*persistence.WorkflowEvents,
) error {

	// NOTE: this function should only be used to workflow which is
	// not the caller, or otherwise deadlock will appear

	if len(eventBatches) == 0 {
		return nil
	}

	domainID := eventBatches[0].DomainID
	workflowID := eventBatches[0].WorkflowID
	runID := eventBatches[0].RunID
	domainCache := c.shard.GetDomainCache()
	clientBean := c.shard.GetService().GetClientBean()
	serializer := c.shard.GetService().GetPayloadSerializer()
	domainEntry, err := domainCache.GetDomainByID(domainID)
	if err != nil {
		return err
	}
	if domainEntry.IsDomainPendingActive() {
		return nil
	}
	var reapplyEvents []*types.HistoryEvent
	for _, events := range eventBatches {
		if events.DomainID != domainID ||
			events.WorkflowID != workflowID {
			return &types.InternalServiceError{
				Message: "workflowExecutionContext encounter mismatch domainID / workflowID in events reapplication.",
			}
		}

		for _, event := range events.Events {
			switch event.GetEventType() {
			case types.EventTypeWorkflowExecutionSignaled:
				reapplyEvents = append(reapplyEvents, event)
			}
		}
	}
	if len(reapplyEvents) == 0 {
		return nil
	}

	// Reapply events only reapply to the current run.
	// The run id is only used for reapply event de-duplication
	execution := &types.WorkflowExecution{
		WorkflowID: common.StringPtr(workflowID),
		RunID:      common.StringPtr(runID),
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultRemoteCallTimeout)
	defer cancel()

	activeCluster := domainEntry.GetReplicationConfig().ActiveClusterName
	if activeCluster == c.shard.GetClusterMetadata().GetCurrentClusterName() {
		return c.shard.GetEngine().ReapplyEvents(
			ctx,
			domainID,
			workflowID,
			runID,
			reapplyEvents,
		)
	}

	// The active cluster of the domain is the same as current cluster.
	// Use the history from the same cluster to reapply events
	reapplyEventsDataBlob, err := serializer.SerializeBatchEvents(
		reapplyEvents,
		common.EncodingTypeThriftRW,
	)
	if err != nil {
		return err
	}
	// The active cluster of the domain is differ from the current cluster
	// Use frontend client to route this request to the active cluster
	// Reapplication only happens in active cluster
	sourceCluster := clientBean.GetRemoteAdminClient(activeCluster)
	if sourceCluster == nil {
		return &types.InternalServiceError{
			Message: fmt.Sprintf("cannot find cluster config %v to do reapply", activeCluster),
		}
	}
	return sourceCluster.ReapplyEvents(
		ctx,
		&types.ReapplyEventsRequest{
			DomainName:        common.StringPtr(domainEntry.GetInfo().Name),
			WorkflowExecution: execution,
			Events:            reapplyEventsDataBlob.ToInternal(),
		},
	)
}

func (c *contextImpl) isPersistenceTimeoutError(
	err error,
) bool {
	// TODO: ideally we only need to check if err has type *persistence.Timeout,
	// but currently only cassandra will return timeout error of that type.
	// so currently this method will return false positives
	switch err.(type) {
	case nil:
		return false
	case *types.WorkflowExecutionAlreadyStartedError,
		*persistence.WorkflowExecutionAlreadyStartedError,
		*persistence.CurrentWorkflowConditionFailedError,
		*persistence.ConditionFailedError,
		*types.ServiceBusyError,
		*types.LimitExceededError,
		*persistence.ShardOwnershipLostError:
		return false
	case *persistence.TimeoutError:
		return true
	default:
		return err != ErrConflict
	}
}
