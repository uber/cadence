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
	"code.uber.internal/devexp/cadence-server/.tmp/.go/goroot/src/encoding/json"
	"context"
	"github.com/uber-common/bark"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/blobstore"
	"github.com/uber/cadence/common/blobstore/blob"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/logging"
	"github.com/uber/cadence/common/persistence"
	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"time"
)

// SystemWorkflow is the system workflow code
// TODO: will have to put logger and metrics in global to get them out here
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
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			// TODO: I need to think through what this should really be
			ExpirationInterval: time.Hour * 24 * 30,
			MaximumAttempts:    0,
			// TODO: some types of non-retryable errors will need to be added here - maybe I map them to hardcoded string
			NonRetriableErrorReasons: []string{},
		},
	}

	actCtx := workflow.WithActivityOptions(ctx, ao)
	switch signal.RequestType {
	case archivalRequest:
		if err := workflow.ExecuteActivity(
			actCtx,
			archivalActivityFnName,
			*signal.ArchiveRequest,
		).Get(ctx, nil); err != nil {
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

// ArchivalActivity is the archival activity code
func ArchivalActivity(ctx context.Context, request ArchiveRequest) error {

	// TODO: check if archival is enabled for domain using domain cache
	logger := ctx.Value(loggerKey).(bark.Logger)
	workflowInfo := activity.GetInfo(ctx)
	logger = logger.WithFields(bark.Fields{
		// TODO: check if I need to add more stuff here...
		logging.TagWorkflowComponent:    logging.TagValueArchivalSystemWorkflowComponent,
		logging.TagArchiveIsEventsV2:    request.EventStoreVersion == persistence.EventStoreVersionV2,
		logging.TagArchiveDomainID:      request.DomainID,
		logging.TagArchiveWorkflowID:    request.WorkflowID,
		logging.TagArchiveRunID:         request.RunID,
		logging.TagWorkflowExecutionID:  workflowInfo.WorkflowExecution.ID,
		logging.TagWorkflowRunID:        workflowInfo.WorkflowExecution.RunID,
		logging.TagArchivalRetryAttempt: workflowInfo.Attempt,
	})
	domainCache := ctx.Value(domainCacheKey).(cache.DomainCache)
	entry, err := domainCache.GetDomainByID(request.DomainID)
	if err != nil {
		// TODO: it is cases like this that I need to think through retrayble vs non-retryable errors
		return err
	}

	// TODO: pull out everything I need in a reasonable order
	// TODO: consider creating a struct that contains everything needed to read history and passing it to iterators
	// config := ctx.Value(configKey).(*Config)
	itr := NewHistoryBlobIterator(
		ctx.Value(historyManagerKey).(persistence.HistoryManager),
		ctx.Value(historyV2ManagerKey).(persistence.HistoryV2Manager),
		request.DomainID,
		request.WorkflowID,
		request.RunID,
		request.EventStoreVersion,
		request.BranchToken,
		request.LastFirstEventID,
		ctx.Value(configKey).(*Config),
		entry.GetInfo().Name,
	)

	bStore := ctx.Value(blobstoreKey).(blobstore.Client)
	for itr.HasNext() {
		historyBlob, err := itr.Next()
		if err != nil {
			// TODO: also need to think through retries here...
			return err
		}
		key, err := NewHistoryBlobKey(
			request.DomainID,
			request.WorkflowID,
			request.RunID,
			*historyBlob.Header.CurrentPageToken,
			*historyBlob.Header.LastFailoverVersion,
		)
		if err != nil {
			// TODO: think through retryable error here
			return err
		}

		body, err := json.Marshal(historyBlob)
		if err != nil {
			// TODO: think through retryable errors here
			return err
		}
		tags, err := ConvertHeaderToTags(historyBlob.Header)
		if err != nil {
			return err
		}
		unwrappedBlob := blob.NewBlob(body, tags)
		blob, err := blob.Wrap(unwrappedBlob, blob.JSONEncoded(), blob.GzipCompressed())
		if err != nil {
			return err
		}

		// TODO: make sure I am plumbing through the correct blobstore so that it has retries and timeouts out of the box
		if err := bStore.Upload(ctx, entry.GetConfig().ArchivalBucket, key, blob); err != nil {
			return err
		}

		// TODO: if at anypoint we encounter a non-retryable error we should clean up everything that we have uploaded to blobstore already, so workflow needs to keep some additional state
	}

	// once once we get here can we delete the history
	if request.EventStoreVersion == persistence.EventStoreVersionV2 {
		// TODO: consider if I want to do type cast assertions or actually check their success - I don't think I need to assert because a workflow panic is fine in this case
		historyV2Manager := ctx.Value(historyV2ManagerKey).(persistence.HistoryV2Manager)

		// TODO: figuring out when this can be retried is tricky
		return persistence.DeleteWorkflowExecutionHistoryV2(historyV2Manager, request.BranchToken, logger)
	}
	historyManager := ctx.Value(historyManagerKey).(persistence.HistoryManager)
	deleteHistoryReq := &persistence.DeleteWorkflowExecutionHistoryRequest{
		DomainID: request.DomainID,
		Execution: shared.WorkflowExecution{
			WorkflowId: common.StringPtr(request.WorkflowID),
			RunId:      common.StringPtr(request.RunID),
		},
	}
	return historyManager.DeleteWorkflowExecutionHistory(deleteHistoryReq)

	// TODO: make sure I am logging and emitting metrics everywhere correctly in activity and in workflow
	logger.Info("andrew test log: archival activity called")
	return nil
}

// BackfillActivity is the backfill activity code
func BackfillActivity(_ context.Context, _ BackfillRequest) error {
	// TODO: write this activity
	return nil
}
