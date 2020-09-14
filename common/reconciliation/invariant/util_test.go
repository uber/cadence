package invariant

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/entity"
)

type UtilSuite struct {
	*require.Assertions
	suite.Suite
}

func TestUtilSuite(t *testing.T) {
	suite.Run(t, new(UtilSuite))
}

func (s *UtilSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *UtilSuite) TestDeleteExecution() {
	testCases := []struct {
		deleteConcreteErr error
		deleteCurrentErr  error
		expectedFixResult *FixResult
	}{
		{
			deleteConcreteErr: errors.New("error deleting concrete execution"),
			expectedFixResult: &FixResult{
				FixResultType: FixResultTypeFailed,
				Info:          "failed to delete concrete workflow execution",
				InfoDetails:   "error deleting concrete execution",
			},
		},
		{
			deleteCurrentErr: errors.New("error deleting current execution"),
			expectedFixResult: &FixResult{
				FixResultType: FixResultTypeFailed,
				Info:          "failed to delete current workflow execution",
				InfoDetails:   "error deleting current execution",
			},
		},
		{
			expectedFixResult: &FixResult{
				FixResultType: FixResultTypeFixed,
			},
		},
	}

	for _, tc := range testCases {
		execManager := &mocks.ExecutionManager{}
		execManager.On("DeleteWorkflowExecution", mock.Anything).Return(tc.deleteConcreteErr).Once()
		if tc.deleteConcreteErr == nil {
			execManager.On("DeleteCurrentWorkflowExecution", mock.Anything).Return(tc.deleteCurrentErr).Once()
		}
		pr := persistence.NewPersistenceRetryer(execManager, nil, common.CreatePersistenceRetryPolicy())
		result := DeleteExecution(&entity.ConcreteExecution{}, pr)
		s.Equal(tc.expectedFixResult, result)
	}
}

func (s *UtilSuite) TestExecutionStillOpen() {
	testCases := []struct {
		getExecResp *persistence.GetWorkflowExecutionResponse
		getExecErr  error
		expectError bool
		expectOpen  bool
	}{
		{
			getExecResp: nil,
			getExecErr:  &shared.EntityNotExistsError{},
			expectError: false,
			expectOpen:  false,
		},
		{
			getExecResp: nil,
			getExecErr:  errors.New("got error"),
			expectError: true,
			expectOpen:  false,
		},
		{
			getExecResp: &persistence.GetWorkflowExecutionResponse{
				State: &persistence.WorkflowMutableState{
					ExecutionInfo: &persistence.WorkflowExecutionInfo{
						State: persistence.WorkflowStateCompleted,
					},
				},
			},
			getExecErr:  nil,
			expectError: false,
			expectOpen:  false,
		},
		{
			getExecResp: &persistence.GetWorkflowExecutionResponse{
				State: &persistence.WorkflowMutableState{
					ExecutionInfo: &persistence.WorkflowExecutionInfo{
						State: persistence.WorkflowStateCreated,
					},
				},
			},
			getExecErr:  nil,
			expectError: false,
			expectOpen:  true,
		},
	}

	for _, tc := range testCases {
		execManager := &mocks.ExecutionManager{}
		execManager.On("GetWorkflowExecution", mock.Anything).Return(tc.getExecResp, tc.getExecErr)
		pr := persistence.NewPersistenceRetryer(execManager, nil, common.CreatePersistenceRetryPolicy())
		open, err := ExecutionStillOpen(&entity.Execution{}, pr)
		if tc.expectError {
			s.Error(err)
		} else {
			s.NoError(err)
		}
		if tc.expectOpen {
			s.True(open)
		} else {
			s.False(open)
		}
	}
}

func (s *UtilSuite) TestExecutionStillExists() {
	testCases := []struct {
		getExecResp  *persistence.GetWorkflowExecutionResponse
		getExecErr   error
		expectError  bool
		expectExists bool
	}{
		{
			getExecResp:  &persistence.GetWorkflowExecutionResponse{},
			getExecErr:   nil,
			expectError:  false,
			expectExists: true,
		},
		{
			getExecResp:  nil,
			getExecErr:   &shared.EntityNotExistsError{},
			expectError:  false,
			expectExists: false,
		},
		{
			getExecResp:  nil,
			getExecErr:   errors.New("got error"),
			expectError:  true,
			expectExists: false,
		},
	}

	for _, tc := range testCases {
		execManager := &mocks.ExecutionManager{}
		execManager.On("GetWorkflowExecution", mock.Anything).Return(tc.getExecResp, tc.getExecErr)
		pr := persistence.NewPersistenceRetryer(execManager, nil, common.CreatePersistenceRetryPolicy())
		exists, err := ExecutionStillExists(&entity.Execution{}, pr)
		if tc.expectError {
			s.Error(err)
		} else {
			s.NoError(err)
		}
		if tc.expectExists {
			s.True(exists)
		} else {
			s.False(exists)
		}
	}
}
