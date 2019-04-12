// Copyright (c) 2016 Uber Technologies, Inc.
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

package host

import (
	"flag"
	"io/ioutil"
	"testing"
	"time"

	"github.com/olivere/elastic"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	workflow "github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
)

type elasticsearchIntegrationSuite struct {
	// override suite.Suite.Assertions with require.Assertions; this means that s.NotNil(nil) will stop the test,
	// not merely log an error
	*require.Assertions
	IntegrationBase
	esClient *elastic.Client
}

// This cluster use customized threshold for history config
func (s *elasticsearchIntegrationSuite) SetupSuite() {
	s.setupSuite("testdata/integration_elasticsearch_cluster.yaml")
	var err error
	s.esClient, err = elastic.NewClient(
		elastic.SetURL(s.testClusterConfig.ESConfig.URL.String()),
		elastic.SetRetrier(elastic.NewBackoffRetrier(elastic.NewExponentialBackoff(128*time.Millisecond, 513*time.Millisecond))),
	)
	s.Require().NoError(err)

	template, err := ioutil.ReadFile("testdata/es_index_template.json")
	s.Require().NoError(err)
	putTemplate, err := s.esClient.IndexPutTemplate("test-visibility-template").BodyString(string(template)).Do(createContext())
	s.Require().NoError(err)
	s.Require().True(putTemplate.Acknowledged)

	index := s.testClusterConfig.ESConfig.Indices[common.VisibilityAppName]
	exists, err := s.esClient.IndexExists(index).Do(createContext())
	s.Require().NoError(err)
	if exists {
		deleteTestIndex, err := s.esClient.DeleteIndex(index).Do(createContext())
		s.Require().Nil(err)
		s.Require().True(deleteTestIndex.Acknowledged)
	}

	createIndex, err := s.esClient.CreateIndex(index).Do(createContext())
	s.Require().NoError(err)
	s.Require().True(createIndex.Acknowledged)
}

func (s *elasticsearchIntegrationSuite) TearDownSuite() {
	s.tearDownSuite()
	deleteTestIndex, err := s.esClient.DeleteIndex(s.testClusterConfig.ESConfig.Indices[common.VisibilityAppName]).Do(createContext())
	s.Nil(err)
	s.True(deleteTestIndex.Acknowledged)

	deleteTemplate, err := s.esClient.IndexDeleteTemplate("test-visibility-template").Do(createContext())
	s.Nil(err)
	s.True(deleteTemplate.Acknowledged)
}

func (s *elasticsearchIntegrationSuite) SetupTest() {
	// Have to define our overridden assertions in the test setup. If we did it earlier, s.T() will return nil
	s.Assertions = require.New(s.T())
}

func TestElasticsearchIntegrationSuite(t *testing.T) {
	flag.Parse()
	suite.Run(t, new(elasticsearchIntegrationSuite))
}

func (s *elasticsearchIntegrationSuite) TestListOpenWorkflow() {
	id := "es-integration-start-workflow-test"
	wt := "es-integration-start-workflow-test-type"
	tl := "es-integration-start-workflow-test-tasklist"
	identity := "worker1"

	workflowType := &workflow.WorkflowType{}
	workflowType.Name = common.StringPtr(wt)

	taskList := &workflow.TaskList{}
	taskList.Name = common.StringPtr(tl)

	request := &workflow.StartWorkflowExecutionRequest{
		RequestId:                           common.StringPtr(uuid.New()),
		Domain:                              common.StringPtr(s.domainName),
		WorkflowId:                          common.StringPtr(id),
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(100),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		Identity:                            common.StringPtr(identity),
	}

	startTime := time.Now().UnixNano()
	we, err := s.engine.StartWorkflowExecution(createContext(), request)
	s.Nil(err)

	startFilter := &workflow.StartTimeFilter{}
	startFilter.EarliestTime = common.Int64Ptr(startTime)
	var openExecution *workflow.WorkflowExecutionInfo
	for i := 0; i < 20; i++ {
		startFilter.LatestTime = common.Int64Ptr(time.Now().UnixNano())
		resp, err := s.engine.ListOpenWorkflowExecutions(createContext(), &workflow.ListOpenWorkflowExecutionsRequest{
			Domain:          common.StringPtr(s.domainName),
			MaximumPageSize: common.Int32Ptr(100),
			StartTimeFilter: startFilter,
			ExecutionFilter: &workflow.WorkflowExecutionFilter{
				WorkflowId: common.StringPtr(id),
			},
		})
		s.Nil(err)
		if len(resp.GetExecutions()) == 1 {
			openExecution = resp.GetExecutions()[0]
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	s.NotNil(openExecution)
	s.Equal(we.GetRunId(), openExecution.GetExecution().GetRunId())
}
