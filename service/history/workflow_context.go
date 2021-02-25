// Copyright (c) 2020 Uber Technologies, Inc.
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

package history

import (
	"context"

	"github.com/uber/cadence/service/history/execution"
)

type (
	// WorkflowContext is an helper interface on top of execution.Context
	WorkflowContext interface {
		GetContext() execution.Context
		GetMutableState() execution.MutableState
		ReloadMutableState(ctx context.Context) (execution.MutableState, error)
		GetReleaseFn() execution.ReleaseFunc
		GetWorkflowID() string
		GetRunID() string
	}

	workflowContextImpl struct {
		context      execution.Context
		mutableState execution.MutableState
		releaseFn    execution.ReleaseFunc
	}
)

// NewWorkflowContext creates a new helper instance on top of execution.Context
func NewWorkflowContext(
	context execution.Context,
	releaseFn execution.ReleaseFunc,
	mutableState execution.MutableState,
) WorkflowContext {

	return &workflowContextImpl{
		context:      context,
		releaseFn:    releaseFn,
		mutableState: mutableState,
	}
}

func (w *workflowContextImpl) GetContext() execution.Context {
	return w.context
}

func (w *workflowContextImpl) GetMutableState() execution.MutableState {
	return w.mutableState
}

func (w *workflowContextImpl) ReloadMutableState(ctx context.Context) (execution.MutableState, error) {
	mutableState, err := w.GetContext().LoadWorkflowExecution(ctx)
	if err != nil {
		return nil, err
	}
	w.mutableState = mutableState
	return mutableState, nil
}

func (w *workflowContextImpl) GetReleaseFn() execution.ReleaseFunc {
	return w.releaseFn
}

func (w *workflowContextImpl) GetWorkflowID() string {
	return w.GetContext().GetExecution().GetWorkflowID()
}

func (w *workflowContextImpl) GetRunID() string {
	return w.GetContext().GetExecution().GetRunID()
}
