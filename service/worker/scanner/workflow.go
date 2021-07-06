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

package scanner

import (
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/service/worker/scanner/executions"
	"github.com/uber/cadence/service/worker/scanner/timers"
)

// @TODO (mantas) replace this with registration through worker
func init() {
	workflow.RegisterWithOptions(executions.ConcreteScannerWorkflow, workflow.RegisterOptions{Name: executions.ConcreteExecutionsScannerWFTypeName})
	workflow.RegisterWithOptions(executions.CurrentScannerWorkflow, workflow.RegisterOptions{Name: executions.CurrentExecutionsScannerWFTypeName})
	workflow.RegisterWithOptions(executions.ConcreteFixerWorkflow, workflow.RegisterOptions{Name: executions.ConcreteExecutionsFixerWFTypeName})
	workflow.RegisterWithOptions(executions.CurrentFixerWorkflow, workflow.RegisterOptions{Name: executions.CurrentExecutionsFixerWFTypeName})
	workflow.RegisterWithOptions(timers.ScannerWorkflow, workflow.RegisterOptions{Name: timers.ScannerWFTypeName})
	workflow.RegisterWithOptions(timers.FixerWorkflow, workflow.RegisterOptions{Name: timers.FixerWFTypeName})
}
