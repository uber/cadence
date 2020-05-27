package shard

import (
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber/cadence/service/worker/scanner/executions/common"
	"testing"
)

type ScannerSuite struct {
	*require.Assertions
	suite.Suite
	controller *gomock.Controller
}

func TestScannerSuite(t *testing.T) {
	suite.Run(t, new(ScannerSuite))
}

func (s *ScannerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.controller = gomock.NewController(s.T())
}

func (s *ScannerSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *ScannerSuite) TestScan_Failure_FirstIteratorError() {
	mockItr := common.NewMockExecutionIterator(s.controller)
	mockItr.EXPECT().HasNext().Return(true).Times(1)
	mockItr.EXPECT().Next().Return(nil, errors.New("iterator error")).Times(1)
	scanner := &scanner{
		shardID: 0,
		itr: mockItr,
	}
	result := scanner.Scan()
	s.Equal(common.ShardScanReport{
		ShardID: 0,
		Stats: common.ShardScanStats{
			CorruptionByType: make(map[common.InvariantType]int64),
		},
		Result: common.ShardScanResult{
			ControlFlowFailure: &common.ControlFlowFailure{
				Info: "persistence iterator returned error",
				InfoDetails: "iterator error",
			},
		},
	}, result)
}

func (s *ScannerSuite) TestScan_Failure_NonFirstError() {
	mockItr := common.NewMockExecutionIterator(s.controller)
	iteratorCallNumber := 0
	mockItr.EXPECT().HasNext().DoAndReturn(func() bool {
		defer func() {
			iteratorCallNumber++
		}()
		return iteratorCallNumber < 5
	}).Times(5)
	mockItr.EXPECT().Next().DoAndReturn(func() (*common.Execution, error) {
		if iteratorCallNumber < 4 {
			return &common.Execution{}, nil
		}
		return nil, fmt.Errorf("iterator got error on: %v", iteratorCallNumber)
	}).Times(5)
	mockInvariantManager := common.NewMockInvariantManager(s.controller)
	mockInvariantManager.EXPECT().RunChecks(gomock.Any()).Return(common.CheckResult{
		CheckResultType: common.CheckResultTypeHealthy,
	}).Times(4)
	scanner := &scanner{
		shardID: 0,
		itr: mockItr,
		invariantManager: mockInvariantManager,
	}
	result := scanner.Scan()
	s.Equal(common.ShardScanReport{
		ShardID: 0,
		Stats: common.ShardScanStats{
			ExecutionsCount: 4,
			CorruptionByType: make(map[common.InvariantType]int64),
		},
		Result: common.ShardScanResult{
			ControlFlowFailure: &common.ControlFlowFailure{
				Info: "persistence iterator returned error",
				InfoDetails: "iterator error",
			},
		},
	}, result)
}

