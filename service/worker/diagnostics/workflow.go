package diagnostics

import (
	"fmt"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/worker/diagnostics/invariants"
	"time"

	"go.uber.org/cadence/workflow"
)

const (
	diagnosticsWorkflow = "diagnostics-workflow"
	tasklist            = "wf-diagnostics"

	retrieveWfExecutionHistoryActivity = "retrieveWfExecutionHistory"
	identifyTimeoutsActivity           = "identifyTimeouts"
)

type DiagnosticsWorkflowInput struct {
	Domain     string
	WorkflowID string
	RunID      string
}

func (w *dw) DiagnosticsWorkflow(ctx workflow.Context, params DiagnosticsWorkflowInput) error {
	activityOptions := workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Second * 10,
		ScheduleToStartTimeout: time.Second * 5,
		StartToCloseTimeout:    time.Second * 5,
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOptions)

	var wfExecutionHistory *types.GetWorkflowExecutionHistoryResponse
	err := workflow.ExecuteActivity(activityCtx, w.retrieveExecutionHistory, retrieveExecutionHistoryInputParams{
		domain: params.Domain,
		execution: &types.WorkflowExecution{
			WorkflowID: params.WorkflowID,
			RunID:      params.RunID,
		}}).Get(ctx, &wfExecutionHistory)
	if err != nil {
		return fmt.Errorf("RetrieveExecutionHistory: %w", err)
	}

	var checkResult []invariants.InvariantCheckResult
	err = workflow.ExecuteActivity(activityCtx, w.identifyTimeouts, identifyTimeoutsInputParams{
		history: wfExecutionHistory}).Get(ctx, &checkResult)
	if err != nil {
		return fmt.Errorf("IdentifyTimeouts: %w", err)
	}

	return nil
}
