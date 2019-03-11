package archiver

import (
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/uber-common/bark"
	"github.com/uber-go/tally"
	"github.com/uber/cadence/common/metrics"
	mmocks "github.com/uber/cadence/common/metrics/mocks"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/workflow"
	"testing"
	"time"
)


var archiverTestGlobalMetricsClient *mmocks.Client

func init() {
	workflow.Register(handleRequestWorkflow)
}

type archiverTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(archiverTestSuite))
}

func (s *archiverTestSuite) SetupTest() {
	archiverTestGlobalMetricsClient = &mmocks.Client{}
}

func (s *archiverTestSuite) TearDownTest() {
	archiverTestGlobalMetricsClient.AssertExpectations(s.T())
}

func (s *archiverTestSuite) TestHandleRequest_Something() {
	archiverTestGlobalMetricsClient.On("StartTimer", metrics.ArchiverScope, metrics.ArchiverHandleRequestLatency).Return(nopTallyStopwatch()).Once()
	archiverTestGlobalMetricsClient.On("StartTimer", metrics.ArchiverScope, metrics.ArchiverUploadWithRetriesLatency).Return(nopTallyStopwatch()).Once()
	archiverTestGlobalMetricsClient.On("IncCounter", metrics.ArchiverScope, metrics.ArchiverUploadSuccessCount).Once()
	archiverTestGlobalMetricsClient.On("StartTimer", metrics.ArchiverScope, metrics.ArchiverDeleteWithRetriesLatency).Return(nopTallyStopwatch()).Once()
	archiverTestGlobalMetricsClient.On("IncCounter", metrics.ArchiverScope, metrics.ArchiverDeleteLocalSuccessCount).Once()
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(uploadHistoryActivityFnName, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity(deleteHistoryActivityFnName, mock.Anything, mock.Anything).Return(nil)


	env.ExecuteWorkflow(handleRequestWorkflow, ArchiveRequest{})
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func handleRequestWorkflow(ctx workflow.Context, request ArchiveRequest) error {
	handleRequest(ctx, bark.NewNopLogger(), archiverTestGlobalMetricsClient, request)
	return nil
}

func nopTallyStopwatch() tally.Stopwatch {
	return tally.NewStopwatch(time.Now(), &nopStopwatchRecorder{})
}

/**

1. uploadHistoryActivityFails (failure case)
2. successfully delete history with local activity (success case)
3. failed to delete history with local activity but successful on normal (success case)
4. failed to delete history all together (failure case)


I need a mock for bark logger, need to be able to assert error logs
 */