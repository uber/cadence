package sysworkflow

import (
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber-common/bark"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	metricsMocks "github.com/uber/cadence/common/metrics/mocks"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service/dynamicconfig"
	"testing"
)

const (
	testDomainID          = "test-domain-id"
	testWorkflowID        = "test-workflow-id"
	testRunID             = "test-run-id"
	testEventStoreVersion = "test-event-store-version"
	testBranchToken       = "test-branch-token"
	testLastFirstEventID  = "test-last-first-event-id"
	testDomain            = "test-domain"
	testClusterName       = "test-cluster-name"
)

type HistoryBlobIteratorSuite struct {
	*require.Assertions
	suite.Suite
	logger               bark.Logger
	metricsClient        *metricsMocks.Client
	mockHistoryManager   *mocks.HistoryManager
	mockHistoryV2Manager *mocks.HistoryV2Manager
}

func TestHistoryBlobIteratorSuite(t *testing.T) {
	suite.Run(t, new(HistoryBlobIteratorSuite))
}

func (s *HistoryBlobIteratorSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.logger = bark.NewNopLogger()
	s.metricsClient = &metricsMocks.Client{}
	s.mockHistoryManager = &mocks.HistoryManager{}
	s.mockHistoryV2Manager = &mocks.HistoryV2Manager{}
}

func (s *HistoryBlobIteratorSuite) TestTest() {
	s.True(true)
}

func (s *HistoryBlobIteratorSuite) constructHistoryEvent(eventId int64) *shared.HistoryEvent {
	return &shared.HistoryEvent{
		EventId: common.Int64Ptr(eventId),
	}
}

func (s *HistoryBlobIteratorSuite) constructConfig(historyPageSize, targetArchivalBlobSize int) *Config {
	return &Config{
		HistoryPageSize:        dynamicconfig.GetIntPropertyFilteredByDomain(historyPageSize),
		TargetArchivalBlobSize: dynamicconfig.GetIntPropertyFilteredByDomain(targetArchivalBlobSize),
	}
}
