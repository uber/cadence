package cli

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/codec"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/quotas"
)

type (
	// CheckResultType is the result type of a check
	CheckResultType string
	// CheckType is the type of check
	CheckType string
)

const (
	// CheckResultTypeHealthy indicates check successfully ran and detected no corruption
	CheckResultTypeHealthy CheckResultType = "healthy"
	// CheckResultTypeCorruption indicates check successfully ran and detected corruption
	CheckResultTypeCorruption = "corruption"
	// CheckResultTypeFailure indicates check failed to run
	CheckResultTypeFailure = "failure"

	// CheckTypeHistoryExists is the check type for history exists
	CheckTypeHistoryExists CheckType = "history_exists"
	// CheckTypeValidFirstEvent is the check type for valid first event
	CheckTypeValidFirstEvent CheckType = "valid_first_event"
	// CheckTypeOrphanExecution is the check type for orphan execution
	CheckTypeOrphanExecution = "orphan_execution"

	historyPageSize = 1
)

type (
	AdminDBCheck interface {
		Check(*CheckRequest) *CheckResult
	}

	historyExistsCheck struct {
		branchDecoder *codec.ThriftRWEncoder
		shardID int
		dbRateLimiter *quotas.DynamicRateLimiter
		historyStore persistence.HistoryStore
		executionStore persistence.ExecutionStore
		payloadSerializer persistence.PayloadSerializer
	}

	firstHistoryEventCheck struct {
		branchDecoder *codec.ThriftRWEncoder
		shardID int
		dbRateLimiter *quotas.DynamicRateLimiter
		historyStore persistence.HistoryStore
		executionStore persistence.ExecutionStore
		payloadSerializer persistence.PayloadSerializer
	}

	orphanExecutionCheck struct {
		branchDecoder *codec.ThriftRWEncoder
		shardID int
		dbRateLimiter *quotas.DynamicRateLimiter
		historyStore persistence.HistoryStore
		executionStore persistence.ExecutionStore
		payloadSerializer persistence.PayloadSerializer
	}

	CheckRequest struct {
		ShardID int
		DomainID string
		WorkflowID string
		RunID string
		TreeID string
		BranchID string
		State int
		PrerequisiteCheckPayload interface{}
	}

	CheckResult struct {
		CheckType CheckType
		CheckResultType CheckResultType
		TotalDatabaseRequests int64
		Payload interface{}
		ErrorInfo *ErrorInfo
	}

	ErrorInfo struct {
		Note string
		Details string
	}
)

func NewHistoryExistsCheck(
	branchDecoder *codec.ThriftRWEncoder,
	shardID int,
	dbRateLimiter *quotas.DynamicRateLimiter,
	historyStore persistence.HistoryStore,
	executionStore persistence.ExecutionStore,
	payloadSerializer persistence.PayloadSerializer,
) AdminDBCheck {
	return &historyExistsCheck{
		branchDecoder: branchDecoder,
		shardID: shardID,
		dbRateLimiter: dbRateLimiter,
		historyStore: historyStore,
		executionStore: executionStore,
		payloadSerializer: payloadSerializer,
	}
}

func NewFirstHistoryEventCheck(
	branchDecoder *codec.ThriftRWEncoder,
	shardID int,
	dbRateLimiter *quotas.DynamicRateLimiter,
	historyStore persistence.HistoryStore,
	executionStore persistence.ExecutionStore,
	payloadSerializer persistence.PayloadSerializer,
) AdminDBCheck {
	return &firstHistoryEventCheck{
		branchDecoder: branchDecoder,
		shardID: shardID,
		dbRateLimiter: dbRateLimiter,
		historyStore: historyStore,
		executionStore: executionStore,
		payloadSerializer: payloadSerializer,
	}
}

func NewOrphanExecutionCheck(
	branchDecoder *codec.ThriftRWEncoder,
	shardID int,
	dbRateLimiter *quotas.DynamicRateLimiter,
	historyStore persistence.HistoryStore,
	executionStore persistence.ExecutionStore,
	payloadSerializer persistence.PayloadSerializer,
) AdminDBCheck {
	return &orphanExecutionCheck{
		branchDecoder: branchDecoder,
		shardID: shardID,
		dbRateLimiter: dbRateLimiter,
		historyStore: historyStore,
		executionStore: executionStore,
		payloadSerializer: payloadSerializer,
	}
}

func (c *historyExistsCheck) Check(cr *CheckRequest) *CheckResult {
	result := &CheckResult{
		CheckType: CheckTypeHistoryExists,
	}
	readHistoryBranchReq := &persistence.InternalReadHistoryBranchRequest{
		TreeID:    cr.TreeID,
		BranchID:  cr.BranchID,
		MinNodeID: common.FirstEventID,
		MaxNodeID: common.EndEventID,
		ShardID:   cr.ShardID,
		PageSize:  historyPageSize,
	}
	history, historyErr := retryReadHistoryBranch(
		c.dbRateLimiter,
		&result.TotalDatabaseRequests,
		c.historyStore,
		readHistoryBranchReq)

	stillExists, executionErr := concreteExecutionStillExists(
		cr,
		c.executionStore,
		c.dbRateLimiter,
		&result.TotalDatabaseRequests)

	if executionErr != nil {
		result.CheckResultType = CheckResultTypeFailure
		result.ErrorInfo = &ErrorInfo{
			Note: "failed to verify if concrete execution still exists",
			Details: executionErr.Error(),
		}
		return result
	}
	if !stillExists {
		result.CheckResultType = CheckResultTypeHealthy
		return result
	}

	if historyErr != nil {
		if historyErr == gocql.ErrNotFound {
			result.CheckResultType = CheckResultTypeCorruption
			result.ErrorInfo = &ErrorInfo{
				Note: "concrete execution exists but history does not",
				Details: historyErr.Error(),
			}
			return result
		}
		result.CheckResultType = CheckResultTypeFailure
		result.ErrorInfo = &ErrorInfo{
			Note: "failed to verify if history exists",
			Details: historyErr.Error(),
		}
		return result
	}
	if history == nil || len(history.History) == 0 {
		result.CheckResultType = CheckResultTypeCorruption
		result.ErrorInfo = &ErrorInfo{
			Note: "got empty history",
		}
		return result
	}
	result.CheckResultType = CheckResultTypeHealthy
	result.Payload = history
	return result
}

func (c *firstHistoryEventCheck) Check(cr *CheckRequest) *CheckResult {
	result := &CheckResult{
		CheckType: CheckTypeValidFirstEvent,
	}
	history := cr.PrerequisiteCheckPayload.(*persistence.InternalReadHistoryBranchResponse)
	firstBatch, err := c.payloadSerializer.DeserializeBatchEvents(history.History[0])
	if err != nil || len(firstBatch) == 0 {
		result.CheckResultType = CheckResultTypeFailure
		result.ErrorInfo = &ErrorInfo{
			Note: "failed to deserialize batch events",
			Details: err.Error(),
		}
		return result
	}
	if firstBatch[0].GetEventId() != common.FirstEventID {
		result.CheckResultType = CheckResultTypeCorruption
		result.ErrorInfo = &ErrorInfo{
			Note: "got unexpected first eventID",
			Details:        fmt.Sprintf("expected: %v but got %v", common.FirstEventID, firstBatch[0].GetEventId()),
		}
		return result
	}
	if firstBatch[0].GetEventType() != shared.EventTypeWorkflowExecutionStarted {
		result.CheckResultType = CheckResultTypeCorruption
		result.ErrorInfo = &ErrorInfo{
			Note: "got unexpected first eventType",
			Details:        fmt.Sprintf("expected: %v but got %v", shared.EventTypeWorkflowExecutionStarted.String(), firstBatch[0].GetEventType().String()),
		}
		return result
	}
	result.CheckResultType = CheckResultTypeHealthy
	return result
}

func (c *orphanExecutionCheck) Check(cr *CheckRequest) *CheckResult {
	result := &CheckResult{
		CheckType: CheckTypeOrphanExecution,
	}
	if !executionOpen(cr) {
		result.CheckResultType = CheckResultTypeHealthy
		return result
	}
	getCurrentExecutionRequest := &persistence.GetCurrentExecutionRequest{
		DomainID:   cr.DomainID,
		WorkflowID: cr.WorkflowID,
	}
	currentExecution, currentErr := retryGetCurrentExecution(
		c.dbRateLimiter,
		&result.TotalDatabaseRequests,
		c.executionStore,
		getCurrentExecutionRequest)

	stillOpen, concreteErr := concreteExecutionStillOpen(cr, c.executionStore, c.dbRateLimiter, &result.TotalDatabaseRequests)
	if concreteErr != nil {
		result.CheckResultType = CheckResultTypeFailure
		result.ErrorInfo = &ErrorInfo{
			Note: "failed to check if concrete execution is still open",
			Details: concreteErr.Error(),
		}
		return result
	}
	if !stillOpen {
		result.CheckResultType = CheckResultTypeHealthy
		return result
	}

	if currentErr != nil {
		switch currentErr.(type) {
		case *shared.EntityNotExistsError:
			result.CheckResultType = CheckResultTypeCorruption
			result.ErrorInfo = &ErrorInfo{
				Note: "execution is open without having current execution",
				Details: currentErr.Error(),
			}
			return result
		}
		result.CheckResultType = CheckResultTypeFailure
		result.ErrorInfo = &ErrorInfo{
			Note: "failed to check if current execution exists",
			Details: currentErr.Error(),
		}
		return result
	}
	if currentExecution.RunID != cr.RunID {
		result.CheckResultType = CheckResultTypeCorruption
		result.ErrorInfo = &ErrorInfo{
			Note: "execution is open but current points at a different execution",
		}
		return result
	}
	result.CheckResultType = CheckResultTypeHealthy
	return result
}

func concreteExecutionStillOpen(
	cr *CheckRequest,
	execStore persistence.ExecutionStore,
	limiter *quotas.DynamicRateLimiter,
	totalDBRequests *int64,
) (bool, error) {
	getConcreteExecution := &persistence.GetWorkflowExecutionRequest{
		DomainID: cr.DomainID,
		Execution: shared.WorkflowExecution{
			WorkflowId: &cr.WorkflowID,
			RunId:      &cr.RunID,
		},
	}
	_, err := retryGetWorkflowExecution(limiter, totalDBRequests, execStore, getConcreteExecution)

	if err != nil {
		switch err.(type) {
		case *shared.EntityNotExistsError:
			return false, nil
		default:
			return false, err
		}
	}
	return executionOpen(cr), nil
}

func concreteExecutionStillExists(
	cr *CheckRequest,
	execStore persistence.ExecutionStore,
	limiter *quotas.DynamicRateLimiter,
	totalDBRequests *int64,
) (bool, error) {
	getConcreteExecution := &persistence.GetWorkflowExecutionRequest{
		DomainID: cr.DomainID,
		Execution: shared.WorkflowExecution{
			WorkflowId: &cr.WorkflowID,
			RunId:      &cr.RunID,
		},
	}
	_, err := retryGetWorkflowExecution(limiter, totalDBRequests, execStore, getConcreteExecution)
	if err == nil {
		return true, nil
	}

	switch err.(type) {
	case *shared.EntityNotExistsError:
		return false, nil
	default:
		return false, err
	}
}

func executionOpen(cr *CheckRequest) bool {
	return cr.State == persistence.WorkflowStateCreated ||
		cr.State == persistence.WorkflowStateRunning
}