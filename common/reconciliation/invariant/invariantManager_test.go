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
	"testing"

	"github.com/uber/cadence/common/reconciliation/entity"
	"github.com/uber/cadence/common/reconciliation/invariant/check"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type InvariantManagerSuite struct {
	*require.Assertions
	suite.Suite
	controller *gomock.Controller
}

func TestInvariantManagerSuite(t *testing.T) {
	suite.Run(t, new(InvariantManagerSuite))
}

func (s *InvariantManagerSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.controller = gomock.NewController(s.T())
}

func (s *InvariantManagerSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *InvariantManagerSuite) TestRunChecks() {
	testCases := []struct {
		checkResults []check.CheckResult
		expected     check.ManagerCheckResult
	}{
		{
			checkResults: nil,
			expected: check.ManagerCheckResult{
				CheckResultType: check.CheckResultTypeHealthy,
				CheckResults:    nil,
			},
		},
		{
			checkResults: []check.CheckResult{
				{
					CheckResultType: check.CheckResultTypeHealthy,
					InvariantType:   check.InvariantType("first"),
					Info:            "invariant 1 info",
					InfoDetails:     "invariant 1 info details",
				},
				{
					CheckResultType: check.CheckResultTypeFailed,
					InvariantType:   check.InvariantType("second"),
					Info:            "invariant 2 info",
					InfoDetails:     "invariant 2 info details",
				},
			},
			expected: check.ManagerCheckResult{
				CheckResultType:          check.CheckResultTypeFailed,
				DeterminingInvariantType: check.InvariantTypePtr("second"),
				CheckResults: []check.CheckResult{
					{
						CheckResultType: check.CheckResultTypeHealthy,
						InvariantType:   check.InvariantType("first"),
						Info:            "invariant 1 info",
						InfoDetails:     "invariant 1 info details",
					},
					{
						CheckResultType: check.CheckResultTypeFailed,
						InvariantType:   check.InvariantType("second"),
						Info:            "invariant 2 info",
						InfoDetails:     "invariant 2 info details",
					},
				},
			},
		},
		{
			checkResults: []check.CheckResult{
				{
					CheckResultType: check.CheckResultTypeHealthy,
					InvariantType:   check.InvariantType("first"),
					Info:            "invariant 1 info",
					InfoDetails:     "invariant 1 info details",
				},
				{
					CheckResultType: check.CheckResultTypeCorrupted,
					InvariantType:   check.InvariantType("second"),
					Info:            "invariant 2 info",
					InfoDetails:     "invariant 2 info details",
				},
			},
			expected: check.ManagerCheckResult{
				CheckResultType:          check.CheckResultTypeCorrupted,
				DeterminingInvariantType: check.InvariantTypePtr("second"),
				CheckResults: []check.CheckResult{
					{
						CheckResultType: check.CheckResultTypeHealthy,
						InvariantType:   check.InvariantType("first"),
						Info:            "invariant 1 info",
						InfoDetails:     "invariant 1 info details",
					},
					{
						CheckResultType: check.CheckResultTypeCorrupted,
						InvariantType:   check.InvariantType("second"),
						Info:            "invariant 2 info",
						InfoDetails:     "invariant 2 info details",
					},
				},
			},
		},
		{
			checkResults: []check.CheckResult{
				{
					CheckResultType: check.CheckResultTypeHealthy,
					InvariantType:   check.InvariantType("first"),
					Info:            "invariant 1 info",
					InfoDetails:     "invariant 1 info details",
				},
				{
					CheckResultType: check.CheckResultTypeHealthy,
					InvariantType:   check.InvariantType("second"),
					Info:            "invariant 2 info",
					InfoDetails:     "invariant 2 info details",
				},
			},
			expected: check.ManagerCheckResult{
				CheckResultType:          check.CheckResultTypeHealthy,
				DeterminingInvariantType: nil,
				CheckResults: []check.CheckResult{
					{
						CheckResultType: check.CheckResultTypeHealthy,
						InvariantType:   check.InvariantType("first"),
						Info:            "invariant 1 info",
						InfoDetails:     "invariant 1 info details",
					},
					{
						CheckResultType: check.CheckResultTypeHealthy,
						InvariantType:   check.InvariantType("second"),
						Info:            "invariant 2 info",
						InfoDetails:     "invariant 2 info details",
					},
				},
			},
		},
		{
			checkResults: []check.CheckResult{
				{
					CheckResultType: check.CheckResultTypeHealthy,
					InvariantType:   check.InvariantType("first"),
					Info:            "invariant 1 info",
					InfoDetails:     "invariant 1 info details",
				},
				{
					CheckResultType: check.CheckResultTypeCorrupted,
					InvariantType:   check.InvariantType("second"),
					Info:            "invariant 2 info",
					InfoDetails:     "invariant 2 info details",
				},
				{
					CheckResultType: check.CheckResultTypeFailed,
					InvariantType:   check.InvariantType("third"),
					Info:            "invariant 3 info",
					InfoDetails:     "invariant 3 info details",
				},
				{
					CheckResultType: check.CheckResultTypeHealthy,
					InvariantType:   check.InvariantType("forth"),
					Info:            "invariant 4 info",
					InfoDetails:     "invariant 4 info details",
				},
			},
			expected: check.ManagerCheckResult{
				CheckResultType:          check.CheckResultTypeFailed,
				DeterminingInvariantType: check.InvariantTypePtr("third"),
				CheckResults: []check.CheckResult{
					{
						CheckResultType: check.CheckResultTypeHealthy,
						InvariantType:   check.InvariantType("first"),
						Info:            "invariant 1 info",
						InfoDetails:     "invariant 1 info details",
					},
					{
						CheckResultType: check.CheckResultTypeCorrupted,
						InvariantType:   check.InvariantType("second"),
						Info:            "invariant 2 info",
						InfoDetails:     "invariant 2 info details",
					},
					{
						CheckResultType: check.CheckResultTypeFailed,
						InvariantType:   check.InvariantType("third"),
						Info:            "invariant 3 info",
						InfoDetails:     "invariant 3 info details",
					},
					{
						CheckResultType: check.CheckResultTypeHealthy,
						InvariantType:   check.InvariantType("forth"),
						Info:            "invariant 4 info",
						InfoDetails:     "invariant 4 info details",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		invariants := make([]check.Invariant, len(tc.checkResults), len(tc.checkResults))
		for i := 0; i < len(tc.checkResults); i++ {
			mockInvariant := check.NewMockInvariant(s.controller)
			mockInvariant.EXPECT().Check(gomock.Any()).Return(tc.checkResults[i])
			invariants[i] = mockInvariant
		}
		manager := &invariantManager{
			invariants: invariants,
		}
		s.Equal(tc.expected, manager.RunChecks(entity.Execution{}))
	}
}

func (s *InvariantManagerSuite) TestRunFixes() {
	testCases := []struct {
		fixResults []check.FixResult
		expected   check.ManagerFixResult
	}{
		{
			fixResults: nil,
			expected: check.ManagerFixResult{
				FixResultType:            check.FixResultTypeSkipped,
				DeterminingInvariantType: nil,
				FixResults:               nil,
			},
		},
		{
			fixResults: []check.FixResult{
				{
					FixResultType: check.FixResultTypeFixed,
					InvariantType: check.InvariantType("first"),
					CheckResult: check.CheckResult{
						CheckResultType: check.CheckResultTypeCorrupted,
						Info:            "invariant 1 check info",
						InfoDetails:     "invariant 1 check info details",
					},
					Info:        "invariant 1 info",
					InfoDetails: "invariant 1 info details",
				},
				{
					FixResultType: check.FixResultTypeFailed,
					InvariantType: check.InvariantType("second"),
					CheckResult: check.CheckResult{
						CheckResultType: check.CheckResultTypeCorrupted,
						Info:            "invariant 2 check info",
						InfoDetails:     "invariant 2 check info details",
					},
					Info:        "invariant 2 info",
					InfoDetails: "invariant 2 info details",
				},
			},
			expected: check.ManagerFixResult{
				FixResultType:            check.FixResultTypeFailed,
				DeterminingInvariantType: check.InvariantTypePtr("second"),
				FixResults: []check.FixResult{
					{
						FixResultType: check.FixResultTypeFixed,
						InvariantType: check.InvariantType("first"),
						CheckResult: check.CheckResult{
							CheckResultType: check.CheckResultTypeCorrupted,
							Info:            "invariant 1 check info",
							InfoDetails:     "invariant 1 check info details",
						},
						Info:        "invariant 1 info",
						InfoDetails: "invariant 1 info details",
					},
					{
						FixResultType: check.FixResultTypeFailed,
						InvariantType: check.InvariantType("second"),
						CheckResult: check.CheckResult{
							CheckResultType: check.CheckResultTypeCorrupted,
							Info:            "invariant 2 check info",
							InfoDetails:     "invariant 2 check info details",
						},
						Info:        "invariant 2 info",
						InfoDetails: "invariant 2 info details",
					},
				},
			},
		},
		{
			fixResults: []check.FixResult{
				{
					FixResultType: check.FixResultTypeSkipped,
					InvariantType: check.InvariantType("first"),
					CheckResult: check.CheckResult{
						CheckResultType: check.CheckResultTypeHealthy,
						Info:            "invariant 1 check info",
						InfoDetails:     "invariant 1 check info details",
					},
					Info:        "invariant 1 info",
					InfoDetails: "invariant 1 info details",
				},
				{
					FixResultType: check.FixResultTypeSkipped,
					InvariantType: check.InvariantType("second"),
					CheckResult: check.CheckResult{
						CheckResultType: check.CheckResultTypeHealthy,
						Info:            "invariant 2 check info",
						InfoDetails:     "invariant 2 check info details",
					},
					Info:        "invariant 2 info",
					InfoDetails: "invariant 2 info details",
				},
			},
			expected: check.ManagerFixResult{
				FixResultType:            check.FixResultTypeSkipped,
				DeterminingInvariantType: nil,
				FixResults: []check.FixResult{
					{
						FixResultType: check.FixResultTypeSkipped,
						InvariantType: check.InvariantType("first"),
						CheckResult: check.CheckResult{
							CheckResultType: check.CheckResultTypeHealthy,
							Info:            "invariant 1 check info",
							InfoDetails:     "invariant 1 check info details",
						},
						Info:        "invariant 1 info",
						InfoDetails: "invariant 1 info details",
					},
					{
						FixResultType: check.FixResultTypeSkipped,
						InvariantType: check.InvariantType("second"),
						CheckResult: check.CheckResult{
							CheckResultType: check.CheckResultTypeHealthy,
							Info:            "invariant 2 check info",
							InfoDetails:     "invariant 2 check info details",
						},
						Info:        "invariant 2 info",
						InfoDetails: "invariant 2 info details",
					},
				},
			},
		},
		{
			fixResults: []check.FixResult{
				{
					FixResultType: check.FixResultTypeSkipped,
					InvariantType: check.InvariantType("first"),
					CheckResult: check.CheckResult{
						CheckResultType: check.CheckResultTypeHealthy,
						Info:            "invariant 1 check info",
						InfoDetails:     "invariant 1 check info details",
					},
					Info:        "invariant 1 info",
					InfoDetails: "invariant 1 info details",
				},
				{
					FixResultType: check.FixResultTypeFixed,
					InvariantType: check.InvariantType("second"),
					CheckResult: check.CheckResult{
						CheckResultType: check.CheckResultTypeCorrupted,
						Info:            "invariant 2 check info",
						InfoDetails:     "invariant 2 check info details",
					},
					Info:        "invariant 2 info",
					InfoDetails: "invariant 2 info details",
				},
				{
					FixResultType: check.FixResultTypeSkipped,
					InvariantType: check.InvariantType("third"),
					CheckResult: check.CheckResult{
						CheckResultType: check.CheckResultTypeHealthy,
						Info:            "invariant 3 check info",
						InfoDetails:     "invariant 3 check info details",
					},
					Info:        "invariant 3 info",
					InfoDetails: "invariant 3 info details",
				},
			},
			expected: check.ManagerFixResult{
				FixResultType:            check.FixResultTypeFixed,
				DeterminingInvariantType: check.InvariantTypePtr("second"),
				FixResults: []check.FixResult{
					{
						FixResultType: check.FixResultTypeSkipped,
						InvariantType: check.InvariantType("first"),
						CheckResult: check.CheckResult{
							CheckResultType: check.CheckResultTypeHealthy,
							Info:            "invariant 1 check info",
							InfoDetails:     "invariant 1 check info details",
						},
						Info:        "invariant 1 info",
						InfoDetails: "invariant 1 info details",
					},
					{
						FixResultType: check.FixResultTypeFixed,
						InvariantType: check.InvariantType("second"),
						CheckResult: check.CheckResult{
							CheckResultType: check.CheckResultTypeCorrupted,
							Info:            "invariant 2 check info",
							InfoDetails:     "invariant 2 check info details",
						},
						Info:        "invariant 2 info",
						InfoDetails: "invariant 2 info details",
					},
					{
						FixResultType: check.FixResultTypeSkipped,
						InvariantType: check.InvariantType("third"),
						CheckResult: check.CheckResult{
							CheckResultType: check.CheckResultTypeHealthy,
							Info:            "invariant 3 check info",
							InfoDetails:     "invariant 3 check info details",
						},
						Info:        "invariant 3 info",
						InfoDetails: "invariant 3 info details",
					},
				},
			},
		},
		{
			fixResults: []check.FixResult{
				{
					FixResultType: check.FixResultTypeSkipped,
					InvariantType: check.InvariantType("first"),
					CheckResult: check.CheckResult{
						CheckResultType: check.CheckResultTypeHealthy,
						Info:            "invariant 1 check info",
						InfoDetails:     "invariant 1 check info details",
					},
					Info:        "invariant 1 info",
					InfoDetails: "invariant 1 info details",
				},
				{
					FixResultType: check.FixResultTypeFixed,
					InvariantType: check.InvariantType("second"),
					CheckResult: check.CheckResult{
						CheckResultType: check.CheckResultTypeCorrupted,
						Info:            "invariant 2 check info",
						InfoDetails:     "invariant 2 check info details",
					},
					Info:        "invariant 2 info",
					InfoDetails: "invariant 2 info details",
				},
				{
					FixResultType: check.FixResultTypeSkipped,
					InvariantType: check.InvariantType("third"),
					CheckResult: check.CheckResult{
						CheckResultType: check.CheckResultTypeHealthy,
						Info:            "invariant 3 check info",
						InfoDetails:     "invariant 3 check info details",
					},
					Info:        "invariant 3 info",
					InfoDetails: "invariant 3 info details",
				},
				{
					FixResultType: check.FixResultTypeFailed,
					InvariantType: check.InvariantType("forth"),
					CheckResult: check.CheckResult{
						CheckResultType: check.CheckResultTypeCorrupted,
						Info:            "invariant 4 check info",
						InfoDetails:     "invariant 2 check info details",
					},
					Info:        "invariant 4 info",
					InfoDetails: "invariant 4 info details",
				},
			},
			expected: check.ManagerFixResult{
				FixResultType:            check.FixResultTypeFailed,
				DeterminingInvariantType: check.InvariantTypePtr("forth"),
				FixResults: []check.FixResult{
					{
						FixResultType: check.FixResultTypeSkipped,
						InvariantType: check.InvariantType("first"),
						CheckResult: check.CheckResult{
							CheckResultType: check.CheckResultTypeHealthy,
							Info:            "invariant 1 check info",
							InfoDetails:     "invariant 1 check info details",
						},
						Info:        "invariant 1 info",
						InfoDetails: "invariant 1 info details",
					},
					{
						FixResultType: check.FixResultTypeFixed,
						InvariantType: check.InvariantType("second"),
						CheckResult: check.CheckResult{
							CheckResultType: check.CheckResultTypeCorrupted,
							Info:            "invariant 2 check info",
							InfoDetails:     "invariant 2 check info details",
						},
						Info:        "invariant 2 info",
						InfoDetails: "invariant 2 info details",
					},
					{
						FixResultType: check.FixResultTypeSkipped,
						InvariantType: check.InvariantType("third"),
						CheckResult: check.CheckResult{
							CheckResultType: check.CheckResultTypeHealthy,
							Info:            "invariant 3 check info",
							InfoDetails:     "invariant 3 check info details",
						},
						Info:        "invariant 3 info",
						InfoDetails: "invariant 3 info details",
					},
					{
						FixResultType: check.FixResultTypeFailed,
						InvariantType: check.InvariantType("forth"),
						CheckResult: check.CheckResult{
							CheckResultType: check.CheckResultTypeCorrupted,
							Info:            "invariant 4 check info",
							InfoDetails:     "invariant 2 check info details",
						},
						Info:        "invariant 4 info",
						InfoDetails: "invariant 4 info details",
					},
				},
			},
		},
		{
			fixResults: []check.FixResult{
				{
					FixResultType: check.FixResultTypeSkipped,
					InvariantType: check.InvariantType("first"),
					CheckResult: check.CheckResult{
						CheckResultType: check.CheckResultTypeHealthy,
						Info:            "invariant 1 check info",
						InfoDetails:     "invariant 1 check info details",
					},
					Info:        "invariant 1 info",
					InfoDetails: "invariant 1 info details",
				},
				{
					FixResultType: check.FixResultTypeFailed,
					InvariantType: check.InvariantType("second"),
					CheckResult: check.CheckResult{
						CheckResultType: check.CheckResultTypeCorrupted,
						Info:            "invariant 4 check info",
						InfoDetails:     "invariant 2 check info details",
					},
					Info:        "invariant 4 info",
					InfoDetails: "invariant 4 info details",
				},
				{
					FixResultType: check.FixResultTypeFixed,
					InvariantType: check.InvariantType("third"),
					CheckResult: check.CheckResult{
						CheckResultType: check.CheckResultTypeCorrupted,
						Info:            "invariant 2 check info",
						InfoDetails:     "invariant 2 check info details",
					},
					Info:        "invariant 2 info",
					InfoDetails: "invariant 2 info details",
				},
				{
					FixResultType: check.FixResultTypeSkipped,
					InvariantType: check.InvariantType("forth"),
					CheckResult: check.CheckResult{
						CheckResultType: check.CheckResultTypeHealthy,
						Info:            "invariant 3 check info",
						InfoDetails:     "invariant 3 check info details",
					},
					Info:        "invariant 3 info",
					InfoDetails: "invariant 3 info details",
				},
			},
			expected: check.ManagerFixResult{
				FixResultType:            check.FixResultTypeFailed,
				DeterminingInvariantType: check.InvariantTypePtr("second"),
				FixResults: []check.FixResult{
					{
						FixResultType: check.FixResultTypeSkipped,
						InvariantType: check.InvariantType("first"),
						CheckResult: check.CheckResult{
							CheckResultType: check.CheckResultTypeHealthy,
							Info:            "invariant 1 check info",
							InfoDetails:     "invariant 1 check info details",
						},
						Info:        "invariant 1 info",
						InfoDetails: "invariant 1 info details",
					},
					{
						FixResultType: check.FixResultTypeFailed,
						InvariantType: check.InvariantType("second"),
						CheckResult: check.CheckResult{
							CheckResultType: check.CheckResultTypeCorrupted,
							Info:            "invariant 4 check info",
							InfoDetails:     "invariant 2 check info details",
						},
						Info:        "invariant 4 info",
						InfoDetails: "invariant 4 info details",
					},
					{
						FixResultType: check.FixResultTypeFixed,
						InvariantType: check.InvariantType("third"),
						CheckResult: check.CheckResult{
							CheckResultType: check.CheckResultTypeCorrupted,
							Info:            "invariant 2 check info",
							InfoDetails:     "invariant 2 check info details",
						},
						Info:        "invariant 2 info",
						InfoDetails: "invariant 2 info details",
					},
					{
						FixResultType: check.FixResultTypeSkipped,
						InvariantType: check.InvariantType("forth"),
						CheckResult: check.CheckResult{
							CheckResultType: check.CheckResultTypeHealthy,
							Info:            "invariant 3 check info",
							InfoDetails:     "invariant 3 check info details",
						},
						Info:        "invariant 3 info",
						InfoDetails: "invariant 3 info details",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		invariants := make([]check.Invariant, len(tc.fixResults), len(tc.fixResults))
		for i := 0; i < len(tc.fixResults); i++ {
			mockInvariant := check.NewMockInvariant(s.controller)
			mockInvariant.EXPECT().Fix(gomock.Any()).Return(tc.fixResults[i])
			invariants[i] = mockInvariant
		}
		manager := &invariantManager{
			invariants: invariants,
		}
		s.Equal(tc.expected, manager.RunFixes(entity.Execution{}))
	}
}
