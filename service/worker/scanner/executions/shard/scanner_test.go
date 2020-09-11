// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package shard

import (
	"errors"
	"fmt"
	"testing"

	"github.com/uber/cadence/common/reconciliation/entity"
	"github.com/uber/cadence/common/reconciliation/invariant/check"

	"github.com/uber/cadence/common/pagination"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/common"
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
	mockItr := pagination.NewMockIterator(s.controller)
	mockItr.EXPECT().HasNext().Return(true).Times(1)
	mockItr.EXPECT().Next().Return(nil, errors.New("iterator error")).Times(1)

	scanner := &Scanner{
		shardID:          0,
		itr:              mockItr,
		progressReportFn: func() {},
	}
	result := scanner.Scan()
	s.Equal(common.ShardScanReport{
		ShardID: 0,
		Stats: common.ShardScanStats{
			CorruptionByType: make(map[check.InvariantType]int64),
		},
		Result: common.ShardScanResult{
			ControlFlowFailure: &common.ControlFlowFailure{
				Info:        "persistence iterator returned error",
				InfoDetails: "iterator error",
			},
		},
	}, result)
}

func (s *ScannerSuite) TestScan_Failure_NonFirstError() {
	mockItr := pagination.NewMockIterator(s.controller)
	iteratorCallNumber := 0
	mockItr.EXPECT().HasNext().DoAndReturn(func() bool {
		return iteratorCallNumber < 5
	}).Times(5)
	mockItr.EXPECT().Next().DoAndReturn(func() (*entity.ConcreteExecution, error) {
		defer func() {
			iteratorCallNumber++
		}()
		if iteratorCallNumber < 4 {
			return &entity.ConcreteExecution{}, nil
		}
		return nil, fmt.Errorf("iterator got error on: %v", iteratorCallNumber)
	}).Times(5)
	mockInvariantManager := common.NewMockInvariantManager(s.controller)
	mockInvariantManager.EXPECT().RunChecks(gomock.Any()).Return(check.ManagerCheckResult{
		CheckResultType: check.CheckResultTypeHealthy,
	}).Times(4)
	scanner := &Scanner{
		shardID:          0,
		itr:              mockItr,
		invariantManager: mockInvariantManager,
		progressReportFn: func() {},
	}
	result := scanner.Scan()
	s.Equal(common.ShardScanReport{
		ShardID: 0,
		Stats: common.ShardScanStats{
			ExecutionsCount:  4,
			CorruptionByType: make(map[check.InvariantType]int64),
		},
		Result: common.ShardScanResult{
			ControlFlowFailure: &common.ControlFlowFailure{
				Info:        "persistence iterator returned error",
				InfoDetails: "iterator got error on: 4",
			},
		},
	}, result)
}

func (s *ScannerSuite) TestScan_Failure_CorruptedWriterError() {
	mockItr := pagination.NewMockIterator(s.controller)
	mockItr.EXPECT().HasNext().Return(true).Times(1)
	mockItr.EXPECT().Next().Return(&entity.ConcreteExecution{}, nil).Times(1)
	mockInvariantManager := common.NewMockInvariantManager(s.controller)
	mockInvariantManager.EXPECT().RunChecks(gomock.Any()).Return(check.ManagerCheckResult{
		CheckResultType: check.CheckResultTypeCorrupted,
	}).Times(1)
	corruptedWriter := common.NewMockExecutionWriter(s.controller)
	corruptedWriter.EXPECT().Add(gomock.Any()).Return(errors.New("corrupted writer add failed")).Times(1)
	scanner := &Scanner{
		shardID:          0,
		itr:              mockItr,
		invariantManager: mockInvariantManager,
		corruptedWriter:  corruptedWriter,
		progressReportFn: func() {},
	}
	result := scanner.Scan()
	s.Equal(common.ShardScanReport{
		ShardID: 0,
		Stats: common.ShardScanStats{
			ExecutionsCount:  1,
			CorruptionByType: make(map[check.InvariantType]int64),
		},
		Result: common.ShardScanResult{
			ControlFlowFailure: &common.ControlFlowFailure{
				Info:        "blobstore add failed for corrupted execution check",
				InfoDetails: "corrupted writer add failed",
			},
		},
	}, result)
}

func (s *ScannerSuite) TestScan_Failure_FailedWriterError() {
	mockItr := pagination.NewMockIterator(s.controller)
	mockItr.EXPECT().HasNext().Return(true).Times(1)
	mockItr.EXPECT().Next().Return(&entity.ConcreteExecution{}, nil).Times(1)
	mockInvariantManager := common.NewMockInvariantManager(s.controller)
	mockInvariantManager.EXPECT().RunChecks(gomock.Any()).Return(check.ManagerCheckResult{
		CheckResultType: check.CheckResultTypeFailed,
	}).Times(1)
	failedWriter := common.NewMockExecutionWriter(s.controller)
	failedWriter.EXPECT().Add(gomock.Any()).Return(errors.New("failed writer add failed")).Times(1)
	scanner := &Scanner{
		shardID:          0,
		itr:              mockItr,
		invariantManager: mockInvariantManager,
		failedWriter:     failedWriter,
		progressReportFn: func() {},
	}
	result := scanner.Scan()
	s.Equal(common.ShardScanReport{
		ShardID: 0,
		Stats: common.ShardScanStats{
			ExecutionsCount:  1,
			CorruptionByType: make(map[check.InvariantType]int64),
		},
		Result: common.ShardScanResult{
			ControlFlowFailure: &common.ControlFlowFailure{
				Info:        "blobstore add failed for failed execution check",
				InfoDetails: "failed writer add failed",
			},
		},
	}, result)
}

func (s *ScannerSuite) TestScan_Failure_FailedWriterFlushError() {
	mockItr := pagination.NewMockIterator(s.controller)
	mockItr.EXPECT().HasNext().Return(false).Times(1)
	failedWriter := common.NewMockExecutionWriter(s.controller)
	failedWriter.EXPECT().Flush().Return(errors.New("failed writer flush failed")).Times(1)
	scanner := &Scanner{
		shardID:          0,
		itr:              mockItr,
		failedWriter:     failedWriter,
		progressReportFn: func() {},
	}
	result := scanner.Scan()
	s.Equal(common.ShardScanReport{
		ShardID: 0,
		Stats: common.ShardScanStats{
			ExecutionsCount:  0,
			CorruptionByType: make(map[check.InvariantType]int64),
		},
		Result: common.ShardScanResult{
			ControlFlowFailure: &common.ControlFlowFailure{
				Info:        "failed to flush for failed execution checks",
				InfoDetails: "failed writer flush failed",
			},
		},
	}, result)
}

func (s *ScannerSuite) TestScan_Failure_CorruptedWriterFlushError() {
	mockItr := pagination.NewMockIterator(s.controller)
	mockItr.EXPECT().HasNext().Return(false).Times(1)
	corruptedWriter := common.NewMockExecutionWriter(s.controller)
	corruptedWriter.EXPECT().Flush().Return(errors.New("corrupted writer flush failed")).Times(1)
	failedWriter := common.NewMockExecutionWriter(s.controller)
	failedWriter.EXPECT().Flush().Return(nil).Times(1)
	scanner := &Scanner{
		shardID:          0,
		itr:              mockItr,
		corruptedWriter:  corruptedWriter,
		failedWriter:     failedWriter,
		progressReportFn: func() {},
	}
	result := scanner.Scan()
	s.Equal(common.ShardScanReport{
		ShardID: 0,
		Stats: common.ShardScanStats{
			ExecutionsCount:  0,
			CorruptionByType: make(map[check.InvariantType]int64),
		},
		Result: common.ShardScanResult{
			ControlFlowFailure: &common.ControlFlowFailure{
				Info:        "failed to flush for corrupted execution checks",
				InfoDetails: "corrupted writer flush failed",
			},
		},
	}, result)
}

func (s *ScannerSuite) TestScan_Success() {
	mockItr := pagination.NewMockIterator(s.controller)
	iteratorCallNumber := 0
	mockItr.EXPECT().HasNext().DoAndReturn(func() bool {
		return iteratorCallNumber < 10
	}).Times(11)
	mockItr.EXPECT().Next().DoAndReturn(func() (*entity.ConcreteExecution, error) {
		defer func() {
			iteratorCallNumber++
		}()
		switch iteratorCallNumber {
		case 0, 1, 2, 3:
			return &entity.ConcreteExecution{
				Execution: entity.Execution{
					DomainID: "healthy",
				},
			}, nil
		case 4, 5, 6:
			return &entity.ConcreteExecution{
				Execution: entity.Execution{
					DomainID: "history_missing",
					State:    persistence.WorkflowStateCompleted,
				},
			}, nil
		case 7:
			return &entity.ConcreteExecution{
				Execution: entity.Execution{
					DomainID: "orphan_execution",
					State:    persistence.WorkflowStateCreated,
				},
			}, nil
		case 8, 9:
			return &entity.ConcreteExecution{
				Execution: entity.Execution{
					DomainID: "failed",
				},
			}, nil
		default:
			panic("should not get here")
		}
	}).Times(10)
	mockInvariantManager := common.NewMockInvariantManager(s.controller)
	mockInvariantManager.EXPECT().RunChecks(&entity.ConcreteExecution{
		Execution: entity.Execution{
			DomainID: "healthy",
		},
	}).Return(check.ManagerCheckResult{
		CheckResultType: check.CheckResultTypeHealthy,
	}).Times(4)
	mockInvariantManager.EXPECT().RunChecks(&entity.ConcreteExecution{
		Execution: entity.Execution{
			DomainID: "history_missing",
			State:    persistence.WorkflowStateCompleted,
		},
	}).Return(check.ManagerCheckResult{
		CheckResultType:          check.CheckResultTypeCorrupted,
		DeterminingInvariantType: check.InvariantTypePtr(check.HistoryExistsInvariantType),
		CheckResults: []check.CheckResult{
			{
				CheckResultType: check.CheckResultTypeCorrupted,
				InvariantType:   check.HistoryExistsInvariantType,
				Info:            "history did not exist",
			},
		},
	}).Times(3)
	mockInvariantManager.EXPECT().RunChecks(&entity.ConcreteExecution{
		Execution: entity.Execution{
			DomainID: "orphan_execution",
			State:    persistence.WorkflowStateCreated,
		}}).Return(check.ManagerCheckResult{
		CheckResultType:          check.CheckResultTypeCorrupted,
		DeterminingInvariantType: check.InvariantTypePtr(check.OpenCurrentExecutionInvariantType),
		CheckResults: []check.CheckResult{
			{
				CheckResultType: check.CheckResultTypeHealthy,
				InvariantType:   check.HistoryExistsInvariantType,
			},
			{
				CheckResultType: check.CheckResultTypeCorrupted,
				InvariantType:   check.OpenCurrentExecutionInvariantType,
				Info:            "execution was orphan",
			},
		},
	}).Times(1)
	mockInvariantManager.EXPECT().RunChecks(&entity.ConcreteExecution{
		Execution: entity.Execution{
			DomainID: "failed",
		}}).Return(check.ManagerCheckResult{
		CheckResultType:          check.CheckResultTypeFailed,
		DeterminingInvariantType: check.InvariantTypePtr(check.HistoryExistsInvariantType),
		CheckResults: []check.CheckResult{
			{
				CheckResultType: check.CheckResultTypeFailed,
				InvariantType:   check.HistoryExistsInvariantType,
				Info:            "failed to check if history exists",
			},
		},
	}).Times(2)

	mockCorruptedWriter := common.NewMockExecutionWriter(s.controller)
	mockCorruptedWriter.EXPECT().Add(common.ScanOutputEntity{
		Execution: &entity.ConcreteExecution{
			Execution: entity.Execution{
				DomainID: "history_missing",
				State:    persistence.WorkflowStateCompleted,
			}},
		Result: check.ManagerCheckResult{
			CheckResultType:          check.CheckResultTypeCorrupted,
			DeterminingInvariantType: check.InvariantTypePtr(check.HistoryExistsInvariantType),
			CheckResults: []check.CheckResult{
				{
					CheckResultType: check.CheckResultTypeCorrupted,
					InvariantType:   check.HistoryExistsInvariantType,
					Info:            "history did not exist",
				},
			},
		},
	}).Times(3)
	mockCorruptedWriter.EXPECT().Add(common.ScanOutputEntity{
		Execution: &entity.ConcreteExecution{
			Execution: entity.Execution{
				DomainID: "orphan_execution",
				State:    persistence.WorkflowStateCreated,
			}},
		Result: check.ManagerCheckResult{
			CheckResultType:          check.CheckResultTypeCorrupted,
			DeterminingInvariantType: check.InvariantTypePtr(check.OpenCurrentExecutionInvariantType),
			CheckResults: []check.CheckResult{
				{
					CheckResultType: check.CheckResultTypeHealthy,
					InvariantType:   check.HistoryExistsInvariantType,
				},
				{
					CheckResultType: check.CheckResultTypeCorrupted,
					InvariantType:   check.OpenCurrentExecutionInvariantType,
					Info:            "execution was orphan",
				},
			},
		},
	}).Times(1)
	mockFailedWriter := common.NewMockExecutionWriter(s.controller)
	mockFailedWriter.EXPECT().Add(common.ScanOutputEntity{
		Execution: &entity.ConcreteExecution{
			Execution: entity.Execution{
				DomainID: "failed",
			}},
		Result: check.ManagerCheckResult{
			CheckResultType:          check.CheckResultTypeFailed,
			DeterminingInvariantType: check.InvariantTypePtr(check.HistoryExistsInvariantType),
			CheckResults: []check.CheckResult{
				{
					CheckResultType: check.CheckResultTypeFailed,
					InvariantType:   check.HistoryExistsInvariantType,
					Info:            "failed to check if history exists",
				},
			},
		},
	}).Times(2)
	mockCorruptedWriter.EXPECT().Flush().Return(nil)
	mockFailedWriter.EXPECT().Flush().Return(nil)
	mockCorruptedWriter.EXPECT().FlushedKeys().Return(&common.Keys{UUID: "corrupt_keys_uuid"})
	mockFailedWriter.EXPECT().FlushedKeys().Return(&common.Keys{UUID: "failed_keys_uuid"})

	scanner := &Scanner{
		shardID:          0,
		invariantManager: mockInvariantManager,
		corruptedWriter:  mockCorruptedWriter,
		failedWriter:     mockFailedWriter,
		itr:              mockItr,
		progressReportFn: func() {},
	}
	result := scanner.Scan()
	s.Equal(common.ShardScanReport{
		ShardID: 0,
		Stats: common.ShardScanStats{
			ExecutionsCount:  10,
			CorruptedCount:   4,
			CheckFailedCount: 2,
			CorruptionByType: map[check.InvariantType]int64{
				check.HistoryExistsInvariantType:        3,
				check.OpenCurrentExecutionInvariantType: 1,
			},
			CorruptedOpenExecutionCount: 1,
		},
		Result: common.ShardScanResult{
			ScanKeys: &common.ShardScanKeys{
				Corrupt: &common.Keys{UUID: "corrupt_keys_uuid"},
				Failed:  &common.Keys{UUID: "failed_keys_uuid"},
			},
		},
	}, result)
}
