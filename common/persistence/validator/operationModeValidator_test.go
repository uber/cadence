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
	"fmt"
	"github.com/uber/cadence/common/persistence"
	"testing"

	"github.com/stretchr/testify/suite"
)

type (
	validateOperationWorkflowModeStateSuite struct {
		suite.Suite
	}
)

func TestValidateOperationWorkflowModeStateSuite(t *testing.T) {
	s := new(validateOperationWorkflowModeStateSuite)
	suite.Run(t, s)
}

func (s *validateOperationWorkflowModeStateSuite) SetupSuite() {
}

func (s *validateOperationWorkflowModeStateSuite) TearDownSuite() {

}

func (s *validateOperationWorkflowModeStateSuite) SetupTest() {

}

func (s *validateOperationWorkflowModeStateSuite) TearDownTest() {

}

func (s *validateOperationWorkflowModeStateSuite) TestCreateMode_UpdateCurrent() {

	stateToError := map[int]bool{
		persistence.WorkflowStateCreated:   false,
		persistence.WorkflowStateRunning:   false,
		persistence.WorkflowStateCompleted: true,
		persistence.WorkflowStateZombie:    true,
	}

	creatModes := []persistence.CreateWorkflowMode{
		persistence.CreateWorkflowModeBrandNew,
		persistence.CreateWorkflowModeWorkflowIDReuse,
		persistence.CreateWorkflowModeContinueAsNew,
	}

	for state, expectError := range stateToError {
		testSnapshot := s.newTestWorkflowSnapshot(state)
		for _, createMode := range creatModes {
			err := ValidateCreateWorkflowModeState(createMode, testSnapshot)
			if !expectError {
				s.NoError(err, err)
			} else {
				s.Error(err, err)
			}
		}
	}
}

func (s *validateOperationWorkflowModeStateSuite) TestCreateMode_BypassCurrent() {

	stateToError := map[int]bool{
		persistence.WorkflowStateCreated:   true,
		persistence.WorkflowStateRunning:   true,
		persistence.WorkflowStateCompleted: true,
		persistence.WorkflowStateZombie:    false,
	}

	for state, expectError := range stateToError {
		testSnapshot := s.newTestWorkflowSnapshot(state)
		err := ValidateCreateWorkflowModeState(persistence.CreateWorkflowModeZombie, testSnapshot)
		if !expectError {
			s.NoError(err, err)
		} else {
			s.Error(err, err)
		}
	}
}

func (s *validateOperationWorkflowModeStateSuite) TestUpdateMode_UpdateCurrent() {

	// only current workflow
	stateToError := map[int]bool{
		persistence.WorkflowStateCreated:   false,
		persistence.WorkflowStateRunning:   false,
		persistence.WorkflowStateCompleted: false,
		persistence.WorkflowStateZombie:    true,
	}
	for state, expectError := range stateToError {
		testCurrentMutation := s.newTestWorkflowMutation(state)
		err := ValidateUpdateWorkflowModeState(
			persistence.UpdateWorkflowModeUpdateCurrent,
			testCurrentMutation,
			nil,
		)
		if !expectError {
			s.NoError(err, err)
		} else {
			s.Error(err, err)
		}
	}

	// current workflow & new workflow
	currentStateToError := map[int]bool{
		persistence.WorkflowStateCreated:   true,
		persistence.WorkflowStateRunning:   true,
		persistence.WorkflowStateCompleted: false,
		persistence.WorkflowStateZombie:    false,
	}
	newStateToError := map[int]bool{
		persistence.WorkflowStateCreated:   false,
		persistence.WorkflowStateRunning:   false,
		persistence.WorkflowStateCompleted: true,
		persistence.WorkflowStateZombie:    true,
	}
	for currentState, currentExpectError := range currentStateToError {
		for newState, newExpectError := range newStateToError {
			testCurrentMutation := s.newTestWorkflowMutation(currentState)
			testNewSnapshot := s.newTestWorkflowSnapshot(newState)
			err := ValidateUpdateWorkflowModeState(
				persistence.UpdateWorkflowModeUpdateCurrent,
				testCurrentMutation,
				&testNewSnapshot,
			)
			if currentExpectError || newExpectError {
				s.Error(err, err)
			} else {
				s.NoError(err, err)
			}
		}
	}
}

func (s *validateOperationWorkflowModeStateSuite) TestUpdateMode_BypassCurrent() {

	// only current workflow
	stateToError := map[int]bool{
		persistence.WorkflowStateCreated:   true,
		persistence.WorkflowStateRunning:   true,
		persistence.WorkflowStateCompleted: false,
		persistence.WorkflowStateZombie:    false,
	}
	for state, expectError := range stateToError {
		testMutation := s.newTestWorkflowMutation(state)
		err := ValidateUpdateWorkflowModeState(
			persistence.UpdateWorkflowModeBypassCurrent,
			testMutation,
			nil,
		)
		if !expectError {
			s.NoError(err, err)
		} else {
			s.Error(err, err)
		}
	}

	// current workflow & new workflow
	currentStateToError := map[int]bool{
		persistence.WorkflowStateCreated:   true,
		persistence.WorkflowStateRunning:   true,
		persistence.WorkflowStateCompleted: false,
		persistence.WorkflowStateZombie:    false,
	}
	newStateToError := map[int]bool{
		persistence.WorkflowStateCreated:   true,
		persistence.WorkflowStateRunning:   true,
		persistence.WorkflowStateCompleted: true,
		persistence.WorkflowStateZombie:    false,
	}
	for currentState, currentExpectError := range currentStateToError {
		for newState, newExpectError := range newStateToError {
			testCurrentMutation := s.newTestWorkflowMutation(currentState)
			testNewSnapshot := s.newTestWorkflowSnapshot(newState)
			err := ValidateUpdateWorkflowModeState(
				persistence.UpdateWorkflowModeBypassCurrent,
				testCurrentMutation,
				&testNewSnapshot,
			)
			if currentExpectError || newExpectError {
				s.Error(err, err)
			} else {
				s.NoError(err, err)
			}
		}
	}
}

func (s *validateOperationWorkflowModeStateSuite) TestConflictResolveMode_UpdateCurrent() {

	// only reset workflow
	stateToError := map[int]bool{
		persistence.WorkflowStateCreated:   false,
		persistence.WorkflowStateRunning:   false,
		persistence.WorkflowStateCompleted: false,
		persistence.WorkflowStateZombie:    true,
	}
	for state, expectError := range stateToError {
		testSnapshot := s.newTestWorkflowSnapshot(state)
		err := ValidateConflictResolveWorkflowModeState(
			persistence.ConflictResolveWorkflowModeUpdateCurrent,
			testSnapshot,
			nil,
			nil,
		)
		if !expectError {
			s.NoError(err, err)
		} else {
			s.Error(err, err)
		}
	}

	// reset workflow & new workflow
	resetStateToError := map[int]bool{
		persistence.WorkflowStateCreated:   true,
		persistence.WorkflowStateRunning:   true,
		persistence.WorkflowStateCompleted: false,
		persistence.WorkflowStateZombie:    true,
	}
	newStateToError := map[int]bool{
		persistence.WorkflowStateCreated:   false,
		persistence.WorkflowStateRunning:   false,
		persistence.WorkflowStateCompleted: true,
		persistence.WorkflowStateZombie:    true,
	}
	for resetState, resetExpectError := range resetStateToError {
		for newState, newExpectError := range newStateToError {
			testResetSnapshot := s.newTestWorkflowSnapshot(resetState)
			testNewSnapshot := s.newTestWorkflowSnapshot(newState)
			err := ValidateConflictResolveWorkflowModeState(
				persistence.ConflictResolveWorkflowModeUpdateCurrent,
				testResetSnapshot,
				&testNewSnapshot,
				nil,
			)
			if resetExpectError || newExpectError {
				s.Error(err, err)
			} else {
				s.NoError(err, err)
			}
		}
	}

	// reset workflow & current workflow
	resetStateToError = map[int]bool{
		persistence.WorkflowStateCreated:   false,
		persistence.WorkflowStateRunning:   false,
		persistence.WorkflowStateCompleted: false,
		persistence.WorkflowStateZombie:    true,
	}
	currentStateToError := map[int]bool{
		persistence.WorkflowStateCreated:   true,
		persistence.WorkflowStateRunning:   true,
		persistence.WorkflowStateCompleted: false,
		persistence.WorkflowStateZombie:    false,
	}
	for resetState, resetExpectError := range resetStateToError {
		for currentState, currentExpectError := range currentStateToError {
			testResetSnapshot := s.newTestWorkflowSnapshot(resetState)
			testCurrentSnapshot := s.newTestWorkflowMutation(currentState)
			err := ValidateConflictResolveWorkflowModeState(
				persistence.ConflictResolveWorkflowModeUpdateCurrent,
				testResetSnapshot,
				nil,
				&testCurrentSnapshot,
			)
			if resetExpectError || currentExpectError {
				s.Error(err, err)
			} else {
				s.NoError(err, err)
			}
		}
	}

	// reset workflow & new workflow & current workflow
	resetStateToError = map[int]bool{
		persistence.WorkflowStateCreated:   true,
		persistence.WorkflowStateRunning:   true,
		persistence.WorkflowStateCompleted: false,
		persistence.WorkflowStateZombie:    true,
	}
	newStateToError = map[int]bool{
		persistence.WorkflowStateCreated:   false,
		persistence.WorkflowStateRunning:   false,
		persistence.WorkflowStateCompleted: true,
		persistence.WorkflowStateZombie:    true,
	}
	currentStateToError = map[int]bool{
		persistence.WorkflowStateCreated:   true,
		persistence.WorkflowStateRunning:   true,
		persistence.WorkflowStateCompleted: false,
		persistence.WorkflowStateZombie:    false,
	}
	for resetState, resetExpectError := range resetStateToError {
		for newState, newExpectError := range newStateToError {
			for currentState, currentExpectError := range currentStateToError {
				testResetSnapshot := s.newTestWorkflowSnapshot(resetState)
				testNewSnapshot := s.newTestWorkflowSnapshot(newState)
				testCurrentSnapshot := s.newTestWorkflowMutation(currentState)
				err := ValidateConflictResolveWorkflowModeState(
					persistence.ConflictResolveWorkflowModeUpdateCurrent,
					testResetSnapshot,
					&testNewSnapshot,
					&testCurrentSnapshot,
				)
				if resetExpectError || newExpectError || currentExpectError {
					s.Error(err, err)
				} else {
					s.NoError(err, err)
				}
			}
		}
	}
}

func (s *validateOperationWorkflowModeStateSuite) TestConflictResolveMode_BypassCurrent() {

	// only reset workflow
	stateToError := map[int]bool{
		persistence.WorkflowStateCreated:   true,
		persistence.WorkflowStateRunning:   true,
		persistence.WorkflowStateCompleted: false,
		persistence.WorkflowStateZombie:    false,
	}
	for state, expectError := range stateToError {
		testSnapshot := s.newTestWorkflowSnapshot(state)
		err := ValidateConflictResolveWorkflowModeState(
			persistence.ConflictResolveWorkflowModeBypassCurrent,
			testSnapshot,
			nil,
			nil,
		)
		if !expectError {
			s.NoError(err, err)
		} else {
			s.Error(err, err)
		}
	}

	// reset workflow & new workflow
	resetStateToError := map[int]bool{
		persistence.WorkflowStateCreated:   true,
		persistence.WorkflowStateRunning:   true,
		persistence.WorkflowStateCompleted: false,
		persistence.WorkflowStateZombie:    true,
	}
	newStateToError := map[int]bool{
		persistence.WorkflowStateCreated:   true,
		persistence.WorkflowStateRunning:   true,
		persistence.WorkflowStateCompleted: true,
		persistence.WorkflowStateZombie:    false,
	}
	for resetState, resetExpectError := range resetStateToError {
		for newState, newExpectError := range newStateToError {
			testResetSnapshot := s.newTestWorkflowSnapshot(resetState)
			testNewSnapshot := s.newTestWorkflowSnapshot(newState)
			err := ValidateConflictResolveWorkflowModeState(
				persistence.ConflictResolveWorkflowModeBypassCurrent,
				testResetSnapshot,
				&testNewSnapshot,
				nil,
			)
			if resetExpectError || newExpectError {
				if err == nil {
					fmt.Print("##")
				}
				s.Error(err, err)
			} else {
				s.NoError(err, err)
			}
		}
	}
}

func (s *validateOperationWorkflowModeStateSuite) newTestWorkflowSnapshot(
	state int,
) persistence.InternalWorkflowSnapshot {
	return persistence.InternalWorkflowSnapshot{
		ExecutionInfo: &persistence.InternalWorkflowExecutionInfo{
			State: state,
		},
	}
}

func (s *validateOperationWorkflowModeStateSuite) newTestWorkflowMutation(
	state int,
) persistence.InternalWorkflowMutation {
	return persistence.InternalWorkflowMutation{
		ExecutionInfo: &persistence.InternalWorkflowExecutionInfo{
			State: state,
		},
	}
}
