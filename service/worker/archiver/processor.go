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

package archiver

import (
	"time"

	"github.com/uber-common/bark"
	"github.com/uber/cadence/common/logging"
	"github.com/uber/cadence/common/metrics"
	"go.uber.org/cadence"
	"go.uber.org/cadence/workflow"
)

type (
	// Processor is used to process archival requests
	Processor interface {
		Start()
		Finished() []uint64
	}

	processor struct {
		ctx           workflow.Context
		logger        bark.Logger
		metricsClient metrics.Client
		concurrency   int
		requestCh     workflow.Channel
		resultCh      workflow.Channel
	}
)

// NewProcessor returns a new Processor
func NewProcessor(
	ctx workflow.Context,
	logger bark.Logger,
	metricsClient metrics.Client,
	concurrency int,
	requestCh workflow.Channel,
) Processor {
	return &processor{
		ctx:           ctx,
		logger:        logger,
		metricsClient: metricsClient,
		concurrency:   concurrency,
		requestCh:     requestCh,
		resultCh:      workflow.NewBufferedChannel(ctx, concurrency),
	}
}

// Start spawns concurrency count of go routines to handle archivals (does not block).
func (a *processor) Start() {
	for i := 0; i < a.concurrency; i++ {
		workflow.Go(a.ctx, func(ctx workflow.Context) {
			var handledHashes []uint64
			for {
				var request ArchiveRequest
				if more := a.requestCh.Receive(ctx, &request); !more {
					break
				}
				handleRequest(ctx, a.logger, request)
				handledHashes = append(handledHashes, hashArchiveRequest(request))
			}
			a.resultCh.Send(a.ctx, handledHashes)
		})
	}
}

// Finished will block until all work has been finished.
// Returns hashes of requests handled.
func (a *processor) Finished() []uint64 {
	var handledHashes []uint64
	for i := 0; i < a.concurrency; i++ {
		var subResult []uint64
		a.resultCh.Receive(a.ctx, &subResult)
		handledHashes = append(handledHashes, subResult...)
	}
	a.resultCh.Close()
	return handledHashes
}

func handleRequest(ctx workflow.Context, logger bark.Logger, request ArchiveRequest) {
	logger = tagLoggerWithRequest(logger, request)
	uploadActOpts := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		HeartbeatTimeout:       time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:          100 * time.Millisecond,
			BackoffCoefficient:       2.0,
			ExpirationInterval:       10 * time.Minute,
			NonRetriableErrorReasons: uploadHistoryActivityNonRetryableErrors,
		},
	}
	uploadActCtx := workflow.WithActivityOptions(ctx, uploadActOpts)
	if err := workflow.ExecuteActivity(uploadActCtx, uploadHistoryActivityFnName, request).Get(uploadActCtx, nil); err != nil {
		logger.WithField(logging.TagErr, err).Error("failed to upload history, moving on to deleting history without archiving")
	} else {
		// emit success metric
	}

	localDeleteActOpts := workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: 1 * time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:          100 * time.Millisecond,
			BackoffCoefficient:       2.0,
			ExpirationInterval:       3 * time.Minute,
			NonRetriableErrorReasons: deleteHistoryActivityNonRetryableErrors,
		},
	}
	localDeleteActCtx := workflow.WithLocalActivityOptions(ctx, localDeleteActOpts)
	err := workflow.ExecuteLocalActivity(localDeleteActCtx, deleteHistoryActivity, request).Get(localDeleteActCtx, nil)
	if err == nil {
		// emit success metric
		return
	}
	// emit metric here indicating that local activity for delete failed
	logger.WithField(logging.TagErr, err).Warn("archivalDeleteHistoryActivity could not be completed as a local activity, attempting to run as normal activity")
	deleteActOpts := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		HeartbeatTimeout:       time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:          100 * time.Millisecond,
			BackoffCoefficient:       2.0,
			ExpirationInterval:       10 * time.Minute,
			NonRetriableErrorReasons: deleteHistoryActivityNonRetryableErrors,
		},
	}
	deleteActCtx := workflow.WithActivityOptions(ctx, deleteActOpts)
	if err := workflow.ExecuteActivity(deleteActCtx, deleteHistoryActivityFnName, request).Get(deleteActCtx, nil); err != nil {
		logger.WithField(logging.TagErr, err).Error("failed to delete history, this means zombie histories are left")
		// emit some failure metric
	} else {
		// emit some success metric
	}
}
