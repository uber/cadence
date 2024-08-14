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
