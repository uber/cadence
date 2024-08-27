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
	"fmt"
	"time"

	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/worker/diagnostics/invariants"
)

const (
	diagnosticsWorkflow = "diagnostics-workflow"
	tasklist            = "wf-diagnostics"

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
	Issues    []invariants.InvariantCheckResult
	RootCause []invariants.InvariantRootCauseResult
	Runbooks  []string
}

func (w *dw) DiagnosticsWorkflow(ctx workflow.Context, params DiagnosticsWorkflowInput) (DiagnosticsWorkflowResult, error) {
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
		return DiagnosticsWorkflowResult{}, fmt.Errorf("RetrieveExecutionHistory: %w", err)
	}

	var checkResult []invariants.InvariantCheckResult
	err = workflow.ExecuteActivity(activityCtx, w.identifyTimeouts, identifyTimeoutsInputParams{
		History: wfExecutionHistory,
		Domain:  params.Domain,
	}).Get(ctx, &checkResult)
	if err != nil {
		return DiagnosticsWorkflowResult{}, fmt.Errorf("IdentifyTimeouts: %w", err)
	}

	var rootCauseResult []invariants.InvariantRootCauseResult
	err = workflow.ExecuteActivity(activityCtx, w.rootCauseTimeouts, rootCauseTimeoutsParams{
		History: wfExecutionHistory,
		Domain:  params.Domain,
		Issues:  checkResult,
	}).Get(ctx, &rootCauseResult)
	if err != nil {
		return DiagnosticsWorkflowResult{}, fmt.Errorf("RootCauseTimeouts: %w", err)
	}

	return DiagnosticsWorkflowResult{
		Issues:    checkResult,
		RootCause: rootCauseResult,
		Runbooks:  []string{linkToTimeoutsRunbook},
	}, nil
}
