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

package types

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetMergedPolicyforNode(t *testing.T) {
	npc := NewNodePolicyCollection([]NodePolicy{
		{
			Path: "*", // level 0
			SplitPolicy: &SplitPolicy{
				PredefinedSplits: []any{"timer", "transfer"},
			},
			DispatchPolicy: &DispatchPolicy{DispatchRPS: 100},
		},
		{
			Path:           "*/.", // level 1 default policy
			SplitPolicy:    &SplitPolicy{},
			DispatchPolicy: &DispatchPolicy{DispatchRPS: 50},
		},
		{
			Path: "*/timer", // level 1 timer node
			SplitPolicy: &SplitPolicy{
				PredefinedSplits: []any{"deletehistory"},
			},
		},
		{
			Path: "*/./.", // level 2 default policy
			SplitPolicy: &SplitPolicy{
				PredefinedSplits: []any{"domain1"},
			},
		},
		{
			Path: "*/timer/deletehistory", // level 2 deletehistory timer node policy
			SplitPolicy: &SplitPolicy{
				Disabled: true,
			},
			DispatchPolicy: &DispatchPolicy{DispatchRPS: 5},
		},
		{
			Path: "*/././*", // level 3 default catch-all node policy
			SplitPolicy: &SplitPolicy{
				Disabled: true,
			},
			DispatchPolicy: &DispatchPolicy{DispatchRPS: 1000},
		},
		{
			Path: "*/././domain1", // level 3 domain node policy
			SplitPolicy: &SplitPolicy{
				Disabled: true,
			},
			DispatchPolicy: &DispatchPolicy{DispatchRPS: 42},
		},
	})

	tests := []struct {
		name string
		path string
		want NodePolicy
	}{
		{
			name: "root node",
			path: "*",
			want: NodePolicy{
				Path: "*",
				SplitPolicy: &SplitPolicy{
					PredefinedSplits: []any{"timer", "transfer"},
				},
				DispatchPolicy: &DispatchPolicy{DispatchRPS: 100},
			},
		},
		{
			name: "level 1 catch-all node",
			path: "*/*",
			want: NodePolicy{
				Path:           "*/*",
				SplitPolicy:    &SplitPolicy{},
				DispatchPolicy: &DispatchPolicy{DispatchRPS: 50},
			},
		},
		{
			name: "level 1 timer node",
			path: "*/timer",
			want: NodePolicy{
				Path: "*/timer",
				SplitPolicy: &SplitPolicy{
					PredefinedSplits: []any{"deletehistory"},
				},
				DispatchPolicy: &DispatchPolicy{DispatchRPS: 50},
			},
		},
		{
			name: "level 2 catch all node",
			path: "*/./*",
			want: NodePolicy{
				Path: "*/./*",
				SplitPolicy: &SplitPolicy{
					PredefinedSplits: []any{"domain1"},
				},
				DispatchPolicy: &DispatchPolicy{DispatchRPS: 50},
			},
		},
		{
			name: "level 2 deletehistory timer node",
			path: "*/timer/deletehistory",
			want: NodePolicy{
				Path: "*/timer/deletehistory",
				SplitPolicy: &SplitPolicy{
					Disabled: true,
				},
				DispatchPolicy: &DispatchPolicy{DispatchRPS: 5},
			},
		},
		{
			name: "level 3 catch-all node",
			path: "*/*/*/*",
			want: NodePolicy{
				Path: "*/*/*/*",
				SplitPolicy: &SplitPolicy{
					Disabled: true,
				},
				DispatchPolicy: &DispatchPolicy{DispatchRPS: 1000},
			},
		},
		{
			name: "level 3 domain1 node for activitytimeout",
			path: "*/timer/activitytimeout/domain1",
			want: NodePolicy{
				Path: "*/timer/activitytimeout/domain1",
				SplitPolicy: &SplitPolicy{
					Disabled: true,
				},
				DispatchPolicy: &DispatchPolicy{DispatchRPS: 42},
			},
		},
		{
			name: "level 3 domain1 node for childwfcompleted",
			path: "*/transfer/childwfcompleted/domain1",
			want: NodePolicy{
				Path: "*/transfer/childwfcompleted/domain1",
				SplitPolicy: &SplitPolicy{
					Disabled: true,
				},
				DispatchPolicy: &DispatchPolicy{DispatchRPS: 42},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := npc.GetMergedPolicyForNode(tc.path)
			if err != nil {
				t.Fatalf("failed to get merged policy for node %v: %v", tc.path, err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatalf("Policy mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
