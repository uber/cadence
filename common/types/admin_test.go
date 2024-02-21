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

	"github.com/stretchr/testify/assert"
)

func TestIsolationGroupConfiguration_ToPartitionList(t *testing.T) {
	tests := map[string]struct {
		in          *IsolationGroupConfiguration
		expectedOut []IsolationGroupPartition
	}{
		"valid value": {
			in: &IsolationGroupConfiguration{
				"zone-1": {
					Name:  "zone-1",
					State: IsolationGroupStateDrained,
				},
				"zone-2": {
					Name:  "zone-2",
					State: IsolationGroupStateHealthy,
				},
				"zone-3": {
					Name:  "zone-3",
					State: IsolationGroupStateDrained,
				},
			},
			expectedOut: []IsolationGroupPartition{
				{
					Name:  "zone-1",
					State: IsolationGroupStateDrained,
				},
				{
					Name:  "zone-2",
					State: IsolationGroupStateHealthy,
				},
				{
					Name:  "zone-3",
					State: IsolationGroupStateDrained,
				},
			},
		},
	}
	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expectedOut, td.in.ToPartitionList())
		})
	}
}

func TestAsyncWorkflowConfigurationDeepCopy(t *testing.T) {
	tests := []struct {
		name  string
		input AsyncWorkflowConfiguration
	}{
		{
			name:  "empty",
			input: AsyncWorkflowConfiguration{},
		},
		{
			name: "predefined queue",
			input: AsyncWorkflowConfiguration{
				Enabled:             true,
				PredefinedQueueName: "test-async-wf-queue",
			},
		},
		{
			name: "custom queue",
			input: AsyncWorkflowConfiguration{
				Enabled:   true,
				QueueType: "custom",
				QueueConfig: &DataBlob{
					EncodingType: EncodingTypeThriftRW.Ptr(),
					Data:         []byte("test-async-wf-queue"),
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.DeepCopy()
			assert.Equal(t, tc.input, got)
			if tc.input.QueueConfig != nil {
				// assert that queue configs look the same but underlying slice is different
				assert.Equal(t, tc.input.QueueConfig, got.QueueConfig)
				if tc.input.QueueConfig.Data != nil && identicalByteArray(tc.input.QueueConfig.Data, got.QueueConfig.Data) {
					t.Error("expected DeepCopy to return a new QueueConfig.Data")
				}
			}
		})
	}
}
