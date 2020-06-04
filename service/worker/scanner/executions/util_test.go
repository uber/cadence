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

package executions

import "github.com/uber/cadence/common"

func (s *workflowsSuite) TestFlattenShards() {
	testCases := []struct {
		input    Shards
		expectedList []int
		expectedMin int
		expectedMax int
	}{
		{
			input: Shards{
				List: []int{1, 2, 3},
			},
			expectedList: []int{1, 2, 3},
			expectedMin: 1,
			expectedMax: 3,
		},
		{
			input: Shards{
				Range: &ShardRange{
					Min: 5,
					Max: 10,
				},
			},
			expectedList: []int{5, 6, 7, 8, 9},
			expectedMin: 5,
			expectedMax: 9,
		},
		{
			input: Shards{
				List: []int{90, 1, 2, 3},
			},
			expectedList: []int{90, 1, 2, 3},
			expectedMin: 1,
			expectedMax: 90,
		},
	}
	for _, tc := range testCases {
		shardList, min, max := flattenShards(tc.input)
		s.Equal(tc.expectedList, shardList)
		s.Equal(tc.expectedMin, min)
		s.Equal(tc.expectedMax, max)
	}
}

func (s *workflowsSuite) TestResolveFixerConfig() {
	result := resolveFixerConfig(FixerWorkflowConfigOverwrites{
		Concurrency: common.IntPtr(1000),
	})
	s.Equal(ResolvedFixerWorkflowConfig{
		Concurrency:             1000,
		BlobstoreFlushThreshold: 1000,
		InvariantCollections: InvariantCollections{
			InvariantCollectionMutableState: true,
			InvariantCollectionHistory:      true,
		},
	}, result)
}

func (s *workflowsSuite) TestValidateShards() {
	testCases := []struct{
		shards Shards
		expectErr bool
	}{
		{
			shards: Shards{},
			expectErr: true,
		},
		{
			shards: Shards{
				List: []int{},
				Range: &ShardRange{},
			},
			expectErr: true,
		},
		{
			shards: Shards{
				List: []int{},
			},
			expectErr: true,
		},
		{
			shards: Shards{
				Range: &ShardRange{
					Min: 0,
					Max: 0,
				},
			},
			expectErr: true,
		},
		{
			shards: Shards{
				Range: &ShardRange{
					Min: 0,
					Max: 1,
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		err := validateShards(tc.shards)
		if tc.expectErr {
			s.Error(err)
		} else {
			s.NoError(err)
		}
	}
}