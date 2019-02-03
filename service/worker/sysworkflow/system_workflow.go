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
	"encoding/json"
	"errors"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/blobstore/blob"
	"go.uber.org/cadence"
	"go.uber.org/cadence/workflow"
	"time"
)

var (
	archivalUploadActivityNonRetryableErr = errors.New(archivalUploadActivityNonRetryableErrStr)
	// TODO: other archival functions will follow the same pattern here...
)

// SystemWorkflow is the system workflow code
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
			InitialInterval:          time.Second,
			BackoffCoefficient:       2.0,
			MaximumInterval:          time.Minute,
			ExpirationInterval:       time.Hour * 24 * 30,
			MaximumAttempts:          0,
			NonRetriableErrorReasons: []string{},
		},
	}

	actCtx := workflow.WithActivityOptions(ctx, ao)
	switch signal.RequestType {
	case archivalRequest:
		//if err := workflow.ExecuteActivity(
		//	actCtx,
		//	archivalActivityFnName,
		//	*signal.ArchiveRequest,
		//).Get(ctx, nil); err != nil {
		//}
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

// ArchivalUploadActivity handles reading history from persistence, generating blobs and uploading blobs.
// Returned string slice represents the keys of all uploaded blobs.
// If error is returned it will either be a retryable or non-retryable.
// All non-retryable errors are mapped to error: archivalUploadActivityNonRetryableErr.
// ArchivalUploadActivity is idempotent.
// TODO: think through activity to make sure it truely is idempotent and that failures and retries are handled reasonably
func ArchivalUploadActivity(ctx context.Context, request ArchiveRequest) ([]string, error) {
	container, ok := ctx.Value(sysWorkerContainerKey).(*SysWorkerContainer)
	if !ok {
		return nil, archivalUploadActivityNonRetryableErr
	}
	domainCache := container.DomainCache
	clusterMetadata := container.ClusterMetadata
	domainCacheEntry, err := domainCache.GetDomainByID(request.DomainID)
	if err != nil {
		return nil, archivalUploadActivityNonRetryableErr
	}
	if !clusterMetadata.IsArchivalEnabled() {
		// for now if archival is disabled simply abort the activity
		// a more in depth design meeting is needed to decide the correct way to handle backfilling/pausing archival
		return nil, archivalUploadActivityNonRetryableErr
	}
	if domainCacheEntry.GetConfig().ArchivalStatus != shared.ArchivalStatusEnabled {
		// for now if archival is disabled simply abort the activity
		// a more in depth design meeting is needed to decide the correct way to handle backfilling/pausing archival
		return nil, archivalUploadActivityNonRetryableErr
	}
	historyBlobItr := NewHistoryBlobIterator(
		container.HistoryManager,
		container.HistoryV2Manager,
		request.DomainID,
		request.WorkflowID,
		request.RunID,
		request.EventStoreVersion,
		request.BranchToken,
		request.LastFirstEventID,
		container.Config,
		domainCacheEntry.GetInfo().Name,
		container.ClusterMetadata.GetCurrentClusterName())

	blobstoreClient := container.Blobstore
	var uploadedBlobKeys []string
	bucket := domainCacheEntry.GetConfig().ArchivalBucket
	for historyBlobItr.HasNext() {
		historyBlob, err := historyBlobItr.Next()
		if err != nil {
			if common.IsPersistenceTransientError(err) {
				return uploadedBlobKeys, err
			}
			return uploadedBlobKeys, archivalUploadActivityNonRetryableErr
		}
		key, err := NewHistoryBlobKey(
			request.DomainID,
			request.WorkflowID,
			request.RunID,
			*historyBlob.Header.CurrentPageToken,
			*historyBlob.Header.LastFailoverVersion,
		)
		if err != nil {
			return nil, archivalUploadActivityNonRetryableErr
		}
		exists, err := blobstoreClient.Exists(ctx, bucket, key)
		if err != nil {
			if common.IsBlobstoreTransientError(err) {
				return uploadedBlobKeys, err
			}
			return uploadedBlobKeys, archivalUploadActivityNonRetryableErr
		}
		if exists {
			continue
		}

		body, err := json.Marshal(historyBlob)
		if err != nil {
			return nil, archivalUploadActivityNonRetryableErr
		}
		tags, err := ConvertHeaderToTags(historyBlob.Header)
		if err != nil {
			return nil, archivalUploadActivityNonRetryableErr
		}
		unwrappedBlob := blob.NewBlob(body, tags)
		blob, err := blob.Wrap(unwrappedBlob, blob.JSONEncoded(), blob.GzipCompressed())
		if err != nil {
			return nil, archivalUploadActivityNonRetryableErr
		}
		if err := blobstoreClient.Upload(ctx, bucket, key, blob); err != nil {
			if common.IsBlobstoreTransientError(err) {
				return uploadedBlobKeys, err
			}
			return uploadedBlobKeys, archivalUploadActivityNonRetryableErr
		}
	}
	return uploadedBlobKeys, nil
}

func ArchivalGarbageCollectActivity(ctx context.Context, blobsToDelete []string) error {
	return nil
}

func ArchivalDeletePersistenceHistory(ctx context.Context, request ArchiveRequest) error {
/**
	// once once we get here can we delete the history
	if request.EventStoreVersion == persistence.EventStoreVersionV2 {
		if err := persistence.DeleteWorkflowExecutionHistoryV2(container.HistoryV2Manager, request.BranchToken, container.Logger); err != nil {
			return nil, err
		}
	}
	deleteHistoryReq := &persistence.DeleteWorkflowExecutionHistoryRequest{
		DomainID: request.DomainID,
		Execution: shared.WorkflowExecution{
			WorkflowId: common.StringPtr(request.WorkflowID),
			RunId:      common.StringPtr(request.RunID),
		},
	}
	if err := container.HistoryManager.DeleteWorkflowExecutionHistory(deleteHistoryReq); err != nil {
		return nil, err
	}
	return nil, nil
 */
 return nil
}

// BackfillActivity is the backfill activity code
func BackfillActivity(_ context.Context, _ BackfillRequest) error {
	// TODO: write this activity
	return nil
}
