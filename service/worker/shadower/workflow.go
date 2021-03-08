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

package shadower

import (
	"context"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/.gen/go/shadower"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
)

const (
	defaultScanWorkflowPageSize = 2000

	// NOTE: do not simply change following values as it may result in workflow non-deterministic errors
	defaultMaxReplayConcurrency = 50
	defaultMaxShadowCountPerRun = 100000
)

func register(worker worker.Worker) {
	worker.RegisterWorkflowWithOptions(
		shadowWorkflow,
		workflow.RegisterOptions{Name: shadower.ShadowWorkflowName},
	)
	worker.RegisterActivity(verifyActiveDomainActivity)
}

func shadowWorkflow(
	ctx workflow.Context,
	params shadower.ShadowWorkflowParams,
) (shadower.ShadowWorkflowResult, error) {
	lao := workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: time.Second * 5,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 1.0,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithLocalActivityOptions(ctx, lao)

	var domainActive bool
	if err := workflow.ExecuteLocalActivity(ctx, verifyActiveDomainActivity, params.GetDomain()).Get(ctx, &domainActive); err != nil {
		return shadower.ShadowWorkflowResult{}, err
	}

	// TODO: we probably should make this configurable by user so that they can control is shadowing workflow
	// should be run in active or passive side
	if !domainActive {
		return shadower.ShadowWorkflowResult{}, nil
	}

	concurrency := params.GetConcurrency()
	if concurrency > defaultMaxReplayConcurrency || concurrency == 0 {
		concurrency = defaultMaxReplayConcurrency
	}

	replayStartTime := workflow.Now(ctx)
	workflowTimeout := time.Duration(workflow.GetInfo(ctx).ExecutionStartToCloseTimeoutSeconds) * time.Second
	retryPolicy := &cadence.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2,
		MaximumInterval:    time.Minute,
		ExpirationInterval: workflowTimeout, // retry until workflow timeout
		NonRetriableErrorReasons: []string{
			shadower.ErrReasonDomainNotExists,
			shadower.ErrReasonInvalidQuery,
			shadower.ShadowerTaskList,
		},
	}
	scanWorkflowCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskList:               params.GetTaskList(),
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		RetryPolicy:            retryPolicy,
	})
	replayWorkflowCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskList:               params.GetTaskList(),
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Hour,
		// do not use a short heartbeat timeout here,
		// as replay may take some time if workflow history is large or retrying due to some transient error
		HeartbeatTimeout: 2 * time.Minute,
		RetryPolicy:      retryPolicy,
	})

	var shadowResult shadower.ShadowWorkflowResult
	if params.GetLastRunResult() != nil {
		shadowResult = *params.GetLastRunResult()
	} else {
		shadowResult = shadower.ShadowWorkflowResult{
			Succeeded: common.Int32Ptr(0),
			Skipped:   common.Int32Ptr(0),
			Failed:    common.Int32Ptr(0),
		}
	}
	scanParams := shadower.ScanWorkflowActivityParams{
		Domain:        params.Domain,
		WorkflowQuery: params.WorkflowQuery,
		NextPageToken: params.NextPageToken,
		PageSize:      common.Int32Ptr(defaultScanWorkflowPageSize),
		SamplingRate:  params.SamplingRate,
	}
	for {
		var scanResult shadower.ScanWorkflowActivityResult
		if err := workflow.ExecuteActivity(scanWorkflowCtx, shadower.ScanWorkflowActivityName, scanParams).Get(scanWorkflowCtx, &scanResult); err != nil {
			return shadowResult, err
		}

		replayFutures := make([]workflow.Future, 0, concurrency)
		for _, executions := range splitExecutions(scanResult.Executions, int(concurrency)) {
			replayParams := shadower.ReplayWorkflowActivityParams{
				Domain:     params.Domain,
				Executions: executions,
			}
			future := workflow.ExecuteActivity(replayWorkflowCtx, shadower.ReplayWorkflowActivityName, replayParams)
			replayFutures = append(replayFutures, future)
		}

		for _, future := range replayFutures {
			var replayResult shadower.ReplayWorkflowActivityResult
			if err := future.Get(replayWorkflowCtx, &replayResult); err != nil {
				return shadowResult, err
			}
			*shadowResult.Succeeded += replayResult.GetSucceeded()
			*shadowResult.Skipped += replayResult.GetSkipped()
			*shadowResult.Failed += replayResult.GetFailed()

			if exitConditionMet(ctx, params.GetExitCondition(), replayStartTime, shadowResult) {
				return shadowResult, nil
			}
		}

		scanParams.NextPageToken = scanResult.NextPageToken
		if len(scanParams.NextPageToken) == 0 {
			break
		}

		if shouldContinueAsNew(shadowResult) {
			return shadowResult, getContinueAsNewError(ctx, params, replayStartTime, shadowResult, scanParams.NextPageToken)
		}
	}

	if params.GetShadowMode() == shadower.ShadowModeContinuous {
		return shadowResult, getContinueAsNewError(ctx, params, replayStartTime, shadowResult, nil)
	}

	return shadowResult, nil
}

func splitExecutions(
	executions []*shared.WorkflowExecution,
	concurrency int,
) [][]*shared.WorkflowExecution {
	var result [][]*shared.WorkflowExecution
	size := (len(executions) + concurrency - 1) / concurrency
	for start := 0; start < len(executions); start += size {
		end := start + size
		if end > len(executions) {
			end = len(executions)
		}
		result = append(result, executions[start:end])
	}
	return result
}

func exitConditionMet(
	ctx workflow.Context,
	exitCondition *shadower.ExitCondition,
	startTime time.Time,
	currentResult shadower.ShadowWorkflowResult,
) bool {
	if exitCondition == nil {
		return false
	}

	expirationInterval := time.Duration(exitCondition.GetExpirationIntervalInSeconds()) * time.Second
	if expirationInterval != 0 &&
		workflow.Now(ctx).Sub(startTime) > expirationInterval {
		return true
	}

	shadowCount := exitCondition.GetShadowCount()
	if shadowCount != 0 &&
		currentResult.GetSucceeded()+currentResult.GetFailed() >= shadowCount {
		return true
	}

	return false
}

func shouldContinueAsNew(
	currentResult shadower.ShadowWorkflowResult,
) bool {
	return currentResult.GetSucceeded()+currentResult.GetSkipped()+currentResult.GetFailed() > defaultMaxShadowCountPerRun
}

func getContinueAsNewError(
	ctx workflow.Context,
	params shadower.ShadowWorkflowParams,
	startTime time.Time,
	currentResult shadower.ShadowWorkflowResult,
	nextPageToken []byte,
) error {
	params.NextPageToken = nextPageToken
	if params.GetExitCondition() != nil {
		if expirationInterval := params.ExitCondition.GetExpirationIntervalInSeconds(); expirationInterval != 0 {
			params.ExitCondition.ExpirationIntervalInSeconds = common.Int32Ptr(expirationInterval - int32(workflow.Now(ctx).Sub(startTime).Seconds()))
		}

		if shadowCount := params.ExitCondition.GetShadowCount(); shadowCount != 0 {
			params.ExitCondition.ShadowCount = common.Int32Ptr(shadowCount - (currentResult.GetSucceeded() + currentResult.GetFailed()))
		}
	}

	params.LastRunResult = &currentResult

	return workflow.NewContinueAsNewError(
		ctx,
		shadower.ShadowWorkflowName,
		params,
	)
}

func verifyActiveDomainActivity(
	ctx context.Context,
	domain string,
) (bool, error) {
	worker := ctx.Value(workerContextKey).(*Worker)
	domainCache := worker.domainCache
	domainEntry, err := domainCache.GetDomain(domain)
	if err != nil {
		return false, err
	}

	domainActive := domainEntry.IsDomainActive() || domainEntry.IsDomainPendingActive()
	return domainActive, nil
}
