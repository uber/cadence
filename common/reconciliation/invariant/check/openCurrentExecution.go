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

package check

import (
	"fmt"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/entity"
)

// OpenCurrentExecutionInvariantType asserts that an open concrete execution must have a valid current execution
const OpenCurrentExecutionInvariantType InvariantType = "open_current_execution"

type (
	openCurrentExecution struct {
		pr persistence.Retryer
	}
)

// NewOpenCurrentExecution returns a new invariant for checking open current execution
func NewOpenCurrentExecution(
	pr persistence.Retryer,
) Invariant {
	return &openCurrentExecution{
		pr: pr,
	}
}

func (o *openCurrentExecution) Check(execution interface{}) CheckResult {
	concreteExecution, ok := execution.(*entity.ConcreteExecution)
	if !ok {
		return CheckResult{
			CheckResultType: CheckResultTypeFailed,
			InvariantType:   o.InvariantType(),
			Info:            "failed to check: expected concrete execution",
		}
	}
	if !Open(concreteExecution.State) {
		return CheckResult{
			CheckResultType: CheckResultTypeHealthy,
			InvariantType:   o.InvariantType(),
		}
	}
	currentExecResp, currentExecErr := o.pr.GetCurrentExecution(&persistence.GetCurrentExecutionRequest{
		DomainID:   concreteExecution.DomainID,
		WorkflowID: concreteExecution.WorkflowID,
	})
	stillOpen, stillOpenErr := ExecutionStillOpen(&concreteExecution.Execution, o.pr)
	if stillOpenErr != nil {
		return CheckResult{
			CheckResultType: CheckResultTypeFailed,
			InvariantType:   o.InvariantType(),
			Info:            "failed to check if concrete execution is still open",
			InfoDetails:     stillOpenErr.Error(),
		}
	}
	if !stillOpen {
		return CheckResult{
			CheckResultType: CheckResultTypeHealthy,
			InvariantType:   o.InvariantType(),
		}
	}
	if currentExecErr != nil {
		switch currentExecErr.(type) {
		case *shared.EntityNotExistsError:
			return CheckResult{
				CheckResultType: CheckResultTypeCorrupted,
				InvariantType:   o.InvariantType(),
				Info:            "execution is open without having current execution",
				InfoDetails:     currentExecErr.Error(),
			}
		default:
			return CheckResult{
				CheckResultType: CheckResultTypeFailed,
				InvariantType:   o.InvariantType(),
				Info:            "failed to check if current execution exists",
				InfoDetails:     currentExecErr.Error(),
			}
		}
	}
	if currentExecResp.RunID != concreteExecution.RunID {
		return CheckResult{
			CheckResultType: CheckResultTypeCorrupted,
			InvariantType:   o.InvariantType(),
			Info:            "execution is open but current points at a different execution",
			InfoDetails:     fmt.Sprintf("current points at %v", currentExecResp.RunID),
		}
	}
	return CheckResult{
		CheckResultType: CheckResultTypeHealthy,
		InvariantType:   o.InvariantType(),
	}
}

func (o *openCurrentExecution) Fix(execution interface{}) FixResult {
	fixResult, checkResult := checkBeforeFix(o, execution)
	if fixResult != nil {
		return *fixResult
	}
	fixResult = DeleteExecution(&execution, o.pr)
	fixResult.CheckResult = *checkResult
	fixResult.InvariantType = o.InvariantType()
	return *fixResult
}

func (o *openCurrentExecution) InvariantType() InvariantType {
	return OpenCurrentExecutionInvariantType
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

// ExecutionStillOpen returns true if execution in persistence exists and is open, false otherwise.
// Returns error on failure to confirm.
func ExecutionStillOpen(
	exec *entity.Execution,
	pr persistence.Retryer,
) (bool, error) {
	req := &persistence.GetWorkflowExecutionRequest{
		DomainID: exec.DomainID,
		Execution: shared.WorkflowExecution{
			WorkflowId: &exec.WorkflowID,
			RunId:      &exec.RunID,
		},
	}
	resp, err := pr.GetWorkflowExecution(req)
	if err != nil {
		switch err.(type) {
		case *shared.EntityNotExistsError:
			return false, nil
		default:
			return false, err
		}
	}
	return Open(resp.State.ExecutionInfo.State), nil
}

// ExecutionStillExists returns true if execution still exists in persistence, false otherwise.
// Returns error on failure to confirm.
func ExecutionStillExists(
	exec *entity.Execution,
	pr persistence.Retryer,
) (bool, error) {
	req := &persistence.GetWorkflowExecutionRequest{
		DomainID: exec.DomainID,
		Execution: shared.WorkflowExecution{
			WorkflowId: &exec.WorkflowID,
			RunId:      &exec.RunID,
		},
	}
	_, err := pr.GetWorkflowExecution(req)
	if err == nil {
		return true, nil
	}
	switch err.(type) {
	case *shared.EntityNotExistsError:
		return false, nil
	default:
		return false, err
	}
}

// Open returns true if workflow state is open false if workflow is closed
func Open(state int) bool {
	return state == persistence.WorkflowStateCreated || state == persistence.WorkflowStateRunning
}
