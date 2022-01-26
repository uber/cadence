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
	"fmt"
	"strings"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	cclient "go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/types"
)

const (
	// workflow constants
	watchdogWFID       = "cadence-sys-watchdog"
	taskListName       = "cadence-sys-tl-watchdog"
	watchdogWFTypeName = "cadence-sys-watchdog-workflow"

	// activities
	handleCorrputedWorkflowActivity = "cadence-sys-watchdog-handle-corrupted-workflow"

	// signals
	corruptWorkflowWatchdogChannelName = "CorruptWorkflowWatchdogChannelName"
)

type (
	Workflow struct {
		watchdog  *WatchDog
		processed cache.Cache
	}

	// Request defines the request for corruptWorkflow maintenance
	CorruptWFRequest struct {
		workflow   types.WorkflowExecution
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

	handleCorruptedWorkflowOptions = workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		RetryPolicy:            &retryPolicy,
	}

	wfOptions = cclient.StartWorkflowOptions{
		ID:                           watchdogWFID,
		TaskList:                     taskListName,
		ExecutionStartToCloseTimeout: 24 * 365 * time.Hour, // 1 year
	}

	corruptWorkflowErrorList = [3]string{
		"unable to get workflow start event",                        // desc + history
		"unable to get activity scheduled event",                    // desc + history
		"corrupted history event batch, eventID is not continouous", // history only
	}
)

func initWorkflow(wd *WatchDog) {
	cacheOpts := &cache.Options{
		InitialCapacity: 100,
		MaxCount:        1000,
		TTL:             24 * time.Hour,
		Pin:             false,
	}

	w := Workflow{
		watchdog:  wd,
		processed: cache.New(cacheOpts),
	}

	workflow.RegisterWithOptions(w.workflowFunc, workflow.RegisterOptions{Name: watchdogWFTypeName})
	activity.RegisterWithOptions(w.handleCorrputedWorkflow, activity.RegisterOptions{Name: handleCorrputedWorkflowActivity})
}

// workflowFunc is the workflow that performs actions for WatchDog
func (w *Workflow) workflowFunc(ctx workflow.Context) error {
	requestCh := workflow.GetSignalChannel(ctx, corruptWorkflowWatchdogChannelName)
	logger := w.watchdog.logger

	for {
		var request CorruptWFRequest
		if more := requestCh.Receive(ctx, &request); !more {
			logger.Info("Corrupt workflow channel closed")
			return cadence.NewCustomError("signal_channel_closed")
		}

		if w.watchdog.config.CorruptWorkflowPause() {
			logger.Warn("Corrupt workflow execution is paused. Enable to continue processing")
			continue
		}
		opt := workflow.WithActivityOptions(ctx, handleCorruptedWorkflowOptions)
		_ = workflow.ExecuteActivity(opt, handleCorrputedWorkflowActivity, request).Get(ctx, nil)
	}
}

func (w *Workflow) getQueryTemplates(ctx context.Context, request *CorruptWFRequest) []func(request *CorruptWFRequest) error {
	// describe workflow
	client := w.watchdog.frontendClient
	return []func(request *CorruptWFRequest) error{
		func(request *CorruptWFRequest) error {
			_, err := client.DescribeWorkflowExecution(ctx, &types.DescribeWorkflowExecutionRequest{
				Domain: request.DomainName,
				Execution: &types.WorkflowExecution{
					WorkflowID: request.workflow.GetWorkflowID(),
					RunID:      request.workflow.GetRunID(),
				},
			})
			return err
		},
		func(request *CorruptWFRequest) error {
			_, err := client.GetWorkflowExecutionHistory(ctx, &types.GetWorkflowExecutionHistoryRequest{
				Domain: request.DomainName,
				Execution: &types.WorkflowExecution{
					WorkflowID: request.workflow.GetWorkflowID(),
					RunID:      request.workflow.GetRunID(),
				},
			})
			return err
		},
	}
}

func (w *Workflow) deleteWorkflow(ctx context.Context, request *CorruptWFRequest, logger *zap.Logger) error {
	domainEntry, err := w.watchdog.domainCache.GetDomain(request.DomainName)
	if err != nil {
		logger.Error("Failed to get domain entry", zap.Error(err), zap.String("DomainName", request.DomainName))
		return err
	}
	clusterName := domainEntry.GetReplicationConfig().ActiveClusterName
	adminClient := w.watchdog.clientBean.GetRemoteAdminClient(clusterName)
	wfDeleteRequest := &types.AdminDeleteWorkflowRequest{
		Domain:    request.DomainName,
		Execution: &request.workflow,
	}
	adminClient.DeleteWorkflow(ctx, wfDeleteRequest)
	return nil
}

func (w *Workflow) getProcessedID(request *CorruptWFRequest) string {
	return fmt.Sprintf("%s-%s", request.workflow.WorkflowID, request.workflow.RunID)
}

// handleCorrputedWorkflowActivity is activity to handle corrupted workflows in DB
func (w *Workflow) handleCorrputedWorkflow(ctx context.Context, request *CorruptWFRequest) error {
	if w.processed.Get(w.getProcessedID(request)) != nil {
		// We already processed this workflow before and couldn't decide if we should delete
		return nil
	}

	logger := activity.GetLogger(ctx)
	logger.Error("Processing",
		zap.String("domainName", request.DomainName),
		zap.String("workflowID", request.workflow.WorkflowID),
		zap.String("runID", request.workflow.RunID),
	)

	tagged := w.watchdog.scopedMetricClient.Tagged(
		metrics.DomainTag(request.DomainName),
	)

	queryTemplates := w.getQueryTemplates(ctx, request)
	for _, queryFunc := range queryTemplates {
		err := queryFunc(request)
		if err == nil {
			continue
		}
		errorMessage := err.Error()
		for _, corruptMessage := range corruptWorkflowErrorList {
			if strings.Contains(errorMessage, corruptMessage) {
				err = w.deleteWorkflow(ctx, request, logger)
				if err == nil {
					tagged.AddCounter(metrics.WatchDogNumDeletedCorruptWorkflows, 1)
				} else {
					tagged.AddCounter(metrics.WatchDogNumFailedToDeleteCorruptWorkflows, 1)
				}
			}
		}
	}

	// We couldn't decide if we should delete this workflow. That means the same logs will continue
	// and the workflow will keep getting signals. So adding workflow to a cache to skip processing
	w.processed.Put(w.getProcessedID(request), struct{}{})

	return nil
}
