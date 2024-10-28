package batcher

import (
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/testsuite"
	"testing"
)

type workflowSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	workflowEnv *testsuite.TestWorkflowEnvironment
}

func TestWorkflowSuite(t *testing.T) {
	suite.Run(t, new(workflowSuite))
}

func (s *workflowSuite) SetupTest() {
	s.workflowEnv = s.NewTestWorkflowEnvironment()
	s.workflowEnv.RegisterWorkflow(BatchWorkflow)

}

func (s *workflowSuite) TestWorkflow() {
	params := BatchParams{
		DomainName:               "test-domain",
		Query:                    "Closetime=missing",
		Reason:                   "unit-test",
		BatchType:                BatchTypeCancel,
		TerminateParams:          TerminateParams{},
		CancelParams:             CancelParams{},
		SignalParams:             SignalParams{},
		ReplicateParams:          ReplicateParams{},
		RPS:                      0,
		Concurrency:              5,
		PageSize:                 10,
		AttemptsOnRetryableError: 0,
		ActivityHeartBeatTimeout: 0,
		NonRetryableErrors:       []string{"HeartbeatTimeoutError"},
		_nonRetryableErrors:      nil,
	}
	activityHeartBeatDeatils := HeartBeatDetails{}
	s.workflowEnv.OnActivity(batchActivityName, mock.Anything, mock.Anything).Return(activityHeartBeatDeatils, nil)
	s.workflowEnv.ExecuteWorkflow(BatchWorkflow, params)
	s.True(s.workflowEnv.IsWorkflowCompleted())
	s.NoError(s.workflowEnv.GetWorkflowError())
}

func (s *workflowSuite) TestWorkflow_ValidationError() {
	params := BatchParams{
		DomainName: "test-domain",
		Query:      "",
		Reason:     "unit-test",
		BatchType:  BatchTypeCancel,
	}
	s.workflowEnv.ExecuteWorkflow(BatchWorkflow, params)
	s.True(s.workflowEnv.IsWorkflowCompleted())
	s.ErrorContains(s.workflowEnv.GetWorkflowError(), "must provide required parameters: BatchType/Reason/DomainName/Query")
}

func (s *workflowSuite) TestWorkflow_UnsupportedBatchType() {
	params := BatchParams{
		DomainName: "test-domain",
		Query:      "Closetime=missing",
		Reason:     "unit-test",
		BatchType:  "invalid",
	}
	s.workflowEnv.ExecuteWorkflow(BatchWorkflow, params)
	s.True(s.workflowEnv.IsWorkflowCompleted())
	s.ErrorContains(s.workflowEnv.GetWorkflowError(), "not supported batch type")
}

func (s *workflowSuite) TestWorkflow_BatchTypeSignalValidation() {
	params := BatchParams{
		DomainName: "test-domain",
		Query:      "Closetime=missing",
		Reason:     "unit-test",
		BatchType:  BatchTypeSignal,
	}
	s.workflowEnv.ExecuteWorkflow(BatchWorkflow, params)
	s.True(s.workflowEnv.IsWorkflowCompleted())
	s.ErrorContains(s.workflowEnv.GetWorkflowError(), "must provide signal name")
}

func (s *workflowSuite) TestWorkflow_BatchTypeReplicateSourceCLusterValidation() {
	paramsWithoutSourceCluster := BatchParams{
		DomainName: "test-domain",
		Query:      "Closetime=missing",
		Reason:     "unit-test",
		BatchType:  BatchTypeReplicate,
	}
	s.workflowEnv.ExecuteWorkflow(BatchWorkflow, paramsWithoutSourceCluster)
	s.True(s.workflowEnv.IsWorkflowCompleted())
	s.ErrorContains(s.workflowEnv.GetWorkflowError(), "must provide source cluster")
}

func (s *workflowSuite) TestWorkflow_BatchTypeReplicateTargetCLusterValidation() {
	paramsWithSourceCluster := BatchParams{
		DomainName: "test-domain",
		Query:      "Closetime=missing",
		Reason:     "unit-test",
		BatchType:  BatchTypeReplicate,
		ReplicateParams: ReplicateParams{
			SourceCluster: "test-dca",
			TargetCluster: "",
		},
	}
	s.workflowEnv.ExecuteWorkflow(BatchWorkflow, paramsWithSourceCluster)
	s.True(s.workflowEnv.IsWorkflowCompleted())
	s.ErrorContains(s.workflowEnv.GetWorkflowError(), "must provide target cluster")
}

func (s *workflowSuite) TearDownTest() {
	s.workflowEnv.AssertExpectations(s.T())
}
