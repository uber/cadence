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

package fetcher

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/uber/cadence/.gen/go/shared"
	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/codec"
	"github.com/uber/cadence/common/persistence"
)

const (
	testTreeID   = "test-tree-id"
	testBranchID = "test-branch-id"
)

var (
	validBranchToken   = []byte{89, 11, 0, 10, 0, 0, 0, 12, 116, 101, 115, 116, 45, 116, 114, 101, 101, 45, 105, 100, 11, 0, 20, 0, 0, 0, 14, 116, 101, 115, 116, 45, 98, 114, 97, 110, 99, 104, 45, 105, 100, 0}
	invalidBranchToken = []byte("invalid")
)

func TestGetBranchToken(t *testing.T) {
	encoder := codec.NewThriftRWEncoder()
	testCases := []struct {
		name        string
		entity      *persistence.ListConcreteExecutionsEntity
		expectError bool
		branchToken []byte
		treeID      string
		branchID    string
	}{
		{
			name: "ValidBranchToken",
			entity: &persistence.ListConcreteExecutionsEntity{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					BranchToken: getValidBranchToken(t, encoder),
				},
			},
			expectError: false,
			branchToken: validBranchToken,
			treeID:      testTreeID,
			branchID:    testBranchID,
		},
		{
			name: "InvalidBranchToken",
			entity: &persistence.ListConcreteExecutionsEntity{
				ExecutionInfo: &persistence.WorkflowExecutionInfo{
					BranchToken: invalidBranchToken,
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			branchToken, branch, err := getBranchToken(tc.entity.ExecutionInfo.BranchToken, tc.entity.VersionHistories, encoder)
			if tc.expectError {
				require.Error(t, err)
				require.Nil(t, branchToken)
				require.Empty(t, branch.GetTreeID())
				require.Empty(t, branch.GetBranchID())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.branchToken, branchToken)
				require.Equal(t, tc.treeID, branch.GetTreeID())
				require.Equal(t, tc.branchID, branch.GetBranchID())
			}
		})
	}
}

func getValidBranchToken(t *testing.T, encoder *codec.ThriftRWEncoder) []byte {
	hb := &shared.HistoryBranch{
		TreeID:   common.StringPtr(testTreeID),
		BranchID: common.StringPtr(testBranchID),
	}
	bytes, err := encoder.Encode(hb)
	require.NoError(t, err)
	return bytes
}
