// Copyright (c) 2022 Uber Technologies, Inc.
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

package watchdog

import (
	"context"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	cclient "go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
)

const (
	// workflow constants
	WatchdogWFID       = "cadence-sys-watchdog"
	taskListName       = "cadence-sys-tl-watchdog"
	watchdogWFTypeName = "cadence-sys-watchdog-workflow"

	// activities
	handleCorruptedWorkflowActivity = "cadence-sys-watchdog-handle-corrupted-workflow"

	// signals
	CorruptWorkflowWatchdogChannelName = "CorruptWorkflowWatchdogChannelName"
)

type (
	Workflow struct {
		watchdog *WatchDog
	}

	CorruptWFRequest struct {
		Workflow   types.WorkflowExecution
		DomainName string
	}
)

var (
	retryPolicy = cadence.RetryPolicy{
		InitialInterval:    10 * time.Second,
		BackoffCoefficient: 1.7,
		MaximumInterval:    5 * time.Minute,
		ExpirationInterval: time.Hour,
	}

	handleCorruptWorkflowOptions = workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		RetryPolicy:            &retryPolicy,
	}

	wfOptions = cclient.StartWorkflowOptions{
		ID:                           WatchdogWFID,
		TaskList:                     taskListName,
		WorkflowIDReusePolicy:        cclient.WorkflowIDReusePolicyTerminateIfRunning,
		ExecutionStartToCloseTimeout: 24 * 365 * time.Hour, // 1 year
	}
)

func initWorkflow(wd *WatchDog) {
	w := Workflow{
		watchdog: wd,
	}

	workflow.RegisterWithOptions(w.workflowFunc, workflow.RegisterOptions{Name: watchdogWFTypeName})
	activity.RegisterWithOptions(w.handleCorruptedWorkflow, activity.RegisterOptions{Name: handleCorruptedWorkflowActivity})
}

// workflowFunc is the workflow that performs actions for WatchDog
func (w *Workflow) workflowFunc(ctx workflow.Context) error {
	requestCh := workflow.GetSignalChannel(ctx, CorruptWorkflowWatchdogChannelName)
	logger := w.watchdog.logger

	for {
		var request CorruptWFRequest
		if more := requestCh.Receive(ctx, &request); !more {
			logger.Info("Corrupt workflow channel closed")
			return cadence.NewCustomError("signal_channel_closed")
		}

		if w.watchdog.config.CorruptWorkflowWatchdogPause() {
			logger.Warn("Corrupt workflow execution is paused. Enable to continue processing")
			continue
		}
		opt := workflow.WithActivityOptions(ctx, handleCorruptWorkflowOptions)
		_ = workflow.ExecuteActivity(opt, handleCorruptedWorkflowActivity, request).Get(ctx, nil)
	}
}

// handleCorruptedWorkflowActivity is activity to handle corrupted workflows in DB
func (w *Workflow) handleCorruptedWorkflow(ctx context.Context, request *CorruptWFRequest) error {
	logger := activity.GetLogger(ctx).With(
		zap.String("DomainName", request.DomainName),
		zap.String("WorkflowID", request.Workflow.GetWorkflowID()),
		zap.String("RunID", request.Workflow.GetRunID()))
	logger.Info("Watchdog processing possible corrupt workflow")
	domainEntry, err := w.watchdog.domainCache.GetDomain(request.DomainName)
	if err != nil {
		logger.Error("Failed to get domain entry", zap.Error(err))
		return err
	}

	tagged := w.watchdog.scopedMetricClient.Tagged(
		metrics.DomainTag(domainEntry.GetInfo().Name),
	)
	clusterName := domainEntry.GetReplicationConfig().ActiveClusterName
	adminClient := w.watchdog.clientBean.GetRemoteAdminClient(clusterName)
	maintainWFRequest := &types.AdminMaintainWorkflowRequest{
		Domain:     request.DomainName,
		Execution:  &request.Workflow,
		SkipErrors: true,
	}
	_, err = adminClient.MaintainCorruptWorkflow(ctx, maintainWFRequest)
	if err != nil {
		return err
	}

	tagged.AddCounter(metrics.WatchDogNumCorruptWorkflowProcessed, 1)

	return nil
}
