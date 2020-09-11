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
	"github.com/uber/cadence/common/reconciliation/invariant/check"
)

type (
	invariantManager struct {
		invariants []check.Invariant
	}
)

// NewInvariantManager handles running a collection of invariants according to the invariant collection provided.
func NewInvariantManager(
	invariants []check.Invariant,
) Manager {
	return &invariantManager{
		invariants: invariants,
	}
}

// RunChecks runs all enabled checks.
func (i *invariantManager) RunChecks(execution interface{}) check.ManagerCheckResult {
	result := check.ManagerCheckResult{
		CheckResultType:          check.CheckResultTypeHealthy,
		DeterminingInvariantType: nil,
		CheckResults:             nil,
	}
	for _, iv := range i.invariants {
		checkResult := iv.Check(execution)
		result.CheckResults = append(result.CheckResults, checkResult)
		checkResultType, updated := i.nextCheckResultType(result.CheckResultType, checkResult.CheckResultType)
		result.CheckResultType = checkResultType
		if updated {
			result.DeterminingInvariantType = &checkResult.InvariantType
		}
	}
	return result
}

// RunFixes runs all enabled fixes.
func (i *invariantManager) RunFixes(execution interface{}) check.ManagerFixResult {
	result := check.ManagerFixResult{
		FixResultType:            check.FixResultTypeSkipped,
		DeterminingInvariantType: nil,
		FixResults:               nil,
	}
	for _, iv := range i.invariants {
		fixResult := iv.Fix(execution)
		result.FixResults = append(result.FixResults, fixResult)
		fixResultType, updated := i.nextFixResultType(result.FixResultType, fixResult.FixResultType)
		result.FixResultType = fixResultType
		if updated {
			result.DeterminingInvariantType = &fixResult.InvariantType
		}
	}
	return result
}

func (i *invariantManager) nextFixResultType(
	currentState check.FixResultType,
	event check.FixResultType,
) (check.FixResultType, bool) {
	switch currentState {
	case check.FixResultTypeSkipped:
		return event, event != check.FixResultTypeSkipped
	case check.FixResultTypeFixed:
		if event == check.FixResultTypeFailed {
			return event, true
		}
		return currentState, false
	case check.FixResultTypeFailed:
		return currentState, false
	default:
		panic("unknown FixResultType")
	}
}

func (i *invariantManager) nextCheckResultType(
	currentState check.CheckResultType,
	event check.CheckResultType,
) (check.CheckResultType, bool) {
	switch currentState {
	case check.CheckResultTypeHealthy:
		return event, event != check.CheckResultTypeHealthy
	case check.CheckResultTypeCorrupted:
		if event == check.CheckResultTypeFailed {
			return event, true
		}
		return currentState, false
	case check.CheckResultTypeFailed:
		return currentState, false
	default:
		panic("unknown CheckResultType")
	}
}

// Manager represents a manager of several invariants.
// It can be used to run a group of invariant checks or fixes.
type Manager interface {
	RunChecks(interface{}) check.ManagerCheckResult
	RunFixes(interface{}) check.ManagerFixResult
}
