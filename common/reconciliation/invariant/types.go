// The MIT License (MIT)
//
// Copyright (c) 2017-2020 Uber Technologies Inc.
//
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

package invariant

import (
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/entity"
)

// Invariant represents an invariant of a single execution.
// It can be used to check that the execution satisfies the invariant.
// It can also be used to fix the invariant for an execution.
type Invariant interface {
	Check(interface{}) CheckResult
	Fix(interface{}) FixResult
	InvariantType() Name
}

// Name is the type of an invariant
type Name string

// Collection is a type which indicates a collection of invariants
type Collection int

// Manager represents a manager of several invariants.
// It can be used to run a group of invariant checks or fixes.
type Manager interface {
	RunChecks(interface{}) ManagerCheckResult
	RunFixes(interface{}) ManagerFixResult
}

// ManagerCheckResult is the result of running a list of checks
type ManagerCheckResult struct {
	CheckResultType          CheckResultType
	DeterminingInvariantType *Name
	CheckResults             []CheckResult
}

// ManagerFixResult is the result of running a list of fixes
type ManagerFixResult struct {
	FixResultType            FixResultType
	DeterminingInvariantType *Name
	FixResults               []FixResult
}

type (
	// CheckResultType is the result type of running an invariant check
	CheckResultType string
	// FixResultType is the result type of running an invariant fix
	FixResultType string
)

// CheckResult is the result of running Check.
type CheckResult struct {
	CheckResultType CheckResultType
	InvariantType   Name
	Info            string
	InfoDetails     string
}

// FixResult is the result of running Fix.
type FixResult struct {
	FixResultType FixResultType
	InvariantType Name
	CheckResult   CheckResult
	Info          string
	InfoDetails   string
}

const (
	// CheckResultTypeFailed indicates a failure occurred while attempting to run check
	CheckResultTypeFailed CheckResultType = "failed"
	// CheckResultTypeCorrupted indicates check successfully ran and detected a corruption
	CheckResultTypeCorrupted CheckResultType = "corrupted"
	// CheckResultTypeHealthy indicates check successfully ran and detected no corruption
	CheckResultTypeHealthy CheckResultType = "healthy"

	// FixResultTypeSkipped indicates that fix skipped execution
	FixResultTypeSkipped FixResultType = "skipped"
	// FixResultTypeFixed indicates that fix successfully fixed an execution
	FixResultTypeFixed FixResultType = "fixed"
	// FixResultTypeFailed indicates that fix attempted to fix an execution but failed to do so
	FixResultTypeFailed FixResultType = "failed"

	// HistoryExistsInvariantType asserts that history must exist if concrete execution exists
	HistoryExistsInvariantType Name = "history_exists"
	// OpenCurrentExecutionInvariantType asserts that an open concrete execution must have a valid current execution
	OpenCurrentExecutionInvariantType Name = "open_current_execution"
	// ConcreteExecutionExistsInvariantType asserts that an open current execution must have a valid concrete execution
	ConcreteExecutionExistsInvariantType Name = "concrete_execution_exists"

	// InvariantCollectionMutableState is the collection of invariants relating to mutable state
	InvariantCollectionMutableState Collection = 0
	// InvariantCollectionHistory is the collection  of invariants relating to history
	InvariantCollectionHistory Collection = 1
)

// InvariantTypePtr returns a pointer to Name
func InvariantTypePtr(t Name) *Name {
	return &t
}

// Open returns true if workflow state is open false if workflow is closed
func Open(state int) bool {
	return state == persistence.WorkflowStateCreated || state == persistence.WorkflowStateRunning
}

// ExecutionOpen returns true if execution state is open false if workflow is closed
func ExecutionOpen(execution interface{}) bool {
	return Open(getExecution(execution).State)
}

// getExecution returns base Execution
func getExecution(execution interface{}) *entity.Execution {
	switch e := execution.(type) {
	case *entity.CurrentExecution:
		return &e.Execution
	case *entity.ConcreteExecution:
		return &e.Execution
	default:
		panic("unexpected execution type")
	}
}

// DeleteExecution deletes concrete execution and
// current execution conditionally on matching runID.
func DeleteExecution(
	exec interface{},
	pr persistence.Retryer,
) *FixResult {
	execution := getExecution(exec)
	if err := pr.DeleteWorkflowExecution(&persistence.DeleteWorkflowExecutionRequest{
		DomainID:   execution.DomainID,
		WorkflowID: execution.WorkflowID,
		RunID:      execution.RunID,
	}); err != nil {
		return &FixResult{
			FixResultType: FixResultTypeFailed,
			Info:          "failed to delete concrete workflow execution",
			InfoDetails:   err.Error(),
		}
	}
	if err := pr.DeleteCurrentWorkflowExecution(&persistence.DeleteCurrentWorkflowExecutionRequest{
		DomainID:   execution.DomainID,
		WorkflowID: execution.WorkflowID,
		RunID:      execution.RunID,
	}); err != nil {
		return &FixResult{
			FixResultType: FixResultTypeFailed,
			Info:          "failed to delete current workflow execution",
			InfoDetails:   err.Error(),
		}
	}
	return &FixResult{
		FixResultType: FixResultTypeFixed,
	}
}
