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
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/blobstore/blob"
	"github.com/uber/cadence/common/persistence"
	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"time"
)

var (
	errArchivalActivityNonRetryable = cadence.NewCustomError(archivalActivityNonRetryableErrStr)
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
			NonRetriableErrorReasons: []string{archivalActivityNonRetryableErrStr},
		},
	}

	actCtx := workflow.WithActivityOptions(ctx, ao)
	switch signal.RequestType {
	case archivalRequest:
		if err := workflow.ExecuteActivity(
			actCtx,
			archivalUploadActivityFnName,
			*signal.ArchiveRequest,
		).Get(ctx, nil); err != nil {
			// log and emit metrics
		}
		if err := workflow.ExecuteActivity(
			actCtx,
			archivalDeleteHistoryActivityFnName,
			*signal.ArchiveRequest,
		).Get(ctx, nil); err != nil {
			// log and emit metrics
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

// ArchivalUploadActivity does the following three things:
// 1. Read history from persistence
// 2. Construct blobs
// 3. Upload blobs
// It is assumed that history is immutable when this activity is running. Under this assumption this activity is idempotent.
// If an error is returned it will be of type archivalActivityNonRetryableErr. All retryable errors are retried forever.
func ArchivalUploadActivity(ctx context.Context, request ArchiveRequest) error {
	container, ok := ctx.Value(sysWorkerContainerKey).(*SysWorkerContainer)
	if !ok {
		return errArchivalActivityNonRetryable
	}
	domainCache := container.DomainCache
	clusterMetadata := container.ClusterMetadata
	domainCacheEntry, err := domainCache.GetDomainByID(request.DomainID)
	if err != nil {
		return errArchivalActivityNonRetryable
	}
	if !clusterMetadata.IsArchivalEnabled() {
		// for now if archival is disabled simply abort the activity
		// a more in depth design meeting is needed to decide the correct way to handle backfilling/pausing archival
		return errArchivalActivityNonRetryable
	}
	if domainCacheEntry.GetConfig().ArchivalStatus != shared.ArchivalStatusEnabled {
		// for now if archival is disabled simply abort the activity
		// a more in depth design meeting is needed to decide the correct way to handle backfilling/pausing archival
		return errArchivalActivityNonRetryable
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
	bucket := domainCacheEntry.GetConfig().ArchivalBucket
	for historyBlobItr.HasNext() {
		historyBlob, err := nextBlobRetryForever(historyBlobItr)
		if err != nil {
			return errArchivalActivityNonRetryable
		}
		key, err := NewHistoryBlobKey(request.DomainID, request.WorkflowID, request.RunID, *historyBlob.Header.CurrentPageToken)
		if err != nil {
			return errArchivalActivityNonRetryable
		}
		if exists, err := blobExistsRetryForever(blobstoreClient, bucket, key); err != nil {
			return errArchivalActivityNonRetryable
		} else if exists {
			continue
		}
		body, err := json.Marshal(historyBlob)
		if err != nil {
			return errArchivalActivityNonRetryable
		}
		tags, err := ConvertHeaderToTags(historyBlob.Header)
		if err != nil {
			return errArchivalActivityNonRetryable
		}
		wrapFunctions := []blob.WrapFn{blob.JSONEncoded()}
		if container.Config.EnableArchivalCompression(domainCacheEntry.GetInfo().Name) {
			wrapFunctions = append(wrapFunctions, blob.GzipCompressed())
		}
		currBlob, err := blob.Wrap(blob.NewBlob(body, tags), wrapFunctions...)
		if err != nil {
			return errArchivalActivityNonRetryable
		}
		if err := blobUploadRetryForever(blobstoreClient, bucket, key, currBlob); err != nil {
			return errArchivalActivityNonRetryable
		}
		activity.RecordHeartbeat(ctx, historyBlob.Header.CurrentPageToken)
	}
	return nil
}

// ArchivalDeleteHistoryActivity deletes the workflow execution history from persistence.
// All retryable errors are retried forever. If an error is returned it is of type archivalActivityNonRetryableErr.
func ArchivalDeleteHistoryActivity(ctx context.Context, request ArchiveRequest) error {
	container, ok := ctx.Value(sysWorkerContainerKey).(*SysWorkerContainer)
	if !ok {
		return errArchivalActivityNonRetryable
	}
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
		return errArchivalActivityNonRetryable
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
	return errArchivalActivityNonRetryable
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
	exists, err := blobstoreClient.Exists(context.Background(), bucket, key)
	for err != nil && common.IsBlobstoreTransientError(err) {
		// blobstoreClient is already retryable so no extra retry/backoff logic is needed here
		exists, err = blobstoreClient.Exists(context.Background(), bucket, key)
	}
	return exists, err
}

func blobUploadRetryForever(blobstoreClient blobstore.Client, bucket string, key blob.Key, blob *blob.Blob) error {
	err := blobstoreClient.Upload(context.Background(), bucket, key, blob)
	for err != nil && common.IsBlobstoreTransientError(err) {
		// blobstoreClient is already retryable so no extra retry/backoff logic is needed here
		err = blobstoreClient.Upload(context.Background(), bucket, key, blob)
	}
	return err
}

// TODO: consider if this retryable activity needs to heartbeat?

// BackfillActivity is the backfill activity code
func BackfillActivity(_ context.Context, _ BackfillRequest) error {
	// TODO: write this activity
	return nil
}
