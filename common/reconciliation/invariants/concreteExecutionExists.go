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

package invariants

import (
	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common/persistence"
	"github.com/uber/cadence/common/reconciliation/common"
)

type (
	concreteExecutionExist struct {
		pr common.PersistenceRetryer
	}
)

// NewConcreteExecutionExists returns a new invariant for checking concrete execution
func NewConcreteExecutionExists(
	pr common.PersistenceRetryer,
) common.Invariant {
	return &concreteExecutionExist{
		pr: pr,
	}
}

func (c *concreteExecutionExist) Check(execution interface{}) common.CheckResult {
	currentExecution, ok := execution.(*common.CurrentExecution)
	if !ok {
		return common.CheckResult{
			CheckResultType: common.CheckResultTypeFailed,
			InvariantType:   c.InvariantType(),
			Info:            "failed to check: expected current execution",
		}
	}
	if !common.Open(currentExecution.State) {
		return common.CheckResult{
			CheckResultType: common.CheckResultTypeHealthy,
			InvariantType:   c.InvariantType(),
		}
	}
	// by this point the corresponding concrete execution must exist and can be already closed
	_, currentExecErr := c.pr.GetConcreteExecution(&persistence.GetConcreteExecutionRequest{
		DomainID:   currentExecution.DomainID,
		WorkflowID: "cadence-sys-executions-scanner",
		RunID:      currentExecution.CurrentRunID,
	})
	if currentExecErr != nil {
		switch currentExecErr.(type) {
		case *shared.EntityNotExistsError:
			return common.CheckResult{
				CheckResultType: common.CheckResultTypeCorrupted,
				InvariantType:   c.InvariantType(),
				Info:            "execution is open without having current execution",
				InfoDetails:     currentExecErr.Error(),
			}
		default:
			return common.CheckResult{
				CheckResultType: common.CheckResultTypeFailed,
				InvariantType:   c.InvariantType(),
				Info:            "failed to check if current execution exists",
				InfoDetails:     currentExecErr.Error(),
			}
		}
	}
	return common.CheckResult{
		CheckResultType: common.CheckResultTypeHealthy,
		InvariantType:   c.InvariantType(),
	}
}

func (c *concreteExecutionExist) Fix(execution interface{}) common.FixResult {
	fixResult, checkResult := checkBeforeFix(c, execution)
	if fixResult != nil {
		return *fixResult
	}
	currentExecution, _ := execution.(*common.CurrentExecution)
	if err := c.pr.DeleteCurrentWorkflowExecution(&persistence.DeleteCurrentWorkflowExecutionRequest{
		DomainID:   currentExecution.DomainID,
		WorkflowID: "cadence-sys-executions-scanner",
		RunID:      currentExecution.RunID,
	}); err != nil {
		return common.FixResult{
			FixResultType: common.FixResultTypeFailed,
			Info:          "failed to delete current workflow execution",
			InfoDetails:   err.Error(),
		}
	}

	fixResult.CheckResult = *checkResult
	fixResult.InvariantType = c.InvariantType()
	return *fixResult
}

func (o *concreteExecutionExist) InvariantType() common.InvariantType {
	return common.ConcreteExecutionExistsInvariantType
}
