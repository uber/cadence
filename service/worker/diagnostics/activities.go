package diagnostics

import (
	"context"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/worker/diagnostics/invariants"
)

type retrieveExecutionHistoryInputParams struct {
	domain    string
	execution *types.WorkflowExecution
}

func (w *dw) retrieveExecutionHistory(ctx context.Context, info retrieveExecutionHistoryInputParams) (*types.GetWorkflowExecutionHistoryResponse, error) {
	frontendClient := w.clientBean.GetFrontendClient()
	return frontendClient.GetWorkflowExecutionHistory(ctx, &types.GetWorkflowExecutionHistoryRequest{
		Domain:    info.domain,
		Execution: info.execution,
	})
}

type identifyTimeoutsInputParams struct {
	history *types.GetWorkflowExecutionHistoryResponse
}

func (w *dw) identifyTimeouts(ctx context.Context, info identifyTimeoutsInputParams) ([]invariants.InvariantCheckResult, error) {
	timeoutInvariant := invariants.NewTimeout(info.history)
	return timeoutInvariant.Check(ctx)
}
