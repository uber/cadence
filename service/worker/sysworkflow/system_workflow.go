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

package sysworkflow

import (
	"context"
	"github.com/uber-common/bark"
	"github.com/uber/cadence/common/logging"
	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"time"
)

// SystemWorkflow is the system workflow code
// TODO: I should be able to have bark logger and metrics client on this context...
func SystemWorkflow(ctx workflow.Context) error {
	ch := workflow.GetSignalChannel(ctx, signalName)
	signalsHandled := 0
	for ; signalsHandled < signalsUntilContinueAsNew; signalsHandled++ {
		var signal signal
		if more := ch.Receive(ctx, &signal); !more {
			break
		}
		selectSystemTask(signal, ctx)
	}

	for {
		var signal signal
		if ok := ch.ReceiveAsync(&signal); !ok {
			break
		}
		selectSystemTask(signal, ctx)
		signalsHandled++
	}
	ctx = workflow.WithExecutionStartToCloseTimeout(ctx, workflowStartToCloseTimeout)
	ctx = workflow.WithWorkflowTaskStartToCloseTimeout(ctx, decisionTaskStartToCloseTimeout)
	return workflow.NewContinueAsNewError(ctx, systemWorkflowFnName)
}

func selectSystemTask(signal signal, ctx workflow.Context) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 10,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			// TODO: I need to think through what this should really be
			ExpirationInterval: time.Hour * 24 * 30,
			MaximumAttempts:    0,
			// TODO: some types of non-retryable errors will need to be added here - maybe I map them to hardcoded string
			NonRetriableErrorReasons: []string{},
		},
	}

	actCtx := workflow.WithActivityOptions(ctx, ao)
	switch signal.RequestType {
	case archivalRequest:
		if err := workflow.ExecuteActivity(
			actCtx,
			archivalActivityFnName,
			*signal.ArchiveRequest,
		).Get(ctx, nil); err != nil {
		}
	case backfillRequest:
		if err := workflow.ExecuteActivity(
			actCtx,
			backfillActivityFnName,
			*signal.BackillRequest,
		).Get(ctx, nil); err != nil {
		}
	default:
	}
}

// ArchivalActivity is the archival activity code
func ArchivalActivity(ctx context.Context, request ArchiveRequest) error {
	logger := ctx.Value(loggerKey).(bark.Logger)
	workflowInfo := activity.GetInfo(ctx)
	logger = logger.WithFields(bark.Fields{
		logging.TagWorkflowComponent:       logging.TagValueArchivalSystemWorkflowComponent,
		logging.TagArchiveIsEventsV2: request.IsEventsV2,
		logging.TagArchiveDomainID:         request.DomainID,
		logging.TagArchiveWorkflowID:       request.WorkflowID,
		logging.TagArchiveRunID:            request.RunID,
		logging.TagArchiveLastWriteVersion: request.LastWriteVersion,
		logging.TagWorkflowExecutionID:     workflowInfo.WorkflowExecution.ID,
		logging.TagWorkflowRunID:           workflowInfo.WorkflowExecution.RunID,
		logging.TagArchivalRetryAttempt:    workflowInfo.Attempt,
	})
	logger.Info("andrew test log: archival activity called")
	return nil
}

// BackfillActivity is the backfill activity code
func BackfillActivity(_ context.Context, _ BackfillRequest) error {
	// TODO: write this activity
	return nil
}
