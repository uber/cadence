// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
//
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

package shardscannertest

import (
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/service/worker/scanner/shardscanner"
)

// NewTestWorkflow is a helper, no-op workflow used for testing purposes.
func NewTestWorkflow(ctx workflow.Context, name string, params shardscanner.ScannerWorkflowParams) error {
	wf, err := shardscanner.NewScannerWorkflow(ctx, name, params)
	if err != nil {
		return err
	}

	return wf.Start(ctx)
}

// NewTestFixerWorkflow is a helper, no-op workflow used for testing purposes.
func NewTestFixerWorkflow(ctx workflow.Context, params shardscanner.FixerWorkflowParams) error {
	wf, err := shardscanner.NewFixerWorkflow(ctx, "test-fixer", params)
	if err != nil {
		return err
	}

	return wf.Start(ctx)

}
