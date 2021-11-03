// Copyright (c) 2017-2021 Uber Technologies, Inc.
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

package shadower

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/.gen/go/shadower"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/persistence"
)

const (
	testActiveDomainName  = "active-domain"
	testStandbyDomainName = "standby-domain"

	testTaskListName  = "test-tl"
	testWorkflowQuery = "some random workflow query"
)

type workflowSuite struct {
	*require.Assertions
	suite.Suite
	testsuite.WorkflowTestSuite

	controller      *gomock.Controller
	mockDomainCache *cache.MockDomainCache

	env *testsuite.TestWorkflowEnvironment
}

func TestWorkflowSuite(t *testing.T) {
	s := new(workflowSuite)
	suite.Run(t, s)
}

func (s *workflowSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockDomainCache = cache.NewMockDomainCache(s.controller)

	activeDomainCache := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{ID: "random domainID", Name: testActiveDomainName},
		&persistence.DomainConfig{Retention: 1},
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestCurrentClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
				{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		1234,
		cluster.GetTestClusterMetadata(true, true),
	)
	s.mockDomainCache.EXPECT().GetDomain(testActiveDomainName).Return(activeDomainCache, nil).AnyTimes()

	standbyDomainCache := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{ID: "random domainID", Name: testStandbyDomainName},
		&persistence.DomainConfig{Retention: 1},
		&persistence.DomainReplicationConfig{
			ActiveClusterName: cluster.TestAlternativeClusterName,
			Clusters: []*persistence.ClusterReplicationConfig{
				{ClusterName: cluster.TestCurrentClusterName},
				{ClusterName: cluster.TestAlternativeClusterName},
			},
		},
		1234,
		cluster.GetTestClusterMetadata(true, true),
	)
	s.mockDomainCache.EXPECT().GetDomain(testStandbyDomainName).Return(standbyDomainCache, nil).AnyTimes()

	activityContext := context.Background()
	activityContext = context.WithValue(activityContext, workerContextKey, &Worker{
		domainCache: s.mockDomainCache,
	})

	s.env = s.NewTestWorkflowEnvironment()
	s.env.SetWorkerOptions(worker.Options{
		BackgroundActivityContext: activityContext,
	})
	s.env.RegisterWorkflowWithOptions(
		shadowWorkflow,
		workflow.RegisterOptions{Name: shadower.WorkflowName},
	)
	s.env.RegisterActivityWithOptions(
		testScanWorkflowActivity,
		activity.RegisterOptions{Name: shadower.ScanWorkflowActivityName},
	)
	s.env.RegisterActivityWithOptions(
		testReplayWorkflowActivity,
		activity.RegisterOptions{Name: shadower.ReplayWorkflowActivityName},
	)
}

func (s *workflowSuite) TearDownTest() {
	s.controller.Finish()
	s.env.AssertExpectations(s.T())
}

func (s *workflowSuite) TestShadowWorkflow_DomainNotSpecified() {
	s.env.ExecuteWorkflow(shadowWorkflow, shadower.WorkflowParams{
		TaskList: common.StringPtr(testTaskListName),
	})

	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
}

func (s *workflowSuite) TestShadowWorkflow_TaskListNotSpecified() {
	s.env.ExecuteWorkflow(shadowWorkflow, shadower.WorkflowParams{
		Domain: common.StringPtr(testStandbyDomainName),
	})

	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
}

func (s *workflowSuite) TestShadowWorkflow_StandbyDomain() {
	s.env.ExecuteWorkflow(shadowWorkflow, shadower.WorkflowParams{
		Domain:   common.StringPtr(testStandbyDomainName),
		TaskList: common.StringPtr(testTaskListName),
	})

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	var result shadower.WorkflowResult
	err := s.env.GetWorkflowResult(&result)
	s.NoError(err)
}

func (s *workflowSuite) TestShadowWorkflow_ScanWorkflowNonRetryableError() {
	s.env.OnActivity(shadower.ScanWorkflowActivityName, mock.Anything, mock.Anything).Return(
		shadower.ScanWorkflowActivityResult{},
		cadence.NewCustomError(shadower.ErrReasonInvalidQuery, "invalid query"),
	).Once()
	s.env.ExecuteWorkflow(shadowWorkflow, shadower.WorkflowParams{
		Domain:        common.StringPtr(testActiveDomainName),
		TaskList:      common.StringPtr(testTaskListName),
		WorkflowQuery: common.StringPtr("invalid workflow query"),
	})

	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
}

func (s *workflowSuite) TestShadowWorkflow_ReplayWorkflowNonRetryableError() {
	s.env.OnActivity(shadower.ScanWorkflowActivityName, mock.Anything, mock.Anything).Return(
		shadower.ScanWorkflowActivityResult{
			Executions: make([]*shared.WorkflowExecution, 10),
		},
		nil,
	).Once()
	s.env.OnActivity(shadower.ReplayWorkflowActivityName, mock.Anything, mock.Anything).Return(
		shadower.ReplayWorkflowActivityResult{Succeeded: common.Int32Ptr(0)},
		nil,
	).Once()
	s.env.OnActivity(shadower.ReplayWorkflowActivityName, mock.Anything, mock.Anything).Return(
		shadower.ReplayWorkflowActivityResult{},
		cadence.NewCustomError(shadower.ErrReasonWorkflowTypeNotRegistered, "workflow not registered"),
	).Once()
	s.env.ExecuteWorkflow(shadowWorkflow, shadower.WorkflowParams{
		Domain:        common.StringPtr(testActiveDomainName),
		TaskList:      common.StringPtr(testTaskListName),
		WorkflowQuery: common.StringPtr(testWorkflowQuery),
		Concurrency:   common.Int32Ptr(3),
	})

	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
}

func (s *workflowSuite) TestShadowWorkflow_ExitCondition_ShadowCount_NoLastResult() {
	shadowCount := 10
	s.env.OnActivity(shadower.ScanWorkflowActivityName, mock.Anything, mock.Anything).Return(
		shadower.ScanWorkflowActivityResult{
			Executions:    make([]*shared.WorkflowExecution, shadowCount/2+1),
			NextPageToken: []byte{1, 2, 3},
		},
		nil,
	).Times(2)
	s.env.OnActivity(shadower.ReplayWorkflowActivityName, mock.Anything, mock.Anything).Return(
		shadower.ReplayWorkflowActivityResult{Succeeded: common.Int32Ptr(int32(shadowCount)/2 + 1)},
		nil,
	).Times(2)

	s.env.ExecuteWorkflow(shadowWorkflow, shadower.WorkflowParams{
		Domain:        common.StringPtr(testActiveDomainName),
		TaskList:      common.StringPtr(testTaskListName),
		WorkflowQuery: common.StringPtr(testWorkflowQuery),
		ExitCondition: &shadower.ExitCondition{
			ShadowCount: common.Int32Ptr(int32(shadowCount)),
		},
	})

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	var result shadower.WorkflowResult
	err := s.env.GetWorkflowResult(&result)
	s.NoError(err)
	s.True(result.GetSucceeded() >= int32(shadowCount))
}

func (s *workflowSuite) TestShadowWorkflow_ExitCondition_ShadowCount_WithLastResult() {
	shadowCount := 10
	s.env.OnActivity(shadower.ScanWorkflowActivityName, mock.Anything, mock.Anything).Return(
		shadower.ScanWorkflowActivityResult{
			Executions:    make([]*shared.WorkflowExecution, shadowCount/2+1),
			NextPageToken: []byte{1, 2, 3},
		},
		nil,
	).Once()
	s.env.OnActivity(shadower.ReplayWorkflowActivityName, mock.Anything, mock.Anything).Return(
		shadower.ReplayWorkflowActivityResult{Succeeded: common.Int32Ptr(int32(shadowCount)/2 + 1)},
		nil,
	).Once()

	lastFailed := shadowCount / 2
	s.env.ExecuteWorkflow(shadowWorkflow, shadower.WorkflowParams{
		Domain:        common.StringPtr(testActiveDomainName),
		TaskList:      common.StringPtr(testTaskListName),
		WorkflowQuery: common.StringPtr(testWorkflowQuery),
		ExitCondition: &shadower.ExitCondition{
			ShadowCount: common.Int32Ptr(int32(shadowCount - lastFailed)),
		},
		LastRunResult: &shadower.WorkflowResult{
			Failed: common.Int32Ptr(int32(lastFailed)),
		},
	})

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	var result shadower.WorkflowResult
	err := s.env.GetWorkflowResult(&result)
	s.NoError(err)
	s.True(result.GetSucceeded()+result.GetFailed() >= int32(shadowCount))
}

func (s *workflowSuite) TestShadowWorkflow_ExitCondition_ExpirationInterval() {
	expirationInterval := time.Minute
	numExecutions := 10
	now := time.Now()
	s.env.OnActivity(shadower.ScanWorkflowActivityName, mock.Anything, mock.Anything).Return(
		shadower.ScanWorkflowActivityResult{
			Executions:    make([]*shared.WorkflowExecution, numExecutions),
			NextPageToken: []byte{1, 2, 3},
		},
		nil,
	).Once()
	s.env.OnActivity(shadower.ReplayWorkflowActivityName, mock.Anything, mock.Anything).Return(
		func(_ context.Context, params shadower.ReplayWorkflowActivityParams) (shadower.ReplayWorkflowActivityResult, error) {
			s.env.SetStartTime(now.Add(2 * expirationInterval))
			return shadower.ReplayWorkflowActivityResult{Succeeded: common.Int32Ptr(int32(len(params.Executions)))}, nil
		},
	).Once()

	s.env.SetStartTime(now)
	s.env.ExecuteWorkflow(shadowWorkflow, shadower.WorkflowParams{
		Domain:        common.StringPtr(testActiveDomainName),
		TaskList:      common.StringPtr(testTaskListName),
		WorkflowQuery: common.StringPtr(testWorkflowQuery),
		ExitCondition: &shadower.ExitCondition{
			ExpirationIntervalInSeconds: common.Int32Ptr(int32(expirationInterval.Seconds())),
		},
	})

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	var result shadower.WorkflowResult
	err := s.env.GetWorkflowResult(&result)
	s.NoError(err)
	s.Equal(int32(numExecutions), result.GetSucceeded())
}

func (s *workflowSuite) TestShadowWorkflow_ContinueAsNew_MaxShadowCount() {
	pages := 100
	pageSize := defaultMaxShadowCountPerRun / pages
	startTime := time.Now()
	replayTimePerPage := time.Second
	s.env.OnActivity(shadower.ScanWorkflowActivityName, mock.Anything, mock.Anything).Return(
		shadower.ScanWorkflowActivityResult{
			Executions:    make([]*shared.WorkflowExecution, pageSize),
			NextPageToken: []byte{1, 2, 3},
		},
		nil,
	).Times(pages)
	s.env.OnActivity(shadower.ReplayWorkflowActivityName, mock.Anything, mock.Anything).Return(
		func(_ context.Context, params shadower.ReplayWorkflowActivityParams) (shadower.ReplayWorkflowActivityResult, error) {
			startTime = startTime.Add(replayTimePerPage)
			s.env.SetStartTime(startTime)
			return shadower.ReplayWorkflowActivityResult{Succeeded: common.Int32Ptr(int32(len(params.Executions)))}, nil
		},
	).Times(pages)

	s.env.SetStartTime(startTime)
	exitCondition := &shadower.ExitCondition{
		ShadowCount:                 common.Int32Ptr(defaultMaxShadowCountPerRun * 10),
		ExpirationIntervalInSeconds: common.Int32Ptr(1 * int32(pages) * 10),
	}
	timerFired := 0
	s.env.SetOnTimerFiredListener(func(_ string) {
		timerFired++
	})
	s.env.ExecuteWorkflow(shadowWorkflow, shadower.WorkflowParams{
		Domain:        common.StringPtr(testActiveDomainName),
		TaskList:      common.StringPtr(testTaskListName),
		WorkflowQuery: common.StringPtr(testWorkflowQuery),
		ExitCondition: exitCondition,
	})

	s.True(s.env.IsWorkflowCompleted())
	fmt.Print("error", s.env.GetWorkflowError())
	continueAsNewErr, ok := s.env.GetWorkflowError().(*workflow.ContinueAsNewError)
	s.True(ok)
	s.Equal(shadower.WorkflowName, continueAsNewErr.WorkflowType().Name)
	shadowParams, ok := continueAsNewErr.Args()[0].(shadower.WorkflowParams)
	s.Equal(testActiveDomainName, shadowParams.GetDomain())
	s.Equal(testTaskListName, shadowParams.GetTaskList())
	s.Equal(testWorkflowQuery, shadowParams.GetWorkflowQuery())
	s.NotNil(shadowParams.NextPageToken)
	s.Equal(1.0, shadowParams.GetSamplingRate())
	s.Equal(shadower.ModeNormal, shadowParams.GetShadowMode())
	shadowedWorkflows := shadowParams.GetLastRunResult().GetSucceeded() + shadowParams.GetLastRunResult().GetFailed()
	s.Equal(exitCondition.GetShadowCount()-shadowedWorkflows, shadowParams.ExitCondition.GetShadowCount())
	s.Equal(int32(pages)*int32(replayTimePerPage.Seconds()), exitCondition.GetExpirationIntervalInSeconds()-shadowParams.GetExitCondition().GetExpirationIntervalInSeconds())
	s.Equal(int32(defaultReplayConcurrency), shadowParams.GetConcurrency())
	totalWorkflows := shadowedWorkflows + shadowParams.GetLastRunResult().GetSkipped()
	s.GreaterOrEqual(totalWorkflows, int32(defaultMaxShadowCountPerRun))
	s.Equal(0, timerFired)
}

func (s *workflowSuite) TestShadowWorkflow_ContinueAsNew_ContinuousShadowing() {
	pageSize := 10
	pages := 5
	samplingRate := 0.5
	s.env.OnActivity(shadower.ScanWorkflowActivityName, mock.Anything, mock.Anything).Return(
		shadower.ScanWorkflowActivityResult{
			Executions:    make([]*shared.WorkflowExecution, pageSize),
			NextPageToken: []byte{1, 2, 3},
		},
		nil,
	).Times(pages - 1)
	s.env.OnActivity(shadower.ScanWorkflowActivityName, mock.Anything, mock.Anything).Return(
		shadower.ScanWorkflowActivityResult{
			Executions: make([]*shared.WorkflowExecution, pageSize),
		},
		nil,
	).Times(1)
	s.env.OnActivity(shadower.ReplayWorkflowActivityName, mock.Anything, mock.Anything).Return(
		shadower.ReplayWorkflowActivityResult{Succeeded: common.Int32Ptr(int32(pageSize))},
		nil,
	).Times(pages)

	timerFired := 0
	s.env.SetOnTimerFiredListener(func(_ string) {
		timerFired++
	})

	s.env.ExecuteWorkflow(shadowWorkflow, shadower.WorkflowParams{
		Domain:        common.StringPtr(testActiveDomainName),
		TaskList:      common.StringPtr(testTaskListName),
		WorkflowQuery: common.StringPtr("some random workflow query"),
		ShadowMode:    shadower.ModeContinuous.Ptr(),
		SamplingRate:  common.Float64Ptr(samplingRate),
	})

	s.True(s.env.IsWorkflowCompleted())
	fmt.Print("error", s.env.GetWorkflowError())
	continueAsNewErr, ok := s.env.GetWorkflowError().(*workflow.ContinueAsNewError)
	s.True(ok)
	s.Equal(shadower.WorkflowName, continueAsNewErr.WorkflowType().Name)
	shadowParams, ok := continueAsNewErr.Args()[0].(shadower.WorkflowParams)
	s.Equal(testActiveDomainName, shadowParams.GetDomain())
	s.Equal(testTaskListName, shadowParams.GetTaskList())
	s.Equal(testWorkflowQuery, shadowParams.GetWorkflowQuery())
	s.Nil(shadowParams.NextPageToken)
	s.Equal(samplingRate, shadowParams.GetSamplingRate())
	s.Equal(shadower.ModeContinuous, shadowParams.GetShadowMode())
	s.Empty(shadowParams.ExitCondition)
	s.Equal(int32(defaultReplayConcurrency), shadowParams.GetConcurrency())
	s.Equal(int32(pages*pageSize), shadowParams.LastRunResult.GetSucceeded())
	s.Equal(1, timerFired)
}

// dummy activity implementations for test
// so that we can register and mock the implementation/result

func testScanWorkflowActivity(
	ctx context.Context,
	params shadower.ScanWorkflowActivityParams,
) (shadower.ScanWorkflowActivityResult, error) {
	return shadower.ScanWorkflowActivityResult{}, nil
}

func testReplayWorkflowActivity(
	ctx context.Context,
	params shadower.ReplayWorkflowActivityParams,
) (shadower.ReplayWorkflowActivityResult, error) {
	return shadower.ReplayWorkflowActivityResult{}, nil
}
