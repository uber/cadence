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

func TestIsolationGroupConfigurationDeepCopy(t *testing.T) {
	tests := []struct {
		name  string
		input IsolationGroupConfiguration
	}{
		{
			name:  "nil",
			input: nil,
		},
		{
			name:  "empty",
			input: IsolationGroupConfiguration{},
		},
		{
			name: "single group",
			input: IsolationGroupConfiguration{
				"zone-1": {
					Name:  "zone-1",
					State: IsolationGroupStateDrained,
				},
			},
		},
		{
			name: "multiple groups",
			input: IsolationGroupConfiguration{
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
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.DeepCopy()
			assert.Equal(t, tc.input, got)

			if tc.input == nil {
				return
			}

			tc.input["new"] = IsolationGroupPartition{Name: "new", State: IsolationGroupStateHealthy}
			assert.NotEqual(t, tc.input, got)
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

func TestAddSearchAttributeRequest_SerializeForLogging(t *testing.T) {
	tests := []struct {
		name     string
		request  *AddSearchAttributeRequest
		expected string
	}{
		{
			name:     "nil",
			request:  nil,
			expected: "",
		},
		{
			name:     "empty",
			request:  &AddSearchAttributeRequest{},
			expected: "{}",
		},
		{
			name: "full",
			request: &AddSearchAttributeRequest{
				SearchAttribute: map[string]IndexedValueType{"key": IndexedValueTypeString},
				SecurityToken:   "test-token",
			},
			expected: `{"searchAttribute":{"key":"STRING"},"securityToken":"test-token"}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			serialized, err := tc.request.SerializeForLogging()

			assert.NoError(t, err)

			assert.Equal(t, tc.expected, serialized)
		})
	}
}

func TestAddSearchAttributeRequest_GetSearchAttribute(t *testing.T) {
	tests := []struct {
		name     string
		request  *AddSearchAttributeRequest
		expected map[string]IndexedValueType
	}{
		{
			name:     "nil",
			request:  nil,
			expected: map[string]IndexedValueType(nil),
		},
		{
			name:     "empty",
			request:  &AddSearchAttributeRequest{},
			expected: map[string]IndexedValueType(nil),
		},
		{
			name: "full",
			request: &AddSearchAttributeRequest{
				SearchAttribute: map[string]IndexedValueType{"key": IndexedValueTypeString},
				SecurityToken:   "test-token",
			},
			expected: map[string]IndexedValueType{"key": IndexedValueTypeString},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			searchAttributes := tc.request.GetSearchAttribute()

			assert.Equal(t, tc.expected, searchAttributes)
		})
	}
}

func TestAdminDescribeWorkflowExecutionRequest_SerializeForLogging(t *testing.T) {
	tests := []struct {
		name     string
		request  *AdminDescribeWorkflowExecutionRequest
		expected string
	}{
		{
			name:     "nil",
			request:  nil,
			expected: "",
		},
		{
			name:     "empty",
			request:  &AdminDescribeWorkflowExecutionRequest{},
			expected: "{}",
		},
		{
			name: "full",
			request: &AdminDescribeWorkflowExecutionRequest{
				Domain: "test-domain",
				Execution: &WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				},
			},
			expected: `{"domain":"test-domain","execution":{"workflowId":"test-workflow-id","runId":"test-run-id"}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			serialized, err := tc.request.SerializeForLogging()

			assert.NoError(t, err)

			assert.Equal(t, tc.expected, serialized)
		})
	}
}

func TestAdminDescribeWorkflowExecutionRequest_GetDomain(t *testing.T) {
	tests := []struct {
		name     string
		request  *AdminDescribeWorkflowExecutionRequest
		expected string
	}{
		{
			name:     "nil",
			request:  nil,
			expected: "",
		},
		{
			name:     "empty",
			request:  &AdminDescribeWorkflowExecutionRequest{},
			expected: "",
		},
		{
			name: "full",
			request: &AdminDescribeWorkflowExecutionRequest{
				Domain: "test-domain",
				Execution: &WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				},
			},
			expected: "test-domain",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			domain := tc.request.GetDomain()

			assert.Equal(t, tc.expected, domain)
		})
	}
}

func TestAdminDescribeWorkflowExecutionResponse_GetShardID(t *testing.T) {
	tests := []struct {
		name     string
		response *AdminDescribeWorkflowExecutionResponse
		expected string
	}{
		{
			name:     "nil",
			response: nil,
			expected: "",
		},
		{
			name:     "empty",
			response: &AdminDescribeWorkflowExecutionResponse{},
			expected: "",
		},
		{
			name: "full",
			response: &AdminDescribeWorkflowExecutionResponse{
				ShardID: "test-shard-id",
			},
			expected: "test-shard-id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			shardID := tc.response.GetShardID()

			assert.Equal(t, tc.expected, shardID)
		})
	}
}

func TestAdminDescribeWorkflowExecutionResponse_GetMutableStateInDatabase(t *testing.T) {
	tests := []struct {
		name     string
		response *AdminDescribeWorkflowExecutionResponse
		expected string
	}{
		{
			name:     "nil",
			response: nil,
			expected: "",
		},
		{
			name:     "empty",
			response: &AdminDescribeWorkflowExecutionResponse{},
			expected: "",
		},
		{
			name: "full",
			response: &AdminDescribeWorkflowExecutionResponse{
				MutableStateInDatabase: "true",
			},
			expected: "true",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mutableStateInDatabase := tc.response.GetMutableStateInDatabase()

			assert.Equal(t, tc.expected, mutableStateInDatabase)
		})
	}
}

func TestGetWorkflowExecutionRawHistoryV2Request_SerializeForLogging(t *testing.T) {
	tests := []struct {
		name     string
		request  *GetWorkflowExecutionRawHistoryV2Request
		expected string
	}{
		{
			name:     "nil",
			request:  nil,
			expected: "",
		},
		{
			name:     "empty",
			request:  &GetWorkflowExecutionRawHistoryV2Request{},
			expected: "{}",
		},
		{
			name: "full",
			request: &GetWorkflowExecutionRawHistoryV2Request{
				Domain: "test-domain",
			},
			expected: `{"domain":"test-domain"}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			serialized, err := tc.request.SerializeForLogging()

			assert.NoError(t, err)

			assert.Equal(t, tc.expected, serialized)
		})
	}
}

func TestGetWorkflowExecutionRawHistoryV2Request_GetDomain(t *testing.T) {
	tests := []struct {
		name     string
		request  *GetWorkflowExecutionRawHistoryV2Request
		expected string
	}{
		{
			name:     "nil",
			request:  nil,
			expected: "",
		},
		{
			name:     "empty",
			request:  &GetWorkflowExecutionRawHistoryV2Request{},
			expected: "",
		},
		{
			name: "full",
			request: &GetWorkflowExecutionRawHistoryV2Request{
				Domain: "test-domain",
			},
			expected: "test-domain",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			domain := tc.request.GetDomain()

			assert.Equal(t, tc.expected, domain)
		})
	}
}

func TestGetWorkflowExecutionRawHistoryV2Request_GetStartEventID(t *testing.T) {
	tests := []struct {
		name     string
		request  *GetWorkflowExecutionRawHistoryV2Request
		expected int64
	}{
		{
			name:     "nil",
			request:  nil,
			expected: 0,
		},
		{
			name:     "empty",
			request:  &GetWorkflowExecutionRawHistoryV2Request{},
			expected: 0,
		},
		{
			name: "full",
			request: &GetWorkflowExecutionRawHistoryV2Request{
				StartEventID: func() *int64 { a := int64(123); return &a }(),
			},
			expected: 123,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			startEventID := tc.request.GetStartEventID()

			assert.Equal(t, tc.expected, startEventID)
		})
	}
}

func TestGetWorkflowExecutionRawHistoryV2Request_GetStartEventVersion(t *testing.T) {
	tests := []struct {
		name     string
		request  *GetWorkflowExecutionRawHistoryV2Request
		expected int64
	}{
		{
			name:     "nil",
			request:  nil,
			expected: 0,
		},
		{
			name:     "empty",
			request:  &GetWorkflowExecutionRawHistoryV2Request{},
			expected: 0,
		},
		{
			name: "full",
			request: &GetWorkflowExecutionRawHistoryV2Request{
				StartEventVersion: func() *int64 { a := int64(123); return &a }(),
			},
			expected: 123,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			startEventVersion := tc.request.GetStartEventVersion()

			assert.Equal(t, tc.expected, startEventVersion)
		})
	}
}

func TestGetWorkflowExecutionRawHistoryV2Request_GetEndEventID(t *testing.T) {
	tests := []struct {
		name     string
		request  *GetWorkflowExecutionRawHistoryV2Request
		expected int64
	}{
		{
			name:     "nil",
			request:  nil,
			expected: 0,
		},
		{
			name:     "empty",
			request:  &GetWorkflowExecutionRawHistoryV2Request{},
			expected: 0,
		},
		{
			name: "full",
			request: &GetWorkflowExecutionRawHistoryV2Request{
				EndEventID: func() *int64 { a := int64(123); return &a }(),
			},
			expected: 123,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			endEventID := tc.request.GetEndEventID()

			assert.Equal(t, tc.expected, endEventID)
		})
	}
}

func TestGetWorkflowExecutionRawHistoryV2Request_GetEndEventVersion(t *testing.T) {
	tests := []struct {
		name     string
		request  *GetWorkflowExecutionRawHistoryV2Request
		expected int64
	}{
		{
			name:     "nil",
			request:  nil,
			expected: 0,
		},
		{
			name:     "empty",
			request:  &GetWorkflowExecutionRawHistoryV2Request{},
			expected: 0,
		},
		{
			name: "full",
			request: &GetWorkflowExecutionRawHistoryV2Request{
				EndEventVersion: func() *int64 { a := int64(123); return &a }(),
			},
			expected: 123,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			endEventVersion := tc.request.GetEndEventVersion()

			assert.Equal(t, tc.expected, endEventVersion)
		})
	}
}

func TestGetWorkflowExecutionRawHistoryV2Request_GetMaximumPageSize(t *testing.T) {
	tests := []struct {
		name     string
		request  *GetWorkflowExecutionRawHistoryV2Request
		expected int32
	}{
		{
			name:     "nil",
			request:  nil,
			expected: 0,
		},
		{
			name:     "empty",
			request:  &GetWorkflowExecutionRawHistoryV2Request{},
			expected: 0,
		},
		{
			name: "full",
			request: &GetWorkflowExecutionRawHistoryV2Request{
				MaximumPageSize: 123,
			},
			expected: 123,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			maximumPageSize := tc.request.GetMaximumPageSize()

			assert.Equal(t, tc.expected, maximumPageSize)
		})
	}
}

func TestGetWorkflowExecutionRawHistoryV2Response_GetHistoryBatches(t *testing.T) {
	tests := []struct {
		name     string
		response *GetWorkflowExecutionRawHistoryV2Response
		expected []*DataBlob
	}{
		{
			name:     "nil",
			response: nil,
			expected: nil,
		},
		{
			name:     "empty",
			response: &GetWorkflowExecutionRawHistoryV2Response{},
			expected: nil,
		},
		{
			name: "full",
			response: &GetWorkflowExecutionRawHistoryV2Response{
				HistoryBatches: []*DataBlob{
					{Data: []byte("test-blob-1")},
					{Data: []byte("test-blob-2")},
				},
			},
			expected: []*DataBlob{
				{Data: []byte("test-blob-1")},
				{Data: []byte("test-blob-2")},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			historyBatches := tc.response.GetHistoryBatches()

			assert.Equal(t, tc.expected, historyBatches)
		})
	}
}

func TestGetWorkflowExecutionRawHistoryV2Response_GetVersionHistory(t *testing.T) {
	tests := []struct {
		name     string
		response *GetWorkflowExecutionRawHistoryV2Response
		expected *VersionHistory
	}{
		{
			name:     "nil",
			response: nil,
			expected: nil,
		},
		{
			name:     "empty",
			response: &GetWorkflowExecutionRawHistoryV2Response{},
			expected: nil,
		},
		{
			name: "full",
			response: &GetWorkflowExecutionRawHistoryV2Response{
				VersionHistory: &VersionHistory{
					BranchToken: []byte("test-branch-token"),
					Items: []*VersionHistoryItem{
						{
							EventID: 123,
							Version: 456,
						},
					},
				},
			},
			expected: &VersionHistory{
				BranchToken: []byte("test-branch-token"),
				Items: []*VersionHistoryItem{
					{
						EventID: 123,
						Version: 456,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			versionHistory := tc.response.GetVersionHistory()

			assert.Equal(t, tc.expected, versionHistory)
		})
	}
}

func TestResendReplicationTasksRequest_SerializeForLogging(t *testing.T) {
	tests := []struct {
		name     string
		request  *ResendReplicationTasksRequest
		expected string
	}{
		{
			name:     "nil",
			request:  nil,
			expected: "",
		},
		{
			name:     "empty",
			request:  &ResendReplicationTasksRequest{},
			expected: "{}",
		},
		{
			name: "full",
			request: &ResendReplicationTasksRequest{
				DomainID:      "test-domain-id",
				WorkflowID:    "test-workflow-id",
				RunID:         "test-run-id",
				RemoteCluster: "test-remote-cluster",
			},
			expected: `{"domainID":"test-domain-id","workflowID":"test-workflow-id","runID":"test-run-id","remoteCluster":"test-remote-cluster"}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			serialized, err := tc.request.SerializeForLogging()

			assert.NoError(t, err)

			assert.Equal(t, tc.expected, serialized)
		})
	}
}

func TestResendReplicationTasksRequest_GetWorkflowID(t *testing.T) {
	tests := []struct {
		name     string
		request  *ResendReplicationTasksRequest
		expected string
	}{
		{
			name:     "nil",
			request:  nil,
			expected: "",
		},
		{
			name:     "empty",
			request:  &ResendReplicationTasksRequest{},
			expected: "",
		},
		{
			name: "full",
			request: &ResendReplicationTasksRequest{
				WorkflowID: "test-workflow-id",
			},
			expected: "test-workflow-id",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			workflowID := tc.request.GetWorkflowID()

			assert.Equal(t, tc.expected, workflowID)
		})
	}
}

func TestResendReplicationTasksRequest_GetRunID(t *testing.T) {
	tests := []struct {
		name     string
		request  *ResendReplicationTasksRequest
		expected string
	}{
		{
			name:     "nil",
			request:  nil,
			expected: "",
		},
		{
			name:     "empty",
			request:  &ResendReplicationTasksRequest{},
			expected: "",
		},
		{
			name: "full",
			request: &ResendReplicationTasksRequest{
				RunID: "test-run-id",
			},
			expected: "test-run-id",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runID := tc.request.GetRunID()

			assert.Equal(t, tc.expected, runID)
		})
	}
}

func TestResendReplicationTasksRequest_GetRemoteCluster(t *testing.T) {
	tests := []struct {
		name     string
		request  *ResendReplicationTasksRequest
		expected string
	}{
		{
			name:     "nil",
			request:  nil,
			expected: "",
		},
		{
			name:     "empty",
			request:  &ResendReplicationTasksRequest{},
			expected: "",
		},
		{
			name: "full",
			request: &ResendReplicationTasksRequest{
				RemoteCluster: "test-remote-cluster",
			},
			expected: "test-remote-cluster",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			remoteCluster := tc.request.GetRemoteCluster()

			assert.Equal(t, tc.expected, remoteCluster)
		})
	}
}

func TestGetDynamicConfigRequest_SerializeForLogging(t *testing.T) {
	tests := []struct {
		name     string
		request  *GetDynamicConfigRequest
		expected string
	}{
		{
			name:     "nil",
			request:  nil,
			expected: "",
		},
		{
			name:     "empty",
			request:  &GetDynamicConfigRequest{},
			expected: "{}",
		},
		{
			name: "full",
			request: &GetDynamicConfigRequest{
				ConfigName: "test-name",
			},
			expected: `{"configName":"test-name"}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			serialized, err := tc.request.SerializeForLogging()

			assert.NoError(t, err)

			assert.Equal(t, tc.expected, serialized)
		})
	}
}

func TestUpdateDynamicConfigRequest_SerializeForLogging(t *testing.T) {
	tests := []struct {
		name     string
		request  *UpdateDynamicConfigRequest
		expected string
	}{
		{
			name:     "nil",
			request:  nil,
			expected: "",
		},
		{
			name:     "empty",
			request:  &UpdateDynamicConfigRequest{},
			expected: "{}",
		},
		{
			name: "full",
			request: &UpdateDynamicConfigRequest{
				ConfigName:   "test-name",
				ConfigValues: []*DynamicConfigValue{{Value: &DataBlob{EncodingType: EncodingTypeThriftRW.Ptr(), Data: []byte("test-value")}}},
			},
			expected: `{"configName":"test-name","configValues":[{"value":{"EncodingType":"ThriftRW","Data":"dGVzdC12YWx1ZQ=="}}]}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			serialized, err := tc.request.SerializeForLogging()

			assert.NoError(t, err)

			assert.Equal(t, tc.expected, serialized)
		})
	}
}

func TestRestoreDynamicConfigRequest_SerializeForLogging(t *testing.T) {
	tests := []struct {
		name     string
		request  *RestoreDynamicConfigRequest
		expected string
	}{
		{
			name:     "nil",
			request:  nil,
			expected: "",
		},
		{
			name:     "empty",
			request:  &RestoreDynamicConfigRequest{},
			expected: "{}",
		},
		{
			name: "full",
			request: &RestoreDynamicConfigRequest{
				ConfigName: "test-name",
			},
			expected: `{"configName":"test-name"}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			serialized, err := tc.request.SerializeForLogging()

			assert.NoError(t, err)

			assert.Equal(t, tc.expected, serialized)
		})
	}
}

func TestAdminDeleteWorkflowRequest_SerializeForLogging(t *testing.T) {
	tests := []struct {
		name     string
		request  *AdminDeleteWorkflowRequest
		expected string
	}{
		{
			name:     "nil",
			request:  nil,
			expected: "",
		},
		{
			name:     "empty",
			request:  &AdminDeleteWorkflowRequest{},
			expected: "{}",
		},
		{
			name: "full",
			request: &AdminDeleteWorkflowRequest{
				Domain: "test-domain",
				Execution: &WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				},
			},
			expected: `{"domain":"test-domain","execution":{"workflowId":"test-workflow-id","runId":"test-run-id"}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			serialized, err := tc.request.SerializeForLogging()

			assert.NoError(t, err)

			assert.Equal(t, tc.expected, serialized)
		})
	}
}

func TestAdminDeleteWorkflowRequest_GetDomain(t *testing.T) {
	tests := []struct {
		name     string
		request  *AdminDeleteWorkflowRequest
		expected string
	}{
		{
			name:     "nil",
			request:  nil,
			expected: "",
		},
		{
			name:     "empty",
			request:  &AdminDeleteWorkflowRequest{},
			expected: "",
		},
		{
			name: "full",
			request: &AdminDeleteWorkflowRequest{
				Domain: "test-domain",
				Execution: &WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				},
			},
			expected: "test-domain",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			domain := tc.request.GetDomain()

			assert.Equal(t, tc.expected, domain)
		})
	}
}

func TestAdminDeleteWorkflowRequest_GetExecution(t *testing.T) {
	tests := []struct {
		name     string
		request  *AdminDeleteWorkflowRequest
		expected *WorkflowExecution
	}{
		{
			name:     "nil",
			request:  nil,
			expected: nil,
		},
		{
			name:     "empty",
			request:  &AdminDeleteWorkflowRequest{},
			expected: nil,
		},
		{
			name: "full",
			request: &AdminDeleteWorkflowRequest{
				Domain: "test-domain",
				Execution: &WorkflowExecution{
					WorkflowID: "test-workflow-id",
					RunID:      "test-run-id",
				},
			},
			expected: &WorkflowExecution{
				WorkflowID: "test-workflow-id",
				RunID:      "test-run-id",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			execution := tc.request.GetExecution()

			assert.Equal(t, tc.expected, execution)
		})
	}
}

func TestAdminDeleteWorkflowRequest_GetSkipErrors(t *testing.T) {
	tests := []struct {
		name     string
		request  *AdminDeleteWorkflowRequest
		expected bool
	}{
		{
			name:     "nil",
			request:  nil,
			expected: false,
		},
		{
			name:     "empty",
			request:  &AdminDeleteWorkflowRequest{},
			expected: false,
		},
		{
			name: "full",
			request: &AdminDeleteWorkflowRequest{
				SkipErrors: true,
			},
			expected: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			skipErrors := tc.request.GetSkipErrors()

			assert.Equal(t, tc.expected, skipErrors)
		})
	}
}

func TestListDynamicConfigRequest_SerializeForLogging(t *testing.T) {
	tests := []struct {
		name     string
		request  *ListDynamicConfigRequest
		expected string
	}{
		{
			name:     "nil",
			request:  nil,
			expected: "",
		},
		{
			name:     "empty",
			request:  &ListDynamicConfigRequest{},
			expected: "{}",
		},
		{
			name: "full",
			request: &ListDynamicConfigRequest{
				ConfigName: "test-name",
			},
			expected: `{"configName":"test-name"}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			serialized, err := tc.request.SerializeForLogging()

			assert.NoError(t, err)

			assert.Equal(t, tc.expected, serialized)
		})
	}
}

func TestFromIsolationGroupPartitionList(t *testing.T) {
	tests := map[string]struct {
		in          []IsolationGroupPartition
		expectedOut IsolationGroupConfiguration
	}{
		"valid value": {
			in: []IsolationGroupPartition{
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
			expectedOut: IsolationGroupConfiguration{
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
		},
		"empty value": {
			in:          nil,
			expectedOut: IsolationGroupConfiguration{},
		},
	}
	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, td.expectedOut, FromIsolationGroupPartitionList(td.in))
		})
	}
}

func TestGetGlobalIsolationGroupsRequest_SerializeForLogging(t *testing.T) {
	tests := map[string]struct {
		request  *GetGlobalIsolationGroupsRequest
		expected string
	}{
		"nil": {
			request:  nil,
			expected: "",
		},
		"empty": {
			request:  &GetGlobalIsolationGroupsRequest{},
			expected: "{}",
		},
	}
	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			serialized, err := td.request.SerializeForLogging()

			assert.NoError(t, err)

			assert.Equal(t, td.expected, serialized)
		})
	}
}

func TestUpdateGlobalIsolationGroupsRequest_SerializeForLogging(t *testing.T) {
	tests := map[string]struct {
		request  *UpdateGlobalIsolationGroupsRequest
		expected string
	}{
		"nil": {
			request:  nil,
			expected: "",
		},
		"empty": {
			request:  &UpdateGlobalIsolationGroupsRequest{},
			expected: `{"IsolationGroups":null}`,
		},
		"full": {
			request: &UpdateGlobalIsolationGroupsRequest{
				IsolationGroups: IsolationGroupConfiguration{
					"zone-1": {
						Name:  "zone-1",
						State: IsolationGroupStateDrained,
					},
				},
			},
			expected: `{"IsolationGroups":{"zone-1":{"Name":"zone-1","State":2}}}`,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			serialized, err := td.request.SerializeForLogging()

			assert.NoError(t, err)

			assert.Equal(t, td.expected, serialized)
		})
	}
}

func TestGetDomainIsolationGroupsRequest_SerializeForLogging(t *testing.T) {
	tests := map[string]struct {
		request  *GetDomainIsolationGroupsRequest
		expected string
	}{
		"nil": {
			request:  nil,
			expected: "",
		},
		"empty": {
			request:  &GetDomainIsolationGroupsRequest{},
			expected: `{"Domain":""}`,
		},
		"full": {
			request: &GetDomainIsolationGroupsRequest{
				Domain: "test-domain",
			},
			expected: `{"Domain":"test-domain"}`,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			serialized, err := td.request.SerializeForLogging()

			assert.NoError(t, err)

			assert.Equal(t, td.expected, serialized)
		})
	}
}

func TestUpdateDomainIsolationGroupsRequest_SerializeForLogging(t *testing.T) {
	tests := map[string]struct {
		request  *UpdateDomainIsolationGroupsRequest
		expected string
	}{
		"nil": {
			request:  nil,
			expected: "",
		},
		"empty": {
			request:  &UpdateDomainIsolationGroupsRequest{},
			expected: `{"Domain":"","IsolationGroups":null}`,
		},
		"full": {
			request: &UpdateDomainIsolationGroupsRequest{
				Domain: "test-domain",
				IsolationGroups: IsolationGroupConfiguration{
					"zone-1": {
						Name:  "zone-1",
						State: IsolationGroupStateDrained,
					},
				},
			},
			expected: `{"Domain":"test-domain","IsolationGroups":{"zone-1":{"Name":"zone-1","State":2}}}`,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			serialized, err := td.request.SerializeForLogging()

			assert.NoError(t, err)

			assert.Equal(t, td.expected, serialized)
		})
	}
}

func TestGetDomainAsyncWorkflowConfiguratonRequest_SerializeForLogging(t *testing.T) {
	tests := map[string]struct {
		request  *GetDomainAsyncWorkflowConfiguratonRequest
		expected string
	}{
		"nil": {
			request:  nil,
			expected: "",
		},
		"empty": {
			request:  &GetDomainAsyncWorkflowConfiguratonRequest{},
			expected: `{"Domain":""}`,
		},
		"full": {
			request: &GetDomainAsyncWorkflowConfiguratonRequest{
				Domain: "test-domain",
			},
			expected: `{"Domain":"test-domain"}`,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			serialized, err := td.request.SerializeForLogging()

			assert.NoError(t, err)

			assert.Equal(t, td.expected, serialized)
		})
	}
}

func TestUpdateDomainAsyncWorkflowConfiguratonRequest_SerializeForLogging(t *testing.T) {
	tests := map[string]struct {
		request  *UpdateDomainAsyncWorkflowConfiguratonRequest
		expected string
	}{
		"nil": {
			request:  nil,
			expected: "",
		},
		"empty": {
			request:  &UpdateDomainAsyncWorkflowConfiguratonRequest{},
			expected: `{"Domain":"","Configuration":null}`,
		},
		"full": {
			request: &UpdateDomainAsyncWorkflowConfiguratonRequest{
				Domain: "test-domain",
				Configuration: &AsyncWorkflowConfiguration{
					Enabled:             true,
					PredefinedQueueName: "test-queue",
					QueueType:           "custom",
					QueueConfig: &DataBlob{
						EncodingType: EncodingTypeThriftRW.Ptr(),
					},
				},
			},
			expected: `{"Domain":"test-domain","Configuration":{"Enabled":true,"PredefinedQueueName":"test-queue","QueueType":"custom","QueueConfig":{"EncodingType":"ThriftRW"}}}`,
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			serialized, err := td.request.SerializeForLogging()

			assert.NoError(t, err)

			assert.Equal(t, td.expected, serialized)
		})
	}
}
