// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

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

package tasklist

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/workflow"

	"github.com/uber/cadence/common/metrics"
	p "github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/resource"
)

type scannerWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestScannerWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(scannerWorkflowTestSuite))
}

func (s *scannerWorkflowTestSuite) TestWorkflowExecutesActivity() {
	a := &Activities{}
	env := s.NewTestWorkflowEnvironment()
	env.RegisterWorkflowWithOptions(Workflow, workflow.RegisterOptions{Name: WFTypeName})
	env.RegisterActivityWithOptions(a.TaskListScavengerActivity, activity.RegisterOptions{Name: taskListScavengerActivityName})
	env.OnActivity(taskListScavengerActivityName, mock.Anything).Times(1).Return(nil)
	env.ExecuteWorkflow(WFTypeName)
	s.True(env.IsWorkflowCompleted())
	env.AssertExpectations(s.T())

}

func (s *scannerWorkflowTestSuite) TestScavengerActivity() {
	controller := gomock.NewController(s.T())
	defer controller.Finish()
	mockResource := resource.NewTest(controller, metrics.Worker)
	a := &Activities{
		resource: mockResource,
	}
	defer mockResource.Finish(s.T())
	mockResource.TaskMgr.On("ListTaskList", mock.Anything, mock.Anything).Return(&p.ListTaskListResponse{}, nil).Times(1)
	env := s.NewTestActivityEnvironment()
	env.RegisterActivityWithOptions(a.TaskListScavengerActivity, activity.RegisterOptions{Name: taskListScavengerActivityName})
	env.SetTestTimeout(time.Second)
	_, err := env.ExecuteActivity(taskListScavengerActivityName)
	s.NoError(err)

}
