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

package workflow

import (
	"context"

	"github.com/uber/cadence/service/history/execution"
)

type (
	// Context is an helper interface on top of execution.Context
	Context interface {
		GetContext() execution.Context
		GetMutableState() execution.MutableState
		ReloadMutableState(ctx context.Context) (execution.MutableState, error)
		GetReleaseFn() execution.ReleaseFunc
		GetWorkflowID() string
		GetRunID() string
	}

	contextImpl struct {
		context      execution.Context
		mutableState execution.MutableState
		releaseFn    execution.ReleaseFunc
	}
)

// NewContext creates a new helper instance on top of execution.Context
func NewContext(
	context execution.Context,
	releaseFn execution.ReleaseFunc,
	mutableState execution.MutableState,
) Context {

	return &contextImpl{
		context:      context,
		releaseFn:    releaseFn,
		mutableState: mutableState,
	}
}

func (w *contextImpl) GetContext() execution.Context {
	return w.context
}

func (w *contextImpl) GetMutableState() execution.MutableState {
	return w.mutableState
}

func (w *contextImpl) ReloadMutableState(ctx context.Context) (execution.MutableState, error) {
	mutableState, err := w.GetContext().LoadWorkflowExecution(ctx)
	if err != nil {
		return nil, err
	}
	w.mutableState = mutableState
	return mutableState, nil
}

func (w *contextImpl) GetReleaseFn() execution.ReleaseFunc {
	return w.releaseFn
}

func (w *contextImpl) GetWorkflowID() string {
	return w.GetContext().GetExecution().GetWorkflowID()
}

func (w *contextImpl) GetRunID() string {
	return w.GetContext().GetExecution().GetRunID()
}
