// Copyright (c) 2017 Uber Technologies, Inc.
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

package validator

import (
	"testing"

	"github.com/uber/cadence/common/persistence"

	"github.com/stretchr/testify/suite"

	"github.com/uber/cadence/common/types"
)

type (
	workflowStateCloseStatusSuite struct {
		suite.Suite
	}
)

func TestWorkflowStateCloseStatusSuite(t *testing.T) {
	s := new(workflowStateCloseStatusSuite)
	suite.Run(t, s)
}

func (s *workflowStateCloseStatusSuite) SetupSuite() {
}

func (s *workflowStateCloseStatusSuite) TearDownSuite() {

}

func (s *workflowStateCloseStatusSuite) SetupTest() {

}

func (s *workflowStateCloseStatusSuite) TearDownTest() {

}

func (s *workflowStateCloseStatusSuite) TestCreateWorkflowStateCloseStatus_WorkflowStateCreated() {
	closeStatuses := []int{
		persistence.WorkflowCloseStatusCompleted,
		persistence.WorkflowCloseStatusFailed,
		persistence.WorkflowCloseStatusCanceled,
		persistence.WorkflowCloseStatusTerminated,
		persistence.WorkflowCloseStatusContinuedAsNew,
		persistence.WorkflowCloseStatusTimedOut,
	}

	s.Nil(ValidateCreateWorkflowStateCloseStatus(persistence.WorkflowStateCreated, persistence.WorkflowCloseStatusNone))

	for _, closeStatus := range closeStatuses {
		s.NotNil(ValidateCreateWorkflowStateCloseStatus(persistence.WorkflowStateCreated, closeStatus))
	}
}

func (s *workflowStateCloseStatusSuite) TestCreateWorkflowStateCloseStatus_WorkflowStateRunning() {
	closeStatuses := []int{
		persistence.WorkflowCloseStatusCompleted,
		persistence.WorkflowCloseStatusFailed,
		persistence.WorkflowCloseStatusCanceled,
		persistence.WorkflowCloseStatusTerminated,
		persistence.WorkflowCloseStatusContinuedAsNew,
		persistence.WorkflowCloseStatusTimedOut,
	}

	s.Nil(ValidateCreateWorkflowStateCloseStatus(persistence.WorkflowStateRunning, persistence.WorkflowCloseStatusNone))

	for _, closeStatus := range closeStatuses {
		s.NotNil(ValidateCreateWorkflowStateCloseStatus(persistence.WorkflowStateRunning, closeStatus))
	}
}

func (s *workflowStateCloseStatusSuite) TestCreateWorkflowStateCloseStatus_WorkflowStateCompleted() {
	closeStatuses := []int{
		persistence.WorkflowCloseStatusNone,
		persistence.WorkflowCloseStatusCompleted,
		persistence.WorkflowCloseStatusFailed,
		persistence.WorkflowCloseStatusCanceled,
		persistence.WorkflowCloseStatusTerminated,
		persistence.WorkflowCloseStatusContinuedAsNew,
		persistence.WorkflowCloseStatusTimedOut,
	}

	for _, closeStatus := range closeStatuses {
		s.NotNil(ValidateCreateWorkflowStateCloseStatus(persistence.WorkflowStateCompleted, closeStatus))
	}
}

func (s *workflowStateCloseStatusSuite) TestCreateWorkflowStateCloseStatus_WorkflowStateZombie() {
	closeStatuses := []int{
		persistence.WorkflowCloseStatusCompleted,
		persistence.WorkflowCloseStatusFailed,
		persistence.WorkflowCloseStatusCanceled,
		persistence.WorkflowCloseStatusTerminated,
		persistence.WorkflowCloseStatusContinuedAsNew,
		persistence.WorkflowCloseStatusTimedOut,
	}

	s.Nil(ValidateCreateWorkflowStateCloseStatus(persistence.WorkflowStateZombie, persistence.WorkflowCloseStatusNone))

	for _, closeStatus := range closeStatuses {
		s.NotNil(ValidateCreateWorkflowStateCloseStatus(persistence.WorkflowStateZombie, closeStatus))
	}
}

// TODO

func (s *workflowStateCloseStatusSuite) TestUpdateWorkflowStateCloseStatus_WorkflowStateCreated() {
	closeStatuses := []int{
		persistence.WorkflowCloseStatusCompleted,
		persistence.WorkflowCloseStatusFailed,
		persistence.WorkflowCloseStatusCanceled,
		persistence.WorkflowCloseStatusTerminated,
		persistence.WorkflowCloseStatusContinuedAsNew,
		persistence.WorkflowCloseStatusTimedOut,
	}

	s.Nil(ValidateUpdateWorkflowStateCloseStatus(persistence.WorkflowStateCreated, persistence.WorkflowCloseStatusNone))

	for _, closeStatus := range closeStatuses {
		s.NotNil(ValidateUpdateWorkflowStateCloseStatus(persistence.WorkflowStateCreated, closeStatus))
	}
}

func (s *workflowStateCloseStatusSuite) TestUpdateWorkflowStateCloseStatus_WorkflowStateRunning() {
	closeStatuses := []int{
		persistence.WorkflowCloseStatusCompleted,
		persistence.WorkflowCloseStatusFailed,
		persistence.WorkflowCloseStatusCanceled,
		persistence.WorkflowCloseStatusTerminated,
		persistence.WorkflowCloseStatusContinuedAsNew,
		persistence.WorkflowCloseStatusTimedOut,
	}

	s.Nil(ValidateUpdateWorkflowStateCloseStatus(persistence.WorkflowStateRunning, persistence.WorkflowCloseStatusNone))

	for _, closeStatus := range closeStatuses {
		s.NotNil(ValidateUpdateWorkflowStateCloseStatus(persistence.WorkflowStateRunning, closeStatus))
	}
}

func (s *workflowStateCloseStatusSuite) TestUpdateWorkflowStateCloseStatus_WorkflowStateCompleted() {
	closeStatuses := []int{
		persistence.WorkflowCloseStatusCompleted,
		persistence.WorkflowCloseStatusFailed,
		persistence.WorkflowCloseStatusCanceled,
		persistence.WorkflowCloseStatusTerminated,
		persistence.WorkflowCloseStatusContinuedAsNew,
		persistence.WorkflowCloseStatusTimedOut,
	}

	s.NotNil(ValidateUpdateWorkflowStateCloseStatus(persistence.WorkflowStateCompleted, persistence.WorkflowCloseStatusNone))

	for _, closeStatus := range closeStatuses {
		s.Nil(ValidateUpdateWorkflowStateCloseStatus(persistence.WorkflowStateCompleted, closeStatus))
	}
}

func (s *workflowStateCloseStatusSuite) TestUpdateWorkflowStateCloseStatus_WorkflowStateZombie() {
	closeStatuses := []int{
		persistence.WorkflowCloseStatusCompleted,
		persistence.WorkflowCloseStatusFailed,
		persistence.WorkflowCloseStatusCanceled,
		persistence.WorkflowCloseStatusTerminated,
		persistence.WorkflowCloseStatusContinuedAsNew,
		persistence.WorkflowCloseStatusTimedOut,
	}

	s.Nil(ValidateUpdateWorkflowStateCloseStatus(persistence.WorkflowStateZombie, persistence.WorkflowCloseStatusNone))

	for _, closeStatus := range closeStatuses {
		s.NotNil(ValidateUpdateWorkflowStateCloseStatus(persistence.WorkflowStateZombie, closeStatus))
	}
}

func (s *workflowStateCloseStatusSuite) TestInternalMapping() {
	s.Nil(ToInternalWorkflowExecutionCloseStatus(persistence.WorkflowCloseStatusNone))
	s.Equal(types.WorkflowExecutionCloseStatusCompleted.Ptr(), ToInternalWorkflowExecutionCloseStatus(persistence.WorkflowCloseStatusCompleted))
	s.Equal(types.WorkflowExecutionCloseStatusFailed.Ptr(), ToInternalWorkflowExecutionCloseStatus(persistence.WorkflowCloseStatusFailed))
	s.Equal(types.WorkflowExecutionCloseStatusCanceled.Ptr(), ToInternalWorkflowExecutionCloseStatus(persistence.WorkflowCloseStatusCanceled))
	s.Equal(types.WorkflowExecutionCloseStatusTerminated.Ptr(), ToInternalWorkflowExecutionCloseStatus(persistence.WorkflowCloseStatusTerminated))
	s.Equal(types.WorkflowExecutionCloseStatusContinuedAsNew.Ptr(), ToInternalWorkflowExecutionCloseStatus(persistence.WorkflowCloseStatusContinuedAsNew))
	s.Equal(types.WorkflowExecutionCloseStatusTimedOut.Ptr(), ToInternalWorkflowExecutionCloseStatus(persistence.WorkflowCloseStatusTimedOut))

	s.Equal(persistence.WorkflowCloseStatusNone, FromInternalWorkflowExecutionCloseStatus(nil))
	s.Equal(persistence.WorkflowCloseStatusCompleted, FromInternalWorkflowExecutionCloseStatus(types.WorkflowExecutionCloseStatusCompleted.Ptr()))
	s.Equal(persistence.WorkflowCloseStatusFailed, FromInternalWorkflowExecutionCloseStatus(types.WorkflowExecutionCloseStatusFailed.Ptr()))
	s.Equal(persistence.WorkflowCloseStatusCanceled, FromInternalWorkflowExecutionCloseStatus(types.WorkflowExecutionCloseStatusCanceled.Ptr()))
	s.Equal(persistence.WorkflowCloseStatusTerminated, FromInternalWorkflowExecutionCloseStatus(types.WorkflowExecutionCloseStatusTerminated.Ptr()))
	s.Equal(persistence.WorkflowCloseStatusContinuedAsNew, FromInternalWorkflowExecutionCloseStatus(types.WorkflowExecutionCloseStatusContinuedAsNew.Ptr()))
	s.Equal(persistence.WorkflowCloseStatusTimedOut, FromInternalWorkflowExecutionCloseStatus(types.WorkflowExecutionCloseStatusTimedOut.Ptr()))
}
