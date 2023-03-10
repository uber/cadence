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

package partition

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/uber/cadence/common/types"
)

func TestDefaultPartitionerIsolationGroupPicking(t *testing.T) {

	isolationGroupsAllHealthy := []types.IsolationGroupPartition{
		{
			Name:   "isolationGroup1",
			Status: types.IsolationGroupStatusHealthy,
		},
		{
			Name:   "isolationGroup2",
			Status: types.IsolationGroupStatusHealthy,
		},
		{
			Name:   "isolationGroup3",
			Status: types.IsolationGroupStatusHealthy,
		},
	}

	isolationGroupsWithOneDrain := []types.IsolationGroupPartition{
		{
			Name:   "isolationGroup1",
			Status: types.IsolationGroupStatusHealthy,
		},
		{
			Name:   "isolationGroup2",
			Status: types.IsolationGroupStatusDrained,
		},
		{
			Name:   "isolationGroup3",
			Status: types.IsolationGroupStatusHealthy,
		},
	}

	tests := map[string]struct {
		isolationGroups []types.IsolationGroupPartition
		wfConfig        DefaultPartitionConfig
		expected        types.IsolationGroupName
	}{
		"default behaviour - no drains - should start in the isolationGroup it's created in 1/3": {
			isolationGroups: isolationGroupsAllHealthy,
			wfConfig: DefaultPartitionConfig{
				WorkflowStartIsolationGroup: "isolationGroup1",
				RunID:                       "d1254200-e457-4998-a2ba-d69c8f59a75f",
			},
			expected: "isolationGroup1",
		},
		"default behaviour - no drains - should start in the isolationGroup it's created in 2/3": {
			isolationGroups: isolationGroupsAllHealthy,
			wfConfig: DefaultPartitionConfig{
				WorkflowStartIsolationGroup: "isolationGroup2",
				RunID:                       "d1254200-e457-4998-a2ba-d69c8f59a75f",
			},
			expected: "isolationGroup2",
		},
		"default behaviour - no drains - should start in the isolationGroup it's created in 3/3": {
			isolationGroups: isolationGroupsAllHealthy,
			wfConfig: DefaultPartitionConfig{
				WorkflowStartIsolationGroup: "isolationGroup3",
				RunID:                       "d1254200-e457-4998-a2ba-d69c8f59a75f",
			},
			expected: "isolationGroup3",
		},
		"failover behaviour - should be able to fallback to a random deterministic partition if the isolationGroup isn't found": {
			isolationGroups: isolationGroupsWithOneDrain,
			wfConfig: DefaultPartitionConfig{
				WorkflowStartIsolationGroup: "isolationGroup2",
				RunID:                       "d1254200-e457-4998-a2ba-d69c8f59a75f",
			},
			expected: "isolationGroup3",
		},
		"regional failover case - should be able to fallback to a random deterministic partition if the isolationGroup isn't found": {
			isolationGroups: isolationGroupsAllHealthy,
			wfConfig: DefaultPartitionConfig{
				WorkflowStartIsolationGroup: "not-found-isolationGroup",
				RunID:                       "d1254200-e457-4998-a2ba-d69c8f59a75f",
			},
			expected: "isolationGroup1",
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expected, pickIsolationGroupAfterDrain(td.isolationGroups, td.wfConfig))
		})
	}
}

func TestDefaultPartitionerPickerDistribution(t *testing.T) {

	countHealthy := make(map[types.IsolationGroupName]int)

	var isolationGroups []types.IsolationGroupPartition

	for i := 0; i < 100; i++ {
		healthy := types.IsolationGroupPartition{
			Name:   types.IsolationGroupName(fmt.Sprintf("isolationGroup-%d", i)),
			Status: types.IsolationGroupStatusHealthy,
		}

		unhealthy := types.IsolationGroupPartition{
			Name:   types.IsolationGroupName(fmt.Sprintf("isolationGroup-%d", i)),
			Status: types.IsolationGroupStatusHealthy,
		}

		isolationGroups = append(isolationGroups, healthy)
		isolationGroups = append(isolationGroups, unhealthy)
		countHealthy[healthy.Name] = 0
	}

	for i := 0; i < 100000; i++ {
		result := pickIsolationGroupAfterDrain(isolationGroups, DefaultPartitionConfig{
			WorkflowStartIsolationGroup: "not-a-present-isolationGroup", // always force a fallback to the simple hash
			RunID:                       uuid.New().String(),
		})

		count, ok := countHealthy[result]
		if !ok {
			t.Fatal("the result wasn't found in the healthy list, something is wrong with the logic for selecting healthy isolationGroups")
		}
		countHealthy[result] = count + 1
	}

	for k, v := range countHealthy {
		assert.True(t, v > 0, "failed to pick a isolationGroup %s", k)
	}
}
