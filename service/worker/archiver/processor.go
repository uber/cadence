package archiver

import (
	"context"
	"encoding/json"
	"github.com/uber-common/bark"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/blobstore/blob"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/logging"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/persistence"
	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"time"
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
// TODO: emit metrics here
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

// TODO: make sure context is respected everywhere to bail out of activities

// Finished will block until all work has been finished.
// Returns hashes of requests handled.
// TODO: emit metrics here
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
	logger := a.logger.WithFields(bark.Fields{
		logging.TagArchiveRequestDomainID:          request.DomainID,
		logging.TagArchiveRequestWorkflowID:        request.WorkflowID,
		logging.TagArchiveRequestRunID:             request.RunID,
		logging.TagArchiveRequestEventStoreVersion: request.EventStoreVersion,
		logging.TagArchiveRequestBranchToken:       string(request.BranchToken),
		logging.TagArchiveRequestNextEventID:       request.NextEventID,
		logging.TagArchiveRequestCloseFailoverVersion: request.CloseFailoverVersion,
	})
	uploadActOpts := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		HeartbeatTimeout:       time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    100 * time.Millisecond,
			BackoffCoefficient: 2.0,
			ExpirationInterval: 15 * time.Minute,
			// TODO: double check these are still correct
			NonRetriableErrorReasons: []string{
				errArchivalUploadActivityGetDomainStr,
				errArchivalUploadActivityNextBlobStr,
				errArchivalUploadActivityConstructKeyStr,
				errArchivalUploadActivityBlobExistsStr,
				errArchivalUploadActivityMarshalBlobStr,
				errArchivalUploadActivityConvertHeaderToTagsStr,
				errArchivalUploadActivityWrapBlobStr,
				errArchivalUploadActivityUploadBlobStr,
			},
		},
	}
	uploadActCtx := workflow.WithActivityOptions(a.ctx, uploadActOpts)
	if err := workflow.ExecuteActivity(uploadActCtx, uploadHistoryActivityFnName, request).Get(uploadActCtx, nil); err != nil {
		logger.WithField(logging.TagErr, err).Error("failed to upload history, moving on to deleting history without archiving")
		a.metricsClient.IncCounter(metrics.ArchiveSystemWorkflowScope, metrics.SysWorkerArchivalUploadActivityNonRetryableFailures)
	} else {
		a.metricsClient.IncCounter(metrics.ArchiveSystemWorkflowScope, metrics.SysWorkerArchivalUploadSuccessful)
	}

	localDeleteActOpts := workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: 1 * time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    100 * time.Millisecond,
			BackoffCoefficient: 2.0,
			ExpirationInterval: 3 * time.Minute,
			NonRetriableErrorReasons: []string{
				errDeleteHistoryActivityDeleteFromV1Str,
				errDeleteHistoryActivityDeleteFromV2Str,
			},
		},
	}
	localDeleteActCtx := workflow.WithLocalActivityOptions(a.ctx, localDeleteActOpts)
	err := workflow.ExecuteLocalActivity(localDeleteActCtx, archivalDeleteHistoryActivity, request).Get(localDeleteActCtx, nil)
	if err == nil {
		a.metricsClient.IncCounter(metrics.ArchiveSystemWorkflowScope, metrics.SysWorkerArchivalDeleteHistorySuccessful)
		return
	}
	// TODO: emit metric here as well
	a.logger.WithField(logging.TagErr, err).Warn("archivalDeleteHistoryActivity could not be completed as a local activity, attempting to run as normal activity")

	deleteActOpts := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
		HeartbeatTimeout:       time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    100 * time.Millisecond,
			BackoffCoefficient: 2.0,
			ExpirationInterval: 15 * time.Minute,
			NonRetriableErrorReasons: []string{
				errDeleteHistoryActivityDeleteFromV1Str,
				errDeleteHistoryActivityDeleteFromV2Str,
			},
		},
	}
	deleteActCtx := workflow.WithActivityOptions(a.ctx, deleteActOpts)
	if err := workflow.ExecuteActivity(deleteActCtx, archivalDeleteHistoryActivityFnName, request).Get(deleteActCtx, nil); err != nil {
		logger.WithField(logging.TagErr, err).Error("failed to delete history, this means zombie histories are left")
		// TODO: rename these metrics because they might not be non-retryable errors anymore
		a.metricsClient.IncCounter(metrics.ArchiveSystemWorkflowScope, metrics.SysWorkerArchivalDeleteHistoryActivityNonRetryableFailures)
	} else {
		a.metricsClient.IncCounter(metrics.ArchiveSystemWorkflowScope, metrics.SysWorkerArchivalDeleteHistorySuccessful)
	}
}


