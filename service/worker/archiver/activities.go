package archiver

import (
	"context"
	"errors"
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
	"github.com/uber/cadence/common/persistence"
	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"

)

const (
	uploadHistoryActivityFnName   = "uploadHistoryActivity"
	deleteHistoryActivityFnName = "deleteHistoryActivity"
	heartbeatTimeout = time.Minute
	blobstoreTimeout = 30 * time.Second

	errGetDomainStr           = "failed to get domain from domain cache"
	errEmptyBucketForEnabledDomain = "got empty bucket for domain which has archival enabled, this means database is in corrupted state"
	errGetNextBlob = "failed to get next history blob from iterator"
	errConstructBlobKey = "failed to construct history blob key"


	errDeleteHistoryActivityDeleteFromV2Str         = "failed to delete history from events v2"
	errDeleteHistoryActivityDeleteFromV1Str         = "failed to delete history from events v1"
)

var (
	errContextExpired = errors.New("activity context expired")
)

func uploadHistoryActivity(ctx context.Context, request ArchiveRequest) error {
	go activityHeartbeat(ctx)
	container := ctx.Value(bootstrapContainerKey).(*BootstrapContainer)
	logger := taggedLogger(container.Logger, request)
	metricsClient := container.MetricsClient
	domainCache := container.DomainCache
	clusterMetadata := container.ClusterMetadata

	// TODO: lets make all these methods smarter: they should log, emit metrics and return the correct error type
	// that way all you need to do here is call method and check if err is not nil, it is not nil just always return the error that method returned
	domainCacheEntry, err := getDomainByID(ctx, domainCache, request.DomainID)
	if err != nil {
		logger.WithError(err).Error("failed to get domain from domain cache")
		if err == errContextExpired {
			return nil
		}
		return cadence.NewCustomError(errGetDomainStr)
	}

	if !clusterMetadata.ArchivalConfig().ConfiguredForArchival() {
		logger.Warn("archival is not enabled for cluster, skipping archival upload")
		return nil
	}
	if domainCacheEntry.GetConfig().ArchivalStatus != shared.ArchivalStatusEnabled {
		logger.Warn("archival is not enabled for domain, skipping archival upload")
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
	if len(bucket) == 0 {
		// log and emit metric in this case
		return cadence.NewCustomError(errEmptyBucketForEnabledDomain)

	}
	for historyBlobItr.HasNext() {
		historyBlob, err := nextBlob(ctx, historyBlobItr)
		if err != nil {
			logger.WithError(err).Error("failed to get next blob from iterator, stopping archival upload")
			return cadence.NewCustomError(errGetNextBlob)
		}
		key, err := NewHistoryBlobKey(request.DomainID, request.WorkflowID, request.RunID, *historyBlob.Header.CurrentPageToken)
		if err != nil {
			logger.WithError(err).Error("failed to construct blob key, stopping archival upload")
			return cadence.NewCustomError(errConstructBlobKey)
		}
		if exists, err := blobExists(ctx, blobstoreClient, bucket, key); err != nil {
			logger.WithError(err).Error("failed to check if blob exists already, stopping archival upload")
			return errArchivalUploadActivityBlobExists
		} else if exists {
			continue
		}
		body, err := json.Marshal(historyBlob)
		if err != nil {
			logger.WithError(err).Error("failed to marshal history blob. stopping archival upload")
			return errArchivalUploadActivityMarshalBlob
		}
		tags, err := ConvertHeaderToTags(historyBlob.Header)
		if err != nil {
			logger.WithError(err).Error("failed to convert header to tags, stopping archival upload")
			return errArchivalUploadActivityConvertHeaderToTags
		}
		wrapFunctions := []blob.WrapFn{blob.JSONEncoded()}
		if container.Config.EnableArchivalCompression(domainName) {
			wrapFunctions = append(wrapFunctions, blob.GzipCompressed())
		}
		currBlob, err := blob.Wrap(blob.NewBlob(body, tags), wrapFunctions...)
		if err != nil {
			logger.WithError(err).Error("failed to wrap blob, stopping archival upload")
			return errArchivalUploadActivityWrapBlob
		}
		if err := uploadBlob(ctx, blobstoreClient, bucket, key, currBlob); err != nil {
			logger.WithError(err).Error("failed to upload blob, stopping archival upload")
			return errArchivalUploadActivityUploadBlob
		}
	}
	return nil
}

func deleteHistoryActivity(ctx context.Context, request ArchiveRequest) error {
	go activityHeartbeat(ctx)
	container := ctx.Value(bootstrapContainerKey).(*BootstrapContainer)
	logger := taggedLogger(container.Logger, request)
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
	return errDeleteHistoryActivityDeleteFromV1
}

func taggedLogger(logger bark.Logger, request ArchiveRequest) bark.Logger {
	return logger.WithFields(bark.Fields{
		logging.TagArchiveRequestDomainID:             request.DomainID,
		logging.TagArchiveRequestWorkflowID:           request.WorkflowID,
		logging.TagArchiveRequestRunID:                request.RunID,
		logging.TagArchiveRequestEventStoreVersion:    request.EventStoreVersion,
		logging.TagArchiveRequestBranchToken:          string(request.BranchToken), // TODO: verify that this logs correctly and lets us delete history
		logging.TagArchiveRequestNextEventID:          request.NextEventID,
		logging.TagArchiveRequestCloseFailoverVersion: request.CloseFailoverVersion,
	})
}

func nextBlob(ctx context.Context, historyBlobItr HistoryBlobIterator) (*HistoryBlob, error) {
	blob, err := historyBlobItr.Next()
	if err == nil {
		return blob, nil
	}
	op := func() error {
		blob, err = historyBlobItr.Next()
		return err
	}
	for err != nil {
		if !common.IsPersistenceTransientError(err) {
			return nil, err
		}
		if err := contextExpired(ctx); err != nil {
			return nil, err
		}
		backoff.Retry(op, common.CreatePersistanceRetryPolicy(), common.IsPersistenceTransientError)
	}
	return blob, nil
}

func blobExists(ctx context.Context, blobstoreClient blobstore.Client, bucket string, key blob.Key) (bool, error) {
	bCtx, cancel := context.WithTimeout(ctx, blobstoreTimeout)
	exists, err := blobstoreClient.Exists(bCtx, bucket, key)
	cancel()
	for err != nil {
		if !common.IsBlobstoreTransientError(err) {
			return false, err
		}
		if err := contextExpired(ctx); err != nil {
			return false, err
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
		if err := contextExpired(ctx); err != nil {
			return err
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
		if err := contextExpired(ctx); err != nil {
			return nil, err
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

func contextExpired(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return errContextExpired
	default:
		return nil
	}
}