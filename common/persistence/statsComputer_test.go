package persistence

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
)

type (
	statsComputerSuite struct {
		sc *statsComputer
		suite.Suite
		// override suite.Suite.Assertions with require.Assertions; this means that s.NotNil(nil) will stop the test,
		// not merely log an error
		*require.Assertions
	}
)

func TestStatsComputerSuite(t *testing.T) {
	s := new(historySerializerSuite)
	suite.Run(t, s)
}

//TODO need to add more tests
func (s *statsComputerSuite) SetupTest() {
	// Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
	s.Assertions = require.New(s.T())
	s.sc = &statsComputer{}
}

func (s *statsComputerSuite) createRequest() *PersistenceUpdateWorkflowExecutionRequest {
	return &PersistenceUpdateWorkflowExecutionRequest{
		ExecutionInfo: &PersistenceWorkflowExecutionInfo{},
	}
}

func (s *statsComputerSuite) TestStatsWithStartedEvent() {
	ms := s.createRequest()
	domainID := "A"
	execution := workflow.WorkflowExecution{
		WorkflowId: common.StringPtr("test-workflow-id"),
		RunId:      common.StringPtr("run_id"),
	}
	workflowType := &workflow.WorkflowType{
		Name: common.StringPtr("test-workflow-type-name"),
	}
	tasklist := &workflow.TaskList{
		Name: common.StringPtr("test-tasklist"),
	}

	ms.ExecutionInfo.DomainID = domainID
	ms.ExecutionInfo.WorkflowID = *execution.WorkflowId
	ms.ExecutionInfo.RunID = *execution.RunId
	ms.ExecutionInfo.WorkflowTypeName = *workflowType.Name
	ms.ExecutionInfo.TaskList = *tasklist.Name

	expectedSize := len(execution.GetWorkflowId()) + len(workflowType.GetName()) + len(tasklist.GetName())

	stats := s.sc.computeMutableStateStats(ms)
	s.Equal(stats.ExecutionInfoSize, expectedSize)
}
