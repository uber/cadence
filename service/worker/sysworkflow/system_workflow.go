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
	"github.com/uber-go/tally"
	"github.com/uber/cadence/common/logging"
	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

// SystemWorkflow is the system workflow code
func SystemWorkflow(ctx workflow.Context) error {
	id := workflow.GetInfo(ctx).WorkflowExecution.ID
	logger := workflow.GetLogger(ctx)
	scope := workflow.GetMetricsScope(ctx).Tagged(map[string]string{SystemWorkflowIDTag: id})
	ch := workflow.GetSignalChannel(ctx, SignalName)
	signalsHandled := 0

	for signalsHandled < SignalsUntilContinueAsNew {
		var signal Signal
		if more := ch.Receive(ctx, &signal); !more {
			scope.Counter(ChannelClosedUnexpectedlyError).Inc(1)
			logger.Error("cadence channel was unexpectedly closed")
			break
		}
		selectSystemTask(scope, signal, ctx, logger)
		signalsHandled++
	}

	for {
		var signal Signal
		if ok := ch.ReceiveAsync(&signal); !ok {
			break
		}
		selectSystemTask(scope, signal, ctx, logger)
		signalsHandled++
	}

	logger.Info("completed current set of iterations, continuing as new",
		zap.Int(logging.TagIterationsUntilContinueAsNew, signalsHandled))
	return workflow.NewContinueAsNewError(ctx, "SystemWorkflow")
}

func selectSystemTask(scope tally.Scope, signal Signal, ctx workflow.Context, logger *zap.Logger) {
	scope.Counter(HandledSignalCount).Inc(1)

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 10,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:          time.Second,
			BackoffCoefficient:       2.0,
			MaximumInterval:          time.Minute,
			ExpirationInterval:       time.Hour * 24 * 30,
			MaximumAttempts:          0,
			NonRetriableErrorReasons: []string{"bad-error"},
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)
	switch signal.RequestType {
	case ArchivalRequest:
		workflow.ExecuteActivity(
			ctx,
			"ArchivalActivity",
			signal.ArchiveRequest.UserWorkflowID,
			signal.ArchiveRequest.UserRunID,
		)
	default:
		scope.Counter(UnknownSignalTypeErr).Inc(1)
		logger.Error("received unknown request type")
	}
}

// ArchivalActivity is the archival activity code
func ArchivalActivity(ctx context.Context, userWorkflowID string, userRunId string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("starting archival",
		zap.String(logging.TagUserWorkflowID, userWorkflowID),
		zap.String(logging.TagUserRunID, userRunId))

	for i := 0; i < 20; i++ {
		time.Sleep(time.Second)
		activity.RecordHeartbeat(ctx, i)

		// if activity failure occurs restart the activity from the beginning
		if 0 == rand.Intn(20) {
			logger.Info("activity failed, will retry...")
			return cadence.NewCustomError("some-retryable-error")
		}
	}

	logger.Info("finished archival",
		zap.String(logging.TagUserWorkflowID, userWorkflowID),
		zap.String(logging.TagUserRunID, userRunId))
	return nil
}
