// Copyright (c) 2018 Uber Technologies, Inc.
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

//go:build pinotintegration
// +build pinotintegration

// to run locally, make sure kafka and pinot is running,
// then run cmd `go test -v ./host -run TestPinotIntegrationSuite -tags pinotintegration`

package host

import (
	"flag"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common/definition"
	pnt "github.com/uber/cadence/common/pinot"

	"testing"

	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/log/loggerimpl"

	"go.uber.org/zap"

	"github.com/uber/cadence/host/pinotutils"

	"encoding/json"
	"fmt"

	"github.com/pborman/uuid"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/types"
)

const (
	numOfRetry        = 50
	waitTimeInMs      = 400
	waitForESToSettle = 4 * time.Second // wait es shards for some time ensure data consistent
)

type PinotIntegrationSuite struct {
	*require.Assertions
	logger log.Logger
	IntegrationBase
	pinotClient pnt.GenericClient

	testSearchAttributeKey string
	testSearchAttributeVal string
}

func TestPinotIntegrationSuite(t *testing.T) {
	flag.Parse()
	clusterConfig, err := GetTestClusterConfig("testdata/integration_pinot_cluster.yaml")
	if err != nil {
		panic(err)
	}
	testCluster := NewPersistenceTestCluster(t, clusterConfig)

	s := new(PinotIntegrationSuite)
	params := IntegrationBaseParams{
		DefaultTestCluster:    testCluster,
		VisibilityTestCluster: testCluster,
		TestClusterConfig:     clusterConfig,
	}
	s.IntegrationBase = NewIntegrationBase(params)
	suite.Run(t, s)
}

func (s *PinotIntegrationSuite) SetupSuite() {
	s.setupSuiteForPinotTest()
	zapLogger, err := zap.NewDevelopment()
	s.Require().NoError(err)
	s.logger = loggerimpl.NewLogger(zapLogger)
	tableName := "cadence_visibility_pinot" //cadence_visibility_pinot_integration_test
	url := "localhost:8099"
	s.pinotClient = pinotutils.CreatePinotClient(s.Suite, url, tableName, s.logger)
}

func (s *PinotIntegrationSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.testSearchAttributeKey = definition.CustomStringField
	s.testSearchAttributeVal = "test value"
}

func (s *PinotIntegrationSuite) TearDownSuite() {
	s.tearDownSuite()
	// check how to clean up test_table
}

func (s *PinotIntegrationSuite) TestListWorkflow() {
	id := "es-integration-list-workflow-test"
	wt := "es-integration-list-workflow-test-type"
	tl := "es-integration-list-workflow-test-tasklist"
	request := s.createStartWorkflowExecutionRequest(id, wt, tl)

	we, err := s.engine.StartWorkflowExecution(createContext(), request)
	s.Nil(err)

	query := fmt.Sprintf(`WorkflowID = "%s"`, id)
	s.testHelperForReadOnce(we.GetRunID(), query, false, false)
}

func (s *PinotIntegrationSuite) createStartWorkflowExecutionRequest(id, wt, tl string) *types.StartWorkflowExecutionRequest {
	identity := "worker1"
	workflowType := &types.WorkflowType{}
	workflowType.Name = wt

	taskList := &types.TaskList{}
	taskList.Name = tl

	request := &types.StartWorkflowExecutionRequest{
		RequestID:                           uuid.New(),
		Domain:                              s.domainName,
		WorkflowID:                          id,
		WorkflowType:                        workflowType,
		TaskList:                            taskList,
		Input:                               nil,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(100),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		Identity:                            identity,
	}
	return request
}

func (s *PinotIntegrationSuite) testHelperForReadOnce(runID, query string, isScan bool, isAnyMatchOk bool) {
	s.testHelperForReadOnceWithDomain(s.domainName, runID, query, isScan, isAnyMatchOk)
}

func (s *PinotIntegrationSuite) testHelperForReadOnceWithDomain(domainName string, runID, query string, isScan bool, isAnyMatchOk bool) {
	var openExecution *types.WorkflowExecutionInfo
	listRequest := &types.ListWorkflowExecutionsRequest{
		Domain:   domainName,
		PageSize: defaultTestValueOfESIndexMaxResultWindow,
		Query:    query,
	}
Retry:
	for i := 0; i < numOfRetry; i++ {
		var resp *types.ListWorkflowExecutionsResponse
		var err error

		if isScan {
			resp, err = s.engine.ScanWorkflowExecutions(createContext(), listRequest)
		} else {
			resp, err = s.engine.ListWorkflowExecutions(createContext(), listRequest)
		}

		s.Nil(err)
		logStr := fmt.Sprintf("Results for query '%s' (desired runId: %s): \n", query, runID)
		s.Logger.Info(logStr)
		for _, e := range resp.GetExecutions() {
			logStr = fmt.Sprintf("Execution: %+v, %+v \n", e.Execution, e)
			s.Logger.Info(logStr)
		}
		if len(resp.GetExecutions()) == 1 {
			openExecution = resp.GetExecutions()[0]
			break
		}
		if isAnyMatchOk {
			for _, e := range resp.GetExecutions() {
				if e.Execution.RunID == runID {
					openExecution = e
					break Retry
				}
			}
		}
		time.Sleep(waitTimeInMs * time.Millisecond)
	}
	s.NotNil(openExecution)
	s.Equal(runID, openExecution.GetExecution().GetRunID())
	s.True(openExecution.GetExecutionTime() >= openExecution.GetStartTime())
	if openExecution.SearchAttributes != nil && len(openExecution.SearchAttributes.GetIndexedFields()) > 0 {
		searchValBytes := openExecution.SearchAttributes.GetIndexedFields()[s.testSearchAttributeKey]
		var searchVal string
		json.Unmarshal(searchValBytes, &searchVal)
		s.Equal(s.testSearchAttributeVal, searchVal)
	}
}
