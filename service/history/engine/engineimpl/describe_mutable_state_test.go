package engineimpl

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/execution/mocks"
)

func TestDescribeMutableState_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockExecutionCache := mocks.NewMockExecutionCache(ctrl)
	mockMutableState := mocks.NewMockMutableState(ctrl)

	historyEngine := &historyEngineImpl{
		executionCache: mockExecutionCache,
	}

	request := &types.DescribeMutableStateRequest{
		DomainUUID: "test-domain-uuid",
		Execution: &types.WorkflowExecution{
			WorkflowID: "test-workflow-id",
			RunID:      "test-run-id",
		},
	}

	mockExecutionCache.EXPECT().GetAndCreateWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, func(error) {}, true, nil)
	mockMutableState.EXPECT().CopyToPersistence().Return(&persistence.WorkflowMutableState{})

	response, err := historyEngine.DescribeMutableState(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestDescribeMutableState_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockExecutionCache := mocks.NewMockExecutionCache(ctrl)

	historyEngine := &historyEngineImpl{
		executionCache: mockExecutionCache,
	}

	request := &types.DescribeMutableStateRequest{
		DomainUUID: "test-domain-uuid",
		Execution: &types.WorkflowExecution{
			WorkflowID: "test-workflow-id",
			RunID:      "test-run-id",
		},
	}

	mockExecutionCache.EXPECT().GetAndCreateWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil, func(error) {}, true, common.ErrPersistenceNonRetryable)

	_, err := historyEngine.DescribeMutableState(context.Background(), request)

	assert.Error(t, err)
}
