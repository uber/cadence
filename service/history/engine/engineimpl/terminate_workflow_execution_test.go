// Filename: terminate_workflow_execution_test.go
// Place this file in the appropriate package directory within your Cadence project

package engineimpl

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/uber/cadence/service/history/constants"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/engine/testdata"
)

// Helper function to setup the state of a workflow in tests
func withTerminationState(execution *types.WorkflowExecution, state *persistence.WorkflowMutableState) func(*testing.T, *testdata.EngineForTest) {
	return func(t *testing.T, engine *testdata.EngineForTest) {
		engine.ShardCtx.Resource.ExecutionMgr.On("GetWorkflowExecution", mock.Anything, mock.MatchedBy(func(req *persistence.GetWorkflowExecutionRequest) bool {
			return req.Execution == *execution
		})).Return(&persistence.GetWorkflowExecutionResponse{
			State: state,
		}, nil)
	}
}

// Creates a request object for the TerminateWorkflowExecution function
func terminationExecutionRequest(execution *types.WorkflowExecution) *types.HistoryTerminateWorkflowExecutionRequest {
	return &types.HistoryTerminateWorkflowExecutionRequest{
		DomainUUID: constants.TestDomainID,
		TerminateRequest: &types.TerminateWorkflowExecutionRequest{
			Domain:            constants.TestDomainName,
			WorkflowExecution: execution,
			Reason:            "Test termination reason",
			Identity:          "Test identity",
		},
	}
}

// Main test function for TerminateWorkflowExecution
func TestTerminateWorkflowExecution(t *testing.T) {
	execution := &types.WorkflowExecution{WorkflowID: "testWorkflowID", RunID: constants.TestRunID}
	state := &persistence.WorkflowMutableState{
		ExecutionInfo: &persistence.WorkflowExecutionInfo{
			WorkflowID: "testWorkflowID",
			RunID:      constants.TestRunID,
			State:      persistence.WorkflowStateRunning,
		},
	}

	eft := testdata.NewEngineForTest(t, NewEngineWithShardContext)
	eft.Engine.Start()
	defer eft.Engine.Stop()
	initFn := withTerminationState(execution, state)
	initFn(t, eft) // Initialize the test with the setup

	request := terminationExecutionRequest(execution)
	err := eft.Engine.TerminateWorkflowExecution(context.Background(), request)

	assert.NoError(t, err)
}
