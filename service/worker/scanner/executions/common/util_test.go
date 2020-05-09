package common

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/codec"
	"github.com/uber/cadence/common/persistence"
	"testing"
)

const (
	domainID = "test-domain-id"
	workflowID = "test-workflow-id"
	runID = "test-run-id"
	treeID = "test-tree-id"
	branchID = "test-branch-id"
)

var (
	validBranchToken = []byte{89, 11, 0, 10, 0, 0, 0, 12, 116, 101, 115, 116, 45, 116, 114, 101, 101, 45, 105, 100, 11, 0, 20, 0, 0, 0, 14, 116 ,101, 115, 116, 45, 98, 114, 97, 110, 99, 104, 45, 105, 100, 0}
	invalidBranchToken = []byte("invalid")
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

func (s *UtilSuite) TestValidateExecution() {
	testCases := []struct{
		execution Execution
		expectError bool
	}{
		{
			execution: Execution{},
			expectError: true,
		},
		{
			execution: Execution{
				ShardID: -1,
			},
			expectError: true,
		},
		{
			execution: Execution{
				ShardID: 0,
			},
			expectError: true,
		},
		{
			execution: Execution{
				ShardID: 0,
				DomainID: domainID,
			},
			expectError: true,
		},
		{
			execution: Execution{
				ShardID: 0,
				DomainID: domainID,
				WorkflowID: workflowID,
			},
			expectError: true,
		},
		{
			execution: Execution{
				ShardID: 0,
				DomainID: domainID,
				WorkflowID: workflowID,
				RunID: runID,
			},
			expectError: true,
		},
		{
			execution: Execution{
				ShardID: 0,
				DomainID: domainID,
				WorkflowID: workflowID,
				RunID: runID,
				BranchToken: []byte{1, 2, 3},
			},
			expectError: true,
		},
		{
			execution: Execution{
				ShardID: 0,
				DomainID: domainID,
				WorkflowID: workflowID,
				RunID: runID,
				BranchToken: []byte{1, 2, 3},
				TreeID: treeID,
			},
			expectError: true,
		},
		{
			execution: Execution{
				ShardID: 0,
				DomainID: domainID,
				WorkflowID: workflowID,
				RunID: runID,
				BranchToken: []byte{1, 2, 3},
				TreeID: treeID,
				BranchID: branchID,
				State: persistence.WorkflowStateCreated - 1,
			},
			expectError: true,
		},
		{
			execution: Execution{
				ShardID: 0,
				DomainID: domainID,
				WorkflowID: workflowID,
				RunID: runID,
				BranchToken: []byte{1, 2, 3},
				TreeID: treeID,
				BranchID: branchID,
				State: persistence.WorkflowStateCorrupted + 1,
			},
			expectError: true,
		},
		{
			execution: Execution{
				ShardID: 0,
				DomainID: domainID,
				WorkflowID: workflowID,
				RunID: runID,
				BranchToken: []byte{1, 2, 3},
				TreeID: treeID,
				BranchID: branchID,
				State: persistence.WorkflowStateCreated,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		err := ValidateExecution(tc.execution)
		if tc.expectError {
			s.Error(err)
		} else {
			s.NoError(err)
		}
	}
}

func (s *UtilSuite) TestGetBranchToken() {
	encoder := codec.NewThriftRWEncoder()
	testCases := []struct{
		entity *persistence.ListConcreteExecutionsEntity
		expectError bool
		branchToken []byte
		treeID string
		branchID string
	}{
		{
			entity: &persistence.ListConcreteExecutionsEntity{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					BranchToken: s.getValidBranchToken(encoder),
				},
			},
			expectError: false,
			branchToken: validBranchToken,
			treeID: treeID,
			branchID: branchID,
		},
		{
			entity: &persistence.ListConcreteExecutionsEntity{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					BranchToken: invalidBranchToken,
				},
				VersionHistories: &persistence.VersionHistories{
					CurrentVersionHistoryIndex: 1,
					Histories: []*persistence.VersionHistory{
						{
							BranchToken: invalidBranchToken,
						},
						{
							BranchToken: validBranchToken,
						},
					},
				},
			},
			expectError: false,
			branchToken: validBranchToken,
			treeID: treeID,
			branchID: branchID,
		},
		{
			entity: &persistence.ListConcreteExecutionsEntity{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					BranchToken: invalidBranchToken,
				},
				VersionHistories: &persistence.VersionHistories{
					CurrentVersionHistoryIndex: 1,
					Histories: []*persistence.VersionHistory{
						{
							BranchToken: validBranchToken,
						},
						{
							BranchToken: invalidBranchToken,
						},
					},
				},
			},
			expectError: true,
		},
		{
			entity: &persistence.ListConcreteExecutionsEntity{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					BranchToken: invalidBranchToken,
				},
				VersionHistories: &persistence.VersionHistories{
					CurrentVersionHistoryIndex: 0,
					Histories: []*persistence.VersionHistory{},
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		branchToken, treeID, branchID, err := GetBranchToken(tc.entity, encoder)
		if tc.expectError {
			s.Error(err)
			s.Nil(branchToken)
			s.Empty(treeID)
			s.Empty(branchID)
		} else {
			s.NoError(err)
			s.Equal(tc.branchToken, branchToken)
			s.Equal(tc.treeID, treeID)
			s.Equal(tc.branchID, branchID)
		}
	}
}

func (s *UtilSuite) getValidBranchToken(encoder *codec.ThriftRWEncoder) []byte {
	hb := &shared.HistoryBranch{
		TreeID: common.StringPtr(treeID),
		BranchID: common.StringPtr(branchID),
	}
	bytes, err := encoder.Encode(hb)
	s.NoError(err)
	return bytes
}