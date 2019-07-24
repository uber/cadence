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
	"context"
	"errors"

	"github.com/uber/cadence/common"
	carchiver "github.com/uber/cadence/common/archiver"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/metrics"
	persistencehelper "github.com/uber/cadence/common/persistence-helper"
	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
)

const (
	uploadHistoryActivityFnName = "uploadHistoryActivity"
	deleteHistoryActivityFnName = "deleteHistoryActivity"

	errCleanUpWorkflow = "failed to clean up workflow from persistence"
	// errDeleteHistoryV1 = "failed to delete history from events_v1"
	// errDeleteHistoryV2 = "failed to delete history from events_v2"

	errActivityPanic       = "cadenceInternal:Panic"
	errTimeoutStartToClose = "cadenceInternal:Timeout START_TO_CLOSE"
)

var (
	uploadHistoryActivityNonRetryableErrors = []string{errActivityPanic, carchiver.ErrArchiveNonRetriable.Error(), errTimeoutStartToClose}
	deleteHistoryActivityNonRetryableErrors = []string{errCleanUpWorkflow}
	errContextTimeout                       = errors.New("activity aborted because context timed out")
)

// uploadHistoryActivity is used to archive a workflow execution history.
// method will retry all errors except timeout errors and archiver.ErrArchiveNonRetriable.
func uploadHistoryActivity(ctx context.Context, request ArchiveRequest) (err error) {
	container := ctx.Value(bootstrapContainerKey).(*BootstrapContainer)
	defer func() {
		if err != nil {
			err = cadence.NewCustomError(err.Error())
		}
	}()
	logger := tagLoggerWithRequest(tagLoggerWithActivityInfo(container.Logger, activity.GetInfo(ctx)), request)
	scheme, err := common.GetArchivalScheme(request.URI)
	if err != nil {
		logger.Error(carchiver.ArchiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason("failed to extract archival scheme"), tag.ArchivalURI(request.URI))
		return carchiver.ErrArchiveNonRetriable
	}
	historyArchiver, err := container.ArchiverProvider.GetHistoryArchiver(scheme, common.WorkerServiceName)
	if err != nil {
		logger.Error(carchiver.ArchiveNonRetriableErrorMsg, tag.ArchivalArchiveFailReason("failed to get history archiver"), tag.Error(err))
		return carchiver.ErrArchiveNonRetriable
	}
	return historyArchiver.Archive(ctx, request.URI, &carchiver.ArchiveHistoryRequest{
		ShardID:              request.ShardID,
		DomainID:             request.DomainID,
		DomainName:           request.DomainName,
		WorkflowID:           request.WorkflowID,
		RunID:                request.RunID,
		EventStoreVersion:    request.EventStoreVersion,
		BranchToken:          request.BranchToken,
		NextEventID:          request.NextEventID,
		CloseFailoverVersion: request.CloseFailoverVersion,
	}, carchiver.GetHeartbeatArchiveOption())
}

// deleteHistoryActivity deletes workflow execution history from persistence.
// method will retry all retryable operations until context expires.
// method will always return either: nil, contextTimeoutErr or an error from deleteHistoryActivityNonRetryableErrors.
func deleteHistoryActivity(ctx context.Context, request ArchiveRequest) (err error) {
	container := ctx.Value(bootstrapContainerKey).(*BootstrapContainer)
	scope := container.MetricsClient.Scope(metrics.ArchiverDeleteHistoryActivityScope, metrics.DomainTag(request.DomainName))
	sw := scope.StartTimer(metrics.CadenceLatency)
	logger := tagLoggerWithRequest(tagLoggerWithActivityInfo(container.Logger, activity.GetInfo(ctx)), request)
	defer func() {
		sw.Stop()
		if err != nil {
			logger.Error("failed to clean up workflow", tag.ArchivalDeleteHistoryFailReason(errCleanUpWorkflow), tag.Error(err))
		}
		if err != nil && !common.IsPersistenceTransientError(err) {
			scope.IncCounter(metrics.ArchiverNonRetryableErrorCount)
		}
	}()

	return container.WorkflowCleaner.CleanUp(&persistencehelper.CleanUpRequest{
		DomainID:          request.DomainID,
		WorkflowID:        request.WorkflowID,
		RunID:             request.RunID,
		ShardID:           request.ShardID,
		FailoverVersion:   request.CloseFailoverVersion,
		EventStoreVersion: request.EventStoreVersion,
		BranchToken:       request.BranchToken,
		TaskID:            request.TaskID,
	})

	// // if request.EventStoreVersion == persistence.EventStoreVersionV2 {
	// // 	if err := deleteHistoryV2(ctx, container, request); err != nil {
	// // 		logger.Error("failed to delete history from events v2", tag.ArchivalDeleteHistoryFailReason(errorDetails(err)), tag.Error(err))
	// // 		return err
	// // 	}
	// // 	return nil
	// // }
	// // if err := deleteHistoryV1(ctx, container, request); err != nil {
	// // 	logger.Error("failed to delete history from events v1", tag.ArchivalDeleteHistoryFailReason(errorDetails(err)), tag.Error(err))
	// // 	return err
	// // }
	// return cleanUpWorkflow(ctx, container, &request)
}

// func cleanUpWorkflow(ctx context.Context, container *BootstrapContainer, request *ArchiveRequest) error {
// 	cleanUpRequest := &persistencehelper.CleanUpRequest{
// 		DomainID:          request.DomainID,
// 		WorkflowID:        request.WorkflowID,
// 		RunID:             request.RunID,
// 		ShardID:           request.ShardID,
// 		FailoverVersion:   request.CloseFailoverVersion,
// 		EventStoreVersion: request.EventStoreVersion,
// 		BranchToken:       request.BranchToken,
// 		TaskID:            request.TaskID,
// 	}
// 	err := container.WorkflowCleaner.CleanUp(cleanUpRequest)
// 	for err != nil {
// 		if !common.IsPersistenceTransientError(err) {
// 			return err
// 		}
// 		if contextExpired(ctx) {
// 			return err
// 		}
// 		err = container.WorkflowCleaner.CleanUp(cleanUpRequest)
// 	}
// 	return nil
// }

// func deleteHistoryV1(ctx context.Context, container *BootstrapContainer, request ArchiveRequest) error {
// 	deleteHistoryReq := &persistence.DeleteWorkflowExecutionHistoryRequest{
// 		DomainID: request.DomainID,
// 		Execution: shared.WorkflowExecution{
// 			WorkflowId: common.StringPtr(request.WorkflowID),
// 			RunId:      common.StringPtr(request.RunID),
// 		},
// 	}
// 	err := container.HistoryManager.DeleteWorkflowExecutionHistory(deleteHistoryReq)
// 	if err == nil {
// 		return nil
// 	}
// 	op := func() error {
// 		return container.HistoryManager.DeleteWorkflowExecutionHistory(deleteHistoryReq)
// 	}
// 	for err != nil {
// 		if !common.IsPersistenceTransientError(err) {
// 			return cadence.NewCustomError(errDeleteHistoryV1, err.Error())
// 		}
// 		if contextExpired(ctx) {
// 			return errContextTimeout
// 		}
// 		err = backoff.Retry(op, common.CreatePersistanceRetryPolicy(), common.IsPersistenceTransientError)
// 	}
// 	return nil
// }

// func deleteHistoryV2(ctx context.Context, container *BootstrapContainer, request ArchiveRequest) error {
// 	err := persistence.DeleteWorkflowExecutionHistoryV2(container.HistoryV2Manager, request.BranchToken, common.IntPtr(request.ShardID), container.Logger)
// 	if err == nil {
// 		return nil
// 	}
// 	op := func() error {
// 		return persistence.DeleteWorkflowExecutionHistoryV2(container.HistoryV2Manager, request.BranchToken, common.IntPtr(request.ShardID), container.Logger)
// 	}
// 	for err != nil {
// 		if !common.IsPersistenceTransientError(err) {
// 			return cadence.NewCustomError(errDeleteHistoryV2, err.Error())
// 		}
// 		if contextExpired(ctx) {
// 			return errContextTimeout
// 		}
// 		err = backoff.Retry(op, common.CreatePersistanceRetryPolicy(), common.IsPersistenceTransientError)
// 	}
// 	return nil
// }
