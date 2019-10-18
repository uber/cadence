package ndc

import (
	"github.com/pborman/uuid"
	"github.com/uber/cadence/.gen/go/cadence/workflowservicetest"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	test "github.com/uber/cadence/common/testing"
	"time"
)

func (s *nDCIntegrationTestSuite) TestReplicationMessageApplication() {

	workflowID := "replication-message-test" + uuid.New()
	runID := uuid.New()
	workflowType := "event-generator-workflow-type"
	tasklist := "event-generator-taskList"

	// active has initial version 0
	historyClient := s.active.GetHistoryClient()

	var historyBatch []*shared.History
	s.generator = test.InitializeHistoryEventGenerator(s.domainName, 1)

	for s.generator.HasNextVertex() {
		events := s.generator.GetNextVertices()
		historyEvents := &shared.History{}
		for _, event := range events {
			historyEvents.Events = append(historyEvents.Events, event.GetData().(*shared.HistoryEvent))
		}
		historyBatch = append(historyBatch, historyEvents)
	}

	versionHistory := s.eventBatchesToVersionHistory(nil, historyBatch)
	standbyClient := s.mockFrontendClient["standby"].(*workflowservicetest.MockClient)

	s.applyEventsThroughFetcher(
		workflowID,
		runID,
		workflowType,
		tasklist,
		versionHistory,
		historyBatch,
		historyClient,
		standbyClient,
	)

	time.Sleep(10 * time.Second)

	s.verifyEventHistory(workflowID, runID, historyBatch)
}

func (s *nDCIntegrationTestSuite) TestReplicationMessageDLQ() {

	workflowID := "replication-message-dlq-test" + uuid.New()
	runID := uuid.New()
	workflowType := "event-generator-workflow-type"
	tasklist := "event-generator-taskList"

	// active has initial version 0
	historyClient := s.active.GetHistoryClient()

	var historyBatch []*shared.History
	s.generator = test.InitializeHistoryEventGenerator(s.domainName, 1)

	for s.generator.HasNextVertex() {
		events := s.generator.GetNextVertices()
		historyEvents := &shared.History{}
		for _, event := range events {
			historyEvents.Events = append(historyEvents.Events, event.GetData().(*shared.HistoryEvent))
		}
		historyBatch = append(historyBatch, historyEvents)
	}

	versionHistory := s.eventBatchesToVersionHistory(nil, historyBatch)

	s.NotNil(historyBatch)
	historyBatch[0].Events[1].Version = common.Int64Ptr(2)
	standbyClient := s.mockFrontendClient["standby"].(*workflowservicetest.MockClient)

	s.applyEventsThroughFetcher(
		workflowID,
		runID,
		workflowType,
		tasklist,
		versionHistory,
		historyBatch,
		historyClient,
		standbyClient,
	)

	time.Sleep(10 * time.Second)

}
