// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

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

package matching

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/uber/cadence/common/partition"
	"github.com/uber/cadence/common/types"
)

func TestIsolationPickWritePartition(t *testing.T) {
	tl := "tl"
	cases := []struct {
		name            string
		group           string
		isolationGroups []string
		numWrite        int
		shouldFallback  bool
		allowed         []string
	}{
		{
			name:            "single partition",
			group:           "a",
			numWrite:        1,
			isolationGroups: []string{"a"},
			allowed:         []string{tl},
		},
		{
			name:            "multiple partitions - single option",
			group:           "b",
			numWrite:        2,
			isolationGroups: []string{"a", "b"},
			allowed:         []string{getPartitionTaskListName(tl, 1)},
		},
		{
			name:            "multiple partitions - multiple options",
			group:           "a",
			numWrite:        2,
			isolationGroups: []string{"a"},
			allowed:         []string{tl, getPartitionTaskListName(tl, 1)},
		},
		{
			name:            "fallback - no group",
			numWrite:        2,
			isolationGroups: []string{"a"},
			shouldFallback:  true,
			allowed:         []string{"fallback"},
		},
		{
			name:            "fallback - no groups",
			group:           "a",
			numWrite:        2,
			isolationGroups: []string{""},
			shouldFallback:  true,
			allowed:         []string{"fallback"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lb, fallback := createWithMocks(t, tc.isolationGroups, tc.numWrite, tc.numWrite)
			req := &types.AddDecisionTaskRequest{
				DomainUUID: "domainId",
				TaskList: &types.TaskList{
					Name: tl,
					Kind: types.TaskListKindSticky.Ptr(),
				},
			}
			if tc.group != "" {
				req.PartitionConfig = map[string]string{
					partition.IsolationGroupKey: tc.group,
				}
			}
			if tc.shouldFallback {
				fallback.EXPECT().PickWritePartition(int(types.TaskListTypeDecision), req).Return("fallback").Times(1)
			}
			p := lb.PickWritePartition(0, req)
			assert.Contains(t, tc.allowed, p)
		})
	}
}

func TestIsolationPickReadPartition(t *testing.T) {
	tl := "tl"
	cases := []struct {
		name            string
		group           string
		isolationGroups []string
		numRead         int
		numWrite        int
		shouldFallback  bool
		allowed         []string
	}{
		{
			name:            "single partition",
			group:           "a",
			numRead:         1,
			numWrite:        1,
			isolationGroups: []string{"a"},
			allowed:         []string{tl},
		},
		{
			name:            "multiple partitions - single option",
			group:           "b",
			numRead:         2,
			numWrite:        2,
			isolationGroups: []string{"a", "b"},
			allowed:         []string{getPartitionTaskListName(tl, 1)},
		},
		{
			name:            "multiple partitions - multiple options",
			group:           "a",
			numRead:         2,
			numWrite:        2,
			isolationGroups: []string{"a"},
			allowed:         []string{tl, getPartitionTaskListName(tl, 1)},
		},
		{
			name:            "scaling - multiple options",
			group:           "d",
			numRead:         4,
			numWrite:        3,
			isolationGroups: []string{"a", "b", "c", "d"},
			// numRead = 4 means tasks for d could be in the last partition (idx=3)
			// numWrite = 3 means new tasks  for d are being written to the root (idx=0)
			allowed: []string{tl, getPartitionTaskListName(tl, 3)},
		},
		{
			name:            "fallback - no group",
			numRead:         2,
			numWrite:        2,
			isolationGroups: []string{"a"},
			shouldFallback:  true,
			allowed:         []string{"fallback"},
		},
		{
			name:            "fallback - no groups",
			group:           "a",
			numRead:         2,
			numWrite:        2,
			isolationGroups: []string{""},
			shouldFallback:  true,
			allowed:         []string{"fallback"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lb, fallback := createWithMocks(t, tc.isolationGroups, tc.numWrite, tc.numRead)
			req := &types.MatchingQueryWorkflowRequest{
				DomainUUID: "domainId",
				TaskList: &types.TaskList{
					Name: tl,
					Kind: types.TaskListKindSticky.Ptr(),
				},
			}
			if tc.shouldFallback {
				fallback.EXPECT().PickReadPartition(int(types.TaskListTypeDecision), req, tc.group).Return("fallback").Times(1)
			}
			p := lb.PickReadPartition(0, req, tc.group)
			assert.Contains(t, tc.allowed, p)
		})
	}
}

func TestIsolationGetPartitionsForGroup(t *testing.T) {
	cases := []struct {
		name            string
		group           string
		isolationGroups []string
		partitions      int
		expected        []int
	}{
		{
			name:            "single partition",
			group:           "a",
			isolationGroups: []string{"a", "b", "c"},
			partitions:      1,
			expected:        []int{0},
		},
		{
			name:            "partitions less than groups",
			group:           "b",
			isolationGroups: []string{"a", "b", "c"},
			partitions:      2,
			expected:        []int{1},
		},
		{
			name:            "partitions equals groups",
			group:           "c",
			isolationGroups: []string{"a", "b", "c"},
			partitions:      3,
			expected:        []int{2},
		},
		{
			name:            "partitions greater than groups",
			group:           "c",
			isolationGroups: []string{"a", "b", "c"},
			partitions:      4,
			expected:        []int{2},
		},
		{
			name:            "partitions greater than groups - multiple assigned",
			group:           "a",
			isolationGroups: []string{"a", "b", "c"},
			partitions:      4,
			expected:        []int{0, 3},
		},
		{
			name:            "not ok - no isolation group",
			group:           "",
			isolationGroups: []string{"a"},
			partitions:      4,
		},
		{
			name:            "not ok - no isolation groups",
			group:           "a",
			isolationGroups: []string{},
			partitions:      4,
		},
		{
			name:            "not ok - unknown isolation group",
			group:           "d",
			isolationGroups: []string{"a", "b", "c"},
			partitions:      4,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lb, _ := createWithMocks(t, tc.isolationGroups, tc.partitions, tc.partitions)
			actual, ok := lb.getPartitionsForGroup(tc.group, tc.partitions)
			if tc.expected == nil {
				assert.Nil(t, actual)
				assert.False(t, ok)
			} else {
				expectedSet := make(map[int]any, len(tc.expected))
				for _, expectedPartition := range tc.expected {
					expectedSet[expectedPartition] = struct{}{}
				}
				assert.Equal(t, expectedSet, actual)
				assert.True(t, ok)
			}
		})
	}
}

func createWithMocks(t *testing.T, isolationGroups []string, writePartitions, readPartitions int) (*isolationLoadBalancer, *MockLoadBalancer) {
	ctrl := gomock.NewController(t)
	fallback := NewMockLoadBalancer(ctrl)
	cfg := NewMockPartitionConfigProvider(ctrl)
	cfg.EXPECT().GetNumberOfWritePartitions(gomock.Any(), gomock.Any(), gomock.Any()).Return(writePartitions).AnyTimes()
	cfg.EXPECT().GetNumberOfReadPartitions(gomock.Any(), gomock.Any(), gomock.Any()).Return(readPartitions).AnyTimes()
	allIsolationGroups := func() []string {
		return isolationGroups
	}
	return &isolationLoadBalancer{
		provider:           cfg,
		fallback:           fallback,
		allIsolationGroups: allIsolationGroups,
	}, fallback
}
