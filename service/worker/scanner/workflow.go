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
	"context"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/service/worker/scanner/tasklist"
)

type contextKey int

const (
	scannerContextKey = contextKey(0)

	maxConcurrentActivityExecutionSize     = 10
	maxConcurrentDecisionTaskExecutionSize = 10

	scannerWFCronSchedule                 = "0 7 * * *"
	scannerWFExecutionStartToCloseTimeout = 23 * time.Hour

	scannerWFTaskListName         = "cadence-sys-scanner-tasklist-0"
	scannerWFName                 = "cadence-sys-scanner-workflow"
	scannerWFID                   = "cadence-sys-scanner"
	taskListScavengerActivityName = "cadence-sys-tasklist-scanner-activity"
)

var (
	scavengerHBInterval = 10 * time.Second
)

var (
	scannerWFRetryPolicy = cadence.RetryPolicy{
		InitialInterval:    10 * time.Second,
		BackoffCoefficient: 1.7,
		MaximumInterval:    5 * time.Minute,
		ExpirationInterval: 20 * time.Hour, // after 20 hours, give up and wait for next run of cron
	}
	scannerActivityRetryPolicy = scannerWFRetryPolicy
)

func init() {
	workflow.RegisterWithOptions(Workflow, workflow.RegisterOptions{Name: scannerWFName})
	activity.RegisterWithOptions(TaskListScavengerActivity, activity.RegisterOptions{Name: taskListScavengerActivityName})
}

// Workflow is the workflow that runs the scanner background sub-system
func Workflow(ctx workflow.Context) error {
	opts := workflow.ActivityOptions{
		ScheduleToStartTimeout: 5 * time.Minute,
		StartToCloseTimeout:    23 * time.Hour,
		HeartbeatTimeout:       5 * time.Minute,
		RetryPolicy:            &scannerActivityRetryPolicy,
	}
	future := workflow.ExecuteActivity(workflow.WithActivityOptions(ctx, opts), taskListScavengerActivityName)
	future.Get(ctx, nil)
	return nil
}

// TaskListScavengerActivity is the activity that runs task list scavenger
func TaskListScavengerActivity(aCtx context.Context) error {
	ctx := aCtx.Value(scannerContextKey).(scannerContext)
	scavenger := tasklist.NewScavenger(ctx.taskDB, ctx.metricsClient, ctx.logger)
	ctx.logger.Info("Starting task list scavenger")
	scavenger.Start()
	for scavenger.Alive() {
		activity.RecordHeartbeat(aCtx)
		if aCtx.Err() != nil {
			ctx.logger.Infof("activity context error, stopping scavenger: %v", aCtx.Err())
			scavenger.Stop()
			return aCtx.Err()
		}
		time.Sleep(scavengerHBInterval)
	}
	return nil
}
