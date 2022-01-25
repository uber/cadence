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

	"github.com/google/uuid"
	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"golang.org/x/time/rate"

	"github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/client/frontend"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
)

type (
	contextKey string
)

const (
	batcherContextKey contextKey = "batcherContext"
	// BatcherTaskListName is the tasklist name
	BatcherTaskListName = "cadence-sys-batcher-tasklist"
	// BatchWFTypeName is the workflow type
	BatchWFTypeName   = "cadence-sys-batch-workflow"
	batchActivityName = "cadence-sys-batch-activity"
	// InfiniteDuration is a long duration(20 yrs) we used for infinite workflow running
	InfiniteDuration = 20 * 365 * 24 * time.Hour
	pageSize         = 1000

	_nonRetriableReason = "non-retriable-error"

	// DefaultRPS is the default RPS
	DefaultRPS = 50
	// DefaultConcurrency is the default concurrency
	DefaultConcurrency = 5
	// DefaultAttemptsOnRetryableError is the default value for AttemptsOnRetryableError
	DefaultAttemptsOnRetryableError = 50
	// DefaultActivityHeartBeatTimeout is the default value for ActivityHeartBeatTimeout
	DefaultActivityHeartBeatTimeout = time.Second * 10
)

const (
	// BatchTypeTerminate is batch type for terminating workflows
	BatchTypeTerminate = "terminate"
	// BatchTypeCancel is the batch type for canceling workflows
	BatchTypeCancel = "cancel"
	// BatchTypeSignal is batch type for signaling workflows
	BatchTypeSignal = "signal"
	// BatchTypeReplicate is batch type for replicating workflows
	BatchTypeReplicate = "replicate"
)

// AllBatchTypes is the batch types we supported
var AllBatchTypes = []string{BatchTypeTerminate, BatchTypeCancel, BatchTypeSignal, BatchTypeReplicate}

type (
	// TerminateParams is the parameters for terminating workflow
	TerminateParams struct {
		// this indicates whether to terminate children workflow. Default to true.
		// TODO https://github.com/uber/cadence/issues/2159
		// Ideally default should be childPolicy of the workflow. But it's currently totally broken.
		TerminateChildren *bool
	}

	// CancelParams is the parameters for canceling workflow
	CancelParams struct {
		// this indicates whether to cancel children workflow. Default to true.
		// TODO https://github.com/uber/cadence/issues/2159
		// Ideally default should be childPolicy of the workflow. But it's currently totally broken.
		CancelChildren *bool
	}

	// SignalParams is the parameters for signaling workflow
	SignalParams struct {
		SignalName string
		Input      string
	}

	// ReplicateParams is the parameters for replicating workflow
	ReplicateParams struct {
		SourceCluster string
		TargetCluster string
	}

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
		// TerminateParams is params only for BatchTypeTerminate
		TerminateParams TerminateParams
		// CancelParams is params only for BatchTypeCancel
		CancelParams CancelParams
		// SignalParams is params only for BatchTypeSignal
		SignalParams SignalParams
		// ReplicateParams is params only for BatchTypeReplicate
		ReplicateParams ReplicateParams
		// RPS of processing. Default to DefaultRPS
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
		execution types.WorkflowExecution
		attempts  int
		// passing along the current heartbeat details to make heartbeat within a task so that it won't timeout
		hbd HeartBeatDetails
	}
)

var (
	batchActivityRetryPolicy = cadence.RetryPolicy{
		InitialInterval:          10 * time.Second,
		BackoffCoefficient:       1.7,
		MaximumInterval:          5 * time.Minute,
		ExpirationInterval:       InfiniteDuration,
		NonRetriableErrorReasons: []string{_nonRetriableReason},
	}

	batchActivityOptions = workflow.ActivityOptions{
		ScheduleToStartTimeout: 5 * time.Minute,
		StartToCloseTimeout:    InfiniteDuration,
		RetryPolicy:            &batchActivityRetryPolicy,
	}
)

func init() {
	workflow.RegisterWithOptions(BatchWorkflow, workflow.RegisterOptions{Name: BatchWFTypeName})
	activity.RegisterWithOptions(BatchActivity, activity.RegisterOptions{Name: batchActivityName})
}

// BatchWorkflow is the workflow that runs a batch job of resetting workflows
func BatchWorkflow(ctx workflow.Context, batchParams BatchParams) (HeartBeatDetails, error) {
	batchParams = setDefaultParams(batchParams)
	err := validateParams(batchParams)
	if err != nil {
		return HeartBeatDetails{}, err
	}
	batchActivityOptions.HeartbeatTimeout = batchParams.ActivityHeartBeatTimeout
	opt := workflow.WithActivityOptions(ctx, batchActivityOptions)
	var result HeartBeatDetails
	err = workflow.ExecuteActivity(opt, batchActivityName, batchParams).Get(ctx, &result)
	return result, err
}

func validateParams(params BatchParams) error {
	if params.BatchType == "" ||
		params.Reason == "" ||
		params.DomainName == "" ||
		params.Query == "" {
		return fmt.Errorf("must provide required parameters: BatchType/Reason/DomainName/Query")
	}
	switch params.BatchType {
	case BatchTypeSignal:
		if params.SignalParams.SignalName == "" {
			return fmt.Errorf("must provide signal name")
		}
		return nil
	case BatchTypeReplicate:
		if params.ReplicateParams.SourceCluster == "" {
			return fmt.Errorf("must provide source cluster")
		}
		if params.ReplicateParams.TargetCluster == "" {
			return fmt.Errorf("must provide target cluster")
		}
		return nil
	case BatchTypeCancel:
		fallthrough
	case BatchTypeTerminate:
		return nil
	default:
		return fmt.Errorf("not supported batch type: %v", params.BatchType)
	}
}

func setDefaultParams(params BatchParams) BatchParams {
	if params.RPS <= 0 {
		params.RPS = DefaultRPS
	}
	if params.Concurrency <= 0 {
		params.Concurrency = DefaultConcurrency
	}
	if params.AttemptsOnRetryableError <= 0 {
		params.AttemptsOnRetryableError = DefaultAttemptsOnRetryableError
	}
	if params.ActivityHeartBeatTimeout <= 0 {
		params.ActivityHeartBeatTimeout = DefaultActivityHeartBeatTimeout
	}
	if len(params.NonRetryableErrors) > 0 {
		params._nonRetryableErrors = make(map[string]struct{}, len(params.NonRetryableErrors))
		for _, estr := range params.NonRetryableErrors {
			params._nonRetryableErrors[estr] = struct{}{}
		}
	}
	if params.TerminateParams.TerminateChildren == nil {
		params.TerminateParams.TerminateChildren = common.BoolPtr(true)
	}
	return params
}

// BatchActivity is activity for processing batch operation
func BatchActivity(ctx context.Context, batchParams BatchParams) (HeartBeatDetails, error) {
	batcher := ctx.Value(batcherContextKey).(*Batcher)
	client := batcher.clientBean.GetFrontendClient()
	var adminClient admin.Client
	if batchParams.BatchType == BatchTypeReplicate {
		currentCluster := batcher.cfg.ClusterMetadata.GetCurrentClusterName()
		if currentCluster != batchParams.ReplicateParams.SourceCluster {
			return HeartBeatDetails{}, cadence.NewCustomError(_nonRetriableReason, fmt.Sprintf("the activity must run in the source cluster, current cluster is %s", currentCluster))
		}
		adminClient = batcher.clientBean.GetRemoteAdminClient(batchParams.ReplicateParams.TargetCluster)
	}

	domainResp, err := client.DescribeDomain(ctx, &types.DescribeDomainRequest{
		Name: &batchParams.DomainName,
	})
	if err != nil {
		return HeartBeatDetails{}, err
	}
	domainID := domainResp.GetDomainInfo().GetUUID()
	hbd := HeartBeatDetails{}
	startOver := true
	if activity.HasHeartbeatDetails(ctx) {
		if err := activity.GetHeartbeatDetails(ctx, &hbd); err == nil {
			startOver = false
		} else {
			batcher := ctx.Value(batcherContextKey).(*Batcher)
			batcher.metricsClient.IncCounter(metrics.BatcherScope, metrics.BatcherProcessorFailures)
			getActivityLogger(ctx).Error("Failed to recover from last heartbeat, start over from beginning", tag.Error(err))
		}
	}

	if startOver {
		resp, err := client.CountWorkflowExecutions(ctx, &types.CountWorkflowExecutionsRequest{
			Domain: batchParams.DomainName,
			Query:  batchParams.Query,
		})
		if err != nil {
			return HeartBeatDetails{}, err
		}
		hbd.TotalEstimate = resp.GetCount()
	}
	rateLimiter := rate.NewLimiter(rate.Limit(batchParams.RPS), batchParams.RPS)
	taskCh := make(chan taskDetail, pageSize)
	respCh := make(chan error, pageSize)
	for i := 0; i < batchParams.Concurrency; i++ {
		go startTaskProcessor(ctx, batchParams, domainID, taskCh, respCh, rateLimiter, client, adminClient)
	}

	for {
		// TODO https://github.com/uber/cadence/issues/2154
		//  Need to improve scan concurrency because it will hold an ES resource until the workflow finishes.
		//  And we can't use list API because terminate / reset will mutate the result.
		resp, err := client.ScanWorkflowExecutions(ctx, &types.ListWorkflowExecutionsRequest{
			Domain:        batchParams.DomainName,
			PageSize:      int32(pageSize),
			NextPageToken: hbd.PageToken,
			Query:         batchParams.Query,
		})
		if err != nil {
			return HeartBeatDetails{}, err
		}
		batchCount := len(resp.Executions)
		if batchCount <= 0 {
			break
		}

		// send all tasks
		for _, wf := range resp.Executions {
			taskCh <- taskDetail{
				execution: *wf.Execution,
				attempts:  0,
				hbd:       hbd,
			}
		}

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
				return HeartBeatDetails{}, ctx.Err()
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

	return hbd, nil
}

func startTaskProcessor(
	ctx context.Context,
	batchParams BatchParams,
	domainID string,
	taskCh chan taskDetail,
	respCh chan error,
	limiter *rate.Limiter,
	client frontend.Client,
	adminClient admin.Client,
) {
	batcher := ctx.Value(batcherContextKey).(*Batcher)
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-taskCh:
			if isDone(ctx) {
				return
			}
			var err error
			requestID := uuid.New().String()

			switch batchParams.BatchType {
			case BatchTypeTerminate:
				err = processTask(ctx, limiter, task, batchParams, client,
					batchParams.TerminateParams.TerminateChildren,
					func(workflowID, runID string) error {
						return client.TerminateWorkflowExecution(ctx, &types.TerminateWorkflowExecutionRequest{
							Domain: batchParams.DomainName,
							WorkflowExecution: &types.WorkflowExecution{
								WorkflowID: workflowID,
								RunID:      runID,
							},
							Reason:   batchParams.Reason,
							Identity: BatchWFTypeName,
						})
					})
			case BatchTypeCancel:
				err = processTask(ctx, limiter, task, batchParams, client,
					batchParams.CancelParams.CancelChildren,
					func(workflowID, runID string) error {
						return client.RequestCancelWorkflowExecution(ctx, &types.RequestCancelWorkflowExecutionRequest{
							Domain: batchParams.DomainName,
							WorkflowExecution: &types.WorkflowExecution{
								WorkflowID: workflowID,
								RunID:      runID,
							},
							Identity:  BatchWFTypeName,
							RequestID: requestID,
						})
					})
			case BatchTypeSignal:
				err = processTask(ctx, limiter, task, batchParams, client, common.BoolPtr(false),
					func(workflowID, runID string) error {
						return client.SignalWorkflowExecution(ctx, &types.SignalWorkflowExecutionRequest{
							Domain: batchParams.DomainName,
							WorkflowExecution: &types.WorkflowExecution{
								WorkflowID: workflowID,
								RunID:      runID,
							},
							Identity:   BatchWFTypeName,
							RequestID:  requestID,
							SignalName: batchParams.SignalParams.SignalName,
							Input:      []byte(batchParams.SignalParams.Input),
						})
					})
			case BatchTypeReplicate:
				err = processTask(ctx, limiter, task, batchParams, client, common.BoolPtr(false),
					func(workflowID, runID string) error {
						return adminClient.ResendReplicationTasks(ctx, &types.ResendReplicationTasksRequest{
							DomainID:      domainID,
							WorkflowID:    workflowID,
							RunID:         runID,
							RemoteCluster: batchParams.ReplicateParams.SourceCluster,
						})
					})
			}
			if err != nil {
				batcher.metricsClient.IncCounter(metrics.BatcherScope, metrics.BatcherProcessorFailures)
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
				batcher.metricsClient.IncCounter(metrics.BatcherScope, metrics.BatcherProcessorSuccess)
				respCh <- nil
			}
		}
	}
}

func processTask(
	ctx context.Context,
	limiter *rate.Limiter,
	task taskDetail,
	batchParams BatchParams,
	client frontend.Client,
	applyOnChild *bool,
	procFn func(string, string) error,
) error {
	wfs := []types.WorkflowExecution{task.execution}
	for len(wfs) > 0 {
		wf := wfs[0]
		wfs = wfs[1:]

		err := limiter.Wait(ctx)
		if err != nil {
			return err
		}
		activity.RecordHeartbeat(ctx, task.hbd)

		err = procFn(wf.GetWorkflowID(), wf.GetRunID())
		if err != nil {
			// EntityNotExistsError means wf is not running or deleted
			if _, ok := err.(*types.EntityNotExistsError); ok {
				continue
			}
			return err
		}
		resp, err := client.DescribeWorkflowExecution(ctx, &types.DescribeWorkflowExecutionRequest{
			Domain: batchParams.DomainName,
			Execution: &types.WorkflowExecution{
				WorkflowID: wf.WorkflowID,
				RunID:      wf.RunID,
			},
		})
		if err != nil {
			// EntityNotExistsError means wf is deleted
			if _, ok := err.(*types.EntityNotExistsError); ok {
				continue
			}
			return err
		}

		// TODO https://github.com/uber/cadence/issues/2159
		// By default should use ChildPolicy, but it is totally broken in Cadence, we need to fix it before using
		if applyOnChild != nil && *applyOnChild && len(resp.PendingChildren) > 0 {
			getActivityLogger(ctx).Info("Found more child workflows to process", tag.Number(int64(len(resp.PendingChildren))))
			for _, ch := range resp.PendingChildren {
				wfs = append(wfs, types.WorkflowExecution{
					WorkflowID: ch.WorkflowID,
					RunID:      ch.RunID,
				})
			}
		}
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
