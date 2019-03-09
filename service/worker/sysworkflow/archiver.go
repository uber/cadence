package sysworkflow

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
	// Archiver is used to handle archival requests
	Archiver interface {
		Start()
		Finished() []uint64
	}

	archiver struct {
		ctx workflow.Context
		logger bark.Logger
		metricsClient metrics.Client
		concurrency int
		requestCh workflow.Channel
		resultCh workflow.Channel
	}
)

func NewArchiver(
	ctx workflow.Context,
	logger bark.Logger,
	metricsClient metrics.Client,
	concurrency int,
	requestCh workflow.Channel,
) Archiver {
	return &archiver{
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
func (a *archiver) Start() {
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
func (a *archiver) Finished() []uint64 {
	var handledHashes []uint64
	for i := 0; i < a.concurrency; i++ {
		var subResult []uint64
		a.resultCh.Receive(a.ctx, &subResult)
		handledHashes = append(handledHashes, subResult...)
	}
	a.resultCh.Close()
	return handledHashes
}

func (a *archiver) handleRequest(request ArchiveRequest) {
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
	if err := workflow.ExecuteActivity(uploadActCtx, archivalUploadActivityFnName, request).Get(uploadActCtx, nil); err != nil {
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

// TODO: now down here I need to make sure I am killing these activities whenever context dies
// TODO: make sure I setup the go routine to heartbeat in order to update the context

// ArchivalUploadActivity does the following three things:
// 1. Read history from persistence
// 2. Construct blobs
// 3. Upload blobs
// It is assumed that history is immutable when this activity is running. Under this assumption this activity is idempotent.
// If an error is returned it will be of type archivalActivityNonRetryableErr. All retryable errors are retried forever.
func archivalUploadActivity(ctx context.Context, request ArchiveRequest) error {
	go func() {
		for {
			<-time.After(heartbeatTimeout / 2)
			activity.RecordHeartbeat(ctx)
		}
	}()
	container := ctx.Value(sysWorkerContainerKey).(*SysWorkerContainer)
	logger := container.Logger.WithFields(bark.Fields{
		logging.TagArchiveRequestDomainID:          request.DomainID,
		logging.TagArchiveRequestWorkflowID:        request.WorkflowID,
		logging.TagArchiveRequestRunID:             request.RunID,
		logging.TagArchiveRequestEventStoreVersion: request.EventStoreVersion,
		logging.TagArchiveRequestNextEventID:       request.NextEventID,
	})
	metricsClient := container.MetricsClient
	domainCache := container.DomainCache
	clusterMetadata := container.ClusterMetadata
	domainCacheEntry, err := getDomainByIDRetryForever(domainCache, request.DomainID)
	if err != nil {
		logger.WithError(err).Error("failed to get domain from domain cache")
		metricsClient.IncCounter(metrics.ArchivalUploadActivityScope, metrics.SysWorkerGetDomainFailures)
		return errArchivalUploadActivityGetDomain
	}

	// if signal is being handled when cluster is paused or enabled then keep processing signal
	// only if cluster sets archival to disabled should the activity be aborted
	if !clusterMetadata.ArchivalConfig().ConfiguredForArchival() {
		logger.Warn("archival is not enabled for cluster, skipping archival upload")
		metricsClient.IncCounter(metrics.ArchivalUploadActivityScope, metrics.SysWorkerArchivalNotEnabledForCluster)
		return nil
	}
	if domainCacheEntry.GetConfig().ArchivalStatus != shared.ArchivalStatusEnabled {
		// for now if archival is disabled simply abort the activity
		// a more in depth design meeting is needed to decide the correct way to handle backfilling/pausing archival
		logger.Warn("archival is not enabled for domain, skipping archival upload")
		metricsClient.IncCounter(metrics.ArchivalUploadActivityScope, metrics.SysWorkerArchivalNotEnabledForDomain)
		return nil
	}

	domainName := domainCacheEntry.GetInfo().Name
	clusterName := container.ClusterMetadata.GetCurrentClusterName()
	historyBlobItr := container.HistoryBlobIterator
	if historyBlobItr == nil {
		historyBlobItr = NewHistoryBlobIterator(logger, metricsClient, request, container, domainName, clusterName)
	}

	blobstoreClient := container.Blobstore
	bucket := domainCacheEntry.GetConfig().ArchivalBucket
	for historyBlobItr.HasNext() {
		historyBlob, err := nextBlobRetryForever(historyBlobItr)
		if err != nil {
			logger.WithError(err).Error("failed to get next blob from iterator, stopping archival upload")
			metricsClient.IncCounter(metrics.ArchivalUploadActivityScope, metrics.SysWorkerNextBlobNonRetryableFailures)
			return errArchivalUploadActivityNextBlob
		}
		key, err := NewHistoryBlobKey(request.DomainID, request.WorkflowID, request.RunID, *historyBlob.Header.CurrentPageToken)
		if err != nil {
			logger.WithError(err).Error("failed to construct blob key, stopping archival upload")
			metricsClient.IncCounter(metrics.ArchivalUploadActivityScope, metrics.SysWorkerKeyConstructionFailures)
			return errArchivalUploadActivityConstructKey
		}
		if exists, err := blobExistsRetryForever(blobstoreClient, bucket, key); err != nil {
			logger.WithError(err).Error("failed to check if blob exists already, stopping archival upload")
			metricsClient.IncCounter(metrics.ArchivalUploadActivityScope, metrics.SysWorkerBlobExistsNonRetryableFailures)
			return errArchivalUploadActivityBlobExists
		} else if exists {
			continue
		}
		body, err := json.Marshal(historyBlob)
		if err != nil {
			logger.WithError(err).Error("failed to marshal history blob. stopping archival upload")
			metricsClient.IncCounter(metrics.ArchivalUploadActivityScope, metrics.SysWorkerMarshalBlobFailures)
			return errArchivalUploadActivityMarshalBlob
		}
		tags, err := ConvertHeaderToTags(historyBlob.Header)
		if err != nil {
			logger.WithError(err).Error("failed to convert header to tags, stopping archival upload")
			metricsClient.IncCounter(metrics.ArchivalUploadActivityScope, metrics.SysWorkerConvertHeaderToTagsFailures)
			return errArchivalUploadActivityConvertHeaderToTags
		}
		wrapFunctions := []blob.WrapFn{blob.JSONEncoded()}
		if container.Config.EnableArchivalCompression(domainName) {
			wrapFunctions = append(wrapFunctions, blob.GzipCompressed())
		}
		currBlob, err := blob.Wrap(blob.NewBlob(body, tags), wrapFunctions...)
		if err != nil {
			logger.WithError(err).Error("failed to wrap blob, stopping archival upload")
			metricsClient.IncCounter(metrics.ArchivalUploadActivityScope, metrics.SysWorkerWrapBlobFailures)
			return errArchivalUploadActivityWrapBlob
		}
		if err := blobUploadRetryForever(blobstoreClient, bucket, key, currBlob); err != nil {
			logger.WithError(err).Error("failed to upload blob, stopping archival upload")
			metricsClient.IncCounter(metrics.ArchivalUploadActivityScope, metrics.SysWorkerBlobUploadNonRetryableFailures)
			return errArchivalUploadActivityUploadBlob
		}
	}
	return nil
}

// ArchivalDeleteHistoryActivity deletes the workflow execution history from persistence.
// All retryable errors are retried forever. If an error is returned it is of type archivalActivityNonRetryableErr.
func archivalDeleteHistoryActivity(ctx context.Context, request ArchiveRequest) error {
	container := ctx.Value(sysWorkerContainerKey).(*SysWorkerContainer)
	logger := container.Logger.WithFields(bark.Fields{
		logging.TagArchiveRequestDomainID:          request.DomainID,
		logging.TagArchiveRequestWorkflowID:        request.WorkflowID,
		logging.TagArchiveRequestRunID:             request.RunID,
		logging.TagArchiveRequestEventStoreVersion: request.EventStoreVersion,
		logging.TagArchiveRequestNextEventID:       request.NextEventID,
	})
	metricsClient := container.MetricsClient
	if request.EventStoreVersion == persistence.EventStoreVersionV2 {
		err := persistence.DeleteWorkflowExecutionHistoryV2(container.HistoryV2Manager, request.BranchToken, container.Logger)
		if err == nil {
			return nil
		}
		op := func() error {
			return persistence.DeleteWorkflowExecutionHistoryV2(container.HistoryV2Manager, request.BranchToken, container.Logger)
		}
		for err != nil && common.IsPersistenceTransientError(err) {
			err = backoff.Retry(op, common.CreatePersistanceRetryPolicy(), common.IsPersistenceTransientError)
		}
		logger.WithError(err).Error("failed to delete history from events v2")
		metricsClient.IncCounter(metrics.ArchivalDeleteHistoryActivityScope, metrics.SysWorkerDeleteHistoryV2NonRetryableFailures)
		return errDeleteHistoryActivityDeleteFromV2
	}
	deleteHistoryReq := &persistence.DeleteWorkflowExecutionHistoryRequest{
		DomainID: request.DomainID,
		Execution: shared.WorkflowExecution{
			WorkflowId: common.StringPtr(request.WorkflowID),
			RunId:      common.StringPtr(request.RunID),
		},
	}
	err := container.HistoryManager.DeleteWorkflowExecutionHistory(deleteHistoryReq)
	if err == nil {
		return nil
	}
	op := func() error {
		return container.HistoryManager.DeleteWorkflowExecutionHistory(deleteHistoryReq)
	}
	for err != nil && common.IsPersistenceTransientError(err) {
		err = backoff.Retry(op, common.CreatePersistanceRetryPolicy(), common.IsPersistenceTransientError)
	}
	logger.WithError(err).Error("failed to delete history from events v1")
	metricsClient.IncCounter(metrics.ArchivalDeleteHistoryActivityScope, metrics.SysWorkerDeleteHistoryV1NonRetryableFailures)
	return errDeleteHistoryActivityDeleteFromV1
}

func nextBlobRetryForever(historyBlobItr HistoryBlobIterator) (*HistoryBlob, error) {
	result, err := historyBlobItr.Next()
	if err == nil {
		return result, nil
	}

	op := func() error {
		result, err = historyBlobItr.Next()
		return err
	}
	for err != nil && common.IsPersistenceTransientError(err) {
		err = backoff.Retry(op, common.CreatePersistanceRetryPolicy(), common.IsPersistenceTransientError)
	}
	return result, err
}

func blobExistsRetryForever(blobstoreClient blobstore.Client, bucket string, key blob.Key) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), blobstoreOperationsDefaultTimeout)
	exists, err := blobstoreClient.Exists(ctx, bucket, key)
	cancel()
	for err != nil && common.IsBlobstoreTransientError(err) {
		// blobstoreClient is already retryable so no extra retry/backoff logic is needed here
		ctx, cancel = context.WithTimeout(context.Background(), blobstoreOperationsDefaultTimeout)
		exists, err = blobstoreClient.Exists(ctx, bucket, key)
		cancel()
	}
	return exists, err
}

func blobUploadRetryForever(blobstoreClient blobstore.Client, bucket string, key blob.Key, blob *blob.Blob) error {
	ctx, cancel := context.WithTimeout(context.Background(), blobstoreOperationsDefaultTimeout)
	err := blobstoreClient.Upload(ctx, bucket, key, blob)
	cancel()
	for err != nil && common.IsBlobstoreTransientError(err) {
		// blobstoreClient is already retryable so no extra retry/backoff logic is needed here
		ctx, cancel = context.WithTimeout(context.Background(), blobstoreOperationsDefaultTimeout)
		err = blobstoreClient.Upload(ctx, bucket, key, blob)
		cancel()
	}
	return err
}

func getDomainByIDRetryForever(domainCache cache.DomainCache, id string) (*cache.DomainCacheEntry, error) {
	entry, err := domainCache.GetDomainByID(id)
	if err == nil {
		return entry, nil
	}
	op := func() error {
		entry, err = domainCache.GetDomainByID(id)
		return err
	}
	for err != nil && common.IsPersistenceTransientError(err) {
		backoff.Retry(op, common.CreatePersistanceRetryPolicy(), common.IsPersistenceTransientError)
	}
	return entry, err
}
