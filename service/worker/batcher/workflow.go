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

package batcher

import (
	"context"
	"fmt"
	"time"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"go.uber.org/cadence"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"golang.org/x/time/rate"
)

const (
	batcherContextKey   = "batcherContext"
	batcherTaskListName = "cadence-sys-batcher-tasklist"
	batchWFTypeName     = "cadence-sys-batch-workflow"
	batchActivityName   = "cadence-sys-batch-activity"

	infiniteDuration = 20 * 365 * 24 * time.Hour
	pageSize         = 1000

	// below are default values for BatchParams
	defaultRPS                      = 50
	defaultConcurrency              = 5
	defaultAttemptsOnRetryableError = 50
	defaultActivityHeartBeatTimeout = time.Minute
)

const (
	BatchTypeTerminate = "terminate"
)

type (
	// BatchParams is the parameters for batch operation workflow
	BatchParams struct {
		// Target domain to execute batch operation
		DomainName string
		// To get the target workflows for processing
		Query string
		// Reason for the operation
		Reason string
		// Supporting: reset,terminate
		BatchType string

		// Below are all optional
		// RPS of processing. Default to defaultRPS
		// TODO we will implement smarter way than this static rate limiter: https://github.com/uber/cadence/issues/2138
		RPS int
		// Number of goroutines running in parallel to process
		Concurrency int
		// Number of attempts for each workflow to process in case of retryable error before giving up
		AttemptsOnRetryableError int
		// timeout for activity heartbeat
		ActivityHeartBeatTimeout time.Duration
		// errors that will not retry which consumes AttemptsOnRetryableError. Default to empty
		NonRetryableErrors []string
		// internal conversion for NonRetryableErrors
		_nonRetryableErrors map[string]struct{}
	}

	// HeartBeatDetails is the struct for heartbeat details
	HeartBeatDetails struct {
		PageToken   []byte
		CurrentPage int
		// This is just an estimation for visibility
		TotalEstimate int64
		// Number of workflows processed successfully
		SuccessCount int
		// Number of workflows that give up due to errors.
		ErrorCount int
	}

	taskDetail struct {
		execution shared.WorkflowExecution
		attempts  int
		// passing along the current heartbeat details to make heartbeat within a task so that it won't timeout
		hbd HeartBeatDetails
	}
)

var (
	batchActivityRetryPolicy = cadence.RetryPolicy{
		InitialInterval:    10 * time.Second,
		BackoffCoefficient: 1.7,
		MaximumInterval:    5 * time.Minute,
		ExpirationInterval: infiniteDuration,
	}

	batchActivityOptions = workflow.ActivityOptions{
		ScheduleToStartTimeout: 5 * time.Minute,
		StartToCloseTimeout:    infiniteDuration,
		RetryPolicy:            &batchActivityRetryPolicy,
	}
)

func init() {
	workflow.RegisterWithOptions(BatchWorkflow, workflow.RegisterOptions{Name: batchWFTypeName})
	activity.RegisterWithOptions(BatchActivity, activity.RegisterOptions{Name: batchActivityName})
}

// BatchWorkflow is the workflow that runs a batch job of resetting workflows
func BatchWorkflow(ctx workflow.Context, batchParams BatchParams) error {
	batchParams, err := setDefaultParams(batchParams)
	if err != nil {
		return err
	}
	batchActivityOptions.HeartbeatTimeout = batchParams.ActivityHeartBeatTimeout
	opt := workflow.WithActivityOptions(ctx, batchActivityOptions)
	return workflow.ExecuteActivity(opt, batchActivityName, batchParams).Get(ctx, nil)
}

func setDefaultParams(params BatchParams) (BatchParams, error) {
	if params.BatchType == "" ||
		params.Reason == "" ||
		params.DomainName == "" ||
		params.Query == "" {
		return BatchParams{}, fmt.Errorf("must provide required parameters: BatchType/Reason/DomainName/Query")
	}
	if params.RPS <= 0 {
		params.RPS = defaultRPS
	}
	if params.Concurrency <= 0 {
		params.Concurrency = defaultConcurrency
	}
	if params.AttemptsOnRetryableError <= 0 {
		params.AttemptsOnRetryableError = defaultAttemptsOnRetryableError
	}
	if params.ActivityHeartBeatTimeout <= 0 {
		params.ActivityHeartBeatTimeout = defaultActivityHeartBeatTimeout
	}
	if len(params.NonRetryableErrors) > 0 {
		params._nonRetryableErrors = make(map[string]struct{}, len(params.NonRetryableErrors))
		for _, estr := range params.NonRetryableErrors {
			params._nonRetryableErrors[estr] = struct{}{}
		}
	}
	switch params.BatchType {
	case BatchTypeTerminate:
		return params, nil
	default:
		return BatchParams{}, fmt.Errorf("not supported batch type: %v", params.BatchType)
	}
}

func BatchActivity(ctx context.Context, batchParams BatchParams) error {
	batcher := ctx.Value(batcherContextKey).(*Batcher)

	hbd := HeartBeatDetails{}
	startOver := true
	if activity.HasHeartbeatDetails(ctx) {
		if err := activity.GetHeartbeatDetails(ctx, &hbd); err == nil {
			startOver = false
		} else {
			getActivityLogger(ctx).Error("Failed to recover from last heartbeat, start over from beginning", tag.Error(err))
		}
	}

	if startOver {
		resp, err := batcher.svcClient.CountWorkflowExecutions(ctx, &shared.CountWorkflowExecutionsRequest{
			Domain: common.StringPtr(batchParams.DomainName),
			Query:  common.StringPtr(batchParams.Query),
		})
		if err != nil {
			return err
		}
		hbd.TotalEstimate = resp.GetCount()
	}
	rateLimiter := rate.NewLimiter(rate.Limit(batchParams.RPS), batchParams.RPS)
	taskCh := make(chan taskDetail, pageSize)
	respCh := make(chan error, pageSize)
	for i := 0; i < batchParams.Concurrency; i++ {
		go startTaskProcessor(ctx, batchParams, taskCh, respCh, rateLimiter)
	}

	for {
		// TODO https://github.com/uber/cadence/issues/2154
		//  Need to improve scan concurrency because it will hold an ES resource until the workflow finishes.
		//  And we can't use list API because terminate / reset will mutate the result.
		resp, err := batcher.svcClient.ScanWorkflowExecutions(ctx, &shared.ListWorkflowExecutionsRequest{
			PageSize:      common.Int32Ptr(int32(pageSize)),
			Domain:        common.StringPtr(batchParams.DomainName),
			NextPageToken: hbd.PageToken,
			Query:         common.StringPtr(batchParams.Query),
		})
		if err != nil {
			return err
		}

		// send all tasks
		for _, wf := range resp.Executions {
			taskCh <- taskDetail{
				execution: *wf.Execution,
				attempts:  0,
				hbd:       hbd,
			}
		}
		batchCount := len(resp.Executions)

		succCount := 0
		errCount := 0
		// wait for counters indicate this batch is done
	Loop:
		for {
			select {
			case err := <-respCh:
				if err == nil {
					succCount++
				} else {
					errCount++
				}
				if succCount+errCount == batchCount {
					break Loop
				}
			case <-ctx.Done():
				return nil
			}
		}

		hbd.CurrentPage++
		hbd.PageToken = resp.NextPageToken
		hbd.SuccessCount += succCount
		hbd.ErrorCount += errCount
		activity.RecordHeartbeat(ctx, hbd)

		if len(hbd.PageToken) == 0 {
			break
		}
	}

	return nil
}

func startTaskProcessor(
	ctx context.Context,
	batchParams BatchParams,
	taskCh chan taskDetail,
	respCh chan error,
	limiter *rate.Limiter) {

	for {
		select {
		case task := <-taskCh:
			if isDone(ctx) {
				return
			}
			var err error

			switch batchParams.BatchType {
			case BatchTypeTerminate:
				err = processOneTerminateTask(ctx, limiter, task, batchParams)
			}
			if err != nil {
				getActivityLogger(ctx).Error("Failed to process batch operation task", tag.Error(err))

				_, ok := batchParams._nonRetryableErrors[err.Error()]
				if ok || task.attempts >= batchParams.AttemptsOnRetryableError {
					respCh <- err
				} else {
					// put back to the channel if less than attemptsOnError
					task.attempts++
					taskCh <- task
				}
			} else {
				respCh <- nil
			}
		}
	}
}

func processOneTerminateTask(ctx context.Context, limiter *rate.Limiter, task taskDetail, param BatchParams) error {
	batcher := ctx.Value(batcherContextKey).(*Batcher)

	wfs := []shared.WorkflowExecution{task.execution}
	for len(wfs) > 0 {
		wf := wfs[0]

		err := limiter.Wait(ctx)
		if err != nil {
			getActivityLogger(ctx).Error("Failed to wait for rateLimiter", tag.Error(err))
		}

		newCtx, cancel := context.WithDeadline(ctx, time.Now().Add(rpcTimeout))
		err = batcher.svcClient.TerminateWorkflowExecution(newCtx, &shared.TerminateWorkflowExecutionRequest{
			Domain:            common.StringPtr(param.DomainName),
			WorkflowExecution: &wf,
			Reason:            common.StringPtr(param.Reason),
			Identity:          common.StringPtr(batchWFTypeName),
		})
		cancel()
		if err != nil {
			_, ok := err.(*shared.EntityNotExistsError)
			if !ok {
				return err
			}
		}
		wfs = wfs[1:]
		newCtx, cancel = context.WithDeadline(ctx, time.Now().Add(rpcTimeout))
		resp, err := batcher.svcClient.DescribeWorkflowExecution(newCtx, &shared.DescribeWorkflowExecutionRequest{
			Domain:    common.StringPtr(param.DomainName),
			Execution: &wf,
		})
		cancel()
		if err != nil {
			return err
		}
		if len(resp.PendingChildren) > 0 {
			getActivityLogger(ctx).Info("Found more child workflows to terminate", tag.Number(int64(len(resp.PendingChildren))))
			for _, ch := range resp.PendingChildren {
				wfs = append(wfs, shared.WorkflowExecution{
					WorkflowId: ch.WorkflowID,
					RunId:      ch.RunID,
				})
			}
		}
		activity.RecordHeartbeat(ctx, task.hbd)
	}

	return nil
}

func isDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func getActivityLogger(ctx context.Context) log.Logger {
	batcher := ctx.Value(batcherContextKey).(*Batcher)
	wfInfo := activity.GetInfo(ctx)
	return batcher.logger.WithTags(
		tag.WorkflowID(wfInfo.WorkflowExecution.ID),
		tag.WorkflowRunID(wfInfo.WorkflowExecution.RunID),
		tag.WorkflowDomainName(wfInfo.WorkflowDomain),
	)
}
