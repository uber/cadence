package invariants

import (
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/service/worker/scanner/executions/common"
)

const (
	domainID    = "test-domain-id"
	workflowID  = "test-workflow-id"
	runID       = "test-run-id"
	shardID     = 0
	treeID      = "test-tree-id"
	branchID    = "test-branch-id"
	openState   = persistence.WorkflowStateCreated
	closedState = persistence.WorkflowStateCompleted
)

var (
	branchToken = []byte{1, 2, 3}
)

func getOpenExecution() common.Execution {
	return common.Execution{
		ShardID:     shardID,
		DomainID:    domainID,
		WorkflowID:  workflowID,
		RunID:       runID,
		BranchToken: branchToken,
		TreeID:      treeID,
		BranchID:    branchID,
		State:       openState,
	}
}

func getClosedExecution() common.Execution {
	return common.Execution{
		ShardID:     shardID,
		DomainID:    domainID,
		WorkflowID:  workflowID,
		RunID:       runID,
		BranchToken: branchToken,
		TreeID:      treeID,
		BranchID:    branchID,
		State:       closedState,
	}
}
