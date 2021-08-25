package testing

import (
	"github.com/pborman/uuid"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/constants"
	"github.com/uber/cadence/service/history/execution"
	"github.com/uber/cadence/service/history/shard"
)

// SetupWorkflow setup a workflow for testing purpose
func SetupWorkflow(
	mockShard *shard.TestContext,
	sourceDomainID string,
) (types.WorkflowExecution, execution.MutableState, int64, error) {
	workflowExecution := types.WorkflowExecution{
		WorkflowID: constants.TestWorkflowID,
		RunID:      constants.TestRunID,
	}
	workflowType := "some random workflow type"
	taskListName := "some random task list"

	entry, err := mockShard.GetDomainCache().GetDomainByID(sourceDomainID)
	if err != nil {
		return types.WorkflowExecution{}, nil, 0, err
	}
	version := entry.GetFailoverVersion()
	mutableState := execution.NewMutableStateBuilderWithVersionHistoriesWithEventV2(
		mockShard,
		mockShard.GetLogger(),
		version,
		workflowExecution.GetRunID(),
		entry,
	)
	_, err = mutableState.AddWorkflowExecutionStartedEvent(
		workflowExecution,
		&types.HistoryStartWorkflowExecutionRequest{
			DomainUUID: sourceDomainID,
			StartRequest: &types.StartWorkflowExecutionRequest{
				WorkflowType:                        &types.WorkflowType{Name: workflowType},
				TaskList:                            &types.TaskList{Name: taskListName},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(2),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
			},
		},
	)
	if err != nil {
		return types.WorkflowExecution{}, nil, 0, err
	}

	di := AddDecisionTaskScheduledEvent(mutableState)
	event := AddDecisionTaskStartedEvent(mutableState, di.ScheduleID, taskListName, uuid.New())
	di.StartedID = event.GetEventID()
	event = AddDecisionTaskCompletedEvent(mutableState, di.ScheduleID, di.StartedID, nil, "some random identity")

	return workflowExecution, mutableState, event.GetEventID(), nil
}

// CreatePersistenceMutableState generated a persistence representation of the mutable state
// a based on the in memory version
func CreatePersistenceMutableState(
	ms execution.MutableState,
	lastEventID int64,
	lastEventVersion int64,
) (*persistence.WorkflowMutableState, error) {

	if ms.GetVersionHistories() != nil {
		currentVersionHistory, err := ms.GetVersionHistories().GetCurrentVersionHistory()
		if err != nil {
			return nil, err
		}

		err = currentVersionHistory.AddOrUpdateItem(persistence.NewVersionHistoryItem(
			lastEventID,
			lastEventVersion,
		))
		if err != nil {
			return nil, err
		}
	}

	return execution.CreatePersistenceMutableState(ms), nil
}
