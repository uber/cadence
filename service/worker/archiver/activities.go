package archiver

import (
	"context"
	"time"

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

)

const (
	uploadHistoryActivityFnName   = "uploadHistoryActivity"
	deleteHistoryActivityFnName = "deleteHistoryActivity"
	heartbeatTimeout = time.Minute
	blobstoreTimeout = 30 * time.Second

	errArchivalUploadActivityGetDomainStr           = "failed to get domain from domain cache"
	errArchivalUploadActivityNextBlobStr            = "failed to get next blob from iterator"
	errArchivalUploadActivityConstructKeyStr        = "failed to construct blob key"
	errArchivalUploadActivityBlobExistsStr          = "failed to check if blob exists already"
	errArchivalUploadActivityMarshalBlobStr         = "failed to marshal history blob"
	errArchivalUploadActivityConvertHeaderToTagsStr = "failed to convert header to tags"
	errArchivalUploadActivityWrapBlobStr            = "failed to wrap blob"
	errArchivalUploadActivityUploadBlobStr          = "failed to upload blob"

	errDeleteHistoryActivityDeleteFromV2Str         = "failed to delete history from events v2"
	errDeleteHistoryActivityDeleteFromV1Str         = "failed to delete history from events v1"
)

var (
	errArchivalUploadActivityGetDomain           = cadence.NewCustomError(errArchivalUploadActivityGetDomainStr)
	errArchivalUploadActivityNextBlob            = cadence.NewCustomError(errArchivalUploadActivityNextBlobStr)
	errArchivalUploadActivityConstructKey        = cadence.NewCustomError(errArchivalUploadActivityConstructKeyStr)
	errArchivalUploadActivityBlobExists          = cadence.NewCustomError(errArchivalUploadActivityBlobExistsStr)
	errArchivalUploadActivityMarshalBlob         = cadence.NewCustomError(errArchivalUploadActivityMarshalBlobStr)
	errArchivalUploadActivityConvertHeaderToTags = cadence.NewCustomError(errArchivalUploadActivityConvertHeaderToTagsStr)
	errArchivalUploadActivityWrapBlob            = cadence.NewCustomError(errArchivalUploadActivityWrapBlobStr)
	errArchivalUploadActivityUploadBlob          = cadence.NewCustomError(errArchivalUploadActivityUploadBlobStr)
	errDeleteHistoryActivityDeleteFromV2         = cadence.NewCustomError(errDeleteHistoryActivityDeleteFromV2Str)
	errDeleteHistoryActivityDeleteFromV1         = cadence.NewCustomError(errDeleteHistoryActivityDeleteFromV1Str)
)

// TODO: now down here I need to make sure I am killing these activities whenever context dies
// TODO: make sure I setup the go routine to heartbeat in order to update the context


func uploadHistoryActivity(ctx context.Context, request ArchiveRequest) error {
	go activityHeartbeat(ctx)
	container := ctx.Value(bootstrapContainerKey).(*BootstrapContainer)
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

func deleteHistoryActivity(ctx context.Context, request ArchiveRequest) error {
	go activityHeartbeat(ctx)
	container := ctx.Value(bootstrapContainerKey).(*BootstrapContainer)
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

func nextBlob(ctx context.Context, historyBlobItr HistoryBlobIterator) (*HistoryBlob, error) {
	result, err := historyBlobItr.Next()
	if err == nil {
		return result, nil
	}

	op := func() error {
		result, err = historyBlobItr.Next()
		return err
	}
	for err != nil {
		if !common.IsPersistenceTransientError(err) {
			return nil, err
		}
		if contextExpired(ctx) {
			return nil, ctx.Err()
		}
		backoff.Retry(op, common.CreatePersistanceRetryPolicy(), common.IsPersistenceTransientError)
	}
	return result, nil
}

func blobExists(ctx context.Context, blobstoreClient blobstore.Client, bucket string, key blob.Key) (bool, error) {
	bCtx, cancel := context.WithTimeout(ctx, blobstoreTimeout)
	exists, err := blobstoreClient.Exists(bCtx, bucket, key)
	cancel()
	for err != nil {
		if !common.IsBlobstoreTransientError(err) {
			return false, err
		}
		if contextExpired(ctx) {
			return false, ctx.Err()
		}
		bCtx, cancel = context.WithTimeout(ctx, blobstoreTimeout)
		exists, err = blobstoreClient.Exists(bCtx, bucket, key)
		cancel()
	}
	return exists, nil
}

func uploadBlob(ctx context.Context, blobstoreClient blobstore.Client, bucket string, key blob.Key, blob *blob.Blob) error {
	bCtx, cancel := context.WithTimeout(ctx, blobstoreTimeout)
	err := blobstoreClient.Upload(bCtx, bucket, key, blob)
	cancel()
	for err != nil {
		if !common.IsBlobstoreTransientError(err) {
			return err
		}
		if contextExpired(ctx) {
			return ctx.Err()
		}
		bCtx, cancel = context.WithTimeout(ctx, blobstoreTimeout)
		err = blobstoreClient.Upload(bCtx, bucket, key, blob)
		cancel()
	}
	return nil
}

func getDomainByID(ctx context.Context, domainCache cache.DomainCache, id string) (*cache.DomainCacheEntry, error) {
	entry, err := domainCache.GetDomainByID(id)
	if err == nil {
		return entry, nil
	}
	op := func() error {
		entry, err = domainCache.GetDomainByID(id)
		return err
	}
	for err != nil {
		if !common.IsPersistenceTransientError(err) {
			return nil, err
		}
		if contextExpired(ctx) {
			return nil, ctx.Err()
		}
		backoff.Retry(op, common.CreatePersistanceRetryPolicy(), common.IsPersistenceTransientError)
	}
	return entry, nil
}

func activityHeartbeat(ctx context.Context) {
	for {
		select {
		case <-time.After(heartbeatTimeout / 2):
			activity.RecordHeartbeat(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func contextExpired(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}