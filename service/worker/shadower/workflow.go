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

	"github.com/uber/cadence/common/types"
)

// The following constants should be identical to those defined in the client repo
// TODO: move those definitions to thrift/proto definition

const (
	scanWorkflowActivityName            = "scanWorkflowActivity"
	replayWorkflowExecutionActivityName = "replayWorkflowExecutionActivity"
)

const (
	errMsgDomainNotExists           = "domain not exists"
	errMsgInvalidQuery              = "invalid visibility query"
	errMsgWorkflowTypeNotRegistered = "workflow type not registered"
)

const (
	shadowWorkflowName = "cadence-shadow-workflow"
)

const (
	defaultScanWorkflowPageSize = 2000

	// NOTE: do not simply change following values as it may result in workflow non-deterministic errors
	defaultMaxReplayConcurrency = 50
	defaultMaxShadowCountPerRun = 100000
)

type (
	shadowWorkflowParams struct {
		Domain   string
		TaskList string

		WorkflowQuery string
		NextPageToken []byte
		SamplingRate  float64

		ShadowMode    shadowMode
		ExitCondition exitCondition

		Concurrency int

		LastRunResult shadowWorkflowResult
	}

	shadowWorkflowResult struct {
		Succeeded int
		Skipped   int
		Failed    int
	}

	scanWorkflowActivityParams struct {
		Domain        string
		WorkflowQuery string
		NextPageToken []byte
		PageSize      int
		SamplingRate  float64
	}

	scanWorkflowActivityResult struct {
		Executions    []types.WorkflowExecution
		NextPageToken []byte
	}

	replayWorkflowActivityParams struct {
		Domain     string
		Executions []types.WorkflowExecution
	}

	replayWorkflowActivityResult struct {
		Succeeded int
		Skipped   int
		Failed    int
		// FailedExecutions []types.WorkflowExecution
	}

	shadowMode int

	exitCondition struct {
		ExpirationInterval time.Duration
		ShadowingCount     int
	}
)

const (
	shadowModeNormal shadowMode = iota + 1
	shadowModeContinuous
)

func register(worker worker.Worker) {
	worker.RegisterWorkflowWithOptions(
		shadowWorkflow,
		workflow.RegisterOptions{Name: shadowWorkflowName},
	)
	worker.RegisterActivity(verifyActiveDomainActivity)
}

func shadowWorkflow(
	ctx workflow.Context,
	params shadowWorkflowParams,
) (shadowWorkflowResult, error) {
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
	if err := workflow.ExecuteLocalActivity(ctx, verifyActiveDomainActivity, params.Domain).Get(ctx, &domainActive); err != nil {
		return shadowWorkflowResult{}, err
	}

	// TODO: we probably should make this configurable by user so that they can control is shadowing workflow
	// should be run in active or passive side
	if !domainActive {
		return shadowWorkflowResult{}, nil
	}

	if params.Concurrency > defaultMaxReplayConcurrency {
		params.Concurrency = defaultMaxReplayConcurrency
	}

	replayStartTime := workflow.Now(ctx)
	workflowTimeout := time.Duration(workflow.GetInfo(ctx).ExecutionStartToCloseTimeoutSeconds) * time.Second
	retryPolicy := &cadence.RetryPolicy{
		InitialInterval:          time.Second,
		BackoffCoefficient:       2,
		MaximumInterval:          time.Minute,
		ExpirationInterval:       workflowTimeout, // retry until workflow timeout
		NonRetriableErrorReasons: []string{errMsgDomainNotExists, errMsgInvalidQuery, errMsgWorkflowTypeNotRegistered},
	}
	scanWorkflowCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskList:               params.TaskList,
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		RetryPolicy:            retryPolicy,
	})
	replayWorkflowCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskList:               params.TaskList,
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Hour,
		// do not use a short heartbeat timeout here,
		// as replay may take some time if workflow history is large or retrying due to some transient error
		HeartbeatTimeout: 2 * time.Minute,
		RetryPolicy:      retryPolicy,
	})

	shadowResult := params.LastRunResult
	scanParams := scanWorkflowActivityParams{
		Domain:        params.Domain,
		WorkflowQuery: params.WorkflowQuery,
		NextPageToken: params.NextPageToken,
		PageSize:      defaultScanWorkflowPageSize,
		SamplingRate:  params.SamplingRate,
	}
	for {
		var scanResult scanWorkflowActivityResult
		if err := workflow.ExecuteActivity(scanWorkflowCtx, scanWorkflowActivityName, scanParams).Get(scanWorkflowCtx, &scanResult); err != nil {
			return shadowResult, err
		}

		replayFutures := make([]workflow.Future, 0, params.Concurrency)
		for _, executions := range splitExecutions(scanResult.Executions, params.Concurrency) {
			replayParams := replayWorkflowActivityParams{
				Domain:     params.Domain,
				Executions: executions,
			}
			future := workflow.ExecuteActivity(replayWorkflowCtx, replayWorkflowExecutionActivityName, replayParams)
			replayFutures = append(replayFutures, future)
		}

		for _, future := range replayFutures {
			var replayResult replayWorkflowActivityResult
			if err := future.Get(replayWorkflowCtx, &replayResult); err != nil {
				return shadowResult, err
			}
			shadowResult.Succeeded += replayResult.Succeeded
			shadowResult.Skipped += replayResult.Skipped
			shadowResult.Failed += replayResult.Failed

			if exitConditionMet(ctx, params.ExitCondition, replayStartTime, shadowResult) {
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

	if params.ShadowMode == shadowModeContinuous {
		return shadowResult, getContinueAsNewError(ctx, params, replayStartTime, shadowResult, nil)
	}

	return shadowResult, nil
}

func splitExecutions(
	executions []types.WorkflowExecution,
	concurrency int,
) [][]types.WorkflowExecution {
	var result [][]types.WorkflowExecution
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
	exitCondition exitCondition,
	startTime time.Time,
	currentResult shadowWorkflowResult,
) bool {
	if exitCondition.ExpirationInterval != 0 &&
		workflow.Now(ctx).Sub(startTime) > exitCondition.ExpirationInterval {
		return true
	}

	if exitCondition.ShadowingCount != 0 &&
		currentResult.Succeeded+currentResult.Failed >= exitCondition.ShadowingCount {
		return true
	}

	return false
}

func shouldContinueAsNew(
	currentResult shadowWorkflowResult,
) bool {
	return currentResult.Succeeded+currentResult.Skipped+currentResult.Failed > defaultMaxShadowCountPerRun
}

func getContinueAsNewError(
	ctx workflow.Context,
	params shadowWorkflowParams,
	startTime time.Time,
	currentResult shadowWorkflowResult,
	nextPageToken []byte,
) error {
	params.NextPageToken = nextPageToken
	if params.ExitCondition.ExpirationInterval != 0 {
		params.ExitCondition.ExpirationInterval -= workflow.Now(ctx).Sub(startTime)
	}
	if params.ExitCondition.ShadowingCount != 0 {
		params.ExitCondition.ShadowingCount -= currentResult.Succeeded + currentResult.Failed
	}
	params.LastRunResult = currentResult

	return workflow.NewContinueAsNewError(
		ctx,
		shadowWorkflowName,
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
