package common

import (
	"errors"
	"fmt"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/codec"
	"github.com/uber/cadence/common/persistence"
)

// ValidateExecution returns an error if Execution is not valid, nil otherwise.
func ValidateExecution(execution *Execution) error {
	if execution.ShardID < 0 {
		return fmt.Errorf("invalid ShardID: %v", execution.ShardID)
	}
	if len(execution.DomainID) == 0 {
		return errors.New("empty DomainID")
	}
	if len(execution.WorkflowID) == 0 {
		return errors.New("empty WorkflowID")
	}
	if len(execution.RunID) == 0 {
		return errors.New("empty RunID")
	}
	if len(execution.BranchToken) == 0 {
		return errors.New("empty BranchToken")
	}
	if len(execution.TreeID) == 0 {
		return errors.New("empty TreeID")
	}
	if len(execution.BranchID) == 0 {
		return errors.New("empty BranchID")
	}
	if execution.State < persistence.WorkflowStateCreated || execution.State > persistence.WorkflowStateCorrupted {
		return fmt.Errorf("unknown workflow state: %v", execution.State)
	}
	return nil
}

// GetBranchToken returns the branchToken, treeID and branchID or error on failure.
func GetBranchToken(
	entity *persistence.ListConcreteExecutionsEntity,
	decoder *codec.ThriftRWEncoder,
) ([]byte, string, string, error) {
	branchToken := entity.ExecutionInfo.BranchToken
	if entity.VersionHistories != nil {
		versionHistory, err := entity.VersionHistories.GetCurrentVersionHistory()
		if err != nil {
			return nil, "", "", err
		}
		branchToken = versionHistory.GetBranchToken()
	}
	var branch shared.HistoryBranch
	if err := decoder.Decode(branchToken, &branch); err != nil {
		return nil, "", "", err
	}
	return branchToken, branch.GetTreeID(), branch.GetBranchID(), nil
}

// ExecutionStillOpen returns true if execution in persistence exists and is open, false otherwise.
// Returns error on failure to confirm.
// TODO: write unit tests for this method and methods below
func ExecutionStillOpen(
	exec *Execution,
	pr PersistenceRetryer,
) (bool, error) {
	req := &persistence.GetWorkflowExecutionRequest{
		DomainID: exec.DomainID,
		Execution: shared.WorkflowExecution{
			WorkflowId: &exec.WorkflowID,
			RunId:      &exec.RunID,
		},
	}
	resp, err := pr.GetWorkflowExecution(req)
	if err != nil {
		switch err.(type) {
		case *shared.EntityNotExistsError:
			return false, nil
		default:
			return false, err
		}
	}
	return Open(resp.State.ExecutionInfo.State), nil
}

// ExecutionStillExists returns true if execution still exists in persistence, false otherwise.
// Returns error on failure to confirm.
func ExecutionStillExists(
	exec *Execution,
	pr PersistenceRetryer,
) (bool, error) {
	req := &persistence.GetWorkflowExecutionRequest{
		DomainID: exec.DomainID,
		Execution: shared.WorkflowExecution{
			WorkflowId: &exec.WorkflowID,
			RunId:      &exec.RunID,
		},
	}
	_, err := pr.GetWorkflowExecution(req)
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

// Open returns true if workflow state is open false if workflow is closed
func Open(state int) bool {
	return state == persistence.WorkflowStateCreated || state == persistence.WorkflowStateRunning
}