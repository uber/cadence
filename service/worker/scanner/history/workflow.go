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

package history

import (
	"context"
	"time"

	"github.com/uber/cadence/common/domain"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/resource"
	"github.com/uber/cadence/common/service/dynamicconfig"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
)

const (
	historyScannerWFID           = "cadence-sys-history-scanner"
	WFTypeName                   = "cadence-sys-history-scanner-workflow"
	TaskListName                 = "cadence-sys-history-scanner-tasklist-0"
	historyScavengerActivityName = "cadence-sys-history-scanner-scvg-activity"
	infiniteDuration             = 20 * 365 * 24 * time.Hour
)

var (
	WFStartOptions = client.StartWorkflowOptions{
		ID:                           historyScannerWFID,
		TaskList:                     TaskListName,
		ExecutionStartToCloseTimeout: infiniteDuration,
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

type Activities struct {
	resource resource.Resource
	// scannerPersistenceMaxQPS the max rate of calls to persistence
	scannerPersistenceMaxQPS   dynamicconfig.IntPropertyFn
	maxWorkflowRetentionInDays dynamicconfig.IntPropertyFn
}

func NewHistoryScanner(w worker.Worker, resource resource.Resource, dc *dynamicconfig.Collection) {
	a := &Activities{
		resource:                   resource,
		scannerPersistenceMaxQPS:   dc.GetIntProperty(dynamicconfig.ScannerPersistenceMaxQPS, 5),
		maxWorkflowRetentionInDays: dc.GetIntProperty(dynamicconfig.MaxRetentionDays, domain.DefaultMaxWorkflowRetentionInDays),
	}

	w.RegisterWorkflowWithOptions(HistoryWorkflow, workflow.RegisterOptions{Name: WFTypeName})
	w.RegisterActivityWithOptions(a.HistoryScavenger, activity.RegisterOptions{Name: historyScavengerActivityName})
}

// HistoryWorkflow is the workflow that runs the history scanner background daemon
func HistoryWorkflow(
	ctx workflow.Context,
) error {
	a := &Activities{}

	future := workflow.ExecuteActivity(
		workflow.WithActivityOptions(ctx, activityOptions),
		a.HistoryScavenger,
	)
	return future.Get(ctx, nil)
}

// HistoryScavenger is the activity that runs history scavenger
func (a *Activities) HistoryScavenger(
	activityCtx context.Context,
) (ScavengerHeartbeatDetails, error) {

	res := a.resource

	hbd := ScavengerHeartbeatDetails{}
	if activity.HasHeartbeatDetails(activityCtx) {
		if err := activity.GetHeartbeatDetails(activityCtx, &hbd); err != nil {
			res.GetLogger().Error("Failed to recover from last heartbeat, start over from beginning", tag.Error(err))
		}
	}

	scavenger := NewScavenger(
		res.GetHistoryManager(),
		a.scannerPersistenceMaxQPS(),
		res.GetHistoryClient(),
		hbd,
		res.GetMetricsClient(),
		res.GetLogger(),
		a.maxWorkflowRetentionInDays,
	)

	return scavenger.Run(activityCtx)
}
