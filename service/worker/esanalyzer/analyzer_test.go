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
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/client"
	"github.com/uber/cadence/client/admin"
	"github.com/uber/cadence/common/cache"
	"github.com/uber/cadence/common/cluster"
	"github.com/uber/cadence/common/dynamicconfig"
	"github.com/uber/cadence/common/elasticsearch"
	esMocks "github.com/uber/cadence/common/elasticsearch/mocks"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/metrics"
	"github.com/uber/cadence/common/metrics/mocks"
	"github.com/uber/cadence/common/persistence"
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
	tallyScope         tally.Scope
	analyzer           *Analyzer
	workflow           *Workflow
	config             Config
	DomainID           string
	DomainName         string
	WorkflowType       string
	WorkflowID         string
	RunID              string
	WorkflowVersion    string
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
	s.WorkflowVersion = "test-workflow-version"
	s.tallyScope = tally.NoopScope
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
	s.resource = resource.NewTest(s.T(), s.controller, metrics.Worker)
	s.mockAdminClient = admin.NewMockClient(s.controller)
	s.clientBean = client.NewMockBean(s.controller)
	s.logger = &log.MockLogger{}
	s.mockMetricClient = &mocks.Client{}
	s.scopedMetricClient = &mocks.Scope{}
	s.mockESClient = &esMocks.GenericClient{}

	//
	// s.mockDomainCache.EXPECT().GetDomainByID(s.DomainID).Return(activeDomainCache, nil).AnyTimes()
	s.mockDomainCache.EXPECT().GetDomain(s.DomainName).Return(activeDomainCache, nil).AnyTimes()

	// SET UP ANALYZER
	s.analyzer = &Analyzer{
		svcClient:   s.resource.GetSDKClient(),
		clientBean:  s.clientBean,
		domainCache: s.mockDomainCache,
		logger:      s.logger,
		esClient:    s.mockESClient,
		config:      &s.config,
		tallyScope:  s.tallyScope,
	}
	s.activityEnv.SetTestTimeout(time.Second * 5)
	s.activityEnv.SetWorkerOptions(worker.Options{BackgroundActivityContext: context.Background()})

	// REGISTER WORKFLOWS AND ACTIVITIES
	s.workflow = &Workflow{analyzer: s.analyzer}
	s.workflowEnv.RegisterWorkflowWithOptions(
		s.workflow.workflowFunc,
		workflow.RegisterOptions{Name: esanalyzerWFTypeName})
	s.workflowEnv.RegisterActivityWithOptions(
		s.workflow.emitWorkflowVersionMetrics,
		activity.RegisterOptions{Name: emitWorkflowVersionMetricsActivity},
	)
	s.activityEnv.RegisterActivityWithOptions(
		s.workflow.emitWorkflowVersionMetrics,
		activity.RegisterOptions{Name: emitWorkflowVersionMetricsActivity},
	)

	s.workflowEnv.RegisterWorkflowWithOptions(
		s.workflow.emitWorkflowTypeCount,
		workflow.RegisterOptions{Name: domainWFTypeCountWorkflowTypeName})
	s.activityEnv.RegisterActivityWithOptions(
		s.workflow.emitWorkflowTypeCountMetrics,
		activity.RegisterOptions{Name: emitDomainWorkflowTypeCountMetricsActivity},
	)
}

func (s *esanalyzerWorkflowTestSuite) TearDownTest() {
	defer s.controller.Finish()
	defer s.resource.Finish(s.T())

	s.workflowEnv.AssertExpectations(s.T())
}

func (s *esanalyzerWorkflowTestSuite) TestExecuteWorkflow() {
	s.workflowEnv.OnActivity(emitWorkflowVersionMetricsActivity, mock.Anything).Return(nil).Times(1)

	s.workflowEnv.ExecuteWorkflow(esanalyzerWFTypeName, mock.Anything)
	err := s.workflowEnv.GetWorkflowResult(nil)
	s.NoError(err)
}

func (s *esanalyzerWorkflowTestSuite) TestEmitWorkflowVersionMetricsActivity() {
	s.config.ESAnalyzerWorkflowVersionDomains = dynamicconfig.GetStringPropertyFn(
		fmt.Sprintf(`["%s"]`, s.DomainName),
	)
	esRaw := `
{
  "took": 642,
  "timed_out": false,
  "_shards": {
    "total": 20,
    "successful": 20,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": 4588,
    "max_score": 0,
    "hits": [

    ]
  },
  "aggregations": {
    "wftypes": {
      "doc_count_error_upper_bound": 0,
      "sum_other_doc_count": 0,
      "buckets": [
        {
          "key": "CourierUpdateWorkflow",
          "doc_count": 4539,
          "versions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [

            ]
          }
        },
        {
          "key": "UpdateTipWorkflow",
          "doc_count": 33,
          "versions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [
              {
                "key": "update tip efp flow-1",
                "doc_count": 3
              }
            ]
          }
        },
        {
          "key": "RoboCourierWorkflow",
          "doc_count": 12,
          "versions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [

            ]
          }
        },
        {
          "key": "ImproveDropoffPinWorkflow",
          "doc_count": 2,
          "versions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [

            ]
          }
        },
        {
          "key": "OrderStateUpdatedWorkflow",
          "doc_count": 2,
          "versions": {
            "doc_count_error_upper_bound": 0,
            "sum_other_doc_count": 0,
            "buckets": [

            ]
          }
        }
      ]
    }
  }
}
	`
	var rawEs elasticsearch.RawResponse
	err := json.Unmarshal([]byte(esRaw), &rawEs)
	s.NoError(err)
	s.mockESClient.On("SearchRaw", mock.Anything, mock.Anything, mock.Anything).Return(
		&rawEs, nil).Times(1)
	_, err = s.activityEnv.ExecuteActivity(s.workflow.emitWorkflowVersionMetrics)
	s.NoError(err)

}

func (s *esanalyzerWorkflowTestSuite) TestEmitWorkflowTypeCountMetricsActivity() {
	s.config.ESAnalyzerWorkflowTypeDomains = dynamicconfig.GetStringPropertyFn(
		fmt.Sprintf(`["%s"]`, s.DomainName),
	)
	esRaw := `
{
  "took": 642,
  "timed_out": false,
  "_shards": {
    "total": 20,
    "successful": 20,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": 4588,
    "max_score": 0,
    "hits": [

    ]
  },
  "aggregations": {
    "wftypes": {
      "doc_count_error_upper_bound": 0,
      "sum_other_doc_count": 0,
      "buckets": [
        {
          "key": "CourierUpdateWorkflow",
          "doc_count": 4539
        },
        {
          "key": "UpdateTipWorkflow",
          "doc_count": 33
        },
        {
          "key": "RoboCourierWorkflow",
          "doc_count": 12
        },
        {
          "key": "ImproveDropoffPinWorkflow",
          "doc_count": 2
        },
        {
          "key": "OrderStateUpdatedWorkflow",
          "doc_count": 2
        }
      ]
    }
  }
}
	`
	var rawEs elasticsearch.RawResponse
	err := json.Unmarshal([]byte(esRaw), &rawEs)
	s.NoError(err)
	s.mockESClient.On("SearchRaw", mock.Anything, mock.Anything, mock.Anything).Return(
		&rawEs, nil).Times(1)
	_, err = s.activityEnv.ExecuteActivity(s.workflow.emitWorkflowTypeCountMetrics)
	s.NoError(err)

}
