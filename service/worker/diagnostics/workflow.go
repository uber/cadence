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
	"github.com/uber/cadence/service/worker/diagnostics/invariants"
)

const (
	diagnosticsWorkflow    = "diagnostics-workflow"
	tasklist               = "diagnostics-wf-tasklist"
	queryDiagnosticsReport = "query-diagnostics-report"

	retrieveWfExecutionHistoryActivity = "retrieveWfExecutionHistory"
	identifyTimeoutsActivity           = "identifyTimeouts"
	rootCauseTimeoutsActivity          = "rootCauseTimeouts"
)

type DiagnosticsWorkflowInput struct {
	Domain     string
	WorkflowID string
	RunID      string
}

type DiagnosticsWorkflowResult struct {
	Timeouts *timeoutDiagnostics
}

type timeoutDiagnostics struct {
	Issues    []*timeoutIssuesResult
	RootCause []*timeoutRootCauseResult
	Runbooks  []string
}

type timeoutIssuesResult struct {
	InvariantType    string
	Reason           string
	ExecutionTimeout *invariants.ExecutionTimeoutMetadata
	ActivityTimeout  *invariants.ActivityTimeoutMetadata
	ChildWfTimeout   *invariants.ChildWfTimeoutMetadata
	DecisionTimeout  *invariants.DecisionTimeoutMetadata
}

type timeoutRootCauseResult struct {
	RootCauseType        string
	PollersMetadata      *invariants.PollersMetadata
	HeartBeatingMetadata *invariants.HeartbeatingMetadata
}

func (w *dw) DiagnosticsWorkflow(ctx workflow.Context, params DiagnosticsWorkflowInput) (*DiagnosticsWorkflowResult, error) {
	scope := w.metricsClient.Scope(metrics.DiagnosticsWorkflowScope, metrics.DomainTag(params.Domain))
	scope.IncCounter(metrics.DiagnosticsWorkflowCount)
	sw := scope.StartTimer(metrics.DiagnosticsWorkflowExecutionLatency)
	defer sw.Stop()

	var timeoutsResult timeoutDiagnostics
	err := workflow.SetQueryHandler(ctx, queryDiagnosticsReport, func() (DiagnosticsWorkflowResult, error) {
		return DiagnosticsWorkflowResult{Timeouts: &timeoutsResult}, nil
	})
	if err != nil {
		return nil, err
	}

	activityOptions := workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Second * 10,
		ScheduleToStartTimeout: time.Second * 5,
		StartToCloseTimeout:    time.Second * 5,
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOptions)

	var wfExecutionHistory *types.GetWorkflowExecutionHistoryResponse
	err = workflow.ExecuteActivity(activityCtx, w.retrieveExecutionHistory, retrieveExecutionHistoryInputParams{
		Domain: params.Domain,
		Execution: &types.WorkflowExecution{
			WorkflowID: params.WorkflowID,
			RunID:      params.RunID,
		}}).Get(ctx, &wfExecutionHistory)
	if err != nil {
		return nil, fmt.Errorf("RetrieveExecutionHistory: %w", err)
	}

	timeoutsResult.Runbooks = []string{linkToTimeoutsRunbook}

	var checkResult []invariants.InvariantCheckResult
	err = workflow.ExecuteActivity(activityCtx, w.identifyTimeouts, identifyTimeoutsInputParams{
		History: wfExecutionHistory,
		Domain:  params.Domain,
	}).Get(ctx, &checkResult)
	if err != nil {
		return nil, fmt.Errorf("IdentifyTimeouts: %w", err)
	}

	timeoutIssues, err := retrieveTimeoutIssues(checkResult)
	if err != nil {
		return nil, fmt.Errorf("RetrieveTimeoutIssues: %w", err)
	}
	timeoutsResult.Issues = timeoutIssues

	var rootCauseResult []invariants.InvariantRootCauseResult
	err = workflow.ExecuteActivity(activityCtx, w.rootCauseTimeouts, rootCauseTimeoutsParams{
		History: wfExecutionHistory,
		Domain:  params.Domain,
		Issues:  checkResult,
	}).Get(ctx, &rootCauseResult)
	if err != nil {
		return nil, fmt.Errorf("RootCauseTimeouts: %w", err)
	}

	timeoutRootCause, err := retrieveTimeoutRootCause(rootCauseResult)
	if err != nil {
		return nil, fmt.Errorf("RetrieveTimeoutRootCause: %w", err)
	}
	timeoutsResult.RootCause = timeoutRootCause

	scope.IncCounter(metrics.DiagnosticsWorkflowSuccess)
	return &DiagnosticsWorkflowResult{Timeouts: &timeoutsResult}, nil
}

func retrieveTimeoutIssues(issues []invariants.InvariantCheckResult) ([]*timeoutIssuesResult, error) {
	result := make([]*timeoutIssuesResult, 0)
	for _, issue := range issues {
		switch issue.InvariantType {
		case invariants.TimeoutTypeExecution.String():
			var metadata invariants.ExecutionTimeoutMetadata
			err := json.Unmarshal(issue.Metadata, &metadata)
			if err != nil {
				return nil, err
			}
			result = append(result, &timeoutIssuesResult{
				InvariantType:    issue.InvariantType,
				Reason:           issue.Reason,
				ExecutionTimeout: &metadata,
			})
		case invariants.TimeoutTypeActivity.String():
			var metadata invariants.ActivityTimeoutMetadata
			err := json.Unmarshal(issue.Metadata, &metadata)
			if err != nil {
				return nil, err
			}
			result = append(result, &timeoutIssuesResult{
				InvariantType:   issue.InvariantType,
				Reason:          issue.Reason,
				ActivityTimeout: &metadata,
			})
		case invariants.TimeoutTypeChildWorkflow.String():
			var metadata invariants.ChildWfTimeoutMetadata
			err := json.Unmarshal(issue.Metadata, &metadata)
			if err != nil {
				return nil, err
			}
			result = append(result, &timeoutIssuesResult{
				InvariantType:  issue.InvariantType,
				Reason:         issue.Reason,
				ChildWfTimeout: &metadata,
			})
		case invariants.TimeoutTypeDecision.String():
			var metadata invariants.DecisionTimeoutMetadata
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

func retrieveTimeoutRootCause(rootCause []invariants.InvariantRootCauseResult) ([]*timeoutRootCauseResult, error) {
	result := make([]*timeoutRootCauseResult, 0)
	for _, rc := range rootCause {
		switch rc.RootCause {
		case invariants.RootCauseTypePollersStatus, invariants.RootCauseTypeMissingPollers:
			var metadata invariants.PollersMetadata
			err := json.Unmarshal(rc.Metadata, &metadata)
			if err != nil {
				return nil, err
			}
			result = append(result, &timeoutRootCauseResult{
				RootCauseType:   rc.RootCause.String(),
				PollersMetadata: &metadata,
			})
		case invariants.RootCauseTypeHeartBeatingNotEnabled, invariants.RootCauseTypeHeartBeatingEnabledMissingHeartbeat:
			var metadata invariants.HeartbeatingMetadata
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
