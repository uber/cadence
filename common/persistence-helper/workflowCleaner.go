// Copyright (c) 2019 Uber Technologies, Inc.
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

package persistencehelper

import (
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/backoff"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/tag"
	"github.com/uber/cadence/common/persistence"
)

type (
	// WorkflowCleaner deletes workflow execution, history and visibility record from persistence
	WorkflowCleaner interface {
		CleanUp(*CleanUpRequest) error
	}

	// CleanUpRequest is the request to clean up a workflow
	CleanUpRequest struct {
		DomainID          string
		WorkflowID        string
		RunID             string
		ShardID           int
		FailoverVersion   int64
		EventStoreVersion int32
		BranchToken       []byte
		TaskID            int64
	}

	workflowCleaner struct {
		executionManager  persistence.ExecutionManager
		historyManager    persistence.HistoryManager
		historyV2Manager  persistence.HistoryV2Manager
		visibilityManager persistence.VisibilityManager
		logger            log.Logger
	}
)

var (
	persistenceOperationRetryPolicy = common.CreatePersistanceRetryPolicy()
)

// NewWorkflowCleaner returns a new workflow cleaner.
func NewWorkflowCleaner(
	executionManager persistence.ExecutionManager,
	historyManager persistence.HistoryManager,
	historyV2Manager persistence.HistoryV2Manager,
	visibilityManager persistence.VisibilityManager,
	logger log.Logger,
) WorkflowCleaner {
	return &workflowCleaner{
		executionManager:  executionManager,
		historyManager:    historyManager,
		historyV2Manager:  historyV2Manager,
		visibilityManager: visibilityManager,
		logger:            logger,
	}
}

func (c *workflowCleaner) CleanUp(request *CleanUpRequest) error {
	if err := c.deleteCurrentWorkflowExecution(request); err != nil && !isNotExistError(err) {
		return err
	}
	if err := c.deleteWorkflowExecution(request); err != nil && !isNotExistError(err) {
		return err
	}
	if err := c.deleteWorkflowHistory(request); err != nil && !isNotExistError(err) {
		return err
	}
	if err := c.deleteWorkflowVisibility(request); err != nil && !isNotExistError(err) {
		return err
	}
	return nil
}

func (c *workflowCleaner) deleteCurrentWorkflowExecution(request *CleanUpRequest) error {
	op := func() error {
		return c.executionManager.DeleteCurrentWorkflowExecution(&persistence.DeleteCurrentWorkflowExecutionRequest{
			DomainID:   request.DomainID,
			WorkflowID: request.WorkflowID,
			RunID:      request.RunID,
		})
	}
	return backoff.Retry(op, persistenceOperationRetryPolicy, common.IsPersistenceTransientError)
}

func (c *workflowCleaner) deleteWorkflowExecution(request *CleanUpRequest) error {
	op := func() error {
		return c.executionManager.DeleteWorkflowExecution(&persistence.DeleteWorkflowExecutionRequest{
			DomainID:   request.DomainID,
			WorkflowID: request.WorkflowID,
			RunID:      request.RunID,
		})
	}
	return backoff.Retry(op, persistenceOperationRetryPolicy, common.IsPersistenceTransientError)
}

func (c *workflowCleaner) deleteWorkflowHistory(request *CleanUpRequest) error {
	op := func() error {
		if request.EventStoreVersion == persistence.EventStoreVersionV2 {
			logger := c.logger.WithTags(tag.WorkflowID(request.WorkflowID),
				tag.WorkflowRunID(request.RunID),
				tag.WorkflowDomainID(request.DomainID),
				tag.ShardID(request.ShardID),
				tag.TaskID(request.TaskID),
				tag.FailoverVersion(request.FailoverVersion))
			return persistence.DeleteWorkflowExecutionHistoryV2(c.historyV2Manager, request.BranchToken, common.IntPtr(request.ShardID), logger)
		}
		return c.historyManager.DeleteWorkflowExecutionHistory(
			&persistence.DeleteWorkflowExecutionHistoryRequest{
				DomainID: request.DomainID,
				Execution: shared.WorkflowExecution{
					WorkflowId: common.StringPtr(request.WorkflowID),
					RunId:      common.StringPtr(request.RunID),
				},
			})
	}
	return backoff.Retry(op, persistenceOperationRetryPolicy, common.IsPersistenceTransientError)
}

func (c *workflowCleaner) deleteWorkflowVisibility(request *CleanUpRequest) error {
	op := func() error {
		return c.visibilityManager.DeleteWorkflowExecution(&persistence.VisibilityDeleteWorkflowExecutionRequest{
			DomainID:   request.DomainID,
			WorkflowID: request.WorkflowID,
			RunID:      request.RunID,
			TaskID:     request.TaskID,
		})
	}
	return backoff.Retry(op, persistenceOperationRetryPolicy, common.IsPersistenceTransientError)
}

func isNotExistError(err error) bool {
	_, ok := err.(*shared.EntityNotExistsError)
	return ok
}
