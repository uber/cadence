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

package tasklist

import (
	"context"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/resource"
)

const (
	infiniteDuration      = 20 * 365 * 24 * time.Hour
	tlScavengerHBInterval = 10 * time.Second

	tlScannerWFID                 = "cadence-sys-tl-scanner"
	WFTypeName                    = "cadence-sys-tl-scanner-workflow"
	tlName                        = "cadence-sys-tl-scanner-tasklist-0"
	taskListScavengerActivityName = "cadence-sys-tl-scanner-scvg-activity"
)

var (
	WFStartOptions = client.StartWorkflowOptions{
		ID:                           tlScannerWFID,
		TaskList:                     tlName,
		ExecutionStartToCloseTimeout: 5 * 24 * time.Hour,
		WorkflowIDReusePolicy:        client.WorkflowIDReusePolicyAllowDuplicate,
		CronSchedule:                 "0 */12 * * *",
	}
	activityRetryPolicy = cadence.RetryPolicy{
		InitialInterval:    10 * time.Second,
		BackoffCoefficient: 1.7,
		MaximumInterval:    5 * time.Minute,
		ExpirationInterval: infiniteDuration,
	}
	activityOptions = workflow.ActivityOptions{
		ScheduleToStartTimeout: 5 * time.Minute,
		StartToCloseTimeout:    infiniteDuration,
		HeartbeatTimeout:       5 * time.Minute,
		RetryPolicy:            &activityRetryPolicy,
	}
)

// Activities is used to pass resources to activities
type Activities struct {
	resource        resource.Resource
	hbSleepInterval time.Duration
	opts            *Options
}

// RegisterWorkflow registers workflow and activity with worker
func RegisterWorkflow(w worker.Worker, resource resource.Resource, opts *Options) {
	a := &Activities{
		resource:        resource,
		hbSleepInterval: tlScavengerHBInterval,
		opts:            opts,
	}
	w.RegisterWorkflowWithOptions(Workflow, workflow.RegisterOptions{Name: WFTypeName})
	w.RegisterActivityWithOptions(a.TaskListScavengerActivity, activity.RegisterOptions{Name: taskListScavengerActivityName})
}

// Workflow is the workflow that runs the task-list scanner background daemon
func Workflow(
	ctx workflow.Context,
) error {

	future := workflow.ExecuteActivity(workflow.WithActivityOptions(ctx, activityOptions), taskListScavengerActivityName)
	return future.Get(ctx, nil)
}

// TaskListScavengerActivity is the activity that runs task list scavenger
func (a *Activities) TaskListScavengerActivity(
	activityCtx context.Context,
) error {
	res := a.resource
	scavenger := NewScavenger(activityCtx,
		res.GetTaskManager(),
		res.GetMetricsClient(),
		res.GetLogger(),
		a.opts,
	)
	scavenger.Start()
	for scavenger.Alive() {
		activity.RecordHeartbeat(activityCtx)
		if activityCtx.Err() != nil {
			res.GetLogger().Info("activity context error, stopping scavenger", tag.Error(activityCtx.Err()))
			scavenger.Stop()
			return activityCtx.Err()
		}
		time.Sleep(a.hbSleepInterval)
	}
	return nil
}
