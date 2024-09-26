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
	"go.uber.org/cadence/workflow"
)

const (
	diagnosticsStarterWorkflow = "diagnostics-starter-workflow"
	queryDiagnosticsReport     = "query-diagnostics-report"
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
	err := workflow.SetQueryHandler(ctx, queryDiagnosticsReport, func() (DiagnosticsStarterWorkflowResult, error) {
		return DiagnosticsStarterWorkflowResult{DiagnosticsResult: &result}, nil
	})
	if err != nil {
		return nil, err
	}

	err = workflow.ExecuteChildWorkflow(ctx, w.DiagnosticsWorkflow, DiagnosticsWorkflowInput{
		Domain:     params.Domain,
		WorkflowID: params.WorkflowID,
		RunID:      params.RunID,
	}).Get(ctx, &result)
	if err != nil {
		return nil, fmt.Errorf("Workflow Diagnostics: %w", err)
	}

	return &DiagnosticsStarterWorkflowResult{DiagnosticsResult: &result}, nil
}
