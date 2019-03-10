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
		ctx workflow.Context
		logger bark.Logger
		metricsClient metrics.Client
		concurrency int
		requestCh workflow.Channel
		resultCh workflow.Channel
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
		ctx: ctx,
		logger: logger,
		metricsClient: metricsClient,
		concurrency: concurrency,
		requestCh: requestCh,
		resultCh: workflow.NewBufferedChannel(ctx, concurrency),
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
				a.handleRequest(request)
				handledHashes = append(handledHashes, HashArchiveRequest(request))
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

func (a *processor) handleRequest(request ArchiveRequest) {
	logger := TagLoggerWithRequest(a.logger, request)
	uploadActOpts := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		HeartbeatTimeout:       time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    100 * time.Millisecond,
			BackoffCoefficient: 2.0,
			ExpirationInterval: 10 * time.Minute,
			NonRetriableErrorReasons: uploadHistoryActivityNonRetryableErrors,
		},
	}
	uploadActCtx := workflow.WithActivityOptions(a.ctx, uploadActOpts)
	if err := workflow.ExecuteActivity(uploadActCtx, uploadHistoryActivityFnName, request).Get(uploadActCtx, nil); err != nil {
		logger.WithField(logging.TagErr, err).Error("failed to upload history, moving on to deleting history without archiving")
	} else {
		// emit success metric
	}

	localDeleteActOpts := workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: 1 * time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    100 * time.Millisecond,
			BackoffCoefficient: 2.0,
			ExpirationInterval: 3 * time.Minute,
			NonRetriableErrorReasons: deleteHistoryActivityNonRetryableErrors,
		},
	}
	localDeleteActCtx := workflow.WithLocalActivityOptions(a.ctx, localDeleteActOpts)
	err := workflow.ExecuteLocalActivity(localDeleteActCtx, deleteHistoryActivity, request).Get(localDeleteActCtx, nil)
	if err == nil {
		// emit success metric
		return
	}
	// emit metric here indicating that local activity for delete failed
	a.logger.WithField(logging.TagErr, err).Warn("archivalDeleteHistoryActivity could not be completed as a local activity, attempting to run as normal activity")
	deleteActOpts := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		HeartbeatTimeout:       time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    100 * time.Millisecond,
			BackoffCoefficient: 2.0,
			ExpirationInterval: 10 * time.Minute,
			NonRetriableErrorReasons: deleteHistoryActivityNonRetryableErrors,
		},
	}
	deleteActCtx := workflow.WithActivityOptions(a.ctx, deleteActOpts)
	if err := workflow.ExecuteActivity(deleteActCtx, deleteHistoryActivityFnName, request).Get(deleteActCtx, nil); err != nil {
		logger.WithField(logging.TagErr, err).Error("failed to delete history, this means zombie histories are left")
		// emit some failure metric
	} else {
		// emit some success metric
	}
}


