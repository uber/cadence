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
	"errors"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/.gen/go/shadower"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
)

const (
	defaultScanWorkflowPageSize     = 1000
	defaultSamplingRate             = 1.0
	defaultReplayConcurrency        = 1
	defaultMaxReplayConcurrency     = 50
	defaultMaxShadowCountPerRun     = 20000
	defaultWaitDurationPerIteration = 5 * time.Minute
)

type (
	workflowConfig struct {
		ScanWorkflowPageSize     int32
		DefaultSamplingRate      float64
		DefaultReplayConcurrency int32
		MaxReplayConcurrency     int32
		MaxShadowCountPerRun     int32
		WaitDurationPerIteration time.Duration
	}
)

func register(worker worker.Worker) {
	worker.RegisterWorkflowWithOptions(
		shadowWorkflow,
		workflow.RegisterOptions{Name: shadower.WorkflowName},
	)
}

func shadowWorkflow(
	ctx workflow.Context,
	params shadower.WorkflowParams,
) (shadower.WorkflowResult, error) {
	profile := beginWorkflow(ctx, &params)

	var config workflowConfig
	config, err := getWorkflowConfig(ctx)
	if err != nil {
		return shadower.WorkflowResult{}, profile.endWorkflow(err)
	}

	if err := validateAndFillWorkflowParams(&params, &config); err != nil {
		return shadower.WorkflowResult{}, profile.endWorkflow(err)
	}

	workflowTimeout := time.Duration(workflow.GetInfo(ctx).ExecutionStartToCloseTimeoutSeconds) * time.Second
	retryPolicy := &cadence.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2,
		MaximumInterval:    time.Minute,
		ExpirationInterval: workflowTimeout, // retry until workflow timeout
		NonRetriableErrorReasons: []string{
			shadower.ErrReasonDomainNotExists,
			shadower.ErrReasonInvalidQuery,
			shadower.ErrReasonWorkflowTypeNotRegistered,
			shadower.ErrNonRetryableType, // java non-retryable error type
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
		StartToCloseTimeout:    time.Duration(config.ScanWorkflowPageSize/params.GetConcurrency()+1) * time.Minute,
		// do not use a short heartbeat timeout here,
		// as replay may take some time if workflow history is large or retrying due to some transient errors
		// this is mainly for java, go replay activity can enable auto heartbeating
		HeartbeatTimeout: 2 * time.Minute,
		RetryPolicy:      retryPolicy,
	})

	shadowResult := shadower.WorkflowResult{
		Succeeded: common.Int32Ptr(0),
		Skipped:   common.Int32Ptr(0),
		Failed:    common.Int32Ptr(0),
	}
	scanParams := shadower.ScanWorkflowActivityParams{
		Domain:        params.Domain,
		WorkflowQuery: params.WorkflowQuery,
		NextPageToken: params.NextPageToken,
		PageSize:      common.Int32Ptr(config.ScanWorkflowPageSize),
		SamplingRate:  params.SamplingRate,
	}
	for {
		var scanResult shadower.ScanWorkflowActivityResult
		if err := workflow.ExecuteActivity(scanWorkflowCtx, shadower.ScanWorkflowActivityName, scanParams).Get(scanWorkflowCtx, &scanResult); err != nil {
			return shadower.WorkflowResult{}, profile.endWorkflow(err)
		}

		replayFutures := make([]workflow.Future, 0, params.GetConcurrency())
		for _, executions := range splitExecutions(scanResult.Executions, int(params.GetConcurrency())) {
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
				return shadower.WorkflowResult{}, profile.endWorkflow(err)
			}
			*shadowResult.Succeeded += replayResult.GetSucceeded()
			*shadowResult.Skipped += replayResult.GetSkipped()
			*shadowResult.Failed += replayResult.GetFailed()

			if exitConditionMet(ctx, params.GetExitCondition(), profile.startTime, shadowResult) {
				return combineShadowResults(shadowResult, params.GetLastRunResult()), profile.endWorkflow(nil)
			}
		}

		scanParams.NextPageToken = scanResult.NextPageToken
		if len(scanParams.NextPageToken) == 0 {
			break
		}

		if shouldContinueAsNew(shadowResult, &config) {
			continueAsNewErr := getContinueAsNewError(ctx, params, profile.startTime, params.GetLastRunResult(), shadowResult, scanParams.NextPageToken)
			return shadower.WorkflowResult{}, profile.endWorkflow(continueAsNewErr)
		}
	}

	if params.GetShadowMode() == shadower.ModeContinuous {
		if err := workflow.Sleep(ctx, config.WaitDurationPerIteration); err != nil {
			return shadower.WorkflowResult{}, profile.endWorkflow(err)
		}
		continueAsNewErr := getContinueAsNewError(ctx, params, profile.startTime, params.GetLastRunResult(), shadowResult, nil)
		return shadower.WorkflowResult{}, profile.endWorkflow(continueAsNewErr)
	}

	return combineShadowResults(shadowResult, params.GetLastRunResult()), profile.endWorkflow(nil)
}

func getWorkflowConfig(
	ctx workflow.Context,
) (workflowConfig, error) {
	var config workflowConfig
	if err := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return workflowConfig{
			ScanWorkflowPageSize:     defaultScanWorkflowPageSize,
			DefaultSamplingRate:      defaultSamplingRate,
			DefaultReplayConcurrency: defaultReplayConcurrency,
			MaxReplayConcurrency:     defaultMaxReplayConcurrency,
			MaxShadowCountPerRun:     defaultMaxShadowCountPerRun,
			WaitDurationPerIteration: defaultWaitDurationPerIteration,
		}
	}).Get(&config); err != nil {
		return workflowConfig{}, err
	}
	return config, nil
}

func validateAndFillWorkflowParams(
	params *shadower.WorkflowParams,
	config *workflowConfig,
) error {
	if len(params.GetDomain()) == 0 {
		return errors.New("Domain is not set on shadower workflow params")
	}

	if len(params.GetTaskList()) == 0 {
		return errors.New("TaskList is not set on shadower workflow params")
	}

	if params.GetSamplingRate() == 0 {
		params.SamplingRate = common.Float64Ptr(config.DefaultSamplingRate)
	}

	if params.GetConcurrency() == 0 {
		params.Concurrency = common.Int32Ptr(config.DefaultReplayConcurrency)
	}

	if params.GetConcurrency() > config.MaxReplayConcurrency {
		params.Concurrency = common.Int32Ptr(config.MaxReplayConcurrency)
	}

	return nil
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
	currentResult shadower.WorkflowResult,
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
	currentResult shadower.WorkflowResult,
	config *workflowConfig,
) bool {
	return currentResult.GetSucceeded()+currentResult.GetSkipped()+currentResult.GetFailed() >= config.MaxShadowCountPerRun
}

func getContinueAsNewError(
	ctx workflow.Context,
	params shadower.WorkflowParams,
	startTime time.Time,
	lastRunResult *shadower.WorkflowResult,
	currentResult shadower.WorkflowResult,
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

	combineShadowResults(currentResult, lastRunResult)
	params.LastRunResult = &currentResult

	return workflow.NewContinueAsNewError(
		ctx,
		shadower.WorkflowName,
		params,
	)
}

func combineShadowResults(
	currentResult shadower.WorkflowResult,
	lastRunResult *shadower.WorkflowResult,
) shadower.WorkflowResult {
	if lastRunResult == nil {
		return currentResult
	}

	*currentResult.Succeeded += lastRunResult.GetSucceeded()
	*currentResult.Skipped += lastRunResult.GetSkipped()
	*currentResult.Failed += lastRunResult.GetFailed()
	return currentResult
}
