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
	"testing"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/locks"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	hcommon "github.com/uber/cadence/service/history/common"
	"github.com/uber/cadence/service/history/events"
	"github.com/uber/cadence/service/history/shard"
)

const (
	defaultRemoteCallTimeout = 30 * time.Second
)

type conflictError struct {
	Cause error
}

func (e *conflictError) Error() string {
	return fmt.Sprintf("conditional update failed: %v", e.Cause)
}

func (e *conflictError) Unwrap() error {
	return e.Cause
}

// NewConflictError is only public because it used in workflow/util_test.go
// TODO: refactor those tests
func NewConflictError(_ *testing.T, cause error) error {
	return &conflictError{cause}
}

// IsConflictError checks whether a conflict has occurred while updating a workflow execution
func IsConflictError(err error) bool {
	var e *conflictError
	return errors.As(err, &e)
}

type (
	// Context is the processing context for all operations on workflow execution
	Context interface {
		GetDomainName() string
		GetDomainID() string
		GetExecution() *types.WorkflowExecution

		GetWorkflowExecution() MutableState
		SetWorkflowExecution(mutableState MutableState)
		LoadWorkflowExecution(ctx context.Context) (MutableState, error)
		LoadWorkflowExecutionWithTaskVersion(ctx context.Context, incomingVersion int64) (MutableState, error)
		LoadExecutionStats(ctx context.Context) (*persistence.ExecutionStats, error)
		Clear()

		Lock(ctx context.Context) error
		Unlock()

		GetHistorySize() int64
		SetHistorySize(size int64)

		ReapplyEvents(
			eventBatches []*persistence.WorkflowEvents,
		) error

		PersistStartWorkflowBatchEvents(
			ctx context.Context,
			workflowEvents *persistence.WorkflowEvents,
		) (events.PersistedBlob, error)
		PersistNonStartWorkflowBatchEvents(
			ctx context.Context,
			workflowEvents *persistence.WorkflowEvents,
		) (events.PersistedBlob, error)

		CreateWorkflowExecution(
			ctx context.Context,
			newWorkflow *persistence.WorkflowSnapshot,
			persistedHistory events.PersistedBlob,
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
		UpdateWorkflowExecutionTasks(
			ctx context.Context,
			now time.Time,
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
	domainName, err := c.shard.GetDomainCache().GetDomainName(c.domainID)
	if err != nil {
		return ""
	}
	return domainName
}

func (c *contextImpl) GetHistorySize() int64 {
	return c.stats.HistorySize
}

func (c *contextImpl) SetHistorySize(size int64) {
	c.stats.HistorySize = size
	if c.mutableState != nil {
		c.mutableState.SetHistorySize(size)
	}
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

func (c *contextImpl) LoadWorkflowExecutionWithTaskVersion(
	ctx context.Context,
	incomingVersion int64,
) (MutableState, error) {

	domainEntry, err := c.shard.GetDomainCache().GetDomainByID(c.domainID)
	if err != nil {
		return nil, err
	}

	if c.mutableState == nil {
		response, err := c.getWorkflowExecutionWithRetry(ctx, &persistence.GetWorkflowExecutionRequest{
			DomainID:   c.domainID,
			Execution:  c.workflowExecution,
			DomainName: domainEntry.GetInfo().Name,
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

	flushBeforeReady, err := c.mutableState.StartTransaction(domainEntry, incomingVersion)
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

	flushBeforeReady, err = c.mutableState.StartTransaction(domainEntry, incomingVersion)
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

	// Use empty version to skip incoming task version validation
	return c.LoadWorkflowExecutionWithTaskVersion(ctx, common.EmptyVersion)
}

func (c *contextImpl) CreateWorkflowExecution(
	ctx context.Context,
	newWorkflow *persistence.WorkflowSnapshot,
	persistedHistory events.PersistedBlob,
	createMode persistence.CreateWorkflowMode,
	prevRunID string,
	prevLastWriteVersion int64,
) (retError error) {

	defer func() {
		if retError != nil {
			c.Clear()
		}
	}()
	domainName := c.GetDomainName()
	createRequest := &persistence.CreateWorkflowExecutionRequest{
		// workflow create mode & prev run ID & version
		Mode:                     createMode,
		PreviousRunID:            prevRunID,
		PreviousLastWriteVersion: prevLastWriteVersion,

		NewWorkflowSnapshot: *newWorkflow,
		DomainName:          domainName,
	}

	historySize := int64(len(persistedHistory.Data))
	historySize += c.GetHistorySize()
	c.SetHistorySize(historySize)
	createRequest.NewWorkflowSnapshot.ExecutionStats = &persistence.ExecutionStats{
		HistorySize: historySize,
	}

	resp, err := c.createWorkflowExecutionWithRetry(ctx, createRequest)
	if err != nil {
		if c.isPersistenceTimeoutError(err) {
			c.notifyTasksFromWorkflowSnapshot(newWorkflow, events.PersistedBlobs{persistedHistory}, true)
		}
		return err
	}

	c.notifyTasksFromWorkflowSnapshot(newWorkflow, events.PersistedBlobs{persistedHistory}, false)

	// finally emit session stats
	emitSessionUpdateStats(
		c.metricsClient,
		domainName,
		resp.MutableStateUpdateSessionStats,
	)

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

	var persistedBlobs events.PersistedBlobs
	resetHistorySize := c.GetHistorySize()
	for _, workflowEvents := range resetWorkflowEventsSeq {
		blob, err := c.PersistNonStartWorkflowBatchEvents(ctx, workflowEvents)
		if err != nil {
			return err
		}
		resetHistorySize += int64(len(blob.Data))
		persistedBlobs = append(persistedBlobs, blob)
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
		blob, err := c.PersistStartWorkflowBatchEvents(ctx, startEvents)
		if err != nil {
			return err
		}
		newWorkflowSizeSize += int64(len(blob.Data))
		newContext.SetHistorySize(newWorkflowSizeSize)
		newWorkflow.ExecutionStats = &persistence.ExecutionStats{
			HistorySize: newWorkflowSizeSize,
		}
		persistedBlobs = append(persistedBlobs, blob)
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
			blob, err := c.PersistNonStartWorkflowBatchEvents(ctx, workflowEvents)
			if err != nil {
				return err
			}
			currentWorkflowSize += int64(len(blob.Data))
			persistedBlobs = append(persistedBlobs, blob)
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
	domain, errorDomainName := c.shard.GetDomainCache().GetDomainName(c.domainID)
	if errorDomainName != nil {
		return errorDomainName
	}
	resp, err := c.shard.ConflictResolveWorkflowExecution(ctx, &persistence.ConflictResolveWorkflowExecutionRequest{
		// RangeID , this is set by shard context
		Mode:                    conflictResolveMode,
		ResetWorkflowSnapshot:   *resetWorkflow,
		NewWorkflowSnapshot:     newWorkflow,
		CurrentWorkflowMutation: currentWorkflow,
		// Encoding, this is set by shard context
		DomainName: domain,
	})
	if err != nil {
		if c.isPersistenceTimeoutError(err) {
			c.notifyTasksFromWorkflowSnapshot(resetWorkflow, persistedBlobs, true)
			c.notifyTasksFromWorkflowSnapshot(newWorkflow, persistedBlobs, true)
			c.notifyTasksFromWorkflowMutation(currentWorkflow, persistedBlobs, true)
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

	c.notifyTasksFromWorkflowSnapshot(resetWorkflow, persistedBlobs, false)
	c.notifyTasksFromWorkflowSnapshot(newWorkflow, persistedBlobs, false)
	c.notifyTasksFromWorkflowMutation(currentWorkflow, persistedBlobs, false)

	// finally emit session stats
	domainName := c.GetDomainName()
	emitWorkflowHistoryStats(
		c.metricsClient,
		domainName,
		int(c.stats.HistorySize),
		int(resetMutableState.GetNextEventID()-1),
	)
	emitSessionUpdateStats(
		c.metricsClient,
		domainName,
		resp.MutableStateUpdateSessionStats,
	)
	// emit workflow completion stats if any
	if resetWorkflow.ExecutionInfo.State == persistence.WorkflowStateCompleted {
		if event, err := resetMutableState.GetCompletionEvent(ctx); err == nil {
			workflowType := resetWorkflow.ExecutionInfo.WorkflowTypeName
			taskList := resetWorkflow.ExecutionInfo.TaskList
			emitWorkflowCompletionStats(c.metricsClient, c.logger,
				domainName, workflowType, c.workflowExecution.GetWorkflowID(), c.workflowExecution.GetRunID(),
				taskList, event)
		}
	}

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

func (c *contextImpl) UpdateWorkflowExecutionTasks(
	ctx context.Context,
	now time.Time,
) (retError error) {

	defer func() {
		if retError != nil {
			c.Clear()
		}
	}()

	currentWorkflow, currentWorkflowEventsSeq, err := c.mutableState.CloseTransactionAsMutation(
		now,
		TransactionPolicyPassive,
	)
	if err != nil {
		return err
	}

	if len(currentWorkflowEventsSeq) != 0 {
		return types.InternalServiceError{
			Message: "UpdateWorkflowExecutionTask can only be used for persisting new workflow tasks, but found new history events",
		}
	}
	currentWorkflow.ExecutionStats = &persistence.ExecutionStats{
		HistorySize: c.GetHistorySize(),
	}
	domainName, errorDomainName := c.shard.GetDomainCache().GetDomainName(c.domainID)
	if errorDomainName != nil {
		return errorDomainName
	}
	resp, err := c.updateWorkflowExecutionWithRetry(ctx, &persistence.UpdateWorkflowExecutionRequest{
		// RangeID , this is set by shard context
		Mode:                   persistence.UpdateWorkflowModeIgnoreCurrent,
		UpdateWorkflowMutation: *currentWorkflow,
		// Encoding, this is set by shard context
		DomainName: domainName,
	})
	if err != nil {
		if c.isPersistenceTimeoutError(err) {
			c.notifyTasksFromWorkflowMutation(currentWorkflow, nil, true)
		}
		return err
	}

	// TODO remove updateCondition in favor of condition in mutable state
	c.updateCondition = currentWorkflow.ExecutionInfo.NextEventID

	// notify current workflow tasks
	c.notifyTasksFromWorkflowMutation(currentWorkflow, nil, false)

	emitSessionUpdateStats(
		c.metricsClient,
		c.GetDomainName(),
		resp.MutableStateUpdateSessionStats,
	)

	return nil
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
	var persistedBlobs events.PersistedBlobs
	currentWorkflowSize := c.GetHistorySize()
	oldWorkflowSize := currentWorkflowSize
	currentWorkflowHistoryCount := c.mutableState.GetNextEventID() - 1
	oldWorkflowHistoryCount := currentWorkflowHistoryCount
	for _, workflowEvents := range currentWorkflowEventsSeq {
		blob, err := c.PersistNonStartWorkflowBatchEvents(ctx, workflowEvents)
		currentWorkflowHistoryCount += int64(len(workflowEvents.Events))
		if err != nil {
			return err
		}
		currentWorkflowSize += int64(len(blob.Data))
		persistedBlobs = append(persistedBlobs, blob)
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
		firstEventID := startEvents.Events[0].ID
		var blob events.PersistedBlob
		if firstEventID == common.FirstEventID {
			blob, err = c.PersistStartWorkflowBatchEvents(ctx, startEvents)
			if err != nil {
				return err
			}
		} else {
			// NOTE: This is the case for reset workflow, reset workflow already inserted a branch record
			blob, err = c.PersistNonStartWorkflowBatchEvents(ctx, startEvents)
			if err != nil {
				return err
			}
		}

		persistedBlobs = append(persistedBlobs, blob)
		newWorkflowSizeSize += int64(len(blob.Data))
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
	domain, errorDomainName := c.shard.GetDomainCache().GetDomainName(c.domainID)
	if errorDomainName != nil {
		return errorDomainName
	}
	resp, err := c.updateWorkflowExecutionWithRetry(ctx, &persistence.UpdateWorkflowExecutionRequest{
		// RangeID , this is set by shard context
		Mode:                   updateMode,
		UpdateWorkflowMutation: *currentWorkflow,
		NewWorkflowSnapshot:    newWorkflow,
		// Encoding, this is set by shard context
		DomainName: domain,
	})
	if err != nil {
		if c.isPersistenceTimeoutError(err) {
			c.notifyTasksFromWorkflowMutation(currentWorkflow, persistedBlobs, true)
			c.notifyTasksFromWorkflowSnapshot(newWorkflow, persistedBlobs, true)
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
	c.notifyTasksFromWorkflowMutation(currentWorkflow, persistedBlobs, false)

	// notify new workflow tasks
	c.notifyTasksFromWorkflowSnapshot(newWorkflow, persistedBlobs, false)

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
	c.emitLargeWorkflowShardIDStats(currentWorkflowSize-oldWorkflowSize, oldWorkflowHistoryCount, oldWorkflowSize, currentWorkflowHistoryCount)
	// emit workflow completion stats if any
	if currentWorkflow.ExecutionInfo.State == persistence.WorkflowStateCompleted {
		if event, err := c.mutableState.GetCompletionEvent(ctx); err == nil {
			workflowType := currentWorkflow.ExecutionInfo.WorkflowTypeName
			taskList := currentWorkflow.ExecutionInfo.TaskList
			emitWorkflowCompletionStats(c.metricsClient, c.logger,
				domainName, workflowType, c.workflowExecution.GetWorkflowID(), c.workflowExecution.GetRunID(),
				taskList, event)
		}
	}

	return nil
}

func (c *contextImpl) notifyTasksFromWorkflowSnapshot(
	workflowSnapShot *persistence.WorkflowSnapshot,
	history events.PersistedBlobs,
	persistenceError bool,
) {
	if workflowSnapShot == nil {
		return
	}

	c.notifyTasks(
		workflowSnapShot.ExecutionInfo,
		workflowSnapShot.VersionHistories,
		activityInfosToMap(workflowSnapShot.ActivityInfos),
		workflowSnapShot.TransferTasks,
		workflowSnapShot.TimerTasks,
		workflowSnapShot.CrossClusterTasks,
		workflowSnapShot.ReplicationTasks,
		history,
		persistenceError,
	)
}

func (c *contextImpl) notifyTasksFromWorkflowMutation(
	workflowMutation *persistence.WorkflowMutation,
	history events.PersistedBlobs,
	persistenceError bool,
) {
	if workflowMutation == nil {
		return
	}

	c.notifyTasks(
		workflowMutation.ExecutionInfo,
		workflowMutation.VersionHistories,
		activityInfosToMap(workflowMutation.UpsertActivityInfos),
		workflowMutation.TransferTasks,
		workflowMutation.TimerTasks,
		workflowMutation.CrossClusterTasks,
		workflowMutation.ReplicationTasks,
		history,
		persistenceError,
	)
}

func activityInfosToMap(ais []*persistence.ActivityInfo) map[int64]*persistence.ActivityInfo {
	m := make(map[int64]*persistence.ActivityInfo, len(ais))
	for _, ai := range ais {
		m[ai.ScheduleID] = ai
	}
	return m
}

func (c *contextImpl) notifyTasks(
	executionInfo *persistence.WorkflowExecutionInfo,
	versionHistories *persistence.VersionHistories,
	activities map[int64]*persistence.ActivityInfo,
	transferTasks []persistence.Task,
	timerTasks []persistence.Task,
	crossClusterTasks []persistence.Task,
	replicationTasks []persistence.Task,
	history events.PersistedBlobs,
	persistenceError bool,
) {
	transferTaskInfo := &hcommon.NotifyTaskInfo{
		ExecutionInfo:    executionInfo,
		Tasks:            transferTasks,
		PersistenceError: persistenceError,
	}
	timerTaskInfo := &hcommon.NotifyTaskInfo{
		ExecutionInfo:    executionInfo,
		Tasks:            timerTasks,
		PersistenceError: persistenceError,
	}
	crossClusterTaskInfo := &hcommon.NotifyTaskInfo{
		ExecutionInfo:    executionInfo,
		Tasks:            crossClusterTasks,
		PersistenceError: persistenceError,
	}
	replicationTaskInfo := &hcommon.NotifyTaskInfo{
		ExecutionInfo:    executionInfo,
		Tasks:            replicationTasks,
		VersionHistories: versionHistories,
		Activities:       activities,
		History:          history,
		PersistenceError: persistenceError,
	}

	c.shard.GetEngine().NotifyNewTransferTasks(transferTaskInfo)
	c.shard.GetEngine().NotifyNewTimerTasks(timerTaskInfo)
	c.shard.GetEngine().NotifyNewCrossClusterTasks(crossClusterTaskInfo)
	c.shard.GetEngine().NotifyNewReplicationTasks(replicationTaskInfo)
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

func (c *contextImpl) PersistStartWorkflowBatchEvents(
	ctx context.Context,
	workflowEvents *persistence.WorkflowEvents,
) (events.PersistedBlob, error) {

	if len(workflowEvents.Events) == 0 {
		return events.PersistedBlob{}, &types.InternalServiceError{
			Message: "cannot persist first workflow events with empty events",
		}
	}

	domainID := workflowEvents.DomainID
	domainName, err := c.shard.GetDomainCache().GetDomainName(domainID)
	if err != nil {
		return events.PersistedBlob{}, err
	}
	workflowID := workflowEvents.WorkflowID
	runID := workflowEvents.RunID
	execution := types.WorkflowExecution{
		WorkflowID: workflowEvents.WorkflowID,
		RunID:      workflowEvents.RunID,
	}

	resp, err := c.appendHistoryV2EventsWithRetry(
		ctx,
		domainID,
		execution,
		&persistence.AppendHistoryNodesRequest{
			IsNewBranch: true,
			Info:        persistence.BuildHistoryGarbageCleanupInfo(domainID, workflowID, runID),
			BranchToken: workflowEvents.BranchToken,
			Events:      workflowEvents.Events,
			DomainName:  domainName,
			// TransactionID is set by shard context
		},
	)
	if err != nil {
		return events.PersistedBlob{}, err
	}
	return events.PersistedBlob{
		DataBlob:     resp.DataBlob,
		BranchToken:  workflowEvents.BranchToken,
		FirstEventID: workflowEvents.Events[0].ID,
	}, nil
}

func (c *contextImpl) PersistNonStartWorkflowBatchEvents(
	ctx context.Context,
	workflowEvents *persistence.WorkflowEvents,
) (events.PersistedBlob, error) {

	if len(workflowEvents.Events) == 0 {
		return events.PersistedBlob{}, nil // allow update workflow without events
	}

	domainID := workflowEvents.DomainID
	domainName, err := c.shard.GetDomainCache().GetDomainName(domainID)
	if err != nil {
		return events.PersistedBlob{}, err
	}
	execution := types.WorkflowExecution{
		WorkflowID: workflowEvents.WorkflowID,
		RunID:      workflowEvents.RunID,
	}

	resp, err := c.appendHistoryV2EventsWithRetry(
		ctx,
		domainID,
		execution,
		&persistence.AppendHistoryNodesRequest{
			IsNewBranch: false,
			BranchToken: workflowEvents.BranchToken,
			Events:      workflowEvents.Events,
			DomainName:  domainName,
			// TransactionID is set by shard context
		},
	)
	if err != nil {
		return events.PersistedBlob{}, err
	}
	return events.PersistedBlob{
		DataBlob:     resp.DataBlob,
		BranchToken:  workflowEvents.BranchToken,
		FirstEventID: workflowEvents.Events[0].ID,
	}, nil
}

func (c *contextImpl) appendHistoryV2EventsWithRetry(
	ctx context.Context,
	domainID string,
	execution types.WorkflowExecution,
	request *persistence.AppendHistoryNodesRequest,
) (*persistence.AppendHistoryNodesResponse, error) {

	var resp *persistence.AppendHistoryNodesResponse
	op := func() error {
		var err error
		resp, err = c.shard.AppendHistoryV2Events(ctx, request, domainID, execution)
		return err
	}

	throttleRetry := backoff.NewThrottleRetry(
		backoff.WithRetryPolicy(common.CreatePersistenceRetryPolicy()),
		backoff.WithRetryableError(persistence.IsTransientError),
	)
	err := throttleRetry.Do(ctx, op)
	return resp, err
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

	isRetryable := func(err error) bool {
		if _, ok := err.(*persistence.TimeoutError); ok {
			// TODO: is timeout error retryable for create workflow?
			// if we treat it as retryable, user may receive workflowAlreadyRunning error
			// on the first start workflow execution request.
			return false
		}
		return persistence.IsTransientError(err)
	}

	throttleRetry := backoff.NewThrottleRetry(
		backoff.WithRetryPolicy(common.CreatePersistenceRetryPolicy()),
		backoff.WithRetryableError(isRetryable),
	)
	err := throttleRetry.Do(ctx, op)
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

	throttleRetry := backoff.NewThrottleRetry(
		backoff.WithRetryPolicy(common.CreatePersistenceRetryPolicy()),
		backoff.WithRetryableError(persistence.IsTransientError),
	)
	err := throttleRetry.Do(ctx, op)
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
	//Preparation for the task Validation.
	//metricsClient := c.shard.GetMetricsClient()
	//domainCache := c.shard.GetDomainCache()
	//executionManager := c.shard.GetExecutionManager()
	//historymanager := c.shard.GetHistoryManager()
	//zapLogger, _ := zap.NewProduction()
	//checker, _ := taskvalidator.NewWfChecker(zapLogger, metricsClient, domainCache, executionManager, historymanager)

	isRetryable := func(err error) bool {
		if _, ok := err.(*persistence.TimeoutError); ok {
			// timeout error is not retryable for update workflow execution
			return false
		}
		return persistence.IsTransientError(err)
	}

	throttleRetry := backoff.NewThrottleRetry(
		backoff.WithRetryPolicy(common.CreatePersistenceRetryPolicy()),
		backoff.WithRetryableError(isRetryable),
	)
	err := throttleRetry.Do(ctx, op)
	switch err.(type) {
	case nil:
		return resp, nil
	case *persistence.ConditionFailedError:
		return nil, &conflictError{err}
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
		//TODO: Call the Task Validation here so that it happens whenever an error happen during Update.
		//err1 := checker.WorkflowCheckforValidation(
		//	ctx,
		//	c.workflowExecution.GetWorkflowID(),
		//	c.domainID,
		//	c.GetDomainName(),
		//	c.workflowExecution.GetRunID(),
		//)
		//if err1 != nil {
		//	return nil, err1
		//}
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
		WorkflowID: workflowID,
		RunID:      runID,
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
			DomainName:        domainEntry.GetInfo().Name,
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
		return !IsConflictError(err)
	}
}
