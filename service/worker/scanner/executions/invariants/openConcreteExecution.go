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
	"github.com/uber/cadence/service/worker/scanner/executions/common"
)

type (
	openConcreteExecution struct {
		pr common.PersistenceRetryer
	}
)

// NewOpenConcreteExecution returns a new invariant for checking open concrete execution
func NewOpenConcreteExecution(
	pr common.PersistenceRetryer,
) common.Invariant {
	return &openConcreteExecution{
		pr: pr,
	}
}

func (o *openConcreteExecution) Check(execution common.Execution) common.CheckResult {
	// TODO consider to share code with openCurrentExecution here
	return common.CheckResult{
		CheckResultType: common.CheckResultTypeHealthy,
		InvariantType:   o.InvariantType(),
	}
}

func (o *openConcreteExecution) Fix(execution common.Execution) common.FixResult {
	// TODO consider to share code with openCurrentExecution here
	panic("implement me")
}

func (o *openConcreteExecution) CanApply(execution common.Execution) bool {
	// TODO return true for current execution
	return true
}

func (o *openConcreteExecution) InvariantType() common.InvariantType {
	return common.OpenConcreteExecutionInvariantType
}
