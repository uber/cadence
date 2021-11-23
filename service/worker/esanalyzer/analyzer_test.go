// Copyright (c) 2017-2021 Uber Technologies Inc.
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
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABxILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package esanalyzer

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/yarpc"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/elasticsearch"
	esMocks "github.com/uber/cadence/common/elasticsearch/mocks"

	"github.com/uber/cadence/client"
	"github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/metrics/mocks"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/types"
	"github.com/uber/cadence/service/history/resource"
)

type esanalyzerWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	activityEnv        *testsuite.TestActivityEnvironment
	workflowEnv        *testsuite.TestWorkflowEnvironment
	controller         *gomock.Controller
	resource           *resource.Test
	mockAdminClient    *admin.MockClient
	mockDomainCache    *cache.MockDomainCache
	clientBean         *client.MockBean
	logger             *log.MockLogger
	mockMetricClient   *mocks.Client
	scopedMetricClient *mocks.Scope
	mockESClient       *esMocks.GenericClient
	analyzer           *Analyzer
	workflow           *Workflow
	config             Config
	DomainID           string
	DomainName         string
	WorkflowType       string
	WorkflowID         string
	RunID              string
}

func TestESAnalyzerWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(esanalyzerWorkflowTestSuite))
}

func (s *esanalyzerWorkflowTestSuite) SetupTest() {
	s.DomainID = "deadbeef-0123-4567-890a-bcdef0123460"
	s.DomainName = "test-domain"
	s.WorkflowType = "test-workflow-type"
	s.WorkflowID = "test-workflow_id"
	s.RunID = "test-run_id"

	activeDomainCache := cache.NewGlobalDomainCacheEntryForTest(
		&persistence.DomainInfo{ID: s.DomainID, Name: s.DomainName},
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

	s.config = Config{
		ESAnalyzerPause:                          dynamicconfig.GetBoolPropertyFn(false),
		ESAnalyzerTimeWindow:                     dynamicconfig.GetDurationPropertyFn(time.Hour * 24 * 30),
		ESAnalyzerMaxNumDomains:                  dynamicconfig.GetIntPropertyFn(500),
		ESAnalyzerMaxNumWorkflowTypes:            dynamicconfig.GetIntPropertyFn(100),
		ESAnalyzerLimitToTypes:                   dynamicconfig.GetStringPropertyFn(""),
		ESAnalyzerLimitToDomains:                 dynamicconfig.GetStringPropertyFn(""),
		ESAnalyzerNumWorkflowsToRefresh:          dynamicconfig.GetIntPropertyFilteredByWorkflowType(2),
		ESAnalyzerBufferWaitTime:                 dynamicconfig.GetDurationPropertyFilteredByWorkflowType(time.Minute * 30),
		ESAnalyzerMinNumWorkflowsForAvg:          dynamicconfig.GetIntPropertyFilteredByWorkflowType(100),
		ESAnalyzerWorkflowDurationWarnThresholds: dynamicconfig.GetStringPropertyFn(""),
	}

	s.activityEnv = s.NewTestActivityEnvironment()
	s.workflowEnv = s.NewTestWorkflowEnvironment()
	s.controller = gomock.NewController(s.T())
	s.mockDomainCache = cache.NewMockDomainCache(s.controller)
	s.resource = resource.NewTest(s.controller, metrics.Worker)
	s.mockAdminClient = admin.NewMockClient(s.controller)
	s.clientBean = client.NewMockBean(s.controller)
	s.logger = &log.MockLogger{}
	s.mockMetricClient = &mocks.Client{}
	s.scopedMetricClient = &mocks.Scope{}
	s.mockESClient = &esMocks.GenericClient{}

	s.mockMetricClient.On("Scope", metrics.ESAnalyzerScope, mock.Anything).Return(s.scopedMetricClient).Once()
	s.scopedMetricClient.On("Tagged", mock.Anything, mock.Anything).Return(s.scopedMetricClient).Once()

	s.mockDomainCache.EXPECT().GetDomainByID(s.DomainID).Return(activeDomainCache, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomain(s.DomainName).Return(activeDomainCache, nil).AnyTimes()
	s.clientBean.EXPECT().GetRemoteAdminClient(cluster.TestCurrentClusterName).Return(s.mockAdminClient).AnyTimes()

	// SET UP ANALYZER
	s.analyzer = &Analyzer{
		svcClient:          s.resource.GetSDKClient(),
		clientBean:         s.clientBean,
		domainCache:        s.mockDomainCache,
		logger:             s.logger,
		scopedMetricClient: getScopedMetricsClient(s.mockMetricClient),
		esClient:           s.mockESClient,
		config:             &s.config,
	}
	s.activityEnv.SetTestTimeout(time.Second * 5)
	s.activityEnv.SetWorkerOptions(worker.Options{BackgroundActivityContext: context.Background()})

	// REGISTER WORKFLOWS AND ACTIVITIES
	s.workflow = &Workflow{analyzer: s.analyzer}
	s.workflowEnv.RegisterWorkflowWithOptions(
		s.workflow.workflowFunc,
		workflow.RegisterOptions{Name: esanalyzerWFTypeName})

	s.workflowEnv.RegisterActivityWithOptions(
		s.workflow.getWorkflowTypes,
		activity.RegisterOptions{Name: getWorkflowTypesActivity})
	s.workflowEnv.RegisterActivityWithOptions(
		s.workflow.findStuckWorkflows,
		activity.RegisterOptions{Name: findStuckWorkflowsActivity})
	s.workflowEnv.RegisterActivityWithOptions(
		s.workflow.refreshStuckWorkflowsFromSameWorkflowType,
		activity.RegisterOptions{Name: refreshStuckWorkflowsActivity},
	)
	s.workflowEnv.RegisterActivityWithOptions(
		s.workflow.findLongRunningWorkflows,
		activity.RegisterOptions{Name: findLongRunningWorkflowsActivity},
	)

	s.activityEnv.RegisterActivityWithOptions(
		s.workflow.getWorkflowTypes,
		activity.RegisterOptions{Name: getWorkflowTypesActivity})
	s.activityEnv.RegisterActivityWithOptions(
		s.workflow.findStuckWorkflows,
		activity.RegisterOptions{Name: findStuckWorkflowsActivity})
	s.activityEnv.RegisterActivityWithOptions(
		s.workflow.refreshStuckWorkflowsFromSameWorkflowType,
		activity.RegisterOptions{Name: refreshStuckWorkflowsActivity},
	)
	s.activityEnv.RegisterActivityWithOptions(
		s.workflow.findLongRunningWorkflows,
		activity.RegisterOptions{Name: findLongRunningWorkflowsActivity},
	)
}

func (s *esanalyzerWorkflowTestSuite) TearDownTest() {
	defer s.controller.Finish()
	defer s.resource.Finish(s.T())

	s.workflowEnv.AssertExpectations(s.T())
}

func (s *esanalyzerWorkflowTestSuite) TestExecuteWorkflow() {
	workflowTypeInfos := []WorkflowTypeInfo{
		{
			Name:         s.WorkflowType,
			NumWorkflows: 564,
			Duration:     Duration{AvgExecTimeNanoseconds: float64(123 * time.Second)},
		},
	}
	s.workflowEnv.OnActivity(getWorkflowTypesActivity, mock.Anything).
		Return(workflowTypeInfos, nil).Times(1)

	workflows := []WorkflowInfo{
		{
			DomainID:   s.DomainID,
			WorkflowID: s.WorkflowID,
			RunID:      s.RunID,
		},
	}
	s.workflowEnv.OnActivity(findStuckWorkflowsActivity, mock.Anything, workflowTypeInfos[0]).
		Return(workflows, nil).Times(1)
	s.workflowEnv.OnActivity(findLongRunningWorkflowsActivity, mock.Anything).
		Return(nil).Times(1)

	s.workflowEnv.OnActivity(refreshStuckWorkflowsActivity, mock.Anything, workflows).Return(nil).Times(1)

	s.workflowEnv.ExecuteWorkflow(esanalyzerWFTypeName)
	err := s.workflowEnv.GetWorkflowResult(nil)
	s.NoError(err)
}

func (s *esanalyzerWorkflowTestSuite) TestExecuteWorkflowMultipleWorkflowTypes() {
	workflowTypeInfos := []WorkflowTypeInfo{
		{
			Name:         s.WorkflowType,
			NumWorkflows: 564,
			Duration:     Duration{AvgExecTimeNanoseconds: float64(123 * time.Second)},
		},
		{
			Name:         "another-workflow-type",
			NumWorkflows: 778,
			Duration:     Duration{AvgExecTimeNanoseconds: float64(332 * time.Second)},
		},
	}
	s.workflowEnv.OnActivity(getWorkflowTypesActivity, mock.Anything).
		Return(workflowTypeInfos, nil).Times(1)

	workflows1 := []WorkflowInfo{
		{
			DomainID:   s.DomainID,
			WorkflowID: s.WorkflowID,
			RunID:      s.RunID,
		},
	}
	workflows2 := []WorkflowInfo{
		{
			DomainID:   s.DomainID,
			WorkflowID: s.WorkflowID,
			RunID:      s.RunID,
		},
	}
	s.workflowEnv.OnActivity(findStuckWorkflowsActivity, mock.Anything, workflowTypeInfos[0]).
		Return(workflows1, nil).Times(1)
	s.workflowEnv.OnActivity(findStuckWorkflowsActivity, mock.Anything, workflowTypeInfos[1]).
		Return(workflows2, nil).Times(1)
	s.workflowEnv.OnActivity(findLongRunningWorkflowsActivity, mock.Anything).
		Return(nil).Times(1)

	s.workflowEnv.OnActivity(refreshStuckWorkflowsActivity, mock.Anything, workflows1).Return(nil).Times(1)
	s.workflowEnv.OnActivity(refreshStuckWorkflowsActivity, mock.Anything, workflows2).Return(nil).Times(1)

	s.workflowEnv.ExecuteWorkflow(esanalyzerWFTypeName)
	err := s.workflowEnv.GetWorkflowResult(nil)
	s.NoError(err)
}

func (s *esanalyzerWorkflowTestSuite) TestRefreshStuckWorkflowsFromSameWorkflowTypeSingleWorkflow() {
	workflows := []WorkflowInfo{
		{
			DomainID:   s.DomainID,
			WorkflowID: s.WorkflowID,
			RunID:      s.RunID,
		},
	}

	s.mockAdminClient.EXPECT().RefreshWorkflowTasks(gomock.Any(), &types.RefreshWorkflowTasksRequest{
		Domain: s.DomainName,
		Execution: &types.WorkflowExecution{
			WorkflowID: s.WorkflowID,
			RunID:      s.RunID,
		},
	}).Return(nil).Times(1)
	s.logger.On("Info", "Refreshed stuck workflow", mock.Anything).Return().Once()
	s.scopedMetricClient.On("IncCounter", metrics.ESAnalyzerNumStuckWorkflowsRefreshed).Return().Once()

	_, err := s.activityEnv.ExecuteActivity(s.workflow.refreshStuckWorkflowsFromSameWorkflowType, workflows)
	s.NoError(err)
}

func (s *esanalyzerWorkflowTestSuite) TestRefreshStuckWorkflowsFromSameWorkflowTypeMultipleWorkflows() {
	anotherWorkflowID := "another-worklow-id"
	anotherRunID := "another-run-id"

	workflows := []WorkflowInfo{
		{
			DomainID:   s.DomainID,
			WorkflowID: s.WorkflowID,
			RunID:      s.RunID,
		},
		{
			DomainID:   s.DomainID,
			WorkflowID: anotherWorkflowID,
			RunID:      anotherRunID,
		},
	}

	expectedWorkflows := map[string]bool{s.WorkflowID: false, anotherWorkflowID: false}
	expectedRunIDs := map[string]bool{s.RunID: false, anotherRunID: false}
	s.mockAdminClient.EXPECT().RefreshWorkflowTasks(gomock.Any(), gomock.Any()).Return(nil).Do(func(
		ctx context.Context,
		request *types.RefreshWorkflowTasksRequest,
		option ...yarpc.CallOption,
	) {
		expectedWorkflows[request.Execution.WorkflowID] = true
		expectedRunIDs[request.Execution.RunID] = true
	}).Times(2)
	s.logger.On("Info", "Refreshed stuck workflow", mock.Anything).Return().Times(2)
	s.scopedMetricClient.On("IncCounter", metrics.ESAnalyzerNumStuckWorkflowsRefreshed).Return().Times(2)

	_, err := s.activityEnv.ExecuteActivity(s.workflow.refreshStuckWorkflowsFromSameWorkflowType, workflows)
	s.NoError(err)

	s.Equal(2, len(expectedWorkflows))
	s.True(expectedWorkflows[s.WorkflowID])
	s.True(expectedWorkflows[anotherWorkflowID])

	s.Equal(2, len(expectedRunIDs))
	s.True(expectedRunIDs[s.RunID])
	s.True(expectedRunIDs[anotherRunID])
}

func (s *esanalyzerWorkflowTestSuite) TestRefreshStuckWorkflowsFromSameWorkflowInconsistentDomain() {
	anotherDomainID := "another-domain-id"
	anotherWorkflowID := "another-worklow-id"
	anotherRunID := "another-run-id"

	workflows := []WorkflowInfo{
		{
			DomainID:   s.DomainID,
			WorkflowID: s.WorkflowID,
			RunID:      s.RunID,
		},
		{
			DomainID:   anotherDomainID,
			WorkflowID: anotherWorkflowID,
			RunID:      anotherRunID,
		},
	}

	expectedWorkflows := map[string]bool{s.WorkflowID: false, anotherWorkflowID: false}
	expectedRunIDs := map[string]bool{s.RunID: false, anotherRunID: false}
	s.mockAdminClient.EXPECT().RefreshWorkflowTasks(gomock.Any(), gomock.Any()).Return(nil).Do(func(
		ctx context.Context,
		request *types.RefreshWorkflowTasksRequest,
		option ...yarpc.CallOption,
	) {
		expectedWorkflows[request.Execution.WorkflowID] = true
		expectedRunIDs[request.Execution.RunID] = true
	}).Times(1)
	s.logger.On("Info", "Refreshed stuck workflow", mock.Anything).Return().Times(2)
	s.scopedMetricClient.On("IncCounter", metrics.ESAnalyzerNumStuckWorkflowsRefreshed).Return().Times(2)

	_, err := s.activityEnv.ExecuteActivity(s.workflow.refreshStuckWorkflowsFromSameWorkflowType, workflows)
	s.Error(err)
	s.EqualError(err, "InternalServiceError{Message: Inconsistent worklow. Expected domainID: deadbeef-0123-4567-890a-bcdef0123460, actual: another-domain-id}")
}

func (s *esanalyzerWorkflowTestSuite) TestFindLongRunningWorkflows() {
	s.mockESClient.On("SearchRaw", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		&elasticsearch.RawResponse{
			TookInMillis: 12,
			Hits: elasticsearch.SearchHits{
				TotalHits: 1234,
			},
		},
		nil).Times(1)

	s.scopedMetricClient.On(
		"AddCounter",
		metrics.ESAnalyzerNumLongRunningWorkflows,
		int64(1234),
	).Return().Times(1)

	s.config.ESAnalyzerWorkflowDurationWarnThresholds = dynamicconfig.GetStringPropertyFn(`{"test-domain/workflow1":"1m"}`)
	_, err := s.activityEnv.ExecuteActivity(s.workflow.findLongRunningWorkflows)
	s.NoError(err)
}

func (s *esanalyzerWorkflowTestSuite) TestFindStuckWorkflows() {
	info := WorkflowTypeInfo{
		DomainID:     s.DomainID,
		Name:         s.WorkflowType,
		NumWorkflows: 123113,
		Duration:     Duration{AvgExecTimeNanoseconds: float64(100 * time.Minute)},
	}

	s.mockESClient.On("SearchRaw", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		&elasticsearch.RawResponse{
			TookInMillis: 12,
			Hits: elasticsearch.SearchHits{
				TotalHits: 2,
				Hits: []*persistence.InternalVisibilityWorkflowExecutionInfo{
					{
						DomainID:   s.DomainID,
						WorkflowID: s.WorkflowID,
						RunID:      s.RunID,
					},
					{
						DomainID:   s.DomainID,
						WorkflowID: "workflow2",
						RunID:      "run2",
					},
				},
			},
		},
		nil).Times(1)
	s.scopedMetricClient.On(
		"AddCounter",
		metrics.ESAnalyzerNumStuckWorkflowsDiscovered,
		int64(2),
	).Return().Times(1)

	actFuture, err := s.activityEnv.ExecuteActivity(s.workflow.findStuckWorkflows, info)
	s.NoError(err)
	var results []WorkflowInfo
	err = actFuture.Get(&results)
	s.NoError(err)
	s.Equal(2, len(results))
	s.Equal(WorkflowInfo{DomainID: s.DomainID, WorkflowID: s.WorkflowID, RunID: s.RunID}, results[0])
	s.Equal(WorkflowInfo{DomainID: s.DomainID, WorkflowID: "workflow2", RunID: "run2"}, results[1])
}

func (s *esanalyzerWorkflowTestSuite) TestFindStuckWorkflowsNotEnoughWorkflows() {
	info := WorkflowTypeInfo{
		DomainID:     s.DomainID,
		Name:         s.WorkflowType,
		NumWorkflows: int64(s.config.ESAnalyzerMinNumWorkflowsForAvg(s.DomainID, s.WorkflowType) - 1),
		Duration:     Duration{AvgExecTimeNanoseconds: float64(100 * time.Minute)},
	}

	s.logger.On("Warn", mock.Anything, mock.Anything).Return().Times(1)

	actFuture, err := s.activityEnv.ExecuteActivity(s.workflow.findStuckWorkflows, info)
	s.NoError(err)
	var results []WorkflowInfo
	err = actFuture.Get(&results)
	s.NoError(err)
	s.Equal(0, len(results))
}

func (s *esanalyzerWorkflowTestSuite) TestFindStuckWorkflowsMinNumWorkflowValidationSkipped() {
	info := WorkflowTypeInfo{
		DomainID:     s.DomainID,
		Name:         s.WorkflowType,
		NumWorkflows: int64(s.config.ESAnalyzerMinNumWorkflowsForAvg(s.DomainID, s.WorkflowType) - 1),
		Duration:     Duration{AvgExecTimeNanoseconds: float64(100 * time.Minute)},
	}

	s.config.ESAnalyzerLimitToTypes = dynamicconfig.GetStringPropertyFn(s.WorkflowType)
	s.logger.On("Info", mock.Anything, mock.Anything).Return().Times(1)
	s.mockESClient.On("SearchRaw", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&elasticsearch.RawResponse{}, nil).Times(1)

	actFuture, err := s.activityEnv.ExecuteActivity(s.workflow.findStuckWorkflows, info)
	s.NoError(err)
	var results []WorkflowInfo
	err = actFuture.Get(&results)
	s.NoError(err)
	s.Equal(0, len(results))
}

func (s *esanalyzerWorkflowTestSuite) TestGetWorkflowTypes() {
	esResult := struct {
		Buckets []DomainInfo `json:"buckets"`
	}{
		Buckets: []DomainInfo{
			{
				DomainID: "aaa-bbb-ccc",
				WFTypeContainer: WorkflowTypeInfoContainer{
					WorkflowTypes: []WorkflowTypeInfo{
						{
							Name:         s.WorkflowType,
							NumWorkflows: 564,
							Duration:     Duration{AvgExecTimeNanoseconds: float64(123 * time.Second)},
						},
						{
							Name:         "another-workflow-type",
							NumWorkflows: 745,
							Duration:     Duration{AvgExecTimeNanoseconds: float64(987 * time.Second)},
						},
					},
				},
			},
		},
	}
	encoded, err := json.Marshal(esResult)
	s.NoError(err)

	s.mockESClient.On("SearchRaw", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		&elasticsearch.RawResponse{
			TookInMillis: 12,
			Aggregations: map[string]json.RawMessage{
				"domains": json.RawMessage(encoded),
			},
		},
		nil).Times(1)
	s.logger.On("Info", mock.Anything, mock.Anything).Return().Once()

	actFuture, err := s.activityEnv.ExecuteActivity(s.workflow.getWorkflowTypes)
	s.NoError(err)
	var results []WorkflowTypeInfo
	err = actFuture.Get(&results)
	s.NoError(err)
	s.Equal(2, len(results))
	s.Equal(normalizeDomainInfos(esResult.Buckets), results)
}

func (s *esanalyzerWorkflowTestSuite) TestGetWorkflowTypesFromConfig() {
	workflowTypes := []WorkflowTypeInfo{
		{DomainID: s.DomainID, Name: "workflow1"},
		{DomainID: s.DomainID, Name: "workflow2"},
	}

	s.config.ESAnalyzerLimitToTypes = dynamicconfig.GetStringPropertyFn(`["test-domain/workflow1","test-domain/workflow2"]`)
	s.logger.On("Info", mock.Anything, mock.Anything).Return().Once()

	actFuture, err := s.activityEnv.ExecuteActivity(s.workflow.getWorkflowTypes)
	s.NoError(err)
	var results []WorkflowTypeInfo
	err = actFuture.Get(&results)
	s.NoError(err)
	s.Equal(2, len(results))
	s.Equal(workflowTypes, results)
}
