package archiver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"github.com/uber/cadence/common/metrics"
	mmocks "github.com/uber/cadence/common/metrics/mocks"
	"github.com/uber/cadence/common/mocks"
	"go.uber.org/cadence"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/workflow"
)

var (
	archiverTestMetrics *mmocks.Client
	archiverTestLogger  *mocks.Logger
)

func init() {
	workflow.Register(handleRequestWorkflow)
}

type archiverSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestArchiverSuite(t *testing.T) {
	suite.Run(t, new(archiverSuite))
}

func (s *archiverSuite) SetupTest() {
	archiverTestMetrics = &mmocks.Client{}
	archiverTestMetrics.On("StartTimer", mock.Anything, mock.Anything).Return(tally.NewStopwatch(time.Now(), &nopStopwatchRecorder{}))
	archiverTestLogger = &mocks.Logger{}
	archiverTestLogger.On("WithFields", mock.Anything).Return(archiverTestLogger)
	archiverTestLogger.On("WithField", mock.Anything, mock.Anything).Return(archiverTestLogger).Maybe()
}

func (s *archiverSuite) TearDownTest() {
	archiverTestMetrics.AssertExpectations(s.T())
	archiverTestLogger.AssertExpectations(s.T())
}

func (s *archiverSuite) TestHandleRequest_UploadFails_NonRetryableError() {
	archiverTestMetrics.On("IncCounter", metrics.ArchiverScope, metrics.ArchiverUploadFailedAllRetriesCount).Once()
	archiverTestMetrics.On("IncCounter", metrics.ArchiverScope, metrics.ArchiverDeleteLocalSuccessCount).Once()
	archiverTestLogger.On("Error", mock.Anything).Once()

	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(uploadHistoryActivityFnName, mock.Anything, mock.Anything).Return(cadence.NewCustomError(errGetDomainByID))
	env.OnActivity(deleteHistoryActivityFnName, mock.Anything, mock.Anything).Return(nil)
	env.ExecuteWorkflow(handleRequestWorkflow, ArchiveRequest{})

	env.AssertExpectations(s.T())
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *archiverSuite) TestHandleRequest_UploadFails_ExpireRetryTimeout() {
	archiverTestMetrics.On("IncCounter", metrics.ArchiverScope, metrics.ArchiverUploadFailedAllRetriesCount).Once()
	archiverTestMetrics.On("IncCounter", metrics.ArchiverScope, metrics.ArchiverDeleteLocalSuccessCount).Once()
	archiverTestLogger.On("Error", mock.Anything).Once()

	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(uploadHistoryActivityFnName, mock.Anything, mock.Anything).Return(workflow.NewTimeoutError(shared.TimeoutTypeStartToClose))
	env.OnActivity(deleteHistoryActivityFnName, mock.Anything, mock.Anything).Return(nil)
	env.ExecuteWorkflow(handleRequestWorkflow, ArchiveRequest{})

	env.AssertExpectations(s.T())
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *archiverSuite) TestHandleRequest_LocalDeleteFails_NonRetryableError() {
	archiverTestMetrics.On("IncCounter", metrics.ArchiverScope, metrics.ArchiverUploadSuccessCount).Once()
	archiverTestMetrics.On("IncCounter", metrics.ArchiverScope, metrics.ArchiverDeleteLocalFailedAllRetriesCount).Once()
	archiverTestMetrics.On("IncCounter", metrics.ArchiverScope, metrics.ArchiverDeleteSuccessCount).Once()
	archiverTestLogger.On("Warn", mock.Anything).Once()

	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(uploadHistoryActivityFnName, mock.Anything, mock.Anything).Return(nil)
	var deleteSucceed bool
	env.OnActivity(deleteHistoryActivityFnName, mock.Anything, mock.Anything).Return(func(context.Context, ArchiveRequest) error {
		if !deleteSucceed {
			deleteSucceed = true
			return cadence.NewCustomError(errDeleteHistoryV1)
		}
		return nil
	})
	env.ExecuteWorkflow(handleRequestWorkflow, ArchiveRequest{})

	env.AssertExpectations(s.T())
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *archiverSuite) TestHandleRequest_LocalDeleteFailsThenSucceeds() {
	archiverTestMetrics.On("IncCounter", metrics.ArchiverScope, metrics.ArchiverUploadSuccessCount).Once()
	archiverTestMetrics.On("IncCounter", metrics.ArchiverScope, metrics.ArchiverDeleteLocalSuccessCount).Once()

	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(uploadHistoryActivityFnName, mock.Anything, mock.Anything).Return(nil)
	firstRun := true
	env.OnActivity(deleteHistoryActivityFnName, mock.Anything, mock.Anything).Return(func(context.Context, ArchiveRequest) error {
		if firstRun {
			firstRun = false
			return errors.New("some retryable error")
		}
		return nil
	})
	env.ExecuteWorkflow(handleRequestWorkflow, ArchiveRequest{})

	env.AssertExpectations(s.T())
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func handleRequestWorkflow(ctx workflow.Context, request ArchiveRequest) error {
	handleRequest(ctx, archiverTestLogger, archiverTestMetrics, request)
	return nil
}
