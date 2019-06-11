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

package batcher

import (
	"context"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
)

const (
	batcherContextKey      = "batcherContext"
	batcherTaskListName    = "cadence-sys-batcher-tasklist"
	batchResetWFTypeName   = "cadence-sys-batch-reset-workflow"
	batchResetActivityName = "cadence-sys-batch-reset-activity"

	infiniteDuration        = 20 * 365 * 24 * time.Hour
	batchActivityHBInterval = 10 * time.Second
)

var (
	batchActivityRetryPolicy = cadence.RetryPolicy{
		InitialInterval:    10 * time.Second,
		BackoffCoefficient: 1.7,
		MaximumInterval:    5 * time.Minute,
		ExpirationInterval: infiniteDuration,
	}

	batchActivityOptions = workflow.ActivityOptions{
		ScheduleToStartTimeout: 5 * time.Minute,
		StartToCloseTimeout:    infiniteDuration,
		HeartbeatTimeout:       5 * time.Minute,
		RetryPolicy:            &batchActivityRetryPolicy,
	}
)

func init() {
	workflow.RegisterWithOptions(BatchResetWorkflow, workflow.RegisterOptions{Name: batchResetWFTypeName})
	activity.RegisterWithOptions(BatchResetActivity, activity.RegisterOptions{Name: batchResetActivityName})
}

// BatchResetWorkflow is the workflow that runs a batch job of resetting workflows
func BatchResetWorkflow(ctx workflow.Context) error {
	future := workflow.ExecuteActivity(workflow.WithActivityOptions(ctx, batchActivityOptions), batchResetActivityName)
	return future.Get(ctx, nil)
}

func BatchResetActivity(aCtx context.Context) error {
	return nil
}
