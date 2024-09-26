package diagnostics

import (
	"fmt"
	"go.uber.org/cadence/workflow"
)

const (
	diagnosticsStarterWorkflow = "diagnostics-starter-workflow"
)

type DiagnosticsStarterWorkflowInput struct {
	Domain     string
	WorkflowID string
	RunID      string
}

type DiagnosticsStarterWorkflowResult struct {
	DiagnosticsResult *DiagnosticsWorkflowResult
}

func (w *dw) DiagnosticsStarterWorkflow(ctx workflow.Context, params DiagnosticsWorkflowInput) (*DiagnosticsStarterWorkflowResult, error) {
	var result DiagnosticsWorkflowResult

	err := workflow.ExecuteChildWorkflow(ctx, w.DiagnosticsWorkflow, DiagnosticsWorkflowInput{
		Domain:     params.Domain,
		WorkflowID: params.WorkflowID,
		RunID:      params.RunID,
	}).Get(ctx, &result)
	if err != nil {
		return nil, fmt.Errorf("Workflow Diagnostics: %w", err)
	}

	return &DiagnosticsStarterWorkflowResult{DiagnosticsResult: &result}, nil
}
