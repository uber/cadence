package invariants

import (
	"errors"
	"testing"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/service/worker/scanner/executions/common"
)

type HistoryExistsSuite struct {
	*require.Assertions
	suite.Suite
}

func TestHistoryExistsSuite(t *testing.T) {
	suite.Run(t, new(HistoryExistsSuite))
}

func (s *HistoryExistsSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *HistoryExistsSuite) TestCheck() {
	testCases := []struct {
		getExecErr                error
		getExecResp               *persistence.GetWorkflowExecutionResponse
		getHistoryErr             error
		getHistoryResp            *persistence.ReadHistoryBranchResponse
		expectedResult            common.CheckResult
		expectedResourcePopulated bool
	}{
		{
			getExecErr:     errors.New("got error checking workflow exists"),
			getHistoryResp: &persistence.ReadHistoryBranchResponse{},
			expectedResult: common.CheckResult{
				CheckResultType: common.CheckResultTypeFailed,
				Info:            "failed to check if concrete execution still exists",
				InfoDetails:     "got error checking workflow exists",
			},
			expectedResourcePopulated: false,
		},
		{
			getExecErr:     &shared.EntityNotExistsError{},
			getHistoryResp: &persistence.ReadHistoryBranchResponse{},
			expectedResult: common.CheckResult{
				CheckResultType: common.CheckResultTypeHealthy,
				Info:            "determined execution was healthy because concrete execution no longer exists",
			},
			expectedResourcePopulated: false,
		},
		{
			getExecResp:    &persistence.GetWorkflowExecutionResponse{},
			getHistoryResp: nil,
			getHistoryErr:  gocql.ErrNotFound,
			expectedResult: common.CheckResult{
				CheckResultType: common.CheckResultTypeCorrupted,
				Info:            "concrete execution exists but history does not exist",
				InfoDetails:     gocql.ErrNotFound.Error(),
			},
			expectedResourcePopulated: false,
		},
		{
			getExecResp:    &persistence.GetWorkflowExecutionResponse{},
			getHistoryResp: nil,
			getHistoryErr:  errors.New("error fetching history"),
			expectedResult: common.CheckResult{
				CheckResultType: common.CheckResultTypeFailed,
				Info:            "failed to verify if history exists",
				InfoDetails:     "error fetching history",
			},
			expectedResourcePopulated: false,
		},
		{
			getExecResp:    &persistence.GetWorkflowExecutionResponse{},
			getHistoryResp: nil,
			expectedResult: common.CheckResult{
				CheckResultType: common.CheckResultTypeCorrupted,
				Info:            "concrete execution exists but got empty history",
			},
			expectedResourcePopulated: false,
		},
		{
			getExecResp: &persistence.GetWorkflowExecutionResponse{},
			getHistoryResp: &persistence.ReadHistoryBranchResponse{
				HistoryEvents: []*shared.HistoryEvent{
					{},
				},
			},
			expectedResult: common.CheckResult{
				CheckResultType: common.CheckResultTypeHealthy,
			},
			expectedResourcePopulated: true,
		},
	}

	for _, tc := range testCases {
		execManager := &mocks.ExecutionManager{}
		historyManager := &mocks.HistoryV2Manager{}
		execManager.On("GetWorkflowExecution", mock.Anything).Return(tc.getExecResp, tc.getExecErr)
		historyManager.On("ReadHistoryBranch", mock.Anything).Return(tc.getHistoryResp, tc.getHistoryErr)
		i := NewHistoryExists(common.NewPersistenceRetryer(execManager, historyManager))
		resources := &common.InvariantResourceBag{}
		result := i.Check(getOpenExecution(), resources)
		s.Equal(tc.expectedResult, result)
		if tc.expectedResourcePopulated {
			s.NotNil(resources.History)
		} else {
			s.Nil(resources.History)
		}
	}
}
