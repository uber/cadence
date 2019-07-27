// Copyright (c) 2019 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package history

import (
	ctx "context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"

	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/clock"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/service"
)

type (
	nDCTransactionMgrForNewWorkflowSuite struct {
		suite.Suite

		logger              log.Logger
		mockExecutionMgr    *mocks.ExecutionManager
		mockHistoryV2Mgr    *mocks.HistoryV2Manager
		mockClusterMetadata *mocks.ClusterMetadata
		mockMetadataMgr     *mocks.MetadataManager
		mockService         service.Service
		mockShard           *shardContextImpl
		mockDomainCache     *cache.DomainCacheMock

		mockTransactionMgr *mockNDCTransactionMgr
		createMgr          *nDCTransactionMgrForNewWorkflowImpl
	}
)

func TestNDCTransactionMgrForNewWorkflowSuite(t *testing.T) {
	s := new(nDCTransactionMgrForNewWorkflowSuite)
	suite.Run(t, s)
}

func (s *nDCTransactionMgrForNewWorkflowSuite) SetupTest() {
	s.logger = loggerimpl.NewDevelopmentForTest(s.Suite)
	s.mockHistoryV2Mgr = &mocks.HistoryV2Manager{}
	s.mockExecutionMgr = &mocks.ExecutionManager{}
	s.mockClusterMetadata = &mocks.ClusterMetadata{}
	s.mockMetadataMgr = &mocks.MetadataManager{}
	metricsClient := metrics.NewClient(tally.NoopScope, metrics.History)
	s.mockService = service.NewTestService(s.mockClusterMetadata, nil, metricsClient, nil)
	s.mockDomainCache = &cache.DomainCacheMock{}

	s.mockShard = &shardContextImpl{
		service:                   s.mockService,
		shardInfo:                 &persistence.ShardInfo{ShardID: 10, RangeID: 1, TransferAckLevel: 0},
		transferSequenceNumber:    1,
		executionManager:          s.mockExecutionMgr,
		historyV2Mgr:              s.mockHistoryV2Mgr,
		maxTransferSequenceNumber: 100000,
		closeCh:                   make(chan int, 100),
		config:                    NewDynamicConfigForTest(),
		logger:                    s.logger,
		domainCache:               s.mockDomainCache,
		metricsClient:             metricsClient,
		timeSource:                clock.NewRealTimeSource(),
	}
	s.mockTransactionMgr = &mockNDCTransactionMgr{}
	s.createMgr = newNDCTransactionMgrForNewWorkflow(s.mockTransactionMgr)
}

func (s *nDCTransactionMgrForNewWorkflowSuite) TearDownTest() {
	s.mockHistoryV2Mgr.AssertExpectations(s.T())
	s.mockExecutionMgr.AssertExpectations(s.T())
	s.mockMetadataMgr.AssertExpectations(s.T())
	s.mockDomainCache.AssertExpectations(s.T())
}

func (s *nDCTransactionMgrForNewWorkflowSuite) TestDispatchForNewWorkflow_Dup() {
	ctx := ctx.Background()
	now := time.Now()

	domainID := "some random domain ID"
	workflowID := "some random workflow ID"
	runID := "some random run ID"

	workflow := &mockNDCWorkflow{}
	defer workflow.AssertExpectations(s.T())
	mutableState := &mockMutableState{}
	defer mutableState.AssertExpectations(s.T())
	workflow.On("getMutableState").Return(mutableState)

	mutableState.On("GetExecutionInfo").Return(&persistence.WorkflowExecutionInfo{
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
	})

	s.mockTransactionMgr.On("getCurrentWorkflowRunID", ctx, domainID, workflowID).Return(runID, nil).Once()

	err := s.createMgr.dispatchForNewWorkflow(ctx, now, workflow)
	s.NoError(err)
}

func (s *nDCTransactionMgrForNewWorkflowSuite) TestDispatchForNewWorkflow_BrandNew() {
	ctx := ctx.Background()
	now := time.Now()

	domainID := "some random domain ID"
	workflowID := "some random workflow ID"
	runID := "some random run ID"

	releaseCalled := false

	workflow := &mockNDCWorkflow{}
	defer workflow.AssertExpectations(s.T())
	context := &mockWorkflowExecutionContext{}
	defer context.AssertExpectations(s.T())
	mutableState := &mockMutableState{}
	defer mutableState.AssertExpectations(s.T())
	var releaseFn releaseWorkflowExecutionFunc = func(error) { releaseCalled = true }
	workflow.On("getContext").Return(context)
	workflow.On("getMutableState").Return(mutableState)
	workflow.On("getReleaseFn").Return(releaseFn)

	workflowSnapshot := &persistence.WorkflowSnapshot{}
	workflowEventsSeq := []*persistence.WorkflowEvents{&persistence.WorkflowEvents{}}
	workflowHistorySize := int64(12345)
	mutableState.On("GetExecutionInfo").Return(&persistence.WorkflowExecutionInfo{
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      runID,
	})
	mutableState.On("CloseTransactionAsSnapshot", now, transactionPolicyPassive).Return(
		workflowSnapshot, workflowEventsSeq, nil,
	).Once()

	s.mockTransactionMgr.On("getCurrentWorkflowRunID", ctx, domainID, workflowID).Return("", nil).Once()

	context.On(
		"persistFirstWorkflowEvents", workflowEventsSeq[0],
	).Return(workflowHistorySize, nil).Once()
	context.On(
		"createWorkflowExecution",
		workflowSnapshot,
		workflowHistorySize,
		now,
		persistence.CreateWorkflowModeBrandNew,
		"",
		int64(0),
	).Return(nil).Once()

	err := s.createMgr.dispatchForNewWorkflow(ctx, now, workflow)
	s.NoError(err)
	s.True(releaseCalled)
}

func (s *nDCTransactionMgrForNewWorkflowSuite) TestDispatchForNewWorkflow_CreateAsCurrent() {
	ctx := ctx.Background()
	now := time.Now()

	domainID := "some random domain ID"
	workflowID := "some random workflow ID"
	targetRunID := "some random run ID"
	currentRunID := "other random runID"
	currentLastWriteVersion := int64(4321)

	targetReleaseCalled := false
	currentReleaseCalled := false

	targetWorkflow := &mockNDCWorkflow{}
	defer targetWorkflow.AssertExpectations(s.T())
	targetContext := &mockWorkflowExecutionContext{}
	defer targetContext.AssertExpectations(s.T())
	targetMutableState := &mockMutableState{}
	defer targetMutableState.AssertExpectations(s.T())
	var targetReleaseFn releaseWorkflowExecutionFunc = func(error) { targetReleaseCalled = true }
	targetWorkflow.On("getContext").Return(targetContext)
	targetWorkflow.On("getMutableState").Return(targetMutableState)
	targetWorkflow.On("getReleaseFn").Return(targetReleaseFn)

	currentWorkflow := &mockNDCWorkflow{}
	defer currentWorkflow.AssertExpectations(s.T())
	currentContext := &mockWorkflowExecutionContext{}
	defer currentContext.AssertExpectations(s.T())
	currentMutableState := &mockMutableState{}
	defer currentMutableState.AssertExpectations(s.T())
	var currentReleaseFn releaseWorkflowExecutionFunc = func(error) { currentReleaseCalled = true }
	currentWorkflow.On("getMutableState").Return(currentMutableState)
	currentWorkflow.On("getReleaseFn").Return(currentReleaseFn)

	targetWorkflowSnapshot := &persistence.WorkflowSnapshot{}
	targetWorkflowEventsSeq := []*persistence.WorkflowEvents{&persistence.WorkflowEvents{}}
	targetWorkflowHistorySize := int64(12345)
	targetMutableState.On("GetExecutionInfo").Return(&persistence.WorkflowExecutionInfo{
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      targetRunID,
	})
	targetMutableState.On("CloseTransactionAsSnapshot", now, transactionPolicyPassive).Return(
		targetWorkflowSnapshot, targetWorkflowEventsSeq, nil,
	).Once()

	s.mockTransactionMgr.On("getCurrentWorkflowRunID", ctx, domainID, workflowID).Return(currentRunID, nil).Once()
	s.mockTransactionMgr.On("loadNDCWorkflow", ctx, domainID, workflowID, currentRunID).Return(currentWorkflow, nil).Once()

	targetWorkflow.On("happensAfter", currentWorkflow).Return(true, nil)
	currentMutableState.On("IsWorkflowExecutionRunning").Return(false)
	currentMutableState.On("GetExecutionInfo").Return(&persistence.WorkflowExecutionInfo{
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      currentRunID,
	})
	currentWorkflow.On("getVectorClock").Return(currentLastWriteVersion, int64(0), nil)

	targetContext.On(
		"persistFirstWorkflowEvents", targetWorkflowEventsSeq[0],
	).Return(targetWorkflowHistorySize, nil).Once()
	targetContext.On(
		"createWorkflowExecution",
		targetWorkflowSnapshot,
		targetWorkflowHistorySize,
		now,
		persistence.CreateWorkflowModeWorkflowIDReuse,
		currentRunID,
		currentLastWriteVersion,
	).Return(nil).Once()

	err := s.createMgr.dispatchForNewWorkflow(ctx, now, targetWorkflow)
	s.NoError(err)
	s.True(targetReleaseCalled)
	s.True(currentReleaseCalled)
}

func (s *nDCTransactionMgrForNewWorkflowSuite) TestDispatchForNewWorkflow_CreateAsZombie() {
	ctx := ctx.Background()
	now := time.Now()

	domainID := "some random domain ID"
	workflowID := "some random workflow ID"
	targetRunID := "some random run ID"
	currentRunID := "other random runID"

	targetReleaseCalled := false
	currentReleaseCalled := false

	targetWorkflow := &mockNDCWorkflow{}
	defer targetWorkflow.AssertExpectations(s.T())
	targetContext := &mockWorkflowExecutionContext{}
	defer targetContext.AssertExpectations(s.T())
	targetMutableState := &mockMutableState{}
	defer targetMutableState.AssertExpectations(s.T())
	var targetReleaseFn releaseWorkflowExecutionFunc = func(error) { targetReleaseCalled = true }
	targetWorkflow.On("getContext").Return(targetContext)
	targetWorkflow.On("getMutableState").Return(targetMutableState)
	targetWorkflow.On("getReleaseFn").Return(targetReleaseFn)

	currentWorkflow := &mockNDCWorkflow{}
	defer currentWorkflow.AssertExpectations(s.T())
	currentContext := &mockWorkflowExecutionContext{}
	defer currentContext.AssertExpectations(s.T())
	currentMutableState := &mockMutableState{}
	defer currentMutableState.AssertExpectations(s.T())
	var currentReleaseFn releaseWorkflowExecutionFunc = func(error) { currentReleaseCalled = true }
	currentWorkflow.On("getReleaseFn").Return(currentReleaseFn)

	targetWorkflowSnapshot := &persistence.WorkflowSnapshot{}
	targetWorkflowEventsSeq := []*persistence.WorkflowEvents{&persistence.WorkflowEvents{}}
	targetWorkflowHistorySize := int64(12345)
	targetMutableState.On("GetExecutionInfo").Return(&persistence.WorkflowExecutionInfo{
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      targetRunID,
	})
	targetMutableState.On("CloseTransactionAsSnapshot", now, transactionPolicyPassive).Return(
		targetWorkflowSnapshot, targetWorkflowEventsSeq, nil,
	).Once()

	s.mockTransactionMgr.On("getCurrentWorkflowRunID", ctx, domainID, workflowID).Return(currentRunID, nil).Once()
	s.mockTransactionMgr.On("loadNDCWorkflow", ctx, domainID, workflowID, currentRunID).Return(currentWorkflow, nil).Once()

	targetWorkflow.On("happensAfter", currentWorkflow).Return(false, nil)
	targetWorkflow.On("suppressWorkflowBy", currentWorkflow).Return(nil).Once()

	targetContext.On(
		"persistFirstWorkflowEvents", targetWorkflowEventsSeq[0],
	).Return(targetWorkflowHistorySize, nil).Once()
	targetContext.On(
		"createWorkflowExecution",
		targetWorkflowSnapshot,
		targetWorkflowHistorySize,
		now,
		persistence.CreateWorkflowModeZombie,
		"",
		int64(0),
	).Return(nil).Once()

	err := s.createMgr.dispatchForNewWorkflow(ctx, now, targetWorkflow)
	s.NoError(err)
	s.True(targetReleaseCalled)
	s.True(currentReleaseCalled)
}

func (s *nDCTransactionMgrForNewWorkflowSuite) TestDispatchForNewWorkflow_SuppressCurrentAndCreateAsCurrent() {
	ctx := ctx.Background()
	now := time.Now()

	domainID := "some random domain ID"
	workflowID := "some random workflow ID"
	targetRunID := "some random run ID"
	currentRunID := "other random runID"

	targetReleaseCalled := false
	currentReleaseCalled := false

	targetWorkflow := &mockNDCWorkflow{}
	defer targetWorkflow.AssertExpectations(s.T())
	targetContext := &mockWorkflowExecutionContext{}
	defer targetContext.AssertExpectations(s.T())
	targetMutableState := &mockMutableState{}
	defer targetMutableState.AssertExpectations(s.T())
	var targetReleaseFn releaseWorkflowExecutionFunc = func(error) { targetReleaseCalled = true }
	targetWorkflow.On("getContext").Return(targetContext)
	targetWorkflow.On("getMutableState").Return(targetMutableState)
	targetWorkflow.On("getReleaseFn").Return(targetReleaseFn)

	currentWorkflow := &mockNDCWorkflow{}
	defer currentWorkflow.AssertExpectations(s.T())
	currentContext := &mockWorkflowExecutionContext{}
	defer currentContext.AssertExpectations(s.T())
	currentMutableState := &mockMutableState{}
	defer currentMutableState.AssertExpectations(s.T())
	var currentReleaseFn releaseWorkflowExecutionFunc = func(error) { currentReleaseCalled = true }
	currentWorkflow.On("getContext").Return(currentContext)
	currentWorkflow.On("getMutableState").Return(currentMutableState)
	currentWorkflow.On("getReleaseFn").Return(currentReleaseFn)

	targetMutableState.On("GetExecutionInfo").Return(&persistence.WorkflowExecutionInfo{
		DomainID:   domainID,
		WorkflowID: workflowID,
		RunID:      targetRunID,
	})

	s.mockTransactionMgr.On("getCurrentWorkflowRunID", ctx, domainID, workflowID).Return(currentRunID, nil).Once()
	s.mockTransactionMgr.On("loadNDCWorkflow", ctx, domainID, workflowID, currentRunID).Return(currentWorkflow, nil).Once()

	targetWorkflow.On("happensAfter", currentWorkflow).Return(true, nil)
	currentMutableState.On("IsWorkflowExecutionRunning").Return(true)
	currentWorkflow.On("suppressWorkflowBy", targetWorkflow).Return(nil).Once()

	currentContext.On(
		"updateWorkflowExecutionWithNew",
		now,
		persistence.UpdateWorkflowModeUpdateCurrent,
		targetContext,
		targetMutableState,
		transactionPolicyActive,
		transactionPolicyPassive.ptr(),
	).Return(nil).Once()

	err := s.createMgr.dispatchForNewWorkflow(ctx, now, targetWorkflow)
	s.NoError(err)
	s.True(targetReleaseCalled)
	s.True(currentReleaseCalled)
}
