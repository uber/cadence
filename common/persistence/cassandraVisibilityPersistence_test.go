package persistence

import (
	"os"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	gen "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
)

type (
	visibilityPersistenceSuite struct {
		suite.Suite
		TestBase
		// override suite.Suite.Assertions with require.Assertions; this means that s.NotNil(nil) will stop the test,
		// not merely log an error
		*require.Assertions
	}
)

func TestVisibilityPersistenceSuite(t *testing.T) {
	s := new(visibilityPersistenceSuite)
	suite.Run(t, s)
}

func (s *visibilityPersistenceSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}

	s.SetupWorkflowStore()
}

func (s *visibilityPersistenceSuite) SetupTest() {
	// Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
	s.Assertions = require.New(s.T())
}

func (s *visibilityPersistenceSuite) TearDownSuite() {
	s.TearDownWorkflowStore()
}

func (s *visibilityPersistenceSuite) TestRecordWorkflowExecutionStarted() {
	workflowExecution := gen.WorkflowExecution{
		WorkflowId: common.StringPtr("visibility-workflow-test"),
		RunId:      common.StringPtr("fb15e4b5-356f-466d-8c6d-a29223e5c536"),
	}

	err := s.VisibilityMgr.RecordWorkflowExecutionStarted(&RecordWorkflowExecutionStartedRequest{
		Execution:        workflowExecution,
		WorkflowTypeName: "visibility-workflow",
		StartTimestamp:   time.Now(),
	})
	s.Nil(err)

	// TODO: validate by reading when List is implemented
}

func (s *visibilityPersistenceSuite) TestRecordWorkflowExecutionClosed() {
	workflowExecution := gen.WorkflowExecution{
		WorkflowId: common.StringPtr("visibility-workflow-test"),
		RunId:      common.StringPtr("cf555273-4443-456b-ad12-076625c50c23"),
	}
	startTimestamp := time.Now().Add(time.Second * -5)

	err0 := s.VisibilityMgr.RecordWorkflowExecutionStarted(&RecordWorkflowExecutionStartedRequest{
		Execution:        workflowExecution,
		WorkflowTypeName: "visibility-workflow",
		StartTimestamp:   startTimestamp,
	})
	s.Nil(err0)

	err1 := s.VisibilityMgr.RecordWorkflowExecutionClosed(&RecordWorkflowExecutionClosedRequest{
		Execution:        workflowExecution,
		WorkflowTypeName: "visibility-workflow",
		StartTimestamp:   startTimestamp,
		CloseTimestamp:   time.Now(),
	})
	s.Nil(err1)

	// TODO: validate by reading when List is implemented
}
