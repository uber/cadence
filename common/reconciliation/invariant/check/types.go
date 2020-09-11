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

//go:generate mockgen -copyright_file ../../../../LICENSE -package $GOPACKAGE -source $GOFILE -destination mocks.go -self_package github.com/uber/cadence/common/reconciliation/invariant/check

package check

// InvariantType is the type of an invariant
type InvariantType string

// CheckResultType is the result type of running an invariant check
type CheckResultType string

// FixResultType is the result type of running an invariant fix
type FixResultType string

// InvariantCollection is a type which indicates a collection of invariants
type InvariantCollection int

// CheckResultTypeFailed indicates a failure occurred while attempting to run check
const CheckResultTypeFailed CheckResultType = "failed"

// CheckResultTypeCorrupted indicates check successfully ran and detected a corruption
const CheckResultTypeCorrupted CheckResultType = "corrupted"

// CheckResultTypeHealthy indicates check successfully ran and detected no corruption
const CheckResultTypeHealthy CheckResultType = "healthy"

// FixResultTypeSkipped indicates that fix skipped execution
const FixResultTypeSkipped FixResultType = "skipped"

// FixResultTypeFixed indicates that fix successfully fixed an execution
const FixResultTypeFixed FixResultType = "fixed"

// FixResultTypeFailed indicates that fix attempted to fix an execution but failed to do so
const FixResultTypeFailed FixResultType = "failed"

// InvariantCollectionMutableState is the collection of invariants relating to mutable state
const InvariantCollectionMutableState InvariantCollection = 0

// InvariantCollectionHistory is the collection  of invariants relating to history
const InvariantCollectionHistory InvariantCollection = 1

// CheckResult is the result of running Check.
type CheckResult struct {
	CheckResultType CheckResultType
	InvariantType   InvariantType
	Info            string
	InfoDetails     string
}

// FixResult is the result of running Fix.
type FixResult struct {
	FixResultType FixResultType
	InvariantType InvariantType
	CheckResult   CheckResult
	Info          string
	InfoDetails   string
}

// InvariantTypePtr returns a pointer to InvariantType
func InvariantTypePtr(t InvariantType) *InvariantType {
	return &t
}

type (
	// ManagerCheckResult is the result of running a list of checks
	ManagerCheckResult struct {
		CheckResultType          CheckResultType
		DeterminingInvariantType *InvariantType
		CheckResults             []CheckResult
	}

	// ManagerFixResult is the result of running a list of fixes
	ManagerFixResult struct {
		FixResultType            FixResultType
		DeterminingInvariantType *InvariantType
		FixResults               []FixResult
	}
)

// Invariant represents an invariant of a single execution.
// It can be used to check that the execution satisfies the invariant.
// It can also be used to fix the invariant for an execution.
type Invariant interface {
	Check(interface{}) CheckResult
	Fix(interface{}) FixResult
	InvariantType() InvariantType
}
