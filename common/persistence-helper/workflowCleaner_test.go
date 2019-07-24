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

package persistencehelper

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/log"
	"github.com/uber/cadence/common/mocks"
	"github.com/uber/cadence/common/persistence"
)

type (
	workflowCleanerSuite struct {
		*require.Assertions
		suite.Suite

		mockExecutionManager  *mocks.ExecutionManager
		mockVisibilityManager *mocks.VisibilityManager
		mockHistoryManager    *mocks.HistoryManager
		mockHistoryV2Manager  *mocks.HistoryV2Manager

		workflowCleaner WorkflowCleaner
		testRequest     *CleanUpRequest
	}
)

func TestWorkflowCleanerSuite(t *testing.T) {
	suite.Run(t, new(workflowCleanerSuite))
}

func (s *workflowCleanerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.mockExecutionManager = &mocks.ExecutionManager{}
	s.mockHistoryManager = &mocks.HistoryManager{}
	s.mockHistoryV2Manager = &mocks.HistoryV2Manager{}
	s.mockVisibilityManager = &mocks.VisibilityManager{}

	s.workflowCleaner = NewWorkflowCleaner(
		s.mockExecutionManager,
		s.mockHistoryManager,
		s.mockHistoryV2Manager,
		s.mockVisibilityManager,
		log.NewNoop(),
	)
	s.testRequest = &CleanUpRequest{
		DomainID:        "some random domainID",
		WorkflowID:      "some random workflowID",
		RunID:           "some random runID",
		ShardID:         123,
		FailoverVersion: 1234,
		TaskID:          12345,
	}
}

func (s *workflowCleanerSuite) TearDownTest() {
	s.mockExecutionManager.AssertExpectations(s.T())
	s.mockHistoryManager.AssertExpectations(s.T())
	s.mockHistoryV2Manager.AssertExpectations(s.T())
	s.mockVisibilityManager.AssertExpectations(s.T())
}

func (s *workflowCleanerSuite) TestCleanUp_Fail_PersistenceNonRetryableErr() {
	s.mockExecutionManager.On("DeleteCurrentWorkflowExecution", mock.Anything).Return(nil).Once()
	s.mockExecutionManager.On("DeleteWorkflowExecution", mock.Anything).Return(nil).Once()
	s.mockHistoryManager.On("DeleteWorkflowExecutionHistory", mock.Anything).Return(errors.New("some random error")).Once()

	err := s.workflowCleaner.CleanUp(s.testRequest)
	s.Error(err)
}

func (s *workflowCleanerSuite) TestCleanUp_Success_EntityNotExistErr() {
	notExistErr := &shared.EntityNotExistsError{}
	s.mockExecutionManager.On("DeleteCurrentWorkflowExecution", mock.Anything).Return(notExistErr).Once()
	s.mockExecutionManager.On("DeleteWorkflowExecution", mock.Anything).Return(notExistErr).Once()
	s.mockHistoryManager.On("DeleteWorkflowExecutionHistory", mock.Anything).Return(notExistErr).Once()
	s.mockVisibilityManager.On("DeleteWorkflowExecution", mock.Anything).Return(notExistErr).Once()

	err := s.workflowCleaner.CleanUp(s.testRequest)
	s.NoError(err)
}

func (s *workflowCleanerSuite) TestCleanUp_Success_EventsV1() {
	s.mockExecutionManager.On("DeleteCurrentWorkflowExecution", mock.Anything).Return(nil).Once()
	s.mockExecutionManager.On("DeleteWorkflowExecution", mock.Anything).Return(nil).Once()
	s.mockHistoryManager.On("DeleteWorkflowExecutionHistory", mock.Anything).Return(nil).Once()
	s.mockVisibilityManager.On("DeleteWorkflowExecution", mock.Anything).Return(nil).Once()

	err := s.workflowCleaner.CleanUp(s.testRequest)
	s.NoError(err)
}

func (s *workflowCleanerSuite) TestCleanUp_Success_EventsV2() {
	s.mockExecutionManager.On("DeleteCurrentWorkflowExecution", mock.Anything).Return(nil).Once()
	s.mockExecutionManager.On("DeleteWorkflowExecution", mock.Anything).Return(nil).Once()
	s.mockHistoryV2Manager.On("DeleteHistoryBranch", mock.Anything).Return(nil).Once()
	s.mockVisibilityManager.On("DeleteWorkflowExecution", mock.Anything).Return(nil).Once()

	s.testRequest.EventStoreVersion = persistence.EventStoreVersionV2
	s.testRequest.BranchToken = []byte{'1', '2', '3'}
	err := s.workflowCleaner.CleanUp(s.testRequest)
	s.NoError(err)
}
