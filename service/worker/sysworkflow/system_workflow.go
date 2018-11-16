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
	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"time"
)

// SystemWorkflow is the system workflow code
func SystemWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	ch := workflow.GetSignalChannel(ctx, SignalName)

	for {
		var signal Signal
		if more := ch.Receive(ctx, &signal); !more {
			logger.Error("cadence channel was unexpectedly closed")
			break
		}
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
			RetryPolicy: &cadence.RetryPolicy{
				InitialInterval:    time.Second,
				BackoffCoefficient: 2.0,
				MaximumInterval:    30 * time.Minute,
				ExpirationInterval: time.Hour * 24 * 30,
				MaximumAttempts:    0,
			},
		}
		ctx = workflow.WithActivityOptions(ctx, ao)
		switch signal.RequestType {
		case ArchivalRequest:
			workflow.ExecuteActivity(ctx, ArchivalActivity, signal)
		default:
			logger.Error("received unknown request type")
		}
	}
	return nil
}

// ArchivalActivity is the archival activity code
func ArchivalActivity(ctx context.Context, signal Signal) error {
	logger := activity.GetLogger(ctx)
	logger.Info("doing archival",
		zap.String("user-workflow-id", signal.ArchiveRequest.UserWorkflowID),
		zap.String("user-run-id", signal.ArchiveRequest.UserRunID))
	return nil
}
