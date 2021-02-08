// Copyright (c) 2017-2021 Uber Technologies, Inc.
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

package history

import (
	"context"
	"time"

	"github.com/pborman/uuid"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
)

type UpdateWorkflowAction struct {
	Noop           bool
	CreateDecision bool
}

var (
	UpdateWorkflowWithNewDecision = &UpdateWorkflowAction{
		CreateDecision: true,
	}
	UpdateWorkflowWithoutDecision = &UpdateWorkflowAction{
		CreateDecision: false,
	}
)

type UpdateWorkflowActionFunc func(execution.Context, execution.MutableState) (*UpdateWorkflowAction, error)

func LoadWorkflowOnce(
	ctx context.Context,
	cache *execution.Cache,
	domainID string,
	workflowID string,
	runID string,
) (WorkflowContext, error) {

	wfContext, release, err := cache.GetOrCreateWorkflowExecution(
		ctx,
		domainID,
		types.WorkflowExecution{
			WorkflowID: workflowID,
			RunID:      runID,
		},
	)
	if err != nil {
		return nil, err
	}

	mutableState, err := wfContext.LoadWorkflowExecution(ctx)
	if err != nil {
		release(err)
		return nil, err
	}

	return NewWorkflowContext(wfContext, release, mutableState), nil
}

func LoadWorkflow(
	ctx context.Context,
	cache *execution.Cache,
	executionManager persistence.ExecutionManager,
	domainID string,
	workflowID string,
	runID string,
) (WorkflowContext, error) {

	if runID != "" {
		return LoadWorkflowOnce(ctx, cache, domainID, workflowID, runID)
	}

	for attempt := 0; attempt < conditionalRetryCount; attempt++ {

		workflowContext, err := LoadWorkflowOnce(ctx, cache, domainID, workflowID, "")
		if err != nil {
			return nil, err
		}

		if workflowContext.GetMutableState().IsWorkflowExecutionRunning() {
			return workflowContext, nil
		}

		// workflow not running, need to check current record
		resp, err := executionManager.GetCurrentExecution(
			ctx,
			&persistence.GetCurrentExecutionRequest{
				DomainID:   domainID,
				WorkflowID: workflowID,
			},
		)
		if err != nil {
			workflowContext.GetReleaseFn()(err)
			return nil, err
		}

		if resp.RunID == workflowContext.GetRunID() {
			return workflowContext, nil
		}
		workflowContext.GetReleaseFn()(nil)
	}

	return nil, &types.InternalServiceError{Message: "unable to locate current workflow execution"}
}

func updateWorkflowHelper(
	ctx context.Context,
	workflowContext WorkflowContext,
	now time.Time,
	action UpdateWorkflowActionFunc,
) (retError error) {

UpdateHistoryLoop:
	for attempt := 0; attempt < conditionalRetryCount; attempt++ {
		wfContext := workflowContext.GetContext()
		mutableState := workflowContext.GetMutableState()

		// conduct caller action
		postActions, err := action(wfContext, mutableState)
		if err != nil {
			if err == ErrStaleState {
				// Handler detected that cached workflow mutable could potentially be stale
				// Reload workflow execution history
				workflowContext.GetContext().Clear()
				if attempt != conditionalRetryCount-1 {
					_, err = workflowContext.ReloadMutableState(ctx)
					if err != nil {
						return err
					}
				}
				continue UpdateHistoryLoop
			}

			// Returned error back to the caller
			return err
		}
		if postActions.Noop {
			return nil
		}

		if postActions.CreateDecision {
			// Create a transfer task to schedule a decision task
			if !mutableState.HasPendingDecision() {
				_, err := mutableState.AddDecisionTaskScheduledEvent(false)
				if err != nil {
					return &types.InternalServiceError{Message: "Failed to add decision scheduled event."}
				}
			}
		}

		err = workflowContext.GetContext().UpdateWorkflowExecutionAsActive(ctx, now)
		if err == execution.ErrConflict {
			if attempt != conditionalRetryCount-1 {
				_, err = workflowContext.ReloadMutableState(ctx)
				if err != nil {
					return err
				}
			}
			continue UpdateHistoryLoop
		}
		return err
	}
	return ErrMaxAttemptsExceeded
}

func UpdateWorkflowExecutionWithAction(
	ctx context.Context,
	cache *execution.Cache,
	domainID string,
	execution types.WorkflowExecution,
	now time.Time,
	action UpdateWorkflowActionFunc,
) (retError error) {

	workflowContext, err := LoadWorkflowOnce(ctx, cache, domainID, execution.GetWorkflowID(), execution.GetRunID())
	if err != nil {
		return err
	}
	defer func() { workflowContext.GetReleaseFn()(retError) }()

	return updateWorkflowHelper(ctx, workflowContext, now, action)
}

func UpdateWorkflow(
	ctx context.Context,
	cache *execution.Cache,
	executionManager persistence.ExecutionManager,
	domainID string,
	execution types.WorkflowExecution,
	now time.Time,
	action UpdateWorkflowActionFunc,
) (retError error) {

	workflowContext, err := LoadWorkflow(ctx, cache, executionManager, domainID, execution.GetWorkflowID(), execution.GetRunID())
	if err != nil {
		return err
	}
	defer func() { workflowContext.GetReleaseFn()(retError) }()

	return updateWorkflowHelper(ctx, workflowContext, now, action)
}

// TODO: remove and use updateWorkflowExecutionWithAction
func UpdateWorkflowExecution(
	ctx context.Context,
	cache *execution.Cache,
	domainID string,
	execution types.WorkflowExecution,
	createDecisionTask bool,
	now time.Time,
	action func(wfContext execution.Context, mutableState execution.MutableState) error,
) error {

	return UpdateWorkflowExecutionWithAction(
		ctx,
		cache,
		domainID,
		execution,
		now,
		getUpdateWorkflowActionFunc(createDecisionTask, action),
	)
}

func getUpdateWorkflowActionFunc(
	createDecisionTask bool,
	action func(wfContext execution.Context, mutableState execution.MutableState) error,
) UpdateWorkflowActionFunc {

	return func(wfContext execution.Context, mutableState execution.MutableState) (*UpdateWorkflowAction, error) {
		err := action(wfContext, mutableState)
		if err != nil {
			return nil, err
		}
		postActions := &UpdateWorkflowAction{
			CreateDecision: createDecisionTask,
		}
		return postActions, nil
	}
}

func ValidateDomainUUID(
	domainUUID string,
) (string, error) {

	if domainUUID == "" {
		return "", &types.BadRequestError{Message: "Missing domain UUID."}
	} else if uuid.Parse(domainUUID) == nil {
		return "", &types.BadRequestError{Message: "Invalid domain UUID."}
	}
	return domainUUID, nil
}

func GetActiveDomainEntry(
	shard shard.Context,
	domainUUID string,
) (*cache.DomainCacheEntry, error) {

	domainID, err := ValidateDomainUUID(domainUUID)
	if err != nil {
		return nil, err
	}

	domainEntry, err := shard.GetDomainCache().GetDomainByID(domainID)
	if err != nil {
		return nil, err
	}
	if err = domainEntry.GetDomainNotActiveErr(); err != nil {
		return domainEntry, err
	}
	return domainEntry, nil
}

func GetPendingActiveDomainEntry(
	shard shard.Context,
	domainUUID string,
) (bool, error) {

	domainID, err := ValidateDomainUUID(domainUUID)
	if err != nil {
		return false, err
	}

	domainEntry, err := shard.GetDomainCache().GetDomainByID(domainID)
	if err != nil {
		return false, err
	}

	return domainEntry.IsDomainPendingActive(), nil
}
