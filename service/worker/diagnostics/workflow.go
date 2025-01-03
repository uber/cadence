// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

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

package diagnostics

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/worker/diagnostics/invariant"
	"github.com/uber/cadence/service/worker/diagnostics/invariant/failure"
	"github.com/uber/cadence/service/worker/diagnostics/invariant/retry"
	"github.com/uber/cadence/service/worker/diagnostics/invariant/timeout"
)

const (
	diagnosticsWorkflow = "diagnostics-workflow"
	tasklist            = "diagnostics-wf-tasklist"

	retrieveWfExecutionHistoryActivity = "retrieveWfExecutionHistory"
	identifyIssuesActivity             = "identifyIssues"
	rootCauseIssuesActivity            = "rootCauseIssues"
)

type DiagnosticsWorkflowInput struct {
	Domain     string
	WorkflowID string
	RunID      string
}

type DiagnosticsWorkflowResult struct {
	Timeouts *timeoutDiagnostics
	Failures *failureDiagnostics
	Retries  *retryDiagnostics
}

type timeoutDiagnostics struct {
	Issues    []*timeoutIssuesResult
	RootCause []*timeoutRootCauseResult
	Runbooks  []string
}

type timeoutIssuesResult struct {
	InvariantType    string
	Reason           string
	ExecutionTimeout *timeout.ExecutionTimeoutMetadata
	ActivityTimeout  *timeout.ActivityTimeoutMetadata
	ChildWfTimeout   *timeout.ChildWfTimeoutMetadata
	DecisionTimeout  *timeout.DecisionTimeoutMetadata
}

type timeoutRootCauseResult struct {
	RootCauseType        string
	PollersMetadata      *timeout.PollersMetadata
	HeartBeatingMetadata *timeout.HeartbeatingMetadata
}

type failureDiagnostics struct {
	Issues    []*failureIssuesResult
	RootCause []*failureRootCauseResult
	Runbooks  []string
}

type failureIssuesResult struct {
	InvariantType string
	Reason        string
	Metadata      *failure.FailureMetadata
}

type failureRootCauseResult struct {
	RootCauseType    string
	BlobSizeMetadata *failure.BlobSizeMetadata
}

type retryDiagnostics struct {
	Issues   []*retryIssuesResult
	Runbooks []string
}

type retryIssuesResult struct {
	InvariantType string
	Reason        string
	Metadata      retry.RetryMetadata
}

func (w *dw) DiagnosticsWorkflow(ctx workflow.Context, params DiagnosticsWorkflowInput) (*DiagnosticsWorkflowResult, error) {
	scope := w.metricsClient.Scope(metrics.DiagnosticsWorkflowScope, metrics.DomainTag(params.Domain))
	scope.IncCounter(metrics.DiagnosticsWorkflowStartedCount)
	sw := scope.StartTimer(metrics.DiagnosticsWorkflowExecutionLatency)
	defer sw.Stop()

	var timeoutsResult timeoutDiagnostics
	var failureResult failureDiagnostics
	var retryResult retryDiagnostics
	var checkResult []invariant.InvariantCheckResult
	var rootCauseResult []invariant.InvariantRootCauseResult

	activityOptions := workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Second * 10,
		ScheduleToStartTimeout: time.Second * 5,
		StartToCloseTimeout:    time.Second * 5,
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOptions)

	var wfExecutionHistory *types.GetWorkflowExecutionHistoryResponse
	err := workflow.ExecuteActivity(activityCtx, w.retrieveExecutionHistory, retrieveExecutionHistoryInputParams{
		Domain: params.Domain,
		Execution: &types.WorkflowExecution{
			WorkflowID: params.WorkflowID,
			RunID:      params.RunID,
		}}).Get(ctx, &wfExecutionHistory)
	if err != nil {
		return nil, fmt.Errorf("RetrieveExecutionHistory: %w", err)
	}

	timeoutsResult.Runbooks = []string{linkToTimeoutsRunbook}

	err = workflow.ExecuteActivity(activityCtx, w.identifyIssues, identifyIssuesParams{
		History: wfExecutionHistory,
		Domain:  params.Domain,
	}).Get(ctx, &checkResult)
	if err != nil {
		return nil, fmt.Errorf("IdentifyIssues: %w", err)
	}

	err = workflow.ExecuteActivity(activityCtx, w.rootCauseIssues, rootCauseIssuesParams{
		History: wfExecutionHistory,
		Domain:  params.Domain,
		Issues:  checkResult,
	}).Get(ctx, &rootCauseResult)
	if err != nil {
		return nil, fmt.Errorf("RootCauseIssues: %w", err)
	}

	timeoutIssues, err := retrieveTimeoutIssues(checkResult)
	if err != nil {
		return nil, fmt.Errorf("RetrieveTimeoutIssues: %w", err)
	}
	timeoutsResult.Issues = timeoutIssues

	timeoutRootCause, err := retrieveTimeoutRootCause(rootCauseResult)
	if err != nil {
		return nil, fmt.Errorf("RetrieveTimeoutRootCause: %w", err)
	}
	timeoutsResult.RootCause = timeoutRootCause

	failureIssues, err := retrieveFailureIssues(checkResult)
	if err != nil {
		return nil, fmt.Errorf("RetrieveFailureIssues: %w", err)
	}
	failureResult.Issues = failureIssues

	failureRootCause, err := retrieveFailureRootCause(rootCauseResult)
	if err != nil {
		return nil, fmt.Errorf("RetrieveFailureRootCause: %w", err)
	}
	failureResult.RootCause = failureRootCause
	failureResult.Runbooks = []string{linkToFailuresRunbook}

	retryIssues, err := retrieveRetryIssues(checkResult)
	if err != nil {
		return nil, fmt.Errorf("RetrieveRetryIssues: %w", err)
	}
	retryResult.Issues = retryIssues
	retryResult.Runbooks = []string{linkToRetriesRunbook}

	scope.IncCounter(metrics.DiagnosticsWorkflowSuccess)
	return &DiagnosticsWorkflowResult{
		Timeouts: &timeoutsResult,
		Failures: &failureResult,
		Retries:  &retryResult,
	}, nil
}

func retrieveTimeoutIssues(issues []invariant.InvariantCheckResult) ([]*timeoutIssuesResult, error) {
	result := make([]*timeoutIssuesResult, 0)
	for _, issue := range issues {
		switch issue.InvariantType {
		case timeout.TimeoutTypeExecution.String():
			var metadata timeout.ExecutionTimeoutMetadata
			err := json.Unmarshal(issue.Metadata, &metadata)
			if err != nil {
				return nil, err
			}
			result = append(result, &timeoutIssuesResult{
				InvariantType:    issue.InvariantType,
				Reason:           issue.Reason,
				ExecutionTimeout: &metadata,
			})
		case timeout.TimeoutTypeActivity.String():
			var metadata timeout.ActivityTimeoutMetadata
			err := json.Unmarshal(issue.Metadata, &metadata)
			if err != nil {
				return nil, err
			}
			result = append(result, &timeoutIssuesResult{
				InvariantType:   issue.InvariantType,
				Reason:          issue.Reason,
				ActivityTimeout: &metadata,
			})
		case timeout.TimeoutTypeChildWorkflow.String():
			var metadata timeout.ChildWfTimeoutMetadata
			err := json.Unmarshal(issue.Metadata, &metadata)
			if err != nil {
				return nil, err
			}
			result = append(result, &timeoutIssuesResult{
				InvariantType:  issue.InvariantType,
				Reason:         issue.Reason,
				ChildWfTimeout: &metadata,
			})
		case timeout.TimeoutTypeDecision.String():
			var metadata timeout.DecisionTimeoutMetadata
			err := json.Unmarshal(issue.Metadata, &metadata)
			if err != nil {
				return nil, err
			}
			result = append(result, &timeoutIssuesResult{
				InvariantType:   issue.InvariantType,
				Reason:          issue.Reason,
				DecisionTimeout: &metadata,
			})
		}
	}
	return result, nil
}

func retrieveTimeoutRootCause(rootCause []invariant.InvariantRootCauseResult) ([]*timeoutRootCauseResult, error) {
	result := make([]*timeoutRootCauseResult, 0)
	for _, rc := range rootCause {
		if rootCausePollersRelated(rc.RootCause) {
			var metadata timeout.PollersMetadata
			err := json.Unmarshal(rc.Metadata, &metadata)
			if err != nil {
				return nil, err
			}
			result = append(result, &timeoutRootCauseResult{
				RootCauseType:   rc.RootCause.String(),
				PollersMetadata: &metadata,
			})
		} else if rootCauseHeartBeatRelated(rc.RootCause) {
			var metadata timeout.HeartbeatingMetadata
			err := json.Unmarshal(rc.Metadata, &metadata)
			if err != nil {
				return nil, err
			}
			result = append(result, &timeoutRootCauseResult{
				RootCauseType:        rc.RootCause.String(),
				HeartBeatingMetadata: &metadata,
			})
		}
	}

	return result, nil
}

func retrieveFailureIssues(issues []invariant.InvariantCheckResult) ([]*failureIssuesResult, error) {
	result := make([]*failureIssuesResult, 0)
	for _, issue := range issues {
		if issue.InvariantType == failure.ActivityFailed.String() || issue.InvariantType == failure.WorkflowFailed.String() || issue.InvariantType == failure.DecisionCausedFailure.String() {
			var data failure.FailureMetadata
			err := json.Unmarshal(issue.Metadata, &data)
			if err != nil {
				return nil, err
			}
			result = append(result, &failureIssuesResult{
				InvariantType: issue.InvariantType,
				Reason:        issue.Reason,
				Metadata:      &data,
			})
		}
	}
	return result, nil
}

func retrieveFailureRootCause(rootCause []invariant.InvariantRootCauseResult) ([]*failureRootCauseResult, error) {
	result := make([]*failureRootCauseResult, 0)
	for _, rc := range rootCause {
		if rc.RootCause == invariant.RootCauseTypeServiceSideIssue || rc.RootCause == invariant.RootCauseTypeServiceSidePanic || rc.RootCause == invariant.RootCauseTypeServiceSideCustomError {
			result = append(result, &failureRootCauseResult{
				RootCauseType: rc.RootCause.String(),
			})
		}
		if rc.RootCause == invariant.RootCauseTypeBlobSizeLimit {
			var metadata failure.BlobSizeMetadata
			err := json.Unmarshal(rc.Metadata, &metadata)
			if err != nil {
				return nil, err
			}
			result = append(result, &failureRootCauseResult{
				RootCauseType:    rc.RootCause.String(),
				BlobSizeMetadata: &metadata,
			})
		}
	}
	return result, nil
}

func retrieveRetryIssues(issues []invariant.InvariantCheckResult) ([]*retryIssuesResult, error) {
	result := make([]*retryIssuesResult, 0)
	for _, issue := range issues {
		if issue.InvariantType == retry.WorkflowRetryIssue.String() || issue.InvariantType == retry.WorkflowRetryInfo.String() || issue.InvariantType == retry.ActivityRetryIssue.String() {
			var data retry.RetryMetadata
			err := json.Unmarshal(issue.Metadata, &data)
			if err != nil {
				return nil, err
			}
			result = append(result, &retryIssuesResult{
				InvariantType: issue.InvariantType,
				Reason:        issue.Reason,
				Metadata:      data,
			})
		}
	}
	return result, nil
}

func rootCauseHeartBeatRelated(rootCause invariant.RootCause) bool {
	for _, rc := range []invariant.RootCause{invariant.RootCauseTypeNoHeartBeatTimeoutNoRetryPolicy,
		invariant.RootCauseTypeHeartBeatingNotEnabledWithRetryPolicy,
		invariant.RootCauseTypeHeartBeatingEnabledWithoutRetryPolicy,
		invariant.RootCauseTypeHeartBeatingEnabledMissingHeartbeat} {
		if rc == rootCause {
			return true
		}
	}
	return false
}

func rootCausePollersRelated(rootCause invariant.RootCause) bool {
	for _, rc := range []invariant.RootCause{invariant.RootCauseTypePollersStatus, invariant.RootCauseTypeMissingPollers} {
		if rc == rootCause {
			return true
		}
	}
	return false
}
